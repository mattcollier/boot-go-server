package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mattcollier/boot-go-server/internal/auth"
	"github.com/mattcollier/boot-go-server/internal/database"
)

type UserPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserData struct {
	Email          string
	HashedPassword string
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	userData, valid := validateUserPayload(w, r)
	if !valid {
		// errors have already been written
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          userData.Email,
		HashedPassword: stringToNullString(userData.HashedPassword),
	})
	if err != nil {
		log.Printf("Error creating user: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
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

func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"Invalid Authorization header"}`))
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"Invalid JWT"}`))
		return
	}

	user, valid := validateUserPayload(w, r)
	if !valid {
		// errors have already been written
		return
	}

	updatedUser, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userId,
		Email:          user.Email,
		HashedPassword: stringToNullString(user.HashedPassword),
	})

	if err != nil {
		log.Printf("Error updating user: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}

	jsonUser, err := json.Marshal(updatedUser)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		log.Printf("Error encoding user: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsonUser)
}

func validateUserPayload(w http.ResponseWriter, r *http.Request) (UserData, bool) {
	decoder := json.NewDecoder(r.Body)
	userPayload := UserPayload{}
	err := decoder.Decode(&userPayload)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		log.Printf("Error decoding message: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return UserData{}, false
	}

	// TODO: add more stringent password requirements
	if userPayload.Password == "" {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write([]byte(`{"error":"Password is required"}`))
		return UserData{}, false
	}

	hashedPassword, err := auth.HashPassword(userPayload.Password)

	if err != nil {
		log.Printf("Error hashing password: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return UserData{}, false
	}

	user := UserData{
		Email:          userPayload.Email,
		HashedPassword: hashedPassword,
	}
	return user, true
}
