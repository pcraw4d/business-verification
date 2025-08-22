package external

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewVerificationAutomatedTester(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil config
	tester := NewVerificationAutomatedTester(nil, logger)
	assert.NotNil(t, tester)
	assert.NotNil(t, tester.config)
	assert.Equal(t, true, tester.config.EnableAutomatedTesting)
	assert.Equal(t, 10, tester.config.MaxConcurrentTests)

	// Test with custom config
	customConfig := &AutomatedTestingConfig{
		EnableAutomatedTesting: false,
		MaxConcurrentTests:     5,
		TestTimeout:            2 * time.Minute,
	}

	tester = NewVerificationAutomatedTester(customConfig, logger)
	assert.NotNil(t, tester)
	assert.Equal(t, false, tester.config.EnableAutomatedTesting)
	assert.Equal(t, 5, tester.config.MaxConcurrentTests)
	assert.Equal(t, 2*time.Minute, tester.config.TestTimeout)
}

func TestDefaultAutomatedTestingConfig(t *testing.T) {
	config := DefaultAutomatedTestingConfig()

	assert.NotNil(t, config)
	assert.Equal(t, true, config.EnableAutomatedTesting)
	assert.Equal(t, true, config.EnableContinuousTesting)
	assert.Equal(t, true, config.EnableRegressionTesting)
	assert.Equal(t, true, config.EnablePerformanceTesting)
	assert.Equal(t, true, config.EnableLoadTesting)
	assert.Equal(t, 1*time.Hour, config.TestInterval)
	assert.Equal(t, 10, config.MaxConcurrentTests)
	assert.Equal(t, 5*time.Minute, config.TestTimeout)
	assert.Equal(t, 1000, config.MaxTestHistory)
	assert.Equal(t, 0.95, config.SuccessThreshold)
	assert.Equal(t, 2*time.Second, config.PerformanceThreshold)
	assert.Equal(t, 5*time.Minute, config.LoadTestDuration)
	assert.Equal(t, 50, config.LoadTestConcurrency)
	assert.Equal(t, 0.05, config.RegressionTestThreshold)
	assert.Equal(t, 30*time.Minute, config.ContinuousTestInterval)
	assert.Equal(t, true, config.AlertOnFailure)
	assert.Equal(t, true, config.AlertOnPerformanceDegrade)
}

