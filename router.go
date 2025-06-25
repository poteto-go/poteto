package poteto

import (
	"errors"
	"net/http"
	"strings"
)

type Router interface {
	add(method, path string, handler HandlerFunc) error

	/*
		Register GET method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	GET(path string, handler HandlerFunc) error

	/*
		Register POST method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	POST(path string, handler HandlerFunc) error

	/*
		Register PUT method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	PUT(path string, handler HandlerFunc) error

	/*
		Register PATCH method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	PATCH(path string, handler HandlerFunc) error

	/*
		Register DELETE method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	DELETE(path string, handler HandlerFunc) error

	/*
		Register HEAD method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	HEAD(path string, handler HandlerFunc) error

	/*
		Register OPTIONS method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	OPTIONS(path string, handler HandlerFunc) error

	/*
		Register TRACE method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	TRACE(path string, handler HandlerFunc) error

	/*
		Register CONNECT method Route

		Trim Suffix "/"
		EX: "/users/" -> "/users"
	*/
	CONNECT(path string, handler HandlerFunc) error

	// DFS route & return linearRouter by method
	//
	// []{
	//   path: string,
	//   handler: HandlerFunc,
	// }
	DFS(method string) []routeLinear

	GetRoutesByMethod(method string) *route
}

// Each Router has TrieTreeRouting by method
type router struct {
	routes map[string]Route
}

/*
Router Provides Radix-Tree Routing

O(logN) ~ N

Supports standard(net/http) methods GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, TRACE, CONNECT

You can use only Router of course.
*/
func NewRouter() Router {
	return &router{
		routes: map[string]Route{
			http.MethodGet:     NewRoute(),
			http.MethodPost:    NewRoute(),
			http.MethodPut:     NewRoute(),
			http.MethodPatch:   NewRoute(),
			http.MethodDelete:  NewRoute(),
			http.MethodHead:    NewRoute(),
			http.MethodOptions: NewRoute(),
			http.MethodTrace:   NewRoute(),
			http.MethodConnect: NewRoute(),
		},
	}
}

func (r *router) add(method, path string, handler HandlerFunc) error {
	routes := r.GetRoutesByMethod(method)
	if routes == nil {
		return errors.New("unexpected method error: " + method)
	}

	thisRoute, _ := routes.Search(path)
	if thisRoute != nil {
		if path == "/" {
			thisRoute.handler = handler
			return nil
		}
		return errors.New("[" + method + "] " + path + " is already used")
	}

	// "/users/" -> "/users"
	// if just "/" -> handler set by above
	path = strings.TrimSuffix(path, "/")

	routes.Insert(path, handler)
	return nil
}

// These are router Method
// Seems redundant, but you can register your own router with poteto
// And call it with `Poteto.GET()` etc.
func (r *router) GET(path string, handler HandlerFunc) error {
	return r.add(http.MethodGet, path, handler)
}

func (r *router) POST(path string, handler HandlerFunc) error {
	return r.add(http.MethodPost, path, handler)
}

func (r *router) PUT(path string, handler HandlerFunc) error {
	return r.add(http.MethodPut, path, handler)
}

func (r *router) PATCH(path string, handler HandlerFunc) error {
	return r.add(http.MethodPatch, path, handler)
}

func (r *router) DELETE(path string, handler HandlerFunc) error {
	return r.add(http.MethodDelete, path, handler)
}

func (r *router) HEAD(path string, handler HandlerFunc) error {
	return r.add(http.MethodHead, path, handler)
}

func (r *router) OPTIONS(path string, handler HandlerFunc) error {
	return r.add(http.MethodOptions, path, handler)
}

func (r *router) TRACE(path string, handler HandlerFunc) error {
	return r.add(http.MethodTrace, path, handler)
}

func (r *router) CONNECT(path string, handler HandlerFunc) error {
	return r.add(http.MethodConnect, path, handler)
}

func (r *router) DFS(method string) []routeLinear {
	routes := r.GetRoutesByMethod(method)
	if routes == nil {
		return []routeLinear{}
	}

	return routes.DFS()
}

func (r *router) GetRoutesByMethod(method string) *route {
	if routes, ok := r.routes[method]; ok {
		return routes.(*route)
	}
	return nil
}
