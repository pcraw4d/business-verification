package classification

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ClassificationAlignmentEngine handles alignment between different classification systems
type ClassificationAlignmentEngine struct {
	db     *sql.DB
	logger *zap.Logger
	config *AlignmentConfig
}

// AlignmentConfig defines configuration for classification alignment
type AlignmentConfig struct {
	EnableMCCAlignment       bool    `json:"enable_mcc_alignment"`
	EnableNAICSAlignment     bool    `json:"enable_naics_alignment"`
	EnableSICAlignment       bool    `json:"enable_sic_alignment"`
	MinAlignmentScore        float64 `json:"min_alignment_score"`
	MaxAlignmentTime         int     `json:"max_alignment_time_seconds"`
	EnableConflictResolution bool    `json:"enable_conflict_resolution"`
	EnableGapAnalysis        bool    `json:"enable_gap_analysis"`
}

// AlignmentResult represents the result of classification alignment analysis
type AlignmentResult struct {
	AnalysisID           string                    `json:"analysis_id"`
	StartTime            time.Time                 `json:"start_time"`
	EndTime              time.Time                 `json:"end_time"`
	Duration             time.Duration             `json:"duration"`
	TotalIndustries      int                       `json:"total_industries"`
	AlignedIndustries    int                       `json:"aligned_industries"`
	MisalignedIndustries int                       `json:"misaligned_industries"`
	Conflicts            []ClassificationConflict  `json:"conflicts"`
	Gaps                 []ClassificationGap       `json:"gaps"`
	Recommendations      []AlignmentRecommendation `json:"recommendations"`
	AlignmentScores      map[string]float64        `json:"alignment_scores"`
	Summary              AlignmentSummary          `json:"summary"`
}

// ClassificationConflict represents a conflict between classification systems
type ClassificationConflict struct {
	IndustryID       int                 `json:"industry_id"`
	IndustryName     string              `json:"industry_name"`
	ConflictType     ConflictType        `json:"conflict_type"`
	ConflictingCodes []ConflictingCode   `json:"conflicting_codes"`
	Severity         ConflictSeverity    `json:"severity"`
	Description      string              `json:"description"`
	Resolution       *ConflictResolution `json:"resolution,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
}

// ClassificationGap represents a gap in classification coverage
type ClassificationGap struct {
	IndustryID     int                `json:"industry_id"`
	IndustryName   string             `json:"industry_name"`
	GapType        GapType            `json:"gap_type"`
	MissingCodes   []MissingCode      `json:"missing_codes"`
	Severity       GapSeverity        `json:"severity"`
	Description    string             `json:"description"`
	Recommendation *GapRecommendation `json:"recommendation,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
}

