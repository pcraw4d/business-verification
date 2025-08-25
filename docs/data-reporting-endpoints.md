# Data Reporting Endpoints

## Overview

The Data Reporting API provides comprehensive report generation capabilities for the KYB platform, allowing users to generate various types of business intelligence reports including verification summaries, analytics dashboards, compliance reports, risk assessments, and custom reports. This API supports both immediate report generation and background job processing with scheduling capabilities.

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
  "report_id": "report_1234567890_1",
  "business_id": "business_123",
  "type": "verification_summary",
  "format": "pdf",
  "title": "Verification Summary Report",
  "status": "success",
  "is_successful": true,
  "file_url": "https://storage.example.com/reports/report_1234567890_1.pdf",
  "file_size": 2097152,
  "page_count": 15,
  "generated_at": "2024-12-19T10:30:00Z",
  "processing_time": "300ms",
  "summary": {
    "total_records": 1500,
    "date_range": {
      "start": "2024-11-19T10:30:00Z",
      "end": "2024-12-19T10:30:00Z"
    },
    "key_metrics": {
      "total_verifications": 1500,
      "success_rate": 0.95,
      "average_score": 0.87,
      "compliance_rate": 0.92
    },
    "charts": [
      {
        "title": "Verification Trends",
        "type": "line_chart",
        "data_points": 30
      }
    ],
    "tables": [
      {
        "title": "Verification Summary",
        "row_count": 50,
        "column_count": 8
      }
    ],
    "recommendations": [
      "Increase verification frequency for high-risk businesses"
    ]
  },
  "metadata": { ... },
  "expires_at": "2025-01-18T10:30:00Z"
}
```

## Supported Report Types

- `verification_summary` - Business verification summary reports
- `analytics` - Analytics and dashboard reports
- `compliance` - Compliance and regulatory reports
- `risk_assessment` - Risk assessment and scoring reports
- `audit_trail` - Audit trail and activity reports
- `performance` - Performance and metrics reports
- `custom` - Custom report configurations

## Supported Report Formats

- `pdf` - Portable Document Format (professional reports)
- `html` - HyperText Markup Language (interactive dashboards)
- `json` - JavaScript Object Notation (data export)
- `excel` - Microsoft Excel format (.xlsx)
- `csv` - Comma-separated values format

## Endpoints

### 1. Generate Report

**POST** `/v1/reports`

Generates a report immediately with the provided configuration.

#### Request Body

```json
{
  "business_id": "business_123",
  "report_type": "verification_summary",
  "format": "pdf",
  "title": "Verification Summary Report",
  "description": "Monthly verification summary with trends and analysis",
  "filters": {
    "status": "completed",
    "date_range": "last_30_days"
  },
  "time_range": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-12-31T23:59:59Z"
  },
  "parameters": {
    "include_charts": true,
    "include_tables": true,
    "include_summary": true,
    "include_details": false
  },
  "include_charts": true,
  "include_tables": true,
  "include_summary": true,
  "include_details": false,
  "custom_template": "custom_template_id",
  "schedule": {
    "type": "monthly",
    "start_date": "2024-01-01T00:00:00Z",
    "day_of_month": 1,
    "time_of_day": "09:00",
    "timezone": "UTC",
    "enabled": true
  },
  "recipients": ["user@example.com", "admin@example.com"],
  "metadata": {
    "source": "verification_data",
    "version": "1.0"
  }
}
```

#### Response

```json
{
  "report_id": "report_1234567890_1",
  "business_id": "business_123",
  "type": "verification_summary",
  "format": "pdf",
  "title": "Verification Summary Report",
  "status": "success",
  "is_successful": true,
  "file_url": "https://storage.example.com/reports/report_1234567890_1.pdf",
  "file_size": 2097152,
  "page_count": 15,
  "generated_at": "2024-12-19T10:30:00Z",
  "processing_time": "300ms",
  "summary": {
    "total_records": 1500,
    "date_range": {
      "start": "2024-11-19T10:30:00Z",
      "end": "2024-12-19T10:30:00Z"
    },
    "key_metrics": {
      "total_verifications": 1500,
      "success_rate": 0.95,
      "average_score": 0.87,
      "compliance_rate": 0.92
    },
    "charts": [
      {
        "title": "Verification Trends",
        "type": "line_chart",
        "data_points": 30
      },
      {
        "title": "Success Rate by Industry",
        "type": "bar_chart",
        "data_points": 10
      }
    ],
    "tables": [
      {
        "title": "Verification Summary",
        "row_count": 50,
        "column_count": 8
      },
      {
        "title": "Risk Assessment",
        "row_count": 25,
        "column_count": 6
      }
    ],
    "recommendations": [
      "Increase verification frequency for high-risk businesses",
      "Implement additional compliance checks for financial services",
      "Consider automated risk scoring for faster processing"
    ]
  },
  "metadata": {
    "source": "verification_data",
    "version": "1.0"
  },
  "expires_at": "2025-01-18T10:30:00Z"
}
```

### 2. Create Report Job

**POST** `/v1/reports/jobs`

Creates a background job for generating large or complex reports.

#### Request Body

Same as the immediate report generation request.

#### Response

```json
{
  "job_id": "report_job_1234567890_1",
  "business_id": "business_123",
  "type": "verification_summary",
  "format": "pdf",
  "title": "Verification Summary Report",
  "status": "pending",
  "progress": 0.0,
  "total_steps": 6,
  "current_step": 0,
  "step_description": "Initializing report generation",
  "created_at": "2024-12-19T10:30:00Z",
  "next_run_at": "2025-01-01T09:00:00Z",
  "metadata": {
    "source": "verification_data",
    "version": "1.0"
  }
}
```

### 3. Get Report Job

**GET** `/v1/reports/jobs?job_id={job_id}`

Retrieves the status and results of a background report generation job.

#### Response

```json
{
  "job_id": "report_job_1234567890_1",
  "business_id": "business_123",
  "type": "verification_summary",
  "format": "pdf",
  "title": "Verification Summary Report",
  "status": "completed",
  "progress": 1.0,
  "total_steps": 6,
  "current_step": 6,
  "step_description": "Report generation completed successfully",
  "result": {
    "report_id": "report_1234567890_1",
    "business_id": "business_123",
    "type": "verification_summary",
    "format": "pdf",
    "title": "Verification Summary Report",
    "status": "success",
    "is_successful": true,
    "file_url": "https://storage.example.com/reports/report_1234567890_1.pdf",
    "file_size": 3145728,
    "page_count": 20,
    "generated_at": "2024-12-19T10:30:00Z"
  },
  "created_at": "2024-12-19T10:30:00Z",
  "started_at": "2024-12-19T10:30:01Z",
  "completed_at": "2024-12-19T10:30:05Z",
  "next_run_at": "2025-01-01T09:00:00Z",
  "metadata": {
    "source": "verification_data",
    "version": "1.0"
  }
}
```

### 4. List Report Jobs

**GET** `/v1/reports/jobs`

Lists all report generation jobs with optional filtering and pagination.

#### Query Parameters

- `status` (optional): Filter by job status (pending, processing, completed, failed, cancelled)
- `business_id` (optional): Filter by business ID
- `report_type` (optional): Filter by report type
- `limit` (optional): Number of jobs to return (default: 50, max: 100)
- `offset` (optional): Number of jobs to skip (default: 0)

#### Response

```json
{
  "jobs": [
    {
      "job_id": "report_job_1234567890_1",
      "business_id": "business_123",
      "type": "verification_summary",
      "format": "pdf",
      "title": "Verification Summary Report",
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

### 5. Get Report Template

**GET** `/v1/reports/templates?template_id={template_id}`

Retrieves a pre-configured report template.

#### Response

```json
{
  "id": "verification_summary_pdf",
  "name": "Verification Summary Report",
  "description": "Comprehensive verification summary with charts and analysis",
  "type": "verification_summary",
  "format": "pdf",
  "parameters": {
    "include_charts": true,
    "include_tables": true,
    "include_summary": true,
    "include_details": false
  },
  "include_charts": true,
  "include_tables": true,
  "include_summary": true,
  "include_details": false,
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

### 6. List Report Templates

**GET** `/v1/reports/templates`

Lists all available report templates with optional filtering and pagination.

#### Query Parameters

- `report_type` (optional): Filter by report type
- `format` (optional): Filter by report format
- `limit` (optional): Number of templates to return (default: 50, max: 100)
- `offset` (optional): Number of templates to skip (default: 0)

#### Response

```json
{
  "templates": [
    {
      "id": "verification_summary_pdf",
      "name": "Verification Summary Report",
      "description": "Comprehensive verification summary with charts and analysis",
      "type": "verification_summary",
      "format": "pdf",
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

### DataReportingRequest

```json
{
  "business_id": "business_123",
  "report_type": "verification_summary",
  "format": "pdf",
  "title": "Verification Summary Report",
  "description": "Monthly verification summary with trends and analysis",
  "filters": {
    "status": "completed",
    "date_range": "last_30_days",
    "score_min": 0.8
  },
  "time_range": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-12-31T23:59:59Z"
  },
  "parameters": {
    "include_charts": true,
    "include_tables": true,
    "include_summary": true,
    "include_details": false,
    "chart_types": ["line", "bar", "pie"],
    "table_columns": ["id", "name", "status", "score", "created_at"]
  },
  "include_charts": true,
  "include_tables": true,
  "include_summary": true,
  "include_details": false,
  "custom_template": "custom_template_id",
  "schedule": {
    "type": "monthly",
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-12-31T23:59:59Z",
    "time_of_day": "09:00",
    "day_of_week": 1,
    "day_of_month": 1,
    "month_of_year": 1,
    "timezone": "UTC",
    "enabled": true,
    "max_occurrences": 12
  },
  "recipients": ["user@example.com", "admin@example.com"],
  "metadata": {
    "source": "verification_data",
    "version": "1.0",
    "generated_by": "user_123"
  }
}
```

### Report Formats

#### PDF Format
- Professional document format
- Includes charts, tables, and summaries
- Supports custom styling and branding
- Page count and file size information
- 30-day expiration

#### HTML Format
- Interactive web-based reports
- Real-time data updates
- Responsive design for mobile devices
- Embedded charts and interactive tables
- Custom CSS styling support

#### JSON Format
- Structured data export
- Includes all report data and metadata
- Machine-readable format
- Suitable for API integration
- No expiration (data format)

#### Excel Format
- Microsoft Excel (.xlsx) format
- Multiple worksheets support
- Chart and table formatting
- Formula and calculation support
- 30-day expiration

#### CSV Format
- Simple tabular data format
- Comma-separated values
- Lightweight and portable
- Suitable for data analysis tools
- 30-day expiration

## Scheduling Options

### Schedule Types

#### One Time
- Single report generation
- No recurring schedule
- Immediate or future execution

#### Daily
- Report generated every day
- Configurable time of day
- Optional end date

#### Weekly
- Report generated on specific day of week
- Configurable time of day
- Day of week: 0-6 (Sunday-Saturday)

#### Monthly
- Report generated on specific day of month
- Configurable time of day
- Day of month: 1-31

#### Quarterly
- Report generated quarterly
- Configurable month and day
- Month of year: 1-12

#### Yearly
- Report generated annually
- Configurable month and day
- Month of year: 1-12

### Schedule Configuration

```json
{
  "type": "monthly",
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-12-31T23:59:59Z",
  "time_of_day": "09:00",
  "day_of_week": 1,
  "day_of_month": 1,
  "month_of_year": 1,
  "timezone": "UTC",
  "enabled": true,
  "max_occurrences": 12
}
```

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
    "message": "Report job not found"
  },
  "timestamp": "2024-12-19T10:30:00Z"
}
```

### Processing Error

```json
{
  "error": {
    "code": "REPORT_ERROR",
    "message": "Failed to generate report"
  },
  "timestamp": "2024-12-19T10:30:00Z"
}
```

## Integration Examples

### JavaScript/TypeScript

```javascript
// Generate report immediately
async function generateReport() {
  const response = await fetch('https://api.kyb-platform.com/v1/reports', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      business_id: 'business_123',
      report_type: 'verification_summary',
      format: 'pdf',
      title: 'Verification Summary Report',
      include_charts: true,
      include_tables: true,
      filters: {
        status: 'completed'
      }
    })
  });

  const reportResult = await response.json();
  return reportResult;
}

// Create a background report job
async function createReportJob() {
  const response = await fetch('https://api.kyb-platform.com/v1/reports/jobs', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      business_id: 'business_123',
      report_type: 'analytics',
      format: 'html',
      title: 'Analytics Dashboard',
      schedule: {
        type: 'daily',
        start_date: new Date().toISOString(),
        time_of_day: '09:00',
        enabled: true
      }
    })
  });

  const job = await response.json();
  return job;
}

