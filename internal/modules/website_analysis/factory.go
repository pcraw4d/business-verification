package website_analysis

import (
	"github.com/pcraw4d/business-verification/internal/architecture"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"

	"go.opentelemetry.io/otel/trace"
)

// WebsiteAnalysisFactory implements ModuleFactory for website analysis modules
type WebsiteAnalysisFactory struct {
	config  *config.Config
	db      database.Database
	logger  *observability.Logger
	metrics *observability.Metrics
	tracer  trace.Tracer
}

// NewWebsiteAnalysisFactory creates a new website analysis factory
func NewWebsiteAnalysisFactory(
	config *config.Config,
	db database.Database,
	logger *observability.Logger,
	metrics *observability.Metrics,
	tracer trace.Tracer,
) *WebsiteAnalysisFactory {
	return &WebsiteAnalysisFactory{
		config:  config,
		db:      db,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// CreateModule creates a new website analysis module
func (f *WebsiteAnalysisFactory) CreateModule(config architecture.ModuleConfig) (architecture.Module, error) {
	module := NewWebsiteAnalysisModule()

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
func (f *WebsiteAnalysisFactory) GetModuleType() string {
	return "website_analysis"
}
