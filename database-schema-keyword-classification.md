# 🗄️ **Database Schema: Keyword Classification System**

## 📋 **Overview**

This document defines the PostgreSQL database schema for the comprehensive keyword classification system. The schema will be implemented using either Railway's PostgreSQL service or Supabase, providing a dynamic, scalable foundation for industry keyword management.

## 🏗️ **Database Architecture**

### **Core Tables**
1. **industries** - Industry definitions and metadata
2. **industry_keywords** - Keywords associated with each industry
3. **classification_codes** - NAICS, MCC, and SIC codes
4. **code_keywords** - Keywords specific to each classification code
5. **industry_patterns** - Advanced pattern matching rules
6. **keyword_weights** - Dynamic keyword importance scoring
7. **audit_logs** - Change tracking and versioning

## 📊 **Detailed Schema Design**

### **1. industries Table**
```sql
CREATE TABLE industries (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    category VARCHAR(50) NOT NULL, -- 'traditional', 'emerging', 'hybrid'
    parent_industry_id INTEGER REFERENCES industries(id),
    confidence_threshold DECIMAL(3,2) DEFAULT 0.80,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_industries_name ON industries(name);
CREATE INDEX idx_industries_category ON industries(category);
CREATE INDEX idx_industries_active ON industries(is_active);
```

### **2. industry_keywords Table**
```sql
CREATE TABLE industry_keywords (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    keyword VARCHAR(100) NOT NULL,
    weight DECIMAL(3,2) DEFAULT 1.00, -- 0.00 to 1.00 importance
    context VARCHAR(50), -- 'business', 'technical', 'general'
    is_primary BOOLEAN DEFAULT false, -- High-priority keywords
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(industry_id, keyword)
);

-- Indexes
CREATE INDEX idx_industry_keywords_industry ON industry_keywords(industry_id);
CREATE INDEX idx_industry_keywords_keyword ON industry_keywords(keyword);
CREATE INDEX idx_industry_keywords_weight ON industry_keywords(weight);
CREATE INDEX idx_industry_keywords_primary ON industry_keywords(is_primary);
```

### **3. classification_codes Table**
```sql
CREATE TABLE classification_codes (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    code_type VARCHAR(10) NOT NULL CHECK (code_type IN ('NAICS', 'MCC', 'SIC')),
    code VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    confidence DECIMAL(3,2) DEFAULT 0.80,
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(code_type, code)
);

-- Indexes
CREATE INDEX idx_classification_codes_industry ON classification_codes(industry_id);
CREATE INDEX idx_classification_codes_type ON classification_codes(code_type);
CREATE INDEX idx_classification_codes_code ON classification_codes(code);
CREATE INDEX idx_classification_codes_primary ON classification_codes(is_primary);
```

### **4. code_keywords Table**
```sql
CREATE TABLE code_keywords (
    id SERIAL PRIMARY KEY,
    code_id INTEGER NOT NULL REFERENCES classification_codes(id) ON DELETE CASCADE,
    keyword VARCHAR(100) NOT NULL,
    relevance_score DECIMAL(3,2) DEFAULT 1.00,
    match_type VARCHAR(20) DEFAULT 'exact', -- 'exact', 'partial', 'synonym'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(code_id, keyword)
);

-- Indexes
CREATE INDEX idx_code_keywords_code ON code_keywords(code_id);
CREATE INDEX idx_code_keywords_keyword ON code_keywords(keyword);
CREATE INDEX idx_code_keywords_relevance ON code_keywords(relevance_score);
```

### **5. industry_patterns Table**
```sql
CREATE TABLE industry_patterns (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    pattern_type VARCHAR(50) NOT NULL, -- 'regex', 'phrase', 'semantic', 'context'
    pattern_data TEXT NOT NULL,
    priority INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_industry_patterns_industry ON industry_patterns(industry_id);
CREATE INDEX idx_industry_patterns_type ON industry_patterns(pattern_type);
CREATE INDEX idx_industry_patterns_active ON industry_patterns(is_active);
```

### **6. keyword_weights Table**
```sql
CREATE TABLE keyword_weights (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    keyword VARCHAR(100) NOT NULL,
    base_weight DECIMAL(3,2) DEFAULT 1.00,
    context_multiplier DECIMAL(3,2) DEFAULT 1.00,
    frequency_boost DECIMAL(3,2) DEFAULT 1.00,
    recency_factor DECIMAL(3,2) DEFAULT 1.00,
    calculated_weight DECIMAL(3,2) GENERATED ALWAYS AS (
        base_weight * context_multiplier * frequency_boost * recency_factor
    ) STORED,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(industry_id, keyword)
);

-- Indexes
CREATE INDEX idx_keyword_weights_industry ON keyword_weights(industry_id);
CREATE INDEX idx_keyword_weights_calculated ON keyword_weights(calculated_weight);
```

### **7. audit_logs Table**
```sql
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    table_name VARCHAR(50) NOT NULL,
    record_id INTEGER NOT NULL,
    action VARCHAR(20) NOT NULL, -- 'INSERT', 'UPDATE', 'DELETE'
    old_values JSONB,
    new_values JSONB,
    user_id VARCHAR(100),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_audit_logs_table ON audit_logs(table_name);
CREATE INDEX idx_audit_logs_record ON audit_logs(record_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);
```

## 🔄 **Sample Data Population**

