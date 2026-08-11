package main

import (
	"net/http"
	"time"

	"github.com/csrrmrvll/chirpy/internal/auth"
	"github.com/csrrmrvll/chirpy/internal/database"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		RefreshToken string `json:"refresh_token"`
	}
	type response struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No Authorization header provided", nil)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	if refreshToken.ExpiresAt.Before(time.Now().UTC()) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token has expired", nil)
		return
	}

	if refreshToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Refresh token has been revoked", nil)
		return
	}

	newAccessToken, err := auth.MakeJWT(
		refreshToken.UserID,
		cfg.jwtSecret,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create new access JWT", err)
		return
	}

	newRefreshToken, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     newAccessToken,
		UserID:    refreshToken.UserID,
		ExpiresAt: time.Now().UTC().Add(60 * 24 * time.Hour),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create new refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken.Token,
	})
}
