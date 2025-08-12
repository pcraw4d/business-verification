# KYB Platform - AWS Migration Strategy

## 🎯 **Executive Summary**

This document outlines the strategy for migrating the KYB platform from Supabase back to AWS as the client base grows and enterprise requirements increase. The strategy focuses on preserving existing work, maintaining abstraction layers, and ensuring a smooth transition with minimal disruption.

**Migration Trigger**: Client base growth requiring enterprise features, higher performance, or specific AWS services
**Timeline**: 6-8 weeks for complete migration
**Goal**: Seamless transition with preserved functionality and improved scalability

---

## 📊 **When to Migrate Back to AWS**

### **Business Triggers**

#### **Client Base Growth**
```yaml
Migration Thresholds:
  - Active Users: > 10,000 users
  - Monthly Revenue: > $50,000
  - Enterprise Customers: > 10 customers
  - Data Volume: > 1TB stored data
  - API Requests: > 10M requests/month
```

#### **Feature Requirements**
```yaml
AWS-Specific Features Needed:
  - Advanced Analytics: Amazon QuickSight, Redshift
  - Machine Learning: SageMaker, Comprehend
  - Advanced Security: GuardDuty, Security Hub
  - Compliance: AWS Artifact, Config
  - Global Distribution: CloudFront, Route 53
  - High Availability: Multi-region deployment
```

#### **Performance Requirements**
```yaml
Performance Thresholds:
  - Response Time: < 100ms (95th percentile)
  - Throughput: > 10,000 requests/second
  - Concurrent Users: > 50,000
  - Data Processing: > 1M records/hour
  - Real-time Analytics: < 1 second latency
```

#### **Cost Considerations**
```yaml
Cost Analysis:
  - Supabase Pro: $25/month (current)
  - Supabase Enterprise: $500+/month
  - AWS Equivalent: $200-500/month
  - Break-even Point: 5,000+ active users
```

---

## 🏗️ **Architecture Preservation Strategy**

### **Abstraction Layer Design**

#### **Database Abstraction**
```go
// internal/database/interface.go - Preserve existing interface
type Database interface {
    Connect(ctx context.Context) error
    Close() error
    Ping(ctx context.Context) error
    BeginTx(ctx context.Context) (Database, error)
    Commit() error
    Rollback() error
    // ... existing methods
}

// internal/database/supabase.go - Supabase implementation
type SupabaseDB struct {
    client *supabase.Client
    config *config.SupabaseConfig
}

// internal/database/aws_postgres.go - AWS RDS implementation
type AWSPostgresDB struct {
    db     *sql.DB
    config *config.DatabaseConfig
}

// internal/database/factory.go - Factory pattern
func NewDatabase(cfg *config.DatabaseConfig) (Database, error) {
    switch cfg.Provider {
    case "supabase":
        return NewSupabaseDB(cfg.Supabase), nil
    case "aws":
        return NewAWSPostgresDB(cfg), nil
    default:
        return nil, fmt.Errorf("unsupported database provider: %s", cfg.Provider)
    }
}
```

#### **Authentication Abstraction**
```go
// internal/auth/interface.go - Preserve existing interface
type AuthService interface {
    SignUp(ctx context.Context, email, password string) (*User, error)
    SignIn(ctx context.Context, email, password string) (*AuthResponse, error)
    ValidateToken(ctx context.Context, token string) (*User, error)
    RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
    // ... existing methods
}

// internal/auth/supabase_auth.go - Supabase implementation
type SupabaseAuthService struct {
    client *supabase.Client
    config *config.SupabaseConfig
}

// internal/auth/aws_auth.go - AWS Cognito implementation
type AWSAuthService struct {
    cognitoClient *cognitoidentityprovider.CognitoIdentityProvider
    config        *config.AWSConfig
}

// internal/auth/factory.go - Factory pattern
func NewAuthService(cfg *config.AuthConfig) (AuthService, error) {
    switch cfg.Provider {
    case "supabase":
        return NewSupabaseAuthService(cfg.Supabase), nil
    case "aws":
        return NewAWSAuthService(cfg.AWS), nil
    default:
        return nil, fmt.Errorf("unsupported auth provider: %s", cfg.Provider)
    }
}
```

#### **Caching Abstraction**
```go
// internal/cache/interface.go - Preserve existing interface
type Cache interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Clear(ctx context.Context) error
}

// internal/cache/supabase_cache.go - Supabase real-time implementation
type SupabaseCache struct {
    client *supabase.Client
}

// internal/cache/aws_redis.go - AWS ElastiCache implementation
type AWSRedisCache struct {
    client redis.Client
    config *config.RedisConfig
}

// internal/cache/factory.go - Factory pattern
func NewCache(cfg *config.CacheConfig) (Cache, error) {
    switch cfg.Provider {
    case "supabase":
        return NewSupabaseCache(cfg.Supabase), nil
    case "aws":
        return NewAWSRedisCache(cfg.AWS), nil
    default:
        return nil, fmt.Errorf("unsupported cache provider: %s", cfg.Provider)
    }
}
```

