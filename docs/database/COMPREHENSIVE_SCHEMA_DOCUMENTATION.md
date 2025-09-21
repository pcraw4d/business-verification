# KYB Platform - Comprehensive Database Schema Documentation

## 📋 **Document Overview**

**Document Version**: 1.0  
**Created**: January 19, 2025  
**Last Updated**: January 19, 2025  
**Purpose**: Comprehensive documentation of the KYB Platform Supabase database schema

This document provides complete documentation of all database tables, relationships, constraints, and data flow patterns implemented in the KYB Platform's enhanced Supabase database schema.

---

## 🗄️ **Database Schema Overview**

### **Schema Architecture**
The KYB Platform database is built on PostgreSQL (via Supabase) with the following architectural principles:

- **Modular Design**: Tables are organized by functional domains
- **Referential Integrity**: Comprehensive foreign key relationships
- **Performance Optimization**: Strategic indexing and query optimization
- **Security**: Row Level Security (RLS) policies
- **Extensibility**: JSONB fields for flexible metadata storage
- **Audit Trail**: Comprehensive timestamp and version tracking

### **Schema Domains**
1. **User Management**: Authentication, authorization, and user profiles
2. **Business Management**: Merchant/business entity management
3. **Classification System**: Industry classification and keyword matching
4. **Risk Management**: Risk assessment and keyword detection
5. **Performance Monitoring**: System performance and accuracy tracking
6. **Compliance & Audit**: Compliance tracking and audit logging

---

## 👥 **User Management Tables**

### **1. users (Consolidated User Table)**

**Purpose**: Centralized user management with comprehensive profile and authentication data.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE,
    password_hash VARCHAR(255),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    full_name VARCHAR(255),
    name VARCHAR(255),
    company VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN (
        'user', 'admin', 'compliance_officer', 'risk_manager', 
        'business_analyst', 'developer', 'other'
    )),
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN (
        'active', 'inactive', 'suspended', 'pending_verification'
    )),
    is_active BOOLEAN DEFAULT TRUE,
    email_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Key Features**:
- **Comprehensive User Data**: Supports both individual and company users
- **Role-Based Access Control**: Predefined roles with specific permissions
- **Security Features**: Account locking, failed login tracking
- **Flexible Metadata**: JSONB field for extensible user attributes
- **Audit Trail**: Created/updated timestamps with automatic triggers

**Indexes**:
- Primary key on `id`
- Unique index on `email`
- Unique index on `username`
- Index on `role` for role-based queries
- Index on `status` for active user filtering
- Index on `is_active` for performance

### **2. api_keys**

**Purpose**: API key management for programmatic access to the platform.

```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    permissions JSONB DEFAULT '[]',
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE
);
```

**Key Features**:
- **Secure Key Storage**: Hashed API keys for security
- **Permission Management**: JSONB array for granular permissions
- **Expiration Support**: Optional key expiration dates
- **Usage Tracking**: Last used timestamp for monitoring
- **Cascade Deletion**: Keys deleted when user is deleted

---

## 🏢 **Business Management Tables**

### **3. merchants (Consolidated Business Table)**

**Purpose**: Comprehensive business/merchant entity management with enhanced data model.

```sql
CREATE TABLE merchants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    legal_name VARCHAR(255) NOT NULL,
    registration_number VARCHAR(100) UNIQUE NOT NULL,
    tax_id VARCHAR(100),
    industry VARCHAR(100),
    industry_code VARCHAR(20),
    business_type VARCHAR(50),
    founded_date DATE,
    employee_count INTEGER,
    annual_revenue DECIMAL(15,2),
    
    -- Address Information
    address_street1 VARCHAR(255),
    address_street2 VARCHAR(255),
    address_city VARCHAR(100),
    address_state VARCHAR(100),
    address_postal_code VARCHAR(20),
    address_country VARCHAR(100),
    address_country_code VARCHAR(10),
    
    -- Contact Information
    contact_phone VARCHAR(50),
    contact_email VARCHAR(255),
    contact_website VARCHAR(255),
    contact_primary_contact VARCHAR(255),
    
    -- Business Relationships
    portfolio_type_id UUID,
    risk_level_id UUID,
    
    -- Status and Compliance
    compliance_status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (compliance_status IN (
        'pending', 'approved', 'rejected', 'under_review', 'suspended'
    )),
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN (
        'active', 'inactive', 'suspended', 'pending_verification'
    )),
    
    -- Audit Fields
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}'
);
```

