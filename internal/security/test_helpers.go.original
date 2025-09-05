package security

import (
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// createTestLogger creates a logger for testing
func createTestLogger() *observability.Logger {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "text",
	}
	return observability.NewLogger(cfg)
}
