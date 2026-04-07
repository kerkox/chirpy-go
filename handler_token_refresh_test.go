package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/kerkox/chirpy-go/internal/database"
)

func TestIsValidRefreshTokenToBeValid(t *testing.T) {

	refreshDb := database.RefreshToken{
		Token:     "123",
		ExpiresAt: time.Now().Add(time.Duration(1) * time.Hour),
		RevokedAt: sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
	}

	result := IsValidRefreshToken(refreshDb)
	if !result {
		t.Error("Should be valid")
	}
}

func TestIsValidRefreshTokenRevoked(t *testing.T) {

	revokedAt := sql.NullTime{
		Time:  time.Now().Add(time.Duration(-1) * time.Hour),
		Valid: true,
	}

	refreshDB := database.RefreshToken{
		Token:     "123",
		RevokedAt: revokedAt,
	}

	result := IsValidRefreshToken(refreshDB)

	if result {
		t.Error("Expected to be invalid because the token is revoked")
	}
}

func TestIsValidRefreshTokenExpired(t *testing.T) {

	refreshDB := database.RefreshToken{
		Token:     "123",
		ExpiresAt: time.Now().Add(time.Duration(-1) * time.Hour),
		RevokedAt: sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
	}

	result := IsValidRefreshToken(refreshDB)
	if result {
		t.Error("Expected to be invalid because the token is expired")
	}
}
