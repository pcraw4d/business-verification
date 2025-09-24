package risk

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestSuiteConfig contains configuration for the test suite
type TestSuiteConfig struct {
	// Test types to run
	RunUnitTests        bool `json:"run_unit_tests"`
	RunIntegrationTests bool `json:"run_integration_tests"`
	RunPerformanceTests bool `json:"run_performance_tests"`
	RunConcurrencyTests bool `json:"run_concurrency_tests"`
	RunMemoryTests      bool `json:"run_memory_tests"`

	// Performance test configuration
	BenchmarkIterations int           `json:"benchmark_iterations"`
	TestTimeout         time.Duration `json:"test_timeout"`

	// Concurrency test configuration
	ConcurrentGoroutines int `json:"concurrent_goroutines"`

	// Memory test configuration
	MemoryTestAssessments int `json:"memory_test_assessments"`

	// Logging configuration
	LogLevel  string `json:"log_level"`
	LogFormat string `json:"log_format"`
}

// DefaultTestSuiteConfig returns a default test suite configuration
func DefaultTestSuiteConfig() *TestSuiteConfig {
	return &TestSuiteConfig{
		RunUnitTests:          true,
		RunIntegrationTests:   true,
		RunPerformanceTests:   true,
		RunConcurrencyTests:   true,
		RunMemoryTests:        true,
		BenchmarkIterations:   1000,
		TestTimeout:           30 * time.Second,
		ConcurrentGoroutines:  10,
		MemoryTestAssessments: 100,
		LogLevel:              "info",
		LogFormat:             "json",
	}
}

// ParseTestSuiteConfigFromFlags parses test suite configuration from command line flags
func ParseTestSuiteConfigFromFlags() *TestSuiteConfig {
	config := DefaultTestSuiteConfig()

	flag.BoolVar(&config.RunUnitTests, "unit", config.RunUnitTests, "Run unit tests")
	flag.BoolVar(&config.RunIntegrationTests, "integration", config.RunIntegrationTests, "Run integration tests")
	flag.BoolVar(&config.RunPerformanceTests, "performance", config.RunPerformanceTests, "Run performance tests")
	flag.BoolVar(&config.RunConcurrencyTests, "concurrency", config.RunConcurrencyTests, "Run concurrency tests")
	flag.BoolVar(&config.RunMemoryTests, "memory", config.RunMemoryTests, "Run memory tests")

	flag.IntVar(&config.BenchmarkIterations, "benchmark-iterations", config.BenchmarkIterations, "Number of benchmark iterations")
	flag.DurationVar(&config.TestTimeout, "test-timeout", config.TestTimeout, "Test timeout duration")
	flag.IntVar(&config.ConcurrentGoroutines, "concurrent-goroutines", config.ConcurrentGoroutines, "Number of concurrent goroutines")
	flag.IntVar(&config.MemoryTestAssessments, "memory-assessments", config.MemoryTestAssessments, "Number of memory test assessments")

	flag.StringVar(&config.LogLevel, "log-level", config.LogLevel, "Log level (debug, info, warn, error)")
	flag.StringVar(&config.LogFormat, "log-format", config.LogFormat, "Log format (json, console)")

	flag.Parse()

	return config
}

// TestSuite runs the comprehensive test suite for enhanced risk services
type TestSuite struct {
	config *TestSuiteConfig
	logger *zap.Logger
}

// NewTestSuite creates a new test suite
func NewTestSuite(config *TestSuiteConfig) (*TestSuite, error) {
	// Configure logger
	var logger *zap.Logger
	var err error

	if config.LogFormat == "json" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &TestSuite{
		config: config,
		logger: logger,
	}, nil
}

