package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/merchant-service/internal/config"
	"kyb-platform/services/merchant-service/internal/supabase"
)

// ClassificationJob represents a classification analysis job
type ClassificationJob struct {
	ID          string
	MerchantID  string
	BusinessName string
	Description  string
	WebsiteURL   string
	Status       JobStatus
	Result       *ClassificationResult
	Error        error
	CreatedAt    time.Time
	UpdatedAt    time.Time
	
	supabaseClient *supabase.Client
	config         *config.Config
	logger         *zap.Logger
	httpClient     *http.Client
}

// ClassificationResult represents the result of classification
type ClassificationResult struct {
	PrimaryIndustry string                 `json:"primaryIndustry"`
	ConfidenceScore float64                `json:"confidenceScore"`
	RiskLevel       string                 `json:"riskLevel"`
	MCCCodes        []IndustryCode         `json:"mccCodes,omitempty"`
	SICCodes        []IndustryCode         `json:"sicCodes,omitempty"`
	NAICSCodes      []IndustryCode          `json:"naicsCodes,omitempty"`
	Status          string                 `json:"status"`
	Explanation     string                 `json:"explanation,omitempty"`     // DistilBART explanation
	ContentSummary  string                 `json:"contentSummary,omitempty"`  // DistilBART content summary
	QuantizationEnabled bool               `json:"quantizationEnabled,omitempty"` // Quantization status
	ModelVersion     string                 `json:"modelVersion,omitempty"`    // Model version
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// IndustryCode represents an industry classification code
type IndustryCode struct {
	Code            string   `json:"code"`
	Description     string   `json:"description"`
	Confidence      float64  `json:"confidence"`
	Source          []string `json:"source,omitempty"`          // ["industry", "keyword", "both"]
	MatchType       string   `json:"matchType,omitempty"`       // "exact", "partial", "synonym"
	RelevanceScore  float64  `json:"relevanceScore,omitempty"`  // From code_keywords table
	Industries      []string `json:"industries,omitempty"`      // Industries that contributed this code
	IsPrimary       bool     `json:"isPrimary,omitempty"`      // From classification_codes.is_primary
}

// NewClassificationJob creates a new classification job
func NewClassificationJob(
	merchantID, businessName, description, websiteURL string,
	supabaseClient *supabase.Client,
	cfg *config.Config,
	logger *zap.Logger,
) *ClassificationJob {
	return &ClassificationJob{
		ID:             fmt.Sprintf("classification_%s_%d", merchantID, time.Now().Unix()),
		MerchantID:     merchantID,
		BusinessName:   businessName,
		Description:    description,
		WebsiteURL:     websiteURL,
		Status:         StatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		supabaseClient: supabaseClient,
		config:         cfg,
		logger:         logger,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetID returns the job ID
func (j *ClassificationJob) GetID() string {
	return j.ID
}

// GetMerchantID returns the merchant ID
func (j *ClassificationJob) GetMerchantID() string {
	return j.MerchantID
}

// GetType returns the job type
func (j *ClassificationJob) GetType() string {
	return "classification"
}

// GetStatus returns the current job status
func (j *ClassificationJob) GetStatus() JobStatus {
	return j.Status
}

// SetStatus sets the job status
func (j *ClassificationJob) SetStatus(status JobStatus) {
	j.Status = status
	j.UpdatedAt = time.Now()
}

// Process executes the classification job
func (j *ClassificationJob) Process(ctx context.Context) error {
	startTime := time.Now()
	j.logger.Info("Starting classification job",
		zap.String("job_id", j.ID),
		zap.String("merchant_id", j.MerchantID),
		zap.String("business_name", j.BusinessName))

	// Update status to processing
	j.SetStatus(StatusProcessing)
	if err := j.updateStatusInDB(ctx, StatusProcessing); err != nil {
		j.logger.Warn("Failed to update status to processing", zap.Error(err))
	}

	// Call classification service
	result, err := j.callClassificationService(ctx)
	if err != nil {
		j.logger.Error("Classification job failed",
			zap.String("job_id", j.ID),
			zap.String("merchant_id", j.MerchantID),
			zap.Error(err))
		
		j.SetStatus(StatusFailed)
		j.Error = err
		j.updateStatusInDB(ctx, StatusFailed)
		return fmt.Errorf("classification failed: %w", err)
	}

	// Save result to database
	if err := j.saveResultToDB(ctx, result); err != nil {
		j.logger.Error("Failed to save classification result",
			zap.String("job_id", j.ID),
			zap.String("merchant_id", j.MerchantID),
			zap.Error(err))
		
		j.SetStatus(StatusFailed)
		j.Error = err
		j.updateStatusInDB(ctx, StatusFailed)
		return fmt.Errorf("failed to save result: %w", err)
	}

	j.Result = result
	j.SetStatus(StatusCompleted)
	j.updateStatusInDB(ctx, StatusCompleted)

	duration := time.Since(startTime)
	j.logger.Info("Classification job completed successfully",
		zap.String("job_id", j.ID),
		zap.String("merchant_id", j.MerchantID),
		zap.Duration("duration", duration),
		zap.String("primary_industry", result.PrimaryIndustry),
		zap.Float64("confidence", result.ConfidenceScore))

	return nil
}

// callClassificationService calls the classification service API
// Falls back to basic classification if the service is unavailable
func (j *ClassificationJob) callClassificationService(ctx context.Context) (*ClassificationResult, error) {
	// Get classification service URL from config or environment
	classificationURL := j.getClassificationServiceURL()
	if classificationURL == "" {
		j.logger.Warn("Classification service URL not configured, using fallback classification",
			zap.String("merchant_id", j.MerchantID))
		return j.performFallbackClassification(), nil
	}

	// Prepare request
	reqBody := map[string]interface{}{
		"business_name": j.BusinessName,
		"description":   j.Description,
		"website_url":   j.WebsiteURL,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", classificationURL+"/v1/classify", bytes.NewBuffer(jsonData))
	if err != nil {
		j.logger.Warn("Failed to create classification request, using fallback",
			zap.String("merchant_id", j.MerchantID),
			zap.Error(err))
		return j.performFallbackClassification(), nil
	}

	req.Header.Set("Content-Type", "application/json")

	// Make request with timeout
	resp, err := j.httpClient.Do(req)
	if err != nil {
		j.logger.Warn("Failed to call classification service, using fallback",
			zap.String("merchant_id", j.MerchantID),
			zap.String("url", classificationURL),
			zap.Error(err))
		return j.performFallbackClassification(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		j.logger.Warn("Classification service returned non-OK status, using fallback",
			zap.String("merchant_id", j.MerchantID),
			zap.Int("status_code", resp.StatusCode))
		return j.performFallbackClassification(), nil
	}

	// Parse response
	var apiResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		j.logger.Warn("Failed to decode classification response, using fallback",
			zap.String("merchant_id", j.MerchantID),
			zap.Error(err))
		return j.performFallbackClassification(), nil
	}

	// Extract classification data from response
	result := j.extractClassificationFromResponse(apiResponse)
	return result, nil
}

// performFallbackClassification performs basic classification when the service is unavailable
func (j *ClassificationJob) performFallbackClassification() *ClassificationResult {
	j.logger.Info("Performing fallback classification",
		zap.String("merchant_id", j.MerchantID),
		zap.String("business_name", j.BusinessName))

	// Use industry from merchant data if available
	primaryIndustry := j.Description
	if primaryIndustry == "" {
		primaryIndustry = "General Business"
	}

	// Basic confidence score for fallback
	confidenceScore := 0.50

	// Determine risk level based on industry keywords
	riskLevel := "medium"
	lowerName := strings.ToLower(j.BusinessName)
	lowerDesc := strings.ToLower(j.Description)
	
	// Low risk indicators
	lowRiskKeywords := []string{"bank", "insurance", "government", "education", "healthcare", "hospital"}
	for _, keyword := range lowRiskKeywords {
		if strings.Contains(lowerName, keyword) || strings.Contains(lowerDesc, keyword) {
			riskLevel = "low"
			confidenceScore = 0.60
			break
		}
	}

	// High risk indicators
	highRiskKeywords := []string{"casino", "gambling", "adult", "cryptocurrency", "crypto"}
	for _, keyword := range highRiskKeywords {
		if strings.Contains(lowerName, keyword) || strings.Contains(lowerDesc, keyword) {
			riskLevel = "high"
			confidenceScore = 0.60
			break
		}
	}

	// Generate industry codes based on business name and description
	mccCodes, sicCodes, naicsCodes := j.generateFallbackIndustryCodes(lowerName, lowerDesc, primaryIndustry)

	result := &ClassificationResult{
		PrimaryIndustry: primaryIndustry,
		ConfidenceScore: confidenceScore,
		RiskLevel:       riskLevel,
		MCCCodes:        mccCodes,
		SICCodes:        sicCodes,
		NAICSCodes:      naicsCodes,
		Status:          "completed",
		Metadata: map[string]interface{}{
			"method":           "fallback",
			"service_unavailable": true,
			"fallback_reason":   "classification service not available",
		},
	}

	return result
}

// generateFallbackIndustryCodes generates basic industry codes based on business name and description
func (j *ClassificationJob) generateFallbackIndustryCodes(businessName, description, industry string) ([]IndustryCode, []IndustryCode, []IndustryCode) {
	var mccCodes []IndustryCode
	var sicCodes []IndustryCode
	var naicsCodes []IndustryCode

	// Combine text for analysis
	combinedText := strings.ToLower(businessName + " " + description + " " + industry)

	// Retail/Store businesses
	if strings.Contains(combinedText, "retail") || strings.Contains(combinedText, "store") || 
	   strings.Contains(combinedText, "shop") || strings.Contains(combinedText, "merchant") ||
	   strings.Contains(combinedText, "outdoor") || strings.Contains(combinedText, "equipment") ||
	   strings.Contains(combinedText, "sporting") || strings.Contains(combinedText, "goods") {
		mccCodes = append(mccCodes, IndustryCode{
			Code:        "5941",
			Description: "Sporting Goods Stores",
			Confidence:  0.65,
		})
		sicCodes = append(sicCodes, IndustryCode{
			Code:        "5941",
			Description: "Sporting Goods Stores",
			Confidence:  0.65,
		})
		naicsCodes = append(naicsCodes, IndustryCode{
			Code:        "451110",
			Description: "Sporting Goods Stores",
			Confidence:  0.65,
		})
	}

	// Wine/Spirits businesses
	if strings.Contains(combinedText, "wine") || strings.Contains(combinedText, "grape") ||
	   strings.Contains(combinedText, "liquor") || strings.Contains(combinedText, "spirits") ||
	   strings.Contains(combinedText, "beverage") || strings.Contains(combinedText, "alcohol") {
		mccCodes = append(mccCodes, IndustryCode{
			Code:        "5921",
			Description: "Package Stores - Beer, Wine, and Liquor",
			Confidence:  0.75,
		})
		sicCodes = append(sicCodes, IndustryCode{
			Code:        "5921",
			Description: "Liquor Stores",
			Confidence:  0.75,
		})
		naicsCodes = append(naicsCodes, IndustryCode{
			Code:        "445310",
			Description: "Beer, Wine, and Liquor Stores",
			Confidence:  0.75,
		})
	}

	// Technology/Software businesses
	if strings.Contains(combinedText, "tech") || strings.Contains(combinedText, "software") ||
	   strings.Contains(combinedText, "computer") || strings.Contains(combinedText, "digital") {
		mccCodes = append(mccCodes, IndustryCode{
			Code:        "5734",
			Description: "Computer Software Stores",
			Confidence:  0.70,
		})
		sicCodes = append(sicCodes, IndustryCode{
			Code:        "7372",
			Description: "Prepackaged Software",
			Confidence:  0.70,
		})
		naicsCodes = append(naicsCodes, IndustryCode{
			Code:        "541511",
			Description: "Custom Computer Programming Services",
			Confidence:  0.70,
		})
	}

	// Food/Restaurant businesses
	if strings.Contains(combinedText, "restaurant") || strings.Contains(combinedText, "food") ||
	   strings.Contains(combinedText, "cafe") || strings.Contains(combinedText, "dining") {
		mccCodes = append(mccCodes, IndustryCode{
			Code:        "5812",
			Description: "Eating Places, Restaurants",
			Confidence:  0.70,
		})
		sicCodes = append(sicCodes, IndustryCode{
			Code:        "5812",
			Description: "Eating Places",
			Confidence:  0.70,
		})
		naicsCodes = append(naicsCodes, IndustryCode{
			Code:        "722511",
			Description: "Full-Service Restaurants",
			Confidence:  0.70,
		})
	}

	// General business fallback (if no specific matches)
	if len(mccCodes) == 0 {
		mccCodes = append(mccCodes, IndustryCode{
			Code:        "5999",
			Description: "Miscellaneous and Specialty Retail Stores",
			Confidence:  0.50,
		})
		sicCodes = append(sicCodes, IndustryCode{
			Code:        "5999",
			Description: "Miscellaneous Retail Stores, Not Elsewhere Classified",
			Confidence:  0.50,
		})
		naicsCodes = append(naicsCodes, IndustryCode{
			Code:        "453998",
			Description: "All Other Miscellaneous Store Retailers",
			Confidence:  0.50,
		})
	}

	return mccCodes, sicCodes, naicsCodes
}

// extractClassificationFromResponse extracts classification data from API response
func (j *ClassificationJob) extractClassificationFromResponse(response map[string]interface{}) *ClassificationResult {
	result := &ClassificationResult{
		Status:   "completed",
		Metadata: make(map[string]interface{}),
	}

	// Extract primary industry (check multiple possible locations)
	if industry, ok := response["primary_industry"].(string); ok {
		result.PrimaryIndustry = industry
	} else if industry, ok := response["industry"].(string); ok {
		result.PrimaryIndustry = industry
	} else if classification, ok := response["classification"].(map[string]interface{}); ok {
		// Check nested classification.industry field
		if industry, ok := classification["industry"].(string); ok {
			result.PrimaryIndustry = industry
		}
	} else if enhanced, ok := response["enhanced_classification"].(map[string]interface{}); ok {
		if industry, ok := enhanced["primary_industry"].(string); ok {
			result.PrimaryIndustry = industry
		}
	}

	// Extract confidence score
	if confidence, ok := response["confidence_score"].(float64); ok {
		result.ConfidenceScore = confidence
	} else if confidence, ok := response["confidence"].(float64); ok {
		result.ConfidenceScore = confidence
	}

	// Extract industry codes
	result.MCCCodes = j.extractIndustryCodes(response, "mcc_codes", "mccCodes")
	result.SICCodes = j.extractIndustryCodes(response, "sic_codes", "sicCodes")
	result.NAICSCodes = j.extractIndustryCodes(response, "naics_codes", "naicsCodes")

	// Extract risk level (default to medium if not found)
	if riskLevel, ok := response["risk_level"].(string); ok {
		result.RiskLevel = riskLevel
	} else {
		result.RiskLevel = "medium"
	}

	// Extract metadata from response
	if metadata, ok := response["metadata"].(map[string]interface{}); ok {
		// Extract page analysis metadata
		if pageAnalysis, ok := metadata["pageAnalysis"].(map[string]interface{}); ok {
			result.Metadata["pageAnalysis"] = pageAnalysis
		}
		
		// Extract brand match metadata
		if brandMatch, ok := metadata["brandMatch"].(map[string]interface{}); ok {
			result.Metadata["brandMatch"] = brandMatch
		}
		
		// Extract analysis method
		if analysisMethod, ok := metadata["analysisMethod"].(string); ok {
			result.Metadata["analysisMethod"] = analysisMethod
		}
		
		// Extract DistilBART enhancement fields
		if explanation, ok := metadata["explanation"].(string); ok {
			result.Explanation = explanation
		}
		if contentSummary, ok := metadata["content_summary"].(string); ok {
			result.ContentSummary = contentSummary
		}
		if quantizationEnabled, ok := metadata["quantization_enabled"].(bool); ok {
			result.QuantizationEnabled = quantizationEnabled
		}
		if modelVersion, ok := metadata["model_version"].(string); ok {
			result.ModelVersion = modelVersion
		}
	}

	// Add data source priority metadata (based on classification priority)
	result.Metadata["dataSourcePriority"] = map[string]string{
		"websiteContent": "primary",
		"businessName":    "secondary", // Only used for brand matches
		"websiteURL":     "fallback",
	}

	// Store full response in metadata for debugging
	result.Metadata["raw_response"] = response

	return result
}

// extractIndustryCodes extracts industry codes from response
func (j *ClassificationJob) extractIndustryCodes(response map[string]interface{}, key1, key2 string) []IndustryCode {
	var codes []IndustryCode

	// Try first key
	if codesData, ok := response[key1].([]interface{}); ok {
		codes = j.parseIndustryCodesArray(codesData)
	} else if codesData, ok := response[key2].([]interface{}); ok {
		codes = j.parseIndustryCodesArray(codesData)
	} else if enhanced, ok := response["enhanced_classification"].(map[string]interface{}); ok {
		if codesData, ok := enhanced[key1].([]interface{}); ok {
			codes = j.parseIndustryCodesArray(codesData)
		} else if codesData, ok := enhanced[key2].([]interface{}); ok {
			codes = j.parseIndustryCodesArray(codesData)
		}
	}

	return codes
}

// parseIndustryCodesArray parses an array of industry codes
func (j *ClassificationJob) parseIndustryCodesArray(data []interface{}) []IndustryCode {
	codes := make([]IndustryCode, 0, len(data))
	
	for _, item := range data {
		var code IndustryCode
		
		if codeMap, ok := item.(map[string]interface{}); ok {
			if codeStr, ok := codeMap["code"].(string); ok {
				code.Code = codeStr
			}
			if desc, ok := codeMap["description"].(string); ok {
				code.Description = desc
			}
			if conf, ok := codeMap["confidence"].(float64); ok {
				code.Confidence = conf
			} else if conf, ok := codeMap["confidence_score"].(float64); ok {
				code.Confidence = conf
			}
			
			// Extract new optional fields
			if source, ok := codeMap["source"].([]interface{}); ok {
				code.Source = make([]string, 0, len(source))
				for _, s := range source {
					if str, ok := s.(string); ok {
						code.Source = append(code.Source, str)
					}
				}
			}
			if matchType, ok := codeMap["matchType"].(string); ok {
				code.MatchType = matchType
			}
			if relevanceScore, ok := codeMap["relevanceScore"].(float64); ok {
				code.RelevanceScore = relevanceScore
			}
			if industries, ok := codeMap["industries"].([]interface{}); ok {
				code.Industries = make([]string, 0, len(industries))
				for _, ind := range industries {
					if str, ok := ind.(string); ok {
						code.Industries = append(code.Industries, str)
					}
				}
			}
			if isPrimary, ok := codeMap["isPrimary"].(bool); ok {
				code.IsPrimary = isPrimary
			}
			
			if code.Code != "" {
				codes = append(codes, code)
			}
		}
	}
	
	return codes
}

// getClassificationServiceURL gets the classification service URL from config or environment
func (j *ClassificationJob) getClassificationServiceURL() string {
	// Check environment variable first
	if url := os.Getenv("CLASSIFICATION_SERVICE_URL"); url != "" {
		return url
	}
	
	// Default to localhost for development
	if j.config.Environment == "development" {
		return "http://localhost:8081"
	}
	
	// Production default
	return "https://classification-service-production.up.railway.app"
}

// updateStatusInDB updates the job status in the database
func (j *ClassificationJob) updateStatusInDB(ctx context.Context, status JobStatus) error {
	updateData := map[string]interface{}{
		"classification_status":     string(status),
		"classification_updated_at": time.Now().Format(time.RFC3339),
	}

	// Check if merchant_analytics record exists
	var existing []map[string]interface{}
	_, err := j.supabaseClient.GetClient().From("merchant_analytics").
		Select("id", "", false).
		Eq("merchant_id", j.MerchantID).
		Limit(1, "").
		ExecuteTo(&existing)

	if err != nil || len(existing) == 0 {
		// Create new record
		insertData := map[string]interface{}{
			"merchant_id":            j.MerchantID,
			"classification_status":   string(status),
			"classification_updated_at": time.Now().Format(time.RFC3339),
			"classification_data":     map[string]interface{}{},
		}
		
		_, _, err := j.supabaseClient.GetClient().From("merchant_analytics").
			Insert(insertData, false, "", "", "").
			Execute()
		
		return err
	}

	// Update existing record
	_, _, err = j.supabaseClient.GetClient().From("merchant_analytics").
		Update(updateData, "", "").
		Eq("merchant_id", j.MerchantID).
		Execute()

	return err
}

// saveResultToDB saves the classification result to the database
func (j *ClassificationJob) saveResultToDB(ctx context.Context, result *ClassificationResult) error {
	// Convert result to JSONB format
	classificationData := map[string]interface{}{
		"primaryIndustry": result.PrimaryIndustry,
		"confidenceScore": result.ConfidenceScore,
		"riskLevel":       result.RiskLevel,
		"status":          result.Status,
	}

	// Add DistilBART enhancement fields
	if result.Explanation != "" {
		classificationData["explanation"] = result.Explanation
	}
	if result.ContentSummary != "" {
		classificationData["contentSummary"] = result.ContentSummary
	}
	if result.QuantizationEnabled {
		classificationData["quantizationEnabled"] = result.QuantizationEnabled
	}
	if result.ModelVersion != "" {
		classificationData["modelVersion"] = result.ModelVersion
	}

	// Add industry codes
	if len(result.MCCCodes) > 0 {
		classificationData["mccCodes"] = result.MCCCodes
	}
	if len(result.SICCodes) > 0 {
		classificationData["sicCodes"] = result.SICCodes
	}
	if len(result.NAICSCodes) > 0 {
		classificationData["naicsCodes"] = result.NAICSCodes
	}

	// Add metadata
	if len(result.Metadata) > 0 {
		classificationData["metadata"] = result.Metadata
	}

	updateData := map[string]interface{}{
		"classification_data":      classificationData,
		"classification_status":    "completed",
		"classification_updated_at": time.Now().Format(time.RFC3339),
	}

	// Check if record exists
	var existing []map[string]interface{}
	_, err := j.supabaseClient.GetClient().From("merchant_analytics").
		Select("id", "", false).
		Eq("merchant_id", j.MerchantID).
		Limit(1, "").
		ExecuteTo(&existing)

	if err != nil || len(existing) == 0 {
		// Create new record
		insertData := map[string]interface{}{
			"merchant_id":            j.MerchantID,
			"classification_data":    classificationData,
			"classification_status":  "completed",
			"classification_updated_at": time.Now().Format(time.RFC3339),
		}
		
		_, _, err := j.supabaseClient.GetClient().From("merchant_analytics").
			Insert(insertData, false, "", "", "").
			Execute()
		
		return err
	}

	// Update existing record
	_, _, err = j.supabaseClient.GetClient().From("merchant_analytics").
		Update(updateData, "", "").
		Eq("merchant_id", j.MerchantID).
		Execute()

	return err
}

