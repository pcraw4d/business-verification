package classification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"kyb-platform/internal/external"
)

// LLMClassifier handles LLM-based classification using Qwen 2.5 7B
type LLMClassifier struct {
	llmServiceURL string
	httpClient    *http.Client
	logger        *log.Logger
}

// NewLLMClassifier creates a new LLM classifier
func NewLLMClassifier(llmServiceURL string, logger *log.Logger) *LLMClassifier {
	if logger == nil {
		logger = log.Default()
	}

	// Trim trailing slash to prevent double-slash in URL construction
	llmServiceURL = strings.TrimSuffix(llmServiceURL, "/")

	return &LLMClassifier{
		llmServiceURL: llmServiceURL,
		httpClient: &http.Client{
			Timeout: 300 * time.Second, // LLM on CPU can take 200+ seconds
		},
		logger: logger,
	}
}

// LLMClassificationResult represents the result of LLM-based classification
type LLMClassificationResult struct {
	PrimaryIndustry          string
	Confidence               float64
	Reasoning                string
	MCC                      []CodeResult
	SIC                      []CodeResult
	NAICS                    []CodeResult
	AlternativeClassifications []string
	ProcessingTimeMs         int64
}

// ClassifyWithLLM performs classification using LLM reasoning
func (l *LLMClassifier) ClassifyWithLLM(
	ctx context.Context,
	content *external.ScrapedContent,
	businessName string,
	description string,
	layer1Result *MultiStrategyResult,
	layer2Result *EmbeddingClassificationResult,
) (*LLMClassificationResult, error) {
	startTime := time.Now()

	l.logger.Printf("🤖 [Layer 3] Starting LLM-based classification for business: %s", businessName)

	// Prepare website content (truncate for context limits)
	websiteContent := l.prepareWebsiteContent(content)

	// Build request payload
	reqBody := map[string]interface{}{
		"context": map[string]interface{}{
			"business_name":   businessName,
			"description":     description,
			"website_content": websiteContent,
		},
		"temperature": 0.1, // Low temperature for consistency
		"max_tokens":  800,
	}

	// Add Layer 1 context if available
	if layer1Result != nil {
		reqBody["context"].(map[string]interface{})["layer1_result"] = map[string]interface{}{
			"industry":   layer1Result.PrimaryIndustry,
			"confidence": layer1Result.Confidence,
			"keywords":   layer1Result.Keywords,
		}
	}

	// Add Layer 2 context if available
	if layer2Result != nil {
		reqBody["context"].(map[string]interface{})["layer2_result"] = map[string]interface{}{
			"top_match":      layer2Result.TopMatch,
			"confidence":     layer2Result.Confidence,
			"top_similarity": layer2Result.TopSimilarity,
		}
	}

	// Call LLM service
	response, err := l.callLLMService(ctx, reqBody)
	if err != nil {
		return nil, fmt.Errorf("LLM service call failed: %w", err)
	}

	// Parse response
	result := &LLMClassificationResult{
		PrimaryIndustry:            l.getString(response, "primary_industry"),
		Confidence:                 l.getFloat64(response, "confidence"),
		Reasoning:                  l.getString(response, "reasoning"),
		AlternativeClassifications: l.parseAlternatives(response["alternative_classifications"]),
		ProcessingTimeMs:           time.Since(startTime).Milliseconds(),
	}

	// Parse codes
	if codesInterface, ok := response["codes"].(map[string]interface{}); ok {
		result.MCC = l.parseCodes(codesInterface["mcc"])
		result.SIC = l.parseCodes(codesInterface["sic"])
		result.NAICS = l.parseCodes(codesInterface["naics"])
	}

	l.logger.Printf("✅ [Layer 3] LLM classification complete (industry: %s, confidence: %.2f%%, duration: %dms)",
		result.PrimaryIndustry, result.Confidence*100, result.ProcessingTimeMs)

	return result, nil
}

// prepareWebsiteContent prepares website content for LLM (truncate to 2000 chars)
func (l *LLMClassifier) prepareWebsiteContent(content *external.ScrapedContent) string {
	parts := []string{}

	// Title
	if content.Title != "" {
		parts = append(parts, fmt.Sprintf("Title: %s", content.Title))
	}

	// Meta description
	if content.MetaDesc != "" {
		parts = append(parts, fmt.Sprintf("Description: %s", content.MetaDesc))
	}

	// About section (truncated)
	if content.AboutText != "" {
		about := content.AboutText
		if len(about) > 800 {
			about = about[:800] + "..."
		}
		parts = append(parts, fmt.Sprintf("About: %s", about))
	}

	// Top headings
	if len(content.Headings) > 0 {
		headings := content.Headings
		if len(headings) > 5 {
			headings = headings[:5]
		}
		parts = append(parts, fmt.Sprintf("Headings: %s", strings.Join(headings, ", ")))
	}

	combined := strings.Join(parts, "\n")

	// Truncate to 2000 chars total
	if len(combined) > 2000 {
		combined = combined[:2000] + "..."
	}

	return combined
}

// callLLMService calls the LLM service and returns the response
func (l *LLMClassifier) callLLMService(
	ctx context.Context,
	reqBody map[string]interface{},
) (map[string]interface{}, error) {
	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		l.llmServiceURL+"/classify",
		bytes.NewReader(reqBodyJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	l.logger.Printf("📡 [Layer 3] Calling LLM service: %s", l.llmServiceURL)

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("LLM service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM service returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if processingTime, ok := result["processing_time_ms"].(float64); ok {
		l.logger.Printf("📊 [Layer 3] LLM service response (processing_time: %.0fms)", processingTime)
	}

	return result, nil
}

// parseCodes converts JSON codes to []CodeResult
func (l *LLMClassifier) parseCodes(codesInterface interface{}) []CodeResult {
	if codesInterface == nil {
		return []CodeResult{}
	}

	codesList, ok := codesInterface.([]interface{})
	if !ok {
		return []CodeResult{}
	}

	results := make([]CodeResult, 0, len(codesList))
	for _, item := range codesList {
		codeMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		code := l.getString(codeMap, "code")
		description := l.getString(codeMap, "description")
		confidence := l.getFloat64(codeMap, "confidence")

		results = append(results, CodeResult{
			Code:        code,
			Description: description,
			Confidence:  confidence,
			Source:      "llm_reasoning",
		})
	}

	return results
}

// parseAlternatives extracts alternative classifications
func (l *LLMClassifier) parseAlternatives(altInterface interface{}) []string {
	if altInterface == nil {
		return []string{}
	}

	altList, ok := altInterface.([]interface{})
	if !ok {
		return []string{}
	}

	results := make([]string, 0, len(altList))
	for _, item := range altList {
		if str, ok := item.(string); ok {
			results = append(results, str)
		}
	}

	return results
}

// Helper functions for safe type assertions
func (l *LLMClassifier) getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (l *LLMClassifier) getFloat64(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0.0
}

