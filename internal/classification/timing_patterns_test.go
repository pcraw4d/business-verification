package classification

import (
	"os"
	"testing"
	"time"
)

func TestGetHumanLikeDelay(t *testing.T) {
	baseDelay := 3 * time.Second
	domain := "example.com"

	t.Run("generates_delay", func(t *testing.T) {
		tpg := NewTimingPatternGenerator()
		delay := tpg.GetHumanLikeDelay(baseDelay, domain)

		if delay < baseDelay {
			t.Errorf("Expected delay to be at least %v, got %v", baseDelay, delay)
		}
	})

	t.Run("respects_minimum", func(t *testing.T) {
		tpg := NewTimingPatternGenerator()
		delay := tpg.GetHumanLikeDelay(baseDelay, domain)

		if delay < baseDelay {
			t.Errorf("Expected delay to respect minimum of %v, got %v", baseDelay, delay)
		}
	})

	t.Run("respects_disabled_setting", func(t *testing.T) {
		os.Setenv("SCRAPING_HUMAN_LIKE_TIMING_ENABLED", "false")
		defer os.Unsetenv("SCRAPING_HUMAN_LIKE_TIMING_ENABLED")

		tpg := NewTimingPatternGenerator()
		delay := tpg.GetHumanLikeDelay(baseDelay, domain)

		// Should still be at least baseDelay
		if delay < baseDelay {
			t.Errorf("Expected delay to be at least %v even when disabled, got %v", baseDelay, delay)
		}
	})

	t.Run("with_crawl_delay", func(t *testing.T) {
		tpg := NewTimingPatternGenerator()
		crawlDelay := 5 * time.Second
		delay := tpg.GetHumanLikeDelayWithCrawlDelay(baseDelay, crawlDelay, domain)

		// Should use the maximum of baseDelay and crawlDelay
		if delay < crawlDelay {
			t.Errorf("Expected delay to respect crawl-delay of %v, got %v", crawlDelay, delay)
		}
	})
}

func TestTimingPatternGenerator_IsEnabled(t *testing.T) {
	t.Run("default_enabled", func(t *testing.T) {
		os.Unsetenv("SCRAPING_HUMAN_LIKE_TIMING_ENABLED")
		tpg := NewTimingPatternGenerator()
		if !tpg.IsEnabled() {
			t.Error("Expected human-like timing to be enabled by default")
		}
	})

	t.Run("explicitly_disabled", func(t *testing.T) {
		os.Setenv("SCRAPING_HUMAN_LIKE_TIMING_ENABLED", "false")
		defer os.Unsetenv("SCRAPING_HUMAN_LIKE_TIMING_ENABLED")

		tpg := NewTimingPatternGenerator()
		if tpg.IsEnabled() {
			t.Error("Expected human-like timing to be disabled")
		}
	})
}

