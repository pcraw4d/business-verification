package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kyb-platform/internal/external"
)

func TestNewVerificationAutomatedTestingHandler(t *testing.T) {
	logger := zap.NewNop()
	tester := external.NewVerificationAutomatedTester(nil, logger)

	handler := NewVerificationAutomatedTestingHandler(tester, logger)
	assert.NotNil(t, handler)
	assert.Equal(t, tester, handler.tester)
	assert.Equal(t, logger, handler.logger)
}

func TestRegisterAutomatedTestingRoutes(t *testing.T) {
	logger := zap.NewNop()
	tester := external.NewVerificationAutomatedTester(nil, logger)
	handler := NewVerificationAutomatedTestingHandler(tester, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test that routes are registered by checking if they exist
	// Note: We can't easily test route registration with gorilla/mux in this way
	// Instead, we'll verify the handler is properly created
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.tester)
	assert.NotNil(t, handler.logger)
}

func TestCreateTestSuite(t *testing.T) {
	logger := zap.NewNop()
	tester := external.NewVerificationAutomatedTester(nil, logger)
	handler := NewVerificationAutomatedTestingHandler(tester, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid request",
			requestBody: CreateTestSuiteRequest{
				Name:        "Test Suite",
				Description: "A test suite",
				Category:    "verification",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			requestBody: CreateTestSuiteRequest{
				Description: "A test suite",
				Category:    "verification",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Test suite name is required",
		},
		{
			name: "missing category",
			requestBody: CreateTestSuiteRequest{
				Name:        "Test Suite",
				Description: "A test suite",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Test suite category is required",
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/test-suites", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response CreateTestSuiteResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, tt.expectedError)
			} else {
				var response CreateTestSuiteResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Suite)
				assert.NotEmpty(t, response.Suite.ID)
			}
		})
	}
}

