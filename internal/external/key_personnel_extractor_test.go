package external

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewKeyPersonnelExtractor(t *testing.T) {
	logger := zap.NewNop()

	t.Run("with default config", func(t *testing.T) {
		extractor := NewKeyPersonnelExtractor(nil, logger)
		assert.NotNil(t, extractor)
		assert.NotNil(t, extractor.config)
		assert.True(t, extractor.config.EnableExecutiveExtraction)
		assert.True(t, extractor.config.EnableTeamExtraction)
		assert.Equal(t, 0.6, extractor.config.MinConfidenceThreshold)
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &KeyPersonnelConfig{
			EnableExecutiveExtraction: false,
			EnableTeamExtraction:      true,
			MinConfidenceThreshold:    0.8,
		}

		extractor := NewKeyPersonnelExtractor(config, logger)
		assert.NotNil(t, extractor)
		assert.False(t, extractor.config.EnableExecutiveExtraction)
		assert.True(t, extractor.config.EnableTeamExtraction)
		assert.Equal(t, 0.8, extractor.config.MinConfidenceThreshold)
	})

	t.Run("with nil logger", func(t *testing.T) {
		extractor := NewKeyPersonnelExtractor(nil, nil)
		assert.NotNil(t, extractor)
		assert.NotNil(t, extractor.logger)
	})
}

func TestExtractKeyPersonnel(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewKeyPersonnelExtractor(nil, logger)
	ctx := context.Background()

	t.Run("extract executives", func(t *testing.T) {
		content := `
			Our leadership team includes John Smith, CEO and founder of the company.
			Sarah Johnson serves as our CTO, leading all technical initiatives.
			Michael Brown is our CFO, managing financial operations.
		`

		result, err := extractor.ExtractKeyPersonnel(ctx, content, "https://example.com/team")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.GreaterOrEqual(t, len(result.Executives), 2)
		assert.Equal(t, "https://example.com/team", result.SourceURL)
		assert.Greater(t, result.ConfidenceScore, 0.0)

		// Check that executives are properly identified
		foundCEO := false
		foundCTO := false
		for _, exec := range result.Executives {
			if strings.Contains(strings.ToLower(exec.Title), "ceo") {
				foundCEO = true
				assert.Equal(t, "executive", exec.Level)
				assert.Greater(t, exec.ConfidenceScore, 0.8)
			}
			if strings.Contains(strings.ToLower(exec.Title), "cto") {
				foundCTO = true
				assert.Equal(t, "executive", exec.Level)
			}
		}
		assert.True(t, foundCEO)
		assert.True(t, foundCTO)
	})

	t.Run("extract senior management", func(t *testing.T) {
		content := `
			Our VP of Engineering is David Wilson, leading our development team.
			Lisa Chen is our Director of Marketing, overseeing all marketing initiatives.
			Robert Davis serves as Head of Sales, driving revenue growth.
		`

		result, err := extractor.ExtractKeyPersonnel(ctx, content, "https://example.com/team")
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(result.SeniorManagement), 2)

		// Check that senior management is properly identified
		foundVP := false
		foundDirector := false
		for _, senior := range result.SeniorManagement {
			if strings.Contains(strings.ToLower(senior.Title), "vp") {
				foundVP = true
				assert.Equal(t, "senior", senior.Level)
			}
			if strings.Contains(strings.ToLower(senior.Title), "director") {
				foundDirector = true
				assert.Equal(t, "senior", senior.Level)
			}
		}
		assert.True(t, foundVP)
		assert.True(t, foundDirector)
	})

	t.Run("extract team members", func(t *testing.T) {
		content := `
			Our development team includes Alex Rodriguez, Senior Software Engineer.
			Emma Thompson is our UX Designer, creating beautiful user experiences.
			James Lee serves as our Data Analyst, providing insights to the team.
		`

		result, err := extractor.ExtractKeyPersonnel(ctx, content, "https://example.com/team")
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(result.TeamMembers), 2)

		// Check that team members are properly identified
		foundEngineer := false
		foundDesigner := false
		for _, member := range result.TeamMembers {
			if strings.Contains(strings.ToLower(member.Title), "engineer") {
				foundEngineer = true
				assert.Equal(t, "team", member.Level)
			}
			if strings.Contains(strings.ToLower(member.Title), "designer") {
				foundDesigner = true
				assert.Equal(t, "team", member.Level)
			}
		}
		assert.True(t, foundEngineer)
		assert.True(t, foundDesigner)
	})

	t.Run("extract with emails and LinkedIn", func(t *testing.T) {
		content := `
			John Smith, CEO - john.smith@company.com
			LinkedIn: https://linkedin.com/in/johnsmith
			Sarah Johnson, CTO - sarah.johnson@company.com
		`

		result, err := extractor.ExtractKeyPersonnel(ctx, content, "https://example.com/team")
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(result.Executives), 1)

		// Check that email and LinkedIn are extracted
		for _, exec := range result.Executives {
			if strings.Contains(exec.Name, "John") {
				assert.Contains(t, exec.Email, "john.smith@company.com")
				assert.Contains(t, exec.LinkedInURL, "linkedin.com")
			}
		}
	})

	t.Run("handle context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(2 * time.Nanosecond) // Ensure timeout

		content := "John Smith, CEO"
		result, err := extractor.ExtractKeyPersonnel(ctx, content, "https://example.com/team")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
		assert.Nil(t, result)
	})

	t.Run("respect confidence threshold", func(t *testing.T) {
		config := getDefaultKeyPersonnelConfig()
		config.MinConfidenceThreshold = 0.95
		extractor := NewKeyPersonnelExtractor(config, logger)

		content := "John Smith, CEO"
		result, err := extractor.ExtractKeyPersonnel(ctx, content, "https://example.com/team")
		require.NoError(t, err)

		// Only high-confidence personnel should be included
		for _, exec := range result.Executives {
			assert.GreaterOrEqual(t, exec.ConfidenceScore, 0.95)
		}
	})

	t.Run("apply data anonymization", func(t *testing.T) {
		config := getDefaultKeyPersonnelConfig()
		config.EnableDataAnonymization = true
		extractor := NewKeyPersonnelExtractor(config, logger)

		content := "John Smith, CEO - john.smith@company.com"
		result, err := extractor.ExtractKeyPersonnel(ctx, content, "https://example.com/team")
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(result.Executives), 1)

		// Check that data is anonymized
		for _, exec := range result.Executives {
			assert.True(t, exec.PrivacyCompliance.IsAnonymized)
			assert.Empty(t, exec.Email)       // Email should be removed
			assert.Empty(t, exec.LinkedInURL) // LinkedIn should be removed
		}
	})
}

