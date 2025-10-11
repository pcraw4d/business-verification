package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"kyb-platform/services/risk-assessment-service/internal/ml/models"
	"kyb-platform/services/risk-assessment-service/internal/ml/validation"

	"go.uber.org/zap"
)

func main() {
	// Command line flags
	var (
		kFolds          = flag.Int("k", 5, "Number of folds for cross-validation")
		totalSamples    = flag.Int("samples", 1000, "Total number of samples to generate")
		timeRange       = flag.Duration("time-range", 365*24*time.Hour, "Time range for historical data")
		confidenceLevel = flag.Float64("confidence", 0.95, "Confidence level for intervals")
		outputFile      = flag.String("output", "", "Output file for validation report (JSON)")
		verbose         = flag.Bool("verbose", false, "Enable verbose logging")
		parallel        = flag.Bool("parallel", false, "Enable parallel fold processing")
		seed            = flag.Int64("seed", time.Now().UnixNano(), "Random seed for reproducibility")
	)
	flag.Parse()

	// Setup logging
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

	logger.Info("Starting ML model validation",
		zap.Int("k_folds", *kFolds),
		zap.Int("total_samples", *totalSamples),
		zap.Duration("time_range", *timeRange),
		zap.Float64("confidence_level", *confidenceLevel),
		zap.Int64("seed", *seed))

	// Create validation service
	validationService := validation.NewValidationService(logger)

	// Create XGBoost model for validation
	model := models.NewXGBoostModel(logger)

	// Configure validation
	config := validation.ValidationConfig{
		CrossValidation: validation.CrossValidationConfig{
			KFolds:          *kFolds,
			ConfidenceLevel: *confidenceLevel,
			RandomSeed:      *seed,
			ParallelFolds:   *parallel,
			MaxConcurrency:  4,
		},
		DataGeneration: validation.DataGenerationConfig{
			TotalSamples:     *totalSamples,
			TimeRange:        *timeRange,
			RiskCategories:   getDefaultRiskCategories(),
			IndustryWeights:  getDefaultIndustryWeights(),
			GeographicBias:   getDefaultGeographicBias(),
			SeasonalPatterns: true,
			TrendStrength:    0.02,
			NoiseLevel:       0.05,
		},
		OutputFormat: "json",
		SaveResults:  *outputFile != "",
		ResultsPath:  *outputFile,
	}

	// Run validation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	report, err := validationService.ValidateModel(ctx, model, config)
	if err != nil {
		logger.Fatal("Model validation failed", zap.Error(err))
	}

	// Display results
	displayResults(report, logger)

	// Save results if requested
	if *outputFile != "" {
		if err := saveReport(report, *outputFile); err != nil {
			logger.Error("Failed to save report", zap.Error(err))
		} else {
			logger.Info("Report saved successfully", zap.String("file", *outputFile))
		}
	}

	logger.Info("Model validation completed successfully")
}

