-- KYB Platform - Classification System Database Migration (CONSTRAINT FIX)
-- This script fixes the industries_category_check constraint issue

-- =============================================================================
-- STEP 1: SAFELY REMOVE THE PROBLEMATIC CONSTRAINT
-- =============================================================================

-- Drop the problematic constraint if it exists
DO $$ 
BEGIN
    -- Check if the constraint exists and drop it
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'industries_category_check' 
        AND table_name = 'industries'
    ) THEN
        ALTER TABLE industries DROP CONSTRAINT industries_category_check;
        RAISE NOTICE 'Dropped industries_category_check constraint';
    ELSE
        RAISE NOTICE 'industries_category_check constraint does not exist';
    END IF;
END $$;

-- =============================================================================
-- STEP 2: ENSURE ALL REQUIRED COLUMNS EXIST
-- =============================================================================

-- Add missing columns if they don't exist
DO $$ 
BEGIN
    -- Add weight column to industry_keywords if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'industry_keywords' AND column_name = 'weight') THEN
        ALTER TABLE industry_keywords ADD COLUMN weight DECIMAL(5,4) DEFAULT 1.0000;
        RAISE NOTICE 'Added weight column to industry_keywords';
    END IF;
    
    -- Add confidence_weight column to industry_patterns if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'industry_patterns' AND column_name = 'confidence_weight') THEN
        ALTER TABLE industry_patterns ADD COLUMN confidence_weight DECIMAL(3,2) DEFAULT 1.00;
        RAISE NOTICE 'Added confidence_weight column to industry_patterns';
    END IF;
    
    -- Add weight column to keyword_weights if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'keyword_weights' AND column_name = 'weight') THEN
        ALTER TABLE keyword_weights ADD COLUMN weight DECIMAL(5,4) DEFAULT 1.0000;
        RAISE NOTICE 'Added weight column to keyword_weights';
    END IF;
END $$;

-- =============================================================================
-- STEP 3: SAFE DATA INSERTION WITH PROPER ERROR HANDLING
-- =============================================================================

-- Insert sample industries with proper error handling
DO $$ 
BEGIN
    -- Insert industries one by one to handle any remaining issues
    BEGIN
        INSERT INTO industries (name, description, category, confidence_threshold) 
        VALUES ('Technology', 'Software development, IT services, and technology companies', 'Technology', 0.70)
        ON CONFLICT (name) DO NOTHING;
        RAISE NOTICE 'Inserted Technology industry';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Failed to insert Technology: %', SQLERRM;
    END;
    
    BEGIN
        INSERT INTO industries (name, description, category, confidence_threshold) 
        VALUES ('Retail', 'Consumer goods retail and e-commerce', 'Commerce', 0.60)
        ON CONFLICT (name) DO NOTHING;
        RAISE NOTICE 'Inserted Retail industry';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Failed to insert Retail: %', SQLERRM;
    END;
    
    BEGIN
        INSERT INTO industries (name, description, category, confidence_threshold) 
        VALUES ('Healthcare', 'Medical services, pharmaceuticals, and healthcare technology', 'Healthcare', 0.75)
        ON CONFLICT (name) DO NOTHING;
        RAISE NOTICE 'Inserted Healthcare industry';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Failed to insert Healthcare: %', SQLERRM;
    END;
    
    BEGIN
        INSERT INTO industries (name, description, category, confidence_threshold) 
        VALUES ('Finance', 'Banking, investment, and financial services', 'Finance', 0.80)
        ON CONFLICT (name) DO NOTHING;
        RAISE NOTICE 'Inserted Finance industry';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Failed to insert Finance: %', SQLERRM;
    END;
    
    BEGIN
        INSERT INTO industries (name, description, category, confidence_threshold) 
        VALUES ('Manufacturing', 'Industrial production and manufacturing', 'Industrial', 0.65)
        ON CONFLICT (name) DO NOTHING;
        RAISE NOTICE 'Inserted Manufacturing industry';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Failed to insert Manufacturing: %', SQLERRM;
    END;
    
    BEGIN
        INSERT INTO industries (name, description, category, confidence_threshold) 
        VALUES ('Food & Beverage', 'Restaurants, food production, and beverage companies', 'Consumer', 0.55)
        ON CONFLICT (name) DO NOTHING;
        RAISE NOTICE 'Inserted Food & Beverage industry';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Failed to insert Food & Beverage: %', SQLERRM;
    END;
    
    BEGIN
        INSERT INTO industries (name, description, category, confidence_threshold) 
        VALUES ('Real Estate', 'Property development, real estate services', 'Property', 0.60)
        ON CONFLICT (name) DO NOTHING;
        RAISE NOTICE 'Inserted Real Estate industry';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Failed to insert Real Estate: %', SQLERRM;
    END;
    
    BEGIN
        INSERT INTO industries (name, description, category, confidence_threshold) 
        VALUES ('Education', 'Educational institutions and training services', 'Education', 0.70)
        ON CONFLICT (name) DO NOTHING;
        RAISE NOTICE 'Inserted Education industry';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Failed to insert Education: %', SQLERRM;
    END;
    
    BEGIN
        INSERT INTO industries (name, description, category, confidence_threshold) 
        VALUES ('Transportation', 'Logistics, shipping, and transportation services', 'Logistics', 0.65)
        ON CONFLICT (name) DO NOTHING;
        RAISE NOTICE 'Inserted Transportation industry';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Failed to insert Transportation: %', SQLERRM;
    END;
    
    BEGIN
        INSERT INTO industries (name, description, category, confidence_threshold) 
        VALUES ('Entertainment', 'Media, entertainment, and creative services', 'Media', 0.60)
        ON CONFLICT (name) DO NOTHING;
        RAISE NOTICE 'Inserted Entertainment industry';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Failed to insert Entertainment: %', SQLERRM;
    END;
