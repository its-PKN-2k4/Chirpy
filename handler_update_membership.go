package main

import (
	"encoding/json"

	"net/http"

	"example.com/m/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUpdateMembership(w http.ResponseWriter, r *http.Request) {
	type payload struct {
		UserID string `json:"user_id"`
	}

	type parameters struct {
		Event string  `json:"event"`
		Data  payload `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	api, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get API key to authenticate", err)
		return
	}

	if cfg.api != api {
		respondWithError(w, http.StatusUnauthorized, "Invalid API key provided", err)
		return
	}

	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	idToUpgrade, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't extract user_id from request", err)
		return
	}

	err1 := cfg.db.UpdateUserMembershipByID(r.Context(), idToUpgrade)
	if err1 != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find user with provided id", err)
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
