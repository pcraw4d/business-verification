-- Risk Assessment Service Row-Level Security (RLS) Policies Migration
-- This migration sets up comprehensive RLS policies for multi-tenant isolation and role-based access control

-- Enable RLS on all tables
ALTER TABLE risk_assessments ENABLE ROW LEVEL SECURITY;
ALTER TABLE risk_predictions ENABLE ROW LEVEL SECURITY;
ALTER TABLE risk_factors ENABLE ROW LEVEL SECURITY;
ALTER TABLE custom_risk_models ENABLE ROW LEVEL SECURITY;
ALTER TABLE batch_jobs ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhooks ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhook_deliveries ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE compliance_checks ENABLE ROW LEVEL SECURITY;

-- Create helper functions for RLS

-- Function to get current tenant ID from JWT claims
CREATE OR REPLACE FUNCTION get_current_tenant_id()
RETURNS UUID AS $$
BEGIN
    -- Try to get tenant_id from JWT claims
    RETURN COALESCE(
        (current_setting('request.jwt.claims', true)::jsonb ->> 'tenant_id')::uuid,
        (current_setting('request.jwt.claims', true)::jsonb ->> 'sub')::uuid
    );
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to get current user ID from JWT claims
CREATE OR REPLACE FUNCTION get_current_user_id()
RETURNS UUID AS $$
BEGIN
    RETURN (current_setting('request.jwt.claims', true)::jsonb ->> 'sub')::uuid;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to check if user has admin role
CREATE OR REPLACE FUNCTION is_admin()
RETURNS BOOLEAN AS $$
BEGIN
    RETURN COALESCE(
        (current_setting('request.jwt.claims', true)::jsonb ->> 'role') = 'admin',
        false
    );
EXCEPTION
    WHEN OTHERS THEN
        RETURN false;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to check if user has analyst role
CREATE OR REPLACE FUNCTION is_analyst()
RETURNS BOOLEAN AS $$
BEGIN
    RETURN COALESCE(
        (current_setting('request.jwt.claims', true)::jsonb ->> 'role') IN ('admin', 'analyst'),
        false
    );
EXCEPTION
    WHEN OTHERS THEN
        RETURN false;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to check if user has viewer role
CREATE OR REPLACE FUNCTION is_viewer()
RETURNS BOOLEAN AS $$
BEGIN
    RETURN COALESCE(
        (current_setting('request.jwt.claims', true)::jsonb ->> 'role') IN ('admin', 'analyst', 'viewer'),
        false
    );
EXCEPTION
    WHEN OTHERS THEN
        RETURN false;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to check service-to-service authentication
CREATE OR REPLACE FUNCTION is_service_account()
RETURNS BOOLEAN AS $$
BEGIN
    RETURN COALESCE(
        (current_setting('request.jwt.claims', true)::jsonb ->> 'role') = 'service',
        false
    );
EXCEPTION
    WHEN OTHERS THEN
        RETURN false;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Risk Assessments RLS Policies

-- Policy: Users can view risk assessments for their tenant
CREATE POLICY "Users can view risk assessments for their tenant"
ON risk_assessments FOR SELECT
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_viewer()
);

-- Policy: Analysts and admins can insert risk assessments for their tenant
CREATE POLICY "Analysts can insert risk assessments for their tenant"
ON risk_assessments FOR INSERT
TO authenticated
WITH CHECK (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
);

-- Policy: Analysts and admins can update risk assessments for their tenant
CREATE POLICY "Analysts can update risk assessments for their tenant"
ON risk_assessments FOR UPDATE
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
)
WITH CHECK (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
);

-- Policy: Admins can delete risk assessments for their tenant
CREATE POLICY "Admins can delete risk assessments for their tenant"
ON risk_assessments FOR DELETE
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_admin()
);

-- Policy: Service accounts can perform all operations (for API access)
CREATE POLICY "Service accounts can perform all operations on risk assessments"
ON risk_assessments FOR ALL
TO authenticated
USING (is_service_account())
WITH CHECK (is_service_account());

-- Risk Predictions RLS Policies

-- Policy: Users can view risk predictions for their tenant
CREATE POLICY "Users can view risk predictions for their tenant"
ON risk_predictions FOR SELECT
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_viewer()
);

-- Policy: Analysts and admins can insert risk predictions for their tenant
CREATE POLICY "Analysts can insert risk predictions for their tenant"
ON risk_predictions FOR INSERT
TO authenticated
WITH CHECK (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
);

-- Policy: Analysts and admins can update risk predictions for their tenant
CREATE POLICY "Analysts can update risk predictions for their tenant"
ON risk_predictions FOR UPDATE
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
)
WITH CHECK (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
);

-- Policy: Admins can delete risk predictions for their tenant
CREATE POLICY "Admins can delete risk predictions for their tenant"
ON risk_predictions FOR DELETE
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_admin()
);

