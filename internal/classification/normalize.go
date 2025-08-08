package classification

import (
	"regexp"
	"strings"
	"unicode"
)

// stopwords is a minimal English stopword set for normalization
var stopwords = map[string]struct{}{
	"the": {}, "and": {}, "of": {}, "for": {}, "to": {}, "in": {}, "a": {}, "an": {},
	"co": {}, "company": {}, "inc": {}, "llc": {}, "corp": {}, "corporation": {}, "ltd": {},
	"services": {}, "solutions": {}, "group": {}, "international": {}, "global": {},
}

var spaceCollapse = regexp.MustCompile(`\s+`)

// normalizeText lowercases, removes punctuation, collapses whitespace
func normalizeText(s string) string {
	if s == "" {
		return s
	}
	// Lowercase and remove punctuation
	b := strings.Builder{}
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			b.WriteRune(unicode.ToLower(r))
		} else {
			// replace punctuation with space to keep token boundaries
			b.WriteRune(' ')
		}
	}
	out := spaceCollapse.ReplaceAllString(b.String(), " ")
	return strings.TrimSpace(out)
}

// tokenize splits normalized text and removes simple stopwords
func tokenize(normalized string) []string {
	if normalized == "" {
		return nil
	}
	parts := strings.Fields(normalized)
	if len(parts) == 0 {
		return nil
	}
	tokens := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		if _, skip := stopwords[p]; skip {
			continue
		}
		if _, exists := seen[p]; exists {
			continue
		}
		seen[p] = struct{}{}
		tokens = append(tokens, p)
	}
	return tokens
}

// normalizeBusinessFields returns normalized free-text and tokens for request fields
func normalizeBusinessFields(name, description, keywords string) (normalized string, tokens []string) {
	n := normalizeText(name)
	d := normalizeText(description)
	k := normalizeText(keywords)
	joined := strings.TrimSpace(strings.Join([]string{n, d, k}, " "))
	return joined, tokenize(joined)
}
