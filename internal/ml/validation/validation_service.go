package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"
)

// ValidationService orchestrates ML model validation processes
type ValidationService struct {
	crossValidator *CrossValidator
	dataGenerator  *HistoricalDataGenerator
	logger         *zap.Logger
}

// ValidationConfig configures the validation process
type ValidationConfig struct {
	CrossValidation CrossValidationConfig `json:"cross_validation"`
	DataGeneration  DataGenerationConfig  `json:"data_generation"`
	OutputFormat    string                `json:"output_format"`
	SaveResults     bool                  `json:"save_results"`
	ResultsPath     string                `json:"results_path"`
}

// CrossValidationConfig configures cross-validation parameters
type CrossValidationConfig struct {
	KFolds          int     `json:"k_folds"`
	ConfidenceLevel float64 `json:"confidence_level"`
	RandomSeed      int64   `json:"random_seed"`
	ParallelFolds   bool    `json:"parallel_folds"`
	MaxConcurrency  int     `json:"max_concurrency"`
}

// ValidationReport contains comprehensive validation results
type ValidationReport struct {
	Summary         ValidationSummary     `json:"summary"`
	CrossValidation *ValidationResult     `json:"cross_validation"`
	HistoricalData  HistoricalDataSummary `json:"historical_data"`
	ModelComparison []ModelComparison     `json:"model_comparison"`
	Recommendations []Recommendation      `json:"recommendations"`
	GeneratedAt     time.Time             `json:"generated_at"`
	ValidationTime  time.Duration         `json:"validation_time"`
	Configuration   ValidationConfig      `json:"configuration"`
}

// ValidationSummary provides a high-level overview of validation results
type ValidationSummary struct {
	OverallScore     float64 `json:"overall_score"`
	AccuracyScore    float64 `json:"accuracy_score"`
	ReliabilityScore float64 `json:"reliability_score"`
	PerformanceScore float64 `json:"performance_score"`
	Recommendation   string  `json:"recommendation"`
	RiskLevel        string  `json:"risk_level"`
	ConfidenceLevel  float64 `json:"confidence_level"`
}

// HistoricalDataSummary summarizes the historical data used for validation
type HistoricalDataSummary struct {
	TotalSamples     int                `json:"total_samples"`
	TimeRange        time.Duration      `json:"time_range"`
	Industries       map[string]int     `json:"industries"`
	Countries        map[string]int     `json:"countries"`
	BusinessSizes    map[string]int     `json:"business_sizes"`
	RiskDistribution map[string]int     `json:"risk_distribution"`
	DataQuality      DataQualityMetrics `json:"data_quality"`
}

// DataQualityMetrics measures the quality of historical data
type DataQualityMetrics struct {
	Completeness   float64 `json:"completeness"`
	Consistency    float64 `json:"consistency"`
	Accuracy       float64 `json:"accuracy"`
	Timeliness     float64 `json:"timeliness"`
	Relevance      float64 `json:"relevance"`
	OverallQuality float64 `json:"overall_quality"`
}

// ModelComparison compares different models
type ModelComparison struct {
	ModelName    string   `json:"model_name"`
	Metrics      Metrics  `json:"metrics"`
	OverallScore float64  `json:"overall_score"`
	Rank         int      `json:"rank"`
	Strengths    []string `json:"strengths"`
	Weaknesses   []string `json:"weaknesses"`
	UseCase      string   `json:"use_case"`
}

// Recommendation provides actionable recommendations
type Recommendation struct {
	Type        string   `json:"type"`
	Priority    string   `json:"priority"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`
	Effort      string   `json:"effort"`
	Timeline    string   `json:"timeline"`
	Actions     []string `json:"actions"`
}

// NewValidationService creates a new validation service
func NewValidationService(logger *zap.Logger) *ValidationService {
	return &ValidationService{
		crossValidator: NewCrossValidator(logger),
		dataGenerator:  NewHistoricalDataGenerator(logger),
		logger:         logger,
	}
}

