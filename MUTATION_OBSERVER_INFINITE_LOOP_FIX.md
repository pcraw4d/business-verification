# MutationObserver Infinite Loop Fix

## Problem
The MutationObserver in `merchant-details.html` was causing an infinite loop:

1. Observer detects tab content → calls `setupRiskIndicatorsButton()` and `setupRiskAssessmentButton()`
2. `setupButtonHandler()` clones and replaces button (triggers DOM mutation)
3. Button replacement triggers MutationObserver again
4. Loop repeats indefinitely, causing performance issues

## Root Cause
- `setupButtonHandler()` uses `cloneNode()` and `replaceChild()` which triggers DOM mutations
- MutationObserver watches `document.body` with `subtree: true`, catching all mutations
- No flag to prevent re-initialization after successful setup
- Observer never disconnects after successful initialization

## Solution Implemented

### 1. Added Initialization Flags
```javascript
window.riskIndicatorsButtonSetup = false;
window.riskAssessmentButtonSetup = false;
window.tabObserver = null;
```

### 2. Added Debouncing
- 500ms debounce timer to prevent rapid re-triggers
- Only processes mutations after 500ms of no changes

### 3. Check Flags Before Re-running
- MutationObserver checks if buttons are already set up before calling setup functions
- Only calls setup if flags are `false`

### 4. Disconnect Observer After Success
- Observer disconnects after both buttons are successfully set up
- Prevents further unnecessary observations

### 5. Prevent Re-setup in Button Handler
- Check for `data-risk-indicators-handler` or `data-risk-assessment-handler` attribute
- Skip setup if handler already exists
- Mark button with attribute after setup

### 6. Mark Buttons as Set Up
- Set `window.riskIndicatorsButtonSetup = true` after successful setup
- Set `window.riskAssessmentButtonSetup = true` after successful setup
- Check both flags and disconnect observer when both are `true`

## Files Modified

1. `cmd/frontend-service/static/merchant-details.html`
   - Added initialization flags
   - Modified MutationObserver with debouncing and exit conditions
   - Updated both button setup handlers

2. `services/frontend/public/merchant-details.html`
   - Applied same fixes for consistency

## Expected Results

- ✅ No infinite loop in console
- ✅ Observer disconnects after both buttons are set up
- ✅ Buttons only set up once
- ✅ Better performance (no excessive DOM queries)
- ✅ Cleaner console output

## Testing

After deployment, verify:
1. Console shows "✅ Both buttons set up - MutationObserver disconnected"
2. No repeated "Looking for Risk Indicators tab button" messages
3. Page loads normally with all tabs functional
4. No performance degradation

