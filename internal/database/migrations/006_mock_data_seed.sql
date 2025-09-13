-- Migration: 006_mock_data_seed.sql
-- Description: Seed database with mock merchant data, portfolio types, and risk levels
-- Created: 2025-01-19
-- Dependencies: 005_merchant_portfolio_schema.sql

-- Insert portfolio types
INSERT INTO portfolio_types (type, description, display_order, is_active) VALUES
('onboarded', 'Fully onboarded and active merchants', 1, true),
('prospective', 'Potential merchants under evaluation', 2, true),
('pending', 'Merchants awaiting approval or processing', 3, true),
('deactivated', 'Deactivated or suspended merchants', 4, true)
ON CONFLICT (type) DO UPDATE SET
    description = EXCLUDED.description,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- Insert risk levels
INSERT INTO risk_levels (level, description, numeric_value, color_code, display_order, is_active) VALUES
('low', 'Low risk merchants with established compliance history', 1, '#10B981', 1, true),
('medium', 'Medium risk merchants requiring standard monitoring', 2, '#F59E0B', 2, true),
('high', 'High risk merchants requiring enhanced due diligence', 3, '#EF4444', 3, true)
ON CONFLICT (level) DO UPDATE SET
    description = EXCLUDED.description,
    numeric_value = EXCLUDED.numeric_value,
    color_code = EXCLUDED.color_code,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- Create a default system user for created_by references if none exists
INSERT INTO users (id, email, username, password_hash, first_name, last_name, role, status, email_verified)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'system@kyb-platform.com',
    'system',
    '$2a$10$dummy.hash.for.system.user',
    'System',
    'User',
    'system',
    'active',
    true
)
ON CONFLICT (id) DO NOTHING;

-- Insert mock merchants with realistic data
INSERT INTO merchants (
    id, name, legal_name, registration_number, tax_id, industry, industry_code, business_type,
    founded_date, employee_count, annual_revenue,
    address_street1, address_street2, address_city, address_state, address_postal_code, address_country, address_country_code,
    contact_phone, contact_email, contact_website, contact_primary_contact,
    portfolio_type_id, risk_level_id, compliance_status, status, created_by
) VALUES
-- Technology Companies (Onboarded - Low Risk)
('10000000-0000-0000-0000-000000000001', 'TechFlow Solutions', 'TechFlow Solutions Inc.', 'TF-2023-001', '12-3456789', 'Technology', '541511', 'Corporation',
'2020-03-15', 45, 2500000.00,
'123 Innovation Drive', NULL, 'San Francisco', 'CA', '94105', 'United States', 'US',
'+1-415-555-0101', 'info@techflow.com', 'https://techflow.com', 'Sarah Johnson',
(SELECT id FROM portfolio_types WHERE type = 'onboarded'), (SELECT id FROM risk_levels WHERE level = 'low'), 'compliant', 'active', '00000000-0000-0000-0000-000000000001'),

('10000000-0000-0000-0000-000000000002', 'DataSync Analytics', 'DataSync Analytics LLC', 'DS-2022-002', '98-7654321', 'Technology', '541512', 'LLC',
'2019-08-22', 28, 1800000.00,
'456 Data Street', 'Suite 200', 'Austin', 'TX', '78701', 'United States', 'US',
'+1-512-555-0102', 'contact@datasync.com', 'https://datasync.com', 'Michael Chen',
(SELECT id FROM portfolio_types WHERE type = 'onboarded'), (SELECT id FROM risk_levels WHERE level = 'low'), 'compliant', 'active', '00000000-0000-0000-0000-000000000001'),

('10000000-0000-0000-0000-000000000003', 'CloudScale Systems', 'CloudScale Systems Inc.', 'CS-2021-003', '45-6789012', 'Technology', '541511', 'Corporation',
'2018-11-10', 67, 4200000.00,
'789 Cloud Avenue', NULL, 'Seattle', 'WA', '98101', 'United States', 'US',
'+1-206-555-0103', 'hello@cloudscale.com', 'https://cloudscale.com', 'David Rodriguez',
(SELECT id FROM portfolio_types WHERE type = 'onboarded'), (SELECT id FROM risk_levels WHERE level = 'low'), 'compliant', 'active', '00000000-0000-0000-0000-000000000001'),

