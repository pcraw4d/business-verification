# Compliance API Reference

## Overview

The Compliance API provides comprehensive endpoints for managing compliance frameworks, assessments, monitoring, and reporting. This reference documents all available endpoints, request/response formats, and usage examples.

## Base URL

```
https://api.kybtool.com/v1/compliance
```

## Authentication

All API requests require authentication using API keys or JWT tokens:

```bash
# Using API Key
Authorization: Bearer YOUR_API_KEY

# Using JWT Token
Authorization: Bearer YOUR_JWT_TOKEN
```

## Common Response Format

All API responses follow a standard format:

```json
{
  "success": true,
  "data": {
    // Response data
  },
  "meta": {
    "request_id": "req-123456789",
    "timestamp": "2024-08-09T20:17:00Z",
    "version": "1.0"
  },
  "errors": null
}
```

## Error Response Format

Error responses include detailed error information:

```json
{
  "success": false,
  "data": null,
  "meta": {
    "request_id": "req-123456789",
    "timestamp": "2024-08-09T20:17:00Z",
    "version": "1.0"
  },
  "errors": [
    {
      "code": "VALIDATION_ERROR",
      "message": "Invalid business ID format",
      "field": "business_id",
      "details": "Business ID must be a valid UUID"
    }
  ]
}
```

## Framework Management

### List Supported Frameworks

**GET** `/frameworks`

Returns a list of all supported compliance frameworks.

#### Response

```json
{
  "success": true,
  "data": {
    "frameworks": [
      {
        "id": "SOC2",
        "name": "SOC 2",
        "version": "2017",
        "type": "Industry Standard",
        "jurisdiction": "United States",
        "description": "Service Organization Control 2 for security, availability, processing integrity, confidentiality, and privacy",
        "categories": ["Security", "Availability", "Processing Integrity", "Confidentiality", "Privacy"]
      },
      {
        "id": "PCI-DSS",
        "name": "PCI DSS",
        "version": "4.0",
        "type": "Industry Standard",
        "jurisdiction": "Global",
        "description": "Payment Card Industry Data Security Standard",
        "categories": ["Build and Maintain a Secure Network", "Protect Account Data", "Maintain a Vulnerability Management Program", "Implement Strong Access Control Measures", "Regularly Monitor and Test Networks", "Maintain an Information Security Policy"]
      },
      {
        "id": "GDPR",
        "name": "GDPR",
        "version": "2018",
        "type": "Privacy Regulation",
        "jurisdiction": "European Union",
        "description": "General Data Protection Regulation",
        "categories": ["Lawfulness, Fairness, and Transparency", "Purpose Limitation", "Data Minimization", "Accuracy", "Storage Limitation", "Integrity and Confidentiality", "Accountability"]
      }
    ]
  }
}
```

### Get Framework Details

**GET** `/frameworks/{framework_id}`

Returns detailed information about a specific framework.

#### Parameters

- `framework_id` (string, required): The framework identifier (e.g., "SOC2", "PCI-DSS", "GDPR")

#### Response

```json
{
  "success": true,
  "data": {
    "framework": {
      "id": "SOC2",
      "name": "SOC 2",
      "version": "2017",
      "type": "Industry Standard",
      "jurisdiction": "United States",
      "description": "Service Organization Control 2 for security, availability, processing integrity, confidentiality, and privacy",
      "effective_date": "2017-01-01T00:00:00Z",
      "last_updated": "2024-08-09T20:17:00Z",
      "categories": [
        {
          "id": "Security",
          "name": "Security",
          "description": "Information and systems are protected against unauthorized access",
          "requirements_count": 25
        }
      ],
      "requirements": [
        {
          "id": "SOC2-SEC-001",
          "title": "Access Control Policy",
          "description": "Establish and maintain access control policies",
          "category": "Security",
          "risk_level": "high",
          "priority": "high"
        }
      ]
    }
  }
}
```

## Compliance Status Management

### Initialize Framework

**POST** `/initialize`

Initialize compliance tracking for a specific framework.

#### Request Body

```json
{
  "business_id": "business-123",
  "framework": "SOC2",
  "version": "2017",
  "type": "SOC 2 Type II",
  "categories": ["Security", "Availability", "Processing Integrity", "Confidentiality", "Privacy"],
  "compliance_officer": "john.doe@acme.com",
  "assessment_frequency": "quarterly"
}
```

