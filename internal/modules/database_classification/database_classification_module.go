package database_classification

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"kyb-platform/internal/architecture"
	"kyb-platform/internal/classification"
	"kyb-platform/internal/database"
	"kyb-platform/internal/shared"
)

// DatabaseClassificationModule implements the architecture.Module interface
// for the new database-driven classification system
type DatabaseClassificationModule struct {
	id                    string
	classificationService *classification.IntegrationService
	logger                *log.Logger
	config                *Config
	metadata              architecture.ModuleMetadata
	status                architecture.ModuleStatus
	startTime             time.Time
}

// Config holds configuration for the database classification module
type Config struct {
	ModuleID          string        `json:"module_id"`
	ModuleName        string        `json:"module_name"`
	ModuleVersion     string        `json:"module_version"`
	ModuleDescription string        `json:"module_description"`
	RequestTimeout    time.Duration `json:"request_timeout"`
	MaxConcurrency    int           `json:"max_concurrency"`
	EnableCaching     bool          `json:"enable_caching"`
	CacheTTL          time.Duration `json:"cache_ttl"`
}

// DefaultConfig returns the default configuration for the module
func DefaultConfig() *Config {
	return &Config{
		ModuleID:          "database_classification",
		ModuleName:        "Database-Driven Classification Module",
		ModuleVersion:     "1.0.0",
		ModuleDescription: "Database-driven business classification using Supabase",
		RequestTimeout:    30 * time.Second,
		MaxConcurrency:    10,
		EnableCaching:     false,
		CacheTTL:          5 * time.Minute,
	}
}

// NewDatabaseClassificationModule creates a new database classification module
func NewDatabaseClassificationModule(
	supabaseClient *database.SupabaseClient,
	logger *log.Logger,
	config *Config,
) (*DatabaseClassificationModule, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if logger == nil {
		logger = log.Default()
	}

	// Create the classification service
	classificationService := classification.NewIntegrationService(supabaseClient, logger)

	// Create module metadata
	metadata := architecture.ModuleMetadata{
		Name:        config.ModuleName,
		Version:     config.ModuleVersion,
		Description: config.ModuleDescription,
		Capabilities: []architecture.ModuleCapability{
			architecture.CapabilityDataExtraction,
			architecture.CapabilityClassification,
			architecture.CapabilityWebAnalysis,
		},
		Priority: architecture.PriorityHigh,
		Tags:     []string{"classification", "database", "supabase", "business-verification"},
	}

	module := &DatabaseClassificationModule{
		id:                    config.ModuleID,
		classificationService: classificationService,
		logger:                logger,
		config:                config,
		metadata:              metadata,
		status:                architecture.ModuleStatusHealthy,
		startTime:             time.Now(),
	}

	logger.Printf("✅ Database Classification Module initialized: %s v%s", config.ModuleName, config.ModuleVersion)

	return module, nil
}

// ID returns the module ID
func (m *DatabaseClassificationModule) ID() string {
	return m.id
}

// Metadata returns the module metadata
func (m *DatabaseClassificationModule) Metadata() architecture.ModuleMetadata {
	return m.metadata
}

// IsRunning returns whether the module is running
func (m *DatabaseClassificationModule) IsRunning() bool {
	return m.status == architecture.ModuleStatusHealthy
}

// Start starts the module
func (m *DatabaseClassificationModule) Start(ctx context.Context) error {
	m.logger.Printf("🚀 Starting Database Classification Module: %s", m.id)
	m.status = architecture.ModuleStatusHealthy
	m.startTime = time.Now()
	return nil
}

// Stop stops the module
func (m *DatabaseClassificationModule) Stop(ctx context.Context) error {
	m.logger.Printf("🛑 Stopping Database Classification Module: %s", m.id)
	m.status = architecture.ModuleStatusStopped
	return nil
}

