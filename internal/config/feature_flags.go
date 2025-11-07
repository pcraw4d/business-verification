package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// FeatureFlag represents a feature flag configuration
type FeatureFlag struct {
	Name        string
	Description string
	Enabled     bool
	Percentage  int // 0-100 for gradual rollout
	StartTime   time.Time
	EndTime     *time.Time // nil means no end time
	Metadata    map[string]interface{}
}

// FeatureFlagManager manages feature flags for the application
type FeatureFlagManager struct {
	flags map[string]*FeatureFlag
	mu    sync.RWMutex
	env   string
}

// NewFeatureFlagManager creates a new feature flag manager
func NewFeatureFlagManager(env string) *FeatureFlagManager {
	fm := &FeatureFlagManager{
		flags: make(map[string]*FeatureFlag),
		env:   env,
	}
	fm.loadDefaultFlags()
	return fm
}

// loadDefaultFlags loads default feature flags based on environment
func (fm *FeatureFlagManager) loadDefaultFlags() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Default flags for modular architecture rollout
	defaultFlags := map[string]*FeatureFlag{
		"modular_architecture": {
			Name:        "modular_architecture",
			Description: "Enable new modular architecture for classification",
			Enabled:     fm.getEnvBool("ENABLE_MODULAR_ARCHITECTURE", false),
			Percentage:  fm.getEnvInt("MODULAR_ARCHITECTURE_PERCENTAGE", 0),
			StartTime:   time.Now(),
			Metadata: map[string]interface{}{
				"modules": []string{"keyword_classification", "ml_classification", "website_analysis", "web_search_analysis"},
			},
		},
		"intelligent_routing": {
			Name:        "intelligent_routing",
			Description: "Enable intelligent routing system for module selection",
			Enabled:     fm.getEnvBool("ENABLE_INTELLIGENT_ROUTING", false),
			Percentage:  fm.getEnvInt("INTELLIGENT_ROUTING_PERCENTAGE", 0),
			StartTime:   time.Now(),
			Metadata: map[string]interface{}{
				"routing_strategy": "performance_based",
			},
		},
		"enhanced_classification": {
			Name:        "enhanced_classification",
			Description: "Enable enhanced classification with all modules",
			Enabled:     fm.getEnvBool("ENABLE_ENHANCED_CLASSIFICATION", false),
			Percentage:  fm.getEnvInt("ENHANCED_CLASSIFICATION_PERCENTAGE", 0),
			StartTime:   time.Now(),
			Metadata: map[string]interface{}{
				"features": []string{"verification", "risk_assessment", "data_extraction"},
			},
		},
		"legacy_compatibility": {
			Name:        "legacy_compatibility",
			Description: "Enable backward compatibility with legacy API endpoints",
			Enabled:     fm.getEnvBool("ENABLE_LEGACY_COMPATIBILITY", true),
			Percentage:  100, // Always enabled for legacy compatibility
			StartTime:   time.Now(),
			Metadata: map[string]interface{}{
				"deprecation_warning": true,
			},
		},
		"a_b_testing": {
			Name:        "a_b_testing",
			Description: "Enable A/B testing for new vs legacy implementations",
			Enabled:     fm.getEnvBool("ENABLE_AB_TESTING", false),
			Percentage:  fm.getEnvInt("AB_TESTING_PERCENTAGE", 10),
			StartTime:   time.Now(),
			Metadata: map[string]interface{}{
				"test_duration": "7d",
				"metrics":       []string{"response_time", "accuracy", "user_satisfaction"},
			},
		},
		"performance_monitoring": {
			Name:        "performance_monitoring",
			Description: "Enable enhanced performance monitoring for modules",
			Enabled:     fm.getEnvBool("ENABLE_PERFORMANCE_MONITORING", true),
			Percentage:  100,
			StartTime:   time.Now(),
			Metadata: map[string]interface{}{
				"metrics": []string{"response_time", "throughput", "error_rate", "resource_usage"},
			},
		},
		"graceful_degradation": {
			Name:        "graceful_degradation",
			Description: "Enable graceful degradation when modules fail",
			Enabled:     fm.getEnvBool("ENABLE_GRACEFUL_DEGRADATION", true),
			Percentage:  100,
			StartTime:   time.Now(),
			Metadata: map[string]interface{}{
				"fallback_strategy": "legacy_implementation",
			},
		},
		// Feature flags for incomplete features - disable in production
		"incomplete_risk_benchmarks": {
			Name:        "incomplete_risk_benchmarks",
			Description: "Enable risk benchmarks endpoint (may be incomplete)",
			Enabled:     fm.getEnvBool("ENABLE_INCOMPLETE_RISK_BENCHMARKS", fm.env != "production"),
			Percentage:  100,
			StartTime:   time.Now(),
			Metadata: map[string]interface{}{
				"status":     "incomplete",
				"todo":       "Complete database queries for benchmarks",
				"production": false,
			},
		},
		"incomplete_risk_predictions": {
			Name:        "incomplete_risk_predictions",
			Description: "Enable risk predictions endpoint (may be incomplete)",
			Enabled:     fm.getEnvBool("ENABLE_INCOMPLETE_RISK_PREDICTIONS", fm.env != "production"),
			Percentage:  100,
			StartTime:   time.Now(),
			Metadata: map[string]interface{}{
				"status":     "incomplete",
				"todo":       "Complete ML service integration",
				"production": false,
			},
		},
		"incomplete_merchant_analytics": {
			Name:        "incomplete_merchant_analytics",
			Description: "Enable merchant analytics endpoint (may be incomplete)",
			Enabled:     fm.getEnvBool("ENABLE_INCOMPLETE_MERCHANT_ANALYTICS", fm.env != "production"),
			Percentage:  100,
			StartTime:   time.Now(),
			Metadata: map[string]interface{}{
				"status":     "incomplete",
				"todo":       "Complete analytics data aggregation",
				"production": false,
			},
		},
	}

	// Load flags from environment
	for name, flag := range defaultFlags {
		fm.flags[name] = flag
	}
}

