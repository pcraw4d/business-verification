-- =============================================================================
-- Restaurant Keywords Addition Script
-- Subtask 1.2.2: Add restaurant keywords with appropriate base weights
-- =============================================================================

-- This script adds comprehensive restaurant keywords to improve
-- classification accuracy for food service businesses

-- =============================================================================
-- 1. ADD RESTAURANT KEYWORDS
-- =============================================================================

-- Restaurant Industry Keywords (Primary Category)
INSERT INTO keyword_weights (industry_id, keyword, base_weight, usage_count, is_active, created_at, last_updated)
SELECT 
    i.id,
    kw.keyword,
    kw.base_weight,
    0 as usage_count,
    true as is_active,
    NOW() as created_at,
    NOW() as last_updated
FROM industries i
CROSS JOIN (VALUES
    -- Core restaurant keywords (High weight - 0.8-1.0)
    ('restaurant', 1.0000),
    ('dining', 0.950),
    ('cuisine', 0.900),
    ('menu', 0.900),
    ('chef', 0.900),
    ('kitchen', 0.850),
    ('food service', 0.850),
    ('meal', 0.800),
    ('dining room', 0.800),
    ('table service', 0.800),
    
    -- Restaurant types (High weight - 0.8-0.9)
    ('fine dining', 0.900),
    ('casual dining', 0.850),
    ('family restaurant', 0.850),
    ('steakhouse', 0.850),
    ('pizzeria', 0.850),
    ('bistro', 0.800),
    ('grill', 0.800),
    ('bar & grill', 0.800),
    ('diner', 0.800),
    ('eatery', 0.750),
    
    -- Cuisine types (Medium-High weight - 0.7-0.8)
    ('italian', 0.800),
    ('chinese', 0.800),
    ('mexican', 0.800),
    ('american', 0.750),
    ('french', 0.800),
    ('japanese', 0.800),
    ('thai', 0.750),
    ('indian', 0.750),
    ('mediterranean', 0.750),
    ('seafood', 0.750),
    ('asian', 0.700),
    ('european', 0.700),
    
    -- Food items (Medium weight - 0.6-0.7)
    ('pasta', 0.700),
    ('pizza', 0.750),
    ('burger', 0.700),
    ('sandwich', 0.650),
    ('salad', 0.600),
    ('soup', 0.600),
    ('dessert', 0.600),
    ('appetizer', 0.650),
    ('entree', 0.700),
    ('main course', 0.700),
    
    -- Beverages (Medium weight - 0.6-0.7)
    ('wine', 0.700),
    ('cocktail', 0.650),
    ('beer', 0.600),
    ('coffee', 0.600),
    ('tea', 0.600),
    ('beverage', 0.650),
    ('drink', 0.600)
) AS kw(keyword, base_weight)
WHERE i.name = 'Restaurants'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    last_updated = NOW();

-- Fast Food Industry Keywords
INSERT INTO keyword_weights (industry_id, keyword, base_weight, usage_count, is_active, created_at, last_updated)
SELECT 
    i.id,
    kw.keyword,
    kw.base_weight,
    0 as usage_count,
    true as is_active,
    NOW() as created_at,
    NOW() as last_updated
