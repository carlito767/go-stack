package main

import (
	"fmt"
	"net/http"

	"github.com/carlito767/go-stack/mux"
)

// Global logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Request from %s for %s\n", r.RemoteAddr, r.URL)
		next.ServeHTTP(w, r)
	})
}

// Global authentication middleware
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add global authentication logic here
		next.ServeHTTP(w, r)
	})
}

// Handlers

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, 404, http.StatusText(404), "[custom 'Not Found' handler]")
}

func pathHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Path with ID:", mux.Params(r)["id"])
}

func main() {
	router := mux.NewRouter()
	router.NotFound = notFoundHandler

	// set routes
	router.GET("/path/:id").
		Use(loggingMiddleware, authMiddleware).
		ThenFunc(pathHandler)

	// start the HTTP server listening on port 8080
	http.ListenAndServe(":8080", router)
}
