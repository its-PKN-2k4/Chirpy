package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// Struct to store stateful (counting requests to an endpoint)
type apiConfig struct {
	fileserverHits atomic.Int32
}

// Middleware to alter/record state and process requests
func (cfg *apiConfig) middlewareMetricsIncrement(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	apiCFG := apiConfig{}                                                                                            // Instance of stateful struct
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

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	srv.ListenAndServe()
}
