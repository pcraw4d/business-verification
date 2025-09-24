package enrichment

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewAudienceAnalyzer(t *testing.T) {
	tests := []struct {
		name     string
		config   *AudienceConfig
		logger   *zap.Logger
		expected *AudienceAnalyzer
	}{
		{
			name:   "with nil inputs",
			config: nil,
			logger: nil,
		},
		{
			name: "with custom config",
			config: &AudienceConfig{
				MinConfidenceThreshold: 0.5,
				MinEvidenceCount:       3,
			},
			logger: zap.NewNop(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAudienceAnalyzer(tt.config, tt.logger)
			assert.NotNil(t, analyzer)
			assert.NotNil(t, analyzer.config)
			assert.NotNil(t, analyzer.logger)
			assert.NotNil(t, analyzer.tracer)

			if tt.config != nil {
				assert.Equal(t, tt.config.MinConfidenceThreshold, analyzer.config.MinConfidenceThreshold)
			}
		})
	}
}

func TestAudienceAnalyzer_AnalyzeAudience(t *testing.T) {
	analyzer := NewAudienceAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name                    string
		content                 string
		sourceURL               string
		expectedPrimaryAudience string
		expectedMinConfidence   float64
		expectedMaxConfidence   float64
	}{
		{
			name: "Enterprise B2B software",
			content: `We provide enterprise software solutions for large corporations and 
			multinational companies. Our platform offers enterprise-grade security, 
			scalability, and dedicated support for Fortune 500 clients. Perfect for 
			IT leaders, CTOs, and enterprise decision makers.`,
			sourceURL:               "https://enterprise-software.com",
			expectedPrimaryAudience: "enterprise",
			expectedMinConfidence:   0.2,
			expectedMaxConfidence:   0.4,
		},
		{
			name: "Consumer marketplace",
			content: `Join our marketplace where individuals and families can buy and sell 
			products directly. Perfect for consumers, households, and everyday users 
			looking for great deals. Easy to use for regular people.`,
			sourceURL:               "https://consumer-marketplace.com",
			expectedPrimaryAudience: "consumer",
			expectedMinConfidence:   0.1,
			expectedMaxConfidence:   0.3,
		},
		{
			name: "SME business tools",
			content: `Designed for small businesses and growing companies. Our affordable 
			solution helps SMEs and startups streamline operations. Perfect for 
			small business owners and entrepreneurs.`,
			sourceURL:               "https://sme-tools.com",
			expectedPrimaryAudience: "sme",
			expectedMinConfidence:   0.2,
			expectedMaxConfidence:   0.5,
		},
		{
			name: "Professional services",
			content: `Tools for professionals, consultants, and freelancers. Designed 
			for experts and specialists who need advanced functionality and 
			professional-grade features.`,
			sourceURL:               "https://professional-tools.com",
			expectedPrimaryAudience: "professional",
			expectedMinConfidence:   0.2,
			expectedMaxConfidence:   0.4,
		},
		{
			name: "Marketplace platform",
			content: `Our platform connects buyers and sellers in a thriving marketplace 
			ecosystem. Join our community of vendors, merchants, and platform users. 
			Perfect for marketplace participants and network members.`,
			sourceURL:               "https://marketplace-platform.com",
			expectedPrimaryAudience: "marketplace_participant",
			expectedMinConfidence:   0.1,
			expectedMaxConfidence:   0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeAudience(context.Background(), tt.content, tt.sourceURL)
			require.NoError(t, err)
			assert.NotNil(t, result)

			// Verify primary audience
			assert.Equal(t, tt.expectedPrimaryAudience, result.PrimaryAudience)
			assert.True(t, result.ConfidenceScore >= tt.expectedMinConfidence,
				"Expected confidence >= %f, got %f", tt.expectedMinConfidence, result.ConfidenceScore)
			assert.True(t, result.ConfidenceScore <= tt.expectedMaxConfidence,
				"Expected confidence <= %f, got %f", tt.expectedMaxConfidence, result.ConfidenceScore)

			// Verify structure
			assert.NotEmpty(t, result.CustomerTypes)
			assert.NotEmpty(t, result.Evidence)
			assert.Equal(t, tt.sourceURL, result.SourceURL)
			assert.NotZero(t, result.AnalyzedAt)
			assert.NotZero(t, result.ProcessingTime)
		})
	}
}

