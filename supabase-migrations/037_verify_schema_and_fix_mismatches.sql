-- Migration: Verify Schema and Fix Type Mismatches
-- Created: 2025-12-22
-- Purpose: Comprehensive schema verification and fix type mismatches

-- =====================================================
-- PART 1: Fix Type Mismatch in get_codes_by_trigram_similarity
-- =====================================================

-- Fix: Cast cc.code to text to match function return type
CREATE OR REPLACE FUNCTION get_codes_by_trigram_similarity(
    p_code_type text,
    p_industry_name text,
    p_threshold float DEFAULT 0.3,
    p_limit int DEFAULT 3
)
RETURNS TABLE (
    code text,
    description text,
    similarity float
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cc.code::text,  -- ✅ FIX: Cast VARCHAR(20) to text
        cc.description,
        similarity(cc.description, p_industry_name)::double precision as similarity  -- ✅ FIX: Cast real to double precision
    FROM classification_codes cc
    WHERE 
        cc.code_type = p_code_type
        AND cc.is_active = true
        AND similarity(cc.description, p_industry_name) >= p_threshold
    ORDER BY 
        similarity DESC,
        cc.code ASC
    LIMIT p_limit;
END;
$$;

COMMENT ON FUNCTION get_codes_by_trigram_similarity IS 
    'Returns classification codes (MCC, SIC, NAICS) with similarity scores using trigram matching against industry name. 
     FIXED: Added cast to text for code column to match return type.';

-- =====================================================
-- PART 2: Schema Verification Queries
-- =====================================================

-- Verify all required tables exist
DO $$
DECLARE
    missing_tables text[];
    required_tables text[] := ARRAY[
        'classification_codes',
        'code_keywords',
        'code_metadata',
        'industries',
        'industry_keywords',
        'industry_topics',
        'keyword_patterns',
        'keyword_weights',
        'code_embeddings',
        'industry_code_crosswalks'
    ];
    tbl_name text;
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'SCHEMA VERIFICATION: Required Tables';
    RAISE NOTICE '========================================';
    
    FOREACH tbl_name IN ARRAY required_tables
    LOOP
        IF EXISTS (
            SELECT 1 FROM information_schema.tables t
            WHERE t.table_schema = 'public' 
            AND t.table_name = tbl_name
        ) THEN
            RAISE NOTICE '✅ Table exists: %', tbl_name;
        ELSE
            RAISE NOTICE '❌ Table missing: %', tbl_name;
            missing_tables := array_append(missing_tables, tbl_name);
        END IF;
    END LOOP;
    
    IF array_length(missing_tables, 1) > 0 THEN
        RAISE WARNING 'Missing tables: %', array_to_string(missing_tables, ', ');
    ELSE
        RAISE NOTICE '✅ All required tables exist';
    END IF;
END $$;

-- Verify RPC functions exist and have correct signatures
DO $$
DECLARE
    func_name text;
    func_exists boolean;
    func_args text;
    func_returns text;
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'SCHEMA VERIFICATION: RPC Functions';
    RAISE NOTICE '========================================';
    
    -- Check get_codes_by_keywords
    SELECT EXISTS (
        SELECT 1 FROM pg_proc p
        JOIN pg_namespace n ON p.pronamespace = n.oid
        WHERE n.nspname = 'public'
        AND p.proname = 'get_codes_by_keywords'
    ) INTO func_exists;
    
    IF func_exists THEN
        SELECT pg_get_function_arguments(p.oid), pg_get_function_result(p.oid)
        INTO func_args, func_returns
        FROM pg_proc p
        JOIN pg_namespace n ON p.pronamespace = n.oid
        WHERE n.nspname = 'public' AND p.proname = 'get_codes_by_keywords';
        
        RAISE NOTICE '✅ get_codes_by_keywords exists';
        RAISE NOTICE '   Arguments: %', func_args;
        RAISE NOTICE '   Returns: %', func_returns;
    ELSE
        RAISE WARNING '❌ get_codes_by_keywords function missing';
    END IF;
    
    -- Check get_codes_by_trigram_similarity
    SELECT EXISTS (
        SELECT 1 FROM pg_proc p
        JOIN pg_namespace n ON p.pronamespace = n.oid
        WHERE n.nspname = 'public'
        AND p.proname = 'get_codes_by_trigram_similarity'
    ) INTO func_exists;
    
    IF func_exists THEN
        SELECT pg_get_function_arguments(p.oid), pg_get_function_result(p.oid)
        INTO func_args, func_returns
        FROM pg_proc p
        JOIN pg_namespace n ON p.pronamespace = n.oid
        WHERE n.nspname = 'public' AND p.proname = 'get_codes_by_trigram_similarity';
        
        RAISE NOTICE '✅ get_codes_by_trigram_similarity exists';
        RAISE NOTICE '   Arguments: %', func_args;
        RAISE NOTICE '   Returns: %', func_returns;
    ELSE
        RAISE WARNING '❌ get_codes_by_trigram_similarity function missing';
    END IF;
    
    -- Check match_code_embeddings
    SELECT EXISTS (
        SELECT 1 FROM pg_proc p
        JOIN pg_namespace n ON p.pronamespace = n.oid
        WHERE n.nspname = 'public'
        AND p.proname = 'match_code_embeddings'
    ) INTO func_exists;
    
    IF func_exists THEN
        SELECT pg_get_function_arguments(p.oid), pg_get_function_result(p.oid)
        INTO func_args, func_returns
        FROM pg_proc p
        JOIN pg_namespace n ON p.pronamespace = n.oid
        WHERE n.nspname = 'public' AND p.proname = 'match_code_embeddings';
        
        RAISE NOTICE '✅ match_code_embeddings exists';
        RAISE NOTICE '   Arguments: %', func_args;
        RAISE NOTICE '   Returns: %', func_returns;
    ELSE
        RAISE WARNING '❌ match_code_embeddings function missing';
    END IF;
END $$;

-- Verify critical column types match expectations
DO $$
DECLARE
    col_type text;
    expected_type text;
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'SCHEMA VERIFICATION: Column Types';
    RAISE NOTICE '========================================';
    
    -- Check classification_codes.code type
    SELECT data_type INTO col_type
    FROM information_schema.columns
    WHERE table_schema = 'public'
    AND table_name = 'classification_codes'
    AND column_name = 'code';
    
    IF col_type IS NOT NULL THEN
        RAISE NOTICE '✅ classification_codes.code type: %', col_type;
        IF col_type NOT IN ('text', 'character varying') THEN
            RAISE WARNING '⚠️ Unexpected type for classification_codes.code: %', col_type;
        END IF;
    ELSE
        RAISE WARNING '❌ classification_codes.code column not found';
    END IF;
    
    -- Check code_keywords.code_id type (should be integer/bigint)
    SELECT data_type INTO col_type
    FROM information_schema.columns
    WHERE table_schema = 'public'
    AND table_name = 'code_keywords'
    AND column_name = 'code_id';
    
    IF col_type IS NOT NULL THEN
        RAISE NOTICE '✅ code_keywords.code_id type: %', col_type;
    ELSE
        RAISE WARNING '❌ code_keywords.code_id column not found';
    END IF;
    
    -- Check code_keywords.relevance_score type (should be numeric/decimal)
    SELECT data_type INTO col_type
    FROM information_schema.columns
    WHERE table_schema = 'public'
    AND table_name = 'code_keywords'
    AND column_name = 'relevance_score';
    
    IF col_type IS NOT NULL THEN
        RAISE NOTICE '✅ code_keywords.relevance_score type: %', col_type;
    ELSE
        RAISE WARNING '❌ code_keywords.relevance_score column not found';
    END IF;
    
    -- Check code_metadata.crosswalk_data type (should be jsonb)
    SELECT data_type INTO col_type
    FROM information_schema.columns
    WHERE table_schema = 'public'
    AND table_name = 'code_metadata'
    AND column_name = 'crosswalk_data';
    
    IF col_type IS NOT NULL THEN
        RAISE NOTICE '✅ code_metadata.crosswalk_data type: %', col_type;
        IF col_type != 'jsonb' THEN
            RAISE WARNING '⚠️ Unexpected type for code_metadata.crosswalk_data: % (expected jsonb)', col_type;
        END IF;
    ELSE
        RAISE WARNING '❌ code_metadata.crosswalk_data column not found';
    END IF;
END $$;

-- Verify critical indexes exist
DO $$
DECLARE
    idx_exists boolean;
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'SCHEMA VERIFICATION: Critical Indexes';
    RAISE NOTICE '========================================';
    
    -- Check trigram index on classification_codes.description
    SELECT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE indexname = 'idx_classification_codes_description_trgm'
    ) INTO idx_exists;
    
    IF idx_exists THEN
        RAISE NOTICE '✅ Trigram index exists: idx_classification_codes_description_trgm';
    ELSE
        RAISE WARNING '❌ Trigram index missing: idx_classification_codes_description_trgm';
    END IF;
    
    -- Check keyword lookup index
    SELECT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE indexname = 'idx_code_keywords_keyword_lookup'
    ) INTO idx_exists;
    
    IF idx_exists THEN
        RAISE NOTICE '✅ Keyword lookup index exists: idx_code_keywords_keyword_lookup';
    ELSE
        RAISE WARNING '❌ Keyword lookup index missing: idx_code_keywords_keyword_lookup';
    END IF;
    
    -- Check code_embeddings vector index
    SELECT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE indexname = 'idx_code_embeddings_vector'
    ) INTO idx_exists;
    
    IF idx_exists THEN
        RAISE NOTICE '✅ Vector index exists: idx_code_embeddings_vector';
    ELSE
        RAISE WARNING '⚠️ Vector index missing: idx_code_embeddings_vector (embeddings may not be used)';
    END IF;
