package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// RequestTrace represents a parsed request trace from logs
type RequestTrace struct {
	RequestID           string
	TotalDuration       time.Duration
	CacheLookupDuration time.Duration
	ScrapingDuration    time.Duration
	ScrapingStrategy    string
	CodeGenDuration     time.Duration
	MCCGenDuration      time.Duration
	NAICSGenDuration    time.Duration
	SICGenDuration      time.Duration
	DBQueryDuration     time.Duration
	MLServiceDuration   time.Duration
	PlaywrightDuration  time.Duration
	SupabaseDuration    time.Duration
	ResponseBuildDuration time.Duration
	QueueWaitDuration   time.Duration
	QueueDepth          int
	StageDurations      map[string]time.Duration
	Industry            string
	WebsiteURL          string
	Error               string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run analyze_slow_requests.go <log_file> [threshold_seconds]")
		fmt.Println("Example: go run analyze_slow_requests.go railway.log 30")
		os.Exit(1)
	}

	logFile := os.Args[1]
	thresholdSeconds := 30.0
	if len(os.Args) > 2 {
		if t, err := strconv.ParseFloat(os.Args[2], 64); err == nil {
			thresholdSeconds = t
		}
	}

	file, err := os.Open(logFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	traces := []RequestTrace{}
	
	// Regex patterns for parsing log entries
	traceCompletePattern := regexp.MustCompile(`\[TRACE-COMPLETE\]`)
	requestIDPattern := regexp.MustCompile(`"request_id":"([^"]+)"`)
	durationPattern := regexp.MustCompile(`"total_duration":([0-9.]+)`)
	
	var currentTrace *RequestTrace
	var traceBuffer strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		
		// Check if this is a trace-complete log entry
		if traceCompletePattern.MatchString(line) {
			// Parse the trace from the accumulated buffer or current line
			if trace := parseTrace(line); trace != nil {
				traces = append(traces, *trace)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	_ = currentTrace
	_ = traceBuffer

	// Filter slow requests
	threshold := time.Duration(thresholdSeconds) * time.Second
	slowRequests := []RequestTrace{}
	for _, trace := range traces {
		if trace.TotalDuration >= threshold {
			slowRequests = append(slowRequests, trace)
		}
	}

	// Sort by duration (slowest first)
	sort.Slice(slowRequests, func(i, j int) bool {
		return slowRequests[i].TotalDuration > slowRequests[j].TotalDuration
	})

	// Generate report
	fmt.Printf("=== Slow Request Analysis Report ===\n\n")
	fmt.Printf("Threshold: %.1f seconds\n", thresholdSeconds)
	fmt.Printf("Total traces found: %d\n", len(traces))
	fmt.Printf("Slow requests (>%.1fs): %d\n\n", thresholdSeconds, len(slowRequests))

	if len(slowRequests) == 0 {
		fmt.Println("No slow requests found.")
		return
	}

	// Top 10 slowest requests
	fmt.Println("=== Top 10 Slowest Requests ===")
	maxPrint := 10
	if len(slowRequests) < maxPrint {
		maxPrint = len(slowRequests)
	}
	for i := 0; i < maxPrint; i++ {
		trace := slowRequests[i]
		fmt.Printf("\n%d. Request ID: %s\n", i+1, trace.RequestID)
		fmt.Printf("   Total Duration: %.2fs\n", trace.TotalDuration.Seconds())
		fmt.Printf("   Cache Lookup: %.2fms\n", trace.CacheLookupDuration.Seconds()*1000)
		fmt.Printf("   Scraping: %.2fms (Strategy: %s)\n", trace.ScrapingDuration.Seconds()*1000, trace.ScrapingStrategy)
		fmt.Printf("   Code Generation: %.2fms\n", trace.CodeGenDuration.Seconds()*1000)
		fmt.Printf("   ML Service: %.2fms\n", trace.MLServiceDuration.Seconds()*1000)
		fmt.Printf("   Queue Wait: %.2fms (Depth: %d)\n", trace.QueueWaitDuration.Seconds()*1000, trace.QueueDepth)
		if trace.Industry != "" {
			fmt.Printf("   Industry: %s\n", trace.Industry)
		}
		if trace.WebsiteURL != "" {
			fmt.Printf("   Website: %s\n", trace.WebsiteURL)
		}
		if trace.Error != "" {
			fmt.Printf("   Error: %s\n", trace.Error)
		}
	}

	// Pattern analysis
	fmt.Println("\n=== Pattern Analysis ===")
	analyzePatterns(slowRequests)
}

func parseTrace(line string) *RequestTrace {
	trace := &RequestTrace{
		StageDurations: make(map[string]time.Duration),
	}

	// Try to parse as JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
		// If not JSON, try regex parsing
		return parseTraceRegex(line)
	}

	// Extract request_id
	if rid, ok := logEntry["request_id"].(string); ok {
		trace.RequestID = rid
	}

	// Extract trace_summary if available
	if summary, ok := logEntry["trace_summary"].(map[string]interface{}); ok {
		if td, ok := summary["total_duration_ms"].(float64); ok {
			trace.TotalDuration = time.Duration(td) * time.Millisecond
		}
		if cd, ok := summary["cache_lookup_ms"].(float64); ok {
			trace.CacheLookupDuration = time.Duration(cd) * time.Millisecond
		}
		if sd, ok := summary["scraping_duration_ms"].(float64); ok {
			trace.ScrapingDuration = time.Duration(sd) * time.Millisecond
		}
		if ss, ok := summary["scraping_strategy"].(string); ok {
			trace.ScrapingStrategy = ss
		}
		if cgd, ok := summary["code_gen_duration_ms"].(float64); ok {
			trace.CodeGenDuration = time.Duration(cgd) * time.Millisecond
		}
		if mld, ok := summary["ml_service_duration_ms"].(float64); ok {
			trace.MLServiceDuration = time.Duration(mld) * time.Millisecond
		}
		if qwd, ok := summary["queue_wait_duration_ms"].(float64); ok {
			trace.QueueWaitDuration = time.Duration(qwd) * time.Millisecond
		}
		if qd, ok := summary["queue_depth"].(float64); ok {
			trace.QueueDepth = int(qd)
		}
	}

	// Extract from metadata if available
	if metadata, ok := logEntry["metadata"].(map[string]interface{}); ok {
		if industry, ok := metadata["primary_industry"].(string); ok {
			trace.Industry = industry
		}
		if url, ok := metadata["website_url"].(string); ok {
			trace.WebsiteURL = url
		}
	}

	return trace
}

func parseTraceRegex(line string) *RequestTrace {
	trace := &RequestTrace{
		StageDurations: make(map[string]time.Duration),
	}

	// Extract request ID
	requestIDPattern := regexp.MustCompile(`"request_id":"([^"]+)"`)
	if matches := requestIDPattern.FindStringSubmatch(line); len(matches) > 1 {
		trace.RequestID = matches[1]
	}

	// Extract total duration
	durationPattern := regexp.MustCompile(`"total_duration":([0-9.]+)`)
	if matches := durationPattern.FindStringSubmatch(line); len(matches) > 1 {
		if d, err := strconv.ParseFloat(matches[1], 64); err == nil {
			trace.TotalDuration = time.Duration(d) * time.Second
		}
	}

	return trace
}

func analyzePatterns(slowRequests []RequestTrace) {
	// Analyze by scraping strategy
	strategyCounts := make(map[string]int)
	strategyDurations := make(map[string]time.Duration)
	
	// Analyze by industry
	industryCounts := make(map[string]int)
	
	// Analyze bottlenecks
	var totalScraping, totalCodeGen, totalML, totalQueueWait time.Duration
	
	for _, trace := range slowRequests {
		if trace.ScrapingStrategy != "" {
			strategyCounts[trace.ScrapingStrategy]++
			strategyDurations[trace.ScrapingStrategy] += trace.ScrapingDuration
		}
		if trace.Industry != "" {
			industryCounts[trace.Industry]++
		}
		totalScraping += trace.ScrapingDuration
		totalCodeGen += trace.CodeGenDuration
		totalML += trace.MLServiceDuration
		totalQueueWait += trace.QueueWaitDuration
	}

	fmt.Printf("Scraping Strategy Distribution:\n")
	for strategy, count := range strategyCounts {
		avgDuration := strategyDurations[strategy] / time.Duration(count)
		fmt.Printf("  %s: %d requests (avg: %.2fms)\n", strategy, count, avgDuration.Seconds()*1000)
	}

	fmt.Printf("\nTop Industries in Slow Requests:\n")
	type industryStat struct {
		industry string
		count    int
	}
	industryStats := []industryStat{}
	for industry, count := range industryCounts {
		industryStats = append(industryStats, industryStat{industry, count})
	}
	sort.Slice(industryStats, func(i, j int) bool {
		return industryStats[i].count > industryStats[j].count
	})
	for i := 0; i < len(industryStats) && i < 5; i++ {
		fmt.Printf("  %s: %d requests\n", industryStats[i].industry, industryStats[i].count)
	}

	fmt.Printf("\nAverage Bottleneck Durations:\n")
	count := len(slowRequests)
	if count > 0 {
		fmt.Printf("  Scraping: %.2fms\n", (totalScraping / time.Duration(count)).Seconds()*1000)
		fmt.Printf("  Code Generation: %.2fms\n", (totalCodeGen / time.Duration(count)).Seconds()*1000)
		fmt.Printf("  ML Service: %.2fms\n", (totalML / time.Duration(count)).Seconds()*1000)
		fmt.Printf("  Queue Wait: %.2fms\n", (totalQueueWait / time.Duration(count)).Seconds()*1000)
	}
}