func TestAudienceAnalyzer_analyzeCustomerTypes(t *testing.T) {
	analyzer := NewAudienceAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name             string
		content          string
		expectedTypes    []string
		minEvidenceCount int
	}{
		{
			name: "Enterprise indicators",
			content: `We serve enterprise clients, large corporations, and Fortune 500 
			companies with enterprise-grade solutions for multinational organizations.`,
			expectedTypes:    []string{"enterprise"},
			minEvidenceCount: 3,
		},
		{
			name: "SME indicators",
			content: `Perfect for small businesses, SMEs, and growing startups. Designed 
			for mid-sized companies and emerging scale-ups.`,
			expectedTypes:    []string{"sme"},
			minEvidenceCount: 3,
		},
		{
			name: "Consumer indicators",
			content: `Great for consumers, individuals, families, and households. Perfect 
			for personal use and everyday users from the general public.`,
			expectedTypes:    []string{"consumer"},
			minEvidenceCount: 3,
		},
		{
			name: "Professional indicators",
			content: `Designed for professionals, experts, consultants, and freelancers. 
			Perfect for specialists and practitioners in professional services.`,
			expectedTypes:    []string{"professional"},
			minEvidenceCount: 3,
		},
		{
			name: "Marketplace indicators",
			content: `Our platform connects buyers and sellers in a thriving marketplace 
			community. Join our ecosystem of vendors and platform participants.`,
			expectedTypes:    []string{"marketplace_participant"},
			minEvidenceCount: 3,
		},
		{
			name: "Mixed indicators",
			content: `We serve both enterprise clients and small businesses, connecting 
			professionals with consumers in our marketplace platform.`,
			expectedTypes:    []string{"enterprise", "sme", "professional", "consumer", "marketplace_participant"},
			minEvidenceCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &AudienceResult{
				Evidence:         []string{},
				ExtractedPhrases: []string{},
			}

			err := analyzer.analyzeCustomerTypes(context.Background(), tt.content, result)
			require.NoError(t, err)

			// Check if all expected types are present
			for _, expectedType := range tt.expectedTypes {
				assert.Contains(t, result.CustomerTypes, expectedType,
					"Expected customer type %s not found", expectedType)
			}

			// Check evidence count
			assert.GreaterOrEqual(t, len(result.Evidence), tt.minEvidenceCount,
				"Expected at least %d evidence items, got %d", tt.minEvidenceCount, len(result.Evidence))
		})
	}
}

func TestAudienceAnalyzer_analyzeDemographics(t *testing.T) {
	analyzer := NewAudienceAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name               string
		content            string
		expectedAgeGroups  []string
		expectedIncome     []string
		expectedEducation  []string
		expectedProfession []string
		expectedTechSavvy  string
	}{
		{
			name: "Young professionals",
			content: `Perfect for millennials and young adults in their 20s and 30s. 
			Designed for university students and college graduates with technical 
			backgrounds in software development.`,
			expectedAgeGroups:  []string{"young_adults"},
			expectedEducation:  []string{"higher_education", "technical"},
			expectedProfession: []string{"technology"},
			expectedTechSavvy:  "high",
		},
		{
			name: "Enterprise executives",
			content: `Designed for C-suite executives and senior professionals in their 
			40s and 50s. Perfect for affluent business leaders and certified experts 
			with advanced business expertise.`,
			expectedAgeGroups:  []string{"middle_aged"},
			expectedIncome:     []string{"high_income"},
			expectedEducation:  []string{"professional"},
			expectedProfession: []string{"business"},
			expectedTechSavvy:  "high",
		},
		{
			name: "Budget-conscious families",
			content: `Affordable solution for families with children and budget-conscious 
			households. Easy and simple to use for parents and everyday users.`,
			expectedAgeGroups: []string{"families"},
			expectedIncome:    []string{"budget_conscious"},
			expectedTechSavvy: "low",
		},
		{
			name: "Healthcare professionals",
			content: `Designed for doctors, nurses, and medical professionals. Perfect 
			for healthcare experts and clinical specialists in pharmaceutical companies.`,
			expectedEducation:  []string{"professional"},
			expectedProfession: []string{"healthcare"},
			expectedTechSavvy:  "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &AudienceResult{
				Evidence: []string{},
			}

			err := analyzer.analyzeDemographics(context.Background(), tt.content, result)
			require.NoError(t, err)

			demographics := result.Demographics

			// Check age groups
			for _, expected := range tt.expectedAgeGroups {
				assert.Contains(t, demographics.AgeGroups, expected,
					"Expected age group %s not found", expected)
			}

			// Check income groups
			for _, expected := range tt.expectedIncome {
				assert.Contains(t, demographics.IncomeGroups, expected,
					"Expected income group %s not found", expected)
			}

			// Check education levels
			for _, expected := range tt.expectedEducation {
				assert.Contains(t, demographics.EducationLevels, expected,
					"Expected education level %s not found", expected)
			}

			// Check profession types
			for _, expected := range tt.expectedProfession {
				assert.Contains(t, demographics.ProfessionTypes, expected,
					"Expected profession type %s not found", expected)
			}

			// Check tech savviness
			if tt.expectedTechSavvy != "" {
				assert.Equal(t, tt.expectedTechSavvy, demographics.TechSavviness)
			}
		})
	}
}

