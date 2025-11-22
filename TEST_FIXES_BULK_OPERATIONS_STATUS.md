# BulkOperationsManager Test Fixes Status

**Date:** 2025-01-27  
**Status:** 10/14 passing (71%) - 4 remaining failures

---

## Fixed Tests ✅

1. ✅ should render merchant list
2. ✅ should display merchant information
3. ✅ should allow selecting individual merchants
4. ✅ should allow selecting all merchants
5. ✅ should show operation selection when merchants are selected
6. ✅ should display operation logs
7. ✅ should allow searching merchants
8. ✅ should handle empty merchant list
9. ✅ should handle loading state
10. ✅ should handle API errors

---

## Remaining Failures ❌

### 1. "should allow deselecting all merchants"
**Issue:** Checkbox filtering logic - merchant checkboxes not being correctly identified after deselect all action.

**Root Cause:** The test expects all merchant checkboxes to be unchecked after clicking "Deselect All", but the checkbox filtering logic isn't correctly identifying merchant checkboxes vs other checkboxes.

**Attempted Fixes:**
- Updated checkbox filtering to check parent div text content for merchant names
- Added validation that merchant checkboxes exist before asserting

**Next Steps:**
- Debug the actual checkbox state after deselect action
- Verify the component's `deselectAll` function is working correctly
- Consider using data-testid attributes for more reliable selection

### 2. "should allow selecting merchants by filter"
**Issue:** No checkboxes are checked after clicking "Select by Filter" button.

**Root Cause:** The `selectByFilter` function filters merchants with `status='pending' OR risk_level='high' OR risk_level='critical'`, but the test might not be waiting long enough for the selection to register, or the filter logic isn't matching the mock data correctly.

**Attempted Fixes:**
- Updated checkbox filtering to use parent div text content
- Added validation that merchant checkboxes exist

**Next Steps:**
- Verify mock merchant data has correct status/risk_level values
- Add debug logging to see which merchants are being selected
- Ensure the component's state update is synchronous

### 3. "should filter merchants by status"
**Issue:** Unable to find "Active" option in the select dropdown.

**Root Cause:** Radix UI Select renders options in a portal, and they might not be immediately available. The test is trying to find the option before it's rendered.

**Attempted Fixes:**
- Added waitFor with querySelector for select content
- Added fallback to find option by text content
- Increased timeout to 5000ms

**Next Steps:**
- Verify Radix UI Select portal rendering in test environment
- Consider using `userEvent.selectOptions` if available
- Add explicit wait for portal to be mounted

### 4. "should filter merchants by risk level"
**Issue:** Unable to find "High" option in the select dropdown.

**Root Cause:** Same as #3 - Radix UI Select portal rendering issue.

**Attempted Fixes:**
- Added waitFor with querySelector for select content
- Added fallback to find option by text content
- Increased timeout to 5000ms

**Next Steps:**
- Same as #3 - verify Radix UI Select portal rendering
- Consider using `userEvent.selectOptions` if available
- Add explicit wait for portal to be mounted

---

## Technical Notes

### Radix UI Select Portal Issues
Radix UI Select components render their options in a portal, which can cause timing issues in tests. The options might not be immediately available after clicking the trigger.

**Potential Solutions:**
1. Use `userEvent.selectOptions` if Vitest supports it
2. Add explicit wait for portal to be mounted using `document.querySelector('[role="listbox"]')`
3. Mock Radix UI Select to render options inline for tests
4. Use `waitFor` with longer timeouts and multiple fallback strategies

### Checkbox Filtering Logic
The component renders merchant checkboxes in div containers, not table rows. The test needs to correctly identify merchant checkboxes by checking if their parent div contains merchant names.

**Current Approach:**
```typescript
const merchantCheckboxes = checkboxes.filter(cb => {
  const parent = cb.closest('div');
  if (!parent) return false;
  const text = parent.textContent || '';
  return text.includes('Business 1') || text.includes('Business 2') || text.includes('Business 3');
});
```

**Potential Improvements:**
- Add `data-testid` attributes to merchant checkboxes in the component
- Use more specific selectors based on component structure
- Verify checkbox state changes are synchronous

---

## Recommendations

1. **Add data-testid attributes** to merchant checkboxes and select options for more reliable test selection
2. **Mock Radix UI Select** to render options inline for tests, avoiding portal timing issues
3. **Verify component state updates** are synchronous and testable
4. **Consider using React Testing Library's `selectOptions`** if available for select interactions

---

**Status:** ⚠️ **4 tests remaining - Portal rendering and checkbox filtering issues**

