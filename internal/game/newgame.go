package game

import (
	"crypto/rand"
	"fmt"
	"math"
	mathrand "math/rand/v2"
	"time"
)

// OptimalGuesses returns the guess limit for Hard difficulty: ceil(log10(possibilities) * 2).
func OptimalGuesses(length int, allowRepeats bool) int {
	n := 1
	for i := 0; i < length; i++ {
		if allowRepeats {
			n *= 10
		} else {
			n *= (10 - i)
		}
	}
	return int(math.Ceil(math.Log10(float64(n)) * 2))
}

// NewGame creates a new game session with the given configuration.
// For player_guesses mode: generates a random secret number.
// For computer_guesses mode: builds the candidate pool and picks the first guess.
func NewGame(cfg Config, mode Mode, difficulty Difficulty, sel Selector) (*Game, error) {
	id, err := newID()
	if err != nil {
		return nil, fmt.Errorf("generating game ID: %w", err)
	}

	g := &Game{
		ID:             id,
		Mode:           mode,
		Config:         cfg,
		Difficulty:     difficulty,
		Status:         StatusInProgress,
		Guesses:        []Guess{},
		selector:       sel,
		LastActivityAt: time.Now(),
	}

	candidates := generateCandidates(cfg.Length, cfg.AllowRepeats)

	switch mode {
	case ModePlayerGuesses:
		g.secret = candidates[mathrand.IntN(len(candidates))]
		switch difficulty {
		case DifficultyHard:
			g.MaxGuesses = OptimalGuesses(cfg.Length, cfg.AllowRepeats)
		case DifficultyMedium:
			g.MaxGuesses = int(math.Ceil(float64(OptimalGuesses(cfg.Length, cfg.AllowRepeats)) * 1.5))
		}
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
