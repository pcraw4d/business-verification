# KYB Platform - Provider Migration Preservation Guide

## 🎯 **Executive Summary**

This guide outlines the key strategies for preserving work and avoiding duplication when migrating between cloud providers (Supabase ↔ AWS). The focus is on maintaining abstraction layers, preserving existing code, and ensuring smooth transitions with minimal rework.

**Key Principle**: Write once, deploy anywhere through proper abstraction layers
**Goal**: Zero code duplication when switching providers
**Benefit**: Future migrations become configuration changes, not rewrites

---

## 🏗️ **Core Preservation Strategies**

### **1. Interface-First Design**

#### **Central Interface Definitions**
```go
// internal/interfaces.go - Single source of truth for all interfaces
package interfaces

import (
    "context"
    "time"
)

// Database interface - Never changes, regardless of provider
type Database interface {
    Connect(ctx context.Context) error
    Close() error
    Ping(ctx context.Context) error
    BeginTx(ctx context.Context) (Database, error)
    Commit() error
    Rollback() error
    // ... all database operations
}

// AuthService interface - Never changes, regardless of provider
type AuthService interface {
    SignUp(ctx context.Context, email, password string) (*User, error)
    SignIn(ctx context.Context, email, password string) (*AuthResponse, error)
    ValidateToken(ctx context.Context, token string) (*User, error)
    RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
    // ... all auth operations
}

// Cache interface - Never changes, regardless of provider
type Cache interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Clear(ctx context.Context) error
}

// Storage interface - Never changes, regardless of provider
type Storage interface {
    Upload(ctx context.Context, key string, data []byte) error
    Download(ctx context.Context, key string) ([]byte, error)
    Delete(ctx context.Context, key string) error
    List(ctx context.Context, prefix string) ([]string, error)
}
```

**Benefits:**
- ✅ **Zero interface changes** when switching providers
- ✅ **All business logic remains unchanged**
- ✅ **Tests work with any provider**
- ✅ **API contracts stay consistent**

### **2. Factory Pattern Implementation**

#### **Provider-Agnostic Factory**
```go
// internal/factory.go - Single factory for all providers
package factory

import (
    "fmt"
    "github.com/pcraw4d/business-verification/internal/config"
    "github.com/pcraw4d/business-verification/internal/interfaces"
)

// Database factory - Supports multiple providers
func NewDatabase(cfg *config.Config) (interfaces.Database, error) {
    switch cfg.Provider.Database {
    case "supabase":
        return database.NewSupabaseDB(cfg.Supabase), nil
    case "aws":
        return database.NewAWSPostgresDB(cfg.AWS), nil
    case "gcp":
        return database.NewGCPPostgresDB(cfg.GCP), nil
    default:
        return nil, fmt.Errorf("unsupported database provider: %s", cfg.Provider.Database)
    }
}

// Auth factory - Supports multiple providers
func NewAuthService(cfg *config.Config) (interfaces.AuthService, error) {
    switch cfg.Provider.Auth {
    case "supabase":
        return auth.NewSupabaseAuthService(cfg.Supabase), nil
    case "aws":
        return auth.NewAWSCognitoService(cfg.AWS), nil
    case "gcp":
        return auth.NewGCPIdentityService(cfg.GCP), nil
    default:
        return nil, fmt.Errorf("unsupported auth provider: %s", cfg.Provider.Auth)
    }
}

// Cache factory - Supports multiple providers
func NewCache(cfg *config.Config) (interfaces.Cache, error) {
    switch cfg.Provider.Cache {
    case "supabase":
        return cache.NewSupabaseCache(cfg.Supabase), nil
    case "aws":
        return cache.NewAWSRedisCache(cfg.AWS), nil
    case "gcp":
        return cache.NewGCPRedisCache(cfg.GCP), nil
    default:
        return nil, fmt.Errorf("unsupported cache provider: %s", cfg.Provider.Cache)
    }
}
```

**Benefits:**
- ✅ **Single factory** handles all providers
- ✅ **Easy to add new providers**
- ✅ **Configuration-driven** provider selection
- ✅ **No code changes** needed for provider switches

### **3. Configuration-Driven Architecture**

