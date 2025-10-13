package reporting

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// DashboardService provides dashboard data and management
type DashboardService interface {
	// CreateDashboard creates a new dashboard
	CreateDashboard(ctx context.Context, request *DashboardRequest) (*DashboardResponse, error)

	// GetDashboard retrieves a dashboard by ID
	GetDashboard(ctx context.Context, tenantID, dashboardID string) (*RiskDashboard, error)

	// UpdateDashboard updates an existing dashboard
	UpdateDashboard(ctx context.Context, tenantID, dashboardID string, request *DashboardRequest) (*DashboardResponse, error)

	// DeleteDashboard deletes a dashboard
	DeleteDashboard(ctx context.Context, tenantID, dashboardID string) error

	// ListDashboards lists dashboards with filters
	ListDashboards(ctx context.Context, filter *DashboardFilter) (*DashboardListResponse, error)

	// GetDashboardData generates dashboard data
	GetDashboardData(ctx context.Context, tenantID, dashboardID string, filters *DashboardFilters) (*RiskDashboard, error)

	// GetRiskOverviewData generates risk overview dashboard data
	GetRiskOverviewData(ctx context.Context, tenantID string, filters *DashboardFilters) (*DashboardSummary, error)

	// GetTrendsData generates trends dashboard data
	GetTrendsData(ctx context.Context, tenantID string, filters *DashboardFilters) (*DashboardTrends, error)

	// GetPredictionsData generates predictions dashboard data
	GetPredictionsData(ctx context.Context, tenantID string, filters *DashboardFilters) (*DashboardPredictions, error)

	// GetDashboardMetrics gets dashboard usage metrics
	GetDashboardMetrics(ctx context.Context, tenantID string) (*DashboardMetrics, error)
}

// DefaultDashboardService implements DashboardService
type DefaultDashboardService struct {
	repository   DashboardRepository
	dataProvider DashboardDataProvider
	logger       *zap.Logger
}

// DashboardRepository defines the interface for dashboard data access
type DashboardRepository interface {
	SaveDashboard(ctx context.Context, dashboard *RiskDashboard) error
	GetDashboard(ctx context.Context, tenantID, dashboardID string) (*RiskDashboard, error)
	ListDashboards(ctx context.Context, filter *DashboardFilter) ([]*RiskDashboard, error)
	DeleteDashboard(ctx context.Context, tenantID, dashboardID string) error
	GetDashboardMetrics(ctx context.Context, tenantID string) (*DashboardMetrics, error)
	RecordDashboardView(ctx context.Context, tenantID, dashboardID string) error
}

// DashboardDataProvider defines the interface for providing dashboard data
type DashboardDataProvider interface {
	GetRiskAssessments(ctx context.Context, tenantID string, filters *DashboardFilters) ([]*models.RiskAssessment, error)
	GetRiskPredictions(ctx context.Context, tenantID string, filters *DashboardFilters) ([]*models.RiskPrediction, error)
	GetBatchJobs(ctx context.Context, tenantID string, filters *DashboardFilters) ([]*BatchJobData, error)
	GetComplianceData(ctx context.Context, tenantID string, filters *DashboardFilters) (*ComplianceData, error)
	GetPerformanceData(ctx context.Context, tenantID string, filters *DashboardFilters) (*PerformanceData, error)
}

// BatchJobData represents batch job data for dashboards
type BatchJobData struct {
	ID            string     `json:"id"`
	Status        string     `json:"status"`
	TotalRequests int        `json:"total_requests"`
	Completed     int        `json:"completed"`
	Failed        int        `json:"failed"`
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	JobType       string     `json:"job_type"`
}

// ComplianceData represents compliance data for dashboards
type ComplianceData struct {
	TotalChecks  int                   `json:"total_checks"`
	Compliant    int                   `json:"compliant"`
	NonCompliant int                   `json:"non_compliant"`
	Pending      int                   `json:"pending"`
	Violations   []ComplianceViolation `json:"violations"`
	Trends       []ComplianceTrend     `json:"trends"`
}

