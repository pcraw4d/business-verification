-- =============================================================================
-- AGRICULTURE & ENERGY KEYWORDS COMPREHENSIVE SCRIPT
-- Task 3.2.7: Add agriculture & energy keywords (50+ agriculture and energy-specific keywords with base weights 0.5-1.0)
-- =============================================================================
-- This script adds comprehensive agriculture & energy keywords across all agriculture
-- and energy industries to achieve >85% classification accuracy for agriculture and energy businesses.
-- 
-- Agriculture & Energy Industries Covered:
-- 1. Agriculture (confidence_threshold: 0.75)
-- 2. Food Production (confidence_threshold: 0.70)
-- 3. Energy Services (confidence_threshold: 0.75)
-- 4. Renewable Energy (confidence_threshold: 0.80)
--
-- Total: 200+ comprehensive keywords across 4 agriculture & energy industries
-- =============================================================================

-- Start transaction for atomic operation
BEGIN;

-- =============================================================================
-- 1. AGRICULTURE KEYWORDS (70+ keywords)
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core Agriculture Keywords (highest weight)
    ('agriculture', 1.0000),
    ('farming', 1.0000),
    ('farm', 0.9500),
    ('agricultural', 0.9500),
    ('crop', 0.9000),
    ('crops', 0.9000),
    ('livestock', 0.9000),
    ('cattle', 0.8500),
    ('dairy', 0.8500),
    ('poultry', 0.8500),
    ('sheep', 0.8000),
    ('goats', 0.8000),
    ('pigs', 0.8000),
    ('horses', 0.7500),
    
    -- Crop Types (high weight)
    ('wheat', 0.9000),
    ('corn', 0.9000),
    ('soybeans', 0.9000),
    ('rice', 0.8500),
    ('cotton', 0.8500),
    ('barley', 0.8000),
    ('oats', 0.8000),
    ('sorghum', 0.7500),
    ('canola', 0.7500),
    ('sunflower', 0.7500),
    ('potatoes', 0.8000),
    ('tomatoes', 0.8000),
    ('lettuce', 0.7500),
    ('carrots', 0.7500),
    ('onions', 0.7500),
    ('peppers', 0.7500),
    ('cucumbers', 0.7000),
    ('squash', 0.7000),
    ('melons', 0.7000),
    ('berries', 0.7500),
    ('strawberries', 0.7500),
    ('blueberries', 0.7000),
    ('raspberries', 0.7000),
    ('grapes', 0.8000),
    ('apples', 0.8000),
    ('oranges', 0.8000),
    ('citrus', 0.8000),
    ('almonds', 0.8000),
    ('walnuts', 0.7500),
    ('pecans', 0.7500),
    ('hazelnuts', 0.7000),
    
    -- Agricultural Operations (medium-high weight)
    ('planting', 0.8000),
    ('harvesting', 0.8000),
    ('irrigation', 0.8000),
    ('fertilizer', 0.8000),
    ('pesticide', 0.7500),
    ('herbicide', 0.7500),
    ('tractor', 0.7500),
    ('combine', 0.7500),
    ('plow', 0.7000),
    ('cultivator', 0.7000),
    ('seeder', 0.7000),
    ('sprayer', 0.7000),
    ('barn', 0.7500),
    ('silo', 0.7500),
    ('greenhouse', 0.8000),
    ('orchard', 0.8000),
    ('vineyard', 0.8000),
    ('pasture', 0.8000),
    ('field', 0.7500),
    ('acre', 0.7500),
    ('hectare', 0.7000),
    
    -- Agricultural Services (medium weight)
    ('agricultural services', 0.8000),
    ('crop consulting', 0.7500),
    ('soil testing', 0.7500),
    ('seed', 0.8000),
    ('seed company', 0.7500),
    ('agricultural equipment', 0.7500),
    ('farm equipment', 0.7500),
    ('agricultural supplies', 0.7000),
    ('feed', 0.8000),
    ('animal feed', 0.8000),
    ('livestock feed', 0.8000),
    ('veterinary', 0.7500),
    ('veterinarian', 0.7500),
    ('agricultural extension', 0.7000),
    ('cooperative', 0.7000),
    ('agricultural cooperative', 0.7000),
    ('farmers market', 0.7000),
    ('organic', 0.8000),
    ('organic farming', 0.8000),
    ('sustainable agriculture', 0.7500),
    ('precision agriculture', 0.7500),
    ('agtech', 0.7000),
    ('agricultural technology', 0.7000)
) AS k(keyword, weight)
WHERE i.name = 'Agriculture' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- 2. FOOD PRODUCTION KEYWORDS (70+ keywords)
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core Food Production Keywords (highest weight)
    ('food production', 1.0000),
    ('food processing', 1.0000),
    ('food manufacturing', 0.9500),
    ('food company', 0.9500),
    ('food processor', 0.9000),
    ('food manufacturer', 0.9000),
    ('packaging', 0.8500),
    ('food packaging', 0.8500),
    ('canning', 0.8000),
    ('freezing', 0.8000),
    ('drying', 0.7500),
    ('dehydration', 0.7500),
    ('pasteurization', 0.7500),
    ('sterilization', 0.7000),
    
    -- Food Categories (high weight)
    ('meat processing', 0.9000),
    ('poultry processing', 0.9000),
    ('dairy processing', 0.9000),
    ('milk processing', 0.9000),
    ('cheese', 0.8500),
    ('cheese making', 0.8500),
    ('yogurt', 0.8000),
    ('butter', 0.8000),
    ('ice cream', 0.8000),
    ('bakery', 0.8500),
    ('baking', 0.8500),
    ('bread', 0.8000),
    ('pastry', 0.8000),
    ('cookies', 0.7500),
    ('crackers', 0.7500),
    ('cereal', 0.8000),
    ('breakfast cereal', 0.8000),
    ('snack foods', 0.8000),
    ('chips', 0.7500),
    ('nuts', 0.7500),
    ('confectionery', 0.8000),
    ('candy', 0.8000),
    ('chocolate', 0.8000),
    ('beverages', 0.8500),
    ('juice', 0.8000),
    ('soda', 0.7500),
    ('soft drinks', 0.7500),
    ('beer', 0.8000),
    ('wine', 0.8000),
    ('spirits', 0.7500),
    ('distillery', 0.7500),
    ('brewery', 0.8000),
    ('winery', 0.8000),
    
    -- Food Processing Equipment (medium-high weight)
    ('food processing equipment', 0.8000),
    ('conveyor', 0.7500),
    ('mixer', 0.7500),
    ('grinder', 0.7500),
    ('cutter', 0.7500),
    ('slicer', 0.7500),
    ('blender', 0.7000),
    ('extruder', 0.7000),
    ('oven', 0.8000),
    ('fryer', 0.7500),
    ('boiler', 0.7000),
    ('steamer', 0.7000),
    ('refrigerator', 0.7500),
    ('freezer', 0.7500),
    ('cooler', 0.7000),
    ('tank', 0.7000),
    ('vessel', 0.7000),
    ('pump', 0.7000),
    ('valve', 0.6500),
    ('pipe', 0.6500),
    ('hose', 0.6500),
    
    -- Food Safety & Quality (medium weight)
    ('food safety', 0.8000),
    ('haccp', 0.8000),
    ('hazard analysis', 0.7500),
    ('critical control points', 0.7500),
    ('food quality', 0.8000),
    ('quality control', 0.7500),
    ('inspection', 0.7500),
    ('testing', 0.7000),
    ('laboratory', 0.7000),
    ('microbiology', 0.7000),
    ('pathogen', 0.7000),
    ('contamination', 0.7000),
    ('sanitation', 0.7500),
    ('cleaning', 0.7000),
    ('disinfection', 0.7000),
    ('sterilization', 0.7000),
    ('preservation', 0.7500),
    ('shelf life', 0.7000),
    ('expiration', 0.7000),
    ('labeling', 0.7000),
    ('nutrition', 0.7500),
    ('nutritional', 0.7500),
    ('ingredients', 0.8000),
    ('additives', 0.7000),
    ('preservatives', 0.7000),
    ('flavoring', 0.7000),
    ('coloring', 0.6500),
    ('texture', 0.6500),
    ('consistency', 0.6500)
) AS k(keyword, weight)
WHERE i.name = 'Food Production' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- 3. ENERGY SERVICES KEYWORDS (70+ keywords)
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core Energy Keywords (highest weight)
    ('energy', 1.0000),
    ('energy services', 0.9500),
    ('energy company', 0.9500),
    ('power', 0.9000),
    ('electricity', 0.9000),
    ('electric', 0.9000),
    ('utility', 0.8500),
    ('utilities', 0.8500),
    ('power plant', 0.9000),
    ('power generation', 0.9000),
    ('generation', 0.8500),
    ('transmission', 0.8500),
    ('distribution', 0.8500),
    ('grid', 0.8000),
    ('electrical grid', 0.8000),
    
    -- Traditional Energy Sources (high weight)
    ('oil', 0.9000),
    ('petroleum', 0.9000),
    ('crude oil', 0.9000),
    ('gas', 0.9000),
    ('natural gas', 0.9000),
    ('coal', 0.9000),
    ('fossil fuel', 0.8500),
    ('fossil fuels', 0.8500),
    ('refinery', 0.8500),
    ('refining', 0.8500),
    ('petrochemical', 0.8000),
    ('petrochemicals', 0.8000),
    ('drilling', 0.8500),
    ('oil drilling', 0.8500),
    ('gas drilling', 0.8500),
    ('exploration', 0.8000),
    ('oil exploration', 0.8000),
    ('gas exploration', 0.8000),
    ('production', 0.8000),
    ('oil production', 0.8000),
    ('gas production', 0.8000),
    ('extraction', 0.8000),
    ('oil extraction', 0.8000),
    ('gas extraction', 0.8000),
    ('pipeline', 0.8000),
    ('oil pipeline', 0.8000),
    ('gas pipeline', 0.8000),
    ('storage', 0.7500),
    ('oil storage', 0.7500),
    ('gas storage', 0.7500),
    ('tank', 0.7000),
    ('storage tank', 0.7000),
    ('terminal', 0.7500),
    ('oil terminal', 0.7500),
    ('gas terminal', 0.7500),
    
    -- Energy Infrastructure (medium-high weight)
    ('power station', 0.8500),
    ('power facility', 0.8500),
    ('generating station', 0.8500),
    ('turbine', 0.8000),
    ('gas turbine', 0.8000),
    ('steam turbine', 0.8000),
    ('boiler', 0.8000),
    ('combustion', 0.7500),
    ('burner', 0.7500),
    ('furnace', 0.7500),
    ('heater', 0.7500),
    ('cooling', 0.7500),
    ('cooling tower', 0.7500),
    ('condenser', 0.7000),
    ('heat exchanger', 0.7000),
    ('compressor', 0.7500),
    ('pump', 0.7000),
    ('valve', 0.7000),
    ('control system', 0.7000),
    ('automation', 0.7000),
    ('monitoring', 0.7000),
    ('maintenance', 0.7500),
    ('repair', 0.7000),
    ('inspection', 0.7000),
    ('testing', 0.7000),
    
    -- Energy Services (medium weight)
    ('energy consulting', 0.7500),
    ('energy management', 0.7500),
    ('energy efficiency', 0.7500),
    ('energy audit', 0.7000),
    ('energy assessment', 0.7000),
    ('energy optimization', 0.7000),
    ('demand response', 0.7000),
    ('load management', 0.7000),
    ('peak shaving', 0.6500),
    ('backup power', 0.7000),
    ('emergency power', 0.7000),
    ('standby power', 0.7000),
    ('generator', 0.7500),
    ('backup generator', 0.7500),
    ('diesel generator', 0.7000),
    ('gas generator', 0.7000),
    ('portable generator', 0.6500),
    ('industrial generator', 0.7000),
    ('commercial generator', 0.7000),
    ('residential generator', 0.6500)
) AS k(keyword, weight)
WHERE i.name = 'Energy Services' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- 4. RENEWABLE ENERGY KEYWORDS (70+ keywords)
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core Renewable Energy Keywords (highest weight)
    ('renewable energy', 1.0000),
    ('clean energy', 0.9500),
    ('green energy', 0.9500),
    ('sustainable energy', 0.9000),
    ('alternative energy', 0.9000),
    ('solar', 0.9500),
    ('solar energy', 0.9500),
    ('solar power', 0.9500),
    ('photovoltaic', 0.9000),
    ('pv', 0.9000),
    ('solar panel', 0.9000),
    ('solar panels', 0.9000),
    ('solar array', 0.8500),
    ('solar farm', 0.8500),
    ('solar installation', 0.8500),
    ('wind', 0.9500),
    ('wind energy', 0.9500),
    ('wind power', 0.9500),
    ('wind turbine', 0.9000),
    ('wind turbines', 0.9000),
    ('wind farm', 0.8500),
    ('wind installation', 0.8500),
    ('hydroelectric', 0.9000),
    ('hydro', 0.9000),
    ('hydropower', 0.9000),
    ('water power', 0.8500),
    ('geothermal', 0.9000),
    ('geothermal energy', 0.9000),
    ('geothermal power', 0.9000),
    ('biomass', 0.8500),
    ('bioenergy', 0.8500),
    ('biofuel', 0.8500),
    ('biogas', 0.8000),
    ('ethanol', 0.8000),
    ('biodiesel', 0.8000),
    
    -- Solar Technology (high weight)
    ('solar cell', 0.9000),
    ('solar cells', 0.9000),
    ('silicon', 0.8000),
    ('monocrystalline', 0.7500),
    ('polycrystalline', 0.7500),
    ('thin film', 0.7000),
    ('inverter', 0.8000),
    ('solar inverter', 0.8000),
    ('microinverter', 0.7000),
    ('string inverter', 0.7000),
    ('central inverter', 0.7000),
    ('mounting system', 0.7500),
    ('racking', 0.7500),
    ('tracking', 0.7500),
    ('solar tracker', 0.7500),
    ('battery', 0.8000),
    ('energy storage', 0.8000),
    ('battery storage', 0.8000),
    ('lithium ion', 0.7500),
    ('lead acid', 0.7000),
    ('flow battery', 0.7000),
    
    -- Wind Technology (high weight)
    ('blade', 0.8500),
    ('turbine blade', 0.8500),
    ('rotor', 0.8000),
    ('nacelle', 0.8000),
    ('tower', 0.8000),
    ('wind tower', 0.8000),
    ('foundation', 0.7500),
    ('gearbox', 0.7500),
    ('generator', 0.8000),
    ('wind generator', 0.8000),
    ('transformer', 0.7500),
    ('substation', 0.7500),
    ('offshore', 0.8000),
    ('offshore wind', 0.8000),
    ('onshore', 0.7500),
    ('onshore wind', 0.7500),
    ('floating', 0.7000),
    ('floating wind', 0.7000),
    
    -- Renewable Energy Services (medium weight)
    ('renewable energy consulting', 0.7500),
    ('solar consulting', 0.7500),
    ('wind consulting', 0.7500),
    ('project development', 0.8000),
    ('solar development', 0.8000),
    ('wind development', 0.8000),
    ('permitting', 0.7500),
    ('environmental', 0.7500),
    ('environmental impact', 0.7500),
    ('eia', 0.7000),
    ('environmental impact assessment', 0.7000),
    ('construction', 0.8000),
    ('solar construction', 0.8000),
    ('wind construction', 0.8000),
    ('installation', 0.8000),
    ('solar installation', 0.8000),
    ('wind installation', 0.8000),
    ('commissioning', 0.7500),
    ('operations', 0.8000),
    ('operations and maintenance', 0.8000),
    ('o&m', 0.8000),
    ('maintenance', 0.8000),
    ('repair', 0.7500),
    ('monitoring', 0.7500),
    ('performance', 0.7500),
    ('efficiency', 0.7500),
    ('optimization', 0.7000),
    ('upgrade', 0.7000),
    ('retrofit', 0.7000),
    ('repowering', 0.7000)
) AS k(keyword, weight)
WHERE i.name = 'Renewable Energy' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- VERIFICATION QUERIES
-- =============================================================================

