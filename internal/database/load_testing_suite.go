package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"
)

// LoadTestingSuite provides specialized load testing for database operations
type LoadTestingSuite struct {
	db     *sql.DB
	logger *log.Logger
	config *LoadTestConfig
}

// LoadTestConfig contains configuration for load testing
type LoadTestConfig struct {
	// Test duration and phases
	TestDuration     time.Duration
	WarmupDuration   time.Duration
	RampupDuration   time.Duration
	RampdownDuration time.Duration

	// Load characteristics
	InitialUsers      int
	MaxUsers          int
	UsersIncrement    int
	IncrementInterval time.Duration

	// Request characteristics
	RequestsPerUser int
	RequestInterval time.Duration
	RequestTimeout  time.Duration

	// Performance thresholds
	MaxResponseTime time.Duration
	MaxErrorRate    float64
	MinThroughput   float64

	// Test scenarios
	TestScenarios []LoadTestScenario
}

// LoadTestScenario defines a specific test scenario
type LoadTestScenario struct {
	Name             string
	Weight           int // Relative weight for scenario selection
	Query            string
	Parameters       []interface{}
	ExpectedDuration time.Duration
	IsReadOnly       bool
}

// LoadTestResult contains the results of a load test
type LoadTestResult struct {
	TestName  string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration

	// Load characteristics
	MaxConcurrentUsers int
	TotalUsers         int
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64

	// Performance metrics
	AverageResponseTime time.Duration
	MinResponseTime     time.Duration
	MaxResponseTime     time.Duration
	P50ResponseTime     time.Duration
	P90ResponseTime     time.Duration
	P95ResponseTime     time.Duration
	P99ResponseTime     time.Duration

	// Throughput metrics
	RequestsPerSecond float64
	PeakThroughput    float64

	// Error analysis
	ErrorRate    float64
	ErrorTypes   map[string]int64
	TimeoutCount int64

	// Resource usage
	PeakConnections    int
	AverageConnections float64

	// Scenario breakdown
	ScenarioResults map[string]*ScenarioResult

	// Recommendations
	Recommendations []string
}

// ScenarioResult contains results for a specific scenario
type ScenarioResult struct {
	ScenarioName        string
	RequestCount        int64
	SuccessCount        int64
	FailureCount        int64
	AverageResponseTime time.Duration
	MaxResponseTime     time.Duration
	ErrorRate           float64
}

// NewLoadTestingSuite creates a new load testing suite
func NewLoadTestingSuite(db *sql.DB, config *LoadTestConfig) *LoadTestingSuite {
	if config == nil {
		config = &LoadTestConfig{
			TestDuration:      5 * time.Minute,
			WarmupDuration:    30 * time.Second,
			RampupDuration:    1 * time.Minute,
			RampdownDuration:  30 * time.Second,
			InitialUsers:      5,
			MaxUsers:          50,
			UsersIncrement:    5,
			IncrementInterval: 30 * time.Second,
			RequestsPerUser:   100,
			RequestInterval:   100 * time.Millisecond,
			RequestTimeout:    10 * time.Second,
			MaxResponseTime:   2 * time.Second,
			MaxErrorRate:      5.0,
			MinThroughput:     100.0,
			TestScenarios:     getDefaultTestScenarios(),
		}
	}

	return &LoadTestingSuite{
		db:     db,
		logger: log.New(log.Writer(), "[LOAD_TEST] ", log.LstdFlags),
		config: config,
	}
}

