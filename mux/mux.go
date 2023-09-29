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

type MatchedRoute struct {
	Method  string
	Pattern string
}

type muxContextKey uint

const (
	currentRouteContextKey muxContextKey = iota
	paramsContextKey
)

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
	route := matchRoutes(r, m.routes)
	if route == nil {
		m.NotFound(w, r)
		return
	}

	ctx := r.Context()

	// set current route in request context
	currentRoute := MatchedRoute{Method: route.method, Pattern: route.pattern}
	ctx = context.WithValue(ctx, currentRouteContextKey, currentRoute)

	// set params in request context
	params := extractParams(route.pattern, r.URL.Path)
	ctx = context.WithValue(ctx, paramsContextKey, params)

	// handle request
	handler := wrapMiddlewares(route.handler, route.middlewares...)
	handler.ServeHTTP(w, r.WithContext(ctx))
}

// CurrentRoute gets matched route from the request context.
func CurrentRoute(r *http.Request) MatchedRoute {
	return r.Context().Value(currentRouteContextKey).(MatchedRoute)
}

// Params gets URL params from the request context.
func Params(r *http.Request) map[string]string {
	return r.Context().Value(paramsContextKey).(map[string]string)
}

func matchRoutes(r *http.Request, routes []Route) *Route {
	method := r.Method
	path := r.URL.Path

	pathParts := strings.Split(path, "/")

	match := func(route *Route) bool {
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

func wrapMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
