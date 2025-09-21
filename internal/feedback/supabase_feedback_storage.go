package feedback

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
)

// SupabaseFeedbackStorage implements FeedbackStorage interface for Supabase
type SupabaseFeedbackStorage struct {
	db     *sql.DB
	logger *log.Logger
}

// NewSupabaseFeedbackStorage creates a new Supabase feedback storage instance
func NewSupabaseFeedbackStorage(db *sql.DB, logger *log.Logger) *SupabaseFeedbackStorage {
	return &SupabaseFeedbackStorage{
		db:     db,
		logger: logger,
	}
}

// StoreFeedback stores user feedback in Supabase
func (sfs *SupabaseFeedbackStorage) StoreFeedback(ctx context.Context, feedback *UserFeedback) error {
	query := `
		INSERT INTO user_feedback (
			id, user_id, category, rating, comments, specific_features,
			improvement_areas, classification_accuracy, performance_rating,
			usability_rating, business_impact, submitted_at, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	// Convert slices to PostgreSQL arrays
	specificFeatures := pq.Array(feedback.SpecificFeatures)
	improvementAreas := pq.Array(feedback.ImprovementAreas)

	// Convert business impact to JSON
	businessImpactJSON, err := json.Marshal(feedback.BusinessImpact)
	if err != nil {
		return fmt.Errorf("failed to marshal business impact: %w", err)
	}

	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(feedback.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = sfs.db.ExecContext(ctx, query,
		feedback.ID,
		feedback.UserID,
		feedback.Category,
		feedback.Rating,
		feedback.Comments,
		specificFeatures,
		improvementAreas,
		feedback.ClassificationAccuracy,
		feedback.PerformanceRating,
		feedback.UsabilityRating,
		businessImpactJSON,
		feedback.SubmittedAt,
		metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to store feedback: %w", err)
	}

	sfs.logger.Printf("Feedback stored successfully: ID=%s", feedback.ID)
	return nil
}

// GetFeedbackByCategory retrieves feedback by category
func (sfs *SupabaseFeedbackStorage) GetFeedbackByCategory(ctx context.Context, category string) ([]*UserFeedback, error) {
	query := `
		SELECT id, user_id, category, rating, comments, specific_features,
			   improvement_areas, classification_accuracy, performance_rating,
			   usability_rating, business_impact, submitted_at, metadata
		FROM user_feedback
		WHERE category = $1
		ORDER BY submitted_at DESC
	`

	rows, err := sfs.db.QueryContext(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to query feedback by category: %w", err)
	}
	defer rows.Close()

	return sfs.scanFeedbackRows(rows)
}

// GetFeedbackByTimeRange retrieves feedback within a time range
func (sfs *SupabaseFeedbackStorage) GetFeedbackByTimeRange(ctx context.Context, start, end time.Time) ([]*UserFeedback, error) {
	query := `
		SELECT id, user_id, category, rating, comments, specific_features,
			   improvement_areas, classification_accuracy, performance_rating,
			   usability_rating, business_impact, submitted_at, metadata
		FROM user_feedback
		WHERE submitted_at BETWEEN $1 AND $2
		ORDER BY submitted_at DESC
	`

	rows, err := sfs.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query feedback by time range: %w", err)
	}
	defer rows.Close()

	return sfs.scanFeedbackRows(rows)
}

// GetFeedbackStats retrieves aggregated feedback statistics
func (sfs *SupabaseFeedbackStorage) GetFeedbackStats(ctx context.Context) (*FeedbackStats, error) {
	// Get total responses
	var totalResponses int
	err := sfs.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_feedback").Scan(&totalResponses)
	if err != nil {
		return nil, fmt.Errorf("failed to get total responses: %w", err)
	}

	// Get average rating
	var averageRating sql.NullFloat64
	err = sfs.db.QueryRowContext(ctx, "SELECT AVG(rating) FROM user_feedback").Scan(&averageRating)
	if err != nil {
		return nil, fmt.Errorf("failed to get average rating: %w", err)
	}

	// Get category breakdown
	categoryQuery := `
		SELECT category, COUNT(*) as count
		FROM user_feedback
		GROUP BY category
		ORDER BY count DESC
	`

	rows, err := sfs.db.QueryContext(ctx, categoryQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get category breakdown: %w", err)
	}
	defer rows.Close()

	categoryBreakdown := make(map[string]int)
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			return nil, fmt.Errorf("failed to scan category breakdown: %w", err)
		}
		categoryBreakdown[category] = count
	}

	// Get rating distribution
	ratingQuery := `
		SELECT rating, COUNT(*) as count
		FROM user_feedback
		GROUP BY rating
		ORDER BY rating
	`

	rows, err = sfs.db.QueryContext(ctx, ratingQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating distribution: %w", err)
	}
	defer rows.Close()

	ratingDistribution := make(map[int]int)
	for rows.Next() {
		var rating, count int
		if err := rows.Scan(&rating, &count); err != nil {
			return nil, fmt.Errorf("failed to scan rating distribution: %w", err)
		}
		ratingDistribution[rating] = count
	}

	// Get common improvements
	improvementsQuery := `
		SELECT unnest(improvement_areas) as improvement, COUNT(*) as count
		FROM user_feedback
		WHERE improvement_areas IS NOT NULL
		GROUP BY improvement
		ORDER BY count DESC
		LIMIT 10
	`

	rows, err = sfs.db.QueryContext(ctx, improvementsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get common improvements: %w", err)
	}
	defer rows.Close()

	var commonImprovements []string
	for rows.Next() {
		var improvement string
		var count int
		if err := rows.Scan(&improvement, &count); err != nil {
			return nil, fmt.Errorf("failed to scan common improvements: %w", err)
		}
		commonImprovements = append(commonImprovements, fmt.Sprintf("%s (%d)", improvement, count))
	}

	// Get business impact averages
	var avgTimeSaved, avgErrorReduction, avgProductivity sql.NullFloat64
	businessImpactQuery := `
		SELECT 
			AVG((business_impact->>'time_saved_minutes')::int),
			AVG((business_impact->>'error_reduction_percentage')::int),
			AVG((business_impact->>'productivity_gain_percentage')::int)
		FROM user_feedback
		WHERE business_impact IS NOT NULL
	`

	err = sfs.db.QueryRowContext(ctx, businessImpactQuery).Scan(&avgTimeSaved, &avgErrorReduction, &avgProductivity)
	if err != nil {
		return nil, fmt.Errorf("failed to get business impact averages: %w", err)
	}

	stats := &FeedbackStats{
		TotalResponses:     totalResponses,
		AverageRating:      averageRating.Float64,
		CategoryBreakdown:  categoryBreakdown,
		RatingDistribution: ratingDistribution,
		CommonImprovements: commonImprovements,
		BusinessImpactAvg: BusinessImpactRating{
			TimeSaved:        int(avgTimeSaved.Float64),
			ErrorReduction:   int(avgErrorReduction.Float64),
			ProductivityGain: int(avgProductivity.Float64),
		},
		ResponseRate: 0.0, // Would need to calculate based on total users
		LastUpdated:  time.Now(),
	}

	return stats, nil
}

// scanFeedbackRows scans database rows into UserFeedback structs
func (sfs *SupabaseFeedbackStorage) scanFeedbackRows(rows *sql.Rows) ([]*UserFeedback, error) {
	var feedback []*UserFeedback

	for rows.Next() {
		f := &UserFeedback{}
		var specificFeatures, improvementAreas pq.StringArray
		var businessImpactJSON, metadataJSON []byte

		err := rows.Scan(
			&f.ID,
			&f.UserID,
			&f.Category,
			&f.Rating,
			&f.Comments,
			&specificFeatures,
			&improvementAreas,
			&f.ClassificationAccuracy,
			&f.PerformanceRating,
			&f.UsabilityRating,
			&businessImpactJSON,
			&f.SubmittedAt,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feedback row: %w", err)
		}

		// Convert arrays
		f.SpecificFeatures = []string(specificFeatures)
		f.ImprovementAreas = []string(improvementAreas)

		// Unmarshal JSON fields
		if len(businessImpactJSON) > 0 {
			if err := json.Unmarshal(businessImpactJSON, &f.BusinessImpact); err != nil {
				return nil, fmt.Errorf("failed to unmarshal business impact: %w", err)
			}
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &f.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		feedback = append(feedback, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating feedback rows: %w", err)
	}

	return feedback, nil
}

// CreateFeedbackTable creates the user_feedback table in Supabase
func (sfs *SupabaseFeedbackStorage) CreateFeedbackTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS user_feedback (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id VARCHAR(255) NOT NULL,
			category VARCHAR(50) NOT NULL CHECK (category IN (
				'database_performance', 'classification_accuracy', 'user_experience',
				'risk_detection', 'overall_satisfaction', 'feature_request', 'bug_report'
			)),
			rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
			comments TEXT,
			specific_features TEXT[],
			improvement_areas TEXT[],
			classification_accuracy DECIMAL(3,2) CHECK (classification_accuracy >= 0 AND classification_accuracy <= 1),
			performance_rating INTEGER NOT NULL CHECK (performance_rating >= 1 AND performance_rating <= 5),
			usability_rating INTEGER NOT NULL CHECK (usability_rating >= 1 AND usability_rating <= 5),
			business_impact JSONB,
			submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			metadata JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Create indexes for better query performance
		CREATE INDEX IF NOT EXISTS idx_user_feedback_category ON user_feedback(category);
		CREATE INDEX IF NOT EXISTS idx_user_feedback_rating ON user_feedback(rating);
		CREATE INDEX IF NOT EXISTS idx_user_feedback_submitted_at ON user_feedback(submitted_at);
		CREATE INDEX IF NOT EXISTS idx_user_feedback_user_id ON user_feedback(user_id);

		-- Create GIN index for JSONB fields
		CREATE INDEX IF NOT EXISTS idx_user_feedback_business_impact ON user_feedback USING GIN (business_impact);
		CREATE INDEX IF NOT EXISTS idx_user_feedback_metadata ON user_feedback USING GIN (metadata);
	`

	_, err := sfs.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create user_feedback table: %w", err)
	}

	sfs.logger.Println("User feedback table created successfully")
	return nil
}
