package ml_classification

import (
	"kyb-platform/internal/architecture"
	"kyb-platform/internal/config"
	"kyb-platform/internal/database"
	"kyb-platform/internal/observability"

	"go.opentelemetry.io/otel/trace"
)

// MLClassificationFactory implements ModuleFactory for ML classification modules
type MLClassificationFactory struct {
	config  *config.Config
	db      database.Database
	logger  *observability.Logger
	metrics *observability.Metrics
	tracer  trace.Tracer
}

// NewMLClassificationFactory creates a new ML classification factory
func NewMLClassificationFactory(
	config *config.Config,
	db database.Database,
	logger *observability.Logger,
	metrics *observability.Metrics,
	tracer trace.Tracer,
) *MLClassificationFactory {
	return &MLClassificationFactory{
		config:  config,
		db:      db,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// CreateModule creates a new ML classification module
func (f *MLClassificationFactory) CreateModule(config architecture.ModuleConfig) (architecture.Module, error) {
	module := NewMLClassificationModule()

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
func (f *MLClassificationFactory) GetModuleType() string {
	return "ml_classification"
}
