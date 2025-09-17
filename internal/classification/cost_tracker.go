package classification

import (
	"context"
	"fmt"
	"time"

	"github.com/pcraw4d/business-verification/internal/shared"
)

// NewCostTracker creates a new cost tracker
func NewCostTracker(logger shared.Logger) *CostTracker {
	tracker := &CostTracker{
		tierBudgets: make(map[CustomerTier]*TierBudget),
		globalBudget: &GlobalBudget{
			DailyBudgetLimit:   100.0,  // $100/day
			MonthlyBudgetLimit: 2000.0, // $2000/month
			EmergencyThreshold: 0.8,    // 80% of budget
		},
		logger: logger,
	}

	// Initialize tier budgets
	tracker.initializeTierBudgets()

	return tracker
}

// initializeTierBudgets initializes budget tracking for all tiers
func (ct *CostTracker) initializeTierBudgets() {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	now := time.Now()
	nextMonth := now.AddDate(0, 1, 0)
	nextMonth = time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, nextMonth.Location())

	// Initialize budgets for each tier
	tierBudgets := map[CustomerTier]float64{
		CustomerTierFree:       10.0,   // $10/month
		CustomerTierStandard:   50.0,   // $50/month
		CustomerTierPremium:    200.0,  // $200/month
		CustomerTierEnterprise: 1000.0, // $1000/month
	}

	for tier, budget := range tierBudgets {
		ct.tierBudgets[tier] = &TierBudget{
			Tier:            tier,
			MonthlyBudget:   budget,
			UsedBudget:      0.0,
			RemainingBudget: budget,
			CallCount:       0,
			LastReset:       now,
			NextReset:       nextMonth,
		}
	}

	ct.logger.Log(context.Background(), shared.LogLevelInfo, "Initialized cost tracking", map[string]interface{}{
		"tier_count": len(ct.tierBudgets),
	})
}

// CheckBudgetConstraints checks if a request can be processed within budget constraints
func (ct *CostTracker) CheckBudgetConstraints(tier CustomerTier, config *TierConfig) error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	// Check global budget constraints
	if err := ct.checkGlobalBudgetConstraints(); err != nil {
		return fmt.Errorf("global budget constraint: %w", err)
	}

	// Check tier-specific budget constraints
	if err := ct.checkTierBudgetConstraints(tier, config); err != nil {
		return fmt.Errorf("tier budget constraint: %w", err)
	}

	return nil
}

// checkGlobalBudgetConstraints checks global budget constraints
func (ct *CostTracker) checkGlobalBudgetConstraints() error {
	now := time.Now()

	// Reset daily budget if needed
	if ct.shouldResetDailyBudget(now) {
		ct.resetDailyBudget(now)
	}

	// Reset monthly budget if needed
	if ct.shouldResetMonthlyBudget(now) {
		ct.resetMonthlyBudget(now)
	}

	// Check daily budget
	if ct.globalBudget.UsedDailyBudget >= ct.globalBudget.DailyBudgetLimit {
		return fmt.Errorf("daily budget limit exceeded: $%.2f/$%.2f",
			ct.globalBudget.UsedDailyBudget, ct.globalBudget.DailyBudgetLimit)
	}

	// Check monthly budget
	if ct.globalBudget.UsedMonthlyBudget >= ct.globalBudget.MonthlyBudgetLimit {
		return fmt.Errorf("monthly budget limit exceeded: $%.2f/$%.2f",
			ct.globalBudget.UsedMonthlyBudget, ct.globalBudget.MonthlyBudgetLimit)
	}

	// Check emergency threshold
	dailyThreshold := ct.globalBudget.DailyBudgetLimit * ct.globalBudget.EmergencyThreshold
	monthlyThreshold := ct.globalBudget.MonthlyBudgetLimit * ct.globalBudget.EmergencyThreshold

	if ct.globalBudget.UsedDailyBudget >= dailyThreshold {
		ct.logger.Log(context.Background(), shared.LogLevelWarning, "Daily budget approaching limit", map[string]interface{}{
			"used_budget": ct.globalBudget.UsedDailyBudget,
			"limit":       ct.globalBudget.DailyBudgetLimit,
			"percentage":  (ct.globalBudget.UsedDailyBudget / ct.globalBudget.DailyBudgetLimit) * 100,
		})
	}

	if ct.globalBudget.UsedMonthlyBudget >= monthlyThreshold {
		ct.logger.Log(context.Background(), shared.LogLevelWarning, "Monthly budget approaching limit", map[string]interface{}{
			"used_budget": ct.globalBudget.UsedMonthlyBudget,
			"limit":       ct.globalBudget.MonthlyBudgetLimit,
			"percentage":  (ct.globalBudget.UsedMonthlyBudget / ct.globalBudget.MonthlyBudgetLimit) * 100,
		})
	}

	return nil
}

