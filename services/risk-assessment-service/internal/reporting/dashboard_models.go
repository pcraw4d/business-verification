package reporting

import (
	"time"
)

// DashboardType represents the type of dashboard
type DashboardType string

const (
	DashboardTypeRiskOverview DashboardType = "risk_overview"
	DashboardTypeTrends       DashboardType = "trends"
	DashboardTypePredictions  DashboardType = "predictions"
	DashboardTypeCompliance   DashboardType = "compliance"
	DashboardTypePerformance  DashboardType = "performance"
	DashboardTypeCustom       DashboardType = "custom"
)

// RiskDashboard represents the main risk assessment dashboard
type RiskDashboard struct {
	ID          string                 `json:"id" db:"id"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	Name        string                 `json:"name" db:"name"`
	Type        DashboardType          `json:"type" db:"type"`
	Summary     DashboardSummary       `json:"summary"`
	Trends      DashboardTrends        `json:"trends"`
	Predictions DashboardPredictions   `json:"predictions"`
	Charts      []DashboardChart       `json:"charts"`
	Filters     DashboardFilters       `json:"filters"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	IsPublic    bool                   `json:"is_public" db:"is_public"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// DashboardSummary provides high-level risk metrics
type DashboardSummary struct {
	TotalAssessments   int                    `json:"total_assessments"`
	AverageRiskScore   float64                `json:"average_risk_score"`
	HighRiskCount      int                    `json:"high_risk_count"`
	MediumRiskCount    int                    `json:"medium_risk_count"`
	LowRiskCount       int                    `json:"low_risk_count"`
	RiskDistribution   []RiskDistribution     `json:"risk_distribution"`
	TopRiskFactors     []RiskFactorData       `json:"top_risk_factors"`
	AssessmentVolume   AssessmentVolumeData   `json:"assessment_volume"`
	ComplianceStatus   ComplianceStatusData   `json:"compliance_status"`
	PerformanceMetrics PerformanceMetricsData `json:"performance_metrics"`
}

// DashboardTrends provides trend analysis data
type DashboardTrends struct {
	RiskScoreOverTime          []TimeSeriesData   `json:"risk_score_over_time"`
	AssessmentVolumeByTime     []TimeSeriesData   `json:"assessment_volume_by_time"`
	AssessmentVolumeByIndustry []ChartData        `json:"assessment_volume_by_industry"`
	AssessmentVolumeByCountry  []ChartData        `json:"assessment_volume_by_country"`
	RiskFactorTrends           []RiskFactorTrend  `json:"risk_factor_trends"`
	ComplianceTrends           []ComplianceTrend  `json:"compliance_trends"`
	PerformanceTrends          []PerformanceTrend `json:"performance_trends"`
}

// DashboardPredictions provides prediction analytics
type DashboardPredictions struct {
	ForecastedRisk      []ForecastData       `json:"forecasted_risk"`
	ModelAccuracy       []ModelPerformance   `json:"model_accuracy"`
	RiskPredictions     []RiskPredictionData `json:"risk_predictions"`
	ConfidenceIntervals []ConfidenceInterval `json:"confidence_intervals"`
	PredictionAccuracy  PredictionAccuracy   `json:"prediction_accuracy"`
	ModelDrift          ModelDriftData       `json:"model_drift"`
}

// DashboardChart represents a chart component
type DashboardChart struct {
	ID          string                 `json:"id"`
	Type        ChartType              `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Data        interface{}            `json:"data"`
	Config      ChartConfig            `json:"config"`
	Position    ChartPosition          `json:"position"`
	Filters     []ChartFilter          `json:"filters"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ChartType represents the type of chart
type ChartType string

const (
	ChartTypeLine    ChartType = "line"
	ChartTypeBar     ChartType = "bar"
	ChartTypePie     ChartType = "pie"
	ChartTypeArea    ChartType = "area"
	ChartTypeScatter ChartType = "scatter"
	ChartTypeHeatmap ChartType = "heatmap"
	ChartTypeGauge   ChartType = "gauge"
	ChartTypeTable   ChartType = "table"
	ChartTypeKPI     ChartType = "kpi"
)

// ChartConfig provides chart configuration
type ChartConfig struct {
	Width        int                    `json:"width"`
	Height       int                    `json:"height"`
	Colors       []string               `json:"colors"`
	ShowLegend   bool                   `json:"show_legend"`
	ShowGrid     bool                   `json:"show_grid"`
	Animation    bool                   `json:"animation"`
	Interactions bool                   `json:"interactions"`
	Options      map[string]interface{} `json:"options"`
}

// ChartPosition defines chart position in dashboard
type ChartPosition struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ChartFilter defines chart filtering options
type ChartFilter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// DashboardFilters provides dashboard-level filtering
type DashboardFilters struct {
	DateRange     DateRangeFilter `json:"date_range"`
	Industry      []string        `json:"industry"`
	Country       []string        `json:"country"`
	RiskLevel     []string        `json:"risk_level"`
	CustomFilters []CustomFilter  `json:"custom_filters"`
}

// DateRangeFilter defines date range filtering
type DateRangeFilter struct {
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	Period    string     `json:"period"` // "7d", "30d", "90d", "1y", "custom"
}

// CustomFilter defines custom filtering options
type CustomFilter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Label    string      `json:"label"`
}

// Data structures for dashboard content

// RiskDistribution represents risk level distribution
type RiskDistribution struct {
	RiskLevel  string  `json:"risk_level"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
	Color      string  `json:"color"`
}

