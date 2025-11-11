# MutationObserver Infinite Loop Fix Plan

## Problem
The MutationObserver in `merchant-details.html` is causing an infinite loop:

1. Observer detects tab content → calls `setupRiskIndicatorsButton()`
2. `setupButtonHandler()` clones and replaces button (line 2528-2529)
3. Button replacement triggers MutationObserver again
4. Loop repeats indefinitely

## Root Cause
- `setupButtonHandler()` uses `cloneNode()` and `replaceChild()` which triggers DOM mutations
- MutationObserver watches `document.body` with `subtree: true`, catching all mutations
- No flag to prevent re-initialization after successful setup
- Observer never disconnects after successful initialization

## Solution

### 1. Add Initialization Flags
- Track if Risk Indicators button is already set up
- Track if Risk Assessment button is already set up
- Prevent re-initialization if already done

### 2. Disconnect Observer After Success
- Disconnect MutationObserver after both buttons are found and set up
- Only reconnect if buttons are actually missing (not just DOM changes)

### 3. Improve Button Handler Setup
- Check if button already has handler before cloning
- Use `addEventListener` with `once: true` or check for existing handler
- Avoid cloning/replacing if handler already exists

### 4. Add Debouncing
- Debounce MutationObserver callbacks to prevent rapid re-triggers
- Only process mutations after a short delay

## Implementation Steps

1. Add flags: `window.riskIndicatorsButtonSetup = false` and `window.riskAssessmentButtonSetup = false`
2. Check flags before calling setup functions
3. Set flags to `true` after successful setup
4. Disconnect observer after both buttons are set up
5. Modify `setupButtonHandler` to check for existing handler before cloning
6. Add debouncing to MutationObserver callback

