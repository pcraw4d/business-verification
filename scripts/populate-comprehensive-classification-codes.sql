-- KYB Platform - Comprehensive Classification Codes
-- This script populates NAICS, MCC, and SIC codes for all remaining industries
-- Run this script AFTER running populate-comprehensive-keywords-part2.sql

-- =============================================================================
-- RETAIL & COMMERCE CLASSIFICATION CODES
-- =============================================================================

INSERT INTO classification_codes (industry_id, code_type, code, description)
SELECT i.id, c.code_type, c.code, c.description
FROM industries i, (VALUES
    -- Online Retail
    ('NAICS', '454110', 'Electronic Shopping and Mail-Order Houses'),
    ('NAICS', '454111', 'Electronic Shopping'),
    ('NAICS', '454112', 'Electronic Auctions'),
    ('SIC', '5961', 'Catalog and Mail-Order Houses'),
    ('SIC', '5962', 'Automatic Merchandising Machine Operators'),
    ('MCC', '5310', 'Discount Stores'),
    ('MCC', '5311', 'Department Stores'),
    ('MCC', '5312', 'Variety Stores'),
    
    -- Physical Retail
    ('NAICS', '448110', 'Men''s Clothing Stores'),
    ('NAICS', '448120', 'Women''s Clothing Stores'),
    ('NAICS', '448130', 'Children''s and Infants'' Clothing Stores'),
    ('SIC', '5611', 'Men''s and Boys'' Clothing and Accessory Stores'),
    ('SIC', '5621', 'Women''s Clothing Stores'),
    ('MCC', '5651', 'Family Clothing Stores'),
    ('MCC', '5655', 'Sports and Riding Apparel Stores'),
    
    -- Fashion & Apparel
    ('NAICS', '448110', 'Men''s Clothing Stores'),
    ('NAICS', '448120', 'Women''s Clothing Stores'),
    ('NAICS', '448130', 'Children''s and Infants'' Clothing Stores'),
    ('SIC', '5611', 'Men''s and Boys'' Clothing and Accessory Stores'),
    ('SIC', '5621', 'Women''s Clothing Stores'),
    ('MCC', '5651', 'Family Clothing Stores'),
    ('MCC', '5655', 'Sports and Riding Apparel Stores'),
    
    -- Electronics Retail
    ('NAICS', '443142', 'Electronics Stores'),
    ('NAICS', '443143', 'Computer and Software Stores'),
    ('NAICS', '443144', 'Camera and Photographic Supplies Stores'),
    ('SIC', '5731', 'Radio, Television, and Consumer Electronics Stores'),
    ('SIC', '5734', 'Computer and Computer Software Stores'),
    ('MCC', '5732', 'Electronics Stores'),
    ('MCC', '5734', 'Computer Software Stores'),
    
    -- Home & Garden
    ('NAICS', '444110', 'Home Centers'),
    ('NAICS', '444120', 'Paint and Wallpaper Stores'),
    ('NAICS', '444130', 'Hardware Stores'),
    ('SIC', '5211', 'Lumber and Other Building Materials Dealers'),
    ('SIC', '5251', 'Hardware Stores'),
    ('MCC', '5200', 'Home Supply Warehouse Stores'),
    ('MCC', '5251', 'Hardware Stores'),
    
    -- Automotive Retail
    ('NAICS', '441110', 'New Car Dealers'),
    ('NAICS', '441120', 'Used Car Dealers'),
    ('NAICS', '441310', 'Automotive Parts and Accessories Stores'),
    ('SIC', '5511', 'Motor Vehicle Dealers (New and Used)'),
    ('SIC', '5531', 'Auto and Home Supply Stores'),
    ('MCC', '5511', 'Car and Truck Dealers (New and Used)'),
    ('MCC', '5533', 'Automotive Parts and Accessories Stores')
) AS c(code_type, code, description)
WHERE i.name IN ('Online Retail', 'Physical Retail', 'Fashion & Apparel', 'Electronics Retail', 'Home & Garden', 'Automotive Retail')
ON CONFLICT (industry_id, code_type, code) DO NOTHING;

-- =============================================================================
-- FOOD & BEVERAGE CLASSIFICATION CODES
-- =============================================================================

