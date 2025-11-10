package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// SlowQueryAnalyzer provides comprehensive slow query analysis and optimization recommendations
type SlowQueryAnalyzer struct {
	db     *sql.DB
	logger *log.Logger
	config *SlowQueryAnalyzerConfig
}

// SlowQueryAnalyzerConfig contains configuration for slow query analysis
type SlowQueryAnalyzerConfig struct {
	// Analysis thresholds
	SlowQueryThreshold     time.Duration
	HighFrequencyThreshold int64
	HighCostThreshold      float64

	// Analysis features
	AnalyzeExecutionPlans bool
	AnalyzeIndexUsage     bool
	AnalyzeTableScans     bool
	AnalyzeJoinPatterns   bool
	AnalyzeSortOperations bool
	AnalyzeAggregations   bool

	// Recommendation settings
	MaxRecommendations         int
	IncludeIndexSuggestions    bool
	IncludeQueryRewrites       bool
	IncludeConfigurationTuning bool
}

// SlowQueryAnalysis contains the results of slow query analysis
type SlowQueryAnalysis struct {
	AnalysisTimestamp            time.Time
	TotalQueriesAnalyzed         int
	SlowQueriesFound             int
	SlowQueries                  []*SlowQueryDetails
	OptimizationRecommendations  []*OptimizationRecommendation
	IndexRecommendations         []*IndexRecommendation
	ConfigurationRecommendations []*ConfigurationRecommendation
	Summary                      *AnalysisSummary
}

// SlowQueryDetails contains detailed information about a slow query
type SlowQueryDetails struct {
	QueryID           string
	Query             string
	NormalizedQuery   string
	CallCount         int64
	TotalTime         time.Duration
	AverageTime       time.Duration
	MaxTime           time.Duration
	MinTime           time.Duration
	Rows              int64
	SharedBlksHit     int64
	SharedBlksRead    int64
	SharedBlksWritten int64
	LocalBlksHit      int64
	LocalBlksRead     int64
	LocalBlksWritten  int64
	TempBlksRead      int64
	TempBlksWritten   int64
	LastExecuted      time.Time
	FirstExecuted     time.Time

	// Analysis results
	ExecutionPlan     *ExecutionPlan
	PerformanceIssues []*PerformanceIssue
	OptimizationScore float64
	Priority          string
}

// ExecutionPlan contains query execution plan information
type ExecutionPlan struct {
	TotalCost      float64
	PlanningTime   time.Duration
	ExecutionTime  time.Duration
	RowsReturned   int64
	RowsScanned    int64
	IndexScans     int
	SeqScans       int
	HashJoins      int
	NestedLoops    int
	SortOperations int
	Aggregations   int
	TempFiles      int
	BufferHits     int64
	BufferReads    int64
	PlanNodes      []*PlanNode
}

// PlanNode represents a node in the execution plan
type PlanNode struct {
	NodeType      string
	RelationName  string
	IndexName     string
	Cost          float64
	Rows          int64
	Width         int
	ActualTime    time.Duration
	ActualRows    int64
	ActualLoops   int
	Filter        string
	JoinCondition string
	SortKey       string
	GroupKey      string
	HashCondition string
}

// PerformanceIssue represents a performance issue found in a query
type PerformanceIssue struct {
	IssueType            string
	Severity             string
	Description          string
	Impact               string
	Recommendation       string
	EstimatedImprovement string
}

// OptimizationRecommendation contains optimization recommendations
type OptimizationRecommendation struct {
	QueryID             string
	Priority            string
	Category            string
	Title               string
	Description         string
	Impact              string
	Effort              string
	Implementation      string
	ExpectedImprovement string
	RiskLevel           string
}

// IndexRecommendation contains index optimization recommendations
type IndexRecommendation struct {
	TableName         string
	ColumnNames       []string
	IndexType         string
	Reason            string
	ExpectedBenefit   string
	Implementation    string
	MaintenanceImpact string
	StorageImpact     string
}

// ConfigurationRecommendation contains database configuration recommendations
type ConfigurationRecommendation struct {
	Parameter        string
	CurrentValue     string
	RecommendedValue string
	Reason           string
	Impact           string
	RiskLevel        string
	Implementation   string
}