func TestAudienceAnalyzer_analyzeIndustries(t *testing.T) {
	analyzer := NewAudienceAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name               string
		content            string
		expectedIndustries []string
		minConfidence      float64
	}{
		{
			name: "Technology sector",
			content: `We provide software solutions, cloud services, and AI-powered 
			analytics for developers and engineers. Our SaaS platform offers 
			cybersecurity and data analytics for digital transformation.`,
			expectedIndustries: []string{"Technology"},
			minConfidence:      0.2,
		},
		{
			name: "Healthcare sector",
			content: `Our healthcare platform serves hospitals, clinics, and medical 
			professionals. We offer pharmaceutical solutions, biotech services, 
			and telemedicine for patient care and clinical diagnosis.`,
			expectedIndustries: []string{"Healthcare"},
			minConfidence:      0.2,
		},
		{
			name: "Financial services",
			content: `We provide banking solutions, investment tools, and fintech 
			services. Our platform offers payment processing, trading capabilities, 
			and risk management for financial institutions.`,
			expectedIndustries: []string{"Financial Services"},
			minConfidence:      0.2,
		},
		{
			name: "Education sector",
			content: `Our edtech platform serves schools, universities, and students. 
			We offer e-learning solutions, online courses, and training programs 
			for academic and corporate learning.`,
			expectedIndustries: []string{"Education"},
			minConfidence:      0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &AudienceResult{
				Evidence: []string{},
			}

			err := analyzer.analyzeIndustries(context.Background(), tt.content, result)
			require.NoError(t, err)

			// Check if expected industries are present
			industryNames := analyzer.getIndustryNames(result.Industries)
			for _, expected := range tt.expectedIndustries {
				assert.Contains(t, industryNames, expected,
					"Expected industry %s not found", expected)
			}

			// Check confidence scores
			for _, industry := range result.Industries {
				assert.GreaterOrEqual(t, industry.ConfidenceScore, tt.minConfidence,
					"Industry %s confidence %.2f below threshold %.2f",
					industry.Name, industry.ConfidenceScore, tt.minConfidence)
			}
		})
	}
}

func TestAudienceAnalyzer_analyzeCompanySizes(t *testing.T) {
	analyzer := NewAudienceAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name          string
		content       string
		expectedSizes []string
	}{
		{
			name: "Enterprise focus",
			content: `We serve enterprise clients, large corporations, and Fortune 500 
			companies. Our enterprise-grade solutions are perfect for multinational 
			organizations and global corporations.`,
			expectedSizes: []string{"enterprise"},
		},
		{
			name: "SME focus",
			content: `Perfect for small businesses, SMEs, and medium-sized companies. 
			Designed for growing businesses and mid-market organizations.`,
			expectedSizes: []string{"sme"},
		},
		{
			name: "Startup focus",
			content: `Ideal for startups, early-stage companies, and emerging scale-ups. 
			Perfect for new companies and young businesses.`,
			expectedSizes: []string{"startup"},
		},
		{
			name: "Mixed sizes",
			content: `We serve enterprise clients, small businesses, and emerging startups. 
			Perfect for companies of all sizes from SMEs to large corporations.`,
			expectedSizes: []string{"enterprise", "sme", "startup"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &AudienceResult{
				Evidence: []string{},
			}

			err := analyzer.analyzeCompanySizes(context.Background(), tt.content, result)
			require.NoError(t, err)

			// Check if all expected sizes are present
			for _, expectedSize := range tt.expectedSizes {
				assert.Contains(t, result.CompanySizes, expectedSize,
					"Expected company size %s not found", expectedSize)
			}
		})
	}
}

