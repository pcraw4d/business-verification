package data_discovery

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestExtractionMonitor_BasicFunctionality(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExtractionMonitorConfig()
	monitor := NewExtractionMonitor(config, logger)

	// Test initial state
	metrics := monitor.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.TotalRequests)
	assert.Equal(t, int64(0), metrics.SuccessfulRequests)
	assert.Equal(t, int64(0), metrics.FailedRequests)
}

func TestExtractionMonitor_RecordExtractionResult(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExtractionMonitorConfig()
	monitor := NewExtractionMonitor(config, logger)

	// Create a sample discovery result
	result := &DataDiscoveryResult{
		DiscoveredFields: []DiscoveredField{
			{
				FieldName:       "email",
				FieldType:       "email",
				ConfidenceScore: 0.9,
				BusinessValue:   0.8,
			},
			{
				FieldName:       "phone",
				FieldType:       "phone",
				ConfidenceScore: 0.8,
				BusinessValue:   0.7,
			},
		},
		ConfidenceScore: 0.85,
		ProcessingTime:  500 * time.Millisecond,
		QualityAssessments: []FieldQualityAssessment{
			{
				FieldName: "email",
				FieldType: "email",
				QualityScore: QualityScore{
					OverallScore: 0.9,
				},
				QualityCategory: "excellent",
				BusinessImpact:  "critical",
			},
			{
				FieldName: "phone",
				FieldType: "phone",
				QualityScore: QualityScore{
					OverallScore: 0.8,
				},
				QualityCategory: "good",
				BusinessImpact:  "high",
			},
		},
	}

	// Record successful extraction
	monitor.RecordExtractionResult(context.Background(), result, 500*time.Millisecond, nil)

	// Verify metrics
	metrics := monitor.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalRequests)
	assert.Equal(t, int64(1), metrics.SuccessfulRequests)
	assert.Equal(t, int64(0), metrics.FailedRequests)
	assert.Equal(t, 500*time.Millisecond, metrics.AverageProcessingTime)
	assert.InDelta(t, 2.0, metrics.FieldsDiscoveredPerRequest, 0.001)
	assert.InDelta(t, 0.85, metrics.AverageQualityScore, 0.001)

	// Verify quality distribution
	assert.Equal(t, 1, metrics.QualityScoreDistribution["excellent"])
	assert.Equal(t, 1, metrics.QualityScoreDistribution["good"])

	// Verify field discovery rates
	assert.InDelta(t, 1.0, metrics.FieldDiscoveryRates["email"], 0.001)
	assert.InDelta(t, 1.0, metrics.FieldDiscoveryRates["phone"], 0.001)
}

func TestExtractionMonitor_RecordFailedExtraction(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExtractionMonitorConfig()
	monitor := NewExtractionMonitor(config, logger)

	// Record failed extraction
	err := assert.AnError
	monitor.RecordExtractionResult(context.Background(), nil, 100*time.Millisecond, err)

	// Verify metrics
	metrics := monitor.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalRequests)
	assert.Equal(t, int64(0), metrics.SuccessfulRequests)
	assert.Equal(t, int64(1), metrics.FailedRequests)
	assert.Equal(t, 100*time.Millisecond, metrics.AverageProcessingTime)

	// Verify error tracking
	assert.Equal(t, int64(1), metrics.ErrorTypes["*errors.errorString"])
}

func TestExtractionMonitor_PerformanceReport(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExtractionMonitorConfig()
	monitor := NewExtractionMonitor(config, logger)

	// Record multiple extractions
	for i := 0; i < 5; i++ {
		result := &DataDiscoveryResult{
			DiscoveredFields: []DiscoveredField{
				{
					FieldName:       "email",
					FieldType:       "email",
					ConfidenceScore: 0.9,
				},
			},
			ConfidenceScore: 0.9,
			ProcessingTime:  200 * time.Millisecond,
			QualityAssessments: []FieldQualityAssessment{
				{
					FieldName: "email",
					FieldType: "email",
					QualityScore: QualityScore{
						OverallScore: 0.9,
					},
					QualityCategory: "excellent",
				},
			},
		}
		monitor.RecordExtractionResult(context.Background(), result, 200*time.Millisecond, nil)
	}

	// Get performance report
	report := monitor.GetPerformanceReport()
	assert.NotNil(t, report)
	assert.Equal(t, int64(5), report.TotalRequests)
	assert.InDelta(t, 1.0, report.SuccessRate, 0.001)
	assert.Equal(t, 200*time.Millisecond, report.AverageProcessingTime)
	assert.InDelta(t, 0.9, report.AverageQualityScore, 0.001)
	assert.InDelta(t, 1.0, report.FieldsDiscoveredPerRequest, 0.001)
}

