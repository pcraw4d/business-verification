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

// BuildSequence builds a temporal sequence for LSTM input with enhanced 6-12 month forecasting
func (tfb *TemporalFeatureBuilder) BuildSequence(business *models.RiskAssessmentRequest, sequenceLength int) ([][]float64, error) {
	// Enhanced temporal data generation for 6-12 month forecasting
	// In a full implementation, this would:
	// 1. Try to fetch real historical data from cache/DB
	// 2. Generate synthetic data based on business characteristics
	// 3. Blend real and synthetic data appropriately
	// 4. Apply trend decomposition and seasonal adjustment

	sequence := make([][]float64, sequenceLength)

	// Generate enhanced synthetic time-series data with multi-step forecasting capabilities
	for i := 0; i < sequenceLength; i++ {
		// Create enhanced feature vector for this timestep
		features := tfb.generateEnhancedTimestepFeatures(business, i, sequenceLength)
		sequence[i] = features
	}

	// Apply trend decomposition and seasonal adjustment
	sequence = tfb.applyTrendDecomposition(sequence)
	sequence = tfb.applySeasonalAdjustment(sequence)

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

// generateEnhancedTimestepFeatures generates enhanced features for 6-12 month forecasting
func (tfb *TemporalFeatureBuilder) generateEnhancedTimestepFeatures(business *models.RiskAssessmentRequest, timestep, totalSteps int) []float64 {
	features := make([]float64, 25) // Expanded to 25 features for enhanced forecasting

	// Normalize timestep to 0-1 range
	timeProgress := float64(timestep) / float64(totalSteps-1)

	// Features 0-19: Original features (enhanced)
	baseFeatures := tfb.generateTimestepFeatures(business, timestep, totalSteps)
	copy(features[:20], baseFeatures)

	// Feature 20: Multi-step ahead prediction confidence
	features[20] = tfb.calculateMultiStepConfidence(timestep, totalSteps)

	// Feature 21: Rolling window volatility (6-month)
	features[21] = tfb.calculateRollingVolatility(timestep, 6)

	// Feature 22: Rolling window volatility (12-month)
	features[22] = tfb.calculateRollingVolatility(timestep, 12)

	// Feature 23: Trend acceleration (second derivative)
	features[23] = tfb.calculateTrendAcceleration(timestep, timeProgress)

	// Feature 24: Seasonal strength indicator
	features[24] = tfb.calculateSeasonalStrength(timestep)

	return features
}

// generateAdvancedTemporalFeatures generates additional advanced temporal features for 6-12 month forecasting
func (tfb *TemporalFeatureBuilder) generateAdvancedTemporalFeatures(business *models.RiskAssessmentRequest, timestep, totalSteps int, historicalSequence [][]float64) []float64 {
	features := make([]float64, 10) // 10 additional advanced features

	// Feature 0: Autocorrelation coefficient (lag-1)
	features[0] = tfb.calculateAutocorrelation(historicalSequence, 1)

	// Feature 1: Momentum indicator (rate of change)
	features[1] = tfb.calculateMomentum(historicalSequence, timestep)

	// Feature 2: Mean reversion tendency
	features[2] = tfb.calculateMeanReversion(historicalSequence, timestep)

	// Feature 3: Volatility clustering indicator
	features[3] = tfb.calculateVolatilityClustering(historicalSequence)

	// Feature 4: Trend persistence score
	features[4] = tfb.calculateTrendPersistence(historicalSequence)

	// Feature 5: Cyclical pattern strength
	features[5] = tfb.calculateCyclicalStrength(historicalSequence, timestep)

	// Feature 6: Regime change probability
	features[6] = tfb.calculateRegimeChangeProbability(historicalSequence, timestep)

	// Feature 7: Long-term memory indicator
	features[7] = tfb.calculateLongTermMemory(historicalSequence)

	// Feature 8: Structural break detection
	features[8] = tfb.calculateStructuralBreak(historicalSequence, timestep)

	// Feature 9: Cross-correlation with economic indicators
	features[9] = tfb.calculateEconomicCorrelation(business, timestep)

	return features
}

// calculateAutocorrelation calculates autocorrelation coefficient for given lag
func (tfb *TemporalFeatureBuilder) calculateAutocorrelation(sequence [][]float64, lag int) float64 {
	if len(sequence) < lag+2 {
		return 0.0
	}

	// Extract a representative time series (using first feature as proxy)
	series := make([]float64, len(sequence))
	for i, timestep := range sequence {
		if len(timestep) > 0 {
			series[i] = timestep[0] // Use first feature as proxy
		}
	}

	// Calculate autocorrelation
	n := len(series)
	if n < lag+2 {
		return 0.0
	}

	// Calculate mean
	mean := 0.0
	for _, val := range series {
		mean += val
	}
	mean /= float64(n)

	// Calculate autocorrelation
	numerator := 0.0
	denominator := 0.0

	for i := 0; i < n-lag; i++ {
		numerator += (series[i] - mean) * (series[i+lag] - mean)
	}

	for i := 0; i < n; i++ {
		denominator += (series[i] - mean) * (series[i] - mean)
	}

	if denominator == 0 {
		return 0.0
	}

	return numerator / denominator
}

// calculateMomentum calculates momentum indicator
func (tfb *TemporalFeatureBuilder) calculateMomentum(sequence [][]float64, currentTimestep int) float64 {
	if len(sequence) < 3 || currentTimestep < 2 {
		return 0.0
	}

	// Use first feature as proxy for the time series
	if len(sequence[currentTimestep]) == 0 || len(sequence[currentTimestep-1]) == 0 || len(sequence[currentTimestep-2]) == 0 {
		return 0.0
	}

	current := sequence[currentTimestep][0]
	previous := sequence[currentTimestep-1][0]
	twoBack := sequence[currentTimestep-2][0]

	// Calculate momentum as weighted average of recent changes
	momentum := (current - previous) + 0.5*(previous-twoBack)
	return math.Max(-1.0, math.Min(1.0, momentum))
}

// calculateMeanReversion calculates mean reversion tendency
func (tfb *TemporalFeatureBuilder) calculateMeanReversion(sequence [][]float64, currentTimestep int) float64 {
	if len(sequence) < 5 {
		return 0.0
	}

	// Calculate mean of recent values
	mean := 0.0
	count := 0
	for i := max(0, currentTimestep-5); i < currentTimestep; i++ {
		if len(sequence[i]) > 0 {
			mean += sequence[i][0]
			count++
		}
	}

	if count == 0 {
		return 0.0
	}
	mean /= float64(count)

	// Calculate current deviation from mean
	if len(sequence[currentTimestep]) == 0 {
		return 0.0
	}
	current := sequence[currentTimestep][0]
	deviation := current - mean

	// Mean reversion tendency (negative correlation with recent deviation)
	return math.Max(-1.0, math.Min(1.0, -deviation*2.0))
}

// calculateVolatilityClustering calculates volatility clustering indicator
func (tfb *TemporalFeatureBuilder) calculateVolatilityClustering(sequence [][]float64) float64 {
	if len(sequence) < 10 {
		return 0.0
	}

	// Calculate returns (first differences)
	returns := make([]float64, len(sequence)-1)
	for i := 1; i < len(sequence); i++ {
		if len(sequence[i]) > 0 && len(sequence[i-1]) > 0 {
			returns[i-1] = sequence[i][0] - sequence[i-1][0]
		}
	}

	// Calculate volatility (standard deviation of returns)
	mean := 0.0
	for _, ret := range returns {
		mean += ret
	}
	mean /= float64(len(returns))

	variance := 0.0
	for _, ret := range returns {
		variance += (ret - mean) * (ret - mean)
	}
	variance /= float64(len(returns))
	volatility := math.Sqrt(variance)

	// Volatility clustering: high volatility tends to be followed by high volatility
	// Calculate correlation between current and lagged volatility
	if len(returns) < 5 {
		return 0.0
	}

	recentVol := 0.0
	for i := max(0, len(returns)-3); i < len(returns); i++ {
		recentVol += math.Abs(returns[i])
	}
	recentVol /= 3.0

	// Normalize and return clustering indicator
	return math.Max(0.0, math.Min(1.0, recentVol/volatility))
}

// calculateTrendPersistence calculates trend persistence score
func (tfb *TemporalFeatureBuilder) calculateTrendPersistence(sequence [][]float64) float64 {
	if len(sequence) < 5 {
		return 0.0
	}

	// Calculate trend direction consistency
	trends := make([]int, len(sequence)-1)
	for i := 1; i < len(sequence); i++ {
		if len(sequence[i]) > 0 && len(sequence[i-1]) > 0 {
			if sequence[i][0] > sequence[i-1][0] {
				trends[i-1] = 1 // Upward
			} else {
				trends[i-1] = -1 // Downward
			}
		}
	}

	// Calculate persistence as consistency of trend direction
	if len(trends) == 0 {
		return 0.0
	}

	positiveCount := 0
	negativeCount := 0
	for _, trend := range trends {
		if trend > 0 {
			positiveCount++
		} else if trend < 0 {
			negativeCount++
		}
	}

	total := positiveCount + negativeCount
	if total == 0 {
		return 0.0
	}

	// Persistence is the maximum of positive or negative consistency
	maxConsistency := math.Max(float64(positiveCount), float64(negativeCount))
	return maxConsistency / float64(total)
}

// calculateCyclicalStrength calculates cyclical pattern strength
func (tfb *TemporalFeatureBuilder) calculateCyclicalStrength(sequence [][]float64, currentTimestep int) float64 {
	if len(sequence) < 8 {
		return 0.0
	}

	// Extract time series
	series := make([]float64, len(sequence))
	for i, timestep := range sequence {
		if len(timestep) > 0 {
			series[i] = timestep[0]
		}
	}

	// Calculate cyclical strength using FFT-like approach (simplified)
	// Look for periodic patterns in the data
	cycleLengths := []int{3, 6, 12} // Quarterly, semi-annual, annual cycles
	maxStrength := 0.0

	for _, cycleLength := range cycleLengths {
		if len(series) < cycleLength*2 {
			continue
		}

		// Calculate correlation with lagged version
		strength := tfb.calculateAutocorrelation(sequence, cycleLength)
		maxStrength = math.Max(maxStrength, math.Abs(strength))
	}

	return maxStrength
}

// calculateRegimeChangeProbability calculates probability of regime change
func (tfb *TemporalFeatureBuilder) calculateRegimeChangeProbability(sequence [][]float64, currentTimestep int) float64 {
	if len(sequence) < 10 {
		return 0.0
	}

	// Calculate rolling statistics to detect regime changes
	windowSize := 5
	if currentTimestep < windowSize {
		return 0.0
	}

	// Calculate mean and variance for recent window
	recentMean := 0.0
	recentVariance := 0.0
	for i := currentTimestep - windowSize; i < currentTimestep; i++ {
		if len(sequence[i]) > 0 {
			recentMean += sequence[i][0]
		}
	}
	recentMean /= float64(windowSize)

	for i := currentTimestep - windowSize; i < currentTimestep; i++ {
		if len(sequence[i]) > 0 {
			diff := sequence[i][0] - recentMean
			recentVariance += diff * diff
		}
	}
	recentVariance /= float64(windowSize)

	// Calculate historical mean and variance
	historicalMean := 0.0
	historicalVariance := 0.0
	historicalCount := 0
	for i := 0; i < currentTimestep-windowSize; i++ {
		if len(sequence[i]) > 0 {
			historicalMean += sequence[i][0]
			historicalCount++
		}
	}
	if historicalCount > 0 {
		historicalMean /= float64(historicalCount)

		for i := 0; i < currentTimestep-windowSize; i++ {
			if len(sequence[i]) > 0 {
				diff := sequence[i][0] - historicalMean
				historicalVariance += diff * diff
			}
		}
		historicalVariance /= float64(historicalCount)
	}

	// Regime change probability based on statistical differences
	meanChange := math.Abs(recentMean - historicalMean)
	varianceChange := math.Abs(recentVariance - historicalVariance)

	// Normalize and combine
	regimeChange := (meanChange + varianceChange*0.5) / 2.0
	return math.Max(0.0, math.Min(1.0, regimeChange))
}

// calculateLongTermMemory calculates long-term memory indicator
func (tfb *TemporalFeatureBuilder) calculateLongTermMemory(sequence [][]float64) float64 {
	if len(sequence) < 10 {
		return 0.0
	}

	// Calculate Hurst exponent approximation
	// Long-term memory is indicated by Hurst exponent > 0.5
	series := make([]float64, len(sequence))
	for i, timestep := range sequence {
		if len(timestep) > 0 {
			series[i] = timestep[0]
		}
	}

	// Simplified Hurst calculation using R/S analysis
	n := len(series)
	if n < 4 {
		return 0.5
	}

	// Calculate mean
	mean := 0.0
	for _, val := range series {
		mean += val
	}
	mean /= float64(n)

	// Calculate cumulative deviations
	cumulativeDeviations := make([]float64, n)
	cumulativeDeviations[0] = series[0] - mean
	for i := 1; i < n; i++ {
		cumulativeDeviations[i] = cumulativeDeviations[i-1] + (series[i] - mean)
	}

	// Calculate range
	minCum := cumulativeDeviations[0]
	maxCum := cumulativeDeviations[0]
	for _, val := range cumulativeDeviations {
		if val < minCum {
			minCum = val
		}
		if val > maxCum {
			maxCum = val
		}
	}
	rangeVal := maxCum - minCum

	// Calculate standard deviation
	stdDev := 0.0
	for _, val := range series {
		diff := val - mean
		stdDev += diff * diff
	}
	stdDev = math.Sqrt(stdDev / float64(n))

	// Calculate R/S ratio
	if stdDev == 0 {
		return 0.5
	}
	rsRatio := rangeVal / stdDev

	// Approximate Hurst exponent
	hurst := math.Log(rsRatio) / math.Log(float64(n))
	hurst = math.Max(0.0, math.Min(1.0, hurst))

	return hurst
}

// calculateStructuralBreak calculates structural break detection
func (tfb *TemporalFeatureBuilder) calculateStructuralBreak(sequence [][]float64, currentTimestep int) float64 {
	if len(sequence) < 8 {
		return 0.0
	}

	// Use Chow test approximation to detect structural breaks
	// Split data into two periods
	splitPoint := currentTimestep / 2
	if splitPoint < 2 {
		return 0.0
	}

	// Calculate means for each period
	mean1 := 0.0
	count1 := 0
	for i := 0; i < splitPoint; i++ {
		if len(sequence[i]) > 0 {
			mean1 += sequence[i][0]
			count1++
		}
	}
	if count1 > 0 {
		mean1 /= float64(count1)
	}

	mean2 := 0.0
	count2 := 0
	for i := splitPoint; i < currentTimestep; i++ {
		if len(sequence[i]) > 0 {
			mean2 += sequence[i][0]
			count2++
		}
	}
	if count2 > 0 {
		mean2 /= float64(count2)
	}

	// Calculate structural break indicator
	breakIndicator := math.Abs(mean2 - mean1)
	return math.Max(0.0, math.Min(1.0, breakIndicator))
}

// calculateEconomicCorrelation calculates correlation with economic indicators
func (tfb *TemporalFeatureBuilder) calculateEconomicCorrelation(business *models.RiskAssessmentRequest, timestep int) float64 {
	// Simulate correlation with economic indicators based on industry
	industryCorrelations := map[string]float64{
		"technology":    0.7, // High correlation with tech sector
		"finance":       0.8, // High correlation with financial markets
		"healthcare":    0.4, // Moderate correlation
		"retail":        0.6, // Moderate correlation with consumer spending
		"manufacturing": 0.5, // Moderate correlation with industrial production
		"construction":  0.3, // Lower correlation
		"restaurant":    0.5, // Moderate correlation
		"consulting":    0.4, // Moderate correlation
		"education":     0.2, // Lower correlation
		"default":       0.5, // Default moderate correlation
	}

	baseCorrelation := industryCorrelations["default"]
	if corr, exists := industryCorrelations[business.Industry]; exists {
		baseCorrelation = corr
	}

	// Add some time-based variation
	timeVariation := math.Sin(float64(timestep)*0.1) * 0.1
	return math.Max(0.0, math.Min(1.0, baseCorrelation+timeVariation))
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// calculateMultiStepConfidence calculates confidence for multi-step ahead predictions
func (tfb *TemporalFeatureBuilder) calculateMultiStepConfidence(timestep, totalSteps int) float64 {
	// Confidence decreases as we look further into the future
	// More recent data points have higher confidence
	recencyFactor := 1.0 - (float64(timestep) / float64(totalSteps-1))

	// Base confidence with some randomness
	baseConfidence := 0.8 + (rand.Float64()-0.5)*0.2

	return math.Max(0.3, math.Min(1.0, baseConfidence*recencyFactor))
}

// calculateRollingVolatility calculates rolling window volatility
func (tfb *TemporalFeatureBuilder) calculateRollingVolatility(timestep, windowSize int) float64 {
	// Simulate rolling volatility calculation
	// In practice, this would use actual historical data

	// Generate synthetic volatility with some persistence
	baseVolatility := 0.3 + math.Sin(float64(timestep)*0.5)*0.2

	// Add some noise
	noise := (rand.Float64() - 0.5) * 0.1

	return math.Max(0.0, math.Min(1.0, baseVolatility+noise))
}

// calculateTrendAcceleration calculates trend acceleration (second derivative)
func (tfb *TemporalFeatureBuilder) calculateTrendAcceleration(timestep int, timeProgress float64) float64 {
	// Simulate trend acceleration
	// Positive values indicate accelerating trends, negative indicate decelerating

	acceleration := math.Sin(timeProgress*math.Pi*3) * 0.3

	// Add some noise
	acceleration += (rand.Float64() - 0.5) * 0.1

	return (acceleration + 1) / 2 // Normalize to 0-1
}

// calculateSeasonalStrength calculates the strength of seasonal patterns
func (tfb *TemporalFeatureBuilder) calculateSeasonalStrength(timestep int) float64 {
	// Calculate how strong seasonal patterns are
	quarter := timestep % 4

	// Different quarters have different seasonal strengths
	seasonalStrengths := []float64{0.8, 0.6, 0.7, 0.9} // Q1, Q2, Q3, Q4

	baseStrength := seasonalStrengths[quarter]

	// Add some variation
	variation := (rand.Float64() - 0.5) * 0.2

	return math.Max(0.0, math.Min(1.0, baseStrength+variation))
}

// applyTrendDecomposition applies trend decomposition to the sequence
func (tfb *TemporalFeatureBuilder) applyTrendDecomposition(sequence [][]float64) [][]float64 {
	if len(sequence) < 3 {
		return sequence
	}

	// Apply simple trend decomposition
	// In practice, this would use more sophisticated methods like STL decomposition

	decomposed := make([][]float64, len(sequence))

	for i, timestep := range sequence {
		decomposed[i] = make([]float64, len(timestep))
		copy(decomposed[i], timestep)

		// Apply trend smoothing to key features
		if i > 0 && i < len(sequence)-1 {
			// Smooth trend-sensitive features (indices 1, 3, 6)
			for _, featureIdx := range []int{1, 3, 6} {
				if featureIdx < len(timestep) {
					// Simple 3-point moving average
					prev := sequence[i-1][featureIdx]
					curr := sequence[i][featureIdx]
					next := sequence[i+1][featureIdx]

					decomposed[i][featureIdx] = (prev + curr + next) / 3.0
				}
			}
		}
	}

	return decomposed
}

// applySeasonalAdjustment applies seasonal adjustment to the sequence
func (tfb *TemporalFeatureBuilder) applySeasonalAdjustment(sequence [][]float64) [][]float64 {
	if len(sequence) < 4 {
		return sequence
	}

	adjusted := make([][]float64, len(sequence))

	// Calculate seasonal indices for each quarter
	seasonalIndices := tfb.calculateSeasonalIndices(sequence)

	for i, timestep := range sequence {
		adjusted[i] = make([]float64, len(timestep))
		copy(adjusted[i], timestep)

		quarter := i % 4
		seasonalIndex := seasonalIndices[quarter]

		// Apply seasonal adjustment to seasonal-sensitive features
		for _, featureIdx := range []int{4, 12, 16} { // Seasonal features
			if featureIdx < len(timestep) {
				// Remove seasonal component
				adjusted[i][featureIdx] = timestep[featureIdx] / seasonalIndex
			}
		}
	}

	return adjusted
}

// calculateSeasonalIndices calculates seasonal indices for each quarter
func (tfb *TemporalFeatureBuilder) calculateSeasonalIndices(sequence [][]float64) []float64 {
	indices := make([]float64, 4)

	// Calculate average for each quarter
	for quarter := 0; quarter < 4; quarter++ {
		var sum float64
		var count int

		for i := quarter; i < len(sequence); i += 4 {
			if len(sequence[i]) > 4 { // Use seasonal feature (index 4)
				sum += sequence[i][4]
				count++
			}
		}

		if count > 0 {
			indices[quarter] = sum / float64(count)
		} else {
			indices[quarter] = 1.0 // Default neutral index
		}
	}

	// Normalize indices so they average to 1.0
	total := 0.0
	for _, idx := range indices {
		total += idx
	}

	if total > 0 {
		normalizationFactor := 4.0 / total
		for i := range indices {
			indices[i] *= normalizationFactor
		}
	}

	return indices
}

// BuildMultiStepSequence builds a sequence for multi-step ahead forecasting
func (tfb *TemporalFeatureBuilder) BuildMultiStepSequence(business *models.RiskAssessmentRequest, sequenceLength int, forecastHorizon int) ([][]float64, error) {
	// Build base sequence
	baseSequence, err := tfb.BuildSequence(business, sequenceLength)
	if err != nil {
		return nil, err
	}

	// Extend sequence for multi-step forecasting
	extendedSequence := make([][]float64, sequenceLength+forecastHorizon)

	// Copy base sequence
	for i, timestep := range baseSequence {
		extendedSequence[i] = make([]float64, len(timestep))
		copy(extendedSequence[i], timestep)
	}

	// Generate forecast steps
	for i := sequenceLength; i < sequenceLength+forecastHorizon; i++ {
		forecastFeatures := tfb.generateForecastFeatures(business, i, sequenceLength, forecastHorizon)
		extendedSequence[i] = forecastFeatures
	}

	return extendedSequence, nil
}

// generateForecastFeatures generates features for forecast steps
func (tfb *TemporalFeatureBuilder) generateForecastFeatures(business *models.RiskAssessmentRequest, timestep, baseLength, forecastHorizon int) []float64 {
	features := make([]float64, 25)

	// Use the last known values as base
	lastTimestep := baseLength - 1

	// Copy base features from last timestep
	baseFeatures := tfb.generateEnhancedTimestepFeatures(business, lastTimestep, baseLength)
	copy(features, baseFeatures)

	// Apply forecast-specific adjustments
	forecastProgress := float64(timestep-baseLength) / float64(forecastHorizon-1)

	// Adjust trend-sensitive features
	trendAdjustment := 1.0 + (forecastProgress * 0.1) // 10% trend continuation
	features[3] *= trendAdjustment                    // Trend feature

	// Adjust seasonal features
	seasonalAdjustment := 1.0 + math.Sin(forecastProgress*math.Pi*2)*0.05
	features[4] *= seasonalAdjustment // Seasonal feature

	// Increase uncertainty for longer forecasts
	uncertaintyIncrease := forecastProgress * 0.2
	features[20] = math.Max(0.1, features[20]-uncertaintyIncrease) // Confidence feature

	return features
}
