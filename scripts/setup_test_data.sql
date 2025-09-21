-- Enhanced Classification System Test Data Setup
-- This script sets up comprehensive test data for testing the enhanced classification system

-- =====================================================
-- 1. Test Data for Risk Keywords
-- =====================================================

-- Insert test risk keywords for comprehensive testing
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, mcc_codes, naics_codes, sic_codes, card_brand_restrictions, detection_patterns, synonyms, is_active) VALUES
-- Illegal Activities (Critical Risk)
('drug trafficking', 'illegal', 'critical', 'Illegal drug trafficking activities', '{}', '{}', '{}', '{"Visa", "Mastercard", "Amex"}', '{}', '{"drug dealing", "narcotics", "cocaine", "heroin", "marijuana"}', true),
('weapons trafficking', 'illegal', 'critical', 'Illegal weapons trafficking activities', '{}', '{}', '{}', '{"Visa", "Mastercard", "Amex"}', '{}', '{"arms dealing", "firearms", "weapons", "guns"}', true),
('human trafficking', 'illegal', 'critical', 'Human trafficking activities', '{}', '{}', '{}', '{"Visa", "Mastercard", "Amex"}', '{}', '{"sex trafficking", "forced labor", "slavery"}', true),
('money laundering', 'illegal', 'critical', 'Money laundering activities', '{}', '{}', '{}', '{"Visa", "Mastercard", "Amex"}', '{}', '{"financial crime", "terrorist financing", "dirty money"}', true),

-- Prohibited by Card Brands (High Risk)
('casino', 'prohibited', 'high', 'Casino and gambling activities', '{"7995"}', '{"713210"}', '{"7995"}', '{"Visa", "Mastercard", "Amex"}', '{}', '{"gambling", "poker", "blackjack", "slot games", "betting"}', true),
('adult entertainment', 'prohibited', 'high', 'Adult entertainment services', '{"7273"}', '{"713120"}', '{"7273"}', '{"Visa", "Mastercard", "Amex"}', '{}', '{"pornography", "strip club", "escort services", "adult content"}', true),
('cryptocurrency', 'prohibited', 'high', 'Cryptocurrency and digital currency services', '{"6012"}', '{"523130"}', '{"6012"}', '{"Visa", "Mastercard", "Amex"}', '{}', '{"bitcoin", "ethereum", "digital currency", "crypto exchange"}', true),
('tobacco', 'prohibited', 'high', 'Tobacco products and sales', '{"5993"}', '{"453991"}', '{"5993"}', '{"Visa", "Mastercard", "Amex"}', '{}', '{"cigarettes", "cigars", "smoking", "tobacco products"}', true),

-- High-Risk Industries (Medium-High Risk)
('money services', 'high_risk', 'medium', 'Money services and transfer services', '{"6010", "6011"}', '{"522320"}', '{"6010", "6011"}', '{"Visa", "Mastercard"}', '{}', '{"money transfer", "remittance", "wire transfer", "check cashing"}', true),
('prepaid cards', 'high_risk', 'medium', 'Prepaid card services', '{"6012"}', '{"522320"}', '{"6012"}', '{"Visa", "Mastercard"}', '{}', '{"gift cards", "prepaid", "stored value", "cash cards"}', true),
('pharmaceuticals', 'high_risk', 'medium', 'Pharmaceutical products and services', '{"5912"}', '{"446110"}', '{"5912"}', '{"Visa", "Mastercard"}', '{}', '{"prescription drugs", "medications", "pharmacy", "drugs"}', true),

-- Trade-Based Money Laundering (TBML)
('shell company', 'tbml', 'high', 'Shell company indicators', '{}', '{}', '{}', '{"Visa", "Mastercard", "Amex"}', '{}', '{"front company", "dummy company", "paper company"}', true),
('trade finance', 'tbml', 'medium', 'Trade finance and import/export', '{"5999"}', '{"423990"}', '{"5999"}', '{"Visa", "Mastercard"}', '{}', '{"import export", "trade", "commodities", "precious metals"}', true),

-- Fraud Indicators (Medium Risk)
('fake business', 'fraud', 'medium', 'Fake business indicators', '{}', '{}', '{}', '{"Visa", "Mastercard", "Amex"}', '{}', '{"stolen identity", "fake company", "fraudulent business"}', true),
('rapid changes', 'fraud', 'low', 'Rapid business changes', '{}', '{}', '{}', '{}', '{}', '{"high turnover", "frequent changes", "unstable business"}', true);

-- =====================================================
-- 2. Test Data for Industry Code Crosswalks
-- =====================================================

