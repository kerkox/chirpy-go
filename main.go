package main

import (
	"net/http"
)

func main() {
	serverMux := new(http.ServeMux)
	server := &http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}

	server.ListenAndServe()
}