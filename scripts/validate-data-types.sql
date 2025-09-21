-- Data Type and Format Validation Script
-- This script validates data types and formats across all columns in the database

-- ============================================================================
-- DATA TYPE VALIDATION ANALYSIS
-- ============================================================================

-- 1. Get all columns and their data types
SELECT 
    'COLUMN DATA TYPES' as test_type,
    t.table_name,
    c.column_name,
    c.data_type,
    c.is_nullable,
    c.character_maximum_length,
    c.numeric_precision,
    c.numeric_scale,
    c.column_default
FROM 
    information_schema.tables t
    JOIN information_schema.columns c ON t.table_name = c.table_name
WHERE 
    t.table_schema = 'public'
    AND t.table_type = 'BASE TABLE'
    AND c.table_schema = 'public'
ORDER BY 
    t.table_name, c.ordinal_position;

-- ============================================================================
-- EMAIL FORMAT VALIDATION
-- ============================================================================

-- Check email columns for valid email format
SELECT 
    'EMAIL FORMAT VALIDATION' as test_type,
    'users.email' as column_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN email IS NULL THEN 1 END) as null_count,
    COUNT(CASE WHEN email IS NOT NULL AND email !~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$' THEN 1 END) as invalid_email_count,
    ROUND(
        COUNT(CASE WHEN email IS NOT NULL AND email !~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$' THEN 1 END) * 100.0 / 
        COUNT(CASE WHEN email IS NOT NULL THEN 1 END), 2
    ) as invalid_email_percentage
FROM users
WHERE EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'email');

-- ============================================================================
-- UUID FORMAT VALIDATION
-- ============================================================================

-- Check UUID columns for valid UUID format
SELECT 
    'UUID FORMAT VALIDATION' as test_type,
    'merchants.id' as column_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN id IS NULL THEN 1 END) as null_count,
    COUNT(CASE WHEN id IS NOT NULL AND id::text !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$' THEN 1 END) as invalid_uuid_count
FROM merchants
WHERE EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'id' AND data_type = 'uuid');

-- ============================================================================
-- PHONE NUMBER FORMAT VALIDATION
-- ============================================================================

-- Check phone columns for valid phone format (E.164 format)
SELECT 
    'PHONE FORMAT VALIDATION' as test_type,
    'merchants.phone' as column_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN phone IS NULL THEN 1 END) as null_count,
    COUNT(CASE WHEN phone IS NOT NULL AND phone !~ '^\+?[1-9]\d{1,14}$' THEN 1 END) as invalid_phone_count,
    ROUND(
        COUNT(CASE WHEN phone IS NOT NULL AND phone !~ '^\+?[1-9]\d{1,14}$' THEN 1 END) * 100.0 / 
        COUNT(CASE WHEN phone IS NOT NULL THEN 1 END), 2
    ) as invalid_phone_percentage
FROM merchants
WHERE EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'phone');

-- ============================================================================
-- URL FORMAT VALIDATION
-- ============================================================================

-- Check URL columns for valid URL format
SELECT 
    'URL FORMAT VALIDATION' as test_type,
    'merchants.website' as column_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN website IS NULL THEN 1 END) as null_count,
    COUNT(CASE WHEN website IS NOT NULL AND website !~ '^https?://[^\s/$.?#].[^\s]*$' THEN 1 END) as invalid_url_count,
    ROUND(
        COUNT(CASE WHEN website IS NOT NULL AND website !~ '^https?://[^\s/$.?#].[^\s]*$' THEN 1 END) * 100.0 / 
        COUNT(CASE WHEN website IS NOT NULL THEN 1 END), 2
    ) as invalid_url_percentage
FROM merchants
WHERE EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'website');

-- ============================================================================
-- DATE FORMAT VALIDATION
-- ============================================================================

