package game

// Score computes Pico/Fermi/Bagel feedback for a guess against a secret.
// Algorithm: match Fermis first (correct digit, correct position), remove
// those positions from consideration, then count Picos from the remainder.
func Score(secret, guess string) Feedback {
	fermis := 0
	secretRem := []rune{}
	guessRem := []rune{}

	for i := range secret {
		if secret[i] == guess[i] {
			fermis++
		} else {
			secretRem = append(secretRem, rune(secret[i]))
			guessRem = append(guessRem, rune(guess[i]))
		}
	}

	picos := 0
	used := make([]bool, len(secretRem))
	for _, g := range guessRem {
		for j, s := range secretRem {
			if !used[j] && g == s {
				picos++
				used[j] = true
				break
			}
		}
	}

	bagel := fermis == 0 && picos == 0
	return Feedback{Pico: picos, Fermi: fermis, Bagel: bagel}
}
