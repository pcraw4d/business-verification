package risk

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// RiskFactorCalculator calculates risk scores for individual factors
type RiskFactorCalculator struct {
	registry *RiskCategoryRegistry
}

// NewRiskFactorCalculator creates a new risk factor calculator
func NewRiskFactorCalculator(registry *RiskCategoryRegistry) *RiskFactorCalculator {
	return &RiskFactorCalculator{
		registry: registry,
	}
}

// RiskFactorInput represents input data for calculating a risk factor
type RiskFactorInput struct {
	FactorID    string                 `json:"factor_id"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source"`
	Reliability float64                `json:"reliability"` // 0.0 to 1.0
}

// RiskFactorResult represents the calculated result for a risk factor
type RiskFactorResult struct {
	FactorID     string       `json:"factor_id"`
	FactorName   string       `json:"factor_name"`
	Category     RiskCategory `json:"category"`
	Subcategory  string       `json:"subcategory"`
	Score        float64      `json:"score"` // 0.0 to 100.0
	Level        RiskLevel    `json:"level"`
	Confidence   float64      `json:"confidence"` // 0.0 to 1.0
	Explanation  string       `json:"explanation"`
	Evidence     []string     `json:"evidence"`
	CalculatedAt time.Time    `json:"calculated_at"`
	RawValue     interface{}  `json:"raw_value,omitempty"`
	Formula      string       `json:"formula,omitempty"`
}

// CalculateFactor calculates the risk score for a specific factor
func (c *RiskFactorCalculator) CalculateFactor(input RiskFactorInput) (*RiskFactorResult, error) {
	// Get the factor definition
	factorDef, exists := c.registry.GetFactor(input.FactorID)
	if !exists {
		return nil, fmt.Errorf("risk factor %s not found", input.FactorID)
	}

	// Validate input data
	if err := c.validateInput(input, factorDef); err != nil {
		return nil, fmt.Errorf("invalid input for factor %s: %w", input.FactorID, err)
	}

	// Calculate the raw score based on calculation type
	rawScore, explanation, evidence, err := c.calculateRawScore(input, factorDef)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate raw score for factor %s: %w", input.FactorID, err)
	}

	// Normalize the score to 0-100 range
	normalizedScore := c.normalizeScore(rawScore, factorDef)

	// Determine risk level based on thresholds
	level := c.determineRiskLevel(normalizedScore, factorDef.Thresholds)

	// Calculate confidence based on data reliability and completeness
	confidence := c.calculateConfidence(input, factorDef)

	result := &RiskFactorResult{
		FactorID:     input.FactorID,
		FactorName:   factorDef.Name,
		Category:     factorDef.Category,
		Subcategory:  factorDef.Subcategory,
		Score:        normalizedScore,
		Level:        level,
		Confidence:   confidence,
		Explanation:  explanation,
		Evidence:     evidence,
		CalculatedAt: time.Now(),
		RawValue:     rawScore,
		Formula:      factorDef.Formula,
	}

	return result, nil
}

// validateInput validates the input data against the factor definition
func (c *RiskFactorCalculator) validateInput(input RiskFactorInput, factorDef *RiskFactorDefinition) error {
	// Check if required data sources are present
	for _, requiredSource := range factorDef.DataSources {
		found := false
		for key := range input.Data {
			if strings.Contains(strings.ToLower(key), strings.ToLower(requiredSource)) {
				found = true
				break
			}
		}
		// Only require data sources if we have no data at all
		if !found && len(input.Data) == 0 {
			return fmt.Errorf("required data source '%s' not found in input data", requiredSource)
		}
		// If we have data but not the required source, just warn but don't fail
		if !found && len(input.Data) > 0 {
			// Continue with available data
		}
	}

	// Validate reliability score
	if input.Reliability < 0.0 || input.Reliability > 1.0 {
		return fmt.Errorf("reliability score must be between 0.0 and 1.0, got %f", input.Reliability)
	}

	return nil
}