func TestAudienceAnalyzer_analyzeGeographicMarkets(t *testing.T) {
	analyzer := NewAudienceAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name            string
		content         string
		expectedMarkets []string
	}{
		{
			name: "North American market",
			content: `We serve clients across the USA, Canada, and North America. 
			Perfect for American and Canadian businesses.`,
			expectedMarkets: []string{"north_america"},
		},
		{
			name: "European market",
			content: `Available across Europe including UK, Germany, France, and EU. 
			Designed for European businesses and organizations.`,
			expectedMarkets: []string{"europe"},
		},
		{
			name: "Global market",
			content: `We operate worldwide with international reach across all regions. 
			Our global platform serves multinational companies globally.`,
			expectedMarkets: []string{"global"},
		},
		{
			name: "Multiple markets",
			content: `Available in USA, Europe, Asia Pacific, and globally. We serve 
			local and international clients across all regions.`,
			expectedMarkets: []string{"north_america", "europe", "asia_pacific", "global", "local"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &AudienceResult{
				Evidence: []string{},
			}

			err := analyzer.analyzeGeographicMarkets(context.Background(), tt.content, result)
			require.NoError(t, err)

			// Check if all expected markets are present
			for _, expectedMarket := range tt.expectedMarkets {
				assert.Contains(t, result.GeographicMarkets, expectedMarket,
					"Expected geographic market %s not found", expectedMarket)
			}
		})
	}
}

func TestAudienceAnalyzer_analyzeBehavioralSegments(t *testing.T) {
	analyzer := NewAudienceAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name             string
		content          string
		expectedSegments []string
	}{
		{
			name: "Early adopters",
			content: `Perfect for early adopters and innovators who want the latest 
			cutting-edge technology. Designed for first movers and new solutions.`,
			expectedSegments: []string{"early_adopters"},
		},
		{
			name: "Budget conscious",
			content: `Affordable and cost-effective solution for budget-conscious users. 
			Economical and cheap option with great value.`,
			expectedSegments: []string{"price_sensitive"},
		},
		{
			name: "Quality focused",
			content: `Premium, high-quality solution for excellence-focused users. 
			Best-in-class and superior quality for demanding customers.`,
			expectedSegments: []string{"quality_focused"},
		},
		{
			name: "Power users",
			content: `Advanced features for expert power users and sophisticated 
			professionals. Complex functionality for technical experts.`,
			expectedSegments: []string{"power_users"},
		},
		{
			name: "Convenience seekers",
			content: `Easy, simple, and quick solution for convenient users. 
			Fast and effortless experience for busy professionals.`,
			expectedSegments: []string{"convenience_seekers"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &AudienceResult{
				Evidence: []string{},
			}

			err := analyzer.analyzeBehavioralSegments(context.Background(), tt.content, result)
			require.NoError(t, err)

			// Check if all expected segments are present
			for _, expectedSegment := range tt.expectedSegments {
				assert.Contains(t, result.BehavioralSegments, expectedSegment,
					"Expected behavioral segment %s not found", expectedSegment)
			}
		})
	}
}

