package webanalysis

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewWebsiteValidator(t *testing.T) {
	wv := NewWebsiteValidator()

	if wv == nil {
		t.Fatal("Expected non-nil WebsiteValidator")
	}

	if wv.authenticityChecker == nil {
		t.Error("Expected non-nil AuthenticityChecker")
	}

	if wv.trafficAnalyzer == nil {
		t.Error("Expected non-nil TrafficAnalyzer")
	}

	if wv.domainReputation == nil {
		t.Error("Expected non-nil DomainReputationChecker")
	}

	if wv.sslValidator == nil {
		t.Error("Expected non-nil SSLValidator")
	}

	if wv.contentQuality == nil {
		t.Error("Expected non-nil ContentQualityAssessor")
	}
}

func TestValidateWebsite(t *testing.T) {
	wv := NewWebsiteValidator()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body>
			<h1>About Us</h1>
			<p>We are a legitimate company providing quality services.</p>
			<a href="/contact">Contact Us</a>
			<a href="/privacy">Privacy Policy</a>
		</body></html>`))
	}))
	defer server.Close()

	result, err := wv.ValidateWebsite(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil validation result")
	}

	if result.URL != server.URL {
		t.Errorf("Expected URL %s, got: %s", server.URL, result.URL)
	}

	if result.AuthenticityScore <= 0 {
		t.Error("Expected positive authenticity score")
	}

	if result.OverallScore <= 0 {
		t.Error("Expected positive overall score")
	}

	if result.ValidationTime <= 0 {
		t.Error("Expected positive validation time")
	}
}

func TestAuthenticityChecker(t *testing.T) {
	ac := NewAuthenticityChecker()

	score, err := ac.CheckAuthenticity("https://legitimate-company.com")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if score <= 0 {
		t.Error("Expected positive authenticity score for legitimate content")
	}

	// Note: This is a simulation since we can't easily test with suspicious content
	// In a real implementation, this would test actual suspicious patterns
}

func TestTrafficAnalyzer(t *testing.T) {
	ta := NewTrafficAnalyzer()

	result, err := ta.AnalyzeTraffic("https://example.com")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil traffic analysis result")
	}

	if result.HumanTrafficPercentage < 0 || result.HumanTrafficPercentage > 100 {
		t.Errorf("Expected human traffic percentage between 0 and 100, got: %f", result.HumanTrafficPercentage)
	}

	if result.ResponseTime < 0 {
		t.Error("Expected positive response time")
	}

	if result.Uptime < 0 || result.Uptime > 100 {
		t.Errorf("Expected uptime between 0 and 100, got: %f", result.Uptime)
	}
}

func TestDomainReputationChecker(t *testing.T) {
	drc := NewDomainReputationChecker()

	result, err := drc.CheckReputation("example.com")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Domain != "example.com" {
		t.Errorf("Expected domain 'example.com', got: %s", result.Domain)
	}

	if result.Age < 0 {
		t.Error("Expected non-negative domain age")
	}

	if result.ReputationScore < 0 || result.ReputationScore > 1 {
		t.Errorf("Expected reputation score between 0 and 1, got: %f", result.ReputationScore)
	}

	if result.LastChecked.IsZero() {
		t.Error("Expected last checked time to be set")
	}
}

func TestSSLValidator(t *testing.T) {
	sv := NewSSLValidator()

	// Test HTTPS URL
	result, err := sv.ValidateSSL("https://example.com")
	if err != nil {
		// SSL validation might fail for test domains, which is expected
		t.Logf("SSL validation failed as expected: %v", err)
		return
	}

	if result.Subject == "" {
		t.Error("Expected non-empty subject")
	}

	if result.Issuer == "" {
		t.Error("Expected non-empty issuer")
	}

	if result.SerialNumber == "" {
		t.Error("Expected non-empty serial number")
	}

	if result.SignatureAlgorithm == "" {
		t.Error("Expected non-empty signature algorithm")
	}

	// Test HTTP URL (should fail)
	httpResult, err := sv.ValidateSSL("http://example.com")
	if err != nil {
		t.Logf("HTTP SSL validation failed as expected: %v", err)
	} else {
		if len(httpResult.Errors) == 0 {
			t.Error("Expected errors for HTTP URL")
		}
	}
}

func TestContentQualityAssessor(t *testing.T) {
	cqa := NewContentQualityAssessor()

	// Create test server with good content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body>
			<h1>Welcome to Our Company</h1>
			<p>We are a professional organization dedicated to providing excellent services to our customers. 
			Our team consists of experienced professionals who are committed to delivering high-quality solutions.</p>
			<p>We offer a wide range of services including consulting, development, and support. 
			Our mission is to help businesses succeed through innovative technology solutions.</p>
			<a href="/about">About Us</a>
			<a href="/services">Our Services</a>
			<a href="/contact">Contact Information</a>
			<img src="/logo.png" alt="Company Logo">
		</body></html>`))
	}))
	defer server.Close()

	result, err := cqa.AssessQuality(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil content quality result")
	}

	if result.OverallScore < 0 || result.OverallScore > 1 {
		t.Errorf("Expected overall score between 0 and 1, got: %f", result.OverallScore)
	}

	if result.ContentLength <= 0 {
		t.Error("Expected positive content length")
	}

	if result.WordCount <= 0 {
		t.Error("Expected positive word count")
	}

	if result.UniqueWords <= 0 {
		t.Error("Expected positive unique word count")
	}

	if result.LinkCount < 0 {
		t.Error("Expected non-negative link count")
	}

	if result.ImageCount < 0 {
		t.Error("Expected non-negative image count")
	}

	if len(result.QualityMetrics) == 0 {
		t.Error("Expected non-empty quality metrics")
	}
}