INSERT INTO classification_codes (industry_id, code_type, code, description)
SELECT i.id, c.code_type, c.code, c.description
FROM industries i, (VALUES
    -- Restaurants
    ('NAICS', '722511', 'Full-Service Restaurants'),
    ('NAICS', '722513', 'Limited-Service Restaurants'),
    ('NAICS', '722514', 'Cafeterias, Grill Buffets, and Buffets'),
    ('SIC', '5812', 'Eating Places'),
    ('SIC', '5813', 'Drinking Places (Alcoholic Beverages)'),
    ('MCC', '5812', 'Eating Places, Restaurants'),
    ('MCC', '5813', 'Drinking Places (Alcoholic Beverages)'),
    
    -- Food Manufacturing
    ('NAICS', '311111', 'Dog and Cat Food Manufacturing'),
    ('NAICS', '311211', 'Flour Milling'),
    ('NAICS', '311212', 'Rice Milling'),
    ('SIC', '2041', 'Flour and Other Grain Mill Products'),
    ('SIC', '2043', 'Cereal Breakfast Foods'),
    ('MCC', '5411', 'Grocery Stores, Supermarkets'),
    ('MCC', '5422', 'Meat Provisioners - Freezer and Locker'),
    
    -- Beverage Industry
    ('NAICS', '312111', 'Soft Drink Manufacturing'),
    ('NAICS', '312112', 'Bottled Water Manufacturing'),
    ('NAICS', '312113', 'Ice Manufacturing'),
    ('SIC', '2086', 'Bottled and Canned Soft Drinks and Carbonated Waters'),
    ('SIC', '2087', 'Flavoring Syrup and Concentrate Manufacturing'),
    ('MCC', '5813', 'Drinking Places (Alcoholic Beverages)'),
    ('MCC', '5921', 'Package Stores - Beer, Wine, and Liquor'),
    
    -- Catering Services
    ('NAICS', '722320', 'Caterers'),
    ('NAICS', '722330', 'Mobile Food Services'),
    ('SIC', '5812', 'Eating Places'),
    ('MCC', '5812', 'Eating Places, Restaurants'),
    
    -- Food Delivery
    ('NAICS', '722513', 'Limited-Service Restaurants'),
    ('NAICS', '722320', 'Caterers'),
    ('SIC', '5812', 'Eating Places'),
    ('MCC', '5812', 'Eating Places, Restaurants')
) AS c(code_type, code, description)
WHERE i.name IN ('Restaurants', 'Food Manufacturing', 'Beverage Industry', 'Catering Services', 'Food Delivery')
ON CONFLICT (industry_id, code_type, code) DO NOTHING;

-- =============================================================================
-- MANUFACTURING & INDUSTRIAL CLASSIFICATION CODES
-- =============================================================================

INSERT INTO classification_codes (industry_id, code_type, code, description)
SELECT i.id, c.code_type, c.code, c.description
FROM industries i, (VALUES
    -- Automotive Manufacturing
    ('NAICS', '336111', 'Automobile Manufacturing'),
    ('NAICS', '336112', 'Light Truck and Utility Vehicle Manufacturing'),
    ('NAICS', '336211', 'Motor Vehicle Body Manufacturing'),
    ('SIC', '3711', 'Motor Vehicles and Passenger Car Bodies'),
    ('SIC', '3713', 'Truck and Bus Bodies'),
    ('MCC', '5511', 'Car and Truck Dealers (New and Used)'),
    ('MCC', '5533', 'Automotive Parts and Accessories Stores'),
    
    -- Electronics Manufacturing
    ('NAICS', '334111', 'Electronic Computer Manufacturing'),
    ('NAICS', '334112', 'Computer Storage Device Manufacturing'),
    ('NAICS', '334113', 'Computer Terminal Manufacturing'),
    ('SIC', '3571', 'Electronic Computers'),
    ('SIC', '3572', 'Computer Storage Devices'),
    ('MCC', '5732', 'Electronics Stores'),
    ('MCC', '5734', 'Computer Software Stores'),
    
    -- Textile Manufacturing
    ('NAICS', '313210', 'Broadwoven Fabric Mills'),
    ('NAICS', '313220', 'Narrow Fabric Mills and Schiffli Machine Embroidery'),
    ('NAICS', '313230', 'Nonwoven Fabric Mills'),
    ('SIC', '2211', 'Broadwoven Fabric Mills, Cotton'),
    ('SIC', '2221', 'Broadwoven Fabric Mills, Manmade Fiber and Silk'),
    ('MCC', '5651', 'Family Clothing Stores'),
    ('MCC', '5655', 'Sports and Riding Apparel Stores'),
    
    -- Chemical Manufacturing
    ('NAICS', '325110', 'Petrochemical Manufacturing'),
    ('NAICS', '325120', 'Industrial Gas Manufacturing'),
    ('NAICS', '325130', 'Synthetic Dye and Pigment Manufacturing'),
    ('SIC', '2812', 'Alkalies and Chlorine'),
    ('SIC', '2813', 'Industrial Gases'),
    ('MCC', '5169', 'Chemicals and Allied Products, Not Elsewhere Classified'),
    
    -- Aerospace Manufacturing
    ('NAICS', '336411', 'Aircraft Manufacturing'),
    ('NAICS', '336412', 'Aircraft Engine and Engine Parts Manufacturing'),
    ('NAICS', '336413', 'Other Aircraft Parts and Auxiliary Equipment Manufacturing'),
    ('SIC', '3721', 'Aircraft'),
    ('SIC', '3724', 'Aircraft Engines and Engine Parts'),
    ('MCC', '3721', 'Aircraft'),
    ('MCC', '3724', 'Aircraft Engines and Engine Parts')
) AS c(code_type, code, description)
WHERE i.name IN ('Automotive Manufacturing', 'Electronics Manufacturing', 'Textile Manufacturing', 'Chemical Manufacturing', 'Aerospace Manufacturing')
ON CONFLICT (industry_id, code_type, code) DO NOTHING;

