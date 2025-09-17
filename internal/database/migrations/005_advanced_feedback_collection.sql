-- Migration: 005_advanced_feedback_collection.sql
-- Description: Create database schema for advanced feedback collection with model versioning
-- Created: 2024-01-01
-- Author: System

BEGIN;

-- Create model versions table for tracking ML model versions
CREATE TABLE IF NOT EXISTS model_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_type VARCHAR(100) NOT NULL,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    metadata JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(model_type, version)
);

-- Create index for model versions
CREATE INDEX IF NOT EXISTS idx_model_versions_type_version ON model_versions(model_type, version);
CREATE INDEX IF NOT EXISTS idx_model_versions_active ON model_versions(is_active) WHERE is_active = true;

-- Enhanced user feedback table with model versioning
CREATE TABLE IF NOT EXISTS user_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(100) NOT NULL,
    business_name VARCHAR(255) NOT NULL,
    original_classification_id UUID REFERENCES business_classifications(id) ON DELETE CASCADE,
    feedback_type VARCHAR(50) NOT NULL CHECK (feedback_type IN (
        'accuracy', 'relevance', 'confidence', 'classification', 'suggestion', 'correction'
    )),
    feedback_value JSONB,
    feedback_text TEXT,
    suggested_classification_id UUID REFERENCES business_classifications(id) ON DELETE SET NULL,
    confidence_score DECIMAL(5,4) CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processed', 'rejected', 'applied')),
    processing_time_ms INTEGER,
    model_version_id UUID REFERENCES model_versions(id) ON DELETE SET NULL,
    classification_method VARCHAR(50) NOT NULL,
    ensemble_weight DECIMAL(5,4),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    CONSTRAINT user_feedback_confidence_range CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0)
);

-- Create indexes for user feedback
CREATE INDEX IF NOT EXISTS idx_user_feedback_user_id ON user_feedback(user_id);
CREATE INDEX IF NOT EXISTS idx_user_feedback_classification_id ON user_feedback(original_classification_id);
CREATE INDEX IF NOT EXISTS idx_user_feedback_type ON user_feedback(feedback_type);
CREATE INDEX IF NOT EXISTS idx_user_feedback_status ON user_feedback(status);
CREATE INDEX IF NOT EXISTS idx_user_feedback_created_at ON user_feedback(created_at);
CREATE INDEX IF NOT EXISTS idx_user_feedback_model_version ON user_feedback(model_version_id);

-- ML model feedback table
CREATE TABLE IF NOT EXISTS ml_model_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_version_id UUID REFERENCES model_versions(id) ON DELETE CASCADE,
    model_type VARCHAR(100) NOT NULL,
    classification_method VARCHAR(50) NOT NULL,
    prediction_id VARCHAR(255) NOT NULL,
    actual_result JSONB NOT NULL,
    predicted_result JSONB NOT NULL,
    accuracy_score DECIMAL(5,4) CHECK (accuracy_score >= 0.0 AND accuracy_score <= 1.0),
    confidence_score DECIMAL(5,4) CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0),
    processing_time_ms INTEGER,
    error_type VARCHAR(100),
    error_description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processed', 'rejected', 'applied')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    CONSTRAINT ml_model_feedback_accuracy_range CHECK (accuracy_score >= 0.0 AND accuracy_score <= 1.0),
    CONSTRAINT ml_model_feedback_confidence_range CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0)
);

-- Create indexes for ML model feedback
CREATE INDEX IF NOT EXISTS idx_ml_model_feedback_model_version ON ml_model_feedback(model_version_id);
CREATE INDEX IF NOT EXISTS idx_ml_model_feedback_model_type ON ml_model_feedback(model_type);
CREATE INDEX IF NOT EXISTS idx_ml_model_feedback_method ON ml_model_feedback(classification_method);
CREATE INDEX IF NOT EXISTS idx_ml_model_feedback_prediction_id ON ml_model_feedback(prediction_id);
CREATE INDEX IF NOT EXISTS idx_ml_model_feedback_status ON ml_model_feedback(status);
CREATE INDEX IF NOT EXISTS idx_ml_model_feedback_created_at ON ml_model_feedback(created_at);

