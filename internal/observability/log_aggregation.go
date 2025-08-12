package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogAggregationSystem provides centralized log aggregation capabilities
type LogAggregationSystem struct {
	logger        *zap.Logger
	elasticClient *elasticsearch.Client
	config        *LogAggregationConfig
	shutdownChan  chan struct{}
}

// LogAggregationConfig holds configuration for log aggregation
type LogAggregationConfig struct {
	// Elasticsearch configuration
	ElasticsearchURL      string
	ElasticsearchUsername string
	ElasticsearchPassword string
	ElasticsearchIndex    string

	// Log shipping configuration
	BatchSize     int
	BatchTimeout  time.Duration
	RetryAttempts int
	RetryDelay    time.Duration

	// Log level configuration
	MinLevel    zapcore.Level
	Environment string
	Application string
	Version     string

	// Output configuration
	EnableConsole bool
	EnableFile    bool
	EnableElastic bool
	LogFilePath   string

	// Buffer configuration
	BufferSize    int
	FlushInterval time.Duration
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Logger      string                 `json:"logger"`
	Environment string                 `json:"environment"`
	Application string                 `json:"application"`
	Version     string                 `json:"version"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Endpoint    string                 `json:"endpoint,omitempty"`
	Method      string                 `json:"method,omitempty"`
	StatusCode  int                    `json:"status_code,omitempty"`
	Duration    float64                `json:"duration,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Stack       string                 `json:"stack,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// LogShipper handles log shipping to external systems
type LogShipper struct {
	elasticClient *elasticsearch.Client
	config        *LogAggregationConfig
	logChan       chan LogEntry
	shutdownChan  chan struct{}
	logger        *zap.Logger
}

// NewLogAggregationSystem creates a new log aggregation system
func NewLogAggregationSystem(config *LogAggregationConfig) (*LogAggregationSystem, error) {
	las := &LogAggregationSystem{
		config:       config,
		shutdownChan: make(chan struct{}),
	}

	// Initialize logger
	if err := las.initializeLogger(); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize Elasticsearch client if enabled
	if config.EnableElastic {
		if err := las.initializeElasticsearch(); err != nil {
			return nil, fmt.Errorf("failed to initialize Elasticsearch: %w", err)
		}
	}

	return las, nil
}

// initializeLogger sets up the structured logger
func (las *LogAggregationSystem) initializeLogger() error {
	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Create core options
	var cores []zapcore.Core

	// Console output
	if las.config.EnableConsole {
		consoleEncoder := zapcore.NewJSONEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			las.config.MinLevel,
		)
		cores = append(cores, consoleCore)
	}

	// File output
	if las.config.EnableFile {
		file, err := os.OpenFile(las.config.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		fileCore := zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(file),
			las.config.MinLevel,
		)
		cores = append(cores, fileCore)
	}

	// Create logger
	if len(cores) == 0 {
		// Fallback to console if no output is configured
		consoleEncoder := zapcore.NewJSONEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			las.config.MinLevel,
		)
		cores = append(cores, consoleCore)
	}

	core := zapcore.NewTee(cores...)
	las.logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// initializeElasticsearch sets up Elasticsearch client
