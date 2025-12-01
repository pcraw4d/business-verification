-- =====================================================
-- Expand Code Metadata Table with Additional Codes
-- Purpose: Add more comprehensive code coverage across industries
-- OPTIMIZATION #6.2: Code Metadata Table Expansion
-- =====================================================

-- =====================================================
-- Part 1: Additional NAICS Codes
-- =====================================================

INSERT INTO code_metadata (code_type, code, official_name, official_description, is_official, is_active)
VALUES
-- Additional Technology Codes
('NAICS', '541330', 'Engineering Services', 
 'This industry comprises establishments primarily engaged in applying physical laws and principles of engineering in the design, development, and utilization of machines, materials, instruments, structures, processes, and systems.', 
 true, true),

('NAICS', '541611', 'Administrative Management and General Management Consulting Services', 
 'This industry comprises establishments primarily engaged in providing advice and assistance to businesses and other organizations on administrative management issues.', 
 true, true),

('NAICS', '541690', 'Other Scientific and Technical Consulting Services', 
 'This industry comprises establishments primarily engaged in providing advice and assistance to businesses and other organizations on scientific and technical issues (except environmental, computer systems design, and management consulting).', 
 true, true),

-- Additional Financial Services
('NAICS', '522210', 'Credit Card Issuing', 
 'This industry comprises establishments primarily engaged in issuing credit cards.', 
 true, true),

('NAICS', '522291', 'Consumer Lending', 
 'This industry comprises establishments primarily engaged in making unsecured cash loans to consumers.', 
 true, true),

('NAICS', '523120', 'Securities Brokerage', 
 'This industry comprises establishments primarily engaged in buying and selling securities on a commission or transaction fee basis.', 
 true, true),

('NAICS', '523130', 'Commodity Contracts Dealing', 
 'This industry comprises establishments primarily engaged in buying and selling commodity contracts on a commission or transaction fee basis.', 
 true, true),

-- Additional Healthcare
('NAICS', '621210', 'Offices of Dentists', 
 'This industry comprises establishments of licensed practitioners having the degree of D.M.D. (Doctor of Dental Medicine), D.D.S. (Doctor of Dental Surgery), or D.D. (Doctor of Dentistry) primarily engaged in the independent practice of general or specialized dentistry or dental surgery.', 
 true, true),

('NAICS', '621310', 'Offices of Chiropractors', 
 'This industry comprises establishments of licensed practitioners having the degree of D.C. (Doctor of Chiropractic) primarily engaged in the independent practice of chiropractic.', 
 true, true),

('NAICS', '621320', 'Offices of Optometrists', 
 'This industry comprises establishments of licensed practitioners having the degree of O.D. (Doctor of Optometry) primarily engaged in the independent practice of optometry.', 
 true, true),

('NAICS', '621340', 'Offices of Physical, Occupational and Speech Therapists, and Audiologists', 
 'This industry comprises establishments of licensed practitioners primarily engaged in the independent practice of physical therapy, occupational therapy, speech-language therapy, or audiology.', 
 true, true),

-- Additional Retail
('NAICS', '452210', 'Warehouse Clubs and Supercenters', 
 'This industry comprises establishments known as warehouse clubs, superstores, or supercenters primarily engaged in retailing a general line of groceries in combination with general lines of new merchandise, such as apparel, furniture, and appliances.', 
 true, true),

('NAICS', '452311', 'Warehouse Clubs and Supercenters', 
 'This industry comprises establishments known as warehouse clubs, superstores, or supercenters primarily engaged in retailing a general line of groceries in combination with general lines of new merchandise.', 
 true, true),

('NAICS', '453110', 'Florists', 
 'This industry comprises establishments primarily engaged in retailing cut flowers, floral arrangements, and potted plants purchased from others.', 
 true, true),

('NAICS', '453210', 'Office Supplies and Stationery Stores', 
 'This industry comprises establishments primarily engaged in retailing new office supplies, stationery, gift wrap, and gift boxes.', 
 true, true),

