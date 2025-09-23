package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"kyb-platform/internal/classification/multi_method_classifier"
	"kyb-platform/internal/config"
	"kyb-platform/internal/testing"

	_ "github.com/lib/pq"
)

// ClassificationTestRunner runs comprehensive accuracy tests
type ClassificationTestRunner struct {
	db         *sql.DB
	tester     *testing.ClassificationAccuracyTester
	classifier *multi_method_classifier.MultiMethodClassifier
	logger     *log.Logger
	config     *config.Config
}

// NewClassificationTestRunner creates a new test runner
func NewClassificationTestRunner(config *config.Config) (*ClassificationTestRunner, error) {
	// Initialize database connection
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize logger
	logger := log.New(os.Stdout, "[ACCURACY_TEST] ", log.LstdFlags|log.Lshortfile)

	// Initialize accuracy tester
	tester := testing.NewClassificationAccuracyTester(db, logger)

	// Initialize classifier (this would integrate with your existing classifier)
	classifier, err := multi_method_classifier.NewMultiMethodClassifier(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize classifier: %w", err)
	}

	return &ClassificationTestRunner{
		db:         db,
		tester:     tester,
		classifier: classifier,
		logger:     logger,
		config:     config,
	}, nil
}

// RunAccuracyTest runs the comprehensive accuracy test
func (ctr *ClassificationTestRunner) RunAccuracyTest(ctx context.Context) error {
	ctr.logger.Println("Starting comprehensive classification accuracy test...")

	startTime := time.Now()

	// Run the accuracy test
	metrics, err := ctr.tester.RunAccuracyTest(ctx, ctr.classifier)
	if err != nil {
		return fmt.Errorf("accuracy test failed: %w", err)
	}

	// Save the results
	if err := ctr.tester.SaveAccuracyReport(ctx, metrics); err != nil {
		ctr.logger.Printf("Warning: Failed to save accuracy report: %v", err)
	}

	// Print results
	ctr.printAccuracyResults(metrics)

	duration := time.Since(startTime)
	ctr.logger.Printf("Accuracy test completed in %v", duration)

	return nil
}

// printAccuracyResults prints comprehensive accuracy results
func (ctr *ClassificationTestRunner) printAccuracyResults(metrics *testing.AccuracyMetrics) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("CLASSIFICATION ACCURACY TEST RESULTS")
	fmt.Println(strings.Repeat("=", 80))

	// Overall metrics
	fmt.Printf("\n📊 OVERALL ACCURACY METRICS:\n")
	fmt.Printf("   Overall Accuracy:     %.2f%%\n", metrics.OverallAccuracy*100)
	fmt.Printf("   MCC Accuracy:         %.2f%%\n", metrics.MCCAccuracy*100)
	fmt.Printf("   NAICS Accuracy:       %.2f%%\n", metrics.NAICSAccuracy*100)
	fmt.Printf("   SIC Accuracy:         %.2f%%\n", metrics.SICAccuracy*100)
	fmt.Printf("   Industry Accuracy:    %.2f%%\n", metrics.IndustryAccuracy*100)
	fmt.Printf("   Confidence Accuracy:  %.2f%%\n", metrics.ConfidenceAccuracy*100)

	// Performance metrics
	fmt.Printf("\n⚡ PERFORMANCE METRICS:\n")
	fmt.Printf("   Average Processing Time: %.0fms\n", float64(metrics.ProcessingMetrics.AvgProcessingTime)/float64(time.Millisecond))
	fmt.Printf("   Min Processing Time:     %.0fms\n", float64(metrics.ProcessingMetrics.MinProcessingTime)/float64(time.Millisecond))
	fmt.Printf("   Max Processing Time:     %.0fms\n", float64(metrics.ProcessingMetrics.MaxProcessingTime)/float64(time.Millisecond))
	fmt.Printf("   95th Percentile:         %.0fms\n", float64(metrics.ProcessingMetrics.P95ProcessingTime)/float64(time.Millisecond))
	fmt.Printf("   99th Percentile:         %.0fms\n", float64(metrics.ProcessingMetrics.P99ProcessingTime)/float64(time.Millisecond))

	// Category metrics
	fmt.Printf("\n📋 CATEGORY METRICS:\n")
	for category, catMetrics := range metrics.CategoryMetrics {
		fmt.Printf("   %s:\n", category)
		fmt.Printf("     Samples: %d, Accuracy: %.2f%%, Avg Confidence: %.2f%%, Processing: %.0fms\n",
			catMetrics.SampleCount,
			catMetrics.Accuracy*100,
			catMetrics.AvgConfidence*100,
			float64(catMetrics.ProcessingTime)/float64(time.Millisecond))
	}

	// Error analysis
	fmt.Printf("\n🚨 ERROR ANALYSIS:\n")
	fmt.Printf("   Total Errors:        %d\n", metrics.ErrorAnalysis.TotalErrors)
	fmt.Printf("   Error Rate:          %.2f%%\n", metrics.ErrorAnalysis.ErrorRate*100)
	fmt.Printf("   False Positives:     %d\n", metrics.ErrorAnalysis.FalsePositives)
	fmt.Printf("   False Negatives:     %d\n", metrics.ErrorAnalysis.FalseNegatives)

	if len(metrics.ErrorAnalysis.ErrorCategories) > 0 {
		fmt.Printf("   Error Categories:\n")
		for category, count := range metrics.ErrorAnalysis.ErrorCategories {
			fmt.Printf("     %s: %d\n", category, count)
		}
	}

	// Recommendations
	if len(metrics.Recommendations) > 0 {
		fmt.Printf("\n💡 RECOMMENDATIONS:\n")
		for i, recommendation := range metrics.Recommendations {
			fmt.Printf("   %d. %s\n", i+1, recommendation)
		}
	}

	// Performance assessment
	fmt.Printf("\n🎯 PERFORMANCE ASSESSMENT:\n")
	ctr.assessPerformance(metrics)

	fmt.Println("\n" + strings.Repeat("=", 80))
}

