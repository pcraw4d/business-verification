package testing

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestSuite represents a collection of unit tests
type TestSuite struct {
	Name         string
	Description  string
	Tests        []*UnitTest
	SetupFunc    func() error
	TeardownFunc func() error
	BeforeEach   func() error
	AfterEach    func() error
	Parallel     bool
	Timeout      time.Duration
	Tags         []string
	logger       *zap.Logger
	mutex        sync.RWMutex
}

// UnitTest represents a single unit test
type UnitTest struct {
	ID          string
	Name        string
	Description string
	Function    TestFunction
	Skipped     bool
	SkipReason  string
	Tags        []string
	Timeout     time.Duration
	Parallel    bool
	Setup       func() error
	Teardown    func() error
	Expected    interface{}
	Actual      interface{}
	Error       error
	Duration    time.Duration
	Status      TestStatus
	Output      []string
	Assertions  []*Assertion
}

// TestFunction represents a test function
type TestFunction func(t *TestContext) error

// TestStatus represents the status of a test
type TestStatus string

const (
	TestStatusPending TestStatus = "pending"
	TestStatusRunning TestStatus = "running"
	TestStatusPassed  TestStatus = "passed"
	TestStatusFailed  TestStatus = "failed"
	TestStatusSkipped TestStatus = "skipped"
	TestStatusTimeout TestStatus = "timeout"
)

// TestContext provides context and utilities for test execution
type TestContext struct {
	Test     *UnitTest
	Suite    *TestSuite
	T        *testing.T
	Context  context.Context
	Logger   *zap.Logger
	Fixtures *TestFixtures
	Mocks    *MockManager
	timer    *time.Timer
	cancel   context.CancelFunc
}

// Assertion represents a test assertion
type Assertion struct {
	ID          string
	Description string
	Expected    interface{}
	Actual      interface{}
	Operator    string
	Success     bool
	Message     string
	StackTrace  string
}

// TestFixtures provides test data and fixtures
type TestFixtures struct {
	Data    map[string]interface{}
	Files   map[string]string
	Configs map[string]interface{}
	mutex   sync.RWMutex
}

// MockManager manages test mocks
type MockManager struct {
	Mocks     map[string]interface{}
	Calls     map[string][]*MockCall
	Behaviors map[string]*MockBehavior
	mutex     sync.RWMutex
}

// MockCall represents a mock function call
type MockCall struct {
	ID        string
	Function  string
	Args      []interface{}
	Result    []interface{}
	Error     error
	Timestamp time.Time
}

// MockBehavior defines mock behavior
type MockBehavior struct {
	ID          string
	Function    string
	ReturnValue []interface{}
	Error       error
	Delay       time.Duration
	CallCount   int
	MaxCalls    int
}

// TestResult represents the result of test execution
type TestResult struct {
	Suite       *TestSuite
	Test        *UnitTest
	Status      TestStatus
	Duration    time.Duration
	Error       error
	Assertions  []*Assertion
	Output      []string
	Coverage    *TestCoverage
	Performance *TestPerformance
}

// TestCoverage represents test coverage information
type TestCoverage struct {
	TotalLines    int
	CoveredLines  int
	Percentage    float64
	UncoveredCode []string
	Functions     map[string]float64
}

// TestPerformance represents test performance metrics
type TestPerformance struct {
	ExecutionTime  time.Duration
	MemoryUsage    int64
	GoroutineCount int
	AllocCount     uint64
	AllocBytes     uint64
	CPUTime        time.Duration
}

// TestRunner manages test execution
type TestRunner struct {
	suites      []*TestSuite
	results     []*TestResult
	logger      *zap.Logger
	config      *TestConfig
	parallel    bool
	timeout     time.Duration
	coverageOut string
	mutex       sync.RWMutex
}

// TestConfig holds configuration for test execution
type TestConfig struct {
	Parallel        bool          `json:"parallel"`
	Timeout         time.Duration `json:"timeout"`
	MaxGoroutines   int           `json:"max_goroutines"`
	CoverageEnabled bool          `json:"coverage_enabled"`
	CoverageOutput  string        `json:"coverage_output"`
	VerboseOutput   bool          `json:"verbose_output"`
	FailFast        bool          `json:"fail_fast"`
	RetryCount      int           `json:"retry_count"`
	RetryDelay      time.Duration `json:"retry_delay"`
	Tags            []string      `json:"tags"`
	SkipTags        []string      `json:"skip_tags"`
}

