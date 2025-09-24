package data_extraction

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"kyb-platform/internal/observability"
	"kyb-platform/internal/shared"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TechnologyExtractor extracts technology stack information from business data
type TechnologyExtractor struct {
	// Configuration
	config *TechnologyConfig

	// Observability
	logger *observability.Logger
	tracer trace.Tracer

	// Pattern matching
	programmingLanguages []*regexp.Regexp
	frameworks           []*regexp.Regexp
	cloudPlatforms       []*regexp.Regexp
	thirdPartyServices   []*regexp.Regexp
	developmentTools     []*regexp.Regexp
}

// TechnologyConfig holds configuration for the technology extractor
type TechnologyConfig struct {
	// Pattern matching settings
	CaseSensitive bool
	MaxPatterns   int

	// Confidence scoring settings
	MinConfidenceThreshold float64
	MaxConfidenceThreshold float64

	// Web scraping settings
	EnableWebScraping bool
	ScrapingTimeout   time.Duration

	// Processing settings
	Timeout time.Duration
}

// TechnologyStack represents extracted technology stack information
type TechnologyStack struct {
	// Programming languages
	ProgrammingLanguages []string           `json:"programming_languages"`
	LanguageConfidence   map[string]float64 `json:"language_confidence"`

	// Frameworks and libraries
	Frameworks          []string           `json:"frameworks"`
	FrameworkConfidence map[string]float64 `json:"framework_confidence"`

	// Cloud platforms
	CloudPlatforms  []string           `json:"cloud_platforms"`
	CloudConfidence map[string]float64 `json:"cloud_confidence"`

	// Third-party services
	ThirdPartyServices []string           `json:"third_party_services"`
	ServiceConfidence  map[string]float64 `json:"service_confidence"`

	// Development tools
	DevelopmentTools []string           `json:"development_tools"`
	ToolConfidence   map[string]float64 `json:"tool_confidence"`

	// Additional details
	TechDetails        map[string]interface{} `json:"tech_details,omitempty"`
	SupportingEvidence []string               `json:"supporting_evidence,omitempty"`

	// Overall assessment
	OverallConfidence float64 `json:"overall_confidence"`

	// Metadata
	ExtractedAt time.Time `json:"extracted_at"`
	DataSources []string  `json:"data_sources"`
}

// Technology categories
const (
	// Programming languages
	TechPython     = "Python"
	TechJavaScript = "JavaScript"
	TechJava       = "Java"
	TechGo         = "Go"
	TechCSharp     = "C#"
	TechPHP        = "PHP"
	TechRuby       = "Ruby"
	TechSwift      = "Swift"
	TechKotlin     = "Kotlin"
	TechRust       = "Rust"
	TechTypeScript = "TypeScript"
	TechCPlusPlus  = "C++"
	TechC          = "C"

	// Frameworks
	TechReact   = "React"
	TechVue     = "Vue.js"
	TechAngular = "Angular"
	TechNodeJS  = "Node.js"
	TechExpress = "Express.js"
	TechDjango  = "Django"
	TechFlask   = "Flask"
	TechSpring  = "Spring"
	TechLaravel = "Laravel"
	TechRails   = "Ruby on Rails"
	TechASPNet  = "ASP.NET"
	TechFastAPI = "FastAPI"
	TechGin     = "Gin"
	TechEcho    = "Echo"

	// Cloud platforms
	TechAWS          = "AWS"
	TechAzure        = "Azure"
	TechGCP          = "Google Cloud Platform"
	TechDigitalOcean = "DigitalOcean"
	TechHeroku       = "Heroku"
	TechVercel       = "Vercel"
	TechNetlify      = "Netlify"
	TechRailway      = "Railway"

	// Third-party services
	TechStripe        = "Stripe"
	TechTwilio        = "Twilio"
	TechSendGrid      = "SendGrid"
	TechMailchimp     = "Mailchimp"
	TechSlack         = "Slack"
	TechDiscord       = "Discord"
	TechZoom          = "Zoom"
	TechNotion        = "Notion"
	TechAirtable      = "Airtable"
	TechZapier        = "Zapier"
	TechShopify       = "Shopify"
	TechWooCommerce   = "WooCommerce"
	TechMongoDB       = "MongoDB"
	TechPostgreSQL    = "PostgreSQL"
	TechMySQL         = "MySQL"
	TechRedis         = "Redis"
	TechElasticsearch = "Elasticsearch"
	TechKafka         = "Apache Kafka"
	TechDocker        = "Docker"
	TechKubernetes    = "Kubernetes"

	// Development tools
	TechGit       = "Git"
	TechGitHub    = "GitHub"
	TechGitLab    = "GitLab"
	TechBitbucket = "Bitbucket"
	TechVSCode    = "VS Code"
	TechIntelliJ  = "IntelliJ IDEA"
	TechEclipse   = "Eclipse"
	TechJira      = "Jira"
	TechTrello    = "Trello"
	TechAsana     = "Asana"
	TechFigma     = "Figma"
	TechSketch    = "Sketch"
	TechAdobe     = "Adobe Creative Suite"
)

