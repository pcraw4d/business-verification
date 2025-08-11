-- KYB Platform - Database Initialization Script
-- This script initializes the database with basic setup and seed data

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Set timezone
SET timezone = 'UTC';

-- Create initial admin user (password: admin123)
INSERT INTO users (
    id, 
    email, 
    password_hash, 
    first_name, 
    last_name, 
    role, 
    is_active, 
    created_at, 
    updated_at
) VALUES (
    uuid_generate_v4(),
    'admin@kybplatform.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- bcrypt hash for 'admin123'
    'Admin',
    'User',
    'admin',
    true,
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Create initial test user (password: test123)
INSERT INTO users (
    id, 
    email, 
    password_hash, 
    first_name, 
    last_name, 
    role, 
    is_active, 
    created_at, 
    updated_at
) VALUES (
    uuid_generate_v4(),
    'test@kybplatform.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- bcrypt hash for 'test123'
    'Test',
    'User',
    'user',
    true,
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Create sample API key for testing
INSERT INTO api_keys (
    id,
    user_id,
    key_hash,
    name,
    permissions,
    expires_at,
    is_active,
    created_at,
    last_used
) VALUES (
    uuid_generate_v4(),
    (SELECT id FROM users WHERE email = 'test@kybplatform.com'),
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- bcrypt hash for 'test-api-key'
    'Test API Key',
    'read,write',
    NOW() + INTERVAL '1 year',
    true,
    NOW(),
    NULL
);

-- Create sample classification data
INSERT INTO classifications (
    id,
    business_name,
    primary_code,
    primary_description,
    confidence,
    method,
    created_at,
    user_id,
    metadata
) VALUES (
    uuid_generate_v4(),
    'Acme Corporation',
    '541511',
    'Custom Computer Programming Services',
    0.95,
    'hybrid',
    NOW(),
    (SELECT id FROM users WHERE email = 'test@kybplatform.com'),
    '{"industry": "technology", "size": "medium"}'
),
(
    uuid_generate_v4(),
    'Tech Solutions Inc',
    '541512',
    'Computer Systems Design Services',
    0.92,
    'hybrid',
    NOW(),
    (SELECT id FROM users WHERE email = 'test@kybplatform.com'),
    '{"industry": "technology", "size": "small"}'
);

-- Create sample risk assessment data
INSERT INTO risk_assessments (
    id,
    business_id,
    overall_score,
    factor_scores,
    risk_level,
    details,
    created_at,
    user_id
) VALUES (
    uuid_generate_v4(),
    'biz-123',
    0.25,
    '{"financial": 0.2, "operational": 0.3, "compliance": 0.2, "market": 0.3}',
    'low',
    '{"factors": ["stable_revenue", "good_credit"], "recommendations": ["monitor_quarterly"]}',
    NOW(),
    (SELECT id FROM users WHERE email = 'test@kybplatform.com')
);

-- Create sample compliance check data
INSERT INTO compliance_checks (
    id,
    business_id,
    framework,
    score,
    status,
    gaps,
    recommendations,
    created_at,
    user_id
) VALUES (
    uuid_generate_v4(),
    'biz-123',
    'SOC2',
    0.85,
    'compliant',
    '[]',
    '{"recommendations": ["implement_advanced_monitoring", "enhance_access_controls"]}',
    NOW(),
    (SELECT id FROM users WHERE email = 'test@kybplatform.com')
);

-- Create sample audit log entries
INSERT INTO audit_logs (
    id,
    user_id,
    action,
    resource_type,
    resource_id,
    details,
    ip_address,
    created_at
) VALUES (
    uuid_generate_v4(),
    (SELECT id FROM users WHERE email = 'admin@kybplatform.com'),
    'login',
    'user',
    (SELECT id FROM users WHERE email = 'admin@kybplatform.com'),
    '{"method": "password", "success": true}',
    '127.0.0.1',
    NOW()
),
(
    uuid_generate_v4(),
    (SELECT id FROM users WHERE email = 'test@kybplatform.com'),
    'classification_create',
    'classification',
    (SELECT id FROM classifications WHERE business_name = 'Acme Corporation'),
    '{"business_name": "Acme Corporation", "confidence": 0.95}',
    '127.0.0.1',
    NOW()
);

-- Grant necessary permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO kyb_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO kyb_user;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_classifications_business_name ON classifications(business_name);
CREATE INDEX IF NOT EXISTS idx_classifications_created_at ON classifications(created_at);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_business_id ON risk_assessments(business_id);
CREATE INDEX IF NOT EXISTS idx_compliance_checks_business_id ON compliance_checks(business_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- Update statistics
ANALYZE;