### **Configuration Management**

#### **Multi-Provider Configuration**
```go
// internal/config/config.go - Enhanced configuration
type Config struct {
    Environment string `json:"environment" yaml:"environment"`
    
    // Provider selection
    Provider ProviderConfig `json:"provider" yaml:"provider"`
    
    // Provider-specific configurations
    Supabase SupabaseConfig `json:"supabase" yaml:"supabase"`
    AWS      AWSConfig      `json:"aws" yaml:"aws"`
    
    // Common configurations
    Server       ServerConfig       `json:"server" yaml:"server"`
    Observability ObservabilityConfig `json:"observability" yaml:"observability"`
    ExternalServices ExternalServicesConfig `json:"external_services" yaml:"external_services"`
}

type ProviderConfig struct {
    Database string `json:"database" yaml:"database"` // "supabase" or "aws"
    Auth     string `json:"auth" yaml:"auth"`         // "supabase" or "aws"
    Cache    string `json:"cache" yaml:"cache"`       // "supabase" or "aws"
    Storage  string `json:"storage" yaml:"storage"`   // "supabase" or "aws"
}

type SupabaseConfig struct {
    URL    string `json:"url" yaml:"url"`
    APIKey string `json:"api_key" yaml:"api_key"`
    JWTSecret string `json:"jwt_secret" yaml:"jwt_secret"`
}

type AWSConfig struct {
    Region          string `json:"region" yaml:"region"`
    AccessKeyID     string `json:"access_key_id" yaml:"access_key_id"`
    SecretAccessKey string `json:"secret_access_key" yaml:"secret_access_key"`
    
    // RDS Configuration
    RDS RDSConfig `json:"rds" yaml:"rds"`
    
    // ElastiCache Configuration
    ElastiCache ElastiCacheConfig `json:"elasticache" yaml:"elasticache"`
    
    // Cognito Configuration
    Cognito CognitoConfig `json:"cognito" yaml:"cognito"`
    
    // S3 Configuration
    S3 S3Config `json:"s3" yaml:"s3"`
}
```

#### **Environment-Specific Configuration**
```bash
# configs/development.env - Development with Supabase
PROVIDER_DATABASE=supabase
PROVIDER_AUTH=supabase
PROVIDER_CACHE=supabase
PROVIDER_STORAGE=supabase

SUPABASE_URL=https://your-dev-project.supabase.co
SUPABASE_API_KEY=your_dev_supabase_anon_key

# configs/production.env - Production with AWS
PROVIDER_DATABASE=aws
PROVIDER_AUTH=aws
PROVIDER_CACHE=aws
PROVIDER_STORAGE=aws

AWS_REGION=us-west-2
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key

# RDS Configuration
AWS_RDS_HOST=kyb-platform-db.cluster-xyz.us-west-2.rds.amazonaws.com
AWS_RDS_PORT=5432
AWS_RDS_DATABASE=kyb_platform
AWS_RDS_USERNAME=kyb_user
AWS_RDS_PASSWORD=your_rds_password

# ElastiCache Configuration
AWS_ELASTICACHE_HOST=kyb-platform-redis.xyz.cache.amazonaws.com
AWS_ELASTICACHE_PORT=6379
AWS_ELASTICACHE_PASSWORD=your_redis_password

# Cognito Configuration
AWS_COGNITO_USER_POOL_ID=us-west-2_xxxxxxxxx
AWS_COGNITO_CLIENT_ID=your_cognito_client_id
AWS_COGNITO_CLIENT_SECRET=your_cognito_client_secret
```

---

## 🔄 **Migration Strategy**

### **Phase 1: Infrastructure Preparation (Week 1-2)**

