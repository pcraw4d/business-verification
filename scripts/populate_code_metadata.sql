-- =====================================================
-- Populate Code Metadata Table
-- Purpose: Import official code descriptions, crosswalks, and hierarchies
-- OPTIMIZATION #6.2: Code Metadata Table Population
-- =====================================================

-- =====================================================
-- Part 1: Populate Official NAICS Code Descriptions
-- =====================================================

-- Insert NAICS codes with official descriptions from Census Bureau
-- Note: This is a sample of common NAICS codes. For full population,
-- you would import from official NAICS data files.

INSERT INTO code_metadata (code_type, code, official_name, official_description, is_official, is_active)
VALUES
-- Technology Sector
('NAICS', '541511', 'Custom Computer Programming Services', 
 'This U.S. industry comprises establishments primarily engaged in writing, modifying, testing, and supporting software to meet the needs of a particular customer.', 
 true, true),

('NAICS', '541512', 'Computer Systems Design Services', 
 'This U.S. industry comprises establishments primarily engaged in planning and designing computer systems that integrate computer hardware, software, and communication technologies.', 
 true, true),

('NAICS', '541519', 'Other Computer Related Services', 
 'This U.S. industry comprises establishments primarily engaged in providing computer related services (except custom programming, systems integration design, and facilities management services).', 
 true, true),

-- Financial Services
('NAICS', '522110', 'Commercial Banking', 
 'This industry comprises establishments primarily engaged in accepting demand and other deposits and making commercial, industrial, and consumer loans.', 
 true, true),

('NAICS', '522120', 'Savings Institutions', 
 'This industry comprises establishments primarily engaged in accepting time deposits and making mortgage and other loans.', 
 true, true),

('NAICS', '523110', 'Investment Banking and Securities Dealing', 
 'This industry comprises establishments primarily engaged in underwriting, originating, and/or maintaining markets for securities of other businesses.', 
 true, true),

-- Healthcare
('NAICS', '621111', 'Offices of Physicians (except Mental Health Specialists)', 
 'This industry comprises establishments of licensed practitioners having the degree of M.D. (Doctor of Medicine) or D.O. (Doctor of Osteopathy) primarily engaged in the independent practice of general or specialized medicine (except psychiatry or psychoanalysis).', 
 true, true),

('NAICS', '621112', 'Offices of Physicians, Mental Health Specialists', 
 'This industry comprises establishments of licensed practitioners having the degree of M.D. (Doctor of Medicine) or D.O. (Doctor of Osteopathy) primarily engaged in the independent practice of psychiatry or psychoanalysis.', 
 true, true),

('NAICS', '622110', 'General Medical and Surgical Hospitals', 
 'This industry comprises establishments primarily engaged in providing general medical and surgical services and other hospital services.', 
 true, true),

-- Retail
('NAICS', '452111', 'Department Stores (except Discount Department Stores)', 
 'This industry comprises establishments known as department stores primarily engaged in retailing a wide range of the following new products with no one merchandise line predominating: apparel, furniture, appliances and home furnishings, and selected additional items.', 
 true, true),

('NAICS', '452112', 'Discount Department Stores', 
 'This industry comprises establishments known as discount department stores primarily engaged in retailing a wide range of the following new products with no one merchandise line predominating: apparel, furniture, appliances and home furnishings, and selected additional items.', 
 true, true),

-- Food & Beverage
('NAICS', '722511', 'Full-Service Restaurants', 
 'This industry comprises establishments primarily engaged in providing food services to patrons who order and are served while seated (i.e., waiter/waitress service) and pay after eating.', 
 true, true),

('NAICS', '722513', 'Limited-Service Restaurants', 
 'This industry comprises establishments primarily engaged in providing food services where patrons generally order or select items and pay before eating.', 
 true, true),

('NAICS', '722515', 'Snack and Nonalcoholic Beverage Bars', 
 'This industry comprises establishments primarily engaged in preparing and serving specialty snacks, such as ice cream, frozen yogurt, cookies, or coffee.', 
 true, true)

ON CONFLICT (code_type, code) DO UPDATE SET
    official_name = EXCLUDED.official_name,
    official_description = EXCLUDED.official_description,
    is_official = EXCLUDED.is_official,
    updated_at = NOW();

-- =====================================================
-- Part 2: Populate Official SIC Code Descriptions
-- =====================================================

INSERT INTO code_metadata (code_type, code, official_name, official_description, is_official, is_active)
VALUES
-- Technology
('SIC', '7371', 'Computer Programming Services', 
 'Establishments primarily engaged in providing computer programming services on a contract or fee basis.', 
 true, true),

