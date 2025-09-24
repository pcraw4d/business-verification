package observability

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DashboardConfigManager provides comprehensive dashboard configuration management
type DashboardConfigManager struct {
	logger        *Logger
	configs       map[string]*DashboardConfigData
	configHistory map[string][]*DashboardConfigData
	config        *DashboardConfigManagerConfig
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	started       bool
}

// DashboardConfigManagerConfig holds configuration for dashboard config manager
type DashboardConfigManagerConfig struct {
	Enabled               bool
	ConfigRetentionPeriod time.Duration
	MaxConfigHistory      int
	AutoSave              bool
	SaveInterval          time.Duration
	Environment           string
	ServiceName           string
	Version               string
}

// DashboardConfigData represents a dashboard configuration
type DashboardConfigData struct {
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
	Version     int                    `json:"version"`
	Enabled     bool                   `json:"enabled"`
	Owner       string                 `json:"owner"`
	Permissions map[string][]string    `json:"permissions"`
}

// DashboardType represents the type of dashboard
type DashboardType string

const (
	DashboardTypeOverview       DashboardType = "overview"
	DashboardTypePerformance    DashboardType = "performance"
	DashboardTypeBusiness       DashboardType = "business"
	DashboardTypeSecurity       DashboardType = "security"
	DashboardTypeInfrastructure DashboardType = "infrastructure"
	DashboardTypeHealth         DashboardType = "health"
	DashboardTypeCustom         DashboardType = "custom"
)

// PanelConfig represents a dashboard panel configuration
type PanelConfig struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Type        PanelType              `json:"type"`
	Position    *PositionConfig        `json:"position"`
	Size        *SizeConfig            `json:"size"`
	DataSource  string                 `json:"data_source"`
	Query       string                 `json:"query"`
	Options     map[string]interface{} `json:"options"`
	Thresholds  []*ThresholdConfig     `json:"thresholds"`
	Format      *FormatConfig          `json:"format"`
	RefreshRate time.Duration          `json:"refresh_rate"`
	Enabled     bool                   `json:"enabled"`
}

// PanelType represents the type of panel
type PanelType string

const (
	PanelTypeGraph     PanelType = "graph"
	PanelTypeStat      PanelType = "stat"
	PanelTypeGauge     PanelType = "gauge"
	PanelTypeTable     PanelType = "table"
	PanelTypeHeatmap   PanelType = "heatmap"
	PanelTypeLogs      PanelType = "logs"
	PanelTypeAlert     PanelType = "alert"
	PanelTypeText      PanelType = "text"
	PanelTypeMap       PanelType = "map"
	PanelTypePie       PanelType = "pie"
	PanelTypeBar       PanelType = "bar"
	PanelTypeHistogram PanelType = "histogram"
)

// PositionConfig represents panel position
type PositionConfig struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// SizeConfig represents panel size
type SizeConfig struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ThresholdConfig represents panel thresholds
type ThresholdConfig struct {
	Value     float64 `json:"value"`
	Color     string  `json:"color"`
	Condition string  `json:"condition"`
	Label     string  `json:"label"`
}

// FormatConfig represents data formatting options
type FormatConfig struct {
	Unit      string   `json:"unit"`
	Decimals  int      `json:"decimals"`
	MinValue  *float64 `json:"min_value,omitempty"`
	MaxValue  *float64 `json:"max_value,omitempty"`
	ColorMode string   `json:"color_mode"`
}

// FilterConfig represents dashboard filters
type FilterConfig struct {
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Field   string                 `json:"field"`
	Value   interface{}            `json:"value"`
	Options map[string]interface{} `json:"options"`
	Enabled bool                   `json:"enabled"`
}

// TimeRangeConfig represents time range configuration
type TimeRangeConfig struct {
	From string `json:"from"`
	To   string `json:"to"`
	Mode string `json:"mode"` // relative, absolute, now
}

// LayoutConfig represents dashboard layout
type LayoutConfig struct {
	Columns    int    `json:"columns"`
	Rows       int    `json:"rows"`
	GridSize   int    `json:"grid_size"`
	Theme      string `json:"theme"`
	Background string `json:"background"`
}