// DefaultTestConfig returns default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		Parallel:        true,
		Timeout:         30 * time.Second,
		MaxGoroutines:   runtime.NumCPU(),
		CoverageEnabled: true,
		CoverageOutput:  "coverage.out",
		VerboseOutput:   false,
		FailFast:        false,
		RetryCount:      0,
		RetryDelay:      100 * time.Millisecond,
		Tags:            []string{},
		SkipTags:        []string{},
	}
}

// NewTestSuite creates a new test suite
func NewTestSuite(name, description string) *TestSuite {
	return &TestSuite{
		Name:        name,
		Description: description,
		Tests:       make([]*UnitTest, 0),
		Parallel:    true,
		Timeout:     30 * time.Second,
		Tags:        make([]string, 0),
		logger:      zap.NewNop(),
	}
}

// AddTest adds a test to the suite
func (s *TestSuite) AddTest(test *UnitTest) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Tests = append(s.Tests, test)
}

// CreateTest creates and adds a new test to the suite
func (s *TestSuite) CreateTest(name, description string, fn TestFunction) *UnitTest {
	test := &UnitTest{
		ID:          generateTestID(),
		Name:        name,
		Description: description,
		Function:    fn,
		Status:      TestStatusPending,
		Timeout:     s.Timeout,
		Parallel:    s.Parallel,
		Tags:        make([]string, 0),
		Output:      make([]string, 0),
		Assertions:  make([]*Assertion, 0),
	}
	s.AddTest(test)
	return test
}

// SetSetup sets the suite setup function
func (s *TestSuite) SetSetup(fn func() error) {
	s.SetupFunc = fn
}

// SetTeardown sets the suite teardown function
func (s *TestSuite) SetTeardown(fn func() error) {
	s.TeardownFunc = fn
}

// SetBeforeEach sets the before each test function
func (s *TestSuite) SetBeforeEach(fn func() error) {
	s.BeforeEach = fn
}

// SetAfterEach sets the after each test function
func (s *TestSuite) SetAfterEach(fn func() error) {
	s.AfterEach = fn
}

// AddTag adds a tag to the suite
func (s *TestSuite) AddTag(tag string) {
	s.Tags = append(s.Tags, tag)
}

// SetLogger sets the logger for the suite
func (s *TestSuite) SetLogger(logger *zap.Logger) {
	s.logger = logger
}

// NewUnitTest creates a new unit test
func NewUnitTest(name, description string, fn TestFunction) *UnitTest {
	return &UnitTest{
		ID:          generateTestID(),
		Name:        name,
		Description: description,
		Function:    fn,
		Status:      TestStatusPending,
		Timeout:     30 * time.Second,
		Parallel:    true,
		Tags:        make([]string, 0),
		Output:      make([]string, 0),
		Assertions:  make([]*Assertion, 0),
	}
}

// AddTag adds a tag to the test
func (t *UnitTest) AddTag(tag string) {
	t.Tags = append(t.Tags, tag)
}

// SetTimeout sets the test timeout
func (t *UnitTest) SetTimeout(timeout time.Duration) {
	t.Timeout = timeout
}

// SetParallel sets whether the test can run in parallel
func (t *UnitTest) SetParallel(parallel bool) {
	t.Parallel = parallel
}

// Skip marks the test to be skipped
func (t *UnitTest) Skip(reason string) {
	t.Skipped = true
	t.SkipReason = reason
}

// NewTestContext creates a new test context
func NewTestContext(test *UnitTest, suite *TestSuite, t *testing.T) *TestContext {
	ctx, cancel := context.WithTimeout(context.Background(), test.Timeout)

	return &TestContext{
		Test:     test,
		Suite:    suite,
		T:        t,
		Context:  ctx,
		Logger:   suite.logger,
		Fixtures: NewTestFixtures(),
		Mocks:    NewMockManager(),
		cancel:   cancel,
	}
}

