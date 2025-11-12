# Export Functionality Testing Guide

## Overview
This document provides comprehensive test cases and procedures for testing the export functionality across all tabs on the merchant-details page. The export system supports multiple formats (CSV, PDF, JSON, Excel) and is available on Business Analytics, Risk Assessment, and Risk Indicators tabs.

**Last Updated:** December 19, 2024  
**Scope:** Export functionality testing for merchant-details page

---

## Export System Architecture

### Components
- **ExportButton**: Reusable component for creating export buttons with format dropdowns
- **Export Service**: Handles actual export generation (client-side or API-based)
- **Export Formats**: CSV, PDF, JSON, Excel (XLSX)

### Export Locations
1. **Business Analytics Tab**: Exports business analytics data
2. **Risk Assessment Tab**: Exports risk assessment reports
3. **Risk Indicators Tab**: Exports risk indicators data

---

## Test Environment Setup

### Prerequisites
1. Frontend service running
2. Backend API services accessible
3. Merchant data loaded in merchant-details page
4. Browser developer tools enabled (Console, Network tabs)
5. File download permissions enabled in browser

### Test Data Requirements
- Merchant with complete data (all fields populated)
- Merchant with partial data (some fields missing)
- Merchant with risk assessment data
- Merchant with risk indicators

---

## Test Scenarios

### 1. Export Button UI Testing

#### Test Case 1.1: Export Button Visibility
**Objective:** Verify export buttons appear on all expected tabs

**Steps:**
1. Navigate to merchant-details page
2. Check Business Analytics tab for export button
3. Check Risk Assessment tab for export button
4. Check Risk Indicators tab for export button

**Expected Results:**
- ✅ Export button visible on Business Analytics tab
- ✅ Export button visible on Risk Assessment tab
- ✅ Export button visible on Risk Indicators tab
- ✅ Export button has download icon
- ✅ Export button is properly styled and positioned

---

#### Test Case 1.2: Export Dropdown Functionality
**Objective:** Verify export format dropdown works correctly

**Steps:**
1. Click export button on any tab
2. Verify dropdown appears
3. Verify all format options are visible (CSV, PDF, JSON, Excel)
4. Click outside dropdown
5. Verify dropdown closes

**Expected Results:**
- ✅ Dropdown appears on button click
- ✅ All format options are visible
- ✅ Format icons are correct
- ✅ Dropdown closes when clicking outside
- ✅ Dropdown closes when selecting a format
- ✅ Dropdown is properly positioned (not cut off)

---

#### Test Case 1.3: Export Button Responsive Design
**Objective:** Verify export buttons work on different screen sizes

**Steps:**
1. Test on desktop (1920x1080)
2. Test on tablet (768x1024)
3. Test on mobile (375x667)
4. Verify button visibility and functionality on each size

**Expected Results:**
- ✅ Export button visible on all screen sizes
- ✅ Dropdown is accessible on all sizes
- ✅ Dropdown doesn't overflow viewport
- ✅ Touch interactions work on mobile

---

### 2. CSV Export Testing

#### Test Case 2.1: Business Analytics CSV Export
**Objective:** Verify CSV export for Business Analytics tab

**Steps:**
1. Navigate to Business Analytics tab
2. Click export button
3. Select "Export as CSV"
4. Wait for export to complete
5. Verify file downloads
6. Open downloaded CSV file
7. Verify data accuracy

**Expected Results:**
- ✅ Export starts without errors
- ✅ Progress indicator shows (if implemented)
- ✅ File downloads successfully
- ✅ Filename is correct format: `Analytics_Export_YYYY-MM-DD.csv`
- ✅ CSV file opens correctly in Excel/Google Sheets
- ✅ All business analytics data is included
- ✅ Data matches what's displayed on page
- ✅ Column headers are correct
- ✅ No data corruption or encoding issues

**Data Validation:**
- Primary Industry
- Confidence Score
- Risk Level
- MCC/SIC/NAICS codes
- Classification methods
- Security indicators
- Quality metrics
- Business intelligence data

---

#### Test Case 2.2: Risk Assessment CSV Export
**Objective:** Verify CSV export for Risk Assessment tab

**Steps:**
1. Navigate to Risk Assessment tab
2. Click export button
3. Select "Export as CSV"
4. Verify file downloads
5. Open and validate CSV content

