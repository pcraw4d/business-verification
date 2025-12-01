-- =====================================================
-- Expand Hierarchy Coverage - Phase 3 Final
-- Purpose: Increase hierarchy coverage from 11.40% to 30%+ (71+ NAICS codes)
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 3 (Final)
-- =====================================================
-- 
-- Current: 62 codes with hierarchy (11.40%)
-- Target: 71+ codes with hierarchy (30%+ of 236 NAICS codes)
-- Need: ~9 additional codes with hierarchy
-- Focus: Only update codes that exist in code_metadata
-- =====================================================

-- =====================================================
-- Part 1: Add hierarchy for existing 5-6 digit NAICS codes
-- Based on official NAICS hierarchy structure
-- =====================================================

-- Technology Sector - Professional Services (541xxx codes)
-- These codes have parent 5415 (Computer Systems Design)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5415',
    'parent_type', 'NAICS',
    'parent_name', 'Computer Systems Design and Related Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('541511', '541512', '541519')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Technology Sector - Management Consulting (5416xx codes)
-- These codes have parent 5416 (Management Consulting Services)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5416',
    'parent_type', 'NAICS',
    'parent_name', 'Management, Scientific, and Technical Consulting Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('541611', '541612', '541613', '541614', '541618', '541620', '541690')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Technology Sector - Research and Development (5417xx codes)
-- These codes have parent 5417 (Scientific Research and Development)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5417',
    'parent_type', 'NAICS',
    'parent_name', 'Scientific Research and Development Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('541713', '541714', '541715', '541720')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Technology Sector - Engineering Services (5413xx codes)
-- These codes have parent 5413 (Architectural, Engineering, and Related Services)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5413',
    'parent_type', 'NAICS',
    'parent_name', 'Architectural, Engineering, and Related Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('541310', '541320', '541330', '541370', '541380')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Technology Sector - Testing Laboratories (5416xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5416',
    'parent_type', 'NAICS',
    'parent_name', 'Management, Scientific, and Technical Consulting Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('541621')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Construction Sector - Residential Building (2361xx codes)
-- These codes have parent 2361 (Residential Building Construction)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '2361',
    'parent_type', 'NAICS',
    'parent_name', 'Residential Building Construction',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('236115', '236116', '236117')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Construction Sector - Nonresidential Building (2362xx codes)
-- These codes have parent 2362 (Nonresidential Building Construction)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '2362',
    'parent_type', 'NAICS',
    'parent_name', 'Nonresidential Building Construction',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('236210', '236220')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Construction Sector - Foundation Contractors (2381xx codes)
-- These codes have parent 2381 (Foundation, Structure, and Building Exterior Contractors)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '2381',
    'parent_type', 'NAICS',
    'parent_name', 'Foundation, Structure, and Building Exterior Contractors',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('238110', '238120', '238130', '238140', '238150', '238160', '238170')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Construction Sector - Building Equipment Contractors (2382xx codes)
-- These codes have parent 2382 (Building Equipment Contractors)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '2382',
    'parent_type', 'NAICS',
    'parent_name', 'Building Equipment Contractors',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('238210', '238220', '238290')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Construction Sector - Building Finishing Contractors (2383xx codes)
-- These codes have parent 2383 (Building Finishing Contractors)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '2383',
    'parent_type', 'NAICS',
    'parent_name', 'Building Finishing Contractors',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('238310', '238320', '238330')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Food Manufacturing - Grain and Oilseed Milling (3112xx codes)
-- These codes have parent 3112 (Grain and Oilseed Milling)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '3112',
    'parent_type', 'NAICS',
    'parent_name', 'Grain and Oilseed Milling',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('311211', '311212', '311213', '311221', '311225', '311230')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Food Manufacturing - Sugar and Confectionery (3113xx codes)
-- These codes have parent 3113 (Sugar and Confectionery Product Manufacturing)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '3113',
    'parent_type', 'NAICS',
    'parent_name', 'Sugar and Confectionery Product Manufacturing',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('311311', '311312', '311313', '311320', '311330', '311340')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Food Manufacturing - Fruit and Vegetable Preserving (3114xx codes)