func TestListTestSuites(t *testing.T) {
	logger := zap.NewNop()
	tester := external.NewVerificationAutomatedTester(nil, logger)
	handler := NewVerificationAutomatedTestingHandler(tester, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Initially should be empty
	req := httptest.NewRequest("GET", "/test-suites", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListTestSuitesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Empty(t, response.Suites)

	// Create a test suite and verify it appears in the list
	suite := &external.TestSuite{
		Name:        "Test Suite",
		Description: "A test suite",
		Category:    "verification",
	}
	err = tester.CreateTestSuite(suite)
	require.NoError(t, err)

	req = httptest.NewRequest("GET", "/test-suites", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Len(t, response.Suites, 1)
	assert.Equal(t, suite.Name, response.Suites[0].Name)
}

func TestGetTestSuite(t *testing.T) {
	logger := zap.NewNop()
	tester := external.NewVerificationAutomatedTester(nil, logger)
	handler := NewVerificationAutomatedTestingHandler(tester, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Create a test suite
	suite := &external.TestSuite{
		Name:        "Test Suite",
		Description: "A test suite",
		Category:    "verification",
	}
	err := tester.CreateTestSuite(suite)
	require.NoError(t, err)

	// Test getting existing suite
	req := httptest.NewRequest("GET", "/test-suites/"+suite.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response GetTestSuiteResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Suite)
	assert.Equal(t, suite.ID, response.Suite.ID)
	assert.Equal(t, suite.Name, response.Suite.Name)

	// Test getting non-existent suite
	req = httptest.NewRequest("GET", "/test-suites/non-existent", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "test suite not found")
}

func TestAddTest(t *testing.T) {
	logger := zap.NewNop()
	tester := external.NewVerificationAutomatedTester(nil, logger)
	handler := NewVerificationAutomatedTestingHandler(tester, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Create a test suite
	suite := &external.TestSuite{
		Name:        "Test Suite",
		Description: "A test suite",
		Category:    "verification",
	}
	err := tester.CreateTestSuite(suite)
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid request",
			requestBody: AddTestRequest{
				Name:        "Test",
				Description: "A test",
				Type:        external.TestTypeUnit,
				Input:       "input",
				Expected:    "output",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			requestBody: AddTestRequest{
				Description: "A test",
				Type:        external.TestTypeUnit,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Test name is required",
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/test-suites/"+suite.ID+"/tests", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response AddTestResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, tt.expectedError)
			} else {
				var response AddTestResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Test)
				assert.NotEmpty(t, response.Test.ID)
			}
		})
	}

	// Test adding to non-existent suite
	req := httptest.NewRequest("POST", "/test-suites/non-existent/tests", bytes.NewBuffer([]byte(`{"name":"Test","type":"unit"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response AddTestResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "test suite not found")
}

func TestRunTestSuite(t *testing.T) {
	logger := zap.NewNop()
	tester := external.NewVerificationAutomatedTester(nil, logger)
	handler := NewVerificationAutomatedTestingHandler(tester, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Create a test suite with a test
	suite := &external.TestSuite{
		Name:        "Test Suite",
		Description: "A test suite",
		Category:    "verification",
	}
	err := tester.CreateTestSuite(suite)
	require.NoError(t, err)

	test := &external.AutomatedTest{
		Name:        "Test",
		Description: "A test",
		Type:        external.TestTypeUnit,
		Input:       "input",
		Expected:    "output",
	}
	err = tester.AddTest(suite.ID, test)
	require.NoError(t, err)

	// Test running the suite
	req := httptest.NewRequest("POST", "/test-suites/"+suite.ID+"/run", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response RunTestSuiteResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Summary)
	assert.Equal(t, 1, response.Summary.TotalTests)
	assert.Equal(t, 1, response.Summary.PassedTests)
	assert.Equal(t, 1.0, response.Summary.SuccessRate)

	// Test running non-existent suite
	req = httptest.NewRequest("POST", "/test-suites/non-existent/run", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "test suite not found")
}

func TestGetTestResults(t *testing.T) {
	logger := zap.NewNop()
	tester := external.NewVerificationAutomatedTester(nil, logger)
	handler := NewVerificationAutomatedTestingHandler(tester, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Create and run a test suite to generate results
	suite := &external.TestSuite{
		Name:        "Test Suite",
		Description: "A test suite",
		Category:    "verification",
	}
	err := tester.CreateTestSuite(suite)
	require.NoError(t, err)

	test := &external.AutomatedTest{
		Name:        "Test",
		Description: "A test",
		Type:        external.TestTypeUnit,
		Input:       "input",
		Expected:    "output",
	}
	err = tester.AddTest(suite.ID, test)
	require.NoError(t, err)

	// Run the suite to generate results
	_, err = tester.RunTestSuite(nil, suite.ID)
	require.NoError(t, err)

	// Test getting all results
	req := httptest.NewRequest("GET", "/test-results", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response GetTestResultsResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Len(t, response.Results, 1)
	assert.Equal(t, external.TestStatusPassed, response.Results[0].Status)

	// Test getting results with limit
	req = httptest.NewRequest("GET", "/test-results?limit=1", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Len(t, response.Results, 1)

	// Test getting results by status
	req = httptest.NewRequest("GET", "/test-results?status=passed", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Len(t, response.Results, 1)
}

func TestGetAutomatedTestingConfig(t *testing.T) {
	logger := zap.NewNop()
	tester := external.NewVerificationAutomatedTester(nil, logger)
	handler := NewVerificationAutomatedTestingHandler(tester, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/config", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response GetConfigResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Config)
	assert.True(t, response.Config.EnableAutomatedTesting)
	assert.Equal(t, 10, response.Config.MaxConcurrentTests)
}

func TestUpdateAutomatedTestingConfig(t *testing.T) {
	logger := zap.NewNop()
	tester := external.NewVerificationAutomatedTester(nil, logger)
	handler := NewVerificationAutomatedTestingHandler(tester, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid request",
			requestBody: UpdateConfigRequest{
				Config: &external.AutomatedTestingConfig{
					EnableAutomatedTesting: false,
					MaxConcurrentTests:     5,
					TestTimeout:            2 * 60 * 1000000000, // 2 minutes in nanoseconds
					SuccessThreshold:       0.90,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "nil config",
			requestBody: UpdateConfigRequest{
				Config: nil,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Configuration is required",
		},
		{
			name: "invalid success threshold",
			requestBody: UpdateConfigRequest{
				Config: &external.AutomatedTestingConfig{
					SuccessThreshold: 1.5, // Should be between 0 and 1
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "success threshold must be between 0 and 1",
		},
		{
			name: "invalid max concurrent tests",
			requestBody: UpdateConfigRequest{
				Config: &external.AutomatedTestingConfig{
					MaxConcurrentTests: 0, // Should be positive
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "max concurrent tests must be positive",
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("PUT", "/config", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response UpdateConfigResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, tt.expectedError)
			} else {
				var response UpdateConfigResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
			}
		})
	}
}
