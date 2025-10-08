package auth

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		// Also fixed dates can be used for the NumericDate
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Issuer:    "chirpy",
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		log.Printf("Error signing jwt: %s", err)
		return "", err
	}
	// fmt.Println(ss, err)
	return ss, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	keyfunc := func(token *jwt.Token) (interface{}, error) {
		// Enforce HS256
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
		}
		return []byte(tokenSecret), nil
	}
	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, keyfunc)
	if err != nil {
		log.Printf("ParseWithClaims error: %v", err)
		return uuid.Nil, err
	}
	if !parsedToken.Valid {
		log.Printf("parsed token is not valid")
		return uuid.Nil, fmt.Errorf("parsed token is not valid")
	}

	// Static claims
	if claims.Issuer != "chirpy" {
		log.Printf("Issuer = %q, want %q", claims.Issuer, "chirpy")
		return uuid.Nil, fmt.Errorf("issuer = %q, want %q", claims.Issuer, "chirpy")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid subject")
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	a := headers.Get("Authorization")

	parts := strings.Split(a, " ")

	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header")
	}

	return parts[1], nil
}