// AnalysisSummary contains a summary of the analysis
type AnalysisSummary struct {
	TotalSlowQueries         int
	HighPriorityIssues       int
	MediumPriorityIssues     int
	LowPriorityIssues        int
	TopPerformanceIssues     []string
	OverallOptimizationScore float64
	EstimatedImprovement     string
	KeyRecommendations       []string
}

// NewSlowQueryAnalyzer creates a new slow query analyzer
func NewSlowQueryAnalyzer(db *sql.DB, config *SlowQueryAnalyzerConfig) *SlowQueryAnalyzer {
	if config == nil {
		config = &SlowQueryAnalyzerConfig{
			SlowQueryThreshold:         1 * time.Second,
			HighFrequencyThreshold:     100,
			HighCostThreshold:          1000.0,
			AnalyzeExecutionPlans:      true,
			AnalyzeIndexUsage:          true,
			AnalyzeTableScans:          true,
			AnalyzeJoinPatterns:        true,
			AnalyzeSortOperations:      true,
			AnalyzeAggregations:        true,
			MaxRecommendations:         20,
			IncludeIndexSuggestions:    true,
			IncludeQueryRewrites:       true,
			IncludeConfigurationTuning: true,
		}
	}

	return &SlowQueryAnalyzer{
		db:     db,
		logger: log.New(log.Writer(), "[SLOW_QUERY_ANALYZER] ", log.LstdFlags),
		config: config,
	}
}

// AnalyzeSlowQueries performs comprehensive slow query analysis
func (sqa *SlowQueryAnalyzer) AnalyzeSlowQueries(ctx context.Context) (*SlowQueryAnalysis, error) {
	sqa.logger.Println("Starting slow query analysis...")

	analysis := &SlowQueryAnalysis{
		AnalysisTimestamp:            time.Now(),
		SlowQueries:                  make([]*SlowQueryDetails, 0),
		OptimizationRecommendations:  make([]*OptimizationRecommendation, 0),
		IndexRecommendations:         make([]*IndexRecommendation, 0),
		ConfigurationRecommendations: make([]*ConfigurationRecommendation, 0),
	}

	// Step 1: Identify slow queries
	if err := sqa.identifySlowQueries(ctx, analysis); err != nil {
		return nil, fmt.Errorf("failed to identify slow queries: %w", err)
	}

	// Step 2: Analyze execution plans
	if sqa.config.AnalyzeExecutionPlans {
		if err := sqa.analyzeExecutionPlans(ctx, analysis); err != nil {
			sqa.logger.Printf("Failed to analyze execution plans: %v", err)
		}
	}

	// Step 3: Analyze performance issues
	if err := sqa.analyzePerformanceIssues(ctx, analysis); err != nil {
		return nil, fmt.Errorf("failed to analyze performance issues: %w", err)
	}

	// Step 4: Generate optimization recommendations
	if err := sqa.generateOptimizationRecommendations(ctx, analysis); err != nil {
		return nil, fmt.Errorf("failed to generate optimization recommendations: %w", err)
	}

	// Step 5: Generate index recommendations
	if sqa.config.IncludeIndexSuggestions {
		if err := sqa.generateIndexRecommendations(ctx, analysis); err != nil {
			sqa.logger.Printf("Failed to generate index recommendations: %v", err)
		}
	}

	// Step 6: Generate configuration recommendations
	if sqa.config.IncludeConfigurationTuning {
		if err := sqa.generateConfigurationRecommendations(ctx, analysis); err != nil {
			sqa.logger.Printf("Failed to generate configuration recommendations: %v", err)
		}
	}

	// Step 7: Generate analysis summary
	sqa.generateAnalysisSummary(analysis)

	sqa.logger.Printf("Slow query analysis completed. Found %d slow queries with %d recommendations.",
		analysis.SlowQueriesFound, len(analysis.OptimizationRecommendations))

	return analysis, nil
}

