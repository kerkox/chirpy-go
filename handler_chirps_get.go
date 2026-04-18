package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"path"

	"github.com/google/uuid"
	"github.com/kerkox/chirpy-go/internal/database"
)

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	query_author_id := query.Get("author_id")

	var dbChirps []database.Chirp
	var err error

	if query_author_id != "" {
		authorId, err := uuid.Parse(query_author_id)

		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Author id invalid not a uuid", err)
			return
		}
		dbChirps, err = cfg.dbQueries.GetChirpsByAuthorId(r.Context(), authorId)

	} else {
		dbChirps, err = cfg.dbQueries.GetChirps(r.Context())
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps", err)
		return
	}

	indexStart := 0
	incrementIndex := 1
	sortDirectionParam := r.URL.Query().Get("sort")

	if sortDirectionParam == "desc" {
		indexStart = len(dbChirps) - 1
		incrementIndex = -1
	}

	chirps := make([]Chirp, len(dbChirps))

	for _, dbChirp := range dbChirps {
		chirps[indexStart] = Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID:    dbChirp.UserID,
			Body:      dbChirp.Body,
		}
		indexStart += incrementIndex
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerChirpGet(w http.ResponseWriter, r *http.Request) {
	chirpId := path.Base(r.URL.Path)
	if chirpId == "" {
		respondWithError(w, http.StatusBadRequest, "Chirp Id is required", nil)
		return
	}
	parseChirpId, err := uuid.Parse(chirpId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}
	chirpDb, err := cfg.dbQueries.GetChirpById(r.Context(), parseChirpId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Chirp Not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't get a chirp", err)
		log.Printf("Error getting chirp: %v", err)
		return
	}
	chirp := Chirp{
		ID:        chirpDb.ID,
		CreatedAt: chirpDb.CreatedAt,
		UpdatedAt: chirpDb.UpdatedAt,
		UserID:    chirpDb.UserID,
		Body:      chirpDb.Body,
	}

	respondWithJSON(w, http.StatusOK, chirp)
}
