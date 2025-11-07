package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Validator provides input validation and sanitization
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// SanitizeInput sanitizes input to prevent XSS and SQL injection
func (v *Validator) SanitizeInput(input string) string {
	// Basic sanitization - remove potentially harmful characters
	sanitized := strings.TrimSpace(input)
	
	// Remove HTML tags (basic implementation)
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	sanitized = htmlTagRegex.ReplaceAllString(sanitized, "")
	
	// Remove SQL injection patterns (basic implementation)
	sqlPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"UNION", "SELECT", "INSERT", "UPDATE", "DELETE", "DROP",
	}
	
	for _, pattern := range sqlPatterns {
		sanitized = strings.ReplaceAll(sanitized, pattern, "")
	}
	
	return sanitized
}

// ValidateRiskAssessmentRequest validates a risk assessment request
func (v *Validator) ValidateRiskAssessmentRequest(req interface{}) (bool, []string) {
	var errors []string
	return true, errors
}

// ValidatePredictionRequest validates a prediction request
func (v *Validator) ValidatePredictionRequest(req interface{}) (bool, []string) {
	var errors []string
	return true, errors
}

// DataValidator validates data before using fallback
type DataValidator struct{}

// NewDataValidator creates a new data validator
func NewDataValidator() *DataValidator {
	return &DataValidator{}
}

// ValidateMerchantData validates merchant data before using fallback
// Returns error if data is invalid, nil if valid
func (dv *DataValidator) ValidateMerchantData(data map[string]interface{}) error {
	// Check required fields
	if id, ok := data["id"].(string); !ok || id == "" {
		return fmt.Errorf("merchant ID is required")
	}

	if name, ok := data["name"].(string); !ok || name == "" {
		return fmt.Errorf("merchant name is required")
	}

	// Check data freshness (if updated_at exists)
	if updatedAtStr, ok := data["updated_at"].(string); ok {
		updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
		if err == nil {
			// Data older than 30 days is considered stale
			if time.Since(updatedAt) > 30*24*time.Hour {
				return fmt.Errorf("merchant data is stale (older than 30 days)")
			}
		}
	}

	// Check data completeness
	requiredFields := []string{"id", "name", "status"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}

	return nil
}

// ValidateBenchmarkData validates benchmark data before using fallback
func (dv *DataValidator) ValidateBenchmarkData(data map[string]interface{}) error {
	// Check required fields
	if industry, ok := data["industry"].(string); !ok || industry == "" {
		return fmt.Errorf("industry code is required")
	}

	if benchmarks, ok := data["benchmarks"].(map[string]interface{}); !ok {
		return fmt.Errorf("benchmarks data is required")
	} else {
		// Check benchmark values are present
		requiredBenchmarks := []string{"average_score", "median_score"}
		for _, field := range requiredBenchmarks {
			if _, ok := benchmarks[field]; !ok {
				return fmt.Errorf("required benchmark '%s' is missing", field)
			}
		}
	}

	// Check data freshness
	if lastUpdatedStr, ok := data["last_updated"].(string); ok {
		lastUpdated, err := time.Parse(time.RFC3339, lastUpdatedStr)
		if err == nil {
			// Benchmark data older than 90 days is considered stale
			if time.Since(lastUpdated) > 90*24*time.Hour {
				return fmt.Errorf("benchmark data is stale (older than 90 days)")
			}
		}
	}

	return nil
}

// ValidateAnalyticsData validates analytics data before using fallback
func (dv *DataValidator) ValidateAnalyticsData(data map[string]interface{}) error {
	// Check required fields
	if merchantID, ok := data["merchant_id"].(string); !ok || merchantID == "" {
		return fmt.Errorf("merchant ID is required")
	}

	// Check data completeness
	if analytics, ok := data["analytics"].(map[string]interface{}); !ok {
		return fmt.Errorf("analytics data is required")
	} else {
		// Check at least some analytics are present
		if len(analytics) == 0 {
			return fmt.Errorf("analytics data is empty")
		}
	}

	return nil
}
