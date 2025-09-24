package testing

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"go.uber.org/zap"
)

// IntegrationTest represents an integration test that tests multiple components working together
type IntegrationTest struct {
	ID          string
	Name        string
	Description string
	Function    IntegrationTestFunction
	Setup       func(ctx *IntegrationContext) error
	Teardown    func(ctx *IntegrationContext) error
	Timeout     time.Duration
	Parallel    bool
	Tags        []string
	Status      TestStatus
	Error       error
	Duration    time.Duration
	Output      []string
	Assertions  []*Assertion
	Components  []string // List of components this test integrates
	Database    bool     // Whether this test requires database
	External    bool     // Whether this test requires external services
	Network     bool     // Whether this test requires network access
}

// IntegrationTestFunction represents an integration test function
type IntegrationTestFunction func(ctx *IntegrationContext) error

// IntegrationContext provides context and utilities for integration test execution
type IntegrationContext struct {
	Test         *IntegrationTest
	Suite        *IntegrationTestSuite
	Context      context.Context
	Logger       *zap.Logger
	Database     *sql.DB
	HTTPClient   *http.Client
	Server       *httptest.Server
	Fixtures     *IntegrationFixtures
	Mocks        *IntegrationMocks
	CleanupFuncs []func() error
	cancel       context.CancelFunc
}

// IntegrationTestSuite represents a collection of integration tests
type IntegrationTestSuite struct {
	Name         string
	Description  string
	Tests        []*IntegrationTest
	SetupFunc    func(ctx *IntegrationContext) error
	TeardownFunc func(ctx *IntegrationContext) error
	BeforeEach   func(ctx *IntegrationContext) error
	AfterEach    func(ctx *IntegrationContext) error
	Parallel     bool
	Timeout      time.Duration
	Tags         []string
	Database     bool
	External     bool
	Network      bool
	logger       *zap.Logger
	mutex        sync.RWMutex
}

// IntegrationFixtures provides fixtures for integration testing
type IntegrationFixtures struct {
	DatabaseFixtures map[string]interface{}
	FileFixtures     map[string]string
	ConfigFixtures   map[string]interface{}
	HTTPFixtures     map[string]*HTTPFixture
	mutex            sync.RWMutex
}

// HTTPFixture represents an HTTP response fixture
type HTTPFixture struct {
	Method     string
	Path       string
	StatusCode int
	Headers    map[string]string
	Body       string
	Delay      time.Duration
	Error      error
}

// IntegrationMocks provides mocking capabilities for integration tests
type IntegrationMocks struct {
	HTTPMocks     map[string]*HTTPMock
	DatabaseMocks map[string]*DatabaseMock
	ExternalMocks map[string]*ExternalMock
	mutex         sync.RWMutex
}

// HTTPMock represents an HTTP service mock
type HTTPMock struct {
	ID       string
	BaseURL  string
	Handlers map[string]http.HandlerFunc
	Server   *httptest.Server
}

// DatabaseMock represents a database mock
type DatabaseMock struct {
	ID      string
	Queries map[string]*QueryMock
	Results map[string][]interface{}
	Errors  map[string]error
}

// QueryMock represents a database query mock
type QueryMock struct {
	Query     string
	Args      []interface{}
	Result    []interface{}
	Error     error
	CallCount int
}

// ExternalMock represents an external service mock
type ExternalMock struct {
	ID      string
	Service string
	Methods map[string]*MethodMock
	CallLog []*ExternalCall
}

// MethodMock represents an external service method mock
type MethodMock struct {
	Method    string
	Args      []interface{}
	Return    []interface{}
	Error     error
	CallCount int
	Delay     time.Duration
}

// ExternalCall represents a call to an external service
type ExternalCall struct {
	ID        string
	Service   string
	Method    string
	Args      []interface{}
	Return    []interface{}
	Error     error
	Duration  time.Duration
	Timestamp time.Time
}

// IntegrationTestConfig represents configuration for integration testing
type IntegrationTestConfig struct {
	DatabaseURL      string
	DatabaseDriver   string
	HTTPTimeout      time.Duration
	MaxRetries       int
	RetryDelay       time.Duration
	Parallel         bool
	MaxGoroutines    int
	FailFast         bool
	CoverageEnabled  bool
	CoverageOutput   string
	LogLevel         string
	Tags             []string
	SkipTags         []string
	Components       []string
	ExternalServices []string
}