// Run runs the test suite
func (ts *TestSuite) Run() error {
	ts.logger.Info("Starting enhanced risk services test suite",
		zap.Bool("unit_tests", ts.config.RunUnitTests),
		zap.Bool("integration_tests", ts.config.RunIntegrationTests),
		zap.Bool("performance_tests", ts.config.RunPerformanceTests),
		zap.Bool("concurrency_tests", ts.config.RunConcurrencyTests),
		zap.Bool("memory_tests", ts.config.RunMemoryTests))

	// Create test configuration
	testConfig := &TestConfig{
		PerformanceTestConfig: PerformanceTestConfig{
			BenchmarkIterations:   ts.config.BenchmarkIterations,
			TestTimeout:           ts.config.TestTimeout,
			MemoryThresholdMB:     100,
			ConcurrentGoroutines:  ts.config.ConcurrentGoroutines,
			MemoryTestAssessments: ts.config.MemoryTestAssessments,
		},
		IntegrationTestConfig: IntegrationTestConfig{
			TestTimeout:                60 * time.Second,
			TestCasesPerSuite:          10,
			EnableDatabaseTests:        false,
			EnableExternalServiceTests: false,
		},
		MockDataConfig: MockDataConfig{
			MockRiskFactorsCount:       5,
			MockHistoricalEntriesCount: 12,
			MockRecommendationsCount:   8,
			MockAlertsCount:            3,
			MockHistoricalTimeRange:    12 * 30 * 24 * time.Hour,
		},
	}

	// Create test runner
	testRunner := NewTestRunner(ts.logger, testConfig)

	// Run tests
	testing.Main(func(pat, str string) (bool, error) { return true, nil },
		[]testing.InternalTest{
			{
				Name: "EnhancedRiskServicesTestSuite",
				F: func(t *testing.T) {
					ts.runTestSuite(t, testRunner)
				},
			},
		},
		[]testing.InternalBenchmark{},
		[]testing.InternalExample{})

	ts.logger.Info("Enhanced risk services test suite completed")
	return nil
}

// runTestSuite runs the test suite with the given test runner
func (ts *TestSuite) runTestSuite(t *testing.T, testRunner *TestRunner) {
	if ts.config.RunUnitTests {
		t.Run("UnitTests", func(t *testing.T) {
			testRunner.RunUnitTests(t)
		})
	}

	if ts.config.RunIntegrationTests {
		t.Run("IntegrationTests", func(t *testing.T) {
			testRunner.RunIntegrationTests(t)
		})
	}

	if ts.config.RunPerformanceTests {
		t.Run("PerformanceTests", func(t *testing.T) {
			testRunner.RunPerformanceTests(t)
		})
	}

	if ts.config.RunConcurrencyTests {
		t.Run("ConcurrencyTests", func(t *testing.T) {
			testRunner.RunConcurrencyTests(t)
		})
	}

	if ts.config.RunMemoryTests {
		t.Run("MemoryTests", func(t *testing.T) {
			testRunner.RunMemoryTests(t)
		})
	}
}

// TestEnhancedRiskServicesTestSuite is the main test function that can be run with `go test`
func TestEnhancedRiskServicesTestSuite(t *testing.T) {
	// Parse configuration from flags
	config := ParseTestSuiteConfigFromFlags()

	// Create test suite
	testSuite, err := NewTestSuite(config)
	if err != nil {
		t.Fatalf("Failed to create test suite: %v", err)
	}

	// Run test suite
	err = testSuite.Run()
	if err != nil {
		t.Fatalf("Test suite failed: %v", err)
	}
}

// TestEnhancedRiskServicesQuick is a quick test that runs only essential tests
func TestEnhancedRiskServicesQuick(t *testing.T) {
	config := &TestSuiteConfig{
		RunUnitTests:          true,
		RunIntegrationTests:   true,
		RunPerformanceTests:   false,
		RunConcurrencyTests:   false,
		RunMemoryTests:        false,
		BenchmarkIterations:   100,
		TestTimeout:           10 * time.Second,
		ConcurrentGoroutines:  5,
		MemoryTestAssessments: 10,
		LogLevel:              "info",
		LogFormat:             "console",
	}

	testSuite, err := NewTestSuite(config)
	if err != nil {
		t.Fatalf("Failed to create test suite: %v", err)
	}

	err = testSuite.Run()
	if err != nil {
		t.Fatalf("Quick test suite failed: %v", err)
	}
}

