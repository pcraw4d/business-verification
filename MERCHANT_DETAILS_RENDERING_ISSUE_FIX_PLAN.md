# Merchant Details Page Rendering Issue - Fix Plan

## Summary

**Status**: ✅ Service Restart Successful | ❌ Content Still Not Rendering

The Go frontend service has been successfully restarted and is now serving the correct `merchant-details.html` file. However, the page content (tabs, merchant name, main content) is still not appearing in the DOM.

## Verification Results

### ✅ Fixed Issues
1. **Wrong File Being Served**: RESOLVED
   - Service restart fixed the issue
   - Correct file is now being served: `merchant-details.html`
   - Page title is correct: "Merchant Details - KYB Platform"

2. **JavaScript Execution**: WORKING
   - Scripts are loading and executing
   - SessionStorage data is being read
   - Page title updates with merchant name

3. **Navigation Component**: WORKING
   - Correctly skipping navigation for merchant-details page
   - Console shows: "Skipping navigation for page: merchant-details"

### ❌ Remaining Issue
**Page Content Not Rendering in DOM**

**Symptoms**:
- HTML file contains correct structure (verified via curl)
- Elements exist in source: `merchantNameText`, `.max-w-7xl`, `nav[aria-label="Tabs"]`, tab buttons
- But diagnostics show these elements are NOT in the DOM:
  ```
  🔍 Main content container (.max-w-7xl): false
  🔍 Tab navigation (nav[aria-label="Tabs"]): false
  🔍 Tab buttons found: 0
  🔍 merchantNameText element: false
  🔍 Total h1 elements: 0
  ```

**Current DOM State**:
- Body has only 1-2 children (banners/warnings)
- Main content structure is missing
- Only banner components are visible

## Root Cause Analysis

The HTML structure is correct and being served, but something is preventing it from rendering in the DOM. Possible causes:

1. **JavaScript Removing Content**: Some script may be clearing the body after page load
2. **Conditional Rendering**: Content might be conditionally rendered and condition isn't met
3. **CSS Hiding Content**: Content might be hidden via CSS (display: none, visibility: hidden)
4. **Template Processing**: If using a template system, content might not be processed
5. **Timing Issue**: Content might be removed before JavaScript can find it

## Investigation Steps

### Step 1: Check Browser DevTools
- Open browser DevTools → Elements tab
- Inspect `<body>` element
- Check if main content elements exist but are hidden
- Check computed styles for display/visibility

### Step 2: Check for JavaScript Interference
- Look for scripts that modify `document.body.innerHTML`
- Check for scripts that remove/hide elements
- Verify no scripts are running before DOM is ready

### Step 3: Check HTML Structure
- Verify HTML is valid (no unclosed tags)
- Check for JavaScript errors that might prevent rendering
- Verify all required scripts are loading

### Step 4: Check CSS
- Look for CSS that might hide content
- Check for `display: none` or `visibility: hidden` on main containers
- Verify Tailwind CSS is loading correctly

## Fix Plan

### Immediate Actions
1. **Inspect Browser DOM**: Use browser DevTools to see actual DOM structure
2. **Check Console Errors**: Look for JavaScript errors that might prevent rendering
3. **Verify Script Load Order**: Ensure scripts load in correct order
4. **Check for Race Conditions**: Verify content isn't being removed before it renders

### Potential Fixes

#### Fix 1: Ensure Content Renders Before JavaScript Runs
- Move critical HTML content outside of any conditional blocks
- Ensure content is in the initial HTML, not injected by JavaScript

#### Fix 2: Fix JavaScript That Removes Content
- If any script is clearing the body, ensure it respects the skipNavigationPages logic
- Add guards to prevent content removal on merchant-details page

#### Fix 3: Fix CSS That Hides Content
- Remove any CSS that hides the main content
- Ensure Tailwind classes are applied correctly

#### Fix 4: Fix Template/Server-Side Rendering
- If using templates, ensure they're being processed
- Verify server is rendering HTML correctly

## Next Steps

1. **Use Browser DevTools** to inspect actual DOM structure
2. **Check for JavaScript errors** in console
3. **Verify HTML is valid** and complete
4. **Test with minimal JavaScript** to isolate the issue
5. **Compare working vs non-working** page structures

## Files to Review

- `cmd/frontend-service/static/merchant-details.html` - Main HTML file
- `services/frontend/public/merchant-details.html` - Source file
- `services/frontend/public/components/navigation.js` - Navigation component
- Any JavaScript that manipulates DOM after page load

## Service Status

✅ **Go Frontend Service**: Running on port 8086
✅ **File Serving**: Correct file being served
✅ **JavaScript Loading**: Scripts executing
❌ **Content Rendering**: Main content not appearing in DOM

