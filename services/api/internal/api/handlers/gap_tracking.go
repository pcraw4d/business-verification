package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// GapTrackingHandler handles gap tracking operations
type GapTrackingHandler struct {
	trackingData []GapTracking
}

// GapTracking represents a tracked compliance gap
type GapTracking struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Priority    string                 `json:"priority"`
	Status      string                 `json:"status"`
	Progress    int                    `json:"progress"`
	AssignedTo  string                 `json:"assigned_to"`
	DueDate     time.Time              `json:"due_date"`
	StartDate   time.Time              `json:"start_date"`
	Framework   string                 `json:"framework"`
	Milestones  []Milestone            `json:"milestones"`
	Team        []string               `json:"team"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Comments    []Comment              `json:"comments"`
	Attachments []Attachment           `json:"attachments"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Milestone represents a project milestone
type Milestone struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
	DueDate     time.Time  `json:"due_date"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	AssignedTo  string     `json:"assigned_to"`
}

// Comment represents a comment on a gap
type Comment struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Attachment represents a file attachment
type Attachment struct {
	ID          string    `json:"id"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	FileType    string    `json:"file_type"`
	UploadedBy  string    `json:"uploaded_by"`
	UploadedAt  time.Time `json:"uploaded_at"`
	DownloadURL string    `json:"download_url"`
}

// TrackingMetrics represents tracking system metrics
type TrackingMetrics struct {
	TotalGaps       int     `json:"total_gaps"`
	InProgressGaps  int     `json:"in_progress_gaps"`
	CompletedGaps   int     `json:"completed_gaps"`
	OverdueGaps     int     `json:"overdue_gaps"`
	AverageProgress float64 `json:"average_progress"`
	CriticalGaps    int     `json:"critical_gaps"`
	HighGaps        int     `json:"high_gaps"`
	MediumGaps      int     `json:"medium_gaps"`
	LowGaps         int     `json:"low_gaps"`
}

// NewGapTrackingHandler creates a new gap tracking handler
func NewGapTrackingHandler() *GapTrackingHandler {
	return &GapTrackingHandler{
		trackingData: getSampleTrackingData(),
	}
}

// GetTrackingMetrics returns overall tracking metrics
func (h *GapTrackingHandler) GetTrackingMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := h.calculateMetrics()

	response := map[string]interface{}{
		"metrics":      metrics,
		"generated_at": time.Now().Format(time.RFC3339),
		"period":       "current",
	}

	json.NewEncoder(w).Encode(response)
}

// GetGapTrackingList returns a list of tracked gaps
func (h *GapTrackingHandler) GetGapTrackingList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	status := r.URL.Query().Get("status")
	priority := r.URL.Query().Get("priority")
	framework := r.URL.Query().Get("framework")
	assignedTo := r.URL.Query().Get("assigned_to")
	overdue := r.URL.Query().Get("overdue") == "true"

	// Filter gaps
	filteredGaps := h.filterGaps(status, priority, framework, assignedTo, overdue)

	response := map[string]interface{}{
		"gaps":        filteredGaps,
		"total_count": len(filteredGaps),
		"filters": map[string]interface{}{
			"status":      status,
			"priority":    priority,
			"framework":   framework,
			"assigned_to": assignedTo,
			"overdue":     overdue,
		},
		"generated_at": time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// GetGapTrackingDetails returns detailed information for a specific gap
func (h *GapTrackingHandler) GetGapTrackingDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gapID := vars["id"]

	w.Header().Set("Content-Type", "application/json")

	gap := h.getGapByID(gapID)
	if gap == nil {
		http.Error(w, "Gap not found", http.StatusNotFound)
		return
	}

	// Get related data
	relatedGaps := h.getRelatedGaps(gapID)
	progressHistory := h.getProgressHistory(gapID)
	teamPerformance := h.getTeamPerformance(gap.Team)

	response := map[string]interface{}{
		"gap":              gap,
		"related_gaps":     relatedGaps,
		"progress_history": progressHistory,
		"team_performance": teamPerformance,
		"risk_assessment":  h.assessGapRisk(gap),
		"recommendations":  h.getGapRecommendations(gapID),
	}

	json.NewEncoder(w).Encode(response)
}