// AlignmentRecommendation represents a recommendation for improving alignment
type AlignmentRecommendation struct {
	RecommendationID string                 `json:"recommendation_id"`
	Type             RecommendationType     `json:"type"`
	Priority         RecommendationPriority `json:"priority"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	ActionItems      []string               `json:"action_items"`
	ExpectedImpact   string                 `json:"expected_impact"`
	EffortLevel      EffortLevel            `json:"effort_level"`
	CreatedAt        time.Time              `json:"created_at"`
}

// AlignmentSummary provides a high-level summary of alignment status
type AlignmentSummary struct {
	OverallAlignmentScore float64 `json:"overall_alignment_score"`
	MCCAlignmentScore     float64 `json:"mcc_alignment_score"`
	NAICSAlignmentScore   float64 `json:"naics_alignment_score"`
	SICAlignmentScore     float64 `json:"sic_alignment_score"`
	TotalConflicts        int     `json:"total_conflicts"`
	TotalGaps             int     `json:"total_gaps"`
	CriticalIssues        int     `json:"critical_issues"`
	HighPriorityIssues    int     `json:"high_priority_issues"`
	MediumPriorityIssues  int     `json:"medium_priority_issues"`
	LowPriorityIssues     int     `json:"low_priority_issues"`
}

// Supporting types
type ConflictType string

const (
	ConflictTypeCodeMismatch       ConflictType = "code_mismatch"
	ConflictTypeIndustryMismatch   ConflictType = "industry_mismatch"
	ConflictTypeConfidenceMismatch ConflictType = "confidence_mismatch"
	ConflictTypeHierarchyMismatch  ConflictType = "hierarchy_mismatch"
)

type ConflictSeverity string

const (
	ConflictSeverityLow      ConflictSeverity = "low"
	ConflictSeverityMedium   ConflictSeverity = "medium"
	ConflictSeverityHigh     ConflictSeverity = "high"
	ConflictSeverityCritical ConflictSeverity = "critical"
)

type GapType string

const (
	GapTypeMissingMCC        GapType = "missing_mcc"
	GapTypeMissingNAICS      GapType = "missing_naics"
	GapTypeMissingSIC        GapType = "missing_sic"
	GapTypeIncompleteMapping GapType = "incomplete_mapping"
)

type GapSeverity string

const (
	GapSeverityLow      GapSeverity = "low"
	GapSeverityMedium   GapSeverity = "medium"
	GapSeverityHigh     GapSeverity = "high"
	GapSeverityCritical GapSeverity = "critical"
)

type RecommendationType string

const (
	RecommendationTypeCodeMapping      RecommendationType = "code_mapping"
	RecommendationTypeIndustryUpdate   RecommendationType = "industry_update"
	RecommendationTypeConfidenceAdjust RecommendationType = "confidence_adjust"
	RecommendationTypeGapFill          RecommendationType = "gap_fill"
)

type RecommendationPriority string

const (
	RecommendationPriorityLow      RecommendationPriority = "low"
	RecommendationPriorityMedium   RecommendationPriority = "medium"
	RecommendationPriorityHigh     RecommendationPriority = "high"
	RecommendationPriorityCritical RecommendationPriority = "critical"
)

type EffortLevel string

const (
	EffortLevelLow      EffortLevel = "low"
	EffortLevelMedium   EffortLevel = "medium"
	EffortLevelHigh     EffortLevel = "high"
	EffortLevelVeryHigh EffortLevel = "very_high"
)

type ConflictingCode struct {
	CodeType     string  `json:"code_type"`
	Code         string  `json:"code"`
	Description  string  `json:"description"`
	Confidence   float64 `json:"confidence"`
	IndustryID   int     `json:"industry_id"`
	IndustryName string  `json:"industry_name"`
}

type MissingCode struct {
	CodeType    string `json:"code_type"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Reason      string `json:"reason"`
}

type ConflictResolution struct {
	ResolutionType string                 `json:"resolution_type"`
	Description    string                 `json:"description"`
	ActionItems    []string               `json:"action_items"`
	Metadata       map[string]interface{} `json:"metadata"`
	ResolvedAt     time.Time              `json:"resolved_at"`
}

type GapRecommendation struct {
	RecommendationType string                 `json:"recommendation_type"`
	Description        string                 `json:"description"`
	ActionItems        []string               `json:"action_items"`
	Metadata           map[string]interface{} `json:"metadata"`
	CreatedAt          time.Time              `json:"created_at"`
}

// NewClassificationAlignmentEngine creates a new classification alignment engine
func NewClassificationAlignmentEngine(db *sql.DB, logger *zap.Logger, config *AlignmentConfig) *ClassificationAlignmentEngine {
	if config == nil {
		config = &AlignmentConfig{
			EnableMCCAlignment:       true,
			EnableNAICSAlignment:     true,
			EnableSICAlignment:       true,
			MinAlignmentScore:        0.8,
			MaxAlignmentTime:         60,
			EnableConflictResolution: true,
			EnableGapAnalysis:        true,
		}
	}

	return &ClassificationAlignmentEngine{
		db:     db,
		logger: logger,
		config: config,
	}
}