// displayResults displays validation results in a formatted way
func displayResults(report *validation.ValidationReport, logger *zap.Logger) {
	fmt.Println("\n" + "="*80)
	fmt.Println("ML MODEL VALIDATION REPORT")
	fmt.Println("=" * 80)

	// Summary
	fmt.Printf("\n📊 VALIDATION SUMMARY\n")
	fmt.Printf("Overall Score: %.3f\n", report.Summary.OverallScore)
	fmt.Printf("Accuracy Score: %.3f\n", report.Summary.AccuracyScore)
	fmt.Printf("Reliability Score: %.3f\n", report.Summary.ReliabilityScore)
	fmt.Printf("Performance Score: %.3f\n", report.Summary.PerformanceScore)
	fmt.Printf("Risk Level: %s\n", report.Summary.RiskLevel)
	fmt.Printf("Confidence Level: %.1f%%\n", report.Summary.ConfidenceLevel*100)
	fmt.Printf("Recommendation: %s\n", report.Summary.Recommendation)

	// Cross-validation results
	if report.CrossValidation != nil {
		fmt.Printf("\n🔬 CROSS-VALIDATION RESULTS\n")
		fmt.Printf("Model: %s\n", report.CrossValidation.ModelName)
		fmt.Printf("K-Folds: %d\n", report.CrossValidation.K)
		fmt.Printf("Total Samples: %d\n", report.CrossValidation.TotalSamples)
		fmt.Printf("Validation Time: %v\n", report.CrossValidation.ValidationTime)

		fmt.Printf("\n📈 PERFORMANCE METRICS\n")
		fmt.Printf("Mean Accuracy: %.3f ± %.3f\n",
			report.CrossValidation.OverallMetrics.MeanAccuracy,
			report.CrossValidation.OverallMetrics.StdAccuracy)
		fmt.Printf("Mean Precision: %.3f ± %.3f\n",
			report.CrossValidation.OverallMetrics.MeanPrecision,
			report.CrossValidation.OverallMetrics.StdPrecision)
		fmt.Printf("Mean Recall: %.3f ± %.3f\n",
			report.CrossValidation.OverallMetrics.MeanRecall,
			report.CrossValidation.OverallMetrics.StdRecall)
		fmt.Printf("Mean F1-Score: %.3f ± %.3f\n",
			report.CrossValidation.OverallMetrics.MeanF1Score,
			report.CrossValidation.OverallMetrics.StdF1Score)
		fmt.Printf("Mean AUC: %.3f ± %.3f\n",
			report.CrossValidation.OverallMetrics.MeanAUC,
			report.CrossValidation.OverallMetrics.StdAUC)
		fmt.Printf("Mean Log Loss: %.3f ± %.3f\n",
			report.CrossValidation.OverallMetrics.MeanLogLoss,
			report.CrossValidation.OverallMetrics.StdLogLoss)

		// Confidence intervals
		fmt.Printf("\n📊 CONFIDENCE INTERVALS (%.0f%%)\n",
			report.CrossValidation.ConfidenceInterval.Confidence*100)
		fmt.Printf("Accuracy: [%.3f, %.3f]\n",
			report.CrossValidation.ConfidenceInterval.Accuracy.Lower,
			report.CrossValidation.ConfidenceInterval.Accuracy.Upper)
		fmt.Printf("Precision: [%.3f, %.3f]\n",
			report.CrossValidation.ConfidenceInterval.Precision.Lower,
			report.CrossValidation.ConfidenceInterval.Precision.Upper)
		fmt.Printf("Recall: [%.3f, %.3f]\n",
			report.CrossValidation.ConfidenceInterval.Recall.Lower,
			report.CrossValidation.ConfidenceInterval.Recall.Upper)
		fmt.Printf("F1-Score: [%.3f, %.3f]\n",
			report.CrossValidation.ConfidenceInterval.F1Score.Lower,
			report.CrossValidation.ConfidenceInterval.F1Score.Upper)
		fmt.Printf("AUC: [%.3f, %.3f]\n",
			report.CrossValidation.ConfidenceInterval.AUC.Lower,
			report.CrossValidation.ConfidenceInterval.AUC.Upper)

		// Fold details
		fmt.Printf("\n📋 FOLD DETAILS\n")
		for i, fold := range report.CrossValidation.FoldResults {
			fmt.Printf("Fold %d: Accuracy=%.3f, F1=%.3f, AUC=%.3f, Time=%v\n",
				fold.FoldIndex, fold.Metrics.Accuracy, fold.Metrics.F1Score,
				fold.Metrics.AUC, fold.TrainingTime+fold.InferenceTime)
		}
	}

	// Historical data summary
	fmt.Printf("\n📚 HISTORICAL DATA SUMMARY\n")
	fmt.Printf("Total Samples: %d\n", report.HistoricalData.TotalSamples)
	fmt.Printf("Time Range: %v\n", report.HistoricalData.TimeRange)
	fmt.Printf("Data Quality: %.3f\n", report.HistoricalData.DataQuality.OverallQuality)

	fmt.Printf("\nIndustries:\n")
	for industry, count := range report.HistoricalData.Industries {
		fmt.Printf("  %s: %d samples\n", industry, count)
	}

	fmt.Printf("\nCountries:\n")
	for country, count := range report.HistoricalData.Countries {
		fmt.Printf("  %s: %d samples\n", country, count)
	}

	fmt.Printf("\nBusiness Sizes:\n")
	for size, count := range report.HistoricalData.BusinessSizes {
		fmt.Printf("  %s: %d samples\n", size, count)
	}

	fmt.Printf("\nRisk Distribution:\n")
	for risk, count := range report.HistoricalData.RiskDistribution {
		fmt.Printf("  %s: %d samples\n", risk, count)
	}

	// Model comparison
	if len(report.ModelComparison) > 0 {
		fmt.Printf("\n🏆 MODEL COMPARISON\n")
		for i, model := range report.ModelComparison {
			fmt.Printf("Rank %d: %s (Score: %.3f)\n",
				model.Rank, model.ModelName, model.OverallScore)
			if len(model.Strengths) > 0 {
				fmt.Printf("  Strengths: %v\n", model.Strengths)
			}
			if len(model.Weaknesses) > 0 {
				fmt.Printf("  Weaknesses: %v\n", model.Weaknesses)
			}
		}
	}

	// Recommendations
	if len(report.Recommendations) > 0 {
		fmt.Printf("\n💡 RECOMMENDATIONS\n")
		for i, rec := range report.Recommendations {
			fmt.Printf("%d. [%s] %s\n", i+1, rec.Priority, rec.Title)
			fmt.Printf("   %s\n", rec.Description)
			fmt.Printf("   Impact: %s, Effort: %s, Timeline: %s\n",
				rec.Impact, rec.Effort, rec.Timeline)
			if len(rec.Actions) > 0 {
				fmt.Printf("   Actions:\n")
				for _, action := range rec.Actions {
					fmt.Printf("     - %s\n", action)
				}
			}
			fmt.Println()
		}
	}

	// Configuration
	fmt.Printf("\n⚙️  VALIDATION CONFIGURATION\n")
	fmt.Printf("K-Folds: %d\n", report.Configuration.CrossValidation.KFolds)
	fmt.Printf("Confidence Level: %.1f%%\n", report.Configuration.CrossValidation.ConfidenceLevel*100)
	fmt.Printf("Total Samples: %d\n", report.Configuration.DataGeneration.TotalSamples)
	fmt.Printf("Time Range: %v\n", report.Configuration.DataGeneration.TimeRange)
	fmt.Printf("Seasonal Patterns: %t\n", report.Configuration.DataGeneration.SeasonalPatterns)
	fmt.Printf("Trend Strength: %.3f\n", report.Configuration.DataGeneration.TrendStrength)
	fmt.Printf("Noise Level: %.3f\n", report.Configuration.DataGeneration.NoiseLevel)

	// Timing
	fmt.Printf("\n⏱️  TIMING INFORMATION\n")
	fmt.Printf("Total Validation Time: %v\n", report.ValidationTime)
	fmt.Printf("Generated At: %s\n", report.GeneratedAt.Format(time.RFC3339))

	fmt.Println("\n" + "="*80)
}

