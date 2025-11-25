-- =====================================================
-- NAICS Alignment Migration Verification Script
-- =====================================================
-- This script verifies that migrations 023, 024, and 025
-- completed successfully and the data is correct.
--
-- Run this after completing all three migrations:
--   023_add_naics_aligned_industries.sql
--   024_reclassify_codes_to_naics_industries.sql
--   025_update_code_keywords_for_naics_industries.sql
-- =====================================================

-- =====================================================
-- 1. Verify New Industries Were Created
-- =====================================================

SELECT 
    '✅ New NAICS Industries Created' as check_name,
    COUNT(*) as count,
    CASE 
        WHEN COUNT(*) >= 11 THEN 'PASS'
        ELSE 'FAIL - Expected at least 11 new industries'
    END as status
FROM industries
WHERE name IN (
    'Agriculture, Forestry, Fishing and Hunting',
    'Mining, Quarrying, and Oil and Gas Extraction',
    'Utilities',
    'Wholesale Trade',
    'Real Estate and Rental and Leasing',
    'Professional, Scientific, and Technical Services',
    'Management of Companies and Enterprises',
    'Administrative and Support Services',
    'Arts, Entertainment, and Recreation',
    'Other Services',
    'Public Administration'
);

-- List all new industries
SELECT 
    'New Industries List' as check_name,
    name as industry_name,
    description,
    is_active
FROM industries
WHERE name IN (
    'Agriculture, Forestry, Fishing and Hunting',
    'Mining, Quarrying, and Oil and Gas Extraction',
    'Utilities',
    'Wholesale Trade',
    'Real Estate and Rental and Leasing',
    'Professional, Scientific, and Technical Services',
    'Management of Companies and Enterprises',
    'Administrative and Support Services',
    'Arts, Entertainment, and Recreation',
    'Other Services',
    'Public Administration'
)
ORDER BY name;

-- =====================================================
-- 2. Verify No NULL industry_id Values
-- =====================================================

SELECT 
    '✅ No NULL industry_id Values' as check_name,
    COUNT(*) as null_count,
    CASE 
        WHEN COUNT(*) = 0 THEN 'PASS'
        ELSE 'FAIL - Found ' || COUNT(*) || ' codes with NULL industry_id'
    END as status
FROM classification_codes
WHERE industry_id IS NULL;

-- If there are NULL values, show them
SELECT 
    'NULL industry_id Codes (if any)' as check_name,
    id,
    code_type,
    code,
    LEFT(description, 50) as description_preview
FROM classification_codes
WHERE industry_id IS NULL
LIMIT 10;

-- =====================================================
-- 3. Verify Code Reclassification
-- =====================================================

-- Check codes reclassified to new NAICS industries
SELECT 
    '✅ Codes Reclassified to NAICS Industries' as check_name,
    COUNT(*) as total_codes,
    COUNT(DISTINCT industry_id) as industries_with_codes,
    CASE 
        WHEN COUNT(*) > 0 THEN 'PASS'
        ELSE 'FAIL - No codes found in new industries'
    END as status
FROM classification_codes cc
INNER JOIN industries i ON i.id = cc.industry_id
WHERE i.name IN (
    'Agriculture, Forestry, Fishing and Hunting',
    'Mining, Quarrying, and Oil and Gas Extraction',
    'Utilities',
    'Wholesale Trade',
    'Real Estate and Rental and Leasing',
    'Professional, Scientific, and Technical Services',
    'Management of Companies and Enterprises',
    'Administrative and Support Services',
    'Arts, Entertainment, and Recreation',
    'Other Services',
    'Public Administration'
);

-- Breakdown by industry
SELECT 
    'Code Distribution by NAICS Industry' as check_name,
    i.name as industry,
    COUNT(cc.id) as code_count,
    COUNT(DISTINCT cc.code_type) as code_types,
    STRING_AGG(DISTINCT cc.code_type, ', ') as types_list
FROM classification_codes cc
INNER JOIN industries i ON i.id = cc.industry_id
WHERE i.name IN (
    'Agriculture, Forestry, Fishing and Hunting',
    'Mining, Quarrying, and Oil and Gas Extraction',
    'Utilities',
    'Wholesale Trade',
    'Real Estate and Rental and Leasing',
    'Professional, Scientific, and Technical Services',
    'Management of Companies and Enterprises',
    'Administrative and Support Services',
    'Arts, Entertainment, and Recreation',
    'Other Services',
    'Public Administration'
)
GROUP BY i.name
ORDER BY code_count DESC;

