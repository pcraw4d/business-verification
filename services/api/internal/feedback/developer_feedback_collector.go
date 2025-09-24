package feedback

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// DeveloperFeedbackCollector handles collection and processing of developer feedback
// on technical implementation and code quality
type DeveloperFeedbackCollector struct {
	storage FeedbackStorage
	logger  *log.Logger
}

// DeveloperFeedback represents structured developer feedback data
type DeveloperFeedback struct {
	ID                     uuid.UUID               `json:"id"`
	DeveloperID            string                  `json:"developer_id"`
	Category               DeveloperCategory       `json:"category"`
	Rating                 int                     `json:"rating"` // 1-5 scale
	Comments               string                  `json:"comments"`
	TechnicalAreas         []string                `json:"technical_areas"`
	ImprovementSuggestions []string                `json:"improvement_suggestions"`
	CodeQualityRating      int                     `json:"code_quality_rating"`    // 1-5 scale
	ArchitectureRating     int                     `json:"architecture_rating"`    // 1-5 scale
	PerformanceRating      int                     `json:"performance_rating"`     // 1-5 scale
	MaintainabilityRating  int                     `json:"maintainability_rating"` // 1-5 scale
	TechnicalDebt          TechnicalDebtAssessment `json:"technical_debt"`
	SubmittedAt            time.Time               `json:"submitted_at"`
	Metadata               map[string]interface{}  `json:"metadata"`
}

// DeveloperCategory represents different types of developer feedback
type DeveloperCategory string

const (
	CategoryCodeQuality      DeveloperCategory = "code_quality"
	CategoryArchitecture     DeveloperCategory = "architecture"
	CategoryPerformance      DeveloperCategory = "performance"
	CategoryMaintainability  DeveloperCategory = "maintainability"
	CategoryTesting          DeveloperCategory = "testing"
	CategoryDocumentation    DeveloperCategory = "documentation"
	CategoryIntegration      DeveloperCategory = "integration"
	CategoryDeployment       DeveloperCategory = "deployment"
	CategorySecurity         DeveloperCategory = "security"
	CategoryOverallTechnical DeveloperCategory = "overall_technical"
)

// TechnicalDebtAssessment represents technical debt evaluation
type TechnicalDebtAssessment struct {
	OverallDebtLevel   string   `json:"overall_debt_level"` // low, medium, high, critical
	DebtAreas          []string `json:"debt_areas"`
	EstimatedEffort    string   `json:"estimated_effort"` // hours, days, weeks, months
	PriorityLevel      string   `json:"priority_level"`   // low, medium, high, critical
	ImpactAssessment   string   `json:"impact_assessment"`
	RecommendedActions []string `json:"recommended_actions"`
}

// DeveloperFeedbackStats provides aggregated developer feedback statistics
type DeveloperFeedbackStats struct {
	TotalResponses       int                  `json:"total_responses"`
	AverageRating        float64              `json:"average_rating"`
	CategoryBreakdown    map[string]int       `json:"category_breakdown"`
	RatingDistribution   map[int]int          `json:"rating_distribution"`
	CommonImprovements   []string             `json:"common_improvements"`
	TechnicalDebtSummary TechnicalDebtSummary `json:"technical_debt_summary"`
	CodeQualityMetrics   CodeQualityMetrics   `json:"code_quality_metrics"`
	ResponseRate         float64              `json:"response_rate"`
	LastUpdated          time.Time            `json:"last_updated"`
}

// TechnicalDebtSummary provides summary of technical debt feedback
type TechnicalDebtSummary struct {
	AverageDebtLevel     string         `json:"average_debt_level"`
	MostCommonDebtAreas  []string       `json:"most_common_debt_areas"`
	PriorityDistribution map[string]int `json:"priority_distribution"`
	EstimatedTotalEffort string         `json:"estimated_total_effort"`
}

// CodeQualityMetrics provides code quality assessment metrics
type CodeQualityMetrics struct {
	AverageCodeQuality     float64 `json:"average_code_quality"`
	AverageArchitecture    float64 `json:"average_architecture"`
	AveragePerformance     float64 `json:"average_performance"`
	AverageMaintainability float64 `json:"average_maintainability"`
	QualityTrend           string  `json:"quality_trend"` // improving, stable, declining
}

