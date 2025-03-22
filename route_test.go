package poteto

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestRoute_DFS(t *testing.T) {
	// Arrange
	mockFunc := func(ctx Context) error {
		return ctx.JSON(200, nil)
	}
	route := NewRoute().(*route)
	route.Insert("/", mockFunc)
	route.Insert("/users/greet/nil", nil)
	route.Insert("/users/:id", mockFunc)
	route.Insert("/users/:id/name", mockFunc)
	route.Insert("/users/social/followers", mockFunc)

	// Act
	results := route.DFS()

	// Assert
	assert.Equal(t, 4, len(results))
}
