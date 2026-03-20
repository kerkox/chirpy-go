package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
	serverMux.HandleFunc("GET /api/healthz", handleHealthz)
	serverMux.HandleFunc("GET /admin/metrics", cfg.getFileServeHits)
	serverMux.HandleFunc("POST /admin/reset", cfg.resetFileServeHits)
	serverMux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serverMux,
	}
	fmt.Println("Starting server on http://localhost:" + port)
	server.ListenAndServe()
}

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type chirpRequest struct {
		Body string `json:"body"`
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	type chirpResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	chirp := chirpRequest{}
	err := decoder.Decode(&chirp)
	if err != nil {
		log.Printf("Error decoding request body: %s", err)
		response := errorResponse{Error: "Somehting went wrong"}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		return
	}
	MAX_CHIRP_LENGTH := 140
	if len(chirp.Body) > MAX_CHIRP_LENGTH {
		response := errorResponse{Error: "Chirp is too long"}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResponse)
		return
	}
	cleanChirp := CleanChirpProfaneWords(chirp.Body)
	response := chirpResponse{CleanedBody: cleanChirp}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func CleanChirpProfaneWords(chirp string) string {
	profaneWords := map[string]bool{"kerfuffle": true, "sharbert": true, "fornax": true}
	var result strings.Builder
	for word := range strings.SplitSeq(chirp, " ") {
		if profaneWords[strings.ToLower(word)] {
			result.WriteString("**** ")
		} else {
			result.WriteString(word + " ")
		}
	}
	return strings.TrimSpace(result.String())
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
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserveHits.Load())))
}

func (cfg *apiConfig) resetFileServeHits(w http.ResponseWriter, r *http.Request) {
	cfg.fileserveHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