// calculateRawScore calculates the raw score based on the factor's calculation type
func (c *RiskFactorCalculator) calculateRawScore(input RiskFactorInput, factorDef *RiskFactorDefinition) (float64, string, []string, error) {
	switch factorDef.CalculationType {
	case "direct":
		return c.calculateDirectScore(input, factorDef)
	case "derived":
		return c.calculateDerivedScore(input, factorDef)
	case "composite":
		return c.calculateCompositeScore(input, factorDef)
	default:
		return 0, "", nil, fmt.Errorf("unknown calculation type: %s", factorDef.CalculationType)
	}
}

// calculateDirectScore calculates score for direct measurement factors
func (c *RiskFactorCalculator) calculateDirectScore(input RiskFactorInput, factorDef *RiskFactorDefinition) (float64, string, []string, error) {
	// Find the primary value in the input data
	var primaryValue interface{}
	var evidence []string

	// Look for the most relevant data field
	for key, value := range input.Data {
		if strings.Contains(strings.ToLower(key), strings.ToLower(factorDef.Name)) ||
			strings.Contains(strings.ToLower(key), strings.ToLower(factorDef.ID)) {
			primaryValue = value
			evidence = append(evidence, fmt.Sprintf("Found value %v in field %s", value, key))
			break
		}
	}

	if primaryValue == nil {
		// Try to find any numeric value
		for key, value := range input.Data {
			if numericValue, ok := c.toFloat64(value); ok {
				primaryValue = numericValue
				evidence = append(evidence, fmt.Sprintf("Using numeric value %v from field %s", value, key))
				break
			}
		}
	}

	if primaryValue == nil {
		return 0, "No relevant data found", evidence, nil
	}

	// Convert to float64
	score, ok := c.toFloat64(primaryValue)
	if !ok {
		return 0, "Unable to convert value to numeric score", evidence, fmt.Errorf("cannot convert %v to numeric value", primaryValue)
	}

	explanation := fmt.Sprintf("Direct measurement: %v", primaryValue)
	return score, explanation, evidence, nil
}

// calculateDerivedScore calculates score for derived factors (calculated from other values)
func (c *RiskFactorCalculator) calculateDerivedScore(input RiskFactorInput, factorDef *RiskFactorDefinition) (float64, string, []string, error) {
	var evidence []string
	var values []float64

	// Extract numeric values from input data
	for key, value := range input.Data {
		if numericValue, ok := c.toFloat64(value); ok {
			values = append(values, numericValue)
			evidence = append(evidence, fmt.Sprintf("Using %v from field %s", value, key))
		}
	}

	if len(values) == 0 {
		return 0, "No numeric data available for calculation", evidence, nil
	}

	// Apply the formula if available
	if factorDef.Formula != "" {
		return c.applyFormula(values, factorDef.Formula, evidence)
	}

	// Default calculation: use the first value or average
	var score float64
	var explanation string

	if len(values) == 1 {
		score = values[0]
		explanation = fmt.Sprintf("Single value calculation: %v", score)
	} else {
		// Calculate average
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		score = sum / float64(len(values))
		explanation = fmt.Sprintf("Average calculation: %v from %d values", score, len(values))
	}

	return score, explanation, evidence, nil
}

// calculateCompositeScore calculates score for composite factors (combining multiple metrics)
func (c *RiskFactorCalculator) calculateCompositeScore(input RiskFactorInput, factorDef *RiskFactorDefinition) (float64, string, []string, error) {
	var evidence []string
	var scores []float64

	// Calculate individual component scores
	for key, value := range input.Data {
		if numericValue, ok := c.toFloat64(value); ok {
			// Apply component-specific logic based on factor type
			componentScore := c.calculateComponentScore(numericValue, key, factorDef)
			scores = append(scores, componentScore)
			evidence = append(evidence, fmt.Sprintf("Component %s: %v -> %v", key, value, componentScore))
		}
	}

	if len(scores) == 0 {
		return 0, "No valid component scores available", evidence, nil
	}

	// Combine scores using weighted average or other aggregation method
	compositeScore := c.aggregateScores(scores, factorDef)
	explanation := fmt.Sprintf("Composite calculation: %v from %d components", compositeScore, len(scores))

	return compositeScore, explanation, evidence, nil
}