// ValidateModel performs comprehensive model validation
func (vs *ValidationService) ValidateModel(
	ctx context.Context,
	model ModelValidator,
	config ValidationConfig,
) (*ValidationReport, error) {
	startTime := time.Now()

	vs.logger.Info("Starting comprehensive model validation",
		zap.String("model", model.GetName()),
		zap.Int("k_folds", config.CrossValidation.KFolds),
		zap.Int("total_samples", config.DataGeneration.TotalSamples))

	// Generate historical data
	vs.logger.Info("Generating historical data for validation")
	samples, historicalSamples, err := vs.dataGenerator.GenerateHistoricalData(ctx, config.DataGeneration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate historical data: %w", err)
	}

	// Perform cross-validation
	vs.logger.Info("Performing cross-validation")
	cvResult, err := vs.crossValidator.CrossValidate(
		ctx,
		model,
		samples,
		config.CrossValidation.KFolds,
		config.CrossValidation.ConfidenceLevel,
	)
	if err != nil {
		return nil, fmt.Errorf("cross-validation failed: %w", err)
	}

	// Generate historical data summary
	historicalSummary := vs.generateHistoricalDataSummary(historicalSamples, config.DataGeneration)

	// Calculate validation summary
	summary := vs.calculateValidationSummary(cvResult, historicalSummary)

	// Generate model comparison (single model for now)
	modelComparison := vs.generateModelComparison(cvResult)

	// Generate recommendations
	recommendations := vs.generateRecommendations(cvResult, historicalSummary, summary)

	validationTime := time.Since(startTime)

	report := &ValidationReport{
		Summary:         summary,
		CrossValidation: cvResult,
		HistoricalData:  historicalSummary,
		ModelComparison: modelComparison,
		Recommendations: recommendations,
		GeneratedAt:     time.Now(),
		ValidationTime:  validationTime,
		Configuration:   config,
	}

	// Save results if requested
	if config.SaveResults {
		if err := vs.saveValidationReport(report, config.ResultsPath); err != nil {
			vs.logger.Warn("Failed to save validation report", zap.Error(err))
		}
	}

	vs.logger.Info("Model validation completed",
		zap.String("model", model.GetName()),
		zap.Float64("overall_score", summary.OverallScore),
		zap.String("recommendation", summary.Recommendation),
		zap.Duration("validation_time", validationTime))

	return report, nil
}

// ValidateMultipleModels compares multiple models
func (vs *ValidationService) ValidateMultipleModels(
	ctx context.Context,
	models []ModelValidator,
	config ValidationConfig,
) (*ValidationReport, error) {
	vs.logger.Info("Starting multi-model validation",
		zap.Int("model_count", len(models)))

	// Generate historical data once
	samples, historicalSamples, err := vs.dataGenerator.GenerateHistoricalData(ctx, config.DataGeneration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate historical data: %w", err)
	}

	// Validate each model
	modelResults := make([]*ValidationResult, len(models))
	for i, model := range models {
		vs.logger.Info("Validating model",
			zap.String("model", model.GetName()),
			zap.Int("index", i+1),
			zap.Int("total", len(models)))

		result, err := vs.crossValidator.CrossValidate(
			ctx,
			model,
			samples,
			config.CrossValidation.KFolds,
			config.CrossValidation.ConfidenceLevel,
		)
		if err != nil {
			vs.logger.Error("Model validation failed",
				zap.String("model", model.GetName()),
				zap.Error(err))
			continue
		}

		modelResults[i] = result
	}

	// Find best model
	bestResult := vs.findBestModel(modelResults)
	if bestResult == nil {
		return nil, fmt.Errorf("no models validated successfully")
	}

	// Generate comprehensive report
	historicalSummary := vs.generateHistoricalDataSummary(historicalSamples, config.DataGeneration)
	summary := vs.calculateValidationSummary(bestResult, historicalSummary)
	modelComparison := vs.generateMultiModelComparison(modelResults)
	recommendations := vs.generateRecommendations(bestResult, historicalSummary, summary)

	report := &ValidationReport{
		Summary:         summary,
		CrossValidation: bestResult,
		HistoricalData:  historicalSummary,
		ModelComparison: modelComparison,
		Recommendations: recommendations,
		GeneratedAt:     time.Now(),
		ValidationTime:  time.Since(time.Now()),
		Configuration:   config,
	}

	return report, nil
}

