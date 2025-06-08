package utils

import "testing"

func TestIsSliceEqual(t *testing.T) {
	tests := []struct {
		name     string
		vec1     []any
		vec2     []any
		expected bool
	}{
		{"TEST same int array", []any{1, 2, 3}, []any{1, 2, 3}, true},
		{"TEST not same length array", []any{1, 2, 3}, []any{1, 2}, false},
		{"TEST same string array", []any{"hello", "world"}, []any{"hello", "world"}, true},
		{"TEST not same value", []any{"hello", "world"}, []any{"not", "world"}, false},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			result := IsSliceEqual(it.vec1, it.vec2)
			if result != it.expected {
				t.Errorf("FATAL")
			}
		})
	}
}

func TestBuildString(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected string
	}{
		{
			name:     "Empty slice",
			values:   []string{},
			expected: "",
		},
		{
			name:     "Single value",
			values:   []string{"hello"},
			expected: "hello",
		},
		{
			name:     "Multiple values",
			values:   []string{"hello", "world", "!!"},
			expected: "hello, world, !!",
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			result := BuildString(it.values...)
			if result != it.expected {
				t.Errorf("Expected %s, got %s", it.expected, result)
			}
		})
	}
}
