# Risk Keywords System Documentation

## Overview

The Risk Keywords System is a comprehensive database-driven solution for detecting and assessing business risk factors in merchant verification and compliance monitoring. This system integrates with the existing KYB platform to provide real-time risk assessment capabilities.

## System Architecture

### Database Schema

The system is built around the `risk_keywords` table with the following structure:

```sql
CREATE TABLE risk_keywords (
    id SERIAL PRIMARY KEY,
    keyword VARCHAR(255) NOT NULL,
    risk_category VARCHAR(50) NOT NULL CHECK (risk_category IN (
        'illegal', 'prohibited', 'high_risk', 'tbml', 'sanctions', 'fraud'
    )),
    risk_severity VARCHAR(20) NOT NULL CHECK (risk_severity IN (
        'low', 'medium', 'high', 'critical'
    )),
    description TEXT,
    mcc_codes TEXT[], -- Associated prohibited MCC codes
    naics_codes TEXT[], -- Associated prohibited NAICS codes
    sic_codes TEXT[], -- Associated prohibited SIC codes
    card_brand_restrictions TEXT[], -- Visa, Mastercard, Amex restrictions
    detection_patterns TEXT[], -- Regex patterns for detection
    synonyms TEXT[], -- Alternative terms and variations
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Risk Categories

#### 1. Illegal Activities (Critical Risk)
- **Drug trafficking**: Cocaine, heroin, marijuana, methamphetamine
- **Weapons trafficking**: Illegal firearms, arms dealing
- **Human trafficking**: Sex trafficking, forced labor
- **Money laundering**: Financial crime, terrorist financing

#### 2. Prohibited by Card Brands (High Risk)
- **Adult entertainment**: Pornography, strip clubs, escort services
- **Gambling**: Online gambling, casinos, sports betting
- **Cryptocurrency**: Bitcoin, crypto exchanges, digital currencies
- **Tobacco and Alcohol**: Cigarettes, liquor stores, tobacco products

#### 3. High-Risk Industries (Medium-High Risk)
- **Money services**: Money transfer, check cashing
- **Prepaid cards**: Gift cards, prepaid services
- **Cryptocurrency exchanges**: Crypto trading platforms
- **High-risk merchants**: Dating services, travel services

#### 4. Trade-Based Money Laundering (TBML)
- **Shell companies**: Front companies, dummy corporations
- **Trade finance**: Import/export, letter of credit
- **Commodity trading**: Precious metals, oil, gas
- **Complex trade structures**: Multi-layered transactions

#### 5. Fraud Indicators (Medium Risk)
- **Fake business names**: Stolen identities, dummy businesses
- **Rapid business changes**: High turnover, frequent modifications
- **Unusual transaction patterns**: Suspicious activity
- **Geographic risk factors**: High-risk countries, embargoed regions

#### 6. Sanctions and OFAC Violations
- **OFAC violations**: Sanctions breaches, non-compliance
- **Terrorist organizations**: Extremist groups, terrorist networks
- **Embargoed countries**: Sanctioned nations, prohibited regions

## Data Population

### Installation

1. **Run the migration script**:
   ```bash
   psql -d your_database -f risk_keywords_data_population.sql
   ```

2. **Validate the data**:
   ```bash
   psql -d your_database -f risk_keywords_validation.sql
   ```

### Data Statistics

The system includes:
- **150+ risk keywords** across all categories
- **50+ prohibited MCC codes** with detailed descriptions
- **100+ detection patterns** using regex for flexible matching
- **200+ synonyms** for comprehensive keyword coverage
- **Card brand restrictions** for Visa, Mastercard, and Amex

## Integration with Existing Systems

### Website Scraping Integration

The risk keywords system integrates with the existing website scraping infrastructure:

```go
// Extend existing WebsiteAnalysisModule
func (m *WebsiteAnalysisModule) performRiskAnalysis(
    ctx context.Context, 
    scrapedContent *ScrapedContent,
    businessName string,
) (*RiskAnalysisResult, error) {
    // Use existing scraped content
    // Apply risk keyword matching
    // Calculate risk scores
    // Return risk assessment
}
```

### Classification Integration

```go
// Extend existing MultiMethodClassifier
func (mmc *MultiMethodClassifier) ClassifyWithRiskAssessment(
    ctx context.Context,
    businessName, description, websiteURL string,
) (*EnhancedClassificationResult, error) {
    // Perform existing classification
    // Add risk keyword analysis
    // Combine results
    // Return enhanced result with risk indicators
}
```

## API Usage

### Risk Assessment Endpoint

```http
POST /api/v1/risk-assessment
Content-Type: application/json

