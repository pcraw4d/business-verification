package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// TestCase represents a test case for keyword matching
type TestCase struct {
	BusinessName     string   `json:"business_name"`
	BusinessDesc     string   `json:"business_description"`
	Keywords         []string `json:"keywords"`
	ExpectedIndustry string   `json:"expected_industry"`
	ExpectedCategory string   `json:"expected_category"`
	TestType         string   `json:"test_type"` // "exact", "partial", "phrase", "synonym"
}

// TestResult represents the result of a keyword matching test
type TestResult struct {
	TestCase        TestCase `json:"test_case"`
	ActualIndustry  string   `json:"actual_industry"`
	ActualCategory  string   `json:"actual_category"`
	Confidence      float64  `json:"confidence"`
	MatchedKeywords []string `json:"matched_keywords"`
	IsCorrect       bool     `json:"is_correct"`
	Error           string   `json:"error,omitempty"`
}

// TestSuite represents a collection of test cases and results
type TestSuite struct {
	TestCases   []TestCase   `json:"test_cases"`
	TestResults []TestResult `json:"test_results"`
	Summary     TestSummary  `json:"summary"`
	Timestamp   time.Time    `json:"timestamp"`
}

// TestSummary provides overall test statistics
type TestSummary struct {
	TotalTests        int                  `json:"total_tests"`
	PassedTests       int                  `json:"passed_tests"`
	FailedTests       int                  `json:"failed_tests"`
	AccuracyRate      float64              `json:"accuracy_rate"`
	AverageConfidence float64              `json:"average_confidence"`
	ByTestType        map[string]TypeStats `json:"by_test_type"`
}

// TypeStats provides statistics for a specific test type
type TypeStats struct {
	Total    int     `json:"total"`
	Passed   int     `json:"passed"`
	Failed   int     `json:"failed"`
	Accuracy float64 `json:"accuracy"`
}

// ClassificationResult represents the result from the classification system
type ClassificationResult struct {
	Industry   *Industry            `json:"industry"`
	Confidence float64              `json:"confidence"`
	Keywords   []string             `json:"keywords"`
	Patterns   []string             `json:"patterns"`
	Codes      []ClassificationCode `json:"codes"`
	Reasoning  string               `json:"reasoning"`
}

// Industry represents an industry from the database
type Industry struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

// ClassificationCode represents a classification code
type ClassificationCode struct {
	Type  string  `json:"type"`
	Code  string  `json:"code"`
	Score float64 `json:"score"`
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Build database connection string from environment variables
	dbURL := buildDatabaseURL()
	if dbURL == "" {
		log.Fatal("Database configuration is incomplete")
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("🧪 Starting comprehensive keyword matching accuracy tests...")

	// Create test cases
	testCases := createTestCases()

	// Run tests
	testSuite, err := runKeywordMatchingTests(db, testCases)
	if err != nil {
		log.Fatalf("Failed to run keyword matching tests: %v", err)
	}

	// Display results
	displayTestResults(testSuite)

	// Save detailed results
	if err := saveTestResults(testSuite); err != nil {
		log.Printf("Warning: Failed to save test results: %v", err)
	}

	fmt.Println("✅ Keyword matching accuracy tests completed successfully!")
}

// buildDatabaseURL builds a PostgreSQL connection string from environment variables
func buildDatabaseURL() string {
	// Check if DATABASE_URL is provided (Railway format)
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		return dbURL
	}

	// Build from individual environment variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	database := os.Getenv("DB_DATABASE")
	sslMode := os.Getenv("DB_SSL_MODE")

	if host == "" || username == "" || database == "" {
		return ""
	}

	if port == "" {
		port = "5432"
	}
	if sslMode == "" {
		sslMode = "require"
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, username, password, database, sslMode)
}