func TestCreateTestSuite(t *testing.T) {
	logger := zap.NewNop()
	tester := NewVerificationAutomatedTester(nil, logger)

	// Test valid test suite creation
	suite := &TestSuite{
		Name:        "Test Suite 1",
		Description: "A test suite for verification",
		Category:    "verification",
	}

	err := tester.CreateTestSuite(suite)
	assert.NoError(t, err)
	assert.NotEmpty(t, suite.ID)
	assert.NotZero(t, suite.CreatedAt)
	assert.NotZero(t, suite.UpdatedAt)
	assert.NotNil(t, suite.Tests)
	assert.NotNil(t, suite.Config)

	// Test nil test suite
	err = tester.CreateTestSuite(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test suite cannot be nil")

	// Test missing name
	err = tester.CreateTestSuite(&TestSuite{
		Description: "A test suite",
		Category:    "verification",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test suite name is required")

	// Test missing category
	err = tester.CreateTestSuite(&TestSuite{
		Name:        "Test Suite 2",
		Description: "A test suite",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test suite category is required")
}

func TestAddTest(t *testing.T) {
	logger := zap.NewNop()
	tester := NewVerificationAutomatedTester(nil, logger)

	// Create a test suite first
	suite := &TestSuite{
		Name:        "Test Suite",
		Description: "A test suite",
		Category:    "verification",
	}
	err := tester.CreateTestSuite(suite)
	require.NoError(t, err)

	// Test valid test addition
	test := &AutomatedTest{
		Name:        "Test 1",
		Description: "A unit test",
		Type:        TestTypeUnit,
		Input:       "test input",
		Expected:    "test output",
	}

	err = tester.AddTest(suite.ID, test)
	assert.NoError(t, err)
	assert.NotEmpty(t, test.ID)
	assert.NotZero(t, test.CreatedAt)
	assert.Equal(t, TestPriorityMedium, test.Priority)
	assert.Equal(t, 1.0, test.Weight)
	assert.NotNil(t, test.Metadata)

	// Test nil test
	err = tester.AddTest(suite.ID, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test cannot be nil")

	// Test missing test name
	err = tester.AddTest(suite.ID, &AutomatedTest{
		Description: "A test",
		Type:        TestTypeUnit,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test name is required")

	// Test non-existent suite
	err = tester.AddTest("non-existent", test)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test suite not found")
}

func TestRunTestSuite(t *testing.T) {
	logger := zap.NewNop()
	tester := NewVerificationAutomatedTester(nil, logger)

	// Create a test suite with tests
	suite := &TestSuite{
		Name:        "Test Suite",
		Description: "A test suite",
		Category:    "verification",
	}
	err := tester.CreateTestSuite(suite)
	require.NoError(t, err)

	// Add various types of tests
	tests := []*AutomatedTest{
		{
			Name:        "Unit Test",
			Description: "A unit test",
			Type:        TestTypeUnit,
			Input:       "test input",
			Expected:    "test output",
		},
		{
			Name:        "Integration Test",
			Description: "An integration test",
			Type:        TestTypeIntegration,
			Input:       "integration input",
			Expected:    "integration output",
		},
		{
			Name:        "Performance Test",
			Description: "A performance test",
			Type:        TestTypePerformance,
			Input:       "performance input",
			Expected:    "performance output",
		},
	}

	for _, test := range tests {
		err = tester.AddTest(suite.ID, test)
		require.NoError(t, err)
	}

	// Run the test suite
	ctx := context.Background()
	summary, err := tester.RunTestSuite(ctx, suite.ID)

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 3, summary.TotalTests)
	assert.Equal(t, 3, summary.PassedTests)
	assert.Equal(t, 0, summary.FailedTests)
	assert.Equal(t, 1.0, summary.SuccessRate)
	assert.NotZero(t, summary.AverageTime)
	assert.NotZero(t, summary.TotalDuration)
	assert.NotNil(t, summary.Performance)

	// Test non-existent suite
	_, err = tester.RunTestSuite(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test suite not found")
}

func TestGetTestSuite(t *testing.T) {
	logger := zap.NewNop()
	tester := NewVerificationAutomatedTester(nil, logger)

	// Create a test suite
	suite := &TestSuite{
		Name:        "Test Suite",
		Description: "A test suite",
		Category:    "verification",
	}
	err := tester.CreateTestSuite(suite)
	require.NoError(t, err)

	// Retrieve the test suite
	retrieved, err := tester.GetTestSuite(suite.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, suite.ID, retrieved.ID)
	assert.Equal(t, suite.Name, retrieved.Name)
	assert.Equal(t, suite.Description, retrieved.Description)
	assert.Equal(t, suite.Category, retrieved.Category)

	// Test non-existent suite
	_, err = tester.GetTestSuite("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test suite not found")
}

func TestListTestSuites(t *testing.T) {
	logger := zap.NewNop()
	tester := NewVerificationAutomatedTester(nil, logger)

	// Initially should be empty
	suites := tester.ListTestSuites()
	assert.Empty(t, suites)

	// Create multiple test suites
	testSuites := []*TestSuite{
		{
			Name:        "Suite 1",
			Description: "First suite",
			Category:    "verification",
		},
		{
			Name:        "Suite 2",
			Description: "Second suite",
			Category:    "performance",
		},
		{
			Name:        "Suite 3",
			Description: "Third suite",
			Category:    "integration",
		},
	}

	for _, suite := range testSuites {
		err := tester.CreateTestSuite(suite)
		require.NoError(t, err)
	}

	// List all test suites
	suites = tester.ListTestSuites()
	assert.Len(t, suites, 3)

	// Verify all suites are present
	suiteNames := make(map[string]bool)
	for _, suite := range suites {
		suiteNames[suite.Name] = true
	}

	assert.True(t, suiteNames["Suite 1"])
	assert.True(t, suiteNames["Suite 2"])
	assert.True(t, suiteNames["Suite 3"])
}

func TestGetTestResults(t *testing.T) {
	logger := zap.NewNop()
	tester := NewVerificationAutomatedTester(nil, logger)

	// Create and run a test suite to generate results
	suite := &TestSuite{
		Name:        "Test Suite",
		Description: "A test suite",
		Category:    "verification",
	}
	err := tester.CreateTestSuite(suite)
	require.NoError(t, err)

	test := &AutomatedTest{
		Name:        "Test",
		Description: "A test",
		Type:        TestTypeUnit,
		Input:       "input",
		Expected:    "output",
	}
	err = tester.AddTest(suite.ID, test)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = tester.RunTestSuite(ctx, suite.ID)
	require.NoError(t, err)

	// Get all results
	results := tester.GetTestResults(0, "")
	assert.Len(t, results, 1)
	assert.Equal(t, TestStatusPassed, results[0].Status)

	// Get results with limit
	results = tester.GetTestResults(1, "")
	assert.Len(t, results, 1)

	// Get results by status
	results = tester.GetTestResults(0, TestStatusPassed)
	assert.Len(t, results, 1)

	results = tester.GetTestResults(0, TestStatusFailed)
	assert.Len(t, results, 0)
}

func TestVerificationAutomatedTestingUpdateConfig(t *testing.T) {
	logger := zap.NewNop()
	tester := NewVerificationAutomatedTester(nil, logger)

	// Test valid config update
	newConfig := &AutomatedTestingConfig{
		EnableAutomatedTesting: false,
		MaxConcurrentTests:     5,
		TestTimeout:            2 * time.Minute,
		SuccessThreshold:       0.90,
	}

	err := tester.UpdateConfig(newConfig)
	assert.NoError(t, err)

	// Verify config was updated
	config := tester.GetConfig()
	assert.Equal(t, false, config.EnableAutomatedTesting)
	assert.Equal(t, 5, config.MaxConcurrentTests)
	assert.Equal(t, 2*time.Minute, config.TestTimeout)
	assert.Equal(t, 0.90, config.SuccessThreshold)

	// Test nil config
	err = tester.UpdateConfig(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")

	// Test invalid success threshold
	invalidConfig := &AutomatedTestingConfig{
		SuccessThreshold: 1.5, // Should be between 0 and 1
	}
	err = tester.UpdateConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "success threshold must be between 0 and 1")

	// Test invalid max concurrent tests
	invalidConfig = &AutomatedTestingConfig{
		MaxConcurrentTests: 0, // Should be positive
	}
	err = tester.UpdateConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max concurrent tests must be positive")

	// Test invalid test timeout
	invalidConfig = &AutomatedTestingConfig{
		TestTimeout: 0, // Should be positive
	}
	err = tester.UpdateConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test timeout must be positive")
}

func TestAutomatedTestExecution(t *testing.T) {
	logger := zap.NewNop()
	tester := NewVerificationAutomatedTester(nil, logger)

	// Test different test types
	testCases := []struct {
		name     string
		testType TestType
		input    interface{}
		expected interface{}
	}{
		{
			name:     "Unit Test",
			testType: TestTypeUnit,
			input:    "unit input",
			expected: "unit output",
		},
		{
			name:     "Integration Test",
			testType: TestTypeIntegration,
			input:    "integration input",
			expected: "integration output",
		},
		{
			name:     "Performance Test",
			testType: TestTypePerformance,
			input:    "performance input",
			expected: "performance output",
		},
		{
			name:     "Load Test",
			testType: TestTypeLoad,
			input:    "load input",
			expected: "load output",
		},
		{
			name:     "Smoke Test",
			testType: TestTypeSmoke,
			input:    "smoke input",
			expected: "smoke output",
		},
		{
			name:     "End-to-End Test",
			testType: TestTypeEndToEnd,
			input:    "e2e input",
			expected: "e2e output",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite := &TestSuite{
				Name:        tc.name + " Suite",
				Description: "Test suite for " + tc.name,
				Category:    "testing",
			}
			err := tester.CreateTestSuite(suite)
			require.NoError(t, err)

			test := &AutomatedTest{
				Name:        tc.name,
				Description: "A " + string(tc.testType) + " test",
				Type:        tc.testType,
				Input:       tc.input,
				Expected:    tc.expected,
			}
			err = tester.AddTest(suite.ID, test)
			require.NoError(t, err)

			ctx := context.Background()
			summary, err := tester.RunTestSuite(ctx, suite.ID)

			assert.NoError(t, err)
			assert.NotNil(t, summary)
			assert.Equal(t, 1, summary.TotalTests)
			assert.Equal(t, 1, summary.PassedTests)
			assert.Equal(t, 1.0, summary.SuccessRate)
		})
	}
}

func TestTestValidation(t *testing.T) {
	logger := zap.NewNop()
	tester := NewVerificationAutomatedTester(nil, logger)

	// Create a test suite
	suite := &TestSuite{
		Name:        "Validation Test Suite",
		Description: "A test suite for validation",
		Category:    "validation",
	}
	err := tester.CreateTestSuite(suite)
	require.NoError(t, err)

	// Test with custom validator
	validationError := "validation failed"
	test := &AutomatedTest{
		Name:        "Validation Test",
		Description: "A test with custom validation",
		Type:        TestTypeUnit,
		Input:       "input",
		Expected:    "output",
		Validator: func(result interface{}) error {
			return fmt.Errorf("%s", validationError)
		},
	}
	err = tester.AddTest(suite.ID, test)
	require.NoError(t, err)

	ctx := context.Background()
	summary, err := tester.RunTestSuite(ctx, suite.ID)

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 1, summary.TotalTests)
	assert.Equal(t, 0, summary.PassedTests)
	assert.Equal(t, 1, summary.FailedTests)
	assert.Equal(t, 0.0, summary.SuccessRate)

	// Check the test result
	results := tester.GetTestResults(0, TestStatusFailed)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Error, validationError)
}

func TestTestSetupAndTeardown(t *testing.T) {
	logger := zap.NewNop()
	tester := NewVerificationAutomatedTester(nil, logger)

	// Create a test suite
	suite := &TestSuite{
		Name:        "Setup Teardown Test Suite",
		Description: "A test suite for setup and teardown",
		Category:    "testing",
	}
	err := tester.CreateTestSuite(suite)
	require.NoError(t, err)

	setupCalled := false
	teardownCalled := false

	test := &AutomatedTest{
		Name:        "Setup Teardown Test",
		Description: "A test with setup and teardown",
		Type:        TestTypeUnit,
		Input:       "input",
		Expected:    "output",
		Setup: func() error {
			setupCalled = true
			return nil
		},
		Teardown: func() error {
			teardownCalled = true
			return nil
		},
	}
	err = tester.AddTest(suite.ID, test)
	require.NoError(t, err)

	ctx := context.Background()
	summary, err := tester.RunTestSuite(ctx, suite.ID)

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 1, summary.PassedTests)
	assert.True(t, setupCalled)
	assert.True(t, teardownCalled)
}

func TestTestSetupFailure(t *testing.T) {
	logger := zap.NewNop()
	tester := NewVerificationAutomatedTester(nil, logger)

	// Create a test suite
	suite := &TestSuite{
		Name:        "Setup Failure Test Suite",
		Description: "A test suite for setup failure",
		Category:    "testing",
	}
	err := tester.CreateTestSuite(suite)
	require.NoError(t, err)

	setupError := "setup failed"
	test := &AutomatedTest{
		Name:        "Setup Failure Test",
		Description: "A test with failing setup",
		Type:        TestTypeUnit,
		Input:       "input",
		Expected:    "output",
		Setup: func() error {
			return fmt.Errorf("%s", setupError)
		},
	}
	err = tester.AddTest(suite.ID, test)
	require.NoError(t, err)

	ctx := context.Background()
	summary, err := tester.RunTestSuite(ctx, suite.ID)

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 1, summary.TotalTests)
	assert.Equal(t, 0, summary.PassedTests)
	assert.Equal(t, 1, summary.FailedTests)

	// Check the test result
	results := tester.GetTestResults(0, TestStatusError)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Error, setupError)
}

func TestPerformanceMetrics(t *testing.T) {
	logger := zap.NewNop()
	tester := NewVerificationAutomatedTester(nil, logger)

	// Create a test suite
	suite := &TestSuite{
		Name:        "Performance Test Suite",
		Description: "A test suite for performance testing",
		Category:    "performance",
	}
	err := tester.CreateTestSuite(suite)
	require.NoError(t, err)

	// Add performance test
	test := &AutomatedTest{
		Name:        "Performance Test",
		Description: "A performance test",
		Type:        TestTypePerformance,
		Input:       "performance input",
		Expected:    "performance output",
	}
	err = tester.AddTest(suite.ID, test)
	require.NoError(t, err)

	ctx := context.Background()
	summary, err := tester.RunTestSuite(ctx, suite.ID)

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 1, summary.PassedTests)
	assert.NotNil(t, summary.Performance)
	assert.NotZero(t, summary.Performance.ResponseTime)
	assert.NotZero(t, summary.Performance.Throughput)
}

func TestLoadTestMetrics(t *testing.T) {
	logger := zap.NewNop()
	tester := NewVerificationAutomatedTester(nil, logger)

	// Create a test suite
	suite := &TestSuite{
		Name:        "Load Test Suite",
		Description: "A test suite for load testing",
		Category:    "load",
	}
	err := tester.CreateTestSuite(suite)
	require.NoError(t, err)

	// Add load test
	test := &AutomatedTest{
		Name:        "Load Test",
		Description: "A load test",
		Type:        TestTypeLoad,
		Input:       "load input",
		Expected:    "load output",
	}
	err = tester.AddTest(suite.ID, test)
	require.NoError(t, err)

	ctx := context.Background()
	summary, err := tester.RunTestSuite(ctx, suite.ID)

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 1, summary.PassedTests)
	assert.NotNil(t, summary.Performance)
	assert.NotZero(t, summary.Performance.ResponseTime)
	assert.NotZero(t, summary.Performance.Throughput)
	assert.NotZero(t, summary.Performance.MemoryUsage)
	assert.NotZero(t, summary.Performance.CPUUsage)
	assert.NotZero(t, summary.Performance.NetworkLatency)
	assert.NotZero(t, summary.Performance.DatabaseQueries)
	assert.NotZero(t, summary.Performance.CacheHits)
	assert.NotZero(t, summary.Performance.CacheMisses)
}

