package website_analysis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/architecture"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/html"
)

// WebsiteAnalysisModule implements the Module interface for website analysis
type WebsiteAnalysisModule struct {
	id        string
	config    architecture.ModuleConfig
	running   bool
	logger    *observability.Logger
	metrics   *observability.Metrics
	tracer    trace.Tracer
	db        database.Database
	appConfig *config.Config

	// Website analysis specific fields
	webScraper          *WebScraper
	contentAnalyzer     *ContentAnalyzer
	semanticAnalyzer    *SemanticAnalyzer
	pageTypeDetector    *PageTypeDetector
	pageDiscovery       *PageDiscovery
	connectionValidator *ConnectionValidator

	// Caching
	resultCache map[string]*WebsiteAnalysisResult
	cacheMutex  sync.RWMutex
	cacheTTL    time.Duration

	// Performance tracking
	analysisTimes map[string]time.Duration
	successRates  map[string]float64
	metricsMutex  sync.RWMutex

	// Configuration
	scrapingConfig ScrapingConfig
	analysisConfig AnalysisConfig
}

// NewWebsiteAnalysisModule creates a new website analysis module
func NewWebsiteAnalysisModule() *WebsiteAnalysisModule {
	return &WebsiteAnalysisModule{
		id: "website_analysis_module",

		// Initialize caching
		resultCache: make(map[string]*WebsiteAnalysisResult),
		cacheTTL:    2 * time.Hour, // Longer cache for website analysis

		// Initialize performance tracking
		analysisTimes: make(map[string]time.Duration),
		successRates:  make(map[string]float64),

		// Initialize configuration
		scrapingConfig: ScrapingConfig{
			Timeout:         30 * time.Second,
			MaxRetries:      3,
			RetryDelay:      2 * time.Second,
			MaxConcurrent:   5,
			RateLimitPerSec: 2,
			UserAgents: []string{
				"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
				"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
				"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
			},
		},
		analysisConfig: AnalysisConfig{
			MaxPages:               5,
			ContentMinLength:       100,
			QualityThreshold:       0.6,
			EnableMetaTags:         true,
			EnableStructuredData:   true,
			EnableSemanticAnalysis: true,
		},
	}
}

// Module interface implementation
func (m *WebsiteAnalysisModule) ID() string {
	return m.id
}

func (m *WebsiteAnalysisModule) Config() architecture.ModuleConfig {
	return m.config
}

func (m *WebsiteAnalysisModule) UpdateConfig(config architecture.ModuleConfig) error {
	m.config = config
	return nil
}

func (m *WebsiteAnalysisModule) Health() architecture.ModuleHealth {
	status := architecture.ModuleStatusStopped
	if m.running {
		status = architecture.ModuleStatusRunning
	}

	return architecture.ModuleHealth{
		Status:    status,
		LastCheck: time.Now(),
		Message:   "Website analysis module health check",
	}
}

func (m *WebsiteAnalysisModule) Metadata() architecture.ModuleMetadata {
	return architecture.ModuleMetadata{
		Name:        "Website Analysis Module",
		Version:     "1.0.0",
		Description: "Performs comprehensive website analysis and content extraction",
		Capabilities: []architecture.ModuleCapability{
			architecture.CapabilityClassification,
			architecture.CapabilityWebAnalysis,
			architecture.CapabilityDataExtraction,
		},
		Priority: architecture.PriorityHigh,
	}
}

func (m *WebsiteAnalysisModule) Start(ctx context.Context) error {
	_, span := m.tracer.Start(ctx, "WebsiteAnalysisModule.Start")
	defer span.End()

	if m.running {
		return fmt.Errorf("module %s is already running", m.id)
	}

	// Initialize website analysis components
	if err := m.initializeComponents(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to initialize components: %w", err)
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
				"start_time": time.Now(),
			},
		})
	}

	span.SetAttributes(attribute.String("module.id", m.id))
	m.logger.WithComponent("website_analysis_module").Info("Module started", map[string]interface{}{
		"module_id": m.id,
	})

	return nil
}

func (m *WebsiteAnalysisModule) Stop(ctx context.Context) error {
	_, span := m.tracer.Start(ctx, "WebsiteAnalysisModule.Stop")
	defer span.End()

	if !m.running {
		return fmt.Errorf("module %s is not running", m.id)
	}

	m.running = false

	// Emit module stopped event
	if emitEvent != nil {
		emitEvent(architecture.Event{
			Type:     architecture.EventTypeModuleStopped,
			Source:   m.id,
			Priority: architecture.EventPriorityNormal,
			Data: map[string]interface{}{
				"module_id": m.id,
				"stop_time": time.Now(),
			},
		})
	}

	span.SetAttributes(attribute.String("module.id", m.id))
	m.logger.WithComponent("website_analysis_module").Info("Module stopped", map[string]interface{}{
		"module_id": m.id,
	})

	return nil
}

