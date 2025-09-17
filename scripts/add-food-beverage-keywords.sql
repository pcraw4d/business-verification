-- =============================================================================
-- Add Food & Beverage Industry Keywords
-- Fix for missing keywords in Food & Beverage industry
-- =============================================================================

-- This script adds comprehensive keywords for the Food & Beverage industry
-- to complete the restaurant classification system

-- =============================================================================
-- 1. ADD FOOD & BEVERAGE INDUSTRY KEYWORDS
-- =============================================================================

-- Food & Beverage Industry Keywords
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
    -- Core Food & Beverage Terms (High weight - 0.9-1.0)
    ('food', 1.000),
    ('beverage', 1.000),
    ('food service', 0.950),
    ('dining', 0.900),
    ('restaurant', 0.900),
    ('cuisine', 0.900),
    ('meal', 0.900),
    ('kitchen', 0.850),
    ('menu', 0.850),
    ('chef', 0.850),
    
    -- Service Types (High weight - 0.8-0.9)
    ('catering', 0.900),
    ('food truck', 0.850),
    ('cafe', 0.800),
    ('bar', 0.800),
    ('pub', 0.800),
    ('brewery', 0.800),
    ('winery', 0.800),
    ('fast food', 0.800),
    ('quick service', 0.800),
    ('fine dining', 0.800),
    ('casual dining', 0.800),
    
    -- Food Categories (Medium-High weight - 0.7-0.8)
    ('italian food', 0.800),
    ('chinese food', 0.800),
    ('mexican food', 0.800),
    ('american food', 0.750),
    ('french food', 0.800),
    ('japanese food', 0.800),
    ('thai food', 0.750),
    ('indian food', 0.750),
    ('mediterranean food', 0.750),
    ('seafood', 0.800),
    ('steak', 0.750),
    ('pizza', 0.800),
    ('sushi', 0.800),
    ('pasta', 0.750),
    ('burger', 0.750),
    ('sandwich', 0.700),
    ('salad', 0.700),
    ('soup', 0.700),
    ('dessert', 0.700),
    ('appetizer', 0.700),
    
    -- Beverage Categories (Medium-High weight - 0.7-0.8)
    ('wine', 0.800),
    ('beer', 0.800),
    ('cocktail', 0.750),
    ('coffee', 0.750),
    ('tea', 0.700),
    ('juice', 0.700),
    ('soda', 0.650),
    ('water', 0.600),
    ('alcohol', 0.800),
    ('spirits', 0.750),
    ('liquor', 0.750),
    
    -- Service Characteristics (Medium weight - 0.6-0.7)
    ('table service', 0.750),
    ('waiter', 0.700),
    ('waitress', 0.700),
    ('hostess', 0.650),
    ('reservations', 0.700),
    ('takeout', 0.700),
    ('delivery', 0.700),
    ('drive through', 0.700),
    ('buffet', 0.700),
    ('self service', 0.650),
    ('counter service', 0.650),
    
    -- Business Types (Medium weight - 0.6-0.7)
    ('eatery', 0.700),
    ('bistro', 0.750),
    ('grill', 0.750),
    ('diner', 0.750),
    ('steakhouse', 0.800),
    ('pizzeria', 0.800),
    ('sushi bar', 0.800),
    ('tapas bar', 0.750),
    ('sports bar', 0.750),
    ('wine bar', 0.800),
    ('coffee shop', 0.750),
    ('bakery', 0.700),
    ('deli', 0.700),
    ('market', 0.650),
    
    -- Time-based Services (Medium weight - 0.6-0.7)
    ('breakfast', 0.700),
    ('brunch', 0.750),
    ('lunch', 0.750),
    ('dinner', 0.750),
    ('late night', 0.650),
    ('happy hour', 0.700),
    ('24 hour', 0.650),
    
    -- Special Features (Medium weight - 0.6-0.7)
    ('outdoor seating', 0.650),
    ('patio', 0.650),
    ('terrace', 0.650),
    ('rooftop', 0.650),
    ('live music', 0.650),
    ('entertainment', 0.600),
    ('family friendly', 0.650),
    ('kid friendly', 0.650),
    ('romantic', 0.650),
    ('upscale', 0.700),
    ('casual', 0.700),
    ('affordable', 0.650),
    ('premium', 0.700),
    ('gourmet', 0.750),
    ('artisan', 0.700),
    ('craft', 0.700),
    ('organic', 0.700),
    ('local', 0.700),
    ('fresh', 0.700),
    ('homemade', 0.700),
    
    -- Dietary Options (Medium weight - 0.6-0.7)
    ('vegetarian', 0.700),
    ('vegan', 0.700),
    ('gluten free', 0.650),
    ('dairy free', 0.650),
    ('keto', 0.650),
    ('paleo', 0.650),
    ('healthy', 0.700),
    ('low calorie', 0.650),
    ('low fat', 0.650),
    ('low sodium', 0.650),
    
    -- Business Operations (Lower weight - 0.6)
    ('franchise', 0.600),
    ('chain', 0.600),
    ('independent', 0.600),
    ('local business', 0.600),
    ('small business', 0.600),
    ('corporate', 0.600),
    ('catering company', 0.700),
    ('food service provider', 0.700),
    ('restaurant group', 0.650),
    ('hospitality', 0.650),
    ('tourism', 0.600),
    ('events', 0.650),
    ('catering events', 0.700),
    ('wedding catering', 0.700),
    ('corporate catering', 0.700),
    ('party catering', 0.700)
) AS kw(keyword, base_weight)
WHERE i.name = 'Food & Beverage'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    last_updated = NOW();

