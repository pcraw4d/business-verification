package classification

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq" // For array scanning
)

// PerformanceOptimizationRecommendations provides comprehensive performance optimization recommendations
type PerformanceOptimizationRecommendations struct {
	db *sql.DB
}

// NewPerformanceOptimizationRecommendations creates a new instance of PerformanceOptimizationRecommendations
func NewPerformanceOptimizationRecommendations(db *sql.DB) *PerformanceOptimizationRecommendations {
	return &PerformanceOptimizationRecommendations{
		db: db,
	}
}

// PerformanceRecommendation represents a performance optimization recommendation
type PerformanceRecommendation struct {
	ID                               int        `json:"id"`
	RecommendationID                 string     `json:"recommendation_id"`
	RecommendationType               string     `json:"recommendation_type"`
	RecommendationCategory           string     `json:"recommendation_category"`
	RecommendationTitle              string     `json:"recommendation_title"`
	RecommendationDescription        string     `json:"recommendation_description"`
	RecommendationPriority           string     `json:"recommendation_priority"`
	RecommendationImpact             string     `json:"recommendation_impact"`
	RecommendationEffort             string     `json:"recommendation_effort"`
	RecommendationBenefit            string     `json:"recommendation_benefit"`
	RecommendationImplementation     string     `json:"recommendation_implementation"`
	RecommendationValidation         string     `json:"recommendation_validation"`
	AffectedSystems                  []string   `json:"affected_systems"`
	RelatedMetrics                   []string   `json:"related_metrics"`
	EstimatedImprovementPercentage   *float64   `json:"estimated_improvement_percentage"`
	EstimatedImplementationTimeHours *float64   `json:"estimated_implementation_time_hours"`
	Prerequisites                    []string   `json:"prerequisites"`
	Dependencies                     []string   `json:"dependencies"`
	Status                           string     `json:"status"`
	AssignedTo                       *string    `json:"assigned_to"`
	AssignedAt                       *time.Time `json:"assigned_at"`
	ImplementedAt                    *time.Time `json:"implemented_at"`
	ImplementationNotes              *string    `json:"implementation_notes"`
	ValidationResults                *string    `json:"validation_results"`
	CreatedAt                        time.Time  `json:"created_at"`
	UpdatedAt                        time.Time  `json:"updated_at"`
}

// RecommendationStatistics represents recommendation statistics
type RecommendationStatistics struct {
	TotalRecommendations       int64            `json:"total_recommendations"`
	PendingRecommendations     int64            `json:"pending_recommendations"`
	ImplementedRecommendations int64            `json:"implemented_recommendations"`
	CriticalRecommendations    int64            `json:"critical_recommendations"`
	HighRecommendations        int64            `json:"high_recommendations"`
	MediumRecommendations      int64            `json:"medium_recommendations"`
	LowRecommendations         int64            `json:"low_recommendations"`
	RecommendationsByCategory  map[string]int64 `json:"recommendations_by_category"`
	RecommendationsByType      map[string]int64 `json:"recommendations_by_type"`
	AvgImplementationTimeHours *float64         `json:"avg_implementation_time_hours"`
	TotalEstimatedImprovement  *float64         `json:"total_estimated_improvement"`
}

// RecommendationValidation represents recommendations setup validation
type RecommendationValidation struct {
	Component      string `json:"component"`
	Status         string `json:"status"`
	Details        string `json:"details"`
	Recommendation string `json:"recommendation"`
}

