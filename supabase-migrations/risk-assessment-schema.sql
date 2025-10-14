-- Risk Assessment Service Database Schema Migration
-- This migration creates all necessary tables for the risk assessment service

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Create custom types
CREATE TYPE risk_level AS ENUM ('low', 'medium', 'high', 'critical');
CREATE TYPE assessment_status AS ENUM ('pending', 'in_progress', 'completed', 'failed', 'cancelled');
CREATE TYPE prediction_horizon AS ENUM ('1_month', '3_months', '6_months', '12_months');
CREATE TYPE batch_job_status AS ENUM ('pending', 'running', 'completed', 'failed', 'cancelled');
CREATE TYPE webhook_event_type AS ENUM ('risk_assessment_completed', 'risk_prediction_updated', 'batch_job_completed', 'custom_model_trained');
CREATE TYPE compliance_status AS ENUM ('compliant', 'non_compliant', 'pending', 'requires_review');

-- Risk Assessments Table
CREATE TABLE risk_assessments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_id VARCHAR(255) NOT NULL,
    business_name VARCHAR(500) NOT NULL,
    business_address TEXT,
    industry VARCHAR(100),
    country VARCHAR(2) NOT NULL,
    tenant_id UUID NOT NULL,
    
    -- Risk Assessment Data
    risk_score DECIMAL(5,4) NOT NULL CHECK (risk_score >= 0 AND risk_score <= 1),
    risk_level risk_level NOT NULL,
    confidence_score DECIMAL(5,4) CHECK (confidence_score >= 0 AND confidence_score <= 1),
    
    -- Assessment Details
    status assessment_status NOT NULL DEFAULT 'pending',
    assessment_type VARCHAR(50) NOT NULL DEFAULT 'standard',
    model_version VARCHAR(50),
    
    -- Risk Factors (JSONB for flexibility)
    risk_factors JSONB,
    external_data JSONB,
    metadata JSONB,
    
    -- Compliance Data
    compliance_status compliance_status,
    sanctions_check BOOLEAN DEFAULT FALSE,
    adverse_media_check BOOLEAN DEFAULT FALSE,
    regulatory_check BOOLEAN DEFAULT FALSE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Audit fields
    created_by UUID,
    updated_by UUID,
    
    CONSTRAINT risk_assessments_risk_score_check CHECK (risk_score >= 0 AND risk_score <= 1),
    CONSTRAINT risk_assessments_confidence_score_check CHECK (confidence_score IS NULL OR (confidence_score >= 0 AND confidence_score <= 1))
);

-- Risk Predictions Table (for time-series forecasts)
CREATE TABLE risk_predictions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_id VARCHAR(255) NOT NULL,
    assessment_id UUID REFERENCES risk_assessments(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    
    -- Prediction Data
    prediction_date DATE NOT NULL,
    horizon prediction_horizon NOT NULL,
    predicted_score DECIMAL(5,4) NOT NULL CHECK (predicted_score >= 0 AND predicted_score <= 1),
    predicted_level risk_level NOT NULL,
    confidence_score DECIMAL(5,4) CHECK (confidence_score >= 0 AND confidence_score <= 1),
    
    -- Model Information
    model_type VARCHAR(50) NOT NULL, -- 'xgboost', 'lstm', 'ensemble'
    model_version VARCHAR(50) NOT NULL,
    
    -- Prediction Factors
    contributing_factors JSONB,
    scenario_analysis JSONB,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT risk_predictions_predicted_score_check CHECK (predicted_score >= 0 AND predicted_score <= 1),
    CONSTRAINT risk_predictions_confidence_score_check CHECK (confidence_score IS NULL OR (confidence_score >= 0 AND confidence_score <= 1))
);

