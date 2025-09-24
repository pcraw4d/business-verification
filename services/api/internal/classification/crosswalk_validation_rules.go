package classification

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"go.uber.org/zap"
)

// CrosswalkValidationRules defines the validation rules engine for crosswalk mappings
type CrosswalkValidationRules struct {
	db     *sql.DB
	logger *zap.Logger
	config *CrosswalkValidationConfig
}

// CrosswalkValidationConfig defines configuration for validation rules
type CrosswalkValidationConfig struct {
	EnableFormatValidation         bool    `json:"enable_format_validation"`
	EnableConsistencyValidation    bool    `json:"enable_consistency_validation"`
	EnableBusinessLogicValidation  bool    `json:"enable_business_logic_validation"`
	MinConfidenceScore             float64 `json:"min_confidence_score"`
	MaxValidationTime              int     `json:"max_validation_time_seconds"`
	EnableCrossReferenceValidation bool    `json:"enable_cross_reference_validation"`
}

// CrosswalkValidationRule defines a single validation rule for crosswalk mappings
type CrosswalkValidationRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        ValidationRuleType     `json:"type"`
	Severity    ValidationSeverity     `json:"severity"`
	Conditions  map[string]interface{} `json:"conditions"`
	Action      ValidationAction       `json:"action"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ValidationRuleType defines the type of validation rule
type ValidationRuleType string

const (
	ValidationRuleTypeFormat      ValidationRuleType = "format"
	ValidationRuleTypeConsistency ValidationRuleType = "consistency"
	ValidationRuleTypeBusiness    ValidationRuleType = "business_logic"
	ValidationRuleTypeCrossRef    ValidationRuleType = "cross_reference"
)

// ValidationSeverity defines the severity level of a validation rule
type ValidationSeverity string

const (
	ValidationSeverityLow      ValidationSeverity = "low"
	ValidationSeverityMedium   ValidationSeverity = "medium"
	ValidationSeverityHigh     ValidationSeverity = "high"
	ValidationSeverityCritical ValidationSeverity = "critical"
)

// ValidationAction defines the action to take when a rule fails
type ValidationAction string

const (
	ValidationActionWarn  ValidationAction = "warn"
	ValidationActionError ValidationAction = "error"
	ValidationActionBlock ValidationAction = "block"
	ValidationActionLog   ValidationAction = "log"
)

// ValidationRuleResult represents the result of a validation rule execution
type ValidationRuleResult struct {
	RuleID        string                 `json:"rule_id"`
	RuleName      string                 `json:"rule_name"`
	Status        ValidationStatus       `json:"status"`
	Severity      ValidationSeverity     `json:"severity"`
	Message       string                 `json:"message"`
	Details       map[string]interface{} `json:"details"`
	Timestamp     time.Time              `json:"timestamp"`
	ExecutionTime time.Duration          `json:"execution_time"`
}

// ValidationStatus defines the status of a validation result
type ValidationStatus string

const (
	ValidationStatusPassed  ValidationStatus = "passed"
	ValidationStatusFailed  ValidationStatus = "failed"
	ValidationStatusSkipped ValidationStatus = "skipped"
	ValidationStatusError   ValidationStatus = "error"
)

// CrosswalkValidationSummary represents the summary of crosswalk validation
type CrosswalkValidationSummary struct {
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	Duration     time.Duration          `json:"duration"`
	TotalRules   int                    `json:"total_rules"`
	PassedRules  int                    `json:"passed_rules"`
	FailedRules  int                    `json:"failed_rules"`
	SkippedRules int                    `json:"skipped_rules"`
	ErrorRules   int                    `json:"error_rules"`
	Results      []ValidationRuleResult `json:"results"`
	Issues       []ValidationIssue      `json:"issues"`
}

// ValidationIssue represents a validation issue that needs attention
type ValidationIssue struct {
	RuleID    string                 `json:"rule_id"`
	RuleName  string                 `json:"rule_name"`
	Severity  ValidationSeverity     `json:"severity"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewCrosswalkValidationRules creates a new crosswalk validation rules engine