-- Additional Food & Beverage
('NAICS', '722410', 'Drinking Places (Alcoholic Beverages)', 
 'This industry comprises establishments primarily engaged in preparing and serving alcoholic beverages for immediate consumption on the premises.', 
 true, true),

('NAICS', '722320', 'Caterers', 
 'This industry comprises establishments primarily engaged in providing single event-based food services.', 
 true, true),

-- Manufacturing
('NAICS', '311811', 'Retail Bakeries', 
 'This industry comprises establishments primarily engaged in retailing bakery products not for immediate consumption made on the premises from flour and other ingredients.', 
 true, true),

('NAICS', '311812', 'Commercial Bakeries', 
 'This industry comprises establishments primarily engaged in manufacturing bread and other bakery products (except cookies and crackers) for wholesale distribution.', 
 true, true),

-- Construction
('NAICS', '236220', 'Commercial and Institutional Building Construction', 
 'This industry comprises establishments primarily engaged in the construction (including new work, additions, alterations, and repairs) of commercial and institutional buildings and related structures.', 
 true, true),

('NAICS', '237310', 'Highway, Street, and Bridge Construction', 
 'This industry comprises establishments primarily engaged in the construction of highways (including elevated), streets, roads, airport runways, public sidewalks, or bridges.', 
 true, true),

-- Real Estate
('NAICS', '531110', 'Lessors of Residential Buildings and Dwellings', 
 'This industry comprises establishments primarily engaged in acting as lessors of buildings used as residences or dwellings.', 
 true, true),

('NAICS', '531120', 'Lessors of Nonresidential Buildings (except Miniwarehouses)', 
 'This industry comprises establishments primarily engaged in acting as lessors of buildings (except miniwarehouses and self-storage units) that are not used as residences or dwellings.', 
 true, true),

('NAICS', '531210', 'Offices of Real Estate Agents and Brokers', 
 'This industry comprises establishments primarily engaged in acting as agents and/or brokers in one or more of the following: (1) selling real estate for others; (2) buying real estate for others; and (3) renting real estate for others.', 
 true, true),

-- Professional Services
('NAICS', '541110', 'Offices of Lawyers', 
 'This industry comprises establishments of legal practitioners known as lawyers or attorneys primarily engaged in the practice of law.', 
 true, true),

('NAICS', '541211', 'Offices of Certified Public Accountants', 
 'This industry comprises establishments of certified public accountants (CPAs) primarily engaged in providing accounting services.', 
 true, true),

('NAICS', '541310', 'Architectural Services', 
 'This industry comprises establishments primarily engaged in planning and designing residential, institutional, leisure, commercial, and industrial buildings and structures by applying knowledge of design, construction procedures, zoning regulations, building codes, and building materials.', 
 true, true),

-- Education
('NAICS', '611110', 'Elementary and Secondary Schools', 
 'This industry comprises establishments primarily engaged in furnishing academic courses and associated course work that comprise a basic preparatory education.', 
 true, true),

('NAICS', '611310', 'Colleges, Universities, and Professional Schools', 
 'This industry comprises establishments primarily engaged in furnishing academic courses and granting degrees at baccalaureate or graduate levels.', 
 true, true),

-- Transportation
('NAICS', '481111', 'Scheduled Passenger Air Transportation', 
 'This industry comprises establishments primarily engaged in providing air transportation of passengers over regular routes and on regular schedules.', 
 true, true),

('NAICS', '485110', 'Urban Transit Systems', 
 'This industry comprises establishments primarily engaged in operating local and suburban passenger transit systems over regular routes and on regular schedules.', 
 true, true),

-- Accommodation
('NAICS', '721110', 'Hotels (except Casino Hotels) and Motels', 
 'This industry comprises establishments primarily engaged in providing short-term lodging in facilities known as hotels, motor hotels, resort hotels, and motels.', 
 true, true),

('NAICS', '721120', 'Casino Hotels', 
 'This industry comprises establishments primarily engaged in providing short-term lodging in hotel facilities with a casino on the premises.', 
 true, true),

