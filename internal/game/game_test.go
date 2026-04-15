package game

import (
	"errors"
	"testing"
)

// fixedSelector always returns a specific value — used to control
// what the computer "guesses" in tests that need a predictable next guess.
type fixedSelector struct{ value string }

func (f fixedSelector) Select(_ []string) string { return f.value }

func newPlayerGame(secret string) *Game {
	return &Game{
		Mode:     ModePlayerGuesses,
		Config:   Config{Length: len(secret), AllowRepeats: true},
		Status:   StatusInProgress,
		Guesses:  []Guess{},
		secret:   secret,
		selector: RandomSelector{},
	}
}

func newComputerGame(currentGuess string, candidates []string) *Game {
	return &Game{
		Mode:         ModeComputerGuesses,
		Config:       Config{Length: len(currentGuess), AllowRepeats: true},
		Status:       StatusInProgress,
		Guesses:      []Guess{},
		CurrentGuess: currentGuess,
		candidates:   candidates,
		selector:     firstSelector{},
	}
}

// --- ApplyGuess ---

func TestApplyGuess_CorrectGuess(t *testing.T) {
	g := newPlayerGame("123")
	fb, err := g.ApplyGuess("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.Fermi != 3 || fb.Pico != 0 || fb.Bagel {
		t.Errorf("expected all fermi, got %+v", fb)
	}
	if g.Status != StatusComplete {
		t.Errorf("Status = %v, want complete", g.Status)
	}
}

func TestApplyGuess_WrongGuess(t *testing.T) {
	g := newPlayerGame("123")
	fb, err := g.ApplyGuess("456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fb.Bagel {
		t.Errorf("expected bagel, got %+v", fb)
	}
	if g.Status != StatusInProgress {
		t.Error("status should remain in_progress after wrong guess")
	}
	if len(g.Guesses) != 1 {
		t.Errorf("expected 1 guess recorded, got %d", len(g.Guesses))
	}
	if g.Guesses[0].Number != 1 {
		t.Errorf("guess number = %d, want 1", g.Guesses[0].Number)
	}
}

func TestApplyGuess_GuessNumberIncrements(t *testing.T) {
	g := newPlayerGame("123")
	g.ApplyGuess("456")
	g.ApplyGuess("789")
	if g.Guesses[1].Number != 2 {
		t.Errorf("second guess number = %d, want 2", g.Guesses[1].Number)
	}
}

func TestApplyGuess_WrongLength(t *testing.T) {
	g := newPlayerGame("123")
	_, err := g.ApplyGuess("12")
	if err == nil {
		t.Error("expected error for wrong length guess")
	}
}

func TestApplyGuess_NonDigit(t *testing.T) {
	g := newPlayerGame("123")
	_, err := g.ApplyGuess("12a")
	if err == nil {
		t.Error("expected error for non-digit character")
	}
}

func TestApplyGuess_RepeatedDigitNotAllowed(t *testing.T) {
	g := &Game{
		Mode:    ModePlayerGuesses,
		Config:  Config{Length: 3, AllowRepeats: false},
		Status:  StatusInProgress,
		Guesses: []Guess{},
		secret:  "123",
	}
	_, err := g.ApplyGuess("112")
	if err == nil {
		t.Error("expected error for repeated digit when not allowed")
	}
}

func TestApplyGuess_RepeatedDigitAllowed(t *testing.T) {
	g := newPlayerGame("112")
	_, err := g.ApplyGuess("112")
	if err != nil {
		t.Errorf("unexpected error for repeated digit when allowed: %v", err)
	}
}

