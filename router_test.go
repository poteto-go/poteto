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
	rtr.GET("users/get", nil)

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
		"https://example.com/v1/users/find/poteto",
		"https://example.com/v1/users/find/potato",
		"https://example.com/v1/users/find/jagaimo",
		"https://example.com/v1/users/create/poteto",
		"https://example.com/v1/users/create/potato",
		"https://example.com/v1/users/create/jagaimo",
		"https://example.com/v1/members/find/poteto",
		"https://example.com/v1/members/find/potato",
		"https://example.com/v1/members/find/jagaimo",
		"https://example.com/v1/members/create/poteto",
		"https://example.com/v1/members/create/potato",
		"https://example.com/v1/members/create/jagaimo",
		"https://example.com/v2/users/find/poteto",
		"https://example.com/v2/users/find/potato",
		"https://example.com/v2/users/find/jagaimo",
		"https://example.com/v2/users/create/poteto",
		"https://example.com/v2/users/create/potato",
		"https://example.com/v2/users/create/jagaimo",
		"https://example.com/v2/members/find/poteto",
		"https://example.com/v2/members/find/potato",
		"https://example.com/v2/members/find/jagaimo",
		"https://example.com/v2/members/create/poteto",
		"https://example.com/v2/members/create/potato",
		"https://example.com/v2/members/create/jagaimo",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		router := NewRouter()
		// Insert
		for _, url := range urls {
			router.GET(url, nil)
			router.POST(url, nil)
			router.PUT(url, nil)
			router.DELETE(url, nil)
		}

		// Search
		for _, url := range urls {
			routesGET := router.GetRoutesByMethod(http.MethodGet)
			routesGET.Search(url)
			routesPOST := router.GetRoutesByMethod(http.MethodPost)
			routesPOST.Search(url)
			routesPUT := router.GetRoutesByMethod(http.MethodPut)
			routesPUT.Search(url)
			routesDELETE := router.GetRoutesByMethod(http.MethodDelete)
			routesDELETE.Search(url)
		}
	}
}
