package perror

import "errors"

var (
	ErrZeroLengthContent       = errors.New("zero length content")
	ErrNotApplicationJson      = errors.New("content-type is not application/json header")
	ErrPathTraversalNotAllowed = errors.New("path traversal not allowed")
	ErrPathLengthExceeded      = errors.New("path length exceeded")
	ErrUnSupportedHTTPMethod   = errors.New("unsupported http method")
)