func TestExtractExecutives(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewKeyPersonnelExtractor(nil, logger)
	ctx := context.Background()

	t.Run("extract CEO", func(t *testing.T) {
		content := "John Smith is our CEO and founder."
		executives, err := extractor.extractExecutives(ctx, content)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(executives), 1)
		assert.Equal(t, "executive", executives[0].Level)
		assert.Contains(t, strings.ToLower(executives[0].Title), "ceo")
		assert.Greater(t, executives[0].ConfidenceScore, 0.8)
	})

	t.Run("extract multiple executives", func(t *testing.T) {
		content := `
			John Smith, CEO
			Sarah Johnson, CTO
			Michael Brown, CFO
		`
		executives, err := extractor.extractExecutives(ctx, content)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(executives), 3)

		titles := make(map[string]bool)
		for _, exec := range executives {
			titles[strings.ToLower(exec.Title)] = true
			assert.Equal(t, "executive", exec.Level)
		}

		assert.True(t, titles["ceo"])
		assert.True(t, titles["cto"])
		assert.True(t, titles["cfo"])
	})

	t.Run("extract with department detection", func(t *testing.T) {
		content := "Sarah Johnson, CTO"
		executives, err := extractor.extractExecutives(ctx, content)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(executives), 1)
		assert.Equal(t, "Engineering", executives[0].Department)
	})
}

func TestExtractSeniorManagement(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewKeyPersonnelExtractor(nil, logger)
	ctx := context.Background()

	t.Run("extract VP", func(t *testing.T) {
		content := "David Wilson, VP of Engineering"
		seniorMgmt, err := extractor.extractSeniorManagement(ctx, content)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(seniorMgmt), 1)
		assert.Equal(t, "senior", seniorMgmt[0].Level)
		assert.Contains(t, strings.ToLower(seniorMgmt[0].Title), "vp")
	})

	t.Run("extract Director", func(t *testing.T) {
		content := "Lisa Chen, Director of Marketing"
		seniorMgmt, err := extractor.extractSeniorManagement(ctx, content)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(seniorMgmt), 1)
		assert.Equal(t, "senior", seniorMgmt[0].Level)
		assert.Contains(t, strings.ToLower(seniorMgmt[0].Title), "director")
	})
}

func TestExtractTeamMembers(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewKeyPersonnelExtractor(nil, logger)
	ctx := context.Background()

	t.Run("extract engineers", func(t *testing.T) {
		content := "Alex Rodriguez, Senior Software Engineer"
		teamMembers, err := extractor.extractTeamMembers(ctx, content)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(teamMembers), 1)
		assert.Equal(t, "team", teamMembers[0].Level)
		assert.Contains(t, strings.ToLower(teamMembers[0].Title), "engineer")
	})

	t.Run("extract designers", func(t *testing.T) {
		content := "Emma Thompson, UX Designer"
		teamMembers, err := extractor.extractTeamMembers(ctx, content)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(teamMembers), 1)
		assert.Equal(t, "team", teamMembers[0].Level)
		assert.Contains(t, strings.ToLower(teamMembers[0].Title), "designer")
	})
}

