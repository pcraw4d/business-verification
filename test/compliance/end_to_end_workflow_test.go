package compliance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kyb-platform/internal/api/routes"
	"kyb-platform/internal/observability"
	"go.uber.org/zap"
)

// TestEndToEndComplianceWorkflow tests the complete compliance workflow from start to finish
func TestEndToEndComplianceWorkflow(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create test server with all compliance routes
	mux := http.NewServeMux()
	routes.RegisterComplianceStatusRoutes(mux, logger)
	routes.RegisterComplianceFrameworkRoutes(mux, logger)
	routes.RegisterComplianceTrackingRoutes(mux, logger)
	routes.RegisterComplianceReportingRoutes(mux, logger)
	routes.RegisterComplianceAlertRoutes(mux, logger)

	// Test business ID for the workflow
	businessID := "test-business-workflow-123"
	frameworkID := "SOC2"
	assessorID := "assessor-workflow-456"

	t.Run("Complete Compliance Workflow", func(t *testing.T) {
		// Step 1: Get initial compliance status
		t.Log("Step 1: Getting initial compliance status...")
		statusReq, _ := http.NewRequest("GET", fmt.Sprintf("/v1/compliance/status/%s", businessID), nil)
		statusReq = statusReq.WithContext(context.WithValue(statusReq.Context(), "request_id", "workflow-test-1"))
		statusRR := httptest.NewRecorder()
		mux.ServeHTTP(statusRR, statusReq)

		if statusRR.Code != http.StatusOK {
			t.Errorf("Failed to get initial compliance status: %d - %s", statusRR.Code, statusRR.Body.String())
		}

		// Step 2: List available compliance frameworks
		t.Log("Step 2: Listing available compliance frameworks...")
		frameworksReq, _ := http.NewRequest("GET", "/v1/compliance/frameworks", nil)
		frameworksReq = frameworksReq.WithContext(context.WithValue(frameworksReq.Context(), "request_id", "workflow-test-2"))
		frameworksRR := httptest.NewRecorder()
		mux.ServeHTTP(frameworksRR, frameworksReq)

		if frameworksRR.Code != http.StatusOK {
			t.Errorf("Failed to list frameworks: %d - %s", frameworksRR.Code, frameworksRR.Body.String())
		}

		// Step 3: Get framework requirements
		t.Log("Step 3: Getting framework requirements...")
		requirementsReq, _ := http.NewRequest("GET", fmt.Sprintf("/v1/compliance/frameworks/%s/requirements", frameworkID), nil)
		requirementsReq = requirementsReq.WithContext(context.WithValue(requirementsReq.Context(), "request_id", "workflow-test-3"))
		requirementsRR := httptest.NewRecorder()
		mux.ServeHTTP(requirementsRR, requirementsReq)

		if requirementsRR.Code != http.StatusOK {
			t.Errorf("Failed to get framework requirements: %d - %s", requirementsRR.Code, requirementsRR.Body.String())
		}

		// Step 4: Create compliance assessment
		t.Log("Step 4: Creating compliance assessment...")
		assessmentData := map[string]interface{}{
			"business_id":     businessID,
			"framework_id":    frameworkID,
			"assessor_id":     assessorID,
			"assessment_type": "initial",
			"scope":           []string{"security", "availability", "processing_integrity", "confidentiality", "privacy"},
		}
		assessmentBody, _ := json.Marshal(assessmentData)
		assessmentReq, _ := http.NewRequest("POST", "/v1/compliance/assessments", bytes.NewBuffer(assessmentBody))
		assessmentReq.Header.Set("Content-Type", "application/json")
		assessmentReq = assessmentReq.WithContext(context.WithValue(assessmentReq.Context(), "request_id", "workflow-test-4"))
		assessmentRR := httptest.NewRecorder()
		mux.ServeHTTP(assessmentRR, assessmentReq)

		if assessmentRR.Code != http.StatusCreated {
			t.Errorf("Failed to create assessment: %d - %s", assessmentRR.Code, assessmentRR.Body.String())
		}

		// Extract assessment ID from response
		var assessmentResponse map[string]interface{}
		if err := json.Unmarshal(assessmentRR.Body.Bytes(), &assessmentResponse); err != nil {
			t.Errorf("Failed to parse assessment response: %v", err)
		}
		assessmentID, ok := assessmentResponse["assessment_id"].(string)
		if !ok {
			t.Errorf("Assessment ID not found in response")
		}

		// Step 5: Update compliance tracking
		t.Log("Step 5: Updating compliance tracking...")
		trackingData := map[string]interface{}{
			"overall_progress": 0.25,
			"compliance_level": "partial",
			"last_assessment":  assessmentID,
			"next_review_date": time.Now().AddDate(0, 3, 0).Format(time.RFC3339),
		}
		trackingBody, _ := json.Marshal(trackingData)
		trackingReq, _ := http.NewRequest("PUT", fmt.Sprintf("/v1/compliance/tracking/%s/%s", businessID, frameworkID), bytes.NewBuffer(trackingBody))
		trackingReq.Header.Set("Content-Type", "application/json")
		trackingReq = trackingReq.WithContext(context.WithValue(trackingReq.Context(), "request_id", "workflow-test-5"))
		trackingRR := httptest.NewRecorder()
		mux.ServeHTTP(trackingRR, trackingReq)

		if trackingRR.Code != http.StatusOK {
			t.Errorf("Failed to update tracking: %d - %s", trackingRR.Code, trackingRR.Body.String())
		}

		// Step 6: Create compliance milestone
		t.Log("Step 6: Creating compliance milestone...")
		milestoneData := map[string]interface{}{
			"business_id":    businessID,
			"framework_id":   frameworkID,
			"name":           "Initial Assessment Complete",
			"type":           "assessment",
			"status":         "completed",
			"target_date":    time.Now().Format(time.RFC3339),
			"completed_date": time.Now().Format(time.RFC3339),
		}
		milestoneBody, _ := json.Marshal(milestoneData)
		milestoneReq, _ := http.NewRequest("POST", "/v1/compliance/milestones", bytes.NewBuffer(milestoneBody))
		milestoneReq.Header.Set("Content-Type", "application/json")
		milestoneReq = milestoneReq.WithContext(context.WithValue(milestoneReq.Context(), "request_id", "workflow-test-6"))
		milestoneRR := httptest.NewRecorder()
		mux.ServeHTTP(milestoneRR, milestoneReq)

		if milestoneRR.Code != http.StatusCreated {
			t.Errorf("Failed to create milestone: %d - %s", milestoneRR.Code, milestoneRR.Body.String())
		}

		// Step 7: Generate compliance report
		t.Log("Step 7: Generating compliance report...")
		reportData := map[string]interface{}{
			"business_id":     businessID,
			"framework_id":    frameworkID,
			"report_type":     "status",
			"generated_by":    assessorID,
			"include_history": true,
		}
		reportBody, _ := json.Marshal(reportData)
		reportReq, _ := http.NewRequest("POST", "/v1/compliance/reports", bytes.NewBuffer(reportBody))
		reportReq.Header.Set("Content-Type", "application/json")
		reportReq = reportReq.WithContext(context.WithValue(reportReq.Context(), "request_id", "workflow-test-7"))
		reportRR := httptest.NewRecorder()
		mux.ServeHTTP(reportRR, reportReq)

		if reportRR.Code != http.StatusCreated {
			t.Errorf("Failed to generate report: %d - %s", reportRR.Code, reportRR.Body.String())
		}

		// Step 8: Create compliance alert
		t.Log("Step 8: Creating compliance alert...")
		alertData := map[string]interface{}{
			"business_id":  businessID,
			"framework_id": frameworkID,
			"alert_type":   "compliance_change",
			"severity":     "medium",
			"title":        "Initial Assessment Completed",
			"description":  "Initial compliance assessment has been completed with partial compliance status",
			"created_by":   assessorID,
		}
		alertBody, _ := json.Marshal(alertData)
		alertReq, _ := http.NewRequest("POST", "/v1/compliance/alerts", bytes.NewBuffer(alertBody))
		alertReq.Header.Set("Content-Type", "application/json")
		alertReq = alertReq.WithContext(context.WithValue(alertReq.Context(), "request_id", "workflow-test-8"))
		alertRR := httptest.NewRecorder()
		mux.ServeHTTP(alertRR, alertReq)

		if alertRR.Code != http.StatusCreated {
			t.Errorf("Failed to create alert: %d - %s", alertRR.Code, alertRR.Body.String())
		}

		// Step 9: Get updated compliance status
		t.Log("Step 9: Getting updated compliance status...")
		updatedStatusReq, _ := http.NewRequest("GET", fmt.Sprintf("/v1/compliance/status/%s", businessID), nil)
		updatedStatusReq = updatedStatusReq.WithContext(context.WithValue(updatedStatusReq.Context(), "request_id", "workflow-test-9"))
		updatedStatusRR := httptest.NewRecorder()
		mux.ServeHTTP(updatedStatusRR, updatedStatusReq)

		if updatedStatusRR.Code != http.StatusOK {
			t.Errorf("Failed to get updated compliance status: %d - %s", updatedStatusRR.Code, updatedStatusRR.Body.String())
		}

		// Step 10: Get compliance history
		t.Log("Step 10: Getting compliance history...")
		historyReq, _ := http.NewRequest("GET", fmt.Sprintf("/v1/compliance/status/%s/history", businessID), nil)
		historyReq = historyReq.WithContext(context.WithValue(historyReq.Context(), "request_id", "workflow-test-10"))
		historyRR := httptest.NewRecorder()
		mux.ServeHTTP(historyRR, historyReq)

		if historyRR.Code != http.StatusOK {
			t.Errorf("Failed to get compliance history: %d - %s", historyRR.Code, historyRR.Body.String())
		}

		// Step 11: Get progress metrics
		t.Log("Step 11: Getting progress metrics...")
		metricsReq, _ := http.NewRequest("GET", fmt.Sprintf("/v1/compliance/metrics/%s/%s", businessID, frameworkID), nil)
		metricsReq = metricsReq.WithContext(context.WithValue(metricsReq.Context(), "request_id", "workflow-test-11"))
		metricsRR := httptest.NewRecorder()
		mux.ServeHTTP(metricsRR, metricsReq)

		if metricsRR.Code != http.StatusOK {
			t.Errorf("Failed to get progress metrics: %d - %s", metricsRR.Code, metricsRR.Body.String())
		}

		// Step 12: Get compliance trends
		t.Log("Step 12: Getting compliance trends...")
		trendsReq, _ := http.NewRequest("GET", fmt.Sprintf("/v1/compliance/trends/%s/%s", businessID, frameworkID), nil)
		trendsReq = trendsReq.WithContext(context.WithValue(trendsReq.Context(), "request_id", "workflow-test-12"))
		trendsRR := httptest.NewRecorder()
		mux.ServeHTTP(trendsRR, trendsReq)

		if trendsRR.Code != http.StatusOK {
			t.Errorf("Failed to get compliance trends: %d - %s", trendsRR.Code, trendsRR.Body.String())
		}

		t.Log("✅ Complete compliance workflow test passed successfully")
	})
}