('SIC', '7372', 'Prepackaged Software', 
 'Establishments primarily engaged in designing, developing, and producing prepackaged computer software.', 
 true, true),

('SIC', '7373', 'Computer Integrated Systems Design', 
 'Establishments primarily engaged in providing computer systems integration and design services.', 
 true, true),

-- Financial Services
('SIC', '6021', 'National Commercial Banks', 
 'Establishments primarily engaged in accepting deposits and making commercial, industrial, and consumer loans.', 
 true, true),

('SIC', '6022', 'State Commercial Banks', 
 'Establishments primarily engaged in accepting deposits and making commercial, industrial, and consumer loans.', 
 true, true),

('SIC', '6211', 'Security Brokers, Dealers, and Flotation Companies', 
 'Establishments primarily engaged in underwriting, originating, and/or maintaining markets for securities.', 
 true, true),

-- Healthcare
('SIC', '8011', 'Offices and Clinics of Doctors of Medicine', 
 'Establishments of licensed practitioners having the degree of M.D. primarily engaged in the independent practice of general or specialized medicine.', 
 true, true),

('SIC', '8062', 'General Medical and Surgical Hospitals', 
 'Establishments primarily engaged in providing general medical and surgical services and other hospital services.', 
 true, true),

-- Retail
('SIC', '5311', 'Department Stores', 
 'Establishments primarily engaged in retailing a wide range of new products with no one merchandise line predominating.', 
 true, true),

-- Food & Beverage
('SIC', '5812', 'Eating Places', 
 'Establishments primarily engaged in providing food services to patrons who order and are served while seated.', 
 true, true)

ON CONFLICT (code_type, code) DO UPDATE SET
    official_name = EXCLUDED.official_name,
    official_description = EXCLUDED.official_description,
    is_official = EXCLUDED.is_official,
    updated_at = NOW();

-- =====================================================
-- Part 3: Populate Official MCC Code Descriptions
-- =====================================================

INSERT INTO code_metadata (code_type, code, official_name, official_description, is_official, is_active)
VALUES
-- Technology
('MCC', '5734', 'Computer Software Stores', 
 'Merchants primarily engaged in retailing computer software and related products.', 
 true, true),

('MCC', '5735', 'Record Stores', 
 'Merchants primarily engaged in retailing records, tapes, and compact discs.', 
 true, true),

-- Financial Services
('MCC', '6010', 'Financial Institutions - Manual Cash Disbursements', 
 'Financial institutions providing manual cash disbursement services.', 
 true, true),

('MCC', '6011', 'Automated Cash Disbursements', 
 'Financial institutions providing automated cash disbursement services (ATMs).', 
 true, true),

('MCC', '6012', 'Financial Institutions - Merchandise, Services', 
 'Financial institutions providing merchandise and services.', 
 true, true),

-- Healthcare
('MCC', '8011', 'Doctors', 
 'Medical practitioners providing healthcare services.', 
 true, true),

('MCC', '8021', 'Dentists, Orthodontists', 
 'Dental practitioners providing dental care services.', 
 true, true),

('MCC', '8041', 'Chiropractors', 
 'Chiropractic practitioners providing healthcare services.', 
 true, true),

('MCC', '8042', 'Optometrists, Ophthalmologists', 
 'Eye care practitioners providing vision care services.', 
 true, true),

('MCC', '8043', 'Opticians, Optical Goods, Eyeglasses', 
 'Merchants providing optical goods and eyeglasses.', 
 true, true),

('MCC', '8049', 'Podiatrists, Chiropodists', 
 'Foot care practitioners providing healthcare services.', 
 true, true),

('MCC', '8050', 'Nursing and Personal Care Facilities', 
 'Facilities providing nursing and personal care services.', 
 true, true),

-- Retail
('MCC', '5310', 'Discount Stores', 
 'Merchants primarily engaged in retailing a wide range of products at discounted prices.', 
 true, true),

('MCC', '5311', 'Department Stores', 
 'Merchants primarily engaged in retailing a wide range of products with no one merchandise line predominating.', 
 true, true),

-- Food & Beverage
('MCC', '5811', 'Caterers', 
 'Merchants providing catering services for events and gatherings.', 
 true, true),

('MCC', '5812', 'Eating Places, Restaurants', 
 'Merchants primarily engaged in providing food services to patrons.', 
 true, true),

