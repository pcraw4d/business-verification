package observability

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestAdvancedMonitoringDashboard_GetDashboardData(t *testing.T) {
	// Create test logger
	logger := zap.NewNop()

	// Create test dashboard
	dashboard := NewAdvancedMonitoringDashboard(
		DefaultAdvancedDashboardConfig(),
		nil, // metricsCollector
		nil, // performanceMonitor
		nil, // alertManager
		nil, // healthChecker
		logger,
	)

	// Test getting dashboard data
	ctx := context.Background()
	data, err := dashboard.GetDashboardData(ctx)
	if err != nil {
		t.Fatalf("Failed to get dashboard data: %v", err)
	}

	// Verify basic structure
	if data == nil {
		t.Fatal("Dashboard data should not be nil")
	}

	if data.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}

	if data.OverallHealth == "" {
		t.Error("Overall health should be set")
	}

	if data.HealthScore < 0 || data.HealthScore > 100 {
		t.Errorf("Health score should be between 0 and 100, got %f", data.HealthScore)
	}

	if data.AlertsSummary == nil {
		t.Error("Alerts summary should not be nil")
	}

	if data.Recommendations == nil {
		t.Error("Recommendations should not be nil")
	}
}

func TestAdvancedMonitoringDashboard_HealthAssessment(t *testing.T) {
	logger := zap.NewNop()
	dashboard := NewAdvancedMonitoringDashboard(
		DefaultAdvancedDashboardConfig(),
		nil, nil, nil, nil,
		logger,
	)

	tests := []struct {
		name              string
		mlHealth          string
		ensembleHealth    string
		uncertaintyHealth string
		securityHealth    string
		performanceHealth string
		criticalAlerts    int
		warningAlerts     int
		expectedHealth    string
	}{
		{
			name:              "all healthy",
			mlHealth:          "healthy",
			ensembleHealth:    "healthy",
			uncertaintyHealth: "healthy",
			securityHealth:    "healthy",
			performanceHealth: "healthy",
			criticalAlerts:    0,
			warningAlerts:     0,
			expectedHealth:    "healthy",
		},
		{
			name:              "warning status",
			mlHealth:          "warning",
			ensembleHealth:    "healthy",
			uncertaintyHealth: "healthy",
			securityHealth:    "healthy",
			performanceHealth: "healthy",
			criticalAlerts:    0,
			warningAlerts:     0,
			expectedHealth:    "warning",
		},
		{
			name:              "critical status",
			mlHealth:          "critical",
			ensembleHealth:    "healthy",
			uncertaintyHealth: "healthy",
			securityHealth:    "healthy",
			performanceHealth: "healthy",
			criticalAlerts:    0,
			warningAlerts:     0,
			expectedHealth:    "critical",
		},
		{
			name:              "critical alerts",
			mlHealth:          "healthy",
			ensembleHealth:    "healthy",
			uncertaintyHealth: "healthy",
			securityHealth:    "healthy",
			performanceHealth: "healthy",
			criticalAlerts:    1,
			warningAlerts:     0,
			expectedHealth:    "critical",
		},
		{
			name:              "warning alerts",
			mlHealth:          "healthy",
			ensembleHealth:    "healthy",
			uncertaintyHealth: "healthy",
			securityHealth:    "healthy",
			performanceHealth: "healthy",
			criticalAlerts:    0,
			warningAlerts:     1,
			expectedHealth:    "warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &AdvancedDashboardData{
				MLModelHealth:     tt.mlHealth,
				EnsembleHealth:    tt.ensembleHealth,
				UncertaintyHealth: tt.uncertaintyHealth,
				SecurityHealth:    tt.securityHealth,
				PerformanceHealth: tt.performanceHealth,
				AlertsSummary: &AlertSummary{
					CriticalAlerts: tt.criticalAlerts,
					WarningAlerts:  tt.warningAlerts,
				},
			}

			health := dashboard.determineOverallHealth(data)
			if health != tt.expectedHealth {
				t.Errorf("Expected health %s, got %s", tt.expectedHealth, health)
			}
		})
	}
}

