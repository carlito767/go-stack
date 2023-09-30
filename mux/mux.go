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

	middlewares []middleware
	routes      []route
}

type middleware = func(http.Handler) http.Handler

type route struct {
	mux         *Mux
	method      string
	pattern     string
	middlewares []middleware
	handler     http.Handler
}

type muxContextKey uint

const (
	paramsContextKey muxContextKey = iota
)

func NewRouter() *Mux {
	return &Mux{NotFound: http.NotFound}
}

// Use adds global middlewares to the router.
func (m *Mux) Use(middlewares ...middleware) *Mux {
	m.middlewares = append(m.middlewares, middlewares...)
	return m
}

// Handle sets a route with a custom HTTP method.
func (m *Mux) Handle(method string, pattern string) *route {
	return &route{mux: m, method: method, pattern: pattern}
}

// GET sets a route with the GET HTTP method.
func (m *Mux) GET(p string) *route {
	return m.Handle("GET", p)
}

// POST sets a route with the POST HTTP method.
func (m *Mux) POST(p string) *route {
	return m.Handle("POST", p)
}

// PUT sets a route with the PUT HTTP method.
func (m *Mux) PUT(p string) *route {
	return m.Handle("PUT", p)
}

// PATCH sets a route with the PATCH HTTP method.
func (m *Mux) PATCH(p string) *route {
	return m.Handle("PATCH", p)
}

// DELETE sets a route with the DELETE HTTP method.
func (m *Mux) DELETE(p string) *route {
	return m.Handle("DELETE", p)
}

// Use adds middlewares to a specific route.
func (r *route) Use(middlewares ...middleware) *route {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

// Then sets the final handler for a route using an http.Handler.
func (r *route) Then(h http.Handler) {
	if r.method == "" {
		panic("method must not be empty")
	}
	if len(r.pattern) < 1 || r.pattern[0] != '/' {
		panic("pattern must begin with '/'")
	}
	if h == nil {
		panic("handler must not be nil")
	}

	middlewares := append(r.mux.middlewares, r.middlewares...)
	for i := range middlewares {
		h = middlewares[len(middlewares)-1-i](h)
	}
	r.handler = h
	r.mux.routes = append(r.mux.routes, *r)
}

// ThenFunc sets the final handler for a route using an http.HandlerFunc.
func (r *route) ThenFunc(h http.HandlerFunc) {
	r.Then(http.HandlerFunc(h))
}

// ServeHTTP implements the http.Handler interface for the router.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := matchRoutes(r, m.routes)
	if route == nil {
		m.NotFound(w, r)
		return
	}

	ctx := r.Context()

	// set params in request context
	params := extractParams(route.pattern, r.URL.Path)
	ctx = context.WithValue(ctx, paramsContextKey, params)

	// handle request
	route.handler.ServeHTTP(w, r.WithContext(ctx))
}

// Params gets URL params from the request context.
func Params(r *http.Request) map[string]string {
	return r.Context().Value(paramsContextKey).(map[string]string)
}

func matchRoutes(r *http.Request, routes []route) *route {
	method := r.Method
	path := r.URL.Path

	pathParts := strings.Split(path, "/")

	match := func(route *route) bool {
		if route.method != method {
			return false
		}

		patternParts := strings.Split(route.pattern, "/")
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

	for _, route := range routes {
		if match(&route) {
			// matched route
			return &route
		}
	}

	return nil
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