func TestConcurrentTestExecution(t *testing.T) {
	logger := zap.NewNop()
	config := &AutomatedTestingConfig{
		MaxConcurrentTests: 3,
		TestTimeout:        10 * time.Second,
	}
	tester := NewVerificationAutomatedTester(config, logger)

	// Create a test suite
	suite := &TestSuite{
		Name:        "Concurrent Test Suite",
		Description: "A test suite for concurrent execution",
		Category:    "concurrent",
	}
	err := tester.CreateTestSuite(suite)
	require.NoError(t, err)

	// Add multiple tests
	for i := 0; i < 5; i++ {
		test := &AutomatedTest{
			Name:        fmt.Sprintf("Test %d", i+1),
			Description: fmt.Sprintf("Test number %d", i+1),
			Type:        TestTypeUnit,
			Input:       fmt.Sprintf("input %d", i+1),
			Expected:    fmt.Sprintf("output %d", i+1),
		}
		err = tester.AddTest(suite.ID, test)
		require.NoError(t, err)
	}

	ctx := context.Background()
	summary, err := tester.RunTestSuite(ctx, suite.ID)

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 5, summary.TotalTests)
	assert.Equal(t, 5, summary.PassedTests)
	assert.Equal(t, 1.0, summary.SuccessRate)
}

func TestTestTimeout(t *testing.T) {
	logger := zap.NewNop()
	config := &AutomatedTestingConfig{
		TestTimeout: 100 * time.Millisecond,
	}
	tester := NewVerificationAutomatedTester(config, logger)

	// Create a test suite
	suite := &TestSuite{
		Name:        "Timeout Test Suite",
		Description: "A test suite for timeout testing",
		Category:    "timeout",
	}
	err := tester.CreateTestSuite(suite)
	require.NoError(t, err)

	// Add a test that takes longer than the timeout
	test := &AutomatedTest{
		Name:        "Timeout Test",
		Description: "A test that should timeout",
		Type:        TestTypeUnit,
		Input:       "input",
		Expected:    "output",
		Setup: func() error {
			time.Sleep(200 * time.Millisecond) // Longer than timeout
			return nil
		},
	}
	err = tester.AddTest(suite.ID, test)
	require.NoError(t, err)

	ctx := context.Background()
	summary, err := tester.RunTestSuite(ctx, suite.ID)

	// The test should fail due to timeout
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 1, summary.TotalTests)
	assert.Equal(t, 0, summary.PassedTests)
	assert.Equal(t, 1, summary.FailedTests)
}