-- Check date columns for valid date format
SELECT 
    'DATE FORMAT VALIDATION' as test_type,
    'merchants.created_at' as column_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN created_at IS NULL THEN 1 END) as null_count,
    COUNT(CASE WHEN created_at IS NOT NULL AND created_at::text !~ '^\d{4}-\d{2}-\d{2}' THEN 1 END) as invalid_date_count
FROM merchants
WHERE EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'created_at');

-- ============================================================================
-- STRING LENGTH VALIDATION
-- ============================================================================

-- Check varchar columns for length constraints
SELECT 
    'STRING LENGTH VALIDATION' as test_type,
    t.table_name,
    c.column_name,
    c.character_maximum_length,
    COUNT(*) as total_records,
    COUNT(CASE WHEN c.column_name IS NOT NULL AND LENGTH(c.column_name::text) > c.character_maximum_length THEN 1 END) as oversized_count
FROM 
    information_schema.tables t
    JOIN information_schema.columns c ON t.table_name = c.table_name
WHERE 
    t.table_schema = 'public'
    AND t.table_type = 'BASE TABLE'
    AND c.table_schema = 'public'
    AND c.data_type = 'character varying'
    AND c.character_maximum_length IS NOT NULL
GROUP BY 
    t.table_name, c.column_name, c.character_maximum_length
ORDER BY 
    t.table_name, c.column_name;

-- ============================================================================
-- NUMERIC RANGE VALIDATION
-- ============================================================================

-- Check numeric columns for valid ranges
SELECT 
    'NUMERIC RANGE VALIDATION' as test_type,
    t.table_name,
    c.column_name,
    c.data_type,
    c.numeric_precision,
    c.numeric_scale,
    COUNT(*) as total_records,
    COUNT(CASE WHEN c.column_name IS NULL THEN 1 END) as null_count
FROM 
    information_schema.tables t
    JOIN information_schema.columns c ON t.table_name = c.table_name
WHERE 
    t.table_schema = 'public'
    AND t.table_type = 'BASE TABLE'
    AND c.table_schema = 'public'
    AND c.data_type IN ('integer', 'bigint', 'smallint', 'decimal', 'numeric', 'real', 'double precision')
GROUP BY 
    t.table_name, c.column_name, c.data_type, c.numeric_precision, c.numeric_scale
ORDER BY 
    t.table_name, c.column_name;

-- ============================================================================
-- BOOLEAN VALIDATION
-- ============================================================================

-- Check boolean columns for valid boolean values
SELECT 
    'BOOLEAN VALIDATION' as test_type,
    t.table_name,
    c.column_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN c.column_name IS NULL THEN 1 END) as null_count,
    COUNT(CASE WHEN c.column_name IS NOT NULL AND c.column_name::text NOT IN ('true', 'false', 't', 'f', '1', '0') THEN 1 END) as invalid_boolean_count
FROM 
    information_schema.tables t
    JOIN information_schema.columns c ON t.table_name = c.table_name
WHERE 
    t.table_schema = 'public'
    AND t.table_type = 'BASE TABLE'
    AND c.table_schema = 'public'
    AND c.data_type = 'boolean'
GROUP BY 
    t.table_name, c.column_name
ORDER BY 
    t.table_name, c.column_name;

-- ============================================================================
-- JSON FORMAT VALIDATION
-- ============================================================================

-- Check JSON/JSONB columns for valid JSON format
SELECT 
    'JSON FORMAT VALIDATION' as test_type,
    t.table_name,
    c.column_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN c.column_name IS NULL THEN 1 END) as null_count,
    COUNT(CASE WHEN c.column_name IS NOT NULL AND NOT (c.column_name::text ~ '^[\{\[].*[\}\]]$') THEN 1 END) as invalid_json_count
FROM 
    information_schema.tables t
    JOIN information_schema.columns c ON t.table_name = c.table_name
WHERE 
    t.table_schema = 'public'
    AND t.table_type = 'BASE TABLE'
    AND c.table_schema = 'public'
    AND c.data_type IN ('json', 'jsonb')
GROUP BY 
    t.table_name, c.column_name
