package webanalysis

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewBotDetectionEvasion(t *testing.T) {
	bde := NewBotDetectionEvasion()
	
	if bde == nil {
		t.Fatal("Expected non-nil BotDetectionEvasion")
	}
	
	if bde.fingerprintManager == nil {
		t.Error("Expected non-nil FingerprintManager")
	}
	
	if bde.requestRandomizer == nil {
		t.Error("Expected non-nil RequestRandomizer")
	}
	
	if bde.captchaDetector == nil {
		t.Error("Expected non-nil CAPTCHADetector")
	}
	
	if bde.behaviorSimulator == nil {
		t.Error("Expected non-nil BehaviorSimulator")
	}
}

func TestGenerateRandomizedRequest(t *testing.T) {
	bde := NewBotDetectionEvasion()
	
	req, err := bde.GenerateRandomizedRequest("https://example.com", "GET")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if req == nil {
		t.Fatal("Expected non-nil request")
	}
	
	if req.URL.Host != "example.com" || req.URL.Scheme != "https" {
		t.Errorf("Expected URL scheme 'https' and host 'example.com', got: %s", req.URL.String())
	}
	
	if req.Method != "GET" {
		t.Errorf("Expected method 'GET', got: %s", req.Method)
	}
	
	// Check that headers are set
	if req.Header.Get("User-Agent") == "" {
		t.Error("Expected User-Agent header to be set")
	}
	
	if req.Header.Get("Accept") == "" {
		t.Error("Expected Accept header to be set")
	}
	
	if req.Header.Get("Accept-Language") == "" {
		t.Error("Expected Accept-Language header to be set")
	}
}

func TestFingerprintManager(t *testing.T) {
	fm := NewFingerprintManager()
	
	// Test user agent randomization
	ua1 := fm.getRandomUserAgent()
	ua2 := fm.getRandomUserAgent()
	
	if ua1 == "" {
		t.Error("Expected non-empty user agent")
	}
	
	if ua2 == "" {
		t.Error("Expected non-empty user agent")
	}
	
	// Test language randomization
	lang1 := fm.getRandomLanguage()
	lang2 := fm.getRandomLanguage()
	
	if lang1 == "" {
		t.Error("Expected non-empty language")
	}
	
	if lang2 == "" {
		t.Error("Expected non-empty language")
	}
	
	// Test platform randomization
	platform := fm.getRandomPlatform()
	if platform == "" {
		t.Error("Expected non-empty platform")
	}
}

func TestRequestRandomizer(t *testing.T) {
	rr := NewRequestRandomizer()
	
	// Test header variation randomization
	accept := rr.getRandomVariation(rr.headerVariations["Accept"])
	if accept == "" {
		t.Error("Expected non-empty Accept header variation")
	}
	
	// Test cookie generation
	cookies := rr.generateRandomCookies()
	if len(cookies) == 0 {
		t.Error("Expected non-empty cookies")
	}
	
	// Test query parameter generation
	queryParams := rr.generateRandomQueryParams()
	if queryParams == "" {
		t.Error("Expected non-empty query parameters")
	}
	
	if !strings.Contains(queryParams, "_=") {
		t.Error("Expected query parameters to contain timestamp")
	}
}

func TestCAPTCHADetection(t *testing.T) {
	bde := NewBotDetectionEvasion()
	
	// Test CAPTCHA detection with CAPTCHA content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body><div class='recaptcha'>Please complete the CAPTCHA</div></body></html>"))
	}))
	defer server.Close()
	
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make test request: %v", err)
	}
	
	if !bde.DetectCAPTCHA(resp) {
		t.Error("Expected CAPTCHA to be detected")
	}
	
	// Test CAPTCHA detection with normal content
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body><h1>Welcome to our website</h1></body></html>"))
	}))
	defer server2.Close()
	
	resp2, err := http.Get(server2.URL)
	if err != nil {
		t.Fatalf("Failed to make test request: %v", err)
	}
	
	if bde.DetectCAPTCHA(resp2) {
		t.Error("Expected no CAPTCHA to be detected")
	}
}

func TestBehaviorSimulation(t *testing.T) {
	bs := NewBehaviorSimulator()
	
	// Test that behavior simulation doesn't panic
	start := time.Now()
	bs.simulateMouseMovement()
	bs.simulateKeyboardInput()
	bs.simulateScrolling()
	bs.simulateClicking()
	duration := time.Since(start)
	
	// Should take some time due to delays
	if duration < time.Millisecond*100 {
		t.Error("Expected behavior simulation to take some time")
	}
}

func TestRandomIntGeneration(t *testing.T) {
	// Test that random int generation works correctly
	bde := NewBotDetectionEvasion()
	
	for i := 0; i < 100; i++ {
		val := bde.randomInt(1, 10)
		if val < 1 || val >= 10 {
			t.Errorf("Expected value between 1 and 9, got: %d", val)
		}
	}
}

func TestJitterGeneration(t *testing.T) {
	bde := NewBotDetectionEvasion()
	
	for i := 0; i < 10; i++ {
		jitter := bde.generateJitter()
		if jitter < 0 || jitter > bde.config.JitterRange {
			t.Errorf("Expected jitter between 0 and %v, got: %v", bde.config.JitterRange, jitter)
		}
	}
}

func TestEvasionConfig(t *testing.T) {
	bde := NewBotDetectionEvasion()
	
	// Test default configuration
	if !bde.config.EnableFingerprintRandomization {
		t.Error("Expected fingerprint randomization to be enabled by default")
	}
	
	if !bde.config.EnableRequestRandomization {
		t.Error("Expected request randomization to be enabled by default")
	}
	
	if !bde.config.EnableCAPTCHADetection {
		t.Error("Expected CAPTCHA detection to be enabled by default")
	}
	
	if !bde.config.EnableBehaviorSimulation {
		t.Error("Expected behavior simulation to be enabled by default")
	}
	
	if bde.config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries to be 3, got: %d", bde.config.MaxRetries)
	}
	
	if bde.config.RetryDelay != time.Second*2 {
		t.Errorf("Expected RetryDelay to be 2 seconds, got: %v", bde.config.RetryDelay)
	}
}

func TestConcurrentAccess(t *testing.T) {
	bde := NewBotDetectionEvasion()
	
	// Test concurrent access to fingerprint manager
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				bde.fingerprintManager.getRandomUserAgent()
				bde.fingerprintManager.getRandomLanguage()
			}
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestRequestRandomizerConcurrent(t *testing.T) {
	rr := NewRequestRandomizer()
	
	// Test concurrent access to request randomizer
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				rr.getRandomVariation(rr.headerVariations["Accept"])
				rr.generateRandomCookies()
				rr.generateRandomQueryParams()
			}
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func BenchmarkGenerateRandomizedRequest(b *testing.B) {
	bde := NewBotDetectionEvasion()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bde.GenerateRandomizedRequest("https://example.com", "GET")
		if err != nil {
			b.Fatalf("Failed to generate request: %v", err)
		}
	}
}

func BenchmarkFingerprintRandomization(b *testing.B) {
	fm := NewFingerprintManager()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fm.getRandomUserAgent()
		fm.getRandomLanguage()
		fm.getRandomPlatform()
	}
}

func BenchmarkRequestRandomization(b *testing.B) {
	rr := NewRequestRandomizer()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr.getRandomVariation(rr.headerVariations["Accept"])
		rr.generateRandomCookies()
		rr.generateRandomQueryParams()
	}
}
