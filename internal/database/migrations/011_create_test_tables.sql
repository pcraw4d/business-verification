-- Migration: Create Missing Test Tables
-- Date: January 2025
-- Purpose: Create merchant_analytics, risk_indicators, enrichment_jobs, and enrichment_sources tables
--          for comprehensive testing of Weeks 2-4 features

BEGIN;

-- ============================================================================
-- merchant_analytics Table
-- ============================================================================
-- Stores analytics data for merchants including classification, security, quality, and intelligence data
CREATE TABLE IF NOT EXISTS merchant_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id VARCHAR(255) NOT NULL,
    
    -- Analytics data stored as JSONB for flexibility
    classification_data JSONB DEFAULT '{}',
    security_data JSONB DEFAULT '{}',
    quality_data JSONB DEFAULT '{}',
    intelligence_data JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for merchant lookups
CREATE INDEX IF NOT EXISTS idx_merchant_analytics_merchant_id ON merchant_analytics(merchant_id);

-- Index for JSONB queries (GIN index for efficient JSONB searches)
CREATE INDEX IF NOT EXISTS idx_merchant_analytics_classification ON merchant_analytics USING GIN (classification_data);
CREATE INDEX IF NOT EXISTS idx_merchant_analytics_security ON merchant_analytics USING GIN (security_data);

-- ============================================================================
-- risk_indicators Table
-- ============================================================================
-- Stores individual risk indicators detected for merchants
CREATE TABLE IF NOT EXISTS risk_indicators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id VARCHAR(255) NOT NULL,
    
    -- Indicator details
    type VARCHAR(100) NOT NULL, -- e.g., 'compliance', 'financial', 'operational', 'reputation'
    name VARCHAR(255) NOT NULL,
    severity VARCHAR(50) NOT NULL DEFAULT 'medium', -- low, medium, high, critical
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, resolved, dismissed
    description TEXT,
    
    -- Scoring and detection
    score DECIMAL(5,2) DEFAULT 0.0, -- 0.00 to 100.00
    detected_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for risk_indicators
CREATE INDEX IF NOT EXISTS idx_risk_indicators_merchant_id ON risk_indicators(merchant_id);
CREATE INDEX IF NOT EXISTS idx_risk_indicators_severity ON risk_indicators(severity);
CREATE INDEX IF NOT EXISTS idx_risk_indicators_status ON risk_indicators(status);
CREATE INDEX IF NOT EXISTS idx_risk_indicators_type ON risk_indicators(type);
CREATE INDEX IF NOT EXISTS idx_risk_indicators_detected_at ON risk_indicators(detected_at DESC);

-- Composite index for common queries
CREATE INDEX IF NOT EXISTS idx_risk_indicators_merchant_status ON risk_indicators(merchant_id, status);

-- ============================================================================
-- enrichment_jobs Table
-- ============================================================================
-- Tracks data enrichment jobs for merchants
CREATE TABLE IF NOT EXISTS enrichment_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id VARCHAR(255) NOT NULL UNIQUE,
    merchant_id VARCHAR(255) NOT NULL,
    source VARCHAR(100) NOT NULL, -- e.g., 'thomson-reuters', 'dun-bradstreet', 'government-registry'
    
    -- Job status
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed
    progress INTEGER DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
    
    -- Job data
    request_data JSONB DEFAULT '{}',
    result_data JSONB DEFAULT '{}',
    error_message TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for enrichment_jobs
CREATE INDEX IF NOT EXISTS idx_enrichment_jobs_job_id ON enrichment_jobs(job_id);
CREATE INDEX IF NOT EXISTS idx_enrichment_jobs_merchant_id ON enrichment_jobs(merchant_id);
CREATE INDEX IF NOT EXISTS idx_enrichment_jobs_status ON enrichment_jobs(status);
CREATE INDEX IF NOT EXISTS idx_enrichment_jobs_source ON enrichment_jobs(source);
CREATE INDEX IF NOT EXISTS idx_enrichment_jobs_created_at ON enrichment_jobs(created_at DESC);

-- Composite index for common queries
CREATE INDEX IF NOT EXISTS idx_enrichment_jobs_merchant_status ON enrichment_jobs(merchant_id, status);

-- ============================================================================
-- enrichment_sources Table
-- ============================================================================
-- Defines available enrichment sources and their configuration
CREATE TABLE IF NOT EXISTS enrichment_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id VARCHAR(100) NOT NULL UNIQUE, -- e.g., 'thomson-reuters', 'dun-bradstreet'
    
    -- Source details
    name VARCHAR(255) NOT NULL,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT true,
    
    -- Configuration
    config JSONB DEFAULT '{}', -- Source-specific configuration
    rate_limit_per_minute INTEGER DEFAULT 60,
    rate_limit_per_day INTEGER DEFAULT 1000,
    
    -- Metadata
    last_used_at TIMESTAMP WITH TIME ZONE,
    usage_count INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for enrichment_sources
CREATE INDEX IF NOT EXISTS idx_enrichment_sources_source_id ON enrichment_sources(source_id);
CREATE INDEX IF NOT EXISTS idx_enrichment_sources_enabled ON enrichment_sources(enabled);

-- Insert default enrichment sources
INSERT INTO enrichment_sources (source_id, name, description, enabled) VALUES
    ('thomson-reuters', 'Thomson Reuters', 'Business intelligence and compliance data', true),
    ('dun-bradstreet', 'Dun & Bradstreet', 'Business credit and company data', true),
    ('government-registry', 'Government Registry', 'Official business registration data', true)
ON CONFLICT (source_id) DO NOTHING;

COMMIT;

-- ============================================================================
-- Verification Queries
-- ============================================================================
-- Run these queries to verify the tables were created successfully:

-- SELECT COUNT(*) FROM merchant_analytics;
-- SELECT COUNT(*) FROM risk_indicators;
-- SELECT COUNT(*) FROM enrichment_jobs;
-- SELECT COUNT(*) FROM enrichment_sources;