// Process processes a module request
func (m *DatabaseClassificationModule) Process(ctx context.Context, req architecture.ModuleRequest) (architecture.ModuleResponse, error) {
	startTime := time.Now()

	m.logger.Printf("🔍 Processing classification request: %s", req.ID)

	// Create timeout context
	processCtx, cancel := context.WithTimeout(ctx, m.config.RequestTimeout)
	defer cancel()

	// Parse the request data
	var businessReq shared.BusinessClassificationRequest
	if req.Data != nil {
		// Convert map to JSON bytes first
		dataBytes, err := json.Marshal(req.Data)
		if err != nil {
			m.logger.Printf("❌ Failed to marshal request data: %v", err)
			return architecture.ModuleResponse{
				ID:      req.ID,
				Success: false,
				Error:   fmt.Sprintf("failed to marshal request data: %v", err),
			}, err
		}

		// Then unmarshal to struct
		if err := json.Unmarshal(dataBytes, &businessReq); err != nil {
			m.logger.Printf("❌ Failed to parse request data: %v", err)
			return architecture.ModuleResponse{
				ID:      req.ID,
				Success: false,
				Error:   fmt.Sprintf("failed to parse request data: %v", err),
			}, err
		}
	}

	// Process the business classification
	rawResult, err := m.classificationService.ProcessBusinessClassification(
		processCtx,
		businessReq.BusinessName,
		businessReq.Description,
		businessReq.WebsiteURL,
	)
	if err != nil {
		return architecture.ModuleResponse{
			ID:      req.ID,
			Success: false,
			Data: map[string]interface{}{
				"error": fmt.Sprintf("classification failed: %v", err),
			},
			Confidence: 0.0,
			Latency:    time.Since(startTime),
			Metadata: map[string]interface{}{
				"processing_time_ms": time.Since(startTime).Milliseconds(),
				"module_id":          m.id,
				"module_type":        "database_classification",
				"error":              err.Error(),
			},
		}, err
	}

	// Convert MultiMethodClassificationResult to BusinessClassificationResponse
	responseData := rawResult

	// Extract website keywords from the actual scraped content, not just the domain name
	var websiteKeywords []string
	if businessReq.WebsiteURL != "" {
		// Use keywords from the primary classification result
		websiteKeywords = responseData.PrimaryClassification.Keywords

		// Fallback: if no keywords from content, extract from domain name
		if len(websiteKeywords) == 0 {
			cleanURL := strings.TrimPrefix(businessReq.WebsiteURL, "https://")
			cleanURL = strings.TrimPrefix(cleanURL, "http://")
			cleanURL = strings.TrimPrefix(cleanURL, "www.")

			parts := strings.Split(cleanURL, ".")
			if len(parts) > 0 {
				domainWords := strings.Fields(strings.ReplaceAll(parts[0], "-", " "))
				for _, word := range domainWords {
					if len(word) > 2 {
						websiteKeywords = append(websiteKeywords, strings.ToLower(word))
					}
				}
			}
		}
	}

	// Convert MultiMethodClassificationResult to the expected response format
	var classifications []shared.IndustryClassification
	var primaryClassification *shared.IndustryClassification

	// Use the primary classification from the multi-method result
	primaryClassification = responseData.PrimaryClassification

	// Add to classifications list
	classifications = append(classifications, *primaryClassification)

	// Create classification codes from the result if available
	var classificationCodes shared.ClassificationCodes
	if primaryClassification.Metadata != nil {
		if codes, exists := primaryClassification.Metadata["classification_codes"]; exists {
			if codesMap, ok := codes.(shared.ClassificationCodes); ok {
				classificationCodes = codesMap
			}
		}
	}

	// If no codes found, create empty structure
	if classificationCodes.MCC == nil {
		classificationCodes = shared.ClassificationCodes{
			MCC:   []shared.MCCCode{},
			NAICS: []shared.NAICSCode{},
			SIC:   []shared.SICCode{},
		}
	}

	// Create the response with the correct structure for the frontend
	response := &shared.BusinessClassificationResponse{
		ID:                    businessReq.ID,
		BusinessName:          businessReq.BusinessName,
		DetectedIndustry:      responseData.PrimaryClassification.IndustryName,
		Confidence:            responseData.EnsembleConfidence,
		Classifications:       classifications,
		PrimaryClassification: primaryClassification,
		ClassificationCodes:   classificationCodes,
		OverallConfidence:     responseData.EnsembleConfidence,
		ClassificationMethod:  "multi_method_ensemble",
		ProcessingTime:        responseData.ProcessingTime,
		ModuleResults:         make(map[string]shared.ModuleResult),
		RawData: map[string]interface{}{
			"multi_method_result": responseData,
			"method_results":      responseData.MethodResults,
			"quality_metrics":     responseData.QualityMetrics,
			"reasoning":           responseData.ClassificationReasoning,
		},
		CreatedAt: responseData.CreatedAt,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"website_keywords":    websiteKeywords,
			"module_id":           m.id,
			"module_type":         "multi_method_classification",
			"processing_time_ms":  time.Since(startTime).Milliseconds(),
			"method_count":        len(responseData.MethodResults),
			"ensemble_confidence": responseData.EnsembleConfidence,
		},
	}

	m.logger.Printf("✅ Classification completed successfully: %s (%dms)",
		businessReq.BusinessName, time.Since(startTime).Milliseconds())

	return architecture.ModuleResponse{
		ID:      req.ID,
		Success: true,
		Data: map[string]interface{}{
			"response": response,
		},
		Confidence: responseData.EnsembleConfidence,
		Latency:    time.Since(startTime),
		Metadata: map[string]interface{}{
			"processing_time_ms": time.Since(startTime).Milliseconds(),
			"module_id":          m.id,
			"module_type":        "multi_method_classification",
		},
	}, nil
}

