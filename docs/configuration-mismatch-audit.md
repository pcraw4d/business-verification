# Configuration Mismatch Investigation - Track 7.2

## Executive Summary

Investigation of configuration mismatches reveals **some inconsistencies in environment variable naming and default values**, but **most configurations are consistent**. Key areas of concern include timeout configurations, service URL configurations, and environment variable naming inconsistencies.

**Status**: ⚠️ **MEDIUM** - Some inconsistencies found, but not critical

## Configuration Loading

### Environment Variable Loading

**Location**: `services/classification-service/internal/config/config.go:162-203`

**Loading Pattern**:
- Uses helper functions: `getEnvAsString()`, `getEnvAsInt()`, `getEnvAsBool()`, `getEnvAsDuration()`, `getEnvAsFloat()`
- Checks `os.LookupEnv()` for existence
- Falls back to default values if not set
- **Status**: ✅ Consistent and correct

**Helper Functions**:
```go
func getEnvAsString(key, defaultValue string) string
func getEnvAsInt(key string, defaultValue int) int
func getEnvAsBool(key string, defaultValue bool) bool
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration
func getEnvAsFloat(key string, defaultValue float64) float64
```

**Status**: ✅ All helper functions implemented correctly

## Configuration Inconsistencies

### 1. Environment Variable Naming ⚠️ **LOW**

**Issue**: Inconsistent environment variable names across services

**API Gateway**:
- Uses `SUPABASE_API_KEY` as fallback for `SUPABASE_ANON_KEY`
- Checks both `SUPABASE_ANON_KEY` and `SUPABASE_API_KEY`

**Classification Service**:
- Only checks `SUPABASE_ANON_KEY`
- No fallback to `SUPABASE_API_KEY`

**Merchant Service**:
- Checks both `ENVIRONMENT` and `ENV` (with `ENVIRONMENT` taking precedence)
- Other services only check `ENVIRONMENT`

**Impact**: ⚠️ **LOW** - Works but inconsistent, may cause confusion

**Recommendation**: Standardize to single variable names
- Use `SUPABASE_ANON_KEY` only (remove `SUPABASE_API_KEY` fallback)
- Use `ENVIRONMENT` only (remove `ENV` fallback)

### 2. Timeout Configuration ⚠️ **MEDIUM**

**Classification Service**:
- `ReadTimeout`: 120s (default)
- `WriteTimeout`: 120s (default)
- `RequestTimeout`: 120s (default)
- `IdleTimeout`: 60s (default)

**API Gateway**:
- `ReadTimeout`: 30s (default)
- `WriteTimeout`: 30s (default)
- HTTP Client Timeout: 30s

**Merchant Service**:
- `ReadTimeout`: 30s (default)
- `WriteTimeout`: 30s (default)
- `RequestTimeout`: 10s (default)

**Inconsistency**: ⚠️ **MEDIUM**
- Classification service has longer timeouts (120s) to accommodate long-running requests
- API Gateway and Merchant Service have shorter timeouts (30s)
- **Impact**: API Gateway may timeout before Classification Service completes

**Recommendation**: 
- Align API Gateway timeout with Classification Service timeout
- Or ensure API Gateway timeout is longer than Classification Service timeout
- **Priority**: MEDIUM (may cause timeout issues)

### 3. Service URL Configuration ⚠️ **MEDIUM**

**Classification Service → Python ML Service**:
- Environment Variable: `PYTHON_ML_SERVICE_URL`
- Expected: `https://python-ml-service-production.up.railway.app`
- **Status**: ⏳ **NEEDS VERIFICATION** - Not in `railway.json`

**Classification Service → Playwright Service**:
- Environment Variable: `PLAYWRIGHT_SERVICE_URL`
- Expected: `https://playwright-scraper-production.up.railway.app` (or similar)
- **Status**: ⏳ **NEEDS VERIFICATION** - Not in `railway.json`

