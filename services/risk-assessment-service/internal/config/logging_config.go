package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggingConfig holds configuration for structured logging
type LoggingConfig struct {
	Level        string `json:"level" yaml:"level"`
	Format       string `json:"format" yaml:"format"`
	Output       string `json:"output" yaml:"output"`
	EnableTrace  bool   `json:"enable_trace" yaml:"enable_trace"`
	EnableCaller bool   `json:"enable_caller" yaml:"enable_caller"`
	EnableStack  bool   `json:"enable_stack" yaml:"enable_stack"`

	// File logging configuration
	FileConfig FileLoggingConfig `json:"file_config" yaml:"file_config"`

	// Structured logging fields
	ServiceName    string `json:"service_name" yaml:"service_name"`
	ServiceVersion string `json:"service_version" yaml:"service_version"`
	Environment    string `json:"environment" yaml:"environment"`

	// Correlation ID configuration
	CorrelationIDHeader string `json:"correlation_id_header" yaml:"correlation_id_header"`
	RequestIDHeader     string `json:"request_id_header" yaml:"request_id_header"`

	// Sampling configuration
	SamplingConfig SamplingConfig `json:"sampling_config" yaml:"sampling_config"`
}

// FileLoggingConfig holds file-specific logging configuration
type FileLoggingConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	Path       string `json:"path" yaml:"path"`
	MaxSize    int    `json:"max_size" yaml:"max_size"` // MB
	MaxBackups int    `json:"max_backups" yaml:"max_backups"`
	MaxAge     int    `json:"max_age" yaml:"max_age"` // days
	Compress   bool   `json:"compress" yaml:"compress"`
	LocalTime  bool   `json:"local_time" yaml:"local_time"`
}

// SamplingConfig holds sampling configuration for high-volume logging
type SamplingConfig struct {
	Enabled    bool          `json:"enabled" yaml:"enabled"`
	Initial    int           `json:"initial" yaml:"initial"`
	Thereafter int           `json:"thereafter" yaml:"thereafter"`
	Tick       time.Duration `json:"tick" yaml:"tick"`
}

// LoadLoggingConfig loads logging configuration from environment variables
func LoadLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		Level:        getEnvAsString("LOG_LEVEL", "info"),
		Format:       getEnvAsString("LOG_FORMAT", "json"),
		Output:       getEnvAsString("LOG_OUTPUT", "stdout"),
		EnableTrace:  getEnvAsBool("LOG_ENABLE_TRACE", true),
		EnableCaller: getEnvAsBool("LOG_ENABLE_CALLER", false),
		EnableStack:  getEnvAsBool("LOG_ENABLE_STACK", false),

		FileConfig: FileLoggingConfig{
			Enabled:    getEnvAsBool("LOG_FILE_ENABLED", false),
			Path:       getEnvAsString("LOG_FILE_PATH", "/var/log/risk-assessment-service.log"),
			MaxSize:    getEnvAsInt("LOG_FILE_MAX_SIZE", 100),
			MaxBackups: getEnvAsInt("LOG_FILE_MAX_BACKUPS", 3),
			MaxAge:     getEnvAsInt("LOG_FILE_MAX_AGE", 7),
			Compress:   getEnvAsBool("LOG_FILE_COMPRESS", true),
			LocalTime:  getEnvAsBool("LOG_FILE_LOCAL_TIME", true),
		},

		ServiceName:    getEnvAsString("SERVICE_NAME", "risk-assessment-service"),
		ServiceVersion: getEnvAsString("SERVICE_VERSION", "1.0.0"),
		Environment:    getEnvAsString("ENVIRONMENT", "development"),

		CorrelationIDHeader: getEnvAsString("CORRELATION_ID_HEADER", "X-Correlation-ID"),
		RequestIDHeader:     getEnvAsString("REQUEST_ID_HEADER", "X-Request-ID"),

		SamplingConfig: SamplingConfig{
			Enabled:    getEnvAsBool("LOG_SAMPLING_ENABLED", false),
			Initial:    getEnvAsInt("LOG_SAMPLING_INITIAL", 100),
			Thereafter: getEnvAsInt("LOG_SAMPLING_THEREAFTER", 100),
			Tick:       getEnvAsDuration("LOG_SAMPLING_TICK", 1*time.Second),
		},
	}
}