func NewCrosswalkValidationRules(db *sql.DB, logger *zap.Logger, config *CrosswalkValidationConfig) *CrosswalkValidationRules {
	if config == nil {
		config = &CrosswalkValidationConfig{
			EnableFormatValidation:         true,
			EnableConsistencyValidation:    true,
			EnableBusinessLogicValidation:  true,
			MinConfidenceScore:             0.7,
			MaxValidationTime:              30,
			EnableCrossReferenceValidation: true,
		}
	}

	return &CrosswalkValidationRules{
		db:     db,
		logger: logger,
		config: config,
	}
}

// CreateValidationRules creates the standard validation rules for crosswalk mappings
func (cvr *CrosswalkValidationRules) CreateValidationRules(ctx context.Context) error {
	cvr.logger.Info("🔧 Creating crosswalk validation rules")

	rules := []CrosswalkValidationRule{
		// Format validation rules
		{
			ID:          "mcc_format_validation",
			Name:        "MCC Format Validation",
			Description: "Validates that MCC codes are 4-digit numeric codes",
			Type:        ValidationRuleTypeFormat,
			Severity:    ValidationSeverityHigh,
			Conditions: map[string]interface{}{
				"pattern": "^[0-9]{4}$",
				"field":   "mcc_code",
			},
			Action:    ValidationActionError,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "naics_format_validation",
			Name:        "NAICS Format Validation",
			Description: "Validates that NAICS codes are 6-digit numeric codes",
			Type:        ValidationRuleTypeFormat,
			Severity:    ValidationSeverityHigh,
			Conditions: map[string]interface{}{
				"pattern": "^[0-9]{6}$",
				"field":   "naics_code",
			},
			Action:    ValidationActionError,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "sic_format_validation",
			Name:        "SIC Format Validation",
			Description: "Validates that SIC codes are 4-digit numeric codes",
			Type:        ValidationRuleTypeFormat,
			Severity:    ValidationSeverityHigh,
			Conditions: map[string]interface{}{
				"pattern": "^[0-9]{4}$",
				"field":   "sic_code",
			},
			Action:    ValidationActionError,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},

		// Consistency validation rules
		{
			ID:          "confidence_score_validation",
			Name:        "Confidence Score Validation",
			Description: "Validates that confidence scores are within acceptable range",
			Type:        ValidationRuleTypeConsistency,
			Severity:    ValidationSeverityMedium,
			Conditions: map[string]interface{}{
				"min_value": 0.0,
				"max_value": 1.0,
				"field":     "confidence_score",
			},
			Action:    ValidationActionWarn,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "industry_mapping_consistency",
			Name:        "Industry Mapping Consistency",
			Description: "Validates that industry mappings are consistent across classification systems",
			Type:        ValidationRuleTypeConsistency,
			Severity:    ValidationSeverityHigh,
			Conditions: map[string]interface{}{
				"check_consistency": true,
				"tolerance":         0.1,
			},
			Action:    ValidationActionError,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},

		// Business logic validation rules
		{
			ID:          "mcc_industry_alignment",
			Name:        "MCC Industry Alignment",
			Description: "Validates that MCC codes align with appropriate industries",
			Type:        ValidationRuleTypeBusiness,
			Severity:    ValidationSeverityHigh,
			Conditions: map[string]interface{}{
				"check_alignment": true,
				"min_confidence":  0.8,
			},
			Action:    ValidationActionError,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "naics_hierarchy_validation",
			Name:        "NAICS Hierarchy Validation",
			Description: "Validates that NAICS codes follow proper hierarchy",
			Type:        ValidationRuleTypeBusiness,
			Severity:    ValidationSeverityMedium,
			Conditions: map[string]interface{}{
				"check_hierarchy":  true,
				"validate_sectors": true,
			},
			Action:    ValidationActionWarn,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "sic_division_validation",
			Name:        "SIC Division Validation",
			Description: "Validates that SIC codes follow proper division structure",
			Type:        ValidationRuleTypeBusiness,
			Severity:    ValidationSeverityMedium,
			Conditions: map[string]interface{}{
				"check_divisions":    true,
				"validate_structure": true,
			},
			Action:    ValidationActionWarn,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},

		// Cross-reference validation rules
		{
			ID:          "crosswalk_completeness",
			Name:        "Crosswalk Completeness",
			Description: "Validates that crosswalk mappings are complete",
			Type:        ValidationRuleTypeCrossRef,
			Severity:    ValidationSeverityMedium,
			Conditions: map[string]interface{}{
				"check_completeness": true,
				"min_coverage":       0.8,
			},
			Action:    ValidationActionWarn,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "duplicate_mapping_validation",
			Name:        "Duplicate Mapping Validation",
			Description: "Validates that there are no duplicate mappings",
			Type:        ValidationRuleTypeCrossRef,
			Severity:    ValidationSeverityHigh,
			Conditions: map[string]interface{}{
				"check_duplicates": true,
				"allow_duplicates": false,
			},
			Action:    ValidationActionError,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Insert validation rules into database
	for _, rule := range rules {
		if err := cvr.insertValidationRule(ctx, rule); err != nil {
			cvr.logger.Error("Failed to insert validation rule",
				zap.String("rule_id", rule.ID),
				zap.Error(err))
			return fmt.Errorf("failed to insert validation rule %s: %w", rule.ID, err)
		}
	}

	cvr.logger.Info("✅ Successfully created crosswalk validation rules",
		zap.Int("rules_count", len(rules)))

	return nil
}

// ValidateCrosswalkMappings validates all crosswalk mappings against the defined rules
func (cvr *CrosswalkValidationRules) ValidateCrosswalkMappings(ctx context.Context) (*CrosswalkValidationSummary, error) {
	startTime := time.Now()
	cvr.logger.Info("🔍 Starting crosswalk validation")

	summary := &CrosswalkValidationSummary{
		StartTime:    startTime,
		TotalRules:   0,
		PassedRules:  0,
		FailedRules:  0,
		SkippedRules: 0,
		ErrorRules:   0,
		Results:      []ValidationRuleResult{},
		Issues:       []ValidationIssue{},
	}

	// Get all active validation rules
	rules, err := cvr.getActiveValidationRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get validation rules: %w", err)
	}

	summary.TotalRules = len(rules)

	// Execute each validation rule
	for _, rule := range rules {
		result, err := cvr.executeValidationRule(ctx, rule)
		if err != nil {
			cvr.logger.Error("Failed to execute validation rule",
				zap.String("rule_id", rule.ID),
				zap.Error(err))

			result = &ValidationRuleResult{
				RuleID:        rule.ID,
				RuleName:      rule.Name,
				Status:        ValidationStatusError,
				Severity:      rule.Severity,
				Message:       fmt.Sprintf("Validation rule execution failed: %v", err),
				Details:       map[string]interface{}{"error": err.Error()},
				Timestamp:     time.Now(),
				ExecutionTime: time.Since(startTime),
			}
			summary.ErrorRules++
		} else {
			switch result.Status {
			case ValidationStatusPassed:
				summary.PassedRules++
			case ValidationStatusFailed:
				summary.FailedRules++
				// Add to issues if severity is high or critical
				if result.Severity == ValidationSeverityHigh || result.Severity == ValidationSeverityCritical {
					summary.Issues = append(summary.Issues, ValidationIssue{
						RuleID:    result.RuleID,
						RuleName:  result.RuleName,
						Severity:  result.Severity,
						Message:   result.Message,
						Details:   result.Details,
						Timestamp: result.Timestamp,
					})
				}
			case ValidationStatusSkipped:
				summary.SkippedRules++
			}
		}

		summary.Results = append(summary.Results, *result)
	}

	// Set end time and duration
	summary.EndTime = time.Now()
	summary.Duration = summary.EndTime.Sub(summary.StartTime)

	cvr.logger.Info("✅ Crosswalk validation completed",
		zap.Duration("duration", summary.Duration),
		zap.Int("total_rules", summary.TotalRules),
		zap.Int("passed_rules", summary.PassedRules),
		zap.Int("failed_rules", summary.FailedRules),
		zap.Int("issues_count", len(summary.Issues)))

	return summary, nil
}

// executeValidationRule executes a single validation rule
func (cvr *CrosswalkValidationRules) executeValidationRule(ctx context.Context, rule CrosswalkValidationRule) (*ValidationRuleResult, error) {
	startTime := time.Now()
	cvr.logger.Debug("Executing validation rule",
		zap.String("rule_id", rule.ID),
		zap.String("rule_name", rule.Name))

	var result *ValidationRuleResult
	var err error

	switch rule.Type {
	case ValidationRuleTypeFormat:
		result, err = cvr.executeFormatValidation(ctx, rule)
	case ValidationRuleTypeConsistency:
		result, err = cvr.executeConsistencyValidation(ctx, rule)
	case ValidationRuleTypeBusiness:
		result, err = cvr.executeBusinessLogicValidation(ctx, rule)
	case ValidationRuleTypeCrossRef:
		result, err = cvr.executeCrossReferenceValidation(ctx, rule)
	default:
		return nil, fmt.Errorf("unknown validation rule type: %s", rule.Type)
	}

	if err != nil {
		return nil, err
	}

	result.ExecutionTime = time.Since(startTime)
	return result, nil
}

// executeFormatValidation executes format validation rules
func (cvr *CrosswalkValidationRules) executeFormatValidation(ctx context.Context, rule CrosswalkValidationRule) (*ValidationRuleResult, error) {
	pattern, ok := rule.Conditions["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("pattern not found in rule conditions")
	}

	field, ok := rule.Conditions["field"].(string)
	if !ok {
		return nil, fmt.Errorf("field not found in rule conditions")
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Get data to validate based on field
	var invalidRecords []map[string]interface{}

	switch field {
	case "mcc_code":
		invalidRecords, err = cvr.validateMCCFormat(ctx, regex)
	case "naics_code":
		invalidRecords, err = cvr.validateNAICSFormat(ctx, regex)
	case "sic_code":
		invalidRecords, err = cvr.validateSICFormat(ctx, regex)
	default:
		return nil, fmt.Errorf("unknown field for format validation: %s", field)
	}

	if err != nil {
		return nil, err
	}

	if len(invalidRecords) == 0 {
		return &ValidationRuleResult{
			RuleID:    rule.ID,
			RuleName:  rule.Name,
			Status:    ValidationStatusPassed,
			Severity:  rule.Severity,
			Message:   "All records passed format validation",
			Details:   map[string]interface{}{"validated_count": 0},
			Timestamp: time.Now(),
		}, nil
	}

	return &ValidationRuleResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Status:   ValidationStatusFailed,
		Severity: rule.Severity,
		Message:  fmt.Sprintf("Found %d records with invalid format", len(invalidRecords)),
		Details: map[string]interface{}{
			"invalid_count":   len(invalidRecords),
			"invalid_records": invalidRecords,
		},
		Timestamp: time.Now(),
	}, nil
}

// executeConsistencyValidation executes consistency validation rules
func (cvr *CrosswalkValidationRules) executeConsistencyValidation(ctx context.Context, rule CrosswalkValidationRule) (*ValidationRuleResult, error) {
	switch rule.ID {
	case "confidence_score_validation":
		return cvr.validateConfidenceScores(ctx, rule)
	case "industry_mapping_consistency":
		return cvr.validateIndustryMappingConsistency(ctx, rule)
	default:
		return nil, fmt.Errorf("unknown consistency validation rule: %s", rule.ID)
	}
}

// executeBusinessLogicValidation executes business logic validation rules
func (cvr *CrosswalkValidationRules) executeBusinessLogicValidation(ctx context.Context, rule CrosswalkValidationRule) (*ValidationRuleResult, error) {
	switch rule.ID {
	case "mcc_industry_alignment":
		return cvr.validateMCCIndustryAlignment(ctx, rule)
	case "naics_hierarchy_validation":
		return cvr.validateNAICSHierarchy(ctx, rule)
	case "sic_division_validation":
		return cvr.validateSICDivision(ctx, rule)
	default:
		return nil, fmt.Errorf("unknown business logic validation rule: %s", rule.ID)
	}
}

// executeCrossReferenceValidation executes cross-reference validation rules
func (cvr *CrosswalkValidationRules) executeCrossReferenceValidation(ctx context.Context, rule CrosswalkValidationRule) (*ValidationRuleResult, error) {
	switch rule.ID {
	case "crosswalk_completeness":
		return cvr.validateCrosswalkCompleteness(ctx, rule)
	case "duplicate_mapping_validation":
		return cvr.validateDuplicateMappings(ctx, rule)
	default:
		return nil, fmt.Errorf("unknown cross-reference validation rule: %s", rule.ID)
	}
}

// Helper methods for specific validations
func (cvr *CrosswalkValidationRules) validateMCCFormat(ctx context.Context, regex *regexp.Regexp) ([]map[string]interface{}, error) {
	query := `
		SELECT mcc_code, description 
		FROM crosswalk_mappings 
		WHERE mcc_code IS NOT NULL AND mcc_code !~ $1
	`

	rows, err := cvr.db.QueryContext(ctx, query, regex.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invalidRecords []map[string]interface{}
	for rows.Next() {
		var mccCode, description string
		if err := rows.Scan(&mccCode, &description); err != nil {
			return nil, err
		}
		invalidRecords = append(invalidRecords, map[string]interface{}{
			"mcc_code":    mccCode,
			"description": description,
		})
	}

	return invalidRecords, nil
}

func (cvr *CrosswalkValidationRules) validateNAICSFormat(ctx context.Context, regex *regexp.Regexp) ([]map[string]interface{}, error) {
	query := `
		SELECT naics_code, description 
		FROM crosswalk_mappings 
		WHERE naics_code IS NOT NULL AND naics_code !~ $1
	`

	rows, err := cvr.db.QueryContext(ctx, query, regex.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invalidRecords []map[string]interface{}
	for rows.Next() {
		var naicsCode, description string
		if err := rows.Scan(&naicsCode, &description); err != nil {
			return nil, err
		}
		invalidRecords = append(invalidRecords, map[string]interface{}{
			"naics_code":  naicsCode,
			"description": description,
		})
	}

	return invalidRecords, nil
}

func (cvr *CrosswalkValidationRules) validateSICFormat(ctx context.Context, regex *regexp.Regexp) ([]map[string]interface{}, error) {
	query := `
		SELECT sic_code, description 
		FROM crosswalk_mappings 
		WHERE sic_code IS NOT NULL AND sic_code !~ $1
	`

	rows, err := cvr.db.QueryContext(ctx, query, regex.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invalidRecords []map[string]interface{}
	for rows.Next() {
		var sicCode, description string
		if err := rows.Scan(&sicCode, &description); err != nil {
			return nil, err
		}
		invalidRecords = append(invalidRecords, map[string]interface{}{
			"sic_code":    sicCode,
			"description": description,
		})
	}

	return invalidRecords, nil
}

func (cvr *CrosswalkValidationRules) validateConfidenceScores(ctx context.Context, rule CrosswalkValidationRule) (*ValidationRuleResult, error) {
	minValue := rule.Conditions["min_value"].(float64)
	maxValue := rule.Conditions["max_value"].(float64)

	query := `
		SELECT COUNT(*) as invalid_count
		FROM crosswalk_mappings 
		WHERE confidence_score < $1 OR confidence_score > $2
	`

	var invalidCount int
	err := cvr.db.QueryRowContext(ctx, query, minValue, maxValue).Scan(&invalidCount)
	if err != nil {
		return nil, err
	}

	if invalidCount == 0 {
		return &ValidationRuleResult{
			RuleID:    rule.ID,
			RuleName:  rule.Name,
			Status:    ValidationStatusPassed,
			Severity:  rule.Severity,
			Message:   "All confidence scores are within valid range",
			Details:   map[string]interface{}{"validated_count": 0},
			Timestamp: time.Now(),
		}, nil
	}

	return &ValidationRuleResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Status:   ValidationStatusFailed,
		Severity: rule.Severity,
		Message:  fmt.Sprintf("Found %d records with invalid confidence scores", invalidCount),
		Details: map[string]interface{}{
			"invalid_count": invalidCount,
			"min_value":     minValue,
			"max_value":     maxValue,
		},
		Timestamp: time.Now(),
	}, nil
}

func (cvr *CrosswalkValidationRules) validateIndustryMappingConsistency(ctx context.Context, rule CrosswalkValidationRule) (*ValidationRuleResult, error) {
	// Check for inconsistent industry mappings across classification systems
	query := `
		SELECT 
			industry_id,
			COUNT(DISTINCT mcc_code) as mcc_count,
			COUNT(DISTINCT naics_code) as naics_count,
			COUNT(DISTINCT sic_code) as sic_count
		FROM crosswalk_mappings 
		WHERE industry_id IS NOT NULL
		GROUP BY industry_id
		HAVING COUNT(DISTINCT mcc_code) = 0 OR COUNT(DISTINCT naics_code) = 0 OR COUNT(DISTINCT sic_code) = 0
	`

	rows, err := cvr.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inconsistentMappings []map[string]interface{}
	for rows.Next() {
		var industryID int
		var mccCount, naicsCount, sicCount int
		if err := rows.Scan(&industryID, &mccCount, &naicsCount, &sicCount); err != nil {
			return nil, err
		}
		inconsistentMappings = append(inconsistentMappings, map[string]interface{}{
			"industry_id": industryID,
			"mcc_count":   mccCount,
			"naics_count": naicsCount,
			"sic_count":   sicCount,
		})
	}

	if len(inconsistentMappings) == 0 {
		return &ValidationRuleResult{
			RuleID:    rule.ID,
			RuleName:  rule.Name,
			Status:    ValidationStatusPassed,
			Severity:  rule.Severity,
			Message:   "All industry mappings are consistent across classification systems",
			Details:   map[string]interface{}{"validated_count": 0},
			Timestamp: time.Now(),
		}, nil
	}

	return &ValidationRuleResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Status:   ValidationStatusFailed,
		Severity: rule.Severity,
		Message:  fmt.Sprintf("Found %d industries with inconsistent mappings", len(inconsistentMappings)),
		Details: map[string]interface{}{
			"inconsistent_count":    len(inconsistentMappings),
			"inconsistent_mappings": inconsistentMappings,
		},
		Timestamp: time.Now(),
	}, nil
}

func (cvr *CrosswalkValidationRules) validateMCCIndustryAlignment(ctx context.Context, rule CrosswalkValidationRule) (*ValidationRuleResult, error) {
	// Check for MCC codes that don't align with their assigned industries
	query := `
		SELECT 
			cm.mcc_code,
			cm.industry_id,
			cm.confidence_score,
			i.name as industry_name
		FROM crosswalk_mappings cm
		JOIN industries i ON cm.industry_id = i.id
		WHERE cm.mcc_code IS NOT NULL 
		AND cm.confidence_score < $1
	`

	minConfidence := rule.Conditions["min_confidence"].(float64)
	rows, err := cvr.db.QueryContext(ctx, query, minConfidence)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var misalignedMappings []map[string]interface{}
	for rows.Next() {
		var mccCode string
		var industryID int
		var confidenceScore float64
		var industryName string
		if err := rows.Scan(&mccCode, &industryID, &confidenceScore, &industryName); err != nil {
			return nil, err
		}
		misalignedMappings = append(misalignedMappings, map[string]interface{}{
			"mcc_code":         mccCode,
			"industry_id":      industryID,
			"industry_name":    industryName,
			"confidence_score": confidenceScore,
		})
	}

	if len(misalignedMappings) == 0 {
		return &ValidationRuleResult{
			RuleID:    rule.ID,
			RuleName:  rule.Name,
			Status:    ValidationStatusPassed,
			Severity:  rule.Severity,
			Message:   "All MCC codes are properly aligned with industries",
			Details:   map[string]interface{}{"validated_count": 0},
			Timestamp: time.Now(),
		}, nil
	}

	return &ValidationRuleResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Status:   ValidationStatusFailed,
		Severity: rule.Severity,
		Message:  fmt.Sprintf("Found %d MCC codes with low industry alignment", len(misalignedMappings)),
		Details: map[string]interface{}{
			"misaligned_count":    len(misalignedMappings),
			"misaligned_mappings": misalignedMappings,
		},
		Timestamp: time.Now(),
	}, nil
}

func (cvr *CrosswalkValidationRules) validateNAICSHierarchy(ctx context.Context, rule CrosswalkValidationRule) (*ValidationRuleResult, error) {
	// Check for NAICS codes that don't follow proper hierarchy
	query := `
		SELECT 
			naics_code,
			description,
			industry_id
		FROM crosswalk_mappings 
		WHERE naics_code IS NOT NULL 
		AND (
			LENGTH(naics_code) != 6 
			OR naics_code !~ '^[0-9]{6}$'
			OR SUBSTRING(naics_code, 1, 2) NOT IN ('11','21','22','23','31','32','33','42','44','45','48','49','51','52','53','54','55','56','61','62','71','72','81','92')
		)
	`

	rows, err := cvr.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invalidNAICS []map[string]interface{}
	for rows.Next() {
		var naicsCode, description string
		var industryID *int
		if err := rows.Scan(&naicsCode, &description, &industryID); err != nil {
			return nil, err
		}
		invalidNAICS = append(invalidNAICS, map[string]interface{}{
			"naics_code":  naicsCode,
			"description": description,
			"industry_id": industryID,
		})
	}

	if len(invalidNAICS) == 0 {
		return &ValidationRuleResult{
			RuleID:    rule.ID,
			RuleName:  rule.Name,
			Status:    ValidationStatusPassed,
			Severity:  rule.Severity,
			Message:   "All NAICS codes follow proper hierarchy",
			Details:   map[string]interface{}{"validated_count": 0},
			Timestamp: time.Now(),
		}, nil
	}

	return &ValidationRuleResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Status:   ValidationStatusFailed,
		Severity: rule.Severity,
		Message:  fmt.Sprintf("Found %d NAICS codes with invalid hierarchy", len(invalidNAICS)),
		Details: map[string]interface{}{
			"invalid_count": len(invalidNAICS),
			"invalid_naics": invalidNAICS,
		},
		Timestamp: time.Now(),
	}, nil
}

