package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/carlito767/go-stack/clp"
	"github.com/carlito767/go-stack/middleware"
	"github.com/carlito767/go-stack/mux"
)

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

	// set global middlewares
	router.Use(middleware.Logger, middleware.Recoverer)

	// set routes
	router.GET("/panic").ThenFunc(panicHandler)
	router.GET("/status/:code").ThenFunc(statusHandler)

	// start server
	addr := fmt.Sprintf("%v:%v", config.Host, config.Port)
	fmt.Printf("server listening on %s\n", addr)
	http.ListenAndServe(addr, router)
}

func panicHandler(w http.ResponseWriter, r *http.Request) {
	panic("panic test")
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Params(r)
	code, _ := strconv.Atoi(params["code"])
	if http.StatusText(code) == "" {
		code = http.StatusBadRequest
	}
	w.WriteHeader(code)
	fmt.Fprintf(w, "Status code: %d %s\n", code, http.StatusText(code))
}