// Cleanup cleans up the test context
func (tc *TestContext) Cleanup() {
	if tc.cancel != nil {
		tc.cancel()
	}
	if tc.timer != nil {
		tc.timer.Stop()
	}
}

// Log logs a message during test execution
func (tc *TestContext) Log(message string) {
	tc.Test.Output = append(tc.Test.Output, message)
	tc.Logger.Info("Test log", zap.String("test", tc.Test.Name), zap.String("message", message))
}

// Logf logs a formatted message during test execution
func (tc *TestContext) Logf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	tc.Log(message)
}

// Assert creates a new assertion
func (tc *TestContext) Assert() *AssertionBuilder {
	return NewAssertionBuilder(tc)
}

// AssertionBuilder provides fluent assertion API
type AssertionBuilder struct {
	context *TestContext
}

// NewAssertionBuilder creates a new assertion builder
func NewAssertionBuilder(context *TestContext) *AssertionBuilder {
	return &AssertionBuilder{context: context}
}

// Equal asserts that two values are equal
func (ab *AssertionBuilder) Equal(expected, actual interface{}, message ...string) bool {
	assertion := &Assertion{
		ID:          generateAssertionID(),
		Description: getDescription(message, "Expected values to be equal"),
		Expected:    expected,
		Actual:      actual,
		Operator:    "equal",
		Success:     reflect.DeepEqual(expected, actual),
		StackTrace:  getStackTrace(),
	}

	if !assertion.Success {
		assertion.Message = fmt.Sprintf("Expected %v, but got %v", expected, actual)
	}

	ab.context.Test.Assertions = append(ab.context.Test.Assertions, assertion)

	if !assertion.Success {
		// Log assertion failure for debugging (not as error to avoid failing outer test)
		ab.context.Logger.Warn("Assertion failed",
			zap.String("test", ab.context.Test.Name),
			zap.String("description", assertion.Description),
			zap.String("message", assertion.Message))
	}

	return assertion.Success
}

// NotEqual asserts that two values are not equal
func (ab *AssertionBuilder) NotEqual(expected, actual interface{}, message ...string) bool {
	assertion := &Assertion{
		ID:          generateAssertionID(),
		Description: getDescription(message, "Expected values to be not equal"),
		Expected:    expected,
		Actual:      actual,
		Operator:    "not_equal",
		Success:     !reflect.DeepEqual(expected, actual),
		StackTrace:  getStackTrace(),
	}

	if !assertion.Success {
		assertion.Message = fmt.Sprintf("Expected %v to not equal %v", expected, actual)
	}

	ab.context.Test.Assertions = append(ab.context.Test.Assertions, assertion)

	if !assertion.Success {
		// Log assertion failure for debugging (not as error to avoid failing outer test)
		ab.context.Logger.Warn("Assertion failed",
			zap.String("test", ab.context.Test.Name),
			zap.String("description", assertion.Description),
			zap.String("message", assertion.Message))
	}

	return assertion.Success
}

// True asserts that a value is true
func (ab *AssertionBuilder) True(value bool, message ...string) bool {
	assertion := &Assertion{
		ID:          generateAssertionID(),
		Description: getDescription(message, "Expected value to be true"),
		Expected:    true,
		Actual:      value,
		Operator:    "true",
		Success:     value,
		StackTrace:  getStackTrace(),
	}

	if !assertion.Success {
		assertion.Message = fmt.Sprintf("Expected true, but got %v", value)
	}

	ab.context.Test.Assertions = append(ab.context.Test.Assertions, assertion)

	if !assertion.Success {
		// Log assertion failure for debugging (not as error to avoid failing outer test)
		ab.context.Logger.Warn("Assertion failed",
			zap.String("test", ab.context.Test.Name),
			zap.String("description", assertion.Description),
			zap.String("message", assertion.Message))
	}

	return assertion.Success
}