-- These codes have parent 3114 (Fruit and Vegetable Preserving and Specialty Food Manufacturing)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '3114',
    'parent_type', 'NAICS',
    'parent_name', 'Fruit and Vegetable Preserving and Specialty Food Manufacturing',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('311410', '311411', '311421', '311422', '311423')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Food Manufacturing - Dairy Product Manufacturing (3115xx codes)
-- These codes have parent 3115 (Dairy Product Manufacturing)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '3115',
    'parent_type', 'NAICS',
    'parent_name', 'Dairy Product Manufacturing',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('311511', '311512', '311513', '311514', '311520')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Beverage Manufacturing - Beverage Manufacturing (3121xx codes)
-- These codes have parent 3121 (Beverage Manufacturing)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '3121',
    'parent_type', 'NAICS',
    'parent_name', 'Beverage Manufacturing',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('312111', '312112', '312113', '312120', '312130', '312140')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Transportation - Air Transportation (481xxx codes)
-- These codes have parent 4811 (Scheduled Passenger Air Transportation) or 4812 (Nonscheduled Air Transportation)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4811',
    'parent_type', 'NAICS',
    'parent_name', 'Scheduled Passenger Air Transportation',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('481111', '481112')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4812',
    'parent_type', 'NAICS',
    'parent_name', 'Nonscheduled Air Transportation',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('481211', '481212')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Transportation - Rail Transportation (482xxx codes)
-- These codes have parent 4821 (Rail Transportation)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4821',
    'parent_type', 'NAICS',
    'parent_name', 'Rail Transportation',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('482111', '482112')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Transportation - Water Transportation (483xxx codes)
-- These codes have parent 4831 (Deep Sea, Coastal, and Great Lakes Water Transportation)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4831',
    'parent_type', 'NAICS',
    'parent_name', 'Deep Sea, Coastal, and Great Lakes Water Transportation',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('483111', '483112', '483113')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Transportation - Truck Transportation (484xxx codes)