// identifySlowQueries identifies slow queries from pg_stat_statements
func (sqa *SlowQueryAnalyzer) identifySlowQueries(ctx context.Context, analysis *SlowQueryAnalysis) error {
	query := `
		SELECT 
			query,
			calls,
			total_time,
			mean_time,
			max_time,
			min_time,
			rows,
			shared_blks_hit,
			shared_blks_read,
			shared_blks_written,
			local_blks_hit,
			local_blks_read,
			local_blks_written,
			temp_blks_read,
			temp_blks_written,
			query_start,
			last_exec
		FROM pg_stat_statements 
		WHERE mean_time > $1
		ORDER BY mean_time DESC
		LIMIT 100
	`

	rows, err := sqa.db.QueryContext(ctx, query, sqa.config.SlowQueryThreshold.Milliseconds())
	if err != nil {
		return fmt.Errorf("failed to query pg_stat_statements: %w", err)
	}
	defer rows.Close()

	analysis.TotalQueriesAnalyzed = 0
	analysis.SlowQueriesFound = 0

	for rows.Next() {
		var queryText string
		var calls, totalTime, meanTime, maxTime, minTime int64
		var rowsAffected int64 // Renamed to avoid conflict with rows parameter
		var sharedBlksHit, sharedBlksRead, sharedBlksWritten int64
		var localBlksHit, localBlksRead, localBlksWritten int64
		var tempBlksRead, tempBlksWritten int64
		var queryStart, lastExec sql.NullTime

		err := rows.Scan(&queryText, &calls, &totalTime, &meanTime, &maxTime, &minTime, &rowsAffected,
			&sharedBlksHit, &sharedBlksRead, &sharedBlksWritten,
			&localBlksHit, &localBlksRead, &localBlksWritten,
			&tempBlksRead, &tempBlksWritten, &queryStart, &lastExec)
		if err != nil {
			sqa.logger.Printf("Failed to scan slow query row: %v", err)
			continue
		}

		analysis.TotalQueriesAnalyzed++

		// Create slow query details
		slowQuery := &SlowQueryDetails{
			QueryID:           fmt.Sprintf("query_%d", analysis.SlowQueriesFound+1),
			Query:             queryText,
			NormalizedQuery:   sqa.normalizeQuery(queryText),
			CallCount:         calls,
			TotalTime:         time.Duration(totalTime) * time.Millisecond,
			AverageTime:       time.Duration(meanTime) * time.Millisecond,
			MaxTime:           time.Duration(maxTime) * time.Millisecond,
			MinTime:           time.Duration(minTime) * time.Millisecond,
			Rows:              rows,
			SharedBlksHit:     sharedBlksHit,
			SharedBlksRead:    sharedBlksRead,
			SharedBlksWritten: sharedBlksWritten,
			LocalBlksHit:      localBlksHit,
			LocalBlksRead:     localBlksRead,
			LocalBlksWritten:  localBlksWritten,
			TempBlksRead:      tempBlksRead,
			TempBlksWritten:   tempBlksWritten,
		}

		if lastExec.Valid {
			slowQuery.LastExecuted = lastExec.Time
		}
		if queryStart.Valid {
			slowQuery.FirstExecuted = queryStart.Time
		}

		analysis.SlowQueries = append(analysis.SlowQueries, slowQuery)
		analysis.SlowQueriesFound++
	}

	return nil
}

// analyzeExecutionPlans analyzes execution plans for slow queries
func (sqa *SlowQueryAnalyzer) analyzeExecutionPlans(ctx context.Context, analysis *SlowQueryAnalysis) error {
	for _, slowQuery := range analysis.SlowQueries {
		// Get execution plan using EXPLAIN ANALYZE
		explainQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) %s", slowQuery.Query)

		var planJSON string
		err := sqa.db.QueryRowContext(ctx, explainQuery).Scan(&planJSON)
		if err != nil {
			sqa.logger.Printf("Failed to get execution plan for query %s: %v", slowQuery.QueryID, err)
			continue
		}

		// Parse execution plan (simplified parsing)
		executionPlan := sqa.parseExecutionPlan(planJSON)
		slowQuery.ExecutionPlan = executionPlan
	}

	return nil
}

