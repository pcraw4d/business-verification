package risk

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// ScoringAlgorithm defines the interface for different risk scoring algorithms
type ScoringAlgorithm interface {
	CalculateScore(factors []RiskFactor, data map[string]interface{}) (float64, float64, error)
	CalculateLevel(score float64, thresholds map[RiskLevel]float64) RiskLevel
	CalculateConfidence(factors []RiskFactor, data map[string]interface{}) float64
}

// WeightedScoringAlgorithm implements weighted risk scoring
type WeightedScoringAlgorithm struct{}

// NewWeightedScoringAlgorithm creates a new weighted scoring algorithm
func NewWeightedScoringAlgorithm() *WeightedScoringAlgorithm {
	return &WeightedScoringAlgorithm{}
}

// CalculateScore calculates a weighted risk score based on factors and data
func (w *WeightedScoringAlgorithm) CalculateScore(factors []RiskFactor, data map[string]interface{}) (float64, float64, error) {
	if len(factors) == 0 {
		return 0.0, 0.0, nil
	}

	var totalScore float64
	var totalWeight float64
	var confidence float64

	for _, factor := range factors {
		// Calculate individual factor score
		factorScore, factorConfidence, err := w.calculateFactorScore(factor, data)
		if err != nil {
			continue // Skip factors that can't be calculated
		}

		// Apply weight to the factor score
		weightedScore := factorScore * factor.Weight
		totalScore += weightedScore
		totalWeight += factor.Weight

		// Accumulate confidence (weighted average)
		confidence += factorConfidence * factor.Weight
	}

	if totalWeight == 0 {
		return 0.0, 0.0, nil
	}

	// Normalize score to 0-100 range
	finalScore := totalScore / totalWeight
	finalConfidence := confidence / totalWeight

	return finalScore, finalConfidence, nil
}

// calculateFactorScore calculates the score for a single risk factor
func (w *WeightedScoringAlgorithm) calculateFactorScore(factor RiskFactor, data map[string]interface{}) (float64, float64, error) {
	// Get factor-specific data
	factorData, exists := data[factor.ID]
	if !exists {
		return 0.0, 0.0, nil // No data available for this factor
	}

	// Calculate score based on factor category
	switch factor.Category {
	case RiskCategoryFinancial:
		return w.calculateFinancialScore(factor, factorData)
	case RiskCategoryOperational:
		return w.calculateOperationalScore(factor, factorData)
	case RiskCategoryRegulatory:
		return w.calculateRegulatoryScore(factor, factorData)
	case RiskCategoryReputational:
		return w.calculateReputationalScore(factor, factorData)
	case RiskCategoryCybersecurity:
		return w.calculateCybersecurityScore(factor, factorData)
	default:
		return w.calculateGenericScore(factor, factorData)
	}
}