-- Arts and Entertainment
('NAICS', '711110', 'Theater Companies and Dinner Theaters', 
 'This industry comprises establishments primarily engaged in producing live theatrical presentations.', 
 true, true),

('NAICS', '713110', 'Amusement and Theme Parks', 
 'This industry comprises establishments primarily engaged in operating amusement and theme parks.', 
 true, true)

ON CONFLICT (code_type, code) DO UPDATE SET
    official_name = EXCLUDED.official_name,
    official_description = EXCLUDED.official_description,
    is_official = EXCLUDED.is_official,
    updated_at = NOW();

-- =====================================================
-- Part 2: Additional SIC Codes
-- =====================================================

INSERT INTO code_metadata (code_type, code, official_name, official_description, is_official, is_active)
VALUES
-- Additional Technology
('SIC', '7374', 'Computer Processing and Data Preparation Services', 
 'Establishments primarily engaged in providing computer processing and data preparation services.', 
 true, true),

('SIC', '7375', 'Information Retrieval Services', 
 'Establishments primarily engaged in providing computerized information retrieval services.', 
 true, true),

('SIC', '7376', 'Computer Facilities Management Services', 
 'Establishments primarily engaged in providing computer facilities management services.', 
 true, true),

-- Additional Financial Services
('SIC', '6029', 'Commercial Banks, Not Elsewhere Classified', 
 'Establishments primarily engaged in accepting deposits and making commercial, industrial, and consumer loans.', 
 true, true),

('SIC', '6035', 'Savings Institutions, Federally Chartered', 
 'Establishments primarily engaged in accepting time deposits and making mortgage and other loans.', 
 true, true),

('SIC', '6211', 'Security Brokers, Dealers, and Flotation Companies', 
 'Establishments primarily engaged in underwriting, originating, and/or maintaining markets for securities.', 
 true, true),

-- Additional Healthcare
('SIC', '8021', 'Offices and Clinics of Dentists', 
 'Establishments of licensed practitioners primarily engaged in the independent practice of general or specialized dentistry.', 
 true, true),

('SIC', '8041', 'Offices and Clinics of Chiropractors', 
 'Establishments of licensed practitioners primarily engaged in the independent practice of chiropractic.', 
 true, true),

('SIC', '8042', 'Offices and Clinics of Optometrists', 
 'Establishments of licensed practitioners primarily engaged in the independent practice of optometry.', 
 true, true),

-- Additional Retail
('SIC', '5331', 'Variety Stores', 
 'Establishments primarily engaged in retailing a general line of merchandise, including apparel, furniture, appliances, and food.', 
 true, true),

('SIC', '5411', 'Grocery Stores', 
 'Establishments primarily engaged in retailing a general line of food products.', 
 true, true),

('SIC', '5441', 'Candy, Nut, and Confectionery Stores', 
 'Establishments primarily engaged in retailing candy, nuts, and confectionery products.', 
 true, true),

-- Additional Food & Beverage
('SIC', '5813', 'Drinking Places (Alcoholic Beverages)', 
 'Establishments primarily engaged in serving alcoholic beverages for consumption on the premises.', 
 true, true),

-- Manufacturing
('SIC', '2051', 'Bread, Cake, and Related Products', 
 'Establishments primarily engaged in manufacturing bread, cake, and related products.', 
 true, true),

-- Construction
('SIC', '1521', 'General Contractors - Single-Family Houses', 
 'Establishments primarily engaged in the construction of single-family houses.', 
 true, true),

('SIC', '1522', 'General Contractors - Residential Buildings, Other Than Single-Family', 
 'Establishments primarily engaged in the construction of residential buildings other than single-family houses.', 
 true, true),

-- Real Estate
('SIC', '6512', 'Operators of Nonresidential Buildings', 
 'Establishments primarily engaged in operating nonresidential buildings.', 
 true, true),

('SIC', '6531', 'Real Estate Agents and Managers', 
 'Establishments primarily engaged in acting as agents and/or brokers in buying, selling, and renting real estate.', 
 true, true),

