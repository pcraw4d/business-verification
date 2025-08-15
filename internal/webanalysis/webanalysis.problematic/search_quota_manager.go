package webanalysis

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SearchQuotaManager provides comprehensive quota management for all search engines
type SearchQuotaManager struct {
	engines     map[string]*EngineQuotaInfo
	globalQuota *GlobalQuotaInfo
	config      QuotaManagerConfig
	mu          sync.RWMutex
}

// QuotaManagerConfig holds configuration for quota management
type QuotaManagerConfig struct {
	EnableQuotaManagement bool          `json:"enable_quota_management"`
	EnableRateLimiting    bool          `json:"enable_rate_limiting"`
	EnableQuotaTracking   bool          `json:"enable_quota_tracking"`
	EnableQuotaAlerts     bool          `json:"enable_quota_alerts"`
	QuotaResetTime        time.Time     `json:"quota_reset_time"`
	QuotaResetInterval    time.Duration `json:"quota_reset_interval"`
	MaxConcurrentRequests int           `json:"max_concurrent_requests"`
	RequestTimeout        time.Duration `json:"request_timeout"`
	AlertThreshold        float64       `json:"alert_threshold"`
	RetryDelay            time.Duration `json:"retry_delay"`
	MaxRetries            int           `json:"max_retries"`
}

// EngineQuotaInfo holds quota information for a specific search engine
type EngineQuotaInfo struct {
	EngineName            string    `json:"engine_name"`
	DailyQuotaUsed        int       `json:"daily_quota_used"`
	DailyQuotaLimit       int       `json:"daily_quota_limit"`
	HourlyQuotaUsed       int       `json:"hourly_quota_used"`
	HourlyQuotaLimit      int       `json:"hourly_quota_limit"`
	MinuteQuotaUsed       int       `json:"minute_quota_used"`
	MinuteQuotaLimit      int       `json:"minute_quota_limit"`
	LastQuotaReset        time.Time `json:"last_quota_reset"`
	LastHourlyReset       time.Time `json:"last_hourly_reset"`
	LastMinuteReset       time.Time `json:"last_minute_reset"`
	ConcurrentRequests    int       `json:"concurrent_requests"`
	MaxConcurrentRequests int       `json:"max_concurrent_requests"`
	IsEnabled             bool      `json:"is_enabled"`
	Priority              int       `json:"priority"`
	FallbackEngines       []string  `json:"fallback_engines"`
}

// GlobalQuotaInfo holds global quota information
type GlobalQuotaInfo struct {
	TotalDailyQuotaUsed        int       `json:"total_daily_quota_used"`
	TotalDailyQuotaLimit       int       `json:"total_daily_quota_limit"`
	TotalHourlyQuotaUsed       int       `json:"total_hourly_quota_used"`
	TotalHourlyQuotaLimit      int       `json:"total_hourly_quota_limit"`
	TotalMinuteQuotaUsed       int       `json:"total_minute_quota_used"`
	TotalMinuteQuotaLimit      int       `json:"total_minute_quota_limit"`
	LastGlobalQuotaReset       time.Time `json:"last_global_quota_reset"`
	LastGlobalHourlyReset      time.Time `json:"last_global_hourly_reset"`
	LastGlobalMinuteReset      time.Time `json:"last_global_minute_reset"`
	TotalConcurrentRequests    int       `json:"total_concurrent_requests"`
	MaxTotalConcurrentRequests int       `json:"max_total_concurrent_requests"`
}

// QuotaRequest represents a quota request
type QuotaRequest struct {
	EngineName string            `json:"engine_name"`
	RequestID  string            `json:"request_id"`
	Priority   int               `json:"priority"`
	Timeout    time.Duration     `json:"timeout"`
	Metadata   map[string]string `json:"metadata"`
}

// QuotaResponse represents a quota response
type QuotaResponse struct {
	RequestID             string            `json:"request_id"`
	EngineName            string            `json:"engine_name"`
	IsAllowed             bool              `json:"is_allowed"`
	WaitTime              time.Duration     `json:"wait_time"`
	QuotaRemaining        int               `json:"quota_remaining"`
	QuotaUsed             int               `json:"quota_used"`
	QuotaLimit            int               `json:"quota_limit"`
	ConcurrentRequests    int               `json:"concurrent_requests"`
	MaxConcurrentRequests int               `json:"max_concurrent_requests"`
	NextResetTime         time.Time         `json:"next_reset_time"`
	Warnings              []string          `json:"warnings"`
	Errors                []string          `json:"errors"`
	Metadata              map[string]string `json:"metadata"`
}