// RiskFactorData represents risk factor information
type RiskFactorData struct {
	Factor     string  `json:"factor"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
	Impact     float64 `json:"impact"`
	Trend      string  `json:"trend"` // "up", "down", "stable"
}

// AssessmentVolumeData represents assessment volume metrics
type AssessmentVolumeData struct {
	Total       int               `json:"total"`
	Daily       int               `json:"daily"`
	Weekly      int               `json:"weekly"`
	Monthly     int               `json:"monthly"`
	GrowthRate  float64           `json:"growth_rate"`
	PeakHours   []PeakHourData    `json:"peak_hours"`
	VolumeByDay []VolumeByDayData `json:"volume_by_day"`
}

// PeakHourData represents peak usage hours
type PeakHourData struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

// VolumeByDayData represents volume by day of week
type VolumeByDayData struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

// ComplianceStatusData represents compliance status
type ComplianceStatusData struct {
	Compliant      int                   `json:"compliant"`
	NonCompliant   int                   `json:"non_compliant"`
	Pending        int                   `json:"pending"`
	ComplianceRate float64               `json:"compliance_rate"`
	TopViolations  []ComplianceViolation `json:"top_violations"`
}

// ComplianceViolation represents compliance violation data
type ComplianceViolation struct {
	Violation string `json:"violation"`
	Count     int    `json:"count"`
	Severity  string `json:"severity"`
}

// PerformanceMetricsData represents performance metrics
type PerformanceMetricsData struct {
	AverageResponseTime float64 `json:"average_response_time_ms"`
	P95ResponseTime     float64 `json:"p95_response_time_ms"`
	P99ResponseTime     float64 `json:"p99_response_time_ms"`
	ErrorRate           float64 `json:"error_rate"`
	Throughput          float64 `json:"throughput_per_minute"`
	Availability        float64 `json:"availability_percentage"`
}

// TimeSeriesData represents time series data points
type TimeSeriesData struct {
	Timestamp time.Time              `json:"timestamp"`
	Value     float64                `json:"value"`
	Label     string                 `json:"label,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ChartData represents generic chart data
type ChartData struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
	Color string  `json:"color,omitempty"`
	Count int     `json:"count,omitempty"`
}

// RiskFactorTrend represents risk factor trend data
type RiskFactorTrend struct {
	Factor    string           `json:"factor"`
	Trend     []TimeSeriesData `json:"trend"`
	Change    float64          `json:"change_percentage"`
	Direction string           `json:"direction"` // "increasing", "decreasing", "stable"
}

// ComplianceTrend represents compliance trend data
type ComplianceTrend struct {
	Metric string           `json:"metric"`
	Trend  []TimeSeriesData `json:"trend"`
	Change float64          `json:"change_percentage"`
	Status string           `json:"status"`
}

// PerformanceTrend represents performance trend data
type PerformanceTrend struct {
	Metric string           `json:"metric"`
	Trend  []TimeSeriesData `json:"trend"`
	Change float64          `json:"change_percentage"`
	Target float64          `json:"target_value"`
	Status string           `json:"status"`
}

// ForecastData represents forecasted data
type ForecastData struct {
	Timestamp  time.Time `json:"timestamp"`
	Value      float64   `json:"value"`
	Confidence float64   `json:"confidence"`
	LowerBound float64   `json:"lower_bound"`
	UpperBound float64   `json:"upper_bound"`
	Model      string    `json:"model"`
}

