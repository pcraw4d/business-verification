# Routing Update - Default to New UI

**Date**: 2025-01-17  
**Status**: ✅ **COMPLETE**

## Summary

Updated routing logic to default to the new UI instead of requiring explicit feature flags. Legacy UI is now only used as a fallback or when explicitly enabled.

## Changes Made

### 1. Updated Route Configuration (`cmd/frontend-service/routing.go`)

#### Before
- Required `USE_NEW_UI=true` or `NEXT_PUBLIC_USE_NEW_UI=true` to enable new UI
- Defaulted to legacy UI if flag not set

#### After
- **New UI is now the default**
- Only uses legacy UI if:
  - `USE_LEGACY_UI=true` is explicitly set, OR
  - Next.js page doesn't exist for the route
- Backward compatible: `USE_NEW_UI=false` also disables new UI

### 2. Updated Routing Logic

The `serveRoute` function now:
1. **Tries Next.js first** (if new UI is enabled, which is default)
2. **Falls back to legacy** only if:
   - Next.js page doesn't exist, OR
   - Legacy UI is explicitly enabled

### 3. Updated Merchant Details Handler

Updated `handleMerchantDetailsRoute` to use the routing config for consistency.

## Environment Variables

### New Behavior

| Variable | Value | Behavior |
|----------|-------|----------|
| (none) | - | **Default: New UI** ✅ |
| `USE_LEGACY_UI` | `true` | Use legacy UI |
| `USE_NEW_UI` | `false` | Use legacy UI (backward compat) |
| `NEXT_PUBLIC_USE_NEW_UI` | `false` | Use legacy UI (backward compat) |
| `USE_NEW_UI` | `true` | Use new UI (explicit) |
| `NEXT_PUBLIC_USE_NEW_UI` | `true` | Use new UI (explicit) |

### Migration Path

**For Railway/Production:**
- **No action needed** - New UI is now default
- To explicitly use legacy: Set `USE_LEGACY_UI=true`
- To explicitly use new UI: Set `USE_NEW_UI=true` (optional, already default)

**For Development:**
- New UI will be used by default
- Set `USE_LEGACY_UI=true` if you need to test legacy UI

## Benefits

1. ✅ **Simpler deployment** - No need to set feature flags
2. ✅ **Better defaults** - New UI is the standard
3. ✅ **Backward compatible** - Legacy UI still available if needed
4. ✅ **Automatic fallback** - Falls back to legacy if Next.js page missing

## Verification

### Test New UI (Default)
```bash
# No environment variables needed
# Should serve new UI by default
curl http://localhost:8086/
```

### Test Legacy UI (Explicit)
```bash
# Set legacy flag
export USE_LEGACY_UI=true
# Should serve legacy UI
curl http://localhost:8086/
```

### Test Fallback
```bash
# Request a route that doesn't exist in Next.js
# Should fall back to legacy if available
curl http://localhost:8086/non-existent-route
```

## Files Modified

- `cmd/frontend-service/routing.go` - Updated default behavior
- `cmd/frontend-service/main.go` - Added strings import, updated merchant details handler

## Next Steps

1. ✅ **Routing updated** - Defaults to new UI
2. ⏳ **Deploy to Railway** - Test in production
3. ⏳ **Monitor** - Watch for any issues
4. ⏳ **Phase 4** - Proceed with legacy UI removal when ready

## Rollback

If issues occur, you can:
1. Set `USE_LEGACY_UI=true` to revert to legacy UI
2. Or set `USE_NEW_UI=false` (backward compatible)

---

**Status**: ✅ **COMPLETE - NEW UI IS NOW DEFAULT**

