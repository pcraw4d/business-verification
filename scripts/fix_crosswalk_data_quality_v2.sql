-- =====================================================
-- Fix Crosswalk Data Quality Issues (Version 2 - Simplified)
-- Purpose: Remove invalid codes and self-references from crosswalk arrays
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Data Quality Fix
-- =====================================================
-- 
-- This script fixes:
-- 1. Removes invalid codes from crosswalk arrays (codes that don't exist)
-- 2. Removes self-references (codes referencing themselves)
-- =====================================================

-- =====================================================
-- Part 1: Clean NAICS arrays - Remove invalid codes and self-references
-- =====================================================

-- Create temporary function to clean arrays
CREATE OR REPLACE FUNCTION clean_naics_array(p_code_type VARCHAR, p_code VARCHAR, p_array JSONB)
RETURNS JSONB AS $$
    SELECT COALESCE(
        jsonb_agg(elem ORDER BY elem),
        '[]'::jsonb
    )
    FROM jsonb_array_elements_text(p_array) AS elem
    WHERE EXISTS (
        SELECT 1 FROM code_metadata cm2 
        WHERE cm2.code_type = 'NAICS' 
        AND cm2.code = elem
        AND cm2.is_active = true
    )
    AND elem != p_code;
$$ LANGUAGE SQL;

CREATE OR REPLACE FUNCTION clean_sic_array(p_code_type VARCHAR, p_code VARCHAR, p_array JSONB)
RETURNS JSONB AS $$
    SELECT COALESCE(
        jsonb_agg(elem ORDER BY elem),
        '[]'::jsonb
    )
    FROM jsonb_array_elements_text(p_array) AS elem
    WHERE EXISTS (
        SELECT 1 FROM code_metadata cm2 
        WHERE cm2.code_type = 'SIC' 
        AND cm2.code = elem
        AND cm2.is_active = true
    )
    AND elem != p_code;
$$ LANGUAGE SQL;

CREATE OR REPLACE FUNCTION clean_mcc_array(p_code_type VARCHAR, p_code VARCHAR, p_array JSONB)
RETURNS JSONB AS $$
    SELECT COALESCE(
        jsonb_agg(elem ORDER BY elem),
        '[]'::jsonb
    )
    FROM jsonb_array_elements_text(p_array) AS elem
    WHERE EXISTS (
        SELECT 1 FROM code_metadata cm2 
        WHERE cm2.code_type = 'MCC' 
        AND cm2.code = elem
        AND cm2.is_active = true
    )
    AND elem != p_code;
$$ LANGUAGE SQL;

-- Update NAICS arrays
UPDATE code_metadata cm
SET crosswalk_data = 
    CASE 
        WHEN crosswalk_data ? 'naics' THEN
            jsonb_set(
                crosswalk_data,
                '{naics}',
                clean_naics_array(cm.code_type, cm.code, crosswalk_data->'naics')
            )
        ELSE crosswalk_data
    END,
    updated_at = NOW()
WHERE is_active = true
  AND crosswalk_data ? 'naics';

-- Update SIC arrays
UPDATE code_metadata cm
SET crosswalk_data = 
    CASE 
        WHEN crosswalk_data ? 'sic' THEN
            jsonb_set(
                crosswalk_data,
                '{sic}',
                clean_sic_array(cm.code_type, cm.code, crosswalk_data->'sic')
            )
        ELSE crosswalk_data
    END,
    updated_at = NOW()
WHERE is_active = true
  AND crosswalk_data ? 'sic';

-- Update MCC arrays
UPDATE code_metadata cm
SET crosswalk_data = 
    CASE 
        WHEN crosswalk_data ? 'mcc' THEN
            jsonb_set(
                crosswalk_data,
                '{mcc}',
                clean_mcc_array(cm.code_type, cm.code, crosswalk_data->'mcc')
            )
        ELSE crosswalk_data
    END,
    updated_at = NOW()
WHERE is_active = true
  AND crosswalk_data ? 'mcc';

-- =====================================================
-- Part 2: Remove empty arrays from crosswalk_data
-- =====================================================

-- Remove empty NAICS arrays
UPDATE code_metadata
SET crosswalk_data = crosswalk_data - 'naics',
    updated_at = NOW()
WHERE is_active = true
  AND crosswalk_data ? 'naics'
  AND (crosswalk_data->'naics' = '[]'::jsonb OR jsonb_array_length(crosswalk_data->'naics') = 0);

-- Remove empty SIC arrays
UPDATE code_metadata
SET crosswalk_data = crosswalk_data - 'sic',
    updated_at = NOW()
WHERE is_active = true
  AND crosswalk_data ? 'sic'
  AND (crosswalk_data->'sic' = '[]'::jsonb OR jsonb_array_length(crosswalk_data->'sic') = 0);

-- Remove empty MCC arrays
UPDATE code_metadata
SET crosswalk_data = crosswalk_data - 'mcc',
    updated_at = NOW()
WHERE is_active = true
  AND crosswalk_data ? 'mcc'
  AND (crosswalk_data->'mcc' = '[]'::jsonb OR jsonb_array_length(crosswalk_data->'mcc') = 0);

-- =====================================================
-- Part 3: Set crosswalk_data to empty if all arrays are removed
-- =====================================================

UPDATE code_metadata
SET crosswalk_data = '{}'::jsonb,
    updated_at = NOW()
WHERE is_active = true
  AND crosswalk_data != '{}'::jsonb
  AND NOT (crosswalk_data ? 'naics')
  AND NOT (crosswalk_data ? 'sic')
  AND NOT (crosswalk_data ? 'mcc');

-- =====================================================
-- Part 4: Clean up temporary functions
-- =====================================================

DROP FUNCTION IF EXISTS clean_naics_array(VARCHAR, VARCHAR, JSONB);
DROP FUNCTION IF EXISTS clean_sic_array(VARCHAR, VARCHAR, JSONB);
DROP FUNCTION IF EXISTS clean_mcc_array(VARCHAR, VARCHAR, JSONB);

-- =====================================================
-- Verification Queries
-- =====================================================

-- Check remaining invalid NAICS codes
SELECT 
    'Remaining Invalid NAICS Codes' AS metric,
    COUNT(*) AS invalid_count
FROM code_metadata cm
CROSS JOIN LATERAL jsonb_array_elements_text(cm.crosswalk_data->'naics') AS naics_code
WHERE cm.is_active = true
  AND cm.crosswalk_data ? 'naics'
  AND NOT EXISTS (
      SELECT 1 FROM code_metadata cm2 
      WHERE cm2.code_type = 'NAICS' 
      AND cm2.code = naics_code
      AND cm2.is_active = true
  );

-- Check remaining invalid SIC codes
SELECT 
    'Remaining Invalid SIC Codes' AS metric,
    COUNT(*) AS invalid_count
FROM code_metadata cm
CROSS JOIN LATERAL jsonb_array_elements_text(cm.crosswalk_data->'sic') AS sic_code
WHERE cm.is_active = true
  AND cm.crosswalk_data ? 'sic'
  AND NOT EXISTS (
      SELECT 1 FROM code_metadata cm2 
      WHERE cm2.code_type = 'SIC' 
      AND cm2.code = sic_code
      AND cm2.is_active = true
  );

-- Check remaining invalid MCC codes
SELECT 
    'Remaining Invalid MCC Codes' AS metric,
    COUNT(*) AS invalid_count
FROM code_metadata cm
CROSS JOIN LATERAL jsonb_array_elements_text(cm.crosswalk_data->'mcc') AS mcc_code
WHERE cm.is_active = true
  AND cm.crosswalk_data ? 'mcc'
  AND NOT EXISTS (
      SELECT 1 FROM code_metadata cm2 
      WHERE cm2.code_type = 'MCC' 
      AND cm2.code = mcc_code
      AND cm2.is_active = true
  );

-- Check remaining self-references
SELECT 
    'Remaining Self-References' AS metric,
    COUNT(DISTINCT cm.code_type || ':' || cm.code) AS self_ref_count
FROM code_metadata cm
WHERE cm.is_active = true
  AND cm.crosswalk_data != '{}'::jsonb
  AND (
      (cm.crosswalk_data ? 'mcc' AND cm.code = ANY(
          SELECT jsonb_array_elements_text(cm.crosswalk_data->'mcc')
      ))
      OR (cm.crosswalk_data ? 'sic' AND cm.code = ANY(
          SELECT jsonb_array_elements_text(cm.crosswalk_data->'sic')
      ))
      OR (cm.crosswalk_data ? 'naics' AND cm.code = ANY(
          SELECT jsonb_array_elements_text(cm.crosswalk_data->'naics')
      ))
  );

-- Summary
SELECT 
    'Crosswalk Data Quality Fix Summary' AS metric,
    (SELECT COUNT(*) FROM code_metadata WHERE is_active = true AND crosswalk_data != '{}'::jsonb) AS codes_with_crosswalks,
    (SELECT COUNT(*) FROM code_metadata WHERE is_active = true) AS total_codes,
    ROUND((SELECT COUNT(*) FROM code_metadata WHERE is_active = true AND crosswalk_data != '{}'::jsonb) * 100.0 / 
          NULLIF((SELECT COUNT(*) FROM code_metadata WHERE is_active = true), 0), 2) AS coverage_percentage;

