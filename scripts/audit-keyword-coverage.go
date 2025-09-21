package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// KeywordAuditResult represents the audit results for keyword coverage
type KeywordAuditResult struct {
	TotalIndustries       int                `json:"total_industries"`
	TotalKeywords         int                `json:"total_keywords"`
	IndustriesWithData    int                `json:"industries_with_data"`
	IndustriesWithoutData int                `json:"industries_without_data"`
	IndustryCoverage      []IndustryCoverage `json:"industry_coverage"`
	MissingKeywords       []MissingKeyword   `json:"missing_keywords"`
	KeywordGaps           []KeywordGap       `json:"keyword_gaps"`
	Recommendations       []string           `json:"recommendations"`
	AuditTimestamp        time.Time          `json:"audit_timestamp"`
}

// IndustryCoverage represents keyword coverage for a specific industry
type IndustryCoverage struct {
	IndustryID      int      `json:"industry_id"`
	IndustryName    string   `json:"industry_name"`
	Category        string   `json:"category"`
	KeywordCount    int      `json:"keyword_count"`
	PrimaryKeywords int      `json:"primary_keywords"`
	AvgWeight       float64  `json:"avg_weight"`
	Keywords        []string `json:"keywords"`
	CoverageScore   float64  `json:"coverage_score"`
}

// MissingKeyword represents a keyword that should be added
type MissingKeyword struct {
	IndustryID   int     `json:"industry_id"`
	IndustryName string  `json:"industry_name"`
	Keyword      string  `json:"keyword"`
	Weight       float64 `json:"weight"`
	Reason       string  `json:"reason"`
	Priority     string  `json:"priority"`
}

// KeywordGap represents a gap in keyword coverage
type KeywordGap struct {
	IndustryID   int      `json:"industry_id"`
	IndustryName string   `json:"industry_name"`
	GapType      string   `json:"gap_type"`
	Description  string   `json:"description"`
	Suggestions  []string `json:"suggestions"`
	Priority     string   `json:"priority"`
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

	fmt.Println("🔍 Starting comprehensive keyword coverage audit...")

	// Perform audit
	auditResult, err := performKeywordAudit(db)
	if err != nil {
		log.Fatalf("Failed to perform keyword audit: %v", err)
	}

	// Generate recommendations
	auditResult.Recommendations = generateRecommendations(auditResult)

	// Save results to file
	outputFile := "keyword_coverage_audit_" + time.Now().Format("2006-01-02") + ".json"
	if err := saveAuditResults(auditResult, outputFile); err != nil {
		log.Fatalf("Failed to save audit results: %v", err)
	}

	// Print summary
	printAuditSummary(auditResult)

	fmt.Printf("✅ Keyword coverage audit completed successfully!\n")
	fmt.Printf("📄 Detailed results saved to: %s\n", outputFile)
}

// performKeywordAudit performs a comprehensive audit of keyword coverage
func performKeywordAudit(db *sql.DB) (*KeywordAuditResult, error) {
	ctx := context.Background()

	result := &KeywordAuditResult{
		AuditTimestamp: time.Now(),
	}

	// Get all industries
	industries, err := getAllIndustries(db, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get industries: %w", err)
	}
	result.TotalIndustries = len(industries)

	// Analyze each industry
	var totalKeywords int
	var industriesWithData int

	for _, industry := range industries {
		coverage, err := analyzeIndustryCoverage(db, ctx, industry)
		if err != nil {
			log.Printf("Warning: Failed to analyze industry %s: %v", industry.Name, err)
			continue
		}

		result.IndustryCoverage = append(result.IndustryCoverage, coverage)
		totalKeywords += coverage.KeywordCount

		if coverage.KeywordCount > 0 {
			industriesWithData++
		}
	}

	result.TotalKeywords = totalKeywords
	result.IndustriesWithData = industriesWithData
	result.IndustriesWithoutData = result.TotalIndustries - industriesWithData

	// Identify missing keywords and gaps
	result.MissingKeywords = identifyMissingKeywords(result.IndustryCoverage)
	result.KeywordGaps = identifyKeywordGaps(result.IndustryCoverage)

	return result, nil
}

// Industry represents an industry from the database
type Industry struct {
	ID                  int
	Name                string
	Description         string
	Category            string
	ParentIndustryID    *int
	ConfidenceThreshold float64
	IsActive            bool
}