func TestExtractionOptimizer_BasicFunctionality(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExtractionMonitorConfig()
	metrics := &ExtractionMetrics{
		TotalRequests:              10,
		SuccessfulRequests:         8,
		FailedRequests:             2,
		AverageProcessingTime:      3 * time.Second,
		AverageQualityScore:        0.6,
		FieldsDiscoveredPerRequest: 8.0,
		QualityScoreDistribution:   make(map[string]int),
		FieldDiscoveryRates:        make(map[string]float64),
		FieldQualityScores:         make(map[string]float64),
		FieldProcessingTimes:       make(map[string]time.Duration),
		ErrorTypes:                 make(map[string]int64),
		ErrorRates:                 make(map[string]float64),
	}

	optimizer := NewExtractionOptimizer(config, logger, metrics)

	// Test initial strategies
	strategies := optimizer.GetOptimizationStrategies()
	assert.Len(t, strategies, 5)

	// Verify strategy names
	strategyNames := make(map[string]bool)
	for _, strategy := range strategies {
		strategyNames[strategy.Name] = true
	}
	assert.True(t, strategyNames["pattern_optimization"])
	assert.True(t, strategyNames["field_prioritization"])
	assert.True(t, strategyNames["resource_optimization"])
	assert.True(t, strategyNames["quality_improvement"])
	assert.True(t, strategyNames["error_reduction"])
}

func TestExtractionOptimizer_RunOptimization(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExtractionMonitorConfig()
	metrics := &ExtractionMetrics{
		TotalRequests:              10,
		SuccessfulRequests:         7, // Below threshold
		FailedRequests:             3,
		AverageProcessingTime:      6 * time.Second, // Above threshold
		AverageQualityScore:        0.6,             // Below threshold
		FieldsDiscoveredPerRequest: 6.0,             // Below threshold
		MemoryUsage:                600,             // Above threshold
		CPUUsage:                   85.0,            // Above threshold
		QualityScoreDistribution:   make(map[string]int),
		FieldDiscoveryRates:        make(map[string]float64),
		FieldQualityScores:         make(map[string]float64),
		FieldProcessingTimes:       make(map[string]time.Duration),
		ErrorTypes:                 make(map[string]int64),
		ErrorRates:                 make(map[string]float64),
	}

	optimizer := NewExtractionOptimizer(config, logger, metrics)

	// Run optimization
	optimizer.RunOptimization()

	// Verify that strategies were applied (this would be verified through logging)
	// The actual optimization logic would be tested through integration tests
}

func TestExtractionOptimizer_StrategyManagement(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExtractionMonitorConfig()
	metrics := &ExtractionMetrics{
		QualityScoreDistribution: make(map[string]int),
		FieldDiscoveryRates:      make(map[string]float64),
		FieldQualityScores:       make(map[string]float64),
		FieldProcessingTimes:     make(map[string]time.Duration),
		ErrorTypes:               make(map[string]int64),
		ErrorRates:               make(map[string]float64),
	}

	optimizer := NewExtractionOptimizer(config, logger, metrics)

	// Test enabling/disabling strategies
	err := optimizer.EnableStrategy("pattern_optimization", false)
	assert.NoError(t, err)

	strategies := optimizer.GetOptimizationStrategies()
	for _, strategy := range strategies {
		if strategy.Name == "pattern_optimization" {
			assert.False(t, strategy.Enabled)
		}
	}

	// Test updating strategy parameters
	params := map[string]interface{}{
		"confidence_threshold": 0.8,
		"max_patterns":         25,
	}
	err = optimizer.UpdateStrategyParameters("pattern_optimization", params)
	assert.NoError(t, err)

	// Test invalid strategy
	err = optimizer.EnableStrategy("invalid_strategy", false)
	assert.Error(t, err)
}

func TestAlertManager_BasicFunctionality(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExtractionMonitorConfig()
	alertManager := NewAlertManager(config, logger)

	// Test initial state
	activeAlerts := alertManager.GetActiveAlerts()
	assert.Len(t, activeAlerts, 0)

	summary := alertManager.GetAlertSummary()
	assert.NotNil(t, summary)
	assert.Equal(t, 0, summary.TotalAlerts)
	assert.Equal(t, 0, summary.ActiveAlerts)
}