// calculateFinancialScore calculates financial risk score
func (w *WeightedScoringAlgorithm) calculateFinancialScore(factor RiskFactor, data interface{}) (float64, float64, error) {
	financialData, ok := data.(map[string]interface{})
	if !ok {
		return 0.0, 0.0, nil
	}

	var score float64
	var confidence float64
	var factors []float64

	// Revenue trend
	if revenue, exists := financialData["revenue"].(float64); exists {
		if revenue < 0 {
			factors = append(factors, 90.0) // High risk for negative revenue
		} else if revenue < 100000 {
			factors = append(factors, 70.0) // Medium-high risk for low revenue
		} else if revenue < 1000000 {
			factors = append(factors, 40.0) // Medium risk
		} else {
			factors = append(factors, 20.0) // Low risk
		}
		confidence += 0.8
	}

	// Debt ratio
	if debtRatio, exists := financialData["debt_ratio"].(float64); exists {
		if debtRatio > 0.8 {
			factors = append(factors, 85.0) // Very high risk
		} else if debtRatio > 0.6 {
			factors = append(factors, 65.0) // High risk
		} else if debtRatio > 0.4 {
			factors = append(factors, 45.0) // Medium risk
		} else {
			factors = append(factors, 25.0) // Low risk
		}
		confidence += 0.9
	}

	// Cash flow
	if cashFlow, exists := financialData["cash_flow"].(float64); exists {
		if cashFlow < 0 {
			factors = append(factors, 80.0) // High risk for negative cash flow
		} else if cashFlow < 50000 {
			factors = append(factors, 50.0) // Medium risk
		} else {
			factors = append(factors, 20.0) // Low risk
		}
		confidence += 0.85
	}

	// Profit margin
	if profitMargin, exists := financialData["profit_margin"].(float64); exists {
		if profitMargin < 0 {
			factors = append(factors, 75.0) // High risk for negative margin
		} else if profitMargin < 0.05 {
			factors = append(factors, 55.0) // Medium-high risk
		} else if profitMargin < 0.15 {
			factors = append(factors, 35.0) // Medium risk
		} else {
			factors = append(factors, 15.0) // Low risk
		}
		confidence += 0.8
	}

	if len(factors) == 0 {
		return 0.0, 0.0, nil
	}

	// Calculate average score
	for _, f := range factors {
		score += f
	}
	score = score / float64(len(factors))
	confidence = confidence / float64(len(factors))

	return score, confidence, nil
}

// calculateOperationalScore calculates operational risk score
func (w *WeightedScoringAlgorithm) calculateOperationalScore(factor RiskFactor, data interface{}) (float64, float64, error) {
	operationalData, ok := data.(map[string]interface{})
	if !ok {
		return 0.0, 0.0, nil
	}

	var score float64
	var confidence float64
	var factors []float64

	// Employee turnover
	if turnover, exists := operationalData["employee_turnover"].(float64); exists {
		if turnover > 0.3 {
			factors = append(factors, 80.0) // High risk
		} else if turnover > 0.2 {
			factors = append(factors, 60.0) // Medium-high risk
		} else if turnover > 0.1 {
			factors = append(factors, 40.0) // Medium risk
		} else {
			factors = append(factors, 20.0) // Low risk
		}
		confidence += 0.8
	}

	// Operational efficiency
	if efficiency, exists := operationalData["operational_efficiency"].(float64); exists {
		if efficiency < 0.5 {
			factors = append(factors, 75.0) // High risk
		} else if efficiency < 0.7 {
			factors = append(factors, 55.0) // Medium-high risk
		} else if efficiency < 0.85 {
			factors = append(factors, 35.0) // Medium risk
		} else {
			factors = append(factors, 15.0) // Low risk
		}
		confidence += 0.85
	}

	// Process maturity
	if maturity, exists := operationalData["process_maturity"].(float64); exists {
		if maturity < 2.0 {
			factors = append(factors, 70.0) // High risk
		} else if maturity < 3.0 {
			factors = append(factors, 50.0) // Medium-high risk
		} else if maturity < 4.0 {
			factors = append(factors, 30.0) // Medium risk
		} else {
			factors = append(factors, 10.0) // Low risk
		}
		confidence += 0.8
	}

	if len(factors) == 0 {
		return 0.0, 0.0, nil
	}

	// Calculate average score
	for _, f := range factors {
		score += f
	}
	score = score / float64(len(factors))
	confidence = confidence / float64(len(factors))

	return score, confidence, nil
}

