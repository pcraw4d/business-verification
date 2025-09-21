package testing

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"time"
)

// ClassificationAccuracyTester provides comprehensive testing for classification accuracy
type ClassificationAccuracyTester struct {
	db     *sql.DB
	logger *log.Logger
}

// TestSample represents a known business sample for testing
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
	TestCategory         string               `json:"test_category"` // primary, edge_case, high_risk, etc.
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
	Confidence       float64   `json:"confidence"` // Manual confidence (0-1)
	Notes            string    `json:"notes"`
	ClassifiedBy     string    `json:"classified_by"`
	ClassifiedAt     time.Time `json:"classified_at"`
}

// ClassificationResult represents the system's classification result
type ClassificationResult struct {
	BusinessID        string                   `json:"business_id"`
	MCCResults        []ClassificationCode     `json:"mcc_results"`
	NAICSResults      []ClassificationCode     `json:"naics_results"`
	SICResults        []ClassificationCode     `json:"sic_results"`
	IndustryResults   []IndustryClassification `json:"industry_results"`
	OverallConfidence float64                  `json:"overall_confidence"`
	ProcessingTime    time.Duration            `json:"processing_time"`
	Timestamp         time.Time                `json:"timestamp"`
}

// ClassificationCode represents a single classification code result
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

// AccuracyMetrics represents comprehensive accuracy metrics
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

// CategoryMetrics represents accuracy metrics for specific test categories
type CategoryMetrics struct {
	CategoryName   string        `json:"category_name"`
	SampleCount    int           `json:"sample_count"`
	Accuracy       float64       `json:"accuracy"`
	AvgConfidence  float64       `json:"avg_confidence"`
	ProcessingTime time.Duration `json:"avg_processing_time"`
	ErrorRate      float64       `json:"error_rate"`
}

// ProcessingMetrics represents system performance metrics
type ProcessingMetrics struct {
	AvgProcessingTime time.Duration `json:"avg_processing_time"`
	MinProcessingTime time.Duration `json:"min_processing_time"`
	MaxProcessingTime time.Duration `json:"max_processing_time"`
	P95ProcessingTime time.Duration `json:"p95_processing_time"`
	P99ProcessingTime time.Duration `json:"p99_processing_time"`
}

// ErrorAnalysis represents detailed error analysis
type ErrorAnalysis struct {
	TotalErrors      int               `json:"total_errors"`
	ErrorRate        float64           `json:"error_rate"`
	ErrorCategories  map[string]int    `json:"error_categories"`
	CommonErrors     []CommonError     `json:"common_errors"`
	FalsePositives   int               `json:"false_positives"`
	FalseNegatives   int               `json:"false_negatives"`
	ConfidenceIssues []ConfidenceIssue `json:"confidence_issues"`
}

