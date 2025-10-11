package models

import (
	"math"
	"math/rand"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// TemporalFeatureBuilder builds time-series sequences for LSTM models
type TemporalFeatureBuilder struct {
	sequenceLength int
	logger         *zap.Logger
}

// NewTemporalFeatureBuilder creates a new temporal feature builder
func NewTemporalFeatureBuilder() *TemporalFeatureBuilder {
	return &TemporalFeatureBuilder{
		sequenceLength: 12,           // 12 months of history
		logger:         zap.NewNop(), // Will be set by the model
	}
}

// BuildSequence builds a temporal sequence for LSTM input
func (tfb *TemporalFeatureBuilder) BuildSequence(business *models.RiskAssessmentRequest, sequenceLength int) ([][]float64, error) {
	// For now, generate synthetic temporal data
	// In a full implementation, this would:
	// 1. Try to fetch real historical data from cache/DB
	// 2. Generate synthetic data based on business characteristics
	// 3. Blend real and synthetic data appropriately

	sequence := make([][]float64, sequenceLength)

	// Generate synthetic time-series data
	for i := 0; i < sequenceLength; i++ {
		// Create feature vector for this timestep
		features := tfb.generateTimestepFeatures(business, i, sequenceLength)
		sequence[i] = features
	}

	return sequence, nil
}

// generateTimestepFeatures generates features for a specific timestep
func (tfb *TemporalFeatureBuilder) generateTimestepFeatures(business *models.RiskAssessmentRequest, timestep, totalSteps int) []float64 {
	features := make([]float64, 20) // 20 features

	// Normalize timestep to 0-1 range
	timeProgress := float64(timestep) / float64(totalSteps-1)

	// Feature 0: Business name length (normalized)
	features[0] = float64(len(business.BusinessName)) / 100.0

	// Feature 1: Industry risk factor
	features[1] = tfb.getIndustryRiskFactor(business.Industry)

	// Feature 2: Address completeness
	features[2] = tfb.getAddressCompleteness(business.BusinessAddress)

	// Feature 3: Time-based trend (simulate business growth/decline)
	features[3] = tfb.generateTrendFeature(timeProgress)

	// Feature 4: Seasonal pattern (quarterly cycles)
	features[4] = tfb.generateSeasonalFeature(timestep)

	// Feature 5: Random walk component
	features[5] = tfb.generateRandomWalkFeature(timestep, business.BusinessName)

	// Feature 6: Economic cycle simulation
	features[6] = tfb.generateEconomicCycleFeature(timeProgress)

	// Feature 7: Compliance risk (simulate compliance events)
	features[7] = tfb.generateComplianceRiskFeature(timestep, business.Industry)

	// Feature 8: Market volatility
	features[8] = tfb.generateMarketVolatilityFeature(timeProgress)

	// Feature 9: Business age simulation (older businesses more stable)
	features[9] = tfb.generateBusinessAgeFeature(timeProgress)

	// Features 10-19: Additional synthetic features
	for i := 10; i < 20; i++ {
		features[i] = tfb.generateAdditionalFeature(i, timestep, business)
	}

	return features
}

// getIndustryRiskFactor returns industry-specific risk factor
func (tfb *TemporalFeatureBuilder) getIndustryRiskFactor(industry string) float64 {
	industryRisks := map[string]float64{
		"technology":    0.2,
		"finance":       0.4,
		"healthcare":    0.3,
		"retail":        0.5,
		"manufacturing": 0.4,
		"construction":  0.6,
		"restaurant":    0.7,
		"consulting":    0.3,
		"education":     0.2,
		"default":       0.5,
	}

	if risk, exists := industryRisks[industry]; exists {
		return risk
	}
	return industryRisks["default"]
}

// getAddressCompleteness calculates address completeness score
func (tfb *TemporalFeatureBuilder) getAddressCompleteness(address string) float64 {
	if address == "" {
		return 0.0
	}

	// Simple completeness scoring based on address length and components
	score := 0.0

	// Check for street number
	if len(address) > 0 && address[0] >= '0' && address[0] <= '9' {
		score += 0.3
	}

	// Check for street name
	if len(address) > 5 {
		score += 0.3
	}

	// Check for city/state/zip
	if len(address) > 15 {
		score += 0.4
	}

	return math.Min(score, 1.0)
}

// generateTrendFeature generates a trend component
func (tfb *TemporalFeatureBuilder) generateTrendFeature(timeProgress float64) float64 {
	// Simulate different business trajectories
	// Some businesses grow, some decline, some stay stable

	// Use a sine wave with different phases for different trends
	trend := math.Sin(timeProgress * math.Pi * 2)

	// Add some randomness
	trend += (rand.Float64() - 0.5) * 0.3

	return (trend + 1) / 2 // Normalize to 0-1
}

// generateSeasonalFeature generates seasonal patterns
func (tfb *TemporalFeatureBuilder) generateSeasonalFeature(timestep int) float64 {
	// Quarterly seasonal pattern
	quarter := timestep % 4
	seasonalValues := []float64{0.8, 1.0, 0.9, 0.7} // Q1, Q2, Q3, Q4

	baseValue := seasonalValues[quarter]

	// Add some noise
	noise := (rand.Float64() - 0.5) * 0.2

	return math.Max(0.0, math.Min(1.0, baseValue+noise))
}

// generateRandomWalkFeature generates random walk component
func (tfb *TemporalFeatureBuilder) generateRandomWalkFeature(timestep int, businessName string) float64 {
	// Use business name as seed for consistent random walk
	seed := int64(0)
	for _, char := range businessName {
		seed += int64(char)
	}

	r := rand.New(rand.NewSource(seed + int64(timestep)))

	// Generate random walk
	walk := 0.0
	for i := 0; i <= timestep; i++ {
		walk += (r.Float64() - 0.5) * 0.1
	}

	// Normalize to 0-1 range
	return math.Max(0.0, math.Min(1.0, (walk+1)/2))
}

// generateEconomicCycleFeature generates economic cycle simulation
func (tfb *TemporalFeatureBuilder) generateEconomicCycleFeature(timeProgress float64) float64 {
	// Simulate economic cycles (boom/bust)
	cycle := math.Sin(timeProgress * math.Pi * 4) // 4 cycles over the sequence

	// Add some phase shift
	cycle += math.Sin(timeProgress*math.Pi*8) * 0.3

	return (cycle + 1) / 2 // Normalize to 0-1
}

// generateComplianceRiskFeature generates compliance risk over time
func (tfb *TemporalFeatureBuilder) generateComplianceRiskFeature(timestep int, industry string) float64 {
	// Some industries have higher compliance risk
	baseRisk := tfb.getIndustryRiskFactor(industry)

	// Simulate compliance events (spikes in risk)
	eventProbability := 0.1 // 10% chance of compliance event per timestep

	if rand.Float64() < eventProbability {
		// Compliance event occurred
		return math.Min(1.0, baseRisk+0.3)
	}

	// Normal compliance risk
	return baseRisk + (rand.Float64()-0.5)*0.1
}

// generateMarketVolatilityFeature generates market volatility
func (tfb *TemporalFeatureBuilder) generateMarketVolatilityFeature(timeProgress float64) float64 {
	// Market volatility tends to cluster (volatile periods followed by calm periods)
	volatility := math.Sin(timeProgress*math.Pi*6)*0.5 + 0.5

	// Add some randomness
	volatility += (rand.Float64() - 0.5) * 0.2

	return math.Max(0.0, math.Min(1.0, volatility))
}

// generateBusinessAgeFeature generates business age simulation
func (tfb *TemporalFeatureBuilder) generateBusinessAgeFeature(timeProgress float64) float64 {
	// Simulate business maturity (older businesses tend to be more stable)
	// This is a reverse time progression (0 = oldest, 1 = newest)
	ageProgress := 1.0 - timeProgress

	// Mature businesses (low ageProgress) have lower risk
	maturityFactor := 1.0 - (ageProgress * 0.3)

	return math.Max(0.0, math.Min(1.0, maturityFactor))
}

// generateAdditionalFeature generates additional synthetic features
func (tfb *TemporalFeatureBuilder) generateAdditionalFeature(featureIndex, timestep int, business *models.RiskAssessmentRequest) float64 {
	// Generate various synthetic features based on feature index
	switch featureIndex {
	case 10:
		// Customer satisfaction simulation
		return 0.7 + (rand.Float64()-0.5)*0.4
	case 11:
		// Employee count simulation
		return math.Min(1.0, float64(len(business.BusinessName))/50.0)
	case 12:
		// Revenue growth simulation
		return 0.5 + math.Sin(float64(timestep)*0.5)*0.3
	case 13:
		// Market share simulation
		return 0.3 + (rand.Float64()-0.5)*0.4
	case 14:
		// Innovation index simulation
		return 0.4 + math.Cos(float64(timestep)*0.3)*0.3
	case 15:
		// Regulatory environment simulation
		return 0.6 + (rand.Float64()-0.5)*0.2
	case 16:
		// Competition intensity simulation
		return 0.5 + math.Sin(float64(timestep)*0.7)*0.3
	case 17:
		// Supply chain risk simulation
		return 0.4 + (rand.Float64()-0.5)*0.3
	case 18:
		// Technology adoption simulation
		return 0.6 + math.Cos(float64(timestep)*0.4)*0.2
	case 19:
		// Environmental risk simulation
		return 0.3 + (rand.Float64()-0.5)*0.4
	default:
		return rand.Float64()
	}
}

// SetLogger sets the logger for the temporal feature builder
func (tfb *TemporalFeatureBuilder) SetLogger(logger *zap.Logger) {
	tfb.logger = logger
}

// SetSequenceLength sets the sequence length for the temporal feature builder
func (tfb *TemporalFeatureBuilder) SetSequenceLength(length int) {
	tfb.sequenceLength = length
}
