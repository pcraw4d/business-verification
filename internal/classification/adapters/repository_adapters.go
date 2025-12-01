package adapters

import (
	"context"
	"time"

	"kyb-platform/internal/classification"
	"kyb-platform/internal/classification/repository"
)

// pageAnalysisAdapter adapts classification.PageAnalysis to repository.PageAnalysisData
type pageAnalysisAdapter struct {
	analysis classification.PageAnalysis
}

func (p *pageAnalysisAdapter) GetURL() string {
	return p.analysis.URL
}

func (p *pageAnalysisAdapter) GetStatusCode() int {
	return p.analysis.StatusCode
}

func (p *pageAnalysisAdapter) GetRelevanceScore() float64 {
	return p.analysis.RelevanceScore
}

func (p *pageAnalysisAdapter) GetKeywords() []string {
	return p.analysis.Keywords
}

func (p *pageAnalysisAdapter) GetIndustryIndicators() []string {
	return p.analysis.IndustryIndicators
}

func (p *pageAnalysisAdapter) GetStructuredData() map[string]interface{} {
	return p.analysis.StructuredData
}

// crawlResultAdapter adapts classification.CrawlResult to repository.CrawlResultInterface
type crawlResultAdapter struct {
	result *classification.CrawlResult
}

func (c *crawlResultAdapter) GetPagesAnalyzed() []repository.PageAnalysisData {
	adapters := make([]repository.PageAnalysisData, len(c.result.PagesAnalyzed))
	for i, page := range c.result.PagesAnalyzed {
		adapters[i] = &pageAnalysisAdapter{analysis: page}
	}
	return adapters
}

func (c *crawlResultAdapter) GetSuccess() bool {
	return c.result.Success
}

func (c *crawlResultAdapter) GetError() string {
	return c.result.Error
}

// structuredDataExtractorAdapter adapts classification.StructuredDataExtractor to repository.StructuredDataExtractorInterface
type structuredDataExtractorAdapter struct {
	extractor *classification.StructuredDataExtractor
}

func (s *structuredDataExtractorAdapter) ExtractStructuredData(htmlContent string) repository.StructuredDataResult {
	result := s.extractor.ExtractStructuredData(htmlContent)
	
	// Convert to repository interface type
	return repository.StructuredDataResult{
		SchemaOrgData:   convertSchemaOrgData(result.SchemaOrgData),
		OpenGraphData:   result.OpenGraphData,
		TwitterCardData: result.TwitterCardData,
		Microdata:       convertMicrodata(result.Microdata),
		BusinessInfo: repository.BusinessInfoData{
			BusinessName: result.BusinessInfo.BusinessName,
			Description:  result.BusinessInfo.Description,
			Services:     result.BusinessInfo.Services,
			Products:     result.BusinessInfo.Products,
			Industry:     result.BusinessInfo.Industry,
			BusinessType: result.BusinessInfo.BusinessType,
		},
		ContactInfo: repository.ContactInfoData{
			Phone:   result.ContactInfo.Phone,
			Email:   result.ContactInfo.Email,
			Address: result.ContactInfo.Address,
			Website: result.ContactInfo.Website,
			Social:  result.ContactInfo.Social,
		},
		ProductInfo:     convertProductInfo(result.ProductInfo),
		ServiceInfo:     convertServiceInfo(result.ServiceInfo),
		EventInfo:       convertEventInfo(result.EventInfo),
		ExtractionScore: result.ExtractionScore,
	}
}

func convertSchemaOrgData(data []classification.SchemaOrgItem) []repository.SchemaOrgItem {
	if data == nil {
		return nil
	}
	result := make([]repository.SchemaOrgItem, len(data))
	for i, item := range data {
		result[i] = repository.SchemaOrgItem{
			Type:       item.Type,
			Properties: item.Properties,
			Context:    item.Context,
			Confidence: item.Confidence,
		}
	}
	return result
}

func convertMicrodata(data []classification.MicrodataItem) []repository.MicrodataItem {
	if data == nil {
		return nil
	}
	result := make([]repository.MicrodataItem, len(data))
	for i, item := range data {
		result[i] = repository.MicrodataItem{
			Type:       item.Type,
			Properties: item.Properties,
		}
	}
	return result
}

func convertProductInfo(data []classification.ProductInfo) []repository.ProductInfoData {
	if data == nil {
		return nil
	}
	result := make([]repository.ProductInfoData, len(data))
	for i, item := range data {
		result[i] = repository.ProductInfoData{
			Name:        item.Name,
			Description: item.Description,
			Price:       item.Price,
			Category:    item.Category,
			Brand:       item.Brand,
			SKU:         item.SKU,
			Image:       item.Image,
			URL:         item.URL,
			Confidence:  item.Confidence,
		}
	}
	return result
}

func convertServiceInfo(data []classification.ServiceInfo) []repository.ServiceInfoData {
	if data == nil {
		return nil
	}
	result := make([]repository.ServiceInfoData, len(data))
	for i, item := range data {
		result[i] = repository.ServiceInfoData{
			Name:        item.Name,
			Description: item.Description,
			Category:    item.Category,
			Price:       item.Price,
			Duration:    item.Duration,
			Features:    item.Features,
			URL:         item.URL,
			Confidence:  item.Confidence,
		}
	}
	return result
}

func convertEventInfo(data []classification.EventInfo) []repository.EventInfoData {
	if data == nil {
		return nil
	}
	result := make([]repository.EventInfoData, len(data))
	for i, item := range data {
		result[i] = repository.EventInfoData{
			Name:        item.Name,
			Description: item.Description,
			StartDate:   item.StartDate,
			EndDate:     item.EndDate,
			Location:    item.Location,
			URL:         item.URL,
			Confidence:  item.Confidence,
		}
	}
	return result
}

// smartWebsiteCrawlerAdapter adapts classification.SmartWebsiteCrawler to repository.SmartWebsiteCrawlerInterface
type smartWebsiteCrawlerAdapter struct {
	crawler *classification.SmartWebsiteCrawler
}

func (s *smartWebsiteCrawlerAdapter) CrawlWebsite(ctx context.Context, websiteURL string) (repository.CrawlResultInterface, error) {
	result, err := s.crawler.CrawlWebsite(ctx, websiteURL)
	if err != nil {
		return nil, err
	}
	return &crawlResultAdapter{result: result}, nil
}

func (s *smartWebsiteCrawlerAdapter) CrawlWebsiteFast(ctx context.Context, websiteURL string, maxTime time.Duration, maxPages int, maxConcurrent int) (repository.CrawlResultInterface, error) {
	result, err := s.crawler.CrawlWebsiteFast(ctx, websiteURL, maxTime, maxPages, maxConcurrent)
	if err != nil {
		return nil, err
	}
	return &crawlResultAdapter{result: result}, nil
}


