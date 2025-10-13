-- Migration: Create reports tables
-- Description: Creates tables for report generation, templates, and scheduling

CREATE TABLE IF NOT EXISTS reports (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('executive_summary', 'compliance', 'risk_audit', 'trend_analysis', 'custom', 'batch_results', 'performance')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'generating', 'completed', 'failed', 'expired')),
    format VARCHAR(20) NOT NULL CHECK (format IN ('pdf', 'excel', 'csv', 'json', 'html')),
    template_id VARCHAR(255),
    data JSONB,
    filters JSONB,
    generated_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    file_size BIGINT DEFAULT 0,
    download_url TEXT,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    error TEXT,
    
    -- Indexes for performance
    INDEX idx_reports_tenant_id (tenant_id),
    INDEX idx_reports_type (type),
    INDEX idx_reports_status (status),
    INDEX idx_reports_format (format),
    INDEX idx_reports_created_by (created_by),
    INDEX idx_reports_created_at (created_at),
    INDEX idx_reports_generated_at (generated_at),
    INDEX idx_reports_expires_at (expires_at),
    
    -- Constraints
    CONSTRAINT fk_reports_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create report templates table
CREATE TABLE IF NOT EXISTS report_templates (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('executive_summary', 'compliance', 'risk_audit', 'trend_analysis', 'custom', 'batch_results', 'performance')),
    description TEXT,
    template JSONB NOT NULL,
    is_public BOOLEAN DEFAULT FALSE,
    is_default BOOLEAN DEFAULT FALSE,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    
    -- Indexes
    INDEX idx_report_templates_tenant_id (tenant_id),
    INDEX idx_report_templates_type (type),
    INDEX idx_report_templates_is_public (is_public),
    INDEX idx_report_templates_is_default (is_default),
    INDEX idx_report_templates_created_by (created_by),
    INDEX idx_report_templates_created_at (created_at),
    
    -- Constraints
    CONSTRAINT fk_report_templates_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create scheduled reports table
CREATE TABLE IF NOT EXISTS scheduled_reports (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    template_id VARCHAR(255) NOT NULL,
    schedule JSONB NOT NULL,
    filters JSONB,
    recipients JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    last_run_at TIMESTAMP WITH TIME ZONE,
    next_run_at TIMESTAMP WITH TIME ZONE,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    
    -- Indexes
    INDEX idx_scheduled_reports_tenant_id (tenant_id),
    INDEX idx_scheduled_reports_template_id (template_id),
    INDEX idx_scheduled_reports_is_active (is_active),
    INDEX idx_scheduled_reports_next_run_at (next_run_at),
    INDEX idx_scheduled_reports_created_by (created_by),
    INDEX idx_scheduled_reports_created_at (created_at),
    
    -- Constraints
    CONSTRAINT fk_scheduled_reports_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_scheduled_reports_template FOREIGN KEY (template_id) REFERENCES report_templates(id) ON DELETE CASCADE
);

