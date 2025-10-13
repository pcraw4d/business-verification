package businessintelligence

import (
	"time"
)

// DataSourceType represents the type of data source
type DataSourceType string

const (
	DataSourceTypeRiskAssessment DataSourceType = "risk_assessment"
	DataSourceTypeBatchJob       DataSourceType = "batch_job"
	DataSourceTypeReport         DataSourceType = "report"
	DataSourceTypeDashboard      DataSourceType = "dashboard"
	DataSourceTypeCustomModel    DataSourceType = "custom_model"
	DataSourceTypeWebhook        DataSourceType = "webhook"
	DataSourceTypePerformance    DataSourceType = "performance"
)

// DataExportFormat represents the format for data export
type DataExportFormat string

const (
	DataExportFormatJSON    DataExportFormat = "json"
	DataExportFormatCSV     DataExportFormat = "csv"
	DataExportFormatExcel   DataExportFormat = "excel"
	DataExportFormatParquet DataExportFormat = "parquet"
	DataExportFormatAvro    DataExportFormat = "avro"
)

// DataSyncStatus represents the status of data synchronization
type DataSyncStatus string

const (
	DataSyncStatusPending   DataSyncStatus = "pending"
	DataSyncStatusSyncing   DataSyncStatus = "syncing"
	DataSyncStatusCompleted DataSyncStatus = "completed"
	DataSyncStatusFailed    DataSyncStatus = "failed"
	DataSyncStatusPaused    DataSyncStatus = "paused"
)

