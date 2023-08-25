package ht

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type QueryHandler func(*url.Values) (int, string, func(io.Writer))

type BodyHandler func(*url.Values, io.Reader) (int, string, func(io.Writer))

type RequestHandler func(req *http.Request) (int, string, func(io.Writer))

type Handler interface {
	QueryHandler | BodyHandler | RequestHandler
}

type ChainHandler func(w http.ResponseWriter, req *http.Request) bool

func (ch ChainHandler) Chain(next ChainHandler) ChainHandler {
	return func(w http.ResponseWriter, req *http.Request) bool {
		if ch(w, req) {
			return next(w, req)
		}
		return false
	}
}

func (ch ChainHandler) Finally(last func(w http.ResponseWriter, req *http.Request)) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if ch(w, req) {
			last(w, req)
		}
	}
}

func BasicAuthHandler(user, pass string) ChainHandler {
	b64auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
	return func(w http.ResponseWriter, req *http.Request) bool {
		if b64auth != req.Header.Get("Authorization") {
			fmt.Println(b64auth)
			fmt.Println(req.Header.Get("Authorization"))
			w.Header().Add("Content-Type", "text/plain")
			w.Header().Add("WWW-Authenticate", `Basic realm="Realm"`)
			w.WriteHeader(401)
			fmt.Fprint(w, "unauthorized")
			return false
		}
		return true
	}
}

func Handle[H Handler](h H) func(w http.ResponseWriter, req *http.Request) {
	switch v := any(h).(type) {
	case QueryHandler:
		return serveQuery(v)
	case BodyHandler:
		return serveBody(v)
	case RequestHandler:
		return serveRequest(v)
	}
	panic("unsupported handler type")
}

func serveQuery(f QueryHandler) func(w http.ResponseWriter, req *http.Request) {
	return serveRequest(func(req *http.Request) (int, string, func(io.Writer)) {
		query := req.URL.Query()
		return f(&query)
	})
}

func serveBody(f BodyHandler) func(w http.ResponseWriter, req *http.Request) {
	return serveRequest(func(req *http.Request) (int, string, func(io.Writer)) {
		query := req.URL.Query()
		return f(&query, req.Body)
	})
}

func serveRequest(f func(req *http.Request) (int, string, func(io.Writer))) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var status int
		var ctype string
		var writer func(io.Writer)
		status, ctype, writer = f(req)
		defer req.Body.Close()
		w.Header().Add("Content-Type", ctype)
		w.WriteHeader(status)
		if writer != nil {
			writer(w)
		}
	}
}

func Html(str string) RequestHandler {
	return func(req *http.Request) (int, string, func(io.Writer)) {
		return Ok(str).AsHtml()
	}
}

func File(mime string, f io.Reader) RequestHandler {
	return func(req *http.Request) (int, string, func(io.Writer)) {
		return Ok(f).ContentType(mime).Build()
	}
}

type ResponseBuilder struct {
	status      int
	contentType string
	body        any
}

func NewResponse() *ResponseBuilder {
	return &ResponseBuilder{}
}

func (rb *ResponseBuilder) Status(status int) *ResponseBuilder {
	rb.status = status
	return rb
}

func (rb *ResponseBuilder) ContentType(contentType string) *ResponseBuilder {
	rb.contentType = contentType
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

func (rb *ResponseBuilder) Build() (int, string, func(io.Writer)) {
	return rb.status, rb.contentType, rb.writer()
}

func (rb *ResponseBuilder) writer() func(io.Writer) {
	switch b := rb.body.(type) {
	case string:
		return stringWriter(b)
	case error:
		return errWriter(b)
	case io.Reader:
		return readerWriter(b)
	default:
		return jsonWriter(b)
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

func Redirect(url string) (int, string, func(io.Writer)) {
	return NewResponse().Status(200).Body(fmt.Sprintf(`<html><header><script>window.location.replace("%s");</script></header><body></body></html>`, url)).AsHtml()
}

func errWriter(err error) func(writer io.Writer) {
	return func(writer io.Writer) {
		if err != nil {
			fmt.Fprintf(writer, "error: %s", err.Error())
		}
	}
}

func stringWriter(str string) func(writer io.Writer) {
	return func(writer io.Writer) {
		fmt.Fprint(writer, str)
	}
}

func jsonWriter(data any) func(writer io.Writer) {
	return func(writer io.Writer) {
		enc := json.NewEncoder(writer)
		enc.SetIndent("", "  ")
		enc.Encode(data)
	}
}

func readerWriter(data io.Reader) func(writer io.Writer) {
	return func(writer io.Writer) {
		io.Copy(writer, data)
	}
}