#### Response

```json
{
  "success": true,
  "data": {
    "compliance_status": {
      "business_id": "business-123",
      "framework": "SOC2",
      "version": "2017",
      "type": "SOC 2 Type II",
      "overall_status": "not_started",
      "compliance_score": 0.0,
      "last_assessment": null,
      "next_assessment": "2024-11-09T20:17:00Z",
      "compliance_officer": "john.doe@acme.com"
    }
  }
}
```

### Get Compliance Status

**GET** `/status/{business_id}/{framework}`

Get the current compliance status for a specific business and framework.

#### Parameters

- `business_id` (string, required): The business identifier
- `framework` (string, required): The framework identifier

#### Response

```json
{
  "success": true,
  "data": {
    "compliance_status": {
      "business_id": "business-123",
      "framework": "SOC2",
      "version": "2017",
      "type": "SOC 2 Type II",
      "overall_status": "in_progress",
      "compliance_score": 65.5,
      "category_status": {
        "Security": {
          "status": "compliant",
          "score": 85.0,
          "requirement_count": 25,
          "implemented_count": 22
        },
        "Availability": {
          "status": "in_progress",
          "score": 45.0,
          "requirement_count": 15,
          "implemented_count": 7
        }
      },
      "last_assessment": "2024-08-01T10:00:00Z",
      "next_assessment": "2024-11-01T10:00:00Z"
    }
  }
}
```

### Update Requirement Status

**PUT** `/requirement/{business_id}/{framework}/{requirement_id}`

Update the status of a specific requirement.

#### Parameters

- `business_id` (string, required): The business identifier
- `framework` (string, required): The framework identifier
- `requirement_id` (string, required): The requirement identifier

#### Request Body

```json
{
  "status": "compliant",
  "implementation_status": "implemented",
  "compliance_score": 100.0,
  "reviewer": "john.doe@acme.com",
  "notes": "Access control policy implemented and tested",
  "evidence": [
    {
      "title": "Access Control Policy",
      "type": "policy_document",
      "file_path": "/documents/access-control-policy.pdf"
    }
  ]
}
```

#### Response

```json
{
  "success": true,
  "data": {
    "requirement_status": {
      "requirement_id": "SOC2-SEC-001",
      "status": "compliant",
      "implementation_status": "implemented",
      "compliance_score": 100.0,
      "last_reviewed": "2024-08-09T20:17:00Z",
      "reviewer": "john.doe@acme.com",
      "evidence_count": 1
    }
  }
}
```

## Assessment Management

### Run Assessment

**POST** `/assess/{business_id}`

Run a compliance assessment for a business.

#### Parameters

- `business_id` (string, required): The business identifier

#### Request Body

```json
{
  "framework": "SOC2",
  "assessment_type": "comprehensive",
  "assessor": "john.doe@acme.com",
  "include_evidence": true,
  "generate_report": true
}
```

#### Response

```json
{
  "success": true,
  "data": {
    "assessment": {
      "id": "assessment-123",
      "business_id": "business-123",
      "framework": "SOC2",
      "assessment_type": "comprehensive",
      "status": "completed",
      "compliance_score": 75.5,
      "gaps_count": 12,
      "recommendations_count": 8,
      "completed_at": "2024-08-09T20:17:00Z",
      "assessor": "john.doe@acme.com"
    }
  }
}
```

### Get Assessment Results

**GET** `/assess/{business_id}/{assessment_id}`

Get detailed results of a specific assessment.

#### Parameters

- `business_id` (string, required): The business identifier
- `assessment_id` (string, required): The assessment identifier

#### Response