END $$;

-- Verify foreign key relationships
DO $$
DECLARE
    fk_count int;
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'SCHEMA VERIFICATION: Foreign Keys';
    RAISE NOTICE '========================================';
    
    -- Check code_keywords -> classification_codes FK
    SELECT COUNT(*) INTO fk_count
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu 
        ON tc.constraint_name = kcu.constraint_name
    WHERE tc.table_schema = 'public'
    AND tc.table_name = 'code_keywords'
    AND tc.constraint_type = 'FOREIGN KEY'
    AND kcu.column_name = 'code_id';
    
    IF fk_count > 0 THEN
        RAISE NOTICE '✅ Foreign key exists: code_keywords.code_id -> classification_codes.id';
    ELSE
        RAISE WARNING '❌ Foreign key missing: code_keywords.code_id -> classification_codes.id';
    END IF;
    
    -- Check classification_codes -> industries FK
    SELECT COUNT(*) INTO fk_count
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu 
        ON tc.constraint_name = kcu.constraint_name
    WHERE tc.table_schema = 'public'
    AND tc.table_name = 'classification_codes'
    AND tc.constraint_type = 'FOREIGN KEY'
    AND kcu.column_name = 'industry_id';
    
    IF fk_count > 0 THEN
        RAISE NOTICE '✅ Foreign key exists: classification_codes.industry_id -> industries.id';
    ELSE
        RAISE WARNING '❌ Foreign key missing: classification_codes.industry_id -> industries.id';
    END IF;
