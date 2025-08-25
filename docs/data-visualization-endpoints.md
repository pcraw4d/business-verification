# Data Visualization Endpoints

## Overview

The Data Visualization API provides comprehensive visualization capabilities for the KYB platform, allowing users to generate various types of charts, graphs, dashboards, and visual representations of business data. This API supports both immediate visualization generation and background job processing for complex visualizations.

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
  "visualization_id": "viz_1234567890_1",
  "business_id": "business_123",
  "type": "line_chart",
  "chart_type": "line",
  "status": "success",
  "is_successful": true,
  "data": { ... },
  "config": { ... },
  "metadata": { ... },
  "generated_at": "2024-12-19T10:30:00Z",
  "processing_time": "150ms"
}
```

## Supported Visualization Types

- `line_chart` - Line charts for time series data
- `bar_chart` - Bar charts for categorical data
- `pie_chart` - Pie charts for proportions
- `area_chart` - Area charts for cumulative data
- `scatter_plot` - Scatter plots for correlation analysis
- `heatmap` - Heatmaps for matrix data
- `gauge` - Gauge charts for single metrics
- `table` - Data tables with sorting and filtering
- `kpi` - Key Performance Indicators
- `dashboard` - Complete dashboard layouts
- `custom` - Custom visualization types

## Supported Chart Types

- `line` - Line charts
- `bar` - Bar charts
- `pie` - Pie charts
- `area` - Area charts
- `scatter` - Scatter plots
- `bubble` - Bubble charts
- `radar` - Radar charts
- `doughnut` - Doughnut charts
- `polar_area` - Polar area charts
- `heatmap` - Heatmaps
- `gauge` - Gauge charts
- `table` - Data tables

## Endpoints

### 1. Generate Visualization

**POST** `/v1/visualize`

Generates a visualization immediately with the provided data and configuration.

#### Request Body

```json
{
  "business_id": "business_123",
  "visualization_type": "line_chart",
  "chart_type": "line",
  "data": {
    "labels": ["Jan", "Feb", "Mar", "Apr", "May", "Jun"],
    "datasets": [
      {
        "label": "Verifications",
        "data": [65, 59, 80, 81, 56, 55]
      }
    ]
  },
  "config": {
    "title": "Monthly Verifications",
    "description": "Number of verifications per month",
    "width": 800,
    "height": 400,
    "colors": ["#36A2EB", "#FF6384"],
    "responsive": true
  },
  "filters": {
    "date_range": "last_6_months",
    "status": "completed"
  },
  "time_range": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-12-31T23:59:59Z"
  },
  "group_by": ["month", "status"],
  "aggregations": ["count", "sum"],
  "include_metadata": true,
  "include_interactivity": true,
  "theme": "light",
  "format": "json",
  "metadata": {
    "source": "verification_data",
    "version": "1.0"
  }
}
```

#### Response

```json
{
  "visualization_id": "viz_1234567890_1",
  "business_id": "business_123",
  "type": "line_chart",
  "chart_type": "line",
  "status": "success",
  "is_successful": true,
  "data": {
    "labels": ["Jan", "Feb", "Mar", "Apr", "May", "Jun"],
    "datasets": [
      {
        "label": "Verifications",
        "data": [65, 59, 80, 81, 56, 55],
        "backgroundColor": "rgba(54, 162, 235, 0.1)",
        "borderColor": "#36A2EB",
        "borderWidth": 2,
        "fill": true,
        "tension": 0.4
      }
    ]
  },
  "config": {
    "title": "Monthly Verifications",
    "description": "Number of verifications per month",
    "width": 800,
    "height": 400,
    "colors": ["#36A2EB", "#FF6384"],
    "responsive": true
  },
  "metadata": {
    "source": "verification_data",
    "version": "1.0"
  },
  "generated_at": "2024-12-19T10:30:00Z",
  "processing_time": "150ms"
}
```

### 2. Create Visualization Job

**POST** `/v1/visualize/jobs`

Creates a background job for generating complex visualizations.

#### Request Body

Same as the immediate visualization request.

#### Response

```json
{
  "job_id": "viz_job_1234567890_1",
  "business_id": "business_123",
  "type": "line_chart",
  "status": "pending",
  "progress": 0.0,
  "total_steps": 5,
  "current_step": 0,
  "step_description": "Initializing visualization job",
  "created_at": "2024-12-19T10:30:00Z",
  "metadata": {
    "source": "verification_data",
    "version": "1.0"
  }
}
```

### 3. Get Visualization Job

**GET** `/v1/visualize/jobs?job_id={job_id}`

Retrieves the status and results of a background visualization job.

#### Response

```json
{
  "job_id": "viz_job_1234567890_1",
  "business_id": "business_123",
  "type": "line_chart",
  "status": "completed",
  "progress": 1.0,
  "total_steps": 5,
  "current_step": 5,
  "step_description": "Visualization completed",
  "result": {
    "visualization_id": "viz_1234567890_1",
    "business_id": "business_123",
    "type": "line_chart",
    "chart_type": "line",
    "status": "success",
    "is_successful": true,
    "data": { ... },
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

### 4. List Visualization Jobs

**GET** `/v1/visualize/jobs`

Lists all visualization jobs with optional filtering and pagination.

#### Query Parameters

- `status` (optional): Filter by job status (pending, processing, completed, failed)
- `business_id` (optional): Filter by business ID
- `limit` (optional): Number of jobs to return (default: 50, max: 100)
- `offset` (optional): Number of jobs to skip (default: 0)

#### Response

```json
{
  "jobs": [
    {
      "job_id": "viz_job_1234567890_1",
      "business_id": "business_123",
      "type": "line_chart",
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

### 5. Get Visualization Schema

**GET** `/v1/visualize/schemas?schema_id={schema_id}`

Retrieves a pre-configured visualization schema.

#### Response

```json
{
  "id": "default_line_chart",
  "name": "Default Line Chart",
  "description": "Standard line chart for time series data",
  "type": "line_chart",
  "chart_type": "line",
  "config": {
    "title": "Time Series Data",
    "description": "Line chart showing data over time",
    "width": 800,
    "height": 400,
    "colors": ["#36A2EB", "#FF6384", "#4BC0C0"],
    "responsive": true
  },
  "data_mapping": {
    "x": "timestamp",
    "y": "value"
  },
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

### 6. List Visualization Schemas

**GET** `/v1/visualize/schemas`

Lists all available visualization schemas with optional filtering and pagination.

#### Query Parameters

- `type` (optional): Filter by visualization type
- `chart_type` (optional): Filter by chart type
- `limit` (optional): Number of schemas to return (default: 50, max: 100)
- `offset` (optional): Number of schemas to skip (default: 0)

#### Response

```json
{
  "schemas": [
    {
      "id": "default_line_chart",
      "name": "Default Line Chart",
      "description": "Standard line chart for time series data",
      "type": "line_chart",
      "chart_type": "line",
      "created_at": "2024-12-19T10:30:00Z",
      "updated_at": "2024-12-19T10:30:00Z"
    }
  ],
  "total_count": 1,
  "limit": 50,
  "offset": 0
}
```

### 7. Generate Dashboard

**POST** `/v1/visualize/dashboard`

Generates a complete dashboard with multiple widgets and layouts.

#### Request Body

```json
{
  "business_id": "business_123",
  "widgets": [
    {
      "id": "total_verifications",
      "type": "kpi",
      "title": "Total Verifications",
      "description": "Total number of verifications",
      "position": {
        "x": 0,
        "y": 0
      },
      "size": {
        "width": 3,
        "height": 2
      },
      "data": {
        "value": 125000,
        "label": "Total Verifications",
        "change": 12.5
      },
      "refresh_rate": 30
    },
    {
      "id": "verification_trend",
      "type": "line_chart",
      "title": "Verification Trend",
      "description": "Monthly verification trend",
      "position": {
        "x": 3,
        "y": 0
      },
      "size": {
        "width": 6,
        "height": 4
      },
      "data": {
        "labels": ["Jan", "Feb", "Mar", "Apr", "May", "Jun"],
        "datasets": [
          {
            "label": "Verifications",
            "data": [65, 59, 80, 81, 56, 55]
          }
        ]
      },
      "refresh_rate": 60
    }
  ],
  "layout": {
    "columns": 12,
    "rows": 8,
    "theme": "light"
  },
  "theme": "light",
  "config": {
    "title": "KYB Platform Dashboard",
    "description": "Comprehensive overview of business verification platform",
    "responsive": true
  },
  "metadata": {
    "version": "1.0",
    "created_by": "user_123"
  }
}
```

#### Response

```json
{
  "dashboard_id": "dashboard_1234567890_1",
  "title": "KYB Platform Dashboard",
  "description": "Comprehensive overview of business verification platform",
  "layout": {
    "columns": 12,
    "rows": 8,
    "widgets": [
      {
        "id": "total_verifications",
        "type": "kpi",
        "position": {"x": 0, "y": 0},
        "size": {"width": 3, "height": 2},
        "data": {
          "value": 125000,
          "label": "Total Verifications",
          "change": 12.5
        }
      },
      {
        "id": "verification_trend",
        "type": "line_chart",
        "position": {"x": 3, "y": 0},
        "size": {"width": 6, "height": 4},
        "data": {
          "labels": ["Jan", "Feb", "Mar", "Apr", "May", "Jun"],
          "datasets": [
            {
              "label": "Verifications",
              "data": [65, 59, 80, 81, 56, 55]
            }
          ]
        }
      }
    ]
  },
  "refresh_rate": 30,
  "theme": "light",
  "created_at": "2024-12-19T10:30:00Z",
  "processing_time": "200ms"
}
```

## Configuration Options

### VisualizationConfig

```json
{
  "title": "Chart Title",
  "description": "Chart description",
  "width": 800,
  "height": 400,
  "colors": ["#36A2EB", "#FF6384", "#4BC0C0"],
  "options": {
    "animation": true,
    "legend": true,
    "tooltips": true
  },
  "axis": {
    "x_axis": {
      "title": "Time",
      "type": "time",
      "grid_lines": true
    },
    "y_axis": {
      "title": "Value",
      "type": "linear",
      "grid_lines": true
    }
  },
  "legend": {
    "display": true,
    "position": "top",
    "align": "center"
  },
  "animation": {
    "enabled": true,
    "duration": 1000,
    "easing": "easeInOutQuart"
  },
  "responsive": true
}
```

### ChartData Structure

```json
{
  "labels": ["Label 1", "Label 2", "Label 3"],
  "datasets": [
    {
      "label": "Dataset 1",
      "data": [10, 20, 30],
      "backgroundColor": "#36A2EB",
      "borderColor": "#36A2EB",
      "borderWidth": 2,
      "fill": true,
      "tension": 0.4,
      "pointRadius": 6,
      "options": {
        "stepped": false,
        "spanGaps": true
      }
    }
  ]
}
```

## Error Responses

### Validation Error

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "visualization_type is required"
  },
  "timestamp": "2024-12-19T10:30:00Z"
}
```

### Job Not Found

```json
{
  "error": {
    "code": "JOB_NOT_FOUND",
    "message": "Visualization job not found"
  },
  "timestamp": "2024-12-19T10:30:00Z"
}
```

### Processing Error

```json
{
  "error": {
    "code": "VISUALIZATION_ERROR",
    "message": "Failed to generate visualization"
  },
  "timestamp": "2024-12-19T10:30:00Z"
}
```

## Integration Examples

### JavaScript/TypeScript

```javascript
// Generate a line chart
async function generateLineChart() {
  const response = await fetch('https://api.kyb-platform.com/v1/visualize', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      visualization_type: 'line_chart',
      chart_type: 'line',
      data: {
        labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
        datasets: [{
          label: 'Verifications',
          data: [65, 59, 80, 81, 56, 55]
        }]
      },
      config: {
        title: 'Monthly Verifications',
        responsive: true
      }
    })
  });

  const visualization = await response.json();
  return visualization;
}