func (cvr *CrosswalkValidationRules) validateSICDivision(ctx context.Context, rule CrosswalkValidationRule) (*ValidationRuleResult, error) {
	// Check for SIC codes that don't follow proper division structure
	query := `
		SELECT 
			sic_code,
			description,
			industry_id
		FROM crosswalk_mappings 
		WHERE sic_code IS NOT NULL 
		AND (
			LENGTH(sic_code) != 4 
			OR sic_code !~ '^[0-9]{4}$'
			OR SUBSTRING(sic_code, 1, 1) NOT IN ('0','1','2','3','4','5','6','7','8','9')
		)
	`

	rows, err := cvr.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invalidSIC []map[string]interface{}
	for rows.Next() {
		var sicCode, description string
		var industryID *int
		if err := rows.Scan(&sicCode, &description, &industryID); err != nil {
			return nil, err
		}
		invalidSIC = append(invalidSIC, map[string]interface{}{
			"sic_code":    sicCode,
			"description": description,
			"industry_id": industryID,
		})
	}

	if len(invalidSIC) == 0 {
		return &ValidationRuleResult{
			RuleID:    rule.ID,
			RuleName:  rule.Name,
			Status:    ValidationStatusPassed,
			Severity:  rule.Severity,
			Message:   "All SIC codes follow proper division structure",
			Details:   map[string]interface{}{"validated_count": 0},
			Timestamp: time.Now(),
		}, nil
	}

	return &ValidationRuleResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Status:   ValidationStatusFailed,
		Severity: rule.Severity,
		Message:  fmt.Sprintf("Found %d SIC codes with invalid division structure", len(invalidSIC)),
		Details: map[string]interface{}{
			"invalid_count": len(invalidSIC),
			"invalid_sic":   invalidSIC,
		},
		Timestamp: time.Now(),
	}, nil
}

