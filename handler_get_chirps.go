package main

import (
	"net/http"
	"sort"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	author := r.URL.Query().Get("author_id")
	if len(author) > 0 {
		authorID, err := uuid.Parse(author)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't parsed provided query parameter as ID", err)
			return
		}

		authorChirps, err := cfg.db.GetChirpsByUserID(r.Context(), authorID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't find chirps created by given user ID", err)
			return
		}

		returnChirps := []Chirp{}

		// Loop through database chirps array returned
		for _, dbChirp := range authorChirps {
			returnChirps = append(returnChirps, Chirp{
				ID:        dbChirp.ID,
				CreatedAt: dbChirp.CreatedAt,
				UpdatedAt: dbChirp.UpdatedAt,
				Body:      dbChirp.Body,
				UserID:    dbChirp.UserID,
			})
		}
		respondWithJSON(w, http.StatusOK, returnChirps)
		return
	}

	chirpsArray, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get all chirps", err)
		return
	}
	returnChirps := []Chirp{}

	// Loop through database chirps array returned
	for _, dbChirp := range chirpsArray {
		returnChirps = append(returnChirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		})
	}

	order := r.URL.Query().Get("sort")
	if order == "desc" {
		sort.Slice(returnChirps, func(i, j int) bool { return returnChirps[i].CreatedAt.After(returnChirps[j].CreatedAt) })
	}

	respondWithJSON(w, http.StatusOK, returnChirps)
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	idToFind, err1 := uuid.Parse(r.PathValue("param1"))
	if err1 != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode url parameters to string for internal use", err1)
		return
	}
	chirpByID, err2 := cfg.db.GetChirpByID(r.Context(), idToFind)
	if err2 != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find chirp that has the given ID", err2)
		return
	}
	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirpByID.ID,
		CreatedAt: chirpByID.CreatedAt,
		UpdatedAt: chirpByID.UpdatedAt,
		Body:      chirpByID.Body,
		UserID:    chirpByID.UserID,
	})
}
