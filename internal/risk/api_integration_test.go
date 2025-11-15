package risk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// NewAPIIntegrationTestSuite creates a new API integration test suite
// Type definition is in test_suite_types.go
func NewAPIIntegrationTestSuite(t *testing.T) *APIIntegrationTestSuite {
	logger := zap.NewNop()
	backupDir := t.TempDir()

	// Create services
	exportSvc := NewExportService(logger)
	backupSvc := NewBackupService(logger, backupDir, 30, false)
	jobManager := NewExportJobManager(logger, exportSvc)
	backupJobManager := NewBackupJobManager(logger, backupSvc)

	// Create handlers
	exportHandler := NewExportHandler(logger, exportSvc, jobManager)
	backupHandler := NewBackupHandler(logger, backupSvc, backupJobManager)

	// Create HTTP mux
	mux := http.NewServeMux()
	exportHandler.RegisterRoutes(mux)
	backupHandler.RegisterRoutes(mux)

	// Create test server
	server := httptest.NewServer(mux)

	return &APIIntegrationTestSuite{
		logger:           logger,
		backupDir:        backupDir,
		exportSvc:        exportSvc,
		backupSvc:        backupSvc,
		jobManager:       jobManager,
		backupJobManager: backupJobManager,
		exportHandler:    exportHandler,
		backupHandler:    backupHandler,
		mux:              mux,
		server:           server,
	}
}

// Close closes the test server
// Method is also declared in test_suite_types.go, but this is the actual implementation
func (suite *APIIntegrationTestSuite) Close() {
	suite.server.Close()
}

// TestExportAPIEndpoints tests all export API endpoints
func TestExportAPIEndpoints(t *testing.T) {
	suite := NewAPIIntegrationTestSuite(t)
	defer suite.Close()

	t.Run("POST /api/v1/export/jobs - Create Export Job", func(t *testing.T) {
		exportData := map[string]interface{}{
			"business_id": "test-business-123",
			"export_type": "assessments",
			"format":      "json",
			"metadata":    map[string]interface{}{"test": "api"},
		}

		reqBody, _ := json.Marshal(exportData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.NotEmpty(t, response["job_id"])
		assert.Equal(t, "pending", response["status"])
	})

	t.Run("GET /api/v1/export/jobs/{job_id} - Get Export Job", func(t *testing.T) {
		// First create a job
		exportData := map[string]interface{}{
			"business_id": "test-business-123",
			"export_type": "assessments",
			"format":      "json",
		}

		reqBody, _ := json.Marshal(exportData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		var createResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createResponse)
		require.NoError(t, err)
		jobID := createResponse["job_id"].(string)

		// Now get the job
		req, err = http.NewRequest("GET", suite.server.URL+"/api/v1/export/jobs/"+jobID, nil)
		require.NoError(t, err)

		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, jobID, response["job"].(map[string]interface{})["id"])
	})

	t.Run("GET /api/v1/export/jobs - List Export Jobs", func(t *testing.T) {
		req, err := http.NewRequest("GET", suite.server.URL+"/api/v1/export/jobs", nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Contains(t, response, "jobs")
		assert.Contains(t, response, "total")
	})

	t.Run("DELETE /api/v1/export/jobs/{job_id} - Cancel Export Job", func(t *testing.T) {
		// First create a job
		exportData := map[string]interface{}{
			"business_id": "test-business-123",
			"export_type": "assessments",
			"format":      "json",
		}

		reqBody, _ := json.Marshal(exportData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		var createResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createResponse)
		require.NoError(t, err)
		jobID := createResponse["job_id"].(string)

		// Now cancel the job
		req, err = http.NewRequest("DELETE", suite.server.URL+"/api/v1/export/jobs/"+jobID, nil)
		require.NoError(t, err)

		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "Export job cancelled successfully", response["message"])
	})

	t.Run("POST /api/v1/export/jobs/cleanup - Cleanup Old Jobs", func(t *testing.T) {
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs/cleanup?hours=24", nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "Old jobs cleaned up successfully", response["message"])
	})
}

