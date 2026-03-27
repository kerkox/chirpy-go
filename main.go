package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	"github.com/kerkox/chirpy-go/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening databse: %s", err)
	}
	dbQueries := database.New(dbConn)

	cfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      dbQueries,
		platform:       platform,
	}

	serverMux := http.NewServeMux()
	appFileServerHandler := http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))
	serverMux.Handle("/app/", cfg.middlewareMetricsInc(appFileServerHandler))

	serverMux.HandleFunc("GET /api/healthz", handlerReadiness)

	serverMux.HandleFunc("POST /api/users", cfg.handlerUsersCreate)

	serverMux.HandleFunc("POST /api/chirps", cfg.handlerChirpsCreate)
	serverMux.HandleFunc("GET /api/chirps", cfg.handlerChirpsRetrieve)

	serverMux.HandleFunc("POST /admin/reset", cfg.handlerReset)
	serverMux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serverMux,
	}
	log.Printf("Starting server on http://localhost:%s\n", port)
	log.Fatal(server.ListenAndServe())
}
