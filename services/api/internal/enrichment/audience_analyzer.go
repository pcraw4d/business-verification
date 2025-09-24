package enrichment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// AudienceAnalyzer analyzes target audience and customer types from website content
type AudienceAnalyzer struct {
	config *AudienceConfig
	logger *zap.Logger
	tracer trace.Tracer
}

// AudienceConfig configuration for audience analysis
type AudienceConfig struct {
	// Analysis thresholds
	MinConfidenceThreshold float64
	MinEvidenceCount       int
	MinContentLength       int

	// Scoring weights
	DemographicWeight float64
	IndustryWeight    float64
	SizeWeight        float64
	GeographicWeight  float64
	BehavioralWeight  float64

	// Validation settings
	RequireMultipleIndicators bool
	EnableFallbackAnalysis    bool
	ValidatePersonas          bool
}

// AudienceResult comprehensive audience analysis result
type AudienceResult struct {
	// Primary classifications
	PrimaryAudience   string            `json:"primary_audience"`
	SecondaryAudience string            `json:"secondary_audience,omitempty"`
	CustomerTypes     []string          `json:"customer_types"`
	CustomerPersonas  []CustomerPersona `json:"customer_personas"`

	// Segmentation analysis
	Demographics       Demographics `json:"demographics"`
	Industries         []Industry   `json:"industries"`
	CompanySizes       []string     `json:"company_sizes"`
	GeographicMarkets  []string     `json:"geographic_markets"`
	BehavioralSegments []string     `json:"behavioral_segments"`

	// Confidence and validation
	ConfidenceScore  float64         `json:"confidence_score"`
	ComponentScores  ComponentScores `json:"component_scores"`
	Evidence         []string        `json:"evidence"`
	ExtractedPhrases []string        `json:"extracted_phrases"`

	// Quality assessment
	IsValidated      bool             `json:"is_validated"`
	ValidationStatus ValidationStatus `json:"validation_status"`
	DataQualityScore float64          `json:"data_quality_score"`
	Reasoning        string           `json:"reasoning"`

	// Metadata
	AnalyzedAt     time.Time     `json:"analyzed_at"`
	ProcessingTime time.Duration `json:"processing_time"`
	SourceURL      string        `json:"source_url,omitempty"`
}

// CustomerPersona represents a detailed customer persona
type CustomerPersona struct {
	Name               string   `json:"name"`
	Type               string   `json:"type"` // enterprise, consumer, marketplace_participant
	Description        string   `json:"description"`
	Characteristics    []string `json:"characteristics"`
	Needs              []string `json:"needs"`
	PainPoints         []string `json:"pain_points"`
	BuyingBehavior     string   `json:"buying_behavior"`
	DecisionMakers     []string `json:"decision_makers"`
	InfluencingFactors []string `json:"influencing_factors"`
	ConfidenceScore    float64  `json:"confidence_score"`
}

// Demographics represents demographic information
type Demographics struct {
	AgeGroups         []string `json:"age_groups"`
	IncomeGroups      []string `json:"income_groups"`
	EducationLevels   []string `json:"education_levels"`
	ProfessionTypes   []string `json:"profession_types"`
	LifestyleSegments []string `json:"lifestyle_segments"`
	TechSavviness     string   `json:"tech_savviness"`
}

// Industry represents an industry vertical
type Industry struct {
	Name            string   `json:"name"`
	Sector          string   `json:"sector"`
	SubIndustries   []string `json:"sub_industries"`
	ConfidenceScore float64  `json:"confidence_score"`
}

// ComponentScores detailed scoring breakdown
type ComponentScores struct {
	DemographicScore float64 `json:"demographic_score"`
	IndustryScore    float64 `json:"industry_score"`
	SizeScore        float64 `json:"size_score"`
	GeographicScore  float64 `json:"geographic_score"`
	BehavioralScore  float64 `json:"behavioral_score"`
	PersonaScore     float64 `json:"persona_score"`
}

// ValidationStatus represents validation status for audience analysis
type ValidationStatus struct {
	IsValid          bool      `json:"is_valid"`
	ValidationErrors []string  `json:"validation_errors"`
	LastValidated    time.Time `json:"last_validated"`
}

