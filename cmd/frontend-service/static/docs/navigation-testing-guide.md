# Navigation and Data Flow Testing Guide

## Overview
This document provides comprehensive test cases and procedures for testing the complete navigation flow from the `add-merchant` form to the `merchant-details` page, including data persistence, error handling, and user experience validation.

**Last Updated:** December 19, 2024  
**Scope:** Frontend navigation and data flow testing

---

## Test Environment Setup

### Prerequisites
1. Frontend service running on `http://localhost:8080` (or configured port)
2. Backend API services running and accessible
3. Browser developer tools enabled (Console, Network, Application/Storage tabs)
4. Test data prepared (various merchant profiles)

### Browser Support
- Chrome/Edge (latest)
- Firefox (latest)
- Safari (latest)
- Mobile browsers (iOS Safari, Chrome Android)

---

## Test Scenarios

### 1. Happy Path: Complete Form Submission Flow

#### Test Case 1.1: Basic Form Submission
**Objective:** Verify successful form submission and navigation to merchant-details page

**Steps:**
1. Navigate to `/add-merchant.html`
2. Fill in all required fields:
   - Business Name: "Test Company Inc"
   - Industry: "Technology"
   - Street Address: "123 Main St"
   - City: "San Francisco"
   - State: "CA"
   - Postal Code: "94102"
   - Phone: "+1-555-123-4567"
   - Email: "test@example.com"
   - Website: "https://www.testcompany.com"
3. Submit the form
4. Verify navigation to `/merchant-details.html`
5. Verify merchant ID is in URL (`?merchantId=...` or `?id=...`)

**Expected Results:**
- ✅ Form submits successfully
- ✅ Navigation occurs without page refresh errors
- ✅ Merchant ID appears in URL parameters
- ✅ Merchant details page loads with correct data
- ✅ All Overview tab cards display submitted data
- ✅ Page title shows business name
- ✅ No console errors

**Validation Points:**
- Check `sessionStorage.getItem('merchantData')` contains form data
- Check `sessionStorage.getItem('merchantApiResults')` contains API response
- Verify all form fields appear correctly in Overview tab cards

---

#### Test Case 1.2: Form Submission with Optional Fields
**Objective:** Verify form submission works with minimal required fields

**Steps:**
1. Navigate to `/add-merchant.html`
2. Fill in only required fields (business name, industry, address)
3. Leave optional fields empty (phone, email, website, financial info)
4. Submit the form
5. Navigate to merchant-details page

**Expected Results:**
- ✅ Form submits successfully
- ✅ Merchant details page loads
- ✅ Required fields display correctly
- ✅ Optional fields show "Not provided" or similar placeholder
- ✅ No errors for missing optional fields

---

#### Test Case 1.3: Form Submission with All Fields
**Objective:** Verify form submission with comprehensive data

**Steps:**
1. Navigate to `/add-merchant.html`
2. Fill in all fields including:
   - All contact information
   - Financial information (revenue, employees, founded year)
   - Business description
   - Additional metadata
3. Submit the form
4. Navigate to merchant-details page

**Expected Results:**
- ✅ All data appears correctly in respective cards
- ✅ Financial card shows revenue, employees, founded year
- ✅ Contact card shows all contact information
- ✅ Data is properly formatted and displayed

---

### 2. Data Persistence Testing

#### Test Case 2.1: Session Storage Persistence
**Objective:** Verify data persists in session storage across navigation

**Steps:**
1. Submit form from add-merchant page
2. Navigate to merchant-details page
3. Open browser DevTools → Application → Session Storage
4. Verify `merchantData` and `merchantApiResults` keys exist
5. Refresh the merchant-details page
6. Verify data still loads correctly

**Expected Results:**
- ✅ Session storage contains valid JSON data
- ✅ Data persists after page refresh
- ✅ Data loads correctly on refresh
- ✅ Merchant ID is preserved

**Validation:**
```javascript
// In browser console:
JSON.parse(sessionStorage.getItem('merchantData'))
JSON.parse(sessionStorage.getItem('merchantApiResults'))
```

