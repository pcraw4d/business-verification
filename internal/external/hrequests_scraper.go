package external

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// HrequestsScraper implements ScraperStrategy interface for hrequests scraping
type HrequestsScraper struct {
	client *HrequestsClient
	logger *zap.Logger
}

// NewHrequestsScraper creates a new hrequests scraper strategy
func NewHrequestsScraper(client *HrequestsClient, logger *zap.Logger) *HrequestsScraper {
	return &HrequestsScraper{
		client: client,
		logger: logger,
	}
}

// Name returns the name of this scraper strategy
func (s *HrequestsScraper) Name() string {
	return "hrequests"
}

// Scrape attempts to scrape a website using the hrequests service
func (s *HrequestsScraper) Scrape(ctx context.Context, url string) (*ScrapedContent, error) {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		s.logger.Warn("⚠️ [Phase1] [Hrequests] Context cancelled before scrape",
			zap.String("url", url),
			zap.Error(ctx.Err()))
		return nil, ctx.Err()
	default:
		// Context is valid, proceed
	}

	s.logger.Info("🔍 [Phase1] [Hrequests] Starting scrape attempt",
		zap.String("url", url))

	// Call hrequests client
	content, err := s.client.Scrape(ctx, url)
	if err != nil {
		s.logger.Warn("⚠️ [Phase1] [Hrequests] Scrape failed",
			zap.String("url", url),
			zap.Error(err))
		return nil, err
	}

	// Validate content quality
	if !s.isHrequestsContentValid(content) {
		s.logger.Warn("⚠️ [Phase1] [Hrequests] Content validation failed",
			zap.String("url", url),
			zap.Float64("quality_score", content.QualityScore),
			zap.Int("word_count", content.WordCount))
		return nil, fmt.Errorf("hrequests content validation failed: quality_score=%.2f, word_count=%d",
			content.QualityScore, content.WordCount)
	}

	s.logger.Info("✅ [Phase1] [Hrequests] Scrape succeeded",
		zap.String("url", url),
		zap.Float64("quality_score", content.QualityScore),
		zap.Int("word_count", content.WordCount))

	return content, nil
}

// isHrequestsContentValid checks if the scraped content meets minimum quality requirements
func (s *HrequestsScraper) isHrequestsContentValid(content *ScrapedContent) bool {
	if content == nil {
		return false
	}

	// Minimum word count requirement
	if content.WordCount < 50 {
		return false
	}

	// Minimum quality score requirement
	if content.QualityScore < 0.5 {
		return false
	}

	// Should have at least title or meta description
	if content.Title == "" && content.MetaDesc == "" {
		return false
	}

	return true
}



