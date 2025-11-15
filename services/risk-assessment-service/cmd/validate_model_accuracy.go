package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/ml/data"
	"kyb-platform/services/risk-assessment-service/internal/ml/ensemble"
	"kyb-platform/services/risk-assessment-service/internal/ml/training"
	"kyb-platform/services/risk-assessment-service/internal/ml/validation"
)

func mainValidateModelAccuracy() {
	// Command line flags
	var (
		numBusinesses      = flag.Int("businesses", 1000, "Number of businesses for validation")
		sequenceLength     = flag.Int("sequence-length", 24, "Time series sequence length")
		cvFolds            = flag.Int("cv-folds", 5, "Cross-validation folds")
		validationHorizons = flag.String("horizons", "3,6,12", "Comma-separated validation horizons in months")
		optimizationMetric = flag.String("metric", "accuracy", "Optimization metric: accuracy, mae, rmse, f1")
		searchStrategy     = flag.String("strategy", "random", "Search strategy: random, grid, bayesian")
		maxTrials          = flag.Int("max-trials", 50, "Maximum number of hyperparameter trials")
		patience           = flag.Int("patience", 10, "Early stopping patience")
		enableTuning       = flag.Bool("tune", false, "Enable hyperparameter tuning")
		verbose            = flag.Bool("verbose", false, "Verbose output")
		outputFile         = flag.String("output", "", "Output file for results (JSON format)")
	)
	flag.Parse()

	// Initialize logger
	var logger *zap.Logger
	var err error
	if *verbose {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("🚀 Starting Model Accuracy Validation",
		zap.Int("num_businesses", *numBusinesses),
		zap.Int("sequence_length", *sequenceLength),
		zap.Int("cv_folds", *cvFolds),
		zap.String("horizons", *validationHorizons),
		zap.String("optimization_metric", *optimizationMetric),
		zap.Bool("enable_tuning", *enableTuning))

	// Parse validation horizons
	horizons, err := parseHorizons(*validationHorizons)
	if err != nil {
		logger.Fatal("Failed to parse validation horizons", zap.Error(err))
	}

	// Initialize components
	syntheticGenerator := data.NewSyntheticDataGenerator()
	historyCollector := data.NewHistoryCollector(nil, nil, logger)
	hybridBlender := data.NewHybridBlender(syntheticGenerator, historyCollector, logger)
	ensembleRouter := ensemble.NewEnsembleRouter(nil, nil, logger)
	validator := validation.NewLSTMValidator(syntheticGenerator, historyCollector, hybridBlender, ensembleRouter, logger)

	// Create validation config
	validationConfig := validation.ValidationConfig{
		NumBusinesses:              *numBusinesses,
		SequenceLength:             *sequenceLength,
		ValidationHorizons:         horizons,
		CrossValidationFolds:       *cvFolds,
		ValidationSplit:            0.8,
		TestSplit:                  0.2,
		TargetAccuracy:             0.90,
		MaxIterations:              100,
		EnableCalibration:          true,
		EnableEnsemble:             false,
		EnableHyperparameterTuning: *enableTuning,
	}

	// Run baseline validation
	logger.Info("Running baseline model validation...")
	baselineResult, err := validator.ValidateModel(context.Background(), validationConfig)
	if err != nil {
		logger.Fatal("Baseline validation failed", zap.Error(err))
	}

	// Print baseline results
	printValidationResults("Baseline", baselineResult, logger)

	// Check if baseline meets 90% accuracy target
	meetsTarget := baselineResult.OverallAccuracy >= 0.90
	logger.Info("Baseline accuracy assessment",
		zap.Float64("accuracy", baselineResult.OverallAccuracy),
		zap.Bool("meets_90_percent_target", meetsTarget))

	// Run hyperparameter tuning if enabled
	var tuningResult *training.TuningResult
	if *enableTuning {
		logger.Info("Starting hyperparameter tuning...")

		// Create hyperparameter tuner
		tuner := training.NewHyperparameterTuner(logger)

		// Create tuning config
		tuningConfig := training.TuningConfig{
			MaxTrials:            *maxTrials,
			Patience:             *patience,
			OptimizationMetric:   *optimizationMetric,
			SearchStrategy:       *searchStrategy,
			EarlyStopping:        true,
			CrossValidationFolds: *cvFolds,
			ValidationHorizons:   horizons,
			NumBusinesses:        *numBusinesses,
			SequenceLength:       *sequenceLength,
			RandomSeed:           time.Now().UnixNano(),
			HyperparameterSpace:  training.GetDefaultHyperparameterSpace(),
		}

		// Run tuning
		tuningResult, err = tuner.TuneHyperparameters(context.Background(), tuningConfig)
		if err != nil {
			logger.Fatal("Hyperparameter tuning failed", zap.Error(err))
		}

		// Print tuning results
		printTuningResults(tuningResult, logger)

		// Check if tuned model meets 90% accuracy target
		meetsTarget = tuningResult.BestTrial.Accuracy >= 0.90
		logger.Info("Tuned model accuracy assessment",
			zap.Float64("accuracy", tuningResult.BestTrial.Accuracy),
			zap.Bool("meets_90_percent_target", meetsTarget))
	}

	// Generate final report
	generateFinalReport(baselineResult, tuningResult, meetsTarget, logger)

	// Save results to file if requested
	if *outputFile != "" {
		saveValidationResultsToFile(baselineResult, tuningResult, *outputFile, logger)
	}

	// Exit with appropriate code
	if meetsTarget {
		logger.Info("✅ Model accuracy target (90%) achieved!")
		os.Exit(0)
	} else {
		logger.Warn("❌ Model accuracy target (90%) not achieved")
		os.Exit(1)
	}
}

// parseHorizons parses comma-separated horizon values
func parseHorizons(horizonsStr string) ([]int, error) {
	var horizons []int
	var current int

	for i, char := range horizonsStr {
		if char == ',' {
			horizons = append(horizons, current)
			current = 0
		} else if char >= '0' && char <= '9' {
			current = current*10 + int(char-'0')
		} else if char != ' ' {
			return nil, fmt.Errorf("invalid character in horizons: %c", char)
		}
	}

	// Add the last horizon
	if current > 0 {
		horizons = append(horizons, current)
	}

	return horizons, nil
}

// printValidationResults prints validation results
func printValidationResults(title string, result *validation.LSTMValidationResult, logger *zap.Logger) {
	fmt.Printf("\n%s Results\n", title)
	fmt.Println("=" + string(make([]byte, len(title)+8)))
	fmt.Printf("Overall Accuracy: %.4f (%.2f%%)\n", result.OverallAccuracy, result.OverallAccuracy*100)
	fmt.Printf("Overall MAE:      %.4f\n", result.OverallMAE)
	fmt.Printf("Overall RMSE:     %.4f\n", result.OverallRMSE)
	fmt.Printf("Total Samples:    %d\n", result.TotalSamples)

	fmt.Println("\nHorizon Results:")
	for horizon, horizonResult := range result.HorizonResults {
		fmt.Printf("  %d months: Accuracy=%.4f, MAE=%.4f, RMSE=%.4f, R²=%.4f\n",
			horizon, horizonResult.Accuracy, horizonResult.MAE, horizonResult.RMSE, horizonResult.R2Score)
	}

	fmt.Println("\nConfidence Analysis:")
	fmt.Printf("  High Confidence Accuracy: %.4f\n", result.ConfidenceAnalysis.HighConfidenceAccuracy)
	fmt.Printf("  Low Confidence Accuracy:  %.4f\n", result.ConfidenceAnalysis.LowConfidenceAccuracy)
	fmt.Printf("  Calibration Error:        %.4f\n", result.ConfidenceAnalysis.CalibrationError)

	fmt.Println()
}

// printTuningResults prints hyperparameter tuning results
func printTuningResults(result *training.TuningResult, logger *zap.Logger) {
	fmt.Println("\nHyperparameter Tuning Results")
	fmt.Println("=============================")
	fmt.Printf("Best Score:        %.4f\n", result.BestScore)
	fmt.Printf("Best Accuracy:     %.4f (%.2f%%)\n", result.BestTrial.Accuracy, result.BestTrial.Accuracy*100)
	fmt.Printf("Best MAE:          %.4f\n", result.BestTrial.MAE)
	fmt.Printf("Best RMSE:         %.4f\n", result.BestTrial.RMSE)
	fmt.Printf("Total Trials:      %d\n", result.TotalTrials)
	fmt.Printf("Completed Trials:  %d\n", result.CompletedTrials)
	fmt.Printf("Failed Trials:     %d\n", result.FailedTrials)
	fmt.Printf("Optimization Time: %v\n", result.OptimizationTime)
	fmt.Printf("Improvement:       %.2f%%\n", result.ImprovementOverBaseline)

	fmt.Println("\nBest Parameters:")
	for name, value := range result.BestParameters {
		fmt.Printf("  %s: %v\n", name, value)
	}

	fmt.Println("\nConvergence Analysis:")
	fmt.Printf("  Converged:        %t\n", result.ConvergenceAnalysis.Converged)
	fmt.Printf("  Convergence Point: %d\n", result.ConvergenceAnalysis.ConvergencePoint)
	fmt.Printf("  Final Improvement: %.4f\n", result.ConvergenceAnalysis.FinalImprovement)
	fmt.Printf("  Stability Score:   %.4f\n", result.ConvergenceAnalysis.StabilityScore)

	fmt.Println()
}

// generateFinalReport generates a final accuracy report
func generateFinalReport(baseline *validation.LSTMValidationResult, tuning *training.TuningResult, meetsTarget bool, logger *zap.Logger) {
	fmt.Println("\n" + string(make([]byte, 60, 60)))
	fmt.Println("FINAL MODEL ACCURACY REPORT")
	fmt.Println(string(make([]byte, 60, 60)))

	fmt.Printf("Validation Date: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("Target Accuracy: 90.00%%\n")

	// Baseline results
	fmt.Printf("\nBaseline Model:\n")
	fmt.Printf("  Accuracy: %.4f (%.2f%%)\n", baseline.OverallAccuracy, baseline.OverallAccuracy*100)
	fmt.Printf("  MAE:      %.4f\n", baseline.OverallMAE)
	fmt.Printf("  RMSE:     %.4f\n", baseline.OverallRMSE)

	// Tuned results (if available)
	if tuning != nil {
		fmt.Printf("\nTuned Model:\n")
		fmt.Printf("  Accuracy: %.4f (%.2f%%)\n", tuning.BestTrial.Accuracy, tuning.BestTrial.Accuracy*100)
		fmt.Printf("  MAE:      %.4f\n", tuning.BestTrial.MAE)
		fmt.Printf("  RMSE:     %.4f\n", tuning.BestTrial.RMSE)
		fmt.Printf("  Improvement: %.2f%%\n", tuning.ImprovementOverBaseline)
	}

	// Target assessment
	fmt.Printf("\nTarget Assessment:\n")
	if meetsTarget {
		fmt.Printf("  ✅ TARGET ACHIEVED: Model accuracy meets 90%% requirement\n")
	} else {
		fmt.Printf("  ❌ TARGET NOT MET: Model accuracy below 90%% requirement\n")
	}

	// Recommendations
	fmt.Printf("\nRecommendations:\n")
	if meetsTarget {
		fmt.Printf("  • Model is ready for production deployment\n")
		fmt.Printf("  • Continue monitoring accuracy in production\n")
		fmt.Printf("  • Consider A/B testing with current model\n")
	} else {
		fmt.Printf("  • Increase training data size\n")
		fmt.Printf("  • Try different model architectures\n")
		fmt.Printf("  • Consider ensemble methods\n")
		fmt.Printf("  • Review feature engineering\n")
		if tuning == nil {
			fmt.Printf("  • Run hyperparameter tuning\n")
		}
	}

	fmt.Println(string(make([]byte, 60, 60)))
}

// saveValidationResultsToFile saves results to a JSON file
func saveValidationResultsToFile(baseline *validation.LSTMValidationResult, tuning *training.TuningResult, filename string, logger *zap.Logger) {
	logger.Info("Saving results to file", zap.String("filename", filename))

	// In a real implementation, this would save the results to a JSON file
	// For now, we'll just log the action
	logger.Info("Results would be saved to JSON file", zap.String("filename", filename))
}