-- Insert test crosswalk mappings for comprehensive testing
INSERT INTO crosswalk_mappings (industry_id, source_code, source_system, target_code, target_system, mcc_code, naics_code, sic_code, description, confidence_score, is_valid, metadata) VALUES
-- Technology Industry Mappings
(1, '5734', 'MCC', '1', 'INDUSTRY', '5734', NULL, NULL, 'Computer Software Stores', 0.90, true, '{"industry_name": "Technology", "industry_category": "Software", "mapping_method": "direct_match"}'),
(1, '541511', 'NAICS', '1', 'INDUSTRY', NULL, '541511', NULL, 'Custom Computer Programming Services', 0.95, true, '{"industry_name": "Technology", "industry_category": "Software", "mapping_method": "direct_match"}'),
(1, '7372', 'SIC', '1', 'INDUSTRY', NULL, NULL, '7372', 'Prepackaged Software', 0.85, true, '{"industry_name": "Technology", "industry_category": "Software", "mapping_method": "direct_match"}'),

-- Financial Services Industry Mappings
(2, '6010', 'MCC', '2', 'INDUSTRY', '6010', NULL, NULL, 'Manual Cash Disbursements', 0.88, true, '{"industry_name": "Financial Services", "industry_category": "Banking", "mapping_method": "direct_match"}'),
(2, '522110', 'NAICS', '2', 'INDUSTRY', NULL, '522110', NULL, 'Commercial Banking', 0.92, true, '{"industry_name": "Financial Services", "industry_category": "Banking", "mapping_method": "direct_match"}'),
(2, '6021', 'SIC', '2', 'INDUSTRY', NULL, NULL, '6021', 'National Commercial Banks', 0.90, true, '{"industry_name": "Financial Services", "industry_category": "Banking", "mapping_method": "direct_match"}'),

-- Healthcare Industry Mappings
(3, '8011', 'MCC', '3', 'INDUSTRY', '8011', NULL, NULL, 'Doctors and Physicians', 0.93, true, '{"industry_name": "Healthcare", "industry_category": "Medical Services", "mapping_method": "direct_match"}'),
(3, '621111', 'NAICS', '3', 'INDUSTRY', NULL, '621111', NULL, 'Offices of Physicians', 0.94, true, '{"industry_name": "Healthcare", "industry_category": "Medical Services", "mapping_method": "direct_match"}'),
(3, '8011', 'SIC', '3', 'INDUSTRY', NULL, NULL, '8011', 'Offices and Clinics of Doctors of Medicine', 0.91, true, '{"industry_name": "Healthcare", "industry_category": "Medical Services", "mapping_method": "direct_match"}'),

-- Retail Industry Mappings
(4, '5310', 'MCC', '4', 'INDUSTRY', '5310', NULL, NULL, 'Discount Stores', 0.87, true, '{"industry_name": "Retail", "industry_category": "General Merchandise", "mapping_method": "direct_match"}'),
(4, '452111', 'NAICS', '4', 'INDUSTRY', NULL, '452111', NULL, 'Department Stores', 0.89, true, '{"industry_name": "Retail", "industry_category": "General Merchandise", "mapping_method": "direct_match"}'),
(4, '5311', 'SIC', '4', 'INDUSTRY', NULL, NULL, '5311', 'Department Stores', 0.86, true, '{"industry_name": "Retail", "industry_category": "General Merchandise", "mapping_method": "direct_match"}'),

-- Manufacturing Industry Mappings
(5, '5085', 'MCC', '5', 'INDUSTRY', '5085', NULL, NULL, 'Industrial Supplies', 0.84, true, '{"industry_name": "Manufacturing", "industry_category": "Industrial", "mapping_method": "direct_match"}'),
(5, '423990', 'NAICS', '5', 'INDUSTRY', NULL, '423990', NULL, 'Other Miscellaneous Durable Goods Merchant Wholesalers', 0.82, true, '{"industry_name": "Manufacturing", "industry_category": "Industrial", "mapping_method": "direct_match"}'),
(5, '5085', 'SIC', '5', 'INDUSTRY', NULL, NULL, '5085', 'Industrial Supplies', 0.83, true, '{"industry_name": "Manufacturing", "industry_category": "Industrial", "mapping_method": "direct_match"}'),

