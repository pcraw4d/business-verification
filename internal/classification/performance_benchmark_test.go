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

// BenchmarkClassificationPerformance benchmarks classification performance
func BenchmarkClassificationPerformance(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping performance benchmark in short mode")
	}

	// Check for Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		b.Skip("Skipping benchmark: SUPABASE_URL or SUPABASE_ANON_KEY not set")
	}

	// Create database client
	config := &database.SupabaseConfig{
		URL:           supabaseURL,
		APIKey:        supabaseKey,
		ServiceRoleKey: supabaseServiceKey,
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	client, err := database.NewSupabaseClient(config, logger)
	if err != nil {
		b.Fatalf("Failed to create Supabase client: %v", err)
	}
	defer client.Close()

	// Create repository and service
	repo := repository.NewSupabaseKeywordRepository(client, logger)
	service := NewIndustryDetectionService(repo, logger)

	// Test cases for benchmarking
	testCases := []struct {
		name        string
		businessName string
		description  string
		websiteURL   string
	}{
		{
			name:         "Technology company with website",
			businessName: "Microsoft Corporation",
			description:  "Software development and cloud computing services",
			websiteURL:   "https://microsoft.com",
		},
		{
			name:         "Retail company with website",
			businessName: "Amazon",
			description:  "E-commerce and retail services",
			websiteURL:   "https://amazon.com",
		},
		{
			name:         "Healthcare company with website",
			businessName: "Mayo Clinic",
			description:  "Medical center and hospital services",
			websiteURL:   "https://mayoclinic.org",
		},
		{
			name:         "Business with description only",
			businessName: "TechCorp Solutions",
			description:  "Technology consulting and software development",
			websiteURL:   "",
		},
		{
			name:         "Business name only",
			businessName: "Local Restaurant",
			description:  "",
			websiteURL:   "",
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			ctx := context.Background()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				start := time.Now()
				_, err := service.DetectIndustry(ctx, tc.businessName, tc.description, tc.websiteURL)
				duration := time.Since(start)

				if err != nil {
					b.Errorf("Classification failed: %v", err)
					continue
				}

				// Check if duration exceeds 5 seconds
				if duration > 5*time.Second {
					b.Errorf("Classification took %v, exceeds 5s target", duration)
				}

				// Log performance metrics
				b.Logf("Iteration %d: %v", i+1, duration)
			}
		})
	}
}

// TestClassificationPerformanceTarget tests that classification completes within 5 seconds
func TestClassificationPerformanceTarget(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Check for Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping performance test: SUPABASE_URL or SUPABASE_ANON_KEY not set")
	}

	// Create database client
	config := &database.SupabaseConfig{
		URL:           supabaseURL,
		APIKey:        supabaseKey,
		ServiceRoleKey: supabaseServiceKey,
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	client, err := database.NewSupabaseClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create Supabase client: %v", err)
	}
	defer client.Close()

	// Create repository and service
	repo := repository.NewSupabaseKeywordRepository(client, logger)
	service := NewIndustryDetectionService(repo, logger)

	// Performance target: 5 seconds
	performanceTarget := 5 * time.Second

	// Test cases with different complexity levels
	testCases := []struct {
		name         string
		businessName string
		description  string
		websiteURL   string
		maxDuration  time.Duration
	}{
		{
			name:         "Simple classification (name only)",
			businessName: "Local Business",
			description:  "",
			websiteURL:   "",
			maxDuration:  2 * time.Second, // Should be fast
		},
		{
			name:         "Medium classification (name + description)",
			businessName: "TechCorp Solutions",
			description:  "Technology consulting and software development services",
			websiteURL:   "",
			maxDuration:  3 * time.Second,
		},
		{
			name:         "Complex classification (name + description + website)",
			businessName: "Microsoft Corporation",
			description:  "Software development and cloud computing services",
			websiteURL:   "https://microsoft.com",
			maxDuration:  performanceTarget, // 5 seconds
		},
		{
			name:         "Complex classification (Amazon)",
			businessName: "Amazon",
			description:  "E-commerce and retail services",
			websiteURL:   "https://amazon.com",
			maxDuration:  performanceTarget,
		},
		{
			name:         "Complex classification (Mayo Clinic)",
			businessName: "Mayo Clinic",
			description:  "Medical center and hospital services",
			websiteURL:   "https://mayoclinic.org",
			maxDuration:  performanceTarget,
		},
	}

	ctx := context.Background()
	results := make(map[string]time.Duration)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()
			result, err := service.DetectIndustry(ctx, tc.businessName, tc.description, tc.websiteURL)
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("Classification failed: %v", err)
			}

			if result == nil {
				t.Fatal("Classification returned nil result")
			}

			// Verify performance target
			if duration > tc.maxDuration {
				t.Errorf("❌ Performance target exceeded: %v > %v (target: %v)", duration, tc.maxDuration, performanceTarget)
			} else {
				t.Logf("✅ Performance target met: %v <= %v", duration, tc.maxDuration)
			}

			// Verify overall 5-second target
			if duration > performanceTarget {
				t.Errorf("❌ Overall performance target exceeded: %v > %v", duration, performanceTarget)
			}

			// Store result for summary
			results[tc.name] = duration

			// Log detailed metrics
			t.Logf("📊 Performance Metrics:")
			t.Logf("   - Duration: %v", duration)
			t.Logf("   - Target: %v", performanceTarget)
			t.Logf("   - Margin: %v", performanceTarget-duration)
			t.Logf("   - Industry: %s", result.IndustryName)
			t.Logf("   - Confidence: %.2f%%", result.Confidence*100)
			t.Logf("   - Method: %s", result.Method)
			t.Logf("   - Keywords: %d", len(result.Keywords))
		})
	}

	// Summary
	t.Logf("\n📊 Performance Summary:")
	t.Logf("   Target: < %v", performanceTarget)
	for name, duration := range results {
		status := "✅"
		if duration > performanceTarget {
			status = "❌"
		}
		t.Logf("   %s %s: %v", status, name, duration)
	}

	// Calculate average
	var totalDuration time.Duration
	for _, duration := range results {
		totalDuration += duration
	}
	avgDuration := totalDuration / time.Duration(len(results))
	t.Logf("   Average: %v", avgDuration)

	if avgDuration > performanceTarget {
		t.Errorf("❌ Average performance exceeds target: %v > %v", avgDuration, performanceTarget)
	} else {
		t.Logf("✅ Average performance within target: %v <= %v", avgDuration, performanceTarget)
	}
}