('MCC', '5813', 'Drinking Places (Alcoholic Beverages)', 
 'Merchants primarily engaged in serving alcoholic beverages for consumption on the premises.', 
 true, true),

('MCC', '5814', 'Fast Food Restaurants', 
 'Merchants primarily engaged in providing quick-service food and beverages.', 
 true, true)

ON CONFLICT (code_type, code) DO UPDATE SET
    official_name = EXCLUDED.official_name,
    official_description = EXCLUDED.official_description,
    is_official = EXCLUDED.is_official,
    updated_at = NOW();

-- =====================================================
-- Part 4: Populate Crosswalk Data (NAICS ↔ SIC ↔ MCC)
-- =====================================================

-- Technology Crosswalks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541511', '541512', '541519'],
    'sic', ARRAY['7371', '7372', '7373'],
    'mcc', ARRAY['5734', '5735']
)
WHERE code_type = 'NAICS' AND code = '541511';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541511', '541512'],
    'sic', ARRAY['7371', '7372'],
    'mcc', ARRAY['5734']
)
WHERE code_type = 'SIC' AND code = '7371';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541511'],
    'sic', ARRAY['7371'],
    'mcc', ARRAY['5734']
)
WHERE code_type = 'MCC' AND code = '5734';

-- Financial Services Crosswalks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['522110', '522120'],
    'sic', ARRAY['6021', '6022'],
    'mcc', ARRAY['6010', '6011', '6012']
)
WHERE code_type = 'NAICS' AND code = '522110';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['522110', '523110'],
    'sic', ARRAY['6021', '6211'],
    'mcc', ARRAY['6010', '6011']
)
WHERE code_type = 'SIC' AND code = '6021';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['522110'],
    'sic', ARRAY['6021'],
    'mcc', ARRAY['6010', '6011', '6012']
)
WHERE code_type = 'MCC' AND code = '6010';

-- Healthcare Crosswalks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621111', '621112'],
    'sic', ARRAY['8011'],
    'mcc', ARRAY['8011', '8021', '8041', '8042', '8043', '8049']
)
WHERE code_type = 'NAICS' AND code = '621111';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621111', '622110'],
    'sic', ARRAY['8011', '8062'],
    'mcc', ARRAY['8011', '8050']
)
WHERE code_type = 'SIC' AND code = '8011';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621111'],
    'sic', ARRAY['8011'],
    'mcc', ARRAY['8011']
)
WHERE code_type = 'MCC' AND code = '8011';

-- Retail Crosswalks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['452111', '452112'],
    'sic', ARRAY['5311'],
    'mcc', ARRAY['5310', '5311']
)
WHERE code_type = 'NAICS' AND code = '452111';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['452111'],
    'sic', ARRAY['5311'],
    'mcc', ARRAY['5311']
)
WHERE code_type = 'SIC' AND code = '5311';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['452111', '452112'],
    'sic', ARRAY['5311'],
    'mcc', ARRAY['5311']
)
WHERE code_type = 'MCC' AND code = '5311';

-- Food & Beverage Crosswalks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['722511', '722513', '722515'],
    'sic', ARRAY['5812'],
    'mcc', ARRAY['5811', '5812', '5813', '5814']
)
WHERE code_type = 'NAICS' AND code = '722511';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['722511', '722513'],
    'sic', ARRAY['5812'],
    'mcc', ARRAY['5812', '5814']
)
WHERE code_type = 'SIC' AND code = '5812';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['722511', '722513'],
    'sic', ARRAY['5812'],
    'mcc', ARRAY['5812', '5814']
)
WHERE code_type = 'MCC' AND code = '5812';

-- =====================================================
-- Part 5: Populate Code Hierarchies
-- =====================================================

-- NAICS Hierarchies (Technology Sector)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5415',
    'parent_type', 'NAICS',
    'parent_name', 'Computer Systems Design and Related Services',
    'child_codes', ARRAY['541511', '541512', '541519']
)
WHERE code_type = 'NAICS' AND code IN ('541511', '541512', '541519');

-- NAICS Hierarchies (Financial Services)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5221',
    'parent_type', 'NAICS',
    'parent_name', 'Depository Credit Intermediation',
    'child_codes', ARRAY['522110', '522120']
)
WHERE code_type = 'NAICS' AND code IN ('522110', '522120');

-- NAICS Hierarchies (Healthcare)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6211',
    'parent_type', 'NAICS',
    'parent_name', 'Offices of Physicians',
    'child_codes', ARRAY['621111', '621112']
)
WHERE code_type = 'NAICS' AND code IN ('621111', '621112');

UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6221',
    'parent_type', 'NAICS',
    'parent_name', 'General Medical and Surgical Hospitals',
    'child_codes', ARRAY['622110']
)
WHERE code_type = 'NAICS' AND code = '622110';

-- NAICS Hierarchies (Retail)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4521',
    'parent_type', 'NAICS',
    'parent_name', 'Department Stores',
    'child_codes', ARRAY['452111', '452112']
)
WHERE code_type = 'NAICS' AND code IN ('452111', '452112');

-- NAICS Hierarchies (Food & Beverage)
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '7225',
    'parent_type', 'NAICS',
    'parent_name', 'Restaurants and Other Eating Places',
    'child_codes', ARRAY['722511', '722513', '722515']
)
WHERE code_type = 'NAICS' AND code IN ('722511', '722513', '722515');

-- =====================================================
-- Part 6: Populate Industry Mappings
-- =====================================================

-- Technology Industry Mappings
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Technology',
    'secondary_industries', ARRAY['Software', 'IT Services', 'Computer Services']
)
WHERE code_type = 'NAICS' AND code IN ('541511', '541512', '541519')
   OR code_type = 'SIC' AND code IN ('7371', '7372', '7373')
   OR code_type = 'MCC' AND code IN ('5734', '5735');

-- Financial Services Industry Mappings
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Financial Services',
    'secondary_industries', ARRAY['Banking', 'Investment Services', 'Financial Institutions']
)
WHERE code_type = 'NAICS' AND code IN ('522110', '522120', '523110')
   OR code_type = 'SIC' AND code IN ('6021', '6022', '6211')
   OR code_type = 'MCC' AND code IN ('6010', '6011', '6012');

-- Healthcare Industry Mappings
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Healthcare',
    'secondary_industries', ARRAY['Medical Services', 'Hospital Services', 'Health Services']
)
WHERE code_type = 'NAICS' AND code IN ('621111', '621112', '622110')
   OR code_type = 'SIC' AND code IN ('8011', '8062')
   OR code_type = 'MCC' AND code IN ('8011', '8021', '8041', '8042', '8043', '8049', '8050');

-- Retail Industry Mappings
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Retail & Commerce',
    'secondary_industries', ARRAY['Department Stores', 'Retail Stores', 'General Merchandise']
)
WHERE code_type = 'NAICS' AND code IN ('452111', '452112')
   OR code_type = 'SIC' AND code = '5311'
   OR code_type = 'MCC' AND code IN ('5310', '5311');

-- Food & Beverage Industry Mappings
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Food & Beverage',
    'secondary_industries', ARRAY['Restaurants', 'Food Services', 'Beverage Services']
)
WHERE code_type = 'NAICS' AND code IN ('722511', '722513', '722515')
   OR code_type = 'SIC' AND code = '5812'
   OR code_type = 'MCC' AND code IN ('5811', '5812', '5813', '5814');

-- =====================================================
-- Verification Queries
-- =====================================================

-- Count total records
SELECT 
    'Total Records' AS metric,
    COUNT(*) AS count
FROM code_metadata;

-- Count by code type
SELECT 
    'Records by Code Type' AS metric,
    code_type,
    COUNT(*) AS count
FROM code_metadata
GROUP BY code_type
ORDER BY code_type;

-- Count with crosswalk data
SELECT 
    'Records with Crosswalk Data' AS metric,
    COUNT(*) AS count
FROM code_metadata
WHERE crosswalk_data != '{}'::jsonb;

-- Count with hierarchy data
SELECT 
    'Records with Hierarchy Data' AS metric,
    COUNT(*) AS count
FROM code_metadata
WHERE hierarchy != '{}'::jsonb;

-- Count with industry mappings
SELECT 
    'Records with Industry Mappings' AS metric,
    COUNT(*) AS count
FROM code_metadata
WHERE industry_mappings != '{}'::jsonb;

-- Sample crosswalk query
SELECT 
    'Sample Crosswalk' AS example,
    code_type,
    code,
    official_name,
    crosswalk_data
FROM code_metadata
WHERE crosswalk_data != '{}'::jsonb
LIMIT 3;

-- Sample hierarchy query
SELECT 
    'Sample Hierarchy' AS example,
    code_type,
    code,
    official_name,
    hierarchy->>'parent_code' AS parent_code,
    hierarchy->>'parent_name' AS parent_name
FROM code_metadata
WHERE hierarchy != '{}'::jsonb
LIMIT 3;

