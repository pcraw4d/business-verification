package testing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// E2ETest represents a single end-to-end test
type E2ETest struct {
	ID          string
	Name        string
	Description string
	Function    E2ETestFunction
	Config      E2ETestConfig
	Tags        []string
	Components  []string
	Timeout     time.Duration
	Parallel    bool
	Skipped     bool
	Priority    E2ETestPriority
	Category    E2ETestCategory
}

// E2ETestFunction defines the signature for end-to-end test functions
type E2ETestFunction func(ctx *E2EContext) error

// E2ETestConfig defines configuration for end-to-end tests
type E2ETestConfig struct {
	Environment     string            // Target environment (dev, staging, prod)
	UserJourney     string            // User journey being tested
	DataSetup       []string          // Data setup requirements
	CleanupRequired bool              // Whether cleanup is required after test
	RetryCount      int               // Number of retry attempts
	RetryDelay      time.Duration     // Delay between retries
	Assertions      []E2EAssertion    // Test assertions
	Checkpoints     []E2ECheckpoint   // Test checkpoints
	Metadata        map[string]string // Additional test metadata
}

// E2EAssertion represents an assertion in an E2E test
type E2EAssertion struct {
	Name        string
	Description string
	Condition   func(ctx *E2EContext) bool
	Message     string
	Critical    bool // Whether failure should stop the test
}

// E2ECheckpoint represents a checkpoint in an E2E test
type E2ECheckpoint struct {
	Name        string
	Description string
	Validate    func(ctx *E2EContext) error
	Required    bool // Whether checkpoint is required to pass
}

// E2ETestPriority represents the priority of an E2E test
type E2ETestPriority string

const (
	PriorityCritical E2ETestPriority = "critical"
	PriorityHigh     E2ETestPriority = "high"
	PriorityMedium   E2ETestPriority = "medium"
	PriorityLow      E2ETestPriority = "low"
)

// E2ETestCategory represents the category of an E2E test
type E2ETestCategory string

const (
	CategoryUserJourney       E2ETestCategory = "user_journey"
	CategoryBusinessFlow      E2ETestCategory = "business_flow"
	CategorySystemIntegration E2ETestCategory = "system_integration"
	CategoryDataFlow          E2ETestCategory = "data_flow"
	CategorySecurity          E2ETestCategory = "security"
	CategoryPerformance       E2ETestCategory = "performance"
	CategoryAccessibility     E2ETestCategory = "accessibility"
	CategoryCompatibility     E2ETestCategory = "compatibility"
)

// E2EContext provides context for end-to-end tests
type E2EContext struct {
	context.Context
	Logger       *zap.Logger
	T            *E2ETest
	StartTime    time.Time
	EndTime      time.Time
	Environment  *E2EEnvironment
	UserJourney  *E2EUserJourney
	DataManager  *E2EDataManager
	Checkpoints  []*E2ECheckpointResult
	Assertions   []*E2EAssertionResult
	CleanupFuncs []func()
	cancel       context.CancelFunc
	mu           sync.Mutex
}

// E2EEnvironment represents the test environment
type E2EEnvironment struct {
	Name         string
	BaseURL      string
	APIEndpoint  string
	DatabaseURL  string
	Credentials  map[string]string
	Config       map[string]interface{}
	Capabilities map[string]bool
}

// E2EUserJourney represents a user journey being tested
type E2EUserJourney struct {
	Name        string
	Description string
	Steps       []E2EJourneyStep
	Data        map[string]interface{}
	State       map[string]interface{}
}

// E2EJourneyStep represents a step in a user journey
type E2EJourneyStep struct {
	Name        string
	Description string
	Action      func(ctx *E2EContext) error
	Validation  func(ctx *E2EContext) error
	Required    bool
}

// E2EDataManager manages test data for E2E tests
type E2EDataManager struct {
	TestData    map[string]interface{}
	SetupData   map[string]interface{}
	CleanupData map[string]interface{}
	mu          sync.RWMutex
}

