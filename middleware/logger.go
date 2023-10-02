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
					switch statusCode {
					case http.StatusOK:
						colorSeq = "\033[32m" // green
					case http.StatusInternalServerError:
						colorSeq = "\033[31m" // red
					default:
						colorSeq = "\033[33m" // yellow
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
