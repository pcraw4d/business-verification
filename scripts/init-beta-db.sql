-- KYB Platform - Beta Database Initialization
-- This script sets up the beta environment with test data

-- Create beta-specific database
CREATE DATABASE IF NOT EXISTS kyb_beta;
CREATE DATABASE IF NOT EXISTS mattermost_beta;

-- Connect to beta database
\c kyb_beta;

-- Run migrations
\i /docker-entrypoint-initdb.d/001_initial_schema.sql
\i /docker-entrypoint-initdb.d/002_rbac_schema.sql
\i /docker-entrypoint-initdb.d/003_performance_indexes.sql
\i /docker-entrypoint-initdb.d/004_supabase_optimizations.sql

-- Create beta-specific tables for feedback collection
CREATE TABLE IF NOT EXISTS beta_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    feedback_type VARCHAR(50) NOT NULL, -- 'survey', 'bug_report', 'feature_request', 'general'
    category VARCHAR(100), -- 'onboarding', 'feature_usage', 'performance', 'usability'
    rating INTEGER CHECK (rating >= 1 AND rating <= 10),
    feedback_text TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS beta_user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    session_id VARCHAR(255) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,
    features_used TEXT[],
    pages_visited TEXT[],
    errors_encountered TEXT[],
    metadata JSONB
);

CREATE TABLE IF NOT EXISTS beta_surveys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_type VARCHAR(50) NOT NULL, -- 'onboarding', 'feature_usage', 'overall_experience', 'performance'
    user_id UUID REFERENCES users(id),
    responses JSONB NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    response_time_seconds INTEGER
);

-- Create indexes for beta analytics
CREATE INDEX IF NOT EXISTS idx_beta_feedback_user_id ON beta_feedback(user_id);
CREATE INDEX IF NOT EXISTS idx_beta_feedback_type ON beta_feedback(feedback_type);
CREATE INDEX IF NOT EXISTS idx_beta_feedback_created_at ON beta_feedback(created_at);
CREATE INDEX IF NOT EXISTS idx_beta_user_sessions_user_id ON beta_user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_beta_user_sessions_start_time ON beta_user_sessions(start_time);
CREATE INDEX IF NOT EXISTS idx_beta_surveys_user_id ON beta_surveys(user_id);
CREATE INDEX IF NOT EXISTS idx_beta_surveys_type ON beta_surveys(survey_type);

-- Insert beta test data for realistic testing scenarios

-- Beta test businesses for different industries
INSERT INTO businesses (id, business_name, business_type, industry, naics_code, confidence_score, created_at, updated_at) VALUES
-- Financial Institutions
('550e8400-e29b-41d4-a716-446655440001', 'Beta Bank Corporation', 'Corporation', 'Financial Services', '522110', 0.95, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440002', 'Test Credit Union', 'Credit Union', 'Financial Services', '522130', 0.92, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440003', 'Beta Fintech Solutions', 'Limited Liability Company', 'Financial Services', '522320', 0.88, NOW(), NOW()),

-- Technology Companies
('550e8400-e29b-41d4-a716-446655440004', 'Test Tech Startup', 'Corporation', 'Technology', '511210', 0.90, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440005', 'Beta Software Solutions', 'Corporation', 'Technology', '541511', 0.93, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440006', 'Test Digital Agency', 'Limited Liability Company', 'Technology', '541810', 0.87, NOW(), NOW()),

-- Legal and Compliance
('550e8400-e29b-41d4-a716-446655440007', 'Beta Legal Associates', 'Partnership', 'Legal Services', '541110', 0.94, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440008', 'Test Compliance Consulting', 'Corporation', 'Professional Services', '541612', 0.89, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440009', 'Beta Risk Management', 'Limited Liability Company', 'Professional Services', '524292', 0.91, NOW(), NOW()),

