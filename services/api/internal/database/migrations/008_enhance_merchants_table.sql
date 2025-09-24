-- Migration: Enhance Merchants Table for Business Consolidation
-- Subtask 2.2.2: Enhance Merchants Table
-- Date: January 19, 2025
-- Purpose: Add missing fields from businesses table to merchants table

-- Start transaction for atomic migration
BEGIN;

-- Add missing fields to merchants table
-- 1. Add metadata field for extensibility (from businesses table)
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}';

-- 2. Add website_url field (separate from contact_website for business website)
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS website_url TEXT;

-- 3. Add description field for business description
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS description TEXT;

-- 4. Enhance field lengths to match businesses table specifications
-- Increase name field length from VARCHAR(255) to VARCHAR(500)
ALTER TABLE merchants ALTER COLUMN name TYPE VARCHAR(500);

-- Increase industry field length from VARCHAR(100) to VARCHAR(255)
ALTER TABLE merchants ALTER COLUMN industry TYPE VARCHAR(255);

-- Increase industry_code field length from VARCHAR(20) to VARCHAR(50)
ALTER TABLE merchants ALTER COLUMN industry_code TYPE VARCHAR(50);

-- 5. Add user_id field for backward compatibility with businesses table
-- This will be used during migration and can be removed later
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;

-- 6. Enhance constraints to match businesses table requirements
-- Make registration_number nullable initially (will be populated during migration)
ALTER TABLE merchants ALTER COLUMN registration_number DROP NOT NULL;

-- Make legal_name nullable initially (will be populated during migration)
ALTER TABLE merchants ALTER COLUMN legal_name DROP NOT NULL;

-- Make portfolio_type_id nullable initially (will be populated during migration)
ALTER TABLE merchants ALTER COLUMN portfolio_type_id DROP NOT NULL;

-- Make risk_level_id nullable initially (will be populated during migration)
ALTER TABLE merchants ALTER COLUMN risk_level_id DROP NOT NULL;

-- Make created_by nullable initially (will be populated during migration)
ALTER TABLE merchants ALTER COLUMN created_by DROP NOT NULL;

-- 7. Add indexes for new fields for performance
CREATE INDEX IF NOT EXISTS idx_merchants_metadata ON merchants USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_merchants_website_url ON merchants (website_url) WHERE website_url IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_merchants_description ON merchants (description) WHERE description IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_merchants_user_id ON merchants (user_id) WHERE user_id IS NOT NULL;

-- 8. Add comments for documentation
COMMENT ON COLUMN merchants.metadata IS 'JSONB field for storing additional business metadata and extensibility';
COMMENT ON COLUMN merchants.website_url IS 'Primary business website URL (separate from contact website)';
COMMENT ON COLUMN merchants.description IS 'Business description and summary';
COMMENT ON COLUMN merchants.user_id IS 'User ID for backward compatibility during migration from businesses table';

-- 9. Create function to handle data migration from businesses to merchants
CREATE OR REPLACE FUNCTION migrate_businesses_to_merchants()
RETURNS INTEGER AS $$
DECLARE
    migrated_count INTEGER := 0;
    business_record RECORD;
    default_portfolio_type_id UUID;
    default_risk_level_id UUID;
