package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
	"kyb-platform/internal/observability"
)

// DeprecatedItem represents a deprecated code item found during scanning
type DeprecatedItem struct {
	File       string    `json:"file"`
	Line       int       `json:"line"`
	Column     int       `json:"column"`
	Content    string    `json:"content"`
	Pattern    string    `json:"pattern"`
	Severity   string    `json:"severity"`
	Suggestion string    `json:"suggestion"`
	Timestamp  time.Time `json:"timestamp"`
	CanAutoFix bool      `json:"can_auto_fix"`
	FixApplied bool      `json:"fix_applied"`
}

// CleanupReport represents the output of a cleanup scan
type CleanupReport struct {
	ScanDate        time.Time              `json:"scan_date"`
	ProjectRoot     string                 `json:"project_root"`
	ScanOptions     ScanOptions            `json:"scan_options"`
	Summary         CleanupSummary         `json:"summary"`
	DeprecatedItems []DeprecatedItem       `json:"deprecated_items"`
	Recommendations []string               `json:"recommendations"`
	FixesApplied    []string               `json:"fixes_applied"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ScanOptions represents the configuration for the cleanup scan
type ScanOptions struct {
	DryRun         bool     `json:"dry_run"`
	Interactive    bool     `json:"interactive"`
	AutoFix        bool     `json:"auto_fix"`
	PatternFilter  string   `json:"pattern_filter"`
	FileFilter     string   `json:"file_filter"`
	ExcludePattern []string `json:"exclude_patterns"`
	Severity       string   `json:"severity"`
	OutputFormat   string   `json:"output_format"`
}

// CleanupSummary provides statistics about the cleanup scan
type CleanupSummary struct {
	TotalItems      int            `json:"total_items"`
	FilesAffected   int            `json:"files_affected"`
	ItemsBySeverity map[string]int `json:"items_by_severity"`
	ItemsByPattern  map[string]int `json:"items_by_pattern"`
	AutoFixable     int            `json:"auto_fixable"`
	FixesApplied    int            `json:"fixes_applied"`
	ScanDuration    time.Duration  `json:"scan_duration"`
}

// DeprecatedCodeCleaner handles the detection and cleanup of deprecated code
type DeprecatedCodeCleaner struct {
	logger               *zap.Logger
	technicalDebtMonitor *observability.TechnicalDebtMonitor
	options              ScanOptions
	patterns             map[string]*regexp.Regexp
	excludePatterns      []*regexp.Regexp
	projectRoot          string
}

// NewDeprecatedCodeCleaner creates a new deprecated code cleaner
func NewDeprecatedCodeCleaner(logger *zap.Logger, projectRoot string, options ScanOptions) *DeprecatedCodeCleaner {
	cleaner := &DeprecatedCodeCleaner{
		logger:      logger,
		options:     options,
		projectRoot: projectRoot,
		patterns:    make(map[string]*regexp.Regexp),
	}

	// Initialize technical debt monitor
	cleaner.technicalDebtMonitor = observability.NewTechnicalDebtMonitor(observability.NewLogger(logger))

	// Initialize patterns
	cleaner.initializePatterns()
	cleaner.initializeExcludePatterns()

	return cleaner
}

// initializePatterns sets up regex patterns for detecting deprecated code
func (dcc *DeprecatedCodeCleaner) initializePatterns() {
	patterns := map[string]string{
		// Comment patterns
		"deprecated_comment": `(?i)(//|#|/\*)\s*(deprecated|fixme|todo.*remove|legacy|obsolete)`,
		"todo_remove":        `(?i)(//|#|/\*)\s*todo.*remove`,
		"fixme_legacy":       `(?i)(//|#|/\*)\s*fixme.*legacy`,

		// Import patterns
		"deprecated_import":  `import\s+.*(".*deprecated.*"|".*legacy.*"|".*old.*")`,
		"problematic_import": `import\s+.*".*webanalysis\.problematic.*"`,

		// Function patterns
		"deprecated_function": `func\s+.*[Dd]eprecated.*\(`,
		"legacy_function":     `func\s+.*[Ll]egacy.*\(`,
		"old_function":        `func\s+.*[Oo]ld.*\(`,

		// Variable patterns
		"deprecated_variable": `(var|const)\s+.*[Dd]eprecated.*\s*=`,
		"legacy_variable":     `(var|const)\s+.*[Ll]egacy.*\s*=`,

		// Type patterns
		"deprecated_type": `type\s+.*[Dd]eprecated.*\s+(struct|interface)`,
		"legacy_type":     `type\s+.*[Ll]egacy.*\s+(struct|interface)`,

		// File patterns
		"test_file":       `_test\.go$`,
		"backup_file":     `\.(bak|backup|orig)$`,
		"temp_file":       `\.(tmp|temp)$|~$`,
		"deprecated_file": `[._](deprecated|legacy|old)[._]`,

		// Code smell patterns
		"hardcoded_url": `https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
		"magic_number":  `\b\d{3,}\b(?!\s*//.*explanation)`,
		"unused_import": `import\s+"[^"]+"\s*$`,
		"empty_catch":   `catch\s*\([^)]*\)\s*\{\s*\}`,

		// Build artifacts
		"generated_file": `// Code generated .* DO NOT EDIT\.`,
		"binary_file":    `\.(exe|bin|so|dll|dylib)$`,
	}

	for name, pattern := range patterns {
		if dcc.options.PatternFilter == "" || dcc.options.PatternFilter == name {
			compiled, err := regexp.Compile(pattern)
			if err != nil {
				dcc.logger.Warn("Failed to compile pattern", zap.String("pattern", name), zap.Error(err))
				continue
			}
			dcc.patterns[name] = compiled
		}
	}
}

// initializeExcludePatterns sets up patterns for files to exclude from scanning
func (dcc *DeprecatedCodeCleaner) initializeExcludePatterns() {
	excludePatterns := []string{
		`vendor/`,
		`node_modules/`,
		`\.git/`,
		`\.cleanup-backups/`,
		`reports/`,
		`logs/`,
		`tmp/`,
		`\.(log|bin|exe|so|dll|dylib)$`,
		`_test\.go$`, // Only exclude if not specifically looking for test files
	}

	// Add custom exclude patterns
	excludePatterns = append(excludePatterns, dcc.options.ExcludePattern...)

	for _, pattern := range excludePatterns {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			dcc.logger.Warn("Failed to compile exclude pattern", zap.String("pattern", pattern), zap.Error(err))
			continue
		}
		dcc.excludePatterns = append(dcc.excludePatterns, compiled)
	}
}