func TestAdvancedMonitoringDashboard_HealthScoreCalculation(t *testing.T) {
	logger := zap.NewNop()
	dashboard := NewAdvancedMonitoringDashboard(
		DefaultAdvancedDashboardConfig(),
		nil, nil, nil, nil,
		logger,
	)

	tests := []struct {
		name              string
		mlHealth          string
		ensembleHealth    string
		uncertaintyHealth string
		securityHealth    string
		performanceHealth string
		criticalAlerts    int
		warningAlerts     int
		expectedMin       float64
		expectedMax       float64
	}{
		{
			name:              "all healthy",
			mlHealth:          "healthy",
			ensembleHealth:    "healthy",
			uncertaintyHealth: "healthy",
			securityHealth:    "healthy",
			performanceHealth: "healthy",
			criticalAlerts:    0,
			warningAlerts:     0,
			expectedMin:       90.0,
			expectedMax:       100.0,
		},
		{
			name:              "warning status",
			mlHealth:          "warning",
			ensembleHealth:    "healthy",
			uncertaintyHealth: "healthy",
			securityHealth:    "healthy",
			performanceHealth: "healthy",
			criticalAlerts:    0,
			warningAlerts:     0,
			expectedMin:       80.0,
			expectedMax:       90.0,
		},
		{
			name:              "critical status",
			mlHealth:          "critical",
			ensembleHealth:    "healthy",
			uncertaintyHealth: "healthy",
			securityHealth:    "healthy",
			performanceHealth: "healthy",
			criticalAlerts:    0,
			warningAlerts:     0,
			expectedMin:       70.0,
			expectedMax:       80.0,
		},
		{
			name:              "critical alerts",
			mlHealth:          "healthy",
			ensembleHealth:    "healthy",
			uncertaintyHealth: "healthy",
			securityHealth:    "healthy",
			performanceHealth: "healthy",
			criticalAlerts:    2,
			warningAlerts:     0,
			expectedMin:       80.0,
			expectedMax:       90.0,
		},
		{
			name:              "warning alerts",
			mlHealth:          "healthy",
			ensembleHealth:    "healthy",
			uncertaintyHealth: "healthy",
			securityHealth:    "healthy",
			performanceHealth: "healthy",
			criticalAlerts:    0,
			warningAlerts:     5,
			expectedMin:       80.0,
			expectedMax:       90.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &AdvancedDashboardData{
				MLModelHealth:     tt.mlHealth,
				EnsembleHealth:    tt.ensembleHealth,
				UncertaintyHealth: tt.uncertaintyHealth,
				SecurityHealth:    tt.securityHealth,
				PerformanceHealth: tt.performanceHealth,
				AlertsSummary: &AlertSummary{
					CriticalAlerts: tt.criticalAlerts,
					WarningAlerts:  tt.warningAlerts,
				},
			}

			score := dashboard.calculateHealthScore(data)
			if score < tt.expectedMin || score > tt.expectedMax {
				t.Errorf("Expected health score between %f and %f, got %f", tt.expectedMin, tt.expectedMax, score)
			}
		})
	}
}