// NewDeveloperFeedbackCollector creates a new developer feedback collector
func NewDeveloperFeedbackCollector(storage FeedbackStorage, logger *log.Logger) *DeveloperFeedbackCollector {
	return &DeveloperFeedbackCollector{
		storage: storage,
		logger:  logger,
	}
}

// CollectDeveloperFeedback processes and stores developer feedback
func (dfc *DeveloperFeedbackCollector) CollectDeveloperFeedback(ctx context.Context, feedback *DeveloperFeedback) error {
	// Validate feedback data
	if err := dfc.validateDeveloperFeedback(feedback); err != nil {
		return fmt.Errorf("developer feedback validation failed: %w", err)
	}

	// Set metadata
	feedback.ID = uuid.New()
	feedback.SubmittedAt = time.Now()

	// Add system metadata
	feedback.Metadata = map[string]interface{}{
		"collector_version": "1.0.0",
		"collection_method": "developer_api",
		"timestamp":         feedback.SubmittedAt,
		"feedback_type":     "developer",
	}

	// Store feedback (using the same storage interface)
	userFeedback := dfc.convertToUserFeedback(feedback)
	if err := dfc.storage.StoreFeedback(ctx, userFeedback); err != nil {
		return fmt.Errorf("failed to store developer feedback: %w", err)
	}

	dfc.logger.Printf("Developer feedback collected successfully: ID=%s, Category=%s, Rating=%d",
		feedback.ID, feedback.Category, feedback.Rating)

	return nil
}

// validateDeveloperFeedback validates the developer feedback data
func (dfc *DeveloperFeedbackCollector) validateDeveloperFeedback(feedback *DeveloperFeedback) error {
	if feedback.DeveloperID == "" {
		return fmt.Errorf("developer ID is required")
	}

	if feedback.Category == "" {
		return fmt.Errorf("developer feedback category is required")
	}

	if feedback.Rating < 1 || feedback.Rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}

	if feedback.CodeQualityRating < 1 || feedback.CodeQualityRating > 5 {
		return fmt.Errorf("code quality rating must be between 1 and 5")
	}

	if feedback.ArchitectureRating < 1 || feedback.ArchitectureRating > 5 {
		return fmt.Errorf("architecture rating must be between 1 and 5")
	}

	if feedback.PerformanceRating < 1 || feedback.PerformanceRating > 5 {
		return fmt.Errorf("performance rating must be between 1 and 5")
	}

	if feedback.MaintainabilityRating < 1 || feedback.MaintainabilityRating > 5 {
		return fmt.Errorf("maintainability rating must be between 1 and 5")
	}

	return nil
}

// convertToUserFeedback converts developer feedback to user feedback format for storage
func (dfc *DeveloperFeedbackCollector) convertToUserFeedback(devFeedback *DeveloperFeedback) *UserFeedback {
	// Convert technical debt to business impact format
	businessImpact := BusinessImpactRating{
		TimeSaved:        dfc.calculateTimeSaved(devFeedback),
		CostReduction:    dfc.calculateCostReduction(devFeedback),
		ErrorReduction:   dfc.calculateErrorReduction(devFeedback),
		ProductivityGain: dfc.calculateProductivityGain(devFeedback),
		ROI:              dfc.calculateROI(devFeedback),
	}

	// Convert technical areas to specific features
	specificFeatures := append(devFeedback.TechnicalAreas,
		fmt.Sprintf("code_quality_%d", devFeedback.CodeQualityRating),
		fmt.Sprintf("architecture_%d", devFeedback.ArchitectureRating),
		fmt.Sprintf("performance_%d", devFeedback.PerformanceRating),
		fmt.Sprintf("maintainability_%d", devFeedback.MaintainabilityRating),
	)

	return &UserFeedback{
		ID:                     devFeedback.ID,
		UserID:                 devFeedback.DeveloperID,
		Category:               FeedbackCategory(devFeedback.Category),
		Rating:                 devFeedback.Rating,
		Comments:               devFeedback.Comments,
		SpecificFeatures:       specificFeatures,
		ImprovementAreas:       devFeedback.ImprovementSuggestions,
		ClassificationAccuracy: dfc.calculateOverallTechnicalScore(devFeedback),
		PerformanceRating:      devFeedback.PerformanceRating,
		UsabilityRating:        devFeedback.MaintainabilityRating,
		BusinessImpact:         businessImpact,
		SubmittedAt:            devFeedback.SubmittedAt,
		Metadata:               devFeedback.Metadata,
	}
}