// shouldExcludeFile checks if a file should be excluded from scanning
func (dcc *DeprecatedCodeCleaner) shouldExcludeFile(filePath string) bool {
	relPath, err := filepath.Rel(dcc.projectRoot, filePath)
	if err != nil {
		relPath = filePath
	}

	for _, pattern := range dcc.excludePatterns {
		if pattern.MatchString(relPath) {
			return true
		}
	}
	return false
}

// scanFile scans a single file for deprecated code patterns
func (dcc *DeprecatedCodeCleaner) scanFile(filePath string) ([]DeprecatedItem, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	lines := strings.Split(string(content), "\n")
	var items []DeprecatedItem

	relPath, _ := filepath.Rel(dcc.projectRoot, filePath)

	for patternName, pattern := range dcc.patterns {
		// Check if pattern matches filename
		if strings.Contains(patternName, "file") && pattern.MatchString(relPath) {
			items = append(items, DeprecatedItem{
				File:       relPath,
				Line:       0,
				Column:     0,
				Content:    fmt.Sprintf("File matches deprecated pattern: %s", patternName),
				Pattern:    patternName,
				Severity:   dcc.getSeverityForPattern(patternName),
				Suggestion: dcc.getSuggestionForPattern(patternName, ""),
				Timestamp:  time.Now(),
				CanAutoFix: dcc.canAutoFixPattern(patternName),
			})
			continue
		}

		// Check content patterns
		for lineNum, line := range lines {
			matches := pattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 0 {
					items = append(items, DeprecatedItem{
						File:       relPath,
						Line:       lineNum + 1,
						Column:     strings.Index(line, match[0]) + 1,
						Content:    strings.TrimSpace(line),
						Pattern:    patternName,
						Severity:   dcc.getSeverityForPattern(patternName),
						Suggestion: dcc.getSuggestionForPattern(patternName, match[0]),
						Timestamp:  time.Now(),
						CanAutoFix: dcc.canAutoFixPattern(patternName),
					})
				}
			}
		}
	}

	return items, nil
}

