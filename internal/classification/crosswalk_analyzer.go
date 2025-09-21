package classification

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// CrosswalkAnalyzer provides comprehensive analysis and mapping between MCC, NAICS, and SIC codes
type CrosswalkAnalyzer struct {
	db     *sql.DB
	logger *zap.Logger
	config *CrosswalkConfig
}

// CrosswalkConfig configuration for crosswalk analysis
type CrosswalkConfig struct {
	MinConfidenceScore    float64       `json:"min_confidence_score"`
	MaxMappingDistance    int           `json:"max_mapping_distance"`
	EnableValidation      bool          `json:"enable_validation"`
	EnableAutoMapping     bool          `json:"enable_auto_mapping"`
	ValidationTimeout     time.Duration `json:"validation_timeout"`
	BatchSize             int           `json:"batch_size"`
	EnableLogging         bool          `json:"enable_logging"`
	EnablePerformanceMode bool          `json:"enable_performance_mode"`
}

// CrosswalkMapping represents a mapping between different classification systems
type CrosswalkMapping struct {
	ID              int                    `json:"id"`
	IndustryID      int                    `json:"industry_id"`
	SourceCode      string                 `json:"source_code"`
	SourceSystem    string                 `json:"source_system"`
	TargetCode      string                 `json:"target_code"`
	TargetSystem    string                 `json:"target_system"`
	MCCCode         string                 `json:"mcc_code"`
	NAICSCode       string                 `json:"naics_code"`
	SICCode         string                 `json:"sic_code"`
	Description     string                 `json:"description"`
	ConfidenceScore float64                `json:"confidence_score"`
	ValidationRules []ValidationRule       `json:"validation_rules"`
	IsValid         bool                   `json:"is_valid"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ValidationRule represents a validation rule for crosswalk mappings
type ValidationRule struct {
	RuleType    string                 `json:"rule_type"`
	RuleName    string                 `json:"rule_name"`
	RuleValue   interface{}            `json:"rule_value"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// IndustryCodeMapping represents the mapping of codes to industries
type IndustryCodeMapping struct {
	IndustryID      int                    `json:"industry_id"`
	IndustryName    string                 `json:"industry_name"`
	MCCCodes        []string               `json:"mcc_codes"`
	NAICSCodes      []string               `json:"naics_codes"`
	SICCodes        []string               `json:"sic_codes"`
	ConfidenceScore float64                `json:"confidence_score"`
	MappingSource   string                 `json:"mapping_source"`
	LastValidated   time.Time              `json:"last_validated"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// CrosswalkAnalysisResult represents the result of crosswalk analysis
type CrosswalkAnalysisResult struct {
	AnalysisID              string                     `json:"analysis_id"`
	StartTime               time.Time                  `json:"start_time"`
	EndTime                 time.Time                  `json:"end_time"`
	Duration                time.Duration              `json:"duration"`
	TotalMappings           int                        `json:"total_mappings"`
	ValidMappings           int                        `json:"valid_mappings"`
	InvalidMappings         int                        `json:"invalid_mappings"`
	MCCToIndustryMappings   map[string][]string        `json:"mcc_to_industry_mappings"`
	NAICSToIndustryMappings map[string][]string        `json:"naics_to_industry_mappings"`
	SICToIndustryMappings   map[string][]string        `json:"sic_to_industry_mappings"`
	CrosswalkMappings       []CrosswalkMapping         `json:"crosswalk_mappings"`
	ValidationResults       *CrosswalkValidationResult `json:"validation_results"`
	Recommendations         []string                   `json:"recommendations"`
	Issues                  []CrosswalkIssue           `json:"issues"`
}

// CrosswalkValidationResult represents validation results for crosswalk mappings
type CrosswalkValidationResult struct {
	TotalValidations               int              `json:"total_validations"`
	PassedValidations              int              `json:"passed_validations"`
	FailedValidations              int              `json:"failed_validations"`
	ValidationAccuracy             float64          `json:"validation_accuracy"`
	FormatValidationPassed         bool             `json:"format_validation_passed"`
	ConsistencyValidationPassed    bool             `json:"consistency_validation_passed"`
	CrossReferenceValidationPassed bool             `json:"cross_reference_validation_passed"`
	Issues                         []CrosswalkIssue `json:"issues"`
}

// CrosswalkIssue represents an issue found during crosswalk analysis
type CrosswalkIssue struct {
	IssueType      string                 `json:"issue_type"`
	IssueCode      string                 `json:"issue_code"`
	Message        string                 `json:"message"`
	SourceCode     string                 `json:"source_code"`
	SourceSystem   string                 `json:"source_system"`
	TargetCode     string                 `json:"target_code"`
	TargetSystem   string                 `json:"target_system"`
	Severity       string                 `json:"severity"`
	Recommendation string                 `json:"recommendation"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// NewCrosswalkAnalyzer creates a new crosswalk analyzer
func NewCrosswalkAnalyzer(db *sql.DB, logger *zap.Logger, config *CrosswalkConfig) *CrosswalkAnalyzer {
	if config == nil {
		config = &CrosswalkConfig{
			MinConfidenceScore:    0.80,
			MaxMappingDistance:    2,
			EnableValidation:      true,
			EnableAutoMapping:     true,
			ValidationTimeout:     30 * time.Second,
			BatchSize:             100,
			EnableLogging:         true,
			EnablePerformanceMode: false,
		}
	}

	return &CrosswalkAnalyzer{
		db:     db,
		logger: logger,
		config: config,
	}
}

// MapMCCCodesToIndustries maps MCC codes to industries
func (ca *CrosswalkAnalyzer) MapMCCCodesToIndustries(ctx context.Context) (*CrosswalkAnalysisResult, error) {
	startTime := time.Now()
	analysisID := fmt.Sprintf("mcc_industry_mapping_%d", startTime.Unix())

	ca.logger.Info("🔍 Starting MCC to Industry mapping analysis",
		zap.String("analysis_id", analysisID),
		zap.Time("start_time", startTime))

	result := &CrosswalkAnalysisResult{
		AnalysisID:            analysisID,
		StartTime:             startTime,
		MCCToIndustryMappings: make(map[string][]string),
		CrosswalkMappings:     []CrosswalkMapping{},
		Recommendations:       []string{},
		Issues:                []CrosswalkIssue{},
	}

	// Get all industries from the database
	industries, err := ca.getIndustries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get industries: %w", err)
	}

	// Get all MCC codes from classification_codes table
	mccCodes, err := ca.getMCCCodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get MCC codes: %w", err)
	}

	ca.logger.Info("📊 Processing MCC to Industry mappings",
		zap.Int("industries_count", len(industries)),
		zap.Int("mcc_codes_count", len(mccCodes)))

	// Map each MCC code to industries
	for _, mccCode := range mccCodes {
		industryMappings, err := ca.mapMCCToIndustries(ctx, mccCode, industries)
		if err != nil {
			ca.logger.Error("Failed to map MCC code to industries",
				zap.String("mcc_code", mccCode.Code),
				zap.Error(err))
			continue
		}

		// Add mappings to result
		for _, mapping := range industryMappings {
			// Get industry name from metadata
			if industryName, ok := mapping.Metadata["industry_name"].(string); ok {
				result.MCCToIndustryMappings[mccCode.Code] = append(
					result.MCCToIndustryMappings[mccCode.Code],
					industryName,
				)
			}
			result.CrosswalkMappings = append(result.CrosswalkMappings, mapping)
		}

		result.TotalMappings += len(industryMappings)
	}

	// Validate mappings if enabled
	if ca.config.EnableValidation {
		validationResult, err := ca.validateMCCIndustryMappings(ctx, result)
		if err != nil {
			ca.logger.Error("Failed to validate MCC industry mappings", zap.Error(err))
		} else {
			result.ValidationResults = validationResult
		}
	}

	// Generate recommendations
	result.Recommendations = ca.generateMCCMappingRecommendations(result)

	// Set end time and duration
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	ca.logger.Info("✅ MCC to Industry mapping analysis completed",
		zap.String("analysis_id", analysisID),
		zap.Duration("duration", result.Duration),
		zap.Int("total_mappings", result.TotalMappings),
		zap.Int("valid_mappings", result.ValidMappings))

	return result, nil
}

