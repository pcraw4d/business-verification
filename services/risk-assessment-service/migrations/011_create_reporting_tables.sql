-- Migration: Create reporting and dashboard tables
-- Description: Creates tables for reports, templates, scheduled reports, and dashboards
-- Version: 1.0.0
-- Date: 2025-01-19

-- Create reports table
CREATE TABLE IF NOT EXISTS reports (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    format VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    template_id VARCHAR(255),
    data JSONB,
    filters JSONB,
    recipients JSONB,
    file_size BIGINT,
    download_url TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    metadata JSONB,
    
    -- Constraints
    CONSTRAINT reports_type_check CHECK (type IN ('executive_summary', 'compliance', 'risk_audit', 'trend_analysis', 'custom', 'batch_results', 'performance')),
    CONSTRAINT reports_format_check CHECK (format IN ('pdf', 'excel', 'csv', 'json', 'html')),
    CONSTRAINT reports_status_check CHECK (status IN ('pending', 'generating', 'completed', 'failed', 'expired'))
);

-- Create report_templates table
CREATE TABLE IF NOT EXISTS report_templates (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    description TEXT,
    template JSONB NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    metadata JSONB,
    
    -- Constraints
    CONSTRAINT report_templates_type_check CHECK (type IN ('executive_summary', 'compliance', 'risk_audit', 'trend_analysis', 'custom', 'batch_results', 'performance'))
);

-- Create scheduled_reports table
CREATE TABLE IF NOT EXISTS scheduled_reports (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    template_id VARCHAR(255) NOT NULL,
    schedule JSONB NOT NULL,
    filters JSONB,
    recipients JSONB,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_run_at TIMESTAMP WITH TIME ZONE,
    next_run_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    metadata JSONB,
    
    -- Foreign key constraint
    CONSTRAINT fk_scheduled_reports_template_id FOREIGN KEY (template_id) REFERENCES report_templates(id) ON DELETE CASCADE
);

