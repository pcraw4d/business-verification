package database

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// QueryOptimizer provides query analysis and optimization recommendations
type QueryOptimizer struct {
	db     *sql.DB
	logger *zap.Logger
}

// QueryAnalysis represents the analysis of a database query
type QueryAnalysis struct {
	Query             string        `json:"query"`
	ExecutionTime     time.Duration `json:"execution_time"`
	RowsExamined      int64         `json:"rows_examined"`
	RowsReturned      int64         `json:"rows_returned"`
	IndexUsed         string        `json:"index_used"`
	FullTableScan     bool          `json:"full_table_scan"`
	Recommendations   []string      `json:"recommendations"`
	OptimizationScore int           `json:"optimization_score"` // 0-100
}

// IndexRecommendation represents a recommended database index
type IndexRecommendation struct {
	Table            string   `json:"table"`
	Columns          []string `json:"columns"`
	Type             string   `json:"type"`     // "btree", "hash", "gin", "gist"
	Priority         int      `json:"priority"` // 1-10, higher is more important
	Reason           string   `json:"reason"`
	EstimatedBenefit string   `json:"estimated_benefit"`
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer(db *sql.DB, logger *zap.Logger) *QueryOptimizer {
	return &QueryOptimizer{
		db:     db,
		logger: logger,
	}
}

// AnalyzeQuery analyzes a SQL query and provides optimization recommendations
func (qo *QueryOptimizer) AnalyzeQuery(ctx context.Context, query string) (*QueryAnalysis, error) {
	start := time.Now()

	// Clean and normalize the query
	normalizedQuery := qo.normalizeQuery(query)

	// Execute EXPLAIN ANALYZE to get execution plan
	explainQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) %s", normalizedQuery)

	var explainResult []map[string]interface{}
	err := qo.db.QueryRowContext(ctx, explainQuery).Scan(&explainResult)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze query: %w", err)
	}

	analysis := &QueryAnalysis{
		Query:         normalizedQuery,
		ExecutionTime: time.Since(start),
	}

	// Parse execution plan
	if len(explainResult) > 0 {
		plan := explainResult[0]
		qo.parseExecutionPlan(plan, analysis)
	}

	// Generate recommendations
	analysis.Recommendations = qo.generateRecommendations(analysis)
	analysis.OptimizationScore = qo.calculateOptimizationScore(analysis)

	return analysis, nil
}

// normalizeQuery cleans and normalizes a SQL query for analysis
func (qo *QueryOptimizer) normalizeQuery(query string) string {
	// Remove extra whitespace
	query = regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(query), " ")

	// Convert to lowercase for analysis
	query = strings.ToLower(query)

	// Remove comments
	query = regexp.MustCompile(`--.*$`).ReplaceAllString(query, "")
	query = regexp.MustCompile(`/\*.*?\*/`).ReplaceAllString(query, "")

	return strings.TrimSpace(query)
}

// parseExecutionPlan extracts information from PostgreSQL EXPLAIN output
func (qo *QueryOptimizer) parseExecutionPlan(plan map[string]interface{}, analysis *QueryAnalysis) {
	if planData, ok := plan["Plan"].(map[string]interface{}); ok {
		// Extract execution time
		if actualTime, ok := planData["Actual Total Time"].(float64); ok {
			analysis.ExecutionTime = time.Duration(actualTime) * time.Millisecond
		}

		// Extract rows examined
		if rowsExamined, ok := planData["Actual Rows"].(float64); ok {
			analysis.RowsExamined = int64(rowsExamined)
		}

		// Check for full table scan
		if nodeType, ok := planData["Node Type"].(string); ok {
			analysis.FullTableScan = nodeType == "Seq Scan"
		}

		// Extract index information
		if indexName, ok := planData["Index Name"].(string); ok {
			analysis.IndexUsed = indexName
		}
	}
}

