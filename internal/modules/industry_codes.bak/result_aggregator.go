package industry_codes

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ResultAggregator provides result aggregation and presentation capabilities
type ResultAggregator struct {
	confidenceScorer *ConfidenceScorer
	rankingEngine    *RankingEngine
	logger           *zap.Logger
}

// NewResultAggregator creates a new result aggregator
func NewResultAggregator(confidenceScorer *ConfidenceScorer, rankingEngine *RankingEngine, logger *zap.Logger) *ResultAggregator {
	return &ResultAggregator{
		confidenceScorer: confidenceScorer,
		rankingEngine:    rankingEngine,
		logger:           logger,
	}
}

// AggregationRequest represents a request for result aggregation
type AggregationRequest struct {
	Results           []*ClassificationResult `json:"results"`
	MaxResultsPerType int                     `json:"max_results_per_type"`
	MinConfidence     float64                 `json:"min_confidence"`
	IncludeMetadata   bool                    `json:"include_metadata"`
	IncludeAnalytics  bool                    `json:"include_analytics"`
	GroupByStrategy   bool                    `json:"group_by_strategy"`
	SortBy            SortCriteria            `json:"sort_by"`
	Presentation      PresentationFormat      `json:"presentation"`
}

// SortCriteria defines how results should be sorted
type SortCriteria string

const (
	SortByConfidence    SortCriteria = "confidence"
	SortByRelevance     SortCriteria = "relevance"
	SortByQuality       SortCriteria = "quality"
	SortByAlphabetical  SortCriteria = "alphabetical"
	SortByCodeType      SortCriteria = "code_type"
	SortByMatchStrength SortCriteria = "match_strength"
)

// PresentationFormat defines the output format
type PresentationFormat string

const (
	PresentationDetailed  PresentationFormat = "detailed"
	PresentationSummary   PresentationFormat = "summary"
	PresentationCompact   PresentationFormat = "compact"
	PresentationExport    PresentationFormat = "export"
	PresentationDashboard PresentationFormat = "dashboard"
	PresentationAPI       PresentationFormat = "api"
)

// AggregatedResults represents the aggregated and presented results
type AggregatedResults struct {
	// Core Results
	TopThreeByType    map[string][]*AggregatedResult `json:"top_three_by_type"`
	OverallTopResults []*AggregatedResult            `json:"overall_top_results"`
	AllResults        []*AggregatedResult            `json:"all_results"`

	// Result Organization
	ResultsByStrategy   map[string][]*AggregatedResult `json:"results_by_strategy,omitempty"`
	ResultsByConfidence map[string][]*AggregatedResult `json:"results_by_confidence,omitempty"`

	// Aggregation Metadata
	AggregationMetadata *AggregationMetadata `json:"aggregation_metadata"`

	// Analytics and Insights
	Analytics *AggregationAnalytics `json:"analytics,omitempty"`

	// Presentation Data
	PresentationData *PresentationData `json:"presentation_data,omitempty"`
}

// AggregatedResult represents a single aggregated classification result
type AggregatedResult struct {
	*ClassificationResult

	// Aggregation-specific data
	AggregationScore float64 `json:"aggregation_score"`
	TypeRank         int     `json:"type_rank"`
	OverallRank      int     `json:"overall_rank"`

	// Enhanced metadata
	ConfidenceLevel   ConfidenceLevel `json:"confidence_level"`
	QualityIndicators []string        `json:"quality_indicators"`
	MatchStrength     MatchStrength   `json:"match_strength"`

	// Cross-reference data
	RelatedCodes     []*RelatedCode     `json:"related_codes,omitempty"`
	AlternativeCodes []*AlternativeCode `json:"alternative_codes,omitempty"`

	// Presentation hints
	DisplayPriority int                    `json:"display_priority"`
	UIHints         map[string]interface{} `json:"ui_hints,omitempty"`
}

// ConfidenceLevel represents the confidence level categorization
type ConfidenceLevel string

const (
	ConfidenceLevelVeryHigh ConfidenceLevel = "very_high"
	ConfidenceLevelHigh     ConfidenceLevel = "high"
	ConfidenceLevelMedium   ConfidenceLevel = "medium"
	ConfidenceLevelLow      ConfidenceLevel = "low"
	ConfidenceLevelVeryLow  ConfidenceLevel = "very_low"
)

// MatchStrength represents the strength of the match
type MatchStrength string

