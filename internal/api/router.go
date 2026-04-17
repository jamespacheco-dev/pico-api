package api

import (
	"net/http"

	"github.com/jamespacheco-dev/pico-api/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(store Store) http.Handler {
	mux := http.NewServeMux()
	h := NewHandler(store)

	mux.HandleFunc("POST /games", metrics.Middleware("POST /games", h.CreateGame))
	mux.HandleFunc("GET /games/{id}", metrics.Middleware("GET /games/{id}", h.GetGame))
	mux.HandleFunc("POST /games/{id}/guesses", metrics.Middleware("POST /games/{id}/guesses", h.CreateGuess))
	mux.HandleFunc("POST /games/{id}/rollback", metrics.Middleware("POST /games/{id}/rollback", h.Rollback))
	mux.Handle("GET /metrics", promhttp.Handler())

	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