func (m *WebsiteAnalysisModule) IsRunning() bool {
	return m.running
}

func (m *WebsiteAnalysisModule) Process(ctx context.Context, req architecture.ModuleRequest) (architecture.ModuleResponse, error) {
	_, span := m.tracer.Start(ctx, "WebsiteAnalysisModule.Process")
	defer span.End()

	span.SetAttributes(
		attribute.String("module.id", m.id),
		attribute.String("request.type", req.Type),
	)

	// Check if this module can handle the request
	if !m.CanHandle(req) {
		return architecture.ModuleResponse{
			ID:      req.ID,
			Success: false,
			Error:   "unsupported request type",
		}, nil
	}

	// Parse the request payload
	analysisReq, err := m.parseAnalysisRequest(req.Data)
	if err != nil {
		span.RecordError(err)
		return architecture.ModuleResponse{
			ID:      req.ID,
			Success: false,
			Error:   fmt.Sprintf("failed to parse request: %v", err),
		}, nil
	}

	// Perform website analysis
	result, err := m.performWebsiteAnalysis(ctx, analysisReq)
	if err != nil {
		span.RecordError(err)
		return architecture.ModuleResponse{
			ID:      req.ID,
			Success: false,
			Error:   fmt.Sprintf("website analysis failed: %v", err),
		}, nil
	}

	// Create response
	response := architecture.ModuleResponse{
		ID:      req.ID,
		Success: true,
		Data: map[string]interface{}{
			"analysis":  result,
			"method":    "website_analysis",
			"module_id": m.id,
		},
	}

	// Emit analysis completed event
	if emitEvent != nil {
		emitEvent(architecture.Event{
			Type:     architecture.EventTypeClassificationCompleted,
			Source:   m.id,
			Priority: architecture.EventPriorityNormal,
			Data: map[string]interface{}{
				"website_url":        analysisReq.WebsiteURL,
				"business_name":      analysisReq.BusinessName,
				"method":             "website_analysis",
				"pages_analyzed":     len(result.PageAnalysis),
				"overall_confidence": result.OverallConfidence,
			},
		})
	}

	// Record metrics
	m.metrics.RecordBusinessClassification("website_analysis", 1.0)

	return response, nil
}

func (m *WebsiteAnalysisModule) CanHandle(req architecture.ModuleRequest) bool {
	return req.Type == "analyze_website"
}

func (m *WebsiteAnalysisModule) HealthCheck(ctx context.Context) error {
	_, span := m.tracer.Start(ctx, "WebsiteAnalysisModule.HealthCheck")
	defer span.End()

	if !m.running {
		return fmt.Errorf("module is not running")
	}

	// Check if components are initialized
	if m.webScraper == nil {
		return fmt.Errorf("web scraper not initialized")
	}

	if m.contentAnalyzer == nil {
		return fmt.Errorf("content analyzer not initialized")
	}

	span.SetAttributes(attribute.String("module.id", m.id))
	return nil
}

func (m *WebsiteAnalysisModule) OnEvent(event architecture.ModuleEvent) error {
	// Handle events if needed
	return nil
}

// Website analysis specific methods

// AnalysisRequest represents a website analysis request
type AnalysisRequest struct {
	BusinessName          string                 `json:"business_name"`
	WebsiteURL            string                 `json:"website_url"`
	MaxPages              int                    `json:"max_pages"`
	IncludeMeta           bool                   `json:"include_meta"`
	IncludeStructuredData bool                   `json:"include_structured_data"`
	Metadata              map[string]interface{} `json:"metadata"`
}

// WebsiteAnalysisResult represents comprehensive website analysis results
type WebsiteAnalysisResult struct {
	URL                    string                         `json:"url"`
	BusinessName           string                         `json:"business_name"`
	ConnectionValidation   *ConnectionValidationResult    `json:"connection_validation"`
	ContentAnalysis        *ContentAnalysisResult         `json:"content_analysis"`
	SemanticAnalysis       *SemanticAnalysisResult        `json:"semantic_analysis"`
	IndustryClassification []IndustryClassificationResult `json:"industry_classification"`
	PageAnalysis           []PageAnalysisResult           `json:"page_analysis"`
	OverallConfidence      float64                        `json:"overall_confidence"`
	AnalysisTime           time.Time                      `json:"analysis_time"`
	AnalysisMetadata       map[string]interface{}         `json:"analysis_metadata"`
}

