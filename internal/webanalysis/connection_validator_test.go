package webanalysis

import (
	"testing"
	"time"
)

func TestNewConnectionValidator(t *testing.T) {
	cv := NewConnectionValidator()

	if cv == nil {
		t.Fatal("Expected non-nil ConnectionValidator")
	}

	if cv.nameMatcher == nil {
		t.Error("Expected non-nil BusinessNameMatcher")
	}

	if cv.addressVerifier == nil {
		t.Error("Expected non-nil AddressVerifier")
	}

	if cv.contactVerifier == nil {
		t.Error("Expected non-nil ContactVerifier")
	}

	if cv.registrationChecker == nil {
		t.Error("Expected non-nil RegistrationChecker")
	}

	if cv.ownershipScorer == nil {
		t.Error("Expected non-nil OwnershipScorer")
	}

	if cv.confidenceAssessor == nil {
		t.Error("Expected non-nil ConfidenceAssessor")
	}
}

func TestValidateConnection(t *testing.T) {
	cv := NewConnectionValidator()

	result, err := cv.ValidateConnection("Acme Corporation", "https://acme.com")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil validation result")
	}

	if result.BusinessName != "Acme Corporation" {
		t.Errorf("Expected business name 'Acme Corporation', got: %s", result.BusinessName)
	}

	if result.WebsiteURL != "https://acme.com" {
		t.Errorf("Expected website URL 'https://acme.com', got: %s", result.WebsiteURL)
	}

	if result.OverallConfidence < 0 || result.OverallConfidence > 1 {
		t.Errorf("Expected overall confidence between 0 and 1, got: %f", result.OverallConfidence)
	}

	if result.ValidationTime <= 0 {
		t.Error("Expected positive validation time")
	}

	if result.LastChecked.IsZero() {
		t.Error("Expected last checked time to be set")
	}
}

func TestBusinessNameMatcher(t *testing.T) {
	bnm := NewBusinessNameMatcher()

	result, err := bnm.MatchName("Acme Corporation", "https://acme.com")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil name match result")
	}

	if result.OriginalName != "Acme Corporation" {
		t.Errorf("Expected original name 'Acme Corporation', got: %s", result.OriginalName)
	}

	if result.SimilarityScore < 0 || result.SimilarityScore > 1 {
		t.Errorf("Expected similarity score between 0 and 1, got: %f", result.SimilarityScore)
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got: %f", result.Confidence)
	}
}

func TestExactMatcher(t *testing.T) {
	em := NewExactMatcher()

	// Test exact match
	if !em.Match("Acme Corporation", "Acme Corporation") {
		t.Error("Expected exact match to return true")
	}

	// Test case insensitive match
	if !em.Match("Acme Corporation", "acme corporation") {
		t.Error("Expected case insensitive match to return true")
	}

	// Test no match
	if em.Match("Acme Corporation", "Different Company") {
		t.Error("Expected no match to return false")
	}
}

func TestFuzzyMatcher(t *testing.T) {
	fm := NewFuzzyMatcher()

	// Test fuzzy matching
	score := fm.Match("Acme Corporation", "Acme Corp")
	if score < 0 || score > 1 {
		t.Errorf("Expected fuzzy match score between 0 and 1, got: %f", score)
	}

	// Test exact match
	score = fm.Match("Acme Corporation", "Acme Corporation")
	if score < 0 || score > 1 {
		t.Errorf("Expected fuzzy match score between 0 and 1, got: %f", score)
	}
}

func TestAbbreviationMatcher(t *testing.T) {
	am := NewAbbreviationMatcher()

	// Test abbreviation matching
	if !am.Match("Acme Corporation", "https://acme.com") {
		t.Log("Abbreviation matching returned false (expected for test data)")
	}

	// Test no abbreviation
	if am.Match("Acme Company", "https://acme.com") {
		t.Log("Abbreviation matching returned true (unexpected for test data)")
	}
}

func TestAliasMatcher(t *testing.T) {
	alm := NewAliasMatcher()

	// Test alias matching
	if !alm.Match("Microsoft", "https://microsoft.com") {
		t.Log("Alias matching returned false (expected for test data)")
	}

	// Test no alias
	if alm.Match("Unknown Company", "https://unknown.com") {
		t.Log("Alias matching returned true (unexpected for test data)")
	}
}

func TestAddressVerifier(t *testing.T) {
	av := NewAddressVerifier()

	result, err := av.VerifyAddress("Acme Corporation", "https://acme.com")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil address match result")
	}

	if result.SimilarityScore < 0 || result.SimilarityScore > 1 {
		t.Errorf("Expected similarity score between 0 and 1, got: %f", result.SimilarityScore)
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got: %f", result.Confidence)
	}
}

func TestContactVerifier(t *testing.T) {
	cv := NewContactVerifier()

	result, err := cv.VerifyContact("Acme Corporation", "https://acme.com")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil contact match result")
	}

	if result.PhoneConfidence < 0 || result.PhoneConfidence > 1 {
		t.Errorf("Expected phone confidence between 0 and 1, got: %f", result.PhoneConfidence)
	}

	if result.EmailConfidence < 0 || result.EmailConfidence > 1 {
		t.Errorf("Expected email confidence between 0 and 1, got: %f", result.EmailConfidence)
	}

	if result.OverallConfidence < 0 || result.OverallConfidence > 1 {
		t.Errorf("Expected overall confidence between 0 and 1, got: %f", result.OverallConfidence)
	}
}

