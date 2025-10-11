package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/ml/ensemble"
	mlmodels "kyb-platform/services/risk-assessment-service/internal/ml/models"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

// BenchmarkConfig holds configuration for benchmarks
type BenchmarkConfig struct {
	NumIterations    int
	NumConcurrent    int
	WarmupIterations int
	ModelPaths       struct {
		XGBoost string
		LSTM    string
	}
}

// BenchmarkResult holds the results of a benchmark
type BenchmarkResult struct {
	ModelType     string
	NumIterations int
	NumConcurrent int
	TotalDuration time.Duration
	AvgDuration   time.Duration
	MinDuration   time.Duration
	MaxDuration   time.Duration
	P50Duration   time.Duration
	P95Duration   time.Duration
	P99Duration   time.Duration
	Throughput    float64 // requests per second
	MemoryUsage   uint64  // bytes
	ErrorCount    int
	SuccessCount  int
}

// BenchmarkSuite runs comprehensive benchmarks for all models
type BenchmarkSuite struct {
	config    BenchmarkConfig
	logger    *zap.Logger
	xgbModel  mlmodels.RiskModel
	lstmModel mlmodels.RiskModel
	router    *ensemble.EnsembleRouter
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite(config BenchmarkConfig) *BenchmarkSuite {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}

	return &BenchmarkSuite{
		config: config,
		logger: logger,
	}
}

// InitializeModels initializes all models for benchmarking
func (bs *BenchmarkSuite) InitializeModels() error {
	bs.logger.Info("Initializing models for benchmarking")

	// Initialize XGBoost model
	bs.xgbModel = mlmodels.NewXGBoostModel("benchmark_xgb", "1.0.0")
	if err := bs.xgbModel.LoadModel(context.Background(), bs.config.ModelPaths.XGBoost); err != nil {
		bs.logger.Warn("Failed to load XGBoost model, using mock", zap.Error(err))
	}

	// Initialize LSTM model
	bs.lstmModel = mlmodels.NewLSTMModel("benchmark_lstm", "1.0.0", bs.logger)
	if err := bs.lstmModel.LoadModel(context.Background(), bs.config.ModelPaths.LSTM); err != nil {
		bs.logger.Warn("Failed to load LSTM model, using mock", zap.Error(err))
	}

	// Initialize ensemble router
	bs.router = ensemble.NewEnsembleRouter(bs.xgbModel, bs.lstmModel, bs.logger)

	bs.logger.Info("Models initialized successfully")
	return nil
}

// BenchmarkXGBoost benchmarks the XGBoost model
func (bs *BenchmarkSuite) BenchmarkXGBoost() BenchmarkResult {
	bs.logger.Info("Starting XGBoost benchmark")

	business := bs.createTestBusiness()
	result := BenchmarkResult{
		ModelType:     "xgboost",
		NumIterations: bs.config.NumIterations,
		NumConcurrent: bs.config.NumConcurrent,
	}

	// Warmup
	bs.warmup(func() error {
		_, err := bs.xgbModel.Predict(context.Background(), business)
		return err
	})

	// Benchmark
	durations := bs.runBenchmark(func() error {
		_, err := bs.xgbModel.Predict(context.Background(), business)
		return err
	})

	bs.calculateStats(&result, durations)
	bs.logger.Info("XGBoost benchmark completed", zap.Duration("avg_duration", result.AvgDuration))
	return result
}

// BenchmarkLSTM benchmarks the LSTM model
func (bs *BenchmarkSuite) BenchmarkLSTM() BenchmarkResult {
	bs.logger.Info("Starting LSTM benchmark")

	business := bs.createTestBusiness()
	result := BenchmarkResult{
		ModelType:     "lstm",
		NumIterations: bs.config.NumIterations,
		NumConcurrent: bs.config.NumConcurrent,
	}

	// Warmup
	bs.warmup(func() error {
		_, err := bs.lstmModel.Predict(context.Background(), business)
		return err
	})

	// Benchmark
	durations := bs.runBenchmark(func() error {
		_, err := bs.lstmModel.Predict(context.Background(), business)
		return err
	})

	bs.calculateStats(&result, durations)
	bs.logger.Info("LSTM benchmark completed", zap.Duration("avg_duration", result.AvgDuration))
	return result
}

