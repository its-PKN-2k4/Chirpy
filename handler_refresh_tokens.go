package main

import (
	"net/http"
	"time"

	"example.com/m/internal/auth"
)

func (cfg *apiConfig) handlerRefreshJWT(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	grab_refresh_tkn, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No refresh token found", err)
		return
	}

	user_with_rtoken, err := cfg.db.GetUserByRefreshToken(r.Context(), grab_refresh_tkn)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No user with matching valid refresh token found", err)
		return
	}

	new_jwt, err := auth.MakeJWT(user_with_rtoken.ID, cfg.key, time.Duration((60*60)*time.Second))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate a new sign-in token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: new_jwt,
	})
}

func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	grab_refresh_tkn, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No login token found", err)
		return
	}

	err1 := cfg.db.RevokeRefreshToken(r.Context(), grab_refresh_tkn)
	if err1 != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke supplied refresh token", err1)
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
}