// DataSync represents a data synchronization job
type DataSync struct {
	ID                string                 `json:"id" db:"id"`
	TenantID          string                 `json:"tenant_id" db:"tenant_id"`
	Name              string                 `json:"name" db:"name"`
	DataSourceType    DataSourceType         `json:"data_source_type" db:"data_source_type"`
	SourceConfig      DataSourceConfig       `json:"source_config" db:"source_config"`
	DestinationConfig DestinationConfig      `json:"destination_config" db:"destination_config"`
	SyncSchedule      SyncSchedule           `json:"sync_schedule" db:"sync_schedule"`
	Status            DataSyncStatus         `json:"status" db:"status"`
	LastSyncAt        *time.Time             `json:"last_sync_at" db:"last_sync_at"`
	NextSyncAt        *time.Time             `json:"next_sync_at" db:"next_sync_at"`
	RecordsSynced     int64                  `json:"records_synced" db:"records_synced"`
	RecordsFailed     int64                  `json:"records_failed" db:"records_failed"`
	Error             string                 `json:"error" db:"error"`
	CreatedBy         string                 `json:"created_by" db:"created_by"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
}

// DataSourceConfig represents the configuration for a data source
type DataSourceConfig struct {
	Filters         map[string]interface{} `json:"filters"`
	DateRange       DateRangeConfig        `json:"date_range"`
	Fields          []string               `json:"fields"`
	Transformations []DataTransformation   `json:"transformations"`
	Aggregations    []DataAggregation      `json:"aggregations"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// DateRangeConfig represents date range configuration
type DateRangeConfig struct {
	StartDate        *time.Time `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	Period           string     `json:"period"` // "7d", "30d", "90d", "1y", "custom"
	Incremental      bool       `json:"incremental"`
	IncrementalField string     `json:"incremental_field"`
}

// DataTransformation represents a data transformation
type DataTransformation struct {
	Type        string                 `json:"type"` // "filter", "map", "aggregate", "calculate"
	Name        string                 `json:"name"`
	Config      map[string]interface{} `json:"config"`
	Description string                 `json:"description"`
}

// DataAggregation represents a data aggregation
type DataAggregation struct {
	Field    string                 `json:"field"`
	Function string                 `json:"function"` // "sum", "avg", "count", "min", "max"
	GroupBy  []string               `json:"group_by"`
	Alias    string                 `json:"alias"`
	Config   map[string]interface{} `json:"config"`
}

// DestinationConfig represents the configuration for a destination
type DestinationConfig struct {
	Type        string                 `json:"type"` // "bi_gateway", "data_warehouse", "api", "file"
	Endpoint    string                 `json:"endpoint"`
	Credentials map[string]interface{} `json:"credentials"`
	Format      DataExportFormat       `json:"format"`
	Compression bool                   `json:"compression"`
	BatchSize   int                    `json:"batch_size"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SyncSchedule represents the schedule for data synchronization
type SyncSchedule struct {
	Frequency ScheduleFrequency      `json:"frequency"`
	Interval  int                    `json:"interval"`
	TimeOfDay string                 `json:"time_of_day"` // HH:MM format
	Timezone  string                 `json:"timezone"`
	Enabled   bool                   `json:"enabled"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ScheduleFrequency represents the frequency of synchronization
type ScheduleFrequency string

const (
	ScheduleFrequencyRealTime ScheduleFrequency = "realtime"
	ScheduleFrequencyHourly   ScheduleFrequency = "hourly"
	ScheduleFrequencyDaily    ScheduleFrequency = "daily"
	ScheduleFrequencyWeekly   ScheduleFrequency = "weekly"
	ScheduleFrequencyMonthly  ScheduleFrequency = "monthly"
	ScheduleFrequencyManual   ScheduleFrequency = "manual"
)

// DataExport represents a data export job
type DataExport struct {
	ID              string                 `json:"id" db:"id"`
	TenantID        string                 `json:"tenant_id" db:"tenant_id"`
	Name            string                 `json:"name" db:"name"`
	DataSourceType  DataSourceType         `json:"data_source_type" db:"data_source_type"`
	SourceConfig    DataSourceConfig       `json:"source_config" db:"source_config"`
	Format          DataExportFormat       `json:"format" db:"format"`
	Status          DataExportStatus       `json:"status" db:"status"`
	FileSize        int64                  `json:"file_size" db:"file_size"`
	DownloadURL     string                 `json:"download_url" db:"download_url"`
	RecordsExported int64                  `json:"records_exported" db:"records_exported"`
	Error           string                 `json:"error" db:"error"`
	CreatedBy       string                 `json:"created_by" db:"created_by"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	CompletedAt     *time.Time             `json:"completed_at" db:"completed_at"`
	ExpiresAt       *time.Time             `json:"expires_at" db:"expires_at"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
}

// DataExportStatus represents the status of a data export
type DataExportStatus string

const (
	DataExportStatusPending   DataExportStatus = "pending"
	DataExportStatusExporting DataExportStatus = "exporting"
	DataExportStatusCompleted DataExportStatus = "completed"
	DataExportStatusFailed    DataExportStatus = "failed"
	DataExportStatusExpired   DataExportStatus = "expired"
)

// BIQuery represents a business intelligence query
type BIQuery struct {
	ID          string                 `json:"id" db:"id"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Query       BIQueryDefinition      `json:"query" db:"query"`
	Parameters  []QueryParameter       `json:"parameters" db:"parameters"`
	IsPublic    bool                   `json:"is_public" db:"is_public"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// BIQueryDefinition represents the definition of a BI query
type BIQueryDefinition struct {
	DataSources  []QueryDataSource      `json:"data_sources"`
	Joins        []QueryJoin            `json:"joins"`
	Filters      []QueryFilter          `json:"filters"`
	GroupBy      []string               `json:"group_by"`
	OrderBy      []QueryOrderBy         `json:"order_by"`
	Limit        int                    `json:"limit"`
	Aggregations []QueryAggregation     `json:"aggregations"`
	Calculations []QueryCalculation     `json:"calculations"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// QueryDataSource represents a data source in a query
type QueryDataSource struct {
	Type    DataSourceType         `json:"type"`
	Alias   string                 `json:"alias"`
	Config  map[string]interface{} `json:"config"`
	Filters []QueryFilter          `json:"filters"`
}

// QueryJoin represents a join between data sources
type QueryJoin struct {
	Type       string           `json:"type"` // "inner", "left", "right", "outer"
	LeftTable  string           `json:"left_table"`
	RightTable string           `json:"right_table"`
	LeftField  string           `json:"left_field"`
	RightField string           `json:"right_field"`
	Conditions []QueryCondition `json:"conditions"`
}

// QueryFilter represents a filter in a query
type QueryFilter struct {
	Field      string           `json:"field"`
	Operator   string           `json:"operator"` // "eq", "ne", "gt", "lt", "gte", "lte", "in", "not_in", "like", "between"
	Value      interface{}      `json:"value"`
	Conditions []QueryCondition `json:"conditions"`
	Logic      string           `json:"logic"` // "and", "or"
}

// QueryCondition represents a condition in a query
type QueryCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// QueryOrderBy represents ordering in a query
type QueryOrderBy struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // "asc", "desc"
}

// QueryAggregation represents an aggregation in a query
type QueryAggregation struct {
	Field    string `json:"field"`
	Function string `json:"function"` // "sum", "avg", "count", "min", "max", "distinct"
	Alias    string `json:"alias"`
}

// QueryCalculation represents a calculated field in a query
type QueryCalculation struct {
	Name       string                 `json:"name"`
	Expression string                 `json:"expression"`
	Type       string                 `json:"type"` // "number", "string", "date", "boolean"
	Config     map[string]interface{} `json:"config"`
}

// QueryParameter represents a parameter in a query
type QueryParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // "string", "number", "date", "boolean"
	Default     interface{} `json:"default"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
}

// BIQueryResult represents the result of a BI query
type BIQueryResult struct {
	ID            string                   `json:"id"`
	QueryID       string                   `json:"query_id"`
	TenantID      string                   `json:"tenant_id"`
	Parameters    map[string]interface{}   `json:"parameters"`
	Data          []map[string]interface{} `json:"data"`
	Columns       []QueryColumn            `json:"columns"`
	RowCount      int64                    `json:"row_count"`
	ExecutionTime time.Duration            `json:"execution_time"`
	CreatedAt     time.Time                `json:"created_at"`
	Metadata      map[string]interface{}   `json:"metadata"`
}

// QueryColumn represents a column in a query result
type QueryColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Format      string `json:"format"`
}

// BIDashboard represents a business intelligence dashboard
type BIDashboard struct {
	ID          string                 `json:"id" db:"id"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Layout      DashboardLayout        `json:"layout" db:"layout"`
	Widgets     []DashboardWidget      `json:"widgets" db:"widgets"`
	Filters     []DashboardFilter      `json:"filters" db:"filters"`
	IsPublic    bool                   `json:"is_public" db:"is_public"`
	IsDefault   bool                   `json:"is_default" db:"is_default"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// DashboardLayout represents the layout of a dashboard
type DashboardLayout struct {
	Rows       []DashboardRow         `json:"rows"`
	Columns    int                    `json:"columns"`
	Spacing    int                    `json:"spacing"`
	Background string                 `json:"background"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// DashboardRow represents a row in a dashboard layout
type DashboardRow struct {
	Height   int                    `json:"height"`
	Widgets  []DashboardWidget      `json:"widgets"`
	Metadata map[string]interface{} `json:"metadata"`
}

// DashboardWidget represents a widget in a dashboard
type DashboardWidget struct {
	ID          string                 `json:"id"`
	Type        WidgetType             `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	QueryID     string                 `json:"query_id"`
	Config      WidgetConfig           `json:"config"`
	Position    WidgetPosition         `json:"position"`
	Size        WidgetSize             `json:"size"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// WidgetType represents the type of a widget
type WidgetType string

const (
	WidgetTypeChart  WidgetType = "chart"
	WidgetTypeTable  WidgetType = "table"
	WidgetTypeKPI    WidgetType = "kpi"
	WidgetTypeGauge  WidgetType = "gauge"
	WidgetTypeMap    WidgetType = "map"
	WidgetTypeText   WidgetType = "text"
	WidgetTypeImage  WidgetType = "image"
	WidgetTypeCustom WidgetType = "custom"
)

// WidgetConfig represents the configuration of a widget
type WidgetConfig struct {
	ChartType       string                 `json:"chart_type,omitempty"`
	Colors          []string               `json:"colors,omitempty"`
	ShowLegend      bool                   `json:"show_legend,omitempty"`
	ShowGrid        bool                   `json:"show_grid,omitempty"`
	Animation       bool                   `json:"animation,omitempty"`
	RefreshInterval int                    `json:"refresh_interval,omitempty"`
	Options         map[string]interface{} `json:"options,omitempty"`
}

// WidgetPosition represents the position of a widget
type WidgetPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// WidgetSize represents the size of a widget
type WidgetSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// DashboardFilter represents a filter in a dashboard
type DashboardFilter struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"` // "date", "select", "multiselect", "range", "text"
	Field    string                 `json:"field"`
	Default  interface{}            `json:"default"`
	Options  []FilterOption         `json:"options,omitempty"`
	Config   map[string]interface{} `json:"config"`
	Metadata map[string]interface{} `json:"metadata"`
}

// FilterOption represents an option in a filter
type FilterOption struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

// BIRequest represents a request to create/update BI resources
type BIRequest struct {
	Name        string                 `json:"name" validate:"required,min=1,max=255"`
	Description string                 `json:"description,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	CreatedBy   string                 `json:"created_by" validate:"required"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BIResponse represents a response for BI resources
type BIResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BIListResponse represents a list response for BI resources
type BIListResponse struct {
	Items    []BIResponse `json:"items"`
	Total    int          `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

// BIFilter represents filters for querying BI resources
type BIFilter struct {
	TenantID  string     `json:"tenant_id,omitempty"`
	CreatedBy string     `json:"created_by,omitempty"`
	IsPublic  *bool      `json:"is_public,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// BIMetrics represents business intelligence metrics
type BIMetrics struct {
	TotalQueries         int                 `json:"total_queries"`
	TotalDashboards      int                 `json:"total_dashboards"`
	TotalDataSyncs       int                 `json:"total_data_syncs"`
	TotalDataExports     int                 `json:"total_data_exports"`
	ActiveDataSyncs      int                 `json:"active_data_syncs"`
	FailedDataSyncs      int                 `json:"failed_data_syncs"`
	DataSyncSuccessRate  float64             `json:"data_sync_success_rate"`
	AverageSyncTime      time.Duration       `json:"average_sync_time"`
	TotalRecordsSynced   int64               `json:"total_records_synced"`
	MostUsedQueries      []QueryUsageData    `json:"most_used_queries"`
	MostViewedDashboards []DashboardViewData `json:"most_viewed_dashboards"`
}

// QueryUsageData represents query usage statistics
type QueryUsageData struct {
	QueryID    string    `json:"query_id"`
	QueryName  string    `json:"query_name"`
	UsageCount int       `json:"usage_count"`
	LastUsed   time.Time `json:"last_used"`
}

// DashboardViewData represents dashboard view statistics
type DashboardViewData struct {
	DashboardID   string    `json:"dashboard_id"`
	DashboardName string    `json:"dashboard_name"`
	ViewCount     int       `json:"view_count"`
	LastViewed    time.Time `json:"last_viewed"`
}
