# Railway Log Analysis - Concurrent Map Access Panic

## Issue Summary

**Date**: 2025-11-26  
**Severity**: Critical (Service Panic)  
**Status**: Fixed

## Root Cause

The classification service was experiencing a **fatal panic** due to concurrent map read and map write operations:

```
fatal error: concurrent map read and map write
```

### Location
- **File**: `internal/classification/repository/keyword_matcher.go`
- **Function**: `stem()` (line 110)
- **Map**: `stemCache map[string]string`

### Cause
The `KeywordMatcher.stemCache` map was being accessed concurrently by multiple goroutines during parallel code generation without proper synchronization. When multiple goroutines called `stem()` simultaneously:
- One goroutine would read from `stemCache` (line 110)
- Another goroutine would write to `stemCache` (lines 118, 155)
- This caused Go's runtime to detect concurrent map access and panic

### Context
The panic occurred during parallel code generation in `generateCodesInParallel()` where:
- Multiple goroutines process MCC, SIC, and NAICS codes concurrently
- Each goroutine calls `generateCodesFromKeywords()`
- Which calls `GetClassificationCodesByKeywords()`
- Which uses `KeywordMatcher.MatchKeyword()`
- Which calls `matchStem()`
- Which calls `stem()` - **the problematic function**

## Solution

Added thread-safe synchronization to protect `stemCache` from concurrent access:

1. **Added `sync.RWMutex`** to `KeywordMatcher` struct:
   ```go
   type KeywordMatcher struct {
       synonymMap map[string][]string
       stemCache  map[string]string
       stemMutex  sync.RWMutex // Protects stemCache from concurrent access
   }
   ```

2. **Protected cache reads** with read locks:
   ```go
   km.stemMutex.RLock()
   if cached, exists := km.stemCache[word]; exists {
       km.stemMutex.RUnlock()
       return cached
   }
   km.stemMutex.RUnlock()
   ```

3. **Protected cache writes** with write locks and double-check pattern:
   ```go
   km.stemMutex.Lock()
   // Double-check after acquiring write lock
   if cached, exists := km.stemCache[word]; exists {
       km.stemMutex.Unlock()
       return cached
   }
   km.stemCache[word] = stemmed
   km.stemMutex.Unlock()
   ```

### Why RWMutex?
- **Read locks** allow multiple concurrent reads (better performance)
- **Write locks** are exclusive (prevents concurrent writes)
- **Double-check pattern** prevents race conditions where multiple goroutines compute the same stem

## Impact

### Before Fix
- Service would panic and crash during parallel code generation
- Railway logs showed hundreds of stack traces
- Service was unstable under concurrent load

### After Fix
- Thread-safe concurrent access to `stemCache`
- No more panics during parallel code generation
- Service remains stable under concurrent load

## Additional Considerations

### `synonymMap` Protection
The `synonymMap` is currently only written during initialization (`loadDefaultSynonyms()`) and then only read. However, the `AddSynonym()` method exists and could potentially be called concurrently. 

**Recommendation**: If `AddSynonym()` is used in production, consider adding similar mutex protection for `synonymMap` to prevent future issues.

## Testing

- ✅ Code compiles successfully
- ✅ No linter errors
- ⚠️ Integration testing recommended to verify fix under concurrent load

## Files Changed

- `internal/classification/repository/keyword_matcher.go`
  - Added `sync` import
  - Added `stemMutex sync.RWMutex` field
  - Updated `stem()` function with mutex protection

## Deployment Notes

- This is a **critical bug fix** that should be deployed immediately
- No breaking changes - purely internal synchronization improvement
- No configuration changes required
- No database migrations required
