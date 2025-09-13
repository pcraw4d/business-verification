-- Migration: 005_merchant_portfolio_schema.sql
-- Description: Add merchant portfolio management tables and schema
-- Created: 2025-01-19
-- Dependencies: 001_initial_schema.sql, 002_rbac_schema.sql

-- Portfolio types lookup table
CREATE TABLE IF NOT EXISTS portfolio_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) UNIQUE NOT NULL CHECK (type IN ('onboarded', 'deactivated', 'prospective', 'pending')),
    description TEXT,
    display_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Risk levels lookup table
CREATE TABLE IF NOT EXISTS risk_levels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    level VARCHAR(50) UNIQUE NOT NULL CHECK (level IN ('high', 'medium', 'low')),
    description TEXT,
    numeric_value INTEGER NOT NULL,
    color_code VARCHAR(7), -- Hex color code for UI
    display_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Merchants table (enhanced version of businesses table for portfolio management)
CREATE TABLE IF NOT EXISTS merchants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    legal_name VARCHAR(255) NOT NULL,
    registration_number VARCHAR(100) UNIQUE NOT NULL,
    tax_id VARCHAR(100),
    industry VARCHAR(100),
    industry_code VARCHAR(20),
    business_type VARCHAR(50),
    founded_date DATE,
    employee_count INTEGER,
    annual_revenue DECIMAL(15,2),
    
    -- Address fields (flattened for better query performance)
    address_street1 VARCHAR(255),
    address_street2 VARCHAR(255),
    address_city VARCHAR(100),
    address_state VARCHAR(100),
    address_postal_code VARCHAR(20),
    address_country VARCHAR(100),
    address_country_code VARCHAR(10),
    
    -- Contact info fields (flattened for better query performance)
    contact_phone VARCHAR(50),
    contact_email VARCHAR(255),
    contact_website VARCHAR(255),
    contact_primary_contact VARCHAR(255),
    
    -- Portfolio management fields
    portfolio_type_id UUID NOT NULL REFERENCES portfolio_types(id),
    risk_level_id UUID NOT NULL REFERENCES risk_levels(id),
    compliance_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    
    -- Audit fields
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Merchant sessions table for single merchant session management
CREATE TABLE IF NOT EXISTS merchant_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_active TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true,
    session_data JSONB, -- Store additional session context
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Enhanced audit logs table for merchant operations
CREATE TABLE IF NOT EXISTS merchant_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    merchant_id UUID REFERENCES merchants(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(100) NOT NULL,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(100),
    session_id UUID REFERENCES merchant_sessions(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Compliance records table for merchant compliance tracking
CREATE TABLE IF NOT EXISTS compliance_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    compliance_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    score DECIMAL(5,4) NOT NULL,
    requirements JSONB,
    check_method VARCHAR(100) NOT NULL,
    source VARCHAR(100) NOT NULL,
    raw_data JSONB,
    checked_by UUID REFERENCES users(id),
    checked_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Merchant analytics table for storing calculated analytics data
CREATE TABLE IF NOT EXISTS merchant_analytics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    risk_score DECIMAL(5,4) NOT NULL,
    compliance_score DECIMAL(5,4) NOT NULL,
    transaction_volume DECIMAL(15,2),
    last_activity TIMESTAMP WITH TIME ZONE,
    flags TEXT[], -- Array of flag strings
    metadata JSONB,
    calculated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Merchant notifications table
CREATE TABLE IF NOT EXISTS merchant_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('risk_alert', 'compliance', 'status_change', 'bulk_operation', 'system')),
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT false,
    priority VARCHAR(20) NOT NULL CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Merchant comparisons table for storing comparison data
CREATE TABLE IF NOT EXISTS merchant_comparisons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    merchant1_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    merchant2_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comparison_data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure merchants are different
    CONSTRAINT check_different_merchants CHECK (merchant1_id != merchant2_id)
);

-- Bulk operations table for tracking bulk operations
CREATE TABLE IF NOT EXISTS bulk_operations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    operation_id VARCHAR(100) UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    operation_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    total_items INTEGER NOT NULL DEFAULT 0,
    processed INTEGER NOT NULL DEFAULT 0,
    successful INTEGER NOT NULL DEFAULT 0,
    failed INTEGER NOT NULL DEFAULT 0,
    errors TEXT[],
    results JSONB,
    metadata JSONB,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Bulk operation items table for individual item tracking