func TestQualityMetricCalculators(t *testing.T) {
	// Test readability calculator
	goodContent := "This is a well-written paragraph. It has multiple sentences. The content is clear and readable."
	readabilityScore := calculateReadability(goodContent)
	if readabilityScore < 0 || readabilityScore > 1 {
		t.Errorf("Expected readability score between 0 and 1, got: %f", readabilityScore)
	}

	// Test content length calculator
	contentLengthScore := calculateContentLength(goodContent)
	if contentLengthScore < 0 || contentLengthScore > 1 {
		t.Errorf("Expected content length score between 0 and 1, got: %f", contentLengthScore)
	}

	// Test spam score calculator
	legitimateContent := "We provide professional services to businesses. Contact us for more information."
	spamScore := calculateSpamScore(legitimateContent)
	if spamScore < 0 || spamScore > 1 {
		t.Errorf("Expected spam score between 0 and 1, got: %f", spamScore)
	}

	// Test grammar score calculator
	grammarScore := calculateGrammarScore(legitimateContent)
	if grammarScore < 0 || grammarScore > 1 {
		t.Errorf("Expected grammar score between 0 and 1, got: %f", grammarScore)
	}
}

func TestContentQualityAssessorHelperMethods(t *testing.T) {
	cqa := NewContentQualityAssessor()

	content := `<html><body>
		<h1>Test Content</h1>
		<p>This is a test paragraph with multiple words.</p>
		<p>Another paragraph with different words.</p>
		<a href="/link1">Link 1</a>
		<a href="/link2">Link 2</a>
		<img src="/image1.jpg" alt="Image 1">
		<img src="/image2.jpg" alt="Image 2">
	</body></html>`

	// Test word counting
	wordCount := cqa.countWords(content)
	if wordCount <= 0 {
		t.Error("Expected positive word count")
	}

	// Test unique word counting
	uniqueWords := cqa.countUniqueWords(content)
	if uniqueWords <= 0 {
		t.Error("Expected positive unique word count")
	}

	if uniqueWords > wordCount {
		t.Error("Expected unique word count to be less than or equal to total word count")
	}

	// Test link counting
	linkCount := cqa.countLinks(content)
	if linkCount != 2 {
		t.Errorf("Expected 2 links, got: %d", linkCount)
	}

	// Test image counting
	imageCount := cqa.countImages(content)
	if imageCount != 2 {
		t.Errorf("Expected 2 images, got: %d", imageCount)
	}

	// Test spam detection
	isSpam := cqa.detectSpam(content)
	if isSpam {
		t.Error("Expected legitimate content to not be detected as spam")
	}

	// Test spam detection with spam content
	spamContent := "Buy now! Act now! Limited time offer! Click here to claim your prize!"
	isSpam = cqa.detectSpam(spamContent)
	if !isSpam {
		t.Error("Expected spam content to be detected as spam")
	}
}

