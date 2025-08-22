package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// CheckRequest specifies what to check for a business
type CheckRequest struct {
	BusinessID string            `json:"business_id"`
	Frameworks []string          `json:"frameworks"` // if empty, check all known in tracking
	Options    EvaluationOptions `json:"options"`
}

// FrameworkCheckResult bundles a framework-level rule evaluation
type FrameworkCheckResult struct {
	FrameworkID string                `json:"framework_id"`
	Summary     ComplianceCheckResult `json:"summary"`
}

// CheckResponse is the aggregate response for a compliance check
type CheckResponse struct {
	BusinessID string                 `json:"business_id"`
	CheckedAt  time.Time              `json:"checked_at"`
	Results    []FrameworkCheckResult `json:"results"`
	Passed     int                    `json:"passed"`
	Failed     int                    `json:"failed"`
}

// CheckEngine orchestrates compliance requirement checking using the rule engine
// It depends on TrackingSystem for current state and FrameworkMappingSystem for framework definitions.
type CheckEngine struct {
	logger     *observability.Logger
	ruleEngine *RuleEngine
	tracking   *TrackingSystem
	mappings   *FrameworkMappingSystem
}

func NewCheckEngine(logger *observability.Logger, ruleEngine *RuleEngine, tracking *TrackingSystem, mappings *FrameworkMappingSystem) *CheckEngine {
	return &CheckEngine{logger: logger, ruleEngine: ruleEngine, tracking: tracking, mappings: mappings}
}

// Check runs compliance checks over the requested frameworks for a business
func (e *CheckEngine) Check(ctx context.Context, req CheckRequest) (*CheckResponse, error) {
	requestID := ctx.Value("request_id").(string)

	e.logger.Info("Running compliance check",
		"request_id", requestID,
		"business_id", req.BusinessID,
		"frameworks", req.Frameworks,
	)

	if e.ruleEngine == nil || e.tracking == nil || e.mappings == nil {
		return nil, fmt.Errorf("check engine dependencies not initialized")
	}

	// Determine frameworks to check based on tracking snapshot
	frameworksToCheck := req.Frameworks
	if len(frameworksToCheck) == 0 {
		// collect from tracking map
		summary, err := e.tracking.GetBusinessComplianceSummary(ctx, req.BusinessID)
		if err != nil {
			return nil, fmt.Errorf("failed to get business tracking: %w", err)
		}
		for fw := range summary {
			frameworksToCheck = append(frameworksToCheck, fw)
		}
	}

	resp := &CheckResponse{BusinessID: req.BusinessID, CheckedAt: time.Now()}

	for _, fw := range frameworksToCheck {
		tracking, err := e.tracking.GetComplianceTracking(ctx, req.BusinessID, fw)
		if err != nil {
			return nil, fmt.Errorf("tracking missing for framework %s: %w", fw, err)
		}
		frameworkDef, err := e.mappings.GetFramework(ctx, fw)
		if err != nil {
			return nil, fmt.Errorf("framework definition missing for %s: %w", fw, err)
		}
		summary, err := e.ruleEngine.EvaluateFramework(ctx, tracking, frameworkDef, req.Options)
		if err != nil {
			return nil, fmt.Errorf("failed evaluating framework %s: %w", fw, err)
		}
		resp.Results = append(resp.Results, FrameworkCheckResult{FrameworkID: fw, Summary: *summary})
		resp.Passed += summary.Passed
		resp.Failed += summary.Failed
	}

	e.logger.Info("Compliance check completed",
		"request_id", requestID,
		"business_id", req.BusinessID,
		"framework_count", len(resp.Results),
		"passed", resp.Passed,
		"failed", resp.Failed,
	)

	return resp, nil
}
