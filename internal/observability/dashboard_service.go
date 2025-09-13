package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// DashboardService provides comprehensive monitoring dashboard functionality
type DashboardService struct {
	logger           *Logger
	metricsCollector *MetricsCollector
	healthChecker    *HealthChecker
	alertManager     *AlertManager
	dashboardConfigs map[string]*DashboardConfig
	dashboardData    map[string]*DashboardData
	exporters        []DashboardDataExporter
	config           *DashboardServiceConfig
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	started          bool
}

// DashboardServiceConfig holds configuration for dashboard service
type DashboardServiceConfig struct {
	Enabled               bool
	RefreshInterval       time.Duration
	DataRetentionPeriod   time.Duration
	MaxDataPoints         int
	ExportEnabled         bool
	ExportInterval        time.Duration
	CustomDashboards      bool
	RealTimeUpdates       bool
	HistoricalDataEnabled bool
	Environment           string
	ServiceName           string
	Version               string
}

// DashboardConfig represents a dashboard configuration
type DashboardConfig struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        DashboardType          `json:"type"`
	Panels      []*PanelConfig         `json:"panels"`
	Layout      *LayoutConfig          `json:"layout"`
	Filters     []*FilterConfig        `json:"filters"`
	TimeRange   *TimeRangeConfig       `json:"time_range"`
	RefreshRate time.Duration          `json:"refresh_rate"`
	Tags        map[string]string      `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Enabled     bool                   `json:"enabled"`
}

// DashboardData represents dashboard data
type DashboardData struct {
	DashboardID string                 `json:"dashboard_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
	Status      string                 `json:"status"`
	Error       string                 `json:"error,omitempty"`
}

// DashboardDataExporter interface for exporting dashboard data
type DashboardDataExporter interface {
	Export(dashboard *DashboardData) error
	Name() string
	Type() string
}

// JSONDashboardDataExporter exports dashboard data as JSON
type JSONDashboardDataExporter struct {
	logger *Logger
}

// NewJSONDashboardDataExporter creates a new JSON dashboard data exporter
func NewJSONDashboardDataExporter(logger *Logger) *JSONDashboardDataExporter {
	return &JSONDashboardDataExporter{
		logger: logger,
	}
}

// Export exports dashboard data as JSON
func (jde *JSONDashboardDataExporter) Export(dashboard *DashboardData) error {
	data, err := json.MarshalIndent(dashboard, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal dashboard data: %w", err)
	}

	jde.logger.Debug("Dashboard data exported as JSON", map[string]interface{}{
		"dashboard_id": dashboard.DashboardID,
		"data_size":    len(data),
	})

	return nil
}

// Name returns the exporter name
func (jde *JSONDashboardDataExporter) Name() string {
	return "json"
}

// Type returns the exporter type
func (jde *JSONDashboardDataExporter) Type() string {
	return "json"
}

// PrometheusDashboardExporter exports dashboard data to Prometheus
type PrometheusDashboardExporter struct {
	logger *Logger
}

// NewPrometheusDashboardExporter creates a new Prometheus dashboard exporter
func NewPrometheusDashboardExporter(logger *Logger) *PrometheusDashboardExporter {
	return &PrometheusDashboardExporter{
		logger: logger,
	}
}

// Export exports dashboard data to Prometheus
func (pde *PrometheusDashboardExporter) Export(dashboard *DashboardData) error {
	pde.logger.Debug("Dashboard data exported to Prometheus", map[string]interface{}{
		"dashboard_id": dashboard.DashboardID,
		"data_points":  len(dashboard.Data),
	})

	// In a real implementation, this would export metrics to Prometheus
	return nil
}

// Name returns the exporter name
func (pde *PrometheusDashboardExporter) Name() string {
	return "prometheus"
}