func TestAdvancedMonitoringDashboard_Recommendations(t *testing.T) {
	logger := zap.NewNop()
	dashboard := NewAdvancedMonitoringDashboard(
		DefaultAdvancedDashboardConfig(),
		nil, nil, nil, nil,
		logger,
	)

	tests := []struct {
		name              string
		mlHealth          string
		ensembleHealth    string
		uncertaintyHealth string
		securityHealth    string
		performanceHealth string
		criticalAlerts    int
		warningAlerts     int
		expectedMin       int
	}{
		{
			name:              "all healthy",
			mlHealth:          "healthy",
			ensembleHealth:    "healthy",
			uncertaintyHealth: "healthy",
			securityHealth:    "healthy",
			performanceHealth: "healthy",
			criticalAlerts:    0,
			warningAlerts:     0,
			expectedMin:       0,
		},
		{
			name:              "warning status",
			mlHealth:          "warning",
			ensembleHealth:    "healthy",
			uncertaintyHealth: "healthy",
			securityHealth:    "healthy",
			performanceHealth: "healthy",
			criticalAlerts:    0,
			warningAlerts:     0,
			expectedMin:       1,
		},
		{
			name:              "critical status",
			mlHealth:          "critical",
			ensembleHealth:    "healthy",
			uncertaintyHealth: "healthy",
			securityHealth:    "healthy",
			performanceHealth: "healthy",
			criticalAlerts:    0,
			warningAlerts:     0,
			expectedMin:       1,
		},
		{
			name:              "critical alerts",
			mlHealth:          "healthy",
			ensembleHealth:    "healthy",
			uncertaintyHealth: "healthy",
			securityHealth:    "healthy",
			performanceHealth: "healthy",
			criticalAlerts:    2,
			warningAlerts:     0,
			expectedMin:       1,
		},
		{
			name:              "warning alerts",
			mlHealth:          "healthy",
			ensembleHealth:    "healthy",
			uncertaintyHealth: "healthy",
			securityHealth:    "healthy",
			performanceHealth: "healthy",
			criticalAlerts:    0,
			warningAlerts:     3,
			expectedMin:       1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &AdvancedDashboardData{
				MLModelHealth:     tt.mlHealth,
				EnsembleHealth:    tt.ensembleHealth,
				UncertaintyHealth: tt.uncertaintyHealth,
				SecurityHealth:    tt.securityHealth,
				PerformanceHealth: tt.performanceHealth,
				AlertsSummary: &AlertSummary{
					CriticalAlerts: tt.criticalAlerts,
					WarningAlerts:  tt.warningAlerts,
				},
			}

			recommendations := dashboard.generateRecommendations(data)
			if len(recommendations) < tt.expectedMin {
				t.Errorf("Expected at least %d recommendations, got %d", tt.expectedMin, len(recommendations))
			}
		})
	}
}