func TestAudienceAnalyzer_generateCustomerPersonas(t *testing.T) {
	analyzer := NewAudienceAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name                 string
		customerTypes        []string
		expectedPersonaTypes []string
		minPersonaCount      int
	}{
		{
			name:                 "Enterprise personas",
			customerTypes:        []string{"enterprise"},
			expectedPersonaTypes: []string{"enterprise"},
			minPersonaCount:      1,
		},
		{
			name:                 "Mixed personas",
			customerTypes:        []string{"enterprise", "sme", "consumer"},
			expectedPersonaTypes: []string{"enterprise", "sme", "consumer"},
			minPersonaCount:      3,
		},
		{
			name:                 "Professional personas",
			customerTypes:        []string{"professional"},
			expectedPersonaTypes: []string{"professional"},
			minPersonaCount:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &AudienceResult{
				CustomerTypes: tt.customerTypes,
			}

			err := analyzer.generateCustomerPersonas(context.Background(), "", result)
			require.NoError(t, err)

			// Check persona count
			assert.GreaterOrEqual(t, len(result.CustomerPersonas), tt.minPersonaCount,
				"Expected at least %d personas, got %d", tt.minPersonaCount, len(result.CustomerPersonas))

			// Check persona types
			personaTypes := make([]string, len(result.CustomerPersonas))
			for i, persona := range result.CustomerPersonas {
				personaTypes[i] = persona.Type
			}

			for _, expectedType := range tt.expectedPersonaTypes {
				assert.Contains(t, personaTypes, expectedType,
					"Expected persona type %s not found", expectedType)
			}

			// Verify persona structure
			for _, persona := range result.CustomerPersonas {
				assert.NotEmpty(t, persona.Name)
				assert.NotEmpty(t, persona.Type)
				assert.NotEmpty(t, persona.Description)
				assert.GreaterOrEqual(t, persona.ConfidenceScore, 0.0)
				assert.LessOrEqual(t, persona.ConfidenceScore, 1.0)
			}
		})
	}
}

func TestAudienceAnalyzer_determinePrimaryAudience(t *testing.T) {
	analyzer := NewAudienceAnalyzer(nil, zap.NewNop())

	tests := []struct {
		name              string
		customerTypes     []string
		expectedPrimary   string
		expectedSecondary string
	}{
		{
			name:              "Enterprise first",
			customerTypes:     []string{"enterprise", "sme", "consumer"},
			expectedPrimary:   "enterprise",
			expectedSecondary: "sme",
		},
		{
			name:              "SME priority",
			customerTypes:     []string{"sme", "consumer", "professional"},
			expectedPrimary:   "sme",
			expectedSecondary: "consumer",
		},
		{
			name:              "Single type",
			customerTypes:     []string{"consumer"},
			expectedPrimary:   "consumer",
			expectedSecondary: "",
		},
		{
			name:              "No types",
			customerTypes:     []string{},
			expectedPrimary:   "unknown",
			expectedSecondary: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &AudienceResult{
				CustomerTypes: tt.customerTypes,
			}

			analyzer.determinePrimaryAudience(result)

			assert.Equal(t, tt.expectedPrimary, result.PrimaryAudience)
			assert.Equal(t, tt.expectedSecondary, result.SecondaryAudience)
		})
	}
}

func TestAudienceAnalyzer_validateResult(t *testing.T) {
	tests := []struct {
		name          string
		config        *AudienceConfig
		result        *AudienceResult
		expectedValid bool
	}{
		{
			name: "Valid result",
			config: &AudienceConfig{
				MinConfidenceThreshold: 0.3,
				MinEvidenceCount:       2,
			},
			result: &AudienceResult{
				PrimaryAudience: "enterprise",
				ConfidenceScore: 0.8,
				Evidence:        []string{"evidence1", "evidence2", "evidence3"},
			},
			expectedValid: true,
		},
		{
			name: "Low confidence",
			config: &AudienceConfig{
				MinConfidenceThreshold: 0.5,
				MinEvidenceCount:       2,
			},
			result: &AudienceResult{
				PrimaryAudience: "enterprise",
				ConfidenceScore: 0.3,
				Evidence:        []string{"evidence1", "evidence2"},
			},
			expectedValid: false,
		},
		{
			name: "Insufficient evidence",
			config: &AudienceConfig{
				MinConfidenceThreshold: 0.3,
				MinEvidenceCount:       3,
			},
			result: &AudienceResult{
				PrimaryAudience: "enterprise",
				ConfidenceScore: 0.8,
				Evidence:        []string{"evidence1"},
			},
			expectedValid: false,
		},
		{
			name: "No primary audience",
			config: &AudienceConfig{
				MinConfidenceThreshold: 0.3,
				MinEvidenceCount:       2,
			},
			result: &AudienceResult{
				PrimaryAudience: "unknown",
				ConfidenceScore: 0.8,
				Evidence:        []string{"evidence1", "evidence2"},
			},
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAudienceAnalyzer(tt.config, zap.NewNop())
			analyzer.validateResult(tt.result)

			assert.Equal(t, tt.expectedValid, tt.result.IsValidated)
			assert.Equal(t, tt.expectedValid, tt.result.ValidationStatus.IsValid)

			if !tt.expectedValid {
				assert.NotEmpty(t, tt.result.ValidationStatus.ValidationErrors)
			}
		})
	}
}

