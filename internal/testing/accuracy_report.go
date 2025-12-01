package testing

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// AccuracyReportGenerator generates detailed accuracy reports
type AccuracyReportGenerator struct {
	logger *log.Logger
}

// NewAccuracyReportGenerator creates a new accuracy report generator
func NewAccuracyReportGenerator(logger *log.Logger) *AccuracyReportGenerator {
	if logger == nil {
		logger = log.Default()
	}
	return &AccuracyReportGenerator{
		logger: logger,
	}
}

// GenerateReport generates a comprehensive accuracy report from metrics
func (arg *AccuracyReportGenerator) GenerateReport(metrics *ComprehensiveAccuracyMetrics) (*AccuracyReport, error) {
	report := &AccuracyReport{
		GeneratedAt:        time.Now(),
		Metrics:            metrics,
		Summary:            arg.generateSummary(metrics),
		CategoryBreakdown:  arg.generateCategoryBreakdown(metrics),
		IndustryBreakdown:  arg.generateIndustryBreakdown(metrics),
		FailureAnalysis:    arg.generateFailureAnalysis(metrics),
		Recommendations:    arg.generateRecommendations(metrics),
	}

	return report, nil
}

// AccuracyReport represents a comprehensive accuracy report
type AccuracyReport struct {
	GeneratedAt        time.Time                        `json:"generated_at"`
	Metrics            *ComprehensiveAccuracyMetrics     `json:"metrics"`
	Summary            *ReportSummary                   `json:"summary"`
	CategoryBreakdown  map[string]*CategoryReport       `json:"category_breakdown"`
	IndustryBreakdown  map[string]*IndustryReport       `json:"industry_breakdown"`
	FailureAnalysis    *FailureAnalysis                 `json:"failure_analysis"`
	Recommendations    []string                         `json:"recommendations"`
}

// ReportSummary provides a high-level summary of the accuracy test results
type ReportSummary struct {
	OverallAccuracy        float64 `json:"overall_accuracy"`
	IndustryAccuracy       float64 `json:"industry_accuracy"`
	CodeAccuracy           float64 `json:"code_accuracy"`
	TotalTestCases         int     `json:"total_test_cases"`
	PassedTestCases        int     `json:"passed_test_cases"`
	FailedTestCases        int     `json:"failed_test_cases"`
	PassRate               float64 `json:"pass_rate"`
	MeetsTarget            bool    `json:"meets_target"`            // 95%+ industry, 90%+ code
	AverageProcessingTime  string  `json:"average_processing_time"`
}

// CategoryReport provides detailed metrics for a specific test category
type CategoryReport struct {
	CategoryName        string  `json:"category_name"`
	TestCount           int     `json:"test_count"`
	OverallAccuracy     float64 `json:"overall_accuracy"`
	IndustryAccuracy    float64 `json:"industry_accuracy"`
	MCCAccuracy         float64 `json:"mcc_accuracy"`
	NAICSAccuracy       float64 `json:"naics_accuracy"`
	SICAccuracy         float64 `json:"sic_accuracy"`
	PassRate            float64 `json:"pass_rate"`
}

// IndustryReport provides detailed metrics for a specific industry
type IndustryReport struct {
	IndustryName        string  `json:"industry_name"`
	TestCount           int     `json:"test_count"`
	OverallAccuracy     float64 `json:"overall_accuracy"`
	IndustryAccuracy    float64 `json:"industry_accuracy"`
	MCCAccuracy         float64 `json:"mcc_accuracy"`
	NAICSAccuracy       float64 `json:"naics_accuracy"`
	SICAccuracy         float64 `json:"sic_accuracy"`
	PassRate            float64 `json:"pass_rate"`
}

