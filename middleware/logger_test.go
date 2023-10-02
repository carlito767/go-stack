package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/carlito767/go-stack/middleware"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name        string
		color       bool
		code        int
		expectedLog string
	}{
		{
			name:        "200 without color",
			color:       false,
			code:        200,
			expectedLog: "[GET] \"/\" (42s)\n200 OK\n",
		},
		{
			name:        "200 with color",
			color:       true,
			code:        200,
			expectedLog: "[GET] \"/\" (42s)\n\x1b[32m200 OK\x1b[0m\n",
		},
		{
			name:        "400 with color",
			color:       true,
			code:        400,
			expectedLog: "[GET] \"/\" (42s)\n\x1b[33m400 Bad Request\x1b[0m\n",
		},
		{
			name:        "500 with color",
			color:       true,
			code:        500,
			expectedLog: "[GET] \"/\" (42s)\n\x1b[31m500 Internal Server Error\x1b[0m\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			osStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() {
				os.Stdout = osStdout
			}()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.code)
			})

			tp := middleware.FakeTimeProvider{}
			logger := middleware.NewLogger(tp, tt.color)(handler)

			req := httptest.NewRequest("GET", "/", nil)
			res := httptest.NewRecorder()

			logger.ServeHTTP(res, req)

			w.Close()
			out, _ := io.ReadAll(r)

			log := string(out)
			if tt.expectedLog != log {
				t.Errorf("log expected:%q, got:%q", tt.expectedLog, log)
			}
		})
	}
}
