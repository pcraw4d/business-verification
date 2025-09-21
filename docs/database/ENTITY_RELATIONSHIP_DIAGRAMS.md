# KYB Platform - Entity Relationship Diagrams

## 📋 **Document Overview**

**Document Version**: 1.0  
**Created**: January 19, 2025  
**Last Updated**: January 19, 2025  
**Purpose**: Visual representation of database table relationships and cardinality

This document provides comprehensive entity relationship diagrams for the KYB Platform database schema, showing table relationships, cardinality, and data flow patterns.

---

## 🗄️ **Core Entity Relationship Diagram**

### **Primary Domain Relationships**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     users       │    │   api_keys      │    │   merchants     │
│                 │    │                 │    │                 │
│ id (PK)         │◄───┤ user_id (FK)    │    │ id (PK)         │
│ email           │    │ name            │    │ name            │
│ username        │    │ key_hash        │    │ legal_name      │
│ role            │    │ permissions     │    │ registration_#  │
│ status          │    │ expires_at      │    │ industry        │
│ is_active       │    │ is_active       │    │ compliance_     │
│ created_at      │    │ created_at      │    │   status        │
│ updated_at      │    │ updated_at      │    │ created_by (FK) │
└─────────────────┘    └─────────────────┘    │ created_at      │
         │                                      │ updated_at      │
         │                                      └─────────────────┘
         │                                               │
         │                                               │
         ▼                                               ▼
┌─────────────────┐                            ┌─────────────────┐
│   audit_logs    │                            │business_risk_   │
│                 │                            │assessments      │
│ id (PK)         │                            │                 │
│ user_id (FK)    │                            │ id (PK)         │
│ action          │                            │ business_id     │
│ resource_type   │                            │ risk_keyword_id │
│ resource_id     │                            │ risk_score      │
│ old_values      │                            │ risk_level      │
│ new_values      │                            │ confidence_     │
│ ip_address      │                            │   score         │
│ created_at      │                            │ assessment_     │
└─────────────────┘                            │   date          │
                                               │ created_at      │
                                               └─────────────────┘
```

---

## 🏭 **Classification System ERD**

### **Industry Classification Relationships**

```
┌─────────────────┐
│   industries    │
│                 │
│ id (PK)         │
│ name            │
│ description     │
│ category        │
│ confidence_     │
│   threshold     │
│ is_active       │
│ created_at      │
│ updated_at      │
└─────────────────┘
         │
         │ 1:N
         ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│industry_keywords│    │classification_  │    │industry_patterns│
│                 │    │codes            │    │                 │
│ id (PK)         │    │                 │    │ id (PK)         │
│ industry_id (FK)│    │ id (PK)         │    │ industry_id (FK)│
│ keyword         │    │ industry_id (FK)│    │ pattern         │
│ weight          │    │ code_type       │    │ pattern_type    │
│ is_active       │    │ code            │    │ confidence_     │
│ created_at      │    │ description     │    │   score         │
│ updated_at      │    │ is_active       │    │ is_active       │
└─────────────────┘    │ created_at      │    │ created_at      │
                       │ updated_at      │    │ updated_at      │
                       └─────────────────┘    └─────────────────┘
                                │
                                │
                                ▼
                       ┌─────────────────┐
                       │keyword_weights  │
                       │                 │
                       │ id (PK)         │
                       │ industry_id (FK)│
                       │ keyword         │
                       │ base_weight     │
                       │ usage_count     │
                       │ success_count   │
                       │ is_active       │
                       │ created_at      │
                       └─────────────────┘
```

---

## ⚠️ **Risk Management System ERD**

### **Risk Keywords and Assessment Relationships**

```
┌─────────────────┐
│  risk_keywords  │
│                 │
│ id (PK)         │
│ keyword         │
│ risk_category   │
│ risk_severity   │
│ description     │
│ mcc_codes[]     │
│ naics_codes[]   │
│ sic_codes[]     │
│ card_brand_     │
│   restrictions[]│
│ detection_      │
│   patterns[]    │
│ synonyms[]      │
│ risk_score_     │
│   weight        │
│ detection_      │
│   confidence    │
│ is_active       │
│ created_at      │
│ updated_at      │
└─────────────────┘
         │
         │ 1:N
         ▼