-- =====================================================
-- 4. Verify NAICS Code Prefix Matching
-- =====================================================

-- Check that NAICS codes starting with specific prefixes are in correct industries
SELECT 
    '✅ NAICS Code Prefix Matching' as check_name,
    LEFT(cc.code, 2) as naics_prefix,
    i.name as expected_industry,
    COUNT(*) as code_count,
    CASE 
        WHEN COUNT(*) > 0 THEN 'PASS'
        ELSE 'WARNING - No codes found for this prefix'
    END as status
FROM classification_codes cc
INNER JOIN industries i ON i.id = cc.industry_id
WHERE cc.code_type = 'NAICS'
  AND (
    (LEFT(cc.code, 2) = '11' AND i.name = 'Agriculture, Forestry, Fishing and Hunting') OR
    (LEFT(cc.code, 2) = '21' AND i.name = 'Mining, Quarrying, and Oil and Gas Extraction') OR
    (LEFT(cc.code, 2) = '22' AND i.name = 'Utilities') OR
    (LEFT(cc.code, 2) = '23' AND i.name = 'Construction') OR
    (LEFT(cc.code, 2) IN ('31', '32', '33') AND i.name = 'Manufacturing') OR
    (LEFT(cc.code, 2) = '42' AND i.name = 'Wholesale Trade') OR
    (LEFT(cc.code, 2) IN ('44', '45') AND i.name = 'Retail') OR
    (LEFT(cc.code, 2) IN ('48', '49') AND i.name = 'Transportation') OR
    (LEFT(cc.code, 2) = '51' AND i.name = 'Technology') OR
    (LEFT(cc.code, 2) = '52' AND i.name = 'Finance') OR
    (LEFT(cc.code, 2) = '53' AND i.name = 'Real Estate and Rental and Leasing') OR
    (LEFT(cc.code, 2) = '54' AND i.name = 'Professional, Scientific, and Technical Services') OR
    (LEFT(cc.code, 2) = '55' AND i.name = 'Management of Companies and Enterprises') OR
    (LEFT(cc.code, 2) = '56' AND i.name = 'Administrative and Support Services') OR
    (LEFT(cc.code, 2) = '61' AND i.name = 'Education') OR
    (LEFT(cc.code, 2) = '62' AND i.name = 'Healthcare') OR
    (LEFT(cc.code, 2) = '71' AND i.name = 'Arts, Entertainment, and Recreation') OR
    (LEFT(cc.code, 2) = '72' AND i.name = 'Food & Beverage') OR
    (LEFT(cc.code, 2) = '81' AND i.name = 'Other Services') OR
    (LEFT(cc.code, 2) = '92' AND i.name = 'Public Administration')
  )
GROUP BY LEFT(cc.code, 2), i.name
ORDER BY LEFT(cc.code, 2);

-- =====================================================
-- 5. Verify Remaining Codes in General Business
-- =====================================================

SELECT 
    '✅ Remaining Codes in General Business' as check_name,
    COUNT(*) as remaining_count,
    COUNT(DISTINCT code_type) as code_types,
    STRING_AGG(DISTINCT code_type, ', ') as types_list,
    CASE 
        WHEN COUNT(*) < 100 THEN 'PASS - Most codes reclassified'
        WHEN COUNT(*) < 500 THEN 'WARNING - Some codes remain'
        ELSE 'FAIL - Too many codes still in General Business'
    END as status
FROM classification_codes
WHERE industry_id = (SELECT id FROM industries WHERE name = 'General Business');

-- Sample of remaining codes
SELECT 
    'Sample Remaining General Business Codes' as check_name,
    code_type,
    code,
    LEFT(description, 60) as description_preview
FROM classification_codes
WHERE industry_id = (SELECT id FROM industries WHERE name = 'General Business')
ORDER BY code_type, code
LIMIT 20;

-- =====================================================
-- 6. Verify Industry Keywords Were Added
-- =====================================================

SELECT 
    '✅ Industry Keywords Added' as check_name,
    COUNT(*) as total_keywords,
    COUNT(DISTINCT industry_id) as industries_with_keywords,
    CASE 
        WHEN COUNT(*) >= 100 THEN 'PASS'
        ELSE 'WARNING - Expected more keywords'
    END as status