const (
	MatchStrengthExact    MatchStrength = "exact"
	MatchStrengthStrong   MatchStrength = "strong"
	MatchStrengthModerate MatchStrength = "moderate"
	MatchStrengthWeak     MatchStrength = "weak"
	MatchStrengthMinimal  MatchStrength = "minimal"
)

// RelatedCode represents a related industry code
type RelatedCode struct {
	Code         *IndustryCode `json:"code"`
	Relationship string        `json:"relationship"`
	Confidence   float64       `json:"confidence"`
}

// AlternativeCode represents an alternative industry code suggestion
type AlternativeCode struct {
	Code        *IndustryCode `json:"code"`
	Reason      string        `json:"reason"`
	Confidence  float64       `json:"confidence"`
	Recommended bool          `json:"recommended"`
}

// AggregationMetadata provides metadata about the aggregation process
type AggregationMetadata struct {
	AggregationTime     time.Duration           `json:"aggregation_time"`
	TotalInputResults   int                     `json:"total_input_results"`
	AggregatedCount     int                     `json:"aggregated_count"`
	FilteredCount       int                     `json:"filtered_count"`
	Strategy            string                  `json:"strategy"`
	Criteria            *AggregationRequest     `json:"criteria"`
	QualityDistribution map[ConfidenceLevel]int `json:"quality_distribution"`
	TypeDistribution    map[string]int          `json:"type_distribution"`
	ProcessingSteps     []ProcessingStep        `json:"processing_steps"`
}

