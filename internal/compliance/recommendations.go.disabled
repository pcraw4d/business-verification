package compliance

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// RecommendationEngine generates actionable recommendations based on gaps and scores
type RecommendationEngine struct {
	logger      *observability.Logger
	scoring     *ScoringEngine
	gapAnalyzer *GapAnalyzer
}

func NewRecommendationEngine(logger *observability.Logger, scoring *ScoringEngine, gapAnalyzer *GapAnalyzer) *RecommendationEngine {
	return &RecommendationEngine{logger: logger, scoring: scoring, gapAnalyzer: gapAnalyzer}
}

// GenerateRecommendations produces recommendations for a business across frameworks
// If frameworks is empty, caller should enumerate known ones.
func (e *RecommendationEngine) GenerateRecommendations(ctx context.Context, businessID string, frameworks []string) ([]ComplianceRecommendation, error) {
	requestID := ctx.Value("request_id").(string)

	e.logger.Info("Generating compliance recommendations",
		"request_id", requestID,
		"business_id", businessID,
		"framework_count", len(frameworks),
	)

	if e.gapAnalyzer == nil {
		return nil, fmt.Errorf("gap analyzer not initialized")
	}

	recs := make([]ComplianceRecommendation, 0)
	now := time.Now()

	for _, fw := range frameworks {
		report, err := e.gapAnalyzer.AnalyzeGaps(ctx, businessID, fw)
		if err != nil {
			return nil, fmt.Errorf("gap analysis failed for %s: %w", fw, err)
		}

		// Requirement gaps -> recommendations
		for i := range report.RequirementGaps {
			gap := report.RequirementGaps[i]
			recs = append(recs, e.recForRequirementGap(fw, &gap, now))
		}

		// Control gaps -> recommendations
		for i := range report.ControlGaps {
			gap := report.ControlGaps[i]
			recs = append(recs, e.recForControlGap(fw, &gap, now))
		}

		// Evidence gaps -> recommendations
		for i := range report.EvidenceGaps {
			gap := report.EvidenceGaps[i]
			recs = append(recs, e.recForEvidenceGap(fw, &gap, now))
		}
	}

	// de-duplicate by (title+requirement/control) to avoid noise
	recs = dedupeRecommendations(recs)

	// stable order: priority (critical->low), then title
	sort.SliceStable(recs, func(i, j int) bool {
		pi := priorityRank(recs[i].Priority)
		pj := priorityRank(recs[j].Priority)
		if pi == pj {
			return recs[i].Title < recs[j].Title
		}
		return pi > pj
	})

	e.logger.Info("Compliance recommendations generated",
		"request_id", requestID,
		"business_id", businessID,
		"count", len(recs),
	)

	return recs, nil
}

