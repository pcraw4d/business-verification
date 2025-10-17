package ml

import (
	"context"
	"fmt"
	"sync"
	"time"

	"kyb-platform/services/risk-assessment-service/internal/ml/optimization"

	"go.uber.org/zap"
)

// ModelWarmup handles model warmup for faster inference
type ModelWarmup struct {
	logger    *zap.Logger
	mu        sync.RWMutex
	stats     *WarmupStats
	config    *WarmupConfig
	quantizer *optimization.ModelQuantizer
	cache     *optimization.InferenceCache
	warmedUp  map[string]bool
}

// WarmupStats represents statistics for model warmup
type WarmupStats struct {
	ModelsWarmedUp    int64         `json:"models_warmed_up"`
	WarmupTime        time.Duration `json:"warmup_time"`
	AverageWarmupTime time.Duration `json:"average_warmup_time"`
	LastWarmup        time.Time     `json:"last_warmup"`
	WarmupFailures    int64         `json:"warmup_failures"`
}

// WarmupConfig represents configuration for model warmup
type WarmupConfig struct {
	EnableWarmup       bool          `json:"enable_warmup"`
	WarmupSamples      int           `json:"warmup_samples"`
	WarmupTimeout      time.Duration `json:"warmup_timeout"`
	EnableQuantization bool          `json:"enable_quantization"`
	EnableCache        bool          `json:"enable_cache"`
	WarmupInterval     time.Duration `json:"warmup_interval"`
}

// ModelInfo represents information about a model
type ModelInfo struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Path        string                 `json:"path"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	LastUpdated time.Time              `json:"last_updated"`
}

// WarmupResult represents the result of model warmup
type WarmupResult struct {
	ModelID       string        `json:"model_id"`
	Success       bool          `json:"success"`
	WarmupTime    time.Duration `json:"warmup_time"`
	InferenceTime time.Duration `json:"inference_time"`
	Error         string        `json:"error,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}

// NewModelWarmup creates a new model warmup handler
func NewModelWarmup(config *WarmupConfig, quantizer *optimization.ModelQuantizer, cache *optimization.InferenceCache, logger *zap.Logger) *ModelWarmup {
	if config == nil {
		config = &WarmupConfig{
			EnableWarmup:       true,
			WarmupSamples:      100,
			WarmupTimeout:      30 * time.Second,
			EnableQuantization: true,
			EnableCache:        true,
			WarmupInterval:     1 * time.Hour,
		}
	}

	return &ModelWarmup{
		logger:    logger,
		stats:     &WarmupStats{},
		config:    config,
		quantizer: quantizer,
		cache:     cache,
		warmedUp:  make(map[string]bool),
	}
}

// WarmupModel warms up a model for faster inference
func (mw *ModelWarmup) WarmupModel(ctx context.Context, model *ModelInfo) (*WarmupResult, error) {
	start := time.Now()

	mw.logger.Info("Starting model warmup",
		zap.String("model_id", model.ID),
		zap.String("model_type", model.Type),
		zap.String("model_name", model.Name))

	// Check if model is already warmed up
	mw.mu.RLock()
	if mw.warmedUp[model.ID] {
		mw.mu.RUnlock()
		mw.logger.Info("Model already warmed up",
			zap.String("model_id", model.ID))

		return &WarmupResult{
			ModelID:    model.ID,
			Success:    true,
			WarmupTime: 0,
			CreatedAt:  time.Now(),
		}, nil
	}
	mw.mu.RUnlock()

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, mw.config.WarmupTimeout)
	defer cancel()

	// Perform warmup
	inferenceTime, err := mw.performWarmup(ctx, model)
	if err != nil {
		mw.mu.Lock()
		mw.stats.WarmupFailures++
		mw.mu.Unlock()

		mw.logger.Error("Model warmup failed",
			zap.String("model_id", model.ID),
			zap.Error(err))

		return &WarmupResult{
			ModelID:    model.ID,
			Success:    false,
			WarmupTime: time.Since(start),
			Error:      err.Error(),
			CreatedAt:  time.Now(),
		}, err
	}

	// Mark model as warmed up
	mw.mu.Lock()
	mw.warmedUp[model.ID] = true
	mw.stats.ModelsWarmedUp++
	mw.stats.WarmupTime = time.Since(start)
	mw.stats.AverageWarmupTime = (mw.stats.AverageWarmupTime + mw.stats.WarmupTime) / 2
	mw.stats.LastWarmup = time.Now()
	mw.mu.Unlock()

	result := &WarmupResult{
		ModelID:       model.ID,
		Success:       true,
		WarmupTime:    time.Since(start),
		InferenceTime: inferenceTime,
		CreatedAt:     time.Now(),
	}

	mw.logger.Info("Model warmup completed",
		zap.String("model_id", model.ID),
		zap.Duration("warmup_time", result.WarmupTime),
		zap.Duration("inference_time", result.InferenceTime))

	return result, nil
}

