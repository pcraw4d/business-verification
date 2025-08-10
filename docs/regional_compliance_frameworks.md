# Regional Compliance Frameworks

## Overview

The KYB Tool implements comprehensive support for regional data protection and privacy compliance frameworks. This document provides detailed information about the supported frameworks, their implementation, and usage guidelines.

## Supported Frameworks

### 1. California Consumer Privacy Act (CCPA)
- **Jurisdiction**: California, United States
- **Version**: 2023
- **Effective Date**: January 1, 2023
- **Scope**: All businesses that collect personal information of California residents
- **Key Features**:
  - Consumer rights (access, deletion, opt-out)
  - Business obligations for data handling
  - Data transparency requirements
  - Enforcement mechanisms

### 2. Lei Geral de Proteção de Dados (LGPD)
- **Jurisdiction**: Brazil
- **Version**: 2021
- **Effective Date**: September 18, 2021
- **Scope**: All organizations processing personal data in Brazil
- **Key Features**:
  - Legal basis for processing
  - Data subject rights
  - Business obligations
  - Data protection measures

### 3. Personal Information Protection and Electronic Documents Act (PIPEDA)
- **Jurisdiction**: Canada
- **Version**: 2015
- **Effective Date**: June 18, 2015
- **Scope**: Private-sector organizations in Canada
- **Key Features**:
  - Consent requirements
  - Limiting collection and use
  - Accuracy and safeguards
  - Individual access rights

### 4. Protection of Personal Information Act (POPIA)
- **Jurisdiction**: South Africa
- **Version**: 2021
- **Effective Date**: July 1, 2021
- **Scope**: All organizations processing personal information in South Africa
- **Key Features**:
  - Processing limitations
  - Purpose specification
  - Information quality
  - Security safeguards

### 5. Personal Data Protection Act (PDPA)
- **Jurisdiction**: Singapore
- **Version**: 2021
- **Effective Date**: February 1, 2021
- **Scope**: All organizations in Singapore
- **Key Features**:
  - Consent requirements
  - Purpose limitation
  - Data breach notification
  - Transfer limitations

### 6. Act on the Protection of Personal Information (APPI)
- **Jurisdiction**: Japan
- **Version**: 2022
- **Effective Date**: April 1, 2022
- **Scope**: All organizations handling personal information in Japan
- **Key Features**:
  - Purpose specification
  - Use limitation
  - Security control measures
  - Individual rights

## Implementation Details

### Framework Structure

Each regional framework is implemented with the following components:

1. **Framework Definition**: Core framework metadata and configuration
2. **Categories**: Logical groupings of requirements
3. **Requirements**: Specific compliance requirements with detailed descriptions
4. **Tracking**: Status tracking and assessment capabilities

### Data Models

#### RegionalFrameworkDefinition
```go
type RegionalFrameworkDefinition struct {
    ID              string
    Name            string
    Version         string
    Description     string
    Type            FrameworkType
    Jurisdiction    string
    GeographicScope []string
    IndustryScope   []string
    EffectiveDate   time.Time
    LastUpdated     time.Time
    NextReviewDate  time.Time
    Requirements    []RegionalRequirement
    Categories      []RegionalCategory
    MappingRules    []FrameworkMapping
    Metadata        map[string]interface{}
}
```

#### RegionalRequirement
```go
type RegionalRequirement struct {
    ID                   string
    RequirementID        string
    Framework            string
    Category             string
    Section              string
    Title                string
    Description          string
    DetailedDescription  string
    LegalBasis           []string
    DataSubjectRights    []string
    RiskLevel            ComplianceRiskLevel
    Priority             CompliancePriority
    ImplementationStatus ImplementationStatus
    EvidenceRequired     bool
    EvidenceDescription  string
    KeyControls          []string
    SubRequirements      []RegionalRequirement
    ParentRequirementID  *string
    EffectiveDate        time.Time
    LastUpdated          time.Time
    NextReviewDate       time.Time
    ReviewFrequency      string
    ComplianceOfficer    string
    Tags                 []string
    Metadata             map[string]interface{}
}
```

## Usage Examples

### Initializing Regional Compliance Tracking

```go
package main

import (
    "context"
    "github.com/pcraw4d/business-verification/internal/compliance"
    "github.com/pcraw4d/business-verification/internal/observability"
)

func main() {
    // Initialize services
    logger := observability.NewLogger()
    statusSystem := compliance.NewComplianceStatusSystem(logger)
    mappingSystem := compliance.NewFrameworkMappingSystem(logger)
    
    // Create regional tracking service
    regionalService := compliance.NewRegionalTrackingService(logger, statusSystem, mappingSystem)
    
    // Initialize tracking for a business
    ctx := context.WithValue(context.Background(), "request_id", "req-123")
    
    err := regionalService.InitializeRegionalTracking(
        ctx,
        "business-123",
        compliance.FrameworkCCPA,
        "California, United States",
        true,  // data controller
        false, // data processor
    )
    
    if err != nil {
        log.Fatal(err)
    }
}
```

### Creating Framework Instances

```go
// Create CCPA framework
ccpaFramework := compliance.NewCCPAFramework()

// Create LGPD framework
lgpdFramework := compliance.NewLGPDFramework()

// Create PIPEDA framework
pipedFramework := compliance.NewPIPEDAFramework()

// Create POPIA framework
popiaFramework := compliance.NewPOPIAFramework()

// Create PDPA framework
pdpaFramework := compliance.NewPDPAFramework()

// Create APPI framework
appiFramework := compliance.NewAPPIFramework()
```

