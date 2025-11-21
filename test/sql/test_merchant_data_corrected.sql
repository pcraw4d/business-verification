-- Test data for merchant-details page testing
-- This script seeds the database with test merchants for various scenarios
-- Updated to match actual database schema (address and contact_info as JSONB)

-- Insert complete merchant with all data
INSERT INTO merchants (
    id, name, legal_name, registration_number, tax_id, industry, industry_code,
    business_type, founded_date, employee_count, annual_revenue,
    address, contact_info,
    portfolio_type, risk_level, compliance_status, status,
    created_by, created_at, updated_at
) VALUES (
    'merchant-complete-123',
    'Complete Test Company',
    'Complete Test Company Inc.',
    'REG-123456',
    'TAX-789012',
    'Technology',
    '541511',
    'Corporation',
    '2019-01-15',
    150,
    5000000.00,
    '{"street1": "123 Main Street", "street2": "Suite 100", "city": "San Francisco", "state": "CA", "postalCode": "94102", "country": "United States", "countryCode": "US"}'::jsonb,
    '{"phone": "+1-555-123-4567", "email": "contact@completetest.com", "website": "https://www.completetest.com", "primaryContact": "John Doe"}'::jsonb,
    'onboarded',
    'low',
    'compliant',
    'active',
    'test-user-1',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Insert partial merchant with missing data
INSERT INTO merchants (
    id, name, legal_name, industry,
    address, contact_info,
    portfolio_type, risk_level, status,
    created_by, created_at, updated_at
) VALUES (
    'merchant-partial-456',
    'Partial Test Company',
    'Partial Test Company LLC',
    'Retail',
    '{"street1": "456 Oak Avenue", "city": "Los Angeles", "state": "CA", "postalCode": "90001", "country": "United States", "countryCode": "US"}'::jsonb,
    '{"email": "info@partialtest.com"}'::jsonb,
    'prospective',
    'medium',
    'active',
    'test-user-1',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Insert minimal merchant
INSERT INTO merchants (
    id, name, legal_name, industry,
    address,
    portfolio_type, status,
    created_by, created_at, updated_at
) VALUES (
    'merchant-minimal-789',
    'Minimal Test Company',
    'Minimal Test Company',
    'Services',
    '{"city": "New York", "state": "NY", "country": "United States", "countryCode": "US"}'::jsonb,
    'pending',
    'pending',
    'test-user-1',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Insert merchant with risk assessment
INSERT INTO merchants (
    id, name, legal_name, industry,
    portfolio_type, risk_level, status,
    created_by, created_at, updated_at
) VALUES (
    'merchant-risk-001',
    'High Risk Test Company',
    'High Risk Test Company Inc.',
    'Financial Services',
    'onboarded',
    'high',
    'active',
    'test-user-1',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Insert merchant with analytics data
INSERT INTO merchants (
    id, name, legal_name, industry, industry_code,
    contact_info,
    portfolio_type, risk_level, status,
    created_by, created_at, updated_at
) VALUES (
    'merchant-analytics-001',
    'Analytics Test Company',
    'Analytics Test Company Inc.',
    'Technology',
    '5734',
    '{"website": "https://www.analyticstest.com"}'::jsonb,
    'onboarded',
    'low',
    'active',
    'test-user-1',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Insert test merchants for error scenarios
INSERT INTO merchants (
    id, name, legal_name, industry,
    portfolio_type, status,
    created_by, created_at, updated_at
) VALUES 
    ('merchant-404', 'Not Found Test', 'Not Found Test Inc.', 'Technology', 'onboarded', 'active', 'test-user-1', NOW(), NOW()),
    ('merchant-500', 'Server Error Test', 'Server Error Test Inc.', 'Technology', 'onboarded', 'active', 'test-user-1', NOW(), NOW()),
    ('merchant-no-risk', 'No Risk Assessment', 'No Risk Assessment Inc.', 'Technology', 'onboarded', 'active', 'test-user-1', NOW(), NOW()),
    ('merchant-no-analytics', 'No Analytics', 'No Analytics Inc.', 'Technology', 'onboarded', 'active', 'test-user-1', NOW(), NOW()),
    ('merchant-no-industry-code', 'No Industry Code', 'No Industry Code Inc.', 'Technology', 'onboarded', 'active', 'test-user-1', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

