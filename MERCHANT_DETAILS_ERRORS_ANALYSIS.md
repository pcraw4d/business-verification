# Merchant Details Page - Error Analysis and Fixes

## Errors Found in Console Logs

### 1. **Tab Buttons Not Found** ⚠️ CRITICAL
**Error**: `❌ Risk Indicators tab button not found after all retries`
**Error**: `❌ Risk Assessment tab button not found after all retries`

**Symptoms**:
- `🔍 All tab buttons found: 0` - QuerySelector finds 0 buttons with `.tab-button` or `button[data-tab]`
- `🔍 Available buttons: [object Object],[object Object],[object Object],[object Object]` - But 4 buttons DO exist when searching all buttons
- Buttons exist in HTML but JavaScript can't find them

**Root Cause Analysis**:
- Buttons are static HTML (lines 409-440 in merchant-details.html)
- Buttons have `class="tab-button"` and `data-tab` attributes
- QuerySelector runs in setTimeout after 2 seconds
- Buttons might not be in DOM when code runs, OR selectors are wrong

**Fix Applied**:
1. ✅ Enhanced button discovery to search ALL buttons first
2. ✅ Added detailed logging for each button found (text, data-tab, classes, id)
3. ✅ Improved fallback logic to check both `data-tab` attribute and text content
4. ✅ Added logging to show how many buttons have `tab-button` class vs `data-tab` attribute

**Next Steps**:
- The enhanced logging will show exactly what buttons exist and their attributes
- This will help diagnose why `.tab-button` selector finds 0 buttons

### 2. **Tab Container Not Found** ⚠️ CRITICAL
**Error**: `❌ Tab container not found after waiting for DOM`
**Error**: `🔍 Available tab-content elements:` (empty)

**Symptoms**:
- `document.getElementById('merchant-details')` returns null
- `document.querySelectorAll('.tab-content')` returns empty array
- Multiple retry attempts fail to find the tab container

**Root Cause Analysis**:
- Tab container exists in HTML at line 447: `<div class="tab-content active" id="merchant-details">`
- Container should be in DOM when page loads
- Either container isn't loading, OR selector is wrong, OR timing issue

**Fix Applied**:
1. ✅ Added multiple selector attempts (getElementById, querySelector with class, attribute selector)
2. ✅ Enhanced error logging to show ALL elements with "merchant" in id
3. ✅ Added logging to show all tab-content elements and their properties
4. ✅ Improved retry logic with better error messages

**Next Steps**:
- Enhanced logging will show what tab-content elements DO exist
- Will help diagnose if container has different id/class than expected

### 3. **merchantNameText Element Not Found** ⚠️ MODERATE
**Error**: `⚠️ merchantNameText element not found after all retries`

**Symptoms**:
- Page title doesn't update with merchant name
- Element exists in HTML at line 389: `<h1 id="merchantNameText">Loading...</h1>`
- Retry logic (10 attempts) fails to find element

**Root Cause Analysis**:
- Element is in HTML but JavaScript can't find it
- Either timing issue, OR element is being removed/replaced, OR selector is wrong

**Fix Applied**:
1. ✅ Added retry mechanism with alternative selector fallback
2. ✅ Improved timing checks

**Status**: Needs further investigation - may be related to tab container issue

### 4. **API Non-JSON Responses** ⚠️ LOW PRIORITY (Expected)
**Errors**:
- `⚠️ API returned non-JSON response, using default data`
- `⚠️ API returned non-JSON response for features, using empty array`
- `⚠️ API returned non-JSON response for supported sources, using empty array`

**Status**: These are expected fallbacks when APIs return HTML error pages. Code handles them gracefully.

### 5. **Missing DOM Elements for Mock Data Warning** ⚠️ LOW PRIORITY (Expected)
**Errors**:
- `⚠️ Element with id "dataSourceValue" not found in DOM`
- `⚠️ Element with id "dataCountValue" not found in DOM`
- `⚠️ Element with id "lastUpdatedValue" not found in DOM`
- `⚠️ Element with id "dataQualityValue" not found in DOM`

**Status**: These elements are created dynamically by MockDataWarning component. Warnings are expected if component hasn't created them yet.

## Fixes Applied

### Enhanced Logging
1. **Button Discovery Logging**:
   - Logs all buttons found in DOM
   - Shows button attributes (data-tab, classes, id, text)
   - Compares counts of buttons with different selectors

2. **Tab Container Discovery Logging**:
   - Logs all tab-content elements found
   - Shows all elements with "merchant" in id
   - Provides detailed element properties

3. **Better Error Messages**:
   - Shows what elements DO exist when expected ones aren't found
   - Helps diagnose selector issues

### Code Improvements
1. **Improved Button Discovery**:
   - Searches ALL buttons first, then filters
   - Checks both `data-tab` attribute and text content
   - More robust fallback logic

2. **Improved Tab Container Discovery**:
   - Multiple selector attempts
   - Better error reporting

## Next Steps

1. **Test Again**: After Railway redeploys, test the flow again
2. **Review Enhanced Logs**: The new logging will show exactly what's in the DOM
3. **Diagnose Root Cause**: Use the detailed logs to understand why elements aren't being found
4. **Apply Targeted Fixes**: Once root cause is identified, apply specific fixes

## Expected Behavior After Fixes

The enhanced logging will provide:
- ✅ Detailed information about all buttons in DOM
- ✅ Information about tab-content elements
- ✅ Better understanding of why selectors aren't matching
- ✅ Data to create targeted fixes

## Files Modified

1. `services/frontend/public/merchant-details.html`
   - Enhanced button discovery logic
   - Improved tab container discovery
   - Added comprehensive error logging