// ConnectionValidationResult represents connection validation results
type ConnectionValidationResult struct {
	IsValid          bool     `json:"is_valid"`
	Confidence       float64  `json:"confidence"`
	ValidationMethod string   `json:"validation_method"`
	BusinessMatch    bool     `json:"business_match"`
	DomainAge        int      `json:"domain_age"`
	SSLValid         bool     `json:"ssl_valid"`
	ValidationErrors []string `json:"validation_errors"`
}

// ContentAnalysisResult represents content analysis results
type ContentAnalysisResult struct {
	ContentQuality     float64                `json:"content_quality"`
	ContentLength      int                    `json:"content_length"`
	MetaTags           map[string]string      `json:"meta_tags"`
	StructuredData     map[string]interface{} `json:"structured_data"`
	IndustryIndicators []string               `json:"industry_indicators"`
	BusinessKeywords   []string               `json:"business_keywords"`
	ContentType        string                 `json:"content_type"`
}

// SemanticAnalysisResult represents semantic analysis results
type SemanticAnalysisResult struct {
	SemanticScore    float64            `json:"semantic_score"`
	TopicModeling    map[string]float64 `json:"topic_modeling"`
	SentimentScore   float64            `json:"sentiment_score"`
	KeyPhrases       []string           `json:"key_phrases"`
	EntityExtraction map[string]string  `json:"entity_extraction"`
}

// IndustryClassificationResult represents industry classification results
type IndustryClassificationResult struct {
	IndustryCode string   `json:"industry_code"`
	IndustryName string   `json:"industry_name"`
	Confidence   float64  `json:"confidence"`
	Keywords     []string `json:"keywords"`
	Evidence     string   `json:"evidence"`
}

// PageAnalysisResult represents analysis results for individual pages
type PageAnalysisResult struct {
	URL            string                         `json:"url"`
	PageType       string                         `json:"page_type"`
	ContentQuality float64                        `json:"content_quality"`
	RelevanceScore float64                        `json:"relevance_score"`
	PriorityScore  float64                        `json:"priority_score"`
	Classification []IndustryClassificationResult `json:"classification"`
	AnalysisTime   time.Time                      `json:"analysis_time"`
}

// ScrapedContent represents scraped website content
type ScrapedContent struct {
	URL           string            `json:"url"`
	Title         string            `json:"title"`
	HTML          string            `json:"html"`
	Text          string            `json:"text"`
	Headers       map[string]string `json:"headers"`
	StatusCode    int               `json:"status_code"`
	ResponseTime  time.Duration     `json:"response_time"`
	ExtractedData map[string]string `json:"extracted_data"`
	Error         string            `json:"error,omitempty"`
}

// Configuration types
type ScrapingConfig struct {
	Timeout         time.Duration `json:"timeout"`
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	MaxConcurrent   int           `json:"max_concurrent"`
	RateLimitPerSec int           `json:"rate_limit_per_sec"`
	UserAgents      []string      `json:"user_agents"`
}

type AnalysisConfig struct {
	MaxPages               int     `json:"max_pages"`
	ContentMinLength       int     `json:"content_min_length"`
	QualityThreshold       float64 `json:"quality_threshold"`
	EnableMetaTags         bool    `json:"enable_meta_tags"`
	EnableStructuredData   bool    `json:"enable_structured_data"`
	EnableSemanticAnalysis bool    `json:"enable_semantic_analysis"`
}

// Component interfaces
type WebScraper struct {
	config ScrapingConfig
	client *http.Client
}

type ContentAnalyzer struct {
	config AnalysisConfig
}

type SemanticAnalyzer struct {
	config AnalysisConfig
}

type PageTypeDetector struct {
	config AnalysisConfig
}

type PageDiscovery struct {
	config AnalysisConfig
}

type ConnectionValidator struct {
	config AnalysisConfig
}

// parseAnalysisRequest parses the module request into an analysis request
func (m *WebsiteAnalysisModule) parseAnalysisRequest(payload map[string]interface{}) (*AnalysisRequest, error) {
	req := &AnalysisRequest{}

	if businessName, ok := payload["business_name"].(string); ok {
		req.BusinessName = businessName
	}

	if websiteURL, ok := payload["website_url"].(string); ok {
		req.WebsiteURL = websiteURL
	}

	if maxPages, ok := payload["max_pages"].(float64); ok {
		req.MaxPages = int(maxPages)
	} else {
		req.MaxPages = m.analysisConfig.MaxPages
	}

	if includeMeta, ok := payload["include_meta"].(bool); ok {
		req.IncludeMeta = includeMeta
	} else {
		req.IncludeMeta = m.analysisConfig.EnableMetaTags
	}

	if includeStructuredData, ok := payload["include_structured_data"].(bool); ok {
		req.IncludeStructuredData = includeStructuredData
	} else {
		req.IncludeStructuredData = m.analysisConfig.EnableStructuredData
	}

	if metadata, ok := payload["metadata"].(map[string]interface{}); ok {
		req.Metadata = metadata
	} else {
		req.Metadata = make(map[string]interface{})
	}

	return req, nil
}

