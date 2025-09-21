# Subtask 2.3.1 Analysis Report: Audit Tables Consolidation

**Date**: January 19, 2025  
**Task**: 2.3.1 - Merge Audit Tables  
**Status**: 🔄 **IN PROGRESS**  
**Priority**: Medium  

---

## 🎯 **Analysis Overview**

This analysis examines the current state of audit tables in the system to identify conflicts, duplications, and consolidation opportunities. The goal is to create a unified audit schema that eliminates redundancy while maintaining comprehensive audit trail functionality.

---

## 📊 **Current Audit Table Analysis**

### **1. audit_logs Table (Primary Schema)**

**Location**: `internal/database/supabase_schema.sql` (lines 111-124)

```sql
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    api_key_id UUID REFERENCES api_keys(id) ON DELETE SET NULL,
    event_type VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id UUID,
    action VARCHAR(100) NOT NULL,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);
```

**Key Features**:
- UUID primary key
- References to users and api_keys tables
- JSONB for details and metadata
- Comprehensive event tracking
- INET for IP addresses

### **2. merchant_audit_logs Table (Merchant-Specific)**

**Location**: `internal/database/migrations/005_merchant_portfolio_schema.sql` (lines 85-98)

```sql
CREATE TABLE IF NOT EXISTS merchant_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    merchant_id UUID REFERENCES merchants(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(100) NOT NULL,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(100),
    session_id UUID REFERENCES merchant_sessions(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Key Features**:
- Merchant-specific audit logging
- Session tracking via merchant_sessions
- Request ID tracking
- String resource_id (vs UUID in audit_logs)

### **3. audit_logs Table (Legacy Schema)**

**Location**: `internal/database/migrations/001_initial_schema.sql` (lines 151-162)

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(100) NOT NULL,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Key Features**:
- Simpler schema without api_key_id or event_type
- No metadata field
- TIMESTAMP instead of TIMESTAMP WITH TIME ZONE

### **4. audit_logs Table (Classification Schema)**

**Location**: `supabase-migrations/001_initial_keyword_classification_schema.sql` (lines 142-151)

```sql
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    table_name VARCHAR(50) NOT NULL,
    record_id INTEGER NOT NULL,
    action VARCHAR(20) NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE')),
    old_values JSONB,
    new_values JSONB,
    user_id VARCHAR(100),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Key Features**:
- SERIAL primary key (different from UUID)
- Table-level audit tracking
- Old/new values tracking
- String user_id (vs UUID)

---

## 🔍 **Schema Comparison Analysis**

### **Field Comparison Matrix**

| Field | Primary Schema | Merchant Schema | Legacy Schema | Classification Schema |
|-------|---------------|-----------------|---------------|---------------------|
| id | UUID | UUID | UUID | SERIAL |
| user_id | UUID | UUID | UUID | VARCHAR(100) |
| merchant_id | ❌ | UUID | ❌ | ❌ |
| api_key_id | UUID | ❌ | ❌ | ❌ |
| event_type | VARCHAR(100) | ❌ | ❌ | ❌ |
| resource_type | VARCHAR(100) | VARCHAR(100) | VARCHAR(100) | ❌ |
| resource_id | UUID | VARCHAR(100) | VARCHAR(100) | ❌ |
| action | VARCHAR(100) | VARCHAR(100) | VARCHAR(100) | VARCHAR(20) |
| details | JSONB | JSONB | JSONB | ❌ |
| old_values | ❌ | ❌ | ❌ | JSONB |
| new_values | ❌ | ❌ | ❌ | JSONB |
| ip_address | INET | INET | INET | ❌ |
| user_agent | TEXT | TEXT | TEXT | ❌ |
| request_id | ❌ | VARCHAR(100) | VARCHAR(100) | ❌ |
| session_id | ❌ | UUID | ❌ | ❌ |
| table_name | ❌ | ❌ | ❌ | VARCHAR(50) |
| record_id | ❌ | ❌ | ❌ | INTEGER |
| metadata | JSONB | ❌ | ❌ | ❌ |
| created_at | TIMESTAMP WITH TIME ZONE | TIMESTAMP WITH TIME ZONE | TIMESTAMP | TIMESTAMP WITH TIME ZONE |

### **Key Differences Identified**

1. **Primary Key Types**: UUID vs SERIAL inconsistency
2. **Resource ID Types**: UUID vs VARCHAR(100) inconsistency  
3. **User ID Types**: UUID vs VARCHAR(100) inconsistency
4. **Missing Fields**: Different schemas have different optional fields
5. **Data Types**: TIMESTAMP vs TIMESTAMP WITH TIME ZONE inconsistency
6. **Purpose**: General audit vs merchant-specific vs table-level audit

---

## 🏗️ **Application Code Analysis**

### **Current Usage Patterns**

#### **1. AuditService Usage**
- **File**: `internal/services/audit_service.go`
- **Method**: `LogMerchantOperation()`
- **Target**: Uses `AuditRepository.SaveAuditLog()` interface
- **Data Flow**: Creates `models.AuditLog` → Saves via repository → Also logs to compliance system