// calculateTimeSaved estimates time saved from technical improvements
func (dfc *DeveloperFeedbackCollector) calculateTimeSaved(feedback *DeveloperFeedback) int {
	// Base calculation on performance and maintainability ratings
	baseTime := 30 // base 30 minutes
	performanceBonus := (feedback.PerformanceRating - 3) * 10
	maintainabilityBonus := (feedback.MaintainabilityRating - 3) * 15

	return baseTime + performanceBonus + maintainabilityBonus
}

// calculateCostReduction estimates cost reduction from technical improvements
func (dfc *DeveloperFeedbackCollector) calculateCostReduction(feedback *DeveloperFeedback) string {
	// Base calculation on overall rating and technical debt level
	baseReduction := 20 // base 20%
	ratingBonus := (feedback.Rating - 3) * 5

	// Adjust based on technical debt
	debtMultiplier := 1.0
	switch feedback.TechnicalDebt.OverallDebtLevel {
	case "low":
		debtMultiplier = 1.2
	case "medium":
		debtMultiplier = 1.0
	case "high":
		debtMultiplier = 0.8
	case "critical":
		debtMultiplier = 0.6
	}

	totalReduction := int(float64(baseReduction+ratingBonus) * debtMultiplier)
	if totalReduction < 0 {
		totalReduction = 0
	}
	if totalReduction > 80 {
		totalReduction = 80
	}

	return fmt.Sprintf("%d%%", totalReduction)
}

// calculateErrorReduction estimates error reduction from technical improvements
func (dfc *DeveloperFeedbackCollector) calculateErrorReduction(feedback *DeveloperFeedback) int {
	// Base calculation on code quality and architecture ratings
	baseReduction := 30 // base 30%
	qualityBonus := (feedback.CodeQualityRating - 3) * 10
	architectureBonus := (feedback.ArchitectureRating - 3) * 8

	totalReduction := baseReduction + qualityBonus + architectureBonus
	if totalReduction < 0 {
		totalReduction = 0
	}
	if totalReduction > 90 {
		totalReduction = 90
	}

	return totalReduction
}

// calculateProductivityGain estimates productivity gain from technical improvements
func (dfc *DeveloperFeedbackCollector) calculateProductivityGain(feedback *DeveloperFeedback) int {
	// Base calculation on maintainability and overall rating
	baseGain := 25 // base 25%
	maintainabilityBonus := (feedback.MaintainabilityRating - 3) * 12
	ratingBonus := (feedback.Rating - 3) * 8

	totalGain := baseGain + maintainabilityBonus + ratingBonus
	if totalGain < 0 {
		totalGain = 0
	}
	if totalGain > 100 {
		totalGain = 100
	}

	return totalGain
}

// calculateROI estimates return on investment from technical improvements
func (dfc *DeveloperFeedbackCollector) calculateROI(feedback *DeveloperFeedback) string {
	// Calculate based on multiple factors
	score := (feedback.Rating + feedback.CodeQualityRating + feedback.ArchitectureRating +
		feedback.PerformanceRating + feedback.MaintainabilityRating) / 5.0

	// Adjust based on technical debt
	debtAdjustment := 1.0
	switch feedback.TechnicalDebt.OverallDebtLevel {
	case "low":
		debtAdjustment = 1.3
	case "medium":
		debtAdjustment = 1.0
	case "high":
		debtAdjustment = 0.7
	case "critical":
		debtAdjustment = 0.4
	}

	adjustedScore := score * debtAdjustment

	if adjustedScore >= 4.5 {
		return "Exceptional"
	} else if adjustedScore >= 4.0 {
		return "Very High"
	} else if adjustedScore >= 3.5 {
		return "High"
	} else if adjustedScore >= 3.0 {
		return "Medium"
	} else {
		return "Low"
	}
}

