package validators

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/pcraw4d/business-verification/internal/datasource"
)

var (
	htmlTagRegex   = regexp.MustCompile(`<[^>]*>`)
	spaceCollapseR = regexp.MustCompile(`\s+`)
)

// SanitizeText removes HTML tags, control characters, collapses whitespace, and trims length
func SanitizeText(input string, maxLen int) string {
	if input == "" {
		return input
	}
	// Strip HTML
	s := htmlTagRegex.ReplaceAllString(input, " ")
	// Remove control chars
	b := strings.Builder{}
	b.Grow(len(s))
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' || unicode.IsSpace(r) || unicode.IsGraphic(r) {
			// keep printable and whitespace
			b.WriteRune(r)
		}
	}
	s = b.String()
	// Collapse whitespace
	s = spaceCollapseR.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	// Enforce max length
	if maxLen > 0 && len(s) > maxLen {
		s = s[:maxLen]
	}
	return s
}

// NormalizeKeywords cleans, lowercases, and de-duplicates keywords
func NormalizeKeywords(keywords []string, maxEachLen int) []string {
	if len(keywords) == 0 {
		return nil
	}
	out := make([]string, 0, len(keywords))
	seen := make(map[string]struct{}, len(keywords))
	for _, k := range keywords {
		k = strings.ToLower(SanitizeText(k, maxEachLen))
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	return out
}

// CleanEnrichmentResult sanitizes an enrichment payload
func CleanEnrichmentResult(er datasource.EnrichmentResult) datasource.EnrichmentResult {
	er.CleanBusinessName = SanitizeText(er.CleanBusinessName, 200)
	er.Industry = SanitizeText(er.Industry, 120)
	er.Description = SanitizeText(er.Description, 2000)
	er.Keywords = NormalizeKeywords(er.Keywords, 64)
	return er
}
