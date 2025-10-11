package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/ml/data"
	"kyb-platform/services/risk-assessment-service/internal/ml/ensemble"
	mlmodels "kyb-platform/services/risk-assessment-service/internal/ml/models"
	"kyb-platform/services/risk-assessment-service/internal/ml/validation"
)

// ValidationConfig holds configuration for model validation
type ValidationConfig struct {
	NumBusinesses        int     `json:"num_businesses"`
	SequenceLength       int     `json:"sequence_length"`
	ValidationHorizons   []int   `json:"validation_horizons"`
	CrossValidationFolds int     `json:"cross_validation_folds"`
	TestDataRatio        float64 `json:"test_data_ratio"`
	RandomSeed           int64   `json:"random_seed"`
	OutputFile           string  `json:"output_file"`
	ModelPaths           struct {
		XGBoost string `json:"xgboost"`
		LSTM    string `json:"lstm"`
	} `json:"model_paths"`
}

// ValidationReport generates a comprehensive validation report
type ValidationReport struct {
	Summary            ValidationSummary                `json:"summary"`
	DetailedResults    *validation.LSTMValidationResult `json:"detailed_results"`
	Recommendations    []string                         `json:"recommendations"`
	PerformanceTargets PerformanceTargets               `json:"performance_targets"`
	GeneratedAt        time.Time                        `json:"generated_at"`
}

// ValidationSummary provides a high-level summary of validation results
type ValidationSummary struct {
	OverallAccuracy      float64 `json:"overall_accuracy"`
	OverallMAE           float64 `json:"overall_mae"`
	OverallRMSE          float64 `json:"overall_rmse"`
	TargetAccuracyMet    bool    `json:"target_accuracy_met"`
	BestHorizon          int     `json:"best_horizon"`
	BestHorizonAccuracy  float64 `json:"best_horizon_accuracy"`
	WorstHorizon         int     `json:"worst_horizon"`
	WorstHorizonAccuracy float64 `json:"worst_horizon_accuracy"`
	ModelComparison      struct {
		BestModel6Months  string  `json:"best_model_6_months"`
		BestModel12Months string  `json:"best_model_12_months"`
		LSTMAdvantage     float64 `json:"lstm_advantage_12_months"`
	} `json:"model_comparison"`
}

// PerformanceTargets defines the performance targets
type PerformanceTargets struct {
	Accuracy6Months  float64 `json:"accuracy_6_months"`
	Accuracy9Months  float64 `json:"accuracy_9_months"`
	Accuracy12Months float64 `json:"accuracy_12_months"`
	MAEThreshold     float64 `json:"mae_threshold"`
	RMSEThreshold    float64 `json:"rmse_threshold"`
}

func validateMain() {
	// Load configuration
	config := loadValidationConfig()

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}
	defer logger.Sync()

	logger.Info("Starting LSTM model validation",
		zap.Int("num_businesses", config.NumBusinesses),
		zap.Ints("horizons", config.ValidationHorizons),
		zap.Int("cv_folds", config.CrossValidationFolds))

	// Initialize models and components
	components, err := initializeComponents(config, logger)
	if err != nil {
		logger.Fatal("Failed to initialize components", zap.Error(err))
	}

	// Create validator
	validator := validation.NewLSTMValidator(
		components.syntheticGenerator,
		components.historyCollector,
		components.hybridBlender,
		components.ensembleRouter,
		logger,
	)

	// Perform validation
	ctx := context.Background()
	validationConfig := validation.ValidationConfig{
		NumBusinesses:        config.NumBusinesses,
		SequenceLength:       config.SequenceLength,
		ValidationHorizons:   config.ValidationHorizons,
		CrossValidationFolds: config.CrossValidationFolds,
		TestDataRatio:        config.TestDataRatio,
		RandomSeed:           config.RandomSeed,
	}

	start := time.Now()
	results, err := validator.ValidateModel(ctx, validationConfig)
	if err != nil {
		logger.Fatal("Validation failed", zap.Error(err))
	}
	duration := time.Since(start)

	logger.Info("Validation completed",
		zap.Duration("duration", duration),
		zap.Float64("overall_accuracy", results.OverallAccuracy),
		zap.Float64("overall_mae", results.OverallMAE),
		zap.Float64("overall_rmse", results.OverallRMSE))

	// Generate report
	report := generateValidationReport(results, config)

	// Print summary
	printValidationSummary(report)

	// Save detailed results
	if err := saveValidationResults(report, config.OutputFile); err != nil {
		logger.Error("Failed to save validation results", zap.Error(err))
	} else {
		logger.Info("Validation results saved", zap.String("file", config.OutputFile))
	}

	// Check if targets are met
	if report.Summary.TargetAccuracyMet {
		logger.Info("✅ All accuracy targets met!")
	} else {
		logger.Warn("❌ Some accuracy targets not met")
	}
}

