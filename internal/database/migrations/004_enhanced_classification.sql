-- Migration: 004_enhanced_classification.sql
-- Description: Update database schema to support enhanced classification features and metrics
-- Created: 2024-01-01
-- Author: System

BEGIN;

-- Add new fields to existing classification table
ALTER TABLE classifications ADD COLUMN IF NOT EXISTS ml_model_version VARCHAR(50);
ALTER TABLE classifications ADD COLUMN IF NOT EXISTS ml_confidence_score DECIMAL(5,4);
ALTER TABLE classifications ADD COLUMN IF NOT EXISTS crosswalk_mappings JSONB;
ALTER TABLE classifications ADD COLUMN IF NOT EXISTS geographic_region VARCHAR(100);
ALTER TABLE classifications ADD COLUMN IF NOT EXISTS region_confidence_score DECIMAL(5,4);
ALTER TABLE classifications ADD COLUMN IF NOT EXISTS industry_specific_data JSONB;
ALTER TABLE classifications ADD COLUMN IF NOT EXISTS classification_algorithm VARCHAR(50);
ALTER TABLE classifications ADD COLUMN IF NOT EXISTS validation_rules_applied JSONB;
ALTER TABLE classifications ADD COLUMN IF NOT EXISTS processing_time_ms INTEGER;
ALTER TABLE classifications ADD COLUMN IF NOT EXISTS enhanced_metadata JSONB;

-- Create feedback collection table
CREATE TABLE IF NOT EXISTS feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(100) NOT NULL,
    business_name VARCHAR(255) NOT NULL,
    original_classification_id UUID REFERENCES classifications(id) ON DELETE CASCADE,
    feedback_type VARCHAR(50) NOT NULL CHECK (feedback_type IN ('accuracy', 'relevance', 'confidence', 'classification', 'suggestion', 'correction')),
    feedback_value JSONB,
    feedback_text TEXT,
    suggested_classification_id UUID REFERENCES classifications(id) ON DELETE SET NULL,
    confidence_score DECIMAL(5,4) CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processed', 'rejected', 'applied')),
    processing_time_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    CONSTRAINT feedback_confidence_range CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0)
);

-- Create accuracy validation table
CREATE TABLE IF NOT EXISTS accuracy_validations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    classification_id UUID REFERENCES classifications(id) ON DELETE CASCADE,
    metric_type VARCHAR(50) NOT NULL CHECK (metric_type IN ('overall', 'industry', 'business_type', 'region', 'confidence', 'time_range')),
    dimension VARCHAR(100) NOT NULL,
    total_classifications INTEGER NOT NULL DEFAULT 0,
    correct_classifications INTEGER NOT NULL DEFAULT 0,
    incorrect_classifications INTEGER NOT NULL DEFAULT 0,
    accuracy_score DECIMAL(5,4) CHECK (accuracy_score >= 0.0 AND accuracy_score <= 1.0),
    confidence_score DECIMAL(5,4) CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0),
    processing_time_ms INTEGER,
    time_range_seconds INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}',
    CONSTRAINT accuracy_score_range CHECK (accuracy_score >= 0.0 AND accuracy_score <= 1.0),
    CONSTRAINT confidence_score_range CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0)
);

-- Create accuracy alerts table
CREATE TABLE IF NOT EXISTS accuracy_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_type VARCHAR(50) NOT NULL CHECK (metric_type IN ('overall', 'industry', 'business_type', 'region', 'confidence', 'time_range')),
    dimension VARCHAR(100) NOT NULL,
    threshold DECIMAL(5,4) NOT NULL CHECK (threshold >= 0.0 AND threshold <= 1.0),
    current_value DECIMAL(5,4) NOT NULL CHECK (current_value >= 0.0 AND current_value <= 1.0),
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    message TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'acknowledged', 'resolved')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    CONSTRAINT alert_threshold_range CHECK (threshold >= 0.0 AND threshold <= 1.0),
    CONSTRAINT alert_current_value_range CHECK (current_value >= 0.0 AND current_value <= 1.0)
);

-- Create accuracy thresholds table
CREATE TABLE IF NOT EXISTS accuracy_thresholds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_type VARCHAR(50) NOT NULL CHECK (metric_type IN ('overall', 'industry', 'business_type', 'region', 'confidence', 'time_range')),
    threshold DECIMAL(5,4) NOT NULL CHECK (threshold >= 0.0 AND threshold <= 1.0),
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    alert_enabled BOOLEAN NOT NULL DEFAULT true,
    description TEXT,
    last_triggered TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}',
    CONSTRAINT threshold_range CHECK (threshold >= 0.0 AND threshold <= 1.0),
    UNIQUE(metric_type, severity)
);