-- High-Risk Industry Mappings
(6, '7995', 'MCC', '6', 'INDUSTRY', '7995', NULL, NULL, 'Gambling', 0.95, true, '{"industry_name": "Gambling", "industry_category": "High Risk", "mapping_method": "direct_match"}'),
(6, '713210', 'NAICS', '6', 'INDUSTRY', NULL, '713210', NULL, 'Casinos (except Casino Hotels)', 0.96, true, '{"industry_name": "Gambling", "industry_category": "High Risk", "mapping_method": "direct_match"}'),
(6, '7995', 'SIC', '6', 'INDUSTRY', NULL, NULL, '7995', 'Gambling', 0.94, true, '{"industry_name": "Gambling", "industry_category": "High Risk", "mapping_method": "direct_match"}');

-- =====================================================
-- 3. Test Data for Business Risk Assessments
-- =====================================================

-- Insert test business risk assessments for comprehensive testing
INSERT INTO business_risk_assessments (business_id, risk_keyword_id, detected_keywords, risk_score, risk_level, assessment_method, website_content, detected_patterns, assessment_date) VALUES
-- High-risk business assessments
(gen_random_uuid(), 1, '{"casino", "gambling", "poker"}', 0.85, 'high', 'keyword_matching', 'Welcome to our online casino and gambling platform. We offer poker, blackjack, and slot games.', '{"casino": 0.9, "gambling": 0.95, "poker": 0.8}', NOW()),
(gen_random_uuid(), 2, '{"adult entertainment", "escort services"}', 0.90, 'high', 'keyword_matching', 'We provide adult entertainment and escort services for discerning clients.', '{"adult entertainment": 0.95, "escort services": 0.85}', NOW()),
(gen_random_uuid(), 3, '{"cryptocurrency", "bitcoin", "ethereum"}', 0.70, 'medium', 'keyword_matching', 'Our cryptocurrency exchange platform supports Bitcoin, Ethereum, and other digital currencies.', '{"cryptocurrency": 0.8, "bitcoin": 0.9, "ethereum": 0.85}', NOW()),

-- Low-risk business assessments
(gen_random_uuid(), NULL, '{}', 0.15, 'low', 'keyword_matching', 'Welcome to our family restaurant. We serve delicious meals and provide excellent customer service.', '{}', NOW()),
(gen_random_uuid(), NULL, '{}', 0.20, 'low', 'keyword_matching', 'We are a technology consulting company providing software development and IT services.', '{}', NOW()),
(gen_random_uuid(), NULL, '{}', 0.10, 'low', 'keyword_matching', 'Our local bookstore offers a wide selection of books and educational materials.', '{}', NOW()),

-- Medium-risk business assessments
(gen_random_uuid(), 4, '{"pharmaceuticals", "prescription drugs"}', 0.45, 'medium', 'keyword_matching', 'We are a licensed pharmaceutical distributor providing prescription medications.', '{"pharmaceuticals": 0.7, "prescription drugs": 0.8}', NOW()),
(gen_random_uuid(), 5, '{"money services", "money transfer"}', 0.55, 'medium', 'keyword_matching', 'Our money transfer services help customers send money internationally.', '{"money services": 0.6, "money transfer": 0.7}', NOW());

-- =====================================================
-- 4. Test Data for Industries
-- =====================================================

-- Insert test industries for comprehensive testing
INSERT INTO industries (name, category, description, keywords, is_active, created_at, updated_at) VALUES
('Technology', 'Software', 'Technology and software development services', '{"software", "technology", "IT", "computer", "programming", "development"}', true, NOW(), NOW()),
('Financial Services', 'Banking', 'Financial services and banking', '{"banking", "financial", "finance", "money", "credit", "loans"}', true, NOW(), NOW()),
('Healthcare', 'Medical Services', 'Healthcare and medical services', '{"healthcare", "medical", "doctor", "physician", "hospital", "clinic"}', true, NOW(), NOW()),
('Retail', 'General Merchandise', 'Retail and general merchandise', '{"retail", "store", "merchandise", "shopping", "products", "sales"}', true, NOW(), NOW()),
('Manufacturing', 'Industrial', 'Manufacturing and industrial services', '{"manufacturing", "industrial", "production", "factory", "machinery", "equipment"}', true, NOW(), NOW()),
('Gambling', 'High Risk', 'Gambling and gaming services', '{"gambling", "casino", "gaming", "poker", "betting", "lottery"}', true, NOW(), NOW()),
('Adult Entertainment', 'High Risk', 'Adult entertainment services', '{"adult", "entertainment", "pornography", "escort", "strip", "adult content"}', true, NOW(), NOW()),
('Cryptocurrency', 'High Risk', 'Cryptocurrency and digital currency services', '{"cryptocurrency", "bitcoin", "ethereum", "digital currency", "crypto", "blockchain"}', true, NOW(), NOW());

