package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// CostOptimizationHandlers handles cost optimization related HTTP requests
type CostOptimizationHandlers struct {
	resourceMonitor ResourceMonitor
	logger          *zap.Logger
}

// ResourceMonitor interface for resource monitoring
type ResourceMonitor interface {
	GetCurrentMetrics() *ResourceMetrics
	GetRecommendations() []CostOptimizationRecommendation
	GetCostTrends(ctx context.Context, period time.Duration) ([]CostTrend, error)
	GetResourceUtilization() map[string]float64
	GetCostBreakdown() *ResourceCosts
	GetHighPriorityRecommendations() []CostOptimizationRecommendation
	GetRecommendationsByType(recType string) []CostOptimizationRecommendation
	UpdateRecommendationStatus(id, status string) error
	GetResourceAlerts() []ResourceAlert
}

// ResourceMetrics represents current resource metrics
type ResourceMetrics struct {
	Timestamp         time.Time      `json:"timestamp"`
	CPUUtilization    float64        `json:"cpu_utilization"`
	MemoryUtilization float64        `json:"memory_utilization"`
	DiskUtilization   float64        `json:"disk_utilization"`
	NetworkIO         NetworkMetrics `json:"network_io"`
	PodCount          int            `json:"pod_count"`
	RequestRate       float64        `json:"request_rate"`
	ResponseTime      float64        `json:"response_time"`
	ErrorRate         float64        `json:"error_rate"`
	CostPerHour       float64        `json:"cost_per_hour"`
	ResourceCosts     ResourceCosts  `json:"resource_costs"`
}

// NetworkMetrics represents network I/O metrics
type NetworkMetrics struct {
	BytesIn    float64 `json:"bytes_in"`
	BytesOut   float64 `json:"bytes_out"`
	PacketsIn  float64 `json:"packets_in"`
	PacketsOut float64 `json:"packets_out"`
}

// ResourceCosts represents cost breakdown by resource type
type ResourceCosts struct {
	ComputeCost  float64 `json:"compute_cost"`
	StorageCost  float64 `json:"storage_cost"`
	NetworkCost  float64 `json:"network_cost"`
	DatabaseCost float64 `json:"database_cost"`
	CacheCost    float64 `json:"cache_cost"`
	CDNCost      float64 `json:"cdn_cost"`
	TotalCost    float64 `json:"total_cost"`
}

