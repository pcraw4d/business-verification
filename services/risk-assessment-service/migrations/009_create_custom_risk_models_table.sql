-- Migration: Create custom risk models table
-- Description: Creates table for storing custom risk models for enterprise customers
-- Version: 009
-- Date: 2024-01-20

-- Create custom_risk_models table
CREATE TABLE IF NOT EXISTS custom_risk_models (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    base_model VARCHAR(100) NOT NULL,
    custom_factors JSONB NOT NULL DEFAULT '[]',
    factor_weights JSONB NOT NULL DEFAULT '{}',
    thresholds JSONB NOT NULL DEFAULT '{}',
    validation_rules JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN NOT NULL DEFAULT true,
    version VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    metadata JSONB DEFAULT '{}'
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_custom_risk_models_tenant_id ON custom_risk_models(tenant_id);
CREATE INDEX IF NOT EXISTS idx_custom_risk_models_base_model ON custom_risk_models(base_model);
CREATE INDEX IF NOT EXISTS idx_custom_risk_models_is_active ON custom_risk_models(is_active);
CREATE INDEX IF NOT EXISTS idx_custom_risk_models_created_at ON custom_risk_models(created_at);
CREATE INDEX IF NOT EXISTS idx_custom_risk_models_tenant_base_active ON custom_risk_models(tenant_id, base_model, is_active);

-- Create composite index for active model lookup
CREATE INDEX IF NOT EXISTS idx_custom_risk_models_active_lookup ON custom_risk_models(tenant_id, base_model) WHERE is_active = true;

-- Add constraints
ALTER TABLE custom_risk_models ADD CONSTRAINT chk_custom_risk_models_name_not_empty CHECK (LENGTH(TRIM(name)) > 0);
ALTER TABLE custom_risk_models ADD CONSTRAINT chk_custom_risk_models_base_model_valid CHECK (base_model IN (
    'fintech', 'healthcare', 'technology', 'retail', 'manufacturing', 
    'real_estate', 'energy', 'transportation', 'general'
));
ALTER TABLE custom_risk_models ADD CONSTRAINT chk_custom_risk_models_version_format CHECK (version ~ '^[0-9]+\.[0-9]+\.[0-9]+$');

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_custom_risk_models_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_custom_risk_models_updated_at
    BEFORE UPDATE ON custom_risk_models
    FOR EACH ROW
    EXECUTE FUNCTION update_custom_risk_models_updated_at();

-- Create table for custom model usage tracking
CREATE TABLE IF NOT EXISTS custom_model_usage (
    id SERIAL PRIMARY KEY,
    model_id VARCHAR(255) NOT NULL REFERENCES custom_risk_models(id) ON DELETE CASCADE,
    tenant_id VARCHAR(255) NOT NULL,
    assessment_id VARCHAR(255) NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    risk_score FLOAT,
    processing_time_ms INTEGER,
    metadata JSONB DEFAULT '{}'
);

-- Create indexes for usage tracking
CREATE INDEX IF NOT EXISTS idx_custom_model_usage_model_id ON custom_model_usage(model_id);
CREATE INDEX IF NOT EXISTS idx_custom_model_usage_tenant_id ON custom_model_usage(tenant_id);
CREATE INDEX IF NOT EXISTS idx_custom_model_usage_used_at ON custom_model_usage(used_at);
CREATE INDEX IF NOT EXISTS idx_custom_model_usage_assessment_id ON custom_model_usage(assessment_id);

-- Create table for custom model performance metrics
CREATE TABLE IF NOT EXISTS custom_model_metrics (
    id SERIAL PRIMARY KEY,
    model_id VARCHAR(255) NOT NULL REFERENCES custom_risk_models(id) ON DELETE CASCADE,
    tenant_id VARCHAR(255) NOT NULL,
    metric_date DATE NOT NULL,
    accuracy FLOAT,
    precision_score FLOAT,
    recall_score FLOAT,
    f1_score FLOAT,
    confidence_score FLOAT,
    total_assessments INTEGER DEFAULT 0,
    successful_assessments INTEGER DEFAULT 0,
    failed_assessments INTEGER DEFAULT 0,
    average_processing_time_ms FLOAT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- Create indexes for metrics
CREATE INDEX IF NOT EXISTS idx_custom_model_metrics_model_id ON custom_model_metrics(model_id);
CREATE INDEX IF NOT EXISTS idx_custom_model_metrics_tenant_id ON custom_model_metrics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_custom_model_metrics_metric_date ON custom_model_metrics(metric_date);
CREATE UNIQUE INDEX IF NOT EXISTS idx_custom_model_metrics_unique ON custom_model_metrics(model_id, metric_date);

-- Add comments for documentation
COMMENT ON TABLE custom_risk_models IS 'Stores custom risk models created by enterprise customers';
COMMENT ON COLUMN custom_risk_models.id IS 'Unique identifier for the custom model';
COMMENT ON COLUMN custom_risk_models.tenant_id IS 'Tenant/organization that owns this model';
COMMENT ON COLUMN custom_risk_models.name IS 'Human-readable name for the model';
COMMENT ON COLUMN custom_risk_models.description IS 'Description of the model and its purpose';
COMMENT ON COLUMN custom_risk_models.base_model IS 'Base industry model this custom model extends';
COMMENT ON COLUMN custom_risk_models.custom_factors IS 'JSON array of custom risk factors';
COMMENT ON COLUMN custom_risk_models.factor_weights IS 'JSON object mapping factor names to weights';
COMMENT ON COLUMN custom_risk_models.thresholds IS 'JSON object mapping risk levels to threshold values';
COMMENT ON COLUMN custom_risk_models.validation_rules IS 'JSON array of validation rules for the model';
COMMENT ON COLUMN custom_risk_models.is_active IS 'Whether this model is currently active';
COMMENT ON COLUMN custom_risk_models.version IS 'Semantic version of the model';
COMMENT ON COLUMN custom_risk_models.created_by IS 'User who created this model';
COMMENT ON COLUMN custom_risk_models.metadata IS 'Additional metadata for the model';

COMMENT ON TABLE custom_model_usage IS 'Tracks usage of custom models for analytics';
COMMENT ON TABLE custom_model_metrics IS 'Stores performance metrics for custom models';

-- Insert sample data for testing (optional)
-- INSERT INTO custom_risk_models (
--     id, tenant_id, name, description, base_model, custom_factors, 
--     factor_weights, thresholds, validation_rules, created_by
-- ) VALUES (
--     'sample_model_1',
--     'tenant_1',
--     'Sample Fintech Model',
--     'A sample custom model for fintech companies',
--     'fintech',
--     '[{"id": "annual_revenue", "name": "Annual Revenue", "weight": 0.3}]',
--     '{"financial": 0.3, "operational": 0.25, "compliance": 0.2, "reputational": 0.15, "regulatory": 0.1}',
--     '{"low": 0.25, "medium": 0.5, "high": 0.75, "critical": 1.0}',
--     '[]',
--     'system'
-- );
