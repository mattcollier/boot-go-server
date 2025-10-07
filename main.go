package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mattcollier/boot-go-server/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		w.Header().Add("Cache-Control", "no-cache")
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	text := fmt.Sprintf(`
<html>
<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
</body>
</html>`, cfg.fileserverHits.Load())
	w.Write([]byte(text))
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	if cfg.platform == "dev" {
		err := cfg.db.DeleteUsers(r.Context())
		if err != nil {
			log.Printf("Error deleting users: %s", err)
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(500)
			w.Write([]byte(`{"error":"Something went wrong"}`))
			return
		}
	}
	cfg.fileserverHits.Store(0)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	text := fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())
	w.Write([]byte(text))
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type payload struct {
		Email string `json:"email"`
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

	user, err := cfg.db.CreateUser(r.Context(), p.Email)
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

func main() {
	godotenv.Load()
	const filepathRoot = "."
	const port = "8080"
	dbURL := os.Getenv("DB_URL")

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("'PLATFORM' env must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("database connection error")
	}
	dbQueries := database.New(db)

	api := apiConfig{
		db:       dbQueries,
		platform: platform,
	}

	h := api.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot)))

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", h))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/chirps", api.handleChirps)
	mux.HandleFunc("GET /api/chirps", api.handleGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirp_id}", api.handleGetChirp)
	mux.HandleFunc("POST /api/users", api.handleCreateUser)
	mux.HandleFunc("GET /admin/metrics", api.handleMetrics)
	mux.HandleFunc("POST /admin/reset", api.resetMetrics)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func cleanMessage(s string) string {
	rValue := make([]string, 0)
	words := strings.Split(s, " ")
	for _, w := range words {
		t := strings.ToUpper(w)
		switch t {
		case "KERFUFFLE":
			fallthrough
		case "SHARBERT":
			fallthrough
		case "FORNAX":
			rValue = append(rValue, "****")
		default:
			rValue = append(rValue, w)
		}
	}
	return strings.Join(rValue, " ")
}