// createTestCases creates comprehensive test cases for keyword matching
func createTestCases() []TestCase {
	return []TestCase{
		// Technology & Software Development
		{
			BusinessName:     "TechCorp Solutions",
			BusinessDesc:     "Custom software development and cloud solutions",
			Keywords:         []string{"software", "development", "cloud", "custom", "solutions"},
			ExpectedIndustry: "Software Development",
			ExpectedCategory: "emerging",
			TestType:         "exact",
		},
		{
			BusinessName:     "AgileDev Inc",
			BusinessDesc:     "Agile software development and DevOps consulting",
			Keywords:         []string{"agile", "devops", "consulting", "software"},
			ExpectedIndustry: "Software Development",
			ExpectedCategory: "emerging",
			TestType:         "phrase",
		},
		{
			BusinessName:     "CloudNative Systems",
			BusinessDesc:     "Microservices and API development",
			Keywords:         []string{"microservices", "api", "cloud", "native"},
			ExpectedIndustry: "Software Development",
			ExpectedCategory: "emerging",
			TestType:         "synonym",
		},

		// Healthcare
		{
			BusinessName:     "Metro Medical Center",
			BusinessDesc:     "Primary care and family medicine practice",
			Keywords:         []string{"medical", "primary", "care", "family", "medicine"},
			ExpectedIndustry: "Medical Practices",
			ExpectedCategory: "traditional",
			TestType:         "exact",
		},
		{
			BusinessName:     "Wellness Clinic",
			BusinessDesc:     "Mental health counseling and therapy services",
			Keywords:         []string{"mental", "health", "counseling", "therapy"},
			ExpectedIndustry: "Mental Health",
			ExpectedCategory: "traditional",
			TestType:         "exact",
		},
		{
			BusinessName:     "HealthTech Innovations",
			BusinessDesc:     "Healthcare technology and medical software",
			Keywords:         []string{"healthcare", "technology", "medical", "software"},
			ExpectedIndustry: "Healthcare Technology",
			ExpectedCategory: "emerging",
			TestType:         "phrase",
		},

		// Financial Services
		{
			BusinessName:     "First National Bank",
			BusinessDesc:     "Commercial banking and financial services",
			Keywords:         []string{"banking", "commercial", "financial", "services"},
			ExpectedIndustry: "Banking",
			ExpectedCategory: "traditional",
			TestType:         "exact",
		},
		{
			BusinessName:     "Secure Insurance Group",
			BusinessDesc:     "Life and health insurance coverage",
			Keywords:         []string{"insurance", "life", "health", "coverage"},
			ExpectedIndustry: "Insurance",
			ExpectedCategory: "traditional",
			TestType:         "exact",
		},
		{
			BusinessName:     "Investment Advisors LLC",
			BusinessDesc:     "Financial planning and investment management",
			Keywords:         []string{"investment", "financial", "planning", "management"},
			ExpectedIndustry: "Investment Services",
			ExpectedCategory: "traditional",
			TestType:         "synonym",
		},

		// Food & Beverage
		{
			BusinessName:     "Bella Vista Restaurant",
			BusinessDesc:     "Fine dining Italian cuisine",
			Keywords:         []string{"restaurant", "dining", "italian", "cuisine"},
			ExpectedIndustry: "Restaurants",
			ExpectedCategory: "traditional",
			TestType:         "exact",
		},
		{
			BusinessName:     "Craft Brewery Co",
			BusinessDesc:     "Artisanal beer brewing and tasting room",
			Keywords:         []string{"brewery", "beer", "brewing", "craft"},
			ExpectedIndustry: "Breweries",
			ExpectedCategory: "traditional",
			TestType:         "exact",
		},
		{
			BusinessName:     "Coffee Corner",
			BusinessDesc:     "Specialty coffee and light breakfast",
			Keywords:         []string{"coffee", "specialty", "breakfast", "cafe"},
			ExpectedIndustry: "Cafes & Coffee Shops",
			ExpectedCategory: "traditional",
			TestType:         "synonym",
		},

		// Professional Services
		{
			BusinessName:     "Legal Associates",
			BusinessDesc:     "Corporate law and legal consulting",
			Keywords:         []string{"legal", "law", "corporate", "consulting"},
			ExpectedIndustry: "Legal Services",
			ExpectedCategory: "traditional",
			TestType:         "exact",
		},
		{
			BusinessName:     "CPA Partners",
			BusinessDesc:     "Accounting and tax preparation services",
			Keywords:         []string{"accounting", "tax", "preparation", "cpa"},
			ExpectedIndustry: "Financial Services",
			ExpectedCategory: "traditional",
			TestType:         "exact",
		},
		{
			BusinessName:     "Marketing Pro",
			BusinessDesc:     "Digital marketing and advertising agency",
			Keywords:         []string{"marketing", "digital", "advertising", "agency"},
			ExpectedIndustry: "Digital Services",
			ExpectedCategory: "emerging",
			TestType:         "phrase",
		},

		// Manufacturing
		{
			BusinessName:     "Precision Manufacturing",
			BusinessDesc:     "Custom metal fabrication and machining",
			Keywords:         []string{"manufacturing", "metal", "fabrication", "machining"},
			ExpectedIndustry: "Manufacturing",
			ExpectedCategory: "traditional",
			TestType:         "exact",
		},
		{
			BusinessName:     "Auto Parts Inc",
			BusinessDesc:     "Automotive parts manufacturing and distribution",
			Keywords:         []string{"automotive", "parts", "manufacturing", "distribution"},
			ExpectedIndustry: "Consumer Manufacturing",
			ExpectedCategory: "traditional",
			TestType:         "synonym",
		},

		// Retail
		{
			BusinessName:     "Fashion Forward",
			BusinessDesc:     "Women's clothing and accessories retail",
			Keywords:         []string{"fashion", "clothing", "accessories", "retail"},
			ExpectedIndustry: "Retail",
			ExpectedCategory: "traditional",
			TestType:         "exact",
		},
		{
			BusinessName:     "Electronics Plus",
			BusinessDesc:     "Consumer electronics and technology retail",
			Keywords:         []string{"electronics", "consumer", "technology", "retail"},
			ExpectedIndustry: "Technology",
			ExpectedCategory: "traditional",
			TestType:         "phrase",
		},

		// Edge cases and challenging scenarios
		{
			BusinessName:     "Tech Solutions",
			BusinessDesc:     "IT consulting and system integration",
			Keywords:         []string{"tech", "solutions", "it", "consulting"},
			ExpectedIndustry: "Technology Services",
			ExpectedCategory: "emerging",
			TestType:         "partial",
		},
		{
			BusinessName:     "Health & Wellness",
			BusinessDesc:     "Holistic health and wellness services",
			Keywords:         []string{"health", "wellness", "holistic", "services"},
			ExpectedIndustry: "Healthcare Services",
			ExpectedCategory: "traditional",
			TestType:         "partial",
		},
		{
			BusinessName:     "Financial Planning",
			BusinessDesc:     "Personal finance and wealth management",
			Keywords:         []string{"financial", "planning", "personal", "wealth"},
			ExpectedIndustry: "Investment Services",
			ExpectedCategory: "traditional",
			TestType:         "partial",
		},
	}
}