// performWebsiteAnalysis performs comprehensive website analysis
func (m *WebsiteAnalysisModule) performWebsiteAnalysis(ctx context.Context, req *AnalysisRequest) (*WebsiteAnalysisResult, error) {
	_, span := m.tracer.Start(ctx, "performWebsiteAnalysis")
	defer span.End()

	span.SetAttributes(
		attribute.String("website_url", req.WebsiteURL),
		attribute.String("business_name", req.BusinessName),
	)

	// Check cache first
	cacheKey := m.generateCacheKey(req)
	if cached, exists := m.getFromCache(cacheKey); exists {
		span.AddEvent("Cache hit")
		return cached, nil
	}

	startTime := time.Now()

	// Step 1: Scrape website content
	scrapedContent, err := m.scrapeWebsite(ctx, req.WebsiteURL)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to scrape website: %w", err)
	}

	// Step 2: Perform connection validation
	connectionValidation, err := m.validateConnection(ctx, req.BusinessName, req.WebsiteURL, scrapedContent)
	if err != nil {
		// Log error but continue with analysis
		m.logger.WithComponent("website_analysis_module").Warn("connection_validation_failed", map[string]interface{}{
			"website_url": req.WebsiteURL,
			"error":       err.Error(),
		})
		connectionValidation = m.createMinimalValidationResult(req.BusinessName, req.WebsiteURL)
	}

	// Step 3: Perform content analysis
	contentAnalysis, err := m.analyzeContent(ctx, scrapedContent, req)
	if err != nil {
		span.RecordError(err)
		contentAnalysis = m.createMinimalContentAnalysis(scrapedContent)
	}

	// Step 4: Perform semantic analysis
	semanticAnalysis, err := m.analyzeSemantic(ctx, scrapedContent, req.BusinessName)
	if err != nil {
		span.RecordError(err)
		semanticAnalysis = m.createMinimalSemanticAnalysis(scrapedContent)
	}

	// Step 5: Perform industry classification
	industryClassification, err := m.classifyIndustries(ctx, scrapedContent, req.BusinessName)
	if err != nil {
		span.RecordError(err)
		industryClassification = m.createMinimalIndustryClassification(req.BusinessName)
	}

	// Step 6: Perform page analysis
	pageAnalysis, err := m.analyzePages(ctx, req.WebsiteURL, req.BusinessName, req.MaxPages)
	if err != nil {
		span.RecordError(err)
		pageAnalysis = []PageAnalysisResult{}
	}

	// Step 7: Calculate overall confidence
	overallConfidence := m.calculateOverallConfidence(
		connectionValidation, contentAnalysis, semanticAnalysis, industryClassification)

	// Create result
	result := &WebsiteAnalysisResult{
		URL:                    req.WebsiteURL,
		BusinessName:           req.BusinessName,
		ConnectionValidation:   connectionValidation,
		ContentAnalysis:        contentAnalysis,
		SemanticAnalysis:       semanticAnalysis,
		IndustryClassification: industryClassification,
		PageAnalysis:           pageAnalysis,
		OverallConfidence:      overallConfidence,
		AnalysisTime:           time.Now(),
		AnalysisMetadata: map[string]interface{}{
			"analysis_duration": time.Since(startTime).Milliseconds(),
			"pages_analyzed":    len(pageAnalysis),
			"cache_key":         cacheKey,
		},
	}

	// Cache the result
	m.cacheResult(cacheKey, result)

	// Update performance metrics
	m.updatePerformanceMetrics(req.WebsiteURL, time.Since(startTime))

	span.SetAttributes(
		attribute.Float64("overall_confidence", overallConfidence),
		attribute.Int("pages_analyzed", len(pageAnalysis)),
		attribute.Int64("analysis_duration_ms", time.Since(startTime).Milliseconds()),
	)

	return result, nil
}

