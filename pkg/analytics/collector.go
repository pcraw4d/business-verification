package analytics

import (
	"context"
	"sync"
	"time"
)

// AnalyticsCollector collects and aggregates business analytics data
type AnalyticsCollector struct {
	mu sync.RWMutex

	// Business metrics
	totalClassifications      int64
	successfulClassifications int64
	failedClassifications     int64

	// Industry metrics
	industryCounts  map[string]int64
	riskLevelCounts map[string]int64

	// Performance metrics
	avgResponseTime   time.Duration
	totalResponseTime time.Duration
	responseCount     int64

	// User metrics
	userActivity map[string]*UserActivity

	// Time-based metrics
	dailyStats map[string]*DailyStats
}

// UserActivity tracks user-specific analytics
type UserActivity struct {
	UserID          string
	Classifications int64
	LastActivity    time.Time
	AvgResponseTime time.Duration
	SuccessRate     float64
}

// DailyStats tracks daily analytics
type DailyStats struct {
	Date            time.Time
	Classifications int64
	SuccessRate     float64
	AvgResponseTime time.Duration
	TopIndustries   []IndustryCount
	TopRiskLevels   []RiskLevelCount
}

// IndustryCount represents industry classification count
type IndustryCount struct {
	Industry   string
	Count      int64
	Percentage float64
}

// RiskLevelCount represents risk level count
type RiskLevelCount struct {
	RiskLevel  string
	Count      int64
	Percentage float64
}

// NewAnalyticsCollector creates a new analytics collector
func NewAnalyticsCollector() *AnalyticsCollector {
	return &AnalyticsCollector{
		industryCounts:  make(map[string]int64),
		riskLevelCounts: make(map[string]int64),
		userActivity:    make(map[string]*UserActivity),
		dailyStats:      make(map[string]*DailyStats),
	}
}

// RecordClassification records a classification event
func (ac *AnalyticsCollector) RecordClassification(ctx context.Context, userID string, success bool, responseTime time.Duration, industry string, riskLevel string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Update overall metrics
	ac.totalClassifications++
	if success {
		ac.successfulClassifications++
	} else {
		ac.failedClassifications++
	}

	// Update response time metrics
	ac.totalResponseTime += responseTime
	ac.responseCount++
	ac.avgResponseTime = ac.totalResponseTime / time.Duration(ac.responseCount)

	// Update industry counts
	if industry != "" {
		ac.industryCounts[industry]++
	}

	// Update risk level counts
	if riskLevel != "" {
		ac.riskLevelCounts[riskLevel]++
	}

	// Update user activity
	if userActivity, exists := ac.userActivity[userID]; exists {
		userActivity.Classifications++
		userActivity.LastActivity = time.Now()
		userActivity.AvgResponseTime = (userActivity.AvgResponseTime + responseTime) / 2
		if success {
			userActivity.SuccessRate = float64(userActivity.Classifications) / float64(ac.totalClassifications) * 100
		}
	} else {
		ac.userActivity[userID] = &UserActivity{
			UserID:          userID,
			Classifications: 1,
			LastActivity:    time.Now(),
			AvgResponseTime: responseTime,
			SuccessRate:     100.0,
		}
	}

	// Update daily stats
	dateKey := time.Now().Format("2006-01-02")
	if dailyStat, exists := ac.dailyStats[dateKey]; exists {
		dailyStat.Classifications++
		dailyStat.SuccessRate = float64(ac.successfulClassifications) / float64(ac.totalClassifications) * 100
		dailyStat.AvgResponseTime = ac.avgResponseTime
	} else {
		ac.dailyStats[dateKey] = &DailyStats{
			Date:            time.Now(),
			Classifications: 1,
			SuccessRate:     100.0,
			AvgResponseTime: responseTime,
			TopIndustries:   ac.getTopIndustries(5),
			TopRiskLevels:   ac.getTopRiskLevels(5),
		}
	}
}

// GetOverallStats returns overall analytics statistics
func (ac *AnalyticsCollector) GetOverallStats() map[string]interface{} {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	successRate := float64(0)
	if ac.totalClassifications > 0 {
		successRate = float64(ac.successfulClassifications) / float64(ac.totalClassifications) * 100
	}

	return map[string]interface{}{
		"total_classifications":      ac.totalClassifications,
		"successful_classifications": ac.successfulClassifications,
		"failed_classifications":     ac.failedClassifications,
		"success_rate":               successRate,
		"avg_response_time":          ac.avgResponseTime.String(),
		"top_industries":             ac.getTopIndustries(10),
		"top_risk_levels":            ac.getTopRiskLevels(10),
		"total_users":                len(ac.userActivity),
	}
}

