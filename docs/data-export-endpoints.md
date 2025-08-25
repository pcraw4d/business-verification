# Data Export Endpoints

## Overview

The Data Export API provides comprehensive data export capabilities for the KYB platform, allowing users to export various types of data in multiple formats including CSV, JSON, Excel, PDF, XML, TSV, and YAML. This API supports both immediate export generation and background job processing for large exports.

## Base URL

```
https://api.kyb-platform.com/v1
```

## Authentication

All endpoints require API key authentication via the `Authorization` header:

```
Authorization: Bearer YOUR_API_KEY
```

## Response Format

All responses are returned in JSON format with the following structure:

```json
{
  "export_id": "export_1234567890_1",
  "business_id": "business_123",
  "type": "verifications",
  "format": "csv",
  "status": "success",
  "is_successful": true,
  "file_url": "https://storage.example.com/exports/export_1234567890_1.csv",
  "file_size": 1048576,
  "row_count": 1000,
  "columns": ["id", "name", "status", "created_at"],
  "metadata": { ... },
  "expires_at": "2024-12-20T10:30:00Z",
  "generated_at": "2024-12-19T10:30:00Z",
  "processing_time": "150ms"
}
```

## Supported Export Formats

- `csv` - Comma-separated values format
- `json` - JavaScript Object Notation format
- `excel` - Microsoft Excel format (.xlsx)
- `pdf` - Portable Document Format
- `xml` - Extensible Markup Language format
- `tsv` - Tab-separated values format
- `yaml` - YAML Ain't Markup Language format

## Supported Export Types

- `verifications` - Business verification data
- `analytics` - Analytics and reporting data
- `reports` - Generated reports and summaries
- `audit_logs` - System audit logs
- `user_data` - User account and activity data
- `business_data` - Business profile and information data
- `custom` - Custom data exports

## Endpoints

### 1. Export Data

**POST** `/v1/export`

Exports data immediately with the provided configuration.

#### Request Body

```json
{
  "business_id": "business_123",
  "export_type": "verifications",
  "format": "csv",
  "filters": {
    "status": "completed",
    "date_range": "last_30_days"
  },
  "time_range": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-12-31T23:59:59Z"
  },
  "columns": ["id", "business_name", "status", "score", "created_at"],
  "sort_by": ["created_at"],
  "sort_order": "desc",
  "include_headers": true,
  "include_metadata": true,
  "compression": false,
  "password": "optional_password",
  "custom_query": "SELECT * FROM verifications WHERE status = 'completed'",
  "batch_size": 1000,
  "max_rows": 50000,
  "metadata": {
    "source": "verification_data",
    "version": "1.0"
  }
}
```

#### Response

```json
{
  "export_id": "export_1234567890_1",
  "business_id": "business_123",
  "type": "verifications",
  "format": "csv",
  "status": "success",
  "is_successful": true,
  "file_url": "https://storage.example.com/exports/export_1234567890_1.csv",
  "file_size": 1048576,
  "row_count": 1000,
  "columns": ["id", "business_name", "status", "score", "created_at"],
  "metadata": {
    "source": "verification_data",
    "version": "1.0"
  },
  "expires_at": "2024-12-20T10:30:00Z",
  "generated_at": "2024-12-19T10:30:00Z",
  "processing_time": "150ms"
}
```

### 2. Create Export Job

**POST** `/v1/export/jobs`

Creates a background job for exporting large datasets.

#### Request Body

Same as the immediate export request.

#### Response

```json
{
  "job_id": "export_job_1234567890_1",
  "business_id": "business_123",
  "type": "verifications",
  "format": "csv",
  "status": "pending",
  "progress": 0.0,
  "total_steps": 5,
  "current_step": 0,
  "step_description": "Initializing export job",
  "created_at": "2024-12-19T10:30:00Z",
  "metadata": {
    "source": "verification_data",
    "version": "1.0"
  }
}
```

### 3. Get Export Job

**GET** `/v1/export/jobs?job_id={job_id}`

Retrieves the status and results of a background export job.

#### Response