// E2ECheckpointResult represents the result of a checkpoint
type E2ECheckpointResult struct {
	Checkpoint *E2ECheckpoint
	Passed     bool
	Error      error
	Timestamp  time.Time
	Duration   time.Duration
}

// E2EAssertionResult represents the result of an assertion
type E2EAssertionResult struct {
	Assertion *E2EAssertion
	Passed    bool
	Error     error
	Timestamp time.Time
	Duration  time.Duration
}

// E2ETestSuite represents a collection of end-to-end tests
type E2ETestSuite struct {
	Name        string
	Description string
	Tests       []*E2ETest
	Setup       func(ctx *E2EContext) error
	Teardown    func(ctx *E2EContext) error
	BeforeEach  func(ctx *E2EContext) error
	AfterEach   func(ctx *E2EContext) error
	Parallel    bool
	Timeout     time.Duration
	Tags        []string
	Components  []string
	Config      E2ETestConfig
	Environment *E2EEnvironment
}

// E2ETestRunner manages end-to-end test execution
type E2ETestRunner struct {
	Suites      []*E2ETestSuite
	Config      E2ETestConfig
	Logger      *zap.Logger
	Results     []*E2ETestResult
	Environment *E2EEnvironment
	mu          sync.Mutex
}

// E2ETestResult represents the result of an end-to-end test
type E2ETestResult struct {
	Test        *E2ETest
	Status      E2ETestStatus
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Checkpoints []*E2ECheckpointResult
	Assertions  []*E2EAssertionResult
	UserJourney *E2EUserJourney
	Environment *E2EEnvironment
	Passed      bool
	Error       error
	RetryCount  int
	RetryErrors []error
	Metadata    map[string]interface{}
}

// E2ETestStatus represents the status of an end-to-end test
type E2ETestStatus string

const (
	E2EStatusPassed   E2ETestStatus = "passed"
	E2EStatusFailed   E2ETestStatus = "failed"
	E2EStatusSkipped  E2ETestStatus = "skipped"
	E2EStatusError    E2ETestStatus = "error"
	E2EStatusRetrying E2ETestStatus = "retrying"
	E2EStatusTimeout  E2ETestStatus = "timeout"
)

// NewE2ETest creates a new end-to-end test
func NewE2ETest(name string, function E2ETestFunction) *E2ETest {
	return &E2ETest{
		ID:         generateE2ETestID(),
		Name:       name,
		Function:   function,
		Config:     DefaultE2ETestConfig(),
		Tags:       []string{},
		Components: []string{},
		Timeout:    5 * time.Minute,
		Parallel:   false,
		Skipped:    false,
		Priority:   PriorityMedium,
		Category:   CategoryUserJourney,
	}
}

// DefaultE2ETestConfig returns default E2E test configuration
func DefaultE2ETestConfig() E2ETestConfig {
	return E2ETestConfig{
		Environment:     "staging",
		UserJourney:     "",
		DataSetup:       []string{},
		CleanupRequired: true,
		RetryCount:      2,
		RetryDelay:      5 * time.Second,
		Assertions:      []E2EAssertion{},
		Checkpoints:     []E2ECheckpoint{},
		Metadata:        map[string]string{},
	}
}

// AddTag adds a tag to the E2E test
func (et *E2ETest) AddTag(tag string) *E2ETest {
	et.Tags = append(et.Tags, tag)
	return et
}

// SetConfig sets the E2E test configuration
func (et *E2ETest) SetConfig(config E2ETestConfig) *E2ETest {
	et.Config = config
	return et
}

// SetTimeout sets the test timeout
func (et *E2ETest) SetTimeout(timeout time.Duration) *E2ETest {
	et.Timeout = timeout
	return et
}

// SetParallel sets whether the test can run in parallel
func (et *E2ETest) SetParallel(parallel bool) *E2ETest {
	et.Parallel = parallel
	return et
}

// Skip marks the test as skipped
func (et *E2ETest) Skip() *E2ETest {
	et.Skipped = true
	return et
}