func TestAudienceAnalyzer_Integration(t *testing.T) {
	analyzer := NewAudienceAnalyzer(nil, zap.NewNop())

	// Test with comprehensive enterprise content
	content := `We are an enterprise software company serving Fortune 500 corporations 
	and large multinational organizations. Our platform offers enterprise-grade 
	security, scalability, and advanced analytics for IT leaders, CTOs, and 
	technical decision makers. Perfect for developers, engineers, and technology 
	professionals in the software industry. We operate globally across North America, 
	Europe, and Asia Pacific with premium, high-quality solutions for sophisticated 
	enterprise customers.`

	result, err := analyzer.AnalyzeAudience(context.Background(), content, "https://enterprise-software.com")
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify primary classification
	assert.Equal(t, "enterprise", result.PrimaryAudience)
	assert.True(t, result.ConfidenceScore > 0.5, "Expected high confidence, got %f", result.ConfidenceScore)

	// Verify comprehensive analysis
	assert.Contains(t, result.CustomerTypes, "enterprise")
	assert.NotEmpty(t, result.Demographics.ProfessionTypes)
	assert.Contains(t, result.Demographics.ProfessionTypes, "technology")
	assert.NotEmpty(t, result.Industries)
	assert.Contains(t, result.CompanySizes, "enterprise")
	assert.NotEmpty(t, result.GeographicMarkets)
	assert.NotEmpty(t, result.BehavioralSegments)

	// Verify personas
	assert.NotEmpty(t, result.CustomerPersonas)
	enterprisePersona := result.CustomerPersonas[0]
	assert.Equal(t, "enterprise", enterprisePersona.Type)
	assert.NotEmpty(t, enterprisePersona.Characteristics)
	assert.NotEmpty(t, enterprisePersona.Needs)
	assert.NotEmpty(t, enterprisePersona.PainPoints)

	// Verify evidence and quality
	assert.NotEmpty(t, result.Evidence)
	assert.True(t, result.DataQualityScore > 0.5)
	assert.NotEmpty(t, result.Reasoning)

	// Verify metadata
	assert.Equal(t, "https://enterprise-software.com", result.SourceURL)
	assert.NotZero(t, result.AnalyzedAt)
	assert.NotZero(t, result.ProcessingTime)
}

func TestAudienceAnalyzer_Performance(t *testing.T) {
	analyzer := NewAudienceAnalyzer(nil, zap.NewNop())

	// Create large content for performance testing
	content := `We are an enterprise software company serving Fortune 500 corporations. ` +
		strings.Repeat("Our platform offers enterprise-grade solutions for large organizations. ", 500)

	result, err := analyzer.AnalyzeAudience(context.Background(), content, "https://test.com")
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify performance
	assert.True(t, result.ProcessingTime.Milliseconds() < 1000,
		"Processing took too long: %v", result.ProcessingTime)

	// Verify analysis still works with large content
	assert.Equal(t, "enterprise", result.PrimaryAudience)
	assert.NotEmpty(t, result.Evidence)
}

func TestGetDefaultAudienceConfig(t *testing.T) {
	config := GetDefaultAudienceConfig()
	assert.NotNil(t, config)
	assert.Equal(t, 0.3, config.MinConfidenceThreshold)
	assert.Equal(t, 2, config.MinEvidenceCount)
	assert.Equal(t, 50, config.MinContentLength)
	assert.True(t, config.RequireMultipleIndicators)
	assert.True(t, config.EnableFallbackAnalysis)
	assert.True(t, config.ValidatePersonas)
}

func TestMinFloat64Helper(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		b        float64
		expected float64
	}{
		{"a smaller", 1.0, 2.0, 1.0},
		{"b smaller", 3.0, 2.0, 2.0},
		{"equal", 1.5, 1.5, 1.5},
		{"negative", -1.0, 0.5, -1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := minFloat64(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