┌─────────────────┐    ┌─────────────────┐
│business_risk_   │    │risk_keyword_    │
│assessments      │    │relationships    │
│                 │    │                 │
│ id (PK)         │    │ id (PK)         │
│ business_id     │    │ parent_keyword_ │
│ risk_keyword_id │    │   id (FK)       │
│ detected_       │    │ child_keyword_  │
│   keywords[]    │    │   id (FK)       │
│ risk_score      │    │ relationship_   │
│ risk_level      │    │   type          │
│ assessment_     │    │ confidence_     │
│   method        │    │   score         │
│ website_content │    │ is_active       │
│ detected_       │    │ created_at      │
│   patterns      │    │ updated_at      │
│ assessment_     │    └─────────────────┘
│   metadata      │
│ confidence_     │
│   score         │
│ assessment_     │
│   date          │
│ expires_at      │
│ created_at      │
│ updated_at      │
└─────────────────┘
```

---

## 🔗 **Code Crosswalk System ERD**

### **Industry Code Crosswalk Relationships**

```
┌─────────────────┐
│   industries    │
│                 │
│ id (PK)         │
│ name            │
│ description     │
│ category        │
│ confidence_     │
│   threshold     │
│ is_active       │
│ created_at      │
│ updated_at      │
└─────────────────┘
         │
         │ 1:N
         ▼
┌─────────────────┐
│industry_code_   │
│crosswalks       │
│                 │
│ id (PK)         │
│ industry_id (FK)│
│ mcc_code        │
│ naics_code      │
│ sic_code        │
│ code_           │
│   description   │
│ confidence_     │
│   score         │
│ is_primary      │
│ is_active       │
│ usage_          │
│   frequency     │
│ last_used       │
│ created_at      │
│ updated_at      │
└─────────────────┘
```

---

## 📊 **Performance Monitoring System ERD**

### **Unified Monitoring Relationships**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│unified_perf_    │    │unified_perf_    │    │unified_perf_    │
│metrics          │    │alerts           │    │reports          │
│                 │    │                 │    │                 │
│ id (PK)         │    │ id (PK)         │    │ id (PK)         │
│ metric_name     │    │ alert_name      │    │ report_name     │
│ metric_type     │    │ alert_type      │    │ report_type     │
│ metric_value    │    │ severity        │    │ report_period_  │
│ metric_unit     │    │ threshold_      │    │   start         │
│ component_name  │    │   value         │    │ report_period_  │
│ service_name    │    │ actual_value    │    │   end           │
│ environment     │    │ component_name  │    │ component_name  │
│ tags            │    │ service_name    │    │ service_name    │
│ metadata        │    │ environment     │    │ environment     │
│ timestamp       │    │ status          │    │ report_data     │
│ created_at      │    │ message         │    │ summary_        │
└─────────────────┘    │ metadata        │    │   metrics       │
                       │ triggered_at    │    │ generated_by    │
                       │ acknowledged_at │    │ generated_at    │
                       │ resolved_at     │    │ created_at      │
                       │ created_at      │    └─────────────────┘
                       └─────────────────┘
```

---

## 🎯 **Classification Performance ERD**

### **Performance Metrics Relationships**

