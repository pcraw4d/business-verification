# Export Functionality Guide

## Overview

The export functionality allows users to export merchant and risk data in multiple formats (CSV, PDF, JSON).

## Features

- Export merchant data
- Export risk assessment data
- Multiple format support (CSV, PDF, JSON)
- Progress tracking
- Download links

## Usage

### Merchant Export

1. Navigate to the merchant detail page
2. Click the "Export" button in the header
3. Select the desired format (CSV, PDF, or JSON)
4. Wait for the export to complete
5. Download the file when ready

### Risk Export

1. Navigate to the Risk Assessment tab on the merchant detail page
2. Click the "Export" button in the risk panel header
3. Select the desired format
4. Download the exported file

## Export Formats

### CSV
- Comma-separated values format
- Suitable for spreadsheet applications
- Includes all merchant/risk data

### PDF
- Portable Document Format
- Includes charts and visualizations
- Formatted for printing

### JSON
- JavaScript Object Notation
- Machine-readable format
- Includes all data with structure

## API Endpoints

- `POST /api/v1/export` - Export merchant/risk data
- `POST /api/v1/reports/export` - Export reports
- `GET /api/v1/export/jobs/{job_id}` - Check export job status

## Implementation

The export functionality uses the `ExportButton` component which can be integrated into any page:

```javascript
const exportButton = new ExportButton({
    container: document.getElementById('exportContainer'),
    exportType: 'merchant',
    formats: ['csv', 'pdf', 'json']
});
await exportButton.init();
```

## Error Handling

If export fails:
1. Check browser console for error messages
2. Verify API endpoint is accessible
3. Ensure you have proper permissions
4. Check network connectivity

