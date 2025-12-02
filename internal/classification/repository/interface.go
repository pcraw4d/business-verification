package repository

import (
	"context"
	"time"
)

// Industry represents a business industry with classification metadata
type Industry struct {
	ID                  int     `json:"id"`
	Name                string  `json:"name"`
	Description         string  `json:"description"`
	Category            string  `json:"category"`
	ConfidenceThreshold float64 `json:"confidence_threshold"`
	IsActive            bool    `json:"is_active"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

// IndustryKeyword represents a keyword associated with an industry
type IndustryKeyword struct {
	ID         int     `json:"id"`
	IndustryID int     `json:"industry_id"`
	Keyword    string  `json:"keyword"`
	Weight     float64 `json:"weight"`
	IsActive   bool    `json:"is_active"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// ClassificationCode represents industry classification codes (NAICS, MCC, SIC)
type ClassificationCode struct {
	ID          int    `json:"id"`
	IndustryID  int    `json:"industry_id"`
	CodeType    string `json:"code_type"`
	Code        string `json:"code"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ClassificationCodeWithMetadata represents a classification code with keyword matching metadata
type ClassificationCodeWithMetadata struct {
	ClassificationCode
	RelevanceScore float64 `json:"relevance_score"` // From code_keywords table
	MatchType      string  `json:"match_type"`        // "exact", "partial", "synonym"
}

// IndustryPattern represents phrase patterns for industry detection
type IndustryPattern struct {
	ID              int     `json:"id"`
	IndustryID      int     `json:"industry_id"`
	Pattern         string  `json:"pattern"`
	PatternType     string  `json:"pattern_type"`
	ConfidenceScore float64 `json:"confidence_score"`
	IsActive        bool    `json:"is_active"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// KeywordWeight represents dynamic keyword weighting and scoring
type KeywordWeight struct {
	ID                int     `json:"id"`
	Keyword           string  `json:"keyword"`
	IndustryID        int     `json:"industry_id"`
	BaseWeight        float64 `json:"base_weight"`
	ContextMultiplier float64 `json:"context_multiplier"`
	UsageCount        int     `json:"usage_count"`
	LastUpdated       string  `json:"last_updated"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

// ClassificationResult represents the result of business classification
type ClassificationResult struct {
	Industry   *Industry            `json:"industry"`
	Confidence float64              `json:"confidence"`
	Keywords   []string             `json:"keywords"`
	Patterns   []string             `json:"patterns"`
	Codes      []ClassificationCode `json:"codes"`
	Reasoning  string               `json:"reasoning"`
}

// SearchFilters represents filters for searching industries and keywords
type SearchFilters struct {
	Category      string  `json:"category"`
	MinConfidence float64 `json:"min_confidence"`
	IsActive      *bool   `json:"is_active"`
	Keyword       string  `json:"keyword"`
	Limit         int     `json:"limit"`
	Offset        int     `json:"offset"`
}

// KeywordRepository defines the interface for keyword classification operations
type KeywordRepository interface {
	// Industry Management
	GetIndustryByID(ctx context.Context, id int) (*Industry, error)
	GetIndustryByName(ctx context.Context, name string) (*Industry, error)
	ListIndustries(ctx context.Context, category string) ([]*Industry, error)
	CreateIndustry(ctx context.Context, industry *Industry) error
	UpdateIndustry(ctx context.Context, industry *Industry) error
	DeleteIndustry(ctx context.Context, id int) error

	// Keyword Management
	GetKeywordsByIndustry(ctx context.Context, industryID int) ([]*IndustryKeyword, error)
	SearchKeywords(ctx context.Context, query string, limit int) ([]*IndustryKeyword, error)
	AddKeywordToIndustry(ctx context.Context, industryID int, keyword string, weight float64) error
	UpdateKeywordWeight(ctx context.Context, keywordID int, weight float64) error
	RemoveKeywordFromIndustry(ctx context.Context, keywordID int) error

	// Classification Codes
	GetClassificationCodesByIndustry(ctx context.Context, industryID int) ([]*ClassificationCode, error)
	GetClassificationCodesByType(ctx context.Context, codeType string) ([]*ClassificationCode, error)
	GetClassificationCodesByKeywords(ctx context.Context, keywords []string, codeType string, minRelevance float64) ([]*ClassificationCodeWithMetadata, error)
	FindCodesByFullTextSearch(ctx context.Context, searchText string, codeType string) ([]*ClassificationCode, error)
	AddClassificationCode(ctx context.Context, code *ClassificationCode) error
	UpdateClassificationCode(ctx context.Context, code *ClassificationCode) error
	DeleteClassificationCode(ctx context.Context, id int) error

	// Cached Classification Codes
	GetCachedClassificationCodes(ctx context.Context, industryID int) ([]*ClassificationCode, error)
	GetCachedClassificationCodesByType(ctx context.Context, codeType string) ([]*ClassificationCode, error)
	InitializeIndustryCodeCache(ctx context.Context) error
	InvalidateIndustryCodeCache(ctx context.Context, patterns []string) error
	GetIndustryCodeCacheStats() *IndustryCodeCacheStats

	// Optimized Batch Queries
	GetBatchClassificationCodes(ctx context.Context, industryIDs []int) (map[int][]*ClassificationCode, error)
	GetBatchIndustries(ctx context.Context, industryIDs []int) (map[int]*Industry, error)
	GetBatchKeywords(ctx context.Context, industryIDs []int) (map[int][]*KeywordWeight, error)

	// Industry Patterns (deprecated - not implemented, using keyword-based classification instead)
	// Phase 5.1: Removed from interface - pattern matching not implemented
	// GetPatternsByIndustry(ctx context.Context, industryID int) ([]*IndustryPattern, error)
	// AddPattern(ctx context.Context, pattern *IndustryPattern) error
	// UpdatePattern(ctx context.Context, pattern *IndustryPattern) error
	// DeletePattern(ctx context.Context, id int) error

	// Keyword Weights
	GetKeywordWeights(ctx context.Context, keyword string) ([]*KeywordWeight, error)
	UpdateKeywordWeightByID(ctx context.Context, weight *KeywordWeight) error
	IncrementUsageCount(ctx context.Context, keyword string, industryID int) error

	// Business Classification
	ClassifyBusiness(ctx context.Context, businessName, websiteURL string) (*ClassificationResult, error)
	ClassifyBusinessByKeywords(ctx context.Context, keywords []string) (*ClassificationResult, error)
	GetTopIndustriesByKeywords(ctx context.Context, keywords []string, limit int) ([]*Industry, error)

	// Advanced Search and Analytics
	SearchIndustriesByPattern(ctx context.Context, pattern string) ([]*Industry, error)
	GetIndustryStatistics(ctx context.Context) (map[string]interface{}, error)
	GetKeywordFrequency(ctx context.Context, industryID int) (map[string]int, error)

	// Bulk Operations
	BulkInsertKeywords(ctx context.Context, keywords []*IndustryKeyword) error
	BulkUpdateKeywords(ctx context.Context, keywords []*IndustryKeyword) error
	BulkDeleteKeywords(ctx context.Context, keywordIDs []int) error

	// Health and Maintenance
	Ping(ctx context.Context) error
	GetDatabaseStats(ctx context.Context) (map[string]interface{}, error)
	CleanupInactiveData(ctx context.Context) error

	// Accuracy Tracking and Calibration (OPTIMIZATION #5.2)
	SaveClassificationAccuracy(ctx context.Context, tracking *ClassificationAccuracyTracking) error
	UpdateClassificationAccuracy(ctx context.Context, requestID string, actualIndustry string, validatedBy string) error
	GetCalibrationStatistics(ctx context.Context, startDate, endDate time.Time) ([]*CalibrationBinStatistics, error)

	// Topic-Industry Mapping (Phase 1.3)
	GetIndustryTopicsByKeywords(ctx context.Context, keywords []string) (map[int]float64, error) // industry_id -> relevance_score
	GetTopicAccuracy(ctx context.Context, industryID int, topic string) (float64, error)

	// Keyword Patterns / Co-Occurrence (Phase 1.4)
	FindIndustriesByPatterns(ctx context.Context, patterns []string) ([]*PatternMatchResult, error)
	GetPatternMatches(ctx context.Context, industryID int, patterns []string) ([]*KeywordPattern, error)

	// Batch Queries (Phase 2.2)
	BatchFindKeywords(ctx context.Context, keywords []string) (map[string][]IndustryMatch, error)
	BatchFindIndustryTopics(ctx context.Context, keywords []string) (map[string][]TopicMatch, error)
}

// PatternMatchResult represents an industry match from keyword pattern analysis
type PatternMatchResult struct {
	IndustryID     int
	IndustryName   string
	PatternMatches int
	AvgScore       float64
	MatchedPatterns []string
}

// KeywordPattern represents a keyword co-occurrence pattern
type KeywordPattern struct {
	ID               int
	IndustryID       int
	KeywordPair      string
	Keyword1         string
	Keyword2         string
	CoOccurrenceScore float64
	PatternType      string
	Frequency        int
}

// IndustryMatch represents a keyword match with industry information
type IndustryMatch struct {
	IndustryID     int
	IndustryName   string
	Weight         float64
	Similarity     float64
}

// TopicMatch represents an industry topic match
type TopicMatch struct {
	IndustryID     int
	IndustryName   string
	RelevanceScore float64
	AccuracyScore  float64
}

// ClassificationAccuracyTracking represents a classification accuracy tracking record
type ClassificationAccuracyTracking struct {
	RequestID            string    `json:"request_id"`
	BusinessName         string    `json:"business_name"`
	WebsiteURL           string    `json:"website_url"`
	PredictedIndustry    string    `json:"predicted_industry"`
	ActualIndustry       string    `json:"actual_industry,omitempty"` // NULL until validated
	PredictedConfidence  float64   `json:"predicted_confidence"`
	ActualConfidence     float64   `json:"actual_confidence,omitempty"` // NULL until validated
	IsCorrect            *bool     `json:"is_correct,omitempty"`       // NULL until validated
	ConfidenceBin        int       `json:"confidence_bin"`
	ClassificationMethod string    `json:"classification_method"`
	KeywordsCount        int       `json:"keywords_count"`
	ProcessingTimeMs     int       `json:"processing_time_ms"`
	CreatedAt            time.Time `json:"created_at"`
	ValidatedAt          *time.Time `json:"validated_at,omitempty"`
	ValidatedBy          string    `json:"validated_by,omitempty"`
}

// CalibrationBinStatistics represents calibration statistics for a confidence bin
type CalibrationBinStatistics struct {
	ConfidenceBin        int     `json:"confidence_bin"`
	TotalClassifications int64   `json:"total_classifications"`
	CorrectClassifications int64 `json:"correct_classifications"`
	PredictedAccuracy    float64 `json:"predicted_accuracy"`
	ActualAccuracy       float64 `json:"actual_accuracy"`
	CalibrationError     float64 `json:"calibration_error"`
}
