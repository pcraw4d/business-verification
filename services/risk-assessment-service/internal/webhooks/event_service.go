package webhooks

import (
	"context"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// EventService handles webhook event triggering for risk assessment operations
type EventService struct {
	webhookManager WebhookManager
	logger         *zap.Logger
}

// NewEventService creates a new webhook event service
func NewEventService(webhookManager WebhookManager, logger *zap.Logger) *EventService {
	return &EventService{
		webhookManager: webhookManager,
		logger:         logger,
	}
}

// TriggerRiskAssessmentStarted triggers a webhook event when risk assessment starts
func (es *EventService) TriggerRiskAssessmentStarted(ctx context.Context, tenantID, businessID string, req *models.RiskAssessmentRequest) error {
	eventData := &WebhookEventData{
		ID:         generateEventID(),
		Type:       EventRiskAssessmentStarted,
		TenantID:   tenantID,
		BusinessID: businessID,
		Data: map[string]interface{}{
			"business_name":      req.BusinessName,
			"business_address":   req.BusinessAddress,
			"industry":           req.Industry,
			"country":            req.Country,
			"status":             "started",
			"prediction_horizon": req.PredictionHorizon,
		},
		Timestamp: time.Now(),
		Source:    "risk_assessment_service",
		Version:   "1.0",
		Metadata: map[string]interface{}{
			"model_type": req.ModelType,
		},
	}

	return es.webhookManager.ProcessEvent(ctx, eventData)
}

// TriggerRiskAssessmentCompleted triggers a webhook event when risk assessment completes
func (es *EventService) TriggerRiskAssessmentCompleted(ctx context.Context, tenantID, businessID string, assessment *models.RiskAssessment) error {
	eventData := &WebhookEventData{
		ID:         generateEventID(),
		Type:       EventRiskAssessmentCompleted,
		TenantID:   tenantID,
		BusinessID: businessID,
		Data: map[string]interface{}{
			"assessment_id":      assessment.ID,
			"business_name":      assessment.BusinessName,
			"business_address":   assessment.BusinessAddress,
			"industry":           assessment.Industry,
			"country":            assessment.Country,
			"status":             string(assessment.Status),
			"risk_score":         assessment.RiskScore,
			"risk_level":         string(assessment.RiskLevel),
			"confidence_score":   assessment.ConfidenceScore,
			"prediction_horizon": assessment.PredictionHorizon,
			"risk_factors":       assessment.RiskFactors,
			"created_at":         assessment.CreatedAt,
			"updated_at":         assessment.UpdatedAt,
		},
		Timestamp: time.Now(),
		Source:    "risk_assessment_service",
		Version:   "1.0",
		Metadata: map[string]interface{}{
			"assessment_id": assessment.ID,
			"model_type":    assessment.Metadata["model_type"],
		},
	}

	return es.webhookManager.ProcessEvent(ctx, eventData)
}

// TriggerRiskAssessmentFailed triggers a webhook event when risk assessment fails
func (es *EventService) TriggerRiskAssessmentFailed(ctx context.Context, tenantID, businessID string, req *models.RiskAssessmentRequest, err error) error {
	eventData := &WebhookEventData{
		ID:         generateEventID(),
		Type:       EventRiskAssessmentFailed,
		TenantID:   tenantID,
		BusinessID: businessID,
		Data: map[string]interface{}{
			"business_name":      req.BusinessName,
			"business_address":   req.BusinessAddress,
			"industry":           req.Industry,
			"country":            req.Country,
			"status":             "failed",
			"error":              err.Error(),
			"prediction_horizon": req.PredictionHorizon,
		},
		Timestamp: time.Now(),
		Source:    "risk_assessment_service",
		Version:   "1.0",
		Metadata: map[string]interface{}{
			"model_type": req.ModelType,
			"error_type": "assessment_failed",
		},
	}

	return es.webhookManager.ProcessEvent(ctx, eventData)
}

// TriggerRiskPredictionCompleted triggers a webhook event when risk prediction completes
func (es *EventService) TriggerRiskPredictionCompleted(ctx context.Context, tenantID, businessID string, prediction *models.RiskPrediction) error {
	eventData := &WebhookEventData{
		ID:         generateEventID(),
		Type:       EventRiskPredictionCompleted,
		TenantID:   tenantID,
		BusinessID: businessID,
		Data: map[string]interface{}{
			"business_id":      prediction.BusinessID,
			"predicted_score":  prediction.PredictedScore,
			"predicted_level":  string(prediction.PredictedLevel),
			"confidence_score": prediction.ConfidenceScore,
			"horizon_months":   prediction.HorizonMonths,
			"prediction_date":  prediction.PredictionDate,
			"created_at":       prediction.CreatedAt,
		},
		Timestamp: time.Now(),
		Source:    "risk_assessment_service",
		Version:   "1.0",
		Metadata: map[string]interface{}{
			"business_id": prediction.BusinessID,
		},
	}

	return es.webhookManager.ProcessEvent(ctx, eventData)
}

// TriggerRiskPredictionFailed triggers a webhook event when risk prediction fails
func (es *EventService) TriggerRiskPredictionFailed(ctx context.Context, tenantID, businessID string, req *models.RiskAssessmentRequest, horizonMonths int, err error) error {
	eventData := &WebhookEventData{
		ID:         generateEventID(),
		Type:       EventRiskPredictionFailed,
		TenantID:   tenantID,
		BusinessID: businessID,
		Data: map[string]interface{}{
			"business_name":    req.BusinessName,
			"business_address": req.BusinessAddress,
			"industry":         req.Industry,
			"country":          req.Country,
			"horizon_months":   horizonMonths,
			"error":            err.Error(),
		},
		Timestamp: time.Now(),
		Source:    "risk_assessment_service",
		Version:   "1.0",
		Metadata: map[string]interface{}{
			"model_type": req.ModelType,
			"error_type": "prediction_failed",
		},
	}

	return es.webhookManager.ProcessEvent(ctx, eventData)
}

// TriggerBatchJobStarted triggers a webhook event when batch job starts
func (es *EventService) TriggerBatchJobStarted(ctx context.Context, tenantID string, jobID string, requestCount int) error {
	eventData := &WebhookEventData{
		ID:       generateEventID(),
		Type:     EventBatchJobStarted,
		TenantID: tenantID,
		Data: map[string]interface{}{
			"job_id":        jobID,
			"request_count": requestCount,
			"status":        "started",
			"started_at":    time.Now(),
		},
		Timestamp: time.Now(),
		Source:    "risk_assessment_service",
		Version:   "1.0",
		Metadata: map[string]interface{}{
			"job_id": jobID,
		},
	}

	return es.webhookManager.ProcessEvent(ctx, eventData)
}

// TriggerBatchJobCompleted triggers a webhook event when batch job completes
func (es *EventService) TriggerBatchJobCompleted(ctx context.Context, tenantID string, jobID string, results map[string]interface{}) error {
	eventData := &WebhookEventData{
		ID:       generateEventID(),
		Type:     EventBatchJobCompleted,
		TenantID: tenantID,
		Data: map[string]interface{}{
			"job_id":       jobID,
			"status":       "completed",
			"completed_at": time.Now(),
			"results":      results,
		},
		Timestamp: time.Now(),
		Source:    "risk_assessment_service",
		Version:   "1.0",
		Metadata: map[string]interface{}{
			"job_id": jobID,
		},
	}

	return es.webhookManager.ProcessEvent(ctx, eventData)
}

// TriggerBatchJobFailed triggers a webhook event when batch job fails
func (es *EventService) TriggerBatchJobFailed(ctx context.Context, tenantID string, jobID string, err error) error {
	eventData := &WebhookEventData{
		ID:       generateEventID(),
		Type:     EventBatchJobFailed,
		TenantID: tenantID,
		Data: map[string]interface{}{
			"job_id":    jobID,
			"status":    "failed",
			"failed_at": time.Now(),
			"error":     err.Error(),
		},
		Timestamp: time.Now(),
		Source:    "risk_assessment_service",
		Version:   "1.0",
		Metadata: map[string]interface{}{
			"job_id":     jobID,
			"error_type": "batch_job_failed",
		},
	}

	return es.webhookManager.ProcessEvent(ctx, eventData)
}

// TriggerBatchJobProgress triggers a webhook event for batch job progress updates
func (es *EventService) TriggerBatchJobProgress(ctx context.Context, tenantID string, jobID string, progress int, total int) error {
	eventData := &WebhookEventData{
		ID:       generateEventID(),
		Type:     EventBatchJobProgress,
		TenantID: tenantID,
		Data: map[string]interface{}{
			"job_id":     jobID,
			"progress":   progress,
			"total":      total,
			"percentage": float64(progress) / float64(total) * 100,
			"updated_at": time.Now(),
		},
		Timestamp: time.Now(),
		Source:    "risk_assessment_service",
		Version:   "1.0",
		Metadata: map[string]interface{}{
			"job_id": jobID,
		},
	}

	return es.webhookManager.ProcessEvent(ctx, eventData)
}

// TriggerCustomModelCreated triggers a webhook event when custom model is created
func (es *EventService) TriggerCustomModelCreated(ctx context.Context, tenantID string, modelID string, modelName string) error {
	eventData := &WebhookEventData{
		ID:       generateEventID(),
		Type:     EventCustomModelCreated,
		TenantID: tenantID,
		Data: map[string]interface{}{
			"model_id":   modelID,
			"model_name": modelName,
			"status":     "created",
			"created_at": time.Now(),
		},
		Timestamp: time.Now(),
		Source:    "risk_assessment_service",
		Version:   "1.0",
		Metadata: map[string]interface{}{
			"model_id": modelID,
		},
	}

	return es.webhookManager.ProcessEvent(ctx, eventData)
}

// TriggerCustomModelUpdated triggers a webhook event when custom model is updated
func (es *EventService) TriggerCustomModelUpdated(ctx context.Context, tenantID string, modelID string, modelName string) error {
	eventData := &WebhookEventData{
		ID:       generateEventID(),
		Type:     EventCustomModelUpdated,
		TenantID: tenantID,
		Data: map[string]interface{}{
			"model_id":   modelID,
			"model_name": modelName,
			"status":     "updated",
			"updated_at": time.Now(),
		},
		Timestamp: time.Now(),
		Source:    "risk_assessment_service",
		Version:   "1.0",
		Metadata: map[string]interface{}{
			"model_id": modelID,
		},
	}

	return es.webhookManager.ProcessEvent(ctx, eventData)
}

// TriggerCustomModelDeleted triggers a webhook event when custom model is deleted
func (es *EventService) TriggerCustomModelDeleted(ctx context.Context, tenantID string, modelID string, modelName string) error {
	eventData := &WebhookEventData{
		ID:       generateEventID(),
		Type:     EventCustomModelDeleted,
		TenantID: tenantID,
		Data: map[string]interface{}{
			"model_id":   modelID,
			"model_name": modelName,
			"status":     "deleted",
			"deleted_at": time.Now(),
		},
		Timestamp: time.Now(),
		Source:    "risk_assessment_service",
		Version:   "1.0",
		Metadata: map[string]interface{}{
			"model_id": modelID,
		},
	}

	return es.webhookManager.ProcessEvent(ctx, eventData)
}

// TriggerCustomModelValidated triggers a webhook event when custom model is validated
func (es *EventService) TriggerCustomModelValidated(ctx context.Context, tenantID string, modelID string, modelName string, validationResult map[string]interface{}) error {
	eventData := &WebhookEventData{
		ID:       generateEventID(),
		Type:     EventCustomModelValidated,
		TenantID: tenantID,
		Data: map[string]interface{}{
			"model_id":          modelID,
			"model_name":        modelName,
			"status":            "validated",
			"validated_at":      time.Now(),
			"validation_result": validationResult,
		},
		Timestamp: time.Now(),
		Source:    "risk_assessment_service",
		Version:   "1.0",
		Metadata: map[string]interface{}{
			"model_id": modelID,
		},
	}

	return es.webhookManager.ProcessEvent(ctx, eventData)
}

// TriggerReportGenerated triggers a webhook event when report is generated
func (es *EventService) TriggerReportGenerated(ctx context.Context, tenantID string, reportID string, reportType string) error {
	eventData := &WebhookEventData{
		ID:       generateEventID(),
		Type:     EventReportGenerated,
		TenantID: tenantID,
		Data: map[string]interface{}{
			"report_id":    reportID,
			"report_type":  reportType,
			"status":       "generated",
			"generated_at": time.Now(),
		},
		Timestamp: time.Now(),
		Source:    "risk_assessment_service",
		Version:   "1.0",
		Metadata: map[string]interface{}{
			"report_id": reportID,
		},
	}

	return es.webhookManager.ProcessEvent(ctx, eventData)
}

// TriggerReportScheduled triggers a webhook event when report is scheduled
func (es *EventService) TriggerReportScheduled(ctx context.Context, tenantID string, reportID string, scheduleType string) error {
	eventData := &WebhookEventData{
		ID:       generateEventID(),
		Type:     EventReportScheduled,
		TenantID: tenantID,
		Data: map[string]interface{}{
			"report_id":     reportID,
			"schedule_type": scheduleType,
			"status":        "scheduled",
			"scheduled_at":  time.Now(),
		},
		Timestamp: time.Now(),
		Source:    "risk_assessment_service",
		Version:   "1.0",
		Metadata: map[string]interface{}{
			"report_id": reportID,
		},
	}

	return es.webhookManager.ProcessEvent(ctx, eventData)
}

// TriggerReportFailed triggers a webhook event when report generation fails
func (es *EventService) TriggerReportFailed(ctx context.Context, tenantID string, reportID string, reportType string, err error) error {
	eventData := &WebhookEventData{
		ID:       generateEventID(),
		Type:     EventReportFailed,
		TenantID: tenantID,
		Data: map[string]interface{}{
			"report_id":   reportID,
			"report_type": reportType,
			"status":      "failed",
			"failed_at":   time.Now(),
			"error":       err.Error(),
		},
		Timestamp: time.Now(),
		Source:    "risk_assessment_service",
		Version:   "1.0",
		Metadata: map[string]interface{}{
			"report_id":  reportID,
			"error_type": "report_failed",
		},
	}

	return es.webhookManager.ProcessEvent(ctx, eventData)
}
