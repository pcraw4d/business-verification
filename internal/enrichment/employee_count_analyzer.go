package enrichment

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// EmployeeCountAnalyzer analyzes website content to identify employee count indicators
type EmployeeCountAnalyzer struct {
	config *EmployeeCountConfig
	logger *zap.Logger
	tracer trace.Tracer
}

// EmployeeCountConfig contains configuration for employee count analysis
type EmployeeCountConfig struct {
	// Analysis settings
	EnableEmployeeExtraction bool `json:"enable_employee_extraction"`
	EnableSizeClassification bool `json:"enable_size_classification"`
	EnableConfidenceScoring  bool `json:"enable_confidence_scoring"`
	EnableValidation         bool `json:"enable_validation"`

	// Employee count patterns and keywords
	EmployeeCountPatterns []string `json:"employee_count_patterns"`
	TeamSizeKeywords      []string `json:"team_size_keywords"`
	CompanySizeKeywords   []string `json:"company_size_keywords"`

	// Size classification thresholds
	StartupThreshold    int `json:"startup_threshold"`    // 1-50 employees
	SMEMinThreshold     int `json:"sme_min_threshold"`    // 51 employees
	SMEMaxThreshold     int `json:"sme_max_threshold"`    // 250 employees
	EnterpriseThreshold int `json:"enterprise_threshold"` // 251+ employees

	// Confidence settings
	MinConfidenceThreshold float64 `json:"min_confidence_threshold"`
	MaxExtractionAttempts  int     `json:"max_extraction_attempts"`

	// Quality settings
	EnableDuplicateDetection bool `json:"enable_duplicate_detection"`
	EnableContextValidation  bool `json:"enable_context_validation"`
}

