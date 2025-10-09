package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mattcollier/boot-go-server/internal/auth"
)

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Invalid Authorization header"}`))
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(401)
		return
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(401)
		return
	}

	jwtTTL := time.Duration(time.Hour)
	newToken, errJwt := auth.MakeJWT(refreshToken.UserID.UUID, cfg.jwtSecret, jwtTTL)
	if errJwt != nil {
		log.Printf("Error in MakeJWT: %s", errJwt)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"token":"%s"}`, newToken)
}