// ComponentContainer holds all initialized components
type ComponentContainer struct {
	syntheticGenerator *data.SyntheticDataGenerator
	historyCollector   *data.HistoryCollector
	hybridBlender      *data.HybridBlender
	ensembleRouter     *ensemble.EnsembleRouter
}

// loadValidationConfig loads validation configuration
func loadValidationConfig() ValidationConfig {
	// Default configuration
	config := ValidationConfig{
		NumBusinesses:        1000,
		SequenceLength:       12,
		ValidationHorizons:   []int{6, 9, 12},
		CrossValidationFolds: 5,
		TestDataRatio:        0.2,
		RandomSeed:           time.Now().UnixNano(),
		OutputFile:           "validation_results.json",
		ModelPaths: struct {
			XGBoost string `json:"xgboost"`
			LSTM    string `json:"lstm"`
		}{
			XGBoost: "./models/xgb_model.json",
			LSTM:    "./models/risk_lstm_v1.onnx",
		},
	}

	// Try to load from config file
	if configFile := "validation_config.json"; fileExists(configFile) {
		data, err := os.ReadFile(configFile)
		if err != nil {
			log.Printf("Warning: Failed to read config file: %v", err)
		} else {
			if err := json.Unmarshal(data, &config); err != nil {
				log.Printf("Warning: Failed to parse config file: %v", err)
			}
		}
	}

	return config
}

// initializeComponents initializes all required components
func initializeComponents(config ValidationConfig, logger *zap.Logger) (*ComponentContainer, error) {
	// Initialize synthetic data generator
	syntheticGenerator := data.NewSyntheticDataGenerator()
	syntheticGenerator.SetLogger(logger)

	// Initialize history collector (mock for validation)
	historyCollector := data.NewHistoryCollector(nil, nil, logger)

	// Initialize hybrid blender
	hybridBlender := data.NewHybridBlender(syntheticGenerator, historyCollector, logger)

	// Initialize models
	xgbModel := mlmodels.NewXGBoostModel("validation_xgb", "1.0.0")
	if err := xgbModel.LoadModel(context.Background(), config.ModelPaths.XGBoost); err != nil {
		logger.Warn("Failed to load XGBoost model, using mock", zap.Error(err))
	}

	lstmModel := mlmodels.NewLSTMModel("validation_lstm", "1.0.0", logger)
	if err := lstmModel.LoadModel(context.Background(), config.ModelPaths.LSTM); err != nil {
		logger.Warn("Failed to load LSTM model, using mock", zap.Error(err))
	}

	// Initialize ensemble router
	ensembleRouter := ensemble.NewEnsembleRouter(xgbModel, lstmModel, logger)

	return &ComponentContainer{
		syntheticGenerator: syntheticGenerator,
		historyCollector:   historyCollector,
		hybridBlender:      hybridBlender,
		ensembleRouter:     ensembleRouter,
	}, nil
}

