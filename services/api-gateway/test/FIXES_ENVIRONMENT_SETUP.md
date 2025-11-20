# Fixes Environment Setup Guide

**Date:** 2025-01-27

## Overview

This guide explains how to set up the environment to test the fixes for:
1. Invalid Merchant ID Error Handling
2. Service Connectivity for Local Development

## Prerequisites

### Required Environment Variables

All services need Supabase credentials to start:

```bash
SUPABASE_URL=https://qpqhuqqmkjxsltzshfam.supabase.co
SUPABASE_ANON_KEY=<your_actual_anon_key>
SUPABASE_SERVICE_ROLE_KEY=<your_actual_service_role_key>
SUPABASE_JWT_SECRET=<your_actual_jwt_secret>
```

### Where to Get Credentials

1. **Supabase Dashboard:**
   - Go to your Supabase project
   - Settings → API
   - Copy the URL, anon key, and service role key

2. **JWT Secret:**
   - Settings → API → JWT Secret

## Setup Options

### Option 1: Use Existing .env File

If you have a `.env` file with valid credentials:

```bash
# Copy .env to railway.env (if needed)
cp .env railway.env

# Or source .env directly
source .env
export ENVIRONMENT=development
```

### Option 2: Update railway.env

Edit `railway.env` and replace placeholder values:

```bash
# Edit railway.env
nano railway.env

# Replace these lines:
SUPABASE_ANON_KEY=your_supabase_anon_key_here
SUPABASE_SERVICE_ROLE_KEY=your_supabase_service_role_key_here
SUPABASE_JWT_SECRET=your_supabase_jwt_secret_here

# With actual values:
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
SUPABASE_SERVICE_ROLE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
SUPABASE_JWT_SECRET=your-actual-jwt-secret
```

### Option 3: Set Environment Variables Directly

```bash
export SUPABASE_URL=https://qpqhuqqmkjxsltzshfam.supabase.co
export SUPABASE_ANON_KEY=<your_key>
export SUPABASE_SERVICE_ROLE_KEY=<your_key>
export SUPABASE_JWT_SECRET=<your_secret>
export ENVIRONMENT=development
```

## Testing the Fixes

### Quick Code Verification (No Services Needed)

```bash
./scripts/test-fixes-simple.sh
```

This verifies:
- ✅ Fix code is present
- ✅ Port configuration is correct
- ⚠️ Runtime testing requires services

### Full Testing (Requires Valid Credentials)

```bash
./scripts/setup-and-test-fixes.sh
```

This will:
1. Load environment variables from `railway.env`
2. Start Merchant Service (port 8083)
3. Start Risk Assessment Service (port 8082)
4. Start API Gateway (port 8080)
5. Test both fixes

### Manual Testing

If services are already running:

```bash
# Test invalid merchant ID (should return 404)
curl -v http://localhost:8080/api/v1/merchants/invalid-id-123

# Test service connectivity (check API Gateway logs)
tail -f logs/api-gateway.log | grep localhost
```

## Expected Results

### Fix 1: Invalid Merchant ID

**Before Fix:**
```bash
$ curl http://localhost:8080/api/v1/merchants/invalid-id-123
HTTP/1.1 200 OK
{"id":"invalid-id-123","name":"Sample Merchant",...}
```

**After Fix:**
```bash
$ curl http://localhost:8080/api/v1/merchants/invalid-id-123
HTTP/1.1 404 Not Found
{"error":{"code":"NOT_FOUND","message":"Merchant not found"}}
```

### Fix 2: Service Connectivity

**Before Fix:**
- API Gateway uses Railway URLs even in development

**After Fix:**
- When `ENVIRONMENT=development`, API Gateway uses localhost URLs
- Check logs: `grep "localhost" logs/api-gateway.log`

## Troubleshooting

### Services Won't Start

**Error:** `Supabase environment variables must be set`

**Solution:**
- Ensure `railway.env` has valid credentials
- Or set environment variables directly
- Or use `.env` file if it has valid credentials

### Invalid Merchant ID Still Returns 200

**Cause:** Merchant service is running old code

**Solution:**
- Restart merchant service with new code
- Run: `./scripts/setup-and-test-fixes.sh`

### API Gateway Not Using Localhost URLs

**Cause:** `ENVIRONMENT` not set to `development`

**Solution:**
```bash
export ENVIRONMENT=development
# Restart API Gateway
```

## Current Status

✅ **Code Complete:** Both fixes are implemented  
⚠️ **Testing:** Requires valid Supabase credentials  
✅ **Documentation:** Complete  

## Next Steps

1. **If you have credentials:**
   - Update `railway.env` with actual values
   - Run `./scripts/setup-and-test-fixes.sh`

2. **If you don't have credentials:**
   - Fixes are code-complete and correct
   - Can test later when credentials are available
   - Continue with implementation plan tasks

3. **If services are already running:**
   - Test fixes through running services
   - Restart services when convenient to apply fixes

