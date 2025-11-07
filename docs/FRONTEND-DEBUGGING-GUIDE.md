# Frontend Debugging Guide - Data Loading Issues

**Date**: January 2025  
**Issue**: UI not updating and data not populating

---

## Problem Summary

User reported:
1. Nothing seems to have changed in the UI
2. None of the data is populating within the cards of the merchant details or business analytics tabs

---

## Root Causes Identified

### 1. Data Service Initialization Issue
**Problem**: `this.dataService` was only initialized if `this.sharedRiskService` was NOT available. If shared service existed but failed, `this.dataService` was undefined.

**Fix**: Always initialize `this.dataService` as a fallback, regardless of shared service availability.

### 2. Session Storage Dependency
**Problem**: Merchant Details tab only loaded data from `sessionStorage`. If sessionStorage was empty, no data would load.

**Fix**: Added API fallback - if sessionStorage is empty, try to load from API using merchantId.

### 3. Missing Error Handling
**Problem**: Errors were silently failing without clear feedback.

**Fix**: Added comprehensive logging and error messages.

### 4. Tab Not Auto-Loading
**Problem**: Risk Indicators tab only loaded when clicked, not on page load.

**Fix**: Added auto-load logic when tab is active on page load.

---

## Fixes Applied

### 1. Data Service Initialization Fix

**Before**:
```javascript
// Fallback to existing data service if shared not available
if (!this.sharedRiskService) {
    this.dataService = new RiskIndicatorsDataService();
}
```

**After**:
```javascript
// Always initialize fallback data service (needed even if shared service exists)
this.dataService = new RiskIndicatorsDataService();
```

### 2. Merchant Data Loading Fix

**Before**:
```javascript
loadMerchantData() {
    // Only loads from sessionStorage
    const merchantDataStr = sessionStorage.getItem('merchantData');
    if (merchantDataStr) {
        this.merchantData = JSON.parse(merchantDataStr);
    }
}
```

**After**:
```javascript
async loadMerchantData() {
    const merchantId = getMerchantId();
    
    // Try sessionStorage first
    const merchantDataStr = sessionStorage.getItem('merchantData');
    if (merchantDataStr) {
        this.merchantData = JSON.parse(merchantDataStr);
    } else if (merchantId) {
        // Fallback to API
        const response = await fetch(endpoints.merchantById(merchantId));
        if (response.ok) {
            this.merchantData = await response.json();
            sessionStorage.setItem('merchantData', JSON.stringify(this.merchantData));
        }
    }
}
```

### 3. Enhanced Logging

Added comprehensive console logging:
- `📊 Loading risk data for merchant: {id}`
- `✅ Using shared risk service` or `📦 Using fallback data service`
- `✅ Risk data loaded: {data}`
- `✅ Risk Indicators tab fully loaded and rendered`
- Error logging with stack traces

### 4. Auto-Load Feature

Added logic to auto-load Risk Indicators tab if:
- Merchant ID is available
- Risk Indicators tab is active on page load

---

## Debugging Steps

### 1. Check Browser Console

Open browser DevTools (F12) and check the Console tab for:
- ✅ Success messages (green checkmarks)
- ⚠️ Warnings (yellow)
- ❌ Errors (red)

### 2. Verify Merchant ID

Check if merchant ID is available:
```javascript
// In browser console
getMerchantId()
```

Should return a merchant ID string, not `null` or `undefined`.

### 3. Check Session Storage

```javascript
// In browser console
sessionStorage.getItem('merchantData')
sessionStorage.getItem('merchantApiResults')
```

If empty, the API fallback should load data.

### 4. Verify API Endpoints

Check if API calls are being made:
- Open Network tab in DevTools
- Look for requests to:
  - `/api/v1/merchants/{id}`
  - `/api/v1/risk/assess`
  - `/api/v1/risk/benchmarks`
  - `/api/v1/risk/predictions/{id}`

### 5. Check Component Initialization

```javascript
// In browser console
typeof MerchantRiskIndicatorsTab
typeof RiskIndicatorsDataService
typeof APIConfig
typeof RealDataIntegration
```

All should return `"function"` or `"object"`, not `"undefined"`.

---

## Common Issues and Solutions

### Issue: "MerchantRiskIndicatorsTab class not found"
**Solution**: Check that `merchant-risk-indicators-tab.js` is loaded before initialization script.

### Issue: "APIConfig not available"
**Solution**: Check that `api-config.js` is loaded before other scripts.

### Issue: "No merchant ID found"
**Solution**: 
- Check URL for `?merchantId=...` or `?id=...`
- Check sessionStorage for merchant data
- Check if merchant data is set in `window.merchantDetails.merchantData`

### Issue: "Failed to load from API"
**Solution**:
- Check API Gateway URL in `api-config.js`
- Verify API Gateway is accessible
- Check CORS settings
- Verify merchant ID is valid

### Issue: "Shared services not available"
**Solution**: This is expected if modules fail to load. The fallback data service will be used.

---

## Testing Checklist

- [ ] Open merchant details page with merchant ID in URL
- [ ] Check browser console for initialization messages
- [ ] Verify merchant data loads (either from sessionStorage or API)
- [ ] Click on Risk Indicators tab
- [ ] Verify risk data loads
- [ ] Check Network tab for API calls
- [ ] Verify data appears in UI cards

---

## Next Steps

1. **Test the fixes**: Open the merchant details page and check console
2. **Verify data loading**: Check if merchant data and risk data load correctly
3. **Check UI updates**: Verify cards populate with data
4. **Report any errors**: Share console errors if issues persist

---

## Files Modified

- `web/js/components/merchant-risk-indicators-tab.js` - Fixed dataService initialization, added logging
- `web/merchant-details.html` - Added API fallback, improved error handling, auto-load feature

---

**Status**: ✅ Fixes applied and committed

