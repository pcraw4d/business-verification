-- Migration: Create webhook tables
-- Description: Creates tables for webhook management, delivery tracking, and retry handling

-- Create webhooks table
CREATE TABLE IF NOT EXISTS webhooks (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    url TEXT NOT NULL,
    events JSONB NOT NULL, -- Array of webhook event types
    secret TEXT, -- For signature verification
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'inactive', 'paused', 'disabled')),
    retry_policy JSONB NOT NULL, -- Retry configuration
    rate_limit JSONB NOT NULL, -- Rate limiting configuration
    headers JSONB DEFAULT '{}', -- Custom headers
    filters JSONB DEFAULT '{}', -- Event filtering configuration
    statistics JSONB DEFAULT '{}', -- Delivery statistics
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    
    -- Indexes for performance
    INDEX idx_webhooks_tenant_id (tenant_id),
    INDEX idx_webhooks_status (status),
    INDEX idx_webhooks_created_by (created_by),
    INDEX idx_webhooks_created_at (created_at),
    INDEX idx_webhooks_last_triggered_at (last_triggered_at),
    INDEX idx_webhooks_events_gin (events) USING GIN, -- For JSONB queries
    
    -- Constraints
    CONSTRAINT fk_webhooks_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create webhook_deliveries table
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id VARCHAR(255) PRIMARY KEY,
    webhook_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    event_id VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL, -- Event payload data
    headers JSONB DEFAULT '{}', -- Request headers
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'sending', 'delivered', 'failed', 'retrying', 'cancelled')),
    attempts INT DEFAULT 0,
    max_attempts INT DEFAULT 3,
    response_code INT,
    response_body TEXT,
    response_headers JSONB DEFAULT '{}',
    latency INTERVAL,
    error TEXT,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    delivered_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    
    -- Indexes for performance
    INDEX idx_webhook_deliveries_webhook_id (webhook_id),
    INDEX idx_webhook_deliveries_tenant_id (tenant_id),
    INDEX idx_webhook_deliveries_status (status),
    INDEX idx_webhook_deliveries_event_type (event_type),
    INDEX idx_webhook_deliveries_created_at (created_at),
    INDEX idx_webhook_deliveries_next_retry_at (next_retry_at),
    INDEX idx_webhook_deliveries_webhook_status (webhook_id, status),
    
    -- Constraints
    CONSTRAINT fk_webhook_deliveries_webhook FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE,
    CONSTRAINT fk_webhook_deliveries_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create webhook_templates table
CREATE TABLE IF NOT EXISTS webhook_templates (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    events JSONB NOT NULL, -- Array of webhook event types
    url_template TEXT NOT NULL,
    headers JSONB DEFAULT '{}',
    retry_policy JSONB NOT NULL,
    rate_limit JSONB NOT NULL,
    filters JSONB DEFAULT '{}',
    is_public BOOLEAN DEFAULT FALSE,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    
    -- Indexes
    INDEX idx_webhook_templates_is_public (is_public),
    INDEX idx_webhook_templates_created_by (created_by),
    INDEX idx_webhook_templates_created_at (created_at),
    INDEX idx_webhook_templates_events_gin (events) USING GIN
);

-- Create webhook_delivery_logs table for detailed logging
CREATE TABLE IF NOT EXISTS webhook_delivery_logs (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    delivery_id VARCHAR(255) NOT NULL,
    webhook_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    attempt_number INT NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('started', 'completed', 'failed', 'cancelled')),
    response_code INT,
    response_body TEXT,
    response_headers JSONB DEFAULT '{}',
    latency INTERVAL,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    
    -- Indexes
    INDEX idx_webhook_delivery_logs_delivery_id (delivery_id),
    INDEX idx_webhook_delivery_logs_webhook_id (webhook_id),
    INDEX idx_webhook_delivery_logs_tenant_id (tenant_id),
    INDEX idx_webhook_delivery_logs_status (status),
    INDEX idx_webhook_delivery_logs_started_at (started_at),
    
    -- Constraints
    CONSTRAINT fk_webhook_delivery_logs_delivery FOREIGN KEY (delivery_id) REFERENCES webhook_deliveries(id) ON DELETE CASCADE,
    CONSTRAINT fk_webhook_delivery_logs_webhook FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE,
    CONSTRAINT fk_webhook_delivery_logs_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create webhook_events table for event tracking
