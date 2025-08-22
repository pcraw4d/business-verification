package compliance

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// GapSeverity indicates business impact of a gap
type GapSeverity string

const (
	GapSeverityLow      GapSeverity = "low"
	GapSeverityMedium   GapSeverity = "medium"
	GapSeverityHigh     GapSeverity = "high"
	GapSeverityCritical GapSeverity = "critical"
)

// GapType categorizes the kind of compliance gap
type GapType string

const (
	GapMissingRequirement  GapType = "missing_requirement"
	GapRequirementNotReady GapType = "requirement_not_implemented"
	GapNonCompliant        GapType = "non_compliant_requirement"
	GapOverdueReview       GapType = "overdue_review"
	GapMissingControl      GapType = "missing_control"
	GapIneffectiveControl  GapType = "ineffective_control"
	GapOverdueControlTest  GapType = "overdue_control_test"
	GapMissingEvidence     GapType = "missing_evidence"
	GapOpenException       GapType = "open_exception"
	GapCoverage            GapType = "coverage_gap"
)

// RequirementGap describes a gap at the requirement level
type RequirementGap struct {
	RequirementID  string      `json:"requirement_id"`
	Title          string      `json:"title"`
	GapType        GapType     `json:"gap_type"`
	Severity       GapSeverity `json:"severity"`
	Description    string      `json:"description"`
	Recommendation string      `json:"recommendation"`
}

// ControlGap describes a gap at the control level
type ControlGap struct {
	ControlID      string      `json:"control_id"`
	RequirementID  string      `json:"requirement_id"`
	Title          string      `json:"title"`
	GapType        GapType     `json:"gap_type"`
	Severity       GapSeverity `json:"severity"`
	Description    string      `json:"description"`
	Recommendation string      `json:"recommendation"`
}

// EvidenceGap describes missing evidence needs
type EvidenceGap struct {
	RequirementID  string      `json:"requirement_id"`
	ControlID      string      `json:"control_id,omitempty"`
	MissingCount   int         `json:"missing_count"`
	Severity       GapSeverity `json:"severity"`
	Recommendation string      `json:"recommendation"`
}

// GapAnalysisReport aggregates all gaps for a business and framework
type GapAnalysisReport struct {
	BusinessID      string           `json:"business_id"`
	FrameworkID     string           `json:"framework_id"`
	GeneratedAt     time.Time        `json:"generated_at"`
	RequirementGaps []RequirementGap `json:"requirement_gaps"`
	ControlGaps     []ControlGap     `json:"control_gaps"`
	EvidenceGaps    []EvidenceGap    `json:"evidence_gaps"`
	Totals          map[string]int   `json:"totals"`
	SeverityCounts  map[string]int   `json:"severity_counts"`
}

// GapAnalyzer finds compliance gaps using tracking and framework definitions
type GapAnalyzer struct {
	logger   *observability.Logger
	tracking *TrackingSystem
	mappings *FrameworkMappingSystem
}

func NewGapAnalyzer(logger *observability.Logger, tracking *TrackingSystem, mappings *FrameworkMappingSystem) *GapAnalyzer {
	return &GapAnalyzer{logger: logger, tracking: tracking, mappings: mappings}
}