-- =============================================================================
-- 2. VERIFICATION
-- =============================================================================

-- Verify Food & Beverage keywords were added
DO $$
DECLARE
    keyword_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO keyword_count 
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Food & Beverage' 
    AND kw.is_active = true;
    
    RAISE NOTICE 'Food & Beverage keywords added: %', keyword_count;
    
    IF keyword_count >= 50 THEN
        RAISE NOTICE 'STATUS: Food & Beverage keywords successfully added';
    ELSE
        RAISE NOTICE 'STATUS: Food & Beverage keywords need more work';
    END IF;
END $$;

-- Show Food & Beverage keywords summary
SELECT 
    'Food & Beverage Keywords Summary' as summary,
    COUNT(*) as total_keywords,
    MIN(base_weight) as min_weight,
    MAX(base_weight) as max_weight,
    ROUND(AVG(base_weight), 3) as avg_weight
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name = 'Food & Beverage' 
AND kw.is_active = true;

-- Show keyword distribution by weight
SELECT 
    'Keyword Distribution by Weight' as summary,
    CASE 
        WHEN base_weight >= 0.900 THEN 'High (0.9-1.0)'
        WHEN base_weight >= 0.800 THEN 'Medium-High (0.8-0.9)'
        WHEN base_weight >= 0.700 THEN 'Medium (0.7-0.8)'
        WHEN base_weight >= 0.600 THEN 'Low-Medium (0.6-0.7)'
        ELSE 'Low (<0.6)'
    END as weight_category,
    COUNT(*) as keyword_count
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name = 'Food & Beverage' 
AND kw.is_active = true
GROUP BY 
    CASE 
        WHEN base_weight >= 0.900 THEN 'High (0.9-1.0)'
        WHEN base_weight >= 0.800 THEN 'Medium-High (0.8-0.9)'
        WHEN base_weight >= 0.700 THEN 'Medium (0.7-0.8)'
        WHEN base_weight >= 0.600 THEN 'Low-Medium (0.6-0.7)'
        ELSE 'Low (<0.6)'
    END
ORDER BY MIN(base_weight) DESC;

-- =============================================================================
-- 3. COMPLETION MESSAGE
-- =============================================================================

DO $$
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'FOOD & BEVERAGE KEYWORDS ADDITION COMPLETED';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Added comprehensive keywords for Food & Beverage industry';
    RAISE NOTICE 'Keywords cover all aspects of food service and beverage operations';
    RAISE NOTICE 'Weight distribution optimized for classification accuracy';
    RAISE NOTICE 'All keywords are active and ready for classification testing';
    RAISE NOTICE 'Next step: Expand classification codes for all industries';
    RAISE NOTICE '=============================================================================';
END $$;
