/*
Package mux implements a request router and dispatcher for matching incoming requests to their respective handler.

Inspired by:

	https://github.com/nmerouze/stack

# Usage

	router := mux.NewRouter()
	router.Use(globalMiddleware1, globalMiddleware2, ...)
	router.Handle(method, path).Use(middleware1, middleware2, ...).Then(handler)
	http.ListenAndServe(addr, router)

	See cmd/server/main.go for example.
*/
package mux

import (
	"fmt"
	"net/http"
)

type Mux struct {
	NotFound http.HandlerFunc

	mux         *http.ServeMux
	middlewares []middleware
}

type middleware = func(http.Handler) http.Handler

type route struct {
	m           *Mux
	method      string
	path        string
	middlewares []middleware
	handler     http.Handler
}

type muxContextKey uint

const (
	paramsContextKey muxContextKey = iota
)

func NewRouter() *Mux {
	return &Mux{
		NotFound: http.NotFound,
		mux:      http.NewServeMux(),
	}
}

// Use adds global middlewares to the router.
func (m *Mux) Use(middlewares ...middleware) *Mux {
	m.middlewares = append(m.middlewares, middlewares...)
	return m
}

// Handle sets a route with a custom HTTP method.
func (m *Mux) Handle(method string, path string) *route {
	return &route{m: m, method: method, path: path}
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
	if len(r.path) < 1 || r.path[0] != '/' {
		panic("path must begin with '/'")
	}
	if h == nil {
		panic("handler must not be nil")
	}

	middlewares := append(r.m.middlewares, r.middlewares...)
	for i := range middlewares {
		h = middlewares[len(middlewares)-1-i](h)
	}
	r.handler = h

	pattern := fmt.Sprintf("%s %s", r.method, r.path)
	r.m.mux.Handle(pattern, h)
}

// ThenFunc sets the final handler for a route using an http.HandlerFunc.
func (r *route) ThenFunc(h http.HandlerFunc) {
	r.Then(http.HandlerFunc(h))
}

// ServeHTTP implements the http.Handler interface for the router.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}