-- Create ML model versions table
CREATE TABLE IF NOT EXISTS ml_model_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_name VARCHAR(100) NOT NULL,
    version VARCHAR(50) NOT NULL,
    model_type VARCHAR(50) NOT NULL CHECK (model_type IN ('bert', 'ensemble', 'custom')),
    file_path VARCHAR(500),
    accuracy_score DECIMAL(5,4) CHECK (accuracy_score >= 0.0 AND accuracy_score <= 1.0),
    training_data_size INTEGER,
    training_date TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT false,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT model_accuracy_range CHECK (accuracy_score >= 0.0 AND accuracy_score <= 1.0),
    UNIQUE(model_name, version)
);

-- Create crosswalk mappings table
CREATE TABLE IF NOT EXISTS crosswalk_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_code VARCHAR(20) NOT NULL,
    source_system VARCHAR(20) NOT NULL CHECK (source_system IN ('naics', 'sic', 'mcc')),
    target_code VARCHAR(20) NOT NULL,
    target_system VARCHAR(20) NOT NULL CHECK (target_system IN ('naics', 'sic', 'mcc')),
    confidence_score DECIMAL(5,4) NOT NULL CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0),
    validation_rules JSONB DEFAULT '[]',
    is_valid BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}',
    CONSTRAINT crosswalk_confidence_range CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0),
    CONSTRAINT crosswalk_different_systems CHECK (source_system != target_system),
    UNIQUE(source_code, source_system, target_code, target_system)
);

-- Create geographic regions table
CREATE TABLE IF NOT EXISTS geographic_regions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region_name VARCHAR(100) NOT NULL,
    region_type VARCHAR(50) NOT NULL CHECK (region_type IN ('country', 'state', 'city', 'postal_code')),
    parent_region_id UUID REFERENCES geographic_regions(id) ON DELETE CASCADE,
    industry_patterns JSONB DEFAULT '{}',
    confidence_modifiers JSONB DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}',
    UNIQUE(region_name, region_type)
);

-- Create industry-specific mappings table
CREATE TABLE IF NOT EXISTS industry_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    industry_type VARCHAR(50) NOT NULL CHECK (industry_type IN ('agriculture', 'retail', 'food', 'manufacturing', 'technology', 'finance', 'healthcare', 'other')),
    industry_name VARCHAR(255) NOT NULL,
    naics_codes TEXT[] DEFAULT '{}',
    sic_codes TEXT[] DEFAULT '{}',
    mcc_codes TEXT[] DEFAULT '{}',
    keywords TEXT[] DEFAULT '{}',
    confidence_score DECIMAL(5,4) NOT NULL CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0),
    validation_rules JSONB DEFAULT '[]',
    classification_algorithm VARCHAR(50) NOT NULL CHECK (classification_algorithm IN ('keyword_based', 'code_density', 'hybrid')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}',
    CONSTRAINT industry_confidence_range CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0),
    UNIQUE(industry_type)
);

-- Create dashboard widgets table
CREATE TABLE IF NOT EXISTS dashboard_widgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    widget_id VARCHAR(100) NOT NULL UNIQUE,
    widget_type VARCHAR(50) NOT NULL CHECK (widget_type IN ('metric', 'chart', 'table', 'alert')),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    position JSONB NOT NULL DEFAULT '{"x": 0, "y": 0, "width": 6, "height": 4}',
    config JSONB DEFAULT '{}',
    data JSONB DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- Create dashboard metrics table