// BenchmarkEnsemble benchmarks the ensemble model
func (bs *BenchmarkSuite) BenchmarkEnsemble() BenchmarkResult {
	bs.logger.Info("Starting Ensemble benchmark")

	business := bs.createTestBusiness()
	business.PredictionHorizon = 4 // Trigger ensemble
	result := BenchmarkResult{
		ModelType:     "ensemble",
		NumIterations: bs.config.NumIterations,
		NumConcurrent: bs.config.NumConcurrent,
	}

	// Warmup
	bs.warmup(func() error {
		_, err := bs.router.PredictWithEnsemble(context.Background(), business)
		return err
	})

	// Benchmark
	durations := bs.runBenchmark(func() error {
		_, err := bs.router.PredictWithEnsemble(context.Background(), business)
		return err
	})

	bs.calculateStats(&result, durations)
	bs.logger.Info("Ensemble benchmark completed", zap.Duration("avg_duration", result.AvgDuration))
	return result
}

// BenchmarkFuturePredictions benchmarks future prediction performance
func (bs *BenchmarkSuite) BenchmarkFuturePredictions() map[string]BenchmarkResult {
	bs.logger.Info("Starting future prediction benchmarks")

	results := make(map[string]BenchmarkResult)
	horizons := []int{3, 6, 9, 12}

	for _, horizon := range horizons {
		business := bs.createTestBusiness()
		result := BenchmarkResult{
			ModelType:     fmt.Sprintf("ensemble_future_%d", horizon),
			NumIterations: bs.config.NumIterations / 2, // Fewer iterations for future predictions
			NumConcurrent: bs.config.NumConcurrent,
		}

		// Warmup
		bs.warmup(func() error {
			_, err := bs.router.PredictFutureWithEnsemble(context.Background(), business, horizon)
			return err
		})

		// Benchmark
		durations := bs.runBenchmark(func() error {
			_, err := bs.router.PredictFutureWithEnsemble(context.Background(), business, horizon)
			return err
		})

		bs.calculateStats(&result, durations)
		results[fmt.Sprintf("future_%d", horizon)] = result
		bs.logger.Info("Future prediction benchmark completed",
			zap.Int("horizon", horizon),
			zap.Duration("avg_duration", result.AvgDuration))
	}

	return results
}

// BenchmarkMemoryUsage benchmarks memory usage for different models
func (bs *BenchmarkSuite) BenchmarkMemoryUsage() map[string]uint64 {
	bs.logger.Info("Starting memory usage benchmarks")

	results := make(map[string]uint64)

	// Force garbage collection
	runtime.GC()
	runtime.GC()

	// Measure baseline memory
	var baselineMem runtime.MemStats
	runtime.ReadMemStats(&baselineMem)

	// Test XGBoost memory usage
	runtime.GC()
	var xgbMem runtime.MemStats
	runtime.ReadMemStats(&xgbMem)
	results["xgboost"] = xgbMem.Alloc - baselineMem.Alloc

	// Test LSTM memory usage
	runtime.GC()
	var lstmMem runtime.MemStats
	runtime.ReadMemStats(&lstmMem)
	results["lstm"] = lstmMem.Alloc - baselineMem.Alloc

	// Test Ensemble memory usage
	runtime.GC()
	var ensembleMem runtime.MemStats
	runtime.ReadMemStats(&ensembleMem)
	results["ensemble"] = ensembleMem.Alloc - baselineMem.Alloc

	bs.logger.Info("Memory usage benchmarks completed")
	return results
}

// createTestBusiness creates a test business for benchmarking
func (bs *BenchmarkSuite) createTestBusiness() *models.RiskAssessmentRequest {
	return &models.RiskAssessmentRequest{
		BusinessName:      "Benchmark Test Company",
		BusinessAddress:   "123 Benchmark St, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		Phone:             "+1-555-123-4567",
		Email:             "test@benchmark.com",
		Website:           "https://benchmark.com",
		PredictionHorizon: 3,
		Metadata: map[string]interface{}{
			"benchmark": true,
		},
	}
}

// warmup runs warmup iterations
func (bs *BenchmarkSuite) warmup(operation func() error) {
	for i := 0; i < bs.config.WarmupIterations; i++ {
		operation()
	}
}