FROM industries i
CROSS JOIN (VALUES
    -- Core fast food keywords (High weight - 0.8-1.0)
    ('fast food', 1.0000),
    ('quick service', 0.950),
    ('drive thru', 0.900),
    ('drive through', 0.900),
    ('takeout', 0.900),
    ('take away', 0.850),
    ('fast casual', 0.850),
    ('counter service', 0.800),
    ('quick meal', 0.800),
    ('fast service', 0.800),
    
    -- Fast food chains (High weight - 0.8-0.9)
    ('mcdonalds', 0.900),
    ('burger king', 0.900),
    ('kfc', 0.900),
    ('subway', 0.900),
    ('taco bell', 0.900),
    ('pizza hut', 0.900),
    ('dominos', 0.900),
    ('wendys', 0.900),
    ('chick fil a', 0.900),
    ('popeyes', 0.900),
    
    -- Fast food items (Medium-High weight - 0.7-0.8)
    ('burger', 0.800),
    ('fries', 0.800),
    ('chicken', 0.800),
    ('sandwich', 0.750),
    ('pizza', 0.750),
    ('taco', 0.750),
    ('wrap', 0.700),
    ('nuggets', 0.700),
    ('combo meal', 0.700),
    ('value meal', 0.700),
    
    -- Service characteristics (Medium weight - 0.6-0.7)
    ('self service', 0.700),
    ('no table service', 0.650),
    ('quick preparation', 0.650),
    ('convenience', 0.600),
    ('affordable', 0.600),
    ('budget friendly', 0.600)
) AS kw(keyword, base_weight)
WHERE i.name = 'Fast Food'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    last_updated = NOW();

-- Fine Dining Industry Keywords
INSERT INTO keyword_weights (industry_id, keyword, base_weight, usage_count, is_active, created_at, last_updated)
SELECT 
    i.id,
    kw.keyword,
    kw.base_weight,
    0 as usage_count,
    true as is_active,
    NOW() as created_at,
    NOW() as last_updated
FROM industries i
CROSS JOIN (VALUES
    -- Core fine dining keywords (High weight - 0.8-1.0)
    ('fine dining', 1.0000),
    ('upscale', 0.950),
    ('gourmet', 0.950),
    ('premium', 0.900),
    ('high end', 0.900),
    ('luxury', 0.900),
    ('elegant', 0.850),
    ('sophisticated', 0.850),
    ('refined', 0.800),
    ('exclusive', 0.800),
    
    -- Fine dining characteristics (High weight - 0.8-0.9)
    ('white tablecloth', 0.900),
    ('formal dining', 0.900),
    ('wine pairing', 0.900),
    ('sommelier', 0.900),
    ('executive chef', 0.900),
    ('tasting menu', 0.900),
    ('degustation', 0.850),
    ('a la carte', 0.850),
    ('reservations required', 0.800),
    ('dress code', 0.800),
    
    -- Premium cuisine (Medium-High weight - 0.7-0.8)
    ('french cuisine', 0.800),
    ('continental', 0.800),
    ('haute cuisine', 0.800),
    ('molecular gastronomy', 0.800),
    ('fusion', 0.750),
    ('artisanal', 0.750),
    ('organic', 0.700),
    ('farm to table', 0.700),
    ('locally sourced', 0.700),
    ('seasonal menu', 0.700),
    
    -- Premium beverages (Medium weight - 0.6-0.7)
    ('wine cellar', 0.800),
    ('craft cocktails', 0.750),
    ('premium spirits', 0.700),
    ('champagne', 0.700),
    ('vintage wine', 0.700),
    ('single malt', 0.650),
    ('artisanal beer', 0.600)
) AS kw(keyword, base_weight)
WHERE i.name = 'Fine Dining'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    last_updated = NOW();

-- Casual Dining Industry Keywords
INSERT INTO keyword_weights (industry_id, keyword, base_weight, usage_count, is_active, created_at, last_updated)
SELECT 
    i.id,
    kw.keyword,
    kw.base_weight,
    0 as usage_count,
    true as is_active,
    NOW() as created_at,
    NOW() as last_updated
