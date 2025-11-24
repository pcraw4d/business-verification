package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// IndustryCode represents an industry classification code
type IndustryCode struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
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
func (j *ClassificationJob) callClassificationService(ctx context.Context) (*ClassificationResult, error) {
	// Get classification service URL from config or environment
	classificationURL := j.getClassificationServiceURL()
	if classificationURL == "" {
		return nil, fmt.Errorf("classification service URL not configured")
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
	req, err := http.NewRequestWithContext(ctx, "POST", classificationURL+"/api/v1/classify", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call classification service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("classification service returned status %d", resp.StatusCode)
	}

	// Parse response
	var apiResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract classification data from response
	result := j.extractClassificationFromResponse(apiResponse)
	return result, nil
}

// extractClassificationFromResponse extracts classification data from API response
func (j *ClassificationJob) extractClassificationFromResponse(response map[string]interface{}) *ClassificationResult {
	result := &ClassificationResult{
		Status:   "completed",
		Metadata: make(map[string]interface{}),
	}

	// Extract primary industry
	if industry, ok := response["primary_industry"].(string); ok {
		result.PrimaryIndustry = industry
	} else if industry, ok := response["industry"].(string); ok {
		result.PrimaryIndustry = industry
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

	// Store full response in metadata
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