// TestEnhancedRiskServicesPerformance is a performance-focused test
func TestEnhancedRiskServicesPerformance(t *testing.T) {
	config := &TestSuiteConfig{
		RunUnitTests:          false,
		RunIntegrationTests:   false,
		RunPerformanceTests:   true,
		RunConcurrencyTests:   true,
		RunMemoryTests:        true,
		BenchmarkIterations:   5000,
		TestTimeout:           60 * time.Second,
		ConcurrentGoroutines:  20,
		MemoryTestAssessments: 500,
		LogLevel:              "warn",
		LogFormat:             "json",
	}

	testSuite, err := NewTestSuite(config)
	if err != nil {
		t.Fatalf("Failed to create test suite: %v", err)
	}

	err = testSuite.Run()
	if err != nil {
		t.Fatalf("Performance test suite failed: %v", err)
	}
}

// TestEnhancedRiskServicesStress is a stress test
func TestEnhancedRiskServicesStress(t *testing.T) {
	config := &TestSuiteConfig{
		RunUnitTests:          false,
		RunIntegrationTests:   true,
		RunPerformanceTests:   true,
		RunConcurrencyTests:   true,
		RunMemoryTests:        true,
		BenchmarkIterations:   10000,
		TestTimeout:           120 * time.Second,
		ConcurrentGoroutines:  50,
		MemoryTestAssessments: 1000,
		LogLevel:              "error",
		LogFormat:             "json",
	}

	testSuite, err := NewTestSuite(config)
	if err != nil {
		t.Fatalf("Failed to create test suite: %v", err)
	}

	err = testSuite.Run()
	if err != nil {
		t.Fatalf("Stress test suite failed: %v", err)
	}
}

// TestEnhancedRiskServicesSmoke is a smoke test for basic functionality
func TestEnhancedRiskServicesSmoke(t *testing.T) {
	logger := zap.NewNop()
	factory := NewEnhancedRiskServiceFactory(logger)
	service := factory.CreateEnhancedRiskService()

	// Test basic functionality
	request := &EnhancedRiskAssessmentRequest{
		AssessmentID: "smoke-test",
		BusinessID:   "smoke-business",
		RiskFactorInputs: []RiskFactorInput{
			{
				FactorType: "financial",
				Data: map[string]interface{}{
					"revenue": 1000000.0,
					"debt":    500000.0,
					"assets":  2000000.0,
				},
				Weight: 0.3,
			},
		},
		IncludeTrendAnalysis:       false,
		IncludeCorrelationAnalysis: false,
	}

	response, err := service.PerformEnhancedRiskAssessment(context.Background(), request)
	if err != nil {
		t.Fatalf("Smoke test failed: %v", err)
	}

	if response == nil {
		t.Fatal("Smoke test returned nil response")
	}

	if response.AssessmentID != request.AssessmentID {
		t.Errorf("Expected assessment ID %s, got %s", request.AssessmentID, response.AssessmentID)
	}

	if response.BusinessID != request.BusinessID {
		t.Errorf("Expected business ID %s, got %s", request.BusinessID, response.BusinessID)
	}

	if response.OverallRiskScore < 0 || response.OverallRiskScore > 1 {
		t.Errorf("Invalid overall risk score %f", response.OverallRiskScore)
	}

	if len(response.RiskFactors) == 0 {
		t.Error("Expected risk factors but got none")
	}
}

// TestEnhancedRiskServicesErrorHandling tests error handling scenarios
func TestEnhancedRiskServicesErrorHandling(t *testing.T) {
	logger := zap.NewNop()
	factory := NewEnhancedRiskServiceFactory(logger)
	service := factory.CreateEnhancedRiskService()

	t.Run("InvalidRequest", func(t *testing.T) {
		// Test with invalid request
		request := &EnhancedRiskAssessmentRequest{
			AssessmentID:     "",
			BusinessID:       "",
			RiskFactorInputs: []RiskFactorInput{},
		}

		_, err := service.PerformEnhancedRiskAssessment(context.Background(), request)
		// Should handle gracefully or return appropriate error
		if err != nil {
			t.Logf("Expected error for invalid request: %v", err)
		}
	})

	t.Run("InvalidFactorData", func(t *testing.T) {
		// Test with invalid factor data
		request := &EnhancedRiskAssessmentRequest{
			AssessmentID: "error-test",
			BusinessID:   "error-business",
			RiskFactorInputs: []RiskFactorInput{
				{
					FactorType: "invalid_factor",
					Data: map[string]interface{}{
						"invalid_field": "invalid_value",
					},
					Weight: -1.0, // Invalid weight
				},
			},
		}

		_, err := service.PerformEnhancedRiskAssessment(context.Background(), request)
		// Should handle gracefully or return appropriate error
		if err != nil {
			t.Logf("Expected error for invalid factor data: %v", err)
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		// Test with cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		request := &EnhancedRiskAssessmentRequest{
			AssessmentID: "cancelled-test",
			BusinessID:   "cancelled-business",
			RiskFactorInputs: []RiskFactorInput{
				{
					FactorType: "financial",
					Data: map[string]interface{}{
						"revenue": 1000000.0,
					},
					Weight: 0.3,
				},
			},
		}

		_, err := service.PerformEnhancedRiskAssessment(ctx, request)
		// Should handle context cancellation gracefully
		if err != nil {
			t.Logf("Expected error for cancelled context: %v", err)
		}
	})
}