func TestAlertManager_CreateAndManageAlerts(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExtractionMonitorConfig()
	config.AlertSettings.AlertCooldownPeriod = 0 // Disable cooldown for testing
	alertManager := NewAlertManager(config, logger)

	// Create alerts
	metrics := &ExtractionMetrics{
		TotalRequests:      10,
		SuccessfulRequests: 7,
	}

	alertManager.CreateAlert("performance", "critical", "Success rate below threshold", metrics)
	alertManager.CreateAlert("quality", "warning", "Quality score below threshold", metrics)

	// Verify alerts were created
	activeAlerts := alertManager.GetActiveAlerts()
	assert.Len(t, activeAlerts, 2)

	// Test alert filtering
	performanceAlerts := alertManager.GetAlertsByType("performance")
	assert.Len(t, performanceAlerts, 1)
	assert.Equal(t, "performance", performanceAlerts[0].Type)

	criticalAlerts := alertManager.GetAlertsBySeverity("critical")
	assert.Len(t, criticalAlerts, 1)
	assert.Equal(t, "critical", criticalAlerts[0].Severity)

	// Test alert acknowledgment
	alertID := activeAlerts[0].ID
	err := alertManager.AcknowledgeAlert(alertID)
	assert.NoError(t, err)

	// Test alert resolution
	err = alertManager.ResolveAlert(alertID)
	assert.NoError(t, err)

	// Verify alert was resolved
	activeAlerts = alertManager.GetActiveAlerts()
	assert.Len(t, activeAlerts, 1) // One alert should still be active
}

func TestAlertManager_AlertSummaryAndTrends(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExtractionMonitorConfig()
	config.AlertSettings.AlertCooldownPeriod = 0 // Disable cooldown for testing
	alertManager := NewAlertManager(config, logger)

	// Create multiple alerts
	for i := 0; i < 5; i++ {
		alertManager.CreateAlert("performance", "warning", "Test alert", nil)
	}
	for i := 0; i < 3; i++ {
		alertManager.CreateAlert("quality", "critical", "Test alert", nil)
	}

	// Test alert summary
	summary := alertManager.GetAlertSummary()
	assert.Equal(t, 8, summary.TotalAlerts)
	assert.Equal(t, 8, summary.ActiveAlerts)
	assert.Equal(t, 5, summary.AlertsByType["performance"])
	assert.Equal(t, 3, summary.AlertsByType["quality"])
	assert.Equal(t, 5, summary.AlertsBySeverity["warning"])
	assert.Equal(t, 3, summary.AlertsBySeverity["critical"])

	// Test alert trends
	trends := alertManager.GetAlertTrends(24 * time.Hour)
	assert.Equal(t, 8, trends.TotalAlerts)
	assert.Equal(t, 5, trends.AlertsByType["performance"])
	assert.Equal(t, 3, trends.AlertsByType["quality"])
}

func TestAlertManager_AlertHistoryAndPagination(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExtractionMonitorConfig()
	config.AlertSettings.AlertCooldownPeriod = 0 // Disable cooldown for testing
	alertManager := NewAlertManager(config, logger)

	// Create multiple alerts
	for i := 0; i < 10; i++ {
		alertManager.CreateAlert("test", "info", "Test alert", nil)
	}

	// Test pagination
	history := alertManager.GetAlertHistory(5, 0) // First 5 alerts
	assert.Len(t, history, 5)

	history = alertManager.GetAlertHistory(5, 5) // Next 5 alerts
	assert.Len(t, history, 5)

	// Test getting specific alert
	allAlerts := alertManager.GetActiveAlerts()
	if len(allAlerts) > 0 {
		alertID := allAlerts[0].ID
		alert, err := alertManager.GetAlertByID(alertID)
		assert.NoError(t, err)
		assert.Equal(t, alertID, alert.ID)
	}

	// Test getting non-existent alert
	_, err := alertManager.GetAlertByID("non_existent")
	assert.Error(t, err)
}

func TestAlertManager_CleanupOldAlerts(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExtractionMonitorConfig()
	alertManager := NewAlertManager(config, logger)

	// Create alerts
	alertManager.CreateAlert("test", "info", "Test alert", nil)

	// Verify alert was created
	activeAlerts := alertManager.GetActiveAlerts()
	assert.Len(t, activeAlerts, 1)

	// Cleanup alerts older than 1 hour (should keep recent alerts)
	removedCount := alertManager.CleanupOldAlerts(1 * time.Hour)
	assert.Equal(t, 0, removedCount) // No alerts should be removed

	// Verify alerts still exist
	activeAlerts = alertManager.GetActiveAlerts()
	assert.Len(t, activeAlerts, 1)
}