// generateCacheKey generates a cache key for the request
func (m *WebsiteAnalysisModule) generateCacheKey(req *AnalysisRequest) string {
	data := fmt.Sprintf("%s|%s|%d|%t|%t",
		req.BusinessName,
		req.WebsiteURL,
		req.MaxPages,
		req.IncludeMeta,
		req.IncludeStructuredData,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// getFromCache retrieves a result from cache
func (m *WebsiteAnalysisModule) getFromCache(cacheKey string) (*WebsiteAnalysisResult, bool) {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	if result, exists := m.resultCache[cacheKey]; exists {
		// Check if cache entry is still valid
		if time.Since(result.AnalysisTime) < m.cacheTTL {
			return result, true
		}
		// Remove expired entry
		delete(m.resultCache, cacheKey)
	}

	return nil, false
}

// cacheResult stores a result in cache
func (m *WebsiteAnalysisModule) cacheResult(cacheKey string, result *WebsiteAnalysisResult) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	m.resultCache[cacheKey] = result
}

// scrapeWebsite scrapes the main website
func (m *WebsiteAnalysisModule) scrapeWebsite(ctx context.Context, url string) (*ScrapedContent, error) {
	_, span := m.tracer.Start(ctx, "scrapeWebsite")
	defer span.End()

	span.SetAttributes(attribute.String("url", url))

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: m.scrapingConfig.Timeout,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", m.getRandomUserAgent())

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch website: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse HTML and extract text content
	htmlContent := string(body)
	textContent := m.extractTextFromHTML(htmlContent)

	// Extract title
	title := m.extractTitleFromHTML(htmlContent)

	// Create scraped content
	content := &ScrapedContent{
		URL:           url,
		Title:         title,
		HTML:          htmlContent,
		Text:          textContent,
		StatusCode:    resp.StatusCode,
		ResponseTime:  time.Since(time.Now()),
		Headers:       make(map[string]string),
		ExtractedData: make(map[string]string),
	}

	// Extract headers
	for key, values := range resp.Header {
		if len(values) > 0 {
			content.Headers[key] = values[0]
		}
	}

	span.SetAttributes(
		attribute.Int("status_code", resp.StatusCode),
		attribute.Int("content_length", len(textContent)),
		attribute.String("title", title),
	)

	return content, nil
}

// validateConnection validates business-website connection
func (m *WebsiteAnalysisModule) validateConnection(ctx context.Context, businessName, websiteURL string, content *ScrapedContent) (*ConnectionValidationResult, error) {
	_, span := m.tracer.Start(ctx, "validateConnection")
	defer span.End()

	// Simple validation logic
	isValid := true
	confidence := 0.8
	validationErrors := []string{}

	// Check if business name appears in content
	businessNameLower := strings.ToLower(businessName)
	contentLower := strings.ToLower(content.Text)

	if !strings.Contains(contentLower, businessNameLower) {
		isValid = false
		confidence -= 0.3
		validationErrors = append(validationErrors, "Business name not found in website content")
	}

	// Check if website is accessible
	if content.StatusCode != 200 {
		isValid = false
		confidence -= 0.2
		validationErrors = append(validationErrors, fmt.Sprintf("Website returned status code: %d", content.StatusCode))
	}

	// Check content quality
	if len(content.Text) < m.analysisConfig.ContentMinLength {
		isValid = false
		confidence -= 0.1
		validationErrors = append(validationErrors, "Insufficient content length")
	}

	// Ensure confidence is within bounds
	if confidence < 0 {
		confidence = 0
	}

	result := &ConnectionValidationResult{
		IsValid:          isValid,
		Confidence:       confidence,
		ValidationMethod: "content_analysis",
		BusinessMatch:    strings.Contains(contentLower, businessNameLower),
		DomainAge:        365,  // Placeholder
		SSLValid:         true, // Placeholder
		ValidationErrors: validationErrors,
	}

	span.SetAttributes(
		attribute.Bool("is_valid", isValid),
		attribute.Float64("confidence", confidence),
		attribute.Bool("business_match", result.BusinessMatch),
	)

	return result, nil
}

// analyzeContent performs content analysis
func (m *WebsiteAnalysisModule) analyzeContent(ctx context.Context, content *ScrapedContent, req *AnalysisRequest) (*ContentAnalysisResult, error) {
	_, span := m.tracer.Start(ctx, "analyzeContent")
	defer span.End()

	analysis := &ContentAnalysisResult{
		ContentQuality:     0.8,
		ContentLength:      len(content.Text),
		MetaTags:           make(map[string]string),
		StructuredData:     make(map[string]interface{}),
		IndustryIndicators: []string{},
		BusinessKeywords:   []string{},
		ContentType:        "business_website",
	}

	// Extract meta tags if enabled
	if req.IncludeMeta {
		analysis.MetaTags = m.extractMetaTags(content.HTML)
	}

	// Extract structured data if enabled
	if req.IncludeStructuredData {
		analysis.StructuredData = m.extractStructuredData(content.HTML)
	}

	// Extract industry indicators
	analysis.IndustryIndicators = m.extractIndustryIndicators(content.Text)

	// Extract business keywords
	analysis.BusinessKeywords = m.extractBusinessKeywords(content.Text, req.BusinessName)

	// Calculate content quality
	analysis.ContentQuality = m.calculateContentQuality(content, analysis)

	span.SetAttributes(
		attribute.Float64("content_quality", analysis.ContentQuality),
		attribute.Int("content_length", analysis.ContentLength),
		attribute.Int("industry_indicators", len(analysis.IndustryIndicators)),
	)

	return analysis, nil
}

// analyzeSemantic performs semantic analysis
func (m *WebsiteAnalysisModule) analyzeSemantic(ctx context.Context, content *ScrapedContent, businessName string) (*SemanticAnalysisResult, error) {
	_, span := m.tracer.Start(ctx, "analyzeSemantic")
	defer span.End()

	analysis := &SemanticAnalysisResult{
		SemanticScore:    0.75,
		TopicModeling:    make(map[string]float64),
		SentimentScore:   0.6,
		KeyPhrases:       []string{},
		EntityExtraction: make(map[string]string),
	}

	// Simple semantic analysis
	text := strings.ToLower(content.Text)

	// Topic modeling (simple keyword-based)
	topics := map[string][]string{
		"technology": {"software", "technology", "digital", "platform", "system"},
		"healthcare": {"health", "medical", "care", "hospital", "patient"},
		"finance":    {"financial", "bank", "credit", "investment", "money"},
		"retail":     {"shop", "store", "retail", "product", "sale"},
	}

	for topic, keywords := range topics {
		score := 0.0
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				score += 0.2
			}
		}
		if score > 0 {
			analysis.TopicModeling[topic] = score
		}
	}

	// Extract key phrases (simple approach)
	analysis.KeyPhrases = m.extractKeyPhrases(content.Text)

	// Entity extraction
	analysis.EntityExtraction["business_name"] = businessName
	analysis.EntityExtraction["website"] = content.URL

	// Calculate semantic score
	analysis.SemanticScore = m.calculateSemanticScore(analysis)

	span.SetAttributes(
		attribute.Float64("semantic_score", analysis.SemanticScore),
		attribute.Float64("sentiment_score", analysis.SentimentScore),
		attribute.Int("key_phrases", len(analysis.KeyPhrases)),
	)

	return analysis, nil
}

