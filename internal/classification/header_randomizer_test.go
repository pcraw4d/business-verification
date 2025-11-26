package classification

import (
	"os"
	"testing"
)

func TestGetRandomizedHeaders(t *testing.T) {
	baseUserAgent := "Mozilla/5.0 (compatible; KYBPlatformBot/1.0; +https://kyb-platform.com/bot-info; Business Verification)"

	t.Run("maintains_user_agent", func(t *testing.T) {
		hr := NewHeaderRandomizer()
		headers := hr.GetRandomizedHeaders(baseUserAgent)

		if headers["User-Agent"] != baseUserAgent {
			t.Errorf("Expected User-Agent to be %s, got %s", baseUserAgent, headers["User-Agent"])
		}
	})

	t.Run("includes_required_headers", func(t *testing.T) {
		hr := NewHeaderRandomizer()
		headers := hr.GetRandomizedHeaders(baseUserAgent)

		requiredHeaders := []string{"Accept", "Accept-Language", "Accept-Encoding", "Connection", "Upgrade-Insecure-Requests"}
		for _, header := range requiredHeaders {
			if headers[header] == "" {
				t.Errorf("Expected header %s to be set", header)
			}
		}
	})

	t.Run("respects_disabled_setting", func(t *testing.T) {
		os.Setenv("SCRAPING_HEADER_RANDOMIZATION_ENABLED", "false")
		defer os.Unsetenv("SCRAPING_HEADER_RANDOMIZATION_ENABLED")

		hr := NewHeaderRandomizer()
		headers := hr.GetRandomizedHeaders(baseUserAgent)

		if headers["User-Agent"] != baseUserAgent {
			t.Errorf("Expected User-Agent to be maintained even when disabled")
		}

		// Should still have basic headers
		if headers["Accept"] == "" {
			t.Errorf("Expected basic headers even when randomization is disabled")
		}
	})

	t.Run("with_referer", func(t *testing.T) {
		hr := NewHeaderRandomizer()
		referer := "https://example.com/previous-page"
		headers := hr.GetRandomizedHeadersWithReferer(baseUserAgent, referer)

		if headers["Referer"] != referer {
			t.Errorf("Expected Referer to be %s, got %s", referer, headers["Referer"])
		}
	})
}

func TestHeaderRandomizer_IsEnabled(t *testing.T) {
	t.Run("default_enabled", func(t *testing.T) {
		os.Unsetenv("SCRAPING_HEADER_RANDOMIZATION_ENABLED")
		hr := NewHeaderRandomizer()
		if !hr.IsEnabled() {
			t.Error("Expected header randomization to be enabled by default")
		}
	})

	t.Run("explicitly_disabled", func(t *testing.T) {
		os.Setenv("SCRAPING_HEADER_RANDOMIZATION_ENABLED", "false")
		defer os.Unsetenv("SCRAPING_HEADER_RANDOMIZATION_ENABLED")

		hr := NewHeaderRandomizer()
		if hr.IsEnabled() {
			t.Error("Expected header randomization to be disabled")
		}
	})
}

