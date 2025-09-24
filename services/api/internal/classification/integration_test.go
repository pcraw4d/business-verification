package classification

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/database"
)

// TestDatabaseIntegration runs integration tests against the actual Supabase database
// These tests require a valid Supabase connection and will be skipped if not available
func TestDatabaseIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Check if we have Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseURL == "" || supabaseKey == "" || supabaseServiceKey == "" {
		t.Skip("skipping integration test: SUPABASE_URL, SUPABASE_ANON_KEY, or SUPABASE_SERVICE_ROLE_KEY not set")
	}

	// Create Supabase client
	config := &database.SupabaseConfig{
		URL:            supabaseURL,
		APIKey:         supabaseKey,
		ServiceRoleKey: supabaseServiceKey,
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	client, err := database.NewSupabaseClient(config, logger)
	if err != nil {
		t.Fatalf("failed to create Supabase client: %v", err)
	}
	defer client.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		t.Skipf("skipping integration test: cannot connect to Supabase: %v", err)
	}

	t.Run("test_industry_retrieval", func(t *testing.T) {
		testIndustryRetrieval(t, client)
	})

	t.Run("test_keyword_search", func(t *testing.T) {
		testKeywordSearch(t, client)
	})

	t.Run("test_classification_codes", func(t *testing.T) {
		testClassificationCodes(t, client)
	})

	t.Run("test_end_to_end_classification", func(t *testing.T) {
		testEndToEndClassification(t, client)
	})
}

func testIndustryRetrieval(t *testing.T, client *database.SupabaseClient) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := repository.NewRepository(client, logger)
	ctx := context.Background()

	// Test getting industries
	industries, err := repo.ListIndustries(ctx, "")
	if err != nil {
		t.Fatalf("failed to list industries: %v", err)
	}

	if len(industries) == 0 {
		t.Fatal("expected at least one industry, got none")
	}

	t.Logf("found %d industries", len(industries))

	// Test getting specific industry
	firstIndustry := industries[0]
	industry, err := repo.GetIndustryByID(ctx, firstIndustry.ID)
	if err != nil {
		t.Fatalf("failed to get industry by ID: %v", err)
	}

	if industry == nil {
		t.Fatal("expected industry, got nil")
	}

	if industry.ID != firstIndustry.ID {
		t.Errorf("expected industry ID %d, got %d", firstIndustry.ID, industry.ID)
	}

	t.Logf("retrieved industry: %s", industry.Name)
}

func testKeywordSearch(t *testing.T, client *database.SupabaseClient) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := repository.NewRepository(client, logger)
	ctx := context.Background()

	// Test keyword search
	keywords, err := repo.SearchKeywords(ctx, "technology", 10)
	if err != nil {
		t.Fatalf("failed to search keywords: %v", err)
	}

	if len(keywords) == 0 {
		t.Log("no technology keywords found (this might be expected)")
		return
	}

	t.Logf("found %d technology keywords", len(keywords))

	// Test getting keywords by industry
	if len(keywords) > 0 {
		firstKeyword := keywords[0]
		industryKeywords, err := repo.GetKeywordsByIndustry(ctx, firstKeyword.IndustryID)
		if err != nil {
			t.Fatalf("failed to get keywords by industry: %v", err)
		}

		t.Logf("found %d keywords for industry %d", len(industryKeywords), firstKeyword.IndustryID)
	}
}

func testClassificationCodes(t *testing.T, client *database.SupabaseClient) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := repository.NewRepository(client, logger)
	ctx := context.Background()

	// Test getting classification codes
	codes, err := repo.GetClassificationCodesByIndustry(ctx, 1) // Assuming industry ID 1 exists
	if err != nil {
		t.Fatalf("failed to get classification codes: %v", err)
	}

	if len(codes) == 0 {
		t.Log("no classification codes found for industry 1 (this might be expected)")
		return
	}

	t.Logf("found %d classification codes for industry 1", len(codes))

	// Test getting top industries by keywords
	topIndustries, err := repo.GetTopIndustriesByKeywords(ctx, []string{"technology", "software"}, 5)
	if err != nil {
		t.Fatalf("failed to get top industries by keywords: %v", err)
	}

	if len(topIndustries) > 0 {
		t.Logf("top industry for 'technology,software': %s", topIndustries[0].Name)
	}
}