```json
{
  "job_id": "export_job_1234567890_1",
  "business_id": "business_123",
  "type": "verifications",
  "format": "csv",
  "status": "completed",
  "progress": 1.0,
  "total_steps": 5,
  "current_step": 5,
  "step_description": "Export completed successfully",
  "result": {
    "export_id": "export_1234567890_1",
    "business_id": "business_123",
    "type": "verifications",
    "format": "csv",
    "status": "success",
    "is_successful": true,
    "file_url": "https://storage.example.com/exports/export_1234567890_1.csv",
    "file_size": 2097152,
    "row_count": 5000,
    "columns": ["id", "business_name", "status", "score", "created_at", "updated_at"],
    "generated_at": "2024-12-19T10:30:00Z"
  },
  "created_at": "2024-12-19T10:30:00Z",
  "started_at": "2024-12-19T10:30:01Z",
  "completed_at": "2024-12-19T10:30:05Z",
  "metadata": {
    "source": "verification_data",
    "version": "1.0"
  }
}
```

### 4. List Export Jobs

**GET** `/v1/export/jobs`

Lists all export jobs with optional filtering and pagination.

#### Query Parameters

- `status` (optional): Filter by job status (pending, processing, completed, failed, cancelled)
- `business_id` (optional): Filter by business ID
- `limit` (optional): Number of jobs to return (default: 50, max: 100)
- `offset` (optional): Number of jobs to skip (default: 0)

#### Response

```json
{
  "jobs": [
    {
      "job_id": "export_job_1234567890_1",
      "business_id": "business_123",
      "type": "verifications",
      "format": "csv",
      "status": "completed",
      "progress": 1.0,
      "created_at": "2024-12-19T10:30:00Z"
    }
  ],
  "total_count": 1,
  "limit": 50,
  "offset": 0
}
```

### 5. Get Export Template

**GET** `/v1/export/templates?template_id={template_id}`

Retrieves a pre-configured export template.

#### Response

```json
{
  "id": "verifications_csv",
  "name": "Verifications CSV Export",
  "description": "Export verification data in CSV format",
  "type": "verifications",
  "format": "csv",
  "columns": ["id", "business_name", "status", "score", "created_at", "updated_at"],
  "sort_by": ["created_at"],
  "sort_order": "desc",
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

### 6. List Export Templates

**GET** `/v1/export/templates`

Lists all available export templates with optional filtering and pagination.

#### Query Parameters

- `type` (optional): Filter by export type
- `format` (optional): Filter by export format
- `limit` (optional): Number of templates to return (default: 50, max: 100)
- `offset` (optional): Number of templates to skip (default: 0)

#### Response

```json
{
  "templates": [
    {
      "id": "verifications_csv",
      "name": "Verifications CSV Export",
      "description": "Export verification data in CSV format",
      "type": "verifications",
      "format": "csv",
      "created_at": "2024-12-19T10:30:00Z",
      "updated_at": "2024-12-19T10:30:00Z"
    }
  ],
  "total_count": 1,
  "limit": 50,
  "offset": 0
}
```

## Configuration Options

### DataExportRequest

```json
{
  "business_id": "business_123",
  "export_type": "verifications",
  "format": "csv",
  "filters": {
    "status": "completed",
    "date_range": "last_30_days",
    "score_min": 0.8
  },
  "time_range": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-12-31T23:59:59Z"
  },
  "columns": ["id", "business_name", "status", "score", "created_at"],
  "sort_by": ["created_at", "business_name"],
  "sort_order": "desc",
  "include_headers": true,
  "include_metadata": true,
  "compression": false,
  "password": "optional_password",
  "custom_query": "SELECT * FROM verifications WHERE status = 'completed'",
  "batch_size": 1000,
  "max_rows": 50000,
  "metadata": {
    "source": "verification_data",
    "version": "1.0",
    "exported_by": "user_123"
  }
}
```

### Export Formats

#### CSV Format
- Comma-separated values
- Includes headers by default
- Supports custom delimiters
- UTF-8 encoding

#### JSON Format
- Structured data format
- Supports nested objects and arrays
- Pretty-printed by default
- UTF-8 encoding

#### Excel Format
- Microsoft Excel (.xlsx) format
- Multiple worksheets support
- Cell formatting and styling
- Charts and graphs support

#### PDF Format
- Portable Document Format
- Professional formatting
- Table layouts
- Header and footer support

#### XML Format
- Extensible Markup Language
- Hierarchical data structure
- Custom schema support
- UTF-8 encoding

#### TSV Format
- Tab-separated values
- Similar to CSV but with tab delimiters
- Includes headers by default
- UTF-8 encoding

#### YAML Format
- YAML Ain't Markup Language
- Human-readable format
- Supports complex data structures
- UTF-8 encoding

## Error Responses

### Validation Error

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "business_id is required"
  },
  "timestamp": "2024-12-19T10:30:00Z"
}
```