// IntegrationTestResult represents the result of an integration test
type IntegrationTestResult struct {
	Suite       *IntegrationTestSuite
	Test        *IntegrationTest
	Status      TestStatus
	Error       error
	Duration    time.Duration
	Assertions  []*Assertion
	Output      []string
	Coverage    *TestCoverage
	Performance *TestPerformance
	Components  []string
	Database    bool
	External    bool
	Network     bool
}

// IntegrationTestRunner manages integration test execution
type IntegrationTestRunner struct {
	suites  []*IntegrationTestSuite
	results []*IntegrationTestResult
	config  *IntegrationTestConfig
	logger  *zap.Logger
	mutex   sync.RWMutex
}

// NewIntegrationTest creates a new integration test
func NewIntegrationTest(name, description string, function IntegrationTestFunction) *IntegrationTest {
	return &IntegrationTest{
		ID:          generateTestID(),
		Name:        name,
		Description: description,
		Function:    function,
		Timeout:     30 * time.Second,
		Parallel:    false,
		Tags:        make([]string, 0),
		Status:      TestStatusPending,
		Output:      make([]string, 0),
		Assertions:  make([]*Assertion, 0),
		Components:  make([]string, 0),
		Database:    false,
		External:    false,
		Network:     false,
	}
}

// AddTag adds a tag to the integration test
func (it *IntegrationTest) AddTag(tag string) {
	it.Tags = append(it.Tags, tag)
}

// SetTimeout sets the timeout for the integration test
func (it *IntegrationTest) SetTimeout(timeout time.Duration) {
	it.Timeout = timeout
}

// SetParallel sets whether the integration test can run in parallel
func (it *IntegrationTest) SetParallel(parallel bool) {
	it.Parallel = parallel
}

// AddComponent adds a component that this test integrates
func (it *IntegrationTest) AddComponent(component string) {
	it.Components = append(it.Components, component)
}

// SetDatabase sets whether this test requires database access
func (it *IntegrationTest) SetDatabase(requires bool) {
	it.Database = requires
}

// SetExternal sets whether this test requires external services
func (it *IntegrationTest) SetExternal(requires bool) {
	it.External = requires
}

// SetNetwork sets whether this test requires network access
func (it *IntegrationTest) SetNetwork(requires bool) {
	it.Network = requires
}

// NewIntegrationTestSuite creates a new integration test suite
func NewIntegrationTestSuite(name, description string) *IntegrationTestSuite {
	return &IntegrationTestSuite{
		Name:        name,
		Description: description,
		Tests:       make([]*IntegrationTest, 0),
		Parallel:    false,
		Timeout:     60 * time.Second,
		Tags:        make([]string, 0),
		Database:    false,
		External:    false,
		Network:     false,
		logger:      zap.NewNop(),
	}
}

// AddTest adds a test to the integration test suite
func (it *IntegrationTestSuite) AddTest(test *IntegrationTest) {
	it.mutex.Lock()
	defer it.mutex.Unlock()
	it.Tests = append(it.Tests, test)
}

// CreateTest creates a new test within the suite
func (it *IntegrationTestSuite) CreateTest(name, description string, function IntegrationTestFunction) *IntegrationTest {
	test := NewIntegrationTest(name, description, function)
	it.AddTest(test)
	return test
}

// SetSetup sets the setup function for the integration test suite
func (it *IntegrationTestSuite) SetSetup(setup func(ctx *IntegrationContext) error) {
	it.SetupFunc = setup
}

// SetTeardown sets the teardown function for the integration test suite
func (it *IntegrationTestSuite) SetTeardown(teardown func(ctx *IntegrationContext) error) {
	it.TeardownFunc = teardown
}

// SetBeforeEach sets the before each function for the integration test suite
func (it *IntegrationTestSuite) SetBeforeEach(beforeEach func(ctx *IntegrationContext) error) {
	it.BeforeEach = beforeEach
}

// SetAfterEach sets the after each function for the integration test suite
func (it *IntegrationTestSuite) SetAfterEach(afterEach func(ctx *IntegrationContext) error) {
	it.AfterEach = afterEach
}

// SetParallel sets whether the integration test suite can run in parallel
func (it *IntegrationTestSuite) SetParallel(parallel bool) {
	it.Parallel = parallel
}

// SetTimeout sets the timeout for the integration test suite
func (it *IntegrationTestSuite) SetTimeout(timeout time.Duration) {
	it.Timeout = timeout
}

