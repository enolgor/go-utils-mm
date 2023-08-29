package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

type ChainHandler func(http.ResponseWriter, *http.Request) bool
type contextKey any

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
		follow := true
		for _, next := range handlers {
			follow = false
			switch v := next.(type) {
			case http.HandlerFunc:
				v(w, req)
			case func(http.ResponseWriter, *http.Request):
				v(w, req)
			case ChainHandler:
				follow = v(w, req)
			case func(http.ResponseWriter, *http.Request) bool:
				follow = v(w, req)
			default:
				panic(fmt.Sprintf("unknown method signature: %T", next))
			}
			if !follow {
				return
			}
		}
	}
}

func AddContextValue(req *http.Request, key, value any) {
	r := req.WithContext(context.WithValue(req.Context(), contextKey(key), value))
	*req = *r
}

func GetContextValue[t any](req *http.Request, key any, value *t) bool {
	v := req.Context().Value(contextKey(key))
	s, ok := v.(t)
	if ok && value != nil {
		*value = s
	}
	return ok
}

func HasContextValue[t any](req *http.Request, key any) bool {
	return GetContextValue[t](req, key, nil)
}

var Get ChainHandler = Method("GET")
var Post ChainHandler = Method("POST")
var Put ChainHandler = Method("PUT")
var Path ChainHandler = Method("PATCH")

func Html(str string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		Response(w).WithBody(str).AsHtml()
	}
}

func File(contentType string, f io.Reader) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		Response(w).WithBody(f).As(contentType)
	}
}
