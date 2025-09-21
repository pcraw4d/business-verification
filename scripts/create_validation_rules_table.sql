-- Create validation_rules table for crosswalk validation
-- This table stores validation rules for crosswalk mappings

CREATE TABLE IF NOT EXISTS validation_rules (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL CHECK (type IN ('format', 'consistency', 'business_logic', 'cross_reference')),
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    conditions JSONB NOT NULL DEFAULT '{}',
    action VARCHAR(20) NOT NULL CHECK (action IN ('warn', 'error', 'block', 'log')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_validation_rules_type ON validation_rules(type);
CREATE INDEX IF NOT EXISTS idx_validation_rules_severity ON validation_rules(severity);
CREATE INDEX IF NOT EXISTS idx_validation_rules_active ON validation_rules(is_active);
CREATE INDEX IF NOT EXISTS idx_validation_rules_conditions ON validation_rules USING GIN(conditions);

-- Create validation_results table to store validation execution results
CREATE TABLE IF NOT EXISTS validation_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rule_id VARCHAR(100) NOT NULL REFERENCES validation_rules(id),
    status VARCHAR(20) NOT NULL CHECK (status IN ('passed', 'failed', 'skipped', 'error')),
    message TEXT,
    details JSONB DEFAULT '{}',
    execution_time INTERVAL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for validation_results
CREATE INDEX IF NOT EXISTS idx_validation_results_rule_id ON validation_results(rule_id);
CREATE INDEX IF NOT EXISTS idx_validation_results_status ON validation_results(status);
CREATE INDEX IF NOT EXISTS idx_validation_results_created_at ON validation_results(created_at);
CREATE INDEX IF NOT EXISTS idx_validation_results_details ON validation_results USING GIN(details);

-- Create validation_summaries table to store validation run summaries
CREATE TABLE IF NOT EXISTS validation_summaries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    duration INTERVAL,
    total_rules INTEGER DEFAULT 0,
    passed_rules INTEGER DEFAULT 0,
    failed_rules INTEGER DEFAULT 0,
    skipped_rules INTEGER DEFAULT 0,
    error_rules INTEGER DEFAULT 0,
    issues_count INTEGER DEFAULT 0,
    summary_data JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for validation_summaries
CREATE INDEX IF NOT EXISTS idx_validation_summaries_start_time ON validation_summaries(start_time);
CREATE INDEX IF NOT EXISTS idx_validation_summaries_duration ON validation_summaries(duration);
CREATE INDEX IF NOT EXISTS idx_validation_summaries_failed_rules ON validation_summaries(failed_rules);
CREATE INDEX IF NOT EXISTS idx_validation_summaries_summary_data ON validation_summaries USING GIN(summary_data);

-- Create validation_issues table to store critical validation issues
CREATE TABLE IF NOT EXISTS validation_issues (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rule_id VARCHAR(100) NOT NULL REFERENCES validation_rules(id),
    rule_name VARCHAR(255) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    message TEXT NOT NULL,
    details JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'open' CHECK (status IN ('open', 'acknowledged', 'resolved', 'ignored')),
    assigned_to VARCHAR(255),
    resolution_notes TEXT,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for validation_issues
CREATE INDEX IF NOT EXISTS idx_validation_issues_rule_id ON validation_issues(rule_id);
CREATE INDEX IF NOT EXISTS idx_validation_issues_severity ON validation_issues(severity);
CREATE INDEX IF NOT EXISTS idx_validation_issues_status ON validation_issues(status);
CREATE INDEX IF NOT EXISTS idx_validation_issues_assigned_to ON validation_issues(assigned_to);
CREATE INDEX IF NOT EXISTS idx_validation_issues_created_at ON validation_issues(created_at);
CREATE INDEX IF NOT EXISTS idx_validation_issues_details ON validation_issues USING GIN(details);

-- Add comments to tables
COMMENT ON TABLE validation_rules IS 'Stores validation rules for crosswalk mappings';
COMMENT ON TABLE validation_results IS 'Stores individual validation rule execution results';
COMMENT ON TABLE validation_summaries IS 'Stores summaries of validation runs';
COMMENT ON TABLE validation_issues IS 'Stores critical validation issues that need attention';

-- Add comments to key columns
COMMENT ON COLUMN validation_rules.type IS 'Type of validation rule: format, consistency, business_logic, cross_reference';
COMMENT ON COLUMN validation_rules.severity IS 'Severity level: low, medium, high, critical';
COMMENT ON COLUMN validation_rules.conditions IS 'JSON conditions for the validation rule';
COMMENT ON COLUMN validation_rules.action IS 'Action to take when rule fails: warn, error, block, log';

COMMENT ON COLUMN validation_results.status IS 'Result status: passed, failed, skipped, error';
COMMENT ON COLUMN validation_results.details IS 'JSON details about the validation result';

COMMENT ON COLUMN validation_issues.status IS 'Issue status: open, acknowledged, resolved, ignored';
COMMENT ON COLUMN validation_issues.details IS 'JSON details about the validation issue';

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_validation_rules_updated_at 
    BEFORE UPDATE ON validation_rules 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_validation_issues_updated_at 
    BEFORE UPDATE ON validation_issues 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert sample validation rules (these will be created by the Go code, but we can add some defaults)
INSERT INTO validation_rules (id, name, description, type, severity, conditions, action, is_active) VALUES
('sample_format_rule', 'Sample Format Rule', 'Sample format validation rule', 'format', 'high', '{"pattern": "^[0-9]{4}$", "field": "mcc_code"}', 'error', true),
('sample_consistency_rule', 'Sample Consistency Rule', 'Sample consistency validation rule', 'consistency', 'medium', '{"min_value": 0.0, "max_value": 1.0, "field": "confidence_score"}', 'warn', true)
ON CONFLICT (id) DO NOTHING;

-- Create view for validation rule statistics
CREATE OR REPLACE VIEW validation_rule_stats AS
SELECT 
    vr.type,
    vr.severity,
    COUNT(*) as total_rules,
    COUNT(CASE WHEN vr.is_active THEN 1 END) as active_rules,
    COUNT(CASE WHEN NOT vr.is_active THEN 1 END) as inactive_rules,
    COUNT(CASE WHEN vr.action = 'error' THEN 1 END) as error_rules,
    COUNT(CASE WHEN vr.action = 'warn' THEN 1 END) as warn_rules,
    COUNT(CASE WHEN vr.action = 'block' THEN 1 END) as block_rules,
    COUNT(CASE WHEN vr.action = 'log' THEN 1 END) as log_rules
FROM validation_rules vr
GROUP BY vr.type, vr.severity
ORDER BY vr.type, vr.severity;

-- Create view for recent validation results
CREATE OR REPLACE VIEW recent_validation_results AS
SELECT 
    vr.name as rule_name,
    vr.type as rule_type,
    vr.severity as rule_severity,
    vres.status,
    vres.message,
    vres.execution_time,
    vres.created_at
FROM validation_results vres
JOIN validation_rules vr ON vres.rule_id = vr.id
WHERE vres.created_at >= NOW() - INTERVAL '7 days'
ORDER BY vres.created_at DESC;

-- Create view for open validation issues
CREATE OR REPLACE VIEW open_validation_issues AS
SELECT 
    vi.id,
    vi.rule_id,
    vi.rule_name,
    vi.severity,
    vi.message,
    vi.details,
    vi.assigned_to,
    vi.created_at,
    vi.updated_at
FROM validation_issues vi
WHERE vi.status = 'open'
ORDER BY 
    CASE vi.severity 
        WHEN 'critical' THEN 1 
        WHEN 'high' THEN 2 
        WHEN 'medium' THEN 3 
        WHEN 'low' THEN 4 
    END,
    vi.created_at DESC;

-- Grant permissions (adjust as needed for your environment)
-- GRANT SELECT, INSERT, UPDATE, DELETE ON validation_rules TO your_app_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON validation_results TO your_app_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON validation_summaries TO your_app_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON validation_issues TO your_app_user;
-- GRANT SELECT ON validation_rule_stats TO your_app_user;
-- GRANT SELECT ON recent_validation_results TO your_app_user;
-- GRANT SELECT ON open_validation_issues TO your_app_user;
