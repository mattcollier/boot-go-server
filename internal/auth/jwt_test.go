package auth

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMakeJWT_ValidTokenAndClaims(t *testing.T) {
	userID := uuid.New()
	secret := "super-secret"
	expiresIn := 90 * time.Minute

	tokenStr, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}
	if tokenStr == "" {
		t.Fatalf("MakeJWT returned empty token")
	}

	claims := &jwt.RegisteredClaims{}
	keyfunc := func(token *jwt.Token) (interface{}, error) {
		// Enforce HS256
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
		}
		return []byte(secret), nil
	}

	parsedToken, err := jwt.ParseWithClaims(tokenStr, claims, keyfunc)
	if err != nil {
		t.Fatalf("ParseWithClaims error: %v", err)
	}
	if !parsedToken.Valid {
		t.Fatalf("parsed token is not valid")
	}

	// Static claims
	if claims.Issuer != "chirpy" {
		t.Errorf("Issuer = %q, want %q", claims.Issuer, "chirpy")
	}
	if claims.Subject != userID.String() {
		t.Errorf("Subject = %q, want %q", claims.Subject, userID.String())
	}

	// Time claims (allow a small clock skew)
	const maxSkew = 2 * time.Second
	now := time.Now()

	iat := claims.IssuedAt.Time
	if iat.Before(now.Add(-maxSkew)) || iat.After(now.Add(maxSkew)) {
		t.Errorf("IssuedAt %v not within %v of now %v", iat, maxSkew, now)
	}

	exp := claims.ExpiresAt.Time
	expected := iat.Add(expiresIn)
	if exp.Before(expected.Add(-maxSkew)) || exp.After(expected.Add(maxSkew)) {
		t.Errorf("ExpiresAt %v not within %v of iat+expiresIn (%v)", exp, maxSkew, expected)
	}
}

func TestMakeJWT_WrongSecretFailsVerification(t *testing.T) {
	userID := uuid.New()
	rightSecret := "right-secret"
	wrongSecret := "wrong-secret"
	expiresIn := time.Minute

	tokenStr, err := MakeJWT(userID, rightSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	_, err = jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
		}
		return []byte(wrongSecret), nil
	})
	if err == nil {
		t.Fatalf("expected signature validation error, got nil")
	}
	if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		t.Errorf("expected ErrTokenSignatureInvalid, got %v", err)
	}
}
