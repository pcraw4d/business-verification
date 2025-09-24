package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"kyb-platform/internal/database"
)

// BulkOperationsService provides business logic for bulk merchant operations
type BulkOperationsService struct {
	db                database.Database
	logger            *log.Logger
	operations        map[string]*BulkOperation
	operationsMutex   sync.RWMutex
	merchantService   *MerchantPortfolioService
	auditService      *AuditService
	complianceService *ComplianceService
}

// NewBulkOperationsService creates a new bulk operations service
func NewBulkOperationsService(
	db database.Database,
	logger *log.Logger,
	merchantService *MerchantPortfolioService,
	auditService *AuditService,
	complianceService *ComplianceService,
) *BulkOperationsService {
	if logger == nil {
		logger = log.Default()
	}

	return &BulkOperationsService{
		db:                db,
		logger:            logger,
		operations:        make(map[string]*BulkOperation),
		merchantService:   merchantService,
		auditService:      auditService,
		complianceService: complianceService,
	}
}

// BulkOperation represents a bulk operation in progress
type BulkOperation struct {
	ID              string                    `json:"id"`
	Type            BulkOperationType         `json:"type"`
	Status          BulkOperationStatus       `json:"status"`
	UserID          string                    `json:"user_id"`
	MerchantIDs     []string                  `json:"merchant_ids"`
	TotalItems      int                       `json:"total_items"`
	ProcessedItems  int                       `json:"processed_items"`
	SuccessfulItems int                       `json:"successful_items"`
	FailedItems     int                       `json:"failed_items"`
	CurrentIndex    int                       `json:"current_index"`
	Results         []BulkOperationItemResult `json:"results"`
	Errors          []string                  `json:"errors"`
	Metadata        map[string]interface{}    `json:"metadata"`
	StartedAt       time.Time                 `json:"started_at"`
	PausedAt        *time.Time                `json:"paused_at,omitempty"`
	ResumedAt       *time.Time                `json:"resumed_at,omitempty"`
	CompletedAt     *time.Time                `json:"completed_at,omitempty"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
	mutex           sync.RWMutex
}

// BulkOperationType represents the type of bulk operation
type BulkOperationType string

const (
	BulkOperationTypeUpdatePortfolioType BulkOperationType = "update_portfolio_type"
	BulkOperationTypeUpdateRiskLevel     BulkOperationType = "update_risk_level"
	BulkOperationTypeUpdateStatus        BulkOperationType = "update_status"
	BulkOperationTypeBulkDelete          BulkOperationType = "bulk_delete"
	BulkOperationTypeBulkExport          BulkOperationType = "bulk_export"
	BulkOperationTypeComplianceCheck     BulkOperationType = "compliance_check"
)

// BulkOperationStatus represents the status of a bulk operation
type BulkOperationStatus string

const (
	BulkOperationStatusPending   BulkOperationStatus = "pending"
	BulkOperationStatusRunning   BulkOperationStatus = "running"
	BulkOperationStatusPaused    BulkOperationStatus = "paused"
	BulkOperationStatusCompleted BulkOperationStatus = "completed"
	BulkOperationStatusFailed    BulkOperationStatus = "failed"
	BulkOperationStatusCancelled BulkOperationStatus = "cancelled"
)

// BulkOperationItemResult represents a single item in a bulk operation
type BulkOperationItemResult struct {
	MerchantID  string                  `json:"merchant_id"`
	Status      BulkOperationItemStatus `json:"status"`
	Error       string                  `json:"error,omitempty"`
	Result      map[string]interface{}  `json:"result,omitempty"`
	ProcessedAt time.Time               `json:"processed_at"`
}

// BulkOperationItemStatus represents the status of a single item
type BulkOperationItemStatus string

const (
	BulkOperationItemStatusPending    BulkOperationItemStatus = "pending"
	BulkOperationItemStatusProcessing BulkOperationItemStatus = "processing"
	BulkOperationItemStatusSuccess    BulkOperationItemStatus = "success"
	BulkOperationItemStatusFailed     BulkOperationItemStatus = "failed"
	BulkOperationItemStatusSkipped    BulkOperationItemStatus = "skipped"
)

// BulkOperationRequest represents a request to start a bulk operation
type BulkOperationRequest struct {
	Type        BulkOperationType      `json:"type"`
	MerchantIDs []string               `json:"merchant_ids"`
	Parameters  map[string]interface{} `json:"parameters"`
	Options     BulkOperationOptions   `json:"options"`
}

// BulkOperationOptions represents options for bulk operations
type BulkOperationOptions struct {
	BatchSize           int           `json:"batch_size"`
	DelayBetweenBatches time.Duration `json:"delay_between_batches"`
	MaxConcurrency      int           `json:"max_concurrency"`
	ContinueOnError     bool          `json:"continue_on_error"`
	ValidateBefore      bool          `json:"validate_before"`
}

// BulkOperationProgress represents the progress of a bulk operation
type BulkOperationProgress struct {
	OperationID            string              `json:"operation_id"`
	Status                 BulkOperationStatus `json:"status"`
	TotalItems             int                 `json:"total_items"`
	ProcessedItems         int                 `json:"processed_items"`
	SuccessfulItems        int                 `json:"successful_items"`
	FailedItems            int                 `json:"failed_items"`
	ProgressPercentage     float64             `json:"progress_percentage"`
	CurrentItem            string              `json:"current_item,omitempty"`
	EstimatedTimeRemaining *time.Duration      `json:"estimated_time_remaining,omitempty"`
	StartedAt              time.Time           `json:"started_at"`
	LastUpdatedAt          time.Time           `json:"last_updated_at"`
}

// Common errors
var (
	ErrBulkOperationInvalidType    = errors.New("invalid bulk operation type")
	ErrBulkOperationNotRunning     = errors.New("bulk operation is not running")
	ErrBulkOperationAlreadyRunning = errors.New("bulk operation is already running")
	ErrBulkOperationNotPaused      = errors.New("bulk operation is not paused")
	ErrBulkOperationEmptyList      = errors.New("merchant list cannot be empty")
	ErrBulkOperationTooLarge       = errors.New("merchant list too large for bulk operation")
)

// =============================================================================
// Bulk Operation Management
// =============================================================================

// StartBulkOperation starts a new bulk operation
func (s *BulkOperationsService) StartBulkOperation(ctx context.Context, req *BulkOperationRequest, userID string) (*BulkOperation, error) {
	s.logger.Printf("Starting bulk operation: %s for %d merchants", req.Type, len(req.MerchantIDs))

	// Validate request
	if err := s.validateBulkOperationRequest(req); err != nil {
		return nil, fmt.Errorf("invalid bulk operation request: %w", err)
	}

	// Check if user has any running operations
	if s.hasRunningOperation(userID) {
		return nil, ErrBulkOperationAlreadyRunning
	}

	// Create operation
	operation := &BulkOperation{
		ID:              s.generateOperationID(),
		Type:            req.Type,
		Status:          BulkOperationStatusPending,
		UserID:          userID,
		MerchantIDs:     req.MerchantIDs,
		TotalItems:      len(req.MerchantIDs),
		ProcessedItems:  0,
		SuccessfulItems: 0,
		FailedItems:     0,
		CurrentIndex:    0,
		Results:         make([]BulkOperationItemResult, len(req.MerchantIDs)),
		Errors:          []string{},
		Metadata:        req.Parameters,
		StartedAt:       time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Initialize results
	for i, merchantID := range req.MerchantIDs {
		operation.Results[i] = BulkOperationItemResult{
			MerchantID: merchantID,
			Status:     BulkOperationItemStatusPending,
		}
	}

	// Store operation
	s.operationsMutex.Lock()
	s.operations[operation.ID] = operation
	s.operationsMutex.Unlock()

	// Log audit event
	if err := s.auditService.LogMerchantOperation(ctx, &LogMerchantOperationRequest{
		UserID:       userID,
		MerchantID:   "", // Bulk operation affects multiple merchants
		Action:       "START_BULK_OPERATION",
		ResourceType: "bulk_operation",
		ResourceID:   operation.ID,
		Details:      fmt.Sprintf("Started bulk operation %s for %d merchants", req.Type, len(req.MerchantIDs)),
		Description:  fmt.Sprintf("Bulk operation %s initiated", req.Type),
		Metadata: map[string]interface{}{
			"operation_type": req.Type,
			"merchant_count": len(req.MerchantIDs),
		},
	}); err != nil {
		s.logger.Printf("Warning: failed to log audit event: %v", err)
	}

	// Start processing in background
	go s.processBulkOperation(ctx, operation, req.Options)

	s.logger.Printf("Bulk operation started: %s", operation.ID)
	return operation, nil
}

// GetBulkOperation retrieves a bulk operation by ID
func (s *BulkOperationsService) GetBulkOperation(ctx context.Context, operationID string) (*BulkOperation, error) {
	s.operationsMutex.RLock()
	defer s.operationsMutex.RUnlock()

	operation, exists := s.operations[operationID]
	if !exists {
		return nil, ErrBulkOperationNotFound
	}

	// Return a copy to avoid race conditions
	operation.mutex.RLock()
	defer operation.mutex.RUnlock()

	// Create a deep copy
	operationCopy := *operation
	operationCopy.Results = make([]BulkOperationItemResult, len(operation.Results))
	copy(operationCopy.Results, operation.Results)
	operationCopy.Errors = make([]string, len(operation.Errors))
	copy(operationCopy.Errors, operation.Errors)

	return &operationCopy, nil
}

// GetBulkOperationProgress retrieves the progress of a bulk operation
func (s *BulkOperationsService) GetBulkOperationProgress(ctx context.Context, operationID string) (*BulkOperationProgress, error) {
	operation, err := s.GetBulkOperation(ctx, operationID)
	if err != nil {
		return nil, err
	}

	operation.mutex.RLock()
	defer operation.mutex.RUnlock()

	progress := &BulkOperationProgress{
		OperationID:        operation.ID,
		Status:             operation.Status,
		TotalItems:         operation.TotalItems,
		ProcessedItems:     operation.ProcessedItems,
		SuccessfulItems:    operation.SuccessfulItems,
		FailedItems:        operation.FailedItems,
		ProgressPercentage: 0,
		StartedAt:          operation.StartedAt,
		LastUpdatedAt:      operation.UpdatedAt,
	}

	// Calculate progress percentage
	if operation.TotalItems > 0 {
		progress.ProgressPercentage = float64(operation.ProcessedItems) / float64(operation.TotalItems) * 100
	}

	// Set current item if processing
	if operation.Status == BulkOperationStatusRunning && operation.CurrentIndex < len(operation.MerchantIDs) {
		progress.CurrentItem = operation.MerchantIDs[operation.CurrentIndex]
	}

	// Calculate estimated time remaining
	if operation.ProcessedItems > 0 && operation.Status == BulkOperationStatusRunning {
		elapsed := time.Since(operation.StartedAt)
		rate := float64(operation.ProcessedItems) / elapsed.Seconds()
		remaining := operation.TotalItems - operation.ProcessedItems
		if rate > 0 {
			estimated := time.Duration(float64(remaining)/rate) * time.Second
			progress.EstimatedTimeRemaining = &estimated
		}
	}

	return progress, nil
}

// PauseBulkOperation pauses a running bulk operation
func (s *BulkOperationsService) PauseBulkOperation(ctx context.Context, operationID, userID string) error {
	s.logger.Printf("Pausing bulk operation: %s", operationID)

	s.operationsMutex.Lock()
	defer s.operationsMutex.Unlock()

	operation, exists := s.operations[operationID]
	if !exists {
		return ErrBulkOperationNotFound
	}

	operation.mutex.Lock()
	defer operation.mutex.Unlock()

	if operation.Status != BulkOperationStatusRunning {
		return ErrBulkOperationNotRunning
	}

	// Update operation status
	operation.Status = BulkOperationStatusPaused
	now := time.Now()
	operation.PausedAt = &now
	operation.UpdatedAt = now

	// Log audit event
	if err := s.auditService.LogMerchantOperation(ctx, &LogMerchantOperationRequest{
		UserID:       userID,
		MerchantID:   "",
		Action:       "PAUSE_BULK_OPERATION",
		ResourceType: "bulk_operation",
		ResourceID:   operationID,
		Details:      "Bulk operation paused",
		Description:  "Bulk operation paused by user",
	}); err != nil {
		s.logger.Printf("Warning: failed to log audit event: %v", err)
	}

	s.logger.Printf("Bulk operation paused: %s", operationID)
	return nil
}

// ResumeBulkOperation resumes a paused bulk operation
func (s *BulkOperationsService) ResumeBulkOperation(ctx context.Context, operationID, userID string) error {
	s.logger.Printf("Resuming bulk operation: %s", operationID)

	s.operationsMutex.Lock()
	defer s.operationsMutex.Unlock()

	operation, exists := s.operations[operationID]
	if !exists {
		return ErrBulkOperationNotFound
	}

	operation.mutex.Lock()
	defer operation.mutex.Unlock()

	if operation.Status != BulkOperationStatusPaused {
		return ErrBulkOperationNotPaused
	}

	// Update operation status
	operation.Status = BulkOperationStatusRunning
	now := time.Now()
	operation.ResumedAt = &now
	operation.UpdatedAt = now

	// Log audit event
	if err := s.auditService.LogMerchantOperation(ctx, &LogMerchantOperationRequest{
		UserID:       userID,
		MerchantID:   "",
		Action:       "RESUME_BULK_OPERATION",
		ResourceType: "bulk_operation",
		ResourceID:   operationID,
		Details:      "Bulk operation resumed",
		Description:  "Bulk operation resumed by user",
	}); err != nil {
		s.logger.Printf("Warning: failed to log audit event: %v", err)
	}

	// Resume processing in background
	go s.processBulkOperation(ctx, operation, BulkOperationOptions{})

	s.logger.Printf("Bulk operation resumed: %s", operationID)
	return nil
}

// CancelBulkOperation cancels a bulk operation
func (s *BulkOperationsService) CancelBulkOperation(ctx context.Context, operationID, userID string) error {
	s.logger.Printf("Cancelling bulk operation: %s", operationID)

	s.operationsMutex.Lock()
	defer s.operationsMutex.Unlock()

	operation, exists := s.operations[operationID]
	if !exists {
		return ErrBulkOperationNotFound
	}

	operation.mutex.Lock()
	defer operation.mutex.Unlock()

	if operation.Status == BulkOperationStatusCompleted || operation.Status == BulkOperationStatusFailed {
		return fmt.Errorf("cannot cancel completed operation")
	}

	// Update operation status
	operation.Status = BulkOperationStatusCancelled
	now := time.Now()
	operation.CompletedAt = &now
	operation.UpdatedAt = now

	// Log audit event
	if err := s.auditService.LogMerchantOperation(ctx, &LogMerchantOperationRequest{
		UserID:       userID,
		MerchantID:   "",
		Action:       "CANCEL_BULK_OPERATION",
		ResourceType: "bulk_operation",
		ResourceID:   operationID,
		Details:      "Bulk operation cancelled",
		Description:  "Bulk operation cancelled by user",
	}); err != nil {
		s.logger.Printf("Warning: failed to log audit event: %v", err)
	}

	s.logger.Printf("Bulk operation cancelled: %s", operationID)
	return nil
}

// ListBulkOperations lists bulk operations for a user
func (s *BulkOperationsService) ListBulkOperations(ctx context.Context, userID string, limit int) ([]*BulkOperation, error) {
	s.operationsMutex.RLock()
	defer s.operationsMutex.RUnlock()

	var userOperations []*BulkOperation
	count := 0

	for _, operation := range s.operations {
		if operation.UserID == userID {
			operation.mutex.RLock()
			operationCopy := *operation
			operationCopy.Results = make([]BulkOperationItemResult, len(operation.Results))
			copy(operationCopy.Results, operation.Results)
			operationCopy.Errors = make([]string, len(operation.Errors))
			copy(operationCopy.Errors, operation.Errors)
			operation.mutex.RUnlock()

			userOperations = append(userOperations, &operationCopy)
			count++

			if limit > 0 && count >= limit {
				break
			}
		}
	}

	return userOperations, nil
}

// =============================================================================
// Bulk Operation Processing
// =============================================================================

// processBulkOperation processes a bulk operation
func (s *BulkOperationsService) processBulkOperation(ctx context.Context, operation *BulkOperation, options BulkOperationOptions) {
	s.logger.Printf("Processing bulk operation: %s", operation.ID)

	// Set default options
	if options.BatchSize <= 0 {
		options.BatchSize = 10
	}
	if options.DelayBetweenBatches <= 0 {
		options.DelayBetweenBatches = 100 * time.Millisecond
	}
	if options.MaxConcurrency <= 0 {
		options.MaxConcurrency = 5
	}

	// Update status to running
	operation.mutex.Lock()
	operation.Status = BulkOperationStatusRunning
	operation.UpdatedAt = time.Now()
	operation.mutex.Unlock()

	// Process merchants in batches
	for i := operation.CurrentIndex; i < len(operation.MerchantIDs); i++ {
		// Check if operation was paused or cancelled
		operation.mutex.RLock()
		status := operation.Status
		operation.mutex.RUnlock()

		if status == BulkOperationStatusPaused {
			s.logger.Printf("Bulk operation paused: %s", operation.ID)
			return
		}

		if status == BulkOperationStatusCancelled {
			s.logger.Printf("Bulk operation cancelled: %s", operation.ID)
			return
		}

		// Process current merchant
		merchantID := operation.MerchantIDs[i]
		s.processMerchant(ctx, operation, i, merchantID)

		// Update current index
		operation.mutex.Lock()
		operation.CurrentIndex = i + 1
		operation.UpdatedAt = time.Now()
		operation.mutex.Unlock()

		// Add delay between items if specified
		if options.DelayBetweenBatches > 0 {
			time.Sleep(options.DelayBetweenBatches)
		}
	}

	// Mark operation as completed
	operation.mutex.Lock()
	operation.Status = BulkOperationStatusCompleted
	now := time.Now()
	operation.CompletedAt = &now
	operation.UpdatedAt = now
	operation.mutex.Unlock()

	// Log audit event
	if err := s.auditService.LogMerchantOperation(ctx, &LogMerchantOperationRequest{
		UserID:       operation.UserID,
		MerchantID:   "",
		Action:       "COMPLETE_BULK_OPERATION",
		ResourceType: "bulk_operation",
		ResourceID:   operation.ID,
		Details:      fmt.Sprintf("Bulk operation completed: %d successful, %d failed", operation.SuccessfulItems, operation.FailedItems),
		Description:  "Bulk operation completed successfully",
		Metadata: map[string]interface{}{
			"successful_items": operation.SuccessfulItems,
			"failed_items":     operation.FailedItems,
		},
	}); err != nil {
		s.logger.Printf("Warning: failed to log audit event: %v", err)
	}

	s.logger.Printf("Bulk operation completed: %s (%d successful, %d failed)", operation.ID, operation.SuccessfulItems, operation.FailedItems)
}

// processMerchant processes a single merchant in a bulk operation
func (s *BulkOperationsService) processMerchant(ctx context.Context, operation *BulkOperation, index int, merchantID string) {
	s.logger.Printf("Processing merchant %s in operation %s", merchantID, operation.ID)

	// Update item status to processing
	operation.mutex.Lock()
	operation.Results[index].Status = BulkOperationItemStatusProcessing
	operation.Results[index].ProcessedAt = time.Now()
	operation.mutex.Unlock()

	// Process based on operation type
	var err error
	var result map[string]interface{}

	switch operation.Type {
	case BulkOperationTypeUpdatePortfolioType:
		err, result = s.processUpdatePortfolioType(ctx, operation, merchantID)
	case BulkOperationTypeUpdateRiskLevel:
		err, result = s.processUpdateRiskLevel(ctx, operation, merchantID)
	case BulkOperationTypeUpdateStatus:
		err, result = s.processUpdateStatus(ctx, operation, merchantID)
	case BulkOperationTypeBulkDelete:
		err, result = s.processBulkDelete(ctx, operation, merchantID)
	case BulkOperationTypeComplianceCheck:
		err, result = s.processComplianceCheck(ctx, operation, merchantID)
	default:
		err = fmt.Errorf("unsupported operation type: %s", operation.Type)
	}

	// Update item result
	operation.mutex.Lock()
	if err != nil {
		operation.Results[index].Status = BulkOperationItemStatusFailed
		operation.Results[index].Error = err.Error()
		operation.FailedItems++
		operation.Errors = append(operation.Errors, fmt.Sprintf("Merchant %s: %v", merchantID, err))
	} else {
		operation.Results[index].Status = BulkOperationItemStatusSuccess
		operation.Results[index].Result = result
		operation.SuccessfulItems++
	}
	operation.ProcessedItems++
	operation.UpdatedAt = time.Now()
	operation.mutex.Unlock()
}

// =============================================================================
// Operation Type Handlers
// =============================================================================

// processUpdatePortfolioType processes portfolio type update for a merchant
func (s *BulkOperationsService) processUpdatePortfolioType(ctx context.Context, operation *BulkOperation, merchantID string) (error, map[string]interface{}) {
	portfolioTypeStr, ok := operation.Metadata["portfolio_type"].(string)
	if !ok {
		return fmt.Errorf("portfolio_type parameter is required"), nil
	}

	portfolioType := PortfolioType(portfolioTypeStr)
	if !s.merchantService.isValidPortfolioType(portfolioType) {
		return fmt.Errorf("invalid portfolio type: %s", portfolioTypeStr), nil
	}

	err := s.merchantService.UpdateMerchantPortfolioType(ctx, merchantID, portfolioType, operation.UserID)
	if err != nil {
		return err, nil
	}

	return nil, map[string]interface{}{
		"portfolio_type": string(portfolioType),
		"updated_at":     time.Now(),
	}
}

// processUpdateRiskLevel processes risk level update for a merchant
func (s *BulkOperationsService) processUpdateRiskLevel(ctx context.Context, operation *BulkOperation, merchantID string) (error, map[string]interface{}) {
	riskLevelStr, ok := operation.Metadata["risk_level"].(string)
	if !ok {
		return fmt.Errorf("risk_level parameter is required"), nil
	}

	riskLevel := RiskLevel(riskLevelStr)
	if !s.merchantService.isValidRiskLevel(riskLevel) {
		return fmt.Errorf("invalid risk level: %s", riskLevelStr), nil
	}

	err := s.merchantService.UpdateMerchantRiskLevel(ctx, merchantID, riskLevel, operation.UserID)
	if err != nil {
		return err, nil
	}

	return nil, map[string]interface{}{
		"risk_level": string(riskLevel),
		"updated_at": time.Now(),
	}
}

// processUpdateStatus processes status update for a merchant
func (s *BulkOperationsService) processUpdateStatus(ctx context.Context, operation *BulkOperation, merchantID string) (error, map[string]interface{}) {
	status, ok := operation.Metadata["status"].(string)
	if !ok {
		return fmt.Errorf("status parameter is required"), nil
	}

	// Get existing merchant
	merchant, err := s.merchantService.GetMerchant(ctx, merchantID)
	if err != nil {
		return err, nil
	}

	// Update status
	merchant.Status = status
	merchant.UpdatedAt = time.Now()

	_, err = s.merchantService.UpdateMerchant(ctx, merchant, operation.UserID)
	if err != nil {
		return err, nil
	}

	return nil, map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
}

// processBulkDelete processes bulk delete for a merchant
func (s *BulkOperationsService) processBulkDelete(ctx context.Context, operation *BulkOperation, merchantID string) (error, map[string]interface{}) {
	err := s.merchantService.DeleteMerchant(ctx, merchantID, operation.UserID)
	if err != nil {
		return err, nil
	}

	return nil, map[string]interface{}{
		"deleted_at": time.Now(),
	}
}

// processComplianceCheck processes compliance check for a merchant
func (s *BulkOperationsService) processComplianceCheck(ctx context.Context, operation *BulkOperation, merchantID string) (error, map[string]interface{}) {
	// Get merchant
	merchant, err := s.merchantService.GetMerchant(ctx, merchantID)
	if err != nil {
		return err, nil
	}

	// Perform compliance check
	complianceResult, err := s.complianceService.ValidateMerchantCompliance(ctx, merchant.ID)
	if err != nil {
		return err, nil
	}

	return nil, map[string]interface{}{
		"compliance_status": complianceResult.OverallStatus,
		"compliance_score":  complianceResult.ComplianceScore,
		"checked_at":        time.Now(),
	}
}

// =============================================================================
// Helper Methods
// =============================================================================

// validateBulkOperationRequest validates a bulk operation request
func (s *BulkOperationsService) validateBulkOperationRequest(req *BulkOperationRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	if req.Type == "" {
		return ErrBulkOperationInvalidType
	}

	if len(req.MerchantIDs) == 0 {
		return ErrBulkOperationEmptyList
	}

	if len(req.MerchantIDs) > 10000 {
		return ErrBulkOperationTooLarge
	}

	// Validate operation type
	switch req.Type {
	case BulkOperationTypeUpdatePortfolioType:
		if _, ok := req.Parameters["portfolio_type"].(string); !ok {
			return errors.New("portfolio_type parameter is required")
		}
	case BulkOperationTypeUpdateRiskLevel:
		if _, ok := req.Parameters["risk_level"].(string); !ok {
			return errors.New("risk_level parameter is required")
		}
	case BulkOperationTypeUpdateStatus:
		if _, ok := req.Parameters["status"].(string); !ok {
			return errors.New("status parameter is required")
		}
	case BulkOperationTypeBulkDelete, BulkOperationTypeComplianceCheck:
		// No additional parameters required
	default:
		return ErrBulkOperationInvalidType
	}

	return nil
}

// hasRunningOperation checks if user has any running operations
func (s *BulkOperationsService) hasRunningOperation(userID string) bool {
	s.operationsMutex.RLock()
	defer s.operationsMutex.RUnlock()

	for _, operation := range s.operations {
		if operation.UserID == userID && operation.Status == BulkOperationStatusRunning {
			return true
		}
	}

	return false
}

// generateOperationID generates a unique operation ID
func (s *BulkOperationsService) generateOperationID() string {
	return fmt.Sprintf("bulk_op_%d_%d", time.Now().UnixNano(), len(s.operations))
}

// CleanupCompletedOperations removes old completed operations
func (s *BulkOperationsService) CleanupCompletedOperations(maxAge time.Duration) {
	s.operationsMutex.Lock()
	defer s.operationsMutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for id, operation := range s.operations {
		if (operation.Status == BulkOperationStatusCompleted ||
			operation.Status == BulkOperationStatusFailed ||
			operation.Status == BulkOperationStatusCancelled) &&
			operation.UpdatedAt.Before(cutoff) {
			delete(s.operations, id)
			s.logger.Printf("Cleaned up old bulk operation: %s", id)
		}
	}
}