// calculateOverallTechnicalScore calculates overall technical score
func (dfc *DeveloperFeedbackCollector) calculateOverallTechnicalScore(feedback *DeveloperFeedback) float64 {
	// Weighted average of all technical ratings
	totalScore := float64(feedback.Rating + feedback.CodeQualityRating +
		feedback.ArchitectureRating + feedback.PerformanceRating +
		feedback.MaintainabilityRating)

	return totalScore / 25.0 // Normalize to 0-1 scale
}

// GetDeveloperFeedbackAnalysis retrieves and analyzes developer feedback data
func (dfc *DeveloperFeedbackCollector) GetDeveloperFeedbackAnalysis(ctx context.Context, category DeveloperCategory) (*DeveloperFeedbackAnalysis, error) {
	// Get feedback for specific category
	userFeedback, err := dfc.storage.GetFeedbackByCategory(ctx, string(category))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve developer feedback: %w", err)
	}

	// Convert back to developer feedback format
	developerFeedback := dfc.convertFromUserFeedback(userFeedback)

	// Analyze feedback
	analysis := dfc.analyzeDeveloperFeedback(developerFeedback)

	return analysis, nil
}

// DeveloperFeedbackAnalysis provides detailed analysis of developer feedback data
type DeveloperFeedbackAnalysis struct {
	Category               DeveloperCategory    `json:"category"`
	TotalResponses         int                  `json:"total_responses"`
	AverageRating          float64              `json:"average_rating"`
	AverageCodeQuality     float64              `json:"average_code_quality"`
	AverageArchitecture    float64              `json:"average_architecture"`
	AveragePerformance     float64              `json:"average_performance"`
	AverageMaintainability float64              `json:"average_maintainability"`
	TopImprovements        []string             `json:"top_improvements"`
	TopTechnicalAreas      []string             `json:"top_technical_areas"`
	TechnicalDebtSummary   TechnicalDebtSummary `json:"technical_debt_summary"`
	CodeQualityMetrics     CodeQualityMetrics   `json:"code_quality_metrics"`
	Recommendations        []string             `json:"recommendations"`
	GeneratedAt            time.Time            `json:"generated_at"`
}

// convertFromUserFeedback converts user feedback back to developer feedback format
func (dfc *DeveloperFeedbackCollector) convertFromUserFeedback(userFeedback []*UserFeedback) []*DeveloperFeedback {
	var developerFeedback []*DeveloperFeedback

	for _, uf := range userFeedback {
		// Extract technical ratings from specific features
		codeQuality, architecture, performance, maintainability := dfc.extractTechnicalRatings(uf.SpecificFeatures)

		// Extract technical debt information from business impact
		technicalDebt := dfc.extractTechnicalDebt(uf.BusinessImpact, uf.ImprovementAreas)

		devFeedback := &DeveloperFeedback{
			ID:                     uf.ID,
			DeveloperID:            uf.UserID,
			Category:               DeveloperCategory(uf.Category),
			Rating:                 uf.Rating,
			Comments:               uf.Comments,
			TechnicalAreas:         dfc.extractTechnicalAreas(uf.SpecificFeatures),
			ImprovementSuggestions: uf.ImprovementAreas,
			CodeQualityRating:      codeQuality,
			ArchitectureRating:     architecture,
			PerformanceRating:      performance,
			MaintainabilityRating:  maintainability,
			TechnicalDebt:          technicalDebt,
			SubmittedAt:            uf.SubmittedAt,
			Metadata:               uf.Metadata,
		}

		developerFeedback = append(developerFeedback, devFeedback)
	}

	return developerFeedback
}

// extractTechnicalRatings extracts technical ratings from specific features
func (dfc *DeveloperFeedbackCollector) extractTechnicalRatings(features []string) (int, int, int, int) {
	codeQuality := 3     // default
	architecture := 3    // default
	performance := 3     // default
	maintainability := 3 // default

	for _, feature := range features {
		if len(feature) > 12 && feature[:12] == "code_quality_" {
			codeQuality = int(feature[12] - '0')
		} else if len(feature) > 13 && feature[:13] == "architecture_" {
			architecture = int(feature[13] - '0')
		} else if len(feature) > 12 && feature[:12] == "performance_" {
			performance = int(feature[12] - '0')
		} else if len(feature) > 15 && feature[:15] == "maintainability_" {
			maintainability = int(feature[15] - '0')
		}
	}

	return codeQuality, architecture, performance, maintainability
}

