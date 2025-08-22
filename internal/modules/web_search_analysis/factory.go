package web_search_analysis

import (
	"github.com/pcraw4d/business-verification/internal/architecture"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"

	"go.opentelemetry.io/otel/trace"
)

// WebSearchAnalysisFactory implements ModuleFactory for web search analysis modules
type WebSearchAnalysisFactory struct {
	config  *config.Config
	db      database.Database
	logger  *observability.Logger
	metrics *observability.Metrics
	tracer  trace.Tracer
}

// NewWebSearchAnalysisFactory creates a new web search analysis factory
func NewWebSearchAnalysisFactory(
	config *config.Config,
	db database.Database,
	logger *observability.Logger,
	metrics *observability.Metrics,
	tracer trace.Tracer,
) *WebSearchAnalysisFactory {
	return &WebSearchAnalysisFactory{
		config:  config,
		db:      db,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// CreateModule creates a new web search analysis module
func (f *WebSearchAnalysisFactory) CreateModule(config architecture.ModuleConfig) (architecture.Module, error) {
	module := NewWebSearchAnalysisModule()

	// Set module configuration
	module.config = config

	// Inject dependencies
	module.logger = f.logger
	module.metrics = f.metrics
	module.tracer = f.tracer
	module.db = f.db
	module.appConfig = f.config

	return module, nil
}

// GetModuleType returns the module type
func (f *WebSearchAnalysisFactory) GetModuleType() string {
	return "web_search_analysis"
}