CREATE TABLE IF NOT EXISTS bulk_operation_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bulk_operation_id UUID NOT NULL REFERENCES bulk_operations(id) ON DELETE CASCADE,
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'skipped')),
    error_message TEXT,
    result_data JSONB,
    processed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance optimization

-- Portfolio types indexes
CREATE INDEX IF NOT EXISTS idx_portfolio_types_type ON portfolio_types(type);
CREATE INDEX IF NOT EXISTS idx_portfolio_types_active ON portfolio_types(is_active);
CREATE INDEX IF NOT EXISTS idx_portfolio_types_display_order ON portfolio_types(display_order);

-- Risk levels indexes
CREATE INDEX IF NOT EXISTS idx_risk_levels_level ON risk_levels(level);
CREATE INDEX IF NOT EXISTS idx_risk_levels_numeric_value ON risk_levels(numeric_value);
CREATE INDEX IF NOT EXISTS idx_risk_levels_active ON risk_levels(is_active);
CREATE INDEX IF NOT EXISTS idx_risk_levels_display_order ON risk_levels(display_order);

-- Merchants indexes
CREATE INDEX IF NOT EXISTS idx_merchants_registration_number ON merchants(registration_number);
CREATE INDEX IF NOT EXISTS idx_merchants_tax_id ON merchants(tax_id);
CREATE INDEX IF NOT EXISTS idx_merchants_industry ON merchants(industry);
CREATE INDEX IF NOT EXISTS idx_merchants_industry_code ON merchants(industry_code);
CREATE INDEX IF NOT EXISTS idx_merchants_status ON merchants(status);
CREATE INDEX IF NOT EXISTS idx_merchants_portfolio_type_id ON merchants(portfolio_type_id);
CREATE INDEX IF NOT EXISTS idx_merchants_risk_level_id ON merchants(risk_level_id);
CREATE INDEX IF NOT EXISTS idx_merchants_compliance_status ON merchants(compliance_status);
CREATE INDEX IF NOT EXISTS idx_merchants_created_by ON merchants(created_by);
CREATE INDEX IF NOT EXISTS idx_merchants_created_at ON merchants(created_at);
CREATE INDEX IF NOT EXISTS idx_merchants_updated_at ON merchants(updated_at);

