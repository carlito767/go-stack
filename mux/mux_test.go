package mux_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/carlito767/go-stack/mux"
)

// Middlewares

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add logic here
		w.Write([]byte("logging\n"))

		next.ServeHTTP(w, r)
	})
}

func middleware1(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add logic here
		w.Write([]byte("1"))

		next.ServeHTTP(w, r)
	})
}

func middleware2(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add logic here
		w.Write([]byte("2"))

		next.ServeHTTP(w, r)
	})
}

func middleware3(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add logic here
		w.Write([]byte("3"))

		next.ServeHTTP(w, r)
	})
}

func TestRouter(t *testing.T) {
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	notFoundHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	// set routes
	router.GET("/nil").Then(nil)

	router.GET("/path/:id").
		Use(middleware1, middleware2, middleware3).
		Then(testHandler)

	// test the router with an invalid handler
	r := httptest.NewRequest("GET", "/nil", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("status code expected: %d, got: %d", http.StatusOK, w.Code)
	}

	// test the router with a valid route
	r = httptest.NewRequest("GET", "/path/123", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("status code expected: %d, got: %d", http.StatusOK, w.Code)
	}

	// check middlewares
	expected := "logging\n123"
	if w.Body.String() != expected {
		t.Fatalf("response body expected: %#v, got: %#v", expected, w.Body.String())
	}

	// test the router with an invalid route
	paths := []string{"/invalid/path", "/1/2/3"}
	for _, path := range paths {
		r = httptest.NewRequest("GET", path, nil)
		w = httptest.NewRecorder()
		router.NotFound = notFoundHandler // set custom 404 handler
		router.ServeHTTP(w, r)
		if w.Code != http.StatusNotFound {
			t.Errorf("status code expected: %d, got: %d", http.StatusNotFound, w.Code)
		}
	}
}

func TestRoutes(t *testing.T) {
	var get, post, put, patch, delete bool

	m := mux.NewRouter()

	m.GET("/get").ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		get = true
	})

	m.POST("/post").ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		post = true
	})

	m.PUT("/put").ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		put = true
	})

	m.PATCH("/patch").ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		patch = true
	})

	m.DELETE("/delete").ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		delete = true
	})

	w := httptest.NewRecorder()

	r, _ := http.NewRequest("GET", "/get", nil)
	m.ServeHTTP(w, r)
	if !get {
		t.Fatalf("routing GET failed")
	}

	r, _ = http.NewRequest("POST", "/post", nil)
	m.ServeHTTP(w, r)
	if !post {
		t.Fatalf("routing POST failed")
	}

	r, _ = http.NewRequest("PUT", "/put", nil)
	m.ServeHTTP(w, r)
	if !put {
		t.Fatalf("routing PUT failed")
	}

	r, _ = http.NewRequest("PATCH", "/patch", nil)
	m.ServeHTTP(w, r)
	if !patch {
		t.Fatalf("routing PATCH failed")
	}

	r, _ = http.NewRequest("DELETE", "/delete", nil)
	m.ServeHTTP(w, r)
	if !delete {
		t.Fatalf("routing DELETE failed")
	}
}

func TestContext(t *testing.T) {
	router := mux.NewRouter()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// verify that current route is correct
		currentRoute := mux.CurrentRoute(r)
		expectedRoute := mux.MatchedRoute{Method: "GET", Pattern: "/path/:id"}
		if currentRoute != expectedRoute {
			t.Errorf("route expected: %v, got: %v", expectedRoute, currentRoute)
		}

		// verify that params are correct
		params := mux.Params(r)
		expectedParams := map[string]string{"id": "123"}
		for key, value := range expectedParams {
			if params[key] != value {
				t.Errorf("param expected: %s=%s, got: %s=%s", key, value, key, params[key])
			}
		}
	})

	router.GET("/path/:id").Then(testHandler)

	req := httptest.NewRequest("GET", "/path/123", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
}