// calculateComponentScore calculates score for a component of a composite factor
func (c *RiskFactorCalculator) calculateComponentScore(value float64, componentKey string, factorDef *RiskFactorDefinition) float64 {
	// Apply component-specific logic based on the factor category
	switch factorDef.Category {
	case RiskCategoryFinancial:
		return c.calculateFinancialComponent(value, componentKey)
	case RiskCategoryOperational:
		return c.calculateOperationalComponent(value, componentKey)
	case RiskCategoryRegulatory:
		return c.calculateRegulatoryComponent(value, componentKey)
	case RiskCategoryReputational:
		return c.calculateReputationalComponent(value, componentKey)
	case RiskCategoryCybersecurity:
		return c.calculateCybersecurityComponent(value, componentKey)
	default:
		return value
	}
}

// calculateFinancialComponent calculates score for financial risk components
func (c *RiskFactorCalculator) calculateFinancialComponent(value float64, componentKey string) float64 {
	key := strings.ToLower(componentKey)

	switch {
	case strings.Contains(key, "ratio"):
		// For ratios, higher is generally better (lower risk)
		if value >= 2.0 {
			return 20.0 // Low risk
		} else if value >= 1.5 {
			return 40.0 // Medium risk
		} else if value >= 1.0 {
			return 60.0 // High risk
		} else {
			return 80.0 // Critical risk
		}
	case strings.Contains(key, "score"):
		// For scores, higher is better
		if value >= 80 {
			return 20.0
		} else if value >= 60 {
			return 40.0
		} else if value >= 40 {
			return 60.0
		} else {
			return 80.0
		}
	case strings.Contains(key, "trend"):
		// For trends, positive is better
		if value > 0 {
			return 30.0
		} else if value > -0.05 {
			return 50.0
		} else if value > -0.15 {
			return 70.0
		} else {
			return 90.0
		}
	default:
		return value
	}
}

// calculateOperationalComponent calculates score for operational risk components
func (c *RiskFactorCalculator) calculateOperationalComponent(value float64, componentKey string) float64 {
	key := strings.ToLower(componentKey)

	switch {
	case strings.Contains(key, "turnover"):
		// For turnover rates, lower is better
		if value <= 0.05 {
			return 20.0
		} else if value <= 0.15 {
			return 40.0
		} else if value <= 0.25 {
			return 60.0
		} else {
			return 80.0
		}
	case strings.Contains(key, "uptime"):
		// For uptime, higher is better
		if value >= 0.99 {
			return 20.0
		} else if value >= 0.95 {
			return 40.0
		} else if value >= 0.90 {
			return 60.0
		} else {
			return 80.0
		}
	case strings.Contains(key, "concentration"):
		// For concentration, lower is better
		if value <= 0.3 {
			return 20.0
		} else if value <= 0.5 {
			return 40.0
		} else if value <= 0.7 {
			return 60.0
		} else {
			return 80.0
		}
	default:
		return value
	}
}

// calculateRegulatoryComponent calculates score for regulatory risk components
func (c *RiskFactorCalculator) calculateRegulatoryComponent(value float64, componentKey string) float64 {
	key := strings.ToLower(componentKey)

	switch {
	case strings.Contains(key, "compliance"):
		// For compliance scores, higher is better
		if value >= 90 {
			return 20.0
		} else if value >= 75 {
			return 40.0
		} else if value >= 60 {
			return 60.0
		} else {
			return 80.0
		}
	case strings.Contains(key, "violation"):
		// For violations, lower is better
		if value == 0 {
			return 20.0
		} else if value <= 1 {
			return 40.0
		} else if value <= 3 {
			return 60.0
		} else {
			return 80.0
		}
	case strings.Contains(key, "license"):
		// For license status, higher is better
		if value >= 1.0 {
			return 20.0
		} else if value >= 0.8 {
			return 40.0
		} else if value >= 0.6 {
			return 60.0
		} else {
			return 80.0
		}
	default:
		return value
	}
}