// getAllIndustries retrieves all industries from the database
func getAllIndustries(db *sql.DB, ctx context.Context) ([]Industry, error) {
	query := `
		SELECT id, name, description, category, parent_industry_id, 
		       confidence_threshold, is_active
		FROM industries 
		WHERE is_active = true
		ORDER BY name
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var industries []Industry
	for rows.Next() {
		var industry Industry
		err := rows.Scan(
			&industry.ID, &industry.Name, &industry.Description,
			&industry.Category, &industry.ParentIndustryID,
			&industry.ConfidenceThreshold, &industry.IsActive,
		)
		if err != nil {
			return nil, err
		}
		industries = append(industries, industry)
	}

	return industries, nil
}

// analyzeIndustryCoverage analyzes keyword coverage for a specific industry
func analyzeIndustryCoverage(db *sql.DB, ctx context.Context, industry Industry) (IndustryCoverage, error) {
	coverage := IndustryCoverage{
		IndustryID:   industry.ID,
		IndustryName: industry.Name,
		Category:     industry.Category,
	}

	// Get keywords for this industry
	query := `
		SELECT keyword, weight
		FROM industry_keywords 
		WHERE industry_id = $1 AND is_active = true
		ORDER BY weight DESC, keyword
	`

	rows, err := db.QueryContext(ctx, query, industry.ID)
	if err != nil {
		return coverage, err
	}
	defer rows.Close()

	var totalWeight float64
	var primaryCount int
	var keywords []string

	for rows.Next() {
		var keyword string
		var weight float64

		err := rows.Scan(&keyword, &weight)
		if err != nil {
			return coverage, err
		}

		keywords = append(keywords, keyword)
		totalWeight += weight

		// Consider keywords with weight >= 0.9 as primary
		if weight >= 0.9 {
			primaryCount++
		}
	}

	coverage.KeywordCount = len(keywords)
	coverage.PrimaryKeywords = primaryCount
	coverage.Keywords = keywords

	if coverage.KeywordCount > 0 {
		coverage.AvgWeight = totalWeight / float64(coverage.KeywordCount)
	}

	// Calculate coverage score based on keyword count, primary keywords, and average weight
	coverage.CoverageScore = calculateCoverageScore(coverage)

	return coverage, nil
}

// calculateCoverageScore calculates a coverage score for an industry
func calculateCoverageScore(coverage IndustryCoverage) float64 {
	// Base score from keyword count (0-40 points)
	keywordScore := float64(coverage.KeywordCount) * 0.4
	if keywordScore > 40 {
		keywordScore = 40
	}

	// Primary keyword bonus (0-30 points)
	primaryScore := float64(coverage.PrimaryKeywords) * 3.0
	if primaryScore > 30 {
		primaryScore = 30
	}

	// Weight quality score (0-30 points)
	weightScore := coverage.AvgWeight * 30

	totalScore := keywordScore + primaryScore + weightScore
	if totalScore > 100 {
		totalScore = 100
	}

	return totalScore
}

// identifyMissingKeywords identifies keywords that should be added
func identifyMissingKeywords(coverages []IndustryCoverage) []MissingKeyword {
	var missing []MissingKeyword

	// Define expected keywords for each industry category
	expectedKeywords := map[string][]string{
		"Technology": {
			"software", "technology", "digital", "tech", "app", "platform", "development",
			"programming", "coding", "computer", "IT", "cyber", "data", "analytics",
			"cloud", "AI", "machine learning", "artificial intelligence", "blockchain",
			"cryptocurrency", "fintech", "edtech", "healthtech", "proptech",
		},
		"Healthcare": {
			"medical", "healthcare", "health", "medical practice", "clinic", "hospital",
			"physician", "doctor", "patient", "medical treatment", "diagnosis", "surgery",
			"pharmaceutical", "pharma", "medication", "drug", "medical device", "medtech",
			"telemedicine", "mental health", "psychology", "therapy", "counseling",
		},
		"Financial Services": {
			"bank", "banking", "finance", "financial", "credit", "loan", "investment",
			"insurance", "accounting", "bookkeeping", "tax", "audit", "financial planning",
			"wealth management", "asset management", "trading", "securities", "fintech",
		},
		"Retail": {
			"retail", "store", "shop", "merchandise", "products", "commerce", "ecommerce",
			"online store", "shopping", "fashion", "apparel", "clothing", "electronics",
			"consumer goods", "brand", "franchise", "chain store",
		},
		"Manufacturing": {
			"manufacturing", "production", "factory", "industrial", "machinery", "assembly",
			"automotive", "electronics", "textile", "chemical", "aerospace", "components",
			"supply chain", "quality control", "engineering",
		},
		"Professional Services": {
			"legal", "law", "attorney", "lawyer", "consulting", "advisory", "marketing",
			"advertising", "real estate", "property", "accounting", "audit", "tax",
			"business services", "professional services",
		},
		"Education": {
			"education", "school", "university", "college", "learning", "teaching", "training",
			"academic", "student", "curriculum", "online education", "e-learning", "edtech",
		},
		"Transportation": {
			"transportation", "logistics", "shipping", "freight", "delivery", "courier",
			"warehousing", "supply chain", "trucking", "aviation", "maritime",
		},
		"Food & Beverage": {
			"restaurant", "food", "dining", "catering", "beverage", "food service",
			"food delivery", "food manufacturing", "food processing", "culinary",
		},
		"Construction": {
			"construction", "building", "contractor", "engineering", "architecture",
			"renovation", "home improvement", "infrastructure", "project management",
		},
		"Agriculture": {
			"agriculture", "farming", "crop", "livestock", "food production", "agricultural",
			"farming equipment", "agricultural services", "crop production",
		},
	}

	for _, coverage := range coverages {
		// Find expected keywords for this industry
		expected, exists := expectedKeywords[coverage.IndustryName]
		if !exists {
			// Try to match by category or partial name
			for category, keywords := range expectedKeywords {
				if strings.Contains(strings.ToLower(coverage.IndustryName), strings.ToLower(category)) ||
					strings.Contains(strings.ToLower(category), strings.ToLower(coverage.IndustryName)) {
					expected = keywords
					break
				}
			}
		}

		if expected != nil {
			// Check which expected keywords are missing
			existingKeywords := make(map[string]bool)
			for _, keyword := range coverage.Keywords {
				existingKeywords[strings.ToLower(keyword)] = true
			}

			for _, expectedKeyword := range expected {
				if !existingKeywords[strings.ToLower(expectedKeyword)] {
					priority := "medium"
					weight := 0.7

					// High priority for core industry terms
					coreTerms := []string{"software", "medical", "bank", "retail", "manufacturing", "legal", "education", "transportation", "restaurant", "construction", "agriculture"}
					for _, coreTerm := range coreTerms {
						if strings.Contains(strings.ToLower(expectedKeyword), coreTerm) {
							priority = "high"
							weight = 0.9
							break
						}
					}

					missing = append(missing, MissingKeyword{
						IndustryID:   coverage.IndustryID,
						IndustryName: coverage.IndustryName,
						Keyword:      expectedKeyword,
						Weight:       weight,
						Reason:       "Missing core industry keyword",
						Priority:     priority,
					})
				}
			}
		}

		// Check for low coverage industries
		if coverage.CoverageScore < 50 {
			missing = append(missing, MissingKeyword{
				IndustryID:   coverage.IndustryID,
				IndustryName: coverage.IndustryName,
				Keyword:      "comprehensive keyword coverage needed",
				Weight:       0.8,
				Reason:       fmt.Sprintf("Low coverage score: %.1f", coverage.CoverageScore),
				Priority:     "high",
			})
		}
	}

	return missing
}

// identifyKeywordGaps identifies gaps in keyword coverage
func identifyKeywordGaps(coverages []IndustryCoverage) []KeywordGap {
	var gaps []KeywordGap

	for _, coverage := range coverages {
		// Check for insufficient primary keywords
		if coverage.PrimaryKeywords < 3 {
			gaps = append(gaps, KeywordGap{
				IndustryID:   coverage.IndustryID,
				IndustryName: coverage.IndustryName,
				GapType:      "insufficient_primary_keywords",
				Description:  fmt.Sprintf("Only %d primary keywords (minimum 3 recommended)", coverage.PrimaryKeywords),
				Suggestions:  []string{"Add more primary keywords with weight 0.9-1.0", "Focus on core industry terms", "Include business name variations"},
				Priority:     "high",
			})
		}

		// Check for low keyword count
		if coverage.KeywordCount < 10 {
			gaps = append(gaps, KeywordGap{
				IndustryID:   coverage.IndustryID,
				IndustryName: coverage.IndustryName,
				GapType:      "insufficient_keywords",
				Description:  fmt.Sprintf("Only %d keywords (minimum 10 recommended)", coverage.KeywordCount),
				Suggestions:  []string{"Add synonyms and variations", "Include technical terms", "Add business model keywords", "Include service/product keywords"},
				Priority:     "high",
			})
		}

		// Check for low average weight
		if coverage.AvgWeight < 0.7 {
			gaps = append(gaps, KeywordGap{
				IndustryID:   coverage.IndustryID,
				IndustryName: coverage.IndustryName,
				GapType:      "low_keyword_weights",
				Description:  fmt.Sprintf("Average weight %.2f (minimum 0.7 recommended)", coverage.AvgWeight),
				Suggestions:  []string{"Increase weights for core keywords", "Review and adjust keyword relevance", "Focus on high-impact terms"},
				Priority:     "medium",
			})
		}

		// Check for missing context diversity
		if coverage.KeywordCount > 0 {
			// Simple check for keyword diversity
			hasBusinessTerms := false
			hasTechnicalTerms := false

			for _, keyword := range coverage.Keywords {
				keywordLower := strings.ToLower(keyword)
				if strings.Contains(keywordLower, "business") || strings.Contains(keywordLower, "service") || strings.Contains(keywordLower, "company") {
					hasBusinessTerms = true
				}
				if strings.Contains(keywordLower, "tech") || strings.Contains(keywordLower, "digital") || strings.Contains(keywordLower, "system") {
					hasTechnicalTerms = true
				}
			}

			if !hasBusinessTerms || !hasTechnicalTerms {
				gaps = append(gaps, KeywordGap{
					IndustryID:   coverage.IndustryID,
					IndustryName: coverage.IndustryName,
					GapType:      "limited_context_diversity",
					Description:  "Keywords lack diversity in business and technical contexts",
					Suggestions:  []string{"Add business model keywords", "Include technical terminology", "Add industry-specific jargon", "Include service/product variations"},
					Priority:     "medium",
				})
			}
		}
	}

	return gaps
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

// generateRecommendations generates actionable recommendations
func generateRecommendations(result *KeywordAuditResult) []string {
	var recommendations []string

	// Overall coverage recommendations
	if result.IndustriesWithoutData > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Add keywords for %d industries that currently have no keyword data", result.IndustriesWithoutData))
	}

	// High priority missing keywords
	highPriorityMissing := 0
	for _, missing := range result.MissingKeywords {
		if missing.Priority == "high" {
			highPriorityMissing++
		}
	}
	if highPriorityMissing > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Add %d high-priority missing keywords to improve classification accuracy", highPriorityMissing))
	}

	// Low coverage industries
	lowCoverageIndustries := 0
	for _, coverage := range result.IndustryCoverage {
		if coverage.CoverageScore < 50 {
			lowCoverageIndustries++
		}
	}
	if lowCoverageIndustries > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Enhance keyword coverage for %d industries with low coverage scores", lowCoverageIndustries))
	}

	// Keyword gap recommendations
	highPriorityGaps := 0
	for _, gap := range result.KeywordGaps {
		if gap.Priority == "high" {
			highPriorityGaps++
		}
	}
	if highPriorityGaps > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Address %d high-priority keyword gaps to improve system performance", highPriorityGaps))
	}

	// General recommendations
	recommendations = append(recommendations,
		"Implement keyword weighting system with primary/secondary classification",
		"Add synonyms and variations for existing keywords",
		"Create keyword validation and testing framework",
		"Establish keyword maintenance and update procedures")

	return recommendations
}

// saveAuditResults saves audit results to a JSON file
func saveAuditResults(result *KeywordAuditResult, filename string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// printAuditSummary prints a summary of the audit results
func printAuditSummary(result *KeywordAuditResult) {
	fmt.Println("\n📊 KEYWORD COVERAGE AUDIT SUMMARY")
	fmt.Println("==================================")
	fmt.Printf("Total Industries: %d\n", result.TotalIndustries)
	fmt.Printf("Industries with Keywords: %d\n", result.IndustriesWithData)
	fmt.Printf("Industries without Keywords: %d\n", result.IndustriesWithoutData)
	fmt.Printf("Total Keywords: %d\n", result.TotalKeywords)
	fmt.Printf("Missing Keywords: %d\n", len(result.MissingKeywords))
	fmt.Printf("Keyword Gaps: %d\n", len(result.KeywordGaps))

	// Show top 5 industries by coverage score
	fmt.Println("\n🏆 TOP 5 INDUSTRIES BY COVERAGE SCORE:")
	sort.Slice(result.IndustryCoverage, func(i, j int) bool {
		return result.IndustryCoverage[i].CoverageScore > result.IndustryCoverage[j].CoverageScore
	})

	for i, coverage := range result.IndustryCoverage[:min(5, len(result.IndustryCoverage))] {
		fmt.Printf("%d. %s: %.1f%% (%d keywords, %d primary)\n",
			i+1, coverage.IndustryName, coverage.CoverageScore,
			coverage.KeywordCount, coverage.PrimaryKeywords)
	}

	// Show bottom 5 industries by coverage score
	fmt.Println("\n⚠️  BOTTOM 5 INDUSTRIES BY COVERAGE SCORE:")
	for i, coverage := range result.IndustryCoverage[max(0, len(result.IndustryCoverage)-5):] {
		fmt.Printf("%d. %s: %.1f%% (%d keywords, %d primary)\n",
			len(result.IndustryCoverage)-4+i, coverage.IndustryName, coverage.CoverageScore,
			coverage.KeywordCount, coverage.PrimaryKeywords)
	}

	// Show high priority recommendations
	fmt.Println("\n🎯 HIGH PRIORITY RECOMMENDATIONS:")
	for i, rec := range result.Recommendations[:min(5, len(result.Recommendations))] {
		fmt.Printf("%d. %s\n", i+1, rec)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