FROM industries i
CROSS JOIN (VALUES
    -- Core casual dining keywords (High weight - 0.8-1.0)
    ('casual dining', 1.0000),
    ('family restaurant', 0.950),
    ('table service', 0.900),
    ('moderate pricing', 0.850),
    ('comfortable atmosphere', 0.800),
    ('relaxed dining', 0.800),
    ('family friendly', 0.800),
    ('kid friendly', 0.750),
    ('group dining', 0.750),
    ('social dining', 0.700),
    
    -- Casual dining characteristics (Medium-High weight - 0.7-0.8)
    ('full service', 0.800),
    ('wait staff', 0.800),
    ('server', 0.800),
    ('hostess', 0.750),
    ('busboy', 0.700),
    ('moderate prices', 0.750),
    ('reasonable prices', 0.700),
    ('value dining', 0.700),
    ('comfort food', 0.750),
    ('home style', 0.700),
    
    -- Popular casual chains (Medium-High weight - 0.7-0.8)
    ('olive garden', 0.800),
    ('applebees', 0.800),
    ('chilis', 0.800),
    ('tgi fridays', 0.800),
    ('red lobster', 0.800),
    ('outback steakhouse', 0.800),
    ('texas roadhouse', 0.800),
    ('cracker barrel', 0.800),
    ('dennys', 0.750),
    ('ihop', 0.750),
    
    -- Menu items (Medium weight - 0.6-0.7)
    ('appetizers', 0.700),
    ('entrees', 0.700),
    ('desserts', 0.650),
    ('kids menu', 0.700),
    ('senior menu', 0.650),
    ('happy hour', 0.700),
    ('specials', 0.650),
    ('daily specials', 0.650)
) AS kw(keyword, base_weight)
WHERE i.name = 'Casual Dining'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    last_updated = NOW();

-- Quick Service Industry Keywords
INSERT INTO keyword_weights (industry_id, keyword, base_weight, usage_count, is_active, created_at, last_updated)
SELECT 
    i.id,
    kw.keyword,
    kw.base_weight,
    0 as usage_count,
    true as is_active,
    NOW() as created_at,
    NOW() as last_updated
FROM industries i
CROSS JOIN (VALUES
    -- Core quick service keywords (High weight - 0.8-1.0)
    ('quick service', 1.0000),
    ('fast casual', 0.950),
    ('limited table service', 0.900),
    ('quick preparation', 0.900),
    ('fast casual dining', 0.850),
    ('counter ordering', 0.850),
    ('self seating', 0.800),
    ('casual atmosphere', 0.800),
    ('modern dining', 0.750),
    ('contemporary', 0.700),
    
    -- Quick service chains (High weight - 0.8-0.9)
    ('chipotle', 0.900),
    ('panera bread', 0.900),
    ('qdoba', 0.900),
    ('moes southwest grill', 0.900),
    ('five guys', 0.900),
    ('shake shack', 0.900),
    ('in n out', 0.900),
    ('whataburger', 0.900),
    ('culvers', 0.900),
    ('raising canes', 0.900),
    
    -- Service model (Medium-High weight - 0.7-0.8)
    ('order at counter', 0.800),
    ('pick up at counter', 0.800),
    ('self serve', 0.750),
    ('limited service', 0.750),
    ('fast preparation', 0.750),
    ('fresh ingredients', 0.700),
    ('made to order', 0.700),
    ('customizable', 0.700),
    ('build your own', 0.700),
    ('assembly line', 0.650),
    
    -- Menu characteristics (Medium weight - 0.6-0.7)
    ('fresh food', 0.700),
    ('healthy options', 0.700),
    ('premium ingredients', 0.700),
    ('artisanal', 0.650),
    ('gourmet', 0.650),
    ('specialty', 0.650),
    ('signature', 0.650),
    ('unique', 0.600)
) AS kw(keyword, base_weight)
WHERE i.name = 'Quick Service'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    last_updated = NOW();

-- Food & Beverage Industry Keywords (General Category)
INSERT INTO keyword_weights (industry_id, keyword, base_weight, usage_count, is_active, created_at, last_updated)
SELECT 
    i.id,
    kw.keyword,
    kw.base_weight,
    0 as usage_count,
    true as is_active,
    NOW() as created_at,
    NOW() as last_updated
