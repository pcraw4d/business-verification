-- =====================================================
-- Migration: Populate Missing Crosswalk Data
-- Purpose: Add crosswalk data for common MCC codes that are missing crosswalk mappings
-- Date: 2025-01-XX
-- Phase 2: Top 3 Codes Gap Filling Enhancement
-- =====================================================

-- This migration populates crosswalk data for MCC codes that are commonly used
-- but missing crosswalk mappings to SIC and NAICS codes.

-- =====================================================
-- Part 1: Ensure code_metadata entries exist for common MCC codes
-- =====================================================

-- Insert or update MCC 5819 (Miscellaneous Food Stores - includes pizza restaurants)
-- Based on actual definition: "Miscellaneous Food Stores - Convenience Stores, Specialty Markets, Vending Machines"
INSERT INTO code_metadata (code_type, code, official_name, official_description, is_official, is_active)
VALUES
('MCC', '5819', 'Miscellaneous Food Stores', 
 'Miscellaneous Food Stores - Convenience Stores, Specialty Markets, Vending Machines, and similar establishments. This code is also used for pizza restaurants and food service establishments not elsewhere classified.', 
 true, true)
ON CONFLICT (code_type, code) DO UPDATE SET
    official_name = EXCLUDED.official_name,
    official_description = EXCLUDED.official_description,
    is_official = EXCLUDED.is_official,
    updated_at = NOW();

-- =====================================================
-- Part 2: Populate Crosswalk Data for MCC 5819
-- =====================================================

-- MCC 5819 -> SIC and NAICS crosswalks
-- Based on official mappings and actual usage:
-- - SIC 5499: Miscellaneous Food Stores (primary match)
-- - SIC 5812: Eating Places, Restaurants (for pizza restaurants that use this code)
-- - SIC 5411: Grocery Stores (for convenience stores)
-- - NAICS 445110: Supermarkets and Other Grocery (except Convenience) Stores
-- - NAICS 445120: Convenience Stores
-- - NAICS 722511: Full-Service Restaurants (for pizza restaurants)
-- - NAICS 722513: Limited-Service Restaurants (for pizza restaurants)

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['445110', '445120', '722511', '722513'],
    'sic', ARRAY['5499', '5812', '5411']
)
WHERE code_type = 'MCC' AND code = '5819';

-- =====================================================
-- Part 3: Populate Crosswalk Data for Other Common Missing MCC Codes
-- =====================================================

-- MCC 5812 (Eating Places, Restaurants) - ensure it has crosswalks
INSERT INTO code_metadata (code_type, code, official_name, official_description, is_official, is_active)
VALUES
('MCC', '5812', 'Eating Places, Restaurants', 
 'Merchants primarily engaged in providing food services to patrons who order and are served while seated (i.e., waiter/waitress service) and pay after eating.', 
 true, true)
ON CONFLICT (code_type, code) DO UPDATE SET
    official_name = EXCLUDED.official_name,
    official_description = EXCLUDED.official_description,
    updated_at = NOW();

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['722511', '722513'],
    'sic', ARRAY['5812']
)
WHERE code_type = 'MCC' AND code = '5812';

-- MCC 5814 (Fast Food Restaurants)
INSERT INTO code_metadata (code_type, code, official_name, official_description, is_official, is_active)
VALUES
('MCC', '5814', 'Fast Food Restaurants', 
 'Merchants primarily engaged in providing quick-service food and beverages.', 
 true, true)
ON CONFLICT (code_type, code) DO UPDATE SET
    official_name = EXCLUDED.official_name,
    official_description = EXCLUDED.official_description,
    updated_at = NOW();

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['722513', '722515'],
    'sic', ARRAY['5812']
)
WHERE code_type = 'MCC' AND code = '5814';

-- =====================================================
-- Part 4: Reverse Crosswalks - Ensure SIC and NAICS codes can find MCC codes
-- =====================================================

-- SIC 5812 -> MCC crosswalks
INSERT INTO code_metadata (code_type, code, official_name, official_description, is_official, is_active)
VALUES
('SIC', '5812', 'Eating Places', 
 'Establishments primarily engaged in providing food services to patrons who order and are served while seated.', 
 true, true)
ON CONFLICT (code_type, code) DO UPDATE SET
    official_name = EXCLUDED.official_name,
    official_description = EXCLUDED.official_description,
    updated_at = NOW();

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['722511', '722513'],
    'mcc', ARRAY['5812', '5814', '5819']
)
WHERE code_type = 'SIC' AND code = '5812';

-- SIC 5499 -> MCC crosswalks
INSERT INTO code_metadata (code_type, code, official_name, official_description, is_official, is_active)
VALUES
('SIC', '5499', 'Miscellaneous Food Stores', 
 'Establishments primarily engaged in retailing miscellaneous food items not elsewhere classified.', 
 true, true)
ON CONFLICT (code_type, code) DO UPDATE SET
    official_name = EXCLUDED.official_name,
    official_description = EXCLUDED.official_description,
    updated_at = NOW();

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['445110', '445120'],
    'mcc', ARRAY['5819', '5411']
)
WHERE code_type = 'SIC' AND code = '5499';

-- NAICS 722511 (Full-Service Restaurants) -> MCC crosswalks
INSERT INTO code_metadata (code_type, code, official_name, official_description, is_official, is_active)
VALUES
('NAICS', '722511', 'Full-Service Restaurants', 
 'This industry comprises establishments primarily engaged in providing food services to patrons who order and are served while seated (i.e., waiter/waitress service) and pay after eating.', 
 true, true)
ON CONFLICT (code_type, code) DO UPDATE SET
    official_name = EXCLUDED.official_name,
    official_description = EXCLUDED.official_description,
    updated_at = NOW();

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['5812'],
    'mcc', ARRAY['5812', '5819']
)
WHERE code_type = 'NAICS' AND code = '722511';

-- NAICS 722513 (Limited-Service Restaurants) -> MCC crosswalks
INSERT INTO code_metadata (code_type, code, official_name, official_description, is_official, is_active)
VALUES
('NAICS', '722513', 'Limited-Service Restaurants', 
 'This industry comprises establishments primarily engaged in providing food services where patrons generally order or select items and pay before eating.', 
 true, true)
ON CONFLICT (code_type, code) DO UPDATE SET
    official_name = EXCLUDED.official_name,
    official_description = EXCLUDED.official_description,
    updated_at = NOW();

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['5812'],
    'mcc', ARRAY['5814', '5819']
)
WHERE code_type = 'NAICS' AND code = '722513';

-- =====================================================
-- Part 5: Verification Queries
-- =====================================================

-- Verify crosswalk data was populated
-- SELECT code_type, code, crosswalk_data 
-- FROM code_metadata 
-- WHERE code_type = 'MCC' AND code = '5819';

-- Verify reverse crosswalks
-- SELECT code_type, code, crosswalk_data 
-- FROM code_metadata 
-- WHERE (code_type = 'SIC' AND code = '5812') 
--    OR (code_type = 'NAICS' AND code IN ('722511', '722513'));
