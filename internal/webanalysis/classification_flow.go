package webanalysis

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ClassificationFlow represents the type of classification flow
type ClassificationFlow string

const (
	FlowURLBased    ClassificationFlow = "url_based"
	FlowSearchBased ClassificationFlow = "search_based"
)

// ClassificationRequest represents a business classification request
type ClassificationRequest struct {
	BusinessName                string             `json:"business_name"`
	BusinessType                string             `json:"business_type,omitempty"`
	Industry                    string             `json:"industry,omitempty"`
	WebsiteURL                  string             `json:"website_url,omitempty"`
	Address                     string             `json:"address,omitempty"`
	ContactInfo                 map[string]string  `json:"contact_info,omitempty"`
	FlowPreference              ClassificationFlow `json:"flow_preference,omitempty"`
	MaxResults                  int                `json:"max_results,omitempty"`
	ConfidenceThreshold         float64            `json:"confidence_threshold,omitempty"`
	IncludeRiskAnalysis         bool               `json:"include_risk_analysis,omitempty"`
	IncludeConnectionValidation bool               `json:"include_connection_validation,omitempty"`
}

// ClassificationResult represents the unified result format
type ClassificationResult struct {
	RequestID            string                   `json:"request_id"`
	BusinessName         string                   `json:"business_name"`
	FlowUsed             ClassificationFlow       `json:"flow_used"`
	ProcessingTime       time.Duration            `json:"processing_time"`
	Confidence           float64                  `json:"confidence"`
	Industries           []IndustryClassification `json:"industries"`
	RiskAssessment       *RiskAssessment          `json:"risk_assessment,omitempty"`
	ConnectionValidation *ConnectionValidation    `json:"connection_validation,omitempty"`
	WebsiteData          *WebsiteAnalysis         `json:"website_data,omitempty"`
	SearchData           *SearchAnalysis          `json:"search_data,omitempty"`
	Errors               []string                 `json:"errors,omitempty"`
	Warnings             []string                 `json:"warnings,omitempty"`
}

// IndustryClassification represents a single industry classification
type IndustryClassification struct {
	Industry   string   `json:"industry"`
	NAICSCode  string   `json:"naics_code,omitempty"`
	SICCode    string   `json:"sic_code,omitempty"`
	Confidence float64  `json:"confidence"`
	Evidence   string   `json:"evidence,omitempty"`
	Keywords   []string `json:"keywords,omitempty"`
}

// RiskAssessment represents risk analysis results
type RiskAssessment struct {
	OverallRisk     string          `json:"overall_risk"`
	RiskScore       float64         `json:"risk_score"`
	RiskFactors     []RiskFactor    `json:"risk_factors"`
	RiskIndicators  []RiskIndicator `json:"risk_indicators"`
	Recommendations []string        `json:"recommendations"`
}

// RiskFactor represents a specific risk factor
type RiskFactor struct {
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
	Confidence  float64 `json:"confidence"`
	Evidence    string  `json:"evidence"`
}

// RiskIndicator represents a risk indicator
type RiskIndicator struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	Source      string  `json:"source"`
}

// ConnectionValidation represents business-website connection validation
type ConnectionValidation struct {
	IsConnected       bool               `json:"is_connected"`
	Confidence        float64            `json:"confidence"`
	Evidence          string             `json:"evidence"`
	ValidationFactors []ValidationFactor `json:"validation_factors"`
	Recommendations   []string           `json:"recommendations"`
}

// ValidationFactor represents a validation factor
type ValidationFactor struct {
	Factor     string  `json:"factor"`
	Match      bool    `json:"match"`
	Confidence float64 `json:"confidence"`
	Details    string  `json:"details"`
}

// WebsiteAnalysis represents website scraping results
type WebsiteAnalysis struct {
	URL           string            `json:"url"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Content       string            `json:"content"`
	ExtractedData map[string]string `json:"extracted_data"`
	PageCount     int               `json:"page_count"`
	ScrapingDepth int               `json:"scraping_depth"`
	QualityScore  float64           `json:"quality_score"`
}

// SearchAnalysis represents web search results
type SearchAnalysis struct {
	SearchQuery  string         `json:"search_query"`
	ResultsCount int            `json:"results_count"`
	TopResults   []SearchResult `json:"top_results"`
	SearchTime   time.Duration  `json:"search_time"`
	SourcesUsed  []string       `json:"sources_used"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Title          string  `json:"title"`
	URL            string  `json:"url"`
	Description    string  `json:"description"`
	RelevanceScore float64 `json:"relevance_score"`
	Source         string  `json:"source"`
}

// ClassificationFlowManager manages the dual-classification flow
type ClassificationFlowManager struct {
	urlFlow          *URLBasedFlow
	searchFlow       *SearchBasedFlow
	flowSelector     *FlowSelector
	resultAggregator *ResultAggregator
	mu               sync.RWMutex
	config           FlowConfig
}

// FlowConfig holds configuration for classification flows
type FlowConfig struct {
	DefaultMaxResults          int           `json:"default_max_results"`
	DefaultConfidenceThreshold float64       `json:"default_confidence_threshold"`
	URLFlowTimeout             time.Duration `json:"url_flow_timeout"`
	SearchFlowTimeout          time.Duration `json:"search_flow_timeout"`
	FallbackEnabled            bool          `json:"fallback_enabled"`
	ParallelProcessing         bool          `json:"parallel_processing"`
	RetryAttempts              int           `json:"retry_attempts"`
}

