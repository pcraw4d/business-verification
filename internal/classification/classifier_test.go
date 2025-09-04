package classification

import (
	"context"
	"log"
	"testing"
)

// TestNewClassificationCodeGenerator tests the constructor
func TestNewClassificationCodeGenerator(t *testing.T) {
	mockRepo := &MockKeywordRepository{}

	// Test with custom logger
	logger := log.Default()
	generator := NewClassificationCodeGenerator(mockRepo, logger)

	if generator == nil {
		t.Error("Expected generator to be created")
	}

	if generator.repo != mockRepo {
		t.Error("Expected repository to be set correctly")
	}

	if generator.logger != logger {
		t.Error("Expected logger to be set correctly")
	}

	// Test with nil logger (should use default)
	generator = NewClassificationCodeGenerator(mockRepo, nil)
	if generator.logger == nil {
		t.Error("Expected default logger to be used when nil is passed")
	}
}

// TestGenerateClassificationCodes tests the main classification code generation
func TestGenerateClassificationCodes(t *testing.T) {
	mockRepo := &MockKeywordRepository{}
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	ctx := context.Background()
	keywords := []string{"software", "technology", "platform"}
	detectedIndustry := "Technology"
	confidence := 0.85

	// Test successful generation
	codes, err := generator.GenerateClassificationCodes(ctx, keywords, detectedIndustry, confidence)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if codes == nil {
		t.Error("Expected codes to be generated")
	}

	// Check that codes were generated for each type
	if len(codes.MCC) == 0 {
		t.Error("Expected MCC codes to be generated")
	}

	if len(codes.SIC) == 0 {
		t.Error("Expected SIC codes to be generated")
	}

	if len(codes.NAICS) == 0 {
		t.Error("Expected NAICS codes to be generated")
	}

	// Test with empty keywords
	codes, err = generator.GenerateClassificationCodes(ctx, []string{}, detectedIndustry, confidence)
	if err != nil {
		t.Errorf("Expected no error with empty keywords, got: %v", err)
	}

	// Test with nil keywords
	codes, err = generator.GenerateClassificationCodes(ctx, nil, detectedIndustry, confidence)
	if err != nil {
		t.Errorf("Expected no error with nil keywords, got: %v", err)
	}
}

// TestGenerateMCCCodes tests MCC code generation
func TestGenerateMCCCodes(t *testing.T) {
	mockRepo := &MockKeywordRepository{}
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	ctx := context.Background()
	confidence := 0.8

	// Test financial keywords
	financialKeywords := []string{"bank", "finance", "credit"}
	codes := &ClassificationCodesInfo{}

	err := generator.generateMCCCodes(ctx, codes, financialKeywords, confidence)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(codes.MCC) == 0 {
		t.Error("Expected MCC codes to be generated for financial keywords")
	}

	// Test technology keywords
	techKeywords := []string{"software", "platform", "digital"}
	codes = &ClassificationCodesInfo{}

	err = generator.generateMCCCodes(ctx, codes, techKeywords, confidence)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(codes.MCC) == 0 {
		t.Error("Expected MCC codes to be generated for technology keywords")
	}

	// Test no matching keywords
	noMatchKeywords := []string{"xyz", "abc", "def"}
	codes = &ClassificationCodesInfo{}

	err = generator.generateMCCCodes(ctx, codes, noMatchKeywords, confidence)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(codes.MCC) != 0 {
		t.Error("Expected no MCC codes for non-matching keywords")
	}
}

