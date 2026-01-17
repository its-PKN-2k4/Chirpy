package main

import (
	"encoding/json"
	"net/http"
	"time"

	"example.com/m/internal/auth"
	"example.com/m/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't encrypt password", err)
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		HashedPassword: hashedPassword,
		Email:          params.Email})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	})
}

func (cfg *apiConfig) handlerUsersLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
		//ExpiresInSeconds *int   `json:"expires_in_seconds"`
	}
	type response struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	const maxExpirySeconds = 60 * 60

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	grab_user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, grab_user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
	if !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	} else {
		created_token, err := auth.MakeJWT(grab_user.ID, cfg.key, time.Duration(maxExpirySeconds)*time.Second)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't generate sign-in token", err)
			return
		}

		raw_refresh_tkn, err := auth.MakeRefreshToken()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't generate refresh token", err)
			return
		}

		refresh_tkn, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
			Token:  raw_refresh_tkn,
			UserID: grab_user.ID,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't generate refresh token", err)
			return
		}

		respondWithJSON(w, http.StatusOK, response{
			ID:           grab_user.ID,
			CreatedAt:    grab_user.CreatedAt,
			UpdatedAt:    grab_user.UpdatedAt,
			Email:        grab_user.Email,
			Token:        created_token,
			RefreshToken: refresh_tkn.Token,
		})
	}
}
