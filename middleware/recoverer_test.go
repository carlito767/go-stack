package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/carlito767/go-stack/middleware"
)

func TestRecoverer(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("some error")
	})

	recoverer := middleware.Recoverer(handler)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	recoverer.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("status code expected:%d, got:%d", http.StatusInternalServerError, res.Code)
	}
}