// runKeywordMatchingTests executes all test cases and returns results
func runKeywordMatchingTests(db *sql.DB, testCases []TestCase) (*TestSuite, error) {
	ctx := context.Background()

	testSuite := &TestSuite{
		TestCases:   testCases,
		TestResults: make([]TestResult, 0, len(testCases)),
		Timestamp:   time.Now(),
	}

	for _, testCase := range testCases {
		result, err := runSingleTest(ctx, db, testCase)
		if err != nil {
			result = TestResult{
				TestCase:  testCase,
				Error:     err.Error(),
				IsCorrect: false,
			}
		}
		testSuite.TestResults = append(testSuite.TestResults, result)
	}

	// Calculate summary statistics
	testSuite.Summary = calculateTestSummary(testSuite.TestResults)

	return testSuite, nil
}

// runSingleTest executes a single test case
func runSingleTest(ctx context.Context, db *sql.DB, testCase TestCase) (TestResult, error) {
	// Simulate the keyword classification logic
	result, err := classifyBusinessByKeywords(ctx, db, testCase.Keywords)
	if err != nil {
		return TestResult{}, err
	}

	// Determine if the classification is correct
	isCorrect := false
	if result.Industry != nil {
		isCorrect = strings.EqualFold(result.Industry.Name, testCase.ExpectedIndustry) ||
			strings.EqualFold(result.Industry.Category, testCase.ExpectedCategory)
	}

	return TestResult{
		TestCase:        testCase,
		ActualIndustry:  getIndustryName(result.Industry),
		ActualCategory:  getIndustryCategory(result.Industry),
		Confidence:      result.Confidence,
		MatchedKeywords: result.Keywords,
		IsCorrect:       isCorrect,
	}, nil
}