// QuotaAlert represents a quota alert
type QuotaAlert struct {
	AlertID    string    `json:"alert_id"`
	EngineName string    `json:"engine_name"`
	AlertType  string    `json:"alert_type"`
	AlertLevel string    `json:"alert_level"`
	Message    string    `json:"message"`
	QuotaUsed  int       `json:"quota_used"`
	QuotaLimit int       `json:"quota_limit"`
	Percentage float64   `json:"percentage"`
	Timestamp  time.Time `json:"timestamp"`
	IsResolved bool      `json:"is_resolved"`
}

// NewSearchQuotaManager creates a new search quota manager
func NewSearchQuotaManager() *SearchQuotaManager {
	config := QuotaManagerConfig{
		EnableQuotaManagement: true,
		EnableRateLimiting:    true,
		EnableQuotaTracking:   true,
		EnableQuotaAlerts:     true,
		QuotaResetTime:        time.Now().Truncate(24 * time.Hour),
		QuotaResetInterval:    24 * time.Hour,
		MaxConcurrentRequests: 10,
		RequestTimeout:        time.Second * 30,
		AlertThreshold:        0.8, // 80%
		RetryDelay:            time.Second * 1,
		MaxRetries:            3,
	}

	manager := &SearchQuotaManager{
		engines: make(map[string]*EngineQuotaInfo),
		globalQuota: &GlobalQuotaInfo{
			TotalDailyQuotaLimit:       10000,
			TotalHourlyQuotaLimit:      1000,
			TotalMinuteQuotaLimit:      100,
			MaxTotalConcurrentRequests: 50,
		},
		config: config,
	}

	// Initialize default engines
	manager.initializeDefaultEngines()

	return manager
}

// initializeDefaultEngines sets up default search engines with their quotas
func (sqm *SearchQuotaManager) initializeDefaultEngines() {
	// Google Custom Search Engine
	sqm.engines["google"] = &EngineQuotaInfo{
		EngineName:            "google",
		DailyQuotaLimit:       10000,
		HourlyQuotaLimit:      1000,
		MinuteQuotaLimit:      10,
		MaxConcurrentRequests: 5,
		IsEnabled:             true,
		Priority:              1,
		FallbackEngines:       []string{"bing", "duckduckgo"},
	}

	// Bing Search Engine
	sqm.engines["bing"] = &EngineQuotaInfo{
		EngineName:            "bing",
		DailyQuotaLimit:       3000,
		HourlyQuotaLimit:      300,
		MinuteQuotaLimit:      3,
		MaxConcurrentRequests: 3,
		IsEnabled:             true,
		Priority:              2,
		FallbackEngines:       []string{"duckduckgo"},
	}

	// DuckDuckGo Search Engine
	sqm.engines["duckduckgo"] = &EngineQuotaInfo{
		EngineName:            "duckduckgo",
		DailyQuotaLimit:       1000,
		HourlyQuotaLimit:      100,
		MinuteQuotaLimit:      1,
		MaxConcurrentRequests: 2,
		IsEnabled:             true,
		Priority:              3,
		FallbackEngines:       []string{},
	}
}

