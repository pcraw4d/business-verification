package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/external"
)

// VerificationAutomatedTestingHandler handles HTTP requests for automated testing functionality
type VerificationAutomatedTestingHandler struct {
	tester *external.VerificationAutomatedTester
	logger *zap.Logger
}

// NewVerificationAutomatedTestingHandler creates a new automated testing handler
func NewVerificationAutomatedTestingHandler(tester *external.VerificationAutomatedTester, logger *zap.Logger) *VerificationAutomatedTestingHandler {
	return &VerificationAutomatedTestingHandler{
		tester: tester,
		logger: logger,
	}
}

// RegisterRoutes registers the automated testing routes
func (h *VerificationAutomatedTestingHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/test-suites", h.CreateTestSuite).Methods("POST")
	router.HandleFunc("/test-suites", h.ListTestSuites).Methods("GET")
	router.HandleFunc("/test-suites/{suiteID}", h.GetTestSuite).Methods("GET")
	router.HandleFunc("/test-suites/{suiteID}/tests", h.AddTest).Methods("POST")
	router.HandleFunc("/test-suites/{suiteID}/run", h.RunTestSuite).Methods("POST")
	router.HandleFunc("/test-results", h.GetTestResults).Methods("GET")
	router.HandleFunc("/config", h.GetConfig).Methods("GET")
	router.HandleFunc("/config", h.UpdateConfig).Methods("PUT")
}

// CreateTestSuiteRequest represents a request to create a test suite
type CreateTestSuiteRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// CreateTestSuiteResponse represents a response for test suite creation
type CreateTestSuiteResponse struct {
	Success bool                `json:"success"`
	Suite   *external.TestSuite `json:"suite,omitempty"`
	Error   string              `json:"error,omitempty"`
}

// CreateTestSuite handles test suite creation
func (h *VerificationAutomatedTestingHandler) CreateTestSuite(w http.ResponseWriter, r *http.Request) {
	var req CreateTestSuiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := CreateTestSuiteResponse{
			Success: false,
			Error:   "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate request
	if req.Name == "" {
		response := CreateTestSuiteResponse{
			Success: false,
			Error:   "Test suite name is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.Category == "" {
		response := CreateTestSuiteResponse{
			Success: false,
			Error:   "Test suite category is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create test suite
	suite := &external.TestSuite{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Config:      req.Config,
	}

	if err := h.tester.CreateTestSuite(suite); err != nil {
		response := CreateTestSuiteResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := CreateTestSuiteResponse{
		Success: true,
		Suite:   suite,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ListTestSuitesResponse represents a response for listing test suites
type ListTestSuitesResponse struct {
	Success bool                  `json:"success"`
	Suites  []*external.TestSuite `json:"suites"`
	Error   string                `json:"error,omitempty"`
}

// ListTestSuites handles listing all test suites
func (h *VerificationAutomatedTestingHandler) ListTestSuites(w http.ResponseWriter, r *http.Request) {
	suites := h.tester.ListTestSuites()

	response := ListTestSuitesResponse{
		Success: true,
		Suites:  suites,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetTestSuiteResponse represents a response for getting a test suite
type GetTestSuiteResponse struct {
	Success bool                `json:"success"`
	Suite   *external.TestSuite `json:"suite,omitempty"`
	Error   string              `json:"error,omitempty"`
}

// GetTestSuite handles getting a specific test suite
func (h *VerificationAutomatedTestingHandler) GetTestSuite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	suiteID := vars["suiteID"]

	suite, err := h.tester.GetTestSuite(suiteID)
	if err != nil {
		response := GetTestSuiteResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := GetTestSuiteResponse{
		Success: true,
		Suite:   suite,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// AddTestRequest represents a request to add a test to a suite
type AddTestRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        external.TestType      `json:"type"`
	Input       interface{}            `json:"input,omitempty"`
	Expected    interface{}            `json:"expected,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Weight      float64                `json:"weight,omitempty"`
	Priority    external.TestPriority  `json:"priority,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
}

// AddTestResponse represents a response for adding a test
type AddTestResponse struct {
	Success bool                    `json:"success"`
	Test    *external.AutomatedTest `json:"test,omitempty"`
	Error   string                  `json:"error,omitempty"`
}

// AddTest handles adding a test to a test suite
func (h *VerificationAutomatedTestingHandler) AddTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	suiteID := vars["suiteID"]

	var req AddTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := AddTestResponse{
			Success: false,
			Error:   "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate request
	if req.Name == "" {
		response := AddTestResponse{
			Success: false,
			Error:   "Test name is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create test
	test := &external.AutomatedTest{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Input:       req.Input,
		Expected:    req.Expected,
		Metadata:    req.Metadata,
		Weight:      req.Weight,
		Priority:    req.Priority,
		Tags:        req.Tags,
	}

	if err := h.tester.AddTest(suiteID, test); err != nil {
		response := AddTestResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := AddTestResponse{
		Success: true,
		Test:    test,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// RunTestSuiteResponse represents a response for running a test suite
type RunTestSuiteResponse struct {
	Success bool                  `json:"success"`
	Summary *external.TestSummary `json:"summary,omitempty"`
	Error   string                `json:"error,omitempty"`
}

// RunTestSuite handles running a test suite
func (h *VerificationAutomatedTestingHandler) RunTestSuite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	suiteID := vars["suiteID"]

	ctx := r.Context()
	summary, err := h.tester.RunTestSuite(ctx, suiteID)
	if err != nil {
		response := RunTestSuiteResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := RunTestSuiteResponse{
		Success: true,
		Summary: summary,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetTestResultsResponse represents a response for getting test results
type GetTestResultsResponse struct {
	Success bool                   `json:"success"`
	Results []*external.TestResult `json:"results"`
	Error   string                 `json:"error,omitempty"`
}

// GetTestResults handles getting test results
func (h *VerificationAutomatedTestingHandler) GetTestResults(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	status := r.URL.Query().Get("status")

	limit := 0
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	var testStatus external.TestStatus
	if status != "" {
		testStatus = external.TestStatus(status)
	}

	results := h.tester.GetTestResults(limit, testStatus)

	response := GetTestResultsResponse{
		Success: true,
		Results: results,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetConfigResponse represents a response for getting configuration
type GetConfigResponse struct {
	Success bool                             `json:"success"`
	Config  *external.AutomatedTestingConfig `json:"config,omitempty"`
	Error   string                           `json:"error,omitempty"`
}

// GetConfig handles getting the current configuration
func (h *VerificationAutomatedTestingHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	config := h.tester.GetConfig()

	response := GetConfigResponse{
		Success: true,
		Config:  config,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateConfigRequest represents a request to update configuration
type UpdateConfigRequest struct {
	Config *external.AutomatedTestingConfig `json:"config"`
}

// UpdateConfigResponse represents a response for updating configuration
type UpdateConfigResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// UpdateConfig handles updating the configuration
func (h *VerificationAutomatedTestingHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := UpdateConfigResponse{
			Success: false,
			Error:   "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.Config == nil {
		response := UpdateConfigResponse{
			Success: false,
			Error:   "Configuration is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if err := h.tester.UpdateConfig(req.Config); err != nil {
		response := UpdateConfigResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := UpdateConfigResponse{
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
