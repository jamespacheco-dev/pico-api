package game

import (
	"strings"
	"testing"
)

// firstSelector always picks the first candidate — deterministic for testing.
type firstSelector struct{}

func (firstSelector) Select(candidates []string) string { return candidates[0] }

func TestNewGame_PlayerGuesses(t *testing.T) {
	cfg := Config{Length: 3, AllowRepeats: false}
	g, err := NewGame(cfg, ModePlayerGuesses, DifficultyEasy, RandomSelector{})
	if err != nil {
		t.Fatalf("NewGame returned error: %v", err)
	}

	if g.ID == "" {
		t.Error("game ID should not be empty")
	}
	if g.Mode != ModePlayerGuesses {
		t.Errorf("Mode = %v, want %v", g.Mode, ModePlayerGuesses)
	}
	if g.Status != StatusInProgress {
		t.Errorf("Status = %v, want %v", g.Status, StatusInProgress)
	}
	if g.CurrentGuess != "" {
		t.Errorf("CurrentGuess should be empty for player_guesses, got %q", g.CurrentGuess)
	}
	if g.secret == "" {
		t.Error("secret should be set for player_guesses mode")
	}
	if len(g.secret) != cfg.Length {
		t.Errorf("secret length = %d, want %d", len(g.secret), cfg.Length)
	}
	if g.candidates != nil {
		t.Error("candidates should not be set for player_guesses mode")
	}
}

func TestNewGame_ComputerGuesses(t *testing.T) {
	cfg := Config{Length: 3, AllowRepeats: false}
	g, err := NewGame(cfg, ModeComputerGuesses, DifficultyEasy, firstSelector{})
	if err != nil {
		t.Fatalf("NewGame returned error: %v", err)
	}

	if g.CurrentGuess == "" {
		t.Error("CurrentGuess should be set for computer_guesses mode")
	}
	if len(g.CurrentGuess) != cfg.Length {
		t.Errorf("CurrentGuess length = %d, want %d", len(g.CurrentGuess), cfg.Length)
	}
	if len(g.candidates) != 720 { // 10*9*8 no repeats
		t.Errorf("candidates count = %d, want 720", len(g.candidates))
	}
	if g.secret != "" {
		t.Error("secret should not be set for computer_guesses mode")
	}
}

func TestNewGame_IDFormat(t *testing.T) {
	cfg := Config{Length: 3, AllowRepeats: false}
	g, err := NewGame(cfg, ModePlayerGuesses, DifficultyEasy, RandomSelector{})
	if err != nil {
		t.Fatalf("NewGame returned error: %v", err)
	}
	// UUID v4 format: 8-4-4-4-12 hex chars
	parts := strings.Split(g.ID, "-")
	if len(parts) != 5 {
		t.Errorf("ID %q does not look like a UUID", g.ID)
	}
}

func TestNewGame_SecretIsValidCandidate(t *testing.T) {
	cfg := Config{Length: 3, AllowRepeats: false}
	g, err := NewGame(cfg, ModePlayerGuesses, DifficultyEasy, RandomSelector{})
	if err != nil {
		t.Fatalf("NewGame returned error: %v", err)
	}

	candidates := generateCandidates(cfg.Length, cfg.AllowRepeats)
	found := false
	for _, c := range candidates {
		if c == g.secret {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("secret %q is not a valid candidate", g.secret)
	}
}
