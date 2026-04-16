package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/kerkox/chirpy-go/internal/auth"
	"github.com/kerkox/chirpy-go/internal/database"
)

func (cfg *apiConfig) handlerPolkWebhookUserUpgraded(w http.ResponseWriter, r *http.Request) {

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error apiKey not found", err)
		return
	}

	if apiKey != cfg.apiKey {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error apiKey: %s not valid", apiKey), err)
		return
	}

	type requestPolka struct {
		Event string `json:"event"`
		Data  struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}

	var req requestPolka
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request", err)
		return
	}

	if req.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	userId, err := uuid.Parse(req.Data.UserId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error parsing the UserId", err)
		return
	}

	user, err := cfg.dbQueries.GetUserById(r.Context(), userId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, fmt.Sprintf("User with id: %s not found", userId), err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error with DB", err)
	}

	if user.IsChirpyRed {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	cfg.dbQueries.UpdateUserIsChirpyRed(r.Context(), database.UpdateUserIsChirpyRedParams{
		ID:          userId,
		IsChirpyRed: true,
	})

	respondWithJSON(w, http.StatusNoContent, nil)

}