// assessPerformance provides performance assessment
func (ctr *ClassificationTestRunner) assessPerformance(metrics *testing.AccuracyMetrics) {
	// Overall accuracy assessment
	if metrics.OverallAccuracy >= 0.95 {
		fmt.Printf("   ✅ Overall accuracy (%.1f%%) meets target (95%%)\n", metrics.OverallAccuracy*100)
	} else if metrics.OverallAccuracy >= 0.90 {
		fmt.Printf("   ⚠️  Overall accuracy (%.1f%%) is close to target (95%%)\n", metrics.OverallAccuracy*100)
	} else {
		fmt.Printf("   ❌ Overall accuracy (%.1f%%) is below target (95%%)\n", metrics.OverallAccuracy*100)
	}

	// Processing time assessment
	avgProcessingMs := float64(metrics.ProcessingMetrics.AvgProcessingTime) / float64(time.Millisecond)
	if avgProcessingMs <= 200 {
		fmt.Printf("   ✅ Processing time (%.0fms) meets target (200ms)\n", avgProcessingMs)
	} else if avgProcessingMs <= 500 {
		fmt.Printf("   ⚠️  Processing time (%.0fms) is acceptable but could be improved\n", avgProcessingMs)
	} else {
		fmt.Printf("   ❌ Processing time (%.0fms) exceeds target (200ms)\n", avgProcessingMs)
	}

	// Error rate assessment
	if metrics.ErrorAnalysis.ErrorRate <= 0.05 {
		fmt.Printf("   ✅ Error rate (%.1f%%) is acceptable\n", metrics.ErrorAnalysis.ErrorRate*100)
	} else if metrics.ErrorAnalysis.ErrorRate <= 0.10 {
		fmt.Printf("   ⚠️  Error rate (%.1f%%) is high but manageable\n", metrics.ErrorAnalysis.ErrorRate*100)
	} else {
		fmt.Printf("   ❌ Error rate (%.1f%%) is too high\n", metrics.ErrorAnalysis.ErrorRate*100)
	}

	// Individual classification type assessment
	classifications := map[string]float64{
		"MCC":      metrics.MCCAccuracy,
		"NAICS":    metrics.NAICSAccuracy,
		"SIC":      metrics.SICAccuracy,
		"Industry": metrics.IndustryAccuracy,
	}

	for name, accuracy := range classifications {
		if accuracy >= 0.90 {
			fmt.Printf("   ✅ %s accuracy (%.1f%%) is good\n", name, accuracy*100)
		} else if accuracy >= 0.80 {
			fmt.Printf("   ⚠️  %s accuracy (%.1f%%) needs improvement\n", name, accuracy*100)
		} else {
			fmt.Printf("   ❌ %s accuracy (%.1f%%) is poor\n", name, accuracy*100)
		}
	}
}

// Close closes the test runner
func (ctr *ClassificationTestRunner) Close() error {
	return ctr.db.Close()
}

func main() {
	// Command line flags
	var (
		configFile = flag.String("config", "configs/dev/config.yaml", "Configuration file path")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set log level
	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Create test runner
	runner, err := NewClassificationTestRunner(cfg)
	if err != nil {
		log.Fatalf("Failed to create test runner: %v", err)
	}
	defer runner.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Run accuracy test
	if err := runner.RunAccuracyTest(ctx); err != nil {
		log.Fatalf("Accuracy test failed: %v", err)
	}

	log.Println("Classification accuracy test completed successfully!")
}