#### **Multi-Provider Configuration**
```go
// internal/config/config.go - Supports all providers
type Config struct {
    Environment string `json:"environment" yaml:"environment"`
    
    // Provider selection - Environment variable driven
    Provider ProviderConfig `json:"provider" yaml:"provider"`
    
    // Provider-specific configurations
    Supabase SupabaseConfig `json:"supabase" yaml:"supabase"`
    AWS      AWSConfig      `json:"aws" yaml:"aws"`
    GCP      GCPConfig      `json:"gcp" yaml:"gcp"`
    
    // Common configurations - Never change
    Server       ServerConfig       `json:"server" yaml:"server"`
    Observability ObservabilityConfig `json:"observability" yaml:"observability"`
    ExternalServices ExternalServicesConfig `json:"external_services" yaml:"external_services"`
}

type ProviderConfig struct {
    Database string `json:"database" yaml:"database"` // "supabase", "aws", "gcp"
    Auth     string `json:"auth" yaml:"auth"`         // "supabase", "aws", "gcp"
    Cache    string `json:"cache" yaml:"cache"`       // "supabase", "aws", "gcp"
    Storage  string `json:"storage" yaml:"storage"`   // "supabase", "aws", "gcp"
}
```

#### **Environment-Based Provider Selection**
```bash
# configs/development.env - Development with Supabase
PROVIDER_DATABASE=supabase
PROVIDER_AUTH=supabase
PROVIDER_CACHE=supabase
PROVIDER_STORAGE=supabase

# configs/staging.env - Staging with AWS
PROVIDER_DATABASE=aws
PROVIDER_AUTH=aws
PROVIDER_CACHE=aws
PROVIDER_STORAGE=aws

# configs/production.env - Production with AWS
PROVIDER_DATABASE=aws
PROVIDER_AUTH=aws
PROVIDER_CACHE=aws
PROVIDER_STORAGE=aws
```

**Benefits:**
- ✅ **Environment-specific** provider selection
- ✅ **No code changes** for provider switches
- ✅ **Easy testing** with different providers
- ✅ **Gradual migration** capability

---

## 🔄 **Migration Workflow**

### **Step 1: Implement New Provider Adapter**

#### **Create Provider-Specific Implementation**
```go
// internal/database/aws_postgres.go - New AWS implementation
package database

import (
    "context"
    "database/sql"
    "fmt"
    
    "github.com/pcraw4d/business-verification/internal/interfaces"
    _ "github.com/lib/pq"
)

type AWSPostgresDB struct {
    db     *sql.DB
    config *config.AWSConfig
    tx     *sql.Tx
}

// Implement all interface methods
func (p *AWSPostgresDB) Connect(ctx context.Context) error {
    // AWS-specific connection logic
}

func (p *AWSPostgresDB) Close() error {
    // AWS-specific cleanup logic
}

// ... implement all other interface methods
```

#### **Add to Factory**
```go
// internal/factory.go - Add new provider
func NewDatabase(cfg *config.Config) (interfaces.Database, error) {
    switch cfg.Provider.Database {
    case "supabase":
        return database.NewSupabaseDB(cfg.Supabase), nil
    case "aws":
        return database.NewAWSPostgresDB(cfg.AWS), nil  // New provider
    default:
        return nil, fmt.Errorf("unsupported database provider: %s", cfg.Provider.Database)
    }
}
```

### **Step 2: Update Configuration**

#### **Add Provider Configuration**
```go
// internal/config/config.go - Add new provider config
type AWSConfig struct {
    Region          string `json:"region" yaml:"region"`
    AccessKeyID     string `json:"access_key_id" yaml:"access_key_id"`
    SecretAccessKey string `json:"secret_access_key" yaml:"secret_access_key"`
    
    RDS RDSConfig `json:"rds" yaml:"rds"`
    // ... other AWS-specific configs
}
```

#### **Update Environment Variables**
```bash
# configs/production.env - Switch to AWS
PROVIDER_DATABASE=aws
PROVIDER_AUTH=aws
PROVIDER_CACHE=aws
PROVIDER_STORAGE=aws

AWS_REGION=us-west-2
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
```

### **Step 3: Test and Validate**

#### **Provider-Agnostic Tests**
```go
// test/provider_agnostic_test.go - Tests work with any provider
func TestDatabaseOperations(t *testing.T) {
    // This test works with Supabase, AWS, GCP, etc.
    cfg := loadTestConfig()
    
    db, err := factory.NewDatabase(cfg)
    if err != nil {
        t.Fatalf("Failed to create database: %v", err)
    }
    
    // Test database operations
    // These tests work regardless of the underlying provider
}
```

### **Step 4: Deploy and Switch**

#### **Blue-Green Deployment**
```yaml
# deployments/kubernetes/blue-green.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kyb-platform-blue
spec:
  template:
    spec:
      containers:
      - name: kyb-platform
        env:
        - name: PROVIDER_DATABASE
          value: "supabase"  # Current provider
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kyb-platform-green
spec:
  template:
    spec:
      containers:
      - name: kyb-platform
        env:
        - name: PROVIDER_DATABASE
          value: "aws"  # New provider
```