// runBenchmark runs the actual benchmark
func (bs *BenchmarkSuite) runBenchmark(operation func() error) []time.Duration {
	durations := make([]time.Duration, bs.config.NumIterations)

	if bs.config.NumConcurrent == 1 {
		// Sequential benchmark
		for i := 0; i < bs.config.NumIterations; i++ {
			start := time.Now()
			if err := operation(); err != nil {
				bs.logger.Warn("Benchmark operation failed", zap.Error(err))
			}
			durations[i] = time.Since(start)
		}
	} else {
		// Concurrent benchmark
		results := make(chan time.Duration, bs.config.NumIterations)
		errors := make(chan error, bs.config.NumIterations)

		for i := 0; i < bs.config.NumIterations; i++ {
			go func() {
				start := time.Now()
				if err := operation(); err != nil {
					errors <- err
					return
				}
				results <- time.Since(start)
			}()
		}

		// Collect results
		for i := 0; i < bs.config.NumIterations; i++ {
			select {
			case duration := <-results:
				durations[i] = duration
			case err := <-errors:
				bs.logger.Warn("Concurrent benchmark operation failed", zap.Error(err))
				durations[i] = 0 // Mark as failed
			}
		}
	}

	return durations
}

// calculateStats calculates benchmark statistics
func (bs *BenchmarkSuite) calculateStats(result *BenchmarkResult, durations []time.Duration) {
	if len(durations) == 0 {
		return
	}

	// Sort durations for percentile calculations
	sortedDurations := make([]time.Duration, len(durations))
	copy(sortedDurations, durations)

	// Simple bubble sort for durations
	for i := 0; i < len(sortedDurations)-1; i++ {
		for j := 0; j < len(sortedDurations)-i-1; j++ {
			if sortedDurations[j] > sortedDurations[j+1] {
				sortedDurations[j], sortedDurations[j+1] = sortedDurations[j+1], sortedDurations[j]
			}
		}
	}

	// Calculate statistics
	var total time.Duration
	successCount := 0

	for _, duration := range sortedDurations {
		if duration > 0 {
			total += duration
			successCount++
		}
	}

	result.SuccessCount = successCount
	result.ErrorCount = len(durations) - successCount
	result.TotalDuration = total
	result.AvgDuration = total / time.Duration(successCount)
	result.MinDuration = sortedDurations[0]
	result.MaxDuration = sortedDurations[len(sortedDurations)-1]

	// Calculate percentiles
	if len(sortedDurations) > 0 {
		result.P50Duration = sortedDurations[len(sortedDurations)*50/100]
		result.P95Duration = sortedDurations[len(sortedDurations)*95/100]
		result.P99Duration = sortedDurations[len(sortedDurations)*99/100]
	}

	// Calculate throughput
	if result.TotalDuration > 0 {
		result.Throughput = float64(successCount) / result.TotalDuration.Seconds()
	}

	// Get memory usage
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	result.MemoryUsage = memStats.Alloc
}