#### **AWS Infrastructure Setup**
```hcl
# deployments/terraform/aws/main.tf - AWS infrastructure
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# VPC and Networking
module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
  version = "5.0.0"
  
  name = "kyb-platform-vpc"
  cidr = var.vpc_cidr
  
  azs             = var.availability_zones
  private_subnets = var.private_subnet_cidrs
  public_subnets  = var.public_subnet_cidrs
  
  enable_nat_gateway = true
  single_nat_gateway = false
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
  }
}

# RDS PostgreSQL
module "rds" {
  source  = "terraform-aws-modules/rds/aws"
  version = "~> 6.0"
  
  identifier = "kyb-platform-db"
  
  engine               = "postgres"
  engine_version       = "15.4"
  instance_class       = "db.t3.micro"
  allocated_storage    = 20
  max_allocated_storage = 100
  
  db_name  = "kyb_platform"
  username = var.db_username
  port     = "5432"
  
  vpc_security_group_ids = [aws_security_group.rds.id]
  subnet_ids             = module.vpc.private_subnets
  
  backup_retention_period = 7
  deletion_protection = true
  
  performance_insights_enabled = true
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
  }
}

# ElastiCache Redis
resource "aws_elasticache_replication_group" "redis" {
  replication_group_id       = "kyb-platform-redis"
  replication_group_description = "KYB Platform Redis cluster"
  
  node_type                  = "cache.t3.micro"
  port                       = 6379
  parameter_group_name       = aws_elasticache_parameter_group.redis.name
  subnet_group_name          = aws_elasticache_subnet_group.redis.name
  security_group_ids         = [aws_security_group.redis.id]
  
  num_cache_clusters = 1
  
  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
  }
}

# Cognito User Pool
resource "aws_cognito_user_pool" "main" {
  name = "kyb-platform-users"
  
  password_policy {
    minimum_length    = 8
    require_lowercase = true
    require_numbers   = true
    require_symbols   = true
    require_uppercase = true
  }
  
  auto_verified_attributes = ["email"]
  
  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
  }
  
  email_configuration {
    email_sending_account = "COGNITO_DEFAULT"
  }
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
  }
}

resource "aws_cognito_user_pool_client" "main" {
  name         = "kyb-platform-client"
  user_pool_id = aws_cognito_user_pool.main.id
  
  generate_secret = true
  
  explicit_auth_flows = [
    "ALLOW_USER_PASSWORD_AUTH",
    "ALLOW_REFRESH_TOKEN_AUTH",
    "ALLOW_USER_SRP_AUTH"
  ]
}
```

#### **Data Migration Strategy**
```bash
# scripts/migrate_data.sh - Data migration script
#!/bin/bash

set -e

echo "Starting data migration from Supabase to AWS..."

# Export data from Supabase
echo "Exporting data from Supabase..."
pg_dump "postgresql://postgres:${SUPABASE_DB_PASSWORD}@${SUPABASE_DB_HOST}:5432/postgres" \
  --schema=public \
  --data-only \
  --no-owner \
  --no-privileges \
  > supabase_data_dump.sql

# Import data to AWS RDS
echo "Importing data to AWS RDS..."
psql "postgresql://${AWS_RDS_USERNAME}:${AWS_RDS_PASSWORD}@${AWS_RDS_HOST}:${AWS_RDS_PORT}/${AWS_RDS_DATABASE}" \
  < supabase_data_dump.sql

echo "Data migration completed successfully!"
```

### **Phase 2: Implementation Updates (Week 3-4)**

#### **AWS Service Implementations**
```go
// internal/database/aws_postgres.go - AWS RDS implementation
package database

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    _ "github.com/lib/pq"
)

type AWSPostgresDB struct {
    db     *sql.DB
    config *config.DatabaseConfig
    tx     *sql.Tx
}

func NewAWSPostgresDB(cfg *config.DatabaseConfig) *AWSPostgresDB {
    return &AWSPostgresDB{
        config: cfg,
    }
}

func (p *AWSPostgresDB) Connect(ctx context.Context) error {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        p.config.Host,
        p.config.Port,
        p.config.Username,
        p.config.Password,
        p.config.Database,
        p.config.SSLMode,
    )
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return fmt.Errorf("failed to open database connection: %w", err)
    }
    
    // Configure connection pool for AWS RDS
    db.SetMaxOpenConns(p.config.MaxOpenConns)
    db.SetMaxIdleConns(p.config.MaxIdleConns)
    db.SetConnMaxLifetime(p.config.ConnMaxLifetime)
    
    // Test the connection
    if err := db.PingContext(ctx); err != nil {
        return fmt.Errorf("failed to ping database: %w", err)
    }
    
    p.db = db
    return nil
}

// Implement all interface methods...
func (p *AWSPostgresDB) Close() error {
    if p.db != nil {
        return p.db.Close()
    }
    return nil
}

func (p *AWSPostgresDB) Ping(ctx context.Context) error {
    if p.db == nil {
        return fmt.Errorf("database not connected")
    }
    return p.db.PingContext(ctx)
}

func (p *AWSPostgresDB) BeginTx(ctx context.Context) (Database, error) {
    if p.db == nil {
        return nil, fmt.Errorf("database not connected")
    }
    
    tx, err := p.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    
    return &AWSPostgresDB{
        db:     p.db,
        config: p.config,
        tx:     tx,
    }, nil
}

func (p *AWSPostgresDB) Commit() error {
    if p.tx == nil {
        return fmt.Errorf("no active transaction")
    }
    return p.tx.Commit()
}

func (p *AWSPostgresDB) Rollback() error {
    if p.tx == nil {
        return fmt.Errorf("no active transaction")
    }
    return p.tx.Rollback()
}
```