// IsEnabled checks if a feature flag is enabled
func (fm *FeatureFlagManager) IsEnabled(flagName string) bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	flag, exists := fm.flags[flagName]
	if !exists {
		return false
	}

	if !flag.Enabled {
		return false
	}

	// Check if flag has expired
	if flag.EndTime != nil && time.Now().After(*flag.EndTime) {
		return false
	}

	return true
}

// IsEnabledForPercentage checks if a feature flag is enabled for a specific percentage of requests
func (fm *FeatureFlagManager) IsEnabledForPercentage(flagName string, requestID string) bool {
	if !fm.IsEnabled(flagName) {
		return false
	}

	fm.mu.RLock()
	defer fm.mu.RUnlock()

	flag, exists := fm.flags[flagName]
	if !exists {
		return false
	}

	// If percentage is 100, always enable
	if flag.Percentage >= 100 {
		return true
	}

	// If percentage is 0, never enable
	if flag.Percentage <= 0 {
		return false
	}

	// Simple hash-based percentage calculation
	hash := fm.hashString(requestID)
	percentage := hash % 100

	return percentage < flag.Percentage
}

// GetFlag retrieves a feature flag by name
func (fm *FeatureFlagManager) GetFlag(flagName string) (*FeatureFlag, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	flag, exists := fm.flags[flagName]
	if !exists {
		return nil, fmt.Errorf("feature flag '%s' not found", flagName)
	}

	return flag, nil
}

// SetFlag sets or updates a feature flag
func (fm *FeatureFlagManager) SetFlag(flag *FeatureFlag) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if flag.Name == "" {
		return fmt.Errorf("feature flag name cannot be empty")
	}

	if flag.Percentage < 0 || flag.Percentage > 100 {
		return fmt.Errorf("feature flag percentage must be between 0 and 100")
	}

	fm.flags[flag.Name] = flag
	return nil
}

// GetAllFlags returns all feature flags
func (fm *FeatureFlagManager) GetAllFlags() map[string]*FeatureFlag {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	result := make(map[string]*FeatureFlag)
	for name, flag := range fm.flags {
		result[name] = flag
	}
	return result
}

// DeleteFlag removes a feature flag
func (fm *FeatureFlagManager) DeleteFlag(flagName string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if _, exists := fm.flags[flagName]; !exists {
		return fmt.Errorf("feature flag '%s' not found", flagName)
	}

	delete(fm.flags, flagName)
	return nil
}

// EnableFlag enables a feature flag
func (fm *FeatureFlagManager) EnableFlag(flagName string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	flag, exists := fm.flags[flagName]
	if !exists {
		return fmt.Errorf("feature flag '%s' not found", flagName)
	}

	flag.Enabled = true
	return nil
}

