package webanalysis

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// createTestWebScraper creates a properly initialized WebScraper for testing
func createTestWebScraper() *WebScraper {
	proxyMgr := NewProxyManager()
	return NewWebScraper(proxyMgr)
}

func TestNewPriorityScrapingQueue(t *testing.T) {
	scraper := createTestWebScraper()
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)

	if queue == nil {
		t.Fatal("Expected non-nil PriorityScrapingQueue instance")
	}

	if queue.scraper != scraper {
		t.Error("Expected scraper to be set correctly")
	}

	if queue.relevanceScorer != relevanceScorer {
		t.Error("Expected relevanceScorer to be set correctly")
	}

	if queue.pageDiscovery != pageDiscovery {
		t.Error("Expected pageDiscovery to be set correctly")
	}

	if len(queue.workers) != queue.config.MaxWorkers {
		t.Errorf("Expected %d workers, got %d", queue.config.MaxWorkers, len(queue.workers))
	}

	if queue.queue == nil {
		t.Error("Expected queue to be initialized")
	}

	if queue.results == nil {
		t.Error("Expected results channel to be initialized")
	}
}

func TestAddJob(t *testing.T) {
	scraper := createTestWebScraper()
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)
	defer queue.Stop()

	job := &PriorityScrapingJob{
		URL:      "https://example.com/about",
		Business: "Test Company",
		Priority: 100,
		Context:  &ScrapingContext{},
		Metadata: make(map[string]interface{}),
	}

	err := queue.AddJob(job)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if queue.GetQueueSize() != 1 {
		t.Errorf("Expected queue size 1, got %d", queue.GetQueueSize())
	}
}

func TestAddJobs(t *testing.T) {
	scraper := createTestWebScraper()
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)
	defer queue.Stop()

	jobs := []*PriorityScrapingJob{
		{
			URL:      "https://example.com/about",
			Business: "Test Company",
			Priority: 100,
			Context:  &ScrapingContext{},
			Metadata: make(map[string]interface{}),
		},
		{
			URL:      "https://example.com/contact",
			Business: "Test Company",
			Priority: 90,
			Context:  &ScrapingContext{},
			Metadata: make(map[string]interface{}),
		},
	}

	err := queue.AddJobs(jobs)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if queue.GetQueueSize() != 2 {
		t.Errorf("Expected queue size 2, got %d", queue.GetQueueSize())
	}
}

func TestCreateJobFromDiscoveryResult(t *testing.T) {
	scraper := createTestWebScraper()
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)

	discoveryResult := &PageDiscoveryResult{
		URL:               "https://example.com/about",
		RelevanceScore:    0.8,
		PageType:          PageTypeAbout,
		Priority:          150,
		Depth:             0,
		DiscoveredAt:      time.Now(),
		ContentIndicators: []string{"company_information"},
		BusinessKeywords:  []string{"test", "company"},
	}

	context := &ScrapingContext{
		Industry: "technology",
		Location: "United States",
	}

	job := queue.CreateJobFromDiscoveryResult(discoveryResult, "Test Company", context)

	if job == nil {
		t.Fatal("Expected non-nil job")
	}

	if job.URL != discoveryResult.URL {
		t.Errorf("Expected URL %s, got %s", discoveryResult.URL, job.URL)
	}

	if job.Business != "Test Company" {
		t.Errorf("Expected business Test Company, got %s", job.Business)
	}

	if job.Priority != discoveryResult.Priority {
		t.Errorf("Expected priority %d, got %d", discoveryResult.Priority, job.Priority)
	}

	if job.RelevanceScore != discoveryResult.RelevanceScore {
		t.Errorf("Expected relevance score %f, got %f", discoveryResult.RelevanceScore, job.RelevanceScore)
	}

	if job.PageType != discoveryResult.PageType {
		t.Errorf("Expected page type %s, got %s", discoveryResult.PageType, job.PageType)
	}

	if job.Depth != discoveryResult.Depth {
		t.Errorf("Expected depth %d, got %d", discoveryResult.Depth, job.Depth)
	}

	if job.Context != context {
		t.Error("Expected context to be set correctly")
	}
}

func TestCreateJobFromURL(t *testing.T) {
	scraper := &WebScraper{}
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)

	context := &ScrapingContext{
		Industry: "technology",
		Location: "United States",
	}

	job := queue.CreateJobFromURL("https://example.com/about", "Test Company", 100, context)

	if job == nil {
		t.Fatal("Expected non-nil job")
	}

	if job.URL != "https://example.com/about" {
		t.Errorf("Expected URL https://example.com/about, got %s", job.URL)
	}

	if job.Business != "Test Company" {
		t.Errorf("Expected business Test Company, got %s", job.Business)
	}

	if job.Priority != 100 {
		t.Errorf("Expected priority 100, got %d", job.Priority)
	}

	if job.Context != context {
		t.Error("Expected context to be set correctly")
	}
}