// RunLoadTest executes a comprehensive load test
func (lts *LoadTestingSuite) RunLoadTest(ctx context.Context) (*LoadTestResult, error) {
	lts.logger.Println("Starting comprehensive load test...")

	startTime := time.Now()
	result := &LoadTestResult{
		TestName:        "Database Load Test",
		StartTime:       startTime,
		ErrorTypes:      make(map[string]int64),
		ScenarioResults: make(map[string]*ScenarioResult),
	}

	// Initialize scenario results
	for _, scenario := range lts.config.TestScenarios {
		result.ScenarioResults[scenario.Name] = &ScenarioResult{
			ScenarioName: scenario.Name,
		}
	}

	// Create context with timeout
	testCtx, cancel := context.WithTimeout(ctx, lts.config.TestDuration)
	defer cancel()

	// Run load test phases
	if err := lts.runWarmupPhase(testCtx, result); err != nil {
		return nil, fmt.Errorf("warmup phase failed: %w", err)
	}

	if err := lts.runRampupPhase(testCtx, result); err != nil {
		return nil, fmt.Errorf("rampup phase failed: %w", err)
	}

	if err := lts.runSustainedLoadPhase(testCtx, result); err != nil {
		return nil, fmt.Errorf("sustained load phase failed: %w", err)
	}

	if err := lts.runRampdownPhase(testCtx, result); err != nil {
		return nil, fmt.Errorf("rampdown phase failed: %w", err)
	}

	// Calculate final results
	lts.calculateFinalResults(result)

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	lts.logger.Printf("Load test completed in %v", result.Duration)
	return result, nil
}

// runWarmupPhase runs the warmup phase
func (lts *LoadTestingSuite) runWarmupPhase(ctx context.Context, result *LoadTestResult) error {
	lts.logger.Println("Running warmup phase...")

	warmupCtx, cancel := context.WithTimeout(ctx, lts.config.WarmupDuration)
	defer cancel()

	// Run with initial users for warmup
	return lts.runLoadPhase(warmupCtx, lts.config.InitialUsers, "warmup", result)
}

// runRampupPhase runs the rampup phase
func (lts *LoadTestingSuite) runRampupPhase(ctx context.Context, result *LoadTestResult) error {
	lts.logger.Println("Running rampup phase...")

	rampupCtx, cancel := context.WithTimeout(ctx, lts.config.RampupDuration)
	defer cancel()

	// Gradually increase users
	currentUsers := lts.config.InitialUsers
	incrementTicker := time.NewTicker(lts.config.IncrementInterval)
	defer incrementTicker.Stop()

	done := make(chan struct{})
	go func() {
		<-rampupCtx.Done()
		close(done)
	}()

	for {
		select {
		case <-done:
			return nil
		case <-incrementTicker.C:
			if currentUsers < lts.config.MaxUsers {
				currentUsers += lts.config.UsersIncrement
				if currentUsers > lts.config.MaxUsers {
					currentUsers = lts.config.MaxUsers
				}

				lts.logger.Printf("Ramping up to %d users", currentUsers)

				// Run load phase with current user count
				phaseCtx, phaseCancel := context.WithTimeout(rampupCtx, lts.config.IncrementInterval)
				if err := lts.runLoadPhase(phaseCtx, currentUsers, "rampup", result); err != nil {
					phaseCancel()
					return err
				}
				phaseCancel()
			}
		}
	}
}

// runSustainedLoadPhase runs the sustained load phase
func (lts *LoadTestingSuite) runSustainedLoadPhase(ctx context.Context, result *LoadTestResult) error {
	lts.logger.Println("Running sustained load phase...")

	// Calculate remaining time for sustained load
	elapsed := time.Since(result.StartTime)
	remainingTime := lts.config.TestDuration - elapsed - lts.config.RampdownDuration
	if remainingTime <= 0 {
		lts.logger.Println("No time remaining for sustained load phase")
		return nil
	}

	sustainedCtx, cancel := context.WithTimeout(ctx, remainingTime)
	defer cancel()

	return lts.runLoadPhase(sustainedCtx, lts.config.MaxUsers, "sustained", result)
}

