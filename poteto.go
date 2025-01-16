package poteto

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"sync"

	stdContext "context"

	"github.com/fatih/color"
	"github.com/poteto-go/poteto/constant"
	"github.com/poteto-go/poteto/utils"
)

type Poteto interface {
	// If requested, call this
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Run(addr string) error
	RunTLS(addr string, cert, key []byte) error
	Stop(ctx stdContext.Context) error
	setupServer() error
	Register(middlewares ...MiddlewareFunc)
	Combine(pattern string, middlewares ...MiddlewareFunc) *middlewareTree
	SetLogger(logger any)
	Leaf(basePath string, handler LeafHandler)

	GET(path string, handler HandlerFunc) error
	POST(path string, handler HandlerFunc) error
	PUT(path string, handler HandlerFunc) error
	PATCH(path string, handler HandlerFunc) error
	DELETE(path string, handler HandlerFunc) error
	HEAD(path string, handler HandlerFunc) error
	OPTIONS(path string, handler HandlerFunc) error
	TRACE(path string, handler HandlerFunc) error
	CONNECT(path string, handler HandlerFunc) error
}

type poteto struct {
	router         Router
	errorHandler   HttpErrorHandler
	middlewareTree MiddlewareTree
	logger         any
	cache          sync.Pool
	option         PotetoOption
	startupMutex   sync.RWMutex
	Server         http.Server
	Listener       net.Listener
}

func New() Poteto {
	return &poteto{
		router:         NewRouter(),
		errorHandler:   &httpErrorHandler{},
		middlewareTree: NewMiddlewareTree(),
		option:         DefaultPotetoOption,
	}
}

func NewWithOption(option PotetoOption) Poteto {
	return &poteto{
		router:         NewRouter(),
		errorHandler:   &httpErrorHandler{},
		middlewareTree: NewMiddlewareTree(),
		option:         option,
	}
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
		ctx.Set(constant.STORE_REQUEST_ID, reqId)
		if id := ctx.GetRequestHeaderParam(constant.HEADER_X_REQUEST_ID); id == "" {
			ctx.SetResponseHeader(constant.HEADER_X_REQUEST_ID, reqId)
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
		ctx.SetParam(constant.PARAM_TYPE_PATH, httpParam)
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

	if !strings.Contains(addr, constant.PARAM_PREFIX) {
		addr = constant.PARAM_PREFIX + addr
	}

	p.Server.Addr = addr
	if err := p.setupServer(); err != nil {
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
	p.Server.TLSConfig = &tls.Config{}
	p.Server.TLSConfig.Certificates = make([]tls.Certificate, 1)
	parsedCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		p.startupMutex.Unlock()
		return err
	}
	p.Server.TLSConfig.Certificates[0] = parsedCert

	// Setup Server
	p.Server.Addr = addr
	if err := p.setupServer(); err != nil {
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

// Leaf makes router group
// You can make your router clear
// with middlewares
func (p *poteto) Leaf(basePath string, handler LeafHandler) {
	leaf := NewLeaf(p, basePath)

	handler(leaf)
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
