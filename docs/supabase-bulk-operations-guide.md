# 📦 **Supabase Bulk Import/Export Operations Guide**

## 📋 **Overview**

This guide provides comprehensive instructions for bulk import/export operations using Supabase's REST API and custom tools. The system includes both Bash and Python scripts for different use cases and complexity levels.

## 🛠️ **Available Tools**

### **1. Bash Script (`supabase-bulk-import-export.sh`)**
- **Purpose**: Simple, fast operations for basic import/export
- **Use Cases**: Quick backups, CSV imports, basic data migration
- **Requirements**: `curl`, `jq`, Bash 4.0+

### **2. Python Script (`supabase_bulk_operations.py`)**
- **Purpose**: Advanced operations with validation and complex data handling
- **Use Cases**: Data validation, complex migrations, environment syncing
- **Requirements**: Python 3.8+, pip packages (see `requirements.txt`)

## 🚀 **Quick Start**

### **Environment Setup**
```bash
# Set required environment variables
export SUPABASE_URL="https://your-project.supabase.co"
export SUPABASE_SERVICE_ROLE_KEY="your-service-role-key"

# For Python script, also install dependencies
pip install -r scripts/requirements.txt
```

### **Basic Operations**
```bash
# Export all data (Bash)
./scripts/supabase-bulk-import-export.sh export-industries
./scripts/supabase-bulk-import-export.sh export-keywords
./scripts/supabase-bulk-import-export.sh export-codes

# Export all data (Python)
python scripts/supabase_bulk_operations.py export-all --output-dir exports/

# Import data (Bash)
./scripts/supabase-bulk-import-export.sh import-keywords /path/to/keywords.csv

# Import data (Python)
python scripts/supabase_bulk_operations.py import-all --input-dir imports/
```

## 📊 **Bash Script Operations**

### **Export Commands**
```bash
# Export industries to CSV
./scripts/supabase-bulk-import-export.sh export-industries

# Export keywords to CSV
./scripts/supabase-bulk-import-export.sh export-keywords

# Export classification codes to CSV
./scripts/supabase-bulk-import-export.sh export-codes

# Create full backup
./scripts/supabase-bulk-import-export.sh backup-all
```

### **Import Commands**
```bash
# Import industries from CSV
./scripts/supabase-bulk-import-export.sh import-industries /path/to/industries.csv

# Import keywords from CSV
./scripts/supabase-bulk-import-export.sh import-keywords /path/to/keywords.csv

# Import classification codes from CSV
./scripts/supabase-bulk-import-export.sh import-codes /path/to/codes.csv

# Restore from backup
./scripts/supabase-bulk-import-export.sh restore-all /path/to/backup_20250119_143022
```

### **CSV Format Requirements**

#### **Industries CSV**
```csv
id,name,description,category,is_active,created_at,updated_at
1,Technology,Software development and technology services,Technology,true,2025-01-19T10:00:00Z,2025-01-19T10:00:00Z
2,Healthcare,Medical and healthcare services,Healthcare,true,2025-01-19T10:00:00Z,2025-01-19T10:00:00Z
```

#### **Keywords CSV**
```csv
id,industry_id,industry_name,keyword,weight,keyword_type,is_active,created_at,updated_at
1,1,Technology,software,0.9,primary,true,2025-01-19T10:00:00Z,2025-01-19T10:00:00Z
2,1,Technology,technology,0.8,primary,true,2025-01-19T10:00:00Z,2025-01-19T10:00:00Z
```

#### **Classification Codes CSV**
```csv
id,code,description,code_type,industry_id,industry_name,is_active,created_at,updated_at
1,541511,Custom Computer Programming Services,NAICS,1,Technology,true,2025-01-19T10:00:00Z,2025-01-19T10:00:00Z
2,541512,Computer Systems Design Services,NAICS,1,Technology,true,2025-01-19T10:00:00Z,2025-01-19T10:00:00Z
```

## 🐍 **Python Script Operations**

### **Export Operations**
```bash
# Export all data to structured format
python scripts/supabase_bulk_operations.py export-all --output-dir exports/

# Export with verbose logging
python scripts/supabase_bulk_operations.py export-all --output-dir exports/ --verbose
```