// calculateRegulatoryScore calculates regulatory risk score
func (w *WeightedScoringAlgorithm) calculateRegulatoryScore(factor RiskFactor, data interface{}) (float64, float64, error) {
	regulatoryData, ok := data.(map[string]interface{})
	if !ok {
		return 0.0, 0.0, nil
	}

	var score float64
	var confidence float64
	var factors []float64

	// Compliance violations
	if violations, exists := regulatoryData["compliance_violations"].(float64); exists {
		if violations > 5 {
			factors = append(factors, 90.0) // Very high risk
		} else if violations > 2 {
			factors = append(factors, 70.0) // High risk
		} else if violations > 0 {
			factors = append(factors, 50.0) // Medium risk
		} else {
			factors = append(factors, 10.0) // Low risk
		}
		confidence += 0.9
	}

	// Regulatory fines
	if fines, exists := regulatoryData["regulatory_fines"].(float64); exists {
		if fines > 100000 {
			factors = append(factors, 85.0) // Very high risk
		} else if fines > 10000 {
			factors = append(factors, 65.0) // High risk
		} else if fines > 0 {
			factors = append(factors, 45.0) // Medium risk
		} else {
			factors = append(factors, 10.0) // Low risk
		}
		confidence += 0.9
	}

	// License status
	if licenseStatus, exists := regulatoryData["license_status"].(string); exists {
		switch licenseStatus {
		case "suspended":
			factors = append(factors, 95.0) // Very high risk
		case "expired":
			factors = append(factors, 80.0) // High risk
		case "pending":
			factors = append(factors, 60.0) // Medium-high risk
		case "active":
			factors = append(factors, 10.0) // Low risk
		}
		confidence += 0.95
	}

	if len(factors) == 0 {
		return 0.0, 0.0, nil
	}

	// Calculate average score
	for _, f := range factors {
		score += f
	}
	score = score / float64(len(factors))
	confidence = confidence / float64(len(factors))

	return score, confidence, nil
}

// calculateReputationalScore calculates reputational risk score
func (w *WeightedScoringAlgorithm) calculateReputationalScore(factor RiskFactor, data interface{}) (float64, float64, error) {
	reputationalData, ok := data.(map[string]interface{})
	if !ok {
		return 0.0, 0.0, nil
	}

	var score float64
	var confidence float64
	var factors []float64

	// Customer satisfaction
	if satisfaction, exists := reputationalData["customer_satisfaction"].(float64); exists {
		if satisfaction < 0.3 {
			factors = append(factors, 85.0) // Very high risk
		} else if satisfaction < 0.5 {
			factors = append(factors, 65.0) // High risk
		} else if satisfaction < 0.7 {
			factors = append(factors, 45.0) // Medium risk
		} else {
			factors = append(factors, 15.0) // Low risk
		}
		confidence += 0.8
	}

	// Negative reviews
	if negativeReviews, exists := reputationalData["negative_reviews"].(float64); exists {
		if negativeReviews > 0.5 {
			factors = append(factors, 80.0) // High risk
		} else if negativeReviews > 0.3 {
			factors = append(factors, 60.0) // Medium-high risk
		} else if negativeReviews > 0.1 {
			factors = append(factors, 40.0) // Medium risk
		} else {
			factors = append(factors, 20.0) // Low risk
		}
		confidence += 0.85
	}

	// Media sentiment
	if sentiment, exists := reputationalData["media_sentiment"].(float64); exists {
		if sentiment < -0.5 {
			factors = append(factors, 75.0) // High risk
		} else if sentiment < -0.2 {
			factors = append(factors, 55.0) // Medium-high risk
		} else if sentiment < 0.2 {
			factors = append(factors, 35.0) // Medium risk
		} else {
			factors = append(factors, 15.0) // Low risk
		}
		confidence += 0.8
	}

	if len(factors) == 0 {
		return 0.0, 0.0, nil
	}

	// Calculate average score
	for _, f := range factors {
		score += f
	}
	score = score / float64(len(factors))
	confidence = confidence / float64(len(factors))

	return score, confidence, nil
}

