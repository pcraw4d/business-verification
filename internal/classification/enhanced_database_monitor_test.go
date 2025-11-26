package classification

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap/zaptest"
)

// createEnhancedTestDB creates an in-memory SQLite database for testing
func createEnhancedTestDB() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	return db
}

func TestEnhancedDatabaseMonitor_NewEnhancedDatabaseMonitor(t *testing.T) {
	tests := []struct {
		name   string
		config *EnhancedDatabaseConfig
	}{
		{
			name:   "with default config",
			config: nil,
		},
		{
			name: "with custom config",
			config: &EnhancedDatabaseConfig{
				Enabled:               true,
				CollectionInterval:    10 * time.Second,
				SlowQueryThreshold:    50 * time.Millisecond,
				MaxQueryStats:         500,
				OptimizationCacheSize: 250,
				AlertingEnabled:       true,
				TrackQueryPlans:       true,
				TrackIndexUsage:       true,
				TrackConnectionPool:   true,
				TrackLockContention:   true,
				TrackDeadlocks:        true,
				TrackCacheHitRatio:    true,
				TrackTableSizes:       true,
				TrackIndexSizes:       true,
				TrackVacuumStats:      true,
				TrackReplicationLag:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := createEnhancedTestDB()
			defer db.Close()
			logger := zaptest.NewLogger(t)

			monitor := NewEnhancedDatabaseMonitor(db, logger, tt.config)

			if monitor == nil {
				t.Fatal("Expected monitor to be created, got nil")
			}

			if monitor.db != db {
				t.Error("Expected monitor to use the provided database")
			}

			if monitor.logger != logger {
				t.Error("Expected monitor to use the provided logger")
			}

			if tt.config == nil {
				// Check default config
				if !monitor.config.Enabled {
					t.Error("Expected default config to have monitoring enabled")
				}
				if monitor.config.CollectionInterval != 30*time.Second {
					t.Error("Expected default collection interval to be 30 seconds")
				}
				if monitor.config.SlowQueryThreshold != 100*time.Millisecond {
					t.Error("Expected default slow query threshold to be 100ms")
				}
			} else {
				// Check custom config
				if monitor.config.Enabled != tt.config.Enabled {
					t.Error("Expected monitor to use custom enabled setting")
				}
				if monitor.config.CollectionInterval != tt.config.CollectionInterval {
					t.Error("Expected monitor to use custom collection interval")
				}
				if monitor.config.SlowQueryThreshold != tt.config.SlowQueryThreshold {
					t.Error("Expected monitor to use custom slow query threshold")
				}
			}

			// Clean up
			monitor.Stop()
		})
	}
}

func TestEnhancedDatabaseMonitor_StartStop_Enhanced(b *testing.T) {
	db := createEnhancedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(b)
	config := &EnhancedDatabaseConfig{
		Enabled:            true,
		CollectionInterval: 100 * time.Millisecond,
	}

	monitor := NewEnhancedDatabaseMonitor(db, logger, config)

	// Test starting
	monitor.Start()
	time.Sleep(50 * time.Millisecond) // Give it a moment to start

	// Test stopping
	monitor.Stop()
	time.Sleep(50 * time.Millisecond) // Give it a moment to stop

	// Verify that the monitor stopped without panicking
	// Further checks could involve inspecting logs or mock channels if they were used
}