// mapMCCToIndustries maps a single MCC code to relevant industries
func (ca *CrosswalkAnalyzer) mapMCCToIndustries(ctx context.Context, mccCode ClassificationCode, industries []Industry) ([]CrosswalkMapping, error) {
	var mappings []CrosswalkMapping

	for _, industry := range industries {
		// Calculate confidence score based on keyword matching and description similarity
		confidenceScore, err := ca.calculateMCCIndustryConfidence(ctx, mccCode, industry)
		if err != nil {
			ca.logger.Error("Failed to calculate confidence score",
				zap.String("mcc_code", mccCode.Code),
				zap.String("industry", industry.Name),
				zap.Error(err))
			continue
		}

		// Only include mappings above minimum confidence threshold
		if confidenceScore >= ca.config.MinConfidenceScore {
			mapping := CrosswalkMapping{
				IndustryID:      industry.ID,
				SourceCode:      mccCode.Code,
				SourceSystem:    "MCC",
				TargetCode:      fmt.Sprintf("%d", industry.ID),
				TargetSystem:    "INDUSTRY",
				MCCCode:         mccCode.Code,
				Description:     mccCode.Description,
				ConfidenceScore: confidenceScore,
				ValidationRules: ca.generateValidationRules(mccCode, industry),
				IsValid:         true,
				Metadata: map[string]interface{}{
					"mcc_description":    mccCode.Description,
					"industry_name":      industry.Name,
					"industry_category":  industry.Category,
					"mapping_method":     "keyword_similarity",
					"confidence_factors": ca.getConfidenceFactors(mccCode, industry),
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			mappings = append(mappings, mapping)
		}
	}

	return mappings, nil
}

// calculateMCCIndustryConfidence calculates confidence score for MCC to industry mapping
func (ca *CrosswalkAnalyzer) calculateMCCIndustryConfidence(ctx context.Context, mccCode ClassificationCode, industry Industry) (float64, error) {
	// Get industry keywords
	keywords, err := ca.getIndustryKeywords(ctx, industry.ID)
	if err != nil {
		return 0.0, fmt.Errorf("failed to get industry keywords: %w", err)
	}

	// Calculate keyword similarity
	keywordScore := ca.calculateKeywordSimilarity(mccCode.Description, keywords)

	// Calculate description similarity
	descriptionScore := ca.calculateDescriptionSimilarity(mccCode.Description, industry.Description)

	// Calculate category alignment
	categoryScore := ca.calculateCategoryAlignment(mccCode, industry)

	// Weighted combination of scores
	confidenceScore := (keywordScore * 0.4) + (descriptionScore * 0.4) + (categoryScore * 0.2)

	return confidenceScore, nil
}

// calculateKeywordSimilarity calculates similarity between MCC description and industry keywords
func (ca *CrosswalkAnalyzer) calculateKeywordSimilarity(description string, keywords []IndustryKeyword) float64 {
	if len(keywords) == 0 {
		return 0.0
	}

	description = strings.ToLower(description)
	totalScore := 0.0
	totalWeight := 0.0

	for _, keyword := range keywords {
		keywordText := strings.ToLower(keyword.Keyword)
		if strings.Contains(description, keywordText) {
			// Calculate partial match score
			matchScore := float64(len(keywordText)) / float64(len(description))
			weightedScore := matchScore * keyword.Weight
			totalScore += weightedScore
		}
		totalWeight += keyword.Weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalScore / totalWeight
}

// calculateDescriptionSimilarity calculates similarity between MCC and industry descriptions
func (ca *CrosswalkAnalyzer) calculateDescriptionSimilarity(mccDescription, industryDescription string) float64 {
	if industryDescription == "" {
		return 0.0
	}

	// Simple word overlap calculation
	mccWords := strings.Fields(strings.ToLower(mccDescription))
	industryWords := strings.Fields(strings.ToLower(industryDescription))

	if len(mccWords) == 0 || len(industryWords) == 0 {
		return 0.0
	}

	overlap := 0
	for _, mccWord := range mccWords {
		for _, industryWord := range industryWords {
			if mccWord == industryWord {
				overlap++
				break
			}
		}
	}

	// Calculate Jaccard similarity
	union := len(mccWords) + len(industryWords) - overlap
	if union == 0 {
		return 0.0
	}

	return float64(overlap) / float64(union)
}

// calculateCategoryAlignment calculates alignment between MCC and industry categories
func (ca *CrosswalkAnalyzer) calculateCategoryAlignment(mccCode ClassificationCode, industry Industry) float64 {
	// Define category alignment rules
	categoryAlignments := map[string]map[string]float64{
		"traditional": {
			"traditional": 1.0,
			"emerging":    0.3,
			"hybrid":      0.7,
		},
		"emerging": {
			"traditional": 0.3,
			"emerging":    1.0,
			"hybrid":      0.8,
		},
		"hybrid": {
			"traditional": 0.7,
			"emerging":    0.8,
			"hybrid":      1.0,
		},
	}

	// Get MCC category (this would need to be determined based on MCC code)
	mccCategory := ca.determineMCCCategory(mccCode.Code)

	if alignment, exists := categoryAlignments[mccCategory][industry.Category]; exists {
		return alignment
	}

	return 0.5 // Default alignment
}

// determineMCCCategory determines the category of an MCC code
func (ca *CrosswalkAnalyzer) determineMCCCategory(mccCode string) string {
	// This is a simplified categorization - in practice, this would be more sophisticated
	code := mccCode
	if len(code) >= 2 {
		firstTwo := code[:2]
		switch {
		case firstTwo >= "00" && firstTwo <= "19":
			return "traditional"
		case firstTwo >= "20" && firstTwo <= "39":
			return "traditional"
		case firstTwo >= "40" && firstTwo <= "59":
			return "traditional"
		case firstTwo >= "60" && firstTwo <= "79":
			return "emerging"
		case firstTwo >= "80" && firstTwo <= "99":
			return "hybrid"
		default:
			return "traditional"
		}
	}
	return "traditional"
}

// getConfidenceFactors returns factors that contributed to the confidence score
func (ca *CrosswalkAnalyzer) getConfidenceFactors(mccCode ClassificationCode, industry Industry) map[string]float64 {
	return map[string]float64{
		"keyword_similarity":     0.4,
		"description_similarity": 0.4,
		"category_alignment":     0.2,
	}
}

// generateValidationRules generates validation rules for a crosswalk mapping
func (ca *CrosswalkAnalyzer) generateValidationRules(mccCode ClassificationCode, industry Industry) []ValidationRule {
	return []ValidationRule{
		{
			RuleType:    "format",
			RuleName:    "mcc_format_validation",
			RuleValue:   "4-digit numeric",
			Description: "MCC code must be 4-digit numeric format",
			Metadata:    map[string]interface{}{"pattern": "^\\d{4}$"},
		},
		{
			RuleType:    "confidence",
			RuleName:    "minimum_confidence",
			RuleValue:   ca.config.MinConfidenceScore,
			Description: "Minimum confidence score for mapping",
			Metadata:    map[string]interface{}{"threshold": ca.config.MinConfidenceScore},
		},
		{
			RuleType:    "consistency",
			RuleName:    "industry_consistency",
			RuleValue:   industry.Name,
			Description: "Industry must be active and valid",
			Metadata:    map[string]interface{}{"industry_id": industry.ID},
		},
	}
}

// Database helper methods

// getIndustries retrieves all industries from the database
func (ca *CrosswalkAnalyzer) getIndustries(ctx context.Context) ([]Industry, error) {
	query := `
		SELECT id, name, description, category, confidence_threshold, is_active, created_at, updated_at
		FROM industries
		WHERE is_active = true
		ORDER BY name
	`

	rows, err := ca.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query industries: %w", err)
	}
	defer rows.Close()

	var industries []Industry
	for rows.Next() {
		var industry Industry
		err := rows.Scan(
			&industry.ID,
			&industry.Name,
			&industry.Description,
			&industry.Category,
			&industry.ConfidenceThreshold,
			&industry.IsActive,
			&industry.CreatedAt,
			&industry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan industry: %w", err)
		}
		industries = append(industries, industry)
	}

	return industries, nil
}

// getMCCCodes retrieves all MCC codes from the database
func (ca *CrosswalkAnalyzer) getMCCCodes(ctx context.Context) ([]ClassificationCode, error) {
	query := `
		SELECT id, industry_id, code_type, code, description, confidence, is_primary, is_active, created_at, updated_at
		FROM classification_codes
		WHERE code_type = 'MCC' AND is_active = true
		ORDER BY code
	`

	rows, err := ca.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query MCC codes: %w", err)
	}
	defer rows.Close()

	var mccCodes []ClassificationCode
	for rows.Next() {
		var code ClassificationCode
		err := rows.Scan(
			&code.ID,
			&code.IndustryID,
			&code.CodeType,
			&code.Code,
			&code.Description,
			&code.Confidence,
			&code.IsPrimary,
			&code.IsActive,
			&code.CreatedAt,
			&code.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan MCC code: %w", err)
		}
		mccCodes = append(mccCodes, code)
	}

	return mccCodes, nil
}

// getIndustryKeywords retrieves keywords for a specific industry
func (ca *CrosswalkAnalyzer) getIndustryKeywords(ctx context.Context, industryID int) ([]IndustryKeyword, error) {
	query := `
		SELECT id, industry_id, keyword, weight, is_active, created_at, updated_at
		FROM industry_keywords
		WHERE industry_id = $1 AND is_active = true
		ORDER BY weight DESC
	`

	rows, err := ca.db.QueryContext(ctx, query, industryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query industry keywords: %w", err)
	}
	defer rows.Close()

	var keywords []IndustryKeyword
	for rows.Next() {
		var keyword IndustryKeyword
		err := rows.Scan(
			&keyword.ID,
			&keyword.IndustryID,
			&keyword.Keyword,
			&keyword.Weight,
			&keyword.IsActive,
			&keyword.CreatedAt,
			&keyword.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan industry keyword: %w", err)
		}
		keywords = append(keywords, keyword)
	}

	return keywords, nil
}

// validateMCCIndustryMappings validates the MCC to industry mappings
func (ca *CrosswalkAnalyzer) validateMCCIndustryMappings(ctx context.Context, result *CrosswalkAnalysisResult) (*CrosswalkValidationResult, error) {
	validationResult := &CrosswalkValidationResult{
		Issues: []CrosswalkIssue{},
	}

	// Validate format consistency
	formatValid := ca.validateMCCFormatConsistency(result)
	validationResult.FormatValidationPassed = formatValid

	// Validate consistency
	consistencyValid := ca.validateMCCConsistency(result)
	validationResult.ConsistencyValidationPassed = consistencyValid

	// Validate cross-references
	crossRefValid := ca.validateMCCCrossReferences(result)
	validationResult.CrossReferenceValidationPassed = crossRefValid

	// Calculate overall validation accuracy
	validationResult.TotalValidations = len(result.CrosswalkMappings)
	validationResult.PassedValidations = 0
	validationResult.FailedValidations = 0

	for _, mapping := range result.CrosswalkMappings {
		if mapping.IsValid {
			validationResult.PassedValidations++
		} else {
			validationResult.FailedValidations++
		}
	}

	if validationResult.TotalValidations > 0 {
		validationResult.ValidationAccuracy = float64(validationResult.PassedValidations) / float64(validationResult.TotalValidations)
	}

	return validationResult, nil
}

// validateMCCFormatConsistency validates MCC format consistency
func (ca *CrosswalkAnalyzer) validateMCCFormatConsistency(result *CrosswalkAnalysisResult) bool {
	// Check that all MCC codes follow 4-digit format
	for mccCode := range result.MCCToIndustryMappings {
		if len(mccCode) != 4 {
			ca.addCrosswalkIssue(result, "format", "INVALID_MCC_FORMAT",
				fmt.Sprintf("MCC code %s is not 4 digits", mccCode),
				mccCode, "MCC", "", "", "high",
				"Ensure all MCC codes are 4-digit numeric format")
			return false
		}
	}
	return true
}

// validateMCCConsistency validates MCC mapping consistency
func (ca *CrosswalkAnalyzer) validateMCCConsistency(result *CrosswalkAnalysisResult) bool {
	// Check for duplicate mappings
	mappingCounts := make(map[string]int)
	for _, mapping := range result.CrosswalkMappings {
		key := fmt.Sprintf("%s_%s", mapping.SourceCode, mapping.TargetCode)
		mappingCounts[key]++
	}

	hasDuplicates := false
	for key, count := range mappingCounts {
		if count > 1 {
			ca.addCrosswalkIssue(result, "consistency", "DUPLICATE_MAPPING",
				fmt.Sprintf("Duplicate mapping found: %s (count: %d)", key, count),
				"", "", "", "", "medium",
				"Remove duplicate mappings")
			hasDuplicates = true
		}
	}

	return !hasDuplicates
}

// validateMCCCrossReferences validates cross-references for MCC mappings
func (ca *CrosswalkAnalyzer) validateMCCCrossReferences(result *CrosswalkAnalysisResult) bool {
	// This would validate against other classification systems
	// For now, return true as a placeholder
	return true
}

// addCrosswalkIssue adds a crosswalk issue to the result
func (ca *CrosswalkAnalyzer) addCrosswalkIssue(result *CrosswalkAnalysisResult, issueType, issueCode, message, sourceCode, sourceSystem, targetCode, targetSystem, severity, recommendation string) {
	issue := CrosswalkIssue{
		IssueType:      issueType,
		IssueCode:      issueCode,
		Message:        message,
		SourceCode:     sourceCode,
		SourceSystem:   sourceSystem,
		TargetCode:     targetCode,
		TargetSystem:   targetSystem,
		Severity:       severity,
		Recommendation: recommendation,
		Metadata:       make(map[string]interface{}),
	}
	result.Issues = append(result.Issues, issue)
}

// generateMCCMappingRecommendations generates recommendations for MCC mappings
func (ca *CrosswalkAnalyzer) generateMCCMappingRecommendations(result *CrosswalkAnalysisResult) []string {
	var recommendations []string

	// Check for low confidence mappings
	lowConfidenceCount := 0
	for _, mapping := range result.CrosswalkMappings {
		if mapping.ConfidenceScore < 0.9 {
			lowConfidenceCount++
		}
	}

	if lowConfidenceCount > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Review %d mappings with confidence scores below 0.9", lowConfidenceCount))
	}

	// Check for unmapped MCC codes
	if len(result.MCCToIndustryMappings) < result.TotalMappings {
		recommendations = append(recommendations,
			"Some MCC codes may not have industry mappings - review for completeness")
	}

	// Performance recommendations
	if result.Duration > 30*time.Second {
		recommendations = append(recommendations,
			"Consider optimizing mapping performance for large datasets")
	}

	return recommendations
}

