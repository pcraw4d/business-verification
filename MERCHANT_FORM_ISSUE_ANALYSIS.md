# Merchant Form Submission Issue Analysis

## Problem
After clicking "Verify Merchant" on the add merchant form, the merchant details page is not being populated/displayed.

## Code Flow Analysis

### 1. Form Submission Flow
- Form has submit button with text "Verify Merchant"
- `handleSubmit()` is called on form submission
- Form validation runs first
- If valid, `processMerchantVerification()` is called
- This function:
  1. Stores data in sessionStorage
  2. Makes 3 API calls in parallel using `Promise.allSettled`
  3. Stores API results in sessionStorage
  4. Redirects to `/merchant-details` after 100ms timeout

### 2. Potential Issues Identified

#### Issue 1: Business Intelligence API Timeout
- The `callBusinessIntelligenceAPI` has a 30-second timeout
- If the API is slow or fails, it could delay the redirect
- However, `Promise.allSettled` should handle this gracefully

#### Issue 2: Error Handling
- The code has try-catch blocks, but errors might be swallowed
- If an error occurs in `processMerchantVerification`, it should still redirect
- But the error might prevent sessionStorage from being set

#### Issue 3: Redirect Timing
- Redirect happens after 100ms timeout
- This should be sufficient, but if there's a JavaScript error, it might not execute

#### Issue 4: SessionStorage Not Set
- If `processMerchantVerification` throws an error before setting sessionStorage
- The merchant details page won't have data to display

#### Issue 5: Merchant Details Page Not Loading Data
- The merchant details page loads data from sessionStorage
- If sessionStorage is empty, it shows warnings but doesn't populate
- The page should still display, just without data

## Recommended Fixes

### Fix 1: Ensure SessionStorage is Set Before API Calls
Move sessionStorage.setItem to happen immediately, before API calls.

### Fix 2: Add Fallback Redirect
Add a fallback redirect that happens even if everything fails.

### Fix 3: Improve Error Handling
Wrap the entire process in better error handling to ensure redirect always happens.

### Fix 4: Add Immediate Redirect Option
Consider redirecting immediately after storing form data, then loading API results on the details page.

### Fix 5: Add More Logging
Add comprehensive logging to track exactly where the flow breaks.

