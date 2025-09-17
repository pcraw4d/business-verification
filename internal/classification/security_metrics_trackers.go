package classification

import (
	"fmt"
	"time"
)

// DataSourceTrustTracker methods

// RecordTrust records a data source trust validation
func (dst *DataSourceTrustTracker) RecordTrust(trusted bool, source string) {
	dst.mu.Lock()
	defer dst.mu.Unlock()

	dst.totalValidations++
	if trusted {
		dst.trustedCount++
	} else {
		dst.untrustedCount++
	}

	dst.trustRate = float64(dst.trustedCount) / float64(dst.totalValidations) * 100
	dst.lastUpdated = time.Now()

	// Add to historical data
	dst.historicalRates = append(dst.historicalRates, TrustRateDataPoint{
		Timestamp:    time.Now(),
		TrustRate:    dst.trustRate,
		TrustedCount: dst.trustedCount,
		TotalCount:   dst.totalValidations,
	})
}

// GetTrustRate returns the current trust rate
func (dst *DataSourceTrustTracker) GetTrustRate() float64 {
	dst.mu.RLock()
	defer dst.mu.RUnlock()
	return dst.trustRate
}

// GetTrustedCount returns the count of trusted sources
func (dst *DataSourceTrustTracker) GetTrustedCount() int64 {
	dst.mu.RLock()
	defer dst.mu.RUnlock()
	return dst.trustedCount
}

// GetTotalValidations returns the total number of validations
func (dst *DataSourceTrustTracker) GetTotalValidations() int64 {
	dst.mu.RLock()
	defer dst.mu.RUnlock()
	return dst.totalValidations
}

// GetLastUpdated returns the last update time
func (dst *DataSourceTrustTracker) GetLastUpdated() time.Time {
	dst.mu.RLock()
	defer dst.mu.RUnlock()
	return dst.lastUpdated
}

// UpdateMetrics updates internal metrics
func (dst *DataSourceTrustTracker) UpdateMetrics() {
	dst.mu.Lock()
	defer dst.mu.Unlock()

	if dst.totalValidations > 0 {
		dst.trustRate = float64(dst.trustedCount) / float64(dst.totalValidations) * 100
	}
	dst.lastUpdated = time.Now()
}

// CleanupOldData removes old historical data
func (dst *DataSourceTrustTracker) CleanupOldData(cutoffTime time.Time) {
	dst.mu.Lock()
	defer dst.mu.Unlock()

	var filtered []TrustRateDataPoint
	for _, dataPoint := range dst.historicalRates {
		if dataPoint.Timestamp.After(cutoffTime) {
			filtered = append(filtered, dataPoint)
		}
	}
	dst.historicalRates = filtered
}

// WebsiteVerificationTracker methods

// RecordVerification records a website verification result
func (wvt *WebsiteVerificationTracker) RecordVerification(success bool, domain, method string) {
	wvt.mu.Lock()
	defer wvt.mu.Unlock()

	wvt.totalAttempts++
	if success {
		wvt.successCount++
	} else {
		wvt.failureCount++
	}

	wvt.successRate = float64(wvt.successCount) / float64(wvt.totalAttempts) * 100
	wvt.lastUpdated = time.Now()

	// Add to historical data
	wvt.historicalRates = append(wvt.historicalRates, VerificationRateDataPoint{
		Timestamp:     time.Now(),
		SuccessRate:   wvt.successRate,
		SuccessCount:  wvt.successCount,
		TotalAttempts: wvt.totalAttempts,
	})
}

// GetSuccessRate returns the current success rate
func (wvt *WebsiteVerificationTracker) GetSuccessRate() float64 {
	wvt.mu.RLock()
	defer wvt.mu.RUnlock()
	return wvt.successRate
}

// GetSuccessCount returns the count of successful verifications
func (wvt *WebsiteVerificationTracker) GetSuccessCount() int64 {
	wvt.mu.RLock()
	defer wvt.mu.RUnlock()
	return wvt.successCount
}

// GetTotalAttempts returns the total number of verification attempts
func (wvt *WebsiteVerificationTracker) GetTotalAttempts() int64 {
	wvt.mu.RLock()
	defer wvt.mu.RUnlock()
	return wvt.totalAttempts
}

// GetLastUpdated returns the last update time
func (wvt *WebsiteVerificationTracker) GetLastUpdated() time.Time {
	wvt.mu.RLock()
	defer wvt.mu.RUnlock()
	return wvt.lastUpdated
}

// UpdateMetrics updates internal metrics
func (wvt *WebsiteVerificationTracker) UpdateMetrics() {
	wvt.mu.Lock()
	defer wvt.mu.Unlock()

	if wvt.totalAttempts > 0 {
		wvt.successRate = float64(wvt.successCount) / float64(wvt.totalAttempts) * 100
	}
	wvt.lastUpdated = time.Now()
}

