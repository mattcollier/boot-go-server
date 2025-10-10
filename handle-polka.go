package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/mattcollier/boot-go-server/internal/auth"
	"github.com/mattcollier/boot-go-server/internal/database"
)

type PolkaWebookPayload struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) handlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"Invalid Authorization header"}`))
		return
	}

	if apiKey != cfg.polkaAPIKey {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(401)
		return
	}

	decoder := json.NewDecoder(r.Body)
	polkaWebookPayload := PolkaWebookPayload{}
	err = decoder.Decode(&polkaWebookPayload)
	if err != nil {
		log.Printf("Error decoding message: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}

	if polkaWebookPayload.Event != "user.upgraded" {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(204)
		return
	}

	userId, err := uuid.Parse(polkaWebookPayload.Data.UserID)
	if err != nil {
		log.Printf("Invalid User ID: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write([]byte(`{"error":"Invalid User ID"}`))
		return
	}

	_, err = cfg.db.UpdateIsChirpyRed(r.Context(), database.UpdateIsChirpyRedParams{
		ID:          userId,
		IsChirpyRed: true,
	})

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(404)
			return
		}
		log.Printf("Error updating user: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(204)
}