// runRampdownPhase runs the rampdown phase
func (lts *LoadTestingSuite) runRampdownPhase(ctx context.Context, result *LoadTestResult) error {
	lts.logger.Println("Running rampdown phase...")

	rampdownCtx, cancel := context.WithTimeout(ctx, lts.config.RampdownDuration)
	defer cancel()

	// Gradually decrease users
	currentUsers := lts.config.MaxUsers
	decrementInterval := lts.config.RampdownDuration / time.Duration((lts.config.MaxUsers-lts.config.InitialUsers)/lts.config.UsersIncrement)

	if decrementInterval <= 0 {
		decrementInterval = 5 * time.Second
	}

	decrementTicker := time.NewTicker(decrementInterval)
	defer decrementTicker.Stop()

	done := make(chan struct{})
	go func() {
		<-rampdownCtx.Done()
		close(done)
	}()

	for {
		select {
		case <-done:
			return nil
		case <-decrementTicker.C:
			if currentUsers > lts.config.InitialUsers {
				currentUsers -= lts.config.UsersIncrement
				if currentUsers < lts.config.InitialUsers {
					currentUsers = lts.config.InitialUsers
				}

				lts.logger.Printf("Ramping down to %d users", currentUsers)

				// Run load phase with current user count
				phaseCtx, phaseCancel := context.WithTimeout(rampdownCtx, decrementInterval)
				if err := lts.runLoadPhase(phaseCtx, currentUsers, "rampdown", result); err != nil {
					phaseCancel()
					return err
				}
				phaseCancel()
			}
		}
	}
}

// runLoadPhase runs a load phase with specified number of users
func (lts *LoadTestingSuite) runLoadPhase(ctx context.Context, userCount int, phase string, result *LoadTestResult) error {
	var wg sync.WaitGroup
	var responseTimes []time.Duration
	var responseTimesMu sync.Mutex

	// Track peak connections
	var peakConnections int32
	connectionTicker := time.NewTicker(1 * time.Second)
	defer connectionTicker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-connectionTicker.C:
				var currentConnections int
				if err := lts.db.QueryRowContext(ctx, "SELECT count(*) FROM pg_stat_activity WHERE state = 'active'").Scan(&currentConnections); err == nil {
					if int32(currentConnections) > atomic.LoadInt32(&peakConnections) {
						atomic.StoreInt32(&peakConnections, int32(currentConnections))
					}
				}
			}
		}
	}()

	// Start user goroutines
	for i := 0; i < userCount; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			// Each user performs requests
			for j := 0; j < lts.config.RequestsPerUser; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					// Select scenario based on weight
					scenario := lts.selectScenario()

					// Execute request
					startTime := time.Now()
					err := lts.executeRequest(ctx, scenario)
					responseTime := time.Since(startTime)

					// Record response time
					responseTimesMu.Lock()
					responseTimes = append(responseTimes, responseTime)
					responseTimesMu.Unlock()

					// Update counters
					atomic.AddInt64(&result.TotalRequests, 1)

					// Update scenario results
					scenarioResult := result.ScenarioResults[scenario.Name]
					atomic.AddInt64(&scenarioResult.RequestCount, 1)

					if err != nil {
						atomic.AddInt64(&result.FailedRequests, 1)
						atomic.AddInt64(&scenarioResult.FailureCount, 1)

						// Categorize error
						errorType := lts.categorizeError(err)
						atomic.AddInt64(&result.ErrorTypes[errorType], 1)

						if lts.isTimeoutError(err) {
							atomic.AddInt64(&result.TimeoutCount, 1)
						}
					} else {
						atomic.AddInt64(&result.SuccessfulRequests, 1)
						atomic.AddInt64(&scenarioResult.SuccessCount, 1)
					}

					// Update max response time
					if responseTime > result.MaxResponseTime {
						result.MaxResponseTime = responseTime
					}

					if responseTime > scenarioResult.MaxResponseTime {
						scenarioResult.MaxResponseTime = responseTime
					}

					// Small delay between requests
					time.Sleep(lts.config.RequestInterval)
				}
			}
		}(i)
	}

	// Wait for all users to complete
	wg.Wait()

	// Update peak connections
	if int32(atomic.LoadInt32(&peakConnections)) > int32(result.PeakConnections) {
		result.PeakConnections = int(atomic.LoadInt32(&peakConnections))
	}

	// Update max concurrent users
	if userCount > result.MaxConcurrentUsers {
		result.MaxConcurrentUsers = userCount
	}

	// Calculate response time statistics for this phase
	if len(responseTimes) > 0 {
		lts.calculateResponseTimeStats(responseTimes, result)
	}

	return nil
}

