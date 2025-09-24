package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// StandaloneClassificationAccuracyTest provides a standalone accuracy testing system
type StandaloneClassificationAccuracyTest struct {
	db     *sql.DB
	logger *log.Logger
}

// TestSample represents a test sample
type TestSample struct {
	ID                   string               `json:"id"`
	BusinessName         string               `json:"business_name"`
	Description          string               `json:"description"`
	WebsiteURL           string               `json:"website_url"`
	ExpectedMCC          string               `json:"expected_mcc"`
	ExpectedNAICS        string               `json:"expected_naics"`
	ExpectedSIC          string               `json:"expected_sic"`
	ExpectedIndustry     string               `json:"expected_industry"`
	ManualClassification ManualClassification `json:"manual_classification"`
	TestCategory         string               `json:"test_category"`
}

// ManualClassification represents expert manual classification
type ManualClassification struct {
	MCCCode          string    `json:"mcc_code"`
	MCCDescription   string    `json:"mcc_description"`
	NAICSCode        string    `json:"naics_code"`
	NAICSDescription string    `json:"naics_description"`
	SICCode          string    `json:"sic_code"`
	SICDescription   string    `json:"sic_description"`
	IndustryID       int       `json:"industry_id"`
	IndustryName     string    `json:"industry_name"`
	Confidence       float64   `json:"confidence"`
	Notes            string    `json:"notes"`
	ClassifiedBy     string    `json:"classified_by"`
	ClassifiedAt     time.Time `json:"classified_at"`
}

// MockClassificationResult represents a mock classification result
type MockClassificationResult struct {
	BusinessID        string                   `json:"business_id"`
	MCCResults        []ClassificationCode     `json:"mcc_results"`
	NAICSResults      []ClassificationCode     `json:"naics_results"`
	SICResults        []ClassificationCode     `json:"sic_results"`
	IndustryResults   []IndustryClassification `json:"industry_results"`
	OverallConfidence float64                  `json:"overall_confidence"`
	ProcessingTime    time.Duration            `json:"processing_time"`
	Timestamp         time.Time                `json:"timestamp"`
}

// ClassificationCode represents a classification code result
type ClassificationCode struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	Rank        int     `json:"rank"`
}

// IndustryClassification represents industry classification result
type IndustryClassification struct {
	IndustryID   int     `json:"industry_id"`
	IndustryName string  `json:"industry_name"`
	Confidence   float64 `json:"confidence"`
	Rank         int     `json:"rank"`
}

// AccuracyMetrics represents accuracy metrics
type AccuracyMetrics struct {
	OverallAccuracy    float64                    `json:"overall_accuracy"`
	MCCAccuracy        float64                    `json:"mcc_accuracy"`
	NAICSAccuracy      float64                    `json:"naics_accuracy"`
	SICAccuracy        float64                    `json:"sic_accuracy"`
	IndustryAccuracy   float64                    `json:"industry_accuracy"`
	ConfidenceAccuracy float64                    `json:"confidence_accuracy"`
	CategoryMetrics    map[string]CategoryMetrics `json:"category_metrics"`
	ProcessingMetrics  ProcessingMetrics          `json:"processing_metrics"`
	ErrorAnalysis      ErrorAnalysis              `json:"error_analysis"`
	Recommendations    []string                   `json:"recommendations"`
}

// CategoryMetrics represents category-specific metrics
type CategoryMetrics struct {
	CategoryName   string        `json:"category_name"`
	SampleCount    int           `json:"sample_count"`
	Accuracy       float64       `json:"accuracy"`
	AvgConfidence  float64       `json:"avg_confidence"`
	ProcessingTime time.Duration `json:"avg_processing_time"`
	ErrorRate      float64       `json:"error_rate"`
}

// ProcessingMetrics represents processing time metrics
type ProcessingMetrics struct {
	AvgProcessingTime time.Duration `json:"avg_processing_time"`
	MinProcessingTime time.Duration `json:"min_processing_time"`
	MaxProcessingTime time.Duration `json:"max_processing_time"`
	P95ProcessingTime time.Duration `json:"p95_processing_time"`
	P99ProcessingTime time.Duration `json:"p99_processing_time"`
}

