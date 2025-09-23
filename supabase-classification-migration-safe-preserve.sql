-- KYB Platform - Classification System Database Migration (SAFE - PRESERVE DATA)
-- This script fixes column issues while preserving existing data

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- =============================================================================
-- SAFE TABLE CREATION - ONLY CREATE IF NOT EXISTS
-- =============================================================================

-- Create industries table only if it doesn't exist
CREATE TABLE IF NOT EXISTS industries (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    category VARCHAR(100),
    confidence_threshold DECIMAL(3,2) DEFAULT 0.50,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create industry_keywords table only if it doesn't exist
CREATE TABLE IF NOT EXISTS industry_keywords (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    keyword VARCHAR(255) NOT NULL,
    weight DECIMAL(5,4) DEFAULT 1.0000,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(industry_id, keyword)
);

-- Create classification_codes table only if it doesn't exist
CREATE TABLE IF NOT EXISTS classification_codes (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    code_type VARCHAR(10) NOT NULL CHECK (code_type IN ('NAICS', 'SIC', 'MCC')),
    code VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(industry_id, code_type, code)
);

-- Create industry_patterns table only if it doesn't exist
CREATE TABLE IF NOT EXISTS industry_patterns (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    pattern_type VARCHAR(50) NOT NULL CHECK (pattern_type IN ('regex', 'keyword', 'phrase', 'domain')),
    pattern_value TEXT NOT NULL,
    confidence_weight DECIMAL(3,2) DEFAULT 1.00 CHECK (confidence_weight >= 0.00 AND confidence_weight <= 2.00),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(industry_id, pattern_type, pattern_value)
);

-- Create keyword_weights table only if it doesn't exist
CREATE TABLE IF NOT EXISTS keyword_weights (
    id SERIAL PRIMARY KEY,
    keyword VARCHAR(255) NOT NULL UNIQUE,
    weight DECIMAL(5,4) DEFAULT 1.0000 CHECK (weight >= 0.0000 AND weight <= 5.0000),
    usage_count INTEGER DEFAULT 0,
    last_used TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create classification_accuracy_metrics table only if it doesn't exist
CREATE TABLE IF NOT EXISTS classification_accuracy_metrics (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    total_classifications INTEGER DEFAULT 0,
    correct_classifications INTEGER DEFAULT 0,
    accuracy_percentage DECIMAL(5,2) DEFAULT 0.00,
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(industry_id)
);

-- =============================================================================
-- SAFE COLUMN ADDITIONS - ADD MISSING COLUMNS IF THEY DON'T EXIST
-- =============================================================================

-- Add weight column to industry_keywords if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'industry_keywords' AND column_name = 'weight') THEN
        ALTER TABLE industry_keywords ADD COLUMN weight DECIMAL(5,4) DEFAULT 1.0000;
    END IF;
END $$;

-- Add confidence_weight column to industry_patterns if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'industry_patterns' AND column_name = 'confidence_weight') THEN
        ALTER TABLE industry_patterns ADD COLUMN confidence_weight DECIMAL(3,2) DEFAULT 1.00;
    END IF;
END $$;

-- Add weight column to keyword_weights if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'keyword_weights' AND column_name = 'weight') THEN
        ALTER TABLE keyword_weights ADD COLUMN weight DECIMAL(5,4) DEFAULT 1.0000;
    END IF;
END $$;

-- =============================================================================
-- SAFE INDEX CREATION - CREATE ONLY IF NOT EXISTS
-- =============================================================================

-- Industries table indexes
CREATE INDEX IF NOT EXISTS idx_industries_name ON industries(name);
CREATE INDEX IF NOT EXISTS idx_industries_category ON industries(category);
CREATE INDEX IF NOT EXISTS idx_industries_active ON industries(is_active);

-- Industry keywords table indexes
CREATE INDEX IF NOT EXISTS idx_industry_keywords_industry_id ON industry_keywords(industry_id);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_keyword ON industry_keywords(keyword);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_active ON industry_keywords(is_active);

-- Classification codes table indexes
CREATE INDEX IF NOT EXISTS idx_classification_codes_industry_id ON classification_codes(industry_id);
CREATE INDEX IF NOT EXISTS idx_classification_codes_type ON classification_codes(code_type);
CREATE INDEX IF NOT EXISTS idx_classification_codes_code ON classification_codes(code);
CREATE INDEX IF NOT EXISTS idx_classification_codes_active ON classification_codes(is_active);

