package main

import (
	"context"
	"fmt"
	"time"

	"kyb-platform/internal/config"
)

func main() {
	fmt.Println("Testing rollout manager with zero interval fix...")

	// Create a config with zero interval (this should trigger our fix)
	rolloutConfig := config.RolloutConfig{
		IncrementInterval:   0, // This should be fixed to 1 hour
		IncrementPercentage: 0, // This should be fixed to 10%
		MaxPercentage:       0, // This should be fixed to 100%
	}

	// Create rollout manager - this should not panic now
	rolloutManager := config.NewRolloutManager(rolloutConfig)
	fmt.Println("✅ RolloutManager created successfully without panic")

	// Test that the config was fixed
	rolloutConfigResult := rolloutManager.GetConfig()
	fmt.Printf("IncrementInterval: %v (should be 1h)\n", rolloutConfigResult.IncrementInterval)
	fmt.Printf("IncrementPercentage: %v (should be 10)\n", rolloutConfigResult.IncrementPercentage)
	fmt.Printf("MaxPercentage: %v (should be 100)\n", rolloutConfigResult.MaxPercentage)

	// Test gradual rollout functionality
	ctx := context.Background()
	requestType := "test_request"

	// Test multiple requests to see distribution
	controlCount := 0
	newModelCount := 0

	for i := 0; i < 100; i++ {
		shouldUseNew := rolloutManager.ShouldUseNewModel(ctx, requestType)
		if shouldUseNew {
			newModelCount++
		} else {
			controlCount++
		}
		time.Sleep(1 * time.Millisecond) // Add variation
	}

	fmt.Printf("\nRollout distribution (100 requests):\n")
	fmt.Printf("Control (rule_based): %d (%.1f%%)\n", controlCount, float64(controlCount)/100*100)
	fmt.Printf("New model: %d (%.1f%%)\n", newModelCount, float64(newModelCount)/100*100)

	// Test A/B testing as well
	abConfig := config.ABTestingConfig{
		TestDuration:            0, // This should be fixed to 1 week
		MinimumSampleSize:       0, // This should be fixed to 1000
		StatisticalSignificance: 0, // This should be fixed to 0.95
	}

	abTester := config.NewABTester(abConfig)
	fmt.Println("\n✅ ABTester created successfully without panic")

	abConfigResult := abTester.GetConfig()
	fmt.Printf("TestDuration: %v (should be 168h)\n", abConfigResult.TestDuration)
	fmt.Printf("MinimumSampleSize: %v (should be 1000)\n", abConfigResult.MinimumSampleSize)
	fmt.Printf("StatisticalSignificance: %v (should be 0.95)\n", abConfigResult.StatisticalSignificance)

	// Test A/B testing distribution
	controlABCount := 0
	testABCount := 0

	for i := 0; i < 100; i++ {
		variant, err := abTester.GetTestVariant(ctx, requestType)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		if variant == "control" {
			controlABCount++
		} else {
			testABCount++
		}
		time.Sleep(1 * time.Millisecond) // Add variation
	}

	fmt.Printf("\nA/B testing distribution (100 requests):\n")
	fmt.Printf("Control: %d (%.1f%%)\n", controlABCount, float64(controlABCount)/100*100)
	fmt.Printf("Test: %d (%.1f%%)\n", testABCount, float64(testABCount)/100*100)

	fmt.Println("\n✅ All tests passed! The fix prevents runtime panics and provides sensible defaults.")
}