func TestGetQueueSize(t *testing.T) {
	scraper := &WebScraper{}
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)
	defer queue.Stop()

	if queue.GetQueueSize() != 0 {
		t.Errorf("Expected initial queue size 0, got %d", queue.GetQueueSize())
	}

	job := &PriorityScrapingJob{
		URL:      "https://example.com/about",
		Business: "Test Company",
		Priority: 100,
		Context:  &ScrapingContext{},
		Metadata: make(map[string]interface{}),
	}

	queue.AddJob(job)

	if queue.GetQueueSize() != 1 {
		t.Errorf("Expected queue size 1, got %d", queue.GetQueueSize())
	}
}

func TestIsEmpty(t *testing.T) {
	scraper := &WebScraper{}
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)
	defer queue.Stop()

	if !queue.IsEmpty() {
		t.Error("Expected queue to be empty initially")
	}

	job := &PriorityScrapingJob{
		URL:      "https://example.com/about",
		Business: "Test Company",
		Priority: 100,
		Context:  &ScrapingContext{},
		Metadata: make(map[string]interface{}),
	}

	queue.AddJob(job)

	if queue.IsEmpty() {
		t.Error("Expected queue to not be empty after adding job")
	}
}

func TestClear(t *testing.T) {
	scraper := &WebScraper{}
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)
	defer queue.Stop()

	// Add some jobs
	jobs := []*PriorityScrapingJob{
		{
			URL:      "https://example.com/about",
			Business: "Test Company",
			Priority: 100,
			Context:  &ScrapingContext{},
			Metadata: make(map[string]interface{}),
		},
		{
			URL:      "https://example.com/contact",
			Business: "Test Company",
			Priority: 90,
			Context:  &ScrapingContext{},
			Metadata: make(map[string]interface{}),
		},
	}

	queue.AddJobs(jobs)

	if queue.GetQueueSize() != 2 {
		t.Errorf("Expected queue size 2, got %d", queue.GetQueueSize())
	}

	queue.Clear()

	if queue.GetQueueSize() != 0 {
		t.Errorf("Expected queue size 0 after clear, got %d", queue.GetQueueSize())
	}

	if !queue.IsEmpty() {
		t.Error("Expected queue to be empty after clear")
	}
}

func TestGetActiveWorkerCount(t *testing.T) {
	scraper := &WebScraper{}
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)
	defer queue.Stop()

	activeWorkers := queue.GetActiveWorkerCount()
	if activeWorkers != queue.config.MaxWorkers {
		t.Errorf("Expected %d active workers, got %d", queue.config.MaxWorkers, activeWorkers)
	}
}

func TestGetStats(t *testing.T) {
	scraper := &WebScraper{}
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)
	defer queue.Stop()

	stats := queue.GetStats()

	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	if stats.TotalJobsQueued != 0 {
		t.Errorf("Expected 0 jobs queued initially, got %d", stats.TotalJobsQueued)
	}

	if stats.QueueSize != 0 {
		t.Errorf("Expected queue size 0, got %d", stats.QueueSize)
	}

	if stats.ActiveWorkers != queue.config.MaxWorkers {
		t.Errorf("Expected %d active workers, got %d", queue.config.MaxWorkers, stats.ActiveWorkers)
	}
}

func TestGetWorkerStats(t *testing.T) {
	scraper := &WebScraper{}
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)
	defer queue.Stop()

	workerStats := queue.GetWorkerStats()

	if len(workerStats) != queue.config.MaxWorkers {
		t.Errorf("Expected %d worker stats, got %d", queue.config.MaxWorkers, len(workerStats))
	}

	for i, stats := range workerStats {
		if stats == nil {
			t.Errorf("Expected non-nil stats for worker %d", i)
		}
	}
}

func TestSetConfig(t *testing.T) {
	scraper := &WebScraper{}
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)
	defer queue.Stop()

	newConfig := QueueConfig{
		MaxWorkers:             10,
		MaxQueueSize:           2000,
		MaxConcurrentPerDomain: 3,
		DefaultTimeout:         time.Second * 60,
		MaxRetries:             5,
		RetryDelay:             time.Second * 5,
		RateLimitPerDomain:     3,
		RateLimitWindow:        time.Second * 2,
		EnableRelevanceScoring: false,
		EnablePageDiscovery:    false,
		MinRelevanceScore:      0.5,
		MaxDiscoveryDepth:      5,
		MaxPagesPerDomain:      100,
	}

	queue.SetConfig(newConfig)

	config := queue.GetConfig()

	if config.MaxWorkers != newConfig.MaxWorkers {
		t.Errorf("Expected MaxWorkers %d, got %d", newConfig.MaxWorkers, config.MaxWorkers)
	}

	if config.MaxQueueSize != newConfig.MaxQueueSize {
		t.Errorf("Expected MaxQueueSize %d, got %d", newConfig.MaxQueueSize, config.MaxQueueSize)
	}

	if config.EnableRelevanceScoring != newConfig.EnableRelevanceScoring {
		t.Errorf("Expected EnableRelevanceScoring %t, got %t", newConfig.EnableRelevanceScoring, config.EnableRelevanceScoring)
	}
}

