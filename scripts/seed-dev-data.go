package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/supabase-community/supabase-go"
	"go.uber.org/zap"
)

// SeedDevData seeds development database with sample data
func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	// Get Supabase configuration from environment
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		logger.Fatal("SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY must be set")
	}

	// Create Supabase client
	client, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		logger.Fatal("Failed to create Supabase client", zap.Error(err))
	}

	ctx := context.Background()

	// Seed merchants
	if err := seedMerchants(ctx, client, logger); err != nil {
		logger.Error("Failed to seed merchants", zap.Error(err))
	}

	// Seed risk benchmarks
	if err := seedRiskBenchmarks(ctx, client, logger); err != nil {
		logger.Error("Failed to seed risk benchmarks", zap.Error(err))
	}

	logger.Info("Development data seeding completed")
}

// seedMerchants seeds sample merchant data
func seedMerchants(ctx context.Context, client *supabase.Client, logger *zap.Logger) error {
	merchants := []map[string]interface{}{
		{
			"id":             "merchant_001",
			"name":           "Acme Corporation",
			"legal_name":     "Acme Corporation LLC",
			"portfolio_type": "onboarded",
			"risk_level":     "low",
			"status":         "active",
			"industry":       "Technology",
			"industry_code":  "5734",
			"created_at":     time.Now().Format(time.RFC3339),
			"updated_at":     time.Now().Format(time.RFC3339),
		},
		{
			"id":             "merchant_002",
			"name":           "Global Trading Co",
			"legal_name":     "Global Trading Company Inc",
			"portfolio_type": "prospective",
			"risk_level":     "medium",
			"status":         "active",
			"industry":       "Retail",
			"industry_code":  "5999",
			"created_at":     time.Now().Format(time.RFC3339),
			"updated_at":     time.Now().Format(time.RFC3339),
		},
		{
			"id":             "merchant_003",
			"name":           "Secure Finance Ltd",
			"legal_name":     "Secure Finance Limited",
			"portfolio_type": "onboarded",
			"risk_level":     "high",
			"status":         "active",
			"industry":       "Financial Services",
			"industry_code":  "6012",
			"created_at":     time.Now().Format(time.RFC3339),
			"updated_at":     time.Now().Format(time.RFC3339),
		},
	}

	for _, merchant := range merchants {
		merchantJSON, _ := json.Marshal(merchant)
		var result []map[string]interface{}

		_, err := client.From("merchants").
			Upsert(string(merchantJSON), "", false).
			ExecuteTo(&result)

		if err != nil {
			logger.Warn("Failed to seed merchant",
				zap.String("merchant_id", merchant["id"].(string)),
				zap.Error(err))
		} else {
			logger.Info("Seeded merchant",
				zap.String("merchant_id", merchant["id"].(string)))
		}
	}

	return nil
}

// seedRiskBenchmarks seeds sample risk benchmark data
func seedRiskBenchmarks(ctx context.Context, client *supabase.Client, logger *zap.Logger) error {
	benchmarks := []map[string]interface{}{
		{
			"industry_code": "5734",
			"industry_type": "mcc",
			"average_score": 75.5,
			"median_score":  76.0,
			"percentile_75": 82.0,
			"percentile_90": 88.0,
			"updated_at":    time.Now().Format(time.RFC3339),
		},
		{
			"industry_code": "5999",
			"industry_type": "mcc",
			"average_score": 65.0,
			"median_score":  66.0,
			"percentile_75": 72.0,
			"percentile_90": 78.0,
			"updated_at":    time.Now().Format(time.RFC3339),
		},
		{
			"industry_code": "6012",
			"industry_type": "mcc",
			"average_score": 55.0,
			"median_score":  56.0,
			"percentile_75": 62.0,
			"percentile_90": 68.0,
			"updated_at":    time.Now().Format(time.RFC3339),
		},
	}

	for _, benchmark := range benchmarks {
		benchmarkJSON, _ := json.Marshal(benchmark)
		var result []map[string]interface{}

		_, err := client.From("risk_benchmarks").
			Upsert(string(benchmarkJSON), "", false).
			ExecuteTo(&result)

		if err != nil {
			logger.Warn("Failed to seed benchmark",
				zap.String("industry_code", benchmark["industry_code"].(string)),
				zap.Error(err))
		} else {
			logger.Info("Seeded benchmark",
				zap.String("industry_code", benchmark["industry_code"].(string)))
		}
	}

	return nil
}
