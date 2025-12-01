package classification

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/machine_learning"
)

// ClassificationMetrics tracks classification accuracy and performance metrics
type ClassificationMetrics struct {
	TotalClassifications    int64              `json:"total_classifications"`
	MLClassifications      int64              `json:"ml_classifications"`
	KeywordClassifications int64              `json:"keyword_classifications"`
	FallbackClassifications int64             `json:"fallback_classifications"`
	IndustryAccuracy       map[string]float64 `json:"industry_accuracy"` // Industry name -> accuracy percentage
	MethodAccuracy         map[string]float64 `json:"method_accuracy"`   // Method name -> accuracy percentage
	IndustryCorrect        map[string]int64   `json:"industry_correct"`  // Industry name -> correct count
	IndustryTotal          map[string]int64   `json:"industry_total"`    // Industry name -> total count
	MethodCorrect          map[string]int64   `json:"method_correct"`     // Method name -> correct count
	MethodTotal            map[string]int64   `json:"method_total"`      // Method name -> total count
	mu                     sync.RWMutex       `json:"-"`
}

// NewClassificationMetrics creates a new classification metrics tracker
func NewClassificationMetrics() *ClassificationMetrics {
	return &ClassificationMetrics{
		IndustryAccuracy:       make(map[string]float64),
		MethodAccuracy:         make(map[string]float64),
		IndustryCorrect:        make(map[string]int64),
		IndustryTotal:          make(map[string]int64),
		MethodCorrect:         make(map[string]int64),
		MethodTotal:           make(map[string]int64),
	}
}

// IndustryDetectionService provides database-driven industry classification
type IndustryDetectionService struct {
	repo                 repository.KeywordRepository
	logger               *log.Logger
	monitor              *ClassificationAccuracyMonitoring
	multiStrategyClassifier *MultiStrategyClassifier
	multiMethodClassifier   *MultiMethodClassifier  // Optional: for ML support
	useML                  bool                    // Flag to enable ML
	metrics                *ClassificationMetrics  // Classification accuracy metrics
}

// NewIndustryDetectionService creates a new industry detection service
func NewIndustryDetectionService(repo repository.KeywordRepository, logger *log.Logger) *IndustryDetectionService {
	if logger == nil {
		logger = log.Default()
	}

	return &IndustryDetectionService{
		repo:                 repo,
		logger:               logger,
		monitor:              nil, // Will be set separately if monitoring is needed
		multiStrategyClassifier: NewMultiStrategyClassifier(repo, logger),
		multiMethodClassifier:   nil,
		useML:                  false,
		metrics:                NewClassificationMetrics(),
	}
}

// NewIndustryDetectionServiceWithML creates a new industry detection service with ML support
func NewIndustryDetectionServiceWithML(
	repo repository.KeywordRepository,
	mlClassifier *machine_learning.ContentClassifier,
	pythonMLService interface{}, // *infrastructure.PythonMLService - using interface to avoid import cycle
	logger *log.Logger,
) *IndustryDetectionService {
	if logger == nil {
		logger = log.Default()
	}

	// Create MultiMethodClassifier with Python ML service support
	multiMethodClassifier := NewMultiMethodClassifier(repo, mlClassifier, logger)
	if pythonMLService != nil {
		multiMethodClassifier.SetPythonMLService(pythonMLService)
		logger.Printf("✅ IndustryDetectionService initialized with ML support (Python ML service enabled)")
	} else {
		logger.Printf("✅ IndustryDetectionService initialized with ML support (Go ML classifier only)")
	}

	return &IndustryDetectionService{
		repo:                 repo,
		logger:               logger,
		monitor:              nil,
		multiStrategyClassifier: NewMultiStrategyClassifier(repo, logger), // Keep for fallback
		multiMethodClassifier:   multiMethodClassifier,
		useML:                  true,
		metrics:                NewClassificationMetrics(),
	}
}

