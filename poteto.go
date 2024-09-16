package poteto

import (
	"net/http"
)

type Poteto interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Run(addr string)
	GET(path string, handler HandlerFunc) error
	POST(path string, handler HandlerFunc) error
	PUT(path string, handler HandlerFunc) error
	DELETE(path string, handler HandlerFunc) error
}

type poteto struct {
	router Router
}

func New() Poteto {
	return &poteto{
		router: NewRouter([]string{"GET", "POST", "PUT", "DELETE"}),
	}
}

func (p *poteto) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(w, r)
	routes := p.router.GetRoutesByMethod(r.Method)

	targetRoute := routes.Search(r.URL.Path)
	handler := targetRoute.GetHandler()

	if targetRoute == nil || handler == nil {
		ctx.WriteHeader(http.StatusNotFound)
		return
	}

	ctx.SetPath(r.URL.Path)
	handler(ctx)
}

func (p *poteto) Run(addr string) {
	if err := http.ListenAndServe(addr, p); err != nil {
		panic(err)
	}
}

func (p *poteto) GET(path string, handler HandlerFunc) error {
	return p.router.GET(path, handler)
}

func (p *poteto) POST(path string, handler HandlerFunc) error {
	return p.router.POST(path, handler)
}

func (p *poteto) PUT(path string, handler HandlerFunc) error {
	return p.router.PUT(path, handler)
}

func (p *poteto) DELETE(path string, handler HandlerFunc) error {
	return p.router.DELETE(path, handler)
}