// WarmupAllModels warms up all available models
func (mw *ModelWarmup) WarmupAllModels(ctx context.Context, models []*ModelInfo) error {
	mw.logger.Info("Starting warmup of all models",
		zap.Int("model_count", len(models)))

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	// Limit concurrency
	semaphore := make(chan struct{}, 5)

	for _, model := range models {
		wg.Add(1)
		go func(m *ModelInfo) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result, err := mw.WarmupModel(ctx, m)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to warmup model %s: %w", m.ID, err))
				mu.Unlock()
			} else {
				mw.logger.Debug("Model warmup completed",
					zap.String("model_id", m.ID),
					zap.Bool("success", result.Success))
			}
		}(model)
	}

	wg.Wait()

	if len(errors) > 0 {
		mw.logger.Error("Some models failed to warmup",
			zap.Int("error_count", len(errors)))
		return fmt.Errorf("failed to warmup %d models", len(errors))
	}

	mw.logger.Info("All models warmed up successfully",
		zap.Int("model_count", len(models)))

	return nil
}

// GetStats returns warmup statistics
func (mw *ModelWarmup) GetStats() *WarmupStats {
	mw.mu.RLock()
	defer mw.mu.RUnlock()

	stats := *mw.stats
	return &stats
}

// IsWarmedUp checks if a model is warmed up
func (mw *ModelWarmup) IsWarmedUp(modelID string) bool {
	mw.mu.RLock()
	defer mw.mu.RUnlock()

	return mw.warmedUp[modelID]
}

// ResetWarmup resets the warmup status for a model
func (mw *ModelWarmup) ResetWarmup(modelID string) {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	delete(mw.warmedUp, modelID)

	mw.logger.Info("Model warmup status reset",
		zap.String("model_id", modelID))
}

// Helper methods

func (mw *ModelWarmup) performWarmup(ctx context.Context, model *ModelInfo) (time.Duration, error) {
	// Generate warmup samples
	samples := mw.generateWarmupSamples(model)

	// Perform warmup inference
	var totalInferenceTime time.Duration
	var successfulInferences int

	for i, sample := range samples {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			// Perform inference
			start := time.Now()
			_, err := mw.performInference(ctx, model, sample)
			inferenceTime := time.Since(start)

			if err != nil {
				mw.logger.Warn("Warmup inference failed",
					zap.String("model_id", model.ID),
					zap.Int("sample_index", i),
					zap.Error(err))
				continue
			}

			totalInferenceTime += inferenceTime
			successfulInferences++

			// Cache the result if enabled
			if mw.config.EnableCache && mw.cache != nil {
				request := &optimization.InferenceRequest{
					ModelID:   model.ID,
					Input:     sample,
					Options:   make(map[string]interface{}),
					RequestID: fmt.Sprintf("warmup_%d", i),
				}

				result := map[string]interface{}{
					"risk_score": 0.75,
					"risk_level": "medium",
					"factors":    []string{"industry_risk", "country_risk"},
				}

				if err := mw.cache.Set(ctx, request, result); err != nil {
					mw.logger.Warn("Failed to cache warmup result",
						zap.String("model_id", model.ID),
						zap.Int("sample_index", i),
						zap.Error(err))
				}
			}
		}
	}

	if successfulInferences == 0 {
		return 0, fmt.Errorf("no successful warmup inferences")
	}

	averageInferenceTime := totalInferenceTime / time.Duration(successfulInferences)

	mw.logger.Info("Warmup inference completed",
		zap.String("model_id", model.ID),
		zap.Int("total_samples", len(samples)),
		zap.Int("successful_inferences", successfulInferences),
		zap.Duration("average_inference_time", averageInferenceTime))

	return averageInferenceTime, nil
}

