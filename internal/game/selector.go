package game

import "math/rand/v2"

// Selector picks one guess from a list of valid candidates.
// The three difficulty levels are distinct implementations of this interface.
type Selector interface {
	Select(candidates []string) string
}

// RandomSelector picks uniformly at random. Used for Easy difficulty.
// Also used by all difficulty levels on the first guess, since with no
// constraints applied yet, every selection strategy reduces to random.
type RandomSelector struct{}

func (RandomSelector) Select(candidates []string) string {
	return candidates[rand.IntN(len(candidates))]
}
