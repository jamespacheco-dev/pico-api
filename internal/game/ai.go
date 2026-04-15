package game

import (
	"math/rand/v2"
	"sort"
)

// partitionScore returns the size of the largest feedback bucket when guess
// is played against all candidates. A lower score means the guess splits the
// candidate pool more evenly, leaving fewer candidates in the worst case.
func partitionScore(guess string, candidates []string) int {
	buckets := make(map[Feedback]int)
	for _, c := range candidates {
		buckets[Score(c, guess)]++
	}
	max := 0
	for _, count := range buckets {
		if count > max {
			max = count
		}
	}
	return max
}

// HardSelector picks the candidate that minimally partitions the remaining
// candidate space — the guess that minimizes the worst-case bucket size.
// Ties are broken by lexicographic order (first in the sorted candidate list).
type HardSelector struct{}

func (HardSelector) Select(candidates []string) string {
	best := len(candidates) + 1
	bestCandidate := candidates[0]
	for _, c := range candidates {
		if score := partitionScore(c, candidates); score < best {
			best = score
			bestCandidate = c
		}
	}
	return bestCandidate
}

// MediumSelector picks randomly from the top 25% of candidates by partition
// score. This approximates a human player who makes reasonable but non-optimal
// guesses. The percentile collapses toward Hard behaviour late in the game
// when the candidate pool is small — mirroring how a human plays more
// deliberately as options narrow.
type MediumSelector struct{}

func (MediumSelector) Select(candidates []string) string {
	type scored struct {
		value string
		score int
	}

	all := make([]scored, len(candidates))
	for i, c := range candidates {
		all[i] = scored{c, partitionScore(c, candidates)}
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].score < all[j].score
	})

	topN := len(all) / 4
	if topN < 1 {
		topN = 1
	}

	return all[rand.IntN(topN)].value
}