// Create a background job
async function createVisualizationJob() {
  const response = await fetch('https://api.kyb-platform.com/v1/visualize/jobs', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      visualization_type: 'dashboard',
      data: { /* complex data */ },
      config: { /* complex config */ }
    })
  });

  const job = await response.json();
  return job;
}

// Poll job status
async function pollJobStatus(jobId) {
  const response = await fetch(`https://api.kyb-platform.com/v1/visualize/jobs?job_id=${jobId}`, {
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY'
    }
  });

  const job = await response.json();
  return job;
}

// Generate dashboard
async function generateDashboard() {
  const response = await fetch('https://api.kyb-platform.com/v1/visualize/dashboard', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      business_id: 'business_123',
      widgets: [
        {
          id: 'kpi_widget',
          type: 'kpi',
          title: 'Total Verifications',
          position: { x: 0, y: 0 },
          size: { width: 3, height: 2 },
          data: { value: 125000, label: 'Total Verifications' }
        }
      ],
      theme: 'light'
    })
  });

  const dashboard = await response.json();
  return dashboard;
}
```

### Python

```python
import requests
import json

# Generate a line chart
def generate_line_chart():
    url = 'https://api.kyb-platform.com/v1/visualize'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
    }
    
    data = {
        'visualization_type': 'line_chart',
        'chart_type': 'line',
        'data': {
            'labels': ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
            'datasets': [{
                'label': 'Verifications',
                'data': [65, 59, 80, 81, 56, 55]
            }]
        },
        'config': {
            'title': 'Monthly Verifications',
            'responsive': True
        }
    }
    
    response = requests.post(url, headers=headers, json=data)
    return response.json()

