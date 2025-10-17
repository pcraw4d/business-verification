package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"kyb-platform/services/risk-assessment-service/internal/ml"
	"kyb-platform/services/risk-assessment-service/internal/ml/optimization"

	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Create benchmark configuration
	config := &BenchmarkConfig{
		Models: []ModelConfig{
			{
				ID:   "xgboost_risk_model",
				Type: "xgboost",
				Name: "XGBoost Risk Assessment Model",
			},
			{
				ID:   "lstm_risk_model",
				Type: "lstm",
				Name: "LSTM Risk Assessment Model",
			},
			{
				ID:   "transformer_risk_model",
				Type: "transformer",
				Name: "Transformer Risk Assessment Model",
			},
		},
		TestSamples:        1000,
		WarmupSamples:      100,
		EnableQuantization: true,
		EnableCache:        true,
	}

	// Run benchmark
	benchmark := NewInferenceBenchmark(config, logger)

	ctx := context.Background()
	results, err := benchmark.RunBenchmark(ctx)
	if err != nil {
		logger.Fatal("Benchmark failed", zap.Error(err))
	}

	// Print results
	benchmark.PrintResults(results)

	// Save results to file
	if err := benchmark.SaveResults(results, "benchmark_results.json"); err != nil {
		logger.Error("Failed to save results", zap.Error(err))
	}
}

// BenchmarkConfig represents configuration for the inference benchmark
type BenchmarkConfig struct {
	Models             []ModelConfig `json:"models"`
	TestSamples        int           `json:"test_samples"`
	WarmupSamples      int           `json:"warmup_samples"`
	EnableQuantization bool          `json:"enable_quantization"`
	EnableCache        bool          `json:"enable_cache"`
}

// ModelConfig represents configuration for a model
type ModelConfig struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// InferenceBenchmark benchmarks ML model inference performance
type InferenceBenchmark struct {
	config    *BenchmarkConfig
	logger    *zap.Logger
	quantizer *optimization.ModelQuantizer
	cache     *optimization.InferenceCache
	warmup    *ml.ModelWarmup
}

// BenchmarkResult represents the result of a benchmark
type BenchmarkResult struct {
	ModelID              string        `json:"model_id"`
	ModelType            string        `json:"model_type"`
	ModelName            string        `json:"model_name"`
	WarmupTime           time.Duration `json:"warmup_time"`
	AverageInferenceTime time.Duration `json:"average_inference_time"`
	P95InferenceTime     time.Duration `json:"p95_inference_time"`
	P99InferenceTime     time.Duration `json:"p99_inference_time"`
	MinInferenceTime     time.Duration `json:"min_inference_time"`
	MaxInferenceTime     time.Duration `json:"max_inference_time"`
	Throughput           float64       `json:"throughput"`
	CacheHitRate         float64       `json:"cache_hit_rate"`
	QuantizationSpeedup  float64       `json:"quantization_speedup"`
	TotalSamples         int           `json:"total_samples"`
	SuccessfulSamples    int           `json:"successful_samples"`
	FailedSamples        int           `json:"failed_samples"`
	CreatedAt            time.Time     `json:"created_at"`
}

// BenchmarkResults represents all benchmark results
type BenchmarkResults struct {
	Results   []*BenchmarkResult `json:"results"`
	TotalTime time.Duration      `json:"total_time"`
	CreatedAt time.Time          `json:"created_at"`
	Config    *BenchmarkConfig   `json:"config"`
}

