-- =============================================================================
-- RETAIL & E-COMMERCE KEYWORDS ADDITION SCRIPT
-- Task 3.2.4: Add retail and e-commerce keywords
-- =============================================================================
-- This script adds comprehensive keyword sets for retail and e-commerce industries
-- to achieve >85% classification accuracy for retail businesses.
-- 
-- Industries covered:
-- 1. Retail (traditional retail stores, brick-and-mortar retail)
-- 2. E-commerce (online retail, e-commerce platforms)
-- 3. Wholesale (wholesale trade, distribution, B2B sales)
-- 4. Consumer Goods (consumer goods manufacturing, retail, distribution)
--
-- Total keywords: 200+ across 4 industries
-- Base weights: 0.5000-1.0000 as specified in plan
-- =============================================================================

-- Start transaction for atomic operation
BEGIN;

-- =============================================================================
-- 1. RETAIL INDUSTRY KEYWORDS (50+ keywords)
-- =============================================================================

-- Get Retail industry ID
DO $$
DECLARE
    retail_industry_id INTEGER;
BEGIN
    SELECT id INTO retail_industry_id 
    FROM industries 
    WHERE name = 'Retail' AND is_active = true;
    
    IF retail_industry_id IS NULL THEN
        RAISE EXCEPTION 'Retail industry not found. Please run add-comprehensive-industries.sql first.';
    END IF;
    
    -- Insert retail keywords with appropriate weights
    INSERT INTO keyword_weights (industry_id, keyword, base_weight, is_active, created_at, updated_at) VALUES
    -- Core retail terms (weight 1.0000)
    (retail_industry_id, 'retail', 1.0000, true, NOW(), NOW()),
    (retail_industry_id, 'store', 1.0000, true, NOW(), NOW()),
    (retail_industry_id, 'shop', 1.0000, true, NOW(), NOW()),
    (retail_industry_id, 'shopping', 1.0000, true, NOW(), NOW()),
    (retail_industry_id, 'merchandise', 1.0000, true, NOW(), NOW()),
    (retail_industry_id, 'inventory', 1.0000, true, NOW(), NOW()),
    (retail_industry_id, 'sales', 1.0000, true, NOW(), NOW()),
    (retail_industry_id, 'customer', 1.0000, true, NOW(), NOW()),
    (retail_industry_id, 'brick and mortar', 1.0000, true, NOW(), NOW()),
    (retail_industry_id, 'physical store', 1.0000, true, NOW(), NOW()),
    
    -- Retail operations (weight 0.9000)
    (retail_industry_id, 'retailer', 0.9000, true, NOW(), NOW()),
    (retail_industry_id, 'retail store', 0.9000, true, NOW(), NOW()),
    (retail_industry_id, 'retail chain', 0.9000, true, NOW(), NOW()),
    (retail_industry_id, 'retail outlet', 0.9000, true, NOW(), NOW()),
    (retail_industry_id, 'retail location', 0.9000, true, NOW(), NOW()),
    (retail_industry_id, 'retail business', 0.9000, true, NOW(), NOW()),
    (retail_industry_id, 'retail operations', 0.9000, true, NOW(), NOW()),
    (retail_industry_id, 'retail management', 0.9000, true, NOW(), NOW()),
    (retail_industry_id, 'retail sales', 0.9000, true, NOW(), NOW()),
    (retail_industry_id, 'retail market', 0.9000, true, NOW(), NOW()),
    
    -- Store types (weight 0.8000)
    (retail_industry_id, 'department store', 0.8000, true, NOW(), NOW()),
    (retail_industry_id, 'specialty store', 0.8000, true, NOW(), NOW()),
    (retail_industry_id, 'boutique', 0.8000, true, NOW(), NOW()),
    (retail_industry_id, 'supermarket', 0.8000, true, NOW(), NOW()),
    (retail_industry_id, 'grocery store', 0.8000, true, NOW(), NOW()),
    (retail_industry_id, 'convenience store', 0.8000, true, NOW(), NOW()),
    (retail_industry_id, 'discount store', 0.8000, true, NOW(), NOW()),
    (retail_industry_id, 'outlet store', 0.8000, true, NOW(), NOW()),
    (retail_industry_id, 'flagship store', 0.8000, true, NOW(), NOW()),
    (retail_industry_id, 'pop-up store', 0.8000, true, NOW(), NOW()),
    
    -- Retail activities (weight 0.7000)
    (retail_industry_id, 'selling', 0.7000, true, NOW(), NOW()),
    (retail_industry_id, 'purchasing', 0.7000, true, NOW(), NOW()),
    (retail_industry_id, 'buying', 0.7000, true, NOW(), NOW()),
    (retail_industry_id, 'stocking', 0.7000, true, NOW(), NOW()),
    (retail_industry_id, 'displaying', 0.7000, true, NOW(), NOW()),
    (retail_industry_id, 'merchandising', 0.7000, true, NOW(), NOW()),
    (retail_industry_id, 'pricing', 0.7000, true, NOW(), NOW()),
    (retail_industry_id, 'promotion', 0.7000, true, NOW(), NOW()),
    (retail_industry_id, 'marketing', 0.7000, true, NOW(), NOW()),
    (retail_industry_id, 'advertising', 0.7000, true, NOW(), NOW()),
    
    -- Retail concepts (weight 0.6000)
    (retail_industry_id, 'commerce', 0.6000, true, NOW(), NOW()),
    (retail_industry_id, 'trade', 0.6000, true, NOW(), NOW()),
    (retail_industry_id, 'business', 0.6000, true, NOW(), NOW()),
    (retail_industry_id, 'commercial', 0.6000, true, NOW(), NOW()),
    (retail_industry_id, 'marketplace', 0.6000, true, NOW(), NOW()),
    (retail_industry_id, 'shopping center', 0.6000, true, NOW(), NOW()),
    (retail_industry_id, 'mall', 0.6000, true, NOW(), NOW()),
    (retail_industry_id, 'plaza', 0.6000, true, NOW(), NOW()),
    (retail_industry_id, 'strip mall', 0.6000, true, NOW(), NOW()),
    (retail_industry_id, 'shopping district', 0.6000, true, NOW(), NOW()),
    
    -- Retail services (weight 0.5000)
    (retail_industry_id, 'customer service', 0.5000, true, NOW(), NOW()),
    (retail_industry_id, 'sales associate', 0.5000, true, NOW(), NOW()),
    (retail_industry_id, 'cashier', 0.5000, true, NOW(), NOW()),
    (retail_industry_id, 'store manager', 0.5000, true, NOW(), NOW()),
    (retail_industry_id, 'retail staff', 0.5000, true, NOW(), NOW()),
    (retail_industry_id, 'store hours', 0.5000, true, NOW(), NOW()),
    (retail_industry_id, 'store location', 0.5000, true, NOW(), NOW()),
    (retail_industry_id, 'store address', 0.5000, true, NOW(), NOW()),
    (retail_industry_id, 'store phone', 0.5000, true, NOW(), NOW()),
    (retail_industry_id, 'store website', 0.5000, true, NOW(), NOW())
    
    ON CONFLICT (industry_id, keyword) DO UPDATE SET
        base_weight = EXCLUDED.base_weight,
        is_active = EXCLUDED.is_active,
        updated_at = NOW();
    
    RAISE NOTICE 'Added 50 retail keywords for industry ID: %', retail_industry_id;