**Key Features**:
- **Comprehensive Business Data**: Legal name, registration, tax ID, financial data
- **Complete Address Model**: Structured address fields for all countries
- **Contact Management**: Multiple contact methods and primary contact
- **Compliance Tracking**: Status tracking for compliance workflows
- **Audit Trail**: Created by user tracking and timestamps
- **Flexible Metadata**: JSONB for extensible business attributes

**Indexes**:
- Primary key on `id`
- Unique index on `registration_number`
- Index on `industry` for industry-based queries
- Index on `compliance_status` for compliance filtering
- Index on `status` for active merchant filtering
- Index on `created_by` for user-based queries

---

## 🏭 **Classification System Tables**

### **4. industries**

**Purpose**: Master table of industry classifications with confidence thresholds.

```sql
CREATE TABLE industries (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    category VARCHAR(100),
    confidence_threshold DECIMAL(3,2) DEFAULT 0.50,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Key Features**:
- **Industry Master Data**: Comprehensive industry taxonomy
- **Confidence Thresholds**: Configurable confidence levels per industry
- **Categorization**: Industry grouping for better organization
- **Active Status**: Soft deletion support

### **5. industry_keywords**

**Purpose**: Keywords associated with industries for classification matching.

```sql
CREATE TABLE industry_keywords (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    keyword VARCHAR(255) NOT NULL,
    weight DECIMAL(5,4) DEFAULT 1.0000,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(industry_id, keyword)
);
```

**Key Features**:
- **Weighted Keywords**: Configurable keyword importance
- **Cascade Deletion**: Keywords deleted when industry is deleted
- **Unique Constraints**: Prevents duplicate keywords per industry
- **Active Status**: Soft deletion support

### **6. classification_codes**

**Purpose**: Industry classification codes (NAICS, SIC, MCC) mapping.

```sql
CREATE TABLE classification_codes (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    code_type VARCHAR(10) NOT NULL CHECK (code_type IN ('NAICS', 'SIC', 'MCC')),
    code VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(industry_id, code_type, code)
);
```

**Key Features**:
- **Multi-Standard Support**: NAICS, SIC, and MCC codes
- **Code Validation**: Check constraints for valid code types
- **Unique Mapping**: Prevents duplicate code assignments
- **Cascade Deletion**: Codes deleted when industry is deleted

### **7. industry_patterns**

**Purpose**: Advanced pattern matching for industry classification.

```sql
CREATE TABLE industry_patterns (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    pattern VARCHAR(500) NOT NULL,
    pattern_type VARCHAR(50) NOT NULL DEFAULT 'phrase',
    confidence_score DECIMAL(3,2) DEFAULT 0.50,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Key Features**:
- **Pattern Matching**: Advanced text pattern recognition
- **Pattern Types**: Support for different pattern matching algorithms
- **Confidence Scoring**: Individual pattern confidence levels
- **Cascade Deletion**: Patterns deleted when industry is deleted

### **8. keyword_weights**

**Purpose**: Dynamic keyword weighting based on usage and success metrics.

```sql
CREATE TABLE keyword_weights (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    keyword VARCHAR(255) NOT NULL,
    base_weight DECIMAL(5,4) DEFAULT 1.0000,
    usage_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(industry_id, keyword)
);
```

**Key Features**:
- **Dynamic Weighting**: Weights adjusted based on performance
- **Usage Tracking**: Count of keyword usage
- **Success Metrics**: Success rate tracking for ML optimization
- **Performance Optimization**: Automatic weight adjustment

---

## ⚠️ **Risk Management Tables**

### **9. risk_keywords**

**Purpose**: Comprehensive risk keyword database for detecting prohibited and high-risk activities.

```sql
CREATE TABLE risk_keywords (
    id SERIAL PRIMARY KEY,
    keyword VARCHAR(255) NOT NULL,
    risk_category VARCHAR(50) NOT NULL CHECK (risk_category IN (
        'illegal', 'prohibited', 'high_risk', 'tbml', 'sanctions', 'fraud', 'regulatory'
    )),
    risk_severity VARCHAR(20) NOT NULL CHECK (risk_severity IN (
        'low', 'medium', 'high', 'critical'
    )),
    description TEXT,
    mcc_codes TEXT[],
    naics_codes TEXT[],
    sic_codes TEXT[],
    card_brand_restrictions TEXT[],
    detection_patterns TEXT[],
    synonyms TEXT[],
    risk_score_weight DECIMAL(3,2) DEFAULT 1.00 CHECK (risk_score_weight >= 0.00 AND risk_score_weight <= 2.00),
    detection_confidence DECIMAL(3,2) DEFAULT 0.80 CHECK (detection_confidence >= 0.00 AND detection_confidence <= 1.00),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(keyword) WHERE is_active = true
);
```

**Key Features**:
- **Comprehensive Risk Categories**: Illegal, prohibited, high-risk, TBML, sanctions, fraud, regulatory
- **Severity Levels**: Low, medium, high, critical risk classification
- **Code Integration**: MCC, NAICS, SIC code associations
- **Card Brand Restrictions**: Visa, Mastercard, Amex specific restrictions
- **Pattern Matching**: Regex patterns for advanced detection
- **Weighted Scoring**: Configurable risk score weights
- **Confidence Metrics**: Detection confidence levels
- **Array Fields**: GIN indexes for efficient array queries

**Indexes**:
- Primary key on `id`
- Unique index on `keyword` (active only)
- Index on `risk_category` for category filtering
- Index on `risk_severity` for severity filtering
- Index on `is_active` for active keyword filtering
- Composite indexes for category/severity combinations
- GIN indexes on all array fields (mcc_codes, naics_codes, etc.)
- Full-text search index for keyword and description

### **10. business_risk_assessments**

**Purpose**: Comprehensive risk assessment results for businesses with metadata and confidence scoring.

```sql
CREATE TABLE business_risk_assessments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_id UUID NOT NULL,
    risk_keyword_id INTEGER REFERENCES risk_keywords(id) ON DELETE SET NULL,
    detected_keywords TEXT[],
    risk_score DECIMAL(3,2) NOT NULL CHECK (risk_score >= 0.00 AND risk_score <= 1.00),
    risk_level VARCHAR(20) NOT NULL CHECK (risk_level IN (
        'low', 'medium', 'high', 'critical'
    )),
    assessment_method VARCHAR(100),
    website_content TEXT,
    detected_patterns JSONB,
    assessment_metadata JSONB,
    confidence_score DECIMAL(3,2) DEFAULT 0.80 CHECK (confidence_score >= 0.00 AND confidence_score <= 1.00),
    assessment_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Key Features**:
- **Comprehensive Risk Scoring**: 0.00-1.00 risk score with validation
- **Risk Level Classification**: Low, medium, high, critical levels
- **Assessment Methods**: Tracking of assessment algorithms used
- **Content Analysis**: Website content storage for analysis
- **Pattern Detection**: JSONB storage of detected patterns
- **Metadata Storage**: Flexible assessment metadata
- **Confidence Scoring**: Assessment confidence levels
- **Expiration Support**: Risk assessment expiration dates

**Indexes**:
- Primary key on `id`
- Index on `business_id` for business-based queries
- Index on `risk_keyword_id` for keyword-based queries
- Index on `risk_score` for score-based filtering
- Index on `risk_level` for level-based filtering
- Index on `assessment_date` for temporal queries
- Index on `expires_at` for expiration management
- GIN indexes on array and JSONB fields

### **11. risk_keyword_relationships**

**Purpose**: Advanced relationships between risk keywords for enhanced detection and pattern recognition.

```sql
CREATE TABLE risk_keyword_relationships (
    id SERIAL PRIMARY KEY,
    parent_keyword_id INTEGER NOT NULL REFERENCES risk_keywords(id) ON DELETE CASCADE,
    child_keyword_id INTEGER NOT NULL REFERENCES risk_keywords(id) ON DELETE CASCADE,
    relationship_type VARCHAR(50) NOT NULL CHECK (relationship_type IN (
        'synonym', 'related', 'subcategory', 'superset', 'conflict', 'enhances'
    )),
    confidence_score DECIMAL(3,2) DEFAULT 0.80 CHECK (confidence_score >= 0.00 AND confidence_score <= 1.00),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CHECK (parent_keyword_id != child_keyword_id),
    UNIQUE(parent_keyword_id, child_keyword_id, relationship_type)
);
```

**Key Features**:
- **Relationship Types**: Synonym, related, subcategory, superset, conflict, enhances
- **Confidence Scoring**: Relationship confidence levels
- **Self-Reference Prevention**: Check constraint prevents self-references
- **Unique Relationships**: Prevents duplicate relationship definitions
- **Cascade Deletion**: Relationships deleted when keywords are deleted

---

## 🔗 **Code Crosswalk Tables**

### **12. industry_code_crosswalks**

**Purpose**: Comprehensive crosswalk mapping between industries and all classification codes with usage tracking.

```sql
CREATE TABLE industry_code_crosswalks (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    mcc_code VARCHAR(10),
    naics_code VARCHAR(10),
    sic_code VARCHAR(10),
    code_description TEXT,
    confidence_score DECIMAL(3,2) DEFAULT 0.80 CHECK (confidence_score >= 0.00 AND confidence_score <= 1.00),
    is_primary BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    usage_frequency INTEGER DEFAULT 0,
    last_used TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(industry_id, mcc_code, naics_code, sic_code)
);
```

**Key Features**:
- **Multi-Code Support**: MCC, NAICS, and SIC code mapping
- **Confidence Scoring**: Crosswalk confidence levels
- **Primary Designation**: Primary code identification
- **Usage Tracking**: Frequency and last used timestamps
- **Unique Constraints**: Prevents duplicate crosswalk entries
- **Cascade Deletion**: Crosswalks deleted when industry is deleted

**Indexes**:
- Primary key on `id`
- Index on `industry_id` for industry-based queries
- Index on `mcc_code` for MCC-based queries
- Index on `naics_code` for NAICS-based queries
- Index on `sic_code` for SIC-based queries
- Index on `is_active` for active crosswalk filtering
- Index on `is_primary` for primary code identification
- Index on `usage_frequency` for usage-based sorting
- Composite indexes for common query patterns

---

## 📊 **Performance Monitoring Tables**

### **13. classification_performance_metrics**

**Purpose**: Comprehensive performance tracking for classification system including risk assessment metrics.

```sql
CREATE TABLE classification_performance_metrics (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    request_id VARCHAR(255),
    business_name VARCHAR(500),
    business_description TEXT,
    website_url VARCHAR(1000),
    predicted_industry VARCHAR(255),
    predicted_confidence DECIMAL(3,2),
    actual_industry VARCHAR(255),
    actual_confidence DECIMAL(3,2),
    accuracy_score DECIMAL(3,2),
    response_time_ms DECIMAL(10,2),
    processing_time_ms DECIMAL(10,2),
    classification_method VARCHAR(100),
    keywords_used TEXT[],
    risk_keywords_detected TEXT[],
    risk_score DECIMAL(3,2),
    risk_level VARCHAR(20),
    confidence_threshold DECIMAL(3,2) DEFAULT 0.50,
    is_correct BOOLEAN,
    error_message TEXT,
    user_feedback TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Key Features**:
- **Comprehensive Metrics**: Classification and risk assessment performance
- **Temporal Tracking**: Timestamp-based performance analysis
- **Method Tracking**: Classification algorithm performance comparison
- **Accuracy Scoring**: Prediction vs. actual accuracy metrics
- **Performance Metrics**: Response and processing time tracking
- **Risk Integration**: Risk assessment performance tracking
- **User Feedback**: Human feedback integration for ML improvement

**Indexes**:
- Primary key on `id`
- Index on `timestamp` for temporal queries
- Index on `request_id` for request tracking
- Index on `predicted_industry` for industry-based analysis
- Index on `accuracy_score` for accuracy-based filtering
- Index on `response_time_ms` for performance analysis
- Index on `classification_method` for method comparison
- Index on `risk_level` for risk-based analysis
- GIN indexes on array fields

### **14. classification_accuracy_metrics**

**Purpose**: Detailed accuracy tracking for classification system validation and improvement.

```sql
CREATE TABLE classification_accuracy_metrics (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    request_id VARCHAR(255),
    business_name VARCHAR(500),
    business_description TEXT,
    website_url VARCHAR(1000),
    predicted_industry VARCHAR(255),
    predicted_confidence DECIMAL(3,2),
    actual_industry VARCHAR(255),
    actual_confidence DECIMAL(3,2),
    accuracy_score DECIMAL(3,2),
    response_time_ms DECIMAL(10,2),
    processing_time_ms DECIMAL(10,2),
    classification_method VARCHAR(100),
    keywords_used TEXT[],
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Key Features**:
- **Accuracy Focus**: Specialized accuracy tracking
- **Method Comparison**: Algorithm performance comparison
- **Confidence Tracking**: Prediction confidence analysis
- **Performance Metrics**: Response time tracking
- **Keyword Analysis**: Keyword usage tracking

---

## 📈 **Unified Monitoring Tables**

### **15. unified_performance_metrics**

**Purpose**: Consolidated performance metrics for all system components.

```sql
CREATE TABLE unified_performance_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    metric_name VARCHAR(255) NOT NULL,
    metric_type VARCHAR(100) NOT NULL CHECK (metric_type IN (
        'classification', 'risk_assessment', 'api_response', 'database_query', 'external_service'
    )),
    metric_value DECIMAL(15,4) NOT NULL,
    metric_unit VARCHAR(50),
    component_name VARCHAR(255),
    service_name VARCHAR(255),
    environment VARCHAR(50) DEFAULT 'production',
    tags JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Key Features**:
- **Unified Metrics**: All system metrics in one table
- **Metric Types**: Classification, risk assessment, API, database, external service
- **Flexible Tagging**: JSONB tags for flexible categorization
- **Environment Support**: Environment-specific metrics
- **Metadata Storage**: Additional metric context

### **16. unified_performance_alerts**

**Purpose**: Consolidated alerting system for performance thresholds.

```sql
CREATE TABLE unified_performance_alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    alert_name VARCHAR(255) NOT NULL,
    alert_type VARCHAR(100) NOT NULL CHECK (alert_type IN (
        'performance', 'accuracy', 'error_rate', 'availability', 'custom'
    )),
    severity VARCHAR(20) NOT NULL CHECK (severity IN (
        'low', 'medium', 'high', 'critical'
    )),
    threshold_value DECIMAL(15,4),
    actual_value DECIMAL(15,4),
    component_name VARCHAR(255),
    service_name VARCHAR(255),
    environment VARCHAR(50) DEFAULT 'production',
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN (
        'active', 'acknowledged', 'resolved', 'suppressed'
    )),
    message TEXT,
    metadata JSONB DEFAULT '{}',
    triggered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Key Features**:
- **Unified Alerting**: All system alerts in one table
- **Alert Types**: Performance, accuracy, error rate, availability, custom
- **Severity Levels**: Low, medium, high, critical
- **Status Tracking**: Active, acknowledged, resolved, suppressed
- **Threshold Monitoring**: Threshold vs. actual value tracking

### **17. unified_performance_reports**

**Purpose**: Consolidated reporting for performance analysis and business intelligence.

```sql
CREATE TABLE unified_performance_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    report_name VARCHAR(255) NOT NULL,
    report_type VARCHAR(100) NOT NULL CHECK (report_type IN (
        'daily', 'weekly', 'monthly', 'quarterly', 'annual', 'ad_hoc'
    )),
    report_period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    report_period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    component_name VARCHAR(255),
    service_name VARCHAR(255),
    environment VARCHAR(50) DEFAULT 'production',
    report_data JSONB NOT NULL,
    summary_metrics JSONB,
    generated_by VARCHAR(255),
    generated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Key Features**:
- **Unified Reporting**: All system reports in one table
- **Report Types**: Daily, weekly, monthly, quarterly, annual, ad hoc
- **Period Tracking**: Start and end timestamps for report periods
- **Data Storage**: JSONB storage for flexible report data
- **Summary Metrics**: High-level metrics for quick analysis

---

## 🔒 **Security and Audit Tables**

### **18. audit_logs**

**Purpose**: Comprehensive audit logging for all system activities.

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Key Features**:
- **Comprehensive Logging**: All user actions tracked
- **Change Tracking**: Old and new values for updates
- **Security Context**: IP address, user agent, session tracking
- **Flexible Metadata**: Additional context storage
- **User Attribution**: User ID for all actions

### **19. compliance_checks**

**Purpose**: Compliance tracking and validation for regulatory requirements.

```sql
CREATE TABLE compliance_checks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_id UUID NOT NULL,
    compliance_framework VARCHAR(100) NOT NULL,
    check_type VARCHAR(100) NOT NULL,
    check_status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (check_status IN (
        'pending', 'passed', 'failed', 'warning', 'not_applicable'
    )),
    check_result JSONB,
    check_metadata JSONB DEFAULT '{}',
    performed_by UUID REFERENCES users(id),
    performed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Key Features**:
- **Framework Support**: Multiple compliance frameworks
- **Check Types**: Various compliance check types
- **Status Tracking**: Pending, passed, failed, warning, not applicable
- **Result Storage**: Detailed check results in JSONB
- **Expiration Support**: Compliance check expiration dates

---

## 🔧 **System Configuration Tables**

### **20. migration_log**

**Purpose**: Database migration tracking and version control.

```sql
CREATE TABLE migration_log (
    id SERIAL PRIMARY KEY,
    migration_name VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN (
        'pending', 'running', 'completed', 'failed', 'rolled_back'
    )),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Key Features**:
- **Migration Tracking**: All database migrations logged
- **Status Management**: Pending, running, completed, failed, rolled back
- **Error Handling**: Error message storage for failed migrations
- **Version Control**: Migration name uniqueness for version tracking

---

## 📋 **Table Relationships Summary**

### **Primary Relationships**

1. **users** → **api_keys** (1:many)
2. **users** → **merchants** (1:many, via created_by)
3. **users** → **audit_logs** (1:many)
4. **users** → **compliance_checks** (1:many, via performed_by)

5. **industries** → **industry_keywords** (1:many)
6. **industries** → **classification_codes** (1:many)
7. **industries** → **industry_patterns** (1:many)
8. **industries** → **keyword_weights** (1:many)
9. **industries** → **industry_code_crosswalks** (1:many)

10. **risk_keywords** → **business_risk_assessments** (1:many)
11. **risk_keywords** → **risk_keyword_relationships** (1:many, self-referencing)

12. **merchants** → **business_risk_assessments** (1:many)
13. **merchants** → **compliance_checks** (1:many)

### **Foreign Key Constraints**

- **Cascade Deletion**: Most classification tables cascade delete when parent industry is deleted
- **Set Null**: Risk keyword references in business risk assessments set to null when keyword is deleted
- **Restrict**: User references prevent deletion of users with active records

---

## 🎯 **Data Flow Patterns**

### **Classification Flow**
1. **Input**: Business name, description, website URL
2. **Processing**: Industry classification using keywords, patterns, and ML models
3. **Output**: Industry classification with confidence scores
4. **Storage**: Results stored in classification_performance_metrics

### **Risk Assessment Flow**
1. **Input**: Business data and website content
2. **Processing**: Risk keyword matching and pattern detection
3. **Scoring**: Risk score calculation with confidence metrics
4. **Storage**: Results stored in business_risk_assessments

### **Performance Monitoring Flow**
1. **Collection**: Metrics collected from all system components
2. **Aggregation**: Metrics aggregated in unified_performance_metrics
3. **Alerting**: Threshold-based alerts in unified_performance_alerts
4. **Reporting**: Periodic reports in unified_performance_reports

---

## 🔍 **Index Strategy**

### **Performance Indexes**
- **Primary Keys**: All tables have UUID or SERIAL primary keys
- **Unique Indexes**: Email, username, registration numbers
- **Foreign Key Indexes**: All foreign key columns indexed
- **Composite Indexes**: Common query patterns optimized
- **GIN Indexes**: Array and JSONB fields for efficient queries
- **Full-Text Indexes**: Text search optimization

### **Query Optimization**
- **Covering Indexes**: Include frequently accessed columns
- **Partial Indexes**: Active records only where applicable
- **Expression Indexes**: Computed columns for performance
- **Hash Indexes**: Equality-only queries optimization

---

## 🛡️ **Security Implementation**

### **Row Level Security (RLS)**
- **Public Read Access**: Classification and risk data publicly readable
- **Authenticated Write**: All write operations require authentication
- **Role-Based Access**: Different permissions for different user roles
- **Data Isolation**: User data isolated by user ID

### **Data Validation**
- **Check Constraints**: Data type and range validation
- **Trigger Validation**: Complex business rule validation
- **Foreign Key Constraints**: Referential integrity enforcement
- **Unique Constraints**: Data uniqueness enforcement

---

## 📊 **Performance Characteristics**

### **Table Sizes (Estimated)**
- **users**: ~10K records
- **merchants**: ~100K records
- **industries**: ~1K records
- **industry_keywords**: ~50K records
- **risk_keywords**: ~10K records
- **business_risk_assessments**: ~500K records
- **classification_performance_metrics**: ~1M records

### **Query Performance Targets**
- **Simple Lookups**: <10ms
- **Complex Joins**: <100ms
- **Aggregation Queries**: <500ms
- **Full-Text Search**: <200ms
- **Risk Assessment**: <50ms

---

## 🔄 **Maintenance Procedures**

### **Regular Maintenance**
- **Index Rebuilding**: Weekly index maintenance
- **Statistics Updates**: Daily statistics refresh
- **Vacuum Operations**: Automated vacuum scheduling
- **Performance Monitoring**: Continuous performance tracking

### **Data Archival**
- **Performance Metrics**: 1-year retention with archival
- **Audit Logs**: 7-year retention for compliance
- **Risk Assessments**: 3-year retention with archival
- **Classification Data**: Permanent retention

---

## 📚 **Documentation Standards**

### **Schema Documentation**
- **Table Comments**: All tables have descriptive comments
- **Column Comments**: Key columns have explanatory comments
- **Constraint Documentation**: All constraints documented
- **Index Documentation**: Index purposes documented

### **Code Documentation**
- **Function Comments**: All functions have GoDoc-style comments
- **Parameter Documentation**: All parameters documented
- **Return Value Documentation**: Return values documented
- **Example Usage**: Code examples provided

---

**Document Status**: ✅ **COMPLETED**  
**Next Review**: Monthly during active development  
**Maintainer**: KYB Platform Development Team
