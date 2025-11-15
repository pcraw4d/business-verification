package data_discovery

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"

	"kyb-platform/internal/machine_learning"
)

// DataDiscoveryService provides automated data point discovery capabilities
type DataDiscoveryService struct {
	config                *DataDiscoveryConfig
	logger                *zap.Logger
	patternDetector       *PatternDetector
	contentClassifier     *machine_learning.ContentClassifier
	fieldAnalyzer         *FieldAnalyzer
	extractionRulesEngine *ExtractionRulesEngine
	qualityScorer         *QualityScorer
	monitor               *ExtractionMonitor // Added for 3.9.4
}

// DataDiscoveryConfig defines configuration for data discovery
type DataDiscoveryConfig struct {
	MaxDiscoveryDepth      int           `json:"max_discovery_depth"`
	MinConfidenceThreshold float64       `json:"min_confidence_threshold"`
	PatternMatchTimeout    time.Duration `json:"pattern_match_timeout"`
	EnableMLClassification bool          `json:"enable_ml_classification"`
	MaxPatternsPerField    int           `json:"max_patterns_per_field"`
	DiscoveryStrategy      string        `json:"discovery_strategy"` // "aggressive", "conservative", "balanced"
}

// DefaultDataDiscoveryConfig returns default configuration
func DefaultDataDiscoveryConfig() *DataDiscoveryConfig {
	return &DataDiscoveryConfig{
		MaxDiscoveryDepth:      5,
		MinConfidenceThreshold: 0.7,
		PatternMatchTimeout:    10 * time.Second,
		EnableMLClassification: true,
		MaxPatternsPerField:    10,
		DiscoveryStrategy:      "balanced",
	}
}

// DataDiscoveryResult represents the results of automated data point discovery
type DataDiscoveryResult struct {
	DiscoveredFields     []DiscoveredField        `json:"discovered_fields"`
	ConfidenceScore      float64                  `json:"confidence_score"`
	ExtractionRules      []ExtractionRule         `json:"extraction_rules"`
	PatternMatches       []PatternMatch           `json:"pattern_matches"`
	ClassificationResult *ClassificationResult    `json:"classification_result"`
	QualityAssessments   []FieldQualityAssessment `json:"quality_assessments"`
	ProcessingTime       time.Duration            `json:"processing_time"`
	Metadata             map[string]interface{}   `json:"metadata"`
}

