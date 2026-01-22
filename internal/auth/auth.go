package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})

	ss, err := token.SignedString([]byte(tokenSecret))
	return ss, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("Error parsing token: %v", err)
	}

	idString := claims.Subject

	realId, err := uuid.Parse(idString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("Error parsing result to id format: %v", err)
	}
	return realId, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if len(auth) == 0 {
		return "", fmt.Errorf("Authorization field not found in header")
	}

	stripped_token := strings.TrimSpace(auth)
	token_chunks := strings.Fields(stripped_token)
	if len(token_chunks) != 2 {
		return "", fmt.Errorf("Malformed authorization header")
	}

	if token_chunks[0] != "Bearer" {
		return "", fmt.Errorf("Malformed authorization header")
	}

	return token_chunks[1], nil
}

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	rand.Read(key)
	return hex.EncodeToString(key), nil
}

func GetAPIKey(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if len(auth) == 0 {
		return "", fmt.Errorf("Authorization field not found in header")
	}

	stripped_api_key := strings.TrimSpace(auth)
	key_chunks := strings.Fields(stripped_api_key)
	if len(key_chunks) != 2 {
		return "", fmt.Errorf("Malformed authorization header")
	}

	if key_chunks[0] != "ApiKey" {
		return "", fmt.Errorf("Malformed authorization header")
	}

	return key_chunks[1], nil
}
