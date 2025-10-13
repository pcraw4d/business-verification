package reporting

import (
	"time"
)

// ReportType represents the type of report
type ReportType string

const (
	ReportTypeExecutiveSummary ReportType = "executive_summary"
	ReportTypeCompliance       ReportType = "compliance"
	ReportTypeRiskAudit        ReportType = "risk_audit"
	ReportTypeTrendAnalysis    ReportType = "trend_analysis"
	ReportTypeCustom           ReportType = "custom"
	ReportTypeBatchResults     ReportType = "batch_results"
	ReportTypePerformance      ReportType = "performance"
)

// ReportStatus represents the status of a report
type ReportStatus string

const (
	ReportStatusPending    ReportStatus = "pending"
	ReportStatusGenerating ReportStatus = "generating"
	ReportStatusCompleted  ReportStatus = "completed"
	ReportStatusFailed     ReportStatus = "failed"
	ReportStatusExpired    ReportStatus = "expired"
)

// ReportFormat represents the output format of a report
type ReportFormat string

const (
	ReportFormatPDF   ReportFormat = "pdf"
	ReportFormatExcel ReportFormat = "excel"
	ReportFormatCSV   ReportFormat = "csv"
	ReportFormatJSON  ReportFormat = "json"
	ReportFormatHTML  ReportFormat = "html"
)

// Report represents a generated report
type Report struct {
	ID          string                 `json:"id" db:"id"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	Name        string                 `json:"name" db:"name"`
	Type        ReportType             `json:"type" db:"type"`
	Status      ReportStatus           `json:"status" db:"status"`
	Format      ReportFormat           `json:"format" db:"format"`
	TemplateID  string                 `json:"template_id" db:"template_id"`
	Data        ReportData             `json:"data" db:"data"`
	Filters     ReportFilters          `json:"filters" db:"filters"`
	GeneratedAt *time.Time             `json:"generated_at" db:"generated_at"`
	ExpiresAt   *time.Time             `json:"expires_at" db:"expires_at"`
	FileSize    int64                  `json:"file_size" db:"file_size"`
	DownloadURL string                 `json:"download_url" db:"download_url"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	Error       string                 `json:"error" db:"error"`
}

// ReportData represents the data content of a report
type ReportData struct {
	Summary         ReportSummary          `json:"summary"`
	Charts          []ReportChart          `json:"charts"`
	Tables          []ReportTable          `json:"tables"`
	Insights        []ReportInsight        `json:"insights"`
	Recommendations []ReportRecommendation `json:"recommendations"`
	RawData         interface{}            `json:"raw_data"`
}

// ReportSummary provides high-level report summary
type ReportSummary struct {
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	Period           string                 `json:"period"`
	TotalRecords     int                    `json:"total_records"`
	KeyMetrics       map[string]interface{} `json:"key_metrics"`
	ExecutiveSummary string                 `json:"executive_summary"`
	GeneratedAt      time.Time              `json:"generated_at"`
}

