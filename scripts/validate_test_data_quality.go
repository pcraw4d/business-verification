package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// TestSample represents a test sample
type TestSample struct {
	ID               string   `json:"id"`
	BusinessName     string   `json:"business_name"`
	Description      string   `json:"description"`
	WebsiteURL       string   `json:"website_url"`
	ExpectedIndustry string   `json:"expected_industry"`
	ExpectedMCC      []string `json:"expected_mcc_codes"`
	ExpectedNAICS    []string `json:"expected_naics_codes"`
	ExpectedSIC      []string `json:"expected_sic_codes"`
	Category         string   `json:"category"`
	Complexity       string   `json:"complexity"`
	ScrapingDifficulty string `json:"scraping_difficulty"`
}

// ValidationIssue represents a data quality issue
type ValidationIssue struct {
	SampleID    string
	IssueType   string
	Issue       string
	Severity    string
	Field       string
	Value       string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run validate_test_data_quality.go <path_to_test_data.json>")
		fmt.Println("Example: go run validate_test_data_quality.go test/data/comprehensive_test_samples.json")
		os.Exit(1)
	}

	filePath := os.Args[1]
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	var testData struct {
		Samples []TestSample `json:"samples"`
	}

	if err := json.Unmarshal(data, &testData); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Test Data Quality Validation ===\n")
	fmt.Printf("Total samples: %d\n\n", len(testData.Samples))

	var issues []ValidationIssue
	issueCounts := make(map[string]int)

	// Validate each sample
	for _, sample := range testData.Samples {
		sampleIssues := validateSample(sample)
		issues = append(issues, sampleIssues...)
		for _, issue := range sampleIssues {
			issueCounts[issue.IssueType]++
		}
	}

	// Print summary
	fmt.Println("=== Validation Summary ===")
	fmt.Printf("Total issues found: %d\n\n", len(issues))

	fmt.Println("=== Issue Distribution ===")
	for issueType, count := range issueCounts {
		fmt.Printf("%s: %d\n", issueType, count)
	}
	fmt.Println()

	// Print detailed issues by severity
	criticalIssues := filterIssues(issues, "CRITICAL")
	highIssues := filterIssues(issues, "HIGH")
	mediumIssues := filterIssues(issues, "MEDIUM")
	lowIssues := filterIssues(issues, "LOW")

	if len(criticalIssues) > 0 {
		fmt.Println("=== CRITICAL Issues ===")
		for _, issue := range criticalIssues {
			fmt.Printf("Sample %s: %s - %s (Field: %s, Value: %s)\n",
				issue.SampleID, issue.IssueType, issue.Issue, issue.Field, issue.Value)
		}
		fmt.Println()
	}

	if len(highIssues) > 0 {
		fmt.Println("=== HIGH Issues ===")
		for _, issue := range highIssues {
			fmt.Printf("Sample %s: %s - %s (Field: %s, Value: %s)\n",
				issue.SampleID, issue.IssueType, issue.Issue, issue.Field, issue.Value)
		}
		fmt.Println()
	}

	if len(mediumIssues) > 0 {
		fmt.Println("=== MEDIUM Issues (showing first 10) ===")
		for i, issue := range mediumIssues {
			if i >= 10 {
				fmt.Printf("... and %d more\n", len(mediumIssues)-10)
				break
			}
			fmt.Printf("Sample %s: %s - %s (Field: %s, Value: %s)\n",
				issue.SampleID, issue.IssueType, issue.Issue, issue.Field, issue.Value)
		}
		fmt.Println()
	}

	// Generate report
	generateReport(issues, issueCounts, len(testData.Samples))
}