// CreateLogger creates a configured zap logger based on the logging configuration
func (lc *LoggingConfig) CreateLogger() (*zap.Logger, error) {
	// Set up encoder configuration
	var encoderConfig zapcore.EncoderConfig
	if lc.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	// Customize encoder configuration for structured logging
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.LevelKey = "level"
	encoderConfig.NameKey = "logger"
	encoderConfig.CallerKey = "caller"
	encoderConfig.MessageKey = "message"
	encoderConfig.StacktraceKey = "stacktrace"
	encoderConfig.LineEnding = zapcore.DefaultLineEnding
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Set up encoder
	var encoder zapcore.Encoder
	if lc.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Set up log level
	level, err := zapcore.ParseLevel(lc.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	// Set up output
	var writeSyncer zapcore.WriteSyncer
	if lc.Output == "stdout" {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else if lc.Output == "stderr" {
		writeSyncer = zapcore.AddSync(os.Stderr)
	} else {
		// File output
		file, err := os.OpenFile(lc.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		writeSyncer = zapcore.AddSync(file)
	}

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// Set up options
	var options []zap.Option

	// Add caller information
	if lc.EnableCaller {
		options = append(options, zap.AddCaller())
	}

	// Add stack trace
	if lc.EnableStack {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// Add sampling if enabled
	if lc.SamplingConfig.Enabled {
		options = append(options, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(
				core,
				lc.SamplingConfig.Tick,
				lc.SamplingConfig.Initial,
				lc.SamplingConfig.Thereafter,
			)
		}))
	}

	// Create logger
	logger := zap.New(core, options...)

	// Add default fields
	logger = logger.With(
		zap.String("service", lc.ServiceName),
		zap.String("version", lc.ServiceVersion),
		zap.String("environment", lc.Environment),
		zap.String("hostname", getHostname()),
		zap.String("pid", fmt.Sprintf("%d", os.Getpid())),
	)

	return logger, nil
}

// CreateRequestLogger creates a logger with request-specific context
func (lc *LoggingConfig) CreateRequestLogger(baseLogger *zap.Logger, correlationID, requestID, userID, tenantID string) *zap.Logger {
	fields := []zap.Field{
		zap.String("correlation_id", correlationID),
		zap.String("request_id", requestID),
	}

	if userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	if tenantID != "" {
		fields = append(fields, zap.String("tenant_id", tenantID))
	}

	return baseLogger.With(fields...)
}

// CreateServiceLogger creates a logger with service-specific context
func (lc *LoggingConfig) CreateServiceLogger(baseLogger *zap.Logger, serviceName, operation string) *zap.Logger {
	return baseLogger.With(
		zap.String("service", serviceName),
		zap.String("operation", operation),
	)
}

// CreateErrorLogger creates a logger with error-specific context
func (lc *LoggingConfig) CreateErrorLogger(baseLogger *zap.Logger, err error, context map[string]interface{}) *zap.Logger {
	fields := []zap.Field{
		zap.Error(err),
	}

	for key, value := range context {
		fields = append(fields, zap.Any(key, value))
	}

	return baseLogger.With(fields...)
}

// LogStructuredData logs structured data as JSON
func (lc *LoggingConfig) LogStructuredData(logger *zap.Logger, level zapcore.Level, message string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Error("Failed to marshal structured data", zap.Error(err))
		return
	}

	var structuredData map[string]interface{}
	if err := json.Unmarshal(jsonData, &structuredData); err != nil {
		logger.Error("Failed to unmarshal structured data", zap.Error(err))
		return
	}

	fields := make([]zap.Field, 0, len(structuredData))
	for key, value := range structuredData {
		fields = append(fields, zap.Any(key, value))
	}

	switch level {
	case zapcore.DebugLevel:
		logger.Debug(message, fields...)
	case zapcore.InfoLevel:
		logger.Info(message, fields...)
	case zapcore.WarnLevel:
		logger.Warn(message, fields...)
	case zapcore.ErrorLevel:
		logger.Error(message, fields...)
	case zapcore.FatalLevel:
		logger.Fatal(message, fields...)
	case zapcore.PanicLevel:
		logger.Panic(message, fields...)
	}
}

// Validate validates the logging configuration
func (lc *LoggingConfig) Validate() error {
	// Validate log level
	_, err := zapcore.ParseLevel(lc.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	// Validate log format
	if lc.Format != "json" && lc.Format != "console" {
		return fmt.Errorf("invalid log format: %s (must be 'json' or 'console')", lc.Format)
	}

	// Validate log output
	if lc.Output != "stdout" && lc.Output != "stderr" && !strings.HasPrefix(lc.Output, "/") {
		return fmt.Errorf("invalid log output: %s (must be 'stdout', 'stderr', or file path)", lc.Output)
	}

	// Validate file configuration if file logging is enabled
	if lc.FileConfig.Enabled {
		if lc.FileConfig.Path == "" {
			return fmt.Errorf("log file path cannot be empty when file logging is enabled")
		}

		if lc.FileConfig.MaxSize <= 0 {
			return fmt.Errorf("log file max size must be positive")
		}

		if lc.FileConfig.MaxBackups < 0 {
			return fmt.Errorf("log file max backups cannot be negative")
		}

		if lc.FileConfig.MaxAge < 0 {
			return fmt.Errorf("log file max age cannot be negative")
		}
	}

	// Validate sampling configuration
	if lc.SamplingConfig.Enabled {
		if lc.SamplingConfig.Initial <= 0 {
			return fmt.Errorf("sampling initial count must be positive")
		}

		if lc.SamplingConfig.Thereafter <= 0 {
			return fmt.Errorf("sampling thereafter count must be positive")
		}

		if lc.SamplingConfig.Tick <= 0 {
			return fmt.Errorf("sampling tick duration must be positive")
		}
	}

	return nil
}

// getHostname returns the hostname of the current machine
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