// UpdateGapProgress updates the progress of a specific gap
func (h *GapTrackingHandler) UpdateGapProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gapID := vars["id"]

	var updateRequest struct {
		Progress   int                    `json:"progress"`
		Status     string                 `json:"status,omitempty"`
		Comment    string                 `json:"comment,omitempty"`
		Milestones []string               `json:"milestones,omitempty"`
		Metadata   map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Find and update gap
	gap := h.getGapByID(gapID)
	if gap == nil {
		http.Error(w, "Gap not found", http.StatusNotFound)
		return
	}

	// Update gap
	gap.Progress = updateRequest.Progress
	if updateRequest.Status != "" {
		gap.Status = updateRequest.Status
	}
	gap.UpdatedAt = time.Now()

	// Add comment if provided
	if updateRequest.Comment != "" {
		comment := Comment{
			ID:        fmt.Sprintf("comment_%d", time.Now().Unix()),
			Content:   updateRequest.Comment,
			Author:    "Current User", // In real implementation, get from auth context
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		gap.Comments = append(gap.Comments, comment)
	}

	// Update milestones if provided
	for _, milestoneID := range updateRequest.Milestones {
		for i, milestone := range gap.Milestones {
			if milestone.ID == milestoneID {
				gap.Milestones[i].Completed = true
				now := time.Now()
				gap.Milestones[i].CompletedAt = &now
				break
			}
		}
	}

	// Update metadata
	if updateRequest.Metadata != nil {
		if gap.Metadata == nil {
			gap.Metadata = make(map[string]interface{})
		}
		for key, value := range updateRequest.Metadata {
			gap.Metadata[key] = value
		}
	}

	response := map[string]interface{}{
		"gap":            gap,
		"message":        "Gap progress updated successfully",
		"updated_at":     gap.UpdatedAt,
		"next_milestone": h.getNextMilestone(gap),
		"risk_level":     h.assessGapRisk(gap),
	}

	json.NewEncoder(w).Encode(response)
}

// CreateGapTracking creates a new gap tracking entry
func (h *GapTrackingHandler) CreateGapTracking(w http.ResponseWriter, r *http.Request) {
	var gapRequest struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Priority    string    `json:"priority"`
		Framework   string    `json:"framework"`
		AssignedTo  string    `json:"assigned_to"`
		DueDate     time.Time `json:"due_date"`
		StartDate   time.Time `json:"start_date"`
		Team        []string  `json:"team"`
		Milestones  []struct {
			Name        string    `json:"name"`
			Description string    `json:"description"`
			DueDate     time.Time `json:"due_date"`
			AssignedTo  string    `json:"assigned_to"`
		} `json:"milestones"`
	}

	if err := json.NewDecoder(r.Body).Decode(&gapRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Create new gap
	gap := GapTracking{
		ID:          fmt.Sprintf("gap_%d", time.Now().Unix()),
		Title:       gapRequest.Title,
		Description: gapRequest.Description,
		Priority:    gapRequest.Priority,
		Status:      "open",
		Progress:    0,
		AssignedTo:  gapRequest.AssignedTo,
		DueDate:     gapRequest.DueDate,
		StartDate:   gapRequest.StartDate,
		Framework:   gapRequest.Framework,
		Team:        gapRequest.Team,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Comments:    []Comment{},
		Attachments: []Attachment{},
		Metadata:    make(map[string]interface{}),
	}

	// Add milestones
	for _, milestoneReq := range gapRequest.Milestones {
		milestone := Milestone{
			ID:          fmt.Sprintf("milestone_%d", time.Now().UnixNano()),
			Name:        milestoneReq.Name,
			Description: milestoneReq.Description,
			Completed:   false,
			DueDate:     milestoneReq.DueDate,
			AssignedTo:  milestoneReq.AssignedTo,
		}
		gap.Milestones = append(gap.Milestones, milestone)
	}

	// Add to tracking data
	h.trackingData = append(h.trackingData, gap)

	response := map[string]interface{}{
		"gap":                  gap,
		"message":              "Gap tracking created successfully",
		"created_at":           gap.CreatedAt,
		"estimated_completion": h.estimateCompletion(gap),
		"risk_assessment":      h.assessGapRisk(&gap),
	}

	json.NewEncoder(w).Encode(response)
}

// GetProgressHistory returns progress history for a gap
func (h *GapTrackingHandler) GetProgressHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gapID := vars["id"]

	w.Header().Set("Content-Type", "application/json")

	history := h.getProgressHistory(gapID)

	response := map[string]interface{}{
		"gap_id":        gapID,
		"history":       history,
		"total_entries": len(history),
		"generated_at":  time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// GetTeamPerformance returns performance metrics for a team
func (h *GapTrackingHandler) GetTeamPerformance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamName := vars["team"]

	w.Header().Set("Content-Type", "application/json")

	performance := h.getTeamPerformance([]string{teamName})

	response := map[string]interface{}{
		"team":         teamName,
		"performance":  performance,
		"generated_at": time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// ExportTrackingReport exports a comprehensive tracking report
func (h *GapTrackingHandler) ExportTrackingReport(w http.ResponseWriter, r *http.Request) {
	// Set response headers for file download
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=gap_tracking_report.json")

	// Generate comprehensive report
	metrics := h.calculateMetrics()
	report := map[string]interface{}{
		"report_date":      time.Now().Format("2006-01-02"),
		"metrics":          metrics,
		"gaps":             h.trackingData,
		"summary":          h.generateTrackingSummary(),
		"recommendations":  h.getSystemRecommendations(),
		"team_performance": h.getAllTeamPerformance(),
	}

	json.NewEncoder(w).Encode(report)
}

// Helper methods

func (h *GapTrackingHandler) calculateMetrics() TrackingMetrics {
	totalGaps := len(h.trackingData)
	inProgressGaps := 0
	completedGaps := 0
	overdueGaps := 0
	criticalGaps := 0
	highGaps := 0
	mediumGaps := 0
	lowGaps := 0
	totalProgress := 0

	for _, gap := range h.trackingData {
		switch gap.Status {
		case "in-progress":
			inProgressGaps++
		case "completed":
			completedGaps++
		}

		if gap.DueDate.Before(time.Now()) && gap.Status != "completed" {
			overdueGaps++
		}

		switch gap.Priority {
		case "critical":
			criticalGaps++
		case "high":
			highGaps++
		case "medium":
			mediumGaps++
		case "low":
			lowGaps++
		}

		totalProgress += gap.Progress
	}

	averageProgress := 0.0
	if totalGaps > 0 {
		averageProgress = float64(totalProgress) / float64(totalGaps)
	}

	return TrackingMetrics{
		TotalGaps:       totalGaps,
		InProgressGaps:  inProgressGaps,
		CompletedGaps:   completedGaps,
		OverdueGaps:     overdueGaps,
		AverageProgress: averageProgress,
		CriticalGaps:    criticalGaps,
		HighGaps:        highGaps,
		MediumGaps:      mediumGaps,
		LowGaps:         lowGaps,
	}
}

func (h *GapTrackingHandler) filterGaps(status, priority, framework, assignedTo string, overdue bool) []GapTracking {
	var filtered []GapTracking

	for _, gap := range h.trackingData {
		// Status filter
		if status != "" && gap.Status != status {
			continue
		}

		// Priority filter
		if priority != "" && gap.Priority != priority {
			continue
		}

		// Framework filter
		if framework != "" && gap.Framework != framework {
			continue
		}

		// Assigned to filter
		if assignedTo != "" && gap.AssignedTo != assignedTo {
			continue
		}

		// Overdue filter
		if overdue && (gap.DueDate.After(time.Now()) || gap.Status == "completed") {
			continue
		}

		filtered = append(filtered, gap)
	}

	return filtered
}

func (h *GapTrackingHandler) getGapByID(id string) *GapTracking {
	for i, gap := range h.trackingData {
		if gap.ID == id {
			return &h.trackingData[i]
		}
	}
	return nil
}

func (h *GapTrackingHandler) getRelatedGaps(gapID string) []GapTracking {
	// Return gaps with same framework or team members
	gap := h.getGapByID(gapID)
	if gap == nil {
		return []GapTracking{}
	}

	var related []GapTracking
	for _, g := range h.trackingData {
		if g.ID != gapID && (g.Framework == gap.Framework || hasCommonTeamMember(g.Team, gap.Team)) {
			related = append(related, g)
		}
	}

	return related
}

func (h *GapTrackingHandler) getProgressHistory(gapID string) []map[string]interface{} {
	// Simulate progress history
	return []map[string]interface{}{
		{
			"date":     time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
			"progress": 25,
			"comment":  "Initial setup completed",
			"author":   "John Smith",
		},
		{
			"date":     time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
			"progress": 45,
			"comment":  "Core implementation in progress",
			"author":   "Sarah Johnson",
		},
		{
			"date":     time.Now().AddDate(0, 0, -3).Format("2006-01-02"),
			"progress": 65,
			"comment":  "Testing phase started",
			"author":   "Mike Chen",
		},
	}
}

func (h *GapTrackingHandler) getTeamPerformance(team []string) map[string]interface{} {
	// Calculate team performance metrics
	teamGaps := 0
	totalProgress := 0
	completedGaps := 0

	for _, gap := range h.trackingData {
		if hasCommonTeamMember(gap.Team, team) {
			teamGaps++
			totalProgress += gap.Progress
			if gap.Status == "completed" {
				completedGaps++
			}
		}
	}

	averageProgress := 0.0
	if teamGaps > 0 {
		averageProgress = float64(totalProgress) / float64(teamGaps)
	}

	return map[string]interface{}{
		"total_gaps":       teamGaps,
		"completed_gaps":   completedGaps,
		"average_progress": averageProgress,
		"completion_rate":  float64(completedGaps) / float64(teamGaps) * 100,
	}
}

func (h *GapTrackingHandler) assessGapRisk(gap *GapTracking) map[string]interface{} {
	riskLevel := "low"
	riskScore := 0

	// Calculate risk based on various factors
	if gap.Priority == "critical" {
		riskScore += 40
	} else if gap.Priority == "high" {
		riskScore += 30
	} else if gap.Priority == "medium" {
		riskScore += 20
	} else {
		riskScore += 10
	}

	// Check if overdue
	if gap.DueDate.Before(time.Now()) && gap.Status != "completed" {
		riskScore += 30
	}

	// Check progress vs timeline
	daysElapsed := int(time.Since(gap.StartDate).Hours() / 24)
	daysTotal := int(gap.DueDate.Sub(gap.StartDate).Hours() / 24)
	if daysTotal > 0 {
		expectedProgress := (float64(daysElapsed) / float64(daysTotal)) * 100
		if gap.Progress < expectedProgress-20 {
			riskScore += 20
		}
	}

	// Determine risk level
	if riskScore >= 70 {
		riskLevel = "high"
	} else if riskScore >= 40 {
		riskLevel = "medium"
	}

	return map[string]interface{}{
		"risk_level": riskLevel,
		"risk_score": riskScore,
		"factors": []string{
			"Priority level",
			"Timeline adherence",
			"Progress rate",
		},
	}
}

func (h *GapTrackingHandler) getGapRecommendations(gapID string) []string {
	return []string{
		"Monitor progress more frequently",
		"Consider additional resources if behind schedule",
		"Review and update timeline if needed",
		"Ensure all team members are aligned on objectives",
	}
}

func (h *GapTrackingHandler) getNextMilestone(gap *GapTracking) *Milestone {
	for _, milestone := range gap.Milestones {
		if !milestone.Completed {
			return &milestone
		}
	}
	return nil
}

func (h *GapTrackingHandler) estimateCompletion(gap GapTracking) time.Time {
	// Simple estimation based on current progress and timeline
	if gap.Progress == 0 {
		return gap.DueDate
	}

	daysElapsed := int(time.Since(gap.StartDate).Hours() / 24)
	daysRemaining := int(gap.DueDate.Sub(gap.StartDate).Hours()/24) - daysElapsed

	if gap.Progress > 0 {
		estimatedDaysRemaining := int(float64(daysRemaining) * (100.0 - float64(gap.Progress)) / float64(gap.Progress))
		return time.Now().AddDate(0, 0, estimatedDaysRemaining)
	}

	return gap.DueDate
}

func (h *GapTrackingHandler) generateTrackingSummary() map[string]interface{} {
	metrics := h.calculateMetrics()

	return map[string]interface{}{
		"overall_status": "on_track",
		"key_achievements": []string{
			fmt.Sprintf("%d gaps completed successfully", metrics.CompletedGaps),
			fmt.Sprintf("%.1f%% average progress across all gaps", metrics.AverageProgress),
		},
		"areas_of_concern": []string{
			fmt.Sprintf("%d gaps are overdue", metrics.OverdueGaps),
			fmt.Sprintf("%d critical gaps require immediate attention", metrics.CriticalGaps),
		},
		"recommendations": []string{
			"Focus on completing overdue gaps",
			"Increase monitoring frequency for critical gaps",
			"Consider resource reallocation for high-priority items",
		},
	}
}

func (h *GapTrackingHandler) getSystemRecommendations() []string {
	return []string{
		"Implement automated progress tracking",
		"Set up proactive alerting for overdue items",
		"Create standardized milestone templates",
		"Establish regular team performance reviews",
		"Integrate with project management tools",
	}
}

func (h *GapTrackingHandler) getAllTeamPerformance() map[string]interface{} {
	teams := make(map[string][]string)

	// Group gaps by team
	for _, gap := range h.trackingData {
		for _, teamMember := range gap.Team {
			if _, exists := teams[teamMember]; !exists {
				teams[teamMember] = []string{}
			}
			teams[teamMember] = append(teams[teamMember], gap.ID)
		}
	}

	performance := make(map[string]interface{})
	for team, gapIDs := range teams {
		performance[team] = h.getTeamPerformance([]string{team})
	}

	return performance
}

// Utility functions

func hasCommonTeamMember(team1, team2 []string) bool {
	for _, member1 := range team1 {
		for _, member2 := range team2 {
			if member1 == member2 {
				return true
			}
		}
	}
	return false
}

func getSampleTrackingData() []GapTracking {
	now := time.Now()
	return []GapTracking{
		{
			ID:          "gap-001",
			Title:       "Multi-Factor Authentication Implementation",
			Description: "Deploy MFA across all critical systems and user accounts",
			Priority:    "critical",
			Status:      "in-progress",
			Progress:    65,
			AssignedTo:  "Security Team",
			DueDate:     now.AddDate(0, 0, 27),
			StartDate:   now.AddDate(0, 0, -18),
			Framework:   "SOC 2",
			Milestones: []Milestone{
				{ID: "mil-001", Name: "System Evaluation", Completed: true, DueDate: now.AddDate(0, 0, -13), CompletedAt: &now},
				{ID: "mil-002", Name: "Vendor Selection", Completed: true, DueDate: now.AddDate(0, 0, -6), CompletedAt: &now},
				{ID: "mil-003", Name: "Pilot Implementation", Completed: false, DueDate: now.AddDate(0, 0, 6)},
				{ID: "mil-004", Name: "Full Rollout", Completed: false, DueDate: now.AddDate(0, 0, 27)},
			},
			Team:        []string{"John Smith", "Sarah Johnson", "Mike Chen"},
			CreatedAt:   now.AddDate(0, 0, -18),
			UpdatedAt:   now,
			Comments:    []Comment{},
			Attachments: []Attachment{},
			Metadata:    make(map[string]interface{}),
		},
		{
			ID:          "gap-002",
			Title:       "Data Encryption Standards",
			Description: "Implement comprehensive data encryption for data at rest and in transit",
			Priority:    "high",
			Status:      "in-progress",
			Progress:    40,
			AssignedTo:  "DevOps Team",
			DueDate:     now.AddDate(0, 0, 42),
			StartDate:   now.AddDate(0, 0, -9),
			Framework:   "GDPR",
			Milestones: []Milestone{
				{ID: "mil-005", Name: "Encryption Audit", Completed: true, DueDate: now.AddDate(0, 0, -4), CompletedAt: &now},
				{ID: "mil-006", Name: "Standards Definition", Completed: true, DueDate: now.AddDate(0, 0, 1), CompletedAt: &now},
				{ID: "mil-007", Name: "Database Encryption", Completed: false, DueDate: now.AddDate(0, 0, 21)},
				{ID: "mil-008", Name: "TLS Configuration", Completed: false, DueDate: now.AddDate(0, 0, 36)},
			},
			Team:        []string{"Alex Rodriguez", "Emma Wilson", "David Kim"},
			CreatedAt:   now.AddDate(0, 0, -9),
			UpdatedAt:   now,
			Comments:    []Comment{},
			Attachments: []Attachment{},
			Metadata:    make(map[string]interface{}),
		},
		{
			ID:          "gap-003",
			Title:       "Security Awareness Training",
			Description: "Develop and implement comprehensive security awareness training",
			Priority:    "medium",
			Status:      "open",
			Progress:    15,
			AssignedTo:  "HR Team",
			DueDate:     now.AddDate(0, 0, 72),
			StartDate:   now.AddDate(0, 0, 1),
			Framework:   "PCI DSS",
			Milestones: []Milestone{
				{ID: "mil-009", Name: "Curriculum Design", Completed: false, DueDate: now.AddDate(0, 0, 12)},
				{ID: "mil-010", Name: "Content Creation", Completed: false, DueDate: now.AddDate(0, 0, 26)},
				{ID: "mil-011", Name: "Pilot Training", Completed: false, DueDate: now.AddDate(0, 0, 42)},
				{ID: "mil-012", Name: "Full Deployment", Completed: false, DueDate: now.AddDate(0, 0, 72)},
			},
			Team:        []string{"Lisa Brown", "Tom Anderson"},
			CreatedAt:   now.AddDate(0, 0, 1),
			UpdatedAt:   now,
			Comments:    []Comment{},
			Attachments: []Attachment{},
			Metadata:    make(map[string]interface{}),
		},
	}
}