// getSeverityForPattern returns the severity level for a pattern
func (dcc *DeprecatedCodeCleaner) getSeverityForPattern(pattern string) string {
	severityMap := map[string]string{
		"deprecated_comment":  "high",
		"deprecated_function": "high",
		"deprecated_type":     "high",
		"deprecated_import":   "high",
		"problematic_import":  "critical",
		"todo_remove":         "medium",
		"fixme_legacy":        "medium",
		"legacy_function":     "medium",
		"legacy_variable":     "medium",
		"legacy_type":         "medium",
		"backup_file":         "low",
		"temp_file":           "low",
		"hardcoded_url":       "low",
		"magic_number":        "low",
		"unused_import":       "medium",
	}

	if severity, exists := severityMap[pattern]; exists {
		return severity
	}
	return "medium"
}

// getSuggestionForPattern returns a suggestion for fixing the pattern
func (dcc *DeprecatedCodeCleaner) getSuggestionForPattern(pattern, match string) string {
	suggestions := map[string]string{
		"deprecated_comment":  "Remove or update deprecated comments",
		"deprecated_function": "Replace with current implementation or remove if unused",
		"deprecated_type":     "Update to use current types",
		"deprecated_import":   "Update import to use current package",
		"problematic_import":  "Remove problematic import and update code",
		"todo_remove":         "Review and remove TODO comment if task is complete",
		"fixme_legacy":        "Fix legacy code or remove if no longer needed",
		"backup_file":         "Remove backup file if no longer needed",
		"temp_file":           "Remove temporary file",
		"hardcoded_url":       "Move URL to configuration",
		"magic_number":        "Replace with named constant",
		"unused_import":       "Remove unused import",
	}

	if suggestion, exists := suggestions[pattern]; exists {
		return suggestion
	}
	return "Review and update deprecated code"
}

// canAutoFixPattern determines if a pattern can be automatically fixed
func (dcc *DeprecatedCodeCleaner) canAutoFixPattern(pattern string) bool {
	autoFixable := map[string]bool{
		"backup_file":        true,
		"temp_file":          true,
		"unused_import":      false, // Requires careful analysis
		"deprecated_comment": false, // Requires manual review
	}

	if canFix, exists := autoFixable[pattern]; exists {
		return canFix
	}
	return false
}