```go
// internal/auth/aws_auth.go - AWS Cognito implementation
package auth

import (
    "context"
    "fmt"
    "time"
    
    "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
    "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type AWSAuthService struct {
    cognitoClient *cognitoidentityprovider.Client
    config        *config.AWSConfig
}

func NewAWSAuthService(cfg *config.AWSConfig) (*AWSAuthService, error) {
    client := cognitoidentityprovider.New(cognitoidentityprovider.Options{
        Region: cfg.Region,
    })
    
    return &AWSAuthService{
        cognitoClient: client,
        config:        cfg,
    }, nil
}

func (a *AWSAuthService) SignUp(ctx context.Context, email, password string) (*User, error) {
    input := &cognitoidentityprovider.SignUpInput{
        ClientId: &a.config.Cognito.ClientID,
        Username: &email,
        Password: &password,
        UserAttributes: []types.AttributeType{
            {
                Name:  aws.String("email"),
                Value: &email,
            },
        },
    }
    
    result, err := a.cognitoClient.SignUp(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("failed to sign up user: %w", err)
    }
    
    return &User{
        ID:    *result.UserSub,
        Email: email,
    }, nil
}

func (a *AWSAuthService) SignIn(ctx context.Context, email, password string) (*AuthResponse, error) {
    input := &cognitoidentityprovider.InitiateAuthInput{
        ClientId: &a.config.Cognito.ClientID,
        AuthFlow: types.AuthFlowTypeUserPasswordAuth,
        AuthParameters: map[string]string{
            "USERNAME": email,
            "PASSWORD": password,
        },
    }
    
    result, err := a.cognitoClient.InitiateAuth(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("failed to sign in user: %w", err)
    }
    
    return &AuthResponse{
        AccessToken:  *result.AuthenticationResult.AccessToken,
        RefreshToken: *result.AuthenticationResult.RefreshToken,
        User: &User{
            ID:    email, // Use email as ID for consistency
            Email: email,
        },
    }, nil
}

func (a *AWSAuthService) ValidateToken(ctx context.Context, token string) (*User, error) {
    input := &cognitoidentityprovider.GetUserInput{
        AccessToken: &token,
    }
    
    result, err := a.cognitoClient.GetUser(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("failed to validate token: %w", err)
    }
    
    var email string
    for _, attr := range result.UserAttributes {
        if *attr.Name == "email" {
            email = *attr.Value
            break
        }
    }
    
    return &User{
        ID:    *result.Username,
        Email: email,
    }, nil
}
```

```go
// internal/cache/aws_redis.go - AWS ElastiCache implementation
package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/redis/go-redis/v9"
)

type AWSRedisCache struct {
    client redis.Client
    config *config.RedisConfig
}

func NewAWSRedisCache(cfg *config.RedisConfig) *AWSRedisCache {
    client := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Password: cfg.Password,
        DB:       cfg.DB,
    })
    
    return &AWSRedisCache{
        client: *client,
        config: cfg,
    }
}

func (r *AWSRedisCache) Get(ctx context.Context, key string) (interface{}, error) {
    val, err := r.client.Get(ctx, key).Result()
    if err != nil {
        return nil, err
    }
    
    var result interface{}
    if err := json.Unmarshal([]byte(val), &result); err != nil {
        return nil, err
    }
    
    return result, nil
}

func (r *AWSRedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    return r.client.Set(ctx, key, string(data), ttl).Err()
}

func (r *AWSRedisCache) Delete(ctx context.Context, key string) error {
    return r.client.Del(ctx, key).Err()
}

func (r *AWSRedisCache) Clear(ctx context.Context) error {
    return r.client.FlushDB(ctx).Err()
}
```

### **Phase 3: Testing and Validation (Week 5-6)**

