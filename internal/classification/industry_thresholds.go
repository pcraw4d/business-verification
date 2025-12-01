package classification

import (
	"strings"
	"sync"
)

// IndustryThresholds manages industry-specific confidence thresholds
// OPTIMIZATION #16: Different confidence thresholds per industry
type IndustryThresholds struct {
	thresholds map[string]float64
	defaultThreshold float64
	minThreshold float64
	maxThreshold float64
	mu sync.RWMutex
}

// NewIndustryThresholds creates a new industry thresholds manager with default values
func NewIndustryThresholds() *IndustryThresholds {
	thresholds := make(map[string]float64)
	
	// High-risk industries require higher confidence
	thresholds["Financial Services"] = 0.7
	thresholds["Finance"] = 0.7
	thresholds["Fintech"] = 0.7
	thresholds["Insurance"] = 0.7
	
	// Healthcare requires high confidence
	thresholds["Healthcare"] = 0.65
	thresholds["Medical Technology"] = 0.65
	
	// Legal and professional services require high confidence
	thresholds["Legal"] = 0.6
	thresholds["Professional, Scientific, and Technical Services"] = 0.6
	
	// Medium-risk industries
	thresholds["Real Estate and Rental and Leasing"] = 0.5
	thresholds["Construction"] = 0.5
	thresholds["Manufacturing"] = 0.45
	
	// Lower-risk industries use default threshold (0.3)
	// Technology, Software Development, Retail, E-commerce, Food & Beverage, etc.
	
	return &IndustryThresholds{
		thresholds: thresholds,
		defaultThreshold: 0.3,
		minThreshold: 0.2,
		maxThreshold: 0.9,
	}
}

// GetThreshold returns the confidence threshold for a given industry
// Returns the industry-specific threshold if found, otherwise returns the default
func (it *IndustryThresholds) GetThreshold(industryName string) float64 {
	it.mu.RLock()
	defer it.mu.RUnlock()
	
	// Normalize industry name (case-insensitive, trim whitespace)
	normalized := strings.TrimSpace(strings.ToLower(industryName))
	
	// Check exact match first
	if threshold, exists := it.thresholds[industryName]; exists {
		return it.clampThreshold(threshold)
	}
	
	// Check case-insensitive match
	for name, threshold := range it.thresholds {
		if strings.ToLower(strings.TrimSpace(name)) == normalized {
			return it.clampThreshold(threshold)
		}
	}
	
	// Check partial match for industry names that might have variations
	// e.g., "Financial" matches "Financial Services"
	for name, threshold := range it.thresholds {
		if strings.Contains(normalized, strings.ToLower(name)) ||
		   strings.Contains(strings.ToLower(name), normalized) {
			return it.clampThreshold(threshold)
		}
	}
	
	// Return default threshold
	return it.clampThreshold(it.defaultThreshold)
}

// SetThreshold sets a custom threshold for an industry (thread-safe)
func (it *IndustryThresholds) SetThreshold(industryName string, threshold float64) {
	it.mu.Lock()
	defer it.mu.Unlock()
	it.thresholds[industryName] = it.clampThreshold(threshold)
}

// clampThreshold ensures threshold is within min/max bounds
func (it *IndustryThresholds) clampThreshold(threshold float64) float64 {
	if threshold < it.minThreshold {
		return it.minThreshold
	}
	if threshold > it.maxThreshold {
		return it.maxThreshold
	}
	return threshold
}

// ShouldTerminateEarly determines if classification should terminate early based on industry threshold
func (it *IndustryThresholds) ShouldTerminateEarly(industryName string, confidence float64, keywordCount int) bool {
	threshold := it.GetThreshold(industryName)
	minKeywords := 2
	
	// Terminate early if confidence is below threshold and keywords are insufficient
	return confidence < threshold && keywordCount < minKeywords
}

// ShouldGenerateCodes determines if code generation should be performed based on industry threshold
func (it *IndustryThresholds) ShouldGenerateCodes(industryName string, confidence float64) bool {
	threshold := it.GetThreshold(industryName)
	// Generate codes if confidence meets or exceeds threshold, or if confidence > 0.5 (fallback)
	return confidence >= threshold || confidence > 0.5
}

