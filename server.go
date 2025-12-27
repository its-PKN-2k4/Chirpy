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

	"example.com/m/internal/database"
	_ "github.com/lib/pq"
)

// Struct to store stateful (counting requests to an endpoint)
type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

// Middleware to alter/record state and process requests
func (cfg *apiConfig) middlewareMetricsIncrement(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

// Function to handle POST request and return stats as json
func JsonHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type returnVals struct {
		Cleaned_body string `json:"cleaned_body"`
	}

	resContent := returnVals{}

	w.Header().Set("Content-Type", "application/json")
	if len(params.Body) <= 140 {
		to_filter := []string{"kerfuffle", "sharbert", "fornax"}
		string_splits := strings.Split(params.Body, " ")
		for i := 0; i < len(string_splits); i++ {
			for j := 0; j < len(to_filter); j++ {
				if strings.ToLower(string_splits[i]) == to_filter[j] {
					string_splits[i] = "****"
				}
			}
		}
		resContent.Cleaned_body = strings.Join(string_splits, " ")
		w.WriteHeader(http.StatusOK)
	} else {
		resContent.Cleaned_body = ""
		w.WriteHeader(http.StatusBadRequest)
	}
	data, err := json.Marshal(resContent)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader((http.StatusInternalServerError))
		return
	}
	w.Write(data)

}

func main() {
	err1 := godotenv.Load()
	if err1 != nil {
		log.Printf("Error loading godotenv lib: %s", err1)
	}
	dbURL := os.Getenv("DB_URL")
	db, err2 := sql.Open("postgres", dbURL)
	if err2 != nil {
		log.Printf("Error opening connection to SQL database: %s", err2)
	}
	dbQueries := database.New(db)

	mux := http.NewServeMux()
	apiCFG := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
	}
	// Instance of stateful struct
	mux.Handle("/app/", apiCFG.middlewareMetricsIncrement(http.StripPrefix("/app", http.FileServer(http.Dir("."))))) // Handler for /app endpoint

	// Handler for /api/healthz endpoint
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Handler for /admin/metrics endpoint
	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		hitCount := fmt.Sprintf(`
		<html>
  			<body>
    			<h1>Welcome, Chirpy Admin</h1>
    			<p>Chirpy has been visited %d times!</p>
  			</body>
		</html>`, apiCFG.fileserverHits.Load())
		w.Write([]byte(hitCount))
	})

	// Handler for /admin/reset endpoint
	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}

		w.WriteHeader(http.StatusOK)
		apiCFG.fileserverHits.Swap(0)
	})

	// Handler for /api/validate_chirp endpoint
	mux.HandleFunc("POST /api/validate_chirp", JsonHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	srv.ListenAndServe()
}