### **Industries Table**
```sql
INSERT INTO industries (name, description, category) VALUES
('Grocery/Retail', 'Food and retail businesses including supermarkets, convenience stores, and specialty food shops', 'traditional'),
('Technology', 'Software, hardware, and digital services companies', 'traditional'),
('Financial Services', 'Banking, insurance, investment, and financial advisory services', 'traditional'),
('Healthcare', 'Medical services, pharmaceuticals, and health-related businesses', 'traditional'),
('Manufacturing', 'Industrial production and manufacturing companies', 'traditional'),
('E-commerce', 'Online retail and digital marketplace businesses', 'emerging'),
('Fintech', 'Financial technology and digital financial services', 'emerging'),
('Healthtech', 'Healthcare technology and digital health solutions', 'emerging');
```

### **Industry Keywords Table**
```sql
-- Grocery/Retail Keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, context, is_primary) VALUES
(1, 'grocery', 1.00, 'business', true),
(1, 'supermarket', 0.95, 'business', true),
(1, 'food', 0.90, 'business', true),
(1, 'fresh', 0.85, 'business', false),
(1, 'produce', 0.85, 'business', false),
(1, 'meat', 0.80, 'business', false),
(1, 'dairy', 0.80, 'business', false),
(1, 'bakery', 0.80, 'business', false),
(1, 'organic', 0.75, 'business', false),
(1, 'local', 0.70, 'business', false);

-- Technology Keywords
INSERT INTO industry_keywords (industry_id, keyword, weight, context, is_primary) VALUES
(2, 'software', 1.00, 'technical', true),
(2, 'platform', 0.95, 'technical', true),
(2, 'digital', 0.90, 'technical', true),
(2, 'api', 0.85, 'technical', false),
(2, 'cloud', 0.85, 'technical', false),
(2, 'ai', 0.80, 'technical', false),
(2, 'machine learning', 0.80, 'technical', false),
(2, 'cybersecurity', 0.75, 'technical', false),
(2, 'blockchain', 0.70, 'technical', false),
(2, 'iot', 0.70, 'technical', false);
```

### **Classification Codes Table**
```sql
-- Grocery/Retail Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, confidence, is_primary) VALUES
(1, 'NAICS', '445110', 'Supermarkets and Other Grocery (except Convenience) Stores', 0.95, true),
(1, 'NAICS', '445120', 'Convenience Stores', 0.90, false),
(1, 'NAICS', '445210', 'Meat Markets', 0.85, false),
(1, 'MCC', '5411', 'Grocery Stores, Supermarkets', 0.95, true),
(1, 'MCC', '5814', 'Fast Food Restaurants', 0.80, false),
(1, 'SIC', '5411', 'Grocery Stores', 0.95, true),
(1, 'SIC', '5421', 'Meat and Fish Markets', 0.85, false);

-- Technology Codes
INSERT INTO classification_codes (industry_id, code_type, code, description, confidence, is_primary) VALUES
(2, 'NAICS', '541511', 'Custom Computer Programming Services', 0.95, true),
(2, 'NAICS', '541512', 'Computer Systems Design Services', 0.90, false),
(2, 'NAICS', '541513', 'Computer Facilities Management Services', 0.85, false),
(2, 'MCC', '5734', 'Computer Software Stores', 0.90, false),
(2, 'MCC', '7372', 'Prepackaged Software', 0.95, true),
(2, 'SIC', '7372', 'Prepackaged Software', 0.95, true),
(2, 'SIC', '7373', 'Computer Integrated Systems Design', 0.90, false);
```

## 🚀 **Implementation Benefits**

### **Dynamic Management**
- **Hot-reloadable keywords** without code deployments
- **Admin interface** for keyword management
- **Bulk import/export** capabilities
- **Version control** and audit trails

### **Scalability**
- **Horizontal scaling** with connection pooling
- **Caching layer** for performance optimization
- **Database indexing** for fast queries
- **Partitioning** for large datasets

### **Reliability**
- **Fallback mechanisms** to hard-coded patterns
- **Health checks** and monitoring
- **Backup and recovery** procedures
- **Transaction safety** for data integrity

### **Flexibility**
- **Multi-tenant support** for different organizations
- **Custom industry definitions** per client
- **Dynamic weight adjustments** based on usage
- **A/B testing** capabilities for keyword effectiveness

## 🔧 **Database Connection Setup**

### **Railway PostgreSQL**
```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

func connectRailwayPostgres() (*sql.DB, error) {
    dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
        os.Getenv("RAILWAY_POSTGRES_HOST"),
        os.Getenv("RAILWAY_POSTGRES_PORT"),
        os.Getenv("RAILWAY_POSTGRES_USER"),
        os.Getenv("RAILWAY_POSTGRES_PASSWORD"),
        os.Getenv("RAILWAY_POSTGRES_DB"),
    )
    
    return sql.Open("postgres", dsn)
}
```

### **Supabase Integration**
```go
import (
    "github.com/supabase-community/supabase-go"
)

func connectSupabase() (*supabase.Client, error) {
    supabaseURL := os.Getenv("SUPABASE_URL")
    supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
    
    return supabase.NewClient(supabaseURL, supabaseKey, nil)
}
```

## 📊 **Performance Considerations**

### **Indexing Strategy**
- **Composite indexes** for common query patterns
- **Partial indexes** for active records only
- **Covering indexes** to avoid table lookups
- **Regular index maintenance** and optimization

### **Caching Strategy**
- **Redis caching** for frequently accessed patterns
- **In-memory caching** for active industry definitions
- **Cache invalidation** on keyword updates
- **Cache warming** for startup performance

### **Query Optimization**
- **Prepared statements** for repeated queries
- **Connection pooling** for concurrent access
- **Query monitoring** and performance analysis
- **Regular query optimization** and tuning

---

**Document Version**: 1.0.0  
**Created**: September 2, 2025  
**Status**: Ready for Implementation  
**Next Review**: After database schema implementation
