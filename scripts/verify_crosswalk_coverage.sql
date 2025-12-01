-- =====================================================
-- Crosswalk Coverage Verification
-- Purpose: Verify crosswalk coverage meets Phase 2 goals
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 2
-- =====================================================

-- =====================================================
-- Part 1: MCC Crosswalk Coverage
-- =====================================================

-- MCC codes with crosswalks
SELECT 
    'MCC Crosswalk Coverage' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata WHERE code_type = 'MCC' AND is_active = true), 2) AS percentage,
    CASE 
        WHEN COUNT(*) >= 30 THEN '✅ PASS - 30+ MCC codes (45%+)'
        ELSE '❌ FAIL - Below 30 MCC codes'
    END AS status
FROM code_metadata
WHERE code_type = 'MCC'
  AND is_active = true
  AND crosswalk_data != '{}'::jsonb
  AND (crosswalk_data ? 'naics' OR crosswalk_data ? 'sic');

-- MCC crosswalk distribution
SELECT 
    'MCC Crosswalk Distribution' AS metric,
    COUNT(*) FILTER (WHERE crosswalk_data ? 'naics') AS with_naics,
    COUNT(*) FILTER (WHERE crosswalk_data ? 'sic') AS with_sic,
    COUNT(*) FILTER (WHERE crosswalk_data ? 'naics' AND crosswalk_data ? 'sic') AS with_both
FROM code_metadata
WHERE code_type = 'MCC'
  AND is_active = true
  AND crosswalk_data != '{}'::jsonb;

-- =====================================================
-- Part 2: SIC Crosswalk Coverage
-- =====================================================

-- SIC codes with crosswalks
SELECT 
    'SIC Crosswalk Coverage' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata WHERE code_type = 'SIC' AND is_active = true), 2) AS percentage,
    CASE 
        WHEN COUNT(*) >= 20 THEN '✅ PASS - 20+ SIC codes (57%+)'
        ELSE '❌ FAIL - Below 20 SIC codes'
    END AS status
FROM code_metadata
WHERE code_type = 'SIC'
  AND is_active = true
  AND crosswalk_data != '{}'::jsonb
  AND (crosswalk_data ? 'naics' OR crosswalk_data ? 'mcc');

-- =====================================================
-- Part 3: NAICS Crosswalk Coverage
-- =====================================================

-- NAICS codes with crosswalks
SELECT 
    'NAICS Crosswalk Coverage' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata WHERE code_type = 'NAICS' AND is_active = true), 2) AS percentage,
    CASE 
        WHEN COUNT(*) >= 30 THEN '✅ PASS - 30+ NAICS codes (61%+)'
        ELSE '❌ FAIL - Below 30 NAICS codes'
    END AS status
FROM code_metadata
WHERE code_type = 'NAICS'
  AND is_active = true
  AND crosswalk_data != '{}'::jsonb
  AND (crosswalk_data ? 'sic' OR crosswalk_data ? 'mcc');

-- =====================================================
-- Part 4: Overall Crosswalk Coverage
-- =====================================================

-- Total codes with crosswalks
SELECT 
    'Total Codes with Crosswalks' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata WHERE is_active = true), 2) AS percentage,
    CASE 
        WHEN COUNT(*) >= 50 THEN '✅ PASS - 50+ codes (50%+)'
        ELSE '❌ FAIL - Below 50 codes'
    END AS status
FROM code_metadata
WHERE is_active = true
  AND crosswalk_data != '{}'::jsonb
  AND (
      crosswalk_data ? 'naics' 
      OR crosswalk_data ? 'sic' 
      OR crosswalk_data ? 'mcc'
  );

-- Crosswalk coverage by code type
SELECT 
    'Crosswalk Coverage by Code Type' AS metric,
    code_type,
    COUNT(*) FILTER (WHERE crosswalk_data != '{}'::jsonb) AS codes_with_crosswalks,
    COUNT(*) AS total_codes,
    ROUND(COUNT(*) FILTER (WHERE crosswalk_data != '{}'::jsonb) * 100.0 / COUNT(*), 2) AS coverage_percentage
FROM code_metadata
WHERE is_active = true
GROUP BY code_type
ORDER BY code_type;

-- =====================================================
-- Part 5: Crosswalk Accuracy Verification
-- =====================================================

-- Verify bidirectional crosswalks (if code A links to code B, code B should link to code A)
-- This is a simplified check - full verification would require more complex queries

-- Sample crosswalk validation
SELECT 
    'Sample Crosswalk Validation' AS example,
    cm1.code_type AS source_type,
    cm1.code AS source_code,
    cm1.official_name AS source_name,
    cm2.code_type AS target_type,
    cm2.code AS target_code,
    cm2.official_name AS target_name
FROM code_metadata cm1
CROSS JOIN LATERAL jsonb_array_elements_text(cm1.crosswalk_data->'naics') AS naics_code
INNER JOIN code_metadata cm2 
    ON cm2.code_type = 'NAICS' 
    AND cm2.code = naics_code
WHERE cm1.is_active = true
  AND cm2.is_active = true
LIMIT 10;

-- =====================================================
-- Part 6: Summary Report
-- =====================================================

SELECT 
    '=== CROSSWALK COVERAGE SUMMARY ===' AS section;

SELECT 
    'MCC Codes with Crosswalks' AS metric,
    COUNT(*) AS value,
    CASE 
        WHEN COUNT(*) >= 30 THEN '✅ PASS'
        ELSE '❌ FAIL'
    END AS status
FROM code_metadata
WHERE code_type = 'MCC'
  AND is_active = true
  AND crosswalk_data != '{}'::jsonb
  AND (crosswalk_data ? 'naics' OR crosswalk_data ? 'sic')

UNION ALL

SELECT 
    'SIC Codes with Crosswalks' AS metric,
    COUNT(*) AS value,
    CASE 
        WHEN COUNT(*) >= 20 THEN '✅ PASS'
        ELSE '❌ FAIL'
    END AS status
FROM code_metadata
WHERE code_type = 'SIC'
  AND is_active = true
  AND crosswalk_data != '{}'::jsonb
  AND (crosswalk_data ? 'naics' OR crosswalk_data ? 'mcc')

UNION ALL

SELECT 
    'NAICS Codes with Crosswalks' AS metric,
    COUNT(*) AS value,
    CASE 
        WHEN COUNT(*) >= 30 THEN '✅ PASS'
        ELSE '❌ FAIL'
    END AS status
FROM code_metadata
WHERE code_type = 'NAICS'
  AND is_active = true
  AND crosswalk_data != '{}'::jsonb
  AND (crosswalk_data ? 'sic' OR crosswalk_data ? 'mcc')

UNION ALL

SELECT 
    'Total Codes with Crosswalks' AS metric,
    COUNT(*) AS value,
    CASE 
        WHEN COUNT(*) >= 50 THEN '✅ PASS'
        ELSE '❌ FAIL'
    END AS status
FROM code_metadata
WHERE is_active = true
  AND crosswalk_data != '{}'::jsonb
  AND (
      crosswalk_data ? 'naics' 
      OR crosswalk_data ? 'sic' 
      OR crosswalk_data ? 'mcc'
  );