func validateSample(sample TestSample) []ValidationIssue {
	var issues []ValidationIssue

	// Validate ID
	if sample.ID == "" {
		issues = append(issues, ValidationIssue{
			SampleID:  sample.ID,
			IssueType: "missing_id",
			Issue:     "Sample ID is missing",
			Severity:  "CRITICAL",
			Field:     "id",
			Value:     "",
		})
	}

	// Validate business name
	if sample.BusinessName == "" {
		issues = append(issues, ValidationIssue{
			SampleID:  sample.ID,
			IssueType: "missing_business_name",
			Issue:     "Business name is missing",
			Severity:  "CRITICAL",
			Field:     "business_name",
			Value:     "",
		})
	}

	// Validate website URL
	if sample.WebsiteURL != "" {
		urlIssues := validateURL(sample.WebsiteURL, sample.ID)
		issues = append(issues, urlIssues...)
	}

	// Validate expected industry
	if sample.ExpectedIndustry == "" {
		issues = append(issues, ValidationIssue{
			SampleID:  sample.ID,
			IssueType: "missing_expected_industry",
			Issue:     "Expected industry is missing",
			Severity:  "HIGH",
			Field:     "expected_industry",
			Value:     "",
		})
	}

	// Validate expected codes
	if len(sample.ExpectedMCC) == 0 && len(sample.ExpectedNAICS) == 0 && len(sample.ExpectedSIC) == 0 {
		issues = append(issues, ValidationIssue{
			SampleID:  sample.ID,
			IssueType: "missing_expected_codes",
			Issue:     "No expected codes provided (MCC, NAICS, or SIC)",
			Severity:  "MEDIUM",
			Field:     "expected_codes",
			Value:     "",
		})
	}

	// Validate code formats
	for _, mcc := range sample.ExpectedMCC {
		if !isValidMCC(mcc) {
			issues = append(issues, ValidationIssue{
				SampleID:  sample.ID,
				IssueType: "invalid_mcc_format",
				Issue:     fmt.Sprintf("Invalid MCC code format: %s", mcc),
				Severity:  "MEDIUM",
				Field:     "expected_mcc_codes",
				Value:     mcc,
			})
		}
	}

	for _, naics := range sample.ExpectedNAICS {
		if !isValidNAICS(naics) {
			issues = append(issues, ValidationIssue{
				SampleID:  sample.ID,
				IssueType: "invalid_naics_format",
				Issue:     fmt.Sprintf("Invalid NAICS code format: %s", naics),
				Severity:  "MEDIUM",
				Field:     "expected_naics_codes",
				Value:     naics,
			})
		}
	}

	for _, sic := range sample.ExpectedSIC {
		if !isValidSIC(sic) {
			issues = append(issues, ValidationIssue{
				SampleID:  sample.ID,
				IssueType: "invalid_sic_format",
				Issue:     fmt.Sprintf("Invalid SIC code format: %s", sic),
				Severity:  "MEDIUM",
				Field:     "expected_sic_codes",
				Value:     sic,
			})
		}
	}

	return issues
}

func validateURL(urlStr string, sampleID string) []ValidationIssue {
	var issues []ValidationIssue

	// Check for malformed URLs
	if strings.Contains(urlStr, "&") && !strings.Contains(urlStr, "?") {
		// URL contains & but no ? (likely malformed)
		issues = append(issues, ValidationIssue{
			SampleID:  sampleID,
			IssueType: "malformed_url",
			Issue:     fmt.Sprintf("URL contains '&' without '?': %s", urlStr),
			Severity:  "HIGH",
			Field:     "website_url",
			Value:     urlStr,
		})
	}

	// Try to parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		issues = append(issues, ValidationIssue{
			SampleID:  sampleID,
			IssueType: "invalid_url",
			Issue:     fmt.Sprintf("URL cannot be parsed: %s", err.Error()),
			Severity:  "CRITICAL",
			Field:     "website_url",
			Value:     urlStr,
		})
		return issues
	}

	// Check for missing scheme
	if parsedURL.Scheme == "" {
		issues = append(issues, ValidationIssue{
			SampleID:  sampleID,
			IssueType: "missing_url_scheme",
			Issue:     "URL missing scheme (http:// or https://)",
			Severity:  "HIGH",
			Field:     "website_url",
			Value:     urlStr,
		})
	}

	// Check for invalid characters
	invalidChars := []string{" ", "\n", "\t", "\r"}
	for _, char := range invalidChars {
		if strings.Contains(urlStr, char) {
			issues = append(issues, ValidationIssue{
				SampleID:  sampleID,
				IssueType: "invalid_url_characters",
				Issue:     fmt.Sprintf("URL contains invalid character: %s", char),
				Severity:  "HIGH",
				Field:     "website_url",
				Value:     urlStr,
			})
		}
	}

	// Check for common malformed patterns
	malformedPatterns := []struct {
		pattern string
		issue   string
	}{
		{`www\.modernarts&`, "Malformed URL pattern: www.modernarts&"},
		{`&entertainment`, "URL contains '&' in domain name"},
		{`http://http://`, "Double http:// scheme"},
		{`https://https://`, "Double https:// scheme"},
	}

	for _, pattern := range malformedPatterns {
		matched, _ := regexp.MatchString(pattern.pattern, urlStr)
		if matched {
			issues = append(issues, ValidationIssue{
				SampleID:  sampleID,
				IssueType: "malformed_url_pattern",
				Issue:     pattern.issue,
				Severity:  "HIGH",
				Field:     "website_url",
				Value:     urlStr,
			})
		}
	}

	return issues
}