// TestEnhancedRiskServicesDataValidation tests data validation
func TestEnhancedRiskServicesDataValidation(t *testing.T) {
	logger := zap.NewNop()
	factory := NewEnhancedRiskServiceFactory(logger)
	service := factory.CreateEnhancedRiskService()

	t.Run("ValidData", func(t *testing.T) {
		request := &EnhancedRiskAssessmentRequest{
			AssessmentID: "validation-test",
			BusinessID:   "validation-business",
			RiskFactorInputs: []RiskFactorInput{
				{
					FactorType: "financial",
					Data: map[string]interface{}{
						"revenue": 1000000.0,
						"debt":    500000.0,
						"assets":  2000000.0,
					},
					Weight: 0.3,
				},
			},
		}

		response, err := service.PerformEnhancedRiskAssessment(context.Background(), request)
		if err != nil {
			t.Errorf("Valid data test failed: %v", err)
			return
		}

		// Validate response structure
		if response.AssessmentID != request.AssessmentID {
			t.Errorf("Assessment ID mismatch: expected %s, got %s", request.AssessmentID, response.AssessmentID)
		}

		if response.BusinessID != request.BusinessID {
			t.Errorf("Business ID mismatch: expected %s, got %s", request.BusinessID, response.BusinessID)
		}

		if response.OverallRiskScore < 0 || response.OverallRiskScore > 1 {
			t.Errorf("Invalid overall risk score: %f", response.OverallRiskScore)
		}

		if response.ConfidenceScore < 0 || response.ConfidenceScore > 1 {
			t.Errorf("Invalid confidence score: %f", response.ConfidenceScore)
		}

		if len(response.RiskFactors) != len(request.RiskFactorInputs) {
			t.Errorf("Risk factors count mismatch: expected %d, got %d", len(request.RiskFactorInputs), len(response.RiskFactors))
		}
	})
}

// TestEnhancedRiskServicesEdgeCases tests edge cases
func TestEnhancedRiskServicesEdgeCases(t *testing.T) {
	logger := zap.NewNop()
	factory := NewEnhancedRiskServiceFactory(logger)
	service := factory.CreateEnhancedRiskService()

	t.Run("EmptyRiskFactors", func(t *testing.T) {
		request := &EnhancedRiskAssessmentRequest{
			AssessmentID:     "empty-factors-test",
			BusinessID:       "empty-factors-business",
			RiskFactorInputs: []RiskFactorInput{},
		}

		response, err := service.PerformEnhancedRiskAssessment(context.Background(), request)
		if err != nil {
			t.Errorf("Empty risk factors test failed: %v", err)
			return
		}

		if response.OverallRiskScore != 0 {
			t.Errorf("Expected overall risk score 0 for empty factors, got %f", response.OverallRiskScore)
		}
	})

	t.Run("SingleRiskFactor", func(t *testing.T) {
		request := &EnhancedRiskAssessmentRequest{
			AssessmentID: "single-factor-test",
			BusinessID:   "single-factor-business",
			RiskFactorInputs: []RiskFactorInput{
				{
					FactorType: "financial",
					Data: map[string]interface{}{
						"revenue": 1000000.0,
					},
					Weight: 1.0,
				},
			},
		}

		response, err := service.PerformEnhancedRiskAssessment(context.Background(), request)
		if err != nil {
			t.Errorf("Single risk factor test failed: %v", err)
			return
		}

		if len(response.RiskFactors) != 1 {
			t.Errorf("Expected 1 risk factor, got %d", len(response.RiskFactors))
		}
	})

	t.Run("HighWeightRiskFactor", func(t *testing.T) {
		request := &EnhancedRiskAssessmentRequest{
			AssessmentID: "high-weight-test",
			BusinessID:   "high-weight-business",
			RiskFactorInputs: []RiskFactorInput{
				{
					FactorType: "financial",
					Data: map[string]interface{}{
						"revenue": 1000000.0,
					},
					Weight: 10.0, // High weight
				},
			},
		}

		response, err := service.PerformEnhancedRiskAssessment(context.Background(), request)
		if err != nil {
			t.Errorf("High weight risk factor test failed: %v", err)
			return
		}

		// Should handle high weights gracefully
		if response.OverallRiskScore < 0 || response.OverallRiskScore > 1 {
			t.Errorf("Invalid overall risk score with high weight: %f", response.OverallRiskScore)
		}
	})
}