#### **Comprehensive Testing Strategy**
```go
// test/migration_test.go - Migration testing
package test

import (
    "context"
    "testing"
    "time"
    
    "github.com/pcraw4d/business-verification/internal/config"
    "github.com/pcraw4d/business-verification/internal/database"
    "github.com/pcraw4d/business-verification/internal/auth"
    "github.com/pcraw4d/business-verification/internal/cache"
)

func TestProviderMigration(t *testing.T) {
    tests := []struct {
        name     string
        provider string
        config   *config.Config
    }{
        {
            name:     "Supabase Provider",
            provider: "supabase",
            config: &config.Config{
                Provider: config.ProviderConfig{
                    Database: "supabase",
                    Auth:     "supabase",
                    Cache:    "supabase",
                },
                Supabase: config.SupabaseConfig{
                    URL:    "https://test.supabase.co",
                    APIKey: "test_key",
                },
            },
        },
        {
            name:     "AWS Provider",
            provider: "aws",
            config: &config.Config{
                Provider: config.ProviderConfig{
                    Database: "aws",
                    Auth:     "aws",
                    Cache:    "aws",
                },
                AWS: config.AWSConfig{
                    Region: "us-west-2",
                    RDS: config.RDSConfig{
                        Host:     "test-rds.amazonaws.com",
                        Port:     5432,
                        Database: "test_db",
                        Username: "test_user",
                        Password: "test_password",
                    },
                    ElastiCache: config.ElastiCacheConfig{
                        Host:     "test-redis.amazonaws.com",
                        Port:     6379,
                        Password: "test_password",
                    },
                    Cognito: config.CognitoConfig{
                        UserPoolID:     "us-west-2_testpool",
                        ClientID:       "test_client_id",
                        ClientSecret:   "test_client_secret",
                    },
                },
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test database provider
            db, err := database.NewDatabase(tt.config)
            if err != nil {
                t.Fatalf("Failed to create database: %v", err)
            }
            
            if err := db.Connect(context.Background()); err != nil {
                t.Fatalf("Failed to connect to database: %v", err)
            }
            defer db.Close()
            
            // Test auth provider
            authService, err := auth.NewAuthService(tt.config)
            if err != nil {
                t.Fatalf("Failed to create auth service: %v", err)
            }
            
            // Test cache provider
            cache, err := cache.NewCache(tt.config)
            if err != nil {
                t.Fatalf("Failed to create cache: %v", err)
            }
            
            // Test basic operations
            testKey := "test_key"
            testValue := "test_value"
            
            if err := cache.Set(context.Background(), testKey, testValue, time.Minute); err != nil {
                t.Fatalf("Failed to set cache value: %v", err)
            }
            
            val, err := cache.Get(context.Background(), testKey)
            if err != nil {
                t.Fatalf("Failed to get cache value: %v", err)
            }
            
            if val != testValue {
                t.Errorf("Expected %s, got %v", testValue, val)
            }
        })
    }
}
```

### **Phase 4: Deployment and Cutover (Week 7-8)**

#### **Blue-Green Deployment Strategy**
```yaml
# deployments/kubernetes/blue-green-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kyb-platform-blue
  namespace: kyb-platform
spec:
  replicas: 3
  selector:
    matchLabels:
      app: kyb-platform
      version: blue
  template:
    metadata:
      labels:
        app: kyb-platform
        version: blue
    spec:
      containers:
      - name: kyb-platform
        image: kyb-platform:blue
        env:
        - name: PROVIDER_DATABASE
          value: "supabase"
        - name: PROVIDER_AUTH
          value: "supabase"
        - name: PROVIDER_CACHE
          value: "supabase"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kyb-platform-green
  namespace: kyb-platform
spec:
  replicas: 0  # Start with 0 replicas
  selector:
    matchLabels:
      app: kyb-platform
      version: green
  template:
    metadata:
      labels:
        app: kyb-platform
        version: green
    spec:
      containers:
      - name: kyb-platform
        image: kyb-platform:green
        env:
        - name: PROVIDER_DATABASE
          value: "aws"
        - name: PROVIDER_AUTH
          value: "aws"
        - name: PROVIDER_CACHE
          value: "aws"
---
apiVersion: v1
kind: Service
metadata:
  name: kyb-platform-service
  namespace: kyb-platform
spec:
  selector:
    app: kyb-platform
    version: blue  # Start with blue
  ports:
  - port: 80
    targetPort: 8080
```

#### **Cutover Script**
```bash
#!/bin/bash
# scripts/cutover.sh - Blue-green deployment cutover

set -e

echo "Starting blue-green deployment cutover..."

# Step 1: Deploy green environment
echo "Deploying green environment..."
kubectl scale deployment kyb-platform-green --replicas=3 -n kyb-platform

# Step 2: Wait for green to be ready
echo "Waiting for green environment to be ready..."
kubectl rollout status deployment kyb-platform-green -n kyb-platform

# Step 3: Run health checks on green
echo "Running health checks on green environment..."
kubectl port-forward service/kyb-platform-green 8080:80 -n kyb-platform &
PORT_FORWARD_PID=$!

sleep 10

# Health check
if curl -f http://localhost:8080/health; then
    echo "Green environment health check passed"
else
    echo "Green environment health check failed"
    kill $PORT_FORWARD_PID
    exit 1
fi

kill $PORT_FORWARD_PID

# Step 4: Switch traffic to green
echo "Switching traffic to green environment..."
kubectl patch service kyb-platform-service -n kyb-platform -p '{"spec":{"selector":{"version":"green"}}}'

# Step 5: Scale down blue
echo "Scaling down blue environment..."
kubectl scale deployment kyb-platform-blue --replicas=0 -n kyb-platform

echo "Cutover completed successfully!"
```

