package poteto

import (
	"net/http"
)

// This is defaultErrorHandler
func DefaultErrorHandler(err error, ctx Context) {
	if ctx.GetResponse().IsCommitted {
		return
	}

	httpErr, ok := err.(*httpError)
	if !ok { // Not Handled
		httpErr = NewHttpError(http.StatusInternalServerError)
	}
	// Unwrap wrapped error
	if httpErr.InternalError != nil {
		if warpedErr, ok := httpErr.InternalError.(*httpError); ok {
			httpErr = warpedErr
		}
	}

	message := httpErr.Message
	switch m := httpErr.Message.(type) {
	case string:
		message = map[string]string{"message": m}
	case []byte:
		message = map[string][]byte{"message": m}
	}

	// Send response
	ctx.JSON(httpErr.Code, message)
}
