package testing

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewIntegrationTest(t *testing.T) {
	test := NewIntegrationTest("Test 1", "First integration test", func(ctx *IntegrationContext) error {
		return nil
	})

	if test.Name != "Test 1" {
		t.Errorf("Expected test name 'Test 1', got '%s'", test.Name)
	}

	if test.Description != "First integration test" {
		t.Errorf("Expected test description 'First integration test', got '%s'", test.Description)
	}

	if test.Status != TestStatusPending {
		t.Errorf("Expected test status pending, got %v", test.Status)
	}

	if test.Parallel {
		t.Error("Expected test to be non-parallel by default")
	}

	if test.Timeout != 30*time.Second {
		t.Errorf("Expected test timeout 30s, got %v", test.Timeout)
	}

	if test.Function == nil {
		t.Error("Expected test function to be set")
	}

	if len(test.Components) != 0 {
		t.Errorf("Expected empty components list, got %d", len(test.Components))
	}

	if test.Database {
		t.Error("Expected database to be false by default")
	}

	if test.External {
		t.Error("Expected external to be false by default")
	}

	if test.Network {
		t.Error("Expected network to be false by default")
	}
}

func TestIntegrationTest_AddTag(t *testing.T) {
	test := NewIntegrationTest("Test 1", "First test", func(ctx *IntegrationContext) error {
		return nil
	})

	test.AddTag("integration")
	test.AddTag("database")

	if len(test.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(test.Tags))
	}

	if test.Tags[0] != "integration" {
		t.Errorf("Expected first tag 'integration', got '%s'", test.Tags[0])
	}

	if test.Tags[1] != "database" {
		t.Errorf("Expected second tag 'database', got '%s'", test.Tags[1])
	}
}

func TestIntegrationTest_SetTimeout(t *testing.T) {
	test := NewIntegrationTest("Test 1", "First test", func(ctx *IntegrationContext) error {
		return nil
	})

	test.SetTimeout(10 * time.Second)

	if test.Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", test.Timeout)
	}
}

func TestIntegrationTest_SetParallel(t *testing.T) {
	test := NewIntegrationTest("Test 1", "First test", func(ctx *IntegrationContext) error {
		return nil
	})

	test.SetParallel(false)

	if test.Parallel {
		t.Error("Expected test to be non-parallel")
	}
}

func TestIntegrationTest_AddComponent(t *testing.T) {
	test := NewIntegrationTest("Test 1", "First test", func(ctx *IntegrationContext) error {
		return nil
	})

	test.AddComponent("database")
	test.AddComponent("api")

	if len(test.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(test.Components))
	}

	if test.Components[0] != "database" {
		t.Errorf("Expected first component 'database', got '%s'", test.Components[0])
	}

	if test.Components[1] != "api" {
		t.Errorf("Expected second component 'api', got '%s'", test.Components[1])
	}
}

func TestIntegrationTest_SetDatabase(t *testing.T) {
	test := NewIntegrationTest("Test 1", "First test", func(ctx *IntegrationContext) error {
		return nil
	})

	test.SetDatabase(true)

	if !test.Database {
		t.Error("Expected database to be true")
	}
}

func TestIntegrationTest_SetExternal(t *testing.T) {
	test := NewIntegrationTest("Test 1", "First test", func(ctx *IntegrationContext) error {
		return nil
	})

	test.SetExternal(true)

	if !test.External {
		t.Error("Expected external to be true")
	}
}

func TestIntegrationTest_SetNetwork(t *testing.T) {
	test := NewIntegrationTest("Test 1", "First test", func(ctx *IntegrationContext) error {
		return nil
	})

	test.SetNetwork(true)

	if !test.Network {
		t.Error("Expected network to be true")
	}
}

func TestNewIntegrationTestSuite(t *testing.T) {
	suite := NewIntegrationTestSuite("Test Suite", "A test suite for integration testing")

	if suite.Name != "Test Suite" {
		t.Errorf("Expected suite name 'Test Suite', got '%s'", suite.Name)
	}

	if suite.Description != "A test suite for integration testing" {
		t.Errorf("Expected suite description 'A test suite for integration testing', got '%s'", suite.Description)
	}

	if suite.Parallel {
		t.Error("Expected suite to be non-parallel by default")
	}

	if suite.Timeout != 60*time.Second {
		t.Errorf("Expected suite timeout 60s, got %v", suite.Timeout)
	}

	if len(suite.Tests) != 0 {
		t.Errorf("Expected empty test list, got %d tests", len(suite.Tests))
	}

	if suite.Database {
		t.Error("Expected database to be false by default")
	}

	if suite.External {
		t.Error("Expected external to be false by default")
	}

	if suite.Network {
		t.Error("Expected network to be false by default")
	}
}

