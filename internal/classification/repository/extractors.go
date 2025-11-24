package repository

import (
	"context"
	"log"
)

// PageAnalysisData represents the data needed from page analysis
// This interface breaks the import cycle by not importing classification.PageAnalysis directly
type PageAnalysisData interface {
	GetURL() string
	GetStatusCode() int
	GetRelevanceScore() float64
	GetKeywords() []string
	GetIndustryIndicators() []string
	GetStructuredData() map[string]interface{}
}

// StructuredDataExtractorInterface defines the interface for structured data extraction
// This breaks the import cycle by not importing classification.StructuredDataExtractor directly
type StructuredDataExtractorInterface interface {
	ExtractStructuredData(htmlContent string) StructuredDataResult
}

// SmartWebsiteCrawlerInterface defines the interface for website crawling
// This breaks the import cycle by not importing classification.SmartWebsiteCrawler directly
type SmartWebsiteCrawlerInterface interface {
	CrawlWebsite(ctx context.Context, websiteURL string) (CrawlResultInterface, error)
}

// CrawlResultInterface defines the interface for crawl results
type CrawlResultInterface interface {
	GetPagesAnalyzed() []PageAnalysisData
	GetSuccess() bool
	GetError() string
}

// StructuredDataResult represents extracted structured data
type StructuredDataResult struct {
	SchemaOrgData   []SchemaOrgItem
	OpenGraphData   map[string]string
	TwitterCardData map[string]string
	Microdata       []MicrodataItem
	BusinessInfo    BusinessInfoData
	ContactInfo     ContactInfoData
	ProductInfo     []ProductInfoData
	ServiceInfo     []ServiceInfoData
	EventInfo       []EventInfoData
	ExtractionScore float64
}

// SchemaOrgItem represents a Schema.org structured data item
type SchemaOrgItem struct {
	Type       string
	Properties map[string]interface{}
	Context    string
	Confidence float64
}

// MicrodataItem represents microdata
type MicrodataItem struct {
	Type       string
	Properties map[string]interface{}
}

// BusinessInfoData represents business information
type BusinessInfoData struct {
	BusinessName string
	Description  string
	Services     []string
	Products     []string
	Industry     string
	BusinessType string
}

// ContactInfoData represents contact information
type ContactInfoData struct {
	Phone   string
	Email   string
	Address string
	Website string
	Social  map[string]string
}

// ProductInfoData represents product information
type ProductInfoData struct {
	Name        string
	Description string
	Price       string
	Category    string
	Brand       string
	SKU         string
	Image       string
	URL         string
	Confidence  float64
}

// ServiceInfoData represents service information
type ServiceInfoData struct {
	Name        string
	Description string
	Category    string
	Price       string
	Duration    string
	Features    []string
	URL         string
	Confidence  float64
}

// EventInfoData represents event information
type EventInfoData struct {
	Name        string
	Description string
	StartDate   string
	EndDate     string
	Location    string
	URL         string
	Confidence  float64
}

// NewStructuredDataExtractorAdapter creates an adapter from classification.StructuredDataExtractor
// This is set by internal/classification/adapters to avoid import cycle
var NewStructuredDataExtractorAdapter func(logger *log.Logger) StructuredDataExtractorInterface

// NewSmartWebsiteCrawlerAdapter creates an adapter from classification.SmartWebsiteCrawler
// This is set by internal/classification/adapters to avoid import cycle
var NewSmartWebsiteCrawlerAdapter func(logger *log.Logger) SmartWebsiteCrawlerInterface

// InitAdapters initializes the adapter functions - must be called before using adapters
// This is called from a package that can import both classification and adapters
func InitAdapters(
	structuredDataAdapter func(logger *log.Logger) StructuredDataExtractorInterface,
	smartCrawlerAdapter func(logger *log.Logger) SmartWebsiteCrawlerInterface,
) {
	NewStructuredDataExtractorAdapter = structuredDataAdapter
	NewSmartWebsiteCrawlerAdapter = smartCrawlerAdapter
}

