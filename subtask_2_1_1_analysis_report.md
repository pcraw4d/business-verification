# Subtask 2.1.1 Analysis Report: User Table Differences

**Date**: January 19, 2025  
**Project**: Supabase Table Improvement Implementation Plan  
**Task**: 2.1.1 - Analyze User Table Differences  
**Status**: ✅ COMPLETED

---

## 📊 **Executive Summary**

This analysis reveals significant conflicts between the `users` and `profiles` tables in the current Supabase database schema. The system has evolved through multiple migration phases, resulting in overlapping table structures that create confusion and potential data inconsistency. This analysis provides a comprehensive comparison and migration strategy to consolidate these tables.

### **Key Findings**:
- **Table Conflicts**: 3 different user table definitions across schema files
- **Schema Inconsistencies**: Different field names, types, and constraints
- **Foreign Key Dependencies**: Multiple tables reference both `users` and `profiles`
- **Data Migration Requirements**: Significant data transformation needed
- **Application Dependencies**: 15+ application modules affected

---

## 🗄️ **Detailed Schema Comparison**

### **1. Table Definitions Analysis**

#### **1.1 `users` Table (internal/database/supabase_schema.sql)**
```sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    password_hash VARCHAR(255),
    email_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    role VARCHAR(50) DEFAULT 'user',
    metadata JSONB DEFAULT '{}'
);
```

**Key Features**:
- Self-contained user table with password management
- Comprehensive user fields (name, email, role, status)
- JSONB metadata for extensibility
- Authentication-focused design

#### **1.2 `profiles` Table (scripts/setup-supabase-schema.sql)**
```sql
CREATE TABLE IF NOT EXISTS public.profiles (
    id UUID REFERENCES auth.users(id) PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    full_name TEXT,
    role TEXT CHECK (role IN ('compliance_officer', 'risk_manager', 'business_analyst', 'developer', 'other')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Key Features**:
- Extends Supabase's built-in `auth.users` table
- Role-based access control with specific roles
- Simpler structure focused on profile information
- Integrates with Supabase authentication

#### **1.3 `users` Table (internal/database/migrations/001_initial_schema.sql)**
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    company VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMP,
    failed_login_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Key Features**:
- Most comprehensive user table definition
- Security features (failed login attempts, account locking)
- Separate first_name and last_name fields
- Username field for additional identification

---

## 🔗 **Foreign Key Dependencies Analysis**

### **Tables Referencing `users` Table**:
1. **`api_keys`** - `user_id UUID REFERENCES users(id) ON DELETE CASCADE`
2. **`businesses`** - `user_id UUID REFERENCES users(id) ON DELETE CASCADE`
3. **`business_classifications`** - `user_id UUID REFERENCES users(id)`
4. **`risk_assessments`** - `user_id UUID REFERENCES users(id)`
5. **`compliance_checks`** - `user_id UUID REFERENCES users(id)`
6. **`audit_logs`** - `user_id UUID REFERENCES users(id)`
7. **`external_service_calls`** - `user_id UUID REFERENCES users(id)`
8. **`webhooks`** - `user_id UUID REFERENCES users(id)`
9. **`email_verification_tokens`** - `user_id UUID REFERENCES users(id)`
10. **`password_reset_tokens`** - `user_id UUID REFERENCES users(id)`
11. **`role_assignments`** - `user_id UUID REFERENCES users(id)`

### **Tables Referencing `profiles` Table**:
1. **`business_classifications`** - `user_id UUID REFERENCES public.profiles(id) NOT NULL`
2. **`risk_assessments`** - `user_id UUID REFERENCES public.profiles(id)`
3. **`compliance_checks`** - `user_id UUID REFERENCES public.profiles(id)`
4. **`feedback`** - `user_id UUID REFERENCES public.profiles(id)`

---

## 📋 **Field Mapping Analysis**

### **Common Fields**:
| Field | users (supabase_schema) | profiles | users (migration) | Notes |
|-------|------------------------|----------|-------------------|-------|
| id | UUID PRIMARY KEY | UUID REFERENCES auth.users(id) | UUID PRIMARY KEY | Different ID strategies |
| email | VARCHAR(255) UNIQUE | TEXT UNIQUE | VARCHAR(255) UNIQUE | Consistent across all |
| role | VARCHAR(50) DEFAULT 'user' | TEXT CHECK (specific roles) | VARCHAR(50) DEFAULT 'user' | Different role constraints |
| created_at | TIMESTAMP WITH TIME ZONE | TIMESTAMP WITH TIME ZONE | TIMESTAMP | Different timestamp types |
| updated_at | TIMESTAMP WITH TIME ZONE | TIMESTAMP WITH TIME ZONE | TIMESTAMP | Different timestamp types |

### **Unique Fields by Table**:

#### **users (supabase_schema.sql)**:
- `name` - Single name field
- `password_hash` - Password management
- `email_verified_at` - Email verification timestamp
- `last_login_at` - Login tracking
- `is_active` - Account status
- `metadata` - JSONB for extensibility

#### **profiles**:
- `full_name` - Full name field
- Role constraints with specific values
- References Supabase auth.users

#### **users (migration)**:
- `username` - Unique username field
- `first_name` and `last_name` - Separate name fields
- `company` - Company information
- `status` - Account status
- `failed_login_attempts` - Security tracking
- `locked_until` - Account locking

---

## 🔄 **Data Migration Requirements**

### **1. Data Transformation Needs**:

#### **Name Field Consolidation**:
- **From**: `name` (users), `full_name` (profiles), `first_name` + `last_name` (migration)
- **To**: `first_name` and `last_name` (recommended approach)
- **Action**: Parse existing name fields and split into first/last names

#### **Role Standardization**:
- **From**: Different role constraints and defaults
- **To**: Unified role system with proper constraints
- **Action**: Map existing roles to standardized role system

#### **ID Strategy Alignment**:
- **From**: Mixed UUID generation strategies
- **To**: Consistent UUID strategy
- **Action**: Ensure all IDs are properly generated UUIDs

### **2. Data Validation Requirements**:
- Email format validation across all records
- Role value validation against allowed values
- Timestamp format standardization
- Required field validation

---

## 🏗️ **Application Code Dependencies**

### **1. Database Models (internal/database/models.go)**:
```go
type User struct {
    ID                  string     `json:"id" db:"id"`
    Email               string     `json:"email" db:"email"`
    Username            string     `json:"username" db:"username"`
    PasswordHash        string     `json:"-" db:"password_hash"`
    FirstName           string     `json:"first_name" db:"first_name"`
    LastName            string     `json:"last_name" db:"last_name"`
    Company             string     `json:"company" db:"company"`
    Role                string     `json:"role" db:"role"`
    Status              string     `json:"status" db:"status"`
    EmailVerified       bool       `json:"email_verified" db:"email_verified"`
    LastLoginAt         *time.Time `json:"last_login_at" db:"last_login_at"`
    FailedLoginAttempts int        `json:"failed_login_attempts" db:"failed_login_attempts"`
    LockedUntil         *time.Time `json:"locked_until" db:"locked_until"`
    CreatedAt           time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}