-- Financial Services (Onboarded - Medium Risk)
('10000000-0000-0000-0000-000000000004', 'Metro Credit Union', 'Metro Credit Union', 'MCU-2020-004', '34-5678901', 'Finance', '522110', 'Credit Union',
'2015-06-30', 125, 8500000.00,
'321 Financial Plaza', 'Floor 15', 'Chicago', 'IL', '60601', 'United States', 'US',
'+1-312-555-0104', 'info@metrocu.org', 'https://metrocu.org', 'Jennifer Williams',
(SELECT id FROM portfolio_types WHERE type = 'onboarded'), (SELECT id FROM risk_levels WHERE level = 'medium'), 'compliant', 'active', '00000000-0000-0000-0000-000000000001'),

('10000000-0000-0000-0000-000000000005', 'Premier Investment Group', 'Premier Investment Group LLC', 'PIG-2019-005', '56-7890123', 'Finance', '523920', 'LLC',
'2017-04-12', 89, 12000000.00,
'654 Investment Way', 'Suite 500', 'New York', 'NY', '10001', 'United States', 'US',
'+1-212-555-0105', 'contact@premierinvest.com', 'https://premierinvest.com', 'Robert Thompson',
(SELECT id FROM portfolio_types WHERE type = 'onboarded'), (SELECT id FROM risk_levels WHERE level = 'medium'), 'compliant', 'active', '00000000-0000-0000-0000-000000000001'),

-- Healthcare (Prospective - Medium Risk)
('10000000-0000-0000-0000-000000000006', 'Wellness Medical Center', 'Wellness Medical Center PLLC', 'WMC-2023-006', '78-9012345', 'Healthcare', '621111', 'PLLC',
'2021-09-05', 156, 9800000.00,
'987 Health Boulevard', NULL, 'Denver', 'CO', '80201', 'United States', 'US',
'+1-303-555-0106', 'info@wellnessmed.com', 'https://wellnessmed.com', 'Dr. Lisa Anderson',
(SELECT id FROM portfolio_types WHERE type = 'prospective'), (SELECT id FROM risk_levels WHERE level = 'medium'), 'pending', 'active', '00000000-0000-0000-0000-000000000001'),

('10000000-0000-0000-0000-000000000007', 'Advanced Dental Care', 'Advanced Dental Care PC', 'ADC-2022-007', '90-1234567', 'Healthcare', '621210', 'PC',
'2020-12-18', 34, 2100000.00,
'147 Dental Drive', 'Building A', 'Phoenix', 'AZ', '85001', 'United States', 'US',
'+1-602-555-0107', 'appointments@advanceddental.com', 'https://advanceddental.com', 'Dr. Mark Davis',
(SELECT id FROM portfolio_types WHERE type = 'prospective'), (SELECT id FROM risk_levels WHERE level = 'medium'), 'pending', 'active', '00000000-0000-0000-0000-000000000001'),

-- Retail (Pending - Low Risk)
('10000000-0000-0000-0000-000000000008', 'Urban Fashion Co.', 'Urban Fashion Company Inc.', 'UFC-2023-008', '12-3456780', 'Retail', '448140', 'Corporation',
'2022-01-20', 78, 5600000.00,
'258 Fashion Street', NULL, 'Los Angeles', 'CA', '90001', 'United States', 'US',
'+1-213-555-0108', 'orders@urbanfashion.com', 'https://urbanfashion.com', 'Amanda Garcia',
(SELECT id FROM portfolio_types WHERE type = 'pending'), (SELECT id FROM risk_levels WHERE level = 'low'), 'pending', 'active', '00000000-0000-0000-0000-000000000001'),

