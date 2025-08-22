package keyword_classification

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/architecture"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

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

	// Keyword classification specific fields
	keywordMappings  map[string][]string
	industryCodes    map[string]string
	confidenceScores map[string]float64
}

// NewKeywordClassificationModule creates a new keyword classification module
func NewKeywordClassificationModule() *KeywordClassificationModule {
	return &KeywordClassificationModule{
		id:               "keyword_classification_module",
		keywordMappings:  make(map[string][]string),
		industryCodes:    make(map[string]string),
		confidenceScores: make(map[string]float64),
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

	// Initialize keyword mappings
	if err := m.initializeKeywordMappings(); err != nil {
		span.RecordError(err)
		if m.logger != nil {
			m.logger.LogModuleError(ctx, "initialize_keyword_mappings", err, map[string]interface{}{
				"operation": "startup",
			})
		}
		return fmt.Errorf("failed to initialize keyword mappings: %w", err)
	}

	// Initialize industry codes
	if err := m.initializeIndustryCodes(); err != nil {
		span.RecordError(err)
		if m.logger != nil {
			m.logger.LogModuleError(ctx, "initialize_industry_codes", err, map[string]interface{}{
				"operation": "startup",
			})
		}
		return fmt.Errorf("failed to initialize industry codes: %w", err)
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

	span.SetAttributes(attribute.String("module.id", m.id))

	// Log performance metrics
	if m.logger != nil {
		m.logger.LogModulePerformance(ctx, "module_start", startTime, time.Now(), map[string]interface{}{
			"keyword_mappings_count": len(m.keywordMappings),
			"industry_codes_count":   len(m.industryCodes),
		})
	}

	return nil
}

func (m *KeywordClassificationModule) Stop(ctx context.Context) error {
	startTime := time.Now()
	_, span := m.tracer.Start(ctx, "KeywordClassificationModule.Stop")
	defer span.End()

	if !m.running {
		return fmt.Errorf("module %s is not running", m.id)
	}

	m.running = false

	// Log module stop
	if m.logger != nil {
		m.logger.LogModuleStop(ctx, "graceful shutdown")
	}

	// Emit module stopped event
	if emitEvent != nil {
		emitEvent(architecture.Event{
			Type:     architecture.EventTypeModuleStopped,
			Source:   m.id,
			Priority: architecture.EventPriorityNormal,
			Data: map[string]interface{}{
				"module_id": m.id,
				"stop_time": startTime,
			},
		})
	}

	span.SetAttributes(attribute.String("module.id", m.id))

	// Log performance metrics
	if m.logger != nil {
		m.logger.LogModulePerformance(ctx, "module_stop", startTime, time.Now(), nil)
	}

	return nil
}

func (m *KeywordClassificationModule) IsRunning() bool {
	return m.running
}

func (m *KeywordClassificationModule) Process(ctx context.Context, req architecture.ModuleRequest) (architecture.ModuleResponse, error) {
	startTime := time.Now()
	_, span := m.tracer.Start(ctx, "KeywordClassificationModule.Process")
	defer span.End()

	span.SetAttributes(
		attribute.String("module.id", m.id),
		attribute.String("request.type", req.Type),
	)

	// Log incoming request
	if m.logger != nil {
		m.logger.LogModuleRequest(ctx, req.Type, req.ID, len(req.Data))
	}

	// Check if this module can handle the request
	if !m.CanHandle(req) {
		response := architecture.ModuleResponse{
			ID:      req.ID,
			Success: false,
			Error:   "unsupported request type",
		}

		// Log response
		if m.logger != nil {
			m.logger.LogModuleResponse(ctx, req.ID, false, 0, time.Since(startTime))
		}

		return response, nil
	}

	// Parse the request payload
	classificationReq, err := m.parseClassificationRequest(req.Data)
	if err != nil {
		span.RecordError(err)
		if m.logger != nil {
			m.logger.LogModuleError(ctx, "parse_classification_request", err, map[string]interface{}{
				"request_id":   req.ID,
				"request_type": req.Type,
			})
		}

		response := architecture.ModuleResponse{
			ID:      req.ID,
			Success: false,
			Error:   fmt.Sprintf("failed to parse request: %v", err),
		}

		// Log response
		if m.logger != nil {
			m.logger.LogModuleResponse(ctx, req.ID, false, 0, time.Since(startTime))
		}

		return response, nil
	}

	// Perform keyword classification
	classifications, err := m.performKeywordClassification(ctx, classificationReq)
	if err != nil {
		span.RecordError(err)
		if m.logger != nil {
			m.logger.LogModuleError(ctx, "perform_keyword_classification", err, map[string]interface{}{
				"request_id":    req.ID,
				"business_name": classificationReq.BusinessName,
			})
		}

		response := architecture.ModuleResponse{
			ID:      req.ID,
			Success: false,
			Error:   fmt.Sprintf("classification failed: %v", err),
		}

		// Log response
		if m.logger != nil {
			m.logger.LogModuleResponse(ctx, req.ID, false, 0, time.Since(startTime))
		}

		return response, nil
	}

	// Create response
	response := architecture.ModuleResponse{
		ID:      req.ID,
		Success: true,
		Data: map[string]interface{}{
			"classifications": classifications,
			"method":          "keyword_classification",
			"module_id":       m.id,
		},
	}

	// Log successful response
	if m.logger != nil {
		m.logger.LogModuleResponse(ctx, req.ID, true, len(classifications), time.Since(startTime))
	}

	return response, nil
}

func (m *KeywordClassificationModule) CanHandle(req architecture.ModuleRequest) bool {
	return req.Type == "classify_by_keywords"
}

func (m *KeywordClassificationModule) HealthCheck(ctx context.Context) error {
	_, span := m.tracer.Start(ctx, "KeywordClassificationModule.HealthCheck")
	defer span.End()

	if !m.running {
		return fmt.Errorf("module is not running")
	}

	// Check if keyword mappings are loaded
	if len(m.keywordMappings) == 0 {
		return fmt.Errorf("keyword mappings not initialized")
	}

	// Check if industry codes are loaded
	if len(m.industryCodes) == 0 {
		return fmt.Errorf("industry codes not initialized")
	}

	span.SetAttributes(attribute.String("module.id", m.id))
	return nil
}

func (m *KeywordClassificationModule) OnEvent(event architecture.ModuleEvent) error {
	// Handle events if needed
	return nil
}

// Keyword classification specific methods

// ClassificationRequest represents a keyword classification request
type ClassificationRequest struct {
	BusinessName       string `json:"business_name"`
	Description        string `json:"description"`
	Keywords           string `json:"keywords"`
	BusinessType       string `json:"business_type"`
	Industry           string `json:"industry"`
	RegistrationNumber string `json:"registration_number"`
	WebsiteURL         string `json:"website_url"`
}

// IndustryClassification represents a classification result
type IndustryClassification struct {
	IndustryCode         string   `json:"industry_code"`
	IndustryName         string   `json:"industry_name"`
	ConfidenceScore      float64  `json:"confidence_score"`
	ClassificationMethod string   `json:"classification_method"`
	Description          string   `json:"description"`
	MatchedKeywords      []string `json:"matched_keywords"`
}

// parseClassificationRequest parses the module request into a classification request
func (m *KeywordClassificationModule) parseClassificationRequest(payload map[string]interface{}) (*ClassificationRequest, error) {
	req := &ClassificationRequest{}

	if businessName, ok := payload["business_name"].(string); ok {
		req.BusinessName = businessName
	}

	if description, ok := payload["description"].(string); ok {
		req.Description = description
	}

	if keywords, ok := payload["keywords"].(string); ok {
		req.Keywords = keywords
	}

	if businessType, ok := payload["business_type"].(string); ok {
		req.BusinessType = businessType
	}

	if industry, ok := payload["industry"].(string); ok {
		req.Industry = industry
	}

	if registrationNumber, ok := payload["registration_number"].(string); ok {
		req.RegistrationNumber = registrationNumber
	}

	if websiteURL, ok := payload["website_url"].(string); ok {
		req.WebsiteURL = websiteURL
	}

	return req, nil
}

// performKeywordClassification performs keyword-based classification
func (m *KeywordClassificationModule) performKeywordClassification(ctx context.Context, req *ClassificationRequest) ([]IndustryClassification, error) {
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

	// Check each industry category for keyword matches
	for industry, keywords := range m.keywordMappings {
		matched := m.findKeywordMatches(normalized, tokens, keywords)
		if len(matched) > 0 {
			// Get industry code
			industryCode := m.industryCodes[industry]
			if industryCode == "" {
				industryCode = industry // Fallback to industry name as code
			}

			// Calculate confidence score
			confidence := m.calculateKeywordConfidence(matched, keywords)

			classification := IndustryClassification{
				IndustryCode:         industryCode,
				IndustryName:         industry,
				ConfidenceScore:      confidence,
				ClassificationMethod: "keyword_classification",
				Description:          fmt.Sprintf("Keyword-based classification with %d matches", len(matched)),
				MatchedKeywords:      matched,
			}

			classifications = append(classifications, classification)
			matchedKeywords[industry] = matched
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
func (m *KeywordClassificationModule) normalizeBusinessFields(businessName, description, keywords string) (string, []string) {
	var fields []string

	// Add business name
	if businessName != "" {
		fields = append(fields, strings.ToLower(strings.TrimSpace(businessName)))
	}

	// Add description
	if description != "" {
		fields = append(fields, strings.ToLower(strings.TrimSpace(description)))
	}

	// Add keywords
	if keywords != "" {
		fields = append(fields, strings.ToLower(strings.TrimSpace(keywords)))
	}

	if len(fields) == 0 {
		return "", nil
	}

	// Join all fields
	normalized := strings.Join(fields, " ")

	// Extract tokens
	tokens := strings.Fields(normalized)

	return normalized, tokens
}

// findKeywordMatches finds keyword matches in the normalized text
func (m *KeywordClassificationModule) findKeywordMatches(normalized string, tokens []string, keywords []string) []string {
	var matches []string

	for _, keyword := range keywords {
		keywordLower := strings.ToLower(keyword)

		// Check for exact match in normalized text
		if strings.Contains(normalized, keywordLower) {
			matches = append(matches, keyword)
			continue
		}

		// Check for token matches
		for _, token := range tokens {
			if strings.Contains(token, keywordLower) || strings.Contains(keywordLower, token) {
				matches = append(matches, keyword)
				break
			}
		}
	}

	return matches
}

// calculateKeywordConfidence calculates confidence score based on keyword matches
func (m *KeywordClassificationModule) calculateKeywordConfidence(matched []string, totalKeywords []string) float64 {
	if len(totalKeywords) == 0 {
		return 0.0
	}

	// Base confidence on match ratio
	matchRatio := float64(len(matched)) / float64(len(totalKeywords))

	// Apply confidence scoring
	confidence := matchRatio * 0.9 // Max 90% confidence for keyword matching

	// Boost confidence for multiple matches
	if len(matched) > 1 {
		confidence += 0.05 * float64(len(matched)-1)
	}

	// Cap at 95%
	if confidence > 0.95 {
		confidence = 0.95
	}

	return confidence
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

// initializeKeywordMappings initializes keyword mappings for different industries
func (m *KeywordClassificationModule) initializeKeywordMappings() error {
	m.keywordMappings = map[string][]string{
		"Grocery & Food Retail": {
			"grocery", "supermarket", "food", "market", "store", "retail", "fresh", "organic",
			"produce", "meat", "dairy", "bakery", "deli", "convenience", "shop",
		},
		"Financial Services": {
			"bank", "financial", "credit", "loan", "mortgage", "investment", "insurance",
			"wealth", "asset", "fund", "capital", "finance", "lending", "savings",
		},
		"Healthcare": {
			"health", "medical", "hospital", "clinic", "doctor", "physician", "nurse",
			"pharmacy", "dental", "therapy", "wellness", "care", "treatment", "medicine",
		},
		"Technology": {
			"tech", "software", "hardware", "computer", "digital", "internet", "web",
			"app", "platform", "system", "data", "cloud", "ai", "machine learning",
		},
		"Real Estate": {
			"real estate", "property", "housing", "apartment", "rental", "leasing",
			"construction", "development", "building", "home", "house", "commercial",
		},
		"Transportation": {
			"transport", "logistics", "shipping", "delivery", "freight", "trucking",
			"warehouse", "storage", "supply chain", "distribution", "courier",
		},
		"Education": {
			"education", "school", "university", "college", "academy", "training",
			"learning", "course", "program", "institute", "center", "tutoring",
		},
		"Entertainment": {
			"entertainment", "media", "film", "music", "gaming", "sports", "recreation",
			"leisure", "amusement", "theater", "cinema", "studio", "production",
		},
		"Manufacturing": {
			"manufacturing", "factory", "production", "industrial", "machinery",
			"equipment", "assembly", "fabrication", "processing", "industrial",
		},
		"Professional Services": {
			"consulting", "legal", "accounting", "advisory", "professional",
			"service", "agency", "firm", "partners", "associates", "group",
		},
	}

	return nil
}

// initializeIndustryCodes initializes industry code mappings
func (m *KeywordClassificationModule) initializeIndustryCodes() error {
	m.industryCodes = map[string]string{
		"Grocery & Food Retail": "445110",
		"Financial Services":    "522110",
		"Healthcare":            "621111",
		"Technology":            "511210",
		"Real Estate":           "531210",
		"Transportation":        "484110",
		"Education":             "611110",
		"Entertainment":         "711110",
		"Manufacturing":         "332996",
		"Professional Services": "541611",
	}

	return nil
}

// Event emission function (will be injected by the module manager)
var emitEvent func(architecture.Event) error

// SetEventEmitter sets the event emission function
func (m *KeywordClassificationModule) SetEventEmitter(emitter func(architecture.Event) error) {
	emitEvent = emitter
}
