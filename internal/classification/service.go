package classification

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"kyb-platform/internal/classification/repository"
)

// IndustryDetectionService provides database-driven industry classification
type IndustryDetectionService struct {
	repo    repository.KeywordRepository
	logger  *log.Logger
	monitor *ClassificationAccuracyMonitoring
}

// NewIndustryDetectionService creates a new industry detection service
func NewIndustryDetectionService(repo repository.KeywordRepository, logger *log.Logger) *IndustryDetectionService {
	if logger == nil {
		logger = log.Default()
	}

	return &IndustryDetectionService{
		repo:    repo,
		logger:  logger,
		monitor: nil, // Will be set separately if monitoring is needed
	}
}

// NewIndustryDetectionServiceWithMonitoring creates a new industry detection service with monitoring
func NewIndustryDetectionServiceWithMonitoring(repo repository.KeywordRepository, logger *log.Logger, monitor *ClassificationAccuracyMonitoring) *IndustryDetectionService {
	if logger == nil {
		logger = log.Default()
	}

	return &IndustryDetectionService{
		repo:    repo,
		logger:  logger,
		monitor: monitor,
	}
}

// IndustryDetectionResult represents the result of industry detection
type IndustryDetectionResult struct {
	IndustryName   string        `json:"industry_name"`
	Confidence     float64       `json:"confidence"`
	Keywords       []string      `json:"keywords"`
	ProcessingTime time.Duration `json:"processing_time"`
	Method         string        `json:"method"`
	Reasoning      string        `json:"reasoning"`
	CreatedAt      time.Time     `json:"created_at"`
}

// DetectIndustry performs database-driven industry detection
func (s *IndustryDetectionService) DetectIndustry(ctx context.Context, businessName, description, websiteURL string) (*IndustryDetectionResult, error) {
	startTime := time.Now()
	requestID := s.generateRequestID()

	s.logger.Printf("🔍 Starting industry detection for: %s (request: %s)", businessName, requestID)

	// Extract keywords using database-driven approach
	keywords, err := s.extractKeywordsFromDatabase(ctx, businessName, websiteURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract keywords: %w", err)
	}

	if len(keywords) == 0 {
		s.logger.Printf("⚠️ No keywords extracted for: %s", businessName)
		return &IndustryDetectionResult{
			IndustryName:   "General Business",
			Confidence:     0.30,
			Keywords:       []string{},
			ProcessingTime: time.Since(startTime),
			Method:         "database_driven",
			Reasoning:      "No relevant keywords found in database",
			CreatedAt:      time.Now(),
		}, nil
	}

	// Classify using database-driven keyword matching
	result, err := s.classifyByKeywords(ctx, keywords)
	if err != nil {
		return nil, fmt.Errorf("failed to classify by keywords: %w", err)
	}

	result.ProcessingTime = time.Since(startTime)
	result.Method = "database_driven"
	result.CreatedAt = time.Now()

	s.logger.Printf("✅ Industry detection completed: %s (confidence: %.2f%%) (request: %s)",
		result.IndustryName, result.Confidence*100, requestID)

	return result, nil
}

// extractKeywordsFromDatabase extracts keywords using database-driven approach
func (s *IndustryDetectionService) extractKeywordsFromDatabase(ctx context.Context, businessName, websiteURL string) ([]string, error) {
	// Use the repository's classification method to get keywords
	result, err := s.repo.ClassifyBusiness(ctx, businessName, websiteURL)
	if err != nil {
		return nil, fmt.Errorf("failed to classify business: %w", err)
	}

	if result == nil {
		return []string{}, nil
	}

	return result.Keywords, nil
}

// classifyByKeywords performs classification using database-driven keyword matching
func (s *IndustryDetectionService) classifyByKeywords(ctx context.Context, keywords []string) (*IndustryDetectionResult, error) {
	// Use the repository's classification method
	classification, err := s.repo.ClassifyBusinessByKeywords(ctx, keywords)
	if err != nil {
		return nil, fmt.Errorf("database classification failed: %w", err)
	}

	if classification == nil {
		return &IndustryDetectionResult{
			IndustryName: "General Business",
			Confidence:   0.30,
			Keywords:     keywords,
			Reasoning:    "No matching industry found in database",
		}, nil
	}

	return &IndustryDetectionResult{
		IndustryName: classification.Industry.Name,
		Confidence:   classification.Confidence,
		Keywords:     keywords,
		Reasoning:    fmt.Sprintf("Matched %d keywords to %s industry", len(keywords), classification.Industry.Name),
	}, nil
}