---

## 🔧 **Preservation Strategies**

### **Code Preservation**

#### **Interface-First Design**
```go
// Preserve all existing interfaces
// internal/interfaces.go - Central interface definitions
package interfaces

import (
    "context"
    "time"
)

// Database interface - preserved from original implementation
type Database interface {
    Connect(ctx context.Context) error
    Close() error
    Ping(ctx context.Context) error
    BeginTx(ctx context.Context) (Database, error)
    Commit() error
    Rollback() error
    // ... all existing methods
}

// AuthService interface - preserved from original implementation
type AuthService interface {
    SignUp(ctx context.Context, email, password string) (*User, error)
    SignIn(ctx context.Context, email, password string) (*AuthResponse, error)
    ValidateToken(ctx context.Context, token string) (*User, error)
    RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
    // ... all existing methods
}

// Cache interface - preserved from original implementation
type Cache interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Clear(ctx context.Context) error
}
```

#### **Factory Pattern Preservation**
```go
// internal/factory.go - Preserved factory pattern
package factory

import (
    "github.com/pcraw4d/business-verification/internal/config"
    "github.com/pcraw4d/business-verification/internal/database"
    "github.com/pcraw4d/business-verification/internal/auth"
    "github.com/pcraw4d/business-verification/internal/cache"
)

// Preserved factory functions
func NewDatabase(cfg *config.Config) (database.Database, error) {
    switch cfg.Provider.Database {
    case "supabase":
        return database.NewSupabaseDB(cfg.Supabase), nil
    case "aws":
        return database.NewAWSPostgresDB(cfg.AWS), nil
    default:
        return nil, fmt.Errorf("unsupported database provider: %s", cfg.Provider.Database)
    }
}

func NewAuthService(cfg *config.Config) (auth.AuthService, error) {
    switch cfg.Provider.Auth {
    case "supabase":
        return auth.NewSupabaseAuthService(cfg.Supabase), nil
    case "aws":
        return auth.NewAWSAuthService(cfg.AWS), nil
    default:
        return nil, fmt.Errorf("unsupported auth provider: %s", cfg.Provider.Auth)
    }
}

func NewCache(cfg *config.Config) (cache.Cache, error) {
    switch cfg.Provider.Cache {
    case "supabase":
        return cache.NewSupabaseCache(cfg.Supabase), nil
    case "aws":
        return cache.NewAWSRedisCache(cfg.AWS), nil
    default:
        return nil, fmt.Errorf("unsupported cache provider: %s", cfg.Provider.Cache)
    }
}
```

### **Configuration Preservation**

#### **Environment-Based Provider Selection**
```bash
# configs/development.env - Development with Supabase
PROVIDER_DATABASE=supabase
PROVIDER_AUTH=supabase
PROVIDER_CACHE=supabase
PROVIDER_STORAGE=supabase

# Supabase configuration
SUPABASE_URL=https://your-dev-project.supabase.co
SUPABASE_API_KEY=your_dev_supabase_anon_key

# configs/staging.env - Staging with AWS
PROVIDER_DATABASE=aws
PROVIDER_AUTH=aws
PROVIDER_CACHE=aws
PROVIDER_STORAGE=aws

# AWS configuration
AWS_REGION=us-west-2
AWS_ACCESS_KEY_ID=your_staging_aws_access_key
AWS_SECRET_ACCESS_KEY=your_staging_aws_secret_key

# configs/production.env - Production with AWS
PROVIDER_DATABASE=aws
PROVIDER_AUTH=aws
PROVIDER_CACHE=aws
PROVIDER_STORAGE=aws

# AWS configuration
AWS_REGION=us-west-2
AWS_ACCESS_KEY_ID=your_production_aws_access_key
AWS_SECRET_ACCESS_KEY=your_production_aws_secret_key
```

### **Testing Preservation**

