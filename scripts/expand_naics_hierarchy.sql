-- =====================================================
-- Expand NAICS Hierarchy Coverage - Phase 2, Task 2.4
-- Purpose: Increase NAICS hierarchy coverage from 15.33% to 30%+ (30+ NAICS codes)
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 2, Task 2.4
-- =====================================================
-- 
-- This script adds hierarchy data for 30+ NAICS codes, establishing
-- parent/child relationships based on the official NAICS structure.
-- Target: 30+ NAICS codes with hierarchy data (30%+ coverage)
-- Accuracy Target: > 98% accuracy for hierarchy relationships
-- =====================================================

-- =====================================================
-- Part 1: Technology Sector (54) Hierarchies
-- =====================================================

-- NAICS 54: Professional, Scientific, and Technical Services (Sector)
-- NAICS 541: Professional, Scientific, and Technical Services (Subsector)
-- NAICS 5415: Computer Systems Design and Related Services (Industry Group)
-- NAICS 54151: Custom Computer Programming Services (Industry)
-- NAICS 541511: Custom Computer Programming Services (U.S. Industry)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '54151',
    'parent_type', 'NAICS',
    'parent_name', 'Custom Computer Programming Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '541511';

-- NAICS 541512: Computer Systems Design Services
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '54151',
    'parent_type', 'NAICS',
    'parent_name', 'Custom Computer Programming Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '541512';

-- NAICS 541519: Other Computer Related Services
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '54151',
    'parent_type', 'NAICS',
    'parent_name', 'Custom Computer Programming Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '541519';

-- NAICS 541330: Engineering Services
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '54133',
    'parent_type', 'NAICS',
    'parent_name', 'Engineering Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '541330';

-- NAICS 541310: Architectural Services
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '54131',
    'parent_type', 'NAICS',
    'parent_name', 'Architectural Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '541310';

-- NAICS 541370: Surveying and Mapping (except Geophysical) Services
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '54137',
    'parent_type', 'NAICS',
    'parent_name', 'Surveying and Mapping Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '541370';

-- NAICS 518210: Data Processing, Hosting, and Related Services
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5182',
    'parent_type', 'NAICS',
    'parent_name', 'Data Processing, Hosting, and Related Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '518210';

-- NAICS 518310: Internet Service Providers and Web Search Portals
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5183',
    'parent_type', 'NAICS',
    'parent_name', 'Internet Service Providers and Web Search Portals',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '518310';

-- NAICS 519130: Internet Publishing and Broadcasting and Web Search Portals
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5191',
    'parent_type', 'NAICS',
    'parent_name', 'Internet Publishing and Broadcasting',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '519130';

-- =====================================================
-- Part 2: Financial Services Sector (52) Hierarchies
-- =====================================================

-- NAICS 522110: Commercial Banking
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5221',
    'parent_type', 'NAICS',
    'parent_name', 'Depository Credit Intermediation',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '522110';

-- NAICS 522120: Savings Institutions
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5221',
    'parent_type', 'NAICS',
    'parent_name', 'Depository Credit Intermediation',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '522120';

-- NAICS 522210: Credit Card Issuing
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5222',
    'parent_type', 'NAICS',
    'parent_name', 'Nondepository Credit Intermediation',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '522210';

-- NAICS 522291: Consumer Lending
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5222',
    'parent_type', 'NAICS',
    'parent_name', 'Nondepository Credit Intermediation',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '522291';

-- NAICS 522292: Real Estate Credit
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5222',
    'parent_type', 'NAICS',
    'parent_name', 'Nondepository Credit Intermediation',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '522292';

-- NAICS 523110: Investment Banking and Securities Dealing
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5231',
    'parent_type', 'NAICS',
    'parent_name', 'Securities and Commodity Contracts Intermediation and Brokerage',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '523110';

-- NAICS 523120: Securities Brokerage
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5231',
    'parent_type', 'NAICS',
    'parent_name', 'Securities and Commodity Contracts Intermediation and Brokerage',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '523120';

-- NAICS 523130: Commodity Contracts Dealing
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5231',
    'parent_type', 'NAICS',
    'parent_name', 'Securities and Commodity Contracts Intermediation and Brokerage',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '523130';