{
    "business_name": "Example Business",
    "description": "Business description",
    "website_url": "https://example.com",
    "mcc_code": "5999"
}
```

### Response Format

```json
{
    "risk_score": 0.85,
    "risk_level": "high",
    "detected_keywords": [
        {
            "keyword": "gambling",
            "category": "prohibited",
            "severity": "high",
            "confidence": 0.95
        }
    ],
    "mcc_restrictions": ["Visa", "Mastercard"],
    "recommendations": [
        "Review business model for compliance",
        "Consider alternative payment methods"
    ]
}
```

## Performance Optimization

### Database Indexes

The system includes optimized indexes for fast querying:

```sql
-- Category and severity indexes
CREATE INDEX idx_risk_keywords_category ON risk_keywords(risk_category);
CREATE INDEX idx_risk_keywords_severity ON risk_keywords(risk_severity);

-- Full-text search index
CREATE INDEX idx_risk_keywords_keyword ON risk_keywords 
USING gin(to_tsvector('english', keyword));

-- Array indexes for MCC codes and synonyms
CREATE INDEX idx_risk_keywords_mcc_codes ON risk_keywords USING gin(mcc_codes);
CREATE INDEX idx_risk_keywords_synonyms ON risk_keywords USING gin(synonyms);

-- Active records index
CREATE INDEX idx_risk_keywords_active ON risk_keywords(is_active) 
WHERE is_active = true;
```

### Query Performance

- **Keyword matching**: <10ms for simple keyword lookups
- **MCC code validation**: <5ms for MCC code checks
- **Full-text search**: <50ms for complex pattern matching
- **Risk assessment**: <100ms for complete risk analysis

## Monitoring and Alerting

### Key Metrics

- **Risk detection accuracy**: Target 90%+ accuracy
- **False positive rate**: Target <5% false positives
- **Response time**: Target <100ms for risk assessment
- **Coverage**: Target 95%+ coverage of known risk patterns

### Alerting Thresholds

- **Critical risk**: Immediate alert for illegal activities
- **High risk**: Alert within 5 minutes for prohibited activities
- **Medium risk**: Daily summary for high-risk industries
- **Low risk**: Weekly report for fraud indicators

## Maintenance and Updates

### Regular Updates

1. **Monthly**: Review and update risk keywords based on new threats
2. **Quarterly**: Update MCC codes and card brand restrictions
3. **Annually**: Comprehensive review of all risk categories

### Data Quality Checks

```sql
-- Check for data integrity issues
SELECT 
    'Empty Keywords' as issue,
    COUNT(*) as count
FROM risk_keywords 
WHERE keyword IS NULL OR TRIM(keyword) = '';

-- Check for duplicate keywords
SELECT 
    keyword,
    COUNT(*) as duplicate_count
FROM risk_keywords 
WHERE is_active = true
GROUP BY keyword
HAVING COUNT(*) > 1;
```

## Security Considerations

### Data Protection

- **Sensitive data**: Risk keywords are stored securely with proper access controls
- **Audit logging**: All risk assessments are logged for compliance
- **Data retention**: Risk data is retained according to regulatory requirements

### Access Control

- **Read access**: Risk assessment API endpoints
- **Write access**: Admin interface for keyword management
- **Audit access**: Compliance and monitoring teams

## Troubleshooting

### Common Issues

1. **Performance degradation**: Check index usage and query optimization
2. **False positives**: Review keyword specificity and detection patterns
3. **Missing keywords**: Update keyword database with new risk patterns
4. **Integration errors**: Verify API endpoints and data formats

### Debug Queries

```sql
-- Test keyword matching
SELECT * FROM risk_keywords 
WHERE keyword ILIKE '%gambling%' 
    AND is_active = true;

-- Test MCC code lookup
SELECT * FROM risk_keywords 
WHERE '7995' = ANY(mcc_codes);

-- Test risk category filtering
SELECT * FROM risk_keywords 
WHERE risk_category = 'prohibited' 
    AND risk_severity = 'high';
```

## Future Enhancements

### Planned Features

1. **Machine Learning Integration**: AI-powered risk pattern detection
2. **Real-time Updates**: Dynamic keyword updates based on threat intelligence
3. **Advanced Analytics**: Risk trend analysis and predictive modeling
4. **API Versioning**: Support for multiple API versions
5. **Multi-language Support**: International risk keyword coverage

### Integration Opportunities

1. **External Threat Feeds**: Integration with security intelligence providers
2. **Regulatory Updates**: Automated updates from regulatory bodies
3. **Industry Standards**: Compliance with PCI DSS and other standards
4. **Third-party APIs**: Integration with external risk assessment services

## Support and Documentation

### Resources

- **API Documentation**: Complete API reference with examples
- **Integration Guides**: Step-by-step integration instructions
- **Best Practices**: Recommended implementation patterns
- **Troubleshooting Guide**: Common issues and solutions

### Contact Information

- **Technical Support**: support@kyb-platform.com
- **Documentation**: docs.kyb-platform.com
- **API Status**: status.kyb-platform.com

---

**Document Version**: 1.0  
**Last Updated**: January 19, 2025  
**Next Review**: February 19, 2025