// NewClassificationFlowManager creates a new classification flow manager
func NewClassificationFlowManager(config FlowConfig) *ClassificationFlowManager {
	return &ClassificationFlowManager{
		urlFlow:          NewURLBasedFlow(),
		searchFlow:       NewSearchBasedFlow(),
		flowSelector:     NewFlowSelector(),
		resultAggregator: NewResultAggregator(),
		config:           config,
	}
}

// ClassifyBusiness performs business classification using the appropriate flow
func (cfm *ClassificationFlowManager) ClassifyBusiness(ctx context.Context, req *ClassificationRequest) (*ClassificationResult, error) {
	start := time.Now()

	// Set default values
	if req.MaxResults == 0 {
		req.MaxResults = cfm.config.DefaultMaxResults
	}
	if req.ConfidenceThreshold == 0 {
		req.ConfidenceThreshold = cfm.config.DefaultConfidenceThreshold
	}

	// Generate request ID
	requestID := generateRequestID()

	// Determine which flow to use
	flow, err := cfm.flowSelector.SelectFlow(req)
	if err != nil {
		return nil, fmt.Errorf("failed to select flow: %w", err)
	}

	var result *ClassificationResult

	// Execute the selected flow
	switch flow {
	case FlowURLBased:
		result, err = cfm.executeURLBasedFlow(ctx, req, requestID)
	case FlowSearchBased:
		result, err = cfm.executeSearchBasedFlow(ctx, req, requestID)
	default:
		return nil, fmt.Errorf("unknown flow type: %s", flow)
	}

	if err != nil {
		// If fallback is enabled and primary flow failed, try the other flow
		if cfm.config.FallbackEnabled {
			fallbackFlow := cfm.getFallbackFlow(flow)
			fallbackResult, fallbackErr := cfm.executeFallbackFlow(ctx, req, requestID, fallbackFlow)
			if fallbackErr == nil {
				fallbackResult.Warnings = append(fallbackResult.Warnings, "Primary flow failed, using fallback flow")
				return fallbackResult, nil
			}
		}
		return nil, err
	}

	// Set processing time
	result.ProcessingTime = time.Since(start)

	return result, nil
}

// executeURLBasedFlow executes the URL-based classification flow
func (cfm *ClassificationFlowManager) executeURLBasedFlow(ctx context.Context, req *ClassificationRequest, requestID string) (*ClassificationResult, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, cfm.config.URLFlowTimeout)
	defer cancel()

	// Execute URL-based flow
	urlResult, err := cfm.urlFlow.Execute(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("URL-based flow failed: %w", err)
	}

	// Create classification result
	result := &ClassificationResult{
		RequestID:    requestID,
		BusinessName: req.BusinessName,
		FlowUsed:     FlowURLBased,
		WebsiteData:  urlResult.WebsiteData,
		Industries:   urlResult.Industries,
		Confidence:   urlResult.Confidence,
	}

	// Add risk assessment if requested
	if req.IncludeRiskAnalysis {
		result.RiskAssessment = urlResult.RiskAssessment
	}

	// Add connection validation if requested
	if req.IncludeConnectionValidation {
		result.ConnectionValidation = urlResult.ConnectionValidation
	}

	return result, nil
}

// executeSearchBasedFlow executes the search-based classification flow
func (cfm *ClassificationFlowManager) executeSearchBasedFlow(ctx context.Context, req *ClassificationRequest, requestID string) (*ClassificationResult, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, cfm.config.SearchFlowTimeout)
	defer cancel()

	// Execute search-based flow
	searchResult, err := cfm.searchFlow.Execute(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("search-based flow failed: %w", err)
	}

	// Create classification result
	result := &ClassificationResult{
		RequestID:    requestID,
		BusinessName: req.BusinessName,
		FlowUsed:     FlowSearchBased,
		SearchData:   searchResult.SearchData,
		Industries:   searchResult.Industries,
		Confidence:   searchResult.Confidence,
	}

	// Add risk assessment if requested
	if req.IncludeRiskAnalysis {
		result.RiskAssessment = searchResult.RiskAssessment
	}

	// Add connection validation if requested
	if req.IncludeConnectionValidation {
		result.ConnectionValidation = searchResult.ConnectionValidation
	}

	return result, nil
}

// executeFallbackFlow executes the fallback flow when primary flow fails
func (cfm *ClassificationFlowManager) executeFallbackFlow(ctx context.Context, req *ClassificationRequest, requestID string, fallbackFlow ClassificationFlow) (*ClassificationResult, error) {
	switch fallbackFlow {
	case FlowURLBased:
		return cfm.executeURLBasedFlow(ctx, req, requestID)
	case FlowSearchBased:
		return cfm.executeSearchBasedFlow(ctx, req, requestID)
	default:
		return nil, fmt.Errorf("unknown fallback flow: %s", fallbackFlow)
	}
}

// getFallbackFlow returns the fallback flow for a given flow
func (cfm *ClassificationFlowManager) getFallbackFlow(primaryFlow ClassificationFlow) ClassificationFlow {
	switch primaryFlow {
	case FlowURLBased:
		return FlowSearchBased
	case FlowSearchBased:
		return FlowURLBased
	default:
		return FlowSearchBased
	}
}

// GetFlowStats returns statistics about flow usage
func (cfm *ClassificationFlowManager) GetFlowStats() map[string]interface{} {
	cfm.mu.RLock()
	defer cfm.mu.RUnlock()

	return map[string]interface{}{
		"url_flow_stats":      cfm.urlFlow.GetStats(),
		"search_flow_stats":   cfm.searchFlow.GetStats(),
		"flow_selector_stats": cfm.flowSelector.GetStats(),
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