// Type returns the exporter type
func (pde *PrometheusDashboardExporter) Type() string {
	return "prometheus"
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(
	logger *Logger,
	metricsCollector *MetricsCollector,
	healthChecker *HealthChecker,
	alertManager *AlertManager,
	config *DashboardServiceConfig,
) *DashboardService {
	ctx, cancel := context.WithCancel(context.Background())

	return &DashboardService{
		logger:           logger,
		metricsCollector: metricsCollector,
		healthChecker:    healthChecker,
		alertManager:     alertManager,
		dashboardConfigs: make(map[string]*DashboardConfig),
		dashboardData:    make(map[string]*DashboardData),
		exporters:        make([]DashboardDataExporter, 0),
		config:           config,
		ctx:              ctx,
		cancel:           cancel,
	}
}

// Start starts the dashboard service
func (ds *DashboardService) Start() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.started {
		return fmt.Errorf("dashboard service already started")
	}

	ds.logger.Info("Starting dashboard service", map[string]interface{}{
		"service_name": ds.config.ServiceName,
		"version":      ds.config.Version,
		"environment":  ds.config.Environment,
	})

	// Initialize default dashboards
	if err := ds.initializeDefaultDashboards(); err != nil {
		return fmt.Errorf("failed to initialize default dashboards: %w", err)
	}

	// Start data collection
	if ds.config.Enabled {
		go ds.startDataCollection()
	}

	// Start data export
	if ds.config.ExportEnabled {
		go ds.startDataExport()
	}

	ds.started = true
	ds.logger.Info("Dashboard service started successfully", map[string]interface{}{})
	return nil
}

// Stop stops the dashboard service
func (ds *DashboardService) Stop() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if !ds.started {
		return fmt.Errorf("dashboard service not started")
	}

	ds.logger.Info("Stopping dashboard service", map[string]interface{}{})

	ds.cancel()
	ds.started = false

	ds.logger.Info("Dashboard service stopped successfully", map[string]interface{}{})
	return nil
}

// AddDashboard adds a new dashboard configuration
func (ds *DashboardService) AddDashboard(config *DashboardConfig) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if config.ID == "" {
		return fmt.Errorf("dashboard ID cannot be empty")
	}

	if _, exists := ds.dashboardConfigs[config.ID]; exists {
		return fmt.Errorf("dashboard with ID %s already exists", config.ID)
	}

	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	ds.dashboardConfigs[config.ID] = config

	ds.logger.Info("Dashboard added", map[string]interface{}{
		"dashboard_id": config.ID,
		"name":         config.Name,
		"type":         config.Type,
	})

	return nil
}

// RemoveDashboard removes a dashboard configuration
func (ds *DashboardService) RemoveDashboard(dashboardID string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if _, exists := ds.dashboardConfigs[dashboardID]; !exists {
		return fmt.Errorf("dashboard with ID %s not found", dashboardID)
	}

	delete(ds.dashboardConfigs, dashboardID)
	delete(ds.dashboardData, dashboardID)

	ds.logger.Info("Dashboard removed", map[string]interface{}{
		"dashboard_id": dashboardID,
	})

	return nil
}

// GetDashboard returns a dashboard configuration
func (ds *DashboardService) GetDashboard(dashboardID string) (*DashboardConfig, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	config, exists := ds.dashboardConfigs[dashboardID]
	if !exists {
		return nil, fmt.Errorf("dashboard with ID %s not found", dashboardID)
	}

	// Return a copy
	return &DashboardConfig{
		ID:          config.ID,
		Name:        config.Name,
		Description: config.Description,
		Type:        config.Type,
		Panels:      config.Panels,
		Layout:      config.Layout,
		Filters:     config.Filters,
		TimeRange:   config.TimeRange,
		RefreshRate: config.RefreshRate,
		Tags:        config.Tags,
		Metadata:    config.Metadata,
		CreatedAt:   config.CreatedAt,
		UpdatedAt:   config.UpdatedAt,
		Enabled:     config.Enabled,
	}, nil
}

// ListDashboards returns all dashboard configurations
func (ds *DashboardService) ListDashboards() []*DashboardConfig {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	dashboards := make([]*DashboardConfig, 0, len(ds.dashboardConfigs))
	for _, config := range ds.dashboardConfigs {
		dashboards = append(dashboards, &DashboardConfig{
			ID:          config.ID,
			Name:        config.Name,
			Description: config.Description,
			Type:        config.Type,
			Panels:      config.Panels,
			Layout:      config.Layout,
			Filters:     config.Filters,
			TimeRange:   config.TimeRange,
			RefreshRate: config.RefreshRate,
			Tags:        config.Tags,
			Metadata:    config.Metadata,
			CreatedAt:   config.CreatedAt,
			UpdatedAt:   config.UpdatedAt,
			Enabled:     config.Enabled,
		})
	}

	return dashboards
}