**Expected Results:**
- ✅ File downloads successfully
- ✅ Filename: `Risk_Export_YYYY-MM-DD.csv`
- ✅ Risk score data included
- ✅ Risk factors included
- ✅ Website risk data included
- ✅ Risk history included (if available)
- ✅ All risk metrics are accurate

---

#### Test Case 2.3: Risk Indicators CSV Export
**Objective:** Verify CSV export for Risk Indicators tab

**Steps:**
1. Navigate to Risk Indicators tab
2. Click export button
3. Select "Export as CSV"
4. Verify file downloads
5. Validate CSV content

**Expected Results:**
- ✅ File downloads successfully
- ✅ Filename: `Risk-indicators_Export_YYYY-MM-DD.csv`
- ✅ All risk indicators listed
- ✅ Indicator status included
- ✅ Severity levels included
- ✅ Timestamps included (if applicable)

---

#### Test Case 2.4: CSV Export with Missing Data
**Objective:** Verify CSV export handles missing data gracefully

**Steps:**
1. Navigate to merchant-details page with incomplete data
2. Export CSV from any tab
3. Verify export completes
4. Check CSV for missing data handling

**Expected Results:**
- ✅ Export completes without errors
- ✅ Missing fields show as empty or "N/A"
- ✅ No null/undefined values in CSV
- ✅ CSV structure remains consistent

---

### 3. PDF Export Testing

#### Test Case 3.1: Business Analytics PDF Export
**Objective:** Verify PDF export for Business Analytics tab

**Steps:**
1. Navigate to Business Analytics tab
2. Click export button
3. Select "Export as PDF"
4. Wait for PDF generation
5. Verify file downloads
6. Open PDF and validate content

**Expected Results:**
- ✅ PDF generation starts
- ✅ Progress indicator shows
- ✅ File downloads successfully
- ✅ Filename: `Analytics_Export_YYYY-MM-DD.pdf`
- ✅ PDF opens correctly in PDF viewer
- ✅ All content is readable
- ✅ Formatting is correct
- ✅ Charts/graphs are included (if applicable)
- ✅ Page breaks are appropriate
- ✅ Header/footer information is present

**Content Validation:**
- Business name and merchant ID
- Export date/timestamp
- All analytics sections included
- Proper formatting and styling
- Print-friendly layout

---

#### Test Case 3.2: Risk Assessment PDF Export
**Objective:** Verify PDF export for Risk Assessment tab

**Steps:**
1. Navigate to Risk Assessment tab
2. Click export button
3. Select "Export as PDF"
4. Verify PDF downloads
5. Validate PDF content

**Expected Results:**
- ✅ PDF downloads successfully
- ✅ Filename: `Risk_Export_YYYY-MM-DD.pdf`
- ✅ Risk score prominently displayed
- ✅ Risk breakdown included
- ✅ Charts/visualizations included
- ✅ Risk factors explained
- ✅ Professional report format

---

#### Test Case 3.3: PDF Export Performance
**Objective:** Verify PDF generation performance

**Steps:**
1. Navigate to tab with large dataset
2. Click export button
3. Select "Export as PDF"
4. Measure time to download
5. Check browser performance

**Expected Results:**
- ✅ PDF generates within 10-15 seconds for normal data
- ✅ No browser freezing or unresponsiveness
- ✅ Progress indicator updates
- ✅ User can cancel if needed (if implemented)

---

### 4. JSON Export Testing

#### Test Case 4.1: Business Analytics JSON Export
**Objective:** Verify JSON export for Business Analytics tab

**Steps:**
1. Navigate to Business Analytics tab
2. Click export button
3. Select "Export as JSON"
4. Verify file downloads
5. Open JSON file
6. Validate JSON structure

**Expected Results:**
- ✅ File downloads successfully
- ✅ Filename: `Analytics_Export_YYYY-MM-DD.json`
- ✅ JSON is valid (no syntax errors)
- ✅ JSON is properly formatted (readable)
- ✅ All data fields are included
- ✅ Data types are correct
- ✅ Nested structures are correct
- ✅ No circular references

**JSON Structure Validation:**
```json
{
  "type": "analytics",
  "merchant_id": "...",
  "timestamp": "...",
  "data": {
    "classification": {...},
    "security": {...},
    "quality": {...},
    "intelligence": {...}
  }
}
```

---

#### Test Case 4.2: JSON Export Data Completeness
**Objective:** Verify JSON export includes all expected data

**Steps:**
1. Export JSON from Business Analytics tab
2. Parse JSON file
3. Verify all expected fields are present
4. Compare with displayed data