BEGIN
    -- Get default portfolio type ID (prospective for migrated businesses)
    SELECT id INTO default_portfolio_type_id 
    FROM portfolio_types 
    WHERE name = 'prospective' 
    LIMIT 1;
    
    -- If no prospective type exists, get the first available type
    IF default_portfolio_type_id IS NULL THEN
        SELECT id INTO default_portfolio_type_id 
        FROM portfolio_types 
        LIMIT 1;
    END IF;
    
    -- Get default risk level ID (medium for migrated businesses)
    SELECT id INTO default_risk_level_id 
    FROM risk_levels 
    WHERE name = 'medium' 
    LIMIT 1;
    
    -- If no medium level exists, get the first available level
    IF default_risk_level_id IS NULL THEN
        SELECT id INTO default_risk_level_id 
        FROM risk_levels 
        LIMIT 1;
    END IF;
    
    -- Migrate data from businesses to merchants
    FOR business_record IN 
        SELECT * FROM businesses 
        WHERE NOT EXISTS (
            SELECT 1 FROM merchants 
            WHERE merchants.registration_number = businesses.registration_number
            OR merchants.name = businesses.name
        )
    LOOP
        INSERT INTO merchants (
            name,
            legal_name,
            registration_number,
            tax_id,
            industry,
            industry_code,
            business_type,
            founded_date,
            employee_count,
            annual_revenue,
            
            -- Address fields (extract from JSONB)
            address_street1,
            address_street2,
            address_city,
            address_state,
            address_postal_code,
            address_country,
            address_country_code,
            
            -- Contact info fields (extract from JSONB)
            contact_phone,
            contact_email,
            contact_website,
            contact_primary_contact,
            
            -- New fields from businesses table
            website_url,
            description,
            metadata,
            user_id,
            
            -- Portfolio management fields (with defaults)
            portfolio_type_id,
            risk_level_id,
            compliance_status,
            status,
            created_by,
            created_at,
            updated_at
        ) VALUES (
            business_record.name,
            COALESCE(business_record.name, '') as legal_name,  -- Use name as legal_name if no separate legal_name
            COALESCE(business_record.registration_number, '') as registration_number,
            '' as tax_id,  -- Default empty, will be populated later
            business_record.industry,
            business_record.industry_code,
            '' as business_type,  -- Default empty, will be populated later
            business_record.founded_date,
            business_record.employee_count,
            business_record.annual_revenue,
            
            -- Extract address from JSONB
            business_record.address->>'street1' as address_street1,
            business_record.address->>'street2' as address_street2,
            business_record.address->>'city' as address_city,
            business_record.address->>'state' as address_state,
            business_record.address->>'postal_code' as address_postal_code,
            business_record.address->>'country' as address_country,
            business_record.country_code as address_country_code,
            
            -- Extract contact info from JSONB
            business_record.contact_info->>'phone' as contact_phone,
            business_record.contact_info->>'email' as contact_email,
            business_record.website_url as contact_website,
            business_record.contact_info->>'primary_contact' as contact_primary_contact,
            
            -- New fields
            business_record.website_url,
            business_record.description,
            business_record.metadata,
            business_record.user_id,
            
            -- Portfolio management defaults
            default_portfolio_type_id,
            default_risk_level_id,
            'pending' as compliance_status,
            'active' as status,
            business_record.user_id as created_by,
            business_record.created_at,
            business_record.updated_at
        );
        
        migrated_count := migrated_count + 1;
    END LOOP;
    
    RETURN migrated_count;
END;
$$ LANGUAGE plpgsql;

-- 10. Create function to validate data integrity after migration
CREATE OR REPLACE FUNCTION validate_merchants_migration()
RETURNS TABLE (
    validation_type TEXT,
    status TEXT,
    message TEXT,
    count INTEGER
) AS $$
BEGIN
    -- Check for duplicate registration numbers
    RETURN QUERY
    SELECT 
        'duplicate_registration'::TEXT as validation_type,
        CASE WHEN COUNT(*) > 1 THEN 'FAIL' ELSE 'PASS' END as status,
        'Check for duplicate registration numbers' as message,
        COUNT(*) as count
    FROM merchants 
    WHERE registration_number IS NOT NULL AND registration_number != ''
    GROUP BY registration_number
    HAVING COUNT(*) > 1;
    
    -- Check for missing required fields
    RETURN QUERY
    SELECT 
        'missing_required_fields'::TEXT as validation_type,
        CASE WHEN COUNT(*) > 0 THEN 'FAIL' ELSE 'PASS' END as status,
        'Check for missing required fields' as message,
        COUNT(*) as count
    FROM merchants 
    WHERE name IS NULL OR name = '';
    
    -- Check for invalid foreign key references
    RETURN QUERY
    SELECT 
        'invalid_portfolio_type'::TEXT as validation_type,
        CASE WHEN COUNT(*) > 0 THEN 'FAIL' ELSE 'PASS' END as status,
        'Check for invalid portfolio type references' as message,
        COUNT(*) as count
    FROM merchants m
    LEFT JOIN portfolio_types pt ON m.portfolio_type_id = pt.id
    WHERE m.portfolio_type_id IS NOT NULL AND pt.id IS NULL;
    
    RETURN QUERY
    SELECT 
        'invalid_risk_level'::TEXT as validation_type,
        CASE WHEN COUNT(*) > 0 THEN 'FAIL' ELSE 'PASS' END as status,
        'Check for invalid risk level references' as message,
        COUNT(*) as count
    FROM merchants m
    LEFT JOIN risk_levels rl ON m.risk_level_id = rl.id
    WHERE m.risk_level_id IS NOT NULL AND rl.id IS NULL;
    
    -- Check for invalid user references
    RETURN QUERY
    SELECT 
        'invalid_user_reference'::TEXT as validation_type,
        CASE WHEN COUNT(*) > 0 THEN 'FAIL' ELSE 'PASS' END as status,
        'Check for invalid user references' as message,
        COUNT(*) as count
    FROM merchants m
    LEFT JOIN users u ON m.created_by = u.id
    WHERE m.created_by IS NOT NULL AND u.id IS NULL;
    
    -- Check total migration count
    RETURN QUERY
    SELECT 
        'total_merchants'::TEXT as validation_type,
        'INFO' as status,
        'Total merchants after migration' as message,
        COUNT(*) as count
    FROM merchants;
    