// Poll job status
async function pollJobStatus(jobId) {
  const response = await fetch(`https://api.kyb-platform.com/v1/reports/jobs?job_id=${jobId}`, {
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY'
    }
  });

  const job = await response.json();
  return job;
}

// Get report template
async function getReportTemplate(templateId) {
  const response = await fetch(`https://api.kyb-platform.com/v1/reports/templates?template_id=${templateId}`, {
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY'
    }
  });

  const template = await response.json();
  return template;
}

// Download report file
async function downloadReportFile(fileUrl) {
  const response = await fetch(fileUrl, {
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY'
    }
  });

  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'report.pdf';
  a.click();
  window.URL.revokeObjectURL(url);
}
```

### Python

```python
import requests
import json

# Generate report immediately
def generate_report():
    url = 'https://api.kyb-platform.com/v1/reports'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
    }
    
    data = {
        'business_id': 'business_123',
        'report_type': 'verification_summary',
        'format': 'pdf',
        'title': 'Verification Summary Report',
        'include_charts': True,
        'include_tables': True,
        'filters': {
            'status': 'completed'
        }
    }
    
    response = requests.post(url, headers=headers, json=data)
    return response.json()

# Create a background report job
def create_report_job():
    url = 'https://api.kyb-platform.com/v1/reports/jobs'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
    }
    
    data = {
        'business_id': 'business_123',
        'report_type': 'analytics',
        'format': 'html',
        'title': 'Analytics Dashboard',
        'schedule': {
            'type': 'daily',
            'start_date': '2024-01-01T00:00:00Z',
            'time_of_day': '09:00',
            'enabled': True
        }
    }
    
    response = requests.post(url, headers=headers, json=data)
    return response.json()

