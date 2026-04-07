package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/kerkox/chirpy-go/internal/auth"
	"github.com/kerkox/chirpy-go/internal/database"
)

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid token", err)
		return
	}

	_, err = cfg.dbQueries.GetRerfreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithJSON(w, http.StatusNoContent, nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error with database", err)
		return
	}

	revokedAt := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	cfg.dbQueries.RevokeRefreshToken(r.Context(), database.RevokeRefreshTokenParams{
		Token:     refreshToken,
		RevokedAt: revokedAt,
	})
	respondWithJSON(w, http.StatusNoContent, nil)

}