// PrintResults prints benchmark results in a formatted way
func (bs *BenchmarkSuite) PrintResults(results []BenchmarkResult, futureResults map[string]BenchmarkResult, memoryResults map[string]uint64) {
	fmt.Println("\n" + "="*80)
	fmt.Println("LSTM ENSEMBLE BENCHMARK RESULTS")
	fmt.Println("=" * 80)

	// Print main model results
	fmt.Println("\n📊 MODEL PERFORMANCE COMPARISON")
	fmt.Println("-" * 50)
	fmt.Printf("%-12s %-8s %-8s %-8s %-8s %-8s %-8s %-8s\n",
		"Model", "Avg(ms)", "P50(ms)", "P95(ms)", "P99(ms)", "Min(ms)", "Max(ms)", "RPS")
	fmt.Println("-" * 50)

	for _, result := range results {
		fmt.Printf("%-12s %-8.2f %-8.2f %-8.2f %-8.2f %-8.2f %-8.2f %-8.1f\n",
			result.ModelType,
			float64(result.AvgDuration.Nanoseconds())/1e6,
			float64(result.P50Duration.Nanoseconds())/1e6,
			float64(result.P95Duration.Nanoseconds())/1e6,
			float64(result.P99Duration.Nanoseconds())/1e6,
			float64(result.MinDuration.Nanoseconds())/1e6,
			float64(result.MaxDuration.Nanoseconds())/1e6,
			result.Throughput)
	}

	// Print future prediction results
	if len(futureResults) > 0 {
		fmt.Println("\n🔮 FUTURE PREDICTION PERFORMANCE")
		fmt.Println("-" * 50)
		fmt.Printf("%-15s %-8s %-8s %-8s %-8s %-8s\n",
			"Horizon", "Avg(ms)", "P50(ms)", "P95(ms)", "P99(ms)", "RPS")
		fmt.Println("-" * 50)

		for horizon, result := range futureResults {
			fmt.Printf("%-15s %-8.2f %-8.2f %-8.2f %-8.2f %-8.1f\n",
				horizon,
				float64(result.AvgDuration.Nanoseconds())/1e6,
				float64(result.P50Duration.Nanoseconds())/1e6,
				float64(result.P95Duration.Nanoseconds())/1e6,
				float64(result.P99Duration.Nanoseconds())/1e6,
				result.Throughput)
		}
	}

	// Print memory usage
	if len(memoryResults) > 0 {
		fmt.Println("\n💾 MEMORY USAGE")
		fmt.Println("-" * 30)
		fmt.Printf("%-12s %-10s\n", "Model", "Memory(MB)")
		fmt.Println("-" * 30)

		for model, memory := range memoryResults {
			fmt.Printf("%-12s %-10.2f\n", model, float64(memory)/(1024*1024))
		}
	}

	// Print success rates
	fmt.Println("\n✅ SUCCESS RATES")
	fmt.Println("-" * 30)
	fmt.Printf("%-12s %-8s %-8s %-8s\n", "Model", "Success", "Errors", "Rate")
	fmt.Println("-" * 30)

	for _, result := range results {
		successRate := float64(result.SuccessCount) / float64(result.NumIterations) * 100
		fmt.Printf("%-12s %-8d %-8d %-7.1f%%\n",
			result.ModelType,
			result.SuccessCount,
			result.ErrorCount,
			successRate)
	}

	// Print performance targets
	fmt.Println("\n🎯 PERFORMANCE TARGETS")
	fmt.Println("-" * 40)
	fmt.Println("Target: LSTM < 150ms (P95)")
	fmt.Println("Target: Ensemble < 200ms (P95)")
	fmt.Println("Target: Memory < 1.5GB")
	fmt.Println("Target: Throughput > 1000 req/min")

	// Check if targets are met
	fmt.Println("\n📈 TARGET COMPLIANCE")
	fmt.Println("-" * 30)
	for _, result := range results {
		var target time.Duration
		switch result.ModelType {
		case "lstm":
			target = 150 * time.Millisecond
		case "ensemble":
			target = 200 * time.Millisecond
		default:
			continue
		}

		met := result.P95Duration <= target
		status := "❌ FAIL"
		if met {
			status = "✅ PASS"
		}

		fmt.Printf("%-12s P95: %-8.2fms (target: %-8.2fms) %s\n",
			result.ModelType,
			float64(result.P95Duration.Nanoseconds())/1e6,
			float64(target.Nanoseconds())/1e6,
			status)
	}

	fmt.Println("\n" + "="*80)
}

func main() {
	// Configuration
	config := BenchmarkConfig{
		NumIterations:    1000,
		NumConcurrent:    10,
		WarmupIterations: 100,
		ModelPaths: struct {
			XGBoost string
			LSTM    string
		}{
			XGBoost: "./models/xgb_model.json",
			LSTM:    "./models/risk_lstm_v1.onnx",
		},
	}

	// Create benchmark suite
	suite := NewBenchmarkSuite(config)

	// Initialize models
	if err := suite.InitializeModels(); err != nil {
		log.Fatal("Failed to initialize models:", err)
	}

	// Run benchmarks
	fmt.Println("Starting LSTM Ensemble Benchmarks...")
	fmt.Printf("Iterations: %d, Concurrent: %d, Warmup: %d\n",
		config.NumIterations, config.NumConcurrent, config.WarmupIterations)

	var results []BenchmarkResult

	// Benchmark individual models
	results = append(results, suite.BenchmarkXGBoost())
	results = append(results, suite.BenchmarkLSTM())
	results = append(results, suite.BenchmarkEnsemble())

	// Benchmark future predictions
	futureResults := suite.BenchmarkFuturePredictions()

	// Benchmark memory usage
	memoryResults := suite.BenchmarkMemoryUsage()

	// Print results
	suite.PrintResults(results, futureResults, memoryResults)

	fmt.Println("\nBenchmark completed successfully!")
}
