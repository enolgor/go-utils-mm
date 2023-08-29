package server

import (
	"fmt"
	"net/http"
	"net/url"
)

type Router struct {
	routes   []route
	notFound http.HandlerFunc
}

type route struct {
	method  string
	path    string
	handler http.HandlerFunc
}

func NewRouter() *Router {
	return &Router{routes: []route{}, notFound: nil}
}

func (r *Router) Register(method string, path string, handler http.HandlerFunc) {
	r.routes = append(r.routes, route{method, path, handler})
}

func (r *Router) NotFound(handler http.HandlerFunc) {
	r.notFound = handler
}

type routerContextKey int

const pathParamsKey routerContextKey = iota

func PathParams(req *http.Request) url.Values {
	values := make(url.Values)
	GetContextValue(req, pathParamsKey, &values)
	return values
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pathParams := make(url.Values)
	AddContextValue(req, pathParamsKey, pathParams)
	for _, route := range r.routes {
		if req.Method == route.method && route.path == req.URL.Path {
			route.handler(w, req)
			return
		}
	}
	if r.notFound != nil {
		r.notFound(w, req)
		return
	}
	Response(w).Status(http.StatusNotFound).WithBody(fmt.Sprintf("%s %s not found", req.Method, req.URL.Path)).AsTextPlain()
}