### **Import Operations**
```bash
# Import all data with validation
python scripts/supabase_bulk_operations.py import-all --input-dir imports/

# Import without validation (faster)
python scripts/supabase_bulk_operations.py import-all --input-dir imports/ --no-validate
```

### **Data Validation**
```bash
# Validate data integrity
python scripts/supabase_bulk_operations.py validate-data

# Validate with verbose output
python scripts/supabase_bulk_operations.py validate-data --verbose
```

### **Environment Syncing**
```bash
# Sync data between environments
python scripts/supabase_bulk_operations.py sync-data \
    --source-url "https://staging-project.supabase.co" \
    --source-key "staging-service-role-key" \
    --tables industries industry_keywords classification_codes

# Sync specific tables only
python scripts/supabase_bulk_operations.py sync-data \
    --source-url "https://staging-project.supabase.co" \
    --source-key "staging-service-role-key" \
    --tables industries
```

### **Sample Data Generation**
```bash
# Generate sample data for testing
python scripts/supabase_bulk_operations.py generate-sample --output-dir samples/
```

## 📁 **File Structure**

### **Export Structure**
```
exports/
├── export_manifest_20250119_143022.json
├── industries_20250119_143022.json
├── keywords_20250119_143022.json
├── codes_20250119_143022.json
├── code_keywords_20250119_143022.json
└── history_20250119_143022.json
```

### **Backup Structure**
```
backups/
└── backup_20250119_143022/
    ├── manifest.json
    ├── industries_export_20250119_143022.csv
    ├── keywords_export_20250119_143022.csv
    └── codes_export_20250119_143022.csv
```

## 🔧 **Advanced Operations**

### **Custom Data Processing**

#### **Filter and Transform Data**
```python
# Example: Filter keywords by weight
import json
from pathlib import Path

# Load exported data
with open('exports/keywords_20250119_143022.json', 'r') as f:
    keywords = json.load(f)

# Filter high-weight keywords
high_weight_keywords = [
    k for k in keywords 
    if k['weight'] > 0.7 and k['is_active']
]

# Save filtered data
with open('filtered_keywords.json', 'w') as f:
    json.dump(high_weight_keywords, f, indent=2)
```

#### **Bulk Data Updates**
```python
# Example: Update keyword weights based on performance
from supabase import create_client

client = create_client(SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY)

# Get keywords with low performance
low_perf_keywords = client.table('industry_keywords').select('*').lt('weight', 0.3).execute()

# Update weights
for keyword in low_perf_keywords.data:
    new_weight = min(keyword['weight'] * 1.2, 1.0)
    client.table('industry_keywords').update({
        'weight': new_weight,
        'updated_at': 'now()'
    }).eq('id', keyword['id']).execute()
```

### **Data Migration Scripts**

#### **Migrate from Legacy System**
```python
#!/usr/bin/env python3
"""
Legacy system migration script
"""

import json
import requests
from supabase import create_client

def migrate_from_legacy():
    # Connect to legacy system
    legacy_data = requests.get('https://legacy-api.com/keywords').json()
    
    # Connect to Supabase
    client = create_client(SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY)
    
    # Transform and migrate data
    for item in legacy_data:
        # Transform legacy format to new format
        new_item = {
            'keyword': item['term'],
            'weight': item['importance'] / 100.0,
            'keyword_type': 'legacy',
            'is_active': True
        }
        
        # Insert into Supabase
        client.table('industry_keywords').insert(new_item).execute()

if __name__ == '__main__':
    migrate_from_legacy()
```

## 🚨 **Error Handling and Troubleshooting**

### **Common Issues**

#### **Authentication Errors**
```bash
# Check environment variables
echo $SUPABASE_URL
echo $SUPABASE_SERVICE_ROLE_KEY

# Test connection
curl -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
     -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
     "$SUPABASE_URL/rest/v1/industries?select=count"
```

#### **Data Validation Errors**
```bash
# Validate CSV format
head -n 1 your_file.csv
# Should match expected headers

# Check for special characters
grep -P '[^\x00-\x7F]' your_file.csv
```