// GetUserStats returns user-specific analytics
func (ac *AnalyticsCollector) GetUserStats(userID string) map[string]interface{} {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	userActivity, exists := ac.userActivity[userID]
	if !exists {
		return map[string]interface{}{
			"user_id":           userID,
			"classifications":   0,
			"last_activity":     nil,
			"avg_response_time": "0s",
			"success_rate":      0.0,
		}
	}

	return map[string]interface{}{
		"user_id":           userID,
		"classifications":   userActivity.Classifications,
		"last_activity":     userActivity.LastActivity,
		"avg_response_time": userActivity.AvgResponseTime.String(),
		"success_rate":      userActivity.SuccessRate,
	}
}

// GetDailyStats returns daily analytics statistics
func (ac *AnalyticsCollector) GetDailyStats(days int) map[string]interface{} {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	stats := make([]*DailyStats, 0)
	cutoff := time.Now().AddDate(0, 0, -days)

	for _, dailyStat := range ac.dailyStats {
		if dailyStat.Date.After(cutoff) {
			stats = append(stats, dailyStat)
		}
	}

	return map[string]interface{}{
		"daily_stats": stats,
		"period_days": days,
		"total_days":  len(stats),
	}
}

// GetIndustryAnalytics returns industry-specific analytics
func (ac *AnalyticsCollector) GetIndustryAnalytics() map[string]interface{} {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	return map[string]interface{}{
		"industry_counts":  ac.industryCounts,
		"top_industries":   ac.getTopIndustries(20),
		"total_industries": len(ac.industryCounts),
	}
}

// GetRiskAnalytics returns risk-level analytics
func (ac *AnalyticsCollector) GetRiskAnalytics() map[string]interface{} {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	return map[string]interface{}{
		"risk_level_counts": ac.riskLevelCounts,
		"top_risk_levels":   ac.getTopRiskLevels(20),
		"total_risk_levels": len(ac.riskLevelCounts),
	}
}

// Helper methods
func (ac *AnalyticsCollector) getTopIndustries(limit int) []IndustryCount {
	topIndustries := make([]IndustryCount, 0)

	// Sort industries by count
	for industry, count := range ac.industryCounts {
		percentage := float64(0)
		if ac.totalClassifications > 0 {
			percentage = float64(count) / float64(ac.totalClassifications) * 100
		}

		topIndustries = append(topIndustries, IndustryCount{
			Industry:   industry,
			Count:      count,
			Percentage: percentage,
		})
	}

	// Sort by count (descending)
	for i := 0; i < len(topIndustries); i++ {
		for j := i + 1; j < len(topIndustries); j++ {
			if topIndustries[i].Count < topIndustries[j].Count {
				topIndustries[i], topIndustries[j] = topIndustries[j], topIndustries[i]
			}
		}
	}

	// Limit results
	if len(topIndustries) > limit {
		topIndustries = topIndustries[:limit]
	}

	return topIndustries
}

func (ac *AnalyticsCollector) getTopRiskLevels(limit int) []RiskLevelCount {
	topRiskLevels := make([]RiskLevelCount, 0)

	// Sort risk levels by count
	for riskLevel, count := range ac.riskLevelCounts {
		percentage := float64(0)
		if ac.totalClassifications > 0 {
			percentage = float64(count) / float64(ac.totalClassifications) * 100
		}

		topRiskLevels = append(topRiskLevels, RiskLevelCount{
			RiskLevel:  riskLevel,
			Count:      count,
			Percentage: percentage,
		})
	}

	// Sort by count (descending)
	for i := 0; i < len(topRiskLevels); i++ {
		for j := i + 1; j < len(topRiskLevels); j++ {
			if topRiskLevels[i].Count < topRiskLevels[j].Count {
				topRiskLevels[i], topRiskLevels[j] = topRiskLevels[j], topRiskLevels[i]
			}
		}
	}

	// Limit results
	if len(topRiskLevels) > limit {
		topRiskLevels = topRiskLevels[:limit]
	}

	return topRiskLevels
}