**Expected Results:**
- ✅ All displayed data is in JSON
- ✅ Additional metadata included (timestamps, IDs)
- ✅ Data structure matches API response structure
- ✅ No data loss or truncation

---

### 5. Excel Export Testing

#### Test Case 5.1: Business Analytics Excel Export
**Objective:** Verify Excel export for Business Analytics tab

**Steps:**
1. Navigate to Business Analytics tab
2. Click export button
3. Select "Export as Excel"
4. Verify file downloads
5. Open Excel file
6. Validate content and formatting

**Expected Results:**
- ✅ File downloads successfully
- ✅ Filename: `Analytics_Export_YYYY-MM-DD.xlsx` or `.xls`
- ✅ File opens in Excel/LibreOffice
- ✅ Multiple sheets if data is categorized
- ✅ Formatting is preserved
- ✅ Charts are included (if applicable)
- ✅ Column widths are appropriate
- ✅ Data types are correct (numbers, dates, text)

---

#### Test Case 5.2: Excel Export with Charts
**Objective:** Verify Excel export includes charts and visualizations

**Steps:**
1. Export Excel from tab with charts
2. Open Excel file
3. Verify charts are present
4. Verify chart data is correct

**Expected Results:**
- ✅ Charts are embedded in Excel
- ✅ Chart data matches displayed data
- ✅ Charts are properly formatted
- ✅ Charts are on appropriate sheets

---

### 6. Error Handling Testing

#### Test Case 6.1: Export with No Data
**Objective:** Verify export handles empty data gracefully

**Steps:**
1. Navigate to tab with no data loaded
2. Click export button
3. Select any format
4. Verify error handling

**Expected Results:**
- ✅ Error message displayed: "No data available to export"
- ✅ User-friendly error message
- ✅ No console errors
- ✅ Export button remains functional
- ✅ User can retry after data loads

---

#### Test Case 6.2: Export with Network Error
**Objective:** Verify export handles API failures

**Steps:**
1. Disable network in DevTools
2. Navigate to merchant-details page
3. Click export button
4. Select any format
5. Verify error handling

**Expected Results:**
- ✅ Network error is caught
- ✅ Error message displayed to user
- ✅ Error logged to console
- ✅ Export button state is restored
- ✅ User can retry when network is restored

---

#### Test Case 6.3: Export with Invalid Format
**Objective:** Verify export handles invalid format requests

**Steps:**
1. Manually trigger export with invalid format (if possible)
2. Verify error handling

**Expected Results:**
- ✅ Invalid format is rejected
- ✅ Error message displayed
- ✅ Fallback to default format (if applicable)

---

#### Test Case 6.4: Export with Authentication Failure
**Objective:** Verify export handles auth token issues

**Steps:**
1. Clear authentication tokens
2. Navigate to merchant-details page
3. Click export button
4. Select any format
5. Verify error handling

**Expected Results:**
- ✅ Authentication error is caught
- ✅ User-friendly error message
- ✅ User is prompted to re-authenticate (if applicable)
- ✅ Error logged appropriately

---

### 7. Performance Testing

#### Test Case 7.1: Export Performance with Large Datasets
**Objective:** Verify export performance with large amounts of data

**Steps:**
1. Navigate to merchant with extensive data
2. Export each format
3. Measure export time
4. Check browser performance

**Expected Results:**
- ✅ CSV export completes quickly (< 5 seconds)
- ✅ JSON export completes quickly (< 3 seconds)
- ✅ PDF export completes within reasonable time (< 15 seconds)
- ✅ Excel export completes within reasonable time (< 10 seconds)
- ✅ No browser freezing
- ✅ Memory usage is reasonable

---

#### Test Case 7.2: Multiple Concurrent Exports
**Objective:** Verify system handles multiple export requests

**Steps:**
1. Click export button on Business Analytics tab
2. Quickly switch to Risk Assessment tab
3. Click export button there
4. Verify both exports complete

**Expected Results:**
- ✅ Both exports can run concurrently
- ✅ No conflicts or race conditions
- ✅ Each export completes independently
- ✅ Progress indicators work for each export

---

#### Test Case 7.3: Export Queue Management
**Objective:** Verify export queue handles multiple requests

**Steps:**
1. Rapidly click export button multiple times
2. Select different formats quickly
3. Verify queue management