END;
$$ LANGUAGE plpgsql;

-- 11. Create rollback function in case migration needs to be reversed
CREATE OR REPLACE FUNCTION rollback_merchants_enhancement()
RETURNS VOID AS $$
BEGIN
    -- Remove added columns
    ALTER TABLE merchants DROP COLUMN IF EXISTS metadata;
    ALTER TABLE merchants DROP COLUMN IF EXISTS website_url;
    ALTER TABLE merchants DROP COLUMN IF EXISTS description;
    ALTER TABLE merchants DROP COLUMN IF EXISTS user_id;
    
    -- Revert field lengths to original
    ALTER TABLE merchants ALTER COLUMN name TYPE VARCHAR(255);
    ALTER TABLE merchants ALTER COLUMN industry TYPE VARCHAR(100);
    ALTER TABLE merchants ALTER COLUMN industry_code TYPE VARCHAR(20);
    
    -- Restore original constraints
    ALTER TABLE merchants ALTER COLUMN registration_number SET NOT NULL;
    ALTER TABLE merchants ALTER COLUMN legal_name SET NOT NULL;
    ALTER TABLE merchants ALTER COLUMN portfolio_type_id SET NOT NULL;
    ALTER TABLE merchants ALTER COLUMN risk_level_id SET NOT NULL;
    ALTER TABLE merchants ALTER COLUMN created_by SET NOT NULL;
    
    -- Drop added indexes
    DROP INDEX IF EXISTS idx_merchants_metadata;
    DROP INDEX IF EXISTS idx_merchants_website_url;
    DROP INDEX IF EXISTS idx_merchants_description;
    DROP INDEX IF EXISTS idx_merchants_user_id;
    
    -- Drop functions
    DROP FUNCTION IF EXISTS migrate_businesses_to_merchants();
    DROP FUNCTION IF EXISTS validate_merchants_migration();
    DROP FUNCTION IF EXISTS rollback_merchants_enhancement();
END;
$$ LANGUAGE plpgsql;

-- Commit the transaction
COMMIT;

-- Log migration completion
DO $$
BEGIN
    RAISE NOTICE 'Merchants table enhancement migration completed successfully';
    RAISE NOTICE 'Added fields: metadata, website_url, description, user_id';
    RAISE NOTICE 'Enhanced field lengths: name(500), industry(255), industry_code(50)';
    RAISE NOTICE 'Created migration functions: migrate_businesses_to_merchants(), validate_merchants_migration()';
    RAISE NOTICE 'Created rollback function: rollback_merchants_enhancement()';
    RAISE NOTICE 'Next step: Run SELECT migrate_businesses_to_merchants(); to migrate data';
END $$;