// TestBackupAPIEndpoints tests all backup API endpoints
func TestBackupAPIEndpoints(t *testing.T) {
	suite := NewAPIIntegrationTestSuite(t)
	defer suite.Close()

	t.Run("POST /api/v1/backup - Create Backup", func(t *testing.T) {
		backupData := map[string]interface{}{
			"business_id":  "test-business-123",
			"backup_type":  "business",
			"include_data": []string{"assessments"},
			"metadata":     map[string]interface{}{"test": "api"},
		}

		reqBody, _ := json.Marshal(backupData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/backup", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.NotEmpty(t, response["backup_id"])
		assert.Equal(t, "completed", response["status"])
	})

	t.Run("GET /api/v1/backup - List Backups", func(t *testing.T) {
		req, err := http.NewRequest("GET", suite.server.URL+"/api/v1/backup", nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Contains(t, response, "backups")
		assert.Contains(t, response, "total")
	})

	t.Run("GET /api/v1/backup/statistics - Get Backup Statistics", func(t *testing.T) {
		req, err := http.NewRequest("GET", suite.server.URL+"/api/v1/backup/statistics", nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Contains(t, response, "statistics")
	})

	t.Run("POST /api/v1/backup/cleanup - Cleanup Expired Backups", func(t *testing.T) {
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/backup/cleanup", nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "Expired backups cleaned up successfully", response["message"])
	})

	t.Run("POST /api/v1/backup/restore - Restore Backup", func(t *testing.T) {
		// First create a backup
		backupData := map[string]interface{}{
			"business_id":  "test-business-123",
			"backup_type":  "business",
			"include_data": []string{"assessments"},
		}

		reqBody, _ := json.Marshal(backupData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/backup", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		var createResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createResponse)
		require.NoError(t, err)
		backupID := createResponse["backup_id"].(string)

		// Now restore the backup
		restoreData := map[string]interface{}{
			"backup_id":    backupID,
			"business_id":  "test-business-123",
			"restore_type": "business",
		}

		reqBody, _ = json.Marshal(restoreData)
		req, err = http.NewRequest("POST", suite.server.URL+"/api/v1/backup/restore", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.NotEmpty(t, response["restore_id"])
		assert.Equal(t, "completed", response["status"])
	})

	t.Run("POST /api/v1/backup/jobs - Create Backup Job", func(t *testing.T) {
		backupData := map[string]interface{}{
			"business_id":  "test-business-123",
			"backup_type":  "business",
			"include_data": []string{"assessments"},
			"metadata":     map[string]interface{}{"test": "api"},
		}

		reqBody, _ := json.Marshal(backupData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/backup/jobs", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-123"))

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.NotEmpty(t, response["job_id"])
		assert.Equal(t, "pending", response["status"])
	})

	t.Run("GET /api/v1/backup/jobs/{job_id} - Get Backup Job", func(t *testing.T) {
		// First create a job
		backupData := map[string]interface{}{
			"business_id":  "test-business-123",
			"backup_type":  "business",
			"include_data": []string{"assessments"},
		}

		reqBody, _ := json.Marshal(backupData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/backup/jobs", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		var createResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createResponse)
		require.NoError(t, err)
		jobID := createResponse["job_id"].(string)

		// Now get the job
		req, err = http.NewRequest("GET", suite.server.URL+"/api/v1/backup/jobs/"+jobID, nil)
		require.NoError(t, err)

		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, jobID, response["job"].(map[string]interface{})["id"])
	})

	t.Run("GET /api/v1/backup/jobs - List Backup Jobs", func(t *testing.T) {
		req, err := http.NewRequest("GET", suite.server.URL+"/api/v1/backup/jobs", nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Contains(t, response, "jobs")
		assert.Contains(t, response, "total")
	})

	t.Run("DELETE /api/v1/backup/jobs/{job_id} - Cancel Backup Job", func(t *testing.T) {
		// First create a job
		backupData := map[string]interface{}{
			"business_id":  "test-business-123",
			"backup_type":  "business",
			"include_data": []string{"assessments"},
		}

		reqBody, _ := json.Marshal(backupData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/backup/jobs", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		var createResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createResponse)
		require.NoError(t, err)
		jobID := createResponse["job_id"].(string)

		// Now cancel the job
		req, err = http.NewRequest("DELETE", suite.server.URL+"/api/v1/backup/jobs/"+jobID, nil)
		require.NoError(t, err)

		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "Backup job cancelled successfully", response["message"])
	})

	t.Run("POST /api/v1/backup/jobs/cleanup - Cleanup Old Backup Jobs", func(t *testing.T) {
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/backup/jobs/cleanup?hours=24", nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "Old jobs cleaned up successfully", response["message"])
	})

	t.Run("POST /api/v1/backup/schedules - Create Backup Schedule", func(t *testing.T) {
		scheduleData := map[string]interface{}{
			"business_id":    "test-business-123",
			"name":           "Daily Backup",
			"description":    "Daily backup schedule",
			"backup_type":    "business",
			"include_data":   []string{"assessments"},
			"schedule":       "0 2 * * *",
			"retention_days": 30,
			"enabled":        true,
		}

		reqBody, _ := json.Marshal(scheduleData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/backup/schedules", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.NotEmpty(t, response["schedule_id"])
		assert.Equal(t, "Daily Backup", response["name"])
	})

	t.Run("GET /api/v1/backup/schedules - List Backup Schedules", func(t *testing.T) {
		req, err := http.NewRequest("GET", suite.server.URL+"/api/v1/backup/schedules", nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Contains(t, response, "schedules")
		assert.Contains(t, response, "total")
	})
}

// TestAPIErrorHandling tests API error handling scenarios
func TestAPIErrorHandling(t *testing.T) {
	suite := NewAPIIntegrationTestSuite(t)
	defer suite.Close()

	t.Run("Invalid JSON Request Body", func(t *testing.T) {
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs", bytes.NewReader([]byte("invalid json")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		exportData := map[string]interface{}{
			"business_id": "test-business-123",
			// Missing export_type and format
		}

		reqBody, _ := json.Marshal(exportData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Invalid Export Type", func(t *testing.T) {
		exportData := map[string]interface{}{
			"business_id": "test-business-123",
			"export_type": "invalid-type",
			"format":      "json",
		}

		reqBody, _ := json.Marshal(exportData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Invalid Backup Type", func(t *testing.T) {
		backupData := map[string]interface{}{
			"business_id":  "test-business-123",
			"backup_type":  "invalid-type",
			"include_data": []string{"assessments"},
		}

		reqBody, _ := json.Marshal(backupData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/backup", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Non-existent Job ID", func(t *testing.T) {
		req, err := http.NewRequest("GET", suite.server.URL+"/api/v1/export/jobs/non-existent-id", nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Non-existent Backup ID", func(t *testing.T) {
		restoreData := map[string]interface{}{
			"backup_id":    "non-existent-backup",
			"business_id":  "test-business-123",
			"restore_type": "business",
		}

		reqBody, _ := json.Marshal(restoreData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/backup/restore", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// TestAPIPerformance tests API performance scenarios
func TestAPIPerformance(t *testing.T) {
	suite := NewAPIIntegrationTestSuite(t)
	defer suite.Close()

	t.Run("Concurrent Export Job Creation", func(t *testing.T) {
		numRequests := 10
		results := make(chan error, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(i int) {
				exportData := map[string]interface{}{
					"business_id": fmt.Sprintf("test-business-%d", i),
					"export_type": "assessments",
					"format":      "json",
				}

				reqBody, _ := json.Marshal(exportData)
				req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs", bytes.NewReader(reqBody))
				if err != nil {
					results <- err
					return
				}
				req.Header.Set("Content-Type", "application/json")

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					results <- err
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusCreated {
					results <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
					return
				}

				results <- nil
			}(i)
		}

		// Wait for all requests to complete
		for i := 0; i < numRequests; i++ {
			err := <-results
			assert.NoError(t, err)
		}
	})

	t.Run("Concurrent Backup Creation", func(t *testing.T) {
		numRequests := 5
		results := make(chan error, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(i int) {
				backupData := map[string]interface{}{
					"business_id":  fmt.Sprintf("test-business-%d", i),
					"backup_type":  "business",
					"include_data": []string{"assessments"},
				}

				reqBody, _ := json.Marshal(backupData)
				req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/backup", bytes.NewReader(reqBody))
				if err != nil {
					results <- err
					return
				}
				req.Header.Set("Content-Type", "application/json")

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					results <- err
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusCreated {
					results <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
					return
				}

				results <- nil
			}(i)
		}

		// Wait for all requests to complete
		for i := 0; i < numRequests; i++ {
			err := <-results
			assert.NoError(t, err)
		}
	})

	t.Run("Large Request Body Handling", func(t *testing.T) {
		// Create a large metadata object
		largeMetadata := make(map[string]interface{})
		for i := 0; i < 1000; i++ {
			largeMetadata[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
		}

		exportData := map[string]interface{}{
			"business_id": "test-business-123",
			"export_type": "assessments",
			"format":      "json",
			"metadata":    largeMetadata,
		}

		reqBody, _ := json.Marshal(exportData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

// TestAPISecurity tests API security scenarios
func TestAPISecurity(t *testing.T) {
	suite := NewAPIIntegrationTestSuite(t)
	defer suite.Close()

	t.Run("SQL Injection Attempt", func(t *testing.T) {
		exportData := map[string]interface{}{
			"business_id": "test-business-123'; DROP TABLE assessments; --",
			"export_type": "assessments",
			"format":      "json",
		}

		reqBody, _ := json.Marshal(exportData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle gracefully without crashing
		assert.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusBadRequest)
	})

	t.Run("XSS Attempt", func(t *testing.T) {
		exportData := map[string]interface{}{
			"business_id": "test-business-123",
			"export_type": "assessments",
			"format":      "json",
			"metadata": map[string]interface{}{
				"script": "<script>alert('xss')</script>",
			},
		}

		reqBody, _ := json.Marshal(exportData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle gracefully
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("Large Request Body", func(t *testing.T) {
		// Create a very large request body
		largeData := make([]byte, 10*1024*1024) // 10MB
		for i := range largeData {
			largeData[i] = 'A'
		}

		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs", bytes.NewReader(largeData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle gracefully
		assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusRequestEntityTooLarge)
	})
}

// TestAPIVersioning tests API versioning scenarios
func TestAPIVersioning(t *testing.T) {
	suite := NewAPIIntegrationTestSuite(t)
	defer suite.Close()

	t.Run("API Version Header", func(t *testing.T) {
		exportData := map[string]interface{}{
			"business_id": "test-business-123",
			"export_type": "assessments",
			"format":      "json",
		}

		reqBody, _ := json.Marshal(exportData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("API-Version", "v1")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("Invalid API Version", func(t *testing.T) {
		exportData := map[string]interface{}{
			"business_id": "test-business-123",
			"export_type": "assessments",
			"format":      "json",
		}

		reqBody, _ := json.Marshal(exportData)
		req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/export/jobs", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("API-Version", "v2")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should still work as we're using v1 endpoints
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}