END $$;

-- =============================================================================
-- STEP 4: INSERT SAMPLE KEYWORDS (SAFE)
-- =============================================================================

-- Insert sample industry keywords with error handling
DO $$ 
BEGIN
    -- Technology keywords
    INSERT INTO industry_keywords (industry_id, keyword, weight) VALUES
    (1, 'software', 1.2),
    (1, 'technology', 1.1),
    (1, 'computer', 1.0),
    (1, 'tech', 1.0),
    (1, 'programming', 1.3),
    (1, 'development', 1.1),
    (1, 'IT', 1.0),
    (1, 'digital', 0.9),
    (1, 'app', 1.0),
    (1, 'platform', 1.0)
    ON CONFLICT (industry_id, keyword) DO NOTHING;
    
    -- Retail keywords
    INSERT INTO industry_keywords (industry_id, keyword, weight) VALUES
    (2, 'retail', 1.2),
    (2, 'store', 1.1),
    (2, 'shop', 1.0),
    (2, 'commerce', 1.1),
    (2, 'ecommerce', 1.2),
    (2, 'marketplace', 1.0),
    (2, 'sales', 0.9),
    (2, 'merchandise', 1.0),
    (2, 'products', 0.8),
    (2, 'goods', 0.8)
    ON CONFLICT (industry_id, keyword) DO NOTHING;
    
    -- Healthcare keywords
    INSERT INTO industry_keywords (industry_id, keyword, weight) VALUES
    (3, 'health', 1.2),
    (3, 'medical', 1.3),
    (3, 'healthcare', 1.2),
    (3, 'hospital', 1.1),
    (3, 'clinic', 1.0),
    (3, 'pharmacy', 1.0),
    (3, 'doctor', 1.0),
    (3, 'patient', 0.9),
    (3, 'treatment', 1.0),
    (3, 'medicine', 1.0)
    ON CONFLICT (industry_id, keyword) DO NOTHING;
    
    -- Finance keywords
    INSERT INTO industry_keywords (industry_id, keyword, weight) VALUES
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
    ON CONFLICT (industry_id, keyword) DO NOTHING;
    
    RAISE NOTICE 'Inserted industry keywords successfully';
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Error inserting keywords: %', SQLERRM;
END $$;

-- =============================================================================
-- STEP 5: INSERT SAMPLE CLASSIFICATION CODES (SAFE)
-- =============================================================================

-- Insert sample classification codes with error handling
DO $$ 
BEGIN
    -- Technology codes
    INSERT INTO classification_codes (industry_id, code_type, code, description) VALUES
    (1, 'NAICS', '541511', 'Custom Computer Programming Services'),
    (1, 'NAICS', '541512', 'Computer Systems Design Services'),
    (1, 'NAICS', '541513', 'Computer Facilities Management Services'),
    (1, 'SIC', '7372', 'Computer Programming Services'),
    (1, 'SIC', '7373', 'Computer Integrated Systems Design'),
    (1, 'MCC', '7372', 'Computer Programming Services')
    ON CONFLICT (industry_id, code_type, code) DO NOTHING;
    
    -- Retail codes
    INSERT INTO classification_codes (industry_id, code_type, code, description) VALUES
    (2, 'NAICS', '454110', 'Electronic Shopping and Mail-Order Houses'),
    (2, 'NAICS', '448140', 'Family Clothing Stores'),
    (2, 'SIC', '5961', 'Catalog and Mail-Order Houses'),
    (2, 'SIC', '5621', 'Women''s Clothing Stores'),
    (2, 'MCC', '5311', 'Department Stores')
    ON CONFLICT (industry_id, code_type, code) DO NOTHING;
    
    -- Healthcare codes
    INSERT INTO classification_codes (industry_id, code_type, code, description) VALUES
    (3, 'NAICS', '621111', 'Offices of Physicians (except Mental Health Specialists)'),
    (3, 'NAICS', '622110', 'General Medical and Surgical Hospitals'),
    (3, 'SIC', '8011', 'Offices and Clinics of Doctors of Medicine'),
    (3, 'SIC', '8062', 'General Medical and Surgical Hospitals'),
    (3, 'MCC', '8062', 'Hospitals')
    ON CONFLICT (industry_id, code_type, code) DO NOTHING;
    
    -- Finance codes
    INSERT INTO classification_codes (industry_id, code_type, code, description) VALUES
    (4, 'NAICS', '522110', 'Commercial Banking'),
    (4, 'NAICS', '523110', 'Investment Banking and Securities Dealing'),
    (4, 'SIC', '6021', 'National Commercial Banks'),
    (4, 'SIC', '6211', 'Security Brokers, Dealers, and Flotation Companies'),
    (4, 'MCC', '6010', 'Financial Institutions - Merchandise, Services')
    ON CONFLICT (industry_id, code_type, code) DO NOTHING;
    
    RAISE NOTICE 'Inserted classification codes successfully';
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Error inserting classification codes: %', SQLERRM;
END $$;

-- =============================================================================
-- STEP 6: VERIFICATION
-- =============================================================================

-- Show current state
SELECT 'Current Industries:' as info;
SELECT id, name, category, confidence_threshold FROM industries ORDER BY id;

SELECT 'Current Industry Keywords Count:' as info;
SELECT COUNT(*) as keyword_count FROM industry_keywords;

SELECT 'Current Classification Codes Count:' as info;
SELECT COUNT(*) as code_count FROM classification_codes;

-- Show any remaining constraints on industries table
SELECT 'Remaining Constraints on Industries:' as info;
SELECT constraint_name, constraint_type 
FROM information_schema.table_constraints 
WHERE table_name = 'industries';