func (cvr *CrosswalkValidationRules) validateCrosswalkCompleteness(ctx context.Context, rule CrosswalkValidationRule) (*ValidationRuleResult, error) {
	// Check for missing crosswalk mappings
	query := `
		SELECT 
			COUNT(*) as total_industries,
			COUNT(CASE WHEN mcc_code IS NOT NULL THEN 1 END) as mcc_mappings,
			COUNT(CASE WHEN naics_code IS NOT NULL THEN 1 END) as naics_mappings,
			COUNT(CASE WHEN sic_code IS NOT NULL THEN 1 END) as sic_mappings
		FROM industries i
		LEFT JOIN crosswalk_mappings cm ON i.id = cm.industry_id
	`

	var totalIndustries, mccMappings, naicsMappings, sicMappings int
	err := cvr.db.QueryRowContext(ctx, query).Scan(&totalIndustries, &mccMappings, &naicsMappings, &sicMappings)
	if err != nil {
		return nil, err
	}

	mccCoverage := float64(mccMappings) / float64(totalIndustries)
	naicsCoverage := float64(naicsMappings) / float64(totalIndustries)
	sicCoverage := float64(sicMappings) / float64(totalIndustries)

	minCoverage := rule.Conditions["min_coverage"].(float64)

	if mccCoverage >= minCoverage && naicsCoverage >= minCoverage && sicCoverage >= minCoverage {
		return &ValidationRuleResult{
			RuleID:   rule.ID,
			RuleName: rule.Name,
			Status:   ValidationStatusPassed,
			Severity: rule.Severity,
			Message:  "Crosswalk mappings have sufficient coverage",
			Details: map[string]interface{}{
				"mcc_coverage":   mccCoverage,
				"naics_coverage": naicsCoverage,
				"sic_coverage":   sicCoverage,
			},
			Timestamp: time.Now(),
		}, nil
	}

	return &ValidationRuleResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Status:   ValidationStatusFailed,
		Severity: rule.Severity,
		Message:  "Crosswalk mappings have insufficient coverage",
		Details: map[string]interface{}{
			"mcc_coverage":   mccCoverage,
			"naics_coverage": naicsCoverage,
			"sic_coverage":   sicCoverage,
			"min_coverage":   minCoverage,
		},
		Timestamp: time.Now(),
	}, nil
}