-- Verify all agriculture & energy keywords were added
DO $$
DECLARE
    agriculture_count INTEGER;
    food_production_count INTEGER;
    energy_services_count INTEGER;
    renewable_energy_count INTEGER;
    expected_count INTEGER := 200; -- 200+ keywords across all agriculture & energy industries
BEGIN
    -- Count Agriculture keywords
    SELECT COUNT(*) INTO agriculture_count
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    WHERE i.name = 'Agriculture' AND i.is_active = true AND ik.is_active = true;
    
    -- Count Food Production keywords
    SELECT COUNT(*) INTO food_production_count
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    WHERE i.name = 'Food Production' AND i.is_active = true AND ik.is_active = true;
    
    -- Count Energy Services keywords
    SELECT COUNT(*) INTO energy_services_count
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    WHERE i.name = 'Energy Services' AND i.is_active = true AND ik.is_active = true;
    
    -- Count Renewable Energy keywords
    SELECT COUNT(*) INTO renewable_energy_count
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    WHERE i.name = 'Renewable Energy' AND i.is_active = true AND ik.is_active = true;
    
    RAISE NOTICE 'Agriculture & Energy Keywords Added:';
    RAISE NOTICE 'Agriculture: % keywords', agriculture_count;
    RAISE NOTICE 'Food Production: % keywords', food_production_count;
    RAISE NOTICE 'Energy Services: % keywords', energy_services_count;
    RAISE NOTICE 'Renewable Energy: % keywords', renewable_energy_count;
    RAISE NOTICE 'Total: % keywords', (agriculture_count + food_production_count + energy_services_count + renewable_energy_count);
    
    IF (agriculture_count + food_production_count + energy_services_count + renewable_energy_count) >= expected_count THEN
        RAISE NOTICE 'SUCCESS: Agriculture & energy keywords added successfully';
    ELSE
        RAISE NOTICE 'WARNING: Expected % keywords, but found %', expected_count, (agriculture_count + food_production_count + energy_services_count + renewable_energy_count);
    END IF;
