# SQL Script Execution Order for Accuracy Plan Enhancements

This document provides the correct order for executing SQL scripts in the Supabase SQL Editor to implement the accuracy plan enhancements.

## Prerequisites

**IMPORTANT**: Ensure the `code_metadata` table exists before running these scripts. If you haven't run the migration yet, run this first:

- `supabase-migrations/029_add_code_metadata.sql` - Creates the `code_metadata` table and indexes

## Execution Order

Execute these scripts in the following order:

### Step 1: Initial Code Metadata Population (Optional)
**Purpose**: Populate initial seed data with basic code metadata

- `scripts/populate_code_metadata.sql` - Initial population with ~40 codes
  - **Note**: This is optional if you want initial seed data. You can skip this if you're starting fresh.

### Step 2: Expand Code Metadata (Phase 1)
**Purpose**: Add 350+ additional codes to reach 500+ total codes

- `scripts/expand_code_metadata_phase1.sql` - Adds 350+ codes across all major industries
  - **Verification**: Run `scripts/verify_code_metadata_coverage.sql` after this to verify 500+ codes exist

### Step 3: Expand Keywords
**Purpose**: Add keywords for 90%+ of codes (15+ keywords per code)

- `scripts/expand_keywords_phase1.sql` - Populates keywords extracted from official descriptions
  - **Verification**: Check that 90%+ of codes have 15+ keywords

### Step 4: Expand Crosswalks (Phase 2)
**Purpose**: Add crosswalk relationships between code types

Execute in this order:

1. `scripts/expand_mcc_crosswalks.sql` - Adds crosswalks for 30+ MCC codes (45%+ coverage)
2. `scripts/expand_sic_crosswalks.sql` - Adds crosswalks for 50+ SIC codes (57%+ coverage)
3. `scripts/expand_naics_crosswalks.sql` - Adds crosswalks for 50+ NAICS codes (61%+ coverage)
   - **Verification**: Run `scripts/verify_crosswalk_coverage.sql` after all crosswalk scripts

### Step 5: Expand NAICS Hierarchy
**Purpose**: Add parent/child relationships for NAICS codes

- `scripts/expand_naics_hierarchy.sql` - Adds hierarchy data for 50+ NAICS codes (30%+ coverage)
  - **Verification**: Check that 30%+ of NAICS codes have hierarchy data

### Step 6: Expand Industry Mappings
**Purpose**: Add primary/secondary industry classifications

- `scripts/expand_industry_mappings.sql` - Adds industry mappings for 120+ codes (80%+ coverage)
  - **Verification**: Check that 80%+ of codes have industry mappings

## Complete Execution Order Summary

```
1. supabase-migrations/029_add_code_metadata.sql (if not already run)
2. scripts/populate_code_metadata.sql (optional - initial seed data)
3. scripts/expand_code_metadata_phase1.sql
4. scripts/expand_keywords_phase1.sql
5. scripts/expand_mcc_crosswalks.sql
6. scripts/expand_sic_crosswalks.sql
7. scripts/expand_naics_crosswalks.sql
8. scripts/expand_naics_hierarchy.sql
9. scripts/expand_industry_mappings.sql
```

## Verification Scripts

After completing all scripts, run these verification scripts to confirm success:

- `scripts/verify_code_metadata_coverage.sql` - Verify code metadata coverage
- `scripts/verify_crosswalk_coverage.sql` - Verify crosswalk coverage
- `scripts/verify_code_metadata_complete.sql` - Comprehensive verification

## Expected Results

After running all scripts in order, you should have:

- ✅ 500+ codes in `code_metadata` table
- ✅ 90%+ of codes with 15+ keywords
- ✅ 50+ codes with crosswalks (50%+ coverage)
- ✅ 30+ NAICS codes with hierarchy (30%+ coverage)
- ✅ 120+ codes with industry mappings (80%+ coverage)

## Notes

- All scripts use `ON CONFLICT` clauses, so they can be safely re-run
- Scripts are idempotent - running them multiple times won't create duplicates
- Each script includes verification queries at the end
- If a script fails, check the error message and fix any issues before proceeding