-- Risk Factors Table (for explainability)
CREATE TABLE risk_factors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    assessment_id UUID REFERENCES risk_assessments(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    
    -- Factor Details
    factor_name VARCHAR(100) NOT NULL,
    factor_category VARCHAR(50) NOT NULL, -- 'financial', 'operational', 'compliance', 'reputational'
    factor_value DECIMAL(10,4),
    factor_weight DECIMAL(5,4) CHECK (factor_weight >= 0 AND factor_weight <= 1),
    impact_score DECIMAL(5,4) CHECK (impact_score >= 0 AND impact_score <= 1),
    
    -- Factor Metadata
    description TEXT,
    source VARCHAR(100), -- 'external_api', 'ml_model', 'manual_input'
    confidence DECIMAL(5,4) CHECK (confidence >= 0 AND confidence <= 1),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Custom Risk Models Table (for enterprise features)
CREATE TABLE custom_risk_models (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    
    -- Model Details
    name VARCHAR(200) NOT NULL,
    description TEXT,
    model_type VARCHAR(50) NOT NULL, -- 'xgboost', 'lstm', 'ensemble'
    
    -- Model Configuration
    configuration JSONB NOT NULL,
    training_data JSONB,
    validation_metrics JSONB,
    
    -- Model Status
    status VARCHAR(20) NOT NULL DEFAULT 'draft', -- 'draft', 'training', 'active', 'inactive', 'archived'
    version VARCHAR(20) NOT NULL DEFAULT '1.0.0',
    is_active BOOLEAN DEFAULT FALSE,
    
    -- Performance Metrics
    accuracy DECIMAL(5,4),
    precision_score DECIMAL(5,4),
    recall_score DECIMAL(5,4),
    f1_score DECIMAL(5,4),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    trained_at TIMESTAMP WITH TIME ZONE,
    activated_at TIMESTAMP WITH TIME ZONE,
    
    -- Audit fields
    created_by UUID,
    updated_by UUID
);

-- Batch Jobs Table (for async processing)
CREATE TABLE batch_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    
    -- Job Details
    job_type VARCHAR(50) NOT NULL, -- 'risk_assessment', 'prediction', 'compliance_check'
    status batch_job_status NOT NULL DEFAULT 'pending',
    
    -- Job Configuration
    configuration JSONB,
    input_data JSONB,
    
    -- Progress Tracking
    total_requests INTEGER DEFAULT 0,
    completed_requests INTEGER DEFAULT 0,
    failed_requests INTEGER DEFAULT 0,
    progress_percentage DECIMAL(5,2) DEFAULT 0,
    
    -- Results
    results JSONB,
    error_log JSONB,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Audit fields
    created_by UUID,
    updated_by UUID
);

-- Webhooks Table (for event notifications)
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    
    -- Webhook Details
    name VARCHAR(200) NOT NULL,
    url TEXT NOT NULL,
    secret_key VARCHAR(255),
    
    -- Event Configuration
    event_types webhook_event_type[] NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Delivery Configuration
    retry_policy JSONB,
    timeout_seconds INTEGER DEFAULT 30,
    max_retries INTEGER DEFAULT 3,
    
    -- Statistics
    total_deliveries INTEGER DEFAULT 0,
    successful_deliveries INTEGER DEFAULT 0,
    failed_deliveries INTEGER DEFAULT 0,
    last_delivery_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Audit fields
    created_by UUID,
    updated_by UUID
);

-- Webhook Deliveries Table (for tracking delivery attempts)
CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    webhook_id UUID REFERENCES webhooks(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    
    -- Delivery Details
    event_type webhook_event_type NOT NULL,
    payload JSONB NOT NULL,
    
    -- Delivery Status
    status VARCHAR(20) NOT NULL, -- 'pending', 'delivered', 'failed', 'retrying'
    response_status INTEGER,
    response_body TEXT,
    
    -- Retry Information
    attempt_number INTEGER DEFAULT 1,
    max_attempts INTEGER DEFAULT 3,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    delivered_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE
);

-- Audit Logs Table (for compliance tracking)
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    
    -- Event Details
    event_type VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    
    -- Action Details
    action VARCHAR(50) NOT NULL, -- 'create', 'update', 'delete', 'view'
    old_values JSONB,
    new_values JSONB,
    
    -- Context
    user_id UUID,
    user_email VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Compliance Checks Table (for regulatory compliance)
