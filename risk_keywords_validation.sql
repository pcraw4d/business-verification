-- ====================================================================
-- Risk Keywords Data Validation Script
-- ====================================================================
-- This script validates the populated risk keywords data and tests
-- the functionality of the risk detection system.
-- ====================================================================

-- ====================================================================
-- 1. DATA INTEGRITY VALIDATION
-- ====================================================================

-- Check total record count
SELECT 
    'Total Risk Keywords' as metric,
    COUNT(*) as count
FROM risk_keywords;

-- Check active vs inactive records
SELECT 
    'Active Keywords' as metric,
    COUNT(*) as count
FROM risk_keywords 
WHERE is_active = true
UNION ALL
SELECT 
    'Inactive Keywords' as metric,
    COUNT(*) as count
FROM risk_keywords 
WHERE is_active = false;

-- Check distribution by risk category
SELECT 
    risk_category,
    COUNT(*) as keyword_count,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2) as percentage
FROM risk_keywords 
WHERE is_active = true
GROUP BY risk_category
ORDER BY keyword_count DESC;

-- Check distribution by risk severity
SELECT 
    risk_severity,
    COUNT(*) as keyword_count,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2) as percentage
FROM risk_keywords 
WHERE is_active = true
GROUP BY risk_severity
ORDER BY 
    CASE risk_severity 
        WHEN 'critical' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        WHEN 'low' THEN 4
    END;

-- ====================================================================
-- 2. CONTENT VALIDATION
-- ====================================================================

-- Check for empty or null keywords
SELECT 
    'Empty Keywords' as issue,
    COUNT(*) as count
FROM risk_keywords 
WHERE keyword IS NULL OR TRIM(keyword) = ''
UNION ALL
SELECT 
    'Missing Descriptions' as issue,
    COUNT(*) as count
FROM risk_keywords 
WHERE description IS NULL OR TRIM(description) = ''
UNION ALL
SELECT 
    'Missing MCC Codes' as issue,
    COUNT(*) as count
FROM risk_keywords 
WHERE mcc_codes IS NULL OR array_length(mcc_codes, 1) IS NULL;

-- Check for duplicate keywords
SELECT 
    keyword,
    COUNT(*) as duplicate_count
FROM risk_keywords 
WHERE is_active = true
GROUP BY keyword
HAVING COUNT(*) > 1
ORDER BY duplicate_count DESC;

-- ====================================================================
-- 3. MCC CODE VALIDATION
-- ====================================================================

-- Check MCC codes distribution
SELECT 
    'Keywords with MCC Codes' as metric,
    COUNT(*) as count
FROM risk_keywords 
WHERE mcc_codes IS NOT NULL AND array_length(mcc_codes, 1) > 0
UNION ALL
SELECT 
    'Keywords without MCC Codes' as metric,
    COUNT(*) as count
FROM risk_keywords 
WHERE mcc_codes IS NULL OR array_length(mcc_codes, 1) IS NULL;

-- List all unique MCC codes
SELECT DISTINCT unnest(mcc_codes) as mcc_code
FROM risk_keywords 
WHERE mcc_codes IS NOT NULL
ORDER BY mcc_code;

-- Check for specific prohibited MCC codes
SELECT 
    'MCC 7995 (Gambling)' as mcc_code,
    COUNT(*) as keyword_count
FROM risk_keywords 
WHERE '7995' = ANY(mcc_codes)
UNION ALL
SELECT 
    'MCC 7273 (Dating)' as mcc_code,
    COUNT(*) as keyword_count
FROM risk_keywords 
WHERE '7273' = ANY(mcc_codes)
UNION ALL
SELECT 
    'MCC 7841 (Video Entertainment)' as mcc_code,
    COUNT(*) as keyword_count
FROM risk_keywords 
WHERE '7841' = ANY(mcc_codes)
UNION ALL
SELECT 
    'MCC 5993 (Cigar Stores)' as mcc_code,
    COUNT(*) as keyword_count
FROM risk_keywords 
WHERE '5993' = ANY(mcc_codes)
UNION ALL
SELECT 
    'MCC 5921 (Package Stores)' as mcc_code,
    COUNT(*) as keyword_count
FROM risk_keywords 
WHERE '5921' = ANY(mcc_codes);

-- ====================================================================
-- 4. CARD BRAND RESTRICTIONS VALIDATION
-- ====================================================================

-- Check card brand restrictions distribution
SELECT 
    'Keywords with Card Restrictions' as metric,
    COUNT(*) as count
FROM risk_keywords 
WHERE card_brand_restrictions IS NOT NULL AND array_length(card_brand_restrictions, 1) > 0
UNION ALL
SELECT 
    'Keywords without Card Restrictions' as metric,
    COUNT(*) as count
FROM risk_keywords 
WHERE card_brand_restrictions IS NULL OR array_length(card_brand_restrictions, 1) IS NULL;

-- List all unique card brand restrictions
SELECT DISTINCT unnest(card_brand_restrictions) as card_brand
FROM risk_keywords 
WHERE card_brand_restrictions IS NOT NULL
ORDER BY card_brand;

-- Check restrictions by card brand
SELECT 
    unnest(card_brand_restrictions) as card_brand,
    COUNT(*) as restriction_count
FROM risk_keywords 
WHERE card_brand_restrictions IS NOT NULL
GROUP BY unnest(card_brand_restrictions)
ORDER BY restriction_count DESC;

-- ====================================================================
-- 5. DETECTION PATTERNS VALIDATION
-- ====================================================================

-- Check detection patterns distribution
SELECT 
    'Keywords with Detection Patterns' as metric,
    COUNT(*) as count
