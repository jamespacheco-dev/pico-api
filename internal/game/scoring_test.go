package game

import "testing"

func TestScore(t *testing.T) {
	tests := []struct {
		name   string
		secret string
		guess  string
		want   Feedback
	}{
		{
			name:   "all fermi",
			secret: "123",
			guess:  "123",
			want:   Feedback{Pico: 0, Fermi: 3, Bagel: false},
		},
		{
			name:   "all pico",
			secret: "123",
			guess:  "231",
			want:   Feedback{Pico: 3, Fermi: 0, Bagel: false},
		},
		{
			name:   "bagel",
			secret: "123",
			guess:  "456",
			want:   Feedback{Pico: 0, Fermi: 0, Bagel: true},
		},
		{
			name:   "mixed pico and fermi",
			secret: "123",
			guess:  "132",
			want:   Feedback{Pico: 2, Fermi: 1, Bagel: false},
		},
		{
			name:   "repeated digit in guess, one match",
			secret: "112",
			guess:  "211",
			want:   Feedback{Pico: 2, Fermi: 1, Bagel: false},
		},
		{
			name:   "repeated digit in secret, fermi takes priority",
			secret: "112",
			guess:  "121",
			want:   Feedback{Pico: 2, Fermi: 1, Bagel: false},
		},
		{
			name:   "no double-counting repeated digits",
			secret: "112",
			guess:  "111",
			want:   Feedback{Pico: 0, Fermi: 2, Bagel: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Score(tt.secret, tt.guess)
			if got != tt.want {
				t.Errorf("Score(%q, %q) = %+v, want %+v", tt.secret, tt.guess, got, tt.want)
			}
		})
	}
}