-- Industry patterns table indexes
CREATE INDEX IF NOT EXISTS idx_industry_patterns_industry_id ON industry_patterns(industry_id);
CREATE INDEX IF NOT EXISTS idx_industry_patterns_type ON industry_patterns(pattern_type);
CREATE INDEX IF NOT EXISTS idx_industry_patterns_active ON industry_patterns(is_active);

-- Keyword weights table indexes
CREATE INDEX IF NOT EXISTS idx_keyword_weights_keyword ON keyword_weights(keyword);
CREATE INDEX IF NOT EXISTS idx_keyword_weights_weight ON keyword_weights(weight);

-- Classification accuracy metrics table indexes
CREATE INDEX IF NOT EXISTS idx_classification_accuracy_industry_id ON classification_accuracy_metrics(industry_id);
CREATE INDEX IF NOT EXISTS idx_classification_accuracy_percentage ON classification_accuracy_metrics(accuracy_percentage);

-- =============================================================================
-- SAFE DATA INSERTION - ONLY INSERT IF DATA DOESN'T EXIST
-- =============================================================================

-- Insert sample industries only if they don't exist
INSERT INTO industries (name, description, category, confidence_threshold) 
SELECT * FROM (VALUES
    ('Technology', 'Software development, IT services, and technology companies', 'Technology', 0.70),
    ('Retail', 'Consumer goods retail and e-commerce', 'Commerce', 0.60),
    ('Healthcare', 'Medical services, pharmaceuticals, and healthcare technology', 'Healthcare', 0.75),
    ('Finance', 'Banking, investment, and financial services', 'Finance', 0.80),
    ('Manufacturing', 'Industrial production and manufacturing', 'Industrial', 0.65),
    ('Food & Beverage', 'Restaurants, food production, and beverage companies', 'Consumer', 0.55),
    ('Real Estate', 'Property development, real estate services', 'Property', 0.60),
    ('Education', 'Educational institutions and training services', 'Education', 0.70),
    ('Transportation', 'Logistics, shipping, and transportation services', 'Logistics', 0.65),
    ('Entertainment', 'Media, entertainment, and creative services', 'Media', 0.60)
) AS v(name, description, category, confidence_threshold)
WHERE NOT EXISTS (SELECT 1 FROM industries WHERE industries.name = v.name);

-- Insert sample industry keywords only if they don't exist
INSERT INTO industry_keywords (industry_id, keyword, weight) 
SELECT * FROM (VALUES
    -- Technology keywords
    (1, 'software', 1.2),
    (1, 'technology', 1.1),
    (1, 'computer', 1.0),
    (1, 'tech', 1.0),
    (1, 'programming', 1.3),
    (1, 'development', 1.1),
    (1, 'IT', 1.0),
    (1, 'digital', 0.9),
    (1, 'app', 1.0),
    (1, 'platform', 1.0),
    
    -- Retail keywords
    (2, 'retail', 1.2),
    (2, 'store', 1.1),
    (2, 'shop', 1.0),
    (2, 'commerce', 1.1),
    (2, 'ecommerce', 1.2),
    (2, 'marketplace', 1.0),
    (2, 'sales', 0.9),
    (2, 'merchandise', 1.0),
    (2, 'products', 0.8),
    (2, 'goods', 0.8),
    
    -- Healthcare keywords
    (3, 'health', 1.2),
    (3, 'medical', 1.3),
    (3, 'healthcare', 1.2),
    (3, 'hospital', 1.1),
    (3, 'clinic', 1.0),
    (3, 'pharmacy', 1.0),
    (3, 'doctor', 1.0),
    (3, 'patient', 0.9),
    (3, 'treatment', 1.0),
    (3, 'medicine', 1.0),
    
    -- Finance keywords
    (4, 'finance', 1.2),
    (4, 'banking', 1.3),
    (4, 'financial', 1.1),
    (4, 'investment', 1.2),
    (4, 'credit', 1.0),
    (4, 'loan', 1.0),
    (4, 'insurance', 1.0),
    (4, 'trading', 1.0),
    (4, 'wealth', 1.0),
    (4, 'capital', 1.0)
) AS v(industry_id, keyword, weight)
WHERE NOT EXISTS (SELECT 1 FROM industry_keywords WHERE industry_keywords.industry_id = v.industry_id AND industry_keywords.keyword = v.keyword);