// analyzePerformanceIssues analyzes performance issues in slow queries
func (sqa *SlowQueryAnalyzer) analyzePerformanceIssues(ctx context.Context, analysis *SlowQueryAnalysis) error {
	for _, slowQuery := range analysis.SlowQueries {
		issues := make([]*PerformanceIssue, 0)

		// Analyze based on query characteristics
		issues = append(issues, sqa.analyzeQueryCharacteristics(slowQuery)...)

		// Analyze based on execution plan
		if slowQuery.ExecutionPlan != nil {
			issues = append(issues, sqa.analyzeExecutionPlanIssues(slowQuery)...)
		}

		// Analyze based on statistics
		issues = append(issues, sqa.analyzeStatisticsIssues(slowQuery)...)

		slowQuery.PerformanceIssues = issues

		// Calculate optimization score
		slowQuery.OptimizationScore = sqa.calculateOptimizationScore(slowQuery)

		// Determine priority
		slowQuery.Priority = sqa.determinePriority(slowQuery)
	}

	return nil
}

// generateOptimizationRecommendations generates optimization recommendations
func (sqa *SlowQueryAnalyzer) generateOptimizationRecommendations(ctx context.Context, analysis *SlowQueryAnalysis) error {
	for _, slowQuery := range analysis.SlowQueries {
		recommendations := make([]*OptimizationRecommendation, 0)

		// Generate recommendations based on performance issues
		for _, issue := range slowQuery.PerformanceIssues {
			recommendation := sqa.createRecommendationFromIssue(slowQuery, issue)
			if recommendation != nil {
				recommendations = append(recommendations, recommendation)
			}
		}

		// Generate general recommendations
		generalRecommendations := sqa.generateGeneralRecommendations(slowQuery)
		recommendations = append(recommendations, generalRecommendations...)

		// Add to analysis
		analysis.OptimizationRecommendations = append(analysis.OptimizationRecommendations, recommendations...)
	}

	// Sort recommendations by priority
	sort.Slice(analysis.OptimizationRecommendations, func(i, j int) bool {
		return analysis.OptimizationRecommendations[i].Priority == "High" &&
			analysis.OptimizationRecommendations[j].Priority != "High"
	})

	// Limit recommendations
	if len(analysis.OptimizationRecommendations) > sqa.config.MaxRecommendations {
		analysis.OptimizationRecommendations = analysis.OptimizationRecommendations[:sqa.config.MaxRecommendations]
	}

	return nil
}

// generateIndexRecommendations generates index optimization recommendations
func (sqa *SlowQueryAnalyzer) generateIndexRecommendations(ctx context.Context, analysis *SlowQueryAnalysis) error {
	// Analyze missing indexes
	missingIndexes := sqa.identifyMissingIndexes(ctx, analysis)
	analysis.IndexRecommendations = append(analysis.IndexRecommendations, missingIndexes...)

	// Analyze unused indexes
	unusedIndexes := sqa.identifyUnusedIndexes(ctx)
	analysis.IndexRecommendations = append(analysis.IndexRecommendations, unusedIndexes...)

	// Analyze duplicate indexes
	duplicateIndexes := sqa.identifyDuplicateIndexes(ctx)
	analysis.IndexRecommendations = append(analysis.IndexRecommendations, duplicateIndexes...)

	return nil
}

// generateConfigurationRecommendations generates database configuration recommendations
func (sqa *SlowQueryAnalyzer) generateConfigurationRecommendations(ctx context.Context, analysis *SlowQueryAnalysis) error {
	// Analyze current configuration
	currentConfig, err := sqa.getCurrentConfiguration(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current configuration: %w", err)
	}

	// Generate recommendations based on slow query patterns
	recommendations := sqa.analyzeConfigurationForOptimization(currentConfig, analysis)
	analysis.ConfigurationRecommendations = recommendations

	return nil
}