// DiscoveredField represents a discovered data field
type DiscoveredField struct {
	FieldName        string                 `json:"field_name"`
	FieldType        string                 `json:"field_type"`
	DataType         string                 `json:"data_type"`
	ConfidenceScore  float64                `json:"confidence_score"`
	ExtractionMethod string                 `json:"extraction_method"`
	SampleValues     []string               `json:"sample_values"`
	ValidationRules  []ValidationRule       `json:"validation_rules"`
	Priority         int                    `json:"priority"`
	BusinessValue    float64                `json:"business_value"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// ExtractionRule defines how to extract a specific type of data
type ExtractionRule struct {
	RuleID          string                 `json:"rule_id"`
	FieldType       string                 `json:"field_type"`
	Pattern         string                 `json:"pattern"`
	XPathSelector   string                 `json:"xpath_selector,omitempty"`
	CSSSelector     string                 `json:"css_selector,omitempty"`
	RegexPattern    string                 `json:"regex_pattern,omitempty"`
	ContextClues    []string               `json:"context_clues"`
	Priority        int                    `json:"priority"`
	ConfidenceScore float64                `json:"confidence_score"`
	ApplicableTypes []string               `json:"applicable_types"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// PatternMatch represents a pattern that was matched in the content
type PatternMatch struct {
	PatternID       string                 `json:"pattern_id"`
	MatchedText     string                 `json:"matched_text"`
	FieldType       string                 `json:"field_type"`
	ConfidenceScore float64                `json:"confidence_score"`
	Context         string                 `json:"context"`
	Position        int                    `json:"position"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ClassificationResult represents the result of content classification
type ClassificationResult struct {
	BusinessType        string                 `json:"business_type"`
	IndustryCategory    string                 `json:"industry_category"`
	ContentCategories   []string               `json:"content_categories"`
	QualityScore        float64                `json:"quality_score"`
	TechnicalIndicators []string               `json:"technical_indicators"`
	ConfidenceScore     float64                `json:"confidence_score"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// ValidationRule defines validation criteria for extracted data
type ValidationRule struct {
	RuleType        string                 `json:"rule_type"`
	Pattern         string                 `json:"pattern,omitempty"`
	MinLength       int                    `json:"min_length,omitempty"`
	MaxLength       int                    `json:"max_length,omitempty"`
	RequiredFormat  string                 `json:"required_format,omitempty"`
	AllowedValues   []string               `json:"allowed_values,omitempty"`
	ConfidenceBoost float64                `json:"confidence_boost"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// NewDataDiscoveryService creates a new data discovery service
func NewDataDiscoveryService(config *DataDiscoveryConfig, logger *zap.Logger) *DataDiscoveryService {
	if config == nil {
		config = DefaultDataDiscoveryConfig()
	}

	// Create ML content classifier with default config
	mlConfig := machine_learning.ContentClassifierConfig{
		ModelType:             "bert",
		MaxSequenceLength:     512,
		BatchSize:             32,
		LearningRate:          0.0001,
		Epochs:                10,
		ValidationSplit:       0.2,
		ConfidenceThreshold:   0.7,
		ExplainabilityEnabled: config.EnableMLClassification,
	}
	mlClassifier := machine_learning.NewContentClassifier(mlConfig)

	return &DataDiscoveryService{
		config:                config,
		logger:                logger,
		patternDetector:       NewPatternDetector(config, logger),
		contentClassifier:     mlClassifier,
		fieldAnalyzer:         NewFieldAnalyzer(config, logger),
		extractionRulesEngine: NewExtractionRulesEngine(config, logger),
		qualityScorer:         NewQualityScorer(config, logger),
		monitor:               NewExtractionMonitor(nil, logger), // Initialize monitor
	}
}

// DiscoverDataPoints performs automated data point discovery on the provided content
func (s *DataDiscoveryService) DiscoverDataPoints(ctx context.Context, content *ContentInput) (*DataDiscoveryResult, error) {
	startTime := time.Now()

	s.logger.Info("Starting automated data point discovery",
		zap.String("content_type", content.ContentType),
		zap.Int("content_length", len(content.RawContent)))

	result := &DataDiscoveryResult{
		DiscoveredFields:   []DiscoveredField{},
		ExtractionRules:    []ExtractionRule{},
		PatternMatches:     []PatternMatch{},
		QualityAssessments: []FieldQualityAssessment{},
		Metadata:           make(map[string]interface{}),
	}

	// Step 1: Classify content to understand context
	// Extract content string and industry from ContentInput
	contentStr := content.RawContent
	if contentStr == "" {
		contentStr = content.HTMLContent
	}
	industry := ""
	if industryVal, ok := content.MetaData["industry"]; ok {
		industry = industryVal
	}

	mlClassification, err := s.contentClassifier.ClassifyContent(ctx, contentStr, industry)
	if err != nil {
		s.logger.Warn("Content classification failed, continuing with discovery",
			zap.Error(err))
	} else {
		// Convert machine_learning.ClassificationResult to local ClassificationResult
		// Extract business type and industry from classifications
		businessType := ""
		industryCategory := ""
		contentCategories := []string{}
		if len(mlClassification.Classifications) > 0 {
			businessType = mlClassification.Classifications[0].Label
			for _, pred := range mlClassification.Classifications {
				contentCategories = append(contentCategories, pred.Label)
			}
		}

		result.ClassificationResult = &ClassificationResult{
			BusinessType:        businessType,
			IndustryCategory:    industryCategory,
			ContentCategories:   contentCategories,
			QualityScore:        mlClassification.QualityScore,
			TechnicalIndicators: []string{}, // Not available in ML result
			ConfidenceScore:     mlClassification.Confidence,
			Metadata:            make(map[string]interface{}),
		}
	}

	// Step 2: Detect patterns in the content
	patterns, err := s.patternDetector.DetectPatterns(ctx, content)
	if err != nil {
		s.logger.Warn("Pattern detection failed, continuing with discovery",
			zap.Error(err))
	} else {
		result.PatternMatches = patterns
	}

	// Step 3: Analyze fields and extract potential data points
	fields, err := s.fieldAnalyzer.AnalyzeFields(ctx, content, patterns, result.ClassificationResult)
	if err != nil {
		s.logger.Warn("Field analysis failed, continuing with discovery",
			zap.Error(err))
	} else {
		result.DiscoveredFields = fields
	}

	// Step 4: Generate extraction rules for discovered fields
	rules, err := s.extractionRulesEngine.GenerateRules(ctx, result.DiscoveredFields, patterns)
	if err != nil {
		s.logger.Warn("Extraction rule generation failed",
			zap.Error(err))
	} else {
		result.ExtractionRules = rules
	}

	// Step 5: Perform quality scoring on discovered fields
	businessContext := s.buildBusinessContext(content, result.ClassificationResult)
	qualityAssessments, err := s.qualityScorer.ScoreDiscoveredFields(ctx, result.DiscoveredFields, patterns, result.ClassificationResult, businessContext)
	if err != nil {
		s.logger.Warn("Quality scoring failed",
			zap.Error(err))
	} else {
		result.QualityAssessments = qualityAssessments
	}

	// Step 6: Calculate overall confidence score
	result.ConfidenceScore = s.calculateOverallConfidence(result)
	result.ProcessingTime = time.Since(startTime)

	// Add metadata
	result.Metadata["discovery_strategy"] = s.config.DiscoveryStrategy
	result.Metadata["fields_discovered"] = len(result.DiscoveredFields)
	result.Metadata["patterns_matched"] = len(result.PatternMatches)
	result.Metadata["rules_generated"] = len(result.ExtractionRules)

	s.logger.Info("Data point discovery completed",
		zap.Float64("confidence_score", result.ConfidenceScore),
		zap.Duration("processing_time", result.ProcessingTime),
		zap.Int("fields_discovered", len(result.DiscoveredFields)))

	// Record metrics with monitor (Added for 3.9.4)
	if s.monitor != nil {
		s.monitor.RecordExtractionResult(ctx, result, result.ProcessingTime, nil)
	}

	return result, nil
}

// GetDiscoveredFieldsByPriority returns discovered fields sorted by priority and business value
func (s *DataDiscoveryService) GetDiscoveredFieldsByPriority(result *DataDiscoveryResult) []DiscoveredField {
	fields := make([]DiscoveredField, len(result.DiscoveredFields))
	copy(fields, result.DiscoveredFields)

	// Sort by priority (lower number = higher priority) and then by business value
	sort.Slice(fields, func(i, j int) bool {
		if fields[i].Priority == fields[j].Priority {
			return fields[i].BusinessValue > fields[j].BusinessValue
		}
		return fields[i].Priority < fields[j].Priority
	})

	return fields
}

// GetHighConfidenceFields returns only fields with confidence above threshold
func (s *DataDiscoveryService) GetHighConfidenceFields(result *DataDiscoveryResult) []DiscoveredField {
	var highConfidenceFields []DiscoveredField

	for _, field := range result.DiscoveredFields {
		if field.ConfidenceScore >= s.config.MinConfidenceThreshold {
			highConfidenceFields = append(highConfidenceFields, field)
		}
	}

	return highConfidenceFields
}

// GenerateExtractionPlan creates an execution plan for extracting discovered data points
func (s *DataDiscoveryService) GenerateExtractionPlan(result *DataDiscoveryResult) *ExtractionPlan {
	plan := &ExtractionPlan{
		PlanID:          generateID(),
		CreatedAt:       time.Now(),
		Fields:          s.GetHighConfidenceFields(result),
		ExtractionRules: result.ExtractionRules,
		Strategy:        s.config.DiscoveryStrategy,
		EstimatedTime:   s.estimateExtractionTime(result),
		Metadata:        make(map[string]interface{}),
	}

	// Group fields by extraction method for optimization
	plan.FieldGroups = s.groupFieldsByExtractionMethod(plan.Fields)

	return plan
}

// ExtractionPlan represents a plan for extracting discovered data points
type ExtractionPlan struct {
	PlanID          string                 `json:"plan_id"`
	CreatedAt       time.Time              `json:"created_at"`
	Fields          []DiscoveredField      `json:"fields"`
	ExtractionRules []ExtractionRule       `json:"extraction_rules"`
	FieldGroups     []FieldGroup           `json:"field_groups"`
	Strategy        string                 `json:"strategy"`
	EstimatedTime   time.Duration          `json:"estimated_time"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// FieldGroup represents a group of fields that can be extracted together
type FieldGroup struct {
	GroupID          string            `json:"group_id"`
	ExtractionMethod string            `json:"extraction_method"`
	Fields           []DiscoveredField `json:"fields"`
	Priority         int               `json:"priority"`
	EstimatedTime    time.Duration     `json:"estimated_time"`
}

// ContentInput represents input content for discovery
type ContentInput struct {
	RawContent     string                 `json:"raw_content"`
	ContentType    string                 `json:"content_type"`
	URL            string                 `json:"url,omitempty"`
	HTMLContent    string                 `json:"html_content,omitempty"`
	StructuredData map[string]interface{} `json:"structured_data,omitempty"`
	MetaData       map[string]string      `json:"metadata,omitempty"`
	Language       string                 `json:"language,omitempty"`
	Encoding       string                 `json:"encoding,omitempty"`
}

// calculateOverallConfidence calculates the overall confidence score for the discovery result
func (s *DataDiscoveryService) calculateOverallConfidence(result *DataDiscoveryResult) float64 {
	if len(result.DiscoveredFields) == 0 {
		return 0.0
	}

	var totalConfidence float64
	var weightSum float64

	// Weight by business value and field priority
	for _, field := range result.DiscoveredFields {
		weight := field.BusinessValue * (1.0 + float64(10-field.Priority)/10.0)
		totalConfidence += field.ConfidenceScore * weight
		weightSum += weight
	}

	if weightSum == 0 {
		return 0.0
	}

	baseConfidence := totalConfidence / weightSum

	// Apply classification confidence if available
	if result.ClassificationResult != nil {
		classificationBonus := result.ClassificationResult.ConfidenceScore * 0.1
		baseConfidence = baseConfidence + classificationBonus
	}

	// Ensure result is within [0, 1] range
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}

	return baseConfidence
}

// estimateExtractionTime estimates the time required to extract all discovered fields
func (s *DataDiscoveryService) estimateExtractionTime(result *DataDiscoveryResult) time.Duration {
	// Base time per field
	baseTimePerField := 100 * time.Millisecond

	// Additional time based on extraction complexity
	var totalTime time.Duration

	for _, field := range result.DiscoveredFields {
		fieldTime := baseTimePerField

		// Adjust based on extraction method
		switch field.ExtractionMethod {
		case "regex":
			fieldTime += 50 * time.Millisecond
		case "xpath":
			fieldTime += 100 * time.Millisecond
		case "ml_classification":
			fieldTime += 500 * time.Millisecond
		case "pattern_matching":
			fieldTime += 200 * time.Millisecond
		}

		// Adjust based on confidence (lower confidence = more time)
		confidenceMultiplier := 2.0 - field.ConfidenceScore
		fieldTime = time.Duration(float64(fieldTime) * confidenceMultiplier)

		totalTime += fieldTime
	}

	return totalTime
}

// groupFieldsByExtractionMethod groups fields by their extraction method for optimization
func (s *DataDiscoveryService) groupFieldsByExtractionMethod(fields []DiscoveredField) []FieldGroup {
	methodGroups := make(map[string][]DiscoveredField)

	for _, field := range fields {
		methodGroups[field.ExtractionMethod] = append(methodGroups[field.ExtractionMethod], field)
	}

	var groups []FieldGroup
	groupIndex := 0

	for method, methodFields := range methodGroups {
		group := FieldGroup{
			GroupID:          fmt.Sprintf("group_%d", groupIndex),
			ExtractionMethod: method,
			Fields:           methodFields,
			Priority:         s.calculateGroupPriority(methodFields),
			EstimatedTime:    s.estimateGroupExtractionTime(methodFields),
		}
		groups = append(groups, group)
		groupIndex++
	}

	// Sort groups by priority
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Priority < groups[j].Priority
	})

	return groups
}