ORDER BY 
    t.table_name, c.column_name;

-- ============================================================================
-- NULL CONSTRAINT VALIDATION
-- ============================================================================

-- Check for NULL values in columns that should not be NULL
SELECT 
    'NULL CONSTRAINT VALIDATION' as test_type,
    t.table_name,
    c.column_name,
    c.is_nullable,
    COUNT(*) as total_records,
    COUNT(CASE WHEN c.column_name IS NULL THEN 1 END) as null_count,
    ROUND(
        COUNT(CASE WHEN c.column_name IS NULL THEN 1 END) * 100.0 / COUNT(*), 2
    ) as null_percentage
FROM 
    information_schema.tables t
    JOIN information_schema.columns c ON t.table_name = c.table_name
WHERE 
    t.table_schema = 'public'
    AND t.table_type = 'BASE TABLE'
    AND c.table_schema = 'public'
    AND c.is_nullable = 'NO'
GROUP BY 
    t.table_name, c.column_name, c.is_nullable
HAVING 
    COUNT(CASE WHEN c.column_name IS NULL THEN 1 END) > 0
ORDER BY 
    null_percentage DESC, t.table_name, c.column_name;

-- ============================================================================
-- DEFAULT VALUE VALIDATION
-- ============================================================================

-- Check columns with default values
SELECT 
    'DEFAULT VALUE VALIDATION' as test_type,
    t.table_name,
    c.column_name,
    c.column_default,
    COUNT(*) as total_records,
    COUNT(CASE WHEN c.column_name IS NULL THEN 1 END) as null_count
FROM 
    information_schema.tables t
    JOIN information_schema.columns c ON t.table_name = c.table_name
WHERE 
    t.table_schema = 'public'
    AND t.table_type = 'BASE TABLE'
    AND c.table_schema = 'public'
    AND c.column_default IS NOT NULL
GROUP BY 
    t.table_name, c.column_name, c.column_default
ORDER BY 
    t.table_name, c.column_name;

-- ============================================================================
-- DATA TYPE CONSISTENCY CHECK
-- ============================================================================

-- Check for data type inconsistencies across similar columns
SELECT 
    'DATA TYPE CONSISTENCY' as test_type,
    c1.table_name as table1,
    c1.column_name as column1,
    c1.data_type as type1,
    c2.table_name as table2,
    c2.column_name as column2,
    c2.data_type as type2,
    CASE 
        WHEN c1.data_type = c2.data_type THEN 'CONSISTENT'
        ELSE 'INCONSISTENT'
    END as consistency_status
FROM 
    information_schema.columns c1
    JOIN information_schema.columns c2 ON c1.column_name = c2.column_name
WHERE 
    c1.table_schema = 'public'
    AND c2.table_schema = 'public'
    AND c1.table_name != c2.table_name
    AND c1.data_type != c2.data_type
    AND c1.column_name IN ('id', 'created_at', 'updated_at', 'user_id', 'merchant_id')
ORDER BY 
    c1.column_name, c1.table_name, c2.table_name;

-- ============================================================================
-- COMPREHENSIVE DATA TYPE SUMMARY
-- ============================================================================

-- Summary of all data types in the database
SELECT 
    'DATA TYPE SUMMARY' as test_type,
    c.data_type,
    COUNT(*) as column_count,
    COUNT(DISTINCT t.table_name) as table_count
FROM 
    information_schema.tables t
    JOIN information_schema.columns c ON t.table_name = c.table_name
WHERE 
    t.table_schema = 'public'
    AND t.table_type = 'BASE TABLE'
    AND c.table_schema = 'public'
GROUP BY 
    c.data_type
ORDER BY 
    column_count DESC;

-- ============================================================================
-- TEST COMPLETION SUMMARY
-- ============================================================================

SELECT 
    'TEST COMPLETION SUMMARY' as test_type,
    'Data Type and Format Validation Complete' as status,
    NOW() as completion_time,
    'Review all results above for any data type or format issues' as next_steps;
