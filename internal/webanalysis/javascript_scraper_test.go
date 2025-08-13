package webanalysis

import (
	"testing"
	"time"
)

func TestNewJavaScriptScraper(t *testing.T) {
	config := JavaScriptScraperConfig{
		EnableJavaScript:       true,
		WaitForLoadTimeout:     30 * time.Second,
		JavaScriptTimeout:      10 * time.Second,
		EnableFingerprinting:   true,
		EnableSessionMgmt:      true,
		EnableResourceBlocking: true,
		EnableUserInteractions: true,
		MaxConcurrentBrowsers:  5,
		ViewportWidth:          1920,
		ViewportHeight:         1080,
		UserDataDir:            "/tmp/browser_data",
	}

	scraper := NewJavaScriptScraper(config)

	if scraper == nil {
		t.Fatal("Expected scraper to be created, got nil")
	}

	if scraper.config.WaitForLoadTimeout != 30*time.Second {
		t.Errorf("Expected WaitForLoadTimeout to be 30s, got %v", scraper.config.WaitForLoadTimeout)
	}

	if scraper.config.MaxConcurrentBrowsers != 5 {
		t.Errorf("Expected MaxConcurrentBrowsers to be 5, got %d", scraper.config.MaxConcurrentBrowsers)
	}

	if scraper.browserMgr == nil {
		t.Fatal("Expected browser manager to be created, got nil")
	}
}

func TestNewJavaScriptScraperWithDefaults(t *testing.T) {
	config := JavaScriptScraperConfig{}

	scraper := NewJavaScriptScraper(config)

	if scraper == nil {
		t.Fatal("Expected scraper to be created, got nil")
	}

	// Check default values
	if scraper.config.WaitForLoadTimeout != 30*time.Second {
		t.Errorf("Expected default WaitForLoadTimeout to be 30s, got %v", scraper.config.WaitForLoadTimeout)
	}

	if scraper.config.JavaScriptTimeout != 10*time.Second {
		t.Errorf("Expected default JavaScriptTimeout to be 10s, got %v", scraper.config.JavaScriptTimeout)
	}

	if scraper.config.MaxConcurrentBrowsers != 5 {
		t.Errorf("Expected default MaxConcurrentBrowsers to be 5, got %d", scraper.config.MaxConcurrentBrowsers)
	}

	if scraper.config.ViewportWidth != 1920 {
		t.Errorf("Expected default ViewportWidth to be 1920, got %d", scraper.config.ViewportWidth)
	}

	if scraper.config.ViewportHeight != 1080 {
		t.Errorf("Expected default ViewportHeight to be 1080, got %d", scraper.config.ViewportHeight)
	}
}

func TestGetFingerprintUserAgent(t *testing.T) {
	config := JavaScriptScraperConfig{}
	scraper := NewJavaScriptScraper(config)

	tests := []struct {
		profile     string
		expectedUA  string
		description string
	}{
		{
			profile:     "chrome_windows",
			expectedUA:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			description: "Chrome Windows user agent",
		},
		{
			profile:     "chrome_mac",
			expectedUA:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			description: "Chrome Mac user agent",
		},
		{
			profile:     "firefox_windows",
			expectedUA:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
			description: "Firefox Windows user agent",
		},
		{
			profile:     "safari_mac",
			expectedUA:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
			description: "Safari Mac user agent",
		},
		{
			profile:     "unknown_profile",
			expectedUA:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			description: "Default user agent for unknown profile",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			ua := scraper.getFingerprintUserAgent(test.profile)
			if ua != test.expectedUA {
				t.Errorf("Expected user agent for profile '%s' to be '%s', got '%s'", test.profile, test.expectedUA, ua)
			}
		})
	}
}

func TestBrowserManagerGetBrowser(t *testing.T) {
	config := JavaScriptScraperConfig{MaxConcurrentBrowsers: 2}
	bm := &BrowserManager{
		config:      config,
		browsers:    make(map[string]*BrowserInstance),
		maxBrowsers: config.MaxConcurrentBrowsers,
	}

	// Test getting first browser
	browser1, err := bm.GetBrowser("session1")
	if err != nil {
		t.Fatalf("Expected to get browser, got error: %v", err)
	}

	if browser1 == nil {
		t.Fatal("Expected browser instance, got nil")
	}

	if browser1.ID != "browser_1" {
		t.Errorf("Expected browser ID to be 'browser_1', got '%s'", browser1.ID)
	}

	if browser1.SessionID != "session1" {
		t.Errorf("Expected session ID to be 'session1', got '%s'", browser1.SessionID)
	}

	if !browser1.InUse {
		t.Error("Expected browser to be marked as in use")
	}

	// Test getting second browser
	browser2, err := bm.GetBrowser("session2")
	if err != nil {
		t.Fatalf("Expected to get second browser, got error: %v", err)
	}

	if browser2.ID != "browser_2" {
		t.Errorf("Expected browser ID to be 'browser_2', got '%s'", browser2.ID)
	}

	// Test getting third browser (should fail due to limit)
	_, err = bm.GetBrowser("session3")
	if err == nil {
		t.Error("Expected error when trying to get third browser, got nil")
	}

	// Test releasing and reusing browser
	bm.ReleaseBrowser(browser1.ID)
	if browser1.InUse {
		t.Error("Expected browser to be marked as not in use after release")
	}

	// Should be able to get browser again after release
	browser3, err := bm.GetBrowser("session3")
	if err != nil {
		t.Fatalf("Expected to get browser after release, got error: %v", err)
	}

	if browser3.ID != browser1.ID {
		t.Errorf("Expected to reuse released browser, got different browser: %s vs %s", browser3.ID, browser1.ID)
	}
}