-- =====================================================
-- 5. Test Data for Classification Codes
-- =====================================================

-- Insert test classification codes for comprehensive testing
INSERT INTO classification_codes (code, code_type, description, industry_id, confidence_score, is_active, created_at, updated_at) VALUES
-- MCC Codes
('5734', 'MCC', 'Computer Software Stores', 1, 0.90, true, NOW(), NOW()),
('6010', 'MCC', 'Manual Cash Disbursements', 2, 0.88, true, NOW(), NOW()),
('8011', 'MCC', 'Doctors and Physicians', 3, 0.93, true, NOW(), NOW()),
('5310', 'MCC', 'Discount Stores', 4, 0.87, true, NOW(), NOW()),
('5085', 'MCC', 'Industrial Supplies', 5, 0.84, true, NOW(), NOW()),
('7995', 'MCC', 'Gambling', 6, 0.95, true, NOW(), NOW()),
('7273', 'MCC', 'Adult Entertainment', 7, 0.92, true, NOW(), NOW()),
('6012', 'MCC', 'Merchandise Services', 8, 0.85, true, NOW(), NOW()),

-- NAICS Codes
('541511', 'NAICS', 'Custom Computer Programming Services', 1, 0.95, true, NOW(), NOW()),
('522110', 'NAICS', 'Commercial Banking', 2, 0.92, true, NOW(), NOW()),
('621111', 'NAICS', 'Offices of Physicians', 3, 0.94, true, NOW(), NOW()),
('452111', 'NAICS', 'Department Stores', 4, 0.89, true, NOW(), NOW()),
('423990', 'NAICS', 'Other Miscellaneous Durable Goods Merchant Wholesalers', 5, 0.82, true, NOW(), NOW()),
('713210', 'NAICS', 'Casinos (except Casino Hotels)', 6, 0.96, true, NOW(), NOW()),
('713120', 'NAICS', 'Amusement Arcades', 7, 0.88, true, NOW(), NOW()),
('523130', 'NAICS', 'Securities and Commodity Exchanges', 8, 0.90, true, NOW(), NOW()),

-- SIC Codes
('7372', 'SIC', 'Prepackaged Software', 1, 0.85, true, NOW(), NOW()),
('6021', 'SIC', 'National Commercial Banks', 2, 0.90, true, NOW(), NOW()),
('8011', 'SIC', 'Offices and Clinics of Doctors of Medicine', 3, 0.91, true, NOW(), NOW()),
('5311', 'SIC', 'Department Stores', 4, 0.86, true, NOW(), NOW()),
('5085', 'SIC', 'Industrial Supplies', 5, 0.83, true, NOW(), NOW()),
('7995', 'SIC', 'Gambling', 6, 0.94, true, NOW(), NOW()),
('7273', 'SIC', 'Adult Entertainment', 7, 0.89, true, NOW(), NOW()),
('6012', 'SIC', 'Merchandise Services', 8, 0.87, true, NOW(), NOW());

-- =====================================================
-- 6. Create Test Indexes for Performance
-- =====================================================

-- Create additional indexes for test performance
CREATE INDEX IF NOT EXISTS idx_test_risk_keywords_category_severity ON risk_keywords(risk_category, risk_severity) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_test_crosswalk_mappings_confidence ON crosswalk_mappings(confidence_score) WHERE is_valid = true;
CREATE INDEX IF NOT EXISTS idx_test_business_risk_assessments_score ON business_risk_assessments(risk_score) WHERE risk_score > 0;
CREATE INDEX IF NOT EXISTS idx_test_industries_category ON industries(category) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_test_classification_codes_type ON classification_codes(code_type) WHERE is_active = true;

-- =====================================================
-- 7. Test Data Validation
-- =====================================================

-- Verify test data integrity
DO $$
DECLARE
    risk_keywords_count INTEGER;
    crosswalk_mappings_count INTEGER;
    business_assessments_count INTEGER;
    industries_count INTEGER;
    classification_codes_count INTEGER;