// RequestQuota requests quota for a specific search engine
func (sqm *SearchQuotaManager) RequestQuota(ctx context.Context, req *QuotaRequest) (*QuotaResponse, error) {
	sqm.mu.Lock()
	defer sqm.mu.Unlock()

	if !sqm.config.EnableQuotaManagement {
		return sqm.createAllowedResponse(req), nil
	}

	// Check if engine exists
	engineInfo, exists := sqm.engines[req.EngineName]
	if !exists {
		return nil, fmt.Errorf("engine %s not found", req.EngineName)
	}

	if !engineInfo.IsEnabled {
		return nil, fmt.Errorf("engine %s is disabled", req.EngineName)
	}

	// Reset quotas if needed
	sqm.resetQuotasIfNeeded(engineInfo)

	// Check global quota
	if !sqm.checkGlobalQuota() {
		return sqm.createDeniedResponse(req, "Global quota exceeded"), nil
	}

	// Check engine-specific quotas
	if !sqm.checkEngineQuota(engineInfo) {
		return sqm.createDeniedResponse(req, "Engine quota exceeded"), nil
	}

	// Check concurrent requests
	if !sqm.checkConcurrentRequests(engineInfo) {
		return sqm.createDeniedResponse(req, "Too many concurrent requests"), nil
	}

	// Increment usage
	sqm.incrementUsage(engineInfo)

	// Check for alerts
	sqm.checkForAlerts(engineInfo)

	return sqm.createAllowedResponse(req), nil
}

// ReleaseQuota releases quota for a completed request
func (sqm *SearchQuotaManager) ReleaseQuota(engineName, requestID string) error {
	sqm.mu.Lock()
	defer sqm.mu.Unlock()

	engineInfo, exists := sqm.engines[engineName]
	if !exists {
		return fmt.Errorf("engine %s not found", engineName)
	}

	// Decrement concurrent requests
	if engineInfo.ConcurrentRequests > 0 {
		engineInfo.ConcurrentRequests--
	}

	if sqm.globalQuota.TotalConcurrentRequests > 0 {
		sqm.globalQuota.TotalConcurrentRequests--
	}

	return nil
}

// GetQuotaStatus returns the current quota status for all engines
func (sqm *SearchQuotaManager) GetQuotaStatus() map[string]interface{} {
	sqm.mu.RLock()
	defer sqm.mu.RUnlock()

	status := make(map[string]interface{})

	// Global quota status
	status["global"] = map[string]interface{}{
		"total_daily_quota_used":    sqm.globalQuota.TotalDailyQuotaUsed,
		"total_daily_quota_limit":   sqm.globalQuota.TotalDailyQuotaLimit,
		"total_hourly_quota_used":   sqm.globalQuota.TotalHourlyQuotaUsed,
		"total_hourly_quota_limit":  sqm.globalQuota.TotalHourlyQuotaLimit,
		"total_minute_quota_used":   sqm.globalQuota.TotalMinuteQuotaUsed,
		"total_minute_quota_limit":  sqm.globalQuota.TotalMinuteQuotaLimit,
		"total_concurrent_requests": sqm.globalQuota.TotalConcurrentRequests,
		"max_concurrent_requests":   sqm.globalQuota.MaxTotalConcurrentRequests,
		"last_reset":                sqm.globalQuota.LastGlobalQuotaReset,
	}

	// Engine-specific quota status
	engines := make(map[string]interface{})
	for name, engine := range sqm.engines {
		engines[name] = map[string]interface{}{
			"engine_name":             engine.EngineName,
			"daily_quota_used":        engine.DailyQuotaUsed,
			"daily_quota_limit":       engine.DailyQuotaLimit,
			"hourly_quota_used":       engine.HourlyQuotaUsed,
			"hourly_quota_limit":      engine.HourlyQuotaLimit,
			"minute_quota_used":       engine.MinuteQuotaUsed,
			"minute_quota_limit":      engine.MinuteQuotaLimit,
			"concurrent_requests":     engine.ConcurrentRequests,
			"max_concurrent_requests": engine.MaxConcurrentRequests,
			"is_enabled":              engine.IsEnabled,
			"priority":                engine.Priority,
			"fallback_engines":        engine.FallbackEngines,
			"last_quota_reset":        engine.LastQuotaReset,
			"daily_quota_remaining":   engine.DailyQuotaLimit - engine.DailyQuotaUsed,
			"hourly_quota_remaining":  engine.HourlyQuotaLimit - engine.HourlyQuotaUsed,
			"minute_quota_remaining":  engine.MinuteQuotaLimit - engine.MinuteQuotaUsed,
			"daily_quota_percentage":  float64(engine.DailyQuotaUsed) / float64(engine.DailyQuotaLimit) * 100,
			"hourly_quota_percentage": float64(engine.HourlyQuotaUsed) / float64(engine.HourlyQuotaLimit) * 100,
			"minute_quota_percentage": float64(engine.MinuteQuotaUsed) / float64(engine.MinuteQuotaLimit) * 100,
		}
	}
	status["engines"] = engines

	// Configuration
	status["config"] = sqm.config

	return status
}

