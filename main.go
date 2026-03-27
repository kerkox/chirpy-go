package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/kerkox/chirpy-go/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserveHits atomic.Int32
	dbQueries     *database.Queries
	platform      string
}

type Chirp struct {
	ID        string `json:"id"`
	Body      string `json:"body"`
	UserId    string `json:"user_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func main() {
	godotenv.Load()
	const filepathRoot = "."
	const port = "8080"

	dbURL := os.Getenv("DB_URL")
	db, _ := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	serverMux := http.NewServeMux()
	appFileServerHandler := http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))
	cfg := &apiConfig{
		fileserveHits: atomic.Int32{},
		dbQueries:     dbQueries,
		platform:      os.Getenv("PLATFORM"),
	}
	serverMux.Handle("/app/", cfg.middlewareMetricsInc(appFileServerHandler))
	serverMux.HandleFunc("GET /api/healthz", handleHealthz)
	serverMux.HandleFunc("GET /admin/metrics", cfg.getFileServeHits)
	serverMux.HandleFunc("POST /admin/reset", cfg.resetFileServeHits)
	serverMux.HandleFunc("POST /api/users", cfg.handleCreateUser)
	serverMux.HandleFunc("POST /api/chirps", cfg.handleCreateChirp)
	serverMux.HandleFunc("GET /api/chirps", cfg.handleGetChirps)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serverMux,
	}
	fmt.Println("Starting server on http://localhost:" + port)
	server.ListenAndServe()
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.GetAllChirpsByAscendingCreatedAt(r.Context())
	if err != nil {
		log.Printf("Error getting chirps: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var response []Chirp
	for _, chirp := range chirps {
		response = append(response, Chirp{
			ID:        chirp.ID.String(),
			UserId:    chirp.UserID.String(),
			Body:      chirp.Body,
			CreatedAt: chirp.CreatedAt.String(),
			UpdatedAt: chirp.UpdatedAt.String(),
		})
	}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type createChirpRequest struct {
		Body   string `json:"body"`
		UserId string `json:"user_id"`
	}
	decoder := json.NewDecoder(r.Body)
	req := createChirpRequest{}
	err := decoder.Decode(&req)
	if err != nil {
		log.Printf("Error decoding request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cleanChirp, err := validateChirp(req.Body)
	if err != nil {
		log.Printf("Error validating chirp: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		log.Printf("Error parsing user ID: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:        uuid.New(),
		UserID:    userID,
		Body:      cleanChirp,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := Chirp{
		ID:        chirp.ID.String(),
		UserId:    chirp.UserID.String(),
		Body:      chirp.Body,
		CreatedAt: chirp.CreatedAt.String(),
		UpdatedAt: chirp.UpdatedAt.String(),
	}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type createUserRequest struct {
		Email string `json:"email"`
	}
	type createUserResponse struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}
	decoder := json.NewDecoder(r.Body)
	user := createUserRequest{}
	err := decoder.Decode(&user)
	if err != nil {
		log.Printf("Error decoding request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userDB, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		Email:     user.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Printf("Error creating user: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := createUserResponse{
		ID:        userDB.ID.String(),
		Email:     userDB.Email,
		CreatedAt: userDB.CreatedAt.String(),
		UpdatedAt: userDB.UpdatedAt.String(),
	}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}

func validateChirp(chirp string) (string, error) {
	MAX_CHIRP_LENGTH := 140
	if len(chirp) > MAX_CHIRP_LENGTH {
		return "", fmt.Errorf("Chirp is too long")
	}
	cleanChirp := CleanChirpProfaneWords(chirp)
	return cleanChirp, nil
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
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}
	cfg.fileserveHits.Store(0)
	cfg.dbQueries.DeleteAllUsers(r.Context())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