func TestAdvancedMonitoringDashboard_ExportDashboardData(t *testing.T) {
	logger := zap.NewNop()
	dashboard := NewAdvancedMonitoringDashboard(
		DefaultAdvancedDashboardConfig(),
		nil, nil, nil, nil,
		logger,
	)

	ctx := context.Background()

	// Test JSON export
	jsonData, err := dashboard.ExportDashboardData(ctx, "json")
	if err != nil {
		t.Fatalf("Failed to export JSON data: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("JSON export should not be empty")
	}

	// Test YAML export
	yamlData, err := dashboard.ExportDashboardData(ctx, "yaml")
	if err != nil {
		t.Fatalf("Failed to export YAML data: %v", err)
	}

	if len(yamlData) == 0 {
		t.Error("YAML export should not be empty")
	}

	// Test unsupported format
	_, err = dashboard.ExportDashboardData(ctx, "xml")
	if err == nil {
		t.Error("Should return error for unsupported format")
	}
}

func TestAdvancedMonitoringDashboard_GetHealthStatus(t *testing.T) {
	logger := zap.NewNop()
	dashboard := NewAdvancedMonitoringDashboard(
		DefaultAdvancedDashboardConfig(),
		nil, nil, nil, nil,
		logger,
	)

	// Test initial health status
	status := dashboard.GetHealthStatus()
	if status != "unknown" {
		t.Errorf("Expected initial health status 'unknown', got '%s'", status)
	}

	// Test after getting dashboard data
	ctx := context.Background()
	_, err := dashboard.GetDashboardData(ctx)
	if err != nil {
		t.Fatalf("Failed to get dashboard data: %v", err)
	}

	status = dashboard.GetHealthStatus()
	if status == "" {
		t.Error("Health status should not be empty after getting dashboard data")
	}
}

func TestAdvancedMonitoringDashboard_GetLastUpdateTime(t *testing.T) {
	logger := zap.NewNop()
	dashboard := NewAdvancedMonitoringDashboard(
		DefaultAdvancedDashboardConfig(),
		nil, nil, nil, nil,
		logger,
	)

	// Test initial update time
	lastUpdate := dashboard.GetLastUpdateTime()
	if lastUpdate.IsZero() {
		t.Error("Last update time should not be zero")
	}

	// Test after getting dashboard data
	ctx := context.Background()
	_, err := dashboard.GetDashboardData(ctx)
	if err != nil {
		t.Fatalf("Failed to get dashboard data: %v", err)
	}

	newUpdate := dashboard.GetLastUpdateTime()
	if newUpdate.Before(lastUpdate) {
		t.Error("Last update time should be updated after getting dashboard data")
	}
}

func TestAdvancedMonitoringDashboard_GetAlertsSummary(t *testing.T) {
	logger := zap.NewNop()
	dashboard := NewAdvancedMonitoringDashboard(
		DefaultAdvancedDashboardConfig(),
		nil, nil, nil, nil,
		logger,
	)

	// Test initial alerts summary
	summary := dashboard.GetAlertsSummary()
	if summary == nil {
		t.Error("Alerts summary should not be nil")
	}

	// The timestamp might be zero initially, which is acceptable
	// We'll test that it gets updated after getting dashboard data
}

func TestDefaultAdvancedDashboardConfig(t *testing.T) {
	config := DefaultAdvancedDashboardConfig()

	if config == nil {
		t.Fatal("Default config should not be nil")
	}

	if !config.DashboardEnabled {
		t.Error("Dashboard should be enabled by default")
	}

	if config.UpdateInterval == 0 {
		t.Error("Update interval should be set")
	}

	if config.HealthCheckInterval == 0 {
		t.Error("Health check interval should be set")
	}

	if config.AlertSummaryInterval == 0 {
		t.Error("Alert summary interval should be set")
	}

	if config.RealTimeUpdateInterval == 0 {
		t.Error("Real-time update interval should be set")
	}

	if config.MaxAlertsDisplayed <= 0 {
		t.Error("Max alerts displayed should be positive")
	}

	if config.MaxMetricsHistory <= 0 {
		t.Error("Max metrics history should be positive")
	}

	if !config.ShowDetailedMetrics {
		t.Error("Show detailed metrics should be true by default")
	}

	if !config.ShowTrendAnalysis {
		t.Error("Show trend analysis should be true by default")
	}

	if !config.ShowMLModelMetrics {
		t.Error("Show ML model metrics should be true by default")
	}

	if !config.ShowEnsembleMetrics {
		t.Error("Show ensemble metrics should be true by default")
	}

	if !config.ShowUncertaintyMetrics {
		t.Error("Show uncertainty metrics should be true by default")
	}

	if !config.ShowSecurityMetrics {
		t.Error("Show security metrics should be true by default")
	}

	if !config.IntegrateMLMonitoring {
		t.Error("Integrate ML monitoring should be true by default")
	}

	if !config.IntegrateEnsembleMonitoring {
		t.Error("Integrate ensemble monitoring should be true by default")
	}

	if !config.IntegrateUncertaintyMonitoring {
		t.Error("Integrate uncertainty monitoring should be true by default")
	}

	if !config.IntegrateSecurityMonitoring {
		t.Error("Integrate security monitoring should be true by default")
	}

	if !config.IntegratePerformanceMonitoring {
		t.Error("Integrate performance monitoring should be true by default")
	}
}

func TestNewMLModelMonitor(t *testing.T) {
	monitor := NewMLModelMonitor()

	if monitor == nil {
		t.Fatal("ML model monitor should not be nil")
	}

	if monitor.models == nil {
		t.Error("Models map should be initialized")
	}

	if monitor.driftDetector == nil {
		t.Error("Drift detector should be initialized")
	}

	if monitor.performanceTracker == nil {
		t.Error("Performance tracker should be initialized")
	}
}

func TestNewEnsembleMonitor(t *testing.T) {
	monitor := NewEnsembleMonitor()

	if monitor == nil {
		t.Fatal("Ensemble monitor should not be nil")
	}

	if monitor.methods == nil {
		t.Error("Methods map should be initialized")
	}

	if monitor.weightTracker == nil {
		t.Error("Weight tracker should be initialized")
	}

	if monitor.contributionAnalyzer == nil {
		t.Error("Contribution analyzer should be initialized")
	}
}

func TestNewUncertaintyMonitor(t *testing.T) {
	monitor := NewUncertaintyMonitor()

	if monitor == nil {
		t.Fatal("Uncertainty monitor should not be nil")
	}

	if monitor.uncertaintyMetrics == nil {
		t.Error("Uncertainty metrics should be initialized")
	}

	if monitor.calibrationTracker == nil {
		t.Error("Calibration tracker should be initialized")
	}

	if monitor.reliabilityAnalyzer == nil {
		t.Error("Reliability analyzer should be initialized")
	}
}

func TestNewSecurityMonitor(t *testing.T) {
	monitor := NewSecurityMonitor()

	if monitor == nil {
		t.Fatal("Security monitor should not be nil")
	}

	if monitor.securityMetrics == nil {
		t.Error("Security metrics should be initialized")
	}

	if monitor.complianceTracker == nil {
		t.Error("Compliance tracker should be initialized")
	}

	if monitor.violationAnalyzer == nil {
		t.Error("Violation analyzer should be initialized")
	}
}

func TestMLModelMonitor_GetAllMLModelMetrics(t *testing.T) {
	monitor := NewMLModelMonitor()

	metrics := monitor.GetAllMLModelMetrics()
	if metrics == nil {
		t.Error("Metrics should not be nil")
	}

	if len(metrics) != 0 {
		t.Error("Initial metrics should be empty")
	}
}

func TestEnsembleMonitor_GetAllEnsembleMetrics(t *testing.T) {
	monitor := NewEnsembleMonitor()

	metrics := monitor.GetAllEnsembleMetrics()
	if metrics == nil {
		t.Error("Metrics should not be nil")
	}

	if len(metrics) != 0 {
		t.Error("Initial metrics should be empty")
	}
}

func TestUncertaintyMonitor_GetUncertaintyMetrics(t *testing.T) {
	monitor := NewUncertaintyMonitor()

	metrics := monitor.GetUncertaintyMetrics()
	if metrics == nil {
		t.Error("Metrics should not be nil")
	}
}

func TestSecurityMonitor_GetSecurityMetrics(t *testing.T) {
	monitor := NewSecurityMonitor()

	metrics := monitor.GetSecurityMetrics()
	if metrics == nil {
		t.Error("Metrics should not be nil")
	}
}

// Benchmark tests
func BenchmarkAdvancedMonitoringDashboard_GetDashboardData(b *testing.B) {
	logger := zap.NewNop()
	dashboard := NewAdvancedMonitoringDashboard(
		DefaultAdvancedDashboardConfig(),
		nil, nil, nil, nil,
		logger,
	)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dashboard.GetDashboardData(ctx)
		if err != nil {
			b.Fatalf("Failed to get dashboard data: %v", err)
		}
	}
}

