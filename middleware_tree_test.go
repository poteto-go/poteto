package poteto

import "testing"

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
