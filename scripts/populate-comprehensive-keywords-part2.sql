-- KYB Platform - Comprehensive Keywords Part 2
-- This script continues populating keywords for remaining industries
-- Run this script AFTER running populate-comprehensive-classification-data.sql

-- =============================================================================
-- RETAIL & COMMERCE KEYWORDS
-- =============================================================================

INSERT INTO industry_keywords (industry_id, keyword, weight) 
SELECT i.id, k.keyword, k.weight
FROM industries i, (VALUES
    -- Online Retail
    ('online retail', 1.0), ('ecommerce', 1.0), ('e-commerce', 1.0), ('online store', 0.9),
    ('digital marketplace', 0.8), ('online shopping', 0.8), ('web store', 0.7), ('online sales', 0.7),
    ('digital retail', 0.7), ('online marketplace', 0.7), ('internet retail', 0.6), ('virtual store', 0.6),
    
    -- Physical Retail
    ('retail store', 1.0), ('brick and mortar', 0.9), ('physical store', 0.9), ('retail shop', 0.8),
    ('storefront', 0.8), ('retail location', 0.7), ('shopping center', 0.7), ('mall store', 0.6),
    ('retail chain', 0.6), ('department store', 0.6), ('boutique', 0.6), ('retail outlet', 0.6),
    
    -- Fashion & Apparel
    ('fashion', 1.0), ('apparel', 1.0), ('clothing', 0.9), ('fashion retail', 0.9),
    ('clothing store', 0.8), ('fashion boutique', 0.8), ('apparel store', 0.8), ('fashion brand', 0.7),
    ('clothing brand', 0.7), ('fashion design', 0.6), ('textile retail', 0.6), ('fashion accessories', 0.6),
    
    -- Electronics Retail
    ('electronics', 1.0), ('consumer electronics', 0.9), ('electronic devices', 0.8), ('tech retail', 0.8),
    ('computer store', 0.7), ('mobile devices', 0.7), ('audio equipment', 0.6), ('video equipment', 0.6),
    ('gaming equipment', 0.6), ('smart home', 0.6), ('electronic accessories', 0.6), ('tech gadgets', 0.6),
    
    -- Home & Garden
    ('home improvement', 1.0), ('garden center', 0.9), ('home decor', 0.8), ('furniture', 0.8),
    ('home goods', 0.8), ('garden supplies', 0.7), ('hardware store', 0.7), ('home renovation', 0.6),
    ('landscaping', 0.6), ('outdoor furniture', 0.6), ('home appliances', 0.6), ('garden tools', 0.6),
    
    -- Automotive Retail
    ('automotive', 1.0), ('car dealership', 0.9), ('auto sales', 0.9), ('vehicle sales', 0.8),
    ('car dealer', 0.8), ('automotive retail', 0.8), ('auto parts', 0.7), ('car service', 0.6),
    ('auto repair', 0.6), ('car maintenance', 0.6), ('automotive accessories', 0.6), ('vehicle financing', 0.6)
) AS k(keyword, weight)
WHERE i.name IN ('Online Retail', 'Physical Retail', 'Fashion & Apparel', 'Electronics Retail', 'Home & Garden', 'Automotive Retail')
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- =============================================================================
-- FOOD & BEVERAGE KEYWORDS
-- =============================================================================

INSERT INTO industry_keywords (industry_id, keyword, weight) 
SELECT i.id, k.keyword, k.weight
FROM industries i, (VALUES
    -- Restaurants
    ('restaurant', 1.0), ('food service', 0.9), ('dining', 0.8), ('restaurant business', 0.8),
    ('food establishment', 0.7), ('dining establishment', 0.7), ('restaurant chain', 0.6), ('fine dining', 0.6),
    ('casual dining', 0.6), ('fast casual', 0.6), ('restaurant management', 0.6), ('culinary', 0.6),
    
    -- Food Manufacturing
    ('food manufacturing', 1.0), ('food production', 0.9), ('food processing', 0.9), ('food company', 0.8),
    ('food factory', 0.7), ('food packaging', 0.7), ('food distribution', 0.6), ('food supply', 0.6),
    ('food ingredients', 0.6), ('food products', 0.6), ('food safety', 0.6), ('food quality', 0.6),
    
    -- Beverage Industry
    ('beverage', 1.0), ('beverage company', 0.9), ('beverage production', 0.8), ('drink manufacturing', 0.8),
    ('beverage distribution', 0.7), ('soft drinks', 0.7), ('alcoholic beverages', 0.7), ('beverage packaging', 0.6),
    ('beverage retail', 0.6), ('beverage service', 0.6), ('beverage industry', 0.6), ('drink company', 0.6),
    
    -- Catering Services
    ('catering', 1.0), ('catering services', 0.9), ('event catering', 0.8), ('catering company', 0.8),
    ('food catering', 0.7), ('catering business', 0.7), ('event food service', 0.6), ('catering management', 0.6),
    ('catering delivery', 0.6), ('catering planning', 0.6), ('catering equipment', 0.6), ('catering staff', 0.6),
    
    -- Food Delivery
    ('food delivery', 1.0), ('delivery service', 0.9), ('food takeout', 0.8), ('delivery app', 0.8),
    ('food courier', 0.7), ('delivery platform', 0.7), ('takeout service', 0.6), ('food logistics', 0.6),
    ('delivery management', 0.6), ('food transportation', 0.6), ('delivery tracking', 0.6), ('food ordering', 0.6)
) AS k(keyword, weight)
WHERE i.name IN ('Restaurants', 'Food Manufacturing', 'Beverage Industry', 'Catering Services', 'Food Delivery')
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- =============================================================================
-- MANUFACTURING & INDUSTRIAL KEYWORDS
-- =============================================================================