// EmployeeCountResult represents the results of employee count analysis
type EmployeeCountResult struct {
	// Extracted employee count information
	EmployeeCount    int     `json:"employee_count"`
	EmployeeRange    string  `json:"employee_range"`
	ConfidenceScore  float64 `json:"confidence_score"`
	ExtractionMethod string  `json:"extraction_method"`
	IsValidated      bool    `json:"is_validated"`

	// Company size classification
	CompanySize    string  `json:"company_size"`
	SizeConfidence float64 `json:"size_confidence"`
	SizeCategory   string  `json:"size_category"`

	// Evidence and reasoning
	Evidence         []string `json:"evidence"`
	Reasoning        string   `json:"reasoning"`
	ExtractedPhrases []string `json:"extracted_phrases"`

	// Quality metrics
	DataQuality      DataQualityMetrics `json:"data_quality"`
	ValidationStatus ValidationStatus   `json:"validation_status"`

	// Metadata
	ExtractedAt time.Time              `json:"extracted_at"`
	SourceURL   string                 `json:"source_url"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DataQualityMetrics represents data quality assessment
type DataQualityMetrics struct {
	Completeness  float64  `json:"completeness"`
	Accuracy      float64  `json:"accuracy"`
	Consistency   float64  `json:"consistency"`
	Timeliness    float64  `json:"timeliness"`
	OverallScore  float64  `json:"overall_score"`
	MissingFields []string `json:"missing_fields"`
	InvalidFields []string `json:"invalid_fields"`
}

// CompanySizeCategory represents company size categories
type CompanySizeCategory struct {
	Category        string   `json:"category"`
	MinEmployees    int      `json:"min_employees"`
	MaxEmployees    int      `json:"max_employees"`
	ConfidenceScore float64  `json:"confidence_score"`
	Evidence        []string `json:"evidence"`
}

// NewEmployeeCountAnalyzer creates a new employee count analyzer
func NewEmployeeCountAnalyzer(config *EmployeeCountConfig, logger *zap.Logger) *EmployeeCountAnalyzer {
	if config == nil {
		config = getDefaultEmployeeCountConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &EmployeeCountAnalyzer{
		config: config,
		logger: logger,
		tracer: otel.Tracer("employee_count_analyzer"),
	}
}

// AnalyzeEmployeeCount analyzes website content to identify employee count indicators
func (eca *EmployeeCountAnalyzer) AnalyzeEmployeeCount(ctx context.Context, content string, sourceURL string) (*EmployeeCountResult, error) {
	ctx, span := eca.tracer.Start(ctx, "AnalyzeEmployeeCount")
	defer span.End()

	startTime := time.Now()

	span.SetAttributes(
		attribute.String("source_url", sourceURL),
		attribute.Int("content_length", len(content)),
	)

	result := &EmployeeCountResult{
		Evidence:         make([]string, 0),
		ExtractedPhrases: make([]string, 0),
		SourceURL:        sourceURL,
		ExtractedAt:      time.Now(),
		Metadata:         make(map[string]interface{}),
	}

	// Check context timeout
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("analysis cancelled: %w", ctx.Err())
	default:
	}

	// Step 1: Extract employee count from content
	if eca.config.EnableEmployeeExtraction {
		employeeCount, method, evidence, phrases, err := eca.extractEmployeeCount(ctx, content)
		if err != nil {
			eca.logger.Error("employee count extraction failed", zap.Error(err))
		} else {
			result.EmployeeCount = employeeCount
			result.ExtractionMethod = method
			result.Evidence = evidence
			result.ExtractedPhrases = phrases
		}
	}

	// Step 2: Classify company size
	if eca.config.EnableSizeClassification {
		sizeCategory, err := eca.classifyCompanySize(ctx, result.EmployeeCount, content)
		if err != nil {
			eca.logger.Error("company size classification failed", zap.Error(err))
		} else {
			result.CompanySize = sizeCategory.Category
			result.SizeConfidence = sizeCategory.ConfidenceScore
			result.SizeCategory = sizeCategory.Category
			result.Evidence = append(result.Evidence, sizeCategory.Evidence...)
		}
	}

	// Step 3: Calculate confidence score
	if eca.config.EnableConfidenceScoring {
		result.ConfidenceScore = eca.calculateConfidenceScore(result)
	}

	// Step 4: Validate results
	if eca.config.EnableValidation {
		result.ValidationStatus = eca.validateResults(result)
		result.IsValidated = result.ValidationStatus.IsValid
	}

	// Step 5: Calculate data quality metrics
	result.DataQuality = eca.calculateDataQuality(result)

	// Add metadata
	result.Metadata["extraction_duration"] = time.Since(startTime)
	result.Metadata["content_length"] = len(content)
	result.Metadata["extraction_methods"] = eca.getExtractionMethods(result)

	// Generate employee range
	result.EmployeeRange = eca.generateEmployeeRange(result.EmployeeCount)

	// Generate reasoning
	result.Reasoning = eca.generateReasoning(result)

	eca.logger.Info("employee count analysis completed",
		zap.String("source_url", sourceURL),
		zap.Int("employee_count", result.EmployeeCount),
		zap.String("company_size", result.CompanySize),
		zap.Float64("confidence_score", result.ConfidenceScore),
		zap.Duration("duration", time.Since(startTime)))

	return result, nil
}

// extractEmployeeCount extracts employee count from website content
func (eca *EmployeeCountAnalyzer) extractEmployeeCount(ctx context.Context, content string) (int, string, []string, []string, error) {
	ctx, span := eca.tracer.Start(ctx, "extractEmployeeCount")
	defer span.End()

	var employeeCount int
	var extractionMethod string
	var evidence []string
	var extractedPhrases []string

	// Normalize content for analysis
	normalizedContent := strings.ToLower(content)

	// Method 1: Direct employee count patterns
	if count, method, ev, phrases := eca.extractDirectEmployeeCount(normalizedContent); count > 0 {
		employeeCount = count
		extractionMethod = method
		evidence = ev
		extractedPhrases = phrases
	}

	// Method 2: Team size indicators
	if employeeCount == 0 {
		if count, method, ev, phrases := eca.extractTeamSizeIndicators(normalizedContent); count > 0 {
			employeeCount = count
			extractionMethod = method
			evidence = ev
			extractedPhrases = phrases
		}
	}

	// Method 3: Company size keywords
	if employeeCount == 0 {
		if count, method, ev, phrases := eca.extractCompanySizeKeywords(normalizedContent); count > 0 {
			employeeCount = count
			extractionMethod = method
			evidence = ev
			extractedPhrases = phrases
		}
	}

	// Method 4: LinkedIn-style team descriptions
	if employeeCount == 0 {
		if count, method, ev, phrases := eca.extractLinkedInStyleIndicators(normalizedContent); count > 0 {
			employeeCount = count
			extractionMethod = method
			evidence = ev
			extractedPhrases = phrases
		}
	}

	span.SetAttributes(
		attribute.Int("employee_count", employeeCount),
		attribute.String("extraction_method", extractionMethod),
		attribute.Int("evidence_count", len(evidence)),
	)

	return employeeCount, extractionMethod, evidence, extractedPhrases, nil
}

// extractDirectEmployeeCount extracts direct employee count mentions
func (eca *EmployeeCountAnalyzer) extractDirectEmployeeCount(content string) (int, string, []string, []string) {
	var evidence []string
	var phrases []string

	// Direct employee count patterns
	patterns := []string{
		`(?i)(\d+)\s*(?:employees?|staff|team members?)`,
		`(?i)(?:we have|our team consists of|we are)\s*(\d+)\s*(?:employees?|people|team members?)`,
		`(?i)(?:team of|staff of)\s*(\d+)\s*(?:employees?|people|professionals?)`,
		`(?i)(\d+)\s*(?:people|professionals?)\s*(?:team|staff|workforce)`,
		`(?i)(?:over|more than|up to)\s*(\d+)\s*(?:employees?|people|team members?)`,
		`(?i)(?:approximately|about|around)\s*(\d+)\s*(?:employees?|people|team members?)`,
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) > 1 {
				countStr := match[1]
				count, err := strconv.Atoi(countStr)
				if err == nil && count > 0 && count <= 100000 { // Reasonable range
					phrase := strings.TrimSpace(match[0])
					evidence = append(evidence, fmt.Sprintf("Direct mention: %s", phrase))
					phrases = append(phrases, phrase)
					return count, "direct_mention", evidence, phrases
				}
			}
		}
	}

	return 0, "", evidence, phrases
}

// extractTeamSizeIndicators extracts team size indicators
func (eca *EmployeeCountAnalyzer) extractTeamSizeIndicators(content string) (int, string, []string, []string) {
	var evidence []string
	var phrases []string

	// Team size indicators
	teamIndicators := map[string]int{
		"small team":            5,
		"tight-knit team":       8,
		"growing team":          15,
		"dedicated team":        12,
		"core team":             6,
		"startup team":          10,
		"boutique team":         8,
		"lean team":             5,
		"agile team":            7,
		"cross-functional team": 12,
	}

	for indicator, estimatedCount := range teamIndicators {
		if strings.Contains(content, indicator) {
			evidence = append(evidence, fmt.Sprintf("Team indicator: %s", indicator))
			phrases = append(phrases, indicator)
			return estimatedCount, "team_indicator", evidence, phrases
		}
	}

	return 0, "", evidence, phrases
}

// extractCompanySizeKeywords extracts company size from keywords
func (eca *EmployeeCountAnalyzer) extractCompanySizeKeywords(content string) (int, string, []string, []string) {
	var evidence []string
	var phrases []string

	// Company size keywords
	sizeKeywords := map[string]int{
		"startup":            15,
		"small business":     25,
		"medium business":    100,
		"large company":      500,
		"enterprise":         1000,
		"fortune 500":        5000,
		"multinational":      2000,
		"global company":     1500,
		"family business":    20,
		"boutique":           15,
		"agency":             30,
		"consulting firm":    50,
		"technology company": 100,
		"software company":   80,
	}

	for keyword, estimatedCount := range sizeKeywords {
		if strings.Contains(content, keyword) {
			evidence = append(evidence, fmt.Sprintf("Size keyword: %s", keyword))
			phrases = append(phrases, keyword)
			return estimatedCount, "size_keyword", evidence, phrases
		}
	}

	return 0, "", evidence, phrases
}

// extractLinkedInStyleIndicators extracts LinkedIn-style team descriptions
func (eca *EmployeeCountAnalyzer) extractLinkedInStyleIndicators(content string) (int, string, []string, []string) {
	var evidence []string
	var phrases []string

	// LinkedIn-style patterns
	patterns := []string{
		`(?i)(\d+)\+?\s*(?:employees?|people|team members?)\s*(?:worldwide|globally|across|in)`,
		`(?i)(?:join our team of|work with our team of)\s*(\d+)\s*(?:employees?|people|professionals?)`,
		`(?i)(?:we're a team of|we are a team of)\s*(\d+)\s*(?:employees?|people|professionals?)`,
		`(?i)(?:growing from|expanded from)\s*(\d+)\s*(?:to|employees?|people)`,
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) > 1 {
				countStr := match[1]
				count, err := strconv.Atoi(countStr)
				if err == nil && count > 0 && count <= 100000 {
					phrase := strings.TrimSpace(match[0])
					evidence = append(evidence, fmt.Sprintf("LinkedIn-style: %s", phrase))
					phrases = append(phrases, phrase)
					return count, "linkedin_style", evidence, phrases
				}
			}
		}
	}

	return 0, "", evidence, phrases
}