// NewIndustryDetectionServiceWithMonitoring creates a new industry detection service with monitoring
func NewIndustryDetectionServiceWithMonitoring(repo repository.KeywordRepository, logger *log.Logger, monitor *ClassificationAccuracyMonitoring) *IndustryDetectionService {
	if logger == nil {
		logger = log.Default()
	}

	return &IndustryDetectionService{
		repo:                 repo,
		logger:               logger,
		monitor:              monitor,
		multiStrategyClassifier: NewMultiStrategyClassifier(repo, logger),
		metrics:                NewClassificationMetrics(),
	}
}

// RecordClassification records a classification result for accuracy tracking
func (s *IndustryDetectionService) RecordClassification(
	result *IndustryDetectionResult,
	expectedIndustry string,
) {
	if s.metrics == nil {
		return
	}

	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	s.metrics.TotalClassifications++

	// Track by method
	switch result.Method {
	case "ml_distilbart", "ml", "ml_fallback":
		s.metrics.MLClassifications++
	case "keyword", "keyword_classification":
		s.metrics.KeywordClassifications++
	default:
		s.metrics.FallbackClassifications++
	}

	// Track accuracy
	isCorrect := strings.EqualFold(result.IndustryName, expectedIndustry)
	
	// Update industry metrics
	s.metrics.IndustryTotal[expectedIndustry]++
	if isCorrect {
		s.metrics.IndustryCorrect[expectedIndustry]++
	}
	
	// Calculate industry accuracy
	if total := s.metrics.IndustryTotal[expectedIndustry]; total > 0 {
		s.metrics.IndustryAccuracy[expectedIndustry] = float64(s.metrics.IndustryCorrect[expectedIndustry]) / float64(total) * 100.0
	}

	// Update method metrics
	s.metrics.MethodTotal[result.Method]++
	if isCorrect {
		s.metrics.MethodCorrect[result.Method]++
	}
	
	// Calculate method accuracy
	if total := s.metrics.MethodTotal[result.Method]; total > 0 {
		s.metrics.MethodAccuracy[result.Method] = float64(s.metrics.MethodCorrect[result.Method]) / float64(total) * 100.0
	}
}