INSERT INTO industry_keywords (industry_id, keyword, weight) 
SELECT i.id, k.keyword, k.weight
FROM industries i, (VALUES
    -- Automotive Manufacturing
    ('automotive manufacturing', 1.0), ('car manufacturing', 0.9), ('vehicle production', 0.9), ('auto parts', 0.8),
    ('automotive industry', 0.8), ('car assembly', 0.7), ('automotive components', 0.7), ('vehicle manufacturing', 0.7),
    ('automotive engineering', 0.6), ('car production', 0.6), ('automotive supply', 0.6), ('vehicle assembly', 0.6),
    
    -- Electronics Manufacturing
    ('electronics manufacturing', 1.0), ('electronic components', 0.9), ('electronic devices', 0.8), ('electronics production', 0.8),
    ('electronic assembly', 0.7), ('electronics industry', 0.7), ('electronic manufacturing', 0.7), ('electronics factory', 0.6),
    ('electronic parts', 0.6), ('electronics supply', 0.6), ('electronic equipment', 0.6), ('electronics design', 0.6),
    
    -- Textile Manufacturing
    ('textile manufacturing', 1.0), ('fabric production', 0.9), ('textile industry', 0.8), ('textile production', 0.8),
    ('fabric manufacturing', 0.7), ('textile factory', 0.7), ('fabric processing', 0.6), ('textile processing', 0.6),
    ('fabric design', 0.6), ('textile design', 0.6), ('fabric supply', 0.6), ('textile supply', 0.6),
    
    -- Chemical Manufacturing
    ('chemical manufacturing', 1.0), ('chemical production', 0.9), ('chemical industry', 0.8), ('chemical processing', 0.8),
    ('chemical factory', 0.7), ('chemical products', 0.7), ('chemical engineering', 0.6), ('chemical supply', 0.6),
    ('chemical materials', 0.6), ('chemical compounds', 0.6), ('chemical research', 0.6), ('chemical development', 0.6),
    
    -- Aerospace Manufacturing
    ('aerospace manufacturing', 1.0), ('aircraft manufacturing', 0.9), ('aerospace industry', 0.8), ('aerospace production', 0.8),
    ('aircraft production', 0.7), ('aerospace components', 0.7), ('aerospace engineering', 0.6), ('aircraft assembly', 0.6),
    ('aerospace supply', 0.6), ('aerospace technology', 0.6), ('aircraft parts', 0.6), ('aerospace systems', 0.6)
) AS k(keyword, weight)
WHERE i.name IN ('Automotive Manufacturing', 'Electronics Manufacturing', 'Textile Manufacturing', 'Chemical Manufacturing', 'Aerospace Manufacturing')
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- =============================================================================
-- PROFESSIONAL SERVICES KEYWORDS
-- =============================================================================

INSERT INTO industry_keywords (industry_id, keyword, weight) 
SELECT i.id, k.keyword, k.weight
FROM industries i, (VALUES
    -- Legal Services
    ('legal services', 1.0), ('law firm', 0.9), ('attorney', 0.8), ('lawyer', 0.8), ('legal practice', 0.8),
    ('legal counsel', 0.7), ('legal advice', 0.7), ('legal representation', 0.6), ('legal consulting', 0.6),
    ('legal support', 0.6), ('legal assistance', 0.6), ('legal expertise', 0.6), ('legal services', 0.6),
    
    -- Accounting Services
    ('accounting services', 1.0), ('accounting firm', 0.9), ('accountant', 0.8), ('accounting practice', 0.8),
    ('financial accounting', 0.7), ('tax services', 0.7), ('bookkeeping', 0.6), ('financial consulting', 0.6),
    ('audit services', 0.6), ('accounting consulting', 0.6), ('financial reporting', 0.6), ('tax preparation', 0.6),
    
    -- Consulting
    ('consulting', 1.0), ('consulting services', 0.9), ('business consulting', 0.8), ('management consulting', 0.8),
    ('consulting firm', 0.7), ('business advisory', 0.7), ('strategic consulting', 0.6), ('consulting practice', 0.6),
    ('business strategy', 0.6), ('management advisory', 0.6), ('consulting expertise', 0.6), ('business development', 0.6),
    
    -- Marketing & Advertising
    ('marketing', 1.0), ('advertising', 1.0), ('marketing agency', 0.9), ('advertising agency', 0.9),
    ('marketing services', 0.8), ('advertising services', 0.8), ('digital marketing', 0.7), ('marketing strategy', 0.7),
    ('brand marketing', 0.6), ('marketing consulting', 0.6), ('advertising campaign', 0.6), ('marketing communications', 0.6),
    
    -- Real Estate Services
    ('real estate', 1.0), ('real estate services', 0.9), ('real estate agency', 0.8), ('property management', 0.8),
    ('real estate broker', 0.7), ('real estate agent', 0.7), ('property sales', 0.6), ('real estate consulting', 0.6),
    ('property development', 0.6), ('real estate investment', 0.6), ('property leasing', 0.6), ('real estate management', 0.6)
) AS k(keyword, weight)
WHERE i.name IN ('Legal Services', 'Accounting Services', 'Consulting', 'Marketing & Advertising', 'Real Estate Services')
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- =============================================================================
-- EDUCATION & TRAINING KEYWORDS
-- =============================================================================