// calculateCybersecurityScore calculates cybersecurity risk score
func (w *WeightedScoringAlgorithm) calculateCybersecurityScore(factor RiskFactor, data interface{}) (float64, float64, error) {
	cyberData, ok := data.(map[string]interface{})
	if !ok {
		return 0.0, 0.0, nil
	}

	var score float64
	var confidence float64
	var factors []float64

	// Security incidents
	if incidents, exists := cyberData["security_incidents"].(float64); exists {
		if incidents > 10 {
			factors = append(factors, 90.0) // Very high risk
		} else if incidents > 5 {
			factors = append(factors, 70.0) // High risk
		} else if incidents > 1 {
			factors = append(factors, 50.0) // Medium risk
		} else {
			factors = append(factors, 20.0) // Low risk
		}
		confidence += 0.9
	}

	// Data breaches
	if breaches, exists := cyberData["data_breaches"].(float64); exists {
		if breaches > 0 {
			factors = append(factors, 95.0) // Very high risk
		} else {
			factors = append(factors, 10.0) // Low risk
		}
		confidence += 0.95
	}

	// Security maturity
	if maturity, exists := cyberData["security_maturity"].(float64); exists {
		if maturity < 2.0 {
			factors = append(factors, 80.0) // High risk
		} else if maturity < 3.0 {
			factors = append(factors, 60.0) // Medium-high risk
		} else if maturity < 4.0 {
			factors = append(factors, 40.0) // Medium risk
		} else {
			factors = append(factors, 20.0) // Low risk
		}
		confidence += 0.85
	}

	if len(factors) == 0 {
		return 0.0, 0.0, nil
	}

	// Calculate average score
	for _, f := range factors {
		score += f
	}
	score = score / float64(len(factors))
	confidence = confidence / float64(len(factors))

	return score, confidence, nil
}

// calculateGenericScore calculates a generic risk score
func (w *WeightedScoringAlgorithm) calculateGenericScore(factor RiskFactor, data interface{}) (float64, float64, error) {
	// Generic scoring based on factor metadata
	var score float64
	var confidence float64

	// Check if data contains a direct score
	if scoreData, ok := data.(map[string]interface{}); ok {
		if directScore, exists := scoreData["score"].(float64); exists {
			score = directScore
			confidence = 0.8
		}
	}

	return score, confidence, nil
}

// CalculateLevel determines the risk level based on score and thresholds
func (w *WeightedScoringAlgorithm) CalculateLevel(score float64, thresholds map[RiskLevel]float64) RiskLevel {
	if len(thresholds) == 0 {
		// Default thresholds if none provided
		thresholds = map[RiskLevel]float64{
			RiskLevelLow:      25.0,
			RiskLevelMedium:   50.0,
			RiskLevelHigh:     75.0,
			RiskLevelCritical: 90.0,
		}
	}

	if score >= thresholds[RiskLevelCritical] {
		return RiskLevelCritical
	} else if score >= thresholds[RiskLevelHigh] {
		return RiskLevelHigh
	} else if score >= thresholds[RiskLevelMedium] {
		return RiskLevelMedium
	} else {
		return RiskLevelLow
	}
}

// CalculateConfidence calculates overall confidence based on data availability and quality
func (w *WeightedScoringAlgorithm) CalculateConfidence(factors []RiskFactor, data map[string]interface{}) float64 {
	if len(factors) == 0 {
		return 0.0
	}

	var totalConfidence float64
	var validFactors int

	for _, factor := range factors {
		if _, exists := data[factor.ID]; exists {
			// Data exists for this factor
			totalConfidence += 0.8
			validFactors++
		} else {
			// No data for this factor
			totalConfidence += 0.2
			validFactors++
		}
	}

	if validFactors == 0 {
		return 0.0
	}

	return totalConfidence / float64(validFactors)
}

// TrendAnalysisAlgorithm analyzes risk trends over time
type TrendAnalysisAlgorithm struct{}

// NewTrendAnalysisAlgorithm creates a new trend analysis algorithm
func NewTrendAnalysisAlgorithm() *TrendAnalysisAlgorithm {
	return &TrendAnalysisAlgorithm{}
}

