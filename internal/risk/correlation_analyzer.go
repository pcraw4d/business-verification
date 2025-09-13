package risk

import (
	"context"
	"fmt"
	"math"
	"strings"

	"go.uber.org/zap"
)

// CorrelationAnalyzer analyzes correlations between risk factors
type CorrelationAnalyzer struct {
	logger *zap.Logger
}

// AnalyzeCorrelations analyzes correlations between the target factor and related factors
// AnalyzeCorrelation analyzes correlation between risk factors
func (ca *CorrelationAnalyzer) AnalyzeCorrelation(ctx context.Context, factorData [][]float64, factorNames []string) (map[string]float64, error) {
	correlations := make(map[string]float64)

	if len(factorData) != len(factorNames) {
		return nil, fmt.Errorf("factor data length (%d) does not match factor names length (%d)", len(factorData), len(factorNames))
	}

	// Calculate correlations between all pairs of factors
	for i := 0; i < len(factorNames); i++ {
		for j := i + 1; j < len(factorNames); j++ {
			if len(factorData[i]) > 0 && len(factorData[j]) > 0 {
				correlation := ca.calculatePearsonCorrelation(factorData[i][0], factorData[j][0])
				key := fmt.Sprintf("%s_%s", factorNames[i], factorNames[j])
				correlations[key] = correlation
			}
		}
	}

	return correlations, nil
}

func (ca *CorrelationAnalyzer) AnalyzeCorrelations(targetFactorID string, relatedFactors []string, data map[string]interface{}) (*CorrelationAnalysis, error) {
	if len(relatedFactors) == 0 {
		return &CorrelationAnalysis{
			CorrelatedFactors:   []CorrelatedFactor{},
			MaxCorrelation:      0,
			AvgCorrelation:      0,
			CorrelationStrength: "none",
		}, nil
	}

	var correlatedFactors []CorrelatedFactor
	var correlations []float64

	// Extract target factor value
	targetValue, err := ca.extractNumericValue(data, targetFactorID)
	if err != nil {
		ca.logger.Warn("Could not extract target factor value",
			zap.String("factor_id", targetFactorID),
			zap.Error(err))
		return &CorrelationAnalysis{
			CorrelatedFactors:   []CorrelatedFactor{},
			MaxCorrelation:      0,
			AvgCorrelation:      0,
			CorrelationStrength: "none",
		}, nil
	}

	// Analyze correlations with each related factor
	for _, relatedFactorID := range relatedFactors {
		correlation, significance, relationship, err := ca.calculateCorrelation(targetFactorID, targetValue, relatedFactorID, data)
		if err != nil {
			ca.logger.Warn("Failed to calculate correlation",
				zap.String("target_factor", targetFactorID),
				zap.String("related_factor", relatedFactorID),
				zap.Error(err))
			continue
		}

		// Only include significant correlations
		if math.Abs(correlation) > 0.1 && significance > 0.05 {
			correlatedFactor := CorrelatedFactor{
				FactorID:     relatedFactorID,
				FactorName:   ca.getFactorName(relatedFactorID),
				Correlation:  correlation,
				Significance: significance,
				Relationship: relationship,
			}
			correlatedFactors = append(correlatedFactors, correlatedFactor)
			correlations = append(correlations, math.Abs(correlation))
		}
	}

	// Calculate summary statistics
	maxCorrelation := 0.0
	avgCorrelation := 0.0
	correlationStrength := "none"

	if len(correlations) > 0 {
		maxCorrelation = ca.calculateMax(correlations)
		avgCorrelation = ca.calculateMean(correlations)
		correlationStrength = ca.determineCorrelationStrength(avgCorrelation)
	}

	return &CorrelationAnalysis{
		CorrelatedFactors:   correlatedFactors,
		MaxCorrelation:      maxCorrelation,
		AvgCorrelation:      avgCorrelation,
		CorrelationStrength: correlationStrength,
	}, nil
}

// calculateCorrelation calculates correlation between two factors
func (ca *CorrelationAnalyzer) calculateCorrelation(targetFactorID string, targetValue float64, relatedFactorID string, data map[string]interface{}) (correlation, significance float64, relationship string, err error) {
	// Extract related factor value
	relatedValue, err := ca.extractNumericValue(data, relatedFactorID)
	if err != nil {
		return 0, 0, "", err
	}

	// Calculate correlation coefficient
	correlation = ca.calculatePearsonCorrelation(targetValue, relatedValue)

	// Calculate significance (simplified)
	significance = ca.calculateSignificance(correlation)

	// Determine relationship type
	relationship = ca.determineRelationship(correlation)

	return correlation, significance, relationship, nil
}