// NewAudienceAnalyzer creates a new audience analyzer
func NewAudienceAnalyzer(config *AudienceConfig, logger *zap.Logger) *AudienceAnalyzer {
	if config == nil {
		config = GetDefaultAudienceConfig()
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	return &AudienceAnalyzer{
		config: config,
		logger: logger,
		tracer: otel.Tracer("audience-analyzer"),
	}
}

// GetDefaultAudienceConfig returns default configuration
func GetDefaultAudienceConfig() *AudienceConfig {
	return &AudienceConfig{
		MinConfidenceThreshold:    0.3,
		MinEvidenceCount:          2,
		MinContentLength:          50,
		DemographicWeight:         0.25,
		IndustryWeight:            0.25,
		SizeWeight:                0.20,
		GeographicWeight:          0.15,
		BehavioralWeight:          0.15,
		RequireMultipleIndicators: true,
		EnableFallbackAnalysis:    true,
		ValidatePersonas:          true,
	}
}

// AnalyzeAudience performs comprehensive audience analysis
func (aa *AudienceAnalyzer) AnalyzeAudience(ctx context.Context, content, sourceURL string) (*AudienceResult, error) {
	ctx, span := aa.tracer.Start(ctx, "audience_analyzer.analyze_audience")
	defer span.End()

	startTime := time.Now()

	// Input validation
	if err := aa.validateInput(content); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	result := &AudienceResult{
		SourceURL:        sourceURL,
		AnalyzedAt:       time.Now(),
		Evidence:         []string{},
		ExtractedPhrases: []string{},
		CustomerPersonas: []CustomerPersona{},
		ComponentScores:  ComponentScores{},
	}

	// Perform analysis components
	if err := aa.analyzeCustomerTypes(ctx, content, result); err != nil {
		aa.logger.Warn("Customer type analysis failed", zap.Error(err))
	}

	if err := aa.analyzeDemographics(ctx, content, result); err != nil {
		aa.logger.Warn("Demographics analysis failed", zap.Error(err))
	}

	if err := aa.analyzeIndustries(ctx, content, result); err != nil {
		aa.logger.Warn("Industry analysis failed", zap.Error(err))
	}

	if err := aa.analyzeCompanySizes(ctx, content, result); err != nil {
		aa.logger.Warn("Company size analysis failed", zap.Error(err))
	}

	if err := aa.analyzeGeographicMarkets(ctx, content, result); err != nil {
		aa.logger.Warn("Geographic analysis failed", zap.Error(err))
	}

	if err := aa.analyzeBehavioralSegments(ctx, content, result); err != nil {
		aa.logger.Warn("Behavioral analysis failed", zap.Error(err))
	}

	if err := aa.generateCustomerPersonas(ctx, content, result); err != nil {
		aa.logger.Warn("Persona generation failed", zap.Error(err))
	}

	// Determine primary and secondary audiences
	aa.determinePrimaryAudience(result)

	// Calculate confidence scores
	aa.calculateConfidenceScores(result)

	// Validate results
	aa.validateResult(result)

	// Generate reasoning
	result.Reasoning = aa.generateReasoning(result)

	// Calculate data quality
	result.DataQualityScore = aa.calculateDataQuality(result)

	result.ProcessingTime = time.Since(startTime)

	aa.logger.Info("Audience analysis completed",
		zap.String("primary_audience", result.PrimaryAudience),
		zap.Float64("confidence", result.ConfidenceScore),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// analyzeCustomerTypes identifies primary customer types
func (aa *AudienceAnalyzer) analyzeCustomerTypes(ctx context.Context, content string, result *AudienceResult) error {
	ctx, span := aa.tracer.Start(ctx, "audience_analyzer.analyze_customer_types")
	defer span.End()

	contentLower := strings.ToLower(content)
	customerTypes := make(map[string]float64)

	// Enterprise customers
	enterpriseIndicators := []string{
		"enterprise", "enterprises", "large companies", "corporations",
		"fortune 500", "multinational", "global companies", "big business",
		"corporate clients", "enterprise customers", "b2b clients",
		"institutional clients", "commercial clients", "business customers",
	}

	for _, indicator := range enterpriseIndicators {
		if strings.Contains(contentLower, indicator) {
			customerTypes["enterprise"] += 1.0
			result.Evidence = append(result.Evidence, fmt.Sprintf("Enterprise indicator: %s", indicator))
			result.ExtractedPhrases = append(result.ExtractedPhrases, indicator)
		}
	}

	// SME customers
	smeIndicators := []string{
		"small business", "small businesses", "sme", "small-medium enterprise",
		"medium business", "growing companies", "startup", "startups",
		"mid-size", "mid-sized", "emerging companies", "scale-ups",
	}

	for _, indicator := range smeIndicators {
		if strings.Contains(contentLower, indicator) {
			customerTypes["sme"] += 1.0
			result.Evidence = append(result.Evidence, fmt.Sprintf("SME indicator: %s", indicator))
			result.ExtractedPhrases = append(result.ExtractedPhrases, indicator)
		}
	}

	// Individual consumers
	consumerIndicators := []string{
		"consumers", "individuals", "personal use", "home users",
		"families", "households", "personal customers", "end users",
		"retail customers", "individual buyers", "personal clients",
		"everyday users", "regular people", "general public",
	}

	for _, indicator := range consumerIndicators {
		if strings.Contains(contentLower, indicator) {
			customerTypes["consumer"] += 1.0
			result.Evidence = append(result.Evidence, fmt.Sprintf("Consumer indicator: %s", indicator))
			result.ExtractedPhrases = append(result.ExtractedPhrases, indicator)
		}
	}

	// Professional users
	professionalIndicators := []string{
		"professionals", "experts", "specialists", "practitioners",
		"consultants", "freelancers", "contractors", "agencies",
		"professional services", "expert users", "industry professionals",
	}

	for _, indicator := range professionalIndicators {
		if strings.Contains(contentLower, indicator) {
			customerTypes["professional"] += 1.0
			result.Evidence = append(result.Evidence, fmt.Sprintf("Professional indicator: %s", indicator))
			result.ExtractedPhrases = append(result.ExtractedPhrases, indicator)
		}
	}

	// Marketplace participants
	marketplaceIndicators := []string{
		"buyers and sellers", "marketplace", "platform users", "community",
		"vendors", "merchants", "suppliers", "partners", "network",
		"ecosystem", "two-sided", "multi-sided", "participants",
	}

	for _, indicator := range marketplaceIndicators {
		if strings.Contains(contentLower, indicator) {
			customerTypes["marketplace_participant"] += 1.0
			result.Evidence = append(result.Evidence, fmt.Sprintf("Marketplace indicator: %s", indicator))
			result.ExtractedPhrases = append(result.ExtractedPhrases, indicator)
		}
	}

	// Convert to sorted list
	for customerType, score := range customerTypes {
		if score > 0 {
			result.CustomerTypes = append(result.CustomerTypes, customerType)
		}
	}

	return nil
}

// analyzeDemographics analyzes demographic indicators
func (aa *AudienceAnalyzer) analyzeDemographics(ctx context.Context, content string, result *AudienceResult) error {
	ctx, span := aa.tracer.Start(ctx, "audience_analyzer.analyze_demographics")
	defer span.End()

	contentLower := strings.ToLower(content)
	demographics := &Demographics{}

	// Age groups
	ageIndicators := map[string][]string{
		"young_adults": {"millennials", "gen z", "young adults", "college students", "university students", "20s", "30s"},
		"middle_aged":  {"gen x", "middle-aged", "professionals", "40s", "50s", "working adults"},
		"seniors":      {"baby boomers", "seniors", "retirees", "60+", "elderly", "mature adults"},
		"families":     {"families", "parents", "children", "kids", "family-friendly", "household"},
	}

	for ageGroup, indicators := range ageIndicators {
		for _, indicator := range indicators {
			if strings.Contains(contentLower, indicator) {
				demographics.AgeGroups = append(demographics.AgeGroups, ageGroup)
				result.Evidence = append(result.Evidence, fmt.Sprintf("Age group indicator: %s", indicator))
				break
			}
		}
	}

	// Income groups
	incomeIndicators := map[string][]string{
		"high_income":      {"luxury", "premium", "high-end", "affluent", "wealthy", "executive", "c-suite"},
		"middle_income":    {"affordable", "value", "mainstream", "middle class", "professional"},
		"budget_conscious": {"budget", "low-cost", "economical", "free", "cheap", "discount"},
	}

	for incomeGroup, indicators := range incomeIndicators {
		for _, indicator := range indicators {
			if strings.Contains(contentLower, indicator) {
				demographics.IncomeGroups = append(demographics.IncomeGroups, incomeGroup)
				result.Evidence = append(result.Evidence, fmt.Sprintf("Income group indicator: %s", indicator))
				break
			}
		}
	}

	// Education levels
	educationIndicators := map[string][]string{
		"higher_education": {"university", "college", "graduates", "phd", "masters", "degree", "academic"},
		"professional":     {"certified", "licensed", "qualified", "expert", "specialist", "trained"},
		"technical":        {"technical", "engineering", "developer", "programmer", "analyst", "scientist"},
	}

	for educationLevel, indicators := range educationIndicators {
		for _, indicator := range indicators {
			if strings.Contains(contentLower, indicator) {
				demographics.EducationLevels = append(demographics.EducationLevels, educationLevel)
				result.Evidence = append(result.Evidence, fmt.Sprintf("Education indicator: %s", indicator))
				break
			}
		}
	}

	// Profession types
	professionIndicators := map[string][]string{
		"technology": {"developers", "engineers", "programmers", "it", "tech", "software", "data"},
		"business":   {"managers", "executives", "analysts", "consultants", "sales", "marketing"},
		"healthcare": {"doctors", "nurses", "medical", "healthcare", "clinical", "pharmaceutical"},
		"finance":    {"financial", "banking", "investment", "accounting", "insurance", "fintech"},
		"education":  {"teachers", "educators", "academic", "research", "university", "school"},
		"creative":   {"designers", "artists", "creative", "media", "advertising", "content"},
	}

	for professionType, indicators := range professionIndicators {
		for _, indicator := range indicators {
			if strings.Contains(contentLower, indicator) {
				demographics.ProfessionTypes = append(demographics.ProfessionTypes, professionType)
				result.Evidence = append(result.Evidence, fmt.Sprintf("Profession indicator: %s", indicator))
				break
			}
		}
	}

	// Tech savviness
	if strings.Contains(contentLower, "technical") || strings.Contains(contentLower, "advanced") ||
		strings.Contains(contentLower, "expert") || strings.Contains(contentLower, "power user") {
		demographics.TechSavviness = "high"
	} else if strings.Contains(contentLower, "easy") || strings.Contains(contentLower, "simple") ||
		strings.Contains(contentLower, "user-friendly") || strings.Contains(contentLower, "intuitive") {
		demographics.TechSavviness = "low"
	} else {
		demographics.TechSavviness = "medium"
	}

	result.Demographics = *demographics
	return nil
}

// analyzeIndustries identifies target industries
func (aa *AudienceAnalyzer) analyzeIndustries(ctx context.Context, content string, result *AudienceResult) error {
	ctx, span := aa.tracer.Start(ctx, "audience_analyzer.analyze_industries")
	defer span.End()

	contentLower := strings.ToLower(content)
	industries := make(map[string]*Industry)

	// Technology sector
	techIndicators := []string{
		"technology", "software", "saas", "cloud", "ai", "machine learning",
		"data", "analytics", "cybersecurity", "fintech", "edtech", "healthtech",
		"developers", "engineers", "programming", "digital transformation",
	}

	techConfidence := 0.0
	for _, indicator := range techIndicators {
		if strings.Contains(contentLower, indicator) {
			techConfidence += 0.2
			result.Evidence = append(result.Evidence, fmt.Sprintf("Technology indicator: %s", indicator))
		}
	}

	if techConfidence > 0 {
		industries["technology"] = &Industry{
			Name:            "Technology",
			Sector:          "Information Technology",
			SubIndustries:   []string{"Software", "Cloud Services", "Data Analytics"},
			ConfidenceScore: minFloat64(techConfidence, 1.0),
		}
	}

	// Healthcare sector
	healthcareIndicators := []string{
		"healthcare", "medical", "hospital", "clinic", "pharmaceutical",
		"biotech", "health", "patient", "clinical", "medical devices",
		"telemedicine", "healthtech", "wellness", "therapy", "diagnosis",
	}

	healthcareConfidence := 0.0
	for _, indicator := range healthcareIndicators {
		if strings.Contains(contentLower, indicator) {
			healthcareConfidence += 0.2
			result.Evidence = append(result.Evidence, fmt.Sprintf("Healthcare indicator: %s", indicator))
		}
	}

	if healthcareConfidence > 0 {
		industries["healthcare"] = &Industry{
			Name:            "Healthcare",
			Sector:          "Healthcare & Life Sciences",
			SubIndustries:   []string{"Medical Devices", "Pharmaceuticals", "Digital Health"},
			ConfidenceScore: minFloat64(healthcareConfidence, 1.0),
		}
	}

	// Financial services
	financeIndicators := []string{
		"financial", "banking", "finance", "investment", "insurance",
		"fintech", "payments", "trading", "wealth", "credit", "loan",
		"accounting", "tax", "audit", "compliance", "risk management",
	}

	financeConfidence := 0.0
	for _, indicator := range financeIndicators {
		if strings.Contains(contentLower, indicator) {
			financeConfidence += 0.2
			result.Evidence = append(result.Evidence, fmt.Sprintf("Finance indicator: %s", indicator))
		}
	}

	if financeConfidence > 0 {
		industries["finance"] = &Industry{
			Name:            "Financial Services",
			Sector:          "Financial Services",
			SubIndustries:   []string{"Banking", "Investment", "Insurance", "Payments"},
			ConfidenceScore: minFloat64(financeConfidence, 1.0),
		}
	}

	// Education sector
	educationIndicators := []string{
		"education", "learning", "training", "school", "university",
		"edtech", "e-learning", "online courses", "students", "teachers",
		"academic", "curriculum", "assessment", "certification",
	}

	educationConfidence := 0.0
	for _, indicator := range educationIndicators {
		if strings.Contains(contentLower, indicator) {
			educationConfidence += 0.2
			result.Evidence = append(result.Evidence, fmt.Sprintf("Education indicator: %s", indicator))
		}
	}

	if educationConfidence > 0 {
		industries["education"] = &Industry{
			Name:            "Education",
			Sector:          "Education & Training",
			SubIndustries:   []string{"K-12", "Higher Education", "Corporate Training", "Online Learning"},
			ConfidenceScore: minFloat64(educationConfidence, 1.0),
		}
	}

	// Convert map to slice
	for _, industry := range industries {
		result.Industries = append(result.Industries, *industry)
	}

	return nil
}

// analyzeCompanySizes identifies target company sizes
func (aa *AudienceAnalyzer) analyzeCompanySizes(ctx context.Context, content string, result *AudienceResult) error {
	ctx, span := aa.tracer.Start(ctx, "audience_analyzer.analyze_company_sizes")
	defer span.End()

	contentLower := strings.ToLower(content)
	companySizes := make(map[string]bool)

	// Enterprise size indicators
	enterpriseIndicators := []string{
		"enterprise", "large companies", "corporations", "multinational",
		"fortune 500", "global", "enterprise-grade", "scalable",
	}

	for _, indicator := range enterpriseIndicators {
		if strings.Contains(contentLower, indicator) {
			companySizes["enterprise"] = true
			result.Evidence = append(result.Evidence, fmt.Sprintf("Enterprise size indicator: %s", indicator))
		}
	}

	// SME size indicators
	smeIndicators := []string{
		"small business", "medium business", "sme", "growing companies",
		"mid-size", "mid-market", "small-medium enterprise",
	}

	for _, indicator := range smeIndicators {
		if strings.Contains(contentLower, indicator) {
			companySizes["sme"] = true
			result.Evidence = append(result.Evidence, fmt.Sprintf("SME size indicator: %s", indicator))
		}
	}

	// Startup indicators
	startupIndicators := []string{
		"startup", "startups", "early-stage", "emerging companies",
		"scale-ups", "new companies", "young companies",
	}

	for _, indicator := range startupIndicators {
		if strings.Contains(contentLower, indicator) {
			companySizes["startup"] = true
			result.Evidence = append(result.Evidence, fmt.Sprintf("Startup size indicator: %s", indicator))
		}
	}

	// Convert map to slice
	for size := range companySizes {
		result.CompanySizes = append(result.CompanySizes, size)
	}

	return nil
}

// analyzeGeographicMarkets identifies geographic target markets
func (aa *AudienceAnalyzer) analyzeGeographicMarkets(ctx context.Context, content string, result *AudienceResult) error {
	ctx, span := aa.tracer.Start(ctx, "audience_analyzer.analyze_geographic_markets")
	defer span.End()

	contentLower := strings.ToLower(content)
	markets := make(map[string]bool)

	// Regional indicators
	regions := map[string][]string{
		"north_america": {"usa", "united states", "america", "canada", "north america", "us", "canadian"},
		"europe":        {"europe", "european", "eu", "uk", "germany", "france", "italy", "spain", "netherlands"},
		"asia_pacific":  {"asia", "pacific", "apac", "china", "japan", "india", "singapore", "australia"},
		"global":        {"global", "worldwide", "international", "multinational", "across regions"},
		"local":         {"local", "regional", "community", "neighborhood", "city", "state"},
	}

	for region, indicators := range regions {
		for _, indicator := range indicators {
			if strings.Contains(contentLower, indicator) {
				markets[region] = true
				result.Evidence = append(result.Evidence, fmt.Sprintf("Geographic indicator: %s", indicator))
				break
			}
		}
	}

	// Convert map to slice
	for market := range markets {
		result.GeographicMarkets = append(result.GeographicMarkets, market)
	}

	return nil
}

// analyzeBehavioralSegments identifies behavioral segments
func (aa *AudienceAnalyzer) analyzeBehavioralSegments(ctx context.Context, content string, result *AudienceResult) error {
	ctx, span := aa.tracer.Start(ctx, "audience_analyzer.analyze_behavioral_segments")
	defer span.End()

	contentLower := strings.ToLower(content)
	segments := make(map[string]bool)

	// Behavioral indicators
	behaviors := map[string][]string{
		"early_adopters":      {"early adopters", "innovators", "first movers", "cutting-edge", "latest", "new"},
		"mainstream":          {"mainstream", "popular", "widely used", "standard", "common", "typical"},
		"price_sensitive":     {"budget", "cost-effective", "affordable", "value", "cheap", "economical"},
		"quality_focused":     {"premium", "high-quality", "best-in-class", "excellence", "superior"},
		"convenience_seekers": {"easy", "simple", "quick", "fast", "convenient", "effortless"},
		"power_users":         {"advanced", "expert", "power user", "professional", "complex", "sophisticated"},
	}

	for segment, indicators := range behaviors {
		for _, indicator := range indicators {
			if strings.Contains(contentLower, indicator) {
				segments[segment] = true
				result.Evidence = append(result.Evidence, fmt.Sprintf("Behavioral indicator: %s", indicator))
				break
			}
		}
	}

	// Convert map to slice
	for segment := range segments {
		result.BehavioralSegments = append(result.BehavioralSegments, segment)
	}

	return nil
}

// generateCustomerPersonas creates detailed customer personas
func (aa *AudienceAnalyzer) generateCustomerPersonas(ctx context.Context, content string, result *AudienceResult) error {
	ctx, span := aa.tracer.Start(ctx, "audience_analyzer.generate_customer_personas")
	defer span.End()

	// Generate personas based on identified customer types and characteristics
	personas := []CustomerPersona{}

	for _, customerType := range result.CustomerTypes {
		persona := aa.createPersonaForType(customerType, result)
		if persona.ConfidenceScore > 0.3 {
			personas = append(personas, persona)
		}
	}

	result.CustomerPersonas = personas
	return nil
}

// createPersonaForType creates a persona for a specific customer type
func (aa *AudienceAnalyzer) createPersonaForType(customerType string, result *AudienceResult) CustomerPersona {
	switch customerType {
	case "enterprise":
		return CustomerPersona{
			Name:        "Enterprise Decision Maker",
			Type:        "enterprise",
			Description: "Senior executives and IT leaders in large organizations",
			Characteristics: []string{
				"Budget authority over $100K+",
				"Multiple stakeholders involved in decisions",
				"Risk-averse and compliance-focused",
				"Requires proven ROI and case studies",
			},
			Needs: []string{
				"Scalable solutions",
				"Enterprise-grade security",
				"Dedicated support",
				"Integration capabilities",
			},
			PainPoints: []string{
				"Complex procurement processes",
				"Integration challenges",
				"Change management resistance",
				"Vendor management overhead",
			},
			BuyingBehavior:  "Committee-based, long sales cycles, multiple touchpoints",
			DecisionMakers:  []string{"CTO", "CIO", "VP Engineering", "Procurement"},
			ConfidenceScore: 0.8,
		}

	case "sme":
		return CustomerPersona{
			Name:        "SME Business Owner",
			Type:        "sme",
			Description: "Founders and managers of small-medium enterprises",
			Characteristics: []string{
				"Limited budget but growth-focused",
				"Hands-on decision making",
				"Efficiency and ROI focused",
				"Technology adoption driven by necessity",
			},
			Needs: []string{
				"Cost-effective solutions",
				"Easy implementation",
				"Quick time to value",
				"Scalable as business grows",
			},
			PainPoints: []string{
				"Limited technical resources",
				"Budget constraints",
				"Time limitations",
				"Lack of specialized expertise",
			},
			BuyingBehavior:  "Direct decision making, price-sensitive, quick decisions",
			DecisionMakers:  []string{"Founder", "CEO", "Operations Manager"},
			ConfidenceScore: 0.7,
		}

	case "consumer":
		return CustomerPersona{
			Name:        "Individual Consumer",
			Type:        "consumer",
			Description: "Individual users seeking personal solutions",
			Characteristics: []string{
				"Personal budget considerations",
				"Individual decision making",
				"Convenience and usability focused",
				"Price and value conscious",
			},
			Needs: []string{
				"User-friendly interface",
				"Affordable pricing",
				"Quick problem resolution",
				"Personal value and benefits",
			},
			PainPoints: []string{
				"Complex interfaces",
				"High costs",
				"Poor customer support",
				"Lack of personalization",
			},
			BuyingBehavior:  "Individual research, trial-based, impulse and considered purchases",
			DecisionMakers:  []string{"Self", "Family members"},
			ConfidenceScore: 0.6,
		}

	case "professional":
		return CustomerPersona{
			Name:        "Professional User",
			Type:        "professional",
			Description: "Individual professionals and consultants",
			Characteristics: []string{
				"Professional budget allocation",
				"Expertise and efficiency focused",
				"Tool and productivity oriented",
				"Industry-specific needs",
			},
			Needs: []string{
				"Professional-grade features",
				"Productivity enhancement",
				"Industry-specific functionality",
				"Professional support",
			},
			PainPoints: []string{
				"Generic solutions",
				"Lack of advanced features",
				"Poor integration with workflow",
				"Inadequate support",
			},
			BuyingBehavior:  "Research-driven, feature comparison, trial periods",
			DecisionMakers:  []string{"Self", "Team lead", "Practice manager"},
			ConfidenceScore: 0.7,
		}

	default:
		return CustomerPersona{
			Name:            "General User",
			Type:            customerType,
			Description:     "General user segment",
			ConfidenceScore: 0.3,
		}
	}
}

// determinePrimaryAudience determines primary and secondary audiences
func (aa *AudienceAnalyzer) determinePrimaryAudience(result *AudienceResult) {
	if len(result.CustomerTypes) == 0 {
		result.PrimaryAudience = "unknown"
		return
	}

	// Priority order for primary audience
	priorityOrder := []string{"enterprise", "sme", "professional", "consumer", "marketplace_participant"}

	for _, priority := range priorityOrder {
		for _, customerType := range result.CustomerTypes {
			if customerType == priority {
				result.PrimaryAudience = customerType
				break
			}
		}
		if result.PrimaryAudience != "" {
			break
		}
	}

	// Set secondary audience
	if len(result.CustomerTypes) > 1 {
		for _, customerType := range result.CustomerTypes {
			if customerType != result.PrimaryAudience {
				result.SecondaryAudience = customerType
				break
			}
		}
	}

	if result.PrimaryAudience == "" {
		result.PrimaryAudience = result.CustomerTypes[0]
	}
}

// calculateConfidenceScores calculates confidence scores for all components
func (aa *AudienceAnalyzer) calculateConfidenceScores(result *AudienceResult) {
	// Calculate component scores
	result.ComponentScores.DemographicScore = aa.calculateDemographicScore(result)
	result.ComponentScores.IndustryScore = aa.calculateIndustryScore(result)
	result.ComponentScores.SizeScore = aa.calculateSizeScore(result)
	result.ComponentScores.GeographicScore = aa.calculateGeographicScore(result)
	result.ComponentScores.BehavioralScore = aa.calculateBehavioralScore(result)
	result.ComponentScores.PersonaScore = aa.calculatePersonaScore(result)

	// Calculate overall confidence
	scores := []float64{
		result.ComponentScores.DemographicScore * aa.config.DemographicWeight,
		result.ComponentScores.IndustryScore * aa.config.IndustryWeight,
		result.ComponentScores.SizeScore * aa.config.SizeWeight,
		result.ComponentScores.GeographicScore * aa.config.GeographicWeight,
		result.ComponentScores.BehavioralScore * aa.config.BehavioralWeight,
	}

	totalScore := 0.0
	for _, score := range scores {
		totalScore += score
	}

	result.ConfidenceScore = totalScore
}

// Helper functions for component scoring
func (aa *AudienceAnalyzer) calculateDemographicScore(result *AudienceResult) float64 {
	score := 0.0
	demographics := result.Demographics

	if len(demographics.AgeGroups) > 0 {
		score += 0.2
	}
	if len(demographics.IncomeGroups) > 0 {
		score += 0.2
	}
	if len(demographics.EducationLevels) > 0 {
		score += 0.2
	}
	if len(demographics.ProfessionTypes) > 0 {
		score += 0.2
	}
	if demographics.TechSavviness != "" {
		score += 0.2
	}

	return score
}

func (aa *AudienceAnalyzer) calculateIndustryScore(result *AudienceResult) float64 {
	if len(result.Industries) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, industry := range result.Industries {
		totalConfidence += industry.ConfidenceScore
	}

	return minFloat64(totalConfidence/float64(len(result.Industries)), 1.0)
}

func (aa *AudienceAnalyzer) calculateSizeScore(result *AudienceResult) float64 {
	return float64(len(result.CompanySizes)) * 0.33
}

func (aa *AudienceAnalyzer) calculateGeographicScore(result *AudienceResult) float64 {
	return minFloat64(float64(len(result.GeographicMarkets))*0.25, 1.0)
}

func (aa *AudienceAnalyzer) calculateBehavioralScore(result *AudienceResult) float64 {
	return minFloat64(float64(len(result.BehavioralSegments))*0.2, 1.0)
}

func (aa *AudienceAnalyzer) calculatePersonaScore(result *AudienceResult) float64 {
	if len(result.CustomerPersonas) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, persona := range result.CustomerPersonas {
		totalConfidence += persona.ConfidenceScore
	}

	return totalConfidence / float64(len(result.CustomerPersonas))
}

// validateInput validates input parameters
func (aa *AudienceAnalyzer) validateInput(content string) error {
	if len(content) < aa.config.MinContentLength {
		return fmt.Errorf("content too short: %d characters (minimum: %d)", len(content), aa.config.MinContentLength)
	}
	return nil
}

// validateResult validates analysis results
func (aa *AudienceAnalyzer) validateResult(result *AudienceResult) {
	status := ValidationStatus{
		IsValid:          true,
		ValidationErrors: []string{},
		LastValidated:    time.Now(),
	}

	// Check confidence threshold
	if result.ConfidenceScore < aa.config.MinConfidenceThreshold {
		status.IsValid = false
		status.ValidationErrors = append(status.ValidationErrors, fmt.Sprintf("Confidence score %.2f below threshold %.2f",
			result.ConfidenceScore, aa.config.MinConfidenceThreshold))
	}

	// Check evidence count
	if len(result.Evidence) < aa.config.MinEvidenceCount {
		status.IsValid = false
		status.ValidationErrors = append(status.ValidationErrors, fmt.Sprintf("Insufficient evidence: %d items (minimum: %d)",
			len(result.Evidence), aa.config.MinEvidenceCount))
	}

	// Check primary audience
	if result.PrimaryAudience == "" || result.PrimaryAudience == "unknown" {
		status.IsValid = false
		status.ValidationErrors = append(status.ValidationErrors, "No primary audience identified")
	}

	result.IsValidated = status.IsValid
	result.ValidationStatus = status
}

// calculateDataQuality calculates data quality score
func (aa *AudienceAnalyzer) calculateDataQuality(result *AudienceResult) float64 {
	qualityFactors := []float64{
		aa.calculateEvidenceQuality(result),
		aa.calculateAnalysisCompleteness(result),
		aa.calculatePersonaQuality(result),
	}

	total := 0.0
	for _, factor := range qualityFactors {
		total += factor
	}

	return total / float64(len(qualityFactors))
}

func (aa *AudienceAnalyzer) calculateEvidenceQuality(result *AudienceResult) float64 {
	if len(result.Evidence) == 0 {
		return 0.0
	}

	return minFloat64(float64(len(result.Evidence))*0.1, 1.0)
}

func (aa *AudienceAnalyzer) calculateAnalysisCompleteness(result *AudienceResult) float64 {
	completeness := 0.0

	if len(result.CustomerTypes) > 0 {
		completeness += 0.2
	}
	if len(result.Demographics.AgeGroups) > 0 || len(result.Demographics.ProfessionTypes) > 0 {
		completeness += 0.2
	}
	if len(result.Industries) > 0 {
		completeness += 0.2
	}
	if len(result.CompanySizes) > 0 {
		completeness += 0.2
	}
	if len(result.BehavioralSegments) > 0 {
		completeness += 0.2
	}

	return completeness
}

func (aa *AudienceAnalyzer) calculatePersonaQuality(result *AudienceResult) float64 {
	if len(result.CustomerPersonas) == 0 {
		return 0.0
	}

	qualitySum := 0.0
	for _, persona := range result.CustomerPersonas {
		quality := 0.0
		if len(persona.Characteristics) > 0 {
			quality += 0.25
		}
		if len(persona.Needs) > 0 {
			quality += 0.25
		}
		if len(persona.PainPoints) > 0 {
			quality += 0.25
		}
		if persona.BuyingBehavior != "" {
			quality += 0.25
		}
		qualitySum += quality
	}

	return qualitySum / float64(len(result.CustomerPersonas))
}

// generateReasoning generates human-readable reasoning
func (aa *AudienceAnalyzer) generateReasoning(result *AudienceResult) string {
	if result.PrimaryAudience == "" {
		return "No clear audience indicators found in the analyzed content."
	}

	reasoning := fmt.Sprintf("Primary audience identified as '%s' with %.1f%% confidence. ",
		result.PrimaryAudience, result.ConfidenceScore*100)

	if len(result.CustomerTypes) > 1 {
		reasoning += fmt.Sprintf("Multiple customer types detected: %v. ", result.CustomerTypes)
	}

	if len(result.Industries) > 0 {
		reasoning += fmt.Sprintf("Target industries include: %v. ", aa.getIndustryNames(result.Industries))
	}

	if len(result.Demographics.ProfessionTypes) > 0 {
		reasoning += fmt.Sprintf("Professional focus areas: %v. ", result.Demographics.ProfessionTypes)
	}

	if len(result.BehavioralSegments) > 0 {
		reasoning += fmt.Sprintf("Behavioral segments: %v. ", result.BehavioralSegments)
	}

	reasoning += fmt.Sprintf("Analysis based on %d pieces of evidence.", len(result.Evidence))

	return reasoning
}

func (aa *AudienceAnalyzer) getIndustryNames(industries []Industry) []string {
	names := make([]string, len(industries))
	for i, industry := range industries {
		names[i] = industry.Name
	}
	return names
}

// Helper function for float64 minimum
func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