---

#### Test Case 2.2: Direct URL Access
**Objective:** Verify merchant-details page works with direct URL access

**Steps:**
1. Submit form and note the merchant ID from URL
2. Open new browser tab/window
3. Navigate directly to `/merchant-details.html?merchantId={id}`
4. Verify page loads correctly

**Expected Results:**
- ✅ Page loads without errors
- ✅ If session data exists, it loads from session storage
- ✅ If session data missing, attempts to fetch from API
- ✅ Falls back to mock data if API unavailable
- ✅ Merchant ID from URL is used correctly

---

#### Test Case 2.3: Multiple Merchant Navigation
**Objective:** Verify navigation between different merchants

**Steps:**
1. Submit form for Merchant A
2. Navigate to merchant-details for Merchant A
3. Navigate back to add-merchant
4. Submit form for Merchant B
5. Navigate to merchant-details for Merchant B
6. Verify correct data for Merchant B displays

**Expected Results:**
- ✅ Session storage updates with new merchant data
- ✅ Merchant B data displays correctly
- ✅ No data from Merchant A appears
- ✅ URL updates with Merchant B ID

---

### 3. Error Handling Testing

#### Test Case 3.1: Missing Session Data
**Objective:** Verify graceful handling when session data is missing

**Steps:**
1. Clear session storage: `sessionStorage.clear()`
2. Navigate to `/merchant-details.html?merchantId=test-123`
3. Verify page behavior

**Expected Results:**
- ✅ Page attempts to fetch data from API using merchant ID
- ✅ If API unavailable, falls back to mock data
- ✅ Mock data tooltips appear on cards
- ✅ No console errors or page crashes
- ✅ User sees informative messages

---

#### Test Case 3.2: Invalid Session Data
**Objective:** Verify handling of corrupted session storage data

**Steps:**
1. Set invalid JSON in session storage:
   ```javascript
   sessionStorage.setItem('merchantData', 'invalid json{')
   ```
2. Navigate to merchant-details page
3. Verify error handling

**Expected Results:**
- ✅ Error is caught and logged
- ✅ Page attempts to fetch from API if merchant ID available
- ✅ Falls back to mock data gracefully
- ✅ No page crashes

---

#### Test Case 3.3: Network Errors
**Objective:** Verify handling of API fetch failures

**Steps:**
1. Disable network in DevTools (Offline mode)
2. Clear session storage
3. Navigate to `/merchant-details.html?merchantId=test-123`
4. Verify page behavior

**Expected Results:**
- ✅ Network error is caught and logged
- ✅ Page falls back to mock data
- ✅ Mock data tooltips appear
- ✅ User can still interact with page
- ✅ Error message is user-friendly

---

#### Test Case 3.4: Malformed API Response
**Objective:** Verify handling of unexpected API response format

**Steps:**
1. Use browser extension or proxy to modify API responses
2. Return malformed JSON from `/api/v1/merchants/{id}`
3. Navigate to merchant-details page
4. Verify error handling

**Expected Results:**
- ✅ JSON parse error is caught
- ✅ Error is logged to console
- ✅ Page falls back to mock data
- ✅ No page crashes

---

### 4. Tab Navigation Testing

#### Test Case 4.1: Tab Switching
**Objective:** Verify all tabs load correctly and display unique content

**Steps:**
1. Navigate to merchant-details page
2. Click each tab: Overview, Business Analytics, Risk Assessment, Risk Indicators
3. Verify content loads for each tab
4. Verify no content duplication between tabs

**Expected Results:**
- ✅ Each tab displays unique content
- ✅ Tab switching is smooth (no flickering)
- ✅ Active tab is visually indicated
- ✅ Tab content loads correctly
- ✅ No console errors

---

#### Test Case 4.2: Responsive Tab Navigation
**Objective:** Verify tab navigation works on different screen sizes

