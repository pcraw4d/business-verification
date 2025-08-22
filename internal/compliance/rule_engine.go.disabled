package compliance

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// EntityType enumerates rule targets
// Allowed: "overall", "framework", "requirement", "control"
type EntityType string

const (
	EntityTypeOverall     EntityType = "overall"
	EntityTypeFramework   EntityType = "framework"
	EntityTypeRequirement EntityType = "requirement"
	EntityTypeControl     EntityType = "control"
)

// Operator represents predicate operators for conditions
type Operator string

const (
	OpEq         Operator = "eq"
	OpNe         Operator = "ne"
	OpGt         Operator = "gt"
	OpGte        Operator = "gte"
	OpLt         Operator = "lt"
	OpLte        Operator = "lte"
	OpContains   Operator = "contains"
	OpIn         Operator = "in"
	OpExists     Operator = "exists"
	OpNotExists  Operator = "not_exists"
	OpRegexMatch Operator = "regex"
)

// Predicate describes a single boolean check
type Predicate struct {
	Attribute string      `json:"attribute"`
	Operator  Operator    `json:"operator"`
	Value     interface{} `json:"value,omitempty"`
}

// Condition represents a boolean expression tree
// Only one of All/Any/Not/Predicate is set
// - All: AND over children
// - Any: OR over children
// - Not: negation of the nested condition
// - Predicate: leaf predicate
// An empty condition evaluates to true
// This flexible structure avoids deep nesting complexity in callers
// and keeps evaluation concise.
type Condition struct {
	All       []Condition `json:"all,omitempty"`
	Any       []Condition `json:"any,omitempty"`
	Not       *Condition  `json:"not,omitempty"`
	Predicate *Predicate  `json:"predicate,omitempty"`
}

// RuleEffect defines optional side effects if a rule fails
// Effects are only applied when ApplyEffects is true in evaluation options.
type RuleEffect struct {
	SetStatus       *ComplianceStatus         `json:"set_status,omitempty"`
	ScoreAdjustment *float64                  `json:"score_adjustment,omitempty"` // positive or negative delta 0..100
	RequireEvidence bool                      `json:"require_evidence,omitempty"`
	Recommendation  *ComplianceRecommendation `json:"recommendation,omitempty"`
}

// Rule defines a compliance rule
// A rule passes if its condition evaluates to true for a given entity.
// If a rule fails, its effect MAY be applied (optional).
type Rule struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	EntityType  EntityType `json:"entity_type"`
	TargetIDs   []string   `json:"target_ids,omitempty"` // optional filter to specific requirement/control IDs
	Condition   Condition  `json:"condition"`
	Severity    string     `json:"severity"` // "low","medium","high","critical"
	Effect      RuleEffect `json:"effect"`
	Enabled     bool       `json:"enabled"`
}