### Job Not Found

```json
{
  "error": {
    "code": "JOB_NOT_FOUND",
    "message": "Export job not found"
  },
  "timestamp": "2024-12-19T10:30:00Z"
}
```

### Processing Error

```json
{
  "error": {
    "code": "EXPORT_ERROR",
    "message": "Failed to generate export"
  },
  "timestamp": "2024-12-19T10:30:00Z"
}
```

## Integration Examples

### JavaScript/TypeScript

```javascript
// Export data immediately
async function exportData() {
  const response = await fetch('https://api.kyb-platform.com/v1/export', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      business_id: 'business_123',
      export_type: 'verifications',
      format: 'csv',
      columns: ['id', 'business_name', 'status', 'score'],
      filters: {
        status: 'completed'
      }
    })
  });

  const exportResult = await response.json();
  return exportResult;
}

// Create a background export job
async function createExportJob() {
  const response = await fetch('https://api.kyb-platform.com/v1/export/jobs', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      business_id: 'business_123',
      export_type: 'analytics',
      format: 'excel',
      max_rows: 100000
    })
  });

  const job = await response.json();
  return job;
}

// Poll job status
async function pollJobStatus(jobId) {
  const response = await fetch(`https://api.kyb-platform.com/v1/export/jobs?job_id=${jobId}`, {
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY'
    }
  });

  const job = await response.json();
  return job;
}

// Get export template
async function getExportTemplate(templateId) {
  const response = await fetch(`https://api.kyb-platform.com/v1/export/templates?template_id=${templateId}`, {
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY'
    }
  });

  const template = await response.json();
  return template;
}

// Download export file
async function downloadExportFile(fileUrl) {
  const response = await fetch(fileUrl, {
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY'
    }
  });

  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'export.csv';
  a.click();
  window.URL.revokeObjectURL(url);
}
```

### Python

```python
import requests
import json

# Export data immediately
def export_data():
    url = 'https://api.kyb-platform.com/v1/export'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
    }
    
    data = {
        'business_id': 'business_123',
        'export_type': 'verifications',
        'format': 'csv',
        'columns': ['id', 'business_name', 'status', 'score'],
        'filters': {
            'status': 'completed'
        }
    }
    
    response = requests.post(url, headers=headers, json=data)
    return response.json()

# Create a background export job
def create_export_job():
    url = 'https://api.kyb-platform.com/v1/export/jobs'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
    }
    
    data = {
        'business_id': 'business_123',
        'export_type': 'analytics',
        'format': 'excel',
        'max_rows': 100000
    }
    
    response = requests.post(url, headers=headers, json=data)
    return response.json()

# Poll job status
def poll_job_status(job_id):
    url = f'https://api.kyb-platform.com/v1/export/jobs?job_id={job_id}'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY'
    }
    
    response = requests.get(url, headers=headers)
    return response.json()

# Get export template
def get_export_template(template_id):
    url = f'https://api.kyb-platform.com/v1/export/templates?template_id={template_id}'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY'
    }
    
    response = requests.get(url, headers=headers)
    return response.json()

# Download export file
def download_export_file(file_url, filename):
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY'
    }
    
    response = requests.get(file_url, headers=headers, stream=True)
    
    with open(filename, 'wb') as f:
        for chunk in response.iter_content(chunk_size=8192):
            f.write(chunk)
    
    return filename
```

### React Component

```jsx
import React, { useState, useEffect } from 'react';

