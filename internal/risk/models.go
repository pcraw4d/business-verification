package risk

import (
	"time"
)

// RiskCategory represents different types of risk that can be assessed
type RiskCategory string

const (
	RiskCategoryOperational   RiskCategory = "operational"
	RiskCategoryFinancial     RiskCategory = "financial"
	RiskCategoryRegulatory    RiskCategory = "regulatory"
	RiskCategoryReputational  RiskCategory = "reputational"
	RiskCategoryCybersecurity RiskCategory = "cybersecurity"
)

// RiskLevel represents the severity level of a risk
type RiskLevel string

const (
	RiskLevelMinimal  RiskLevel = "minimal"
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// DateRange represents a date range for filtering
// Note: Also defined in automated_alerts_stub.go - using that definition
// type DateRange struct {
// 	StartDate time.Time `json:"start_date"`
// 	EndDate   time.Time `json:"end_date"`
// }

// RiskFactor represents a specific risk factor that contributes to overall risk assessment
type RiskFactor struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    RiskCategory           `json:"category"`
	Weight      float64                `json:"weight"` // 0.0 to 1.0
	Thresholds  map[RiskLevel]float64  `json:"thresholds"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// RiskScore represents a calculated risk score for a specific factor or overall
type RiskScore struct {
	FactorID     string       `json:"factor_id"`
	FactorName   string       `json:"factor_name"`
	Category     RiskCategory `json:"category"`
	Score        float64      `json:"score"` // 0.0 to 100.0
	Level        RiskLevel    `json:"level"`
	Confidence   float64      `json:"confidence"` // 0.0 to 1.0
	Explanation  string       `json:"explanation"`
	Evidence     []string     `json:"evidence"`
	CalculatedAt time.Time    `json:"calculated_at"`
}

// RiskAssessment represents a complete risk assessment for a business
type RiskAssessment struct {
	ID              string                     `json:"id"`
	BusinessID      string                     `json:"business_id"`
	BusinessName    string                     `json:"business_name"`
	OverallScore    float64                    `json:"overall_score"`
	OverallLevel    RiskLevel                  `json:"overall_level"`
	CategoryScores  map[RiskCategory]RiskScore `json:"category_scores"`
	FactorScores    []RiskScore                `json:"factor_scores"`
	Recommendations []RiskRecommendation       `json:"recommendations"`
	Alerts          []RiskAlert                `json:"alerts,omitempty"`
	AlertLevel      RiskLevel                  `json:"alert_level"`
	AssessedAt      time.Time                  `json:"assessed_at"`
	ValidUntil      time.Time                  `json:"valid_until"`
	Metadata        map[string]interface{}     `json:"metadata,omitempty"`
}

// RiskRecommendation represents a recommendation to mitigate or address a risk
type RiskRecommendation struct {
	ID          string    `json:"id"`
	RiskFactor  string    `json:"risk_factor"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    RiskLevel `json:"priority"`
	Action      string    `json:"action"`
	Impact      string    `json:"impact"`
	Timeline    string    `json:"timeline"`
	CreatedAt   time.Time `json:"created_at"`
}

// RiskThreshold represents configurable thresholds for risk levels
type RiskThreshold struct {
	Category    RiskCategory `json:"category"`
	LowMax      float64      `json:"low_max"`
	MediumMax   float64      `json:"medium_max"`
	HighMax     float64      `json:"high_max"`
	CriticalMin float64      `json:"critical_min"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// RiskAlert represents an alert triggered by risk assessment
type RiskAlert struct {
	ID             string     `json:"id"`
	BusinessID     string     `json:"business_id"`
	RiskFactor     string     `json:"risk_factor"`
	Level          RiskLevel  `json:"level"`
	Message        string     `json:"message"`
	Score          float64    `json:"score"`
	Threshold      float64    `json:"threshold"`
	TriggeredAt    time.Time  `json:"triggered_at"`
	Acknowledged   bool       `json:"acknowledged"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
}

// RiskTrend represents historical risk trend data
type RiskTrend struct {
	BusinessID   string       `json:"business_id"`
	Category     RiskCategory `json:"category"`
	Score        float64      `json:"score"`
	Level        RiskLevel    `json:"level"`
	RecordedAt   time.Time    `json:"recorded_at"`
	ChangeFrom   float64      `json:"change_from"`
	ChangePeriod string       `json:"change_period"`
}