// GenerateDatabasePerformanceRecommendations generates database performance recommendations
func (por *PerformanceOptimizationRecommendations) GenerateDatabasePerformanceRecommendations(ctx context.Context) ([]PerformanceRecommendation, error) {
	query := `SELECT * FROM generate_database_performance_recommendations()`

	rows, err := por.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate database performance recommendations: %w", err)
	}
	defer rows.Close()

	var results []PerformanceRecommendation
	for rows.Next() {
		var result PerformanceRecommendation
		var affectedSystems, relatedMetrics, prerequisites, dependencies pq.StringArray

		err := rows.Scan(
			&result.RecommendationID,
			&result.RecommendationType,
			&result.RecommendationCategory,
			&result.RecommendationTitle,
			&result.RecommendationDescription,
			&result.RecommendationPriority,
			&result.RecommendationImpact,
			&result.RecommendationEffort,
			&result.RecommendationBenefit,
			&result.RecommendationImplementation,
			&result.RecommendationValidation,
			&affectedSystems,
			&relatedMetrics,
			&result.EstimatedImprovementPercentage,
			&result.EstimatedImplementationTimeHours,
			&prerequisites,
			&dependencies,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan database performance recommendation: %w", err)
		}

		result.AffectedSystems = []string(affectedSystems)
		result.RelatedMetrics = []string(relatedMetrics)
		result.Prerequisites = []string(prerequisites)
		result.Dependencies = []string(dependencies)
		results = append(results, result)
	}

	return results, nil
}

// GenerateClassificationPerformanceRecommendations generates classification performance recommendations
func (por *PerformanceOptimizationRecommendations) GenerateClassificationPerformanceRecommendations(ctx context.Context) ([]PerformanceRecommendation, error) {
	query := `SELECT * FROM generate_classification_performance_recommendations()`

	rows, err := por.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate classification performance recommendations: %w", err)
	}
	defer rows.Close()

	var results []PerformanceRecommendation
	for rows.Next() {
		var result PerformanceRecommendation
		var affectedSystems, relatedMetrics, prerequisites, dependencies pq.StringArray

		err := rows.Scan(
			&result.RecommendationID,
			&result.RecommendationType,
			&result.RecommendationCategory,
			&result.RecommendationTitle,
			&result.RecommendationDescription,
			&result.RecommendationPriority,
			&result.RecommendationImpact,
			&result.RecommendationEffort,
			&result.RecommendationBenefit,
			&result.RecommendationImplementation,
			&result.RecommendationValidation,
			&affectedSystems,
			&relatedMetrics,
			&result.EstimatedImprovementPercentage,
			&result.EstimatedImplementationTimeHours,
			&prerequisites,
			&dependencies,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan classification performance recommendation: %w", err)
		}

		result.AffectedSystems = []string(affectedSystems)
		result.RelatedMetrics = []string(relatedMetrics)
		result.Prerequisites = []string(prerequisites)
		result.Dependencies = []string(dependencies)
		results = append(results, result)
	}

	return results, nil
}

// GenerateSystemResourceRecommendations generates system resource recommendations
func (por *PerformanceOptimizationRecommendations) GenerateSystemResourceRecommendations(ctx context.Context) ([]PerformanceRecommendation, error) {
	query := `SELECT * FROM generate_system_resource_recommendations()`

	rows, err := por.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate system resource recommendations: %w", err)
	}
	defer rows.Close()

	var results []PerformanceRecommendation
	for rows.Next() {
		var result PerformanceRecommendation
		var affectedSystems, relatedMetrics, prerequisites, dependencies pq.StringArray

		err := rows.Scan(
			&result.RecommendationID,
			&result.RecommendationType,
			&result.RecommendationCategory,
			&result.RecommendationTitle,
			&result.RecommendationDescription,
			&result.RecommendationPriority,
			&result.RecommendationImpact,
			&result.RecommendationEffort,
			&result.RecommendationBenefit,
			&result.RecommendationImplementation,
			&result.RecommendationValidation,
			&affectedSystems,
			&relatedMetrics,
			&result.EstimatedImprovementPercentage,
			&result.EstimatedImplementationTimeHours,
			&prerequisites,
			&dependencies,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan system resource recommendation: %w", err)
		}

		result.AffectedSystems = []string(affectedSystems)
		result.RelatedMetrics = []string(relatedMetrics)
		result.Prerequisites = []string(prerequisites)
		result.Dependencies = []string(dependencies)
		results = append(results, result)
	}

	return results, nil
}