**Expected Results:**
- ✅ Exports are queued properly
- ✅ No duplicate exports
- ✅ Exports process in order
- ✅ User feedback is clear

---

### 8. Data Accuracy Testing

#### Test Case 8.1: Data Accuracy - Business Analytics
**Objective:** Verify exported data matches displayed data

**Steps:**
1. Note all displayed values on Business Analytics tab
2. Export CSV
3. Compare exported data with displayed data
4. Export JSON
5. Compare JSON data with displayed data

**Expected Results:**
- ✅ All displayed values match exported values
- ✅ No data transformation errors
- ✅ Calculations are correct
- ✅ Percentages are accurate
- ✅ Dates are formatted correctly
- ✅ Currency values are correct

---

#### Test Case 8.2: Data Accuracy - Risk Assessment
**Objective:** Verify risk data accuracy in exports

**Steps:**
1. Note risk scores and factors on Risk Assessment tab
2. Export all formats
3. Compare exported data with displayed data

**Expected Results:**
- ✅ Risk scores match exactly
- ✅ Risk factors are complete
- ✅ Risk levels are correct
- ✅ Risk history is accurate
- ✅ Trend data is correct

---

#### Test Case 8.3: Data Consistency Across Formats
**Objective:** Verify data is consistent across all export formats

**Steps:**
1. Export same data in CSV, PDF, JSON, Excel
2. Compare data across all formats
3. Verify consistency

**Expected Results:**
- ✅ Same data in all formats
- ✅ No discrepancies
- ✅ Formatting differences are acceptable (e.g., dates)
- ✅ Core data values are identical

---

### 9. File Download Testing

#### Test Case 9.1: File Download Functionality
**Objective:** Verify files download correctly

**Steps:**
1. Export each format
2. Verify browser download behavior
3. Check download location
4. Verify file integrity

**Expected Results:**
- ✅ Files download to default download folder
- ✅ Browser download notification appears
- ✅ Files are not corrupted
- ✅ Files can be opened immediately after download
- ✅ File sizes are reasonable

---

#### Test Case 9.2: File Naming Convention
**Objective:** Verify file naming is consistent and meaningful

**Steps:**
1. Export from each tab in each format
2. Verify filename format
3. Check for naming conflicts

**Expected Results:**
- ✅ Filenames follow pattern: `{Type}_Export_{Date}.{ext}`
- ✅ Dates are in YYYY-MM-DD format
- ✅ No special characters that cause issues
- ✅ Filenames are descriptive
- ✅ No duplicate filenames (if exporting multiple times)

---

#### Test Case 9.3: File Size Validation
**Objective:** Verify exported files have reasonable sizes

**Steps:**
1. Export each format
2. Check file sizes
3. Verify they're not excessively large

**Expected Results:**
- ✅ CSV files: < 1MB for normal data
- ✅ JSON files: < 1MB for normal data
- ✅ PDF files: < 5MB for normal reports
- ✅ Excel files: < 2MB for normal data
- ✅ File sizes scale appropriately with data volume

---

### 10. User Experience Testing

#### Test Case 10.1: Export Progress Indicators
**Objective:** Verify users see progress during export

**Steps:**
1. Click export button
2. Select format
3. Observe progress indicators

**Expected Results:**
- ✅ Loading spinner or progress bar appears
- ✅ Button shows "Exporting..." state
- ✅ Progress updates (if applicable)
- ✅ User can see export is in progress
- ✅ Button is disabled during export

---

#### Test Case 10.2: Export Success Feedback
**Objective:** Verify users receive success confirmation

**Steps:**
1. Complete export
2. Verify success feedback

**Expected Results:**
- ✅ Success message appears (toast/notification)
- ✅ Button state returns to normal
- ✅ File download starts
- ✅ User knows export completed successfully

---

#### Test Case 10.3: Export Error Feedback
**Objective:** Verify users receive clear error messages

**Steps:**
1. Trigger export error (network failure, etc.)
2. Verify error message

**Expected Results:**
- ✅ Error message is user-friendly
- ✅ Error message explains what went wrong
- ✅ Error message suggests solutions (if applicable)
- ✅ Button state is restored
- ✅ User can retry export

---

### 11. Accessibility Testing

#### Test Case 11.1: Keyboard Navigation
**Objective:** Verify export buttons are keyboard accessible

**Steps:**
1. Navigate to tab using keyboard only
2. Tab to export button
3. Activate button with Enter/Space
4. Navigate dropdown with keyboard
5. Select format with keyboard

