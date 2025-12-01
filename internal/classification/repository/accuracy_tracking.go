package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// SaveClassificationAccuracy saves a classification accuracy tracking record to the database
// OPTIMIZATION #5.2: Confidence Calibration - Database Persistence
func (r *SupabaseKeywordRepository) SaveClassificationAccuracy(ctx context.Context, tracking *ClassificationAccuracyTracking) error {
	if r.client == nil {
		return fmt.Errorf("database client is nil")
	}

	// Get PostgREST client
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return fmt.Errorf("PostgREST client is nil")
	}

	// Convert to map for insertion
	data := map[string]interface{}{
		"request_id":             tracking.RequestID,
		"business_name":          tracking.BusinessName,
		"website_url":            tracking.WebsiteURL,
		"predicted_industry":     tracking.PredictedIndustry,
		"predicted_confidence":   tracking.PredictedConfidence,
		"confidence_bin":         tracking.ConfidenceBin,
		"classification_method":  tracking.ClassificationMethod,
		"keywords_count":         tracking.KeywordsCount,
		"processing_time_ms":    tracking.ProcessingTimeMs,
		"created_at":            tracking.CreatedAt.Format(time.RFC3339),
	}

	// Only include actual_industry if provided (validation)
	if tracking.ActualIndustry != "" {
		data["actual_industry"] = tracking.ActualIndustry
	}

	// Only include actual_confidence if provided
	if tracking.ActualConfidence > 0 {
		data["actual_confidence"] = tracking.ActualConfidence
	}

	// Only include is_correct if provided
	if tracking.IsCorrect != nil {
		data["is_correct"] = *tracking.IsCorrect
	}

	// Only include validated_at if provided
	if tracking.ValidatedAt != nil {
		data["validated_at"] = tracking.ValidatedAt.Format(time.RFC3339)
	}

	// Only include validated_by if provided
	if tracking.ValidatedBy != "" {
		data["validated_by"] = tracking.ValidatedBy
	}

	// Insert into database using PostgREST
	// PostgREST expects the data as interface{} (map or slice), not JSON bytes
	// Wrap data in array for PostgREST Insert
	insertData := []map[string]interface{}{data}
	_, _, err := postgrestClient.From("classification_accuracy_tracking").
		Insert(insertData, false, "", "", "").
		Execute()

	if err != nil {
		r.logger.Printf("⚠️ [Accuracy Tracking] Failed to save classification accuracy: %v", err)
		return fmt.Errorf("failed to save classification accuracy: %w", err)
	}

	r.logger.Printf("✅ [Accuracy Tracking] Saved classification accuracy for request_id: %s", tracking.RequestID)
	return nil
}

// UpdateClassificationAccuracy updates the actual industry and validation status for a classification
// OPTIMIZATION #5.2: Confidence Calibration - Validation API
func (r *SupabaseKeywordRepository) UpdateClassificationAccuracy(
	ctx context.Context,
	requestID string,
	actualIndustry string,
	validatedBy string,
) error {
	if r.client == nil {
		return fmt.Errorf("database client is nil")
	}

	// Get PostgREST client
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return fmt.Errorf("PostgREST client is nil")
	}

	// Prepare update data
	updateData := map[string]interface{}{
		"actual_industry": actualIndustry,
		"is_correct":      nil, // Will be calculated based on predicted vs actual
		"validated_at":    time.Now().Format(time.RFC3339),
		"validated_by":    validatedBy,
	}

	// Get the existing record to check if predicted matches actual
	var existingRecord struct {
		PredictedIndustry string `json:"predicted_industry"`
	}

	// First, get the existing record
	response, _, err := postgrestClient.From("classification_accuracy_tracking").
		Select("predicted_industry", "", false).
		Eq("request_id", requestID).
		Single().
		Execute()

	if err != nil {
		return fmt.Errorf("failed to find classification record: %w", err)
	}

	if err := json.Unmarshal(response, &existingRecord); err != nil {
		return fmt.Errorf("failed to unmarshal existing record: %w", err)
	}

	// Calculate is_correct
	isCorrect := existingRecord.PredictedIndustry == actualIndustry
	updateData["is_correct"] = isCorrect

	// Marshal update data
	jsonData, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("failed to marshal update data: %w", err)
	}

	// Update the record
	_, _, err = postgrestClient.From("classification_accuracy_tracking").
		Update(jsonData, "", "").
		Eq("request_id", requestID).
		Execute()

	if err != nil {
		r.logger.Printf("⚠️ [Accuracy Tracking] Failed to update classification accuracy: %v", err)
		return fmt.Errorf("failed to update classification accuracy: %w", err)
	}

	r.logger.Printf("✅ [Accuracy Tracking] Updated classification accuracy for request_id: %s (is_correct: %v)",
		requestID, isCorrect)
	return nil
}

// GetCalibrationStatistics retrieves calibration statistics for a date range
// OPTIMIZATION #5.2: Confidence Calibration - Statistics
func (r *SupabaseKeywordRepository) GetCalibrationStatistics(
	ctx context.Context,
	startDate time.Time,
	endDate time.Time,
) ([]*CalibrationBinStatistics, error) {
	if r.client == nil {
		return nil, fmt.Errorf("database client is nil")
	}

	// Get PostgREST client
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return nil, fmt.Errorf("PostgREST client is nil")
	}

	// Use the database function to get statistics
	// Note: We'll use a direct query since PostgREST doesn't easily support function calls
	// For now, we'll query the view instead
	response, _, err := postgrestClient.From("classification_accuracy_by_bin").
		Select("*", "", false).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to get calibration statistics: %w", err)
	}

	var results []struct {
		ConfidenceBin         int     `json:"confidence_bin"`
		TotalClassifications  int64   `json:"total_classifications"`
		CorrectClassifications int64  `json:"correct_classifications"`
		AvgPredictedConfidence float64 `json:"avg_predicted_confidence"`
		ActualAccuracy        float64 `json:"actual_accuracy"`
		CalibrationError      float64 `json:"calibration_error"`
	}

	if err := json.Unmarshal(response, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal calibration statistics: %w", err)
	}

	// Convert to CalibrationBinStatistics
	statistics := make([]*CalibrationBinStatistics, 0, len(results))
	for _, result := range results {
		statistics = append(statistics, &CalibrationBinStatistics{
			ConfidenceBin:         result.ConfidenceBin,
			TotalClassifications:  result.TotalClassifications,
			CorrectClassifications: result.CorrectClassifications,
			PredictedAccuracy:     result.AvgPredictedConfidence,
			ActualAccuracy:        result.ActualAccuracy,
			CalibrationError:      result.CalibrationError,
		})
	}

	return statistics, nil
}