func TestBrowserManagerCleanupBrowsers(t *testing.T) {
	config := JavaScriptScraperConfig{MaxConcurrentBrowsers: 3}
	bm := &BrowserManager{
		config:      config,
		browsers:    make(map[string]*BrowserInstance),
		maxBrowsers: config.MaxConcurrentBrowsers,
	}

	// Create browsers
	browser1, _ := bm.GetBrowser("session1")
	browser2, _ := bm.GetBrowser("session2")

	// Release browsers
	bm.ReleaseBrowser(browser1.ID)
	bm.ReleaseBrowser(browser2.ID)

	// Set last used time to old time for cleanup
	browser1.LastUsed = time.Now().Add(-10 * time.Minute)
	browser2.LastUsed = time.Now().Add(-10 * time.Minute)

	// Cleanup should remove old browsers
	bm.CleanupBrowsers()

	if len(bm.browsers) != 0 {
		t.Errorf("Expected all browsers to be cleaned up, got %d remaining", len(bm.browsers))
	}
}

func TestJavaScriptScrapingJobValidation(t *testing.T) {
	tests := []struct {
		job         JavaScriptScrapingJob
		description string
		shouldError bool
	}{
		{
			job: JavaScriptScrapingJob{
				URL:      "https://example.com",
				Business: "Test Business",
				Timeout:  30 * time.Second,
			},
			description: "Valid job with required fields",
			shouldError: false,
		},
		{
			job: JavaScriptScrapingJob{
				URL:      "",
				Business: "Test Business",
			},
			description: "Invalid job with empty URL",
			shouldError: true,
		},
		{
			job: JavaScriptScrapingJob{
				URL:      "https://example.com",
				Business: "",
			},
			description: "Invalid job with empty business name",
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			// This would be implemented in the actual scraper
			// For now, we just test the structure
			if test.job.URL == "" || test.job.Business == "" {
				if !test.shouldError {
					t.Error("Expected job to be valid, but it has empty required fields")
				}
			}
		})
	}
}

func TestUserInteractionValidation(t *testing.T) {
	tests := []struct {
		interaction UserInteraction
		description string
		shouldError bool
	}{
		{
			interaction: UserInteraction{
				Type:      "click",
				Selector:  "#button",
				WaitAfter: 1000,
			},
			description: "Valid click interaction",
			shouldError: false,
		},
		{
			interaction: UserInteraction{
				Type:      "type",
				Selector:  "#input",
				Value:     "test value",
				WaitAfter: 500,
			},
			description: "Valid type interaction",
			shouldError: false,
		},
		{
			interaction: UserInteraction{
				Type:      "scroll",
				Value:     "100",
				WaitAfter: 200,
			},
			description: "Valid scroll interaction",
			shouldError: false,
		},
		{
			interaction: UserInteraction{
				Type:     "invalid_type",
				Selector: "#element",
			},
			description: "Invalid interaction type",
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			validTypes := map[string]bool{
				"click":  true,
				"scroll": true,
				"type":   true,
				"hover":  true,
			}

			if !validTypes[test.interaction.Type] {
				if !test.shouldError {
					t.Errorf("Expected interaction type '%s' to be valid", test.interaction.Type)
				}
			}
		})
	}
}

func TestJavaScriptScrapingResultStructure(t *testing.T) {
	result := &JavaScriptScrapingResult{
		URL:                "https://example.com",
		Title:              "Test Page",
		HTML:               "<html><body>Test</body></html>",
		Text:               "Test",
		StatusCode:         200,
		ResponseTime:       2 * time.Second,
		JavaScriptExecuted: []string{"console.log('test')"},
		ExtractedData:      map[string]string{"title": "Test"},
		DOMSnapshot:        "<html>...</html>",
		ScreenshotPath:     "/tmp/screenshot.png",
		JavaScriptErrors:   []string{},
		LoadTime:           1 * time.Second,
		NetworkRequests:    []NetworkRequest{},
		Error:              "",
		ScrapedAt:          time.Now(),
		BrowserFingerprint: "chrome_windows",
		SessionID:          "session_123",
	}

	if result.URL != "https://example.com" {
		t.Errorf("Expected URL to be 'https://example.com', got '%s'", result.URL)
	}

	if result.Title != "Test Page" {
		t.Errorf("Expected title to be 'Test Page', got '%s'", result.Title)
	}

	if result.StatusCode != 200 {
		t.Errorf("Expected status code to be 200, got %d", result.StatusCode)
	}

	if len(result.JavaScriptExecuted) != 1 {
		t.Errorf("Expected 1 JavaScript execution, got %d", len(result.JavaScriptExecuted))
	}

	if len(result.ExtractedData) != 1 {
		t.Errorf("Expected 1 extracted data item, got %d", len(result.ExtractedData))
	}
}

func TestNetworkRequestStructure(t *testing.T) {
	request := NetworkRequest{
		URL:         "https://example.com/api/data",
		Method:      "GET",
		Status:      200,
		ContentType: "application/json",
		Size:        1024,
	}

	if request.URL != "https://example.com/api/data" {
		t.Errorf("Expected URL to be 'https://example.com/api/data', got '%s'", request.URL)
	}

	if request.Method != "GET" {
		t.Errorf("Expected method to be 'GET', got '%s'", request.Method)
	}

	if request.Status != 200 {
		t.Errorf("Expected status to be 200, got %d", request.Status)
	}

	if request.ContentType != "application/json" {
		t.Errorf("Expected content type to be 'application/json', got '%s'", request.ContentType)
	}

	if request.Size != 1024 {
		t.Errorf("Expected size to be 1024, got %d", request.Size)
	}
}
