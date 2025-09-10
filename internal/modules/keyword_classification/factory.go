package keyword_classification

import (
	"github.com/pcraw4d/business-verification/internal/architecture"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"

	"go.opentelemetry.io/otel/trace"
)

// KeywordClassificationFactory implements ModuleFactory for keyword classification modules
type KeywordClassificationFactory struct {
	config  *config.Config
	db      database.Database
	logger  *observability.Logger
	metrics *observability.Metrics
	tracer  trace.Tracer
}

// NewKeywordClassificationFactory creates a new keyword classification factory
func NewKeywordClassificationFactory(
	config *config.Config,
	db database.Database,
	logger *observability.Logger,
	metrics *observability.Metrics,
	tracer trace.Tracer,
) *KeywordClassificationFactory {
	return &KeywordClassificationFactory{
		config:  config,
		db:      db,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// CreateModule creates a new keyword classification module
func (f *KeywordClassificationFactory) CreateModule(config architecture.ModuleConfig) (architecture.Module, error) {
	// TODO: For now, create a basic implementation to get compilation working
	// This will need to be properly implemented with the correct database interface

	// Create module with nil repository (will fail at runtime but allows compilation)
	module := NewKeywordClassificationModule(nil)

	// Set module configuration
	module.config = config

	// Create module logger
	moduleLogger := observability.NewModuleLogger(f.logger.GetZapLogger(), module.ID())

	// Inject dependencies
	module.logger = moduleLogger
	module.metrics = f.metrics
	module.tracer = f.tracer
	module.db = f.db
	module.appConfig = f.config

	return module, nil
}

// GetModuleType returns the module type
func (f *KeywordClassificationFactory) GetModuleType() string {
	return "keyword_classification"
}