// AddEngine adds a new search engine to the quota manager
func (sqm *SearchQuotaManager) AddEngine(engineName string, engineInfo *EngineQuotaInfo) error {
	sqm.mu.Lock()
	defer sqm.mu.Unlock()

	if _, exists := sqm.engines[engineName]; exists {
		return fmt.Errorf("engine %s already exists", engineName)
	}

	engineInfo.EngineName = engineName
	engineInfo.LastQuotaReset = time.Now()
	engineInfo.LastHourlyReset = time.Now()
	engineInfo.LastMinuteReset = time.Now()

	sqm.engines[engineName] = engineInfo

	return nil
}

// UpdateEngine updates an existing search engine's quota information
func (sqm *SearchQuotaManager) UpdateEngine(engineName string, engineInfo *EngineQuotaInfo) error {
	sqm.mu.Lock()
	defer sqm.mu.Unlock()

	if _, exists := sqm.engines[engineName]; !exists {
		return fmt.Errorf("engine %s not found", engineName)
	}

	engineInfo.EngineName = engineName
	sqm.engines[engineName] = engineInfo

	return nil
}

// RemoveEngine removes a search engine from the quota manager
func (sqm *SearchQuotaManager) RemoveEngine(engineName string) error {
	sqm.mu.Lock()
	defer sqm.mu.Unlock()

	if _, exists := sqm.engines[engineName]; !exists {
		return fmt.Errorf("engine %s not found", engineName)
	}

	delete(sqm.engines, engineName)

	return nil
}

// EnableEngine enables a search engine
func (sqm *SearchQuotaManager) EnableEngine(engineName string) error {
	sqm.mu.Lock()
	defer sqm.mu.Unlock()

	engineInfo, exists := sqm.engines[engineName]
	if !exists {
		return fmt.Errorf("engine %s not found", engineName)
	}

	engineInfo.IsEnabled = true

	return nil
}

// DisableEngine disables a search engine
func (sqm *SearchQuotaManager) DisableEngine(engineName string) error {
	sqm.mu.Lock()
	defer sqm.mu.Unlock()

	engineInfo, exists := sqm.engines[engineName]
	if !exists {
		return fmt.Errorf("engine %s not found", engineName)
	}

	engineInfo.IsEnabled = false

	return nil
}

// ResetQuotas resets quotas for all engines
func (sqm *SearchQuotaManager) ResetQuotas() {
	sqm.mu.Lock()
	defer sqm.mu.Unlock()

	now := time.Now()

	// Reset global quotas
	sqm.globalQuota.TotalDailyQuotaUsed = 0
	sqm.globalQuota.TotalHourlyQuotaUsed = 0
	sqm.globalQuota.TotalMinuteQuotaUsed = 0
	sqm.globalQuota.TotalConcurrentRequests = 0
	sqm.globalQuota.LastGlobalQuotaReset = now
	sqm.globalQuota.LastGlobalHourlyReset = now
	sqm.globalQuota.LastGlobalMinuteReset = now

	// Reset engine quotas
	for _, engine := range sqm.engines {
		engine.DailyQuotaUsed = 0
		engine.HourlyQuotaUsed = 0
		engine.MinuteQuotaUsed = 0
		engine.ConcurrentRequests = 0
		engine.LastQuotaReset = now
		engine.LastHourlyReset = now
		engine.LastMinuteReset = now
	}
}

// GetAvailableEngines returns a list of available engines sorted by priority
func (sqm *SearchQuotaManager) GetAvailableEngines() []string {
	sqm.mu.RLock()
	defer sqm.mu.RUnlock()

	var availableEngines []string
	for name, engine := range sqm.engines {
		if engine.IsEnabled && sqm.hasQuotaRemaining(engine) {
			availableEngines = append(availableEngines, name)
		}
	}

	// Sort by priority (lower number = higher priority)
	// This is a simple implementation - in a real system, you might want more sophisticated sorting
	return availableEngines
}

