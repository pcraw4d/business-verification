-- =====================================================
-- Crosswalk Data Quality Verification
-- Purpose: Verify crosswalk data doesn't have incorrect code type assignments
-- Date: 2025-01-27
-- =====================================================
-- 
-- This script checks for potential issues:
-- 1. MCC codes in SIC crosswalk arrays (or vice versa)
-- 2. Codes that appear in their own crosswalk arrays
-- 3. Invalid code type assignments
-- =====================================================

-- =====================================================
-- Part 1: Check for MCC codes incorrectly in SIC arrays
-- =====================================================

-- Find MCC codes that have SIC codes in their crosswalk that match MCC code values
SELECT 
    'MCC Codes with Matching SIC Codes in Crosswalk' AS issue_type,
    cm.code_type,
    cm.code,
    cm.official_name,
    sic_code_in_crosswalk,
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM code_metadata cm2 
            WHERE cm2.code_type = 'MCC' 
            AND cm2.code = sic_code_in_crosswalk
        ) THEN '⚠️ This SIC code also exists as an MCC code'
        ELSE '✅ OK'
    END AS status
FROM code_metadata cm
CROSS JOIN LATERAL jsonb_array_elements_text(cm.crosswalk_data->'sic') AS sic_code_in_crosswalk
WHERE cm.code_type = 'MCC'
  AND cm.is_active = true
  AND cm.crosswalk_data ? 'sic'
  AND EXISTS (
      SELECT 1 FROM code_metadata cm2 
      WHERE cm2.code_type = 'MCC' 
      AND cm2.code = sic_code_in_crosswalk
  )
LIMIT 20;

-- =====================================================
-- Part 2: Check for SIC codes incorrectly in MCC arrays
-- =====================================================

-- Find SIC codes that have MCC codes in their crosswalk that match SIC code values
SELECT 
    'SIC Codes with Matching MCC Codes in Crosswalk' AS issue_type,
    cm.code_type,
    cm.code,
    cm.official_name,
    mcc_code_in_crosswalk,
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM code_metadata cm2 
            WHERE cm2.code_type = 'SIC' 
            AND cm2.code = mcc_code_in_crosswalk
        ) THEN '⚠️ This MCC code also exists as a SIC code'
        ELSE '✅ OK'
    END AS status
FROM code_metadata cm
CROSS JOIN LATERAL jsonb_array_elements_text(cm.crosswalk_data->'mcc') AS mcc_code_in_crosswalk
WHERE cm.code_type = 'SIC'
  AND cm.is_active = true
  AND cm.crosswalk_data ? 'mcc'
  AND EXISTS (
      SELECT 1 FROM code_metadata cm2 
      WHERE cm2.code_type = 'SIC' 
      AND cm2.code = mcc_code_in_crosswalk
  )
LIMIT 20;

-- =====================================================
-- Part 3: Check for codes referencing themselves
-- =====================================================

-- Find codes that reference themselves in crosswalk arrays
SELECT 
    'Codes Referencing Themselves in Crosswalk' AS issue_type,
    cm.code_type,
    cm.code,
    cm.official_name,
    'mcc' AS crosswalk_array,
    cm.code AS self_reference
FROM code_metadata cm
WHERE cm.code_type = 'MCC'
  AND cm.is_active = true
  AND cm.crosswalk_data ? 'mcc'
  AND cm.code = ANY(
      SELECT jsonb_array_elements_text(cm.crosswalk_data->'mcc')
  )

UNION ALL

SELECT 
    'Codes Referencing Themselves in Crosswalk' AS issue_type,
    cm.code_type,
    cm.code,
    cm.official_name,
    'sic' AS crosswalk_array,
    cm.code AS self_reference
FROM code_metadata cm
WHERE cm.code_type = 'SIC'
  AND cm.is_active = true
  AND cm.crosswalk_data ? 'sic'
  AND cm.code = ANY(
      SELECT jsonb_array_elements_text(cm.crosswalk_data->'sic')
  )

UNION ALL

SELECT 
    'Codes Referencing Themselves in Crosswalk' AS issue_type,
    cm.code_type,
    cm.code,
    cm.official_name,
    'naics' AS crosswalk_array,
    cm.code AS self_reference
FROM code_metadata cm
WHERE cm.code_type = 'NAICS'
  AND cm.is_active = true
  AND cm.crosswalk_data ? 'naics'
  AND cm.code = ANY(
      SELECT jsonb_array_elements_text(cm.crosswalk_data->'naics')
  );

-- =====================================================
-- Part 4: Check for codes that exist in multiple code types
-- =====================================================

-- Find codes that exist as both MCC and SIC (same numeric value)
SELECT 
    'Codes Existing in Multiple Code Types' AS issue_type,
    mcc.code AS code_value,
    'MCC' AS first_type,
    mcc.official_name AS mcc_name,
    'SIC' AS second_type,
    sic.official_name AS sic_name,
    CASE 
        WHEN mcc.official_name = sic.official_name THEN '⚠️ Same name - likely duplicate'
        ELSE '⚠️ Different names - may be intentional'
    END AS status