// extractTechnicalAreas extracts technical areas from specific features
func (dfc *DeveloperFeedbackCollector) extractTechnicalAreas(features []string) []string {
	var technicalAreas []string

	for _, feature := range features {
		// Skip rating features
		if len(feature) > 12 && (feature[:12] == "code_quality_" ||
			feature[:13] == "architecture_" || feature[:12] == "performance_" ||
			feature[:15] == "maintainability_") {
			continue
		}
		technicalAreas = append(technicalAreas, feature)
	}

	return technicalAreas
}

// extractTechnicalDebt extracts technical debt information
func (dfc *DeveloperFeedbackCollector) extractTechnicalDebt(businessImpact BusinessImpactRating, improvements []string) TechnicalDebtAssessment {
	// Determine debt level based on cost reduction
	debtLevel := "medium"
	if businessImpact.CostReduction == "0%" {
		debtLevel = "critical"
	} else if businessImpact.CostReduction == "10%" || businessImpact.CostReduction == "20%" {
		debtLevel = "high"
	} else if businessImpact.CostReduction == "40%" || businessImpact.CostReduction == "50%" {
		debtLevel = "low"
	}

	// Determine priority based on ROI
	priority := "medium"
	if businessImpact.ROI == "Low" {
		priority = "critical"
	} else if businessImpact.ROI == "High" || businessImpact.ROI == "Very High" {
		priority = "low"
	}

	return TechnicalDebtAssessment{
		OverallDebtLevel:   debtLevel,
		DebtAreas:          improvements,
		EstimatedEffort:    "2-4 weeks",
		PriorityLevel:      priority,
		ImpactAssessment:   "Moderate impact on development velocity",
		RecommendedActions: []string{"Code review", "Refactoring", "Testing improvements"},
	}
}

// analyzeDeveloperFeedback performs comprehensive analysis of developer feedback data
func (dfc *DeveloperFeedbackCollector) analyzeDeveloperFeedback(feedback []*DeveloperFeedback) *DeveloperFeedbackAnalysis {
	if len(feedback) == 0 {
		return &DeveloperFeedbackAnalysis{
			GeneratedAt: time.Now(),
		}
	}

	analysis := &DeveloperFeedbackAnalysis{
		Category:       feedback[0].Category,
		TotalResponses: len(feedback),
		GeneratedAt:    time.Now(),
	}

	// Calculate averages
	var totalRating, totalCodeQuality, totalArchitecture, totalPerformance, totalMaintainability float64
	improvementCounts := make(map[string]int)
	technicalAreaCounts := make(map[string]int)
	debtLevelCounts := make(map[string]int)
	priorityCounts := make(map[string]int)

	for _, f := range feedback {
		totalRating += float64(f.Rating)
		totalCodeQuality += float64(f.CodeQualityRating)
		totalArchitecture += float64(f.ArchitectureRating)
		totalPerformance += float64(f.PerformanceRating)
		totalMaintainability += float64(f.MaintainabilityRating)

		// Count improvements
		for _, improvement := range f.ImprovementSuggestions {
			improvementCounts[improvement]++
		}

		// Count technical areas
		for _, area := range f.TechnicalAreas {
			technicalAreaCounts[area]++
		}

		// Count debt levels and priorities
		debtLevelCounts[f.TechnicalDebt.OverallDebtLevel]++
		priorityCounts[f.TechnicalDebt.PriorityLevel]++
	}

	// Calculate averages
	count := float64(len(feedback))
	analysis.AverageRating = totalRating / count
	analysis.AverageCodeQuality = totalCodeQuality / count
	analysis.AverageArchitecture = totalArchitecture / count
	analysis.AveragePerformance = totalPerformance / count
	analysis.AverageMaintainability = totalMaintainability / count

	// Find top improvements and technical areas
	analysis.TopImprovements = dfc.getTopItems(improvementCounts, 5)
	analysis.TopTechnicalAreas = dfc.getTopItems(technicalAreaCounts, 5)

	// Calculate technical debt summary
	analysis.TechnicalDebtSummary = dfc.calculateTechnicalDebtSummary(debtLevelCounts, priorityCounts)

	// Calculate code quality metrics
	analysis.CodeQualityMetrics = dfc.calculateCodeQualityMetrics(analysis)

	// Generate recommendations
	analysis.Recommendations = dfc.generateDeveloperRecommendations(analysis)

	return analysis
}