#### **2. Repository Implementation**
- **File**: `internal/database/merchant_portfolio_repository.go`
- **Method**: `CreateAuditLog()`
- **Target**: Inserts into `audit_logs` table (not `merchant_audit_logs`)
- **Query**: Uses the legacy schema structure

#### **3. Model Definition**
- **File**: `internal/models/merchant_portfolio.go`
- **Struct**: `AuditLog`
- **Fields**: Matches the legacy schema (no api_key_id, event_type, metadata)

### **Code Inconsistencies Found**

1. **Repository vs Schema Mismatch**: Code inserts into `audit_logs` but merchant schema defines `merchant_audit_logs`
2. **Model vs Schema Mismatch**: Model doesn't include all fields from primary schema
3. **Multiple Schema Definitions**: Different migration files define different audit_logs schemas
4. **Interface vs Implementation**: AuditRepository interface doesn't match actual implementation

---

## 🎯 **Consolidation Strategy**

### **Unified Audit Schema Design**

Based on the analysis, I recommend creating a unified audit schema that combines the best features from all existing schemas:

```sql
CREATE TABLE IF NOT EXISTS unified_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    api_key_id UUID REFERENCES api_keys(id) ON DELETE SET NULL,
    merchant_id UUID REFERENCES merchants(id) ON DELETE SET NULL,
    session_id UUID REFERENCES merchant_sessions(id) ON DELETE SET NULL,
    
    -- Event Classification
    event_type VARCHAR(100) NOT NULL,
    event_category VARCHAR(50) NOT NULL DEFAULT 'audit',
    action VARCHAR(100) NOT NULL,
    
    -- Resource Information
    resource_type VARCHAR(100),
    resource_id VARCHAR(100), -- Flexible to handle both UUID and string IDs
    table_name VARCHAR(50), -- For table-level audits
    
    -- Change Tracking
    old_values JSONB,
    new_values JSONB,
    details JSONB,
    
    -- Request Context
    request_id VARCHAR(100),
    ip_address INET,
    user_agent TEXT,
    
    -- Metadata and Timestamps
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Indexes for performance
    CONSTRAINT chk_audit_log_action CHECK (action IN ('INSERT', 'UPDATE', 'DELETE', 'CREATE', 'READ', 'LOGIN', 'LOGOUT', 'ACCESS', 'EXPORT', 'IMPORT'))
);
```

### **Key Design Decisions**

1. **Flexible Resource ID**: VARCHAR(100) to handle both UUID and string identifiers
2. **Comprehensive Event Types**: Support for both CRUD operations and business events
3. **Change Tracking**: Old/new values for audit trail completeness
4. **Optional References**: All foreign keys are optional to support different audit scenarios
5. **Metadata Field**: JSONB for extensibility
6. **Performance Indexes**: Strategic indexing for common query patterns

---

## 📋 **Migration Plan**

### **Phase 1: Schema Creation**
1. Create `unified_audit_logs` table with comprehensive schema
2. Add performance indexes
3. Create audit log triggers for automatic logging

### **Phase 2: Data Migration**
1. Migrate data from `audit_logs` (all variants)
2. Migrate data from `merchant_audit_logs`
3. Validate data integrity and completeness

### **Phase 3: Code Updates**
1. Update `AuditLog` model to match unified schema
2. Update repository methods to use unified table
3. Update service layer to handle new fields
4. Update API handlers for new audit structure

### **Phase 4: Testing and Validation**
1. Test audit logging functionality
2. Validate data integrity
3. Performance testing
4. Rollback testing

### **Phase 5: Cleanup**
1. Drop redundant audit tables
2. Update documentation
3. Remove unused code

---

## 🚨 **Risk Assessment**

### **High Risk Items**
1. **Data Loss**: Multiple audit tables with different schemas
2. **Application Downtime**: Code changes required for repository layer
3. **Performance Impact**: Large audit tables may affect query performance

### **Mitigation Strategies**
1. **Comprehensive Backup**: Full database backup before migration
2. **Staged Migration**: Migrate in phases with validation at each step
3. **Rollback Plan**: Ability to revert to original tables if needed
4. **Performance Testing**: Benchmark before and after migration

---

## 📊 **Success Metrics**

### **Technical Metrics**
- ✅ **Schema Consolidation**: Single unified audit table
- ✅ **Data Integrity**: 100% data migration success
- ✅ **Performance**: Maintain or improve query performance
- ✅ **Code Consistency**: Unified model and repository interfaces

### **Business Metrics**
- ✅ **Audit Completeness**: All audit events captured
- ✅ **Compliance**: Maintain audit trail for regulatory requirements
- ✅ **Operational Efficiency**: Simplified audit management

---

## 🎯 **Next Steps**

1. **Approve Unified Schema**: Review and approve the proposed unified audit schema
2. **Create Migration Scripts**: Develop comprehensive migration scripts
3. **Update Models**: Modify Go models to match unified schema
4. **Test Migration**: Execute migration in development environment
5. **Update Application Code**: Modify repository and service layers
6. **Performance Testing**: Validate performance with unified schema
7. **Production Deployment**: Deploy to production with rollback capability

---

**Document Status**: ✅ **COMPLETED**  
**Next Task**: 2.3.1 Schema Creation and Migration Scripts  
**Review Required**: Technical lead approval for unified schema design