-- These codes have parent 4841 (General Freight Trucking) or 4842 (Specialized Freight Trucking)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4841',
    'parent_type', 'NAICS',
    'parent_name', 'General Freight Trucking',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('484110', '484121', '484122')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4842',
    'parent_type', 'NAICS',
    'parent_name', 'Specialized Freight Trucking',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('484210', '484220', '484230')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Transportation - Bus Transportation (485xxx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4852',
    'parent_type', 'NAICS',
    'parent_name', 'Interurban and Rural Bus Transportation',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('485210')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4853',
    'parent_type', 'NAICS',
    'parent_name', 'Taxi and Limousine Service',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('485310', '485320')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Transportation - Couriers and Messengers (492xxx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4921',
    'parent_type', 'NAICS',
    'parent_name', 'Couriers and Express Delivery Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('492110')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4922',
    'parent_type', 'NAICS',
    'parent_name', 'Local Messengers and Local Delivery',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('492210')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Financial Services - Depository Credit Intermediation (5221xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5221',
    'parent_type', 'NAICS',
    'parent_name', 'Depository Credit Intermediation',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('522110', '522120', '522130', '522190')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Financial Services - Nondepository Credit Intermediation (5222xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5222',
    'parent_type', 'NAICS',
    'parent_name', 'Nondepository Credit Intermediation',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('522220', '522230', '522291', '522292', '522293', '522294', '522298')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Financial Services - Securities Brokerage (5231xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5231',
    'parent_type', 'NAICS',
    'parent_name', 'Securities and Commodity Contracts Intermediation and Brokerage',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('523110', '523120', '523130', '523140')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Financial Services - Other Financial Investment (5239xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5239',
    'parent_type', 'NAICS',
    'parent_name', 'Other Financial Investment Activities',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('523910', '523920', '523930', '523991', '523999')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Financial Services - Insurance Carriers (5241xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5241',
    'parent_type', 'NAICS',
    'parent_name', 'Insurance Carriers',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('524113', '524114', '524126', '524127', '524128', '524130')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Healthcare - Offices of Physicians (6211xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6211',
    'parent_type', 'NAICS',
    'parent_name', 'Offices of Physicians',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('621111', '621112')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Healthcare - Offices of Dentists (6212xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6212',
    'parent_type', 'NAICS',
    'parent_name', 'Offices of Dentists',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('621210')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Healthcare - Offices of Other Health Practitioners (6213xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6213',
    'parent_type', 'NAICS',
    'parent_name', 'Offices of Other Health Practitioners',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('621310', '621320', '621330', '621340')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Healthcare - Outpatient Care Centers (6214xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6214',
    'parent_type', 'NAICS',
    'parent_name', 'Outpatient Care Centers',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('621410', '621420', '621491', '621492', '621493', '621498')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Healthcare - Medical and Diagnostic Laboratories (6215xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6215',
    'parent_type', 'NAICS',
    'parent_name', 'Medical and Diagnostic Laboratories',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('621511', '621512')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Healthcare - Home Health Care Services (6216xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6216',
    'parent_type', 'NAICS',
    'parent_name', 'Home Health Care Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('621610')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Healthcare - Other Ambulatory Health Care (6219xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6219',
    'parent_type', 'NAICS',
    'parent_name', 'Other Ambulatory Health Care Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('621910', '621991', '621999')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Healthcare - General Medical and Surgical Hospitals (6221xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6221',
    'parent_type', 'NAICS',
    'parent_name', 'General Medical and Surgical Hospitals',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('622110')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Education - Elementary and Secondary Schools (6111xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6111',
    'parent_type', 'NAICS',
    'parent_name', 'Elementary and Secondary Schools',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('611110')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Education - Colleges, Universities, and Professional Schools (6113xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6113',
    'parent_type', 'NAICS',
    'parent_name', 'Colleges, Universities, and Professional Schools',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('611310')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Education - Business Schools and Computer Training (6114xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6114',
    'parent_type', 'NAICS',
    'parent_name', 'Business Schools and Computer and Management Training',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('611410', '611420', '611430')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Education - Technical and Trade Schools (6115xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6115',
    'parent_type', 'NAICS',
    'parent_name', 'Technical and Trade Schools',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('611511', '611512', '611519')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Education - Other Schools and Instruction (6116xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6116',
    'parent_type', 'NAICS',
    'parent_name', 'Other Schools and Instruction',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('611610', '611620', '611630', '611691', '611692', '611699')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Real Estate - Lessors of Real Estate (5311xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5311',
    'parent_type', 'NAICS',
    'parent_name', 'Lessors of Real Estate',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('531110', '531120', '531190')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Real Estate - Offices of Real Estate Agents (5312xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5312',
    'parent_type', 'NAICS',
    'parent_name', 'Offices of Real Estate Agents and Brokers',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('531210')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Real Estate - Activities Related to Real Estate (5313xx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5313',
    'parent_type', 'NAICS',
    'parent_name', 'Activities Related to Real Estate',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('531311', '531312', '531320', '531390')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- Retail - General Merchandise Stores (452xxx codes)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4529',
    'parent_type', 'NAICS',
    'parent_name', 'Other General Merchandise Stores',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' 
  AND code IN ('452990')
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- =====================================================
-- Verification Query
-- =====================================================

-- Verify hierarchy coverage after supplement
SELECT 
    'Hierarchy Coverage After Final Supplement' AS metric,
    COUNT(*) AS codes_with_hierarchy,
    (SELECT COUNT(*) FROM code_metadata WHERE code_type = 'NAICS' AND is_active = true) AS total_naics_codes,
    ROUND(COUNT(*) * 100.0 / NULLIF((SELECT COUNT(*) FROM code_metadata WHERE code_type = 'NAICS' AND is_active = true), 0), 2) AS coverage_percentage,
    CASE 
        WHEN COUNT(*) * 100.0 / NULLIF((SELECT COUNT(*) FROM code_metadata WHERE code_type = 'NAICS' AND is_active = true), 0) >= 30.0 THEN '✅ PASS - 30%+ coverage'
        ELSE '❌ FAIL - Below 30% coverage'
    END AS status
FROM code_metadata
WHERE code_type = 'NAICS'
  AND is_active = true
  AND hierarchy != '{}'::jsonb
  AND hierarchy IS NOT NULL;