('10000000-0000-0000-0000-000000000009', 'Green Earth Organics', 'Green Earth Organics LLC', 'GEO-2023-009', '23-4567890', 'Retail', '445299', 'LLC',
'2021-07-14', 42, 3200000.00,
'369 Organic Lane', NULL, 'Portland', 'OR', '97201', 'United States', 'US',
'+1-503-555-0109', 'info@greenearth.com', 'https://greenearth.com', 'Elena Martinez',
(SELECT id FROM portfolio_types WHERE type = 'pending'), (SELECT id FROM risk_levels WHERE level = 'low'), 'pending', 'active', '00000000-0000-0000-0000-000000000001'),

-- Manufacturing (High Risk - Deactivated)
('10000000-0000-0000-0000-000000000010', 'Precision Manufacturing', 'Precision Manufacturing Corp', 'PMC-2018-010', '34-5678900', 'Manufacturing', '332710', 'Corporation',
'2016-03-08', 234, 18500000.00,
'741 Industrial Park', 'Building 3', 'Detroit', 'MI', '48201', 'United States', 'US',
'+1-313-555-0110', 'info@precisionmfg.com', 'https://precisionmfg.com', 'James Wilson',
(SELECT id FROM portfolio_types WHERE type = 'deactivated'), (SELECT id FROM risk_levels WHERE level = 'high'), 'non_compliant', 'inactive', '00000000-0000-0000-0000-000000000001'),

-- Additional diverse businesses for comprehensive testing
('10000000-0000-0000-0000-000000000011', 'Creative Design Studio', 'Creative Design Studio LLC', 'CDS-2023-011', '45-6789010', 'Professional Services', '541810', 'LLC',
'2022-05-30', 15, 850000.00,
'123 Creative Lane', 'Studio 2B', 'Miami', 'FL', '33101', 'United States', 'US',
'+1-305-555-0111', 'hello@creativedesign.com', 'https://creativedesign.com', 'Sophia Lee',
(SELECT id FROM portfolio_types WHERE type = 'onboarded'), (SELECT id FROM risk_levels WHERE level = 'low'), 'compliant', 'active', '00000000-0000-0000-0000-000000000001'),

('10000000-0000-0000-0000-000000000012', 'Global Logistics Inc', 'Global Logistics Inc', 'GLI-2019-012', '56-7890120', 'Transportation', '484122', 'Corporation',
'2018-11-25', 189, 12500000.00,
'456 Logistics Way', NULL, 'Atlanta', 'GA', '30301', 'United States', 'US',
'+1-404-555-0112', 'info@globallogistics.com', 'https://globallogistics.com', 'Carlos Rodriguez',
(SELECT id FROM portfolio_types WHERE type = 'onboarded'), (SELECT id FROM risk_levels WHERE level = 'medium'), 'compliant', 'active', '00000000-0000-0000-0000-000000000001'),

('10000000-0000-0000-0000-000000000013', 'EcoTech Solutions', 'EcoTech Solutions Inc', 'ETS-2023-013', '67-8901230', 'Environmental Services', '562211', 'Corporation',
'2023-02-14', 32, 1800000.00,
'789 Green Street', 'Suite 100', 'San Diego', 'CA', '92101', 'United States', 'US',
'+1-619-555-0113', 'contact@ecotech.com', 'https://ecotech.com', 'Rachel Green',
(SELECT id FROM portfolio_types WHERE type = 'prospective'), (SELECT id FROM risk_levels WHERE level = 'low'), 'pending', 'active', '00000000-0000-0000-0000-000000000001'),

('10000000-0000-0000-0000-000000000014', 'Metro Construction', 'Metro Construction LLC', 'MCL-2020-014', '78-9012340', 'Construction', '236220', 'LLC',
'2019-08-10', 67, 4500000.00,
'321 Construction Blvd', NULL, 'Houston', 'TX', '77001', 'United States', 'US',
'+1-713-555-0114', 'info@metroconstruction.com', 'https://metroconstruction.com', 'Thomas Brown',
(SELECT id FROM portfolio_types WHERE type = 'pending'), (SELECT id FROM risk_levels WHERE level = 'medium'), 'pending', 'active', '00000000-0000-0000-0000-000000000001'),

