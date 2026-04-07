package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/kerkox/chirpy-go/internal/auth"
	"github.com/kerkox/chirpy-go/internal/database"
)

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	type refreshResponse struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid token", err)
		return
	}

	refreshTokenDB, err := cfg.dbQueries.GetRerfreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "Error Unauthorized", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error with database", err)
		return
	}

	expiresIn := 3600
	if IsValidRefreshToken(refreshTokenDB) {
		token, err := auth.MakeJWT(refreshTokenDB.UserID, cfg.jwtSecret, time.Duration(expiresIn)*time.Second)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT", err)
			return
		}
		response := refreshResponse{
			Token: token,
		}
		respondWithJSON(w, http.StatusOK, response)
		return
	}
	respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", nil)
}

func IsValidRefreshToken(refreshToken database.RefreshToken) bool {
	if refreshToken.RevokedAt.Valid {
		return false
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		return false
	}

	return true
}