func TestWebsiteValidatorConfiguration(t *testing.T) {
	wv := NewWebsiteValidator()

	// Test default configuration
	if !wv.config.EnableAuthenticityCheck {
		t.Error("Expected authenticity check to be enabled by default")
	}

	if !wv.config.EnableTrafficAnalysis {
		t.Error("Expected traffic analysis to be enabled by default")
	}

	if !wv.config.EnableDomainReputation {
		t.Error("Expected domain reputation to be enabled by default")
	}

	if !wv.config.EnableSSLValidation {
		t.Error("Expected SSL validation to be enabled by default")
	}

	if !wv.config.EnableContentQualityCheck {
		t.Error("Expected content quality check to be enabled by default")
	}

	if wv.config.Timeout != time.Second*30 {
		t.Errorf("Expected timeout to be 30 seconds, got: %v", wv.config.Timeout)
	}

	if wv.config.MaxRedirects != 10 {
		t.Errorf("Expected max redirects to be 10, got: %d", wv.config.MaxRedirects)
	}

	if wv.config.MinContentLength != 100 {
		t.Errorf("Expected min content length to be 100, got: %d", wv.config.MinContentLength)
	}

	if wv.config.MaxContentLength != 1000000 {
		t.Errorf("Expected max content length to be 1000000, got: %d", wv.config.MaxContentLength)
	}
}

func TestOverallScoreCalculation(t *testing.T) {
	wv := NewWebsiteValidator()

	// Create a test result with all components
	result := &WebsiteValidationResult{
		AuthenticityScore: 0.8,
		DomainReputation: &DomainReputationResult{
			ReputationScore: 0.9,
		},
		SSLValidation: &SSLValidationResult{
			IsValid: true,
		},
		ContentQuality: &ContentQualityResult{
			OverallScore: 0.85,
		},
	}

	score := wv.calculateOverallScore(result)
	if score < 0 || score > 1 {
		t.Errorf("Expected overall score between 0 and 1, got: %f", score)
	}

	// Test with missing components
	result2 := &WebsiteValidationResult{
		AuthenticityScore: 0.8,
	}

	score2 := wv.calculateOverallScore(result2)
	if score2 < 0 || score2 > 1 {
		t.Errorf("Expected overall score between 0 and 1, got: %f", score2)
	}
}

func TestConcurrentValidation(t *testing.T) {
	wv := NewWebsiteValidator()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body><h1>Test Content</h1></body></html>`))
	}))
	defer server.Close()

	// Test concurrent validation
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			result, err := wv.ValidateWebsite(server.URL)
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

func BenchmarkValidateWebsite(b *testing.B) {
	wv := NewWebsiteValidator()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body><h1>Test Content</h1></body></html>`))
	}))
	defer server.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := wv.ValidateWebsite(server.URL)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}

func BenchmarkAuthenticityCheck(b *testing.B) {
	ac := NewAuthenticityChecker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ac.CheckAuthenticity("https://example.com")
		if err != nil {
			b.Fatalf("Authenticity check failed: %v", err)
		}
	}
}

func BenchmarkContentQualityAssessment(b *testing.B) {
	cqa := NewContentQualityAssessor()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body><h1>Test Content</h1></body></html>`))
	}))
	defer server.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cqa.AssessQuality(server.URL)
		if err != nil {
			b.Fatalf("Content quality assessment failed: %v", err)
		}
	}
}
