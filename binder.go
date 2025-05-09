package poteto

import (
	"strings"

	validator "github.com/go-playground/validator/v10"
	"github.com/poteto-go/poteto/constant"
	"github.com/poteto-go/poteto/perror"
)

type Binder interface {
	// Bind request body -> &object
	//
	// if unset "Content-Type: application/json", return perror.ErrUnsetHeaderApplicationJson
	//
	// if zero length content, return perror.ErrZeroLengthContent
	//
	Bind(ctx Context, object any) error

	// Bind with github.com/go-playground/validator/v10
	BindWithValidate(ctx Context, object any) error
}

type binder struct{}

func NewBinder() Binder {
	return &binder{}
}

func (b *binder) Bind(ctx Context, object any) error {
	req := ctx.GetRequest()
	if req.ContentLength == 0 {
		return perror.ErrZeroLengthContent
	}

	base, _, _ := strings.Cut(
		ctx.GetRequestHeaderParam(constant.HeaderContentType), ";",
	)
	mediaType := strings.TrimSpace(base)

	if mediaType != constant.APPLICATION_JSON {
		// if not application/json
		return perror.ErrNotApplicationJson
	}

	if err := ctx.JsonDeserialize(object); err != nil {
		return err
	}
	return nil
}

func (b *binder) BindWithValidate(ctx Context, object any) error {
	if err := b.Bind(ctx, object); err != nil {
		return err
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(object); err != nil {
		return err
	}

	return nil
}
