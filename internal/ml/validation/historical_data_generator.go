package validation

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// HistoricalDataGenerator generates realistic historical risk assessment data
type HistoricalDataGenerator struct {
	logger *zap.Logger
}

// DataGenerationConfig configures the historical data generation
type DataGenerationConfig struct {
	TotalSamples     int                `json:"total_samples"`
	TimeRange        time.Duration      `json:"time_range"`
	RiskCategories   []RiskCategory     `json:"risk_categories"`
	IndustryWeights  map[string]float64 `json:"industry_weights"`
	GeographicBias   map[string]float64 `json:"geographic_bias"`
	SeasonalPatterns bool               `json:"seasonal_patterns"`
	TrendStrength    float64            `json:"trend_strength"`
	NoiseLevel       float64            `json:"noise_level"`
}

// RiskCategory defines a risk category with its characteristics
type RiskCategory struct {
	Name           string             `json:"name"`
	BaseRisk       float64            `json:"base_risk"`
	Volatility     float64            `json:"volatility"`
	IndustryBias   map[string]float64 `json:"industry_bias"`
	GeographicBias map[string]float64 `json:"geographic_bias"`
	SizeBias       map[string]float64 `json:"size_bias"`
	AgeBias        map[string]float64 `json:"age_bias"`
}

// HistoricalSample represents a historical risk assessment sample
type HistoricalSample struct {
	RiskSample
	Timestamp      time.Time          `json:"timestamp"`
	BusinessID     string             `json:"business_id"`
	Industry       string             `json:"industry"`
	Country        string             `json:"country"`
	BusinessSize   string             `json:"business_size"`
	BusinessAge    int                `json:"business_age"`
	ActualOutcome  float64            `json:"actual_outcome"`
	PredictedRisk  float64            `json:"predicted_risk"`
	RiskFactors    map[string]float64 `json:"risk_factors"`
	ExternalEvents []ExternalEvent    `json:"external_events"`
}

