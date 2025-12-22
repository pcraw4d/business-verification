package main

import (
	"encoding/json"
	"fmt"
	"os"
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
	Method           string  `json:"method,omitempty"`
}

// ConfidenceAnalysis represents analysis of confidence scores
type ConfidenceAnalysis struct {
	HighConfidence   []TestResult // >= 0.70
	MediumConfidence []TestResult // 0.30 - 0.69
	LowConfidence    []TestResult // < 0.30
	AverageConfidence float64
	TotalResults     int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run analyze_confidence_scores.go <test_results.json>")
		fmt.Println("Example: go run analyze_confidence_scores.go test/integration/test/results/railway_e2e_classification_20251222_093927.json")
		os.Exit(1)
	}

	filePath := os.Args[1]
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Extract results
	var results []TestResult
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
				if method, ok := rMap["method"].(string); ok {
					result.Method = method
				}
				results = append(results, result)
			}
		}
	}

	// Analyze confidence scores
	analysis := analyzeConfidence(results)

	// Print analysis
	fmt.Println("=== Confidence Score Analysis ===\n")
	fmt.Printf("Total Results: %d\n", analysis.TotalResults)
	fmt.Printf("Average Confidence: %.2f%%\n", analysis.AverageConfidence*100)
	fmt.Println()

	fmt.Printf("High Confidence (≥70%%): %d (%.1f%%)\n", 
		len(analysis.HighConfidence), 
		float64(len(analysis.HighConfidence))/float64(analysis.TotalResults)*100)
	fmt.Printf("Medium Confidence (30-69%%): %d (%.1f%%)\n", 
		len(analysis.MediumConfidence), 
		float64(len(analysis.MediumConfidence))/float64(analysis.TotalResults)*100)
	fmt.Printf("Low Confidence (<30%%): %d (%.1f%%)\n", 
		len(analysis.LowConfidence), 
		float64(len(analysis.LowConfidence))/float64(analysis.TotalResults)*100)
	fmt.Println()

	// Compare high vs low confidence accuracy
	fmt.Println("=== Accuracy by Confidence Level ===")
	highAcc := calculateAccuracy(analysis.HighConfidence)
	mediumAcc := calculateAccuracy(analysis.MediumConfidence)
	lowAcc := calculateAccuracy(analysis.LowConfidence)

	fmt.Printf("High Confidence Accuracy: %.1f%%\n", highAcc*100)
	fmt.Printf("Medium Confidence Accuracy: %.1f%%\n", mediumAcc*100)
	fmt.Printf("Low Confidence Accuracy: %.1f%%\n", lowAcc*100)
	fmt.Println()

	// Method distribution
	fmt.Println("=== Classification Method Distribution ===")
	methodDist := calculateMethodDistribution(results)
	for method, count := range methodDist {
		fmt.Printf("%s: %d (%.1f%%)\n", method, count, float64(count)/float64(analysis.TotalResults)*100)
	}
	fmt.Println()

	// Confidence distribution by method
	fmt.Println("=== Average Confidence by Method ===")
	methodConf := calculateMethodConfidence(results)
	for method, conf := range methodConf {
		fmt.Printf("%s: %.2f%%\n", method, conf*100)
	}
	fmt.Println()

	// Write detailed report
	writeConfidenceReport(analysis, results, filePath)
}

func analyzeConfidence(results []TestResult) ConfidenceAnalysis {
	analysis := ConfidenceAnalysis{
		HighConfidence:   []TestResult{},
		MediumConfidence: []TestResult{},
		LowConfidence:    []TestResult{},
		TotalResults:     len(results),
	}

	totalConf := 0.0
	for _, result := range results {
		totalConf += result.Confidence

		if result.Confidence >= 0.70 {
			analysis.HighConfidence = append(analysis.HighConfidence, result)
		} else if result.Confidence >= 0.30 {
			analysis.MediumConfidence = append(analysis.MediumConfidence, result)
		} else {
			analysis.LowConfidence = append(analysis.LowConfidence, result)
		}
	}

	if len(results) > 0 {
		analysis.AverageConfidence = totalConf / float64(len(results))
	}

	return analysis
}