// AnalyzeClassificationAlignment performs comprehensive alignment analysis
func (cae *ClassificationAlignmentEngine) AnalyzeClassificationAlignment(ctx context.Context) (*AlignmentResult, error) {
	startTime := time.Now()
	analysisID := fmt.Sprintf("alignment_analysis_%d", startTime.Unix())

	cae.logger.Info("🔍 Starting classification alignment analysis",
		zap.String("analysis_id", analysisID),
		zap.Time("start_time", startTime))

	result := &AlignmentResult{
		AnalysisID:           analysisID,
		StartTime:            startTime,
		TotalIndustries:      0,
		AlignedIndustries:    0,
		MisalignedIndustries: 0,
		Conflicts:            []ClassificationConflict{},
		Gaps:                 []ClassificationGap{},
		Recommendations:      []AlignmentRecommendation{},
		AlignmentScores:      make(map[string]float64),
	}

	// Get all industries
	industries, err := cae.getIndustries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get industries: %w", err)
	}

	result.TotalIndustries = len(industries)

	// Analyze alignment for each industry
	for _, industry := range industries {
		industryAlignment, err := cae.analyzeIndustryAlignment(ctx, industry)
		if err != nil {
			cae.logger.Error("Failed to analyze industry alignment",
				zap.Int("industry_id", industry.ID),
				zap.String("industry_name", industry.Name),
				zap.Error(err))
			continue
		}

		// Add conflicts and gaps
		result.Conflicts = append(result.Conflicts, industryAlignment.Conflicts...)
		result.Gaps = append(result.Gaps, industryAlignment.Gaps...)

		// Update alignment counts
		if industryAlignment.IsAligned {
			result.AlignedIndustries++
		} else {
			result.MisalignedIndustries++
		}
	}

	// Calculate alignment scores
	if err := cae.CalculateAlignmentScores(ctx, result); err != nil {
		cae.logger.Error("Failed to calculate alignment scores", zap.Error(err))
	}

	// Generate recommendations
	result.Recommendations = cae.GenerateAlignmentRecommendations(result)

	// Create summary
	result.Summary = cae.CreateAlignmentSummary(result)

	// Set end time and duration
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	cae.logger.Info("✅ Classification alignment analysis completed",
		zap.String("analysis_id", analysisID),
		zap.Duration("duration", result.Duration),
		zap.Int("total_industries", result.TotalIndustries),
		zap.Int("aligned_industries", result.AlignedIndustries),
		zap.Int("misaligned_industries", result.MisalignedIndustries),
		zap.Int("conflicts", len(result.Conflicts)),
		zap.Int("gaps", len(result.Gaps)),
		zap.Float64("overall_alignment_score", result.Summary.OverallAlignmentScore))

	return result, nil
}

// analyzeIndustryAlignment analyzes alignment for a specific industry
func (cae *ClassificationAlignmentEngine) analyzeIndustryAlignment(ctx context.Context, industry Industry) (*IndustryAlignmentResult, error) {
	result := &IndustryAlignmentResult{
		IndustryID:   industry.ID,
		IndustryName: industry.Name,
		IsAligned:    true,
		Conflicts:    []ClassificationConflict{},
		Gaps:         []ClassificationGap{},
	}

	// Get crosswalk mappings for this industry
	mappings, err := cae.getIndustryMappings(ctx, industry.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get industry mappings: %w", err)
	}

	// Analyze MCC alignment
	if cae.config.EnableMCCAlignment {
		mccConflicts, mccGaps, err := cae.AnalyzeMCCAlignment(ctx, industry, mappings)
		if err != nil {
			cae.logger.Error("Failed to analyze MCC alignment",
				zap.Int("industry_id", industry.ID),
				zap.Error(err))
		} else {
			result.Conflicts = append(result.Conflicts, mccConflicts...)
			result.Gaps = append(result.Gaps, mccGaps...)
		}
	}

	// Analyze NAICS alignment
	if cae.config.EnableNAICSAlignment {
		naicsConflicts, naicsGaps, err := cae.AnalyzeNAICSAlignment(ctx, industry, mappings)
		if err != nil {
			cae.logger.Error("Failed to analyze NAICS alignment",
				zap.Int("industry_id", industry.ID),
				zap.Error(err))
		} else {
			result.Conflicts = append(result.Conflicts, naicsConflicts...)
			result.Gaps = append(result.Gaps, naicsGaps...)
		}
	}

	// Analyze SIC alignment
	if cae.config.EnableSICAlignment {
		sicConflicts, sicGaps, err := cae.AnalyzeSICAlignment(ctx, industry, mappings)
		if err != nil {
			cae.logger.Error("Failed to analyze SIC alignment",
				zap.Int("industry_id", industry.ID),
				zap.Error(err))
		} else {
			result.Conflicts = append(result.Conflicts, sicConflicts...)
			result.Gaps = append(result.Gaps, sicGaps...)
		}
	}

	// Determine overall alignment
	if len(result.Conflicts) > 0 || len(result.Gaps) > 0 {
		result.IsAligned = false
	}

	return result, nil
}