const ExportComponent = () => {
  const [exportData, setExportData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [jobStatus, setJobStatus] = useState(null);

  const exportDataImmediately = async () => {
    setLoading(true);
    try {
      const response = await fetch('https://api.kyb-platform.com/v1/export', {
        method: 'POST',
        headers: {
          'Authorization': 'Bearer YOUR_API_KEY',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          business_id: 'business_123',
          export_type: 'verifications',
          format: 'csv',
          columns: ['id', 'business_name', 'status', 'score']
        })
      });

      const result = await response.json();
      setExportData(result);
    } catch (error) {
      console.error('Error exporting data:', error);
    } finally {
      setLoading(false);
    }
  };

  const createExportJob = async () => {
    setLoading(true);
    try {
      const response = await fetch('https://api.kyb-platform.com/v1/export/jobs', {
        method: 'POST',
        headers: {
          'Authorization': 'Bearer YOUR_API_KEY',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          business_id: 'business_123',
          export_type: 'analytics',
          format: 'excel',
          max_rows: 100000
        })
      });

      const job = await response.json();
      setJobStatus(job);
      
      // Start polling for job status
      pollJobStatus(job.job_id);
    } catch (error) {
      console.error('Error creating export job:', error);
    } finally {
      setLoading(false);
    }
  };

  const pollJobStatus = async (jobId) => {
    const interval = setInterval(async () => {
      try {
        const response = await fetch(`https://api.kyb-platform.com/v1/export/jobs?job_id=${jobId}`, {
          headers: {
            'Authorization': 'Bearer YOUR_API_KEY'
          }
        });

        const job = await response.json();
        setJobStatus(job);

        if (job.status === 'completed' || job.status === 'failed') {
          clearInterval(interval);
        }
      } catch (error) {
        console.error('Error polling job status:', error);
        clearInterval(interval);
      }
    }, 2000); // Poll every 2 seconds
  };

  const downloadFile = async (fileUrl) => {
    try {
      const response = await fetch(fileUrl, {
        headers: {
          'Authorization': 'Bearer YOUR_API_KEY'
        }
      });

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'export.csv';
      a.click();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Error downloading file:', error);
    }
  };

  return (
    <div>
      <h2>Data Export</h2>
      
      <button onClick={exportDataImmediately} disabled={loading}>
        {loading ? 'Exporting...' : 'Export Data Immediately'}
      </button>
      
      <button onClick={createExportJob} disabled={loading}>
        {loading ? 'Creating Job...' : 'Create Export Job'}
      </button>

      {exportData && (
        <div>
          <h3>Export Result</h3>
          <p>Export ID: {exportData.export_id}</p>
          <p>Status: {exportData.status}</p>
          <p>File Size: {exportData.file_size} bytes</p>
          <p>Row Count: {exportData.row_count}</p>
          <button onClick={() => downloadFile(exportData.file_url)}>
            Download File
          </button>
        </div>
      )}

      {jobStatus && (
        <div>
          <h3>Export Job Status</h3>
          <p>Job ID: {jobStatus.job_id}</p>
          <p>Status: {jobStatus.status}</p>
          <p>Progress: {(jobStatus.progress * 100).toFixed(1)}%</p>
          <p>Step: {jobStatus.step_description}</p>
          
          {jobStatus.result && (
            <div>
              <h4>Job Completed</h4>
              <p>File URL: {jobStatus.result.file_url}</p>
              <p>File Size: {jobStatus.result.file_size} bytes</p>
              <p>Row Count: {jobStatus.result.row_count}</p>
              <button onClick={() => downloadFile(jobStatus.result.file_url)}>
                Download File
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default ExportComponent;
```

## Best Practices

### 1. Data Preparation

- Ensure data is properly filtered and validated before export
- Use appropriate column selection to reduce file size
- Consider data privacy and compliance requirements
- Validate export parameters before submission

### 2. Format Selection

- Use CSV for simple tabular data and compatibility
- Use JSON for structured data and API integration
- Use Excel for complex data with formatting requirements
- Use PDF for reports and documentation
- Use XML for data exchange with external systems

### 3. Background Jobs

- Use background jobs for large datasets (> 10,000 rows)
- Implement proper polling mechanisms for job status
- Handle job failures gracefully
- Consider job cleanup and retention policies

### 4. File Management

- Download files promptly (they expire after 24 hours)
- Implement proper file storage and backup
- Consider file compression for large exports
- Use secure file URLs with authentication

### 5. Error Handling

- Implement proper error handling for all API calls
- Display meaningful error messages to users
- Retry failed requests when appropriate
- Log errors for debugging and monitoring

### 6. Performance

- Use appropriate batch sizes for large exports
- Consider data pagination for very large datasets
- Optimize query performance with proper indexing
- Use background jobs for heavy processing

### 7. Security

- Validate all input data and parameters
- Implement proper access controls
- Use secure file storage and URLs
- Monitor for abuse and rate limiting

## Rate Limiting

- **Standard Exports**: 50 requests per minute per API key
- **Background Jobs**: 10 job creations per minute per API key
- **Template Retrieval**: 100 requests per minute per API key
- **File Downloads**: 200 requests per minute per API key

## Monitoring and Observability

### Key Metrics

- **Export Request Rate**: Number of export requests per minute
- **Success Rate**: Percentage of successful exports
- **Processing Time**: Average time to generate exports
- **Error Rate**: Percentage of failed exports
- **Job Completion Rate**: Percentage of completed background jobs
- **File Download Rate**: Number of file downloads per minute

### Health Checks

Monitor the following endpoints for system health:

```bash
# Check export service health
curl -X GET "https://api.kyb-platform.com/v1/health/export" \
  -H "Authorization: Bearer YOUR_API_KEY"

# Check background job processing
curl -X GET "https://api.kyb-platform.com/v1/health/jobs" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Logging

All export operations are logged with the following information:

- Request ID for correlation
- Business ID for tracking
- Export type and format
- Processing time and performance metrics
- Error details and stack traces
- File size and row count information

## Troubleshooting

### Common Issues

1. **Validation Errors**
   - Ensure all required fields are provided
   - Check data format and types
   - Verify export type and format compatibility

2. **Job Failures**
   - Check job status and error messages
   - Verify data size and complexity
   - Monitor system resources

3. **Performance Issues**
   - Use background jobs for large exports
   - Optimize query parameters
   - Consider data filtering and pagination

4. **File Access Issues**
   - Verify file URLs are still valid (24-hour expiration)
   - Check authentication and permissions
   - Ensure proper file storage configuration

5. **Authentication Errors**
   - Verify API key is valid and active
   - Check API key permissions
   - Ensure proper header format

### Debug Information

Enable debug logging by including the `X-Debug` header:

```bash
curl -X POST "https://api.kyb-platform.com/v1/export" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "X-Debug: true" \
  -H "Content-Type: application/json" \
  -d '{"business_id": "business_123", "export_type": "verifications", "format": "csv"}'
```

### Support

For additional support and troubleshooting:

- Check the API documentation for detailed endpoint information
- Review error logs and monitoring dashboards
- Contact support with request IDs and error details
- Provide reproducible examples for complex issues

## Migration Guide

### From Previous Versions

If migrating from previous export APIs:

1. **Update Endpoint URLs**: Use the new `/v1/export` endpoints
2. **Update Request Format**: Follow the new request structure
3. **Update Response Handling**: Handle the new response format
4. **Test Thoroughly**: Verify all exports work correctly
5. **Update Documentation**: Update client documentation and examples

### Breaking Changes

- New authentication requirements
- Updated request/response formats
- New error codes and messages
- Enhanced validation rules
- Improved performance characteristics

## Future Enhancements

### Planned Features

1. **Real-time Exports**: Streaming exports for large datasets
2. **Advanced Formats**: More export formats and configurations
3. **Scheduled Exports**: Automated export scheduling
4. **Export Templates**: Custom export template creation
5. **Data Transformation**: Built-in data transformation capabilities
6. **Export Analytics**: Export usage analytics and insights
7. **Batch Processing**: Batch export operations
8. **Export Notifications**: Email and webhook notifications for completed exports

### API Versioning

The export API follows semantic versioning:

- **v1**: Current stable version
- **v2**: Planned major version with new features
- **Beta**: Experimental features and endpoints

Check the API documentation for the latest version information and migration guides.