END $$;

-- =============================================================================
-- 2. E-COMMERCE INDUSTRY KEYWORDS (50+ keywords)
-- =============================================================================

-- Get E-commerce industry ID
DO $$
DECLARE
    ecommerce_industry_id INTEGER;
BEGIN
    SELECT id INTO ecommerce_industry_id 
    FROM industries 
    WHERE name = 'E-commerce' AND is_active = true;
    
    IF ecommerce_industry_id IS NULL THEN
        RAISE EXCEPTION 'E-commerce industry not found. Please run add-comprehensive-industries.sql first.';
    END IF;
    
    -- Insert e-commerce keywords with appropriate weights
    INSERT INTO keyword_weights (industry_id, keyword, base_weight, is_active, created_at, updated_at) VALUES
    -- Core e-commerce terms (weight 1.0000)
    (ecommerce_industry_id, 'ecommerce', 1.0000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'e-commerce', 1.0000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online store', 1.0000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online shop', 1.0000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online retail', 1.0000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'digital commerce', 1.0000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'web store', 1.0000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'internet store', 1.0000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online marketplace', 1.0000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'digital storefront', 1.0000, true, NOW(), NOW()),
    
    -- E-commerce platforms (weight 0.9000)
    (ecommerce_industry_id, 'shopify', 0.9000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'woocommerce', 0.9000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'magento', 0.9000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'bigcommerce', 0.9000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'squarespace', 0.9000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'wix', 0.9000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'amazon', 0.9000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'ebay', 0.9000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'etsy', 0.9000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'alibaba', 0.9000, true, NOW(), NOW()),
    
    -- E-commerce activities (weight 0.8000)
    (ecommerce_industry_id, 'online selling', 0.8000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online shopping', 0.8000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online buying', 0.8000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online purchasing', 0.8000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online ordering', 0.8000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online payment', 0.8000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online checkout', 0.8000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online cart', 0.8000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online catalog', 0.8000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online inventory', 0.8000, true, NOW(), NOW()),
    
    -- E-commerce technology (weight 0.7000)
    (ecommerce_industry_id, 'website', 0.7000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'web platform', 0.7000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online platform', 0.7000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'digital platform', 0.7000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'mobile commerce', 0.7000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'mcommerce', 0.7000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'app store', 0.7000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'mobile app', 0.7000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'responsive design', 0.7000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'user experience', 0.7000, true, NOW(), NOW()),
    
    -- E-commerce concepts (weight 0.6000)
    (ecommerce_industry_id, 'digital transformation', 0.6000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online business', 0.6000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'internet business', 0.6000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'web business', 0.6000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'digital business', 0.6000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online presence', 0.6000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'web presence', 0.6000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'digital presence', 0.6000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online marketing', 0.6000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'digital marketing', 0.6000, true, NOW(), NOW()),
    
    -- E-commerce services (weight 0.5000)
    (ecommerce_industry_id, 'online customer service', 0.5000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'live chat', 0.5000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online support', 0.5000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'digital support', 0.5000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'online help', 0.5000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'shipping', 0.5000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'delivery', 0.5000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'fulfillment', 0.5000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'logistics', 0.5000, true, NOW(), NOW()),
    (ecommerce_industry_id, 'order processing', 0.5000, true, NOW(), NOW())
    
    ON CONFLICT (industry_id, keyword) DO UPDATE SET
        base_weight = EXCLUDED.base_weight,
        is_active = EXCLUDED.is_active,
        updated_at = NOW();
    
    RAISE NOTICE 'Added 50 e-commerce keywords for industry ID: %', ecommerce_industry_id;
END $$;

-- =============================================================================
-- 3. WHOLESALE INDUSTRY KEYWORDS (50+ keywords)
-- =============================================================================

-- Get Wholesale industry ID
DO $$
DECLARE
    wholesale_industry_id INTEGER;
BEGIN
    SELECT id INTO wholesale_industry_id 
    FROM industries 
    WHERE name = 'Wholesale' AND is_active = true;
    
    IF wholesale_industry_id IS NULL THEN
        RAISE EXCEPTION 'Wholesale industry not found. Please run add-comprehensive-industries.sql first.';
    END IF;
    
    -- Insert wholesale keywords with appropriate weights
    INSERT INTO keyword_weights (industry_id, keyword, base_weight, is_active, created_at, updated_at) VALUES
    -- Core wholesale terms (weight 1.0000)
    (wholesale_industry_id, 'wholesale', 1.0000, true, NOW(), NOW()),
    (wholesale_industry_id, 'wholesaler', 1.0000, true, NOW(), NOW()),
    (wholesale_industry_id, 'wholesale trade', 1.0000, true, NOW(), NOW()),
    (wholesale_industry_id, 'wholesale business', 1.0000, true, NOW(), NOW()),
    (wholesale_industry_id, 'wholesale distribution', 1.0000, true, NOW(), NOW()),
    (wholesale_industry_id, 'wholesale sales', 1.0000, true, NOW(), NOW()),
    (wholesale_industry_id, 'wholesale operations', 1.0000, true, NOW(), NOW()),
    (wholesale_industry_id, 'wholesale market', 1.0000, true, NOW(), NOW()),
    (wholesale_industry_id, 'wholesale pricing', 1.0000, true, NOW(), NOW()),
    (wholesale_industry_id, 'wholesale supplier', 1.0000, true, NOW(), NOW()),
    
    -- B2B terms (weight 0.9000)
    (wholesale_industry_id, 'b2b', 0.9000, true, NOW(), NOW()),
    (wholesale_industry_id, 'business to business', 0.9000, true, NOW(), NOW()),
    (wholesale_industry_id, 'b2b sales', 0.9000, true, NOW(), NOW()),
    (wholesale_industry_id, 'b2b trade', 0.9000, true, NOW(), NOW()),
    (wholesale_industry_id, 'b2b commerce', 0.9000, true, NOW(), NOW()),
    (wholesale_industry_id, 'b2b marketplace', 0.9000, true, NOW(), NOW()),
    (wholesale_industry_id, 'b2b platform', 0.9000, true, NOW(), NOW()),
    (wholesale_industry_id, 'b2b services', 0.9000, true, NOW(), NOW()),
    (wholesale_industry_id, 'b2b solutions', 0.9000, true, NOW(), NOW()),
    (wholesale_industry_id, 'b2b network', 0.9000, true, NOW(), NOW()),
    
    -- Distribution terms (weight 0.8000)
    (wholesale_industry_id, 'distribution', 0.8000, true, NOW(), NOW()),
    (wholesale_industry_id, 'distributor', 0.8000, true, NOW(), NOW()),
    (wholesale_industry_id, 'distribution center', 0.8000, true, NOW(), NOW()),
    (wholesale_industry_id, 'distribution network', 0.8000, true, NOW(), NOW()),
    (wholesale_industry_id, 'distribution channel', 0.8000, true, NOW(), NOW()),
    (wholesale_industry_id, 'distribution system', 0.8000, true, NOW(), NOW()),
    (wholesale_industry_id, 'distribution services', 0.8000, true, NOW(), NOW()),
    (wholesale_industry_id, 'distribution operations', 0.8000, true, NOW(), NOW()),
    (wholesale_industry_id, 'distribution management', 0.8000, true, NOW(), NOW()),
    (wholesale_industry_id, 'distribution logistics', 0.8000, true, NOW(), NOW()),
    
    -- Trade terms (weight 0.7000)
    (wholesale_industry_id, 'trade', 0.7000, true, NOW(), NOW()),
    (wholesale_industry_id, 'trading', 0.7000, true, NOW(), NOW()),
    (wholesale_industry_id, 'trader', 0.7000, true, NOW(), NOW()),
    (wholesale_industry_id, 'trading company', 0.7000, true, NOW(), NOW()),
    (wholesale_industry_id, 'trading house', 0.7000, true, NOW(), NOW()),
    (wholesale_industry_id, 'trading platform', 0.7000, true, NOW(), NOW()),
    (wholesale_industry_id, 'trading network', 0.7000, true, NOW(), NOW()),
    (wholesale_industry_id, 'trading services', 0.7000, true, NOW(), NOW()),
    (wholesale_industry_id, 'trading operations', 0.7000, true, NOW(), NOW()),
    (wholesale_industry_id, 'trading business', 0.7000, true, NOW(), NOW()),
    
    -- Supply chain terms (weight 0.6000)
    (wholesale_industry_id, 'supply chain', 0.6000, true, NOW(), NOW()),
    (wholesale_industry_id, 'supply chain management', 0.6000, true, NOW(), NOW()),
    (wholesale_industry_id, 'supply chain services', 0.6000, true, NOW(), NOW()),
    (wholesale_industry_id, 'supply chain solutions', 0.6000, true, NOW(), NOW()),
    (wholesale_industry_id, 'supply chain network', 0.6000, true, NOW(), NOW()),
    (wholesale_industry_id, 'supply chain operations', 0.6000, true, NOW(), NOW()),
    (wholesale_industry_id, 'supply chain logistics', 0.6000, true, NOW(), NOW()),
    (wholesale_industry_id, 'supply chain optimization', 0.6000, true, NOW(), NOW()),
    (wholesale_industry_id, 'supply chain integration', 0.6000, true, NOW(), NOW()),
    (wholesale_industry_id, 'supply chain technology', 0.6000, true, NOW(), NOW()),
    
    -- Wholesale services (weight 0.5000)
    (wholesale_industry_id, 'bulk sales', 0.5000, true, NOW(), NOW()),
    (wholesale_industry_id, 'volume sales', 0.5000, true, NOW(), NOW()),
    (wholesale_industry_id, 'quantity discounts', 0.5000, true, NOW(), NOW()),
    (wholesale_industry_id, 'bulk pricing', 0.5000, true, NOW(), NOW()),
    (wholesale_industry_id, 'volume pricing', 0.5000, true, NOW(), NOW()),
    (wholesale_industry_id, 'wholesale pricing', 0.5000, true, NOW(), NOW()),
    (wholesale_industry_id, 'trade pricing', 0.5000, true, NOW(), NOW()),
    (wholesale_industry_id, 'dealer pricing', 0.5000, true, NOW(), NOW()),
    (wholesale_industry_id, 'reseller pricing', 0.5000, true, NOW(), NOW()),
    (wholesale_industry_id, 'distributor pricing', 0.5000, true, NOW(), NOW())
    
    ON CONFLICT (industry_id, keyword) DO UPDATE SET
        base_weight = EXCLUDED.base_weight,
        is_active = EXCLUDED.is_active,
        updated_at = NOW();
    
    RAISE NOTICE 'Added 50 wholesale keywords for industry ID: %', wholesale_industry_id;
END $$;

-- =============================================================================
-- 4. CONSUMER GOODS INDUSTRY KEYWORDS (50+ keywords)
-- =============================================================================

-- Get Consumer Goods industry ID
DO $$
DECLARE
    consumer_goods_industry_id INTEGER;
BEGIN
    SELECT id INTO consumer_goods_industry_id 
    FROM industries 
    WHERE name = 'Consumer Goods' AND is_active = true;
    
    IF consumer_goods_industry_id IS NULL THEN
        RAISE EXCEPTION 'Consumer Goods industry not found. Please run add-comprehensive-industries.sql first.';
    END IF;
    
    -- Insert consumer goods keywords with appropriate weights
    INSERT INTO keyword_weights (industry_id, keyword, base_weight, is_active, created_at, updated_at) VALUES
    -- Core consumer goods terms (weight 1.0000)
    (consumer_goods_industry_id, 'consumer goods', 1.0000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'consumer products', 1.0000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'consumer items', 1.0000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'consumer merchandise', 1.0000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'consumer brands', 1.0000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'consumer market', 1.0000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'consumer sales', 1.0000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'consumer business', 1.0000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'consumer industry', 1.0000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'consumer sector', 1.0000, true, NOW(), NOW()),
    
    -- Product categories (weight 0.9000)
    (consumer_goods_industry_id, 'household goods', 0.9000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'personal care', 0.9000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'beauty products', 0.9000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'cosmetics', 0.9000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'skincare', 0.9000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'hair care', 0.9000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'body care', 0.9000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'health products', 0.9000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'wellness products', 0.9000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'lifestyle products', 0.9000, true, NOW(), NOW()),
    
    -- Manufacturing terms (weight 0.8000)
    (consumer_goods_industry_id, 'manufacturing', 0.8000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'manufacturer', 0.8000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'manufacturing company', 0.8000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'manufacturing business', 0.8000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'manufacturing operations', 0.8000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'production', 0.8000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'producer', 0.8000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'production company', 0.8000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'production facility', 0.8000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'production line', 0.8000, true, NOW(), NOW()),
    
    -- Brand terms (weight 0.7000)
    (consumer_goods_industry_id, 'brand', 0.7000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'branding', 0.7000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'brand management', 0.7000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'brand development', 0.7000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'brand strategy', 0.7000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'brand marketing', 0.7000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'brand promotion', 0.7000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'brand awareness', 0.7000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'brand recognition', 0.7000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'brand loyalty', 0.7000, true, NOW(), NOW()),
    
    -- Market terms (weight 0.6000)
    (consumer_goods_industry_id, 'market', 0.6000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'marketplace', 0.6000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'market research', 0.6000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'market analysis', 0.6000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'market strategy', 0.6000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'market development', 0.6000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'market penetration', 0.6000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'market share', 0.6000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'market position', 0.6000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'market leader', 0.6000, true, NOW(), NOW()),
    
    -- Sales terms (weight 0.5000)
    (consumer_goods_industry_id, 'sales', 0.5000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'sales team', 0.5000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'sales management', 0.5000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'sales strategy', 0.5000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'sales development', 0.5000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'sales growth', 0.5000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'sales performance', 0.5000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'sales targets', 0.5000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'sales goals', 0.5000, true, NOW(), NOW()),
    (consumer_goods_industry_id, 'sales objectives', 0.5000, true, NOW(), NOW())
    
    ON CONFLICT (industry_id, keyword) DO UPDATE SET
        base_weight = EXCLUDED.base_weight,
        is_active = EXCLUDED.is_active,
        updated_at = NOW();
    
    RAISE NOTICE 'Added 50 consumer goods keywords for industry ID: %', consumer_goods_industry_id;
END $$;

-- =============================================================================
-- VERIFICATION QUERIES
-- =============================================================================

-- Verify all retail & e-commerce keywords were added
DO $$
DECLARE
    retail_keyword_count INTEGER;
    ecommerce_keyword_count INTEGER;
    wholesale_keyword_count INTEGER;
    consumer_goods_keyword_count INTEGER;
    total_keyword_count INTEGER;
BEGIN
    -- Count keywords for each industry
    SELECT COUNT(*) INTO retail_keyword_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Retail' AND kw.is_active = true;
    
    SELECT COUNT(*) INTO ecommerce_keyword_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'E-commerce' AND kw.is_active = true;
    
    SELECT COUNT(*) INTO wholesale_keyword_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Wholesale' AND kw.is_active = true;
    
    SELECT COUNT(*) INTO consumer_goods_keyword_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Consumer Goods' AND kw.is_active = true;
    
    total_keyword_count := retail_keyword_count + ecommerce_keyword_count + wholesale_keyword_count + consumer_goods_keyword_count;
    
    -- Report results
    RAISE NOTICE 'Retail keywords added: %', retail_keyword_count;
    RAISE NOTICE 'E-commerce keywords added: %', ecommerce_keyword_count;
    RAISE NOTICE 'Wholesale keywords added: %', wholesale_keyword_count;
    RAISE NOTICE 'Consumer Goods keywords added: %', consumer_goods_keyword_count;
    RAISE NOTICE 'Total retail & e-commerce keywords: %', total_keyword_count;
    
    -- Verify minimum keyword counts
    IF retail_keyword_count >= 50 AND ecommerce_keyword_count >= 50 AND wholesale_keyword_count >= 50 AND consumer_goods_keyword_count >= 50 THEN
        RAISE NOTICE 'SUCCESS: All industries have 50+ keywords as required';
    ELSE
        RAISE NOTICE 'WARNING: Some industries may not have sufficient keywords';
    END IF;
END $$;

-- Display keyword summary by industry
SELECT 
    'RETAIL & E-COMMERCE KEYWORDS SUMMARY' as summary_type,
    '' as spacer;

SELECT 
    i.name as industry_name,
    COUNT(kw.keyword) as keyword_count,
    ROUND(MIN(kw.base_weight), 4) as min_weight,
    ROUND(MAX(kw.base_weight), 4) as max_weight,
    ROUND(AVG(kw.base_weight), 4) as avg_weight
FROM industries i
JOIN keyword_weights kw ON i.id = kw.industry_id
WHERE i.name IN ('Retail', 'E-commerce', 'Wholesale', 'Consumer Goods')
AND kw.is_active = true
GROUP BY i.name
ORDER BY keyword_count DESC;

-- Display sample keywords for each industry
SELECT 
    'SAMPLE KEYWORDS BY INDUSTRY' as summary_type,
    '' as spacer;

SELECT 
    i.name as industry_name,
    kw.keyword,
    kw.base_weight
FROM industries i
JOIN keyword_weights kw ON i.id = kw.industry_id
WHERE i.name IN ('Retail', 'E-commerce', 'Wholesale', 'Consumer Goods')
AND kw.is_active = true
AND kw.base_weight >= 0.9000
ORDER BY i.name, kw.base_weight DESC
LIMIT 20;

COMMIT;

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================
DO $$
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'RETAIL & E-COMMERCE KEYWORDS ADDITION COMPLETED';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Keywords added: 200+ across 4 industries';
    RAISE NOTICE 'Industries covered: Retail, E-commerce, Wholesale, Consumer Goods';
    RAISE NOTICE 'Base weights: 0.5000-1.0000 as specified in plan';
    RAISE NOTICE 'Status: Ready for testing and validation (Subtask 3.2.4.5)';
    RAISE NOTICE 'Next step: Test keyword relevance and classification accuracy';
    RAISE NOTICE '=============================================================================';
END $$;
