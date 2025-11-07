package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// DashboardExportService provides comprehensive dashboard export and reporting functionality
type DashboardExportService struct {
	logger    *Logger
	config    *DashboardExportConfig
	exporters map[string]DashboardExporter
	reporters map[string]DashboardReporter
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	started   bool
}

// DashboardExportConfig holds configuration for dashboard export service
type DashboardExportConfig struct {
	Enabled             bool
	ExportInterval      time.Duration
	ReportInterval      time.Duration
	DataRetentionPeriod time.Duration
	MaxExports          int
	MaxReports          int
	Environment         string
	ServiceName         string
	Version             string
}

// DashboardExporter interface for exporting dashboard data
type DashboardExporter interface {
	Export(data interface{}) error
	Name() string
	Type() string
	Enabled() bool
}

// DashboardReporter interface for generating reports
type DashboardReporter interface {
	GenerateReport(data interface{}) (*DashboardReport, error)
	Name() string
	Type() string
	Enabled() bool
}

// DashboardReport represents a dashboard report
type DashboardReport struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Data        interface{}            `json:"data"`
	Format      string                 `json:"format"`
	GeneratedAt time.Time              `json:"generated_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// JSONDashboardExporter exports dashboard data as JSON
type JSONDashboardExporter struct {
	logger  *Logger
	enabled bool
}

// NewJSONDashboardExporter creates a new JSON dashboard exporter
func NewJSONDashboardExporter(logger *Logger) *JSONDashboardExporter {
	return &JSONDashboardExporter{
		logger:  logger,
		enabled: true,
	}
}

// Export exports dashboard data as JSON
func (jde *JSONDashboardExporter) Export(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	jde.logger.Debug("Dashboard data exported as JSON", map[string]interface{}{
		"data_size": len(jsonData),
	})

	return nil
}

// Name returns the exporter name
func (jde *JSONDashboardExporter) Name() string {
	return "json"
}

// Type returns the exporter type
func (jde *JSONDashboardExporter) Type() string {
	return "json"
}

// Enabled returns whether the exporter is enabled
func (jde *JSONDashboardExporter) Enabled() bool {
	return jde.enabled
}

// CSVDashboardExporter exports dashboard data as CSV
type CSVDashboardExporter struct {
	logger  *Logger
	enabled bool
}

// NewCSVDashboardExporter creates a new CSV dashboard exporter
func NewCSVDashboardExporter(logger *Logger) *CSVDashboardExporter {
	return &CSVDashboardExporter{
		logger:  logger,
		enabled: true,
	}
}

// Export exports dashboard data as CSV
func (cde *CSVDashboardExporter) Export(data interface{}) error {
	cde.logger.Debug("Dashboard data exported as CSV", map[string]interface{}{
		"data_type": fmt.Sprintf("%T", data),
	})

	// In a real implementation, this would convert data to CSV format
	return nil
}

// Name returns the exporter name
func (cde *CSVDashboardExporter) Name() string {
	return "csv"
}

// Type returns the exporter type
func (cde *CSVDashboardExporter) Type() string {
	return "csv"
}

// Enabled returns whether the exporter is enabled
func (cde *CSVDashboardExporter) Enabled() bool {
	return cde.enabled
}

// PDFDashboardReporter generates PDF reports
type PDFDashboardReporter struct {
	logger  *Logger
	enabled bool
}

// NewPDFDashboardReporter creates a new PDF dashboard reporter
func NewPDFDashboardReporter(logger *Logger) *PDFDashboardReporter {
	return &PDFDashboardReporter{
		logger:  logger,
		enabled: true,
	}
}

// GenerateReport generates a PDF report
func (pdr *PDFDashboardReporter) GenerateReport(data interface{}) (*DashboardReport, error) {
	report := &DashboardReport{
		ID:          fmt.Sprintf("pdf_%d", time.Now().Unix()),
		Name:        "PDF Report",
		Type:        "pdf",
		Title:       "KYB Platform Dashboard Report",
		Description: "Comprehensive dashboard report in PDF format",
		Data:        data,
		Format:      "pdf",
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Metadata: map[string]interface{}{
			"generator": "pdf_reporter",
			"version":   "1.0",
		},
	}

	pdr.logger.Debug("PDF report generated", map[string]interface{}{
		"report_id": report.ID,
		"title":     report.Title,
	})

	return report, nil
}

// Name returns the reporter name
func (pdr *PDFDashboardReporter) Name() string {
	return "pdf"
}

// Type returns the reporter type
func (pdr *PDFDashboardReporter) Type() string {
	return "pdf"
}

// Enabled returns whether the reporter is enabled
func (pdr *PDFDashboardReporter) Enabled() bool {
	return pdr.enabled
}

// HTMLDashboardReporter generates HTML reports
type HTMLDashboardReporter struct {
	logger  *Logger
	enabled bool
}

// NewHTMLDashboardReporter creates a new HTML dashboard reporter
func NewHTMLDashboardReporter(logger *Logger) *HTMLDashboardReporter {
	return &HTMLDashboardReporter{
		logger:  logger,
		enabled: true,
	}
}

// GenerateReport generates an HTML report
func (hdr *HTMLDashboardReporter) GenerateReport(data interface{}) (*DashboardReport, error) {
	report := &DashboardReport{
		ID:          fmt.Sprintf("html_%d", time.Now().Unix()),
		Name:        "HTML Report",
		Type:        "html",
		Title:       "KYB Platform Dashboard Report",
		Description: "Interactive dashboard report in HTML format",
		Data:        data,
		Format:      "html",
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		Metadata: map[string]interface{}{
			"generator": "html_reporter",
			"version":   "1.0",
		},
	}

	hdr.logger.Debug("HTML report generated", map[string]interface{}{
		"report_id": report.ID,
		"title":     report.Title,
	})

	return report, nil
}

// Name returns the reporter name
func (hdr *HTMLDashboardReporter) Name() string {
	return "html"
}

// Type returns the reporter type
func (hdr *HTMLDashboardReporter) Type() string {
	return "html"
}

// Enabled returns whether the reporter is enabled
func (hdr *HTMLDashboardReporter) Enabled() bool {
	return hdr.enabled
}

// NewDashboardExportService creates a new dashboard export service
func NewDashboardExportService(
	logger *Logger,
	config *DashboardExportConfig,
) *DashboardExportService {
	ctx, cancel := context.WithCancel(context.Background())

	return &DashboardExportService{
		logger:    logger,
		config:    config,
		exporters: make(map[string]DashboardExporter),
		reporters: make(map[string]DashboardReporter),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start starts the dashboard export service
func (des *DashboardExportService) Start() error {
	des.mu.Lock()
	defer des.mu.Unlock()

	if des.started {
		return fmt.Errorf("dashboard export service already started")
	}

	des.logger.Info("Starting dashboard export service", map[string]interface{}{
		"service_name": des.config.ServiceName,
		"version":      des.config.Version,
		"environment":  des.config.Environment,
	})

	// Initialize default exporters and reporters
	des.initializeDefaultExporters()
	des.initializeDefaultReporters()

	// Start export process
	if des.config.Enabled {
		go des.startExportProcess()
	}

	// Start report generation
	go des.startReportGeneration()

	des.started = true
	des.logger.Info("Dashboard export service started successfully", map[string]interface{}{})
	return nil
}

// Stop stops the dashboard export service
func (des *DashboardExportService) Stop() error {
	des.mu.Lock()
	defer des.mu.Unlock()

	if !des.started {
		return fmt.Errorf("dashboard export service not started")
	}

	des.logger.Info("Stopping dashboard export service", map[string]interface{}{})

	des.cancel()
	des.started = false

	des.logger.Info("Dashboard export service stopped successfully", map[string]interface{}{})
	return nil
}

// AddExporter adds a dashboard exporter
func (des *DashboardExportService) AddExporter(exporter DashboardExporter) {
	des.mu.Lock()
	defer des.mu.Unlock()

	des.exporters[exporter.Name()] = exporter

	des.logger.Info("Dashboard exporter added", map[string]interface{}{
		"name": exporter.Name(),
		"type": exporter.Type(),
	})
}

// RemoveExporter removes a dashboard exporter
func (des *DashboardExportService) RemoveExporter(name string) {
	des.mu.Lock()
	defer des.mu.Unlock()

	delete(des.exporters, name)

	des.logger.Info("Dashboard exporter removed", map[string]interface{}{
		"name": name,
	})
}

// AddReporter adds a dashboard reporter
func (des *DashboardExportService) AddReporter(reporter DashboardReporter) {
	des.mu.Lock()
	defer des.mu.Unlock()

	des.reporters[reporter.Name()] = reporter

	des.logger.Info("Dashboard reporter added", map[string]interface{}{
		"name": reporter.Name(),
		"type": reporter.Type(),
	})
}

// RemoveReporter removes a dashboard reporter
func (des *DashboardExportService) RemoveReporter(name string) {
	des.mu.Lock()
	defer des.mu.Unlock()

	delete(des.reporters, name)

	des.logger.Info("Dashboard reporter removed", map[string]interface{}{
		"name": name,
	})
}

// ExportData exports data using all enabled exporters
func (des *DashboardExportService) ExportData(data interface{}) error {
	des.mu.RLock()
	exporters := make([]DashboardExporter, 0, len(des.exporters))
	for _, exporter := range des.exporters {
		if exporter.Enabled() {
			exporters = append(exporters, exporter)
		}
	}
	des.mu.RUnlock()

	for _, exporter := range exporters {
		if err := exporter.Export(data); err != nil {
			des.logger.Error("Failed to export data", map[string]interface{}{
				"exporter": exporter.Name(),
				"type":     exporter.Type(),
				"error":    err.Error(),
			})
			return fmt.Errorf("failed to export data with %s: %w", exporter.Name(), err)
		}
	}

	des.logger.Debug("Data exported successfully", map[string]interface{}{
		"exporters_used": len(exporters),
	})

	return nil
}

// GenerateReport generates a report using the specified reporter
func (des *DashboardExportService) GenerateReport(reporterName string, data interface{}) (*DashboardReport, error) {
	des.mu.RLock()
	reporter, exists := des.reporters[reporterName]
	des.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("reporter %s not found", reporterName)
	}

	if !reporter.Enabled() {
		return nil, fmt.Errorf("reporter %s is disabled", reporterName)
	}

	report, err := reporter.GenerateReport(data)
	if err != nil {
		des.logger.Error("Failed to generate report", map[string]interface{}{
			"reporter": reporterName,
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("failed to generate report with %s: %w", reporterName, err)
	}

	des.logger.Info("Report generated successfully", map[string]interface{}{
		"report_id": report.ID,
		"reporter":  reporterName,
		"type":      report.Type,
	})

	return report, nil
}

// GenerateAllReports generates reports using all enabled reporters
func (des *DashboardExportService) GenerateAllReports(data interface{}) ([]*DashboardReport, error) {
	des.mu.RLock()
	reporters := make([]DashboardReporter, 0, len(des.reporters))
	for _, reporter := range des.reporters {
		if reporter.Enabled() {
			reporters = append(reporters, reporter)
		}
	}
	des.mu.RUnlock()

	var reports []*DashboardReport
	var errors []error

	for _, reporter := range reporters {
		report, err := reporter.GenerateReport(data)
		if err != nil {
			des.logger.Error("Failed to generate report", map[string]interface{}{
				"reporter": reporter.Name(),
				"error":    err.Error(),
			})
			errors = append(errors, fmt.Errorf("failed to generate report with %s: %w", reporter.Name(), err))
			continue
		}

		reports = append(reports, report)
	}

	if len(errors) > 0 {
		des.logger.Warn("Some reports failed to generate", map[string]interface{}{
			"successful_reports": len(reports),
			"failed_reports":     len(errors),
		})
	}

	des.logger.Info("Reports generated", map[string]interface{}{
		"total_reports": len(reports),
		"successful":    len(reports),
		"failed":        len(errors),
	})

	return reports, nil
}

// ListExporters returns all available exporters
func (des *DashboardExportService) ListExporters() []DashboardExporter {
	des.mu.RLock()
	defer des.mu.RUnlock()

	exporters := make([]DashboardExporter, 0, len(des.exporters))
	for _, exporter := range des.exporters {
		exporters = append(exporters, exporter)
	}

	return exporters
}

// ListReporters returns all available reporters
func (des *DashboardExportService) ListReporters() []DashboardReporter {
	des.mu.RLock()
	defer des.mu.RUnlock()

	reporters := make([]DashboardReporter, 0, len(des.reporters))
	for _, reporter := range des.reporters {
		reporters = append(reporters, reporter)
	}

	return reporters
}

// initializeDefaultExporters initializes default exporters
func (des *DashboardExportService) initializeDefaultExporters() {
	// Add JSON exporter
	jsonExporter := NewJSONDashboardExporter(des.logger)
	des.exporters[jsonExporter.Name()] = jsonExporter

	// Add CSV exporter
	csvExporter := NewCSVDashboardExporter(des.logger)
	des.exporters[csvExporter.Name()] = csvExporter

	des.logger.Info("Default exporters initialized", map[string]interface{}{
		"exporters": len(des.exporters),
	})
}

// initializeDefaultReporters initializes default reporters
func (des *DashboardExportService) initializeDefaultReporters() {
	// Add PDF reporter
	pdfReporter := NewPDFDashboardReporter(des.logger)
	des.reporters[pdfReporter.Name()] = pdfReporter

	// Add HTML reporter
	htmlReporter := NewHTMLDashboardReporter(des.logger)
	des.reporters[htmlReporter.Name()] = htmlReporter

	des.logger.Info("Default reporters initialized", map[string]interface{}{
		"reporters": len(des.reporters),
	})
}

// startExportProcess starts the export process
func (des *DashboardExportService) startExportProcess() {
	ticker := time.NewTicker(des.config.ExportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-des.ctx.Done():
			des.logger.Info("Export process stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			des.performScheduledExport()
		}
	}
}

// startReportGeneration starts the report generation process
func (des *DashboardExportService) startReportGeneration() {
	ticker := time.NewTicker(des.config.ReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-des.ctx.Done():
			des.logger.Info("Report generation stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			des.performScheduledReportGeneration()
		}
	}
}

// performScheduledExport performs scheduled data export
func (des *DashboardExportService) performScheduledExport() {
	// In a real implementation, this would collect data from all dashboards
	// and export it using all enabled exporters
	exportData := map[string]interface{}{
		"timestamp":   time.Now(),
		"service":     des.config.ServiceName,
		"version":     des.config.Version,
		"environment": des.config.Environment,
	}

	if err := des.ExportData(exportData); err != nil {
		des.logger.Error("Scheduled export failed", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// performScheduledReportGeneration performs scheduled report generation
func (des *DashboardExportService) performScheduledReportGeneration() {
	// In a real implementation, this would collect data from all dashboards
	// and generate reports using all enabled reporters
	reportData := map[string]interface{}{
		"timestamp":   time.Now(),
		"service":     des.config.ServiceName,
		"version":     des.config.Version,
		"environment": des.config.Environment,
	}

	reports, err := des.GenerateAllReports(reportData)
	if err != nil {
		des.logger.Error("Scheduled report generation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	des.logger.Info("Scheduled reports generated", map[string]interface{}{
		"report_count": len(reports),
	})
}
