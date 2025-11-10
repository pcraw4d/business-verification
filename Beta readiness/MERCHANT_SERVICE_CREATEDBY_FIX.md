# Merchant Service CreatedBy Field Fix

**Date**: 2025-11-10  
**Status**: ✅ Completed

---

## Summary

Implemented extraction of user ID from authentication context/headers for the `CreatedBy` field in merchant creation, replacing the hardcoded "system" value.

---

## Issue

- **Location**: `services/merchant-service/internal/handlers/merchant.go:340`
- **Status**: TODO - Get from auth context
- **Impact**: Low - Currently hardcoded to "system"

---

## Implementation

### Changes Made

1. **Created `getUserIDFromRequest` function**
   - Extracts user ID from multiple sources
   - Priority order:
     1. Context value `user_id` (set by API Gateway auth middleware)
     2. `X-User-ID` header
     3. Fallback to "system"

2. **Updated `createMerchant` function**
   - Added `userID` parameter
   - Uses provided user ID instead of hardcoded "system"

3. **Updated `HandleCreateMerchant` function**
   - Extracts user ID before calling `createMerchant`
   - Passes user ID to `createMerchant`

---

## Code Changes

### Before
```go
CreatedBy: "system", // TODO: Get from auth context
```

### After
```go
// Extract user ID from request (headers or context)
userID := h.getUserIDFromRequest(r)

// Create merchant
merchant, err := h.createMerchant(ctx, &req, startTime, userID)

// In createMerchant:
CreatedBy: userID,
```

---

## User ID Extraction Logic

```go
func (h *MerchantHandler) getUserIDFromRequest(r *http.Request) string {
    ctx := r.Context()

    // 1. Try context (from API Gateway auth middleware)
    if userID := ctx.Value("user_id"); userID != nil {
        if id, ok := userID.(string); ok && id != "" {
            return id
        }
    }

    // 2. Try X-User-ID header
    if userID := r.Header.Get("X-User-ID"); userID != "" {
        return userID
    }

    // 3. Fallback to "system"
    return "system"
}
```

---

## Benefits

1. **Proper Attribution**: Merchants are attributed to the creating user
2. **Audit Trail**: Better tracking of who created merchants
3. **Flexibility**: Works with API Gateway or direct calls
4. **Backward Compatible**: Falls back to "system" if no auth available

---

## Testing Recommendations

1. **Test With API Gateway**: Verify user ID is extracted from context
2. **Test With Header**: Verify X-User-ID header works
3. **Test Without Auth**: Verify fallback to "system" works
4. **Test Direct Calls**: Verify service works when called directly

---

## Future Enhancements

1. **JWT Decoding**: Decode JWT token from Authorization header to extract user ID
2. **User Service Integration**: Query user service to get user details
3. **Validation**: Validate user ID exists before using it

---

## Next Steps

1. ✅ Changes committed and pushed
2. ⏳ Test user ID extraction in deployed environment
3. ⏳ Verify merchants are properly attributed
4. ⏳ Monitor logs for user ID extraction

---

**Last Updated**: 2025-11-10

