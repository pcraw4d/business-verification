-- =====================================================
-- Unified Compliance Schema Migration
-- Task 2.3.2: Merge Compliance Tables
-- =====================================================
-- This migration consolidates compliance_checks and compliance_records
-- into a unified compliance_tracking table that supports both
-- compliance checking and record tracking functionality.

-- Create unified compliance tracking table
CREATE TABLE IF NOT EXISTS compliance_tracking (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Entity reference (aligned with Task 2.2 merchants consolidation)
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    
    -- Compliance identification
    compliance_type VARCHAR(100) NOT NULL,
    compliance_framework VARCHAR(100), -- e.g., AML, KYC, KYB, FATF, GDPR, PCI, etc.
    check_type VARCHAR(100) NOT NULL, -- e.g., 'automated', 'manual', 'periodic', 'ad_hoc'
    
    -- Compliance status and scoring
    status VARCHAR(50) NOT NULL CHECK (status IN (
        'pending', 'in_progress', 'completed', 'failed', 'overdue', 'expired', 'cancelled'
    )),
    score DECIMAL(5,4), -- Compliance score (0.0000 to 1.0000)
    risk_level VARCHAR(20) CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    
    -- Compliance details
    requirements JSONB, -- Structured compliance requirements
    check_method VARCHAR(100) NOT NULL, -- How the check was performed
    source VARCHAR(100) NOT NULL, -- Source of the compliance check
    raw_data JSONB, -- Raw data from external sources or internal systems
    
    -- Results and findings
    result JSONB, -- Detailed compliance check results
    findings JSONB, -- Specific findings from the compliance check
    recommendations JSONB, -- Recommendations for compliance improvement
    evidence JSONB, -- Evidence supporting compliance status
    
    -- Audit trail (enhanced from compliance_records)
    checked_by UUID REFERENCES users(id), -- User who performed the check
    checked_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reviewed_by UUID REFERENCES users(id), -- User who reviewed the results
    reviewed_at TIMESTAMP WITH TIME ZONE,
    approved_by UUID REFERENCES users(id), -- User who approved the compliance
    approved_at TIMESTAMP WITH TIME ZONE,
    
    -- Compliance lifecycle management
    due_date TIMESTAMP WITH TIME ZONE, -- When compliance is due
    expires_at TIMESTAMP WITH TIME ZONE, -- When compliance expires
    next_review_date TIMESTAMP WITH TIME ZONE, -- When next review is scheduled
    
    -- Priority and assignment
    priority VARCHAR(20) DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    assigned_to UUID REFERENCES users(id), -- User assigned to handle this compliance item
    
    -- Additional metadata
    tags TEXT[], -- Tags for categorization and filtering
    notes TEXT, -- Additional notes or comments
    metadata JSONB DEFAULT '{}', -- Additional metadata
    
    -- Standard timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT valid_score_range CHECK (score IS NULL OR (score >= 0.0000 AND score <= 1.0000)),
    CONSTRAINT valid_dates CHECK (
        (due_date IS NULL OR expires_at IS NULL OR due_date <= expires_at) AND
        (checked_at IS NULL OR reviewed_at IS NULL OR checked_at <= reviewed_at) AND
        (reviewed_at IS NULL OR approved_at IS NULL OR reviewed_at <= approved_at)
    )
);

-- Create comprehensive indexes for performance
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_merchant_id ON compliance_tracking(merchant_id);
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_compliance_type ON compliance_tracking(compliance_type);
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_compliance_framework ON compliance_tracking(compliance_framework);
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_status ON compliance_tracking(status);
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_risk_level ON compliance_tracking(risk_level);
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_priority ON compliance_tracking(priority);
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_checked_by ON compliance_tracking(checked_by);
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_assigned_to ON compliance_tracking(assigned_to);
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_due_date ON compliance_tracking(due_date);
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_expires_at ON compliance_tracking(expires_at);
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_next_review_date ON compliance_tracking(next_review_date);
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_created_at ON compliance_tracking(created_at);
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_updated_at ON compliance_tracking(updated_at);

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_merchant_status ON compliance_tracking(merchant_id, status);
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_merchant_type ON compliance_tracking(merchant_id, compliance_type);
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_status_priority ON compliance_tracking(status, priority);
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_due_overdue ON compliance_tracking(due_date) WHERE due_date < CURRENT_TIMESTAMP AND status NOT IN ('completed', 'cancelled');
CREATE INDEX IF NOT EXISTS idx_compliance_tracking_expiring ON compliance_tracking(expires_at) WHERE expires_at < CURRENT_TIMESTAMP + INTERVAL '30 days';

-- Create trigger for updated_at timestamp
CREATE OR REPLACE FUNCTION update_compliance_tracking_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_compliance_tracking_updated_at
    BEFORE UPDATE ON compliance_tracking
    FOR EACH ROW
    EXECUTE FUNCTION update_compliance_tracking_updated_at();