// CleanupOldData removes old historical data
func (wvt *WebsiteVerificationTracker) CleanupOldData(cutoffTime time.Time) {
	wvt.mu.Lock()
	defer wvt.mu.Unlock()

	var filtered []VerificationRateDataPoint
	for _, dataPoint := range wvt.historicalRates {
		if dataPoint.Timestamp.After(cutoffTime) {
			filtered = append(filtered, dataPoint)
		}
	}
	wvt.historicalRates = filtered
}

// SecurityViolationTracker methods

// RecordViolation records a security violation
func (svt *SecurityViolationTracker) RecordViolation(violation SecurityViolation) {
	svt.mu.Lock()
	defer svt.mu.Unlock()

	svt.totalViolations++
	svt.violationsByType[violation.Type]++
	svt.recentViolations = append(svt.recentViolations, violation)

	// Keep only recent violations (last 100)
	if len(svt.recentViolations) > 100 {
		svt.recentViolations = svt.recentViolations[len(svt.recentViolations)-100:]
	}
}

// GetTotalViolations returns the total number of violations
func (svt *SecurityViolationTracker) GetTotalViolations() int64 {
	svt.mu.RLock()
	defer svt.mu.RUnlock()
	return svt.totalViolations
}

// GetViolationsByType returns violations grouped by type
func (svt *SecurityViolationTracker) GetViolationsByType() map[string]int64 {
	svt.mu.RLock()
	defer svt.mu.RUnlock()

	result := make(map[string]int64)
	for k, v := range svt.violationsByType {
		result[k] = v
	}
	return result
}

// GetRecentViolations returns recent violations
func (svt *SecurityViolationTracker) GetRecentViolations(limit int) []SecurityViolation {
	svt.mu.RLock()
	defer svt.mu.RUnlock()

	if limit <= 0 || limit >= len(svt.recentViolations) {
		return svt.recentViolations
	}

	start := len(svt.recentViolations) - limit
	return svt.recentViolations[start:]
}

// GetLastUpdated returns the last update time
func (svt *SecurityViolationTracker) GetLastUpdated() time.Time {
	svt.mu.RLock()
	defer svt.mu.RUnlock()

	if len(svt.recentViolations) == 0 {
		return time.Time{}
	}
	return svt.recentViolations[len(svt.recentViolations)-1].Timestamp
}

// ConfidenceIntegrityTracker methods

// RecordScore records a confidence score for integrity tracking
func (cit *ConfidenceIntegrityTracker) RecordScore(score float64, valid bool, source string) {
	cit.mu.Lock()
	defer cit.mu.Unlock()

	cit.totalScores++
	if valid {
		cit.validScores++
	} else {
		cit.invalidScores++
	}

	cit.integrityRate = float64(cit.validScores) / float64(cit.totalScores)
	cit.lastUpdated = time.Now()

	// Add to historical data
	cit.historicalRates = append(cit.historicalRates, ConfidenceIntegrityDataPoint{
		Timestamp:     time.Now(),
		IntegrityRate: cit.integrityRate,
		ValidScores:   cit.validScores,
		TotalScores:   cit.totalScores,
	})
}

// GetIntegrityRate returns the current integrity rate
func (cit *ConfidenceIntegrityTracker) GetIntegrityRate() float64 {
	cit.mu.RLock()
	defer cit.mu.RUnlock()
	return cit.integrityRate
}

// GetValidScores returns the count of valid scores
func (cit *ConfidenceIntegrityTracker) GetValidScores() int64 {
	cit.mu.RLock()
	defer cit.mu.RUnlock()
	return cit.validScores
}

// GetTotalScores returns the total number of scores
func (cit *ConfidenceIntegrityTracker) GetTotalScores() int64 {
	cit.mu.RLock()
	defer cit.mu.RUnlock()
	return cit.totalScores
}

// GetLastUpdated returns the last update time
func (cit *ConfidenceIntegrityTracker) GetLastUpdated() time.Time {
	cit.mu.RLock()
	defer cit.mu.RUnlock()
	return cit.lastUpdated
}

// UpdateMetrics updates internal metrics
func (cit *ConfidenceIntegrityTracker) UpdateMetrics() {
	cit.mu.Lock()
	defer cit.mu.Unlock()

	if cit.totalScores > 0 {
		cit.integrityRate = float64(cit.validScores) / float64(cit.totalScores)
	}
	cit.lastUpdated = time.Now()
}

// CleanupOldData removes old historical data
func (cit *ConfidenceIntegrityTracker) CleanupOldData(cutoffTime time.Time) {
	cit.mu.Lock()
	defer cit.mu.Unlock()

	var filtered []ConfidenceIntegrityDataPoint
	for _, dataPoint := range cit.historicalRates {
		if dataPoint.Timestamp.After(cutoffTime) {
			filtered = append(filtered, dataPoint)
		}
	}
	cit.historicalRates = filtered
}

// SecurityAlertManager methods