func (las *LogAggregationSystem) initializeElasticsearch() error {
	cfg := elasticsearch.Config{
		Addresses: []string{las.config.ElasticsearchURL},
		Username:  las.config.ElasticsearchUsername,
		Password:  las.config.ElasticsearchPassword,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Test connection
	res, err := client.Info()
	if err != nil {
		return fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	las.elasticClient = client
	las.logger.Info("Elasticsearch connection established")

	return nil
}

// GetLogger returns the configured logger
func (las *LogAggregationSystem) GetLogger() *zap.Logger {
	return las.logger
}

// CreateLogShipper creates a new log shipper for external systems
func (las *LogAggregationSystem) CreateLogShipper() *LogShipper {
	shipper := &LogShipper{
		elasticClient: las.elasticClient,
		config:        las.config,
		logChan:       make(chan LogEntry, las.config.BufferSize),
		shutdownChan:  make(chan struct{}),
		logger:        las.logger,
	}

	// Start log shipping goroutine
	go shipper.startLogShipping()

	return shipper
}

// startLogShipping handles log shipping to external systems
func (shipper *LogShipper) startLogShipping() {
	ticker := time.NewTicker(shipper.config.FlushInterval)
	defer ticker.Stop()

	var batch []LogEntry

	for {
		select {
		case <-shipper.shutdownChan:
			// Flush remaining logs
			if len(batch) > 0 {
				shipper.shipBatch(batch)
			}
			return

		case logEntry := <-shipper.logChan:
			batch = append(batch, logEntry)

			// Ship batch if it reaches the size limit
			if len(batch) >= shipper.config.BatchSize {
				shipper.shipBatch(batch)
				batch = batch[:0] // Reset batch
			}

		case <-ticker.C:
			// Flush batch on timer
			if len(batch) > 0 {
				shipper.shipBatch(batch)
				batch = batch[:0] // Reset batch
			}
		}
	}
}

// shipBatch ships a batch of log entries to Elasticsearch
func (shipper *LogShipper) shipBatch(batch []LogEntry) {
	if shipper.elasticClient == nil {
		return
	}

	// Create bulk request body
	var bulkBody strings.Builder
	for _, entry := range batch {
		// Add index action
		indexAction := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": shipper.config.ElasticsearchIndex,
			},
		}
		indexActionBytes, _ := json.Marshal(indexAction)
		bulkBody.WriteString(string(indexActionBytes) + "\n")

		// Add document
		docBytes, _ := json.Marshal(entry)
		bulkBody.WriteString(string(docBytes) + "\n")
	}

	// Execute bulk request with retries
	for attempt := 0; attempt < shipper.config.RetryAttempts; attempt++ {
		res, err := shipper.elasticClient.Bulk(
			strings.NewReader(bulkBody.String()),
			shipper.elasticClient.Bulk.WithIndex(shipper.config.ElasticsearchIndex),
		)
		if err == nil && !res.IsError() {
			shipper.logger.Debug("Successfully shipped log batch", zap.Int("count", len(batch)))
			return
		}

		if err != nil {
			shipper.logger.Error("Failed to ship log batch", zap.Error(err), zap.Int("attempt", attempt+1))
		} else {
			shipper.logger.Error("Bulk request had errors", zap.Int("attempt", attempt+1))
		}

		// Wait before retry
		if attempt < shipper.config.RetryAttempts-1 {
			time.Sleep(shipper.config.RetryDelay)
		}
	}

	shipper.logger.Error("Failed to ship log batch after all retries", zap.Int("count", len(batch)))
}

// ShipLog sends a log entry to the log shipper
func (shipper *LogShipper) ShipLog(entry LogEntry) {
	select {
	case shipper.logChan <- entry:
		// Log entry queued successfully
	default:
		// Channel is full, log locally
		shipper.logger.Warn("Log shipper buffer full, dropping log entry", zap.String("message", entry.Message))
	}
}

// Shutdown gracefully shuts down the log aggregation system
func (las *LogAggregationSystem) Shutdown() {
	close(las.shutdownChan)
	if las.logger != nil {
		las.logger.Sync()
	}
}

// CreateStructuredLogger creates a structured logger with correlation IDs
func (las *LogAggregationSystem) CreateStructuredLogger(ctx context.Context) *zap.Logger {
	logger := las.logger

	// Add correlation IDs from context
	if traceID := GetTraceIDFromContext(ctx); traceID != "" {
		logger = logger.With(zap.String("trace_id", traceID))
	}

	if spanID := GetSpanIDFromContext(ctx); spanID != "" {
		logger = logger.With(zap.String("span_id", spanID))
	}

	if requestID := GetRequestIDFromContext(ctx); requestID != "" {
		logger = logger.With(zap.String("request_id", requestID))
	}

	if userID := GetUserIDFromContext(ctx); userID != "" {
		logger = logger.With(zap.String("user_id", userID))
	}

	return logger
}