// TestClassificationPerformanceConsistency tests that performance is consistent across multiple runs
func TestClassificationPerformanceConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping consistency test in short mode")
	}

	// Check for Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping consistency test: SUPABASE_URL or SUPABASE_ANON_KEY not set")
	}

	// Create database client
	config := &database.SupabaseConfig{
		URL:           supabaseURL,
		APIKey:        supabaseKey,
		ServiceRoleKey: supabaseServiceKey,
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	client, err := database.NewSupabaseClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create Supabase client: %v", err)
	}
	defer client.Close()

	// Create repository and service
	repo := repository.NewSupabaseKeywordRepository(client, logger)
	service := NewIndustryDetectionService(repo, logger)

	// Test the same business multiple times
	businessName := "Microsoft Corporation"
	description := "Software development and cloud computing services"
	websiteURL := "https://microsoft.com"

	ctx := context.Background()
	numRuns := 5
	durations := make([]time.Duration, numRuns)
	performanceTarget := 5 * time.Second

	for i := 0; i < numRuns; i++ {
		start := time.Now()
		result, err := service.DetectIndustry(ctx, businessName, description, websiteURL)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Classification failed on run %d: %v", i+1, err)
		}

		if result == nil {
			t.Fatalf("Classification returned nil result on run %d", i+1)
		}

		durations[i] = duration

		// Verify each run is under target
		if duration > performanceTarget {
			t.Errorf("❌ Run %d exceeded target: %v > %v", i+1, duration, performanceTarget)
		} else {
			t.Logf("✅ Run %d: %v (target: %v)", i+1, duration, performanceTarget)
		}
	}

	// Calculate statistics
	var totalDuration time.Duration
	var minDuration, maxDuration time.Duration = durations[0], durations[0]

	for _, d := range durations {
		totalDuration += d
		if d < minDuration {
			minDuration = d
		}
		if d > maxDuration {
			maxDuration = d
		}
	}

	avgDuration := totalDuration / time.Duration(numRuns)
	variance := time.Duration(0)
	for _, d := range durations {
		diff := d - avgDuration
		if diff < 0 {
			diff = -diff
		}
		variance += diff
	}
	avgVariance := variance / time.Duration(numRuns)

	t.Logf("\n📊 Consistency Metrics:")
	t.Logf("   Runs: %d", numRuns)
	t.Logf("   Min: %v", minDuration)
	t.Logf("   Max: %v", maxDuration)
	t.Logf("   Average: %v", avgDuration)
	t.Logf("   Average Variance: %v", avgVariance)
	t.Logf("   Target: %v", performanceTarget)

	// Verify consistency (max should not be more than 2x min)
	if maxDuration > 2*minDuration {
		t.Errorf("❌ Performance inconsistency detected: max (%v) > 2x min (%v)", maxDuration, minDuration)
	} else {
		t.Logf("✅ Performance is consistent: max (%v) <= 2x min (%v)", maxDuration, minDuration)
	}

	// Verify average is under target
	if avgDuration > performanceTarget {
		t.Errorf("❌ Average performance exceeds target: %v > %v", avgDuration, performanceTarget)
	} else {
		t.Logf("✅ Average performance within target: %v <= %v", avgDuration, performanceTarget)
	}
}

