package handlers

import (
	"time"
)

// GovernancePolicy represents a data governance policy
type GovernancePolicy struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Rules       []PolicyRule `json:"rules"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// PolicyRule represents a policy rule
type PolicyRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Enabled     bool   `json:"enabled"`
}

// GovernanceStatistics represents governance statistics
type GovernanceStatistics struct {
	TotalPolicies   int     `json:"total_policies"`
	ActivePolicies  int     `json:"active_policies"`
	TotalRules      int     `json:"total_rules"`
	ActiveRules     int     `json:"active_rules"`
	ComplianceScore float64 `json:"compliance_score"`
}

// LineageProcess represents a data lineage process
type LineageProcess struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LineageJob represents a data lineage job
type LineageJob struct {
	ID          string     `json:"id"`
	ProcessID   string     `json:"process_id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// QualityIssue represents a data quality issue
type QualityIssue struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Severity    string     `json:"severity"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Enabled     bool   `json:"enabled"`
	Expression  string `json:"expression"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
}

// UpdateConfigRequest represents a request to update configuration
type UpdateConfigRequest struct {
	Key    string                 `json:"key"`
	Value  interface{}            `json:"value"`
	Config map[string]interface{} `json:"config"`
}

// WorkflowStatus represents workflow status
type WorkflowStatus string

const (
	WorkflowStatusPending    WorkflowStatus = "pending"
	WorkflowStatusInProgress WorkflowStatus = "in_progress"
	WorkflowStatusCompleted  WorkflowStatus = "completed"
	WorkflowStatusFailed     WorkflowStatus = "failed"
	WorkflowStatusCancelled  WorkflowStatus = "cancelled"
)