// ModelPerformance represents ML model performance metrics
type ModelPerformance struct {
	ModelName    string    `json:"model_name"`
	Accuracy     float64   `json:"accuracy"`
	Precision    float64   `json:"precision"`
	Recall       float64   `json:"recall"`
	F1Score      float64   `json:"f1_score"`
	AUC          float64   `json:"auc"`
	LastUpdated  time.Time `json:"last_updated"`
	TrainingSize int       `json:"training_size"`
	TestSize     int       `json:"test_size"`
}

// RiskPredictionData represents risk prediction data
type RiskPredictionData struct {
	BusinessID    string    `json:"business_id"`
	BusinessName  string    `json:"business_name"`
	CurrentRisk   float64   `json:"current_risk"`
	PredictedRisk float64   `json:"predicted_risk"`
	Confidence    float64   `json:"confidence"`
	TimeHorizon   int       `json:"time_horizon_months"`
	PredictedAt   time.Time `json:"predicted_at"`
	Factors       []string  `json:"key_factors"`
}

// ConfidenceInterval represents confidence interval data
type ConfidenceInterval struct {
	Metric       string    `json:"metric"`
	Value        float64   `json:"value"`
	LowerBound   float64   `json:"lower_bound"`
	UpperBound   float64   `json:"upper_bound"`
	Confidence   float64   `json:"confidence"`
	SampleSize   int       `json:"sample_size"`
	CalculatedAt time.Time `json:"calculated_at"`
}

// PredictionAccuracy represents prediction accuracy metrics
type PredictionAccuracy struct {
	OverallAccuracy float64            `json:"overall_accuracy"`
	ByRiskLevel     map[string]float64 `json:"by_risk_level"`
	ByTimeHorizon   map[string]float64 `json:"by_time_horizon"`
	ByIndustry      map[string]float64 `json:"by_industry"`
	LastUpdated     time.Time          `json:"last_updated"`
}

// ModelDriftData represents model drift detection data
type ModelDriftData struct {
	ModelName      string    `json:"model_name"`
	DriftDetected  bool      `json:"drift_detected"`
	DriftScore     float64   `json:"drift_score"`
	DriftThreshold float64   `json:"drift_threshold"`
	LastChecked    time.Time `json:"last_checked"`
	DriftFactors   []string  `json:"drift_factors"`
	Recommendation string    `json:"recommendation"`
}

// DashboardRequest represents a request to create/update a dashboard
type DashboardRequest struct {
	Name      string                 `json:"name" validate:"required,min=1,max=255"`
	Type      DashboardType          `json:"type" validate:"required"`
	Charts    []DashboardChart       `json:"charts,omitempty"`
	Filters   DashboardFilters       `json:"filters,omitempty"`
	IsPublic  bool                   `json:"is_public,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy string                 `json:"created_by" validate:"required"`
}

// DashboardResponse represents a dashboard response
type DashboardResponse struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Type      DashboardType `json:"type"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	CreatedBy string        `json:"created_by"`
	IsPublic  bool          `json:"is_public"`
}

// DashboardListResponse represents a list of dashboards
type DashboardListResponse struct {
	Dashboards []DashboardResponse `json:"dashboards"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
}

// DashboardFilter represents filters for querying dashboards
type DashboardFilter struct {
	TenantID  string        `json:"tenant_id,omitempty"`
	Type      DashboardType `json:"type,omitempty"`
	CreatedBy string        `json:"created_by,omitempty"`
	IsPublic  *bool         `json:"is_public,omitempty"`
	StartDate *time.Time    `json:"start_date,omitempty"`
	EndDate   *time.Time    `json:"end_date,omitempty"`
	Limit     int           `json:"limit,omitempty"`
	Offset    int           `json:"offset,omitempty"`
}

// DashboardMetrics represents dashboard usage metrics
type DashboardMetrics struct {
	TotalDashboards   int                 `json:"total_dashboards"`
	PublicDashboards  int                 `json:"public_dashboards"`
	PrivateDashboards int                 `json:"private_dashboards"`
	MostViewed        []DashboardViewData `json:"most_viewed"`
	AverageViews      float64             `json:"average_views"`
	TotalViews        int                 `json:"total_views"`
}

// DashboardViewData represents dashboard view statistics
type DashboardViewData struct {
	DashboardID   string    `json:"dashboard_id"`
	DashboardName string    `json:"dashboard_name"`
	ViewCount     int       `json:"view_count"`
	LastViewed    time.Time `json:"last_viewed"`
}
