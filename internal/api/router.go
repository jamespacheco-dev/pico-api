package api

import "net/http"

func NewRouter(store Store) http.Handler {
	mux := http.NewServeMux()
	h := NewHandler(store)

	mux.HandleFunc("POST /games", h.CreateGame)
	mux.HandleFunc("GET /games/{id}", h.GetGame)
	mux.HandleFunc("POST /games/{id}/guesses", h.CreateGuess)
	mux.HandleFunc("POST /games/{id}/rollback", h.Rollback)

	return mux
}