// classifyIndustries performs industry classification
func (m *WebsiteAnalysisModule) classifyIndustries(ctx context.Context, content *ScrapedContent, businessName string) ([]IndustryClassificationResult, error) {
	_, span := m.tracer.Start(ctx, "classifyIndustries")
	defer span.End()

	classifications := []IndustryClassificationResult{}

	// Simple industry classification based on keywords
	text := strings.ToLower(content.Text + " " + businessName)

	industryPatterns := map[string]struct {
		code     string
		name     string
		keywords []string
	}{
		"technology": {
			code:     "511210",
			name:     "Technology",
			keywords: []string{"software", "technology", "digital", "platform", "system", "app", "web"},
		},
		"healthcare": {
			code:     "621111",
			name:     "Healthcare",
			keywords: []string{"health", "medical", "care", "hospital", "patient", "doctor", "clinic"},
		},
		"finance": {
			code:     "522110",
			name:     "Financial Services",
			keywords: []string{"financial", "bank", "credit", "investment", "money", "loan", "insurance"},
		},
		"retail": {
			code:     "445110",
			name:     "Retail",
			keywords: []string{"shop", "store", "retail", "product", "sale", "buy", "purchase"},
		},
	}

	for _, pattern := range industryPatterns {
		confidence := 0.0
		matchedKeywords := []string{}

		for _, keyword := range pattern.keywords {
			if strings.Contains(text, keyword) {
				confidence += 0.2
				matchedKeywords = append(matchedKeywords, keyword)
			}
		}

		if confidence > 0.3 {
			classifications = append(classifications, IndustryClassificationResult{
				IndustryCode: pattern.code,
				IndustryName: pattern.name,
				Confidence:   confidence,
				Keywords:     matchedKeywords,
				Evidence:     fmt.Sprintf("Matched keywords: %s", strings.Join(matchedKeywords, ", ")),
			})
		}
	}

	// Sort by confidence
	if len(classifications) > 0 {
		// Simple sorting (in a real implementation, use sort.Slice)
		// For now, just return the first one
		classifications = classifications[:1]
	}

	span.SetAttributes(
		attribute.Int("classifications_count", len(classifications)),
	)

	return classifications, nil
}