// calculateGroupPriority calculates the priority of a field group
func (s *DataDiscoveryService) calculateGroupPriority(fields []DiscoveredField) int {
	if len(fields) == 0 {
		return 10
	}

	totalPriority := 0
	for _, field := range fields {
		totalPriority += field.Priority
	}

	return totalPriority / len(fields)
}

// estimateGroupExtractionTime estimates extraction time for a group of fields
func (s *DataDiscoveryService) estimateGroupExtractionTime(fields []DiscoveredField) time.Duration {
	if len(fields) == 0 {
		return 0
	}

	// Base time for group setup
	baseTime := 50 * time.Millisecond

	// Time per field in group (optimized for batch processing)
	timePerField := 50 * time.Millisecond

	return baseTime + time.Duration(len(fields))*timePerField
}

// buildBusinessContext creates business context from content and classification
func (s *DataDiscoveryService) buildBusinessContext(content *ContentInput, classification *ClassificationResult) *BusinessContext {
	context := &BusinessContext{
		UseCaseProfile: "verification", // Default use case
		CustomWeights:  make(map[string]float64),
	}

	// Extract context from content metadata
	if content.MetaData != nil {
		if industry, exists := content.MetaData["industry"]; exists {
			context.Industry = industry
		}
		if businessType, exists := content.MetaData["business_type"]; exists {
			context.BusinessType = businessType
		}
		if geography, exists := content.MetaData["geography"]; exists {
			context.Geography = geography
		}
	}

	// Use classification results to enhance context
	if classification != nil {
		if context.Industry == "" {
			context.Industry = classification.IndustryCategory
		}
		if context.BusinessType == "" {
			context.BusinessType = classification.BusinessType
		}
	}

	// Set default priority fields for business verification
	context.PriorityFields = []string{"email", "phone", "address", "tax_id"}

	// Set default weights
	context.CustomWeights["email"] = 1.0
	context.CustomWeights["phone"] = 1.0
	context.CustomWeights["address"] = 0.9
	context.CustomWeights["tax_id"] = 0.8
	context.CustomWeights["url"] = 0.7
	context.CustomWeights["social_media"] = 0.6

	return context
}