// TestComplianceWorkflowWithMultipleFrameworks tests workflow with multiple compliance frameworks
func TestComplianceWorkflowWithMultipleFrameworks(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create test server with all compliance routes
	mux := http.NewServeMux()
	routes.RegisterComplianceStatusRoutes(mux, logger)
	routes.RegisterComplianceFrameworkRoutes(mux, logger)
	routes.RegisterComplianceTrackingRoutes(mux, logger)
	routes.RegisterComplianceReportingRoutes(mux, logger)
	routes.RegisterComplianceAlertRoutes(mux, logger)

	businessID := "test-business-multi-framework-123"
	frameworks := []string{"SOC2", "GDPR", "PCI-DSS"}
	assessorID := "assessor-multi-456"

	t.Run("Multi-Framework Compliance Workflow", func(t *testing.T) {
		for _, frameworkID := range frameworks {
			t.Logf("Testing workflow for framework: %s", frameworkID)

			// Create assessment for each framework
			assessmentData := map[string]interface{}{
				"business_id":     businessID,
				"framework_id":    frameworkID,
				"assessor_id":     assessorID,
				"assessment_type": "initial",
			}
			assessmentBody, _ := json.Marshal(assessmentData)
			assessmentReq, _ := http.NewRequest("POST", "/v1/compliance/assessments", bytes.NewBuffer(assessmentBody))
			assessmentReq.Header.Set("Content-Type", "application/json")
			assessmentReq = assessmentReq.WithContext(context.WithValue(assessmentReq.Context(), "request_id", fmt.Sprintf("multi-framework-test-%s", frameworkID)))
			assessmentRR := httptest.NewRecorder()
			mux.ServeHTTP(assessmentRR, assessmentReq)

			if assessmentRR.Code != http.StatusCreated {
				t.Errorf("Failed to create assessment for %s: %d - %s", frameworkID, assessmentRR.Code, assessmentRR.Body.String())
			}

			// Update tracking for each framework
			trackingData := map[string]interface{}{
				"overall_progress": 0.30,
				"compliance_level": "partial",
			}
			trackingBody, _ := json.Marshal(trackingData)
			trackingReq, _ := http.NewRequest("PUT", fmt.Sprintf("/v1/compliance/tracking/%s/%s", businessID, frameworkID), bytes.NewBuffer(trackingBody))
			trackingReq.Header.Set("Content-Type", "application/json")
			trackingReq = trackingReq.WithContext(context.WithValue(trackingReq.Context(), "request_id", fmt.Sprintf("multi-tracking-test-%s", frameworkID)))
			trackingRR := httptest.NewRecorder()
			mux.ServeHTTP(trackingRR, trackingReq)

			if trackingRR.Code != http.StatusOK {
				t.Errorf("Failed to update tracking for %s: %d - %s", frameworkID, trackingRR.Code, trackingRR.Body.String())
			}
		}

		// Get overall compliance status
		statusReq, _ := http.NewRequest("GET", fmt.Sprintf("/v1/compliance/status/%s", businessID), nil)
		statusReq = statusReq.WithContext(context.WithValue(statusReq.Context(), "request_id", "multi-framework-status-test"))
		statusRR := httptest.NewRecorder()
		mux.ServeHTTP(statusRR, statusReq)

		if statusRR.Code != http.StatusOK {
			t.Errorf("Failed to get overall compliance status: %d - %s", statusRR.Code, statusRR.Body.String())
		}

		t.Log("✅ Multi-framework compliance workflow test passed successfully")
	})
}