// checkTierBudgetConstraints checks tier-specific budget constraints
func (ct *CostTracker) checkTierBudgetConstraints(tier CustomerTier, config *TierConfig) error {
	budget, exists := ct.tierBudgets[tier]
	if !exists {
		return fmt.Errorf("no budget tracking for tier %s", tier)
	}

	// Reset tier budget if needed
	if ct.shouldResetTierBudget(budget) {
		ct.resetTierBudget(budget)
	}

	// Check tier budget
	if budget.UsedBudget >= budget.MonthlyBudget {
		return fmt.Errorf("tier %s budget limit exceeded: $%.2f/$%.2f",
			tier, budget.UsedBudget, budget.MonthlyBudget)
	}

	// Check per-call cost limit
	if config.MaxCostPerCall > 0 && config.MaxCostPerCall > budget.RemainingBudget {
		return fmt.Errorf("insufficient budget for tier %s: need $%.4f, have $%.2f",
			tier, config.MaxCostPerCall, budget.RemainingBudget)
	}

	return nil
}

// RecordCost records the cost of a classification request
func (ct *CostTracker) RecordCost(tier CustomerTier, cost float64, method string) error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	// Update tier budget
	budget, exists := ct.tierBudgets[tier]
	if !exists {
		return fmt.Errorf("no budget tracking for tier %s", tier)
	}

	budget.UsedBudget += cost
	budget.RemainingBudget = budget.MonthlyBudget - budget.UsedBudget
	budget.CallCount++

	// Update average cost per call
	if budget.CallCount > 0 {
		budget.AverageCostPerCall = budget.UsedBudget / float64(budget.CallCount)
	}

	// Update global budget
	ct.globalBudget.UsedDailyBudget += cost
	ct.globalBudget.UsedMonthlyBudget += cost

	ct.logger.Log(context.Background(), shared.LogLevelInfo, "Recorded cost", map[string]interface{}{
		"tier":           tier,
		"cost":           cost,
		"method":         method,
		"total_used":     budget.UsedBudget,
		"monthly_budget": budget.MonthlyBudget,
	})

	return nil
}

// GetBudgetStatus returns the current budget status for a tier
func (ct *CostTracker) GetBudgetStatus(tier CustomerTier) (*TierBudget, error) {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	budget, exists := ct.tierBudgets[tier]
	if !exists {
		return nil, fmt.Errorf("no budget tracking for tier %s", tier)
	}

	// Create a copy to avoid race conditions
	budgetCopy := *budget
	return &budgetCopy, nil
}

// GetGlobalBudgetStatus returns the current global budget status
func (ct *CostTracker) GetGlobalBudgetStatus() *GlobalBudget {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	// Create a copy to avoid race conditions
	budgetCopy := *ct.globalBudget
	return &budgetCopy
}

// GetAllBudgetStatuses returns budget statuses for all tiers
func (ct *CostTracker) GetAllBudgetStatuses() map[CustomerTier]*TierBudget {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	statuses := make(map[CustomerTier]*TierBudget)
	for tier, budget := range ct.tierBudgets {
		// Create a copy to avoid race conditions
		budgetCopy := *budget
		statuses[tier] = &budgetCopy
	}

	return statuses
}