// GetFallbackEngine returns the best fallback engine for a given engine
func (sqm *SearchQuotaManager) GetFallbackEngine(engineName string) string {
	sqm.mu.RLock()
	defer sqm.mu.RUnlock()

	engineInfo, exists := sqm.engines[engineName]
	if !exists {
		return ""
	}

	// Check fallback engines in order
	for _, fallbackName := range engineInfo.FallbackEngines {
		fallbackEngine, exists := sqm.engines[fallbackName]
		if exists && fallbackEngine.IsEnabled && sqm.hasQuotaRemaining(fallbackEngine) {
			return fallbackName
		}
	}

	return ""
}

// UpdateConfig updates the quota manager configuration
func (sqm *SearchQuotaManager) UpdateConfig(config QuotaManagerConfig) {
	sqm.mu.Lock()
	defer sqm.mu.Unlock()
	sqm.config = config
}

// GetConfig returns the current configuration
func (sqm *SearchQuotaManager) GetConfig() QuotaManagerConfig {
	sqm.mu.RLock()
	defer sqm.mu.RUnlock()
	return sqm.config
}

// resetQuotasIfNeeded resets quotas if the reset interval has passed
func (sqm *SearchQuotaManager) resetQuotasIfNeeded(engineInfo *EngineQuotaInfo) {
	now := time.Now()

	// Reset daily quota
	if now.Sub(engineInfo.LastQuotaReset) >= 24*time.Hour {
		engineInfo.DailyQuotaUsed = 0
		engineInfo.LastQuotaReset = now
	}

	// Reset hourly quota
	if now.Sub(engineInfo.LastHourlyReset) >= time.Hour {
		engineInfo.HourlyQuotaUsed = 0
		engineInfo.LastHourlyReset = now
	}

	// Reset minute quota
	if now.Sub(engineInfo.LastMinuteReset) >= time.Minute {
		engineInfo.MinuteQuotaUsed = 0
		engineInfo.LastMinuteReset = now
	}

	// Reset global quotas
	if now.Sub(sqm.globalQuota.LastGlobalQuotaReset) >= 24*time.Hour {
		sqm.globalQuota.TotalDailyQuotaUsed = 0
		sqm.globalQuota.LastGlobalQuotaReset = now
	}

	if now.Sub(sqm.globalQuota.LastGlobalHourlyReset) >= time.Hour {
		sqm.globalQuota.TotalHourlyQuotaUsed = 0
		sqm.globalQuota.LastGlobalHourlyReset = now
	}

	if now.Sub(sqm.globalQuota.LastGlobalMinuteReset) >= time.Minute {
		sqm.globalQuota.TotalMinuteQuotaUsed = 0
		sqm.globalQuota.LastGlobalMinuteReset = now
	}
}

// checkGlobalQuota checks if global quota allows the request
func (sqm *SearchQuotaManager) checkGlobalQuota() bool {
	return sqm.globalQuota.TotalDailyQuotaUsed < sqm.globalQuota.TotalDailyQuotaLimit &&
		sqm.globalQuota.TotalHourlyQuotaUsed < sqm.globalQuota.TotalHourlyQuotaLimit &&
		sqm.globalQuota.TotalMinuteQuotaUsed < sqm.globalQuota.TotalMinuteQuotaLimit
}

// checkEngineQuota checks if engine-specific quota allows the request
func (sqm *SearchQuotaManager) checkEngineQuota(engineInfo *EngineQuotaInfo) bool {
	return engineInfo.DailyQuotaUsed < engineInfo.DailyQuotaLimit &&
		engineInfo.HourlyQuotaUsed < engineInfo.HourlyQuotaLimit &&
		engineInfo.MinuteQuotaUsed < engineInfo.MinuteQuotaLimit
}

// checkConcurrentRequests checks if concurrent request limits allow the request
func (sqm *SearchQuotaManager) checkConcurrentRequests(engineInfo *EngineQuotaInfo) bool {
	return engineInfo.ConcurrentRequests < engineInfo.MaxConcurrentRequests &&
		sqm.globalQuota.TotalConcurrentRequests < sqm.globalQuota.MaxTotalConcurrentRequests
}

