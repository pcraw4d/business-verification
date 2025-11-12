-- Test data for merchant-details page testing
-- This script seeds the database with test merchants for various scenarios

-- Insert complete merchant with all data
INSERT INTO merchants (
    id, name, legal_name, registration_number, tax_id, industry, industry_code,
    business_type, founded_date, employee_count, annual_revenue,
    address_street1, address_street2, address_city, address_state,
    address_postal_code, address_country, address_country_code,
    contact_phone, contact_email, contact_website, contact_primary_contact,
    portfolio_type_id, risk_level_id, compliance_status, status,
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
    '123 Main Street',
    'Suite 100',
    'San Francisco',
    'CA',
    '94102',
    'United States',
    'US',
    '+1-555-123-4567',
    'contact@completetest.com',
    'https://www.completetest.com',
    'John Doe',
    (SELECT id FROM portfolio_types WHERE type = 'onboarded' LIMIT 1),
    (SELECT id FROM risk_levels WHERE level = 'low' LIMIT 1),
    'compliant',
    'active',
    'test-user-1',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Insert partial merchant with missing data
INSERT INTO merchants (
    id, name, legal_name, industry,
    address_street1, address_city, address_state,
    address_postal_code, address_country, address_country_code,
    contact_email,
    portfolio_type_id, risk_level_id, status,
    created_by, created_at, updated_at
) VALUES (
    'merchant-partial-456',
    'Partial Test Company',
    'Partial Test Company LLC',
    'Retail',
    '456 Oak Avenue',
    'Los Angeles',
    'CA',
    '90001',
    'United States',
    'US',
    'info@partialtest.com',
    (SELECT id FROM portfolio_types WHERE type = 'prospective' LIMIT 1),
    (SELECT id FROM risk_levels WHERE level = 'medium' LIMIT 1),
    'active',
    'test-user-1',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Insert minimal merchant
INSERT INTO merchants (
    id, name, legal_name, industry,
    address_city, address_state, address_country, address_country_code,
    portfolio_type_id, status,
    created_by, created_at, updated_at
) VALUES (
    'merchant-minimal-789',
    'Minimal Test Company',
    'Minimal Test Company',
    'Services',
    'New York',
    'NY',
    'United States',
    'US',
    (SELECT id FROM portfolio_types WHERE type = 'pending' LIMIT 1),
    'pending',
    'test-user-1',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Insert merchant with risk assessment
INSERT INTO merchants (
    id, name, legal_name, industry,
    portfolio_type_id, risk_level_id, status,
    created_by, created_at, updated_at
) VALUES (
    'merchant-risk-001',
    'High Risk Test Company',
    'High Risk Test Company Inc.',
    'Financial Services',
    (SELECT id FROM portfolio_types WHERE type = 'onboarded' LIMIT 1),
    (SELECT id FROM risk_levels WHERE level = 'high' LIMIT 1),
    'active',
    'test-user-1',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Insert merchant with analytics data
INSERT INTO merchants (
    id, name, legal_name, industry, industry_code,
    contact_website,
    portfolio_type_id, risk_level_id, status,
    created_by, created_at, updated_at
) VALUES (
    'merchant-analytics-001',
    'Analytics Test Company',
    'Analytics Test Company Inc.',
    'Technology',
    '5734',
    'https://www.analyticstest.com',
    (SELECT id FROM portfolio_types WHERE type = 'onboarded' LIMIT 1),
    (SELECT id FROM risk_levels WHERE level = 'low' LIMIT 1),
    'active',
    'test-user-1',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Insert classification data for analytics merchant
INSERT INTO business_classifications (
    business_name, primary_industry, confidence_score, classification_metadata
) VALUES (
    'Analytics Test Company',
    '{"name": "Technology", "code": "541511"}'::jsonb,
    0.95,
    '{
        "mcc_codes": [
            {"code": "5734", "description": "Computer Software Stores", "confidence": 0.92},
            {"code": "7372", "description": "Computer Programming Services", "confidence": 0.88}
        ],
        "sic_codes": [
            {"code": "7372", "description": "Prepackaged Software", "confidence": 0.90}
        ],
        "naics_codes": [
            {"code": "541511", "description": "Custom Computer Programming Services", "confidence": 0.93}
        ]
    }'::jsonb
) ON CONFLICT DO NOTHING;

-- Insert risk assessment records
INSERT INTO risk_assessments (
    id, merchant_id, status, options, progress, estimated_completion, created_at, updated_at
) VALUES (
    'assess-pending-001',
    'merchant-risk-001',
    'pending',
    '{"includeHistory": true, "includePredictions": true}'::jsonb,
    0,
    NOW() + INTERVAL '5 minutes',
    NOW(),
    NOW()
),
(
    'assess-processing-001',
    'merchant-risk-001',
    'processing',
    '{"includeHistory": true, "includePredictions": false}'::jsonb,
    65,
    NOW() + INTERVAL '2 minutes',
    NOW(),
    NOW()
),
(
    'assess-completed-001',
    'merchant-risk-001',
    'completed',
    '{"includeHistory": true, "includePredictions": true}'::jsonb,
    100,
    NULL,
    NOW(),
    NOW(),
    NOW() - INTERVAL '30 seconds'
) ON CONFLICT (id) DO NOTHING;

-- Update completed assessment with result
UPDATE risk_assessments
SET result = '{
    "overallScore": 0.75,
    "riskLevel": "medium",
    "factors": [
        {"name": "Financial Stability", "score": 0.8, "weight": 0.3},
        {"name": "Business History", "score": 0.7, "weight": 0.25},
        {"name": "Compliance", "score": 0.75, "weight": 0.25},
        {"name": "Industry Risk", "score": 0.7, "weight": 0.2}
    ]
}'::jsonb
WHERE id = 'assess-completed-001';

