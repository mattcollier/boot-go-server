package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mattcollier/boot-go-server/internal/database"
)

func (cfg *apiConfig) handleChirps(w http.ResponseWriter, r *http.Request) {
	type messageBody struct {
		Body   string        `json:"body"`
		UserID uuid.NullUUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	mb := messageBody{}
	err := decoder.Decode(&mb)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		log.Printf("Error decoding message: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}
	if len(mb.Body) > 140 {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write([]byte(`{"error":"Chirp is too long"}`))
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   mb.Body,
		UserID: mb.UserID,
	})
	if err != nil {

	}

	type myChirp struct {
		ID        uuid.UUID     `json:"id"`
		CreatedAt time.Time     `json:"created_at"`
		UpdatedAt time.Time     `json:"updated_at"`
		Body      string        `json:"body"`
		UserID    uuid.NullUUID `json:"user_id"`
	}

	// move data into a properly tagged struct
	taggedChirp := myChirp{}
	taggedChirp.ID = chirp.ID
	taggedChirp.CreatedAt = chirp.CreatedAt
	taggedChirp.UpdatedAt = chirp.UpdatedAt
	taggedChirp.Body = chirp.Body
	taggedChirp.UserID = chirp.UserID

	jsonChirp, err := json.Marshal(taggedChirp)
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
	w.Write(jsonChirp)
}
