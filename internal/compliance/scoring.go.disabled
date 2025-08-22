package compliance

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ScoreWeights defines configurable weights for scoring
type ScoreWeights struct {
	// Requirement-level weights
	StatusWeight         float64 // weight for requirement status
	ImplementationWeight float64 // weight for implementation status
	ControlsWeight       float64 // weight for control effectiveness aggregation
	EvidenceWeight       float64 // weight for evidence presence
	RecencyWeight        float64 // weight for freshness of review/tests
	ExceptionPenalty     float64 // penalty per active exception

	// Framework aggregation
	HighPriorityWeight     float64 // multiplier for high priority requirements
	CriticalPriorityWeight float64 // multiplier for critical priority requirements

	// Global clamps
	MinScore float64
	MaxScore float64
}

// DefaultScoreWeights returns sane defaults
func DefaultScoreWeights() ScoreWeights {
	return ScoreWeights{
		StatusWeight:           0.25,
		ImplementationWeight:   0.20,
		ControlsWeight:         0.45,
		EvidenceWeight:         0.05,
		RecencyWeight:          0.05,
		ExceptionPenalty:       5.0,
		HighPriorityWeight:     1.25,
		CriticalPriorityWeight: 1.5,
		MinScore:               0.0,
		MaxScore:               100.0,
	}
}

// ScoringEngine calculates compliance scores
type ScoringEngine struct {
	logger  *observability.Logger
	weights ScoreWeights
}

func NewScoringEngine(logger *observability.Logger, weights ScoreWeights) *ScoringEngine {
	return &ScoringEngine{logger: logger, weights: weights}
}

// RequirementScore computes a score (0-100) for a requirement
func (e *ScoringEngine) RequirementScore(req *RequirementTracking) float64 {
	if req == nil {
		return 0
	}

	statusScore := e.scoreByRequirementStatus(req.Status)
	implScore := e.scoreByImplementationStatus(req.ImplementationStatus)
	controlsScore := e.controlsAggregateScore(req.Controls)
	evidenceScore := e.evidenceScore(len(req.Evidence))
	recencyScore := e.recencyScore(req.LastReviewed, 30*24*time.Hour) // monthly target

	weighted := statusScore*e.weights.StatusWeight +
		implScore*e.weights.ImplementationWeight +
		controlsScore*e.weights.ControlsWeight +
		evidenceScore*e.weights.EvidenceWeight +
		recencyScore*e.weights.RecencyWeight

	// Exception penalties
	penalty := float64(len(req.Exceptions)) * e.weights.ExceptionPenalty
	score := clamp(weighted-penalty, e.weights.MinScore, e.weights.MaxScore)
	return score
}

// FrameworkScore computes average weighted score across requirements
func (e *ScoringEngine) FrameworkScore(tracking *ComplianceTracking, requirements map[string]*ComplianceRequirement) float64 {
	if tracking == nil || len(tracking.Requirements) == 0 {
		return 0
	}
	total := 0.0
	weightSum := 0.0
	for i := range tracking.Requirements {
		reqTrack := &tracking.Requirements[i]
		reqScore := e.RequirementScore(reqTrack)
		w := 1.0
		// apply priority multiplier when available
		if requirements != nil {
			if reqDef, ok := requirements[reqTrack.RequirementID]; ok {
				switch reqDef.Priority {
				case CompliancePriorityHigh:
					w = e.weights.HighPriorityWeight
				case CompliancePriorityCritical:
					w = e.weights.CriticalPriorityWeight
				}
			}
		}
		total += reqScore * w
		weightSum += w
	}
	if weightSum == 0 {
		return 0
	}
	return clamp(total/weightSum, e.weights.MinScore, e.weights.MaxScore)
}