func TestEnhancedDatabaseMonitor_RecordQueryExecution(t *testing.T) {
	tests := []struct {
		name          string
		queryText     string
		executionTime time.Duration
		rowsReturned  int64
		rowsExamined  int64
		errorOccurred bool
		errorMessage  string
		expectedStats int
	}{
		{
			name:          "successful query execution",
			queryText:     "SELECT * FROM users WHERE id = ?",
			executionTime: 50 * time.Millisecond,
			rowsReturned:  1,
			rowsExamined:  1,
			errorOccurred: false,
			errorMessage:  "",
			expectedStats: 1,
		},
		{
			name:          "slow query execution",
			queryText:     "SELECT * FROM large_table WHERE complex_condition = ?",
			executionTime: 200 * time.Millisecond,
			rowsReturned:  10,
			rowsExamined:  1000,
			errorOccurred: false,
			errorMessage:  "",
			expectedStats: 1,
		},
		{
			name:          "query with error",
			queryText:     "SELECT * FROM non_existent_table",
			executionTime: 10 * time.Millisecond,
			rowsReturned:  0,
			rowsExamined:  0,
			errorOccurred: true,
			errorMessage:  "table does not exist",
			expectedStats: 1,
		},
		{
			name:          "multiple executions of same query",
			queryText:     "SELECT * FROM users WHERE active = true",
			executionTime: 30 * time.Millisecond,
			rowsReturned:  5,
			rowsExamined:  5,
			errorOccurred: false,
			errorMessage:  "",
			expectedStats: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := createEnhancedTestDB()
			defer db.Close()
			logger := zaptest.NewLogger(t)
			config := &EnhancedDatabaseConfig{
				Enabled:               true,
				CollectionInterval:    30 * time.Second,
				SlowQueryThreshold:    100 * time.Millisecond,
				MaxQueryStats:         1000,
				OptimizationCacheSize: 500,
				AlertingEnabled:       true,
			}

			monitor := NewEnhancedDatabaseMonitor(db, logger, config)
			defer monitor.Stop()

			// Record multiple executions for the same query to test statistics
			executions := 3
			for i := 0; i < executions; i++ {
				monitor.RecordQueryExecution(
					context.Background(),
					tt.queryText,
					tt.executionTime,
					tt.rowsReturned,
					tt.rowsExamined,
					tt.errorOccurred,
					tt.errorMessage,
				)
			}

			// Get query stats
			stats := monitor.GetQueryStats(10)
			if len(stats) != tt.expectedStats {
				t.Errorf("Expected %d query stats, got %d", tt.expectedStats, len(stats))
			}

			// Verify the stats for the recorded query
			for _, stat := range stats {
				if stat.QueryText != tt.queryText {
					continue
				}

				if stat.ExecutionCount != int64(executions) {
					t.Errorf("Expected execution count %d, got %d", executions, stat.ExecutionCount)
				}

				if stat.AverageExecutionTime != float64(tt.executionTime.Milliseconds()) {
					t.Errorf("Expected average execution time %.2f, got %.2f",
						float64(tt.executionTime.Milliseconds()), stat.AverageExecutionTime)
				}

				if stat.RowsReturned != tt.rowsReturned*int64(executions) {
					t.Errorf("Expected rows returned %d, got %d",
						tt.rowsReturned*int64(executions), stat.RowsReturned)
				}

				if stat.RowsExamined != tt.rowsExamined*int64(executions) {
					t.Errorf("Expected rows examined %d, got %d",
						tt.rowsExamined*int64(executions), stat.RowsExamined)
				}

				if tt.errorOccurred && stat.ErrorCount != int64(executions) {
					t.Errorf("Expected error count %d, got %d", executions, stat.ErrorCount)
				}

				// Check performance category
				if tt.executionTime > config.SlowQueryThreshold {
					if stat.PerformanceCategory != "poor" && stat.PerformanceCategory != "critical" {
						t.Errorf("Expected performance category to be poor or critical for slow query, got %s",
							stat.PerformanceCategory)
					}
				}

				// Check optimization score
				if stat.OptimizationScore < 0 || stat.OptimizationScore > 100 {
					t.Errorf("Expected optimization score between 0 and 100, got %.2f", stat.OptimizationScore)
				}
			}
		})
	}
}