// UpdateBudgetLimits updates budget limits for a tier
func (ct *CostTracker) UpdateBudgetLimits(tier CustomerTier, monthlyBudget float64) error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	budget, exists := ct.tierBudgets[tier]
	if !exists {
		return fmt.Errorf("no budget tracking for tier %s", tier)
	}

	oldBudget := budget.MonthlyBudget
	budget.MonthlyBudget = monthlyBudget
	budget.RemainingBudget = monthlyBudget - budget.UsedBudget

	ct.logger.Log(context.Background(), shared.LogLevelInfo, "Updated budget for tier", map[string]interface{}{
		"tier":       tier,
		"old_budget": oldBudget,
		"new_budget": monthlyBudget,
	})

	return nil
}

// UpdateGlobalBudgetLimits updates global budget limits
func (ct *CostTracker) UpdateGlobalBudgetLimits(dailyLimit, monthlyLimit float64) error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	oldDaily := ct.globalBudget.DailyBudgetLimit
	oldMonthly := ct.globalBudget.MonthlyBudgetLimit

	ct.globalBudget.DailyBudgetLimit = dailyLimit
	ct.globalBudget.MonthlyBudgetLimit = monthlyLimit

	ct.logger.Log(context.Background(), shared.LogLevelInfo, "Updated global budget limits", map[string]interface{}{
		"old_daily":   oldDaily,
		"new_daily":   dailyLimit,
		"old_monthly": oldMonthly,
		"new_monthly": monthlyLimit,
	})

	return nil
}

// ResetBudget resets the budget for a specific tier
func (ct *CostTracker) ResetBudget(tier CustomerTier) error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	budget, exists := ct.tierBudgets[tier]
	if !exists {
		return fmt.Errorf("no budget tracking for tier %s", tier)
	}

	ct.resetTierBudget(budget)

	ct.logger.Log(context.Background(), shared.LogLevelInfo, "Reset budget for tier", map[string]interface{}{
		"tier": tier,
	})

	return nil
}

// ResetAllBudgets resets all budgets
func (ct *CostTracker) ResetAllBudgets() error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	now := time.Now()

	// Reset all tier budgets
	for _, budget := range ct.tierBudgets {
		ct.resetTierBudget(budget)
	}

	// Reset global budget
	ct.resetDailyBudget(now)
	ct.resetMonthlyBudget(now)

	ct.logger.Log(context.Background(), shared.LogLevelInfo, "Reset all budgets", map[string]interface{}{})

	return nil
}

// Helper functions for budget management
func (ct *CostTracker) shouldResetDailyBudget(now time.Time) bool {
	// Reset at midnight
	return now.Hour() == 0 && now.Minute() == 0
}

func (ct *CostTracker) shouldResetMonthlyBudget(now time.Time) bool {
	// Reset on the first day of the month
	return now.Day() == 1 && now.Hour() == 0 && now.Minute() == 0
}

func (ct *CostTracker) shouldResetTierBudget(budget *TierBudget) bool {
	now := time.Now()
	return now.After(budget.NextReset)
}

func (ct *CostTracker) resetDailyBudget(now time.Time) {
	ct.globalBudget.UsedDailyBudget = 0.0
	ct.logger.Log(context.Background(), shared.LogLevelInfo, "Reset daily budget", map[string]interface{}{})
}

func (ct *CostTracker) resetMonthlyBudget(now time.Time) {
	ct.globalBudget.UsedMonthlyBudget = 0.0
	ct.logger.Log(context.Background(), shared.LogLevelInfo, "Reset monthly budget", map[string]interface{}{})
}