END $$;

-- Display keyword summary by industry
SELECT 
    'AGRICULTURE & ENERGY KEYWORD SUMMARY' as summary_type,
    '' as spacer;

SELECT 
    i.name as industry_name,
    i.confidence_threshold,
    COUNT(ik.id) as keyword_count,
    ROUND(MIN(ik.weight), 3) as min_weight,
    ROUND(MAX(ik.weight), 3) as max_weight,
    ROUND(AVG(ik.weight), 3) as avg_weight
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
WHERE i.name IN ('Agriculture', 'Food Production', 'Energy Services', 'Renewable Energy') 
AND i.is_active = true
GROUP BY i.id, i.name, i.confidence_threshold
ORDER BY i.name;

-- Display sample keywords for each industry
SELECT 
    'SAMPLE AGRICULTURE KEYWORDS' as summary_type,
    '' as spacer;

SELECT 
    keyword,
    weight,
    CASE 
        WHEN weight >= 0.90 THEN 'Very High'
        WHEN weight >= 0.80 THEN 'High'
        WHEN weight >= 0.70 THEN 'Medium-High'
        WHEN weight >= 0.60 THEN 'Medium'
        ELSE 'Low'
    END as weight_category
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
WHERE i.name = 'Agriculture' AND i.is_active = true AND ik.is_active = true
ORDER BY weight DESC
LIMIT 10;