// GetDashboardData returns dashboard data
func (ds *DashboardService) GetDashboardData(dashboardID string) (*DashboardData, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	data, exists := ds.dashboardData[dashboardID]
	if !exists {
		return nil, fmt.Errorf("dashboard data for ID %s not found", dashboardID)
	}

	// Return a copy
	return &DashboardData{
		DashboardID: data.DashboardID,
		Timestamp:   data.Timestamp,
		Data:        data.Data,
		Metadata:    data.Metadata,
		Status:      data.Status,
		Error:       data.Error,
	}, nil
}

// RefreshDashboard refreshes dashboard data
func (ds *DashboardService) RefreshDashboard(dashboardID string) error {
	ds.mu.RLock()
	config, exists := ds.dashboardConfigs[dashboardID]
	ds.mu.RUnlock()

	if !exists {
		return fmt.Errorf("dashboard with ID %s not found", dashboardID)
	}

	if !config.Enabled {
		return fmt.Errorf("dashboard %s is disabled", dashboardID)
	}

	// Collect data for the dashboard
	data, err := ds.collectDashboardData(config)
	if err != nil {
		ds.logger.Error("Failed to collect dashboard data", map[string]interface{}{
			"dashboard_id": dashboardID,
			"error":        err.Error(),
		})
		return fmt.Errorf("failed to collect dashboard data: %w", err)
	}

	// Store the data
	ds.mu.Lock()
	ds.dashboardData[dashboardID] = data
	ds.mu.Unlock()

	ds.logger.Debug("Dashboard data refreshed", map[string]interface{}{
		"dashboard_id": dashboardID,
		"data_points":  len(data.Data),
	})

	return nil
}

// AddExporter adds a dashboard exporter
func (ds *DashboardService) AddExporter(exporter DashboardDataExporter) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.exporters = append(ds.exporters, exporter)

	ds.logger.Info("Dashboard exporter added", map[string]interface{}{
		"exporter": exporter.Name(),
		"type":     exporter.Type(),
	})
}

