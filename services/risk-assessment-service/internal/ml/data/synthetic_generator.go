package data

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// SyntheticDataGenerator generates synthetic time-series data for LSTM training and inference
type SyntheticDataGenerator struct {
	industryPatterns map[string]RiskPattern
	randomSeed       int64
	logger           *zap.Logger
}

// RiskPattern defines the risk characteristics for an industry
type RiskPattern struct {
	BaseRisk            float64
	Volatility          float64
	Seasonality         []float64 // 12 months of seasonal factors
	Trend               float64   // Annual trend
	EconomicSensitivity float64   // Sensitivity to economic cycles
}

// RiskDataPoint represents a single data point in the time series
type RiskDataPoint struct {
	Timestamp        time.Time
	RiskScore        float64
	FinancialHealth  float64
	ComplianceScore  float64
	MarketConditions float64
	RevenueTrend     float64
	EmployeeGrowth   float64
	RiskVolatility   float64
}

// NewSyntheticDataGenerator creates a new synthetic data generator
func NewSyntheticDataGenerator() *SyntheticDataGenerator {
	return &SyntheticDataGenerator{
		industryPatterns: initializeIndustryPatterns(),
		randomSeed:       time.Now().UnixNano(),
		logger:           zap.NewNop(),
	}
}

// SetLogger sets the logger for the synthetic data generator
func (sdg *SyntheticDataGenerator) SetLogger(logger *zap.Logger) {
	sdg.logger = logger
}

// GenerateHistoricalSequence generates a historical sequence of risk data for a business
func (sdg *SyntheticDataGenerator) GenerateHistoricalSequence(business *models.RiskAssessmentRequest, months int) []RiskDataPoint {
	// Get industry pattern
	pattern, exists := sdg.industryPatterns[business.Industry]
	if !exists {
		pattern = sdg.industryPatterns["default"]
	}

	// Set random seed for reproducible results
	rng := rand.New(rand.NewSource(sdg.randomSeed + int64(len(business.BusinessName))))

	// Generate sequence
	sequence := make([]RiskDataPoint, months)
	baseTime := time.Now().AddDate(0, -months, 0)

	for i := 0; i < months; i++ {
		currentTime := baseTime.AddDate(0, i, 0)
		monthIndex := int(currentTime.Month()) - 1

		// Generate base risk score with trend and seasonality
		trendFactor := float64(i) * pattern.Trend / 12.0
		seasonalFactor := pattern.Seasonality[monthIndex]
		baseRisk := pattern.BaseRisk + trendFactor + seasonalFactor

		// Add random walk component
		randomWalk := (rng.Float64() - 0.5) * pattern.Volatility
		riskScore := baseRisk + randomWalk

		// Ensure risk score is in valid range
		riskScore = math.Max(0.0, math.Min(1.0, riskScore))

		// Generate correlated features
		financialHealth := sdg.generateCorrelatedFeature(riskScore, 0.7, rng)
		complianceScore := sdg.generateCorrelatedFeature(riskScore, 0.6, rng)
		marketConditions := sdg.generateMarketConditions(currentTime, rng)
		revenueTrend := sdg.generateCorrelatedFeature(riskScore, 0.5, rng)
		employeeGrowth := sdg.generateCorrelatedFeature(riskScore, 0.4, rng)
		riskVolatility := sdg.generateVolatility(riskScore, pattern.Volatility, rng)

		sequence[i] = RiskDataPoint{
			Timestamp:        currentTime,
			RiskScore:        riskScore,
			FinancialHealth:  financialHealth,
			ComplianceScore:  complianceScore,
			MarketConditions: marketConditions,
			RevenueTrend:     revenueTrend,
			EmployeeGrowth:   employeeGrowth,
			RiskVolatility:   riskVolatility,
		}
	}

	sdg.logger.Debug("Generated synthetic historical sequence",
		zap.String("business_name", business.BusinessName),
		zap.String("industry", business.Industry),
		zap.Int("months", months),
		zap.Float64("avg_risk_score", sdg.calculateAverageRisk(sequence)))

	return sequence
}