// scanProject scans the entire project for deprecated code
func (dcc *DeprecatedCodeCleaner) scanProject(ctx context.Context) (*CleanupReport, error) {
	startTime := time.Now()
	dcc.logger.Info("Starting project scan for deprecated code", zap.String("project_root", dcc.projectRoot))

	var allItems []DeprecatedItem
	filesScanned := 0
	filesWithIssues := 0

	err := filepath.WalkDir(dcc.projectRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			dcc.logger.Warn("Error walking directory", zap.String("path", path), zap.Error(err))
			return nil
		}

		if d.IsDir() {
			return nil
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip excluded files
		if dcc.shouldExcludeFile(path) {
			return nil
		}

		// Apply file filter if specified
		if dcc.options.FileFilter != "" && !strings.Contains(path, dcc.options.FileFilter) {
			return nil
		}

		filesScanned++
		items, err := dcc.scanFile(path)
		if err != nil {
			dcc.logger.Warn("Error scanning file", zap.String("file", path), zap.Error(err))
			return nil
		}

		if len(items) > 0 {
			filesWithIssues++
			allItems = append(allItems, items...)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning project: %w", err)
	}

	// Generate summary
	summary := dcc.generateSummary(allItems, time.Since(startTime))

	// Apply fixes if requested
	var fixesApplied []string
	if dcc.options.AutoFix && !dcc.options.DryRun {
		fixesApplied = dcc.applyAutoFixes(allItems)
	}

	// Generate recommendations
	recommendations := dcc.generateRecommendations(allItems)

	report := &CleanupReport{
		ScanDate:        time.Now(),
		ProjectRoot:     dcc.projectRoot,
		ScanOptions:     dcc.options,
		Summary:         summary,
		DeprecatedItems: allItems,
		Recommendations: recommendations,
		FixesApplied:    fixesApplied,
		Metadata: map[string]interface{}{
			"files_scanned":     filesScanned,
			"files_with_issues": filesWithIssues,
			"scan_duration":     time.Since(startTime).String(),
		},
	}

	dcc.logger.Info("Project scan completed",
		zap.Int("total_items", len(allItems)),
		zap.Int("files_scanned", filesScanned),
		zap.Int("files_with_issues", filesWithIssues),
		zap.Duration("duration", time.Since(startTime)),
	)

	return report, nil
}

// generateSummary creates a summary of the scan results
func (dcc *DeprecatedCodeCleaner) generateSummary(items []DeprecatedItem, duration time.Duration) CleanupSummary {
	summary := CleanupSummary{
		TotalItems:      len(items),
		ItemsBySeverity: make(map[string]int),
		ItemsByPattern:  make(map[string]int),
		ScanDuration:    duration,
	}

	filesSet := make(map[string]bool)
	autoFixableCount := 0
	fixesAppliedCount := 0

	for _, item := range items {
		filesSet[item.File] = true
		summary.ItemsBySeverity[item.Severity]++
		summary.ItemsByPattern[item.Pattern]++

		if item.CanAutoFix {
			autoFixableCount++
		}
		if item.FixApplied {
			fixesAppliedCount++
		}
	}

	summary.FilesAffected = len(filesSet)
	summary.AutoFixable = autoFixableCount
	summary.FixesApplied = fixesAppliedCount

	return summary
}

// applyAutoFixes applies automatic fixes for eligible items
func (dcc *DeprecatedCodeCleaner) applyAutoFixes(items []DeprecatedItem) []string {
	var fixesApplied []string

	for i := range items {
		item := &items[i]
		if !item.CanAutoFix {
			continue
		}

		switch item.Pattern {
		case "backup_file", "temp_file":
			filePath := filepath.Join(dcc.projectRoot, item.File)
			if err := os.Remove(filePath); err != nil {
				dcc.logger.Warn("Failed to remove file", zap.String("file", filePath), zap.Error(err))
			} else {
				item.FixApplied = true
				fixesApplied = append(fixesApplied, fmt.Sprintf("Removed %s: %s", item.Pattern, item.File))
				dcc.logger.Info("Auto-removed file", zap.String("file", filePath), zap.String("pattern", item.Pattern))
			}
		}
	}

	return fixesApplied
}

// generateRecommendations creates actionable recommendations based on findings
func (dcc *DeprecatedCodeCleaner) generateRecommendations(items []DeprecatedItem) []string {
	var recommendations []string
	patternCounts := make(map[string]int)

	for _, item := range items {
		patternCounts[item.Pattern]++
	}

	// Sort patterns by count
	type patternCount struct {
		pattern string
		count   int
	}
	var sortedPatterns []patternCount
	for pattern, count := range patternCounts {
		sortedPatterns = append(sortedPatterns, patternCount{pattern, count})
	}
	sort.Slice(sortedPatterns, func(i, j int) bool {
		return sortedPatterns[i].count > sortedPatterns[j].count
	})

	// Generate specific recommendations
	for _, pc := range sortedPatterns {
		switch pc.pattern {
		case "deprecated_comment":
			recommendations = append(recommendations, fmt.Sprintf("Review and remove %d deprecated comments", pc.count))
		case "backup_file", "temp_file":
			recommendations = append(recommendations, fmt.Sprintf("Clean up %d backup/temporary files", pc.count))
		case "deprecated_import":
			recommendations = append(recommendations, fmt.Sprintf("Update %d deprecated imports", pc.count))
		case "magic_number":
			recommendations = append(recommendations, fmt.Sprintf("Replace %d magic numbers with named constants", pc.count))
		}
	}

	// Add general recommendations
	if len(items) > 0 {
		recommendations = append(recommendations, "Run tests after applying fixes to ensure functionality")
		recommendations = append(recommendations, "Consider setting up pre-commit hooks to prevent deprecated code")
		recommendations = append(recommendations, "Schedule regular cleanup sessions")
	}

	return recommendations
}

// saveReport saves the cleanup report to a file
func (dcc *DeprecatedCodeCleaner) saveReport(report *CleanupReport, outputPath string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	dcc.logger.Info("Report saved", zap.String("path", outputPath))
	return nil
}

// integrateWithTechnicalDebtMonitor sends cleanup metrics to the technical debt monitoring system
func (dcc *DeprecatedCodeCleaner) integrateWithTechnicalDebtMonitor(report *CleanupReport) {
	if dcc.technicalDebtMonitor == nil {
		return
	}

	// Record deprecated API calls if any
	// TODO: Implement RecordDeprecatedAPICall method in TechnicalDebtMonitor
	// for _, item := range report.DeprecatedItems {
	// 	if item.Pattern == "deprecated_import" || item.Pattern == "deprecated_function" {
	// 		dcc.technicalDebtMonitor.RecordDeprecatedAPICall(item.File)
	// 	}
	// }

	dcc.logger.Info("Integrated with technical debt monitoring system",
		zap.Int("total_items", report.Summary.TotalItems),
		zap.Int("files_affected", report.Summary.FilesAffected),
	)
}

func main() {
	// Command line flags
	var (
		projectRoot    = flag.String("project", ".", "Project root directory")
		dryRun         = flag.Bool("dry-run", false, "Perform dry run without making changes")
		interactive    = flag.Bool("interactive", false, "Run in interactive mode")
		autoFix        = flag.Bool("auto-fix", false, "Automatically fix simple issues")
		patternFilter  = flag.String("pattern", "", "Filter by specific pattern")
		fileFilter     = flag.String("file", "", "Filter by specific file")
		severity       = flag.String("severity", "", "Filter by severity (critical,high,medium,low)")
		outputFormat   = flag.String("format", "json", "Output format (json,text)")
		outputFile     = flag.String("output", "", "Output file path")
		logLevel       = flag.String("log-level", "info", "Log level (debug,info,warn,error)")
		excludePattern = flag.String("exclude", "", "Additional exclude patterns (comma-separated)")
	)
	flag.Parse()

	// Initialize logger
	var logger *zap.Logger
	var err error
	switch *logLevel {
	case "debug":
		logger, err = zap.NewDevelopment()
	default:
		logger, err = zap.NewProduction()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Parse exclude patterns
	var excludePatterns []string
	if *excludePattern != "" {
		excludePatterns = strings.Split(*excludePattern, ",")
	}

	// Create scan options
	options := ScanOptions{
		DryRun:         *dryRun,
		Interactive:    *interactive,
		AutoFix:        *autoFix,
		PatternFilter:  *patternFilter,
		FileFilter:     *fileFilter,
		ExcludePattern: excludePatterns,
		Severity:       *severity,
		OutputFormat:   *outputFormat,
	}

	// Create cleaner
	cleaner := NewDeprecatedCodeCleaner(logger, *projectRoot, options)

	// Scan project
	ctx := context.Background()
	report, err := cleaner.scanProject(ctx)
	if err != nil {
		logger.Fatal("Failed to scan project", zap.Error(err))
	}

	// Integrate with technical debt monitor
	cleaner.integrateWithTechnicalDebtMonitor(report)

	// Determine output file
	if *outputFile == "" {
		timestamp := time.Now().Format("20060102-150405")
		*outputFile = fmt.Sprintf("deprecated-code-report-%s.json", timestamp)
	}

	// Save report
	if err := cleaner.saveReport(report, *outputFile); err != nil {
		logger.Fatal("Failed to save report", zap.Error(err))
	}

	// Print summary
	fmt.Printf("Cleanup scan completed successfully!\n")
	fmt.Printf("Total deprecated items found: %d\n", report.Summary.TotalItems)
	fmt.Printf("Files affected: %d\n", report.Summary.FilesAffected)
	fmt.Printf("Auto-fixable items: %d\n", report.Summary.AutoFixable)
	if options.AutoFix {
		fmt.Printf("Fixes applied: %d\n", report.Summary.FixesApplied)
	}
	fmt.Printf("Report saved to: %s\n", *outputFile)

	// Print recommendations
	if len(report.Recommendations) > 0 {
		fmt.Printf("\nRecommendations:\n")
		for i, rec := range report.Recommendations {
			fmt.Printf("%d. %s\n", i+1, rec)
		}
	}
}
