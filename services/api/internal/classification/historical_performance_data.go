package classification

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// NewHistoricalPerformanceData creates a new historical performance data manager
func NewHistoricalPerformanceData(dataFile string, logger *log.Logger) *HistoricalPerformanceData {
	if logger == nil {
		logger = log.Default()
	}

	hpd := &HistoricalPerformanceData{
		DataFile: dataFile,
		Data:     make(map[string]*MethodPerformanceData),
	}

	// Load existing data
	if err := hpd.LoadData(); err != nil {
		logger.Printf("⚠️ Failed to load historical performance data: %v", err)
	}

	return hpd
}

// LoadData loads historical performance data from file
func (hpd *HistoricalPerformanceData) LoadData() error {
	hpd.mutex.Lock()
	defer hpd.mutex.Unlock()

	// Check if file exists
	if _, err := os.Stat(hpd.DataFile); os.IsNotExist(err) {
		// File doesn't exist, create directory and return
		dir := filepath.Dir(hpd.DataFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create data directory: %w", err)
		}
		return nil
	}

	// Read file
	data, err := os.ReadFile(hpd.DataFile)
	if err != nil {
		return fmt.Errorf("failed to read data file: %w", err)
	}

	// Unmarshal data
	if err := json.Unmarshal(data, &hpd.Data); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// SaveData saves historical performance data to file
func (hpd *HistoricalPerformanceData) SaveData() error {
	hpd.mutex.Lock()
	defer hpd.mutex.Unlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(hpd.DataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Marshal data
	data, err := json.MarshalIndent(hpd.Data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Write file
	if err := os.WriteFile(hpd.DataFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write data file: %w", err)
	}

	return nil
}

// SaveWeightAdjustments saves weight adjustment history
func (hpd *HistoricalPerformanceData) SaveWeightAdjustments(adjustments map[string]WeightAdjustment) error {
	hpd.mutex.Lock()
	defer hpd.mutex.Unlock()

	// Create weight adjustments file
	adjustmentsFile := "data/weight_adjustments.json"

	// Load existing adjustments
	var existingAdjustments []WeightAdjustment
	if data, err := os.ReadFile(adjustmentsFile); err == nil {
		json.Unmarshal(data, &existingAdjustments)
	}

	// Add new adjustments
	for _, adjustment := range adjustments {
		existingAdjustments = append(existingAdjustments, adjustment)
	}

	// Keep only last 1000 adjustments
	if len(existingAdjustments) > 1000 {
		existingAdjustments = existingAdjustments[len(existingAdjustments)-1000:]
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(adjustmentsFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Save adjustments
	data, err := json.MarshalIndent(existingAdjustments, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal adjustments: %w", err)
	}

	if err := os.WriteFile(adjustmentsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write adjustments file: %w", err)
	}

	return nil
}

// GetWeightAdjustmentHistory returns weight adjustment history for a method
func (hpd *HistoricalPerformanceData) GetWeightAdjustmentHistory(methodName string) ([]WeightAdjustment, error) {
	adjustmentsFile := "data/weight_adjustments.json"

	// Load adjustments
	var adjustments []WeightAdjustment
	if data, err := os.ReadFile(adjustmentsFile); err != nil {
		return nil, fmt.Errorf("failed to read adjustments file: %w", err)
	} else {
		if err := json.Unmarshal(data, &adjustments); err != nil {
			return nil, fmt.Errorf("failed to unmarshal adjustments: %w", err)
		}
	}

	// Filter by method name
	var methodAdjustments []WeightAdjustment
	for _, adjustment := range adjustments {
		if adjustment.MethodName == methodName {
			methodAdjustments = append(methodAdjustments, adjustment)
		}
	}

	return methodAdjustments, nil
}

// GetPerformanceHistory returns performance history for a method
func (hpd *HistoricalPerformanceData) GetPerformanceHistory(methodName string) (*MethodPerformanceData, bool) {
	hpd.mutex.RLock()
	defer hpd.mutex.RUnlock()

	data, exists := hpd.Data[methodName]
	if !exists {
		return nil, false
	}

	// Return a copy
	copy := &MethodPerformanceData{
		MethodName:         data.MethodName,
		TotalRequests:      data.TotalRequests,
		SuccessfulRequests: data.SuccessfulRequests,
		FailedRequests:     data.FailedRequests,
		AverageAccuracy:    data.AverageAccuracy,
		AverageLatency:     data.AverageLatency,
		LastAccuracy:       data.LastAccuracy,
		LastLatency:        data.LastLatency,
		LastUpdated:        data.LastUpdated,
		CurrentWeight:      data.CurrentWeight,
	}

	// Copy slices
	copy.AccuracyHistory = make([]AccuracyDataPoint, len(data.AccuracyHistory))
	for i, point := range data.AccuracyHistory {
		copy.AccuracyHistory[i] = point
	}

	copy.LatencyHistory = make([]LatencyDataPoint, len(data.LatencyHistory))
	for i, point := range data.LatencyHistory {
		copy.LatencyHistory[i] = point
	}

	copy.WeightHistory = make([]WeightDataPoint, len(data.WeightHistory))
	for i, point := range data.WeightHistory {
		copy.WeightHistory[i] = point
	}

	return copy, true
}

// UpdatePerformanceData updates performance data for a method
func (hpd *HistoricalPerformanceData) UpdatePerformanceData(methodName string, data *MethodPerformanceData) {
	hpd.mutex.Lock()
	defer hpd.mutex.Unlock()

	hpd.Data[methodName] = data
}

// GetHistoricalSummary returns a summary of historical performance data
func (hpd *HistoricalPerformanceData) GetHistoricalSummary() map[string]interface{} {
	hpd.mutex.RLock()
	defer hpd.mutex.RUnlock()

	summary := make(map[string]interface{})

	summary["total_methods"] = len(hpd.Data)
	summary["data_file"] = hpd.DataFile

	// Method summaries
	methodSummaries := make(map[string]interface{})
	for methodName, data := range hpd.Data {
		methodSummary := map[string]interface{}{
			"total_requests":          data.TotalRequests,
			"successful_requests":     data.SuccessfulRequests,
			"failed_requests":         data.FailedRequests,
			"average_accuracy":        data.AverageAccuracy,
			"average_latency_ms":      data.AverageLatency.Milliseconds(),
			"current_weight":          data.CurrentWeight,
			"last_updated":            data.LastUpdated,
			"accuracy_history_points": len(data.AccuracyHistory),
			"latency_history_points":  len(data.LatencyHistory),
			"weight_history_points":   len(data.WeightHistory),
		}

		if data.TotalRequests > 0 {
			methodSummary["success_rate"] = float64(data.SuccessfulRequests) / float64(data.TotalRequests)
			methodSummary["error_rate"] = float64(data.FailedRequests) / float64(data.TotalRequests)
		} else {
			methodSummary["success_rate"] = 0.0
			methodSummary["error_rate"] = 0.0
		}

		methodSummaries[methodName] = methodSummary
	}

	summary["methods"] = methodSummaries

	return summary
}

// CleanupOldData removes old data based on retention policy
func (hpd *HistoricalPerformanceData) CleanupOldData(retentionDays int) error {
	hpd.mutex.Lock()
	defer hpd.mutex.Unlock()

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	for methodName, data := range hpd.Data {
		// Clean up old accuracy history
		var newAccuracyHistory []AccuracyDataPoint
		for _, point := range data.AccuracyHistory {
			if point.Timestamp.After(cutoffTime) {
				newAccuracyHistory = append(newAccuracyHistory, point)
			}
		}
		data.AccuracyHistory = newAccuracyHistory

		// Clean up old latency history
		var newLatencyHistory []LatencyDataPoint
		for _, point := range data.LatencyHistory {
			if point.Timestamp.After(cutoffTime) {
				newLatencyHistory = append(newLatencyHistory, point)
			}
		}
		data.LatencyHistory = newLatencyHistory

		// Clean up old weight history
		var newWeightHistory []WeightDataPoint
		for _, point := range data.WeightHistory {
			if point.Timestamp.After(cutoffTime) {
				newWeightHistory = append(newWeightHistory, point)
			}
		}
		data.WeightHistory = newWeightHistory

		// Update the data
		hpd.Data[methodName] = data
	}

	// Save cleaned data
	return hpd.SaveData()
}

// ExportData exports historical data to a file
func (hpd *HistoricalPerformanceData) ExportData(filename string) error {
	hpd.mutex.RLock()
	defer hpd.mutex.RUnlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	// Marshal data
	data, err := json.MarshalIndent(hpd.Data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Write file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	return nil
}

// ImportData imports historical data from a file
func (hpd *HistoricalPerformanceData) ImportData(filename string) error {
	hpd.mutex.Lock()
	defer hpd.mutex.Unlock()

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	// Unmarshal data
	var importedData map[string]*MethodPerformanceData
	if err := json.Unmarshal(data, &importedData); err != nil {
		return fmt.Errorf("failed to unmarshal import data: %w", err)
	}

	// Merge with existing data
	for methodName, methodData := range importedData {
		hpd.Data[methodName] = methodData
	}

	// Save merged data
	return hpd.SaveData()
}