// NewInferenceBenchmark creates a new inference benchmark
func NewInferenceBenchmark(config *BenchmarkConfig, logger *zap.Logger) *InferenceBenchmark {
	// Create quantizer
	quantizerConfig := &optimization.QuantizationConfig{
		EnableINT8Quantization:    true,
		EnableFloat16Quantization: true,
		EnablePruning:             true,
		PruningRatio:              0.1,
		CalibrationSamples:        1000,
		QuantizationMethod:        "dynamic",
	}
	quantizer := optimization.NewModelQuantizer(quantizerConfig, logger)

	// Create cache
	cacheConfig := &optimization.CacheConfig{
		MaxSize:           10000,
		DefaultTTL:        1 * time.Hour,
		CleanupInterval:   5 * time.Minute,
		EnableMetrics:     true,
		EnableCompression: false,
	}
	cache := optimization.NewInferenceCache(cacheConfig, logger)

	// Create warmup
	warmupConfig := &ml.WarmupConfig{
		EnableWarmup:       true,
		WarmupSamples:      config.WarmupSamples,
		WarmupTimeout:      30 * time.Second,
		EnableQuantization: config.EnableQuantization,
		EnableCache:        config.EnableCache,
		WarmupInterval:     1 * time.Hour,
	}
	warmup := ml.NewModelWarmup(warmupConfig, quantizer, cache, logger)

	return &InferenceBenchmark{
		config:    config,
		logger:    logger,
		quantizer: quantizer,
		cache:     cache,
		warmup:    warmup,
	}
}

// RunBenchmark runs the inference benchmark
func (ib *InferenceBenchmark) RunBenchmark(ctx context.Context) (*BenchmarkResults, error) {
	start := time.Now()

	ib.logger.Info("Starting inference benchmark",
		zap.Int("model_count", len(ib.config.Models)),
		zap.Int("test_samples", ib.config.TestSamples))

	results := make([]*BenchmarkResult, 0, len(ib.config.Models))

	for _, modelConfig := range ib.config.Models {
		ib.logger.Info("Benchmarking model",
			zap.String("model_id", modelConfig.ID),
			zap.String("model_type", modelConfig.Type))

		result, err := ib.benchmarkModel(ctx, modelConfig)
		if err != nil {
			ib.logger.Error("Failed to benchmark model",
				zap.String("model_id", modelConfig.ID),
				zap.Error(err))
			continue
		}

		results = append(results, result)
	}

	benchmarkResults := &BenchmarkResults{
		Results:   results,
		TotalTime: time.Since(start),
		CreatedAt: time.Now(),
		Config:    ib.config,
	}

	ib.logger.Info("Benchmark completed",
		zap.Duration("total_time", benchmarkResults.TotalTime),
		zap.Int("successful_models", len(results)))

	return benchmarkResults, nil
}