func TestJobHeapPriority(t *testing.T) {
	scraper := &WebScraper{}
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)
	defer queue.Stop()

	// Add jobs with different priorities
	jobs := []*PriorityScrapingJob{
		{
			URL:            "https://example.com/low",
			Business:       "Test Company",
			Priority:       50,
			RelevanceScore: 0.5,
			Context:        &ScrapingContext{},
			Metadata:       make(map[string]interface{}),
		},
		{
			URL:            "https://example.com/high",
			Business:       "Test Company",
			Priority:       100,
			RelevanceScore: 0.8,
			Context:        &ScrapingContext{},
			Metadata:       make(map[string]interface{}),
		},
		{
			URL:            "https://example.com/medium",
			Business:       "Test Company",
			Priority:       75,
			RelevanceScore: 0.6,
			Context:        &ScrapingContext{},
			Metadata:       make(map[string]interface{}),
		},
	}

	for _, job := range jobs {
		queue.AddJob(job)
	}

	// The highest priority job should be processed first
	// Note: In a real test, we would need to mock the scraper to verify processing order
	if queue.GetQueueSize() != 3 {
		t.Errorf("Expected queue size 3, got %d", queue.GetQueueSize())
	}
}

func TestJobHeapRelevanceScore(t *testing.T) {
	scraper := &WebScraper{}
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)
	defer queue.Stop()

	// Add jobs with same priority but different relevance scores
	jobs := []*PriorityScrapingJob{
		{
			URL:            "https://example.com/low-relevance",
			Business:       "Test Company",
			Priority:       100,
			RelevanceScore: 0.3,
			Context:        &ScrapingContext{},
			Metadata:       make(map[string]interface{}),
		},
		{
			URL:            "https://example.com/high-relevance",
			Business:       "Test Company",
			Priority:       100,
			RelevanceScore: 0.9,
			Context:        &ScrapingContext{},
			Metadata:       make(map[string]interface{}),
		},
		{
			URL:            "https://example.com/medium-relevance",
			Business:       "Test Company",
			Priority:       100,
			RelevanceScore: 0.6,
			Context:        &ScrapingContext{},
			Metadata:       make(map[string]interface{}),
		},
	}

	for _, job := range jobs {
		queue.AddJob(job)
	}

	if queue.GetQueueSize() != 3 {
		t.Errorf("Expected queue size 3, got %d", queue.GetQueueSize())
	}
}

func TestJobHeapCreationTime(t *testing.T) {
	scraper := &WebScraper{}
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)
	defer queue.Stop()

	// Add jobs with same priority and relevance score but different creation times
	now := time.Now()
	jobs := []*PriorityScrapingJob{
		{
			URL:            "https://example.com/later",
			Business:       "Test Company",
			Priority:       100,
			RelevanceScore: 0.5,
			CreatedAt:      now.Add(time.Second),
			Context:        &ScrapingContext{},
			Metadata:       make(map[string]interface{}),
		},
		{
			URL:            "https://example.com/earlier",
			Business:       "Test Company",
			Priority:       100,
			RelevanceScore: 0.5,
			CreatedAt:      now,
			Context:        &ScrapingContext{},
			Metadata:       make(map[string]interface{}),
		},
	}

	for _, job := range jobs {
		queue.AddJob(job)
	}

	if queue.GetQueueSize() != 2 {
		t.Errorf("Expected queue size 2, got %d", queue.GetQueueSize())
	}
}

