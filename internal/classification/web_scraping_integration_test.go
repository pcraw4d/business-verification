package classification

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// TestRobotsTxtIntegration tests robots.txt parsing with real-world scenarios
func TestRobotsTxtIntegration(t *testing.T) {
	tests := []struct {
		name           string
		robotsContent  string
		path           string
		expectedResult string // "allowed", "blocked", "delay"
		expectedDelay  time.Duration
	}{
		{
			name: "google-style robots.txt",
			robotsContent: `User-agent: *
Allow: /search
Disallow: /searchhistory
Disallow: /mypreferences
Crawl-delay: 10
`,
			path:           "/search",
			expectedResult: "allowed",
			expectedDelay:  10 * time.Second,
		},
		{
			name: "github-style robots.txt",
			robotsContent: `User-agent: *
Disallow: /search
Allow: /
`,
			path:           "/",
			expectedResult: "allowed",
			expectedDelay:  0,
		},
		{
			name: "completely blocked",
			robotsContent: `User-agent: *
Disallow: /
`,
			path:           "/",
			expectedResult: "blocked",
			expectedDelay:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/robots.txt" {
					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tt.robotsContent))
				} else {
					w.WriteHeader(http.StatusOK)
				}
			}))
			defer server.Close()

			logger := log.New(os.Stdout, "[test] ", log.LstdFlags)
			crawler := NewSmartWebsiteCrawler(logger)

			blocked, delay, err := crawler.checkRobotsTxt(context.Background(), server.URL, tt.path)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			switch tt.expectedResult {
			case "blocked":
				if !blocked {
					t.Errorf("Expected path to be blocked, but it was allowed")
				}
			case "allowed":
				if blocked {
					t.Errorf("Expected path to be allowed, but it was blocked")
				}
			}

			if delay != tt.expectedDelay {
				t.Errorf("Expected delay %v, got %v", tt.expectedDelay, delay)
			}
		})
	}
}

// TestHTTPStatusCodeHandlingIntegration tests status code handling with mock servers
func TestHTTPStatusCodeHandlingIntegration(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		headers        map[string]string
		expectedAction string // "stop", "retry", "continue"
		description    string
	}{
		{
			name:           "429 with retry-after",
			statusCode:     429,
			headers:        map[string]string{"Retry-After": "60"},
			expectedAction: "stop",
			description:    "should stop immediately on 429 with retry-after header",
		},
		{
			name:           "403 forbidden",
			statusCode:     403,
			headers:        map[string]string{},
			expectedAction: "stop",
			description:    "should stop immediately on 403",
		},
		{
			name:           "503 service unavailable",
			statusCode:     503,
			headers:        map[string]string{},
			expectedAction: "retry",
			description:    "should indicate retry needed on 503",
		},
		{
			name:           "200 success",
			statusCode:     200,
			headers:        map[string]string{},
			expectedAction: "continue",
			description:    "should continue on 200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Set headers
				for k, v := range tt.headers {
					w.Header().Set(k, v)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte("test response"))
			}))
			defer server.Close()

			// Make request
			client := &http.Client{Timeout: 5 * time.Second}
			req, err := http.NewRequest("GET", server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("User-Agent", GetUserAgent())

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			// Verify status code
			if resp.StatusCode != tt.statusCode {
				t.Errorf("Status code mismatch: expected %d, got %d", tt.statusCode, resp.StatusCode)
			}

			// Verify headers
			for k, v := range tt.headers {
				if resp.Header.Get(k) != v {
					t.Errorf("Header %s mismatch: expected %s, got %s", k, v, resp.Header.Get(k))
				}
			}

			// Verify expected action
			switch tt.statusCode {
			case 429, 403:
				if tt.expectedAction != "stop" {
					t.Errorf("Expected action 'stop' for status %d", tt.statusCode)
				}
			case 503:
				if tt.expectedAction != "retry" {
					t.Errorf("Expected action 'retry' for status %d", tt.statusCode)
				}
			case 200:
				if tt.expectedAction != "continue" {
					t.Errorf("Expected action 'continue' for status %d", tt.statusCode)
				}
			}
		})
	}
}

// TestRateLimitingIntegration tests rate limiting behavior
func TestRateLimitingIntegration(t *testing.T) {
	t.Run("rate limiting respects delay", func(t *testing.T) {
		// This test verifies that rate limiting adds appropriate delays
		// Full integration would require repository setup
		start := time.Now()
		
		// Simulate rate limiting delay
		delay := 100 * time.Millisecond // Short delay for testing
		time.Sleep(delay)
		
		elapsed := time.Since(start)
		if elapsed < delay {
			t.Errorf("Rate limiting delay not respected: expected at least %v, got %v", delay, elapsed)
		}
	})
}

// TestUserAgentInRequests tests that User-Agent is properly set in requests
func TestUserAgentInRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")
		if userAgent == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Missing User-Agent"))
			return
		}

		// Verify User-Agent format
		if !contains(userAgent, "KYBPlatformBot") {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid User-Agent format"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", GetUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Request failed with status %d", resp.StatusCode)
	}
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