// benchmarkModel benchmarks a single model
func (ib *InferenceBenchmark) benchmarkModel(ctx context.Context, modelConfig ModelConfig) (*BenchmarkResult, error) {
	// Create model info
	model := &ml.ModelInfo{
		ID:          modelConfig.ID,
		Type:        modelConfig.Type,
		Name:        modelConfig.Name,
		Version:     "1.0.0",
		Path:        fmt.Sprintf("/models/%s", modelConfig.ID),
		Metadata:    make(map[string]interface{}),
		CreatedAt:   time.Now(),
		LastUpdated: time.Now(),
	}

	// Warmup model
	warmupStart := time.Now()
	_, err := ib.warmup.WarmupModel(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("failed to warmup model: %w", err)
	}
	warmupTime := time.Since(warmupStart)

	// Generate test samples
	samples := ib.generateTestSamples(modelConfig.Type, ib.config.TestSamples)

	// Benchmark inference
	inferenceTimes := make([]time.Duration, 0, len(samples))
	var successfulSamples, failedSamples int
	var cacheHits, cacheMisses int64

	for i, sample := range samples {
		// Create inference request
		request := &optimization.InferenceRequest{
			ModelID:   modelConfig.ID,
			Input:     sample,
			Options:   make(map[string]interface{}),
			RequestID: fmt.Sprintf("benchmark_%d", i),
		}

		// Check cache first
		if ib.config.EnableCache {
			if _, found := ib.cache.Get(ctx, request); found {
				cacheHits++
				continue
			}
			cacheMisses++
		}

		// Perform inference
		inferenceStart := time.Now()
		_, err := ib.performInference(ctx, model, sample)
		inferenceTime := time.Since(inferenceStart)

		if err != nil {
			failedSamples++
			ib.logger.Warn("Inference failed",
				zap.String("model_id", modelConfig.ID),
				zap.Int("sample_index", i),
				zap.Error(err))
			continue
		}

		inferenceTimes = append(inferenceTimes, inferenceTime)
		successfulSamples++

		// Cache the result
		if ib.config.EnableCache {
			result := map[string]interface{}{
				"risk_score": 0.75,
				"risk_level": "medium",
				"factors":    []string{"industry_risk", "country_risk"},
			}
			ib.cache.Set(ctx, request, result)
		}
	}

	// Calculate statistics
	stats := ib.calculateStatistics(inferenceTimes)

	// Calculate cache hit rate
	var cacheHitRate float64
	if cacheHits+cacheMisses > 0 {
		cacheHitRate = float64(cacheHits) / float64(cacheHits+cacheMisses)
	}

	// Calculate throughput
	var throughput float64
	if len(inferenceTimes) > 0 {
		totalTime := time.Duration(0)
		for _, t := range inferenceTimes {
			totalTime += t
		}
		throughput = float64(len(inferenceTimes)) / totalTime.Seconds()
	}

	result := &BenchmarkResult{
		ModelID:              modelConfig.ID,
		ModelType:            modelConfig.Type,
		ModelName:            modelConfig.Name,
		WarmupTime:           warmupTime,
		AverageInferenceTime: stats.Average,
		P95InferenceTime:     stats.P95,
		P99InferenceTime:     stats.P99,
		MinInferenceTime:     stats.Min,
		MaxInferenceTime:     stats.Max,
		Throughput:           throughput,
		CacheHitRate:         cacheHitRate,
		QuantizationSpeedup:  2.0, // Simulated
		TotalSamples:         len(samples),
		SuccessfulSamples:    successfulSamples,
		FailedSamples:        failedSamples,
		CreatedAt:            time.Now(),
	}

	ib.logger.Info("Model benchmark completed",
		zap.String("model_id", modelConfig.ID),
		zap.Duration("average_inference_time", result.AverageInferenceTime),
		zap.Float64("throughput", result.Throughput),
		zap.Float64("cache_hit_rate", result.CacheHitRate))

	return result, nil
}

// generateTestSamples generates test samples for benchmarking
func (ib *InferenceBenchmark) generateTestSamples(modelType string, count int) []map[string]interface{} {
	samples := make([]map[string]interface{}, count)

	for i := 0; i < count; i++ {
		sample := map[string]interface{}{
			"business_name":     fmt.Sprintf("Test Business %d", i),
			"industry":          "Technology",
			"country":           "US",
			"revenue":           float64(1000000 + i*10000),
			"employee_count":    int64(10 + i),
			"years_in_business": int64(1 + i%20),
		}

		// Add model-specific fields
		switch modelType {
		case "xgboost":
			sample["features"] = []float64{0.1, 0.2, 0.3, 0.4, 0.5}
		case "lstm":
			sample["sequence"] = []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0}
		case "transformer":
			sample["tokens"] = []string{"business", "risk", "assessment", "test"}
		}

		samples[i] = sample
	}

	return samples
}