-- Security validation feedback table
CREATE TABLE IF NOT EXISTS security_validation_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    validation_type VARCHAR(100) NOT NULL,
    data_source_type VARCHAR(100) NOT NULL,
    website_url VARCHAR(1000),
    validation_result JSONB NOT NULL,
    trust_score DECIMAL(5,4) CHECK (trust_score >= 0.0 AND trust_score <= 1.0),
    verification_status VARCHAR(50) NOT NULL,
    security_violations TEXT[],
    processing_time_ms INTEGER,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processed', 'rejected', 'applied')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    CONSTRAINT security_validation_feedback_trust_range CHECK (trust_score >= 0.0 AND trust_score <= 1.0)
);

-- Create indexes for security validation feedback
CREATE INDEX IF NOT EXISTS idx_security_validation_feedback_type ON security_validation_feedback(validation_type);
CREATE INDEX IF NOT EXISTS idx_security_validation_feedback_data_source ON security_validation_feedback(data_source_type);
CREATE INDEX IF NOT EXISTS idx_security_validation_feedback_website ON security_validation_feedback(website_url);
CREATE INDEX IF NOT EXISTS idx_security_validation_feedback_status ON security_validation_feedback(status);
CREATE INDEX IF NOT EXISTS idx_security_validation_feedback_created_at ON security_validation_feedback(created_at);

-- Feedback trends table for aggregated analysis
CREATE TABLE IF NOT EXISTS feedback_trends (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    method VARCHAR(50) NOT NULL,
    time_window VARCHAR(20) NOT NULL,
    total_feedback INTEGER NOT NULL DEFAULT 0,
    positive_feedback INTEGER NOT NULL DEFAULT 0,
    negative_feedback INTEGER NOT NULL DEFAULT 0,
    average_accuracy DECIMAL(5,4) CHECK (average_accuracy >= 0.0 AND average_accuracy <= 1.0),
    average_confidence DECIMAL(5,4) CHECK (average_confidence >= 0.0 AND average_confidence <= 1.0),
    average_processing_time INTEGER,
    error_rate DECIMAL(5,4) CHECK (error_rate >= 0.0 AND error_rate <= 1.0),
    security_violation_rate DECIMAL(5,4) CHECK (security_violation_rate >= 0.0 AND security_violation_rate <= 1.0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT feedback_trends_accuracy_range CHECK (average_accuracy >= 0.0 AND average_accuracy <= 1.0),
    CONSTRAINT feedback_trends_confidence_range CHECK (average_confidence >= 0.0 AND average_confidence <= 1.0),
    CONSTRAINT feedback_trends_error_rate_range CHECK (error_rate >= 0.0 AND error_rate <= 1.0),
    CONSTRAINT feedback_trends_security_violation_range CHECK (security_violation_rate >= 0.0 AND security_violation_rate <= 1.0)
);

-- Create indexes for feedback trends
CREATE INDEX IF NOT EXISTS idx_feedback_trends_method ON feedback_trends(method);
CREATE INDEX IF NOT EXISTS idx_feedback_trends_time_window ON feedback_trends(time_window);
CREATE INDEX IF NOT EXISTS idx_feedback_trends_created_at ON feedback_trends(created_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_feedback_trends_method_time_window ON feedback_trends(method, time_window);

-- Feedback processing queue table for async processing
CREATE TABLE IF NOT EXISTS feedback_processing_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feedback_type VARCHAR(50) NOT NULL,
    feedback_id UUID NOT NULL,
    priority INTEGER DEFAULT 5 CHECK (priority >= 1 AND priority <= 10),
    status VARCHAR(20) NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'processing', 'completed', 'failed', 'retry')),
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    error_message TEXT,
    scheduled_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- Create indexes for feedback processing queue