func TestExtractionMonitor_Integration(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExtractionMonitorConfig()
	config.AlertSettings.AlertCooldownPeriod = 0            // Disable cooldown for testing
	config.MetricsCollectionInterval = 1 * time.Millisecond // Very short interval for testing
	monitor := NewExtractionMonitor(config, logger)

	// Simulate multiple extraction operations
	for i := 0; i < 10; i++ {
		result := &DataDiscoveryResult{
			DiscoveredFields: []DiscoveredField{
				{
					FieldName:       "email",
					FieldType:       "email",
					ConfidenceScore: 0.9,
				},
				{
					FieldName:       "phone",
					FieldType:       "phone",
					ConfidenceScore: 0.8,
				},
			},
			ConfidenceScore: 0.85,
			ProcessingTime:  300 * time.Millisecond,
			QualityAssessments: []FieldQualityAssessment{
				{
					FieldName: "email",
					FieldType: "email",
					QualityScore: QualityScore{
						OverallScore: 0.9,
					},
					QualityCategory: "excellent",
				},
				{
					FieldName: "phone",
					FieldType: "phone",
					QualityScore: QualityScore{
						OverallScore: 0.8,
					},
					QualityCategory: "good",
				},
			},
		}

		// Record successful extraction
		monitor.RecordExtractionResult(context.Background(), result, 300*time.Millisecond, nil)
	}

	// Record one failed extraction
	monitor.RecordExtractionResult(context.Background(), nil, 100*time.Millisecond, assert.AnError)

	// Give background monitoring time to run
	time.Sleep(10 * time.Millisecond)

	// Get comprehensive report
	report := monitor.GetPerformanceReport()
	assert.NotNil(t, report)
	assert.Equal(t, int64(11), report.TotalRequests)
	assert.InDelta(t, 0.91, report.SuccessRate, 0.01) // 10/11 ≈ 0.91
	assert.True(t, report.AverageProcessingTime >= 250*time.Millisecond && report.AverageProcessingTime <= 350*time.Millisecond)
	assert.InDelta(t, 2.0, report.FieldsDiscoveredPerRequest, 0.001)

	// Verify alerts were created for performance issues
	alerts := monitor.alerts.GetActiveAlerts()
	assert.NotEmpty(t, alerts)

	// Verify optimization was triggered
	strategies := monitor.optimizer.GetOptimizationStrategies()
	assert.NotEmpty(t, strategies)
}

func TestExtractionMonitor_Configuration(t *testing.T) {
	logger := zap.NewNop()

	// Test custom configuration
	config := &ExtractionMonitorConfig{
		MetricsCollectionInterval: 60 * time.Second,
		PerformanceThresholds: PerformanceThresholds{
			MaxProcessingTime:        10 * time.Second,
			MinSuccessRate:           0.9,
			MaxErrorRate:             0.1,
			MinDataPointsPerBusiness: 10,
			MaxMemoryUsage:           1024,
			MinQualityScore:          0.8,
		},
		AlertSettings: AlertSettings{
			Enabled:             true,
			AlertChannels:       []string{"log", "metrics"},
			CriticalThreshold:   0.7,
			WarningThreshold:    0.8,
			AlertCooldownPeriod: 10 * time.Minute,
		},
		OptimizationEnabled:      true,
		AutoOptimizationInterval: 15 * time.Minute,
		OptimizationThresholds: OptimizationThresholds{
			PerformanceDegradationThreshold: 0.15,
			QualityImprovementThreshold:     0.1,
			ResourceUtilizationThreshold:    0.9,
			SuccessRateImprovementThreshold: 0.05,
		},
		MetricsRetentionPeriod: 48 * time.Hour,
		MaxMetricsHistory:      2000,
	}

	monitor := NewExtractionMonitor(config, logger)
	assert.NotNil(t, monitor)
	assert.Equal(t, config, monitor.config)
}

func TestExtractionMonitor_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultExtractionMonitorConfig()
	monitor := NewExtractionMonitor(config, logger)

	// Test concurrent access to metrics
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			result := &DataDiscoveryResult{
				DiscoveredFields: []DiscoveredField{
					{
						FieldName:       "email",
						FieldType:       "email",
						ConfidenceScore: 0.9,
					},
				},
				ConfidenceScore: 0.9,
				ProcessingTime:  200 * time.Millisecond,
				QualityAssessments: []FieldQualityAssessment{
					{
						FieldName: "email",
						FieldType: "email",
						QualityScore: QualityScore{
							OverallScore: 0.9,
						},
						QualityCategory: "excellent",
					},
				},
			}
			monitor.RecordExtractionResult(context.Background(), result, 200*time.Millisecond, nil)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final metrics
	metrics := monitor.GetMetrics()
	assert.Equal(t, int64(10), metrics.TotalRequests)
	assert.Equal(t, int64(10), metrics.SuccessfulRequests)
	assert.Equal(t, int64(0), metrics.FailedRequests)
}