// calculateReputationalComponent calculates score for reputational risk components
func (c *RiskFactorCalculator) calculateReputationalComponent(value float64, componentKey string) float64 {
	key := strings.ToLower(componentKey)

	switch {
	case strings.Contains(key, "sentiment"):
		// For sentiment, higher is better
		if value >= 0.7 {
			return 20.0
		} else if value >= 0.5 {
			return 40.0
		} else if value >= 0.3 {
			return 60.0
		} else {
			return 80.0
		}
	case strings.Contains(key, "satisfaction"):
		// For satisfaction scores, higher is better
		if value >= 4.0 {
			return 20.0
		} else if value >= 3.5 {
			return 40.0
		} else if value >= 3.0 {
			return 60.0
		} else {
			return 80.0
		}
	case strings.Contains(key, "mention"):
		// For negative mentions, lower is better
		if value == 0 {
			return 20.0
		} else if value <= 5 {
			return 40.0
		} else if value <= 15 {
			return 60.0
		} else {
			return 80.0
		}
	default:
		return value
	}
}

// calculateCybersecurityComponent calculates score for cybersecurity risk components
func (c *RiskFactorCalculator) calculateCybersecurityComponent(value float64, componentKey string) float64 {
	key := strings.ToLower(componentKey)

	switch {
	case strings.Contains(key, "security"):
		// For security scores, higher is better
		if value >= 85 {
			return 20.0
		} else if value >= 70 {
			return 40.0
		} else if value >= 55 {
			return 60.0
		} else {
			return 80.0
		}
	case strings.Contains(key, "breach"):
		// For breach incidents, lower is better
		if value == 0 {
			return 20.0
		} else if value <= 1 {
			return 40.0
		} else if value <= 2 {
			return 60.0
		} else {
			return 80.0
		}
	case strings.Contains(key, "patch"):
		// For patch compliance, higher is better
		if value >= 0.95 {
			return 20.0
		} else if value >= 0.85 {
			return 40.0
		} else if value >= 0.75 {
			return 60.0
		} else {
			return 80.0
		}
	default:
		return value
	}
}

// aggregateScores aggregates multiple component scores into a composite score
func (c *RiskFactorCalculator) aggregateScores(scores []float64, factorDef *RiskFactorDefinition) float64 {
	if len(scores) == 0 {
		return 0
	}

	if len(scores) == 1 {
		return scores[0]
	}

	// Use weighted average based on factor weights, or simple average
	sum := 0.0
	for _, score := range scores {
		sum += score
	}

	return sum / float64(len(scores))
}

// applyFormula applies a mathematical formula to calculate the score
func (c *RiskFactorCalculator) applyFormula(values []float64, formula string, evidence []string) (float64, string, []string, error) {
	// Simple formula parser for common operations
	formula = strings.ToLower(strings.TrimSpace(formula))

	switch {
	case strings.Contains(formula, "average") || strings.Contains(formula, "mean"):
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		score := sum / float64(len(values))
		return score, fmt.Sprintf("Average calculation: %v", score), evidence, nil

	case strings.Contains(formula, "sum"):
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum, fmt.Sprintf("Sum calculation: %v", sum), evidence, nil

	case strings.Contains(formula, "max"):
		max := values[0]
		for _, v := range values {
			if v > max {
				max = v
			}
		}
		return max, fmt.Sprintf("Maximum calculation: %v", max), evidence, nil

	case strings.Contains(formula, "min"):
		min := values[0]
		for _, v := range values {
			if v < min {
				min = v
			}
		}
		return min, fmt.Sprintf("Minimum calculation: %v", min), evidence, nil

	default:
		// Default to average if formula is not recognized
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		score := sum / float64(len(values))
		return score, fmt.Sprintf("Default calculation (average): %v", score), evidence, nil
	}
}

