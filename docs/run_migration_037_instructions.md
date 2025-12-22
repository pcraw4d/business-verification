# Instructions to Run Migration 037

## Option 1: Using Supabase Dashboard (Recommended)

1. **Open Supabase Dashboard**
   - Go to your Supabase project dashboard
   - Navigate to **SQL Editor**

2. **Run the Migration**
   - Copy the contents of `supabase-migrations/037_verify_schema_and_fix_mismatches.sql`
   - Paste into the SQL Editor
   - Click **Run** or press `Ctrl+Enter` (Windows/Linux) or `Cmd+Enter` (Mac)

3. **Review Results**
   - The script will output NOTICE messages showing:
     - ✅ Tables that exist
     - ✅ Functions that exist
     - ✅ Column types verified
     - ✅ Indexes verified
     - ⚠️ Any warnings or missing components

## Option 2: Using Supabase CLI (If Project is Linked)

```bash
# First, link your project (if not already linked)
supabase link --project-ref YOUR_PROJECT_REF

# Then push the migration
supabase db push
```

## Option 3: Using psql Directly

If you have the database connection string:

```bash
# Set connection string (replace with your actual values)
export DATABASE_URL="postgresql://postgres:[YOUR-PASSWORD]@[YOUR-PROJECT-REF].supabase.co:5432/postgres"

# Run the migration
psql "$DATABASE_URL" -f supabase-migrations/037_verify_schema_and_fix_mismatches.sql
```

## What the Migration Does

1. **Fixes Type Mismatch**: Updates `get_codes_by_trigram_similarity` to cast `cc.code` to `text`
2. **Verifies Tables**: Checks all 10 required tables exist
3. **Verifies Functions**: Checks all 3 RPC functions exist with correct signatures
4. **Verifies Column Types**: Validates critical column types match expectations
5. **Verifies Indexes**: Checks critical indexes exist
6. **Verifies Foreign Keys**: Validates foreign key relationships
7. **Tests Functions**: Runs sample queries to ensure functions work

## Expected Output

You should see output like:

```
========================================
SCHEMA VERIFICATION: Required Tables
========================================
✅ Table exists: classification_codes
✅ Table exists: code_keywords
...
✅ All required tables exist

========================================
SCHEMA VERIFICATION: RPC Functions
========================================
✅ get_codes_by_keywords exists
   Arguments: p_code_type text, p_keywords text[], p_limit integer DEFAULT 10
   Returns: TABLE(code text, description text, max_weight double precision)
...
```

## Troubleshooting

If you see errors:
- **"function already exists"**: This is OK - the function is being replaced
- **"table does not exist"**: Check that previous migrations have been run
- **"permission denied"**: Ensure you're using the service role key or have proper permissions