('10000000-0000-0000-0000-000000000015', 'Digital Marketing Pro', 'Digital Marketing Pro LLC', 'DMP-2022-015', '89-0123450', 'Marketing', '541810', 'LLC',
'2021-12-01', 24, 1200000.00,
'654 Marketing Avenue', 'Floor 5', 'Boston', 'MA', '02101', 'United States', 'US',
'+1-617-555-0115', 'hello@digitalmarketingpro.com', 'https://digitalmarketingpro.com', 'Jennifer Taylor',
(SELECT id FROM portfolio_types WHERE type = 'onboarded'), (SELECT id FROM risk_levels WHERE level = 'low'), 'compliant', 'active', '00000000-0000-0000-0000-000000000001'),

-- High-risk businesses for testing
('10000000-0000-0000-0000-000000000016', 'Offshore Trading Co', 'Offshore Trading Company Ltd', 'OTC-2021-016', '90-1234560', 'Finance', '523110', 'Corporation',
'2020-06-15', 45, 8500000.00,
'789 Offshore Plaza', 'Suite 2000', 'Miami', 'FL', '33131', 'United States', 'US',
'+1-305-555-0116', 'info@offshoretrading.com', 'https://offshoretrading.com', 'Alexander Volkov',
(SELECT id FROM portfolio_types WHERE type = 'deactivated'), (SELECT id FROM risk_levels WHERE level = 'high'), 'non_compliant', 'inactive', '00000000-0000-0000-0000-000000000001'),

('10000000-0000-0000-0000-000000000017', 'Cryptocurrency Exchange', 'CryptoExchange LLC', 'CE-2023-017', '01-2345670', 'Finance', '523130', 'LLC',
'2023-01-10', 89, 15000000.00,
'456 Crypto Street', 'Floor 20', 'San Francisco', 'CA', '94105', 'United States', 'US',
'+1-415-555-0117', 'support@cryptoexchange.com', 'https://cryptoexchange.com', 'Michael Chen',
(SELECT id FROM portfolio_types WHERE type = 'prospective'), (SELECT id FROM risk_levels WHERE level = 'high'), 'pending', 'active', '00000000-0000-0000-0000-000000000001'),

-- International businesses
('10000000-0000-0000-0000-000000000018', 'Canadian Import Export', 'Canadian Import Export Ltd', 'CIE-2022-018', '12-3456789', 'Wholesale Trade', '423820', 'Corporation',
'2021-09-20', 56, 6800000.00,
'123 International Way', NULL, 'Toronto', 'ON', 'M5H 2N2', 'Canada', 'CA',
'+1-416-555-0118', 'info@canadianimport.com', 'https://canadianimport.com', 'Jean-Pierre Dubois',
(SELECT id FROM portfolio_types WHERE type = 'onboarded'), (SELECT id FROM risk_levels WHERE level = 'medium'), 'compliant', 'active', '00000000-0000-0000-0000-000000000001'),

('10000000-0000-0000-0000-000000000019', 'European Consulting Group', 'European Consulting Group GmbH', 'ECG-2020-019', 'DE123456789', 'Professional Services', '541611', 'Corporation',
'2019-04-15', 34, 2200000.00,
'456 Business District', 'Floor 8', 'Berlin', 'Berlin', '10115', 'Germany', 'DE',
'+49-30-555-0119', 'info@europeanconsulting.de', 'https://europeanconsulting.de', 'Klaus Mueller',
(SELECT id FROM portfolio_types WHERE type = 'prospective'), (SELECT id FROM risk_levels WHERE level = 'medium'), 'pending', 'active', '00000000-0000-0000-0000-000000000001'),