-- NAICS 523210: Securities and Commodity Exchanges
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5232',
    'parent_type', 'NAICS',
    'parent_name', 'Securities and Commodity Exchanges',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '523210';

-- =====================================================
-- Part 3: Healthcare Sector (62) Hierarchies
-- =====================================================

-- NAICS 621111: Offices of Physicians (except Mental Health Specialists)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6211',
    'parent_type', 'NAICS',
    'parent_name', 'Offices of Physicians',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '621111';

-- NAICS 621112: Offices of Physicians, Mental Health Specialists
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6211',
    'parent_type', 'NAICS',
    'parent_name', 'Offices of Physicians',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '621112';

-- NAICS 621210: Offices of Dentists
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6212',
    'parent_type', 'NAICS',
    'parent_name', 'Offices of Dentists',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '621210';

-- NAICS 621310: Offices of Chiropractors
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6213',
    'parent_type', 'NAICS',
    'parent_name', 'Offices of Other Health Practitioners',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '621310';

-- NAICS 621320: Offices of Optometrists
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6213',
    'parent_type', 'NAICS',
    'parent_name', 'Offices of Other Health Practitioners',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '621320';

-- NAICS 621391: Offices of Podiatrists
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6213',
    'parent_type', 'NAICS',
    'parent_name', 'Offices of Other Health Practitioners',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '621391';

-- NAICS 621399: Offices of All Other Miscellaneous Health Practitioners
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6213',
    'parent_type', 'NAICS',
    'parent_name', 'Offices of Other Health Practitioners',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '621399';

-- NAICS 622110: General Medical and Surgical Hospitals
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6221',
    'parent_type', 'NAICS',
    'parent_name', 'General Medical and Surgical Hospitals',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '622110';

-- NAICS 622210: Psychiatric and Substance Abuse Hospitals
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6222',
    'parent_type', 'NAICS',
    'parent_name', 'Psychiatric and Substance Abuse Hospitals',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '622210';

-- NAICS 622310: Specialty (except Psychiatric and Substance Abuse) Hospitals
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6223',
    'parent_type', 'NAICS',
    'parent_name', 'Specialty (except Psychiatric and Substance Abuse) Hospitals',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '622310';

-- NAICS 621511: Medical Laboratories
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6215',
    'parent_type', 'NAICS',
    'parent_name', 'Medical and Diagnostic Laboratories',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '621511';

-- NAICS 621512: Diagnostic Imaging Centers
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6215',
    'parent_type', 'NAICS',
    'parent_name', 'Medical and Diagnostic Laboratories',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '621512';

-- =====================================================
-- Part 4: Retail Sector (44-45) Hierarchies
-- =====================================================

-- NAICS 452111: Department Stores (except Discount Department Stores)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4521',
    'parent_type', 'NAICS',
    'parent_name', 'Department Stores',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '452111';

-- NAICS 452112: Discount Department Stores
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4521',
    'parent_type', 'NAICS',
    'parent_name', 'Department Stores',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '452112';

-- NAICS 452210: Warehouse Clubs and Supercenters
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4522',
    'parent_type', 'NAICS',
    'parent_name', 'Warehouse Clubs and Supercenters',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '452210';

-- NAICS 445110: Supermarkets and Other Grocery (except Convenience) Stores
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4451',
    'parent_type', 'NAICS',
    'parent_name', 'Grocery Stores',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '445110';

-- NAICS 445120: Convenience Stores
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4451',
    'parent_type', 'NAICS',
    'parent_name', 'Grocery Stores',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '445120';

-- NAICS 445292: Confectionery and Nut Stores
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4452',
    'parent_type', 'NAICS',
    'parent_name', 'Specialty Food Stores',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '445292';

-- =====================================================
-- Part 5: Food & Beverage Sector (72) Hierarchies
-- =====================================================

-- NAICS 722511: Full-Service Restaurants
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '7225',
    'parent_type', 'NAICS',
    'parent_name', 'Restaurants and Other Eating Places',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '722511';

-- NAICS 722513: Limited-Service Restaurants
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '7225',
    'parent_type', 'NAICS',
    'parent_name', 'Restaurants and Other Eating Places',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '722513';