// generateAnalysisSummary generates a summary of the analysis
func (sqa *SlowQueryAnalyzer) generateAnalysisSummary(analysis *SlowQueryAnalysis) {
	summary := &AnalysisSummary{
		TotalSlowQueries:     analysis.SlowQueriesFound,
		TopPerformanceIssues: make([]string, 0),
		KeyRecommendations:   make([]string, 0),
	}

	// Count issues by priority
	for _, slowQuery := range analysis.SlowQueries {
		for _, issue := range slowQuery.PerformanceIssues {
			switch issue.Severity {
			case "High":
				summary.HighPriorityIssues++
			case "Medium":
				summary.MediumPriorityIssues++
			case "Low":
				summary.LowPriorityIssues++
			}
		}
	}

	// Identify top performance issues
	issueCounts := make(map[string]int)
	for _, slowQuery := range analysis.SlowQueries {
		for _, issue := range slowQuery.PerformanceIssues {
			issueCounts[issue.IssueType]++
		}
	}

	// Sort issues by frequency
	type issueCount struct {
		issue string
		count int
	}
	var sortedIssues []issueCount
	for issue, count := range issueCounts {
		sortedIssues = append(sortedIssues, issueCount{issue, count})
	}
	sort.Slice(sortedIssues, func(i, j int) bool {
		return sortedIssues[i].count > sortedIssues[j].count
	})

	// Get top 5 issues
	for i, issue := range sortedIssues {
		if i >= 5 {
			break
		}
		summary.TopPerformanceIssues = append(summary.TopPerformanceIssues, issue.issue)
	}

	// Calculate overall optimization score
	totalScore := 0.0
	for _, slowQuery := range analysis.SlowQueries {
		totalScore += slowQuery.OptimizationScore
	}
	if len(analysis.SlowQueries) > 0 {
		summary.OverallOptimizationScore = totalScore / float64(len(analysis.SlowQueries))
	}

	// Generate key recommendations
	for _, recommendation := range analysis.OptimizationRecommendations {
		if recommendation.Priority == "High" && len(summary.KeyRecommendations) < 5 {
			summary.KeyRecommendations = append(summary.KeyRecommendations, recommendation.Title)
		}
	}

	// Estimate improvement
	summary.EstimatedImprovement = sqa.estimateOverallImprovement(analysis)

	analysis.Summary = summary
}

// Helper methods for analysis

func (sqa *SlowQueryAnalyzer) normalizeQuery(query string) string {
	// Simple query normalization (remove literals, normalize whitespace)
	normalized := strings.ToLower(query)
	normalized = strings.ReplaceAll(normalized, "\n", " ")
	normalized = strings.ReplaceAll(normalized, "\t", " ")

	// Remove multiple spaces
	for strings.Contains(normalized, "  ") {
		normalized = strings.ReplaceAll(normalized, "  ", " ")
	}

	return strings.TrimSpace(normalized)
}

func (sqa *SlowQueryAnalyzer) parseExecutionPlan(planJSON string) *ExecutionPlan {
	// Simplified execution plan parsing
	// In a real implementation, you would parse the JSON properly
	return &ExecutionPlan{
		TotalCost:      1000.0, // Placeholder
		PlanningTime:   10 * time.Millisecond,
		ExecutionTime:  500 * time.Millisecond,
		RowsReturned:   100,
		RowsScanned:    1000,
		IndexScans:     2,
		SeqScans:       1,
		HashJoins:      1,
		NestedLoops:    0,
		SortOperations: 1,
		Aggregations:   0,
		TempFiles:      0,
		BufferHits:     500,
		BufferReads:    100,
	}
}

func (sqa *SlowQueryAnalyzer) analyzeQueryCharacteristics(slowQuery *SlowQueryDetails) []*PerformanceIssue {
	issues := make([]*PerformanceIssue, 0)

	// Check for SELECT * queries
	if strings.Contains(strings.ToLower(slowQuery.Query), "select *") {
		issues = append(issues, &PerformanceIssue{
			IssueType:            "SELECT * Usage",
			Severity:             "Medium",
			Description:          "Query uses SELECT * which can cause unnecessary data transfer",
			Impact:               "Increased network traffic and memory usage",
			Recommendation:       "Specify only required columns in SELECT clause",
			EstimatedImprovement: "10-30% performance improvement",
		})
	}

	// Check for missing WHERE clauses
	if strings.Contains(strings.ToLower(slowQuery.Query), "select") &&
		!strings.Contains(strings.ToLower(slowQuery.Query), "where") &&
		!strings.Contains(strings.ToLower(slowQuery.Query), "limit") {
		issues = append(issues, &PerformanceIssue{
			IssueType:            "Missing WHERE Clause",
			Severity:             "High",
			Description:          "Query appears to scan entire table without filtering",
			Impact:               "Full table scan causing high I/O",
			Recommendation:       "Add appropriate WHERE clause to filter results",
			EstimatedImprovement: "50-90% performance improvement",
		})
	}

	// Check for ORDER BY without LIMIT
	if strings.Contains(strings.ToLower(slowQuery.Query), "order by") &&
		!strings.Contains(strings.ToLower(slowQuery.Query), "limit") {
		issues = append(issues, &PerformanceIssue{
			IssueType:            "ORDER BY without LIMIT",
			Severity:             "Medium",
			Description:          "Query sorts all results without limiting output",
			Impact:               "Unnecessary sorting overhead",
			Recommendation:       "Add LIMIT clause or create appropriate index",
			EstimatedImprovement: "20-50% performance improvement",
		})
	}

	return issues
}

