package main

import (
	"encoding/json"
	"net/http"

	"github.com/kerkox/chirpy-go/internal/auth"
	"github.com/kerkox/chirpy-go/internal/database"
)

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid token", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	userDB, err := cfg.dbQueries.GetUserById(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't fetch user", err)
		return
	}

	if userDB.ID != userID {
		respondWithError(w, http.StatusForbidden, "You can only update your own account", nil)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}
	_, err = cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		HashedPassword: hashedPassword,
		Email:          params.Email,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	type response struct {
		User
	}

	userResponse := response{
		User: User{
			ID:        userDB.ID,
			Email:     params.Email,
			CreatedAt: userDB.CreatedAt,
			UpdatedAt: userDB.UpdatedAt,
		},
	}
	respondWithJSON(w, http.StatusOK, userResponse)
}
