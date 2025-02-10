package poteto

import (
	"strings"

	"github.com/poteto-go/poteto/constant"
)

type Binder interface {
	Bind(ctx Context, object any) error
}

type binder struct{}

func NewBinder() Binder {
	return &binder{}
}

func (*binder) Bind(ctx Context, object any) error {
	req := ctx.GetRequest()
	if req.ContentLength == 0 {
		return nil
	}

	base, _, _ := strings.Cut(
		ctx.GetRequestHeaderParam(constant.HEADER_CONTENT_TYPE), ";",
	)
	mediaType := strings.TrimSpace(base)

	switch mediaType {
	case constant.APPLICATION_JSON:
		if err := ctx.JsonDeserialize(object); err != nil {
			return err
		}
	}

	// if not application/json
	// return nil
	return nil
}