func TestKeyPersonnelUtilityFunctions(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewKeyPersonnelExtractor(nil, logger)

	t.Run("extract name near title", func(t *testing.T) {
		content := "John Smith is our CEO"
		name := extractor.extractNameNearTitle(content, "CEO")
		assert.Equal(t, "John Smith", name)
	})

	t.Run("extract email for person", func(t *testing.T) {
		content := "John Smith, CEO - john.smith@company.com"
		email := extractor.extractEmailForPerson("John Smith", content)
		assert.Equal(t, "john.smith@company.com", email)
	})

	t.Run("extract LinkedIn URL", func(t *testing.T) {
		content := "John Smith - https://linkedin.com/in/johnsmith"
		linkedIn := extractor.extractLinkedInURL("John Smith", content)
		assert.Contains(t, linkedIn, "linkedin.com")
	})

	t.Run("extract bio", func(t *testing.T) {
		content := "John Smith is our CEO. He has over 20 years of experience."
		bio := extractor.extractBioForPerson("John Smith", content)
		assert.Contains(t, bio, "He has over 20 years of experience")
	})

	t.Run("determine department", func(t *testing.T) {
		assert.Equal(t, "Engineering", extractor.determineDepartment("CTO"))
		assert.Equal(t, "Finance", extractor.determineDepartment("CFO"))
		assert.Equal(t, "Marketing", extractor.determineDepartment("CMO"))
		assert.Equal(t, "General", extractor.determineDepartment("Unknown Title"))
	})
}

func TestKeyPersonnelConfidenceCalculation(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewKeyPersonnelExtractor(nil, logger)

	t.Run("executive confidence", func(t *testing.T) {
		confidence := extractor.calculateExecutiveConfidence("John Smith", "CEO")
		assert.Greater(t, confidence, 0.9)

		confidence = extractor.calculateExecutiveConfidence("John Smith", "Chief Executive Officer")
		assert.Greater(t, confidence, 0.9)
	})

	t.Run("senior confidence", func(t *testing.T) {
		confidence := extractor.calculateSeniorConfidence("David Wilson", "VP of Engineering")
		assert.Greater(t, confidence, 0.8)

		confidence = extractor.calculateSeniorConfidence("Lisa Chen", "Director of Marketing")
		assert.Greater(t, confidence, 0.8)
	})

	t.Run("team confidence", func(t *testing.T) {
		confidence := extractor.calculateTeamConfidence("Alex Rodriguez", "Software Engineer")
		assert.Greater(t, confidence, 0.7)

		confidence = extractor.calculateTeamConfidence("Emma Thompson", "Senior UX Designer")
		assert.Greater(t, confidence, 0.7)
	})
}

func TestDataProcessing(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewKeyPersonnelExtractor(nil, logger)

	t.Run("deduplicate personnel", func(t *testing.T) {
		personnel := []ExecutiveTeamMember{
			{Name: "John Smith", Title: "CEO"},
			{Name: "John Smith", Title: "CEO"}, // Duplicate
			{Name: "Sarah Johnson", Title: "CTO"},
		}

		deduplicated := extractor.deduplicatePersonnel(personnel)
		assert.Equal(t, 2, len(deduplicated))
	})

	t.Run("filter by confidence", func(t *testing.T) {
		personnel := []ExecutiveTeamMember{
			{Name: "John Smith", Title: "CEO", ConfidenceScore: 0.9},
			{Name: "Sarah Johnson", Title: "CTO", ConfidenceScore: 0.5}, // Below threshold
		}

		config := getDefaultKeyPersonnelConfig()
		config.MinConfidenceThreshold = 0.6
		extractor.config = config

		filtered := extractor.filterByConfidence(personnel)
		assert.Equal(t, 1, len(filtered))
		assert.Equal(t, "John Smith", filtered[0].Name)
	})

	t.Run("anonymize personnel", func(t *testing.T) {
		personnel := []ExecutiveTeamMember{
			{
				Name:        "John Smith",
				Title:       "CEO",
				Email:       "john.smith@company.com",
				LinkedInURL: "https://linkedin.com/in/johnsmith",
				Bio:         "John has 20 years of experience",
			},
		}

		anonymized := extractor.anonymizePersonnel(personnel)
		assert.Equal(t, 1, len(anonymized))
		assert.Equal(t, "J. Smith", anonymized[0].Name)
		assert.Empty(t, anonymized[0].Email)
		assert.Empty(t, anonymized[0].LinkedInURL)
		assert.Equal(t, "Professional bio available", anonymized[0].Bio)
		assert.True(t, anonymized[0].PrivacyCompliance.IsAnonymized)
	})
}