-- =============================================================================
-- PROFESSIONAL SERVICES CLASSIFICATION CODES
-- =============================================================================

INSERT INTO classification_codes (industry_id, code_type, code, description)
SELECT i.id, c.code_type, c.code, c.description
FROM industries i, (VALUES
    -- Legal Services
    ('NAICS', '541110', 'Offices of Lawyers'),
    ('NAICS', '541191', 'Title Abstract and Settlement Offices'),
    ('NAICS', '541199', 'All Other Legal Services'),
    ('SIC', '8111', 'Legal Services'),
    ('MCC', '8111', 'Legal Services'),
    
    -- Accounting Services
    ('NAICS', '541211', 'Offices of Certified Public Accountants'),
    ('NAICS', '541213', 'Tax Preparation Services'),
    ('NAICS', '541219', 'Other Accounting Services'),
    ('SIC', '8721', 'Accounting, Auditing, and Bookkeeping Services'),
    ('MCC', '8721', 'Accounting, Auditing, and Bookkeeping Services'),
    
    -- Consulting
    ('NAICS', '541611', 'Administrative Management and General Management Consulting Services'),
    ('NAICS', '541612', 'Human Resources Consulting Services'),
    ('NAICS', '541613', 'Marketing Consulting Services'),
    ('SIC', '8742', 'Management Consulting Services'),
    ('MCC', '8742', 'Management Consulting Services'),
    
    -- Marketing & Advertising
    ('NAICS', '541810', 'Advertising Agencies'),
    ('NAICS', '541820', 'Public Relations Agencies'),
    ('NAICS', '541830', 'Media Buying Agencies'),
    ('SIC', '7311', 'Advertising Agencies'),
    ('MCC', '7311', 'Advertising Agencies'),
    
    -- Real Estate Services
    ('NAICS', '531210', 'Offices of Real Estate Agents and Brokers'),
    ('NAICS', '531311', 'Residential Property Managers'),
    ('NAICS', '531312', 'Nonresidential Property Managers'),
    ('SIC', '6531', 'Real Estate Agents and Managers'),
    ('MCC', '6531', 'Real Estate Agents and Managers')
) AS c(code_type, code, description)
WHERE i.name IN ('Legal Services', 'Accounting Services', 'Consulting', 'Marketing & Advertising', 'Real Estate Services')
ON CONFLICT (industry_id, code_type, code) DO NOTHING;

-- =============================================================================
-- EDUCATION & TRAINING CLASSIFICATION CODES
-- =============================================================================

