package poteto

import (
	"net/http"
	"testing"
)

var rtr *router

func TestAdd(t *testing.T) {
	rtr = NewRouter().(*router)
	tests := []struct {
		name   string
		method string
		path   string
		want   bool
	}{
		{"success add new route", http.MethodGet, "/users/find", false},
		{"fail add already existed route", http.MethodGet, "/users/find", true},
		{"success add new method already existed route", http.MethodPost, "/users/find", false},
		{"success add new method already existed route", http.MethodPut, "/users/find", false},
		{"success add new method already existed route", http.MethodDelete, "/users/find", false},
		{"success add new method already existed route", http.MethodPatch, "/users/find", false},
		{"success add new method already existed route", http.MethodHead, "/users/find", false},
		{"success add new method already existed route", http.MethodOptions, "/users/find", false},
		{"success add new method already existed route", http.MethodTrace, "/users/find", false},
		{"success add new method already existed route", http.MethodConnect, "/users/find", false},
		{"return nil unexpected method", "UNEXPECTED", "/users/find", true},
	}

	for _, it := range tests {
		t.Run(it.name, func(tt *testing.T) {
			err := rtr.add(it.method, it.path, nil)
			if it.want {
				if err == nil {
					t.Errorf("FATAL: success already existed route")
				}
			} else {
				if err != nil {
					t.Errorf("FATAL: fail new route")
				}
			}
		})
	}
}

func TestGetRoutesByMethod(t *testing.T) {
	rtr.GET("/users/get", nil)

	routes := rtr.GetRoutesByMethod("GET")
	child, ok := routes.children["users"].(*route)
	if !ok || child.key != "users" {
		t.Errorf("FATAL add top param")
	}

	cchild, ok := child.children["get"].(*route)
	if !ok || cchild.key != "get" {
		t.Errorf("FATAL add bottom param")
	}
}

func BenchmarkInsertAndSearchRouter(b *testing.B) {
	urls := []string{
		"/",
		"/example.com/v1/users/find/poteto",
		"/example.com/v1/users/find/potato",
		"/example.com/v1/users/find/jagaimo",
		"/example.com/v1/users/create/poteto",
		"/example.com/v1/users/create/potato",
		"/example.com/v1/users/create/jagaimo",
		"/example.com/v1/members/find/poteto",
		"/example.com/v1/members/find/potato",
		"/example.com/v1/members/find/jagaimo",
		"/example.com/v1/members/create/poteto",
		"/example.com/v1/members/create/potato",
		"/example.com/v1/members/create/jagaimo",
		"/example.com/v2/users/find/poteto",
		"/example.com/v2/users/find/potato",
		"/example.com/v2/users/find/jagaimo",
		"/example.com/v2/users/create/poteto",
		"/example.com/v2/users/create/potato",
		"/example.com/v2/users/create/jagaimo",
		"/example.com/v2/members/find/poteto",
		"/example.com/v2/members/find/potato",
		"/example.com/v2/members/find/jagaimo",
		"/example.com/v2/members/create/poteto",
		"/example.com/v2/members/create/potato",
		"/example.com/v2/members/create/jagaimo",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		router := NewRouter()
		// Insert
		for _, url := range urls {
			router.GET(url, nil)
			router.POST(url, nil)
			router.PUT(url, nil)
			router.PATCH(url, nil)
			router.DELETE(url, nil)
			router.HEAD(url, nil)
			router.OPTIONS(url, nil)
			router.TRACE(url, nil)
			router.CONNECT(url, nil)
		}

		// Search
		for _, url := range urls {
			routesGET := router.GetRoutesByMethod(http.MethodGet)
			routesGET.Search(url)
			routesPOST := router.GetRoutesByMethod(http.MethodPost)
			routesPOST.Search(url)
			routesPUT := router.GetRoutesByMethod(http.MethodPut)
			routesPUT.Search(url)
			routesPATCH := router.GetRoutesByMethod(http.MethodPatch)
			routesPATCH.Search(url)
			routesDELETE := router.GetRoutesByMethod(http.MethodDelete)
			routesDELETE.Search(url)
			routesHEAD := router.GetRoutesByMethod(http.MethodHead)
			routesHEAD.Search(url)
			routesOPTIONS := router.GetRoutesByMethod(http.MethodOptions)
			routesOPTIONS.Search(url)
			routesTRACE := router.GetRoutesByMethod(http.MethodTrace)
			routesTRACE.Search(url)
			routesCONNECT := router.GetRoutesByMethod(http.MethodConnect)
			routesCONNECT.Search(url)
		}
	}
}