// LogHTTPRequest logs HTTP request details
func (las *LogAggregationSystem) LogHTTPRequest(ctx context.Context, r *http.Request, statusCode int, duration time.Duration) {
	logger := las.CreateStructuredLogger(ctx)

	logger.Info("HTTP request completed",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("query", r.URL.RawQuery),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
		zap.String("user_agent", r.UserAgent()),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("referer", r.Referer()),
		zap.String("content_type", r.Header.Get("Content-Type")),
		zap.String("accept", r.Header.Get("Accept")),
	)
}

// LogError logs error details with stack trace
func (las *LogAggregationSystem) LogError(ctx context.Context, err error, message string, fields ...zap.Field) {
	logger := las.CreateStructuredLogger(ctx)

	allFields := append([]zap.Field{
		zap.Error(err),
		zap.String("error_type", fmt.Sprintf("%T", err)),
	}, fields...)

	logger.Error(message, allFields...)
}

// LogBusinessEvent logs business-specific events
func (las *LogAggregationSystem) LogBusinessEvent(ctx context.Context, eventType, eventName string, data map[string]interface{}) {
	logger := las.CreateStructuredLogger(ctx)

	fields := []zap.Field{
		zap.String("event_type", eventType),
		zap.String("event_name", eventName),
	}

	// Add data fields
	for key, value := range data {
		fields = append(fields, zap.Any(key, value))
	}

	logger.Info("Business event", fields...)
}

// LogSecurityEvent logs security-related events
func (las *LogAggregationSystem) LogSecurityEvent(ctx context.Context, eventType, eventName string, severity string, data map[string]interface{}) {
	logger := las.CreateStructuredLogger(ctx)

	fields := []zap.Field{
		zap.String("event_type", eventType),
		zap.String("event_name", eventName),
		zap.String("severity", severity),
		zap.String("category", "security"),
	}

	// Add data fields
	for key, value := range data {
		fields = append(fields, zap.Any(key, value))
	}

	logger.Warn("Security event", fields...)
}

// LogPerformanceEvent logs performance-related events
func (las *LogAggregationSystem) LogPerformanceEvent(ctx context.Context, operation string, duration time.Duration, data map[string]interface{}) {
	logger := las.CreateStructuredLogger(ctx)

	fields := []zap.Field{
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.String("category", "performance"),
	}

	// Add data fields
	for key, value := range data {
		fields = append(fields, zap.Any(key, value))
	}

	logger.Info("Performance event", fields...)
}

// LogDatabaseEvent logs database-related events
func (las *LogAggregationSystem) LogDatabaseEvent(ctx context.Context, operation, query string, duration time.Duration, data map[string]interface{}) {
	logger := las.CreateStructuredLogger(ctx)

	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("query", query),
		zap.Duration("duration", duration),
		zap.String("category", "database"),
	}

	// Add data fields
	for key, value := range data {
		fields = append(fields, zap.Any(key, value))
	}

	logger.Info("Database event", fields...)
}

// LogExternalAPIEvent logs external API calls
func (las *LogAggregationSystem) LogExternalAPIEvent(ctx context.Context, provider, endpoint, method string, statusCode int, duration time.Duration, data map[string]interface{}) {
	logger := las.CreateStructuredLogger(ctx)

	fields := []zap.Field{
		zap.String("provider", provider),
		zap.String("endpoint", endpoint),
		zap.String("method", method),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
		zap.String("category", "external_api"),
	}

	// Add data fields
	for key, value := range data {
		fields = append(fields, zap.Any(key, value))
	}

	logger.Info("External API event", fields...)
}