func (e *RecommendationEngine) recForRequirementGap(framework string, gap *RequirementGap, now time.Time) ComplianceRecommendation {
	id := fmt.Sprintf("rec_req_%s_%s_%d", framework, gap.RequirementID, now.UnixNano())
	priority := priorityFromGapSeverity(gap.Severity)
	return ComplianceRecommendation{
		ID:            id,
		Type:          RecommendationTypeImplementation,
		Priority:      priority,
		Title:         fmt.Sprintf("%s: %s", framework, gap.Title),
		Description:   fmt.Sprintf("Gap: %s - %s", gap.GapType, gap.Description),
		Action:        gap.Recommendation,
		Timeline:      timelineFromPriority(priority),
		Impact:        impactFromPriority(priority),
		Effort:        effortFromGapType(gap.GapType),
		RequirementID: strPtr(gap.RequirementID),
		AssignedTo:    "compliance_officer",
		Status:        RecommendationStatusOpen,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (e *RecommendationEngine) recForControlGap(framework string, gap *ControlGap, now time.Time) ComplianceRecommendation {
	id := fmt.Sprintf("rec_ctrl_%s_%s_%d", framework, gap.ControlID, now.UnixNano())
	priority := priorityFromGapSeverity(gap.Severity)
	return ComplianceRecommendation{
		ID:            id,
		Type:          RecommendationTypeImplementation,
		Priority:      priority,
		Title:         fmt.Sprintf("%s Control: %s", framework, gap.Title),
		Description:   fmt.Sprintf("Gap: %s - %s", gap.GapType, gap.Description),
		Action:        gap.Recommendation,
		Timeline:      timelineFromPriority(priority),
		Impact:        impactFromPriority(priority),
		Effort:        effortFromGapType(gap.GapType),
		RequirementID: strPtr(gap.RequirementID),
		ControlID:     strPtr(gap.ControlID),
		AssignedTo:    "control_owner",
		Status:        RecommendationStatusOpen,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (e *RecommendationEngine) recForEvidenceGap(framework string, gap *EvidenceGap, now time.Time) ComplianceRecommendation {
	id := fmt.Sprintf("rec_evd_%s_%s_%s_%d", framework, gap.RequirementID, gap.ControlID, now.UnixNano())
	priority := CompliancePriorityMedium
	return ComplianceRecommendation{
		ID:            id,
		Type:          RecommendationTypeDocumentation,
		Priority:      priority,
		Title:         fmt.Sprintf("%s Evidence: Requirement %s", framework, gap.RequirementID),
		Description:   "Missing required evidence",
		Action:        gap.Recommendation,
		Timeline:      timelineFromPriority(priority),
		Impact:        impactFromPriority(priority),
		Effort:        "Low",
		RequirementID: strPtr(gap.RequirementID),
		ControlID:     strPtr(gap.ControlID),
		AssignedTo:    "compliance_analyst",
		Status:        RecommendationStatusOpen,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// helpers

func strPtr(s string) *string { return &s }

func dedupeRecommendations(in []ComplianceRecommendation) []ComplianceRecommendation {
	if len(in) == 0 {
		return in
	}
	seen := make(map[string]bool)
	out := make([]ComplianceRecommendation, 0, len(in))
	for _, r := range in {
		key := fmt.Sprintf("%s|%s|%s|%s", r.Title, r.Description, valOrEmpty(r.RequirementID), valOrEmpty(r.ControlID))
		if !seen[key] {
			seen[key] = true
			out = append(out, r)
		}
	}
	return out
}

func valOrEmpty(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func priorityFromGapSeverity(sev GapSeverity) CompliancePriority {
	switch sev {
	case GapSeverityCritical:
		return CompliancePriorityCritical
	case GapSeverityHigh:
		return CompliancePriorityHigh
	case GapSeverityMedium:
		return CompliancePriorityMedium
	default:
		return CompliancePriorityLow
	}
}

func priorityRank(p CompliancePriority) int {
	switch p {
	case CompliancePriorityCritical:
		return 4
	case CompliancePriorityHigh:
		return 3
	case CompliancePriorityMedium:
		return 2
	case CompliancePriorityLow:
		return 1
	default:
		return 0
	}
}

func timelineFromPriority(p CompliancePriority) string {
	switch p {
	case CompliancePriorityCritical:
		return "1-2 weeks"
	case CompliancePriorityHigh:
		return "2-4 weeks"
	case CompliancePriorityMedium:
		return "4-8 weeks"
	default:
		return "> 8 weeks"
	}
}

func impactFromPriority(p CompliancePriority) string {
	switch p {
	case CompliancePriorityCritical:
		return "Critical - audit blocking"
	case CompliancePriorityHigh:
		return "High - significant risk reduction"
	case CompliancePriorityMedium:
		return "Medium - improves evidence and posture"
	default:
		return "Low - hygiene improvement"
	}
}

func effortFromGapType(gt GapType) string {
	switch gt {
	case GapMissingRequirement, GapMissingControl:
		return "High"
	case GapNonCompliant, GapIneffectiveControl:
		return "Medium"
	case GapOverdueReview, GapOverdueControlTest, GapMissingEvidence, GapOpenException, GapRequirementNotReady:
		return "Low"
	default:
		return "Medium"
	}
}
