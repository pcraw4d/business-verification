# CSS Diagnostic Testing Instructions

**Date**: 2025-01-11  
**Status**: 🔍 **READY FOR TESTING**

## Overview

A comprehensive CSS diagnostic function (`window.diagnoseCSSStyles()`) has been added to investigate why Business Analytics content remains visible even when its tab is hidden.

## How to Test

1. **Navigate to the merchant details page** on Railway:
   ```
   https://frontend-service-production-b225.up.railway.app/merchant-details.html
   ```

2. **Open the browser console** (F12 or right-click → Inspect → Console)

3. **Click the Overview tab** to switch away from Business Analytics

4. **Run the CSS diagnostic function** in the console:
   ```javascript
   window.diagnoseCSSStyles()
   ```

5. **Review the output** which will show:
   - Computed styles for `#business-analytics` tab
   - Computed styles for `#coreResults` element
   - Inline styles on both elements
   - Parent chain styles (up to 10 levels)
   - CSS rules that might override `display: none`
   - Viewport visibility using `getBoundingClientRect()`
   - Comparison with Overview tab styles

## What to Look For

The diagnostic will help identify:

1. **Position Properties**: Check if `position: absolute` or `position: fixed` is making content visible
2. **CSS Overrides**: Look for CSS rules with higher specificity than `display: none !important`
3. **Z-Index Issues**: Check if high z-index values are causing content to appear above other elements
4. **Opacity/Transform**: Check if opacity or transform properties are affecting visibility
5. **Viewport Visibility**: Verify if content is actually in the viewport despite being hidden

## Expected Output Format

The diagnostic will output detailed information in the console, including:
- Computed styles object with all CSS properties
- Inline styles object
- Parent chain traversal showing styles at each level
- CSS rules that match Business Analytics selectors
- Viewport visibility calculations

## Next Steps

After running the diagnostic:
1. Copy the console output
2. Analyze the computed styles to identify the root cause
3. Apply fixes based on the findings