func TestEnhancedDatabaseMonitor_GetPerformanceAlerts(t *testing.T) {
	db := createEnhancedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(t)
	config := &EnhancedDatabaseConfig{
		Enabled:               true,
		CollectionInterval:    30 * time.Second,
		SlowQueryThreshold:    50 * time.Millisecond, // Low threshold to trigger alerts
		MaxQueryStats:         1000,
		OptimizationCacheSize: 500,
		AlertingEnabled:       true,
	}

	monitor := NewEnhancedDatabaseMonitor(db, logger, config)
	defer monitor.Stop()

	// Record a slow query to trigger an alert
	monitor.RecordQueryExecution(
		context.Background(),
		"SELECT * FROM slow_table",
		200*time.Millisecond, // Exceeds threshold
		1,
		1000, // High rows examined
		false,
		"",
	)

	// Record a query with errors to trigger an error alert
	for i := 0; i < 15; i++ { // Need more than 10 executions with >10% error rate
		monitor.RecordQueryExecution(
			context.Background(),
			"SELECT * FROM error_table",
			10*time.Millisecond,
			0,
			0,
			i < 3, // 3 errors out of 15 = 20% error rate
			"test error",
		)
	}

	// Get active alerts
	alerts := monitor.GetPerformanceAlerts(false, 10)
	if len(alerts) == 0 {
		t.Error("Expected to have performance alerts, but got none")
	}

	// Check alert types
	alertTypes := make(map[string]bool)
	for _, alert := range alerts {
		alertTypes[alert.AlertType] = true

		if alert.Severity == "" {
			t.Error("Expected alert to have a severity level")
		}

		if alert.Message == "" {
			t.Error("Expected alert to have a message")
		}

		if len(alert.Recommendations) == 0 {
			t.Error("Expected alert to have recommendations")
		}
	}

	// Should have slow query and high error rate alerts
	if !alertTypes["slow_query"] {
		t.Error("Expected to have slow_query alert type")
	}
	if !alertTypes["high_error_rate"] {
		t.Error("Expected to have high_error_rate alert type")
	}
}

func TestEnhancedDatabaseMonitor_GetOptimizationRecommendations(t *testing.T) {
	db := createEnhancedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(t)
	config := &EnhancedDatabaseConfig{
		Enabled:               true,
		CollectionInterval:    30 * time.Second,
		SlowQueryThreshold:    50 * time.Millisecond,
		MaxQueryStats:         1000,
		OptimizationCacheSize: 500,
		AlertingEnabled:       true,
	}

	monitor := NewEnhancedDatabaseMonitor(db, logger, config)
	defer monitor.Stop()

	// Record multiple slow queries to generate optimization recommendations
	for i := 0; i < 10; i++ {
		monitor.RecordQueryExecution(
			context.Background(),
			"SELECT * FROM slow_table WHERE complex_condition = ?",
			200*time.Millisecond, // Slow query
			1,
			1000, // High rows examined
			false,
			"",
		)
	}

	// Get optimization recommendations
	recommendations := monitor.GetOptimizationRecommendations("", 10)
	if len(recommendations) == 0 {
		t.Error("Expected to have optimization recommendations, but got none")
	}

	// Check recommendation structure
	for _, optimization := range recommendations {
		if optimization.QueryID == "" {
			t.Error("Expected optimization to have a query ID")
		}

		if optimization.QueryText == "" {
			t.Error("Expected optimization to have query text")
		}

		if optimization.OptimizationScore < 0 || optimization.OptimizationScore > 100 {
			t.Error("Expected optimization score between 0 and 100")
		}

		if optimization.Priority == "" {
			t.Error("Expected optimization to have a priority")
		}

		if len(optimization.Recommendations) == 0 {
			t.Error("Expected optimization to have recommendations")
		}

		// Check recommendation structure
		for _, rec := range optimization.Recommendations {
			if rec.Type == "" {
				t.Error("Expected recommendation to have a type")
			}

			if rec.Description == "" {
				t.Error("Expected recommendation to have a description")
			}

			if rec.Impact == "" {
				t.Error("Expected recommendation to have an impact level")
			}

			if rec.Effort == "" {
				t.Error("Expected recommendation to have an effort level")
			}
		}
	}
}

