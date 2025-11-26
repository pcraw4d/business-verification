package classification

import (
	"context"
	"log"
	"testing"
)

func TestNewConfidenceCalibrator(t *testing.T) {
	calibrator := NewConfidenceCalibrator(log.Default())
	
	if calibrator == nil {
		t.Fatal("NewConfidenceCalibrator() returned nil")
	}
	
	if calibrator.numBins != 10 {
		t.Errorf("Expected 10 bins, got %d", calibrator.numBins)
	}
	
	if calibrator.targetAccuracy != 0.95 {
		t.Errorf("Expected target accuracy 0.95, got %.2f", calibrator.targetAccuracy)
	}
	
	if len(calibrator.calibrationBins) != 10 {
		t.Errorf("Expected 10 calibration bins, got %d", len(calibrator.calibrationBins))
	}
}

func TestRecordClassification(t *testing.T) {
	calibrator := NewConfidenceCalibrator(log.Default())
	
	// Record some classifications
	err := calibrator.RecordClassification(context.Background(), 0.9, "Technology", "Technology", true)
	if err != nil {
		t.Fatalf("RecordClassification() failed: %v", err)
	}
	
	err = calibrator.RecordClassification(context.Background(), 0.8, "Healthcare", "Healthcare", true)
	if err != nil {
		t.Fatalf("RecordClassification() failed: %v", err)
	}
	
	err = calibrator.RecordClassification(context.Background(), 0.7, "Technology", "Healthcare", false)
	if err != nil {
		t.Fatalf("RecordClassification() failed: %v", err)
	}
	
	stats := calibrator.GetStatistics()
	if stats["total_classifications"].(int64) != 3 {
		t.Errorf("Expected 3 classifications, got %d", stats["total_classifications"])
	}
	
	if stats["correct_classifications"].(int64) != 2 {
		t.Errorf("Expected 2 correct classifications, got %d", stats["correct_classifications"])
	}
}

func TestCalibrate(t *testing.T) {
	calibrator := NewConfidenceCalibrator(log.Default())
	
	// Record enough classifications to trigger calibration
	for i := 0; i < 50; i++ {
		confidence := 0.5 + float64(i%5)*0.1
		isCorrect := i%10 != 0 // 90% accuracy
		calibrator.RecordClassification(context.Background(), confidence, "Industry", "Industry", isCorrect)
	}
	
	result, err := calibrator.Calibrate(context.Background())
	if err != nil {
		t.Fatalf("Calibrate() failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("Calibrate() returned nil result")
	}
	
	if len(result.CalibrationBins) == 0 {
		t.Error("Expected calibration bins, got none")
	}
	
	if result.RecommendedThreshold < 0.0 || result.RecommendedThreshold > 1.0 {
		t.Errorf("Recommended threshold out of range: %.2f", result.RecommendedThreshold)
	}
}

func TestAdjustConfidence(t *testing.T) {
	calibrator := NewConfidenceCalibrator(log.Default())
	
	// Record some classifications to build calibration data
	for i := 0; i < 20; i++ {
		confidence := 0.8
		isCorrect := i < 18 // 90% accuracy (should calibrate down)
		calibrator.RecordClassification(context.Background(), confidence, "Industry", "Industry", isCorrect)
	}
	
	// Calibrate
	_, err := calibrator.Calibrate(context.Background())
	if err != nil {
		t.Fatalf("Calibrate() failed: %v", err)
	}
	
	// Test adjustment
	originalConfidence := 0.8
	adjusted := calibrator.AdjustConfidence(originalConfidence)
	
	if adjusted < 0.0 || adjusted > 1.0 {
		t.Errorf("Adjusted confidence out of range: %.2f", adjusted)
	}
}

func TestGetRecommendedConfidenceThreshold(t *testing.T) {
	calibrator := NewConfidenceCalibrator(log.Default())
	
	// Record classifications with varying accuracy
	for i := 0; i < 100; i++ {
		confidence := 0.5 + float64(i%5)*0.1
		// Higher confidence should have higher accuracy
		isCorrect := confidence >= 0.8 || (confidence >= 0.6 && i%3 == 0)
		calibrator.RecordClassification(context.Background(), confidence, "Industry", "Industry", isCorrect)
	}
	
	threshold, err := calibrator.GetRecommendedConfidenceThreshold(context.Background())
	if err != nil {
		t.Fatalf("GetRecommendedConfidenceThreshold() failed: %v", err)
	}
	
	if threshold < 0.0 || threshold > 1.0 {
		t.Errorf("Recommended threshold out of range: %.2f", threshold)
	}
}

func TestSetTargetAccuracy(t *testing.T) {
	calibrator := NewConfidenceCalibrator(log.Default())
	
	calibrator.SetTargetAccuracy(0.90)
	
	if calibrator.GetTargetAccuracy() != 0.90 {
		t.Errorf("Expected target accuracy 0.90, got %.2f", calibrator.GetTargetAccuracy())
	}
}

func TestShouldRecalibrate(t *testing.T) {
	calibrator := NewConfidenceCalibrator(log.Default())
	
	// Should recalibrate if never calibrated
	if !calibrator.ShouldRecalibrate() {
		t.Error("Should recalibrate when never calibrated")
	}
	
	// Calibrate
	_, err := calibrator.Calibrate(context.Background())
	if err != nil {
		t.Fatalf("Calibrate() failed: %v", err)
	}
	
	// Should not recalibrate immediately after calibration
	if calibrator.ShouldRecalibrate() {
		t.Error("Should not recalibrate immediately after calibration")
	}
}

func BenchmarkRecordClassification(b *testing.B) {
	calibrator := NewConfidenceCalibrator(log.Default())
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		confidence := 0.5 + float64(i%10)*0.05
		isCorrect := i%10 != 0
		calibrator.RecordClassification(context.Background(), confidence, "Industry", "Industry", isCorrect)
	}
}

func BenchmarkAdjustConfidence(b *testing.B) {
	calibrator := NewConfidenceCalibrator(log.Default())
	
	// Build calibration data
	for i := 0; i < 100; i++ {
		confidence := 0.5 + float64(i%10)*0.05
		isCorrect := i%10 != 0
		calibrator.RecordClassification(context.Background(), confidence, "Industry", "Industry", isCorrect)
	}
	
	calibrator.Calibrate(context.Background())
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		confidence := 0.5 + float64(i%10)*0.05
		calibrator.AdjustConfidence(confidence)
	}
}