**Expected Results:**
- ✅ Export button is focusable
- ✅ Focus indicator is visible
- ✅ Button activates with Enter/Space
- ✅ Dropdown opens with keyboard
- ✅ Format options are keyboard navigable
- ✅ Format selection works with keyboard

---

#### Test Case 11.2: Screen Reader Compatibility
**Objective:** Verify export functionality works with screen readers

**Steps:**
1. Enable screen reader
2. Navigate to export button
3. Verify button is announced correctly
4. Verify dropdown options are announced

**Expected Results:**
- ✅ Button purpose is announced
- ✅ Dropdown state is announced
- ✅ Format options are announced
- ✅ Export progress is announced
- ✅ Success/error messages are announced

---

### 12. Integration Testing

#### Test Case 12.1: Export with Real API
**Objective:** Verify export works with backend API

**Steps:**
1. Ensure backend API is running
2. Export from each tab
3. Verify API calls are made
4. Verify API responses are handled

**Expected Results:**
- ✅ API calls are made to correct endpoints
- ✅ Request payload is correct
- ✅ API responses are handled
- ✅ File downloads from API URLs work
- ✅ Error responses are handled

---

#### Test Case 12.2: Export with Mock Data
**Objective:** Verify export works with mock/fallback data

**Steps:**
1. Navigate to page with mock data
2. Export from each tab
3. Verify exports complete

**Expected Results:**
- ✅ Exports work with mock data
- ✅ Mock data is clearly marked in exports (if applicable)
- ✅ No errors with mock data
- ✅ File structure is consistent

---

## Test Checklist

### Pre-Testing Checklist
- [ ] Frontend service is running
- [ ] Backend API is accessible
- [ ] Merchant data is loaded
- [ ] Browser download permissions enabled
- [ ] Browser DevTools open (Console, Network tabs)

### Export Format Checklist
- [ ] CSV export tested on all tabs
- [ ] PDF export tested on all tabs
- [ ] JSON export tested on all tabs
- [ ] Excel export tested on all tabs

### Post-Testing Checklist
- [ ] All test cases executed
- [ ] All exported files validated
- [ ] Data accuracy verified
- [ ] Error scenarios tested
- [ ] Performance acceptable
- [ ] Issues documented

---

## Common Issues and Solutions

### Issue: Export button not appearing
**Solution:** Check ExportButton component is loaded, verify container element exists, check console for errors

### Issue: Export fails with "No data available"
**Solution:** Verify merchant data is loaded, check dataSource configuration, verify merchant ID is available

### Issue: File downloads but is corrupted
**Solution:** Check API response format, verify file encoding, check browser download settings

### Issue: Export is slow
**Solution:** Check data volume, verify API performance, consider client-side export for small datasets

### Issue: PDF formatting issues
**Solution:** Check PDF template, verify CSS is included, check page size settings

---

## Automated Testing Recommendations

### Unit Tests
- Test ExportButton component initialization
- Test format selection
- Test data gathering functions
- Test filename generation
- Test error handling

### Integration Tests
- Test export API calls
- Test file download handling
- Test export with various data states
- Test export error scenarios

### E2E Tests (using Playwright/Cypress)
- Complete export flow for each format
- Verify file downloads
- Validate exported file content
- Test error handling flows

---

## Performance Benchmarks

### Expected Performance
- **CSV Export**: < 3 seconds
- **JSON Export**: < 2 seconds
- **PDF Export**: < 15 seconds
- **Excel Export**: < 10 seconds

### File Size Guidelines
- **CSV**: 50KB - 500KB (normal data)
- **JSON**: 50KB - 500KB (normal data)
- **PDF**: 200KB - 2MB (normal reports)
- **Excel**: 100KB - 1MB (normal data)

---

## Reporting Test Results

### Test Result Template
```
Test Case: [ID] - [Name]
Format: [CSV/PDF/JSON/Excel]
Tab: [Business Analytics/Risk Assessment/Risk Indicators]
Status: ✅ Pass / ❌ Fail / ⚠️ Partial
File Size: [Size]
Export Time: [Time]
Data Accuracy: ✅ Accurate / ❌ Issues Found
Notes: [Any observations]
Screenshots: [Links if applicable]
Exported File: [File path/name]
```

---

**Document Version:** 1.0.0  
**Last Updated:** December 19, 2024  
**Next Review:** March 19, 2025