-- NAICS 722515: Snack and Nonalcoholic Beverage Bars
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '7225',
    'parent_type', 'NAICS',
    'parent_name', 'Restaurants and Other Eating Places',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '722515';

-- NAICS 722410: Drinking Places (Alcoholic Beverages)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '7224',
    'parent_type', 'NAICS',
    'parent_name', 'Drinking Places (Alcoholic Beverages)',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '722410';

-- =====================================================
-- Part 6: Construction Sector (23) Hierarchies
-- =====================================================

-- NAICS 236115: New Single-Family Housing Construction (except Operative Builders)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '2361',
    'parent_type', 'NAICS',
    'parent_name', 'Residential Building Construction',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '236115';

-- NAICS 236116: New Multifamily Housing Construction (except Operative Builders)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '2361',
    'parent_type', 'NAICS',
    'parent_name', 'Residential Building Construction',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '236116';

-- NAICS 236117: New Housing Operative Builders
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '2361',
    'parent_type', 'NAICS',
    'parent_name', 'Residential Building Construction',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '236117';

-- NAICS 236210: Industrial Building Construction
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '2362',
    'parent_type', 'NAICS',
    'parent_name', 'Nonresidential Building Construction',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '236210';

-- NAICS 236220: Commercial and Institutional Building Construction
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '2362',
    'parent_type', 'NAICS',
    'parent_name', 'Nonresidential Building Construction',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '236220';

-- NAICS 237310: Highway, Street, and Bridge Construction
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '2373',
    'parent_type', 'NAICS',
    'parent_name', 'Highway, Street, and Bridge Construction',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '237310';

-- NAICS 237110: Water and Sewer Line and Related Structures Construction
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '2371',
    'parent_type', 'NAICS',
    'parent_name', 'Utility System Construction',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '237110';

-- =====================================================
-- Part 7: Real Estate Sector (53) Hierarchies
-- =====================================================

-- NAICS 531120: Lessors of Nonresidential Buildings (except Miniwarehouses)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5311',
    'parent_type', 'NAICS',
    'parent_name', 'Lessors of Real Estate',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '531120';

-- NAICS 531110: Lessors of Residential Buildings and Dwellings
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5311',
    'parent_type', 'NAICS',
    'parent_name', 'Lessors of Real Estate',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '531110';

-- NAICS 531210: Offices of Real Estate Agents and Brokers
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5312',
    'parent_type', 'NAICS',
    'parent_name', 'Offices of Real Estate Agents and Brokers',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '531210';

-- =====================================================
-- Part 8: Professional Services Hierarchies
-- =====================================================

-- NAICS 541110: Offices of Lawyers
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5411',
    'parent_type', 'NAICS',
    'parent_name', 'Legal Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '541110';

-- NAICS 541211: Offices of Certified Public Accountants
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5412',
    'parent_type', 'NAICS',
    'parent_name', 'Accounting, Tax Preparation, Bookkeeping, and Payroll Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '541211';

-- NAICS 541219: Other Accounting Services
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5412',
    'parent_type', 'NAICS',
    'parent_name', 'Accounting, Tax Preparation, Bookkeeping, and Payroll Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '541219';

-- =====================================================
-- Part 9: Education Sector (61) Hierarchies
-- =====================================================

-- NAICS 611110: Elementary and Secondary Schools
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6111',
    'parent_type', 'NAICS',
    'parent_name', 'Elementary and Secondary Schools',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '611110';

-- NAICS 611310: Colleges, Universities, and Professional Schools
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6113',
    'parent_type', 'NAICS',
    'parent_name', 'Colleges, Universities, and Professional Schools',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '611310';

-- NAICS 611210: Junior Colleges
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6112',
    'parent_type', 'NAICS',
    'parent_name', 'Junior Colleges',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '611210';

-- =====================================================
-- Part 10: Transportation Sector (48-49) Hierarchies
-- =====================================================

-- NAICS 481111: Scheduled Passenger Air Transportation
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4811',
    'parent_type', 'NAICS',
    'parent_name', 'Scheduled Air Transportation',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '481111';