-- Search indexes for merchant search functionality
CREATE INDEX IF NOT EXISTS idx_merchants_name_trgm ON merchants USING gin(name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_merchants_legal_name_trgm ON merchants USING gin(legal_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_merchants_contact_email ON merchants(contact_email);
CREATE INDEX IF NOT EXISTS idx_merchants_contact_phone ON merchants(contact_phone);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_merchants_portfolio_risk ON merchants(portfolio_type_id, risk_level_id);
CREATE INDEX IF NOT EXISTS idx_merchants_status_compliance ON merchants(status, compliance_status);
CREATE INDEX IF NOT EXISTS idx_merchants_created_by_status ON merchants(created_by, status);

-- Merchant sessions indexes
CREATE INDEX IF NOT EXISTS idx_merchant_sessions_user_id ON merchant_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_merchant_sessions_merchant_id ON merchant_sessions(merchant_id);
CREATE INDEX IF NOT EXISTS idx_merchant_sessions_is_active ON merchant_sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_merchant_sessions_last_active ON merchant_sessions(last_active);
CREATE INDEX IF NOT EXISTS idx_merchant_sessions_user_active ON merchant_sessions(user_id, is_active);

-- Merchant audit logs indexes
CREATE INDEX IF NOT EXISTS idx_merchant_audit_logs_user_id ON merchant_audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_merchant_audit_logs_merchant_id ON merchant_audit_logs(merchant_id);
CREATE INDEX IF NOT EXISTS idx_merchant_audit_logs_action ON merchant_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_merchant_audit_logs_resource_type ON merchant_audit_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_merchant_audit_logs_resource_id ON merchant_audit_logs(resource_id);
CREATE INDEX IF NOT EXISTS idx_merchant_audit_logs_created_at ON merchant_audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_merchant_audit_logs_request_id ON merchant_audit_logs(request_id);
CREATE INDEX IF NOT EXISTS idx_merchant_audit_logs_session_id ON merchant_audit_logs(session_id);

-- Compliance records indexes
CREATE INDEX IF NOT EXISTS idx_compliance_records_merchant_id ON compliance_records(merchant_id);
CREATE INDEX IF NOT EXISTS idx_compliance_records_compliance_type ON compliance_records(compliance_type);
CREATE INDEX IF NOT EXISTS idx_compliance_records_status ON compliance_records(status);
CREATE INDEX IF NOT EXISTS idx_compliance_records_score ON compliance_records(score);
CREATE INDEX IF NOT EXISTS idx_compliance_records_checked_by ON compliance_records(checked_by);
CREATE INDEX IF NOT EXISTS idx_compliance_records_checked_at ON compliance_records(checked_at);
CREATE INDEX IF NOT EXISTS idx_compliance_records_expires_at ON compliance_records(expires_at);

-- Merchant analytics indexes
CREATE INDEX IF NOT EXISTS idx_merchant_analytics_merchant_id ON merchant_analytics(merchant_id);
CREATE INDEX IF NOT EXISTS idx_merchant_analytics_risk_score ON merchant_analytics(risk_score);
CREATE INDEX IF NOT EXISTS idx_merchant_analytics_compliance_score ON merchant_analytics(compliance_score);
CREATE INDEX IF NOT EXISTS idx_merchant_analytics_calculated_at ON merchant_analytics(calculated_at);

-- Merchant notifications indexes
CREATE INDEX IF NOT EXISTS idx_merchant_notifications_merchant_id ON merchant_notifications(merchant_id);
CREATE INDEX IF NOT EXISTS idx_merchant_notifications_user_id ON merchant_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_merchant_notifications_type ON merchant_notifications(type);
CREATE INDEX IF NOT EXISTS idx_merchant_notifications_is_read ON merchant_notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_merchant_notifications_priority ON merchant_notifications(priority);
CREATE INDEX IF NOT EXISTS idx_merchant_notifications_created_at ON merchant_notifications(created_at);
CREATE INDEX IF NOT EXISTS idx_merchant_notifications_user_unread ON merchant_notifications(user_id, is_read);

-- Merchant comparisons indexes
CREATE INDEX IF NOT EXISTS idx_merchant_comparisons_merchant1_id ON merchant_comparisons(merchant1_id);
CREATE INDEX IF NOT EXISTS idx_merchant_comparisons_merchant2_id ON merchant_comparisons(merchant2_id);
CREATE INDEX IF NOT EXISTS idx_merchant_comparisons_user_id ON merchant_comparisons(user_id);
CREATE INDEX IF NOT EXISTS idx_merchant_comparisons_created_at ON merchant_comparisons(created_at);

-- Bulk operations indexes
CREATE INDEX IF NOT EXISTS idx_bulk_operations_operation_id ON bulk_operations(operation_id);
CREATE INDEX IF NOT EXISTS idx_bulk_operations_user_id ON bulk_operations(user_id);
CREATE INDEX IF NOT EXISTS idx_bulk_operations_operation_type ON bulk_operations(operation_type);
CREATE INDEX IF NOT EXISTS idx_bulk_operations_status ON bulk_operations(status);
CREATE INDEX IF NOT EXISTS idx_bulk_operations_started_at ON bulk_operations(started_at);
CREATE INDEX IF NOT EXISTS idx_bulk_operations_completed_at ON bulk_operations(completed_at);

-- Bulk operation items indexes
CREATE INDEX IF NOT EXISTS idx_bulk_operation_items_bulk_operation_id ON bulk_operation_items(bulk_operation_id);
CREATE INDEX IF NOT EXISTS idx_bulk_operation_items_merchant_id ON bulk_operation_items(merchant_id);
CREATE INDEX IF NOT EXISTS idx_bulk_operation_items_status ON bulk_operation_items(status);
CREATE INDEX IF NOT EXISTS idx_bulk_operation_items_processed_at ON bulk_operation_items(processed_at);

-- Add triggers for updated_at columns
CREATE TRIGGER update_portfolio_types_updated_at 
    BEFORE UPDATE ON portfolio_types 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_risk_levels_updated_at 
    BEFORE UPDATE ON risk_levels 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_merchants_updated_at 
    BEFORE UPDATE ON merchants 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_merchant_sessions_updated_at 
    BEFORE UPDATE ON merchant_sessions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_compliance_records_updated_at 
    BEFORE UPDATE ON compliance_records 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_merchant_analytics_updated_at 
    BEFORE UPDATE ON merchant_analytics 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_merchant_comparisons_updated_at 
    BEFORE UPDATE ON merchant_comparisons 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_bulk_operations_updated_at 
    BEFORE UPDATE ON bulk_operations 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create views for common queries

-- View for merchant portfolio summary
CREATE OR REPLACE VIEW merchant_portfolio_summary AS
SELECT 
    pt.type as portfolio_type,
    rl.level as risk_level,
    COUNT(*) as merchant_count,
    AVG(ma.risk_score) as avg_risk_score,
    AVG(ma.compliance_score) as avg_compliance_score
FROM merchants m
JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
JOIN risk_levels rl ON m.risk_level_id = rl.id
LEFT JOIN merchant_analytics ma ON m.id = ma.merchant_id
WHERE m.status = 'active'
GROUP BY pt.type, rl.level;

-- View for active merchant sessions
CREATE OR REPLACE VIEW active_merchant_sessions AS
SELECT 
    ms.id,
    ms.user_id,
    ms.merchant_id,
    m.name as merchant_name,
    u.email as user_email,
    ms.started_at,
    ms.last_active,
    ms.is_active,
    CASE 
        WHEN ms.last_active < (CURRENT_TIMESTAMP - INTERVAL '24 hours') THEN true
        ELSE false
    END as is_expired
FROM merchant_sessions ms
JOIN merchants m ON ms.merchant_id = m.id
JOIN users u ON ms.user_id = u.id
WHERE ms.is_active = true;

-- View for merchant compliance status
CREATE OR REPLACE VIEW merchant_compliance_status AS
SELECT 
    m.id as merchant_id,
    m.name as merchant_name,
    m.compliance_status,
    cr.compliance_type,
    cr.status as check_status,
    cr.score as compliance_score,
    cr.checked_at,
    cr.expires_at,
    CASE 
        WHEN cr.expires_at IS NOT NULL AND cr.expires_at < CURRENT_TIMESTAMP THEN 'expired'
        WHEN cr.status = 'passed' THEN 'compliant'
        WHEN cr.status = 'failed' THEN 'non_compliant'
        ELSE 'pending'
    END as overall_status
FROM merchants m
LEFT JOIN compliance_records cr ON m.id = cr.merchant_id
WHERE m.status = 'active';

-- Add comments for documentation
COMMENT ON TABLE portfolio_types IS 'Lookup table for merchant portfolio types (onboarded, deactivated, prospective, pending)';
COMMENT ON TABLE risk_levels IS 'Lookup table for merchant risk levels (high, medium, low) with numeric values for comparison';
COMMENT ON TABLE merchants IS 'Main merchants table for portfolio management, enhanced version of businesses table';
COMMENT ON TABLE merchant_sessions IS 'Tracks active merchant sessions for single merchant session management';
COMMENT ON TABLE merchant_audit_logs IS 'Audit trail for all merchant operations and changes';
COMMENT ON TABLE compliance_records IS 'Compliance check records for merchants with expiration tracking';
COMMENT ON TABLE merchant_analytics IS 'Calculated analytics data for merchants including risk and compliance scores';
COMMENT ON TABLE merchant_notifications IS 'Notifications for merchants including risk alerts and compliance updates';
COMMENT ON TABLE merchant_comparisons IS 'Stores comparison data between two merchants';
COMMENT ON TABLE bulk_operations IS 'Tracks bulk operations on merchants with progress monitoring';
COMMENT ON TABLE bulk_operation_items IS 'Individual items within bulk operations for detailed tracking';

COMMENT ON VIEW merchant_portfolio_summary IS 'Summary view of merchant portfolio by type and risk level';
COMMENT ON VIEW active_merchant_sessions IS 'View of currently active merchant sessions with expiration status';
COMMENT ON VIEW merchant_compliance_status IS 'View of merchant compliance status with check details';

-- Enable trigram extension for text search if not already enabled
CREATE EXTENSION IF NOT EXISTS pg_trgm;