func testEndToEndClassification(t *testing.T, client *database.SupabaseClient) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := repository.NewRepository(client, logger)
	ctx := context.Background()

	// Test business classification
	result, err := repo.ClassifyBusiness(ctx, "TechCorp Solutions",
		"We develop innovative software solutions for businesses using cloud technology",
		"https://techcorp.com")

	if err != nil {
		t.Fatalf("failed to classify business: %v", err)
	}

	if result == nil {
		t.Fatal("expected classification result, got nil")
	}

	t.Logf("business classified as: %s (confidence: %.2f%%)",
		result.Industry.Name, result.Confidence)

	// Test that we have classification codes
	if len(result.Codes) > 0 {
		t.Logf("found %d classification codes", len(result.Codes))

		// Log some codes
		for i, code := range result.Codes {
			if i >= 3 { // Limit to first 3
				break
			}
			t.Logf("  %s: %s (%s)", code.CodeType, code.Code, code.Description)
		}
	} else {
		t.Log("no classification codes found (this might be expected)")
	}
}

// TestDatabasePerformance runs performance tests against the database
func TestDatabasePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	// Check if we have Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("skipping performance test: SUPABASE_URL or SUPABASE_ANON_KEY not set")
	}

	// Create Supabase client
	config := &database.SupabaseConfig{
		URL:    supabaseURL,
		APIKey: supabaseKey,
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	client, err := database.NewSupabaseClient(config, logger)
	if err != nil {
		t.Fatalf("failed to create Supabase client: %v", err)
	}
	defer client.Close()

	repo := repository.NewRepository(client, logger)
	ctx := context.Background()

	t.Run("test_query_performance", func(t *testing.T) {
		testQueryPerformance(t, repo, ctx)
	})

	t.Run("test_concurrent_requests", func(t *testing.T) {
		testConcurrentRequests(t, repo, ctx)
	})
}

func testQueryPerformance(t *testing.T, repo repository.KeywordRepository, ctx context.Context) {
	// Test industry listing performance
	start := time.Now()
	_, err := repo.ListIndustries(ctx, "")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("failed to list industries: %v", err)
	}

	t.Logf("list industries (100): %v", duration)

	// Performance threshold: should complete within 1 second
	if duration > time.Second {
		t.Errorf("industry listing took too long: %v", duration)
	}

	// Test keyword search performance
	start = time.Now()
	_, err = repo.SearchKeywords(ctx, "business", 50)
	duration = time.Since(start)

	if err != nil {
		t.Fatalf("failed to search keywords: %v", err)
	}

	t.Logf("keyword search (50 results): %v", duration)

	// Performance threshold: should complete within 500ms
	if duration > 500*time.Millisecond {
		t.Errorf("keyword search took too long: %v", duration)
	}
}

func testConcurrentRequests(t *testing.T, repo repository.KeywordRepository, ctx context.Context) {
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	// Start concurrent requests
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			// Simulate different search queries
			query := "business"
			if id%2 == 0 {
				query = "technology"
			}

			_, err := repo.SearchKeywords(ctx, query, 10)
			results <- err
		}(i)
	}

	// Collect results
	var errors []error
	for i := 0; i < numGoroutines; i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}

	// Check for errors
	if len(errors) > 0 {
		t.Errorf("encountered %d errors in concurrent requests", len(errors))
		for _, err := range errors {
			t.Logf("concurrent request error: %v", err)
		}
	}

	t.Logf("successfully completed %d concurrent requests", numGoroutines)
}
