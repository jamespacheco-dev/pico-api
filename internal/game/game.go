package game

import (
	"errors"
	"fmt"
	"time"
)

// Mode represents which player is guessing.
type Mode string

const (
	ModePlayerGuesses   Mode = "player_guesses"
	ModeComputerGuesses Mode = "computer_guesses"
)

// Difficulty controls the AI opponent's selection strategy.
type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

// Status represents the current state of a game session.
type Status string

const (
	StatusInProgress Status = "in_progress"
	StatusComplete   Status = "complete"
	StatusLost       Status = "lost"
)

// Config holds the settings chosen at game creation.
type Config struct {
	Length       int  `json:"length"`
	AllowRepeats bool `json:"allow_repeats"`
}

// Feedback is the structured result of evaluating a guess.
type Feedback struct {
	Pico  int  `json:"pico"`
	Fermi int  `json:"fermi"`
	Bagel bool `json:"bagel"`
}

// Guess is one round of play, with the value guessed and the feedback given.
type Guess struct {
	Number   int      `json:"number"`
	Value    string   `json:"value"`
	Feedback Feedback `json:"feedback"`
}

// Game is a single game session.
type Game struct {
	ID             string     `json:"id"`
	Mode           Mode       `json:"mode"`
	Config         Config     `json:"config"`
	Difficulty     Difficulty `json:"difficulty"`
	Status         Status     `json:"status"`
	MaxGuesses     int        `json:"max_guesses"`
	CurrentGuess   string     `json:"current_guess,omitempty"`
	RevealedSecret string     `json:"revealed_secret,omitempty"`
	Guesses        []Guess    `json:"guesses"`
	LastActivityAt time.Time  `json:"last_activity_at"`

	secret     string
	candidates []string
	selector   Selector
}

// Sentinel errors for use with errors.Is in the handler layer.
var (
	ErrWrongMode     = errors.New("wrong mode for this operation")
	ErrGameComplete  = errors.New("game is already complete")
	ErrContradictory = errors.New("contradictory feedback: no valid candidates remain")
	ErrOutOfRange    = errors.New("to_guess is out of range")
)

// IsComplete reports whether the game has been won.
func (g *Game) IsComplete() bool {
	return g.Status == StatusComplete
}

// IsGameOver reports whether the game has ended (won or lost).
func (g *Game) IsGameOver() bool {
	return g.Status == StatusComplete || g.Status == StatusLost
}

// ApplyGuess processes a player's guess in player_guesses mode.
// Returns the feedback for the guess.
func (g *Game) ApplyGuess(guess string) (Feedback, error) {
	if g.Mode != ModePlayerGuesses {
		return Feedback{}, ErrWrongMode
	}
	if g.Status != StatusInProgress {
		return Feedback{}, ErrGameComplete
	}
	if err := g.validateGuess(guess); err != nil {
		return Feedback{}, err
	}

	fb := Score(g.secret, guess)
	g.Guesses = append(g.Guesses, Guess{
		Number:   len(g.Guesses) + 1,
		Value:    guess,
		Feedback: fb,
	})

	if fb.Fermi == g.Config.Length {
		g.Status = StatusComplete
	} else if g.MaxGuesses > 0 && len(g.Guesses) >= g.MaxGuesses {
		g.Status = StatusLost
		g.RevealedSecret = g.secret
	}

	g.LastActivityAt = time.Now()
	return fb, nil
}

// ApplyFeedback processes player feedback on the computer's current guess
// in computer_guesses mode. Bagel is derived from pico and fermi — the
// client's bagel value is ignored. Returns the computer's next guess, or
// empty string if the game is now complete.
func (g *Game) ApplyFeedback(fb Feedback) (string, error) {
	if g.Mode != ModeComputerGuesses {
		return "", ErrWrongMode
	}
	if g.Status != StatusInProgress {
		return "", ErrGameComplete
	}

	fb.Bagel = fb.Pico == 0 && fb.Fermi == 0

	if err := g.validateFeedback(fb); err != nil {
		return "", err
	}

	// Filter before committing so a contradictory response leaves state unchanged.
	filtered := filterCandidates(g.candidates, g.CurrentGuess, fb)
	if len(filtered) == 0 {
		return "", ErrContradictory
	}

	g.Guesses = append(g.Guesses, Guess{
		Number:   len(g.Guesses) + 1,
		Value:    g.CurrentGuess,
		Feedback: fb,
	})
	g.candidates = filtered

	if fb.Fermi == g.Config.Length {
		g.Status = StatusComplete
		g.CurrentGuess = ""
		return "", nil
	}

	g.CurrentGuess = g.selector.Select(g.candidates)
	g.LastActivityAt = time.Now()
	return g.CurrentGuess, nil
}

// Rollback truncates guess history to the first toGuess entries, rebuilds
// the candidate pool by replaying that feedback, and picks a new current guess.
// Only valid in computer_guesses mode.
func (g *Game) Rollback(toGuess int) error {
	if g.Mode != ModeComputerGuesses {
		return ErrWrongMode
	}
	if toGuess < 0 || toGuess > len(g.Guesses) {
		return ErrOutOfRange
	}

	g.Guesses = g.Guesses[:toGuess]

	g.candidates = generateCandidates(g.Config.Length, g.Config.AllowRepeats)
	for _, guess := range g.Guesses {
		g.candidates = filterCandidates(g.candidates, guess.Value, guess.Feedback)
	}

	g.Status = StatusInProgress
	g.CurrentGuess = g.selector.Select(g.candidates)
	g.LastActivityAt = time.Now()
	return nil
}

func (g *Game) validateGuess(guess string) error {
	if len(guess) != g.Config.Length {
		return fmt.Errorf("guess must be %d digits", g.Config.Length)
	}
	var seen [10]bool
	for _, c := range guess {
		if c < '0' || c > '9' {
			return errors.New("guess must contain only digits")
		}
		idx := c - '0'
		if !g.Config.AllowRepeats && seen[idx] {
			return errors.New("repeated digits are not allowed")
		}
		seen[idx] = true
	}
	return nil
}

func (g *Game) validateFeedback(fb Feedback) error {
	if fb.Pico < 0 || fb.Fermi < 0 {
		return errors.New("feedback values cannot be negative")
	}
	if fb.Pico+fb.Fermi > g.Config.Length {
		return fmt.Errorf("pico + fermi cannot exceed number length (%d)", g.Config.Length)
	}
	return nil
}