// SaveCrosswalkMappings saves crosswalk mappings to the database
func (ca *CrosswalkAnalyzer) SaveCrosswalkMappings(ctx context.Context, mappings []CrosswalkMapping) error {
	if len(mappings) == 0 {
		return nil
	}

	// Prepare batch insert
	query := `
		INSERT INTO crosswalk_mappings (
			id, source_code, source_system, target_code, target_system,
			confidence_score, validation_rules, is_valid, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (source_code, source_system, target_code, target_system)
		DO UPDATE SET
			confidence_score = EXCLUDED.confidence_score,
			validation_rules = EXCLUDED.validation_rules,
			is_valid = EXCLUDED.is_valid,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
	`

	tx, err := ca.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, mapping := range mappings {
		validationRulesJSON, err := json.Marshal(mapping.ValidationRules)
		if err != nil {
			ca.logger.Error("Failed to marshal validation rules", zap.Error(err))
			continue
		}

		metadataJSON, err := json.Marshal(mapping.Metadata)
		if err != nil {
			ca.logger.Error("Failed to marshal metadata", zap.Error(err))
			continue
		}

		_, err = stmt.ExecContext(ctx,
			mapping.ID,
			mapping.SourceCode,
			mapping.SourceSystem,
			mapping.TargetCode,
			mapping.TargetSystem,
			mapping.ConfidenceScore,
			validationRulesJSON,
			mapping.IsValid,
			metadataJSON,
			mapping.CreatedAt,
			mapping.UpdatedAt,
		)
		if err != nil {
			ca.logger.Error("Failed to insert crosswalk mapping",
				zap.Int("mapping_id", mapping.ID),
				zap.Error(err))
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	ca.logger.Info("✅ Crosswalk mappings saved successfully",
		zap.Int("mappings_count", len(mappings)))

	return nil
}

// MapNAICSCodesToIndustries maps NAICS codes to industries
func (ca *CrosswalkAnalyzer) MapNAICSCodesToIndustries(ctx context.Context) (*CrosswalkAnalysisResult, error) {
	startTime := time.Now()
	analysisID := fmt.Sprintf("naics_industry_mapping_%d", startTime.Unix())

	ca.logger.Info("🔍 Starting NAICS to Industry mapping analysis",
		zap.String("analysis_id", analysisID),
		zap.Time("start_time", startTime))

	result := &CrosswalkAnalysisResult{
		AnalysisID:              analysisID,
		StartTime:               startTime,
		NAICSToIndustryMappings: make(map[string][]string),
		CrosswalkMappings:       []CrosswalkMapping{},
		Recommendations:         []string{},
		Issues:                  []CrosswalkIssue{},
	}

	// Get all industries from the database
	industries, err := ca.getIndustries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get industries: %w", err)
	}

	// Get all NAICS codes from classification_codes table
	naicsCodes, err := ca.getNAICSCodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get NAICS codes: %w", err)
	}

	ca.logger.Info("📊 Processing NAICS to Industry mappings",
		zap.Int("industries_count", len(industries)),
		zap.Int("naics_codes_count", len(naicsCodes)))

	// Map each NAICS code to industries
	for _, naicsCode := range naicsCodes {
		industryMappings, err := ca.mapNAICSToIndustries(ctx, naicsCode, industries)
		if err != nil {
			ca.logger.Error("Failed to map NAICS code to industries",
				zap.String("naics_code", naicsCode.Code),
				zap.Error(err))
			continue
		}

		// Add mappings to result
		for _, mapping := range industryMappings {
			// Get industry name from metadata
			if industryName, ok := mapping.Metadata["industry_name"].(string); ok {
				result.NAICSToIndustryMappings[naicsCode.Code] = append(
					result.NAICSToIndustryMappings[naicsCode.Code],
					industryName,
				)
			}
			result.CrosswalkMappings = append(result.CrosswalkMappings, mapping)
		}

		result.TotalMappings += len(industryMappings)
	}

	// Validate mappings if enabled
	if ca.config.EnableValidation {
		validationResult, err := ca.validateNAICSIndustryMappings(ctx, result)
		if err != nil {
			ca.logger.Error("Failed to validate NAICS industry mappings", zap.Error(err))
		} else {
			result.ValidationResults = validationResult
		}
	}

	// Generate recommendations
	result.Recommendations = ca.generateNAICSMappingRecommendations(result)

	// Set end time and duration
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	ca.logger.Info("✅ NAICS to Industry mapping analysis completed",
		zap.String("analysis_id", analysisID),
		zap.Duration("duration", result.Duration),
		zap.Int("total_mappings", result.TotalMappings),
		zap.Int("valid_mappings", result.ValidMappings))

	return result, nil
}

