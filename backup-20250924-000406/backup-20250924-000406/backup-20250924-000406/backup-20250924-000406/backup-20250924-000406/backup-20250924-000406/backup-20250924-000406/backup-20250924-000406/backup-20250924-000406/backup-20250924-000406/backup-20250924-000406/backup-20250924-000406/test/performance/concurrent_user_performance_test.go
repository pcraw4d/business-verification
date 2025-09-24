package performance

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestConcurrentUserPerformance tests concurrent user scenarios for MVP (20 users)
func TestConcurrentUserPerformance(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.ConcurrentUsers = 20
	config.TestDuration = 2 * time.Minute
	config.ResponseTimeLimit = 5 * time.Second

	suite := NewPerformanceTestSuite(config)

	// Add concurrent user performance tests
	suite.AddTest(PerformanceTest{
		Name:        "ConcurrentUserPortfolioAccess",
		Description: "Test 20 concurrent users accessing merchant portfolio",
		TestFunc:    testConcurrentUserPortfolioAccess,
	})

	suite.AddTest(PerformanceTest{
		Name:        "ConcurrentUserMerchantDetailView",
		Description: "Test 20 concurrent users viewing merchant details",
		TestFunc:    testConcurrentUserMerchantDetailView,
	})

	suite.AddTest(PerformanceTest{
		Name:        "ConcurrentUserSearchOperations",
		Description: "Test 20 concurrent users performing search operations",
		TestFunc:    testConcurrentUserSearchOperations,
	})

	suite.AddTest(PerformanceTest{
		Name:        "ConcurrentUserBulkOperations",
		Description: "Test 20 concurrent users performing bulk operations",
		TestFunc:    testConcurrentUserBulkOperations,
	})

	suite.AddTest(PerformanceTest{
		Name:        "ConcurrentUserSessionManagement",
		Description: "Test 20 concurrent users with session management",
		TestFunc:    testConcurrentUserSessionManagement,
	})

	// Run all tests
	suite.RunAllTests(t)
}

// testConcurrentUserPortfolioAccess tests concurrent portfolio access
func testConcurrentUserPortfolioAccess(t *testing.T, config *PerformanceTestConfig) error {
	helper := NewBenchmarkHelper(config)
	runner := NewPerformanceTestRunner(config)
	runner.Start()

	// Generate test data
	testData := helper.GenerateTestData(1000)

	// Create concurrent user simulators
	var wg sync.WaitGroup
	userSimulators := make([]*ConcurrentUserSimulator, config.ConcurrentUsers)

	// Initialize user simulators
	for i := 0; i < config.ConcurrentUsers; i++ {
		userSimulators[i] = NewConcurrentUserSimulator(i, config, runner)
		userSimulators[i].Start()
	}

	// Simulate concurrent portfolio access
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			simulator := userSimulators[userID]
			ctx, cancel := context.WithTimeout(context.Background(), config.TestDuration)
			defer cancel()

			// Simulate user behavior patterns
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Simulate different portfolio operations
					operation := func() {
						start := time.Now()

						// Randomly choose operation type
						operationType := rand.Intn(4)
						switch operationType {
						case 0:
							// List merchants
							simulateListMerchants(testData)
						case 1:
							// Search merchants
							simulateSearchMerchants(testData)
						case 2:
							// Filter merchants
							simulateFilterMerchants(testData)
						case 3:
							// Paginate merchants
							simulatePaginateMerchants(testData)
						}

						responseTime := time.Since(start)
						runner.RecordRequest(responseTime, true)
					}

					simulator.AddOperation(operation)

					// Simulate user think time
					thinkTime := time.Duration(rand.Intn(2000)+500) * time.Millisecond
					time.Sleep(thinkTime)
				}
			}
		}(i)
	}

	// Wait for test duration
	time.Sleep(config.TestDuration)

	// Stop all user simulators
	for _, simulator := range userSimulators {
		simulator.Stop()
	}

	wg.Wait()
	runner.Stop()

	metrics := runner.GetMetrics()
	runner.PrintMetrics("Concurrent User Portfolio Access Performance")

	// Assert performance requirements for concurrent users
	assert.GreaterOrEqual(t, metrics.TotalRequests, int64(config.ConcurrentUsers*10),
		"Should handle at least 10 requests per user during test duration")
	assert.LessOrEqual(t, metrics.AverageResponseTime, 2*time.Second,
		"Average response time should be within 2 seconds under concurrent load")
	assert.Less(t, metrics.ErrorRate, 5.0,
		"Error rate should be less than 5%% under concurrent load")
	assert.GreaterOrEqual(t, metrics.RequestsPerSecond, float64(config.ConcurrentUsers),
		"Should handle at least 1 request per second per concurrent user")

	return nil
}