// classifyCompanySize classifies company size based on employee count
func (eca *EmployeeCountAnalyzer) classifyCompanySize(ctx context.Context, employeeCount int, content string) (*CompanySizeCategory, error) {
	ctx, span := eca.tracer.Start(ctx, "classifyCompanySize")
	defer span.End()

	var category string
	var confidence float64
	var evidence []string

	// Determine category based on employee count
	switch {
	case employeeCount == 0:
		category = "unknown"
		confidence = 0.0
		evidence = append(evidence, "No employee count data available")
	case employeeCount <= eca.config.StartupThreshold:
		category = "startup"
		confidence = 0.9
		evidence = append(evidence, fmt.Sprintf("Employee count (%d) within startup range (1-%d)", employeeCount, eca.config.StartupThreshold))
	case employeeCount <= eca.config.SMEMaxThreshold:
		category = "sme"
		confidence = 0.85
		evidence = append(evidence, fmt.Sprintf("Employee count (%d) within SME range (%d-%d)", employeeCount, eca.config.SMEMinThreshold, eca.config.SMEMaxThreshold))
	case employeeCount <= eca.config.EnterpriseThreshold:
		category = "mid_enterprise"
		confidence = 0.8
		evidence = append(evidence, fmt.Sprintf("Employee count (%d) within mid-enterprise range (%d-%d)", employeeCount, eca.config.SMEMaxThreshold+1, eca.config.EnterpriseThreshold))
	default:
		category = "enterprise"
		confidence = 0.9
		evidence = append(evidence, fmt.Sprintf("Employee count (%d) indicates large enterprise", employeeCount))
	}

	// Adjust confidence based on content indicators
	confidence = eca.adjustSizeConfidence(confidence, content, category)

	span.SetAttributes(
		attribute.String("category", category),
		attribute.Float64("confidence", confidence),
		attribute.Int("employee_count", employeeCount),
	)

	return &CompanySizeCategory{
		Category:        category,
		MinEmployees:    eca.getMinEmployees(category),
		MaxEmployees:    eca.getMaxEmployees(category),
		ConfidenceScore: confidence,
		Evidence:        evidence,
	}, nil
}