```
┌─────────────────┐    ┌─────────────────┐
│classification_  │    │classification_  │
│performance_     │    │accuracy_        │
│metrics          │    │metrics          │
│                 │    │                 │
│ id (PK)         │    │ id (PK)         │
│ timestamp       │    │ timestamp       │
│ request_id      │    │ request_id      │
│ business_name   │    │ business_name   │
│ business_desc   │    │ business_desc   │
│ website_url     │    │ website_url     │
│ predicted_      │    │ predicted_      │
│   industry      │    │   industry      │
│ predicted_      │    │ predicted_      │
│   confidence    │    │   confidence    │
│ actual_industry │    │ actual_industry │
│ actual_         │    │ actual_         │
│   confidence    │    │   confidence    │
│ accuracy_score  │    │ accuracy_score  │
│ response_time_  │    │ response_time_  │
│   ms            │    │   ms            │
│ processing_     │    │ processing_     │
│   time_ms       │    │   time_ms       │
│ classification_ │    │ classification_ │
│   method        │    │   method        │
│ keywords_used[] │    │ keywords_used[] │
│ risk_keywords_  │    │ created_at      │
│   detected[]    │    └─────────────────┘
│ risk_score      │
│ risk_level      │
│ confidence_     │
│   threshold     │
│ is_correct      │
│ error_message   │
│ user_feedback   │
│ created_at      │
└─────────────────┘
```

---

## 🔒 **Security and Compliance ERD**

### **Audit and Compliance Relationships**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     users       │    │   merchants     │    │compliance_      │
│                 │    │                 │    │checks           │
│ id (PK)         │    │ id (PK)         │    │                 │
│ email           │    │ name            │    │ id (PK)         │
│ username        │    │ legal_name      │    │ business_id     │
│ role            │    │ registration_#  │    │ compliance_     │
│ status          │    │ industry        │    │   framework     │
│ is_active       │    │ compliance_     │    │ check_type      │
│ created_at      │    │   status        │    │ check_status    │
│ updated_at      │    │ created_by (FK) │    │ check_result    │
└─────────────────┘    │ created_at      │    │ check_metadata  │
         │              │ updated_at      │    │ performed_by    │
         │              └─────────────────┘    │ performed_at    │
         │                       │              │ expires_at      │
         │                       │              │ created_at      │
         │                       │              │ updated_at      │
         │                       │              └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       │
┌─────────────────┐             │                       │
│   audit_logs    │             │                       │
│                 │             │                       │
│ id (PK)         │             │                       │
│ user_id (FK)    │             │                       │
│ action          │             │                       │
│ resource_type   │             │                       │
│ resource_id     │             │                       │
│ old_values      │             │                       │
│ new_values      │             │                       │
│ ip_address      │             │                       │
│ user_agent      │             │                       │
│ session_id      │             │                       │
│ metadata        │             │                       │
│ created_at      │             │                       │
└─────────────────┘             │                       │
                                │                       │
                                ▼                       ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │business_risk_   │    │     users       │
                       │assessments      │    │                 │
                       │                 │    │ id (PK)         │
                       │ id (PK)         │    │ email           │
                       │ business_id     │    │ username        │
                       │ risk_keyword_id │    │ role            │
                       │ detected_       │    │ status          │
                       │   keywords[]    │    │ is_active       │
                       │ risk_score      │    │ created_at      │
                       │ risk_level      │    │ updated_at      │
                       │ assessment_     │    └─────────────────┘
                       │   method        │
                       │ website_content │
                       │ detected_       │
                       │   patterns      │
                       │ assessment_     │
                       │   metadata      │
                       │ confidence_     │
                       │   score         │
                       │ assessment_     │
                       │   date          │
                       │ expires_at      │
                       │ created_at      │
                       │ updated_at      │
                       └─────────────────┘