FROM industry_keywords ik
INNER JOIN industries i ON i.id = ik.industry_id
WHERE i.name IN (
    'Agriculture, Forestry, Fishing and Hunting',
    'Mining, Quarrying, and Oil and Gas Extraction',
    'Utilities',
    'Wholesale Trade',
    'Real Estate and Rental and Leasing',
    'Professional, Scientific, and Technical Services',
    'Management of Companies and Enterprises',
    'Administrative and Support Services',
    'Arts, Entertainment, and Recreation',
    'Other Services',
    'Public Administration'
);

-- Keywords per industry
SELECT 
    'Keywords per NAICS Industry' as check_name,
    i.name as industry,
    COUNT(ik.id) as keyword_count,
    ROUND(AVG(ik.weight), 2) as avg_weight
FROM industry_keywords ik
INNER JOIN industries i ON i.id = ik.industry_id
WHERE i.name IN (
    'Agriculture, Forestry, Fishing and Hunting',
    'Mining, Quarrying, and Oil and Gas Extraction',
    'Utilities',
    'Wholesale Trade',
    'Real Estate and Rental and Leasing',
    'Professional, Scientific, and Technical Services',
    'Management of Companies and Enterprises',
    'Administrative and Support Services',
    'Arts, Entertainment, and Recreation',
    'Other Services',
    'Public Administration'
)
GROUP BY i.name
ORDER BY keyword_count DESC;

-- =====================================================
-- 7. Verify Code Keywords Were Updated
-- =====================================================

SELECT 
    '✅ Code Keywords Updated for NAICS Industries' as check_name,
    COUNT(*) as total_keywords,
    COUNT(DISTINCT code_id) as codes_with_keywords,
    ROUND(COUNT(*)::NUMERIC / NULLIF(COUNT(DISTINCT code_id), 0), 2) as avg_keywords_per_code,
    CASE 
        WHEN COUNT(*) > 0 AND COUNT(DISTINCT code_id) > 0 THEN 'PASS'
        ELSE 'FAIL - No keywords found'
    END as status
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
INNER JOIN industries i ON i.id = cc.industry_id
WHERE i.name IN (
    'Agriculture, Forestry, Fishing and Hunting',
    'Mining, Quarrying, and Oil and Gas Extraction',
    'Utilities',
    'Wholesale Trade',
    'Real Estate and Rental and Leasing',
    'Professional, Scientific, and Technical Services',
    'Management of Companies and Enterprises',
    'Administrative and Support Services',
    'Arts, Entertainment, and Recreation',
    'Other Services',
    'Public Administration'
);

-- Code keywords breakdown by industry
SELECT 
    'Code Keywords by NAICS Industry' as check_name,
    i.name as industry,
    COUNT(DISTINCT ck.code_id) as codes_with_keywords,
    COUNT(ck.id) as total_keywords,
    ROUND(AVG(ck.relevance_score), 2) as avg_relevance,
    COUNT(DISTINCT CASE WHEN ck.match_type = 'exact' THEN ck.id END) as exact_matches,
    COUNT(DISTINCT CASE WHEN ck.match_type = 'partial' THEN ck.id END) as partial_matches
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
INNER JOIN industries i ON i.id = cc.industry_id
WHERE i.name IN (
    'Agriculture, Forestry, Fishing and Hunting',
    'Mining, Quarrying, and Oil and Gas Extraction',
    'Utilities',
    'Wholesale Trade',
    'Real Estate and Rental and Leasing',
    'Professional, Scientific, and Technical Services',
    'Management of Companies and Enterprises',
    'Administrative and Support Services',
    'Arts, Entertainment, and Recreation',
    'Other Services',
    'Public Administration'
)
GROUP BY i.name
ORDER BY i.name;

-- =====================================================
-- 8. Verify Data Integrity
-- =====================================================

-- Check for orphaned codes (codes without valid industry)
SELECT 
    '✅ No Orphaned Codes' as check_name,
    COUNT(*) as orphaned_count,
    CASE 
        WHEN COUNT(*) = 0 THEN 'PASS'
        ELSE 'FAIL - Found ' || COUNT(*) || ' orphaned codes'
    END as status
FROM classification_codes cc
LEFT JOIN industries i ON i.id = cc.industry_id
WHERE cc.industry_id IS NOT NULL
  AND i.id IS NULL;

-- Check for codes with invalid industry_id
SELECT 
    '✅ All Codes Have Valid Industry' as check_name,
    COUNT(*) as invalid_count,
    CASE 
        WHEN COUNT(*) = 0 THEN 'PASS'
        ELSE 'FAIL - Found ' || COUNT(*) || ' codes with invalid industry_id'
    END as status
