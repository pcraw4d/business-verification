package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// TestResult represents a single test result
type TestResult struct {
	ExpectedIndustry string  `json:"expected_industry"`
	ActualIndustry   string  `json:"actual_industry"`
	Confidence       float64 `json:"confidence"`
	BusinessName     string  `json:"business_name"`
	Description      string  `json:"description"`
	WebsiteURL       string  `json:"website_url"`
	Success          bool    `json:"success"`
	Error            string  `json:"error,omitempty"`
}

// ClassificationMetrics represents the metrics structure
type ClassificationMetrics struct {
	AverageConfidence    float64            `json:"average_confidence"`
	ClassificationAccuracy float64          `json:"classification_accuracy"`
	IndustryAccuracy     map[string]float64 `json:"industry_accuracy"`
}

// TestData represents the full test data structure
type TestData struct {
	ClassificationMetrics ClassificationMetrics `json:"classification_metrics"`
	Results               []TestResult          `json:"results,omitempty"`
}

// MisclassificationPattern represents a pattern of misclassifications
type MisclassificationPattern struct {
	ExpectedIndustry string
	ActualIndustry   string
	Count            int
	Examples         []string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run analyze_classification_accuracy.go <test_results.json>")
		fmt.Println("Example: go run analyze_classification_accuracy.go test/integration/test/results/railway_e2e_classification_20251222_093927.json")
		os.Exit(1)
	}

	filePath := os.Args[1]
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	var testData TestData
	if err := json.Unmarshal(data, &testData); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Extract results if available
	var results []TestResult
	if testData.Results != nil {
		results = testData.Results
	} else {
		// Try to extract from a different structure
		var rawData map[string]interface{}
		if err := json.Unmarshal(data, &rawData); err == nil {
			if resultsRaw, ok := rawData["results"].([]interface{}); ok {
				for _, r := range resultsRaw {
					if rMap, ok := r.(map[string]interface{}); ok {
						result := TestResult{}
						if expected, ok := rMap["expected_industry"].(string); ok {
							result.ExpectedIndustry = expected
						}
						if actual, ok := rMap["actual_industry"].(string); ok {
							result.ActualIndustry = actual
						}
						if conf, ok := rMap["confidence"].(float64); ok {
							result.Confidence = conf
						}
						if name, ok := rMap["business_name"].(string); ok {
							result.BusinessName = name
						}
						if desc, ok := rMap["description"].(string); ok {
							result.Description = desc
						}
						if url, ok := rMap["website_url"].(string); ok {
							result.WebsiteURL = url
						}
						if success, ok := rMap["success"].(bool); ok {
							result.Success = success
						}
						results = append(results, result)
					}
				}
			}
		}
	}

	// Analyze patterns
	fmt.Println("=== Classification Accuracy Analysis ===\n")
	fmt.Printf("Total Results: %d\n", len(results))
	fmt.Printf("Average Confidence: %.2f%%\n", testData.ClassificationMetrics.AverageConfidence*100)
	fmt.Printf("Overall Accuracy: %.2f%%\n", testData.ClassificationMetrics.ClassificationAccuracy*100)
	fmt.Println()

	// Industry accuracy breakdown
	fmt.Println("=== Industry Accuracy Breakdown ===")
	industryAcc := testData.ClassificationMetrics.IndustryAccuracy
	type accStat struct {
		industry string
		accuracy float64
	}
	accStats := []accStat{}
	for industry, accuracy := range industryAcc {
		accStats = append(accStats, accStat{industry, accuracy})
	}
	sort.Slice(accStats, func(i, j int) bool {
		return accStats[i].accuracy < accStats[j].accuracy
	})

	for _, stat := range accStats {
		status := "✅"
		if stat.accuracy < 0.5 {
			status = "❌"
		} else if stat.accuracy < 0.8 {
			status = "⚠️"
		}
		fmt.Printf("%s %s: %.1f%%\n", status, stat.industry, stat.accuracy*100)
	}
	fmt.Println()

	// Analyze misclassification patterns
	if len(results) > 0 {
		analyzeMisclassifications(results, testData)
	}

	// Write detailed report
	writeDetailedReport(testData, results, filePath)
}

