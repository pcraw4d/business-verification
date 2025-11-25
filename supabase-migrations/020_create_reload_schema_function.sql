-- Create reload_schema function to refresh PostgREST schema cache
-- This function can be called via RPC to force PostgREST to reload its schema cache

CREATE OR REPLACE FUNCTION reload_schema()
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
    -- Notify PostgREST to reload schema
    -- This is done by calling NOTIFY on a channel that PostgREST listens to
    PERFORM pg_notify('pgrst', 'reload schema');
    
    -- Alternative: Force a schema reload by touching the schema
    -- This works by querying information_schema which forces PostgREST to refresh
    PERFORM 1 FROM information_schema.tables WHERE table_schema = 'public';
END;
$$;

-- Grant execute permission to authenticated users and service role
GRANT EXECUTE ON FUNCTION reload_schema() TO authenticated;
GRANT EXECUTE ON FUNCTION reload_schema() TO service_role;
GRANT EXECUTE ON FUNCTION reload_schema() TO anon;

-- Add comment
COMMENT ON FUNCTION reload_schema() IS 'Forces PostgREST to reload its schema cache. Call via RPC: POST /rest/v1/rpc/reload_schema';