// GetAllPerformanceRecommendations gets all performance recommendations
func (por *PerformanceOptimizationRecommendations) GetAllPerformanceRecommendations(ctx context.Context) ([]PerformanceRecommendation, error) {
	query := `SELECT * FROM get_all_performance_recommendations()`

	rows, err := por.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all performance recommendations: %w", err)
	}
	defer rows.Close()

	var results []PerformanceRecommendation
	for rows.Next() {
		var result PerformanceRecommendation
		var affectedSystems, relatedMetrics, prerequisites, dependencies pq.StringArray

		err := rows.Scan(
			&result.RecommendationID,
			&result.RecommendationType,
			&result.RecommendationCategory,
			&result.RecommendationTitle,
			&result.RecommendationDescription,
			&result.RecommendationPriority,
			&result.RecommendationImpact,
			&result.RecommendationEffort,
			&result.RecommendationBenefit,
			&result.RecommendationImplementation,
			&result.RecommendationValidation,
			&affectedSystems,
			&relatedMetrics,
			&result.EstimatedImprovementPercentage,
			&result.EstimatedImplementationTimeHours,
			&prerequisites,
			&dependencies,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance recommendation: %w", err)
		}

		result.AffectedSystems = []string(affectedSystems)
		result.RelatedMetrics = []string(relatedMetrics)
		result.Prerequisites = []string(prerequisites)
		result.Dependencies = []string(dependencies)
		results = append(results, result)
	}

	return results, nil
}

// SavePerformanceRecommendations saves performance recommendations to the database
func (por *PerformanceOptimizationRecommendations) SavePerformanceRecommendations(ctx context.Context) (int, error) {
	query := `SELECT save_performance_recommendations()`

	var savedCount int
	err := por.db.QueryRowContext(ctx, query).Scan(&savedCount)

	if err != nil {
		return 0, fmt.Errorf("failed to save performance recommendations: %w", err)
	}

	return savedCount, nil
}

// GetRecommendationsByPriority gets recommendations by priority
func (por *PerformanceOptimizationRecommendations) GetRecommendationsByPriority(ctx context.Context, priority string) ([]PerformanceRecommendation, error) {
	query := `SELECT * FROM get_recommendations_by_priority($1)`

	rows, err := por.db.QueryContext(ctx, query, priority)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations by priority: %w", err)
	}
	defer rows.Close()

	var results []PerformanceRecommendation
	for rows.Next() {
		var result PerformanceRecommendation
		var affectedSystems, relatedMetrics, prerequisites, dependencies pq.StringArray

		err := rows.Scan(
			&result.RecommendationID,
			&result.RecommendationType,
			&result.RecommendationCategory,
			&result.RecommendationTitle,
			&result.RecommendationDescription,
			&result.RecommendationPriority,
			&result.RecommendationImpact,
			&result.RecommendationEffort,
			&result.RecommendationBenefit,
			&result.RecommendationImplementation,
			&result.RecommendationValidation,
			&affectedSystems,
			&relatedMetrics,
			&result.EstimatedImprovementPercentage,
			&result.EstimatedImplementationTimeHours,
			&prerequisites,
			&dependencies,
			&result.Status,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recommendation by priority: %w", err)
		}

		result.AffectedSystems = []string(affectedSystems)
		result.RelatedMetrics = []string(relatedMetrics)
		result.Prerequisites = []string(prerequisites)
		result.Dependencies = []string(dependencies)
		results = append(results, result)
	}

	return results, nil
}

// GetRecommendationsByCategory gets recommendations by category
func (por *PerformanceOptimizationRecommendations) GetRecommendationsByCategory(ctx context.Context, category string) ([]PerformanceRecommendation, error) {
	query := `SELECT * FROM get_recommendations_by_category($1)`

	rows, err := por.db.QueryContext(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations by category: %w", err)
	}
	defer rows.Close()

	var results []PerformanceRecommendation
	for rows.Next() {
		var result PerformanceRecommendation
		var affectedSystems, relatedMetrics, prerequisites, dependencies pq.StringArray

		err := rows.Scan(
			&result.RecommendationID,
			&result.RecommendationType,
			&result.RecommendationCategory,
			&result.RecommendationTitle,
			&result.RecommendationDescription,
			&result.RecommendationPriority,
			&result.RecommendationImpact,
			&result.RecommendationEffort,
			&result.RecommendationBenefit,
			&result.RecommendationImplementation,
			&result.RecommendationValidation,
			&affectedSystems,
			&relatedMetrics,
			&result.EstimatedImprovementPercentage,
			&result.EstimatedImplementationTimeHours,
			&prerequisites,
			&dependencies,
			&result.Status,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recommendation by category: %w", err)
		}

		result.AffectedSystems = []string(affectedSystems)
		result.RelatedMetrics = []string(relatedMetrics)
		result.Prerequisites = []string(prerequisites)
		result.Dependencies = []string(dependencies)
		results = append(results, result)
	}

	return results, nil
}