// False asserts that a value is false
func (ab *AssertionBuilder) False(value bool, message ...string) bool {
	assertion := &Assertion{
		ID:          generateAssertionID(),
		Description: getDescription(message, "Expected value to be false"),
		Expected:    false,
		Actual:      value,
		Operator:    "false",
		Success:     !value,
		StackTrace:  getStackTrace(),
	}

	if !assertion.Success {
		assertion.Message = fmt.Sprintf("Expected false, but got %v", value)
	}

	ab.context.Test.Assertions = append(ab.context.Test.Assertions, assertion)

	if !assertion.Success {
		// Log assertion failure for debugging (not as error to avoid failing outer test)
		ab.context.Logger.Warn("Assertion failed",
			zap.String("test", ab.context.Test.Name),
			zap.String("description", assertion.Description),
			zap.String("message", assertion.Message))
	}

	return assertion.Success
}

// Nil asserts that a value is nil
func (ab *AssertionBuilder) Nil(value interface{}, message ...string) bool {
	assertion := &Assertion{
		ID:          generateAssertionID(),
		Description: getDescription(message, "Expected value to be nil"),
		Expected:    nil,
		Actual:      value,
		Operator:    "nil",
		Success:     value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()),
		StackTrace:  getStackTrace(),
	}

	if !assertion.Success {
		assertion.Message = fmt.Sprintf("Expected nil, but got %v", value)
	}

	ab.context.Test.Assertions = append(ab.context.Test.Assertions, assertion)

	if !assertion.Success {
		// Log assertion failure for debugging (not as error to avoid failing outer test)
		ab.context.Logger.Warn("Assertion failed",
			zap.String("test", ab.context.Test.Name),
			zap.String("description", assertion.Description),
			zap.String("message", assertion.Message))
	}

	return assertion.Success
}

// NotNil asserts that a value is not nil
func (ab *AssertionBuilder) NotNil(value interface{}, message ...string) bool {
	assertion := &Assertion{
		ID:          generateAssertionID(),
		Description: getDescription(message, "Expected value to be not nil"),
		Expected:    "not nil",
		Actual:      value,
		Operator:    "not_nil",
		Success:     value != nil && !(reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()),
		StackTrace:  getStackTrace(),
	}

	if !assertion.Success {
		assertion.Message = "Expected value to be not nil, but got nil"
	}

	ab.context.Test.Assertions = append(ab.context.Test.Assertions, assertion)

	if !assertion.Success {
		// Log assertion failure for debugging (not as error to avoid failing outer test)
		ab.context.Logger.Warn("Assertion failed",
			zap.String("test", ab.context.Test.Name),
			zap.String("description", assertion.Description),
			zap.String("message", assertion.Message))
	}

	return assertion.Success
}

// Error asserts that an error occurred
func (ab *AssertionBuilder) Error(err error, message ...string) bool {
	assertion := &Assertion{
		ID:          generateAssertionID(),
		Description: getDescription(message, "Expected an error to occur"),
		Expected:    "error",
		Actual:      err,
		Operator:    "error",
		Success:     err != nil,
		StackTrace:  getStackTrace(),
	}

	if !assertion.Success {
		assertion.Message = "Expected an error, but got nil"
	}

	ab.context.Test.Assertions = append(ab.context.Test.Assertions, assertion)

	if !assertion.Success {
		// Log assertion failure for debugging (not as error to avoid failing outer test)
		ab.context.Logger.Warn("Assertion failed",
			zap.String("test", ab.context.Test.Name),
			zap.String("description", assertion.Description),
			zap.String("message", assertion.Message))
	}

	return assertion.Success
}

// NoError asserts that no error occurred
func (ab *AssertionBuilder) NoError(err error, message ...string) bool {
	assertion := &Assertion{
		ID:          generateAssertionID(),
		Description: getDescription(message, "Expected no error to occur"),
		Expected:    nil,
		Actual:      err,
		Operator:    "no_error",
		Success:     err == nil,
		StackTrace:  getStackTrace(),
	}

	if !assertion.Success {
		assertion.Message = fmt.Sprintf("Expected no error, but got: %v", err)
	}

	ab.context.Test.Assertions = append(ab.context.Test.Assertions, assertion)

	if !assertion.Success {
		// Log assertion failure for debugging (not as error to avoid failing outer test)
		ab.context.Logger.Warn("Assertion failed",
			zap.String("test", ab.context.Test.Name),
			zap.String("description", assertion.Description),
			zap.String("message", assertion.Message))
	}

	return assertion.Success
}