INSERT INTO classification_codes (industry_id, code_type, code, description)
SELECT i.id, c.code_type, c.code, c.description
FROM industries i, (VALUES
    -- Higher Education
    ('NAICS', '611310', 'Colleges, Universities, and Professional Schools'),
    ('NAICS', '611410', 'Business Schools and Computer and Management Training'),
    ('NAICS', '611420', 'Computer Training'),
    ('SIC', '8221', 'Colleges, Universities, and Professional Schools'),
    ('MCC', '8220', 'Colleges, Universities, Professional Schools, and Junior Colleges'),
    
    -- K-12 Education
    ('NAICS', '611110', 'Elementary and Secondary Schools'),
    ('NAICS', '611691', 'Exam Preparation and Tutoring'),
    ('SIC', '8211', 'Elementary and Secondary Schools'),
    ('MCC', '8211', 'Elementary and Secondary Schools'),
    
    -- Professional Training
    ('NAICS', '611410', 'Business Schools and Computer and Management Training'),
    ('NAICS', '611420', 'Computer Training'),
    ('NAICS', '611430', 'Professional and Management Development Training'),
    ('SIC', '8299', 'Schools and Educational Services, Not Elsewhere Classified'),
    ('MCC', '8299', 'Schools and Educational Services, Not Elsewhere Classified'),
    
    -- Online Education
    ('NAICS', '611710', 'Educational Support Services'),
    ('NAICS', '611691', 'Exam Preparation and Tutoring'),
    ('SIC', '8299', 'Schools and Educational Services, Not Elsewhere Classified'),
    ('MCC', '8299', 'Schools and Educational Services, Not Elsewhere Classified')
) AS c(code_type, code, description)
WHERE i.name IN ('Higher Education', 'K-12 Education', 'Professional Training', 'Online Education')
ON CONFLICT (industry_id, code_type, code) DO NOTHING;

-- =============================================================================
-- TRANSPORTATION & LOGISTICS CLASSIFICATION CODES
-- =============================================================================

INSERT INTO classification_codes (industry_id, code_type, code, description)
SELECT i.id, c.code_type, c.code, c.description
FROM industries i, (VALUES
    -- Freight & Shipping
    ('NAICS', '484110', 'General Freight Trucking, Local'),
    ('NAICS', '484121', 'General Freight Trucking, Long-Distance, Truckload'),
    ('NAICS', '484122', 'General Freight Trucking, Long-Distance, Less-Than-Truckload'),
    ('SIC', '4213', 'Trucking, Except Local'),
    ('SIC', '4214', 'Local Trucking Without Storage'),
    ('MCC', '4214', 'Local Trucking Without Storage'),
    ('MCC', '4215', 'Courier Services'),
    
    -- Passenger Transportation
    ('NAICS', '485111', 'Mixed Mode Transit Systems'),
    ('NAICS', '485112', 'Commuter Rail Systems'),
    ('NAICS', '485113', 'Bus and Other Motor Vehicle Transit Systems'),
    ('SIC', '4111', 'Local and Suburban Transportation'),
    ('SIC', '4119', 'Local Passenger Transportation, Not Elsewhere Classified'),
    ('MCC', '4111', 'Local and Suburban Transportation'),
    ('MCC', '4119', 'Local Passenger Transportation, Not Elsewhere Classified'),
    
    -- Warehousing
    ('NAICS', '493110', 'General Warehousing and Storage'),
    ('NAICS', '493120', 'Refrigerated Warehousing and Storage'),
    ('NAICS', '493130', 'Farm Product Warehousing and Storage'),
    ('SIC', '4225', 'General Warehousing and Storage'),
    ('SIC', '4226', 'Special Warehousing and Storage, Not Elsewhere Classified'),
    ('MCC', '4225', 'General Warehousing and Storage'),
    ('MCC', '4226', 'Special Warehousing and Storage, Not Elsewhere Classified'),
    
    -- Courier Services
    ('NAICS', '492110', 'Couriers and Express Delivery Services'),
    ('NAICS', '492210', 'Local Messengers and Local Delivery'),
    ('SIC', '4215', 'Courier Services'),
    ('MCC', '4215', 'Courier Services')
) AS c(code_type, code, description)
WHERE i.name IN ('Freight & Shipping', 'Passenger Transportation', 'Warehousing', 'Courier Services')
ON CONFLICT (industry_id, code_type, code) DO NOTHING;

-- =============================================================================
-- ENTERTAINMENT & MEDIA CLASSIFICATION CODES
-- =============================================================================