// BusinessScore aggregates across frameworks present in summary
func (e *ScoringEngine) BusinessScore(ctx context.Context, trackingSummary map[string]*ComplianceTracking, frameworkDefs *FrameworkMappingSystem) float64 {
	if len(trackingSummary) == 0 || frameworkDefs == nil {
		return 0
	}
	// stabilize ordering
	keys := make([]string, 0, len(trackingSummary))
	for k := range trackingSummary {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	total := 0.0
	for _, fw := range keys {
		tr := trackingSummary[fw]
		// build requirement definition map for priority weighting
		reqDefs := make(map[string]*ComplianceRequirement)
		if fwDef, ok := frameworkDefs.frameworks[fw]; ok && fwDef != nil {
			for i := range fwDef.Requirements {
				req := fwDef.Requirements[i]
				reqCopy := req // capture
				reqDefs[req.RequirementID] = &reqCopy
			}
		}
		fwScore := e.FrameworkScore(tr, reqDefs)
		total += fwScore
	}
	return clamp(total/float64(len(keys)), e.weights.MinScore, e.weights.MaxScore)
}

// scoring primitives

func (e *ScoringEngine) scoreByRequirementStatus(s ComplianceStatus) float64 {
	switch s {
	case ComplianceStatusVerified:
		return 100
	case ComplianceStatusImplemented:
		return 80
	case ComplianceStatusInProgress:
		return 50
	case ComplianceStatusNotStarted:
		return 10
	case ComplianceStatusNonCompliant:
		return 0
	case ComplianceStatusExempt:
		return 70
	default:
		return 0
	}
}

func (e *ScoringEngine) scoreByImplementationStatus(s ImplementationStatus) float64 {
	switch s {
	case ImplementationStatusDeployed:
		return 100
	case ImplementationStatusTested:
		return 85
	case ImplementationStatusImplemented:
		return 75
	case ImplementationStatusInProgress:
		return 50
	case ImplementationStatusPlanned:
		return 25
	case ImplementationStatusNotImplemented:
		return 0
	default:
		return 0
	}
}

func (e *ScoringEngine) controlsAggregateScore(controls []ControlTracking) float64 {
	if len(controls) == 0 {
		return 0
	}
	sum := 0.0
	for i := range controls {
		sum += e.controlScore(&controls[i])
	}
	return sum / float64(len(controls))
}

func (e *ScoringEngine) controlScore(ctrl *ControlTracking) float64 {
	if ctrl == nil {
		return 0
	}
	base := 0.0
	switch ctrl.Effectiveness {
	case ControlEffectivenessHighlyEffective:
		base = 100
	case ControlEffectivenessEffective:
		base = 80
	case ControlEffectivenessPartiallyEffective:
		base = 50
	case ControlEffectivenessIneffective:
		base = 10
	}
	// recency: tests within 30 days preferred
	recency := e.recencyScore(ptrTimeOrZero(ctrl.LastTested), 30*24*time.Hour)
	// evidence presence bonus (capped small)
	evidence := e.evidenceScore(len(ctrl.Evidence))
	// combine with small weights to avoid dominance
	score := base*0.85 + recency*0.10 + evidence*0.05
	return clamp(score, e.weights.MinScore, e.weights.MaxScore)
}

func (e *ScoringEngine) evidenceScore(count int) float64 {
	if count <= 0 {
		return 0
	}
	if count >= 5 {
		return 100
	}
	return (float64(count) / 5.0) * 100.0
}

// recencyScore returns 100 if within targetAge; declines to 0 as age approaches 4x target
func (e *ScoringEngine) recencyScore(last time.Time, targetAge time.Duration) float64 {
	if last.IsZero() {
		return 0
	}
	age := time.Since(last)
	if age <= targetAge {
		return 100
	}
	// linear decay to 0 at 4x target
	maxAge := 4 * targetAge
	if age >= maxAge {
		return 0
	}
	ratio := 1.0 - (float64(age) / float64(maxAge))
	return clamp(ratio*100.0, 0, 100)
}

func clamp(v, min, max float64) float64 {
	return math.Max(min, math.Min(max, v))
}

func ptrTimeOrZero(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
