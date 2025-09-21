# 🚀 Classification Schema Migration Execution Guide

## 📋 Overview

This guide provides step-by-step instructions for executing **Subtask 1.2.1: Execute Classification Schema Migration** from the Supabase Table Improvement Implementation Plan.

## 🎯 Objective

Execute the `supabase-classification-migration.sql` script to create 6 critical classification tables that form the foundation of our enhanced classification system.

## 📊 Tables to be Created

The migration will create the following 6 tables:

1. **`industries`** - Core industry definitions and metadata
2. **`industry_keywords`** - Keywords associated with each industry
3. **`classification_codes`** - NAICS, SIC, and MCC codes for industries
4. **`industry_patterns`** - Pattern matching rules for industry detection
5. **`keyword_weights`** - Dynamic keyword weighting system
6. **`classification_accuracy_metrics`** - Performance tracking and analytics

## 🔧 Prerequisites

### Required Environment Variables

Before executing the migration, ensure the following environment variables are set:

```bash
# Supabase Configuration
SUPABASE_URL=https://your-project-id.supabase.co
SUPABASE_API_KEY=your_supabase_anon_key
SUPABASE_SERVICE_ROLE_KEY=your_supabase_service_role_key
SUPABASE_JWT_SECRET=your_supabase_jwt_secret
```

### Required Tools

- **Supabase Account**: Active Supabase project
- **psql** (optional): PostgreSQL command-line client for direct execution
- **curl** (optional): For API-based execution

## 🚀 Execution Methods

### Method 1: Supabase SQL Editor (Recommended)

This is the most reliable method for executing the migration:

#### Step 1: Access Supabase Dashboard
1. Navigate to [app.supabase.com](https://app.supabase.com)
2. Select your project
3. Go to **SQL Editor** in the left sidebar

#### Step 2: Execute Migration
1. Click **"New Query"**
2. Copy the entire contents of `supabase-classification-migration.sql`
3. Paste into the SQL Editor
4. Click **"Run"** to execute

#### Step 3: Verify Execution
1. Check the **"Results"** tab for any errors
2. Navigate to **Table Editor** to verify tables were created
3. Confirm all 6 tables are present:
   - `industries`
   - `industry_keywords`
   - `classification_codes`
   - `industry_patterns`
   - `keyword_weights`
   - `classification_accuracy_metrics`

### Method 2: Automated Script Execution

If you have the required environment variables configured:

```bash
# Make the script executable
chmod +x scripts/execute-subtask-1-2-1.sh

# Execute the migration
./scripts/execute-subtask-1-2-1.sh
```

### Method 3: Go Validator (Post-Migration)

After executing the migration, validate the results:

```bash
# Run the validation script
go run scripts/validate-classification-migration.go
```

## 📋 Migration Script Contents

The `supabase-classification-migration.sql` script includes:

### 1. Extensions
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
```

### 2. Table Creation
- **Industries Table**: Core industry definitions
- **Industry Keywords Table**: Keyword associations with weights
- **Classification Codes Table**: NAICS, SIC, MCC code mappings
- **Industry Patterns Table**: Pattern matching rules
- **Keyword Weights Table**: Dynamic weighting system
- **Classification Accuracy Metrics Table**: Performance tracking

### 3. Indexes
- Performance-optimized indexes for all tables
- Composite indexes for common query patterns
- Partial indexes for active records

### 4. Row Level Security (RLS)
- Public read access for classification data
- Service role access for metrics insertion
- Secure data access policies

### 5. Sample Data
- Technology and Retail industry examples
- Sample keywords and classification codes
- Test data for validation

## ✅ Validation Checklist

After executing the migration, verify the following:

### Table Existence
- [ ] `industries` table created
- [ ] `industry_keywords` table created
- [ ] `classification_codes` table created
- [ ] `industry_patterns` table created
- [ ] `keyword_weights` table created
- [ ] `classification_accuracy_metrics` table created

### Sample Data
- [ ] Technology industry inserted
- [ ] Retail industry inserted
- [ ] Sample keywords added
- [ ] Sample classification codes added

### Constraints and Relationships
- [ ] Foreign key constraints working
- [ ] Unique constraints enforced
- [ ] Check constraints validated
- [ ] RLS policies active

### Performance
- [ ] Indexes created successfully
- [ ] Query performance acceptable
- [ ] No blocking locks

## 🔍 Troubleshooting

### Common Issues

#### 1. Permission Errors
**Error**: `permission denied for table`
**Solution**: Ensure you're using the service role key, not the anon key

#### 2. Extension Errors
**Error**: `extension "uuid-ossp" does not exist`
**Solution**: Contact Supabase support to enable extensions

#### 3. Constraint Violations
**Error**: `duplicate key value violates unique constraint`
**Solution**: The script uses `ON CONFLICT DO NOTHING` to handle duplicates

#### 4. RLS Policy Errors
**Error**: `new row violates row-level security policy`
**Solution**: Ensure RLS policies are correctly configured

### Debug Steps

1. **Check Supabase Logs**
   - Go to **Logs** in Supabase dashboard
   - Look for error messages

2. **Verify Environment Variables**
   ```bash
   echo $SUPABASE_URL
   echo $SUPABASE_SERVICE_ROLE_KEY
   ```

3. **Test Connection**
   ```bash
   go run test-supabase-connection.go
   ```

4. **Manual Table Check**
   - Use Supabase Table Editor
   - Verify table structures match expected schema

## 📊 Expected Results

### Table Counts
After successful migration, you should see:
- **6 tables** created
- **2 sample industries** (Technology, Retail)
- **20+ sample keywords**
- **14+ sample classification codes**
- **Multiple indexes** for performance

### Performance Metrics
- **Migration time**: < 30 seconds
- **Table creation**: < 5 seconds per table
- **Index creation**: < 10 seconds total
- **Sample data insertion**: < 5 seconds

## 🎯 Success Criteria

The migration is considered successful when:

1. ✅ All 6 tables are created without errors
2. ✅ Sample data is inserted successfully
3. ✅ All constraints and relationships work
4. ✅ RLS policies are active and functional
5. ✅ Indexes are created for performance
6. ✅ Validation script passes all tests

## 📋 Next Steps

After successful migration execution:

1. **Proceed to Subtask 1.2.2**: Populate Classification Data
2. **Add comprehensive industry data**
3. **Populate NAICS, MCC, SIC codes**
4. **Create industry patterns for detection**
5. **Test classification system functionality**

## 🔗 Related Files

- **Migration Script**: `supabase-classification-migration.sql`
- **Execution Script**: `scripts/execute-subtask-1-2-1.sh`
- **Validation Script**: `scripts/validate-classification-migration.go`
- **Implementation Plan**: `SUPABASE_TABLE_IMPROVEMENT_IMPLEMENTATION_PLAN.md`

## 📞 Support

If you encounter issues:

1. Check the troubleshooting section above
2. Review Supabase documentation
3. Check project logs in Supabase dashboard
4. Verify environment configuration
5. Test with simplified queries first

---

**Document Version**: 1.0  
**Created**: January 19, 2025  
**Last Updated**: January 19, 2025  
**Next Review**: After migration execution