#### **Import Failures**
```bash
# Check for duplicate keys
python scripts/supabase_bulk_operations.py validate-data

# Check for orphaned references
# (See validation output for specific issues)
```

### **Recovery Procedures**

#### **Restore from Backup**
```bash
# List available backups
ls -la backups/

# Restore from specific backup
./scripts/supabase-bulk-import-export.sh restore-all backups/backup_20250119_143022/
```

#### **Partial Data Recovery**
```python
# Restore specific tables only
python scripts/supabase_bulk_operations.py import-all \
    --input-dir backups/backup_20250119_143022/ \
    --no-validate
```

## 📈 **Performance Optimization**

### **Large Dataset Handling**

#### **Batch Processing**
```python
def import_large_dataset(data, batch_size=1000):
    """Import large datasets in batches."""
    for i in range(0, len(data), batch_size):
        batch = data[i:i + batch_size]
        result = ops._make_request('POST', 'industry_keywords', batch)
        logger.info(f"Imported batch {i//batch_size + 1}: {len(result)} records")
```

#### **Parallel Processing**
```python
import concurrent.futures
from multiprocessing import Pool

def parallel_import(data_chunks):
    """Import data chunks in parallel."""
    with Pool(processes=4) as pool:
        results = pool.map(import_chunk, data_chunks)
    return results
```

### **Memory Optimization**
```python
# For very large datasets, use streaming
def stream_import(csv_file):
    """Stream import for large CSV files."""
    import pandas as pd
    
    chunk_size = 1000
    for chunk in pd.read_csv(csv_file, chunksize=chunk_size):
        # Process chunk
        result = ops._make_request('POST', 'industry_keywords', chunk.to_dict('records'))
        yield len(result)
```

## 🔒 **Security Considerations**

### **API Key Management**
```bash
# Use environment variables (recommended)
export SUPABASE_SERVICE_ROLE_KEY="your-key"

# Or use .env file
echo "SUPABASE_SERVICE_ROLE_KEY=your-key" > .env
```

### **Data Privacy**
```python
# Sanitize sensitive data before export
def sanitize_export_data(data):
    """Remove sensitive information from export data."""
    sanitized = []
    for item in data:
        # Remove or mask sensitive fields
        if 'sensitive_field' in item:
            item['sensitive_field'] = '***REDACTED***'
        sanitized.append(item)
    return sanitized
```

## 📊 **Monitoring and Logging**

### **Operation Logging**
```python
import logging

# Configure detailed logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('bulk_operations.log'),
        logging.StreamHandler()
    ]
)
```

### **Performance Monitoring**
```python
import time

def monitor_operation(func):
    """Decorator to monitor operation performance."""
    def wrapper(*args, **kwargs):
        start_time = time.time()
        result = func(*args, **kwargs)
        end_time = time.time()
        
        logger.info(f"Operation {func.__name__} completed in {end_time - start_time:.2f} seconds")
        return result
    return wrapper
```

## 📚 **Best Practices**

### **Data Management**
1. **Always backup before major operations**
2. **Validate data before import**
3. **Use staging environment for testing**
4. **Monitor operation logs**
5. **Test with small datasets first**

### **Performance**
1. **Use batch processing for large datasets**
2. **Implement proper error handling**
3. **Use appropriate batch sizes (100-1000 records)**
4. **Monitor memory usage**
5. **Consider parallel processing for large operations**

### **Security**
1. **Never commit API keys to version control**
2. **Use environment variables for configuration**
3. **Implement proper access controls**
4. **Audit data access and changes**
5. **Sanitize sensitive data in exports**

## 🔗 **Useful Resources**

- [Supabase REST API Documentation](https://supabase.com/docs/reference/api)
- [PostgREST API Documentation](https://postgrest.org/en/stable/api.html)
- [Python Supabase Client](https://github.com/supabase/supabase-py)
- [Bash Scripting Best Practices](https://google.github.io/styleguide/shellguide.html)

## 📞 **Support**

For issues with bulk operations:
1. Check the logs for specific error messages
2. Validate your data format and structure
3. Test with small datasets first
4. Contact the development team for system-specific issues

---

**Last Updated**: January 2025  
**Version**: 1.0.0  
**Maintained By**: Business Verification Development Team
