package server

import (
	"fmt"
	"net/http"

	"github.com/enolgor/go-utils/server/path"
)

type Router struct {
	routes      []route
	notFound    http.HandlerFunc
	internalErr http.HandlerFunc
}

type RouterBuilder struct {
	router *Router
}

type route struct {
	method   string
	pathExpr string
	matcher  func(string, map[any]string) bool
	handler  http.HandlerFunc
}

func NewRouterBuilder() *RouterBuilder {
	return &RouterBuilder{&Router{routes: []route{}, notFound: nil, internalErr: nil}}
}

func (r *RouterBuilder) register(method string, pathExpr string, handler http.HandlerFunc) *RouterBuilder {
	r.router.routes = append(r.router.routes, route{method, pathExpr, nil, handler})
	return r
}

func (r *RouterBuilder) Get(pathExpr string, handler http.HandlerFunc) *RouterBuilder {
	return r.register("GET", pathExpr, handler)
}

func (r *RouterBuilder) Post(pathExpr string, handler http.HandlerFunc) *RouterBuilder {
	return r.register("POST", pathExpr, handler)
}

func (r *RouterBuilder) Put(pathExpr string, handler http.HandlerFunc) *RouterBuilder {
	return r.register("PUT", pathExpr, handler)
}

func (r *RouterBuilder) Patch(pathExpr string, handler http.HandlerFunc) *RouterBuilder {
	return r.register("PATCH", pathExpr, handler)
}

func (r *RouterBuilder) Delete(pathExpr string, handler http.HandlerFunc) *RouterBuilder {
	return r.register("DELETE", pathExpr, handler)
}

func (r *RouterBuilder) NotFound(handler http.HandlerFunc) *RouterBuilder {
	r.router.notFound = handler
	return r
}

func (r *RouterBuilder) InternalErr(handler http.HandlerFunc) *RouterBuilder {
	r.router.internalErr = handler
	return r
}

var defaultNotFound = func(w http.ResponseWriter, req *http.Request) {
	Response(w).Status(http.StatusNotFound).WithBody(fmt.Sprintf("%s %s not found", req.Method, req.URL.Path)).AsTextPlain()
}

var defaultInternalErr = func(w http.ResponseWriter, req *http.Request) {
	err := Recover(req)
	if err == nil {
		err = "uknown"
	}
	Response(w).Status(http.StatusInternalServerError).WithBody(fmt.Sprintf("internal server error: %s", err)).AsTextPlain()
}

func (r *RouterBuilder) Build() (*Router, error) {
	var err error
	for i := range r.router.routes {
		if r.router.routes[i].matcher, err = path.Matcher(r.router.routes[i].pathExpr); err != nil {
			return nil, err
		}
	}
	if r.router.notFound == nil {
		r.router.notFound = defaultNotFound
	}
	if r.router.internalErr == nil {
		r.router.internalErr = defaultInternalErr
	}
	return r.router, nil
}

type routerContextKey int

const (
	pathParamsKey routerContextKey = iota
	panicKey
)

func PathParams(req *http.Request) map[any]string {
	values := map[any]string{}
	GetContextValue(req, pathParamsKey, &values)
	return values
}

func Recover(req *http.Request) any {
	var err any
	GetContextValue(req, panicKey, &err)
	return err
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			AddContextValue(req, panicKey, rec)
			r.internalErr(w, req)
		}
	}()
	pathParams := make(map[any]string)
	for _, route := range r.routes {
		if req.Method == route.method && route.matcher(req.URL.Path, pathParams) {
			AddContextValue(req, pathParamsKey, pathParams)
			route.handler(w, req)
			return
		}
	}
	r.notFound(w, req)
}