// AddComponent adds a component to the E2E test
func (et *E2ETest) AddComponent(component string) *E2ETest {
	et.Components = append(et.Components, component)
	return et
}

// SetPriority sets the test priority
func (et *E2ETest) SetPriority(priority E2ETestPriority) *E2ETest {
	et.Priority = priority
	return et
}

// SetCategory sets the test category
func (et *E2ETest) SetCategory(category E2ETestCategory) *E2ETest {
	et.Category = category
	return et
}

// AddAssertion adds an assertion to the test
func (et *E2ETest) AddAssertion(assertion E2EAssertion) *E2ETest {
	et.Config.Assertions = append(et.Config.Assertions, assertion)
	return et
}

// AddCheckpoint adds a checkpoint to the test
func (et *E2ETest) AddCheckpoint(checkpoint E2ECheckpoint) *E2ETest {
	et.Config.Checkpoints = append(et.Config.Checkpoints, checkpoint)
	return et
}

// NewE2ETestSuite creates a new E2E test suite
func NewE2ETestSuite(name string) *E2ETestSuite {
	return &E2ETestSuite{
		Name:        name,
		Description: "",
		Tests:       []*E2ETest{},
		Setup:       nil,
		Teardown:    nil,
		BeforeEach:  nil,
		AfterEach:   nil,
		Parallel:    false,
		Timeout:     10 * time.Minute,
		Tags:        []string{},
		Components:  []string{},
		Config:      DefaultE2ETestConfig(),
		Environment: nil,
	}
}

// AddTest adds a test to the suite
func (ets *E2ETestSuite) AddTest(test *E2ETest) *E2ETestSuite {
	ets.Tests = append(ets.Tests, test)
	return ets
}

// CreateTest creates and adds a new test to the suite
func (ets *E2ETestSuite) CreateTest(name string, function E2ETestFunction) *E2ETest {
	test := NewE2ETest(name, function)
	ets.AddTest(test)
	return test
}

// SetSetup sets the suite setup function
func (ets *E2ETestSuite) SetSetup(setup func(ctx *E2EContext) error) *E2ETestSuite {
	ets.Setup = setup
	return ets
}

// SetTeardown sets the suite teardown function
func (ets *E2ETestSuite) SetTeardown(teardown func(ctx *E2EContext) error) *E2ETestSuite {
	ets.Teardown = teardown
	return ets
}

// SetBeforeEach sets the before each function
func (ets *E2ETestSuite) SetBeforeEach(beforeEach func(ctx *E2EContext) error) *E2ETestSuite {
	ets.BeforeEach = beforeEach
	return ets
}

// SetAfterEach sets the after each function
func (ets *E2ETestSuite) SetAfterEach(afterEach func(ctx *E2EContext) error) *E2ETestSuite {
	ets.AfterEach = afterEach
	return ets
}

// SetParallel sets whether tests can run in parallel
func (ets *E2ETestSuite) SetParallel(parallel bool) *E2ETestSuite {
	ets.Parallel = parallel
	return ets
}

// SetTimeout sets the suite timeout
func (ets *E2ETestSuite) SetTimeout(timeout time.Duration) *E2ETestSuite {
	ets.Timeout = timeout
	return ets
}

// AddTag adds a tag to the suite
func (ets *E2ETestSuite) AddTag(tag string) *E2ETestSuite {
	ets.Tags = append(ets.Tags, tag)
	return ets
}

// AddComponent adds a component to the suite
func (ets *E2ETestSuite) AddComponent(component string) *E2ETestSuite {
	ets.Components = append(ets.Components, component)
	return ets
}

// SetConfig sets the suite configuration
func (ets *E2ETestSuite) SetConfig(config E2ETestConfig) *E2ETestSuite {
	ets.Config = config
	return ets
}

// SetEnvironment sets the test environment
func (ets *E2ETestSuite) SetEnvironment(env *E2EEnvironment) *E2ETestSuite {
	ets.Environment = env
	return ets
}

