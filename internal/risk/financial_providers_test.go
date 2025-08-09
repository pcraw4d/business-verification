package risk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// createTestLogger creates a logger for testing
func createTestLogger() *observability.Logger {
	return observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "json",
	})
}

func TestRealFinancialProvider_GetCompanyFinancials(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Return mock financial data
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"business_id": "test-business-123",
			"provider": "test_provider",
			"last_updated": "2024-01-01T00:00:00Z",
			"revenue": {
				"total_revenue": 1000000.0,
				"revenue_growth": 5.2,
				"currency": "USD"
			},
			"risk_indicators": [
				{
					"indicator": "Test Risk Indicator",
					"value": 0.5,
					"threshold": 0.7,
					"risk_level": "medium",
					"description": "Test risk description",
					"recommendation": "Test recommendation"
				}
			]
		}`))
	}))
	defer server.Close()

	// Create logger
	logger := createTestLogger()

	// Create provider
	provider := NewRealFinancialProvider("test_provider", "test-api-key", server.URL, logger)

	// Test successful request
	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	data, err := provider.GetCompanyFinancials(ctx, "test-business-123")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.BusinessID != "test-business-123" {
		t.Errorf("Expected business ID 'test-business-123', got %s", data.BusinessID)
	}

	if data.Provider != "test_provider" {
		t.Errorf("Expected provider 'test_provider', got %s", data.Provider)
	}

	if data.Revenue == nil {
		t.Error("Expected revenue data to be present")
	} else if data.Revenue.TotalRevenue != 1000000.0 {
		t.Errorf("Expected total revenue 1000000.0, got %f", data.Revenue.TotalRevenue)
	}

	if len(data.RiskIndicators) != 1 {
		t.Errorf("Expected 1 risk indicator, got %d", len(data.RiskIndicators))
	}
}

func TestRealFinancialProvider_GetCreditScore(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"business_id": "test-business-123",
			"provider": "test_provider",
			"score": 750,
			"score_range": "good",
			"last_updated": "2024-01-01T00:00:00Z",
			"trend": "stable",
			"risk_level": "low"
		}`))
	}))
	defer server.Close()

	logger := createTestLogger()
	provider := NewRealFinancialProvider("test_provider", "test-api-key", server.URL, logger)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	score, err := provider.GetCreditScore(ctx, "test-business-123")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if score.BusinessID != "test-business-123" {
		t.Errorf("Expected business ID 'test-business-123', got %s", score.BusinessID)
	}

	if score.Score != 750 {
		t.Errorf("Expected score 750, got %d", score.Score)
	}

	if score.ScoreRange != "good" {
		t.Errorf("Expected score range 'good', got %s", score.ScoreRange)
	}

	if score.RiskLevel != RiskLevelLow {
		t.Errorf("Expected risk level 'low', got %s", score.RiskLevel)
	}
}

func TestRealFinancialProvider_GetPaymentHistory(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"business_id": "test-business-123",
			"provider": "test_provider",
			"total_payments": 120,
			"on_time_payments": 115,
			"late_payments": 3,
			"defaulted_payments": 2,
			"payment_rate": 95.8,
			"average_days_late": 5.2,
			"last_payment_date": "2024-01-01T00:00:00Z",
			"payment_trend": "improving",
			"risk_level": "low"
		}`))
	}))
	defer server.Close()

	logger := observability.NewLogger("test", "debug")
	provider := NewRealFinancialProvider("test_provider", "test-api-key", server.URL, logger)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	history, err := provider.GetPaymentHistory(ctx, "test-business-123")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if history.BusinessID != "test-business-123" {
		t.Errorf("Expected business ID 'test-business-123', got %s", history.BusinessID)
	}

	if history.TotalPayments != 120 {
		t.Errorf("Expected total payments 120, got %d", history.TotalPayments)
	}

	if history.PaymentRate != 95.8 {
		t.Errorf("Expected payment rate 95.8, got %f", history.PaymentRate)
	}

	if history.RiskLevel != RiskLevelLow {
		t.Errorf("Expected risk level 'low', got %s", history.RiskLevel)
	}
}

func TestRealFinancialProvider_GetBankruptcyInfo(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"business_id": "test-business-123",
			"provider": "test_provider",
			"has_bankruptcy": false,
			"risk_level": "low"
		}`))
	}))
	defer server.Close()

	logger := observability.NewLogger("test", "debug")
	provider := NewRealFinancialProvider("test_provider", "test-api-key", server.URL, logger)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	bankruptcy, err := provider.GetBankruptcyInfo(ctx, "test-business-123")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if bankruptcy.BusinessID != "test-business-123" {
		t.Errorf("Expected business ID 'test-business-123', got %s", bankruptcy.BusinessID)
	}

	if bankruptcy.HasBankruptcy != false {
		t.Errorf("Expected has_bankruptcy false, got %t", bankruptcy.HasBankruptcy)
	}

	if bankruptcy.RiskLevel != RiskLevelLow {
		t.Errorf("Expected risk level 'low', got %s", bankruptcy.RiskLevel)
	}
}

