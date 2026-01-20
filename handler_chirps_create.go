package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"example.com/m/internal/auth"
	"example.com/m/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	bearer_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error obtaining sign-in token", err)
		return
	}

	id_from_token, err := auth.ValidateJWT(bearer_token, cfg.key)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error validating sign-in token", err)
		return
	}

	cleaned, err := validateChirp(params.Body)
	log.Printf("Decoded cleaned: body=%q", cleaned)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: id_from_token,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    id_from_token,
	})
}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(body, badWords)
	return cleaned, nil
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	idChirpToFind, err1 := uuid.Parse(r.PathValue("chirpID"))
	if err1 != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode url parameters to string for internal use", err1)
		return
	}

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error obtaining sign-in token", err)
		return
	}

	idFromToken, err := auth.ValidateJWT(bearerToken, cfg.key)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error validating sign-in token", err)
		return
	}

	chirpByID, err2 := cfg.db.GetChirpByID(r.Context(), idChirpToFind)
	if err2 != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find chirp that has the given ID", err2)
		return
	}

	if chirpByID.UserID.String() != idFromToken.String() {
		respondWithError(w, http.StatusForbidden, "You are not allowed to perform modifications to this chirp", nil)
		return
	}

	err3 := cfg.db.DeleteChirpWithID(r.Context(), chirpByID.ID)
	if err3 != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find chirp with parsed ID for deletion", err3)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