// NewE2EContext creates a new E2E context
func NewE2EContext(ctx context.Context, test *E2ETest, logger *zap.Logger, env *E2EEnvironment) *E2EContext {
	ctx, cancel := context.WithTimeout(ctx, test.Timeout)
	return &E2EContext{
		Context:      ctx,
		Logger:       logger,
		T:            test,
		StartTime:    time.Now(),
		Environment:  env,
		UserJourney:  &E2EUserJourney{},
		DataManager:  NewE2EDataManager(),
		Checkpoints:  []*E2ECheckpointResult{},
		Assertions:   []*E2EAssertionResult{},
		CleanupFuncs: []func(){},
		cancel:       cancel,
		mu:           sync.Mutex{},
	}
}

// NewE2EDataManager creates a new E2E data manager
func NewE2EDataManager() *E2EDataManager {
	return &E2EDataManager{
		TestData:    map[string]interface{}{},
		SetupData:   map[string]interface{}{},
		CleanupData: map[string]interface{}{},
		mu:          sync.RWMutex{},
	}
}

// Cleanup performs cleanup operations
func (ec *E2EContext) Cleanup() {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	// Execute cleanup functions in reverse order
	for i := len(ec.CleanupFuncs) - 1; i >= 0; i-- {
		if ec.CleanupFuncs[i] != nil {
			ec.CleanupFuncs[i]()
		}
	}

	if ec.cancel != nil {
		ec.cancel()
	}

	ec.EndTime = time.Now()
}

// AddCleanup adds a cleanup function
func (ec *E2EContext) AddCleanup(cleanup func()) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.CleanupFuncs = append(ec.CleanupFuncs, cleanup)
}

// Log logs a message
func (ec *E2EContext) Log(msg string, fields ...zap.Field) {
	if ec.Logger != nil {
		ec.Logger.Info(msg, fields...)
	}
}

// Logf logs a formatted message
func (ec *E2EContext) Logf(format string, args ...interface{}) {
	if ec.Logger != nil {
		ec.Logger.Sugar().Infof(format, args...)
	}
}

// SetUserJourney sets the user journey for the test
func (ec *E2EContext) SetUserJourney(journey *E2EUserJourney) {
	ec.UserJourney = journey
}

// AddJourneyStep adds a step to the user journey
func (ec *E2EContext) AddJourneyStep(step E2EJourneyStep) {
	ec.UserJourney.Steps = append(ec.UserJourney.Steps, step)
}

// ExecuteJourney executes the user journey
func (ec *E2EContext) ExecuteJourney() error {
	for i, step := range ec.UserJourney.Steps {
		ec.Log("Executing journey step",
			zap.String("step_name", step.Name),
			zap.Int("step_number", i+1),
			zap.Int("total_steps", len(ec.UserJourney.Steps)))

		// Execute step action
		if step.Action != nil {
			if err := step.Action(ec); err != nil {
				if step.Required {
					return fmt.Errorf("required step '%s' failed: %w", step.Name, err)
				}
				ec.Log("Optional step failed, continuing",
					zap.String("step_name", step.Name),
					zap.Error(err))
			}
		}

		// Execute step validation
		if step.Validation != nil {
			if err := step.Validation(ec); err != nil {
				if step.Required {
					return fmt.Errorf("required step validation '%s' failed: %w", step.Name, err)
				}
				ec.Log("Optional step validation failed, continuing",
					zap.String("step_name", step.Name),
					zap.Error(err))
			}
		}
	}

	return nil
}

// SetTestData sets test data
func (dm *E2EDataManager) SetTestData(key string, value interface{}) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.TestData[key] = value
}

// GetTestData gets test data
func (dm *E2EDataManager) GetTestData(key string) (interface{}, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	value, exists := dm.TestData[key]
	return value, exists
}

// SetSetupData sets setup data
func (dm *E2EDataManager) SetSetupData(key string, value interface{}) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.SetupData[key] = value
}

