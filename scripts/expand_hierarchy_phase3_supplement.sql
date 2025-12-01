-- =====================================================
-- Expand Hierarchy Coverage - Phase 3 Supplement
-- Purpose: Increase hierarchy coverage from 11.40% to 30%+ (163+ codes)
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 3 (Supplement)
-- =====================================================
-- 
-- Current: 62 codes with hierarchy (11.40%)
-- Target: 163+ codes with hierarchy (30%+)
-- Need: ~101 additional codes with hierarchy
-- Focus: NAICS codes (MCC and SIC don't typically have hierarchies)
-- =====================================================

-- =====================================================
-- Part 1: Technology Sector (54) Hierarchies (20 codes)
-- =====================================================

-- NAICS 541: Professional, Scientific, and Technical Services (Subsector)
-- Parent relationships for 5-digit codes
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '54',
    'parent_type', 'NAICS',
    'parent_name', 'Professional, Scientific, and Technical Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '541'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 5413: Architectural, Engineering, and Related Services (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '541',
    'parent_type', 'NAICS',
    'parent_name', 'Professional, Scientific, and Technical Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '5413'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 54131: Architectural Services (Industry)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5413',
    'parent_type', 'NAICS',
    'parent_name', 'Architectural, Engineering, and Related Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '54131'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 54132: Landscape Architectural Services (Industry)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5413',
    'parent_type', 'NAICS',
    'parent_name', 'Architectural, Engineering, and Related Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '54132'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 54133: Engineering Services (Industry)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5413',
    'parent_type', 'NAICS',
    'parent_name', 'Architectural, Engineering, and Related Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '54133'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 54138: Testing Laboratories (Industry)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5413',
    'parent_type', 'NAICS',
    'parent_name', 'Architectural, Engineering, and Related Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '54138'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 5416: Management, Scientific, and Technical Consulting Services (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '541',
    'parent_type', 'NAICS',
    'parent_name', 'Professional, Scientific, and Technical Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '5416'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 54161: Management Consulting Services (Industry)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5416',
    'parent_type', 'NAICS',
    'parent_name', 'Management, Scientific, and Technical Consulting Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '54161'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 54162: Environmental Consulting Services (Industry)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5416',
    'parent_type', 'NAICS',
    'parent_name', 'Management, Scientific, and Technical Consulting Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '54162'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 54169: Other Scientific and Technical Consulting Services (Industry)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5416',
    'parent_type', 'NAICS',
    'parent_name', 'Management, Scientific, and Technical Consulting Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '54169'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 5417: Scientific Research and Development Services (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '541',
    'parent_type', 'NAICS',
    'parent_name', 'Professional, Scientific, and Technical Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '5417'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 54171: Research and Development in the Physical, Engineering, and Life Sciences (Industry)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5417',
    'parent_type', 'NAICS',
    'parent_name', 'Scientific Research and Development Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '54171'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 54172: Research and Development in the Social Sciences and Humanities (Industry)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5417',
    'parent_type', 'NAICS',
    'parent_name', 'Scientific Research and Development Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '54172'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- =====================================================
-- Part 2: Construction Sector (23) Hierarchies (15 codes)
-- =====================================================

-- NAICS 23: Construction (Sector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', NULL,
    'parent_type', NULL,
    'parent_name', NULL,
    'child_codes', ARRAY['236', '237', '238']::text[]
)
WHERE code_type = 'NAICS' AND code = '23'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 236: Construction of Buildings (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '23',
    'parent_type', 'NAICS',
    'parent_name', 'Construction',
    'child_codes', ARRAY['2361', '2362']::text[]
)
WHERE code_type = 'NAICS' AND code = '236'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 2361: Residential Building Construction (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '236',
    'parent_type', 'NAICS',
    'parent_name', 'Construction of Buildings',
    'child_codes', ARRAY['236115', '236116', '236117']::text[]
)
WHERE code_type = 'NAICS' AND code = '2361'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 2362: Nonresidential Building Construction (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '236',
    'parent_type', 'NAICS',
    'parent_name', 'Construction of Buildings',
    'child_codes', ARRAY['236210', '236220']::text[]
)
WHERE code_type = 'NAICS' AND code = '2362'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 237: Heavy and Civil Engineering Construction (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '23',
    'parent_type', 'NAICS',
    'parent_name', 'Construction',
    'child_codes', ARRAY['2371', '2372', '2373']::text[]
)
WHERE code_type = 'NAICS' AND code = '237'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 2371: Utility System Construction (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '237',
    'parent_type', 'NAICS',
    'parent_name', 'Heavy and Civil Engineering Construction',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '2371'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 2372: Land Subdivision (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '237',
    'parent_type', 'NAICS',
    'parent_name', 'Heavy and Civil Engineering Construction',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '2372'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 2373: Highway, Street, and Bridge Construction (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '237',
    'parent_type', 'NAICS',
    'parent_name', 'Heavy and Civil Engineering Construction',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '2373'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 238: Specialty Trade Contractors (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '23',
    'parent_type', 'NAICS',
    'parent_name', 'Construction',
    'child_codes', ARRAY['2381', '2382', '2383', '2389']::text[]
)
WHERE code_type = 'NAICS' AND code = '238'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 2381: Foundation, Structure, and Building Exterior Contractors (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '238',
    'parent_type', 'NAICS',
    'parent_name', 'Specialty Trade Contractors',
    'child_codes', ARRAY['238110', '238120', '238130', '238140', '238150', '238160', '238170']::text[]
)
WHERE code_type = 'NAICS' AND code = '2381'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 2382: Building Equipment Contractors (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '238',
    'parent_type', 'NAICS',
    'parent_name', 'Specialty Trade Contractors',
    'child_codes', ARRAY['238210', '238220', '238290']::text[]
)
WHERE code_type = 'NAICS' AND code = '2382'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 2383: Building Finishing Contractors (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '238',
    'parent_type', 'NAICS',
    'parent_name', 'Specialty Trade Contractors',
    'child_codes', ARRAY['238310', '238320', '238330']::text[]
)
WHERE code_type = 'NAICS' AND code = '2383'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 2389: Other Specialty Trade Contractors (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '238',
    'parent_type', 'NAICS',
    'parent_name', 'Specialty Trade Contractors',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '2389'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- =====================================================
-- Part 3: Manufacturing Sector (31-33) Hierarchies (20 codes)
-- =====================================================

-- NAICS 31-33: Manufacturing (Sector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', NULL,
    'parent_type', NULL,
    'parent_name', NULL,
    'child_codes', ARRAY['311', '312', '313', '314', '315']::text[]
)
WHERE code_type = 'NAICS' AND code = '31'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 311: Food Manufacturing (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '31',
    'parent_type', 'NAICS',
    'parent_name', 'Manufacturing',
    'child_codes', ARRAY['3111', '3112', '3113', '3114', '3115', '3116']::text[]
)
WHERE code_type = 'NAICS' AND code = '311'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 3111: Animal Food Manufacturing (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '311',
    'parent_type', 'NAICS',
    'parent_name', 'Food Manufacturing',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '3111'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 3112: Grain and Oilseed Milling (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '311',
    'parent_type', 'NAICS',
    'parent_name', 'Food Manufacturing',
    'child_codes', ARRAY['311211', '311212', '311213', '311221', '311225', '311230']::text[]
)
WHERE code_type = 'NAICS' AND code = '3112'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 3113: Sugar and Confectionery Product Manufacturing (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '311',
    'parent_type', 'NAICS',
    'parent_name', 'Food Manufacturing',
    'child_codes', ARRAY['311311', '311312', '311313', '311320', '311330', '311340']::text[]
)
WHERE code_type = 'NAICS' AND code = '3113'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 3114: Fruit and Vegetable Preserving and Specialty Food Manufacturing (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '311',
    'parent_type', 'NAICS',
    'parent_name', 'Food Manufacturing',
    'child_codes', ARRAY['311410', '311411', '311421', '311422', '311423']::text[]
)
WHERE code_type = 'NAICS' AND code = '3114'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 3115: Dairy Product Manufacturing (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '311',
    'parent_type', 'NAICS',
    'parent_name', 'Food Manufacturing',
    'child_codes', ARRAY['311511', '311512', '311513', '311514', '311520']::text[]
)
WHERE code_type = 'NAICS' AND code = '3115'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 3116: Animal Slaughtering and Processing (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '311',
    'parent_type', 'NAICS',
    'parent_name', 'Food Manufacturing',
    'child_codes', ARRAY['311615']::text[]
)
WHERE code_type = 'NAICS' AND code = '3116'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 312: Beverage and Tobacco Product Manufacturing (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '31',
    'parent_type', 'NAICS',
    'parent_name', 'Manufacturing',
    'child_codes', ARRAY['3121', '3122']::text[]
)
WHERE code_type = 'NAICS' AND code = '312'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 3121: Beverage Manufacturing (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '312',
    'parent_type', 'NAICS',
    'parent_name', 'Beverage and Tobacco Product Manufacturing',
    'child_codes', ARRAY['312111', '312112', '312113', '312120', '312130', '312140']::text[]
)
WHERE code_type = 'NAICS' AND code = '3121'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 3122: Tobacco Manufacturing (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '312',
    'parent_type', 'NAICS',
    'parent_name', 'Beverage and Tobacco Product Manufacturing',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '3122'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- =====================================================
-- Part 4: Transportation Sector (48) Hierarchies (15 codes)
-- =====================================================

-- NAICS 48-49: Transportation and Warehousing (Sector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', NULL,
    'parent_type', NULL,
    'parent_name', NULL,
    'child_codes', ARRAY['481', '482', '483', '484', '485', '486', '487', '488', '491', '492', '493']::text[]
)
WHERE code_type = 'NAICS' AND code = '48'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 481: Air Transportation (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '48',
    'parent_type', 'NAICS',
    'parent_name', 'Transportation and Warehousing',
    'child_codes', ARRAY['4811', '4812']::text[]
)
WHERE code_type = 'NAICS' AND code = '481'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 4811: Scheduled Passenger Air Transportation (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '481',
    'parent_type', 'NAICS',
    'parent_name', 'Air Transportation',
    'child_codes', ARRAY['481111', '481112']::text[]
)
WHERE code_type = 'NAICS' AND code = '4811'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 4812: Nonscheduled Air Transportation (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '481',
    'parent_type', 'NAICS',
    'parent_name', 'Air Transportation',
    'child_codes', ARRAY['481211', '481212']::text[]
)
WHERE code_type = 'NAICS' AND code = '4812'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 482: Rail Transportation (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '48',
    'parent_type', 'NAICS',
    'parent_name', 'Transportation and Warehousing',
    'child_codes', ARRAY['4821']::text[]
)
WHERE code_type = 'NAICS' AND code = '482'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 4821: Rail Transportation (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '482',
    'parent_type', 'NAICS',
    'parent_name', 'Rail Transportation',
    'child_codes', ARRAY['482111', '482112']::text[]
)
WHERE code_type = 'NAICS' AND code = '4821'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 483: Water Transportation (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '48',
    'parent_type', 'NAICS',
    'parent_name', 'Transportation and Warehousing',
    'child_codes', ARRAY['4831']::text[]
)
WHERE code_type = 'NAICS' AND code = '483'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 4831: Deep Sea, Coastal, and Great Lakes Water Transportation (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '483',
    'parent_type', 'NAICS',
    'parent_name', 'Water Transportation',
    'child_codes', ARRAY['483111', '483112']::text[]
)
WHERE code_type = 'NAICS' AND code = '4831'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 484: Truck Transportation (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '48',
    'parent_type', 'NAICS',
    'parent_name', 'Transportation and Warehousing',
    'child_codes', ARRAY['4841', '4842']::text[]
)
WHERE code_type = 'NAICS' AND code = '484'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 4841: General Freight Trucking (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '484',
    'parent_type', 'NAICS',
    'parent_name', 'Truck Transportation',
    'child_codes', ARRAY['484110', '484121', '484122']::text[]
)
WHERE code_type = 'NAICS' AND code = '4841'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 4842: Specialized Freight Trucking (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '484',
    'parent_type', 'NAICS',
    'parent_name', 'Truck Transportation',
    'child_codes', ARRAY['484210', '484220', '484230']::text[]
)
WHERE code_type = 'NAICS' AND code = '4842'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 485: Transit and Ground Passenger Transportation (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '48',
    'parent_type', 'NAICS',
    'parent_name', 'Transportation and Warehousing',
    'child_codes', ARRAY['4851', '4852', '4853', '4854', '4855', '4859']::text[]
)
WHERE code_type = 'NAICS' AND code = '485'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 4852: Interurban and Rural Bus Transportation (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '485',
    'parent_type', 'NAICS',
    'parent_name', 'Transit and Ground Passenger Transportation',
    'child_codes', ARRAY['485210']::text[]
)
WHERE code_type = 'NAICS' AND code = '4852'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 492: Couriers and Messengers (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '48',
    'parent_type', 'NAICS',
    'parent_name', 'Transportation and Warehousing',
    'child_codes', ARRAY['4921', '4922']::text[]
)
WHERE code_type = 'NAICS' AND code = '492'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 4921: Couriers and Express Delivery Services (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '492',
    'parent_type', 'NAICS',
    'parent_name', 'Couriers and Messengers',
    'child_codes', ARRAY['492110']::text[]
)
WHERE code_type = 'NAICS' AND code = '4921'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 4922: Local Messengers and Local Delivery (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '492',
    'parent_type', 'NAICS',
    'parent_name', 'Couriers and Messengers',
    'child_codes', ARRAY['492210']::text[]
)
WHERE code_type = 'NAICS' AND code = '4922'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- =====================================================
-- Part 5: Financial Services Sector (52) Hierarchies (10 codes)
-- =====================================================

-- NAICS 52: Finance and Insurance (Sector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', NULL,
    'parent_type', NULL,
    'parent_name', NULL,
    'child_codes', ARRAY['521', '522', '523', '524', '525']::text[]
)
WHERE code_type = 'NAICS' AND code = '52'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 522: Credit Intermediation and Related Activities (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '52',
    'parent_type', 'NAICS',
    'parent_name', 'Finance and Insurance',
    'child_codes', ARRAY['5221', '5222', '5223']::text[]
)
WHERE code_type = 'NAICS' AND code = '522'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 5221: Depository Credit Intermediation (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '522',
    'parent_type', 'NAICS',
    'parent_name', 'Credit Intermediation and Related Activities',
    'child_codes', ARRAY['522110', '522120', '522130', '522190']::text[]
)
WHERE code_type = 'NAICS' AND code = '5221'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 5222: Nondepository Credit Intermediation (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '522',
    'parent_type', 'NAICS',
    'parent_name', 'Credit Intermediation and Related Activities',
    'child_codes', ARRAY['522210', '220', '522230', '522291', '522292', '522293', '522294', '522298']::text[]
)
WHERE code_type = 'NAICS' AND code = '5222'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 5223: Activities Related to Credit Intermediation (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '522',
    'parent_type', 'NAICS',
    'parent_name', 'Credit Intermediation and Related Activities',
    'child_codes', ARRAY['522310', '522320', '522390']::text[]
)
WHERE code_type = 'NAICS' AND code = '5223'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 523: Securities, Commodity Contracts, and Other Financial Investments (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '52',
    'parent_type', 'NAICS',
    'parent_name', 'Finance and Insurance',
    'child_codes', ARRAY['5231', '5232', '5239']::text[]
)
WHERE code_type = 'NAICS' AND code = '523'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 5231: Securities and Commodity Contracts Intermediation and Brokerage (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '523',
    'parent_type', 'NAICS',
    'parent_name', 'Securities, Commodity Contracts, and Other Financial Investments',
    'child_codes', ARRAY['523110', '523120', '523130', '523140']::text[]
)
WHERE code_type = 'NAICS' AND code = '5231'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 5232: Securities and Commodity Exchanges (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '523',
    'parent_type', 'NAICS',
    'parent_name', 'Securities, Commodity Contracts, and Other Financial Investments',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '5222'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 5239: Other Financial Investment Activities (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '523',
    'parent_type', 'NAICS',
    'parent_name', 'Securities, Commodity Contracts, and Other Financial Investments',
    'child_codes', ARRAY['523910', '523920', '523930', '523991', '523999']::text[]
)
WHERE code_type = 'NAICS' AND code = '5239'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 524: Insurance Carriers and Related Activities (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '52',
    'parent_type', 'NAICS',
    'parent_name', 'Finance and Insurance',
    'child_codes', ARRAY['5241', '5242']::text[]
)
WHERE code_type = 'NAICS' AND code = '524'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 5241: Insurance Carriers (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '524',
    'parent_type', 'NAICS',
    'parent_name', 'Insurance Carriers and Related Activities',
    'child_codes', ARRAY['524113', '524114', '524126', '524127', '524128', '524130']::text[]
)
WHERE code_type = 'NAICS' AND code = '5241'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- =====================================================
-- Part 6: Healthcare Sector (62) Hierarchies (15 codes)
-- =====================================================

-- NAICS 62: Health Care and Social Assistance (Sector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', NULL,
    'parent_type', NULL,
    'parent_name', NULL,
    'child_codes', ARRAY['621', '622', '623', '624']::text[]
)
WHERE code_type = 'NAICS' AND code = '62'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 621: Ambulatory Health Care Services (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '62',
    'parent_type', 'NAICS',
    'parent_name', 'Health Care and Social Assistance',
    'child_codes', ARRAY['6211', '6212', '6213', '6214', '6215', '6216', '6219']::text[]
)
WHERE code_type = 'NAICS' AND code = '621'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 6211: Offices of Physicians (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '621',
    'parent_type', 'NAICS',
    'parent_name', 'Ambulatory Health Care Services',
    'child_codes', ARRAY['621111', '621112']::text[]
)
WHERE code_type = 'NAICS' AND code = '6211'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 6212: Offices of Dentists (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '621',
    'parent_type', 'NAICS',
    'parent_name', 'Ambulatory Health Care Services',
    'child_codes', ARRAY['621210']::text[]
)
WHERE code_type = 'NAICS' AND code = '6212'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 6213: Offices of Other Health Practitioners (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '621',
    'parent_type', 'NAICS',
    'parent_name', 'Ambulatory Health Care Services',
    'child_codes', ARRAY['621310', '621320', '621330', '621340']::text[]
)
WHERE code_type = 'NAICS' AND code = '6213'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 6214: Outpatient Care Centers (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '621',
    'parent_type', 'NAICS',
    'parent_name', 'Ambulatory Health Care Services',
    'child_codes', ARRAY['621410', '621420', '621491', '621492', '621493', '621498']::text[]
)
WHERE code_type = 'NAICS' AND code = '6214'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 6215: Medical and Diagnostic Laboratories (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '621',
    'parent_type', 'NAICS',
    'parent_name', 'Ambulatory Health Care Services',
    'child_codes', ARRAY['621511', '621512']::text[]
)
WHERE code_type = 'NAICS' AND code = '6215'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 6216: Home Health Care Services (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '621',
    'parent_type', 'NAICS',
    'parent_name', 'Ambulatory Health Care Services',
    'child_codes', ARRAY['621610']::text[]
)
WHERE code_type = 'NAICS' AND code = '6216'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 6219: Other Ambulatory Health Care Services (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '621',
    'parent_type', 'NAICS',
    'parent_name', 'Ambulatory Health Care Services',
    'child_codes', ARRAY['621910', '621991', '621999']::text[]
)
WHERE code_type = 'NAICS' AND code = '6219'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 622: Hospitals (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '62',
    'parent_type', 'NAICS',
    'parent_name', 'Health Care and Social Assistance',
    'child_codes', ARRAY['6221', '6222', '6223']::text[]
)
WHERE code_type = 'NAICS' AND code = '622'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 6221: General Medical and Surgical Hospitals (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '622',
    'parent_type', 'NAICS',
    'parent_name', 'Hospitals',
    'child_codes', ARRAY['622110']::text[]
)
WHERE code_type = 'NAICS' AND code = '6221'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- =====================================================
-- Part 7: Education Sector (61) Hierarchies (10 codes)
-- =====================================================

-- NAICS 61: Educational Services (Sector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', NULL,
    'parent_type', NULL,
    'parent_name', NULL,
    'child_codes', ARRAY['611']::text[]
)
WHERE code_type = 'NAICS' AND code = '61'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 611: Educational Services (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '61',
    'parent_type', 'NAICS',
    'parent_name', 'Educational Services',
    'child_codes', ARRAY['6111', '6112', '6113', '6114', '6115', '6116', '6117']::text[]
)
WHERE code_type = 'NAICS' AND code = '611'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 6111: Elementary and Secondary Schools (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '611',
    'parent_type', 'NAICS',
    'parent_name', 'Educational Services',
    'child_codes', ARRAY['611110']::text[]
)
WHERE code_type = 'NAICS' AND code = '6111'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 6113: Colleges, Universities, and Professional Schools (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '611',
    'parent_type', 'NAICS',
    'parent_name', 'Educational Services',
    'child_codes', ARRAY['611310']::text[]
)
WHERE code_type = 'NAICS' AND code = '6113'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 6114: Business Schools and Computer and Management Training (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '611',
    'parent_type', 'NAICS',
    'parent_name', 'Educational Services',
    'child_codes', ARRAY['611410', '611420', '611430']::text[]
)
WHERE code_type = 'NAICS' AND code = '6114'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 6115: Technical and Trade Schools (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '611',
    'parent_type', 'NAICS',
    'parent_name', 'Educational Services',
    'child_codes', ARRAY['611511', '611512', '611519']::text[]
)
WHERE code_type = 'NAICS' AND code = '6115'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 6116: Other Schools and Instruction (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '611',
    'parent_type', 'NAICS',
    'parent_name', 'Educational Services',
    'child_codes', ARRAY['611610', '611620', '611630', '611691', '611692', '611699']::text[]
)
WHERE code_type = 'NAICS' AND code = '6116'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- =====================================================
-- Part 8: Real Estate Sector (53) Hierarchies (6 codes)
-- =====================================================

-- NAICS 53: Real Estate and Rental and Leasing (Sector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', NULL,
    'parent_type', NULL,
    'parent_name', NULL,
    'child_codes', ARRAY['531', '532', '533']::text[]
)
WHERE code_type = 'NAICS' AND code = '53'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 531: Real Estate (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '53',
    'parent_type', 'NAICS',
    'parent_name', 'Real Estate and Rental and Leasing',
    'child_codes', ARRAY['5311', '5312', '5313']::text[]
)
WHERE code_type = 'NAICS' AND code = '531'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 5311: Lessors of Real Estate (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '531',
    'parent_type', 'NAICS',
    'parent_name', 'Real Estate',
    'child_codes', ARRAY['531110', '531120', '531190']::text[]
)
WHERE code_type = 'NAICS' AND code = '5311'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 5312: Offices of Real Estate Agents and Brokers (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '531',
    'parent_type', 'NAICS',
    'parent_name', 'Real Estate',
    'child_codes', ARRAY['531210']::text[]
)
WHERE code_type = 'NAICS' AND code = '5312'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 5313: Activities Related to Real Estate (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '531',
    'parent_type', 'NAICS',
    'parent_name', 'Real Estate',
    'child_codes', ARRAY['531311', '531312', '531320', '531390']::text[]
)
WHERE code_type = 'NAICS' AND code = '5313'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- =====================================================
-- Part 9: Retail Sector (44-45) Hierarchies (5 codes)
-- =====================================================

-- NAICS 44-45: Retail Trade (Sector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', NULL,
    'parent_type', NULL,
    'parent_name', NULL,
    'child_codes', ARRAY['441', '442', '443', '444', '445', '446', '447', '448', '451', '452', '453', '454']::text[]
)
WHERE code_type = 'NAICS' AND code = '44'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 452: General Merchandise Stores (Subsector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '44',
    'parent_type', 'NAICS',
    'parent_name', 'Retail Trade',
    'child_codes', ARRAY['4521', '4529']::text[]
)
WHERE code_type = 'NAICS' AND code = '452'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- NAICS 4529: Other General Merchandise Stores (Industry Group)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '452',
    'parent_type', 'NAICS',
    'parent_name', 'General Merchandise Stores',
    'child_codes', ARRAY['452990']::text[]
)
WHERE code_type = 'NAICS' AND code = '4529'
  AND (hierarchy = '{}'::jsonb OR hierarchy IS NULL);

-- =====================================================
-- Verification Query
-- =====================================================

-- Verify hierarchy coverage after supplement
SELECT 
    'Hierarchy Coverage After Supplement' AS metric,
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