func (ct *CostTracker) resetTierBudget(budget *TierBudget) {
	now := time.Now()
	nextMonth := now.AddDate(0, 1, 0)
	nextMonth = time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, nextMonth.Location())

	budget.UsedBudget = 0.0
	budget.RemainingBudget = budget.MonthlyBudget
	budget.CallCount = 0
	budget.AverageCostPerCall = 0.0
	budget.LastReset = now
	budget.NextReset = nextMonth

	ct.logger.Log(context.Background(), shared.LogLevelInfo, "Reset budget for tier", map[string]interface{}{
		"tier": budget.Tier,
	})
}

// GetCostAnalytics returns cost analytics for reporting
func (ct *CostTracker) GetCostAnalytics(ctx context.Context) (*CostAnalytics, error) {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	analytics := &CostAnalytics{
		GeneratedAt:       time.Now(),
		GlobalBudget:      *ct.globalBudget,
		TierBudgets:       make(map[CustomerTier]*TierBudget),
		TotalCostsByTier:  make(map[CustomerTier]float64),
		TotalCallsByTier:  make(map[CustomerTier]int64),
		AverageCostByTier: make(map[CustomerTier]float64),
		CostEfficiency:    make(map[CustomerTier]float64),
	}

	// Calculate analytics for each tier
	for tier, budget := range ct.tierBudgets {
		// Create a copy
		budgetCopy := *budget
		analytics.TierBudgets[tier] = &budgetCopy

		// Calculate metrics
		analytics.TotalCostsByTier[tier] = budget.UsedBudget
		analytics.TotalCallsByTier[tier] = budget.CallCount

		if budget.CallCount > 0 {
			analytics.AverageCostByTier[tier] = budget.AverageCostPerCall
			// Cost efficiency = accuracy per dollar spent (simplified)
			analytics.CostEfficiency[tier] = 0.8 / budget.AverageCostPerCall // Placeholder calculation
		}
	}

	// Calculate global metrics
	analytics.TotalGlobalCost = ct.globalBudget.UsedMonthlyBudget
	analytics.TotalGlobalCalls = 0
	for _, calls := range analytics.TotalCallsByTier {
		analytics.TotalGlobalCalls += calls
	}

	if analytics.TotalGlobalCalls > 0 {
		analytics.AverageGlobalCost = analytics.TotalGlobalCost / float64(analytics.TotalGlobalCalls)
	}

	return analytics, nil
}

// CostAnalytics represents cost analytics data
type CostAnalytics struct {
	GeneratedAt       time.Time                    `json:"generated_at"`
	GlobalBudget      GlobalBudget                 `json:"global_budget"`
	TierBudgets       map[CustomerTier]*TierBudget `json:"tier_budgets"`
	TotalCostsByTier  map[CustomerTier]float64     `json:"total_costs_by_tier"`
	TotalCallsByTier  map[CustomerTier]int64       `json:"total_calls_by_tier"`
	AverageCostByTier map[CustomerTier]float64     `json:"average_cost_by_tier"`
	CostEfficiency    map[CustomerTier]float64     `json:"cost_efficiency"`
	TotalGlobalCost   float64                      `json:"total_global_cost"`
	TotalGlobalCalls  int64                        `json:"total_global_calls"`
	AverageGlobalCost float64                      `json:"average_global_cost"`
}

// ExportCostData exports cost data for external analysis
func (ct *CostTracker) ExportCostData(ctx context.Context, format string) ([]byte, error) {
	analytics, err := ct.GetCostAnalytics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost analytics: %w", err)
	}

	switch format {
	case "json":
		return ct.exportJSON(analytics)
	case "csv":
		return ct.exportCSV(analytics)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// exportJSON exports cost data as JSON
func (ct *CostTracker) exportJSON(analytics *CostAnalytics) ([]byte, error) {
	// This would use json.Marshal in a real implementation
	// For now, return a placeholder
	return []byte(`{"message": "JSON export not implemented"}`), nil
}

// exportCSV exports cost data as CSV
func (ct *CostTracker) exportCSV(analytics *CostAnalytics) ([]byte, error) {
	// This would generate CSV data in a real implementation
	// For now, return a placeholder
	return []byte(`tier,used_budget,monthly_budget,call_count,average_cost_per_call`), nil
}