func TestRealFinancialProvider_GetLegalActions(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"business_id": "test-business-123",
			"provider": "test_provider",
			"total_actions": 2,
			"active_actions": 1,
			"resolved_actions": 1,
			"actions": [
				{
					"action_id": "action-123",
					"action_type": "lawsuit",
					"filing_date": "2024-01-01T00:00:00Z",
					"status": "active",
					"amount": 50000.0,
					"description": "Test lawsuit",
					"risk_level": "medium"
				}
			],
			"risk_level": "medium"
		}`))
	}))
	defer server.Close()

	logger := observability.NewLogger("test", "debug")
	provider := NewRealFinancialProvider("test_provider", "test-api-key", server.URL, logger)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	actions, err := provider.GetLegalActions(ctx, "test-business-123")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if actions.BusinessID != "test-business-123" {
		t.Errorf("Expected business ID 'test-business-123', got %s", actions.BusinessID)
	}

	if actions.TotalActions != 2 {
		t.Errorf("Expected total actions 2, got %d", actions.TotalActions)
	}

	if actions.ActiveActions != 1 {
		t.Errorf("Expected active actions 1, got %d", actions.ActiveActions)
	}

	if len(actions.Actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(actions.Actions))
	}

	if actions.RiskLevel != RiskLevelMedium {
		t.Errorf("Expected risk level 'medium', got %s", actions.RiskLevel)
	}
}

func TestRealFinancialProvider_GetIndustryBenchmarks(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"industry": "technology",
			"provider": "test_provider",
			"last_updated": "2024-01-01T00:00:00Z",
			"revenue_benchmarks": {
				"median_revenue": 5000000.0,
				"average_revenue": 7500000.0,
				"revenue_growth": 8.5,
				"revenue_stability": 75.0
			},
			"profitability_benchmarks": {
				"median_gross_margin": 0.25,
				"median_net_margin": 0.15,
				"median_roa": 0.12,
				"median_roe": 0.18
			}
		}`))
	}))
	defer server.Close()

	logger := observability.NewLogger("test", "debug")
	provider := NewRealFinancialProvider("test_provider", "test-api-key", server.URL, logger)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	benchmarks, err := provider.GetIndustryBenchmarks(ctx, "technology")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if benchmarks.Industry != "technology" {
		t.Errorf("Expected industry 'technology', got %s", benchmarks.Industry)
	}

	if benchmarks.Provider != "test_provider" {
		t.Errorf("Expected provider 'test_provider', got %s", benchmarks.Provider)
	}

	if benchmarks.RevenueBenchmarks == nil {
		t.Error("Expected revenue benchmarks to be present")
	} else if benchmarks.RevenueBenchmarks.MedianRevenue != 5000000.0 {
		t.Errorf("Expected median revenue 5000000.0, got %f", benchmarks.RevenueBenchmarks.MedianRevenue)
	}

	if benchmarks.ProfitabilityBenchmarks == nil {
		t.Error("Expected profitability benchmarks to be present")
	} else if benchmarks.ProfitabilityBenchmarks.MedianGrossMargin != 0.25 {
		t.Errorf("Expected median gross margin 0.25, got %f", benchmarks.ProfitabilityBenchmarks.MedianGrossMargin)
	}
}

func TestRealFinancialProvider_ErrorHandling(t *testing.T) {
	// Create test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	logger := observability.NewLogger("test", "debug")
	provider := NewRealFinancialProvider("test_provider", "test-api-key", server.URL, logger)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	_, err := provider.GetCompanyFinancials(ctx, "test-business-123")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != "provider returned status 500" {
		t.Errorf("Expected error 'provider returned status 500', got %s", err.Error())
	}
}