**Steps:**
1. Test on desktop (1920x1080)
2. Test on tablet (768x1024) - verify "More" dropdown appears
3. Test on mobile (375x667) - verify horizontal scrolling works
4. Switch tabs on each device size

**Expected Results:**
- ✅ Desktop: All tabs visible
- ✅ Tablet: Overflow tabs in "More" dropdown
- ✅ Mobile: Horizontal scrolling enabled
- ✅ Tab switching works on all sizes
- ✅ Touch interactions work on mobile

---

#### Test Case 4.3: Tab Content Loading
**Objective:** Verify each tab loads its specific data

**Steps:**
1. Navigate to merchant-details page
2. Open Network tab in DevTools
3. Switch to Business Analytics tab
4. Verify API calls are made (if applicable)
5. Switch to Risk Assessment tab
6. Verify risk data loads
7. Switch to Risk Indicators tab
8. Verify indicators load

**Expected Results:**
- ✅ Each tab makes appropriate API calls
- ✅ Data loads correctly for each tab
- ✅ Loading states appear during data fetch
- ✅ Error states appear if data fetch fails

---

### 5. Data Display Testing

#### Test Case 5.1: Overview Tab Cards
**Objective:** Verify all Overview tab cards display correctly

**Steps:**
1. Navigate to merchant-details page
2. Verify Overview tab is active
3. Check each card: Overview, Contact, Financial, Compliance
4. Verify data in each card matches form submission

**Expected Results:**
- ✅ Overview card shows: Business ID, Industry, Status
- ✅ Contact card shows: Address, Phone, Email, Website
- ✅ Financial card shows: Revenue, Employees, Founded Year
- ✅ Compliance card shows: KYB Status, Last Verification, Score
- ✅ All cards have loading states
- ✅ All cards have error states

---

#### Test Case 5.2: Mock Data Indicators
**Objective:** Verify mock data tooltips appear when data is mock

**Steps:**
1. Navigate to merchant-details page without session data
2. Verify mock data is displayed
3. Hover over info icons on cards
4. Verify tooltips appear with correct content

**Expected Results:**
- ✅ Mock data tooltips appear on cards with mock data
- ✅ Tooltips have correct content for each card type
- ✅ Tooltips auto-position correctly
- ✅ Tooltips are accessible (keyboard navigation)

---

#### Test Case 5.3: Real Data Display
**Objective:** Verify real data displays correctly when available

**Steps:**
1. Submit form with complete data
2. Navigate to merchant-details page
3. Verify all cards show real data
4. Verify no mock data tooltips appear

**Expected Results:**
- ✅ All cards display real data from form
- ✅ No mock data tooltips on real data cards
- ✅ Data formatting is correct
- ✅ All fields populated correctly

---

### 6. Navigation Component Testing

#### Test Case 6.1: Left Sidebar Navigation
**Objective:** Verify left sidebar navigation works correctly

**Steps:**
1. Navigate to merchant-details page
2. Verify left sidebar is visible
3. Click sidebar toggle (if available)
4. Verify sidebar collapses/expands
5. Click navigation links
6. Verify navigation works

**Expected Results:**
- ✅ Sidebar is visible on merchant-details page
- ✅ Sidebar toggle works (if implemented)
- ✅ Navigation links work correctly
- ✅ Active page is highlighted
- ✅ Layout adjusts correctly when sidebar toggles

---

#### Test Case 6.2: Fixed Footer
**Objective:** Verify fixed footer with session buttons works

**Steps:**
1. Navigate to merchant-details page
2. Scroll to bottom
3. Verify fixed footer is visible
4. Click "History" button
5. Click "End Session" button
6. Verify button functionality

**Expected Results:**
- ✅ Fixed footer is always visible at bottom
- ✅ "History" button navigates to history page
- ✅ "End Session" button clears session and navigates
- ✅ Footer is responsive on mobile
- ✅ Footer is accessible (keyboard navigation)

---

### 7. Performance Testing

#### Test Case 7.1: Page Load Performance
**Objective:** Verify page loads within acceptable time

