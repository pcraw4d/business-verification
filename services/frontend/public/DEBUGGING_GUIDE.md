# Comprehensive Form Flow Debugging Guide

## Overview

A comprehensive debugging system has been implemented to track the entire add-merchant form flow from submission to merchant details page population. This system provides real-time monitoring and detailed logging of all critical operations.

## Features

### 1. **Floating Debug Panel**
- **Toggle**: Press `Ctrl+Shift+\`` (backtick) or `Cmd+Shift+\`` (Mac) or click the toggle button (🔍) in bottom-right corner
- **Location**: Top-right corner of the page (when open)
- **Toggle Button**: Small green circular button (🔍) in bottom-right corner (always visible)
- **Display**: Real-time logs with color-coded categories
- **Auto-scroll**: Automatically scrolls to show latest logs

### 2. **Comprehensive Monitoring**

The debugger tracks:

#### **Form Events**
- Form element discovery
- Submit button clicks
- Form validation
- Form data collection

#### **API Calls**
- API request initiation
- API response handling
- Error tracking

#### **SessionStorage Operations**
- `setItem` operations (with value previews)
- `getItem` operations
- Current sessionStorage state

#### **Redirect Operations**
- `window.location.href` assignments
- `window.location.assign` calls
- `window.location.replace` calls
- URL changes

#### **DOM Changes**
- Merchant details tab container detection
- Merchant detail field discovery
- DOM mutation tracking

### 3. **Debug Categories**

Logs are color-coded by category:
- 🟢 **system**: System initialization and setup
- 🔵 **form**: Form-related events
- 🟡 **api**: API calls and responses
- 🟣 **storage**: SessionStorage operations
- 🔴 **redirect**: Navigation and redirects
- 🔵 **dom**: DOM element discovery and changes
- 🔴 **error**: Errors and warnings

## Usage

### Accessing the Debug Panel

1. **Keyboard Shortcut**: Press `Ctrl+Shift+\`` (backtick) or `Cmd+Shift+\`` (Mac) anywhere on the page
2. **Toggle Button**: Click the "Toggle" button in the debug panel (if visible)

### Getting a Summary

Open browser console and run:
```javascript
window.debugFormFlow.summary()
```

This will show:
- Total number of logs
- Count by category
- Total duration
- Current URL
- SessionStorage state

### Exporting Logs

To export all logs for analysis:
```javascript
window.debugFormFlow.export()
```

This downloads a JSON file with all logs, timestamps, and data.

### Accessing Raw Logs

```javascript
window.debugFormFlow.logs
```

Returns the full array of log entries.

## What to Look For

### **Form Submission Issues**
- Check for `form` category logs
- Verify `handleSubmit called` appears
- Check if validation passes or fails
- Look for form data collection logs

### **Redirect Issues**
- Check `redirect` category logs
- Verify `finalizeRedirect called` appears
- Check if `window.location.href` is set
- Look for URL change detection

### **Data Population Issues**
- Check `dom` category logs
- Look for `Tab container found` message
- Verify merchant detail fields are discovered
- Check for `populateMerchantDetails called`

### **SessionStorage Issues**
- Check `storage` category logs
- Verify `merchantData` and `merchantApiResults` are set
- Check if data is retrieved on merchant details page

## Debugging Workflow

1. **Open the add-merchant page**
   - Debug panel should initialize automatically
   - Press `Ctrl+Shift+\`` (backtick) or `Cmd+Shift+\`` (Mac) to show the panel, or click the toggle button (🔍) in the bottom-right corner

2. **Fill out and submit the form**
   - Watch the debug panel for real-time logs
   - All operations will be logged automatically

3. **After redirect to merchant details page**
   - Debug panel persists across navigation
   - Check for DOM discovery logs
   - Verify tab container is found

4. **If issues occur**
   - Export logs: `window.debugFormFlow.export()`
   - Get summary: `window.debugFormFlow.summary()`
   - Review logs by category

## Common Issues and Solutions

### **Tab Container Not Found**
- Check `dom` category logs
- Look for `Tab container found` message
- If missing, check `Document body structure` log
- Verify HTML structure matches expected IDs

### **Redirect Not Working**
- Check `redirect` category logs
- Verify `finalizeRedirect called` appears
- Check if `window.location.href` is set
- Look for URL change detection

### **Data Not Populating**
- Check `storage` category logs
- Verify data is in sessionStorage
- Check `dom` logs for field discovery
- Verify tab container is found

## Integration Points

The debugger is integrated into:
- `add-merchant.html` - Form submission tracking
- `merchant-details.html` - DOM discovery and population tracking
- `merchant-form.js` - Form handling and redirect tracking

## Technical Details

- **Performance**: Minimal overhead, uses efficient event listeners
- **Persistence**: Logs persist across page navigation (within same session)
- **Storage**: Logs kept in memory (last 100 entries)
- **Export**: Full logs can be exported as JSON

## Next Steps

After Railway redeploys:
1. Test the form submission flow
2. Use the debug panel to monitor the entire process
3. Export logs if issues occur
4. Share logs for analysis

The debugger will help identify exactly where in the flow issues occur, making it much easier to fix problems systematically.

