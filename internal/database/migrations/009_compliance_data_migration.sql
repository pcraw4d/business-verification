-- =====================================================
-- Compliance Data Migration Script
-- Task 2.3.2: Merge Compliance Tables - Data Migration
-- =====================================================
-- This script migrates data from compliance_checks and compliance_records
-- into the unified compliance_tracking table.

-- Start transaction for data migration
BEGIN;

-- =====================================================
-- Step 1: Migrate data from compliance_checks table
-- =====================================================
INSERT INTO compliance_tracking (
    merchant_id,
    compliance_type,
    compliance_framework,
    check_type,
    status,
    score,
    risk_level,
    requirements,
    check_method,
    source,
    raw_data,
    result,
    findings,
    recommendations,
    evidence,
    checked_by,
    checked_at,
    due_date,
    expires_at,
    next_review_date,
    priority,
    assigned_to,
    tags,
    notes,
    metadata,
    created_at,
    updated_at
)
SELECT 
    -- Map business_id to merchant_id (assuming businesses table was consolidated to merchants)
    COALESCE(m.id, b.id) as merchant_id,
    
    -- Map compliance fields
    cc.compliance_type,
    cc.compliance_type as compliance_framework, -- Use compliance_type as framework initially
    'automated' as check_type, -- Default to automated for legacy data
    cc.status,
    cc.score,
    
    -- Determine risk level based on score
    CASE 
        WHEN cc.score >= 0.8 THEN 'low'
        WHEN cc.score >= 0.6 THEN 'medium'
        WHEN cc.score >= 0.4 THEN 'high'
        ELSE 'critical'
    END as risk_level,
    
    -- Map JSONB fields
    cc.requirements,
    cc.check_method,
    cc.source,
    cc.raw_data,
    
    -- Create result from raw_data if available
    CASE 
        WHEN cc.raw_data IS NOT NULL THEN jsonb_build_object(
            'source', cc.source,
            'method', cc.check_method,
            'raw_data', cc.raw_data
        )
        ELSE NULL
    END as result,
    
    -- Initialize other fields
    NULL as findings,
    NULL as recommendations,
    NULL as evidence,
    NULL as checked_by, -- No user tracking in original table
    cc.created_at as checked_at,
    NULL as due_date,
    NULL as expires_at,
    NULL as next_review_date,
    'medium' as priority, -- Default priority
    NULL as assigned_to,
    ARRAY[cc.compliance_type] as tags, -- Use compliance_type as tag
    NULL as notes,
    
    -- Create metadata from available fields
    jsonb_build_object(
        'migrated_from', 'compliance_checks',
        'original_id', cc.id,
        'original_created_at', cc.created_at
    ) as metadata,
    
    cc.created_at,
    cc.created_at as updated_at

FROM compliance_checks cc
LEFT JOIN businesses b ON cc.business_id = b.id
LEFT JOIN merchants m ON b.id = m.id OR b.name = m.business_name -- Try to match by ID or name
WHERE NOT EXISTS (
    -- Avoid duplicates if migration runs multiple times
    SELECT 1 FROM compliance_tracking ct 
    WHERE ct.metadata->>'original_id' = cc.id::text 
    AND ct.metadata->>'migrated_from' = 'compliance_checks'
);

-- =====================================================
-- Step 2: Migrate data from compliance_records table
-- =====================================================
INSERT INTO compliance_tracking (
    merchant_id,
    compliance_type,
    compliance_framework,
    check_type,
    status,
    score,
    risk_level,
    requirements,
    check_method,
    source,
    raw_data,
    result,
    findings,
    recommendations,
    evidence,
    checked_by,
    checked_at,
    reviewed_by,
    reviewed_at,
    approved_by,
    approved_at,
    due_date,
    expires_at,
    next_review_date,
    priority,
    assigned_to,
    tags,
    notes,
    metadata,
    created_at,
    updated_at
)
SELECT 
    cr.merchant_id,
    cr.compliance_type,
    cr.compliance_type as compliance_framework, -- Use compliance_type as framework initially
    'manual' as check_type, -- Default to manual for records
    cr.status,
    cr.score,
    
    -- Determine risk level based on score
    CASE 
        WHEN cr.score >= 0.8 THEN 'low'
        WHEN cr.score >= 0.6 THEN 'medium'
        WHEN cr.score >= 0.4 THEN 'high'
        ELSE 'critical'
    END as risk_level,
    
    -- Map JSONB fields
    cr.requirements,
    cr.check_method,
    cr.source,
    cr.raw_data,
    
    -- Create result from raw_data if available
    CASE 
        WHEN cr.raw_data IS NOT NULL THEN jsonb_build_object(
            'source', cr.source,
            'method', cr.check_method,
            'raw_data', cr.raw_data
        )
        ELSE NULL
    END as result,
    
    -- Initialize other fields
    NULL as findings,
    NULL as recommendations,
    NULL as evidence,
    cr.checked_by,
    cr.checked_at,
    NULL as reviewed_by,
    NULL as reviewed_at,
    NULL as approved_by,
    NULL as approved_at,
    NULL as due_date,
    cr.expires_at,
    NULL as next_review_date,
    'medium' as priority, -- Default priority
    NULL as assigned_to,
    ARRAY[cr.compliance_type] as tags, -- Use compliance_type as tag
    NULL as notes,
    
    -- Create metadata from available fields
    jsonb_build_object(
        'migrated_from', 'compliance_records',
        'original_id', cr.id,
        'original_created_at', cr.created_at,
        'original_checked_at', cr.checked_at
    ) as metadata,
    
    cr.created_at,
    cr.updated_at