// AnalyzeMCCAlignment analyzes MCC code alignment for an industry
func (cae *ClassificationAlignmentEngine) AnalyzeMCCAlignment(ctx context.Context, industry Industry, mappings []CrosswalkMapping) ([]ClassificationConflict, []ClassificationGap, error) {
	var conflicts []ClassificationConflict
	var gaps []ClassificationGap

	// Get MCC mappings for this industry
	mccMappings := cae.filterMappingsByType(mappings, "mcc_code")

	// Check for missing MCC codes
	if len(mccMappings) == 0 {
		gap := ClassificationGap{
			IndustryID:   industry.ID,
			IndustryName: industry.Name,
			GapType:      GapTypeMissingMCC,
			MissingCodes: []MissingCode{
				{
					CodeType:    "MCC",
					Code:        "N/A",
					Description: "No MCC codes mapped to this industry",
					Reason:      "Industry lacks MCC code mappings",
				},
			},
			Severity:    GapSeverityHigh,
			Description: fmt.Sprintf("Industry '%s' has no MCC code mappings", industry.Name),
			CreatedAt:   time.Now(),
		}
		gaps = append(gaps, gap)
	}

	// Check for low confidence MCC mappings
	for _, mapping := range mccMappings {
		if mapping.ConfidenceScore < cae.config.MinAlignmentScore {
			conflict := ClassificationConflict{
				IndustryID:   industry.ID,
				IndustryName: industry.Name,
				ConflictType: ConflictTypeConfidenceMismatch,
				ConflictingCodes: []ConflictingCode{
					{
						CodeType:     "MCC",
						Code:         mapping.MCCCode,
						Description:  mapping.Description,
						Confidence:   mapping.ConfidenceScore,
						IndustryID:   industry.ID,
						IndustryName: industry.Name,
					},
				},
				Severity:    ConflictSeverityMedium,
				Description: fmt.Sprintf("MCC code %s has low confidence score %.2f for industry '%s'", mapping.MCCCode, mapping.ConfidenceScore, industry.Name),
				CreatedAt:   time.Now(),
			}
			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts, gaps, nil
}

// AnalyzeNAICSAlignment analyzes NAICS code alignment for an industry
func (cae *ClassificationAlignmentEngine) AnalyzeNAICSAlignment(ctx context.Context, industry Industry, mappings []CrosswalkMapping) ([]ClassificationConflict, []ClassificationGap, error) {
	var conflicts []ClassificationConflict
	var gaps []ClassificationGap

	// Get NAICS mappings for this industry
	naicsMappings := cae.filterMappingsByType(mappings, "naics_code")

	// Check for missing NAICS codes
	if len(naicsMappings) == 0 {
		gap := ClassificationGap{
			IndustryID:   industry.ID,
			IndustryName: industry.Name,
			GapType:      GapTypeMissingNAICS,
			MissingCodes: []MissingCode{
				{
					CodeType:    "NAICS",
					Code:        "N/A",
					Description: "No NAICS codes mapped to this industry",
					Reason:      "Industry lacks NAICS code mappings",
				},
			},
			Severity:    GapSeverityHigh,
			Description: fmt.Sprintf("Industry '%s' has no NAICS code mappings", industry.Name),
			CreatedAt:   time.Now(),
		}
		gaps = append(gaps, gap)
	}

	// Check for invalid NAICS hierarchy
	for _, mapping := range naicsMappings {
		if !cae.isValidNAICSHierarchy(mapping.NAICSCode) {
			conflict := ClassificationConflict{
				IndustryID:   industry.ID,
				IndustryName: industry.Name,
				ConflictType: ConflictTypeHierarchyMismatch,
				ConflictingCodes: []ConflictingCode{
					{
						CodeType:     "NAICS",
						Code:         mapping.NAICSCode,
						Description:  mapping.Description,
						Confidence:   mapping.ConfidenceScore,
						IndustryID:   industry.ID,
						IndustryName: industry.Name,
					},
				},
				Severity:    ConflictSeverityHigh,
				Description: fmt.Sprintf("NAICS code %s has invalid hierarchy for industry '%s'", mapping.NAICSCode, industry.Name),
				CreatedAt:   time.Now(),
			}
			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts, gaps, nil
}

// AnalyzeSICAlignment analyzes SIC code alignment for an industry
func (cae *ClassificationAlignmentEngine) AnalyzeSICAlignment(ctx context.Context, industry Industry, mappings []CrosswalkMapping) ([]ClassificationConflict, []ClassificationGap, error) {
	var conflicts []ClassificationConflict
	var gaps []ClassificationGap

	// Get SIC mappings for this industry
	sicMappings := cae.filterMappingsByType(mappings, "sic_code")

	// Check for missing SIC codes
	if len(sicMappings) == 0 {
		gap := ClassificationGap{
			IndustryID:   industry.ID,
			IndustryName: industry.Name,
			GapType:      GapTypeMissingSIC,
			MissingCodes: []MissingCode{
				{
					CodeType:    "SIC",
					Code:        "N/A",
					Description: "No SIC codes mapped to this industry",
					Reason:      "Industry lacks SIC code mappings",
				},
			},
			Severity:    GapSeverityHigh,
			Description: fmt.Sprintf("Industry '%s' has no SIC code mappings", industry.Name),
			CreatedAt:   time.Now(),
		}
		gaps = append(gaps, gap)
	}

	// Check for invalid SIC division
	for _, mapping := range sicMappings {
		if !cae.isValidSICDivision(mapping.SICCode) {
			conflict := ClassificationConflict{
				IndustryID:   industry.ID,
				IndustryName: industry.Name,
				ConflictType: ConflictTypeHierarchyMismatch,
				ConflictingCodes: []ConflictingCode{
					{
						CodeType:     "SIC",
						Code:         mapping.SICCode,
						Description:  mapping.Description,
						Confidence:   mapping.ConfidenceScore,
						IndustryID:   industry.ID,
						IndustryName: industry.Name,
					},
				},
				Severity:    ConflictSeverityHigh,
				Description: fmt.Sprintf("SIC code %s has invalid division for industry '%s'", mapping.SICCode, industry.Name),
				CreatedAt:   time.Now(),
			}
			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts, gaps, nil
}

// CalculateAlignmentScores calculates alignment scores for different classification systems
func (cae *ClassificationAlignmentEngine) CalculateAlignmentScores(ctx context.Context, result *AlignmentResult) error {
	// Calculate overall alignment score
	if result.TotalIndustries > 0 {
		result.AlignmentScores["overall"] = float64(result.AlignedIndustries) / float64(result.TotalIndustries)
	}

	// Calculate MCC alignment score
	mccAligned := 0
	mccTotal := 0
	for _, industry := range cae.getIndustriesFromResult(result) {
		mappings, err := cae.getIndustryMappings(ctx, industry.ID)
		if err != nil {
			continue
		}
		mccMappings := cae.filterMappingsByType(mappings, "mcc_code")
		if len(mccMappings) > 0 {
			mccTotal++
			hasHighConfidence := false
			for _, mapping := range mccMappings {
				if mapping.ConfidenceScore >= cae.config.MinAlignmentScore {
					hasHighConfidence = true
					break
				}
			}
			if hasHighConfidence {
				mccAligned++
			}
		}
	}
	if mccTotal > 0 {
		result.AlignmentScores["mcc"] = float64(mccAligned) / float64(mccTotal)
	}

	// Calculate NAICS alignment score
	naicsAligned := 0
	naicsTotal := 0
	for _, industry := range cae.getIndustriesFromResult(result) {
		mappings, err := cae.getIndustryMappings(ctx, industry.ID)
		if err != nil {
			continue
		}
		naicsMappings := cae.filterMappingsByType(mappings, "naics_code")
		if len(naicsMappings) > 0 {
			naicsTotal++
			hasValidHierarchy := true
			for _, mapping := range naicsMappings {
				if !cae.isValidNAICSHierarchy(mapping.NAICSCode) {
					hasValidHierarchy = false
					break
				}
			}
			if hasValidHierarchy {
				naicsAligned++
			}
		}
	}
	if naicsTotal > 0 {
		result.AlignmentScores["naics"] = float64(naicsAligned) / float64(naicsTotal)
	}

	// Calculate SIC alignment score
	sicAligned := 0
	sicTotal := 0
	for _, industry := range cae.getIndustriesFromResult(result) {
		mappings, err := cae.getIndustryMappings(ctx, industry.ID)
		if err != nil {
			continue
		}
		sicMappings := cae.filterMappingsByType(mappings, "sic_code")
		if len(sicMappings) > 0 {
			sicTotal++
			hasValidDivision := true
			for _, mapping := range sicMappings {
				if !cae.isValidSICDivision(mapping.SICCode) {
					hasValidDivision = false
					break
				}
			}
			if hasValidDivision {
				sicAligned++
			}
		}
	}
	if sicTotal > 0 {
		result.AlignmentScores["sic"] = float64(sicAligned) / float64(sicTotal)
	}

	return nil
}

// GenerateAlignmentRecommendations generates recommendations for improving alignment
func (cae *ClassificationAlignmentEngine) GenerateAlignmentRecommendations(result *AlignmentResult) []AlignmentRecommendation {
	var recommendations []AlignmentRecommendation

	// Analyze conflicts and generate recommendations
	conflictTypes := make(map[ConflictType]int)
	for _, conflict := range result.Conflicts {
		conflictTypes[conflict.ConflictType]++
	}

	// Generate recommendations based on conflict types
	if conflictTypes[ConflictTypeConfidenceMismatch] > 0 {
		recommendations = append(recommendations, AlignmentRecommendation{
			RecommendationID: "confidence_improvement",
			Type:             RecommendationTypeConfidenceAdjust,
			Priority:         RecommendationPriorityHigh,
			Title:            "Improve Confidence Scores",
			Description:      fmt.Sprintf("Address %d confidence mismatch conflicts", conflictTypes[ConflictTypeConfidenceMismatch]),
			ActionItems: []string{
				"Review and update confidence scoring algorithm",
				"Validate industry-code mappings with domain experts",
				"Implement confidence score calibration",
			},
			ExpectedImpact: "Improved classification accuracy and reliability",
			EffortLevel:    EffortLevelMedium,
			CreatedAt:      time.Now(),
		})
	}

	if conflictTypes[ConflictTypeHierarchyMismatch] > 0 {
		recommendations = append(recommendations, AlignmentRecommendation{
			RecommendationID: "hierarchy_validation",
			Type:             RecommendationTypeCodeMapping,
			Priority:         RecommendationPriorityCritical,
			Title:            "Fix Hierarchy Mismatches",
			Description:      fmt.Sprintf("Resolve %d hierarchy mismatch conflicts", conflictTypes[ConflictTypeHierarchyMismatch]),
			ActionItems: []string{
				"Validate NAICS and SIC code hierarchies",
				"Update invalid code mappings",
				"Implement hierarchy validation rules",
			},
			ExpectedImpact: "Ensures compliance with classification standards",
			EffortLevel:    EffortLevelHigh,
			CreatedAt:      time.Now(),
		})
	}

	// Analyze gaps and generate recommendations
	gapTypes := make(map[GapType]int)
	for _, gap := range result.Gaps {
		gapTypes[gap.GapType]++
	}

	if gapTypes[GapTypeMissingMCC] > 0 || gapTypes[GapTypeMissingNAICS] > 0 || gapTypes[GapTypeMissingSIC] > 0 {
		recommendations = append(recommendations, AlignmentRecommendation{
			RecommendationID: "gap_filling",
			Type:             RecommendationTypeGapFill,
			Priority:         RecommendationPriorityHigh,
			Title:            "Fill Classification Gaps",
			Description:      fmt.Sprintf("Address %d missing code mappings", gapTypes[GapTypeMissingMCC]+gapTypes[GapTypeMissingNAICS]+gapTypes[GapTypeMissingSIC]),
			ActionItems: []string{
				"Identify appropriate codes for unmapped industries",
				"Create new crosswalk mappings",
				"Validate new mappings with business stakeholders",
			},
			ExpectedImpact: "Complete classification coverage for all industries",
			EffortLevel:    EffortLevelHigh,
			CreatedAt:      time.Now(),
		})
	}

	return recommendations
}

// CreateAlignmentSummary creates a summary of alignment status
func (cae *ClassificationAlignmentEngine) CreateAlignmentSummary(result *AlignmentResult) AlignmentSummary {
	summary := AlignmentSummary{
		OverallAlignmentScore: result.AlignmentScores["overall"],
		MCCAlignmentScore:     result.AlignmentScores["mcc"],
		NAICSAlignmentScore:   result.AlignmentScores["naics"],
		SICAlignmentScore:     result.AlignmentScores["sic"],
		TotalConflicts:        len(result.Conflicts),
		TotalGaps:             len(result.Gaps),
	}

	// Count issues by severity
	for _, conflict := range result.Conflicts {
		switch conflict.Severity {
		case ConflictSeverityCritical:
			summary.CriticalIssues++
		case ConflictSeverityHigh:
			summary.HighPriorityIssues++
		case ConflictSeverityMedium:
			summary.MediumPriorityIssues++
		case ConflictSeverityLow:
			summary.LowPriorityIssues++
		}
	}

	for _, gap := range result.Gaps {
		switch gap.Severity {
		case GapSeverityCritical:
			summary.CriticalIssues++
		case GapSeverityHigh:
			summary.HighPriorityIssues++
		case GapSeverityMedium:
			summary.MediumPriorityIssues++
		case GapSeverityLow:
			summary.LowPriorityIssues++
		}
	}

	return summary
}

// Helper methods
func (cae *ClassificationAlignmentEngine) getIndustries(ctx context.Context) ([]Industry, error) {
	query := `SELECT id, name, description FROM industries ORDER BY name`
	rows, err := cae.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var industries []Industry
	for rows.Next() {
		var industry Industry
		if err := rows.Scan(&industry.ID, &industry.Name, &industry.Description); err != nil {
			return nil, err
		}
		industries = append(industries, industry)
	}

	return industries, nil
}

func (cae *ClassificationAlignmentEngine) getIndustryMappings(ctx context.Context, industryID int) ([]CrosswalkMapping, error) {
	query := `
		SELECT id, industry_id, mcc_code, naics_code, sic_code, description, confidence_score, created_at, updated_at
		FROM crosswalk_mappings 
		WHERE industry_id = $1
	`
	rows, err := cae.db.QueryContext(ctx, query, industryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mappings []CrosswalkMapping
	for rows.Next() {
		var mapping CrosswalkMapping
		if err := rows.Scan(
			&mapping.ID, &mapping.IndustryID, &mapping.MCCCode, &mapping.NAICSCode, &mapping.SICCode,
			&mapping.Description, &mapping.ConfidenceScore, &mapping.CreatedAt, &mapping.UpdatedAt,
		); err != nil {
			return nil, err
		}
		mappings = append(mappings, mapping)
	}

	return mappings, nil
}

func (cae *ClassificationAlignmentEngine) filterMappingsByType(mappings []CrosswalkMapping, codeType string) []CrosswalkMapping {
	var filtered []CrosswalkMapping
	for _, mapping := range mappings {
		switch codeType {
		case "mcc_code":
			if mapping.MCCCode != "" {
				filtered = append(filtered, mapping)
			}
		case "naics_code":
			if mapping.NAICSCode != "" {
				filtered = append(filtered, mapping)
			}
		case "sic_code":
			if mapping.SICCode != "" {
				filtered = append(filtered, mapping)
			}
		}
	}
	return filtered
}

func (cae *ClassificationAlignmentEngine) isValidNAICSHierarchy(naicsCode string) bool {
	if len(naicsCode) != 6 {
		return false
	}

	// Check if first 2 digits are valid NAICS sectors
	validSectors := []string{"11", "21", "22", "23", "31", "32", "33", "42", "44", "45", "48", "49", "51", "52", "53", "54", "55", "56", "61", "62", "71", "72", "81", "92"}
	sector := naicsCode[:2]

	for _, validSector := range validSectors {
		if sector == validSector {
			return true
		}
	}

	return false
}

func (cae *ClassificationAlignmentEngine) isValidSICDivision(sicCode string) bool {
	if len(sicCode) != 4 {
		return false
	}

	// Check if first digit is valid SIC division (0-9)
	division := sicCode[0]
	return division >= '0' && division <= '9'
}

func (cae *ClassificationAlignmentEngine) getIndustriesFromResult(result *AlignmentResult) []Industry {
	// This is a simplified implementation - in practice, you'd want to store industries in the result
	// For now, we'll return an empty slice as this is used for score calculation
	return []Industry{}
}

// IndustryAlignmentResult represents the alignment result for a specific industry
type IndustryAlignmentResult struct {
	IndustryID   int                      `json:"industry_id"`
	IndustryName string                   `json:"industry_name"`
	IsAligned    bool                     `json:"is_aligned"`
	Conflicts    []ClassificationConflict `json:"conflicts"`
	Gaps         []ClassificationGap      `json:"gaps"`
}