// ImplementRecommendation implements a recommendation
func (por *PerformanceOptimizationRecommendations) ImplementRecommendation(ctx context.Context, recommendationID, implementedBy string, implementationNotes *string) (bool, error) {
	query := `SELECT implement_recommendation($1, $2, $3)`

	var implemented bool
	err := por.db.QueryRowContext(ctx, query, recommendationID, implementedBy, implementationNotes).Scan(&implemented)

	if err != nil {
		return false, fmt.Errorf("failed to implement recommendation: %w", err)
	}

	return implemented, nil
}

// GetRecommendationStatistics gets recommendation statistics
func (por *PerformanceOptimizationRecommendations) GetRecommendationStatistics(ctx context.Context) (*RecommendationStatistics, error) {
	query := `SELECT * FROM get_recommendation_statistics()`

	var result RecommendationStatistics
	var recommendationsByCategory, recommendationsByType []byte

	err := por.db.QueryRowContext(ctx, query).Scan(
		&result.TotalRecommendations,
		&result.PendingRecommendations,
		&result.ImplementedRecommendations,
		&result.CriticalRecommendations,
		&result.HighRecommendations,
		&result.MediumRecommendations,
		&result.LowRecommendations,
		&recommendationsByCategory,
		&recommendationsByType,
		&result.AvgImplementationTimeHours,
		&result.TotalEstimatedImprovement,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get recommendation statistics: %w", err)
	}

	// Parse JSON fields
	if err := parseJSONField(recommendationsByCategory, &result.RecommendationsByCategory); err != nil {
		log.Printf("Warning: failed to parse recommendations by category: %v", err)
	}

	if err := parseJSONField(recommendationsByType, &result.RecommendationsByType); err != nil {
		log.Printf("Warning: failed to parse recommendations by type: %v", err)
	}

	return &result, nil
}