CREATE INDEX IF NOT EXISTS idx_feedback_processing_queue_status ON feedback_processing_queue(status);
CREATE INDEX IF NOT EXISTS idx_feedback_processing_queue_priority ON feedback_processing_queue(priority);
CREATE INDEX IF NOT EXISTS idx_feedback_processing_queue_scheduled_at ON feedback_processing_queue(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_feedback_processing_queue_feedback_type ON feedback_processing_queue(feedback_type);

-- Create triggers for updating timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers to tables with updated_at columns
CREATE TRIGGER update_model_versions_updated_at BEFORE UPDATE ON model_versions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_feedback_trends_updated_at BEFORE UPDATE ON feedback_trends
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function to automatically update feedback trends
CREATE OR REPLACE FUNCTION update_feedback_trends()
RETURNS TRIGGER AS $$
DECLARE
    method_name VARCHAR(50);
    time_window_name VARCHAR(20);
BEGIN
    -- Determine method from the feedback
    IF TG_TABLE_NAME = 'user_feedback' THEN
        method_name := NEW.classification_method;
    ELSIF TG_TABLE_NAME = 'ml_model_feedback' THEN
        method_name := NEW.classification_method;
    ELSIF TG_TABLE_NAME = 'security_validation_feedback' THEN
        method_name := 'security';
    END IF;
    
    -- Determine time window (daily for now)
    time_window_name := 'daily';
    
    -- Insert or update feedback trends
    INSERT INTO feedback_trends (
        method, time_window, total_feedback, positive_feedback, negative_feedback,
        average_accuracy, average_confidence, average_processing_time, error_rate, security_violation_rate
    ) VALUES (
        method_name, time_window_name, 1, 
        CASE WHEN NEW.status = 'processed' THEN 1 ELSE 0 END,
        CASE WHEN NEW.status = 'rejected' THEN 1 ELSE 0 END,
        COALESCE(NEW.accuracy_score, 0.0),
        COALESCE(NEW.confidence_score, 0.0),
        COALESCE(NEW.processing_time_ms, 0),
        0.0, 0.0
    )
    ON CONFLICT (method, time_window) 
    DO UPDATE SET
        total_feedback = feedback_trends.total_feedback + 1,
        positive_feedback = feedback_trends.positive_feedback + 
            CASE WHEN NEW.status = 'processed' THEN 1 ELSE 0 END,
        negative_feedback = feedback_trends.negative_feedback + 
            CASE WHEN NEW.status = 'rejected' THEN 1 ELSE 0 END,
        average_accuracy = (feedback_trends.average_accuracy * feedback_trends.total_feedback + 
            COALESCE(NEW.accuracy_score, 0.0)) / (feedback_trends.total_feedback + 1),
        average_confidence = (feedback_trends.average_confidence * feedback_trends.total_feedback + 
            COALESCE(NEW.confidence_score, 0.0)) / (feedback_trends.total_feedback + 1),
        average_processing_time = (feedback_trends.average_processing_time * feedback_trends.total_feedback + 
            COALESCE(NEW.processing_time_ms, 0)) / (feedback_trends.total_feedback + 1),
        updated_at = NOW();
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers to update feedback trends
CREATE TRIGGER update_user_feedback_trends AFTER INSERT ON user_feedback
    FOR EACH ROW EXECUTE FUNCTION update_feedback_trends();

CREATE TRIGGER update_ml_model_feedback_trends AFTER INSERT ON ml_model_feedback
    FOR EACH ROW EXECUTE FUNCTION update_feedback_trends();

CREATE TRIGGER update_security_validation_feedback_trends AFTER INSERT ON security_validation_feedback
    FOR EACH ROW EXECUTE FUNCTION update_feedback_trends();

-- Create views for easy querying
CREATE OR REPLACE VIEW feedback_summary AS
SELECT 
    'user' as feedback_source,
    uf.id,
    uf.user_id,
    uf.feedback_type,
    uf.status,
    uf.confidence_score,
    uf.processing_time_ms,
    uf.classification_method,
    uf.created_at
FROM user_feedback uf
UNION ALL
SELECT 
    'ml_model' as feedback_source,
    mf.id,
    mf.model_type as user_id,
    'ml_performance' as feedback_type,
    mf.status,
    mf.confidence_score,
    mf.processing_time_ms,
    mf.classification_method,
    mf.created_at
FROM ml_model_feedback mf
UNION ALL
SELECT 
    'security' as feedback_source,
    svf.id,
    svf.validation_type as user_id,
    'security_validation' as feedback_type,
    svf.status,
    svf.trust_score as confidence_score,
    svf.processing_time_ms,
    'security' as classification_method,
    svf.created_at
FROM security_validation_feedback svf;

-- Create view for method performance analysis
CREATE OR REPLACE VIEW method_performance_analysis AS
SELECT 
    ft.method,
    ft.time_window,
    ft.total_feedback,
    ft.positive_feedback,
    ft.negative_feedback,
    ft.average_accuracy,
    ft.average_confidence,
    ft.average_processing_time,
    ft.error_rate,
    ft.security_violation_rate,
    CASE 
        WHEN ft.total_feedback > 0 THEN 
            (ft.positive_feedback::DECIMAL / ft.total_feedback::DECIMAL) * 100
        ELSE 0 
    END as success_rate_percentage,
    ft.created_at,
    ft.updated_at
FROM feedback_trends ft
ORDER BY ft.method, ft.time_window, ft.created_at DESC;

COMMIT;
