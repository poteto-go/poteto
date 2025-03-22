package poteto

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	stdContext "context"

	"github.com/caarlos0/env/v11"
	"github.com/fatih/color"
	"github.com/poteto-go/poteto/constant"
	"github.com/poteto-go/poteto/utils"
)

type Poteto interface {
	Router() *router
	MiddlewareTree() *middlewareTree

	// If requested, call this
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Run(addr string) error
	RunTLS(addr string, cert, key []byte) error
	Stop(ctx stdContext.Context) error
	setupServer() error
	Register(middlewares ...MiddlewareFunc)
	Combine(pattern string, middlewares ...MiddlewareFunc) *middlewareTree
	SetLogger(logger any)

	// Leaf makes router group
	// You can make your router clear
	// with middlewares
	Leaf(basePath string, handler LeafHandler)

	// add router & middleware tree from api (Poteto)
	AddApi(api Poteto)

	// workflow is a function that is executed when the server starts | end
	// - constant.StartUpWorkflow: "startUp"
	//  - This is a workflow that is executed when the server starts
	RegisterWorkflow(workflowType string, priority uint, workflow WorkflowFunc)

	GET(path string, handler HandlerFunc) error
	POST(path string, handler HandlerFunc) error
	PUT(path string, handler HandlerFunc) error
	PATCH(path string, handler HandlerFunc) error
	DELETE(path string, handler HandlerFunc) error
	HEAD(path string, handler HandlerFunc) error
	OPTIONS(path string, handler HandlerFunc) error
	TRACE(path string, handler HandlerFunc) error
	CONNECT(path string, handler HandlerFunc) error

	// poteto.Play make ut w/o server
	// EX:
	//  p := poteto.New()
	//  p.GET("/users", func(ctx poteto.Context) error {
	//    return ctx.JSON(http.StatusOK, map[string]string{
	//      "id":   "1",
	//      "name": "tester",
	//    })
	//  })
	//  res := p.Play(http.MethodGet, "/users")
	//  resBodyStr := res.Body.String
	//  // => {"id":"1","name":"tester"}
	Play(method, path string, body ...string) *httptest.ResponseRecorder

	// check if target route has handler.
	// func main() {
	//   p := poteto.New()
	//
	//   p.GET("/users", handler)
	//
	//   ok := p.Check(http.MethodGet, "/users") // -> return true
	//   ng := p.Check(http.MethodPost, "/users") // -> return false
	// }
	Check(method, path string) bool

	// Chain Handler & middleware
	//
	// This is high-readability on just-one-path middleware
	//
	// func main() {
	//   p := poteto.New()
	//
	//   p.GET(
	//     "/users",
	//     p.Chain(
	//       middleware1,
	//       middleware2,
	//     )(handler)
	//   )
	// }
	Chain(middlewares ...MiddlewareFunc) func(HandlerFunc) HandlerFunc
}

type poteto struct {
	router          Router
	errorHandler    HttpErrorHandler
	middlewareTree  MiddlewareTree
	logger          any
	cache           sync.Pool
	option          PotetoOption
	startupMutex    sync.RWMutex
	Server          http.Server
	Listener        net.Listener
	potetoWorkflows PotetoWorkflows
}

func Api(basePath string, handler LeafHandler) *poteto {
	api := New().(*poteto)

	api.Leaf(basePath, handler)

	return api
}

func New() Poteto {
	var DefaultPotetoOption PotetoOption
	if err := env.Parse(&DefaultPotetoOption); err != nil {
		panic(err)
	}

	return &poteto{
		router:          NewRouter(),
		errorHandler:    &httpErrorHandler{},
		middlewareTree:  NewMiddlewareTree(),
		option:          DefaultPotetoOption,
		potetoWorkflows: NewPotetoWorkflows(),
	}
}

func NewWithOption(option PotetoOption) Poteto {
	return &poteto{
		router:          NewRouter(),
		errorHandler:    &httpErrorHandler{},
		middlewareTree:  NewMiddlewareTree(),
		option:          option,
		potetoWorkflows: NewPotetoWorkflows(),
	}
}

func (p *poteto) Router() *router {
	return p.router.(*router)
}

func (p *poteto) MiddlewareTree() *middlewareTree {
	return p.middlewareTree.(*middlewareTree)
}

// Cashed context | NewContext
func (p *poteto) initializeContext(w http.ResponseWriter, r *http.Request) *context {
	if ctx, ok := p.cache.Get().(*context); ok {
		ctx.Reset(w, r)
		return ctx
	}

	newCtx := NewContext(w, r).(*context)
	if p.logger != nil {
		newCtx.SetLogger(p.logger)
	}
	return newCtx
}

