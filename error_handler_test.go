package poteto

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultErrorHandler(t *testing.T) {
	t.Run("handled case:", func(t *testing.T) {
		tests := []struct {
			name         string
			err          error
			expectedCode int
			expected     string
		}{
			{
				"Test Not Handled Error -> Server Error",
				errors.New("not httpError"),
				http.StatusInternalServerError,
				`{"message":"Internal Server Error"}`,
			},
			{
				"Test Handled Error",
				NewHttpError(http.StatusBadRequest),
				http.StatusBadRequest,
				`{"message":"Bad Request"}`,
			},
			{
				"Test wrapped Error",
				&httpError{
					Code:          http.StatusBadRequest,
					Message:       "",
					InternalError: NewHttpError(http.StatusBadRequest),
				},
				http.StatusBadRequest,
				`{"message":"Bad Request"}`,
			},
		}

		for _, it := range tests {
			t.Run(it.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				ctx := NewContext(w, nil)

				DefaultErrorHandler(it.err, ctx)

				assert.Equal(t, it.expectedCode, w.Result().StatusCode)
				assert.Contains(t, w.Body.String(), it.expected)
			})
		}
	})

	t.Run("has already committed => return 200", func(t *testing.T) {
		// Arrange
		w := httptest.NewRecorder()
		ctx := NewContext(w, nil)
		ctx.GetResponse().IsCommitted = true

		// Act
		DefaultErrorHandler(NewHttpError(http.StatusBadRequest), ctx)

		// Assert
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})
}