**API Gateway → Classification Service**:
- Environment Variable: `CLASSIFICATION_SERVICE_URL`
- **Status**: ⏳ **NEEDS VERIFICATION** - Not in `railway.json`

**Impact**: ⚠️ **MEDIUM** - Service URLs may not be configured correctly in Railway

**Recommendation**: 
- Verify all service URLs are set in Railway dashboard
- Document expected service URLs
- Add service URL validation on startup

### 4. Default Values ⚠️ **LOW**

**Consistent Defaults**:
- ✅ Server timeouts: Most services use 30s read/write, 60s idle
- ✅ Port defaults: Different per service (8080, 8081, 8082) - **Correct**
- ✅ Logging: All use "info" level, "json" format
- ✅ Supabase configuration: All use same variable names

**Inconsistent Defaults**:
- ⚠️ Classification Service: 120s timeouts (longer for long-running requests)
- ⚠️ Request timeouts: Different values (10s, 30s, 120s)

**Impact**: ⚠️ **LOW** - Different defaults are acceptable for different service needs

**Recommendation**: Document why defaults differ (service-specific requirements)

## Service-to-Service Configuration

### API Gateway → Classification Service

**Configuration**:
- Service URL: `CLASSIFICATION_SERVICE_URL`
- HTTP Client Timeout: 30s (default)
- **Issue**: Classification Service timeout is 120s, but API Gateway timeout is 30s
- **Impact**: ⚠️ **HIGH** - API Gateway may timeout before Classification Service completes

**Recommendation**: 
- Increase API Gateway HTTP client timeout to match Classification Service timeout (120s)
- Or implement async request pattern for long-running requests

### Classification Service → Python ML Service

**Configuration**:
- Service URL: `PYTHON_ML_SERVICE_URL`
- HTTP Client Timeout: 30s (in Python ML service client)
- Classification Service Request Timeout: 120s
- **Status**: ✅ Timeout alignment looks correct (30s < 120s)

**Recommendation**: ✅ No changes needed

### Classification Service → Playwright Service

**Configuration**:
- Service URL: `PLAYWRIGHT_SERVICE_URL`
- HTTP Client Timeout: 60s (in Playwright scraper)
- Classification Service Request Timeout: 120s
- **Status**: ✅ Timeout alignment looks correct (60s < 120s)

**Recommendation**: ✅ No changes needed

### Classification Service → Supabase

**Configuration**:
- URL: `SUPABASE_URL`
- API Key: `SUPABASE_ANON_KEY`
- Service Role Key: `SUPABASE_SERVICE_ROLE_KEY`
- Health Check Timeout: 5s
- **Status**: ✅ Configuration looks correct

**Recommendation**: ✅ No changes needed

## Railway Configuration

### railway.json Analysis

**Location**: `railway.json`

**Current Configuration**:
- Only sets `ENV` and `LOG_LEVEL` for production/staging
- Does not set service URLs
- Does not set timeout configurations
- Does not set feature flags

**Missing Variables**:
- `PYTHON_ML_SERVICE_URL`
- `PLAYWRIGHT_SERVICE_URL`
- `CLASSIFICATION_SERVICE_URL` (for API Gateway)
- `REQUEST_TIMEOUT`
- Feature flags (if different from defaults)

**Impact**: ⚠️ **MEDIUM** - Variables may not be set in Railway, relying on defaults

**Recommendation**: 
- Document all required environment variables
- Verify variables are set in Railway dashboard
- Consider adding to `railway.json` for documentation

## Configuration Validation

### Current Validation

**Location**: `services/classification-service/internal/config/config.go:153-156`

**Validation**:
```go
if cfg.Supabase.URL == "" || cfg.Supabase.APIKey == "" || cfg.Supabase.ServiceRoleKey == "" {
    return nil, fmt.Errorf("Supabase environment variables must be set")
}
```

**Status**: ✅ Validates required Supabase configuration