func (p *poteto) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get from cache & reset context
	ctx := p.initializeContext(w, r)

	/* default get & set RequestId
	/  you can make WithRequestIdOption false: you can faster request
	/  option := poteto.PotetoOption{
	/    WithRequestId: false,
	/    ListenerNetwork: "tcp",
	/  }
	/  p := poteto.NewWithOption(option)
	*/
	if p.option.WithRequestId {
		reqId := ctx.RequestId()
		ctx.Set(constant.StoredRequestId, reqId)
		if id := ctx.GetRequestHeaderParam(constant.HeaderRequestId); id == "" {
			ctx.SetResponseHeader(constant.HeaderRequestId, reqId)
		}
	}

	routes := p.router.GetRoutesByMethod(r.Method)

	targetRoute, httpParams := routes.Search(r.URL.Path)
	if targetRoute == nil {
		ctx.WriteHeader(http.StatusNotFound)
		return
	}

	handler := targetRoute.GetHandler()
	if handler == nil {
		ctx.WriteHeader(http.StatusNotFound)
		return
	}

	ctx.SetQueryParam(r.URL.Query())
	ctx.SetPath(r.URL.Path)
	for _, httpParam := range httpParams {
		ctx.SetParam(constant.ParamTypePath, httpParam)
	}

	// Search middleware
	middlewares := p.middlewareTree.SearchMiddlewares(r.URL.Path)
	handler = p.applyMiddleware(middlewares, handler)
	if err := handler(ctx); err != nil {
		p.errorHandler.HandleHttpError(err, ctx)
	}

	// cached context
	p.cache.Put(ctx)
}

func (p *poteto) applyMiddleware(middlewares []MiddlewareFunc, handler HandlerFunc) HandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

func (p *poteto) Run(addr string) error {
	p.startupMutex.Lock()

	if !strings.Contains(addr, constant.ParamPrefix) {
		addr = constant.ParamPrefix + addr
	}

	p.Server.Addr = addr
	if err := p.setupServer(); err != nil {
		if p.option.DebugMode {
			utils.PotetoPrint(
				fmt.Sprintf(
					"poteto.setupServer reverted with %s",
					err.Error(),
				),
			)
		}
		p.startupMutex.Unlock()
		return err
	}

	// Run StartUpWorkflows just before the server starts
	workflows := p.potetoWorkflows.(*potetoWorkflows)
	if err := workflows.ApplyStartUpWorkflows(); err != nil {
		if p.option.DebugMode {
			utils.PotetoPrint(
				fmt.Sprintf(
					"workflows.ApplyStartUpWorkflows reverted with %s",
					err.Error(),
				),
			)
		}
		p.startupMutex.Unlock()
		return err
	}

	utils.PotetoPrint("server is available at http://127.0.0.1" + addr + "\n")

	p.startupMutex.Unlock()
	return p.Server.Serve(p.Listener)
}

// RunTLS required file byte not file path
func (p *poteto) RunTLS(addr string, cert, key []byte) error {
	p.startupMutex.Lock()

	// Setting TLS
	p.Server.TLSConfig = &tls.Config{MinVersion: 0x0303} // Version 1.2
	p.Server.TLSConfig.Certificates = make([]tls.Certificate, 1)
	parsedCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		if p.option.DebugMode {
			utils.PotetoPrint(
				fmt.Sprintf(
					"tls.X509KeyPair reverted with %s",
					err.Error(),
				),
			)
		}
		p.startupMutex.Unlock()
		return err
	}
	p.Server.TLSConfig.Certificates[0] = parsedCert

	// Setup Server
	p.Server.Addr = addr
	if err := p.setupServer(); err != nil {
		if p.option.DebugMode {
			utils.PotetoPrint(
				fmt.Sprintf(
					"poteto.setupServer reverted with %s",
					err.Error(),
				),
			)
		}
		p.startupMutex.Unlock()
		return err
	}

	// Run StartUpWorkflows just before the server starts
	workflows := p.potetoWorkflows.(*potetoWorkflows)
	if err := workflows.ApplyStartUpWorkflows(); err != nil {
		if p.option.DebugMode {
			utils.PotetoPrint(
				fmt.Sprintf(
					"workflows.ApplyStartUpWorkflows reverted with %s",
					err.Error(),
				),
			)
		}
		p.startupMutex.Unlock()
		return err
	}

	utils.PotetoPrint("server is available at https://127.0.0.1" + addr + "\n")

	p.startupMutex.Unlock()
	return p.Server.Serve(p.Listener)
}

