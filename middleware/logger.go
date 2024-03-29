package middleware

import (
	"fmt"
	"net/http"
)

// Logger is a middleware that logs every request with some useful data.
var Logger = NewLogger(MockTimeProvider{}, true)

// NewLogger creates a middleware that logs every request with some useful data.
func NewLogger(tp TimeProvider, color bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			lrw := newLoggingResponseWriter(w)

			t := tp.Now()
			defer func() {
				statusCode := lrw.statusCode
				var colorSeq, resetSeq string
				if color {
					// https://zetcode.com/golang/terminal-colour/
					switch {
					case statusCode >= 100 && statusCode <= 199:
						// 1xx informational response
						colorSeq = "\033[34m" // blue
					case statusCode >= 200 && statusCode <= 299:
						// 2xx success
						colorSeq = "\033[32m" // green
					case statusCode >= 300 && statusCode <= 399:
						// 3xx redirection
						colorSeq = "\033[33m" // yellow
					case statusCode >= 400 && statusCode <= 499:
						// 4xx client errors
						colorSeq = "\033[31m" // red
					case statusCode >= 500 && statusCode <= 599:
						// 5xx server errors
						colorSeq = "\033[35m" // purple
					}
					resetSeq = "\033[0m" // reset color
				}
				fmt.Printf("[%s] %q (%v)\n", r.Method, r.URL.String(), tp.Since(t))
				msg := fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode))
				fmt.Printf("%s%s%s\n", colorSeq, msg, resetSeq)
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