('10000000-0000-0000-0000-000000000020', 'Asia Pacific Trading', 'Asia Pacific Trading Pte Ltd', 'APT-2023-020', 'SG123456789', 'Wholesale Trade', '423820', 'Corporation',
'2022-11-30', 78, 9500000.00,
'789 Marina Bay', 'Tower 1, Level 15', 'Singapore', 'Singapore', '018956', 'Singapore', 'SG',
'+65-6123-0120', 'contact@asiapacific.sg', 'https://asiapacific.sg', 'Li Wei',
(SELECT id FROM portfolio_types WHERE type = 'pending'), (SELECT id FROM risk_levels WHERE level = 'medium'), 'pending', 'active', '00000000-0000-0000-0000-000000000001')

ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    legal_name = EXCLUDED.legal_name,
    registration_number = EXCLUDED.registration_number,
    tax_id = EXCLUDED.tax_id,
    industry = EXCLUDED.industry,
    industry_code = EXCLUDED.industry_code,
    business_type = EXCLUDED.business_type,
    founded_date = EXCLUDED.founded_date,
    employee_count = EXCLUDED.employee_count,
    annual_revenue = EXCLUDED.annual_revenue,
    address_street1 = EXCLUDED.address_street1,
    address_street2 = EXCLUDED.address_street2,
    address_city = EXCLUDED.address_city,
    address_state = EXCLUDED.address_state,
    address_postal_code = EXCLUDED.address_postal_code,
    address_country = EXCLUDED.address_country,
    address_country_code = EXCLUDED.address_country_code,
    contact_phone = EXCLUDED.contact_phone,
    contact_email = EXCLUDED.contact_email,
    contact_website = EXCLUDED.contact_website,
    contact_primary_contact = EXCLUDED.contact_primary_contact,
    portfolio_type_id = EXCLUDED.portfolio_type_id,
    risk_level_id = EXCLUDED.risk_level_id,
    compliance_status = EXCLUDED.compliance_status,
    status = EXCLUDED.status,
    updated_at = CURRENT_TIMESTAMP;

-- Insert sample merchant analytics data
INSERT INTO merchant_analytics (merchant_id, risk_score, compliance_score, transaction_volume, last_activity, flags, metadata)
SELECT 
    m.id,
    CASE 
        WHEN rl.level = 'low' THEN 0.15 + (random() * 0.20) -- 0.15-0.35
        WHEN rl.level = 'medium' THEN 0.35 + (random() * 0.30) -- 0.35-0.65
        WHEN rl.level = 'high' THEN 0.65 + (random() * 0.30) -- 0.65-0.95
    END as risk_score,
    CASE 
        WHEN m.compliance_status = 'compliant' THEN 0.80 + (random() * 0.20) -- 0.80-1.00
        WHEN m.compliance_status = 'pending' THEN 0.40 + (random() * 0.40) -- 0.40-0.80
        WHEN m.compliance_status = 'non_compliant' THEN 0.00 + (random() * 0.40) -- 0.00-0.40
    END as compliance_score,
    m.annual_revenue * (0.5 + random() * 1.0) as transaction_volume, -- 50-150% of annual revenue
    CURRENT_TIMESTAMP - (random() * INTERVAL '30 days') as last_activity,
    CASE 
        WHEN rl.level = 'high' THEN ARRAY['high_risk', 'enhanced_monitoring']
        WHEN m.compliance_status = 'non_compliant' THEN ARRAY['compliance_issue']
        WHEN m.employee_count > 100 THEN ARRAY['large_enterprise']
        ELSE ARRAY[]::TEXT[]
    END as flags,
    jsonb_build_object(
        'industry_risk', CASE 
            WHEN m.industry IN ('Finance', 'Cryptocurrency') THEN 'high'
            WHEN m.industry IN ('Healthcare', 'Manufacturing') THEN 'medium'
            ELSE 'low'
        END,
        'geographic_risk', CASE 
            WHEN m.address_country_code NOT IN ('US', 'CA', 'GB', 'DE', 'AU') THEN 'high'
            ELSE 'low'
        END,
        'size_category', CASE 
            WHEN m.employee_count > 200 THEN 'large'
            WHEN m.employee_count > 50 THEN 'medium'
            ELSE 'small'
        END
    ) as metadata