// FailureAnalysis provides analysis of test failures
type FailureAnalysis struct {
	TotalFailures           int                      `json:"total_failures"`
	FailureRate             float64                  `json:"failure_rate"`
	CommonFailureReasons    map[string]int           `json:"common_failure_reasons"`
	IndustryFailurePatterns map[string]int           `json:"industry_failure_patterns"`
	CodeTypeFailureRates    map[string]float64      `json:"code_type_failure_rates"`
	SampleFailures          []*FailureSample         `json:"sample_failures"`
}

// FailureSample represents a sample failure case
type FailureSample struct {
	BusinessName         string   `json:"business_name"`
	ExpectedIndustry     string   `json:"expected_industry"`
	ActualIndustry       string   `json:"actual_industry"`
	ExpectedMCCCodes     []string `json:"expected_mcc_codes"`
	ActualMCCCodes       []string `json:"actual_mcc_codes"`
	ExpectedNAICSCodes   []string `json:"expected_naics_codes"`
	ActualNAICSCodes     []string `json:"actual_naics_codes"`
	ExpectedSICCodes     []string `json:"expected_sic_codes"`
	ActualSICCodes       []string `json:"actual_sic_codes"`
	FailureReason        string   `json:"failure_reason"`
}

// generateSummary generates a high-level summary
func (arg *AccuracyReportGenerator) generateSummary(metrics *ComprehensiveAccuracyMetrics) *ReportSummary {
	passRate := 0.0
	if metrics.TotalTestCases > 0 {
		passRate = float64(metrics.PassedTestCases) / float64(metrics.TotalTestCases)
	}

	meetsTarget := metrics.IndustryAccuracy >= 0.95 && metrics.CodeAccuracy >= 0.90

	return &ReportSummary{
		OverallAccuracy:       metrics.OverallAccuracy,
		IndustryAccuracy:      metrics.IndustryAccuracy,
		CodeAccuracy:          metrics.CodeAccuracy,
		TotalTestCases:        metrics.TotalTestCases,
		PassedTestCases:       metrics.PassedTestCases,
		FailedTestCases:       metrics.FailedTestCases,
		PassRate:              passRate,
		MeetsTarget:           meetsTarget,
		AverageProcessingTime: metrics.AverageProcessingTime.String(),
	}
}

// generateCategoryBreakdown generates category-specific breakdown
func (arg *AccuracyReportGenerator) generateCategoryBreakdown(metrics *ComprehensiveAccuracyMetrics) map[string]*CategoryReport {
	breakdown := make(map[string]*CategoryReport)

	// Group results by category
	categoryResults := make(map[string][]*AccuracyTestResult)
	for _, result := range metrics.TestResults {
		if result.Error == "" {
			categoryResults[result.TestCategory] = append(categoryResults[result.TestCategory], result)
		}
	}

	// Calculate metrics per category
	for category, results := range categoryResults {
		if len(results) == 0 {
			continue
		}

		var totalIndustryAccuracy float64
		var totalMCCAccuracy float64
		var totalNAICSAccuracy float64
		var totalSICAccuracy float64
		var totalOverallAccuracy float64
		passed := 0

		for _, result := range results {
			totalIndustryAccuracy += result.IndustryAccuracy
			totalMCCAccuracy += result.MCCAccuracy
			totalNAICSAccuracy += result.NAICSAccuracy
			totalSICAccuracy += result.SICAccuracy
			totalOverallAccuracy += result.OverallAccuracy
			if result.OverallAccuracy >= 0.80 { // 80% threshold for pass
				passed++
			}
		}

		count := float64(len(results))
		breakdown[category] = &CategoryReport{
			CategoryName:     category,
			TestCount:        len(results),
			OverallAccuracy:  totalOverallAccuracy / count,
			IndustryAccuracy: totalIndustryAccuracy / count,
			MCCAccuracy:      totalMCCAccuracy / count,
			NAICSAccuracy:    totalNAICSAccuracy / count,
			SICAccuracy:      totalSICAccuracy / count,
			PassRate:         float64(passed) / count,
		}
	}

	return breakdown
}