// ExternalEvent represents external events that affect risk
type ExternalEvent struct {
	Type        string    `json:"type"`
	Impact      float64   `json:"impact"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}

// NewHistoricalDataGenerator creates a new historical data generator
func NewHistoricalDataGenerator(logger *zap.Logger) *HistoricalDataGenerator {
	return &HistoricalDataGenerator{
		logger: logger,
	}
}

// GenerateHistoricalData generates realistic historical risk assessment data
func (hdg *HistoricalDataGenerator) GenerateHistoricalData(
	ctx context.Context,
	config DataGenerationConfig,
) ([]RiskSample, []HistoricalSample, error) {
	hdg.logger.Info("Generating historical risk assessment data",
		zap.Int("total_samples", config.TotalSamples),
		zap.Duration("time_range", config.TimeRange),
		zap.Int("risk_categories", len(config.RiskCategories)))

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Generate base samples
	samples := make([]RiskSample, 0, config.TotalSamples)
	historicalSamples := make([]HistoricalSample, 0, config.TotalSamples)

	// Calculate time step
	timeStep := config.TimeRange / time.Duration(config.TotalSamples)
	startTime := time.Now().Add(-config.TimeRange)

	// Generate samples over time
	for i := 0; i < config.TotalSamples; i++ {
		sampleTime := startTime.Add(timeStep * time.Duration(i))

		// Generate business characteristics
		businessID := fmt.Sprintf("BUS_%d_%d", i, sampleTime.Unix())
		industry := hdg.selectIndustry(config.IndustryWeights)
		country := hdg.selectCountry(config.GeographicBias)
		businessSize := hdg.selectBusinessSize()
		businessAge := hdg.generateBusinessAge(sampleTime)

		// Generate risk factors
		riskFactors := hdg.generateRiskFactors(industry, country, businessSize, businessAge, sampleTime, config)

		// Calculate base risk score
		baseRisk := hdg.calculateBaseRisk(riskFactors, config.RiskCategories)

		// Apply temporal effects
		temporalRisk := hdg.applyTemporalEffects(baseRisk, sampleTime, config)

		// Apply external events
		externalEvents := hdg.generateExternalEvents(sampleTime, config)
		finalRisk := hdg.applyExternalEvents(temporalRisk, externalEvents)

		// Add noise
		noisyRisk := hdg.addNoise(finalRisk, config.NoiseLevel)

		// Generate features for ML model
		features := hdg.generateFeatures(riskFactors, sampleTime, config)

		// Create samples
		riskSample := RiskSample{
			Features: features,
			Label:    noisyRisk,
			Metadata: map[string]interface{}{
				"business_id":   businessID,
				"industry":      industry,
				"country":       country,
				"business_size": businessSize,
				"business_age":  businessAge,
				"timestamp":     sampleTime,
			},
		}

		historicalSample := HistoricalSample{
			RiskSample:     riskSample,
			Timestamp:      sampleTime,
			BusinessID:     businessID,
			Industry:       industry,
			Country:        country,
			BusinessSize:   businessSize,
			BusinessAge:    businessAge,
			ActualOutcome:  noisyRisk,
			PredictedRisk:  baseRisk,
			RiskFactors:    riskFactors,
			ExternalEvents: externalEvents,
		}

		samples = append(samples, riskSample)
		historicalSamples = append(historicalSamples, historicalSample)
	}

	hdg.logger.Info("Historical data generation completed",
		zap.Int("generated_samples", len(samples)),
		zap.Int("historical_samples", len(historicalSamples)))

	return samples, historicalSamples, nil
}

// selectIndustry selects an industry based on weights
func (hdg *HistoricalDataGenerator) selectIndustry(weights map[string]float64) string {
	industries := []string{
		"Technology", "Finance", "Healthcare", "Manufacturing", "Retail",
		"Real Estate", "Construction", "Transportation", "Energy", "Education",
		"Entertainment", "Food & Beverage", "Agriculture", "Mining", "Utilities",
	}

	if len(weights) == 0 {
		return industries[rand.Intn(len(industries))]
	}

	// Weighted selection
	totalWeight := 0.0
	for _, weight := range weights {
		totalWeight += weight
	}

	random := rand.Float64() * totalWeight
	current := 0.0

	for industry, weight := range weights {
		current += weight
		if random <= current {
			return industry
		}
	}

	return industries[rand.Intn(len(industries))]
}

// selectCountry selects a country based on geographic bias
func (hdg *HistoricalDataGenerator) selectCountry(bias map[string]float64) string {
	countries := []string{
		"United States", "Canada", "United Kingdom", "Germany", "France",
		"Japan", "Australia", "Brazil", "India", "China", "Mexico", "Italy",
		"Spain", "Netherlands", "Sweden", "Norway", "Switzerland", "Singapore",
	}

	if len(bias) == 0 {
		return countries[rand.Intn(len(countries))]
	}

	// Weighted selection
	totalWeight := 0.0
	for _, weight := range bias {
		totalWeight += weight
	}

	random := rand.Float64() * totalWeight
	current := 0.0

	for country, weight := range bias {
		current += weight
		if random <= current {
			return country
		}
	}

	return countries[rand.Intn(len(countries))]
}

// selectBusinessSize selects a business size
func (hdg *HistoricalDataGenerator) selectBusinessSize() string {
	sizes := []string{"Micro", "Small", "Medium", "Large", "Enterprise"}
	weights := []float64{0.3, 0.4, 0.2, 0.08, 0.02} // Realistic distribution

	random := rand.Float64()
	current := 0.0

	for i, weight := range weights {
		current += weight
		if random <= current {
			return sizes[i]
		}
	}

	return sizes[0]
}

// generateBusinessAge generates a realistic business age
func (hdg *HistoricalDataGenerator) generateBusinessAge(sampleTime time.Time) int {
	// Most businesses are relatively young
	ages := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 15, 20, 30, 50}
	weights := []float64{0.15, 0.12, 0.10, 0.08, 0.07, 0.06, 0.05, 0.04, 0.03, 0.03, 0.08, 0.06, 0.04, 0.02}

	random := rand.Float64()
	current := 0.0

	for i, weight := range weights {
		current += weight
		if random <= current {
			return ages[i]
		}
	}

	return ages[0]
}

// generateRiskFactors generates risk factors for a business
func (hdg *HistoricalDataGenerator) generateRiskFactors(
	industry, country, businessSize string,
	businessAge int,
	sampleTime time.Time,
	config DataGenerationConfig,
) map[string]float64 {
	factors := make(map[string]float64)

	// Financial health factors
	factors["revenue_growth"] = hdg.generateRevenueGrowth(businessAge, industry)
	factors["profit_margin"] = hdg.generateProfitMargin(industry, businessSize)
	factors["debt_to_equity"] = hdg.generateDebtToEquity(businessSize, industry)
	factors["cash_flow"] = hdg.generateCashFlow(industry, businessSize)
	factors["credit_score"] = hdg.generateCreditScore(businessSize, businessAge)

	// Operational factors
	factors["employee_count"] = hdg.generateEmployeeCount(businessSize)
	factors["market_share"] = hdg.generateMarketShare(industry, businessSize)
	factors["customer_concentration"] = hdg.generateCustomerConcentration(businessSize)
	factors["supplier_dependency"] = hdg.generateSupplierDependency(industry)
	factors["regulatory_compliance"] = hdg.generateRegulatoryCompliance(industry, country)

	// Market factors
	factors["market_volatility"] = hdg.generateMarketVolatility(industry, sampleTime)
	factors["competition_level"] = hdg.generateCompetitionLevel(industry)
	factors["barriers_to_entry"] = hdg.generateBarriersToEntry(industry)
	factors["economic_cycle"] = hdg.generateEconomicCycle(sampleTime, config)

	// Technology factors
	factors["digital_maturity"] = hdg.generateDigitalMaturity(industry, businessAge)
	factors["cyber_security"] = hdg.generateCyberSecurity(industry, businessSize)
	factors["innovation_index"] = hdg.generateInnovationIndex(industry, businessSize)

	// Environmental factors
	factors["climate_risk"] = hdg.generateClimateRisk(industry, country)
	factors["sustainability_score"] = hdg.generateSustainabilityScore(industry, businessSize)
	factors["esg_rating"] = hdg.generateESGRating(industry, country)

	return factors
}

// generateFeatures converts risk factors to ML features
func (hdg *HistoricalDataGenerator) generateFeatures(
	riskFactors map[string]float64,
	sampleTime time.Time,
	config DataGenerationConfig,
) []float64 {
	features := make([]float64, 0, len(riskFactors)+10)

	// Add risk factors as features
	for _, value := range riskFactors {
		features = append(features, value)
	}

	// Add temporal features
	features = append(features, float64(sampleTime.Year()))
	features = append(features, float64(sampleTime.Month()))
	features = append(features, float64(sampleTime.Day()))
	features = append(features, float64(sampleTime.Weekday()))
	features = append(features, float64(sampleTime.Hour()))

	// Add seasonal features
	if config.SeasonalPatterns {
		features = append(features, hdg.calculateSeasonalFactor(sampleTime))
		features = append(features, hdg.calculateQuarterlyFactor(sampleTime))
	}

	// Add trend features
	features = append(features, hdg.calculateTrendFactor(sampleTime, config))

	// Add noise to features
	for i := range features {
		features[i] += rand.NormFloat64() * config.NoiseLevel * 0.1
	}

	return features
}

// calculateBaseRisk calculates base risk score from risk factors
func (hdg *HistoricalDataGenerator) calculateBaseRisk(
	riskFactors map[string]float64,
	categories []RiskCategory,
) float64 {
	if len(categories) == 0 {
		// Default risk calculation
		return hdg.calculateDefaultRisk(riskFactors)
	}

	// Calculate risk for each category
	totalRisk := 0.0
	totalWeight := 0.0

	for _, category := range categories {
		categoryRisk := hdg.calculateCategoryRisk(riskFactors, category)
		weight := 1.0 / float64(len(categories)) // Equal weight for now

		totalRisk += categoryRisk * weight
		totalWeight += weight
	}

	if totalWeight > 0 {
		return totalRisk / totalWeight
	}

	return hdg.calculateDefaultRisk(riskFactors)
}

// calculateDefaultRisk calculates default risk score
func (hdg *HistoricalDataGenerator) calculateDefaultRisk(riskFactors map[string]float64) float64 {
	// Weighted combination of key risk factors
	weights := map[string]float64{
		"revenue_growth":         0.15,
		"profit_margin":          0.12,
		"debt_to_equity":         0.10,
		"cash_flow":              0.10,
		"credit_score":           0.08,
		"market_volatility":      0.08,
		"competition_level":      0.07,
		"regulatory_compliance":  0.06,
		"customer_concentration": 0.05,
		"supplier_dependency":    0.05,
		"cyber_security":         0.04,
		"esg_rating":             0.10,
	}

	totalRisk := 0.0
	totalWeight := 0.0

	for factor, weight := range weights {
		if value, exists := riskFactors[factor]; exists {
			// Normalize and weight the factor
			normalizedValue := hdg.normalizeRiskFactor(factor, value)
			totalRisk += normalizedValue * weight
			totalWeight += weight
		}
	}

	if totalWeight > 0 {
		return math.Max(0, math.Min(1, totalRisk/totalWeight))
	}

	return 0.5 // Default neutral risk
}

// normalizeRiskFactor normalizes a risk factor to 0-1 scale
func (hdg *HistoricalDataGenerator) normalizeRiskFactor(factor string, value float64) float64 {
	// Factor-specific normalization
	switch factor {
	case "revenue_growth":
		return math.Max(0, math.Min(1, (value+0.5)/1.0)) // -50% to +50%
	case "profit_margin":
		return math.Max(0, math.Min(1, (value+0.2)/0.4)) // -20% to +20%
	case "debt_to_equity":
		return math.Max(0, math.Min(1, value/2.0)) // 0 to 2
	case "credit_score":
		return value / 850.0 // 0 to 850
	case "market_volatility":
		return math.Max(0, math.Min(1, value/0.5)) // 0 to 50%
	default:
		return math.Max(0, math.Min(1, value))
	}
}

// Helper methods for generating specific risk factors
func (hdg *HistoricalDataGenerator) generateRevenueGrowth(businessAge int, industry string) float64 {
	// Younger businesses tend to have higher growth rates
	baseGrowth := 0.1 - float64(businessAge)*0.01
	industryMultiplier := hdg.getIndustryMultiplier(industry, "growth")
	return baseGrowth*industryMultiplier + rand.NormFloat64()*0.1
}

func (hdg *HistoricalDataGenerator) generateProfitMargin(industry, businessSize string) float64 {
	baseMargin := 0.15
	industryMultiplier := hdg.getIndustryMultiplier(industry, "margin")
	sizeMultiplier := hdg.getSizeMultiplier(businessSize, "margin")
	return baseMargin*industryMultiplier*sizeMultiplier + rand.NormFloat64()*0.05
}

func (hdg *HistoricalDataGenerator) generateDebtToEquity(businessSize, industry string) float64 {
	baseRatio := 0.5
	sizeMultiplier := hdg.getSizeMultiplier(businessSize, "debt")
	industryMultiplier := hdg.getIndustryMultiplier(industry, "debt")
	return baseRatio*sizeMultiplier*industryMultiplier + rand.NormFloat64()*0.2
}

func (hdg *HistoricalDataGenerator) generateCashFlow(industry, businessSize string) float64 {
	baseFlow := 0.1
	industryMultiplier := hdg.getIndustryMultiplier(industry, "cashflow")
	sizeMultiplier := hdg.getSizeMultiplier(businessSize, "cashflow")
	return baseFlow*industryMultiplier*sizeMultiplier + rand.NormFloat64()*0.05
}

func (hdg *HistoricalDataGenerator) generateCreditScore(businessSize string, businessAge int) float64 {
	baseScore := 650.0
	ageBonus := float64(businessAge) * 2.0
	sizeBonus := hdg.getSizeMultiplier(businessSize, "credit") * 50.0
	return math.Max(300, math.Min(850, baseScore+ageBonus+sizeBonus+rand.NormFloat64()*50))
}

func (hdg *HistoricalDataGenerator) generateEmployeeCount(businessSize string) float64 {
	switch businessSize {
	case "Micro":
		return float64(rand.Intn(10) + 1)
	case "Small":
		return float64(rand.Intn(50) + 10)
	case "Medium":
		return float64(rand.Intn(200) + 50)
	case "Large":
		return float64(rand.Intn(1000) + 200)
	case "Enterprise":
		return float64(rand.Intn(5000) + 1000)
	default:
		return float64(rand.Intn(100) + 1)
	}
}

func (hdg *HistoricalDataGenerator) generateMarketShare(industry, businessSize string) float64 {
	baseShare := 0.01
	sizeMultiplier := hdg.getSizeMultiplier(businessSize, "marketshare")
	industryMultiplier := hdg.getIndustryMultiplier(industry, "marketshare")
	return baseShare*sizeMultiplier*industryMultiplier + rand.NormFloat64()*0.005
}

func (hdg *HistoricalDataGenerator) generateCustomerConcentration(businessSize string) float64 {
	// Smaller businesses tend to have higher customer concentration
	baseConcentration := 0.3
	sizeMultiplier := hdg.getSizeMultiplier(businessSize, "concentration")
	return baseConcentration*sizeMultiplier + rand.NormFloat64()*0.1
}

func (hdg *HistoricalDataGenerator) generateSupplierDependency(industry string) float64 {
	baseDependency := 0.2
	industryMultiplier := hdg.getIndustryMultiplier(industry, "supplier")
	return baseDependency*industryMultiplier + rand.NormFloat64()*0.1
}

func (hdg *HistoricalDataGenerator) generateRegulatoryCompliance(industry, country string) float64 {
	baseCompliance := 0.8
	industryMultiplier := hdg.getIndustryMultiplier(industry, "compliance")
	countryMultiplier := hdg.getCountryMultiplier(country, "compliance")
	return baseCompliance*industryMultiplier*countryMultiplier + rand.NormFloat64()*0.1
}

func (hdg *HistoricalDataGenerator) generateMarketVolatility(industry string, sampleTime time.Time) float64 {
	baseVolatility := 0.2
	industryMultiplier := hdg.getIndustryMultiplier(industry, "volatility")
	seasonalFactor := hdg.calculateSeasonalFactor(sampleTime)
	return baseVolatility*industryMultiplier*(1+seasonalFactor*0.2) + rand.NormFloat64()*0.05
}

func (hdg *HistoricalDataGenerator) generateCompetitionLevel(industry string) float64 {
	baseCompetition := 0.5
	industryMultiplier := hdg.getIndustryMultiplier(industry, "competition")
	return baseCompetition*industryMultiplier + rand.NormFloat64()*0.1
}

func (hdg *HistoricalDataGenerator) generateBarriersToEntry(industry string) float64 {
	baseBarriers := 0.3
	industryMultiplier := hdg.getIndustryMultiplier(industry, "barriers")
	return baseBarriers*industryMultiplier + rand.NormFloat64()*0.1
}

func (hdg *HistoricalDataGenerator) generateEconomicCycle(sampleTime time.Time, config DataGenerationConfig) float64 {
	// Simulate economic cycle
	cycleLength := 7.0 * 365 * 24 * 60 * 60 // 7 years in seconds
	cyclePosition := float64(sampleTime.Unix()%int64(cycleLength)) / cycleLength
	return math.Sin(cyclePosition*2*math.Pi)*0.3 + 0.5
}

func (hdg *HistoricalDataGenerator) generateDigitalMaturity(industry string, businessAge int) float64 {
	baseMaturity := 0.6
	industryMultiplier := hdg.getIndustryMultiplier(industry, "digital")
	ageFactor := math.Max(0, 1-float64(businessAge)*0.02) // Younger businesses more digital
	return baseMaturity*industryMultiplier*ageFactor + rand.NormFloat64()*0.1
}

func (hdg *HistoricalDataGenerator) generateCyberSecurity(industry, businessSize string) float64 {
	baseSecurity := 0.7
	industryMultiplier := hdg.getIndustryMultiplier(industry, "cybersecurity")
	sizeMultiplier := hdg.getSizeMultiplier(businessSize, "cybersecurity")
	return baseSecurity*industryMultiplier*sizeMultiplier + rand.NormFloat64()*0.1
}

func (hdg *HistoricalDataGenerator) generateInnovationIndex(industry, businessSize string) float64 {
	baseInnovation := 0.5
	industryMultiplier := hdg.getIndustryMultiplier(industry, "innovation")
	sizeMultiplier := hdg.getSizeMultiplier(businessSize, "innovation")
	return baseInnovation*industryMultiplier*sizeMultiplier + rand.NormFloat64()*0.1
}

func (hdg *HistoricalDataGenerator) generateClimateRisk(industry, country string) float64 {
	baseRisk := 0.3
	industryMultiplier := hdg.getIndustryMultiplier(industry, "climate")
	countryMultiplier := hdg.getCountryMultiplier(country, "climate")
	return baseRisk*industryMultiplier*countryMultiplier + rand.NormFloat64()*0.1
}

func (hdg *HistoricalDataGenerator) generateSustainabilityScore(industry, businessSize string) float64 {
	baseScore := 0.6
	industryMultiplier := hdg.getIndustryMultiplier(industry, "sustainability")
	sizeMultiplier := hdg.getSizeMultiplier(businessSize, "sustainability")
	return baseScore*industryMultiplier*sizeMultiplier + rand.NormFloat64()*0.1
}

func (hdg *HistoricalDataGenerator) generateESGRating(industry, country string) float64 {
	baseRating := 0.7
	industryMultiplier := hdg.getIndustryMultiplier(industry, "esg")
	countryMultiplier := hdg.getCountryMultiplier(country, "esg")
	return baseRating*industryMultiplier*countryMultiplier + rand.NormFloat64()*0.1
}

// Helper methods for multipliers
func (hdg *HistoricalDataGenerator) getIndustryMultiplier(industry, factor string) float64 {
	multipliers := map[string]map[string]float64{
		"Technology":    {"growth": 1.5, "margin": 1.2, "debt": 0.8, "cashflow": 1.1, "credit": 1.1, "marketshare": 0.8, "concentration": 0.9, "supplier": 0.7, "compliance": 0.9, "volatility": 1.3, "competition": 1.4, "barriers": 0.6, "digital": 1.4, "cybersecurity": 1.2, "innovation": 1.5, "climate": 0.8, "sustainability": 1.1, "esg": 1.0},
		"Finance":       {"growth": 1.0, "margin": 1.3, "debt": 1.2, "cashflow": 1.2, "credit": 1.3, "marketshare": 1.1, "concentration": 1.0, "supplier": 0.8, "compliance": 1.4, "volatility": 1.1, "competition": 1.2, "barriers": 1.3, "digital": 1.2, "cybersecurity": 1.4, "innovation": 1.1, "climate": 0.9, "sustainability": 1.0, "esg": 1.1},
		"Healthcare":    {"growth": 1.1, "margin": 1.1, "debt": 0.9, "cashflow": 1.0, "credit": 1.2, "marketshare": 1.0, "concentration": 0.8, "supplier": 1.1, "compliance": 1.3, "volatility": 0.8, "competition": 0.9, "barriers": 1.2, "digital": 1.0, "cybersecurity": 1.3, "innovation": 1.2, "climate": 0.9, "sustainability": 1.1, "esg": 1.2},
		"Manufacturing": {"growth": 0.9, "margin": 0.8, "debt": 1.1, "cashflow": 0.9, "credit": 0.9, "marketshare": 1.2, "concentration": 1.1, "supplier": 1.3, "compliance": 1.0, "volatility": 1.0, "competition": 1.1, "barriers": 1.0, "digital": 0.8, "cybersecurity": 0.9, "innovation": 0.9, "climate": 1.2, "sustainability": 0.9, "esg": 0.9},
		"Retail":        {"growth": 1.0, "margin": 0.7, "debt": 1.0, "cashflow": 0.8, "credit": 0.8, "marketshare": 1.0, "concentration": 1.2, "supplier": 1.2, "compliance": 0.9, "volatility": 1.2, "competition": 1.3, "barriers": 0.7, "digital": 1.1, "cybersecurity": 0.8, "innovation": 1.0, "climate": 0.8, "sustainability": 1.0, "esg": 0.9},
	}

	if industryData, exists := multipliers[industry]; exists {
		if multiplier, exists := industryData[factor]; exists {
			return multiplier
		}
	}

	return 1.0 // Default multiplier
}

func (hdg *HistoricalDataGenerator) getSizeMultiplier(businessSize, factor string) float64 {
	multipliers := map[string]map[string]float64{
		"Micro":      {"margin": 0.7, "debt": 1.2, "cashflow": 0.6, "credit": 0.7, "marketshare": 0.3, "concentration": 1.5, "cybersecurity": 0.6, "innovation": 1.2, "sustainability": 0.8},
		"Small":      {"margin": 0.8, "debt": 1.1, "cashflow": 0.7, "credit": 0.8, "marketshare": 0.5, "concentration": 1.3, "cybersecurity": 0.7, "innovation": 1.1, "sustainability": 0.9},
		"Medium":     {"margin": 1.0, "debt": 1.0, "cashflow": 1.0, "credit": 1.0, "marketshare": 1.0, "concentration": 1.0, "cybersecurity": 1.0, "innovation": 1.0, "sustainability": 1.0},
		"Large":      {"margin": 1.1, "debt": 0.9, "cashflow": 1.1, "credit": 1.1, "marketshare": 1.5, "concentration": 0.8, "cybersecurity": 1.2, "innovation": 0.9, "sustainability": 1.1},
		"Enterprise": {"margin": 1.2, "debt": 0.8, "cashflow": 1.2, "credit": 1.2, "marketshare": 2.0, "concentration": 0.6, "cybersecurity": 1.3, "innovation": 0.8, "sustainability": 1.2},
	}

	if sizeData, exists := multipliers[businessSize]; exists {
		if multiplier, exists := sizeData[factor]; exists {
			return multiplier
		}
	}

	return 1.0 // Default multiplier
}

func (hdg *HistoricalDataGenerator) getCountryMultiplier(country, factor string) float64 {
	multipliers := map[string]map[string]float64{
		"United States":  {"compliance": 1.0, "climate": 1.0, "esg": 1.0},
		"Canada":         {"compliance": 1.1, "climate": 0.9, "esg": 1.1},
		"United Kingdom": {"compliance": 1.1, "climate": 0.8, "esg": 1.2},
		"Germany":        {"compliance": 1.2, "climate": 0.8, "esg": 1.3},
		"France":         {"compliance": 1.1, "climate": 0.9, "esg": 1.2},
		"Japan":          {"compliance": 1.0, "climate": 1.1, "esg": 1.0},
		"Australia":      {"compliance": 1.0, "climate": 1.2, "esg": 1.0},
		"Brazil":         {"compliance": 0.8, "climate": 1.3, "esg": 0.8},
		"India":          {"compliance": 0.7, "climate": 1.4, "esg": 0.7},
		"China":          {"compliance": 0.8, "climate": 1.2, "esg": 0.8},
	}

	if countryData, exists := multipliers[country]; exists {
		if multiplier, exists := countryData[factor]; exists {
			return multiplier
		}
	}

	return 1.0 // Default multiplier
}

// Temporal effect methods
func (hdg *HistoricalDataGenerator) applyTemporalEffects(baseRisk float64, sampleTime time.Time, config DataGenerationConfig) float64 {
	risk := baseRisk

	// Apply seasonal effects
	if config.SeasonalPatterns {
		seasonalFactor := hdg.calculateSeasonalFactor(sampleTime)
		risk += seasonalFactor * 0.1
	}

	// Apply trend effects
	trendFactor := hdg.calculateTrendFactor(sampleTime, config)
	risk += trendFactor * config.TrendStrength

	return math.Max(0, math.Min(1, risk))
}

func (hdg *HistoricalDataGenerator) calculateSeasonalFactor(sampleTime time.Time) float64 {
	// Quarterly seasonal pattern
	quarter := (sampleTime.Month() - 1) / 3
	seasonalValues := []float64{0.1, -0.05, -0.1, 0.05} // Q1, Q2, Q3, Q4
	return seasonalValues[quarter]
}

func (hdg *HistoricalDataGenerator) calculateQuarterlyFactor(sampleTime time.Time) float64 {
	// Monthly pattern within quarter
	monthInQuarter := (sampleTime.Month() - 1) % 3
	quarterlyValues := []float64{0.05, 0.0, -0.05}
	return quarterlyValues[monthInQuarter]
}

func (hdg *HistoricalDataGenerator) calculateTrendFactor(sampleTime time.Time, config DataGenerationConfig) float64 {
	// Linear trend over time
	yearsSinceStart := sampleTime.Sub(time.Now().Add(-config.TimeRange)).Hours() / (24 * 365)
	return yearsSinceStart * 0.02 // 2% per year trend
}

// External events methods
func (hdg *HistoricalDataGenerator) generateExternalEvents(sampleTime time.Time, config DataGenerationConfig) []ExternalEvent {
	events := make([]ExternalEvent, 0)

	// Random chance of external events
	if rand.Float64() < 0.1 { // 10% chance
		eventTypes := []string{"economic_crisis", "regulatory_change", "market_disruption", "natural_disaster", "pandemic", "technology_breakthrough"}
		eventType := eventTypes[rand.Intn(len(eventTypes))]

		impact := rand.Float64()*0.3 - 0.15 // -15% to +15% impact

		events = append(events, ExternalEvent{
			Type:        eventType,
			Impact:      impact,
			Timestamp:   sampleTime,
			Description: fmt.Sprintf("%s event affecting business risk", eventType),
		})
	}

	return events
}

func (hdg *HistoricalDataGenerator) applyExternalEvents(baseRisk float64, events []ExternalEvent) float64 {
	risk := baseRisk

	for _, event := range events {
		risk += event.Impact
	}

	return math.Max(0, math.Min(1, risk))
}

func (hdg *HistoricalDataGenerator) addNoise(risk float64, noiseLevel float64) float64 {
	noise := rand.NormFloat64() * noiseLevel
	return math.Max(0, math.Min(1, risk+noise))
}

// calculateCategoryRisk calculates risk for a specific category
func (hdg *HistoricalDataGenerator) calculateCategoryRisk(riskFactors map[string]float64, category RiskCategory) float64 {
	// This would be implemented based on the specific risk category
	// For now, return a weighted combination of relevant factors
	return category.BaseRisk + rand.NormFloat64()*category.Volatility
}