// AnalyzeTrend analyzes risk trends and predicts future scores
func (t *TrendAnalysisAlgorithm) AnalyzeTrend(trends []RiskTrend, horizon time.Duration) (float64, float64, error) {
	if len(trends) < 2 {
		return 0.0, 0.0, nil
	}

	// Sort trends by time
	sort.Slice(trends, func(i, j int) bool {
		return trends[i].RecordedAt.Before(trends[j].RecordedAt)
	})

	// Calculate trend slope using linear regression
	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(trends))

	for i, trend := range trends {
		x := float64(i)
		y := trend.Score
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope (rate of change)
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// Predict future score
	lastScore := trends[len(trends)-1].Score
	timeSteps := horizon.Hours() / 24.0 // Convert to days
	predictedScore := lastScore + (slope * timeSteps)

	// Cap predicted score between 0 and 100
	if predictedScore < 0 {
		predictedScore = 0
	} else if predictedScore > 100 {
		predictedScore = 100
	}

	// Calculate confidence based on trend consistency
	var variance float64
	mean := sumY / n
	for _, trend := range trends {
		variance += math.Pow(trend.Score-mean, 2)
	}
	variance /= n

	// Lower variance = higher confidence
	confidence := math.Max(0.1, 1.0-math.Sqrt(variance)/100.0)

	return predictedScore, confidence, nil
}

// CompositeScoringAlgorithm combines multiple scoring methods
type CompositeScoringAlgorithm struct {
	algorithms []ScoringAlgorithm
	weights    []float64
}

// NewCompositeScoringAlgorithm creates a new composite scoring algorithm
func NewCompositeScoringAlgorithm(algorithms []ScoringAlgorithm, weights []float64) *CompositeScoringAlgorithm {
	return &CompositeScoringAlgorithm{
		algorithms: algorithms,
		weights:    weights,
	}
}

// CalculateScore calculates a composite score using multiple algorithms
func (c *CompositeScoringAlgorithm) CalculateScore(factors []RiskFactor, data map[string]interface{}) (float64, float64, error) {
	if len(c.algorithms) == 0 {
		return 0.0, 0.0, nil
	}

	var totalScore float64
	var totalConfidence float64
	var totalWeight float64

	for i, algorithm := range c.algorithms {
		weight := 1.0
		if i < len(c.weights) {
			weight = c.weights[i]
		}

		score, confidence, err := algorithm.CalculateScore(factors, data)
		if err != nil {
			continue
		}

		totalScore += score * weight
		totalConfidence += confidence * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0, 0.0, nil
	}

	return totalScore / totalWeight, totalConfidence / totalWeight, nil
}

// CalculateLevel determines risk level using the first algorithm's thresholds
func (c *CompositeScoringAlgorithm) CalculateLevel(score float64, thresholds map[RiskLevel]float64) RiskLevel {
	if len(c.algorithms) == 0 {
		return RiskLevelLow
	}
	return c.algorithms[0].CalculateLevel(score, thresholds)
}

// CalculateConfidence calculates composite confidence
func (c *CompositeScoringAlgorithm) CalculateConfidence(factors []RiskFactor, data map[string]interface{}) float64 {
	if len(c.algorithms) == 0 {
		return 0.0
	}

	var totalConfidence float64
	var totalWeight float64

	for i, algorithm := range c.algorithms {
		weight := 1.0
		if i < len(c.weights) {
			weight = c.weights[i]
		}

		confidence := algorithm.CalculateConfidence(factors, data)
		totalConfidence += confidence * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalConfidence / totalWeight
}

// RiskPredictionAlgorithm implements risk prediction models
type RiskPredictionAlgorithm struct {
	trendAnalyzer *TrendAnalysisAlgorithm
}

// NewRiskPredictionAlgorithm creates a new risk prediction algorithm
func NewRiskPredictionAlgorithm() *RiskPredictionAlgorithm {
	return &RiskPredictionAlgorithm{
		trendAnalyzer: NewTrendAnalysisAlgorithm(),
	}
}

// PredictRiskScore predicts future risk scores based on historical data
func (p *RiskPredictionAlgorithm) PredictRiskScore(
	historicalTrends []RiskTrend,
	horizon time.Duration,
	confidenceThreshold float64,
) (*RiskPrediction, error) {
	if len(historicalTrends) < 3 {
		return nil, fmt.Errorf("insufficient historical data for prediction (minimum 3 data points required)")
	}

	// Analyze trend to get base prediction
	predictedScore, confidence, err := p.trendAnalyzer.AnalyzeTrend(historicalTrends, horizon)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze trend: %w", err)
	}

	// Apply confidence threshold
	if confidence < confidenceThreshold {
		confidence = confidenceThreshold * 0.5 // Reduce confidence for low-quality predictions
	}

	// Determine predicted risk level
	level := p.determinePredictedLevel(predictedScore)

	// Calculate contributing factors
	factors := p.identifyContributingFactors(historicalTrends, predictedScore)

	// Generate prediction ID
	predictionID := fmt.Sprintf("prediction_%d", time.Now().UnixNano())

	prediction := &RiskPrediction{
		ID:             predictionID,
		BusinessID:     historicalTrends[0].BusinessID,
		FactorID:       "composite_risk",
		PredictedScore: predictedScore,
		PredictedLevel: level,
		Confidence:     confidence,
		Horizon:        p.formatHorizon(horizon),
		PredictedAt:    time.Now(),
		Factors:        factors,
	}

	return prediction, nil
}