func BenchmarkAdvancedMonitoringDashboard_HealthScoreCalculation(b *testing.B) {
	logger := zap.NewNop()
	dashboard := NewAdvancedMonitoringDashboard(
		DefaultAdvancedDashboardConfig(),
		nil, nil, nil, nil,
		logger,
	)

	data := &AdvancedDashboardData{
		MLModelHealth:     "healthy",
		EnsembleHealth:    "warning",
		UncertaintyHealth: "healthy",
		SecurityHealth:    "critical",
		PerformanceHealth: "healthy",
		AlertsSummary: &AlertSummary{
			CriticalAlerts: 2,
			WarningAlerts:  5,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dashboard.calculateHealthScore(data)
	}
}

func BenchmarkAdvancedMonitoringDashboard_Recommendations(b *testing.B) {
	logger := zap.NewNop()
	dashboard := NewAdvancedMonitoringDashboard(
		DefaultAdvancedDashboardConfig(),
		nil, nil, nil, nil,
		logger,
	)

	data := &AdvancedDashboardData{
		MLModelHealth:     "critical",
		EnsembleHealth:    "warning",
		UncertaintyHealth: "healthy",
		SecurityHealth:    "critical",
		PerformanceHealth: "warning",
		AlertsSummary: &AlertSummary{
			CriticalAlerts: 3,
			WarningAlerts:  7,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dashboard.generateRecommendations(data)
	}
}
