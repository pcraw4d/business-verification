# Code Metadata Population Guide

## Overview

This guide explains how to populate the `code_metadata` table with official code descriptions, crosswalk data, and hierarchies.

## What Was Created

The script `scripts/populate_code_metadata.sql` includes:

1. **Official Code Descriptions**:
   - 15 NAICS codes (Technology, Financial Services, Healthcare, Retail, Food & Beverage)
   - 10 SIC codes (matching industries)
   - 15 MCC codes (matching industries)

2. **Crosswalk Data**:
   - NAICS ↔ SIC ↔ MCC mappings
   - Bidirectional relationships between code types

3. **Code Hierarchies**:
   - Parent/child relationships for NAICS codes
   - Industry groupings

4. **Industry Mappings**:
   - Primary and secondary industry classifications
   - Industry associations for each code

## Running the Population Script

### Step 1: Run in Supabase

1. Open Supabase SQL Editor
2. Copy the entire contents of `scripts/populate_code_metadata.sql`
3. Paste and execute
4. Verify results using the verification queries at the end of the script

### Step 2: Verify Population

After running, you should see:
- **Total Records**: ~40 records
- **Records by Code Type**: 
  - NAICS: ~15
  - SIC: ~10
  - MCC: ~15
- **Records with Crosswalk Data**: ~40 (all should have crosswalks)
- **Records with Hierarchy Data**: ~15 (NAICS codes with hierarchies)
- **Records with Industry Mappings**: ~40 (all should have mappings)

## Using the Code Metadata

### Query Crosswalks

```sql
-- Find all codes related to a NAICS code
SELECT * FROM code_crosswalk_view 
WHERE code_type = 'NAICS' AND code = '541511';

-- Find all MCC codes for a given industry
SELECT mcc_code 
FROM code_crosswalk_view 
WHERE code_type = 'NAICS' AND code = '541511' 
AND mcc_code IS NOT NULL;
```

### Query Hierarchies

```sql
-- Find parent code for a given code
SELECT parent_code, parent_name 
FROM code_hierarchy_view 
WHERE code_type = 'NAICS' AND code = '541511';

-- Find all child codes
SELECT child_code 
FROM code_hierarchy_view 
WHERE code_type = 'NAICS' AND parent_code = '5415';
```

### Query Official Descriptions

```sql
-- Get official description for a code
SELECT official_name, official_description 
FROM code_metadata 
WHERE code_type = 'NAICS' AND code = '541511';
```

## Expanding the Data

### Adding More Codes

To add more codes, follow this pattern:

```sql
INSERT INTO code_metadata (code_type, code, official_name, official_description, is_official, is_active)
VALUES
('NAICS', 'XXXXXX', 'Official Name', 'Official description from Census Bureau', true, true)
ON CONFLICT (code_type, code) DO UPDATE SET
    official_name = EXCLUDED.official_name,
    official_description = EXCLUDED.official_description,
    updated_at = NOW();
```

### Adding Crosswalks

```sql
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['541511', '541512'],
    'sic', ARRAY['7371', '7372'],
    'mcc', ARRAY['5734']
)
WHERE code_type = 'SIC' AND code = '7371';
```

### Adding Hierarchies

```sql
UPDATE code_metadata
SET hierarchy = jsonb_build_object(
    'parent_code', '5415',
    'parent_type', 'NAICS',
    'parent_name', 'Computer Systems Design and Related Services',
    'child_codes', ARRAY['541511', '541512', '541519']
)
WHERE code_type = 'NAICS' AND code IN ('541511', '541512', '541519');
```

## Data Sources

### Official NAICS Data
- **Source**: U.S. Census Bureau
- **URL**: https://www.census.gov/naics/
- **Format**: CSV/Excel files available for download

### Official SIC Data
- **Source**: U.S. Securities and Exchange Commission (SEC)
- **URL**: https://www.sec.gov/info/edgar/siccodes.htm
- **Format**: Text files available

### Official MCC Data
- **Source**: Payment card networks (Visa, Mastercard)
- **URL**: Various payment processor documentation
- **Format**: CSV/Excel files

## Best Practices

1. **Always use official sources** for code descriptions
2. **Verify crosswalk mappings** with multiple sources
3. **Keep hierarchies consistent** with official code structures
4. **Update regularly** as codes change or new codes are added
5. **Mark unofficial data** with `is_official = false`
6. **Use `ON CONFLICT` clauses** to handle updates gracefully

## Maintenance

### Regular Updates

- **NAICS**: Updated every 5 years (last update: 2022)
- **SIC**: Generally stable, but check for updates
- **MCC**: Updated periodically by payment processors

### Monitoring

```sql
-- Check for codes without descriptions
SELECT code_type, code 
FROM code_metadata 
WHERE official_description IS NULL;

-- Check for codes without crosswalks
SELECT code_type, code 
FROM code_metadata 
WHERE crosswalk_data = '{}'::jsonb;

-- Check for codes without industry mappings
SELECT code_type, code 
FROM code_metadata 
WHERE industry_mappings = '{}'::jsonb;
```

## Next Steps

1. ✅ Run the population script
2. ✅ Verify data was inserted correctly
3. ✅ Test crosswalk queries
4. ✅ Test hierarchy queries
5. ⏭️ Expand with more codes as needed
6. ⏭️ Integrate with classification service

