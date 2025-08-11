# KYB Platform - Architecture Diagrams

This document provides comprehensive architecture diagrams for the KYB Platform using Mermaid syntax. These diagrams illustrate the system design, component relationships, data flow, and deployment architecture.

## Table of Contents

1. [System Overview](#system-overview)
2. [Component Architecture](#component-architecture)
3. [Data Flow Diagrams](#data-flow-diagrams)
4. [API Architecture](#api-architecture)
5. [Database Schema](#database-schema)
6. [Deployment Architecture](#deployment-architecture)
7. [Sequence Diagrams](#sequence-diagrams)
8. [Security Architecture](#security-architecture)
9. [Monitoring Architecture](#monitoring-architecture)

## System Overview

### High-Level System Architecture

```mermaid
graph TB
    subgraph "Client Applications"
        Web[Web Dashboard]
        Mobile[Mobile App]
        API_Client[API Client]
    end
    
    subgraph "API Gateway Layer"
        Gateway[API Gateway]
        LoadBalancer[Load Balancer]
        CDN[CDN]
    end
    
    subgraph "Application Layer"
        Auth[Authentication Service]
        Classify[Classification Service]
        Risk[Risk Assessment Service]
        Compliance[Compliance Service]
        User[User Management Service]
    end
    
    subgraph "Data Layer"
        PostgreSQL[(PostgreSQL)]
        Redis[(Redis Cache)]
        Elastic[Elasticsearch]
    end
    
    subgraph "External Services"
        Data_Providers[Data Providers]
        ML_Models[ML Models]
        Notification[Notification Service]
    end
    
    subgraph "Infrastructure"
        Monitoring[Monitoring Stack]
        Logging[Logging System]
        Backup[Backup System]
    end
    
    Web --> CDN
    Mobile --> CDN
    API_Client --> CDN
    CDN --> LoadBalancer
    LoadBalancer --> Gateway
    
    Gateway --> Auth
    Gateway --> Classify
    Gateway --> Risk
    Gateway --> Compliance
    Gateway --> User
    
    Auth --> PostgreSQL
    Auth --> Redis
    Classify --> PostgreSQL
    Classify --> Redis
    Classify --> Elastic
    Risk --> PostgreSQL
    Risk --> Redis
    Compliance --> PostgreSQL
    Compliance --> Redis
    User --> PostgreSQL
    User --> Redis
    
    Classify --> Data_Providers
    Risk --> ML_Models
    Compliance --> Data_Providers
    
    Auth --> Notification
    Risk --> Notification
    Compliance --> Notification
    
    Auth --> Monitoring
    Classify --> Monitoring
    Risk --> Monitoring
    Compliance --> Monitoring
    User --> Monitoring
    
    PostgreSQL --> Backup
    Redis --> Backup
```

## Component Architecture

### Service Layer Architecture

```mermaid
graph TB
    subgraph "API Gateway"
        Router[Router]
        Middleware[Middleware Stack]
        RateLimit[Rate Limiter]
        Auth[Auth Middleware]
    end
    
    subgraph "Service Layer"
        subgraph "Classification Service"
            ClassifyHandler[Classification Handler]
            ClassifyService[Classification Service]
            ClassifyRepo[Classification Repository]
            ClassifyCache[Classification Cache]
        end
        
        subgraph "Risk Assessment Service"
            RiskHandler[Risk Handler]
            RiskService[Risk Service]
            RiskRepo[Risk Repository]
            RiskCache[Risk Cache]
        end
        
        subgraph "Compliance Service"
            ComplianceHandler[Compliance Handler]
            ComplianceService[Compliance Service]
            ComplianceRepo[Compliance Repository]
            ComplianceCache[Compliance Cache]
        end
        
        subgraph "Authentication Service"
            AuthHandler[Auth Handler]
            AuthService[Auth Service]
            AuthRepo[Auth Repository]
            JWT[JWT Manager]
        end
    end
    
    subgraph "Data Access Layer"
        DB[Database Interface]
        Cache[Cache Interface]
        External[External APIs]
    end
    
    subgraph "Infrastructure"
        Logger[Logger]
        Metrics[Metrics]
        Tracing[Tracing]
        Config[Configuration]
    end
    
    Router --> Middleware
    Middleware --> RateLimit
    RateLimit --> Auth
    
    Auth --> ClassifyHandler
    Auth --> RiskHandler
    Auth --> ComplianceHandler
    Auth --> AuthHandler
    
    ClassifyHandler --> ClassifyService
    RiskHandler --> RiskService
    ComplianceHandler --> ComplianceService
    AuthHandler --> AuthService
    
    ClassifyService --> ClassifyRepo
    ClassifyService --> ClassifyCache
    RiskService --> RiskRepo
    RiskService --> RiskCache
    ComplianceService --> ComplianceRepo
    ComplianceService --> ComplianceCache
    AuthService --> AuthRepo
    AuthService --> JWT
    
    ClassifyRepo --> DB
    RiskRepo --> DB
    ComplianceRepo --> DB
    AuthRepo --> DB
    
    ClassifyCache --> Cache
    RiskCache --> Cache
    ComplianceCache --> Cache
    
    ClassifyService --> External
    RiskService --> External
    ComplianceService --> External
    
    ClassifyService --> Logger
    RiskService --> Logger
    ComplianceService --> Logger
    AuthService --> Logger
    
    ClassifyService --> Metrics
    RiskService --> Metrics
    ComplianceService --> Metrics
    AuthService --> Metrics
    
    ClassifyService --> Tracing
    RiskService --> Tracing
    ComplianceService --> Tracing
    AuthService --> Tracing
    
    ClassifyService --> Config
    RiskService --> Config
    ComplianceService --> Config
    AuthService --> Config
```

## Data Flow Diagrams

### Business Classification Flow

```mermaid
flowchart TD
    Start([Client Request]) --> Validate{Validate Request}
    Validate -->|Invalid| Error[Return Error]
    Validate -->|Valid| Cache{Check Cache}
    
    Cache -->|Hit| Return[Return Cached Result]
    Cache -->|Miss| Normalize[Normalize Business Name]
    
    Normalize --> Keyword[Keyword Classification]
    Normalize --> Fuzzy[Fuzzy Matching]
    Normalize --> Rules[Rule-Based Classification]
    
    Keyword --> Combine[Combine Results]
    Fuzzy --> Combine
    Rules --> Combine
    
    Combine --> Score[Calculate Confidence Score]
    Score --> Select[Select Primary Classification]
    Select --> Store[Store in Database]
    Store --> CacheStore[Store in Cache]
    CacheStore --> Return
    
    Error --> End([End])
    Return --> End
```

### Risk Assessment Flow

```mermaid
flowchart TD
    Start([Business Data]) --> Validate{Validate Data}
    Validate -->|Invalid| Error[Return Error]
    Validate -->|Valid| Factors[Calculate Risk Factors]
    
    Factors --> Financial[Financial Risk]
    Factors --> Operational[Operational Risk]
    Factors --> Compliance[Compliance Risk]
    Factors --> Market[Market Risk]
    
    Financial --> Weight[Apply Weights]
    Operational --> Weight
    Compliance --> Weight
    Market --> Weight
    
    Weight --> Combine[Combine Risk Scores]
    Combine --> Industry[Apply Industry Adjustments]
    Industry --> Threshold{Check Thresholds}
    
    Threshold -->|High Risk| Alert[Generate Alert]
    Threshold -->|Normal Risk| Store[Store Assessment]
    
    Alert --> Store
    Store --> Return[Return Risk Score]
    
    Error --> End([End])
    Return --> End
```

### Compliance Checking Flow

```mermaid
flowchart TD
    Start([Compliance Request]) --> Frameworks{Select Frameworks}
    Frameworks --> SOC2[SOC 2 Framework]
    Frameworks --> PCI[PCI DSS Framework]
    Frameworks --> GDPR[GDPR Framework]
    Frameworks --> Regional[Regional Frameworks]
    
    SOC2 --> Rules[Load Framework Rules]
    PCI --> Rules
    GDPR --> Rules
    Regional --> Rules
    
    Rules --> Evaluate[Evaluate Each Rule]
    Evaluate --> Data{Data Available?}
    Data -->|No| Gap[Mark as Gap]
    Data -->|Yes| Check{Check Compliance}
    
    Check -->|Compliant| Pass[Mark as Pass]
    Check -->|Non-Compliant| Fail[Mark as Fail]
    
    Pass --> Score[Calculate Framework Score]
    Fail --> Score
    Gap --> Score
    
    Score --> Overall[Calculate Overall Score]
    Overall --> Report[Generate Report]
    Report --> Store[Store Results]
    Store --> Return[Return Compliance Status]
    
    Return --> End([End])
```

## API Architecture

### REST API Structure

```mermaid
graph TB
    subgraph "API Gateway"
        Gateway[API Gateway]
        Auth[Authentication]
        RateLimit[Rate Limiting]
        Validation[Request Validation]
        Logging[Request Logging]
    end
    
    subgraph "API Endpoints"
        subgraph "Authentication"
            POST_login[POST /v1/auth/login]
            POST_register[POST /v1/auth/register]
            POST_refresh[POST /v1/auth/refresh]
            POST_logout[POST /v1/auth/logout]
        end
        
        subgraph "Classification"
            POST_classify[POST /v1/classify]
            POST_batch[POST /v1/classify/batch]
            GET_history[GET /v1/classify/history]
            GET_result[GET /v1/classify/{id}]
        end
        
        subgraph "Risk Assessment"
            POST_assess[POST /v1/risk/assess]
            GET_risk[GET /v1/risk/{id}]
            GET_alerts[GET /v1/risk/alerts]
            POST_threshold[POST /v1/risk/thresholds]
        end
        
        subgraph "Compliance"
            POST_check[POST /v1/compliance/check]
            GET_status[GET /v1/compliance/status]
            GET_report[GET /v1/compliance/report]
            POST_framework[POST /v1/compliance/frameworks]
        end
        
        subgraph "User Management"
            GET_profile[GET /v1/users/profile]
            PUT_profile[PUT /v1/users/profile]
            GET_api_keys[GET /v1/users/api-keys]
            POST_api_keys[POST /v1/users/api-keys]
        end
        
        subgraph "System"
            GET_health[GET /health]
            GET_metrics[GET /metrics]
            GET_docs[GET /docs]
        end
    end
    
    Gateway --> Auth
    Auth --> RateLimit
    RateLimit --> Validation
    Validation --> Logging
    
    Logging --> POST_login
    Logging --> POST_register
    Logging --> POST_refresh
    Logging --> POST_logout
    Logging --> POST_classify
    Logging --> POST_batch
    Logging --> GET_history
    Logging --> GET_result
    Logging --> POST_assess
    Logging --> GET_risk
    Logging --> GET_alerts
    Logging --> POST_threshold
    Logging --> POST_check
    Logging --> GET_status
    Logging --> GET_report
    Logging --> POST_framework
    Logging --> GET_profile
    Logging --> PUT_profile
    Logging --> GET_api_keys
    Logging --> POST_api_keys
    Logging --> GET_health
    Logging --> GET_metrics
    Logging --> GET_docs
```

## Database Schema

### Core Database Schema

```mermaid
erDiagram
    users {
        uuid id PK
        string email UK
        string password_hash
        string first_name
        string last_name
        string role
        boolean is_active
        timestamp created_at
        timestamp updated_at
        timestamp last_login
    }
    
    api_keys {
        uuid id PK
        uuid user_id FK
        string key_hash
        string name
        string permissions
        timestamp expires_at
        boolean is_active
        timestamp created_at
        timestamp last_used
    }
    
    classifications {
        uuid id PK
        string business_name
        string primary_code
        string primary_description
        float confidence
        string method
        jsonb metadata
        timestamp created_at
        uuid user_id FK
    }
    
    classification_alternatives {
        uuid id PK
        uuid classification_id FK
        string code
        string description
        float confidence
    }
    
    risk_assessments {
        uuid id PK
        uuid business_id FK
        float overall_score
        jsonb factor_scores
        string risk_level
        jsonb details
        timestamp created_at
        uuid user_id FK
    }
    
    compliance_checks {
        uuid id PK
        uuid business_id FK
        string framework
        float score
        string status
        jsonb gaps
        jsonb recommendations
        timestamp created_at
        uuid user_id FK
    }
    
    audit_logs {
        uuid id PK
        uuid user_id FK
        string action
        string resource_type
        uuid resource_id
        jsonb details
        string ip_address
        timestamp created_at
    }
    
    users ||--o{ api_keys : "has"
    users ||--o{ classifications : "creates"
    users ||--o{ risk_assessments : "creates"
    users ||--o{ compliance_checks : "creates"
    users ||--o{ audit_logs : "generates"
    
    classifications ||--o{ classification_alternatives : "has"
```

## Deployment Architecture

### Production Deployment

```mermaid
graph TB
    subgraph "Load Balancer Layer"
        ALB[Application Load Balancer]
        WAF[Web Application Firewall]
    end
    
    subgraph "Application Layer"
        subgraph "Auto Scaling Group"
            App1[App Instance 1]
            App2[App Instance 2]
            App3[App Instance 3]
            AppN[App Instance N]
        end
    end
    
    subgraph "Data Layer"
        subgraph "Database Cluster"
            Primary[(Primary DB)]
            Replica1[(Read Replica 1)]
            Replica2[(Read Replica 2)]
        end
        
        subgraph "Cache Cluster"
            Redis1[(Redis Primary)]
            Redis2[(Redis Replica)]
        end
        
        subgraph "Search Cluster"
            ES1[(Elasticsearch 1)]
            ES2[(Elasticsearch 2)]
            ES3[(Elasticsearch 3)]
        end
    end
    
    subgraph "Monitoring Layer"
        Prometheus[Prometheus]
        Grafana[Grafana]
        AlertManager[Alert Manager]
        Jaeger[Jaeger]
    end
    
    subgraph "Storage Layer"
        S3[S3 Bucket]
        CloudWatch[CloudWatch Logs]
        Backup[Backup Storage]
    end
    
    Internet --> WAF
    WAF --> ALB
    ALB --> App1
    ALB --> App2
    ALB --> App3
    ALB --> AppN
    
    App1 --> Primary
    App2 --> Primary
    App3 --> Primary
    AppN --> Primary
    
    App1 --> Redis1
    App2 --> Redis1
    App3 --> Redis1
    AppN --> Redis1
    
    App1 --> ES1
    App2 --> ES2
    App3 --> ES3
    AppN --> ES1
    
    Primary --> Replica1
    Primary --> Replica2
    Redis1 --> Redis2
    
    App1 --> Prometheus
    App2 --> Prometheus
    App3 --> Prometheus
    AppN --> Prometheus
    
    Prometheus --> Grafana
    Prometheus --> AlertManager
    
    App1 --> Jaeger
    App2 --> Jaeger
    App3 --> Jaeger
    AppN --> Jaeger
    
    App1 --> S3
    App2 --> S3
    App3 --> S3
    AppN --> S3
    
    App1 --> CloudWatch
    App2 --> CloudWatch
    App3 --> CloudWatch
    AppN --> CloudWatch
    
    Primary --> Backup
    Redis1 --> Backup
    ES1 --> Backup
```

## Sequence Diagrams

### Business Classification Sequence

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant Auth
    participant Classify
    participant Cache
    participant DB
    participant External
    
    Client->>Gateway: POST /v1/classify
    Gateway->>Auth: Validate Token
    Auth-->>Gateway: Token Valid
    Gateway->>Classify: Forward Request
    
    Classify->>Cache: Check Cache
    alt Cache Hit
        Cache-->>Classify: Return Cached Result
        Classify-->>Gateway: Return Result
        Gateway-->>Client: 200 OK
    else Cache Miss
        Cache-->>Classify: Cache Miss
        Classify->>External: Get External Data
        External-->>Classify: External Data
        Classify->>Classify: Perform Classification
        Classify->>DB: Store Result
        Classify->>Cache: Cache Result
        Classify-->>Gateway: Return Result
        Gateway-->>Client: 200 OK
    end
```

### Risk Assessment Sequence

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant Auth
    participant Risk
    participant DB
    participant ML
    participant Alert
    
    Client->>Gateway: POST /v1/risk/assess
    Gateway->>Auth: Validate Token
    Auth-->>Gateway: Token Valid
    Gateway->>Risk: Forward Request
    
    Risk->>DB: Get Business Data
    DB-->>Risk: Business Data
    Risk->>ML: Calculate Risk Factors
    ML-->>Risk: Risk Factors
    Risk->>Risk: Calculate Overall Score
    Risk->>DB: Store Assessment
    
    alt High Risk
        Risk->>Alert: Generate Alert
        Alert-->>Risk: Alert Created
    end
    
    Risk-->>Gateway: Return Assessment
    Gateway-->>Client: 200 OK
```

### Compliance Check Sequence

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant Auth
    participant Compliance
    participant DB
    participant Rules
    participant Report
    
    Client->>Gateway: POST /v1/compliance/check
    Gateway->>Auth: Validate Token
    Auth-->>Gateway: Token Valid
    Gateway->>Compliance: Forward Request
    
    Compliance->>Rules: Load Framework Rules
    Rules-->>Compliance: Framework Rules
    Compliance->>DB: Get Business Data
    DB-->>Compliance: Business Data
    
    loop For Each Rule
        Compliance->>Compliance: Evaluate Rule
        Compliance->>Compliance: Check Compliance
    end
    
    Compliance->>Compliance: Calculate Framework Score
    Compliance->>Report: Generate Report
    Report-->>Compliance: Compliance Report
    Compliance->>DB: Store Results
    Compliance-->>Gateway: Return Results
    Gateway-->>Client: 200 OK
```

## Security Architecture

### Security Layers

```mermaid
graph TB
    subgraph "Network Security"
        WAF[Web Application Firewall]
        DDoS[DDoS Protection]
        VPN[VPN Access]
    end
    
    subgraph "Application Security"
        Auth[Authentication]
        Authz[Authorization]
        RateLimit[Rate Limiting]
        Validation[Input Validation]
        Encryption[Data Encryption]
    end
    
    subgraph "Data Security"
        DB_Encryption[Database Encryption]
        Backup_Encryption[Backup Encryption]
        Key_Management[Key Management]
        Audit[Audit Logging]
    end
    
    subgraph "Infrastructure Security"
        Network_Seg[Network Segmentation]
        Access_Control[Access Control]
        Monitoring[Security Monitoring]
        Incident_Response[Incident Response]
    end
    
    Internet --> WAF
    WAF --> DDoS
    DDoS --> VPN
    
    VPN --> Auth
    Auth --> Authz
    Authz --> RateLimit
    RateLimit --> Validation
    Validation --> Encryption
    
    Encryption --> DB_Encryption
    Encryption --> Backup_Encryption
    DB_Encryption --> Key_Management
    Backup_Encryption --> Key_Management
    
    Key_Management --> Audit
    Audit --> Network_Seg
    Network_Seg --> Access_Control
    Access_Control --> Monitoring
    Monitoring --> Incident_Response
```

## Monitoring Architecture

### Observability Stack

```mermaid
graph TB
    subgraph "Application Layer"
        App1[App Instance 1]
        App2[App Instance 2]
        App3[App Instance 3]
    end
    
    subgraph "Data Collection"
        Prometheus[Prometheus]
        Jaeger[Jaeger]
        Fluentd[Fluentd]
        Filebeat[Filebeat]
    end
    
    subgraph "Data Storage"
        TSDB[Time Series DB]
        Log_Storage[Log Storage]
        Trace_Storage[Trace Storage]
    end
    
    subgraph "Visualization & Alerting"
        Grafana[Grafana]
        AlertManager[Alert Manager]
        Dashboard[Dashboards]
    end
    
    subgraph "External Services"
        PagerDuty[PagerDuty]
        Slack[Slack]
        Email[Email]
    end
    
    App1 --> Prometheus
    App2 --> Prometheus
    App3 --> Prometheus
    
    App1 --> Jaeger
    App2 --> Jaeger
    App3 --> Jaeger
    
    App1 --> Fluentd
    App2 --> Fluentd
    App3 --> Fluentd
    
    App1 --> Filebeat
    App2 --> Filebeat
    App3 --> Filebeat
    
    Prometheus --> TSDB
    Fluentd --> Log_Storage
    Filebeat --> Log_Storage
    Jaeger --> Trace_Storage
    
    TSDB --> Grafana
    Log_Storage --> Grafana
    Trace_Storage --> Grafana
    
    TSDB --> AlertManager
    Log_Storage --> AlertManager
    
    AlertManager --> PagerDuty
    AlertManager --> Slack
    AlertManager --> Email
    
    Grafana --> Dashboard
```

### Metrics Collection

```mermaid
graph LR
    subgraph "Application Metrics"
        HTTP_Metrics[HTTP Metrics]
        Business_Metrics[Business Metrics]
        System_Metrics[System Metrics]
    end
    
    subgraph "Infrastructure Metrics"
        CPU[CPU Usage]
        Memory[Memory Usage]
        Disk[Disk Usage]
        Network[Network Usage]
    end
    
    subgraph "Database Metrics"
        Query_Performance[Query Performance]
        Connection_Pool[Connection Pool]
        Lock_Stats[Lock Statistics]
    end
    
    subgraph "External Service Metrics"
        API_Latency[API Latency]
        Error_Rates[Error Rates]
        Availability[Availability]
    end
    
    HTTP_Metrics --> Prometheus
    Business_Metrics --> Prometheus
    System_Metrics --> Prometheus
    
    CPU --> Prometheus
    Memory --> Prometheus
    Disk --> Prometheus
    Network --> Prometheus
    
    Query_Performance --> Prometheus
    Connection_Pool --> Prometheus
    Lock_Stats --> Prometheus
    
    API_Latency --> Prometheus
    Error_Rates --> Prometheus
    Availability --> Prometheus
    
    Prometheus --> Grafana
    Prometheus --> AlertManager
```

---

## Diagram Usage Guidelines

### Creating New Diagrams

1. **Use Mermaid Syntax**: All diagrams use Mermaid syntax for consistency
2. **Keep Diagrams Focused**: Each diagram should illustrate one specific aspect
3. **Use Clear Labels**: All nodes and edges should have descriptive labels
4. **Maintain Consistency**: Use consistent colors and shapes for similar components
5. **Include Legends**: Add legends for complex diagrams with multiple component types

### Updating Diagrams

1. **Version Control**: Track diagram changes in version control
2. **Documentation**: Update this document when architecture changes
3. **Review Process**: Review diagrams during architecture reviews
4. **Automation**: Consider automated diagram generation from code

### Best Practices

1. **Simplicity**: Keep diagrams simple and easy to understand
2. **Hierarchy**: Use proper hierarchy to show relationships
3. **Flow Direction**: Use consistent flow direction (top-to-bottom or left-to-right)
4. **Grouping**: Group related components using subgraphs
5. **Annotations**: Add notes for complex interactions or decisions

---

*Last updated: January 2024*