// generateIndustryBreakdown generates industry-specific breakdown
func (arg *AccuracyReportGenerator) generateIndustryBreakdown(metrics *ComprehensiveAccuracyMetrics) map[string]*IndustryReport {
	breakdown := make(map[string]*IndustryReport)

	// Group results by expected industry
	industryResults := make(map[string][]*AccuracyTestResult)
	for _, result := range metrics.TestResults {
		if result.Error == "" {
			industryResults[result.ExpectedIndustry] = append(industryResults[result.ExpectedIndustry], result)
		}
	}

	// Calculate metrics per industry
	for industry, results := range industryResults {
		if len(results) == 0 {
			continue
		}

		var totalIndustryAccuracy float64
		var totalMCCAccuracy float64
		var totalNAICSAccuracy float64
		var totalSICAccuracy float64
		var totalOverallAccuracy float64
		passed := 0

		for _, result := range results {
			totalIndustryAccuracy += result.IndustryAccuracy
			totalMCCAccuracy += result.MCCAccuracy
			totalNAICSAccuracy += result.NAICSAccuracy
			totalSICAccuracy += result.SICAccuracy
			totalOverallAccuracy += result.OverallAccuracy
			if result.OverallAccuracy >= 0.80 {
				passed++
			}
		}

		count := float64(len(results))
		breakdown[industry] = &IndustryReport{
			IndustryName:     industry,
			TestCount:        len(results),
			OverallAccuracy: totalOverallAccuracy / count,
			IndustryAccuracy: totalIndustryAccuracy / count,
			MCCAccuracy:      totalMCCAccuracy / count,
			NAICSAccuracy:    totalNAICSAccuracy / count,
			SICAccuracy:      totalSICAccuracy / count,
			PassRate:         float64(passed) / count,
		}
	}

	return breakdown
}

// generateFailureAnalysis generates failure analysis
func (arg *AccuracyReportGenerator) generateFailureAnalysis(metrics *ComprehensiveAccuracyMetrics) *FailureAnalysis {
	analysis := &FailureAnalysis{
		TotalFailures:           metrics.FailedTestCases,
		CommonFailureReasons:    make(map[string]int),
		IndustryFailurePatterns: make(map[string]int),
		CodeTypeFailureRates:    make(map[string]float64),
		SampleFailures:          []*FailureSample{},
	}

	if metrics.TotalTestCases > 0 {
		analysis.FailureRate = float64(metrics.FailedTestCases) / float64(metrics.TotalTestCases)
	}

	// Analyze failures
	failureReasons := make(map[string]int)
	industryFailures := make(map[string]int)
	var mccFailures, naicsFailures, sicFailures int
	var mccTotal, naicsTotal, sicTotal int

	for _, result := range metrics.TestResults {
		if result.Error != "" {
			failureReasons[result.Error]++
			industryFailures[result.ExpectedIndustry]++
			continue
		}

		// Track code type failures
		if len(result.ExpectedMCCCodes) > 0 {
			mccTotal++
			if result.MCCAccuracy < 0.5 {
				mccFailures++
			}
		}
		if len(result.ExpectedNAICSCodes) > 0 {
			naicsTotal++
			if result.NAICSAccuracy < 0.5 {
				naicsFailures++
			}
		}
		if len(result.ExpectedSICCodes) > 0 {
			sicTotal++
			if result.SICAccuracy < 0.5 {
				sicFailures++
			}
		}

		// Collect sample failures (overall accuracy < 0.80)
		if result.OverallAccuracy < 0.80 && len(analysis.SampleFailures) < 10 {
			failureReason := "Low overall accuracy"
			if !result.IndustryMatch {
				failureReason = "Industry mismatch"
			} else if result.MCCAccuracy < 0.5 || result.NAICSAccuracy < 0.5 || result.SICAccuracy < 0.5 {
				failureReason = "Code accuracy below threshold"
			}

			analysis.SampleFailures = append(analysis.SampleFailures, &FailureSample{
				BusinessName:       result.BusinessName,
				ExpectedIndustry:   result.ExpectedIndustry,
				ActualIndustry:     result.ActualIndustry,
				ExpectedMCCCodes:   result.ExpectedMCCCodes,
				ActualMCCCodes:     result.ActualMCCCodes,
				ExpectedNAICSCodes: result.ExpectedNAICSCodes,
				ActualNAICSCodes:   result.ActualNAICSCodes,
				ExpectedSICCodes:   result.ExpectedSICCodes,
				ActualSICCodes:     result.ActualSICCodes,
				FailureReason:      failureReason,
			})
		}
	}

	analysis.CommonFailureReasons = failureReasons
	analysis.IndustryFailurePatterns = industryFailures

	if mccTotal > 0 {
		analysis.CodeTypeFailureRates["MCC"] = float64(mccFailures) / float64(mccTotal)
	}
	if naicsTotal > 0 {
		analysis.CodeTypeFailureRates["NAICS"] = float64(naicsFailures) / float64(naicsTotal)
	}
	if sicTotal > 0 {
		analysis.CodeTypeFailureRates["SIC"] = float64(sicFailures) / float64(sicTotal)
	}

	return analysis
}

