package main

import (
	"fmt"
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	serverMux := http.NewServeMux()
	serverMux.Handle("/", http.FileServer(http.Dir(filepathRoot)))
	server := &http.Server{
		Addr:    ":" + port,
		Handler: serverMux,
	}
	fmt.Println("Starting server on http://localhost:" + port)
	server.ListenAndServe()
}