// getTopItems returns the top N items by count
func (dfc *DeveloperFeedbackCollector) getTopItems(counts map[string]int, n int) []string {
	var items []string
	for item, count := range counts {
		items = append(items, fmt.Sprintf("%s (%d)", item, count))
	}

	// Sort by count (simplified - in production, use proper sorting)
	if len(items) > n {
		items = items[:n]
	}

	return items
}

// calculateTechnicalDebtSummary calculates technical debt summary
func (dfc *DeveloperFeedbackCollector) calculateTechnicalDebtSummary(debtLevelCounts, priorityCounts map[string]int) TechnicalDebtSummary {
	// Find most common debt level
	mostCommonDebt := "medium"
	maxCount := 0
	for level, count := range debtLevelCounts {
		if count > maxCount {
			maxCount = count
			mostCommonDebt = level
		}
	}

	// Find most common debt areas (simplified)
	mostCommonAreas := []string{"code_quality", "architecture", "testing"}

	// Calculate estimated effort
	estimatedEffort := "2-4 weeks"
	if mostCommonDebt == "critical" {
		estimatedEffort = "1-2 months"
	} else if mostCommonDebt == "high" {
		estimatedEffort = "3-6 weeks"
	} else if mostCommonDebt == "low" {
		estimatedEffort = "1-2 weeks"
	}

	return TechnicalDebtSummary{
		AverageDebtLevel:     mostCommonDebt,
		MostCommonDebtAreas:  mostCommonAreas,
		PriorityDistribution: priorityCounts,
		EstimatedTotalEffort: estimatedEffort,
	}
}

// calculateCodeQualityMetrics calculates code quality metrics
func (dfc *DeveloperFeedbackCollector) calculateCodeQualityMetrics(analysis *DeveloperFeedbackAnalysis) CodeQualityMetrics {
	// Determine quality trend based on ratings
	trend := "stable"
	if analysis.AverageCodeQuality >= 4.0 && analysis.AverageArchitecture >= 4.0 {
		trend = "improving"
	} else if analysis.AverageCodeQuality <= 2.5 || analysis.AverageArchitecture <= 2.5 {
		trend = "declining"
	}

	return CodeQualityMetrics{
		AverageCodeQuality:     analysis.AverageCodeQuality,
		AverageArchitecture:    analysis.AverageArchitecture,
		AveragePerformance:     analysis.AveragePerformance,
		AverageMaintainability: analysis.AverageMaintainability,
		QualityTrend:           trend,
	}
}

// generateDeveloperRecommendations generates actionable recommendations based on analysis
func (dfc *DeveloperFeedbackCollector) generateDeveloperRecommendations(analysis *DeveloperFeedbackAnalysis) []string {
	var recommendations []string

	// Code quality recommendations
	if analysis.AverageCodeQuality < 3.5 {
		recommendations = append(recommendations, "Implement comprehensive code review processes")
		recommendations = append(recommendations, "Add automated code quality checks and linting")
	}

	// Architecture recommendations
	if analysis.AverageArchitecture < 3.5 {
		recommendations = append(recommendations, "Review and refactor system architecture")
		recommendations = append(recommendations, "Implement design patterns and best practices")
	}

	// Performance recommendations
	if analysis.AveragePerformance < 3.5 {
		recommendations = append(recommendations, "Optimize database queries and indexing")
		recommendations = append(recommendations, "Implement caching strategies")
	}

	// Maintainability recommendations
	if analysis.AverageMaintainability < 3.5 {
		recommendations = append(recommendations, "Improve code documentation and comments")
		recommendations = append(recommendations, "Implement comprehensive testing strategies")
	}

	// Technical debt recommendations
	if analysis.TechnicalDebtSummary.AverageDebtLevel == "high" || analysis.TechnicalDebtSummary.AverageDebtLevel == "critical" {
		recommendations = append(recommendations, "Prioritize technical debt reduction")
		recommendations = append(recommendations, "Allocate dedicated time for refactoring")
	}

	// General recommendations
	if analysis.AverageRating < 3.5 {
		recommendations = append(recommendations, "Conduct team retrospectives to identify improvement areas")
		recommendations = append(recommendations, "Implement continuous improvement processes")
	}

	return recommendations
}