// generateRecommendations generates recommendations based on metrics
func (arg *AccuracyReportGenerator) generateRecommendations(metrics *ComprehensiveAccuracyMetrics) []string {
	recommendations := []string{}

	// Industry accuracy recommendations
	if metrics.IndustryAccuracy < 0.95 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Industry accuracy (%.2f%%) is below target (95%%). Consider improving keyword matching and industry detection algorithms.", 
				metrics.IndustryAccuracy*100))
	}

	// Code accuracy recommendations
	if metrics.CodeAccuracy < 0.90 {
		recommendations = append(recommendations,
			fmt.Sprintf("Code accuracy (%.2f%%) is below target (90%%). Review code generation logic and keyword-to-code mappings.",
				metrics.CodeAccuracy*100))
	}

	// Specific code type recommendations
	if metrics.MCCAccuracy < 0.85 {
		recommendations = append(recommendations,
			fmt.Sprintf("MCC code accuracy (%.2f%%) needs improvement. Enhance MCC keyword matching and crosswalk data.",
				metrics.MCCAccuracy*100))
	}
	if metrics.NAICSAccuracy < 0.85 {
		recommendations = append(recommendations,
			fmt.Sprintf("NAICS code accuracy (%.2f%%) needs improvement. Review NAICS hierarchy and keyword mappings.",
				metrics.NAICSAccuracy*100))
	}
	if metrics.SICAccuracy < 0.85 {
		recommendations = append(recommendations,
			fmt.Sprintf("SIC code accuracy (%.2f%%) needs improvement. Enhance SIC keyword matching and crosswalk data.",
				metrics.SICAccuracy*100))
	}

	// Edge case recommendations
	if metrics.EdgeCaseAccuracy > 0 && metrics.EdgeCaseAccuracy < 0.70 {
		recommendations = append(recommendations,
			"Edge case accuracy is low. Consider adding more edge case training data and improving ambiguous business name handling.")
	}

	// Performance recommendations
	if metrics.AverageProcessingTime > 2*time.Second {
		recommendations = append(recommendations,
			fmt.Sprintf("Average processing time (%.2fs) is high. Consider optimizing classification algorithms and database queries.",
				metrics.AverageProcessingTime.Seconds()))
	}

	// Category-specific recommendations
	worstCategory := ""
	worstAccuracy := 1.0
	for category, accuracy := range metrics.AccuracyByCategory {
		if accuracy < worstAccuracy {
			worstAccuracy = accuracy
			worstCategory = category
		}
	}
	if worstCategory != "" && worstAccuracy < 0.80 {
		recommendations = append(recommendations,
			fmt.Sprintf("Category '%s' has low accuracy (%.2f%%). Review test cases and improve classification for this category.",
				worstCategory, worstAccuracy*100))
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "✅ All accuracy targets met! System is performing well.")
	}

	return recommendations
}