// TestGenerateSICCodes tests SIC code generation
func TestGenerateSICCodes(t *testing.T) {
	mockRepo := &MockKeywordRepository{}
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	ctx := context.Background()
	confidence := 0.8
	keywords := []string{"software", "technology"}

	// Test Technology industry
	codes := &ClassificationCodesInfo{}
	err := generator.generateSICCodes(ctx, codes, keywords, "Technology", confidence)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(codes.SIC) == 0 {
		t.Error("Expected SIC codes to be generated for Technology industry")
	}

	// Test Financial Services industry
	codes = &ClassificationCodesInfo{}
	err = generator.generateSICCodes(ctx, codes, keywords, "Financial Services", confidence)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(codes.SIC) == 0 {
		t.Error("Expected SIC codes to be generated for Financial Services industry")
	}

	// Test unknown industry
	codes = &ClassificationCodesInfo{}
	err = generator.generateSICCodes(ctx, codes, keywords, "Unknown Industry", confidence)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(codes.SIC) != 0 {
		t.Error("Expected no SIC codes for unknown industry")
	}
}

// TestGenerateNAICSCodes tests NAICS code generation
func TestGenerateNAICSCodes(t *testing.T) {
	mockRepo := &MockKeywordRepository{}
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	ctx := context.Background()
	confidence := 0.8
	keywords := []string{"software", "technology"}

	// Test Technology industry
	codes := &ClassificationCodesInfo{}
	err := generator.generateNAICSCodes(ctx, codes, keywords, "Technology", confidence)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(codes.NAICS) == 0 {
		t.Error("Expected NAICS codes to be generated for Technology industry")
	}

	// Test Manufacturing industry
	codes = &ClassificationCodesInfo{}
	err = generator.generateNAICSCodes(ctx, codes, keywords, "Manufacturing", confidence)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(codes.NAICS) == 0 {
		t.Error("Expected NAICS codes to be generated for Manufacturing industry")
	}

	// Test unknown industry
	codes = &ClassificationCodesInfo{}
	err = generator.generateNAICSCodes(ctx, codes, keywords, "Unknown Industry", confidence)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(codes.NAICS) != 0 {
		t.Error("Expected no NAICS codes for unknown industry")
	}
}

// TestValidateClassificationCodes tests code validation
func TestValidateClassificationCodes(t *testing.T) {
	mockRepo := &MockKeywordRepository{}
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	// Test nil codes
	err := generator.ValidateClassificationCodes(nil, "Technology")
	if err == nil {
		t.Error("Expected error for nil codes")
	}

	// Test valid codes
	validCodes := &ClassificationCodesInfo{
		MCC: []MCCCode{
			{Code: "5734", Description: "Computer Software Stores", Confidence: 0.9, Keywords: []string{"software"}},
		},
		SIC: []SICCode{
			{Code: "7372", Description: "Prepackaged Software", Confidence: 0.85, Keywords: []string{"software"}},
		},
		NAICS: []NAICSCode{
			{Code: "541511", Description: "Custom Computer Programming Services", Confidence: 0.9, Keywords: []string{"software"}},
		},
	}

	err = generator.ValidateClassificationCodes(validCodes, "Technology")
	if err != nil {
		t.Errorf("Expected no error for valid codes, got: %v", err)
	}

	// Test invalid confidence score
	invalidCodes := &ClassificationCodesInfo{
		MCC: []MCCCode{
			{Code: "5734", Description: "Computer Software Stores", Confidence: 1.5, Keywords: []string{"software"}},
		},
	}

	err = generator.ValidateClassificationCodes(invalidCodes, "Technology")
	if err == nil {
		t.Error("Expected error for invalid confidence score")
	}
}

