# Merchant Service Build Fix

**Date:** November 23, 2025  
**Issue:** Build failure due to unused variable

## Problem

The merchant-service build failed with the following error:
```
internal/handlers/merchant.go:1218:2: declared and not used: ctx
```

## Root Cause

When updating `HandleMerchantStatistics` to return the correct schema format, the `ctx` variable was declared but no longer used after removing the Supabase database call.

## Fix

Removed the unused `ctx` variable declaration:
```go
// Before
func (h *MerchantHandler) HandleMerchantStatistics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	ctx := r.Context()  // ❌ Unused variable

// After
func (h *MerchantHandler) HandleMerchantStatistics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()  // ✅ No unused variables
```

## Verification

✅ Build now succeeds:
```bash
go build -o /tmp/merchant-service-test ./cmd/main.go
# No errors
```

## Status

✅ **FIXED** - Code committed and pushed  
⏳ **DEPLOYING** - Railway auto-deploy will trigger deployment

---

**Commit:** Fix build error: remove unused ctx variable in HandleMerchantStatistics

