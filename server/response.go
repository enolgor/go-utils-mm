package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ResponseBuilder interface {
	Status(status int) ResponseBuilder
	WithBody(body any) ResponseBuilder
	WithHeader(key, value string) ResponseBuilder
	WithCookie(cookie *http.Cookie) ResponseBuilder
	Redirect(redirect string)
	As(contentType string)
	AsTextPlain()
	AsJson()
	AsHtml()
}

type responseBuilder struct {
	w      http.ResponseWriter
	status int
	body   any
}

func Response(w http.ResponseWriter) ResponseBuilder {
	return &responseBuilder{w: w, status: http.StatusOK, body: nil}
}

func (rb *responseBuilder) Status(status int) ResponseBuilder {
	rb.status = status
	return rb
}

func (rb *responseBuilder) WithBody(body any) ResponseBuilder {
	rb.body = body
	return rb
}

func (rb *responseBuilder) WithHeader(key, value string) ResponseBuilder {
	rb.w.Header().Add(key, value)
	return rb
}

func (rb *responseBuilder) WithCookie(cookie *http.Cookie) ResponseBuilder {
	http.SetCookie(rb.w, cookie)
	return rb
}

func (rb *responseBuilder) writeBody() {
	if rb.body == nil {
		return
	}
	switch b := rb.body.(type) {
	case string:
		fmt.Fprint(rb.w, b)
	case error:
		fmt.Fprintf(rb.w, "error: %s", b.Error())
	case io.Reader:
		io.Copy(rb.w, b)
	case func(io.Writer):
		b(rb.w)
	default:
		enc := json.NewEncoder(rb.w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(b); err != nil {
			panic(err)
		}
	}
}

func (rb *responseBuilder) As(contentType string) {
	rb.WithHeader("Content-Type", contentType)
	rb.w.WriteHeader(rb.status)
	rb.writeBody()
}

func (rb *responseBuilder) AsTextPlain() {
	rb.As("text/plain")
}

func (rb *responseBuilder) AsJson() {
	rb.As("application/json")
}

func (rb *responseBuilder) AsHtml() {
	rb.As("text/html")
}

func (rb *responseBuilder) Redirect(redirect string) {
	rb.WithBody(fmt.Sprintf(`<html><header><script>window.location.replace("%s");</script></header><body></body></html>`, redirect)).AsHtml()
}