// generateHistoricalDataSummary creates a summary of historical data
func (vs *ValidationService) generateHistoricalDataSummary(
	samples []HistoricalSample,
	config DataGenerationConfig,
) HistoricalDataSummary {
	industries := make(map[string]int)
	countries := make(map[string]int)
	businessSizes := make(map[string]int)
	riskDistribution := make(map[string]int)

	for _, sample := range samples {
		// Count industries
		industries[sample.Industry]++

		// Count countries
		countries[sample.Country]++

		// Count business sizes
		businessSizes[sample.BusinessSize]++

		// Count risk levels
		riskLevel := vs.categorizeRisk(sample.ActualOutcome)
		riskDistribution[riskLevel]++
	}

	// Calculate data quality metrics
	dataQuality := vs.calculateDataQuality(samples)

	return HistoricalDataSummary{
		TotalSamples:     len(samples),
		TimeRange:        config.TimeRange,
		Industries:       industries,
		Countries:        countries,
		BusinessSizes:    businessSizes,
		RiskDistribution: riskDistribution,
		DataQuality:      dataQuality,
	}
}

// categorizeRisk categorizes risk into levels
func (vs *ValidationService) categorizeRisk(risk float64) string {
	switch {
	case risk < 0.2:
		return "Low"
	case risk < 0.4:
		return "Medium-Low"
	case risk < 0.6:
		return "Medium"
	case risk < 0.8:
		return "Medium-High"
	default:
		return "High"
	}
}

// calculateDataQuality calculates data quality metrics
func (vs *ValidationService) calculateDataQuality(samples []HistoricalSample) DataQualityMetrics {
	if len(samples) == 0 {
		return DataQualityMetrics{}
	}

	// Calculate completeness (percentage of non-zero values)
	completeness := vs.calculateCompleteness(samples)

	// Calculate consistency (low variance in similar samples)
	consistency := vs.calculateConsistency(samples)

	// Calculate accuracy (based on feature consistency)
	accuracy := vs.calculateAccuracy(samples)

	// Calculate timeliness (based on data recency)
	timeliness := vs.calculateTimeliness(samples)

	// Calculate relevance (based on feature importance)
	relevance := vs.calculateRelevance(samples)

	// Overall quality is weighted average
	overallQuality := (completeness*0.25 + consistency*0.25 + accuracy*0.25 + timeliness*0.15 + relevance*0.10)

	return DataQualityMetrics{
		Completeness:   completeness,
		Consistency:    consistency,
		Accuracy:       accuracy,
		Timeliness:     timeliness,
		Relevance:      relevance,
		OverallQuality: overallQuality,
	}
}

// calculateCompleteness calculates data completeness
func (vs *ValidationService) calculateCompleteness(samples []HistoricalSample) float64 {
	if len(samples) == 0 {
		return 0
	}

	totalFields := 0
	nonZeroFields := 0

	for _, sample := range samples {
		for _, value := range sample.RiskFactors {
			totalFields++
			if value != 0 {
				nonZeroFields++
			}
		}
	}

	if totalFields == 0 {
		return 0
	}

	return float64(nonZeroFields) / float64(totalFields)
}