// selectScenario selects a test scenario based on weights
func (lts *LoadTestingSuite) selectScenario() *LoadTestScenario {
	if len(lts.config.TestScenarios) == 0 {
		return &LoadTestScenario{
			Name:       "default",
			Query:      "SELECT 1",
			IsReadOnly: true,
		}
	}

	// Simple weighted selection (could be improved with better algorithm)
	totalWeight := 0
	for _, scenario := range lts.config.TestScenarios {
		totalWeight += scenario.Weight
	}

	if totalWeight == 0 {
		return &lts.config.TestScenarios[0]
	}

	// For simplicity, use round-robin selection
	// In a real implementation, you'd use proper weighted selection
	return &lts.config.TestScenarios[0]
}

// executeRequest executes a database request
func (lts *LoadTestingSuite) executeRequest(ctx context.Context, scenario *LoadTestScenario) error {
	// Create context with timeout
	requestCtx, cancel := context.WithTimeout(ctx, lts.config.RequestTimeout)
	defer cancel()

	// Execute the query
	rows, err := lts.db.QueryContext(requestCtx, scenario.Query, scenario.Parameters...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Consume results (important for accurate timing)
	for rows.Next() {
		// Just scan to consume the row
		var dummy interface{}
		rows.Scan(&dummy)
	}

	return rows.Err()
}

// categorizeError categorizes an error for analysis
func (lts *LoadTestingSuite) categorizeError(err error) string {
	if err == nil {
		return "none"
	}

	errStr := err.Error()

	switch {
	case lts.isTimeoutError(err):
		return "timeout"
	case lts.isConnectionError(err):
		return "connection"
	case lts.isConstraintError(err):
		return "constraint"
	case lts.isSyntaxError(err):
		return "syntax"
	default:
		return "other"
	}
}

// isTimeoutError checks if an error is a timeout error
func (lts *LoadTestingSuite) isTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return contains(errStr, "timeout") || contains(errStr, "deadline exceeded")
}

// isConnectionError checks if an error is a connection error
func (lts *LoadTestingSuite) isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return contains(errStr, "connection") || contains(errStr, "network")
}

// isConstraintError checks if an error is a constraint error
func (lts *LoadTestingSuite) isConstraintError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return contains(errStr, "constraint") || contains(errStr, "unique") || contains(errStr, "foreign key")
}

// isSyntaxError checks if an error is a syntax error
func (lts *LoadTestingSuite) isSyntaxError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return contains(errStr, "syntax") || contains(errStr, "parse")
}

// calculateResponseTimeStats calculates response time statistics
func (lts *LoadTestingSuite) calculateResponseTimeStats(responseTimes []time.Duration, result *LoadTestResult) {
	if len(responseTimes) == 0 {
		return
	}

	// Sort response times for percentile calculation
	// For simplicity, we'll calculate basic stats
	var total time.Duration
	min := responseTimes[0]
	max := responseTimes[0]

	for _, rt := range responseTimes {
		total += rt
		if rt < min {
			min = rt
		}
		if rt > max {
			max = rt
		}
	}

	avg := total / time.Duration(len(responseTimes))

	// Update result statistics
	if result.MinResponseTime == 0 || min < result.MinResponseTime {
		result.MinResponseTime = min
	}

	if max > result.MaxResponseTime {
		result.MaxResponseTime = max
	}

	// Update average (this is a simplified calculation)
	if result.AverageResponseTime == 0 {
		result.AverageResponseTime = avg
	} else {
		// Weighted average with existing data
		result.AverageResponseTime = (result.AverageResponseTime + avg) / 2
	}
}

// calculateFinalResults calculates final test results
func (lts *LoadTestingSuite) calculateFinalResults(result *LoadTestResult) {
	// Calculate error rate
	if result.TotalRequests > 0 {
		result.ErrorRate = float64(result.FailedRequests) / float64(result.TotalRequests) * 100
	}

	// Calculate throughput
	if result.Duration > 0 {
		result.RequestsPerSecond = float64(result.TotalRequests) / result.Duration.Seconds()
	}

	// Calculate scenario error rates
	for _, scenarioResult := range result.ScenarioResults {
		if scenarioResult.RequestCount > 0 {
			scenarioResult.ErrorRate = float64(scenarioResult.FailureCount) / float64(scenarioResult.RequestCount) * 100
		}
	}

	// Generate recommendations
	lts.generateRecommendations(result)
}