func TestEnhancedDatabaseMonitor_GetDatabaseSummary(t *testing.T) {
	db := createEnhancedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(t)
	config := &EnhancedDatabaseConfig{
		Enabled:               true,
		CollectionInterval:    30 * time.Second,
		SlowQueryThreshold:    50 * time.Millisecond,
		MaxQueryStats:         1000,
		OptimizationCacheSize: 500,
		AlertingEnabled:       true,
	}

	monitor := NewEnhancedDatabaseMonitor(db, logger, config)
	defer monitor.Stop()

	// Record various types of queries
	queries := []struct {
		text          string
		executionTime time.Duration
		rowsReturned  int64
		rowsExamined  int64
		errorOccurred bool
		executions    int
	}{
		{"SELECT * FROM fast_table", 10 * time.Millisecond, 1, 1, false, 5},
		{"SELECT * FROM slow_table", 200 * time.Millisecond, 1, 1000, false, 3},
		{"SELECT * FROM error_table", 20 * time.Millisecond, 0, 0, true, 10},
	}

	for _, q := range queries {
		for i := 0; i < q.executions; i++ {
			monitor.RecordQueryExecution(
				context.Background(),
				q.text,
				q.executionTime,
				q.rowsReturned,
				q.rowsExamined,
				q.errorOccurred,
				"test error",
			)
		}
	}

	// Get database summary
	summary := monitor.GetDatabaseSummary()
	if summary == nil {
		t.Fatal("Expected database summary, got nil")
	}

	// Check summary structure
	requiredKeys := []string{"timestamp", "queries", "alerts", "optimizations", "configuration"}
	for _, key := range requiredKeys {
		if _, exists := summary[key]; !exists {
			t.Errorf("Expected summary to have key: %s", key)
		}
	}

	// Check queries section
	queriesSection, ok := summary["queries"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected queries section to be a map")
	}

	requiredQueryKeys := []string{"total_queries", "slow_queries", "high_error_queries", "avg_optimization_score"}
	for _, key := range requiredQueryKeys {
		if _, exists := queriesSection[key]; !exists {
			t.Errorf("Expected queries section to have key: %s", key)
		}
	}

	// Check alerts section
	alertsSection, ok := summary["alerts"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected alerts section to be a map")
	}

	requiredAlertKeys := []string{"total_alerts", "active_alerts", "resolved_alerts"}
	for _, key := range requiredAlertKeys {
		if _, exists := alertsSection[key]; !exists {
			t.Errorf("Expected alerts section to have key: %s", key)
		}
	}

	// Check optimizations section
	optimizationsSection, ok := summary["optimizations"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected optimizations section to be a map")
	}

	requiredOptimizationKeys := []string{"total_optimizations", "high_priority", "medium_priority", "low_priority"}
	for _, key := range requiredOptimizationKeys {
		if _, exists := optimizationsSection[key]; !exists {
			t.Errorf("Expected optimizations section to have key: %s", key)
		}
	}

	// Check configuration section
	configSection, ok := summary["configuration"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected configuration section to be a map")
	}

	requiredConfigKeys := []string{"slow_query_threshold_ms", "collection_interval_sec", "max_query_stats", "alerting_enabled"}
	for _, key := range requiredConfigKeys {
		if _, exists := configSection[key]; !exists {
			t.Errorf("Expected configuration section to have key: %s", key)
		}
	}
}

func TestEnhancedDatabaseMonitor_QueryHashGeneration(t *testing.T) {
	db := createEnhancedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(t)
	config := &EnhancedDatabaseConfig{
		Enabled: true,
	}

	monitor := NewEnhancedDatabaseMonitor(db, logger, config)
	defer monitor.Stop()

	// Test that same query text generates same hash
	query1 := "SELECT * FROM users WHERE id = ?"
	query2 := "SELECT * FROM users WHERE id = ?"
	query3 := "SELECT * FROM users WHERE name = ?"

	hash1 := monitor.generateQueryHash(query1)
	hash2 := monitor.generateQueryHash(query2)
	hash3 := monitor.generateQueryHash(query3)

	if hash1 != hash2 {
		t.Error("Expected same query text to generate same hash")
	}

	if hash1 == hash3 {
		t.Error("Expected different query text to generate different hash")
	}
}