-- Professional Services
('SIC', '8111', 'Legal Services', 
 'Establishments primarily engaged in providing legal services.', 
 true, true),

('SIC', '8721', 'Accounting, Auditing, and Bookkeeping Services', 
 'Establishments primarily engaged in providing accounting, auditing, and bookkeeping services.', 
 true, true),

-- Education
('SIC', '8211', 'Elementary and Secondary Schools', 
 'Establishments primarily engaged in furnishing academic courses and associated course work.', 
 true, true),

('SIC', '8221', 'Colleges, Universities, and Professional Schools', 
 'Establishments primarily engaged in furnishing academic courses and granting degrees.', 
 true, true),

-- Transportation
('SIC', '4512', 'Air Transportation, Scheduled', 
 'Establishments primarily engaged in providing scheduled air transportation of passengers.', 
 true, true),

('SIC', '4111', 'Local and Suburban Transit', 
 'Establishments primarily engaged in operating local and suburban passenger transit systems.', 
 true, true),

-- Accommodation
('SIC', '7011', 'Hotels and Motels', 
 'Establishments primarily engaged in providing short-term lodging in facilities known as hotels, motor hotels, resort hotels, and motels.', 
 true, true),

-- Arts and Entertainment
('SIC', '7922', 'Theatrical Producers (except Motion Picture) and Ticket Agencies', 
 'Establishments primarily engaged in producing live theatrical presentations.', 
 true, true),

('SIC', '7996', 'Amusement Parks', 
 'Establishments primarily engaged in operating amusement and theme parks.', 
 true, true)

ON CONFLICT (code_type, code) DO UPDATE SET
    official_name = EXCLUDED.official_name,
    official_description = EXCLUDED.official_description,
    is_official = EXCLUDED.is_official,
    updated_at = NOW();

-- =====================================================
-- Part 3: Additional MCC Codes
-- =====================================================

INSERT INTO code_metadata (code_type, code, official_name, official_description, is_official, is_active)
VALUES
-- Additional Technology
('MCC', '5045', 'Computers, Computer Peripheral Equipment, Software', 
 'Merchants primarily engaged in retailing computers, computer peripheral equipment, and software.', 
 true, true),

('MCC', '5733', 'Radio, Television, and Consumer Electronics Stores', 
 'Merchants primarily engaged in retailing radio, television, and consumer electronics.', 
 true, true),

-- Additional Financial Services
('MCC', '6012', 'Financial Institutions - Merchandise, Services', 
 'Financial institutions providing merchandise and services.', 
 true, true),

-- Additional Healthcare
-- Note: MCC codes 8011, 8021, 8041, 8042, 8043, 8049, 8050 are already in populate_code_metadata.sql
-- They will be handled by ON CONFLICT if that script was run first

('MCC', '8062', 'Hospitals', 
 'Establishments primarily engaged in providing general medical and surgical services.', 
 true, true),

('MCC', '8071', 'Medical and Dental Laboratories', 
 'Laboratories providing medical and dental testing services.', 
 true, true),

-- Additional Retail
('MCC', '5411', 'Grocery Stores, Supermarkets', 
 'Merchants primarily engaged in retailing a general line of food products.', 
 true, true),

('MCC', '5441', 'Candy, Nut, and Confectionery Stores', 
 'Merchants primarily engaged in retailing candy, nuts, and confectionery products.', 
 true, true),

('MCC', '5451', 'Dairy Products Stores', 
 'Merchants primarily engaged in retailing dairy products.', 
 true, true),

('MCC', '5462', 'Bakeries', 
 'Merchants primarily engaged in retailing bakery products.', 
 true, true),

('MCC', '5499', 'Miscellaneous Food Stores - Convenience Stores, Markets, Specialty Stores', 
 'Merchants primarily engaged in retailing miscellaneous food products.', 
 true, true),

('MCC', '5611', 'Men''s and Women''s Clothing Stores', 
 'Merchants primarily engaged in retailing men''s and women''s clothing.', 
 true, true),

