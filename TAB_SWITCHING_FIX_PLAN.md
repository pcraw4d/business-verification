# Tab Switching Fix Plan

## Problem

Multiple tabs on the merchant details page are displaying the same content (Business Analytics content) instead of their own unique content. When clicking on "Merchant Detail", "Overview", "Contact", "Financial", and "Compliance" tabs, they all show the Business Analytics content.

## Root Cause Analysis

### Current Implementation

1. **CSS**: `.tab-content` has `display: none` by default, and `.tab-content.active` has `display: block` - this is correct.

2. **Tab Structure**: All tabs are correctly structured with proper IDs:
   - `merchant-details` (has `active` class by default)
   - `business-analytics`
   - `risk-assessment`
   - `risk-indicators`
   - `overview`
   - `contact`
   - `financial`
   - `compliance`

3. **Tab Switching Logic**: The `switchTab()` method:
   - Removes `active` class from all tabs and buttons
   - Adds `active` class to selected tab and button
   - Calls special handlers for certain tabs (risk-indicators, risk-assessment, overview, contact, financial, compliance)

### Hypothesis

The issue is likely that:
1. The Business Analytics tab content is somehow always visible (maybe it's being shown by JavaScript after tab switching)
2. The tab switching logic isn't properly hiding the Business Analytics tab when other tabs are clicked
3. There might be JavaScript that's showing the Business Analytics content regardless of which tab is active

## Investigation Steps

1. **Check if Business Analytics tab is being forced visible**
   - Search for any JavaScript that sets `display: block` or adds `active` class to `business-analytics` tab
   - Check if there's any code that shows Business Analytics content on page load or after tab switching

2. **Verify tab switching is being called**
   - Add console logging to `switchTab()` method to verify it's being called with correct `tabId`
   - Verify that `tabElement` and `buttonElement` are being found correctly

3. **Check for CSS conflicts**
   - Verify no other CSS rules are overriding the `.tab-content` display rules
   - Check if inline styles are being applied that override the CSS

4. **Check if Business Analytics content is outside tab container**
   - Verify that Business Analytics content is actually inside the `#business-analytics` tab content container
   - Check if there's duplicate content being rendered outside the tab containers

## Fix Strategy

### Option 1: Ensure Proper Tab Hiding (Recommended)

Modify the `switchTab()` method to explicitly hide all tabs before showing the selected one:

```javascript
switchTab(tabId) {
    // Explicitly hide ALL tabs first
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.remove('active');
        tab.style.display = 'none'; // Force hide
    });
    
    // Remove active class from all buttons
    document.querySelectorAll('.tab-button').forEach(button => {
        button.classList.remove('active');
    });

    // Show only the selected tab
    const tabElement = document.getElementById(tabId);
    const buttonElement = document.querySelector(`[data-tab="${tabId}"]`);
    
    if (tabElement) {
        tabElement.classList.add('active');
        tabElement.style.display = 'block'; // Force show
    }
    if (buttonElement) {
        buttonElement.classList.add('active');
    }
    
    // Handle special tab initializations
    // ... rest of the code
}
```

### Option 2: Add Debugging and Verify Tab IDs

Add comprehensive logging to understand what's happening:

```javascript
switchTab(tabId) {
    console.log('🔄 Switching to tab:', tabId);
    
    // Log all tabs before switching
    const allTabs = document.querySelectorAll('.tab-content');
    console.log('📋 All tabs before switch:', Array.from(allTabs).map(t => ({
        id: t.id,
        hasActive: t.classList.contains('active'),
        display: window.getComputedStyle(t).display
    })));
    
    // Remove active class from all tabs and buttons
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.remove('active');
    });
    document.querySelectorAll('.tab-button').forEach(button => {
        button.classList.remove('active');
    });

    // Add active class to selected tab and button
    const tabElement = document.getElementById(tabId);
    const buttonElement = document.querySelector(`[data-tab="${tabId}"]`);
    
    console.log('🎯 Tab element found:', tabElement ? tabElement.id : 'NOT FOUND');
    console.log('🎯 Button element found:', buttonElement ? buttonElement.getAttribute('data-tab') : 'NOT FOUND');
    
    if (tabElement) {
        tabElement.classList.add('active');
        console.log('✅ Added active class to tab:', tabElement.id);
    } else {
        console.error('❌ Tab element not found for ID:', tabId);
    }
    
    if (buttonElement) {
        buttonElement.classList.add('active');
        console.log('✅ Added active class to button');
    }
    
    // Log all tabs after switching
    console.log('📋 All tabs after switch:', Array.from(document.querySelectorAll('.tab-content')).map(t => ({
        id: t.id,
        hasActive: t.classList.contains('active'),
        display: window.getComputedStyle(t).display
    })));
    
    // Handle special tab initializations
    // ... rest of the code
}
```

### Option 3: Check for JavaScript Interference

Search for any code that might be showing Business Analytics content:

```javascript
// Search for:
- document.getElementById('business-analytics')
- document.querySelector('#business-analytics')
- .style.display = 'block'
- .classList.add('active')
- business-analytics
```

## Implementation Plan

1. **Add debugging to `switchTab()` method** to understand what's happening
2. **Force hide/show tabs** using inline styles as a temporary fix
3. **Test each tab** to verify they show their own content
4. **Remove debugging** once the issue is fixed
5. **Test on Railway** to ensure fix works in production

## Files to Modify

- `cmd/frontend-service/static/merchant-details.html` - Fix `switchTab()` method
- `services/frontend/public/merchant-details.html` - Apply same fix

## Expected Outcome

- Each tab should display its own unique content
- Only one tab should be visible at a time
- Tab switching should work smoothly without showing duplicate content