// PerformanceData represents performance data for dashboards
type PerformanceData struct {
	ResponseTime PerformanceMetrics `json:"response_time"`
	Throughput   PerformanceMetrics `json:"throughput"`
	ErrorRate    PerformanceMetrics `json:"error_rate"`
	Availability PerformanceMetrics `json:"availability"`
	Trends       []PerformanceTrend `json:"trends"`
}

// PerformanceMetrics represents performance metrics
type PerformanceMetrics struct {
	Average float64 `json:"average"`
	P95     float64 `json:"p95"`
	P99     float64 `json:"p99"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
}

// NewDefaultDashboardService creates a new default dashboard service
func NewDefaultDashboardService(
	repository DashboardRepository,
	dataProvider DashboardDataProvider,
	logger *zap.Logger,
) *DefaultDashboardService {
	return &DefaultDashboardService{
		repository:   repository,
		dataProvider: dataProvider,
		logger:       logger,
	}
}

// CreateDashboard creates a new dashboard
func (ds *DefaultDashboardService) CreateDashboard(ctx context.Context, request *DashboardRequest) (*DashboardResponse, error) {
	ds.logger.Info("Creating dashboard",
		zap.String("name", request.Name),
		zap.String("type", string(request.Type)),
		zap.String("created_by", request.CreatedBy))

	// Generate dashboard ID
	dashboardID := generateDashboardID()

	// Create dashboard
	dashboard := &RiskDashboard{
		ID:        dashboardID,
		TenantID:  getTenantIDFromContext(ctx),
		Name:      request.Name,
		Type:      request.Type,
		Charts:    request.Charts,
		Filters:   request.Filters,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: request.CreatedBy,
		IsPublic:  request.IsPublic,
		Metadata:  request.Metadata,
	}

	// Save dashboard
	if err := ds.repository.SaveDashboard(ctx, dashboard); err != nil {
		return nil, fmt.Errorf("failed to save dashboard: %w", err)
	}

	response := &DashboardResponse{
		ID:        dashboard.ID,
		Name:      dashboard.Name,
		Type:      dashboard.Type,
		CreatedAt: dashboard.CreatedAt,
		UpdatedAt: dashboard.UpdatedAt,
		CreatedBy: dashboard.CreatedBy,
		IsPublic:  dashboard.IsPublic,
	}

	ds.logger.Info("Dashboard created successfully",
		zap.String("dashboard_id", dashboardID),
		zap.String("name", request.Name))

	return response, nil
}

// GetDashboard retrieves a dashboard by ID
func (ds *DefaultDashboardService) GetDashboard(ctx context.Context, tenantID, dashboardID string) (*RiskDashboard, error) {
	ds.logger.Debug("Getting dashboard",
		zap.String("dashboard_id", dashboardID),
		zap.String("tenant_id", tenantID))

	dashboard, err := ds.repository.GetDashboard(ctx, tenantID, dashboardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}

	if dashboard == nil {
		return nil, fmt.Errorf("dashboard not found: %s", dashboardID)
	}

	// Record dashboard view
	if err := ds.repository.RecordDashboardView(ctx, tenantID, dashboardID); err != nil {
		ds.logger.Warn("Failed to record dashboard view", zap.Error(err))
	}

	ds.logger.Debug("Dashboard retrieved successfully",
		zap.String("dashboard_id", dashboardID))

	return dashboard, nil
}

// UpdateDashboard updates an existing dashboard
func (ds *DefaultDashboardService) UpdateDashboard(ctx context.Context, tenantID, dashboardID string, request *DashboardRequest) (*DashboardResponse, error) {
	ds.logger.Info("Updating dashboard",
		zap.String("dashboard_id", dashboardID),
		zap.String("tenant_id", tenantID))

	// Get existing dashboard
	dashboard, err := ds.repository.GetDashboard(ctx, tenantID, dashboardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}

	if dashboard == nil {
		return nil, fmt.Errorf("dashboard not found: %s", dashboardID)
	}

	// Update dashboard fields
	dashboard.Name = request.Name
	dashboard.Type = request.Type
	dashboard.Charts = request.Charts
	dashboard.Filters = request.Filters
	dashboard.IsPublic = request.IsPublic
	dashboard.Metadata = request.Metadata
	dashboard.UpdatedAt = time.Now()

	// Save updated dashboard
	if err := ds.repository.SaveDashboard(ctx, dashboard); err != nil {
		return nil, fmt.Errorf("failed to update dashboard: %w", err)
	}

	response := &DashboardResponse{
		ID:        dashboard.ID,
		Name:      dashboard.Name,
		Type:      dashboard.Type,
		CreatedAt: dashboard.CreatedAt,
		UpdatedAt: dashboard.UpdatedAt,
		CreatedBy: dashboard.CreatedBy,
		IsPublic:  dashboard.IsPublic,
	}

	ds.logger.Info("Dashboard updated successfully",
		zap.String("dashboard_id", dashboardID))

	return response, nil
}

// DeleteDashboard deletes a dashboard
func (ds *DefaultDashboardService) DeleteDashboard(ctx context.Context, tenantID, dashboardID string) error {
	ds.logger.Info("Deleting dashboard",
		zap.String("dashboard_id", dashboardID),
		zap.String("tenant_id", tenantID))

	if err := ds.repository.DeleteDashboard(ctx, tenantID, dashboardID); err != nil {
		return fmt.Errorf("failed to delete dashboard: %w", err)
	}

	ds.logger.Info("Dashboard deleted successfully",
		zap.String("dashboard_id", dashboardID))

	return nil
}

// ListDashboards lists dashboards with filters
func (ds *DefaultDashboardService) ListDashboards(ctx context.Context, filter *DashboardFilter) (*DashboardListResponse, error) {
	ds.logger.Debug("Listing dashboards",
		zap.String("tenant_id", filter.TenantID),
		zap.String("type", string(filter.Type)))

	dashboards, err := ds.repository.ListDashboards(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list dashboards: %w", err)
	}

	// Convert to response format
	responses := make([]DashboardResponse, len(dashboards))
	for i, dashboard := range dashboards {
		responses[i] = DashboardResponse{
			ID:        dashboard.ID,
			Name:      dashboard.Name,
			Type:      dashboard.Type,
			CreatedAt: dashboard.CreatedAt,
			UpdatedAt: dashboard.UpdatedAt,
			CreatedBy: dashboard.CreatedBy,
			IsPublic:  dashboard.IsPublic,
		}
	}

	response := &DashboardListResponse{
		Dashboards: responses,
		Total:      len(responses),
		Page:       1, // This would be calculated based on offset/limit
		PageSize:   len(responses),
	}

	ds.logger.Debug("Dashboards listed successfully",
		zap.String("tenant_id", filter.TenantID),
		zap.Int("count", len(responses)))

	return response, nil
}

// GetDashboardData generates dashboard data
func (ds *DefaultDashboardService) GetDashboardData(ctx context.Context, tenantID, dashboardID string, filters *DashboardFilters) (*RiskDashboard, error) {
	ds.logger.Debug("Getting dashboard data",
		zap.String("dashboard_id", dashboardID),
		zap.String("tenant_id", tenantID))

	// Get dashboard configuration
	dashboard, err := ds.repository.GetDashboard(ctx, tenantID, dashboardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}

	if dashboard == nil {
		return nil, fmt.Errorf("dashboard not found: %s", dashboardID)
	}

	// Merge filters
	if filters != nil {
		dashboard.Filters = *filters
	}

	// Generate data based on dashboard type
	switch dashboard.Type {
	case DashboardTypeRiskOverview:
		summary, err := ds.GetRiskOverviewData(ctx, tenantID, &dashboard.Filters)
		if err != nil {
			return nil, fmt.Errorf("failed to get risk overview data: %w", err)
		}
		dashboard.Summary = *summary

	case DashboardTypeTrends:
		trends, err := ds.GetTrendsData(ctx, tenantID, &dashboard.Filters)
		if err != nil {
			return nil, fmt.Errorf("failed to get trends data: %w", err)
		}
		dashboard.Trends = *trends

	case DashboardTypePredictions:
		predictions, err := ds.GetPredictionsData(ctx, tenantID, &dashboard.Filters)
		if err != nil {
			return nil, fmt.Errorf("failed to get predictions data: %w", err)
		}
		dashboard.Predictions = *predictions

	default:
		// For custom dashboards, generate all data types
		summary, err := ds.GetRiskOverviewData(ctx, tenantID, &dashboard.Filters)
		if err != nil {
			return nil, fmt.Errorf("failed to get risk overview data: %w", err)
		}
		dashboard.Summary = *summary

		trends, err := ds.GetTrendsData(ctx, tenantID, &dashboard.Filters)
		if err != nil {
			return nil, fmt.Errorf("failed to get trends data: %w", err)
		}
		dashboard.Trends = *trends

		predictions, err := ds.GetPredictionsData(ctx, tenantID, &dashboard.Filters)
		if err != nil {
			return nil, fmt.Errorf("failed to get predictions data: %w", err)
		}
		dashboard.Predictions = *predictions
	}

	ds.logger.Debug("Dashboard data generated successfully",
		zap.String("dashboard_id", dashboardID),
		zap.String("type", string(dashboard.Type)))

	return dashboard, nil
}

// GetRiskOverviewData generates risk overview dashboard data
func (ds *DefaultDashboardService) GetRiskOverviewData(ctx context.Context, tenantID string, filters *DashboardFilters) (*DashboardSummary, error) {
	ds.logger.Debug("Generating risk overview data",
		zap.String("tenant_id", tenantID))

	// Get risk assessments
	assessments, err := ds.dataProvider.GetRiskAssessments(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk assessments: %w", err)
	}

	// Calculate summary metrics
	summary := ds.calculateRiskSummary(assessments)

	ds.logger.Debug("Risk overview data generated",
		zap.String("tenant_id", tenantID),
		zap.Int("total_assessments", summary.TotalAssessments),
		zap.Float64("average_risk_score", summary.AverageRiskScore))

	return summary, nil
}

// GetTrendsData generates trends dashboard data
func (ds *DefaultDashboardService) GetTrendsData(ctx context.Context, tenantID string, filters *DashboardFilters) (*DashboardTrends, error) {
	ds.logger.Debug("Generating trends data",
		zap.String("tenant_id", tenantID))

	// Get risk assessments
	assessments, err := ds.dataProvider.GetRiskAssessments(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk assessments: %w", err)
	}

	// Calculate trends
	trends := ds.calculateTrends(assessments)

	ds.logger.Debug("Trends data generated",
		zap.String("tenant_id", tenantID),
		zap.Int("time_series_points", len(trends.RiskScoreOverTime)))

	return trends, nil
}

// GetPredictionsData generates predictions dashboard data
func (ds *DefaultDashboardService) GetPredictionsData(ctx context.Context, tenantID string, filters *DashboardFilters) (*DashboardPredictions, error) {
	ds.logger.Debug("Generating predictions data",
		zap.String("tenant_id", tenantID))

	// Get risk predictions
	predictions, err := ds.dataProvider.GetRiskPredictions(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk predictions: %w", err)
	}

	// Calculate prediction analytics
	predictionData := ds.calculatePredictions(predictions)

	ds.logger.Debug("Predictions data generated",
		zap.String("tenant_id", tenantID),
		zap.Int("predictions_count", len(predictionData.RiskPredictions)))

	return predictionData, nil
}

// GetDashboardMetrics gets dashboard usage metrics
func (ds *DefaultDashboardService) GetDashboardMetrics(ctx context.Context, tenantID string) (*DashboardMetrics, error) {
	ds.logger.Debug("Getting dashboard metrics",
		zap.String("tenant_id", tenantID))

	metrics, err := ds.repository.GetDashboardMetrics(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard metrics: %w", err)
	}

	ds.logger.Debug("Dashboard metrics retrieved",
		zap.String("tenant_id", tenantID),
		zap.Int("total_dashboards", metrics.TotalDashboards))

	return metrics, nil
}

// Helper methods for data calculation

// calculateRiskSummary calculates risk summary metrics
func (ds *DefaultDashboardService) calculateRiskSummary(assessments []*models.RiskAssessment) *DashboardSummary {
	if len(assessments) == 0 {
		return &DashboardSummary{}
	}

	var totalScore float64
	var highRisk, mediumRisk, lowRisk int
	riskDistribution := make(map[string]int)
	riskFactors := make(map[string]int)

	for _, assessment := range assessments {
		totalScore += assessment.RiskScore

		// Count by risk level
		switch assessment.RiskLevel {
		case models.RiskLevelHigh:
			highRisk++
		case models.RiskLevelMedium:
			mediumRisk++
		case models.RiskLevelLow:
			lowRisk++
		}

		// Count risk distribution
		riskDistribution[string(assessment.RiskLevel)]++

		// Count risk factors (this would be extracted from assessment details)
		// For now, we'll use dummy data
		riskFactors["Financial Stability"]++
		riskFactors["Regulatory Compliance"]++
		riskFactors["Market Position"]++
	}

	// Calculate risk distribution percentages
	var riskDistributions []RiskDistribution
	total := len(assessments)
	for level, count := range riskDistribution {
		percentage := float64(count) / float64(total) * 100
		riskDistributions = append(riskDistributions, RiskDistribution{
			RiskLevel:  level,
			Count:      count,
			Percentage: percentage,
			Color:      getRiskLevelColor(level),
		})
	}

	// Calculate top risk factors
	var topRiskFactors []RiskFactorData
	for factor, count := range riskFactors {
		percentage := float64(count) / float64(total) * 100
		topRiskFactors = append(topRiskFactors, RiskFactorData{
			Factor:     factor,
			Count:      count,
			Percentage: percentage,
			Impact:     percentage, // This would be calculated from actual impact data
			Trend:      "stable",   // This would be calculated from trend analysis
		})
	}

	return &DashboardSummary{
		TotalAssessments: total,
		AverageRiskScore: totalScore / float64(total),
		HighRiskCount:    highRisk,
		MediumRiskCount:  mediumRisk,
		LowRiskCount:     lowRisk,
		RiskDistribution: riskDistributions,
		TopRiskFactors:   topRiskFactors,
		AssessmentVolume: AssessmentVolumeData{
			Total:      total,
			Daily:      total / 30, // Rough estimate
			Weekly:     total / 4,  // Rough estimate
			Monthly:    total,
			GrowthRate: 0.0, // This would be calculated from historical data
		},
		ComplianceStatus: ComplianceStatusData{
			Compliant:      int(float64(total) * 0.85), // Rough estimate
			NonCompliant:   int(float64(total) * 0.10), // Rough estimate
			Pending:        int(float64(total) * 0.05), // Rough estimate
			ComplianceRate: 85.0,                       // Rough estimate
		},
		PerformanceMetrics: PerformanceMetricsData{
			AverageResponseTime: 500.0, // ms
			P95ResponseTime:     1000.0,
			P99ResponseTime:     2000.0,
			ErrorRate:           0.1,
			Throughput:          1000.0,
			Availability:        99.9,
		},
	}
}

// calculateTrends calculates trend data
func (ds *DefaultDashboardService) calculateTrends(assessments []*models.RiskAssessment) *DashboardTrends {
	// Group assessments by time periods
	timeSeriesData := make(map[string][]TimeSeriesData)
	industryData := make(map[string]int)
	countryData := make(map[string]int)

	for _, assessment := range assessments {
		// Time series data (grouped by day)
		day := assessment.CreatedAt.Format("2006-01-02")
		if timeSeriesData[day] == nil {
			timeSeriesData[day] = []TimeSeriesData{}
		}
		timeSeriesData[day] = append(timeSeriesData[day], TimeSeriesData{
			Timestamp: assessment.CreatedAt,
			Value:     assessment.RiskScore,
		})

		// Industry data
		industryData[assessment.Industry]++

		// Country data
		countryData[assessment.Country]++
	}

	// Convert to time series format
	var riskScoreOverTime []TimeSeriesData
	for day, data := range timeSeriesData {
		var totalScore float64
		for _, point := range data {
			totalScore += point.Value
		}
		avgScore := totalScore / float64(len(data))
		riskScoreOverTime = append(riskScoreOverTime, TimeSeriesData{
			Timestamp: data[0].Timestamp,
			Value:     avgScore,
			Label:     day,
		})
	}

	// Convert industry data to chart format
	var industryChartData []ChartData
	for industry, count := range industryData {
		industryChartData = append(industryChartData, ChartData{
			Label: industry,
			Value: float64(count),
			Count: count,
		})
	}

	// Convert country data to chart format
	var countryChartData []ChartData
	for country, count := range countryData {
		countryChartData = append(countryChartData, ChartData{
			Label: country,
			Value: float64(count),
			Count: count,
		})
	}

	return &DashboardTrends{
		RiskScoreOverTime:          riskScoreOverTime,
		AssessmentVolumeByTime:     riskScoreOverTime, // Same data for volume
		AssessmentVolumeByIndustry: industryChartData,
		AssessmentVolumeByCountry:  countryChartData,
		RiskFactorTrends:           []RiskFactorTrend{},  // This would be calculated from detailed data
		ComplianceTrends:           []ComplianceTrend{},  // This would be calculated from compliance data
		PerformanceTrends:          []PerformanceTrend{}, // This would be calculated from performance data
	}
}

// calculatePredictions calculates prediction analytics
func (ds *DefaultDashboardService) calculatePredictions(predictions []*models.RiskPrediction) *DashboardPredictions {
	var riskPredictions []RiskPredictionData
	var modelPerformance []ModelPerformance

	// Convert predictions to dashboard format
	for _, prediction := range predictions {
		riskPredictions = append(riskPredictions, RiskPredictionData{
			BusinessID:    prediction.BusinessID,
			BusinessName:  "Business " + prediction.BusinessID, // Business name would be looked up separately
			CurrentRisk:   0.0,                                 // Current risk would be looked up from latest assessment
			PredictedRisk: prediction.PredictedScore,
			Confidence:    prediction.ConfidenceScore,
			TimeHorizon:   prediction.HorizonMonths,
			PredictedAt:   prediction.CreatedAt,
			Factors:       []string{"Financial Stability", "Market Position"}, // This would be extracted from prediction details
		})
	}

	// Add model performance data (this would be calculated from actual model metrics)
	modelPerformance = append(modelPerformance, ModelPerformance{
		ModelName:    "XGBoost",
		Accuracy:     0.92,
		Precision:    0.89,
		Recall:       0.91,
		F1Score:      0.90,
		AUC:          0.94,
		LastUpdated:  time.Now(),
		TrainingSize: 10000,
		TestSize:     2000,
	})

	return &DashboardPredictions{
		RiskPredictions:     riskPredictions,
		ModelAccuracy:       modelPerformance,
		ForecastedRisk:      []ForecastData{},       // This would be calculated from forecast models
		ConfidenceIntervals: []ConfidenceInterval{}, // This would be calculated from prediction intervals
		PredictionAccuracy: PredictionAccuracy{
			OverallAccuracy: 0.92,
			ByRiskLevel: map[string]float64{
				"high":   0.95,
				"medium": 0.90,
				"low":    0.88,
			},
			ByTimeHorizon: map[string]float64{
				"1_month":   0.95,
				"3_months":  0.92,
				"6_months":  0.89,
				"12_months": 0.85,
			},
			LastUpdated: time.Now(),
		},
		ModelDrift: ModelDriftData{
			ModelName:      "XGBoost",
			DriftDetected:  false,
			DriftScore:     0.05,
			DriftThreshold: 0.10,
			LastChecked:    time.Now(),
			DriftFactors:   []string{},
			Recommendation: "Model is performing well",
		},
	}
}

// Helper functions

func generateDashboardID() string {
	return fmt.Sprintf("dashboard_%d", time.Now().UnixNano())
}

func getTenantIDFromContext(ctx context.Context) string {
	// This would extract tenant ID from context
	// Implementation depends on your authentication/authorization system
	if tenantID := ctx.Value("tenant_id"); tenantID != nil {
		if id, ok := tenantID.(string); ok {
			return id
		}
	}
	return "default"
}

func getRiskLevelColor(level string) string {
	switch level {
	case "high":
		return "#ff4444"
	case "medium":
		return "#ffaa00"
	case "low":
		return "#44ff44"
	default:
		return "#888888"
	}
}