FROM classification_codes cc
WHERE industry_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM industries i WHERE i.id = cc.industry_id);

-- =====================================================
-- 9. Overall Summary
-- =====================================================

SELECT 
    '📊 OVERALL MIGRATION SUMMARY' as summary_section,
    '' as detail;

SELECT 
    'Total Industries' as metric,
    COUNT(*)::TEXT as value
FROM industries;

SELECT 
    'NAICS-Aligned Industries' as metric,
    COUNT(*)::TEXT as value
FROM industries
WHERE name IN (
    'Agriculture, Forestry, Fishing and Hunting',
    'Mining, Quarrying, and Oil and Gas Extraction',
    'Utilities',
    'Wholesale Trade',
    'Real Estate and Rental and Leasing',
    'Professional, Scientific, and Technical Services',
    'Management of Companies and Enterprises',
    'Administrative and Support Services',
    'Arts, Entertainment, and Recreation',
    'Other Services',
    'Public Administration'
);

SELECT 
    'Total Classification Codes' as metric,
    COUNT(*)::TEXT as value
FROM classification_codes;

SELECT 
    'Codes in NAICS Industries' as metric,
    COUNT(*)::TEXT as value
FROM classification_codes cc
INNER JOIN industries i ON i.id = cc.industry_id
WHERE i.name IN (
    'Agriculture, Forestry, Fishing and Hunting',
    'Mining, Quarrying, and Oil and Gas Extraction',
    'Utilities',
    'Wholesale Trade',
    'Real Estate and Rental and Leasing',
    'Professional, Scientific, and Technical Services',
    'Management of Companies and Enterprises',
    'Administrative and Support Services',
    'Arts, Entertainment, and Recreation',
    'Other Services',
    'Public Administration'
);

SELECT 
    'Codes in General Business' as metric,
    COUNT(*)::TEXT as value
FROM classification_codes
WHERE industry_id = (SELECT id FROM industries WHERE name = 'General Business');

SELECT 
    'Total Code Keywords' as metric,
    COUNT(*)::TEXT as value
FROM code_keywords;

SELECT 
    'Code Keywords for NAICS Industries' as metric,
    COUNT(*)::TEXT as value
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
INNER JOIN industries i ON i.id = cc.industry_id
WHERE i.name IN (
    'Agriculture, Forestry, Fishing and Hunting',
    'Mining, Quarrying, and Oil and Gas Extraction',
    'Utilities',
    'Wholesale Trade',
    'Real Estate and Rental and Leasing',
    'Professional, Scientific, and Technical Services',
    'Management of Companies and Enterprises',
    'Administrative and Support Services',
    'Arts, Entertainment, and Recreation',
    'Other Services',
    'Public Administration'
);

-- =====================================================
-- 10. Sample Data Verification
-- =====================================================

-- Show sample codes from each new industry
SELECT 
    'Sample Codes by NAICS Industry' as check_name,
    i.name as industry,
    cc.code_type,
    cc.code,
    LEFT(cc.description, 50) as description_preview
FROM classification_codes cc
INNER JOIN industries i ON i.id = cc.industry_id
WHERE i.name IN (
    'Agriculture, Forestry, Fishing and Hunting',
    'Mining, Quarrying, and Oil and Gas Extraction',
    'Utilities',
    'Wholesale Trade',
    'Real Estate and Rental and Leasing',
    'Professional, Scientific, and Technical Services',
    'Management of Companies and Enterprises',
    'Administrative and Support Services',
    'Arts, Entertainment, and Recreation',
    'Other Services',
    'Public Administration'
)
ORDER BY i.name, cc.code_type, cc.code
LIMIT 50;

-- Show sample keywords for a few codes
SELECT 
    'Sample Code Keywords' as check_name,
    cc.code_type,
    cc.code,
    LEFT(cc.description, 40) as description_preview,
    ck.keyword,
    ck.relevance_score,
    ck.match_type
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
INNER JOIN industries i ON i.id = cc.industry_id
WHERE i.name IN (
    'Professional, Scientific, and Technical Services',
    'Real Estate and Rental and Leasing',
    'Wholesale Trade'
)
ORDER BY cc.code_type, cc.code, ck.relevance_score DESC
LIMIT 30;

-- =====================================================
-- VERIFICATION COMPLETE
-- =====================================================
-- Review all results above. All checks should show:
--   ✅ PASS status for critical checks
--   ✅ Reasonable counts and distributions
--   ✅ No NULL or orphaned data
-- =====================================================