// TestGetCodeStatistics tests code statistics generation
func TestGetCodeStatistics(t *testing.T) {
	mockRepo := &MockKeywordRepository{}
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	// Test with nil codes
	stats := generator.GetCodeStatistics(nil)

	if stats["total_codes"] != 0 {
		t.Error("Expected total_codes to be 0 for nil codes")
	}

	if stats["avg_confidence"] != 0.0 {
		t.Error("Expected avg_confidence to be 0.0 for nil codes")
	}

	// Test with valid codes
	validCodes := &ClassificationCodesInfo{
		MCC: []MCCCode{
			{Code: "5734", Description: "Computer Software Stores", Confidence: 0.9, Keywords: []string{"software"}},
			{Code: "7372", Description: "Prepackaged Software", Confidence: 0.85, Keywords: []string{"software"}},
		},
		SIC: []SICCode{
			{Code: "7372", Description: "Prepackaged Software", Confidence: 0.8, Keywords: []string{"software"}},
		},
		NAICS: []NAICSCode{
			{Code: "541511", Description: "Custom Computer Programming Services", Confidence: 0.95, Keywords: []string{"software"}},
		},
	}

	stats = generator.GetCodeStatistics(validCodes)

	expectedTotal := 4
	if stats["total_codes"] != expectedTotal {
		t.Errorf("Expected total_codes to be %d, got %v", expectedTotal, stats["total_codes"])
	}

	expectedMCC := 2
	if stats["mcc_count"] != expectedMCC {
		t.Errorf("Expected mcc_count to be %d, got %v", expectedMCC, stats["mcc_count"])
	}

	expectedSIC := 1
	if stats["sic_count"] != expectedSIC {
		t.Errorf("Expected sic_count to be %d, got %v", expectedSIC, stats["sic_count"])
	}

	expectedNAICS := 1
	if stats["naics_count"] != expectedNAICS {
		t.Errorf("Expected naics_count to be %d, got %v", expectedNAICS, stats["naics_count"])
	}

	// Check average confidence calculation
	expectedAvg := (0.9 + 0.85 + 0.8 + 0.95) / 4.0
	if stats["avg_confidence"] != expectedAvg {
		t.Errorf("Expected avg_confidence to be %.2f, got %v", expectedAvg, stats["avg_confidence"])
	}
}

// TestHelperMethods tests the helper methods
func TestHelperMethods(t *testing.T) {
	mockRepo := &MockKeywordRepository{}
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	// Test containsAny
	source := []string{"apple", "banana", "cherry"}
	targets := []string{"banana", "grape"}

	if !generator.containsAny(source, targets) {
		t.Error("Expected containsAny to return true for matching targets")
	}

	noMatchTargets := []string{"grape", "orange"}
	if generator.containsAny(source, noMatchTargets) {
		t.Error("Expected containsAny to return false for non-matching targets")
	}

	// Test findMatchingKeywords
	keywords := []string{"software", "platform", "digital"}
	targets = []string{"software", "web"}

	matches := generator.findMatchingKeywords(keywords, targets)
	expectedMatches := 1
	if len(matches) != expectedMatches {
		t.Errorf("Expected %d matches, got %d", expectedMatches, len(matches))
	}

	// Test with nil keywords
	matchesNil := generator.findMatchingKeywords(nil, targets)
	if len(matchesNil) != 0 {
		t.Error("Expected no matches for nil keywords")
	}
}

// TestIntegration tests the integration between different methods
func TestIntegration(t *testing.T) {
	mockRepo := &MockKeywordRepository{}
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	ctx := context.Background()

	// Test complete workflow
	keywords := []string{"bank", "finance", "credit", "investment"}
	detectedIndustry := "Financial Services"
	confidence := 0.9

	// Generate codes
	codes, err := generator.GenerateClassificationCodes(ctx, keywords, detectedIndustry, confidence)
	if err != nil {
		t.Errorf("Expected no error in code generation, got: %v", err)
	}

	// Validate codes
	err = generator.ValidateClassificationCodes(codes, detectedIndustry)
	if err != nil {
		t.Errorf("Expected no error in validation, got: %v", err)
	}

	// Get statistics
	stats := generator.GetCodeStatistics(codes)

	// Verify we have codes for all types
	if stats["total_codes"].(int) == 0 {
		t.Error("Expected codes to be generated in integration test")
	}

	// Verify confidence scores are reasonable
	if stats["avg_confidence"].(float64) < 0.5 {
		t.Error("Expected reasonable confidence scores in integration test")
	}
}