// Contains asserts that a string contains a substring
func (ab *AssertionBuilder) Contains(str, substr string, message ...string) bool {
	assertion := &Assertion{
		ID:          generateAssertionID(),
		Description: getDescription(message, "Expected string to contain substring"),
		Expected:    substr,
		Actual:      str,
		Operator:    "contains",
		Success:     strings.Contains(str, substr),
		StackTrace:  getStackTrace(),
	}

	if !assertion.Success {
		assertion.Message = fmt.Sprintf("Expected '%s' to contain '%s'", str, substr)
	}

	ab.context.Test.Assertions = append(ab.context.Test.Assertions, assertion)

	if !assertion.Success {
		// Log assertion failure for debugging (not as error to avoid failing outer test)
		ab.context.Logger.Warn("Assertion failed",
			zap.String("test", ab.context.Test.Name),
			zap.String("description", assertion.Description),
			zap.String("message", assertion.Message))
	}

	return assertion.Success
}

// Len asserts that a slice, map, or string has a specific length
func (ab *AssertionBuilder) Len(obj interface{}, expectedLen int, message ...string) bool {
	v := reflect.ValueOf(obj)
	var actualLen int

	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String, reflect.Chan:
		actualLen = v.Len()
	default:
		assertion := &Assertion{
			ID:          generateAssertionID(),
			Description: getDescription(message, "Expected object to have length"),
			Expected:    expectedLen,
			Actual:      obj,
			Operator:    "len",
			Success:     false,
			Message:     fmt.Sprintf("Object of type %T does not have a length", obj),
			StackTrace:  getStackTrace(),
		}
		ab.context.Test.Assertions = append(ab.context.Test.Assertions, assertion)
		// Log assertion failure for debugging (not as error to avoid failing outer test)
		ab.context.Logger.Warn("Assertion failed",
			zap.String("test", ab.context.Test.Name),
			zap.String("description", assertion.Description),
			zap.String("message", assertion.Message))
		return false
	}

	assertion := &Assertion{
		ID:          generateAssertionID(),
		Description: getDescription(message, "Expected object to have specific length"),
		Expected:    expectedLen,
		Actual:      actualLen,
		Operator:    "len",
		Success:     actualLen == expectedLen,
		StackTrace:  getStackTrace(),
	}

	if !assertion.Success {
		assertion.Message = fmt.Sprintf("Expected length %d, but got %d", expectedLen, actualLen)
	}

	ab.context.Test.Assertions = append(ab.context.Test.Assertions, assertion)

	if !assertion.Success {
		// Log assertion failure for debugging (not as error to avoid failing outer test)
		ab.context.Logger.Warn("Assertion failed",
			zap.String("test", ab.context.Test.Name),
			zap.String("description", assertion.Description),
			zap.String("message", assertion.Message))
	}

	return assertion.Success
}

// NewTestFixtures creates a new test fixtures manager
func NewTestFixtures() *TestFixtures {
	return &TestFixtures{
		Data:    make(map[string]interface{}),
		Files:   make(map[string]string),
		Configs: make(map[string]interface{}),
	}
}

// SetData sets test data
func (tf *TestFixtures) SetData(key string, value interface{}) {
	tf.mutex.Lock()
	defer tf.mutex.Unlock()
	tf.Data[key] = value
}

// GetData gets test data
func (tf *TestFixtures) GetData(key string) (interface{}, bool) {
	tf.mutex.RLock()
	defer tf.mutex.RUnlock()
	value, exists := tf.Data[key]
	return value, exists
}

// SetFile sets test file content
func (tf *TestFixtures) SetFile(filename, content string) {
	tf.mutex.Lock()
	defer tf.mutex.Unlock()
	tf.Files[filename] = content
}

// GetFile gets test file content
func (tf *TestFixtures) GetFile(filename string) (string, bool) {
	tf.mutex.RLock()
	defer tf.mutex.RUnlock()
	content, exists := tf.Files[filename]
	return content, exists
}

