package poteto

import (
	"net/http"

	"github.com/poteto-go/poteto/utils"
)

type Response interface {
	/*
		Write statusCode to http.ResponseWriter
	*/
	WriteHeader(code int)

	/*
		Write to http.ResponseWriter
	*/
	Write(b []byte) (int, error)

	/*
		Set status code
	*/
	SetStatus(code int)

	/*
		return http.ResponseWriter.Header()
	*/
	Header() http.Header

	/*
		Set Header if key doesn't include
	*/
	SetHeader(key, value string)

	/*
		Internal call http.ResponseWriter.Header().Add
	*/
	AddHeader(key, value string)

	/*
		fullfil interface for (making) responseController

		you can assign to responseController

		func hoge() {
			res := NewResponse(w)
			rc := http.NewResponseController(res)
		}

		https://go.dev/src/net/http/responsecontroller.go
	*/
	Unwrap() http.ResponseWriter

	/*
		reset response
	*/
	Reset(w http.ResponseWriter)
}

type response struct {
	Writer      http.ResponseWriter
	Status      int
	Size        int64
	IsCommitted bool
}

func NewResponse(w http.ResponseWriter) Response {
	return &response{Writer: w}
}

func (r *response) WriteHeader(code int) {
	if r.IsCommitted {
		utils.PotetoPrint("response has already committed\n")
		return
	}

	r.SetStatus(code)
	r.Writer.WriteHeader(r.Status)
	r.IsCommitted = true
}

func (r *response) SetHeader(key, value string) {
	if r.Writer.Header().Get(key) != "" {
		return
	}

	r.Writer.Header().Set(key, value)
}

func (r *response) AddHeader(key, value string) {
	r.Writer.Header().Add(key, value)
}

func (r *response) Write(b []byte) (int, error) {
	if !r.IsCommitted {
		if r.Status == 0 {
			r.SetStatus(http.StatusOK)
		}
		r.WriteHeader(r.Status)
		r.IsCommitted = true
	}

	n, err := r.Writer.Write(b)
	r.Size += int64(n)

	return n, err
}

func (r *response) SetStatus(code int) {
	r.Status = code
}

func (r *response) Header() http.Header {
	return r.Writer.Header()
}

func (r *response) Unwrap() http.ResponseWriter {
	return r.Writer
}

func (r *response) Reset(w http.ResponseWriter) {
	r.Writer = w
	r.Status = 0
	r.Size = 0
	r.IsCommitted = false
}