// classifyBusinessByKeywords simulates the keyword classification logic
func classifyBusinessByKeywords(ctx context.Context, db *sql.DB, keywords []string) (*ClassificationResult, error) {
	if len(keywords) == 0 {
		return &ClassificationResult{
			Industry:   &Industry{Name: "General Business", ID: 26},
			Confidence: 0.50,
			Keywords:   []string{},
			Patterns:   []string{},
			Codes:      []ClassificationCode{},
			Reasoning:  "No keywords provided for classification",
		}, nil
	}

	// Build keyword index for classification
	keywordIndex, err := buildKeywordIndex(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to build keyword index: %w", err)
	}

	// Use the same algorithm as the repository
	industryScores := make(map[int]float64)
	industryMatches := make(map[int][]string)

	for _, inputKeyword := range keywords {
		normalizedKeyword := strings.ToLower(strings.TrimSpace(inputKeyword))
		isPhrase := strings.Contains(normalizedKeyword, " ")
		phraseMultiplier := 1.0

		if isPhrase {
			phraseMultiplier = 1.5
		}

		// Direct lookup
		if matches, exists := keywordIndex[normalizedKeyword]; exists {
			for _, match := range matches {
				weight := match.Weight * phraseMultiplier
				industryScores[match.IndustryID] += weight
				industryMatches[match.IndustryID] = append(industryMatches[match.IndustryID], match.Keyword)
			}
		}

		// Partial matching
		for keyword, matches := range keywordIndex {
			if normalizedKeyword == keyword {
				continue
			}

			if isPhrase && strings.Contains(keyword, " ") {
				if hasPhraseOverlap(normalizedKeyword, keyword) {
					for _, match := range matches {
						partialWeight := match.Weight * 0.8
						industryScores[match.IndustryID] += partialWeight
						industryMatches[match.IndustryID] = append(industryMatches[match.IndustryID], match.Keyword)
					}
				}
			} else if strings.Contains(normalizedKeyword, keyword) || strings.Contains(keyword, normalizedKeyword) {
				for _, match := range matches {
					partialWeight := match.Weight * 0.5
					industryScores[match.IndustryID] += partialWeight
					industryMatches[match.IndustryID] = append(industryMatches[match.IndustryID], match.Keyword)
				}
			}
		}
	}

	// Find best industry
	bestIndustryID := 26 // Default
	bestScore := 0.0
	for industryID, score := range industryScores {
		if score > bestScore {
			bestScore = score
			bestIndustryID = industryID
		}
	}

	// Get industry details
	industry, err := getIndustryByID(ctx, db, bestIndustryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get industry: %w", err)
	}

	// Calculate confidence
	confidence := bestScore / float64(len(keywords))
	if confidence > 1.0 {
		confidence = 1.0
	}

	// Get matched keywords
	var matchedKeywords []string
	if matches, exists := industryMatches[bestIndustryID]; exists {
		matchedKeywords = matches
	}

	return &ClassificationResult{
		Industry:   industry,
		Confidence: confidence,
		Keywords:   matchedKeywords,
		Patterns:   []string{},
		Codes:      []ClassificationCode{},
		Reasoning:  fmt.Sprintf("Classified based on %d keyword matches with score %.2f", len(matchedKeywords), bestScore),
	}, nil
}

// KeywordMatch represents a keyword match with an industry
type KeywordMatch struct {
	IndustryID int
	Keyword    string
	Weight     float64
}

