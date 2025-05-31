package poteto

import (
	"github.com/poteto-go/poteto/perror"
	"github.com/poteto-go/poteto/utils"
)

/*
/ Leaf Make Router Great
/ p.Leaf("/users", func(leaf Leaf) {
/   leaf.Register(sampleMiddleware)
/   leaf.GET("/", getAllUserForTest)
/   leaf.POST("/create", getAllUserForTest)
/   leaf.PUT("/change", getAllUserForTest)
/   leaf.DELETE("/delete", getAllUserForTest)
/ })
*/

type leaf struct {
	poteto   Poteto
	basePath string
}

type Leaf interface {
	// internal call Poteto.Combine w/ base path
	Register(middlewares ...MiddlewareFunc) *middlewareTree

	// internal call Poteto.GET w/ base path
	GET(addPath string, handler HandlerFunc) error

	// internal call Poteto.POST w/ base path
	POST(addPath string, handler HandlerFunc) error

	// internal call Poteto.PUT w/ base path
	PUT(addPath string, handler HandlerFunc) error

	// internal call Poteto.PATCH w/ base path
	PATCH(path string, handler HandlerFunc) error

	// internal call Poteto.DELETE w/ base path
	DELETE(addPath string, handler HandlerFunc) error

	// internal call Poteto.HEAD w/ base path
	HEAD(path string, handler HandlerFunc) error

	// internal call Poteto.OPTIONS w/ base path
	OPTIONS(path string, handler HandlerFunc) error

	// internal call Poteto.TRACE w/ base path
	TRACE(path string, handler HandlerFunc) error

	// internal call Poteto.CONNECT w/ base path
	CONNECT(path string, handler HandlerFunc) error
}

func NewLeaf(poteto Poteto, basePath string) Leaf {
	return &leaf{
		poteto:   poteto,
		basePath: basePath,
	}
}

func (l *leaf) Register(middlewares ...MiddlewareFunc) *middlewareTree {
	return l.poteto.Combine(l.basePath, middlewares...)
}

func (l *leaf) GET(addPath string, handler HandlerFunc) error {
	return leafAdd(l, GET{}, addPath, handler)
}

func (l *leaf) POST(addPath string, handler HandlerFunc) error {
	return leafAdd(l, POST{}, addPath, handler)
}

func (l *leaf) PUT(addPath string, handler HandlerFunc) error {
	return leafAdd(l, PUT{}, addPath, handler)
}

func (l *leaf) PATCH(addPath string, handler HandlerFunc) error {
	return leafAdd(l, PATCH{}, addPath, handler)
}

func (l *leaf) DELETE(addPath string, handler HandlerFunc) error {
	return leafAdd(l, DELETE{}, addPath, handler)
}

func (l *leaf) HEAD(addPath string, handler HandlerFunc) error {
	return leafAdd(l, HEAD{}, addPath, handler)
}

func (l *leaf) OPTIONS(addPath string, handler HandlerFunc) error {
	return leafAdd(l, OPTIONS{}, addPath, handler)
}

func (l *leaf) TRACE(addPath string, handler HandlerFunc) error {
	return leafAdd(l, TRACE{}, addPath, handler)
}

func (l *leaf) CONNECT(addPath string, handler HandlerFunc) error {
	return leafAdd(l, CONNECT{}, addPath, handler)
}

func leafAdd[M HTTPMethod](l *leaf, method M, addPath string, handler HandlerFunc) error {
	path, err := utils.BuildSafeUrl(l.basePath, addPath)
	if err != nil {
		return err
	}

	switch any(method).(type) {
	case GET:
		return l.poteto.GET(path, handler)
	case POST:
		return l.poteto.POST(path, handler)
	case PUT:
		return l.poteto.PUT(path, handler)
	case PATCH:
		return l.poteto.PATCH(path, handler)
	case DELETE:
		return l.poteto.DELETE(path, handler)
	case HEAD:
		return l.poteto.HEAD(path, handler)
	case OPTIONS:
		return l.poteto.OPTIONS(path, handler)
	case TRACE:
		return l.poteto.TRACE(path, handler)
	case CONNECT:
		return l.poteto.CONNECT(path, handler)
	default:
		// not run
		return perror.ErrUnSupportedHTTPMethod
	}
}
