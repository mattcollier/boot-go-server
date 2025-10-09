package main

import (
	"database/sql"
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
	jwtSecret      string
}

func main() {
	godotenv.Load()
	const filepathRoot = "."
	const port = "8080"
	dbURL := os.Getenv("DB_URL")

	jwtSecret := os.Getenv("JWT_TOKEN_SECRET")
	if jwtSecret == "" {
		log.Fatal("'JWT_TOKEN_SECRET' env must be set")
	}

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
		db:        dbQueries,
		platform:  platform,
		jwtSecret: jwtSecret,
	}

	h := api.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot)))

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", h))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/chirps", api.handleChirps)
	mux.HandleFunc("GET /api/chirps", api.handleGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirp_id}", api.handleGetChirp)
	mux.HandleFunc("POST /api/users", api.handleCreateUser)
	mux.HandleFunc("POST /api/login", api.handleLogin)
	mux.HandleFunc("POST /api/refresh", api.handleRefreshToken)
	mux.HandleFunc("POST /api/revoke", api.handleRevokeRefreshToken)
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

// stringToNullString is a helper function to convert a string to sql.NullString
func stringToNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "", // Valid is true if the string is not empty
	}
}