FROM risk_keywords 
WHERE detection_patterns IS NOT NULL AND array_length(detection_patterns, 1) > 0
UNION ALL
SELECT 
    'Keywords without Detection Patterns' as metric,
    COUNT(*) as count
FROM risk_keywords 
WHERE detection_patterns IS NULL OR array_length(detection_patterns, 1) IS NULL;

-- Check for valid regex patterns (basic validation)
SELECT 
    keyword,
    detection_patterns
FROM risk_keywords 
WHERE detection_patterns IS NOT NULL
    AND array_length(detection_patterns, 1) > 0
    AND (
        -- Check for basic regex structure
        EXISTS (
            SELECT 1 FROM unnest(detection_patterns) as pattern 
            WHERE pattern LIKE '%(?i)%' OR pattern LIKE '%[%]%' OR pattern LIKE '%(%'
        )
    )
LIMIT 10;

-- ====================================================================
-- 6. SYNONYMS VALIDATION
-- ====================================================================

-- Check synonyms distribution
SELECT 
    'Keywords with Synonyms' as metric,
    COUNT(*) as count
FROM risk_keywords 
WHERE synonyms IS NOT NULL AND array_length(synonyms, 1) > 0
UNION ALL
SELECT 
    'Keywords without Synonyms' as metric,
    COUNT(*) as count
FROM risk_keywords 
WHERE synonyms IS NULL OR array_length(synonyms, 1) IS NULL;

-- Show examples of keywords with synonyms
SELECT 
    keyword,
    synonyms,
    array_length(synonyms, 1) as synonym_count
FROM risk_keywords 
WHERE synonyms IS NOT NULL 
    AND array_length(synonyms, 1) > 0
ORDER BY array_length(synonyms, 1) DESC
LIMIT 10;

-- ====================================================================
-- 7. RISK ASSESSMENT TESTING
-- ====================================================================

-- Test risk keyword matching function (if exists)
-- This would test the actual risk detection functionality
SELECT 
    'Risk Detection Test' as test_name,
    'Testing keyword matching functionality' as description;

-- Sample test cases for risk detection
WITH test_cases AS (
    SELECT 'drug trafficking business' as test_text
    UNION ALL SELECT 'online casino gambling'
    UNION ALL SELECT 'adult entertainment club'
    UNION ALL SELECT 'bitcoin cryptocurrency exchange'
    UNION ALL SELECT 'tobacco cigarette store'
    UNION ALL SELECT 'money laundering service'
    UNION ALL SELECT 'shell company front'
    UNION ALL SELECT 'human trafficking network'
)
SELECT 
    tc.test_text,
    rk.keyword,
    rk.risk_category,
    rk.risk_severity,
    rk.description
FROM test_cases tc
CROSS JOIN risk_keywords rk
WHERE rk.is_active = true
    AND (
        -- Simple keyword matching (case insensitive)
        LOWER(tc.test_text) LIKE '%' || LOWER(rk.keyword) || '%'
        OR 
        -- Synonym matching
        EXISTS (
            SELECT 1 FROM unnest(rk.synonyms) as synonym
            WHERE LOWER(tc.test_text) LIKE '%' || LOWER(synonym) || '%'
        )
    )
ORDER BY tc.test_text, 
    CASE rk.risk_severity 
        WHEN 'critical' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        WHEN 'low' THEN 4
    END;

-- ====================================================================
-- 8. PERFORMANCE VALIDATION
-- ====================================================================

-- Test index performance
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM risk_keywords 
WHERE risk_category = 'illegal' 
    AND risk_severity = 'critical'
    AND is_active = true;

-- Test keyword search performance
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM risk_keywords 
WHERE to_tsvector('english', keyword) @@ to_tsquery('english', 'drug | gambling | adult');

-- Test MCC code lookup performance
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM risk_keywords 
WHERE '7995' = ANY(mcc_codes);

-- ====================================================================
-- 9. COMPREHENSIVE SUMMARY REPORT
-- ====================================================================

-- Generate comprehensive summary report
SELECT 
    '=== RISK KEYWORDS DATA VALIDATION SUMMARY ===' as report_section
UNION ALL
SELECT 
    'Total Keywords: ' || COUNT(*)::text
FROM risk_keywords
UNION ALL
SELECT 
    'Active Keywords: ' || COUNT(*)::text
FROM risk_keywords 
WHERE is_active = true
UNION ALL
SELECT 
    'Critical Risk Keywords: ' || COUNT(*)::text
FROM risk_keywords 
WHERE risk_severity = 'critical' AND is_active = true
UNION ALL
SELECT 
    'High Risk Keywords: ' || COUNT(*)::text
FROM risk_keywords 
WHERE risk_severity = 'high' AND is_active = true
UNION ALL
SELECT 
    'Keywords with MCC Codes: ' || COUNT(*)::text
FROM risk_keywords 
WHERE mcc_codes IS NOT NULL AND array_length(mcc_codes, 1) > 0
UNION ALL
SELECT 
    'Keywords with Card Restrictions: ' || COUNT(*)::text
FROM risk_keywords 
WHERE card_brand_restrictions IS NOT NULL AND array_length(card_brand_restrictions, 1) > 0
UNION ALL
SELECT 
    'Keywords with Detection Patterns: ' || COUNT(*)::text
FROM risk_keywords 
WHERE detection_patterns IS NOT NULL AND array_length(detection_patterns, 1) > 0
UNION ALL
SELECT 
    'Keywords with Synonyms: ' || COUNT(*)::text
FROM risk_keywords 
WHERE synonyms IS NOT NULL AND array_length(synonyms, 1) > 0
UNION ALL
SELECT 
    '=== VALIDATION COMPLETE ===' as report_section;

-- ====================================================================
-- END OF VALIDATION SCRIPT
-- ====================================================================