#### **Provider-Agnostic Tests**
```go
// test/provider_agnostic_test.go - Tests that work with any provider
package test

import (
    "context"
    "testing"
    "time"
    
    "github.com/pcraw4d/business-verification/internal/config"
    "github.com/pcraw4d/business-verification/internal/factory"
)

func TestDatabaseOperations(t *testing.T) {
    // This test works with any database provider
    cfg := loadTestConfig()
    
    db, err := factory.NewDatabase(cfg)
    if err != nil {
        t.Fatalf("Failed to create database: %v", err)
    }
    
    if err := db.Connect(context.Background()); err != nil {
        t.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()
    
    // Test database operations
    // These tests work regardless of the underlying provider
}

func TestAuthOperations(t *testing.T) {
    // This test works with any auth provider
    cfg := loadTestConfig()
    
    authService, err := factory.NewAuthService(cfg)
    if err != nil {
        t.Fatalf("Failed to create auth service: %v", err)
    }
    
    // Test auth operations
    // These tests work regardless of the underlying provider
}

func TestCacheOperations(t *testing.T) {
    // This test works with any cache provider
    cfg := loadTestConfig()
    
    cache, err := factory.NewCache(cfg)
    if err != nil {
        t.Fatalf("Failed to create cache: %v", err)
    }
    
    // Test cache operations
    // These tests work regardless of the underlying provider
}
```

---

## 📊 **Cost Analysis and ROI**

### **Migration Cost Breakdown**

#### **Development Costs**
```yaml
Development Effort:
  - Infrastructure Setup: 2 weeks
  - Implementation: 2 weeks
  - Testing: 1 week
  - Deployment: 1 week
  - Total: 6 weeks

Development Team:
  - Senior Backend Developer: $8,000/week
  - DevOps Engineer: $7,000/week
  - QA Engineer: $5,000/week
  - Total: $20,000/week

Total Development Cost: $120,000
```

#### **Infrastructure Costs**
```yaml
AWS Infrastructure (Monthly):
  - EC2: $60
  - RDS: $15
  - ElastiCache: $15
  - Load Balancer: $20
  - CloudWatch: $10
  - S3/CloudFront: $15
  - Other: $65
  - Total: $200/month

Supabase Infrastructure (Monthly):
  - Supabase Pro: $25
  - Custom Domain: $1
  - Monitoring: $5
  - Total: $31/month

Cost Difference: $169/month ($2,028/year)
```

#### **ROI Calculation**
```yaml
Break-even Analysis:
  - Development Cost: $120,000
  - Annual Infrastructure Savings: $2,028
  - Break-even Time: 59 years

However, this doesn't account for:
  - Enterprise features needed
  - Performance requirements
  - Compliance requirements
  - Scalability needs

Realistic ROI Factors:
  - Enterprise customer acquisition: $50,000-100,000/year per customer
  - Performance improvements: 20-50% better response times
  - Compliance certifications: Required for enterprise sales
  - Scalability: Support for 10x more users
```

### **Business Justification**

#### **Enterprise Requirements**
```yaml
Enterprise Features Needed:
  - Advanced Security: AWS GuardDuty, Security Hub
  - Compliance: SOC 2, PCI DSS, GDPR
  - High Availability: 99.99% uptime SLA
  - Global Distribution: Multi-region deployment
  - Advanced Analytics: QuickSight, Redshift
  - Machine Learning: SageMaker integration
  - Custom Integrations: Enterprise SSO, LDAP
  - Dedicated Support: 24/7 enterprise support
```

#### **Performance Requirements**
```yaml
Performance Thresholds:
  - Response Time: < 100ms (95th percentile)
  - Throughput: > 10,000 requests/second
  - Concurrent Users: > 50,000
  - Data Processing: > 1M records/hour
  - Real-time Analytics: < 1 second latency
```

---

## 🚨 **Risk Mitigation**

### **Technical Risks**

#### **Data Migration Risks**
```yaml
Risk: Data loss during migration
Mitigation:
  - Comprehensive backup strategy
  - Incremental migration approach
  - Data validation at each step
  - Rollback procedures
  - Parallel running during transition
```

#### **Performance Risks**
```yaml
Risk: Performance degradation
Mitigation:
  - Performance testing before cutover
  - Load testing with production data
  - Monitoring during transition
  - Gradual traffic migration
  - Rollback capability
```

#### **Compatibility Risks**
```yaml
Risk: Feature incompatibility
Mitigation:
  - Feature parity testing
  - API compatibility testing
  - Client application testing
  - Documentation updates
  - Training for development team
```

### **Business Risks**

#### **Downtime Risks**
```yaml
Risk: Service interruption during migration
Mitigation:
  - Blue-green deployment strategy
  - Zero-downtime migration
  - Rollback procedures
  - Communication plan
  - Customer notification
```

#### **Cost Overrun Risks**
```yaml
Risk: Migration costs exceeding budget
Mitigation:
  - Detailed cost planning
  - Phased migration approach
  - Regular cost monitoring
  - Contingency budget
  - ROI tracking
```

---

## 📋 **Migration Checklist**

