package classification

import (
	"sort"
)

// mapIndustryTextToNAICS returns NAICS codes most relevant to a free-text industry label
func mapIndustryTextToNAICS(industryText string, data *IndustryCodeData) []string {
	if data == nil || industryText == "" {
		return nil
	}
	// Combine keyword and fuzzy approaches
	seen := make(map[string]struct{})
	var results []string

	for _, c := range data.SearchNAICSByKeyword(industryText) {
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		results = append(results, c)
	}
	for _, c := range data.SearchNAICSByFuzzy(industryText, 0.82) {
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		results = append(results, c)
	}
	return results
}

type scoredCode struct {
	code  string
	score float64
}

// crosswalkFromNAICS finds related MCC and SIC codes for a given NAICS by text similarity
func crosswalkFromNAICS(naicsCode string, data *IndustryCodeData) (mcc []string, sic []string) {
	if data == nil {
		return nil, nil
	}
	title := data.GetNAICSName(naicsCode)
	if title == "" || title == "Unknown NAICS Industry" {
		return nil, nil
	}

	// Score all MCC/SIC descriptions against NAICS title and keep best matches
	var mccScored []scoredCode
	for code, desc := range data.MCC {
		// Use bidirectional comparison and token overlap to account for asymmetry
		s := bidirectionalSimilarity(title, desc)
		if s < 0.55 {
			// Fallback: token overlap heuristic
			if overlap := tokenOverlapCount(title, desc); overlap >= 1 {
				s = 0.6 + 0.05*float64(overlap) // boost score modestly with overlap
			}
		}
		if s >= 0.55 {
			mccScored = append(mccScored, scoredCode{code: code, score: s})
		}
	}
	sort.Slice(mccScored, func(i, j int) bool { return mccScored[i].score > mccScored[j].score })
	// take top 3
	for i := 0; i < len(mccScored) && i < 3; i++ {
		mcc = append(mcc, mccScored[i].code)
	}

	var sicScored []scoredCode
	for code, desc := range data.SIC {
		s := bidirectionalSimilarity(title, desc)
		if s < 0.55 {
			if overlap := tokenOverlapCount(title, desc); overlap >= 1 {
				s = 0.6 + 0.05*float64(overlap)
			}
		}
		if s >= 0.55 {
			sicScored = append(sicScored, scoredCode{code: code, score: s})
		}
	}
	sort.Slice(sicScored, func(i, j int) bool { return sicScored[i].score > sicScored[j].score })
	for i := 0; i < len(sicScored) && i < 3; i++ {
		sic = append(sic, sicScored[i].code)
	}

	return mcc, sic
}

// bidirectionalSimilarity takes the max similarity in both directions to mitigate length asymmetry
func bidirectionalSimilarity(a, b string) float64 {
	sa := tokenMaxSimilarity(a, b)
	sb := tokenMaxSimilarity(b, a)
	if sa > sb {
		return sa
	}
	return sb
}

// tokenOverlapCount counts overlapping informative tokens between two strings
func tokenOverlapCount(a, b string) int {
	na := normalizeText(a)
	nb := normalizeText(b)
	ta := tokenize(na)
	tb := tokenize(nb)
	if len(ta) == 0 || len(tb) == 0 {
		return 0
	}
	set := make(map[string]struct{}, len(ta))
	for _, t := range ta {
		set[t] = struct{}{}
	}
	count := 0
	for _, t := range tb {
		if _, ok := set[t]; ok {
			count++
		}
	}
	return count
}
