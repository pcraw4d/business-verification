package performance

import (
	"context"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/internal/risk"
)

// BenchmarkClassificationService tests classification service performance
func BenchmarkClassificationService(b *testing.B) {
	// Setup test environment
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})

	// Create classification service
	service := classification.NewClassificationService(nil, nil, logger, nil)

	// Test data
	testBusinesses := []struct {
		name  string
		input *classification.ClassificationRequest
	}{
		{
			name: "Technology Corporation",
			input: &classification.ClassificationRequest{
				BusinessName: "Acme Technology Corporation",
				BusinessType: "Corporation",
				Industry:     "Technology",
			},
		},
		{
			name: "Financial Services LLC",
			input: &classification.ClassificationRequest{
				BusinessName: "Global Financial Services LLC",
				BusinessType: "LLC",
				Industry:     "Financial Services",
			},
		},
		{
			name: "Manufacturing Company",
			input: &classification.ClassificationRequest{
				BusinessName: "Industrial Manufacturing Co",
				BusinessType: "Corporation",
				Industry:     "Manufacturing",
			},
		},
	}

	b.ResetTimer()

	for _, testCase := range testBusinesses {
		b.Run(testCase.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ctx := context.Background()
				_, err := service.ClassifyBusiness(ctx, testCase.input)
				if err != nil {
					b.Fatalf("Classification failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkBatchClassification tests batch classification performance
func BenchmarkBatchClassification(b *testing.B) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})

	service := classification.NewClassificationService(nil, nil, logger, nil)

	// Create batch of test businesses
	batchSize := 100
	businesses := make([]classification.ClassificationRequest, batchSize)

	for i := 0; i < batchSize; i++ {
		businesses[i] = classification.ClassificationRequest{
			BusinessName: "Test Business " + string(rune(i+'A')),
			BusinessType: "Corporation",
			Industry:     "Technology",
		}
	}

	b.ResetTimer()

	b.Run("Batch Classification", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			batchReq := &classification.BatchClassificationRequest{Businesses: businesses}
			_, err := service.ClassifyBusinessesBatch(ctx, batchReq)
			if err != nil {
				b.Fatalf("Batch classification failed: %v", err)
			}
		}
	})
}

// BenchmarkRiskAssessment tests risk assessment performance
func BenchmarkRiskAssessment(b *testing.B) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})

	service := risk.NewService(nil, logger)

	testCases := []struct {
		name  string
		input risk.RiskAssessmentInput
	}{
		{
			name: "Low Risk Business",
			input: risk.RiskAssessmentInput{
				BusinessID:   "low-risk-123",
				BusinessName: "Stable Corporation",
				BusinessType: "Corporation",
				Industry:     "Technology",
				FinancialData: risk.FinancialData{
					AnnualRevenue:     10000000,
					EmployeeCount:     500,
					YearsInBusiness:   10,
					ProfitMargin:      0.15,
					DebtToEquityRatio: 0.3,
				},
				Location: risk.Location{
					Country: "US",
					State:   "CA",
					City:    "San Francisco",
				},
			},
		},
		{
			name: "High Risk Business",
			input: risk.RiskAssessmentInput{
				BusinessID:   "high-risk-456",
				BusinessName: "Startup Ventures",
				BusinessType: "LLC",
				Industry:     "Financial Services",
				FinancialData: risk.FinancialData{
					AnnualRevenue:     100000,
					EmployeeCount:     5,
					YearsInBusiness:   1,
					ProfitMargin:      -0.2,
					DebtToEquityRatio: 2.5,
				},
				Location: risk.Location{
					Country: "US",
					State:   "NY",
					City:    "New York",
				},
			},
		},
		{
			name: "Medium Risk Business",
			input: risk.RiskAssessmentInput{
				BusinessID:   "medium-risk-789",
				BusinessName: "Growing Manufacturing",
				BusinessType: "Corporation",
				Industry:     "Manufacturing",
				FinancialData: risk.FinancialData{
					AnnualRevenue:     5000000,
					EmployeeCount:     100,
					YearsInBusiness:   5,
					ProfitMargin:      0.08,
					DebtToEquityRatio: 0.8,
				},
				Location: risk.Location{
					Country: "US",
					State:   "TX",
					City:    "Houston",
				},
			},
		},
	}

	b.ResetTimer()

	for _, testCase := range testCases {
		b.Run(testCase.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ctx := context.Background()
				_, err := service.AssessRisk(ctx, testCase.input)
				if err != nil {
					b.Fatalf("Risk assessment failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkDatabaseOperations tests database operation performance
func BenchmarkDatabaseOperations(b *testing.B) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})

	// This would connect to a test database
	// For now, we'll simulate database operations
	db := &MockDatabase{}

	b.Run("Database Read Operations", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			_, err := db.GetBusiness(ctx, "test-business-id")
			if err != nil {
				b.Fatalf("Database read failed: %v", err)
			}
		}
	})

	b.Run("Database Write Operations", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			business := &database.Business{
				ID:           "test-business-" + string(rune(i)),
				Name:         "Test Business",
				BusinessType: "Corporation",
				Industry:     "Technology",
				CreatedAt:    time.Now(),
			}
			err := db.CreateBusiness(ctx, business)
			if err != nil {
				b.Fatalf("Database write failed: %v", err)
			}
		}
	})
}