-- Insert sample classification codes only if they don't exist
INSERT INTO classification_codes (industry_id, code_type, code, description) 
SELECT * FROM (VALUES
    -- Technology codes
    (1, 'NAICS', '541511', 'Custom Computer Programming Services'),
    (1, 'NAICS', '541512', 'Computer Systems Design Services'),
    (1, 'NAICS', '541513', 'Computer Facilities Management Services'),
    (1, 'SIC', '7372', 'Computer Programming Services'),
    (1, 'SIC', '7373', 'Computer Integrated Systems Design'),
    (1, 'MCC', '7372', 'Computer Programming Services'),
    
    -- Retail codes
    (2, 'NAICS', '454110', 'Electronic Shopping and Mail-Order Houses'),
    (2, 'NAICS', '448140', 'Family Clothing Stores'),
    (2, 'SIC', '5961', 'Catalog and Mail-Order Houses'),
    (2, 'SIC', '5621', 'Women''s Clothing Stores'),
    (2, 'MCC', '5311', 'Department Stores'),
    
    -- Healthcare codes
    (3, 'NAICS', '621111', 'Offices of Physicians (except Mental Health Specialists)'),
    (3, 'NAICS', '622110', 'General Medical and Surgical Hospitals'),
    (3, 'SIC', '8011', 'Offices and Clinics of Doctors of Medicine'),
    (3, 'SIC', '8062', 'General Medical and Surgical Hospitals'),
    (3, 'MCC', '8062', 'Hospitals'),
    
    -- Finance codes
    (4, 'NAICS', '522110', 'Commercial Banking'),
    (4, 'NAICS', '523110', 'Investment Banking and Securities Dealing'),
    (4, 'SIC', '6021', 'National Commercial Banks'),
    (4, 'SIC', '6211', 'Security Brokers, Dealers, and Flotation Companies'),
    (4, 'MCC', '6010', 'Financial Institutions - Merchandise, Services')
) AS v(industry_id, code_type, code, description)
WHERE NOT EXISTS (SELECT 1 FROM classification_codes WHERE classification_codes.industry_id = v.industry_id AND classification_codes.code_type = v.code_type AND classification_codes.code = v.code);

-- Insert sample industry patterns only if they don't exist
INSERT INTO industry_patterns (industry_id, pattern_type, pattern_value, confidence_weight) 
SELECT * FROM (VALUES
    -- Technology patterns
    (1, 'domain', '\.com$', 0.8),
    (1, 'domain', '\.io$', 1.2),
    (1, 'domain', '\.tech$', 1.3),
    (1, 'keyword', 'software', 1.2),
    (1, 'keyword', 'technology', 1.1),
    (1, 'regex', '.*tech.*', 1.0),
    
    -- Retail patterns
    (2, 'domain', '\.store$', 1.2),
    (2, 'domain', '\.shop$', 1.1),
    (2, 'keyword', 'retail', 1.2),
    (2, 'keyword', 'commerce', 1.1),
    (2, 'regex', '.*shop.*', 1.0),
    
    -- Healthcare patterns
    (3, 'domain', '\.health$', 1.3),
    (3, 'domain', '\.medical$', 1.2),
    (3, 'keyword', 'health', 1.2),
    (3, 'keyword', 'medical', 1.3),
    (3, 'regex', '.*health.*', 1.0),
    
    -- Finance patterns
    (4, 'domain', '\.finance$', 1.2),
    (4, 'domain', '\.bank$', 1.3),
    (4, 'keyword', 'finance', 1.2),
    (4, 'keyword', 'banking', 1.3),
    (4, 'regex', '.*finance.*', 1.0)
) AS v(industry_id, pattern_type, pattern_value, confidence_weight)
WHERE NOT EXISTS (SELECT 1 FROM industry_patterns WHERE industry_patterns.industry_id = v.industry_id AND industry_patterns.pattern_type = v.pattern_type AND industry_patterns.pattern_value = v.pattern_value);

-- Insert sample keyword weights only if they don't exist
INSERT INTO keyword_weights (keyword, weight, usage_count) 
SELECT * FROM (VALUES
    ('software', 1.2, 0),
    ('technology', 1.1, 0),
    ('retail', 1.2, 0),
    ('health', 1.2, 0),
    ('finance', 1.2, 0),
    ('medical', 1.3, 0),
    ('banking', 1.3, 0),
    ('commerce', 1.1, 0),
    ('development', 1.1, 0),
    ('programming', 1.3, 0)
) AS v(keyword, weight, usage_count)
WHERE NOT EXISTS (SELECT 1 FROM keyword_weights WHERE keyword_weights.keyword = v.keyword);

