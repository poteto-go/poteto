package poteto

type WorkflowFunc func() error

type HandlerFunc func(ctx Context) error

type MiddlewareFunc func(next HandlerFunc) HandlerFunc

type LeafHandler func(leaf Leaf)

type ErrorHandlerFunc func(err error, ctx Context)

type (
	GET     struct{}
	POST    struct{}
	PUT     struct{}
	PATCH   struct{}
	DELETE  struct{}
	HEAD    struct{}
	OPTIONS struct{}
	TRACE   struct{}
	CONNECT struct{}
)

type HTTPMethod interface {
	GET | POST | PUT | PATCH | DELETE | HEAD | OPTIONS | TRACE | CONNECT
}
