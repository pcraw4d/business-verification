package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewFeatureFlagManager(t *testing.T) {
	fm := NewFeatureFlagManager("test")

	if fm == nil {
		t.Fatal("Expected FeatureFlagManager to be created")
	}

	if fm.env != "test" {
		t.Errorf("Expected environment to be 'test', got '%s'", fm.env)
	}

	// Check that default flags are loaded
	flags := fm.GetAllFlags()
	if len(flags) == 0 {
		t.Error("Expected default flags to be loaded")
	}

	// Check for specific default flags
	expectedFlags := []string{
		"modular_architecture",
		"intelligent_routing",
		"enhanced_classification",
		"legacy_compatibility",
		"a_b_testing",
		"performance_monitoring",
		"graceful_degradation",
	}

	for _, flagName := range expectedFlags {
		if _, exists := flags[flagName]; !exists {
			t.Errorf("Expected flag '%s' to be loaded", flagName)
		}
	}
}

func TestFeatureFlagManager_IsEnabled(t *testing.T) {
	fm := NewFeatureFlagManager("test")

	// Test existing flag
	if !fm.IsEnabled("legacy_compatibility") {
		t.Error("Expected legacy_compatibility to be enabled by default")
	}

	// Test non-existent flag
	if fm.IsEnabled("non_existent_flag") {
		t.Error("Expected non-existent flag to be disabled")
	}

	// Test disabled flag
	flag := &FeatureFlag{
		Name:    "test_flag",
		Enabled: false,
	}
	fm.SetFlag(flag)

	if fm.IsEnabled("test_flag") {
		t.Error("Expected disabled flag to return false")
	}
}

func TestFeatureFlagManager_IsEnabledForPercentage(t *testing.T) {
	fm := NewFeatureFlagManager("test")

	// Test 100% rollout
	flag := &FeatureFlag{
		Name:       "test_flag_100",
		Enabled:    true,
		Percentage: 100,
	}
	fm.SetFlag(flag)

	if !fm.IsEnabledForPercentage("test_flag_100", "test_request") {
		t.Error("Expected 100% rollout to always be enabled")
	}

	// Test 0% rollout
	flag = &FeatureFlag{
		Name:       "test_flag_0",
		Enabled:    true,
		Percentage: 0,
	}
	fm.SetFlag(flag)

	if fm.IsEnabledForPercentage("test_flag_0", "test_request") {
		t.Error("Expected 0% rollout to never be enabled")
	}

	// Test 50% rollout with consistent request ID
	flag = &FeatureFlag{
		Name:       "test_flag_50",
		Enabled:    true,
		Percentage: 50,
	}
	fm.SetFlag(flag)

	// Same request ID should always get the same result
	result1 := fm.IsEnabledForPercentage("test_flag_50", "consistent_request")
	result2 := fm.IsEnabledForPercentage("test_flag_50", "consistent_request")

	if result1 != result2 {
		t.Error("Expected consistent results for same request ID")
	}
}

func TestFeatureFlagManager_SetFlag(t *testing.T) {
	fm := NewFeatureFlagManager("test")

	// Test valid flag
	flag := &FeatureFlag{
		Name:        "test_flag",
		Description: "Test flag",
		Enabled:     true,
		Percentage:  50,
	}

	err := fm.SetFlag(flag)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test invalid percentage
	flag = &FeatureFlag{
		Name:       "test_flag_invalid",
		Percentage: 150,
	}

	err = fm.SetFlag(flag)
	if err == nil {
		t.Error("Expected error for invalid percentage")
	}

	// Test empty name
	flag = &FeatureFlag{
		Name:       "",
		Percentage: 50,
	}

	err = fm.SetFlag(flag)
	if err == nil {
		t.Error("Expected error for empty name")
	}
}

func TestFeatureFlagManager_GetFlag(t *testing.T) {
	fm := NewFeatureFlagManager("test")

	// Test existing flag
	flag, err := fm.GetFlag("legacy_compatibility")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if flag.Name != "legacy_compatibility" {
		t.Errorf("Expected flag name 'legacy_compatibility', got '%s'", flag.Name)
	}

	// Test non-existent flag
	_, err = fm.GetFlag("non_existent_flag")
	if err == nil {
		t.Error("Expected error for non-existent flag")
	}
}