func TestIntegrationTestSuite_AddTest(t *testing.T) {
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")
	test := NewIntegrationTest("Test 1", "First test", func(ctx *IntegrationContext) error {
		return nil
	})

	suite.AddTest(test)

	if len(suite.Tests) != 1 {
		t.Errorf("Expected 1 test, got %d", len(suite.Tests))
	}

	if suite.Tests[0] != test {
		t.Error("Expected test to be added to suite")
	}
}

func TestIntegrationTestSuite_CreateTest(t *testing.T) {
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")

	test := suite.CreateTest("Test 1", "First test", func(ctx *IntegrationContext) error {
		return nil
	})

	if test == nil {
		t.Fatal("Expected test to be created")
	}

	if test.Name != "Test 1" {
		t.Errorf("Expected test name 'Test 1', got '%s'", test.Name)
	}

	if test.Description != "First test" {
		t.Errorf("Expected test description 'First test', got '%s'", test.Description)
	}

	if len(suite.Tests) != 1 {
		t.Errorf("Expected 1 test in suite, got %d", len(suite.Tests))
	}
}

func TestIntegrationTestSuite_SetupTeardown(t *testing.T) {
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")

	setupCalled := false
	teardownCalled := false

	suite.SetSetup(func(ctx *IntegrationContext) error {
		setupCalled = true
		return nil
	})

	suite.SetTeardown(func(ctx *IntegrationContext) error {
		teardownCalled = true
		return nil
	})

	if suite.SetupFunc == nil {
		t.Error("Expected setup function to be set")
	}

	if suite.TeardownFunc == nil {
		t.Error("Expected teardown function to be set")
	}

	// Test setup function
	ctx := &IntegrationContext{}
	if err := suite.SetupFunc(ctx); err != nil {
		t.Errorf("Expected no error from setup, got %v", err)
	}

	if !setupCalled {
		t.Error("Expected setup function to be called")
	}

	// Test teardown function
	if err := suite.TeardownFunc(ctx); err != nil {
		t.Errorf("Expected no error from teardown, got %v", err)
	}

	if !teardownCalled {
		t.Error("Expected teardown function to be called")
	}
}

func TestIntegrationTestSuite_BeforeAfterEach(t *testing.T) {
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")

	beforeCalled := false
	afterCalled := false

	suite.SetBeforeEach(func(ctx *IntegrationContext) error {
		beforeCalled = true
		return nil
	})

	suite.SetAfterEach(func(ctx *IntegrationContext) error {
		afterCalled = true
		return nil
	})

	if suite.BeforeEach == nil {
		t.Error("Expected before each function to be set")
	}

	if suite.AfterEach == nil {
		t.Error("Expected after each function to be set")
	}

	// Test before each function
	ctx := &IntegrationContext{}
	if err := suite.BeforeEach(ctx); err != nil {
		t.Errorf("Expected no error from before each, got %v", err)
	}

	if !beforeCalled {
		t.Error("Expected before each function to be called")
	}

	// Test after each function
	if err := suite.AfterEach(ctx); err != nil {
		t.Errorf("Expected no error from after each, got %v", err)
	}

	if !afterCalled {
		t.Error("Expected after each function to be called")
	}
}

func TestIntegrationTestSuite_SetParallel(t *testing.T) {
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")

	suite.SetParallel(false)

	if suite.Parallel {
		t.Error("Expected suite to be non-parallel")
	}
}

func TestIntegrationTestSuite_SetTimeout(t *testing.T) {
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")

	suite.SetTimeout(10 * time.Second)

	if suite.Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", suite.Timeout)
	}
}

func TestIntegrationTestSuite_AddTag(t *testing.T) {
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")

	suite.AddTag("integration")
	suite.AddTag("database")

	if len(suite.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(suite.Tags))
	}

	if suite.Tags[0] != "integration" {
		t.Errorf("Expected first tag 'integration', got '%s'", suite.Tags[0])
	}

	if suite.Tags[1] != "database" {
		t.Errorf("Expected second tag 'database', got '%s'", suite.Tags[1])
	}
}