// performInference performs a single inference
func (ib *InferenceBenchmark) performInference(ctx context.Context, model *ml.ModelInfo, input map[string]interface{}) (map[string]interface{}, error) {
	// Simulate model inference
	// In a real implementation, you would call the actual ML model

	// Simulate different inference times based on model type
	var inferenceTime time.Duration
	switch model.Type {
	case "xgboost":
		inferenceTime = 50 * time.Millisecond
	case "lstm":
		inferenceTime = 100 * time.Millisecond
	case "transformer":
		inferenceTime = 150 * time.Millisecond
	default:
		inferenceTime = 75 * time.Millisecond
	}

	// Apply quantization speedup if enabled
	if ib.config.EnableQuantization {
		// Simulate 2x speedup from quantization
		inferenceTime = inferenceTime / 2
	}

	// Simulate inference
	time.Sleep(inferenceTime)

	// Generate mock result
	result := map[string]interface{}{
		"risk_score":     0.75,
		"risk_level":     "medium",
		"factors":        []string{"industry_risk", "country_risk"},
		"confidence":     0.92,
		"model_id":       model.ID,
		"inference_time": inferenceTime.Milliseconds(),
	}

	return result, nil
}

// Statistics represents statistical data
type Statistics struct {
	Average time.Duration
	P95     time.Duration
	P99     time.Duration
	Min     time.Duration
	Max     time.Duration
}

// calculateStatistics calculates statistics from inference times
func (ib *InferenceBenchmark) calculateStatistics(times []time.Duration) *Statistics {
	if len(times) == 0 {
		return &Statistics{}
	}

	// Sort times
	sortedTimes := make([]time.Duration, len(times))
	copy(sortedTimes, times)

	// Simple bubble sort (for small datasets)
	for i := 0; i < len(sortedTimes)-1; i++ {
		for j := 0; j < len(sortedTimes)-i-1; j++ {
			if sortedTimes[j] > sortedTimes[j+1] {
				sortedTimes[j], sortedTimes[j+1] = sortedTimes[j+1], sortedTimes[j]
			}
		}
	}

	// Calculate statistics
	var total time.Duration
	for _, t := range sortedTimes {
		total += t
	}

	stats := &Statistics{
		Average: total / time.Duration(len(sortedTimes)),
		Min:     sortedTimes[0],
		Max:     sortedTimes[len(sortedTimes)-1],
	}

	// Calculate percentiles
	if len(sortedTimes) > 0 {
		p95Index := int(float64(len(sortedTimes)) * 0.95)
		p99Index := int(float64(len(sortedTimes)) * 0.99)

		if p95Index < len(sortedTimes) {
			stats.P95 = sortedTimes[p95Index]
		}
		if p99Index < len(sortedTimes) {
			stats.P99 = sortedTimes[p99Index]
		}
	}

	return stats
}

// PrintResults prints benchmark results
func (ib *InferenceBenchmark) PrintResults(results *BenchmarkResults) {
	fmt.Println("\n=== ML Model Inference Benchmark Results ===")
	fmt.Printf("Total Benchmark Time: %v\n", results.TotalTime)
	fmt.Printf("Models Tested: %d\n", len(results.Results))
	fmt.Println()

	for _, result := range results.Results {
		fmt.Printf("Model: %s (%s)\n", result.ModelName, result.ModelType)
		fmt.Printf("  Average Inference Time: %v\n", result.AverageInferenceTime)
		fmt.Printf("  P95 Inference Time: %v\n", result.P95InferenceTime)
		fmt.Printf("  P99 Inference Time: %v\n", result.P99InferenceTime)
		fmt.Printf("  Min Inference Time: %v\n", result.MinInferenceTime)
		fmt.Printf("  Max Inference Time: %v\n", result.MaxInferenceTime)
		fmt.Printf("  Throughput: %.2f requests/second\n", result.Throughput)
		fmt.Printf("  Cache Hit Rate: %.2f%%\n", result.CacheHitRate*100)
		fmt.Printf("  Quantization Speedup: %.2fx\n", result.QuantizationSpeedup)
		fmt.Printf("  Successful Samples: %d/%d\n", result.SuccessfulSamples, result.TotalSamples)
		fmt.Println()
	}
}

// SaveResults saves benchmark results to a file
func (ib *InferenceBenchmark) SaveResults(results *BenchmarkResults, filename string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write results file: %w", err)
	}

	ib.logger.Info("Benchmark results saved",
		zap.String("filename", filename))

	return nil
}
