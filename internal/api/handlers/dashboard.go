package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/internal/risk"
)

// DashboardHandler handles dashboard-related API endpoints
type DashboardHandler struct {
	logger      *observability.Logger
	riskService *risk.RiskService
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(logger *observability.Logger, riskService *risk.RiskService) *DashboardHandler {
	return &DashboardHandler{
		logger:      logger,
		riskService: riskService,
	}
}

// DashboardOverview represents the main dashboard overview data
type DashboardOverview struct {
	TotalBusinesses    int                    `json:"total_businesses"`
	ActiveAlerts       int                    `json:"active_alerts"`
	CriticalAlerts     int                    `json:"critical_alerts"`
	HighRiskBusinesses int                    `json:"high_risk_businesses"`
	RecentAssessments  int                    `json:"recent_assessments"`
	AverageRiskScore   float64                `json:"average_risk_score"`
	RiskDistribution   map[string]int         `json:"risk_distribution"`
	RecentAlerts       []risk.RiskAlert       `json:"recent_alerts"`
	MonitoringStatus   *risk.MonitoringStatus `json:"monitoring_status"`
	LastUpdated        time.Time              `json:"last_updated"`
}

// DashboardBusiness represents business-specific dashboard data
type DashboardBusiness struct {
	BusinessID       string                 `json:"business_id"`
	BusinessName     string                 `json:"business_name"`
	OverallRiskScore float64                `json:"overall_risk_score"`
	OverallRiskLevel risk.RiskLevel         `json:"overall_risk_level"`
	CategoryScores   map[string]float64     `json:"category_scores"`
	RecentAlerts     []risk.RiskAlert       `json:"recent_alerts"`
	AutomatedAlerts  []risk.AutomatedAlert  `json:"automated_alerts"`
	RiskTrends       []risk.RiskTrend       `json:"risk_trends"`
	ThresholdAlerts  []risk.ThresholdAlert  `json:"threshold_alerts"`
	LastAssessment   *risk.RiskAssessment   `json:"last_assessment"`
	MonitoringStatus *risk.MonitoringStatus `json:"monitoring_status"`
	LastUpdated      time.Time              `json:"last_updated"`
}

// DashboardAnalytics represents analytics data for the dashboard
type DashboardAnalytics struct {
	RiskScoreDistribution map[string]int        `json:"risk_score_distribution"`
	CategoryRiskAverages  map[string]float64    `json:"category_risk_averages"`
	AlertTrends           []AlertTrendData      `json:"alert_trends"`
	AssessmentTrends      []AssessmentTrendData `json:"assessment_trends"`
	TopRiskFactors        []RiskFactorData      `json:"top_risk_factors"`
	GeographicRiskData    []GeographicRiskData  `json:"geographic_risk_data"`
	IndustryRiskData      []IndustryRiskData    `json:"industry_risk_data"`
	TimeRange             string                `json:"time_range"`
	LastUpdated           time.Time             `json:"last_updated"`
}

// AlertTrendData represents alert trend data
type AlertTrendData struct {
	Date        time.Time `json:"date"`
	TotalAlerts int       `json:"total_alerts"`
	Critical    int       `json:"critical"`
	High        int       `json:"high"`
	Medium      int       `json:"medium"`
	Low         int       `json:"low"`
}

// AssessmentTrendData represents assessment trend data
type AssessmentTrendData struct {
	Date             time.Time `json:"date"`
	TotalAssessments int       `json:"total_assessments"`
	AverageScore     float64   `json:"average_score"`
	HighRiskCount    int       `json:"high_risk_count"`
	CriticalCount    int       `json:"critical_count"`
}

// RiskFactorData represents risk factor data
type RiskFactorData struct {
	FactorID     string         `json:"factor_id"`
	FactorName   string         `json:"factor_name"`
	Category     string         `json:"category"`
	AverageScore float64        `json:"average_score"`
	Occurrences  int            `json:"occurrences"`
	RiskLevel    risk.RiskLevel `json:"risk_level"`
}

// GeographicRiskData represents geographic risk data
type GeographicRiskData struct {
	Region        string  `json:"region"`
	Country       string  `json:"country"`
	State         string  `json:"state"`
	City          string  `json:"city"`
	BusinessCount int     `json:"business_count"`
	AverageRisk   float64 `json:"average_risk"`
	HighRiskCount int     `json:"high_risk_count"`
}

// IndustryRiskData represents industry risk data
type IndustryRiskData struct {
	IndustryCode  string  `json:"industry_code"`
	IndustryName  string  `json:"industry_name"`
	BusinessCount int     `json:"business_count"`
	AverageRisk   float64 `json:"average_risk"`
	HighRiskCount int     `json:"high_risk_count"`
	CriticalCount int     `json:"critical_count"`
}

// GetDashboardOverviewHandler handles GET /v1/dashboard/overview requests
func (h *DashboardHandler) GetDashboardOverviewHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Dashboard overview request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
	)

	// Parse query parameters
	timeRange := r.URL.Query().Get("time_range")
	if timeRange == "" {
		timeRange = "7d" // default to 7 days
	}

	// Get monitoring status
	monitoringStatus, err := h.riskService.GetMonitoringStatus(r.Context())
	if err != nil {
		h.logger.Error("Failed to get monitoring status",
			"request_id", requestID,
			"error", err.Error(),
		)
		// Don't fail the entire request if monitoring status fails
		monitoringStatus = &risk.MonitoringStatus{
			ActiveMonitors:   0,
			TotalAlerts:      0,
			CriticalAlerts:   0,
			WarningAlerts:    0,
			MonitoringHealth: "unknown",
			Uptime:           0,
		}
	}

	// Create mock overview data (in real implementation, this would query the database)
	overview := &DashboardOverview{
		TotalBusinesses:    1250,
		ActiveAlerts:       45,
		CriticalAlerts:     12,
		HighRiskBusinesses: 89,
		RecentAssessments:  156,
		AverageRiskScore:   67.3,
		RiskDistribution: map[string]int{
			"low":      234,
			"medium":   567,
			"high":     345,
			"critical": 104,
		},
		RecentAlerts: []risk.RiskAlert{
			{
				ID:           "alert_1",
				BusinessID:   "business_123",
				RiskFactor:   "financial_stability",
				Level:        risk.RiskLevelCritical,
				Message:      "Critical financial stability risk detected",
				Score:        85.0,
				Threshold:    80.0,
				TriggeredAt:  time.Now().Add(-2 * time.Hour),
				Acknowledged: false,
			},
			{
				ID:           "alert_2",
				BusinessID:   "business_456",
				RiskFactor:   "operational_efficiency",
				Level:        risk.RiskLevelHigh,
				Message:      "High operational efficiency risk detected",
				Score:        75.0,
				Threshold:    70.0,
				TriggeredAt:  time.Now().Add(-1 * time.Hour),
				Acknowledged: false,
			},
		},
		MonitoringStatus: monitoringStatus,
		LastUpdated:      time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(overview); err != nil {
		h.logger.Error("Failed to encode dashboard overview response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Dashboard overview request completed",
		"request_id", requestID,
		"time_range", timeRange,
		"duration_ms", duration.Milliseconds(),
		"status_code", http.StatusOK,
	)
}

// GetDashboardBusinessHandler handles GET /v1/dashboard/business/{businessID} requests
func (h *DashboardHandler) GetDashboardBusinessHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Dashboard business request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
	)

	// Extract business ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid business ID", http.StatusBadRequest)
		return
	}
	businessID := pathParts[4]

	// Parse query parameters
	includeHistory := r.URL.Query().Get("include_history") == "true"
	timeRange := r.URL.Query().Get("time_range")
	if timeRange == "" {
		timeRange = "30d" // default to 30 days
	}

	// Get business-specific data
	ctx := context.WithValue(r.Context(), "request_id", requestID)

	// Get recent alerts (mock data for now)
	alerts := []risk.RiskAlert{
		{
			ID:           "alert_1",
			BusinessID:   businessID,
			RiskFactor:   "financial_stability",
			Level:        risk.RiskLevelCritical,
			Message:      "Critical financial stability risk detected",
			Score:        85.0,
			Threshold:    80.0,
			TriggeredAt:  time.Now().Add(-2 * time.Hour),
			Acknowledged: false,
		},
		{
			ID:           "alert_2",
			BusinessID:   businessID,
			RiskFactor:   "operational_efficiency",
			Level:        risk.RiskLevelHigh,
			Message:      "High operational efficiency risk detected",
			Score:        75.0,
			Threshold:    70.0,
			TriggeredAt:  time.Now().Add(-1 * time.Hour),
			Acknowledged: false,
		},
	}

	// Get automated alert history
	automatedAlerts, err := h.riskService.GetAutomatedAlertHistory(ctx, businessID)
	if err != nil {
		h.logger.Error("Failed to get automated alert history",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		automatedAlerts = []risk.AutomatedAlert{}
	}

	// Get monitoring status
	monitoringStatus, err := h.riskService.GetMonitoringStatus(ctx)
	if err != nil {
		h.logger.Error("Failed to get monitoring status",
			"request_id", requestID,
			"error", err.Error(),
		)
		monitoringStatus = &risk.MonitoringStatus{
			ActiveMonitors:   0,
			TotalAlerts:      0,
			CriticalAlerts:   0,
			WarningAlerts:    0,
			MonitoringHealth: "unknown",
			Uptime:           0,
		}
	}

	// Create mock business dashboard data
	businessData := &DashboardBusiness{
		BusinessID:       businessID,
		BusinessName:     "Sample Business Corp",
		OverallRiskScore: 72.5,
		OverallRiskLevel: risk.RiskLevelHigh,
		CategoryScores: map[string]float64{
			"financial":     78.2,
			"operational":   65.4,
			"regulatory":    82.1,
			"reputational":  68.9,
			"cybersecurity": 75.6,
		},
		RecentAlerts:     alerts,
		AutomatedAlerts:  automatedAlerts,
		RiskTrends:       []risk.RiskTrend{},
		ThresholdAlerts:  []risk.ThresholdAlert{},
		LastAssessment:   nil,
		MonitoringStatus: monitoringStatus,
		LastUpdated:      time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(businessData); err != nil {
		h.logger.Error("Failed to encode dashboard business response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Dashboard business request completed",
		"request_id", requestID,
		"business_id", businessID,
		"include_history", includeHistory,
		"time_range", timeRange,
		"duration_ms", duration.Milliseconds(),
		"status_code", http.StatusOK,
	)
}

// GetDashboardAnalyticsHandler handles GET /v1/dashboard/analytics requests
func (h *DashboardHandler) GetDashboardAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Dashboard analytics request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
	)

	// Parse query parameters
	timeRange := r.URL.Query().Get("time_range")
	if timeRange == "" {
		timeRange = "30d" // default to 30 days
	}

	category := r.URL.Query().Get("category")
	region := r.URL.Query().Get("region")
	industry := r.URL.Query().Get("industry")

	// Create mock analytics data
	analytics := &DashboardAnalytics{
		RiskScoreDistribution: map[string]int{
			"0-20":   45,
			"21-40":  123,
			"41-60":  234,
			"61-80":  345,
			"81-100": 156,
		},
		CategoryRiskAverages: map[string]float64{
			"financial":     72.3,
			"operational":   68.7,
			"regulatory":    75.2,
			"reputational":  69.8,
			"cybersecurity": 73.1,
		},
		AlertTrends: []AlertTrendData{
			{
				Date:        time.Now().AddDate(0, 0, -6),
				TotalAlerts: 12,
				Critical:    3,
				High:        5,
				Medium:      3,
				Low:         1,
			},
			{
				Date:        time.Now().AddDate(0, 0, -5),
				TotalAlerts: 15,
				Critical:    4,
				High:        6,
				Medium:      4,
				Low:         1,
			},
			{
				Date:        time.Now().AddDate(0, 0, -4),
				TotalAlerts: 18,
				Critical:    5,
				High:        7,
				Medium:      4,
				Low:         2,
			},
			{
				Date:        time.Now().AddDate(0, 0, -3),
				TotalAlerts: 14,
				Critical:    3,
				High:        6,
				Medium:      4,
				Low:         1,
			},
			{
				Date:        time.Now().AddDate(0, 0, -2),
				TotalAlerts: 16,
				Critical:    4,
				High:        7,
				Medium:      3,
				Low:         2,
			},
			{
				Date:        time.Now().AddDate(0, 0, -1),
				TotalAlerts: 19,
				Critical:    5,
				High:        8,
				Medium:      4,
				Low:         2,
			},
			{
				Date:        time.Now(),
				TotalAlerts: 22,
				Critical:    6,
				High:        9,
				Medium:      5,
				Low:         2,
			},
		},
		AssessmentTrends: []AssessmentTrendData{
			{
				Date:             time.Now().AddDate(0, 0, -6),
				TotalAssessments: 45,
				AverageScore:     68.2,
				HighRiskCount:    12,
				CriticalCount:    3,
			},
			{
				Date:             time.Now().AddDate(0, 0, -5),
				TotalAssessments: 52,
				AverageScore:     69.1,
				HighRiskCount:    15,
				CriticalCount:    4,
			},
			{
				Date:             time.Now().AddDate(0, 0, -4),
				TotalAssessments: 48,
				AverageScore:     70.3,
				HighRiskCount:    14,
				CriticalCount:    5,
			},
			{
				Date:             time.Now().AddDate(0, 0, -3),
				TotalAssessments: 55,
				AverageScore:     71.2,
				HighRiskCount:    18,
				CriticalCount:    6,
			},
			{
				Date:             time.Now().AddDate(0, 0, -2),
				TotalAssessments: 51,
				AverageScore:     72.8,
				HighRiskCount:    16,
				CriticalCount:    7,
			},
			{
				Date:             time.Now().AddDate(0, 0, -1),
				TotalAssessments: 58,
				AverageScore:     73.5,
				HighRiskCount:    20,
				CriticalCount:    8,
			},
			{
				Date:             time.Now(),
				TotalAssessments: 62,
				AverageScore:     74.1,
				HighRiskCount:    22,
				CriticalCount:    9,
			},
		},
		TopRiskFactors: []RiskFactorData{
			{
				FactorID:     "financial_stability",
				FactorName:   "Financial Stability",
				Category:     "financial",
				AverageScore: 78.5,
				Occurrences:  234,
				RiskLevel:    risk.RiskLevelHigh,
			},
			{
				FactorID:     "operational_efficiency",
				FactorName:   "Operational Efficiency",
				Category:     "operational",
				AverageScore: 75.2,
				Occurrences:  198,
				RiskLevel:    risk.RiskLevelHigh,
			},
			{
				FactorID:     "regulatory_compliance",
				FactorName:   "Regulatory Compliance",
				Category:     "regulatory",
				AverageScore: 82.1,
				Occurrences:  156,
				RiskLevel:    risk.RiskLevelCritical,
			},
			{
				FactorID:     "cybersecurity_posture",
				FactorName:   "Cybersecurity Posture",
				Category:     "cybersecurity",
				AverageScore: 73.8,
				Occurrences:  145,
				RiskLevel:    risk.RiskLevelHigh,
			},
			{
				FactorID:     "reputation_management",
				FactorName:   "Reputation Management",
				Category:     "reputational",
				AverageScore: 69.4,
				Occurrences:  123,
				RiskLevel:    risk.RiskLevelMedium,
			},
		},
		GeographicRiskData: []GeographicRiskData{
			{
				Region:        "North America",
				Country:       "United States",
				State:         "California",
				City:          "San Francisco",
				BusinessCount: 234,
				AverageRisk:   72.3,
				HighRiskCount: 45,
			},
			{
				Region:        "North America",
				Country:       "United States",
				State:         "New York",
				City:          "New York City",
				BusinessCount: 198,
				AverageRisk:   75.1,
				HighRiskCount: 52,
			},
			{
				Region:        "Europe",
				Country:       "United Kingdom",
				State:         "England",
				City:          "London",
				BusinessCount: 156,
				AverageRisk:   68.7,
				HighRiskCount: 34,
			},
			{
				Region:        "Europe",
				Country:       "Germany",
				State:         "Berlin",
				City:          "Berlin",
				BusinessCount: 123,
				AverageRisk:   71.2,
				HighRiskCount: 28,
			},
		},
		IndustryRiskData: []IndustryRiskData{
			{
				IndustryCode:  "541511",
				IndustryName:  "Custom Computer Programming Services",
				BusinessCount: 89,
				AverageRisk:   73.5,
				HighRiskCount: 23,
				CriticalCount: 5,
			},
			{
				IndustryCode:  "522110",
				IndustryName:  "Commercial Banking",
				BusinessCount: 67,
				AverageRisk:   78.2,
				HighRiskCount: 31,
				CriticalCount: 8,
			},
			{
				IndustryCode:  "541214",
				IndustryName:  "Payroll Services",
				BusinessCount: 45,
				AverageRisk:   75.8,
				HighRiskCount: 18,
				CriticalCount: 6,
			},
			{
				IndustryCode:  "541519",
				IndustryName:  "Other Computer Related Services",
				BusinessCount: 78,
				AverageRisk:   70.1,
				HighRiskCount: 22,
				CriticalCount: 4,
			},
		},
		TimeRange:   timeRange,
		LastUpdated: time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(analytics); err != nil {
		h.logger.Error("Failed to encode dashboard analytics response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Dashboard analytics request completed",
		"request_id", requestID,
		"time_range", timeRange,
		"category", category,
		"region", region,
		"industry", industry,
		"duration_ms", duration.Milliseconds(),
		"status_code", http.StatusOK,
	)
}

// GetDashboardAlertsHandler handles GET /v1/dashboard/alerts requests
func (h *DashboardHandler) GetDashboardAlertsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Dashboard alerts request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
	)

	// Parse query parameters
	level := r.URL.Query().Get("level")
	acknowledged := r.URL.Query().Get("acknowledged")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	businessID := r.URL.Query().Get("business_id")

	limit := 50 // default limit
	offset := 0 // default offset

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Create mock alerts data
	alerts := []risk.RiskAlert{
		{
			ID:           "alert_1",
			BusinessID:   "business_123",
			RiskFactor:   "financial_stability",
			Level:        risk.RiskLevelCritical,
			Message:      "Critical financial stability risk detected",
			Score:        85.0,
			Threshold:    80.0,
			TriggeredAt:  time.Now().Add(-2 * time.Hour),
			Acknowledged: false,
		},
		{
			ID:           "alert_2",
			BusinessID:   "business_456",
			RiskFactor:   "operational_efficiency",
			Level:        risk.RiskLevelHigh,
			Message:      "High operational efficiency risk detected",
			Score:        75.0,
			Threshold:    70.0,
			TriggeredAt:  time.Now().Add(-1 * time.Hour),
			Acknowledged: false,
		},
		{
			ID:           "alert_3",
			BusinessID:   "business_789",
			RiskFactor:   "regulatory_compliance",
			Level:        risk.RiskLevelCritical,
			Message:      "Critical regulatory compliance risk detected",
			Score:        88.0,
			Threshold:    85.0,
			TriggeredAt:  time.Now().Add(-30 * time.Minute),
			Acknowledged: false,
		},
	}

	// Filter by level if specified
	if level != "" {
		var filteredAlerts []risk.RiskAlert
		for _, alert := range alerts {
			if string(alert.Level) == level {
				filteredAlerts = append(filteredAlerts, alert)
			}
		}
		alerts = filteredAlerts
	}

	// Filter by business ID if specified
	if businessID != "" {
		var filteredAlerts []risk.RiskAlert
		for _, alert := range alerts {
			if alert.BusinessID == businessID {
				filteredAlerts = append(filteredAlerts, alert)
			}
		}
		alerts = filteredAlerts
	}

	// Apply pagination
	if offset < len(alerts) {
		end := offset + limit
		if end > len(alerts) {
			end = len(alerts)
		}
		alerts = alerts[offset:end]
	} else {
		alerts = []risk.RiskAlert{}
	}

	// Create response
	response := map[string]interface{}{
		"alerts":       alerts,
		"total_count":  len(alerts),
		"limit":        limit,
		"offset":       offset,
		"has_more":     false,
		"last_updated": time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode dashboard alerts response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Dashboard alerts request completed",
		"request_id", requestID,
		"level", level,
		"acknowledged", acknowledged,
		"limit", limit,
		"offset", offset,
		"business_id", businessID,
		"alert_count", len(alerts),
		"duration_ms", duration.Milliseconds(),
		"status_code", http.StatusOK,
	)
}

