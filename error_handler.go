package poteto

import (
	"net/http"
)

type HttpErrorHandler interface {
	HandleHttpError(err error, ctx Context)
}

type httpErrorHandler struct{}

func (heh *httpErrorHandler) HandleHttpError(err error, ctx Context) {
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
	err = ctx.JSON(httpErr.Code, message)
	_ = err
}