// AddTag adds a tag to the integration test suite
func (it *IntegrationTestSuite) AddTag(tag string) {
	it.Tags = append(it.Tags, tag)
}

// SetDatabase sets whether this suite requires database access
func (it *IntegrationTestSuite) SetDatabase(requires bool) {
	it.Database = requires
}

// SetExternal sets whether this suite requires external services
func (it *IntegrationTestSuite) SetExternal(requires bool) {
	it.External = requires
}

// SetNetwork sets whether this suite requires network access
func (it *IntegrationTestSuite) SetNetwork(requires bool) {
	it.Network = requires
}

// NewIntegrationContext creates a new integration context
func NewIntegrationContext(test *IntegrationTest, suite *IntegrationTestSuite) *IntegrationContext {
	ctx, cancel := context.WithTimeout(context.Background(), test.Timeout)

	return &IntegrationContext{
		Test:         test,
		Suite:        suite,
		Context:      ctx,
		Logger:       suite.logger,
		Fixtures:     NewIntegrationFixtures(),
		Mocks:        NewIntegrationMocks(),
		CleanupFuncs: make([]func() error, 0),
		cancel:       cancel,
	}
}

// Cleanup performs cleanup operations for the integration context
func (ic *IntegrationContext) Cleanup() {
	if ic.cancel != nil {
		ic.cancel()
	}

	// Run cleanup functions in reverse order
	for i := len(ic.CleanupFuncs) - 1; i >= 0; i-- {
		if err := ic.CleanupFuncs[i](); err != nil {
			if ic.Logger != nil {
				ic.Logger.Error("Cleanup function failed", zap.Error(err))
			}
		}
	}
}

// AddCleanup adds a cleanup function to be executed when the context is cleaned up
func (ic *IntegrationContext) AddCleanup(cleanup func() error) {
	ic.CleanupFuncs = append(ic.CleanupFuncs, cleanup)
}

// Log logs a message to the test output
func (ic *IntegrationContext) Log(message string) {
	ic.Test.Output = append(ic.Test.Output, message)
	ic.Logger.Info(message, zap.String("test", ic.Test.Name))
}

// Logf logs a formatted message to the test output
func (ic *IntegrationContext) Logf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	ic.Log(message)
}

// Assert returns an assertion builder for the integration context
func (ic *IntegrationContext) Assert() *AssertionBuilder {
	return &AssertionBuilder{context: &TestContext{Test: &UnitTest{}}}
}

// NewIntegrationFixtures creates new integration fixtures
func NewIntegrationFixtures() *IntegrationFixtures {
	return &IntegrationFixtures{
		DatabaseFixtures: make(map[string]interface{}),
		FileFixtures:     make(map[string]string),
		ConfigFixtures:   make(map[string]interface{}),
		HTTPFixtures:     make(map[string]*HTTPFixture),
	}
}

// NewIntegrationMocks creates new integration mocks
func NewIntegrationMocks() *IntegrationMocks {
	return &IntegrationMocks{
		HTTPMocks:     make(map[string]*HTTPMock),
		DatabaseMocks: make(map[string]*DatabaseMock),
		ExternalMocks: make(map[string]*ExternalMock),
	}
}

// DefaultIntegrationTestConfig returns the default integration test configuration
func DefaultIntegrationTestConfig() *IntegrationTestConfig {
	return &IntegrationTestConfig{
		DatabaseURL:      "",
		DatabaseDriver:   "postgres",
		HTTPTimeout:      30 * time.Second,
		MaxRetries:       3,
		RetryDelay:       1 * time.Second,
		Parallel:         false,
		MaxGoroutines:    10,
		FailFast:         false,
		CoverageEnabled:  true,
		CoverageOutput:   "integration_coverage.out",
		LogLevel:         "info",
		Tags:             make([]string, 0),
		SkipTags:         make([]string, 0),
		Components:       make([]string, 0),
		ExternalServices: make([]string, 0),
	}
}

