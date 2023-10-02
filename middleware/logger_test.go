package middleware_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/carlito767/go-stack/middleware"
)

func TestLogger(t *testing.T) {
	osStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = osStdout
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	tp := middleware.FakeTimeProvider{}
	logger := middleware.NewLogger(tp)(handler)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	logger.ServeHTTP(res, req)

	w.Close()
	out, _ := io.ReadAll(r)

	expectedLog := fmt.Sprintf("[GET] \"/\" (%v)\n404 Not Found\n", tp.Since(tp.Now()))
	actualLog := string(out)
	if expectedLog != actualLog {
		t.Errorf("log expected:\n%s\n, got:\n%s\n", expectedLog, actualLog)
	}
}
