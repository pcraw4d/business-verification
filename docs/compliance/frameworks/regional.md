# Regional Compliance Frameworks

## Overview

The KYB Tool supports comprehensive regional compliance frameworks for data protection and privacy regulations across different jurisdictions. This documentation provides information about regional frameworks and links to detailed documentation.

## Supported Regional Frameworks

### 1. California Consumer Privacy Act (CCPA)
- **Jurisdiction**: California, United States
- **Version**: 2023
- **Effective Date**: January 1, 2023
- **Scope**: All businesses that collect personal information of California residents

### 2. Lei Geral de Proteção de Dados (LGPD)
- **Jurisdiction**: Brazil
- **Version**: 2021
- **Effective Date**: September 18, 2021
- **Scope**: All organizations processing personal data in Brazil

### 3. Personal Information Protection and Electronic Documents Act (PIPEDA)
- **Jurisdiction**: Canada
- **Version**: 2015
- **Effective Date**: June 18, 2015
- **Scope**: Private-sector organizations in Canada

### 4. Protection of Personal Information Act (POPIA)
- **Jurisdiction**: South Africa
- **Version**: 2021
- **Effective Date**: July 1, 2021
- **Scope**: All organizations processing personal information in South Africa

### 5. Personal Data Protection Act (PDPA)
- **Jurisdiction**: Singapore
- **Version**: 2021
- **Effective Date**: February 1, 2021
- **Scope**: All organizations in Singapore

### 6. Act on the Protection of Personal Information (APPI)
- **Jurisdiction**: Japan
- **Version**: 2022
- **Effective Date**: April 1, 2022
- **Scope**: All organizations handling personal information in Japan

## Detailed Documentation

For comprehensive documentation on regional compliance frameworks, including implementation details, requirements, and usage guidelines, please refer to:

**[Regional Compliance Frameworks Documentation](../../regional_compliance_frameworks.md)**

This documentation includes:

- **Framework Details**: Complete information about each regional framework
- **Implementation Guide**: How to implement and use regional frameworks
- **API Documentation**: Regional framework-specific API endpoints
- **Usage Examples**: Code examples and implementation patterns
- **Best Practices**: Industry best practices for regional compliance
- **Testing**: Comprehensive test coverage and validation
- **Troubleshooting**: Common issues and solutions

## Quick Start

### Initialize Regional Framework

```go
// Initialize regional compliance tracking
err := regionalService.InitializeRegionalTracking(
    ctx,
    businessID,
    compliance.FrameworkCCPA,
    "California, United States",
    true,  // data controller
    false, // data processor
)
```

### Get Regional Status

```go
// Get regional compliance status
status, err := regionalService.GetRegionalStatus(ctx, businessID, compliance.FrameworkCCPA)
```

### Assess Regional Compliance

```go
// Assess regional compliance
assessment, err := regionalService.AssessRegionalCompliance(ctx, businessID, compliance.FrameworkCCPA)
```

## API Endpoints

### Regional Framework Endpoints

- `POST /v1/compliance/regional/initialize` - Initialize regional compliance tracking
- `GET /v1/compliance/regional/status/{businessID}/{framework}` - Get regional compliance status
- `PUT /v1/compliance/regional/requirement/{businessID}/{framework}/{requirementID}` - Update requirement status
- `PUT /v1/compliance/regional/category/{businessID}/{framework}/{categoryID}` - Update category status
- `POST /v1/compliance/regional/assess/{businessID}/{framework}` - Assess regional compliance
- `GET /v1/compliance/regional/report/{businessID}/{framework}` - Generate compliance report
- `GET /v1/compliance/regional/frameworks` - Get supported frameworks

## Framework Selection

When choosing which regional frameworks to implement, consider:

1. **Geographic Presence**: Where does your business operate?
2. **Data Processing**: Where do you process personal data?
3. **Customer Base**: Where are your customers located?
4. **Legal Requirements**: What are your legal obligations?
5. **Business Objectives**: What are your compliance goals?

## Implementation Strategy

### Phase 1: Core Frameworks
- Start with frameworks that apply to your primary markets
- Focus on high-risk, high-impact requirements
- Establish baseline compliance

### Phase 2: Expansion
- Add frameworks for secondary markets
- Implement advanced compliance features
- Enhance monitoring and reporting

### Phase 3: Optimization
- Optimize compliance processes
- Implement automation
- Continuous improvement

## Support

For questions or issues related to regional compliance frameworks:

1. **Documentation**: Refer to the detailed regional frameworks documentation
2. **Examples**: Check the examples directory for implementation examples
3. **Community**: Join the community forum for discussions
4. **Support**: Contact support for technical assistance

---

**Last Updated**: August 2024  
**Version**: 1.0  
**Regional Frameworks Version**: 1.0
