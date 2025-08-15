package webanalysis

import (
	"fmt"
	"net/url"
	"strings"
)

// FlowSelector determines which classification flow to use
type FlowSelector struct {
	config FlowSelectorConfig
}

// FlowSelectorConfig holds configuration for flow selection
type FlowSelectorConfig struct {
	DefaultFlow         ClassificationFlow
	URLFlowPriority     float64
	SearchFlowPriority  float64
	EnableAutoSelection bool
}

// NewFlowSelector creates a new flow selector
func NewFlowSelector() *FlowSelector {
	config := FlowSelectorConfig{
		DefaultFlow:         FlowURLBased,
		URLFlowPriority:     0.8,
		SearchFlowPriority:  0.6,
		EnableAutoSelection: true,
	}

	return &FlowSelector{
		config: config,
	}
}

// SelectFlow determines which flow to use based on the request
func (fs *FlowSelector) SelectFlow(req *ClassificationRequest) (ClassificationFlow, error) {
	// If user has specified a flow preference, use it
	if req.FlowPreference != "" {
		return fs.validateFlowPreference(req.FlowPreference)
	}

	// If auto-selection is disabled, use default flow
	if !fs.config.EnableAutoSelection {
		return fs.config.DefaultFlow, nil
	}

	// Auto-select flow based on request parameters
	return fs.autoSelectFlow(req)
}

// validateFlowPreference validates the user's flow preference
func (fs *FlowSelector) validateFlowPreference(preference ClassificationFlow) (ClassificationFlow, error) {
	switch preference {
	case FlowURLBased, FlowSearchBased:
		return preference, nil
	default:
		return fs.config.DefaultFlow, fmt.Errorf("invalid flow preference: %s", preference)
	}
}

// autoSelectFlow automatically selects the best flow based on request parameters
func (fs *FlowSelector) autoSelectFlow(req *ClassificationRequest) (ClassificationFlow, error) {
	// Calculate scores for each flow
	urlScore := fs.calculateURLFlowScore(req)
	searchScore := fs.calculateSearchFlowScore(req)

	// Select flow with higher score
	if urlScore > searchScore {
		return FlowURLBased, nil
	} else {
		return FlowSearchBased, nil
	}
}

// calculateURLFlowScore calculates the score for URL-based flow
func (fs *FlowSelector) calculateURLFlowScore(req *ClassificationRequest) float64 {
	score := 0.0

	// Check if website URL is provided
	if req.WebsiteURL != "" {
		score += 0.8

		// Validate URL format
		if fs.isValidURL(req.WebsiteURL) {
			score += 0.2
		}
	}

	// Check if business name is provided (required for both flows)
	if req.BusinessName != "" {
		score += 0.1
	}

	// Check if additional business information is provided
	if req.BusinessType != "" {
		score += 0.05
	}
	if req.Industry != "" {
		score += 0.05
	}
	if req.Address != "" {
		score += 0.05
	}
	if req.ContactInfo != nil && len(req.ContactInfo) > 0 {
		score += 0.05
	}

	// Apply priority multiplier
	score *= fs.config.URLFlowPriority

	return score
}

// calculateSearchFlowScore calculates the score for search-based flow
func (fs *FlowSelector) calculateSearchFlowScore(req *ClassificationRequest) float64 {
	score := 0.0

	// Check if business name is provided (required for search flow)
	if req.BusinessName != "" {
		score += 0.6
	} else {
		// Business name is required for search flow
		return 0.0
	}

	// Check if additional business information is provided
	if req.BusinessType != "" {
		score += 0.1
	}
	if req.Industry != "" {
		score += 0.1
	}
	if req.Address != "" {
		score += 0.1
	}
	if req.ContactInfo != nil && len(req.ContactInfo) > 0 {
		score += 0.1
	}

	// If no website URL is provided, search flow is preferred
	if req.WebsiteURL == "" {
		score += 0.3
	}

	// Apply priority multiplier
	score *= fs.config.SearchFlowPriority

	return score
}

// isValidURL checks if a URL is valid
func (fs *FlowSelector) isValidURL(urlString string) bool {
	// Basic URL validation
	if !strings.HasPrefix(urlString, "http://") && !strings.HasPrefix(urlString, "https://") {
		return false
	}

	// Try to parse the URL
	_, err := url.Parse(urlString)
	return err == nil
}

// GetFlowRecommendation provides a recommendation for which flow to use
func (fs *FlowSelector) GetFlowRecommendation(req *ClassificationRequest) (ClassificationFlow, string, error) {
	urlScore := fs.calculateURLFlowScore(req)
	searchScore := fs.calculateSearchFlowScore(req)

	var recommendedFlow ClassificationFlow
	var reason string

	if urlScore > searchScore {
		recommendedFlow = FlowURLBased
		if req.WebsiteURL != "" {
			reason = "Website URL provided - URL-based flow will provide more accurate results"
		} else {
			reason = "URL-based flow preferred based on available information"
		}
	} else {
		recommendedFlow = FlowSearchBased
		if req.WebsiteURL == "" {
			reason = "No website URL provided - search-based flow will find relevant information"
		} else {
			reason = "Search-based flow preferred based on available information"
		}
	}

	return recommendedFlow, reason, nil
}

// GetFlowComparison provides a comparison of both flows for the request
func (fs *FlowSelector) GetFlowComparison(req *ClassificationRequest) map[string]interface{} {
	urlScore := fs.calculateURLFlowScore(req)
	searchScore := fs.calculateSearchFlowScore(req)

	urlRecommendation, urlReason, _ := fs.GetFlowRecommendation(req)

	return map[string]interface{}{
		"url_flow_score":         urlScore,
		"search_flow_score":      searchScore,
		"recommended_flow":       urlRecommendation,
		"recommendation_reason":  urlReason,
		"url_flow_available":     req.WebsiteURL != "",
		"search_flow_available":  req.BusinessName != "",
		"auto_selection_enabled": fs.config.EnableAutoSelection,
	}
}

// GetStats returns statistics about flow selection
func (fs *FlowSelector) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_selections":       0, // TODO: Implement counter
		"url_flow_selections":    0, // TODO: Implement counter
		"search_flow_selections": 0, // TODO: Implement counter
		"auto_selections":        0, // TODO: Implement counter
		"manual_selections":      0, // TODO: Implement counter
	}
}
