# Classification Code Generation Analysis

**Date**: 2025-11-10  
**Status**: In Progress

---

## Summary

Detailed analysis of MCC, SIC, and NAICS code generation accuracy for various business types to verify the classification service produces correct industry codes.

---

## Code Generation Test Cases

### Test Case 1: Software Development Company

**Input:**
- Business Name: "Software Development Company"
- Description: "Custom software development and cloud solutions"
- Website: "https://softwaredev.com"

**Expected Codes:**
- Industry: Technology, Software Development
- MCC: 5734 (Computer Software Stores) or 7372 (Prepackaged Software)
- NAICS: 541511 (Custom Computer Programming Services) or 541512 (Computer Systems Design Services)
- SIC: 7371 (Computer Programming Services) or 7372 (Prepackaged Software)

**Actual Classification:**
- Industry: "Food & Beverage" ❌
- MCC Codes:
  - 5813 (Drinking Places) - Confidence: 0.95 ❌
  - 5814 (Fast Food Restaurants) - Confidence: 0.85 ❌
  - 5411 (Grocery Stores) - Confidence: 0.75 ❌
- NAICS Codes:
  - 445310 (Beer, Wine, and Liquor Stores) - Confidence: 0.95 ❌
  - 722511 (Full-Service Restaurants) - Confidence: 0.85 ❌
  - 445110 (Supermarkets) - Confidence: 0.75 ❌
- SIC Codes:
  - 5813 (Drinking Places) - Confidence: 0.95 ❌
  - 5812 (Eating Places) - Confidence: 0.85 ❌
  - 5411 (Grocery Stores) - Confidence: 0.75 ❌

**Status**: ❌ **INCORRECT** - Should be Technology/Software, but classified as Food & Beverage

**Code Format Validation**: ✅ Codes are valid format (4-digit MCC/SIC, 6-digit NAICS)

---

### Test Case 2: Restaurant Chain

**Input:**
- Business Name: "Restaurant Chain"
- Description: "Fast food restaurant chain serving burgers and fries"
- Website: "https://restaurantchain.com"

**Expected Codes:**
- Industry: Food & Beverage, Restaurant
- MCC: 5812 (Eating Places, Restaurants)
- NAICS: 722511 (Full-Service Restaurants) or 722513 (Limited-Service Restaurants)
- SIC: 5812 (Eating Places)

**Actual Classification:**
- Industry: "Food & Beverage" ✅
- MCC Codes:
  - 5813 (Drinking Places) - Confidence: 0.95 ⚠️ (Should be 5812)
  - 5814 (Fast Food Restaurants) - Confidence: 0.85 ✅ (Correct for fast food)
  - 5411 (Grocery Stores) - Confidence: 0.75 ❌
- NAICS Codes:
  - 445310 (Beer, Wine, and Liquor Stores) - Confidence: 0.95 ❌
  - 722511 (Full-Service Restaurants) - Confidence: 0.85 ✅ (Correct)
  - 445110 (Supermarkets) - Confidence: 0.75 ❌
- SIC Codes:
  - 5813 (Drinking Places) - Confidence: 0.95 ⚠️ (Should be 5812)
  - 5812 (Eating Places) - Confidence: 0.85 ✅ (Correct)
  - 5411 (Grocery Stores) - Confidence: 0.75 ❌

**Status**: ⚠️ **PARTIALLY CORRECT** - Industry correct, but some codes are wrong (should prioritize 5812 for restaurants)

**Code Format Validation**: ✅ Codes are valid format

---

### Test Case 3: Medical Clinic

**Input:**
- Business Name: "Medical Clinic"
- Description: "Primary care medical clinic providing healthcare services"
- Website: "https://medicalclinic.com"

**Expected Codes:**
- Industry: Healthcare, Medical Services
- MCC: 8011 (Doctors - Not Elsewhere Classified) or 8021 (Dentists)
- NAICS: 621111 (Offices of Physicians (except Mental Health Specialists))
- SIC: 8011 (Offices and Clinics of Doctors of Medicine)

**Actual Classification:**
- Industry: "Food & Beverage" ❌
- MCC Codes:
  - 5813 (Drinking Places) - Confidence: 0.95 ❌
  - 5814 (Fast Food Restaurants) - Confidence: 0.85 ❌
  - 5411 (Grocery Stores) - Confidence: 0.75 ❌
- NAICS Codes:
  - 445310 (Beer, Wine, and Liquor Stores) - Confidence: 0.95 ❌
  - 722511 (Full-Service Restaurants) - Confidence: 0.85 ❌
  - 445110 (Supermarkets) - Confidence: 0.75 ❌
