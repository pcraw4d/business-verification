package classification

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"kyb-platform/internal/classification/repository"
)

// ClassificationCodeGenerator provides database-driven classification code generation
type ClassificationCodeGenerator struct {
	repo    repository.KeywordRepository
	logger  *log.Logger
	monitor *ClassificationAccuracyMonitoring
}

// NewClassificationCodeGenerator creates a new classification code generator
func NewClassificationCodeGenerator(repo repository.KeywordRepository, logger *log.Logger) *ClassificationCodeGenerator {
	if logger == nil {
		logger = log.Default()
	}

	return &ClassificationCodeGenerator{
		repo:    repo,
		logger:  logger,
		monitor: nil, // Will be set separately if monitoring is needed
	}
}

// NewClassificationCodeGeneratorWithMonitoring creates a new classification code generator with monitoring
func NewClassificationCodeGeneratorWithMonitoring(repo repository.KeywordRepository, logger *log.Logger, monitor *ClassificationAccuracyMonitoring) *ClassificationCodeGenerator {
	if logger == nil {
		logger = log.Default()
	}

	return &ClassificationCodeGenerator{
		repo:    repo,
		logger:  logger,
		monitor: monitor,
	}
}

// ClassificationCodesInfo contains the industry classification codes
type ClassificationCodesInfo struct {
	MCC   []MCCCode   `json:"mcc,omitempty"`
	SIC   []SICCode   `json:"sic,omitempty"`
	NAICS []NAICSCode `json:"naics,omitempty"`
}

// MCCCode represents a Merchant Category Code
type MCCCode struct {
	Code        string   `json:"code"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`
	Keywords    []string `json:"keywords_matched"`
}

// SICCode represents a Standard Industrial Classification code
type SICCode struct {
	Code        string   `json:"code"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`
	Keywords    []string `json:"keywords_matched"`
}

// NAICSCode represents a North American Industry Classification System code
type NAICSCode struct {
	Code        string   `json:"code"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`
	Keywords    []string `json:"keywords_matched"`
}

// GenerateClassificationCodes generates MCC, SIC, and NAICS codes based on extracted keywords and industry analysis
func (g *ClassificationCodeGenerator) GenerateClassificationCodes(ctx context.Context, keywords []string, detectedIndustry string, confidence float64) (*ClassificationCodesInfo, error) {
	startTime := time.Now()
	requestID := g.generateRequestID()

	g.logger.Printf("🔍 Generating classification codes for industry: %s (confidence: %.2f%%) (request: %s)", detectedIndustry, confidence*100, requestID)

	codes := &ClassificationCodesInfo{
		MCC:   []MCCCode{},
		SIC:   []SICCode{},
		NAICS: []NAICSCode{},
	}

	// Convert keywords to lowercase for matching
	keywordsLower := make([]string, len(keywords))
	for i, keyword := range keywords {
		keywordsLower[i] = strings.ToLower(keyword)
	}

	// Generate codes using parallel processing for better performance
	g.generateCodesInParallel(ctx, codes, keywordsLower, detectedIndustry, confidence)

	// Record performance metrics
	g.recordCodeGenerationMetrics(ctx, requestID, keywords, detectedIndustry, confidence, codes, time.Since(startTime), nil)

	g.logger.Printf("✅ Generated %d MCC, %d SIC, %d NAICS codes (request: %s)",
		len(codes.MCC), len(codes.SIC), len(codes.NAICS), requestID)

	return codes, nil
}