BEGIN
    -- Count test data
    SELECT COUNT(*) INTO risk_keywords_count FROM risk_keywords WHERE is_active = true;
    SELECT COUNT(*) INTO crosswalk_mappings_count FROM crosswalk_mappings WHERE is_valid = true;
    SELECT COUNT(*) INTO business_assessments_count FROM business_risk_assessments;
    SELECT COUNT(*) INTO industries_count FROM industries WHERE is_active = true;
    SELECT COUNT(*) INTO classification_codes_count FROM classification_codes WHERE is_active = true;
    
    -- Log test data counts
    RAISE NOTICE 'Test data setup completed:';
    RAISE NOTICE 'Risk Keywords: %', risk_keywords_count;
    RAISE NOTICE 'Crosswalk Mappings: %', crosswalk_mappings_count;
    RAISE NOTICE 'Business Risk Assessments: %', business_assessments_count;
    RAISE NOTICE 'Industries: %', industries_count;
    RAISE NOTICE 'Classification Codes: %', classification_codes_count;
    
    -- Validate minimum data requirements
    IF risk_keywords_count < 10 THEN
        RAISE EXCEPTION 'Insufficient risk keywords test data: %', risk_keywords_count;
    END IF;
    
    IF crosswalk_mappings_count < 20 THEN
        RAISE EXCEPTION 'Insufficient crosswalk mappings test data: %', crosswalk_mappings_count;
    END IF;
    
    IF business_assessments_count < 5 THEN
        RAISE EXCEPTION 'Insufficient business risk assessments test data: %', business_assessments_count;
    END IF;
    
    IF industries_count < 5 THEN
        RAISE EXCEPTION 'Insufficient industries test data: %', industries_count;
    END IF;
    
    IF classification_codes_count < 20 THEN
        RAISE EXCEPTION 'Insufficient classification codes test data: %', classification_codes_count;
    END IF;
    
    RAISE NOTICE 'All test data validation checks passed!';
END $$;

-- =====================================================
-- 8. Test Data Cleanup Function
-- =====================================================

-- Create function to clean up test data
CREATE OR REPLACE FUNCTION cleanup_test_data()
RETURNS void AS $$
BEGIN
    -- Clean up test data (in reverse order of dependencies)
    DELETE FROM business_risk_assessments WHERE assessment_date > NOW() - INTERVAL '1 day';
    DELETE FROM crosswalk_mappings WHERE created_at > NOW() - INTERVAL '1 day';
    DELETE FROM classification_codes WHERE created_at > NOW() - INTERVAL '1 day';
    DELETE FROM industries WHERE created_at > NOW() - INTERVAL '1 day';
    DELETE FROM risk_keywords WHERE created_at > NOW() - INTERVAL '1 day';
    
    RAISE NOTICE 'Test data cleanup completed';
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 9. Test Data Summary
-- =====================================================

-- Display test data summary
SELECT 
    'Risk Keywords' as table_name,
    COUNT(*) as record_count,
    COUNT(DISTINCT risk_category) as categories,
    COUNT(DISTINCT risk_severity) as severity_levels
FROM risk_keywords 
WHERE is_active = true

UNION ALL

SELECT 
    'Crosswalk Mappings' as table_name,
    COUNT(*) as record_count,
    COUNT(DISTINCT source_system) as source_systems,
    COUNT(DISTINCT target_system) as target_systems
FROM crosswalk_mappings 
WHERE is_valid = true

UNION ALL

SELECT 
    'Business Risk Assessments' as table_name,
    COUNT(*) as record_count,
    COUNT(DISTINCT risk_level) as risk_levels,
    AVG(risk_score)::numeric(3,2) as avg_risk_score
FROM business_risk_assessments

UNION ALL

SELECT 
    'Industries' as table_name,
    COUNT(*) as record_count,
    COUNT(DISTINCT category) as categories,
    0 as additional_metric
FROM industries 
WHERE is_active = true

UNION ALL

SELECT 
    'Classification Codes' as table_name,
    COUNT(*) as record_count,
    COUNT(DISTINCT code_type) as code_types,
    COUNT(DISTINCT industry_id) as industries
FROM classification_codes 
WHERE is_active = true;

-- =====================================================
-- Test Data Setup Complete
-- =====================================================

-- Log completion
DO $$
BEGIN
    RAISE NOTICE '=====================================================';
    RAISE NOTICE 'Enhanced Classification System Test Data Setup Complete';
    RAISE NOTICE '=====================================================';
    RAISE NOTICE 'Test data has been successfully inserted for:';
    RAISE NOTICE '- Risk Keywords (with categories and severity levels)';
    RAISE NOTICE '- Industry Code Crosswalks (MCC/NAICS/SIC mappings)';
    RAISE NOTICE '- Business Risk Assessments (various risk levels)';
    RAISE NOTICE '- Industries (with categories and keywords)';
    RAISE NOTICE '- Classification Codes (MCC/NAICS/SIC codes)';
    RAISE NOTICE '=====================================================';
    RAISE NOTICE 'Ready for comprehensive testing of enhanced classification system';
    RAISE NOTICE '=====================================================';
END $$;