('MCC', '5621', 'Women''s Ready-to-Wear Stores', 
 'Merchants primarily engaged in retailing women''s ready-to-wear clothing.', 
 true, true),

('MCC', '5631', 'Women''s Accessory and Specialty Stores', 
 'Merchants primarily engaged in retailing women''s accessories and specialty items.', 
 true, true),

('MCC', '5641', 'Children''s and Infants'' Wear Stores', 
 'Merchants primarily engaged in retailing children''s and infants'' clothing.', 
 true, true),

('MCC', '5651', 'Family Clothing Stores', 
 'Merchants primarily engaged in retailing family clothing.', 
 true, true),

('MCC', '5661', 'Shoe Stores', 
 'Merchants primarily engaged in retailing shoes.', 
 true, true),

('MCC', '5691', 'Men''s and Women''s Clothing Stores', 
 'Merchants primarily engaged in retailing men''s and women''s clothing.', 
 true, true),

('MCC', '5712', 'Furniture, Home Furnishings, and Equipment Stores, Except Appliances', 
 'Merchants primarily engaged in retailing furniture, home furnishings, and equipment.', 
 true, true),

('MCC', '5713', 'Floor Covering Stores', 
 'Merchants primarily engaged in retailing floor coverings.', 
 true, true),

('MCC', '5714', 'Drapery, Window Covering, and Upholstery Stores', 
 'Merchants primarily engaged in retailing drapery, window coverings, and upholstery.', 
 true, true),

('MCC', '5718', 'Fireplace, Fireplace Screens, and Accessories Stores', 
 'Merchants primarily engaged in retailing fireplaces, fireplace screens, and accessories.', 
 true, true),

('MCC', '5719', 'Miscellaneous Home Furnishing Specialty Stores', 
 'Merchants primarily engaged in retailing miscellaneous home furnishings.', 
 true, true),

('MCC', '5722', 'Household Appliance Stores', 
 'Merchants primarily engaged in retailing household appliances.', 
 true, true),

('MCC', '5732', 'Electronics Stores', 
 'Merchants primarily engaged in retailing electronics.', 
 true, true),

-- Note: MCC codes 5734, 5735, 5811, 5813, 5814 are already in populate_code_metadata.sql
-- They will be handled by ON CONFLICT if that script was run first

-- Professional Services
('MCC', '8111', 'Legal Services', 
 'Merchants providing legal services.', 
 true, true),

('MCC', '8911', 'Architectural, Engineering, and Surveying Services', 
 'Merchants providing architectural, engineering, and surveying services.', 
 true, true),

('MCC', '8931', 'Accounting, Auditing, and Bookkeeping Services', 
 'Merchants providing accounting, auditing, and bookkeeping services.', 
 true, true),

-- Education
('MCC', '8211', 'Elementary and Secondary Schools', 
 'Establishments primarily engaged in furnishing academic courses and associated course work.', 
 true, true),

('MCC', '8220', 'Colleges, Universities, and Professional Schools', 
 'Establishments primarily engaged in furnishing academic courses and granting degrees.', 
 true, true),

-- Transportation
('MCC', '4111', 'Local and Suburban Passenger Transportation', 
 'Merchants providing local and suburban passenger transportation services.', 
 true, true),

('MCC', '4112', 'Passenger Railways', 
 'Merchants providing passenger railway transportation services.', 
 true, true),

('MCC', '4119', 'Ambulance Services', 
 'Merchants providing ambulance services.', 
 true, true),

('MCC', '4121', 'Taxicabs and Limousines', 
 'Merchants providing taxicab and limousine services.', 
 true, true),

('MCC', '4131', 'Bus Lines', 
 'Merchants providing bus line transportation services.', 
 true, true),

('MCC', '4511', 'Airlines, Air Carriers', 
 'Merchants providing airline transportation services.', 
 true, true),

-- Accommodation
('MCC', '7011', 'Hotels, Motels, and Resorts', 
 'Merchants providing short-term lodging in facilities known as hotels, motor hotels, resort hotels, and motels.', 
 true, true),