```json
{
  "success": true,
  "data": {
    "assessment": {
      "id": "assessment-123",
      "business_id": "business-123",
      "framework": "SOC2",
      "assessment_type": "comprehensive",
      "status": "completed",
      "compliance_score": 75.5,
      "category_results": [
        {
          "category": "Security",
          "status": "compliant",
          "score": 85.0,
          "requirement_count": 25,
          "implemented_count": 22,
          "gaps": [
            {
              "requirement_id": "SOC2-SEC-003",
              "title": "Change Management",
              "description": "Change management process not fully documented",
              "risk_level": "medium",
              "recommendation": "Document change management procedures"
            }
          ]
        }
      ],
      "gaps": [
        {
          "requirement_id": "SOC2-SEC-003",
          "title": "Change Management",
          "category": "Security",
          "risk_level": "medium",
          "priority": "medium",
          "description": "Change management process not fully documented",
          "recommendation": "Document change management procedures"
        }
      ],
      "recommendations": [
        {
          "id": "rec-001",
          "title": "Document Change Management Procedures",
          "description": "Create comprehensive change management documentation",
          "priority": "medium",
          "estimated_effort": "2 weeks",
          "assigned_to": "jane.smith@acme.com"
        }
      ]
    }
  }
}
```

## Gap Analysis

### Analyze Gaps

**POST** `/gaps/analyze/{business_id}`

Perform gap analysis for a business.

#### Parameters

- `business_id` (string, required): The business identifier

#### Request Body

```json
{
  "framework": "SOC2",
  "include_recommendations": true,
  "prioritize_by_risk": true
}
```

#### Response

```json
{
  "success": true,
  "data": {
    "gap_analysis": {
      "business_id": "business-123",
      "framework": "SOC2",
      "analysis_date": "2024-08-09T20:17:00Z",
      "overall_compliance_score": 75.5,
      "gaps": [
        {
          "requirement_id": "SOC2-SEC-003",
          "title": "Change Management",
          "category": "Security",
          "risk_level": "medium",
          "priority": "medium",
          "description": "Change management process not fully documented",
          "current_status": "not_implemented",
          "target_status": "compliant",
          "recommendation": "Document change management procedures",
          "estimated_effort": "2 weeks",
          "estimated_cost": "$5000"
        }
      ],
      "summary": {
        "total_requirements": 100,
        "compliant_requirements": 75,
        "non_compliant_requirements": 15,
        "in_progress_requirements": 10,
        "high_risk_gaps": 3,
        "medium_risk_gaps": 8,
        "low_risk_gaps": 4
      }
    }
  }
}
```

## Reporting

### Generate Report

**POST** `/reports/generate/{business_id}`

Generate a compliance report.

#### Parameters

- `business_id` (string, required): The business identifier

#### Request Body

```json
{
  "framework": "SOC2",
  "report_type": "comprehensive",
  "format": "pdf",
  "include_evidence": true,
  "include_recommendations": true,
  "date_range": {
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-08-09T20:17:00Z"
  }
}
```

#### Response

```json
{
  "success": true,
  "data": {
    "report": {
      "id": "report-123",
      "business_id": "business-123",
      "framework": "SOC2",
      "report_type": "comprehensive",
      "format": "pdf",
      "status": "generated",
      "file_url": "https://api.kybtool.com/reports/report-123.pdf",
      "generated_at": "2024-08-09T20:17:00Z",
      "file_size": "2.5MB",
      "expires_at": "2024-09-09T20:17:00Z"
    }
  }
}
```

### Get Report Status

**GET** `/reports/{report_id}`

Get the status of a report generation request.

#### Parameters

- `report_id` (string, required): The report identifier

#### Response

```json
{
  "success": true,
  "data": {
    "report": {
      "id": "report-123",
      "business_id": "business-123",
      "framework": "SOC2",
      "report_type": "comprehensive",
      "format": "pdf",
      "status": "generated",
      "file_url": "https://api.kybtool.com/reports/report-123.pdf",
      "generated_at": "2024-08-09T20:17:00Z",
      "file_size": "2.5MB",
      "expires_at": "2024-09-09T20:17:00Z"
    }
  }
}
```

## Monitoring and Alerts

### Set Up Monitoring

**POST** `/monitoring/setup/{business_id}`

Set up compliance monitoring for a business.

#### Parameters

- `business_id` (string, required): The business identifier

#### Request Body

```json
{
  "framework": "SOC2",
  "monitoring_frequency": "daily",
  "alert_threshold": 80.0,
  "recipients": ["john.doe@acme.com", "jane.smith@acme.com"],
  "enabled": true
}
```

#### Response

```json
{
  "success": true,
  "data": {
    "monitoring": {
      "business_id": "business-123",
      "framework": "SOC2",
      "monitoring_frequency": "daily",
      "alert_threshold": 80.0,
      "recipients": ["john.doe@acme.com", "jane.smith@acme.com"],
      "enabled": true,
      "last_check": "2024-08-09T20:17:00Z",
      "next_check": "2024-08-10T20:17:00Z"
    }
  }
}
```

