package goweb

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
)

// Route is an HTTP Route with optional children.
type Route struct {
	method     string
	path       string
	handler    http.Handler
	children   []*Route
	middleware []Middleware
}

// Handler creates a simple route from an [http.Handler].
func Handler(method, path string, handler http.Handler) *Route {
	return &Route{
		method:  method,
		path:    path,
		handler: handler,
	}
}

// Func creates a simple route from a handler function
func Func(method, path string, f http.HandlerFunc) *Route {
	return &Route{
		method:  method,
		path:    path,
		handler: f,
	}
}

// Group creates a route group without an own handler.
func Group(path string, children []*Route) *Route {
	return &Route{
		path:     path,
		children: children,
	}
}

// Prefix wraps a route with a path prefix.
func Prefix(path string, r *Route) *Route {
	return &Route{
		path:     path,
		children: []*Route{r},
	}
}

// Middleware adds middleware to the route.
func (r *Route) Middleware(mw ...Middleware) *Route {
	r.middleware = append(r.middleware, mw...)
	return r
}

// Build the route into an HTTP handler.
func (r *Route) Build() (*httprouter.Router, error) {
	router := httprouter.New()
	err := r.register(router, "/", nil)
	if err != nil {
		return nil, err
	}
	return router, nil
}

// register the route and all its children.
func (r *Route) register(router *httprouter.Router, prefix string, mw []Middleware) error {
	// Prepend the parent path prefix
	path, err := url.JoinPath(prefix, r.path)
	if err != nil {
		return err
	}

	// Prepend the parent's middleware
	mw = append(mw, r.middleware...)

	// Register the route handler if it has one.
	if r.handler != nil {
		handler := r.handler
		for i := len(mw) - 1; i >= 0; i-- {
			handler = mw[i].Handler(handler)
		}
		router.Handler(r.method, path, handler)
	}

	// Recursively register the route's children.
	for _, child := range r.children {
		err := child.register(router, path, mw)
		if err != nil {
			return err
		}
	}

	return nil
}

// Dump returns string representations of the route and its children.
func (r *Route) Dump() []string {
	return r.dump("/")
}

// dump returns string representations of the route and its children.
func (r *Route) dump(prefix string) []string {
	path, _ := url.JoinPath(prefix, r.path)

	var routes []string
	if r.handler != nil {
		routes = append(routes, fmt.Sprintf("%s %s", r.method, path))
	}

	for _, child := range r.children {
		routes = append(routes, child.dump(path)...)
	}

	return routes
}

// Middleware is the interface for an HTTP middleware
type Middleware interface {
	Handler(http.Handler) http.Handler
}

// MiddlewareFunc wraps a middleware function to satisfy the Middleware interface.
type MiddlewareFunc func(http.Handler) http.Handler

// Handler returns the middleware's handler.
func (f MiddlewareFunc) Handler(h http.Handler) http.Handler {
	return f(h)
}
