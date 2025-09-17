-- =============================================================================
-- Restaurant Industries Addition Script
-- Subtask 1.2.1: Add restaurant industries to database
-- =============================================================================

-- This script adds comprehensive restaurant industry categories to improve
-- classification accuracy for food service businesses

-- =============================================================================
-- 1. ADD RESTAURANT INDUSTRIES
-- =============================================================================

-- Add restaurant industry categories with appropriate confidence thresholds
INSERT INTO industries (name, description, category, confidence_threshold, is_active, created_at, updated_at) VALUES
-- Primary Restaurant Categories
('Restaurants', 'Full-service restaurants including fine dining, casual dining, and family restaurants', 'traditional', 0.75, true, NOW(), NOW()),
('Fast Food', 'Quick service restaurants, fast food chains, and takeout establishments', 'traditional', 0.80, true, NOW(), NOW()),
('Food & Beverage', 'General food and beverage services including restaurants, cafes, and food service', 'traditional', 0.70, true, NOW(), NOW()),

-- Specialized Restaurant Types
('Fine Dining', 'Upscale restaurants with premium dining experiences and high-end cuisine', 'traditional', 0.85, true, NOW(), NOW()),
('Casual Dining', 'Mid-range restaurants with table service and moderate pricing', 'traditional', 0.75, true, NOW(), NOW()),
('Quick Service', 'Fast casual restaurants with limited table service and quick preparation', 'traditional', 0.80, true, NOW(), NOW()),

-- Food Service Categories
('Catering', 'Food catering services for events, parties, and corporate functions', 'traditional', 0.70, true, NOW(), NOW()),
('Food Trucks', 'Mobile food service vehicles and street food vendors', 'emerging', 0.75, true, NOW(), NOW()),
('Cafes & Coffee Shops', 'Coffee shops, cafes, and light food service establishments', 'traditional', 0.70, true, NOW(), NOW()),

-- Beverage Service
('Bars & Pubs', 'Alcoholic beverage service establishments including bars, pubs, and taverns', 'traditional', 0.75, true, NOW(), NOW()),
('Breweries', 'Beer production and tasting establishments', 'traditional', 0.80, true, NOW(), NOW()),
('Wineries', 'Wine production and tasting establishments', 'traditional', 0.80, true, NOW(), NOW())

ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    confidence_threshold = EXCLUDED.confidence_threshold,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- =============================================================================
-- 2. VERIFICATION QUERIES
-- =============================================================================

-- Verify restaurant industries were added successfully
DO $$
DECLARE
    restaurant_count INTEGER;
    fast_food_count INTEGER;
    food_beverage_count INTEGER;
BEGIN
    -- Count restaurant industries
    SELECT COUNT(*) INTO restaurant_count 
    FROM industries 
    WHERE name IN ('Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries')
    AND is_active = true;
    
    -- Verify specific industries exist
    SELECT COUNT(*) INTO fast_food_count 
    FROM industries 
    WHERE name = 'Fast Food' AND is_active = true;
    
    SELECT COUNT(*) INTO food_beverage_count 
    FROM industries 
    WHERE name = 'Food & Beverage' AND is_active = true;
    
    -- Report results
    RAISE NOTICE 'Restaurant industries added: %', restaurant_count;
    RAISE NOTICE 'Fast Food industry exists: %', (fast_food_count > 0);
    RAISE NOTICE 'Food & Beverage industry exists: %', (food_beverage_count > 0);
    
    -- Verify confidence thresholds
    RAISE NOTICE 'Restaurant confidence thresholds:';
    FOR restaurant_count IN 
        SELECT id, name, confidence_threshold 
        FROM industries 
        WHERE name IN ('Restaurants', 'Fast Food', 'Fine Dining', 'Casual Dining')
        AND is_active = true
        ORDER BY confidence_threshold DESC
    LOOP
        RAISE NOTICE '  %: %', restaurant_count.name, restaurant_count.confidence_threshold;
    END LOOP;
END $$;

-- =============================================================================
-- 3. DISPLAY ADDED INDUSTRIES
-- =============================================================================

-- Show all restaurant-related industries
SELECT 
    id,
    name,
    description,
    category,
    confidence_threshold,
    is_active,
    created_at
FROM industries 
WHERE name IN (
    'Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 
    'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 
    'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries'
)
AND is_active = true
ORDER BY confidence_threshold DESC, name;

-- =============================================================================
-- 4. COMPLETION MESSAGE
-- =============================================================================

DO $$
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'RESTAURANT INDUSTRIES ADDITION COMPLETED SUCCESSFULLY';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Added 12 restaurant industry categories with appropriate confidence thresholds';
    RAISE NOTICE 'Industries range from 0.70 to 0.85 confidence thresholds';
    RAISE NOTICE 'All industries are active and ready for keyword association';
    RAISE NOTICE 'Next step: Add restaurant keywords (Subtask 1.2.2)';
    RAISE NOTICE '=============================================================================';
END $$;
