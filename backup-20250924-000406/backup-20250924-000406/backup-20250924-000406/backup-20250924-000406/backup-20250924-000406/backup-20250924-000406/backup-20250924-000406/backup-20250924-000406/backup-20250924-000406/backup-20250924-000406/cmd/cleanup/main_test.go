package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewDeprecatedCodeCleaner(t *testing.T) {
	logger := zap.NewNop()
	options := ScanOptions{
		DryRun:      true,
		Interactive: false,
	}

	cleaner := NewDeprecatedCodeCleaner(logger, ".", options)

	assert.NotNil(t, cleaner)
	assert.Equal(t, logger, cleaner.logger)
	assert.Equal(t, options, cleaner.options)
	assert.NotNil(t, cleaner.patterns)
	assert.NotNil(t, cleaner.excludePatterns)
	assert.NotNil(t, cleaner.technicalDebtMonitor)
}

func TestShouldExcludeFile(t *testing.T) {
	logger := zap.NewNop()
	options := ScanOptions{}
	cleaner := NewDeprecatedCodeCleaner(logger, ".", options)

	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{
			name:     "Go file should not be excluded",
			filePath: "internal/test.go",
			expected: false,
		},
		{
			name:     "Vendor directory should be excluded",
			filePath: "vendor/some/package.go",
			expected: true,
		},
		{
			name:     "Git directory should be excluded",
			filePath: ".git/config",
			expected: true,
		},
		{
			name:     "Log file should be excluded",
			filePath: "app.log",
			expected: true,
		},
		{
			name:     "Binary file should be excluded",
			filePath: "app.bin",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleaner.shouldExcludeFile(tt.filePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestScanFile(t *testing.T) {
	logger := zap.NewNop()
	options := ScanOptions{}
	cleaner := NewDeprecatedCodeCleaner(logger, ".", options)

	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.go")
	content := `package test

// DEPRECATED: This function is deprecated
func oldFunction() {
	// TODO: Remove this function
	return
}

// Legacy code - should be refactored
func legacyFunction() {
	return
}

// Normal function
func normalFunction() {
	return
}`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Scan the file
	items, err := cleaner.scanFile(testFile)
	require.NoError(t, err)

	// Should find deprecated patterns
	assert.NotEmpty(t, items)

	// Check for specific patterns
	foundDeprecatedComment := false
	foundTodoRemove := false

	for _, item := range items {
		if item.Pattern == "deprecated_comment" {
			foundDeprecatedComment = true
		}
		if item.Pattern == "todo_remove" {
			foundTodoRemove = true
		}
	}

	assert.True(t, foundDeprecatedComment, "Should find deprecated comment")
	assert.True(t, foundTodoRemove, "Should find TODO remove comment")
}

func TestScanFileWithProblematicImport(t *testing.T) {
	logger := zap.NewNop()
	options := ScanOptions{}
	cleaner := NewDeprecatedCodeCleaner(logger, ".", options)

	// Create a temporary test file with problematic import
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.go")
	content := `package test

import "webanalysis.problematic/some/package"

func testFunction() {
	return
}`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Scan the file
	items, err := cleaner.scanFile(testFile)
	require.NoError(t, err)

	// Should find problematic import
	assert.NotEmpty(t, items)

	foundProblematicImport := false
	for _, item := range items {
		if item.Pattern == "problematic_import" {
			foundProblematicImport = true
			assert.Equal(t, "critical", item.Severity)
		}
	}

	assert.True(t, foundProblematicImport, "Should find problematic import")
}

func TestGetSeverityForPattern(t *testing.T) {
	logger := zap.NewNop()
	options := ScanOptions{}
	cleaner := NewDeprecatedCodeCleaner(logger, ".", options)

	tests := []struct {
		pattern  string
		expected string
	}{
		{"deprecated_comment", "high"},
		{"deprecated_function", "high"},
		{"problematic_import", "critical"},
		{"backup_file", "low"},
		{"unknown_pattern", "medium"},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			result := cleaner.getSeverityForPattern(tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCanAutoFixPattern(t *testing.T) {
	logger := zap.NewNop()
	options := ScanOptions{}
	cleaner := NewDeprecatedCodeCleaner(logger, ".", options)

	tests := []struct {
		pattern  string
		expected bool
	}{
		{"backup_file", true},
		{"temp_file", true},
		{"deprecated_comment", false},
		{"deprecated_function", false},
		{"unknown_pattern", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			result := cleaner.canAutoFixPattern(tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSummary(t *testing.T) {
	logger := zap.NewNop()
	options := ScanOptions{}
	cleaner := NewDeprecatedCodeCleaner(logger, ".", options)

	items := []DeprecatedItem{
		{
			File:       "file1.go",
			Pattern:    "deprecated_comment",
			Severity:   "high",
			CanAutoFix: false,
		},
		{
			File:       "file2.go",
			Pattern:    "backup_file",
			Severity:   "low",
			CanAutoFix: true,
		},
		{
			File:       "file1.go",
			Pattern:    "todo_remove",
			Severity:   "medium",
			CanAutoFix: false,
		},
	}

	duration := 5 * time.Second
	summary := cleaner.generateSummary(items, duration)

	assert.Equal(t, 3, summary.TotalItems)
	assert.Equal(t, 2, summary.FilesAffected) // file1.go and file2.go
	assert.Equal(t, 1, summary.AutoFixable)
	assert.Equal(t, 0, summary.FixesApplied)
	assert.Equal(t, duration, summary.ScanDuration)

	// Check severity counts
	assert.Equal(t, 1, summary.ItemsBySeverity["high"])
	assert.Equal(t, 1, summary.ItemsBySeverity["medium"])
	assert.Equal(t, 1, summary.ItemsBySeverity["low"])

	// Check pattern counts
	assert.Equal(t, 1, summary.ItemsByPattern["deprecated_comment"])
	assert.Equal(t, 1, summary.ItemsByPattern["backup_file"])
	assert.Equal(t, 1, summary.ItemsByPattern["todo_remove"])
}

func TestApplyAutoFixes(t *testing.T) {
	logger := zap.NewNop()
	options := ScanOptions{AutoFix: true}

	// Create temporary directory and files
	tempDir := t.TempDir()
	backupFile := filepath.Join(tempDir, "test.bak")
	tempFile := filepath.Join(tempDir, "test.tmp")

	// Create test files
	err := os.WriteFile(backupFile, []byte("backup content"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(tempFile, []byte("temp content"), 0644)
	require.NoError(t, err)

	cleaner := NewDeprecatedCodeCleaner(logger, tempDir, options)

	items := []DeprecatedItem{
		{
			File:       filepath.Base(backupFile),
			Pattern:    "backup_file",
			CanAutoFix: true,
		},
		{
			File:       filepath.Base(tempFile),
			Pattern:    "temp_file",
			CanAutoFix: true,
		},
		{
			File:       "some_file.go",
			Pattern:    "deprecated_comment",
			CanAutoFix: false,
		},
	}

	fixesApplied := cleaner.applyAutoFixes(items)

	// Check that fixes were applied
	assert.Len(t, fixesApplied, 2)
	assert.Contains(t, fixesApplied[0], "backup_file")
	assert.Contains(t, fixesApplied[1], "temp_file")

	// Check that files were actually removed
	_, err = os.Stat(backupFile)
	assert.True(t, os.IsNotExist(err), "Backup file should be removed")

	_, err = os.Stat(tempFile)
	assert.True(t, os.IsNotExist(err), "Temp file should be removed")

	// Check that FixApplied flag was set
	assert.True(t, items[0].FixApplied)
	assert.True(t, items[1].FixApplied)
	assert.False(t, items[2].FixApplied)
}

func TestGenerateRecommendations(t *testing.T) {
	logger := zap.NewNop()
	options := ScanOptions{}
	cleaner := NewDeprecatedCodeCleaner(logger, ".", options)

	items := []DeprecatedItem{
		{Pattern: "deprecated_comment"},
		{Pattern: "deprecated_comment"},
		{Pattern: "backup_file"},
		{Pattern: "magic_number"},
		{Pattern: "deprecated_import"},
	}

	recommendations := cleaner.generateRecommendations(items)

	assert.NotEmpty(t, recommendations)

	// Should include specific recommendations based on patterns
	hasDeprecatedCommentRec := false
	hasGeneralRec := false

	for _, rec := range recommendations {
		if strings.Contains(rec, "deprecated comments") {
			hasDeprecatedCommentRec = true
		}
		if strings.Contains(rec, "Run tests after applying fixes") {
			hasGeneralRec = true
		}
	}

	assert.True(t, hasDeprecatedCommentRec, "Should have deprecated comment recommendation")
	assert.True(t, hasGeneralRec, "Should have general recommendation")
}

func TestScanProject(t *testing.T) {
	logger := zap.NewNop()
	options := ScanOptions{
		DryRun: true,
	}

	// Create temporary project structure
	tempDir := t.TempDir()

	// Create test files
	goFile := filepath.Join(tempDir, "main.go")
	goContent := `package main

// DEPRECATED: Old main function
func main() {
	// TODO: Remove this
	return
}`
	err := os.WriteFile(goFile, []byte(goContent), 0644)
	require.NoError(t, err)

	backupFile := filepath.Join(tempDir, "test.bak")
	err = os.WriteFile(backupFile, []byte("backup"), 0644)
	require.NoError(t, err)

	// Create exclude directory
	vendorDir := filepath.Join(tempDir, "vendor")
	err = os.MkdirAll(vendorDir, 0755)
	require.NoError(t, err)
	vendorFile := filepath.Join(vendorDir, "ignored.go")
	err = os.WriteFile(vendorFile, []byte("// DEPRECATED: should be ignored"), 0644)
	require.NoError(t, err)

	cleaner := NewDeprecatedCodeCleaner(logger, tempDir, options)

	ctx := context.Background()
	report, err := cleaner.scanProject(ctx)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Check report structure
	assert.NotZero(t, report.ScanDate)
	assert.Equal(t, tempDir, report.ProjectRoot)
	assert.Equal(t, options, report.ScanOptions)
	assert.NotZero(t, report.Summary.TotalItems)
	assert.NotEmpty(t, report.DeprecatedItems)
	assert.NotEmpty(t, report.Recommendations)

	// Should find items in main.go and backup file, but not in vendor
	foundMainGo := false
	foundBackup := false
	foundVendor := false

	for _, item := range report.DeprecatedItems {
		if strings.Contains(item.File, "main.go") {
			foundMainGo = true
		}
		if strings.Contains(item.File, ".bak") {
			foundBackup = true
		}
		if strings.Contains(item.File, "vendor") {
			foundVendor = true
		}
	}

	assert.True(t, foundMainGo, "Should find deprecated items in main.go")
	assert.True(t, foundBackup, "Should find backup file")
	assert.False(t, foundVendor, "Should not find items in vendor directory")
}

func TestScanProjectWithPatternFilter(t *testing.T) {
	logger := zap.NewNop()
	options := ScanOptions{
		DryRun:        true,
		PatternFilter: "deprecated_comment",
	}

	// Create temporary project structure
	tempDir := t.TempDir()

	// Create test file with multiple patterns
	goFile := filepath.Join(tempDir, "main.go")
	goContent := `package main

// DEPRECATED: Old function
func oldFunc() {
	return
}

func legacyFunc() {
	return
}`
	err := os.WriteFile(goFile, []byte(goContent), 0644)
	require.NoError(t, err)

	cleaner := NewDeprecatedCodeCleaner(logger, tempDir, options)

	ctx := context.Background()
	report, err := cleaner.scanProject(ctx)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Should only find deprecated_comment pattern
	foundDeprecatedComment := false
	foundLegacyFunction := false

	for _, item := range report.DeprecatedItems {
		if item.Pattern == "deprecated_comment" {
			foundDeprecatedComment = true
		}
		if item.Pattern == "legacy_function" {
			foundLegacyFunction = true
		}
	}

	assert.True(t, foundDeprecatedComment, "Should find deprecated comment")
	assert.False(t, foundLegacyFunction, "Should not find legacy function (filtered out)")
}

func TestScanProjectWithFileFilter(t *testing.T) {
	logger := zap.NewNop()
	options := ScanOptions{
		DryRun:     true,
		FileFilter: "main.go",
	}

	// Create temporary project structure
	tempDir := t.TempDir()

	// Create multiple test files
	mainFile := filepath.Join(tempDir, "main.go")
	mainContent := `// DEPRECATED: Main function`
	err := os.WriteFile(mainFile, []byte(mainContent), 0644)
	require.NoError(t, err)

	utilFile := filepath.Join(tempDir, "util.go")
	utilContent := `// DEPRECATED: Util function`
	err = os.WriteFile(utilFile, []byte(utilContent), 0644)
	require.NoError(t, err)

	cleaner := NewDeprecatedCodeCleaner(logger, tempDir, options)

	ctx := context.Background()
	report, err := cleaner.scanProject(ctx)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Should only find items in main.go
	foundMain := false
	foundUtil := false

	for _, item := range report.DeprecatedItems {
		if strings.Contains(item.File, "main.go") {
			foundMain = true
		}
		if strings.Contains(item.File, "util.go") {
			foundUtil = true
		}
	}

	assert.True(t, foundMain, "Should find items in main.go")
	assert.False(t, foundUtil, "Should not find items in util.go (filtered out)")
}

func TestDeprecatedItemJSONSerialization(t *testing.T) {
	item := DeprecatedItem{
		File:       "test.go",
		Line:       10,
		Column:     5,
		Content:    "// DEPRECATED: old code",
		Pattern:    "deprecated_comment",
		Severity:   "high",
		Suggestion: "Remove deprecated comment",
		Timestamp:  time.Now(),
		CanAutoFix: false,
		FixApplied: false,
	}

	// Test JSON marshaling
	data, err := json.Marshal(item)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Test JSON unmarshaling
	var unmarshaled DeprecatedItem
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, item.File, unmarshaled.File)
	assert.Equal(t, item.Pattern, unmarshaled.Pattern)
	assert.Equal(t, item.Severity, unmarshaled.Severity)
}