// CreateAlert creates a new security alert
func (sam *SecurityAlertManager) CreateAlert(alertType, severity, message, metric string, value, threshold float64) {
	if !sam.enabled {
		return
	}

	sam.mu.Lock()
	defer sam.mu.Unlock()

	// Check cooldown to prevent spam
	lastAlertTime, exists := sam.lastAlertTimes[alertType]
	if exists && time.Since(lastAlertTime) < sam.alertCooldown {
		return
	}

	alert := SecurityAlert{
		ID:           fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		Timestamp:    time.Now(),
		Type:         alertType,
		Severity:     severity,
		Message:      message,
		Metric:       metric,
		Value:        value,
		Threshold:    threshold,
		Acknowledged: false,
	}

	sam.alerts = append(sam.alerts, alert)
	sam.lastAlertTimes[alertType] = time.Now()

	// Keep only recent alerts (last 100)
	if len(sam.alerts) > 100 {
		sam.alerts = sam.alerts[len(sam.alerts)-100:]
	}
}

// GetRecentAlerts returns recent alerts
func (sam *SecurityAlertManager) GetRecentAlerts(limit int) []SecurityAlert {
	sam.mu.RLock()
	defer sam.mu.RUnlock()

	if limit <= 0 || limit >= len(sam.alerts) {
		return sam.alerts
	}

	start := len(sam.alerts) - limit
	return sam.alerts[start:]
}

// CleanupOldAlerts removes old alerts
func (sam *SecurityAlertManager) CleanupOldAlerts(cutoffTime time.Time) {
	sam.mu.Lock()
	defer sam.mu.Unlock()

	var filtered []SecurityAlert
	for _, alert := range sam.alerts {
		if alert.Timestamp.After(cutoffTime) {
			filtered = append(filtered, alert)
		}
	}
	sam.alerts = filtered
}

// SecurityPerformanceTracker methods

// RecordCollectionTime records collection time
func (spt *SecurityPerformanceTracker) RecordCollectionTime(duration time.Duration) {
	if !spt.enabled {
		return
	}

	spt.mu.Lock()
	defer spt.mu.Unlock()

	spt.collectionTimes = append(spt.collectionTimes, duration)

	// Keep only recent times (last 100)
	if len(spt.collectionTimes) > 100 {
		spt.collectionTimes = spt.collectionTimes[len(spt.collectionTimes)-100:]
	}
}

// RecordProcessingTime records processing time
func (spt *SecurityPerformanceTracker) RecordProcessingTime(duration time.Duration) {
	if !spt.enabled {
		return
	}

	spt.mu.Lock()
	defer spt.mu.Unlock()

	spt.processingTimes = append(spt.processingTimes, duration)

	// Keep only recent times (last 100)
	if len(spt.processingTimes) > 100 {
		spt.processingTimes = spt.processingTimes[len(spt.processingTimes)-100:]
	}
}

// RecordAlertTime records alert processing time
func (spt *SecurityPerformanceTracker) RecordAlertTime(duration time.Duration) {
	if !spt.enabled {
		return
	}

	spt.mu.Lock()
	defer spt.mu.Unlock()

	spt.alertTimes = append(spt.alertTimes, duration)

	// Keep only recent times (last 100)
	if len(spt.alertTimes) > 100 {
		spt.alertTimes = spt.alertTimes[len(spt.alertTimes)-100:]
	}
}

// GetPerformanceMetrics returns performance metrics
func (spt *SecurityPerformanceTracker) GetPerformanceMetrics() *SecurityPerformanceMetrics {
	spt.mu.RLock()
	defer spt.mu.RUnlock()

	metrics := &SecurityPerformanceMetrics{
		TotalCollections: int64(len(spt.collectionTimes)),
		LastUpdated:      time.Now(),
	}

	// Calculate average collection time
	if len(spt.collectionTimes) > 0 {
		var total time.Duration
		for _, duration := range spt.collectionTimes {
			total += duration
		}
		metrics.AverageCollectionTime = total / time.Duration(len(spt.collectionTimes))
	}

	// Calculate average processing time
	if len(spt.processingTimes) > 0 {
		var total time.Duration
		for _, duration := range spt.processingTimes {
			total += duration
		}
		metrics.AverageProcessingTime = total / time.Duration(len(spt.processingTimes))
	}

	// Calculate average alert time
	if len(spt.alertTimes) > 0 {
		var total time.Duration
		for _, duration := range spt.alertTimes {
			total += duration
		}
		metrics.AverageAlertTime = total / time.Duration(len(spt.alertTimes))
	}

	return metrics
}

// CleanupOldData removes old performance data
func (spt *SecurityPerformanceTracker) CleanupOldData(cutoffTime time.Time) {
	spt.mu.Lock()
	defer spt.mu.Unlock()

	// Keep only recent data (last 100 entries)
	if len(spt.collectionTimes) > 100 {
		spt.collectionTimes = spt.collectionTimes[len(spt.collectionTimes)-100:]
	}
	if len(spt.processingTimes) > 100 {
		spt.processingTimes = spt.processingTimes[len(spt.processingTimes)-100:]
	}
	if len(spt.alertTimes) > 100 {
		spt.alertTimes = spt.alertTimes[len(spt.alertTimes)-100:]
	}
}