// GetDashboardMonitoringHandler handles GET /v1/dashboard/monitoring requests
func (h *DashboardHandler) GetDashboardMonitoringHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Dashboard monitoring request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
	)

	// Get monitoring status
	monitoringStatus, err := h.riskService.GetMonitoringStatus(r.Context())
	if err != nil {
		h.logger.Error("Failed to get monitoring status",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to get monitoring status", http.StatusInternalServerError)
		return
	}

	// Get automated alert rules
	alertRules, err := h.riskService.GetAutomatedAlertRules(r.Context())
	if err != nil {
		h.logger.Error("Failed to get automated alert rules",
			"request_id", requestID,
			"error", err.Error(),
		)
		alertRules = []*risk.AutomatedAlertRule{}
	}

	// Create monitoring response
	response := map[string]interface{}{
		"monitoring_status": monitoringStatus,
		"alert_rules":       alertRules,
		"active_monitors":   monitoringStatus.ActiveMonitors,
		"system_health":     monitoringStatus.MonitoringHealth,
		"last_updated":      time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode dashboard monitoring response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Dashboard monitoring request completed",
		"request_id", requestID,
		"active_monitors", monitoringStatus.ActiveMonitors,
		"system_health", monitoringStatus.MonitoringHealth,
		"duration_ms", duration.Milliseconds(),
		"status_code", http.StatusOK,
	)
}

