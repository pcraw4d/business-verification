package classification

import (
	"math"
	"strings"
)

// levenshteinDistance computes the Levenshtein edit distance between two strings
func levenshteinDistance(a, b string) int {
	la := len(a)
	lb := len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// Initialize two rolling rows to reduce memory
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		ai := a[i-1]
		for j := 1; j <= lb; j++ {
			cost := 0
			if ai != b[j-1] {
				cost = 1
			}
			deletion := prev[j] + 1
			insertion := curr[j-1] + 1
			substitution := prev[j-1] + cost
			curr[j] = minInt(deletion, insertion, substitution)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func minInt(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// similarity computes a normalized similarity score in [0,1] based on Levenshtein distance
// 1.0 means identical, 0.0 means completely different. Empty strings return 0 unless both empty.
func similarity(a, b string) float64 {
	if a == "" && b == "" {
		return 1.0
	}
	if a == "" || b == "" {
		return 0.0
	}
	// Normalize both inputs using our normalization pipeline
	na := normalizeText(a)
	nb := normalizeText(b)
	if na == "" && nb == "" {
		return 1.0
	}
	if na == "" || nb == "" {
		return 0.0
	}
	dist := float64(levenshteinDistance(na, nb))
	denom := float64(maxInt(len(na), len(nb)))
	if denom == 0 {
		return 1.0
	}
	return 1.0 - (dist / denom)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// tokenMaxSimilarity returns the maximum similarity between the query and any token of text
func tokenMaxSimilarity(query, text string) float64 {
	nq := normalizeText(query)
	nt := normalizeText(text)
	if nq == "" || nt == "" {
		return 0.0
	}
	maxScore := 0.0
	for _, tok := range strings.Fields(nt) {
		if len(tok) < 3 {
			continue
		}
		s := similarity(nq, tok)
		if s > maxScore {
			maxScore = s
		}
		if maxScore >= 0.999 {
			// Early exit for exact match
			break
		}
	}
	// Also compare against full text to capture multi-word closeness
	maxScore = math.Max(maxScore, similarity(nq, nt))
	return maxScore
}