// incrementUsage increments usage counters
func (sqm *SearchQuotaManager) incrementUsage(engineInfo *EngineQuotaInfo) {
	engineInfo.DailyQuotaUsed++
	engineInfo.HourlyQuotaUsed++
	engineInfo.MinuteQuotaUsed++
	engineInfo.ConcurrentRequests++

	sqm.globalQuota.TotalDailyQuotaUsed++
	sqm.globalQuota.TotalHourlyQuotaUsed++
	sqm.globalQuota.TotalMinuteQuotaUsed++
	sqm.globalQuota.TotalConcurrentRequests++
}

// hasQuotaRemaining checks if an engine has quota remaining
func (sqm *SearchQuotaManager) hasQuotaRemaining(engineInfo *EngineQuotaInfo) bool {
	return engineInfo.DailyQuotaUsed < engineInfo.DailyQuotaLimit &&
		engineInfo.HourlyQuotaUsed < engineInfo.HourlyQuotaLimit &&
		engineInfo.MinuteQuotaUsed < engineInfo.MinuteQuotaLimit &&
		engineInfo.ConcurrentRequests < engineInfo.MaxConcurrentRequests
}

// checkForAlerts checks if quota usage triggers alerts
func (sqm *SearchQuotaManager) checkForAlerts(engineInfo *EngineQuotaInfo) {
	if !sqm.config.EnableQuotaAlerts {
		return
	}

	// Check daily quota alert
	dailyPercentage := float64(engineInfo.DailyQuotaUsed) / float64(engineInfo.DailyQuotaLimit)
	if dailyPercentage >= sqm.config.AlertThreshold {
		// In a real implementation, you would send an alert here
		fmt.Printf("ALERT: Engine %s daily quota usage at %.1f%%\n", engineInfo.EngineName, dailyPercentage*100)
	}

	// Check hourly quota alert
	hourlyPercentage := float64(engineInfo.HourlyQuotaUsed) / float64(engineInfo.HourlyQuotaLimit)
	if hourlyPercentage >= sqm.config.AlertThreshold {
		fmt.Printf("ALERT: Engine %s hourly quota usage at %.1f%%\n", engineInfo.EngineName, hourlyPercentage*100)
	}
}

// createAllowedResponse creates a response for an allowed request
func (sqm *SearchQuotaManager) createAllowedResponse(req *QuotaRequest) *QuotaResponse {
	engineInfo := sqm.engines[req.EngineName]

	return &QuotaResponse{
		RequestID:             req.RequestID,
		EngineName:            req.EngineName,
		IsAllowed:             true,
		WaitTime:              0,
		QuotaRemaining:        engineInfo.DailyQuotaLimit - engineInfo.DailyQuotaUsed,
		QuotaUsed:             engineInfo.DailyQuotaUsed,
		QuotaLimit:            engineInfo.DailyQuotaLimit,
		ConcurrentRequests:    engineInfo.ConcurrentRequests,
		MaxConcurrentRequests: engineInfo.MaxConcurrentRequests,
		NextResetTime:         engineInfo.LastQuotaReset.Add(24 * time.Hour),
		Warnings:              []string{},
		Errors:                []string{},
		Metadata:              req.Metadata,
	}
}

// createDeniedResponse creates a response for a denied request
func (sqm *SearchQuotaManager) createDeniedResponse(req *QuotaRequest, reason string) *QuotaResponse {
	engineInfo := sqm.engines[req.EngineName]

	return &QuotaResponse{
		RequestID:             req.RequestID,
		EngineName:            req.EngineName,
		IsAllowed:             false,
		WaitTime:              sqm.config.RetryDelay,
		QuotaRemaining:        engineInfo.DailyQuotaLimit - engineInfo.DailyQuotaUsed,
		QuotaUsed:             engineInfo.DailyQuotaUsed,
		QuotaLimit:            engineInfo.DailyQuotaLimit,
		ConcurrentRequests:    engineInfo.ConcurrentRequests,
		MaxConcurrentRequests: engineInfo.MaxConcurrentRequests,
		NextResetTime:         engineInfo.LastQuotaReset.Add(24 * time.Hour),
		Warnings:              []string{},
		Errors:                []string{reason},
		Metadata:              req.Metadata,
	}
}
