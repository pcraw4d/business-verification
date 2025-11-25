# NAICS Alignment Migrations - Ready to Run

## ✅ Migration 023 - COMPLETED

You have successfully run `023_add_naics_aligned_industries.sql`. This migration:

- Added 11 new NAICS-aligned industries
- Updated existing industry descriptions
- Added comprehensive keywords for each new industry

---

## 📋 Migration 024 - Reclassify Codes to NAICS Industries

**Purpose**: Reclassifies existing `classification_codes` from "General Business" to the appropriate NAICS-aligned industries based on their NAICS 2-digit prefixes and description keywords.

**Execution Order**: Run this **AFTER** migration 023.

**What it does**:

1. Reclassifies NAICS codes based on their 2-digit prefix (e.g., codes starting with "11" → Agriculture)
2. Reclassifies codes based on description keywords for codes without NAICS codes
3. Displays a summary of reclassified codes

**Copy the entire file below and paste into Supabase SQL Editor:**

---

## 📋 Migration 025 - Update code_keywords for NAICS Industries

**Purpose**: Updates the `code_keywords` table for newly classified codes, ensuring all codes have appropriate keyword mappings.

**Execution Order**: Run this **AFTER** migration 024.

**What it does**:

1. Extracts keywords from code descriptions for newly classified codes
2. Links industry keywords to codes in the same industry
3. Adds NAICS-specific keywords for each new industry
4. Displays a summary of updated keywords

**Copy the entire file below and paste into Supabase SQL Editor:**

---

## 🚀 Execution Instructions

1. **Open Supabase SQL Editor**

   - Navigate to your Supabase project
   - Go to SQL Editor

2. **Run Migration 024**

   - Copy the entire contents of `supabase-migrations/024_reclassify_codes_to_naics_industries.sql`
   - Paste into SQL Editor
   - Click "Run" or press `Ctrl+Enter` (Windows/Linux) or `Cmd+Enter` (Mac)
   - Wait for completion and review the summary output

3. **Run Migration 025**

   - Copy the entire contents of `supabase-migrations/025_update_code_keywords_for_naics_industries.sql`
   - Paste into SQL Editor
   - Click "Run" or press `Ctrl+Enter` (Windows/Linux) or `Cmd+Enter` (Mac)
   - Wait for completion and review the summary output

4. **Refresh PostgREST Schema Cache**

   - After both migrations complete, run:

   ```bash
   ./scripts/refresh_postgrest_schema_api.sh
   ```

5. **Verify Results**
   - Check the summary output from each migration
   - Run verification queries (see below)

---

## ✅ Verification Queries

After running both migrations, you can verify the results with these queries:

### Check Reclassified Codes by Industry

```sql
SELECT
    i.name as industry,
    COUNT(cc.id) as code_count,
    COUNT(DISTINCT cc.code_type) as code_types,
    STRING_AGG(DISTINCT cc.code_type, ', ') as types
FROM classification_codes cc
INNER JOIN industries i ON i.id = cc.industry_id
WHERE i.name IN (
    'Agriculture, Forestry, Fishing and Hunting',
    'Mining, Quarrying, and Oil and Gas Extraction',
    'Utilities',
    'Wholesale Trade',
    'Real Estate and Rental and Leasing',
    'Professional, Scientific, and Technical Services',
    'Management of Companies and Enterprises',
    'Administrative and Support Services',
    'Arts, Entertainment, and Recreation',
    'Other Services',
    'Public Administration'
)
GROUP BY i.name
ORDER BY code_count DESC;
```

### Check Remaining Codes in General Business

```sql
SELECT
    COUNT(*) as remaining_codes,
    COUNT(DISTINCT code_type) as code_types,
    STRING_AGG(DISTINCT code_type, ', ') as types
FROM classification_codes
WHERE industry_id = (SELECT id FROM industries WHERE name = 'General Business');
```

### Check Code Keywords for New Industries

```sql
SELECT
    i.name as industry,
    COUNT(DISTINCT ck.code_id) as codes_with_keywords,
    COUNT(ck.id) as total_keywords,
    ROUND(AVG(ck.relevance_score), 2) as avg_relevance
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
INNER JOIN industries i ON i.id = cc.industry_id
WHERE i.name IN (
    'Agriculture, Forestry, Fishing and Hunting',
    'Mining, Quarrying, and Oil and Gas Extraction',
    'Utilities',
    'Wholesale Trade',
    'Real Estate and Rental and Leasing',
    'Professional, Scientific, and Technical Services',
    'Management of Companies and Enterprises',
    'Administrative and Support Services',
    'Arts, Entertainment, and Recreation',
    'Other Services',
    'Public Administration'
)
GROUP BY i.name
ORDER BY i.name;
```

---

## 📊 Expected Results

After running both migrations, you should see:

1. **Reclassified Codes**: Most codes previously in "General Business" should now be assigned to specific NAICS-aligned industries
2. **Keywords Updated**: All newly classified codes should have keyword mappings in the `code_keywords` table
3. **Industry Distribution**: Codes should be distributed across the 11 new NAICS-aligned industries
4. **Minimal Remaining**: Only a small number of codes (if any) should remain in "General Business"

---

## ⚠️ Notes

- Both migrations are **idempotent** - they can be run multiple times safely
- Migration 024 uses `ON CONFLICT DO NOTHING` where applicable
- Migration 025 uses `ON CONFLICT DO UPDATE` to update existing keywords
- The migrations include summary queries at the end to show what was changed
- If you encounter any errors, check:
  - That migration 023 completed successfully
  - That all industry names match exactly (case-sensitive)
  - That you have the necessary permissions

---

## 🔄 Next Steps

After completing both migrations:

1. ✅ Refresh PostgREST schema cache
2. ✅ Run verification queries
3. ✅ Test hybrid code generation with the new industries
4. ✅ Run test suite: `./scripts/run_hybrid_tests.sh`