// generateCorrelatedFeature generates a feature that is correlated with the risk score
func (sdg *SyntheticDataGenerator) generateCorrelatedFeature(riskScore, correlation float64, rng *rand.Rand) float64 {
	// Generate a feature that is inversely correlated with risk (lower risk = higher feature value)
	baseValue := 1.0 - riskScore*correlation
	noise := (rng.Float64() - 0.5) * 0.2
	return math.Max(0.0, math.Min(1.0, baseValue+noise))
}

// generateMarketConditions generates market conditions based on time and economic cycles
func (sdg *SyntheticDataGenerator) generateMarketConditions(timestamp time.Time, rng *rand.Rand) float64 {
	// Simulate economic cycles (boom/bust cycles)
	yearsSinceEpoch := timestamp.Year() - 2020
	cyclePhase := math.Sin(float64(yearsSinceEpoch) * 2 * math.Pi / 7) // 7-year cycle

	// Add some randomness
	noise := (rng.Float64() - 0.5) * 0.3

	// Convert to 0-1 range
	marketConditions := (cyclePhase+1)/2 + noise
	return math.Max(0.0, math.Min(1.0, marketConditions))
}

// generateVolatility generates risk volatility based on current risk and industry volatility
func (sdg *SyntheticDataGenerator) generateVolatility(riskScore, industryVolatility float64, rng *rand.Rand) float64 {
	// Higher risk scores tend to have higher volatility
	baseVolatility := industryVolatility * (0.5 + riskScore*0.5)
	noise := (rng.Float64() - 0.5) * 0.1
	return math.Max(0.0, math.Min(1.0, baseVolatility+noise))
}

// calculateAverageRisk calculates the average risk score in a sequence
func (sdg *SyntheticDataGenerator) calculateAverageRisk(sequence []RiskDataPoint) float64 {
	if len(sequence) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, point := range sequence {
		sum += point.RiskScore
	}
	return sum / float64(len(sequence))
}

// initializeIndustryPatterns initializes risk patterns for different industries
func initializeIndustryPatterns() map[string]RiskPattern {
	patterns := make(map[string]RiskPattern)

	// Technology industry
	patterns["technology"] = RiskPattern{
		BaseRisk:            0.25,
		Volatility:          0.15,
		Seasonality:         []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		Trend:               0.02,
		EconomicSensitivity: 0.3,
	}

	// Healthcare industry
	patterns["healthcare"] = RiskPattern{
		BaseRisk:            0.20,
		Volatility:          0.10,
		Seasonality:         []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		Trend:               -0.01,
		EconomicSensitivity: 0.1,
	}

	// Financial services
	patterns["financial"] = RiskPattern{
		BaseRisk:            0.35,
		Volatility:          0.25,
		Seasonality:         []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		Trend:               0.01,
		EconomicSensitivity: 0.8,
	}

	// Manufacturing
	patterns["manufacturing"] = RiskPattern{
		BaseRisk:            0.30,
		Volatility:          0.20,
		Seasonality:         []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		Trend:               0.0,
		EconomicSensitivity: 0.6,
	}

	// Retail
	patterns["retail"] = RiskPattern{
		BaseRisk:            0.40,
		Volatility:          0.30,
		Seasonality:         []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		Trend:               0.0,
		EconomicSensitivity: 0.7,
	}

	// Default pattern
	patterns["default"] = RiskPattern{
		BaseRisk:            0.30,
		Volatility:          0.20,
		Seasonality:         []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		Trend:               0.0,
		EconomicSensitivity: 0.5,
	}

	return patterns
}