// extractNumericValue extracts numeric value from data for a factor
func (ca *CorrelationAnalyzer) extractNumericValue(data map[string]interface{}, factorID string) (float64, error) {
	// Try direct key match first
	if value, exists := data[factorID]; exists {
		return ca.convertToFloat64(value)
	}

	// Try case-insensitive search
	factorIDLower := strings.ToLower(factorID)
	for key, value := range data {
		if strings.ToLower(key) == factorIDLower {
			return ca.convertToFloat64(value)
		}
	}

	// Try partial match
	for key, value := range data {
		if strings.Contains(strings.ToLower(key), strings.ToLower(factorID)) {
			return ca.convertToFloat64(value)
		}
	}

	// Try to find any numeric value if factor not found
	for _, value := range data {
		if numericValue, err := ca.convertToFloat64(value); err == nil {
			return numericValue, nil
		}
	}

	return 0, fmt.Errorf("no numeric value found for factor %s", factorID)
}

// convertToFloat64 converts various types to float64
func (ca *CorrelationAnalyzer) convertToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		// Try to parse as float
		if f, err := ca.parseFloat(v); err == nil {
			return f, nil
		}
		// Try to parse as percentage
		if strings.Contains(v, "%") {
			clean := strings.TrimSpace(strings.Replace(v, "%", "", -1))
			if f, err := ca.parseFloat(clean); err == nil {
				return f / 100.0, nil
			}
		}
		return 0, fmt.Errorf("cannot convert string to float: %s", v)
	default:
		return 0, fmt.Errorf("cannot convert type %T to float64", value)
	}
}

// parseFloat parses a string to float64
func (ca *CorrelationAnalyzer) parseFloat(s string) (float64, error) {
	// Remove common formatting
	s = strings.TrimSpace(s)
	s = strings.Replace(s, ",", "", -1)
	s = strings.Replace(s, "$", "", -1)

	// Try to parse
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

// calculatePearsonCorrelation calculates Pearson correlation coefficient
func (ca *CorrelationAnalyzer) calculatePearsonCorrelation(x, y float64) float64 {
	// For single data points, we can't calculate correlation
	// This is a simplified version that assumes we have some context
	// In a real implementation, you'd need multiple data points

	// For now, we'll use a heuristic based on the values
	// This is a placeholder - in practice you'd need historical data

	// Simple heuristic: if both values are in similar ranges, assume positive correlation
	xNormalized := ca.normalizeValue(x)
	yNormalized := ca.normalizeValue(y)

	// Calculate similarity
	diff := math.Abs(xNormalized - yNormalized)
	similarity := 1.0 - diff

	// Convert to correlation coefficient
	correlation := (similarity - 0.5) * 2.0

	return math.Max(-1.0, math.Min(1.0, correlation))
}

// normalizeValue normalizes a value to 0-1 range
func (ca *CorrelationAnalyzer) normalizeValue(value float64) float64 {
	// Simple normalization - in practice you'd use proper scaling
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 1.0
	}
	return value / 100.0
}

// calculateSignificance calculates significance of correlation
func (ca *CorrelationAnalyzer) calculateSignificance(correlation float64) float64 {
	// Simplified significance calculation
	// In practice, you'd use proper statistical tests

	absCorrelation := math.Abs(correlation)

	if absCorrelation > 0.8 {
		return 0.01 // Highly significant
	} else if absCorrelation > 0.6 {
		return 0.05 // Significant
	} else if absCorrelation > 0.4 {
		return 0.1 // Marginally significant
	} else {
		return 0.5 // Not significant
	}
}

// determineRelationship determines the type of relationship
func (ca *CorrelationAnalyzer) determineRelationship(correlation float64) string {
	absCorrelation := math.Abs(correlation)

	if absCorrelation < 0.1 {
		return "none"
	} else if absCorrelation < 0.3 {
		return "weak"
	} else if absCorrelation < 0.7 {
		return "moderate"
	} else {
		return "strong"
	}
}