func TestIntegrationTestSuite_SetDatabase(t *testing.T) {
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")

	suite.SetDatabase(true)

	if !suite.Database {
		t.Error("Expected database to be true")
	}
}

func TestIntegrationTestSuite_SetExternal(t *testing.T) {
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")

	suite.SetExternal(true)

	if !suite.External {
		t.Error("Expected external to be true")
	}
}

func TestIntegrationTestSuite_SetNetwork(t *testing.T) {
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")

	suite.SetNetwork(true)

	if !suite.Network {
		t.Error("Expected network to be true")
	}
}

func TestNewIntegrationContext(t *testing.T) {
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")
	test := NewIntegrationTest("Test 1", "First test", func(ctx *IntegrationContext) error {
		return nil
	})

	testCtx := NewIntegrationContext(test, suite)
	defer testCtx.Cleanup()

	if testCtx.Test != test {
		t.Error("Expected test context to reference test")
	}

	if testCtx.Suite != suite {
		t.Error("Expected test context to reference suite")
	}

	if testCtx.Context == nil {
		t.Error("Expected test context to have context")
	}

	if testCtx.Fixtures == nil {
		t.Error("Expected test context to have fixtures")
	}

	if testCtx.Mocks == nil {
		t.Error("Expected test context to have mocks")
	}

	if len(testCtx.CleanupFuncs) != 0 {
		t.Errorf("Expected empty cleanup functions, got %d", len(testCtx.CleanupFuncs))
	}
}

func TestIntegrationContext_Log(t *testing.T) {
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")
	test := NewIntegrationTest("Test 1", "First test", func(ctx *IntegrationContext) error {
		return nil
	})

	testCtx := NewIntegrationContext(test, suite)
	defer testCtx.Cleanup()

	testCtx.Log("Test message")

	if len(test.Output) != 1 {
		t.Errorf("Expected 1 log message, got %d", len(test.Output))
	}

	if test.Output[0] != "Test message" {
		t.Errorf("Expected log message 'Test message', got '%s'", test.Output[0])
	}
}

func TestIntegrationContext_Logf(t *testing.T) {
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")
	test := NewIntegrationTest("Test 1", "First test", func(ctx *IntegrationContext) error {
		return nil
	})

	testCtx := NewIntegrationContext(test, suite)
	defer testCtx.Cleanup()

	testCtx.Logf("Test message %d", 42)

	if len(test.Output) != 1 {
		t.Errorf("Expected 1 log message, got %d", len(test.Output))
	}

	if test.Output[0] != "Test message 42" {
		t.Errorf("Expected log message 'Test message 42', got '%s'", test.Output[0])
	}
}

func TestIntegrationContext_AddCleanup(t *testing.T) {
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")
	test := NewIntegrationTest("Test 1", "First test", func(ctx *IntegrationContext) error {
		return nil
	})

	testCtx := NewIntegrationContext(test, suite)
	defer testCtx.Cleanup()

	cleanupCalled := false

	testCtx.AddCleanup(func() error {
		cleanupCalled = true
		return nil
	})

	if len(testCtx.CleanupFuncs) != 1 {
		t.Errorf("Expected 1 cleanup function, got %d", len(testCtx.CleanupFuncs))
	}

	// Cleanup should be called when defer executes
	if cleanupCalled {
		t.Error("Expected cleanup to not be called yet")
	}
}

func TestIntegrationFixtures(t *testing.T) {
	fixtures := NewIntegrationFixtures()

	// Test database fixtures
	fixtures.SetDatabaseFixture("user", map[string]interface{}{"id": 1, "name": "test"})
	value, exists := fixtures.GetDatabaseFixture("user")
	if !exists {
		t.Error("Expected database fixture to exist")
	}
	if value == nil {
		t.Error("Expected database fixture to be non-nil")
	}

	// Test file fixtures
	fixtures.SetFileFixture("config.json", `{"setting": "value"}`)
	content, exists := fixtures.GetFileFixture("config.json")
	if !exists {
		t.Error("Expected file fixture to exist")
	}
	if content != `{"setting": "value"}` {
		t.Errorf("Expected content '{\"setting\": \"value\"}', got '%s'", content)
	}

	// Test config fixtures
	config := map[string]interface{}{"setting": "value"}
	fixtures.SetConfigFixture("app", config)
	retrievedConfig, exists := fixtures.GetConfigFixture("app")
	if !exists {
		t.Error("Expected config fixture to exist")
	}
	if retrievedConfig == nil {
		t.Error("Expected config fixture to be non-nil")
	}

	// Test HTTP fixtures
	httpFixture := &HTTPFixture{
		Method:     "GET",
		Path:       "/api/test",
		StatusCode: 200,
		Body:       `{"result": "success"}`,
	}
	fixtures.SetHTTPFixture("api_test", httpFixture)
	retrievedHTTPFixture, exists := fixtures.GetHTTPFixture("api_test")
	if !exists {
		t.Error("Expected HTTP fixture to exist")
	}
	if retrievedHTTPFixture.Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", retrievedHTTPFixture.Method)
	}
}