// GetSetupData gets setup data
func (dm *E2EDataManager) GetSetupData(key string) (interface{}, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	value, exists := dm.SetupData[key]
	return value, exists
}

// AddCleanupData adds cleanup data
func (dm *E2EDataManager) AddCleanupData(key string, value interface{}) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.CleanupData[key] = value
}

// GetCleanupData gets cleanup data
func (dm *E2EDataManager) GetCleanupData(key string) (interface{}, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	value, exists := dm.CleanupData[key]
	return value, exists
}

// RunCheckpoint runs a checkpoint
func (ec *E2EContext) RunCheckpoint(checkpoint *E2ECheckpoint) *E2ECheckpointResult {
	startTime := time.Now()

	result := &E2ECheckpointResult{
		Checkpoint: checkpoint,
		Timestamp:  startTime,
	}

	if checkpoint.Validate != nil {
		if err := checkpoint.Validate(ec); err != nil {
			result.Passed = false
			result.Error = err
		} else {
			result.Passed = true
		}
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(startTime)
	ec.Checkpoints = append(ec.Checkpoints, result)

	return result
}

// RunAssertion runs an assertion
func (ec *E2EContext) RunAssertion(assertion *E2EAssertion) *E2EAssertionResult {
	startTime := time.Now()

	result := &E2EAssertionResult{
		Assertion: assertion,
		Timestamp: startTime,
	}

	if assertion.Condition != nil {
		if passed := assertion.Condition(ec); !passed {
			result.Passed = false
			if assertion.Message != "" {
				result.Error = fmt.Errorf("%s", assertion.Message)
			} else {
				result.Error = fmt.Errorf("assertion '%s' failed", assertion.Name)
			}
		} else {
			result.Passed = true
		}
	} else {
		result.Passed = true
	}

	result.Duration = time.Since(startTime)
	ec.Assertions = append(ec.Assertions, result)

	return result
}

// NewE2ETestRunner creates a new E2E test runner
func NewE2ETestRunner(logger *zap.Logger, env *E2EEnvironment) *E2ETestRunner {
	return &E2ETestRunner{
		Suites:      []*E2ETestSuite{},
		Config:      DefaultE2ETestConfig(),
		Logger:      logger,
		Results:     []*E2ETestResult{},
		Environment: env,
		mu:          sync.Mutex{},
	}
}

// AddSuite adds a test suite to the runner
func (etr *E2ETestRunner) AddSuite(suite *E2ETestSuite) *E2ETestRunner {
	etr.mu.Lock()
	defer etr.mu.Unlock()
	etr.Suites = append(etr.Suites, suite)
	return etr
}

// RunAllSuites runs all test suites
func (etr *E2ETestRunner) RunAllSuites(ctx context.Context) error {
	etr.mu.Lock()
	defer etr.mu.Unlock()

	for _, suite := range etr.Suites {
		if err := etr.runSuite(ctx, suite); err != nil {
			return fmt.Errorf("suite %s failed: %w", suite.Name, err)
		}
	}

	return nil
}

// runSuite runs a single test suite
func (etr *E2ETestRunner) runSuite(ctx context.Context, suite *E2ETestSuite) error {
	etr.Logger.Info("Running E2E test suite", zap.String("suite", suite.Name))

	// Create suite context
	suiteCtx, cancel := context.WithTimeout(ctx, suite.Timeout)
	defer cancel()

	// Run setup if provided
	if suite.Setup != nil {
		testCtx := NewE2EContext(suiteCtx, &E2ETest{Name: "setup"}, etr.Logger, suite.Environment)
		if err := suite.Setup(testCtx); err != nil {
			return fmt.Errorf("suite setup failed: %w", err)
		}
		testCtx.Cleanup()
	}

	// Run tests
	if suite.Parallel {
		if err := etr.runTestsParallel(suiteCtx, suite); err != nil {
			return fmt.Errorf("parallel test execution failed: %w", err)
		}
	} else {
		if err := etr.runTestsSequential(suiteCtx, suite); err != nil {
			return fmt.Errorf("sequential test execution failed: %w", err)
		}
	}

	// Run teardown if provided
	if suite.Teardown != nil {
		testCtx := NewE2EContext(suiteCtx, &E2ETest{Name: "teardown"}, etr.Logger, suite.Environment)
		if err := suite.Teardown(testCtx); err != nil {
			etr.Logger.Error("Suite teardown failed", zap.Error(err))
		}
		testCtx.Cleanup()
	}

	return nil
}

// runTestsParallel runs tests in parallel
func (etr *E2ETestRunner) runTestsParallel(ctx context.Context, suite *E2ETestSuite) error {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 3) // Limit concurrent E2E tests

	for _, test := range suite.Tests {
		if test.Skipped {
			continue
		}

		wg.Add(1)
		go func(t *E2ETest) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := etr.runSingleTest(ctx, suite, t); err != nil {
				etr.Logger.Error("Test failed", zap.String("test", t.Name), zap.Error(err))
			}
		}(test)
	}

	wg.Wait()
	return nil
}

