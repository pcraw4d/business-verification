-- KYB Platform - Classification System Database Migration
-- Run this script in the Supabase SQL Editor to create classification tables

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- =============================================================================
-- INDUSTRIES TABLE
-- =============================================================================
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

-- =============================================================================
-- INDUSTRY KEYWORDS TABLE
-- =============================================================================
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

-- =============================================================================
-- CLASSIFICATION CODES TABLE
-- =============================================================================
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

-- =============================================================================
-- INDUSTRY PATTERNS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS industry_patterns (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    pattern VARCHAR(500) NOT NULL,
    pattern_type VARCHAR(50) NOT NULL DEFAULT 'phrase',
    confidence_score DECIMAL(3,2) DEFAULT 0.50,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- KEYWORD WEIGHTS TABLE (for dynamic weighting)
-- =============================================================================
CREATE TABLE IF NOT EXISTS keyword_weights (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    keyword VARCHAR(255) NOT NULL,
    base_weight DECIMAL(5,4) DEFAULT 1.0000,
    usage_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(industry_id, keyword)
);

-- =============================================================================
-- CLASSIFICATION ACCURACY METRICS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS classification_accuracy_metrics (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    request_id VARCHAR(255),
    business_name VARCHAR(500),
    business_description TEXT,
    website_url VARCHAR(1000),
    predicted_industry VARCHAR(255),
    predicted_confidence DECIMAL(3,2),
    actual_industry VARCHAR(255),
    actual_confidence DECIMAL(3,2),
    accuracy_score DECIMAL(3,2),
    response_time_ms DECIMAL(10,2),
    processing_time_ms DECIMAL(10,2),
    classification_method VARCHAR(100),
    keywords_used TEXT[],
    confidence_threshold DECIMAL(3,2) DEFAULT 0.50,
    is_correct BOOLEAN,
    error_message TEXT,
    user_feedback TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- INDEXES FOR PERFORMANCE
-- =============================================================================

-- Industries indexes
CREATE INDEX IF NOT EXISTS idx_industries_name ON industries(name);
CREATE INDEX IF NOT EXISTS idx_industries_active ON industries(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_industries_category ON industries(category);

-- Industry keywords indexes
CREATE INDEX IF NOT EXISTS idx_industry_keywords_industry ON industry_keywords(industry_id);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_keyword ON industry_keywords(keyword);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_active ON industry_keywords(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_industry_keywords_weight ON industry_keywords(weight DESC);

-- Classification codes indexes
CREATE INDEX IF NOT EXISTS idx_classification_codes_industry ON classification_codes(industry_id);
CREATE INDEX IF NOT EXISTS idx_classification_codes_type ON classification_codes(code_type);
CREATE INDEX IF NOT EXISTS idx_classification_codes_code ON classification_codes(code);
CREATE INDEX IF NOT EXISTS idx_classification_codes_active ON classification_codes(is_active) WHERE is_active = true;

-- Industry patterns indexes
CREATE INDEX IF NOT EXISTS idx_industry_patterns_industry ON industry_patterns(industry_id);
CREATE INDEX IF NOT EXISTS idx_industry_patterns_pattern ON industry_patterns(pattern);
CREATE INDEX IF NOT EXISTS idx_industry_patterns_active ON industry_patterns(is_active) WHERE is_active = true;

-- Keyword weights indexes
CREATE INDEX IF NOT EXISTS idx_keyword_weights_industry ON keyword_weights(industry_id);
CREATE INDEX IF NOT EXISTS idx_keyword_weights_keyword ON keyword_weights(keyword);
CREATE INDEX IF NOT EXISTS idx_keyword_weights_active ON keyword_weights(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_keyword_weights_usage ON keyword_weights(usage_count DESC);

-- Classification accuracy metrics indexes
CREATE INDEX IF NOT EXISTS idx_classification_accuracy_timestamp ON classification_accuracy_metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_classification_accuracy_request_id ON classification_accuracy_metrics(request_id);
CREATE INDEX IF NOT EXISTS idx_classification_accuracy_industry ON classification_accuracy_metrics(predicted_industry);

-- =============================================================================
-- SAMPLE DATA INSERTION
-- =============================================================================

-- Insert sample industries
INSERT INTO industries (name, description, category, confidence_threshold) VALUES
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
ON CONFLICT (name) DO NOTHING;

-- Insert sample keywords for Technology industry
INSERT INTO industry_keywords (industry_id, keyword, weight) 
SELECT i.id, k.keyword, k.weight
FROM industries i, (VALUES
    ('software', 1.0),
    ('technology', 1.0),
    ('development', 0.9),
    ('programming', 0.9),
    ('computer', 0.8),
    ('digital', 0.8),
    ('tech', 0.7),
    ('app', 0.7),
    ('platform', 0.6),
    ('system', 0.6)
) AS k(keyword, weight)
WHERE i.name = 'Technology'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- Insert sample keywords for Retail industry
INSERT INTO industry_keywords (industry_id, keyword, weight)
SELECT i.id, k.keyword, k.weight
FROM industries i, (VALUES
    ('retail', 1.0),
    ('store', 0.9),
    ('shop', 0.9),
    ('commerce', 0.8),
    ('sales', 0.8),
    ('merchandise', 0.7),
    ('products', 0.7),
    ('ecommerce', 0.6),
    ('online', 0.6),
    ('marketplace', 0.5)
) AS k(keyword, weight)
WHERE i.name = 'Retail'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- Insert sample classification codes for Technology
INSERT INTO classification_codes (industry_id, code_type, code, description)
SELECT i.id, c.code_type, c.code, c.description
FROM industries i, (VALUES
    ('NAICS', '541511', 'Custom Computer Programming Services'),
    ('NAICS', '541512', 'Computer Systems Design Services'),
    ('NAICS', '541513', 'Computer Facilities Management Services'),
    ('SIC', '7372', 'Computer Programming Services'),
    ('SIC', '7373', 'Computer Integrated Systems Design'),
    ('MCC', '7372', 'Computer Programming Services'),
    ('MCC', '7373', 'Computer Integrated Systems Design')
) AS c(code_type, code, description)
WHERE i.name = 'Technology'
ON CONFLICT (industry_id, code_type, code) DO NOTHING;

-- Insert sample classification codes for Retail
INSERT INTO classification_codes (industry_id, code_type, code, description)
SELECT i.id, c.code_type, c.code, c.description
FROM industries i, (VALUES
    ('NAICS', '448110', 'Men''s Clothing Stores'),
    ('NAICS', '448120', 'Women''s Clothing Stores'),
    ('NAICS', '448130', 'Children''s and Infants'' Clothing Stores'),
    ('SIC', '5611', 'Men''s and Boys'' Clothing and Accessory Stores'),
    ('SIC', '5621', 'Women''s Clothing Stores'),
    ('MCC', '5651', 'Family Clothing Stores'),
    ('MCC', '5655', 'Sports and Riding Apparel Stores')
) AS c(code_type, code, description)
WHERE i.name = 'Retail'
ON CONFLICT (industry_id, code_type, code) DO NOTHING;

-- =============================================================================
-- ROW LEVEL SECURITY (RLS) POLICIES
-- =============================================================================

-- Enable RLS on all tables
ALTER TABLE industries ENABLE ROW LEVEL SECURITY;
ALTER TABLE industry_keywords ENABLE ROW LEVEL SECURITY;
ALTER TABLE classification_codes ENABLE ROW LEVEL SECURITY;
ALTER TABLE industry_patterns ENABLE ROW LEVEL SECURITY;
ALTER TABLE keyword_weights ENABLE ROW LEVEL SECURITY;
ALTER TABLE classification_accuracy_metrics ENABLE ROW LEVEL SECURITY;

-- Create policies for public read access (for classification)
CREATE POLICY "Allow public read access to industries" ON industries
    FOR SELECT USING (is_active = true);

CREATE POLICY "Allow public read access to industry keywords" ON industry_keywords
    FOR SELECT USING (is_active = true);

CREATE POLICY "Allow public read access to classification codes" ON classification_codes
    FOR SELECT USING (is_active = true);

CREATE POLICY "Allow public read access to industry patterns" ON industry_patterns
    FOR SELECT USING (is_active = true);

CREATE POLICY "Allow public read access to keyword weights" ON keyword_weights
    FOR SELECT USING (is_active = true);

-- Allow service role to insert accuracy metrics
CREATE POLICY "Allow service role to insert accuracy metrics" ON classification_accuracy_metrics
    FOR INSERT WITH CHECK (true);

-- =============================================================================
-- FUNCTIONS AND TRIGGERS
-- =============================================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_industries_updated_at BEFORE UPDATE ON industries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_industry_keywords_updated_at BEFORE UPDATE ON industry_keywords
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_classification_codes_updated_at BEFORE UPDATE ON classification_codes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_industry_patterns_updated_at BEFORE UPDATE ON industry_patterns
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================

DO $$
BEGIN
    RAISE NOTICE 'KYB Platform Classification System database migration completed successfully!';
    RAISE NOTICE 'Tables created: industries, industry_keywords, classification_codes, industry_patterns, keyword_weights, classification_accuracy_metrics';
    RAISE NOTICE 'Sample data inserted for Technology and Retail industries';
    RAISE NOTICE 'Indexes and RLS policies configured';
END $$;
