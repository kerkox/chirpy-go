package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kerkox/chirpy-go/internal/auth"
)

func (cfg *apiConfig) handlerUsersLogin(w http.ResponseWriter, r *http.Request) {
	type loginRequest struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds *int   `json:"expires_in_seconds"`
	}
	type loginResponse struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Email     string `json:"email"`
		Token     string `json:"token"`
	}
	var req loginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password", err)
		return
	}

	ok, err := auth.CheckPasswordHash(req.Password, user.HashedPassword)
	if err != nil || !ok {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password", err)
		return
	}

	expiresIn := 3600 // default to 1 hour
	if req.ExpiresInSeconds != nil && *req.ExpiresInSeconds > 0 && *req.ExpiresInSeconds < 3600 {
		expiresIn = *req.ExpiresInSeconds
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Duration(expiresIn)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT", err)
		return
	}

	resp := loginResponse{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email:     user.Email,
		Token:     token,
	}
	respondWithJSON(w, http.StatusOK, resp)

}
