package poteto

import (
	"errors"
	"net/http"
	"testing"
)

func TestError(t *testing.T) {
	tests := []struct {
		name           string
		isHaveInternal bool
		expected       string
	}{
		{"Test w/o internal", false, "code=400, message=Bad Request"},
		{"Test w internal", true, "code=400, message=Bad Request, internal=internalError"},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			httpError := NewHttpError(http.StatusBadRequest)
			if it.isHaveInternal {
				httpError.SetInternalError(errors.New("internalError"))
			}

			result := httpError.Error()

			if result != it.expected {
				t.Errorf("Unmatched")
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	internalError := errors.New("internalError")

	httpError := NewHttpError(http.StatusBadRequest, "BadRequest")
	httpError.SetInternalError(internalError)

	result := httpError.Unwrap()
	if result != internalError {
		t.Errorf("Unmatched")
	}
}
