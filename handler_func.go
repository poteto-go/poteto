package poteto

type WorkflowFunc func() error

type HandlerFunc func(ctx Context) error

type MiddlewareFunc func(next HandlerFunc) HandlerFunc

type LeafHandler func(leaf Leaf)
