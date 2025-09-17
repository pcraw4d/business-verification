package main

import (
	"context"
	"log"

	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/pcraw4d/business-verification/internal/database"
)

func main() {
	// Create a mock Supabase client (we won't actually connect)
	supabaseClient := &database.SupabaseClient{}

	// Create logger
	logger := log.New(log.Writer(), "TEST ", log.LstdFlags)

	// Create integration service
	integrationService := classification.NewIntegrationService(supabaseClient, logger)

	// Test the multi-method classification
	ctx := context.Background()
	result, err := integrationService.ProcessBusinessClassification(
		ctx,
		"Test Restaurant",
		"A local restaurant serving Italian cuisine",
		"https://testrestaurant.com",
	)

	if err != nil {
		logger.Printf("❌ Classification failed: %v", err)
		return
	}

	logger.Printf("✅ Multi-method classification successful!")
	logger.Printf("Primary Classification: %s", result.PrimaryClassification.IndustryName)
	logger.Printf("Ensemble Confidence: %.2f%%", result.EnsembleConfidence*100)
	logger.Printf("Method Count: %d", len(result.MethodResults))
	logger.Printf("Processing Time: %v", result.ProcessingTime)

	// Print method breakdown
	for i, method := range result.MethodResults {
		logger.Printf("Method %d: %s (%s) - Confidence: %.2f%% - Success: %v",
			i+1, method.MethodName, method.MethodType, method.Confidence*100, method.Success)
	}
}