// PredictMultipleHorizons predicts risk scores for multiple time horizons
func (p *RiskPredictionAlgorithm) PredictMultipleHorizons(
	historicalTrends []RiskTrend,
	horizons []time.Duration,
) ([]RiskPrediction, error) {
	var predictions []RiskPrediction

	for _, horizon := range horizons {
		prediction, err := p.PredictRiskScore(historicalTrends, horizon, 0.7)
		if err != nil {
			continue // Skip predictions that can't be calculated
		}
		predictions = append(predictions, *prediction)
	}

	return predictions, nil
}

// determinePredictedLevel determines the risk level for a predicted score
func (p *RiskPredictionAlgorithm) determinePredictedLevel(score float64) RiskLevel {
	switch {
	case score < 25:
		return RiskLevelLow
	case score < 50:
		return RiskLevelMedium
	case score < 75:
		return RiskLevelHigh
	default:
		return RiskLevelCritical
	}
}

// identifyContributingFactors identifies the main factors contributing to the prediction
func (p *RiskPredictionAlgorithm) identifyContributingFactors(trends []RiskTrend, predictedScore float64) []string {
	var factors []string

	// Analyze recent trends to identify contributing factors
	if len(trends) >= 2 {
		recentChange := trends[len(trends)-1].Score - trends[len(trends)-2].Score
		
		if recentChange > 10 {
			factors = append(factors, "accelerating_risk_trend")
		} else if recentChange < -10 {
			factors = append(factors, "improving_risk_profile")
		}

		// Identify category-specific factors
		categoryCounts := make(map[RiskCategory]int)
		for _, trend := range trends {
			categoryCounts[trend.Category]++
		}

		// Find dominant category
		var dominantCategory RiskCategory
		maxCount := 0
		for category, count := range categoryCounts {
			if count > maxCount {
				maxCount = count
				dominantCategory = category
			}
		}

		if maxCount > 0 {
			factors = append(factors, string(dominantCategory)+"_dominant")
		}
	}

	// Add prediction confidence factor
	if predictedScore > 80 {
		factors = append(factors, "high_risk_prediction")
	} else if predictedScore < 20 {
		factors = append(factors, "low_risk_prediction")
	}

	return factors
}

// formatHorizon formats the prediction horizon as a string
func (p *RiskPredictionAlgorithm) formatHorizon(horizon time.Duration) string {
	days := int(horizon.Hours() / 24)
	
	switch {
	case days <= 30:
		return "1month"
	case days <= 90:
		return "3months"
	case days <= 180:
		return "6months"
	case days <= 365:
		return "1year"
	default:
		return fmt.Sprintf("%dmonths", days/30)
	}
}

