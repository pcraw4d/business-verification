# Crosswalk and Keyword Data Investigation

## Summary

Investigated missing crosswalk and keyword data that's causing test failures.

## Issues Identified

### 1. Crosswalk Data Gaps ✅

**Problem:** MCC 5819 (and likely other codes) missing crosswalk data in `code_metadata` table.

**Root Cause:**
- `code_metadata` table exists with `crosswalk_data` JSONB field
- Crosswalk data is stored as JSONB: `{"naics": ["..."], "sic": ["..."], "mcc": ["..."]}`
- Many MCC codes don't have crosswalk data populated

**MCC 5819 Definition:**
- Official: "Miscellaneous Food Stores - Convenience Stores, Specialty Markets, Vending Machines"
- Also used for: Pizza restaurants and food service establishments not elsewhere classified

**Solution:**
- Created migration `041_populate_missing_crosswalks.sql` to populate crosswalk data for:
  - MCC 5819 → SIC 5499, 5812, 5411; NAICS 445110, 445120, 722511, 722513
  - MCC 5812 → SIC 5812; NAICS 722511, 722513
  - MCC 5814 → SIC 5812; NAICS 722513, 722515
  - Reverse crosswalks (SIC/NAICS → MCC)

### 2. Keyword Data Gaps ✅

**Problem:** Technology keywords like "cloud", "computing", "software" not found in `industry_keywords` table.

**Root Cause:**
- `industry_keywords` table exists and is queried by `GetIndustriesByKeyword`
- Some technology keywords may be missing or have low weights
- Keywords need to match exactly (case-insensitive via `Ilike`)

**Solution:**
- Created migration `042_populate_technology_keywords.sql` to add:
  - "cloud", "cloud computing", "cloud services" (weight: 0.95)
  - "computing" (weight: 0.90)
  - "software", "saas" (weight: 0.90)
  - "platform", "technology", "tech" (weight: 0.85-0.80)
  - "it services", "information technology", "solutions", "services"

## Migrations Created

### 1. `041_populate_missing_crosswalks.sql`
- Populates crosswalk data for common MCC codes (5819, 5812, 5814)
- Creates reverse crosswalks (SIC/NAICS → MCC)
- Ensures bidirectional mapping for gap filling

### 2. `042_populate_technology_keywords.sql`
- Adds technology keywords to "Technology" industry
- Also adds to "Software" industry if it exists separately
- Uses appropriate weights and contexts

## Verification Script

Created `scripts/verify_crosswalk_keyword_data.sql` to:
- Check which MCC codes have crosswalk data
- Check which technology keywords exist
- Count missing crosswalks
- Verify keyword distribution by industry

## Next Steps

1. **Run Migrations:**
   ```bash
   # Apply migrations to populate data
   # Migration 041: Crosswalk data
   # Migration 042: Technology keywords
   ```

2. **Verify Data:**
   ```sql
   -- Run verification script
   \i scripts/verify_crosswalk_keyword_data.sql
   ```

3. **Test After Migration:**
   - Re-run test suite to verify gap filling works
   - Check if Cloud Services keywords are extracted
   - Verify Top 3 Codes test passes

## Expected Improvements

After running migrations:
- **Top 3 Codes:** Should fill gaps using crosswalk data
- **Cloud Services Keywords:** Should extract "cloud" and "computing" from description
- **Gap Filling:** Should work for MCC → SIC/NAICS and reverse