func (mw *ModelWarmup) generateWarmupSamples(model *ModelInfo) []map[string]interface{} {
	samples := make([]map[string]interface{}, mw.config.WarmupSamples)

	// Generate diverse samples based on model type
	for i := 0; i < mw.config.WarmupSamples; i++ {
		sample := map[string]interface{}{
			"business_name":     fmt.Sprintf("Sample Business %d", i),
			"industry":          "Technology",
			"country":           "US",
			"revenue":           float64(1000000 + i*10000),
			"employee_count":    int64(10 + i),
			"years_in_business": int64(1 + i%20),
		}

		// Add model-specific fields
		switch model.Type {
		case "xgboost":
			sample["features"] = []float64{0.1, 0.2, 0.3, 0.4, 0.5}
		case "lstm":
			sample["sequence"] = []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0}
		case "transformer":
			sample["tokens"] = []string{"business", "risk", "assessment", "sample"}
		}

		samples[i] = sample
	}

	return samples
}

func (mw *ModelWarmup) performInference(ctx context.Context, model *ModelInfo, input map[string]interface{}) (map[string]interface{}, error) {
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
	if mw.config.EnableQuantization && mw.quantizer != nil {
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

// WarmupScheduler schedules periodic model warmup
type WarmupScheduler struct {
	warmup   *ModelWarmup
	logger   *zap.Logger
	models   []*ModelInfo
	interval time.Duration
	running  bool
	stopCh   chan struct{}
}

// NewWarmupScheduler creates a new warmup scheduler
func NewWarmupScheduler(warmup *ModelWarmup, models []*ModelInfo, interval time.Duration, logger *zap.Logger) *WarmupScheduler {
	return &WarmupScheduler{
		warmup:   warmup,
		logger:   logger,
		models:   models,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start starts the warmup scheduler
func (ws *WarmupScheduler) Start() error {
	if ws.running {
		return fmt.Errorf("warmup scheduler is already running")
	}

	ws.running = true

	go ws.scheduleLoop()

	ws.logger.Info("Warmup scheduler started",
		zap.Duration("interval", ws.interval))

	return nil
}

// Stop stops the warmup scheduler
func (ws *WarmupScheduler) Stop() error {
	if !ws.running {
		return fmt.Errorf("warmup scheduler is not running")
	}

	ws.running = false
	close(ws.stopCh)

	ws.logger.Info("Warmup scheduler stopped")

	return nil
}

func (ws *WarmupScheduler) scheduleLoop() {
	ticker := time.NewTicker(ws.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ws.stopCh:
			return
		case <-ticker.C:
			ws.performScheduledWarmup()
		}
	}
}

func (ws *WarmupScheduler) performScheduledWarmup() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ws.logger.Info("Performing scheduled warmup",
		zap.Int("model_count", len(ws.models)))

	if err := ws.warmup.WarmupAllModels(ctx, ws.models); err != nil {
		ws.logger.Error("Scheduled warmup failed",
			zap.Error(err))
	} else {
		ws.logger.Info("Scheduled warmup completed successfully")
	}
}
