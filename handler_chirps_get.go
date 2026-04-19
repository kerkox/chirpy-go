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

	sortDirection := query.Get("sort")
	if sortDirection != "desc" {
		sortDirection = "asc"
	}

	var dbChirps []database.Chirp
	var err error

	if query_author_id != "" {
		authorId, err := uuid.Parse(query_author_id)

		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Author id invalid not a uuid", err)
			return
		}
		dbChirps, err = cfg.dbQueries.GetChirpsByAuthorId(r.Context(), database.GetChirpsByAuthorIdParams{
			UserID:        authorId,
			SortDirection: sortDirection,
		})

	} else {
		dbChirps, err = cfg.dbQueries.GetChirps(r.Context(), sortDirection)
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps", err)
		return
	}

	chirps := make([]Chirp, len(dbChirps))

	for i, dbChirp := range dbChirps {
		chirps[i] = Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID:    dbChirp.UserID,
			Body:      dbChirp.Body,
		}
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
