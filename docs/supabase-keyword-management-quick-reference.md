# 🚀 **Supabase Keyword Management - Quick Reference**

## 📋 **Quick Access Links**

- **Supabase Dashboard**: [https://supabase.com/dashboard](https://supabase.com/dashboard)
- **Table Editor**: Dashboard → Table Editor
- **SQL Editor**: Dashboard → SQL Editor
- **API Docs**: Dashboard → API Documentation

## 🔍 **Common Queries**

### **View All Industries**
```sql
SELECT * FROM industries ORDER BY name;
```

### **View Keywords for Industry**
```sql
SELECT ik.*, i.name as industry_name 
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
WHERE i.name = 'Technology'
ORDER BY ik.weight DESC;
```

### **View Classification Codes**
```sql
SELECT cc.*, i.name as industry_name
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
WHERE cc.code_type = 'NAICS'
ORDER BY cc.code;
```

## ➕ **Adding Data**

### **Add New Industry**
1. Table Editor → `industries` → Insert → Insert row
2. Fill: `name`, `description`, `category`, `is_active: true`

### **Add Keywords**
```sql
INSERT INTO industry_keywords (industry_id, keyword, weight, keyword_type, is_active, created_at)
VALUES (1, 'software', 0.9, 'primary', true, NOW());
```

### **Add Classification Code**
```sql
INSERT INTO classification_codes (code, description, code_type, industry_id, is_active, created_at)
VALUES ('541511', 'Custom Computer Programming Services', 'NAICS', 1, true, NOW());
```

## 🔧 **Bulk Operations**

### **Import CSV Data**
1. Table Editor → Insert → Import data from CSV
2. Upload CSV file
3. Map columns to fields
4. Click Import

### **Bulk Update Keywords**
```sql
UPDATE industry_keywords 
SET weight = weight * 1.1
WHERE keyword_type = 'primary';
```

### **Bulk Deactivate Keywords**
```sql
UPDATE industry_keywords 
SET is_active = false
WHERE weight < 0.3;
```

## 🧹 **Data Cleanup**

### **Remove Duplicates**
```sql
DELETE FROM industry_keywords 
WHERE id NOT IN (
    SELECT MIN(id) FROM industry_keywords 
    GROUP BY industry_id, keyword
);
```

### **Clean Old Data**
```sql
DELETE FROM classification_history 
WHERE created_at < NOW() - INTERVAL '6 months';
```

## 📊 **Monitoring**

### **System Health Check**
```sql
SELECT 
    'Industries' as type, COUNT(*) as total, COUNT(CASE WHEN is_active THEN 1 END) as active
FROM industries
UNION ALL
SELECT 
    'Keywords' as type, COUNT(*) as total, COUNT(CASE WHEN is_active THEN 1 END) as active
FROM industry_keywords;
```

### **Performance Metrics**
```sql
SELECT 
    DATE(created_at) as date,
    COUNT(*) as classifications,
    AVG(confidence_score) as avg_confidence
FROM classification_history
WHERE created_at >= NOW() - INTERVAL '7 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

## 🚨 **Troubleshooting**

### **Check for Issues**
```sql
-- Orphaned keywords
SELECT * FROM industry_keywords 
WHERE industry_id NOT IN (SELECT id FROM industries);

-- Invalid weights
SELECT * FROM industry_keywords 
WHERE weight < 0 OR weight > 1;

-- Empty keywords
SELECT * FROM industry_keywords 
WHERE keyword IS NULL OR TRIM(keyword) = '';
```

### **Fix Common Issues**
```sql
-- Fix invalid weights
UPDATE industry_keywords 
SET weight = 0.5 
WHERE weight < 0 OR weight > 1;

-- Fix empty keywords
DELETE FROM industry_keywords 
WHERE keyword IS NULL OR TRIM(keyword) = '';
```

## 🔑 **Environment Variables**

Make sure these are set in your environment:
```bash
SUPABASE_URL=your_supabase_url
SUPABASE_API_KEY=your_api_key
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key
```

## 📞 **Need Help?**

1. **Documentation**: Check the full guide in `docs/supabase-table-editor-keyword-management.md`
2. **SQL Queries**: Use queries from `configs/supabase/keyword_management_queries.sql`
3. **Support**: Contact the development team

---

**Quick Reference Version**: 1.0.0  
**Last Updated**: January 2025