// DisableFlag disables a feature flag
func (fm *FeatureFlagManager) DisableFlag(flagName string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	flag, exists := fm.flags[flagName]
	if !exists {
		return fmt.Errorf("feature flag '%s' not found", flagName)
	}

	flag.Enabled = false
	return nil
}

// SetPercentage sets the rollout percentage for a feature flag
func (fm *FeatureFlagManager) SetPercentage(flagName string, percentage int) error {
	if percentage < 0 || percentage > 100 {
		return fmt.Errorf("percentage must be between 0 and 100")
	}

	fm.mu.Lock()
	defer fm.mu.Unlock()

	flag, exists := fm.flags[flagName]
	if !exists {
		return fmt.Errorf("feature flag '%s' not found", flagName)
	}

	flag.Percentage = percentage
	return nil
}

// GetRolloutStatus returns the current rollout status for all flags
func (fm *FeatureFlagManager) GetRolloutStatus() map[string]interface{} {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	status := make(map[string]interface{})
	for name, flag := range fm.flags {
		status[name] = map[string]interface{}{
			"enabled":    flag.Enabled,
			"percentage": flag.Percentage,
			"start_time": flag.StartTime,
			"end_time":   flag.EndTime,
			"metadata":   flag.Metadata,
		}
	}
	return status
}

// ShouldUseModularArchitecture determines if the request should use the new modular architecture
func (fm *FeatureFlagManager) ShouldUseModularArchitecture(ctx context.Context, requestID string) bool {
	// Check if modular architecture is enabled
	if !fm.IsEnabledForPercentage("modular_architecture", requestID) {
		return false
	}

	// Check if intelligent routing is enabled
	if !fm.IsEnabledForPercentage("intelligent_routing", requestID) {
		return false
	}

	return true
}

// ShouldUseLegacyImplementation determines if the request should fall back to legacy implementation
func (fm *FeatureFlagManager) ShouldUseLegacyImplementation(ctx context.Context, requestID string) bool {
	// Always allow legacy if compatibility is enabled
	if fm.IsEnabled("legacy_compatibility") {
		return true
	}

	// Use legacy if modular architecture is not enabled
	return !fm.ShouldUseModularArchitecture(ctx, requestID)
}

// ShouldEnableABTesting determines if A/B testing should be enabled for the request
func (fm *FeatureFlagManager) ShouldEnableABTesting(ctx context.Context, requestID string) bool {
	return fm.IsEnabledForPercentage("a_b_testing", requestID)
}

// ShouldEnableGracefulDegradation determines if graceful degradation should be enabled
func (fm *FeatureFlagManager) ShouldEnableGracefulDegradation(ctx context.Context) bool {
	return fm.IsEnabled("graceful_degradation")
}

// ShouldEnablePerformanceMonitoring determines if performance monitoring should be enabled
func (fm *FeatureFlagManager) ShouldEnablePerformanceMonitoring(ctx context.Context) bool {
	return fm.IsEnabled("performance_monitoring")
}

// Helper methods for environment variable parsing
func (fm *FeatureFlagManager) getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

func (fm *FeatureFlagManager) getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// Simple hash function for percentage-based rollout
func (fm *FeatureFlagManager) hashString(s string) int {
	hash := 0
	for _, char := range s {
		hash = ((hash << 5) - hash) + int(char)
		hash = hash & hash // Convert to 32-bit integer
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// FeatureFlagContext adds feature flag information to context
func (fm *FeatureFlagManager) FeatureFlagContext(ctx context.Context, requestID string) context.Context {
	flags := make(map[string]bool)

	// Check all relevant flags
	flagNames := []string{
		"modular_architecture",
		"intelligent_routing",
		"enhanced_classification",
		"legacy_compatibility",
		"a_b_testing",
		"graceful_degradation",
		"performance_monitoring",
	}

	for _, flagName := range flagNames {
		flags[flagName] = fm.IsEnabledForPercentage(flagName, requestID)
	}

	// Add to context
	ctx = context.WithValue(ctx, "feature_flags", flags)
	ctx = context.WithValue(ctx, "request_id", requestID)

	return ctx
}

// GetFeatureFlagsFromContext retrieves feature flags from context
func GetFeatureFlagsFromContext(ctx context.Context) map[string]bool {
	if flags, ok := ctx.Value("feature_flags").(map[string]bool); ok {
		return flags
	}
	return make(map[string]bool)
}

// GetRequestIDFromContext retrieves request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}
