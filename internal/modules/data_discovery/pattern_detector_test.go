package data_discovery

import (
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAddressRegexPattern(t *testing.T) {
	// Test the exact regex pattern from the pattern detector
	addressPattern := `\d+\s+[A-Za-z\s]+(?:Ave|Street|St|Avenue|Road|Rd|Boulevard|Blvd|Lane|Ln|Drive|Dr)[,\s]+[A-Za-z\s]+(?:Suite\s+\d+)?[,\s]+[A-Z]{2}\s+\d{5}(?:-\d{4})?`

	regex := regexp.MustCompile(addressPattern)

	testAddress := "123 Business Ave, Suite 100, Anytown, NY 12345"

	matches := regex.FindAllString(testAddress, -1)

	t.Logf("Testing address: %s", testAddress)
	t.Logf("Regex pattern: %s", addressPattern)
	t.Logf("Found %d matches: %v", len(matches), matches)

	// Also test without "Address:" prefix
	testAddress2 := "123 Business Ave, Suite 100, Anytown, NY 12345"
	matches2 := regex.FindAllString(testAddress2, -1)
	t.Logf("Testing address (no prefix): %s", testAddress2)
	t.Logf("Found %d matches: %v", len(matches2), matches2)

	// Test a simpler pattern
	simplePattern := `\d+\s+[A-Za-z\s]+Ave[,\s]+[A-Za-z\s]+[,\s]+[A-Z]{2}\s+\d{5}`
	simpleRegex := regexp.MustCompile(simplePattern)
	simpleMatches := simpleRegex.FindAllString(testAddress, -1)
	t.Logf("Simple pattern: %s", simplePattern)
	t.Logf("Found %d matches: %v", len(simpleMatches), simpleMatches)

	// Test step by step
	t.Logf("=== Step by step debugging ===")

	// Test just the number and street name
	step1Pattern := `\d+\s+[A-Za-z\s]+Ave`
	step1Regex := regexp.MustCompile(step1Pattern)
	step1Matches := step1Regex.FindAllString(testAddress, -1)
	t.Logf("Step 1 (number + street): %s", step1Pattern)
	t.Logf("Found %d matches: %v", len(step1Matches), step1Matches)

	// Test with city
	step2Pattern := `\d+\s+[A-Za-z\s]+Ave[,\s]+[A-Za-z\s]+`
	step2Regex := regexp.MustCompile(step2Pattern)
	step2Matches := step2Regex.FindAllString(testAddress, -1)
	t.Logf("Step 2 (with city): %s", step2Pattern)
	t.Logf("Found %d matches: %v", len(step2Matches), step2Matches)

	// Test with state and zip
	step3Pattern := `\d+\s+[A-Za-z\s]+Ave[,\s]+[A-Za-z\s]+[,\s]+[A-Z]{2}\s+\d{5}`
	step3Regex := regexp.MustCompile(step3Pattern)
	step3Matches := step3Regex.FindAllString(testAddress, -1)
	t.Logf("Step 3 (with state/zip): %s", step3Pattern)
	t.Logf("Found %d matches: %v", len(step3Matches), step3Matches)

	// Test the exact string we're looking for
	exactPattern := `123 Business Ave, Suite 100, Anytown, NY 12345`
	exactRegex := regexp.MustCompile(regexp.QuoteMeta(exactPattern))
	exactMatches := exactRegex.FindAllString(testAddress, -1)
	t.Logf("Exact match: %s", exactPattern)
	t.Logf("Found %d matches: %v", len(exactMatches), exactMatches)

	// Test the new flexible pattern
	newPattern := `\d+\s+[A-Za-z\s]+(?:Ave|Street|St|Avenue|Road|Rd|Boulevard|Blvd|Lane|Ln|Drive|Dr)[,\s]+(?:[A-Za-z\s]+(?:Suite\s+\d+)?[,\s]+)?[A-Za-z\s]+[,\s]+[A-Z]{2}\s+\d{5}(?:-\d{4})?`
	newRegex := regexp.MustCompile(newPattern)
	newMatches := newRegex.FindAllString(testAddress, -1)
	t.Logf("New flexible pattern: %s", newPattern)
	t.Logf("Found %d matches: %v", len(newMatches), newMatches)
}

func TestPatternDetector_AddressDetection(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	detector := NewPatternDetector(config, logger)

	// Test the exact address from the integration test
	content := &ContentInput{
		RawContent:  "Address: 123 Business Ave, Suite 100, Anytown, NY 12345",
		ContentType: "text",
	}

	ctx := context.Background()
	matches, err := detector.DetectPatterns(ctx, content)

	assert.NoError(t, err)

	t.Logf("Found %d pattern matches:", len(matches))
	for _, match := range matches {
		t.Logf("  - %s: %s (confidence: %.2f)", match.PatternID, match.MatchedText, match.ConfidenceScore)
	}

	// Check if address was detected
	var addressFound bool
	for _, match := range matches {
		if match.FieldType == "address" {
			addressFound = true
			t.Logf("Address found: %s", match.MatchedText)
			break
		}
	}

	assert.True(t, addressFound, "Address pattern should be detected")
}

func TestPatternDetector_AllPatterns(t *testing.T) {
	config := DefaultDataDiscoveryConfig()
	logger := zap.NewNop()
	detector := NewPatternDetector(config, logger)

	// Test with the full content from integration test
	content := &ContentInput{
		RawContent: `
			Acme Corporation
			Contact: info@acme.com | Phone: (555) 123-4567
			Address: 123 Business Ave, Suite 100, Anytown, NY 12345
			Website: https://www.acme.com
			Follow us: https://facebook.com/acme | https://twitter.com/acme
			Business Hours: Monday-Friday 9:00 AM - 6:00 PM
			Services: Software Development, Consulting, Training
			Founded: 2010
			EIN: 12-3456789
		`,
		ContentType: "html",
	}

	ctx := context.Background()
	matches, err := detector.DetectPatterns(ctx, content)

	assert.NoError(t, err)

	t.Logf("Found %d pattern matches:", len(matches))
	for _, match := range matches {
		t.Logf("  - %s (%s): %s (confidence: %.2f)",
			match.PatternID, match.FieldType, match.MatchedText, match.ConfidenceScore)
	}

	// Check for expected field types
	fieldTypes := make(map[string]bool)
	for _, match := range matches {
		fieldTypes[match.FieldType] = true
	}

	t.Logf("Found field types: %v", fieldTypes)

	expectedTypes := []string{"email", "phone", "address", "url", "social_media"}
	for _, expected := range expectedTypes {
		assert.True(t, fieldTypes[expected], "Should detect %s pattern", expected)
	}
}