### Converting to Regulatory Framework

```go
// Convert regional framework to regulatory framework
regional := compliance.NewCCPAFramework()
regulatory := regional.ConvertRegionalToRegulatoryFramework()

// Use regulatory framework in compliance system
// ... implementation details
```

## API Endpoints

### Regional Compliance Management

The following endpoints are available for regional compliance management:

- `POST /v1/compliance/regional/initialize` - Initialize regional compliance tracking
- `GET /v1/compliance/regional/status/{businessID}/{framework}` - Get regional compliance status
- `PUT /v1/compliance/regional/requirement/{businessID}/{framework}/{requirementID}` - Update requirement status
- `PUT /v1/compliance/regional/category/{businessID}/{framework}/{categoryID}` - Update category status
- `POST /v1/compliance/regional/assess/{businessID}/{framework}` - Assess regional compliance
- `GET /v1/compliance/regional/report/{businessID}/{framework}` - Generate compliance report
- `GET /v1/compliance/regional/frameworks` - Get supported frameworks

## Testing

### Running Tests

```bash
# Run all regional framework tests
go test ./internal/compliance -v -run TestRegionalFramework

# Run specific test
go test ./internal/compliance -v -run TestRegionalFrameworkCreation

# Run with coverage
go test ./internal/compliance -v -run TestRegionalFramework -cover
```

### Test Coverage

The regional framework implementation includes comprehensive tests covering:

- Framework creation and validation
- Requirement and category structure
- Framework conversion to regulatory format
- Timestamp validation
- Metadata handling
- Geographic scope validation

## Configuration

### Environment Variables

```bash
# Regional compliance settings
REGIONAL_COMPLIANCE_ENABLED=true
REGIONAL_FRAMEWORK_UPDATE_INTERVAL=24h
REGIONAL_ASSESSMENT_FREQUENCY=annually
REGIONAL_REVIEW_FREQUENCY=semi-annually
```

### Framework Configuration

Each framework can be configured with:

- **Update Intervals**: How often framework definitions are updated
- **Assessment Frequency**: How often compliance assessments are performed
- **Review Frequency**: How often requirements are reviewed
- **Geographic Scope**: Specific regions where the framework applies
- **Industry Scope**: Specific industries where the framework applies

## Best Practices

### Implementation Guidelines

1. **Framework Selection**: Choose frameworks based on business operations and geographic presence
2. **Role Definition**: Clearly define whether the business is a data controller, processor, or both
3. **Regular Assessment**: Conduct regular compliance assessments and reviews
4. **Evidence Management**: Maintain proper evidence for all compliance requirements
5. **Documentation**: Keep comprehensive documentation of compliance activities

### Risk Management

1. **High-Risk Requirements**: Prioritize high-risk and high-priority requirements
2. **Evidence Collection**: Ensure all required evidence is collected and maintained
3. **Review Cycles**: Establish regular review cycles for compliance status
4. **Remediation Plans**: Develop and track remediation plans for non-compliant items

### Monitoring and Reporting

1. **Status Tracking**: Monitor compliance status across all frameworks
2. **Trend Analysis**: Analyze compliance trends over time
3. **Reporting**: Generate regular compliance reports for stakeholders
4. **Alerting**: Set up alerts for compliance issues and deadlines

## Troubleshooting

### Common Issues

1. **Framework Not Found**: Ensure the framework is properly implemented and registered
2. **Invalid Jurisdiction**: Verify the jurisdiction matches the framework requirements
3. **Missing Requirements**: Check that all required fields are populated
4. **Conversion Errors**: Ensure proper data types and field mappings

### Debugging

```go
// Enable debug logging
logger.SetLevel("debug")

// Check framework status
status, err := regionalService.GetRegionalStatus(ctx, businessID, framework)
if err != nil {
    log.Printf("Error getting status: %v", err)
}

// Validate framework definition
framework := compliance.NewCCPAFramework()
if framework.ID == "" {
    log.Error("Framework ID is empty")
}
```

## Future Enhancements

### Planned Features

1. **Additional Frameworks**: Support for more regional frameworks
2. **Automated Assessment**: Automated compliance assessment capabilities
3. **Integration**: Integration with external compliance tools
4. **Reporting**: Enhanced reporting and analytics features
5. **Workflow**: Compliance workflow management

### Framework Updates

The system is designed to support framework updates and new versions:

1. **Version Management**: Track framework versions and changes
2. **Migration Support**: Support for migrating between framework versions
3. **Backward Compatibility**: Maintain compatibility with previous versions
4. **Update Notifications**: Notify users of framework updates

## Support

For questions or issues related to regional compliance frameworks:

1. **Documentation**: Refer to this documentation and API documentation
2. **Testing**: Use the provided test suite to validate implementations
3. **Logging**: Enable debug logging for troubleshooting
4. **Community**: Check the project repository for updates and discussions

## References

- [CCPA Official Website](https://oag.ca.gov/privacy/ccpa)
- [LGPD Official Website](https://www.gov.br/anpd/pt-br)
- [PIPEDA Official Website](https://www.priv.gc.ca/en/privacy-topics/privacy-laws-in-canada/the-personal-information-protection-and-electronic-documents-act-pipeda/)
- [POPIA Official Website](https://www.justice.gov.za/inforeg/docs/InfoRegSA-POPIA-act2013-004.pdf)
- [PDPA Official Website](https://www.pdpc.gov.sg/Overview-of-PDPA/The-Legislation/Personal-Data-Protection-Act)
- [APPI Official Website](https://www.ppc.go.jp/en/)