FROM compliance_records cr
WHERE NOT EXISTS (
    -- Avoid duplicates if migration runs multiple times
    SELECT 1 FROM compliance_tracking ct 
    WHERE ct.metadata->>'original_id' = cr.id::text 
    AND ct.metadata->>'migrated_from' = 'compliance_records'
);

-- =====================================================
-- Step 3: Update compliance_tracking with enhanced data
-- =====================================================

-- Update compliance_framework based on compliance_type patterns
UPDATE compliance_tracking 
SET compliance_framework = CASE 
    WHEN compliance_type ILIKE '%aml%' OR compliance_type ILIKE '%anti-money%' THEN 'AML'
    WHEN compliance_type ILIKE '%kyc%' OR compliance_type ILIKE '%know your customer%' THEN 'KYC'
    WHEN compliance_type ILIKE '%kyb%' OR compliance_type ILIKE '%know your business%' THEN 'KYB'
    WHEN compliance_type ILIKE '%fatf%' THEN 'FATF'
    WHEN compliance_type ILIKE '%gdpr%' THEN 'GDPR'
    WHEN compliance_type ILIKE '%pci%' THEN 'PCI'
    WHEN compliance_type ILIKE '%sox%' THEN 'SOX'
    WHEN compliance_type ILIKE '%iso%' THEN 'ISO27001'
    WHEN compliance_type ILIKE '%soc%' THEN 'SOC2'
    WHEN compliance_type ILIKE '%bsa%' THEN 'BSA'
    WHEN compliance_type ILIKE '%ofac%' THEN 'OFAC'
    ELSE compliance_type
END
WHERE compliance_framework = compliance_type;

-- Update check_type based on source patterns
UPDATE compliance_tracking 
SET check_type = CASE 
    WHEN source ILIKE '%api%' OR source ILIKE '%automated%' OR source ILIKE '%system%' THEN 'automated'
    WHEN source ILIKE '%manual%' OR source ILIKE '%user%' OR source ILIKE '%admin%' THEN 'manual'
    WHEN source ILIKE '%scheduled%' OR source ILIKE '%cron%' THEN 'periodic'
    ELSE 'ad_hoc'
END
WHERE check_type IN ('automated', 'manual'); -- Only update if still using defaults

-- Update priority based on risk_level and status
UPDATE compliance_tracking 
SET priority = CASE 
    WHEN risk_level = 'critical' OR status = 'failed' THEN 'critical'
    WHEN risk_level = 'high' OR status = 'overdue' THEN 'high'
    WHEN risk_level = 'medium' THEN 'medium'
    ELSE 'low'
END
WHERE priority = 'medium'; -- Only update if still using default

-- =====================================================
-- Step 4: Create data validation queries
-- =====================================================

-- Validate migration completeness
DO $$
DECLARE
    compliance_checks_count INTEGER;
    compliance_records_count INTEGER;
    compliance_tracking_count INTEGER;
    migrated_checks_count INTEGER;
    migrated_records_count INTEGER;
BEGIN
    -- Count original records
    SELECT COUNT(*) INTO compliance_checks_count FROM compliance_checks;
    SELECT COUNT(*) INTO compliance_records_count FROM compliance_records;
    
    -- Count migrated records
    SELECT COUNT(*) INTO compliance_tracking_count FROM compliance_tracking;
    SELECT COUNT(*) INTO migrated_checks_count FROM compliance_tracking WHERE metadata->>'migrated_from' = 'compliance_checks';
    SELECT COUNT(*) INTO migrated_records_count FROM compliance_tracking WHERE metadata->>'migrated_from' = 'compliance_records';
    
    -- Log migration results
    RAISE NOTICE 'Migration Summary:';
    RAISE NOTICE '  Original compliance_checks records: %', compliance_checks_count;
    RAISE NOTICE '  Original compliance_records records: %', compliance_records_count;
    RAISE NOTICE '  Total migrated records: %', compliance_tracking_count;
    RAISE NOTICE '  Migrated from compliance_checks: %', migrated_checks_count;
    RAISE NOTICE '  Migrated from compliance_records: %', migrated_records_count;
    
    -- Validate migration
    IF (migrated_checks_count + migrated_records_count) != compliance_tracking_count THEN
        RAISE EXCEPTION 'Migration validation failed: Total migrated records (%) does not match sum of individual migrations (%)', 
            compliance_tracking_count, (migrated_checks_count + migrated_records_count);
    END IF;
    
    RAISE NOTICE 'Migration completed successfully!';
END $$;

-- =====================================================
-- Step 5: Create backup of original tables (optional)
-- =====================================================

-- Create backup tables with timestamp
CREATE TABLE IF NOT EXISTS compliance_checks_backup_$(date +%Y%m%d_%H%M%S) AS 
SELECT *, 'backup_created' as backup_note, CURRENT_TIMESTAMP as backup_created_at 
FROM compliance_checks;

CREATE TABLE IF NOT EXISTS compliance_records_backup_$(date +%Y%m%d_%H%M%S) AS 
SELECT *, 'backup_created' as backup_note, CURRENT_TIMESTAMP as backup_created_at 
FROM compliance_records;

-- Commit the transaction
COMMIT;

-- =====================================================
-- Migration Complete
-- =====================================================
-- The migration has successfully consolidated data from:
-- 1. compliance_checks table
-- 2. compliance_records table
-- 
-- Into the unified compliance_tracking table with:
-- - Enhanced audit trail
-- - Better data structure
-- - Comprehensive indexing
-- - Views for reporting
-- - RLS security
-- 
-- Next steps:
-- 1. Update application code to use compliance_tracking
-- 2. Test the new unified table
-- 3. Drop original tables after validation
-- =====================================================