// GetConfig returns the module configuration
func (m *DatabaseClassificationModule) GetConfig() *Config {
	return m.config
}

// GetClassificationService returns the underlying classification service
func (m *DatabaseClassificationModule) GetClassificationService() *classification.IntegrationService {
	return m.classificationService
}

// Config returns the module configuration
func (m *DatabaseClassificationModule) Config() architecture.ModuleConfig {
	return architecture.ModuleConfig{
		Enabled:    true,
		Timeout:    m.config.RequestTimeout,
		RetryCount: 3,
		Parameters: map[string]interface{}{
			"module_id":       m.config.ModuleID,
			"module_name":     m.config.ModuleName,
			"module_version":  m.config.ModuleVersion,
			"max_concurrency": m.config.MaxConcurrency,
			"enable_caching":  m.config.EnableCaching,
			"cache_ttl":       m.config.CacheTTL.String(),
		},
		Dependencies: []string{"supabase"},
	}
}

// Health returns the module health status
func (m *DatabaseClassificationModule) Health() architecture.ModuleHealth {
	return architecture.ModuleHealth{
		Status:      m.status,
		LastCheck:   time.Now(),
		ErrorCount:  0,   // TODO: Track error count
		SuccessRate: 1.0, // TODO: Track success rate
		Latency:     time.Since(m.startTime),
		Message:     "Database classification module is running",
	}
}

// HealthCheck performs a health check on the module
func (m *DatabaseClassificationModule) HealthCheck(ctx context.Context) error {
	// For now, just check if the module is running
	if !m.IsRunning() {
		m.status = architecture.ModuleStatusUnhealthy
		return fmt.Errorf("module is not running")
	}

	// TODO: Add more comprehensive health checks
	m.status = architecture.ModuleStatusHealthy
	return nil
}

// CanHandle determines if this module can handle the given request
func (m *DatabaseClassificationModule) CanHandle(req architecture.ModuleRequest) bool {
	// This module can handle any business classification request
	// Check if the request has the required fields
	if req.Data == nil {
		return false
	}

	// Check for business name or description
	businessName, hasName := req.Data["business_name"].(string)
	description, hasDescription := req.Data["description"].(string)

	return hasName && businessName != "" || hasDescription && description != ""
}

// OnEvent handles module events
func (m *DatabaseClassificationModule) OnEvent(event architecture.ModuleEvent) error {
	m.logger.Printf("📢 Received event: %s for module %s", event.Type, m.id)

	// Handle different event types
	switch event.Type {
	case "shutdown":
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return m.Stop(ctx)
	case "health_check":
		return m.HealthCheck(context.Background())
	default:
		// Log unknown events but don't fail
		m.logger.Printf("⚠️ Unknown event type: %s", event.Type)
	}

	return nil
}

