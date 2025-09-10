-- Create tables for keyword classification system

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

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
    keyword VARCHAR(100) NOT NULL,
    weight DECIMAL(3,2) DEFAULT 1.00 CHECK (weight >= 0.00 AND weight <= 1.00),
    context VARCHAR(50) CHECK (context IN ('business', 'technical', 'general')),
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(industry_id, keyword)
);

-- Indexes for industry_keywords table
CREATE INDEX IF NOT EXISTS idx_industry_keywords_industry ON industry_keywords(industry_id);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_keyword ON industry_keywords(keyword);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_weight ON industry_keywords(weight);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_primary ON industry_keywords(is_primary);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_context ON industry_keywords(context);

-- 3. classification_codes Table
CREATE TABLE IF NOT EXISTS classification_codes (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    code_type VARCHAR(10) NOT NULL CHECK (code_type IN ('NAICS', 'MCC', 'SIC')),
    code VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    confidence DECIMAL(3,2) DEFAULT 0.80 CHECK (confidence >= 0.00 AND confidence <= 1.00),
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(code_type, code)
);

-- Indexes for classification_codes table
CREATE INDEX IF NOT EXISTS idx_classification_codes_industry ON classification_codes(industry_id);
CREATE INDEX IF NOT EXISTS idx_classification_codes_type ON classification_codes(code_type);
CREATE INDEX IF NOT EXISTS idx_classification_codes_code ON classification_codes(code);
CREATE INDEX IF NOT EXISTS idx_classification_codes_primary ON classification_codes(is_primary);

-- Insert sample data
INSERT INTO industries (name, description, category, confidence_threshold) VALUES
('Technology', 'Technology and software companies', 'traditional', 0.80),
('Healthcare', 'Healthcare and medical services', 'traditional', 0.85),
('Financial Services', 'Banking, finance, and investment services', 'traditional', 0.90),
('Retail', 'Retail and consumer goods', 'traditional', 0.75),
('Manufacturing', 'Manufacturing and industrial production', 'traditional', 0.80)
ON CONFLICT (name) DO NOTHING;

-- Insert sample keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, context, is_primary) VALUES
-- Technology keywords
(1, 'software', 0.9, 'technical', true),
(1, 'technology', 0.8, 'general', true),
(1, 'development', 0.7, 'technical', false),
(1, 'platform', 0.6, 'technical', false),
(1, 'digital', 0.5, 'general', false),
-- Healthcare keywords
(2, 'medical', 0.9, 'business', true),
(2, 'healthcare', 0.8, 'business', true),
(2, 'patient', 0.6, 'business', false),
(2, 'clinic', 0.7, 'business', false),
-- Financial Services keywords
(3, 'bank', 0.9, 'business', true),
(3, 'finance', 0.8, 'business', true),
(3, 'credit', 0.7, 'business', false),
(3, 'investment', 0.6, 'business', false)
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- Insert sample classification codes
INSERT INTO classification_codes (industry_id, code_type, code, description, confidence, is_primary) VALUES
-- Technology codes
(1, 'NAICS', '541511', 'Custom Computer Programming Services', 0.9, true),
(1, 'NAICS', '541512', 'Computer Systems Design Services', 0.85, false),
(1, 'MCC', '5734', 'Computer Software Stores', 0.8, true),
(1, 'SIC', '7372', 'Prepackaged Software', 0.85, true),
-- Healthcare codes
(2, 'NAICS', '621111', 'Offices of Physicians', 0.9, true),
(2, 'NAICS', '621112', 'Offices of Physicians, Mental Health Specialists', 0.85, false),
(2, 'MCC', '8099', 'Health Practitioners, Not Elsewhere Classified', 0.8, true),
(2, 'SIC', '8011', 'Offices and Clinics of Doctors of Medicine', 0.85, true),
-- Financial Services codes
(3, 'NAICS', '522110', 'Commercial Banking', 0.9, true),
(3, 'NAICS', '522120', 'Savings Institutions', 0.85, false),
(3, 'MCC', '6011', 'Automated Teller Machine Services', 0.8, true),
(3, 'SIC', '6021', 'National Commercial Banks', 0.85, true)
ON CONFLICT (code_type, code) DO NOTHING;