func calculateAccuracy(results []TestResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	correct := 0
	for _, result := range results {
		expected := normalizeIndustry(result.ExpectedIndustry)
		actual := normalizeIndustry(result.ActualIndustry)
		if expected == actual && actual != "" {
			correct++
		}
	}

	return float64(correct) / float64(len(results))
}

func calculateMethodDistribution(results []TestResult) map[string]int {
	dist := make(map[string]int)
	for _, result := range results {
		method := result.Method
		if method == "" {
			method = "unknown"
		}
		dist[method]++
	}
	return dist
}

func calculateMethodConfidence(results []TestResult) map[string]float64 {
	methodConf := make(map[string]float64)
	methodCount := make(map[string]int)

	for _, result := range results {
		method := result.Method
		if method == "" {
			method = "unknown"
		}
		methodConf[method] += result.Confidence
		methodCount[method]++
	}

	for method := range methodConf {
		if count := methodCount[method]; count > 0 {
			methodConf[method] /= float64(count)
		}
	}

	return methodConf
}

func normalizeIndustry(industry string) string {
	industry = strings.ToLower(strings.TrimSpace(industry))
	replacements := map[string]string{
		"arts & entertainment": "arts and entertainment",
		"food & beverage":      "food and beverage",
		"financial services":   "banking",
	}
	if normalized, ok := replacements[industry]; ok {
		return normalized
	}
	return industry
}

func writeConfidenceReport(analysis ConfidenceAnalysis, results []TestResult, filePath string) {
	reportFile := "docs/confidence-score-analysis.md"
	file, err := os.Create(reportFile)
	if err != nil {
		fmt.Printf("Warning: Could not create report file: %v\n", err)
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "# Confidence Score Analysis\n\n")
	fmt.Fprintf(file, "Generated from: %s\n\n", filePath)
	fmt.Fprintf(file, "## Summary\n\n")
	fmt.Fprintf(file, "- **Total Results**: %d\n", analysis.TotalResults)
	fmt.Fprintf(file, "- **Average Confidence**: %.2f%%\n", analysis.AverageConfidence*100)
	fmt.Fprintf(file, "- **Target**: >70%%\n\n")

	fmt.Fprintf(file, "## Confidence Distribution\n\n")
	fmt.Fprintf(file, "| Level | Count | Percentage | Accuracy |\n")
	fmt.Fprintf(file, "|-------|-------|------------|----------|\n")
	
	highAcc := calculateAccuracy(analysis.HighConfidence)
	mediumAcc := calculateAccuracy(analysis.MediumConfidence)
	lowAcc := calculateAccuracy(analysis.LowConfidence)

	fmt.Fprintf(file, "| High (≥70%%) | %d | %.1f%% | %.1f%% |\n", 
		len(analysis.HighConfidence), 
		float64(len(analysis.HighConfidence))/float64(analysis.TotalResults)*100,
		highAcc*100)
	fmt.Fprintf(file, "| Medium (30-69%%) | %d | %.1f%% | %.1f%% |\n", 
		len(analysis.MediumConfidence), 
		float64(len(analysis.MediumConfidence))/float64(analysis.TotalResults)*100,
		mediumAcc*100)
	fmt.Fprintf(file, "| Low (<30%%) | %d | %.1f%% | %.1f%% |\n", 
		len(analysis.LowConfidence), 
		float64(len(analysis.LowConfidence))/float64(analysis.TotalResults)*100,
		lowAcc*100)

	fmt.Fprintf(file, "\n## Key Findings\n\n")
	fmt.Fprintf(file, "1. **Average confidence is %.2f%%** (target: >70%%)\n", analysis.AverageConfidence*100)
	fmt.Fprintf(file, "2. **Only %.1f%% of results have high confidence** (≥70%%)\n", 
		float64(len(analysis.HighConfidence))/float64(analysis.TotalResults)*100)
	fmt.Fprintf(file, "3. **%.1f%% of results have low confidence** (<30%%)\n", 
		float64(len(analysis.LowConfidence))/float64(analysis.TotalResults)*100)

	fmt.Printf("\nDetailed report written to: %s\n", reportFile)
}

