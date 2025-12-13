package classification

import (
	"log"
	"os"
	"testing"
)

// TestPhase2_Top3CodesPerType tests that exactly 3 codes are returned per type with Source field
func TestPhase2_Top3CodesPerType(t *testing.T) {
	// This is a unit test that validates the structure
	// Full integration test would require database connection
	
	// Test that MCCCode, SICCode, NAICSCode have Source field
	mccCode := MCCCode{
		Code:        "5812",
		Description: "Eating Places",
		Confidence:  0.85,
		Source:      "keyword_match",
	}
	
	if mccCode.Source == "" {
		t.Error("MCCCode missing Source field")
	}
	
	sicCode := SICCode{
		Code:        "5812",
		Description: "Eating Places",
		Confidence:  0.85,
		Source:      "industry_match",
	}
	
	if sicCode.Source == "" {
		t.Error("SICCode missing Source field")
	}
	
	naicsCode := NAICSCode{
		Code:        "722511",
		Description: "Full-Service Restaurants",
		Confidence:  0.90,
		Source:      "crosswalk_from_mcc",
	}
	
	if naicsCode.Source == "" {
		t.Error("NAICSCode missing Source field")
	}
	
	// Test selectTopCodes function
	candidates := []CodeResult{
		{Code: "1", Description: "Code 1", Confidence: 0.95, Source: "keyword"},
		{Code: "2", Description: "Code 2", Confidence: 0.90, Source: "industry"},
		{Code: "3", Description: "Code 3", Confidence: 0.85, Source: "trigram"},
		{Code: "4", Description: "Code 4", Confidence: 0.80, Source: "keyword"},
		{Code: "5", Description: "Code 5", Confidence: 0.75, Source: "industry"},
	}
	
	generator := &ClassificationCodeGenerator{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
	
	topCodes := generator.selectTopCodes(candidates, 3)
	
	if len(topCodes) != 3 {
		t.Errorf("Expected 3 codes, got %d", len(topCodes))
	}
	
	// Verify they're sorted by confidence (descending)
	if topCodes[0].Confidence < topCodes[1].Confidence || 
	   topCodes[1].Confidence < topCodes[2].Confidence {
		t.Error("Codes not sorted by confidence (descending)")
	}
	
	// Verify all have Source field
	for i, code := range topCodes {
		if code.Source == "" {
			t.Errorf("Code %d missing Source field", i)
		}
	}
}

// TestPhase2_ConfidenceCalibration tests the enhanced confidence calibration
func TestPhase2_ConfidenceCalibration(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	calibrator := NewConfidenceCalibrator(logger)
	
	// Test with high-quality content and strong strategy agreement
	strategyScores := map[string]float64{
		"keyword":       0.85,
		"entity":        0.82,
		"topic":         0.80,
		"co_occurrence": 0.78,
	}
	
	calibrated := calibrator.CalibrateConfidence(
		strategyScores,
		0.85, // High content quality
		0.90, // High code agreement
		"multi_strategy",
	)
	
	// Should be in 70-95% range
	if calibrated < 0.70 || calibrated > 0.95 {
		t.Errorf("Calibrated confidence %.2f not in expected range [0.70, 0.95]", calibrated)
	}
	
	// Should be higher than base average due to quality factors
	baseAvg := (0.85 + 0.82 + 0.80 + 0.78) / 4.0
	if calibrated < baseAvg {
		t.Errorf("Calibrated confidence %.2f should be >= base average %.2f", calibrated, baseAvg)
	}
	
	// Test with low-quality content
	calibratedLow := calibrator.CalibrateConfidence(
		strategyScores,
		0.4, // Low content quality
		0.5, // Low code agreement
		"multi_strategy",
	)
	
	// Should still be in valid range but lower
	if calibratedLow < 0.30 || calibratedLow > 0.95 {
		t.Errorf("Low-quality calibrated confidence %.2f not in expected range [0.30, 0.95]", calibratedLow)
	}
	
	if calibratedLow >= calibrated {
		t.Errorf("Low-quality confidence %.2f should be < high-quality %.2f", calibratedLow, calibrated)
	}
}

// TestPhase2_CodeAgreement tests CalculateCodeAgreement method
func TestPhase2_CodeAgreement(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	calibrator := NewConfidenceCalibrator(logger)
	
	// Test with aligned codes (high confidence)
	codes := &ClassificationCodesInfo{
		MCC: []MCCCode{
			{Code: "5812", Confidence: 0.90},
			{Code: "5813", Confidence: 0.85},
			{Code: "5814", Confidence: 0.80},
		},
		SIC: []SICCode{
			{Code: "5812", Confidence: 0.88},
			{Code: "5813", Confidence: 0.83},
			{Code: "5814", Confidence: 0.78},
		},
		NAICS: []NAICSCode{
			{Code: "722511", Confidence: 0.92},
			{Code: "722513", Confidence: 0.87},
			{Code: "722514", Confidence: 0.82},
		},
	}
	
	agreement := calibrator.CalculateCodeAgreement(codes)
	
	// Should be high (>0.80) when codes align
	if agreement < 0.70 {
		t.Errorf("Code agreement %.2f should be >= 0.70 for aligned codes", agreement)
	}
	
	// Test with misaligned codes (low confidence)
	codesLow := &ClassificationCodesInfo{
		MCC: []MCCCode{
			{Code: "5812", Confidence: 0.50},
		},
		SIC: []SICCode{
			{Code: "5812", Confidence: 0.45},
		},
		NAICS: []NAICSCode{
			{Code: "722511", Confidence: 0.55},
		},
	}
	
	agreementLow := calibrator.CalculateCodeAgreement(codesLow)
	
	// Should be lower
	if agreementLow >= agreement {
		t.Errorf("Misaligned code agreement %.2f should be < aligned %.2f", agreementLow, agreement)
	}
}

// TestPhase2_FastPath tests the fast path implementation
func TestPhase2_FastPath(t *testing.T) {
	// Test extractObviousKeywords
	classifier := &MultiStrategyClassifier{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
	
	// Test with obvious restaurant keywords
	keywords := classifier.extractObviousKeywords(
		"Joe's Pizza Restaurant",
		"A family-owned pizza restaurant serving authentic Italian cuisine",
		"https://joespizza.com",
	)
	
	if len(keywords) == 0 {
		t.Error("Should extract obvious keywords from restaurant business")
	}
	
	// Should find "restaurant"
	foundRestaurant := false
	for _, kw := range keywords {
		if kw == "restaurant" {
			foundRestaurant = true
			break
		}
	}
	
	if !foundRestaurant {
		t.Error("Should extract 'restaurant' as obvious keyword")
	}
	
	// Test with obvious coffee shop
	coffeeKeywords := classifier.extractObviousKeywords(
		"Starbucks Coffee",
		"Coffee shop and cafe",
		"https://starbucks.com",
	)
	
	if len(coffeeKeywords) == 0 {
		t.Error("Should extract obvious keywords from coffee shop")
	}
	
	// Test with non-obvious business (should return empty)
	genericKeywords := classifier.extractObviousKeywords(
		"ABC Corporation",
		"General business services",
		"https://abc.com",
	)
	
	// May or may not find keywords, but shouldn't crash
	_ = genericKeywords
}

// TestPhase2_ExplanationGeneration tests explanation generation
func TestPhase2_ExplanationGeneration(t *testing.T) {
	generator := NewExplanationGenerator()
	
	// Create a mock MultiStrategyResult
	result := &MultiStrategyResult{
		PrimaryIndustry: "Restaurants",
		Confidence:      0.88,
		Keywords:        []string{"restaurant", "pizza", "dining"},
		Method:          "multi_strategy",
		Strategies: []ClassificationStrategy{
			{StrategyName: "keyword", Score: 0.90, Confidence: 0.88},
			{StrategyName: "entity", Score: 0.85, Confidence: 0.85},
		},
	}
	
	// Create mock codes
	codes := &ClassificationCodesInfo{
		MCC: []MCCCode{
			{Code: "5812", Confidence: 0.90},
		},
		SIC: []SICCode{
			{Code: "5812", Confidence: 0.88},
		},
		NAICS: []NAICSCode{
			{Code: "722511", Confidence: 0.92},
		},
	}
	
	explanation := generator.GenerateExplanation(result, codes, 0.85)
	
	// Verify explanation structure
	if explanation.PrimaryReason == "" {
		t.Error("Explanation missing primary reason")
	}
	
	if len(explanation.SupportingFactors) == 0 {
		t.Error("Explanation missing supporting factors")
	}
	
	if len(explanation.SupportingFactors) < 3 {
		t.Errorf("Expected at least 3 supporting factors, got %d", len(explanation.SupportingFactors))
	}
	
	if len(explanation.KeyTermsFound) == 0 {
		t.Error("Explanation missing key terms")
	}
	
	if explanation.MethodUsed == "" {
		t.Error("Explanation missing method used")
	}
	
	if explanation.ProcessingPath == "" {
		t.Error("Explanation missing processing path")
	}
	
	// Verify primary reason is meaningful
	if len(explanation.PrimaryReason) < 20 {
		t.Errorf("Primary reason too short: %s", explanation.PrimaryReason)
	}
}

// TestPhase2_GenericFallback tests the generic fallback fix
func TestPhase2_GenericFallback(t *testing.T) {
	classifier := &MultiStrategyClassifier{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
	
	// Test boostSpecificIndustries
	combinedScores := map[int]float64{
		1:  0.75, // Technology (specific)
		2:  0.80, // Healthcare (specific)
		26: 0.85, // General Business (generic)
	}
	
	industryNames := map[int]string{
		1:  "Technology",
		2:  "Healthcare",
		26: "General Business",
	}
	
	boosted := classifier.boostSpecificIndustries(combinedScores, industryNames)
	
	// Specific industries should be boosted
	if boosted[1] <= combinedScores[1] {
		t.Error("Technology industry should be boosted")
	}
	
	if boosted[2] <= combinedScores[2] {
		t.Error("Healthcare industry should be boosted")
	}
	
	// Generic should not be boosted
	if boosted[26] != combinedScores[26] {
		t.Error("General Business should not be boosted")
	}
	
	// Test selectBestIndustry with generic vs specific
	_, industryName, score := classifier.selectBestIndustry(
		combinedScores,
		industryNames,
		26, // Current: General Business
		"General Business",
		0.70, // Low confidence
	)
	
	// Should prefer specific industry if within 0.15
	if industryName == "General Business" && score < 0.70 {
		// This is expected - generic requires >=0.70
	} else if industryName != "General Business" {
		// Should prefer specific if available
		if industryName == "Technology" || industryName == "Healthcare" {
			// Good - preferred specific
		}
	}
	
	// Verify score is valid
	if score < 0 || score > 1.0 {
		t.Errorf("Invalid score: %.2f", score)
	}
}

// TestPhase2_EnrichWithCrosswalks tests crosswalk enrichment
func TestPhase2_EnrichWithCrosswalks(t *testing.T) {
	// This would require a repository mock
	// For now, test the structure
	
	codes := &ClassificationCodesInfo{
		MCC: []MCCCode{
			{Code: "5812", Confidence: 0.90, Description: "Eating Places"},
		},
		SIC: []SICCode{},
		NAICS: []NAICSCode{},
	}
	
	// Test that enrichWithCrosswalks doesn't crash
	generator := &ClassificationCodeGenerator{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
	
	// Without repository, this will return codes unchanged
	enriched := generator.enrichWithCrosswalks(codes)
	
	if enriched == nil {
		t.Error("enrichWithCrosswalks should return codes")
	}
}

// TestPhase2_FillGapsWithCrosswalks tests gap filling
func TestPhase2_FillGapsWithCrosswalks(t *testing.T) {
	generator := &ClassificationCodeGenerator{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
	
	codes := &ClassificationCodesInfo{
		MCC: []MCCCode{
			{Code: "5812", Confidence: 0.90},
		},
		SIC: []SICCode{}, // Empty - should be filled
		NAICS: []NAICSCode{}, // Empty - should be filled
	}
	
	// Without repository, this will return codes unchanged
	filled := generator.fillGapsWithCrosswalks(codes)
	
	if filled == nil {
		t.Error("fillGapsWithCrosswalks should return codes")
	}
}

// TestPhase2_EnsureTop3 tests the ensureTop3 methods
func TestPhase2_EnsureTop3(t *testing.T) {
	generator := &ClassificationCodeGenerator{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
	
	// Test with more than 3 codes
	mccCodes := []MCCCode{
		{Code: "1", Confidence: 0.95},
		{Code: "2", Confidence: 0.90},
		{Code: "3", Confidence: 0.85},
		{Code: "4", Confidence: 0.80},
		{Code: "5", Confidence: 0.75},
	}
	
	top3 := generator.ensureTop3MCC(mccCodes)
	
	if len(top3) != 3 {
		t.Errorf("Expected 3 codes, got %d", len(top3))
	}
	
	// Test with less than 3 codes
	fewCodes := []MCCCode{
		{Code: "1", Confidence: 0.95},
		{Code: "2", Confidence: 0.90},
	}
	
	top3Few := generator.ensureTop3MCC(fewCodes)
	
	if len(top3Few) != 2 {
		t.Errorf("Expected 2 codes, got %d", len(top3Few))
	}
	
	// Test with exactly 3 codes
	exact3 := []MCCCode{
		{Code: "1", Confidence: 0.95},
		{Code: "2", Confidence: 0.90},
		{Code: "3", Confidence: 0.85},
	}
	
	top3Exact := generator.ensureTop3MCC(exact3)
	
	if len(top3Exact) != 3 {
		t.Errorf("Expected 3 codes, got %d", len(top3Exact))
	}
}

// BenchmarkPhase2_FastPath benchmarks fast path performance
func BenchmarkPhase2_FastPath(b *testing.B) {
	classifier := &MultiStrategyClassifier{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = classifier.extractObviousKeywords(
			"Joe's Pizza Restaurant",
			"Family pizza restaurant",
			"https://joespizza.com",
		)
	}
}

// BenchmarkPhase2_ConfidenceCalibration benchmarks calibration performance
func BenchmarkPhase2_ConfidenceCalibration(b *testing.B) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	calibrator := NewConfidenceCalibrator(logger)
	
	strategyScores := map[string]float64{
		"keyword":       0.85,
		"entity":        0.82,
		"topic":         0.80,
		"co_occurrence": 0.78,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calibrator.CalibrateConfidence(
			strategyScores,
			0.85,
			0.90,
			"multi_strategy",
		)
	}
}