-- Healthcare
('550e8400-e29b-41d4-a716-446655440010', 'Test Medical Center', 'Corporation', 'Healthcare', '622110', 0.96, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440011', 'Beta Healthcare Solutions', 'Corporation', 'Healthcare', '621111', 0.92, NOW(), NOW()),

-- Manufacturing
('550e8400-e29b-41d4-a716-446655440012', 'Test Manufacturing Co', 'Corporation', 'Manufacturing', '332996', 0.88, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440013', 'Beta Industrial Products', 'Corporation', 'Manufacturing', '332312', 0.90, NOW(), NOW()),

-- Retail and E-commerce
('550e8400-e29b-41d4-a716-446655440014', 'Test Retail Store', 'Corporation', 'Retail', '452111', 0.85, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440015', 'Beta E-commerce Platform', 'Corporation', 'Retail', '454110', 0.87, NOW(), NOW()),

-- Consulting
('550e8400-e29b-41d4-a716-446655440016', 'Test Management Consulting', 'Corporation', 'Professional Services', '541611', 0.89, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440017', 'Beta Strategy Group', 'Limited Liability Company', 'Professional Services', '541618', 0.86, NOW(), NOW()),

-- Real Estate
('550e8400-e29b-41d4-a716-446655440018', 'Test Real Estate Agency', 'Corporation', 'Real Estate', '531210', 0.84, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440019', 'Beta Property Management', 'Limited Liability Company', 'Real Estate', '531312', 0.88, NOW(), NOW()),

-- Energy
('550e8400-e29b-41d4-a716-446655440020', 'Test Energy Solutions', 'Corporation', 'Energy', '221111', 0.91, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Insert beta test users for different user types
INSERT INTO users (id, email, username, first_name, last_name, company, role, status, email_verified, created_at, updated_at) VALUES
-- Financial Institution Users
('660e8400-e29b-41d4-a716-446655440001', 'banker@betabank.com', 'banker1', 'John', 'Smith', 'Beta Bank Corporation', 'user', 'active', true, NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440002', 'compliance@testcredit.com', 'compliance1', 'Sarah', 'Johnson', 'Test Credit Union', 'user', 'active', true, NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440003', 'risk@betafintech.com', 'risk1', 'Michael', 'Brown', 'Beta Fintech Solutions', 'user', 'active', true, NOW(), NOW()),

-- Technology Company Users
('660e8400-e29b-41d4-a716-446655440004', 'cto@testtech.com', 'cto1', 'Emily', 'Davis', 'Test Tech Startup', 'user', 'active', true, NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440005', 'dev@betasoftware.com', 'dev1', 'David', 'Wilson', 'Beta Software Solutions', 'user', 'active', true, NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440006', 'pm@testdigital.com', 'pm1', 'Lisa', 'Anderson', 'Test Digital Agency', 'user', 'active', true, NOW(), NOW()),

-- Legal and Compliance Users
('660e8400-e29b-41d4-a716-446655440007', 'partner@betalegal.com', 'partner1', 'Robert', 'Taylor', 'Beta Legal Associates', 'user', 'active', true, NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440008', 'consultant@testcompliance.com', 'consultant1', 'Jennifer', 'Martinez', 'Test Compliance Consulting', 'user', 'active', true, NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440009', 'manager@betarisk.com', 'manager1', 'Christopher', 'Garcia', 'Beta Risk Management', 'user', 'active', true, NOW(), NOW()),

-- Healthcare Users
('660e8400-e29b-41d4-a716-446655440010', 'admin@testmedical.com', 'admin1', 'Amanda', 'Rodriguez', 'Test Medical Center', 'user', 'active', true, NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440011', 'director@betahealthcare.com', 'director1', 'James', 'Lee', 'Beta Healthcare Solutions', 'user', 'active', true, NOW(), NOW()),

-- Manufacturing Users
('660e8400-e29b-41d4-a716-446655440012', 'operations@testmanufacturing.com', 'ops1', 'Patricia', 'White', 'Test Manufacturing Co', 'user', 'active', true, NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440013', 'ceo@betaindustrial.com', 'ceo1', 'Thomas', 'Harris', 'Beta Industrial Products', 'user', 'active', true, NOW(), NOW()),

-- Retail Users
('660e8400-e29b-41d4-a716-446655440014', 'manager@testretail.com', 'retail1', 'Nancy', 'Clark', 'Test Retail Store', 'user', 'active', true, NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440015', 'founder@betaecommerce.com', 'founder1', 'Kevin', 'Lewis', 'Beta E-commerce Platform', 'user', 'active', true, NOW(), NOW()),

-- Consulting Users
('660e8400-e29b-41d4-a716-446655440016', 'partner@testconsulting.com', 'partner2', 'Michelle', 'Robinson', 'Test Management Consulting', 'user', 'active', true, NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440017', 'analyst@betastrategy.com', 'analyst1', 'Daniel', 'Walker', 'Beta Strategy Group', 'user', 'active', true, NOW(), NOW()),

-- Real Estate Users
('660e8400-e29b-41d4-a716-446655440018', 'agent@testrealestate.com', 'agent1', 'Stephanie', 'Perez', 'Test Real Estate Agency', 'user', 'active', true, NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440019', 'manager@betaproperty.com', 'prop1', 'Ryan', 'Hall', 'Beta Property Management', 'user', 'active', true, NOW(), NOW()),

-- Energy Users
('660e8400-e29b-41d4-a716-446655440020', 'engineer@testenergy.com', 'engineer1', 'Laura', 'Young', 'Test Energy Solutions', 'user', 'active', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Insert sample risk assessments for beta testing
INSERT INTO risk_assessments (id, business_id, overall_score, overall_level, category_scores, factor_scores, recommendations, alerts, assessed_at, valid_until) VALUES
('770e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 0.25, 'low', 
 '{"financial": {"score": 0.20, "level": "low"}, "operational": {"score": 0.30, "level": "low"}, "regulatory": {"score": 0.15, "level": "low"}, "reputational": {"score": 0.25, "level": "low"}, "cybersecurity": {"score": 0.35, "level": "medium"}}',
 '{"profit_margin": {"score": 0.20, "level": "low"}, "credit_score": {"score": 0.15, "level": "low"}, "compliance_score": {"score": 0.10, "level": "low"}}',
 '["Maintain current risk management practices", "Consider cybersecurity enhancements"]',
 '[]',
 NOW(), NOW() + INTERVAL '30 days'),

('770e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440004', 0.75, 'high',
 '{"financial": {"score": 0.80, "level": "high"}, "operational": {"score": 0.70, "level": "high"}, "regulatory": {"score": 0.60, "level": "medium"}, "reputational": {"score": 0.75, "level": "high"}, "cybersecurity": {"score": 0.85, "level": "critical"}}',
 '{"profit_margin": {"score": 0.90, "level": "critical"}, "credit_score": {"score": 0.70, "level": "high"}, "compliance_score": {"score": 0.65, "level": "medium"}}',
 '["Implement comprehensive risk mitigation strategy", "Enhance cybersecurity measures", "Improve financial controls"]',
 '["High financial risk detected", "Critical cybersecurity vulnerabilities"]',
 NOW(), NOW() + INTERVAL '30 days'),

('770e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440007', 0.45, 'medium',
 '{"financial": {"score": 0.40, "level": "medium"}, "operational": {"score": 0.50, "level": "medium"}, "regulatory": {"score": 0.35, "level": "medium"}, "reputational": {"score": 0.45, "level": "medium"}, "cybersecurity": {"score": 0.55, "level": "medium"}}',
 '{"profit_margin": {"score": 0.45, "level": "medium"}, "credit_score": {"score": 0.40, "level": "medium"}, "compliance_score": {"score": 0.35, "level": "medium"}}',
 '["Monitor risk factors closely", "Consider operational improvements"]',
 '["Medium operational risk detected"]',
 NOW(), NOW() + INTERVAL '30 days')
ON CONFLICT (id) DO NOTHING;

-- Insert sample compliance tracking for beta testing
INSERT INTO compliance_status (id, business_id, framework, status, score, requirements, last_updated) VALUES
('880e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'SOC2', 'compliant', 0.85, 
 '{"Access Control": "implemented", "Data Encryption": "implemented", "Audit Logging": "implemented", "Incident Response": "implemented"}',
 NOW()),

('880e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'PCI-DSS', 'compliant', 0.90,
 '{"Build and Maintain a Secure Network": "implemented", "Protect Cardholder Data": "implemented", "Maintain Vulnerability Management": "implemented", "Implement Strong Access Controls": "implemented"}',
 NOW()),

('880e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440004', 'SOC2', 'non_compliant', 0.35,
 '{"Access Control": "partial", "Data Encryption": "implemented", "Audit Logging": "missing", "Incident Response": "missing"}',
 NOW()),

('880e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440007', 'GDPR', 'compliant', 0.88,
 '{"Data Minimization": "implemented", "Consent Management": "implemented", "Data Subject Rights": "implemented", "Data Retention": "implemented"}',
 NOW())
ON CONFLICT (id) DO NOTHING;

-- Insert sample feedback data for beta testing
INSERT INTO beta_feedback (user_id, feedback_type, category, rating, feedback_text, metadata) VALUES
('660e8400-e29b-41d4-a716-446655440001', 'survey', 'onboarding', 9, 'The onboarding process was very smooth and intuitive. The training materials were comprehensive.', '{"survey_id": "onboarding_001", "completion_time": 1200}'),
('660e8400-e29b-41d4-a716-446655440002', 'survey', 'feature_usage', 8, 'The risk assessment feature is excellent. Very accurate and provides valuable insights.', '{"survey_id": "feature_001", "features_used": ["risk_assessment", "compliance_check"]}'),
('660e8400-e29b-41d4-a716-446655440003', 'bug_report', 'performance', 6, 'Sometimes the platform is slow when processing large datasets.', '{"bug_id": "PERF-001", "severity": "medium", "browser": "chrome"}'),
('660e8400-e29b-41d4-a716-446655440004', 'feature_request', 'usability', 7, 'Would love to see bulk upload functionality for business data.', '{"feature_id": "BULK-001", "priority": "high", "impact": "high"}'),
('660e8400-e29b-41d4-a716-446655440005', 'survey', 'overall_experience', 9, 'Overall, this platform has significantly improved our due diligence process.', '{"survey_id": "overall_001", "nps_score": 9, "willingness_to_pay": true}')
ON CONFLICT DO NOTHING;

-- Insert sample user sessions for analytics
INSERT INTO beta_user_sessions (user_id, session_id, start_time, end_time, duration_seconds, features_used, pages_visited, errors_encountered) VALUES
('660e8400-e29b-41d4-a716-446655440001', 'session_001', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '1 hour 45 minutes', 900, 
 '{"business_classification", "risk_assessment", "dashboard"}', 
 '{"dashboard", "classify", "risk", "compliance"}', 
 '{}'),

('660e8400-e29b-41d4-a716-446655440002', 'session_002', NOW() - INTERVAL '4 hours', NOW() - INTERVAL '3 hours 30 minutes', 1800,
 '{"compliance_check", "reporting", "analytics"}',
 '{"compliance", "reports", "analytics", "settings"}',
 '{"timeout_error"}'),

('660e8400-e29b-41d4-a716-446655440003', 'session_003', NOW() - INTERVAL '6 hours', NOW() - INTERVAL '5 hours 15 minutes', 2700,
 '{"risk_assessment", "compliance_check", "business_classification"}',
 '{"risk", "compliance", "classify", "dashboard"}',
 '{}')
ON CONFLICT DO NOTHING;

-- Insert sample survey responses
INSERT INTO beta_surveys (survey_type, user_id, responses, response_time_seconds) VALUES
('onboarding', '660e8400-e29b-41d4-a716-446655440001', 
 '{"ease_of_onboarding": 9, "training_helpful": true, "confidence_level": 8, "challenges": "None", "improvements": "More video tutorials"}', 
 180),

('feature_usage', '660e8400-e29b-41d4-a716-446655440002',
 '{"features_used": ["risk_assessment", "compliance_check"], "satisfaction_scores": {"risk_assessment": 9, "compliance_check": 8}, "missing_features": "Bulk processing", "comparison": "Better than alternatives", "value_improvements": "API integration"}',
 240),

('overall_experience', '660e8400-e29b-41d4-a716-446655440003',
 '{"nps_score": 9, "overall_satisfaction": 9, "recommendation_likelihood": 9, "value_for_money": 8, "willingness_to_pay": true, "changes_requested": "Mobile app", "benefits": "Time savings, accuracy improvement"}',
 300)
ON CONFLICT DO NOTHING;

-- Create beta-specific views for analytics
CREATE OR REPLACE VIEW beta_user_engagement AS
SELECT 
    u.id as user_id,
    u.email,
    u.company,
    COUNT(DISTINCT s.session_id) as total_sessions,
    AVG(s.duration_seconds) as avg_session_duration,
    COUNT(DISTINCT f.id) as total_feedback,
    AVG(f.rating) as avg_rating,
    COUNT(DISTINCT sur.id) as surveys_completed
FROM users u
LEFT JOIN beta_user_sessions s ON u.id = s.user_id
LEFT JOIN beta_feedback f ON u.id = f.user_id
LEFT JOIN beta_surveys sur ON u.id = sur.user_id
WHERE u.email LIKE '%@beta%' OR u.email LIKE '%@test%'
GROUP BY u.id, u.email, u.company;

CREATE OR REPLACE VIEW beta_feature_usage AS
SELECT 
    feature,
    COUNT(*) as usage_count,
    COUNT(DISTINCT user_id) as unique_users
FROM (
    SELECT user_id, unnest(features_used) as feature
    FROM beta_user_sessions
    WHERE features_used IS NOT NULL
) feature_usage
GROUP BY feature
ORDER BY usage_count DESC;

CREATE OR REPLACE VIEW beta_feedback_summary AS
SELECT 
    feedback_type,
    category,
    COUNT(*) as total_feedback,
    AVG(rating) as avg_rating,
    COUNT(CASE WHEN rating >= 8 THEN 1 END) as positive_feedback,
    COUNT(CASE WHEN rating <= 4 THEN 1 END) as negative_feedback
FROM beta_feedback
WHERE rating IS NOT NULL
GROUP BY feedback_type, category
ORDER BY feedback_type, avg_rating DESC;

-- Grant necessary permissions
GRANT ALL PRIVILEGES ON DATABASE kyb_beta TO kyb_beta_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO kyb_beta_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO kyb_beta_user;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO kyb_beta_user;

-- Create beta-specific roles and permissions
INSERT INTO roles (id, name, description, permissions, created_at, updated_at) VALUES
('990e8400-e29b-41d4-a716-446655440001', 'beta_user', 'Beta testing user with full access', 
 '["read:businesses", "write:businesses", "read:risk_assessments", "write:risk_assessments", "read:compliance", "write:compliance", "read:feedback", "write:feedback"]',
 NOW(), NOW()),
('990e8400-e29b-41d4-a716-446655440002', 'beta_admin', 'Beta testing administrator', 
 '["read:all", "write:all", "admin:users", "admin:feedback", "admin:analytics"]',
 NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Assign beta roles to users
INSERT INTO role_assignments (id, user_id, role_id, assigned_at, expires_at) VALUES
('aa0e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440001', '990e8400-e29b-41d4-a716-446655440001', NOW(), NOW() + INTERVAL '90 days'),
('aa0e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440002', '990e8400-e29b-41d4-a716-446655440001', NOW(), NOW() + INTERVAL '90 days'),
('aa0e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440003', '990e8400-e29b-41d4-a716-446655440001', NOW(), NOW() + INTERVAL '90 days')
ON CONFLICT (id) DO NOTHING;

-- Log beta environment setup
INSERT INTO audit_logs (id, user_id, action, resource_type, resource_id, details, ip_address, user_agent, created_at) VALUES
('bb0e8400-e29b-41d4-a716-446655440001', NULL, 'beta_environment_setup', 'system', 'beta_init', 
 '{"action": "beta_database_initialization", "tables_created": 4, "test_data_inserted": 60, "views_created": 3}', 
 '127.0.0.1', 'beta-setup-script', NOW());

-- Print completion message
SELECT 'Beta database initialization completed successfully!' as status;