// NewIntegrationTestRunner creates a new integration test runner
func NewIntegrationTestRunner(config *IntegrationTestConfig, logger *zap.Logger) *IntegrationTestRunner {
	if config == nil {
		config = DefaultIntegrationTestConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &IntegrationTestRunner{
		suites:  make([]*IntegrationTestSuite, 0),
		results: make([]*IntegrationTestResult, 0),
		config:  config,
		logger:  logger,
	}
}

// AddSuite adds a test suite to the integration test runner
func (itr *IntegrationTestRunner) AddSuite(suite *IntegrationTestSuite) {
	itr.mutex.Lock()
	defer itr.mutex.Unlock()
	itr.suites = append(itr.suites, suite)
}

// RunAllSuites runs all integration test suites
func (itr *IntegrationTestRunner) RunAllSuites() []*IntegrationTestResult {
	itr.mutex.Lock()
	defer itr.mutex.Unlock()

	itr.results = make([]*IntegrationTestResult, 0)

	for _, suite := range itr.suites {
		results := itr.runSuite(suite)
		itr.results = append(itr.results, results...)
	}

	return itr.results
}

// runSuite runs a single integration test suite
func (itr *IntegrationTestRunner) runSuite(suite *IntegrationTestSuite) []*IntegrationTestResult {
	results := make([]*IntegrationTestResult, 0)

	// Create integration context for suite
	ctx := &IntegrationContext{
		Suite:        suite,
		Logger:       itr.logger,
		Fixtures:     NewIntegrationFixtures(),
		Mocks:        NewIntegrationMocks(),
		CleanupFuncs: make([]func() error, 0),
	}
	defer ctx.Cleanup()

	// Run suite setup
	if suite.SetupFunc != nil {
		if err := suite.SetupFunc(ctx); err != nil {
			itr.logger.Error("Suite setup failed", zap.String("suite", suite.Name), zap.Error(err))
			return results
		}
	}

	// Run tests
	if itr.config.Parallel && suite.Parallel {
		results = itr.runTestsParallel(suite, ctx)
	} else {
		results = itr.runTestsSequential(suite, ctx)
	}

	// Run suite teardown
	if suite.TeardownFunc != nil {
		if err := suite.TeardownFunc(ctx); err != nil {
			itr.logger.Error("Suite teardown failed", zap.String("suite", suite.Name), zap.Error(err))
		}
	}

	return results
}

// runTestsParallel runs integration tests in parallel
func (itr *IntegrationTestRunner) runTestsParallel(suite *IntegrationTestSuite, ctx *IntegrationContext) []*IntegrationTestResult {
	results := make([]*IntegrationTestResult, len(suite.Tests))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, itr.config.MaxGoroutines)

	for i, test := range suite.Tests {
		if itr.shouldSkipTest(test) {
			results[i] = &IntegrationTestResult{
				Suite:  suite,
				Test:   test,
				Status: TestStatusSkipped,
			}
			continue
		}

		wg.Add(1)
		go func(idx int, test *IntegrationTest) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			results[idx] = itr.runSingleTest(suite, test, ctx)
		}(i, test)
	}

	wg.Wait()
	return results
}

// runTestsSequential runs integration tests sequentially
func (itr *IntegrationTestRunner) runTestsSequential(suite *IntegrationTestSuite, ctx *IntegrationContext) []*IntegrationTestResult {
	results := make([]*IntegrationTestResult, 0)

	for _, test := range suite.Tests {
		if itr.shouldSkipTest(test) {
			results = append(results, &IntegrationTestResult{
				Suite:  suite,
				Test:   test,
				Status: TestStatusSkipped,
			})
			continue
		}

		result := itr.runSingleTest(suite, test, ctx)
		results = append(results, result)

		if itr.config.FailFast && result.Status == TestStatusFailed {
			break
		}
	}

	return results
}

