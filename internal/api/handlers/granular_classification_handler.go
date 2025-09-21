package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pcraw4d/business-verification/internal/api/routing"
	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/modules/risk_assessment"
)

// GranularClassificationHandler handles classification requests with granular feature flag support
type GranularClassificationHandler struct {
	// Core components
	featureFlagManager *config.GranularFeatureFlagManager
	intelligentRouter  *routing.IntelligentRouter
	classifier         *classification.MultiMethodClassifier

	// Configuration
	config GranularClassificationConfig

	// Logging
	logger *log.Logger
}

// GranularClassificationConfig holds configuration for the granular classification handler
type GranularClassificationConfig struct {
	// Request handling
	MaxRequestSize    int64         `json:"max_request_size"`
	RequestTimeout    time.Duration `json:"request_timeout"`
	MaxConcurrentReqs int           `json:"max_concurrent_requests"`

	// Feature flag configuration
	FeatureFlagEnabled bool `json:"feature_flag_enabled"`
	FallbackEnabled    bool `json:"fallback_enabled"`

	// Performance monitoring
	MetricsEnabled      bool `json:"metrics_enabled"`
	PerformanceTracking bool `json:"performance_tracking"`

	// A/B testing
	ABTestingEnabled bool `json:"ab_testing_enabled"`

	// Rollout configuration
	RolloutEnabled bool `json:"rollout_enabled"`
}

// ClassificationRequest represents a classification request
type ClassificationRequest struct {
	// Request information
	RequestID    string `json:"request_id"`
	BusinessName string `json:"business_name"`
	Description  string `json:"description"`
	WebsiteURL   string `json:"website_url"`
	RequestType  string `json:"request_type"` // classification, risk_detection

	// Request metadata
	UserID    string            `json:"user_id"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata"`

	// Performance requirements
	MaxLatency  time.Duration `json:"max_latency"`
	MinAccuracy float64       `json:"min_accuracy"`
	Priority    string        `json:"priority"` // low, medium, high, critical
}