# Create a background job
def create_visualization_job():
    url = 'https://api.kyb-platform.com/v1/visualize/jobs'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
    }
    
    data = {
        'visualization_type': 'dashboard',
        'data': {},  # complex data
        'config': {}  # complex config
    }
    
    response = requests.post(url, headers=headers, json=data)
    return response.json()

# Poll job status
def poll_job_status(job_id):
    url = f'https://api.kyb-platform.com/v1/visualize/jobs?job_id={job_id}'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY'
    }
    
    response = requests.get(url, headers=headers)
    return response.json()

# Generate dashboard
def generate_dashboard():
    url = 'https://api.kyb-platform.com/v1/visualize/dashboard'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
    }
    
    data = {
        'business_id': 'business_123',
        'widgets': [
            {
                'id': 'kpi_widget',
                'type': 'kpi',
                'title': 'Total Verifications',
                'position': {'x': 0, 'y': 0},
                'size': {'width': 3, 'height': 2},
                'data': {'value': 125000, 'label': 'Total Verifications'}
            }
        ],
        'theme': 'light'
    }
    
    response = requests.post(url, headers=headers, json=data)
    return response.json()
```

### React Component

```jsx
import React, { useState, useEffect } from 'react';
import { Line } from 'react-chartjs-2';

const VisualizationComponent = () => {
  const [chartData, setChartData] = useState(null);
  const [loading, setLoading] = useState(false);

  const generateChart = async () => {
    setLoading(true);
    try {
      const response = await fetch('https://api.kyb-platform.com/v1/visualize', {
        method: 'POST',
        headers: {
          'Authorization': 'Bearer YOUR_API_KEY',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          visualization_type: 'line_chart',
          chart_type: 'line',
          data: {
            labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
            datasets: [{
              label: 'Verifications',
              data: [65, 59, 80, 81, 56, 55]
            }]
          },
          config: {
            title: 'Monthly Verifications',
            responsive: true
          }
        })
      });

      const visualization = await response.json();
      setChartData(visualization.data);
    } catch (error) {
      console.error('Error generating chart:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    generateChart();
  }, []);

  if (loading) {
    return <div>Loading chart...</div>;
  }

  if (!chartData) {
    return <div>No chart data available</div>;
  }

  return (
    <div>
      <h2>Monthly Verifications</h2>
      <Line data={chartData} />
    </div>
  );
};

