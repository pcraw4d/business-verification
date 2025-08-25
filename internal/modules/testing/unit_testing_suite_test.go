package testing

import (
	"errors"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewTestSuite(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite for testing")

	if suite.Name != "Test Suite" {
		t.Errorf("Expected suite name 'Test Suite', got '%s'", suite.Name)
	}

	if suite.Description != "A test suite for testing" {
		t.Errorf("Expected suite description 'A test suite for testing', got '%s'", suite.Description)
	}

	if !suite.Parallel {
		t.Error("Expected suite to be parallel by default")
	}

	if suite.Timeout != 30*time.Second {
		t.Errorf("Expected suite timeout 30s, got %v", suite.Timeout)
	}

	if len(suite.Tests) != 0 {
		t.Errorf("Expected empty test list, got %d tests", len(suite.Tests))
	}
}

func TestTestSuite_AddTest(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
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

func TestTestSuite_CreateTest(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")

	test := suite.CreateTest("Test 1", "First test", func(tc *TestContext) error {
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

func TestTestSuite_SetupTeardown(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")

	setupCalled := false
	teardownCalled := false

	suite.SetSetup(func() error {
		setupCalled = true
		return nil
	})

	suite.SetTeardown(func() error {
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
	if err := suite.SetupFunc(); err != nil {
		t.Errorf("Expected no error from setup, got %v", err)
	}

	if !setupCalled {
		t.Error("Expected setup function to be called")
	}

	// Test teardown function
	if err := suite.TeardownFunc(); err != nil {
		t.Errorf("Expected no error from teardown, got %v", err)
	}

	if !teardownCalled {
		t.Error("Expected teardown function to be called")
	}
}

func TestTestSuite_BeforeAfterEach(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")

	beforeCalled := false
	afterCalled := false

	suite.SetBeforeEach(func() error {
		beforeCalled = true
		return nil
	})

	suite.SetAfterEach(func() error {
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
	if err := suite.BeforeEach(); err != nil {
		t.Errorf("Expected no error from before each, got %v", err)
	}

	if !beforeCalled {
		t.Error("Expected before each function to be called")
	}

	// Test after each function
	if err := suite.AfterEach(); err != nil {
		t.Errorf("Expected no error from after each, got %v", err)
	}

	if !afterCalled {
		t.Error("Expected after each function to be called")
	}
}

func TestNewUnitTest(t *testing.T) {
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	if test.Name != "Test 1" {
		t.Errorf("Expected test name 'Test 1', got '%s'", test.Name)
	}

	if test.Description != "First test" {
		t.Errorf("Expected test description 'First test', got '%s'", test.Description)
	}

	if test.Status != TestStatusPending {
		t.Errorf("Expected test status pending, got %v", test.Status)
	}

	if !test.Parallel {
		t.Error("Expected test to be parallel by default")
	}

	if test.Timeout != 30*time.Second {
		t.Errorf("Expected test timeout 30s, got %v", test.Timeout)
	}

	if test.Function == nil {
		t.Error("Expected test function to be set")
	}
}

func TestUnitTest_AddTag(t *testing.T) {
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	test.AddTag("unit")
	test.AddTag("fast")

	if len(test.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(test.Tags))
	}

	if test.Tags[0] != "unit" {
		t.Errorf("Expected first tag 'unit', got '%s'", test.Tags[0])
	}

	if test.Tags[1] != "fast" {
		t.Errorf("Expected second tag 'fast', got '%s'", test.Tags[1])
	}
}

func TestUnitTest_SetTimeout(t *testing.T) {
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	test.SetTimeout(10 * time.Second)

	if test.Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", test.Timeout)
	}
}

func TestUnitTest_SetParallel(t *testing.T) {
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	test.SetParallel(false)

	if test.Parallel {
		t.Error("Expected test to be non-parallel")
	}
}

func TestUnitTest_Skip(t *testing.T) {
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	test.Skip("Not implemented yet")

	if !test.Skipped {
		t.Error("Expected test to be skipped")
	}

	if test.SkipReason != "Not implemented yet" {
		t.Errorf("Expected skip reason 'Not implemented yet', got '%s'", test.SkipReason)
	}
}

func TestNewTestContext(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	testCtx := NewTestContext(test, suite, t)
	defer testCtx.Cleanup()

	if testCtx.Test != test {
		t.Error("Expected test context to reference test")
	}

	if testCtx.Suite != suite {
		t.Error("Expected test context to reference suite")
	}

	if testCtx.T != t {
		t.Error("Expected test context to reference testing.T")
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
}

func TestTestContext_Log(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	testCtx := NewTestContext(test, suite, t)
	defer testCtx.Cleanup()

	testCtx.Log("Test message")

	if len(test.Output) != 1 {
		t.Errorf("Expected 1 log message, got %d", len(test.Output))
	}

	if test.Output[0] != "Test message" {
		t.Errorf("Expected log message 'Test message', got '%s'", test.Output[0])
	}
}

func TestTestContext_Logf(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	testCtx := NewTestContext(test, suite, t)
	defer testCtx.Cleanup()

	testCtx.Logf("Test message %d", 42)

	if len(test.Output) != 1 {
		t.Errorf("Expected 1 log message, got %d", len(test.Output))
	}

	if test.Output[0] != "Test message 42" {
		t.Errorf("Expected log message 'Test message 42', got '%s'", test.Output[0])
	}
}

func TestAssertionBuilder_Equal(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	testCtx := NewTestContext(test, suite, &testing.T{})
	defer testCtx.Cleanup()

	// Test successful assertion
	result := testCtx.Assert().Equal(42, 42, "Values should be equal")
	if !result {
		t.Error("Expected equal assertion to pass")
	}

	if len(test.Assertions) != 1 {
		t.Errorf("Expected 1 assertion, got %d", len(test.Assertions))
	}

	assertion := test.Assertions[0]
	if !assertion.Success {
		t.Error("Expected assertion to be successful")
	}

	if assertion.Operator != "equal" {
		t.Errorf("Expected operator 'equal', got '%s'", assertion.Operator)
	}

	// Test failed assertion
	result = testCtx.Assert().Equal(42, 43, "Values should not be equal")
	if result {
		t.Error("Expected unequal assertion to fail")
	}

	if len(test.Assertions) != 2 {
		t.Errorf("Expected 2 assertions, got %d", len(test.Assertions))
	}

	failedAssertion := test.Assertions[1]
	if failedAssertion.Success {
		t.Error("Expected assertion to fail")
	}
}

func TestAssertionBuilder_NotEqual(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	testCtx := NewTestContext(test, suite, &testing.T{})
	defer testCtx.Cleanup()

	// Test successful assertion
	result := testCtx.Assert().NotEqual(42, 43, "Values should not be equal")
	if !result {
		t.Error("Expected not equal assertion to pass")
	}

	// Test failed assertion
	result = testCtx.Assert().NotEqual(42, 42, "Values should be equal")
	if result {
		t.Error("Expected not equal assertion to fail when values are equal")
	}
}

func TestAssertionBuilder_True(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	testCtx := NewTestContext(test, suite, &testing.T{})
	defer testCtx.Cleanup()

	// Test successful assertion
	result := testCtx.Assert().True(true, "Value should be true")
	if !result {
		t.Error("Expected true assertion to pass")
	}

	// Test failed assertion
	result = testCtx.Assert().True(false, "Value should be true")
	if result {
		t.Error("Expected true assertion to fail when value is false")
	}
}

func TestAssertionBuilder_False(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	testCtx := NewTestContext(test, suite, &testing.T{})
	defer testCtx.Cleanup()

	// Test successful assertion
	result := testCtx.Assert().False(false, "Value should be false")
	if !result {
		t.Error("Expected false assertion to pass")
	}

	// Test failed assertion
	result = testCtx.Assert().False(true, "Value should be false")
	if result {
		t.Error("Expected false assertion to fail when value is true")
	}
}

func TestAssertionBuilder_Nil(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	testCtx := NewTestContext(test, suite, &testing.T{})
	defer testCtx.Cleanup()

	// Test successful assertion
	result := testCtx.Assert().Nil(nil, "Value should be nil")
	if !result {
		t.Error("Expected nil assertion to pass")
	}

	// Test failed assertion
	result = testCtx.Assert().Nil("not nil", "Value should be nil")
	if result {
		t.Error("Expected nil assertion to fail when value is not nil")
	}
}

func TestAssertionBuilder_NotNil(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	testCtx := NewTestContext(test, suite, &testing.T{})
	defer testCtx.Cleanup()

	// Test successful assertion
	result := testCtx.Assert().NotNil("not nil", "Value should not be nil")
	if !result {
		t.Error("Expected not nil assertion to pass")
	}

	// Test failed assertion
	result = testCtx.Assert().NotNil(nil, "Value should not be nil")
	if result {
		t.Error("Expected not nil assertion to fail when value is nil")
	}
}

func TestAssertionBuilder_Error(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	testCtx := NewTestContext(test, suite, &testing.T{})
	defer testCtx.Cleanup()

	// Test successful assertion
	err := errors.New("test error")
	result := testCtx.Assert().Error(err, "Should have an error")
	if !result {
		t.Error("Expected error assertion to pass")
	}

	// Test failed assertion
	result = testCtx.Assert().Error(nil, "Should have an error")
	if result {
		t.Error("Expected error assertion to fail when no error")
	}
}

func TestAssertionBuilder_NoError(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	testCtx := NewTestContext(test, suite, &testing.T{})
	defer testCtx.Cleanup()

	// Test successful assertion
	result := testCtx.Assert().NoError(nil, "Should have no error")
	if !result {
		t.Error("Expected no error assertion to pass")
	}

	// Test failed assertion
	err := errors.New("test error")
	result = testCtx.Assert().NoError(err, "Should have no error")
	if result {
		t.Error("Expected no error assertion to fail when there is an error")
	}
}

func TestAssertionBuilder_Contains(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	testCtx := NewTestContext(test, suite, &testing.T{})
	defer testCtx.Cleanup()

	// Test successful assertion
	result := testCtx.Assert().Contains("hello world", "world", "String should contain substring")
	if !result {
		t.Error("Expected contains assertion to pass")
	}

	// Test failed assertion
	result = testCtx.Assert().Contains("hello world", "foo", "String should contain substring")
	if result {
		t.Error("Expected contains assertion to fail when substring not found")
	}
}

func TestAssertionBuilder_Len(t *testing.T) {
	suite := NewTestSuite("Test Suite", "A test suite")
	test := NewUnitTest("Test 1", "First test", func(tc *TestContext) error {
		return nil
	})

	testCtx := NewTestContext(test, suite, &testing.T{})
	defer testCtx.Cleanup()

	// Test successful assertion with slice
	slice := []int{1, 2, 3}
	result := testCtx.Assert().Len(slice, 3, "Slice should have length 3")
	if !result {
		t.Error("Expected len assertion to pass for slice")
	}

	// Test successful assertion with string
	str := "hello"
	result = testCtx.Assert().Len(str, 5, "String should have length 5")
	if !result {
		t.Error("Expected len assertion to pass for string")
	}

	// Test failed assertion
	result = testCtx.Assert().Len(slice, 2, "Slice should have length 2")
	if result {
		t.Error("Expected len assertion to fail when length is wrong")
	}

	// Test with invalid type
	result = testCtx.Assert().Len(42, 1, "Number should have length")
	if result {
		t.Error("Expected len assertion to fail for invalid type")
	}
}

func TestTestFixtures(t *testing.T) {
	fixtures := NewTestFixtures()

	// Test data operations
	fixtures.SetData("key1", "value1")
	value, exists := fixtures.GetData("key1")
	if !exists {
		t.Error("Expected data to exist")
	}
	if value != "value1" {
		t.Errorf("Expected value 'value1', got '%v'", value)
	}

	// Test non-existent data
	_, exists = fixtures.GetData("nonexistent")
	if exists {
		t.Error("Expected data not to exist")
	}

	// Test file operations
	fixtures.SetFile("test.txt", "file content")
	content, exists := fixtures.GetFile("test.txt")
	if !exists {
		t.Error("Expected file to exist")
	}
	if content != "file content" {
		t.Errorf("Expected content 'file content', got '%s'", content)
	}

	// Test config operations
	config := map[string]interface{}{"setting": "value"}
	fixtures.SetConfig("app", config)
	retrievedConfig, exists := fixtures.GetConfig("app")
	if !exists {
		t.Error("Expected config to exist")
	}
	if retrievedConfig == nil {
		t.Error("Expected config to be non-nil")
	}
}

func TestMockManager(t *testing.T) {
	mockManager := NewMockManager()

	// Test mock operations
	mockService := "test service"
	mockManager.AddMock("service", mockService)

	mock, exists := mockManager.GetMock("service")
	if !exists {
		t.Error("Expected mock to exist")
	}
	if mock != mockService {
		t.Errorf("Expected mock '%v', got '%v'", mockService, mock)
	}

	// Test call recording
	args := []interface{}{"arg1", "arg2"}
	result := []interface{}{"result1"}
	err := errors.New("test error")
	mockManager.RecordCall("TestFunction", args, result, err)

	calls := mockManager.GetCalls("TestFunction")
	if len(calls) != 1 {
		t.Errorf("Expected 1 call, got %d", len(calls))
	}

	call := calls[0]
	if call.Function != "TestFunction" {
		t.Errorf("Expected function 'TestFunction', got '%s'", call.Function)
	}
	if len(call.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(call.Args))
	}
	if call.Error != err {
		t.Errorf("Expected error '%v', got '%v'", err, call.Error)
	}

	// Test behavior setting
	behavior := &MockBehavior{
		ID:          "behavior1",
		Function:    "TestFunction",
		ReturnValue: []interface{}{"mocked result"},
		Error:       nil,
		Delay:       time.Millisecond,
		CallCount:   0,
		MaxCalls:    5,
	}
	mockManager.SetBehavior("TestFunction", behavior)

	retrievedBehavior, exists := mockManager.GetBehavior("TestFunction")
	if !exists {
		t.Error("Expected behavior to exist")
	}
	if retrievedBehavior.ID != behavior.ID {
		t.Errorf("Expected behavior ID '%s', got '%s'", behavior.ID, retrievedBehavior.ID)
	}
}

func TestDefaultTestConfig(t *testing.T) {
	config := DefaultTestConfig()

	if !config.Parallel {
		t.Error("Expected parallel to be true by default")
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.Timeout)
	}

	if config.MaxGoroutines <= 0 {
		t.Errorf("Expected positive max goroutines, got %d", config.MaxGoroutines)
	}

	if !config.CoverageEnabled {
		t.Error("Expected coverage to be enabled by default")
	}

	if config.CoverageOutput != "coverage.out" {
		t.Errorf("Expected coverage output 'coverage.out', got '%s'", config.CoverageOutput)
	}

	if config.FailFast {
		t.Error("Expected fail fast to be false by default")
	}
}

func TestNewTestRunner(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultTestConfig()
	runner := NewTestRunner(config, logger)

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

func TestTestRunner_AddSuite(t *testing.T) {
	logger := zap.NewNop()
	runner := NewTestRunner(nil, logger)
	suite := NewTestSuite("Test Suite", "A test suite")

	runner.AddSuite(suite)

	if len(runner.suites) != 1 {
		t.Errorf("Expected 1 suite, got %d", len(runner.suites))
	}

	if runner.suites[0] != suite {
		t.Error("Expected suite to be added to runner")
	}
}

func TestTestRunner_RunAllSuites(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultTestConfig()
	config.Parallel = false // For predictable testing
	runner := NewTestRunner(config, logger)

	// Create a test suite with a simple test
	suite := NewTestSuite("Test Suite", "A test suite")
	testCount := 0

	suite.CreateTest("Test 1", "First test", func(tc *TestContext) error {
		testCount++
		tc.Assert().Equal(1, 1, "One should equal one")
		return nil
	})

	suite.CreateTest("Test 2", "Second test", func(tc *TestContext) error {
		testCount++
		tc.Assert().True(true, "True should be true")
		return nil
	})

	runner.AddSuite(suite)

	// Run all suites
	results := runner.RunAllSuites(t)

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

func TestTestRunner_GenerateSummary(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultTestConfig()
	config.Parallel = false
	runner := NewTestRunner(config, logger)

	// Create a test suite with passing and failing tests
	suite := NewTestSuite("Test Suite", "A test suite")

	suite.CreateTest("Passing Test", "This test passes", func(tc *TestContext) error {
		tc.Assert().True(true, "Should pass")
		return nil
	})

	suite.CreateTest("Failing Test", "This test fails", func(tc *TestContext) error {
		tc.Assert().True(false, "Should fail")
		return nil
	})

	skippedTest := suite.CreateTest("Skipped Test", "This test is skipped", func(tc *TestContext) error {
		return nil
	})
	skippedTest.Skip("Not implemented")

	runner.AddSuite(suite)

	// Run tests
	runner.RunAllSuites(t)

	// Generate summary
	summary := runner.GenerateSummary()

	if summary.TotalTests != 3 {
		t.Errorf("Expected 3 total tests, got %d", summary.TotalTests)
	}

	if summary.PassedTests != 1 {
		t.Errorf("Expected 1 passed test, got %d", summary.PassedTests)
	}

	if summary.FailedTests != 1 {
		t.Errorf("Expected 1 failed test, got %d", summary.FailedTests)
	}

	if summary.SkippedTests != 1 {
		t.Errorf("Expected 1 skipped test, got %d", summary.SkippedTests)
	}

	if summary.TotalDuration <= 0 {
		t.Error("Expected positive total duration")
	}

	// Check suite summary
	suiteSummary, exists := summary.Suites["Test Suite"]
	if !exists {
		t.Error("Expected suite summary to exist")
	}

	if suiteSummary.TotalTests != 3 {
		t.Errorf("Expected 3 tests in suite summary, got %d", suiteSummary.TotalTests)
	}

	if suiteSummary.PassedTests != 1 {
		t.Errorf("Expected 1 passed test in suite summary, got %d", suiteSummary.PassedTests)
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test generateTestID
	id1 := generateTestID()
	id2 := generateTestID()
	if id1 == id2 {
		t.Error("Expected unique test IDs")
	}
	if !strings.HasPrefix(id1, "test_") {
		t.Errorf("Expected test ID to start with 'test_', got '%s'", id1)
	}

	// Test generateAssertionID
	assertionID1 := generateAssertionID()
	assertionID2 := generateAssertionID()
	if assertionID1 == assertionID2 {
		t.Error("Expected unique assertion IDs")
	}
	if !strings.HasPrefix(assertionID1, "assertion_") {
		t.Errorf("Expected assertion ID to start with 'assertion_', got '%s'", assertionID1)
	}

	// Test generateCallID
	callID1 := generateCallID()
	callID2 := generateCallID()
	if callID1 == callID2 {
		t.Error("Expected unique call IDs")
	}
	if !strings.HasPrefix(callID1, "call_") {
		t.Errorf("Expected call ID to start with 'call_', got '%s'", callID1)
	}

	// Test getDescription
	desc1 := getDescription([]string{"Custom message"}, "Default message")
	if desc1 != "Custom message" {
		t.Errorf("Expected 'Custom message', got '%s'", desc1)
	}

	desc2 := getDescription([]string{}, "Default message")
	if desc2 != "Default message" {
		t.Errorf("Expected 'Default message', got '%s'", desc2)
	}

	// Test getStackTrace
	stackTrace := getStackTrace()
	if len(stackTrace) == 0 {
		t.Error("Expected non-empty stack trace")
	}
}