-- Create dashboards table
CREATE TABLE IF NOT EXISTS dashboards (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    summary JSONB,
    trends JSONB,
    predictions JSONB,
    charts JSONB,
    filters JSONB,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    metadata JSONB,
    
    -- Constraints
    CONSTRAINT dashboards_type_check CHECK (type IN ('risk_overview', 'trends', 'predictions', 'compliance', 'performance', 'custom'))
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_reports_tenant_id ON reports(tenant_id);
CREATE INDEX IF NOT EXISTS idx_reports_type ON reports(type);
CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status);
CREATE INDEX IF NOT EXISTS idx_reports_created_at ON reports(created_at);
CREATE INDEX IF NOT EXISTS idx_reports_tenant_type ON reports(tenant_id, type);
CREATE INDEX IF NOT EXISTS idx_reports_tenant_status ON reports(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_reports_created_by ON reports(created_by);
CREATE INDEX IF NOT EXISTS idx_reports_expires_at ON reports(expires_at);

CREATE INDEX IF NOT EXISTS idx_report_templates_tenant_id ON report_templates(tenant_id);
CREATE INDEX IF NOT EXISTS idx_report_templates_type ON report_templates(type);
CREATE INDEX IF NOT EXISTS idx_report_templates_is_public ON report_templates(is_public);
CREATE INDEX IF NOT EXISTS idx_report_templates_is_default ON report_templates(is_default);
CREATE INDEX IF NOT EXISTS idx_report_templates_created_at ON report_templates(created_at);
CREATE INDEX IF NOT EXISTS idx_report_templates_tenant_type ON report_templates(tenant_id, type);

CREATE INDEX IF NOT EXISTS idx_scheduled_reports_tenant_id ON scheduled_reports(tenant_id);
CREATE INDEX IF NOT EXISTS idx_scheduled_reports_is_active ON scheduled_reports(is_active);
CREATE INDEX IF NOT EXISTS idx_scheduled_reports_next_run_at ON scheduled_reports(next_run_at);
CREATE INDEX IF NOT EXISTS idx_scheduled_reports_created_at ON scheduled_reports(created_at);
CREATE INDEX IF NOT EXISTS idx_scheduled_reports_tenant_active ON scheduled_reports(tenant_id, is_active);

CREATE INDEX IF NOT EXISTS idx_dashboards_tenant_id ON dashboards(tenant_id);
CREATE INDEX IF NOT EXISTS idx_dashboards_type ON dashboards(type);
CREATE INDEX IF NOT EXISTS idx_dashboards_is_public ON dashboards(is_public);
CREATE INDEX IF NOT EXISTS idx_dashboards_created_at ON dashboards(created_at);
CREATE INDEX IF NOT EXISTS idx_dashboards_tenant_type ON dashboards(tenant_id, type);

-- Create composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_reports_tenant_created ON reports(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reports_status_created ON reports(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_report_templates_tenant_created ON report_templates(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_scheduled_reports_next_run ON scheduled_reports(next_run_at) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_dashboards_tenant_created ON dashboards(tenant_id, created_at DESC);

-- Add comments for documentation
COMMENT ON TABLE reports IS 'Stores generated reports and their metadata';
COMMENT ON TABLE report_templates IS 'Stores report templates for generating reports';
COMMENT ON TABLE scheduled_reports IS 'Stores scheduled report configurations';
COMMENT ON TABLE dashboards IS 'Stores dashboard configurations and data';

COMMENT ON COLUMN reports.id IS 'Unique identifier for the report';
COMMENT ON COLUMN reports.tenant_id IS 'Tenant identifier for multi-tenancy';
COMMENT ON COLUMN reports.type IS 'Type of report (executive_summary, compliance, etc.)';
COMMENT ON COLUMN reports.format IS 'Output format (pdf, excel, csv, json, html)';
COMMENT ON COLUMN reports.status IS 'Current status of the report';
COMMENT ON COLUMN reports.template_id IS 'Reference to the template used';
COMMENT ON COLUMN reports.data IS 'Report data in JSON format';
COMMENT ON COLUMN reports.filters IS 'Filters applied to the report in JSON format';
COMMENT ON COLUMN reports.recipients IS 'Report recipients in JSON format';
COMMENT ON COLUMN reports.file_size IS 'Size of the generated report file in bytes';
COMMENT ON COLUMN reports.download_url IS 'URL to download the report file';
COMMENT ON COLUMN reports.expires_at IS 'When the report expires and should be deleted';

COMMENT ON COLUMN report_templates.id IS 'Unique identifier for the template';
COMMENT ON COLUMN report_templates.tenant_id IS 'Tenant identifier for multi-tenancy';
COMMENT ON COLUMN report_templates.type IS 'Type of report this template is for';
COMMENT ON COLUMN report_templates.template IS 'Template configuration in JSON format';
COMMENT ON COLUMN report_templates.is_public IS 'Whether the template is publicly available';
COMMENT ON COLUMN report_templates.is_default IS 'Whether this is a default template';

COMMENT ON COLUMN scheduled_reports.id IS 'Unique identifier for the scheduled report';
COMMENT ON COLUMN scheduled_reports.tenant_id IS 'Tenant identifier for multi-tenancy';
COMMENT ON COLUMN scheduled_reports.template_id IS 'Reference to the template to use';
COMMENT ON COLUMN scheduled_reports.schedule IS 'Schedule configuration in JSON format';
COMMENT ON COLUMN scheduled_reports.is_active IS 'Whether the scheduled report is active';
COMMENT ON COLUMN scheduled_reports.last_run_at IS 'When the report was last executed';
COMMENT ON COLUMN scheduled_reports.next_run_at IS 'When the report should be executed next';

COMMENT ON COLUMN dashboards.id IS 'Unique identifier for the dashboard';
COMMENT ON COLUMN dashboards.tenant_id IS 'Tenant identifier for multi-tenancy';
COMMENT ON COLUMN dashboards.type IS 'Type of dashboard (risk_overview, trends, etc.)';
COMMENT ON COLUMN dashboards.summary IS 'Dashboard summary data in JSON format';
COMMENT ON COLUMN dashboards.trends IS 'Dashboard trends data in JSON format';
COMMENT ON COLUMN dashboards.predictions IS 'Dashboard predictions data in JSON format';
COMMENT ON COLUMN dashboards.charts IS 'Dashboard charts configuration in JSON format';
COMMENT ON COLUMN dashboards.filters IS 'Dashboard filters in JSON format';
COMMENT ON COLUMN dashboards.is_public IS 'Whether the dashboard is publicly accessible';