// TestComplianceWorkflowErrorScenarios tests error handling in the compliance workflow
func TestComplianceWorkflowErrorScenarios(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create test server with all compliance routes
	mux := http.NewServeMux()
	routes.RegisterComplianceStatusRoutes(mux, logger)
	routes.RegisterComplianceFrameworkRoutes(mux, logger)
	routes.RegisterComplianceTrackingRoutes(mux, logger)
	routes.RegisterComplianceReportingRoutes(mux, logger)
	routes.RegisterComplianceAlertRoutes(mux, logger)

	t.Run("Error Scenario Testing", func(t *testing.T) {
		// Test 1: Invalid business ID
		t.Log("Testing invalid business ID scenario...")
		invalidStatusReq, _ := http.NewRequest("GET", "/v1/compliance/status/", nil)
		invalidStatusReq = invalidStatusReq.WithContext(context.WithValue(invalidStatusReq.Context(), "request_id", "error-test-1"))
		invalidStatusRR := httptest.NewRecorder()
		mux.ServeHTTP(invalidStatusRR, invalidStatusReq)

		if invalidStatusRR.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 for invalid business ID, got %d", invalidStatusRR.Code)
		}

		// Test 2: Non-existent framework
		t.Log("Testing non-existent framework scenario...")
		invalidFrameworkReq, _ := http.NewRequest("GET", "/v1/compliance/frameworks/INVALID-FRAMEWORK/requirements", nil)
		invalidFrameworkReq = invalidFrameworkReq.WithContext(context.WithValue(invalidFrameworkReq.Context(), "request_id", "error-test-2"))
		invalidFrameworkRR := httptest.NewRecorder()
		mux.ServeHTTP(invalidFrameworkRR, invalidFrameworkReq)

		if invalidFrameworkRR.Code != http.StatusNotFound {
			t.Errorf("Expected 404 for non-existent framework, got %d", invalidFrameworkRR.Code)
		}

		// Test 3: Invalid assessment data
		t.Log("Testing invalid assessment data scenario...")
		invalidAssessmentData := map[string]interface{}{
			"business_id": "test-business-123",
			// Missing required fields: framework_id, assessor_id
		}
		invalidAssessmentBody, _ := json.Marshal(invalidAssessmentData)
		invalidAssessmentReq, _ := http.NewRequest("POST", "/v1/compliance/assessments", bytes.NewBuffer(invalidAssessmentBody))
		invalidAssessmentReq.Header.Set("Content-Type", "application/json")
		invalidAssessmentReq = invalidAssessmentReq.WithContext(context.WithValue(invalidAssessmentReq.Context(), "request_id", "error-test-3"))
		invalidAssessmentRR := httptest.NewRecorder()
		mux.ServeHTTP(invalidAssessmentRR, invalidAssessmentReq)

		if invalidAssessmentRR.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 for invalid assessment data, got %d", invalidAssessmentRR.Code)
		}

		// Test 4: Invalid tracking data
		t.Log("Testing invalid tracking data scenario...")
		invalidTrackingData := map[string]interface{}{
			"overall_progress": 1.5, // Invalid: should be 0-1
			"compliance_level": "invalid_level",
		}
		invalidTrackingBody, _ := json.Marshal(invalidTrackingData)
		invalidTrackingReq, _ := http.NewRequest("PUT", "/v1/compliance/tracking/test-business-123/SOC2", bytes.NewBuffer(invalidTrackingBody))
		invalidTrackingReq.Header.Set("Content-Type", "application/json")
		invalidTrackingReq = invalidTrackingReq.WithContext(context.WithValue(invalidTrackingReq.Context(), "request_id", "error-test-4"))
		invalidTrackingRR := httptest.NewRecorder()
		mux.ServeHTTP(invalidTrackingRR, invalidTrackingReq)

		if invalidTrackingRR.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 for invalid tracking data, got %d", invalidTrackingRR.Code)
		}

		// Test 5: Invalid report type
		t.Log("Testing invalid report type scenario...")
		invalidReportData := map[string]interface{}{
			"business_id":  "test-business-123",
			"framework_id": "SOC2",
			"report_type":  "invalid_report_type",
			"generated_by": "test-user",
		}
		invalidReportBody, _ := json.Marshal(invalidReportData)
		invalidReportReq, _ := http.NewRequest("POST", "/v1/compliance/reports", bytes.NewBuffer(invalidReportBody))
		invalidReportReq.Header.Set("Content-Type", "application/json")
		invalidReportReq = invalidReportReq.WithContext(context.WithValue(invalidReportReq.Context(), "request_id", "error-test-5"))
		invalidReportRR := httptest.NewRecorder()
		mux.ServeHTTP(invalidReportRR, invalidReportReq)

		if invalidReportRR.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 for invalid report type, got %d", invalidReportRR.Code)
		}

		t.Log("✅ Error scenario testing completed successfully")
	})
}

