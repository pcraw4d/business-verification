-- =====================================================
-- Railway Classifications Table Migration (Fix Version)
-- KYB Platform - Railway Server Integration
-- =====================================================
-- 
-- This script fixes the existing tables and ensures they have the correct schema
-- for the Railway server to work properly.
--
-- Author: KYB Platform Development Team
-- Date: January 22, 2025
-- Version: 1.2 (Fix)
-- =====================================================

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- FIX EXISTING TABLES
-- =====================================================

-- Fix merchants table - add missing columns if they don't exist
DO $$
BEGIN
    -- Add description column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'merchants' AND column_name = 'description') THEN
        ALTER TABLE merchants ADD COLUMN description TEXT;
    END IF;
    
    -- Add website_url column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'merchants' AND column_name = 'website_url') THEN
        ALTER TABLE merchants ADD COLUMN website_url VARCHAR(1000);
    END IF;
    
    -- Add created_at column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'merchants' AND column_name = 'created_at') THEN
        ALTER TABLE merchants ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
    END IF;
    
    -- Add updated_at column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'merchants' AND column_name = 'updated_at') THEN
        ALTER TABLE merchants ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
    END IF;
END $$;

-- Fix mock_merchants table - add missing columns if they don't exist
DO $$
BEGIN
    -- Add description column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'mock_merchants' AND column_name = 'description') THEN
        ALTER TABLE mock_merchants ADD COLUMN description TEXT;
    END IF;
    
    -- Add website_url column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'mock_merchants' AND column_name = 'website_url') THEN
        ALTER TABLE mock_merchants ADD COLUMN website_url VARCHAR(1000);
    END IF;
    
    -- Add created_at column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'mock_merchants' AND column_name = 'created_at') THEN
        ALTER TABLE mock_merchants ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
    END IF;
    
    -- Add updated_at column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'mock_merchants' AND column_name = 'updated_at') THEN
        ALTER TABLE mock_merchants ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
    END IF;
END $$;

-- =====================================================
-- CREATE MISSING TABLES
-- =====================================================

-- Create classifications table if it doesn't exist
CREATE TABLE IF NOT EXISTS classifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_id VARCHAR(255) NOT NULL,
    business_name VARCHAR(500) NOT NULL,
    description TEXT,
    website_url VARCHAR(1000),
    classification JSONB NOT NULL,
    confidence_score DECIMAL(3,2) NOT NULL CHECK (confidence_score >= 0.00 AND confidence_score <= 1.00),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(business_id)
);

-- Create merchants table if it doesn't exist (with all required columns)
CREATE TABLE IF NOT EXISTS merchants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    industry VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active',
    description TEXT,
    website_url VARCHAR(1000),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create mock_merchants table if it doesn't exist (with all required columns)
CREATE TABLE IF NOT EXISTS mock_merchants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    industry VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active',
    description TEXT,
    website_url VARCHAR(1000),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =====================================================
-- CREATE INDEXES
-- =====================================================

-- Classifications indexes
CREATE INDEX IF NOT EXISTS idx_classifications_business_id ON classifications(business_id);
CREATE INDEX IF NOT EXISTS idx_classifications_business_name ON classifications(business_name);
CREATE INDEX IF NOT EXISTS idx_classifications_created_at ON classifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_classifications_confidence_score ON classifications(confidence_score DESC);

-- Merchants indexes
CREATE INDEX IF NOT EXISTS idx_merchants_name ON merchants(name);
CREATE INDEX IF NOT EXISTS idx_merchants_industry ON merchants(industry);
CREATE INDEX IF NOT EXISTS idx_merchants_status ON merchants(status);
CREATE INDEX IF NOT EXISTS idx_merchants_created_at ON merchants(created_at DESC);

-- Mock merchants indexes
CREATE INDEX IF NOT EXISTS idx_mock_merchants_name ON mock_merchants(name);
CREATE INDEX IF NOT EXISTS idx_mock_merchants_industry ON mock_merchants(industry);
CREATE INDEX IF NOT EXISTS idx_mock_merchants_status ON mock_merchants(status);

-- =====================================================
-- SETUP ROW LEVEL SECURITY
-- =====================================================

-- Enable RLS on all tables
ALTER TABLE classifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE merchants ENABLE ROW LEVEL SECURITY;
ALTER TABLE mock_merchants ENABLE ROW LEVEL SECURITY;

