# Code Metadata Expansion Usage Guide

## Issue Fixed

The original `expand_code_metadata.sql` had duplicate codes within the same INSERT statement, causing the error:

```
ERROR: ON CONFLICT DO UPDATE command cannot affect row a second time
```

## Solution

I've created a **clean version** that:

1. ✅ Removes duplicate MCC code '5733' (was appearing twice)
2. ✅ Removes codes that already exist in `populate_code_metadata.sql`:
   - MCC: 5734, 5735, 5811, 5812, 5813, 5814, 8011, 8021, 8041, 8042, 8043, 8049, 8050
3. ✅ Uses proper PostgreSQL syntax (`''` instead of `\'` for apostrophes)

## Usage

### Option 1: Use the Clean Version (Recommended)

**File**: `scripts/expand_code_metadata_clean.sql`

This version excludes codes that are already in `populate_code_metadata.sql` to avoid confusion.

```sql
-- In Supabase SQL Editor, run:
-- scripts/expand_code_metadata_clean.sql
```

### Option 2: Use the Fixed Original

**File**: `scripts/expand_code_metadata.sql`

This version has been fixed but includes some codes that may already exist. The `ON CONFLICT` clause will handle updates gracefully.

```sql
-- In Supabase SQL Editor, run:
-- scripts/expand_code_metadata.sql
```

## Execution Order

1. **First**: Run `populate_code_metadata.sql` (if not already done)

   - Adds ~40 codes with basic metadata

2. **Second**: Run `expand_code_metadata_clean.sql` (recommended)
   - Adds ~100+ additional codes
   - Excludes duplicates from step 1

## What Gets Added

### New Codes Added (in clean version):

- **NAICS**: ~30 additional codes
- **SIC**: ~20 additional codes
- **MCC**: ~50 additional codes (excluding duplicates)

### All Codes Include:

- ✅ Official descriptions
- ✅ Crosswalk mappings (NAICS ↔ SIC ↔ MCC)
- ✅ Industry mappings
- ✅ Code hierarchies (for NAICS)

## Verification

After running, verify with:

```sql
-- Count total records
SELECT COUNT(*) FROM code_metadata;

-- Should show ~140+ total codes after expansion
-- (40 from populate + 100+ from expand)
```

## Troubleshooting

### Error: "duplicate key value violates unique constraint"

**Cause**: Code already exists from a previous run
**Solution**: This is fine - `ON CONFLICT DO UPDATE` will update existing records

### Error: "syntax error at or near"

**Cause**: Apostrophe escaping issue
**Solution**: Use the clean version which has proper `''` escaping

### Error: "ON CONFLICT DO UPDATE command cannot affect row a second time"

**Cause**: Duplicate codes in the same INSERT statement
**Solution**: Use `expand_code_metadata_clean.sql` which has duplicates removed

## Summary

- ✅ **Fixed**: All SQL syntax errors
- ✅ **Fixed**: Removed duplicate MCC '5733'
- ✅ **Created**: Clean version without conflicts
- ✅ **Ready**: Both scripts are ready to run

**Recommended**: Use `expand_code_metadata_clean.sql` for the cleanest execution.
