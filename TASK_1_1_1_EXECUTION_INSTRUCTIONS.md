# 🔧 Task 1.1.1 Execution Instructions

## 📋 **Task Overview**
**Task**: Add missing `is_active` column to `keyword_weights` table  
**Duration**: 2 hours  
**Priority**: CRITICAL  
**Dependencies**: None  

## 🎯 **Objective**
Fix the critical database schema issue that prevents the classification system from building the keyword index. The `keyword_weights` table is missing the `is_active` column, causing "is_active does not exist" errors.

## 🔍 **Current State Verification**
Before executing this task, we've confirmed:
- ✅ Supabase connection is working
- ✅ `keyword_weights` table exists and is accessible
- ❌ `is_active` column is missing (confirmed by test: `column keyword_weights.is_active does not exist`)

## 🚀 **Execution Steps**

### **Step 1: Access Supabase SQL Editor**
1. **Open Supabase Dashboard**: Go to [https://supabase.com/dashboard](https://supabase.com/dashboard)
2. **Select Project**: Choose your KYB Platform project
3. **Navigate to SQL Editor**: Click on "SQL Editor" in the left sidebar
4. **Create New Query**: Click "New Query" button

### **Step 2: Execute the SQL Script**
1. **Copy the SQL Script**: Open `scripts/task-1-1-1-sql-script.sql`
2. **Paste into SQL Editor**: Copy the entire contents and paste into the SQL Editor
3. **Execute the Script**: Click the "Run" button (or press Ctrl+Enter)

### **Step 3: Verify Execution Results**
After running the script, you should see the following results:

#### **Verification 1: Column Check**
```
test_name    | column_name | data_type | is_nullable | column_default
-------------|-------------|-----------|-------------|---------------
Column Check | is_active   | boolean   | YES         | true
```

#### **Verification 2: Record Check**
```
test_name    | total_records | active_records | inactive_records | null_records
-------------|---------------|----------------|------------------|-------------
Record Check | [number]      | [same number]  | 0                | 0
```

#### **Verification 3: Index Check**
```
test_name    | indexname                              | indexdef
-------------|----------------------------------------|----------------------------------------
Index Check  | idx_keyword_weights_active             | CREATE INDEX ...
Index Check  | idx_keyword_weights_industry_active    | CREATE INDEX ...
```

#### **Verification 4: Performance Test**
```
test_name        | filtered_count
-----------------|---------------
Performance Test | [same as total_records]
```

## ✅ **Success Criteria**
- [ ] `is_active` column exists in `keyword_weights` table
- [ ] All existing records have `is_active = true`
- [ ] Performance indexes are created
- [ ] No database errors when querying the table
- [ ] Classification system can build keyword index without errors

## 🧪 **Post-Execution Testing**

### **Test 1: Verify Column Exists**
Run the following Go test to confirm the column was added:
```bash
SUPABASE_URL=$(grep SUPABASE_URL .env | cut -d'=' -f2) \
SUPABASE_API_KEY=$(grep SUPABASE_API_KEY .env | cut -d'=' -f2) \
SUPABASE_SERVICE_ROLE_KEY=$(grep SUPABASE_SERVICE_ROLE_KEY .env | cut -d'=' -f2) \
SUPABASE_JWT_SECRET=$(grep SUPABASE_JWT_SECRET .env | cut -d'=' -f2) \
go run test-keyword-weights-schema.go
```

**Expected Output**: All tests should pass, including "✅ is_active column exists and is accessible"

### **Test 2: Test Classification System**
Run the classification system to ensure it no longer produces "is_active does not exist" errors:
```bash
# Start the server
go run cmd/railway-server/main.go

# In another terminal, test classification
curl -X POST http://localhost:8080/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Restaurant",
    "description": "Fine dining restaurant serving Italian cuisine",
    "website_url": ""
  }'
```

**Expected Result**: No "is_active does not exist" errors in server logs

## 📊 **Expected Impact**
After completing Task 1.1.1:
- ✅ Database schema errors eliminated
- ✅ Keyword index building will work correctly
- ✅ Classification system can process requests without database errors
- ✅ Foundation ready for Task 1.1.2 and Task 1.1.3

## 🔄 **Rollback Plan**
If issues occur, you can rollback by:
1. **Remove the column**: `ALTER TABLE keyword_weights DROP COLUMN IF EXISTS is_active;`
2. **Remove the indexes**: 
   ```sql
   DROP INDEX IF EXISTS idx_keyword_weights_active;
   DROP INDEX IF EXISTS idx_keyword_weights_industry_active;
   ```

## 📝 **Documentation Update**
After successful completion:
1. Update the TODO list to mark Task 1.1.1 as completed
2. Document any issues encountered and their resolution
3. Prepare for Task 1.1.2 (Update existing records) - which is already included in this task
4. Prepare for Task 1.1.3 (Create performance indexes) - which is already included in this task

## 🎯 **Next Steps**
After Task 1.1.1 is completed:
- **Task 1.1.2**: Update existing records (✅ Already completed in this task)
- **Task 1.1.3**: Create performance indexes (✅ Already completed in this task)
- **Task 1.2**: Add Restaurant Industry Data
- **Task 1.3**: Test Restaurant Classification

---

**Note**: This task is critical and must be completed before proceeding with any other classification improvements. The missing `is_active` column is blocking the entire classification system from functioning properly.