func TestIntegrationMocks(t *testing.T) {
	mocks := NewIntegrationMocks()

	// Test HTTP mocks
	httpMock := mocks.AddHTTPMock("api", "http://localhost:8080")
	if httpMock == nil {
		t.Error("Expected HTTP mock to be created")
	}
	if httpMock.ID != "api" {
		t.Errorf("Expected ID 'api', got '%s'", httpMock.ID)
	}
	if httpMock.BaseURL != "http://localhost:8080" {
		t.Errorf("Expected base URL 'http://localhost:8080', got '%s'", httpMock.BaseURL)
	}

	retrievedHTTPMock, exists := mocks.GetHTTPMock("api")
	if !exists {
		t.Error("Expected HTTP mock to exist")
	}
	if retrievedHTTPMock != httpMock {
		t.Error("Expected retrieved HTTP mock to match created mock")
	}

	// Test database mocks
	dbMock := mocks.AddDatabaseMock("main_db")
	if dbMock == nil {
		t.Error("Expected database mock to be created")
	}
	if dbMock.ID != "main_db" {
		t.Errorf("Expected ID 'main_db', got '%s'", dbMock.ID)
	}

	retrievedDBMock, exists := mocks.GetDatabaseMock("main_db")
	if !exists {
		t.Error("Expected database mock to exist")
	}
	if retrievedDBMock != dbMock {
		t.Error("Expected retrieved database mock to match created mock")
	}

	// Test external mocks
	extMock := mocks.AddExternalMock("payment_service", "payment")
	if extMock == nil {
		t.Error("Expected external mock to be created")
	}
	if extMock.ID != "payment_service" {
		t.Errorf("Expected ID 'payment_service', got '%s'", extMock.ID)
	}
	if extMock.Service != "payment" {
		t.Errorf("Expected service 'payment', got '%s'", extMock.Service)
	}

	retrievedExtMock, exists := mocks.GetExternalMock("payment_service")
	if !exists {
		t.Error("Expected external mock to exist")
	}
	if retrievedExtMock != extMock {
		t.Error("Expected retrieved external mock to match created mock")
	}
}

func TestHTTPMock(t *testing.T) {
	mock := &HTTPMock{
		ID:       "test",
		BaseURL:  "http://localhost:8080",
		Handlers: make(map[string]http.HandlerFunc),
	}

	// Add handler
	mock.AddHandler("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"status": "ok"}`))
	})

	if len(mock.Handlers) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(mock.Handlers))
	}

	// Test start/stop
	mock.Start()
	if mock.Server == nil {
		t.Error("Expected server to be started")
	}

	mock.Stop()
	// Server should be closed
}

func TestDatabaseMock(t *testing.T) {
	mock := &DatabaseMock{
		ID:      "test",
		Queries: make(map[string]*QueryMock),
		Results: make(map[string][]interface{}),
		Errors:  make(map[string]error),
	}

	// Add query mock
	mock.AddQuery("SELECT * FROM users", []interface{}{1, "test"}, nil)

	if len(mock.Queries) != 1 {
		t.Errorf("Expected 1 query mock, got %d", len(mock.Queries))
	}

	queryMock, exists := mock.GetQuery("SELECT * FROM users")
	if !exists {
		t.Error("Expected query mock to exist")
	}
	if queryMock.Query != "SELECT * FROM users" {
		t.Errorf("Expected query 'SELECT * FROM users', got '%s'", queryMock.Query)
	}
}