### **Pre-Migration Tasks**

- [ ] **Business Justification**
  - [ ] Document enterprise requirements
  - [ ] Calculate ROI and break-even
  - [ ] Get stakeholder approval
  - [ ] Set migration timeline

- [ ] **Technical Preparation**
  - [ ] Set up AWS infrastructure
  - [ ] Configure AWS services (RDS, ElastiCache, Cognito)
  - [ ] Set up monitoring and alerting
  - [ ] Prepare data migration scripts

- [ ] **Development Preparation**
  - [ ] Implement AWS service adapters
  - [ ] Update configuration management
  - [ ] Create provider abstraction layers
  - [ ] Update factory patterns

### **Migration Tasks**

- [ ] **Data Migration**
  - [ ] Export data from Supabase
  - [ ] Import data to AWS RDS
  - [ ] Verify data integrity
  - [ ] Test all database operations

- [ ] **Application Migration**
  - [ ] Deploy with AWS providers
  - [ ] Test all functionality
  - [ ] Performance testing
  - [ ] Security testing

- [ ] **Traffic Migration**
  - [ ] Blue-green deployment
  - [ ] Gradual traffic shift
  - [ ] Monitor performance
  - [ ] Validate functionality

### **Post-Migration Tasks**

- [ ] **Validation**
  - [ ] Complete functionality testing
  - [ ] Performance validation
  - [ ] Security validation
  - [ ] User acceptance testing

- [ ] **Optimization**
  - [ ] Performance tuning
  - [ ] Cost optimization
  - [ ] Monitoring setup
  - [ ] Alerting configuration

- [ ] **Documentation**
  - [ ] Update deployment documentation
  - [ ] Update API documentation
  - [ ] Create operational runbooks
  - [ ] Train operations team

---

## 🎯 **Success Metrics**

### **Technical Metrics**

- **Migration Success Rate**: 100% of features working
- **Performance**: Response times within 10% of Supabase performance
- **Uptime**: 99.9% availability maintained
- **Error Rate**: < 0.1% error rate

### **Business Metrics**

- **Cost Efficiency**: Infrastructure costs within budget
- **Feature Parity**: 100% of Supabase features working
- **User Experience**: No degradation in user experience
- **Enterprise Readiness**: All enterprise features available

### **Operational Metrics**

- **Deployment Time**: < 1 hour for new deployments
- **Recovery Time**: < 5 minutes for service recovery
- **Scaling Time**: < 10 minutes for auto-scaling
- **Monitoring Coverage**: 100% of critical metrics

---

## 📚 **Additional Resources**

### **AWS Documentation**
- [AWS RDS PostgreSQL](https://docs.aws.amazon.com/rds/latest/userguide/CHAP_PostgreSQL.html)
- [AWS ElastiCache Redis](https://docs.aws.amazon.com/elasticache/latest/red-ug/)
- [AWS Cognito](https://docs.aws.amazon.com/cognito/latest/developerguide/)
- [AWS Migration Hub](https://docs.aws.amazon.com/migrationhub/)

### **Migration Tools**
- [AWS Database Migration Service](https://docs.aws.amazon.com/dms/)
- [AWS Schema Conversion Tool](https://docs.aws.amazon.com/SchemaConversionTool/)
- [AWS CloudFormation](https://docs.aws.amazon.com/cloudformation/)

### **Best Practices**
- [AWS Well-Architected Framework](https://aws.amazon.com/architecture/well-architected/)
- [AWS Migration Best Practices](https://aws.amazon.com/migration/)
- [AWS Security Best Practices](https://aws.amazon.com/security/security-learning/)

---

## 🏁 **Conclusion**

The AWS migration strategy provides a comprehensive approach to transitioning from Supabase back to AWS as your client base grows. The key to success is preserving existing work through:

**Preservation Strategies:**
- **Interface-first design** that abstracts provider differences
- **Factory patterns** that allow easy provider switching
- **Configuration-driven** provider selection
- **Provider-agnostic tests** that work with any backend

**Migration Benefits:**
- **Enterprise features** for large customers
- **Better performance** and scalability
- **Advanced security** and compliance
- **Global distribution** capabilities
- **Cost optimization** at scale

**Risk Mitigation:**
- **Blue-green deployment** for zero downtime
- **Comprehensive testing** at each stage
- **Rollback procedures** for safety
- **Gradual migration** to minimize risk

The strategy ensures that your existing Supabase implementation is preserved and can be easily switched back if needed, while providing a clear path to AWS for enterprise growth. The abstraction layers mean that future migrations between providers will be much simpler and less risky.

**Recommendation**: Proceed with the AWS migration when enterprise requirements justify the investment, using the preservation strategies to maintain flexibility for future provider changes.