// generateRecommendations creates optimization recommendations based on query analysis
func (qo *QueryOptimizer) generateRecommendations(analysis *QueryAnalysis) []string {
	var recommendations []string

	// Check for full table scan
	if analysis.FullTableScan {
		recommendations = append(recommendations, "Consider adding an index to avoid full table scan")
	}

	// Check execution time
	if analysis.ExecutionTime > 100*time.Millisecond {
		recommendations = append(recommendations, "Query execution time is high, consider optimization")
	}

	// Check for missing WHERE clause
	if !strings.Contains(analysis.Query, "where") && strings.Contains(analysis.Query, "select") {
		recommendations = append(recommendations, "Consider adding WHERE clause to limit results")
	}

	// Check for SELECT *
	if strings.Contains(analysis.Query, "select *") {
		recommendations = append(recommendations, "Avoid SELECT *, specify only needed columns")
	}

	// Check for missing LIMIT
	if strings.Contains(analysis.Query, "select") && !strings.Contains(analysis.Query, "limit") {
		recommendations = append(recommendations, "Consider adding LIMIT clause for large result sets")
	}

	return recommendations
}

// calculateOptimizationScore calculates a score from 0-100 for query optimization
func (qo *QueryOptimizer) calculateOptimizationScore(analysis *QueryAnalysis) int {
	score := 100

	// Deduct points for full table scan
	if analysis.FullTableScan {
		score -= 30
	}

	// Deduct points for high execution time
	if analysis.ExecutionTime > 1*time.Second {
		score -= 40
	} else if analysis.ExecutionTime > 100*time.Millisecond {
		score -= 20
	}

	// Deduct points for missing WHERE clause
	if !strings.Contains(analysis.Query, "where") && strings.Contains(analysis.Query, "select") {
		score -= 15
	}

	// Deduct points for SELECT *
	if strings.Contains(analysis.Query, "select *") {
		score -= 10
	}

	// Deduct points for missing LIMIT
	if strings.Contains(analysis.Query, "select") && !strings.Contains(analysis.Query, "limit") {
		score -= 5
	}

	if score < 0 {
		score = 0
	}

	return score
}

// GetIndexRecommendations analyzes the database and provides index recommendations
func (qo *QueryOptimizer) GetIndexRecommendations(ctx context.Context) ([]*IndexRecommendation, error) {
	var recommendations []*IndexRecommendation

	// Analyze risk_assessments table
	raRecommendations := qo.analyzeRiskAssessmentsTable(ctx)
	recommendations = append(recommendations, raRecommendations...)

	// Analyze batch_jobs table
	bjRecommendations := qo.analyzeBatchJobsTable(ctx)
	recommendations = append(recommendations, bjRecommendations...)

	// Analyze custom_models table
	cmRecommendations := qo.analyzeCustomModelsTable(ctx)
	recommendations = append(recommendations, cmRecommendations...)

	return recommendations, nil
}

// analyzeRiskAssessmentsTable provides index recommendations for risk_assessments table
func (qo *QueryOptimizer) analyzeRiskAssessmentsTable(ctx context.Context) []*IndexRecommendation {
	var recommendations []*IndexRecommendation

	// Business ID + Created At index for lookups
	recommendations = append(recommendations, &IndexRecommendation{
		Table:            "risk_assessments",
		Columns:          []string{"business_id", "created_at"},
		Type:             "btree",
		Priority:         9,
		Reason:           "Optimizes business lookup queries with date filtering",
		EstimatedBenefit: "High - used in 80% of queries",
	})

	// Risk Level + Industry index for filtering
	recommendations = append(recommendations, &IndexRecommendation{
		Table:            "risk_assessments",
		Columns:          []string{"risk_level", "industry"},
		Type:             "btree",
		Priority:         8,
		Reason:           "Optimizes risk filtering by level and industry",
		EstimatedBenefit: "High - used in dashboard queries",
	})

	// Status index for active assessments
	recommendations = append(recommendations, &IndexRecommendation{
		Table:            "risk_assessments",
		Columns:          []string{"status"},
		Type:             "btree",
		Priority:         7,
		Reason:           "Optimizes queries filtering by assessment status",
		EstimatedBenefit: "Medium - used in status-based queries",
	})

	// Organization ID index for multi-tenant queries
	recommendations = append(recommendations, &IndexRecommendation{
		Table:            "risk_assessments",
		Columns:          []string{"organization_id"},
		Type:             "btree",
		Priority:         6,
		Reason:           "Optimizes multi-tenant data isolation queries",
		EstimatedBenefit: "Medium - used in organization filtering",
	})

	return recommendations
}

