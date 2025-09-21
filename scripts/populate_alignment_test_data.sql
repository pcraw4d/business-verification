-- Populate alignment test data for classification alignment testing
-- This script creates sample industries and crosswalk mappings for testing

-- Insert test industries
INSERT INTO industries (id, name, description, created_at, updated_at) VALUES
(1, 'Technology', 'Technology and software companies', NOW(), NOW()),
(2, 'Healthcare', 'Healthcare and medical services', NOW(), NOW()),
(3, 'Manufacturing', 'Manufacturing and production companies', NOW(), NOW()),
(4, 'Financial Services', 'Banking and financial services', NOW(), NOW()),
(5, 'Retail', 'Retail and consumer goods', NOW(), NOW()),
(6, 'Education', 'Educational institutions and services', NOW(), NOW()),
(7, 'Transportation', 'Transportation and logistics', NOW(), NOW()),
(8, 'Energy', 'Energy and utilities', NOW(), NOW()),
(9, 'Construction', 'Construction and building services', NOW(), NOW()),
(10, 'Agriculture', 'Agriculture and farming', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = EXCLUDED.updated_at;

-- Insert test crosswalk mappings with various alignment scenarios

-- Well-aligned mappings (high confidence, proper codes)
INSERT INTO crosswalk_mappings (id, industry_id, mcc_code, naics_code, sic_code, description, confidence_score, created_at, updated_at) VALUES
-- Technology industry - well aligned
(1, 1, '5734', '541511', '7371', 'Computer Software Stores', 0.95, NOW(), NOW()),
(2, 1, '5735', '541512', '7372', 'Computer Software Publishers', 0.92, NOW(), NOW()),
(3, 1, '5814', '541511', '7371', 'Fast Food Restaurants (Tech)', 0.88, NOW(), NOW()),

-- Healthcare industry - well aligned
(4, 2, '8011', '621111', '8011', 'Offices of Physicians', 0.96, NOW(), NOW()),
(5, 2, '8021', '621112', '8021', 'Offices of Dentists', 0.94, NOW(), NOW()),
(6, 2, '8062', '622110', '8062', 'General Medical and Surgical Hospitals', 0.97, NOW(), NOW()),

-- Manufacturing industry - well aligned
(7, 3, '3571', '334111', '3571', 'Electronic Computers', 0.93, NOW(), NOW()),
(8, 3, '3572', '334112', '3572', 'Computer Storage Devices', 0.91, NOW(), NOW()),
(9, 3, '3577', '334118', '3577', 'Computer Terminals', 0.89, NOW(), NOW()),

-- Financial Services industry - well aligned
(10, 4, '6011', '522110', '6021', 'Automated Teller Machine Services', 0.95, NOW(), NOW()),
(11, 4, '6012', '522110', '6022', 'Financial Transactions', 0.93, NOW(), NOW()),
(12, 4, '6019', '522110', '6029', 'Other Financial Services', 0.90, NOW(), NOW()),

-- Retail industry - well aligned
(13, 5, '5310', '448140', '5311', 'Department Stores', 0.94, NOW(), NOW()),
(14, 5, '5311', '448140', '5311', 'Department Stores', 0.92, NOW(), NOW()),
(15, 5, '5331', '448140', '5331', 'Variety Stores', 0.88, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    industry_id = EXCLUDED.industry_id,
    mcc_code = EXCLUDED.mcc_code,
    naics_code = EXCLUDED.naics_code,
    sic_code = EXCLUDED.sic_code,
    description = EXCLUDED.description,
    confidence_score = EXCLUDED.confidence_score,
    updated_at = EXCLUDED.updated_at;

-- Misaligned mappings (low confidence, hierarchy issues)
INSERT INTO crosswalk_mappings (id, industry_id, mcc_code, naics_code, sic_code, description, confidence_score, created_at, updated_at) VALUES
-- Education industry - misaligned (low confidence)
(16, 6, '8220', '611110', '8221', 'Colleges and Universities', 0.65, NOW(), NOW()),
(17, 6, '8221', '611110', '8221', 'Colleges and Universities', 0.62, NOW(), NOW()),

-- Transportation industry - misaligned (hierarchy issues)
(18, 7, '4111', '481111', '4111', 'Local and Suburban Transportation', 0.75, NOW(), NOW()),
(19, 7, '4119', '481112', '4119', 'Local and Suburban Transportation', 0.70, NOW(), NOW()),

-- Energy industry - misaligned (mixed issues)
(20, 8, '4911', '221122', '4911', 'Electric Services', 0.68, NOW(), NOW()),
(21, 8, '4922', '221122', '4922', 'Natural Gas Distribution', 0.72, NOW(), NOW()),

-- Construction industry - misaligned (low confidence)
(22, 9, '1520', '236115', '1520', 'General Contractors', 0.58, NOW(), NOW()),
(23, 9, '1531', '236116', '1531', 'Operative Builders', 0.61, NOW(), NOW()),

-- Agriculture industry - misaligned (hierarchy issues)
(24, 10, '0111', '111110', '0111', 'Wheat Farming', 0.66, NOW(), NOW()),
(25, 10, '0112', '111120', '0112', 'Rice Farming', 0.69, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    industry_id = EXCLUDED.industry_id,
    mcc_code = EXCLUDED.mcc_code,
    naics_code = EXCLUDED.naics_code,
    sic_code = EXCLUDED.sic_code,
    description = EXCLUDED.description,
    confidence_score = EXCLUDED.confidence_score,
    updated_at = EXCLUDED.updated_at;

-- Mappings with gaps (missing codes)
INSERT INTO crosswalk_mappings (id, industry_id, mcc_code, naics_code, sic_code, description, confidence_score, created_at, updated_at) VALUES
-- Technology industry - missing NAICS and SIC
(26, 1, '5734', '', '', 'Computer Software Stores (MCC only)', 0.90, NOW(), NOW()),

-- Healthcare industry - missing MCC and SIC
(27, 2, '', '621111', '', 'Offices of Physicians (NAICS only)', 0.95, NOW(), NOW()),

-- Manufacturing industry - missing MCC and NAICS
(28, 3, '', '', '3571', 'Electronic Computers (SIC only)', 0.88, NOW(), NOW()),

-- Financial Services industry - missing SIC
(29, 4, '6011', '522110', '', 'ATM Services (MCC and NAICS only)', 0.92, NOW(), NOW()),

-- Retail industry - missing NAICS
(30, 5, '5310', '', '5311', 'Department Stores (MCC and SIC only)', 0.89, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    industry_id = EXCLUDED.industry_id,
    mcc_code = EXCLUDED.mcc_code,
    naics_code = EXCLUDED.naics_code,
    sic_code = EXCLUDED.sic_code,
    description = EXCLUDED.description,
    confidence_score = EXCLUDED.confidence_score,
    updated_at = EXCLUDED.updated_at;

-- Invalid hierarchy mappings (for testing hierarchy validation)
INSERT INTO crosswalk_mappings (id, industry_id, mcc_code, naics_code, sic_code, description, confidence_score, created_at, updated_at) VALUES
-- Invalid NAICS hierarchy (invalid sector)
(31, 1, '5734', '999999', '7371', 'Invalid NAICS Sector', 0.85, NOW(), NOW()),

-- Invalid SIC hierarchy (invalid division)
(32, 2, '8011', '621111', '9999', 'Invalid SIC Division', 0.87, NOW(), NOW()),

-- Invalid format codes
(33, 3, '123', '12345', '123', 'Invalid Format Codes', 0.80, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    industry_id = EXCLUDED.industry_id,
    mcc_code = EXCLUDED.mcc_code,
    naics_code = EXCLUDED.naics_code,
    sic_code = EXCLUDED.sic_code,
    description = EXCLUDED.description,
    confidence_score = EXCLUDED.confidence_score,
    updated_at = EXCLUDED.updated_at;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_industry_id ON crosswalk_mappings(industry_id);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_mcc_code ON crosswalk_mappings(mcc_code);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_naics_code ON crosswalk_mappings(naics_code);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_sic_code ON crosswalk_mappings(sic_code);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_confidence_score ON crosswalk_mappings(confidence_score);

-- Create views for alignment analysis
CREATE OR REPLACE VIEW alignment_summary AS
SELECT 
    i.id as industry_id,
    i.name as industry_name,
    COUNT(CASE WHEN cm.mcc_code IS NOT NULL AND cm.mcc_code != '' THEN 1 END) as mcc_count,
    COUNT(CASE WHEN cm.naics_code IS NOT NULL AND cm.naics_code != '' THEN 1 END) as naics_count,
    COUNT(CASE WHEN cm.sic_code IS NOT NULL AND cm.sic_code != '' THEN 1 END) as sic_count,
    AVG(cm.confidence_score) as avg_confidence_score,
    MIN(cm.confidence_score) as min_confidence_score,
    MAX(cm.confidence_score) as max_confidence_score
FROM industries i
LEFT JOIN crosswalk_mappings cm ON i.id = cm.industry_id
GROUP BY i.id, i.name
ORDER BY i.name;

CREATE OR REPLACE VIEW alignment_issues AS
SELECT 
    i.id as industry_id,
    i.name as industry_name,
    CASE 
        WHEN COUNT(CASE WHEN cm.mcc_code IS NOT NULL AND cm.mcc_code != '' THEN 1 END) = 0 THEN 'Missing MCC'
        WHEN COUNT(CASE WHEN cm.naics_code IS NOT NULL AND cm.naics_code != '' THEN 1 END) = 0 THEN 'Missing NAICS'
        WHEN COUNT(CASE WHEN cm.sic_code IS NOT NULL AND cm.sic_code != '' THEN 1 END) = 0 THEN 'Missing SIC'
        WHEN AVG(cm.confidence_score) < 0.8 THEN 'Low Confidence'
        ELSE 'No Issues'
    END as issue_type,
    AVG(cm.confidence_score) as avg_confidence_score,
    COUNT(cm.id) as total_mappings
FROM industries i
LEFT JOIN crosswalk_mappings cm ON i.id = cm.industry_id
GROUP BY i.id, i.name
HAVING 
    COUNT(CASE WHEN cm.mcc_code IS NOT NULL AND cm.mcc_code != '' THEN 1 END) = 0 OR
    COUNT(CASE WHEN cm.naics_code IS NOT NULL AND cm.naics_code != '' THEN 1 END) = 0 OR
    COUNT(CASE WHEN cm.sic_code IS NOT NULL AND cm.sic_code != '' THEN 1 END) = 0 OR
    AVG(cm.confidence_score) < 0.8
ORDER BY avg_confidence_score ASC;

-- Create function to validate NAICS hierarchy
CREATE OR REPLACE FUNCTION is_valid_naics_sector(naics_code TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    IF LENGTH(naics_code) != 6 THEN
        RETURN FALSE;
    END IF;
    
    RETURN SUBSTRING(naics_code, 1, 2) IN (
        '11', '21', '22', '23', '31', '32', '33', '42', '44', '45', 
        '48', '49', '51', '52', '53', '54', '55', '56', '61', '62', 
        '71', '72', '81', '92'
    );
END;
$$ LANGUAGE plpgsql;

-- Create function to validate SIC division
CREATE OR REPLACE FUNCTION is_valid_sic_division(sic_code TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    IF LENGTH(sic_code) != 4 THEN
        RETURN FALSE;
    END IF;
    
    RETURN SUBSTRING(sic_code, 1, 1) ~ '^[0-9]$';
END;
$$ LANGUAGE plpgsql;

-- Create view for hierarchy validation
CREATE OR REPLACE VIEW hierarchy_validation AS
SELECT 
    cm.id,
    cm.industry_id,
    i.name as industry_name,
    cm.mcc_code,
    cm.naics_code,
    cm.sic_code,
    cm.confidence_score,
    CASE 
        WHEN cm.naics_code IS NOT NULL AND cm.naics_code != '' AND NOT is_valid_naics_sector(cm.naics_code) THEN 'Invalid NAICS'
        WHEN cm.sic_code IS NOT NULL AND cm.sic_code != '' AND NOT is_valid_sic_division(cm.sic_code) THEN 'Invalid SIC'
        WHEN LENGTH(cm.mcc_code) != 4 AND cm.mcc_code IS NOT NULL AND cm.mcc_code != '' THEN 'Invalid MCC'
        ELSE 'Valid'
    END as validation_status
FROM crosswalk_mappings cm
JOIN industries i ON cm.industry_id = i.id
WHERE 
    (cm.naics_code IS NOT NULL AND cm.naics_code != '' AND NOT is_valid_naics_sector(cm.naics_code)) OR
    (cm.sic_code IS NOT NULL AND cm.sic_code != '' AND NOT is_valid_sic_division(cm.sic_code)) OR
    (LENGTH(cm.mcc_code) != 4 AND cm.mcc_code IS NOT NULL AND cm.mcc_code != '')
ORDER BY cm.industry_id, cm.id;

-- Add comments
COMMENT ON VIEW alignment_summary IS 'Summary of alignment status for each industry';
COMMENT ON VIEW alignment_issues IS 'Industries with alignment issues';
COMMENT ON VIEW hierarchy_validation IS 'Crosswalk mappings with hierarchy validation issues';
COMMENT ON FUNCTION is_valid_naics_sector(TEXT) IS 'Validates NAICS code sector (first 2 digits)';
COMMENT ON FUNCTION is_valid_sic_division(TEXT) IS 'Validates SIC code division (first digit)';

-- Grant permissions (adjust as needed for your environment)
-- GRANT SELECT ON alignment_summary TO your_app_user;
-- GRANT SELECT ON alignment_issues TO your_app_user;
-- GRANT SELECT ON hierarchy_validation TO your_app_user;
-- GRANT EXECUTE ON FUNCTION is_valid_naics_sector(TEXT) TO your_app_user;
-- GRANT EXECUTE ON FUNCTION is_valid_sic_division(TEXT) TO your_app_user;

