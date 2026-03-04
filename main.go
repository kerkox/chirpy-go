package main

import (
	"fmt"
	"net/http"
)

func main() {
	serverMux := http.NewServeMux()
	serverMux.Handle("/", http.FileServer(http.Dir("./")))
	server := &http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}
	fmt.Println("Starting server on http://localhost:8080")
	server.ListenAndServe()
}