('MCC', '7012', 'Timeshares', 
 'Merchants providing timeshare lodging services.', 
 true, true),

-- Arts and Entertainment
('MCC', '7922', 'Theatrical Producers and Ticket Agencies', 
 'Merchants providing theatrical production and ticket agency services.', 
 true, true),

('MCC', '7929', 'Bands, Orchestras, and Miscellaneous Entertainers', 
 'Merchants providing entertainment services.', 
 true, true),

('MCC', '7933', 'Bowling Alleys', 
 'Merchants providing bowling alley services.', 
 true, true),

('MCC', '7941', 'Commercial Sports, Athletic Fields, Recreation, and Parks', 
 'Merchants providing commercial sports, athletic fields, recreation, and park services.', 
 true, true),

('MCC', '7991', 'Tourist Attractions and Exhibits', 
 'Merchants providing tourist attraction and exhibit services.', 
 true, true),

('MCC', '7992', 'Public Golf Courses', 
 'Merchants providing public golf course services.', 
 true, true),

('MCC', '7993', 'Video Amusement Game Supplies', 
 'Merchants providing video amusement game supplies.', 
 true, true),

('MCC', '7994', 'Video Game Arcades', 
 'Merchants providing video game arcade services.', 
 true, true),

('MCC', '7995', 'Betting (including Lottery Tickets, Casino Gaming Chips, Off-Track Betting, and Wagers)', 
 'Merchants providing betting services.', 
 true, true),

('MCC', '7996', 'Amusement Parks, Circuses, Carnivals, and Fortune Tellers', 
 'Merchants providing amusement park, circus, carnival, and fortune teller services.', 
 true, true),

('MCC', '7997', 'Membership Clubs (Sports, Recreation, Athletic), Country Clubs, and Private Golf Courses', 
 'Merchants providing membership club services.', 
 true, true),

('MCC', '7998', 'Aquariums, Seaquariums, Dolphinariums', 
 'Merchants providing aquarium, seaquarium, and dolphinarium services.', 
 true, true),

('MCC', '7999', 'Recreation Services, Not Elsewhere Classified', 
 'Merchants providing recreation services not elsewhere classified.', 
 true, true)

ON CONFLICT (code_type, code) DO UPDATE SET
    official_name = EXCLUDED.official_name,
    official_description = EXCLUDED.official_description,
    is_official = EXCLUDED.is_official,
    updated_at = NOW();

-- =====================================================
-- Part 4: Expand Crosswalk Data for New Codes
-- =====================================================

-- Technology Crosswalks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541330', '541511', '541512'],
    'sic', ARRAY['7371', '7372', '7373', '7374'],
    'mcc', ARRAY['5045', '5733', '5734']
)
WHERE code_type = 'NAICS' AND code = '541330';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541511', '541512'],
    'sic', ARRAY['7371', '7372', '7374', '7375'],
    'mcc', ARRAY['5045', '5734']
)
WHERE code_type = 'SIC' AND code IN ('7374', '7375', '7376');

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541511', '541512'],
    'sic', ARRAY['7371', '7372'],
    'mcc', ARRAY['5045', '5733', '5734']
)
WHERE code_type = 'MCC' AND code IN ('5045', '5733');

-- Healthcare Crosswalks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621210'],
    'sic', ARRAY['8021'],
    'mcc', ARRAY['8021']
)
WHERE code_type = 'NAICS' AND code = '621210';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621310'],
    'sic', ARRAY['8041'],
    'mcc', ARRAY['8041']
)
WHERE code_type = 'NAICS' AND code = '621310';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621320'],
    'sic', ARRAY['8042'],
    'mcc', ARRAY['8042', '8043']
)
WHERE code_type = 'NAICS' AND code = '621320';

-- Retail Crosswalks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['452210', '452311'],
    'sic', ARRAY['5331', '5411'],
    'mcc', ARRAY['5411', '5310', '5311']
)
WHERE code_type = 'NAICS' AND code IN ('452210', '452311');

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['453110'],
    'sic', ARRAY[]::text[],
    'mcc', ARRAY['5992']
)
WHERE code_type = 'NAICS' AND code = '453110';

