package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"path"

	"github.com/google/uuid"
	"github.com/kerkox/chirpy-go/internal/auth"
)

func (cfg *apiConfig) handlerChirpDelete(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authentication is required", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authentication is required", err)
		return
	}

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

	if chirpDb.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You're not allowed to delete this chirp, it's not yours", nil)
		return
	}

	cfg.dbQueries.DeleteChirpById(r.Context(), chirpDb.ID)
	respondWithJSON(w, http.StatusNoContent, "")
}
