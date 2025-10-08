package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mattcollier/boot-go-server/internal/auth"
)

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type loginDetails struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds *int   `json:"expires_in_seconds"` // optional
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	ld := loginDetails{}
	err := decoder.Decode(&ld)
	if err != nil {
		log.Printf("Error decoding message: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}

	ttl := time.Duration(time.Hour)
	if ld.ExpiresInSeconds != nil {
		// do not allow values greater than an hour
		if *ld.ExpiresInSeconds <= 0 || *ld.ExpiresInSeconds > 3600 {
			// use the default
		} else {
			ttl = time.Duration(*ld.ExpiresInSeconds) * time.Second
		}
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), ld.Email)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(401)
			w.Write([]byte(`{"error":"Incorrect email or password"}`))
		} else {
			log.Printf("Database error: %s", err)
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(500)
			w.Write([]byte(`{"error":"Something went wrong"}`))
		}

		return
	}
	passwordValid, err := auth.CheckPasswordHash(ld.Password, user.HashedPassword.String)

	if err != nil {
		log.Printf("Error validating password: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}

	if !passwordValid {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"Incorrect email or password"}`))
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, ttl)
	if err != nil {
		log.Printf("Error in MakeJWT: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}

	type redactedUser struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}
	ru := redactedUser{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	}

	jsonRedacted, err := json.Marshal(ru)

	if err != nil {
		log.Printf("Error encoding response: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsonRedacted)
}