FROM code_metadata mcc
INNER JOIN code_metadata sic 
    ON mcc.code = sic.code
WHERE mcc.code_type = 'MCC'
  AND sic.code_type = 'SIC'
  AND mcc.is_active = true
  AND sic.is_active = true
ORDER BY mcc.code
LIMIT 20;

-- =====================================================
-- Part 5: Verify crosswalk array contents are correct types
-- =====================================================

-- Check that 'mcc' arrays only contain codes that exist as MCC
SELECT 
    'Invalid MCC Codes in Crosswalk Arrays' AS issue_type,
    cm.code_type AS source_type,
    cm.code AS source_code,
    cm.official_name AS source_name,
    mcc_code_in_crosswalk AS invalid_mcc_code,
    CASE 
        WHEN NOT EXISTS (
            SELECT 1 FROM code_metadata cm2 
            WHERE cm2.code_type = 'MCC' 
            AND cm2.code = mcc_code_in_crosswalk
            AND cm2.is_active = true
        ) THEN '❌ MCC code does not exist'
        ELSE '✅ MCC code exists'
    END AS status
FROM code_metadata cm
CROSS JOIN LATERAL jsonb_array_elements_text(cm.crosswalk_data->'mcc') AS mcc_code_in_crosswalk
WHERE cm.is_active = true
  AND cm.crosswalk_data ? 'mcc'
  AND NOT EXISTS (
      SELECT 1 FROM code_metadata cm2 
      WHERE cm2.code_type = 'MCC' 
      AND cm2.code = mcc_code_in_crosswalk
      AND cm2.is_active = true
  )
LIMIT 20;

-- Check that 'sic' arrays only contain codes that exist as SIC
SELECT 
    'Invalid SIC Codes in Crosswalk Arrays' AS issue_type,
    cm.code_type AS source_type,
    cm.code AS source_code,
    cm.official_name AS source_name,
    sic_code_in_crosswalk AS invalid_sic_code,
    CASE 
        WHEN NOT EXISTS (
            SELECT 1 FROM code_metadata cm2 
            WHERE cm2.code_type = 'SIC' 
            AND cm2.code = sic_code_in_crosswalk
            AND cm2.is_active = true
        ) THEN '❌ SIC code does not exist'
        ELSE '✅ SIC code exists'
    END AS status
FROM code_metadata cm
CROSS JOIN LATERAL jsonb_array_elements_text(cm.crosswalk_data->'sic') AS sic_code_in_crosswalk
WHERE cm.is_active = true
  AND cm.crosswalk_data ? 'sic'
  AND NOT EXISTS (
      SELECT 1 FROM code_metadata cm2 
      WHERE cm2.code_type = 'SIC' 
      AND cm2.code = sic_code_in_crosswalk
      AND cm2.is_active = true
  )
LIMIT 20;

-- Check that 'naics' arrays only contain codes that exist as NAICS
SELECT 
    'Invalid NAICS Codes in Crosswalk Arrays' AS issue_type,
    cm.code_type AS source_type,
    cm.code AS source_code,
    cm.official_name AS source_name,
    naics_code_in_crosswalk AS invalid_naics_code,
    CASE 
        WHEN NOT EXISTS (
            SELECT 1 FROM code_metadata cm2 
            WHERE cm2.code_type = 'NAICS' 
            AND cm2.code = naics_code_in_crosswalk
            AND cm2.is_active = true
        ) THEN '❌ NAICS code does not exist'
        ELSE '✅ NAICS code exists'
    END AS status
FROM code_metadata cm
CROSS JOIN LATERAL jsonb_array_elements_text(cm.crosswalk_data->'naics') AS naics_code_in_crosswalk
WHERE cm.is_active = true
  AND cm.crosswalk_data ? 'naics'
  AND NOT EXISTS (
      SELECT 1 FROM code_metadata cm2 
      WHERE cm2.code_type = 'NAICS' 
      AND cm2.code = naics_code_in_crosswalk
      AND cm2.is_active = true
  )
LIMIT 20;

-- =====================================================
-- Part 6: Sample problematic crosswalks
-- =====================================================

-- Show examples where MCC and SIC have the same code value
SELECT 
    'Sample: Codes with Same Value in MCC and SIC' AS example,
    mcc.code AS code_value,
    mcc.official_name AS mcc_name,
    sic.official_name AS sic_name,
    mcc.crosswalk_data AS mcc_crosswalk,
    sic.crosswalk_data AS sic_crosswalk
FROM code_metadata mcc
INNER JOIN code_metadata sic 
    ON mcc.code = sic.code
WHERE mcc.code_type = 'MCC'
  AND sic.code_type = 'SIC'
  AND mcc.is_active = true
  AND sic.is_active = true
LIMIT 10;