// GetDashboardThresholdsHandler handles GET /v1/dashboard/thresholds requests
func (h *DashboardHandler) GetDashboardThresholdsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Context().Value("request_id").(string)

	h.logger.Info("Dashboard thresholds request received",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
	)

	// Parse query parameters
	category := r.URL.Query().Get("category")

	// Get threshold configurations for different categories
	ctx := context.WithValue(r.Context(), "request_id", requestID)

	thresholds := make(map[string]*risk.ThresholdMonitoringConfig)

	categories := []risk.RiskCategory{
		risk.RiskCategoryFinancial,
		risk.RiskCategoryOperational,
		risk.RiskCategoryRegulatory,
		risk.RiskCategoryReputational,
		risk.RiskCategoryCybersecurity,
	}

	for _, cat := range categories {
		if category == "" || string(cat) == category {
			config, err := h.riskService.GetThresholdConfig(ctx, cat)
			if err != nil {
				h.logger.Error("Failed to get threshold config",
					"request_id", requestID,
					"category", cat,
					"error", err.Error(),
				)
				continue
			}
			thresholds[string(cat)] = config
		}
	}

	// Create response
	response := map[string]interface{}{
		"thresholds":   thresholds,
		"categories":   categories,
		"last_updated": time.Now(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode dashboard thresholds response",
			"request_id", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	h.logger.Info("Dashboard thresholds request completed",
		"request_id", requestID,
		"category", category,
		"threshold_count", len(thresholds),
		"duration_ms", duration.Milliseconds(),
		"status_code", http.StatusOK,
	)
}