// CommonError represents frequently occurring errors
type CommonError struct {
	ErrorType   string   `json:"error_type"`
	Count       int      `json:"count"`
	Percentage  float64  `json:"percentage"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
}

// ConfidenceIssue represents confidence scoring issues
type ConfidenceIssue struct {
	IssueType     string  `json:"issue_type"`
	Count         int     `json:"count"`
	AvgConfidence float64 `json:"avg_confidence"`
	Description   string  `json:"description"`
}

// NewClassificationAccuracyTester creates a new accuracy tester
func NewClassificationAccuracyTester(db *sql.DB, logger *log.Logger) *ClassificationAccuracyTester {
	return &ClassificationAccuracyTester{
		db:     db,
		logger: logger,
	}
}

// LoadTestSamples loads test samples from the database
func (cat *ClassificationAccuracyTester) LoadTestSamples(ctx context.Context) ([]TestSample, error) {
	query := `
		SELECT 
			id, business_name, description, website_url,
			expected_mcc, expected_naics, expected_sic, expected_industry,
			manual_classification, test_category
		FROM classification_test_samples
		WHERE is_active = true
		ORDER BY test_category, business_name
	`

	rows, err := cat.db.QueryContext(ctx, query)
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

	cat.logger.Printf("Loaded %d test samples", len(samples))
	return samples, nil
}

// RunAccuracyTest runs comprehensive accuracy testing
func (cat *ClassificationAccuracyTester) RunAccuracyTest(ctx context.Context, classifier Classifier) (*AccuracyMetrics, error) {
	cat.logger.Println("Starting classification accuracy test...")

	// Load test samples
	samples, err := cat.LoadTestSamples(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load test samples: %w", err)
	}

	if len(samples) == 0 {
		return nil, fmt.Errorf("no test samples found")
	}

	// Run classification on all samples
	results, err := cat.runClassificationOnSamples(ctx, classifier, samples)
	if err != nil {
		return nil, fmt.Errorf("failed to run classification: %w", err)
	}

	// Calculate accuracy metrics
	metrics, err := cat.calculateAccuracyMetrics(samples, results)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate accuracy metrics: %w", err)
	}

	// Generate recommendations
	metrics.Recommendations = cat.generateRecommendations(metrics)

	cat.logger.Printf("Accuracy test completed. Overall accuracy: %.2f%%", metrics.OverallAccuracy*100)
	return metrics, nil
}

// runClassificationOnSamples runs classification on all test samples
func (cat *ClassificationAccuracyTester) runClassificationOnSamples(
	ctx context.Context,
	classifier Classifier,
	samples []TestSample,
) ([]ClassificationResult, error) {
	var results []ClassificationResult

	for i, sample := range samples {
		cat.logger.Printf("Processing sample %d/%d: %s", i+1, len(samples), sample.BusinessName)

		startTime := time.Now()

		// Run classification
		result, err := classifier.Classify(ctx, sample.BusinessName, sample.Description, sample.WebsiteURL)
		if err != nil {
			cat.logger.Printf("Classification failed for sample %s: %v", sample.BusinessName, err)
			// Continue with other samples
			continue
		}

		processingTime := time.Since(startTime)

		// Convert to our result format
		classificationResult := ClassificationResult{
			BusinessID:        sample.ID,
			MCCResults:        result.MCCResults,
			NAICSResults:      result.NAICSResults,
			SICResults:        result.SICResults,
			IndustryResults:   result.IndustryResults,
			OverallConfidence: result.OverallConfidence,
			ProcessingTime:    processingTime,
			Timestamp:         time.Now(),
		}

		results = append(results, classificationResult)
	}

	return results, nil
}

// calculateAccuracyMetrics calculates comprehensive accuracy metrics
func (cat *ClassificationAccuracyTester) calculateAccuracyMetrics(
	samples []TestSample,
	results []ClassificationResult,
) (*AccuracyMetrics, error) {
	metrics := &AccuracyMetrics{
		CategoryMetrics: make(map[string]CategoryMetrics),
		ErrorAnalysis: ErrorAnalysis{
			ErrorCategories: make(map[string]int),
		},
	}

	// Create result map for easy lookup
	resultMap := make(map[string]ClassificationResult)
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
			continue // Skip samples that failed classification
		}

		// Calculate accuracies for each classification type
		mccAccuracy := cat.calculateMCCAccuracy(sample, result)
		naicsAccuracy := cat.calculateNAICSAccuracy(sample, result)
		sicAccuracy := cat.calculateSICAccuracy(sample, result)
		industryAccuracy := cat.calculateIndustryAccuracy(sample, result)
		confidenceAccuracy := cat.calculateConfidenceAccuracy(sample, result)

		mccAccuracies = append(mccAccuracies, mccAccuracy)
		naicsAccuracies = append(naicsAccuracies, naicsAccuracy)
		sicAccuracies = append(sicAccuracies, sicAccuracy)
		industryAccuracies = append(industryAccuracies, industryAccuracy)
		confidenceAccuracies = append(confidenceAccuracies, confidenceAccuracy)
		processingTimes = append(processingTimes, result.ProcessingTime)

		// Update category metrics
		cat.updateCategoryMetrics(metrics, sample, result, mccAccuracy, naicsAccuracy, sicAccuracy, industryAccuracy)

		// Analyze errors
		cat.analyzeErrors(metrics, sample, result)
	}

	// Calculate overall metrics
	metrics.MCCAccuracy = cat.calculateAverage(mccAccuracies)
	metrics.NAICSAccuracy = cat.calculateAverage(naicsAccuracies)
	metrics.SICAccuracy = cat.calculateAverage(sicAccuracies)
	metrics.IndustryAccuracy = cat.calculateAverage(industryAccuracies)
	metrics.ConfidenceAccuracy = cat.calculateAverage(confidenceAccuracies)
	metrics.OverallAccuracy = (metrics.MCCAccuracy + metrics.NAICSAccuracy + metrics.SICAccuracy + metrics.IndustryAccuracy) / 4

	// Calculate processing metrics
	metrics.ProcessingMetrics = cat.calculateProcessingMetrics(processingTimes)

	// Calculate error analysis
	metrics.ErrorAnalysis.ErrorRate = float64(metrics.ErrorAnalysis.TotalErrors) / float64(len(samples))

	return metrics, nil
}

// calculateMCCAccuracy calculates MCC classification accuracy
func (cat *ClassificationAccuracyTester) calculateMCCAccuracy(sample TestSample, result ClassificationResult) float64 {
	if sample.ExpectedMCC == "" {
		return 1.0 // No expected MCC, consider as correct
	}

	// Check if expected MCC is in top 3 results
	for _, mccResult := range result.MCCResults {
		if mccResult.Code == sample.ExpectedMCC {
			// Higher rank (lower number) gets higher score
			return 1.0 - (float64(mccResult.Rank-1) * 0.1)
		}
	}

	return 0.0 // Expected MCC not found in results
}

// calculateNAICSAccuracy calculates NAICS classification accuracy
func (cat *ClassificationAccuracyTester) calculateNAICSAccuracy(sample TestSample, result ClassificationResult) float64 {
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

// calculateSICAccuracy calculates SIC classification accuracy
func (cat *ClassificationAccuracyTester) calculateSICAccuracy(sample TestSample, result ClassificationResult) float64 {
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

// calculateIndustryAccuracy calculates industry classification accuracy
func (cat *ClassificationAccuracyTester) calculateIndustryAccuracy(sample TestSample, result ClassificationResult) float64 {
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

// calculateConfidenceAccuracy calculates confidence scoring accuracy
func (cat *ClassificationAccuracyTester) calculateConfidenceAccuracy(sample TestSample, result ClassificationResult) float64 {
	// Compare system confidence with manual confidence
	manualConfidence := sample.ManualClassification.Confidence
	systemConfidence := result.OverallConfidence

	// Calculate accuracy as 1 - absolute difference
	difference := math.Abs(manualConfidence - systemConfidence)
	return math.Max(0, 1.0-difference)
}

// updateCategoryMetrics updates metrics for specific test categories
func (cat *ClassificationAccuracyTester) updateCategoryMetrics(
	metrics *AccuracyMetrics,
	sample TestSample,
	result ClassificationResult,
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

// analyzeErrors analyzes classification errors
func (cat *ClassificationAccuracyTester) analyzeErrors(
	metrics *AccuracyMetrics,
	sample TestSample,
	result ClassificationResult,
) {
	// Check for various types of errors
	hasError := false

	// MCC error
	if sample.ExpectedMCC != "" && !cat.hasCodeInResults(sample.ExpectedMCC, result.MCCResults) {
		metrics.ErrorAnalysis.ErrorCategories["mcc_miss"]++
		hasError = true
	}

	// NAICS error
	if sample.ExpectedNAICS != "" && !cat.hasCodeInResults(sample.ExpectedNAICS, result.NAICSResults) {
		metrics.ErrorAnalysis.ErrorCategories["naics_miss"]++
		hasError = true
	}

	// SIC error
	if sample.ExpectedSIC != "" && !cat.hasCodeInResults(sample.ExpectedSIC, result.SICResults) {
		metrics.ErrorAnalysis.ErrorCategories["sic_miss"]++
		hasError = true
	}

	// Industry error
	if sample.ExpectedIndustry != "" && !cat.hasIndustryInResults(sample.ExpectedIndustry, result.IndustryResults) {
		metrics.ErrorAnalysis.ErrorCategories["industry_miss"]++
		hasError = true
	}

	// Confidence issues
	if math.Abs(sample.ManualClassification.Confidence-result.OverallConfidence) > 0.3 {
		metrics.ErrorAnalysis.ErrorCategories["confidence_issue"]++
	}

	if hasError {
		metrics.ErrorAnalysis.TotalErrors++
	}
}

// hasCodeInResults checks if a code exists in classification results
func (cat *ClassificationAccuracyTester) hasCodeInResults(expectedCode string, results []ClassificationCode) bool {
	for _, result := range results {
		if result.Code == expectedCode {
			return true
		}
	}
	return false
}

// hasIndustryInResults checks if an industry exists in classification results
func (cat *ClassificationAccuracyTester) hasIndustryInResults(expectedIndustry string, results []IndustryClassification) bool {
	for _, result := range results {
		if result.IndustryName == expectedIndustry {
			return true
		}
	}
	return false
}

// calculateAverage calculates the average of a slice of float64 values
func (cat *ClassificationAccuracyTester) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

// calculateProcessingMetrics calculates processing time metrics
func (cat *ClassificationAccuracyTester) calculateProcessingMetrics(processingTimes []time.Duration) ProcessingMetrics {
	if len(processingTimes) == 0 {
		return ProcessingMetrics{}
	}

	// Sort processing times for percentile calculations
	sort.Slice(processingTimes, func(i, j int) bool {
		return processingTimes[i] < processingTimes[j]
	})

	metrics := ProcessingMetrics{
		MinProcessingTime: processingTimes[0],
		MaxProcessingTime: processingTimes[len(processingTimes)-1],
	}

	// Calculate average
	sum := time.Duration(0)
	for _, duration := range processingTimes {
		sum += duration
	}
	metrics.AvgProcessingTime = sum / time.Duration(len(processingTimes))

	// Calculate percentiles
	metrics.P95ProcessingTime = processingTimes[int(float64(len(processingTimes))*0.95)]
	metrics.P99ProcessingTime = processingTimes[int(float64(len(processingTimes))*0.99)]

	return metrics
}

// generateRecommendations generates improvement recommendations based on metrics
func (cat *ClassificationAccuracyTester) generateRecommendations(metrics *AccuracyMetrics) []string {
	var recommendations []string

	// Overall accuracy recommendations
	if metrics.OverallAccuracy < 0.95 {
		recommendations = append(recommendations,
			fmt.Sprintf("Overall accuracy (%.1f%%) is below target (95%). Consider enhancing keyword database and classification algorithms.",
				metrics.OverallAccuracy*100))
	}

	// Individual classification type recommendations
	if metrics.MCCAccuracy < 0.90 {
		recommendations = append(recommendations,
			fmt.Sprintf("MCC accuracy (%.1f%%) needs improvement. Review MCC code mappings and keyword associations.",
				metrics.MCCAccuracy*100))
	}

	if metrics.NAICSAccuracy < 0.90 {
		recommendations = append(recommendations,
			fmt.Sprintf("NAICS accuracy (%.1f%%) needs improvement. Enhance NAICS code classification logic.",
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

	// Confidence scoring recommendations
	if metrics.ConfidenceAccuracy < 0.80 {
		recommendations = append(recommendations,
			fmt.Sprintf("Confidence scoring accuracy (%.1f%%) needs improvement. Review confidence calculation algorithms.",
				metrics.ConfidenceAccuracy*100))
	}

	// Processing time recommendations
	if metrics.ProcessingMetrics.AvgProcessingTime > 200*time.Millisecond {
		recommendations = append(recommendations,
			fmt.Sprintf("Average processing time (%.0fms) exceeds target (200ms). Consider performance optimization.",
				float64(metrics.ProcessingMetrics.AvgProcessingTime)/float64(time.Millisecond)))
	}

	// Error analysis recommendations
	if metrics.ErrorAnalysis.ErrorRate > 0.05 {
		recommendations = append(recommendations,
			fmt.Sprintf("Error rate (%.1f%%) is high. Focus on reducing classification errors.",
				metrics.ErrorAnalysis.ErrorRate*100))
	}

	// Category-specific recommendations
	for category, catMetrics := range metrics.CategoryMetrics {
		if catMetrics.Accuracy < 0.90 {
			recommendations = append(recommendations,
				fmt.Sprintf("Category '%s' accuracy (%.1f%%) needs improvement. Review classification logic for this category.",
					category, catMetrics.Accuracy*100))
		}
	}

	return recommendations
}

// convertToClassificationCodes converts generic results to ClassificationCode format
func (cat *ClassificationAccuracyTester) convertToClassificationCodes(results []interface{}) []ClassificationCode {
	var codes []ClassificationCode
	for i, result := range results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			code := ClassificationCode{
				Code:        fmt.Sprintf("%v", resultMap["code"]),
				Description: fmt.Sprintf("%v", resultMap["description"]),
				Confidence:  resultMap["confidence"].(float64),
				Rank:        i + 1,
			}
			codes = append(codes, code)
		}
	}
	return codes
}

// convertToIndustryClassifications converts generic results to IndustryClassification format
func (cat *ClassificationAccuracyTester) convertToIndustryClassifications(results []interface{}) []IndustryClassification {
	var industries []IndustryClassification
	for i, result := range results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			industry := IndustryClassification{
				IndustryID:   int(resultMap["industry_id"].(float64)),
				IndustryName: fmt.Sprintf("%v", resultMap["industry_name"]),
				Confidence:   resultMap["confidence"].(float64),
				Rank:         i + 1,
			}
			industries = append(industries, industry)
		}
	}
	return industries
}

// Classifier interface for classification systems
type Classifier interface {
	Classify(ctx context.Context, businessName, description, websiteURL string) (*ClassificationResult, error)
}

// SaveAccuracyReport saves accuracy test results to database
func (cat *ClassificationAccuracyTester) SaveAccuracyReport(ctx context.Context, metrics *AccuracyMetrics) error {
	query := `
		INSERT INTO classification_accuracy_reports (
			overall_accuracy, mcc_accuracy, naics_accuracy, sic_accuracy, 
			industry_accuracy, confidence_accuracy, category_metrics,
			processing_metrics, error_analysis, recommendations, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	categoryMetricsJSON, _ := json.Marshal(metrics.CategoryMetrics)
	processingMetricsJSON, _ := json.Marshal(metrics.ProcessingMetrics)
	errorAnalysisJSON, _ := json.Marshal(metrics.ErrorAnalysis)
	recommendationsJSON, _ := json.Marshal(metrics.Recommendations)

	_, err := cat.db.ExecContext(ctx, query,
		metrics.OverallAccuracy,
		metrics.MCCAccuracy,
		metrics.NAICSAccuracy,
		metrics.SICAccuracy,
		metrics.IndustryAccuracy,
		metrics.ConfidenceAccuracy,
		string(categoryMetricsJSON),
		string(processingMetricsJSON),
		string(errorAnalysisJSON),
		string(recommendationsJSON),
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to save accuracy report: %w", err)
	}

	cat.logger.Println("Accuracy report saved successfully")
	return nil
}