// determineCorrelationStrength determines overall correlation strength
func (ca *CorrelationAnalyzer) determineCorrelationStrength(avgCorrelation float64) string {
	if avgCorrelation < 0.1 {
		return "none"
	} else if avgCorrelation < 0.3 {
		return "weak"
	} else if avgCorrelation < 0.5 {
		return "moderate"
	} else if avgCorrelation < 0.7 {
		return "strong"
	} else {
		return "very_strong"
	}
}

// getFactorName gets a human-readable name for a factor
func (ca *CorrelationAnalyzer) getFactorName(factorID string) string {
	// Convert factor ID to readable name
	name := strings.ReplaceAll(factorID, "_", " ")
	name = strings.Title(name)
	return name
}

// calculateMean calculates mean of values
func (ca *CorrelationAnalyzer) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateMax calculates maximum of values
func (ca *CorrelationAnalyzer) calculateMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

// AnalyzeFactorDependencies analyzes dependencies between multiple factors
func (ca *CorrelationAnalyzer) AnalyzeFactorDependencies(factors []string, data map[string]interface{}) (*FactorDependencyAnalysis, error) {
	if len(factors) < 2 {
		return &FactorDependencyAnalysis{
			Dependencies:  []FactorDependency{},
			MaxDependency: 0,
			AvgDependency: 0,
		}, nil
	}

	var dependencies []FactorDependency
	var dependencyStrengths []float64

	// Analyze dependencies between all pairs of factors
	for i := 0; i < len(factors); i++ {
		for j := i + 1; j < len(factors); j++ {
			factor1 := factors[i]
			factor2 := factors[j]

			dependency, strength, err := ca.calculateFactorDependency(factor1, factor2, data)
			if err != nil {
				ca.logger.Warn("Failed to calculate factor dependency",
					zap.String("factor1", factor1),
					zap.String("factor2", factor2),
					zap.Error(err))
				continue
			}

			if strength > 0.1 { // Only include significant dependencies
				dependencies = append(dependencies, dependency)
				dependencyStrengths = append(dependencyStrengths, strength)
			}
		}
	}

	// Calculate summary statistics
	maxDependency := 0.0
	avgDependency := 0.0

	if len(dependencyStrengths) > 0 {
		maxDependency = ca.calculateMax(dependencyStrengths)
		avgDependency = ca.calculateMean(dependencyStrengths)
	}

	return &FactorDependencyAnalysis{
		Dependencies:  dependencies,
		MaxDependency: maxDependency,
		AvgDependency: avgDependency,
	}, nil
}

// FactorDependencyAnalysis contains factor dependency analysis results
type FactorDependencyAnalysis struct {
	Dependencies  []FactorDependency `json:"dependencies"`
	MaxDependency float64            `json:"max_dependency"`
	AvgDependency float64            `json:"avg_dependency"`
}

// FactorDependency represents a dependency between two factors
type FactorDependency struct {
	Factor1ID    string  `json:"factor1_id"`
	Factor1Name  string  `json:"factor1_name"`
	Factor2ID    string  `json:"factor2_id"`
	Factor2Name  string  `json:"factor2_name"`
	Dependency   float64 `json:"dependency"`
	Strength     float64 `json:"strength"`
	Relationship string  `json:"relationship"`
}

// calculateFactorDependency calculates dependency between two factors
func (ca *CorrelationAnalyzer) calculateFactorDependency(factor1ID, factor2ID string, data map[string]interface{}) (FactorDependency, float64, error) {
	// Extract values for both factors
	value1, err := ca.extractNumericValue(data, factor1ID)
	if err != nil {
		return FactorDependency{}, 0, err
	}

	value2, err := ca.extractNumericValue(data, factor2ID)
	if err != nil {
		return FactorDependency{}, 0, err
	}

	// Calculate correlation
	correlation := ca.calculatePearsonCorrelation(value1, value2)
	strength := math.Abs(correlation)
	relationship := ca.determineRelationship(correlation)

	dependency := FactorDependency{
		Factor1ID:    factor1ID,
		Factor1Name:  ca.getFactorName(factor1ID),
		Factor2ID:    factor2ID,
		Factor2Name:  ca.getFactorName(factor2ID),
		Dependency:   correlation,
		Strength:     strength,
		Relationship: relationship,
	}

	return dependency, strength, nil
}