// convertRawResultToBusinessClassificationResponse converts the raw result from the integration service
// to a proper BusinessClassificationResponse
func (m *DatabaseClassificationModule) convertRawResultToBusinessClassificationResponse(
	rawResult map[string]interface{},
	businessReq shared.BusinessClassificationRequest,
	startTime time.Time,
) (*shared.BusinessClassificationResponse, error) {
	// Extract classification data
	classificationData, ok := rawResult["classification_data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid classification data format")
	}

	// Extract industry detection data
	industryDetection, ok := classificationData["industry_detection"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid industry detection data format")
	}

	// Extract basic fields
	detectedIndustry, _ := industryDetection["detected_industry"].(string)
	confidence, _ := industryDetection["confidence"].(float64)
	keywordsMatched, _ := industryDetection["keywords_matched"].([]string)

	// Extract classification codes
	var classificationCodes shared.ClassificationCodes
	if codesData, ok := classificationData["classification_codes"].(map[string]interface{}); ok {
		// Convert MCC codes
		if mccData, ok := codesData["mcc"].([]interface{}); ok {
			for _, code := range mccData {
				if codeMap, ok := code.(map[string]interface{}); ok {
					classificationCodes.MCC = append(classificationCodes.MCC, shared.MCCCode{
						Code:        getString(codeMap, "code"),
						Description: getString(codeMap, "description"),
						Confidence:  getFloat64(codeMap, "confidence"),
					})
				}
			}
		}

		// Convert SIC codes
		if sicData, ok := codesData["sic"].([]interface{}); ok {
			for _, code := range sicData {
				if codeMap, ok := code.(map[string]interface{}); ok {
					classificationCodes.SIC = append(classificationCodes.SIC, shared.SICCode{
						Code:        getString(codeMap, "code"),
						Description: getString(codeMap, "description"),
						Confidence:  getFloat64(codeMap, "confidence"),
					})
				}
			}
		}

		// Convert NAICS codes
		if naicsData, ok := codesData["naics"].([]interface{}); ok {
			for _, code := range naicsData {
				if codeMap, ok := code.(map[string]interface{}); ok {
					classificationCodes.NAICS = append(classificationCodes.NAICS, shared.NAICSCode{
						Code:        getString(codeMap, "code"),
						Description: getString(codeMap, "description"),
						Confidence:  getFloat64(codeMap, "confidence"),
					})
				}
			}
		}
	}

	// Extract website keywords from the request
	var websiteKeywords []string
	if businessReq.WebsiteURL != "" {
		// Extract keywords from website URL (domain name)
		cleanURL := strings.TrimPrefix(businessReq.WebsiteURL, "https://")
		cleanURL = strings.TrimPrefix(cleanURL, "http://")
		cleanURL = strings.TrimPrefix(cleanURL, "www.")

		parts := strings.Split(cleanURL, ".")
		if len(parts) > 0 {
			domainWords := strings.Fields(strings.ReplaceAll(parts[0], "-", " "))
			for _, word := range domainWords {
				if len(word) > 2 {
					websiteKeywords = append(websiteKeywords, strings.ToLower(word))
				}
			}
		}
	}

	// Create the response
	response := &shared.BusinessClassificationResponse{
		ID:                  businessReq.ID,
		BusinessName:        businessReq.BusinessName,
		DetectedIndustry:    detectedIndustry,
		Confidence:          confidence,
		ClassificationCodes: classificationCodes,
		ProcessingTime:      time.Since(startTime),
		Timestamp:           time.Now(),
		CreatedAt:           time.Now(),
		Metadata: map[string]interface{}{
			"module_id":          m.id,
			"module_type":        "database_classification",
			"keywords_used":      keywordsMatched,
			"website_keywords":   websiteKeywords,
			"processing_time_ms": time.Since(startTime).Milliseconds(),
			"raw_result":         rawResult,
		},
	}

	return response, nil
}

// Helper functions for safe type conversion
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0.0
}

// convertMCCCodes converts internal MCC codes to shared format
func convertMCCCodes(codes []classification.MCCCode) []shared.MCCCode {
	result := make([]shared.MCCCode, len(codes))
	for i, code := range codes {
		result[i] = shared.MCCCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
		}
	}
	return result
}

// convertSICCodes converts internal SIC codes to shared format
func convertSICCodes(codes []classification.SICCode) []shared.SICCode {
	result := make([]shared.SICCode, len(codes))
	for i, code := range codes {
		result[i] = shared.SICCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
		}
	}
	return result
}

// convertNAICSCodes converts internal NAICS codes to shared format
func convertNAICSCodes(codes []classification.NAICSCode) []shared.NAICSCode {
	result := make([]shared.NAICSCode, len(codes))
	for i, code := range codes {
		result[i] = shared.NAICSCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
		}
	}
	return result
}