FROM industries i
CROSS JOIN (VALUES
    -- Core food & beverage keywords (High weight - 0.8-1.0)
    ('food & beverage', 1.0000),
    ('food service', 0.950),
    ('beverage service', 0.900),
    ('food establishment', 0.900),
    ('dining establishment', 0.850),
    ('food business', 0.850),
    ('hospitality', 0.800),
    ('food industry', 0.800),
    ('catering', 0.750),
    ('food production', 0.700),
    
    -- General food terms (Medium-High weight - 0.7-0.8)
    ('food', 0.800),
    ('beverage', 0.800),
    ('drink', 0.750),
    ('refreshment', 0.700),
    ('sustenance', 0.650),
    ('nourishment', 0.650),
    ('culinary', 0.750),
    ('gastronomy', 0.700),
    ('cooking', 0.700),
    ('preparation', 0.650),
    
    -- Service types (Medium weight - 0.6-0.7)
    ('dine in', 0.700),
    ('takeout', 0.700),
    ('delivery', 0.700),
    ('catering', 0.700),
    ('banquet', 0.650),
    ('event', 0.650),
    ('party', 0.600),
    ('celebration', 0.600),
    ('gathering', 0.600),
    ('function', 0.600)
) AS kw(keyword, base_weight)
WHERE i.name = 'Food & Beverage'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    last_updated = NOW();

-- =============================================================================
-- 2. ADD SPECIALIZED RESTAURANT KEYWORDS
-- =============================================================================

-- Catering Industry Keywords
INSERT INTO keyword_weights (industry_id, keyword, base_weight, usage_count, is_active, created_at, last_updated)
SELECT 
    i.id,
    kw.keyword,
    kw.base_weight,
    0 as usage_count,
    true as is_active,
    NOW() as created_at,
    NOW() as last_updated
FROM industries i
CROSS JOIN (VALUES
    ('catering', 1.0000),
    ('event catering', 0.950),
    ('party catering', 0.900),
    ('corporate catering', 0.900),
    ('wedding catering', 0.900),
    ('banquet', 0.850),
    ('buffet', 0.850),
    ('off premise', 0.800),
    ('on premise', 0.800),
    ('full service catering', 0.800),
    ('drop off catering', 0.750),
    ('catering service', 0.800),
    ('food delivery', 0.700),
    ('event planning', 0.700),
    ('special events', 0.700)
) AS kw(keyword, base_weight)
WHERE i.name = 'Catering'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    last_updated = NOW();

-- Food Trucks Industry Keywords
INSERT INTO keyword_weights (industry_id, keyword, base_weight, usage_count, is_active, created_at, last_updated)
SELECT 
    i.id,
    kw.keyword,
    kw.base_weight,
    0 as usage_count,
    true as is_active,
    NOW() as created_at,
    NOW() as last_updated
FROM industries i
CROSS JOIN (VALUES
    ('food truck', 1.0000),
    ('mobile food', 0.950),
    ('street food', 0.900),
    ('food cart', 0.900),
    ('mobile kitchen', 0.850),
    ('roaming food', 0.800),
    ('truck food', 0.800),
    ('mobile dining', 0.750),
    ('street vendor', 0.750),
    ('food stand', 0.700),
    ('portable food', 0.700),
    ('nomadic food', 0.650),
    ('traveling food', 0.650),
    ('pop up food', 0.600),
    ('temporary food', 0.600)
) AS kw(keyword, base_weight)
WHERE i.name = 'Food Trucks'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    last_updated = NOW();

-- Cafes & Coffee Shops Industry Keywords
INSERT INTO keyword_weights (industry_id, keyword, base_weight, usage_count, is_active, created_at, last_updated)
SELECT 
    i.id,
    kw.keyword,
    kw.base_weight,
    0 as usage_count,
    true as is_active,
    NOW() as created_at,
    NOW() as last_updated
