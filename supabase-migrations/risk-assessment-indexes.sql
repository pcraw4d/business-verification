-- Risk Assessment Service Performance Indexes Migration
-- This migration adds optimized indexes for high-performance queries

-- Composite indexes for common query patterns

-- Risk Assessments - Dashboard queries (tenant_id + risk_level + created_at)
CREATE INDEX CONCURRENTLY idx_risk_assessments_tenant_risk_created 
ON risk_assessments(tenant_id, risk_level, created_at DESC);

-- Risk Assessments - Business lookup with tenant isolation
CREATE INDEX CONCURRENTLY idx_risk_assessments_tenant_business 
ON risk_assessments(tenant_id, business_id, created_at DESC);

-- Risk Assessments - Status filtering for active assessments
CREATE INDEX CONCURRENTLY idx_risk_assessments_tenant_status_created 
ON risk_assessments(tenant_id, status, created_at DESC) 
WHERE status IN ('pending', 'in_progress', 'completed');

-- Risk Assessments - Industry analysis
CREATE INDEX CONCURRENTLY idx_risk_assessments_tenant_industry_risk 
ON risk_assessments(tenant_id, industry, risk_level, created_at DESC);

-- Risk Assessments - Country-based filtering
CREATE INDEX CONCURRENTLY idx_risk_assessments_tenant_country_created 
ON risk_assessments(tenant_id, country, created_at DESC);

-- Risk Assessments - Compliance status filtering
CREATE INDEX CONCURRENTLY idx_risk_assessments_tenant_compliance 
ON risk_assessments(tenant_id, compliance_status, created_at DESC) 
WHERE compliance_status IS NOT NULL;

-- Risk Predictions - Time-series queries
CREATE INDEX CONCURRENTLY idx_risk_predictions_tenant_business_date 
ON risk_predictions(tenant_id, business_id, prediction_date DESC);

-- Risk Predictions - Model performance analysis
CREATE INDEX CONCURRENTLY idx_risk_predictions_tenant_model_horizon 
ON risk_predictions(tenant_id, model_type, horizon, created_at DESC);

-- Risk Predictions - Assessment relationship
CREATE INDEX CONCURRENTLY idx_risk_predictions_assessment_horizon 
ON risk_predictions(assessment_id, horizon, prediction_date DESC);

-- Risk Factors - Category-based analysis
CREATE INDEX CONCURRENTLY idx_risk_factors_assessment_category 
ON risk_factors(assessment_id, factor_category, impact_score DESC);

-- Risk Factors - High-impact factors
CREATE INDEX CONCURRENTLY idx_risk_factors_tenant_high_impact 
ON risk_factors(tenant_id, impact_score DESC, created_at DESC) 
WHERE impact_score > 0.7;

-- Custom Risk Models - Active models lookup
CREATE INDEX CONCURRENTLY idx_custom_risk_models_tenant_active 
ON custom_risk_models(tenant_id, is_active, status, created_at DESC) 
WHERE is_active = true;

-- Custom Risk Models - Model type filtering
CREATE INDEX CONCURRENTLY idx_custom_risk_models_tenant_type 
ON custom_risk_models(tenant_id, model_type, status, created_at DESC);

-- Batch Jobs - Status and progress tracking
CREATE INDEX CONCURRENTLY idx_batch_jobs_tenant_status_progress 
ON batch_jobs(tenant_id, status, progress_percentage, created_at DESC);

-- Batch Jobs - Job type filtering
CREATE INDEX CONCURRENTLY idx_batch_jobs_tenant_type_status 
ON batch_jobs(tenant_id, job_type, status, created_at DESC);

-- Webhooks - Active webhooks by event type
CREATE INDEX CONCURRENTLY idx_webhooks_tenant_active_events 
ON webhooks(tenant_id, is_active, event_types) 
WHERE is_active = true;

-- Webhook Deliveries - Status and retry tracking
CREATE INDEX CONCURRENTLY idx_webhook_deliveries_webhook_status_retry 
ON webhook_deliveries(webhook_id, status, next_retry_at) 
WHERE status IN ('pending', 'retrying');

-- Webhook Deliveries - Failed deliveries for retry
CREATE INDEX CONCURRENTLY idx_webhook_deliveries_failed_retry 
ON webhook_deliveries(webhook_id, status, attempt_number, created_at DESC) 
WHERE status = 'failed' AND attempt_number < max_attempts;

-- Audit Logs - Entity-based queries
CREATE INDEX CONCURRENTLY idx_audit_logs_tenant_entity 
ON audit_logs(tenant_id, entity_type, entity_id, created_at DESC);