// BenchmarkConcurrentOperations tests performance under concurrent load
func BenchmarkConcurrentOperations(b *testing.B) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})

	classificationService := classification.NewService(nil, logger)
	riskService := risk.NewService(nil, logger)

	b.Run("Concurrent Classification", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			input := classification.BusinessInput{
				BusinessName: "Concurrent Test Business",
				BusinessType: "Corporation",
				Industry:     "Technology",
				Location: classification.Location{
					Country: "US",
					State:   "CA",
					City:    "San Francisco",
				},
			}

			for pb.Next() {
				ctx := context.Background()
				_, err := classificationService.ClassifyBusiness(ctx, input)
				if err != nil {
					b.Fatalf("Concurrent classification failed: %v", err)
				}
			}
		})
	})

	b.Run("Concurrent Risk Assessment", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			input := risk.RiskAssessmentInput{
				BusinessID:   "concurrent-test",
				BusinessName: "Concurrent Test Business",
				BusinessType: "Corporation",
				Industry:     "Technology",
				FinancialData: risk.FinancialData{
					AnnualRevenue:     1000000,
					EmployeeCount:     50,
					YearsInBusiness:   3,
					ProfitMargin:      0.1,
					DebtToEquityRatio: 0.5,
				},
				Location: risk.Location{
					Country: "US",
					State:   "CA",
					City:    "San Francisco",
				},
			}

			for pb.Next() {
				ctx := context.Background()
				_, err := riskService.AssessRisk(ctx, input)
				if err != nil {
					b.Fatalf("Concurrent risk assessment failed: %v", err)
				}
			}
		})
	})
}

// BenchmarkMemoryUsage tests memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})

	service := classification.NewService(nil, logger)

	// Test with large batch of businesses
	batchSize := 1000
	businesses := make([]classification.BusinessInput, batchSize)

	for i := 0; i < batchSize; i++ {
		businesses[i] = classification.BusinessInput{
			BusinessName: "Large Batch Business " + string(rune(i%26+'A')),
			BusinessType: "Corporation",
			Industry:     "Technology",
			Location: classification.Location{
				Country: "US",
				State:   "CA",
				City:    "San Francisco",
			},
		}
	}

	b.ResetTimer()

	b.Run("Large Batch Processing", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			_, err := service.ClassifyBusinesses(ctx, businesses)
			if err != nil {
				b.Fatalf("Large batch processing failed: %v", err)
			}
		}
	})
}

// MockDatabase provides a mock database for testing
type MockDatabase struct{}

func (m *MockDatabase) GetBusiness(ctx context.Context, id string) (*database.Business, error) {
	// Simulate database read
	time.Sleep(1 * time.Millisecond)
	return &database.Business{
		ID:           id,
		Name:         "Test Business",
		BusinessType: "Corporation",
		Industry:     "Technology",
		CreatedAt:    time.Now(),
	}, nil
}

func (m *MockDatabase) CreateBusiness(ctx context.Context, business *database.Business) error {
	// Simulate database write
	time.Sleep(2 * time.Millisecond)
	return nil
}

func (m *MockDatabase) Close() error {
	return nil
}