// SetConfig sets test configuration
func (tf *TestFixtures) SetConfig(key string, config interface{}) {
	tf.mutex.Lock()
	defer tf.mutex.Unlock()
	tf.Configs[key] = config
}

// GetConfig gets test configuration
func (tf *TestFixtures) GetConfig(key string) (interface{}, bool) {
	tf.mutex.RLock()
	defer tf.mutex.RUnlock()
	config, exists := tf.Configs[key]
	return config, exists
}

// NewMockManager creates a new mock manager
func NewMockManager() *MockManager {
	return &MockManager{
		Mocks:     make(map[string]interface{}),
		Calls:     make(map[string][]*MockCall),
		Behaviors: make(map[string]*MockBehavior),
	}
}

// AddMock adds a mock to the manager
func (mm *MockManager) AddMock(name string, mock interface{}) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	mm.Mocks[name] = mock
}

// GetMock gets a mock from the manager
func (mm *MockManager) GetMock(name string) (interface{}, bool) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	mock, exists := mm.Mocks[name]
	return mock, exists
}

// RecordCall records a mock call
func (mm *MockManager) RecordCall(function string, args []interface{}, result []interface{}, err error) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	call := &MockCall{
		ID:        generateCallID(),
		Function:  function,
		Args:      args,
		Result:    result,
		Error:     err,
		Timestamp: time.Now(),
	}

	mm.Calls[function] = append(mm.Calls[function], call)
}

// GetCalls gets recorded calls for a function
func (mm *MockManager) GetCalls(function string) []*MockCall {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	return mm.Calls[function]
}

// SetBehavior sets mock behavior
func (mm *MockManager) SetBehavior(function string, behavior *MockBehavior) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	mm.Behaviors[function] = behavior
}

// GetBehavior gets mock behavior
func (mm *MockManager) GetBehavior(function string) (*MockBehavior, bool) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	behavior, exists := mm.Behaviors[function]
	return behavior, exists
}

// NewTestRunner creates a new test runner
func NewTestRunner(config *TestConfig, logger *zap.Logger) *TestRunner {
	if config == nil {
		config = DefaultTestConfig()
	}

	return &TestRunner{
		suites:      make([]*TestSuite, 0),
		results:     make([]*TestResult, 0),
		logger:      logger,
		config:      config,
		parallel:    config.Parallel,
		timeout:     config.Timeout,
		coverageOut: config.CoverageOutput,
	}
}

// AddSuite adds a test suite to the runner
func (tr *TestRunner) AddSuite(suite *TestSuite) {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()
	tr.suites = append(tr.suites, suite)
}

// RunAllSuites runs all test suites
func (tr *TestRunner) RunAllSuites(t *testing.T) []*TestResult {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()

	tr.results = make([]*TestResult, 0)

	for _, suite := range tr.suites {
		results := tr.runSuite(t, suite)
		tr.results = append(tr.results, results...)
	}

	return tr.results
}

// runSuite runs a single test suite
func (tr *TestRunner) runSuite(t *testing.T, suite *TestSuite) []*TestResult {
	results := make([]*TestResult, 0)

	// Run suite setup
	if suite.SetupFunc != nil {
		if err := suite.SetupFunc(); err != nil {
			tr.logger.Error("Suite setup failed", zap.String("suite", suite.Name), zap.Error(err))
			return results
		}
	}

	// Run tests
	if tr.parallel && suite.Parallel {
		results = tr.runTestsParallel(t, suite)
	} else {
		results = tr.runTestsSequential(t, suite)
	}

	// Run suite teardown
	if suite.TeardownFunc != nil {
		if err := suite.TeardownFunc(); err != nil {
			tr.logger.Error("Suite teardown failed", zap.String("suite", suite.Name), zap.Error(err))
		}
	}

	return results
}