// mapNAICSToIndustries maps a single NAICS code to relevant industries
func (ca *CrosswalkAnalyzer) mapNAICSToIndustries(ctx context.Context, naicsCode ClassificationCode, industries []Industry) ([]CrosswalkMapping, error) {
	var mappings []CrosswalkMapping

	for _, industry := range industries {
		// Calculate confidence score based on keyword matching and description similarity
		confidenceScore, err := ca.calculateNAICSIndustryConfidence(ctx, naicsCode, industry)
		if err != nil {
			ca.logger.Error("Failed to calculate confidence score",
				zap.String("naics_code", naicsCode.Code),
				zap.String("industry", industry.Name),
				zap.Error(err))
			continue
		}

		// Only include mappings above minimum confidence threshold
		if confidenceScore >= ca.config.MinConfidenceScore {
			mapping := CrosswalkMapping{
				IndustryID:      industry.ID,
				SourceCode:      naicsCode.Code,
				SourceSystem:    "NAICS",
				TargetCode:      fmt.Sprintf("%d", industry.ID),
				TargetSystem:    "INDUSTRY",
				NAICSCode:       naicsCode.Code,
				Description:     naicsCode.Description,
				ConfidenceScore: confidenceScore,
				ValidationRules: ca.generateNAICSValidationRules(naicsCode, industry),
				IsValid:         true,
				Metadata: map[string]interface{}{
					"naics_description":  naicsCode.Description,
					"industry_name":      industry.Name,
					"industry_category":  industry.Category,
					"mapping_method":     "keyword_similarity",
					"confidence_factors": ca.getNAICSConfidenceFactors(naicsCode, industry),
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			mappings = append(mappings, mapping)
		}
	}

	return mappings, nil
}

// calculateNAICSIndustryConfidence calculates confidence score for NAICS to industry mapping
func (ca *CrosswalkAnalyzer) calculateNAICSIndustryConfidence(ctx context.Context, naicsCode ClassificationCode, industry Industry) (float64, error) {
	// Get industry keywords
	keywords, err := ca.getIndustryKeywords(ctx, industry.ID)
	if err != nil {
		return 0.0, fmt.Errorf("failed to get industry keywords: %w", err)
	}

	// Calculate keyword similarity
	keywordScore := ca.calculateKeywordSimilarity(naicsCode.Description, keywords)

	// Calculate description similarity
	descriptionScore := ca.calculateDescriptionSimilarity(naicsCode.Description, industry.Description)

	// Calculate NAICS hierarchy alignment
	hierarchyScore := ca.calculateNAICSHierarchyAlignment(naicsCode, industry)

	// Weighted combination of scores
	confidenceScore := (keywordScore * 0.4) + (descriptionScore * 0.3) + (hierarchyScore * 0.3)

	return confidenceScore, nil
}

// calculateNAICSHierarchyAlignment calculates alignment based on NAICS hierarchy
func (ca *CrosswalkAnalyzer) calculateNAICSHierarchyAlignment(naicsCode ClassificationCode, industry Industry) float64 {
	// NAICS hierarchy levels: 2-digit (sector), 3-digit (subsector), 4-digit (industry group), 5-digit (NAICS industry), 6-digit (US industry)
	code := naicsCode.Code
	if len(code) < 2 {
		return 0.0
	}

	// Define sector mappings to industry categories
	sectorMappings := map[string]map[string]float64{
		"11": {"traditional": 1.0, "emerging": 0.2, "hybrid": 0.5}, // Agriculture, Forestry, Fishing and Hunting
		"21": {"traditional": 1.0, "emerging": 0.3, "hybrid": 0.6}, // Mining, Quarrying, and Oil and Gas Extraction
		"22": {"traditional": 1.0, "emerging": 0.4, "hybrid": 0.7}, // Utilities
		"23": {"traditional": 1.0, "emerging": 0.2, "hybrid": 0.5}, // Construction
		"31": {"traditional": 1.0, "emerging": 0.3, "hybrid": 0.6}, // Manufacturing
		"32": {"traditional": 1.0, "emerging": 0.3, "hybrid": 0.6}, // Manufacturing
		"33": {"traditional": 1.0, "emerging": 0.3, "hybrid": 0.6}, // Manufacturing
		"42": {"traditional": 1.0, "emerging": 0.4, "hybrid": 0.7}, // Wholesale Trade
		"44": {"traditional": 1.0, "emerging": 0.5, "hybrid": 0.8}, // Retail Trade
		"45": {"traditional": 1.0, "emerging": 0.5, "hybrid": 0.8}, // Retail Trade
		"48": {"traditional": 1.0, "emerging": 0.4, "hybrid": 0.7}, // Transportation and Warehousing
		"49": {"traditional": 1.0, "emerging": 0.4, "hybrid": 0.7}, // Transportation and Warehousing
		"51": {"traditional": 0.8, "emerging": 0.6, "hybrid": 0.9}, // Information
		"52": {"traditional": 0.7, "emerging": 0.8, "hybrid": 0.9}, // Finance and Insurance
		"53": {"traditional": 0.6, "emerging": 0.7, "hybrid": 0.8}, // Real Estate and Rental and Leasing
		"54": {"traditional": 0.5, "emerging": 0.8, "hybrid": 0.9}, // Professional, Scientific, and Technical Services
		"55": {"traditional": 0.6, "emerging": 0.7, "hybrid": 0.8}, // Management of Companies and Enterprises
		"56": {"traditional": 0.5, "emerging": 0.8, "hybrid": 0.9}, // Administrative and Support and Waste Management and Remediation Services
		"61": {"traditional": 1.0, "emerging": 0.3, "hybrid": 0.6}, // Educational Services
		"62": {"traditional": 1.0, "emerging": 0.4, "hybrid": 0.7}, // Health Care and Social Assistance
		"71": {"traditional": 1.0, "emerging": 0.4, "hybrid": 0.7}, // Arts, Entertainment, and Recreation
		"72": {"traditional": 1.0, "emerging": 0.3, "hybrid": 0.6}, // Accommodation and Food Services
		"81": {"traditional": 0.8, "emerging": 0.6, "hybrid": 0.8}, // Other Services (except Public Administration)
		"92": {"traditional": 1.0, "emerging": 0.2, "hybrid": 0.5}, // Public Administration
	}

	sector := code[:2]
	if alignment, exists := sectorMappings[sector][industry.Category]; exists {
		return alignment
	}

	return 0.5 // Default alignment
}

// getNAICSConfidenceFactors returns factors that contributed to the confidence score
func (ca *CrosswalkAnalyzer) getNAICSConfidenceFactors(naicsCode ClassificationCode, industry Industry) map[string]float64 {
	return map[string]float64{
		"keyword_similarity":     0.4,
		"description_similarity": 0.3,
		"hierarchy_alignment":    0.3,
	}
}

// generateNAICSValidationRules generates validation rules for NAICS crosswalk mapping
func (ca *CrosswalkAnalyzer) generateNAICSValidationRules(naicsCode ClassificationCode, industry Industry) []ValidationRule {
	return []ValidationRule{
		{
			RuleType:    "format",
			RuleName:    "naics_format_validation",
			RuleValue:   "6-digit numeric",
			Description: "NAICS code must be 6-digit numeric format",
			Metadata:    map[string]interface{}{"pattern": "^\\d{6}$"},
		},
		{
			RuleType:    "confidence",
			RuleName:    "minimum_confidence",
			RuleValue:   ca.config.MinConfidenceScore,
			Description: "Minimum confidence score for mapping",
			Metadata:    map[string]interface{}{"threshold": ca.config.MinConfidenceScore},
		},
		{
			RuleType:    "hierarchy",
			RuleName:    "naics_hierarchy_validation",
			RuleValue:   "valid_sector",
			Description: "NAICS code must have valid sector code",
			Metadata:    map[string]interface{}{"sector_codes": []string{"11", "21", "22", "23", "31", "32", "33", "42", "44", "45", "48", "49", "51", "52", "53", "54", "55", "56", "61", "62", "71", "72", "81", "92"}},
		},
		{
			RuleType:    "consistency",
			RuleName:    "industry_consistency",
			RuleValue:   industry.Name,
			Description: "Industry must be active and valid",
			Metadata:    map[string]interface{}{"industry_id": industry.ID},
		},
	}
}

// getNAICSCodes retrieves all NAICS codes from the database
func (ca *CrosswalkAnalyzer) getNAICSCodes(ctx context.Context) ([]ClassificationCode, error) {
	query := `
		SELECT id, industry_id, code_type, code, description, confidence, is_primary, is_active, created_at, updated_at
		FROM classification_codes
		WHERE code_type = 'NAICS' AND is_active = true
		ORDER BY code
	`

	rows, err := ca.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query NAICS codes: %w", err)
	}
	defer rows.Close()

	var naicsCodes []ClassificationCode
	for rows.Next() {
		var code ClassificationCode
		err := rows.Scan(
			&code.ID,
			&code.IndustryID,
			&code.CodeType,
			&code.Code,
			&code.Description,
			&code.Confidence,
			&code.IsPrimary,
			&code.IsActive,
			&code.CreatedAt,
			&code.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan NAICS code: %w", err)
		}
		naicsCodes = append(naicsCodes, code)
	}

	return naicsCodes, nil
}

// validateNAICSIndustryMappings validates the NAICS to industry mappings
func (ca *CrosswalkAnalyzer) validateNAICSIndustryMappings(ctx context.Context, result *CrosswalkAnalysisResult) (*CrosswalkValidationResult, error) {
	validationResult := &CrosswalkValidationResult{
		Issues: []CrosswalkIssue{},
	}

	// Validate format consistency
	formatValid := ca.validateNAICSFormatConsistency(result)
	validationResult.FormatValidationPassed = formatValid

	// Validate consistency
	consistencyValid := ca.validateNAICSConsistency(result)
	validationResult.ConsistencyValidationPassed = consistencyValid

	// Validate cross-references
	crossRefValid := ca.validateNAICSCrossReferences(result)
	validationResult.CrossReferenceValidationPassed = crossRefValid

	// Calculate overall validation accuracy
	validationResult.TotalValidations = len(result.CrosswalkMappings)
	validationResult.PassedValidations = 0
	validationResult.FailedValidations = 0

	for _, mapping := range result.CrosswalkMappings {
		if mapping.IsValid {
			validationResult.PassedValidations++
		} else {
			validationResult.FailedValidations++
		}
	}

	if validationResult.TotalValidations > 0 {
		validationResult.ValidationAccuracy = float64(validationResult.PassedValidations) / float64(validationResult.TotalValidations)
	}

	return validationResult, nil
}

// validateNAICSFormatConsistency validates NAICS format consistency
func (ca *CrosswalkAnalyzer) validateNAICSFormatConsistency(result *CrosswalkAnalysisResult) bool {
	// Check that all NAICS codes follow 6-digit format
	for naicsCode := range result.NAICSToIndustryMappings {
		if len(naicsCode) != 6 {
			ca.addCrosswalkIssue(result, "format", "INVALID_NAICS_FORMAT",
				fmt.Sprintf("NAICS code %s is not 6 digits", naicsCode),
				naicsCode, "NAICS", "", "", "high",
				"Ensure all NAICS codes are 6-digit numeric format")
			return false
		}
	}
	return true
}

// validateNAICSConsistency validates NAICS mapping consistency
func (ca *CrosswalkAnalyzer) validateNAICSConsistency(result *CrosswalkAnalysisResult) bool {
	// Check for duplicate mappings
	mappingCounts := make(map[string]int)
	for _, mapping := range result.CrosswalkMappings {
		key := fmt.Sprintf("%s_%s", mapping.SourceCode, mapping.TargetCode)
		mappingCounts[key]++
	}

	hasDuplicates := false
	for key, count := range mappingCounts {
		if count > 1 {
			ca.addCrosswalkIssue(result, "consistency", "DUPLICATE_MAPPING",
				fmt.Sprintf("Duplicate mapping found: %s (count: %d)", key, count),
				"", "", "", "", "medium",
				"Remove duplicate mappings")
			hasDuplicates = true
		}
	}

	return !hasDuplicates
}

// validateNAICSCrossReferences validates cross-references for NAICS mappings
func (ca *CrosswalkAnalyzer) validateNAICSCrossReferences(result *CrosswalkAnalysisResult) bool {
	// This would validate against other classification systems
	// For now, return true as a placeholder
	return true
}

// generateNAICSMappingRecommendations generates recommendations for NAICS mappings
func (ca *CrosswalkAnalyzer) generateNAICSMappingRecommendations(result *CrosswalkAnalysisResult) []string {
	var recommendations []string

	// Check for low confidence mappings
	lowConfidenceCount := 0
	for _, mapping := range result.CrosswalkMappings {
		if mapping.ConfidenceScore < 0.9 {
			lowConfidenceCount++
		}
	}

	if lowConfidenceCount > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Review %d mappings with confidence scores below 0.9", lowConfidenceCount))
	}

	// Check for unmapped NAICS codes
	if len(result.NAICSToIndustryMappings) < result.TotalMappings {
		recommendations = append(recommendations,
			"Some NAICS codes may not have industry mappings - review for completeness")
	}

	// Performance recommendations
	if result.Duration > 30*time.Second {
		recommendations = append(recommendations,
			"Consider optimizing mapping performance for large datasets")
	}

	// NAICS-specific recommendations
	recommendations = append(recommendations,
		"Consider implementing NAICS hierarchy validation for better accuracy")
	recommendations = append(recommendations,
		"Review sector-level mappings for consistency across subsectors")

	return recommendations
}