func TestKeyPersonnelStatisticsCalculation(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewKeyPersonnelExtractor(nil, logger)

	t.Run("calculate personnel stats", func(t *testing.T) {
		result := &PersonnelExtractionResult{
			Executives: []ExecutiveTeamMember{
				{Name: "John Smith", Title: "CEO", ConfidenceScore: 0.9},
				{Name: "Sarah Johnson", Title: "CTO", ConfidenceScore: 0.85},
			},
			SeniorManagement: []ExecutiveTeamMember{
				{Name: "David Wilson", Title: "VP Engineering", ConfidenceScore: 0.8},
			},
			TeamMembers: []ExecutiveTeamMember{
				{Name: "Alex Rodriguez", Title: "Engineer", ConfidenceScore: 0.7},
			},
		}

		stats := extractor.calculatePersonnelStats(result)
		assert.Equal(t, 4, stats.TotalMatches)
		assert.Equal(t, 2, stats.ValidExecutives)
		assert.Equal(t, 1, stats.ValidSeniorMgmt)
		assert.Equal(t, 1, stats.ValidTeamMembers)
		assert.InDelta(t, 0.8125, stats.AverageConfidence, 0.01) // (0.9 + 0.85 + 0.8 + 0.7) / 4
	})

	t.Run("calculate overall confidence", func(t *testing.T) {
		result := &PersonnelExtractionResult{
			Executives: []ExecutiveTeamMember{
				{Name: "John Smith", Title: "CEO", ConfidenceScore: 0.9},
			},
			SeniorManagement: []ExecutiveTeamMember{
				{Name: "David Wilson", Title: "VP Engineering", ConfidenceScore: 0.8},
			},
			TeamMembers: []ExecutiveTeamMember{
				{Name: "Alex Rodriguez", Title: "Engineer", ConfidenceScore: 0.7},
			},
		}

		confidence := extractor.calculateOverallConfidence(result)
		// Weighted calculation: (0.9 * 0.5 + 0.8 * 0.3 + 0.7 * 0.2) / (0.5 + 0.3 + 0.2) = 0.83
		assert.InDelta(t, 0.83, confidence, 0.01)
	})
}

func TestKeyPersonnelConfigurationMethods(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewKeyPersonnelExtractor(nil, logger)

	t.Run("update config", func(t *testing.T) {
		newConfig := &KeyPersonnelConfig{
			EnableExecutiveExtraction: false,
			MinConfidenceThreshold:    0.8,
		}

		err := extractor.UpdateConfig(newConfig)
		assert.NoError(t, err)
		assert.False(t, extractor.config.EnableExecutiveExtraction)
		assert.Equal(t, 0.8, extractor.config.MinConfidenceThreshold)
	})

	t.Run("update config with nil", func(t *testing.T) {
		err := extractor.UpdateConfig(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("get config", func(t *testing.T) {
		config := extractor.GetConfig()
		assert.NotNil(t, config)
		assert.Equal(t, extractor.config, config)
	})
}

func TestDefaultConfiguration(t *testing.T) {
	config := getDefaultKeyPersonnelConfig()

	t.Run("default settings", func(t *testing.T) {
		assert.True(t, config.EnableExecutiveExtraction)
		assert.True(t, config.EnableTeamExtraction)
		assert.True(t, config.EnableRoleDetection)
		assert.False(t, config.EnableLinkedInIntegration)
		assert.Equal(t, 0.6, config.MinConfidenceThreshold)
		assert.True(t, config.EnableDuplicateDetection)
		assert.True(t, config.EnableContextValidation)
		assert.False(t, config.EnableDataAnonymization)
		assert.Equal(t, 50, config.MaxPersonnelCount)
	})

	t.Run("executive titles", func(t *testing.T) {
		assert.Contains(t, config.ExecutiveTitles, "CEO")
		assert.Contains(t, config.ExecutiveTitles, "CTO")
		assert.Contains(t, config.ExecutiveTitles, "CFO")
		assert.Contains(t, config.ExecutiveTitles, "Chief Executive Officer")
	})

	t.Run("senior titles", func(t *testing.T) {
		assert.Contains(t, config.SeniorTitles, "VP")
		assert.Contains(t, config.SeniorTitles, "Vice President")
		assert.Contains(t, config.SeniorTitles, "Director")
	})

	t.Run("department titles", func(t *testing.T) {
		assert.Contains(t, config.DepartmentTitles, "Engineering")
		assert.Contains(t, config.DepartmentTitles, "Finance")
		assert.Contains(t, config.DepartmentTitles, "Marketing")
	})
}
