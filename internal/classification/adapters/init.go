package adapters

import (
	"log"

	"kyb-platform/internal/classification"
	"kyb-platform/internal/classification/repository"
)

// Init initializes the adapters and registers them with the repository package
// This must be called before using repository functions that need adapters
func Init() {
	repository.InitAdapters(
		NewStructuredDataExtractorAdapter,
		NewSmartWebsiteCrawlerAdapter,
	)
}

// NewStructuredDataExtractorAdapter creates an adapter from classification.StructuredDataExtractor
func NewStructuredDataExtractorAdapter(logger *log.Logger) repository.StructuredDataExtractorInterface {
	extractor := classification.NewStructuredDataExtractor(logger)
	return &structuredDataExtractorAdapter{extractor: extractor}
}

// NewSmartWebsiteCrawlerAdapter creates an adapter from classification.SmartWebsiteCrawler
func NewSmartWebsiteCrawlerAdapter(logger *log.Logger) repository.SmartWebsiteCrawlerInterface {
	crawler := classification.NewSmartWebsiteCrawler(logger)
	return &smartWebsiteCrawlerAdapter{crawler: crawler}
}