// calculateConsistency calculates data consistency
func (vs *ValidationService) calculateConsistency(samples []HistoricalSample) float64 {
	if len(samples) < 2 {
		return 1.0
	}

	// Group samples by similar characteristics
	groups := make(map[string][]HistoricalSample)
	for _, sample := range samples {
		key := fmt.Sprintf("%s_%s_%s", sample.Industry, sample.Country, sample.BusinessSize)
		groups[key] = append(groups[key], sample)
	}

	totalVariance := 0.0
	totalGroups := 0

	for _, group := range groups {
		if len(group) > 1 {
			variance := vs.calculateGroupVariance(group)
			totalVariance += variance
			totalGroups++
		}
	}

	if totalGroups == 0 {
		return 1.0
	}

	avgVariance := totalVariance / float64(totalGroups)
	// Convert variance to consistency score (lower variance = higher consistency)
	return math.Max(0, 1-avgVariance)
}

// calculateGroupVariance calculates variance within a group
func (vs *ValidationService) calculateGroupVariance(group []HistoricalSample) float64 {
	if len(group) < 2 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, sample := range group {
		sum += sample.ActualOutcome
	}
	mean := sum / float64(len(group))

	// Calculate variance
	variance := 0.0
	for _, sample := range group {
		diff := sample.ActualOutcome - mean
		variance += diff * diff
	}
	variance /= float64(len(group) - 1)

	return variance
}

// calculateAccuracy calculates data accuracy
func (vs *ValidationService) calculateAccuracy(samples []HistoricalSample) float64 {
	// This is a simplified accuracy calculation
	// In practice, you would compare against known ground truth
	return 0.85 // Placeholder value
}

// calculateTimeliness calculates data timeliness
func (vs *ValidationService) calculateTimeliness(samples []HistoricalSample) float64 {
	if len(samples) == 0 {
		return 0
	}

	now := time.Now()
	totalScore := 0.0

	for _, sample := range samples {
		age := now.Sub(sample.Timestamp)
		// Score decreases with age (newer data is better)
		score := math.Max(0, 1-age.Hours()/(365*24)) // 1 year = 0 score
		totalScore += score
	}

	return totalScore / float64(len(samples))
}

// calculateRelevance calculates data relevance
func (vs *ValidationService) calculateRelevance(samples []HistoricalSample) float64 {
	// This is a simplified relevance calculation
	// In practice, you would analyze feature importance and correlation
	return 0.80 // Placeholder value
}

// calculateValidationSummary calculates overall validation summary
func (vs *ValidationService) calculateValidationSummary(
	cvResult *ValidationResult,
	historicalSummary HistoricalDataSummary,
) ValidationSummary {
	// Calculate scores based on cross-validation results
	accuracyScore := cvResult.OverallMetrics.MeanAccuracy
	reliabilityScore := 1.0 - cvResult.OverallMetrics.StdAccuracy // Lower std = higher reliability
	performanceScore := (cvResult.OverallMetrics.MeanF1Score + cvResult.OverallMetrics.MeanAUC) / 2.0

	// Overall score is weighted combination
	overallScore := (accuracyScore*0.4 + reliabilityScore*0.3 + performanceScore*0.3)

	// Determine recommendation
	var recommendation string
	var riskLevel string
	var confidenceLevel float64

	if overallScore >= 0.9 {
		recommendation = "Excellent model performance. Ready for production deployment."
		riskLevel = "Low"
		confidenceLevel = 0.95
	} else if overallScore >= 0.8 {
		recommendation = "Good model performance. Consider minor improvements before production."
		riskLevel = "Low-Medium"
		confidenceLevel = 0.85
	} else if overallScore >= 0.7 {
		recommendation = "Acceptable model performance. Significant improvements recommended."
		riskLevel = "Medium"
		confidenceLevel = 0.75
	} else if overallScore >= 0.6 {
		recommendation = "Poor model performance. Major improvements required before deployment."
		riskLevel = "High"
		confidenceLevel = 0.65
	} else {
		recommendation = "Unacceptable model performance. Model needs complete retraining."
		riskLevel = "Very High"
		confidenceLevel = 0.50
	}

	return ValidationSummary{
		OverallScore:     overallScore,
		AccuracyScore:    accuracyScore,
		ReliabilityScore: reliabilityScore,
		PerformanceScore: performanceScore,
		Recommendation:   recommendation,
		RiskLevel:        riskLevel,
		ConfidenceLevel:  confidenceLevel,
	}
}

