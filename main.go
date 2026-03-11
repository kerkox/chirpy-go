package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserveHits atomic.Int32
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	serverMux := http.NewServeMux()
	appFileServerHandler := http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))
	cfg := &apiConfig{
		fileserveHits: atomic.Int32{},
	}
	serverMux.Handle("/app/", cfg.middlewareMetricsInc(appFileServerHandler))
	serverMux.HandleFunc("/healthz", handleHealthz)
	serverMux.HandleFunc("/metrics", cfg.getFileServeHits)
	serverMux.HandleFunc("/reset", cfg.resetFileServeHits)

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

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserveHits.Add(1)
		fmt.Println("Hits: ", cfg.fileserveHits.Load())
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) getFileServeHits(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserveHits.Load())))
}

func (cfg *apiConfig) resetFileServeHits(w http.ResponseWriter, r *http.Request) {
	cfg.fileserveHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