func TestExternalMock(t *testing.T) {
	mock := &ExternalMock{
		ID:      "test",
		Service: "payment",
		Methods: make(map[string]*MethodMock),
		CallLog: make([]*ExternalCall, 0),
	}

	// Add method mock
	mock.AddMethod("process_payment", []interface{}{"order_id", 100.0}, []interface{}{"success", "txn_123"}, nil)

	if len(mock.Methods) != 1 {
		t.Errorf("Expected 1 method mock, got %d", len(mock.Methods))
	}

	methodMock, exists := mock.GetMethod("process_payment")
	if !exists {
		t.Error("Expected method mock to exist")
	}
	if methodMock.Method != "process_payment" {
		t.Errorf("Expected method 'process_payment', got '%s'", methodMock.Method)
	}

	// Test call recording
	mock.RecordCall("process_payment", []interface{}{"order_id", 100.0}, []interface{}{"success", "txn_123"}, nil, 100*time.Millisecond)

	if len(mock.CallLog) != 1 {
		t.Errorf("Expected 1 call log entry, got %d", len(mock.CallLog))
	}

	call := mock.CallLog[0]
	if call.Method != "process_payment" {
		t.Errorf("Expected method 'process_payment', got '%s'", call.Method)
	}
	if call.Service != "payment" {
		t.Errorf("Expected service 'payment', got '%s'", call.Service)
	}
}

func TestDefaultIntegrationTestConfig(t *testing.T) {
	config := DefaultIntegrationTestConfig()

	if config.DatabaseURL != "" {
		t.Errorf("Expected empty database URL, got '%s'", config.DatabaseURL)
	}

	if config.DatabaseDriver != "postgres" {
		t.Errorf("Expected database driver 'postgres', got '%s'", config.DatabaseDriver)
	}

	if config.HTTPTimeout != 30*time.Second {
		t.Errorf("Expected HTTP timeout 30s, got %v", config.HTTPTimeout)
	}

	if config.MaxRetries != 3 {
		t.Errorf("Expected max retries 3, got %d", config.MaxRetries)
	}

	if config.RetryDelay != 1*time.Second {
		t.Errorf("Expected retry delay 1s, got %v", config.RetryDelay)
	}

	if config.Parallel {
		t.Error("Expected parallel to be false by default")
	}

	if config.MaxGoroutines != 10 {
		t.Errorf("Expected max goroutines 10, got %d", config.MaxGoroutines)
	}

	if config.FailFast {
		t.Error("Expected fail fast to be false by default")
	}

	if !config.CoverageEnabled {
		t.Error("Expected coverage to be enabled by default")
	}

	if config.CoverageOutput != "integration_coverage.out" {
		t.Errorf("Expected coverage output 'integration_coverage.out', got '%s'", config.CoverageOutput)
	}

	if config.LogLevel != "info" {
		t.Errorf("Expected log level 'info', got '%s'", config.LogLevel)
	}

	if len(config.Tags) != 0 {
		t.Errorf("Expected empty tags, got %d", len(config.Tags))
	}

	if len(config.SkipTags) != 0 {
		t.Errorf("Expected empty skip tags, got %d", len(config.SkipTags))
	}

	if len(config.Components) != 0 {
		t.Errorf("Expected empty components, got %d", len(config.Components))
	}

	if len(config.ExternalServices) != 0 {
		t.Errorf("Expected empty external services, got %d", len(config.ExternalServices))
	}
}

func TestNewIntegrationTestRunner(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultIntegrationTestConfig()
	runner := NewIntegrationTestRunner(config, logger)

	if runner.logger != logger {
		t.Error("Expected runner to use provided logger")
	}

	if runner.config != config {
		t.Error("Expected runner to use provided config")
	}

	if len(runner.suites) != 0 {
		t.Errorf("Expected empty suites list, got %d", len(runner.suites))
	}

	if len(runner.results) != 0 {
		t.Errorf("Expected empty results list, got %d", len(runner.results))
	}
}

func TestIntegrationTestRunner_AddSuite(t *testing.T) {
	logger := zap.NewNop()
	runner := NewIntegrationTestRunner(nil, logger)
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")

	runner.AddSuite(suite)

	if len(runner.suites) != 1 {
		t.Errorf("Expected 1 suite, got %d", len(runner.suites))
	}

	if runner.suites[0] != suite {
		t.Error("Expected suite to be added to runner")
	}
}

