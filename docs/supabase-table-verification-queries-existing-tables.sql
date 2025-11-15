-- Supabase Table Structure Verification - Existing Tables Only
-- Run these queries for tables that exist: merchants and risk_assessments

-- ============================================================================
-- Check merchants Table Structure
-- ============================================================================

SELECT 
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_schema = 'public'
    AND table_name = 'merchants'
ORDER BY ordinal_position;

-- ============================================================================
-- Check risk_assessments Table Structure
-- ============================================================================

SELECT 
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_schema = 'public'
    AND table_name = 'risk_assessments'
ORDER BY ordinal_position;

-- ============================================================================
-- Expected Columns for merchants Table (from migration 005)
-- ============================================================================
-- Expected columns:
-- - id (UUID)
-- - name (VARCHAR)
-- - legal_name (VARCHAR)
-- - registration_number (VARCHAR)
-- - tax_id (VARCHAR)
-- - industry (VARCHAR)
-- - industry_code (VARCHAR)
-- - business_type (VARCHAR)
-- - founded_date (DATE)
-- - employee_count (INTEGER)
-- - annual_revenue (DECIMAL)
-- - address_street1, address_street2, address_city, address_state, etc.
-- - contact_phone, contact_email, contact_website
-- - portfolio_type_id (UUID)
-- - risk_level_id (UUID)
-- - compliance_status (VARCHAR)
-- - status (VARCHAR)
-- - created_by (UUID)
-- - created_at (TIMESTAMP)
-- - updated_at (TIMESTAMP)

-- ============================================================================
-- Expected Columns for risk_assessments Table (from migration 010)
-- ============================================================================
-- Expected columns:
-- - id (UUID)
-- - merchant_id (VARCHAR) - Added in migration 010
-- - status (VARCHAR) - pending, processing, completed, failed (Added in migration 010)
-- - options (JSONB) - Added in migration 010
-- - result (JSONB) - Added in migration 010
-- - progress (INTEGER) - 0-100 (Added in migration 010)
-- - estimated_completion (TIMESTAMP) - Added in migration 010
-- - completed_at (TIMESTAMP) - Added in migration 010
-- - overall_score (DECIMAL) - May be named risk_score in some schemas
-- - risk_level (VARCHAR)
-- - risk_factors (JSONB) - May be named risk_factors or factors
-- - assessment_method (VARCHAR)
-- - assessment_date (TIMESTAMP)
-- - expires_at (TIMESTAMP)
-- - created_at (TIMESTAMP)
-- - updated_at (TIMESTAMP)
-- - business_id (UUID) - May exist from older schema
-- - user_id (UUID) - May exist from older schema

