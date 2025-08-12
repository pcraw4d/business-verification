package webanalysis

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"container/heap"
)

// PriorityScrapingJob represents a prioritized scraping task
type PriorityScrapingJob struct {
	URL            string                 `json:"url"`
	Business       string                 `json:"business"`
	Priority       int                    `json:"priority"`
	RelevanceScore float64                `json:"relevance_score"`
	PageType       PageType               `json:"page_type"`
	Depth          int                    `json:"depth"`
	Retries        int                    `json:"retries"`
	MaxRetries     int                    `json:"max_retries"`
	Timeout        time.Duration          `json:"timeout"`
	CreatedAt      time.Time              `json:"created_at"`
	ScheduledAt    time.Time              `json:"scheduled_at"`
	Context        *ScrapingContext       `json:"context"`
	Metadata       map[string]interface{} `json:"metadata"`
	index          int                    // Used by heap interface
}

// ScrapingContext provides context for scraping operations
type ScrapingContext struct {
	Domain           string            `json:"domain"`
	Industry         string            `json:"industry"`
	Location         string            `json:"location"`
	BusinessType     string            `json:"business_type"`
	DiscoverySession string            `json:"discovery_session"`
	ParentURL        string            `json:"parent_url"`
	ScoringContext   *ScoringContext   `json:"scoring_context"`
	Metadata         map[string]string `json:"metadata"`
}

// ScrapingResult represents the result of a scraping job
type ScrapingResult struct {
	Job             *PriorityScrapingJob `json:"job"`
	Content         *ScrapedContent      `json:"content"`
	RelevanceScore  *PageRelevanceScore  `json:"relevance_score"`
	ProcessingTime  time.Duration        `json:"processing_time"`
	Success         bool                 `json:"success"`
	Error           string               `json:"error,omitempty"`
	CompletedAt     time.Time            `json:"completed_at"`
	DiscoveredLinks []string             `json:"discovered_links,omitempty"`
	NextJobs        []*PriorityScrapingJob `json:"next_jobs,omitempty"`
}

// PriorityScrapingQueue manages priority-based scraping operations
type PriorityScrapingQueue struct {
	queue           *JobHeap
	scraper         *WebScraper
	relevanceScorer *PageRelevanceScorer
	pageDiscovery   *IntelligentPageDiscovery
	workers         []*ScrapingWorker
	results         chan *ScrapingResult
	config          QueueConfig
	mu              sync.RWMutex
	stats           *QueueStats
	ctx             context.Context
	cancel          context.CancelFunc
}

// JobHeap implements heap.Interface for priority queue
type JobHeap []*PriorityScrapingJob

// QueueConfig holds configuration for the priority scraping queue
type QueueConfig struct {
	MaxWorkers             int           `json:"max_workers"`
	MaxQueueSize           int           `json:"max_queue_size"`
	MaxConcurrentPerDomain int           `json:"max_concurrent_per_domain"`
	DefaultTimeout         time.Duration `json:"default_timeout"`
	MaxRetries             int           `json:"max_retries"`
	RetryDelay             time.Duration `json:"retry_delay"`
	RateLimitPerDomain     int           `json:"rate_limit_per_domain"`
	RateLimitWindow        time.Duration `json:"rate_limit_window"`
	EnableRelevanceScoring bool          `json:"enable_relevance_scoring"`
	EnablePageDiscovery    bool          `json:"enable_page_discovery"`
	MinRelevanceScore      float64       `json:"min_relevance_score"`
	MaxDiscoveryDepth      int           `json:"max_discovery_depth"`
	MaxPagesPerDomain      int           `json:"max_pages_per_domain"`
}

// QueueStats holds statistics about the queue
type QueueStats struct {
	TotalJobsQueued       int           `json:"total_jobs_queued"`
	TotalJobsCompleted    int           `json:"total_jobs_completed"`
	TotalJobsFailed       int           `json:"total_jobs_failed"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	QueueSize             int           `json:"queue_size"`
	ActiveWorkers         int           `json:"active_workers"`
	LastUpdated           time.Time     `json:"last_updated"`
	mu                    sync.RWMutex
}

// ScrapingWorker represents a worker that processes scraping jobs
type ScrapingWorker struct {
	id            int
	queue         *PriorityScrapingQueue
	active        bool
	processingJob *PriorityScrapingJob
	startedAt     time.Time
	stats         *WorkerStats
}

// WorkerStats holds statistics for a worker
type WorkerStats struct {
	JobsProcessed         int           `json:"jobs_processed"`
	JobsSucceeded         int           `json:"jobs_succeeded"`
	JobsFailed            int           `json:"jobs_failed"`
	TotalProcessingTime   time.Duration `json:"total_processing_time"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	LastJobAt             time.Time     `json:"last_job_at"`
}