// RuleSet groups rules for a framework
// Example: SOC2 Common Controls v1
// Rule order is preserved for deterministic evaluation
// and results are sorted by severity then name in reports.
type RuleSet struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Framework string    `json:"framework"`
	Version   string    `json:"version"`
	Rules     []Rule    `json:"rules"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RuleOutcome captures the evaluation of a single rule against a single entity
type RuleOutcome struct {
	RuleID     string     `json:"rule_id"`
	RuleName   string     `json:"rule_name"`
	EntityType EntityType `json:"entity_type"`
	EntityID   string     `json:"entity_id"`
	Passed     bool       `json:"passed"`
	Severity   string     `json:"severity"`
	Details    string     `json:"details"`
}

// ComplianceCheckResult aggregates outcomes for a framework evaluation
type ComplianceCheckResult struct {
	BusinessID string        `json:"business_id"`
	Framework  string        `json:"framework"`
	Evaluated  time.Time     `json:"evaluated"`
	Passed     int           `json:"passed"`
	Failed     int           `json:"failed"`
	Outcomes   []RuleOutcome `json:"outcomes"`
}

// EvaluationOptions controls evaluation behavior
type EvaluationOptions struct {
	ApplyEffects bool `json:"apply_effects"`
}

// RuleEngine evaluates rule sets for compliance checks
// Thread-safe registration and evaluation.
type RuleEngine struct {
	logger   *observability.Logger
	mu       sync.RWMutex
	ruleSets map[string][]*RuleSet // framework -> rule sets
}

// NewRuleEngine constructs a RuleEngine
func NewRuleEngine(logger *observability.Logger) *RuleEngine {
	return &RuleEngine{
		logger:   logger,
		ruleSets: make(map[string][]*RuleSet),
	}
}

// RegisterRuleSet registers or replaces a rule set for a framework
func (e *RuleEngine) RegisterRuleSet(ctx context.Context, set *RuleSet) error {
	if set == nil {
		return fmt.Errorf("nil rule set")
	}
	if set.Framework == "" {
		return fmt.Errorf("rule set framework is required")
	}
	if set.ID == "" {
		return fmt.Errorf("rule set ID is required")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// replace by ID if exists; else append
	list := e.ruleSets[set.Framework]
	replaced := false
	for i := range list {
		if list[i].ID == set.ID {
			set.UpdatedAt = time.Now()
			list[i] = set
			replaced = true
			break
		}
	}
	if !replaced {
		set.CreatedAt = time.Now()
		set.UpdatedAt = set.CreatedAt
		list = append(list, set)
	}
	e.ruleSets[set.Framework] = list

	return nil
}

// GetRuleSets returns rule sets for a framework (copy)
func (e *RuleEngine) GetRuleSets(framework string) []*RuleSet {
	e.mu.RLock()
	defer e.mu.RUnlock()
	list := e.ruleSets[framework]
	out := make([]*RuleSet, len(list))
	copy(out, list)
	return out
}

// EvaluateFramework evaluates all registered rules for a framework against tracking data
func (e *RuleEngine) EvaluateFramework(
	ctx context.Context,
	tracking *ComplianceTracking,
	frameworkDef *RegulatoryFramework,
	options EvaluationOptions,
) (*ComplianceCheckResult, error) {
	requestID := ctx.Value("request_id").(string)

	if tracking == nil {
		return nil, fmt.Errorf("tracking is required")
	}

	e.mu.RLock()
	sets := e.ruleSets[tracking.Framework]
	e.mu.RUnlock()

	result := &ComplianceCheckResult{
		BusinessID: tracking.BusinessID,
		Framework:  tracking.Framework,
		Evaluated:  time.Now(),
		Outcomes:   make([]RuleOutcome, 0),
	}

	// Build quick lookup maps for current tracking snapshot
	reqMap := make(map[string]*RequirementTracking, len(tracking.Requirements))
	for i := range tracking.Requirements {
		req := &tracking.Requirements[i]
		reqMap[req.RequirementID] = req
	}

	controlMap := make(map[string]*ControlTracking)
	for i := range tracking.Requirements {
		for j := range tracking.Requirements[i].Controls {
			ctrl := &tracking.Requirements[i].Controls[j]
			controlMap[ctrl.ControlID] = ctrl
		}
	}

	// Iterate rule sets and rules
	for _, set := range sets {
		for _, rule := range set.Rules {
			if !rule.Enabled {
				continue
			}

			switch rule.EntityType {
			case EntityTypeRequirement:
				// Targets: all requirements or filtered
				candidateIDs := filterTargetRequirementIDs(rule.TargetIDs, reqMap)
				for _, rid := range candidateIDs {
					req := reqMap[rid]
					passed := e.evaluateConditionRequirement(&rule.Condition, req)
					details := e.buildDetails(passed)
					out := RuleOutcome{RuleID: rule.ID, RuleName: rule.Name, EntityType: rule.EntityType, EntityID: rid, Passed: passed, Severity: rule.Severity, Details: details}
					result.Outcomes = append(result.Outcomes, out)

					if !passed && options.ApplyEffects {
						e.applyRequirementEffect(rule.Effect, req)
					}
				}

			case EntityTypeControl:
				candidateIDs := filterTargetControlIDs(rule.TargetIDs, controlMap)
				for _, cid := range candidateIDs {
					ctrl := controlMap[cid]
					passed := e.evaluateConditionControl(&rule.Condition, ctrl)
					details := e.buildDetails(passed)
					out := RuleOutcome{RuleID: rule.ID, RuleName: rule.Name, EntityType: rule.EntityType, EntityID: cid, Passed: passed, Severity: rule.Severity, Details: details}
					result.Outcomes = append(result.Outcomes, out)

					if !passed && options.ApplyEffects {
						e.applyControlEffect(rule.Effect, ctrl)
					}
				}

			case EntityTypeFramework, EntityTypeOverall:
				// Simple framework-level evaluation using tracking fields
				passed := e.evaluateConditionFramework(&rule.Condition, tracking)
				details := e.buildDetails(passed)
				entityID := tracking.Framework
				out := RuleOutcome{RuleID: rule.ID, RuleName: rule.Name, EntityType: rule.EntityType, EntityID: entityID, Passed: passed, Severity: rule.Severity, Details: details}
				result.Outcomes = append(result.Outcomes, out)
			}
		}
	}

	// Tally
	for _, o := range result.Outcomes {
		if o.Passed {
			result.Passed++
		} else {
			result.Failed++
		}
	}

	// Sort outcomes by severity, then name for readability
	sort.SliceStable(result.Outcomes, func(i, j int) bool {
		if result.Outcomes[i].Severity == result.Outcomes[j].Severity {
			return result.Outcomes[i].RuleName < result.Outcomes[j].RuleName
		}
		return severityRank(result.Outcomes[i].Severity) > severityRank(result.Outcomes[j].Severity)
	})

	e.logger.Info("Compliance rules evaluated",
		"request_id", requestID,
		"business_id", tracking.BusinessID,
		"framework", tracking.Framework,
		"rule_sets", len(sets),
		"passed", result.Passed,
		"failed", result.Failed,
	)

	return result, nil
}

// Helpers

func (e *RuleEngine) buildDetails(passed bool) string {
	if passed {
		return "condition satisfied"
	}
	return "condition failed"
}

func severityRank(s string) int {
	s = strings.ToLower(s)
	switch s {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// Condition evaluation

func (e *RuleEngine) evaluateConditionRequirement(cond *Condition, req *RequirementTracking) bool {
	if cond == nil {
		return true
	}
	// All
	if len(cond.All) > 0 {
		for i := range cond.All {
			if !e.evaluateConditionRequirement(&cond.All[i], req) {
				return false
			}
		}
		return true
	}
	// Any
	if len(cond.Any) > 0 {
		for i := range cond.Any {
			if e.evaluateConditionRequirement(&cond.Any[i], req) {
				return true
			}
		}
		return false
	}
	// Not
	if cond.Not != nil {
		return !e.evaluateConditionRequirement(cond.Not, req)
	}
	// Predicate
	if cond.Predicate != nil {
		attr := strings.ToLower(cond.Predicate.Attribute)
		return evalPredicate(cond.Predicate.Operator, e.reqAttrValue(attr, req), cond.Predicate.Value)
	}
	return true
}

func (e *RuleEngine) evaluateConditionControl(cond *Condition, ctrl *ControlTracking) bool {
	if cond == nil {
		return true
	}
	if len(cond.All) > 0 {
		for i := range cond.All {
			if !e.evaluateConditionControl(&cond.All[i], ctrl) {
				return false
			}
		}
		return true
	}
	if len(cond.Any) > 0 {
		for i := range cond.Any {
			if e.evaluateConditionControl(&cond.Any[i], ctrl) {
				return true
			}
		}
		return false
	}
	if cond.Not != nil {
		return !e.evaluateConditionControl(cond.Not, ctrl)
	}
	if cond.Predicate != nil {
		attr := strings.ToLower(cond.Predicate.Attribute)
		return evalPredicate(cond.Predicate.Operator, e.ctrlAttrValue(attr, ctrl), cond.Predicate.Value)
	}
	return true
}

func (e *RuleEngine) evaluateConditionFramework(cond *Condition, tracking *ComplianceTracking) bool {
	if cond == nil {
		return true
	}
	if len(cond.All) > 0 {
		for i := range cond.All {
			if !e.evaluateConditionFramework(&cond.All[i], tracking) {
				return false
			}
		}
		return true
	}
	if len(cond.Any) > 0 {
		for i := range cond.Any {
			if e.evaluateConditionFramework(&cond.Any[i], tracking) {
				return true
			}
		}
		return false
	}
	if cond.Not != nil {
		return !e.evaluateConditionFramework(cond.Not, tracking)
	}
	if cond.Predicate != nil {
		attr := strings.ToLower(cond.Predicate.Attribute)
		return evalPredicate(cond.Predicate.Operator, e.frameworkAttrValue(attr, tracking), cond.Predicate.Value)
	}
	return true
}

// Attribute retrieval helpers

// filterTargetRequirementIDs returns requirement IDs filtered by targets (or all if empty)
func filterTargetRequirementIDs(targets []string, source map[string]*RequirementTracking) []string {
	if len(targets) == 0 {
		ids := make([]string, 0, len(source))
		for id := range source {
			ids = append(ids, id)
		}
		return ids
	}
	var ids []string
	for _, id := range targets {
		if _, ok := source[id]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}

// filterTargetControlIDs returns control IDs filtered by targets (or all if empty)
func filterTargetControlIDs(targets []string, source map[string]*ControlTracking) []string {
	if len(targets) == 0 {
		ids := make([]string, 0, len(source))
		for id := range source {
			ids = append(ids, id)
		}
		return ids
	}
	var ids []string
	for _, id := range targets {
		if _, ok := source[id]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}

func (e *RuleEngine) reqAttrValue(attr string, req *RequirementTracking) interface{} {
	switch attr {
	case "status":
		return string(req.Status)
	case "implementation_status":
		return string(req.ImplementationStatus)
	case "score", "compliance_score":
		return req.ComplianceScore
	case "evidence_count":
		return len(req.Evidence)
	case "exception_count":
		return len(req.Exceptions)
	case "control_count":
		return len(req.Controls)
	case "last_reviewed_ts":
		return req.LastReviewed.Unix()
	default:
		return nil
	}
}

func (e *RuleEngine) ctrlAttrValue(attr string, ctrl *ControlTracking) interface{} {
	switch attr {
	case "status":
		return string(ctrl.Status)
	case "implementation_status":
		return string(ctrl.ImplementationStatus)
	case "effectiveness":
		return string(ctrl.Effectiveness)
	case "test_count":
		return len(ctrl.TestResults)
	case "evidence_count":
		return len(ctrl.Evidence)
	case "pass_rate":
		return e.computeControlPassRate(ctrl)
	case "last_tested_ts":
		if ctrl.LastTested != nil {
			return ctrl.LastTested.Unix()
		}
		return int64(0)
	default:
		return nil
	}
}

func (e *RuleEngine) frameworkAttrValue(attr string, tracking *ComplianceTracking) interface{} {
	switch attr {
	case "overall_status":
		return string(tracking.OverallStatus)
	case "compliance_score":
		return tracking.ComplianceScore
	case "requirement_count":
		return len(tracking.Requirements)
	default:
		return nil
	}
}

func (e *RuleEngine) computeControlPassRate(ctrl *ControlTracking) float64 {
	if len(ctrl.TestResults) == 0 {
		return 0.0
	}
	passes := 0
	for i := range ctrl.TestResults {
		if ctrl.TestResults[i].Result == TestResultPass {
			passes++
		}
	}
	return float64(passes) / float64(len(ctrl.TestResults)) * 100.0
}

// Effects application (only if ApplyEffects)

func (e *RuleEngine) applyRequirementEffect(effect RuleEffect, req *RequirementTracking) {
	if effect.ScoreAdjustment != nil {
		req.ComplianceScore = clamp01(req.ComplianceScore + *effect.ScoreAdjustment)
	}
	if effect.SetStatus != nil {
		req.Status = *effect.SetStatus
	}
	// RequireEvidence is a policy flag; concrete enforcement would occur elsewhere
}

func (e *RuleEngine) applyControlEffect(effect RuleEffect, ctrl *ControlTracking) {
	if effect.SetStatus != nil {
		ctrl.Status = *effect.SetStatus
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

// Predicate evaluation core

func evalPredicate(op Operator, left interface{}, right interface{}) bool {
	switch op {
	case OpExists:
		return left != nil
	case OpNotExists:
		return left == nil
	case OpEq:
		return compareEq(left, right)
	case OpNe:
		return !compareEq(left, right)
	case OpGt:
		return compareOrd(left, right) > 0
	case OpGte:
		return compareOrd(left, right) >= 0
	case OpLt:
		return compareOrd(left, right) < 0
	case OpLte:
		return compareOrd(left, right) <= 0
	case OpContains:
		return contains(left, right)
	case OpIn:
		return inSet(left, right)
	case OpRegexMatch:
		return regexMatch(left, right)
	default:
		return false
	}
}

func compareEq(a, b interface{}) bool {
	switch av := a.(type) {
	case string:
		bv := toString(b)
		return strings.EqualFold(av, bv)
	case float64:
		bf := toFloat(b)
		return av == bf
	case int:
		bf := toFloat(b)
		return float64(av) == bf
	case int64:
		bf := toFloat(b)
		return float64(av) == bf
	case nil:
		return b == nil
	default:
		return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
	}
}

// compareOrd returns 1 if a>b, 0 if a==b, -1 if a<b for numeric/string
func compareOrd(a, b interface{}) int {
	// numeric first
	af := toFloat(a)
	bf := toFloat(b)
	if !isNaN(af) && !isNaN(bf) {
		if af > bf {
			return 1
		}
		if af < bf {
			return -1
		}
		return 0
	}
	// string fallback
	sa := strings.ToLower(toString(a))
	sb := strings.ToLower(toString(b))
	if sa > sb {
		return 1
	}
	if sa < sb {
		return -1
	}
	return 0
}

func contains(a, b interface{}) bool {
	sa := strings.ToLower(toString(a))
	sb := strings.ToLower(toString(b))
	return strings.Contains(sa, sb)
}

func inSet(a, b interface{}) bool {
	sa := strings.ToLower(toString(a))
	switch bv := b.(type) {
	case []string:
		for _, s := range bv {
			if strings.ToLower(s) == sa {
				return true
			}
		}
	case []interface{}:
		for _, s := range bv {
			if strings.ToLower(toString(s)) == sa {
				return true
			}
		}
	}
	return false
}

func regexMatch(a, b interface{}) bool {
	pattern := toString(b)
	if pattern == "" {
		return false
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(toString(a))
}

func toString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	case float64:
		return fmt.Sprintf("%g", t)
	case int:
		return fmt.Sprintf("%d", t)
	case int64:
		return fmt.Sprintf("%d", t)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", t)
	}
}

func toFloat(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case string:
		// best-effort parse
		var f float64
		_, _ = fmt.Sscan(t, &f)
		return f
	default:
		return 0
	}
}

func isNaN(f float64) bool { return f != f }