// NewTechnologyExtractor creates a new technology stack extractor
func NewTechnologyExtractor(
	config *TechnologyConfig,
	logger *observability.Logger,
	tracer trace.Tracer,
) *TechnologyExtractor {
	// Set default configuration
	if config == nil {
		config = &TechnologyConfig{
			CaseSensitive:          false,
			MaxPatterns:            100,
			MinConfidenceThreshold: 0.3,
			MaxConfidenceThreshold: 1.0,
			EnableWebScraping:      false,
			ScrapingTimeout:        10 * time.Second,
			Timeout:                30 * time.Second,
		}
	}

	extractor := &TechnologyExtractor{
		config: config,
		logger: logger,
		tracer: tracer,
	}

	// Initialize pattern matching
	extractor.initializePatterns()

	return extractor
}

// initializePatterns initializes all pattern matching regexes
func (te *TechnologyExtractor) initializePatterns() {
	// Programming languages patterns
	te.programmingLanguages = []*regexp.Regexp{
		regexp.MustCompile(`(?i)python`),
		regexp.MustCompile(`(?i)javascript`),
		regexp.MustCompile(`(?i)js\b`),
		regexp.MustCompile(`(?i)java\b`),
		regexp.MustCompile(`(?i)go\b`),
		regexp.MustCompile(`(?i)golang`),
		regexp.MustCompile(`(?i)c#`),
		regexp.MustCompile(`(?i)csharp`),
		regexp.MustCompile(`(?i)php`),
		regexp.MustCompile(`(?i)ruby`),
		regexp.MustCompile(`(?i)swift`),
		regexp.MustCompile(`(?i)kotlin`),
		regexp.MustCompile(`(?i)rust`),
		regexp.MustCompile(`(?i)typescript`),
		regexp.MustCompile(`(?i)ts\b`),
		regexp.MustCompile(`(?i)c\+\+`),
		regexp.MustCompile(`(?i)cpp`),
		regexp.MustCompile(`(?i)c\b`),
	}

	// Framework patterns
	te.frameworks = []*regexp.Regexp{
		regexp.MustCompile(`(?i)react`),
		regexp.MustCompile(`(?i)vue`),
		regexp.MustCompile(`(?i)angular`),
		regexp.MustCompile(`(?i)node\.js`),
		regexp.MustCompile(`(?i)nodejs`),
		regexp.MustCompile(`(?i)express`),
		regexp.MustCompile(`(?i)django`),
		regexp.MustCompile(`(?i)flask`),
		regexp.MustCompile(`(?i)spring`),
		regexp.MustCompile(`(?i)laravel`),
		regexp.MustCompile(`(?i)rails`),
		regexp.MustCompile(`(?i)ruby on rails`),
		regexp.MustCompile(`(?i)asp\.net`),
		regexp.MustCompile(`(?i)aspnet`),
		regexp.MustCompile(`(?i)fastapi`),
		regexp.MustCompile(`(?i)gin`),
		regexp.MustCompile(`(?i)echo`),
	}

	// Cloud platform patterns
	te.cloudPlatforms = []*regexp.Regexp{
		regexp.MustCompile(`(?i)aws`),
		regexp.MustCompile(`(?i)amazon web services`),
		regexp.MustCompile(`(?i)azure`),
		regexp.MustCompile(`(?i)microsoft azure`),
		regexp.MustCompile(`(?i)google cloud`),
		regexp.MustCompile(`(?i)gcp`),
		regexp.MustCompile(`(?i)digitalocean`),
		regexp.MustCompile(`(?i)heroku`),
		regexp.MustCompile(`(?i)vercel`),
		regexp.MustCompile(`(?i)netlify`),
		regexp.MustCompile(`(?i)railway`),
	}

	// Third-party service patterns
	te.thirdPartyServices = []*regexp.Regexp{
		regexp.MustCompile(`(?i)stripe`),
		regexp.MustCompile(`(?i)twilio`),
		regexp.MustCompile(`(?i)sendgrid`),
		regexp.MustCompile(`(?i)mailchimp`),
		regexp.MustCompile(`(?i)slack`),
		regexp.MustCompile(`(?i)discord`),
		regexp.MustCompile(`(?i)zoom`),
		regexp.MustCompile(`(?i)notion`),
		regexp.MustCompile(`(?i)airtable`),
		regexp.MustCompile(`(?i)zapier`),
		regexp.MustCompile(`(?i)shopify`),
		regexp.MustCompile(`(?i)woocommerce`),
		regexp.MustCompile(`(?i)mongodb`),
		regexp.MustCompile(`(?i)postgresql`),
		regexp.MustCompile(`(?i)postgres`),
		regexp.MustCompile(`(?i)mysql`),
		regexp.MustCompile(`(?i)redis`),
		regexp.MustCompile(`(?i)elasticsearch`),
		regexp.MustCompile(`(?i)kafka`),
		regexp.MustCompile(`(?i)docker`),
		regexp.MustCompile(`(?i)kubernetes`),
		regexp.MustCompile(`(?i)k8s`),
	}

	// Development tool patterns
	te.developmentTools = []*regexp.Regexp{
		regexp.MustCompile(`(?i)git`),
		regexp.MustCompile(`(?i)github`),
		regexp.MustCompile(`(?i)gitlab`),
		regexp.MustCompile(`(?i)bitbucket`),
		regexp.MustCompile(`(?i)vs code`),
		regexp.MustCompile(`(?i)vscode`),
		regexp.MustCompile(`(?i)intellij`),
		regexp.MustCompile(`(?i)eclipse`),
		regexp.MustCompile(`(?i)jira`),
		regexp.MustCompile(`(?i)trello`),
		regexp.MustCompile(`(?i)asana`),
		regexp.MustCompile(`(?i)figma`),
		regexp.MustCompile(`(?i)sketch`),
		regexp.MustCompile(`(?i)adobe`),
	}
}

