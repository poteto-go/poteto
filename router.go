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

	GetRoutesByMethod(method string) *route
}

// Each Router has TrieTreeRouting by method
type router struct {
	routesGET     Route
	routesPOST    Route
	routesPUT     Route
	routesPATCH   Route
	routesDELETE  Route
	routesHEAD    Route
	routesOPTIONS Route
	routesTRACE   Route
	routesCONNECT Route
}

/*
Router Provides Radix-Tree Routing

O(logN) ~ N

Supports standard(net/http) methods GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, TRACE, CONNECT

You can use only Router of course.
*/
func NewRouter() Router {
	return &router{
		routesGET:     NewRoute(),
		routesPOST:    NewRoute(),
		routesPUT:     NewRoute(),
		routesPATCH:   NewRoute(),
		routesDELETE:  NewRoute(),
		routesHEAD:    NewRoute(),
		routesOPTIONS: NewRoute(),
		routesTRACE:   NewRoute(),
		routesCONNECT: NewRoute(),
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

func (r *router) GetRoutesByMethod(method string) *route {
	switch method {
	case http.MethodGet:
		return r.routesGET.(*route)
	case http.MethodPost:
		return r.routesPOST.(*route)
	case http.MethodPut:
		return r.routesPUT.(*route)
	case http.MethodPatch:
		return r.routesPATCH.(*route)
	case http.MethodDelete:
		return r.routesDELETE.(*route)
	case http.MethodHead:
		return r.routesHEAD.(*route)
	case http.MethodOptions:
		return r.routesOPTIONS.(*route)
	case http.MethodTrace:
		return r.routesTRACE.(*route)
	case http.MethodConnect:
		return r.routesCONNECT.(*route)
	default:
		return nil
	}
}