// adjustSizeConfidence adjusts confidence based on content indicators
func (eca *EmployeeCountAnalyzer) adjustSizeConfidence(baseConfidence float64, content string, category string) float64 {
	confidence := baseConfidence

	// Positive indicators for the category
	positiveIndicators := map[string][]string{
		"startup":    {"startup", "early stage", "seed", "series a", "founded", "new company"},
		"sme":        {"small business", "medium business", "family owned", "local business"},
		"enterprise": {"enterprise", "fortune 500", "multinational", "global", "large company"},
	}

	// Negative indicators for the category
	negativeIndicators := map[string][]string{
		"startup":    {"enterprise", "fortune 500", "large company", "multinational"},
		"sme":        {"startup", "enterprise", "fortune 500"},
		"enterprise": {"startup", "small team", "boutique"},
	}

	// Check positive indicators
	if indicators, exists := positiveIndicators[category]; exists {
		for _, indicator := range indicators {
			if strings.Contains(strings.ToLower(content), indicator) {
				confidence += 0.05
				break
			}
		}
	}

	// Check negative indicators
	if indicators, exists := negativeIndicators[category]; exists {
		for _, indicator := range indicators {
			if strings.Contains(strings.ToLower(content), indicator) {
				confidence -= 0.1
				break
			}
		}
	}

	// Ensure confidence stays within bounds
	if confidence > 1.0 {
		confidence = 1.0
	} else if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// getMinEmployees returns minimum employees for a category
func (eca *EmployeeCountAnalyzer) getMinEmployees(category string) int {
	switch category {
	case "startup":
		return 1
	case "sme":
		return eca.config.SMEMinThreshold
	case "mid_enterprise":
		return eca.config.SMEMaxThreshold + 1
	case "enterprise":
		return eca.config.EnterpriseThreshold + 1
	default:
		return 0
	}
}

// getMaxEmployees returns maximum employees for a category
func (eca *EmployeeCountAnalyzer) getMaxEmployees(category string) int {
	switch category {
	case "startup":
		return eca.config.StartupThreshold
	case "sme":
		return eca.config.SMEMaxThreshold
	case "mid_enterprise":
		return eca.config.EnterpriseThreshold
	case "enterprise":
		return 100000 // Large enterprise
	default:
		return 0
	}
}

// calculateConfidenceScore calculates overall confidence score
func (eca *EmployeeCountAnalyzer) calculateConfidenceScore(result *EmployeeCountResult) float64 {
	confidence := 0.5 // Base confidence

	// Employee count confidence
	if result.EmployeeCount > 0 {
		confidence += 0.2
	}

	// Evidence confidence
	if len(result.Evidence) > 0 {
		confidence += 0.15
	}

	// Size classification confidence
	if result.SizeConfidence > 0 {
		confidence += result.SizeConfidence * 0.15
	}

	// Validation confidence
	if result.IsValidated {
		confidence += 0.1
	}

	// Ensure confidence stays within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// validateResults validates the analysis results
func (eca *EmployeeCountAnalyzer) validateResults(result *EmployeeCountResult) ValidationStatus {
	status := ValidationStatus{
		IsValid:          true,
		ValidationErrors: make([]string, 0),
		LastValidated:    time.Now(),
	}

	// Validate employee count
	if result.EmployeeCount < 0 {
		status.IsValid = false
		status.ValidationErrors = append(status.ValidationErrors, "Employee count cannot be negative")
	}

	if result.EmployeeCount > 1000000 {
		status.IsValid = false
		status.ValidationErrors = append(status.ValidationErrors, "Employee count seems unreasonably high")
	}

	// Validate confidence score
	if result.ConfidenceScore < 0 || result.ConfidenceScore > 1 {
		status.IsValid = false
		status.ValidationErrors = append(status.ValidationErrors, "Confidence score must be between 0 and 1")
	}

	// Validate company size
	if result.CompanySize == "" {
		status.IsValid = false
		status.ValidationErrors = append(status.ValidationErrors, "Company size classification is missing")
	}

	return status
}

// calculateDataQuality calculates data quality metrics
func (eca *EmployeeCountAnalyzer) calculateDataQuality(result *EmployeeCountResult) DataQualityMetrics {
	quality := DataQualityMetrics{
		Completeness:  0.5,
		Accuracy:      0.8,
		Consistency:   0.75,
		Timeliness:    1.0,
		MissingFields: make([]string, 0),
		InvalidFields: make([]string, 0),
	}

	// Completeness
	if result.EmployeeCount > 0 {
		quality.Completeness += 0.3
	} else {
		quality.MissingFields = append(quality.MissingFields, "employee_count")
	}

	if result.CompanySize != "" {
		quality.Completeness += 0.2
	} else {
		quality.MissingFields = append(quality.MissingFields, "company_size")
	}

	// Accuracy based on confidence
	quality.Accuracy = result.ConfidenceScore

	// Consistency
	if len(result.Evidence) > 0 {
		quality.Consistency += 0.15
	}

	if result.IsValidated {
		quality.Consistency += 0.1
	}

	// Calculate overall score
	quality.OverallScore = (quality.Completeness + quality.Accuracy + quality.Consistency + quality.Timeliness) / 4

	return quality
}

// generateEmployeeRange generates a human-readable employee range
func (eca *EmployeeCountAnalyzer) generateEmployeeRange(employeeCount int) string {
	if employeeCount == 0 {
		return "Unknown"
	}

	switch {
	case employeeCount <= 10:
		return "1-10 employees"
	case employeeCount <= 50:
		return "11-50 employees"
	case employeeCount <= 100:
		return "51-100 employees"
	case employeeCount <= 250:
		return "101-250 employees"
	case employeeCount <= 500:
		return "251-500 employees"
	case employeeCount <= 1000:
		return "501-1,000 employees"
	case employeeCount <= 5000:
		return "1,001-5,000 employees"
	case employeeCount <= 10000:
		return "5,001-10,000 employees"
	default:
		return "10,000+ employees"
	}
}

// generateReasoning generates human-readable reasoning for the analysis
func (eca *EmployeeCountAnalyzer) generateReasoning(result *EmployeeCountResult) string {
	if result.EmployeeCount == 0 {
		return "No employee count information could be extracted from the website content."
	}

	reasoning := fmt.Sprintf("Based on analysis of the website content, the company appears to have approximately %d employees. ", result.EmployeeCount)

	if len(result.Evidence) > 0 {
		reasoning += "This conclusion is supported by: "
		for i, evidence := range result.Evidence {
			if i > 0 {
				reasoning += "; "
			}
			reasoning += evidence
		}
		reasoning += ". "
	}

	reasoning += fmt.Sprintf("The company is classified as a %s with %d%% confidence. ",
		strings.Title(result.CompanySize), int(result.ConfidenceScore*100))

	if result.SizeConfidence > 0.8 {
		reasoning += "This classification is highly confident based on multiple indicators. "
	} else if result.SizeConfidence > 0.6 {
		reasoning += "This classification is moderately confident. "
	} else {
		reasoning += "This classification has lower confidence and may require additional verification. "
	}

	return reasoning
}

// getExtractionMethods returns the methods used for extraction
func (eca *EmployeeCountAnalyzer) getExtractionMethods(result *EmployeeCountResult) []string {
	methods := []string{result.ExtractionMethod}

	if result.CompanySize != "" {
		methods = append(methods, "size_classification")
	}

	if result.ConfidenceScore > 0 {
		methods = append(methods, "confidence_scoring")
	}

	if result.IsValidated {
		methods = append(methods, "validation")
	}

	return methods
}

// getDefaultEmployeeCountConfig returns default configuration
func getDefaultEmployeeCountConfig() *EmployeeCountConfig {
	return &EmployeeCountConfig{
		EnableEmployeeExtraction: true,
		EnableSizeClassification: true,
		EnableConfidenceScoring:  true,
		EnableValidation:         true,

		StartupThreshold:    50,
		SMEMinThreshold:     51,
		SMEMaxThreshold:     250,
		EnterpriseThreshold: 251,

		MinConfidenceThreshold: 0.3,
		MaxExtractionAttempts:  5,

		EnableDuplicateDetection: true,
		EnableContextValidation:  true,
	}
}