// runTestsSequential runs tests sequentially
func (etr *E2ETestRunner) runTestsSequential(ctx context.Context, suite *E2ETestSuite) error {
	for _, test := range suite.Tests {
		if test.Skipped {
			continue
		}

		if err := etr.runSingleTest(ctx, suite, test); err != nil {
			return fmt.Errorf("test %s failed: %w", test.Name, err)
		}
	}

	return nil
}

// runSingleTest runs a single test
func (etr *E2ETestRunner) runSingleTest(ctx context.Context, suite *E2ETestSuite, test *E2ETest) error {
	etr.Logger.Info("Running E2E test", zap.String("test", test.Name))

	// Create test context
	testCtx := NewE2EContext(ctx, test, etr.Logger, suite.Environment)
	defer testCtx.Cleanup()

	// Run before each if provided
	if suite.BeforeEach != nil {
		if err := suite.BeforeEach(testCtx); err != nil {
			return fmt.Errorf("before each failed: %w", err)
		}
	}

	// Execute test function with retries
	result := etr.executeTestFunctionWithRetries(testCtx, test)

	// Run after each if provided
	if suite.AfterEach != nil {
		if err := suite.AfterEach(testCtx); err != nil {
			etr.Logger.Error("After each failed", zap.Error(err))
		}
	}

	// Store result
	etr.mu.Lock()
	etr.Results = append(etr.Results, result)
	etr.mu.Unlock()

	return nil
}

// executeTestFunctionWithRetries executes the test function with retry logic
func (etr *E2ETestRunner) executeTestFunctionWithRetries(ctx *E2EContext, test *E2ETest) *E2ETestResult {
	startTime := time.Now()
	var lastError error
	var retryErrors []error

	// Execute test with retries
	for attempt := 0; attempt <= test.Config.RetryCount; attempt++ {
		if attempt > 0 {
			etr.Logger.Info("Retrying E2E test",
				zap.String("test", test.Name),
				zap.Int("attempt", attempt+1),
				zap.Int("max_attempts", test.Config.RetryCount+1))

			// Wait before retry
			time.Sleep(test.Config.RetryDelay)
		}

		// Execute test function
		err := test.Function(ctx)

		if err == nil {
			// Test passed
			endTime := time.Now()
			duration := endTime.Sub(startTime)

			result := &E2ETestResult{
				Test:        test,
				Status:      E2EStatusPassed,
				StartTime:   startTime,
				EndTime:     endTime,
				Duration:    duration,
				Checkpoints: ctx.Checkpoints,
				Assertions:  ctx.Assertions,
				UserJourney: ctx.UserJourney,
				Environment: ctx.Environment,
				Passed:      true,
				RetryCount:  attempt,
				RetryErrors: retryErrors,
				Metadata:    map[string]interface{}{},
			}

			// Evaluate checkpoints and assertions
			result.Passed = etr.evaluateTestResult(result)

			return result
		}

		// Test failed
		lastError = err
		retryErrors = append(retryErrors, err)

		if attempt == test.Config.RetryCount {
			// Final attempt failed
			endTime := time.Now()
			duration := endTime.Sub(startTime)

			result := &E2ETestResult{
				Test:        test,
				Status:      E2EStatusFailed,
				StartTime:   startTime,
				EndTime:     endTime,
				Duration:    duration,
				Checkpoints: ctx.Checkpoints,
				Assertions:  ctx.Assertions,
				UserJourney: ctx.UserJourney,
				Environment: ctx.Environment,
				Passed:      false,
				Error:       lastError,
				RetryCount:  attempt,
				RetryErrors: retryErrors,
				Metadata:    map[string]interface{}{},
			}

			return result
		}
	}

	// This should never be reached, but just in case
	return &E2ETestResult{
		Test:   test,
		Status: E2EStatusError,
		Error:  fmt.Errorf("unexpected retry loop exit"),
	}
}