**Missing Validation**:
- Service URLs (Python ML, Playwright)
- Timeout configurations
- Feature flags

**Recommendation**: 
- Add validation for critical service URLs
- Add validation for timeout configurations
- Log warnings for missing optional configurations

## Root Cause Analysis

### Potential Issues

1. **API Gateway Timeout Mismatch** ⚠️ **HIGH**
   - API Gateway timeout: 30s
   - Classification Service timeout: 120s
   - **Impact**: API Gateway may timeout before Classification Service completes
   - **Evidence**: Classification Service has longer timeouts for long-running requests

2. **Service URLs Not Verified** ⚠️ **MEDIUM**
   - Service URLs not in `railway.json`
   - May not be set in Railway dashboard
   - **Impact**: Services may not be able to communicate
   - **Evidence**: Track 6.1, 6.2 findings

3. **Environment Variable Naming Inconsistencies** ⚠️ **LOW**
   - Different services use different variable names
   - **Impact**: Confusion, but functionality works
   - **Evidence**: API Gateway uses `SUPABASE_API_KEY` fallback

4. **Missing Configuration Validation** ⚠️ **LOW**
   - Only validates Supabase configuration
   - Does not validate service URLs
   - **Impact**: Services may fail at runtime instead of startup
   - **Evidence**: No validation for service URLs

## Recommendations

### Immediate Actions (High Priority)

1. **Fix API Gateway Timeout**:
   - Increase API Gateway HTTP client timeout to 120s (or higher)
   - Ensure timeout is longer than Classification Service timeout
   - **Expected Impact**: Prevents timeout errors

2. **Verify Service URLs**:
   - Check Railway dashboard for all service URLs
   - Verify `PYTHON_ML_SERVICE_URL` is set
   - Verify `PLAYWRIGHT_SERVICE_URL` is set
   - Verify `CLASSIFICATION_SERVICE_URL` is set (for API Gateway)

### Medium Priority Actions

3. **Standardize Environment Variable Names**:
   - Use `SUPABASE_ANON_KEY` only (remove `SUPABASE_API_KEY` fallback)
   - Use `ENVIRONMENT` only (remove `ENV` fallback)
   - Update all services to use consistent names

4. **Add Configuration Validation**:
   - Validate critical service URLs on startup
   - Log warnings for missing optional configurations
   - Fail fast if required configurations are missing

5. **Document Configuration**:
   - Create `.env.example` files for each service
   - Document all environment variables
   - Document expected service URLs

### Low Priority Actions

6. **Document Default Values**:
   - Document why defaults differ between services
   - Explain service-specific requirements
   - Create configuration guide

## Code Locations

- **Config Loading**: `services/classification-service/internal/config/config.go:86-159`
- **Helper Functions**: `services/classification-service/internal/config/config.go:162-203`
- **Railway Config**: `railway.json`
- **Environment Variables Doc**: `RAILWAY_ENVIRONMENT_VARIABLES.md`

## Next Steps

1. ✅ **Complete Track 7.2 Investigation** - This document
2. **Verify Service URLs in Railway** - Check dashboard
3. **Fix API Gateway Timeout** - Increase to 120s
4. **Add Configuration Validation** - Validate service URLs
5. **Standardize Variable Names** - Remove inconsistencies
6. **Document Configuration** - Create `.env.example` files

## Expected Impact

After fixing issues:

1. **Timeout Errors**: Reduced with aligned timeouts
2. **Service Communication**: Improved with verified URLs
3. **Configuration Clarity**: Improved with standardization
4. **Startup Failures**: Reduced with validation

## References

- Config Implementation: `services/classification-service/internal/config/config.go`
- Railway Config: `railway.json`
- Environment Variables: `RAILWAY_ENVIRONMENT_VARIABLES.md`
- Track 6.1: `docs/python-ml-service-connectivity-audit.md`
- Track 6.2: `docs/playwright-service-connectivity-audit.md`