func TestEnhancedDatabaseMonitor_QueryIDGeneration(t *testing.T) {
	db := createEnhancedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(t)
	config := &EnhancedDatabaseConfig{
		Enabled: true,
	}

	monitor := NewEnhancedDatabaseMonitor(db, logger, config)
	defer monitor.Stop()

	// Test that query IDs are unique
	query := "SELECT * FROM users"

	id1 := monitor.generateQueryID(query)
	time.Sleep(1 * time.Millisecond) // Ensure different timestamp
	id2 := monitor.generateQueryID(query)

	if id1 == id2 {
		t.Error("Expected query IDs to be unique")
	}

	if id1 == "" {
		t.Error("Expected query ID to be non-empty")
	}
}

func TestEnhancedDatabaseMonitor_AlertSeverityDetermination(t *testing.T) {
	db := createEnhancedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(t)
	config := &EnhancedDatabaseConfig{
		Enabled: true,
	}

	monitor := NewEnhancedDatabaseMonitor(db, logger, config)
	defer monitor.Stop()

	tests := []struct {
		actual    float64
		threshold float64
		expected  string
	}{
		{50, 100, "low"},       // 0.5x threshold
		{150, 100, "low"},      // 1.5x threshold
		{200, 100, "medium"},   // 2x threshold
		{300, 100, "high"},     // 3x threshold
		{500, 100, "critical"}, // 5x threshold
		{600, 100, "critical"}, // 6x threshold
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("actual_%.0f_threshold_%.0f", tt.actual, tt.threshold), func(t *testing.T) {
			severity := monitor.determineAlertSeverity(tt.actual, tt.threshold)
			if severity != tt.expected {
				t.Errorf("Expected severity %s, got %s", tt.expected, severity)
			}
		})
	}
}

func BenchmarkEnhancedDatabaseMonitor_RecordQueryExecution(b *testing.B) {
	db := createEnhancedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(b)
	config := &EnhancedDatabaseConfig{
		Enabled:               true,
		CollectionInterval:    30 * time.Second,
		SlowQueryThreshold:    100 * time.Millisecond,
		MaxQueryStats:         1000,
		OptimizationCacheSize: 500,
		AlertingEnabled:       false, // Disable alerting for benchmark
	}

	monitor := NewEnhancedDatabaseMonitor(db, logger, config)
	defer monitor.Stop()

	queryText := "SELECT * FROM benchmark_table WHERE id = ?"
	executionTime := 50 * time.Millisecond
	rowsReturned := int64(1)
	rowsExamined := int64(1)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			monitor.RecordQueryExecution(
				context.Background(),
				queryText,
				executionTime,
				rowsReturned,
				rowsExamined,
				false,
				"",
			)
		}
	})
}

func BenchmarkEnhancedDatabaseMonitor_GetQueryStats(b *testing.B) {
	db := createEnhancedTestDB()
	defer db.Close()
	logger := zaptest.NewLogger(b)
	config := &EnhancedDatabaseConfig{
		Enabled:               true,
		CollectionInterval:    30 * time.Second,
		SlowQueryThreshold:    100 * time.Millisecond,
		MaxQueryStats:         1000,
		OptimizationCacheSize: 500,
		AlertingEnabled:       false,
	}

	monitor := NewEnhancedDatabaseMonitor(db, logger, config)
	defer monitor.Stop()

	// Pre-populate with some query stats
	for i := 0; i < 100; i++ {
		monitor.RecordQueryExecution(
			context.Background(),
			fmt.Sprintf("SELECT * FROM table_%d WHERE id = ?", i),
			50*time.Millisecond,
			1,
			1,
			false,
			"",
		)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.GetQueryStats(50)
	}
}