func analyzeMisclassifications(results []TestResult, testData TestData) {
	fmt.Println("=== Misclassification Patterns ===")

	// Count correct vs incorrect
	correct := 0
	incorrect := 0
	empty := 0

	// Pattern tracking
	patterns := make(map[string]*MisclassificationPattern)

	for _, result := range results {
		expected := normalizeIndustry(result.ExpectedIndustry)
		actual := normalizeIndustry(result.ActualIndustry)

		if actual == "" {
			empty++
			continue
		}

		if expected == actual {
			correct++
		} else {
			incorrect++
			patternKey := fmt.Sprintf("%s -> %s", expected, actual)
			if patterns[patternKey] == nil {
				patterns[patternKey] = &MisclassificationPattern{
					ExpectedIndustry: expected,
					ActualIndustry:   actual,
					Examples:         []string{},
				}
			}
			patterns[patternKey].Count++
			if len(patterns[patternKey].Examples) < 3 {
				example := result.BusinessName
				if example == "" {
					example = result.Description[:50] + "..."
				}
				patterns[patternKey].Examples = append(patterns[patternKey].Examples, example)
			}
		}
	}

	fmt.Printf("Correct: %d (%.1f%%)\n", correct, float64(correct)/float64(len(results))*100)
	fmt.Printf("Incorrect: %d (%.1f%%)\n", incorrect, float64(incorrect)/float64(len(results))*100)
	fmt.Printf("Empty/Unknown: %d (%.1f%%)\n", empty, float64(empty)/float64(len(results))*100)
	fmt.Println()

	// Top misclassification patterns
	type patternStat struct {
		pattern string
		count   int
	}
	patternStats := []patternStat{}
	for patternKey, pattern := range patterns {
		patternStats = append(patternStats, patternStat{patternKey, pattern.Count})
	}
	sort.Slice(patternStats, func(i, j int) bool {
		return patternStats[i].count > patternStats[j].count
	})

	fmt.Println("Top Misclassification Patterns:")
	for i, stat := range patternStats {
		if i >= 10 {
			break
		}
		pattern := patterns[stat.pattern]
		fmt.Printf("  %d. %s (%d occurrences)\n", i+1, stat.pattern, stat.count)
		if len(pattern.Examples) > 0 {
			fmt.Printf("     Examples: %s\n", strings.Join(pattern.Examples, ", "))
		}
	}
	fmt.Println()

	// Industries with 0% accuracy
	fmt.Println("Industries with 0% Accuracy:")
	industryAcc := testData.ClassificationMetrics.IndustryAccuracy
	for industry, accuracy := range industryAcc {
		if accuracy == 0 {
			fmt.Printf("  - %s\n", industry)
		}
	}
}

func normalizeIndustry(industry string) string {
	industry = strings.ToLower(strings.TrimSpace(industry))
	// Normalize common variations
	replacements := map[string]string{
		"arts & entertainment": "arts and entertainment",
		"food & beverage":      "food and beverage",
		"financial services":   "banking",
		"professional services": "professional services",
	}
	if normalized, ok := replacements[industry]; ok {
		return normalized
	}
	return industry
}

func writeDetailedReport(testData TestData, results []TestResult, filePath string) {
	reportFile := "docs/classification-accuracy-analysis.md"
	file, err := os.Create(reportFile)
	if err != nil {
		fmt.Printf("Warning: Could not create report file: %v\n", err)
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "# Classification Accuracy Analysis\n\n")
	fmt.Fprintf(file, "Generated from: %s\n\n", filePath)
	fmt.Fprintf(file, "## Summary\n\n")
	fmt.Fprintf(file, "- **Overall Accuracy**: %.2f%%\n", testData.ClassificationMetrics.ClassificationAccuracy*100)
	fmt.Fprintf(file, "- **Average Confidence**: %.2f%%\n", testData.ClassificationMetrics.AverageConfidence*100)
	fmt.Fprintf(file, "- **Total Results**: %d\n\n", len(results))

	fmt.Fprintf(file, "## Industry Accuracy Breakdown\n\n")
	fmt.Fprintf(file, "| Industry | Accuracy | Status |\n")
	fmt.Fprintf(file, "|----------|----------|--------|\n")
	industryAcc := testData.ClassificationMetrics.IndustryAccuracy
	for industry, accuracy := range industryAcc {
		status := "✅"
		if accuracy < 0.5 {
			status = "❌"
		} else if accuracy < 0.8 {
			status = "⚠️"
		}
		fmt.Fprintf(file, "| %s | %.1f%% | %s |\n", industry, accuracy*100, status)
	}

	fmt.Fprintf(file, "\n## Key Findings\n\n")
	fmt.Fprintf(file, "1. **Industries with 0%% Accuracy**:\n")
	for industry, accuracy := range industryAcc {
		if accuracy == 0 {
			fmt.Fprintf(file, "   - %s\n", industry)
		}
	}

	fmt.Fprintf(file, "\n2. **Industries with Low Accuracy (<50%%)**:\n")
	for industry, accuracy := range industryAcc {
		if accuracy > 0 && accuracy < 0.5 {
			fmt.Fprintf(file, "   - %s: %.1f%%\n", industry, accuracy*100)
		}
	}

	fmt.Printf("\nDetailed report written to: %s\n", reportFile)
}