// AnalyzeGaps returns a gap analysis report for a business and framework
func (g *GapAnalyzer) AnalyzeGaps(ctx context.Context, businessID, frameworkID string) (*GapAnalysisReport, error) {
	requestID := ctx.Value("request_id").(string)

	g.logger.Info("Analyzing compliance gaps",
		"request_id", requestID,
		"business_id", businessID,
		"framework_id", frameworkID,
	)

	if g.tracking == nil || g.mappings == nil {
		return nil, fmt.Errorf("gap analyzer dependencies not initialized")
	}

	tracking, err := g.tracking.GetComplianceTracking(ctx, businessID, frameworkID)
	if err != nil {
		return nil, fmt.Errorf("tracking not found: %w", err)
	}
	framework, err := g.mappings.GetFramework(ctx, frameworkID)
	if err != nil {
		return nil, fmt.Errorf("framework not found: %w", err)
	}

	report := &GapAnalysisReport{
		BusinessID:     businessID,
		FrameworkID:    frameworkID,
		GeneratedAt:    time.Now(),
		Totals:         make(map[string]int),
		SeverityCounts: make(map[string]int),
	}

	// Build quick index for tracking by requirement and control
	reqTrackMap := make(map[string]*RequirementTracking)
	for i := range tracking.Requirements {
		req := &tracking.Requirements[i]
		reqTrackMap[req.RequirementID] = req
	}
	ctrlTrackMap := make(map[string]*ControlTracking)
	for i := range tracking.Requirements {
		for j := range tracking.Requirements[i].Controls {
			ctrl := &tracking.Requirements[i].Controls[j]
			ctrlTrackMap[ctrl.ControlID] = ctrl
		}
	}

	// Scan each framework requirement for gaps
	for i := range framework.Requirements {
		reqDef := &framework.Requirements[i]
		reqTrack, exists := reqTrackMap[reqDef.RequirementID]
		if !exists {
			// No tracking exists: coverage gap
			g.addReqGap(report, RequirementGap{
				RequirementID:  reqDef.RequirementID,
				Title:          reqDef.Title,
				GapType:        GapMissingRequirement,
				Severity:       g.severityFromDef(reqDef),
				Description:    "Requirement not present in tracking; no assessment available",
				Recommendation: "Initialize tracking for this requirement and begin implementation",
			})
			continue
		}

		// Status-based gaps
		switch reqTrack.Status {
		case ComplianceStatusNotStarted:
			g.addReqGap(report, RequirementGap{
				RequirementID:  reqDef.RequirementID,
				Title:          reqDef.Title,
				GapType:        GapRequirementNotReady,
				Severity:       g.severityFromDef(reqDef),
				Description:    "Requirement not started",
				Recommendation: "Define implementation plan, assign owner, and set target dates",
			})
		case ComplianceStatusNonCompliant:
			g.addReqGap(report, RequirementGap{
				RequirementID:  reqDef.RequirementID,
				Title:          reqDef.Title,
				GapType:        GapNonCompliant,
				Severity:       GapSeverityCritical,
				Description:    "Requirement marked non-compliant",
				Recommendation: "Prioritize remediation with clear corrective controls and evidence",
			})
		}

		// Review recency gap (overdue review)
		if time.Since(reqTrack.LastReviewed) > 90*24*time.Hour { // 90 days
			g.addReqGap(report, RequirementGap{
				RequirementID:  reqDef.RequirementID,
				Title:          reqDef.Title,
				GapType:        GapOverdueReview,
				Severity:       GapSeverityMedium,
				Description:    "Requirement review overdue (>90 days)",
				Recommendation: "Schedule and complete a requirement review; update evidence",
			})
		}

		// Evidence requirement gaps
		if reqDef.EvidenceRequired && len(reqTrack.Evidence) == 0 {
			report.EvidenceGaps = append(report.EvidenceGaps, EvidenceGap{
				RequirementID:  reqDef.RequirementID,
				MissingCount:   1,
				Severity:       g.severityFromDef(reqDef),
				Recommendation: "Collect and attach required evidence per framework guidance",
			})
			g.bump(report, string(GapMissingEvidence))
			g.bumpSeverity(report, string(g.severityFromDef(reqDef)))
		}

		// Open exceptions
		for _, ex := range reqTrack.Exceptions {
			if ex.Status == ExceptionStatusApproved || ex.Status == ExceptionStatusPending {
				g.addReqGap(report, RequirementGap{
					RequirementID:  reqDef.RequirementID,
					Title:          reqDef.Title,
					GapType:        GapOpenException,
					Severity:       GapSeverityHigh,
					Description:    "Active exception exists for this requirement",
					Recommendation: "Review exception validity, define compensating controls, and set expiry",
				})
			}
		}

		// Control-level gaps compared to framework definition
		ctrlDefByID := make(map[string]*ComplianceControl)
		for c := range reqDef.Controls {
			cd := &reqDef.Controls[c]
			ctrlDefByID[cd.ID] = cd
		}

		// Missing controls from definition
		for ctrlID, cd := range ctrlDefByID {
			var ctrlTrack *ControlTracking
			for j := range reqTrack.Controls {
				if reqTrack.Controls[j].ControlID == ctrlID {
					ctrlTrack = &reqTrack.Controls[j]
					break
				}
			}
			if ctrlTrack == nil {
				report.ControlGaps = append(report.ControlGaps, ControlGap{
					ControlID:      ctrlID,
					RequirementID:  reqDef.RequirementID,
					Title:          cd.Title,
					GapType:        GapMissingControl,
					Severity:       g.severityFromDef(reqDef),
					Description:    "Required control missing in tracking",
					Recommendation: "Create and implement the control; define tests and evidence",
				})
				g.bump(report, string(GapMissingControl))
				g.bumpSeverity(report, string(g.severityFromDef(reqDef)))
				continue
			}

			// Ineffective control
			if ctrlTrack.Effectiveness == ControlEffectivenessIneffective || ctrlTrack.Effectiveness == ControlEffectivenessPartiallyEffective {
				report.ControlGaps = append(report.ControlGaps, ControlGap{
					ControlID:      ctrlTrack.ControlID,
					RequirementID:  reqDef.RequirementID,
					Title:          cd.Title,
					GapType:        GapIneffectiveControl,
					Severity:       GapSeverityHigh,
					Description:    "Control effectiveness is low",
					Recommendation: "Improve design or operation; increase testing and monitoring",
				})
				g.bump(report, string(GapIneffectiveControl))
				g.bumpSeverity(report, string(GapSeverityHigh))
			}

			// Overdue testing
			if ctrlTrack.LastTested == nil || time.Since(*ctrlTrack.LastTested) > 60*24*time.Hour {
				report.ControlGaps = append(report.ControlGaps, ControlGap{
					ControlID:      ctrlTrack.ControlID,
					RequirementID:  reqDef.RequirementID,
					Title:          cd.Title,
					GapType:        GapOverdueControlTest,
					Severity:       GapSeverityMedium,
					Description:    "Control testing overdue (>60 days)",
					Recommendation: "Schedule and perform control tests; attach results",
				})
				g.bump(report, string(GapOverdueControlTest))
				g.bumpSeverity(report, string(GapSeverityMedium))
			}

			// Control evidence missing
			if len(ctrlTrack.Evidence) == 0 && cd != nil {
				report.EvidenceGaps = append(report.EvidenceGaps, EvidenceGap{
					RequirementID:  reqDef.RequirementID,
					ControlID:      ctrlTrack.ControlID,
					MissingCount:   1,
					Severity:       GapSeverityMedium,
					Recommendation: "Attach operating evidence for the control (e.g., logs, screenshots)",
				})
				g.bump(report, string(GapMissingEvidence))
				g.bumpSeverity(report, string(GapSeverityMedium))
			}
		}
	}

	// Sort outputs for stable presentation
	sort.SliceStable(report.RequirementGaps, func(i, j int) bool {
		return report.RequirementGaps[i].RequirementID < report.RequirementGaps[j].RequirementID
	})
	sort.SliceStable(report.ControlGaps, func(i, j int) bool {
		return report.ControlGaps[i].RequirementID+report.ControlGaps[i].ControlID < report.ControlGaps[j].RequirementID+report.ControlGaps[j].ControlID
	})

	g.logger.Info("Gap analysis completed",
		"request_id", requestID,
		"business_id", businessID,
		"framework_id", frameworkID,
		"req_gaps", len(report.RequirementGaps),
		"ctrl_gaps", len(report.ControlGaps),
		"evidence_gaps", len(report.EvidenceGaps),
	)

	return report, nil
}

func (g *GapAnalyzer) severityFromDef(req *ComplianceRequirement) GapSeverity {
	switch req.RiskLevel {
	case ComplianceRiskLevelCritical:
		return GapSeverityCritical
	case ComplianceRiskLevelHigh:
		return GapSeverityHigh
	case ComplianceRiskLevelMedium:
		return GapSeverityMedium
	default:
		return GapSeverityLow
	}
}

func (g *GapAnalyzer) addReqGap(report *GapAnalysisReport, gap RequirementGap) {
	report.RequirementGaps = append(report.RequirementGaps, gap)
	g.bump(report, string(gap.GapType))
	g.bumpSeverity(report, string(gap.Severity))
}

func (g *GapAnalyzer) bump(report *GapAnalysisReport, key string) {
	report.Totals[key] = report.Totals[key] + 1
}

func (g *GapAnalyzer) bumpSeverity(report *GapAnalysisReport, sev string) {
	report.SeverityCounts[sev] = report.SeverityCounts[sev] + 1
}
