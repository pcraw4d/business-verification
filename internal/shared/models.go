package shared

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Core Business Classification Models
// =============================================================================

// BusinessClassificationRequest represents a unified business classification request
// that can be used across all classification modules
type BusinessClassificationRequest struct {
	ID                 string                 `json:"id"`
	BusinessName       string                 `json:"business_name" validate:"required"`
	BusinessType       string                 `json:"business_type,omitempty"`
	Industry           string                 `json:"industry,omitempty"`
	Description        string                 `json:"description,omitempty"`
	Keywords           []string               `json:"keywords,omitempty"`
	WebsiteURL         string                 `json:"website_url,omitempty"`
	RegistrationNumber string                 `json:"registration_number,omitempty"`
	TaxID              string                 `json:"tax_id,omitempty"`
	Address            string                 `json:"address,omitempty"`
	GeographicRegion   string                 `json:"geographic_region,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	RequestedAt        time.Time              `json:"requested_at"`
}

// BusinessClassificationResponse represents a unified business classification response
// that aggregates results from multiple classification modules
type BusinessClassificationResponse struct {
	ID                    string                   `json:"id"`
	BusinessName          string                   `json:"business_name"`
	DetectedIndustry      string                   `json:"detected_industry,omitempty"`
	Confidence            float64                  `json:"confidence"`
	Classifications       []IndustryClassification `json:"classifications"`
	PrimaryClassification *IndustryClassification  `json:"primary_classification,omitempty"`
	ClassificationCodes   ClassificationCodes      `json:"classification_codes,omitempty"`
	OverallConfidence     float64                  `json:"overall_confidence"`
	ClassificationMethod  string                   `json:"classification_method"`
	ProcessingTime        time.Duration            `json:"processing_time"`
	ModuleResults         map[string]ModuleResult  `json:"module_results,omitempty"`
	RawData               map[string]interface{}   `json:"raw_data,omitempty"`
	CreatedAt             time.Time                `json:"created_at"`
	Timestamp             time.Time                `json:"timestamp"`
	Metadata              map[string]interface{}   `json:"metadata,omitempty"`

	// Enhanced multi-method classification fields
	MethodBreakdown         []ClassificationMethodResult `json:"method_breakdown,omitempty"`
	EnsembleConfidence      float64                      `json:"ensemble_confidence,omitempty"`
	ClassificationReasoning string                       `json:"classification_reasoning,omitempty"`
	QualityMetrics          *ClassificationQuality       `json:"quality_metrics,omitempty"`
}

// IndustryClassification represents a standardized industry classification result
// that can be used across all classification modules
type IndustryClassification struct {
	IndustryCode         string                 `json:"industry_code"`
	IndustryName         string                 `json:"industry_name"`
	PrimaryIndustry      string                 `json:"primary_industry,omitempty"`      // 1. Primary Industry
	ConfidenceScore      float64                `json:"confidence_score"`                 // 1. Confidence Level
	ClassificationMethod string                 `json:"classification_method"`
	Keywords             []string               `json:"keywords,omitempty"`
	Description          string                 `json:"description,omitempty"`
	Evidence             string                 `json:"evidence,omitempty"`
	ProcessingTime       time.Duration          `json:"processing_time,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
	
	// DistilBART enhancement fields
	ContentSummary       string                 `json:"content_summary,omitempty"`       // Website content summary
	Explanation          string                 `json:"explanation,omitempty"`          // 4. Explanation
	ClassificationReasoning string               `json:"classification_reasoning,omitempty"`
	AllIndustryScores    map[string]float64     `json:"all_industry_scores,omitempty"`
	QuantizationEnabled  bool                   `json:"quantization_enabled,omitempty"`
	ModelVersion         string                 `json:"model_version,omitempty"`
	
	// Risk level (5. Risk Level)
	RiskLevel            string                 `json:"risk_level,omitempty"`
	
	// Code distribution (3. Industry code distribution)
	CodeDistribution     *CodeDistribution     `json:"code_distribution,omitempty"`
}

