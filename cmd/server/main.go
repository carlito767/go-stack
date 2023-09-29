package main

import (
	"fmt"
	"net/http"

	"github.com/carlito767/go-stack/clp"
	"github.com/carlito767/go-stack/mux"
)

// Middlewares

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%s] %v\n", r.Method, r.URL)

		next.ServeHTTP(w, r)
	})
}

func currentRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentRoute := mux.CurrentRoute(r)
		fmt.Printf("matched pattern: %s\n", currentRoute.Pattern)

		next.ServeHTTP(w, r)
	})
}

// Handlers

func pathHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Path with ID:", mux.Params(r)["id"])
}

func main() {
	// load config
	config := struct {
		Host string `name:"host,h"`
		Port uint   `name:"port,p"`
	}{
		Host: "localhost",
		Port: 8080,
	}
	if err := clp.ParseOptions(&config); err != nil {
		fmt.Println("error:", err)
		return
	}

	// create router
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	// set routes
	router.GET("/path/:id").Use(currentRouteMiddleware).ThenFunc(pathHandler)

	// start server
	addr := fmt.Sprintf("%v:%v", config.Host, config.Port)
	fmt.Printf("server listening on %s\n", addr)
	http.ListenAndServe(addr, router)
}