// generateModelComparison generates model comparison for single model
func (vs *ValidationService) generateModelComparison(cvResult *ValidationResult) []ModelComparison {
	strengths := []string{}
	weaknesses := []string{}

	// Analyze strengths and weaknesses
	if cvResult.OverallMetrics.MeanAccuracy > 0.85 {
		strengths = append(strengths, "High accuracy")
	} else if cvResult.OverallMetrics.MeanAccuracy < 0.7 {
		weaknesses = append(weaknesses, "Low accuracy")
	}

	if cvResult.OverallMetrics.StdAccuracy < 0.05 {
		strengths = append(strengths, "Consistent performance")
	} else if cvResult.OverallMetrics.StdAccuracy > 0.1 {
		weaknesses = append(weaknesses, "Inconsistent performance")
	}

	if cvResult.OverallMetrics.MeanF1Score > 0.8 {
		strengths = append(strengths, "Good F1 score")
	} else if cvResult.OverallMetrics.MeanF1Score < 0.6 {
		weaknesses = append(weaknesses, "Poor F1 score")
	}

	if cvResult.OverallMetrics.MeanAUC > 0.8 {
		strengths = append(strengths, "Good AUC")
	} else if cvResult.OverallMetrics.MeanAUC < 0.6 {
		weaknesses = append(weaknesses, "Poor AUC")
	}

	overallScore := (cvResult.OverallMetrics.MeanAccuracy + cvResult.OverallMetrics.MeanF1Score + cvResult.OverallMetrics.MeanAUC) / 3.0

	return []ModelComparison{
		{
			ModelName: cvResult.ModelName,
			Metrics: Metrics{
				Accuracy:  cvResult.OverallMetrics.MeanAccuracy,
				Precision: cvResult.OverallMetrics.MeanPrecision,
				Recall:    cvResult.OverallMetrics.MeanRecall,
				F1Score:   cvResult.OverallMetrics.MeanF1Score,
				AUC:       cvResult.OverallMetrics.MeanAUC,
				LogLoss:   cvResult.OverallMetrics.MeanLogLoss,
			},
			OverallScore: overallScore,
			Rank:         1,
			Strengths:    strengths,
			Weaknesses:   weaknesses,
			UseCase:      "General risk assessment",
		},
	}
}

// generateMultiModelComparison generates comparison for multiple models
func (vs *ValidationService) generateMultiModelComparison(results []*ValidationResult) []ModelComparison {
	comparisons := make([]ModelComparison, len(results))

	for i, result := range results {
		if result == nil {
			continue
		}

		overallScore := (result.OverallMetrics.MeanAccuracy + result.OverallMetrics.MeanF1Score + result.OverallMetrics.MeanAUC) / 3.0

		comparisons[i] = ModelComparison{
			ModelName: result.ModelName,
			Metrics: Metrics{
				Accuracy:  result.OverallMetrics.MeanAccuracy,
				Precision: result.OverallMetrics.MeanPrecision,
				Recall:    result.OverallMetrics.MeanRecall,
				F1Score:   result.OverallMetrics.MeanF1Score,
				AUC:       result.OverallMetrics.MeanAUC,
				LogLoss:   result.OverallMetrics.MeanLogLoss,
			},
			OverallScore: overallScore,
			Rank:         i + 1,
			Strengths:    []string{}, // Would be calculated based on comparison
			Weaknesses:   []string{}, // Would be calculated based on comparison
			UseCase:      "General risk assessment",
		}
	}

	// Sort by overall score
	sort.Slice(comparisons, func(i, j int) bool {
		return comparisons[i].OverallScore > comparisons[j].OverallScore
	})

	// Update ranks
	for i := range comparisons {
		comparisons[i].Rank = i + 1
	}

	return comparisons
}