// GetDeveloperFeedbackStats retrieves overall developer feedback statistics
func (dfc *DeveloperFeedbackCollector) GetDeveloperFeedbackStats(ctx context.Context) (*DeveloperFeedbackStats, error) {
	// Get all developer feedback
	allFeedback, err := dfc.storage.GetFeedbackByTimeRange(ctx, time.Now().AddDate(0, -1, 0), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve developer feedback stats: %w", err)
	}

	// Filter for developer feedback (based on metadata)
	var developerFeedback []*UserFeedback
	for _, feedback := range allFeedback {
		if feedback.Metadata != nil && feedback.Metadata["feedback_type"] == "developer" {
			developerFeedback = append(developerFeedback, feedback)
		}
	}

	// Convert to developer feedback format
	devFeedback := dfc.convertFromUserFeedback(developerFeedback)

	// Calculate statistics
	stats := dfc.calculateDeveloperFeedbackStats(devFeedback)

	return stats, nil
}

// calculateDeveloperFeedbackStats calculates comprehensive developer feedback statistics
func (dfc *DeveloperFeedbackCollector) calculateDeveloperFeedbackStats(feedback []*DeveloperFeedback) *DeveloperFeedbackStats {
	if len(feedback) == 0 {
		return &DeveloperFeedbackStats{
			LastUpdated: time.Now(),
		}
	}

	stats := &DeveloperFeedbackStats{
		TotalResponses: len(feedback),
		LastUpdated:    time.Now(),
	}

	// Calculate averages
	var totalRating, totalCodeQuality, totalArchitecture, totalPerformance, totalMaintainability float64
	categoryCounts := make(map[string]int)
	ratingCounts := make(map[int]int)
	improvementCounts := make(map[string]int)
	debtLevelCounts := make(map[string]int)
	priorityCounts := make(map[string]int)

	for _, f := range feedback {
		totalRating += float64(f.Rating)
		totalCodeQuality += float64(f.CodeQualityRating)
		totalArchitecture += float64(f.ArchitectureRating)
		totalPerformance += float64(f.PerformanceRating)
		totalMaintainability += float64(f.MaintainabilityRating)

		// Count categories and ratings
		categoryCounts[string(f.Category)]++
		ratingCounts[f.Rating]++

		// Count improvements
		for _, improvement := range f.ImprovementSuggestions {
			improvementCounts[improvement]++
		}

		// Count debt levels and priorities
		debtLevelCounts[f.TechnicalDebt.OverallDebtLevel]++
		priorityCounts[f.TechnicalDebt.PriorityLevel]++
	}

	// Calculate averages
	count := float64(len(feedback))
	stats.AverageRating = totalRating / count

	// Set breakdowns
	stats.CategoryBreakdown = categoryCounts
	stats.RatingDistribution = ratingCounts
	stats.CommonImprovements = dfc.getTopItems(improvementCounts, 10)

	// Calculate technical debt summary
	stats.TechnicalDebtSummary = dfc.calculateTechnicalDebtSummary(debtLevelCounts, priorityCounts)

	// Calculate code quality metrics
	stats.CodeQualityMetrics = CodeQualityMetrics{
		AverageCodeQuality:     totalCodeQuality / count,
		AverageArchitecture:    totalArchitecture / count,
		AveragePerformance:     totalPerformance / count,
		AverageMaintainability: totalMaintainability / count,
		QualityTrend:           "stable", // Would need historical data for trend analysis
	}

	return stats
}