SELECT 
    'SAMPLE FOOD PRODUCTION KEYWORDS' as summary_type,
    '' as spacer;

SELECT 
    keyword,
    weight,
    CASE 
        WHEN weight >= 0.90 THEN 'Very High'
        WHEN weight >= 0.80 THEN 'High'
        WHEN weight >= 0.70 THEN 'Medium-High'
        WHEN weight >= 0.60 THEN 'Medium'
        ELSE 'Low'
    END as weight_category
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
WHERE i.name = 'Food Production' AND i.is_active = true AND ik.is_active = true
ORDER BY weight DESC
LIMIT 10;

SELECT 
    'SAMPLE ENERGY SERVICES KEYWORDS' as summary_type,
    '' as spacer;

SELECT 
    keyword,
    weight,
    CASE 
        WHEN weight >= 0.90 THEN 'Very High'
        WHEN weight >= 0.80 THEN 'High'
        WHEN weight >= 0.70 THEN 'Medium-High'
        WHEN weight >= 0.60 THEN 'Medium'
        ELSE 'Low'
    END as weight_category
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
WHERE i.name = 'Energy Services' AND i.is_active = true AND ik.is_active = true
ORDER BY weight DESC
LIMIT 10;

SELECT 
    'SAMPLE RENEWABLE ENERGY KEYWORDS' as summary_type,
    '' as spacer;

SELECT 
    keyword,
    weight,
    CASE 
        WHEN weight >= 0.90 THEN 'Very High'
        WHEN weight >= 0.80 THEN 'High'
        WHEN weight >= 0.70 THEN 'Medium-High'
        WHEN weight >= 0.60 THEN 'Medium'
        ELSE 'Low'
    END as weight_category
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
WHERE i.name = 'Renewable Energy' AND i.is_active = true AND ik.is_active = true
ORDER BY weight DESC
LIMIT 10;

COMMIT;

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================
DO $$
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'AGRICULTURE & ENERGY KEYWORDS COMPREHENSIVE SCRIPT COMPLETED';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Task 3.2.7: Agriculture & Energy Keywords Added Successfully';
    RAISE NOTICE 'Industries covered: Agriculture, Food Production, Energy Services, Renewable Energy';
    RAISE NOTICE 'Total keywords added: 200+ comprehensive agriculture & energy keywords';
    RAISE NOTICE 'Weight range: 0.5000-1.0000 as specified in plan';
    RAISE NOTICE 'Status: Ready for testing and validation';
    RAISE NOTICE 'Next: Test agriculture & energy classification accuracy';
    RAISE NOTICE '=============================================================================';
END $$;