### Get Alerts

**GET** `/alerts/{business_id}`

Get compliance alerts for a business.

#### Parameters

- `business_id` (string, required): The business identifier

#### Query Parameters

- `status` (string, optional): Filter by alert status (active, resolved)
- `severity` (string, optional): Filter by severity (high, medium, low)
- `framework` (string, optional): Filter by framework
- `limit` (integer, optional): Number of alerts to return (default: 50)
- `offset` (integer, optional): Number of alerts to skip (default: 0)

#### Response

```json
{
  "success": true,
  "data": {
    "alerts": [
      {
        "id": "alert-123",
        "business_id": "business-123",
        "framework": "SOC2",
        "alert_type": "compliance_score_drop",
        "severity": "high",
        "title": "Compliance Score Dropped Below Threshold",
        "description": "SOC 2 compliance score dropped to 75.5% (below 80% threshold)",
        "status": "active",
        "created_at": "2024-08-09T20:17:00Z",
        "resolved_at": null
      }
    ],
    "meta": {
      "total": 5,
      "active": 3,
      "resolved": 2
    }
  }
}
```

## Data Models

### ComplianceStatus

```json
{
  "business_id": "string",
  "framework": "string",
  "version": "string",
  "type": "string",
  "overall_status": "string",
  "compliance_score": "number",
  "category_status": "object",
  "requirements_status": "object",
  "last_assessment": "string",
  "next_assessment": "string",
  "compliance_officer": "string"
}
```

### RequirementStatus

```json
{
  "requirement_id": "string",
  "status": "string",
  "implementation_status": "string",
  "compliance_score": "number",
  "last_reviewed": "string",
  "reviewer": "string",
  "evidence_count": "number",
  "notes": "string"
}
```

### Gap

```json
{
  "requirement_id": "string",
  "title": "string",
  "category": "string",
  "risk_level": "string",
  "priority": "string",
  "description": "string",
  "current_status": "string",
  "target_status": "string",
  "recommendation": "string",
  "estimated_effort": "string",
  "estimated_cost": "string"
}
```

### Alert

```json
{
  "id": "string",
  "business_id": "string",
  "framework": "string",
  "alert_type": "string",
  "severity": "string",
  "title": "string",
  "description": "string",
  "status": "string",
  "created_at": "string",
  "resolved_at": "string"
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Request validation failed |
| `BUSINESS_NOT_FOUND` | Business not found |
| `FRAMEWORK_NOT_FOUND` | Framework not found |
| `REQUIREMENT_NOT_FOUND` | Requirement not found |
| `ASSESSMENT_NOT_FOUND` | Assessment not found |
| `REPORT_NOT_FOUND` | Report not found |
| `UNAUTHORIZED` | Authentication required |
| `FORBIDDEN` | Insufficient permissions |
| `RATE_LIMIT_EXCEEDED` | Rate limit exceeded |
| `INTERNAL_ERROR` | Internal server error |

## Rate Limits

- **Standard Plan**: 1000 requests per hour
- **Professional Plan**: 5000 requests per hour
- **Enterprise Plan**: 20000 requests per hour

## Pagination

For endpoints that return lists, pagination is supported using `limit` and `offset` query parameters:

```
GET /compliance/status?limit=50&offset=100
```

Response includes pagination metadata:

```json
{
  "success": true,
  "data": {
    "items": [...],
    "meta": {
      "total": 500,
      "limit": 50,
      "offset": 100,
      "has_more": true
    }
  }
}
```

## SDKs and Libraries

Official SDKs are available for:

- [Go SDK](https://github.com/pcraw4d/kyb-tool-go-sdk)
- [Python SDK](https://github.com/pcraw4d/kyb-tool-python-sdk)
- [JavaScript SDK](https://github.com/pcraw4d/kyb-tool-js-sdk)

## Support

For API support:

1. **Documentation**: Refer to this API reference
2. **Examples**: Check the examples directory
3. **Community**: Join the community forum
4. **Support**: Contact support for technical assistance

---

**Last Updated**: August 2024  
**Version**: 1.0  
**API Version**: 1.0
