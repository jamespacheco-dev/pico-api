package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jamespacheco-dev/pico-api/internal/game"
)

// mockStore implements Store with controllable behaviour for testing.
type mockStore struct {
	games     map[string]*game.Game
	createErr error
	getErr    error
	saveErr   error
}

func newMockStore() *mockStore {
	return &mockStore{games: make(map[string]*game.Game)}
}

func (s *mockStore) Create(g *game.Game) error {
	if s.createErr != nil {
		return s.createErr
	}
	s.games[g.ID] = g
	return nil
}

func (s *mockStore) Get(id string) (*game.Game, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	g, ok := s.games[id]
	if !ok {
		return nil, ErrNotFound
	}
	return g, nil
}

func (s *mockStore) Save(g *game.Game) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.games[g.ID] = g
	return nil
}

func (s *mockStore) Sweep(_ time.Duration) int { return 0 }

// --- Helpers ---

func newTestRouter(store Store) http.Handler {
	return NewRouter(store)
}

func post(t *testing.T, router http.Handler, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func get(t *testing.T, router http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func decodeGame(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var g map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&g); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return g
}

// --- CreateGame ---

func TestCreateGame_PlayerGuesses(t *testing.T) {
	store := newMockStore()
	router := newTestRouter(store)

	rr := post(t, router, "/games", map[string]any{
		"mode": "player_guesses",
	})

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rr.Code)
	}
	g := decodeGame(t, rr)
	if g["mode"] != "player_guesses" {
		t.Errorf("mode = %v, want player_guesses", g["mode"])
	}
	if g["id"] == "" {
		t.Error("id should be set")
	}
	if g["current_guess"] != nil {
		t.Error("current_guess should not be set for player_guesses mode")
	}
}

func TestCreateGame_ComputerGuesses(t *testing.T) {
	store := newMockStore()
	router := newTestRouter(store)

	rr := post(t, router, "/games", map[string]any{
		"mode": "computer_guesses",
	})

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rr.Code)
	}
	g := decodeGame(t, rr)
	if g["current_guess"] == nil || g["current_guess"] == "" {
		t.Error("current_guess should be set for computer_guesses mode")
	}
}

func TestCreateGame_DefaultLength(t *testing.T) {
	store := newMockStore()
	router := newTestRouter(store)

	rr := post(t, router, "/games", map[string]any{"mode": "player_guesses"})
	g := decodeGame(t, rr)

	cfg := g["config"].(map[string]any)
	if cfg["length"].(float64) != 3 {
		t.Errorf("default length = %v, want 3", cfg["length"])
	}
}

