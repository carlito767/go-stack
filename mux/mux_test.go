package mux_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/carlito767/go-stack/mux"
)

func m(msg string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(msg))
			next.ServeHTTP(w, r)
		})
	}
}

func TestPanic(t *testing.T) {
	h := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

	tests := []struct {
		name    string
		method  string
		path    string
		handler http.Handler
		msg     string
	}{
		{
			name:    "empty path",
			method:  "GET",
			path:    "",
			handler: h,
			msg:     "path must begin with '/'",
		},
		{
			name:    "invalid path",
			method:  "GET",
			path:    "invalid/path",
			handler: h,
			msg:     "path must begin with '/'",
		},
		{
			name:    "invalid handler",
			method:  "GET",
			path:    "/",
			handler: nil,
			msg:     "handler must not be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("the code did not panic (%q)", tt.name)
				} else {
					msg := r.(string)
					if msg != tt.msg {
						t.Errorf("panic message expected: '%s', got: '%s'", tt.msg, msg)
					}
				}
			}()

			router := mux.NewRouter()
			router.Handle(tt.method, tt.path).Use().Then(tt.handler)
		})
	}
}

func TestRouter(t *testing.T) {
	router := mux.NewRouter()
	router.Use(m("1"), m("2"))

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// set routes
	router.GET("/path/{id}").
		Use(m("3"), m("4")).
		Then(testHandler)

	// test the router with a valid route
	req := httptest.NewRequest("GET", "/path/123", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Errorf("status code expected: %d, got: %d", http.StatusOK, res.Code)
	}

	// check middlewares
	expected := "1234"
	if res.Body.String() != expected {
		t.Fatalf("response body expected: %#v, got: %#v", expected, res.Body.String())
	}

	// test the router with an invalid route
	paths := []string{"/invalid/path", "/1/2/3"}
	for _, path := range paths {
		req = httptest.NewRequest("GET", path, nil)
		res = httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != http.StatusNotFound {
			t.Errorf("status code expected: %d, got: %d", http.StatusNotFound, res.Code)
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

	res := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/get", nil)
	m.ServeHTTP(res, req)
	if !get {
		t.Fatalf("routing GET failed")
	}

	req, _ = http.NewRequest("POST", "/post", nil)
	m.ServeHTTP(res, req)
	if !post {
		t.Fatalf("routing POST failed")
	}

	req, _ = http.NewRequest("PUT", "/put", nil)
	m.ServeHTTP(res, req)
	if !put {
		t.Fatalf("routing PUT failed")
	}

	req, _ = http.NewRequest("PATCH", "/patch", nil)
	m.ServeHTTP(res, req)
	if !patch {
		t.Fatalf("routing PATCH failed")
	}

	req, _ = http.NewRequest("DELETE", "/delete", nil)
	m.ServeHTTP(res, req)
	if !delete {
		t.Fatalf("routing DELETE failed")
	}
}

func TestParams(t *testing.T) {
	router := mux.NewRouter()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// verify that params are correct
		expectedParams := map[string]string{"id": "123"}
		for key, value := range expectedParams {
			v := r.PathValue(key)
			if v != value {
				t.Errorf("param expected: %s=%s, got: %s=%s", key, value, key, v)
			}
		}
	})

	router.GET("/path/{id}").Then(testHandler)

	req := httptest.NewRequest("GET", "/path/123", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
}

func TestNewSubRouter(t *testing.T) {
	rooter := mux.NewRouter()
	apiV0 := rooter.NewSubRouter("/api/v0")

	// register handlers on root and subrouter
	rooter.GET("/").ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root"))
	})

	apiV0.GET("/hello").ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello v0"))
	})

	// create test server
	server := httptest.NewServer(rooter)
	defer server.Close()

	tests := []struct {
		path string
		want string
	}{
		{"/", "root"},
		{"/api/v0/hello", "hello v0"},
	}

	for _, tt := range tests {
		resp, err := http.Get(server.URL + tt.path)
		if err != nil {
			t.Fatalf("GET %s failed: %v", tt.path, err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		got := string(body)
		if got != tt.want {
			t.Errorf("GET %q = %q; want %q", tt.path, got, tt.want)
		}
	}
}

func TestSubRouterPrefixChaining(t *testing.T) {
	rooter := mux.NewRouter()
	apiV0 := rooter.NewSubRouter("/api/v0")
	admin := apiV0.NewSubRouter("/admin")

	admin.GET("/status").ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("admin status"))
	})

	server := httptest.NewServer(rooter)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v0/admin/status")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if got := string(body); got != "admin status" {
		t.Errorf("GET /api/v0/admin/status = %q; want %q", got, "admin status")
	}
}
