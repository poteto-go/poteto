package poteto

import (
	"testing"
)

func TestNewRoute(t *testing.T) {
	got := NewRoute().(*route)

	if len(got.children) != 0 {
		t.Errorf("Cannot initialize Route: method")
	}
}

func TestInsertAlreadyExistPath(t *testing.T) {
	route := NewRoute().(*route)

	route.Insert("/", getAllUserForTest)
	route.Insert("/", getAllUserForTest)
}

func TestInsertAndSearch(t *testing.T) {
	url := "/example.com/v1/users/find/poteto"

	route := NewRoute().(*route)

	route.Insert("/", nil)
	route.Insert(url, nil)
	route.Insert("/users/:id", nil)
	route.Insert("/users/:id/name", nil)

	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"FIND empty", "/", ""},
		{"FIND", "/example.com", "example.com"},
		{"NOT FOUND", "/test.com", ""},
		{"PARAM ROUTING", "/users/1", ":id"},
		{"PARAM ROUTING", "/users/1/name", "name"},
	}

	for _, it := range tests {
		t.Run(it.name, func(tt *testing.T) {
			route.Search(it.arg)
		})
	}
}

func BenchmarkInsertAndSearch(b *testing.B) {
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
		// Insert
		route := NewRoute().(*route)
		for _, url := range urls {
			route.Insert(url, nil)
		}

		// Search
		for _, url := range urls {
			route.Search(url)
		}
	}
}
