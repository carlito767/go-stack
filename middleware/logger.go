package middleware

import (
	"fmt"
	"net/http"
)

// Logger is a middleware that logs every request with some useful data.
var Logger = NewLogger(MockTimeProvider{})

// NewLogger creates a middleware that logs every request with some useful data.
func NewLogger(tp TimeProvider) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			lrw := newLoggingResponseWriter(w)

			t := tp.Now()
			defer func() {
				statusCode := lrw.statusCode
				fmt.Printf("[%s] %q (%v)\n%d %s\n", r.Method, r.URL.String(), tp.Since(t), statusCode, http.StatusText(statusCode))
			}()

			next.ServeHTTP(lrw, r)
		}

		return http.HandlerFunc(fn)
	}
}

//
// Logging Response Writer
// https://ndersson.me/post/capturing_status_code_in_net_http/
//

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	// WriteHeader(int) is not called if our response implicitly returns 200 OK, so
	// we default to that status code.
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