// CostOptimizationRecommendation represents a cost optimization recommendation
type CostOptimizationRecommendation struct {
	ID               string    `json:"id"`
	Type             string    `json:"type"`
	Priority         string    `json:"priority"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	PotentialSavings float64   `json:"potential_savings"`
	Impact           string    `json:"impact"`
	Effort           string    `json:"effort"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CostTrend represents cost trend data
type CostTrend struct {
	Timestamp time.Time `json:"timestamp"`
	Cost      float64   `json:"cost"`
	Type      string    `json:"type"`
}

// ResourceAlert represents a resource utilization alert
type ResourceAlert struct {
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Timestamp time.Time `json:"timestamp"`
}

// NewCostOptimizationHandlers creates a new cost optimization handlers instance
func NewCostOptimizationHandlers(resourceMonitor ResourceMonitor, logger *zap.Logger) *CostOptimizationHandlers {
	return &CostOptimizationHandlers{
		resourceMonitor: resourceMonitor,
		logger:          logger,
	}
}

// GetCurrentMetrics returns current resource metrics
func (h *CostOptimizationHandlers) GetCurrentMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.resourceMonitor.GetCurrentMetrics()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		h.logger.Error("Failed to encode metrics", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetRecommendations returns cost optimization recommendations
func (h *CostOptimizationHandlers) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	recommendations := h.resourceMonitor.GetRecommendations()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(recommendations); err != nil {
		h.logger.Error("Failed to encode recommendations", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetHighPriorityRecommendations returns high priority recommendations
func (h *CostOptimizationHandlers) GetHighPriorityRecommendations(w http.ResponseWriter, r *http.Request) {
	recommendations := h.resourceMonitor.GetHighPriorityRecommendations()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(recommendations); err != nil {
		h.logger.Error("Failed to encode high priority recommendations", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetRecommendationsByType returns recommendations filtered by type
func (h *CostOptimizationHandlers) GetRecommendationsByType(w http.ResponseWriter, r *http.Request) {
	recType := r.URL.Query().Get("type")
	if recType == "" {
		http.Error(w, "Type parameter is required", http.StatusBadRequest)
		return
	}

	recommendations := h.resourceMonitor.GetRecommendationsByType(recType)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(recommendations); err != nil {
		h.logger.Error("Failed to encode recommendations by type", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetCostTrends returns cost trends for the specified period
func (h *CostOptimizationHandlers) GetCostTrends(w http.ResponseWriter, r *http.Request) {
	periodStr := r.URL.Query().Get("period")
	if periodStr == "" {
		periodStr = "24h" // Default to 24 hours
	}

	period, err := time.ParseDuration(periodStr)
	if err != nil {
		http.Error(w, "Invalid period format", http.StatusBadRequest)
		return
	}

	trends, err := h.resourceMonitor.GetCostTrends(r.Context(), period)
	if err != nil {
		h.logger.Error("Failed to get cost trends", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(trends); err != nil {
		h.logger.Error("Failed to encode cost trends", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetResourceUtilization returns current resource utilization
func (h *CostOptimizationHandlers) GetResourceUtilization(w http.ResponseWriter, r *http.Request) {
	utilization := h.resourceMonitor.GetResourceUtilization()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(utilization); err != nil {
		h.logger.Error("Failed to encode resource utilization", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetCostBreakdown returns current cost breakdown
func (h *CostOptimizationHandlers) GetCostBreakdown(w http.ResponseWriter, r *http.Request) {
	costBreakdown := h.resourceMonitor.GetCostBreakdown()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(costBreakdown); err != nil {
		h.logger.Error("Failed to encode cost breakdown", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetResourceAlerts returns current resource alerts
func (h *CostOptimizationHandlers) GetResourceAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := h.resourceMonitor.GetResourceAlerts()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(alerts); err != nil {
		h.logger.Error("Failed to encode resource alerts", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// UpdateRecommendationStatus updates the status of a recommendation
func (h *CostOptimizationHandlers) UpdateRecommendationStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.ID == "" || request.Status == "" {
		http.Error(w, "ID and status are required", http.StatusBadRequest)
		return
	}

	// Validate status
	validStatuses := []string{"pending", "approved", "rejected", "implemented", "archived"}
	valid := false
	for _, status := range validStatuses {
		if request.Status == status {
			valid = true
			break
		}
	}

	if !valid {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	if err := h.resourceMonitor.UpdateRecommendationStatus(request.ID, request.Status); err != nil {
		h.logger.Error("Failed to update recommendation status", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"message": "Recommendation status updated successfully",
		"id":      request.ID,
		"status":  request.Status,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetCostOptimizationDashboard returns a comprehensive cost optimization dashboard
func (h *CostOptimizationHandlers) GetCostOptimizationDashboard(w http.ResponseWriter, r *http.Request) {
	// Get all dashboard data
	metrics := h.resourceMonitor.GetCurrentMetrics()
	recommendations := h.resourceMonitor.GetRecommendations()
	costBreakdown := h.resourceMonitor.GetCostBreakdown()
	resourceUtilization := h.resourceMonitor.GetResourceUtilization()
	alerts := h.resourceMonitor.GetResourceAlerts()

	// Get cost trends for the last 24 hours
	trends, err := h.resourceMonitor.GetCostTrends(r.Context(), 24*time.Hour)
	if err != nil {
		h.logger.Error("Failed to get cost trends for dashboard", zap.Error(err))
		trends = []CostTrend{} // Use empty slice if trends fail
	}

	// Calculate summary statistics
	totalRecommendations := len(recommendations)
	highPriorityRecommendations := 0
	totalPotentialSavings := 0.0

	for _, rec := range recommendations {
		if rec.Priority == "high" || rec.Priority == "critical" {
			highPriorityRecommendations++
		}
		if rec.PotentialSavings > 0 {
			totalPotentialSavings += rec.PotentialSavings
		}
	}

	// Create dashboard response
	dashboard := map[string]interface{}{
		"timestamp": time.Now(),
		"summary": map[string]interface{}{
			"current_cost_per_hour":         metrics.CostPerHour,
			"total_recommendations":         totalRecommendations,
			"high_priority_recommendations": highPriorityRecommendations,
			"total_potential_savings":       totalPotentialSavings,
			"active_alerts":                 len(alerts),
		},
		"metrics":              metrics,
		"cost_breakdown":       costBreakdown,
		"resource_utilization": resourceUtilization,
		"recommendations":      recommendations,
		"cost_trends":          trends,
		"alerts":               alerts,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dashboard); err != nil {
		h.logger.Error("Failed to encode dashboard", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetCostOptimizationReport generates a detailed cost optimization report
func (h *CostOptimizationHandlers) GetCostOptimizationReport(w http.ResponseWriter, r *http.Request) {
	// Get report parameters
	periodStr := r.URL.Query().Get("period")
	if periodStr == "" {
		periodStr = "7d" // Default to 7 days
	}

	period, err := time.ParseDuration(periodStr)
	if err != nil {
		http.Error(w, "Invalid period format", http.StatusBadRequest)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json" // Default to JSON
	}

	// Get report data
	metrics := h.resourceMonitor.GetCurrentMetrics()
	recommendations := h.resourceMonitor.GetRecommendations()
	costBreakdown := h.resourceMonitor.GetCostBreakdown()
	trends, err := h.resourceMonitor.GetCostTrends(r.Context(), period)
	if err != nil {
		h.logger.Error("Failed to get cost trends for report", zap.Error(err))
		trends = []CostTrend{}
	}

	// Generate report
	report := map[string]interface{}{
		"generated_at": time.Now(),
		"period":       period.String(),
		"format":       format,
		"executive_summary": map[string]interface{}{
			"current_hourly_cost":    metrics.CostPerHour,
			"projected_daily_cost":   metrics.CostPerHour * 24,
			"projected_monthly_cost": metrics.CostPerHour * 24 * 30,
			"total_recommendations":  len(recommendations),
			"potential_savings":      calculateTotalSavings(recommendations),
		},
		"cost_analysis": map[string]interface{}{
			"breakdown":   costBreakdown,
			"trends":      trends,
			"utilization": h.resourceMonitor.GetResourceUtilization(),
		},
		"recommendations": map[string]interface{}{
			"all":                    recommendations,
			"by_priority":            groupRecommendationsByPriority(recommendations),
			"by_type":                groupRecommendationsByType(recommendations),
			"implementation_roadmap": generateImplementationRoadmap(recommendations),
		},
		"alerts": h.resourceMonitor.GetResourceAlerts(),
	}

	// Set appropriate content type based on format
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(report); err != nil {
			h.logger.Error("Failed to encode report", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=cost_optimization_report.csv")
		// TODO: Implement CSV export
		http.Error(w, "CSV format not implemented yet", http.StatusNotImplemented)
		return
	default:
		http.Error(w, "Unsupported format", http.StatusBadRequest)
		return
	}
}

// Helper functions

func calculateTotalSavings(recommendations []CostOptimizationRecommendation) float64 {
	total := 0.0
	for _, rec := range recommendations {
		if rec.PotentialSavings > 0 {
			total += rec.PotentialSavings
		}
	}
	return total
}

func groupRecommendationsByPriority(recommendations []CostOptimizationRecommendation) map[string][]CostOptimizationRecommendation {
	grouped := make(map[string][]CostOptimizationRecommendation)

	for _, rec := range recommendations {
		grouped[rec.Priority] = append(grouped[rec.Priority], rec)
	}

	return grouped
}

func groupRecommendationsByType(recommendations []CostOptimizationRecommendation) map[string][]CostOptimizationRecommendation {
	grouped := make(map[string][]CostOptimizationRecommendation)

	for _, rec := range recommendations {
		grouped[rec.Type] = append(grouped[rec.Type], rec)
	}

	return grouped
}

func generateImplementationRoadmap(recommendations []CostOptimizationRecommendation) []map[string]interface{} {
	var roadmap []map[string]interface{}

	// Sort recommendations by priority and potential savings
	priorityOrder := map[string]int{
		"critical": 1,
		"high":     2,
		"medium":   3,
		"low":      4,
	}

	// Group by priority and create roadmap phases
	phases := make(map[string][]CostOptimizationRecommendation)
	for _, rec := range recommendations {
		phases[rec.Priority] = append(phases[rec.Priority], rec)
	}

	// Create roadmap phases in priority order
	for priority := 1; priority <= 4; priority++ {
		for prio, recs := range phases {
			if priorityOrder[prio] == priority && len(recs) > 0 {
				phase := map[string]interface{}{
					"phase":             prio,
					"priority":          priority,
					"recommendations":   recs,
					"estimated_effort":  calculatePhaseEffort(recs),
					"potential_savings": calculatePhaseSavings(recs),
				}
				roadmap = append(roadmap, phase)
			}
		}
	}

	return roadmap
}

func calculatePhaseEffort(recommendations []CostOptimizationRecommendation) string {
	effortMap := map[string]int{
		"low":    1,
		"medium": 2,
		"high":   3,
	}

	maxEffort := 0
	for _, rec := range recommendations {
		if effortMap[rec.Effort] > maxEffort {
			maxEffort = effortMap[rec.Effort]
		}
	}

	switch maxEffort {
	case 1:
		return "low"
	case 2:
		return "medium"
	case 3:
		return "high"
	default:
		return "unknown"
	}
}

func calculatePhaseSavings(recommendations []CostOptimizationRecommendation) float64 {
	total := 0.0
	for _, rec := range recommendations {
		if rec.PotentialSavings > 0 {
			total += rec.PotentialSavings
		}
	}
	return total
}