// findBestModel finds the best performing model
func (vs *ValidationService) findBestModel(results []*ValidationResult) *ValidationResult {
	var best *ValidationResult
	bestScore := 0.0

	for _, result := range results {
		if result == nil {
			continue
		}

		score := (result.OverallMetrics.MeanAccuracy + result.OverallMetrics.MeanF1Score + result.OverallMetrics.MeanAUC) / 3.0
		if score > bestScore {
			bestScore = score
			best = result
		}
	}

	return best
}

// generateRecommendations generates actionable recommendations
func (vs *ValidationService) generateRecommendations(
	cvResult *ValidationResult,
	historicalSummary HistoricalDataSummary,
	summary ValidationSummary,
) []Recommendation {
	recommendations := []Recommendation{}

	// Accuracy recommendations
	if cvResult.OverallMetrics.MeanAccuracy < 0.8 {
		recommendations = append(recommendations, Recommendation{
			Type:        "Model Improvement",
			Priority:    "High",
			Title:       "Improve Model Accuracy",
			Description: "Model accuracy is below acceptable threshold. Consider feature engineering, hyperparameter tuning, or additional training data.",
			Impact:      "High",
			Effort:      "Medium",
			Timeline:    "2-4 weeks",
			Actions: []string{
				"Analyze misclassified samples",
				"Feature engineering and selection",
				"Hyperparameter optimization",
				"Additional training data collection",
			},
		})
	}

	// Consistency recommendations
	if cvResult.OverallMetrics.StdAccuracy > 0.1 {
		recommendations = append(recommendations, Recommendation{
			Type:        "Model Stability",
			Priority:    "Medium",
			Title:       "Improve Model Consistency",
			Description: "Model shows high variance across folds. Consider regularization, ensemble methods, or more robust training.",
			Impact:      "Medium",
			Effort:      "Medium",
			Timeline:    "1-2 weeks",
			Actions: []string{
				"Add regularization techniques",
				"Implement ensemble methods",
				"Cross-validation with more folds",
				"Stratified sampling",
			},
		})
	}

	// Data quality recommendations
	if historicalSummary.DataQuality.OverallQuality < 0.8 {
		recommendations = append(recommendations, Recommendation{
			Type:        "Data Quality",
			Priority:    "High",
			Title:       "Improve Data Quality",
			Description: "Historical data quality is below optimal levels. Focus on data collection and preprocessing improvements.",
			Impact:      "High",
			Effort:      "High",
			Timeline:    "4-6 weeks",
			Actions: []string{
				"Data collection process review",
				"Data validation and cleaning",
				"Missing data imputation",
				"Data quality monitoring",
			},
		})
	}

	// Performance recommendations
	if cvResult.ValidationTime > 5*time.Minute {
		recommendations = append(recommendations, Recommendation{
			Type:        "Performance",
			Priority:    "Low",
			Title:       "Optimize Model Performance",
			Description: "Model training and validation time is high. Consider optimization for production deployment.",
			Impact:      "Low",
			Effort:      "Medium",
			Timeline:    "1-2 weeks",
			Actions: []string{
				"Model optimization",
				"Feature selection",
				"Parallel processing",
				"Caching strategies",
			},
		})
	}

	return recommendations
}

// saveValidationReport saves the validation report to file
func (vs *ValidationService) saveValidationReport(report *ValidationReport, path string) error {
	// This would implement file saving logic
	// For now, just log the report
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	vs.logger.Info("Validation report generated",
		zap.String("path", path),
		zap.String("report", string(reportJSON)))

	return nil
}