FROM merchants m
JOIN risk_levels rl ON m.risk_level_id = rl.id
WHERE m.status = 'active'
ON CONFLICT (merchant_id) DO UPDATE SET
    risk_score = EXCLUDED.risk_score,
    compliance_score = EXCLUDED.compliance_score,
    transaction_volume = EXCLUDED.transaction_volume,
    last_activity = EXCLUDED.last_activity,
    flags = EXCLUDED.flags,
    metadata = EXCLUDED.metadata,
    updated_at = CURRENT_TIMESTAMP;

-- Insert sample compliance records
INSERT INTO compliance_records (merchant_id, compliance_type, status, score, requirements, check_method, source, raw_data, checked_by)
SELECT 
    m.id,
    'kyc_verification',
    CASE 
        WHEN m.compliance_status = 'compliant' THEN 'passed'
        WHEN m.compliance_status = 'non_compliant' THEN 'failed'
        ELSE 'pending'
    END,
    CASE 
        WHEN m.compliance_status = 'compliant' THEN 0.85 + (random() * 0.15)
        WHEN m.compliance_status = 'non_compliant' THEN 0.00 + (random() * 0.30)
        ELSE 0.30 + (random() * 0.50)
    END,
    jsonb_build_object(
        'documents_verified', CASE WHEN m.compliance_status = 'compliant' THEN true ELSE false END,
        'identity_verified', CASE WHEN m.compliance_status = 'compliant' THEN true ELSE false END,
        'address_verified', CASE WHEN m.compliance_status = 'compliant' THEN true ELSE false END,
        'business_registration_verified', CASE WHEN m.compliance_status = 'compliant' THEN true ELSE false END
    ),
    'automated_screening',
    'internal_system',
    jsonb_build_object(
        'check_date', CURRENT_TIMESTAMP - (random() * INTERVAL '90 days'),
        'verification_method', 'document_analysis',
        'confidence_level', CASE 
            WHEN m.compliance_status = 'compliant' THEN 'high'
            WHEN m.compliance_status = 'non_compliant' THEN 'low'
            ELSE 'medium'
        END
    ),
    '00000000-0000-0000-0000-000000000001'
FROM merchants m
WHERE m.status = 'active'
ON CONFLICT DO NOTHING;

-- Insert sample audit logs for merchant operations
INSERT INTO merchant_audit_logs (user_id, merchant_id, action, resource_type, resource_id, details, ip_address, user_agent, request_id)
SELECT 
    '00000000-0000-0000-0000-000000000001',
    m.id,
    'merchant_created',
    'merchant',
    m.id::TEXT,
    jsonb_build_object(
        'merchant_name', m.name,
        'portfolio_type', pt.type,
        'risk_level', rl.level,
        'compliance_status', m.compliance_status
    ),
    '127.0.0.1'::INET,
    'KYB-Platform/1.0',
    'req_' || m.id
FROM merchants m
JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
JOIN risk_levels rl ON m.risk_level_id = rl.id
ON CONFLICT DO NOTHING;

-- Insert sample notifications for high-risk merchants
INSERT INTO merchant_notifications (merchant_id, user_id, type, title, message, priority, is_read)
SELECT 
    m.id,
    '00000000-0000-0000-0000-000000000001',
    'risk_alert',
    'High Risk Merchant Alert',
    'Merchant ' || m.name || ' has been flagged as high risk and requires enhanced monitoring.',
    'high',
    false
FROM merchants m
JOIN risk_levels rl ON m.risk_level_id = rl.id
WHERE rl.level = 'high' AND m.status = 'active'
ON CONFLICT DO NOTHING;

-- Insert sample notifications for compliance issues
INSERT INTO merchant_notifications (merchant_id, user_id, type, title, message, priority, is_read)
SELECT 
    m.id,
    '00000000-0000-0000-0000-000000000001',
    'compliance',
    'Compliance Review Required',
    'Merchant ' || m.name || ' requires compliance review due to status: ' || m.compliance_status,
    'medium',
    false