// testConcurrentUserMerchantDetailView tests concurrent merchant detail views
func testConcurrentUserMerchantDetailView(t *testing.T, config *PerformanceTestConfig) error {
	helper := NewBenchmarkHelper(config)
	runner := NewPerformanceTestRunner(config)
	runner.Start()

	// Generate test data
	testData := helper.GenerateTestData(500)

	// Create concurrent user simulators
	var wg sync.WaitGroup
	userSimulators := make([]*ConcurrentUserSimulator, config.ConcurrentUsers)

	// Initialize user simulators
	for i := 0; i < config.ConcurrentUsers; i++ {
		userSimulators[i] = NewConcurrentUserSimulator(i, config, runner)
		userSimulators[i].Start()
	}

	// Simulate concurrent merchant detail views
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			simulator := userSimulators[userID]
			ctx, cancel := context.WithTimeout(context.Background(), config.TestDuration)
			defer cancel()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Simulate merchant detail view
					operation := func() {
						start := time.Now()

						// Randomly select a merchant
						merchant := testData[rand.Intn(len(testData))]

						// Simulate loading merchant details
						simulateLoadMerchantDetails(merchant)

						responseTime := time.Since(start)
						runner.RecordRequest(responseTime, true)
					}

					simulator.AddOperation(operation)

					// Simulate user viewing time
					viewTime := time.Duration(rand.Intn(5000)+1000) * time.Millisecond
					time.Sleep(viewTime)
				}
			}
		}(i)
	}

	// Wait for test duration
	time.Sleep(config.TestDuration)

	// Stop all user simulators
	for _, simulator := range userSimulators {
		simulator.Stop()
	}

	wg.Wait()
	runner.Stop()

	metrics := runner.GetMetrics()
	runner.PrintMetrics("Concurrent User Merchant Detail View Performance")

	// Assert performance requirements
	assert.GreaterOrEqual(t, metrics.TotalRequests, int64(config.ConcurrentUsers*5),
		"Should handle at least 5 detail view requests per user")
	assert.LessOrEqual(t, metrics.AverageResponseTime, 1*time.Second,
		"Merchant detail view should load within 1 second")
	assert.Less(t, metrics.ErrorRate, 2.0,
		"Error rate should be less than 2%% for detail views")

	return nil
}

// testConcurrentUserSearchOperations tests concurrent search operations
func testConcurrentUserSearchOperations(t *testing.T, config *PerformanceTestConfig) error {
	helper := NewBenchmarkHelper(config)
	runner := NewPerformanceTestRunner(config)
	runner.Start()

	// Generate test data
	testData := helper.GenerateTestData(2000)

	// Create search queries
	searchQueries := []string{
		"Test Merchant",
		"Technology",
		"@test.com",
		"555-",
		"Finance",
		"High Risk",
		"Onboarded",
		"Deactivated",
	}

	// Create concurrent user simulators
	var wg sync.WaitGroup
	userSimulators := make([]*ConcurrentUserSimulator, config.ConcurrentUsers)

	// Initialize user simulators
	for i := 0; i < config.ConcurrentUsers; i++ {
		userSimulators[i] = NewConcurrentUserSimulator(i, config, runner)
		userSimulators[i].Start()
	}

	// Simulate concurrent search operations
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			simulator := userSimulators[userID]
			ctx, cancel := context.WithTimeout(context.Background(), config.TestDuration)
			defer cancel()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Simulate search operation
					operation := func() {
						start := time.Now()

						// Randomly select search query
						query := searchQueries[rand.Intn(len(searchQueries))]

						// Simulate search operation
						simulateSearchOperation(testData, query)

						responseTime := time.Since(start)
						runner.RecordRequest(responseTime, true)
					}

					simulator.AddOperation(operation)

					// Simulate user search behavior
					searchInterval := time.Duration(rand.Intn(3000)+1000) * time.Millisecond
					time.Sleep(searchInterval)
				}
			}
		}(i)
	}

	// Wait for test duration
	time.Sleep(config.TestDuration)

	// Stop all user simulators
	for _, simulator := range userSimulators {
		simulator.Stop()
	}

	wg.Wait()
	runner.Stop()

	metrics := runner.GetMetrics()
	runner.PrintMetrics("Concurrent User Search Operations Performance")

	// Assert performance requirements
	assert.GreaterOrEqual(t, metrics.TotalRequests, int64(config.ConcurrentUsers*8),
		"Should handle at least 8 search requests per user")
	assert.LessOrEqual(t, metrics.AverageResponseTime, 1*time.Second,
		"Search operations should complete within 1 second")
	assert.Less(t, metrics.ErrorRate, 3.0,
		"Error rate should be less than 3%% for search operations")
	assert.GreaterOrEqual(t, metrics.RequestsPerSecond, float64(config.ConcurrentUsers)*0.5,
		"Should handle at least 0.5 search requests per second per user")

	return nil
}

