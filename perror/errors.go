package perror

import "errors"

var (
	ErrZeroLengthContent  = errors.New("zero length content")
	ErrNotApplicationJson = errors.New("content-type is not application/json header")
)
