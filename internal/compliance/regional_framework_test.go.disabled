package compliance

import (
	"testing"
	"time"
)

func TestRegionalFrameworkCreation(t *testing.T) {
	tests := []struct {
		name     string
		creator  func() *RegionalFrameworkDefinition
		expected string
	}{
		{
			name:     "CCPA Framework",
			creator:  NewCCPAFramework,
			expected: FrameworkCCPA,
		},
		{
			name:     "LGPD Framework",
			creator:  NewLGPDFramework,
			expected: FrameworkLGPD,
		},
		{
			name:     "PIPEDA Framework",
			creator:  NewPIPEDAFramework,
			expected: FrameworkPIPEDA,
		},
		{
			name:     "POPIA Framework",
			creator:  NewPOPIAFramework,
			expected: FrameworkPOPIA,
		},
		{
			name:     "PDPA Framework",
			creator:  NewPDPAFramework,
			expected: FrameworkPDPA,
		},
		{
			name:     "APPI Framework",
			creator:  NewAPPIFramework,
			expected: FrameworkAPPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			framework := tt.creator()

			if framework.ID != tt.expected {
				t.Errorf("Expected framework ID %s, got %s", tt.expected, framework.ID)
			}

			if framework.Name == "" {
				t.Error("Framework name should not be empty")
			}

			if framework.Version == "" {
				t.Error("Framework version should not be empty")
			}

			if framework.Description == "" {
				t.Error("Framework description should not be empty")
			}

			if framework.Jurisdiction == "" {
				t.Error("Framework jurisdiction should not be empty")
			}

			if len(framework.GeographicScope) == 0 {
				t.Error("Framework geographic scope should not be empty")
			}

			if len(framework.IndustryScope) == 0 {
				t.Error("Framework industry scope should not be empty")
			}

			if framework.EffectiveDate.IsZero() {
				t.Error("Framework effective date should not be zero")
			}

			if len(framework.Requirements) == 0 {
				t.Error("Framework should have requirements")
			}

			if len(framework.Categories) == 0 {
				t.Error("Framework should have categories")
			}
		})
	}
}

func TestRegionalFrameworkRequirements(t *testing.T) {
	tests := []struct {
		name    string
		creator func() *RegionalFrameworkDefinition
		minReqs int
		minCats int
	}{
		{
			name:    "CCPA Requirements",
			creator: NewCCPAFramework,
			minReqs: 4,
			minCats: 4,
		},
		{
			name:    "LGPD Requirements",
			creator: NewLGPDFramework,
			minReqs: 3,
			minCats: 4,
		},
		{
			name:    "PIPEDA Requirements",
			creator: NewPIPEDAFramework,
			minReqs: 3,
			minCats: 8,
		},
		{
			name:    "POPIA Requirements",
			creator: NewPOPIAFramework,
			minReqs: 3,
			minCats: 6,
		},
		{
			name:    "PDPA Requirements",
			creator: NewPDPAFramework,
			minReqs: 3,
			minCats: 8,
		},
		{
			name:    "APPI Requirements",
			creator: NewAPPIFramework,
			minReqs: 3,
			minCats: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			framework := tt.creator()

			if len(framework.Requirements) < tt.minReqs {
				t.Errorf("Expected at least %d requirements, got %d", tt.minReqs, len(framework.Requirements))
			}

			if len(framework.Categories) < tt.minCats {
				t.Errorf("Expected at least %d categories, got %d", tt.minCats, len(framework.Categories))
			}

			// Check that all requirements have valid data
			for i, req := range framework.Requirements {
				if req.ID == "" {
					t.Errorf("Requirement %d ID should not be empty", i)
				}
				if req.RequirementID == "" {
					t.Errorf("Requirement %d RequirementID should not be empty", i)
				}
				if req.Framework == "" {
					t.Errorf("Requirement %d Framework should not be empty", i)
				}
				if req.Category == "" {
					t.Errorf("Requirement %d Category should not be empty", i)
				}
				if req.Title == "" {
					t.Errorf("Requirement %d Title should not be empty", i)
				}
				if req.Description == "" {
					t.Errorf("Requirement %d Description should not be empty", i)
				}
				if req.DetailedDescription == "" {
					t.Errorf("Requirement %d DetailedDescription should not be empty", i)
				}
				if req.EffectiveDate.IsZero() {
					t.Errorf("Requirement %d EffectiveDate should not be zero", i)
				}
			}

			// Check that all categories have valid data
			for i, cat := range framework.Categories {
				if cat.ID == "" {
					t.Errorf("Category %d ID should not be empty", i)
				}
				if cat.Name == "" {
					t.Errorf("Category %d Name should not be empty", i)
				}
				if cat.Code == "" {
					t.Errorf("Category %d Code should not be empty", i)
				}
				if cat.Description == "" {
					t.Errorf("Category %d Description should not be empty", i)
				}
				if len(cat.Requirements) == 0 {
					t.Errorf("Category %d should have requirements", i)
				}
			}
		})
	}
}