// GetQualityAssessmentsByScore returns quality assessments sorted by overall score
func (s *DataDiscoveryService) GetQualityAssessmentsByScore(result *DataDiscoveryResult) []FieldQualityAssessment {
	assessments := make([]FieldQualityAssessment, len(result.QualityAssessments))
	copy(assessments, result.QualityAssessments)

	// Sort by overall quality score (descending)
	sort.Slice(assessments, func(i, j int) bool {
		return assessments[i].QualityScore.OverallScore > assessments[j].QualityScore.OverallScore
	})

	return assessments
}

// GetHighQualityFields returns only fields with quality score above threshold
func (s *DataDiscoveryService) GetHighQualityFields(result *DataDiscoveryResult, threshold float64) []FieldQualityAssessment {
	var highQualityFields []FieldQualityAssessment

	for _, assessment := range result.QualityAssessments {
		if assessment.QualityScore.OverallScore >= threshold {
			highQualityFields = append(highQualityFields, assessment)
		}
	}

	return highQualityFields
}

// GetCriticalBusinessImpactFields returns fields with critical business impact
func (s *DataDiscoveryService) GetCriticalBusinessImpactFields(result *DataDiscoveryResult) []FieldQualityAssessment {
	var criticalFields []FieldQualityAssessment

	for _, assessment := range result.QualityAssessments {
		if assessment.BusinessImpact == "critical" {
			criticalFields = append(criticalFields, assessment)
		}
	}

	return criticalFields
}

