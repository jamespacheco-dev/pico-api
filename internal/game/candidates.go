package game

// generateCandidates returns all digit strings of the given length.
// If allowRepeats is false, only strings with distinct digits are included.
// Results are in lexicographic order.
func generateCandidates(length int, allowRepeats bool) []string {
	var result []string
	current := make([]byte, 0, length)

	var build func(depth int, used [10]bool)
	build = func(depth int, used [10]bool) {
		if depth == length {
			result = append(result, string(current))
			return
		}
		for d := byte('0'); d <= '9'; d++ {
			idx := d - '0'
			if !allowRepeats && used[idx] {
				continue
			}
			current = append(current, d)
			used[idx] = true
			build(depth+1, used)
			current = current[:len(current)-1]
			used[idx] = false
		}
	}

	build(0, [10]bool{})
	return result
}

// filterCandidates returns only those candidates consistent with the given
// guess having produced the given feedback. A candidate is consistent if,
// treating it as the secret, scoring guess against it yields the same feedback.
func filterCandidates(candidates []string, guess string, fb Feedback) []string {
	result := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if Score(c, guess) == fb {
			result = append(result, c)
		}
	}
	return result
}