// MapSICCodesToIndustries maps SIC codes to industries
func (ca *CrosswalkAnalyzer) MapSICCodesToIndustries(ctx context.Context) (*CrosswalkAnalysisResult, error) {
	startTime := time.Now()
	analysisID := fmt.Sprintf("sic_industry_mapping_%d", startTime.Unix())

	ca.logger.Info("🔍 Starting SIC to Industry mapping analysis",
		zap.String("analysis_id", analysisID),
		zap.Time("start_time", startTime))

	result := &CrosswalkAnalysisResult{
		AnalysisID:            analysisID,
		StartTime:             startTime,
		SICToIndustryMappings: make(map[string][]string),
		CrosswalkMappings:     []CrosswalkMapping{},
		Recommendations:       []string{},
		Issues:                []CrosswalkIssue{},
	}

	// Get all industries from the database
	industries, err := ca.getIndustries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get industries: %w", err)
	}

	// Get all SIC codes from classification_codes table
	sicCodes, err := ca.getSICCodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SIC codes: %w", err)
	}

	ca.logger.Info("📊 Processing SIC to Industry mappings",
		zap.Int("industries_count", len(industries)),
		zap.Int("sic_codes_count", len(sicCodes)))

	// Map each SIC code to industries
	for _, sicCode := range sicCodes {
		industryMappings, err := ca.mapSICToIndustries(ctx, sicCode, industries)
		if err != nil {
			ca.logger.Error("Failed to map SIC code to industries",
				zap.String("sic_code", sicCode.Code),
				zap.Error(err))
			continue
		}

		// Add mappings to result
		for _, mapping := range industryMappings {
			// Get industry name from metadata
			if industryName, ok := mapping.Metadata["industry_name"].(string); ok {
				result.SICToIndustryMappings[sicCode.Code] = append(
					result.SICToIndustryMappings[sicCode.Code],
					industryName,
				)
			}
			result.CrosswalkMappings = append(result.CrosswalkMappings, mapping)
		}

		result.TotalMappings += len(industryMappings)
	}

	// Validate mappings if enabled
	if ca.config.EnableValidation {
		validationResult, err := ca.validateSICIndustryMappings(ctx, result)
		if err != nil {
			ca.logger.Error("Failed to validate SIC industry mappings", zap.Error(err))
		} else {
			result.ValidationResults = validationResult
		}
	}

	// Generate recommendations
	result.Recommendations = ca.generateSICMappingRecommendations(result)

	// Set end time and duration
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	ca.logger.Info("✅ SIC to Industry mapping analysis completed",
		zap.String("analysis_id", analysisID),
		zap.Duration("duration", result.Duration),
		zap.Int("total_mappings", result.TotalMappings),
		zap.Int("valid_mappings", result.ValidMappings))

	return result, nil
}