func TestFeatureFlagManager_EnableDisableFlag(t *testing.T) {
	fm := NewFeatureFlagManager("test")

	// Create a disabled flag
	flag := &FeatureFlag{
		Name:    "test_flag",
		Enabled: false,
	}
	fm.SetFlag(flag)

	// Test initial state
	if fm.IsEnabled("test_flag") {
		t.Error("Expected flag to be disabled initially")
	}

	// Enable flag
	err := fm.EnableFlag("test_flag")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !fm.IsEnabled("test_flag") {
		t.Error("Expected flag to be enabled after EnableFlag")
	}

	// Disable flag
	err = fm.DisableFlag("test_flag")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if fm.IsEnabled("test_flag") {
		t.Error("Expected flag to be disabled after DisableFlag")
	}
}

func TestFeatureFlagManager_SetPercentage(t *testing.T) {
	fm := NewFeatureFlagManager("test")

	// Create a flag
	flag := &FeatureFlag{
		Name:       "test_flag",
		Enabled:    true,
		Percentage: 0,
	}
	fm.SetFlag(flag)

	// Set percentage
	err := fm.SetPercentage("test_flag", 75)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify percentage was set
	flag, err = fm.GetFlag("test_flag")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if flag.Percentage != 75 {
		t.Errorf("Expected percentage 75, got %d", flag.Percentage)
	}

	// Test invalid percentage
	err = fm.SetPercentage("test_flag", 150)
	if err == nil {
		t.Error("Expected error for invalid percentage")
	}

	err = fm.SetPercentage("test_flag", -10)
	if err == nil {
		t.Error("Expected error for negative percentage")
	}
}

func TestFeatureFlagManager_DeleteFlag(t *testing.T) {
	fm := NewFeatureFlagManager("test")

	// Create a flag
	flag := &FeatureFlag{
		Name:    "test_flag",
		Enabled: true,
	}
	fm.SetFlag(flag)

	// Verify flag exists
	if !fm.IsEnabled("test_flag") {
		t.Error("Expected flag to exist before deletion")
	}

	// Delete flag
	err := fm.DeleteFlag("test_flag")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify flag is deleted
	if fm.IsEnabled("test_flag") {
		t.Error("Expected flag to be deleted")
	}

	// Test deleting non-existent flag
	err = fm.DeleteFlag("non_existent_flag")
	if err == nil {
		t.Error("Expected error for deleting non-existent flag")
	}
}

func TestFeatureFlagManager_ShouldUseModularArchitecture(t *testing.T) {
	fm := NewFeatureFlagManager("test")

	// Test with modular architecture disabled
	if fm.ShouldUseModularArchitecture(context.Background(), "test_request") {
		t.Error("Expected modular architecture to be disabled by default")
	}

	// Enable modular architecture
	fm.EnableFlag("modular_architecture")
	fm.SetPercentage("modular_architecture", 100)

	// Test with intelligent routing disabled
	if fm.ShouldUseModularArchitecture(context.Background(), "test_request") {
		t.Error("Expected modular architecture to be disabled when intelligent routing is disabled")
	}

	// Enable intelligent routing
	fm.EnableFlag("intelligent_routing")
	fm.SetPercentage("intelligent_routing", 100)

	// Now should be enabled
	if !fm.ShouldUseModularArchitecture(context.Background(), "test_request") {
		t.Error("Expected modular architecture to be enabled when both flags are enabled")
	}
}

func TestFeatureFlagManager_ShouldUseLegacyImplementation(t *testing.T) {
	fm := NewFeatureFlagManager("test")

	// Test with legacy compatibility enabled (default)
	if !fm.ShouldUseLegacyImplementation(context.Background(), "test_request") {
		t.Error("Expected legacy implementation to be available when compatibility is enabled")
	}

	// Disable legacy compatibility
	fm.DisableFlag("legacy_compatibility")

	// Test with modular architecture disabled
	if !fm.ShouldUseLegacyImplementation(context.Background(), "test_request") {
		t.Error("Expected legacy implementation when modular architecture is disabled")
	}

	// Enable modular architecture
	fm.EnableFlag("modular_architecture")
	fm.SetPercentage("modular_architecture", 100)
	fm.EnableFlag("intelligent_routing")
	fm.SetPercentage("intelligent_routing", 100)

	// Now should not use legacy
	if fm.ShouldUseLegacyImplementation(context.Background(), "test_request") {
		t.Error("Expected not to use legacy implementation when modular architecture is enabled")
	}
}