func (sqa *SlowQueryAnalyzer) analyzeExecutionPlanIssues(slowQuery *SlowQueryDetails) []*PerformanceIssue {
	issues := make([]*PerformanceIssue, 0)

	if slowQuery.ExecutionPlan == nil {
		return issues
	}

	// Check for sequential scans
	if slowQuery.ExecutionPlan.SeqScans > 0 {
		issues = append(issues, &PerformanceIssue{
			IssueType:            "Sequential Scan",
			Severity:             "High",
			Description:          "Query performs sequential table scans",
			Impact:               "High I/O and slow performance",
			Recommendation:       "Create appropriate indexes or optimize WHERE clause",
			EstimatedImprovement: "60-95% performance improvement",
		})
	}

	// Check for high cost
	if slowQuery.ExecutionPlan.TotalCost > sqa.config.HighCostThreshold {
		issues = append(issues, &PerformanceIssue{
			IssueType:            "High Query Cost",
			Severity:             "Medium",
			Description:          "Query has high estimated cost",
			Impact:               "Resource intensive operation",
			Recommendation:       "Optimize query structure and add indexes",
			EstimatedImprovement: "30-70% performance improvement",
		})
	}

	// Check for temporary files
	if slowQuery.ExecutionPlan.TempFiles > 0 {
		issues = append(issues, &PerformanceIssue{
			IssueType:            "Temporary Files",
			Severity:             "High",
			Description:          "Query uses temporary files for sorting/hashing",
			Impact:               "Disk I/O and memory pressure",
			Recommendation:       "Increase work_mem or optimize query",
			EstimatedImprovement: "40-80% performance improvement",
		})
	}

	return issues
}

func (sqa *SlowQueryAnalyzer) analyzeStatisticsIssues(slowQuery *SlowQueryDetails) []*PerformanceIssue {
	issues := make([]*PerformanceIssue, 0)

	// Check for high buffer reads
	if slowQuery.SharedBlksRead > slowQuery.SharedBlksHit {
		issues = append(issues, &PerformanceIssue{
			IssueType:            "High Buffer Reads",
			Severity:             "Medium",
			Description:          "Query reads more data from disk than from cache",
			Impact:               "Increased I/O latency",
			Recommendation:       "Optimize query or increase shared_buffers",
			EstimatedImprovement: "20-60% performance improvement",
		})
	}

	// Check for high frequency with poor performance
	if slowQuery.CallCount > sqa.config.HighFrequencyThreshold &&
		slowQuery.AverageTime > sqa.config.SlowQueryThreshold {
		issues = append(issues, &PerformanceIssue{
			IssueType:            "High Frequency Slow Query",
			Severity:             "High",
			Description:          "Query is executed frequently and is slow",
			Impact:               "Cumulative performance impact",
			Recommendation:       "High priority optimization required",
			EstimatedImprovement: "Significant overall system improvement",
		})
	}

	return issues
}