// GetClassificationMetrics returns the current classification metrics
func (s *IndustryDetectionService) GetClassificationMetrics() *ClassificationMetrics {
	if s.metrics == nil {
		return NewClassificationMetrics()
	}

	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	// Return a copy of metrics
	metricsCopy := &ClassificationMetrics{
		TotalClassifications:     s.metrics.TotalClassifications,
		MLClassifications:         s.metrics.MLClassifications,
		KeywordClassifications:   s.metrics.KeywordClassifications,
		FallbackClassifications:   s.metrics.FallbackClassifications,
		IndustryAccuracy:          make(map[string]float64),
		MethodAccuracy:           make(map[string]float64),
		IndustryCorrect:          make(map[string]int64),
		IndustryTotal:            make(map[string]int64),
		MethodCorrect:           make(map[string]int64),
		MethodTotal:             make(map[string]int64),
	}

	// Copy maps
	for k, v := range s.metrics.IndustryAccuracy {
		metricsCopy.IndustryAccuracy[k] = v
	}
	for k, v := range s.metrics.MethodAccuracy {
		metricsCopy.MethodAccuracy[k] = v
	}
	for k, v := range s.metrics.IndustryCorrect {
		metricsCopy.IndustryCorrect[k] = v
	}
	for k, v := range s.metrics.IndustryTotal {
		metricsCopy.IndustryTotal[k] = v
	}
	for k, v := range s.metrics.MethodCorrect {
		metricsCopy.MethodCorrect[k] = v
	}
	for k, v := range s.metrics.MethodTotal {
		metricsCopy.MethodTotal[k] = v
	}

	return metricsCopy
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

// DetectIndustry performs database-driven industry detection using multi-strategy or multi-method classification
func (s *IndustryDetectionService) DetectIndustry(ctx context.Context, businessName, description, websiteURL string) (*IndustryDetectionResult, error) {
	startTime := time.Now()
	requestID := s.generateRequestID()

	s.logger.Printf("🔍 Starting industry detection for: %s (request: %s)", businessName, requestID)

	// Use MultiMethodClassifier with ML if available, otherwise use MultiStrategyClassifier
	if s.useML && s.multiMethodClassifier != nil {
		s.logger.Printf("🤖 Using MultiMethodClassifier with ML support (request: %s)", requestID)
		return s.detectIndustryWithML(ctx, businessName, description, websiteURL, startTime, requestID)
	}

	// Use multi-strategy classifier for improved accuracy (keyword-based)
	s.logger.Printf("📝 Using MultiStrategyClassifier (keyword-based) (request: %s)", requestID)
	multiResult, err := s.multiStrategyClassifier.ClassifyWithMultiStrategy(
		ctx, businessName, description, websiteURL)
	if err != nil {
		// Fallback to keyword-based classification if multi-strategy fails
		s.logger.Printf("⚠️ Multi-strategy classification failed, falling back to keyword-based: %v", err)
		return s.fallbackToKeywordClassification(ctx, businessName, description, websiteURL, startTime, requestID)
	}

	if multiResult == nil {
		s.logger.Printf("⚠️ Multi-strategy returned nil, falling back to keyword-based")
		return s.fallbackToKeywordClassification(ctx, businessName, description, websiteURL, startTime, requestID)
	}

	// Convert MultiStrategyResult to IndustryDetectionResult
	// Use Confidence field which contains the calibrated value
	result := &IndustryDetectionResult{
		IndustryName:   multiResult.PrimaryIndustry,
		Confidence:     multiResult.Confidence, // This is already calibrated
		Keywords:       multiResult.Keywords,
		ProcessingTime: multiResult.ProcessingTime,
		Method:         "multi_strategy",
		Reasoning:      multiResult.Reasoning,
		CreatedAt:      time.Now(),
	}

	// Record metrics if monitoring is enabled
	if s.monitor != nil {
		// Note: RecordClassificationMetrics signature may need adjustment
		// For now, we'll skip monitoring if method doesn't exist
		// This can be added later when monitoring is fully integrated
	}

	s.logger.Printf("✅ Industry detection completed: %s (confidence: %.2f%%, calibrated: %.2f%%) (request: %s)",
		result.IndustryName, multiResult.Confidence*100, result.Confidence*100, requestID)

	return result, nil
}

// detectIndustryWithML performs industry detection using MultiMethodClassifier with ML support
func (s *IndustryDetectionService) detectIndustryWithML(
	ctx context.Context,
	businessName, description, websiteURL string,
	startTime time.Time,
	requestID string,
) (*IndustryDetectionResult, error) {
	// Use MultiMethodClassifier which combines keyword + ML + description methods
	multiMethodResult, err := s.multiMethodClassifier.ClassifyWithMultipleMethods(
		ctx, businessName, description, websiteURL)
	if err != nil {
		s.logger.Printf("⚠️ Multi-method classification failed, falling back to keyword-based: %v (request: %s)", err, requestID)
		return s.fallbackToKeywordClassification(ctx, businessName, description, websiteURL, startTime, requestID)
	}

	if multiMethodResult == nil || multiMethodResult.PrimaryClassification == nil {
		s.logger.Printf("⚠️ Multi-method returned nil, falling back to keyword-based (request: %s)", requestID)
		return s.fallbackToKeywordClassification(ctx, businessName, description, websiteURL, startTime, requestID)
	}

	// Convert MultiMethodClassificationResult to IndustryDetectionResult
	primary := multiMethodResult.PrimaryClassification
	result := &IndustryDetectionResult{
		IndustryName:   primary.IndustryName,
		Confidence:     primary.ConfidenceScore,
		Keywords:       primary.Keywords,
		ProcessingTime: multiMethodResult.ProcessingTime,
		Method:         primary.ClassificationMethod, // Will be "ml_distilbart" or "ml" or "keyword"
		Reasoning:      multiMethodResult.ClassificationReasoning,
		CreatedAt:      time.Now(),
	}

	// Log which methods were used
	var methodNames []string
	for _, methodResult := range multiMethodResult.MethodResults {
		if methodResult.Success {
			methodNames = append(methodNames, methodResult.MethodName)
		}
	}
	s.logger.Printf("✅ Multi-method classification completed: %s (confidence: %.2f%%) using methods: %v (request: %s)",
		result.IndustryName, result.Confidence*100, methodNames, requestID)

	return result, nil
}

// fallbackToKeywordClassification provides fallback when multi-strategy fails
func (s *IndustryDetectionService) fallbackToKeywordClassification(
	ctx context.Context,
	businessName, description, websiteURL string,
	startTime time.Time,
	requestID string,
) (*IndustryDetectionResult, error) {
	// Extract keywords using database-driven approach (with description fallback)
	keywords, err := s.extractKeywordsFromDatabase(ctx, businessName, description, websiteURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract keywords: %w", err)
	}

	if len(keywords) == 0 {
		// Enhanced fallback: Extract keywords from business name and description
		s.logger.Printf("⚠️ No keywords extracted from website for: %s, trying fallback extraction", businessName)
		keywords = s.extractKeywordsFromNameAndDescription(businessName, description)

	if len(keywords) == 0 {
		s.logger.Printf("⚠️ No keywords extracted for: %s", businessName)
		return &IndustryDetectionResult{
			IndustryName:   "General Business",
			Confidence:     0.30,
			Keywords:       []string{},
			ProcessingTime: time.Since(startTime),
			Method:         "database_driven",
				Reasoning:      "No relevant keywords found in database or fallback extraction",
			CreatedAt:      time.Now(),
		}, nil
		}
		s.logger.Printf("✅ Fallback extraction found %d keywords from name/description", len(keywords))
	}

	// Classify using database-driven keyword matching
	result, err := s.classifyByKeywords(ctx, keywords)
	if err != nil {
		return nil, fmt.Errorf("failed to classify by keywords: %w", err)
	}

	result.ProcessingTime = time.Since(startTime)
	result.Method = "database_driven"
	result.CreatedAt = time.Now()

	s.logger.Printf("✅ Industry detection completed (fallback): %s (confidence: %.2f%%) (request: %s)",
		result.IndustryName, result.Confidence*100, requestID)

	return result, nil
}

// extractKeywordsFromDatabase extracts keywords using database-driven approach
// Now includes description for enhanced fallback keyword extraction
func (s *IndustryDetectionService) extractKeywordsFromDatabase(ctx context.Context, businessName, description, websiteURL string) ([]string, error) {
	// Use the repository's classification method to get keywords
	// Note: ClassifyBusiness may return "General Business" if only URL keywords are available,
	// but it will still return the expanded keywords from the fallback chain.
	// The actual industry classification happens in classifyByKeywords which uses these expanded keywords.
	result, err := s.repo.ClassifyBusiness(ctx, businessName, websiteURL)
	if err != nil {
		return nil, fmt.Errorf("failed to classify business: %w", err)
	}

	if result == nil {
		return []string{}, nil
	}

	keywords := result.Keywords

	// Enhanced: Always supplement with description keywords if available
	// This ensures we use description even when website keywords exist
	if description != "" {
		descKeywords := s.extractKeywordsFromNameAndDescription(businessName, description)
		// Merge keywords, avoiding duplicates
		keywordSet := make(map[string]bool)
		for _, kw := range keywords {
			keywordSet[kw] = true
		}
		for _, kw := range descKeywords {
			if !keywordSet[kw] {
				keywords = append(keywords, kw)
				keywordSet[kw] = true
			}
		}
		if len(descKeywords) > 0 {
			s.logger.Printf("✅ Supplemented with %d keywords from description (total: %d)", len(descKeywords), len(keywords))
		}
	}

	// If still no keywords, try description-only extraction
	if len(keywords) == 0 && description != "" {
		s.logger.Printf("⚠️ No keywords from website, extracting from description for: %s", businessName)
		fallbackKeywords := s.extractKeywordsFromNameAndDescription(businessName, description)
		if len(fallbackKeywords) > 0 {
			keywords = fallbackKeywords
			s.logger.Printf("✅ Extracted %d keywords from description fallback", len(keywords))
		}
	}

	// Return the keywords - these should be the expanded keywords from extractKeywords
	// The industry from ClassifyBusiness may be "General Business" if only 4 URL keywords were found,
	// but classifyByKeywords will use these keywords (which may be expanded by keyword index matching)
	// to correctly identify the industry (e.g., "Wineries")
	return keywords, nil
}

// extractKeywordsFromNameAndDescription extracts keywords from business name and description
// This is a fallback when website scraping fails
func (s *IndustryDetectionService) extractKeywordsFromNameAndDescription(businessName, description string) []string {
	var keywords []string
	seen := make(map[string]bool)

	// Combine business name and description
	text := strings.ToLower(businessName)
	if description != "" {
		text += " " + strings.ToLower(description)
	}

	// Extract meaningful words (3+ characters, not stop words)
	words := strings.Fields(text)
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"is": true, "was": true, "are": true, "were": true, "be": true,
		"been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true,
		"should": true, "could": true, "may": true, "might": true, "must": true,
		"can": true, "this": true, "that": true, "these": true, "those": true,
		"it": true, "its": true, "they": true, "them": true, "their": true,
		"our": true, "your": true, "my": true, "his": true, "her": true,
		"he": true, "she": true, "we": true, "you": true, "i": true, "me": true, "us": true,
		"inc": true, "llc": true, "ltd": true, "corp": true, "corporation": true,
		"company": true, "co": true, "com": true, "www": true, "http": true, "https": true,
		"services": true, "service": true, // Remove generic service words to focus on industry-specific terms
	}

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:()[]{}\"'")
		if len(word) >= 3 && !stopWords[word] && !seen[word] {
			seen[word] = true
			keywords = append(keywords, word)
		}
	}

	// Also extract industry-specific terms from common patterns
	industryPatterns := map[string][]string{
		"technology":     {"software", "tech", "digital", "computer", "internet", "web", "app", "platform", "system", "data", "cloud", "ai", "ml", "api"},
		"healthcare":     {"health", "medical", "hospital", "clinic", "doctor", "patient", "care", "treatment", "pharmacy", "diagnostic", "therapy"},
		"financial":      {"finance", "financial", "bank", "investment", "credit", "loan", "insurance", "trading", "accounting", "tax", "money"},
		"retail":         {"retail", "store", "shop", "merchandise", "product", "sale", "customer", "shopping", "boutique", "outlet"},
		"manufacturing": {"manufacturing", "production", "factory", "industrial", "machinery", "equipment", "assembly", "fabrication"},
		"construction":  {"construction", "contractor", "building", "construction", "architect", "engineering", "renovation", "development"},
		"transportation": {"transport", "transportation", "logistics", "shipping", "delivery", "freight", "trucking", "airline", "railway"},
		"professional":  {"consulting", "professional", "service", "advisory", "legal", "accounting", "management", "consultant"},
	}

	textLower := strings.ToLower(text)
	for _, terms := range industryPatterns {
		for _, term := range terms {
			if strings.Contains(textLower, term) && !seen[term] {
				seen[term] = true
				keywords = append(keywords, term)
			}
		}
	}

	return keywords
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