// ValidateRecommendationsSetup validates the recommendations setup
func (por *PerformanceOptimizationRecommendations) ValidateRecommendationsSetup(ctx context.Context) ([]RecommendationValidation, error) {
	query := `SELECT * FROM validate_recommendations_setup()`

	rows, err := por.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to validate recommendations setup: %w", err)
	}
	defer rows.Close()

	var results []RecommendationValidation
	for rows.Next() {
		var result RecommendationValidation
		err := rows.Scan(
			&result.Component,
			&result.Status,
			&result.Details,
			&result.Recommendation,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recommendation validation: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetCurrentRecommendationsStatus gets current recommendations status summary
func (por *PerformanceOptimizationRecommendations) GetCurrentRecommendationsStatus(ctx context.Context) (map[string]interface{}, error) {
	status := make(map[string]interface{})

	// Get all recommendations
	allRecommendations, err := por.GetAllPerformanceRecommendations(ctx)
	if err != nil {
		log.Printf("Warning: failed to get all recommendations: %v", err)
	} else {
		status["all_recommendations"] = allRecommendations
		status["total_recommendations"] = len(allRecommendations)
	}

	// Get critical recommendations
	criticalRecommendations, err := por.GetRecommendationsByPriority(ctx, "CRITICAL")
	if err != nil {
		log.Printf("Warning: failed to get critical recommendations: %v", err)
	} else {
		status["critical_recommendations"] = criticalRecommendations
		status["critical_count"] = len(criticalRecommendations)
	}

	// Get high priority recommendations
	highRecommendations, err := por.GetRecommendationsByPriority(ctx, "HIGH")
	if err != nil {
		log.Printf("Warning: failed to get high recommendations: %v", err)
	} else {
		status["high_recommendations"] = highRecommendations
		status["high_count"] = len(highRecommendations)
	}

	// Get recommendation statistics
	stats, err := por.GetRecommendationStatistics(ctx)
	if err != nil {
		log.Printf("Warning: failed to get recommendation statistics: %v", err)
	} else {
		status["recommendation_statistics"] = stats
	}

	// Determine overall status
	overallStatus := "OK"
	if len(criticalRecommendations) > 0 {
		overallStatus = "CRITICAL"
	} else if len(highRecommendations) > 0 {
		overallStatus = "WARNING"
	} else if len(allRecommendations) > 0 {
		overallStatus = "FAIR"
	}

	status["overall_status"] = overallStatus
	status["last_checked"] = time.Now()

	return status, nil
}

// GenerateAndSaveRecommendations generates and saves performance recommendations
func (por *PerformanceOptimizationRecommendations) GenerateAndSaveRecommendations(ctx context.Context) (int, error) {
	// Generate recommendations
	allRecommendations, err := por.GetAllPerformanceRecommendations(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to generate recommendations: %w", err)
	}

	// Save recommendations
	savedCount, err := por.SavePerformanceRecommendations(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to save recommendations: %w", err)
	}

	log.Printf("Generated %d recommendations and saved %d to database", len(allRecommendations), savedCount)

	return savedCount, nil
}

// GetPerformanceOptimizationSummary gets a comprehensive performance optimization summary
func (por *PerformanceOptimizationRecommendations) GetPerformanceOptimizationSummary(ctx context.Context) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// Get current status
	status, err := por.GetCurrentRecommendationsStatus(ctx)
	if err != nil {
		log.Printf("Warning: failed to get current recommendations status: %v", err)
	} else {
		summary["current_status"] = status
	}

	// Get recommendation statistics
	stats, err := por.GetRecommendationStatistics(ctx)
	if err != nil {
		log.Printf("Warning: failed to get recommendation statistics: %v", err)
	} else {
		summary["recommendation_statistics"] = stats
	}

	// Get database performance recommendations
	dbRecommendations, err := por.GenerateDatabasePerformanceRecommendations(ctx)
	if err != nil {
		log.Printf("Warning: failed to get database performance recommendations: %v", err)
	} else {
		summary["database_recommendations"] = dbRecommendations
	}

	// Get classification performance recommendations
	classificationRecommendations, err := por.GenerateClassificationPerformanceRecommendations(ctx)
	if err != nil {
		log.Printf("Warning: failed to get classification performance recommendations: %v", err)
	} else {
		summary["classification_recommendations"] = classificationRecommendations
	}

	// Get system resource recommendations
	systemRecommendations, err := por.GenerateSystemResourceRecommendations(ctx)
	if err != nil {
		log.Printf("Warning: failed to get system resource recommendations: %v", err)
	} else {
		summary["system_recommendations"] = systemRecommendations
	}

	summary["last_updated"] = time.Now()

	return summary, nil
}

// MonitorRecommendationsContinuously starts continuous recommendations monitoring
func (por *PerformanceOptimizationRecommendations) MonitorRecommendationsContinuously(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting continuous performance recommendations monitoring with interval: %v", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping continuous performance recommendations monitoring")
			return
		case <-ticker.C:
			// Generate and save recommendations
			savedCount, err := por.GenerateAndSaveRecommendations(ctx)
			if err != nil {
				log.Printf("Error generating and saving recommendations: %v", err)
				continue
			}

			log.Printf("Generated and saved %d performance recommendations", savedCount)

			// Get current recommendations status
			status, err := por.GetCurrentRecommendationsStatus(ctx)
			if err != nil {
				log.Printf("Error getting recommendations status: %v", err)
			} else {
				log.Printf("Recommendations status: %s", status["overall_status"])
			}
		}
	}
}