// runSingleTest runs a single integration test
func (itr *IntegrationTestRunner) runSingleTest(suite *IntegrationTestSuite, test *IntegrationTest, suiteCtx *IntegrationContext) *IntegrationTestResult {
	result := &IntegrationTestResult{
		Suite:       suite,
		Test:        test,
		Status:      TestStatusRunning,
		Assertions:  make([]*Assertion, 0),
		Output:      make([]string, 0),
		Coverage:    &TestCoverage{},
		Performance: &TestPerformance{},
		Components:  test.Components,
		Database:    test.Database,
		External:    test.External,
		Network:     test.Network,
	}

	start := time.Now()
	test.Status = TestStatusRunning

	// Create test context
	testCtx := NewIntegrationContext(test, suite)
	defer testCtx.Cleanup()

	// Run before each
	if suite.BeforeEach != nil {
		if err := suite.BeforeEach(testCtx); err != nil {
			result.Status = TestStatusFailed
			result.Error = fmt.Errorf("before each failed: %w", err)
			return result
		}
	}

	// Run test setup
	if test.Setup != nil {
		if err := test.Setup(testCtx); err != nil {
			result.Status = TestStatusFailed
			result.Error = fmt.Errorf("test setup failed: %w", err)
			return result
		}
	}

	// Run the test function
	err := itr.executeTestFunction(testCtx, test)

	// Run test teardown
	if test.Teardown != nil {
		if teardownErr := test.Teardown(testCtx); teardownErr != nil {
			itr.logger.Warn("Test teardown failed", zap.String("test", test.Name), zap.Error(teardownErr))
		}
	}

	// Run after each
	if suite.AfterEach != nil {
		if afterErr := suite.AfterEach(testCtx); afterErr != nil {
			itr.logger.Warn("After each failed", zap.String("test", test.Name), zap.Error(afterErr))
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

// executeTestFunction executes the integration test function
func (itr *IntegrationTestRunner) executeTestFunction(testCtx *IntegrationContext, test *IntegrationTest) error {
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
func (itr *IntegrationTestRunner) shouldSkipTest(test *IntegrationTest) bool {
	// Check tag filters
	if len(itr.config.Tags) > 0 {
		found := false
		for _, tag := range itr.config.Tags {
			for _, testTag := range test.Tags {
				if tag == testTag {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return true
		}
	}

	// Check skip tags
	for _, skipTag := range itr.config.SkipTags {
		for _, testTag := range test.Tags {
			if skipTag == testTag {
				return true
			}
		}
	}

	// Check component filters
	if len(itr.config.Components) > 0 {
		found := false
		for _, component := range itr.config.Components {
			for _, testComponent := range test.Components {
				if component == testComponent {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return true
		}
	}

	return false
}

// GenerateSummary generates a summary of integration test execution
func (itr *IntegrationTestRunner) GenerateSummary() *TestSummary {
	itr.mutex.RLock()
	defer itr.mutex.RUnlock()

	summary := &TestSummary{
		TotalTests:    len(itr.results),
		PassedTests:   0,
		FailedTests:   0,
		SkippedTests:  0,
		TimeoutTests:  0,
		TotalDuration: 0,
		Suites:        make(map[string]*SuiteSummary),
	}

	for _, result := range itr.results {
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
		if summary.Suites[suiteName] == nil {
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

// SetDatabaseFixture sets a database fixture
func (fixtures *IntegrationFixtures) SetDatabaseFixture(key string, value interface{}) {
	fixtures.mutex.Lock()
	defer fixtures.mutex.Unlock()
	fixtures.DatabaseFixtures[key] = value
}

// GetDatabaseFixture gets a database fixture
func (fixtures *IntegrationFixtures) GetDatabaseFixture(key string) (interface{}, bool) {
	fixtures.mutex.RLock()
	defer fixtures.mutex.RUnlock()
	value, exists := fixtures.DatabaseFixtures[key]
	return value, exists
}

// SetFileFixture sets a file fixture
func (fixtures *IntegrationFixtures) SetFileFixture(key, content string) {
	fixtures.mutex.Lock()
	defer fixtures.mutex.Unlock()
	fixtures.FileFixtures[key] = content
}

// GetFileFixture gets a file fixture
func (fixtures *IntegrationFixtures) GetFileFixture(key string) (string, bool) {
	fixtures.mutex.RLock()
	defer fixtures.mutex.RUnlock()
	content, exists := fixtures.FileFixtures[key]
	return content, exists
}

// SetConfigFixture sets a configuration fixture
func (fixtures *IntegrationFixtures) SetConfigFixture(key string, value interface{}) {
	fixtures.mutex.Lock()
	defer fixtures.mutex.Unlock()
	fixtures.ConfigFixtures[key] = value
}

// GetConfigFixture gets a configuration fixture
func (fixtures *IntegrationFixtures) GetConfigFixture(key string) (interface{}, bool) {
	fixtures.mutex.RLock()
	defer fixtures.mutex.RUnlock()
	value, exists := fixtures.ConfigFixtures[key]
	return value, exists
}

// SetHTTPFixture sets an HTTP fixture
func (fixtures *IntegrationFixtures) SetHTTPFixture(key string, fixture *HTTPFixture) {
	fixtures.mutex.Lock()
	defer fixtures.mutex.Unlock()
	fixtures.HTTPFixtures[key] = fixture
}

// GetHTTPFixture gets an HTTP fixture
func (fixtures *IntegrationFixtures) GetHTTPFixture(key string) (*HTTPFixture, bool) {
	fixtures.mutex.RLock()
	defer fixtures.mutex.RUnlock()
	fixture, exists := fixtures.HTTPFixtures[key]
	return fixture, exists
}

// AddHTTPMock adds an HTTP mock
func (im *IntegrationMocks) AddHTTPMock(id, baseURL string) *HTTPMock {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	mock := &HTTPMock{
		ID:       id,
		BaseURL:  baseURL,
		Handlers: make(map[string]http.HandlerFunc),
	}

	im.HTTPMocks[id] = mock
	return mock
}

// GetHTTPMock gets an HTTP mock
func (im *IntegrationMocks) GetHTTPMock(id string) (*HTTPMock, bool) {
	im.mutex.RLock()
	defer im.mutex.RUnlock()
	mock, exists := im.HTTPMocks[id]
	return mock, exists
}

// AddDatabaseMock adds a database mock
func (im *IntegrationMocks) AddDatabaseMock(id string) *DatabaseMock {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	mock := &DatabaseMock{
		ID:      id,
		Queries: make(map[string]*QueryMock),
		Results: make(map[string][]interface{}),
		Errors:  make(map[string]error),
	}

	im.DatabaseMocks[id] = mock
	return mock
}

// GetDatabaseMock gets a database mock
func (im *IntegrationMocks) GetDatabaseMock(id string) (*DatabaseMock, bool) {
	im.mutex.RLock()
	defer im.mutex.RUnlock()
	mock, exists := im.DatabaseMocks[id]
	return mock, exists
}

// AddExternalMock adds an external service mock
func (im *IntegrationMocks) AddExternalMock(id, service string) *ExternalMock {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	mock := &ExternalMock{
		ID:      id,
		Service: service,
		Methods: make(map[string]*MethodMock),
		CallLog: make([]*ExternalCall, 0),
	}

	im.ExternalMocks[id] = mock
	return mock
}

// GetExternalMock gets an external service mock
func (im *IntegrationMocks) GetExternalMock(id string) (*ExternalMock, bool) {
	im.mutex.RLock()
	defer im.mutex.RUnlock()
	mock, exists := im.ExternalMocks[id]
	return mock, exists
}

// AddHandler adds a handler to an HTTP mock
func (hm *HTTPMock) AddHandler(path string, handler http.HandlerFunc) {
	hm.Handlers[path] = handler
}

// Start starts the HTTP mock server
func (hm *HTTPMock) Start() {
	mux := http.NewServeMux()
	for path, handler := range hm.Handlers {
		mux.HandleFunc(path, handler)
	}
	hm.Server = httptest.NewServer(mux)
}

// Stop stops the HTTP mock server
func (hm *HTTPMock) Stop() {
	if hm.Server != nil {
		hm.Server.Close()
	}
}

// AddQuery adds a query mock to a database mock
func (dm *DatabaseMock) AddQuery(query string, result []interface{}, err error) {
	dm.Queries[query] = &QueryMock{
		Query:  query,
		Result: result,
		Error:  err,
	}
}

// GetQuery gets a query mock
func (dm *DatabaseMock) GetQuery(query string) (*QueryMock, bool) {
	qm, exists := dm.Queries[query]
	return qm, exists
}

// AddMethod adds a method mock to an external service mock
func (em *ExternalMock) AddMethod(method string, args []interface{}, returnValues []interface{}, err error) {
	em.Methods[method] = &MethodMock{
		Method: method,
		Args:   args,
		Return: returnValues,
		Error:  err,
	}
}

// GetMethod gets a method mock
func (em *ExternalMock) GetMethod(method string) (*MethodMock, bool) {
	mm, exists := em.Methods[method]
	return mm, exists
}

// RecordCall records a call to an external service
func (em *ExternalMock) RecordCall(method string, args []interface{}, returnValues []interface{}, err error, duration time.Duration) {
	call := &ExternalCall{
		ID:        generateCallID(),
		Service:   em.Service,
		Method:    method,
		Args:      args,
		Return:    returnValues,
		Error:     err,
		Duration:  duration,
		Timestamp: time.Now(),
	}

	em.CallLog = append(em.CallLog, call)
}
