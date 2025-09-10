-- Create Classification System Database Schema
-- This script creates the required tables for the keyword classification system

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 1. industries Table
CREATE TABLE IF NOT EXISTS industries (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    category VARCHAR(50) NOT NULL CHECK (category IN ('traditional', 'emerging', 'hybrid')),
    parent_industry_id INTEGER REFERENCES industries(id),
    confidence_threshold DECIMAL(3,2) DEFAULT 0.80,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for industries table
CREATE INDEX IF NOT EXISTS idx_industries_name ON industries(name);
CREATE INDEX IF NOT EXISTS idx_industries_category ON industries(category);
CREATE INDEX IF NOT EXISTS idx_industries_active ON industries(is_active);
CREATE INDEX IF NOT EXISTS idx_industries_parent ON industries(parent_industry_id);

-- 2. industry_keywords Table
CREATE TABLE IF NOT EXISTS industry_keywords (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    keyword VARCHAR(255) NOT NULL,
    weight DECIMAL(5,3) DEFAULT 1.000,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(industry_id, keyword)
);

-- Indexes for industry_keywords table
CREATE INDEX IF NOT EXISTS idx_industry_keywords_industry_id ON industry_keywords(industry_id);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_keyword ON industry_keywords(keyword);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_weight ON industry_keywords(weight);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_active ON industry_keywords(is_active);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_keyword_trgm ON industry_keywords USING gin(keyword gin_trgm_ops);

-- 3. classification_codes Table
CREATE TABLE IF NOT EXISTS classification_codes (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    code_type VARCHAR(10) NOT NULL CHECK (code_type IN ('NAICS', 'MCC', 'SIC')),
    code VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(code_type, code)
);

-- Indexes for classification_codes table
CREATE INDEX IF NOT EXISTS idx_classification_codes_industry_id ON classification_codes(industry_id);
CREATE INDEX IF NOT EXISTS idx_classification_codes_type ON classification_codes(code_type);
CREATE INDEX IF NOT EXISTS idx_classification_codes_code ON classification_codes(code);
CREATE INDEX IF NOT EXISTS idx_classification_codes_active ON classification_codes(is_active);
CREATE INDEX IF NOT EXISTS idx_classification_codes_description_trgm ON classification_codes USING gin(description gin_trgm_ops);

-- 4. industry_patterns Table
CREATE TABLE IF NOT EXISTS industry_patterns (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    pattern TEXT NOT NULL,
    pattern_type VARCHAR(50) NOT NULL CHECK (pattern_type IN ('phrase', 'regex', 'keyword_combination')),
    confidence_score DECIMAL(3,2) DEFAULT 0.80,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for industry_patterns table
CREATE INDEX IF NOT EXISTS idx_industry_patterns_industry_id ON industry_patterns(industry_id);
CREATE INDEX IF NOT EXISTS idx_industry_patterns_type ON industry_patterns(pattern_type);
CREATE INDEX IF NOT EXISTS idx_industry_patterns_confidence ON industry_patterns(confidence_score);
CREATE INDEX IF NOT EXISTS idx_industry_patterns_active ON industry_patterns(is_active);
CREATE INDEX IF NOT EXISTS idx_industry_patterns_pattern_trgm ON industry_patterns USING gin(pattern gin_trgm_ops);

-- 5. keyword_weights Table
CREATE TABLE IF NOT EXISTS keyword_weights (
    id SERIAL PRIMARY KEY,
    keyword VARCHAR(255) NOT NULL,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    base_weight DECIMAL(5,3) DEFAULT 1.000,
    context_multiplier DECIMAL(5,3) DEFAULT 1.000,
    usage_count INTEGER DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(keyword, industry_id)
);

-- Indexes for keyword_weights table
CREATE INDEX IF NOT EXISTS idx_keyword_weights_keyword ON keyword_weights(keyword);
CREATE INDEX IF NOT EXISTS idx_keyword_weights_industry_id ON keyword_weights(industry_id);
CREATE INDEX IF NOT EXISTS idx_keyword_weights_base_weight ON keyword_weights(base_weight);
CREATE INDEX IF NOT EXISTS idx_keyword_weights_usage_count ON keyword_weights(usage_count);
CREATE INDEX IF NOT EXISTS idx_keyword_weights_keyword_trgm ON keyword_weights USING gin(keyword gin_trgm_ops);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at columns
CREATE TRIGGER update_industries_updated_at BEFORE UPDATE ON industries FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_industry_keywords_updated_at BEFORE UPDATE ON industry_keywords FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_classification_codes_updated_at BEFORE UPDATE ON classification_codes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_industry_patterns_updated_at BEFORE UPDATE ON industry_patterns FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_keyword_weights_updated_at BEFORE UPDATE ON keyword_weights FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert sample data for testing
INSERT INTO industries (name, description, category, confidence_threshold) VALUES
('Technology', 'Technology and software companies', 'traditional', 0.85),
('Financial Services', 'Banking, insurance, and financial institutions', 'traditional', 0.90),
('Healthcare', 'Medical and healthcare services', 'traditional', 0.88),
('Manufacturing', 'Industrial manufacturing and production', 'traditional', 0.82),
('Retail', 'Retail and consumer goods', 'traditional', 0.80),
('General Business', 'General business services', 'traditional', 0.50)
ON CONFLICT (name) DO NOTHING;

-- Insert sample keywords
INSERT INTO industry_keywords (industry_id, keyword, weight) VALUES
(1, 'software', 1.0),
(1, 'technology', 1.0),
(1, 'platform', 0.9),
(1, 'digital', 0.8),
(1, 'tech', 0.7),
(2, 'bank', 1.0),
(2, 'finance', 1.0),
(2, 'credit', 0.9),
(2, 'investment', 0.8),
(2, 'insurance', 0.8),
(3, 'healthcare', 1.0),
(3, 'medical', 1.0),
(3, 'hospital', 0.9),
(3, 'clinic', 0.8),
(3, 'pharmacy', 0.7),
(4, 'manufacturing', 1.0),
(4, 'factory', 0.9),
(4, 'production', 0.8),
(4, 'industrial', 0.7),
(5, 'retail', 1.0),
(5, 'store', 0.9),
(5, 'shop', 0.8),
(5, 'merchandise', 0.7)
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- Insert sample classification codes
INSERT INTO classification_codes (industry_id, code_type, code, description) VALUES
(1, 'MCC', '5734', 'Computer Software Stores'),
(1, 'SIC', '7372', 'Prepackaged Software'),
(1, 'NAICS', '541511', 'Custom Computer Programming Services'),
(2, 'MCC', '6011', 'Automated Teller Machine Services'),
(2, 'MCC', '6012', 'Financial Institutions - Manual Cash Disbursements'),
(2, 'SIC', '6021', 'National Commercial Banks'),
(2, 'SIC', '6022', 'State Commercial Banks'),
(2, 'NAICS', '522110', 'Commercial Banking'),
(2, 'NAICS', '522120', 'Savings Institutions'),
(3, 'MCC', '8062', 'Hospitals'),
(3, 'SIC', '8062', 'General Medical and Surgical Hospitals'),
(3, 'NAICS', '622110', 'General Medical and Surgical Hospitals'),
(4, 'MCC', '5085', 'Industrial Supplies'),
(4, 'SIC', '3089', 'Plastics Products, Not Elsewhere Classified'),
(4, 'NAICS', '326199', 'All Other Miscellaneous Plastics Product Manufacturing'),
(5, 'MCC', '5310', 'Department Stores'),
(5, 'SIC', '5311', 'Department Stores'),
(5, 'NAICS', '452111', 'Department Stores (except Discount Department Stores)')
ON CONFLICT (code_type, code) DO NOTHING;

-- Insert sample patterns
INSERT INTO industry_patterns (industry_id, pattern, pattern_type, confidence_score) VALUES
(1, 'software development', 'phrase', 0.95),
(1, 'technology company', 'phrase', 0.90),
(1, 'digital platform', 'phrase', 0.85),
(2, 'financial institution', 'phrase', 0.95),
(2, 'banking services', 'phrase', 0.90),
(2, 'credit union', 'phrase', 0.85),
(3, 'healthcare provider', 'phrase', 0.95),
(3, 'medical services', 'phrase', 0.90),
(3, 'hospital system', 'phrase', 0.85),
(4, 'manufacturing company', 'phrase', 0.95),
(4, 'industrial production', 'phrase', 0.90),
(4, 'factory operations', 'phrase', 0.85),
(5, 'retail store', 'phrase', 0.95),
(5, 'shopping center', 'phrase', 0.90),
(5, 'merchandise sales', 'phrase', 0.85)
ON CONFLICT DO NOTHING;

-- Insert sample keyword weights
INSERT INTO keyword_weights (keyword, industry_id, base_weight, context_multiplier, usage_count) VALUES
('software', 1, 1.0, 1.0, 0),
('technology', 1, 1.0, 1.0, 0),
('platform', 1, 0.9, 1.0, 0),
('digital', 1, 0.8, 1.0, 0),
('tech', 1, 0.7, 1.0, 0),
('bank', 2, 1.0, 1.0, 0),
('finance', 2, 1.0, 1.0, 0),
('credit', 2, 0.9, 1.0, 0),
('investment', 2, 0.8, 1.0, 0),
('insurance', 2, 0.8, 1.0, 0),
('healthcare', 3, 1.0, 1.0, 0),
('medical', 3, 1.0, 1.0, 0),
('hospital', 3, 0.9, 1.0, 0),
('clinic', 3, 0.8, 1.0, 0),
('pharmacy', 3, 0.7, 1.0, 0),
('manufacturing', 4, 1.0, 1.0, 0),
('factory', 4, 0.9, 1.0, 0),
('production', 4, 0.8, 1.0, 0),
('industrial', 4, 0.7, 1.0, 0),
('retail', 5, 1.0, 1.0, 0),
('store', 5, 0.9, 1.0, 0),
('shop', 5, 0.8, 1.0, 0),
('merchandise', 5, 0.7, 1.0, 0)
ON CONFLICT (keyword, industry_id) DO NOTHING;

-- Create views for easier querying
CREATE OR REPLACE VIEW industry_classification_summary AS
SELECT 
    i.id,
    i.name as industry_name,
    i.category,
    i.confidence_threshold,
    COUNT(DISTINCT ik.id) as keyword_count,
    COUNT(DISTINCT cc.id) as classification_code_count,
    COUNT(DISTINCT ip.id) as pattern_count
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
LEFT JOIN classification_codes cc ON i.id = cc.industry_id AND cc.is_active = true
LEFT JOIN industry_patterns ip ON i.id = ip.industry_id AND ip.is_active = true
WHERE i.is_active = true
GROUP BY i.id, i.name, i.category, i.confidence_threshold;

-- Grant permissions (adjust as needed for your setup)
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO your_app_user;
-- GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO your_app_user;

COMMENT ON TABLE industries IS 'Business industries with classification metadata';
COMMENT ON TABLE industry_keywords IS 'Keywords associated with industries for classification';
COMMENT ON TABLE classification_codes IS 'Industry classification codes (NAICS, MCC, SIC)';
COMMENT ON TABLE industry_patterns IS 'Phrase patterns for industry detection';
COMMENT ON TABLE keyword_weights IS 'Dynamic keyword weighting and scoring';