func TestApplyGuess_WrongMode(t *testing.T) {
	g := newComputerGame("123", []string{"123", "456"})
	_, err := g.ApplyGuess("123")
	if !errors.Is(err, ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestApplyGuess_GameAlreadyComplete(t *testing.T) {
	g := newPlayerGame("123")
	g.Status = StatusComplete
	_, err := g.ApplyGuess("123")
	if !errors.Is(err, ErrGameComplete) {
		t.Errorf("expected ErrGameComplete, got %v", err)
	}
}

// --- ApplyFeedback ---

func TestApplyFeedback_AllFermi(t *testing.T) {
	candidates := generateCandidates(3, true)
	g := newComputerGame("123", candidates)
	next, err := g.ApplyFeedback(Feedback{Pico: 0, Fermi: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next != "" {
		t.Errorf("next guess should be empty when game complete, got %q", next)
	}
	if g.Status != StatusComplete {
		t.Errorf("Status = %v, want complete", g.Status)
	}
}

func TestApplyFeedback_Normal(t *testing.T) {
	candidates := generateCandidates(3, false)
	g := newComputerGame("012", candidates)
	next, err := g.ApplyFeedback(Feedback{Pico: 1, Fermi: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next == "" {
		t.Error("expected a next guess")
	}
	if g.Status != StatusInProgress {
		t.Error("status should remain in_progress")
	}
	if len(g.Guesses) != 1 {
		t.Errorf("expected 1 guess recorded, got %d", len(g.Guesses))
	}
}

func TestApplyFeedback_BagelDerivedFromPicoFermi(t *testing.T) {
	candidates := generateCandidates(3, false)
	g := newComputerGame("012", candidates)
	// Client sends bagel=false but pico=0, fermi=0 — server should derive bagel=true
	_, err := g.ApplyFeedback(Feedback{Pico: 0, Fermi: 0, Bagel: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !g.Guesses[0].Feedback.Bagel {
		t.Error("bagel should be derived as true when pico=0 and fermi=0")
	}
}

func TestApplyFeedback_Contradictory(t *testing.T) {
	// Only one candidate: "123". Giving feedback that "123" is wrong is contradictory.
	g := newComputerGame("123", []string{"123"})
	_, err := g.ApplyFeedback(Feedback{Pico: 1, Fermi: 0})
	if !errors.Is(err, ErrContradictory) {
		t.Errorf("expected ErrContradictory, got %v", err)
	}
	// State should be unchanged
	if len(g.Guesses) != 0 {
		t.Error("guess should not be recorded after contradictory feedback")
	}
}

func TestApplyFeedback_InvalidFeedback(t *testing.T) {
	candidates := generateCandidates(3, false)
	g := newComputerGame("012", candidates)
	_, err := g.ApplyFeedback(Feedback{Pico: 2, Fermi: 2}) // 4 > length of 3
	if err == nil {
		t.Error("expected error for pico+fermi > length")
	}
}

func TestApplyFeedback_WrongMode(t *testing.T) {
	g := newPlayerGame("123")
	_, err := g.ApplyFeedback(Feedback{Pico: 1, Fermi: 1})
	if !errors.Is(err, ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

func TestApplyFeedback_GameAlreadyComplete(t *testing.T) {
	g := newComputerGame("123", []string{"123"})
	g.Status = StatusComplete
	_, err := g.ApplyFeedback(Feedback{Pico: 0, Fermi: 3})
	if !errors.Is(err, ErrGameComplete) {
		t.Errorf("expected ErrGameComplete, got %v", err)
	}
}

// --- Rollback ---

func TestRollback_Valid(t *testing.T) {
	candidates := generateCandidates(3, false)
	g := newComputerGame("012", candidates)

	g.ApplyFeedback(Feedback{Pico: 0, Fermi: 0, Bagel: true})
	g.ApplyFeedback(Feedback{Pico: 1, Fermi: 0})

	if len(g.Guesses) != 2 {
		t.Fatalf("expected 2 guesses before rollback, got %d", len(g.Guesses))
	}

	err := g.Rollback(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(g.Guesses) != 1 {
		t.Errorf("expected 1 guess after rollback, got %d", len(g.Guesses))
	}
	if g.CurrentGuess == "" {
		t.Error("expected a new current guess after rollback")
	}
	if g.Status != StatusInProgress {
		t.Errorf("Status = %v, want in_progress", g.Status)
	}
}

func TestRollback_OutOfRange(t *testing.T) {
	candidates := generateCandidates(3, false)
	g := newComputerGame("012", candidates)
	g.ApplyFeedback(Feedback{Pico: 1, Fermi: 0})

	if err := g.Rollback(5); !errors.Is(err, ErrOutOfRange) {
		t.Errorf("expected ErrOutOfRange, got %v", err)
	}
	if err := g.Rollback(0); !errors.Is(err, ErrOutOfRange) {
		t.Errorf("expected ErrOutOfRange for 0, got %v", err)
	}
}

func TestRollback_NoGuesses(t *testing.T) {
	g := newComputerGame("012", generateCandidates(3, false))
	if err := g.Rollback(1); !errors.Is(err, ErrNoGuesses) {
		t.Errorf("expected ErrNoGuesses, got %v", err)
	}
}

func TestRollback_WrongMode(t *testing.T) {
	g := newPlayerGame("123")
	if err := g.Rollback(1); !errors.Is(err, ErrWrongMode) {
		t.Errorf("expected ErrWrongMode, got %v", err)
	}
}

// --- IsComplete ---

func TestIsComplete(t *testing.T) {
	g := newPlayerGame("123")
	if g.IsComplete() {
		t.Error("new game should not be complete")
	}
	g.ApplyGuess("123")
	if !g.IsComplete() {
		t.Error("game should be complete after correct guess")
	}
}