// ModuleResult represents the result from a specific classification module
type ModuleResult struct {
	ModuleID        string                   `json:"module_id"`
	ModuleType      string                   `json:"module_type"`
	Success         bool                     `json:"success"`
	Error           string                   `json:"error,omitempty"`
	Classifications []IndustryClassification `json:"classifications,omitempty"`
	ProcessingTime  time.Duration            `json:"processing_time"`
	Confidence      float64                  `json:"confidence"`
	RawData         map[string]interface{}   `json:"raw_data,omitempty"`
	Metadata        map[string]interface{}   `json:"metadata,omitempty"`
}

// ClassificationMethodResult represents a classification method and its result
type ClassificationMethodResult struct {
	MethodName     string                  `json:"method_name"`
	MethodType     string                  `json:"method_type"` // "keyword", "ml", "description"
	Confidence     float64                 `json:"confidence"`
	ProcessingTime time.Duration           `json:"processing_time"`
	Result         *IndustryClassification `json:"result"`
	Evidence       []string                `json:"evidence"`
	Keywords       []string                `json:"keywords"`
	Error          string                  `json:"error,omitempty"`
	Success        bool                    `json:"success"`
}

// ClassificationQuality represents quality metrics for the classification
type ClassificationQuality struct {
	OverallQuality     float64 `json:"overall_quality"`
	MethodAgreement    float64 `json:"method_agreement"`
	ConfidenceVariance float64 `json:"confidence_variance"`
	EvidenceStrength   float64 `json:"evidence_strength"`
	DataCompleteness   float64 `json:"data_completeness"`
}

// =============================================================================
// Batch Processing Models
// =============================================================================

// BatchClassificationRequest represents a batch classification request
type BatchClassificationRequest struct {
	ID          string                          `json:"id"`
	Requests    []BusinessClassificationRequest `json:"requests"`
	BatchSize   int                             `json:"batch_size,omitempty"`
	Concurrency int                             `json:"concurrency,omitempty"`
	Timeout     time.Duration                   `json:"timeout,omitempty"`
	Metadata    map[string]interface{}          `json:"metadata,omitempty"`
	RequestedAt time.Time                       `json:"requested_at"`
}

// BatchClassificationResponse represents a batch classification response
type BatchClassificationResponse struct {
	ID             string                           `json:"id"`
	Responses      []BusinessClassificationResponse `json:"responses"`
	TotalCount     int                              `json:"total_count"`
	SuccessCount   int                              `json:"success_count"`
	ErrorCount     int                              `json:"error_count"`
	ProcessingTime time.Duration                    `json:"processing_time"`
	Errors         []BatchError                     `json:"errors,omitempty"`
	Metadata       map[string]interface{}           `json:"metadata,omitempty"`
	CompletedAt    time.Time                        `json:"completed_at"`
}

// BatchError represents an error in batch processing
type BatchError struct {
	Index        int    `json:"index"`
	BusinessName string `json:"business_name"`
	Error        string `json:"error"`
	ModuleID     string `json:"module_id,omitempty"`
}

// =============================================================================
// Enhanced Classification Models
// =============================================================================

// EnhancedClassification represents an enhanced classification with all features
type EnhancedClassification struct {
	ID                   uuid.UUID `json:"id"`
	BusinessName         string    `json:"business_name"`
	IndustryCode         string    `json:"industry_code"`
	IndustryName         string    `json:"industry_name"`
	ConfidenceScore      float64   `json:"confidence_score"`
	ClassificationMethod string    `json:"classification_method"`
	Description          string    `json:"description,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`

	// Enhanced fields
	MLModelVersion          *string                `json:"ml_model_version,omitempty"`
	MLConfidenceScore       *float64               `json:"ml_confidence_score,omitempty"`
	CrosswalkMappings       map[string]interface{} `json:"crosswalk_mappings,omitempty"`
	GeographicRegion        *string                `json:"geographic_region,omitempty"`
	RegionConfidenceScore   *float64               `json:"region_confidence_score,omitempty"`
	IndustrySpecificData    map[string]interface{} `json:"industry_specific_data,omitempty"`
	ClassificationAlgorithm *string                `json:"classification_algorithm,omitempty"`
	ValidationRulesApplied  map[string]interface{} `json:"validation_rules_applied,omitempty"`
	ProcessingTimeMS        *int                   `json:"processing_time_ms,omitempty"`
	EnhancedMetadata        map[string]interface{} `json:"enhanced_metadata,omitempty"`
}