-- Create view for compliance summary (for reporting and dashboards)
CREATE OR REPLACE VIEW compliance_summary AS
SELECT 
    merchant_id,
    COUNT(*) as total_checks,
    COUNT(*) FILTER (WHERE status = 'completed') as completed_checks,
    COUNT(*) FILTER (WHERE status = 'pending') as pending_checks,
    COUNT(*) FILTER (WHERE status = 'failed') as failed_checks,
    COUNT(*) FILTER (WHERE status = 'overdue') as overdue_checks,
    COUNT(*) FILTER (WHERE due_date < CURRENT_TIMESTAMP AND status NOT IN ('completed', 'cancelled')) as past_due_checks,
    AVG(score) as average_score,
    MAX(checked_at) as last_check_date,
    MIN(next_review_date) as next_review_date,
    COUNT(DISTINCT compliance_type) as compliance_types_covered
FROM compliance_tracking
GROUP BY merchant_id;

-- Create view for compliance alerts (for monitoring and notifications)
CREATE OR REPLACE VIEW compliance_alerts AS
SELECT 
    id,
    merchant_id,
    compliance_type,
    compliance_framework,
    status,
    priority,
    risk_level,
    due_date,
    expires_at,
    assigned_to,
    created_at,
    updated_at,
    CASE 
        WHEN due_date < CURRENT_TIMESTAMP AND status NOT IN ('completed', 'cancelled') THEN 'overdue'
        WHEN expires_at < CURRENT_TIMESTAMP + INTERVAL '7 days' THEN 'expiring_soon'
        WHEN status = 'failed' THEN 'failed'
        WHEN priority = 'critical' AND status != 'completed' THEN 'critical_pending'
        ELSE NULL
    END as alert_type
FROM compliance_tracking
WHERE 
    (due_date < CURRENT_TIMESTAMP AND status NOT IN ('completed', 'cancelled')) OR
    (expires_at < CURRENT_TIMESTAMP + INTERVAL '7 days') OR
    (status = 'failed') OR
    (priority = 'critical' AND status != 'completed');

-- Add comments for documentation
COMMENT ON TABLE compliance_tracking IS 'Unified compliance tracking table consolidating compliance_checks and compliance_records functionality';
COMMENT ON COLUMN compliance_tracking.merchant_id IS 'Reference to the merchant (aligned with Task 2.2 consolidation)';
COMMENT ON COLUMN compliance_tracking.compliance_type IS 'Type of compliance check (e.g., AML, KYC, KYB, PCI, GDPR)';
COMMENT ON COLUMN compliance_tracking.compliance_framework IS 'Regulatory framework (e.g., FATF, SOX, ISO27001)';
COMMENT ON COLUMN compliance_tracking.check_type IS 'Method of compliance check (automated, manual, periodic, ad_hoc)';
COMMENT ON COLUMN compliance_tracking.score IS 'Compliance score from 0.0000 to 1.0000';
COMMENT ON COLUMN compliance_tracking.requirements IS 'Structured compliance requirements in JSON format';
COMMENT ON COLUMN compliance_tracking.result IS 'Detailed compliance check results';
COMMENT ON COLUMN compliance_tracking.findings IS 'Specific findings from the compliance check';
COMMENT ON COLUMN compliance_tracking.recommendations IS 'Recommendations for compliance improvement';
COMMENT ON COLUMN compliance_tracking.evidence IS 'Evidence supporting compliance status';
COMMENT ON COLUMN compliance_tracking.checked_by IS 'User who performed the compliance check';
COMMENT ON COLUMN compliance_tracking.reviewed_by IS 'User who reviewed the compliance results';
COMMENT ON COLUMN compliance_tracking.approved_by IS 'User who approved the compliance status';
COMMENT ON COLUMN compliance_tracking.due_date IS 'When compliance is due';
COMMENT ON COLUMN compliance_tracking.expires_at IS 'When compliance expires';
COMMENT ON COLUMN compliance_tracking.next_review_date IS 'When next review is scheduled';
COMMENT ON COLUMN compliance_tracking.priority IS 'Priority level for compliance handling';
COMMENT ON COLUMN compliance_tracking.assigned_to IS 'User assigned to handle this compliance item';
COMMENT ON COLUMN compliance_tracking.tags IS 'Tags for categorization and filtering';
COMMENT ON COLUMN compliance_tracking.metadata IS 'Additional metadata in JSON format';

COMMENT ON VIEW compliance_summary IS 'Summary view of compliance status per merchant';
COMMENT ON VIEW compliance_alerts IS 'View of compliance items requiring attention or alerts';

-- Enable Row Level Security (RLS) for data protection
ALTER TABLE compliance_tracking ENABLE ROW LEVEL SECURITY;

-- Create RLS policies (basic implementation - can be enhanced based on requirements)
CREATE POLICY compliance_tracking_policy ON compliance_tracking
    FOR ALL TO authenticated
    USING (true); -- Adjust based on your security requirements

-- Grant necessary permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON compliance_tracking TO authenticated;
GRANT SELECT ON compliance_summary TO authenticated;
GRANT SELECT ON compliance_alerts TO authenticated;