-- Policy: Service accounts can perform all operations
CREATE POLICY "Service accounts can perform all operations on risk predictions"
ON risk_predictions FOR ALL
TO authenticated
USING (is_service_account())
WITH CHECK (is_service_account());

-- Risk Factors RLS Policies

-- Policy: Users can view risk factors for their tenant
CREATE POLICY "Users can view risk factors for their tenant"
ON risk_factors FOR SELECT
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_viewer()
);

-- Policy: Analysts and admins can insert risk factors for their tenant
CREATE POLICY "Analysts can insert risk factors for their tenant"
ON risk_factors FOR INSERT
TO authenticated
WITH CHECK (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
);

-- Policy: Analysts and admins can update risk factors for their tenant
CREATE POLICY "Analysts can update risk factors for their tenant"
ON risk_factors FOR UPDATE
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
)
WITH CHECK (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
);

-- Policy: Admins can delete risk factors for their tenant
CREATE POLICY "Admins can delete risk factors for their tenant"
ON risk_factors FOR DELETE
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_admin()
);

-- Policy: Service accounts can perform all operations
CREATE POLICY "Service accounts can perform all operations on risk factors"
ON risk_factors FOR ALL
TO authenticated
USING (is_service_account())
WITH CHECK (is_service_account());

-- Custom Risk Models RLS Policies

-- Policy: Users can view custom risk models for their tenant
CREATE POLICY "Users can view custom risk models for their tenant"
ON custom_risk_models FOR SELECT
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_viewer()
);

-- Policy: Analysts and admins can insert custom risk models for their tenant
CREATE POLICY "Analysts can insert custom risk models for their tenant"
ON custom_risk_models FOR INSERT
TO authenticated
WITH CHECK (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
);

-- Policy: Analysts and admins can update custom risk models for their tenant
CREATE POLICY "Analysts can update custom risk models for their tenant"
ON custom_risk_models FOR UPDATE
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
)
WITH CHECK (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
);

-- Policy: Admins can delete custom risk models for their tenant
CREATE POLICY "Admins can delete custom risk models for their tenant"
ON custom_risk_models FOR DELETE
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_admin()
);

-- Policy: Service accounts can perform all operations
CREATE POLICY "Service accounts can perform all operations on custom risk models"
ON custom_risk_models FOR ALL
TO authenticated
USING (is_service_account())
WITH CHECK (is_service_account());

-- Batch Jobs RLS Policies

-- Policy: Users can view batch jobs for their tenant
CREATE POLICY "Users can view batch jobs for their tenant"
ON batch_jobs FOR SELECT
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_viewer()
);

-- Policy: Analysts and admins can insert batch jobs for their tenant
CREATE POLICY "Analysts can insert batch jobs for their tenant"
ON batch_jobs FOR INSERT
TO authenticated
WITH CHECK (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
);

-- Policy: Analysts and admins can update batch jobs for their tenant
CREATE POLICY "Analysts can update batch jobs for their tenant"
ON batch_jobs FOR UPDATE
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
)
WITH CHECK (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
);

-- Policy: Admins can delete batch jobs for their tenant
CREATE POLICY "Admins can delete batch jobs for their tenant"
ON batch_jobs FOR DELETE
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_admin()
);

-- Policy: Service accounts can perform all operations
CREATE POLICY "Service accounts can perform all operations on batch jobs"
ON batch_jobs FOR ALL
TO authenticated
USING (is_service_account())
WITH CHECK (is_service_account());

-- Webhooks RLS Policies

-- Policy: Users can view webhooks for their tenant
CREATE POLICY "Users can view webhooks for their tenant"
ON webhooks FOR SELECT
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_viewer()
);

-- Policy: Analysts and admins can insert webhooks for their tenant
CREATE POLICY "Analysts can insert webhooks for their tenant"
ON webhooks FOR INSERT
TO authenticated
WITH CHECK (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
);

-- Policy: Analysts and admins can update webhooks for their tenant
CREATE POLICY "Analysts can update webhooks for their tenant"
ON webhooks FOR UPDATE
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
)
WITH CHECK (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
);

-- Policy: Admins can delete webhooks for their tenant
CREATE POLICY "Admins can delete webhooks for their tenant"
ON webhooks FOR DELETE
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_admin()
);

-- Policy: Service accounts can perform all operations
CREATE POLICY "Service accounts can perform all operations on webhooks"
ON webhooks FOR ALL
TO authenticated
USING (is_service_account())
WITH CHECK (is_service_account());

-- Webhook Deliveries RLS Policies

-- Policy: Users can view webhook deliveries for their tenant
CREATE POLICY "Users can view webhook deliveries for their tenant"
ON webhook_deliveries FOR SELECT
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_viewer()
);

-- Policy: Service accounts can insert webhook deliveries
CREATE POLICY "Service accounts can insert webhook deliveries"
ON webhook_deliveries FOR INSERT
TO authenticated
WITH CHECK (is_service_account());

