package enrichment

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TechnologyStackAnalyzer analyzes website technology stack from content and headers
type TechnologyStackAnalyzer struct {
	logger *zap.Logger
	tracer trace.Tracer
	config *TechnologyStackAnalyzerConfig
}

// TechnologyStackAnalyzerConfig contains configuration for technology stack analysis
type TechnologyStackAnalyzerConfig struct {
	// CMS detection patterns
	CMSIndicators map[string][]string
	// Framework detection patterns
	FrameworkIndicators map[string][]string
	// Development tools detection
	ToolIndicators map[string][]string
	// Analytics and tracking services
	AnalyticsIndicators map[string][]string
	// Hosting and infrastructure
	HostingIndicators map[string][]string
	// Confidence thresholds
	MinConfidenceScore float64
	// Maximum technologies to detect per category
	MaxTechnologiesPerCategory int
}

// Technology represents a detected technology
type Technology struct {
	Name            string    `json:"name"`             // Technology name
	Category        string    `json:"category"`         // cms, framework, tool, analytics, hosting
	Type            string    `json:"type"`             // frontend, backend, fullstack, etc.
	Version         string    `json:"version"`          // Version if detected
	ConfidenceScore float64   `json:"confidence_score"` // Detection confidence
	Evidence        []string  `json:"evidence"`         // Evidence supporting detection
	DetectedAt      time.Time `json:"detected_at"`      // When detected
	Source          string    `json:"source"`           // html, headers, meta, etc.
}

// TechnologyStack represents the complete technology stack
type TechnologyStack struct {
	CMS        []Technology `json:"cms"`        // Content Management Systems
	Frameworks []Technology `json:"frameworks"` // Frontend and backend frameworks
	Tools      []Technology `json:"tools"`      // Development and deployment tools
	Analytics  []Technology `json:"analytics"`  // Analytics and tracking services
	Hosting    []Technology `json:"hosting"`    // Hosting and infrastructure
	Database   []Technology `json:"database"`   // Database technologies
	Security   []Technology `json:"security"`   // Security and CDN services
}

// TechnologyStackResult contains the results of technology stack analysis
type TechnologyStackResult struct {
	TechnologyStack  *TechnologyStack `json:"technology_stack"`  // Detected technologies
	PrimaryCMS       *Technology      `json:"primary_cms"`       // Main CMS if detected
	PrimaryFramework *Technology      `json:"primary_framework"` // Main framework if detected
	StackType        string           `json:"stack_type"`        // modern, traditional, headless, etc.
	Complexity       string           `json:"complexity"`        // simple, moderate, complex
	ConfidenceScore  float64          `json:"confidence_score"`  // Overall confidence
	Evidence         []string         `json:"evidence"`          // Supporting evidence
	ProcessingTime   time.Duration    `json:"processing_time"`   // Time taken to process
}

// NewTechnologyStackAnalyzer creates a new technology stack analyzer
func NewTechnologyStackAnalyzer(logger *zap.Logger, config *TechnologyStackAnalyzerConfig) *TechnologyStackAnalyzer {
	if config == nil {
		config = getDefaultTechnologyStackAnalyzerConfig()
	}

	return &TechnologyStackAnalyzer{
		logger: logger,
		tracer: trace.NewNoopTracerProvider().Tracer("technology_stack_analyzer"),
		config: config,
	}
}

// AnalyzeTechnologyStack analyzes technology stack from website content and headers
func (tsa *TechnologyStackAnalyzer) AnalyzeTechnologyStack(ctx context.Context, content string, headers map[string]string) (*TechnologyStackResult, error) {
	ctx, span := tsa.tracer.Start(ctx, "technology_stack_analyzer.analyze_technology_stack")
	defer span.End()

	startTime := time.Now()

	// Normalize content for analysis
	lowerContent := strings.ToLower(content)
	lowerHeaders := make(map[string]string)
	for key, value := range headers {
		lowerHeaders[strings.ToLower(key)] = strings.ToLower(value)
	}

	// Detect technologies by category
	cmsTechnologies := tsa.detectCMS(lowerContent, lowerHeaders)
	frameworkTechnologies := tsa.detectFrameworks(lowerContent, lowerHeaders)
	toolTechnologies := tsa.detectTools(lowerContent, lowerHeaders)
	analyticsTechnologies := tsa.detectAnalytics(lowerContent, lowerHeaders)
	hostingTechnologies := tsa.detectHosting(lowerContent, lowerHeaders)
	databaseTechnologies := tsa.detectDatabases(lowerContent, lowerHeaders)
	securityTechnologies := tsa.detectSecurity(lowerContent, lowerHeaders)

	// Create technology stack
	technologyStack := &TechnologyStack{
		CMS:        cmsTechnologies,
		Frameworks: frameworkTechnologies,
		Tools:      toolTechnologies,
		Analytics:  analyticsTechnologies,
		Hosting:    hostingTechnologies,
		Database:   databaseTechnologies,
		Security:   securityTechnologies,
	}

	// Determine primary technologies
	primaryCMS := tsa.determinePrimaryTechnology(cmsTechnologies)
	primaryFramework := tsa.determinePrimaryTechnology(frameworkTechnologies)

	// Analyze stack characteristics
	stackType := tsa.determineStackType(technologyStack)
	complexity := tsa.determineComplexity(technologyStack)

	// Calculate overall confidence
	confidenceScore := tsa.calculateOverallConfidence(technologyStack)

	// Collect evidence
	evidence := tsa.collectEvidence(technologyStack)

	result := &TechnologyStackResult{
		TechnologyStack:  technologyStack,
		PrimaryCMS:       primaryCMS,
		PrimaryFramework: primaryFramework,
		StackType:        stackType,
		Complexity:       complexity,
		ConfidenceScore:  confidenceScore,
		Evidence:         evidence,
		ProcessingTime:   time.Since(startTime),
	}

	tsa.logger.Info("technology stack analysis completed",
		zap.Int("cms_count", len(cmsTechnologies)),
		zap.Int("framework_count", len(frameworkTechnologies)),
		zap.Int("tool_count", len(toolTechnologies)),
		zap.String("stack_type", stackType),
		zap.String("complexity", complexity),
		zap.Float64("confidence", confidenceScore),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}