-- Professional Services Crosswalks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541110'],
    'sic', ARRAY['8111'],
    'mcc', ARRAY['8111']
)
WHERE code_type = 'NAICS' AND code = '541110';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541211'],
    'sic', ARRAY['8721'],
    'mcc', ARRAY['8931']
)
WHERE code_type = 'NAICS' AND code = '541211';

-- Education Crosswalks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['611110'],
    'sic', ARRAY['8211'],
    'mcc', ARRAY['8211']
)
WHERE code_type = 'NAICS' AND code = '611110';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['611310'],
    'sic', ARRAY['8221'],
    'mcc', ARRAY['8220']
)
WHERE code_type = 'NAICS' AND code = '611310';

-- Transportation Crosswalks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['481111'],
    'sic', ARRAY['4512'],
    'mcc', ARRAY['4511']
)
WHERE code_type = 'NAICS' AND code = '481111';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['485110'],
    'sic', ARRAY['4111'],
    'mcc', ARRAY['4111', '4112', '4131']
)
WHERE code_type = 'NAICS' AND code = '485110';

-- Accommodation Crosswalks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['721110', '721120'],
    'sic', ARRAY['7011'],
    'mcc', ARRAY['7011', '7012']
)
WHERE code_type = 'NAICS' AND code IN ('721110', '721120');

-- Arts and Entertainment Crosswalks
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['711110'],
    'sic', ARRAY['7922'],
    'mcc', ARRAY['7922', '7929']
)
WHERE code_type = 'NAICS' AND code = '711110';

UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['713110'],
    'sic', ARRAY['7996'],
    'mcc', ARRAY['7996', '7991', '7992']
)
WHERE code_type = 'NAICS' AND code = '713110';

-- =====================================================
-- Part 5: Expand Industry Mappings
-- =====================================================

-- Professional Services Industry Mappings
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Professional, Scientific, and Technical Services',
    'secondary_industries', ARRAY['Legal Services', 'Accounting Services', 'Consulting Services']
)
WHERE code_type = 'NAICS' AND code IN ('541110', '541211', '541611')
   OR code_type = 'SIC' AND code IN ('8111', '8721')
   OR code_type = 'MCC' AND code IN ('8111', '8931');

-- Education Industry Mappings
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Education',
    'secondary_industries', ARRAY['Elementary Education', 'Secondary Education', 'Higher Education']
)
WHERE code_type = 'NAICS' AND code IN ('611110', '611310')
   OR code_type = 'SIC' AND code IN ('8211', '8221')
   OR code_type = 'MCC' AND code IN ('8211', '8220');

-- Transportation Industry Mappings
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Transportation and Warehousing',
    'secondary_industries', ARRAY['Air Transportation', 'Ground Transportation', 'Transit Services']
)
WHERE code_type = 'NAICS' AND code IN ('481111', '485110')
   OR code_type = 'SIC' AND code IN ('4512', '4111')
   OR code_type = 'MCC' AND code IN ('4511', '4111', '4112', '4131');

-- Accommodation Industry Mappings
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Accommodation and Food Services',
    'secondary_industries', ARRAY['Hotels', 'Motels', 'Resorts', 'Lodging']
)
WHERE code_type = 'NAICS' AND code IN ('721110', '721120')
   OR code_type = 'SIC' AND code = '7011'
   OR code_type = 'MCC' AND code IN ('7011', '7012');

-- Arts and Entertainment Industry Mappings
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Arts, Entertainment, and Recreation',
    'secondary_industries', ARRAY['Theater', 'Amusement Parks', 'Entertainment Services']
)
WHERE code_type = 'NAICS' AND code IN ('711110', '713110')
   OR code_type = 'SIC' AND code IN ('7922', '7996')
   OR code_type = 'MCC' AND code IN ('7922', '7929', '7996', '7991', '7992');