CREATE TABLE compliance_checks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    assessment_id UUID REFERENCES risk_assessments(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    
    -- Check Details
    check_type VARCHAR(100) NOT NULL, -- 'sanctions', 'adverse_media', 'regulatory', 'pep'
    check_source VARCHAR(100) NOT NULL, -- 'ofac', 'worldcheck', 'newsapi', 'government'
    
    -- Check Results
    status compliance_status NOT NULL,
    risk_score DECIMAL(5,4) CHECK (risk_score >= 0 AND risk_score <= 1),
    match_count INTEGER DEFAULT 0,
    
    -- Check Data
    search_criteria JSONB,
    results JSONB,
    metadata JSONB,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance (will be detailed in next migration)
-- Basic indexes for foreign keys and common queries
CREATE INDEX idx_risk_assessments_business_id ON risk_assessments(business_id);
CREATE INDEX idx_risk_assessments_tenant_id ON risk_assessments(tenant_id);
CREATE INDEX idx_risk_assessments_created_at ON risk_assessments(created_at);
CREATE INDEX idx_risk_assessments_status ON risk_assessments(status);
CREATE INDEX idx_risk_assessments_risk_level ON risk_assessments(risk_level);

CREATE INDEX idx_risk_predictions_business_id ON risk_predictions(business_id);
CREATE INDEX idx_risk_predictions_tenant_id ON risk_predictions(tenant_id);
CREATE INDEX idx_risk_predictions_assessment_id ON risk_predictions(assessment_id);
CREATE INDEX idx_risk_predictions_prediction_date ON risk_predictions(prediction_date);

CREATE INDEX idx_risk_factors_assessment_id ON risk_factors(assessment_id);
CREATE INDEX idx_risk_factors_tenant_id ON risk_factors(tenant_id);
CREATE INDEX idx_risk_factors_category ON risk_factors(factor_category);

CREATE INDEX idx_custom_risk_models_tenant_id ON custom_risk_models(tenant_id);
CREATE INDEX idx_custom_risk_models_status ON custom_risk_models(status);
CREATE INDEX idx_custom_risk_models_is_active ON custom_risk_models(is_active);

CREATE INDEX idx_batch_jobs_tenant_id ON batch_jobs(tenant_id);
CREATE INDEX idx_batch_jobs_status ON batch_jobs(status);
CREATE INDEX idx_batch_jobs_created_at ON batch_jobs(created_at);

CREATE INDEX idx_webhooks_tenant_id ON webhooks(tenant_id);
CREATE INDEX idx_webhooks_is_active ON webhooks(is_active);

CREATE INDEX idx_webhook_deliveries_webhook_id ON webhook_deliveries(webhook_id);
CREATE INDEX idx_webhook_deliveries_tenant_id ON webhook_deliveries(tenant_id);
CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status);
CREATE INDEX idx_webhook_deliveries_created_at ON webhook_deliveries(created_at);

CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_entity_type ON audit_logs(entity_type);
CREATE INDEX idx_audit_logs_entity_id ON audit_logs(entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

CREATE INDEX idx_compliance_checks_assessment_id ON compliance_checks(assessment_id);
CREATE INDEX idx_compliance_checks_tenant_id ON compliance_checks(tenant_id);
CREATE INDEX idx_compliance_checks_check_type ON compliance_checks(check_type);
CREATE INDEX idx_compliance_checks_status ON compliance_checks(status);

-- Add updated_at triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers to relevant tables
CREATE TRIGGER update_risk_assessments_updated_at BEFORE UPDATE ON risk_assessments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_risk_predictions_updated_at BEFORE UPDATE ON risk_predictions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_risk_factors_updated_at BEFORE UPDATE ON risk_factors FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_custom_risk_models_updated_at BEFORE UPDATE ON custom_risk_models FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_batch_jobs_updated_at BEFORE UPDATE ON batch_jobs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_webhooks_updated_at BEFORE UPDATE ON webhooks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_compliance_checks_updated_at BEFORE UPDATE ON compliance_checks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE risk_assessments IS 'Main table storing risk assessment results for businesses';
COMMENT ON TABLE risk_predictions IS 'Time-series predictions for future risk levels';
COMMENT ON TABLE risk_factors IS 'Detailed risk factors contributing to overall risk score';
COMMENT ON TABLE custom_risk_models IS 'Enterprise custom machine learning models';
COMMENT ON TABLE batch_jobs IS 'Asynchronous batch processing jobs';
COMMENT ON TABLE webhooks IS 'Webhook configurations for event notifications';
COMMENT ON TABLE webhook_deliveries IS 'Individual webhook delivery attempts and results';
COMMENT ON TABLE audit_logs IS 'Comprehensive audit trail for compliance';
COMMENT ON TABLE compliance_checks IS 'Regulatory compliance check results';