// generateID generates a unique identifier
func generateID() string {
	return fmt.Sprintf("discovery_%d", time.Now().UnixNano())
}

// Monitoring and Optimization Methods (Added for 3.9.4)

// GetPerformanceReport returns a comprehensive performance report
func (s *DataDiscoveryService) GetPerformanceReport() *PerformanceReport {
	if s.monitor == nil {
		return nil
	}
	return s.monitor.GetPerformanceReport()
}

// GetMetrics returns current extraction metrics
func (s *DataDiscoveryService) GetMetrics() *ExtractionMetrics {
	if s.monitor == nil {
		return nil
	}
	return s.monitor.GetMetrics()
}

// GetOptimizationStrategies returns current optimization strategies
func (s *DataDiscoveryService) GetOptimizationStrategies() []OptimizationStrategy {
	if s.monitor == nil || s.monitor.optimizer == nil {
		return []OptimizationStrategy{}
	}
	return s.monitor.optimizer.GetOptimizationStrategies()
}

// EnableOptimizationStrategy enables or disables an optimization strategy
func (s *DataDiscoveryService) EnableOptimizationStrategy(strategyName string, enabled bool) error {
	if s.monitor == nil || s.monitor.optimizer == nil {
		return fmt.Errorf("monitor or optimizer not initialized")
	}
	return s.monitor.optimizer.EnableStrategy(strategyName, enabled)
}