func (p *poteto) setupServer() error {
	// Print Banner
	coloredBanner := color.HiGreenString(Banner)
	utils.PotetoPrint(coloredBanner)

	// setting handler
	p.Server.Handler = p

	// set listener
	if p.Listener == nil {
		ln, err := net.Listen(p.option.ListenerNetwork, p.Server.Addr)
		if err != nil {
			if p.option.DebugMode {
				utils.PotetoPrint(
					fmt.Sprintf(
						"net.Listen reverted with %s",
						err.Error(),
					),
				)
			}
			return err
		}

		if p.Server.TLSConfig == nil {
			p.Listener = ln
			return nil
		}

		// tls mode
		p.Listener = tls.NewListener(ln, p.Server.TLSConfig)
	}

	return nil
}

// Shutdown stops the server gracefully.
// It internally calls `http.Server#Shutdown()`.
func (p *poteto) Stop(ctx stdContext.Context) error {
	p.startupMutex.Lock()

	if err := p.Server.Shutdown(ctx); err != nil {
		if p.option.DebugMode {
			utils.PotetoPrint(
				fmt.Sprintf(
					"poteto.Server.Shutdown reverted with %s",
					err.Error(),
				),
			)
		}
		p.startupMutex.Unlock()
		return err
	}

	p.startupMutex.Unlock()
	return nil
}

func (p *poteto) Register(middlewares ...MiddlewareFunc) {
	p.middlewareTree.Insert("", middlewares...)
}

func (p *poteto) Combine(pattern string, middlewares ...MiddlewareFunc) *middlewareTree {
	return p.middlewareTree.Insert(pattern, middlewares...)
}

func (p *poteto) SetLogger(logger any) {
	p.logger = logger
}

func (p *poteto) RegisterWorkflow(workflowType string, priority uint, workflow WorkflowFunc) {
	p.potetoWorkflows.(*potetoWorkflows).RegisterWorkflow(workflowType, priority, workflow)
}

func (p *poteto) Leaf(basePath string, yield LeafHandler) {
	leaf := NewLeaf(p, basePath)

	yield(leaf)
}

var allHttpMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodHead,
	http.MethodOptions,
	http.MethodTrace,
	http.MethodConnect,
}

func (p *poteto) AddApi(api Poteto) {
	// add router
	for _, method := range allHttpMethods {
		linearRouter := api.Router().DFS(method)

		for _, lr := range linearRouter {
			err := p.router.add(method, lr.path, lr.handler)
			panic(err)
		}
	}

	// add middleware tree
	linearMiddleware := api.MiddlewareTree().DFS()
	for _, lm := range linearMiddleware {
		p.middlewareTree.Insert(lm.path, lm.handler)
	}
}

func (p *poteto) GET(path string, handler HandlerFunc) error {
	return p.router.GET(path, handler)
}

func (p *poteto) POST(path string, handler HandlerFunc) error {
	return p.router.POST(path, handler)
}

func (p *poteto) PATCH(path string, handler HandlerFunc) error {
	return p.router.PATCH(path, handler)
}

func (p *poteto) PUT(path string, handler HandlerFunc) error {
	return p.router.PUT(path, handler)
}

func (p *poteto) DELETE(path string, handler HandlerFunc) error {
	return p.router.DELETE(path, handler)
}

func (p *poteto) HEAD(path string, handler HandlerFunc) error {
	return p.router.HEAD(path, handler)
}

func (p *poteto) OPTIONS(path string, handler HandlerFunc) error {
	return p.router.OPTIONS(path, handler)
}

func (p *poteto) TRACE(path string, handler HandlerFunc) error {
	return p.router.TRACE(path, handler)
}

func (p *poteto) CONNECT(path string, handler HandlerFunc) error {
	return p.router.CONNECT(path, handler)
}

func (p *poteto) Play(method, path string, body ...string) *httptest.ResponseRecorder {
	if len(body) > 2 {
		panic("should be len(body) = 0 | 1")
	}

	resp, req := func() (*httptest.ResponseRecorder, *http.Request) {
		switch len(body) {
		case 1:
			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, path, strings.NewReader(body[0]))
			return w, req
		default:
			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, path, nil)
			return w, req
		}
	}()

	p.ServeHTTP(resp, req)

	return resp
}

func (p *poteto) Check(method, path string) bool {
	routes := p.router.GetRoutesByMethod(method)
	targetRoute, _ := routes.Search(path)
	if targetRoute == nil {
		return false
	}

	return targetRoute.GetHandler() != nil
}

func (p *poteto) Chain(middlewares ...MiddlewareFunc) func(HandlerFunc) HandlerFunc {
	return func(handler HandlerFunc) HandlerFunc {
		for _, middleware := range middlewares {
			handler = middleware(handler)
		}
		return handler
	}
}