-- Insert sample classification accuracy metrics only if they don't exist
INSERT INTO classification_accuracy_metrics (industry_id, total_classifications, correct_classifications, accuracy_percentage) 
SELECT * FROM (VALUES
    (1, 0, 0, 0.00),
    (2, 0, 0, 0.00),
    (3, 0, 0, 0.00),
    (4, 0, 0, 0.00),
    (5, 0, 0, 0.00),
    (6, 0, 0, 0.00),
    (7, 0, 0, 0.00),
    (8, 0, 0, 0.00),
    (9, 0, 0, 0.00),
    (10, 0, 0, 0.00)
) AS v(industry_id, total_classifications, correct_classifications, accuracy_percentage)
WHERE NOT EXISTS (SELECT 1 FROM classification_accuracy_metrics WHERE classification_accuracy_metrics.industry_id = v.industry_id);

-- =============================================================================
-- SAFE RLS AND TRIGGERS SETUP
-- =============================================================================

-- Enable RLS on all tables (safe to run multiple times)
ALTER TABLE industries ENABLE ROW LEVEL SECURITY;
ALTER TABLE industry_keywords ENABLE ROW LEVEL SECURITY;
ALTER TABLE classification_codes ENABLE ROW LEVEL SECURITY;
ALTER TABLE industry_patterns ENABLE ROW LEVEL SECURITY;
ALTER TABLE keyword_weights ENABLE ROW LEVEL SECURITY;
ALTER TABLE classification_accuracy_metrics ENABLE ROW LEVEL SECURITY;

-- Create policies for public read access (safe to run multiple times)
DROP POLICY IF EXISTS "Enable read access for all users" ON industries;
CREATE POLICY "Enable read access for all users" ON industries FOR SELECT USING (true);

DROP POLICY IF EXISTS "Enable read access for all users" ON industry_keywords;
CREATE POLICY "Enable read access for all users" ON industry_keywords FOR SELECT USING (true);

DROP POLICY IF EXISTS "Enable read access for all users" ON classification_codes;
CREATE POLICY "Enable read access for all users" ON classification_codes FOR SELECT USING (true);

DROP POLICY IF EXISTS "Enable read access for all users" ON industry_patterns;
CREATE POLICY "Enable read access for all users" ON industry_patterns FOR SELECT USING (true);

DROP POLICY IF EXISTS "Enable read access for all users" ON keyword_weights;
CREATE POLICY "Enable read access for all users" ON keyword_weights FOR SELECT USING (true);

DROP POLICY IF EXISTS "Enable read access for all users" ON classification_accuracy_metrics;
CREATE POLICY "Enable read access for all users" ON classification_accuracy_metrics FOR SELECT USING (true);

-- Function to update updated_at timestamp (safe to run multiple times)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at (safe to run multiple times)
DROP TRIGGER IF EXISTS update_industries_updated_at ON industries;
CREATE TRIGGER update_industries_updated_at BEFORE UPDATE ON industries FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_industry_keywords_updated_at ON industry_keywords;
CREATE TRIGGER update_industry_keywords_updated_at BEFORE UPDATE ON industry_keywords FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_classification_codes_updated_at ON classification_codes;
CREATE TRIGGER update_classification_codes_updated_at BEFORE UPDATE ON classification_codes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_industry_patterns_updated_at ON industry_patterns;
CREATE TRIGGER update_industry_patterns_updated_at BEFORE UPDATE ON industry_patterns FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_keyword_weights_updated_at ON keyword_weights;
CREATE TRIGGER update_keyword_weights_updated_at BEFORE UPDATE ON keyword_weights FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_classification_accuracy_metrics_updated_at ON classification_accuracy_metrics;
CREATE TRIGGER update_classification_accuracy_metrics_updated_at BEFORE UPDATE ON classification_accuracy_metrics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- VERIFICATION QUERIES
-- =============================================================================

-- Verify tables and data
SELECT 'industries' as table_name, COUNT(*) as row_count FROM industries
UNION ALL
SELECT 'industry_keywords', COUNT(*) FROM industry_keywords
UNION ALL
SELECT 'classification_codes', COUNT(*) FROM classification_codes
UNION ALL
SELECT 'industry_patterns', COUNT(*) FROM industry_patterns
UNION ALL
SELECT 'keyword_weights', COUNT(*) FROM keyword_weights
UNION ALL
SELECT 'classification_accuracy_metrics', COUNT(*) FROM classification_accuracy_metrics;

-- Show sample data to verify everything is working
SELECT 'Sample Industries:' as info;
SELECT id, name, category, confidence_threshold FROM industries LIMIT 5;

SELECT 'Sample Industry Keywords:' as info;
SELECT ik.id, i.name as industry, ik.keyword, ik.weight FROM industry_keywords ik 
JOIN industries i ON ik.industry_id = i.id LIMIT 10;