// SearchLogs searches logs in Elasticsearch
func (las *LogAggregationSystem) SearchLogs(ctx context.Context, query map[string]interface{}, from, size int) ([]LogEntry, error) {
	if las.elasticClient == nil {
		return nil, fmt.Errorf("Elasticsearch not configured")
	}

	// Create search request
	searchBody, err := json.Marshal(map[string]interface{}{
		"query": query,
		"from":  from,
		"size":  size,
		"sort": []map[string]interface{}{
			{"timestamp": map[string]interface{}{"order": "desc"}},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}

	// Execute search
	res, err := las.elasticClient.Search(
		las.elasticClient.Search.WithContext(ctx),
		las.elasticClient.Search.WithIndex(las.config.ElasticsearchIndex),
		las.elasticClient.Search.WithBody(strings.NewReader(string(searchBody))),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	// Parse response
	var searchResponse map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// Extract hits
	hits, ok := searchResponse["hits"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid search response format")
	}

	hitsArray, ok := hits["hits"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid hits format")
	}

	// Parse log entries
	var entries []LogEntry
	for _, hit := range hitsArray {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		source, ok := hitMap["_source"].(map[string]interface{})
		if !ok {
			continue
		}

		// Convert source to LogEntry
		entryBytes, err := json.Marshal(source)
		if err != nil {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal(entryBytes, &entry); err != nil {
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// GetLogStatistics gets log statistics from Elasticsearch
func (las *LogAggregationSystem) GetLogStatistics(ctx context.Context, timeRange time.Duration) (map[string]interface{}, error) {
	if las.elasticClient == nil {
		return nil, fmt.Errorf("Elasticsearch not configured")
	}

	// Create aggregation query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"gte": "now-" + timeRange.String(),
				},
			},
		},
		"aggs": map[string]interface{}{
			"log_levels": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "level",
				},
			},
			"endpoints": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "endpoint",
					"size":  10,
				},
			},
			"error_rate": map[string]interface{}{
				"filter": map[string]interface{}{
					"term": map[string]interface{}{
						"level": "ERROR",
					},
				},
			},
		},
	}

	searchBody, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	// Execute search
	res, err := las.elasticClient.Search(
		las.elasticClient.Search.WithContext(ctx),
		las.elasticClient.Search.WithIndex(las.config.ElasticsearchIndex),
		las.elasticClient.Search.WithBody(strings.NewReader(string(searchBody))),
		las.elasticClient.Search.WithSize(0),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	// Parse response
	var searchResponse map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return searchResponse, nil
}

// CreateLogEntry creates a structured log entry
func (las *LogAggregationSystem) CreateLogEntry(ctx context.Context, level, message string, fields map[string]interface{}) LogEntry {
	entry := LogEntry{
		Timestamp:   time.Now().UTC(),
		Level:       level,
		Message:     message,
		Logger:      "kyb-platform",
		Environment: las.config.Environment,
		Application: las.config.Application,
		Version:     las.config.Version,
		Fields:      fields,
		Metadata:    make(map[string]interface{}),
	}

	// Add correlation IDs
	if traceID := GetTraceIDFromContext(ctx); traceID != "" {
		entry.TraceID = traceID
	}

	if spanID := GetSpanIDFromContext(ctx); spanID != "" {
		entry.SpanID = spanID
	}

	if requestID := GetRequestIDFromContext(ctx); requestID != "" {
		entry.RequestID = requestID
	}

	if userID := GetUserIDFromContext(ctx); userID != "" {
		entry.UserID = userID
	}

	return entry
}

// Context helper functions for correlation IDs

// GetTraceIDFromContext extracts trace ID from context
func GetTraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return traceID
	}
	return ""
}

// GetSpanIDFromContext extracts span ID from context
func GetSpanIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if spanID, ok := ctx.Value("span_id").(string); ok {
		return spanID
	}
	return ""
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}