FROM industries i
CROSS JOIN (VALUES
    ('cafe', 1.0000),
    ('coffee shop', 1.0000),
    ('coffee house', 0.950),
    ('coffee bar', 0.900),
    ('espresso bar', 0.900),
    ('coffee', 0.900),
    ('espresso', 0.850),
    ('latte', 0.800),
    ('cappuccino', 0.800),
    ('americano', 0.750),
    ('mocha', 0.750),
    ('frappuccino', 0.700),
    ('pastry', 0.700),
    ('bakery', 0.700),
    ('light food', 0.700),
    ('breakfast', 0.700),
    ('brunch', 0.700),
    ('sandwich', 0.650),
    ('salad', 0.650),
    ('soup', 0.600),
    ('wifi', 0.600),
    ('study space', 0.600),
    ('meeting space', 0.600)
) AS kw(keyword, base_weight)
WHERE i.name = 'Cafes & Coffee Shops'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    last_updated = NOW();

-- Bars & Pubs Industry Keywords
INSERT INTO keyword_weights (industry_id, keyword, base_weight, usage_count, is_active, created_at, last_updated)
SELECT 
    i.id,
    kw.keyword,
    kw.base_weight,
    0 as usage_count,
    true as is_active,
    NOW() as created_at,
    NOW() as last_updated
FROM industries i
CROSS JOIN (VALUES
    ('bar', 1.0000),
    ('pub', 1.0000),
    ('tavern', 0.950),
    ('lounge', 0.900),
    ('cocktail bar', 0.900),
    ('sports bar', 0.900),
    ('wine bar', 0.900),
    ('beer bar', 0.900),
    ('alcohol', 0.850),
    ('drinks', 0.850),
    ('cocktails', 0.850),
    ('beer', 0.800),
    ('wine', 0.800),
    ('spirits', 0.800),
    ('liquor', 0.800),
    ('happy hour', 0.750),
    ('nightlife', 0.750),
    ('entertainment', 0.700),
    ('live music', 0.700),
    ('darts', 0.650),
    ('pool', 0.650),
    ('tv', 0.650),
    ('sports', 0.650)
) AS kw(keyword, base_weight)
WHERE i.name = 'Bars & Pubs'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    last_updated = NOW();

-- Breweries Industry Keywords
INSERT INTO keyword_weights (industry_id, keyword, base_weight, usage_count, is_active, created_at, last_updated)
SELECT 
    i.id,
    kw.keyword,
    kw.base_weight,
    0 as usage_count,
    true as is_active,
    NOW() as created_at,
    NOW() as last_updated
FROM industries i
CROSS JOIN (VALUES
    ('brewery', 1.0000),
    ('brewing', 0.950),
    ('beer production', 0.900),
    ('craft beer', 0.900),
    ('microbrewery', 0.900),
    ('brewpub', 0.900),
    ('beer', 0.850),
    ('ale', 0.800),
    ('lager', 0.800),
    ('ipa', 0.800),
    ('stout', 0.750),
    ('porter', 0.750),
    ('pilsner', 0.750),
    ('wheat beer', 0.700),
    ('seasonal beer', 0.700),
    ('limited edition', 0.700),
    ('tasting room', 0.800),
    ('beer tasting', 0.800),
    ('brewery tour', 0.750),
    ('beer garden', 0.750),
    ('taproom', 0.750),
    ('beer hall', 0.700)
) AS kw(keyword, base_weight)
WHERE i.name = 'Breweries'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    last_updated = NOW();

-- Wineries Industry Keywords
INSERT INTO keyword_weights (industry_id, keyword, base_weight, usage_count, is_active, created_at, last_updated)
SELECT 
    i.id,
    kw.keyword,
    kw.base_weight,
    0 as usage_count,
    true as is_active,
    NOW() as created_at,
    NOW() as last_updated