// NewDashboardConfigManager creates a new dashboard config manager
func NewDashboardConfigManager(
	logger *Logger,
	config *DashboardConfigManagerConfig,
) *DashboardConfigManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &DashboardConfigManager{
		logger:        logger,
		configs:       make(map[string]*DashboardConfigData),
		configHistory: make(map[string][]*DashboardConfigData),
		config:        config,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start starts the dashboard config manager
func (dcm *DashboardConfigManager) Start() error {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	if dcm.started {
		return fmt.Errorf("dashboard config manager already started")
	}

	dcm.logger.Info("Starting dashboard config manager", map[string]interface{}{
		"service_name": dcm.config.ServiceName,
		"version":      dcm.config.Version,
		"environment":  dcm.config.Environment,
	})

	// Initialize default configurations
	if err := dcm.initializeDefaultConfigs(); err != nil {
		return fmt.Errorf("failed to initialize default configs: %w", err)
	}

	// Start auto-save process
	if dcm.config.AutoSave {
		go dcm.startAutoSave()
	}

	dcm.started = true
	dcm.logger.Info("Dashboard config manager started successfully", map[string]interface{}{})
	return nil
}

// Stop stops the dashboard config manager
func (dcm *DashboardConfigManager) Stop() error {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	if !dcm.started {
		return fmt.Errorf("dashboard config manager not started")
	}

	dcm.logger.Info("Stopping dashboard config manager", map[string]interface{}{})

	dcm.cancel()
	dcm.started = false

	dcm.logger.Info("Dashboard config manager stopped successfully", map[string]interface{}{})
	return nil
}

// CreateDashboard creates a new dashboard configuration
func (dcm *DashboardConfigManager) CreateDashboard(config *DashboardConfigData) error {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	if config.ID == "" {
		return fmt.Errorf("dashboard ID cannot be empty")
	}

	if _, exists := dcm.configs[config.ID]; exists {
		return fmt.Errorf("dashboard with ID %s already exists", config.ID)
	}

	// Set default values
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	config.Version = 1
	config.Enabled = true

	// Validate configuration
	if err := dcm.validateDashboardConfig(config); err != nil {
		return fmt.Errorf("invalid dashboard configuration: %w", err)
	}

	dcm.configs[config.ID] = config

	// Add to history
	dcm.addToHistory(config)

	dcm.logger.Info("Dashboard created", map[string]interface{}{
		"dashboard_id": config.ID,
		"name":         config.Name,
		"type":         config.Type,
		"owner":        config.Owner,
	})

	return nil
}

// UpdateDashboard updates an existing dashboard configuration
func (dcm *DashboardConfigManager) UpdateDashboard(dashboardID string, updates *DashboardConfigData) error {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	existing, exists := dcm.configs[dashboardID]
	if !exists {
		return fmt.Errorf("dashboard with ID %s not found", dashboardID)
	}

	// Create updated configuration
	updated := &DashboardConfigData{
		ID:          existing.ID,
		Name:        updates.Name,
		Description: updates.Description,
		Type:        updates.Type,
		Panels:      updates.Panels,
		Layout:      updates.Layout,
		Filters:     updates.Filters,
		TimeRange:   updates.TimeRange,
		RefreshRate: updates.RefreshRate,
		Tags:        updates.Tags,
		Metadata:    updates.Metadata,
		CreatedAt:   existing.CreatedAt,
		UpdatedAt:   time.Now(),
		Version:     existing.Version + 1,
		Enabled:     updates.Enabled,
		Owner:       existing.Owner,
		Permissions: updates.Permissions,
	}

	// Validate updated configuration
	if err := dcm.validateDashboardConfig(updated); err != nil {
		return fmt.Errorf("invalid updated dashboard configuration: %w", err)
	}

	// Update configuration
	dcm.configs[dashboardID] = updated

	// Add to history
	dcm.addToHistory(updated)

	dcm.logger.Info("Dashboard updated", map[string]interface{}{
		"dashboard_id": dashboardID,
		"name":         updated.Name,
		"version":      updated.Version,
	})

	return nil
}

// DeleteDashboard deletes a dashboard configuration
func (dcm *DashboardConfigManager) DeleteDashboard(dashboardID string) error {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	if _, exists := dcm.configs[dashboardID]; !exists {
		return fmt.Errorf("dashboard with ID %s not found", dashboardID)
	}

	delete(dcm.configs, dashboardID)
	delete(dcm.configHistory, dashboardID)

	dcm.logger.Info("Dashboard deleted", map[string]interface{}{
		"dashboard_id": dashboardID,
	})

	return nil
}

// GetDashboard returns a dashboard configuration
func (dcm *DashboardConfigManager) GetDashboard(dashboardID string) (*DashboardConfigData, error) {
	dcm.mu.RLock()
	defer dcm.mu.RUnlock()

	config, exists := dcm.configs[dashboardID]
	if !exists {
		return nil, fmt.Errorf("dashboard with ID %s not found", dashboardID)
	}

	// Return a copy
	return dcm.copyDashboardConfig(config), nil
}

// ListDashboards returns all dashboard configurations
func (dcm *DashboardConfigManager) ListDashboards() []*DashboardConfigData {
	dcm.mu.RLock()
	defer dcm.mu.RUnlock()

	dashboards := make([]*DashboardConfigData, 0, len(dcm.configs))
	for _, config := range dcm.configs {
		dashboards = append(dashboards, dcm.copyDashboardConfig(config))
	}

	return dashboards
}

// GetDashboardHistory returns the configuration history for a dashboard
func (dcm *DashboardConfigManager) GetDashboardHistory(dashboardID string) ([]*DashboardConfigData, error) {
	dcm.mu.RLock()
	defer dcm.mu.RUnlock()

	history, exists := dcm.configHistory[dashboardID]
	if !exists {
		return nil, fmt.Errorf("dashboard with ID %s not found", dashboardID)
	}

	// Return copies
	historyCopy := make([]*DashboardConfigData, len(history))
	for i, config := range history {
		historyCopy[i] = dcm.copyDashboardConfig(config)
	}

	return historyCopy, nil
}

// RestoreDashboard restores a dashboard configuration from history
func (dcm *DashboardConfigManager) RestoreDashboard(dashboardID string, version int) error {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	history, exists := dcm.configHistory[dashboardID]
	if !exists {
		return fmt.Errorf("dashboard with ID %s not found", dashboardID)
	}

	// Find the specified version
	var targetConfig *DashboardConfigData
	for _, config := range history {
		if config.Version == version {
			targetConfig = config
			break
		}
	}

	if targetConfig == nil {
		return fmt.Errorf("version %d not found for dashboard %s", version, dashboardID)
	}

	// Create restored configuration
	restored := &DashboardConfigData{
		ID:          targetConfig.ID,
		Name:        targetConfig.Name,
		Description: targetConfig.Description,
		Type:        targetConfig.Type,
		Panels:      targetConfig.Panels,
		Layout:      targetConfig.Layout,
		Filters:     targetConfig.Filters,
		TimeRange:   targetConfig.TimeRange,
		RefreshRate: targetConfig.RefreshRate,
		Tags:        targetConfig.Tags,
		Metadata:    targetConfig.Metadata,
		CreatedAt:   targetConfig.CreatedAt,
		UpdatedAt:   time.Now(),
		Version:     targetConfig.Version + 1,
		Enabled:     targetConfig.Enabled,
		Owner:       targetConfig.Owner,
		Permissions: targetConfig.Permissions,
	}

	// Update configuration
	dcm.configs[dashboardID] = restored

	// Add to history
	dcm.addToHistory(restored)

	dcm.logger.Info("Dashboard restored", map[string]interface{}{
		"dashboard_id":     dashboardID,
		"restored_version": version,
		"new_version":      restored.Version,
	})

	return nil
}

// CloneDashboard clones an existing dashboard configuration
func (dcm *DashboardConfigManager) CloneDashboard(dashboardID string, newID string, newName string) error {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	existing, exists := dcm.configs[dashboardID]
	if !exists {
		return fmt.Errorf("dashboard with ID %s not found", dashboardID)
	}

	if _, exists := dcm.configs[newID]; exists {
		return fmt.Errorf("dashboard with ID %s already exists", newID)
	}

	// Create cloned configuration
	cloned := &DashboardConfigData{
		ID:          newID,
		Name:        newName,
		Description: existing.Description + " (Cloned)",
		Type:        existing.Type,
		Panels:      existing.Panels,
		Layout:      existing.Layout,
		Filters:     existing.Filters,
		TimeRange:   existing.TimeRange,
		RefreshRate: existing.RefreshRate,
		Tags:        existing.Tags,
		Metadata:    existing.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     1,
		Enabled:     true,
		Owner:       existing.Owner,
		Permissions: existing.Permissions,
	}

	// Validate cloned configuration
	if err := dcm.validateDashboardConfig(cloned); err != nil {
		return fmt.Errorf("invalid cloned dashboard configuration: %w", err)
	}

	dcm.configs[newID] = cloned

	// Add to history
	dcm.addToHistory(cloned)

	dcm.logger.Info("Dashboard cloned", map[string]interface{}{
		"original_id": dashboardID,
		"cloned_id":   newID,
		"name":        newName,
	})

	return nil
}

// validateDashboardConfig validates a dashboard configuration
func (dcm *DashboardConfigManager) validateDashboardConfig(config *DashboardConfigData) error {
	if config.Name == "" {
		return fmt.Errorf("dashboard name cannot be empty")
	}

	if config.Type == "" {
		return fmt.Errorf("dashboard type cannot be empty")
	}

	if config.Layout == nil {
		return fmt.Errorf("dashboard layout cannot be nil")
	}

	if config.Layout.Columns <= 0 || config.Layout.Rows <= 0 {
		return fmt.Errorf("invalid layout dimensions")
	}

	// Validate panels
	for _, panel := range config.Panels {
		if err := dcm.validatePanelConfig(panel); err != nil {
			return fmt.Errorf("invalid panel configuration: %w", err)
		}
	}

	return nil
}

// validatePanelConfig validates a panel configuration
func (dcm *DashboardConfigManager) validatePanelConfig(panel *PanelConfig) error {
	if panel.ID == "" {
		return fmt.Errorf("panel ID cannot be empty")
	}

	if panel.Title == "" {
		return fmt.Errorf("panel title cannot be empty")
	}

	if panel.Type == "" {
		return fmt.Errorf("panel type cannot be empty")
	}

	if panel.Position == nil {
		return fmt.Errorf("panel position cannot be nil")
	}

	if panel.Size == nil {
		return fmt.Errorf("panel size cannot be nil")
	}

	if panel.Size.Width <= 0 || panel.Size.Height <= 0 {
		return fmt.Errorf("invalid panel size")
	}

	return nil
}

// copyDashboardConfig creates a deep copy of a dashboard configuration
func (dcm *DashboardConfigManager) copyDashboardConfig(config *DashboardConfigData) *DashboardConfigData {
	// Create a deep copy of the configuration
	copied := &DashboardConfigData{
		ID:          config.ID,
		Name:        config.Name,
		Description: config.Description,
		Type:        config.Type,
		Panels:      make([]*PanelConfig, len(config.Panels)),
		Layout:      config.Layout,
		Filters:     make([]*FilterConfig, len(config.Filters)),
		TimeRange:   config.TimeRange,
		RefreshRate: config.RefreshRate,
		Tags:        make(map[string]string),
		Metadata:    make(map[string]interface{}),
		CreatedAt:   config.CreatedAt,
		UpdatedAt:   config.UpdatedAt,
		Version:     config.Version,
		Enabled:     config.Enabled,
		Owner:       config.Owner,
		Permissions: make(map[string][]string),
	}

	// Copy panels
	for i, panel := range config.Panels {
		copied.Panels[i] = &PanelConfig{
			ID:          panel.ID,
			Title:       panel.Title,
			Type:        panel.Type,
			Position:    panel.Position,
			Size:        panel.Size,
			DataSource:  panel.DataSource,
			Query:       panel.Query,
			Options:     panel.Options,
			Thresholds:  panel.Thresholds,
			Format:      panel.Format,
			RefreshRate: panel.RefreshRate,
			Enabled:     panel.Enabled,
		}
	}

	// Copy filters
	for i, filter := range config.Filters {
		copied.Filters[i] = &FilterConfig{
			Name:    filter.Name,
			Type:    filter.Type,
			Field:   filter.Field,
			Value:   filter.Value,
			Options: filter.Options,
			Enabled: filter.Enabled,
		}
	}

	// Copy tags
	for k, v := range config.Tags {
		copied.Tags[k] = v
	}

	// Copy metadata
	for k, v := range config.Metadata {
		copied.Metadata[k] = v
	}

	// Copy permissions
	for k, v := range config.Permissions {
		copied.Permissions[k] = make([]string, len(v))
		copy(copied.Permissions[k], v)
	}

	return copied
}

// addToHistory adds a configuration to history
func (dcm *DashboardConfigManager) addToHistory(config *DashboardConfigData) {
	history := dcm.configHistory[config.ID]
	history = append(history, dcm.copyDashboardConfig(config))

	// Limit history size
	if len(history) > dcm.config.MaxConfigHistory {
		history = history[len(history)-dcm.config.MaxConfigHistory:]
	}

	dcm.configHistory[config.ID] = history
}

// initializeDefaultConfigs initializes default dashboard configurations
func (dcm *DashboardConfigManager) initializeDefaultConfigs() error {
	// Overview Dashboard
	overviewDashboard := &DashboardConfigData{
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
			"environment": dcm.config.Environment,
			"service":     dcm.config.ServiceName,
		},
		Enabled: true,
		Owner:   "system",
	}

	// Performance Dashboard
	performanceDashboard := &DashboardConfigData{
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
			"environment": dcm.config.Environment,
			"service":     dcm.config.ServiceName,
		},
		Enabled: true,
		Owner:   "system",
	}

	// Add default configurations
	if err := dcm.CreateDashboard(overviewDashboard); err != nil {
		return fmt.Errorf("failed to create overview dashboard: %w", err)
	}

	if err := dcm.CreateDashboard(performanceDashboard); err != nil {
		return fmt.Errorf("failed to create performance dashboard: %w", err)
	}

	return nil
}

// startAutoSave starts the auto-save process
func (dcm *DashboardConfigManager) startAutoSave() {
	ticker := time.NewTicker(dcm.config.SaveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dcm.ctx.Done():
			dcm.logger.Info("Auto-save stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			dcm.performAutoSave()
		}
	}
}

// performAutoSave performs auto-save of configurations
func (dcm *DashboardConfigManager) performAutoSave() {
	dcm.logger.Debug("Auto-save performed", map[string]interface{}{
		"config_count": len(dcm.configs),
	})

	// In a real implementation, this would save configurations to persistent storage
}