-- Create report downloads table for tracking download statistics
CREATE TABLE IF NOT EXISTS report_downloads (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    report_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    downloaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    user_id VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    
    -- Indexes
    INDEX idx_report_downloads_report_id (report_id),
    INDEX idx_report_downloads_tenant_id (tenant_id),
    INDEX idx_report_downloads_downloaded_at (downloaded_at),
    
    -- Constraints
    CONSTRAINT fk_report_downloads_report FOREIGN KEY (report_id) REFERENCES reports(id) ON DELETE CASCADE,
    CONSTRAINT fk_report_downloads_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create report shares table for sharing reports
CREATE TABLE IF NOT EXISTS report_shares (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    report_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    shared_with_tenant_id VARCHAR(255),
    shared_with_user_id VARCHAR(255),
    permission VARCHAR(20) DEFAULT 'read' CHECK (permission IN ('read', 'download')),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    
    -- Indexes
    INDEX idx_report_shares_report_id (report_id),
    INDEX idx_report_shares_tenant_id (tenant_id),
    INDEX idx_report_shares_shared_with (shared_with_tenant_id, shared_with_user_id),
    INDEX idx_report_shares_expires_at (expires_at),
    
    -- Constraints
    CONSTRAINT fk_report_shares_report FOREIGN KEY (report_id) REFERENCES reports(id) ON DELETE CASCADE,
    CONSTRAINT fk_report_shares_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_report_shares_shared_tenant FOREIGN KEY (shared_with_tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create report comments table
CREATE TABLE IF NOT EXISTS report_comments (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    report_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    comment TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_report_comments_report_id (report_id),
    INDEX idx_report_comments_tenant_id (tenant_id),
    INDEX idx_report_comments_user_id (user_id),
    INDEX idx_report_comments_created_at (created_at),
    
    -- Constraints
    CONSTRAINT fk_report_comments_report FOREIGN KEY (report_id) REFERENCES reports(id) ON DELETE CASCADE,
    CONSTRAINT fk_report_comments_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create report favorites table
CREATE TABLE IF NOT EXISTS report_favorites (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    report_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_report_favorites_report_id (report_id),
    INDEX idx_report_favorites_tenant_id (tenant_id),
    INDEX idx_report_favorites_user_id (user_id),
    
    -- Constraints
    CONSTRAINT fk_report_favorites_report FOREIGN KEY (report_id) REFERENCES reports(id) ON DELETE CASCADE,
    CONSTRAINT fk_report_favorites_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    UNIQUE(report_id, tenant_id, user_id)
);

-- Create report generation queue table for async processing
CREATE TABLE IF NOT EXISTS report_generation_queue (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    report_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    priority INTEGER DEFAULT 5,
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Indexes
    INDEX idx_report_generation_queue_report_id (report_id),
    INDEX idx_report_generation_queue_tenant_id (tenant_id),
    INDEX idx_report_generation_queue_status (status),
    INDEX idx_report_generation_queue_priority (priority),
    INDEX idx_report_generation_queue_created_at (created_at),
    
    -- Constraints
    CONSTRAINT fk_report_generation_queue_report FOREIGN KEY (report_id) REFERENCES reports(id) ON DELETE CASCADE,
    CONSTRAINT fk_report_generation_queue_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_reports_tenant_type ON reports(tenant_id, type);
CREATE INDEX IF NOT EXISTS idx_reports_tenant_status ON reports(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_reports_tenant_created_by ON reports(tenant_id, created_by);
CREATE INDEX IF NOT EXISTS idx_reports_tenant_created_at ON reports(tenant_id, created_at);

CREATE INDEX IF NOT EXISTS idx_report_templates_tenant_type ON report_templates(tenant_id, type);
CREATE INDEX IF NOT EXISTS idx_report_templates_tenant_public ON report_templates(tenant_id, is_public);

CREATE INDEX IF NOT EXISTS idx_scheduled_reports_tenant_active ON scheduled_reports(tenant_id, is_active);
CREATE INDEX IF NOT EXISTS idx_scheduled_reports_next_run ON scheduled_reports(next_run_at) WHERE is_active = true;

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_report_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_reports_updated_at
    BEFORE UPDATE ON reports
    FOR EACH ROW
    EXECUTE FUNCTION update_report_updated_at();

CREATE TRIGGER update_report_templates_updated_at
    BEFORE UPDATE ON report_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_report_updated_at();

CREATE TRIGGER update_scheduled_reports_updated_at
    BEFORE UPDATE ON scheduled_reports
    FOR EACH ROW
    EXECUTE FUNCTION update_report_updated_at();

CREATE TRIGGER update_report_comments_updated_at
    BEFORE UPDATE ON report_comments
    FOR EACH ROW
    EXECUTE FUNCTION update_report_updated_at();

-- Create function to clean up expired reports
CREATE OR REPLACE FUNCTION cleanup_expired_reports()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM reports 
    WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Create function to clean up old report downloads (older than 1 year)
CREATE OR REPLACE FUNCTION cleanup_old_report_downloads()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM report_downloads 
    WHERE downloaded_at < CURRENT_TIMESTAMP - INTERVAL '1 year';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Create function to clean up expired report shares
CREATE OR REPLACE FUNCTION cleanup_expired_report_shares()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM report_shares 
    WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Create function to clean up old report generation queue items (older than 7 days)
CREATE OR REPLACE FUNCTION cleanup_old_report_generation_queue()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM report_generation_queue 
    WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '7 days'
    AND status IN ('completed', 'failed');
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Insert sample report templates for testing
INSERT INTO report_templates (
    id, tenant_id, name, type, description, is_public, is_default, created_by
) VALUES (
    'executive_summary_template', 'default', 'Executive Summary Template', 'executive_summary', 
    'Standard executive summary report template with key metrics and insights', true, true, 'system'
), (
    'compliance_template', 'default', 'Compliance Report Template', 'compliance',
    'Comprehensive compliance report template with violation tracking', true, true, 'system'
), (
    'risk_audit_template', 'default', 'Risk Audit Template', 'risk_audit',
    'Detailed risk audit report template with findings and recommendations', true, true, 'system'
), (
    'trend_analysis_template', 'default', 'Trend Analysis Template', 'trend_analysis',
    'Risk trend analysis template with historical data and predictions', true, true, 'system'
), (
    'performance_template', 'default', 'Performance Metrics Template', 'performance',
    'System performance metrics template with optimization recommendations', true, true, 'system'
) ON CONFLICT (id) DO NOTHING;

-- Add comments to tables for documentation
COMMENT ON TABLE reports IS 'Stores generated reports with metadata and file information';
COMMENT ON TABLE report_templates IS 'Stores report templates for different report types';
COMMENT ON TABLE scheduled_reports IS 'Stores scheduled report configurations for automated generation';
COMMENT ON TABLE report_downloads IS 'Tracks report download statistics and access logs';
COMMENT ON TABLE report_shares IS 'Manages report sharing permissions between tenants and users';
COMMENT ON TABLE report_comments IS 'Stores user comments on reports';
COMMENT ON TABLE report_favorites IS 'Tracks user favorite reports';
COMMENT ON TABLE report_generation_queue IS 'Queue for asynchronous report generation processing';
