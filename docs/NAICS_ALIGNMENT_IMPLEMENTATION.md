# NAICS Alignment Implementation Guide

## Overview

This document describes the migration process to align the classification system with official NAICS 2-digit sector codes. This replaces the broad "General Business" category with specific, standardized industries.

## Migration Files

### 1. `023_add_naics_aligned_industries.sql`

- Adds 11 new NAICS-aligned industries
- Updates descriptions of existing industries to reference NAICS codes
- Adds comprehensive keywords for each new industry

### 2. `024_reclassify_codes_to_naics_industries.sql`

- Reclassifies existing codes from "General Business" to appropriate NAICS-aligned industries
- Uses NAICS code prefixes (2-digit) for automatic classification
- Uses description keywords as fallback for non-NAICS codes

### 3. `025_update_code_keywords_for_naics_industries.sql`

- Updates code_keywords mappings for newly classified codes
- Links industry keywords to codes
- Adds NAICS-specific keywords

## NAICS Sector Mapping

| NAICS Code | Sector Name                                      | Industry Name in System                          |
| ---------- | ------------------------------------------------ | ------------------------------------------------ |
| 11         | Agriculture, Forestry, Fishing and Hunting       | Agriculture, Forestry, Fishing and Hunting       |
| 21         | Mining, Quarrying, and Oil and Gas Extraction    | Mining, Quarrying, and Oil and Gas Extraction    |
| 22         | Utilities                                        | Utilities                                        |
| 23         | Construction                                     | Construction                                     |
| 31-33      | Manufacturing                                    | Manufacturing                                    |
| 42         | Wholesale Trade                                  | Wholesale Trade                                  |
| 44-45      | Retail Trade                                     | Retail                                           |
| 48-49      | Transportation and Warehousing                   | Transportation                                   |
| 51         | Information                                      | Technology                                       |
| 52         | Finance and Insurance                            | Finance                                          |
| 53         | Real Estate and Rental and Leasing               | Real Estate and Rental and Leasing               |
| 54         | Professional, Scientific, and Technical Services | Professional, Scientific, and Technical Services |
| 55         | Management of Companies and Enterprises          | Management of Companies and Enterprises          |
| 56         | Administrative and Support Services              | Administrative and Support Services              |
| 61         | Educational Services                             | Education                                        |
| 62         | Health Care and Social Assistance                | Healthcare                                       |
| 71         | Arts, Entertainment, and Recreation              | Arts, Entertainment, and Recreation              |
| 72         | Accommodation and Food Services                  | Food & Beverage                                  |
| 81         | Other Services                                   | Other Services                                   |
| 92         | Public Administration                            | Public Administration                            |

## Execution Order

1. **Run Migration 023**: Adds new industries and keywords

   ```sql
   -- In Supabase SQL Editor, run:
   -- supabase-migrations/023_add_naics_aligned_industries.sql
   ```

2. **Run Migration 024**: Reclassifies existing codes

   ```sql
   -- In Supabase SQL Editor, run:
   -- supabase-migrations/024_reclassify_codes_to_naics_industries.sql
   ```

3. **Run Migration 025**: Updates code keywords

   ```sql
   -- In Supabase SQL Editor, run:
   -- supabase-migrations/025_update_code_keywords_for_naics_industries.sql
   ```

4. **Refresh PostgREST Schema Cache**

   ```bash
   ./scripts/refresh_postgrest_schema_api.sh
   ```

5. **Verify Results**
   ```bash
   ./scripts/run_hybrid_tests.sh
   ```

## Expected Results

After running all migrations:

- **11 new industries** added
- **Existing industries** updated with NAICS references
- **Codes reclassified** from "General Business" to specific industries
- **Keywords added** for all new industries
- **Code keywords updated** for newly classified codes
- **"General Business"** remains as ultimate fallback (should have minimal codes)

## Verification Queries

Run these in Supabase SQL Editor to verify:

```sql
-- Check industry count
SELECT COUNT(*) as total_industries FROM industries WHERE is_active = true;

-- Check codes per industry
SELECT
    i.name,
    COUNT(cc.id) as code_count
FROM industries i
LEFT JOIN classification_codes cc ON cc.industry_id = i.id
WHERE i.is_active = true
GROUP BY i.name
ORDER BY code_count DESC;

-- Check remaining codes in General Business
SELECT COUNT(*) as remaining_in_general_business
FROM classification_codes
WHERE industry_id = (SELECT id FROM industries WHERE name = 'General Business');

-- Check keywords per new industry
SELECT
    i.name,
    COUNT(ik.id) as keyword_count
FROM industries i
LEFT JOIN industry_keywords ik ON ik.industry_id = i.id
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

## Benefits

1. **Standardized Classification**: Aligned with official NAICS structure
2. **Better Accuracy**: More specific industries lead to better classification
3. **Industry Standard**: Uses recognized industry naming conventions
4. **Easier Integration**: Can integrate with external NAICS data sources
5. **Reduced Fallback**: Fewer businesses will fall into "General Business"
6. **Better Analytics**: More granular industry data for reporting

## Notes

- "General Business" (ID 26) is kept as the ultimate fallback
- Existing industries are updated but not renamed (to maintain compatibility)
- New industries use official NAICS sector names
- Keywords are comprehensive and based on NAICS descriptions
- Code reclassification is based on NAICS prefixes for automatic matching