func (cvr *CrosswalkValidationRules) validateDuplicateMappings(ctx context.Context, rule CrosswalkValidationRule) (*ValidationRuleResult, error) {
	// Check for duplicate mappings
	query := `
		SELECT 
			industry_id,
			mcc_code,
			naics_code,
			sic_code,
			COUNT(*) as duplicate_count
		FROM crosswalk_mappings 
		GROUP BY industry_id, mcc_code, naics_code, sic_code
		HAVING COUNT(*) > 1
	`

	rows, err := cvr.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var duplicateMappings []map[string]interface{}
	for rows.Next() {
		var industryID int
		var mccCode, naicsCode, sicCode *string
		var duplicateCount int
		if err := rows.Scan(&industryID, &mccCode, &naicsCode, &sicCode, &duplicateCount); err != nil {
			return nil, err
		}
		duplicateMappings = append(duplicateMappings, map[string]interface{}{
			"industry_id":     industryID,
			"mcc_code":        mccCode,
			"naics_code":      naicsCode,
			"sic_code":        sicCode,
			"duplicate_count": duplicateCount,
		})
	}

	if len(duplicateMappings) == 0 {
		return &ValidationRuleResult{
			RuleID:    rule.ID,
			RuleName:  rule.Name,
			Status:    ValidationStatusPassed,
			Severity:  rule.Severity,
			Message:   "No duplicate mappings found",
			Details:   map[string]interface{}{"validated_count": 0},
			Timestamp: time.Now(),
		}, nil
	}

	return &ValidationRuleResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Status:   ValidationStatusFailed,
		Severity: rule.Severity,
		Message:  fmt.Sprintf("Found %d duplicate mappings", len(duplicateMappings)),
		Details: map[string]interface{}{
			"duplicate_count":    len(duplicateMappings),
			"duplicate_mappings": duplicateMappings,
		},
		Timestamp: time.Now(),
	}, nil
}

// Database helper methods
func (cvr *CrosswalkValidationRules) insertValidationRule(ctx context.Context, rule CrosswalkValidationRule) error {
	query := `
		INSERT INTO validation_rules (id, name, description, type, severity, conditions, action, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			type = EXCLUDED.type,
			severity = EXCLUDED.severity,
			conditions = EXCLUDED.conditions,
			action = EXCLUDED.action,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at
	`

	_, err := cvr.db.ExecContext(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.Type, rule.Severity,
		rule.Conditions, rule.Action, rule.IsActive, rule.CreatedAt, rule.UpdatedAt)

	return err
}

func (cvr *CrosswalkValidationRules) getActiveValidationRules(ctx context.Context) ([]CrosswalkValidationRule, error) {
	query := `
		SELECT id, name, description, type, severity, conditions, action, is_active, created_at, updated_at
		FROM validation_rules 
		WHERE is_active = true
		ORDER BY type, severity DESC
	`

	rows, err := cvr.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []CrosswalkValidationRule
	for rows.Next() {
		var rule CrosswalkValidationRule
		if err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Severity,
			&rule.Conditions, &rule.Action, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, nil
}