-- Construction Industry Mappings
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Construction',
    'secondary_industries', ARRAY['Building Construction', 'Infrastructure Construction', 'Construction Services']
)
WHERE code_type = 'NAICS' AND code IN ('236220', '237310')
   OR code_type = 'SIC' AND code IN ('1521', '1522');

-- Real Estate Industry Mappings
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Real Estate and Rental and Leasing',
    'secondary_industries', ARRAY['Property Management', 'Real Estate Services', 'Rental Services']
)
WHERE code_type = 'NAICS' AND code IN ('531110', '531120', '531210')
   OR code_type = 'SIC' AND code IN ('6512', '6531');

-- Manufacturing Industry Mappings
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Manufacturing',
    'secondary_industries', ARRAY['Food Manufacturing', 'Bakery Products', 'Food Processing']
)
WHERE code_type = 'NAICS' AND code IN ('311811', '311812')
   OR code_type = 'SIC' AND code = '2051';

-- =====================================================
-- Part 6: Expand Hierarchies for New Codes
-- =====================================================

-- Professional Services Hierarchies
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5411',
    'parent_type', 'NAICS',
    'parent_name', 'Legal Services',
    'child_codes', ARRAY['541110']
)
WHERE code_type = 'NAICS' AND code = '541110';

UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5412',
    'parent_type', 'NAICS',
    'parent_name', 'Accounting, Tax Preparation, Bookkeeping, and Payroll Services',
    'child_codes', ARRAY['541211']
)
WHERE code_type = 'NAICS' AND code = '541211';

-- Education Hierarchies
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6111',
    'parent_type', 'NAICS',
    'parent_name', 'Elementary and Secondary Schools',
    'child_codes', ARRAY['611110']
)
WHERE code_type = 'NAICS' AND code = '611110';

UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '6113',
    'parent_type', 'NAICS',
    'parent_name', 'Colleges, Universities, and Professional Schools',
    'child_codes', ARRAY['611310']
)
WHERE code_type = 'NAICS' AND code = '611310';

-- Transportation Hierarchies
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4811',
    'parent_type', 'NAICS',
    'parent_name', 'Scheduled Passenger Air Transportation',
    'child_codes', ARRAY['481111']
)
WHERE code_type = 'NAICS' AND code = '481111';

UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '4851',
    'parent_type', 'NAICS',
    'parent_name', 'Urban Transit Systems',
    'child_codes', ARRAY['485110']
)
WHERE code_type = 'NAICS' AND code = '485110';

-- Accommodation Hierarchies
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '7211',
    'parent_type', 'NAICS',
    'parent_name', 'Traveler Accommodation',
    'child_codes', ARRAY['721110', '721120']
)
WHERE code_type = 'NAICS' AND code IN ('721110', '721120');

-- Arts and Entertainment Hierarchies
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '7111',
    'parent_type', 'NAICS',
    'parent_name', 'Theater Companies and Dinner Theaters',
    'child_codes', ARRAY['711110']
)
WHERE code_type = 'NAICS' AND code = '711110';

UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '7131',
    'parent_type', 'NAICS',
    'parent_name', 'Amusement and Theme Parks',
    'child_codes', ARRAY['713110']
)
WHERE code_type = 'NAICS' AND code = '713110';

-- =====================================================
-- Verification Queries
-- =====================================================

-- Count total records after expansion
SELECT 
    'Total Records After Expansion' AS metric,
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

-- Sample expanded crosswalk query
SELECT 
    'Sample Expanded Crosswalk' AS example,
    code_type,
    code,
    official_name,
    jsonb_array_length(crosswalk_data->'naics') AS naics_count,
    jsonb_array_length(crosswalk_data->'sic') AS sic_count,
    jsonb_array_length(crosswalk_data->'mcc') AS mcc_count
FROM code_metadata
WHERE crosswalk_data != '{}'::jsonb
ORDER BY (jsonb_array_length(crosswalk_data->'naics') + 
          jsonb_array_length(crosswalk_data->'sic') + 
          jsonb_array_length(crosswalk_data->'mcc')) DESC
LIMIT 5;

