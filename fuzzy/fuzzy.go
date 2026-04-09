// Package fuzzy provides a simple fuzzy matching algorithm used to rank
// Makefile targets against a user query.
package fuzzy

import (
	"sort"
	"strings"
)

// Match represents a fuzzy match result.
type Match struct {
	// Target is the original string that was matched.
	Target string
	// Score is the match score; lower is better.
	Score int
	// Positions is the list of rune positions in Target that matched the
	// query. The positions are sorted in ascending order.
	Positions []int
}

// Filter returns the subset of targets that match the query, ordered by
// score (best first). An empty query returns all targets in their original
// order with empty position slices.
func Filter(query string, targets []string) []Match {
	if query == "" {
		out := make([]Match, len(targets))
		for i, t := range targets {
			out[i] = Match{Target: t, Score: 0, Positions: nil}
		}
		return out
	}

	results := make([]Match, 0, len(targets))
	for _, t := range targets {
		if m, ok := Score(query, t); ok {
			results = append(results, m)
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Score < results[j].Score
	})
	return results
}

// Score computes the fuzzy match score for a query against a target. The
// returned bool indicates whether the target matched at all.
//
// Scoring rules (lower is better):
//   - prefix match: -100
//   - runs of consecutive matching characters: -(run*10) per run
//   - early match start position: -(50 / (startIndex+1))
func Score(query, target string) (Match, bool) {
	if query == "" {
		return Match{Target: target, Score: 0}, true
	}

	lowerQuery := strings.ToLower(query)
	lowerTarget := strings.ToLower(target)

	qr := []rune(lowerQuery)
	tr := []rune(lowerTarget)

	positions := make([]int, 0, len(qr))
	qi := 0
	for i, r := range tr {
		if qi < len(qr) && r == qr[qi] {
			positions = append(positions, i)
			qi++
		}
	}
	if qi != len(qr) {
		return Match{}, false
	}

	score := 0

	// Prefix bonus.
	if strings.HasPrefix(lowerTarget, lowerQuery) {
		score -= 100
	}

	// Consecutive run bonus.
	if len(positions) > 0 {
		run := 1
		for i := 1; i < len(positions); i++ {
			if positions[i] == positions[i-1]+1 {
				run++
			} else {
				score -= run * 10
				run = 1
			}
		}
		score -= run * 10
	}

	// Early start position bonus.
	// Literal reading of the spec: "position の逆数 × 5".
	// Scaled by 10 to keep integer arithmetic meaningful.
	if len(positions) > 0 {
		score -= 50 / (positions[0] + 1)
	}

	return Match{
		Target:    target,
		Score:     score,
		Positions: positions,
	}, true
}

// PositionSet returns the match positions as a set for quick lookup during
// rendering.
func (m Match) PositionSet() map[int]struct{} {
	set := make(map[int]struct{}, len(m.Positions))
	for _, p := range m.Positions {
		set[p] = struct{}{}
	}
	return set
}