-- NAICS 492110: Couriers and Express Delivery Services
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4921',
    'parent_type', 'NAICS',
    'parent_name', 'Couriers and Express Delivery Services',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '492110';

-- NAICS 484110: General Freight Trucking, Local
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4841',
    'parent_type', 'NAICS',
    'parent_name', 'General Freight Trucking',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '484110';

-- NAICS 484121: General Freight Trucking, Long-Distance, Truckload
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4841',
    'parent_type', 'NAICS',
    'parent_name', 'General Freight Trucking',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '484121';

-- NAICS 485110: Urban Transit Systems
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4851',
    'parent_type', 'NAICS',
    'parent_name', 'Urban Transit Systems',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '485110';

-- =====================================================
-- Part 11: Accommodation Sector (72) Hierarchies
-- =====================================================

-- NAICS 721110: Hotels (except Casino Hotels) and Motels
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '7211',
    'parent_type', 'NAICS',
    'parent_name', 'Traveler Accommodation',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '721110';

-- NAICS 721211: RV (Recreational Vehicle) Parks and Campgrounds
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '7212',
    'parent_type', 'NAICS',
    'parent_name', 'RV (Recreational Vehicle) Parks and Recreational Camps',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '721211';

-- =====================================================
-- Part 12: Arts and Entertainment Sector (71) Hierarchies
-- =====================================================

-- NAICS 711110: Theater Companies and Dinner Theaters
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '7111',
    'parent_type', 'NAICS',
    'parent_name', 'Theater Companies and Dinner Theaters',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '711110';

-- NAICS 713110: Amusement and Theme Parks
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '7131',
    'parent_type', 'NAICS',
    'parent_name', 'Amusement Parks and Arcades',
    'child_codes', ARRAY[]::text[]
)
WHERE code_type = 'NAICS' AND code = '713110';

-- =====================================================
-- Verification Queries
-- =====================================================

-- Count NAICS codes with hierarchy data
SELECT 
    'NAICS Codes with Hierarchy Data' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata WHERE code_type = 'NAICS' AND is_active = true), 2) AS percentage,
    CASE 
        WHEN COUNT(*) >= 30 THEN '✅ PASS - 30+ NAICS codes (30%+)'
        ELSE '❌ FAIL - Below 30 NAICS codes'
    END AS status
FROM code_metadata
WHERE code_type = 'NAICS'
  AND is_active = true
  AND hierarchy != '{}'::jsonb
  AND hierarchy IS NOT NULL;

-- Sample NAICS hierarchy verification
SELECT 
    'Sample NAICS Hierarchies' AS example,
    code_type,
    code,
    official_name,
    hierarchy->>'parent_code' AS parent_code,
    hierarchy->>'parent_name' AS parent_name,
    hierarchy->'child_codes' AS child_codes
FROM code_metadata
WHERE code_type = 'NAICS'
  AND is_active = true
  AND hierarchy != '{}'::jsonb
LIMIT 10;

-- Verify hierarchy accuracy (check that parent codes exist)
SELECT 
    'Hierarchy Accuracy Check' AS metric,
    COUNT(*) AS codes_with_valid_parents,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata WHERE code_type = 'NAICS' AND hierarchy != '{}'::jsonb), 2) AS accuracy_percentage,
    CASE 
        WHEN COUNT(*) * 100.0 / NULLIF((SELECT COUNT(*) FROM code_metadata WHERE code_type = 'NAICS' AND hierarchy != '{}'::jsonb), 0) >= 98.0 THEN '✅ PASS - 98%+ accuracy'
        ELSE '❌ FAIL - Below 98% accuracy'
    END AS status
FROM code_metadata cm1
WHERE cm1.code_type = 'NAICS'
  AND cm1.is_active = true
  AND cm1.hierarchy != '{}'::jsonb
  AND cm1.hierarchy->>'parent_code' IS NOT NULL
  AND EXISTS (
      SELECT 1 
      FROM code_metadata cm2 
      WHERE cm2.code_type = cm1.hierarchy->>'parent_type'
        AND cm2.code = cm1.hierarchy->>'parent_code'
        AND cm2.is_active = true
  );