INSERT INTO industry_keywords (industry_id, keyword, weight) 
SELECT i.id, k.keyword, k.weight
FROM industries i, (VALUES
    -- Higher Education
    ('higher education', 1.0), ('university', 0.9), ('college', 0.9), ('academic institution', 0.8),
    ('educational institution', 0.8), ('university system', 0.7), ('college system', 0.7), ('academic programs', 0.6),
    ('degree programs', 0.6), ('academic research', 0.6), ('higher learning', 0.6), ('academic excellence', 0.6),
    
    -- K-12 Education
    ('k-12 education', 1.0), ('primary education', 0.9), ('secondary education', 0.9), ('school system', 0.8),
    ('elementary school', 0.8), ('high school', 0.8), ('middle school', 0.7), ('public school', 0.7),
    ('private school', 0.7), ('educational services', 0.6), ('student services', 0.6), ('academic support', 0.6),
    
    -- Professional Training
    ('professional training', 1.0), ('corporate training', 0.9), ('training services', 0.8), ('professional development', 0.8),
    ('training programs', 0.7), ('skill development', 0.7), ('certification training', 0.6), ('training consulting', 0.6),
    ('employee training', 0.6), ('training solutions', 0.6), ('professional education', 0.6), ('training management', 0.6),
    
    -- Online Education
    ('online education', 1.0), ('e-learning', 0.9), ('online learning', 0.9), ('distance learning', 0.8),
    ('online courses', 0.8), ('digital education', 0.7), ('online training', 0.7), ('virtual learning', 0.6),
    ('online programs', 0.6), ('educational technology', 0.6), ('online platform', 0.6), ('digital learning', 0.6)
) AS k(keyword, weight)
WHERE i.name IN ('Higher Education', 'K-12 Education', 'Professional Training', 'Online Education')
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- =============================================================================
-- TRANSPORTATION & LOGISTICS KEYWORDS
-- =============================================================================

INSERT INTO industry_keywords (industry_id, keyword, weight) 
SELECT i.id, k.keyword, k.weight
FROM industries i, (VALUES
    -- Freight & Shipping
    ('freight', 1.0), ('shipping', 1.0), ('freight services', 0.9), ('shipping services', 0.9),
    ('logistics', 0.8), ('freight transportation', 0.8), ('cargo shipping', 0.7), ('freight forwarding', 0.7),
    ('shipping company', 0.6), ('freight company', 0.6), ('logistics services', 0.6), ('transportation services', 0.6),
    
    -- Passenger Transportation
    ('passenger transportation', 1.0), ('public transportation', 0.9), ('transportation services', 0.8), ('passenger services', 0.8),
    ('transit services', 0.7), ('transportation company', 0.7), ('passenger transport', 0.6), ('transportation system', 0.6),
    ('public transit', 0.6), ('transportation network', 0.6), ('passenger mobility', 0.6), ('transportation solutions', 0.6),
    
    -- Warehousing
    ('warehousing', 1.0), ('warehouse services', 0.9), ('storage services', 0.8), ('distribution center', 0.8),
    ('warehouse management', 0.7), ('storage facilities', 0.7), ('warehouse operations', 0.6), ('inventory management', 0.6),
    ('warehouse logistics', 0.6), ('storage solutions', 0.6), ('warehouse services', 0.6), ('distribution services', 0.6),
    
    -- Courier Services
    ('courier services', 1.0), ('package delivery', 0.9), ('delivery services', 0.8), ('courier company', 0.8),
    ('express delivery', 0.7), ('package shipping', 0.7), ('delivery company', 0.6), ('courier delivery', 0.6),
    ('package transport', 0.6), ('delivery logistics', 0.6), ('courier logistics', 0.6), ('express shipping', 0.6)
) AS k(keyword, weight)
WHERE i.name IN ('Freight & Shipping', 'Passenger Transportation', 'Warehousing', 'Courier Services')
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================

DO $$
BEGIN
    RAISE NOTICE 'KYB Platform Comprehensive Keywords Part 2 completed successfully!';
    RAISE NOTICE 'Added comprehensive keywords for Retail, Food & Beverage, Manufacturing, Professional Services, Education, and Transportation industries';
    RAISE NOTICE 'Total keyword coverage now includes all major industry sectors';
    RAISE NOTICE 'Ready for classification codes population';
END $$;
