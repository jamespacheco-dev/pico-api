package game

import "testing"

func TestGenerateCandidates(t *testing.T) {
	tests := []struct {
		name         string
		length       int
		allowRepeats bool
		wantCount    int
	}{
		{"3 digits no repeats", 3, false, 720},  // 10 * 9 * 8
		{"3 digits with repeats", 3, true, 1000}, // 10^3
		{"1 digit no repeats", 1, false, 10},
		{"1 digit with repeats", 1, true, 10},
		{"2 digits no repeats", 2, false, 90}, // 10 * 9
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateCandidates(tt.length, tt.allowRepeats)
			if len(got) != tt.wantCount {
				t.Errorf("generateCandidates(%d, %v) returned %d candidates, want %d",
					tt.length, tt.allowRepeats, len(got), tt.wantCount)
			}
			for _, c := range got {
				if len(c) != tt.length {
					t.Errorf("candidate %q has wrong length", c)
				}
				if !tt.allowRepeats {
					seen := [10]bool{}
					for _, ch := range c {
						idx := ch - '0'
						if seen[idx] {
							t.Errorf("candidate %q has repeated digit", c)
						}
						seen[idx] = true
					}
				}
			}
		})
	}
}

func TestFilterCandidates(t *testing.T) {
	candidates := []string{"123", "456", "789", "132", "213"}

	tests := []struct {
		name      string
		guess     string
		feedback  Feedback
		wantCount int
		wantIn    []string
	}{
		{
			name:      "all fermi keeps only exact match",
			guess:     "123",
			feedback:  Feedback{Pico: 0, Fermi: 3, Bagel: false},
			wantCount: 1,
			wantIn:    []string{"123"},
		},
		{
			name:      "bagel keeps candidates with no shared digits",
			guess:     "123",
			feedback:  Feedback{Pico: 0, Fermi: 0, Bagel: true},
			wantCount: 2,
			wantIn:    []string{"456", "789"},
		},
		{
			name:      "pico fermi filters correctly",
			guess:     "123",
			feedback:  Feedback{Pico: 2, Fermi: 1, Bagel: false},
			wantCount: 2,
			wantIn:    []string{"132", "213"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterCandidates(candidates, tt.guess, tt.feedback)
			if len(got) != tt.wantCount {
				t.Errorf("filterCandidates got %v, want %d results", got, tt.wantCount)
			}
			gotSet := make(map[string]bool)
			for _, c := range got {
				gotSet[c] = true
			}
			for _, want := range tt.wantIn {
				if !gotSet[want] {
					t.Errorf("expected %q in results, got %v", want, got)
				}
			}
		})
	}
}
