package auth

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func makeSignedToken(t *testing.T, alg jwt.SigningMethod, secret string, claims jwt.RegisteredClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(alg, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func baseClaims(userID uuid.UUID, now time.Time, ttl time.Duration) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Issuer:    "chirpy",
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}
}

func TestValidateJWT_Success(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	secret := "super-secret"

	tokenStr := makeSignedToken(t, jwt.SigningMethodHS256, secret, baseClaims(userID, now, time.Hour))

	got, err := ValidateJWT(tokenStr, secret)
	if err != nil {
		t.Fatalf("ValidateJWT returned error: %v", err)
	}
	if got != userID {
		t.Fatalf("ValidateJWT userID = %v, want %v", got, userID)
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	rightSecret := "right"
	wrongSecret := "wrong"

	tokenStr := makeSignedToken(t, jwt.SigningMethodHS256, rightSecret, baseClaims(userID, now, time.Hour))

	got, err := ValidateJWT(tokenStr, wrongSecret)
	if err == nil {
		t.Fatalf("expected error, got nil (userID=%v)", got)
	}
	// Signature invalid should bubble up from jwt/v5
	if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		t.Fatalf("expected ErrTokenSignatureInvalid, got %v", err)
	}
	if got != uuid.Nil {
		t.Fatalf("expected uuid.Nil on error, got %v", got)
	}
}

func TestValidateJWT_WrongAlg(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	secret := "secret"

	// Sign with HS512 but validator enforces HS256 in keyfunc
	tokenStr := makeSignedToken(t, jwt.SigningMethodHS512, secret, baseClaims(userID, now, time.Hour))

	got, err := ValidateJWT(tokenStr, secret)
	if err == nil {
		t.Fatalf("expected error for unexpected signing method, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected signing method") {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != uuid.Nil {
		t.Fatalf("expected uuid.Nil on error, got %v", got)
	}
}

func TestValidateJWT_WrongIssuer(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	secret := "secret"

	claims := baseClaims(userID, now, time.Hour)
	claims.Issuer = "not-chirpy"

	tokenStr := makeSignedToken(t, jwt.SigningMethodHS256, secret, claims)

	got, err := ValidateJWT(tokenStr, secret)
	if err == nil {
		t.Fatalf("expected issuer error, got nil")
	}
	if !strings.Contains(err.Error(), "issuer =") {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != uuid.Nil {
		t.Fatalf("expected uuid.Nil on error, got %v", got)
	}
}

func TestValidateJWT_InvalidSubject(t *testing.T) {
	now := time.Now()
	secret := "secret"

	claims := baseClaims(uuid.New(), now, time.Hour)
	claims.Subject = "not-a-uuid"

	tokenStr := makeSignedToken(t, jwt.SigningMethodHS256, secret, claims)

	got, err := ValidateJWT(tokenStr, secret)
	if err == nil {
		t.Fatalf("expected invalid subject error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid subject") {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != uuid.Nil {
		t.Fatalf("expected uuid.Nil on error, got %v", got)
	}
}

func TestValidateJWT_Expired(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	secret := "secret"

	claims := baseClaims(userID, now.Add(-2*time.Hour), -time.Hour) // exp in the past
	tokenStr := makeSignedToken(t, jwt.SigningMethodHS256, secret, claims)

	got, err := ValidateJWT(tokenStr, secret)
	if err == nil {
		t.Fatalf("expected expiration error, got nil")
	}
	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
	if got != uuid.Nil {
		t.Fatalf("expected uuid.Nil on error, got %v", got)
	}
}

func TestValidateJWT_NotYetValid(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	secret := "secret"

	claims := baseClaims(userID, now, time.Hour)
	claims.NotBefore = jwt.NewNumericDate(now.Add(30 * time.Second)) // not valid yet

	tokenStr := makeSignedToken(t, jwt.SigningMethodHS256, secret, claims)

	got, err := ValidateJWT(tokenStr, secret)
	if err == nil {
		t.Fatalf("expected not-valid-yet error, got nil")
	}
	if !errors.Is(err, jwt.ErrTokenNotValidYet) {
		t.Fatalf("expected ErrTokenNotValidYet, got %v", err)
	}
	if got != uuid.Nil {
		t.Fatalf("expected uuid.Nil on error, got %v", got)
	}
}

func TestValidateJWT_MalformedToken(t *testing.T) {
	secret := "secret"
	malformed := "definitely-not-a-jwt" // lacks the three-part structure

	got, err := ValidateJWT(malformed, secret)
	if err == nil {
		t.Fatalf("expected malformed token error, got nil")
	}
	if !errors.Is(err, jwt.ErrTokenMalformed) {
		t.Fatalf("expected ErrTokenMalformed, got %v", err)
	}
	if got != uuid.Nil {
		t.Fatalf("expected uuid.Nil on error, got %v", got)
	}
}