// GenerateTrainingDataset generates a large dataset for model training
func (sdg *SyntheticDataGenerator) GenerateTrainingDataset(businessCount int, monthsPerBusiness int) map[string][]RiskDataPoint {
	dataset := make(map[string][]RiskDataPoint)

	// Generate data for multiple businesses
	for i := 0; i < businessCount; i++ {
		businessName := fmt.Sprintf("Training_Business_%d", i)
		industry := sdg.getRandomIndustry()

		business := &models.RiskAssessmentRequest{
			BusinessName:    businessName,
			BusinessAddress: fmt.Sprintf("%d Training St", i),
			Industry:        industry,
			Country:         "US",
		}

		sequence := sdg.GenerateHistoricalSequence(business, monthsPerBusiness)
		dataset[businessName] = sequence
	}

	sdg.logger.Info("Generated training dataset",
		zap.Int("business_count", businessCount),
		zap.Int("months_per_business", monthsPerBusiness),
		zap.Int("total_data_points", businessCount*monthsPerBusiness))

	return dataset
}

// getRandomIndustry returns a random industry for training data generation
func (sdg *SyntheticDataGenerator) getRandomIndustry() string {
	industries := []string{"technology", "healthcare", "financial", "manufacturing", "retail"}
	rng := rand.New(rand.NewSource(sdg.randomSeed + time.Now().UnixNano()))
	return industries[rng.Intn(len(industries))]
}

// AddComplianceEvent adds a compliance event to the sequence
func (sdg *SyntheticDataGenerator) AddComplianceEvent(sequence []RiskDataPoint, eventMonth int, severity float64) []RiskDataPoint {
	if eventMonth < 0 || eventMonth >= len(sequence) {
		return sequence
	}

	// Adjust risk score and compliance score for the event month
	sequence[eventMonth].RiskScore = math.Min(1.0, sequence[eventMonth].RiskScore+severity*0.3)
	sequence[eventMonth].ComplianceScore = math.Max(0.0, sequence[eventMonth].ComplianceScore-severity*0.4)

	// Add some spillover effect to adjacent months
	if eventMonth > 0 {
		sequence[eventMonth-1].RiskScore = math.Min(1.0, sequence[eventMonth-1].RiskScore+severity*0.1)
	}
	if eventMonth < len(sequence)-1 {
		sequence[eventMonth+1].RiskScore = math.Min(1.0, sequence[eventMonth+1].RiskScore+severity*0.1)
	}

	return sequence
}

// AddEconomicShock adds an economic shock to the sequence
func (sdg *SyntheticDataGenerator) AddEconomicShock(sequence []RiskDataPoint, shockMonth int, severity float64, duration int) []RiskDataPoint {
	if shockMonth < 0 || shockMonth >= len(sequence) {
		return sequence
	}

	// Apply shock to multiple months
	for i := 0; i < duration && shockMonth+i < len(sequence); i++ {
		monthIndex := shockMonth + i
		decayFactor := 1.0 - float64(i)/float64(duration) // Decay over time

		// Adjust market conditions and risk score
		sequence[monthIndex].MarketConditions = math.Max(0.0, sequence[monthIndex].MarketConditions-severity*decayFactor*0.5)
		sequence[monthIndex].RiskScore = math.Min(1.0, sequence[monthIndex].RiskScore+severity*decayFactor*0.2)
	}

	return sequence
}

// GetIndustryPattern returns the risk pattern for a specific industry
func (sdg *SyntheticDataGenerator) GetIndustryPattern(industry string) (RiskPattern, bool) {
	pattern, exists := sdg.industryPatterns[industry]
	return pattern, exists
}

// SetIndustryPattern sets a custom risk pattern for an industry
func (sdg *SyntheticDataGenerator) SetIndustryPattern(industry string, pattern RiskPattern) {
	sdg.industryPatterns[industry] = pattern
	sdg.logger.Info("Updated industry pattern",
		zap.String("industry", industry),
		zap.Float64("base_risk", pattern.BaseRisk),
		zap.Float64("volatility", pattern.Volatility))
}