-- Audit Logs - User activity tracking
CREATE INDEX CONCURRENTLY idx_audit_logs_tenant_user_action 
ON audit_logs(tenant_id, user_id, action, created_at DESC) 
WHERE user_id IS NOT NULL;

-- Audit Logs - Event type filtering
CREATE INDEX CONCURRENTLY idx_audit_logs_tenant_event_type 
ON audit_logs(tenant_id, event_type, created_at DESC);

-- Compliance Checks - Assessment relationship
CREATE INDEX CONCURRENTLY idx_compliance_checks_assessment_type 
ON compliance_checks(assessment_id, check_type, status, checked_at DESC);

-- Compliance Checks - Status filtering
CREATE INDEX CONCURRENTLY idx_compliance_checks_tenant_status 
ON compliance_checks(tenant_id, status, check_type, checked_at DESC);

-- Compliance Checks - High-risk results
CREATE INDEX CONCURRENTLY idx_compliance_checks_tenant_high_risk 
ON compliance_checks(tenant_id, risk_score DESC, check_type, checked_at DESC) 
WHERE risk_score > 0.7;

-- GIN indexes for JSONB columns (full-text search and complex queries)

-- Risk Assessments - Risk factors search
CREATE INDEX CONCURRENTLY idx_risk_assessments_risk_factors_gin 
ON risk_assessments USING GIN (risk_factors);

-- Risk Assessments - External data search
CREATE INDEX CONCURRENTLY idx_risk_assessments_external_data_gin 
ON risk_assessments USING GIN (external_data);

-- Risk Assessments - Metadata search
CREATE INDEX CONCURRENTLY idx_risk_assessments_metadata_gin 
ON risk_assessments USING GIN (metadata);

-- Risk Predictions - Contributing factors search
CREATE INDEX CONCURRENTLY idx_risk_predictions_contributing_factors_gin 
ON risk_predictions USING GIN (contributing_factors);

-- Risk Predictions - Scenario analysis search
CREATE INDEX CONCURRENTLY idx_risk_predictions_scenario_analysis_gin 
ON risk_predictions USING GIN (scenario_analysis);

-- Custom Risk Models - Configuration search
CREATE INDEX CONCURRENTLY idx_custom_risk_models_configuration_gin 
ON custom_risk_models USING GIN (configuration);

-- Custom Risk Models - Training data search
CREATE INDEX CONCURRENTLY idx_custom_risk_models_training_data_gin 
ON custom_risk_models USING GIN (training_data);

-- Custom Risk Models - Validation metrics search
CREATE INDEX CONCURRENTLY idx_custom_risk_models_validation_metrics_gin 
ON custom_risk_models USING GIN (validation_metrics);

-- Batch Jobs - Configuration search
CREATE INDEX CONCURRENTLY idx_batch_jobs_configuration_gin 
ON batch_jobs USING GIN (configuration);

-- Batch Jobs - Input data search
CREATE INDEX CONCURRENTLY idx_batch_jobs_input_data_gin 
ON batch_jobs USING GIN (input_data);

-- Batch Jobs - Results search
CREATE INDEX CONCURRENTLY idx_batch_jobs_results_gin 
ON batch_jobs USING GIN (results);

-- Webhooks - Retry policy search
CREATE INDEX CONCURRENTLY idx_webhooks_retry_policy_gin 
ON webhooks USING GIN (retry_policy);

-- Webhook Deliveries - Payload search
CREATE INDEX CONCURRENTLY idx_webhook_deliveries_payload_gin 
ON webhook_deliveries USING GIN (payload);

-- Audit Logs - Old values search
CREATE INDEX CONCURRENTLY idx_audit_logs_old_values_gin 
ON audit_logs USING GIN (old_values);

-- Audit Logs - New values search
CREATE INDEX CONCURRENTLY idx_audit_logs_new_values_gin 
ON audit_logs USING GIN (new_values);

-- Compliance Checks - Search criteria search
CREATE INDEX CONCURRENTLY idx_compliance_checks_search_criteria_gin 
ON compliance_checks USING GIN (search_criteria);

-- Compliance Checks - Results search
CREATE INDEX CONCURRENTLY idx_compliance_checks_results_gin 
ON compliance_checks USING GIN (results);

-- Compliance Checks - Metadata search
CREATE INDEX CONCURRENTLY idx_compliance_checks_metadata_gin 
ON compliance_checks USING GIN (metadata);

-- Partial indexes for specific use cases