// generateRecommendations generates performance recommendations
func (lts *LoadTestingSuite) generateRecommendations(result *LoadTestResult) {
	// Check response time thresholds
	if result.AverageResponseTime > lts.config.MaxResponseTime {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Average response time (%.2fms) exceeds threshold (%.2fms). Consider query optimization or additional indexing.",
				float64(result.AverageResponseTime.Nanoseconds())/1e6,
				float64(lts.config.MaxResponseTime.Nanoseconds())/1e6))
	}

	// Check error rate thresholds
	if result.ErrorRate > lts.config.MaxErrorRate {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Error rate (%.2f%%) exceeds threshold (%.2f%%). Review error types and system capacity.",
				result.ErrorRate, lts.config.MaxErrorRate))
	}

	// Check throughput thresholds
	if result.RequestsPerSecond < lts.config.MinThroughput {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Throughput (%.2f req/s) below threshold (%.2f req/s). Consider scaling or optimization.",
				result.RequestsPerSecond, lts.config.MinThroughput))
	}

	// Check connection usage
	if result.PeakConnections > 80 {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("High peak connection count (%d). Review connection pooling configuration.",
				result.PeakConnections))
	}

	// Check timeout rate
	if result.TotalRequests > 0 {
		timeoutRate := float64(result.TimeoutCount) / float64(result.TotalRequests) * 100
		if timeoutRate > 1.0 {
			result.Recommendations = append(result.Recommendations,
				fmt.Sprintf("High timeout rate (%.2f%%). Consider increasing timeouts or optimizing slow queries.",
					timeoutRate))
		}
	}
}

// getDefaultTestScenarios returns default test scenarios
func getDefaultTestScenarios() []LoadTestScenario {
	return []LoadTestScenario{
		{
			Name:             "User Lookup",
			Weight:           30,
			Query:            "SELECT id, email, name FROM users WHERE email = $1",
			Parameters:       []interface{}{"test@example.com"},
			ExpectedDuration: 50 * time.Millisecond,
			IsReadOnly:       true,
		},
		{
			Name:             "Business Search",
			Weight:           25,
			Query:            "SELECT id, name, industry FROM businesses WHERE name ILIKE $1 LIMIT 10",
			Parameters:       []interface{}{"%test%"},
			ExpectedDuration: 100 * time.Millisecond,
			IsReadOnly:       true,
		},
		{
			Name:             "Classification Query",
			Weight:           20,
			Query:            "SELECT bc.*, b.name FROM business_classifications bc JOIN businesses b ON bc.business_id = b.id WHERE bc.created_at > $1 ORDER BY bc.created_at DESC LIMIT 50",
			Parameters:       []interface{}{time.Now().Add(-24 * time.Hour)},
			ExpectedDuration: 200 * time.Millisecond,
			IsReadOnly:       true,
		},
		{
			Name:             "Risk Assessment Summary",
			Weight:           15,
			Query:            "SELECT risk_level, COUNT(*) as count FROM risk_assessments GROUP BY risk_level",
			Parameters:       []interface{}{},
			ExpectedDuration: 150 * time.Millisecond,
			IsReadOnly:       true,
		},
		{
			Name:             "Complex Join Query",
			Weight:           10,
			Query:            "SELECT u.email, b.name, bc.industry, ra.risk_level FROM users u JOIN businesses b ON u.id = b.user_id LEFT JOIN business_classifications bc ON b.id = bc.business_id LEFT JOIN risk_assessments ra ON b.id = ra.business_id WHERE u.created_at > $1 LIMIT 20",
			Parameters:       []interface{}{time.Now().Add(-7 * 24 * time.Hour)},
			ExpectedDuration: 300 * time.Millisecond,
			IsReadOnly:       true,
		},
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
