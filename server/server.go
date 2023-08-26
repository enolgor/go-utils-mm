package server

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type ChainHandler func(w http.ResponseWriter, req *http.Request) bool

func Method(method string) ChainHandler {
	return func(w http.ResponseWriter, req *http.Request) bool {
		if req.Method == method {
			return true
		}
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(400)
		fmt.Fprint(w, "unsupported method")
		return false
	}
}

func BasicAuthHandler(user, pass string) ChainHandler {
	b64auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
	return func(w http.ResponseWriter, req *http.Request) bool {
		if b64auth != req.Header.Get("Authorization") {
			w.Header().Add("Content-Type", "text/plain")
			w.Header().Add("WWW-Authenticate", `Basic realm="Realm"`)
			w.WriteHeader(401)
			fmt.Fprint(w, "unauthorized")
			return false
		}
		return true
	}
}

func NopHandler() ChainHandler {
	return func(w http.ResponseWriter, req *http.Request) bool {
		return true
	}
}

func Handle(handlers ...any) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		stop := false
		for _, next := range handlers {
			stop = true
			switch v := next.(type) {
			case func(*url.Values) (int, string, func(io.Writer)):
				serveQuery(v)(w, req)
			case func(*url.Values, io.Reader) (int, string, func(io.Writer)):
				serveBody(v)(w, req)
			case func(req *http.Request) (int, string, func(io.Writer)):
				serveRequest(v)(w, req)
			case ChainHandler:
				stop = !v(w, req)
			default:
				stop = false
			}
			if stop {
				return
			}
		}
	}
}

var Get ChainHandler = Method("GET")
var Post ChainHandler = Method("POST")
var Put ChainHandler = Method("PUT")
var Path ChainHandler = Method("PATCH")

func serveQuery(f func(*url.Values) (int, string, func(io.Writer))) func(w http.ResponseWriter, req *http.Request) {
	return serveRequest(func(req *http.Request) (int, string, func(io.Writer)) {
		query := req.URL.Query()
		return f(&query)
	})
}

func serveBody(f func(*url.Values, io.Reader) (int, string, func(io.Writer))) func(w http.ResponseWriter, req *http.Request) {
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

func Html(str string) func(req *http.Request) (int, string, func(io.Writer)) {
	return func(req *http.Request) (int, string, func(io.Writer)) {
		return Ok(str).AsHtml()
	}
}

func File(mime string, f io.Reader) func(req *http.Request) (int, string, func(io.Writer)) {
	return func(req *http.Request) (int, string, func(io.Writer)) {
		return Ok(f).As(mime)
	}
}