// TestComplianceWorkflowPerformance tests performance of the complete compliance workflow
func TestComplianceWorkflowPerformance(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create test server with all compliance routes
	mux := http.NewServeMux()
	routes.RegisterComplianceStatusRoutes(mux, logger)
	routes.RegisterComplianceFrameworkRoutes(mux, logger)
	routes.RegisterComplianceTrackingRoutes(mux, logger)
	routes.RegisterComplianceReportingRoutes(mux, logger)
	routes.RegisterComplianceAlertRoutes(mux, logger)

	businessID := "test-business-performance-123"
	frameworkID := "SOC2"
	assessorID := "assessor-performance-456"

	t.Run("Workflow Performance Testing", func(t *testing.T) {
		// Measure complete workflow execution time
		start := time.Now()

		// Step 1: Get compliance status
		statusReq, _ := http.NewRequest("GET", fmt.Sprintf("/v1/compliance/status/%s", businessID), nil)
		statusReq = statusReq.WithContext(context.WithValue(statusReq.Context(), "request_id", "perf-test-1"))
		statusRR := httptest.NewRecorder()
		mux.ServeHTTP(statusRR, statusReq)

		// Step 2: Create assessment
		assessmentData := map[string]interface{}{
			"business_id":  businessID,
			"framework_id": frameworkID,
			"assessor_id":  assessorID,
		}
		assessmentBody, _ := json.Marshal(assessmentData)
		assessmentReq, _ := http.NewRequest("POST", "/v1/compliance/assessments", bytes.NewBuffer(assessmentBody))
		assessmentReq.Header.Set("Content-Type", "application/json")
		assessmentReq = assessmentReq.WithContext(context.WithValue(assessmentReq.Context(), "request_id", "perf-test-2"))
		assessmentRR := httptest.NewRecorder()
		mux.ServeHTTP(assessmentRR, assessmentReq)

		// Step 3: Update tracking
		trackingData := map[string]interface{}{
			"overall_progress": 0.50,
			"compliance_level": "partial",
		}
		trackingBody, _ := json.Marshal(trackingData)
		trackingReq, _ := http.NewRequest("PUT", fmt.Sprintf("/v1/compliance/tracking/%s/%s", businessID, frameworkID), bytes.NewBuffer(trackingBody))
		trackingReq.Header.Set("Content-Type", "application/json")
		trackingReq = trackingReq.WithContext(context.WithValue(trackingReq.Context(), "request_id", "perf-test-3"))
		trackingRR := httptest.NewRecorder()
		mux.ServeHTTP(trackingRR, trackingReq)

		// Step 4: Generate report
		reportData := map[string]interface{}{
			"business_id":  businessID,
			"framework_id": frameworkID,
			"report_type":  "status",
			"generated_by": assessorID,
		}
		reportBody, _ := json.Marshal(reportData)
		reportReq, _ := http.NewRequest("POST", "/v1/compliance/reports", bytes.NewBuffer(reportBody))
		reportReq.Header.Set("Content-Type", "application/json")
		reportReq = reportReq.WithContext(context.WithValue(reportReq.Context(), "request_id", "perf-test-4"))
		reportRR := httptest.NewRecorder()
		mux.ServeHTTP(reportRR, reportReq)

		duration := time.Since(start)

		// Check that all requests were successful
		if statusRR.Code != http.StatusOK {
			t.Errorf("Status request failed: %d", statusRR.Code)
		}
		if assessmentRR.Code != http.StatusCreated {
			t.Errorf("Assessment request failed: %d", assessmentRR.Code)
		}
		if trackingRR.Code != http.StatusOK {
			t.Errorf("Tracking request failed: %d", trackingRR.Code)
		}
		if reportRR.Code != http.StatusCreated {
			t.Errorf("Report request failed: %d", reportRR.Code)
		}

		// Check performance (should complete within 2 seconds)
		maxDuration := 2 * time.Second
		if duration > maxDuration {
			t.Errorf("Workflow took %v, expected less than %v", duration, maxDuration)
		}

		t.Logf("✅ Workflow performance test passed: completed in %v", duration)
	})
}