func TestRealFinancialProvider_Unauthorized(t *testing.T) {
	// Create test server that returns unauthorized
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	logger := observability.NewLogger("test", "debug")
	provider := NewRealFinancialProvider("test_provider", "invalid-key", server.URL, logger)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	_, err := provider.GetCompanyFinancials(ctx, "test-business-123")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != "provider returned status 401" {
		t.Errorf("Expected error 'provider returned status 401', got %s", err.Error())
	}
}

func TestRealFinancialProvider_NetworkError(t *testing.T) {
	logger := observability.NewLogger("test", "debug")
	provider := NewRealFinancialProvider("test_provider", "test-api-key", "http://invalid-url", logger)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	_, err := provider.GetCompanyFinancials(ctx, "test-business-123")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Should contain network error information
	if err.Error() == "" {
		t.Error("Expected error message, got empty string")
	}
}

func TestCreditBureauProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"business_id": "test-business-123",
			"provider": "credit_bureau",
			"score": 800,
			"score_range": "excellent",
			"last_updated": "2024-01-01T00:00:00Z",
			"trend": "improving",
			"risk_level": "low"
		}`))
	}))
	defer server.Close()

	logger := observability.NewLogger("test", "debug")
	provider := NewCreditBureauProvider("test-api-key", server.URL, logger)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	score, err := provider.GetCreditScore(ctx, "test-business-123")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if score.Provider != "credit_bureau" {
		t.Errorf("Expected provider 'credit_bureau', got %s", score.Provider)
	}

	if score.Score != 800 {
		t.Errorf("Expected score 800, got %d", score.Score)
	}
}

func TestFinancialDataProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"business_id": "test-business-123",
			"provider": "financial_data",
			"last_updated": "2024-01-01T00:00:00Z",
			"revenue": {
				"total_revenue": 2000000.0,
				"revenue_growth": 10.5,
				"currency": "USD"
			}
		}`))
	}))
	defer server.Close()

	logger := observability.NewLogger("test", "debug")
	provider := NewFinancialDataProvider("test-api-key", server.URL, logger)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	data, err := provider.GetCompanyFinancials(ctx, "test-business-123")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.Provider != "financial_data" {
		t.Errorf("Expected provider 'financial_data', got %s", data.Provider)
	}

	if data.Revenue.TotalRevenue != 2000000.0 {
		t.Errorf("Expected total revenue 2000000.0, got %f", data.Revenue.TotalRevenue)
	}
}

func TestRegulatoryDataProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"business_id": "test-business-123",
			"provider": "regulatory_data",
			"has_bankruptcy": false,
			"risk_level": "low"
		}`))
	}))
	defer server.Close()

	logger := observability.NewLogger("test", "debug")
	provider := NewRegulatoryDataProvider("test-api-key", server.URL, logger)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	bankruptcy, err := provider.GetBankruptcyInfo(ctx, "test-business-123")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if bankruptcy.Provider != "regulatory_data" {
		t.Errorf("Expected provider 'regulatory_data', got %s", bankruptcy.Provider)
	}

	if bankruptcy.HasBankruptcy != false {
		t.Errorf("Expected has_bankruptcy false, got %t", bankruptcy.HasBankruptcy)
	}
}

func TestRealFinancialProvider_Availability(t *testing.T) {
	logger := observability.NewLogger("test", "debug")
	provider := NewRealFinancialProvider("test_provider", "test-api-key", "http://example.com", logger)

	// Test default availability
	if !provider.IsAvailable() {
		t.Error("Expected provider to be available by default")
	}

	// Test setting availability
	provider.SetAvailable(false)
	if provider.IsAvailable() {
		t.Error("Expected provider to be unavailable after setting to false")
	}

	provider.SetAvailable(true)
	if !provider.IsAvailable() {
		t.Error("Expected provider to be available after setting to true")
	}
}

func TestRealFinancialProvider_ProviderName(t *testing.T) {
	logger := observability.NewLogger("test", "debug")
	provider := NewRealFinancialProvider("test_provider", "test-api-key", "http://example.com", logger)

	if provider.GetProviderName() != "test_provider" {
		t.Errorf("Expected provider name 'test_provider', got %s", provider.GetProviderName())
	}
}