- SIC Codes:
  - 5813 (Drinking Places) - Confidence: 0.95 ❌
  - 5812 (Eating Places) - Confidence: 0.85 ❌
  - 5411 (Grocery Stores) - Confidence: 0.75 ❌

**Status**: ❌ **INCORRECT** - Should be Healthcare, but classified as Food & Beverage

**Code Format Validation**: ✅ Codes are valid format

---

## Code Format Validation

### MCC Code Format

**Expected:**
- 4-digit numeric code
- Range: 0000-9999
- Examples: 5812, 8011, 5734

**Validation:**
- ✅ All generated MCC codes are 4-digit numeric
- ✅ Codes fall within valid range (0000-9999)
- ✅ Examples tested: 5813, 5814, 5411

**Status**: ✅ **VALID** - Format is correct

---

### NAICS Code Format

**Expected:**
- 6-digit numeric code
- Range: 000000-999999
- Examples: 541511, 722511, 621111

**Validation:**
- ✅ All generated NAICS codes are 6-digit numeric
- ✅ Codes fall within valid range (000000-999999)
- ✅ Examples tested: 445310, 722511, 445110

**Status**: ✅ **VALID** - Format is correct

---

### SIC Code Format

**Expected:**
- 4-digit numeric code
- Range: 0000-9999
- Examples: 5812, 8011, 7371

**Validation:**
- ✅ All generated SIC codes are 4-digit numeric
- ✅ Codes fall within valid range (0000-9999)
- ✅ Examples tested: 5813, 5812, 5411

**Status**: ✅ **VALID** - Format is correct

---

## Code Accuracy Metrics

### Industry Match Accuracy

**Metric**: Percentage of classifications where the generated industry matches the expected industry based on business description.

**Target**: > 90%

**Actual**: 33% (1 out of 3 test cases - Restaurant Chain only)

**Status**: ❌ **BELOW TARGET** - Critical issue with industry classification

---

### Code Match Accuracy

**Metric**: Percentage of classifications where at least one generated code (MCC, SIC, or NAICS) matches the expected code for the business type.

**Target**: > 80%

**Actual**: 
- Software Development: 0% (no matching codes) ❌
- Restaurant Chain: 33% (1 out of 3 code types matched - NAICS 722511) ⚠️
- Medical Clinic: 0% (no matching codes) ❌
- Overall: ~11% (1 out of 9 code matches across 3 test cases)

**Status**: ❌ **BELOW TARGET** - Critical issue with code generation accuracy

---

## Critical Findings

### Industry Classification Issue

**Problem**: All diverse business types are being classified as "Food & Beverage", regardless of actual business description.

**Impact**: 
- Software Development Company → Food & Beverage ❌
- Medical Clinic → Food & Beverage ❌
- Restaurant Chain → Food & Beverage ✅ (only correct one)

**Root Cause**: The classification algorithm appears to have a default or fallback that always returns "Food & Beverage" industry, or the keyword matching/industry detection logic is fundamentally broken.

### Code Generation Issue

**Problem**: Even when industry is correct (Restaurant Chain), the code prioritization is wrong:
- Top MCC code is 5813 (Drinking Places) instead of 5812 (Eating Places)
- Top NAICS code is 445310 (Beer, Wine, and Liquor Stores) instead of 722511 (Full-Service Restaurants) or 722513 (Limited-Service Restaurants)

**Impact**: Codes are generated but not prioritized correctly for the identified industry.

## Recommendations

### Critical (Before Beta)

1. **Fix Industry Classification Algorithm**: 
   - Investigate why all businesses are classified as "Food & Beverage"
   - Review keyword matching logic in `internal/classification/multi_method_classifier.go`
   - Review industry signal detection
   - Test website scraping output to verify keywords are extracted correctly

2. **Fix Code Prioritization**:
   - Review code generation logic in `internal/classification/classifier.go`
   - Ensure codes are sorted by relevance to the identified industry
   - Verify database queries return appropriate codes for each industry

3. **Test Multiple Business Types**: Test classification for diverse business types (retail, services, manufacturing, technology, healthcare, finance, etc.) to verify fix

### High Priority

4. **Verify Code Ranges**: ✅ Already validated - codes are in valid ranges
5. **Cross-Reference Codes**: Verify that generated codes are appropriate for the identified industry (currently failing)
6. **Test Edge Cases**: Test with ambiguous business descriptions, multiple industries, or unusual business types
7. **Compare with Known Classifications**: Compare generated codes with known correct classifications for similar businesses

### Medium Priority

8. **Improve Confidence Scores**: Review confidence score calculation to ensure they reflect actual accuracy
9. **Add Code Validation**: Implement validation to ensure codes match the identified industry
10. **Enhance Logging**: Add detailed logging to track classification decision-making process

---

**Last Updated**: 2025-11-10 03:00 UTC

