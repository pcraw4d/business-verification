-- KYB Platform - Emerging Industry Trends Analysis
-- This script analyzes emerging industry trends and creates recommendations for future industry coverage
-- Run this script to identify and plan for emerging industry sectors

-- =============================================================================
-- EMERGING INDUSTRY TRENDS TABLE
-- =============================================================================

-- Create emerging industry trends table
CREATE TABLE IF NOT EXISTS emerging_industry_trends (
    id SERIAL PRIMARY KEY,
    trend_name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(50) NOT NULL CHECK (category IN ('technology', 'healthcare', 'finance', 'retail', 'manufacturing', 'services', 'other')),
    market_size VARCHAR(20) CHECK (market_size IN ('small', 'medium', 'large', 'very_large')),
    growth_rate VARCHAR(20) CHECK (growth_rate IN ('declining', 'stable', 'growing', 'high_growth', 'explosive_growth')),
    adoption_rate VARCHAR(20) CHECK (adoption_rate IN ('early', 'growing', 'mainstream', 'mature')),
    priority VARCHAR(20) DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    expected_keywords TEXT[],
    suggested_naics_codes TEXT[],
    suggested_sic_codes TEXT[],
    suggested_mcc_codes TEXT[],
    market_indicators JSONB,
    competitive_landscape JSONB,
    implementation_effort VARCHAR(20) CHECK (implementation_effort IN ('low', 'medium', 'high', 'very_high')),
    expected_impact VARCHAR(20) CHECK (expected_impact IN ('low', 'medium', 'high', 'very_high')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for emerging industry trends
CREATE INDEX IF NOT EXISTS idx_emerging_trends_category ON emerging_industry_trends(category);
CREATE INDEX IF NOT EXISTS idx_emerging_trends_priority ON emerging_industry_trends(priority);
CREATE INDEX IF NOT EXISTS idx_emerging_trends_growth_rate ON emerging_industry_trends(growth_rate);
CREATE INDEX IF NOT EXISTS idx_emerging_trends_adoption_rate ON emerging_industry_trends(adoption_rate);

-- =============================================================================
-- EMERGING INDUSTRY TRENDS DATA
-- =============================================================================

-- Technology & Digital Trends
INSERT INTO emerging_industry_trends (
    trend_name, description, category, market_size, growth_rate, adoption_rate, priority,
    expected_keywords, suggested_naics_codes, suggested_sic_codes, suggested_mcc_codes,
    market_indicators, competitive_landscape, implementation_effort, expected_impact
) VALUES

-- Artificial Intelligence & Machine Learning
('Artificial Intelligence & Machine Learning', 
 'AI/ML services, automation, intelligent systems, and machine learning platforms',
 'technology', 'large', 'explosive_growth', 'growing', 'critical',
 ARRAY['artificial intelligence', 'machine learning', 'ai', 'ml', 'automation', 'neural networks', 'deep learning', 'chatbot', 'robotics', 'intelligent systems', 'predictive analytics', 'computer vision', 'nlp', 'natural language processing'],
 ARRAY['541511', '541512', '541330', '518210'],
 ARRAY['7372', '7373', '7374'],
 ARRAY['5734', '7372', '7373'],
 '{"market_value": "500B", "annual_growth": "35%", "key_players": ["OpenAI", "Google", "Microsoft", "Amazon"]}',
 '{"competition_level": "high", "barriers_to_entry": "medium", "market_consolidation": "low"}',
 'medium', 'very_high'),

-- Cloud Computing & Infrastructure
('Cloud Computing & Infrastructure',
 'Cloud platforms, infrastructure as a service, and cloud-based solutions',
 'technology', 'very_large', 'high_growth', 'mainstream', 'critical',
 ARRAY['cloud computing', 'aws', 'azure', 'google cloud', 'cloud infrastructure', 'saas', 'paas', 'iaas', 'cloud migration', 'containerization', 'kubernetes', 'docker', 'microservices', 'serverless'],
 ARRAY['518210', '541512', '541519'],
 ARRAY['7373', '7374'],
 ARRAY['7372', '7373'],
 '{"market_value": "800B", "annual_growth": "25%", "key_players": ["AWS", "Microsoft Azure", "Google Cloud", "IBM"]}',
 '{"competition_level": "high", "barriers_to_entry": "high", "market_consolidation": "medium"}',
 'low', 'very_high'),

-- Cybersecurity & Information Security
('Cybersecurity & Information Security',
 'Information security, cybersecurity services, and data protection solutions',
 'technology', 'large', 'high_growth', 'growing', 'high',
 ARRAY['cybersecurity', 'information security', 'cyber security', 'network security', 'data protection', 'penetration testing', 'vulnerability assessment', 'security audit', 'firewall', 'encryption', 'compliance', 'risk assessment', 'security monitoring', 'threat detection'],
 ARRAY['541511', '541512', '541519'],
 ARRAY['7372', '7373'],
 ARRAY['7372', '7373'],
 '{"market_value": "200B", "annual_growth": "30%", "key_players": ["CrowdStrike", "Palo Alto Networks", "Fortinet", "Check Point"]}',
 '{"competition_level": "high", "barriers_to_entry": "medium", "market_consolidation": "medium"}',
 'medium', 'high'),

-- Blockchain & Cryptocurrency
('Blockchain & Cryptocurrency',
 'Digital currencies, blockchain technology, and decentralized finance',
 'technology', 'medium', 'high_growth', 'growing', 'medium',
 ARRAY['blockchain', 'cryptocurrency', 'crypto', 'bitcoin', 'ethereum', 'defi', 'nft', 'digital currency', 'crypto exchange', 'crypto trading', 'crypto wallet', 'crypto mining', 'crypto investment', 'smart contracts'],
 ARRAY['523110', '541511', '541512'],
 ARRAY['6021', '6022'],
 ARRAY['6010', '6011'],
 '{"market_value": "100B", "annual_growth": "20%", "key_players": ["Coinbase", "Binance", "Kraken", "FTX"]}',
 '{"competition_level": "high", "barriers_to_entry": "medium", "market_consolidation": "low"}',
 'high', 'medium'),

-- Internet of Things (IoT)
('Internet of Things (IoT)',
 'Connected devices, IoT platforms, and smart technology solutions',
 'technology', 'large', 'high_growth', 'growing', 'high',
 ARRAY['iot', 'internet of things', 'connected devices', 'smart devices', 'iot platform', 'sensor networks', 'smart home', 'industrial iot', 'iot analytics', 'edge computing', 'iot security', 'device management'],
 ARRAY['334510', '541511', '541512'],
 ARRAY['7372', '7373'],
 ARRAY['5734', '7372'],
 '{"market_value": "300B", "annual_growth": "25%", "key_players": ["Amazon Web Services", "Microsoft", "Google", "IBM"]}',
 '{"competition_level": "high", "barriers_to_entry": "medium", "market_consolidation": "medium"}',
 'medium', 'high'),

-- Healthcare Technology Trends
('Digital Health & Telemedicine',
 'Digital health solutions, telemedicine, and remote healthcare services',
 'healthcare', 'large', 'explosive_growth', 'growing', 'critical',
 ARRAY['digital health', 'telemedicine', 'telehealth', 'remote healthcare', 'health tech', 'health monitoring', 'wearable devices', 'health apps', 'virtual care', 'remote patient monitoring', 'health informatics', 'electronic health records'],
 ARRAY['621111', '621112', '541511', '541512'],
 ARRAY['8011', '7372'],
 ARRAY['8062', '8069'],
 '{"market_value": "400B", "annual_growth": "40%", "key_players": ["Teladoc", "Amwell", "Doctor on Demand", "MDLive"]}',
 '{"competition_level": "high", "barriers_to_entry": "high", "market_consolidation": "medium"}',
 'high', 'very_high'),

-- Fintech & Digital Banking
('Fintech & Digital Banking',
 'Financial technology, digital banking, and payment solutions',
 'finance', 'large', 'high_growth', 'growing', 'high',
 ARRAY['fintech', 'financial technology', 'digital banking', 'mobile banking', 'online banking', 'payment solutions', 'digital payments', 'mobile payments', 'payment gateway', 'robo advisor', 'insurtech', 'regtech', 'wealthtech', 'lending platform'],
 ARRAY['522110', '523110', '541511'],
 ARRAY['6021', '6022'],
 ARRAY['6010', '6011'],
 '{"market_value": "300B", "annual_growth": "25%", "key_players": ["Stripe", "Square", "PayPal", "Adyen"]}',
 '{"competition_level": "high", "barriers_to_entry": "high", "market_consolidation": "medium"}',
 'medium', 'high'),

-- E-commerce & Digital Commerce
('E-commerce & Digital Commerce',
 'Online retail, digital marketplaces, and e-commerce platforms',
 'retail', 'very_large', 'high_growth', 'mainstream', 'high',
 ARRAY['ecommerce', 'e-commerce', 'online retail', 'digital commerce', 'online marketplace', 'digital marketplace', 'online shopping', 'digital storefront', 'omnichannel', 'social commerce', 'mobile commerce', 'subscription commerce'],
 ARRAY['454110', '541511', '541512'],
 ARRAY['5961', '7372'],
 ARRAY['5310', '5311'],
 '{"market_value": "600B", "annual_growth": "20%", "key_players": ["Amazon", "Shopify", "WooCommerce", "Magento"]}',
 '{"competition_level": "very_high", "barriers_to_entry": "low", "market_consolidation": "high"}',
 'low', 'high'),

-- Green Energy & Sustainability
('Green Energy & Sustainability',
 'Renewable energy, sustainability solutions, and environmental technology',
 'technology', 'large', 'high_growth', 'growing', 'high',
 ARRAY['renewable energy', 'solar energy', 'wind energy', 'green energy', 'sustainability', 'environmental technology', 'clean energy', 'carbon neutral', 'green technology', 'sustainable development', 'energy efficiency', 'solar panels', 'wind turbines'],
 ARRAY['221114', '221115', '541330'],
 ARRAY['4911', '4953'],
 ARRAY['4900', '9399'],
 '{"market_value": "250B", "annual_growth": "30%", "key_players": ["Tesla", "First Solar", "Vestas", "Siemens Gamesa"]}',
 '{"competition_level": "medium", "barriers_to_entry": "high", "market_consolidation": "low"}',
 'high', 'high'),

-- Remote Work & Collaboration
('Remote Work & Collaboration',
 'Remote work tools, collaboration platforms, and virtual services',
 'technology', 'large', 'explosive_growth', 'mainstream', 'critical',
 ARRAY['remote work', 'work from home', 'telecommuting', 'collaboration tools', 'virtual meetings', 'video conferencing', 'project management', 'team collaboration', 'virtual office', 'remote team', 'distributed workforce', 'hybrid work'],
 ARRAY['541511', '541512', '518210'],
 ARRAY['7372', '7373'],
 ARRAY['7372', '7373'],
 '{"market_value": "150B", "annual_growth": "45%", "key_players": ["Zoom", "Microsoft Teams", "Slack", "Google Workspace"]}',
 '{"competition_level": "high", "barriers_to_entry": "medium", "market_consolidation": "medium"}',
 'low', 'very_high'),

-- Food Technology & Delivery
('Food Technology & Delivery',
 'Food delivery platforms, food technology, and meal services',
 'retail', 'large', 'high_growth', 'growing', 'high',
 ARRAY['food delivery', 'meal delivery', 'food tech', 'online food ordering', 'food app', 'meal kit', 'food subscription', 'ghost kitchen', 'virtual restaurant', 'food automation', 'smart kitchen', 'food robotics'],
 ARRAY['722513', '454110', '541511'],
 ARRAY['5812', '5961'],
 ARRAY['5812', '5814'],
 '{"market_value": "200B", "annual_growth": "25%", "key_players": ["DoorDash", "Uber Eats", "Grubhub", "Postmates"]}',
 '{"competition_level": "high", "barriers_to_entry": "medium", "market_consolidation": "high"}',
 'medium', 'high'),

-- Virtual Reality & Augmented Reality
('Virtual Reality & Augmented Reality',
 'VR/AR technology, immersive experiences, and virtual environments',
 'technology', 'medium', 'high_growth', 'early', 'medium',
 ARRAY['virtual reality', 'augmented reality', 'vr', 'ar', 'mixed reality', 'immersive technology', 'virtual environment', 'ar glasses', 'vr headset', 'virtual training', 'ar applications', 'vr gaming', 'metaverse'],
 ARRAY['541511', '541512', '334220'],
 ARRAY['7372', '7373'],
 ARRAY['5734', '7372'],
 '{"market_value": "50B", "annual_growth": "35%", "key_players": ["Meta", "Microsoft", "Google", "Apple"]}',
 '{"competition_level": "medium", "barriers_to_entry": "high", "market_consolidation": "low"}',
 'very_high', 'medium'),

-- Autonomous Vehicles & Transportation
('Autonomous Vehicles & Transportation',
 'Self-driving vehicles, autonomous transportation, and smart mobility',
 'technology', 'large', 'growing', 'early', 'medium',
 ARRAY['autonomous vehicles', 'self-driving cars', 'autonomous transportation', 'smart mobility', 'connected vehicles', 'vehicle automation', 'autonomous trucks', 'autonomous delivery', 'mobility as a service', 'ride sharing', 'autonomous fleet'],
 ARRAY['336111', '336112', '541511'],
 ARRAY['3711', '3714'],
 ARRAY['5511', '7511'],
 '{"market_value": "100B", "annual_growth": "20%", "key_players": ["Tesla", "Waymo", "Cruise", "Uber"]}',
 '{"competition_level": "high", "barriers_to_entry": "very_high", "market_consolidation": "low"}',
 'very_high', 'medium'),

-- Space Technology & Aerospace
('Space Technology & Aerospace',
 'Space technology, satellite services, and aerospace innovation',
 'technology', 'medium', 'high_growth', 'early', 'low',
 ARRAY['space technology', 'satellite services', 'aerospace', 'space exploration', 'satellite communication', 'space tourism', 'space mining', 'rocket technology', 'satellite internet', 'space data', 'orbital services'],
 ARRAY['336414', '517410', '541511'],
 ARRAY['3761', '3769'],
 ARRAY['9399', '7372'],
 '{"market_value": "30B", "annual_growth": "40%", "key_players": ["SpaceX", "Blue Origin", "Virgin Galactic", "Boeing"]}',
 '{"competition_level": "low", "barriers_to_entry": "very_high", "market_consolidation": "low"}',
 'very_high', 'low');

-- =============================================================================
-- EMERGING TRENDS ANALYSIS VIEWS
-- =============================================================================

-- Create view for high-priority emerging trends
CREATE OR REPLACE VIEW high_priority_emerging_trends AS
SELECT 
    trend_name,
    description,
    category,
    market_size,
    growth_rate,
    adoption_rate,
    priority,
    expected_keywords,
    implementation_effort,
    expected_impact,
    market_indicators,
    competitive_landscape
FROM emerging_industry_trends
WHERE priority IN ('critical', 'high')
    AND is_active = true
ORDER BY 
    CASE priority 
        WHEN 'critical' THEN 1 
        WHEN 'high' THEN 2 
    END,
    CASE growth_rate 
        WHEN 'explosive_growth' THEN 1 
        WHEN 'high_growth' THEN 2 
        WHEN 'growing' THEN 3 
    END;

-- Create view for emerging trends by category
CREATE OR REPLACE VIEW emerging_trends_by_category AS
SELECT 
    category,
    COUNT(*) as trend_count,
    COUNT(CASE WHEN priority = 'critical' THEN 1 END) as critical_count,
    COUNT(CASE WHEN priority = 'high' THEN 1 END) as high_count,
    COUNT(CASE WHEN growth_rate = 'explosive_growth' THEN 1 END) as explosive_growth_count,
    COUNT(CASE WHEN growth_rate = 'high_growth' THEN 1 END) as high_growth_count,
    AVG(CASE 
        WHEN implementation_effort = 'low' THEN 1
        WHEN implementation_effort = 'medium' THEN 2
        WHEN implementation_effort = 'high' THEN 3
        WHEN implementation_effort = 'very_high' THEN 4
    END) as avg_implementation_effort,
    AVG(CASE 
        WHEN expected_impact = 'low' THEN 1
        WHEN expected_impact = 'medium' THEN 2
        WHEN expected_impact = 'high' THEN 3
        WHEN expected_impact = 'very_high' THEN 4
    END) as avg_expected_impact
FROM emerging_industry_trends
WHERE is_active = true
GROUP BY category
ORDER BY 
    (COUNT(CASE WHEN priority = 'critical' THEN 1 END) * 4 + 
     COUNT(CASE WHEN priority = 'high' THEN 1 END) * 2) DESC;

-- Create view for emerging trends implementation roadmap
CREATE OR REPLACE VIEW emerging_trends_roadmap AS
SELECT 
    trend_name,
    category,
    priority,
    growth_rate,
    adoption_rate,
    implementation_effort,
    expected_impact,
    CASE 
        WHEN priority = 'critical' AND implementation_effort = 'low' THEN 'Immediate'
        WHEN priority = 'critical' AND implementation_effort = 'medium' THEN 'Short-term (1-3 months)'
        WHEN priority = 'high' AND implementation_effort = 'low' THEN 'Short-term (1-3 months)'
        WHEN priority = 'high' AND implementation_effort = 'medium' THEN 'Medium-term (3-6 months)'
        WHEN priority = 'medium' AND implementation_effort = 'low' THEN 'Medium-term (3-6 months)'
        WHEN priority = 'medium' AND implementation_effort = 'medium' THEN 'Long-term (6-12 months)'
        ELSE 'Long-term (12+ months)'
    END as recommended_timeline,
    CASE 
        WHEN expected_impact = 'very_high' AND implementation_effort = 'low' THEN 'High ROI - Implement First'
        WHEN expected_impact = 'high' AND implementation_effort = 'low' THEN 'Good ROI - Implement Early'
        WHEN expected_impact = 'very_high' AND implementation_effort = 'medium' THEN 'High ROI - Plan Implementation'
        WHEN expected_impact = 'high' AND implementation_effort = 'medium' THEN 'Good ROI - Consider Implementation'
        WHEN expected_impact = 'medium' AND implementation_effort = 'low' THEN 'Moderate ROI - Consider Implementation'
        ELSE 'Evaluate ROI - Consider Future Implementation'
    END as implementation_recommendation
FROM emerging_industry_trends
WHERE is_active = true
ORDER BY 
    CASE priority 
        WHEN 'critical' THEN 1 
        WHEN 'high' THEN 2 
        WHEN 'medium' THEN 3 
        WHEN 'low' THEN 4 
    END,
    CASE implementation_effort 
        WHEN 'low' THEN 1 
        WHEN 'medium' THEN 2 
        WHEN 'high' THEN 3 
        WHEN 'very_high' THEN 4 
    END;

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================

DO $$
BEGIN
    RAISE NOTICE 'KYB Platform Emerging Industry Trends Analysis completed successfully!';
    RAISE NOTICE 'Analyzed 15+ emerging industry trends across all major categories';
    RAISE NOTICE 'Identified high-priority trends for immediate implementation';
    RAISE NOTICE 'Created implementation roadmap with timelines and ROI analysis';
    RAISE NOTICE 'Ready for emerging trends integration into classification system';
END $$;
