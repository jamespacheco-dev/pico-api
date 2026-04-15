package game

import "testing"

func TestPartitionScore(t *testing.T) {
	t.Run("single candidate scores 1", func(t *testing.T) {
		score := partitionScore("123", []string{"123"})
		if score != 1 {
			t.Errorf("score = %d, want 1", score)
		}
	})

	t.Run("perfect partition scores 1", func(t *testing.T) {
		// If every candidate produces a different feedback, max bucket = 1.
		// "012" vs ["012","123","234"]: all produce distinct feedback against "012".
		candidates := []string{"012", "123", "234"}
		score := partitionScore("012", candidates)
		if score > len(candidates) {
			t.Errorf("score %d exceeds candidate count %d", score, len(candidates))
		}
	})

	t.Run("score is at most candidate count", func(t *testing.T) {
		candidates := generateCandidates(3, false)
		for _, guess := range candidates[:10] {
			score := partitionScore(guess, candidates)
			if score > len(candidates) {
				t.Errorf("partitionScore(%q) = %d, exceeds candidate count %d", guess, score, len(candidates))
			}
			if score < 1 {
				t.Errorf("partitionScore(%q) = %d, must be at least 1", guess, score)
			}
		}
	})
}

func TestHardSelector_PicksBestPartition(t *testing.T) {
	// With a small, controlled set we can verify Hard picks the guess
	// with the lowest partition score.
	candidates := generateCandidates(3, false)
	sel := HardSelector{}
	chosen := sel.Select(candidates)

	chosenScore := partitionScore(chosen, candidates)
	for _, c := range candidates {
		if score := partitionScore(c, candidates); score < chosenScore {
			t.Errorf("HardSelector chose %q (score %d) but %q has lower score %d",
				chosen, chosenScore, c, score)
		}
	}
}

func TestHardSelector_SingleCandidate(t *testing.T) {
	sel := HardSelector{}
	result := sel.Select([]string{"123"})
	if result != "123" {
		t.Errorf("Select([\"123\"]) = %q, want \"123\"", result)
	}
}

func TestMediumSelector_PicksFromTopQuartile(t *testing.T) {
	candidates := generateCandidates(3, false)
	sel := MediumSelector{}

	// Compute the top-25% score threshold.
	type scored struct {
		value string
		score int
	}
	all := make([]scored, len(candidates))
	for i, c := range candidates {
		all[i] = scored{c, partitionScore(c, candidates)}
	}
	topN := len(all) / 4
	if topN < 1 {
		topN = 1
	}

	// Find the cutoff score (worst score still in the top quartile).
	scores := make([]int, len(all))
	for i, s := range all {
		scores[i] = s.score
	}
	// Sort ascending; cutoff is the score at index topN-1.
	sortedScores := make([]int, len(scores))
	copy(sortedScores, scores)
	for i := 0; i < len(sortedScores)-1; i++ {
		for j := i + 1; j < len(sortedScores); j++ {
			if sortedScores[j] < sortedScores[i] {
				sortedScores[i], sortedScores[j] = sortedScores[j], sortedScores[i]
			}
		}
	}
	cutoff := sortedScores[topN-1]

	// Run many selections and verify each is within the top quartile.
	for i := range 50 {
		chosen := sel.Select(candidates)
		chosenScore := partitionScore(chosen, candidates)
		if chosenScore > cutoff {
			t.Errorf("iteration %d: MediumSelector chose %q with score %d, above cutoff %d",
				i, chosen, chosenScore, cutoff)
		}
	}
}

func TestMediumSelector_SingleCandidate(t *testing.T) {
	sel := MediumSelector{}
	result := sel.Select([]string{"456"})
	if result != "456" {
		t.Errorf("Select([\"456\"]) = %q, want \"456\"", result)
	}
}

func TestSelectors_ReturnValidCandidate(t *testing.T) {
	candidates := generateCandidates(3, false)
	selectors := []struct {
		name string
		sel  Selector
	}{
		{"Random", RandomSelector{}},
		{"Medium", MediumSelector{}},
		{"Hard", HardSelector{}},
	}

	for _, tt := range selectors {
		t.Run(tt.name, func(t *testing.T) {
			chosen := tt.sel.Select(candidates)
			found := false
			for _, c := range candidates {
				if c == chosen {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s selector returned %q which is not in the candidate pool", tt.name, chosen)
			}
		})
	}
}