// TestClassificationPerformanceUnderLoad tests performance under concurrent load
func TestClassificationPerformanceUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	// Check for Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping load test: SUPABASE_URL or SUPABASE_ANON_KEY not set")
	}

	// Create database client
	config := &database.SupabaseConfig{
		URL:           supabaseURL,
		APIKey:        supabaseKey,
		ServiceRoleKey: supabaseServiceKey,
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	client, err := database.NewSupabaseClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create Supabase client: %v", err)
	}
	defer client.Close()

	// Create repository and service
	repo := repository.NewSupabaseKeywordRepository(client, logger)
	service := NewIndustryDetectionService(repo, logger)

	// Test cases for concurrent execution
	testCases := []struct {
		businessName string
		description  string
		websiteURL   string
	}{
		{"Microsoft Corporation", "Software development", "https://microsoft.com"},
		{"Amazon", "E-commerce services", "https://amazon.com"},
		{"Mayo Clinic", "Medical services", "https://mayoclinic.org"},
		{"TechCorp Solutions", "Technology consulting", ""},
		{"Local Restaurant", "Food service", ""},
	}

	ctx := context.Background()
	concurrency := 3 // Run 3 concurrent requests
	performanceTarget := 5 * time.Second

	type result struct {
		name     string
		duration time.Duration
		err      error
	}

	results := make(chan result, len(testCases)*concurrency)

	// Run concurrent requests
	for _, tc := range testCases {
		for i := 0; i < concurrency; i++ {
			go func(businessName, description, websiteURL string, index int) {
				start := time.Now()
				_, err := service.DetectIndustry(ctx, businessName, description, websiteURL)
				duration := time.Since(start)

				results <- result{
					name:     businessName,
					duration: duration,
					err:      err,
				}
			}(tc.businessName, tc.description, tc.websiteURL, i)
		}
	}

	// Collect results
	totalRequests := len(testCases) * concurrency
	completed := 0
	durations := make([]time.Duration, 0, totalRequests)
	errors := 0

	for completed < totalRequests {
		select {
		case res := <-results:
			completed++
			if res.err != nil {
				t.Errorf("❌ Request failed for %s: %v", res.name, res.err)
				errors++
			} else {
				durations = append(durations, res.duration)
				if res.duration > performanceTarget {
					t.Errorf("❌ Request for %s exceeded target: %v > %v", res.name, res.duration, performanceTarget)
				} else {
					t.Logf("✅ Request for %s: %v (target: %v)", res.name, res.duration, performanceTarget)
				}
			}
		case <-time.After(30 * time.Second):
			t.Fatalf("Timeout waiting for results")
		}
	}

	// Calculate statistics
	if len(durations) == 0 {
		t.Fatal("No successful requests")
	}

	var totalDuration time.Duration
	var minDuration, maxDuration time.Duration = durations[0], durations[0]

	for _, d := range durations {
		totalDuration += d
		if d < minDuration {
			minDuration = d
		}
		if d > maxDuration {
			maxDuration = d
		}
	}

	avgDuration := totalDuration / time.Duration(len(durations))

	t.Logf("\n📊 Load Test Results:")
	t.Logf("   Total Requests: %d", totalRequests)
	t.Logf("   Successful: %d", len(durations))
	t.Logf("   Errors: %d", errors)
	t.Logf("   Min: %v", minDuration)
	t.Logf("   Max: %v", maxDuration)
	t.Logf("   Average: %v", avgDuration)
	t.Logf("   Target: %v", performanceTarget)

	// Verify average is under target
	if avgDuration > performanceTarget {
		t.Errorf("❌ Average performance under load exceeds target: %v > %v", avgDuration, performanceTarget)
	} else {
		t.Logf("✅ Average performance under load within target: %v <= %v", avgDuration, performanceTarget)
	}

	// Verify max is reasonable (should not exceed 2x target)
	if maxDuration > 2*performanceTarget {
		t.Errorf("❌ Max performance under load exceeds 2x target: %v > %v", maxDuration, 2*performanceTarget)
	} else {
		t.Logf("✅ Max performance under load acceptable: %v <= %v", maxDuration, 2*performanceTarget)
	}
}