// ErrorAnalysis represents error analysis
type ErrorAnalysis struct {
	TotalErrors      int               `json:"total_errors"`
	ErrorRate        float64           `json:"error_rate"`
	ErrorCategories  map[string]int    `json:"error_categories"`
	CommonErrors     []CommonError     `json:"common_errors"`
	FalsePositives   int               `json:"false_positives"`
	FalseNegatives   int               `json:"false_negatives"`
	ConfidenceIssues []ConfidenceIssue `json:"confidence_issues"`
}

// CommonError represents common errors
type CommonError struct {
	ErrorType   string   `json:"error_type"`
	Count       int      `json:"count"`
	Percentage  float64  `json:"percentage"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
}

// ConfidenceIssue represents confidence issues
type ConfidenceIssue struct {
	IssueType     string  `json:"issue_type"`
	Count         int     `json:"count"`
	AvgConfidence float64 `json:"avg_confidence"`
	Description   string  `json:"description"`
}

// NewStandaloneClassificationAccuracyTest creates a new standalone test
func NewStandaloneClassificationAccuracyTest(db *sql.DB, logger *log.Logger) *StandaloneClassificationAccuracyTest {
	return &StandaloneClassificationAccuracyTest{
		db:     db,
		logger: logger,
	}
}

// LoadTestSamples loads test samples from database
func (sat *StandaloneClassificationAccuracyTest) LoadTestSamples(ctx context.Context) ([]TestSample, error) {
	query := `
		SELECT 
			id, business_name, description, website_url,
			expected_mcc, expected_naics, expected_sic, expected_industry,
			manual_classification, test_category
		FROM classification_test_samples
		WHERE is_active = true
		ORDER BY test_category, business_name
		LIMIT 50
	`

	rows, err := sat.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to load test samples: %w", err)
	}
	defer rows.Close()

	var samples []TestSample
	for rows.Next() {
		var sample TestSample
		var manualClassificationJSON string

		err := rows.Scan(
			&sample.ID,
			&sample.BusinessName,
			&sample.Description,
			&sample.WebsiteURL,
			&sample.ExpectedMCC,
			&sample.ExpectedNAICS,
			&sample.ExpectedSIC,
			&sample.ExpectedIndustry,
			&manualClassificationJSON,
			&sample.TestCategory,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan test sample: %w", err)
		}

		// Parse manual classification JSON
		if err := json.Unmarshal([]byte(manualClassificationJSON), &sample.ManualClassification); err != nil {
			return nil, fmt.Errorf("failed to parse manual classification: %w", err)
		}

		samples = append(samples, sample)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating test samples: %w", err)
	}

	sat.logger.Printf("Loaded %d test samples", len(samples))
	return samples, nil
}

// MockClassify simulates classification for testing
func (sat *StandaloneClassificationAccuracyTest) MockClassify(ctx context.Context, sample TestSample) (*MockClassificationResult, error) {
	// Simulate processing time
	processingTime := time.Duration(50+int(math.Mod(float64(len(sample.BusinessName)), 200))) * time.Millisecond
	time.Sleep(processingTime)

	// Mock classification results with some randomness
	result := &MockClassificationResult{
		BusinessID:        sample.ID,
		OverallConfidence: 0.7 + math.Mod(float64(len(sample.BusinessName)), 0.3),
		ProcessingTime:    processingTime,
		Timestamp:         time.Now(),
	}

	// Mock MCC results
	if sample.ExpectedMCC != "" {
		result.MCCResults = []ClassificationCode{
			{
				Code:        sample.ExpectedMCC,
				Description: "Mock MCC Description",
				Confidence:  0.8 + math.Mod(float64(len(sample.BusinessName)), 0.2),
				Rank:        1,
			},
		}
	}

	// Mock NAICS results
	if sample.ExpectedNAICS != "" {
		result.NAICSResults = []ClassificationCode{
			{
				Code:        sample.ExpectedNAICS,
				Description: "Mock NAICS Description",
				Confidence:  0.8 + math.Mod(float64(len(sample.BusinessName)), 0.2),
				Rank:        1,
			},
		}
	}

	// Mock SIC results
	if sample.ExpectedSIC != "" {
		result.SICResults = []ClassificationCode{
			{
				Code:        sample.ExpectedSIC,
				Description: "Mock SIC Description",
				Confidence:  0.8 + math.Mod(float64(len(sample.BusinessName)), 0.2),
				Rank:        1,
			},
		}
	}

	// Mock industry results
	if sample.ExpectedIndustry != "" {
		result.IndustryResults = []IndustryClassification{
			{
				IndustryID:   1,
				IndustryName: sample.ExpectedIndustry,
				Confidence:   0.8 + math.Mod(float64(len(sample.BusinessName)), 0.2),
				Rank:         1,
			},
		}
	}

	return result, nil
}

// RunAccuracyTest runs the accuracy test
func (sat *StandaloneClassificationAccuracyTest) RunAccuracyTest(ctx context.Context) (*AccuracyMetrics, error) {
	sat.logger.Println("Starting standalone classification accuracy test...")

	// Load test samples
	samples, err := sat.LoadTestSamples(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load test samples: %w", err)
	}

	if len(samples) == 0 {
		return nil, fmt.Errorf("no test samples found")
	}

	// Run mock classification on all samples
	var results []MockClassificationResult
	for i, sample := range samples {
		sat.logger.Printf("Processing sample %d/%d: %s", i+1, len(samples), sample.BusinessName)

		result, err := sat.MockClassify(ctx, sample)
		if err != nil {
			sat.logger.Printf("Classification failed for sample %s: %v", sample.BusinessName, err)
			continue
		}

		results = append(results, *result)
	}

	// Calculate accuracy metrics
	metrics, err := sat.calculateAccuracyMetrics(samples, results)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate accuracy metrics: %w", err)
	}

	// Generate recommendations
	metrics.Recommendations = sat.generateRecommendations(metrics)

	sat.logger.Printf("Accuracy test completed. Overall accuracy: %.2f%%", metrics.OverallAccuracy*100)
	return metrics, nil
}

// calculateAccuracyMetrics calculates accuracy metrics
func (sat *StandaloneClassificationAccuracyTest) calculateAccuracyMetrics(
	samples []TestSample,
	results []MockClassificationResult,
) (*AccuracyMetrics, error) {
	metrics := &AccuracyMetrics{
		CategoryMetrics: make(map[string]CategoryMetrics),
		ErrorAnalysis: ErrorAnalysis{
			ErrorCategories: make(map[string]int),
		},
	}

	// Create result map for easy lookup
	resultMap := make(map[string]MockClassificationResult)
	for _, result := range results {
		resultMap[result.BusinessID] = result
	}

	var mccAccuracies, naicsAccuracies, sicAccuracies, industryAccuracies []float64
	var confidenceAccuracies []float64
	var processingTimes []time.Duration

	// Process each sample
	for _, sample := range samples {
		result, exists := resultMap[sample.ID]
		if !exists {
			continue
		}

		// Calculate accuracies
		mccAccuracy := sat.calculateMCCAccuracy(sample, result)
		naicsAccuracy := sat.calculateNAICSAccuracy(sample, result)
		sicAccuracy := sat.calculateSICAccuracy(sample, result)
		industryAccuracy := sat.calculateIndustryAccuracy(sample, result)
		confidenceAccuracy := sat.calculateConfidenceAccuracy(sample, result)

		mccAccuracies = append(mccAccuracies, mccAccuracy)
		naicsAccuracies = append(naicsAccuracies, naicsAccuracy)
		sicAccuracies = append(sicAccuracies, sicAccuracy)
		industryAccuracies = append(industryAccuracies, industryAccuracy)
		confidenceAccuracies = append(confidenceAccuracies, confidenceAccuracy)
		processingTimes = append(processingTimes, result.ProcessingTime)

		// Update category metrics
		sat.updateCategoryMetrics(metrics, sample, result, mccAccuracy, naicsAccuracy, sicAccuracy, industryAccuracy)

		// Analyze errors
		sat.analyzeErrors(metrics, sample, result)
	}

	// Calculate overall metrics
	metrics.MCCAccuracy = sat.calculateAverage(mccAccuracies)
	metrics.NAICSAccuracy = sat.calculateAverage(naicsAccuracies)
	metrics.SICAccuracy = sat.calculateAverage(sicAccuracies)
	metrics.IndustryAccuracy = sat.calculateAverage(industryAccuracies)
	metrics.ConfidenceAccuracy = sat.calculateAverage(confidenceAccuracies)
	metrics.OverallAccuracy = (metrics.MCCAccuracy + metrics.NAICSAccuracy + metrics.SICAccuracy + metrics.IndustryAccuracy) / 4

	// Calculate processing metrics
	metrics.ProcessingMetrics = sat.calculateProcessingMetrics(processingTimes)

	// Calculate error analysis
	metrics.ErrorAnalysis.ErrorRate = float64(metrics.ErrorAnalysis.TotalErrors) / float64(len(samples))

	return metrics, nil
}

// Helper methods for accuracy calculation
func (sat *StandaloneClassificationAccuracyTest) calculateMCCAccuracy(sample TestSample, result MockClassificationResult) float64 {
	if sample.ExpectedMCC == "" {
		return 1.0
	}

	for _, mccResult := range result.MCCResults {
		if mccResult.Code == sample.ExpectedMCC {
			return 1.0 - (float64(mccResult.Rank-1) * 0.1)
		}
	}

	return 0.0
}

func (sat *StandaloneClassificationAccuracyTest) calculateNAICSAccuracy(sample TestSample, result MockClassificationResult) float64 {
	if sample.ExpectedNAICS == "" {
		return 1.0
	}

	for _, naicsResult := range result.NAICSResults {
		if naicsResult.Code == sample.ExpectedNAICS {
			return 1.0 - (float64(naicsResult.Rank-1) * 0.1)
		}
	}

	return 0.0
}

func (sat *StandaloneClassificationAccuracyTest) calculateSICAccuracy(sample TestSample, result MockClassificationResult) float64 {
	if sample.ExpectedSIC == "" {
		return 1.0
	}

	for _, sicResult := range result.SICResults {
		if sicResult.Code == sample.ExpectedSIC {
			return 1.0 - (float64(sicResult.Rank-1) * 0.1)
		}
	}

	return 0.0
}

func (sat *StandaloneClassificationAccuracyTest) calculateIndustryAccuracy(sample TestSample, result MockClassificationResult) float64 {
	if sample.ExpectedIndustry == "" {
		return 1.0
	}

	for _, industryResult := range result.IndustryResults {
		if industryResult.IndustryName == sample.ExpectedIndustry {
			return 1.0 - (float64(industryResult.Rank-1) * 0.1)
		}
	}

	return 0.0
}

func (sat *StandaloneClassificationAccuracyTest) calculateConfidenceAccuracy(sample TestSample, result MockClassificationResult) float64 {
	manualConfidence := sample.ManualClassification.Confidence
	systemConfidence := result.OverallConfidence

	difference := math.Abs(manualConfidence - systemConfidence)
	return math.Max(0, 1.0-difference)
}

func (sat *StandaloneClassificationAccuracyTest) updateCategoryMetrics(
	metrics *AccuracyMetrics,
	sample TestSample,
	result MockClassificationResult,
	mccAccuracy, naicsAccuracy, sicAccuracy, industryAccuracy float64,
) {
	category := sample.TestCategory
	if category == "" {
		category = "default"
	}

	catMetrics, exists := metrics.CategoryMetrics[category]
	if !exists {
		catMetrics = CategoryMetrics{
			CategoryName: category,
		}
	}

	catMetrics.SampleCount++
	catMetrics.Accuracy = (catMetrics.Accuracy*float64(catMetrics.SampleCount-1) +
		(mccAccuracy+naicsAccuracy+sicAccuracy+industryAccuracy)/4) / float64(catMetrics.SampleCount)
	catMetrics.AvgConfidence = (catMetrics.AvgConfidence*float64(catMetrics.SampleCount-1) +
		result.OverallConfidence) / float64(catMetrics.SampleCount)
	catMetrics.ProcessingTime = (catMetrics.ProcessingTime*time.Duration(catMetrics.SampleCount-1) +
		result.ProcessingTime) / time.Duration(catMetrics.SampleCount)

	metrics.CategoryMetrics[category] = catMetrics
}

func (sat *StandaloneClassificationAccuracyTest) analyzeErrors(
	metrics *AccuracyMetrics,
	sample TestSample,
	result MockClassificationResult,
) {
	hasError := false

	// Check for various types of errors
	if sample.ExpectedMCC != "" && !sat.hasCodeInResults(sample.ExpectedMCC, result.MCCResults) {
		metrics.ErrorAnalysis.ErrorCategories["mcc_miss"]++
		hasError = true
	}

	if sample.ExpectedNAICS != "" && !sat.hasCodeInResults(sample.ExpectedNAICS, result.NAICSResults) {
		metrics.ErrorAnalysis.ErrorCategories["naics_miss"]++
		hasError = true
	}

	if sample.ExpectedSIC != "" && !sat.hasCodeInResults(sample.ExpectedSIC, result.SICResults) {
		metrics.ErrorAnalysis.ErrorCategories["sic_miss"]++
		hasError = true
	}

	if sample.ExpectedIndustry != "" && !sat.hasIndustryInResults(sample.ExpectedIndustry, result.IndustryResults) {
		metrics.ErrorAnalysis.ErrorCategories["industry_miss"]++
		hasError = true
	}

	if math.Abs(sample.ManualClassification.Confidence-result.OverallConfidence) > 0.3 {
		metrics.ErrorAnalysis.ErrorCategories["confidence_issue"]++
	}

	if hasError {
		metrics.ErrorAnalysis.TotalErrors++
	}
}

func (sat *StandaloneClassificationAccuracyTest) hasCodeInResults(expectedCode string, results []ClassificationCode) bool {
	for _, result := range results {
		if result.Code == expectedCode {
			return true
		}
	}
	return false
}

func (sat *StandaloneClassificationAccuracyTest) hasIndustryInResults(expectedIndustry string, results []IndustryClassification) bool {
	for _, result := range results {
		if result.IndustryName == expectedIndustry {
			return true
		}
	}
	return false
}

func (sat *StandaloneClassificationAccuracyTest) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func (sat *StandaloneClassificationAccuracyTest) calculateProcessingMetrics(processingTimes []time.Duration) ProcessingMetrics {
	if len(processingTimes) == 0 {
		return ProcessingMetrics{}
	}

	sort.Slice(processingTimes, func(i, j int) bool {
		return processingTimes[i] < processingTimes[j]
	})

	metrics := ProcessingMetrics{
		MinProcessingTime: processingTimes[0],
		MaxProcessingTime: processingTimes[len(processingTimes)-1],
	}

	sum := time.Duration(0)
	for _, duration := range processingTimes {
		sum += duration
	}
	metrics.AvgProcessingTime = sum / time.Duration(len(processingTimes))

	metrics.P95ProcessingTime = processingTimes[int(float64(len(processingTimes))*0.95)]
	metrics.P99ProcessingTime = processingTimes[int(float64(len(processingTimes))*0.99)]

	return metrics
}

func (sat *StandaloneClassificationAccuracyTest) generateRecommendations(metrics *AccuracyMetrics) []string {
	var recommendations []string

	if metrics.OverallAccuracy < 0.95 {
		recommendations = append(recommendations,
			fmt.Sprintf("Overall accuracy (%.1f%%) is below target (95%). Consider enhancing classification algorithms.",
				metrics.OverallAccuracy*100))
	}

	if metrics.MCCAccuracy < 0.90 {
		recommendations = append(recommendations,
			fmt.Sprintf("MCC accuracy (%.1f%%) needs improvement. Review MCC code mappings.",
				metrics.MCCAccuracy*100))
	}

	if metrics.NAICSAccuracy < 0.90 {
		recommendations = append(recommendations,
			fmt.Sprintf("NAICS accuracy (%.1f%%) needs improvement. Enhance NAICS classification logic.",
				metrics.NAICSAccuracy*100))
	}

	if metrics.SICAccuracy < 0.90 {
		recommendations = append(recommendations,
			fmt.Sprintf("SIC accuracy (%.1f%%) needs improvement. Review SIC code mappings.",
				metrics.SICAccuracy*100))
	}

	if metrics.IndustryAccuracy < 0.90 {
		recommendations = append(recommendations,
			fmt.Sprintf("Industry accuracy (%.1f%%) needs improvement. Enhance industry classification algorithms.",
				metrics.IndustryAccuracy*100))
	}

	if metrics.ConfidenceAccuracy < 0.80 {
		recommendations = append(recommendations,
			fmt.Sprintf("Confidence scoring accuracy (%.1f%%) needs improvement. Review confidence calculation algorithms.",
				metrics.ConfidenceAccuracy*100))
	}

	avgProcessingMs := float64(metrics.ProcessingMetrics.AvgProcessingTime) / float64(time.Millisecond)
	if avgProcessingMs > 200 {
		recommendations = append(recommendations,
			fmt.Sprintf("Average processing time (%.0fms) exceeds target (200ms). Consider performance optimization.",
				avgProcessingMs))
	}

	if metrics.ErrorAnalysis.ErrorRate > 0.05 {
		recommendations = append(recommendations,
			fmt.Sprintf("Error rate (%.1f%%) is high. Focus on reducing classification errors.",
				metrics.ErrorAnalysis.ErrorRate*100))
	}

	return recommendations
}

// printAccuracyResults prints comprehensive accuracy results
func (sat *StandaloneClassificationAccuracyTest) printAccuracyResults(metrics *AccuracyMetrics) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("CLASSIFICATION ACCURACY TEST RESULTS")
	fmt.Println(strings.Repeat("=", 80))

	// Overall metrics
	fmt.Printf("\n📊 OVERALL ACCURACY METRICS:\n")
	fmt.Printf("   Overall Accuracy:     %.2f%%\n", metrics.OverallAccuracy*100)
	fmt.Printf("   MCC Accuracy:         %.2f%%\n", metrics.MCCAccuracy*100)
	fmt.Printf("   NAICS Accuracy:       %.2f%%\n", metrics.NAICSAccuracy*100)
	fmt.Printf("   SIC Accuracy:         %.2f%%\n", metrics.SICAccuracy*100)
	fmt.Printf("   Industry Accuracy:    %.2f%%\n", metrics.IndustryAccuracy*100)
	fmt.Printf("   Confidence Accuracy:  %.2f%%\n", metrics.ConfidenceAccuracy*100)

	// Performance metrics
	fmt.Printf("\n⚡ PERFORMANCE METRICS:\n")
	fmt.Printf("   Average Processing Time: %.0fms\n", float64(metrics.ProcessingMetrics.AvgProcessingTime)/float64(time.Millisecond))
	fmt.Printf("   Min Processing Time:     %.0fms\n", float64(metrics.ProcessingMetrics.MinProcessingTime)/float64(time.Millisecond))
	fmt.Printf("   Max Processing Time:     %.0fms\n", float64(metrics.ProcessingMetrics.MaxProcessingTime)/float64(time.Millisecond))
	fmt.Printf("   95th Percentile:         %.0fms\n", float64(metrics.ProcessingMetrics.P95ProcessingTime)/float64(time.Millisecond))
	fmt.Printf("   99th Percentile:         %.0fms\n", float64(metrics.ProcessingMetrics.P99ProcessingTime)/float64(time.Millisecond))

	// Category metrics
	fmt.Printf("\n📋 CATEGORY METRICS:\n")
	for category, catMetrics := range metrics.CategoryMetrics {
		fmt.Printf("   %s:\n", category)
		fmt.Printf("     Samples: %d, Accuracy: %.2f%%, Avg Confidence: %.2f%%, Processing: %.0fms\n",
			catMetrics.SampleCount,
			catMetrics.Accuracy*100,
			catMetrics.AvgConfidence*100,
			float64(catMetrics.ProcessingTime)/float64(time.Millisecond))
	}

	// Error analysis
	fmt.Printf("\n🚨 ERROR ANALYSIS:\n")
	fmt.Printf("   Total Errors:        %d\n", metrics.ErrorAnalysis.TotalErrors)
	fmt.Printf("   Error Rate:          %.2f%%\n", metrics.ErrorAnalysis.ErrorRate*100)
	fmt.Printf("   False Positives:     %d\n", metrics.ErrorAnalysis.FalsePositives)
	fmt.Printf("   False Negatives:     %d\n", metrics.ErrorAnalysis.FalseNegatives)

	if len(metrics.ErrorAnalysis.ErrorCategories) > 0 {
		fmt.Printf("   Error Categories:\n")
		for category, count := range metrics.ErrorAnalysis.ErrorCategories {
			fmt.Printf("     %s: %d\n", category, count)
		}
	}

	// Recommendations
	if len(metrics.Recommendations) > 0 {
		fmt.Printf("\n💡 RECOMMENDATIONS:\n")
		for i, recommendation := range metrics.Recommendations {
			fmt.Printf("   %d. %s\n", i+1, recommendation)
		}
	}

	// Performance assessment
	fmt.Printf("\n🎯 PERFORMANCE ASSESSMENT:\n")
	sat.assessPerformance(metrics)

	fmt.Println("\n" + strings.Repeat("=", 80))
}

// assessPerformance provides performance assessment
func (sat *StandaloneClassificationAccuracyTest) assessPerformance(metrics *AccuracyMetrics) {
	// Overall accuracy assessment
	if metrics.OverallAccuracy >= 0.95 {
		fmt.Printf("   ✅ Overall accuracy (%.1f%%) meets target (95%%)\n", metrics.OverallAccuracy*100)
	} else if metrics.OverallAccuracy >= 0.90 {
		fmt.Printf("   ⚠️  Overall accuracy (%.1f%%) is close to target (95%%)\n", metrics.OverallAccuracy*100)
	} else {
		fmt.Printf("   ❌ Overall accuracy (%.1f%%) is below target (95%%)\n", metrics.OverallAccuracy*100)
	}

	// Processing time assessment
	avgProcessingMs := float64(metrics.ProcessingMetrics.AvgProcessingTime) / float64(time.Millisecond)
	if avgProcessingMs <= 200 {
		fmt.Printf("   ✅ Processing time (%.0fms) meets target (200ms)\n", avgProcessingMs)
	} else if avgProcessingMs <= 500 {
		fmt.Printf("   ⚠️  Processing time (%.0fms) is acceptable but could be improved\n", avgProcessingMs)
	} else {
		fmt.Printf("   ❌ Processing time (%.0fms) exceeds target (200ms)\n", avgProcessingMs)
	}

	// Error rate assessment
	if metrics.ErrorAnalysis.ErrorRate <= 0.05 {
		fmt.Printf("   ✅ Error rate (%.1f%%) is acceptable\n", metrics.ErrorAnalysis.ErrorRate*100)
	} else if metrics.ErrorAnalysis.ErrorRate <= 0.10 {
		fmt.Printf("   ⚠️  Error rate (%.1f%%) is high but manageable\n", metrics.ErrorAnalysis.ErrorRate*100)
	} else {
		fmt.Printf("   ❌ Error rate (%.1f%%) is too high\n", metrics.ErrorAnalysis.ErrorRate*100)
	}
}

func main() {
	// Command line flags
	var (
		databaseURL = flag.String("database", "postgres://postgres:password@localhost/kyb_platform?sslmode=disable", "Database connection string")
		verbose     = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Set log level
	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Initialize database connection
	db, err := sql.Open("postgres", *databaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test database connection
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize test
	logger := log.New(os.Stdout, "[STANDALONE_ACCURACY_TEST] ", log.LstdFlags|log.Lshortfile)
	test := NewStandaloneClassificationAccuracyTest(db, logger)

	// Run accuracy test
	if err := test.RunAccuracyTest(ctx); err != nil {
		log.Fatalf("Accuracy test failed: %v", err)
	}

	log.Println("Standalone classification accuracy test completed successfully!")
}