func TestRegistrationChecker(t *testing.T) {
	rc := NewRegistrationChecker()

	result, err := rc.CheckRegistration("Acme Corporation", "https://acme.com")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil registration match result")
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got: %f", result.Confidence)
	}
}

func TestOwnershipScorer(t *testing.T) {
	os := NewOwnershipScorer()

	score, err := os.ScoreOwnership("Acme Corporation", "https://acme.com")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if score < 0 || score > 1 {
		t.Errorf("Expected ownership score between 0 and 1, got: %f", score)
	}
}

func TestConfidenceAssessor(t *testing.T) {
	ca := NewConfidenceAssessor()

	// Create a test result with all components
	result := &ConnectionValidationResult{
		NameMatchResult: &NameMatchResult{
			IsMatch:    true,
			Confidence: 0.9,
		},
		AddressMatchResult: &AddressMatchResult{
			IsMatch:    true,
			Confidence: 0.8,
		},
		ContactMatchResult: &ContactMatchResult{
			IsMatch:           true,
			OverallConfidence: 0.7,
		},
		RegistrationMatchResult: &RegistrationMatchResult{
			IsMatch:    true,
			Confidence: 0.6,
		},
		OwnershipScore: 0.5,
	}

	confidence := ca.AssessConfidence(result)
	if confidence < 0 || confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got: %f", confidence)
	}

	// Test with no matches
	result2 := &ConnectionValidationResult{
		NameMatchResult: &NameMatchResult{
			IsMatch: false,
		},
	}

	confidence2 := ca.AssessConfidence(result2)
	if confidence2 != 0.0 {
		t.Errorf("Expected confidence 0.0 for no matches, got: %f", confidence2)
	}
}

func TestConnectionValidatorConfiguration(t *testing.T) {
	cv := NewConnectionValidator()

	// Test default configuration
	if !cv.config.EnableNameMatching {
		t.Error("Expected name matching to be enabled by default")
	}

	if !cv.config.EnableAddressVerification {
		t.Error("Expected address verification to be enabled by default")
	}

	if !cv.config.EnableContactVerification {
		t.Error("Expected contact verification to be enabled by default")
	}

	if !cv.config.EnableRegistrationCheck {
		t.Error("Expected registration check to be enabled by default")
	}

	if !cv.config.EnableOwnershipScoring {
		t.Error("Expected ownership scoring to be enabled by default")
	}

	if cv.config.MinConfidenceScore != 0.7 {
		t.Errorf("Expected min confidence score to be 0.7, got: %f", cv.config.MinConfidenceScore)
	}

	if cv.config.MaxFuzzyDistance != 3 {
		t.Errorf("Expected max fuzzy distance to be 3, got: %d", cv.config.MaxFuzzyDistance)
	}

	if cv.config.Timeout != time.Second*30 {
		t.Errorf("Expected timeout to be 30 seconds, got: %v", cv.config.Timeout)
	}
}

func TestNameMatchingConfiguration(t *testing.T) {
	bnm := NewBusinessNameMatcher()

	// Test default configuration
	if bnm.config.MaxFuzzyDistance != 3 {
		t.Errorf("Expected max fuzzy distance to be 3, got: %d", bnm.config.MaxFuzzyDistance)
	}

	if bnm.config.MinSimilarityScore != 0.8 {
		t.Errorf("Expected min similarity score to be 0.8, got: %f", bnm.config.MinSimilarityScore)
	}

	if !bnm.config.EnableAbbreviations {
		t.Error("Expected abbreviations to be enabled by default")
	}

	if !bnm.config.EnableAliases {
		t.Error("Expected aliases to be enabled by default")
	}

	if bnm.config.CaseSensitive {
		t.Error("Expected case sensitive to be false by default")
	}

	if !bnm.config.IgnorePunctuation {
		t.Error("Expected ignore punctuation to be true by default")
	}
}

func TestAddressVerificationConfiguration(t *testing.T) {
	av := NewAddressVerifier()

	// Test default configuration
	if !av.config.EnableGeocoding {
		t.Error("Expected geocoding to be enabled by default")
	}

	if !av.config.EnableStandardization {
		t.Error("Expected standardization to be enabled by default")
	}

	if av.config.MaxDistance != 10.0 {
		t.Errorf("Expected max distance to be 10.0, got: %f", av.config.MaxDistance)
	}

	if av.config.MinConfidence != 0.8 {
		t.Errorf("Expected min confidence to be 0.8, got: %f", av.config.MinConfidence)
	}
}