# Poll job status
def poll_job_status(job_id):
    url = f'https://api.kyb-platform.com/v1/reports/jobs?job_id={job_id}'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY'
    }
    
    response = requests.get(url, headers=headers)
    return response.json()

# Get report template
def get_report_template(template_id):
    url = f'https://api.kyb-platform.com/v1/reports/templates?template_id={template_id}'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY'
    }
    
    response = requests.get(url, headers=headers)
    return response.json()

# Download report file
def download_report_file(file_url, filename):
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

const ReportComponent = () => {
  const [reportData, setReportData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [jobStatus, setJobStatus] = useState(null);

  const generateReportImmediately = async () => {
    setLoading(true);
    try {
      const response = await fetch('https://api.kyb-platform.com/v1/reports', {
        method: 'POST',
        headers: {
          'Authorization': 'Bearer YOUR_API_KEY',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          business_id: 'business_123',
          report_type: 'verification_summary',
          format: 'pdf',
          title: 'Verification Summary Report',
          include_charts: true,
          include_tables: true
        })
      });

      const result = await response.json();
      setReportData(result);
    } catch (error) {
      console.error('Error generating report:', error);
    } finally {
      setLoading(false);
    }
  };

  const createReportJob = async () => {
    setLoading(true);
    try {
      const response = await fetch('https://api.kyb-platform.com/v1/reports/jobs', {
        method: 'POST',
        headers: {
          'Authorization': 'Bearer YOUR_API_KEY',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          business_id: 'business_123',
          report_type: 'analytics',
          format: 'html',
          title: 'Analytics Dashboard',
          schedule: {
            type: 'daily',
            start_date: new Date().toISOString(),
            time_of_day: '09:00',
            enabled: true
          }
        })
      });

      const job = await response.json();
      setJobStatus(job);
      
      // Start polling for job status
      pollJobStatus(job.job_id);
    } catch (error) {
      console.error('Error creating report job:', error);
    } finally {
      setLoading(false);
    }
  };

  const pollJobStatus = async (jobId) => {
    const interval = setInterval(async () => {
      try {
        const response = await fetch(`https://api.kyb-platform.com/v1/reports/jobs?job_id=${jobId}`, {
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
      a.download = 'report.pdf';
      a.click();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Error downloading file:', error);
    }
  };

  return (
    <div>
      <h2>Data Reporting</h2>
      
      <button onClick={generateReportImmediately} disabled={loading}>
        {loading ? 'Generating...' : 'Generate Report Immediately'}
      </button>
      
      <button onClick={createReportJob} disabled={loading}>
        {loading ? 'Creating Job...' : 'Create Report Job'}
      </button>

      {reportData && (
        <div>
          <h3>Report Result</h3>
          <p>Report ID: {reportData.report_id}</p>
          <p>Status: {reportData.status}</p>
          <p>File Size: {reportData.file_size} bytes</p>
          <p>Page Count: {reportData.page_count}</p>
          <p>Processing Time: {reportData.processing_time}</p>
          <button onClick={() => downloadFile(reportData.file_url)}>
            Download Report
          </button>
          
          {reportData.summary && (
            <div>
              <h4>Report Summary</h4>
              <p>Total Records: {reportData.summary.total_records}</p>
              <p>Charts: {reportData.summary.charts.length}</p>
              <p>Tables: {reportData.summary.tables.length}</p>
              <p>Recommendations: {reportData.summary.recommendations.length}</p>
            </div>
          )}
        </div>
      )}

      {jobStatus && (
        <div>
          <h3>Report Job Status</h3>
          <p>Job ID: {jobStatus.job_id}</p>
          <p>Status: {jobStatus.status}</p>
          <p>Progress: {(jobStatus.progress * 100).toFixed(1)}%</p>
          <p>Step: {jobStatus.step_description}</p>
          
          {jobStatus.result && (
            <div>
              <h4>Job Completed</h4>
              <p>File URL: {jobStatus.result.file_url}</p>
              <p>File Size: {jobStatus.result.file_size} bytes</p>
              <p>Page Count: {jobStatus.result.page_count}</p>
              <button onClick={() => downloadFile(jobStatus.result.file_url)}>
                Download Report
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default ReportComponent;
```

## Best Practices

### 1. Report Design

- Use appropriate report types for different use cases
- Include relevant charts and tables for data visualization
- Provide clear summaries and recommendations
- Consider audience and purpose when selecting format

### 2. Scheduling

- Use scheduling for regular reports and dashboards
- Set appropriate time zones for global teams
- Consider business hours and peak usage times
- Implement proper error handling for failed schedules

### 3. Performance

- Use background jobs for large or complex reports
- Implement proper polling mechanisms for job status
- Consider report caching for frequently accessed data
- Optimize query performance for large datasets

### 4. File Management

- Download files promptly (they expire after 30 days)
- Implement proper file storage and backup
- Consider file compression for large reports
- Use secure file URLs with authentication

### 5. Error Handling

- Implement proper error handling for all API calls
- Display meaningful error messages to users
- Retry failed requests when appropriate
- Log errors for debugging and monitoring

### 6. Security

- Validate all input data and parameters
- Implement proper access controls
- Use secure file storage and URLs
- Monitor for abuse and rate limiting

### 7. Monitoring

- Track report generation success rates
- Monitor job completion times
- Alert on failed report generations
- Monitor file download patterns

## Rate Limiting

- **Standard Reports**: 20 requests per minute per API key
- **Background Jobs**: 5 job creations per minute per API key
- **Template Retrieval**: 50 requests per minute per API key
- **File Downloads**: 100 requests per minute per API key

## Monitoring and Observability

### Key Metrics

- **Report Generation Rate**: Number of reports generated per minute
- **Success Rate**: Percentage of successful report generations
- **Processing Time**: Average time to generate reports
- **Error Rate**: Percentage of failed report generations
- **Job Completion Rate**: Percentage of completed background jobs
- **File Download Rate**: Number of file downloads per minute

### Health Checks

Monitor the following endpoints for system health:

```bash
# Check report service health
curl -X GET "https://api.kyb-platform.com/v1/health/reports" \
  -H "Authorization: Bearer YOUR_API_KEY"

# Check background job processing
curl -X GET "https://api.kyb-platform.com/v1/health/jobs" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Logging

All report operations are logged with the following information:

- Request ID for correlation
- Business ID for tracking
- Report type and format
- Processing time and performance metrics
- Error details and stack traces
- File size and page count information

## Troubleshooting

### Common Issues

1. **Validation Errors**
   - Ensure all required fields are provided
   - Check data format and types
   - Verify report type and format compatibility

2. **Job Failures**
   - Check job status and error messages
   - Verify data size and complexity
   - Monitor system resources

3. **Performance Issues**
   - Use background jobs for large reports
   - Optimize query parameters
   - Consider data filtering and pagination

4. **File Access Issues**
   - Verify file URLs are still valid (30-day expiration)
   - Check authentication and permissions
   - Ensure proper file storage configuration

5. **Authentication Errors**
   - Verify API key is valid and active
   - Check API key permissions
   - Ensure proper header format

### Debug Information

Enable debug logging by including the `X-Debug` header:

```bash
curl -X POST "https://api.kyb-platform.com/v1/reports" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "X-Debug: true" \
  -H "Content-Type: application/json" \
  -d '{"business_id": "business_123", "report_type": "verification_summary", "format": "pdf", "title": "Test Report"}'
```

### Support

For additional support and troubleshooting:

- Check the API documentation for detailed endpoint information
- Review error logs and monitoring dashboards
- Contact support with request IDs and error details
- Provide reproducible examples for complex issues

## Migration Guide

### From Previous Versions

If migrating from previous reporting APIs:

1. **Update Endpoint URLs**: Use the new `/v1/reports` endpoints
2. **Update Request Format**: Follow the new request structure
3. **Update Response Handling**: Handle the new response format
4. **Test Thoroughly**: Verify all reports work correctly
5. **Update Documentation**: Update client documentation and examples

### Breaking Changes

- New authentication requirements
- Updated request/response formats
- New error codes and messages
- Enhanced validation rules
- Improved performance characteristics

## Future Enhancements

### Planned Features

1. **Real-time Reports**: Streaming reports for live data
2. **Advanced Formats**: More report formats and configurations
3. **Custom Templates**: User-defined report templates
4. **Report Analytics**: Report usage analytics and insights
5. **Collaborative Reports**: Shared and collaborative reporting
6. **Report Notifications**: Email and webhook notifications for completed reports
7. **Advanced Scheduling**: More flexible scheduling options
8. **Report Versioning**: Version control for report configurations

### API Versioning

The reporting API follows semantic versioning:

- **v1**: Current stable version
- **v2**: Planned major version with new features
- **Beta**: Experimental features and endpoints

Check the API documentation for the latest version information and migration guides.