// NewPriorityScrapingQueue creates a new priority-based scraping queue
func NewPriorityScrapingQueue(scraper *WebScraper, relevanceScorer *PageRelevanceScorer, pageDiscovery *IntelligentPageDiscovery) *PriorityScrapingQueue {
	ctx, cancel := context.WithCancel(context.Background())

	queue := &PriorityScrapingQueue{
		queue:           &JobHeap{},
		scraper:         scraper,
		relevanceScorer: relevanceScorer,
		pageDiscovery:   pageDiscovery,
		workers:         []*ScrapingWorker{},
		results:         make(chan *ScrapingResult, 1000),
		config: QueueConfig{
			MaxWorkers:             5,
			MaxQueueSize:           1000,
			MaxConcurrentPerDomain: 2,
			DefaultTimeout:         time.Second * 30,
			MaxRetries:             3,
			RetryDelay:             time.Second * 2,
			RateLimitPerDomain:     2,
			RateLimitWindow:        time.Second,
			EnableRelevanceScoring: true,
			EnablePageDiscovery:    true,
			MinRelevanceScore:      0.3,
			MaxDiscoveryDepth:      3,
			MaxPagesPerDomain:      50,
		},
		stats: &QueueStats{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	heap.Init(queue.queue)
	queue.startWorkers()

	return queue
}

// startWorkers starts the scraping workers
func (psq *PriorityScrapingQueue) startWorkers() {
	for i := 0; i < psq.config.MaxWorkers; i++ {
		worker := &ScrapingWorker{
			id:     i,
			queue:  psq,
			active: true,
			stats:  &WorkerStats{},
		}
		psq.workers = append(psq.workers, worker)
		go worker.start()
	}
}

// AddJob adds a scraping job to the priority queue
func (psq *PriorityScrapingQueue) AddJob(job *PriorityScrapingJob) error {
	psq.mu.Lock()
	defer psq.mu.Unlock()

	// Check queue size limit
	if psq.queue.Len() >= psq.config.MaxQueueSize {
		return fmt.Errorf("queue is full (max size: %d)", psq.config.MaxQueueSize)
	}

	// Set default values
	if job.Timeout == 0 {
		job.Timeout = psq.config.DefaultTimeout
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = psq.config.MaxRetries
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.ScheduledAt.IsZero() {
		job.ScheduledAt = time.Now()
	}

	// Add to priority queue
	heap.Push(psq.queue, job)

	// Update stats
	psq.stats.mu.Lock()
	psq.stats.TotalJobsQueued++
	psq.stats.QueueSize = psq.queue.Len()
	psq.stats.LastUpdated = time.Now()
	psq.stats.mu.Unlock()

	log.Printf("Added job to queue: %s (priority: %d, relevance: %.2f)", job.URL, job.Priority, job.RelevanceScore)
	return nil
}

// AddJobs adds multiple scraping jobs to the priority queue
func (psq *PriorityScrapingQueue) AddJobs(jobs []*PriorityScrapingJob) error {
	for _, job := range jobs {
		if err := psq.AddJob(job); err != nil {
			return fmt.Errorf("failed to add job %s: %w", job.URL, err)
		}
	}
	return nil
}

// GetResult retrieves a scraping result from the results channel
func (psq *PriorityScrapingQueue) GetResult(ctx context.Context) (*ScrapingResult, error) {
	select {
	case result := <-psq.results:
		return result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetResults retrieves multiple scraping results
func (psq *PriorityScrapingQueue) GetResults(ctx context.Context, count int) ([]*ScrapingResult, error) {
	var results []*ScrapingResult

	for i := 0; i < count; i++ {
		result, err := psq.GetResult(ctx)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}

	return results, nil
}

// GetStats returns current queue statistics
func (psq *PriorityScrapingQueue) GetStats() *QueueStats {
	psq.stats.mu.RLock()
	defer psq.stats.mu.RUnlock()

	// Update current stats
	psq.stats.QueueSize = psq.queue.Len()
	psq.stats.ActiveWorkers = psq.getActiveWorkerCount()
	psq.stats.LastUpdated = time.Now()

	return psq.stats
}

// getActiveWorkerCount returns the number of active workers
func (psq *PriorityScrapingQueue) getActiveWorkerCount() int {
	count := 0
	for _, worker := range psq.workers {
		if worker.active {
			count++
		}
	}
	return count
}

// Stop stops the priority scraping queue
func (psq *PriorityScrapingQueue) Stop() {
	psq.cancel()

	// Stop all workers
	for _, worker := range psq.workers {
		worker.active = false
	}

	close(psq.results)
}

// CreateJobFromDiscoveryResult creates a scraping job from a page discovery result
func (psq *PriorityScrapingQueue) CreateJobFromDiscoveryResult(result *PageDiscoveryResult, business string, context *ScrapingContext) *PriorityScrapingJob {
	return &PriorityScrapingJob{
		URL:            result.URL,
		Business:       business,
		Priority:       result.Priority,
		RelevanceScore: result.RelevanceScore,
		PageType:       result.PageType,
		Depth:          result.Depth,
		Context:        context,
		Metadata: map[string]interface{}{
			"discovered_at":      result.DiscoveredAt,
			"content_indicators": result.ContentIndicators,
			"business_keywords":  result.BusinessKeywords,
		},
	}
}

// CreateJobFromURL creates a scraping job from a URL
func (psq *PriorityScrapingQueue) CreateJobFromURL(url, business string, priority int, context *ScrapingContext) *PriorityScrapingJob {
	return &PriorityScrapingJob{
		URL:      url,
		Business: business,
		Priority: priority,
		Context:  context,
		Metadata: make(map[string]interface{}),
	}
}

// start starts the worker processing loop
func (w *ScrapingWorker) start() {
	log.Printf("Worker %d started", w.id)

	for w.active {
		select {
		case <-w.queue.ctx.Done():
			w.active = false
			return
		default:
			w.processNextJob()
		}
	}

	log.Printf("Worker %d stopped", w.id)
}

// processNextJob processes the next job from the queue
func (w *ScrapingWorker) processNextJob() {
	// Get next job from queue
	job := w.getNextJob()
	if job == nil {
		time.Sleep(100 * time.Millisecond)
		return
	}

	w.processingJob = job
	w.startedAt = time.Now()

	log.Printf("Worker %d processing job: %s (priority: %d)", w.id, job.URL, job.Priority)

	// Process the job
	result := w.processJob(job)

	// Send result to results channel
	select {
	case w.queue.results <- result:
		// Result sent successfully
	default:
		log.Printf("Warning: Results channel is full, dropping result for %s", job.URL)
	}

	// Update worker stats
	w.updateStats(result)

	w.processingJob = nil
}

// getNextJob gets the next job from the priority queue
func (w *ScrapingWorker) getNextJob() *PriorityScrapingJob {
	w.queue.mu.Lock()
	defer w.queue.mu.Unlock()

	if w.queue.queue.Len() == 0 {
		return nil
	}

	job := heap.Pop(w.queue.queue).(*PriorityScrapingJob)

	// Update queue stats
	w.queue.stats.mu.Lock()
	w.queue.stats.QueueSize = w.queue.queue.Len()
	w.queue.stats.mu.Unlock()

	return job
}

// processJob processes a single scraping job
func (w *ScrapingWorker) processJob(job *PriorityScrapingJob) *ScrapingResult {
	startTime := time.Now()
	result := &ScrapingResult{
		Job:         job,
		CompletedAt: time.Now(),
	}

	// Create scraping job for the scraper
	scrapingJob := &ScrapingJob{
		URL:        job.URL,
		Business:   job.Business,
		Priority:   job.Priority,
		Timeout:    job.Timeout,
		MaxRetries: job.MaxRetries,
	}

	// Scrape the page
	content, err := w.queue.scraper.ScrapeWebsite(scrapingJob)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.ProcessingTime = time.Since(startTime)
		return result
	}

	result.Content = content
	result.Success = true

	// Calculate relevance score if enabled
	if w.queue.config.EnableRelevanceScoring && w.queue.relevanceScorer != nil {
		scoringContext := job.Context.ScoringContext
		if scoringContext == nil {
			scoringContext = &ScoringContext{
				Industry:     job.Context.Industry,
				Location:     job.Context.Location,
				BusinessType: job.Context.BusinessType,
			}
		}

		relevanceScore := w.queue.relevanceScorer.ScorePage(content, job.Business, scoringContext)
		result.RelevanceScore = relevanceScore

		// Update job with relevance score
		job.RelevanceScore = relevanceScore.OverallScore
	}

	// Discover additional pages if enabled
	if w.queue.config.EnablePageDiscovery && w.queue.pageDiscovery != nil && job.Depth < w.queue.config.MaxDiscoveryDepth {
		w.discoverAdditionalPages(job, content, result)
	}

	result.ProcessingTime = time.Since(startTime)
	return result
}

// discoverAdditionalPages discovers additional pages from scraped content
func (w *ScrapingWorker) discoverAdditionalPages(job *PriorityScrapingJob, content *ScrapedContent, result *ScrapingResult) {
	// Extract internal links
	links := w.queue.pageDiscovery.extractInternalLinks(job.URL, content.HTML)

	// Filter and create new jobs
	var nextJobs []*PriorityScrapingJob
	for _, link := range links {
		// Check if we should include this link
		if !w.queue.pageDiscovery.shouldIncludeURL(link) {
			continue
		}

		// Create new job with lower priority and higher depth
		nextJob := &PriorityScrapingJob{
			URL:        link,
			Business:   job.Business,
			Priority:   job.Priority - 10, // Lower priority for discovered pages
			Depth:      job.Depth + 1,
			Context:    job.Context,
			Timeout:    job.Timeout,
			MaxRetries: job.MaxRetries,
			Metadata:   make(map[string]interface{}),
		}

		// Add to queue if not at max depth
		if nextJob.Depth <= w.queue.config.MaxDiscoveryDepth {
			nextJobs = append(nextJobs, nextJob)
		}
	}

	result.DiscoveredLinks = links
	result.NextJobs = nextJobs

	// Add discovered jobs to queue
	if len(nextJobs) > 0 {
		w.queue.AddJobs(nextJobs)
	}
}

// updateStats updates worker statistics
func (w *ScrapingWorker) updateStats(result *ScrapingResult) {
	w.stats.JobsProcessed++
	w.stats.TotalProcessingTime += result.ProcessingTime
	w.stats.LastJobAt = time.Now()

	if result.Success {
		w.stats.JobsSucceeded++
	} else {
		w.stats.JobsFailed++
	}

	// Calculate average processing time
	if w.stats.JobsProcessed > 0 {
		w.stats.AverageProcessingTime = w.stats.TotalProcessingTime / time.Duration(w.stats.JobsProcessed)
	}

	// Update queue stats
	w.queue.stats.mu.Lock()
	w.queue.stats.TotalJobsCompleted++
	if !result.Success {
		w.queue.stats.TotalJobsFailed++
	}
	w.queue.stats.LastUpdated = time.Now()
	w.queue.stats.mu.Unlock()
}

// Heap interface implementation for JobHeap

func (h JobHeap) Len() int { return len(h) }

func (h JobHeap) Less(i, j int) bool {
	// Higher priority jobs come first
	if h[i].Priority != h[j].Priority {
		return h[i].Priority > h[j].Priority
	}

	// If priorities are equal, higher relevance score comes first
	if h[i].RelevanceScore != h[j].RelevanceScore {
		return h[i].RelevanceScore > h[j].RelevanceScore
	}

	// If relevance scores are equal, earlier creation time comes first
	return h[i].CreatedAt.Before(h[j].CreatedAt)
}

func (h JobHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *JobHeap) Push(x interface{}) {
	n := len(*h)
	job := x.(*PriorityScrapingJob)
	job.index = n
	*h = append(*h, job)
}

func (h *JobHeap) Pop() interface{} {
	old := *h
	n := len(old)
	job := old[n-1]
	old[n-1] = nil // avoid memory leak
	job.index = -1 // for safety
	*h = old[0 : n-1]
	return job
}

// GetWorkerStats returns statistics for all workers
func (psq *PriorityScrapingQueue) GetWorkerStats() []*WorkerStats {
	var stats []*WorkerStats
	for _, worker := range psq.workers {
		stats = append(stats, worker.stats)
	}
	return stats
}

// GetQueueSize returns the current queue size
func (psq *PriorityScrapingQueue) GetQueueSize() int {
	psq.mu.RLock()
	defer psq.mu.RUnlock()
	return psq.queue.Len()
}

// IsEmpty returns true if the queue is empty
func (psq *PriorityScrapingQueue) IsEmpty() bool {
	return psq.GetQueueSize() == 0
}

// Clear clears all jobs from the queue
func (psq *PriorityScrapingQueue) Clear() {
	psq.mu.Lock()
	defer psq.mu.Unlock()

	// Clear the heap
	psq.queue = &JobHeap{}
	heap.Init(psq.queue)

	// Update stats
	psq.stats.mu.Lock()
	psq.stats.QueueSize = 0
	psq.stats.LastUpdated = time.Now()
	psq.stats.mu.Unlock()
}

// GetJobCount returns the number of jobs in the queue
func (psq *PriorityScrapingQueue) GetJobCount() int {
	return psq.GetQueueSize()
}

// GetActiveWorkerCount returns the number of active workers
func (psq *PriorityScrapingQueue) GetActiveWorkerCount() int {
	count := 0
	for _, worker := range psq.workers {
		if worker.active {
			count++
		}
	}
	return count
}

// SetConfig updates the queue configuration
func (psq *PriorityScrapingQueue) SetConfig(config QueueConfig) {
	psq.mu.Lock()
	defer psq.mu.Unlock()
	psq.config = config
}

// GetConfig returns the current queue configuration
func (psq *PriorityScrapingQueue) GetConfig() QueueConfig {
	psq.mu.RLock()
	defer psq.mu.RUnlock()
	return psq.config
}