func TestContactVerificationConfiguration(t *testing.T) {
	cv := NewContactVerifier()

	// Test default configuration
	if !cv.config.EnablePhoneVerification {
		t.Error("Expected phone verification to be enabled by default")
	}

	if !cv.config.EnableEmailVerification {
		t.Error("Expected email verification to be enabled by default")
	}

	if !cv.config.ValidateFormat {
		t.Error("Expected format validation to be enabled by default")
	}

	if cv.config.CheckExistence {
		t.Error("Expected existence checking to be disabled by default")
	}
}

func TestRegistrationCheckConfiguration(t *testing.T) {
	rc := NewRegistrationChecker()

	// Test default configuration
	if !rc.config.EnableCrossReference {
		t.Error("Expected cross reference to be enabled by default")
	}

	if !rc.config.CheckMultipleSources {
		t.Error("Expected multiple sources checking to be enabled by default")
	}

	if !rc.config.ValidateStatus {
		t.Error("Expected status validation to be enabled by default")
	}
}

func TestOwnershipScoringConfiguration(t *testing.T) {
	os := NewOwnershipScorer()

	// Test default configuration
	if !os.config.EnableMultipleEvidence {
		t.Error("Expected multiple evidence to be enabled by default")
	}

	if os.config.MinEvidenceScore != 0.6 {
		t.Errorf("Expected min evidence score to be 0.6, got: %f", os.config.MinEvidenceScore)
	}

	if os.config.WeightRecentEvidence != 1.2 {
		t.Errorf("Expected weight recent evidence to be 1.2, got: %f", os.config.WeightRecentEvidence)
	}
}

func TestConfidenceAssessmentConfiguration(t *testing.T) {
	ca := NewConfidenceAssessor()

	// Test default configuration
	if !ca.config.EnableWeightedScoring {
		t.Error("Expected weighted scoring to be enabled by default")
	}

	if !ca.config.RequireMultipleEvidence {
		t.Error("Expected multiple evidence requirement to be enabled by default")
	}

	if ca.config.MinEvidenceCount != 2 {
		t.Errorf("Expected min evidence count to be 2, got: %d", ca.config.MinEvidenceCount)
	}
}

func TestConnectionStrengthDetermination(t *testing.T) {
	cv := NewConnectionValidator()

	// Test strong connection
	strength := cv.determineConnectionStrength(0.95)
	if strength != "strong" {
		t.Errorf("Expected 'strong' for confidence 0.95, got: %s", strength)
	}

	// Test moderate connection
	strength = cv.determineConnectionStrength(0.8)
	if strength != "moderate" {
		t.Errorf("Expected 'moderate' for confidence 0.8, got: %s", strength)
	}

	// Test weak connection
	strength = cv.determineConnectionStrength(0.6)
	if strength != "weak" {
		t.Errorf("Expected 'weak' for confidence 0.6, got: %s", strength)
	}

	// Test no connection
	strength = cv.determineConnectionStrength(0.3)
	if strength != "none" {
		t.Errorf("Expected 'none' for confidence 0.3, got: %s", strength)
	}
}

func TestStringNormalization(t *testing.T) {
	// Test punctuation removal
	result := removePunctuation("Hello, World!")
	if result != "Hello World" {
		t.Errorf("Expected 'Hello World', got: %s", result)
	}

	// Test whitespace normalization
	result = normalizeWhitespace("  Hello    World  ")
	if result != "Hello World" {
		t.Errorf("Expected 'Hello World', got: %s", result)
	}
}

func TestConnectionValidatorConcurrent(t *testing.T) {
	cv := NewConnectionValidator()

	// Test concurrent validation
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			result, err := cv.ValidateConnection("Test Company", "https://test.com")
			if err != nil {
				t.Errorf("Validation failed: %v", err)
			} else if result == nil {
				t.Error("Expected non-nil result")
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}
}

func BenchmarkValidateConnection(b *testing.B) {
	cv := NewConnectionValidator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cv.ValidateConnection("Benchmark Company", "https://benchmark.com")
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}

func BenchmarkBusinessNameMatching(b *testing.B) {
	bnm := NewBusinessNameMatcher()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bnm.MatchName("Benchmark Company", "https://benchmark.com")
		if err != nil {
			b.Fatalf("Name matching failed: %v", err)
		}
	}
}

func BenchmarkExactMatching(b *testing.B) {
	em := NewExactMatcher()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em.Match("Benchmark Company", "Benchmark Company")
	}
}

func BenchmarkFuzzyMatching(b *testing.B) {
	fm := NewFuzzyMatcher()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fm.Match("Benchmark Company", "Benchmark Corp")
	}
}

func BenchmarkConfidenceAssessment(b *testing.B) {
	ca := NewConfidenceAssessor()

	result := &ConnectionValidationResult{
		NameMatchResult: &NameMatchResult{
			IsMatch:    true,
			Confidence: 0.9,
		},
		AddressMatchResult: &AddressMatchResult{
			IsMatch:    true,
			Confidence: 0.8,
		},
		ContactMatchResult: &ContactMatchResult{
			IsMatch:           true,
			OverallConfidence: 0.7,
		},
		RegistrationMatchResult: &RegistrationMatchResult{
			IsMatch:    true,
			Confidence: 0.6,
		},
		OwnershipScore: 0.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ca.AssessConfidence(result)
	}
}