// saveReport saves the validation report to a JSON file
func saveReport(report *validation.ValidationReport, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode report: %w", err)
	}

	return nil
}

// getDefaultRiskCategories returns default risk categories
func getDefaultRiskCategories() []validation.RiskCategory {
	return []validation.RiskCategory{
		{
			Name:       "Financial Risk",
			BaseRisk:   0.3,
			Volatility: 0.1,
			IndustryBias: map[string]float64{
				"Finance": 1.2, "Technology": 0.8, "Healthcare": 0.9,
			},
			GeographicBias: map[string]float64{
				"United States": 1.0, "Canada": 0.9, "United Kingdom": 1.1,
			},
			SizeBias: map[string]float64{
				"Micro": 1.3, "Small": 1.1, "Medium": 1.0, "Large": 0.9, "Enterprise": 0.8,
			},
			AgeBias: map[string]float64{
				"1-2": 1.2, "3-5": 1.0, "6-10": 0.9, "11+": 0.8,
			},
		},
		{
			Name:       "Operational Risk",
			BaseRisk:   0.25,
			Volatility: 0.08,
			IndustryBias: map[string]float64{
				"Manufacturing": 1.3, "Construction": 1.2, "Transportation": 1.1,
			},
			GeographicBias: map[string]float64{
				"United States": 1.0, "Germany": 0.9, "Japan": 0.8,
			},
			SizeBias: map[string]float64{
				"Micro": 1.4, "Small": 1.2, "Medium": 1.0, "Large": 0.8, "Enterprise": 0.7,
			},
			AgeBias: map[string]float64{
				"1-2": 1.3, "3-5": 1.1, "6-10": 1.0, "11+": 0.9,
			},
		},
		{
			Name:       "Market Risk",
			BaseRisk:   0.2,
			Volatility: 0.15,
			IndustryBias: map[string]float64{
				"Retail": 1.4, "Technology": 1.2, "Entertainment": 1.3,
			},
			GeographicBias: map[string]float64{
				"United States": 1.0, "China": 1.2, "Brazil": 1.3,
			},
			SizeBias: map[string]float64{
				"Micro": 1.2, "Small": 1.1, "Medium": 1.0, "Large": 0.9, "Enterprise": 0.8,
			},
			AgeBias: map[string]float64{
				"1-2": 1.1, "3-5": 1.0, "6-10": 0.9, "11+": 0.8,
			},
		},
	}
}

// getDefaultIndustryWeights returns default industry weights
func getDefaultIndustryWeights() map[string]float64 {
	return map[string]float64{
		"Technology":      0.20,
		"Finance":         0.15,
		"Healthcare":      0.12,
		"Manufacturing":   0.10,
		"Retail":          0.08,
		"Real Estate":     0.06,
		"Construction":    0.05,
		"Transportation":  0.04,
		"Energy":          0.04,
		"Education":       0.03,
		"Entertainment":   0.03,
		"Food & Beverage": 0.03,
		"Agriculture":     0.02,
		"Mining":          0.02,
		"Utilities":       0.02,
		"Other":           0.01,
	}
}

// getDefaultGeographicBias returns default geographic bias
func getDefaultGeographicBias() map[string]float64 {
	return map[string]float64{
		"United States":  0.35,
		"Canada":         0.08,
		"United Kingdom": 0.07,
		"Germany":        0.06,
		"France":         0.05,
		"Japan":          0.05,
		"Australia":      0.04,
		"Brazil":         0.04,
		"India":          0.04,
		"China":          0.04,
		"Mexico":         0.03,
		"Italy":          0.03,
		"Spain":          0.03,
		"Netherlands":    0.02,
		"Sweden":         0.02,
		"Norway":         0.02,
		"Switzerland":    0.02,
		"Singapore":      0.02,
		"Other":          0.05,
	}
}
