package game

import (
	"crypto/rand"
	"fmt"
	mathrand "math/rand/v2"
)

// NewGame creates a new game session with the given configuration.
// For player_guesses mode: generates a random secret number.
// For computer_guesses mode: builds the candidate pool and picks the first guess.
func NewGame(cfg Config, mode Mode, difficulty Difficulty, sel Selector) (*Game, error) {
	id, err := newID()
	if err != nil {
		return nil, fmt.Errorf("generating game ID: %w", err)
	}

	g := &Game{
		ID:         id,
		Mode:       mode,
		Config:     cfg,
		Difficulty: difficulty,
		Status:     StatusInProgress,
		Guesses:    []Guess{},
		selector:   sel,
	}

	candidates := generateCandidates(cfg.Length, cfg.AllowRepeats)

	switch mode {
	case ModePlayerGuesses:
		g.secret = candidates[mathrand.IntN(len(candidates))]
	case ModeComputerGuesses:
		g.candidates = candidates
		g.CurrentGuess = sel.Select(candidates)
	}

	return g, nil
}

// newID generates a random UUID v4 string.
func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