CREATE TABLE IF NOT EXISTS webhook_events (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    business_id VARCHAR(255),
    data JSONB NOT NULL,
    source VARCHAR(100) NOT NULL,
    version VARCHAR(20) DEFAULT '1.0',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    
    -- Indexes
    INDEX idx_webhook_events_tenant_id (tenant_id),
    INDEX idx_webhook_events_event_type (event_type),
    INDEX idx_webhook_events_business_id (business_id),
    INDEX idx_webhook_events_created_at (created_at),
    INDEX idx_webhook_events_tenant_type (tenant_id, event_type),
    
    -- Constraints
    CONSTRAINT fk_webhook_events_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create webhook_subscriptions table for event subscriptions
CREATE TABLE IF NOT EXISTS webhook_subscriptions (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    webhook_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    filters JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_webhook_subscriptions_webhook_id (webhook_id),
    INDEX idx_webhook_subscriptions_tenant_id (tenant_id),
    INDEX idx_webhook_subscriptions_event_type (event_type),
    INDEX idx_webhook_subscriptions_is_active (is_active),
    
    -- Constraints
    CONSTRAINT fk_webhook_subscriptions_webhook FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE,
    CONSTRAINT fk_webhook_subscriptions_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    UNIQUE(webhook_id, event_type)
);

-- Create webhook_health_checks table for monitoring webhook health
CREATE TABLE IF NOT EXISTS webhook_health_checks (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    webhook_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('healthy', 'degraded', 'unhealthy')),
    last_successful TIMESTAMP WITH TIME ZONE,
    last_failed TIMESTAMP WITH TIME ZONE,
    consecutive_fails INT DEFAULT 0,
    average_latency FLOAT DEFAULT 0,
    success_rate FLOAT DEFAULT 0,
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    
    -- Indexes
    INDEX idx_webhook_health_checks_webhook_id (webhook_id),
    INDEX idx_webhook_health_checks_tenant_id (tenant_id),
    INDEX idx_webhook_health_checks_status (status),
    INDEX idx_webhook_health_checks_checked_at (checked_at),
    
    -- Constraints
    CONSTRAINT fk_webhook_health_checks_webhook FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE,
    CONSTRAINT fk_webhook_health_checks_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create webhook_rate_limits table for rate limiting tracking
CREATE TABLE IF NOT EXISTS webhook_rate_limits (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    webhook_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    requests INT DEFAULT 0,
    window_start TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    window_size INTERVAL NOT NULL,
    max_requests INT NOT NULL,
    burst INT DEFAULT 0,
    current_burst INT DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_webhook_rate_limits_webhook_id (webhook_id),
    INDEX idx_webhook_rate_limits_tenant_id (tenant_id),
    INDEX idx_webhook_rate_limits_window_start (window_start),
    
    -- Constraints
    CONSTRAINT fk_webhook_rate_limits_webhook FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE,
    CONSTRAINT fk_webhook_rate_limits_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    UNIQUE(webhook_id)
);

-- Create webhook_circuit_breakers table for circuit breaker state
CREATE TABLE IF NOT EXISTS webhook_circuit_breakers (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    webhook_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    state VARCHAR(20) NOT NULL CHECK (state IN ('closed', 'open', 'half_open')),
    failure_count INT DEFAULT 0,
    last_failure_time TIMESTAMP WITH TIME ZONE,
    next_attempt_time TIMESTAMP WITH TIME ZONE,
    failure_threshold INT DEFAULT 5,
    recovery_timeout INTERVAL DEFAULT '60 seconds',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_webhook_circuit_breakers_webhook_id (webhook_id),
    INDEX idx_webhook_circuit_breakers_tenant_id (tenant_id),
    INDEX idx_webhook_circuit_breakers_state (state),
    INDEX idx_webhook_circuit_breakers_next_attempt_time (next_attempt_time),
    
    -- Constraints
    CONSTRAINT fk_webhook_circuit_breakers_webhook FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE,
    CONSTRAINT fk_webhook_circuit_breakers_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    UNIQUE(webhook_id)
);

-- Create webhook_signatures table for signature verification
CREATE TABLE IF NOT EXISTS webhook_signatures (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    webhook_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    algorithm VARCHAR(20) NOT NULL DEFAULT 'sha256',
    signature TEXT NOT NULL,
    timestamp VARCHAR(20) NOT NULL,
    nonce VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_webhook_signatures_webhook_id (webhook_id),
    INDEX idx_webhook_signatures_tenant_id (tenant_id),
    INDEX idx_webhook_signatures_created_at (created_at),
    
    -- Constraints
    CONSTRAINT fk_webhook_signatures_webhook FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE,
    CONSTRAINT fk_webhook_signatures_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_webhooks_tenant_status ON webhooks(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_webhooks_tenant_created_by ON webhooks(tenant_id, created_by);
CREATE INDEX IF NOT EXISTS idx_webhooks_tenant_created_at ON webhooks(tenant_id, created_at);

CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook_created_at ON webhook_deliveries(webhook_id, created_at);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_tenant_status ON webhook_deliveries(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_tenant_created_at ON webhook_deliveries(tenant_id, created_at);

CREATE INDEX IF NOT EXISTS idx_webhook_templates_is_public_created_at ON webhook_templates(is_public, created_at);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_webhook_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_webhooks_updated_at
    BEFORE UPDATE ON webhooks
    FOR EACH ROW
    EXECUTE FUNCTION update_webhook_updated_at();

CREATE TRIGGER update_webhook_templates_updated_at
    BEFORE UPDATE ON webhook_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_webhook_updated_at();

CREATE TRIGGER update_webhook_subscriptions_updated_at
    BEFORE UPDATE ON webhook_subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_webhook_updated_at();

-- Create function to clean up old webhook deliveries (older than 30 days)
CREATE OR REPLACE FUNCTION cleanup_old_webhook_deliveries()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM webhook_deliveries 
    WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '30 days';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Create function to clean up old webhook delivery logs (older than 7 days)
CREATE OR REPLACE FUNCTION cleanup_old_webhook_delivery_logs()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM webhook_delivery_logs 
    WHERE started_at < CURRENT_TIMESTAMP - INTERVAL '7 days';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Create function to clean up old webhook events (older than 90 days)
CREATE OR REPLACE FUNCTION cleanup_old_webhook_events()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM webhook_events 
    WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '90 days';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Create function to clean up old webhook signatures (older than 1 day)
CREATE OR REPLACE FUNCTION cleanup_old_webhook_signatures()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM webhook_signatures 
    WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '1 day';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Create function to update webhook statistics
CREATE OR REPLACE FUNCTION update_webhook_statistics(webhook_id_param VARCHAR(255))
RETURNS VOID AS $$
DECLARE
    total_deliveries INTEGER;
    successful_deliveries INTEGER;
    failed_deliveries INTEGER;
    success_rate FLOAT;
    average_latency FLOAT;
    last_delivery_at TIMESTAMP WITH TIME ZONE;
BEGIN
    -- Get delivery statistics
    SELECT 
        COUNT(*),
        COUNT(*) FILTER (WHERE status = 'delivered'),
        COUNT(*) FILTER (WHERE status = 'failed'),
        AVG(EXTRACT(EPOCH FROM latency) * 1000) FILTER (WHERE latency IS NOT NULL),
        MAX(created_at)
    INTO 
        total_deliveries,
        successful_deliveries,
        failed_deliveries,
        average_latency,
        last_delivery_at
    FROM webhook_deliveries 
    WHERE webhook_id = webhook_id_param;
    
    -- Calculate success rate
    IF total_deliveries > 0 THEN
        success_rate := (successful_deliveries::FLOAT / total_deliveries::FLOAT) * 100;
    ELSE
        success_rate := 0;
    END IF;
    
    -- Update webhook statistics
    UPDATE webhooks 
    SET statistics = jsonb_build_object(
        'total_deliveries', total_deliveries,
        'successful_deliveries', successful_deliveries,
        'failed_deliveries', failed_deliveries,
        'success_rate', success_rate,
        'average_latency', COALESCE(average_latency, 0),
        'last_delivery_at', last_delivery_at
    ),
    updated_at = CURRENT_TIMESTAMP
    WHERE id = webhook_id_param;
END;
$$ language 'plpgsql';

-- Create function to check webhook health
CREATE OR REPLACE FUNCTION check_webhook_health(webhook_id_param VARCHAR(255))
RETURNS TABLE(
    webhook_id VARCHAR(255),
    status VARCHAR(20),
    last_successful TIMESTAMP WITH TIME ZONE,
    last_failed TIMESTAMP WITH TIME ZONE,
    consecutive_fails INTEGER,
    average_latency FLOAT,
    success_rate FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        w.id,
        CASE 
            WHEN w.statistics->>'success_rate' IS NULL THEN 'unknown'::VARCHAR(20)
            WHEN (w.statistics->>'success_rate')::FLOAT >= 95 AND 
                 (w.statistics->>'consecutive_fails')::INTEGER < 3 THEN 'healthy'::VARCHAR(20)
            WHEN (w.statistics->>'success_rate')::FLOAT >= 80 AND 
                 (w.statistics->>'consecutive_fails')::INTEGER < 5 THEN 'degraded'::VARCHAR(20)
            ELSE 'unhealthy'::VARCHAR(20)
        END as status,
        (w.statistics->>'last_successful')::TIMESTAMP WITH TIME ZONE as last_successful,
        (w.statistics->>'last_failed')::TIMESTAMP WITH TIME ZONE as last_failed,
        (w.statistics->>'consecutive_fails')::INTEGER as consecutive_fails,
        (w.statistics->>'average_latency')::FLOAT as average_latency,
        (w.statistics->>'success_rate')::FLOAT as success_rate
    FROM webhooks w
    WHERE w.id = webhook_id_param;
END;
$$ language 'plpgsql';

-- Insert sample webhook templates for testing
INSERT INTO webhook_templates (
    id, name, description, events, url_template, is_public, created_by
) VALUES (
    'risk_assessment_template', 'Risk Assessment Webhook', 'Webhook for risk assessment events', 
    '["risk_assessment.started", "risk_assessment.completed", "risk_assessment.failed"]'::jsonb,
    'https://your-app.com/webhooks/risk-assessment', true, 'system'
), (
    'batch_job_template', 'Batch Job Webhook', 'Webhook for batch job events',
    '["batch_job.started", "batch_job.completed", "batch_job.failed", "batch_job.progress"]'::jsonb,
    'https://your-app.com/webhooks/batch-jobs', true, 'system'
), (
    'compliance_template', 'Compliance Webhook', 'Webhook for compliance events',
    '["risk_threshold.exceeded", "risk_threshold.recovered"]'::jsonb,
    'https://your-app.com/webhooks/compliance', true, 'system'
) ON CONFLICT (id) DO NOTHING;

-- Add comments to tables for documentation
COMMENT ON TABLE webhooks IS 'Stores webhook configurations and settings';
COMMENT ON TABLE webhook_deliveries IS 'Tracks webhook delivery attempts and results';
COMMENT ON TABLE webhook_templates IS 'Stores reusable webhook templates';
COMMENT ON TABLE webhook_delivery_logs IS 'Detailed logs for webhook delivery attempts';
COMMENT ON TABLE webhook_events IS 'Stores webhook events for processing';
COMMENT ON TABLE webhook_subscriptions IS 'Manages webhook event subscriptions';
COMMENT ON TABLE webhook_health_checks IS 'Tracks webhook health status';
COMMENT ON TABLE webhook_rate_limits IS 'Manages webhook rate limiting';
COMMENT ON TABLE webhook_circuit_breakers IS 'Manages webhook circuit breaker state';
COMMENT ON TABLE webhook_signatures IS 'Stores webhook signature verification data';