FROM merchants m
WHERE m.compliance_status IN ('pending', 'non_compliant') AND m.status = 'active'
ON CONFLICT DO NOTHING;

-- Create sample bulk operation for testing
INSERT INTO bulk_operations (operation_id, user_id, operation_type, status, total_items, processed, successful, failed, metadata)
VALUES (
    'bulk_001_' || EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::TEXT,
    '00000000-0000-0000-0000-000000000001',
    'portfolio_type_update',
    'completed',
    20,
    20,
    18,
    2,
    jsonb_build_object(
        'target_portfolio_type', 'onboarded',
        'source_portfolio_type', 'prospective',
        'operation_date', CURRENT_TIMESTAMP,
        'notes', 'Bulk update of prospective merchants to onboarded status'
    )
)
ON CONFLICT (operation_id) DO NOTHING;

-- Insert sample bulk operation items
INSERT INTO bulk_operation_items (bulk_operation_id, merchant_id, status, error_message, result_data, processed_at)
SELECT 
    bo.id,
    m.id,
    CASE 
        WHEN m.compliance_status = 'compliant' THEN 'completed'
        WHEN m.compliance_status = 'non_compliant' THEN 'failed'
        ELSE 'completed'
    END,
    CASE 
        WHEN m.compliance_status = 'non_compliant' THEN 'Compliance check failed'
        ELSE NULL
    END,
    jsonb_build_object(
        'previous_portfolio_type', 'prospective',
        'new_portfolio_type', 'onboarded',
        'updated_at', CURRENT_TIMESTAMP
    ),
    CURRENT_TIMESTAMP - (random() * INTERVAL '1 hour')
FROM bulk_operations bo
CROSS JOIN merchants m
JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
WHERE bo.operation_type = 'portfolio_type_update' 
    AND pt.type = 'onboarded'
    AND m.status = 'active'
LIMIT 20
ON CONFLICT DO NOTHING;

-- Add comments for documentation
COMMENT ON TABLE portfolio_types IS 'Lookup table for merchant portfolio types - seeded with onboarded, prospective, pending, deactivated';
COMMENT ON TABLE risk_levels IS 'Lookup table for merchant risk levels - seeded with low, medium, high risk levels';
COMMENT ON TABLE merchants IS 'Main merchants table - seeded with 20 diverse mock merchants across different industries and risk levels';

-- Verify data insertion
DO $$
DECLARE
    portfolio_count INTEGER;
    risk_count INTEGER;
    merchant_count INTEGER;
    analytics_count INTEGER;
    compliance_count INTEGER;
    audit_count INTEGER;
    notification_count INTEGER;
    bulk_op_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO portfolio_count FROM portfolio_types;
    SELECT COUNT(*) INTO risk_count FROM risk_levels;
    SELECT COUNT(*) INTO merchant_count FROM merchants;
    SELECT COUNT(*) INTO analytics_count FROM merchant_analytics;
    SELECT COUNT(*) INTO compliance_count FROM compliance_records;
    SELECT COUNT(*) INTO audit_count FROM merchant_audit_logs;
    SELECT COUNT(*) INTO notification_count FROM merchant_notifications;
    SELECT COUNT(*) INTO bulk_op_count FROM bulk_operations;
    
    RAISE NOTICE 'Seed data insertion completed:';
    RAISE NOTICE '  Portfolio Types: %', portfolio_count;
    RAISE NOTICE '  Risk Levels: %', risk_count;
    RAISE NOTICE '  Merchants: %', merchant_count;
    RAISE NOTICE '  Analytics Records: %', analytics_count;
    RAISE NOTICE '  Compliance Records: %', compliance_count;
    RAISE NOTICE '  Audit Logs: %', audit_count;
    RAISE NOTICE '  Notifications: %', notification_count;
    RAISE NOTICE '  Bulk Operations: %', bulk_op_count;
END $$;