export default VisualizationComponent;
```

## Best Practices

### 1. Data Preparation

- Ensure data is properly formatted and validated before sending
- Use appropriate data types for different chart types
- Include meaningful labels and descriptions
- Consider data size and performance implications

### 2. Configuration

- Use responsive configurations for better user experience
- Choose appropriate colors for accessibility
- Set meaningful titles and descriptions
- Configure animations appropriately for the use case

### 3. Background Jobs

- Use background jobs for complex visualizations
- Implement proper polling mechanisms for job status
- Handle job failures gracefully
- Consider job cleanup and retention policies

### 4. Error Handling

- Implement proper error handling for all API calls
- Display meaningful error messages to users
- Retry failed requests when appropriate
- Log errors for debugging and monitoring

### 5. Performance

- Use appropriate chart types for data size
- Consider caching for frequently accessed visualizations
- Optimize data payload size
- Use background jobs for heavy processing

### 6. Security

- Validate all input data
- Sanitize user-provided configurations
- Implement proper access controls
- Monitor for abuse and rate limiting

## Rate Limiting

- **Standard Rate Limit**: 100 requests per minute per API key
- **Background Jobs**: 10 job creations per minute per API key
- **Schema Retrieval**: 200 requests per minute per API key

## Monitoring and Observability

### Key Metrics

- **Request Rate**: Number of visualization requests per minute
- **Success Rate**: Percentage of successful visualizations
- **Processing Time**: Average time to generate visualizations
- **Error Rate**: Percentage of failed visualizations
- **Job Completion Rate**: Percentage of completed background jobs

### Health Checks

Monitor the following endpoints for system health:

```bash
# Check visualization service health
curl -X GET "https://api.kyb-platform.com/v1/health/visualization" \
  -H "Authorization: Bearer YOUR_API_KEY"

