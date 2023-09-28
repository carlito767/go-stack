/*
Package mux implements a request router and dispatcher for matching incoming requests to their respective handler.

Inspired by:

	https://github.com/nmerouze/stack
*/
package mux

import (
	"context"
	"net/http"
	"strings"
)

type Mux struct {
	NotFound http.HandlerFunc

	middlewares []Middleware
	routes      []Route
}

type Middleware = func(http.Handler) http.Handler

type Route struct {
	mux         *Mux
	method      string
	pattern     string
	middlewares []Middleware
	handler     http.Handler
}

func NewRouter() *Mux {
	return &Mux{NotFound: http.NotFound}
}

// Use adds global middlewares to the router.
func (m *Mux) Use(middlewares ...Middleware) *Mux {
	m.middlewares = append(m.middlewares, middlewares...)
	return m
}

// GET sets a route with the GET HTTP method.
func (m *Mux) GET(p string) *Route {
	return &Route{mux: m, method: "GET", pattern: p, middlewares: m.middlewares}
}

// POST sets a route with the POST HTTP method.
func (m *Mux) POST(p string) *Route {
	return &Route{mux: m, method: "POST", pattern: p, middlewares: m.middlewares}
}

// PUT sets a route with the PUT HTTP method.
func (m *Mux) PUT(p string) *Route {
	return &Route{mux: m, method: "PUT", pattern: p, middlewares: m.middlewares}
}

// PATCH sets a route with the PATCH HTTP method.
func (m *Mux) PATCH(p string) *Route {
	return &Route{mux: m, method: "PATCH", pattern: p, middlewares: m.middlewares}
}

// DELETE sets a route with the DELETE HTTP method.
func (m *Mux) DELETE(p string) *Route {
	return &Route{mux: m, method: "DELETE", pattern: p, middlewares: m.middlewares}
}

// Use adds middlewares to a specific route.
func (r *Route) Use(middlewares ...Middleware) *Route {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

// Then sets the final handler for a route using an http.Handler.
func (r *Route) Then(h http.Handler) {
	r.handler = h
	r.mux.routes = append(r.mux.routes, *r)
}

// ThenFunc sets the final handler for a route using an http.HandlerFunc.
func (r *Route) ThenFunc(h http.HandlerFunc) {
	r.Then(http.HandlerFunc(h))
}

// ServeHTTP implements the http.Handler interface for the router.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	for _, route := range m.routes {
		if route.method == method && pathMatch(route.pattern, path) {
			// set params in request context
			params := extractParams(route.pattern, path)
			ctx := r.Context()
			ctx = context.WithValue(ctx, "params", params)
			r = r.WithContext(ctx)

			// handle request
			handler := wrapMiddlewares(route.handler, route.middlewares...)
			handler.ServeHTTP(w, r)
			return
		}
	}

	m.NotFound(w, r)
}

// Params gets URL params from the request context.
func Params(r *http.Request) map[string]string {
	return r.Context().Value("params").(map[string]string)
}

func pathMatch(pattern, path string) bool {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, part := range patternParts {
		if part != pathParts[i] && !strings.HasPrefix(part, ":") {
			return false
		}
	}

	return true
}

func extractParams(pattern, path string) map[string]string {
	params := make(map[string]string)
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") && i < len(pathParts) {
			paramName := part[1:]
			params[paramName] = pathParts[i]
		}
	}

	return params
}

func wrapMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