func TestQueueFull(t *testing.T) {
	scraper := &WebScraper{}
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)
	defer queue.Stop()

	// Set a small queue size for testing
	queue.SetConfig(QueueConfig{
		MaxQueueSize: 2,
		MaxWorkers:   1,
	})

	// Add jobs up to the limit
	jobs := []*PriorityScrapingJob{
		{
			URL:      "https://example.com/1",
			Business: "Test Company",
			Priority: 100,
			Context:  &ScrapingContext{},
			Metadata: make(map[string]interface{}),
		},
		{
			URL:      "https://example.com/2",
			Business: "Test Company",
			Priority: 90,
			Context:  &ScrapingContext{},
			Metadata: make(map[string]interface{}),
		},
	}

	for _, job := range jobs {
		err := queue.AddJob(job)
		if err != nil {
			t.Errorf("Expected no error adding job, got %v", err)
		}
	}

	// Try to add one more job
	extraJob := &PriorityScrapingJob{
		URL:      "https://example.com/3",
		Business: "Test Company",
		Priority: 80,
		Context:  &ScrapingContext{},
		Metadata: make(map[string]interface{}),
	}

	err := queue.AddJob(extraJob)
	if err == nil {
		t.Error("Expected error when queue is full")
	}

	if queue.GetQueueSize() != 2 {
		t.Errorf("Expected queue size 2, got %d", queue.GetQueueSize())
	}
}

func TestGetResult(t *testing.T) {
	scraper := &WebScraper{}
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)
	defer queue.Stop()

	// Test getting result with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := queue.GetResult(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded error, got %v", err)
	}
}

func TestGetResults(t *testing.T) {
	scraper := &WebScraper{}
	relevanceScorer := NewPageRelevanceScorer()
	pageDiscovery := NewIntelligentPageDiscovery(scraper)

	queue := NewPriorityScrapingQueue(scraper, relevanceScorer, pageDiscovery)
	defer queue.Stop()

	// Test getting multiple results with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := queue.GetResults(ctx, 3)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded error, got %v", err)
	}
}

func TestPriorityScrapingJobJSON(t *testing.T) {
	job := &PriorityScrapingJob{
		URL:            "https://example.com/about",
		Business:       "Test Company",
		Priority:       100,
		RelevanceScore: 0.8,
		PageType:       PageTypeAbout,
		Depth:          0,
		Retries:        0,
		MaxRetries:     3,
		Timeout:        time.Second * 30,
		CreatedAt:      time.Now(),
		ScheduledAt:    time.Now(),
		Context: &ScrapingContext{
			Industry: "technology",
			Location: "United States",
		},
		Metadata: map[string]interface{}{
			"test_key": "test_value",
		},
	}

	// Test that the struct can be marshaled to JSON
	_, err := json.Marshal(job)
	if err != nil {
		t.Errorf("Failed to marshal PriorityScrapingJob to JSON: %v", err)
	}
}

func TestScrapingResultJSON(t *testing.T) {
	result := &ScrapingResult{
		Job: &PriorityScrapingJob{
			URL:      "https://example.com/about",
			Business: "Test Company",
			Priority: 100,
			Context:  &ScrapingContext{},
			Metadata: make(map[string]interface{}),
		},
		Content: &ScrapedContent{
			URL:   "https://example.com/about",
			Title: "About Us",
			Text:  "This is about our company",
		},
		RelevanceScore: &PageRelevanceScore{
			OverallScore: 0.8,
		},
		ProcessingTime: time.Second * 2,
		Success:        true,
		CompletedAt:    time.Now(),
		DiscoveredLinks: []string{
			"https://example.com/contact",
			"https://example.com/services",
		},
		NextJobs: []*PriorityScrapingJob{
			{
				URL:      "https://example.com/contact",
				Business: "Test Company",
				Priority: 90,
				Context:  &ScrapingContext{},
				Metadata: make(map[string]interface{}),
			},
		},
	}

	// Test that the struct can be marshaled to JSON
	_, err := json.Marshal(result)
	if err != nil {
		t.Errorf("Failed to marshal ScrapingResult to JSON: %v", err)
	}
}

func TestQueueStatsJSON(t *testing.T) {
	stats := &QueueStats{
		TotalJobsQueued:       100,
		TotalJobsCompleted:    80,
		TotalJobsFailed:       5,
		AverageProcessingTime: time.Second * 2,
		QueueSize:             15,
		ActiveWorkers:         3,
		LastUpdated:           time.Now(),
	}

	// Test that the struct can be marshaled to JSON
	_, err := json.Marshal(stats)
	if err != nil {
		t.Errorf("Failed to marshal QueueStats to JSON: %v", err)
	}
}

func TestWorkerStatsJSON(t *testing.T) {
	stats := &WorkerStats{
		JobsProcessed:         50,
		JobsSucceeded:         45,
		JobsFailed:            5,
		TotalProcessingTime:   time.Second * 100,
		AverageProcessingTime: time.Second * 2,
		LastJobAt:             time.Now(),
	}

	// Test that the struct can be marshaled to JSON
	_, err := json.Marshal(stats)
	if err != nil {
		t.Errorf("Failed to marshal WorkerStats to JSON: %v", err)
	}
}