// analyzePages performs analysis of multiple pages
func (m *WebsiteAnalysisModule) analyzePages(ctx context.Context, websiteURL, businessName string, maxPages int) ([]PageAnalysisResult, error) {
	_, span := m.tracer.Start(ctx, "analyzePages")
	defer span.End()

	// For now, just analyze the main page
	// In a real implementation, this would discover and analyze multiple pages
	pageResults := []PageAnalysisResult{}

	// Analyze main page
	mainPageResult := PageAnalysisResult{
		URL:            websiteURL,
		PageType:       "home",
		ContentQuality: 0.8,
		RelevanceScore: 0.9,
		PriorityScore:  1.0,
		Classification: []IndustryClassificationResult{},
		AnalysisTime:   time.Now(),
	}

	pageResults = append(pageResults, mainPageResult)

	span.SetAttributes(
		attribute.Int("pages_analyzed", len(pageResults)),
	)

	return pageResults, nil
}

// calculateOverallConfidence calculates overall confidence
func (m *WebsiteAnalysisModule) calculateOverallConfidence(
	connectionValidation *ConnectionValidationResult,
	contentAnalysis *ContentAnalysisResult,
	semanticAnalysis *SemanticAnalysisResult,
	industryClassification []IndustryClassificationResult) float64 {

	// Weight factors
	weights := map[string]float64{
		"connection": 0.3,
		"content":    0.2,
		"semantic":   0.2,
		"industry":   0.3,
	}

	// Connection confidence
	connectionConfidence := connectionValidation.Confidence

	// Content quality confidence
	contentConfidence := contentAnalysis.ContentQuality

	// Semantic confidence
	semanticConfidence := semanticAnalysis.SemanticScore

	// Industry classification confidence
	industryConfidence := 0.0
	if len(industryClassification) > 0 {
		industryConfidence = industryClassification[0].Confidence
	}

	// Calculate weighted average
	overallConfidence := connectionConfidence*weights["connection"] +
		contentConfidence*weights["content"] +
		semanticConfidence*weights["semantic"] +
		industryConfidence*weights["industry"]

	// Normalize to 0-1 range
	if overallConfidence > 1.0 {
		overallConfidence = 1.0
	}

	return overallConfidence
}

// Helper methods

func (m *WebsiteAnalysisModule) getRandomUserAgent() string {
	if len(m.scrapingConfig.UserAgents) == 0 {
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	}
	return m.scrapingConfig.UserAgents[0] // Simplified for now
}

func (m *WebsiteAnalysisModule) extractTextFromHTML(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return ""
	}

	var textContent strings.Builder
	m.extractText(doc, &textContent)

	text := textContent.String()
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	return text
}

func (m *WebsiteAnalysisModule) extractText(n *html.Node, text *strings.Builder) {
	if n.Type == html.TextNode {
		text.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		m.extractText(c, text)
	}
}

func (m *WebsiteAnalysisModule) extractTitleFromHTML(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return ""
	}

	var title string
	m.findTitle(doc, &title)
	return title
}

func (m *WebsiteAnalysisModule) findTitle(n *html.Node, title *string) {
	if n.Type == html.ElementNode && n.Data == "title" {
		if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
			*title = n.FirstChild.Data
		}
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		m.findTitle(c, title)
	}
}

func (m *WebsiteAnalysisModule) extractMetaTags(htmlContent string) map[string]string {
	metaTags := make(map[string]string)
	// Simple meta tag extraction
	// In a real implementation, this would parse HTML properly
	return metaTags
}

func (m *WebsiteAnalysisModule) extractStructuredData(htmlContent string) map[string]interface{} {
	structuredData := make(map[string]interface{})
	// Simple structured data extraction
	// In a real implementation, this would parse JSON-LD, Microdata, etc.
	return structuredData
}

func (m *WebsiteAnalysisModule) extractIndustryIndicators(text string) []string {
	indicators := []string{}
	textLower := strings.ToLower(text)

	// Simple industry indicator extraction
	industryKeywords := []string{
		"software", "technology", "health", "medical", "financial", "bank",
		"retail", "shop", "education", "consulting", "legal", "real estate",
	}

	for _, keyword := range industryKeywords {
		if strings.Contains(textLower, keyword) {
			indicators = append(indicators, keyword)
		}
	}

	return indicators
}

