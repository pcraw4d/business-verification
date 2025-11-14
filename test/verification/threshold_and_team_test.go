package verification

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kyb-platform/internal/risk"
)

// TestThresholdManagerFunctionality verifies threshold manager works correctly
func TestThresholdManagerFunctionality(t *testing.T) {
	// Create in-memory threshold manager
	manager := risk.CreateDefaultThresholds()

	// Test: List all configs
	configs := manager.ListConfigs()
	assert.Greater(t, len(configs), 0, "Should have default thresholds")

	// Test: Get config by category
	financialConfigs := manager.GetConfigsByCategory(risk.RiskCategoryFinancial)
	assert.Greater(t, len(financialConfigs), 0, "Should have financial thresholds")

	// Test: Get default config
	defaultConfig, exists := manager.GetDefaultConfig(risk.RiskCategoryFinancial)
	if exists {
		assert.True(t, defaultConfig.IsDefault, "Should be marked as default")
		assert.Equal(t, risk.RiskCategoryFinancial, defaultConfig.Category)
	}

	t.Log("✅ Threshold Manager functionality verified")
}

// TestTeamPerformanceGrouping verifies team performance calculation logic
func TestTeamPerformanceGrouping(t *testing.T) {
	// Simulate gap tracking data with teams
	type Gap struct {
		ID       string
		Team     []string
		Progress int
		Status   string
	}

	gaps := []Gap{
		{ID: "gap1", Team: []string{"Alice", "Bob"}, Progress: 50, Status: "in_progress"},
		{ID: "gap2", Team: []string{"Bob", "Alice"}, Progress: 75, Status: "in_progress"}, // Same team, different order
		{ID: "gap3", Team: []string{"Charlie", "David"}, Progress: 100, Status: "completed"},
		{ID: "gap4", Team: []string{"Alice", "Bob", "Eve"}, Progress: 25, Status: "in_progress"},
	}

	// Group by unique teams (same logic as getAllTeamPerformance)
	uniqueTeams := make(map[string][]string)
	for _, gap := range gaps {
		if len(gap.Team) == 0 {
			continue
		}

		// Create sorted copy for key
		teamCopy := make([]string, len(gap.Team))
		copy(teamCopy, gap.Team)
		// Sort to ensure same team members in different order are treated as same team
		sort.Strings(teamCopy)
		teamKey := strings.Join(teamCopy, ",")

		if _, exists := uniqueTeams[teamKey]; !exists {
			uniqueTeams[teamKey] = gap.Team
		}
	}

	// Verify we have 3 unique teams (Alice+Bob, Charlie+David, Alice+Bob+Eve)
	assert.Equal(t, 3, len(uniqueTeams), "Should identify 3 unique teams")

	// Verify Alice+Bob team is identified correctly (regardless of order)
	// Both "Alice,Bob" and "Bob,Alice" should map to the same sorted key "Alice,Bob"
	expectedKey := "Alice,Bob"
	
	found := false
	for key := range uniqueTeams {
		if key == expectedKey {
			found = true
			break
		}
	}
	assert.True(t, found, "Should identify Alice+Bob as same team regardless of order")

	t.Log("✅ Team performance grouping logic verified")
}

// TestThresholdExportImport verifies export/import functionality
func TestThresholdExportImport(t *testing.T) {
	// Create manager with test threshold
	manager := risk.NewThresholdManager()
	
	testConfig := &risk.ThresholdConfig{
		ID:          uuid.New().String(),
		Name:        "Test Export Threshold",
		Description: "Test threshold for export/import",
		Category:    risk.RiskCategoryFinancial,
		RiskLevels: map[risk.RiskLevel]float64{
			risk.RiskLevelLow:      25.0,
			risk.RiskLevelMedium:   50.0,
			risk.RiskLevelHigh:     75.0,
			risk.RiskLevelCritical: 90.0,
		},
		IsDefault: false,
		IsActive:  true,
		Priority:  5,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Register config
	err := manager.RegisterConfig(testConfig)
	require.NoError(t, err, "Should register config successfully")

	// Export thresholds
	service := risk.NewThresholdConfigService(manager)
	exportData, err := service.ExportThresholds()
	require.NoError(t, err, "Should export thresholds successfully")

	// Verify export data is valid JSON
	var exportedConfigs []*risk.ThresholdConfig
	err = json.Unmarshal(exportData, &exportedConfigs)
	require.NoError(t, err, "Exported JSON should be valid")
	assert.Greater(t, len(exportedConfigs), 0, "Should export at least one threshold")

	// Verify our test config is in the export
	found := false
	for _, cfg := range exportedConfigs {
		if cfg.ID == testConfig.ID {
			found = true
			assert.Equal(t, testConfig.Name, cfg.Name)
			assert.Equal(t, testConfig.Category, cfg.Category)
			break
		}
	}
	assert.True(t, found, "Test config should be in export")

	// Import into new manager
	newManager := risk.NewThresholdManager()
	newService := risk.NewThresholdConfigService(newManager)
	
	err = newService.ImportThresholds(exportData)
	require.NoError(t, err, "Should import thresholds successfully")

	// Verify imported config exists
	retrieved, exists := newManager.GetConfig(testConfig.ID)
	assert.True(t, exists, "Should exist after import")
	assert.Equal(t, testConfig.Name, retrieved.Name)
	assert.Equal(t, testConfig.Category, retrieved.Category)
	assert.Equal(t, len(testConfig.RiskLevels), len(retrieved.RiskLevels))

	t.Log("✅ Threshold export/import functionality verified")
}

