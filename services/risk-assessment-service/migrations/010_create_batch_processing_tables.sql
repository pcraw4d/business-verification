-- Migration: Create batch processing tables
-- Description: Creates tables for batch job management and processing
-- Version: 1.0.0
-- Date: 2025-01-19

-- Create batch_jobs table
CREATE TABLE IF NOT EXISTS batch_jobs (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    job_type VARCHAR(100) NOT NULL,
    total_requests INTEGER NOT NULL DEFAULT 0,
    completed INTEGER NOT NULL DEFAULT 0,
    failed INTEGER NOT NULL DEFAULT 0,
    progress DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    last_updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    priority INTEGER NOT NULL DEFAULT 5,
    metadata JSONB,
    results JSONB,
    error TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    
    -- Indexes
    CONSTRAINT batch_jobs_status_check CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled', 'paused')),
    CONSTRAINT batch_jobs_priority_check CHECK (priority >= 1 AND priority <= 10),
    CONSTRAINT batch_jobs_progress_check CHECK (progress >= 0.00 AND progress <= 100.00)
);

-- Create batch_results table
CREATE TABLE IF NOT EXISTS batch_results (
    id VARCHAR(255) PRIMARY KEY,
    job_id VARCHAR(255) NOT NULL,
    request_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    result_data JSONB,
    error_message TEXT,
    processing_time_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Foreign key constraint
    CONSTRAINT fk_batch_results_job_id FOREIGN KEY (job_id) REFERENCES batch_jobs(id) ON DELETE CASCADE,
    
    -- Indexes
    CONSTRAINT batch_results_status_check CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'skipped'))
);

-- Create batch_job_metrics table for performance tracking
CREATE TABLE IF NOT EXISTS batch_job_metrics (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    job_id VARCHAR(255) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(15,4) NOT NULL,
    metric_unit VARCHAR(50),
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Foreign key constraint
    CONSTRAINT fk_batch_job_metrics_job_id FOREIGN KEY (job_id) REFERENCES batch_jobs(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_batch_jobs_tenant_id ON batch_jobs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_batch_jobs_status ON batch_jobs(status);
CREATE INDEX IF NOT EXISTS idx_batch_jobs_created_at ON batch_jobs(created_at);
CREATE INDEX IF NOT EXISTS idx_batch_jobs_priority ON batch_jobs(priority);
CREATE INDEX IF NOT EXISTS idx_batch_jobs_tenant_status ON batch_jobs(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_batch_jobs_created_by ON batch_jobs(created_by);

CREATE INDEX IF NOT EXISTS idx_batch_results_job_id ON batch_results(job_id);
CREATE INDEX IF NOT EXISTS idx_batch_results_status ON batch_results(status);
CREATE INDEX IF NOT EXISTS idx_batch_results_created_at ON batch_results(created_at);
CREATE INDEX IF NOT EXISTS idx_batch_results_job_status ON batch_results(job_id, status);

CREATE INDEX IF NOT EXISTS idx_batch_job_metrics_tenant_id ON batch_job_metrics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_batch_job_metrics_job_id ON batch_job_metrics(job_id);
CREATE INDEX IF NOT EXISTS idx_batch_job_metrics_name ON batch_job_metrics(metric_name);
CREATE INDEX IF NOT EXISTS idx_batch_job_metrics_recorded_at ON batch_job_metrics(recorded_at);

-- Create composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_batch_jobs_tenant_created ON batch_jobs(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_batch_jobs_status_priority ON batch_jobs(status, priority DESC);
CREATE INDEX IF NOT EXISTS idx_batch_results_job_created ON batch_results(job_id, created_at DESC);

-- Add comments for documentation
COMMENT ON TABLE batch_jobs IS 'Stores batch job information and status';
COMMENT ON TABLE batch_results IS 'Stores individual results for batch job requests';
COMMENT ON TABLE batch_job_metrics IS 'Stores performance metrics for batch jobs';

COMMENT ON COLUMN batch_jobs.id IS 'Unique identifier for the batch job';
COMMENT ON COLUMN batch_jobs.tenant_id IS 'Tenant identifier for multi-tenancy';
COMMENT ON COLUMN batch_jobs.status IS 'Current status of the batch job';
COMMENT ON COLUMN batch_jobs.job_type IS 'Type of batch job (e.g., risk_assessment, compliance_check)';
COMMENT ON COLUMN batch_jobs.total_requests IS 'Total number of requests in the batch';
COMMENT ON COLUMN batch_jobs.completed IS 'Number of completed requests';
COMMENT ON COLUMN batch_jobs.failed IS 'Number of failed requests';
COMMENT ON COLUMN batch_jobs.progress IS 'Progress percentage (0-100)';
COMMENT ON COLUMN batch_jobs.priority IS 'Job priority (1-10, higher is more urgent)';
COMMENT ON COLUMN batch_jobs.metadata IS 'Additional job metadata in JSON format';
COMMENT ON COLUMN batch_jobs.results IS 'Job results summary in JSON format';
COMMENT ON COLUMN batch_jobs.retry_count IS 'Number of retry attempts';
COMMENT ON COLUMN batch_jobs.max_retries IS 'Maximum number of retry attempts';

COMMENT ON COLUMN batch_results.id IS 'Unique identifier for the batch result';
COMMENT ON COLUMN batch_results.job_id IS 'Reference to the batch job';
COMMENT ON COLUMN batch_results.request_id IS 'Unique identifier for the individual request';
COMMENT ON COLUMN batch_results.status IS 'Status of the individual request';
COMMENT ON COLUMN batch_results.result_data IS 'Result data in JSON format';
COMMENT ON COLUMN batch_results.processing_time_ms IS 'Processing time in milliseconds';

COMMENT ON COLUMN batch_job_metrics.id IS 'Unique identifier for the metric record';
COMMENT ON COLUMN batch_job_metrics.tenant_id IS 'Tenant identifier for multi-tenancy';
COMMENT ON COLUMN batch_job_metrics.job_id IS 'Reference to the batch job';
COMMENT ON COLUMN batch_job_metrics.metric_name IS 'Name of the metric (e.g., avg_processing_time, throughput)';
COMMENT ON COLUMN batch_job_metrics.metric_value IS 'Value of the metric';
COMMENT ON COLUMN batch_job_metrics.metric_unit IS 'Unit of measurement (e.g., ms, requests/sec)';