func TestIntegrationTestRunner_RunAllSuites(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultIntegrationTestConfig()
	config.Parallel = false // For predictable testing
	runner := NewIntegrationTestRunner(config, logger)

	// Create a test suite with a simple test
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")
	testCount := 0

	suite.CreateTest("Test 1", "First test", func(ctx *IntegrationContext) error {
		testCount++
		ctx.Log("Test 1 executed")
		return nil
	})

	suite.CreateTest("Test 2", "Second test", func(ctx *IntegrationContext) error {
		testCount++
		ctx.Log("Test 2 executed")
		return nil
	})

	runner.AddSuite(suite)

	// Run all suites
	results := runner.RunAllSuites()

	if len(results) != 2 {
		t.Errorf("Expected 2 test results, got %d", len(results))
	}

	if testCount != 2 {
		t.Errorf("Expected 2 tests to be executed, got %d", testCount)
	}

	// Check test results
	for _, result := range results {
		if result.Status != TestStatusPassed {
			t.Errorf("Expected test to pass, got status %v", result.Status)
		}
		if result.Error != nil {
			t.Errorf("Expected no error, got %v", result.Error)
		}
	}
}

func TestIntegrationTestRunner_GenerateSummary(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultIntegrationTestConfig()
	config.Parallel = false
	runner := NewIntegrationTestRunner(config, logger)

	// Create a test suite with passing and failing tests
	suite := NewIntegrationTestSuite("Test Suite", "A test suite")

	suite.CreateTest("Passing Test", "This test passes", func(ctx *IntegrationContext) error {
		ctx.Log("Passing test executed")
		return nil
	})

	suite.CreateTest("Failing Test", "This test fails", func(ctx *IntegrationContext) error {
		ctx.Log("Failing test executed")
		return errors.New("test failed")
	})

	runner.AddSuite(suite)

	// Run tests
	runner.RunAllSuites()

	// Generate summary
	summary := runner.GenerateSummary()

	if summary.TotalTests != 2 {
		t.Errorf("Expected 2 total tests, got %d", summary.TotalTests)
	}

	if summary.PassedTests != 1 {
		t.Errorf("Expected 1 passed test, got %d", summary.PassedTests)
	}

	if summary.FailedTests != 1 {
		t.Errorf("Expected 1 failed test, got %d", summary.FailedTests)
	}

	if summary.TotalDuration <= 0 {
		t.Error("Expected positive total duration")
	}

	// Check suite summary
	suiteSummary, exists := summary.Suites["Test Suite"]
	if !exists {
		t.Error("Expected suite summary to exist")
	}

	if suiteSummary.TotalTests != 2 {
		t.Errorf("Expected 2 tests in suite summary, got %d", suiteSummary.TotalTests)
	}

	if suiteSummary.PassedTests != 1 {
		t.Errorf("Expected 1 passed test in suite summary, got %d", suiteSummary.PassedTests)
	}
}

func TestIntegrationTestRunner_shouldSkipTest(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultIntegrationTestConfig()
	runner := NewIntegrationTestRunner(config, logger)

	// Test with tag filters
	config.Tags = []string{"integration"}
	test := NewIntegrationTest("Test 1", "First test", func(ctx *IntegrationContext) error {
		return nil
	})
	test.AddTag("unit")

	// Test should be skipped because it doesn't have the required tag
	if !runner.shouldSkipTest(test) {
		t.Error("Expected test to be skipped due to tag filter")
	}

	// Test with matching tag
	test.AddTag("integration")
	if runner.shouldSkipTest(test) {
		t.Error("Expected test to not be skipped with matching tag")
	}

	// Test with skip tags
	config.Tags = []string{}
	config.SkipTags = []string{"slow"}
	test = NewIntegrationTest("Test 2", "Second test", func(ctx *IntegrationContext) error {
		return nil
	})
	test.AddTag("slow")

	// Test should be skipped because it has a skip tag
	if !runner.shouldSkipTest(test) {
		t.Error("Expected test to be skipped due to skip tag")
	}

	// Test with component filters
	config.SkipTags = []string{}
	config.Components = []string{"database"}
	test = NewIntegrationTest("Test 3", "Third test", func(ctx *IntegrationContext) error {
		return nil
	})
	test.AddComponent("api")

	// Test should be skipped because it doesn't have the required component
	if !runner.shouldSkipTest(test) {
		t.Error("Expected test to be skipped due to component filter")
	}

	// Test with matching component
	test.AddComponent("database")
	if runner.shouldSkipTest(test) {
		t.Error("Expected test to not be skipped with matching component")
	}
}
