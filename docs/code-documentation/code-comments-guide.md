# KYB Platform - Code Comments Guide

This document provides comprehensive guidelines for adding code comments to the KYB Platform. It covers standards, best practices, and examples for documenting Go code effectively.

## Table of Contents

1. [Comment Standards](#comment-standards)
2. [Package Documentation](#package-documentation)
3. [Function Documentation](#function-documentation)
4. [Type Documentation](#type-documentation)
5. [Variable Documentation](#variable-documentation)
6. [Example Comments](#example-comments)
7. [Best Practices](#best-practices)
8. [Tools and Automation](#tools-and-automation)

## Comment Standards

### General Principles

**1. Clarity Over Brevity**
- Comments should explain "why" not "what"
- Use clear, concise language
- Avoid redundant comments that restate obvious code

**2. Consistency**
- Follow GoDoc conventions
- Use consistent formatting and style
- Maintain consistent terminology

**3. Completeness**
- Document all exported functions, types, and packages
- Include examples for complex functionality
- Provide context for business logic

### Comment Format

**Single-line Comments:**
```go
// This is a single-line comment
```

**Multi-line Comments:**
```go
/*
This is a multi-line comment
that spans multiple lines
*/
```

**GoDoc Comments:**
```go
// PackageName provides functionality for...
package packagename

// FunctionName does something specific...
func FunctionName() {
    // Implementation
}
```

## Package Documentation

### Package-Level Comments

Every package should have a package-level comment that explains its purpose and functionality.

**Example:**
```go
// Package classification provides business classification functionality
// for the KYB Platform. It includes methods for classifying businesses
// using industry-standard codes (NAICS, SIC, MCC) and provides
// confidence scoring for classification accuracy.
//
// The package supports multiple classification methods:
//   - Keyword-based classification
//   - Fuzzy matching algorithms
//   - Hybrid classification approach
//
// Example usage:
//
//	client := classification.NewClient()
//	result, err := client.Classify("Acme Corporation")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("NAICS Code: %s\n", result.PrimaryClassification.NAICSCode)
package classification
```

### Package Organization Comments

Use comments to organize package structure and explain relationships between components.

**Example:**
```go
// Package auth provides authentication and authorization functionality.
//
// The package is organized into the following components:
//   - Service: Core authentication logic
//   - Middleware: HTTP middleware for authentication
//   - Models: Data structures for users and roles
//   - Utils: Helper functions for token management
package auth
```

## Function Documentation

### Exported Function Comments

All exported functions must have GoDoc comments that explain their purpose, parameters, return values, and any side effects.

**Example:**
```go
// Classify performs business classification using the provided business name
// and optional additional information. It returns a ClassificationResult
// containing the primary classification and confidence score.
//
// The function uses a hybrid classification approach combining keyword
// matching and fuzzy algorithms to achieve high accuracy. Confidence
// scores range from 0.0 to 1.0, with higher scores indicating greater
// reliability.
//
// Parameters:
//   - businessName: The name of the business to classify (required)
//   - options: Optional classification parameters including address,
//     website, and description
//
// Returns:
//   - ClassificationResult: Contains classification details and confidence
//   - error: Returns an error if classification fails or input is invalid
//
// Example:
//
//	result, err := Classify("Acme Software Solutions", ClassificationOptions{
//	    Address: "123 Tech Street, San Francisco, CA 94105",
//	    Website: "https://acmesoftware.com",
//	})
//	if err != nil {
//	    log.Printf("Classification failed: %v", err)
//	    return
//	}
//	fmt.Printf("NAICS Code: %s (Confidence: %.2f)\n",
//	    result.PrimaryClassification.NAICSCode,
//	    result.ConfidenceScore)
func Classify(businessName string, options ClassificationOptions) (*ClassificationResult, error) {
    // Implementation
}
```

### Private Function Comments

Private functions should have comments that explain complex logic or non-obvious behavior.

**Example:**
```go
// calculateConfidenceScore computes the confidence score for a classification
// result based on multiple factors including keyword matches, fuzzy similarity,
// and historical accuracy. The score is normalized to a 0.0-1.0 range.
//
// The algorithm weights different factors as follows:
//   - Keyword matches: 40%
//   - Fuzzy similarity: 30%
//   - Historical accuracy: 20%
//   - Data quality: 10%
func calculateConfidenceScore(matches []KeywordMatch, similarity float64, historicalAccuracy float64) float64 {
    // Implementation
}
```

### Error Handling Comments

Document error conditions and their meanings.

**Example:**
```go
// validateBusinessName checks if the provided business name is valid
// for classification. It returns an error if the name is empty,
// too short, or contains invalid characters.
//
// Error conditions:
//   - ErrEmptyName: Business name is empty or whitespace only
//   - ErrNameTooShort: Business name is less than 2 characters
//   - ErrInvalidCharacters: Business name contains disallowed characters
func validateBusinessName(name string) error {
    if strings.TrimSpace(name) == "" {
        return ErrEmptyName
    }
    if len(strings.TrimSpace(name)) < 2 {
        return ErrNameTooShort
    }
    // Additional validation...
    return nil
}
```

## Type Documentation

### Struct Documentation

Document structs with their purpose, field meanings, and usage examples.

**Example:**
```go
// ClassificationResult represents the result of a business classification
// operation. It contains the primary classification, alternative options,
// and confidence metrics.
type ClassificationResult struct {
    // BusinessID is the unique identifier for the classified business.
    // It is generated based on the business name and address.
    BusinessID string `json:"business_id"`

    // PrimaryClassification contains the most likely industry classification
    // based on the analysis. This is the recommended classification to use.
    PrimaryClassification *Classification `json:"primary_classification"`

    // AlternativeClassifications contains other possible classifications
    // with lower confidence scores. These can be used for validation
    // or when the primary classification is uncertain.
    AlternativeClassifications []*Classification `json:"alternative_classifications,omitempty"`

    // ConfidenceScore indicates the reliability of the classification
    // result. Scores range from 0.0 (low confidence) to 1.0 (high confidence).
    // Scores above 0.8 are considered highly reliable.
    ConfidenceScore float64 `json:"confidence_score"`

    // ClassificationMethod indicates which algorithm was used for
    // the classification (e.g., "keyword", "fuzzy", "hybrid").
    ClassificationMethod string `json:"classification_method"`

    // KeywordsMatched contains the keywords that contributed to
    // the classification decision.
    KeywordsMatched []string `json:"keywords_matched,omitempty"`

    // CreatedAt is the timestamp when the classification was performed.
    CreatedAt time.Time `json:"created_at"`
}
```

### Interface Documentation

Document interfaces with their purpose and expected behavior.

**Example:**
```go
// Classifier defines the interface for business classification operations.
// Implementations should provide thread-safe classification functionality
// with proper error handling and performance characteristics.
type Classifier interface {
    // Classify performs classification of a single business entity.
    // The method should return results within 500ms for optimal performance.
    // Implementations must handle concurrent requests safely.
    Classify(ctx context.Context, business *Business) (*ClassificationResult, error)

    // ClassifyBatch performs classification of multiple business entities.
    // The method should process businesses efficiently and provide progress
    // updates for large batches. Results should be returned in the same
    // order as the input businesses.
    ClassifyBatch(ctx context.Context, businesses []*Business) ([]*ClassificationResult, error)

    // GetClassificationHistory retrieves historical classification results
    // for a business. The method should support pagination and filtering
    // by date range and confidence score.
    GetClassificationHistory(ctx context.Context, businessID string, options HistoryOptions) ([]*ClassificationResult, error)
}
```

### Enum Documentation

Document enums and constants with their meanings and usage.

**Example:**
```go
// RiskLevel represents the overall risk assessment level for a business.
// Higher risk levels require more frequent monitoring and review.
type RiskLevel string

const (
    // RiskLevelLow indicates minimal risk with standard monitoring requirements.
    // Businesses with low risk typically require annual reviews.
    RiskLevelLow RiskLevel = "low"

    // RiskLevelMedium indicates moderate risk requiring enhanced monitoring.
    // Businesses with medium risk typically require quarterly reviews.
    RiskLevelMedium RiskLevel = "medium"

    // RiskLevelHigh indicates elevated risk requiring frequent monitoring.
    // Businesses with high risk typically require monthly reviews.
    RiskLevelHigh RiskLevel = "high"

    // RiskLevelCritical indicates high risk requiring immediate attention.
    // Businesses with critical risk require immediate review and action.
    RiskLevelCritical RiskLevel = "critical"
)
```

## Variable Documentation

### Global Variable Comments

Document global variables and constants with their purpose and usage.

**Example:**
```go
// DefaultClassificationTimeout is the default timeout for classification
// operations. It is used when no specific timeout is provided in the
// context or options.
const DefaultClassificationTimeout = 30 * time.Second

// MaxBatchSize is the maximum number of businesses that can be processed
// in a single batch operation. Larger batches are automatically split
// into smaller chunks for processing.
const MaxBatchSize = 1000

// SupportedClassificationSystems contains the industry classification
// systems supported by the platform. Each system has different coverage
// and accuracy characteristics.
var SupportedClassificationSystems = []string{
    "NAICS", // North American Industry Classification System
    "SIC",   // Standard Industrial Classification
    "MCC",   // Merchant Category Code
}
```

### Local Variable Comments

Add comments for complex local variables or non-obvious logic.

**Example:**
```go
func processClassificationBatch(businesses []*Business) []*ClassificationResult {
    // Use a buffered channel to limit concurrent processing
    // and prevent memory exhaustion with large batches
    semaphore := make(chan struct{}, maxConcurrency)
    
    // Track results in a thread-safe manner using a mutex
    // to ensure proper ordering of results
    var results []*ClassificationResult
    var mu sync.Mutex
    
    // Process businesses concurrently with controlled parallelism
    var wg sync.WaitGroup
    for _, business := range businesses {
        wg.Add(1)
        go func(b *Business) {
            defer wg.Done()
            
            // Acquire semaphore to limit concurrent operations
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            // Perform classification with timeout
            result, err := classifyWithTimeout(b, DefaultClassificationTimeout)
            if err != nil {
                log.Printf("Classification failed for %s: %v", b.Name, err)
                return
            }
            
            // Safely append result to shared slice
            mu.Lock()
            results = append(results, result)
            mu.Unlock()
        }(business)
    }
    
    wg.Wait()
    return results
}
```

## Example Comments

### Complex Algorithm Comments

**Example:**
```go
// calculateFuzzySimilarity computes the similarity between two business names
// using the Levenshtein distance algorithm with additional optimizations
// for common business name patterns.
//
// The algorithm applies the following optimizations:
// 1. Normalize business names (remove punctuation, convert to lowercase)
// 2. Apply common abbreviations (Inc -> Incorporated, Corp -> Corporation)
// 3. Weight different types of differences:
//    - Character substitutions: 1 point
//    - Character insertions/deletions: 1 point
//    - Word order changes: 0.5 points
// 4. Normalize final score to 0.0-1.0 range
//
// Performance characteristics:
// - Time complexity: O(m*n) where m,n are string lengths
// - Space complexity: O(min(m,n))
// - Typical execution time: <1ms for business names
func calculateFuzzySimilarity(name1, name2 string) float64 {
    // Normalize input strings
    normalized1 := normalizeBusinessName(name1)
    normalized2 := normalizeBusinessName(name2)
    
    // Apply common business name abbreviations
    normalized1 = applyAbbreviations(normalized1)
    normalized2 = applyAbbreviations(normalized2)
    
    // Calculate base Levenshtein distance
    distance := levenshteinDistance(normalized1, normalized2)
    
    // Apply word order penalty for multi-word names
    if strings.Contains(normalized1, " ") && strings.Contains(normalized2, " ") {
        distance += calculateWordOrderPenalty(normalized1, normalized2)
    }
    
    // Normalize to 0.0-1.0 range
    maxLen := float64(max(len(normalized1), len(normalized2)))
    if maxLen == 0 {
        return 1.0
    }
    
    return 1.0 - (float64(distance) / maxLen)
}
```

### Business Logic Comments

**Example:**
```go
// assessFinancialRisk evaluates the financial risk of a business based on
// multiple factors including revenue stability, credit history, and
// financial ratios.
//
// Risk factors and weights:
// - Revenue volatility (30%): Higher volatility increases risk
// - Credit score (25%): Lower scores indicate higher risk
// - Debt-to-equity ratio (20%): Higher ratios increase risk
// - Cash flow stability (15%): Unstable cash flow increases risk
// - Industry benchmarks (10%): Compared to industry averages
//
// The function returns a risk score from 0.0 (low risk) to 1.0 (high risk)
// and a list of contributing factors for transparency.
func assessFinancialRisk(business *Business, financialData *FinancialData) (*RiskAssessment, error) {
    // Validate input data
    if business == nil || financialData == nil {
        return nil, ErrInvalidInput
    }
    
    // Calculate revenue volatility score
    // Uses coefficient of variation (std dev / mean) over 3 years
    revenueVolatility := calculateRevenueVolatility(financialData.RevenueHistory)
    
    // Assess credit score impact
    // Normalize credit score to 0-1 range and invert (higher score = lower risk)
    creditScoreRisk := 1.0 - (financialData.CreditScore / 850.0)
    
    // Calculate debt-to-equity risk
    // Higher ratios indicate more leverage and higher risk
    debtEquityRisk := min(financialData.DebtToEquityRatio/2.0, 1.0)
    
    // Evaluate cash flow stability
    // Uses standard deviation of monthly cash flows
    cashFlowRisk := calculateCashFlowRisk(financialData.CashFlowHistory)
    
    // Compare to industry benchmarks
    // Adjust risk based on industry-specific financial patterns
    industryRisk := getIndustryBenchmarkRisk(business.Industry, financialData)
    
    // Calculate weighted risk score
    totalRisk := (revenueVolatility * 0.30) +
                 (creditScoreRisk * 0.25) +
                 (debtEquityRisk * 0.20) +
                 (cashFlowRisk * 0.15) +
                 (industryRisk * 0.10)
    
    // Ensure risk score is within valid range
    totalRisk = max(0.0, min(1.0, totalRisk))
    
    return &RiskAssessment{
        RiskScore: totalRisk,
        Factors: []RiskFactor{
            {Name: "Revenue Volatility", Score: revenueVolatility, Weight: 0.30},
            {Name: "Credit Score", Score: creditScoreRisk, Weight: 0.25},
            {Name: "Debt-to-Equity", Score: debtEquityRisk, Weight: 0.20},
            {Name: "Cash Flow Stability", Score: cashFlowRisk, Weight: 0.15},
            {Name: "Industry Benchmark", Score: industryRisk, Weight: 0.10},
        },
    }, nil
}
```

## Best Practices

### Do's and Don'ts

**Do:**
- Comment on "why" not "what"
- Use clear, concise language
- Document exported functions, types, and packages
- Include examples for complex functionality
- Explain business logic and algorithms
- Document error conditions and edge cases
- Keep comments up to date with code changes

**Don't:**
- Comment obvious code
- Use outdated or incorrect information
- Write comments that duplicate code logic
- Use unclear or ambiguous language
- Forget to update comments when code changes
- Use comments to explain poor code design

### Comment Maintenance

**Regular Review:**
- Review comments during code reviews
- Update comments when functionality changes
- Remove obsolete or incorrect comments
- Ensure examples are current and working

**Automation:**
- Use linters to check comment coverage
- Automate comment generation where possible
- Validate comment formatting and style
- Check for comment-code consistency

## Tools and Automation

### GoDoc Generation

**Generate Documentation:**
```bash
# Generate documentation for all packages
godoc -http=:6060

# Generate documentation for specific package
godoc ./internal/classification

# Generate HTML documentation
godoc -html ./internal/classification > docs/classification.html
```

### Linting Tools

**golangci-lint Configuration:**
```yaml
linters:
  enable:
    - godot        # Check comment formatting
    - gocritic     # Check comment quality
    - misspell     # Check spelling in comments
    - revive       # Check exported function documentation

linters-settings:
  godot:
    # Require period at end of sentences
    check-all: true
  gocritic:
    # Check comment quality
    enabled-tags:
      - diagnostic
      - style
```

### Comment Templates

**Function Template:**
```go
// FunctionName performs a specific operation with clear purpose.
//
// Parameters:
//   - param1: Description of first parameter
//   - param2: Description of second parameter
//
// Returns:
//   - returnType: Description of return value
//   - error: Description of error conditions
//
// Example:
//
//	result, err := FunctionName("example", Options{})
//	if err != nil {
//	    return err
//	}
func FunctionName(param1 string, param2 Options) (returnType, error) {
    // Implementation
}
```

**Type Template:**
```go
// TypeName represents a specific concept or data structure.
// It is used for [specific purpose] and provides [key functionality].
//
// Fields:
//   - Field1: Description of first field
//   - Field2: Description of second field
//
// Example usage:
//
//	instance := TypeName{
//	    Field1: "value1",
//	    Field2: "value2",
//	}
type TypeName struct {
    Field1 string `json:"field1"`
    Field2 string `json:"field2"`
}
```

---

## Implementation Checklist

- [ ] Add package-level comments to all packages
- [ ] Document all exported functions with GoDoc comments
- [ ] Add comments for complex private functions
- [ ] Document all exported types and interfaces
- [ ] Add comments for global variables and constants
- [ ] Include examples for complex functionality
- [ ] Document error conditions and edge cases
- [ ] Review and update existing comments
- [ ] Set up automated comment checking
- [ ] Generate and review GoDoc documentation

---

*Last updated: January 2024*
