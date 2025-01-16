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

func TestStrArrayToStr(t *testing.T) {
	tests := []struct {
		name     string
		targets  []string
		expected string
	}{
		{"Test 1 len array", []string{"hello"}, "hello"},
		{"Test multi len array", []string{"hello", "world", "!!"}, "hello,world,!!"},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			result := StrArrayToStr(it.targets)

			if result != it.expected {
				t.Errorf("Unmatched")
			}
		})
	}
}

func BenchmarkStrArrayToStr(b *testing.B) {
	input := []string{}
	for i := 0; i < 100; i++ {
		input = append(input, "hello")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		output := StrArrayToStr(input)
		_ = output
	}
}
