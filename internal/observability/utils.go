package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateRequestID generates a unique request ID
func GenerateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// MonitoringSystem represents a monitoring system interface
type MonitoringSystem interface {
	RecordMetric(name string, value float64, tags map[string]string)
	IncrementCounter(name string, tags map[string]string)
	SetGauge(name string, value float64, tags map[string]string)
}

// ErrorTrackingSystem represents an error tracking system interface
type ErrorTrackingSystem interface {
	CaptureError(err error, tags map[string]string)
	CaptureMessage(message string, level string, tags map[string]string)
}

// DefaultMonitoringSystem is a basic implementation of MonitoringSystem
type DefaultMonitoringSystem struct{}

// NewMonitoringSystem creates a new monitoring system
func NewMonitoringSystem() MonitoringSystem {
	return &DefaultMonitoringSystem{}
}

// RecordMetric records a metric
func (m *DefaultMonitoringSystem) RecordMetric(name string, value float64, tags map[string]string) {
	// In a real implementation, this would send to a monitoring system
	fmt.Printf("Metric: %s = %f, tags: %v\n", name, value, tags)
}

// IncrementCounter increments a counter
func (m *DefaultMonitoringSystem) IncrementCounter(name string, tags map[string]string) {
	// In a real implementation, this would increment a counter
	fmt.Printf("Counter: %s++, tags: %v\n", name, tags)
}

// SetGauge sets a gauge value
func (m *DefaultMonitoringSystem) SetGauge(name string, value float64, tags map[string]string) {
	// In a real implementation, this would set a gauge
	fmt.Printf("Gauge: %s = %f, tags: %v\n", name, value, tags)
}

// DefaultErrorTrackingSystem is a basic implementation of ErrorTrackingSystem
type DefaultErrorTrackingSystem struct{}

// NewErrorTrackingSystem creates a new error tracking system
func NewErrorTrackingSystem() ErrorTrackingSystem {
	return &DefaultErrorTrackingSystem{}
}

// CaptureError captures an error
func (e *DefaultErrorTrackingSystem) CaptureError(err error, tags map[string]string) {
	// In a real implementation, this would send to an error tracking system
	fmt.Printf("Error: %v, tags: %v\n", err, tags)
}

// CaptureMessage captures a message
func (e *DefaultErrorTrackingSystem) CaptureMessage(message string, level string, tags map[string]string) {
	// In a real implementation, this would send to an error tracking system
	fmt.Printf("Message [%s]: %s, tags: %v\n", level, message, tags)
}

// Tracer represents a tracing interface
type Tracer interface {
	StartSpan(name string) Span
	StartSpanWithContext(name string, parent Span) Span
	Start(ctx context.Context, name string) (context.Context, Span)
}

// Span represents a tracing span
type Span interface {
	SetTag(key string, value interface{})
	SetError(err error)
	Finish()
	End()
	Context() interface{}
}

// DefaultSpan is a basic implementation of Span
type DefaultSpan struct {
	name    string
	start   time.Time
	tags    map[string]interface{}
	context interface{}
}

// NewSpan creates a new span
func NewSpan(name string) Span {
	return &DefaultSpan{
		name:  name,
		start: time.Now(),
		tags:  make(map[string]interface{}),
	}
}

// SetTag sets a tag on the span
func (s *DefaultSpan) SetTag(key string, value interface{}) {
	s.tags[key] = value
}

// SetError sets an error on the span
func (s *DefaultSpan) SetError(err error) {
	s.tags["error"] = err.Error()
}

// Finish finishes the span
func (s *DefaultSpan) Finish() {
	duration := time.Since(s.start)
	fmt.Printf("Span [%s] finished in %v, tags: %v\n", s.name, duration, s.tags)
}

// End finishes the span (alias for Finish for compatibility)
func (s *DefaultSpan) End() {
	s.Finish()
}

// Context returns the span context
func (s *DefaultSpan) Context() interface{} {
	return s.context
}

// DefaultTracer is a basic implementation of Tracer
type DefaultTracer struct{}

// NewTracer creates a new tracer
func NewTracer() Tracer {
	return &DefaultTracer{}
}

// StartSpan starts a new span
func (t *DefaultTracer) StartSpan(name string) Span {
	return NewSpan(name)
}

// StartSpanWithContext starts a new span with parent context
func (t *DefaultTracer) StartSpanWithContext(name string, parent Span) Span {
	span := NewSpan(name)
	// In a real implementation, this would set up proper parent-child relationship
	return span
}

// Start starts a new span with context
func (t *DefaultTracer) Start(ctx context.Context, name string) (context.Context, Span) {
	span := t.StartSpan(name)
	// In a real implementation, this would add the span to the context
	return ctx, span
}