**Steps:**
1. Open DevTools → Network tab
2. Navigate to merchant-details page
3. Measure page load time
4. Check resource loading

**Expected Results:**
- ✅ Page loads within 2-3 seconds
- ✅ No blocking resources
- ✅ Images are optimized
- ✅ JavaScript loads efficiently

---

#### Test Case 7.2: Tab Switching Performance
**Objective:** Verify tab switching is smooth and fast

**Steps:**
1. Navigate to merchant-details page
2. Rapidly switch between tabs
3. Verify no lag or jank
4. Check console for errors

**Expected Results:**
- ✅ Tab switching is instant (< 100ms)
- ✅ No visual lag or jank
- ✅ Smooth animations
- ✅ No memory leaks

---

### 8. Accessibility Testing

#### Test Case 8.1: Keyboard Navigation
**Objective:** Verify all functionality is keyboard accessible

**Steps:**
1. Navigate to merchant-details page using keyboard only
2. Tab through all interactive elements
3. Verify focus indicators are visible
4. Verify all actions can be performed with keyboard

**Expected Results:**
- ✅ All tabs are keyboard accessible
- ✅ Tab order is logical
- ✅ Focus indicators are visible
- ✅ All buttons/links are keyboard accessible
- ✅ Skip navigation link works

---

#### Test Case 8.2: Screen Reader Compatibility
**Objective:** Verify page works with screen readers

**Steps:**
1. Enable screen reader (NVDA, JAWS, VoiceOver)
2. Navigate through merchant-details page
3. Verify all content is announced correctly
4. Verify ARIA labels are present

**Expected Results:**
- ✅ All content is announced
- ✅ ARIA labels are correct
- ✅ Tab roles are announced
- ✅ Button purposes are clear
- ✅ Form labels are associated

---

### 9. Cross-Browser Testing

#### Test Case 9.1: Browser Compatibility
**Objective:** Verify functionality works across browsers

**Steps:**
1. Test in Chrome/Edge
2. Test in Firefox
3. Test in Safari
4. Verify consistent behavior

**Expected Results:**
- ✅ All features work in all browsers
- ✅ Visual appearance is consistent
- ✅ No browser-specific errors
- ✅ Performance is acceptable in all browsers

---

## Test Checklist

### Pre-Testing Checklist
- [ ] Frontend service is running
- [ ] Backend API is accessible
- [ ] Browser DevTools are open
- [ ] Network tab is monitoring requests
- [ ] Console is clear of errors

### Post-Testing Checklist
- [ ] All test cases executed
- [ ] All issues documented
- [ ] Screenshots captured for failures
- [ ] Console logs saved
- [ ] Network requests reviewed

---

## Common Issues and Solutions

### Issue: Data not persisting in session storage
**Solution:** Check browser settings allow session storage, verify no extensions blocking storage

### Issue: Navigation not working
**Solution:** Check console for JavaScript errors, verify event listeners are attached

### Issue: Tabs not switching
**Solution:** Check tab click handlers, verify tab IDs match, check for CSS conflicts

### Issue: Mock data not showing
**Solution:** Verify data source tracking, check tooltip initialization, verify card elements exist

---

## Reporting Test Results

### Test Result Template
```
Test Case: [ID] - [Name]
Status: ✅ Pass / ❌ Fail / ⚠️ Partial
Browser: [Browser and version]
Date: [Date]
Notes: [Any observations]
Screenshots: [Links if applicable]
Console Errors: [List any errors]
```

---

## Automated Testing Recommendations

### Unit Tests
- Test data validation functions
- Test data normalization functions
- Test merchant ID extraction
- Test session storage operations

### Integration Tests
- Test form submission → navigation flow
- Test session storage → page load flow
- Test API fetch → data display flow

### E2E Tests (using Playwright/Cypress)
- Complete form submission flow
- Tab navigation flow
- Error handling scenarios
- Responsive design verification

---

**Document Version:** 1.0.0  
**Last Updated:** December 19, 2024  
**Next Review:** March 19, 2025