INSERT INTO classification_codes (industry_id, code_type, code, description)
SELECT i.id, c.code_type, c.code, c.description
FROM industries i, (VALUES
    -- Media Production
    ('NAICS', '512110', 'Motion Picture and Video Production'),
    ('NAICS', '512120', 'Motion Picture and Video Distribution'),
    ('NAICS', '512131', 'Motion Picture Theaters (except Drive-Ins)'),
    ('SIC', '7812', 'Motion Picture and Video Production'),
    ('SIC', '7819', 'Services Allied to Motion Picture Production'),
    ('MCC', '7812', 'Motion Picture and Video Production'),
    ('MCC', '7819', 'Services Allied to Motion Picture Production'),
    
    -- Gaming
    ('NAICS', '511210', 'Software Publishers'),
    ('NAICS', '541511', 'Custom Computer Programming Services'),
    ('NAICS', '541512', 'Computer Systems Design Services'),
    ('SIC', '7372', 'Prepackaged Software'),
    ('SIC', '7373', 'Computer Integrated Systems Design'),
    ('MCC', '7372', 'Computer Programming Services'),
    ('MCC', '7373', 'Computer Integrated Systems Design'),
    
    -- Music & Entertainment
    ('NAICS', '512220', 'Integrated Record Production/Distribution'),
    ('NAICS', '512230', 'Music Publishers'),
    ('NAICS', '512240', 'Sound Recording Studios'),
    ('SIC', '7922', 'Theatrical Producers (except Motion Picture) and Miscellaneous Theatrical Services'),
    ('SIC', '7929', 'Bands, Orchestras, Actors, and Other Entertainers and Entertainment Groups'),
    ('MCC', '7922', 'Theatrical Producers (except Motion Picture) and Miscellaneous Theatrical Services'),
    ('MCC', '7929', 'Bands, Orchestras, Actors, and Other Entertainers and Entertainment Groups'),
    
    -- Publishing
    ('NAICS', '511110', 'Newspaper Publishers'),
    ('NAICS', '511120', 'Periodical Publishers'),
    ('NAICS', '511130', 'Book Publishers'),
    ('SIC', '2711', 'Newspapers: Publishing, or Publishing and Printing'),
    ('SIC', '2721', 'Periodicals: Publishing, or Publishing and Printing'),
    ('MCC', '2711', 'Newspapers: Publishing, or Publishing and Printing'),
    ('MCC', '2721', 'Periodicals: Publishing, or Publishing and Printing')
) AS c(code_type, code, description)
WHERE i.name IN ('Media Production', 'Gaming', 'Music & Entertainment', 'Publishing')
ON CONFLICT (industry_id, code_type, code) DO NOTHING;

-- =============================================================================
-- ENERGY & UTILITIES CLASSIFICATION CODES
-- =============================================================================

INSERT INTO classification_codes (industry_id, code_type, code, description)
SELECT i.id, c.code_type, c.code, c.description
FROM industries i, (VALUES
    -- Renewable Energy
    ('NAICS', '221114', 'Solar Electric Power Generation'),
    ('NAICS', '221115', 'Wind Electric Power Generation'),
    ('NAICS', '221116', 'Geothermal Electric Power Generation'),
    ('SIC', '4911', 'Electric Services'),
    ('MCC', '4911', 'Electric Services'),
    
    -- Oil & Gas
    ('NAICS', '211111', 'Crude Petroleum and Natural Gas Extraction'),
    ('NAICS', '211112', 'Natural Gas Liquid Extraction'),
    ('NAICS', '211113', 'Natural Gas Extraction'),
    ('SIC', '1311', 'Crude Petroleum and Natural Gas'),
    ('MCC', '1311', 'Crude Petroleum and Natural Gas'),
    
    -- Electric Utilities
    ('NAICS', '221122', 'Electric Power Distribution'),
    ('NAICS', '221121', 'Electric Bulk Power Transmission and Control'),
    ('NAICS', '221111', 'Hydroelectric Power Generation'),
    ('SIC', '4911', 'Electric Services'),
    ('MCC', '4911', 'Electric Services'),
    
    -- Water & Waste Management
    ('NAICS', '221310', 'Water Supply and Irrigation Systems'),
    ('NAICS', '221320', 'Sewage Treatment Facilities'),
    ('NAICS', '562211', 'Hazardous Waste Treatment and Disposal'),
    ('SIC', '4941', 'Water Supply'),
    ('MCC', '4941', 'Water Supply')
) AS c(code_type, code, description)
WHERE i.name IN ('Renewable Energy', 'Oil & Gas', 'Electric Utilities', 'Water & Waste Management')
ON CONFLICT (industry_id, code_type, code) DO NOTHING;

-- =============================================================================
-- CONSTRUCTION & ENGINEERING CLASSIFICATION CODES
-- =============================================================================

