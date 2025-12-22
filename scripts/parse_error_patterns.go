package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ErrorCategory represents a category of errors
type ErrorCategory struct {
	Name        string
	Count       int
	Examples    []string
	Patterns    []string
	RequestIDs  []string
}

// ErrorPattern represents a parsed error from logs
type ErrorPattern struct {
	RequestID   string
	Timestamp   string
	Level       string
	Message     string
	Error       string
	Category    string
	Context     map[string]interface{}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run parse_error_patterns.go <log_file>")
		fmt.Println("Example: go run parse_error_patterns.go docs/railway\\ log/complete\\ log.json")
		os.Exit(1)
	}

	logFile := os.Args[1]
	file, err := os.Open(logFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	errors := []ErrorPattern{}
	
	// Patterns for error detection
	dnsPattern := regexp.MustCompile(`(?i)(dns|lookup|resolve|no such host|NXDOMAIN)`)
	timeoutPattern := regexp.MustCompile(`(?i)(timeout|deadline|context deadline|exceeded|canceled)`)
	http4xxPattern := regexp.MustCompile(`(?i)(4\d{2}|bad request|unauthorized|forbidden|not found|method not allowed|conflict|too many requests)`)
	http5xxPattern := regexp.MustCompile(`(?i)(5\d{2}|internal server error|bad gateway|service unavailable|gateway timeout)`)
	networkPattern := regexp.MustCompile(`(?i)(network|connection|refused|reset|unreachable|no route to host)`)
	contextPattern := regexp.MustCompile(`(?i)(context canceled|context deadline exceeded|context expired)`)
	parsePattern := regexp.MustCompile(`(?i)(parse|unmarshal|invalid json|syntax error|malformed)`)
	
	// Try to parse as JSON array first
	decoder := json.NewDecoder(file)
	var logEntries []map[string]interface{}
	if err := decoder.Decode(&logEntries); err == nil {
		// Successfully parsed as JSON array
		for _, logEntry := range logEntries {
			processLogEntry(logEntry, &errors, dnsPattern, timeoutPattern, 
				http4xxPattern, http5xxPattern, networkPattern, contextPattern, parsePattern)
		}
	} else {
		// Try line-delimited JSON
		file.Seek(0, 0)
		scanner := bufio.NewScanner(file)
		// Increase buffer size for large lines
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024) // 1MB buffer
		
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			
			// Try to parse as JSON log entry
			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
				// If not JSON, try to extract errors from plain text
				if strings.Contains(strings.ToLower(line), "error") || 
				   strings.Contains(strings.ToLower(line), "failed") ||
				   strings.Contains(strings.ToLower(line), "panic") {
					errorPattern := extractErrorFromText(line, lineNum)
					if errorPattern != nil {
						errors = append(errors, *errorPattern)
					}
				}
				continue
			}
			
			processLogEntry(logEntry, &errors, dnsPattern, timeoutPattern, 
				http4xxPattern, http5xxPattern, networkPattern, contextPattern, parsePattern)
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
	}

	// Categorize errors
	categories := make(map[string]*ErrorCategory)
	
	for _, err := range errors {
		category := err.Category
		if category == "" {
			category = "unknown"
		}
		
		if categories[category] == nil {
			categories[category] = &ErrorCategory{
				Name:     category,
				Examples: []string{},
				Patterns: []string{},
				RequestIDs: []string{},
			}
		}
		
		cat := categories[category]
		cat.Count++
		
		// Add example (limit to 5 per category)
		if len(cat.Examples) < 5 {
			example := err.Error
			if example == "" {
				example = err.Message
			}
			if example != "" && !contains(cat.Examples, example) {
				cat.Examples = append(cat.Examples, example)
			}
		}
		
		// Add request ID (limit to 10 per category)
		if err.RequestID != "" && len(cat.RequestIDs) < 10 && !contains(cat.RequestIDs, err.RequestID) {
			cat.RequestIDs = append(cat.RequestIDs, err.RequestID)
		}
	}

	// Generate report
	fmt.Printf("=== Error Pattern Analysis Report ===\n\n")
	fmt.Printf("Total errors found: %d\n\n", len(errors))
	
	// Sort categories by count
	type catStat struct {
		name  string
		count int
	}
	catStats := []catStat{}
	for name, cat := range categories {
		catStats = append(catStats, catStat{name, cat.Count})
	}
	sort.Slice(catStats, func(i, j int) bool {
		return catStats[i].count > catStats[j].count
	})
	
	fmt.Println("=== Error Distribution by Category ===")
	for _, stat := range catStats {
		cat := categories[stat.name]
		percentage := float64(cat.Count) / float64(len(errors)) * 100
		fmt.Printf("\n%s: %d errors (%.1f%%)\n", cat.Name, cat.Count, percentage)
		
		if len(cat.Examples) > 0 {
			fmt.Printf("  Examples:\n")
			for i, example := range cat.Examples {
				if i >= 3 {
					break
				}
				// Truncate long examples
				if len(example) > 100 {
					example = example[:100] + "..."
				}
				fmt.Printf("    - %s\n", example)
			}
		}
		
		if len(cat.RequestIDs) > 0 {
			fmt.Printf("  Sample Request IDs:\n")
			for i, reqID := range cat.RequestIDs {
				if i >= 5 {
					break
				}
				fmt.Printf("    - %s\n", reqID)
			}
		}
	}
	
	// Write detailed report to file
	writeDetailedReport(categories, errors)
}

