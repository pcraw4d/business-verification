-- =====================================================
-- Verification Script: Check Crosswalk and Keyword Data
-- Purpose: Verify what data exists and what is missing
-- =====================================================

-- =====================================================
-- Part 1: Check Crosswalk Data for Common MCC Codes
-- =====================================================

SELECT 
    'Crosswalk Data Check' as check_type,
    code_type,
    code,
    official_name,
    CASE 
        WHEN crosswalk_data IS NULL OR crosswalk_data = '{}'::jsonb THEN '❌ Missing'
        ELSE '✅ Exists'
    END as crosswalk_status,
    crosswalk_data
FROM code_metadata
WHERE code_type = 'MCC' 
AND code IN ('5819', '5812', '5814', '5734', '6010', '8011', '5311')
ORDER BY code;

-- =====================================================
-- Part 2: Check Keyword Data for Technology Terms
-- =====================================================

SELECT 
    'Keyword Data Check' as check_type,
    i.name as industry_name,
    ik.keyword,
    ik.weight,
    ik.context
FROM industry_keywords ik
JOIN industries i ON i.id = ik.industry_id
WHERE LOWER(ik.keyword) IN ('cloud', 'computing', 'software', 'saas', 'platform', 'technology', 'tech', 'it services')
ORDER BY ik.weight DESC, i.name;

-- =====================================================
-- Part 3: Check if MCC 5819 has any crosswalk data
-- =====================================================

SELECT 
    'MCC 5819 Crosswalk Check' as check_type,
    code_type,
    code,
    official_name,
    crosswalk_data->'naics' as naics_codes,
    crosswalk_data->'sic' as sic_codes
FROM code_metadata
WHERE code_type = 'MCC' AND code = '5819';

-- =====================================================
-- Part 4: Count Missing Crosswalks
-- =====================================================

SELECT 
    'Missing Crosswalks Summary' as check_type,
    COUNT(*) as total_mcc_codes,
    COUNT(CASE WHEN crosswalk_data IS NULL OR crosswalk_data = '{}'::jsonb THEN 1 END) as missing_crosswalks,
    ROUND(100.0 * COUNT(CASE WHEN crosswalk_data IS NULL OR crosswalk_data = '{}'::jsonb THEN 1 END) / COUNT(*), 2) as missing_percentage
FROM code_metadata
WHERE code_type = 'MCC' AND is_active = true;

-- =====================================================
-- Part 5: Check Technology Keywords by Industry
-- =====================================================

SELECT 
    'Technology Keywords by Industry' as check_type,
    i.name as industry_name,
    COUNT(*) as keyword_count,
    STRING_AGG(ik.keyword, ', ' ORDER BY ik.weight DESC) as keywords
FROM industry_keywords ik
JOIN industries i ON i.id = ik.industry_id
WHERE LOWER(i.name) IN ('technology', 'software', 'cloud computing', 'it services', 'information technology', 'software development')
GROUP BY i.name
ORDER BY keyword_count DESC;