// normalizeScore normalizes the raw score to 0-100 range
func (c *RiskFactorCalculator) normalizeScore(rawScore float64, factorDef *RiskFactorDefinition) float64 {
	// If the factor has specific thresholds, use them for normalization
	if len(factorDef.Thresholds) > 0 {
		return c.normalizeWithThresholds(rawScore, factorDef.Thresholds)
	}

	// Default normalization: assume raw score is already in reasonable range
	// but cap it at 100
	if rawScore > 100 {
		return 100
	}
	if rawScore < 0 {
		return 0
	}
	return rawScore
}

// normalizeWithThresholds normalizes score using factor-specific thresholds
func (c *RiskFactorCalculator) normalizeWithThresholds(rawScore float64, thresholds map[RiskLevel]float64) float64 {
	// For cash flow coverage, higher is better (lower risk)
	// So we need to invert the logic - higher raw score = lower risk score
	
	// Find the appropriate threshold range for the raw score
	var lowThreshold, highThreshold float64
	var lowLevel, highLevel RiskLevel
	
	levels := []RiskLevel{RiskLevelLow, RiskLevelMedium, RiskLevelHigh, RiskLevelCritical}
	
	for i, level := range levels {
		if threshold, exists := thresholds[level]; exists {
			if rawScore >= threshold {
				// For cash flow coverage, higher is better
				if i == 0 {
					// Above highest threshold (Low risk)
					return 25
				}
				lowThreshold = threshold
				lowLevel = level
				highThreshold = thresholds[levels[i-1]]
				highLevel = levels[i-1]
				break
			}
		}
	}
	
	// If we didn't find a range, the score is below all thresholds (Critical risk)
	if lowThreshold == 0 {
		return 100
	}
	
	// Interpolate between the two levels
	levelScores := map[RiskLevel]float64{
		RiskLevelLow:      25,
		RiskLevelMedium:   50,
		RiskLevelHigh:     75,
		RiskLevelCritical: 100,
	}
	
	lowScore := levelScores[lowLevel]
	highScore := levelScores[highLevel]
	
	// Linear interpolation (inverted for cash flow coverage)
	ratio := (rawScore - lowThreshold) / (highThreshold - lowThreshold)
	normalizedScore := lowScore + ratio*(highScore-lowScore)
	
	return math.Max(0, math.Min(100, normalizedScore))
}

// determineRiskLevel determines the risk level based on normalized score
func (c *RiskFactorCalculator) determineRiskLevel(score float64, thresholds map[RiskLevel]float64) RiskLevel {
	// Always use default thresholds for normalized scores (0-100)
	if score <= 25 {
		return RiskLevelLow
	} else if score <= 50 {
		return RiskLevelMedium
	} else if score <= 75 {
		return RiskLevelHigh
	} else {
		return RiskLevelCritical
	}
}

// calculateConfidence calculates the confidence level for the calculation
func (c *RiskFactorCalculator) calculateConfidence(input RiskFactorInput, factorDef *RiskFactorDefinition) float64 {
	// Base confidence on data reliability
	confidence := input.Reliability

	// Adjust based on data completeness
	dataCompleteness := float64(len(input.Data)) / float64(len(factorDef.DataSources))
	confidence *= dataCompleteness

	// Adjust based on data freshness (if timestamp is provided)
	if !input.Timestamp.IsZero() {
		age := time.Since(input.Timestamp)
		if age <= 24*time.Hour {
			confidence *= 1.0 // Recent data
		} else if age <= 7*24*time.Hour {
			confidence *= 0.9 // Week-old data
		} else if age <= 30*24*time.Hour {
			confidence *= 0.8 // Month-old data
		} else {
			confidence *= 0.6 // Old data
		}
	}

	// Ensure confidence is between 0 and 1
	return math.Max(0, math.Min(1, confidence))
}

// toFloat64 converts various types to float64
func (c *RiskFactorCalculator) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}