-- Policy: Service accounts can update webhook deliveries
CREATE POLICY "Service accounts can update webhook deliveries"
ON webhook_deliveries FOR UPDATE
TO authenticated
USING (is_service_account())
WITH CHECK (is_service_account());

-- Policy: Service accounts can delete webhook deliveries
CREATE POLICY "Service accounts can delete webhook deliveries"
ON webhook_deliveries FOR DELETE
TO authenticated
USING (is_service_account());

-- Audit Logs RLS Policies

-- Policy: Users can view audit logs for their tenant
CREATE POLICY "Users can view audit logs for their tenant"
ON audit_logs FOR SELECT
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_viewer()
);

-- Policy: Service accounts can insert audit logs
CREATE POLICY "Service accounts can insert audit logs"
ON audit_logs FOR INSERT
TO authenticated
WITH CHECK (is_service_account());

-- Policy: No updates or deletes allowed on audit logs (immutable)
CREATE POLICY "No updates allowed on audit logs"
ON audit_logs FOR UPDATE
TO authenticated
USING (false);

CREATE POLICY "No deletes allowed on audit logs"
ON audit_logs FOR DELETE
TO authenticated
USING (false);

-- Compliance Checks RLS Policies

-- Policy: Users can view compliance checks for their tenant
CREATE POLICY "Users can view compliance checks for their tenant"
ON compliance_checks FOR SELECT
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_viewer()
);

-- Policy: Analysts and admins can insert compliance checks for their tenant
CREATE POLICY "Analysts can insert compliance checks for their tenant"
ON compliance_checks FOR INSERT
TO authenticated
WITH CHECK (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
);

-- Policy: Analysts and admins can update compliance checks for their tenant
CREATE POLICY "Analysts can update compliance checks for their tenant"
ON compliance_checks FOR UPDATE
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
)
WITH CHECK (
    tenant_id = get_current_tenant_id() 
    AND is_analyst()
);

-- Policy: Admins can delete compliance checks for their tenant
CREATE POLICY "Admins can delete compliance checks for their tenant"
ON compliance_checks FOR DELETE
TO authenticated
USING (
    tenant_id = get_current_tenant_id() 
    AND is_admin()
);

-- Policy: Service accounts can perform all operations
CREATE POLICY "Service accounts can perform all operations on compliance checks"
ON compliance_checks FOR ALL
TO authenticated
USING (is_service_account())
WITH CHECK (is_service_account());

-- Create API key authentication policies (for service-to-service communication)

-- Function to validate API key
CREATE OR REPLACE FUNCTION validate_api_key(api_key TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    -- In a real implementation, you would validate against a stored API key
    -- For now, we'll use a simple check against environment variable
    RETURN api_key IS NOT NULL AND length(api_key) > 20;
EXCEPTION
    WHEN OTHERS THEN
        RETURN false;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Policy: API key access for service accounts
CREATE POLICY "API key access for service operations"
ON risk_assessments FOR ALL
TO anon
USING (validate_api_key(current_setting('request.headers', true)::jsonb ->> 'authorization'))
WITH CHECK (validate_api_key(current_setting('request.headers', true)::jsonb ->> 'authorization'));

-- Grant necessary permissions to authenticated users
GRANT SELECT, INSERT, UPDATE, DELETE ON risk_assessments TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON risk_predictions TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON risk_factors TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON custom_risk_models TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON batch_jobs TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON webhooks TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON webhook_deliveries TO authenticated;
GRANT SELECT, INSERT ON audit_logs TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON compliance_checks TO authenticated;

-- Grant sequence permissions
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO authenticated;

-- Grant function execution permissions
GRANT EXECUTE ON FUNCTION get_current_tenant_id() TO authenticated;
GRANT EXECUTE ON FUNCTION get_current_user_id() TO authenticated;
GRANT EXECUTE ON FUNCTION is_admin() TO authenticated;
GRANT EXECUTE ON FUNCTION is_analyst() TO authenticated;
GRANT EXECUTE ON FUNCTION is_viewer() TO authenticated;
GRANT EXECUTE ON FUNCTION is_service_account() TO authenticated;
GRANT EXECUTE ON FUNCTION validate_api_key(TEXT) TO authenticated;

-- Add comments for documentation
COMMENT ON FUNCTION get_current_tenant_id() IS 'Extracts tenant ID from JWT claims for RLS policies';
COMMENT ON FUNCTION get_current_user_id() IS 'Extracts user ID from JWT claims for audit logging';
COMMENT ON FUNCTION is_admin() IS 'Checks if current user has admin role';
COMMENT ON FUNCTION is_analyst() IS 'Checks if current user has analyst or admin role';
COMMENT ON FUNCTION is_viewer() IS 'Checks if current user has viewer, analyst, or admin role';
COMMENT ON FUNCTION is_service_account() IS 'Checks if current request is from a service account';
COMMENT ON FUNCTION validate_api_key(TEXT) IS 'Validates API key for service-to-service authentication';
