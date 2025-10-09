package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mattcollier/boot-go-server/internal/auth"
	"github.com/mattcollier/boot-go-server/internal/database"
)

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	p := payload{}
	err := decoder.Decode(&p)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		log.Printf("Error decoding message: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}

	// TODO: add more stringent password requirements
	if p.Password == "" {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write([]byte(`{"error":"Password is required"}`))
		return
	}

	hashedPassword, err := auth.HashPassword(p.Password)

	if err != nil {
		log.Printf("Error hashing password: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          p.Email,
		HashedPassword: stringToNullString(hashedPassword),
	})
	if err != nil {

	}

	jsonUser, err := json.Marshal(user)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		log.Printf("Error decoding message: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(jsonUser)
}