// testConcurrentUserBulkOperations tests concurrent bulk operations
func testConcurrentUserBulkOperations(t *testing.T, config *PerformanceTestConfig) error {
	helper := NewBenchmarkHelper(config)
	runner := NewPerformanceTestRunner(config)
	runner.Start()

	// Generate test data
	testData := helper.GenerateTestData(1000)

	// Create concurrent user simulators
	var wg sync.WaitGroup
	userSimulators := make([]*ConcurrentUserSimulator, config.ConcurrentUsers)

	// Initialize user simulators
	for i := 0; i < config.ConcurrentUsers; i++ {
		userSimulators[i] = NewConcurrentUserSimulator(i, config, runner)
		userSimulators[i].Start()
	}

	// Simulate concurrent bulk operations
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			simulator := userSimulators[userID]
			ctx, cancel := context.WithTimeout(context.Background(), config.TestDuration)
			defer cancel()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Simulate bulk operation
					operation := func() {
						start := time.Now()

						// Randomly choose bulk operation type
						operationType := rand.Intn(3)
						switch operationType {
						case 0:
							// Bulk update
							simulateBulkUpdate(testData, 50)
						case 1:
							// Bulk export
							simulateBulkExport(testData, 100)
						case 2:
							// Bulk status change
							simulateBulkStatusChange(testData, 75)
						}

						responseTime := time.Since(start)
						runner.RecordRequest(responseTime, true)
					}

					simulator.AddOperation(operation)

					// Simulate bulk operation interval
					operationInterval := time.Duration(rand.Intn(10000)+5000) * time.Millisecond
					time.Sleep(operationInterval)
				}
			}
		}(i)
	}

	// Wait for test duration
	time.Sleep(config.TestDuration)

	// Stop all user simulators
	for _, simulator := range userSimulators {
		simulator.Stop()
	}

	wg.Wait()
	runner.Stop()

	metrics := runner.GetMetrics()
	runner.PrintMetrics("Concurrent User Bulk Operations Performance")

	// Assert performance requirements
	assert.GreaterOrEqual(t, metrics.TotalRequests, int64(config.ConcurrentUsers*2),
		"Should handle at least 2 bulk operations per user")
	assert.LessOrEqual(t, metrics.AverageResponseTime, 10*time.Second,
		"Bulk operations should complete within 10 seconds")
	assert.Less(t, metrics.ErrorRate, 10.0,
		"Error rate should be less than 10%% for bulk operations")

	return nil
}