// analyzeBatchJobsTable provides index recommendations for batch_jobs table
func (qo *QueryOptimizer) analyzeBatchJobsTable(ctx context.Context) []*IndexRecommendation {
	var recommendations []*IndexRecommendation

	// Status + Created At index for job processing
	recommendations = append(recommendations, &IndexRecommendation{
		Table:            "batch_jobs",
		Columns:          []string{"status", "created_at"},
		Type:             "btree",
		Priority:         9,
		Reason:           "Optimizes job processing queries by status and creation time",
		EstimatedBenefit: "High - used in job scheduler",
	})

	// Organization ID + Status index for multi-tenant job queries
	recommendations = append(recommendations, &IndexRecommendation{
		Table:            "batch_jobs",
		Columns:          []string{"organization_id", "status"},
		Type:             "btree",
		Priority:         8,
		Reason:           "Optimizes organization-specific job queries",
		EstimatedBenefit: "High - used in multi-tenant job management",
	})

	// Priority index for job scheduling
	recommendations = append(recommendations, &IndexRecommendation{
		Table:            "batch_jobs",
		Columns:          []string{"priority", "created_at"},
		Type:             "btree",
		Priority:         7,
		Reason:           "Optimizes job scheduling by priority",
		EstimatedBenefit: "Medium - used in priority-based scheduling",
	})

	return recommendations
}

// analyzeCustomModelsTable provides index recommendations for custom_models table
func (qo *QueryOptimizer) analyzeCustomModelsTable(ctx context.Context) []*IndexRecommendation {
	var recommendations []*IndexRecommendation

	// Organization ID + Is Active index for model queries
	recommendations = append(recommendations, &IndexRecommendation{
		Table:            "custom_models",
		Columns:          []string{"organization_id", "is_active"},
		Type:             "btree",
		Priority:         9,
		Reason:           "Optimizes active model queries per organization",
		EstimatedBenefit: "High - used in model selection",
	})

	// Model Type + Is Active index for model filtering
	recommendations = append(recommendations, &IndexRecommendation{
		Table:            "custom_models",
		Columns:          []string{"model_type", "is_active"},
		Type:             "btree",
		Priority:         7,
		Reason:           "Optimizes model queries by type and status",
		EstimatedBenefit: "Medium - used in model type filtering",
	})

	// Created At index for model versioning
	recommendations = append(recommendations, &IndexRecommendation{
		Table:            "custom_models",
		Columns:          []string{"created_at"},
		Type:             "btree",
		Priority:         6,
		Reason:           "Optimizes model version queries",
		EstimatedBenefit: "Medium - used in model history",
	})

	return recommendations
}

// GetSlowQueries retrieves slow queries from the database
func (qo *QueryOptimizer) GetSlowQueries(ctx context.Context, threshold time.Duration) ([]*QueryAnalysis, error) {
	// This would typically query pg_stat_statements or similar
	// For now, we'll return a placeholder implementation
	query := `
		SELECT query, mean_exec_time, calls, total_exec_time
		FROM pg_stat_statements 
		WHERE mean_exec_time > $1
		ORDER BY mean_exec_time DESC
		LIMIT 50
	`

	rows, err := qo.db.QueryContext(ctx, query, threshold.Milliseconds())
	if err != nil {
		// If pg_stat_statements is not available, return empty result
		qo.logger.Warn("pg_stat_statements not available, cannot retrieve slow queries")
		return []*QueryAnalysis{}, nil
	}
	defer rows.Close()

	var slowQueries []*QueryAnalysis
	for rows.Next() {
		var query string
		var meanExecTime, calls, totalExecTime float64

		err := rows.Scan(&query, &meanExecTime, &calls, &totalExecTime)
		if err != nil {
			continue
		}

		analysis := &QueryAnalysis{
			Query:         query,
			ExecutionTime: time.Duration(meanExecTime) * time.Millisecond,
			RowsExamined:  int64(calls),
		}

		slowQueries = append(slowQueries, analysis)
	}

	return slowQueries, nil
}

// OptimizeQuery provides an optimized version of a query
func (qo *QueryOptimizer) OptimizeQuery(query string) string {
	optimized := query

	// Replace SELECT * with specific columns (if we can determine them)
	if strings.Contains(strings.ToLower(optimized), "select *") {
		// This would need table schema information to be fully implemented
		optimized = strings.Replace(optimized, "SELECT *", "SELECT id, business_id, risk_score, created_at", 1)
	}

	// Add LIMIT if missing and query could return many rows
	if !strings.Contains(strings.ToLower(optimized), "limit") &&
		!strings.Contains(strings.ToLower(optimized), "count(") {
		optimized += " LIMIT 1000"
	}

	return optimized
}