// generateValidationReport generates a comprehensive validation report
func generateValidationReport(results *validation.LSTMValidationResult, config ValidationConfig) ValidationReport {
	// Define performance targets
	targets := PerformanceTargets{
		Accuracy6Months:  0.88,
		Accuracy9Months:  0.86,
		Accuracy12Months: 0.85,
		MAEThreshold:     0.15,
		RMSEThreshold:    0.20,
	}

	// Calculate summary
	summary := ValidationSummary{
		OverallAccuracy:   results.OverallAccuracy,
		OverallMAE:        results.OverallMAE,
		OverallRMSE:       results.OverallRMSE,
		TargetAccuracyMet: true, // Will be updated below
	}

	// Find best and worst horizons
	bestAccuracy := 0.0
	worstAccuracy := 1.0
	for horizon, result := range results.HorizonResults {
		if result.Accuracy > bestAccuracy {
			bestAccuracy = result.Accuracy
			summary.BestHorizon = horizon
			summary.BestHorizonAccuracy = result.Accuracy
		}
		if result.Accuracy < worstAccuracy {
			worstAccuracy = result.Accuracy
			summary.WorstHorizon = horizon
			summary.WorstHorizonAccuracy = result.Accuracy
		}

		// Check if targets are met
		switch horizon {
		case 6:
			if result.Accuracy < targets.Accuracy6Months {
				summary.TargetAccuracyMet = false
			}
		case 9:
			if result.Accuracy < targets.Accuracy9Months {
				summary.TargetAccuracyMet = false
			}
		case 12:
			if result.Accuracy < targets.Accuracy12Months {
				summary.TargetAccuracyMet = false
			}
		}
	}

	// Analyze model comparison
	if results.ModelComparison.XGBoostResults != nil && results.ModelComparison.LSTMResults != nil {
		if xgb6, exists := results.ModelComparison.XGBoostResults[6]; exists {
			summary.ModelComparison.BestModel6Months = "xgboost"
			if lstm6, exists := results.ModelComparison.LSTMResults[6]; exists && lstm6.Accuracy > xgb6.Accuracy {
				summary.ModelComparison.BestModel6Months = "lstm"
			}
		}

		if xgb12, exists := results.ModelComparison.XGBoostResults[12]; exists {
			summary.ModelComparison.BestModel12Months = "xgboost"
			if lstm12, exists := results.ModelComparison.LSTMResults[12]; exists {
				if lstm12.Accuracy > xgb12.Accuracy {
					summary.ModelComparison.BestModel12Months = "lstm"
					summary.ModelComparison.LSTMAdvantage = lstm12.Accuracy - xgb12.Accuracy
				} else {
					summary.ModelComparison.LSTMAdvantage = lstm12.Accuracy - xgb12.Accuracy
				}
			}
		}
	}

	// Generate recommendations
	recommendations := generateRecommendations(results, summary, targets)

	return ValidationReport{
		Summary:            summary,
		DetailedResults:    results,
		Recommendations:    recommendations,
		PerformanceTargets: targets,
		GeneratedAt:        time.Now(),
	}
}

// generateRecommendations generates recommendations based on validation results
func generateRecommendations(results *validation.LSTMValidationResult, summary ValidationSummary, targets PerformanceTargets) []string {
	var recommendations []string

	// Accuracy recommendations
	if !summary.TargetAccuracyMet {
		recommendations = append(recommendations, "❌ Accuracy targets not met - consider retraining with more data or hyperparameter tuning")
	}

	if summary.OverallAccuracy < 0.8 {
		recommendations = append(recommendations, "⚠️ Overall accuracy below 80% - model may need significant improvement")
	}

	// Horizon-specific recommendations
	for horizon, result := range results.HorizonResults {
		if result.Accuracy < 0.8 {
			recommendations = append(recommendations, fmt.Sprintf("⚠️ %d-month horizon accuracy (%.2f%%) below 80%% - consider model adjustments", horizon, result.Accuracy*100))
		}
	}

	// Model comparison recommendations
	if summary.ModelComparison.LSTMAdvantage < 0.02 {
		recommendations = append(recommendations, "💡 LSTM shows minimal advantage over XGBoost for 12-month predictions - consider ensemble approach")
	}

	// MAE/RMSE recommendations
	if summary.OverallMAE > targets.MAEThreshold {
		recommendations = append(recommendations, fmt.Sprintf("📊 MAE (%.3f) exceeds threshold (%.3f) - consider feature engineering", summary.OverallMAE, targets.MAEThreshold))
	}

	if summary.OverallRMSE > targets.RMSEThreshold {
		recommendations = append(recommendations, fmt.Sprintf("📊 RMSE (%.3f) exceeds threshold (%.3f) - consider model regularization", summary.OverallRMSE, targets.RMSEThreshold))
	}

	// Confidence analysis recommendations
	if results.ConfidenceAnalysis.CalibrationError > 0.1 {
		recommendations = append(recommendations, "🎯 High calibration error - model confidence may not be well-calibrated")
	}

	// Positive recommendations
	if summary.TargetAccuracyMet {
		recommendations = append(recommendations, "✅ All accuracy targets met - model ready for production")
	}

	if summary.ModelComparison.LSTMAdvantage > 0.05 {
		recommendations = append(recommendations, "🚀 LSTM shows significant advantage for long-term predictions - excellent for 12-month forecasts")
	}

	return recommendations
}

