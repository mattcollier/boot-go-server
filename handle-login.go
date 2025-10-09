package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mattcollier/boot-go-server/internal/auth"
	"github.com/mattcollier/boot-go-server/internal/database"
)

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type loginDetails struct {
		Password string `json:"password"`
		Email    string `json:"email"`
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

	jwtTTL := time.Duration(time.Hour)
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, jwtTTL)
	if err != nil {
		log.Printf("Error in MakeJWT: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}

	// 60 days
	refreshTokenTTL := time.Duration(time.Hour * 24 * 60)
	now := time.Now()
	refreshToken, _ := auth.MakeRefreshToken()
	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		ExpiresAt: now.Add(refreshTokenTTL),
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
	})

	if err != nil {
		log.Printf("Database error: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}

	type redactedUser struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}
	ru := redactedUser{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
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