func (m *WebsiteAnalysisModule) extractBusinessKeywords(text string, businessName string) []string {
	keywords := []string{}
	textLower := strings.ToLower(text)

	// Extract business name words
	businessWords := strings.Fields(strings.ToLower(businessName))
	for _, word := range businessWords {
		if len(word) > 2 && strings.Contains(textLower, word) {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

func (m *WebsiteAnalysisModule) calculateContentQuality(content *ScrapedContent, analysis *ContentAnalysisResult) float64 {
	quality := 0.5

	// Length factor
	if analysis.ContentLength > 500 {
		quality += 0.2
	} else if analysis.ContentLength > 200 {
		quality += 0.1
	}

	// Meta tags factor
	if len(analysis.MetaTags) > 0 {
		quality += 0.1
	}

	// Structured data factor
	if len(analysis.StructuredData) > 0 {
		quality += 0.1
	}

	// Industry indicators factor
	if len(analysis.IndustryIndicators) > 0 {
		quality += 0.1
	}

	return quality
}

func (m *WebsiteAnalysisModule) extractKeyPhrases(text string) []string {
	phrases := []string{}
	// Simple key phrase extraction
	// In a real implementation, this would use NLP techniques
	return phrases
}

func (m *WebsiteAnalysisModule) calculateSemanticScore(analysis *SemanticAnalysisResult) float64 {
	score := 0.5

	// Topic modeling factor
	if len(analysis.TopicModeling) > 0 {
		score += 0.2
	}

	// Key phrases factor
	if len(analysis.KeyPhrases) > 0 {
		score += 0.2
	}

	// Entity extraction factor
	if len(analysis.EntityExtraction) > 0 {
		score += 0.1
	}

	return score
}

func (m *WebsiteAnalysisModule) updatePerformanceMetrics(websiteURL string, duration time.Duration) {
	m.metricsMutex.Lock()
	defer m.metricsMutex.Unlock()

	m.analysisTimes[websiteURL] = duration
}

// Fallback methods for error cases

func (m *WebsiteAnalysisModule) createMinimalValidationResult(businessName, websiteURL string) *ConnectionValidationResult {
	return &ConnectionValidationResult{
		IsValid:          false,
		Confidence:       0.0,
		ValidationMethod: "fallback",
		BusinessMatch:    false,
		DomainAge:        0,
		SSLValid:         false,
		ValidationErrors: []string{"Validation failed, using fallback"},
	}
}

func (m *WebsiteAnalysisModule) createMinimalContentAnalysis(content *ScrapedContent) *ContentAnalysisResult {
	return &ContentAnalysisResult{
		ContentQuality:     0.3,
		ContentLength:      len(content.Text),
		MetaTags:           make(map[string]string),
		StructuredData:     make(map[string]interface{}),
		IndustryIndicators: []string{},
		BusinessKeywords:   []string{},
		ContentType:        "unknown",
	}
}

func (m *WebsiteAnalysisModule) createMinimalSemanticAnalysis(content *ScrapedContent) *SemanticAnalysisResult {
	return &SemanticAnalysisResult{
		SemanticScore:    0.3,
		TopicModeling:    make(map[string]float64),
		SentimentScore:   0.5,
		KeyPhrases:       []string{},
		EntityExtraction: make(map[string]string),
	}
}

func (m *WebsiteAnalysisModule) createMinimalIndustryClassification(businessName string) []IndustryClassificationResult {
	return []IndustryClassificationResult{
		{
			IndustryCode: "000000",
			IndustryName: "Unknown",
			Confidence:   0.1,
			Keywords:     []string{},
			Evidence:     "Fallback classification due to analysis failure",
		},
	}
}

// initializeComponents initializes website analysis components
func (m *WebsiteAnalysisModule) initializeComponents() error {
	// Initialize web scraper
	m.webScraper = &WebScraper{
		config: m.scrapingConfig,
		client: &http.Client{
			Timeout: m.scrapingConfig.Timeout,
		},
	}

	// Initialize content analyzer
	m.contentAnalyzer = &ContentAnalyzer{
		config: m.analysisConfig,
	}

	// Initialize semantic analyzer
	m.semanticAnalyzer = &SemanticAnalyzer{
		config: m.analysisConfig,
	}

	// Initialize page type detector
	m.pageTypeDetector = &PageTypeDetector{
		config: m.analysisConfig,
	}

	// Initialize page discovery
	m.pageDiscovery = &PageDiscovery{
		config: m.analysisConfig,
	}

	// Initialize connection validator
	m.connectionValidator = &ConnectionValidator{
		config: m.analysisConfig,
	}

	return nil
}

// Event emission function (will be injected by the module manager)
var emitEvent func(architecture.Event) error

// SetEventEmitter sets the event emission function
func (m *WebsiteAnalysisModule) SetEventEmitter(emitter func(architecture.Event) error) {
	emitEvent = emitter
}
