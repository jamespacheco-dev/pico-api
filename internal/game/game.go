package game

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
	ID           string     `json:"id"`
	Mode         Mode       `json:"mode"`
	Config       Config     `json:"config"`
	Difficulty   Difficulty `json:"difficulty"`
	Status       Status     `json:"status"`
	CurrentGuess string     `json:"current_guess,omitempty"`
	Guesses      []Guess    `json:"guesses"`

	secret     string
	candidates []string
}