// initializeDefaultDashboards initializes default dashboard configurations
func (ds *DashboardService) initializeDefaultDashboards() error {
	// Overview Dashboard
	overviewDashboard := &DashboardConfig{
		ID:          "overview",
		Name:        "KYB Platform Overview",
		Description: "High-level overview of KYB Platform health and performance",
		Type:        DashboardTypeOverview,
		Panels: []*PanelConfig{
			{
				ID:         "system_health",
				Title:      "System Health",
				Type:       PanelTypeStat,
				Position:   &PositionConfig{X: 0, Y: 0},
				Size:       &SizeConfig{Width: 6, Height: 4},
				DataSource: "health",
				Query:      "overall_status",
				Enabled:    true,
			},
			{
				ID:         "request_rate",
				Title:      "Request Rate",
				Type:       PanelTypeGraph,
				Position:   &PositionConfig{X: 6, Y: 0},
				Size:       &SizeConfig{Width: 6, Height: 4},
				DataSource: "metrics",
				Query:      "rate(kyb_http_requests_total[5m])",
				Enabled:    true,
			},
			{
				ID:         "response_time",
				Title:      "Response Time",
				Type:       PanelTypeGraph,
				Position:   &PositionConfig{X: 12, Y: 0},
				Size:       &SizeConfig{Width: 6, Height: 4},
				DataSource: "metrics",
				Query:      "histogram_quantile(0.95, rate(kyb_http_request_duration_seconds_bucket[5m]))",
				Enabled:    true,
			},
			{
				ID:         "error_rate",
				Title:      "Error Rate",
				Type:       PanelTypeGraph,
				Position:   &PositionConfig{X: 18, Y: 0},
				Size:       &SizeConfig{Width: 6, Height: 4},
				DataSource: "metrics",
				Query:      "rate(kyb_http_requests_total{status=~\"5..\"}[5m]) / rate(kyb_http_requests_total[5m]) * 100",
				Enabled:    true,
			},
		},
		Layout: &LayoutConfig{
			Columns:  24,
			Rows:     12,
			GridSize: 1,
			Theme:    "dark",
		},
		TimeRange: &TimeRangeConfig{
			From: "now-1h",
			To:   "now",
			Mode: "relative",
		},
		RefreshRate: 30 * time.Second,
		Tags: map[string]string{
			"environment": ds.config.Environment,
			"service":     ds.config.ServiceName,
		},
		Enabled: true,
	}

	// Performance Dashboard
	performanceDashboard := &DashboardConfig{
		ID:          "performance",
		Name:        "Performance Dashboard",
		Description: "Detailed performance metrics and system resources",
		Type:        DashboardTypePerformance,
		Panels: []*PanelConfig{
			{
				ID:         "cpu_usage",
				Title:      "CPU Usage",
				Type:       PanelTypeGauge,
				Position:   &PositionConfig{X: 0, Y: 0},
				Size:       &SizeConfig{Width: 6, Height: 6},
				DataSource: "metrics",
				Query:      "kyb_system_cpu_usage",
				Thresholds: []*ThresholdConfig{
					{Value: 70, Color: "yellow", Condition: "gt", Label: "Warning"},
					{Value: 85, Color: "red", Condition: "gt", Label: "Critical"},
				},
				Enabled: true,
			},
			{
				ID:         "memory_usage",
				Title:      "Memory Usage",
				Type:       PanelTypeGauge,
				Position:   &PositionConfig{X: 6, Y: 0},
				Size:       &SizeConfig{Width: 6, Height: 6},
				DataSource: "metrics",
				Query:      "kyb_system_memory_usage",
				Thresholds: []*ThresholdConfig{
					{Value: 70, Color: "yellow", Condition: "gt", Label: "Warning"},
					{Value: 85, Color: "red", Condition: "gt", Label: "Critical"},
				},
				Enabled: true,
			},
			{
				ID:         "goroutines",
				Title:      "Goroutines",
				Type:       PanelTypeGauge,
				Position:   &PositionConfig{X: 12, Y: 0},
				Size:       &SizeConfig{Width: 6, Height: 6},
				DataSource: "metrics",
				Query:      "kyb_system_goroutines",
				Thresholds: []*ThresholdConfig{
					{Value: 500, Color: "yellow", Condition: "gt", Label: "Warning"},
					{Value: 1000, Color: "red", Condition: "gt", Label: "Critical"},
				},
				Enabled: true,
			},
			{
				ID:         "heap_memory",
				Title:      "Heap Memory",
				Type:       PanelTypeGauge,
				Position:   &PositionConfig{X: 18, Y: 0},
				Size:       &SizeConfig{Width: 6, Height: 6},
				DataSource: "metrics",
				Query:      "kyb_system_heap_alloc / 1024 / 1024",
				Format:     &FormatConfig{Unit: "MB"},
				Enabled:    true,
			},
		},
		Layout: &LayoutConfig{
			Columns:  24,
			Rows:     12,
			GridSize: 1,
			Theme:    "dark",
		},
		TimeRange: &TimeRangeConfig{
			From: "now-1h",
			To:   "now",
			Mode: "relative",
		},
		RefreshRate: 15 * time.Second,
		Tags: map[string]string{
			"environment": ds.config.Environment,
			"service":     ds.config.ServiceName,
		},
		Enabled: true,
	}

	// Add dashboards
	if err := ds.AddDashboard(overviewDashboard); err != nil {
		return fmt.Errorf("failed to add overview dashboard: %w", err)
	}

	if err := ds.AddDashboard(performanceDashboard); err != nil {
		return fmt.Errorf("failed to add performance dashboard: %w", err)
	}

	return nil
}

// collectDashboardData collects data for a dashboard
func (ds *DashboardService) collectDashboardData(config *DashboardConfig) (*DashboardData, error) {
	data := &DashboardData{
		DashboardID: config.ID,
		Timestamp:   time.Now(),
		Data:        make(map[string]interface{}),
		Metadata: map[string]interface{}{
			"dashboard_name": config.Name,
			"dashboard_type": config.Type,
			"panel_count":    len(config.Panels),
		},
		Status: "success",
	}

	// Collect data for each panel
	for _, panel := range config.Panels {
		if !panel.Enabled {
			continue
		}

		panelData, err := ds.collectPanelData(panel)
		if err != nil {
			ds.logger.Warn("Failed to collect panel data", map[string]interface{}{
				"dashboard_id": config.ID,
				"panel_id":     panel.ID,
				"error":        err.Error(),
			})
			panelData = map[string]interface{}{
				"error": err.Error(),
			}
		}

		data.Data[panel.ID] = panelData
	}

	return data, nil
}

