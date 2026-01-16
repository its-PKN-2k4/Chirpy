package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNormalToken(t *testing.T) {
	userId := uuid.New()
	ss, err1 := MakeJWT(userId, "Vertigo", time.Hour)
	if err1 != nil {
		t.Fatalf("Error making a JWT token: %v", err1)
	}
	validatedId, err2 := ValidateJWT(ss, "Vertigo")
	if err2 != nil {
		t.Fatalf("Error validating JWT token: %v", err2)
	}

	if userId != validatedId {
		t.Errorf("initial id: %v, received id: %v", userId, validatedId)
	}
}

func TestExpiredToken(t *testing.T) {
	userId := uuid.New()
	ss, err1 := MakeJWT(userId, "Vertigo", -time.Hour)
	if err1 != nil {
		t.Fatalf("Error making a JWT token: %v", err1)
	}
	_, err2 := ValidateJWT(ss, "Vertigo")
	if err2 == nil {
		t.Fatalf("Error invalidating an expired JWT token")
	}
}

func TestWrongSecret(t *testing.T) {
	userId := uuid.New()
	ss, err1 := MakeJWT(userId, "secret1", -time.Hour)
	if err1 != nil {
		t.Fatalf("Error making a JWT token: %v", err1)
	}
	_, err2 := ValidateJWT(ss, "secret2")
	if err2 == nil {
		t.Fatalf("Error invalidating a JWT with mismatching secrets")
	}
}

func TestGarbageTokenString(t *testing.T) {
	userId := uuid.New()
	ss, err1 := MakeJWT(userId, "Vertigo", -time.Hour)
	if err1 != nil {
		t.Fatalf("Error making a JWT token: %v", err1)
	}
	_, err2 := ValidateJWT("Modified token"+ss, "Vertigo")
	if err2 == nil {
		t.Fatalf("Error invalidating a modified JWT token")
	}
}
