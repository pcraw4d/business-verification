# 🗄️ **Supabase Table Editor: Keyword Management Guide**

## 📋 **Overview**

This guide provides comprehensive instructions for using Supabase's built-in table editor to manage keywords, industries, and classification codes for the business verification system. The Supabase table editor offers a user-friendly interface for CRUD operations without requiring custom admin interfaces.

## 🏗️ **Database Schema Overview**

The keyword classification system uses the following tables:

### **Core Tables**
1. **`industries`** - Industry definitions and metadata
2. **`industry_keywords`** - Keywords associated with industries
3. **`classification_codes`** - NAICS, MCC, and SIC codes
4. **`code_keywords`** - Keywords associated with classification codes
5. **`keyword_patterns`** - Advanced keyword patterns and rules
6. **`classification_history`** - Historical classification data
7. **`system_metrics`** - Performance and usage metrics

## 🚀 **Accessing the Supabase Table Editor**

### **Step 1: Navigate to Supabase Dashboard**
1. Go to [https://supabase.com/dashboard](https://supabase.com/dashboard)
2. Sign in to your account
3. Select your project: `business-verification`

### **Step 2: Access Table Editor**
1. In the left sidebar, click **"Table Editor"**
2. You'll see all tables in your database
3. Click on any table name to open the editor

## 📊 **Table-Specific Management Guides**

### **1. Industries Table Management**

#### **Viewing Industries**
```sql
-- View all industries
SELECT * FROM industries ORDER BY name;

-- View industry with keyword count
SELECT 
    i.*,
    COUNT(ik.id) as keyword_count
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id
GROUP BY i.id
ORDER BY keyword_count DESC;
```

#### **Adding New Industries**
1. Click **"Insert"** → **"Insert row"**
2. Fill in the required fields:
   - `name`: Industry name (e.g., "E-commerce")
   - `description`: Industry description
   - `category`: Industry category
   - `is_active`: true/false
   - `created_at`: Current timestamp
   - `updated_at`: Current timestamp

#### **Editing Industries**
1. Click on any row to edit
2. Modify the fields as needed
3. Click **"Save"** to commit changes

#### **Deleting Industries**
1. Select the row(s) to delete
2. Click **"Delete"** button
3. Confirm the deletion

### **2. Industry Keywords Table Management**

#### **Viewing Keywords**
```sql
-- View keywords for a specific industry
SELECT 
    ik.*,
    i.name as industry_name
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
WHERE i.name = 'Technology'
ORDER BY ik.weight DESC;

-- View all keywords with their industries
SELECT 
    ik.keyword,
    ik.weight,
    ik.keyword_type,
    i.name as industry_name
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
ORDER BY i.name, ik.weight DESC;
```

#### **Adding New Keywords**
1. Click **"Insert"** → **"Insert row"**
2. Fill in the required fields:
   - `industry_id`: Select from dropdown (or use ID)
   - `keyword`: The keyword text (e.g., "software")
   - `weight`: Importance weight (0.0 to 1.0)
   - `keyword_type`: Type (e.g., "primary", "secondary", "contextual")
   - `is_active`: true/false
   - `created_at`: Current timestamp

#### **Bulk Keyword Operations**
```sql
-- Add multiple keywords for an industry
INSERT INTO industry_keywords (industry_id, keyword, weight, keyword_type, is_active, created_at)
VALUES 
    (1, 'artificial intelligence', 0.9, 'primary', true, NOW()),
    (1, 'machine learning', 0.8, 'primary', true, NOW()),
    (1, 'deep learning', 0.7, 'secondary', true, NOW()),
    (1, 'neural networks', 0.6, 'secondary', true, NOW());
```

### **3. Classification Codes Table Management**

#### **Viewing Classification Codes**
```sql
-- View codes by type
SELECT * FROM classification_codes 
WHERE code_type = 'NAICS' 
ORDER BY code;

-- View codes with descriptions
SELECT 
    code,
    description,
    code_type,
    industry_id,
    i.name as industry_name
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
WHERE code_type = 'MCC'
ORDER BY code;
```

#### **Adding New Classification Codes**
1. Click **"Insert"** → **"Insert row"**
2. Fill in the required fields:
   - `code`: The classification code (e.g., "541511")
   - `description`: Code description
   - `code_type`: "NAICS", "MCC", or "SIC"
   - `industry_id`: Associated industry ID
   - `is_active`: true/false
   - `created_at`: Current timestamp

### **4. Code Keywords Table Management**

#### **Viewing Code Keywords**
```sql
-- View keywords for specific codes
SELECT 
    ck.*,
    cc.code,
    cc.description,
    cc.code_type
FROM code_keywords ck
JOIN classification_codes cc ON ck.code_id = cc.id
WHERE cc.code_type = 'NAICS'
ORDER BY cc.code, ck.weight DESC;
```

#### **Adding Code Keywords**
1. Click **"Insert"** → **"Insert row"**
2. Fill in the required fields:
   - `code_id`: Classification code ID
   - `keyword`: Associated keyword
   - `weight`: Importance weight (0.0 to 1.0)
   - `is_active`: true/false
   - `created_at`: Current timestamp

## 🔍 **Advanced Query Operations**

### **Keyword Analysis Queries**
```sql
-- Find industries with most keywords
SELECT 
    i.name,
    COUNT(ik.id) as keyword_count,
    AVG(ik.weight) as avg_weight
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id
GROUP BY i.id, i.name
ORDER BY keyword_count DESC;

-- Find keywords used across multiple industries
SELECT 
    ik.keyword,
    COUNT(DISTINCT ik.industry_id) as industry_count,
    STRING_AGG(i.name, ', ') as industries
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
GROUP BY ik.keyword
HAVING COUNT(DISTINCT ik.industry_id) > 1
ORDER BY industry_count DESC;
```

### **Classification Code Analysis**
```sql
-- Find codes with most keywords
SELECT 
    cc.code,
    cc.description,
    cc.code_type,
    COUNT(ck.id) as keyword_count
FROM classification_codes cc
LEFT JOIN code_keywords ck ON cc.id = ck.code_id
GROUP BY cc.id, cc.code, cc.description, cc.code_type
ORDER BY keyword_count DESC;

-- Find missing keyword mappings
SELECT 
    cc.code,
    cc.description,
    cc.code_type
FROM classification_codes cc
LEFT JOIN code_keywords ck ON cc.id = ck.code_id
WHERE ck.id IS NULL
ORDER BY cc.code_type, cc.code;
```

## 🛠️ **Data Validation and Quality Checks**

### **Data Integrity Queries**
```sql
-- Check for orphaned keywords
SELECT * FROM industry_keywords 
WHERE industry_id NOT IN (SELECT id FROM industries);

-- Check for duplicate keywords within same industry
SELECT 
    industry_id,
    keyword,
    COUNT(*) as duplicate_count
FROM industry_keywords
GROUP BY industry_id, keyword
HAVING COUNT(*) > 1;

-- Check for invalid weights
SELECT * FROM industry_keywords 
WHERE weight < 0 OR weight > 1;

-- Check for empty keywords
SELECT * FROM industry_keywords 
WHERE keyword IS NULL OR TRIM(keyword) = '';
```

### **Performance Monitoring Queries**
```sql
-- Check classification history trends
SELECT 
    DATE(created_at) as date,
    COUNT(*) as classification_count,
    AVG(confidence_score) as avg_confidence
FROM classification_history
WHERE created_at >= NOW() - INTERVAL '7 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;

-- Check system metrics
SELECT 
    metric_name,
    metric_value,
    recorded_at
FROM system_metrics
WHERE recorded_at >= NOW() - INTERVAL '1 hour'
ORDER BY recorded_at DESC;
```

## 📈 **Bulk Operations and Data Import**

### **CSV Import Process**
1. **Prepare CSV File**:
   ```csv
   industry_id,keyword,weight,keyword_type,is_active
   1,software,0.9,primary,true
   1,technology,0.8,primary,true
   1,development,0.7,secondary,true
   ```

2. **Import via Supabase Dashboard**:
   - Go to Table Editor
   - Click **"Insert"** → **"Import data from CSV"**
   - Upload your CSV file
   - Map columns to table fields
   - Click **"Import"**

### **Bulk Update Operations**
```sql
-- Update keyword weights based on performance
UPDATE industry_keywords 
SET weight = weight * 1.1
WHERE keyword IN (
    SELECT DISTINCT keyword 
    FROM classification_history 
    WHERE confidence_score > 0.8
    AND created_at >= NOW() - INTERVAL '30 days'
);

-- Deactivate low-performing keywords
UPDATE industry_keywords 
SET is_active = false
WHERE weight < 0.3
AND keyword NOT IN (
    SELECT DISTINCT keyword 
    FROM classification_history 
    WHERE confidence_score > 0.7
    AND created_at >= NOW() - INTERVAL '7 days'
);
```

## 🔒 **Security and Access Control**

### **Row Level Security (RLS)**
The system uses Supabase's Row Level Security for data protection:

```sql
-- Enable RLS on all tables
ALTER TABLE industries ENABLE ROW LEVEL SECURITY;
ALTER TABLE industry_keywords ENABLE ROW LEVEL SECURITY;
ALTER TABLE classification_codes ENABLE ROW LEVEL SECURITY;
ALTER TABLE code_keywords ENABLE ROW LEVEL SECURITY;

-- Create policies for authenticated users
CREATE POLICY "Allow authenticated users to read all data" 
ON industries FOR SELECT 
TO authenticated 
USING (true);

CREATE POLICY "Allow authenticated users to modify data" 
ON industry_keywords FOR ALL 
TO authenticated 
USING (true);
```

### **API Key Management**
- **Service Role Key**: Full access (use for server-side operations)
- **Anon Key**: Limited access (use for client-side operations)
- **API Keys**: Stored securely in environment variables

## 📊 **Monitoring and Analytics**

### **Real-time Monitoring**
```sql
-- Monitor active classifications
SELECT 
    COUNT(*) as active_classifications,
    AVG(confidence_score) as avg_confidence
FROM classification_history
WHERE created_at >= NOW() - INTERVAL '1 hour';

-- Monitor keyword usage
SELECT 
    ik.keyword,
    COUNT(ch.id) as usage_count,
    AVG(ch.confidence_score) as avg_confidence
FROM industry_keywords ik
JOIN classification_history ch ON ch.keywords_used @> ARRAY[ik.keyword]
WHERE ch.created_at >= NOW() - INTERVAL '24 hours'
GROUP BY ik.keyword
ORDER BY usage_count DESC
LIMIT 10;
```

### **Performance Metrics**
```sql
-- Query performance metrics
SELECT 
    metric_name,
    AVG(metric_value) as avg_value,
    MAX(metric_value) as max_value,
    MIN(metric_value) as min_value
FROM system_metrics
WHERE recorded_at >= NOW() - INTERVAL '1 day'
GROUP BY metric_name
ORDER BY avg_value DESC;
```

## 🚨 **Troubleshooting Common Issues**

### **Connection Issues**
1. **Check API Keys**: Verify SUPABASE_URL and SUPABASE_API_KEY
2. **Check Network**: Ensure firewall allows Supabase connections
3. **Check Quotas**: Monitor usage against free tier limits

### **Data Issues**
1. **Foreign Key Violations**: Ensure referenced IDs exist
2. **Duplicate Keys**: Check for unique constraint violations
3. **Data Type Mismatches**: Verify field types match expected values

### **Performance Issues**
1. **Slow Queries**: Use EXPLAIN ANALYZE to identify bottlenecks
2. **Missing Indexes**: Add indexes for frequently queried columns
3. **Large Result Sets**: Use LIMIT and pagination

## 📚 **Best Practices**

### **Data Management**
1. **Regular Backups**: Use Supabase's automatic backups
2. **Data Validation**: Implement constraints and triggers
3. **Version Control**: Track schema changes
4. **Testing**: Use staging environment for changes

### **Performance Optimization**
1. **Indexing**: Add indexes for search columns
2. **Query Optimization**: Use efficient SQL patterns
3. **Caching**: Leverage Supabase's built-in caching
4. **Connection Pooling**: Use connection pooling for high traffic

### **Security**
1. **RLS Policies**: Implement proper row-level security
2. **API Key Rotation**: Regularly rotate API keys
3. **Access Control**: Limit access to sensitive operations
4. **Audit Logging**: Monitor data access and changes

## 🔗 **Useful Links**

- [Supabase Table Editor Documentation](https://supabase.com/docs/guides/database/tables)
- [Supabase SQL Editor](https://supabase.com/docs/guides/database/sql-editor)
- [Supabase Row Level Security](https://supabase.com/docs/guides/auth/row-level-security)
- [Supabase API Documentation](https://supabase.com/docs/reference/api)

## 📞 **Support**

For issues with the Supabase table editor:
1. Check the [Supabase Documentation](https://supabase.com/docs)
2. Visit the [Supabase Community](https://github.com/supabase/supabase/discussions)
3. Contact the development team for system-specific issues

---

**Last Updated**: January 2025  
**Version**: 1.0.0  
**Maintained By**: Business Verification Development Team
