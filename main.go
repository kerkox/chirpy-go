package main

import (
	"fmt"
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	serverMux := http.NewServeMux()
	serverMux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot))))
	serverMux.HandleFunc("/healthz", handleHealthz)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serverMux,
	}
	fmt.Println("Starting server on http://localhost:" + port)
	server.ListenAndServe()
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