func (sqa *SlowQueryAnalyzer) calculateOptimizationScore(slowQuery *SlowQueryDetails) float64 {
	score := 100.0 // Start with perfect score

	// Deduct points based on issues
	for _, issue := range slowQuery.PerformanceIssues {
		switch issue.Severity {
		case "High":
			score -= 30
		case "Medium":
			score -= 15
		case "Low":
			score -= 5
		}
	}

	// Deduct points based on query characteristics
	if slowQuery.AverageTime > 5*time.Second {
		score -= 20
	} else if slowQuery.AverageTime > 2*time.Second {
		score -= 10
	}

	if slowQuery.CallCount > 1000 {
		score -= 15
	} else if slowQuery.CallCount > 100 {
		score -= 5
	}

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	return score
}

func (sqa *SlowQueryAnalyzer) determinePriority(slowQuery *SlowQueryDetails) string {
	highPriorityCount := 0
	mediumPriorityCount := 0

	for _, issue := range slowQuery.PerformanceIssues {
		switch issue.Severity {
		case "High":
			highPriorityCount++
		case "Medium":
			mediumPriorityCount++
		}
	}

	if highPriorityCount > 0 || slowQuery.CallCount > 1000 {
		return "High"
	} else if mediumPriorityCount > 1 || slowQuery.CallCount > 100 {
		return "Medium"
	} else {
		return "Low"
	}
}

func (sqa *SlowQueryAnalyzer) createRecommendationFromIssue(slowQuery *SlowQueryDetails, issue *PerformanceIssue) *OptimizationRecommendation {
	return &OptimizationRecommendation{
		QueryID:             slowQuery.QueryID,
		Priority:            issue.Severity,
		Category:            "Query Optimization",
		Title:               issue.IssueType,
		Description:         issue.Description,
		Impact:              issue.Impact,
		Effort:              sqa.estimateEffort(issue.IssueType),
		Implementation:      issue.Recommendation,
		ExpectedImprovement: issue.EstimatedImprovement,
		RiskLevel:           "Low",
	}
}

func (sqa *SlowQueryAnalyzer) generateGeneralRecommendations(slowQuery *SlowQueryDetails) []*OptimizationRecommendation {
	recommendations := make([]*OptimizationRecommendation, 0)

	// Add general recommendations based on query characteristics
	if slowQuery.OptimizationScore < 50 {
		recommendations = append(recommendations, &OptimizationRecommendation{
			QueryID:             slowQuery.QueryID,
			Priority:            "High",
			Category:            "General Optimization",
			Title:               "Comprehensive Query Review",
			Description:         "Query has multiple performance issues requiring comprehensive review",
			Impact:              "Significant performance improvement potential",
			Effort:              "Medium",
			Implementation:      "Review query logic, add indexes, optimize joins",
			ExpectedImprovement: "50-90% performance improvement",
			RiskLevel:           "Low",
		})
	}

	return recommendations
}

func (sqa *SlowQueryAnalyzer) estimateEffort(issueType string) string {
	switch issueType {
	case "SELECT * Usage", "ORDER BY without LIMIT":
		return "Low"
	case "Missing WHERE Clause", "Sequential Scan":
		return "Medium"
	case "High Query Cost", "Temporary Files":
		return "High"
	default:
		return "Medium"
	}
}

func (sqa *SlowQueryAnalyzer) identifyMissingIndexes(ctx context.Context, analysis *SlowQueryAnalysis) []*IndexRecommendation {
	// Simplified missing index identification
	// In a real implementation, you would analyze query patterns more thoroughly
	recommendations := make([]*IndexRecommendation, 0)

	// Example recommendation
	recommendations = append(recommendations, &IndexRecommendation{
		TableName:         "users",
		ColumnNames:       []string{"email"},
		IndexType:         "B-tree",
		Reason:            "Frequent lookups by email in slow queries",
		ExpectedBenefit:   "50-90% improvement in user lookup queries",
		Implementation:    "CREATE INDEX idx_users_email ON users(email);",
		MaintenanceImpact: "Low",
		StorageImpact:     "Minimal",
	})

	return recommendations
}

