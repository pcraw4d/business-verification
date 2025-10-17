package test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"

	"kyb-platform/internal/classification"
)

// TestCrosswalkValidationRules tests the crosswalk validation rules engine
func TestCrosswalkValidationRules(t *testing.T) {
	// Setup test database
	db, err := setupCrosswalkTestDatabase()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer db.Close()

	// Setup logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Create validation rules engine
	config := &classification.CrosswalkValidationConfig{
		EnableFormatValidation:         true,
		EnableConsistencyValidation:    true,
		EnableBusinessLogicValidation:  true,
		MinConfidenceScore:             0.7,
		MaxValidationTime:              30,
		EnableCrossReferenceValidation: true,
	}

	validator := classification.NewCrosswalkValidationRules(db, logger, config)

	// Test creating validation rules
	t.Run("CreateValidationRules", func(t *testing.T) {
		ctx := context.Background()
		err := validator.CreateValidationRules(ctx)
		if err != nil {
			t.Fatalf("Failed to create validation rules: %v", err)
		}

		logger.Info("✅ Successfully created validation rules")
	})

	// Test validating crosswalk mappings
	t.Run("ValidateCrosswalkMappings", func(t *testing.T) {
		ctx := context.Background()
		summary, err := validator.ValidateCrosswalkMappings(ctx)
		if err != nil {
			t.Fatalf("Failed to validate crosswalk mappings: %v", err)
		}

		// Validate summary
		if summary == nil {
			t.Fatal("Validation summary is nil")
		}

		if summary.TotalRules == 0 {
			t.Error("No validation rules were executed")
		}

		// Log results
		logger.Info("Crosswalk validation completed",
			zap.Duration("duration", summary.Duration),
			zap.Int("total_rules", summary.TotalRules),
			zap.Int("passed_rules", summary.PassedRules),
			zap.Int("failed_rules", summary.FailedRules),
			zap.Int("skipped_rules", summary.SkippedRules),
			zap.Int("error_rules", summary.ErrorRules),
			zap.Int("issues_count", len(summary.Issues)))

		// Save test results
		if err := saveValidationTestResults("crosswalk_validation_test", summary); err != nil {
			t.Errorf("Failed to save test results: %v", err)
		}

		// Validate that we have some results
		if len(summary.Results) == 0 {
			t.Error("No validation results found")
		}

		// Check that we have the expected rule types
		ruleTypes := make(map[string]int)
		for _, result := range summary.Results {
			ruleTypes[result.RuleID]++
		}

		expectedRules := []string{
			"mcc_format_validation",
			"naics_format_validation",
			"sic_format_validation",
			"confidence_score_validation",
			"industry_mapping_consistency",
			"mcc_industry_alignment",
			"naics_hierarchy_validation",
			"sic_division_validation",
			"crosswalk_completeness",
			"duplicate_mapping_validation",
		}

		for _, expectedRule := range expectedRules {
			if ruleTypes[expectedRule] == 0 {
				t.Errorf("Expected rule %s not found in results", expectedRule)
			}
		}
	})

	// Test individual validation rule types
	t.Run("FormatValidation", func(t *testing.T) {
		ctx := context.Background()

		// Test MCC format validation
		rule := classification.CrosswalkValidationRule{
			ID:          "test_mcc_format",
			Name:        "Test MCC Format",
			Description: "Test MCC format validation",
			Type:        classification.ValidationRuleTypeFormat,
			Severity:    classification.ValidationSeverityHigh,
			Conditions: map[string]interface{}{
				"pattern": "^[0-9]{4}$",
				"field":   "mcc_code",
			},
			Action:    classification.ValidationActionError,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Insert test rule
		if err := insertTestValidationRule(ctx, db, rule); err != nil {
			t.Fatalf("Failed to insert test rule: %v", err)
		}

		// Execute validation
		summary, err := validator.ValidateCrosswalkMappings(ctx)
		if err != nil {
			t.Fatalf("Failed to validate crosswalk mappings: %v", err)
		}

		// Check that format validation was executed
		found := false
		for _, result := range summary.Results {
			if result.RuleID == "test_mcc_format" {
				found = true
				if result.Status != classification.ValidationStatusPassed && result.Status != classification.ValidationStatusFailed {
					t.Errorf("Unexpected validation status: %s", result.Status)
				}
				break
			}
		}

		if !found {
			t.Error("Test MCC format validation rule not found in results")
		}
	})

	// Test consistency validation
	t.Run("ConsistencyValidation", func(t *testing.T) {
		ctx := context.Background()

		// Test confidence score validation
		rule := classification.CrosswalkValidationRule{
			ID:          "test_confidence_score",
			Name:        "Test Confidence Score",
			Description: "Test confidence score validation",
			Type:        classification.ValidationRuleTypeConsistency,
			Severity:    classification.ValidationSeverityMedium,
			Conditions: map[string]interface{}{
				"min_value": 0.0,
				"max_value": 1.0,
				"field":     "confidence_score",
			},
			Action:    classification.ValidationActionWarn,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Insert test rule
		if err := insertTestValidationRule(ctx, db, rule); err != nil {
			t.Fatalf("Failed to insert test rule: %v", err)
		}

		// Execute validation
		summary, err := validator.ValidateCrosswalkMappings(ctx)
		if err != nil {
			t.Fatalf("Failed to validate crosswalk mappings: %v", err)
		}

		// Check that consistency validation was executed
		found := false
		for _, result := range summary.Results {
			if result.RuleID == "test_confidence_score" {
				found = true
				if result.Status != classification.ValidationStatusPassed && result.Status != classification.ValidationStatusFailed {
					t.Errorf("Unexpected validation status: %s", result.Status)
				}
				break
			}
		}

		if !found {
			t.Error("Test confidence score validation rule not found in results")
		}
	})

	// Test business logic validation
	t.Run("BusinessLogicValidation", func(t *testing.T) {
		ctx := context.Background()

		// Test MCC industry alignment
		rule := classification.CrosswalkValidationRule{
			ID:          "test_mcc_alignment",
			Name:        "Test MCC Alignment",
			Description: "Test MCC industry alignment",
			Type:        classification.ValidationRuleTypeBusiness,
			Severity:    classification.ValidationSeverityHigh,
			Conditions: map[string]interface{}{
				"check_alignment": true,
				"min_confidence":  0.8,
			},
			Action:    classification.ValidationActionError,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Insert test rule
		if err := insertTestValidationRule(ctx, db, rule); err != nil {
			t.Fatalf("Failed to insert test rule: %v", err)
		}

		// Execute validation
		summary, err := validator.ValidateCrosswalkMappings(ctx)
		if err != nil {
			t.Fatalf("Failed to validate crosswalk mappings: %v", err)
		}

		// Check that business logic validation was executed
		found := false
		for _, result := range summary.Results {
			if result.RuleID == "test_mcc_alignment" {
				found = true
				if result.Status != classification.ValidationStatusPassed && result.Status != classification.ValidationStatusFailed {
					t.Errorf("Unexpected validation status: %s", result.Status)
				}
				break
			}
		}

		if !found {
			t.Error("Test MCC alignment validation rule not found in results")
		}
	})

	// Test cross-reference validation
	t.Run("CrossReferenceValidation", func(t *testing.T) {
		ctx := context.Background()

		// Test crosswalk completeness
		rule := classification.CrosswalkValidationRule{
			ID:          "test_crosswalk_completeness",
			Name:        "Test Crosswalk Completeness",
			Description: "Test crosswalk completeness",
			Type:        classification.ValidationRuleTypeCrossRef,
			Severity:    classification.ValidationSeverityMedium,
			Conditions: map[string]interface{}{
				"check_completeness": true,
				"min_coverage":       0.8,
			},
			Action:    classification.ValidationActionWarn,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Insert test rule
		if err := insertTestValidationRule(ctx, db, rule); err != nil {
			t.Fatalf("Failed to insert test rule: %v", err)
		}

		// Execute validation
		summary, err := validator.ValidateCrosswalkMappings(ctx)
		if err != nil {
			t.Fatalf("Failed to validate crosswalk mappings: %v", err)
		}

		// Check that cross-reference validation was executed
		found := false
		for _, result := range summary.Results {
			if result.RuleID == "test_crosswalk_completeness" {
				found = true
				if result.Status != classification.ValidationStatusPassed && result.Status != classification.ValidationStatusFailed {
					t.Errorf("Unexpected validation status: %s", result.Status)
				}
				break
			}
		}

		if !found {
			t.Error("Test crosswalk completeness validation rule not found in results")
		}
	})

	// Test validation rule management
	t.Run("ValidationRuleManagement", func(t *testing.T) {
		ctx := context.Background()

		// Test getting active validation rules
		rules, err := getActiveValidationRules(ctx, db)
		if err != nil {
			t.Fatalf("Failed to get active validation rules: %v", err)
		}

		if len(rules) == 0 {
			t.Error("No active validation rules found")
		}

		// Check that we have rules of different types
		ruleTypes := make(map[string]int)
		for _, rule := range rules {
			ruleTypes[string(rule.Type)]++
		}

		expectedTypes := []string{
			string(classification.ValidationRuleTypeFormat),
			string(classification.ValidationRuleTypeConsistency),
			string(classification.ValidationRuleTypeBusiness),
			string(classification.ValidationRuleTypeCrossRef),
		}

		for _, expectedType := range expectedTypes {
			if ruleTypes[expectedType] == 0 {
				t.Errorf("No rules found for type: %s", expectedType)
			}
		}

		logger.Info("Validation rule management test completed",
			zap.Int("total_rules", len(rules)),
			zap.Any("rule_types", ruleTypes))
	})

	// Test validation performance
	t.Run("ValidationPerformance", func(t *testing.T) {
		ctx := context.Background()

		startTime := time.Now()
		summary, err := validator.ValidateCrosswalkMappings(ctx)
		duration := time.Since(startTime)

		if err != nil {
			t.Fatalf("Failed to validate crosswalk mappings: %v", err)
		}

		// Check that validation completed within reasonable time
		if duration > 30*time.Second {
			t.Errorf("Validation took too long: %v", duration)
		}

		// Check that we have reasonable execution times for individual rules
		for _, result := range summary.Results {
			if result.ExecutionTime > 5*time.Second {
				t.Errorf("Rule %s took too long: %v", result.RuleID, result.ExecutionTime)
			}
		}

		logger.Info("Validation performance test completed",
			zap.Duration("total_duration", duration),
			zap.Int("total_rules", summary.TotalRules),
			zap.Duration("avg_rule_time", duration/time.Duration(summary.TotalRules)))
	})
}

// Helper functions for testing
func insertTestValidationRule(ctx context.Context, db *sql.DB, rule classification.CrosswalkValidationRule) error {
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

	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.Type, rule.Severity,
		conditionsJSON, rule.Action, rule.IsActive, rule.CreatedAt, rule.UpdatedAt)

	return err
}

func getActiveValidationRules(ctx context.Context, db *sql.DB) ([]classification.CrosswalkValidationRule, error) {
	query := `
		SELECT id, name, description, type, severity, conditions, action, is_active, created_at, updated_at
		FROM validation_rules 
		WHERE is_active = true
		ORDER BY type, severity DESC
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []classification.CrosswalkValidationRule
	for rows.Next() {
		var rule classification.CrosswalkValidationRule
		var conditionsJSON []byte

		if err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Severity,
			&conditionsJSON, &rule.Action, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
			return nil, err
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func saveValidationTestResults(testName string, summary *classification.CrosswalkValidationSummary) error {
	// Create test results directory if it doesn't exist
	if err := os.MkdirAll("test_results", 0755); err != nil {
		return err
	}

	// Create filename with timestamp
	filename := fmt.Sprintf("test_results/%s_%d.json", testName, time.Now().Unix())

	// Marshal summary to JSON
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(filename, data, 0644)
}

// Benchmark tests
func BenchmarkCrosswalkValidation(b *testing.B) {
	// Setup test database
	db, err := setupCrosswalkTestDatabase()
	if err != nil {
		b.Fatalf("Failed to setup test database: %v", err)
	}
	defer db.Close()

	// Setup logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Create validation rules engine
	config := &classification.CrosswalkValidationConfig{
		EnableFormatValidation:         true,
		EnableConsistencyValidation:    true,
		EnableBusinessLogicValidation:  true,
		MinConfidenceScore:             0.7,
		MaxValidationTime:              30,
		EnableCrossReferenceValidation: true,
	}

	validator := classification.NewCrosswalkValidationRules(db, logger, config)

	// Create validation rules
	ctx := context.Background()
	if err := validator.CreateValidationRules(ctx); err != nil {
		b.Fatalf("Failed to create validation rules: %v", err)
	}

	// Benchmark validation
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		summary, err := validator.ValidateCrosswalkMappings(ctx)
		if err != nil {
			b.Fatalf("Failed to validate crosswalk mappings: %v", err)
		}

		if summary == nil {
			b.Fatal("Validation summary is nil")
		}
	}
}