// printValidationSummary prints a formatted validation summary
func printValidationSummary(report ValidationReport) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("LSTM MODEL VALIDATION REPORT")
	fmt.Println(strings.Repeat("=", 80))

	// Overall results
	fmt.Println("\n📊 OVERALL RESULTS")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Overall Accuracy: %.2f%%\n", report.Summary.OverallAccuracy*100)
	fmt.Printf("Overall MAE:      %.3f\n", report.Summary.OverallMAE)
	fmt.Printf("Overall RMSE:     %.3f\n", report.Summary.OverallRMSE)
	fmt.Printf("Targets Met:      %s\n", getStatusEmoji(report.Summary.TargetAccuracyMet))

	// Horizon results
	fmt.Println("\n🎯 HORIZON-SPECIFIC RESULTS")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("%-8s %-10s %-8s %-8s %-8s %-8s\n", "Horizon", "Accuracy", "MAE", "RMSE", "R²", "Samples")
	fmt.Println(strings.Repeat("-", 40))
	for horizon, result := range report.DetailedResults.HorizonResults {
		target := getTargetForHorizon(horizon, report.PerformanceTargets)
		status := getStatusEmoji(result.Accuracy >= target)
		fmt.Printf("%-8d %-9.2f%% %-8.3f %-8.3f %-8.3f %-8d %s\n",
			horizon, result.Accuracy*100, result.MAE, result.RMSE, result.R2Score, result.SampleCount, status)
	}

	// Model comparison
	fmt.Println("\n🔄 MODEL COMPARISON")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Best 6-month model:  %s\n", report.Summary.ModelComparison.BestModel6Months)
	fmt.Printf("Best 12-month model: %s\n", report.Summary.ModelComparison.BestModel12Months)
	fmt.Printf("LSTM advantage (12m): %.2f%%\n", report.Summary.ModelComparison.LSTMAdvantage*100)

	// Confidence analysis
	fmt.Println("\n🎯 CONFIDENCE ANALYSIS")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("High confidence accuracy: %.2f%%\n", report.DetailedResults.ConfidenceAnalysis.HighConfidenceAccuracy*100)
	fmt.Printf("Low confidence accuracy:  %.2f%%\n", report.DetailedResults.ConfidenceAnalysis.LowConfidenceAccuracy*100)
	fmt.Printf("Calibration error:        %.3f\n", report.DetailedResults.ConfidenceAnalysis.CalibrationError)

	// Cross-validation results
	fmt.Println("\n📈 CROSS-VALIDATION RESULTS")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("%-6s %-10s %-8s %-8s\n", "Fold", "Accuracy", "MAE", "RMSE")
	fmt.Println(strings.Repeat("-", 40))
	for _, cvResult := range report.DetailedResults.CrossValidationResults {
		fmt.Printf("%-6d %-9.2f%% %-8.3f %-8.3f\n",
			cvResult.Fold, cvResult.OverallAccuracy*100, cvResult.OverallMAE, cvResult.OverallRMSE)
	}

	// Recommendations
	fmt.Println("\n💡 RECOMMENDATIONS")
	fmt.Println(strings.Repeat("-", 40))
	for i, rec := range report.Recommendations {
		fmt.Printf("%d. %s\n", i+1, rec)
	}

	// Performance targets
	fmt.Println("\n🎯 PERFORMANCE TARGETS")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("6-month accuracy:  ≥%.0f%% (target)\n", report.PerformanceTargets.Accuracy6Months*100)
	fmt.Printf("9-month accuracy:  ≥%.0f%% (target)\n", report.PerformanceTargets.Accuracy9Months*100)
	fmt.Printf("12-month accuracy: ≥%.0f%% (target)\n", report.PerformanceTargets.Accuracy12Months*100)
	fmt.Printf("MAE threshold:     ≤%.3f\n", report.PerformanceTargets.MAEThreshold)
	fmt.Printf("RMSE threshold:    ≤%.3f\n", report.PerformanceTargets.RMSEThreshold)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("Report generated at: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 80))
}

// saveValidationResults saves validation results to file
func saveValidationResults(report ValidationReport, filename string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// Helper functions
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func getStatusEmoji(condition bool) string {
	if condition {
		return "✅"
	}
	return "❌"
}

func getTargetForHorizon(horizon int, targets PerformanceTargets) float64 {
	switch horizon {
	case 6:
		return targets.Accuracy6Months
	case 9:
		return targets.Accuracy9Months
	case 12:
		return targets.Accuracy12Months
	default:
		return 0.8 // Default target
	}
}