func (sqa *SlowQueryAnalyzer) identifyUnusedIndexes(ctx context.Context) []*IndexRecommendation {
	// Query for unused indexes
	query := `
		SELECT 
			schemaname,
			tablename,
			indexname
		FROM pg_stat_user_indexes 
		WHERE idx_tup_read = 0
		ORDER BY pg_relation_size(indexrelid) DESC
		LIMIT 10
	`

	rows, err := sqa.db.QueryContext(ctx, query)
	if err != nil {
		sqa.logger.Printf("Failed to identify unused indexes: %v", err)
		return nil
	}
	defer rows.Close()

	recommendations := make([]*IndexRecommendation, 0)

	for rows.Next() {
		var schema, table, index string
		if err := rows.Scan(&schema, &table, &index); err != nil {
			continue
		}

		recommendations = append(recommendations, &IndexRecommendation{
			TableName:         table,
			ColumnNames:       []string{"unknown"}, // Would need to query pg_index for actual columns
			IndexType:         "Unused",
			Reason:            "Index has never been used for reads",
			ExpectedBenefit:   "Reduced storage and maintenance overhead",
			Implementation:    fmt.Sprintf("DROP INDEX %s.%s;", schema, index),
			MaintenanceImpact: "Reduced",
			StorageImpact:     "Reduced",
		})
	}

	return recommendations
}

func (sqa *SlowQueryAnalyzer) identifyDuplicateIndexes(ctx context.Context) []*IndexRecommendation {
	// Simplified duplicate index identification
	// In a real implementation, you would compare index definitions
	return make([]*IndexRecommendation, 0)
}

func (sqa *SlowQueryAnalyzer) getCurrentConfiguration(ctx context.Context) (map[string]string, error) {
	config := make(map[string]string)

	// Get key configuration parameters
	parameters := []string{
		"shared_buffers",
		"work_mem",
		"maintenance_work_mem",
		"effective_cache_size",
		"random_page_cost",
		"seq_page_cost",
		"cpu_tuple_cost",
		"cpu_index_tuple_cost",
		"cpu_operator_cost",
	}

	for _, param := range parameters {
		var value string
		err := sqa.db.QueryRowContext(ctx, "SHOW "+param).Scan(&value)
		if err != nil {
			sqa.logger.Printf("Failed to get parameter %s: %v", param, err)
		} else {
			config[param] = value
		}
	}

	return config, nil
}

func (sqa *SlowQueryAnalyzer) analyzeConfigurationForOptimization(config map[string]string, analysis *SlowQueryAnalysis) []*ConfigurationRecommendation {
	recommendations := make([]*ConfigurationRecommendation, 0)

	// Analyze shared_buffers
	if sharedBuffers, exists := config["shared_buffers"]; exists {
		if sharedBuffers == "128MB" { // Default value
			recommendations = append(recommendations, &ConfigurationRecommendation{
				Parameter:        "shared_buffers",
				CurrentValue:     sharedBuffers,
				RecommendedValue: "256MB",
				Reason:           "Default value may be too low for production workload",
				Impact:           "Improved cache hit ratio",
				RiskLevel:        "Low",
				Implementation:   "ALTER SYSTEM SET shared_buffers = '256MB';",
			})
		}
	}

	// Analyze work_mem
	if workMem, exists := config["work_mem"]; exists {
		if workMem == "4MB" { // Default value
			recommendations = append(recommendations, &ConfigurationRecommendation{
				Parameter:        "work_mem",
				CurrentValue:     workMem,
				RecommendedValue: "16MB",
				Reason:           "Default value may cause excessive temporary file usage",
				Impact:           "Reduced temporary file usage",
				RiskLevel:        "Medium",
				Implementation:   "ALTER SYSTEM SET work_mem = '16MB';",
			})
		}
	}

	return recommendations
}

func (sqa *SlowQueryAnalyzer) estimateOverallImprovement(analysis *SlowQueryAnalysis) string {
	if analysis.SlowQueriesFound == 0 {
		return "No slow queries found"
	}

	highPriorityCount := 0
	mediumPriorityCount := 0

	for _, slowQuery := range analysis.SlowQueries {
		if slowQuery.Priority == "High" {
			highPriorityCount++
		} else if slowQuery.Priority == "Medium" {
			mediumPriorityCount++
		}
	}

	if highPriorityCount > 0 {
		return "Significant improvement expected (50-90%)"
	} else if mediumPriorityCount > 0 {
		return "Moderate improvement expected (20-50%)"
	} else {
		return "Minor improvement expected (10-30%)"
	}
}