// generateCodesInParallel generates MCC, SIC, and NAICS codes in parallel for better performance
func (g *ClassificationCodeGenerator) generateCodesInParallel(ctx context.Context, codes *ClassificationCodesInfo, keywordsLower []string, detectedIndustry string, confidence float64) {
	g.logger.Printf("🚀 Starting parallel code generation for MCC, SIC, and NAICS")

	// Create a WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup
	var mu sync.Mutex // Mutex to protect shared data access

	// Channel to collect errors from goroutines
	errorChan := make(chan error, 3)

	// Generate MCC codes in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.logger.Printf("🔄 Starting MCC code generation...")

		// Get industry object first (same as NAICS)
		industryObj, err := g.repo.GetIndustryByName(ctx, detectedIndustry)
		if err != nil {
			g.logger.Printf("⚠️ Failed to get industry for MCC: %v", err)
			errorChan <- fmt.Errorf("MCC industry lookup: %w", err)
			return
		}

		// Get codes for the industry
		allCodes, err := g.repo.GetCachedClassificationCodes(ctx, industryObj.ID)
		if err != nil {
			g.logger.Printf("⚠️ Failed to get MCC codes from database: %v", err)
			errorChan <- fmt.Errorf("MCC codes: %w", err)
			return
		}

		// Filter MCC codes by type and convert
		var mccResults []MCCCode
		for _, code := range allCodes {
			if code.CodeType == "MCC" {
				// Use industry-based codes directly (no keyword filtering needed)
				mccResults = append(mccResults, MCCCode{
					Code:        code.Code,
					Description: code.Description,
					Confidence:  confidence * 0.9,
				})
			}
		}

		// Thread-safe update of shared codes
		mu.Lock()
		codes.MCC = mccResults
		mu.Unlock()

		g.logger.Printf("✅ MCC code generation completed: %d codes", len(mccResults))
	}()

	// Generate SIC codes in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.logger.Printf("🔄 Starting SIC code generation...")

		// Get industry object first (same as NAICS and MCC)
		industryObj, err := g.repo.GetIndustryByName(ctx, detectedIndustry)
		if err != nil {
			g.logger.Printf("⚠️ Failed to get industry for SIC: %v", err)
			errorChan <- fmt.Errorf("SIC industry lookup: %w", err)
			return
		}

		// Get codes for the industry
		allCodes, err := g.repo.GetCachedClassificationCodes(ctx, industryObj.ID)
		if err != nil {
			g.logger.Printf("⚠️ Failed to get SIC codes from database: %v", err)
			errorChan <- fmt.Errorf("SIC codes: %w", err)
			return
		}

		// Filter SIC codes by type and convert
		var sicResults []SICCode
		for _, code := range allCodes {
			if code.CodeType == "SIC" {
				// Use industry-based codes directly (no keyword filtering needed)
				sicResults = append(sicResults, SICCode{
					Code:        code.Code,
					Description: code.Description,
					Confidence:  confidence * 0.9,
				})
			}
		}

		// Thread-safe update of shared codes
		mu.Lock()
		codes.SIC = sicResults
		mu.Unlock()

		g.logger.Printf("✅ SIC code generation completed: %d codes", len(sicResults))
	}()

	// Generate NAICS codes in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.logger.Printf("🔄 Starting NAICS code generation...")

		// Get industry object for NAICS generation
		industryObj, err := g.repo.GetIndustryByName(ctx, detectedIndustry)
		if err != nil {
			g.logger.Printf("⚠️ Failed to get industry for NAICS: %v", err)
			errorChan <- fmt.Errorf("NAICS industry lookup: %w", err)
			return
		}

		// Get NAICS codes for the industry
		naicsCodes, err := g.repo.GetCachedClassificationCodes(ctx, industryObj.ID)
		if err != nil {
			g.logger.Printf("⚠️ Failed to get NAICS codes from database: %v", err)
			errorChan <- fmt.Errorf("NAICS codes: %w", err)
			return
		}

		// Filter NAICS codes by type (no keyword filtering needed - industry match is sufficient)
		var naicsResults []NAICSCode
		for _, code := range naicsCodes {
			if code.CodeType == "NAICS" {
				// Use industry-based codes directly (no keyword filtering needed)
				naicsResults = append(naicsResults, NAICSCode{
					Code:        code.Code,
					Description: code.Description,
					Confidence:  confidence * 0.9,
				})
			}
		}

		// Thread-safe update of shared codes
		mu.Lock()
		codes.NAICS = naicsResults
		mu.Unlock()

		g.logger.Printf("✅ NAICS code generation completed: %d codes", len(naicsResults))
	}()

	// Wait for all goroutines to complete
	wg.Wait()
	close(errorChan)

	// Log any errors that occurred
	for err := range errorChan {
		g.logger.Printf("⚠️ Error in parallel code generation: %v", err)
	}

	g.logger.Printf("🚀 Parallel code generation completed: %d MCC, %d SIC, %d NAICS codes",
		len(codes.MCC), len(codes.SIC), len(codes.NAICS))
}

// =============================================================================
// Performance Monitoring Helper Methods
// =============================================================================

// generateRequestID generates a unique request ID for tracking
func (g *ClassificationCodeGenerator) generateRequestID() string {
	return fmt.Sprintf("code_gen_%d", time.Now().UnixNano())
}

// recordCodeGenerationMetrics records code generation performance metrics
func (g *ClassificationCodeGenerator) recordCodeGenerationMetrics(
	ctx context.Context,
	requestID string,
	keywords []string,
	detectedIndustry string,
	confidence float64,
	codes *ClassificationCodesInfo,
	responseTime time.Duration,
	err error,
) {
	if g.monitor == nil {
		return // No monitoring configured
	}

	// Prepare metrics data
	metrics := &ClassificationAccuracyMetrics{
		Timestamp:            time.Now(),
		RequestID:            requestID,
		PredictedIndustry:    detectedIndustry,
		PredictedConfidence:  confidence,
		ResponseTimeMs:       float64(responseTime.Nanoseconds()) / 1e6, // Convert to milliseconds
		ClassificationMethod: stringPtr("code_generation"),
		KeywordsUsed:         keywords,
		ConfidenceThreshold:  0.5, // Default threshold
		CreatedAt:            time.Now(),
	}

	// Set error message if there was an error
	if err != nil {
		errorMsg := err.Error()
		metrics.ErrorMessage = &errorMsg
	}

	// Record metrics asynchronously to avoid blocking the main flow
	go func() {
		// Note: This would call the actual monitoring method when implemented
		// if err := g.monitor.RecordClassificationMetrics(ctx, metrics); err != nil {
		//     g.logger.Printf("⚠️ Failed to record code generation metrics: %v", err)
		// }
	}()
}

