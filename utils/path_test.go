package utils

import (
	"strings"
	"testing"

	"github.com/poteto-go/poteto/perror"
	"github.com/stretchr/testify/assert"
)

func TestBuildSafeUrl(t *testing.T) {
	t.Run("normal case", func(t *testing.T) {
		tests := []struct {
			basePath string
			addPath  string
			expected string
		}{
			{"/api", "/users", "/api/users"},
			{"/api/", "users", "/api/users"},
			{"/api", "", "/api"},
			{"/", "users", "/users"},
			{"/", "/", "/"},
			{"", "users", "/users"},
			{"", "", "/"},
		}

		for _, test := range tests {
			// Act
			result, err := BuildSafeUrl(test.basePath, test.addPath)

			// Assert
			assert.Nil(t, err)
			assert.Equal(t, test.expected, result)
		}
	})

	t.Run("error case", func(t *testing.T) {
		tests := []struct {
			basePath string
			addPath  string
			err      error
		}{
			{"..", "/hello", perror.ErrPathTraversalNotAllowed},
			{"/api", "..", perror.ErrPathTraversalNotAllowed},
			{
				strings.Repeat("s", 200),
				strings.Repeat("s", 200),
				perror.ErrPathLengthExceeded,
			},
		}

		for _, it := range tests {
			// Act
			_, err := BuildSafeUrl(it.basePath, it.addPath)

			// Assert
			assert.ErrorIs(t, err, it.err)
		}
	})
}