func TestRegionalFrameworkConversion(t *testing.T) {
	tests := []struct {
		name    string
		creator func() *RegionalFrameworkDefinition
	}{
		{"CCPA", NewCCPAFramework},
		{"LGPD", NewLGPDFramework},
		{"PIPEDA", NewPIPEDAFramework},
		{"POPIA", NewPOPIAFramework},
		{"PDPA", NewPDPAFramework},
		{"APPI", NewAPPIFramework},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regional := tt.creator()
			regulatory := regional.ConvertRegionalToRegulatoryFramework()

			if regulatory.ID != regional.ID {
				t.Errorf("Expected regulatory ID %s, got %s", regional.ID, regulatory.ID)
			}

			if regulatory.Name != regional.Name {
				t.Errorf("Expected regulatory name %s, got %s", regional.Name, regulatory.Name)
			}

			if regulatory.Version != regional.Version {
				t.Errorf("Expected regulatory version %s, got %s", regional.Version, regulatory.Version)
			}

			if len(regulatory.Requirements) != len(regional.Requirements) {
				t.Errorf("Expected %d regulatory requirements, got %d", len(regional.Requirements), len(regulatory.Requirements))
			}

			// Check that requirements are properly converted
			for i, req := range regulatory.Requirements {
				if req.Framework != regional.ID {
					t.Errorf("Requirement %d framework should be %s, got %s", i, regional.ID, req.Framework)
				}
				if req.Status != ComplianceStatusNotStarted {
					t.Errorf("Requirement %d status should be NotStarted, got %s", i, req.Status)
				}
			}
		})
	}
}

func TestRegionalFrameworkTimestamps(t *testing.T) {
	framework := NewCCPAFramework()

	// Check that timestamps are reasonable
	now := time.Now()

	if framework.LastUpdated.After(now.Add(time.Minute)) {
		t.Error("LastUpdated should not be in the future")
	}

	if framework.NextReviewDate.Before(now) {
		t.Error("NextReviewDate should be in the future")
	}

	if framework.EffectiveDate.After(now) {
		t.Error("EffectiveDate should be in the past for existing frameworks")
	}
}

func TestRegionalFrameworkMetadata(t *testing.T) {
	framework := NewCCPAFramework()

	if framework.Metadata == nil {
		t.Error("Framework metadata should be initialized")
	}

	// Test adding metadata
	framework.Metadata["test_key"] = "test_value"

	if framework.Metadata["test_key"] != "test_value" {
		t.Error("Metadata should store and retrieve values correctly")
	}
}

func TestRegionalFrameworkGeographicScope(t *testing.T) {
	tests := []struct {
		name     string
		creator  func() *RegionalFrameworkDefinition
		expected []string
	}{
		{
			name:     "CCPA Geographic Scope",
			creator:  NewCCPAFramework,
			expected: []string{"California", "United States"},
		},
		{
			name:     "LGPD Geographic Scope",
			creator:  NewLGPDFramework,
			expected: []string{"Brazil"},
		},
		{
			name:     "PIPEDA Geographic Scope",
			creator:  NewPIPEDAFramework,
			expected: []string{"Canada"},
		},
		{
			name:     "POPIA Geographic Scope",
			creator:  NewPOPIAFramework,
			expected: []string{"South Africa"},
		},
		{
			name:     "PDPA Geographic Scope",
			creator:  NewPDPAFramework,
			expected: []string{"Singapore"},
		},
		{
			name:     "APPI Geographic Scope",
			creator:  NewAPPIFramework,
			expected: []string{"Japan"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			framework := tt.creator()

			if len(framework.GeographicScope) != len(tt.expected) {
				t.Errorf("Expected %d geographic scopes, got %d", len(tt.expected), len(framework.GeographicScope))
			}

			for _, expected := range tt.expected {
				found := false
				for _, actual := range framework.GeographicScope {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected geographic scope %s not found", expected)
				}
			}
		})
	}
}