-- Risk Assessments - Recent high-risk assessments
CREATE INDEX CONCURRENTLY idx_risk_assessments_recent_high_risk 
ON risk_assessments(tenant_id, created_at DESC) 
WHERE risk_level IN ('high', 'critical') AND created_at > NOW() - INTERVAL '30 days';

-- Risk Assessments - Pending assessments
CREATE INDEX CONCURRENTLY idx_risk_assessments_pending 
ON risk_assessments(tenant_id, created_at ASC) 
WHERE status = 'pending';

-- Risk Predictions - Recent predictions
CREATE INDEX CONCURRENTLY idx_risk_predictions_recent 
ON risk_predictions(tenant_id, prediction_date DESC) 
WHERE prediction_date >= CURRENT_DATE;

-- Batch Jobs - Running jobs
CREATE INDEX CONCURRENTLY idx_batch_jobs_running 
ON batch_jobs(tenant_id, started_at DESC) 
WHERE status = 'running';

-- Webhook Deliveries - Pending deliveries
CREATE INDEX CONCURRENTLY idx_webhook_deliveries_pending 
ON webhook_deliveries(webhook_id, created_at ASC) 
WHERE status = 'pending';

-- Audit Logs - Recent activity
CREATE INDEX CONCURRENTLY idx_audit_logs_recent 
ON audit_logs(tenant_id, created_at DESC) 
WHERE created_at > NOW() - INTERVAL '90 days';

-- Compliance Checks - Recent checks
CREATE INDEX CONCURRENTLY idx_compliance_checks_recent 
ON compliance_checks(tenant_id, checked_at DESC) 
WHERE checked_at > NOW() - INTERVAL '30 days';

-- Text search indexes for business names and descriptions

-- Risk Assessments - Business name search (trigram)
CREATE INDEX CONCURRENTLY idx_risk_assessments_business_name_trgm 
ON risk_assessments USING GIN (business_name gin_trgm_ops);

-- Custom Risk Models - Name and description search
CREATE INDEX CONCURRENTLY idx_custom_risk_models_name_trgm 
ON custom_risk_models USING GIN (name gin_trgm_ops);

CREATE INDEX CONCURRENTLY idx_custom_risk_models_description_trgm 
ON custom_risk_models USING GIN (description gin_trgm_ops);

-- Webhooks - Name search
CREATE INDEX CONCURRENTLY idx_webhooks_name_trgm 
ON webhooks USING GIN (name gin_trgm_ops);

-- Statistics and monitoring indexes

-- Risk Assessments - Score distribution analysis
CREATE INDEX CONCURRENTLY idx_risk_assessments_score_distribution 
ON risk_assessments(tenant_id, risk_score, created_at DESC);

-- Risk Predictions - Confidence analysis
CREATE INDEX CONCURRENTLY idx_risk_predictions_confidence_analysis 
ON risk_predictions(tenant_id, confidence_score, model_type, created_at DESC) 
WHERE confidence_score IS NOT NULL;

-- Batch Jobs - Performance analysis
CREATE INDEX CONCURRENTLY idx_batch_jobs_performance 
ON batch_jobs(tenant_id, job_type, total_requests, completed_requests, created_at DESC);

-- Webhook Deliveries - Success rate analysis
CREATE INDEX CONCURRENTLY idx_webhook_deliveries_success_analysis 
ON webhook_deliveries(webhook_id, status, created_at DESC);

-- Add comments for documentation
COMMENT ON INDEX idx_risk_assessments_tenant_risk_created IS 'Optimized for dashboard queries filtering by tenant, risk level, and date';
COMMENT ON INDEX idx_risk_assessments_tenant_business IS 'Optimized for business-specific risk assessment lookups';
COMMENT ON INDEX idx_risk_assessments_risk_factors_gin IS 'GIN index for complex queries on risk factors JSONB data';
COMMENT ON INDEX idx_risk_assessments_business_name_trgm IS 'Trigram index for fuzzy business name searches';
COMMENT ON INDEX idx_risk_predictions_tenant_business_date IS 'Optimized for time-series prediction queries';
COMMENT ON INDEX idx_batch_jobs_tenant_status_progress IS 'Optimized for batch job monitoring and progress tracking';
COMMENT ON INDEX idx_webhook_deliveries_failed_retry IS 'Optimized for webhook retry logic and failed delivery recovery';
COMMENT ON INDEX idx_audit_logs_tenant_entity IS 'Optimized for audit trail queries by entity type and ID';
COMMENT ON INDEX idx_compliance_checks_assessment_type IS 'Optimized for compliance check lookups by assessment and type';