CREATE TABLE IF NOT EXISTS dashboard_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_id VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    value DECIMAL(10,4) NOT NULL,
    unit VARCHAR(50),
    trend VARCHAR(20) CHECK (trend IN ('up', 'down', 'stable')),
    trend_value DECIMAL(10,4),
    status VARCHAR(20) NOT NULL CHECK (status IN ('good', 'warning', 'critical')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- Create alerting rules table
CREATE TABLE IF NOT EXISTS alerting_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metric_type VARCHAR(50) NOT NULL,
    condition VARCHAR(20) NOT NULL CHECK (condition IN ('above', 'below', 'equals')),
    threshold DECIMAL(10,4) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    actions TEXT[] DEFAULT '{}',
    last_triggered TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- Create accuracy reports table
CREATE TABLE IF NOT EXISTS accuracy_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id VARCHAR(100) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    report_type VARCHAR(50) NOT NULL CHECK (report_type IN ('daily', 'weekly', 'monthly')),
    time_range_seconds INTEGER NOT NULL,
    data JSONB NOT NULL DEFAULT '{}',
    summary TEXT,
    recommendations TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- Add indexes for performance optimization

-- Classifications table indexes
CREATE INDEX IF NOT EXISTS idx_classifications_ml_model_version ON classifications(ml_model_version);
CREATE INDEX IF NOT EXISTS idx_classifications_ml_confidence_score ON classifications(ml_confidence_score);
CREATE INDEX IF NOT EXISTS idx_classifications_geographic_region ON classifications(geographic_region);
CREATE INDEX IF NOT EXISTS idx_classifications_classification_algorithm ON classifications(classification_algorithm);
CREATE INDEX IF NOT EXISTS idx_classifications_created_at ON classifications(created_at);
CREATE INDEX IF NOT EXISTS idx_classifications_business_name ON classifications(business_name);
CREATE INDEX IF NOT EXISTS idx_classifications_industry_code ON classifications(industry_code);

-- Feedback table indexes
CREATE INDEX IF NOT EXISTS idx_feedback_user_id ON feedback(user_id);
CREATE INDEX IF NOT EXISTS idx_feedback_business_name ON feedback(business_name);
CREATE INDEX IF NOT EXISTS idx_feedback_feedback_type ON feedback(feedback_type);
CREATE INDEX IF NOT EXISTS idx_feedback_status ON feedback(status);
CREATE INDEX IF NOT EXISTS idx_feedback_created_at ON feedback(created_at);
CREATE INDEX IF NOT EXISTS idx_feedback_original_classification_id ON feedback(original_classification_id);

-- Accuracy validations table indexes
CREATE INDEX IF NOT EXISTS idx_accuracy_validations_metric_type ON accuracy_validations(metric_type);
CREATE INDEX IF NOT EXISTS idx_accuracy_validations_dimension ON accuracy_validations(dimension);
CREATE INDEX IF NOT EXISTS idx_accuracy_validations_accuracy_score ON accuracy_validations(accuracy_score);
CREATE INDEX IF NOT EXISTS idx_accuracy_validations_created_at ON accuracy_validations(created_at);
CREATE INDEX IF NOT EXISTS idx_accuracy_validations_classification_id ON accuracy_validations(classification_id);

-- Accuracy alerts table indexes
CREATE INDEX IF NOT EXISTS idx_accuracy_alerts_metric_type ON accuracy_alerts(metric_type);
CREATE INDEX IF NOT EXISTS idx_accuracy_alerts_severity ON accuracy_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_accuracy_alerts_status ON accuracy_alerts(status);
CREATE INDEX IF NOT EXISTS idx_accuracy_alerts_created_at ON accuracy_alerts(created_at);

-- Accuracy thresholds table indexes
CREATE INDEX IF NOT EXISTS idx_accuracy_thresholds_metric_type ON accuracy_thresholds(metric_type);
CREATE INDEX IF NOT EXISTS idx_accuracy_thresholds_severity ON accuracy_thresholds(severity);
CREATE INDEX IF NOT EXISTS idx_accuracy_thresholds_alert_enabled ON accuracy_thresholds(alert_enabled);

-- ML model versions table indexes
CREATE INDEX IF NOT EXISTS idx_ml_model_versions_model_name ON ml_model_versions(model_name);
CREATE INDEX IF NOT EXISTS idx_ml_model_versions_model_type ON ml_model_versions(model_type);
CREATE INDEX IF NOT EXISTS idx_ml_model_versions_is_active ON ml_model_versions(is_active);
CREATE INDEX IF NOT EXISTS idx_ml_model_versions_accuracy_score ON ml_model_versions(accuracy_score);

-- Crosswalk mappings table indexes
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_source_code ON crosswalk_mappings(source_code);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_source_system ON crosswalk_mappings(source_system);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_target_system ON crosswalk_mappings(target_system);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_confidence_score ON crosswalk_mappings(confidence_score);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_is_valid ON crosswalk_mappings(is_valid);

-- Geographic regions table indexes
CREATE INDEX IF NOT EXISTS idx_geographic_regions_region_name ON geographic_regions(region_name);
CREATE INDEX IF NOT EXISTS idx_geographic_regions_region_type ON geographic_regions(region_type);
CREATE INDEX IF NOT EXISTS idx_geographic_regions_parent_region_id ON geographic_regions(parent_region_id);
CREATE INDEX IF NOT EXISTS idx_geographic_regions_is_active ON geographic_regions(is_active);

-- Industry mappings table indexes
CREATE INDEX IF NOT EXISTS idx_industry_mappings_industry_type ON industry_mappings(industry_type);
CREATE INDEX IF NOT EXISTS idx_industry_mappings_industry_name ON industry_mappings(industry_name);
CREATE INDEX IF NOT EXISTS idx_industry_mappings_confidence_score ON industry_mappings(confidence_score);
CREATE INDEX IF NOT EXISTS idx_industry_mappings_is_active ON industry_mappings(is_active);

-- Dashboard widgets table indexes
CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_widget_type ON dashboard_widgets(widget_type);
CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_is_active ON dashboard_widgets(is_active);

-- Dashboard metrics table indexes
CREATE INDEX IF NOT EXISTS idx_dashboard_metrics_status ON dashboard_metrics(status);
CREATE INDEX IF NOT EXISTS idx_dashboard_metrics_is_active ON dashboard_metrics(is_active);

-- Alerting rules table indexes
CREATE INDEX IF NOT EXISTS idx_alerting_rules_metric_type ON alerting_rules(metric_type);
CREATE INDEX IF NOT EXISTS idx_alerting_rules_severity ON alerting_rules(severity);
CREATE INDEX IF NOT EXISTS idx_alerting_rules_is_enabled ON alerting_rules(is_enabled);

-- Accuracy reports table indexes
CREATE INDEX IF NOT EXISTS idx_accuracy_reports_report_type ON accuracy_reports(report_type);
CREATE INDEX IF NOT EXISTS idx_accuracy_reports_created_at ON accuracy_reports(created_at);

-- Add composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_classifications_business_region ON classifications(business_name, geographic_region);
CREATE INDEX IF NOT EXISTS idx_feedback_user_status ON feedback(user_id, status);
CREATE INDEX IF NOT EXISTS idx_accuracy_validations_type_dimension ON accuracy_validations(metric_type, dimension);
CREATE INDEX IF NOT EXISTS idx_accuracy_alerts_type_severity ON accuracy_alerts(metric_type, severity);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_source_target ON crosswalk_mappings(source_system, target_system);

-- Add GIN indexes for JSONB columns
CREATE INDEX IF NOT EXISTS idx_classifications_crosswalk_mappings_gin ON classifications USING GIN (crosswalk_mappings);
CREATE INDEX IF NOT EXISTS idx_classifications_industry_specific_data_gin ON classifications USING GIN (industry_specific_data);
CREATE INDEX IF NOT EXISTS idx_classifications_validation_rules_applied_gin ON classifications USING GIN (validation_rules_applied);
CREATE INDEX IF NOT EXISTS idx_classifications_enhanced_metadata_gin ON classifications USING GIN (enhanced_metadata);
CREATE INDEX IF NOT EXISTS idx_feedback_feedback_value_gin ON feedback USING GIN (feedback_value);
CREATE INDEX IF NOT EXISTS idx_feedback_metadata_gin ON feedback USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_accuracy_validations_metadata_gin ON accuracy_validations USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_accuracy_alerts_metadata_gin ON accuracy_alerts USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_accuracy_thresholds_metadata_gin ON accuracy_thresholds USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_ml_model_versions_metadata_gin ON ml_model_versions USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_validation_rules_gin ON crosswalk_mappings USING GIN (validation_rules);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_metadata_gin ON crosswalk_mappings USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_geographic_regions_industry_patterns_gin ON geographic_regions USING GIN (industry_patterns);
CREATE INDEX IF NOT EXISTS idx_geographic_regions_confidence_modifiers_gin ON geographic_regions USING GIN (confidence_modifiers);
CREATE INDEX IF NOT EXISTS idx_geographic_regions_metadata_gin ON geographic_regions USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_industry_mappings_validation_rules_gin ON industry_mappings USING GIN (validation_rules);
CREATE INDEX IF NOT EXISTS idx_industry_mappings_metadata_gin ON industry_mappings USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_position_gin ON dashboard_widgets USING GIN (position);
CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_config_gin ON dashboard_widgets USING GIN (config);
CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_data_gin ON dashboard_widgets USING GIN (data);
CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_metadata_gin ON dashboard_widgets USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_dashboard_metrics_metadata_gin ON dashboard_metrics USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_alerting_rules_actions_gin ON alerting_rules USING GIN (actions);
CREATE INDEX IF NOT EXISTS idx_alerting_rules_metadata_gin ON alerting_rules USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_accuracy_reports_data_gin ON accuracy_reports USING GIN (data);
CREATE INDEX IF NOT EXISTS idx_accuracy_reports_recommendations_gin ON accuracy_reports USING GIN (recommendations);
CREATE INDEX IF NOT EXISTS idx_accuracy_reports_metadata_gin ON accuracy_reports USING GIN (metadata);

-- Add triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_accuracy_thresholds_updated_at BEFORE UPDATE ON accuracy_thresholds FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ml_model_versions_updated_at BEFORE UPDATE ON ml_model_versions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_crosswalk_mappings_updated_at BEFORE UPDATE ON crosswalk_mappings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_geographic_regions_updated_at BEFORE UPDATE ON geographic_regions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_industry_mappings_updated_at BEFORE UPDATE ON industry_mappings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_dashboard_widgets_updated_at BEFORE UPDATE ON dashboard_widgets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_dashboard_metrics_updated_at BEFORE UPDATE ON dashboard_metrics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_alerting_rules_updated_at BEFORE UPDATE ON alerting_rules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add data validation constraints
ALTER TABLE classifications ADD CONSTRAINT chk_ml_confidence_score CHECK (ml_confidence_score >= 0.0 AND ml_confidence_score <= 1.0);
ALTER TABLE classifications ADD CONSTRAINT chk_region_confidence_score CHECK (region_confidence_score >= 0.0 AND region_confidence_score <= 1.0);
ALTER TABLE classifications ADD CONSTRAINT chk_processing_time_ms CHECK (processing_time_ms >= 0);

-- Insert default data
INSERT INTO accuracy_thresholds (metric_type, threshold, severity, alert_enabled, description) VALUES
('overall', 0.8, 'high', true, 'Overall accuracy below 80%'),
('industry', 0.7, 'medium', true, 'Industry accuracy below 70%'),
('confidence', 0.6, 'low', true, 'Confidence-based accuracy below 60%')
ON CONFLICT (metric_type, severity) DO NOTHING;

INSERT INTO ml_model_versions (model_name, version, model_type, accuracy_score, is_active) VALUES
('bert_classifier', '1.0.0', 'bert', 0.85, true),
('ensemble_classifier', '1.0.0', 'ensemble', 0.88, false)
ON CONFLICT (model_name, version) DO NOTHING;

INSERT INTO industry_mappings (industry_type, industry_name, confidence_score, classification_algorithm) VALUES
('agriculture', 'Agriculture, Forestry, Fishing and Hunting', 0.85, 'hybrid'),
('retail', 'Retail Trade', 0.80, 'keyword_based'),
('food', 'Food Services and Drinking Places', 0.90, 'keyword_based'),
('manufacturing', 'Manufacturing', 0.75, 'code_density')
ON CONFLICT (industry_type) DO NOTHING;

INSERT INTO dashboard_widgets (widget_id, widget_type, title, description, position) VALUES
('overall_accuracy', 'metric', 'Overall Accuracy', 'Overall classification accuracy', '{"x": 0, "y": 0, "width": 6, "height": 4}'),
('industry_accuracy', 'chart', 'Accuracy by Industry', 'Classification accuracy broken down by industry', '{"x": 6, "y": 0, "width": 6, "height": 4}'),
('confidence_accuracy', 'chart', 'Accuracy by Confidence', 'Classification accuracy by confidence level', '{"x": 0, "y": 4, "width": 6, "height": 4}'),
('active_alerts', 'table', 'Active Alerts', 'Currently active accuracy alerts', '{"x": 6, "y": 4, "width": 6, "height": 4}')
ON CONFLICT (widget_id) DO NOTHING;

INSERT INTO alerting_rules (rule_id, name, description, metric_type, condition, threshold, severity, is_enabled, actions) VALUES
('overall_accuracy_low', 'Overall Accuracy Below 80%', 'Alert when overall accuracy drops below 80%', 'overall_accuracy', 'below', 0.8, 'high', true, ARRAY['email', 'slack']),
('industry_accuracy_low', 'Industry Accuracy Below 70%', 'Alert when any industry accuracy drops below 70%', 'industry_accuracy', 'below', 0.7, 'medium', true, ARRAY['slack']),
('confidence_accuracy_low', 'Low Confidence Accuracy Below 60%', 'Alert when low confidence accuracy drops below 60%', 'confidence_accuracy', 'below', 0.6, 'low', true, ARRAY['slack'])
ON CONFLICT (rule_id) DO NOTHING;

COMMIT;