func TestCreateGame_InvalidMode(t *testing.T) {
	store := newMockStore()
	router := newTestRouter(store)

	rr := post(t, router, "/games", map[string]any{"mode": "not_a_mode"})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestCreateGame_MissingMode(t *testing.T) {
	store := newMockStore()
	router := newTestRouter(store)

	rr := post(t, router, "/games", map[string]any{})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

// --- GetGame ---

func TestGetGame_Found(t *testing.T) {
	store := newMockStore()
	router := newTestRouter(store)

	rr := post(t, router, "/games", map[string]any{"mode": "player_guesses"})
	id := decodeGame(t, rr)["id"].(string)

	rr = get(t, router, "/games/"+id)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestGetGame_NotFound(t *testing.T) {
	store := newMockStore()
	router := newTestRouter(store)

	rr := get(t, router, "/games/does-not-exist")
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestGetGame_SecretNotExposed(t *testing.T) {
	store := newMockStore()
	router := newTestRouter(store)

	rr := post(t, router, "/games", map[string]any{"mode": "player_guesses"})
	id := decodeGame(t, rr)["id"].(string)

	rr = get(t, router, "/games/"+id)
	g := decodeGame(t, rr)
	if _, ok := g["secret"]; ok {
		t.Error("secret should not be exposed in API response")
	}
}

// --- CreateGuess (player_guesses) ---

func TestCreateGuess_PlayerGuesses_InvalidLength(t *testing.T) {
	store := newMockStore()
	router := newTestRouter(store)

	rr := post(t, router, "/games", map[string]any{"mode": "player_guesses"})
	id := decodeGame(t, rr)["id"].(string)

	rr = post(t, router, "/games/"+id+"/guesses", map[string]any{"guess": "12"})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestCreateGuess_PlayerGuesses_MissingGuess(t *testing.T) {
	store := newMockStore()
	router := newTestRouter(store)

	rr := post(t, router, "/games", map[string]any{"mode": "player_guesses"})
	id := decodeGame(t, rr)["id"].(string)

	rr = post(t, router, "/games/"+id+"/guesses", map[string]any{})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestCreateGuess_PlayerGuesses_RecordsGuess(t *testing.T) {
	store := newMockStore()
	router := newTestRouter(store)

	rr := post(t, router, "/games", map[string]any{"mode": "player_guesses"})
	id := decodeGame(t, rr)["id"].(string)

	rr = post(t, router, "/games/"+id+"/guesses", map[string]any{"guess": "123"})
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	g := decodeGame(t, rr)
	guesses := g["guesses"].([]any)
	if len(guesses) != 1 {
		t.Errorf("expected 1 guess recorded, got %d", len(guesses))
	}
}

// --- CreateGuess (computer_guesses) ---

func TestCreateGuess_ComputerGuesses_MissingFeedback(t *testing.T) {
	store := newMockStore()
	router := newTestRouter(store)

	rr := post(t, router, "/games", map[string]any{"mode": "computer_guesses"})
	id := decodeGame(t, rr)["id"].(string)

	rr = post(t, router, "/games/"+id+"/guesses", map[string]any{})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestCreateGuess_ComputerGuesses_RecordsFeedback(t *testing.T) {
	store := newMockStore()
	router := newTestRouter(store)

	rr := post(t, router, "/games", map[string]any{"mode": "computer_guesses"})
	id := decodeGame(t, rr)["id"].(string)

	rr = post(t, router, "/games/"+id+"/guesses", map[string]any{
		"feedback": map[string]any{"pico": 1, "fermi": 0},
	})
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200: %s", rr.Code, rr.Body.String())
	}
	g := decodeGame(t, rr)
	guesses := g["guesses"].([]any)
	if len(guesses) != 1 {
		t.Errorf("expected 1 guess recorded, got %d", len(guesses))
	}
}

// --- Rollback ---

func TestRollback_WrongMode(t *testing.T) {
	store := newMockStore()
	router := newTestRouter(store)

	rr := post(t, router, "/games", map[string]any{"mode": "player_guesses"})
	id := decodeGame(t, rr)["id"].(string)

	rr = post(t, router, "/games/"+id+"/rollback", map[string]any{"to_guess": 1})
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rr.Code)
	}
}

func TestRollback_Valid(t *testing.T) {
	store := newMockStore()
	router := newTestRouter(store)

	rr := post(t, router, "/games", map[string]any{"mode": "computer_guesses"})
	id := decodeGame(t, rr)["id"].(string)

	// Submit two rounds of feedback
	post(t, router, "/games/"+id+"/guesses", map[string]any{
		"feedback": map[string]any{"pico": 1, "fermi": 0},
	})
	post(t, router, "/games/"+id+"/guesses", map[string]any{
		"feedback": map[string]any{"pico": 0, "fermi": 1},
	})

	rr = post(t, router, "/games/"+id+"/rollback", map[string]any{"to_guess": 1})
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200: %s", rr.Code, rr.Body.String())
	}
	g := decodeGame(t, rr)
	guesses := g["guesses"].([]any)
	if len(guesses) != 1 {
		t.Errorf("expected 1 guess after rollback, got %d", len(guesses))
	}
}

func TestRollback_NotFound(t *testing.T) {
	store := newMockStore()
	router := newTestRouter(store)

	rr := post(t, router, "/games/no-such-game/rollback", map[string]any{"to_guess": 1})
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}