// runTestsParallel runs tests in parallel
func (tr *TestRunner) runTestsParallel(t *testing.T, suite *TestSuite) []*TestResult {
	results := make([]*TestResult, len(suite.Tests))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, tr.config.MaxGoroutines)

	for i, test := range suite.Tests {
		if tr.shouldSkipTest(test) {
			results[i] = &TestResult{
				Suite:  suite,
				Test:   test,
				Status: TestStatusSkipped,
			}
			continue
		}

		wg.Add(1)
		go func(idx int, test *UnitTest) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			results[idx] = tr.runSingleTest(t, suite, test)
		}(i, test)
	}

	wg.Wait()
	return results
}

// runTestsSequential runs tests sequentially
func (tr *TestRunner) runTestsSequential(t *testing.T, suite *TestSuite) []*TestResult {
	results := make([]*TestResult, 0)

	for _, test := range suite.Tests {
		if tr.shouldSkipTest(test) {
			results = append(results, &TestResult{
				Suite:  suite,
				Test:   test,
				Status: TestStatusSkipped,
			})
			continue
		}

		result := tr.runSingleTest(t, suite, test)
		results = append(results, result)

		if tr.config.FailFast && result.Status == TestStatusFailed {
			break
		}
	}

	return results
}

// runSingleTest runs a single test
func (tr *TestRunner) runSingleTest(t *testing.T, suite *TestSuite, test *UnitTest) *TestResult {
	result := &TestResult{
		Suite:       suite,
		Test:        test,
		Status:      TestStatusRunning,
		Assertions:  make([]*Assertion, 0),
		Output:      make([]string, 0),
		Coverage:    &TestCoverage{},
		Performance: &TestPerformance{},
	}

	start := time.Now()
	test.Status = TestStatusRunning

	// Create test context
	testCtx := NewTestContext(test, suite, t)
	defer testCtx.Cleanup()

	// Run before each
	if suite.BeforeEach != nil {
		if err := suite.BeforeEach(); err != nil {
			result.Status = TestStatusFailed
			result.Error = fmt.Errorf("before each failed: %w", err)
			return result
		}
	}

	// Run test setup
	if test.Setup != nil {
		if err := test.Setup(); err != nil {
			result.Status = TestStatusFailed
			result.Error = fmt.Errorf("test setup failed: %w", err)
			return result
		}
	}

	// Run the test function
	err := tr.executeTestFunction(testCtx, test)

	// Run test teardown
	if test.Teardown != nil {
		if teardownErr := test.Teardown(); teardownErr != nil {
			tr.logger.Warn("Test teardown failed", zap.String("test", test.Name), zap.Error(teardownErr))
		}
	}

	// Run after each
	if suite.AfterEach != nil {
		if afterErr := suite.AfterEach(); afterErr != nil {
			tr.logger.Warn("After each failed", zap.String("test", test.Name), zap.Error(afterErr))
		}
	}

	// Calculate duration and status
	duration := time.Since(start)
	result.Duration = duration
	test.Duration = duration

	if err != nil {
		result.Status = TestStatusFailed
		result.Error = err
		test.Status = TestStatusFailed
		test.Error = err
	} else {
		// Check if any assertions failed
		failed := false
		for _, assertion := range test.Assertions {
			if !assertion.Success {
				failed = true
				break
			}
		}

		if failed {
			result.Status = TestStatusFailed
			test.Status = TestStatusFailed
		} else {
			result.Status = TestStatusPassed
			test.Status = TestStatusPassed
		}
	}

	result.Assertions = test.Assertions
	result.Output = test.Output

	return result
}

// executeTestFunction executes the test function with timeout
func (tr *TestRunner) executeTestFunction(testCtx *TestContext, test *UnitTest) error {
	done := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("test panicked: %v", r)
			}
		}()
		done <- test.Function(testCtx)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(test.Timeout):
		test.Status = TestStatusTimeout
		return fmt.Errorf("test timed out after %v", test.Timeout)
	case <-testCtx.Context.Done():
		return fmt.Errorf("test context cancelled: %w", testCtx.Context.Err())
	}
}