// GenerateJSONReport generates a JSON-formatted report
func (arg *AccuracyReportGenerator) GenerateJSONReport(metrics *ComprehensiveAccuracyMetrics) ([]byte, error) {
	report, err := arg.GenerateReport(metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal report: %w", err)
	}

	return jsonData, nil
}

// GenerateTextReport generates a human-readable text report
func (arg *AccuracyReportGenerator) GenerateTextReport(metrics *ComprehensiveAccuracyMetrics) (string, error) {
	report, err := arg.GenerateReport(metrics)
	if err != nil {
		return "", fmt.Errorf("failed to generate report: %w", err)
	}

	var sb strings.Builder

	sb.WriteString("=" + strings.Repeat("=", 78) + "\n")
	sb.WriteString("COMPREHENSIVE ACCURACY TEST REPORT\n")
	sb.WriteString("=" + strings.Repeat("=", 78) + "\n")
	sb.WriteString(fmt.Sprintf("Generated At: %s\n\n", report.GeneratedAt.Format(time.RFC3339)))

	// Summary
	sb.WriteString("SUMMARY\n")
	sb.WriteString(strings.Repeat("-", 80) + "\n")
	sb.WriteString(fmt.Sprintf("Overall Accuracy:     %.2f%%\n", report.Summary.OverallAccuracy*100))
	sb.WriteString(fmt.Sprintf("Industry Accuracy:    %.2f%% (Target: 95%%)\n", report.Summary.IndustryAccuracy*100))
	sb.WriteString(fmt.Sprintf("Code Accuracy:        %.2f%% (Target: 90%%)\n", report.Summary.CodeAccuracy*100))
	sb.WriteString(fmt.Sprintf("Total Test Cases:     %d\n", report.Summary.TotalTestCases))
	sb.WriteString(fmt.Sprintf("Passed:               %d\n", report.Summary.PassedTestCases))
	sb.WriteString(fmt.Sprintf("Failed:               %d\n", report.Summary.FailedTestCases))
	sb.WriteString(fmt.Sprintf("Pass Rate:            %.2f%%\n", report.Summary.PassRate*100))
	sb.WriteString(fmt.Sprintf("Meets Target:         %v\n", report.Summary.MeetsTarget))
	sb.WriteString(fmt.Sprintf("Avg Processing Time:   %s\n\n", report.Summary.AverageProcessingTime))

	// Category Breakdown
	if len(report.CategoryBreakdown) > 0 {
		sb.WriteString("CATEGORY BREAKDOWN\n")
		sb.WriteString(strings.Repeat("-", 80) + "\n")
		for category, catReport := range report.CategoryBreakdown {
			sb.WriteString(fmt.Sprintf("\n%s:\n", category))
			sb.WriteString(fmt.Sprintf("  Test Count:        %d\n", catReport.TestCount))
			sb.WriteString(fmt.Sprintf("  Overall Accuracy:   %.2f%%\n", catReport.OverallAccuracy*100))
			sb.WriteString(fmt.Sprintf("  Industry Accuracy:  %.2f%%\n", catReport.IndustryAccuracy*100))
			sb.WriteString(fmt.Sprintf("  MCC Accuracy:       %.2f%%\n", catReport.MCCAccuracy*100))
			sb.WriteString(fmt.Sprintf("  NAICS Accuracy:     %.2f%%\n", catReport.NAICSAccuracy*100))
			sb.WriteString(fmt.Sprintf("  SIC Accuracy:       %.2f%%\n", catReport.SICAccuracy*100))
			sb.WriteString(fmt.Sprintf("  Pass Rate:          %.2f%%\n", catReport.PassRate*100))
		}
		sb.WriteString("\n")
	}

	// Recommendations
	if len(report.Recommendations) > 0 {
		sb.WriteString("RECOMMENDATIONS\n")
		sb.WriteString(strings.Repeat("-", 80) + "\n")
		for i, rec := range report.Recommendations {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