func TestFeatureFlagManager_FeatureFlagContext(t *testing.T) {
	fm := NewFeatureFlagManager("test")

	// Enable some flags
	fm.EnableFlag("modular_architecture")
	fm.SetPercentage("modular_architecture", 100)
	fm.EnableFlag("intelligent_routing")
	fm.SetPercentage("intelligent_routing", 100)

	ctx := context.Background()
	requestID := "test_request_123"

	// Add feature flags to context
	ctx = fm.FeatureFlagContext(ctx, requestID)

	// Retrieve flags from context
	flags := GetFeatureFlagsFromContext(ctx)
	if len(flags) == 0 {
		t.Error("Expected flags to be in context")
	}

	// Check specific flags
	if !flags["modular_architecture"] {
		t.Error("Expected modular_architecture to be enabled in context")
	}

	if !flags["intelligent_routing"] {
		t.Error("Expected intelligent_routing to be enabled in context")
	}

	// Check request ID
	retrievedRequestID := GetRequestIDFromContext(ctx)
	if retrievedRequestID != requestID {
		t.Errorf("Expected request ID '%s', got '%s'", requestID, retrievedRequestID)
	}
}

func TestFeatureFlagManager_EnvironmentVariables(t *testing.T) {
	// Test environment variable parsing
	os.Setenv("ENABLE_MODULAR_ARCHITECTURE", "true")
	os.Setenv("MODULAR_ARCHITECTURE_PERCENTAGE", "75")
	os.Setenv("ENABLE_AB_TESTING", "false")

	fm := NewFeatureFlagManager("test")

	// Check that environment variables are respected
	flag, err := fm.GetFlag("modular_architecture")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !flag.Enabled {
		t.Error("Expected modular_architecture to be enabled from environment")
	}

	if flag.Percentage != 75 {
		t.Errorf("Expected percentage 75, got %d", flag.Percentage)
	}

	// Clean up
	os.Unsetenv("ENABLE_MODULAR_ARCHITECTURE")
	os.Unsetenv("MODULAR_ARCHITECTURE_PERCENTAGE")
	os.Unsetenv("ENABLE_AB_TESTING")
}

func TestFeatureFlagManager_GetRolloutStatus(t *testing.T) {
	fm := NewFeatureFlagManager("test")

	// Enable and configure a flag
	fm.EnableFlag("modular_architecture")
	fm.SetPercentage("modular_architecture", 50)

	status := fm.GetRolloutStatus()

	modularStatus, exists := status["modular_architecture"]
	if !exists {
		t.Error("Expected modular_architecture status to exist")
	}

	statusMap, ok := modularStatus.(map[string]interface{})
	if !ok {
		t.Error("Expected status to be a map")
	}

	if !statusMap["enabled"].(bool) {
		t.Error("Expected modular_architecture to be enabled in status")
	}

	if statusMap["percentage"].(int) != 50 {
		t.Error("Expected percentage to be 50 in status")
	}
}

func TestFeatureFlagManager_ExpiredFlag(t *testing.T) {
	fm := NewFeatureFlagManager("test")

	// Create a flag with expiration
	endTime := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
	flag := &FeatureFlag{
		Name:    "expired_flag",
		Enabled: true,
		EndTime: &endTime,
	}
	fm.SetFlag(flag)

	// Flag should be disabled due to expiration
	if fm.IsEnabled("expired_flag") {
		t.Error("Expected expired flag to be disabled")
	}

	// Create a flag with future expiration
	endTime = time.Now().Add(1 * time.Hour) // Expires in 1 hour
	flag = &FeatureFlag{
		Name:    "future_flag",
		Enabled: true,
		EndTime: &endTime,
	}
	fm.SetFlag(flag)

	// Flag should be enabled
	if !fm.IsEnabled("future_flag") {
		t.Error("Expected future flag to be enabled")
	}
}

func TestFeatureFlagManager_ConcurrentAccess(t *testing.T) {
	fm := NewFeatureFlagManager("test")

	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				fm.IsEnabled("legacy_compatibility")
				fm.GetAllFlags()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent writes
	for i := 0; i < 10; i++ {
		go func(id int) {
			flag := &FeatureFlag{
				Name:       fmt.Sprintf("concurrent_flag_%d", id),
				Enabled:    true,
				Percentage: 50,
			}
			fm.SetFlag(flag)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all flags were set
	flags := fm.GetAllFlags()
	concurrentFlags := 0
	for name := range flags {
		if strings.HasPrefix(name, "concurrent_flag_") {
			concurrentFlags++
		}
	}

	if concurrentFlags != 10 {
		t.Errorf("Expected 10 concurrent flags, got %d", concurrentFlags)
	}
}