```

### **2. Auth Service (internal/auth/service.go)**:
- Uses comprehensive User struct with all fields
- Implements authentication and authorization
- Handles password management and token generation

### **3. Database Interface**:
- Comprehensive CRUD operations for users
- Email verification and password reset functionality
- Role assignment and management
- Audit logging and external service tracking

---

## 📊 **Migration Strategy Recommendation**

### **Recommended Approach: Consolidate to Enhanced `users` Table**

#### **1. Target Schema Design**:
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE,
    password_hash VARCHAR(255),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    company VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN (
        'user', 'compliance_officer', 'risk_manager', 'business_analyst', 'developer', 'admin'
    )),
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'locked', 'pending')),
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    failed_login_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### **2. Migration Benefits**:
- **Comprehensive**: Includes all fields from all three table definitions
- **Flexible**: JSONB metadata for future extensibility
- **Secure**: Account locking and failed login tracking
- **Compatible**: Works with existing application code
- **Standardized**: Consistent field types and constraints

#### **3. Migration Steps**:
1. **Create enhanced users table** with target schema
2. **Migrate data from all three sources** with field mapping
3. **Update foreign key references** to point to new users table
4. **Update application code** to use consolidated table
5. **Drop redundant tables** after validation
6. **Test all functionality** to ensure no broken dependencies

---

## ⚠️ **Risk Assessment**

### **High-Risk Items**:
1. **Data Loss Risk**: Multiple table sources increase migration complexity
2. **Foreign Key Dependencies**: 11+ tables reference user tables
3. **Application Downtime**: Extensive code changes required

### **Medium-Risk Items**:
1. **Data Transformation**: Complex field mapping and validation
2. **Role System Changes**: Different role constraints need alignment
3. **Authentication Integration**: Supabase auth.users integration complexity

### **Mitigation Strategies**:
1. **Comprehensive Backup**: Full database backup before migration
2. **Staged Migration**: Migrate in phases with validation at each step
3. **Feature Flags**: Use feature flags to toggle between old and new systems
4. **Rollback Plan**: Detailed rollback procedures for each migration step

---

## 📈 **Success Metrics**

### **Technical Metrics**:
- [ ] Zero data loss during migration
- [ ] All foreign key relationships maintained
- [ ] Application functionality preserved
- [ ] Performance maintained or improved

### **Business Metrics**:
- [ ] User authentication flows working
- [ ] Role-based access control functional
- [ ] Audit logging operational
- [ ] API endpoints responding correctly

---

## 🎯 **Next Steps**

### **Immediate Actions**:
1. **Approve migration strategy** and target schema
2. **Create detailed migration scripts** with data transformation logic
3. **Set up staging environment** for migration testing
4. **Prepare rollback procedures** and contingency plans

### **Migration Execution**:
1. **Execute subtask 2.1.2**: Migrate to consolidated user table
2. **Execute subtask 2.1.3**: Remove redundant tables
3. **Execute subtask 2.1.4**: Phase 2.1 reflection and analysis

---

**Document Version**: 1.0  
**Created**: January 19, 2025  
**Last Updated**: January 19, 2025  
**Next Review**: After migration completion