func isValidMCC(code string) bool {
	// MCC codes are typically 4 digits
	matched, _ := regexp.MatchString(`^\d{4}$`, code)
	return matched
}

func isValidNAICS(code string) bool {
	// NAICS codes are typically 5-6 digits
	matched, _ := regexp.MatchString(`^\d{5,6}$`, code)
	return matched
}

func isValidSIC(code string) bool {
	// SIC codes are typically 4 digits
	matched, _ := regexp.MatchString(`^\d{4}$`, code)
	return matched
}

func filterIssues(issues []ValidationIssue, severity string) []ValidationIssue {
	var filtered []ValidationIssue
	for _, issue := range issues {
		if issue.Severity == severity {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

func generateReport(issues []ValidationIssue, issueCounts map[string]int, totalSamples int) {
	report := fmt.Sprintf(`# Test Data Quality Validation Report

**Date**: %s
**Total Samples**: %d
**Total Issues**: %d

## Issue Summary

| Issue Type | Count | Percentage |
|------------|-------|------------|
`, time.Now().Format("2006-01-02 15:04:05"), totalSamples, len(issues))

	for issueType, count := range issueCounts {
		percentage := float64(count) / float64(totalSamples) * 100
		report += fmt.Sprintf("| %s | %d | %.1f%% |\n", issueType, count, percentage)
	}

	report += "\n## Recommendations\n\n"

	if issueCounts["malformed_url"] > 0 || issueCounts["invalid_url"] > 0 {
		report += "1. **Fix Malformed URLs**: Clean up URLs with invalid characters or malformed patterns\n"
	}

	if issueCounts["missing_expected_industry"] > 0 {
		report += "2. **Add Expected Industries**: Ensure all samples have expected industry values\n"
	}

	if issueCounts["missing_expected_codes"] > 0 {
		report += "3. **Add Expected Codes**: Ensure all samples have at least one expected code (MCC, NAICS, or SIC)\n"
	}

	if issueCounts["invalid_mcc_format"] > 0 || issueCounts["invalid_naics_format"] > 0 || issueCounts["invalid_sic_format"] > 0 {
		report += "4. **Fix Code Formats**: Validate and fix code format issues\n"
	}

	report += "\n## Detailed Issues\n\n"

	// Group by severity
	severities := []string{"CRITICAL", "HIGH", "MEDIUM", "LOW"}
	for _, severity := range severities {
		severityIssues := filterIssues(issues, severity)
		if len(severityIssues) > 0 {
			report += fmt.Sprintf("### %s Issues (%d)\n\n", severity, len(severityIssues))
			for _, issue := range severityIssues {
				report += fmt.Sprintf("- **Sample %s**: %s - %s (Field: %s, Value: %s)\n",
					issue.SampleID, issue.IssueType, issue.Issue, issue.Field, issue.Value)
			}
			report += "\n"
		}
	}

	outputPath := "docs/test-data-quality-audit.md"
	err := os.WriteFile(outputPath, []byte(report), 0644)
	if err != nil {
		fmt.Printf("Error writing report: %v\n", err)
		return
	}

	fmt.Printf("\nDetailed report written to: %s\n", outputPath)
}