// shouldSkipTest determines if a test should be skipped
func (tr *TestRunner) shouldSkipTest(test *UnitTest) bool {
	if test.Skipped {
		return true
	}

	// Check tag filters
	if len(tr.config.Tags) > 0 {
		hasRequiredTag := false
		for _, tag := range tr.config.Tags {
			for _, testTag := range test.Tags {
				if tag == testTag {
					hasRequiredTag = true
					break
				}
			}
			if hasRequiredTag {
				break
			}
		}
		if !hasRequiredTag {
			return true
		}
	}

	// Check skip tags
	for _, skipTag := range tr.config.SkipTags {
		for _, testTag := range test.Tags {
			if skipTag == testTag {
				return true
			}
		}
	}

	return false
}

// GetResults returns test results
func (tr *TestRunner) GetResults() []*TestResult {
	tr.mutex.RLock()
	defer tr.mutex.RUnlock()
	return tr.results
}

// GenerateSummary generates a test execution summary
func (tr *TestRunner) GenerateSummary() *TestSummary {
	tr.mutex.RLock()
	defer tr.mutex.RUnlock()

	summary := &TestSummary{
		TotalTests:    len(tr.results),
		PassedTests:   0,
		FailedTests:   0,
		SkippedTests:  0,
		TimeoutTests:  0,
		TotalDuration: 0,
		Suites:        make(map[string]*SuiteSummary),
	}

	for _, result := range tr.results {
		summary.TotalDuration += result.Duration

		switch result.Status {
		case TestStatusPassed:
			summary.PassedTests++
		case TestStatusFailed:
			summary.FailedTests++
		case TestStatusSkipped:
			summary.SkippedTests++
		case TestStatusTimeout:
			summary.TimeoutTests++
		}

		// Update suite summary
		suiteName := result.Suite.Name
		if _, exists := summary.Suites[suiteName]; !exists {
			summary.Suites[suiteName] = &SuiteSummary{
				Name:         suiteName,
				TotalTests:   0,
				PassedTests:  0,
				FailedTests:  0,
				SkippedTests: 0,
				Duration:     0,
			}
		}

		suiteSummary := summary.Suites[suiteName]
		suiteSummary.TotalTests++
		suiteSummary.Duration += result.Duration

		switch result.Status {
		case TestStatusPassed:
			suiteSummary.PassedTests++
		case TestStatusFailed:
			suiteSummary.FailedTests++
		case TestStatusSkipped:
			suiteSummary.SkippedTests++
		}
	}

	return summary
}

// TestSummary represents a summary of test execution
type TestSummary struct {
	TotalTests    int                      `json:"total_tests"`
	PassedTests   int                      `json:"passed_tests"`
	FailedTests   int                      `json:"failed_tests"`
	SkippedTests  int                      `json:"skipped_tests"`
	TimeoutTests  int                      `json:"timeout_tests"`
	TotalDuration time.Duration            `json:"total_duration"`
	Suites        map[string]*SuiteSummary `json:"suites"`
}

// SuiteSummary represents a summary of test suite execution
type SuiteSummary struct {
	Name         string        `json:"name"`
	TotalTests   int           `json:"total_tests"`
	PassedTests  int           `json:"passed_tests"`
	FailedTests  int           `json:"failed_tests"`
	SkippedTests int           `json:"skipped_tests"`
	Duration     time.Duration `json:"duration"`
}

// Helper functions
var (
	testIDCounter      int64
	assertionIDCounter int64
	callIDCounter      int64
	idMutex            sync.Mutex
)

func generateTestID() string {
	idMutex.Lock()
	defer idMutex.Unlock()
	testIDCounter++
	return fmt.Sprintf("test_%d_%d", time.Now().UnixNano(), testIDCounter)
}

func generateAssertionID() string {
	idMutex.Lock()
	defer idMutex.Unlock()
	assertionIDCounter++
	return fmt.Sprintf("assertion_%d_%d", time.Now().UnixNano(), assertionIDCounter)
}

func generateCallID() string {
	idMutex.Lock()
	defer idMutex.Unlock()
	callIDCounter++
	return fmt.Sprintf("call_%d_%d", time.Now().UnixNano(), callIDCounter)
}

func getDescription(messages []string, defaultDesc string) string {
	if len(messages) > 0 {
		return messages[0]
	}
	return defaultDesc
}

func getStackTrace() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}
