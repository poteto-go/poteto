package poteto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertAndSearchMiddlewares(t *testing.T) {
	mg := NewMiddlewareTree()

	mg.Insert("/users", sampleMiddleware)
	mg.Insert("/users/hello", sampleMiddleware2)
	tests := []struct {
		name     string
		target   string
		expected int
	}{
		{"Test middlewares", "/users", 1},
		{"Test not found middlewares", "/test", 0},
		{"Test found two node", "/users/hello", 2},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			middlewares := mg.SearchMiddlewares(it.target)
			if len(middlewares) != it.expected {
				t.Errorf("Unmatched")
			}
		})
	}
}

func TestMiddlewareTree_DFS(t *testing.T) {
	t.Run("empty tree", func(t *testing.T) {
		// Arrange
		tree := NewMiddlewareTree()

		// Act
		results := tree.DFS()

		// Assert
		assert.Equal(t, 0, len(results))
	})

	t.Run("normal case", func(t *testing.T) {
		// Arrange
		mockFunc := func(next HandlerFunc) HandlerFunc {
			return func(ctx Context) error {
				return next(ctx)
			}
		}

		tree := NewMiddlewareTree()
		tree.Insert("/users", mockFunc)
		tree.Insert("/users/hello", mockFunc, mockFunc)
		tree.Insert("/users/hello/world", nil)

		// Act
		results := tree.DFS()

		// Assert
		assert.Equal(t, 3, len(results))
	})
}