// RiskPrediction represents a risk prediction for future assessment
type RiskPrediction struct {
	ID             string    `json:"id"`
	BusinessID     string    `json:"business_id"`
	FactorID       string    `json:"factor_id"`
	PredictedScore float64   `json:"predicted_score"`
	PredictedLevel RiskLevel `json:"predicted_level"`
	Confidence     float64   `json:"confidence"`
	Horizon        string    `json:"horizon"` // e.g., "3months", "6months", "1year"
	PredictedAt    time.Time `json:"predicted_at"`
	Factors        []string  `json:"factors"` // contributing factors
}

// RiskData represents external data used for risk assessment
type RiskData struct {
	ID          string                 `json:"id"`
	BusinessID  string                 `json:"business_id"`
	Source      string                 `json:"source"`
	DataType    string                 `json:"data_type"`
	Data        map[string]interface{} `json:"data"`
	Reliability float64                `json:"reliability"` // 0.0 to 1.0
	CollectedAt time.Time              `json:"collected_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// RiskAssessmentRequest represents a request to perform a risk assessment
type RiskAssessmentRequest struct {
	BusinessID         string                 `json:"business_id"`
	BusinessName       string                 `json:"business_name"`
	Categories         []RiskCategory         `json:"categories,omitempty"`
	Factors            []string               `json:"factors,omitempty"`
	IncludeHistory     bool                   `json:"include_history"`
	IncludePredictions bool                   `json:"include_predictions"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// RiskAssessmentResponse represents the response from a risk assessment
type RiskAssessmentResponse struct {
	Assessment  *RiskAssessment  `json:"assessment"`
	Trends      []RiskTrend      `json:"trends,omitempty"`
	Predictions []RiskPrediction `json:"predictions,omitempty"`
	Alerts      []RiskAlert      `json:"alerts,omitempty"`
	GeneratedAt time.Time        `json:"generated_at"`
}

// ExportFormat represents the format for data export
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatXML  ExportFormat = "xml"
	ExportFormatPDF  ExportFormat = "pdf"
	ExportFormatXLSX ExportFormat = "xlsx"
)

// ExportType represents the type of data to export
type ExportType string

const (
	ExportTypeAssessments ExportType = "assessments"
	ExportTypeFactors     ExportType = "factors"
	ExportTypeTrends      ExportType = "trends"
	ExportTypeAlerts      ExportType = "alerts"
	ExportTypeReports     ExportType = "reports"
	ExportTypeAll         ExportType = "all"
	ExportTypeAllData     ExportType = "all_data" // Alias for ExportTypeAll
)

// ExportRequest represents a request to export risk data
type ExportRequest struct {
	BusinessID    string                 `json:"business_id"`
	ExportType    ExportType             `json:"export_type"`
	Format        ExportFormat           `json:"format"`
	DateRange     *DateRange             `json:"date_range,omitempty"`
	Categories    []RiskCategory         `json:"categories,omitempty"`
	IncludeCharts bool                   `json:"include_charts"`
	IncludeTrends bool                   `json:"include_trends"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ExportResponse represents the response from a data export request
type ExportResponse struct {
	ExportID    string                 `json:"export_id"`
	BusinessID  string                 `json:"business_id"`
	ExportType  ExportType             `json:"export_type"`
	Format      ExportFormat           `json:"format"`
	Data        interface{}            `json:"data"`
	RecordCount int                    `json:"record_count"`
	FileSize    int64                  `json:"file_size,omitempty"`
	DownloadURL string                 `json:"download_url,omitempty"`
	GeneratedAt time.Time              `json:"generated_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ExportJob represents a background export job
type ExportJob struct {
	ID          string                 `json:"id"`
	BusinessID  string                 `json:"business_id"`
	ExportType  ExportType             `json:"export_type"`
	Format      ExportFormat           `json:"format"`
	Status      string                 `json:"status"`   // "pending", "processing", "completed", "failed"
	Progress    int                    `json:"progress"` // 0-100
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Result      *ExportResponse        `json:"result,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