FROM industries i
CROSS JOIN (VALUES
    ('winery', 1.0000),
    ('wine production', 0.950),
    ('vineyard', 0.950),
    ('wine making', 0.900),
    ('viticulture', 0.900),
    ('wine', 0.850),
    ('red wine', 0.800),
    ('white wine', 0.800),
    ('rose wine', 0.750),
    ('sparkling wine', 0.750),
    ('champagne', 0.750),
    ('wine tasting', 0.850),
    ('tasting room', 0.800),
    ('wine cellar', 0.800),
    ('wine tour', 0.750),
    ('wine club', 0.700),
    ('wine bar', 0.700),
    ('sommelier', 0.700),
    ('wine pairing', 0.700),
    ('grape', 0.650),
    ('harvest', 0.650),
    ('vintage', 0.650)
) AS kw(keyword, base_weight)
WHERE i.name = 'Wineries'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    last_updated = NOW();

-- =============================================================================
-- 3. VERIFICATION QUERIES
-- =============================================================================

-- Verify restaurant keywords were added successfully
DO $$
DECLARE
    total_keywords INTEGER;
    restaurant_keywords INTEGER;
    fast_food_keywords INTEGER;
    fine_dining_keywords INTEGER;
BEGIN
    -- Count total keywords added
    SELECT COUNT(*) INTO total_keywords 
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name IN (
        'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
        'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
        'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
    )
    AND kw.is_active = true;
    
    -- Count keywords per major industry
    SELECT COUNT(*) INTO restaurant_keywords 
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Restaurants' AND kw.is_active = true;
    
    SELECT COUNT(*) INTO fast_food_keywords 
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Fast Food' AND kw.is_active = true;
    
    SELECT COUNT(*) INTO fine_dining_keywords 
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Fine Dining' AND kw.is_active = true;
    
    -- Report results
    RAISE NOTICE 'Restaurant keywords added: %', total_keywords;
    RAISE NOTICE 'Restaurants industry keywords: %', restaurant_keywords;
    RAISE NOTICE 'Fast Food industry keywords: %', fast_food_keywords;
    RAISE NOTICE 'Fine Dining industry keywords: %', fine_dining_keywords;
    
    -- Verify weight ranges
    RAISE NOTICE 'Keyword weight verification:';
    FOR total_keywords IN 
        SELECT i.name, MIN(kw.base_weight) as min_weight, MAX(kw.base_weight) as max_weight, COUNT(*) as keyword_count
        FROM keyword_weights kw
        JOIN industries i ON kw.industry_id = i.id
        WHERE i.name IN ('Restaurants', 'Fast Food', 'Fine Dining')
        AND kw.is_active = true
        GROUP BY i.name
        ORDER BY keyword_count DESC
    LOOP
        RAISE NOTICE '  %: % keywords, weights %.4f-%.4f', 
            total_keywords.name, total_keywords.keyword_count, 
            total_keywords.min_weight, total_keywords.max_weight;
    END LOOP;
END $$;

-- =============================================================================
-- 4. DISPLAY ADDED KEYWORDS SUMMARY
-- =============================================================================

-- Show keyword count per restaurant industry
SELECT 
    i.name as industry_name,
    COUNT(kw.keyword) as keyword_count,
    MIN(kw.base_weight) as min_weight,
    MAX(kw.base_weight) as max_weight,
    ROUND(AVG(kw.base_weight), 4) as avg_weight
FROM industries i
LEFT JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
WHERE i.name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND i.is_active = true
GROUP BY i.name
ORDER BY keyword_count DESC, i.name;

-- =============================================================================
-- 5. COMPLETION MESSAGE
-- =============================================================================

DO $$
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'RESTAURANT KEYWORDS ADDITION COMPLETED SUCCESSFULLY';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Added comprehensive keywords for 12 restaurant industry categories';
    RAISE NOTICE 'Keywords include core terms, industry-specific terms, and service characteristics';
    RAISE NOTICE 'Base weights range from 0.600 to 1.0000 based on relevance and specificity';
    RAISE NOTICE 'All keywords are active and ready for classification testing';
    RAISE NOTICE 'Next step: Add restaurant classification codes (Subtask 1.2.3)';
    RAISE NOTICE '=============================================================================';
END $$;
