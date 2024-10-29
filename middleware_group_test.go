package poteto

import "testing"

func TestInsertAndSearchMG(t *testing.T) {
	mg := middlewareGroup{
		children: make(map[string]MiddlewareGroup),
	}

	mg.Insert("/users", nil)

	tests := []struct {
		name     string
		target   string
		expected string
	}{
		{"Test found case", "/users", "users"},
		{"Test not found case", "/test", ""},
		{"Test found onetime", "/users/hello", "users"},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			sg := mg.Search(it.target)
			if sg.key != it.expected {
				t.Errorf("Unmatched")
			}
		})
	}
}
