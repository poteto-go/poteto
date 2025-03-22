package poteto

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
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
	if !ok {
		t.Errorf("FATAL add top param")
	}

	_, ok = child.children["get"].(*route)
	if !ok {
		t.Errorf("FATAL add bottom param")
	}
}

func TestRouter_DFS(t *testing.T) {
	t.Run("return empty array", func(t *testing.T) {
		rtr = NewRouter().(*router)
		got := rtr.DFS(http.MethodGet)
		assert.Equal(t, 0, len(got))
	})

	t.Run("call route dfs", func(t *testing.T) {
		rtr = NewRouter().(*router)
		rtr.GET("/users/get", func(ctx Context) error {
			return ctx.JSON(200, nil)
		})
		got := rtr.DFS(http.MethodGet)
		assert.Equal(t, 1, len(got))
	})
}