---

## 📊 **Preservation Benefits**

### **Code Preservation**

#### **What Gets Preserved**
```yaml
Preserved Components:
  - All business logic: 100% preserved
  - API endpoints: 100% preserved
  - Data models: 100% preserved
  - Validation logic: 100% preserved
  - Error handling: 100% preserved
  - Tests: 100% preserved
  - Documentation: 100% preserved
```

#### **What Gets Added**
```yaml
New Components:
  - Provider-specific adapters: Minimal new code
  - Provider configurations: Configuration only
  - Migration scripts: One-time use
  - Provider-specific tests: Minimal new tests
```

### **Effort Comparison**

#### **Without Preservation Strategy**
```yaml
Traditional Migration:
  - Business Logic Rewrite: 80% of code
  - API Changes: 60% of endpoints
  - Data Model Changes: 40% of models
  - Test Rewrite: 90% of tests
  - Documentation Update: 70% of docs
  - Total Effort: 6-12 months
```

#### **With Preservation Strategy**
```yaml
Preserved Migration:
  - Business Logic Rewrite: 0% (preserved)
  - API Changes: 0% (preserved)
  - Data Model Changes: 0% (preserved)
  - Test Rewrite: 0% (preserved)
  - Documentation Update: 10% (minor updates)
  - Total Effort: 2-4 weeks
```

---

## 🚀 **Implementation Checklist**

### **Phase 1: Setup Abstraction Layers**

- [ ] **Create Interface Definitions**
  - [ ] Define Database interface
  - [ ] Define AuthService interface
  - [ ] Define Cache interface
  - [ ] Define Storage interface

- [ ] **Implement Factory Pattern**
  - [ ] Create factory functions
  - [ ] Add provider selection logic
  - [ ] Implement error handling

- [ ] **Update Configuration**
  - [ ] Add provider selection config
  - [ ] Add provider-specific configs
  - [ ] Update environment variables

### **Phase 2: Implement Provider Adapters**

- [ ] **Database Adapters**
  - [ ] Supabase PostgreSQL adapter
  - [ ] AWS RDS PostgreSQL adapter
  - [ ] GCP Cloud SQL adapter

- [ ] **Auth Adapters**
  - [ ] Supabase Auth adapter
  - [ ] AWS Cognito adapter
  - [ ] GCP Identity adapter

- [ ] **Cache Adapters**
  - [ ] Supabase real-time adapter
  - [ ] AWS ElastiCache adapter
  - [ ] GCP Memorystore adapter

### **Phase 3: Testing and Validation**

- [ ] **Provider-Agnostic Tests**
  - [ ] Database operation tests
  - [ ] Auth operation tests
  - [ ] Cache operation tests
  - [ ] Integration tests

- [ ] **Provider-Specific Tests**
  - [ ] Provider connection tests
  - [ ] Provider-specific feature tests
  - [ ] Performance tests

### **Phase 4: Deployment and Migration**

- [ ] **Environment Setup**
  - [ ] Configure new provider
  - [ ] Set up monitoring
  - [ ] Prepare migration scripts

- [ ] **Blue-Green Deployment**
  - [ ] Deploy with new provider
  - [ ] Test functionality
  - [ ] Switch traffic
  - [ ] Monitor performance

---

## 🎯 **Best Practices**

### **1. Interface Design Principles**

#### **Keep Interfaces Stable**
```go
// ✅ Good: Stable interface
type Database interface {
    Connect(ctx context.Context) error
    Close() error
    Query(ctx context.Context, sql string, args ...interface{}) (*Rows, error)
}

// ❌ Bad: Provider-specific interface
type Database interface {
    Connect(ctx context.Context) error
    Close() error
    SupabaseQuery(ctx context.Context, sql string) (*SupabaseRows, error)  // Provider-specific
}
```

#### **Use Common Data Types**
```go
// ✅ Good: Common data types
type User struct {
    ID    string `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
}

// ❌ Bad: Provider-specific data types
type User struct {
    SupabaseID string `json:"supabase_id"`  // Provider-specific
    Email      string `json:"email"`
    Name       string `json:"name"`
}
```

### **2. Configuration Management**

#### **Environment-Based Selection**
```bash
# ✅ Good: Environment-driven provider selection
PROVIDER_DATABASE=aws
PROVIDER_AUTH=aws
PROVIDER_CACHE=aws