// TestEnhancedRiskServicesMain is the main test function that can be run standalone
func TestEnhancedRiskServicesMain(t *testing.T) {
	// This test can be run with: go test -run TestEnhancedRiskServicesMain
	// It will run the complete test suite with default configuration

	config := DefaultTestSuiteConfig()
	testSuite, err := NewTestSuite(config)
	if err != nil {
		t.Fatalf("Failed to create test suite: %v", err)
	}

	err = testSuite.Run()
	if err != nil {
		t.Fatalf("Main test suite failed: %v", err)
	}
}

// TestEnhancedRiskServicesMainWithFlags is the main test function that can be run with flags
func TestEnhancedRiskServicesMainWithFlags(t *testing.T) {
	// This test can be run with: go test -run TestEnhancedRiskServicesMainWithFlags -unit=true -integration=true
	// It will parse flags and run the test suite accordingly

	config := ParseTestSuiteConfigFromFlags()
	testSuite, err := NewTestSuite(config)
	if err != nil {
		t.Fatalf("Failed to create test suite: %v", err)
	}

	err = testSuite.Run()
	if err != nil {
		t.Fatalf("Main test suite with flags failed: %v", err)
	}
}

// TestEnhancedRiskServicesMainWithArgs is the main test function that can be run with command line arguments
func TestEnhancedRiskServicesMainWithArgs(t *testing.T) {
	// This test can be run with: go test -run TestEnhancedRiskServicesMainWithArgs -args -unit=true -integration=true
	// It will parse command line arguments and run the test suite accordingly

	config := ParseTestSuiteConfigFromFlags()
	testSuite, err := NewTestSuite(config)
	if err != nil {
		t.Fatalf("Failed to create test suite: %v", err)
	}

	err = testSuite.Run()
	if err != nil {
		t.Fatalf("Main test suite with args failed: %v", err)
	}
}

// TestEnhancedRiskServicesMainWithEnv is the main test function that can be run with environment variables
func TestEnhancedRiskServicesMainWithEnv(t *testing.T) {
	// This test can be run with: ENHANCED_RISK_UNIT_TESTS=true ENHANCED_RISK_INTEGRATION_TESTS=true go test -run TestEnhancedRiskServicesMainWithEnv
	// It will parse environment variables and run the test suite accordingly

	config := DefaultTestSuiteConfig()

	// Override with environment variables if set
	if os.Getenv("ENHANCED_RISK_UNIT_TESTS") == "true" {
		config.RunUnitTests = true
	}
	if os.Getenv("ENHANCED_RISK_INTEGRATION_TESTS") == "true" {
		config.RunIntegrationTests = true
	}
	if os.Getenv("ENHANCED_RISK_PERFORMANCE_TESTS") == "true" {
		config.RunPerformanceTests = true
	}
	if os.Getenv("ENHANCED_RISK_CONCURRENCY_TESTS") == "true" {
		config.RunConcurrencyTests = true
	}
	if os.Getenv("ENHANCED_RISK_MEMORY_TESTS") == "true" {
		config.RunMemoryTests = true
	}

	testSuite, err := NewTestSuite(config)
	if err != nil {
		t.Fatalf("Failed to create test suite: %v", err)
	}

	err = testSuite.Run()
	if err != nil {
		t.Fatalf("Main test suite with env failed: %v", err)
	}
}
