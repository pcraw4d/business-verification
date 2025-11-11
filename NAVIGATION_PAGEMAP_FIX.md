# Navigation PageMap Fix
## Date: November 11, 2025

---

## Summary

**Status**: ✅ **FIXED** - Missing PageMap Entries Added

Fixed missing entries in the `pageMap` object in `getCurrentPage()` method that were causing incorrect page detection and breaking active link highlighting.

---

## Bugs Fixed

### Bug 1: Missing 'risk-assessment-portfolio' Entry
- **Issue**: The `pageMap` was missing the 'risk-assessment-portfolio' entry
- **Impact**: Pages at `/risk-assessment-portfolio` or `risk-assessment-portfolio.html` would incorrectly return 'home' instead of 'risk-assessment-portfolio'
- **Fix**: Added 'risk-assessment-portfolio': 'risk-assessment-portfolio' to pageMap

### Bug 2: Missing 'analytics-insights' Entry
- **Issue**: The `pageMap` was missing the 'analytics-insights' entry
- **Impact**: Pages at `/analytics-insights` would incorrectly return 'home' instead of 'analytics-insights'
- **Fix**: Added 'analytics-insights': 'analytics-insights' to pageMap

### Bug 3: Missing Admin Page Entries
- **Issue**: The `pageMap` was missing entries for admin pages
- **Impact**: Admin pages would incorrectly return 'home' instead of their correct page identifiers
- **Fix**: Added proper path handling for admin routes:
  - `/admin` → 'admin-dashboard'
  - `/admin/models` → 'admin-models'
  - `/admin/queue` → 'admin-queue'
  - `/sessions` → 'sessions'

---

## Files Updated

1. ✅ `services/frontend/public/js/components/navigation.js`
2. ✅ `cmd/frontend-service/static/js/components/navigation.js`

**Note**: The `components/navigation.js` files (non-js directory) already had 'risk-assessment-portfolio' and don't include admin/analytics links, so they didn't need updates.

---

## Code Changes

### Enhanced Path Handling

Added special handling for multi-segment paths (like `/admin/models`):

```javascript
getCurrentPage() {
    const path = window.location.pathname;
    
    // Handle paths with multiple segments (e.g., /admin/models, /admin/queue)
    if (path.startsWith('/admin/')) {
        const segment = path.split('/').pop();
        if (segment === 'models') {
            return 'admin-models';
        } else if (segment === 'queue') {
            return 'admin-queue';
        }
    }
    
    // Handle root admin path
    if (path === '/admin' || path === '/admin/') {
        return 'admin-dashboard';
    }
    
    // Handle other paths
    const filename = path.split('/').pop().replace('.html', '');
    
    const pageMap = {
        // ... existing entries ...
        'risk-assessment-portfolio': 'risk-assessment-portfolio',
        'analytics-insights': 'analytics-insights',
        'admin': 'admin-dashboard',
        'sessions': 'sessions'
    };

    return pageMap[filename] || 'home';
}
```

---

## Expected Results

After this fix:

1. ✅ **Active link highlighting works correctly** - Navigation links highlight when on their respective pages
2. ✅ **Page-specific initialization works** - Any logic that depends on `this.currentPage` will work correctly
3. ✅ **Risk Assessment Portfolio** - Correctly identified as 'risk-assessment-portfolio'
4. ✅ **Analytics Insights** - Correctly identified as 'analytics-insights'
5. ✅ **Admin Dashboard** - Correctly identified as 'admin-dashboard'
6. ✅ **Admin Models** - Correctly identified as 'admin-models'
7. ✅ **Admin Queue** - Correctly identified as 'admin-queue'
8. ✅ **Sessions** - Correctly identified as 'sessions'

---

## Testing Checklist

After deployment, verify:

- [ ] Navigate to `/risk-assessment-portfolio` - active link highlights correctly
- [ ] Navigate to `/analytics-insights` - active link highlights correctly
- [ ] Navigate to `/admin` - active link highlights correctly
- [ ] Navigate to `/admin/models` - active link highlights correctly
- [ ] Navigate to `/admin/queue` - active link highlights correctly
- [ ] Navigate to `/sessions` - active link highlights correctly
- [ ] Check browser console - no errors related to page detection
- [ ] Verify `this.currentPage` returns correct values in all cases

---

## Related Files

- `services/frontend/public/js/components/navigation.js` - Main navigation component (js directory)
- `cmd/frontend-service/static/js/components/navigation.js` - Static navigation component
- `services/frontend/public/components/navigation.js` - Alternative navigation component (already had risk-assessment-portfolio)

---

## Notes

- The fix handles both `.html` file paths and route paths (with or without leading slash)
- Admin routes require special handling because they have multiple path segments
- The components/navigation.js files don't include admin/analytics links, so they don't need these entries
- All pageMap entries now match their corresponding `data-page` attributes in navigation links

