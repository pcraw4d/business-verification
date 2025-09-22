package keyword_classification

import (
	"context"
	"fmt"
	"strings"
	"time"

	"kyb-platform/internal/architecture"
	"kyb-platform/internal/config"
	"kyb-platform/internal/database"
	"kyb-platform/internal/observability"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ClassificationRequest represents a request for business classification
type ClassificationRequest struct {
	BusinessName string   `json:"business_name"`
	Description  string   `json:"description"`
	Keywords     []string `json:"keywords"`
}

// IndustryClassification represents the result of industry classification
type IndustryClassification struct {
	IndustryCode         string   `json:"industry_code"`
	IndustryName         string   `json:"industry_name"`
	ConfidenceScore      float64  `json:"confidence_score"`
	ClassificationMethod string   `json:"classification_method"`
	Description          string   `json:"description"`
	MatchedKeywords      []string `json:"matched_keywords"`
}

// KeywordClassificationModule implements the Module interface for keyword-based classification
type KeywordClassificationModule struct {
	id        string
	config    architecture.ModuleConfig
	running   bool
	logger    *observability.ModuleLogger
	metrics   *observability.Metrics
	tracer    trace.Tracer
	db        database.Database
	appConfig *config.Config

	// Database-driven keyword classification
	keywordRepo database.KeywordRepository
}

// NewKeywordClassificationModule creates a new keyword classification module
func NewKeywordClassificationModule(keywordRepo database.KeywordRepository) *KeywordClassificationModule {
	return &KeywordClassificationModule{
		id:          "keyword_classification_module",
		keywordRepo: keywordRepo,
	}
}

// Module interface implementation
func (m *KeywordClassificationModule) ID() string {
	return m.id
}

func (m *KeywordClassificationModule) Config() architecture.ModuleConfig {
	return m.config
}

func (m *KeywordClassificationModule) UpdateConfig(config architecture.ModuleConfig) error {
	m.config = config

	// Log configuration update
	if m.logger != nil {
		m.logger.LogModuleConfig(context.Background(), "module_config", map[string]interface{}{
			"enabled": config.Enabled,
		})
	}

	return nil
}

func (m *KeywordClassificationModule) Health() architecture.ModuleHealth {
	status := architecture.ModuleStatusStopped
	if m.running {
		status = architecture.ModuleStatusRunning
	}

	health := architecture.ModuleHealth{
		Status:    status,
		LastCheck: time.Now(),
		Message:   "Keyword classification module health check",
	}

	// Log health status
	if m.logger != nil {
		m.logger.LogModuleHealth(context.Background(), m.running, health.Message, map[string]interface{}{
			"status":     string(status),
			"last_check": health.LastCheck,
		})
	}

	return health
}

func (m *KeywordClassificationModule) Metadata() architecture.ModuleMetadata {
	return architecture.ModuleMetadata{
		Name:        "Keyword Classification Module",
		Version:     "1.0.0",
		Description: "Performs business classification using keyword analysis",
		Capabilities: []architecture.ModuleCapability{
			architecture.CapabilityClassification,
		},
		Priority: architecture.PriorityHigh,
	}
}

func (m *KeywordClassificationModule) Start(ctx context.Context) error {
	startTime := time.Now()
	_, span := m.tracer.Start(ctx, "KeywordClassificationModule.Start")
	defer span.End()

	if m.running {
		return fmt.Errorf("module %s is already running", m.id)
	}

	// Log module start
	if m.logger != nil {
		m.logger.LogModuleStart(ctx, map[string]interface{}{
			"start_time": startTime,
			"module_id":  m.id,
		})
	}

	// Initialize database connection and validate data
	if m.keywordRepo == nil {
		span.RecordError(fmt.Errorf("keyword repository not initialized"))
		return fmt.Errorf("keyword repository not initialized")
	}

	// Test database connectivity
	if err := m.testDatabaseConnection(ctx); err != nil {
		span.RecordError(err)
		if m.logger != nil {
			m.logger.LogModuleError(ctx, "database_connection_test", err, map[string]interface{}{
				"operation": "startup",
			})
		}
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Log successful database connection
	if m.logger != nil {
		m.logger.Info("Database connected", map[string]interface{}{
			"module_id": m.id,
			"status":    "connected",
		})
	}

	m.running = true

	// Emit module started event
	if emitEvent != nil {
		emitEvent(architecture.Event{
			Type:     architecture.EventTypeModuleStarted,
			Source:   m.id,
			Priority: architecture.EventPriorityNormal,
			Data: map[string]interface{}{
				"module_id":  m.id,
				"start_time": startTime,
			},
		})
	}

	// Log performance metrics
	if m.logger != nil {
		m.logger.LogModulePerformance(ctx, "module_start", startTime, time.Now(), map[string]interface{}{
			"database_connected": true,
			"module_type":        "database_driven",
		})
	}

	span.SetAttributes(attribute.String("module.id", m.id))

	return nil
}

func (m *KeywordClassificationModule) Stop(ctx context.Context) error {
	_, span := m.tracer.Start(ctx, "KeywordClassificationModule.Stop")
	defer span.End()

	if !m.running {
		return fmt.Errorf("module %s is not running", m.id)
	}

	m.running = false

	// Log module stop
	if m.logger != nil {
		m.logger.LogModuleStop(ctx, "manual_stop")
	}

	// Emit module stopped event
	if emitEvent != nil {
		emitEvent(architecture.Event{
			Type:     architecture.EventTypeModuleStopped,
			Source:   m.id,
			Priority: architecture.EventPriorityNormal,
			Data: map[string]interface{}{
				"module_id": m.id,
				"reason":    "manual_stop",
			},
		})
	}

	span.SetAttributes(attribute.String("module.id", m.id))

	return nil
}

func (m *KeywordClassificationModule) IsRunning() bool {
	return m.running
}

func (m *KeywordClassificationModule) HealthCheck(ctx context.Context) error {
	_, span := m.tracer.Start(ctx, "KeywordClassificationModule.HealthCheck")
	defer span.End()

	if !m.running {
		return fmt.Errorf("module is not running")
	}

	// Check if database repository is available
	if m.keywordRepo == nil {
		return fmt.Errorf("keyword repository not initialized")
	}

	// Test database connectivity
	_, err := m.keywordRepo.ListIndustries(ctx, "")
	if err != nil {
		return fmt.Errorf("database connectivity check failed: %w", err)
	}

	span.SetAttributes(attribute.String("module.id", m.id))
	return nil
}

// CanHandle determines if this module can handle the given request
func (m *KeywordClassificationModule) CanHandle(req architecture.ModuleRequest) bool {
	// This module can handle classification requests
	if req.Type == "classification" || req.Type == "keyword_classification" {
		return true
	}
	return false
}

// Process processes a module request
func (m *KeywordClassificationModule) Process(ctx context.Context, req architecture.ModuleRequest) (architecture.ModuleResponse, error) {
	startTime := time.Now()

	// Convert the request to a classification request
	classificationReq := &ClassificationRequest{
		BusinessName: req.Data["business_name"].(string),
		Description:  req.Data["description"].(string),
	}

	// Process the classification
	classifications, err := m.PerformKeywordClassification(ctx, classificationReq)
	if err != nil {
		return architecture.ModuleResponse{
			ID:         req.ID,
			Success:    false,
			Error:      err.Error(),
			Confidence: 0.0,
			Latency:    time.Since(startTime),
		}, err
	}

	// Calculate overall confidence from classifications
	overallConfidence := 0.0
	if len(classifications) > 0 {
		overallConfidence = classifications[0].ConfidenceScore // Use highest confidence
	}

	// Return the results
	return architecture.ModuleResponse{
		ID:         req.ID,
		Success:    true,
		Confidence: overallConfidence,
		Latency:    time.Since(startTime),
		Data: map[string]interface{}{
			"classifications": classifications,
		},
	}, nil
}

// OnEvent handles module events
func (m *KeywordClassificationModule) OnEvent(event architecture.ModuleEvent) error {
	// Log the event
	if m.logger != nil {
		m.logger.Info("Module event received", map[string]interface{}{
			"event_type": event.Type,
			"module_id":  event.ModuleID,
			"timestamp":  event.Timestamp,
		})
	}
	return nil
}

// PerformKeywordClassification performs database-driven keyword-based classification
func (m *KeywordClassificationModule) PerformKeywordClassification(ctx context.Context, req *ClassificationRequest) ([]IndustryClassification, error) {
	_, span := m.tracer.Start(ctx, "performKeywordClassification")
	defer span.End()

	span.SetAttributes(attribute.String("business_name", req.BusinessName))

	// Normalize business fields
	normalized, tokens := m.normalizeBusinessFields(req.BusinessName, req.Description, req.Keywords)
	if normalized == "" {
		return nil, fmt.Errorf("no valid business data to classify")
	}

	var classifications []IndustryClassification
	matchedKeywords := make(map[string][]string)

	// Get all industries from database
	industries, err := m.keywordRepo.ListIndustries(ctx, "")
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list industries: %w", err)
	}

	// Process each industry to find keyword matches
	for _, industry := range industries {
		// Get keywords for this industry
		keywords, err := m.keywordRepo.GetKeywordsByIndustry(ctx, industry.ID)
		if err != nil {
			// Log error but continue with other industries
			if m.logger != nil {
				m.logger.LogModuleError(ctx, "get_keywords_for_industry", err, map[string]interface{}{
					"industry_id":   industry.ID,
					"industry_name": industry.Name,
				})
			}
			continue
		}

		// Find keyword matches
		matched := m.findKeywordMatchesInDatabase(normalized, tokens, keywords)
		if len(matched) > 0 {
			// Get classification codes for this industry
			codes, err := m.keywordRepo.GetClassificationCodesByIndustry(ctx, industry.ID)
			if err != nil {
				// Log error but continue
				if m.logger != nil {
					m.logger.LogModuleError(ctx, "get_classification_codes", err, map[string]interface{}{
						"industry_id": industry.ID,
					})
				}
			}

			// Calculate confidence score based on keyword matches and weights
			confidence := m.calculateDatabaseConfidence(matched, keywords)

			// Get primary industry code (prefer NAICS, then MCC, then SIC)
			industryCode := m.getPrimaryIndustryCode(codes)

			classification := IndustryClassification{
				IndustryCode:         industryCode,
				IndustryName:         industry.Name,
				ConfidenceScore:      confidence,
				ClassificationMethod: "database_keyword_classification",
				Description:          fmt.Sprintf("Database-driven classification with %d keyword matches", len(matched)),
				MatchedKeywords:      matched,
			}

			classifications = append(classifications, classification)
			matchedKeywords[industry.Name] = matched
		}
	}

	// Sort by confidence score (highest first)
	classifications = m.sortByConfidence(classifications)

	span.SetAttributes(
		attribute.Int("classifications_count", len(classifications)),
		attribute.Int("matched_keywords_total", len(matchedKeywords)),
	)

	return classifications, nil
}

// normalizeBusinessFields normalizes business fields for keyword matching
func (m *KeywordClassificationModule) normalizeBusinessFields(businessName, description string, keywords []string) (string, []string) {
	// Combine all text fields
	combined := strings.Join([]string{businessName, description, strings.Join(keywords, " ")}, " ")

	// Normalize to lowercase
	normalized := strings.ToLower(combined)

	// Remove common stop words and split into tokens
	tokens := strings.Fields(normalized)

	// Filter out short tokens and common stop words
	var filteredTokens []string
	stopWords := map[string]bool{
		"the": true, "and": true, "or": true, "but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "a": true, "an": true, "is": true, "are": true,
	}

	for _, token := range tokens {
		if len(token) > 2 && !stopWords[token] {
			filteredTokens = append(filteredTokens, token)
		}
	}

	return normalized, filteredTokens
}

// findKeywordMatchesInDatabase finds keyword matches using database keywords
func (m *KeywordClassificationModule) findKeywordMatchesInDatabase(normalized string, tokens []string, keywords []*database.IndustryKeyword) []string {
	var matches []string

	for _, keyword := range keywords {
		if !keyword.IsActive {
			continue
		}

		// Check for exact matches in normalized text
		if strings.Contains(strings.ToLower(normalized), strings.ToLower(keyword.Keyword)) {
			matches = append(matches, keyword.Keyword)
		}

		// Check for token matches
		for _, token := range tokens {
			if strings.EqualFold(token, keyword.Keyword) {
				matches = append(matches, keyword.Keyword)
			}
		}
	}

	return matches
}

// calculateDatabaseConfidence calculates confidence score based on database keyword matches and weights
func (m *KeywordClassificationModule) calculateDatabaseConfidence(matched []string, keywords []*database.IndustryKeyword) float64 {
	if len(matched) == 0 {
		return 0.0
	}

	// Create a map of keyword weights for quick lookup
	keywordWeights := make(map[string]float64)
	for _, keyword := range keywords {
		keywordWeights[keyword.Keyword] = keyword.Weight
	}

	// Calculate weighted confidence
	totalWeight := 0.0
	matchedWeight := 0.0

	for _, keyword := range keywords {
		totalWeight += keyword.Weight
	}

	for _, match := range matched {
		if weight, exists := keywordWeights[match]; exists {
			matchedWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0.0
	}

	// Base confidence on weighted match ratio
	confidence := matchedWeight / totalWeight

	// Apply confidence scaling (max 95%)
	confidence = confidence * 0.95

	// Boost confidence for multiple matches
	if len(matched) > 1 {
		confidence += 0.02 * float64(len(matched)-1)
	}

	// Cap at 95%
	if confidence > 0.95 {
		confidence = 0.95
	}

	return confidence
}

// getPrimaryIndustryCode gets the primary industry code, preferring NAICS > MCC > SIC
func (m *KeywordClassificationModule) getPrimaryIndustryCode(codes []*database.ClassificationCode) string {
	if len(codes) == 0 {
		return ""
	}

	// Prefer NAICS codes
	for _, code := range codes {
		if code.CodeType == "naics" && code.IsActive {
			return code.Code
		}
	}

	// Fall back to MCC codes
	for _, code := range codes {
		if code.CodeType == "mcc" && code.IsActive {
			return code.Code
		}
	}

	// Fall back to SIC codes
	for _, code := range codes {
		if code.CodeType == "sic" && code.IsActive {
			return code.Code
		}
	}

	// Return the first active code if no preferred type found
	for _, code := range codes {
		if code.IsActive {
			return code.Code
		}
	}

	return ""
}

// sortByConfidence sorts classifications by confidence score (highest first)
func (m *KeywordClassificationModule) sortByConfidence(classifications []IndustryClassification) []IndustryClassification {
	// Simple bubble sort for small lists
	for i := 0; i < len(classifications)-1; i++ {
		for j := 0; j < len(classifications)-i-1; j++ {
			if classifications[j].ConfidenceScore < classifications[j+1].ConfidenceScore {
				classifications[j], classifications[j+1] = classifications[j+1], classifications[j]
			}
		}
	}
	return classifications
}

// testDatabaseConnection tests the database connection and validates data availability
func (m *KeywordClassificationModule) testDatabaseConnection(ctx context.Context) error {
	// Test basic connectivity by listing industries
	industries, err := m.keywordRepo.ListIndustries(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to list industries: %w", err)
	}

	if len(industries) == 0 {
		return fmt.Errorf("no industries found in database")
	}

	// Test keyword retrieval for the first industry
	if len(industries) > 0 {
		keywords, err := m.keywordRepo.GetKeywordsByIndustry(ctx, industries[0].ID)
		if err != nil {
			return fmt.Errorf("failed to get keywords for industry %d: %w", industries[0].ID, err)
		}

		// Log database statistics
		if m.logger != nil {
			m.logger.Info("Database validation", map[string]interface{}{
				"total_industries": len(industries),
				"sample_industry":  industries[0].Name,
				"sample_keywords":  len(keywords),
			})
		}
	}

	return nil
}

// Event emission function (will be injected by the module manager)
var emitEvent func(architecture.Event) error

// SetEventEmitter sets the event emission function
func (m *KeywordClassificationModule) SetEventEmitter(emitter func(architecture.Event) error) {
	emitEvent = emitter
}