// collectPanelData collects data for a specific panel
func (ds *DashboardService) collectPanelData(panel *PanelConfig) (map[string]interface{}, error) {
	switch panel.DataSource {
	case "metrics":
		return ds.collectMetricsData(panel)
	case "health":
		return ds.collectHealthData(panel)
	case "alerts":
		return ds.collectAlertsData(panel)
	default:
		return nil, fmt.Errorf("unknown data source: %s", panel.DataSource)
	}
}

// collectMetricsData collects metrics data for a panel
func (ds *DashboardService) collectMetricsData(panel *PanelConfig) (map[string]interface{}, error) {
	// In a real implementation, this would query the metrics collector
	// For now, return mock data
	return map[string]interface{}{
		"value":     100.0,
		"timestamp": time.Now(),
		"query":     panel.Query,
		"source":    "metrics",
	}, nil
}

// collectHealthData collects health data for a panel
func (ds *DashboardService) collectHealthData(panel *PanelConfig) (map[string]interface{}, error) {
	if ds.healthChecker == nil {
		return nil, fmt.Errorf("health checker not available")
	}

	status := ds.healthChecker.GetStatus()
	return map[string]interface{}{
		"status":    status["overall_status"],
		"checks":    status["checks"],
		"timestamp": time.Now(),
		"source":    "health",
	}, nil
}

// collectAlertsData collects alerts data for a panel
func (ds *DashboardService) collectAlertsData(panel *PanelConfig) (map[string]interface{}, error) {
	if ds.alertManager == nil {
		return nil, fmt.Errorf("alert manager not available")
	}

	alerts := ds.alertManager.GetActiveAlerts()
	return map[string]interface{}{
		"alerts":    alerts,
		"count":     len(alerts),
		"timestamp": time.Now(),
		"source":    "alerts",
	}, nil
}

// startDataCollection starts the data collection process
func (ds *DashboardService) startDataCollection() {
	ticker := time.NewTicker(ds.config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ds.ctx.Done():
			ds.logger.Info("Data collection stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			ds.refreshAllDashboards()
		}
	}
}

// refreshAllDashboards refreshes all enabled dashboards
func (ds *DashboardService) refreshAllDashboards() {
	ds.mu.RLock()
	dashboards := make([]*DashboardConfig, 0, len(ds.dashboardConfigs))
	for _, config := range ds.dashboardConfigs {
		if config.Enabled {
			dashboards = append(dashboards, config)
		}
	}
	ds.mu.RUnlock()

	for _, config := range dashboards {
		if err := ds.RefreshDashboard(config.ID); err != nil {
			ds.logger.Error("Failed to refresh dashboard", map[string]interface{}{
				"dashboard_id": config.ID,
				"error":        err.Error(),
			})
		}
	}
}

// startDataExport starts the data export process
func (ds *DashboardService) startDataExport() {
	ticker := time.NewTicker(ds.config.ExportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ds.ctx.Done():
			ds.logger.Info("Data export stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			ds.exportAllDashboardData()
		}
	}
}

// exportAllDashboardData exports all dashboard data
func (ds *DashboardService) exportAllDashboardData() {
	ds.mu.RLock()
	data := make([]*DashboardData, 0, len(ds.dashboardData))
	for _, dashboardData := range ds.dashboardData {
		data = append(data, dashboardData)
	}
	ds.mu.RUnlock()

	for _, dashboardData := range data {
		for _, exporter := range ds.exporters {
			if err := exporter.Export(dashboardData); err != nil {
				ds.logger.Error("Failed to export dashboard data", map[string]interface{}{
					"dashboard_id": dashboardData.DashboardID,
					"exporter":     exporter.Name(),
					"error":        err.Error(),
				})
			}
		}
	}
}
