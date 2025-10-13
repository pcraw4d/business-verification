-- Migration: Create dashboards table
-- Description: Creates the dashboards table for storing dashboard configurations and data

CREATE TABLE IF NOT EXISTS dashboards (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('risk_overview', 'trends', 'predictions', 'compliance', 'performance', 'custom')),
    summary JSONB,
    trends JSONB,
    predictions JSONB,
    charts JSONB,
    filters JSONB,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    metadata JSONB,
    
    -- Indexes for performance
    INDEX idx_dashboards_tenant_id (tenant_id),
    INDEX idx_dashboards_type (type),
    INDEX idx_dashboards_created_by (created_by),
    INDEX idx_dashboards_is_public (is_public),
    INDEX idx_dashboards_created_at (created_at),
    
    -- Constraints
    CONSTRAINT fk_dashboards_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create dashboard views table for tracking usage
CREATE TABLE IF NOT EXISTS dashboard_views (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    dashboard_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    viewed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    user_id VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    
    -- Indexes
    INDEX idx_dashboard_views_dashboard_id (dashboard_id),
    INDEX idx_dashboard_views_tenant_id (tenant_id),
    INDEX idx_dashboard_views_viewed_at (viewed_at),
    
    -- Constraints
    CONSTRAINT fk_dashboard_views_dashboard FOREIGN KEY (dashboard_id) REFERENCES dashboards(id) ON DELETE CASCADE,
    CONSTRAINT fk_dashboard_views_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create dashboard shares table for sharing dashboards
CREATE TABLE IF NOT EXISTS dashboard_shares (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    dashboard_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    shared_with_tenant_id VARCHAR(255),
    shared_with_user_id VARCHAR(255),
    permission VARCHAR(20) DEFAULT 'read' CHECK (permission IN ('read', 'write', 'admin')),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    
    -- Indexes
    INDEX idx_dashboard_shares_dashboard_id (dashboard_id),
    INDEX idx_dashboard_shares_tenant_id (tenant_id),
    INDEX idx_dashboard_shares_shared_with (shared_with_tenant_id, shared_with_user_id),
    INDEX idx_dashboard_shares_expires_at (expires_at),
    
    -- Constraints
    CONSTRAINT fk_dashboard_shares_dashboard FOREIGN KEY (dashboard_id) REFERENCES dashboards(id) ON DELETE CASCADE,
    CONSTRAINT fk_dashboard_shares_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_dashboard_shares_shared_tenant FOREIGN KEY (shared_with_tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create dashboard favorites table
CREATE TABLE IF NOT EXISTS dashboard_favorites (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    dashboard_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_dashboard_favorites_dashboard_id (dashboard_id),
    INDEX idx_dashboard_favorites_tenant_id (tenant_id),
    INDEX idx_dashboard_favorites_user_id (user_id),
    
    -- Constraints
    CONSTRAINT fk_dashboard_favorites_dashboard FOREIGN KEY (dashboard_id) REFERENCES dashboards(id) ON DELETE CASCADE,
    CONSTRAINT fk_dashboard_favorites_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    UNIQUE(dashboard_id, tenant_id, user_id)
);

-- Create dashboard comments table
CREATE TABLE IF NOT EXISTS dashboard_comments (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    dashboard_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    comment TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_dashboard_comments_dashboard_id (dashboard_id),
    INDEX idx_dashboard_comments_tenant_id (tenant_id),
    INDEX idx_dashboard_comments_user_id (user_id),
    INDEX idx_dashboard_comments_created_at (created_at),
    
    -- Constraints
    CONSTRAINT fk_dashboard_comments_dashboard FOREIGN KEY (dashboard_id) REFERENCES dashboards(id) ON DELETE CASCADE,
    CONSTRAINT fk_dashboard_comments_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create dashboard snapshots table for versioning
CREATE TABLE IF NOT EXISTS dashboard_snapshots (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    dashboard_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    version INTEGER NOT NULL,
    name VARCHAR(255),
    summary JSONB,
    trends JSONB,
    predictions JSONB,
    charts JSONB,
    filters JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    
    -- Indexes
    INDEX idx_dashboard_snapshots_dashboard_id (dashboard_id),
    INDEX idx_dashboard_snapshots_tenant_id (tenant_id),
    INDEX idx_dashboard_snapshots_version (version),
    INDEX idx_dashboard_snapshots_created_at (created_at),
    
    -- Constraints
    CONSTRAINT fk_dashboard_snapshots_dashboard FOREIGN KEY (dashboard_id) REFERENCES dashboards(id) ON DELETE CASCADE,
    CONSTRAINT fk_dashboard_snapshots_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    UNIQUE(dashboard_id, version)
);

-- Create dashboard alerts table for monitoring
CREATE TABLE IF NOT EXISTS dashboard_alerts (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    dashboard_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    condition JSONB NOT NULL, -- Alert condition configuration
    threshold JSONB NOT NULL, -- Alert threshold values
    is_enabled BOOLEAN DEFAULT TRUE,
    last_triggered TIMESTAMP WITH TIME ZONE,
    trigger_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    
    -- Indexes
    INDEX idx_dashboard_alerts_dashboard_id (dashboard_id),
    INDEX idx_dashboard_alerts_tenant_id (tenant_id),
    INDEX idx_dashboard_alerts_is_enabled (is_enabled),
    INDEX idx_dashboard_alerts_last_triggered (last_triggered),
    
    -- Constraints
    CONSTRAINT fk_dashboard_alerts_dashboard FOREIGN KEY (dashboard_id) REFERENCES dashboards(id) ON DELETE CASCADE,
    CONSTRAINT fk_dashboard_alerts_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create dashboard alert notifications table
CREATE TABLE IF NOT EXISTS dashboard_alert_notifications (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    alert_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    notification_type VARCHAR(50) NOT NULL CHECK (notification_type IN ('email', 'webhook', 'slack', 'teams')),
    recipient VARCHAR(255) NOT NULL, -- Email, webhook URL, etc.
    subject VARCHAR(255),
    message TEXT,
    sent_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed', 'delivered')),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_dashboard_alert_notifications_alert_id (alert_id),
    INDEX idx_dashboard_alert_notifications_tenant_id (tenant_id),
    INDEX idx_dashboard_alert_notifications_status (status),
    INDEX idx_dashboard_alert_notifications_sent_at (sent_at),
    
    -- Constraints
    CONSTRAINT fk_dashboard_alert_notifications_alert FOREIGN KEY (alert_id) REFERENCES dashboard_alerts(id) ON DELETE CASCADE,
    CONSTRAINT fk_dashboard_alert_notifications_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_dashboards_tenant_type ON dashboards(tenant_id, type);
CREATE INDEX IF NOT EXISTS idx_dashboards_tenant_created_by ON dashboards(tenant_id, created_by);
CREATE INDEX IF NOT EXISTS idx_dashboards_tenant_updated_at ON dashboards(tenant_id, updated_at);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_dashboard_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_dashboards_updated_at
    BEFORE UPDATE ON dashboards
    FOR EACH ROW
    EXECUTE FUNCTION update_dashboard_updated_at();

-- Create function to clean up old dashboard views (older than 90 days)
CREATE OR REPLACE FUNCTION cleanup_old_dashboard_views()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM dashboard_views 
    WHERE viewed_at < CURRENT_TIMESTAMP - INTERVAL '90 days';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Create function to clean up expired dashboard shares
CREATE OR REPLACE FUNCTION cleanup_expired_dashboard_shares()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM dashboard_shares 
    WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Insert sample dashboard data for testing
INSERT INTO dashboards (
    id, tenant_id, name, type, created_by, is_public
) VALUES (
    'sample_risk_overview', 'default', 'Risk Overview Dashboard', 'risk_overview', 'system', true
), (
    'sample_trends', 'default', 'Risk Trends Dashboard', 'trends', 'system', true
), (
    'sample_predictions', 'default', 'Risk Predictions Dashboard', 'predictions', 'system', true
) ON CONFLICT (id) DO NOTHING;

-- Add comments to tables for documentation
COMMENT ON TABLE dashboards IS 'Stores dashboard configurations and data for risk assessment reporting';
COMMENT ON TABLE dashboard_views IS 'Tracks dashboard view statistics for analytics';
COMMENT ON TABLE dashboard_shares IS 'Manages dashboard sharing permissions between tenants and users';
COMMENT ON TABLE dashboard_favorites IS 'Tracks user favorite dashboards';
COMMENT ON TABLE dashboard_comments IS 'Stores user comments on dashboards';
COMMENT ON TABLE dashboard_snapshots IS 'Version history for dashboard configurations';
COMMENT ON TABLE dashboard_alerts IS 'Dashboard alert configurations for monitoring';
COMMENT ON TABLE dashboard_alert_notifications IS 'Alert notification delivery tracking';