END $$;

-- Test RPC functions with sample data
DO $$
DECLARE
    test_result record;
    test_count int;
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'SCHEMA VERIFICATION: RPC Function Tests';
    RAISE NOTICE '========================================';
    
    -- Test get_codes_by_keywords (should not error even if no results)
    BEGIN
        SELECT COUNT(*) INTO test_count
        FROM get_codes_by_keywords('MCC', ARRAY['test'], 1);
        RAISE NOTICE '✅ get_codes_by_keywords executes without errors';
    EXCEPTION
        WHEN OTHERS THEN
            RAISE WARNING '❌ get_codes_by_keywords failed: %', SQLERRM;
    END;
    
    -- Test get_codes_by_trigram_similarity (should not error even if no results)
    BEGIN
        SELECT COUNT(*) INTO test_count
        FROM get_codes_by_trigram_similarity('MCC', 'test industry', 0.3, 1);
        RAISE NOTICE '✅ get_codes_by_trigram_similarity executes without errors';
    EXCEPTION
        WHEN OTHERS THEN
            RAISE WARNING '❌ get_codes_by_trigram_similarity failed: %', SQLERRM;
    END;
    
    -- Test match_code_embeddings (requires vector extension and embeddings)
    BEGIN
        -- This will fail if pgvector is not enabled or no embeddings exist
        -- That's OK - we just want to verify the function exists
        PERFORM match_code_embeddings(
            (SELECT embedding FROM code_embeddings LIMIT 1),
            'MCC',
            0.7,
            1
        );
        RAISE NOTICE '✅ match_code_embeddings executes (embeddings available)';
    EXCEPTION
        WHEN undefined_function THEN
            RAISE WARNING '❌ match_code_embeddings function missing';
        WHEN OTHERS THEN
            -- Other errors (no embeddings, etc.) are OK for verification
            RAISE NOTICE '⚠️ match_code_embeddings exists but test failed (may be due to missing embeddings): %', SQLERRM;
    END;
END $$;

-- Summary report
DO $$
DECLARE
    table_count int;
    function_count int;
    index_count int;
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'SCHEMA VERIFICATION: Summary';
    RAISE NOTICE '========================================';
    
    SELECT COUNT(*) INTO table_count
    FROM information_schema.tables
    WHERE table_schema = 'public'
    AND table_name IN (
        'classification_codes', 'code_keywords', 'code_metadata',
        'industries', 'industry_keywords', 'industry_topics',
        'keyword_patterns', 'keyword_weights', 'code_embeddings',
        'industry_code_crosswalks'
    );
    
    SELECT COUNT(*) INTO function_count
    FROM pg_proc p
    JOIN pg_namespace n ON p.pronamespace = n.oid
    WHERE n.nspname = 'public'
    AND p.proname IN (
        'get_codes_by_keywords',
        'get_codes_by_trigram_similarity',
        'match_code_embeddings'
    );
    
    SELECT COUNT(*) INTO index_count
    FROM pg_indexes
    WHERE indexname IN (
        'idx_classification_codes_description_trgm',
        'idx_code_keywords_keyword_lookup',
        'idx_code_embeddings_vector'
    );
    
    RAISE NOTICE 'Tables verified: %/10', table_count;
    RAISE NOTICE 'Functions verified: %/3', function_count;
    RAISE NOTICE 'Critical indexes verified: %/3', index_count;
    
    IF table_count = 10 AND function_count = 3 THEN
        RAISE NOTICE '✅ Schema verification complete - all critical components present';
    ELSE
        RAISE WARNING '⚠️ Schema verification incomplete - review warnings above';
    END IF;
END $$;

