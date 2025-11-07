/**
 * Shared Export Service
 * Provides unified export functionality for all pages
 * 
 * Supports multiple formats (PDF, Excel, CSV, JSON) and templates
 */

function getEventBusInstance() {
    if (typeof getEventBus !== 'undefined') {
        return getEventBus();
    }
    return {
        emit: () => {},
        on: () => () => {},
        off: () => {},
        once: () => {}
    };
}

class SharedExportService {
    constructor(config = {}) {
        this.eventBus = config.eventBus || getEventBusInstance();
        this.exportQueue = [];
        this.isExporting = false;
        this.templates = new Map();
        this.initializeTemplates();
    }
    
    /**
     * Initialize export templates
     */
    initializeTemplates() {
        // Default template
        this.templates.set('default', {
            name: 'Default Export',
            sections: ['summary', 'data']
        });
        
        // Risk report template
        this.templates.set('riskReport', {
            name: 'Risk Assessment Report',
            sections: ['executive-summary', 'risk-overview', 'risk-factors', 'recommendations']
        });
        
        // Compliance report template
        this.templates.set('complianceReport', {
            name: 'Compliance Report',
            sections: ['compliance-status', 'regulatory-requirements', 'gaps', 'recommendations']
        });
        
        // Quick summary template (for Risk Indicators tab)
        this.templates.set('quickSummary', {
            name: 'Quick Risk Summary',
            sections: ['summary', 'alerts', 'recommendations']
        });
    }
    
    /**
     * Export data
     * @param {Object} data - Data to export
     * @param {Object} options - Export options
     * @returns {Promise<void>}
     */
    async exportData(data, options = {}) {
        const {
            format = 'pdf',
            template = 'default',
            filename = null,
            includeCharts = false
        } = options;
        
        if (this.isExporting) {
            this.exportQueue.push({ data, options });
            return;
        }
        
        this.isExporting = true;
        
        try {
            this.eventBus.emit('export-started', { format, template });
            
            let result;
            switch (format.toLowerCase()) {
                case 'pdf':
                    result = await this.exportToPDF(data, template, { filename, includeCharts });
                    break;
                case 'excel':
                case 'xlsx':
                    result = await this.exportToExcel(data, template, { filename });
                    break;
                case 'csv':
                    result = await this.exportToCSV(data, template, { filename });
                    break;
                case 'json':
                    result = await this.exportToJSON(data, template, { filename });
                    break;
                default:
                    throw new Error(`Unsupported export format: ${format}`);
            }
            
            this.eventBus.emit('export-completed', { format, template, filename: result.filename });
            return result;
            
        } catch (error) {
            console.error('Export error:', error);
            this.eventBus.emit('export-failed', { format, template, error: error.message });
            throw error;
        } finally {
            this.isExporting = false;
            this.processExportQueue();
        }
    }
    
    /**
     * Export to PDF
     * @param {Object} data - Data to export
     * @param {string} template - Template name
     * @param {Object} options - Export options
     * @returns {Promise<Object>} Export result
     */
    async exportToPDF(data, template, options = {}) {
        if (typeof window === 'undefined' || !window.jspdf) {
            throw new Error('jsPDF library not loaded. Please include jsPDF before exporting to PDF.');
        }
        
        const { jsPDF } = window.jspdf;
        const doc = new jsPDF();
        const templateConfig = this.templates.get(template) || this.templates.get('default');
        
        // Set document properties
        doc.setProperties({
            title: templateConfig.name,
            subject: 'KYB Platform Export',
            author: 'KYB Platform',
            creator: 'KYB Platform'
        });
        
        // Add header
        this.addPDFHeader(doc, data, templateConfig);
        
        // Add sections
        let yPosition = 50;
        for (const section of templateConfig.sections) {
            yPosition = await this.addPDFSection(doc, section, data, yPosition);
            
            if (yPosition > 250) {
                doc.addPage();
                yPosition = 30;
            }
        }
        
        // Add charts if requested
        if (options.includeCharts && data.charts) {
            yPosition = await this.addPDFCharts(doc, data.charts, yPosition);
        }
        
        // Add footer
        this.addPDFFooter(doc);
        
        // Generate filename
        const filename = options.filename || this.generateFilename(template, 'pdf');
        
        // Save PDF
        doc.save(filename);
        
        return { filename, format: 'pdf' };
    }
    
    /**
     * Export to Excel
     * @param {Object} data - Data to export
     * @param {string} template - Template name
     * @param {Object} options - Export options
     * @returns {Promise<Object>} Export result
     */
    async exportToExcel(data, template, options = {}) {
        // Check if XLSX library is available
        if (typeof XLSX === 'undefined') {
            throw new Error('XLSX library not loaded. Please include SheetJS before exporting to Excel.');
        }
        
        const workbook = XLSX.utils.book_new();
        
        // Create worksheets based on template
        const templateConfig = this.templates.get(template) || this.templates.get('default');
        
        // Main data sheet
        const mainSheet = this.prepareExcelData(data, templateConfig);
        const worksheet = XLSX.utils.aoa_to_sheet(mainSheet);
        XLSX.utils.book_append_sheet(workbook, worksheet, 'Data');
        
        // Generate filename
        const filename = options.filename || this.generateFilename(template, 'xlsx');
        
        // Save Excel file
        XLSX.writeFile(workbook, filename);
        
        return { filename, format: 'excel' };
    }
    
    /**
     * Export to CSV
     * @param {Object} data - Data to export
     * @param {string} template - Template name
     * @param {Object} options - Export options
     * @returns {Promise<Object>} Export result
     */
    async exportToCSV(data, template, options = {}) {
        // Convert data to CSV format
        const csv = this.convertToCSV(data, template);
        
        // Create blob and download
        const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        
        const filename = options.filename || this.generateFilename(template, 'csv');
        link.download = filename;
        link.click();
        
        URL.revokeObjectURL(url);
        
        return { filename, format: 'csv' };
    }
    