// ClassificationResponse represents a classification response
type ClassificationResponse struct {
	// Response information
	RequestID      string        `json:"request_id"`
	ModelUsed      string        `json:"model_used"`
	ServiceUsed    string        `json:"service_used"`
	ProcessingTime time.Duration `json:"processing_time"`

	// Classification results
	Classification *ClassificationResult                 `json:"classification"`
	RiskAssessment *risk_assessment.RiskAssessmentResult `json:"risk_assessment"`

	// Performance metrics
	Latency    time.Duration `json:"latency"`
	Accuracy   float64       `json:"accuracy"`
	Confidence float64       `json:"confidence"`

	// Feature flag information
	FeatureFlags *FeatureFlagInfo `json:"feature_flags"`

	// Metadata
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ClassificationResult represents classification results
type ClassificationResult struct {
	// Classification data
	IndustryCodes []IndustryCode `json:"industry_codes"`
	Confidence    float64        `json:"confidence"`
	Reasoning     string         `json:"reasoning"`

	// Model information
	ModelVersion string `json:"model_version"`
	ModelType    string `json:"model_type"`
}

// IndustryCode represents an industry classification code
type IndustryCode struct {
	Code        string  `json:"code"`
	Type        string  `json:"type"` // mcc, naics, sic
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// RiskFactor represents a risk factor
type RiskFactor struct {
	Factor      string  `json:"factor"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// FeatureFlagInfo represents feature flag information for the request
type FeatureFlagInfo struct {
	// Service flags
	PythonMLServiceEnabled bool `json:"python_ml_service_enabled"`
	GoRuleEngineEnabled    bool `json:"go_rule_engine_enabled"`

	// Model flags
	BERTClassificationEnabled       bool `json:"bert_classification_enabled"`
	DistilBERTClassificationEnabled bool `json:"distilbert_classification_enabled"`
	CustomNeuralNetEnabled          bool `json:"custom_neural_net_enabled"`

	// A/B testing
	ABTestingEnabled bool   `json:"ab_testing_enabled"`
	TestVariant      string `json:"test_variant"`

	// Rollout
	RolloutEnabled    bool `json:"rollout_enabled"`
	RolloutPercentage int  `json:"rollout_percentage"`
}

// NewGranularClassificationHandler creates a new granular classification handler
func NewGranularClassificationHandler(
	featureFlagManager *config.GranularFeatureFlagManager,
	intelligentRouter *routing.IntelligentRouter,
	classifier *classification.MultiMethodClassifier,
	config GranularClassificationConfig,
	logger *log.Logger,
) *GranularClassificationHandler {
	if logger == nil {
		logger = log.Default()
	}

	return &GranularClassificationHandler{
		featureFlagManager: featureFlagManager,
		intelligentRouter:  intelligentRouter,
		classifier:         classifier,
		config:             config,
		logger:             logger,
	}
}

// ServeHTTP handles HTTP requests for classification
func (gch *GranularClassificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Handle different HTTP methods
	switch r.Method {
	case http.MethodPost:
		gch.handleClassificationRequest(w, r, startTime)
	case http.MethodGet:
		gch.handleStatusRequest(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleClassificationRequest handles classification requests
func (gch *GranularClassificationHandler) handleClassificationRequest(w http.ResponseWriter, r *http.Request, startTime time.Time) {
	// Parse request
	var req ClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := gch.validateRequest(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Set request ID if not provided
	if req.RequestID == "" {
		req.RequestID = gch.generateRequestID()
	}

	// Set timestamp if not provided
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), gch.config.RequestTimeout)
	defer cancel()

	// Process classification request
	response, err := gch.processClassificationRequest(ctx, &req, startTime)
	if err != nil {
		gch.logger.Printf("Classification request failed: %v", err)
		http.Error(w, fmt.Sprintf("Classification failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		gch.logger.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	gch.logger.Printf("Classification request %s completed successfully in %v", req.RequestID, time.Since(startTime))
}

// handleStatusRequest handles status requests
func (gch *GranularClassificationHandler) handleStatusRequest(w http.ResponseWriter, r *http.Request) {
	// Get feature flag status
	flags := gch.featureFlagManager.GetFlags()

	// Get routing metrics
	metrics := gch.intelligentRouter.GetMetrics()

	// Get endpoints status
	endpoints := gch.intelligentRouter.GetEndpoints()

	status := map[string]interface{}{
		"status":          "healthy",
		"timestamp":       time.Now(),
		"feature_flags":   flags,
		"routing_metrics": metrics,
		"endpoints":       endpoints,
		"config":          gch.config,
	}

	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Failed to encode status", http.StatusInternalServerError)
		return
	}
}

// processClassificationRequest processes a classification request
func (gch *GranularClassificationHandler) processClassificationRequest(
	ctx context.Context,
	req *ClassificationRequest,
	startTime time.Time,
) (*ClassificationResponse, error) {
	// Get feature flags
	flags := gch.featureFlagManager.GetFlags()

	// Route request to optimal endpoint
	endpoint, err := gch.intelligentRouter.RouteRequest(ctx, &routing.ClassificationRequest{
		RequestID:    req.RequestID,
		BusinessName: req.BusinessName,
		Description:  req.Description,
		WebsiteURL:   req.WebsiteURL,
		RequestType:  req.RequestType,
		UserID:       req.UserID,
		Timestamp:    req.Timestamp,
		Metadata:     req.Metadata,
		MaxLatency:   req.MaxLatency,
		MinAccuracy:  req.MinAccuracy,
		Priority:     req.Priority,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to route request: %w", err)
	}

	if endpoint == nil {
		return nil, fmt.Errorf("no available endpoint for request")
	}

	// Process classification based on endpoint type
	var classificationResult *ClassificationResult
	var riskAssessmentResult *risk_assessment.RiskAssessmentResult
	var modelUsed, serviceUsed string

	switch endpoint.Type {
	case "python_ml_service":
		classificationResult, riskAssessmentResult, err = gch.processMLClassification(ctx, req, endpoint)
		modelUsed = endpoint.ModelType
		serviceUsed = "python_ml_service"
	case "go_rule_engine":
		classificationResult, riskAssessmentResult, err = gch.processRuleBasedClassification(ctx, req, endpoint)
		modelUsed = endpoint.ModelType
		serviceUsed = "go_rule_engine"
	default:
		return nil, fmt.Errorf("unsupported endpoint type: %s", endpoint.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to process classification: %w", err)
	}

	// Create feature flag info
	featureFlagInfo := &FeatureFlagInfo{
		PythonMLServiceEnabled:          flags.Services.PythonMLServiceEnabled,
		GoRuleEngineEnabled:             flags.Services.GoRuleEngineEnabled,
		BERTClassificationEnabled:       flags.Models.BERTClassificationEnabled,
		DistilBERTClassificationEnabled: flags.Models.DistilBERTClassificationEnabled,
		CustomNeuralNetEnabled:          flags.Models.CustomNeuralNetEnabled,
		ABTestingEnabled:                flags.ABTesting.Enabled,
		TestVariant:                     gch.getTestVariant(ctx, req),
		RolloutEnabled:                  flags.Rollout.GradualRolloutEnabled,
		RolloutPercentage:               flags.Rollout.RolloutPercentage,
	}

	// Create response
	response := &ClassificationResponse{
		RequestID:      req.RequestID,
		ModelUsed:      modelUsed,
		ServiceUsed:    serviceUsed,
		ProcessingTime: time.Since(startTime),
		Classification: classificationResult,
		RiskAssessment: riskAssessmentResult,
		Latency:        time.Since(startTime),
		Accuracy:       gch.calculateAccuracy(classificationResult, riskAssessmentResult),
		Confidence:     riskAssessmentResult.ConfidenceScore,
		FeatureFlags:   featureFlagInfo,
		Timestamp:      time.Now(),
		Metadata: map[string]interface{}{
			"endpoint_url": endpoint.URL,
			"request_type": req.RequestType,
			"user_id":      req.UserID,
		},
	}

	return response, nil
}

// processMLClassification processes classification using ML models
func (gch *GranularClassificationHandler) processMLClassification(
	ctx context.Context,
	req *ClassificationRequest,
	endpoint *routing.ServiceEndpoint,
) (*ClassificationResult, *risk_assessment.RiskAssessmentResult, error) {
	// Use the existing classifier with ML methods
	result, err := gch.classifier.ClassifyWithMultipleMethods(
		ctx,
		req.BusinessName,
		req.Description,
		req.WebsiteURL,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("ML classification failed: %w", err)
	}

	// Convert to our response format
	classificationResult := &ClassificationResult{
		IndustryCodes: gch.convertIndustryCodesFromResult(result),
		Confidence:    result.EnsembleConfidence,
		Reasoning:     result.ClassificationReasoning,
		ModelVersion:  "v1.0",
		ModelType:     endpoint.ModelType,
	}

	// Create risk assessment result
	riskScore := gch.calculateRiskScoreFromResult(result)
	riskAssessmentResult := &risk_assessment.RiskAssessmentResult{
		RequestID:           req.RequestID,
		BusinessName:        req.BusinessName,
		WebsiteURL:          req.WebsiteURL,
		AssessmentTimestamp: time.Now(),
		OverallRiskScore:    riskScore,
		RiskLevel:           risk_assessment.RiskLevel(gch.determineRiskLevel(riskScore)),
		RiskCategory:        risk_assessment.RiskCategoryOperational,
		RiskFactors:         gch.extractRiskFactorsFromResult(result),
		ConfidenceScore:     result.EnsembleConfidence,
	}

	return classificationResult, riskAssessmentResult, nil
}

// processRuleBasedClassification processes classification using rule-based methods
func (gch *GranularClassificationHandler) processRuleBasedClassification(
	ctx context.Context,
	req *ClassificationRequest,
	endpoint *routing.ServiceEndpoint,
) (*ClassificationResult, *risk_assessment.RiskAssessmentResult, error) {
	// Use the existing classifier with rule-based methods
	result, err := gch.classifier.ClassifyWithMultipleMethods(
		ctx,
		req.BusinessName,
		req.Description,
		req.WebsiteURL,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("rule-based classification failed: %w", err)
	}

	// Convert to our response format
	classificationResult := &ClassificationResult{
		IndustryCodes: gch.convertIndustryCodesFromResult(result),
		Confidence:    result.EnsembleConfidence,
		Reasoning:     result.ClassificationReasoning,
		ModelVersion:  "v1.0",
		ModelType:     endpoint.ModelType,
	}

	// Create risk assessment result
	riskScore := gch.calculateRiskScoreFromResult(result)
	riskAssessmentResult := &risk_assessment.RiskAssessmentResult{
		RequestID:           req.RequestID,
		BusinessName:        req.BusinessName,
		WebsiteURL:          req.WebsiteURL,
		AssessmentTimestamp: time.Now(),
		OverallRiskScore:    riskScore,
		RiskLevel:           risk_assessment.RiskLevel(gch.determineRiskLevel(riskScore)),
		RiskCategory:        risk_assessment.RiskCategoryOperational,
		RiskFactors:         gch.extractRiskFactorsFromResult(result),
		ConfidenceScore:     result.EnsembleConfidence,
	}

	return classificationResult, riskAssessmentResult, nil
}

// validateRequest validates a classification request
func (gch *GranularClassificationHandler) validateRequest(req *ClassificationRequest) error {
	if req.BusinessName == "" {
		return fmt.Errorf("business name is required")
	}

	if req.RequestType == "" {
		req.RequestType = "classification"
	}

	if req.Priority == "" {
		req.Priority = "medium"
	}

	return nil
}

// generateRequestID generates a unique request ID
func (gch *GranularClassificationHandler) generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// getTestVariant returns the A/B test variant for the request
func (gch *GranularClassificationHandler) getTestVariant(ctx context.Context, req *ClassificationRequest) string {
	// Simple A/B test variant determination
	// In a real implementation, this would use proper A/B testing logic
	if req.RequestID != "" {
		// Use request ID to determine variant
		if len(req.RequestID)%2 == 0 {
			return "control"
		}
		return "test"
	}
	return "control"
}

// convertIndustryCodesFromResult converts industry codes from classification result to our format
func (gch *GranularClassificationHandler) convertIndustryCodesFromResult(result *classification.MultiMethodClassificationResult) []IndustryCode {
	// Extract industry codes from the primary classification
	var codes []IndustryCode

	if result.PrimaryClassification != nil {
		// Add the primary classification as an industry code
		codes = append(codes, IndustryCode{
			Code:        result.PrimaryClassification.IndustryCode,
			Type:        "primary",
			Description: result.PrimaryClassification.IndustryName,
			Confidence:  result.PrimaryClassification.ConfidenceScore,
		})
	}

	return codes
}

// calculateRiskScoreFromResult calculates a risk score from classification results
func (gch *GranularClassificationHandler) calculateRiskScoreFromResult(result *classification.MultiMethodClassificationResult) float64 {
	// Simple risk score calculation
	// In a real implementation, this would use proper risk assessment logic
	baseScore := 0.1

	// Add risk based on primary classification
	if result.PrimaryClassification != nil {
		// Check for high-risk industry codes
		if gch.isHighRiskMCC(result.PrimaryClassification.IndustryCode) {
			baseScore += 0.3
		}
	}

	// Add risk based on confidence
	if result.EnsembleConfidence < 0.7 {
		baseScore += 0.2
	}

	// Ensure score is between 0 and 1
	if baseScore > 1.0 {
		baseScore = 1.0
	}

	return baseScore
}

// determineRiskLevel determines risk level from risk score
func (gch *GranularClassificationHandler) determineRiskLevel(riskScore float64) string {
	if riskScore < 0.3 {
		return "low"
	} else if riskScore < 0.6 {
		return "medium"
	} else if riskScore < 0.8 {
		return "high"
	} else {
		return "critical"
	}
}

// extractRiskFactorsFromResult extracts risk factors from classification results
func (gch *GranularClassificationHandler) extractRiskFactorsFromResult(result *classification.MultiMethodClassificationResult) []risk_assessment.RiskFactor {
	var factors []risk_assessment.RiskFactor

	// Check for high-risk industry codes
	if result.PrimaryClassification != nil {
		if gch.isHighRiskMCC(result.PrimaryClassification.IndustryCode) {
			factors = append(factors, risk_assessment.RiskFactor{
				Category:    risk_assessment.RiskCategoryOperational,
				Factor:      "high_risk_industry",
				Description: fmt.Sprintf("High-risk industry code: %s", result.PrimaryClassification.IndustryCode),
				Severity:    risk_assessment.RiskLevelHigh,
				Score:       result.PrimaryClassification.ConfidenceScore,
				Evidence:    fmt.Sprintf("Industry code %s identified as high-risk", result.PrimaryClassification.IndustryCode),
				Impact:      "Increased regulatory scrutiny and compliance requirements",
			})
		}
	}

	// Check for low confidence
	if result.EnsembleConfidence < 0.7 {
		factors = append(factors, risk_assessment.RiskFactor{
			Category:    risk_assessment.RiskCategoryOperational,
			Factor:      "low_confidence",
			Description: "Low classification confidence",
			Severity:    risk_assessment.RiskLevelMedium,
			Score:       1.0 - result.EnsembleConfidence,
			Evidence:    fmt.Sprintf("Classification confidence: %.2f", result.EnsembleConfidence),
			Impact:      "Uncertainty in business classification may affect risk assessment",
		})
	}

	return factors
}

// isHighRiskMCC checks if an MCC code is high-risk
func (gch *GranularClassificationHandler) isHighRiskMCC(mccCode string) bool {
	// Simple high-risk MCC code check
	// In a real implementation, this would use a comprehensive list
	highRiskMCCs := map[string]bool{
		"7995": true, // Gambling
		"7273": true, // Dating services
		"5967": true, // Direct marketing
		"5993": true, // Cigar stores
		"5994": true, // News dealers
	}
	return highRiskMCCs[mccCode]
}

// calculateAccuracy calculates overall accuracy from results
func (gch *GranularClassificationHandler) calculateAccuracy(
	classification *ClassificationResult,
	riskAssessment *risk_assessment.RiskAssessmentResult,
) float64 {
	// Simple accuracy calculation
	// In a real implementation, this would use proper accuracy metrics
	accuracy := classification.Confidence
	if riskAssessment != nil {
		accuracy = (accuracy + (1.0 - riskAssessment.OverallRiskScore)) / 2.0
	}
	return accuracy
}

// calculateConfidence calculates overall confidence from results
func (gch *GranularClassificationHandler) calculateConfidence(
	classification *ClassificationResult,
	riskAssessment *risk_assessment.RiskAssessmentResult,
) float64 {
	// Simple confidence calculation
	// In a real implementation, this would use proper confidence metrics
	confidence := classification.Confidence
	if riskAssessment != nil {
		confidence = (confidence + riskAssessment.ConfidenceScore) / 2.0
	}
	return confidence
}
