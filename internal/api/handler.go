package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jamespacheco-dev/pico-api/internal/game"
	"github.com/jamespacheco-dev/pico-api/internal/metrics"
)

type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

// --- Request types ---

type createGameRequest struct {
	Mode         game.Mode       `json:"mode"`
	Length       int             `json:"length"`
	AllowRepeats bool            `json:"allow_repeats"`
	Difficulty   game.Difficulty `json:"difficulty"`
}

type feedbackRequest struct {
	Pico  int `json:"pico"`
	Fermi int `json:"fermi"`
}

type createGuessRequest struct {
	Guess    string           `json:"guess"`
	Feedback *feedbackRequest `json:"feedback"`
}

type rollbackRequest struct {
	ToGuess int `json:"to_guess"`
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}

func selectorFor(d game.Difficulty) game.Selector {
	switch d {
	case game.DifficultyMedium:
		return game.MediumSelector{}
	case game.DifficultyHard:
		return game.HardSelector{}
	default:
		return game.RandomSelector{}
	}
}

// --- Handlers ---

func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var req createGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}

	if req.Mode != game.ModePlayerGuesses && req.Mode != game.ModeComputerGuesses {
		writeError(w, http.StatusBadRequest, "invalid_mode", "mode must be player_guesses or computer_guesses")
		return
	}

	if req.Length == 0 {
		req.Length = 3
	}
	if req.Difficulty == "" {
		req.Difficulty = game.DifficultyEasy
	}

	cfg := game.Config{Length: req.Length, AllowRepeats: req.AllowRepeats}
	g, err := game.NewGame(cfg, req.Mode, req.Difficulty, selectorFor(req.Difficulty))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create game")
		return
	}

	if err := h.store.Create(g); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to store game")
		return
	}

	metrics.GamesCreated.WithLabelValues(string(req.Mode), fmt.Sprintf("%d", req.Length)).Inc()
	metrics.ActiveGames.Inc()

	writeJSON(w, http.StatusCreated, g)
}

func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	g, err := h.store.Get(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "game not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to retrieve game")
		return
	}
	writeJSON(w, http.StatusOK, g)
}

func (h *Handler) CreateGuess(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	g, err := h.store.Get(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "game not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to retrieve game")
		return
	}

	var req createGuessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}

	switch g.Mode {
	case game.ModePlayerGuesses:
		if req.Guess == "" {
			writeError(w, http.StatusBadRequest, "missing_guess", "guess is required for player_guesses mode")
			return
		}
		if _, err := g.ApplyGuess(req.Guess); err != nil {
			writeGuessError(w, err)
			return
		}

	case game.ModeComputerGuesses:
		if req.Feedback == nil {
			writeError(w, http.StatusBadRequest, "missing_feedback", "feedback is required for computer_guesses mode")
			return
		}
		fb := game.Feedback{Pico: req.Feedback.Pico, Fermi: req.Feedback.Fermi}
		start := time.Now()
		_, err := g.ApplyFeedback(fb)
		if err != nil {
			writeGuessError(w, err)
			return
		}
		metrics.AIGuessLatency.WithLabelValues(string(g.Difficulty)).Observe(time.Since(start).Seconds())
	}

	if err := h.store.Save(g); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to save game")
		return
	}

	if g.IsComplete() {
		mode, diff := string(g.Mode), string(g.Difficulty)
		metrics.GamesCompleted.WithLabelValues(mode, diff).Inc()
		metrics.GuessesPerGame.WithLabelValues(mode, diff).Observe(float64(len(g.Guesses)))
		metrics.ActiveGames.Dec()
	}

	writeJSON(w, http.StatusOK, g)
}

func (h *Handler) Rollback(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	g, err := h.store.Get(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "game not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to retrieve game")
		return
	}

	var req rollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}

	if err := g.Rollback(req.ToGuess); err != nil {
		switch {
		case errors.Is(err, game.ErrWrongMode):
			writeError(w, http.StatusUnprocessableEntity, "rollback_not_supported", "rollback is only available in computer_guesses mode")
		case errors.Is(err, game.ErrOutOfRange):
			writeError(w, http.StatusBadRequest, "invalid_guess_number", "to_guess is out of range")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		}
		return
	}

	metrics.Rollbacks.Inc()

	if err := h.store.Save(g); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to save game")
		return
	}

	writeJSON(w, http.StatusOK, g)
}

func writeGuessError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, game.ErrWrongMode):
		writeError(w, http.StatusUnprocessableEntity, "wrong_mode", "this operation is not valid for the current game mode")
	case errors.Is(err, game.ErrGameComplete):
		writeError(w, http.StatusUnprocessableEntity, "game_complete", "the game is already complete")
	case errors.Is(err, game.ErrContradictory):
		writeError(w, http.StatusUnprocessableEntity, "contradictory_feedback", "no valid candidates remain — check your feedback and try again")
	default:
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
	}
}