// isTechnicalTerm checks if a word is a technical term that should be filtered out
func (s *IndustryDetectionService) isTechnicalTerm(word string) bool {
	// Technical terms that should be filtered out
	technicalTerms := map[string]bool{
		// HTML/CSS/JavaScript terms
		"html": true, "css": true, "javascript": true, "js": true, "jquery": true,
		"bootstrap": true, "react": true, "angular": true, "vue": true, "node": true,
		"php": true, "python": true, "java": true, "csharp": true, "ruby": true,
		"sql": true, "mysql": true, "postgresql": true, "mongodb": true, "redis": true,
		"api": true, "rest": true, "graphql": true, "json": true, "xml": true,
		"http": true, "https": true, "ssl": true, "tls": true, "dns": true,
		"cdn": true, "aws": true, "azure": true, "gcp": true, "docker": true,
		"kubernetes": true, "git": true, "github": true, "gitlab": true,

		// Common web development terms
		"div": true, "span": true, "class": true, "id": true, "src": true,
		"href": true, "alt": true, "title": true, "meta": true, "script": true,
		"style": true, "link": true, "img": true, "button": true, "input": true,
		"form": true, "table": true, "tr": true, "td": true, "th": true,

		// Common programming terms
		"function": true, "variable": true, "array": true, "object": true,
		"string": true, "integer": true, "boolean": true, "null": true,
		"undefined": true, "true": true, "false": true, "if": true,
		"else": true, "for": true, "while": true, "return": true,

		// Common system terms
		"system": true, "server": true, "client": true, "database": true,
		"cache": true, "session": true, "cookie": true, "token": true,
		"auth": true, "login": true, "logout": true, "register": true,
		"admin": true, "user": true, "guest": true, "public": true,
		"private": true, "protected": true, "static": true, "dynamic": true,
	}

	return technicalTerms[strings.ToLower(word)]
}

// generateRequestID generates a unique request ID for tracking
func (s *IndustryDetectionService) generateRequestID() string {
	return fmt.Sprintf("industry_detection_%d", time.Now().UnixNano())
}

// GetIndustryDetectionMetrics returns current industry detection performance metrics
func (s *IndustryDetectionService) GetIndustryDetectionMetrics(ctx context.Context) (*ClassificationAccuracyStats, error) {
	if s.monitor == nil {
		return nil, fmt.Errorf("monitoring not configured")
	}

	// Note: This would call the actual monitoring method when implemented
	// return s.monitor.GetClassificationAccuracyStats(ctx, 24*time.Hour)
	return nil, fmt.Errorf("monitoring not fully implemented")
}

// ValidateIndustryDetectionResult validates that the detection result is consistent
func (s *IndustryDetectionService) ValidateIndustryDetectionResult(result *IndustryDetectionResult) error {
	if result == nil {
		return fmt.Errorf("industry detection result cannot be nil")
	}

	// Validate industry name
	if result.IndustryName == "" {
		return fmt.Errorf("industry name cannot be empty")
	}

	// Validate confidence score
	if result.Confidence < 0.0 || result.Confidence > 1.0 {
		return fmt.Errorf("invalid confidence score: %.2f (must be between 0.0 and 1.0)", result.Confidence)
	}

	// Validate processing time
	if result.ProcessingTime < 0 {
		return fmt.Errorf("invalid processing time: %v (must be non-negative)", result.ProcessingTime)
	}

	// Validate method
	if result.Method == "" {
		return fmt.Errorf("detection method cannot be empty")
	}

	return nil
}

// GetIndustryDetectionStatistics returns statistics about industry detection
func (s *IndustryDetectionService) GetIndustryDetectionStatistics() map[string]interface{} {
	return map[string]interface{}{
		"service_name":       "IndustryDetectionService",
		"version":            "2.0.0",
		"database_driven":    true,
		"hardcoded_patterns": false,
		"monitoring_enabled": s.monitor != nil,
		"created_at":         time.Now(),
	}
}