// buildKeywordIndex builds the keyword index for classification
func buildKeywordIndex(ctx context.Context, db *sql.DB) (map[string][]KeywordMatch, error) {
	query := `
		SELECT ik.industry_id, ik.keyword, ik.weight
		FROM industry_keywords ik
		WHERE ik.is_active = true
		ORDER BY ik.industry_id, ik.weight DESC
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	index := make(map[string][]KeywordMatch)
	for rows.Next() {
		var industryID int
		var keyword string
		var weight float64

		err := rows.Scan(&industryID, &keyword, &weight)
		if err != nil {
			return nil, err
		}

		normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
		index[normalizedKeyword] = append(index[normalizedKeyword], KeywordMatch{
			IndustryID: industryID,
			Keyword:    keyword,
			Weight:     weight,
		})
	}

	return index, nil
}

// hasPhraseOverlap checks if two phrases have significant overlap
func hasPhraseOverlap(phrase1, phrase2 string) bool {
	words1 := strings.Fields(phrase1)
	words2 := strings.Fields(phrase2)

	overlap := 0
	for _, word1 := range words1 {
		for _, word2 := range words2 {
			if strings.EqualFold(word1, word2) {
				overlap++
				break
			}
		}
	}

	// Consider overlap if at least 50% of words match
	return float64(overlap)/float64(len(words1)) >= 0.5 || float64(overlap)/float64(len(words2)) >= 0.5
}

// getIndustryByID retrieves an industry by its ID
func getIndustryByID(ctx context.Context, db *sql.DB, id int) (*Industry, error) {
	query := `SELECT id, name, category FROM industries WHERE id = $1`

	var industry Industry
	err := db.QueryRowContext(ctx, query, id).Scan(&industry.ID, &industry.Name, &industry.Category)
	if err != nil {
		return nil, err
	}

	return &industry, nil
}

// getIndustryName safely gets the industry name
func getIndustryName(industry *Industry) string {
	if industry == nil {
		return "Unknown"
	}
	return industry.Name
}

// getIndustryCategory safely gets the industry category
func getIndustryCategory(industry *Industry) string {
	if industry == nil {
		return "Unknown"
	}
	return industry.Category
}

// calculateTestSummary calculates summary statistics for the test results
func calculateTestSummary(results []TestResult) TestSummary {
	summary := TestSummary{
		TotalTests: len(results),
		ByTestType: make(map[string]TypeStats),
	}

	var totalConfidence float64
	validConfidenceCount := 0

	for _, result := range results {
		if result.IsCorrect {
			summary.PassedTests++
		} else {
			summary.FailedTests++
		}

		if result.Confidence > 0 {
			totalConfidence += result.Confidence
			validConfidenceCount++
		}

		// Update type-specific stats
		typeStats := summary.ByTestType[result.TestCase.TestType]
		typeStats.Total++
		if result.IsCorrect {
			typeStats.Passed++
		} else {
			typeStats.Failed++
		}
		summary.ByTestType[result.TestCase.TestType] = typeStats
	}

	// Calculate accuracy rate
	if summary.TotalTests > 0 {
		summary.AccuracyRate = float64(summary.PassedTests) / float64(summary.TotalTests) * 100
	}

	// Calculate average confidence
	if validConfidenceCount > 0 {
		summary.AverageConfidence = totalConfidence / float64(validConfidenceCount)
	}

	// Calculate accuracy for each test type
	for testType, stats := range summary.ByTestType {
		if stats.Total > 0 {
			stats.Accuracy = float64(stats.Passed) / float64(stats.Total) * 100
			summary.ByTestType[testType] = stats
		}
	}

	return summary
}

// displayTestResults displays the test results in a formatted manner
func displayTestResults(testSuite *TestSuite) {
	fmt.Println("\n📊 KEYWORD MATCHING ACCURACY TEST RESULTS")
	fmt.Println("==========================================")

	summary := testSuite.Summary
	fmt.Printf("Total Tests: %d\n", summary.TotalTests)
	fmt.Printf("Passed Tests: %d\n", summary.PassedTests)
	fmt.Printf("Failed Tests: %d\n", summary.FailedTests)
	fmt.Printf("Overall Accuracy: %.1f%%\n", summary.AccuracyRate)
	fmt.Printf("Average Confidence: %.2f\n", summary.AverageConfidence)

	fmt.Println("\n📈 ACCURACY BY TEST TYPE:")
	for testType, stats := range summary.ByTestType {
		fmt.Printf("  %s: %.1f%% (%d/%d)\n",
			strings.Title(testType), stats.Accuracy, stats.Passed, stats.Total)
	}

	fmt.Println("\n✅ PASSED TESTS:")
	passedCount := 0
	for _, result := range testSuite.TestResults {
		if result.IsCorrect {
			passedCount++
			fmt.Printf("  %d. %s -> %s (%.2f confidence)\n",
				passedCount, result.TestCase.BusinessName, result.ActualIndustry, result.Confidence)
		}
	}

	fmt.Println("\n❌ FAILED TESTS:")
	failedCount := 0
	for _, result := range testSuite.TestResults {
		if !result.IsCorrect {
			failedCount++
			expected := result.TestCase.ExpectedIndustry
			actual := result.ActualIndustry
			fmt.Printf("  %d. %s -> Expected: %s, Got: %s (%.2f confidence)\n",
				failedCount, result.TestCase.BusinessName, expected, actual, result.Confidence)
		}
	}

	// Performance recommendations
	fmt.Println("\n🎯 RECOMMENDATIONS:")
	if summary.AccuracyRate < 80 {
		fmt.Println("  ⚠️  Accuracy below 80% - consider enhancing keyword coverage")
	}
	if summary.AverageConfidence < 0.7 {
		fmt.Println("  ⚠️  Low average confidence - review keyword weighting")
	}

	// Type-specific recommendations
	for testType, stats := range summary.ByTestType {
		if stats.Accuracy < 70 {
			fmt.Printf("  ⚠️  %s tests need improvement (%.1f%% accuracy)\n",
				strings.Title(testType), stats.Accuracy)
		}
	}
}

// saveTestResults saves the detailed test results to a JSON file
func saveTestResults(testSuite *TestSuite) error {
	filename := fmt.Sprintf("keyword_matching_test_results_%s.json",
		time.Now().Format("2006-01-02"))

	data, err := json.MarshalIndent(testSuite, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