// ConfidenceInterval represents the confidence interval for a risk prediction
type ConfidenceInterval struct {
	LowerBound float64 `json:"lower_bound"`
	UpperBound float64 `json:"upper_bound"`
	Confidence float64 `json:"confidence"` // e.g., 0.95 for 95% confidence interval
}

// PredictRiskScoreWithConfidenceInterval predicts future risk scores with confidence intervals
func (p *RiskPredictionAlgorithm) PredictRiskScoreWithConfidenceInterval(
	historicalTrends []RiskTrend,
	horizon time.Duration,
	confidenceLevel float64, // e.g., 0.95 for 95% confidence interval
) (*RiskPrediction, *ConfidenceInterval, error) {
	if len(historicalTrends) < 3 {
		return nil, nil, fmt.Errorf("insufficient historical data for prediction (minimum 3 data points required)")
	}

	// Calculate base prediction
	prediction, err := p.PredictRiskScore(historicalTrends, horizon, 0.7)
	if err != nil {
		return nil, nil, err
	}

	// Calculate confidence interval
	interval, err := p.calculateConfidenceInterval(historicalTrends, prediction.PredictedScore, confidenceLevel)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to calculate confidence interval: %w", err)
	}

	return prediction, interval, nil
}

// calculateConfidenceInterval calculates the confidence interval for a prediction
func (p *RiskPredictionAlgorithm) calculateConfidenceInterval(
	trends []RiskTrend,
	predictedScore float64,
	confidenceLevel float64,
) (*ConfidenceInterval, error) {
	if len(trends) < 3 {
		return nil, fmt.Errorf("insufficient data for confidence interval calculation")
	}

	// Calculate standard deviation of historical scores
	var scores []float64
	for _, trend := range trends {
		scores = append(scores, trend.Score)
	}

	mean := p.calculateMean(scores)
	stdDev := p.calculateStandardDeviation(scores, mean)

	// Calculate margin of error using t-distribution approximation
	// For small samples, we use a simplified approach
	marginOfError := p.calculateMarginOfError(stdDev, len(scores), confidenceLevel)

	// Calculate bounds
	lowerBound := predictedScore - marginOfError
	upperBound := predictedScore + marginOfError

	// Ensure bounds are within valid range (0-100)
	if lowerBound < 0 {
		lowerBound = 0
	}
	if upperBound > 100 {
		upperBound = 100
	}

	return &ConfidenceInterval{
		LowerBound: lowerBound,
		UpperBound: upperBound,
		Confidence: confidenceLevel,
	}, nil
}

// calculateMean calculates the mean of a slice of float64 values
func (p *RiskPredictionAlgorithm) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

// calculateStandardDeviation calculates the standard deviation of a slice of float64 values
func (p *RiskPredictionAlgorithm) calculateStandardDeviation(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	sumSquaredDiff := 0.0
	for _, value := range values {
		diff := value - mean
		sumSquaredDiff += diff * diff
	}

	variance := sumSquaredDiff / float64(len(values)-1)
	return math.Sqrt(variance)
}

// calculateMarginOfError calculates the margin of error for a confidence interval
func (p *RiskPredictionAlgorithm) calculateMarginOfError(stdDev float64, sampleSize int, confidenceLevel float64) float64 {
	// Simplified t-distribution approximation
	// For 95% confidence level, use 2.0 as the critical value
	// For 90% confidence level, use 1.645
	// For 99% confidence level, use 2.576
	
	var criticalValue float64
	switch {
	case confidenceLevel >= 0.99:
		criticalValue = 2.576
	case confidenceLevel >= 0.95:
		criticalValue = 2.0
	case confidenceLevel >= 0.90:
		criticalValue = 1.645
	default:
		criticalValue = 1.96 // Default to 95% confidence
	}

	// Standard error = standard deviation / sqrt(sample size)
	standardError := stdDev / math.Sqrt(float64(sampleSize))
	
	// Margin of error = critical value * standard error
	return criticalValue * standardError
}