// ProcessingStep represents a step in the aggregation process
type ProcessingStep struct {
	Step        string                 `json:"step"`
	Description string                 `json:"description"`
	Duration    time.Duration          `json:"duration"`
	InputCount  int                    `json:"input_count"`
	OutputCount int                    `json:"output_count"`
	Success     bool                   `json:"success"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// AggregationAnalytics provides analytics about the aggregated results
type AggregationAnalytics struct {
	// Confidence Analytics
	ConfidenceStats *ConfidenceStatistics `json:"confidence_stats"`

	// Coverage Analytics
	TypeCoverage     map[string]float64 `json:"type_coverage"`
	IndustryCoverage map[string]float64 `json:"industry_coverage"`

	// Quality Analytics
	QualityMetrics *QualityAnalytics `json:"quality_metrics"`

	// Diversity Analytics
	DiversityMetrics *DiversityAnalytics `json:"diversity_metrics"`

	// Recommendation Analytics
	RecommendationScore float64 `json:"recommendation_score"`
	Certainty           float64 `json:"certainty"`

	// Comparison Analytics
	CrossTypeAnalysis *CrossTypeAnalysis `json:"cross_type_analysis,omitempty"`
}

// ConfidenceStatistics provides statistical analysis of confidence scores
type ConfidenceStatistics struct {
	Mean      float64   `json:"mean"`
	Median    float64   `json:"median"`
	Mode      float64   `json:"mode"`
	StdDev    float64   `json:"std_dev"`
	Min       float64   `json:"min"`
	Max       float64   `json:"max"`
	Range     float64   `json:"range"`
	Quartiles []float64 `json:"quartiles"`
}

// QualityAnalytics provides quality analysis of results
type QualityAnalytics struct {
	OverallQuality    float64            `json:"overall_quality"`
	QualityByType     map[string]float64 `json:"quality_by_type"`
	QualityIndicators []string           `json:"quality_indicators"`
	QualityIssues     []string           `json:"quality_issues"`
	Recommendations   []string           `json:"recommendations"`
}

// DiversityAnalytics provides diversity analysis of results
type DiversityAnalytics struct {
	TypeDiversity      float64 `json:"type_diversity"`
	CategoryDiversity  float64 `json:"category_diversity"`
	IndustrySpread     float64 `json:"industry_spread"`
	ConcentrationIndex float64 `json:"concentration_index"`
	DiversityScore     float64 `json:"diversity_score"`
}

// CrossTypeAnalysis provides analysis across different code types
type CrossTypeAnalysis struct {
	TypeCorrelations   map[string]map[string]float64 `json:"type_correlations"`
	ConsistencyScore   float64                       `json:"consistency_score"`
	ConflictingCodes   []CodeConflict                `json:"conflicting_codes"`
	RecommendedPrimary string                        `json:"recommended_primary"`
}

// CodeConflict represents a conflict between different code types
type CodeConflict struct {
	Type1        string `json:"type1"`
	Code1        string `json:"code1"`
	Type2        string `json:"type2"`
	Code2        string `json:"code2"`
	ConflictType string `json:"conflict_type"`
	Severity     string `json:"severity"`
	Resolution   string `json:"resolution"`
}

// PresentationData provides data formatted for specific presentation needs
type PresentationData struct {
	Format          PresentationFormat `json:"format"`
	Title           string             `json:"title"`
	Summary         string             `json:"summary"`
	KeyFindings     []string           `json:"key_findings"`
	Recommendations []string           `json:"recommendations"`

	// Format-specific data
	DetailedView  *DetailedPresentation  `json:"detailed_view,omitempty"`
	SummaryView   *SummaryPresentation   `json:"summary_view,omitempty"`
	CompactView   *CompactPresentation   `json:"compact_view,omitempty"`
	ExportData    *ExportPresentation    `json:"export_data,omitempty"`
	DashboardData *DashboardPresentation `json:"dashboard_data,omitempty"`
	APIResponse   *APIPresentation       `json:"api_response,omitempty"`
}

// DetailedPresentation provides detailed presentation format
type DetailedPresentation struct {
	FullResults           []*AggregatedResult `json:"full_results"`
	DetailedAnalytics     interface{}         `json:"detailed_analytics"`
	MethodologyNotes      []string            `json:"methodology_notes"`
	ConfidenceExplanation string              `json:"confidence_explanation"`
}

// SummaryPresentation provides summary presentation format
type SummaryPresentation struct {
	TopThree          []*AggregatedResult `json:"top_three"`
	KeyMetrics        map[string]float64  `json:"key_metrics"`
	QuickSummary      string              `json:"quick_summary"`
	RecommendedAction string              `json:"recommended_action"`
}

// CompactPresentation provides compact presentation format
type CompactPresentation struct {
	BestMatch           *AggregatedResult   `json:"best_match"`
	AlternativeMatches  []*AggregatedResult `json:"alternative_matches"`
	ConfidenceIndicator string              `json:"confidence_indicator"`
}

// ExportPresentation provides export-ready data
type ExportPresentation struct {
	CSVData        [][]string             `json:"csv_data"`
	Headers        []string               `json:"headers"`
	StructuredData map[string]interface{} `json:"structured_data"`
	ExportMetadata map[string]interface{} `json:"export_metadata"`
}

// DashboardPresentation provides dashboard-ready data
type DashboardPresentation struct {
	Widgets        []DashboardWidget  `json:"widgets"`
	Charts         []ChartData        `json:"charts"`
	KPIs           map[string]float64 `json:"kpis"`
	AlertsWarnings []string           `json:"alerts_warnings"`
}

// DashboardWidget represents a dashboard widget
type DashboardWidget struct {
	Type   string                 `json:"type"`
	Title  string                 `json:"title"`
	Data   interface{}            `json:"data"`
	Config map[string]interface{} `json:"config"`
	Size   string                 `json:"size"`
}

// ChartData represents chart data for visualizations
type ChartData struct {
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	Labels   []string               `json:"labels"`
	Datasets []ChartDataset         `json:"datasets"`
	Options  map[string]interface{} `json:"options"`
}

// ChartDataset represents a dataset for charts
type ChartDataset struct {
	Label string    `json:"label"`
	Data  []float64 `json:"data"`
	Color string    `json:"color"`
}

// APIPresentation provides API-optimized presentation
type APIPresentation struct {
	Status     string                 `json:"status"`
	Data       interface{}            `json:"data"`
	Metadata   map[string]interface{} `json:"metadata"`
	Links      map[string]string      `json:"links"`
	Pagination *PaginationInfo        `json:"pagination,omitempty"`
}

// PaginationInfo provides pagination information
type PaginationInfo struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// AggregateAndPresent aggregates classification results and prepares them for presentation
func (ra *ResultAggregator) AggregateAndPresent(ctx context.Context, request *AggregationRequest) (*AggregatedResults, error) {
	startTime := time.Now()

	ra.logger.Info("Starting result aggregation and presentation",
		zap.Int("input_results", len(request.Results)),
		zap.String("presentation", string(request.Presentation)),
		zap.String("sort_by", string(request.SortBy)))

	if len(request.Results) == 0 {
		return ra.createEmptyResults(request), nil
	}

	// Set defaults
	if request.MaxResultsPerType == 0 {
		request.MaxResultsPerType = 3
	}

	var processingSteps []ProcessingStep

	// Step 1: Deduplicate and merge results
	stepStart := time.Now()
	dedupedResults := ra.deduplicateResults(request.Results)
	processingSteps = append(processingSteps, ProcessingStep{
		Step:        "deduplication",
		Description: "Remove duplicate codes and merge results",
		Duration:    time.Since(stepStart),
		InputCount:  len(request.Results),
		OutputCount: len(dedupedResults),
		Success:     true,
	})

	// Step 2: Calculate enhanced scores
	stepStart = time.Now()
	enhancedResults, err := ra.calculateEnhancedScores(ctx, dedupedResults)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate enhanced scores: %w", err)
	}
	processingSteps = append(processingSteps, ProcessingStep{
		Step:        "score_calculation",
		Description: "Calculate aggregation and confidence scores",
		Duration:    time.Since(stepStart),
		InputCount:  len(dedupedResults),
		OutputCount: len(enhancedResults),
		Success:     true,
	})

	// Step 3: Apply filtering
	stepStart = time.Now()
	filteredResults := ra.applyFiltering(enhancedResults, request)
	processingSteps = append(processingSteps, ProcessingStep{
		Step:        "filtering",
		Description: "Apply confidence and quality filtering",
		Duration:    time.Since(stepStart),
		InputCount:  len(enhancedResults),
		OutputCount: len(filteredResults),
		Success:     true,
	})

	// Step 4: Sort results
	stepStart = time.Now()
	sortedResults := ra.sortResults(filteredResults, request.SortBy)
	processingSteps = append(processingSteps, ProcessingStep{
		Step:        "sorting",
		Description: fmt.Sprintf("Sort results by %s", request.SortBy),
		Duration:    time.Since(stepStart),
		InputCount:  len(filteredResults),
		OutputCount: len(sortedResults),
		Success:     true,
	})

	// Step 5: Group by type and select top results
	stepStart = time.Now()
	topThreeByType := ra.groupAndSelectTopByType(sortedResults, request.MaxResultsPerType)
	processingSteps = append(processingSteps, ProcessingStep{
		Step:        "type_grouping",
		Description: "Group by code type and select top results",
		Duration:    time.Since(stepStart),
		InputCount:  len(sortedResults),
		OutputCount: ra.countResultsInMap(topThreeByType),
		Success:     true,
	})

	// Step 6: Create aggregation metadata
	metadata := &AggregationMetadata{
		AggregationTime:     time.Since(startTime),
		TotalInputResults:   len(request.Results),
		AggregatedCount:     len(sortedResults),
		FilteredCount:       len(request.Results) - len(sortedResults),
		Strategy:            "enhanced_aggregation",
		Criteria:            request,
		QualityDistribution: ra.calculateQualityDistribution(sortedResults),
		TypeDistribution:    ra.calculateTypeDistribution(sortedResults),
		ProcessingSteps:     processingSteps,
	}

	// Step 7: Create analytics (if requested)
	var analytics *AggregationAnalytics
	if request.IncludeAnalytics {
		analytics = ra.calculateAnalytics(sortedResults, topThreeByType)
	}

	// Step 8: Create presentation data
	presentationData := ra.createPresentationData(sortedResults, topThreeByType, request.Presentation, analytics)

	// Step 9: Organize results
	overallTop := ra.getOverallTopResults(sortedResults, request.MaxResultsPerType*3)
	var resultsByStrategy map[string][]*AggregatedResult
	if request.GroupByStrategy {
		resultsByStrategy = ra.groupByStrategy(sortedResults)
	}

	results := &AggregatedResults{
		TopThreeByType:      topThreeByType,
		OverallTopResults:   overallTop,
		AllResults:          sortedResults,
		ResultsByStrategy:   resultsByStrategy,
		AggregationMetadata: metadata,
		Analytics:           analytics,
		PresentationData:    presentationData,
	}

	ra.logger.Info("Completed result aggregation and presentation",
		zap.Duration("total_time", time.Since(startTime)),
		zap.Int("aggregated_results", len(sortedResults)),
		zap.Int("type_groups", len(topThreeByType)))

	return results, nil
}

// Helper methods for the aggregator will be implemented in the next part...