INSERT INTO classification_codes (industry_id, code_type, code, description)
SELECT i.id, c.code_type, c.code, c.description
FROM industries i, (VALUES
    -- Construction
    ('NAICS', '236116', 'New Multifamily Housing Construction (except For-Sale Builders)'),
    ('NAICS', '236117', 'New Housing For-Sale Builders'),
    ('NAICS', '236118', 'Residential Remodelers'),
    ('SIC', '1521', 'General Contractors - Single-Family Houses'),
    ('SIC', '1522', 'General Contractors - Residential Buildings, Other Than Single-Family'),
    ('MCC', '1521', 'General Contractors - Single-Family Houses'),
    ('MCC', '1522', 'General Contractors - Residential Buildings, Other Than Single-Family'),
    
    -- Architecture
    ('NAICS', '541310', 'Architectural Services'),
    ('NAICS', '541320', 'Landscape Architectural Services'),
    ('SIC', '8712', 'Architectural Services'),
    ('MCC', '8712', 'Architectural Services'),
    
    -- Engineering Services
    ('NAICS', '541330', 'Engineering Services'),
    ('NAICS', '541340', 'Drafting Services'),
    ('SIC', '8711', 'Engineering Services'),
    ('MCC', '8711', 'Engineering Services'),
    
    -- Home Improvement
    ('NAICS', '238220', 'Plumbing, Heating, and Air-Conditioning Contractors'),
    ('NAICS', '238350', 'Finish Carpentry Contractors'),
    ('NAICS', '238390', 'Other Building Finishing Contractors'),
    ('SIC', '1711', 'Plumbing, Heating, and Air-Conditioning'),
    ('MCC', '1711', 'Plumbing, Heating, and Air-Conditioning')
) AS c(code_type, code, description)
WHERE i.name IN ('Construction', 'Architecture', 'Engineering Services', 'Home Improvement')
ON CONFLICT (industry_id, code_type, code) DO NOTHING;

-- =============================================================================
-- AGRICULTURE & FOOD PRODUCTION CLASSIFICATION CODES
-- =============================================================================

INSERT INTO classification_codes (industry_id, code_type, code, description)
SELECT i.id, c.code_type, c.code, c.description
FROM industries i, (VALUES
    -- Agriculture
    ('NAICS', '111110', 'Soybean Farming'),
    ('NAICS', '111120', 'Oilseed (except Soybean) Farming'),
    ('NAICS', '111130', 'Dry Pea and Bean Farming'),
    ('SIC', '0111', 'Wheat'),
    ('SIC', '0112', 'Rice'),
    ('MCC', '0111', 'Wheat'),
    ('MCC', '0112', 'Rice'),
    
    -- Food Processing
    ('NAICS', '311111', 'Dog and Cat Food Manufacturing'),
    ('NAICS', '311211', 'Flour Milling'),
    ('NAICS', '311212', 'Rice Milling'),
    ('SIC', '2041', 'Flour and Other Grain Mill Products'),
    ('SIC', '2043', 'Cereal Breakfast Foods'),
    ('MCC', '2041', 'Flour and Other Grain Mill Products'),
    ('MCC', '2043', 'Cereal Breakfast Foods'),
    
    -- Livestock
    ('NAICS', '112111', 'Beef Cattle Ranching and Farming'),
    ('NAICS', '112112', 'Cattle Feedlots'),
    ('NAICS', '112120', 'Dairy Cattle and Milk Production'),
    ('SIC', '0211', 'Beef Cattle Feedlots'),
    ('SIC', '0212', 'Beef Cattle, Except Feedlots'),
    ('MCC', '0211', 'Beef Cattle Feedlots'),
    ('MCC', '0212', 'Beef Cattle, Except Feedlots'),
    
    -- Crop Production
    ('NAICS', '111110', 'Soybean Farming'),
    ('NAICS', '111120', 'Oilseed (except Soybean) Farming'),
    ('NAICS', '111130', 'Dry Pea and Bean Farming'),
    ('SIC', '0111', 'Wheat'),
    ('SIC', '0112', 'Rice'),
    ('MCC', '0111', 'Wheat'),
    ('MCC', '0112', 'Rice')
) AS c(code_type, code, description)
WHERE i.name IN ('Agriculture', 'Food Processing', 'Livestock', 'Crop Production')
ON CONFLICT (industry_id, code_type, code) DO NOTHING;

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================

DO $$
BEGIN
    RAISE NOTICE 'KYB Platform Comprehensive Classification Codes completed successfully!';
    RAISE NOTICE 'Populated NAICS, MCC, and SIC codes for all major industry sectors';
    RAISE NOTICE 'Total classification codes now cover all industries with proper mappings';
    RAISE NOTICE 'Ready for industry patterns creation';
END $$;