// UpdateOptimizationStrategy updates parameters for an optimization strategy
func (s *DataDiscoveryService) UpdateOptimizationStrategy(strategyName string, parameters map[string]interface{}) error {
	if s.monitor == nil || s.monitor.optimizer == nil {
		return fmt.Errorf("monitor or optimizer not initialized")
	}
	return s.monitor.optimizer.UpdateStrategyParameters(strategyName, parameters)
}

// GetActiveAlerts returns all active monitoring alerts
func (s *DataDiscoveryService) GetActiveAlerts() []Alert {
	if s.monitor == nil || s.monitor.alerts == nil {
		return []Alert{}
	}
	return s.monitor.alerts.GetActiveAlerts()
}

// GetAlertSummary returns a summary of alert statistics
func (s *DataDiscoveryService) GetAlertSummary() *AlertSummary {
	if s.monitor == nil || s.monitor.alerts == nil {
		return nil
	}
	return s.monitor.alerts.GetAlertSummary()
}

// AcknowledgeAlert acknowledges an alert by ID
func (s *DataDiscoveryService) AcknowledgeAlert(alertID string) error {
	if s.monitor == nil || s.monitor.alerts == nil {
		return fmt.Errorf("monitor or alerts not initialized")
	}
	return s.monitor.alerts.AcknowledgeAlert(alertID)
}

// ResolveAlert resolves an alert by ID
func (s *DataDiscoveryService) ResolveAlert(alertID string) error {
	if s.monitor == nil || s.monitor.alerts == nil {
		return fmt.Errorf("monitor or alerts not initialized")
	}
	return s.monitor.alerts.ResolveAlert(alertID)
}

// RunOptimization manually triggers optimization
func (s *DataDiscoveryService) RunOptimization() {
	if s.monitor == nil || s.monitor.optimizer == nil {
		s.logger.Warn("Cannot run optimization: monitor or optimizer not initialized")
		return
	}
	s.monitor.optimizer.RunOptimization()
}