// testConcurrentUserSessionManagement tests concurrent session management
func testConcurrentUserSessionManagement(t *testing.T, config *PerformanceTestConfig) error {
	helper := NewBenchmarkHelper(config)
	runner := NewPerformanceTestRunner(config)
	runner.Start()

	// Generate test data
	testData := helper.GenerateTestData(500)

	// Create session manager simulation
	sessionManager := NewSessionManagerSimulator()

	// Create concurrent user simulators
	var wg sync.WaitGroup
	userSimulators := make([]*ConcurrentUserSimulator, config.ConcurrentUsers)

	// Initialize user simulators
	for i := 0; i < config.ConcurrentUsers; i++ {
		userSimulators[i] = NewConcurrentUserSimulator(i, config, runner)
		userSimulators[i].Start()
	}

	// Simulate concurrent session management
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			simulator := userSimulators[userID]
			ctx, cancel := context.WithTimeout(context.Background(), config.TestDuration)
			defer cancel()

			// Simulate user session lifecycle
			for {
				select {
				case <-ctx.Done():
					// Cleanup session
					sessionManager.EndSession(userID)
					return
				default:
					// Simulate session operations
					operation := func() {
						start := time.Now()

						// Randomly choose session operation
						operationType := rand.Intn(4)
						switch operationType {
						case 0:
							// Start session
							sessionManager.StartSession(userID)
						case 1:
							// Switch merchant
							merchant := testData[rand.Intn(len(testData))]
							sessionManager.SwitchMerchant(userID, merchant["id"].(string))
						case 2:
							// Get current session
							sessionManager.GetCurrentSession(userID)
						case 3:
							// End session
							sessionManager.EndSession(userID)
						}

						responseTime := time.Since(start)
						runner.RecordRequest(responseTime, true)
					}

					simulator.AddOperation(operation)

					// Simulate user session behavior
					sessionInterval := time.Duration(rand.Intn(5000)+2000) * time.Millisecond
					time.Sleep(sessionInterval)
				}
			}
		}(i)
	}

	// Wait for test duration
	time.Sleep(config.TestDuration)

	// Stop all user simulators
	for _, simulator := range userSimulators {
		simulator.Stop()
	}

	wg.Wait()
	runner.Stop()

	metrics := runner.GetMetrics()
	runner.PrintMetrics("Concurrent User Session Management Performance")

	// Assert performance requirements
	assert.GreaterOrEqual(t, metrics.TotalRequests, int64(config.ConcurrentUsers*5),
		"Should handle at least 5 session operations per user")
	assert.LessOrEqual(t, metrics.AverageResponseTime, 500*time.Millisecond,
		"Session operations should complete within 500ms")
	assert.Less(t, metrics.ErrorRate, 1.0,
		"Error rate should be less than 1%% for session operations")

	return nil
}

// SessionManagerSimulator simulates session management for testing
type SessionManagerSimulator struct {
	sessions map[int]string
	mutex    sync.RWMutex
}

// NewSessionManagerSimulator creates a new session manager simulator
func NewSessionManagerSimulator() *SessionManagerSimulator {
	return &SessionManagerSimulator{
		sessions: make(map[int]string),
	}
}

// StartSession starts a new session for a user
func (sms *SessionManagerSimulator) StartSession(userID int) {
	sms.mutex.Lock()
	defer sms.mutex.Unlock()

	sms.sessions[userID] = ""
	time.Sleep(10 * time.Millisecond) // Simulate session creation time
}

// SwitchMerchant switches the current merchant for a user session
func (sms *SessionManagerSimulator) SwitchMerchant(userID int, merchantID string) {
	sms.mutex.Lock()
	defer sms.mutex.Unlock()

	sms.sessions[userID] = merchantID
	time.Sleep(5 * time.Millisecond) // Simulate merchant switch time
}

// GetCurrentSession gets the current session for a user
func (sms *SessionManagerSimulator) GetCurrentSession(userID int) string {
	sms.mutex.RLock()
	defer sms.mutex.RUnlock()

	time.Sleep(2 * time.Millisecond) // Simulate session lookup time
	return sms.sessions[userID]
}

// EndSession ends a session for a user
func (sms *SessionManagerSimulator) EndSession(userID int) {
	sms.mutex.Lock()
	defer sms.mutex.Unlock()

	delete(sms.sessions, userID)
	time.Sleep(5 * time.Millisecond) // Simulate session cleanup time
}

// BenchmarkConcurrentUsers benchmarks concurrent user operations
func BenchmarkConcurrentUsers(b *testing.B) {
	config := DefaultPerformanceConfig()
	config.ConcurrentUsers = 20

	helper := NewBenchmarkHelper(config)
	testData := helper.GenerateTestData(1000)

	b.ResetTimer()

	b.Run("ConcurrentPortfolioAccess", func(b *testing.B) {
		var wg sync.WaitGroup
		for i := 0; i < config.ConcurrentUsers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < b.N/config.ConcurrentUsers; j++ {
					simulateListMerchants(testData)
				}
			}()
		}
		wg.Wait()
	})

	b.Run("ConcurrentSearchOperations", func(b *testing.B) {
		var wg sync.WaitGroup
		for i := 0; i < config.ConcurrentUsers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < b.N/config.ConcurrentUsers; j++ {
					simulateSearchOperation(testData, "Test")
				}
			}()
		}
		wg.Wait()
	})
}