    /**
     * Export to JSON
     * @param {Object} data - Data to export
     * @param {string} template - Template name
     * @param {Object} options - Export options
     * @returns {Promise<Object>} Export result
     */
    async exportToJSON(data, template, options = {}) {
        const json = JSON.stringify(data, null, 2);
        
        // Create blob and download
        const blob = new Blob([json], { type: 'application/json;charset=utf-8;' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        
        const filename = options.filename || this.generateFilename(template, 'json');
        link.download = filename;
        link.click();
        
        URL.revokeObjectURL(url);
        
        return { filename, format: 'json' };
    }
    
    /**
     * Add PDF header
     */
    addPDFHeader(doc, data, templateConfig) {
        doc.setFontSize(20);
        doc.setFont('helvetica', 'bold');
        doc.text('KYB Platform', 20, 20);
        
        doc.setFontSize(16);
        doc.text(templateConfig.name, 20, 35);
        
        if (data.merchant) {
            doc.setFontSize(12);
            doc.setFont('helvetica', 'normal');
            doc.text(`Merchant: ${data.merchant.name || data.merchant.id}`, 20, 45);
        }
        
        doc.text(`Date: ${new Date().toLocaleDateString()}`, 20, 50);
    }
    
    /**
     * Add PDF section
     */
    async addPDFSection(doc, section, data, yPosition) {
        doc.setFontSize(14);
        doc.setFont('helvetica', 'bold');
        
        const sectionTitle = this.getSectionTitle(section);
        doc.text(sectionTitle, 20, yPosition);
        yPosition += 10;
        
        doc.setFontSize(10);
        doc.setFont('helvetica', 'normal');
        
        const content = this.generateSectionContent(section, data);
        const lines = doc.splitTextToSize(content, 170);
        
        for (const line of lines) {
            if (yPosition > 270) {
                doc.addPage();
                yPosition = 30;
            }
            doc.text(line, 20, yPosition);
            yPosition += 5;
        }
        
        return yPosition + 10;
    }
    
    /**
     * Add PDF charts
     */
    async addPDFCharts(doc, charts, yPosition) {
        for (const [chartName, chartDataUrl] of Object.entries(charts)) {
            if (yPosition > 200) {
                doc.addPage();
                yPosition = 30;
            }
            
            try {
                // Add chart image to PDF
                doc.addImage(chartDataUrl, 'PNG', 20, yPosition, 170, 100);
                yPosition += 110;
            } catch (error) {
                console.warn(`Failed to add chart ${chartName} to PDF:`, error);
            }
        }
        
        return yPosition;
    }
    
    /**
     * Add PDF footer
     */
    addPDFFooter(doc) {
        const pageCount = doc.internal.getNumberOfPages();
        for (let i = 1; i <= pageCount; i++) {
            doc.setPage(i);
            doc.setFontSize(8);
            doc.setFont('helvetica', 'normal');
            doc.text(
                `Page ${i} of ${pageCount} - Generated by KYB Platform`,
                105,
                285,
                { align: 'center' }
            );
        }
    }
    
    /**
     * Get section title
     */
    getSectionTitle(section) {
        const titles = {
            'executive-summary': 'Executive Summary',
            'risk-overview': 'Risk Overview',
            'risk-factors': 'Risk Factors',
            'recommendations': 'Recommendations',
            'summary': 'Summary',
            'data': 'Data',
            'alerts': 'Alerts'
        };
        return titles[section] || section;
    }
    
    /**
     * Generate section content
     */
    generateSectionContent(section, data) {
        // Basic content generation - can be extended
        if (section === 'summary' && data.summary) {
            return data.summary;
        }
        
        if (section === 'risk-overview' && data.riskAssessment) {
            return `Overall Risk Score: ${data.riskAssessment.overallScore || 'N/A'}`;
        }
        
        return `Content for ${section} section.`;
    }
    
    /**
     * Prepare Excel data
     */
    prepareExcelData(data, templateConfig) {
        const rows = [];
        
        // Header row
        rows.push([templateConfig.name]);
        rows.push([]);
        
        // Add data rows based on template
        if (data.riskAssessment) {
            rows.push(['Overall Risk Score', data.riskAssessment.overallScore || 'N/A']);
        }
        
        return rows;
    }
    
    /**
     * Convert to CSV
     */
    convertToCSV(data, template) {
        const rows = [];
        
        // Add header
        rows.push(['Field', 'Value']);
        
        // Add data
        if (data.riskAssessment) {
            rows.push(['Overall Risk Score', data.riskAssessment.overallScore || 'N/A']);
        }
        
        return rows.map(row => row.map(cell => `"${cell}"`).join(',')).join('\n');
    }
    
    /**
     * Generate filename
     */
    generateFilename(template, extension) {
        const timestamp = new Date().toISOString().split('T')[0];
        return `${template}-${timestamp}.${extension}`;
    }
    
    /**
     * Process export queue
     */
    async processExportQueue() {
        if (this.exportQueue.length > 0 && !this.isExporting) {
            const { data, options } = this.exportQueue.shift();
            await this.exportData(data, options);
        }
    }
}

// Export singleton instance
let exportServiceInstance = null;

export function getExportService(config) {
    if (!exportServiceInstance) {
        exportServiceInstance = new SharedExportService(config);
    }
    return exportServiceInstance;
}

export { SharedExportService };

// Make available globally for non-module environments
if (typeof window !== 'undefined') {
    window.getExportService = getExportService;
    window.SharedExportService = SharedExportService;
}