// =============================================================================
// ML Classification Models
// =============================================================================

// MLClassificationRequest represents a request for ML-based classification
type MLClassificationRequest struct {
	BusinessName        string                 `json:"business_name"`
	BusinessDescription string                 `json:"business_description,omitempty"`
	Keywords            []string               `json:"keywords,omitempty"`
	WebsiteContent      string                 `json:"website_content,omitempty"`
	IndustryHints       []string               `json:"industry_hints,omitempty"`
	GeographicRegion    string                 `json:"geographic_region,omitempty"`
	BusinessType        string                 `json:"business_type,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// MLClassificationResult represents the result of ML-based classification
type MLClassificationResult struct {
	IndustryCode       string                 `json:"industry_code"`
	IndustryName       string                 `json:"industry_name"`
	ConfidenceScore    float64                `json:"confidence_score"`
	ModelType          ModelType              `json:"model_type"`
	ModelVersion       string                 `json:"model_version"`
	InferenceTime      time.Duration          `json:"inference_time"`
	ModelPredictions   []ModelPrediction      `json:"model_predictions,omitempty"`
	EnsembleScore      float64                `json:"ensemble_score,omitempty"`
	FeatureImportance  map[string]float64     `json:"feature_importance,omitempty"`
	ProcessingMetadata map[string]interface{} `json:"processing_metadata,omitempty"`
}

// ModelPrediction represents a prediction from a single model
type ModelPrediction struct {
	ModelID         string    `json:"model_id"`
	ModelType       ModelType `json:"model_type"`
	IndustryCode    string    `json:"industry_code"`
	IndustryName    string    `json:"industry_name"`
	ConfidenceScore float64   `json:"confidence_score"`
	RawScore        float64   `json:"raw_score"`
}

// ModelType represents the type of ML model
type ModelType string

const (
	ModelTypeBERT        ModelType = "bert"
	ModelTypeEnsemble    ModelType = "ensemble"
	ModelTypeTransformer ModelType = "transformer"
	ModelTypeCustom      ModelType = "custom"
)

// ModelInfo represents information about a loaded model
type ModelInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        ModelType              `json:"type"`
	Version     string                 `json:"version"`
	Status      ModelStatus            `json:"status"`
	LoadedAt    time.Time              `json:"loaded_at"`
	LastUsed    time.Time              `json:"last_used"`
	UsageCount  int64                  `json:"usage_count"`
	Performance *ModelPerformance      `json:"performance,omitempty"`
	Config      *ModelConfig           `json:"config,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ModelStatus represents the current status of a model
type ModelStatus string

const (
	ModelStatusLoading    ModelStatus = "loading"
	ModelStatusReady      ModelStatus = "ready"
	ModelStatusError      ModelStatus = "error"
	ModelStatusUpdating   ModelStatus = "updating"
	ModelStatusDeprecated ModelStatus = "deprecated"
)

// ModelPerformance represents performance metrics for a model
type ModelPerformance struct {
	Accuracy        float64   `json:"accuracy"`
	Precision       float64   `json:"precision"`
	Recall          float64   `json:"recall"`
	F1Score         float64   `json:"f1_score"`
	InferenceTime   float64   `json:"inference_time_ms"`
	Throughput      float64   `json:"throughput_requests_per_sec"`
	MemoryUsage     float64   `json:"memory_usage_mb"`
	LastEvaluated   time.Time `json:"last_evaluated"`
	EvaluationCount int       `json:"evaluation_count"`
}

// ModelConfig represents configuration for a model
type ModelConfig struct {
	ModelType         ModelType              `json:"model_type"`
	MaxSequenceLength int                    `json:"max_sequence_length"`
	BatchSize         int                    `json:"batch_size"`
	LearningRate      float64                `json:"learning_rate"`
	Epochs            int                    `json:"epochs"`
	ValidationSplit   float64                `json:"validation_split"`
	Hyperparameters   map[string]interface{} `json:"hyperparameters,omitempty"`
}

// =============================================================================
// Website Analysis Models
// =============================================================================

// WebsiteAnalysisRequest represents a website analysis request
type WebsiteAnalysisRequest struct {
	BusinessName          string                 `json:"business_name"`
	WebsiteURL            string                 `json:"website_url"`
	MaxPages              int                    `json:"max_pages,omitempty"`
	IncludeMeta           bool                   `json:"include_meta,omitempty"`
	IncludeStructuredData bool                   `json:"include_structured_data,omitempty"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
}

// WebsiteAnalysisResult represents comprehensive website analysis results
type WebsiteAnalysisResult struct {
	WebsiteURL             string                         `json:"website_url"`
	BusinessName           string                         `json:"business_name"`
	ConnectionValidation   *ConnectionValidationResult    `json:"connection_validation,omitempty"`
	ContentAnalysis        *ContentAnalysisResult         `json:"content_analysis,omitempty"`
	SemanticAnalysis       *SemanticAnalysisResult        `json:"semantic_analysis,omitempty"`
	IndustryClassification []IndustryClassificationResult `json:"industry_classification,omitempty"`
	PageAnalysis           []PageAnalysisResult           `json:"page_analysis,omitempty"`
	OverallConfidence      float64                        `json:"overall_confidence"`
	AnalysisTime           time.Time                      `json:"analysis_time"`
	AnalysisMetadata       map[string]interface{}         `json:"analysis_metadata,omitempty"`
}

// ConnectionValidationResult represents connection validation results
type ConnectionValidationResult struct {
	IsValid          bool     `json:"is_valid"`
	Confidence       float64  `json:"confidence"`
	ValidationMethod string   `json:"validation_method"`
	BusinessMatch    bool     `json:"business_match"`
	DomainAge        int      `json:"domain_age"`
	SSLValid         bool     `json:"ssl_valid"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
}

// ContentAnalysisResult represents content analysis results
type ContentAnalysisResult struct {
	ContentQuality     float64                `json:"content_quality"`
	ContentLength      int                    `json:"content_length"`
	MetaTags           map[string]string      `json:"meta_tags,omitempty"`
	StructuredData     map[string]interface{} `json:"structured_data,omitempty"`
	IndustryIndicators []string               `json:"industry_indicators,omitempty"`
	BusinessKeywords   []string               `json:"business_keywords,omitempty"`
	ContentType        string                 `json:"content_type"`
}

// SemanticAnalysisResult represents semantic analysis results
type SemanticAnalysisResult struct {
	SemanticScore    float64            `json:"semantic_score"`
	TopicModeling    map[string]float64 `json:"topic_modeling,omitempty"`
	SentimentScore   float64            `json:"sentiment_score"`
	KeyPhrases       []string           `json:"key_phrases,omitempty"`
	EntityExtraction map[string]string  `json:"entity_extraction,omitempty"`
}

// IndustryClassificationResult represents industry classification results
type IndustryClassificationResult struct {
	IndustryCode string   `json:"industry_code"`
	IndustryName string   `json:"industry_name"`
	Confidence   float64  `json:"confidence"`
	Keywords     []string `json:"keywords,omitempty"`
	Evidence     string   `json:"evidence,omitempty"`
}

// PageAnalysisResult represents page analysis results
type PageAnalysisResult struct {
	URL            string  `json:"url"`
	PageType       string  `json:"page_type"`
	ContentQuality float64 `json:"content_quality"`
	Relevance      float64 `json:"relevance"`
	Priority       int     `json:"priority"`
}

// =============================================================================
// Web Search Analysis Models
// =============================================================================

// WebSearchAnalysisRequest represents a web search analysis request
type WebSearchAnalysisRequest struct {
	BusinessName  string                 `json:"business_name"`
	SearchQuery   string                 `json:"search_query,omitempty"`
	BusinessType  string                 `json:"business_type,omitempty"`
	Industry      string                 `json:"industry,omitempty"`
	Address       string                 `json:"address,omitempty"`
	MaxResults    int                    `json:"max_results,omitempty"`
	SearchEngines []string               `json:"search_engines,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// WebSearchAnalysisResult represents comprehensive web search analysis results
type WebSearchAnalysisResult struct {
	SearchQuery            string                         `json:"search_query"`
	BusinessName           string                         `json:"business_name"`
	SearchResults          []SearchResult                 `json:"search_results,omitempty"`
	AnalysisResults        *SearchAnalysisResults         `json:"analysis_results,omitempty"`
	IndustryClassification []IndustryClassificationResult `json:"industry_classification,omitempty"`
	BusinessExtraction     *BusinessExtractionResult      `json:"business_extraction,omitempty"`
	OverallConfidence      float64                        `json:"overall_confidence"`
	SearchTime             time.Time                      `json:"search_time"`
	AnalysisMetadata       map[string]interface{}         `json:"analysis_metadata,omitempty"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Title          string            `json:"title"`
	URL            string            `json:"url"`
	Description    string            `json:"description"`
	Content        string            `json:"content,omitempty"`
	RelevanceScore float64           `json:"relevance_score"`
	Rank           int               `json:"rank"`
	Source         string            `json:"source"`
	PublishedDate  *time.Time        `json:"published_date,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// SearchAnalysisResults represents analysis of search results
type SearchAnalysisResults struct {
	TotalResults       int            `json:"total_results"`
	FilteredResults    int            `json:"filtered_results"`
	AverageRelevance   float64        `json:"average_relevance"`
	TopKeywords        []string       `json:"top_keywords,omitempty"`
	SpamDetected       int            `json:"spam_detected"`
	DuplicatesRemoved  int            `json:"duplicates_removed"`
	ContentQuality     float64        `json:"content_quality"`
	SourceDistribution map[string]int `json:"source_distribution,omitempty"`
}

// BusinessExtractionResult represents extracted business information
type BusinessExtractionResult struct {
	BusinessName    string            `json:"business_name"`
	WebsiteURL      string            `json:"website_url,omitempty"`
	PhoneNumber     string            `json:"phone_number,omitempty"`
	EmailAddress    string            `json:"email_address,omitempty"`
	Address         string            `json:"address,omitempty"`
	SocialMedia     map[string]string `json:"social_media,omitempty"`
	Confidence      float64           `json:"confidence"`
	ExtractedFields map[string]string `json:"extracted_fields,omitempty"`
}

// =============================================================================
// Feedback and Validation Models
// =============================================================================

// FeedbackModel represents feedback data model
type FeedbackModel struct {
	ID                        uuid.UUID              `json:"id"`
	UserID                    string                 `json:"user_id"`
	BusinessName              string                 `json:"business_name"`
	OriginalClassificationID  *uuid.UUID             `json:"original_classification_id,omitempty"`
	FeedbackType              string                 `json:"feedback_type"`
	FeedbackValue             map[string]interface{} `json:"feedback_value,omitempty"`
	FeedbackText              *string                `json:"feedback_text,omitempty"`
	SuggestedClassificationID *uuid.UUID             `json:"suggested_classification_id,omitempty"`
	ConfidenceScore           *float64               `json:"confidence_score,omitempty"`
	Status                    string                 `json:"status"`
	ProcessingTimeMS          *int                   `json:"processing_time_ms,omitempty"`
	CreatedAt                 time.Time              `json:"created_at"`
	ProcessedAt               *time.Time             `json:"processed_at,omitempty"`
	Metadata                  map[string]interface{} `json:"metadata,omitempty"`
}

// AccuracyValidationModel represents accuracy validation data model
type AccuracyValidationModel struct {
	ID                       uuid.UUID              `json:"id"`
	ClassificationID         *uuid.UUID             `json:"classification_id,omitempty"`
	MetricType               string                 `json:"metric_type"`
	Dimension                string                 `json:"dimension"`
	TotalClassifications     int                    `json:"total_classifications"`
	CorrectClassifications   int                    `json:"correct_classifications"`
	IncorrectClassifications int                    `json:"incorrect_classifications"`
	AccuracyScore            *float64               `json:"accuracy_score,omitempty"`
	ConfidenceScore          *float64               `json:"confidence_score,omitempty"`
	ProcessingTimeMS         *int                   `json:"processing_time_ms,omitempty"`
	TimeRangeSeconds         *int                   `json:"time_range_seconds,omitempty"`
	CreatedAt                time.Time              `json:"created_at"`
	Metadata                 map[string]interface{} `json:"metadata,omitempty"`
}

// =============================================================================
// Common Types and Enums
// =============================================================================

// ClassificationMethod represents the method used for classification
type ClassificationMethod string

const (
	ClassificationMethodKeyword   ClassificationMethod = "keyword"
	ClassificationMethodML        ClassificationMethod = "ml"
	ClassificationMethodWebsite   ClassificationMethod = "website"
	ClassificationMethodWebSearch ClassificationMethod = "web_search"
	ClassificationMethodEnsemble  ClassificationMethod = "ensemble"
	ClassificationMethodHybrid    ClassificationMethod = "hybrid"
)

// IndustryType represents different industry categories
type IndustryType string

const (
	IndustryTypeAgriculture   IndustryType = "agriculture"
	IndustryTypeRetail        IndustryType = "retail"
	IndustryTypeFood          IndustryType = "food"
	IndustryTypeManufacturing IndustryType = "manufacturing"
	IndustryTypeTechnology    IndustryType = "technology"
	IndustryTypeFinance       IndustryType = "finance"
	IndustryTypeHealthcare    IndustryType = "healthcare"
	IndustryTypeOther         IndustryType = "other"
)

// ConfidenceLevel represents confidence level categories
type ConfidenceLevel string

const (
	ConfidenceLevelHigh   ConfidenceLevel = "high"
	ConfidenceLevelMedium ConfidenceLevel = "medium"
	ConfidenceLevelLow    ConfidenceLevel = "low"
)

// ProcessingStatus represents the status of processing
type ProcessingStatus string

const (
	ProcessingStatusPending   ProcessingStatus = "pending"
	ProcessingStatusRunning   ProcessingStatus = "running"
	ProcessingStatusCompleted ProcessingStatus = "completed"
	ProcessingStatusFailed    ProcessingStatus = "failed"
	ProcessingStatusCancelled ProcessingStatus = "cancelled"
)

// =============================================================================
// Utility Functions
// =============================================================================

// GetConfidenceLevel returns the confidence level based on score
func GetConfidenceLevel(score float64) ConfidenceLevel {
	switch {
	case score >= 0.8:
		return ConfidenceLevelHigh
	case score >= 0.5:
		return ConfidenceLevelMedium
	default:
		return ConfidenceLevelLow
	}
}

// IsValidClassificationMethod checks if a classification method is valid
func IsValidClassificationMethod(method ClassificationMethod) bool {
	validMethods := []ClassificationMethod{
		ClassificationMethodKeyword,
		ClassificationMethodML,
		ClassificationMethodWebsite,
		ClassificationMethodWebSearch,
		ClassificationMethodEnsemble,
		ClassificationMethodHybrid,
	}

	for _, valid := range validMethods {
		if method == valid {
			return true
		}
	}
	return false
}

// IsValidIndustryType checks if an industry type is valid
func IsValidIndustryType(industryType IndustryType) bool {
	validTypes := []IndustryType{
		IndustryTypeAgriculture,
		IndustryTypeRetail,
		IndustryTypeFood,
		IndustryTypeManufacturing,
		IndustryTypeTechnology,
		IndustryTypeFinance,
		IndustryTypeHealthcare,
		IndustryTypeOther,
	}

	for _, valid := range validTypes {
		if industryType == valid {
			return true
		}
	}
	return false
}

// =============================================================================
// Classification Code Types
// =============================================================================

// MCCCode represents a Merchant Category Code
type MCCCode struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// SICCode represents a Standard Industrial Classification code
type SICCode struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// NAICSCode represents a North American Industry Classification System code
type NAICSCode struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// ClassificationCodes represents all classification codes for a business
type ClassificationCodes struct {
	MCC   []MCCCode   `json:"mcc,omitempty"`
	SIC   []SICCode   `json:"sic,omitempty"`
	NAICS []NAICSCode `json:"naics,omitempty"`
}

// GetTopMCC returns top N MCC codes sorted by confidence (default 3)
func (cc *ClassificationCodes) GetTopMCC(limit int) []MCCCode {
	if limit <= 0 {
		limit = 3
	}
	if len(cc.MCC) == 0 {
		return []MCCCode{}
	}
	
	// Sort by confidence (descending)
	sorted := make([]MCCCode, len(cc.MCC))
	copy(sorted, cc.MCC)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Confidence > sorted[j].Confidence
	})
	
	if len(sorted) > limit {
		return sorted[:limit]
	}
	return sorted
}

// GetTopSIC returns top N SIC codes sorted by confidence (default 3)
func (cc *ClassificationCodes) GetTopSIC(limit int) []SICCode {
	if limit <= 0 {
		limit = 3
	}
	if len(cc.SIC) == 0 {
		return []SICCode{}
	}
	
	// Sort by confidence (descending)
	sorted := make([]SICCode, len(cc.SIC))
	copy(sorted, cc.SIC)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Confidence > sorted[j].Confidence
	})
	
	if len(sorted) > limit {
		return sorted[:limit]
	}
	return sorted
}

// GetTopNAICS returns top N NAICS codes sorted by confidence (default 3)
func (cc *ClassificationCodes) GetTopNAICS(limit int) []NAICSCode {
	if limit <= 0 {
		limit = 3
	}
	if len(cc.NAICS) == 0 {
		return []NAICSCode{}
	}
	
	// Sort by confidence (descending)
	sorted := make([]NAICSCode, len(cc.NAICS))
	copy(sorted, cc.NAICS)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Confidence > sorted[j].Confidence
	})
	
	if len(sorted) > limit {
		return sorted[:limit]
	}
	return sorted
}

// CodeDistribution represents the distribution of codes across types
type CodeDistribution struct {
	MCC    CodeDistributionStats `json:"mcc,omitempty"`
	SIC    CodeDistributionStats `json:"sic,omitempty"`
	NAICS  CodeDistributionStats `json:"naics,omitempty"`
	Total  int                   `json:"total"`
}

// CodeDistributionStats provides statistics for a code type
type CodeDistributionStats struct {
	Count            int                `json:"count"`
	TopCodes         []CodeWithConfidence `json:"top_codes,omitempty"`
	AverageConfidence float64            `json:"average_confidence,omitempty"`
}

// CodeWithConfidence represents a code with its confidence
type CodeWithConfidence struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// CalculateCodeDistribution calculates distribution statistics
func (cc *ClassificationCodes) CalculateCodeDistribution() CodeDistribution {
	dist := CodeDistribution{
		MCC: CodeDistributionStats{
			Count:    len(cc.MCC),
			TopCodes: make([]CodeWithConfidence, 0),
		},
		SIC: CodeDistributionStats{
			Count:    len(cc.SIC),
			TopCodes: make([]CodeWithConfidence, 0),
		},
		NAICS: CodeDistributionStats{
			Count:    len(cc.NAICS),
			TopCodes: make([]CodeWithConfidence, 0),
		},
	}
	
	// Calculate MCC distribution
	if len(cc.MCC) > 0 {
		// Sort by confidence (descending) before taking top 3
		sortedMCC := make([]MCCCode, len(cc.MCC))
		copy(sortedMCC, cc.MCC)
		sort.Slice(sortedMCC, func(i, j int) bool {
			return sortedMCC[i].Confidence > sortedMCC[j].Confidence
		})
		
		totalConfidence := 0.0
		topCount := 3
		if len(sortedMCC) < topCount {
			topCount = len(sortedMCC)
		}
		
		for i := 0; i < topCount; i++ {
			dist.MCC.TopCodes = append(dist.MCC.TopCodes, CodeWithConfidence{
				Code:        sortedMCC[i].Code,
				Description: sortedMCC[i].Description,
				Confidence:  sortedMCC[i].Confidence,
			})
		}
		
		for _, code := range cc.MCC {
			totalConfidence += code.Confidence
		}
		dist.MCC.AverageConfidence = totalConfidence / float64(len(cc.MCC))
	}
	
	// Calculate SIC distribution
	if len(cc.SIC) > 0 {
		// Sort by confidence (descending) before taking top 3
		sortedSIC := make([]SICCode, len(cc.SIC))
		copy(sortedSIC, cc.SIC)
		sort.Slice(sortedSIC, func(i, j int) bool {
			return sortedSIC[i].Confidence > sortedSIC[j].Confidence
		})
		
		totalConfidence := 0.0
		topCount := 3
		if len(sortedSIC) < topCount {
			topCount = len(sortedSIC)
		}
		
		for i := 0; i < topCount; i++ {
			dist.SIC.TopCodes = append(dist.SIC.TopCodes, CodeWithConfidence{
				Code:        sortedSIC[i].Code,
				Description: sortedSIC[i].Description,
				Confidence:  sortedSIC[i].Confidence,
			})
		}
		
		for _, code := range cc.SIC {
			totalConfidence += code.Confidence
		}
		dist.SIC.AverageConfidence = totalConfidence / float64(len(cc.SIC))
	}
	
	// Calculate NAICS distribution
	if len(cc.NAICS) > 0 {
		// Sort by confidence (descending) before taking top 3
		sortedNAICS := make([]NAICSCode, len(cc.NAICS))
		copy(sortedNAICS, cc.NAICS)
		sort.Slice(sortedNAICS, func(i, j int) bool {
			return sortedNAICS[i].Confidence > sortedNAICS[j].Confidence
		})
		
		totalConfidence := 0.0
		topCount := 3
		if len(sortedNAICS) < topCount {
			topCount = len(sortedNAICS)
		}
		
		for i := 0; i < topCount; i++ {
			dist.NAICS.TopCodes = append(dist.NAICS.TopCodes, CodeWithConfidence{
				Code:        sortedNAICS[i].Code,
				Description: sortedNAICS[i].Description,
				Confidence:  sortedNAICS[i].Confidence,
			})
		}
		
		for _, code := range cc.NAICS {
			totalConfidence += code.Confidence
		}
		dist.NAICS.AverageConfidence = totalConfidence / float64(len(cc.NAICS))
	}
	
	dist.Total = dist.MCC.Count + dist.SIC.Count + dist.NAICS.Count
	
	return dist
}

// =============================================================================
// Conversion Functions
// =============================================================================

// ConvertModuleDataToBusinessClassificationRequest converts module request data to business classification request
func ConvertModuleDataToBusinessClassificationRequest(data map[string]interface{}) (*BusinessClassificationRequest, error) {
	req := &BusinessClassificationRequest{}

	if businessName, ok := data["business_name"].(string); ok {
		req.BusinessName = businessName
	}

	if businessType, ok := data["business_type"].(string); ok {
		req.BusinessType = businessType
	}

	if industry, ok := data["industry"].(string); ok {
		req.Industry = industry
	}

	if description, ok := data["description"].(string); ok {
		req.Description = description
	}

	if keywords, ok := data["keywords"].([]string); ok {
		req.Keywords = keywords
	}

	if websiteURL, ok := data["website_url"].(string); ok {
		req.WebsiteURL = websiteURL
	}

	if registrationNumber, ok := data["registration_number"].(string); ok {
		req.RegistrationNumber = registrationNumber
	}

	if taxID, ok := data["tax_id"].(string); ok {
		req.TaxID = taxID
	}

	if address, ok := data["address"].(string); ok {
		req.Address = address
	}

	if geographicRegion, ok := data["geographic_region"].(string); ok {
		req.GeographicRegion = geographicRegion
	}

	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		req.Metadata = metadata
	}

	req.RequestedAt = time.Now()

	return req, nil
}

// ConvertBusinessClassificationResponseToModuleData converts business classification response to module response data
func ConvertBusinessClassificationResponseToModuleData(response *BusinessClassificationResponse) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"id":                     response.ID,
		"business_name":          response.BusinessName,
		"detected_industry":      response.DetectedIndustry,
		"confidence":             response.Confidence,
		"classifications":        response.Classifications,
		"primary_classification": response.PrimaryClassification,
		"classification_codes":   response.ClassificationCodes,
		"overall_confidence":     response.OverallConfidence,
		"classification_method":  response.ClassificationMethod,
		"processing_time":        response.ProcessingTime,
		"module_results":         response.ModuleResults,
		"raw_data":               response.RawData,
		"created_at":             response.CreatedAt,
		"timestamp":              response.Timestamp,
		"metadata":               response.Metadata,
	}

	return data, nil
}
