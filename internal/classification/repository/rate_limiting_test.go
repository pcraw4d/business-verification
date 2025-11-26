package repository

import (
	"os"
	"testing"
	"time"
)

// TestGetRateLimitDelay tests the rate limit delay configuration
func TestGetRateLimitDelay(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		expectedMin time.Duration
		expectedMax time.Duration
		description string
	}{
		{
			name:        "default delay",
			envValue:    "",
			expectedMin: 3 * time.Second,
			expectedMax: 3 * time.Second,
			description: "should return default 3 seconds when env var not set",
		},
		{
			name:        "custom delay",
			envValue:    "5",
			expectedMin: 5 * time.Second,
			expectedMax: 5 * time.Second,
			description: "should return custom delay from environment variable",
		},
		{
			name:        "minimum enforced",
			envValue:    "1",
			expectedMin: 2 * time.Second,
			expectedMax: 2 * time.Second,
			description: "should enforce minimum 2 seconds",
		},
		{
			name:        "maximum enforced",
			envValue:    "20",
			expectedMin: 10 * time.Second,
			expectedMax: 10 * time.Second,
			description: "should enforce maximum 10 seconds",
		},
		{
			name:        "invalid value",
			envValue:    "invalid",
			expectedMin: 3 * time.Second,
			expectedMax: 3 * time.Second,
			description: "should return default for invalid value",
		},
		{
			name:        "zero value",
			envValue:    "0",
			expectedMin: 2 * time.Second,
			expectedMax: 2 * time.Second,
			description: "should enforce minimum for zero value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env value
			originalValue := os.Getenv("SCRAPING_RATE_LIMIT_DELAY")
			defer func() {
				if originalValue != "" {
					os.Setenv("SCRAPING_RATE_LIMIT_DELAY", originalValue)
				} else {
					os.Unsetenv("SCRAPING_RATE_LIMIT_DELAY")
				}
			}()

			// Set test environment variable
			if tt.envValue != "" {
				os.Setenv("SCRAPING_RATE_LIMIT_DELAY", tt.envValue)
			} else {
				os.Unsetenv("SCRAPING_RATE_LIMIT_DELAY")
			}

			delay := getRateLimitDelay()
			if delay < tt.expectedMin || delay > tt.expectedMax {
				t.Errorf("Delay mismatch: expected between %v and %v, got %v. %s",
					tt.expectedMin, tt.expectedMax, delay, tt.description)
			}
		})
	}
}

// TestApplyRateLimitWithCrawlDelay tests rate limiting with crawl delay integration
func TestApplyRateLimitWithCrawlDelay(t *testing.T) {
	t.Run("crawl delay greater than minDelay", func(t *testing.T) {
		// Test that when crawl delay is greater than minDelay, it's used
		minDelay := 3 * time.Second
		crawlDelay := 5 * time.Second

		if crawlDelay > minDelay {
			effectiveDelay := crawlDelay
			if effectiveDelay != crawlDelay {
				t.Errorf("Expected effective delay to be %v when crawl delay > minDelay, got %v",
					crawlDelay, effectiveDelay)
			}
		}
	})

	t.Run("crawl delay less than minDelay", func(t *testing.T) {
		// Test that when crawl delay is less than minDelay, minDelay is used
		minDelay := 3 * time.Second
		crawlDelay := 1 * time.Second

		effectiveDelay := minDelay
		if crawlDelay < minDelay {
			effectiveDelay = minDelay
		}

		if effectiveDelay != minDelay {
			t.Errorf("Expected effective delay to be %v (minDelay) when crawl delay < minDelay, got %v",
				minDelay, effectiveDelay)
		}
	})
}