```

---

## 🔄 **Data Flow Relationships**

### **Primary Data Flow Patterns**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Input Data    │    │  Processing     │    │   Output Data   │
│                 │    │                 │    │                 │
│ Business Name   │───►│ Classification  │───►│ Industry        │
│ Description     │    │ Engine          │    │ Classification  │
│ Website URL     │    │                 │    │ Confidence      │
│                 │    │ Risk Assessment │    │ Score           │
│                 │    │ Engine          │    │                 │
│                 │    │                 │    │ Risk Score      │
│                 │    │                 │    │ Risk Level      │
│                 │    │                 │    │ Detected        │
│                 │    │                 │    │ Keywords        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Storage       │    │   Monitoring    │    │   Reporting     │
│                 │    │                 │    │                 │
│ Performance     │    │ Unified         │    │ Performance     │
│ Metrics         │    │ Metrics         │    │ Reports         │
│                 │    │                 │    │                 │
│ Accuracy        │    │ Alerts          │    │ Business        │
│ Metrics         │    │                 │    │ Intelligence    │
│                 │    │ Performance     │    │                 │
│ Risk            │    │ Tracking        │    │ Compliance      │
│ Assessments     │    │                 │    │ Reports         │
│                 │    │ Error           │    │                 │
│ Audit Logs      │    │ Monitoring      │    │ Risk            │
│                 │    │                 │    │ Reports         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

---

## 📋 **Relationship Summary**

### **Cardinality Summary**

| Parent Table | Child Table | Relationship | Cardinality |
|--------------|-------------|--------------|-------------|
| users | api_keys | user_id | 1:N |
| users | merchants | created_by | 1:N |
| users | audit_logs | user_id | 1:N |
| users | compliance_checks | performed_by | 1:N |
| industries | industry_keywords | industry_id | 1:N |
| industries | classification_codes | industry_id | 1:N |
| industries | industry_patterns | industry_id | 1:N |
| industries | keyword_weights | industry_id | 1:N |
| industries | industry_code_crosswalks | industry_id | 1:N |
| risk_keywords | business_risk_assessments | risk_keyword_id | 1:N |
| risk_keywords | risk_keyword_relationships | parent_keyword_id | 1:N |
| risk_keywords | risk_keyword_relationships | child_keyword_id | 1:N |
| merchants | business_risk_assessments | business_id | 1:N |
| merchants | compliance_checks | business_id | 1:N |

### **Foreign Key Constraints**

| Table | Column | References | Constraint Type |
|-------|--------|------------|-----------------|
| api_keys | user_id | users(id) | CASCADE DELETE |
| merchants | created_by | users(id) | RESTRICT |
| audit_logs | user_id | users(id) | RESTRICT |
| compliance_checks | performed_by | users(id) | RESTRICT |
| industry_keywords | industry_id | industries(id) | CASCADE DELETE |
| classification_codes | industry_id | industries(id) | CASCADE DELETE |
| industry_patterns | industry_id | industries(id) | CASCADE DELETE |
| keyword_weights | industry_id | industries(id) | CASCADE DELETE |
| industry_code_crosswalks | industry_id | industries(id) | CASCADE DELETE |
| business_risk_assessments | risk_keyword_id | risk_keywords(id) | SET NULL |
| risk_keyword_relationships | parent_keyword_id | risk_keywords(id) | CASCADE DELETE |
| risk_keyword_relationships | child_keyword_id | risk_keywords(id) | CASCADE DELETE |
| compliance_checks | business_id | merchants(id) | RESTRICT |

---

## 🎯 **Key Design Patterns**

### **1. Hierarchical Relationships**
- **Industries** → **Keywords/Patterns/Codes**: One-to-many relationships with cascade deletion
- **Risk Keywords** → **Relationships**: Self-referencing many-to-many with relationship types

### **2. Audit Trail Pattern**
- **All Tables**: created_at, updated_at timestamps
- **User Attribution**: created_by, performed_by fields where applicable
- **Change Tracking**: audit_logs table for all system changes

### **3. Soft Delete Pattern**
- **is_active** fields on most tables for soft deletion
- **status** fields for state management
- **expires_at** fields for time-based data expiration

### **4. Flexible Metadata Pattern**
- **JSONB fields**: metadata, tags, patterns for extensible data
- **Array fields**: keywords, codes, restrictions for multi-value storage
- **Enum constraints**: status, type, level fields for controlled values

### **5. Performance Optimization Pattern**
- **Comprehensive indexing**: Primary keys, foreign keys, composite indexes
- **GIN indexes**: Array and JSONB fields for efficient queries
- **Partial indexes**: Active records only where applicable

---

**Document Status**: ✅ **COMPLETED**  
**Next Review**: Monthly during active development  
**Maintainer**: KYB Platform Development Team