# ❌ Bad: Hard-coded provider selection
DATABASE_PROVIDER=aws  # Hard to change
```

#### **Provider-Specific Configs**
```go
// ✅ Good: Separate provider configs
type Config struct {
    Provider ProviderConfig
    Supabase SupabaseConfig
    AWS      AWSConfig
    GCP      GCPConfig
}

// ❌ Bad: Mixed provider configs
type Config struct {
    DatabaseHost string  // Which provider?
    AuthToken    string  // Which provider?
}
```

### **3. Testing Strategy**

#### **Provider-Agnostic Tests**
```go
// ✅ Good: Tests work with any provider
func TestUserRegistration(t *testing.T) {
    cfg := loadTestConfig()
    authService, _ := factory.NewAuthService(cfg)
    
    user, err := authService.SignUp(ctx, "test@example.com", "password")
    assert.NoError(t, err)
    assert.Equal(t, "test@example.com", user.Email)
}

// ❌ Bad: Provider-specific tests
func TestSupabaseUserRegistration(t *testing.T) {
    // Only works with Supabase
    supabaseClient := supabase.NewClient(...)
    // ...
}
```

---

## 📚 **Migration Examples**

### **Supabase → AWS Migration**

#### **1. Add AWS Implementation**
```go
// internal/database/aws_postgres.go
type AWSPostgresDB struct {
    db     *sql.DB
    config *config.AWSConfig
}

func (p *AWSPostgresDB) Connect(ctx context.Context) error {
    // AWS RDS connection logic
}
```

#### **2. Update Factory**
```go
// internal/factory.go
func NewDatabase(cfg *config.Config) (interfaces.Database, error) {
    switch cfg.Provider.Database {
    case "supabase":
        return database.NewSupabaseDB(cfg.Supabase), nil
    case "aws":
        return database.NewAWSPostgresDB(cfg.AWS), nil  // New
    default:
        return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider.Database)
    }
}
```

#### **3. Update Configuration**
```bash
# configs/production.env
PROVIDER_DATABASE=aws  # Changed from supabase
AWS_REGION=us-west-2
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret
```

#### **4. Deploy and Test**
```bash
# Deploy with new configuration
kubectl apply -f deployments/kubernetes/

# Test functionality
go test ./... -v

# Switch traffic
kubectl patch service kyb-platform-service -p '{"spec":{"selector":{"version":"green"}}}'
```

### **AWS → GCP Migration**

#### **1. Add GCP Implementation**
```go
// internal/database/gcp_postgres.go
type GCPPostgresDB struct {
    db     *sql.DB
    config *config.GCPConfig
}

func (p *GCPPostgresDB) Connect(ctx context.Context) error {
    // GCP Cloud SQL connection logic
}
```

#### **2. Update Factory**
```go
// internal/factory.go
func NewDatabase(cfg *config.Config) (interfaces.Database, error) {
    switch cfg.Provider.Database {
    case "supabase":
        return database.NewSupabaseDB(cfg.Supabase), nil
    case "aws":
        return database.NewAWSPostgresDB(cfg.AWS), nil
    case "gcp":
        return database.NewGCPPostgresDB(cfg.GCP), nil  // New
    default:
        return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider.Database)
    }
}
```

#### **3. Update Configuration**
```bash
# configs/production.env
PROVIDER_DATABASE=gcp  # Changed from aws
GCP_PROJECT_ID=your-project
GCP_REGION=us-central1
GCP_CREDENTIALS_FILE=/path/to/credentials.json
```

---

## 🏁 **Conclusion**

The provider migration preservation strategy ensures that:

**Code Preservation:**
- ✅ **100% of business logic** is preserved
- ✅ **Zero API changes** required
- ✅ **All tests continue** to work
- ✅ **Documentation remains** valid

**Effort Reduction:**
- ✅ **90% less migration effort** compared to traditional approaches
- ✅ **2-4 weeks** instead of 6-12 months
- ✅ **Configuration changes** instead of code rewrites
- ✅ **Risk-free migrations** with rollback capability

**Future Flexibility:**
- ✅ **Easy to add new providers**
- ✅ **Simple provider switching**
- ✅ **Environment-specific** provider selection
- ✅ **Gradual migration** capability

**Key Success Factors:**
1. **Interface-first design** - Define stable contracts
2. **Factory pattern** - Abstract provider selection
3. **Configuration-driven** - Environment-based switching
4. **Provider-agnostic tests** - Tests work with any provider
5. **Blue-green deployment** - Zero-downtime migrations

This approach transforms provider migrations from major rewrites into simple configuration changes, preserving all your hard work and enabling future flexibility.