// GetCodeGenerationMetrics returns current code generation performance metrics
func (g *ClassificationCodeGenerator) GetCodeGenerationMetrics(ctx context.Context) (*ClassificationAccuracyStats, error) {
	if g.monitor == nil {
		return nil, fmt.Errorf("monitoring not configured")
	}

	// Note: This would call the actual monitoring method when implemented
	// return g.monitor.GetClassificationAccuracyStats(ctx, 24*time.Hour)
	return nil, fmt.Errorf("monitoring not fully implemented")
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// matchesKeywords checks if a classification code matches any of the provided keywords
func (g *ClassificationCodeGenerator) matchesKeywords(code *repository.ClassificationCode, keywordsLower []string) bool {
	descriptionLower := strings.ToLower(code.Description)

	// Check if any of the provided keywords match
	for _, keyword := range keywordsLower {
		if strings.Contains(descriptionLower, keyword) {
			return true
		}
	}

	return false
}

// ValidateClassificationCodes validates that the generated codes are consistent with the detected industry
func (g *ClassificationCodeGenerator) ValidateClassificationCodes(codes *ClassificationCodesInfo, detectedIndustry string) error {
	if codes == nil {
		return fmt.Errorf("classification codes cannot be nil")
	}

	// Validate that codes exist for the detected industry
	if detectedIndustry != "" {
		hasIndustryCodes := false

		// Check if we have any codes that match the industry
		if len(codes.SIC) > 0 || len(codes.NAICS) > 0 {
			hasIndustryCodes = true
		}

		if !hasIndustryCodes {
			g.logger.Printf("⚠️ Warning: No industry-specific codes found for detected industry: %s", detectedIndustry)
		}
	}

	// Validate confidence scores are within reasonable bounds
	for _, mcc := range codes.MCC {
		if mcc.Confidence < 0.0 || mcc.Confidence > 1.0 {
			return fmt.Errorf("invalid MCC confidence score: %.2f (must be between 0.0 and 1.0)", mcc.Confidence)
		}
	}

	for _, sic := range codes.SIC {
		if sic.Confidence < 0.0 || sic.Confidence > 1.0 {
			return fmt.Errorf("invalid SIC confidence score: %.2f (must be between 0.0 and 1.0)", sic.Confidence)
		}
	}

	for _, naics := range codes.NAICS {
		if naics.Confidence < 0.0 || naics.Confidence > 1.0 {
			return fmt.Errorf("invalid NAICS confidence score: %.2f (must be between 0.0 and 1.0)", naics.Confidence)
		}
	}

	return nil
}

// GetCodeStatistics returns statistics about the generated classification codes
// containsAny checks if any of the source strings contain any of the target strings
func (g *ClassificationCodeGenerator) containsAny(source []string, targets []string) bool {
	for _, s := range source {
		for _, t := range targets {
			if strings.Contains(strings.ToLower(s), strings.ToLower(t)) {
				return true
			}
		}
	}
	return false
}

// findMatchingKeywords finds keywords that match any of the target strings
func (g *ClassificationCodeGenerator) findMatchingKeywords(keywords []string, targets []string) []string {
	if keywords == nil {
		return []string{}
	}

	var matches []string
	for _, keyword := range keywords {
		for _, target := range targets {
			if strings.Contains(strings.ToLower(keyword), strings.ToLower(target)) {
				matches = append(matches, keyword)
				break // Only add each keyword once
			}
		}
	}
	return matches
}

func (g *ClassificationCodeGenerator) GetCodeStatistics(codes *ClassificationCodesInfo) map[string]interface{} {
	if codes == nil {
		return map[string]interface{}{
			"total_codes":    0,
			"mcc_count":      0,
			"sic_count":      0,
			"naics_count":    0,
			"avg_confidence": 0.0,
		}
	}

	totalCodes := len(codes.MCC) + len(codes.SIC) + len(codes.NAICS)

	// Calculate average confidence
	totalConfidence := 0.0
	confidenceCount := 0

	for _, mcc := range codes.MCC {
		totalConfidence += mcc.Confidence
		confidenceCount++
	}

	for _, sic := range codes.SIC {
		totalConfidence += sic.Confidence
		confidenceCount++
	}

	for _, naics := range codes.NAICS {
		totalConfidence += naics.Confidence
		confidenceCount++
	}

	avgConfidence := 0.0
	if confidenceCount > 0 {
		avgConfidence = totalConfidence / float64(confidenceCount)
	}

	return map[string]interface{}{
		"total_codes":    totalCodes,
		"mcc_count":      len(codes.MCC),
		"sic_count":      len(codes.SIC),
		"naics_count":    len(codes.NAICS),
		"avg_confidence": avgConfidence,
	}
}