// evaluateTestResult evaluates test results based on checkpoints and assertions
func (etr *E2ETestRunner) evaluateTestResult(result *E2ETestResult) bool {
	// Check if all required checkpoints passed
	for _, checkpoint := range result.Checkpoints {
		if checkpoint.Checkpoint.Required && !checkpoint.Passed {
			etr.Logger.Error("Required checkpoint failed",
				zap.String("test", result.Test.Name),
				zap.String("checkpoint", checkpoint.Checkpoint.Name),
				zap.Error(checkpoint.Error))
			return false
		}
	}

	// Check if all critical assertions passed
	for _, assertion := range result.Assertions {
		if assertion.Assertion.Critical && !assertion.Passed {
			etr.Logger.Error("Critical assertion failed",
				zap.String("test", result.Test.Name),
				zap.String("assertion", assertion.Assertion.Name),
				zap.Error(assertion.Error))
			return false
		}
	}

	return true
}

// GenerateSummary generates a summary of all test results
func (etr *E2ETestRunner) GenerateSummary() *E2ETestSummary {
	etr.mu.Lock()
	defer etr.mu.Unlock()

	summary := &E2ETestSummary{
		TotalTests:     len(etr.Results),
		PassedTests:    0,
		FailedTests:    0,
		SkippedTests:   0,
		ErrorTests:     0,
		TimeoutTests:   0,
		RetryingTests:  0,
		TotalDuration:  0,
		Results:        etr.Results,
		Environment:    etr.Environment,
		TestCategories: map[E2ETestCategory]int{},
		TestPriorities: map[E2ETestPriority]int{},
	}

	for _, result := range etr.Results {
		summary.TotalDuration += result.Duration

		// Count by status
		switch result.Status {
		case E2EStatusPassed:
			summary.PassedTests++
		case E2EStatusFailed:
			summary.FailedTests++
		case E2EStatusSkipped:
			summary.SkippedTests++
		case E2EStatusError:
			summary.ErrorTests++
		case E2EStatusTimeout:
			summary.TimeoutTests++
		case E2EStatusRetrying:
			summary.RetryingTests++
		}

		// Count by category
		summary.TestCategories[result.Test.Category]++

		// Count by priority
		summary.TestPriorities[result.Test.Priority]++
	}

	return summary
}

// E2ETestSummary represents a summary of E2E test results
type E2ETestSummary struct {
	TotalTests     int
	PassedTests    int
	FailedTests    int
	SkippedTests   int
	ErrorTests     int
	TimeoutTests   int
	RetryingTests  int
	TotalDuration  time.Duration
	Results        []*E2ETestResult
	Environment    *E2EEnvironment
	TestCategories map[E2ETestCategory]int
	TestPriorities map[E2ETestPriority]int
}

// generateE2ETestID generates a unique E2E test ID
func generateE2ETestID() string {
	return fmt.Sprintf("e2e_%d", time.Now().UnixNano())
}