-- Drop existing policies if they exist
DROP POLICY IF EXISTS "Allow public read access to classifications" ON classifications;
DROP POLICY IF EXISTS "Allow public read access to merchants" ON merchants;
DROP POLICY IF EXISTS "Allow public read access to mock_merchants" ON mock_merchants;
DROP POLICY IF EXISTS "Allow authenticated users to manage classifications" ON classifications;
DROP POLICY IF EXISTS "Allow authenticated users to manage merchants" ON merchants;
DROP POLICY IF EXISTS "Allow authenticated users to manage mock_merchants" ON mock_merchants;

-- Create policies for public read access
CREATE POLICY "Allow public read access to classifications" ON classifications
    FOR SELECT USING (true);

CREATE POLICY "Allow public read access to merchants" ON merchants
    FOR SELECT USING (true);

CREATE POLICY "Allow public read access to mock_merchants" ON mock_merchants
    FOR SELECT USING (true);

-- Create policies for authenticated write access
CREATE POLICY "Allow authenticated users to manage classifications" ON classifications
    FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Allow authenticated users to manage merchants" ON merchants
    FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Allow authenticated users to manage mock_merchants" ON mock_merchants
    FOR ALL USING (auth.role() = 'authenticated');

-- =====================================================
-- CREATE TRIGGERS
-- =====================================================

-- Create or replace the update_updated_at_column function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Drop existing triggers if they exist
DROP TRIGGER IF EXISTS update_classifications_updated_at ON classifications;
DROP TRIGGER IF EXISTS update_merchants_updated_at ON merchants;
DROP TRIGGER IF EXISTS update_mock_merchants_updated_at ON mock_merchants;

-- Create triggers for updated_at columns
CREATE TRIGGER update_classifications_updated_at 
    BEFORE UPDATE ON classifications 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_merchants_updated_at 
    BEFORE UPDATE ON merchants 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_mock_merchants_updated_at 
    BEFORE UPDATE ON mock_merchants 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- INSERT SAMPLE DATA
-- =====================================================

-- Insert sample merchants (only if they don't exist)
INSERT INTO merchants (id, name, industry, status, description) VALUES
('merch_1', 'Acme Technology Corp', 'Technology', 'active', 'Leading software development company'),
('merch_2', 'Global Retail Solutions', 'Retail', 'active', 'E-commerce platform provider'),
('merch_3', 'HealthTech Innovations', 'Healthcare', 'active', 'Medical technology solutions'),
('merch_4', 'FinanceFlow Systems', 'Finance', 'inactive', 'Financial services platform')
ON CONFLICT (id) DO NOTHING;

-- Insert sample mock merchants (only if they don't exist)
INSERT INTO mock_merchants (id, name, industry, status, description) VALUES
('mock_1', 'Mock Technology Company', 'Technology', 'active', 'Sample technology business'),
('mock_2', 'Mock Retail Store', 'Retail', 'active', 'Sample retail business'),
('mock_3', 'Mock Healthcare Provider', 'Healthcare', 'inactive', 'Sample healthcare business')
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- VERIFICATION
-- =====================================================

-- Check table structure
SELECT 
    table_name,
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns 
WHERE table_name IN ('classifications', 'merchants', 'mock_merchants')
    AND table_schema = 'public'
ORDER BY table_name, ordinal_position;

-- Check sample data
SELECT 'merchants' as table_name, COUNT(*) as record_count FROM merchants
UNION ALL
SELECT 'mock_merchants' as table_name, COUNT(*) as record_count FROM mock_merchants;

-- =====================================================
-- COMPLETION MESSAGE
-- =====================================================

DO $$
BEGIN
    RAISE NOTICE '=====================================================';
    RAISE NOTICE 'Railway Classifications Migration (Fix) Completed Successfully!';
    RAISE NOTICE '=====================================================';
    RAISE NOTICE 'Tables Fixed/Created:';
    RAISE NOTICE '  ✅ classifications (for Railway server classification storage)';
    RAISE NOTICE '  ✅ merchants (fixed schema, added missing columns)';
    RAISE NOTICE '  ✅ mock_merchants (fixed schema, added missing columns)';
    RAISE NOTICE '';
    RAISE NOTICE 'Schema Fixes Applied:';
    RAISE NOTICE '  ✅ Added missing description columns';
    RAISE NOTICE '  ✅ Added missing website_url columns';
    RAISE NOTICE '  ✅ Added missing timestamp columns';
    RAISE NOTICE '  ✅ Created all required indexes and policies';
    RAISE NOTICE '';
    RAISE NOTICE 'Railway server should now work properly!';
    RAISE NOTICE '=====================================================';
END $$;