# Check background job processing
curl -X GET "https://api.kyb-platform.com/v1/health/jobs" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Logging

All visualization operations are logged with the following information:

- Request ID for correlation
- Business ID for tracking
- Visualization type and configuration
- Processing time and performance metrics
- Error details and stack traces

## Troubleshooting

### Common Issues

1. **Validation Errors**
   - Ensure all required fields are provided
   - Check data format and types
   - Verify chart type compatibility

2. **Job Failures**
   - Check job status and error messages
   - Verify data size and complexity
   - Monitor system resources

3. **Performance Issues**
   - Use background jobs for complex visualizations
   - Optimize data payload size
   - Consider caching strategies

4. **Authentication Errors**
   - Verify API key is valid and active
   - Check API key permissions
   - Ensure proper header format

### Debug Information

Enable debug logging by including the `X-Debug` header:

```bash
curl -X POST "https://api.kyb-platform.com/v1/visualize" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "X-Debug: true" \
  -H "Content-Type: application/json" \
  -d '{"visualization_type": "line_chart", ...}'
```

### Support

For additional support and troubleshooting:

- Check the API documentation for detailed endpoint information
- Review error logs and monitoring dashboards
- Contact support with request IDs and error details
- Provide reproducible examples for complex issues

## Migration Guide

### From Previous Versions

If migrating from previous visualization APIs:

1. **Update Endpoint URLs**: Use the new `/v1/visualize` endpoints
2. **Update Request Format**: Follow the new request structure
3. **Update Response Handling**: Handle the new response format
4. **Test Thoroughly**: Verify all visualizations work correctly
5. **Update Documentation**: Update client documentation and examples

### Breaking Changes

- New authentication requirements
- Updated request/response formats
- New error codes and messages
- Enhanced validation rules
- Improved performance characteristics

## Future Enhancements

### Planned Features

1. **Real-time Visualizations**: WebSocket support for live data updates
2. **Advanced Chart Types**: More specialized chart types and configurations
3. **Interactive Features**: Enhanced interactivity and user interactions
4. **Export Capabilities**: Export visualizations to various formats
5. **Templates**: Pre-built visualization templates for common use cases
6. **Collaboration**: Shared visualizations and collaborative editing
7. **Analytics**: Built-in analytics and insights for visualizations
8. **Mobile Optimization**: Optimized visualizations for mobile devices

### API Versioning

The visualization API follows semantic versioning:

- **v1**: Current stable version
- **v2**: Planned major version with new features
- **Beta**: Experimental features and endpoints

Check the API documentation for the latest version information and migration guides.