func processLogEntry(logEntry map[string]interface{}, errors *[]ErrorPattern, 
	dnsPattern, timeoutPattern, http4xxPattern, http5xxPattern, networkPattern, 
	contextPattern, parsePattern *regexp.Regexp) {
	
	// Check if this is an error log entry
	// Railway logs have message in "message" field, and attributes.level
	message, _ := logEntry["message"].(string)
	attributes, _ := logEntry["attributes"].(map[string]interface{})
	level := ""
	if attributes != nil {
		level, _ = attributes["level"].(string)
	}
	
	// Also check for error field
	errMsg := ""
	if attributes != nil {
		errMsg, _ = attributes["error"].(string)
	}
	
	// Check if this is an error or warning
	isError := level == "error" || level == "warn" || errMsg != "" || 
		strings.Contains(strings.ToLower(message), "error") ||
		strings.Contains(strings.ToLower(message), "failed") ||
		strings.Contains(strings.ToLower(message), "❌") ||
		strings.Contains(strings.ToLower(message), "panic")
	
	if isError {
		errorPattern := ErrorPattern{
			RequestID: extractRequestID(logEntry),
			Timestamp: extractTimestamp(logEntry),
			Level:     level,
			Message:   message,
			Error:     errMsg,
			Context:   logEntry,
		}
		
		// Categorize the error
		errorPattern.Category = categorizeError(errorPattern, dnsPattern, timeoutPattern, 
			http4xxPattern, http5xxPattern, networkPattern, contextPattern, parsePattern)
		
		*errors = append(*errors, errorPattern)
	}
}

func categorizeError(err ErrorPattern, dnsPattern, timeoutPattern, http4xxPattern, 
	http5xxPattern, networkPattern, contextPattern, parsePattern *regexp.Regexp) string {
	
	combinedText := strings.ToLower(err.Error + " " + err.Message)
	
	if dnsPattern.MatchString(combinedText) {
		return "dns_failure"
	}
	if timeoutPattern.MatchString(combinedText) {
		return "timeout"
	}
	if contextPattern.MatchString(combinedText) {
		return "context_cancellation"
	}
	if http5xxPattern.MatchString(combinedText) {
		return "http_5xx"
	}
	if http4xxPattern.MatchString(combinedText) {
		return "http_4xx"
	}
	if networkPattern.MatchString(combinedText) {
		return "network_error"
	}
	if parsePattern.MatchString(combinedText) {
		return "parse_error"
	}
	
	// Check for specific error types in context
	if err.Context != nil {
		if errorType, ok := err.Context["error_type"].(string); ok {
			return strings.ToLower(errorType)
		}
		if statusCode, ok := err.Context["status_code"].(float64); ok {
			code := int(statusCode)
			if code >= 500 {
				return "http_5xx"
			}
			if code >= 400 {
				return "http_4xx"
			}
		}
	}
	
	return "other"
}

func extractRequestID(logEntry map[string]interface{}) string {
	if rid, ok := logEntry["request_id"].(string); ok {
		return rid
	}
	if rid, ok := logEntry["requestId"].(string); ok {
		return rid
	}
	return ""
}

func extractTimestamp(logEntry map[string]interface{}) string {
	if ts, ok := logEntry["ts"].(float64); ok {
		// Convert Unix timestamp to readable format
		return fmt.Sprintf("%.0f", ts)
	}
	if ts, ok := logEntry["timestamp"].(string); ok {
		return ts
	}
	if ts, ok := logEntry["time"].(string); ok {
		return ts
	}
	return ""
}

func extractErrorFromText(line string, lineNum int) *ErrorPattern {
	// Try to extract error information from plain text
	lowerLine := strings.ToLower(line)
	if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "failed") {
		return &ErrorPattern{
			Message:  line,
			Category: "unknown",
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func writeDetailedReport(categories map[string]*ErrorCategory, errors []ErrorPattern) {
	reportFile := "docs/error-pattern-analysis.md"
	file, err := os.Create(reportFile)
	if err != nil {
		fmt.Printf("Warning: Could not create report file: %v\n", err)
		return
	}
	defer file.Close()
	
	fmt.Fprintf(file, "# Error Pattern Analysis Report\n\n")
	fmt.Fprintf(file, "Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "Total errors analyzed: %d\n\n", len(errors))
	
	// Sort categories by count
	type catStat struct {
		name  string
		count int
	}
	catStats := []catStat{}
	for name, cat := range categories {
		catStats = append(catStats, catStat{name, cat.Count})
	}
	sort.Slice(catStats, func(i, j int) bool {
		return catStats[i].count > catStats[j].count
	})
	
	fmt.Fprintf(file, "## Error Distribution\n\n")
	fmt.Fprintf(file, "| Category | Count | Percentage |\n")
	fmt.Fprintf(file, "|----------|-------|------------|\n")
	total := len(errors)
	for _, stat := range catStats {
		cat := categories[stat.name]
		percentage := float64(cat.Count) / float64(total) * 100
		fmt.Fprintf(file, "| %s | %d | %.1f%% |\n", cat.Name, cat.Count, percentage)
	}
	
	fmt.Fprintf(file, "\n## Category Details\n\n")
	for _, stat := range catStats {
		cat := categories[stat.name]
		fmt.Fprintf(file, "### %s (%d errors)\n\n", cat.Name, cat.Count)
		
		if len(cat.Examples) > 0 {
			fmt.Fprintf(file, "**Examples:**\n\n")
			for _, example := range cat.Examples {
				fmt.Fprintf(file, "- %s\n", example)
			}
			fmt.Fprintf(file, "\n")
		}
		
		if len(cat.RequestIDs) > 0 {
			fmt.Fprintf(file, "**Sample Request IDs:**\n\n")
			for _, reqID := range cat.RequestIDs {
				fmt.Fprintf(file, "- %s\n", reqID)
			}
			fmt.Fprintf(file, "\n")
		}
	}
	
	fmt.Printf("\nDetailed report written to: %s\n", reportFile)
}

