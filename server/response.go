package server

import (
	"encoding/json"
	"fmt"
	"io"
)

type ResponseBuilder struct {
	status int
	body   any
}

func NewResponse() *ResponseBuilder {
	return &ResponseBuilder{}
}

func (rb *ResponseBuilder) Status(status int) *ResponseBuilder {
	rb.status = status
	return rb
}

func (rb *ResponseBuilder) Body(body any) *ResponseBuilder {
	rb.body = body
	return rb
}

func (rb *ResponseBuilder) AsTextPlain() (int, string, func(io.Writer)) {
	return rb.status, "text/plain", rb.writer()
}

func (rb *ResponseBuilder) AsJson() (int, string, func(io.Writer)) {
	return rb.status, "application/json", rb.writer()
}

func (rb *ResponseBuilder) AsHtml() (int, string, func(io.Writer)) {
	return rb.status, "text/html", rb.writer()
}

func (rb *ResponseBuilder) As(contentType string) (int, string, func(io.Writer)) {
	return rb.status, contentType, rb.writer()
}

func (rb *ResponseBuilder) writer() func(io.Writer) {
	if rb.body == nil {
		return func(io.Writer) {

		}
	}
	switch b := rb.body.(type) {
	case string:
		return StringWriter(b)
	case error:
		return ErrWriter(b)
	case io.Reader:
		return ReaderWriter(b)
	case func(io.Writer):
		return b
	default:
		return JsonWriter(b)
	}
}

func Ok(body any) *ResponseBuilder {
	return NewResponse().Status(200).Body(body)
}

func ErrNotFound(body any) *ResponseBuilder {
	return NewResponse().Status(404).Body(body)
}

func ErrBadRequest(body any) *ResponseBuilder {
	return NewResponse().Status(400).Body(body)
}

func ErrInternal(body any) *ResponseBuilder {
	return NewResponse().Status(500).Body(body)
}

func StringWriter(str string) func(writer io.Writer) {
	return func(writer io.Writer) {
		fmt.Fprint(writer, str)
	}
}

func ErrWriter(err error) func(writer io.Writer) {
	return func(writer io.Writer) {
		if err != nil {
			fmt.Fprintf(writer, "error: %s", err.Error())
		}
	}
}

func JsonWriter(data any) func(writer io.Writer) {
	return func(writer io.Writer) {
		enc := json.NewEncoder(writer)
		enc.SetIndent("", "  ")
		enc.Encode(data)
	}
}

func ReaderWriter(data io.Reader) func(writer io.Writer) {
	return func(writer io.Writer) {
		io.Copy(writer, data)
	}
}