// ReportChart represents a chart in the report
type ReportChart struct {
	ID          string                 `json:"id"`
	Type        ChartType              `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Data        interface{}            `json:"data"`
	Config      ChartConfig            `json:"config"`
	Position    ChartPosition          `json:"position"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ReportTable represents a table in the report
type ReportTable struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Headers     []string               `json:"headers"`
	Rows        [][]interface{}        `json:"rows"`
	Summary     map[string]interface{} `json:"summary"`
	Position    ChartPosition          `json:"position"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ReportInsight represents an insight in the report
type ReportInsight struct {
	ID          string                 `json:"id"`
	Type        InsightType            `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Impact      InsightImpact          `json:"impact"`
	Confidence  float64                `json:"confidence"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// InsightType represents the type of insight
type InsightType string

const (
	InsightTypeRiskTrend      InsightType = "risk_trend"
	InsightTypeCompliance     InsightType = "compliance"
	InsightTypePerformance    InsightType = "performance"
	InsightTypeAnomaly        InsightType = "anomaly"
	InsightTypePrediction     InsightType = "prediction"
	InsightTypeRecommendation InsightType = "recommendation"
)

// InsightImpact represents the impact level of an insight
type InsightImpact string

const (
	InsightImpactLow      InsightImpact = "low"
	InsightImpactMedium   InsightImpact = "medium"
	InsightImpactHigh     InsightImpact = "high"
	InsightImpactCritical InsightImpact = "critical"
)

// ReportRecommendation represents a recommendation in the report
type ReportRecommendation struct {
	ID          string                 `json:"id"`
	Type        RecommendationType     `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Priority    RecommendationPriority `json:"priority"`
	Action      string                 `json:"action"`
	Timeline    string                 `json:"timeline"`
	Resources   []string               `json:"resources"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RecommendationType represents the type of recommendation
type RecommendationType string

const (
	RecommendationTypeRiskMitigation RecommendationType = "risk_mitigation"
	RecommendationTypeCompliance     RecommendationType = "compliance"
	RecommendationTypeProcess        RecommendationType = "process"
	RecommendationTypeTechnology     RecommendationType = "technology"
	RecommendationTypeTraining       RecommendationType = "training"
)

// RecommendationPriority represents the priority of a recommendation
type RecommendationPriority string

const (
	RecommendationPriorityLow      RecommendationPriority = "low"
	RecommendationPriorityMedium   RecommendationPriority = "medium"
	RecommendationPriorityHigh     RecommendationPriority = "high"
	RecommendationPriorityCritical RecommendationPriority = "critical"
)

// ReportFilters represents filters applied to the report
type ReportFilters struct {
	DateRange     DateRangeFilter `json:"date_range"`
	Industry      []string        `json:"industry"`
	Country       []string        `json:"country"`
	RiskLevel     []string        `json:"risk_level"`
	BusinessID    []string        `json:"business_id"`
	CustomFilters []CustomFilter  `json:"custom_filters"`
}

// ReportTemplate represents a report template
type ReportTemplate struct {
	ID          string                 `json:"id" db:"id"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	Name        string                 `json:"name" db:"name"`
	Type        ReportType             `json:"type" db:"type"`
	Description string                 `json:"description" db:"description"`
	Template    ReportTemplateConfig   `json:"template" db:"template"`
	IsPublic    bool                   `json:"is_public" db:"is_public"`
	IsDefault   bool                   `json:"is_default" db:"is_default"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// ReportTemplateConfig represents the configuration of a report template
type ReportTemplateConfig struct {
	Layout   ReportLayout           `json:"layout"`
	Sections []ReportSection        `json:"sections"`
	Charts   []ReportChartTemplate  `json:"charts"`
	Tables   []ReportTableTemplate  `json:"tables"`
	Styling  ReportStyling          `json:"styling"`
	Branding ReportBranding         `json:"branding"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ReportLayout represents the layout configuration
type ReportLayout struct {
	PageSize    string                 `json:"page_size"`   // A4, Letter, etc.
	Orientation string                 `json:"orientation"` // portrait, landscape
	Margins     ReportMargins          `json:"margins"`
	Header      ReportHeaderFooter     `json:"header"`
	Footer      ReportHeaderFooter     `json:"footer"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ReportMargins represents page margins
type ReportMargins struct {
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
}

// ReportHeaderFooter represents header/footer configuration
type ReportHeaderFooter struct {
	Enabled   bool                   `json:"enabled"`
	Content   string                 `json:"content"`
	Height    float64                `json:"height"`
	FontSize  float64                `json:"font_size"`
	Alignment string                 `json:"alignment"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ReportSection represents a section in the report
type ReportSection struct {
	ID       string                 `json:"id"`
	Title    string                 `json:"title"`
	Type     SectionType            `json:"type"`
	Content  interface{}            `json:"content"`
	Order    int                    `json:"order"`
	Visible  bool                   `json:"visible"`
	Metadata map[string]interface{} `json:"metadata"`
}

// SectionType represents the type of report section
type SectionType string

const (
	SectionTypeSummary         SectionType = "summary"
	SectionTypeCharts          SectionType = "charts"
	SectionTypeTables          SectionType = "tables"
	SectionTypeInsights        SectionType = "insights"
	SectionTypeRecommendations SectionType = "recommendations"
	SectionTypeRawData         SectionType = "raw_data"
	SectionTypeCustom          SectionType = "custom"
)

// ReportChartTemplate represents a chart template
type ReportChartTemplate struct {
	ID         string                 `json:"id"`
	Type       ChartType              `json:"type"`
	Title      string                 `json:"title"`
	DataSource string                 `json:"data_source"`
	Config     ChartConfig            `json:"config"`
	Position   ChartPosition          `json:"position"`
	Visible    bool                   `json:"visible"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ReportTableTemplate represents a table template
type ReportTableTemplate struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	DataSource string                 `json:"data_source"`
	Columns    []TableColumn          `json:"columns"`
	Position   ChartPosition          `json:"position"`
	Visible    bool                   `json:"visible"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TableColumn represents a table column configuration
type TableColumn struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Format     string                 `json:"format"`
	Width      float64                `json:"width"`
	Alignment  string                 `json:"alignment"`
	Sortable   bool                   `json:"sortable"`
	Filterable bool                   `json:"filterable"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ReportStyling represents styling configuration
type ReportStyling struct {
	FontFamily string                 `json:"font_family"`
	FontSize   float64                `json:"font_size"`
	Colors     ReportColors           `json:"colors"`
	Spacing    ReportSpacing          `json:"spacing"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ReportColors represents color scheme
type ReportColors struct {
	Primary    string `json:"primary"`
	Secondary  string `json:"secondary"`
	Accent     string `json:"accent"`
	Background string `json:"background"`
	Text       string `json:"text"`
	Header     string `json:"header"`
	Footer     string `json:"footer"`
}

// ReportSpacing represents spacing configuration
type ReportSpacing struct {
	SectionSpacing float64 `json:"section_spacing"`
	ElementSpacing float64 `json:"element_spacing"`
	LineSpacing    float64 `json:"line_spacing"`
}

// ReportBranding represents branding configuration
type ReportBranding struct {
	LogoURL     string                 `json:"logo_url"`
	CompanyName string                 `json:"company_name"`
	ContactInfo string                 `json:"contact_info"`
	Website     string                 `json:"website"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ScheduledReport represents a scheduled report
type ScheduledReport struct {
	ID         string                 `json:"id" db:"id"`
	TenantID   string                 `json:"tenant_id" db:"tenant_id"`
	Name       string                 `json:"name" db:"name"`
	TemplateID string                 `json:"template_id" db:"template_id"`
	Schedule   ReportSchedule         `json:"schedule" db:"schedule"`
	Filters    ReportFilters          `json:"filters" db:"filters"`
	Recipients []ReportRecipient      `json:"recipients" db:"recipients"`
	IsActive   bool                   `json:"is_active" db:"is_active"`
	LastRunAt  *time.Time             `json:"last_run_at" db:"last_run_at"`
	NextRunAt  *time.Time             `json:"next_run_at" db:"next_run_at"`
	CreatedBy  string                 `json:"created_by" db:"created_by"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" db:"updated_at"`
	Metadata   map[string]interface{} `json:"metadata" db:"metadata"`
}

// ReportSchedule represents the schedule configuration
type ReportSchedule struct {
	Frequency   ScheduleFrequency      `json:"frequency"`
	Interval    int                    `json:"interval"`
	DaysOfWeek  []int                  `json:"days_of_week"` // 0=Sunday, 1=Monday, etc.
	DaysOfMonth []int                  `json:"days_of_month"`
	TimeOfDay   string                 `json:"time_of_day"` // HH:MM format
	Timezone    string                 `json:"timezone"`
	StartDate   *time.Time             `json:"start_date"`
	EndDate     *time.Time             `json:"end_date"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ScheduleFrequency represents the frequency of report generation
type ScheduleFrequency string

const (
	ScheduleFrequencyOnce      ScheduleFrequency = "once"
	ScheduleFrequencyDaily     ScheduleFrequency = "daily"
	ScheduleFrequencyWeekly    ScheduleFrequency = "weekly"
	ScheduleFrequencyMonthly   ScheduleFrequency = "monthly"
	ScheduleFrequencyQuarterly ScheduleFrequency = "quarterly"
	ScheduleFrequencyYearly    ScheduleFrequency = "yearly"
)

// ReportRecipient represents a report recipient
type ReportRecipient struct {
	Type     RecipientType          `json:"type"`
	Value    string                 `json:"value"`
	Name     string                 `json:"name"`
	Format   ReportFormat           `json:"format"`
	Metadata map[string]interface{} `json:"metadata"`
}

// RecipientType represents the type of recipient
type RecipientType string

const (
	RecipientTypeEmail   RecipientType = "email"
	RecipientTypeWebhook RecipientType = "webhook"
	RecipientTypeSlack   RecipientType = "slack"
	RecipientTypeTeams   RecipientType = "teams"
)

// ReportRequest represents a request to generate a report
type ReportRequest struct {
	Name       string                 `json:"name" validate:"required,min=1,max=255"`
	Type       ReportType             `json:"type" validate:"required"`
	TemplateID string                 `json:"template_id,omitempty"`
	Format     ReportFormat           `json:"format" validate:"required"`
	Filters    ReportFilters          `json:"filters,omitempty"`
	Recipients []ReportRecipient      `json:"recipients,omitempty"`
	ExpiresIn  int                    `json:"expires_in,omitempty"` // hours
	CreatedBy  string                 `json:"created_by" validate:"required"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ReportResponse represents a report response
type ReportResponse struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Type        ReportType   `json:"type"`
	Status      ReportStatus `json:"status"`
	Format      ReportFormat `json:"format"`
	GeneratedAt *time.Time   `json:"generated_at"`
	ExpiresAt   *time.Time   `json:"expires_at"`
	FileSize    int64        `json:"file_size"`
	DownloadURL string       `json:"download_url"`
	CreatedBy   string       `json:"created_by"`
}

// ReportListResponse represents a list of reports
type ReportListResponse struct {
	Reports  []ReportResponse `json:"reports"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// ReportFilter represents filters for querying reports
type ReportFilter struct {
	TenantID  string       `json:"tenant_id,omitempty"`
	Type      ReportType   `json:"type,omitempty"`
	Status    ReportStatus `json:"status,omitempty"`
	Format    ReportFormat `json:"format,omitempty"`
	CreatedBy string       `json:"created_by,omitempty"`
	StartDate *time.Time   `json:"start_date,omitempty"`
	EndDate   *time.Time   `json:"end_date,omitempty"`
	Limit     int          `json:"limit,omitempty"`
	Offset    int          `json:"offset,omitempty"`
}

// ReportMetrics represents report usage metrics
type ReportMetrics struct {
	TotalReports      int                  `json:"total_reports"`
	ReportsByType     map[ReportType]int   `json:"reports_by_type"`
	ReportsByStatus   map[ReportStatus]int `json:"reports_by_status"`
	ReportsByFormat   map[ReportFormat]int `json:"reports_by_format"`
	TotalFileSize     int64                `json:"total_file_size"`
	AverageFileSize   float64              `json:"average_file_size"`
	MostUsedTemplates []TemplateUsageData  `json:"most_used_templates"`
	GenerationTime    ReportGenerationTime `json:"generation_time"`
}

// TemplateUsageData represents template usage statistics
type TemplateUsageData struct {
	TemplateID   string    `json:"template_id"`
	TemplateName string    `json:"template_name"`
	UsageCount   int       `json:"usage_count"`
	LastUsed     time.Time `json:"last_used"`
}

// ReportGenerationTime represents report generation time metrics
type ReportGenerationTime struct {
	Average float64 `json:"average_seconds"`
	Min     float64 `json:"min_seconds"`
	Max     float64 `json:"max_seconds"`
	P95     float64 `json:"p95_seconds"`
	P99     float64 `json:"p99_seconds"`
}