// ExtractTechnologyStack extracts technology stack information from business data
func (te *TechnologyExtractor) ExtractTechnologyStack(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
) (*TechnologyStack, error) {
	ctx, span := te.tracer.Start(ctx, "TechnologyExtractor.ExtractTechnologyStack")
	defer span.End()

	span.SetAttributes(
		attribute.String("business_name", businessData.BusinessName),
		attribute.String("website", businessData.WebsiteURL),
	)

	// Create result structure
	result := &TechnologyStack{
		ExtractedAt:         time.Now(),
		DataSources:         []string{"text_analysis", "pattern_matching"},
		TechDetails:         make(map[string]interface{}),
		LanguageConfidence:  make(map[string]float64),
		FrameworkConfidence: make(map[string]float64),
		CloudConfidence:     make(map[string]float64),
		ServiceConfidence:   make(map[string]float64),
		ToolConfidence:      make(map[string]float64),
	}

	// Extract programming languages
	if err := te.extractProgrammingLanguages(ctx, businessData, result); err != nil {
		te.logger.Warn("failed to extract programming languages", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Extract frameworks
	if err := te.extractFrameworks(ctx, businessData, result); err != nil {
		te.logger.Warn("failed to extract frameworks", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Extract cloud platforms
	if err := te.extractCloudPlatforms(ctx, businessData, result); err != nil {
		te.logger.Warn("failed to extract cloud platforms", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Extract third-party services
	if err := te.extractThirdPartyServices(ctx, businessData, result); err != nil {
		te.logger.Warn("failed to extract third-party services", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Extract development tools
	if err := te.extractDevelopmentTools(ctx, businessData, result); err != nil {
		te.logger.Warn("failed to extract development tools", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	// Calculate overall confidence
	te.calculateOverallConfidence(result)

	// Validate results
	if err := te.validateResults(result); err != nil {
		te.logger.Warn("technology stack validation failed", map[string]interface{}{
			"business_name": businessData.BusinessName,
			"error":         err.Error(),
		})
	}

	te.logger.Info("technology stack extraction completed", map[string]interface{}{
		"business_name":         businessData.BusinessName,
		"programming_languages": len(result.ProgrammingLanguages),
		"frameworks":            len(result.Frameworks),
		"cloud_platforms":       len(result.CloudPlatforms),
		"third_party_services":  len(result.ThirdPartyServices),
		"development_tools":     len(result.DevelopmentTools),
		"overall_confidence":    result.OverallConfidence,
	})

	return result, nil
}

// extractProgrammingLanguages extracts programming languages
func (te *TechnologyExtractor) extractProgrammingLanguages(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *TechnologyStack,
) error {
	ctx, span := te.tracer.Start(ctx, "TechnologyExtractor.extractProgrammingLanguages")
	defer span.End()

	text := te.combineText(businessData)
	detectedLanguages := make(map[string]float64)

	// Check programming language patterns
	for _, pattern := range te.programmingLanguages {
		if pattern.MatchString(text) {
			language := te.mapPatternToLanguage(pattern.String())
			if language != "" {
				detectedLanguages[language] = 0.8
				result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
			}
		}
	}

	// Add detected languages to result
	for language, confidence := range detectedLanguages {
		result.ProgrammingLanguages = append(result.ProgrammingLanguages, language)
		result.LanguageConfidence[language] = confidence
	}

	span.SetAttributes(
		attribute.StringSlice("languages", result.ProgrammingLanguages),
		attribute.Int("count", len(result.ProgrammingLanguages)),
	)

	return nil
}

// extractFrameworks extracts frameworks and libraries
func (te *TechnologyExtractor) extractFrameworks(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *TechnologyStack,
) error {
	ctx, span := te.tracer.Start(ctx, "TechnologyExtractor.extractFrameworks")
	defer span.End()

	text := te.combineText(businessData)
	detectedFrameworks := make(map[string]float64)

	// Check framework patterns
	for _, pattern := range te.frameworks {
		if pattern.MatchString(text) {
			framework := te.mapPatternToFramework(pattern.String())
			if framework != "" {
				detectedFrameworks[framework] = 0.8
				result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
			}
		}
	}

	// Add detected frameworks to result
	for framework, confidence := range detectedFrameworks {
		result.Frameworks = append(result.Frameworks, framework)
		result.FrameworkConfidence[framework] = confidence
	}

	span.SetAttributes(
		attribute.StringSlice("frameworks", result.Frameworks),
		attribute.Int("count", len(result.Frameworks)),
	)

	return nil
}

// extractCloudPlatforms extracts cloud platforms
func (te *TechnologyExtractor) extractCloudPlatforms(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *TechnologyStack,
) error {
	ctx, span := te.tracer.Start(ctx, "TechnologyExtractor.extractCloudPlatforms")
	defer span.End()

	text := te.combineText(businessData)
	detectedPlatforms := make(map[string]float64)

	// Check cloud platform patterns
	for _, pattern := range te.cloudPlatforms {
		if pattern.MatchString(text) {
			platform := te.mapPatternToCloudPlatform(pattern.String())
			if platform != "" {
				detectedPlatforms[platform] = 0.8
				result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
			}
		}
	}

	// Add detected platforms to result
	for platform, confidence := range detectedPlatforms {
		result.CloudPlatforms = append(result.CloudPlatforms, platform)
		result.CloudConfidence[platform] = confidence
	}

	span.SetAttributes(
		attribute.StringSlice("cloud_platforms", result.CloudPlatforms),
		attribute.Int("count", len(result.CloudPlatforms)),
	)

	return nil
}

// extractThirdPartyServices extracts third-party services
func (te *TechnologyExtractor) extractThirdPartyServices(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *TechnologyStack,
) error {
	ctx, span := te.tracer.Start(ctx, "TechnologyExtractor.extractThirdPartyServices")
	defer span.End()

	text := te.combineText(businessData)
	detectedServices := make(map[string]float64)

	// Check third-party service patterns
	for _, pattern := range te.thirdPartyServices {
		if pattern.MatchString(text) {
			service := te.mapPatternToService(pattern.String())
			if service != "" {
				detectedServices[service] = 0.8
				result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
			}
		}
	}

	// Add detected services to result
	for service, confidence := range detectedServices {
		result.ThirdPartyServices = append(result.ThirdPartyServices, service)
		result.ServiceConfidence[service] = confidence
	}

	span.SetAttributes(
		attribute.StringSlice("third_party_services", result.ThirdPartyServices),
		attribute.Int("count", len(result.ThirdPartyServices)),
	)

	return nil
}

// extractDevelopmentTools extracts development tools
func (te *TechnologyExtractor) extractDevelopmentTools(
	ctx context.Context,
	businessData *shared.BusinessClassificationRequest,
	result *TechnologyStack,
) error {
	ctx, span := te.tracer.Start(ctx, "TechnologyExtractor.extractDevelopmentTools")
	defer span.End()

	text := te.combineText(businessData)
	detectedTools := make(map[string]float64)

	// Check development tool patterns
	for _, pattern := range te.developmentTools {
		if pattern.MatchString(text) {
			tool := te.mapPatternToTool(pattern.String())
			if tool != "" {
				detectedTools[tool] = 0.8
				result.SupportingEvidence = append(result.SupportingEvidence, pattern.String())
			}
		}
	}

	// Add detected tools to result
	for tool, confidence := range detectedTools {
		result.DevelopmentTools = append(result.DevelopmentTools, tool)
		result.ToolConfidence[tool] = confidence
	}

	span.SetAttributes(
		attribute.StringSlice("development_tools", result.DevelopmentTools),
		attribute.Int("count", len(result.DevelopmentTools)),
	)

	return nil
}

// combineText combines all available text for analysis
func (te *TechnologyExtractor) combineText(businessData *shared.BusinessClassificationRequest) string {
	var parts []string

	// Add business name
	if businessData.BusinessName != "" {
		parts = append(parts, businessData.BusinessName)
	}

	// Add description
	if businessData.Description != "" {
		parts = append(parts, businessData.Description)
	}

	// Add keywords
	if len(businessData.Keywords) > 0 {
		parts = append(parts, strings.Join(businessData.Keywords, " "))
	}

	// Add address
	if businessData.Address != "" {
		parts = append(parts, businessData.Address)
	}

	// Combine all parts
	text := strings.Join(parts, " ")

	// Normalize text
	if !te.config.CaseSensitive {
		text = strings.ToLower(text)
	}

	return text
}

// Mapping functions to convert patterns to standardized names
func (te *TechnologyExtractor) mapPatternToLanguage(pattern string) string {
	pattern = strings.ToLower(pattern)
	switch {
	case strings.Contains(pattern, "python"):
		return TechPython
	case strings.Contains(pattern, "javascript") || strings.Contains(pattern, "js\\b"):
		return TechJavaScript
	case strings.Contains(pattern, "java\\b"):
		return TechJava
	case strings.Contains(pattern, "go\\b") || strings.Contains(pattern, "golang"):
		return TechGo
	case strings.Contains(pattern, "c#") || strings.Contains(pattern, "csharp"):
		return TechCSharp
	case strings.Contains(pattern, "php"):
		return TechPHP
	case strings.Contains(pattern, "ruby"):
		return TechRuby
	case strings.Contains(pattern, "swift"):
		return TechSwift
	case strings.Contains(pattern, "kotlin"):
		return TechKotlin
	case strings.Contains(pattern, "rust"):
		return TechRust
	case strings.Contains(pattern, "typescript") || strings.Contains(pattern, "ts\\b"):
		return TechTypeScript
	case strings.Contains(pattern, "c\\+\\+") || strings.Contains(pattern, "cpp"):
		return TechCPlusPlus
	case strings.Contains(pattern, "c\\b"):
		return TechC
	default:
		return ""
	}
}

func (te *TechnologyExtractor) mapPatternToFramework(pattern string) string {
	pattern = strings.ToLower(pattern)
	switch {
	case strings.Contains(pattern, "react"):
		return TechReact
	case strings.Contains(pattern, "vue"):
		return TechVue
	case strings.Contains(pattern, "angular"):
		return TechAngular
	case strings.Contains(pattern, "node\\.js") || strings.Contains(pattern, "nodejs"):
		return TechNodeJS
	case strings.Contains(pattern, "express"):
		return TechExpress
	case strings.Contains(pattern, "django"):
		return TechDjango
	case strings.Contains(pattern, "flask"):
		return TechFlask
	case strings.Contains(pattern, "spring"):
		return TechSpring
	case strings.Contains(pattern, "laravel"):
		return TechLaravel
	case strings.Contains(pattern, "rails") || strings.Contains(pattern, "ruby on rails"):
		return TechRails
	case strings.Contains(pattern, "asp\\.net") || strings.Contains(pattern, "aspnet"):
		return TechASPNet
	case strings.Contains(pattern, "fastapi"):
		return TechFastAPI
	case strings.Contains(pattern, "gin"):
		return TechGin
	case strings.Contains(pattern, "echo"):
		return TechEcho
	default:
		return ""
	}
}

func (te *TechnologyExtractor) mapPatternToCloudPlatform(pattern string) string {
	pattern = strings.ToLower(pattern)
	switch {
	case strings.Contains(pattern, "aws") || strings.Contains(pattern, "amazon web services"):
		return TechAWS
	case strings.Contains(pattern, "azure") || strings.Contains(pattern, "microsoft azure"):
		return TechAzure
	case strings.Contains(pattern, "google cloud") || strings.Contains(pattern, "gcp"):
		return TechGCP
	case strings.Contains(pattern, "digitalocean"):
		return TechDigitalOcean
	case strings.Contains(pattern, "heroku"):
		return TechHeroku
	case strings.Contains(pattern, "vercel"):
		return TechVercel
	case strings.Contains(pattern, "netlify"):
		return TechNetlify
	case strings.Contains(pattern, "railway"):
		return TechRailway
	default:
		return ""
	}
}

func (te *TechnologyExtractor) mapPatternToService(pattern string) string {
	pattern = strings.ToLower(pattern)
	switch {
	case strings.Contains(pattern, "stripe"):
		return TechStripe
	case strings.Contains(pattern, "twilio"):
		return TechTwilio
	case strings.Contains(pattern, "sendgrid"):
		return TechSendGrid
	case strings.Contains(pattern, "mailchimp"):
		return TechMailchimp
	case strings.Contains(pattern, "slack"):
		return TechSlack
	case strings.Contains(pattern, "discord"):
		return TechDiscord
	case strings.Contains(pattern, "zoom"):
		return TechZoom
	case strings.Contains(pattern, "notion"):
		return TechNotion
	case strings.Contains(pattern, "airtable"):
		return TechAirtable
	case strings.Contains(pattern, "zapier"):
		return TechZapier
	case strings.Contains(pattern, "shopify"):
		return TechShopify
	case strings.Contains(pattern, "woocommerce"):
		return TechWooCommerce
	case strings.Contains(pattern, "mongodb"):
		return TechMongoDB
	case strings.Contains(pattern, "postgresql") || strings.Contains(pattern, "postgres"):
		return TechPostgreSQL
	case strings.Contains(pattern, "mysql"):
		return TechMySQL
	case strings.Contains(pattern, "redis"):
		return TechRedis
	case strings.Contains(pattern, "elasticsearch"):
		return TechElasticsearch
	case strings.Contains(pattern, "kafka"):
		return TechKafka
	case strings.Contains(pattern, "docker"):
		return TechDocker
	case strings.Contains(pattern, "kubernetes") || strings.Contains(pattern, "k8s"):
		return TechKubernetes
	default:
		return ""
	}
}

func (te *TechnologyExtractor) mapPatternToTool(pattern string) string {
	pattern = strings.ToLower(pattern)
	switch {
	case strings.Contains(pattern, "git"):
		return TechGit
	case strings.Contains(pattern, "github"):
		return TechGitHub
	case strings.Contains(pattern, "gitlab"):
		return TechGitLab
	case strings.Contains(pattern, "bitbucket"):
		return TechBitbucket
	case strings.Contains(pattern, "vs code") || strings.Contains(pattern, "vscode"):
		return TechVSCode
	case strings.Contains(pattern, "intellij"):
		return TechIntelliJ
	case strings.Contains(pattern, "eclipse"):
		return TechEclipse
	case strings.Contains(pattern, "jira"):
		return TechJira
	case strings.Contains(pattern, "trello"):
		return TechTrello
	case strings.Contains(pattern, "asana"):
		return TechAsana
	case strings.Contains(pattern, "figma"):
		return TechFigma
	case strings.Contains(pattern, "sketch"):
		return TechSketch
	case strings.Contains(pattern, "adobe"):
		return TechAdobe
	default:
		return ""
	}
}

// calculateOverallConfidence calculates the overall confidence score
func (te *TechnologyExtractor) calculateOverallConfidence(result *TechnologyStack) {
	var scores []float64
	var weights []float64

	// Calculate average confidence for each category
	if len(result.LanguageConfidence) > 0 {
		avgConfidence := te.calculateAverageConfidence(result.LanguageConfidence)
		scores = append(scores, avgConfidence)
		weights = append(weights, 0.25) // 25% weight
	}

	if len(result.FrameworkConfidence) > 0 {
		avgConfidence := te.calculateAverageConfidence(result.FrameworkConfidence)
		scores = append(scores, avgConfidence)
		weights = append(weights, 0.25) // 25% weight
	}

	if len(result.CloudConfidence) > 0 {
		avgConfidence := te.calculateAverageConfidence(result.CloudConfidence)
		scores = append(scores, avgConfidence)
		weights = append(weights, 0.2) // 20% weight
	}

	if len(result.ServiceConfidence) > 0 {
		avgConfidence := te.calculateAverageConfidence(result.ServiceConfidence)
		scores = append(scores, avgConfidence)
		weights = append(weights, 0.2) // 20% weight
	}

	if len(result.ToolConfidence) > 0 {
		avgConfidence := te.calculateAverageConfidence(result.ToolConfidence)
		scores = append(scores, avgConfidence)
		weights = append(weights, 0.1) // 10% weight
	}

	// Calculate weighted average
	if len(scores) > 0 {
		totalWeight := 0.0
		weightedSum := 0.0

		for i, score := range scores {
			weight := weights[i]
			weightedSum += score * weight
			totalWeight += weight
		}

		if totalWeight > 0 {
			result.OverallConfidence = weightedSum / totalWeight
		}
	} else {
		result.OverallConfidence = 0.0
	}
}

// calculateAverageConfidence calculates the average confidence for a map of confidences
func (te *TechnologyExtractor) calculateAverageConfidence(confidences map[string]float64) float64 {
	if len(confidences) == 0 {
		return 0.0
	}

	total := 0.0
	for _, confidence := range confidences {
		total += confidence
	}

	return total / float64(len(confidences))
}

// validateResults validates the extracted results
func (te *TechnologyExtractor) validateResults(result *TechnologyStack) error {
	// Validate confidence scores
	if result.OverallConfidence < 0 || result.OverallConfidence > 1 {
		return fmt.Errorf("overall confidence score %f is out of range [0,1]", result.OverallConfidence)
	}

	// Validate that we have at least some technology detected
	totalTechnologies := len(result.ProgrammingLanguages) + len(result.Frameworks) +
		len(result.CloudPlatforms) + len(result.ThirdPartyServices) + len(result.DevelopmentTools)

	if totalTechnologies == 0 {
		te.logger.Warn("no technologies detected", map[string]interface{}{
			"total_technologies": totalTechnologies,
		})
	}

	return nil
}
