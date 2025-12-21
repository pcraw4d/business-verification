# Railway E2E Test - Sample Size Increase to 385

## Overview

The Railway E2E classification test sample size has been increased from **12 samples to 385 samples** to achieve **95% statistical confidence** with a **5% margin of error**.

## Statistical Justification

- **Previous Sample Size**: 12 samples
  - Margin of error: ~±28% at 95% confidence
  - Statistical power: Insufficient for reliable conclusions

- **New Sample Size**: 385 samples
  - Margin of error: ±5% at 95% confidence
  - Statistical power: Sufficient for reliable accuracy and performance metrics

## Sample Distribution

### Real-World Businesses (~50 samples)

Well-known companies across major industries:
- **E-commerce & Retail**: Amazon, Shopify, eBay, Walmart, Target
- **Technology**: Microsoft, Apple, Google, Meta, Stripe, Salesforce, Oracle, IBM
- **Food & Beverage**: Starbucks, McDonald's, Coca-Cola, PepsiCo, Domino's, Subway
- **Healthcare**: UnitedHealth Group, CVS Health, Walgreens, Mayo Clinic
- **Financial Services**: JPMorgan Chase, Bank of America, Wells Fargo, Goldman Sachs, PayPal
- **Manufacturing**: Tesla, Ford, General Electric
- **Entertainment**: Netflix, Disney, Spotify
- **Professional Services**: Deloitte, PwC, EY
- **Construction**: Caterpillar
- **Real Estate**: Zillow
- **Transportation**: Uber, FedEx
- **Energy**: ExxonMobil

### Programmatically Generated Samples (~335 samples)

Diverse samples generated across 14 major industries:

| Industry | Sample Count | MCC Codes | Categories | Complexity Levels | Scraping Difficulty |
|----------|--------------|-----------|------------|-------------------|---------------------|
| Technology | 50 | 5734, 7372 | saas, software, tech | low, medium, high | low, medium |
| Retail | 45 | 5311, 5331, 5999 | ecommerce, retail, marketplace | low, medium, high | low, medium, high |
| Food & Beverage | 40 | 5812, 5814, 5499 | restaurant, cafe, bar | low, medium | none, low, medium |
| Healthcare | 35 | 8011, 5912, 6300 | healthcare, medical, pharmacy | low, medium, high | low, medium |
| Financial Services | 35 | 6011, 6012, 6211 | banking, fintech, insurance | medium, high | low, medium |
| Manufacturing | 30 | 5511, 5533, 5084 | manufacturing, industrial | medium, high | low, medium |
| Professional Services | 30 | 8931, 8999, 7392 | consulting, legal, accounting | medium, high | low, medium |
| Construction | 25 | 1711, 1521, 1541 | construction, contractor | low, medium | none, low |
| Arts & Entertainment | 25 | 7829, 7832, 5735 | entertainment, media, streaming | low, medium | low, medium |
| Real Estate | 20 | 6513, 6531, 1521 | real estate, property | low, medium | low, medium |
| Transportation | 20 | 4121, 4214, 4789 | transportation, logistics | low, medium, high | low, medium |
| Education | 15 | 8299, 8220, 8241 | education, training | low, medium | low, medium |
| Energy | 15 | 5542, 5541, 5983 | energy, utilities | medium, high | low, medium |
| Agriculture | 10 | 5999, 5261, 5193 | agriculture, farming | low, medium | none, low |

### Small Business Samples (No Website)

Additional samples representing small businesses without websites:
- Local restaurants (pizza, bakery, cafe)
- Service providers (plumbing, auto repair, cleaning)
- Retail shops (grocery, print shop)
- Professional services (dentistry, tech support)

## Sample Characteristics

### Complexity Distribution
- **Low Complexity**: ~30% (115 samples)
- **Medium Complexity**: ~40% (154 samples)
- **High Complexity**: ~30% (116 samples)

### Scraping Difficulty Distribution
- **None** (No website): ~10% (38 samples)
- **Low**: ~40% (154 samples)
- **Medium**: ~40% (154 samples)
- **High**: ~10% (39 samples)

### Website Coverage
- **With Website**: ~90% (347 samples)
- **Without Website**: ~10% (38 samples)

## Implementation Details

### Code Changes

**File**: `test/integration/railway_comprehensive_e2e_classification_test.go`

**Function**: `generateComprehensiveTestSamples()`

**Changes**:
1. Expanded real-world business samples from 12 to ~50
2. Added programmatic sample generation for 14 industries
3. Implemented intelligent distribution across complexity and scraping difficulty levels
4. Added small business samples without websites
5. Ensured exactly 385 samples are generated

### Sample Generation Algorithm

1. **Real-World Samples**: Hardcoded list of well-known businesses
2. **Programmatic Generation**: 
   - Industry templates with MCC codes, categories, and difficulty levels
   - Business name generation using prefixes and suffixes
   - URL generation for businesses with websites
   - Balanced distribution across all dimensions
3. **Small Business Samples**: Rotating list of local business types without websites

## Expected Test Duration

With 385 samples:
- **Concurrency**: 3 concurrent requests
- **Delay**: 500ms between requests
- **Estimated Duration**: ~2-3 hours
  - 385 samples ÷ 3 concurrent = ~128 batches
  - ~180 seconds per request (timeout)
  - ~64 seconds per batch (with delays)
  - Total: ~2-3 hours

## Benefits

1. **Statistical Confidence**: 95% confidence with ±5% margin of error
2. **Industry Coverage**: Comprehensive coverage across 14+ industries
3. **Complexity Diversity**: Tests across all complexity levels
4. **Scraping Scenarios**: Includes various scraping difficulty levels
5. **Real-World Validation**: Includes 50 well-known businesses for validation
6. **Edge Cases**: Includes small businesses without websites

## Validation

The test suite will validate:
- **Overall Classification Accuracy**: ≥80%
- **Code Generation Rate**: ≥90%
- **Overall Code Accuracy**: ≥70%
- **MCC Top 3 Accuracy**: ≥60%
- **Scraping Success Rate**: ≥70%
- **Average Latency**: <10,000ms

## Next Steps

1. Run the enhanced E2E test suite:
   ```bash
   ./test/scripts/run_railway_e2e_classification_tests.sh
   ```

2. Review generated reports:
   - `test/results/railway_e2e_classification_*.json`
   - `test/results/railway_e2e_analysis_*.json`

3. Analyze results for:
   - Industry-specific accuracy patterns
   - Code accuracy by industry
   - Performance bottlenecks
   - Scraping success rates by difficulty

## Notes

- The test will take significantly longer (~2-3 hours) but provides statistically valid results
- Consider running during off-peak hours or scheduling as a nightly job
- Monitor Railway service limits and rate limiting during execution
- Results will provide actionable insights for algorithm improvements

---

**Date**: 2025-01-20  
**Sample Size**: 385 samples  
**Statistical Confidence**: 95% (±5% margin of error)