// mapSICToIndustries maps a single SIC code to relevant industries
func (ca *CrosswalkAnalyzer) mapSICToIndustries(ctx context.Context, sicCode ClassificationCode, industries []Industry) ([]CrosswalkMapping, error) {
	var mappings []CrosswalkMapping

	for _, industry := range industries {
		// Calculate confidence score based on keyword matching and description similarity
		confidenceScore, err := ca.calculateSICIndustryConfidence(ctx, sicCode, industry)
		if err != nil {
			ca.logger.Error("Failed to calculate confidence score",
				zap.String("sic_code", sicCode.Code),
				zap.String("industry", industry.Name),
				zap.Error(err))
			continue
		}

		// Only include mappings above minimum confidence threshold
		if confidenceScore >= ca.config.MinConfidenceScore {
			mapping := CrosswalkMapping{
				IndustryID:      industry.ID,
				SourceCode:      sicCode.Code,
				SourceSystem:    "SIC",
				TargetCode:      fmt.Sprintf("%d", industry.ID),
				TargetSystem:    "INDUSTRY",
				SICCode:         sicCode.Code,
				Description:     sicCode.Description,
				ConfidenceScore: confidenceScore,
				ValidationRules: ca.generateSICValidationRules(sicCode, industry),
				IsValid:         true,
				Metadata: map[string]interface{}{
					"sic_description":    sicCode.Description,
					"industry_name":      industry.Name,
					"industry_category":  industry.Category,
					"mapping_method":     "keyword_similarity",
					"confidence_factors": ca.getSICConfidenceFactors(sicCode, industry),
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			mappings = append(mappings, mapping)
		}
	}

	return mappings, nil
}

// calculateSICIndustryConfidence calculates confidence score for SIC to industry mapping
func (ca *CrosswalkAnalyzer) calculateSICIndustryConfidence(ctx context.Context, sicCode ClassificationCode, industry Industry) (float64, error) {
	// Get industry keywords
	keywords, err := ca.getIndustryKeywords(ctx, industry.ID)
	if err != nil {
		return 0.0, fmt.Errorf("failed to get industry keywords: %w", err)
	}

	// Calculate keyword similarity
	keywordScore := ca.calculateKeywordSimilarity(sicCode.Description, keywords)

	// Calculate description similarity
	descriptionScore := ca.calculateDescriptionSimilarity(sicCode.Description, industry.Description)

	// Calculate SIC division alignment
	divisionScore := ca.calculateSICDivisionAlignment(sicCode, industry)

	// Weighted combination of scores
	confidenceScore := (keywordScore * 0.4) + (descriptionScore * 0.3) + (divisionScore * 0.3)

	return confidenceScore, nil
}

// calculateSICDivisionAlignment calculates alignment based on SIC division
func (ca *CrosswalkAnalyzer) calculateSICDivisionAlignment(sicCode ClassificationCode, industry Industry) float64 {
	// SIC structure: Division (1 digit), Major Group (2 digits), Industry Group (3 digits), Industry (4 digits)
	code := sicCode.Code
	if len(code) < 1 {
		return 0.0
	}

	// Define division mappings to industry categories
	divisionMappings := map[string]map[string]float64{
		"A": {"traditional": 1.0, "emerging": 0.2, "hybrid": 0.5}, // Agriculture, Forestry, and Fishing
		"B": {"traditional": 1.0, "emerging": 0.3, "hybrid": 0.6}, // Mining
		"C": {"traditional": 1.0, "emerging": 0.4, "hybrid": 0.7}, // Construction
		"D": {"traditional": 1.0, "emerging": 0.3, "hybrid": 0.6}, // Manufacturing
		"E": {"traditional": 1.0, "emerging": 0.4, "hybrid": 0.7}, // Transportation, Communications, Electric, Gas, and Sanitary Services
		"F": {"traditional": 1.0, "emerging": 0.4, "hybrid": 0.7}, // Wholesale Trade
		"G": {"traditional": 1.0, "emerging": 0.5, "hybrid": 0.8}, // Retail Trade
		"H": {"traditional": 0.8, "emerging": 0.6, "hybrid": 0.9}, // Finance, Insurance, and Real Estate
		"I": {"traditional": 0.7, "emerging": 0.8, "hybrid": 0.9}, // Services
		"J": {"traditional": 1.0, "emerging": 0.3, "hybrid": 0.6}, // Public Administration
		"K": {"traditional": 0.8, "emerging": 0.6, "hybrid": 0.8}, // Nonclassifiable Establishments
	}

	division := string(code[0])
	if alignment, exists := divisionMappings[division][industry.Category]; exists {
		return alignment
	}

	return 0.5 // Default alignment
}

// getSICConfidenceFactors returns factors that contributed to the confidence score
func (ca *CrosswalkAnalyzer) getSICConfidenceFactors(sicCode ClassificationCode, industry Industry) map[string]float64 {
	return map[string]float64{
		"keyword_similarity":     0.4,
		"description_similarity": 0.3,
		"division_alignment":     0.3,
	}
}

// generateSICValidationRules generates validation rules for SIC crosswalk mapping
func (ca *CrosswalkAnalyzer) generateSICValidationRules(sicCode ClassificationCode, industry Industry) []ValidationRule {
	return []ValidationRule{
		{
			RuleType:    "format",
			RuleName:    "sic_format_validation",
			RuleValue:   "4-digit numeric",
			Description: "SIC code must be 4-digit numeric format",
			Metadata:    map[string]interface{}{"pattern": "^\\d{4}$"},
		},
		{
			RuleType:    "confidence",
			RuleName:    "minimum_confidence",
			RuleValue:   ca.config.MinConfidenceScore,
			Description: "Minimum confidence score for mapping",
			Metadata:    map[string]interface{}{"threshold": ca.config.MinConfidenceScore},
		},
		{
			RuleType:    "division",
			RuleName:    "sic_division_validation",
			RuleValue:   "valid_division",
			Description: "SIC code must have valid division code",
			Metadata:    map[string]interface{}{"division_codes": []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K"}},
		},
		{
			RuleType:    "consistency",
			RuleName:    "industry_consistency",
			RuleValue:   industry.Name,
			Description: "Industry must be active and valid",
			Metadata:    map[string]interface{}{"industry_id": industry.ID},
		},
	}
}

// getSICCodes retrieves all SIC codes from the database
func (ca *CrosswalkAnalyzer) getSICCodes(ctx context.Context) ([]ClassificationCode, error) {
	query := `
		SELECT id, industry_id, code_type, code, description, confidence, is_primary, is_active, created_at, updated_at
		FROM classification_codes
		WHERE code_type = 'SIC' AND is_active = true
		ORDER BY code
	`

	rows, err := ca.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query SIC codes: %w", err)
	}
	defer rows.Close()

	var sicCodes []ClassificationCode
	for rows.Next() {
		var code ClassificationCode
		err := rows.Scan(
			&code.ID,
			&code.IndustryID,
			&code.CodeType,
			&code.Code,
			&code.Description,
			&code.Confidence,
			&code.IsPrimary,
			&code.IsActive,
			&code.CreatedAt,
			&code.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SIC code: %w", err)
		}
		sicCodes = append(sicCodes, code)
	}

	return sicCodes, nil
}

// validateSICIndustryMappings validates the SIC to industry mappings
func (ca *CrosswalkAnalyzer) validateSICIndustryMappings(ctx context.Context, result *CrosswalkAnalysisResult) (*CrosswalkValidationResult, error) {
	validationResult := &CrosswalkValidationResult{
		Issues: []CrosswalkIssue{},
	}

	// Validate format consistency
	formatValid := ca.validateSICFormatConsistency(result)
	validationResult.FormatValidationPassed = formatValid

	// Validate consistency
	consistencyValid := ca.validateSICConsistency(result)
	validationResult.ConsistencyValidationPassed = consistencyValid

	// Validate cross-references
	crossRefValid := ca.validateSICCrossReferences(result)
	validationResult.CrossReferenceValidationPassed = crossRefValid

	// Calculate overall validation accuracy
	validationResult.TotalValidations = len(result.CrosswalkMappings)
	validationResult.PassedValidations = 0
	validationResult.FailedValidations = 0

	for _, mapping := range result.CrosswalkMappings {
		if mapping.IsValid {
			validationResult.PassedValidations++
		} else {
			validationResult.FailedValidations++
		}
	}

	if validationResult.TotalValidations > 0 {
		validationResult.ValidationAccuracy = float64(validationResult.PassedValidations) / float64(validationResult.TotalValidations)
	}

	return validationResult, nil
}

// validateSICFormatConsistency validates SIC format consistency
func (ca *CrosswalkAnalyzer) validateSICFormatConsistency(result *CrosswalkAnalysisResult) bool {
	// Check that all SIC codes follow 4-digit format
	for sicCode := range result.SICToIndustryMappings {
		if len(sicCode) != 4 {
			ca.addCrosswalkIssue(result, "format", "INVALID_SIC_FORMAT",
				fmt.Sprintf("SIC code %s is not 4 digits", sicCode),
				sicCode, "SIC", "", "", "high",
				"Ensure all SIC codes are 4-digit numeric format")
			return false
		}
	}
	return true
}

// validateSICConsistency validates SIC mapping consistency
func (ca *CrosswalkAnalyzer) validateSICConsistency(result *CrosswalkAnalysisResult) bool {
	// Check for duplicate mappings
	mappingCounts := make(map[string]int)
	for _, mapping := range result.CrosswalkMappings {
		key := fmt.Sprintf("%s_%s", mapping.SourceCode, mapping.TargetCode)
		mappingCounts[key]++
	}

	hasDuplicates := false
	for key, count := range mappingCounts {
		if count > 1 {
			ca.addCrosswalkIssue(result, "consistency", "DUPLICATE_MAPPING",
				fmt.Sprintf("Duplicate mapping found: %s (count: %d)", key, count),
				"", "", "", "", "medium",
				"Remove duplicate mappings")
			hasDuplicates = true
		}
	}

	return !hasDuplicates
}

// validateSICCrossReferences validates cross-references for SIC mappings
func (ca *CrosswalkAnalyzer) validateSICCrossReferences(result *CrosswalkAnalysisResult) bool {
	// This would validate against other classification systems
	// For now, return true as a placeholder
	return true
}

// generateSICMappingRecommendations generates recommendations for SIC mappings
func (ca *CrosswalkAnalyzer) generateSICMappingRecommendations(result *CrosswalkAnalysisResult) []string {
	var recommendations []string

	// Check for low confidence mappings
	lowConfidenceCount := 0
	for _, mapping := range result.CrosswalkMappings {
		if mapping.ConfidenceScore < 0.9 {
			lowConfidenceCount++
		}
	}

	if lowConfidenceCount > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Review %d mappings with confidence scores below 0.9", lowConfidenceCount))
	}

	// Check for unmapped SIC codes
	if len(result.SICToIndustryMappings) < result.TotalMappings {
		recommendations = append(recommendations,
			"Some SIC codes may not have industry mappings - review for completeness")
	}

	// Performance recommendations
	if result.Duration > 30*time.Second {
		recommendations = append(recommendations,
			"Consider optimizing mapping performance for large datasets")
	}

	// SIC-specific recommendations
	recommendations = append(recommendations,
		"Consider implementing SIC division validation for better accuracy")
	recommendations = append(recommendations,
		"Review major group mappings for consistency across industry groups")

	return recommendations
}
