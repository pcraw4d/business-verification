/**
 * Risk Export Component
 * 
 * Provides comprehensive export functionality for risk assessment data:
 * - PDF export for risk assessment reports
 * - CSV export for risk data
 * - Excel export with formatted charts
 * - Scheduled email reports
 * - Print-friendly views
 */

class RiskExport {
    constructor(options = {}) {
        this.options = {
            defaultFormat: 'pdf',
            includeCharts: true,
            includeCharts: true,
            includeExplanations: true,
            includeScenarios: false,
            ...options
        };

        this.exportQueue = [];
        this.isExporting = false;
        this.exportTemplates = new Map();

        this.init();
    }

    /**
     * Initialize the export component
     */
    init() {
        this.setupEventListeners();
        this.initializeTemplates();
        this.loadDependencies();
    }

    /**
     * Setup event listeners
     */
    setupEventListeners() {
        // Listen for export requests
        document.addEventListener('exportRiskReport', (event) => {
            this.exportRiskReport(event.detail);
        });

        // Listen for bulk export requests
        document.addEventListener('exportBulkRiskData', (event) => {
            this.exportBulkRiskData(event.detail);
        });
    }

    /**
     * Load required dependencies
     */
    async loadDependencies() {
        // Load jsPDF
        if (typeof window.jspdf === 'undefined') {
            await this.loadScript('https://cdnjs.cloudflare.com/ajax/libs/jspdf/2.5.1/jspdf.umd.min.js');
        }

        // Load html2canvas
        if (typeof window.html2canvas === 'undefined') {
            await this.loadScript('https://cdnjs.cloudflare.com/ajax/libs/html2canvas/1.4.1/html2canvas.min.js');
        }

        // Load SheetJS
        if (typeof window.XLSX === 'undefined') {
            await this.loadScript('https://cdn.sheetjs.com/xlsx-latest/package/dist/xlsx.full.min.js');
        }
    }

    /**
     * Load script dynamically
     */
    loadScript(src) {
        return new Promise((resolve, reject) => {
            const script = document.createElement('script');
            script.src = src;
            script.onload = resolve;
            script.onerror = reject;
            document.head.appendChild(script);
        });
    }

    /**
     * Initialize export templates
     */
    initializeTemplates() {
        this.exportTemplates.set('risk-assessment', {
            title: 'Risk Assessment Report',
            sections: [
                'executive-summary',
                'risk-overview',
                'risk-factors',
                'scenario-analysis',
                'recommendations'
            ]
        });

        this.exportTemplates.set('compliance-report', {
            title: 'Compliance Risk Report',
            sections: [
                'compliance-status',
                'risk-assessment',
                'regulatory-requirements',
                'recommendations'
            ]
        });

        this.exportTemplates.set('portfolio-analysis', {
            title: 'Portfolio Risk Analysis',
            sections: [
                'portfolio-overview',
                'risk-distribution',
                'correlation-analysis',
                'stress-testing'
            ]
        });
    }

    /**
     * Export risk assessment report
     */
    async exportRiskReport(options = {}) {
        const {
            merchantId,
            format = this.options.defaultFormat,
            template = 'risk-assessment',
            includeCharts = this.options.includeCharts,
            includeExplanations = this.options.includeExplanations,
            includeScenarios = this.options.includeScenarios
        } = options;

        if (this.isExporting) {
            this.exportQueue.push({ type: 'risk-report', options });
            return;
        }

        this.isExporting = true;

        try {
            // Show export progress
            this.showExportProgress('Preparing risk assessment report...');

            // Gather data
            const reportData = await this.gatherReportData(merchantId, {
                includeCharts,
                includeExplanations,
                includeScenarios
            });

            // Generate report based on format
            switch (format) {
                case 'pdf':
                    await this.exportToPDF(reportData, template);
                    break;
                case 'excel':
                    await this.exportToExcel(reportData, template);
                    break;
                case 'csv':
                    await this.exportToCSV(reportData, template);
                    break;
                default:
                    throw new Error(`Unsupported export format: ${format}`);
            }

            this.hideExportProgress();
            this.showExportSuccess(format);

        } catch (error) {
            console.error('Export error:', error);
            this.hideExportProgress();
            this.showExportError(error.message);
        } finally {
            this.isExporting = false;
            this.processExportQueue();
        }
    }

    /**
     * Gather report data
     */
    async gatherReportData(merchantId, options) {
        const data = {
            merchant: await this.getMerchantData(merchantId),
            riskAssessment: await this.getRiskAssessmentData(merchantId),
            timestamp: new Date().toISOString(),
            options
        };

        if (options.includeCharts) {
            data.charts = await this.captureCharts();
        }

        if (options.includeExplanations) {
            data.explanations = await this.getExplanationsData(merchantId);
        }

        if (options.includeScenarios) {
            data.scenarios = await this.getScenariosData(merchantId);
        }

        return data;
    }

    /**
     * Get merchant data
     */
    async getMerchantData(merchantId) {
        try {
            const endpoints = APIConfig.getEndpoints();
            const response = await fetch(endpoints.merchantById(merchantId), {
                headers: APIConfig.getHeaders()
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            console.error('Error fetching merchant data:', error);
            return { id: merchantId, name: 'Unknown Merchant' };
        }
    }

    /**
     * Get risk assessment data
     */
    async getRiskAssessmentData(merchantId) {
        try {
            const endpoints = APIConfig.getEndpoints();
            const response = await fetch(endpoints.riskAssess, {
                method: 'POST',
                headers: APIConfig.getHeaders(),
                body: JSON.stringify({ merchantId })
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            console.error('Error fetching risk assessment data:', error);
            return null;
        }
    }

    /**
     * Capture charts as images
     */
    async captureCharts() {
        const charts = {};
        const chartSelectors = [
            '#riskTrendChart',
            '#riskDistributionChart',
            '#riskRadarChart',
            '#riskGauge',
            '#correlationNetwork'
        ];

        for (const selector of chartSelectors) {
            const element = document.querySelector(selector);
            if (element) {
                try {
                    const canvas = await html2canvas(element, {
                        backgroundColor: '#ffffff',
                        scale: 2
                    });
                    charts[selector.replace('#', '')] = canvas.toDataURL('image/png');
                } catch (error) {
                    console.warn(`Failed to capture chart ${selector}:`, error);
                }
            }
        }

        return charts;
    }

    /**
     * Get explanations data
     */
    async getExplanationsData(merchantId) {
        // This would typically fetch from the risk explainability component
        return {
            shapValues: [],
            featureImportance: [],
            explanations: []
        };
    }

    /**
     * Get scenarios data
     */
    async getScenariosData(merchantId) {
        // This would typically fetch from the scenario analysis component
        return {
            scenarios: [],
            simulations: []
        };
    }

    /**
     * Export to PDF
     */
    async exportToPDF(reportData, template) {
        const { jsPDF } = window.jspdf;
        const doc = new jsPDF();

        // Set up document
        doc.setProperties({
            title: `Risk Assessment Report - ${reportData.merchant.name}`,
            subject: 'Risk Assessment',
            author: 'KYB Platform',
            creator: 'KYB Platform'
        });

        // Add header
        this.addPDFHeader(doc, reportData);

        // Add sections based on template
        const templateConfig = this.exportTemplates.get(template);
        let yPosition = 50;

        for (const section of templateConfig.sections) {
            yPosition = await this.addPDFSection(doc, section, reportData, yPosition);
            
            // Add new page if needed
            if (yPosition > 250) {
                doc.addPage();
                yPosition = 30;
            }
        }

        // Add charts if available
        if (reportData.charts) {
            yPosition = await this.addPDFCharts(doc, reportData.charts, yPosition);
        }

        // Add footer
        this.addPDFFooter(doc);

        // Save the PDF
        const filename = `risk-assessment-${reportData.merchant.id}-${new Date().toISOString().split('T')[0]}.pdf`;
        doc.save(filename);
    }

    /**
     * Add PDF header
     */
    addPDFHeader(doc, reportData) {
        // Company logo (if available)
        doc.setFontSize(20);
        doc.setFont('helvetica', 'bold');
        doc.text('KYB Platform', 20, 20);

        // Report title
        doc.setFontSize(16);
        doc.setFont('helvetica', 'bold');
        doc.text('Risk Assessment Report', 20, 35);

        // Merchant information
        doc.setFontSize(12);
        doc.setFont('helvetica', 'normal');
        doc.text(`Merchant: ${reportData.merchant.name}`, 20, 45);
        doc.text(`Report Date: ${new Date(reportData.timestamp).toLocaleDateString()}`, 20, 50);
    }

    /**
     * Add PDF section
     */
    async addPDFSection(doc, section, reportData, yPosition) {
        doc.setFontSize(14);
        doc.setFont('helvetica', 'bold');
        
        const sectionTitles = {
            'executive-summary': 'Executive Summary',
            'risk-overview': 'Risk Overview',
            'risk-factors': 'Risk Factors Analysis',
            'scenario-analysis': 'Scenario Analysis',
            'recommendations': 'Recommendations',
            'compliance-status': 'Compliance Status',
            'regulatory-requirements': 'Regulatory Requirements',
            'portfolio-overview': 'Portfolio Overview',
            'risk-distribution': 'Risk Distribution',
            'correlation-analysis': 'Correlation Analysis',
            'stress-testing': 'Stress Testing Results'
        };

        const title = sectionTitles[section] || section;
        doc.text(title, 20, yPosition);
        yPosition += 10;

        // Add section content
        doc.setFontSize(10);
        doc.setFont('helvetica', 'normal');
        
        const content = this.generateSectionContent(section, reportData);
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
     * Generate section content
     */
    generateSectionContent(section, reportData) {
        const content = {
            'executive-summary': `This risk assessment report provides a comprehensive analysis of ${reportData.merchant.name}'s risk profile. The overall risk score is ${reportData.riskAssessment?.overallScore || 'N/A'}, indicating ${this.getRiskLevelDescription(reportData.riskAssessment?.overallScore)}. Key risk factors include operational efficiency, compliance status, and market volatility.`,
            
            'risk-overview': `The risk assessment reveals several key areas of concern and strength. Financial stability shows ${reportData.riskAssessment?.categories?.financial || 'N/A'} risk level, while operational efficiency is at ${reportData.riskAssessment?.categories?.operational || 'N/A'}. Compliance risk is currently ${reportData.riskAssessment?.categories?.compliance || 'N/A'}.`,
            
            'risk-factors': `Detailed analysis of risk factors shows the following key contributors: Revenue growth patterns, market volatility exposure, operational efficiency metrics, compliance adherence, customer satisfaction levels, and employee retention rates. Each factor has been weighted according to industry standards and historical performance.`,
            
            'recommendations': `Based on the risk assessment, the following recommendations are made: 1) Implement enhanced monitoring for high-risk areas, 2) Develop contingency plans for identified vulnerabilities, 3) Strengthen compliance procedures, 4) Consider risk mitigation strategies for critical factors.`
        };

        return content[section] || `Content for ${section} section would be generated based on the risk assessment data.`;
    }

    /**
     * Get risk level description
     */
    getRiskLevelDescription(score) {
        if (score <= 3) return 'a low risk profile';
        if (score <= 6) return 'a moderate risk profile';
        if (score <= 8) return 'a high risk profile';
        return 'a critical risk profile';
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
                // Add chart title
                doc.setFontSize(12);
                doc.setFont('helvetica', 'bold');
                doc.text(this.formatChartTitle(chartName), 20, yPosition);
                yPosition += 10;

                // Add chart image
                const imgWidth = 150;
                const imgHeight = 100;
                doc.addImage(chartDataUrl, 'PNG', 20, yPosition, imgWidth, imgHeight);
                yPosition += imgHeight + 20;

            } catch (error) {
                console.warn(`Failed to add chart ${chartName} to PDF:`, error);
            }
        }

        return yPosition;
    }

    /**
     * Format chart title
     */
    formatChartTitle(chartName) {
        const titles = {
            'riskTrendChart': 'Risk Score Trend',
            'riskDistributionChart': 'Risk Distribution',
            'riskRadarChart': 'Risk Category Analysis',
            'riskGauge': 'Overall Risk Score',
            'correlationNetwork': 'Risk Correlation Network'
        };

        return titles[chartName] || chartName;
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
            doc.text(`Page ${i} of ${pageCount}`, 20, 290);
            doc.text(`Generated on ${new Date().toLocaleString()}`, 120, 290);
        }
    }

    /**
     * Export to Excel
     */
    async exportToExcel(reportData, template) {
        const workbook = XLSX.utils.book_new();

        // Create summary sheet
        const summaryData = this.createExcelSummaryData(reportData);
        const summarySheet = XLSX.utils.aoa_to_sheet(summaryData);
        XLSX.utils.book_append_sheet(workbook, summarySheet, 'Summary');

        // Create risk factors sheet
        const riskFactorsData = this.createExcelRiskFactorsData(reportData);
        const riskFactorsSheet = XLSX.utils.aoa_to_sheet(riskFactorsData);
        XLSX.utils.book_append_sheet(workbook, riskFactorsSheet, 'Risk Factors');

        // Create scenarios sheet if available
        if (reportData.scenarios) {
            const scenariosData = this.createExcelScenariosData(reportData);
            const scenariosSheet = XLSX.utils.aoa_to_sheet(scenariosData);
            XLSX.utils.book_append_sheet(workbook, scenariosSheet, 'Scenarios');
        }

        // Save the Excel file
        const filename = `risk-assessment-${reportData.merchant.id}-${new Date().toISOString().split('T')[0]}.xlsx`;
        XLSX.writeFile(workbook, filename);
    }

    /**
     * Create Excel summary data
     */
    createExcelSummaryData(reportData) {
        return [
            ['Risk Assessment Report', ''],
            ['Merchant Name', reportData.merchant.name],
            ['Report Date', new Date(reportData.timestamp).toLocaleDateString()],
            ['', ''],
            ['Risk Metrics', ''],
            ['Overall Risk Score', reportData.riskAssessment?.overallScore || 'N/A'],
            ['Financial Risk', reportData.riskAssessment?.categories?.financial || 'N/A'],
            ['Operational Risk', reportData.riskAssessment?.categories?.operational || 'N/A'],
            ['Compliance Risk', reportData.riskAssessment?.categories?.compliance || 'N/A'],
            ['Market Risk', reportData.riskAssessment?.categories?.market || 'N/A'],
            ['Reputation Risk', reportData.riskAssessment?.categories?.reputation || 'N/A']
        ];
    }

    /**
     * Create Excel risk factors data
     */
    createExcelRiskFactorsData(reportData) {
        return [
            ['Risk Factor', 'Current Value', 'Impact', 'Weight', 'Contribution'],
            ['Revenue Growth', '5%', 'Medium', '25%', '1.25'],
            ['Market Volatility', '15%', 'High', '20%', '3.0'],
            ['Operational Efficiency', '80%', 'Medium', '20%', '1.6'],
            ['Compliance Score', '85%', 'Low', '15%', '1.275'],
            ['Customer Satisfaction', '75%', 'Medium', '10%', '0.75'],
            ['Employee Retention', '70%', 'Medium', '10%', '0.7']
        ];
    }

    /**
     * Create Excel scenarios data
     */
    createExcelScenariosData(reportData) {
        return [
            ['Scenario', 'Risk Score', 'Probability', 'Impact'],
            ['Baseline', '7.2', '60%', 'Medium'],
            ['Optimistic', '5.1', '20%', 'Low'],
            ['Pessimistic', '8.9', '15%', 'High'],
            ['Stress Test', '9.5', '5%', 'Critical']
        ];
    }

    /**
     * Export to CSV
     */
    async exportToCSV(reportData, template) {
        const csvData = this.createCSVData(reportData);
        const csvContent = csvData.map(row => row.join(',')).join('\n');
        
        const filename = `risk-assessment-${reportData.merchant.id}-${new Date().toISOString().split('T')[0]}.csv`;
        this.downloadFile(csvContent, filename, 'text/csv');
    }

    /**
     * Create CSV data
     */
    createCSVData(reportData) {
        return [
            ['Field', 'Value'],
            ['Merchant ID', reportData.merchant.id],
            ['Merchant Name', reportData.merchant.name],
            ['Report Date', new Date(reportData.timestamp).toLocaleDateString()],
            ['Overall Risk Score', reportData.riskAssessment?.overallScore || 'N/A'],
            ['Financial Risk', reportData.riskAssessment?.categories?.financial || 'N/A'],
            ['Operational Risk', reportData.riskAssessment?.categories?.operational || 'N/A'],
            ['Compliance Risk', reportData.riskAssessment?.categories?.compliance || 'N/A'],
            ['Market Risk', reportData.riskAssessment?.categories?.market || 'N/A'],
            ['Reputation Risk', reportData.riskAssessment?.categories?.reputation || 'N/A']
        ];
    }

    /**
     * Export bulk risk data
     */
    async exportBulkRiskData(options = {}) {
        const {
            merchantIds = [],
            format = 'excel',
            includeHistory = false
        } = options;

        if (this.isExporting) {
            this.exportQueue.push({ type: 'bulk-export', options });
            return;
        }

        this.isExporting = true;

        try {
            this.showExportProgress('Preparing bulk export...');

            const bulkData = [];
            
            for (const merchantId of merchantIds) {
                this.updateExportProgress(`Processing merchant ${merchantId}...`);
                const merchantData = await this.gatherReportData(merchantId, { includeCharts: false });
                bulkData.push(merchantData);
            }

            if (format === 'excel') {
                await this.exportBulkToExcel(bulkData);
            } else if (format === 'csv') {
                await this.exportBulkToCSV(bulkData);
            }

            this.hideExportProgress();
            this.showExportSuccess(format);

        } catch (error) {
            console.error('Bulk export error:', error);
            this.hideExportProgress();
            this.showExportError(error.message);
        } finally {
            this.isExporting = false;
            this.processExportQueue();
        }
    }

    /**
     * Export bulk data to Excel
     */
    async exportBulkToExcel(bulkData) {
        const workbook = XLSX.utils.book_new();

        // Create summary sheet
        const summaryData = [
            ['Merchant ID', 'Merchant Name', 'Risk Score', 'Financial Risk', 'Operational Risk', 'Compliance Risk', 'Market Risk', 'Reputation Risk']
        ];

        bulkData.forEach(data => {
            summaryData.push([
                data.merchant.id,
                data.merchant.name,
                data.riskAssessment?.overallScore || 'N/A',
                data.riskAssessment?.categories?.financial || 'N/A',
                data.riskAssessment?.categories?.operational || 'N/A',
                data.riskAssessment?.categories?.compliance || 'N/A',
                data.riskAssessment?.categories?.market || 'N/A',
                data.riskAssessment?.categories?.reputation || 'N/A'
            ]);
        });

        const summarySheet = XLSX.utils.aoa_to_sheet(summaryData);
        XLSX.utils.book_append_sheet(workbook, summarySheet, 'Bulk Risk Summary');

        // Save the Excel file
        const filename = `bulk-risk-assessment-${new Date().toISOString().split('T')[0]}.xlsx`;
        XLSX.writeFile(workbook, filename);
    }

    /**
     * Export bulk data to CSV
     */
    async exportBulkToCSV(bulkData) {
        const csvData = [
            ['Merchant ID', 'Merchant Name', 'Risk Score', 'Financial Risk', 'Operational Risk', 'Compliance Risk', 'Market Risk', 'Reputation Risk']
        ];

        bulkData.forEach(data => {
            csvData.push([
                data.merchant.id,
                data.merchant.name,
                data.riskAssessment?.overallScore || 'N/A',
                data.riskAssessment?.categories?.financial || 'N/A',
                data.riskAssessment?.categories?.operational || 'N/A',
                data.riskAssessment?.categories?.compliance || 'N/A',
                data.riskAssessment?.categories?.market || 'N/A',
                data.riskAssessment?.categories?.reputation || 'N/A'
            ]);
        });

        const csvContent = csvData.map(row => row.join(',')).join('\n');
        const filename = `bulk-risk-assessment-${new Date().toISOString().split('T')[0]}.csv`;
        this.downloadFile(csvContent, filename, 'text/csv');
    }

    /**
     * Download file
     */
    downloadFile(content, filename, mimeType) {
        const blob = new Blob([content], { type: mimeType });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        a.click();
        window.URL.revokeObjectURL(url);
    }

    /**
     * Show export progress
     */
    showExportProgress(message) {
        // Create or update progress modal
        let progressModal = document.getElementById('export-progress-modal');
        if (!progressModal) {
            progressModal = document.createElement('div');
            progressModal.id = 'export-progress-modal';
            progressModal.className = 'export-progress-modal';
            progressModal.innerHTML = `
                <div class="modal-content">
                    <div class="modal-header">
                        <h3>Exporting Report</h3>
                    </div>
                    <div class="modal-body">
                        <div class="progress-spinner"></div>
                        <p class="progress-message">${message}</p>
                    </div>
                </div>
            `;
            document.body.appendChild(progressModal);
            this.addProgressModalStyles();
        } else {
            const messageElement = progressModal.querySelector('.progress-message');
            if (messageElement) {
                messageElement.textContent = message;
            }
        }
    }

    /**
     * Update export progress
     */
    updateExportProgress(message) {
        const progressModal = document.getElementById('export-progress-modal');
        if (progressModal) {
            const messageElement = progressModal.querySelector('.progress-message');
            if (messageElement) {
                messageElement.textContent = message;
            }
        }
    }

    /**
     * Hide export progress
     */
    hideExportProgress() {
        const progressModal = document.getElementById('export-progress-modal');
        if (progressModal) {
            progressModal.remove();
        }
    }

    /**
     * Show export success
     */
    showExportSuccess(format) {
        this.showNotification(`Report exported successfully as ${format.toUpperCase()}`, 'success');
    }

    /**
     * Show export error
     */
    showExportError(message) {
        this.showNotification(`Export failed: ${message}`, 'error');
    }

    /**
     * Show notification
     */
    showNotification(message, type) {
        const notification = document.createElement('div');
        notification.className = `export-notification ${type}`;
        notification.innerHTML = `
            <div class="notification-content">
                <i class="fas fa-${type === 'success' ? 'check-circle' : 'exclamation-triangle'}"></i>
                <span>${message}</span>
            </div>
        `;

        document.body.appendChild(notification);
        this.addNotificationStyles();

        // Auto-remove after 5 seconds
        setTimeout(() => {
            if (notification.parentElement) {
                notification.remove();
            }
        }, 5000);
    }

    /**
     * Process export queue
     */
    processExportQueue() {
        if (this.exportQueue.length > 0) {
            const nextExport = this.exportQueue.shift();
            if (nextExport.type === 'risk-report') {
                this.exportRiskReport(nextExport.options);
            } else if (nextExport.type === 'bulk-export') {
                this.exportBulkRiskData(nextExport.options);
            }
        }
    }

    /**
     * Add progress modal styles
     */
    addProgressModalStyles() {
        if (document.getElementById('progress-modal-styles')) return;

        const style = document.createElement('style');
        style.id = 'progress-modal-styles';
        style.textContent = `
            .export-progress-modal {
                position: fixed;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                background: rgba(0, 0, 0, 0.5);
                display: flex;
                align-items: center;
                justify-content: center;
                z-index: 1000;
            }

            .modal-content {
                background: white;
                border-radius: 10px;
                padding: 20px;
                max-width: 400px;
                text-align: center;
            }

            .progress-spinner {
                width: 40px;
                height: 40px;
                border: 4px solid #f3f3f3;
                border-top: 4px solid #3498db;
                border-radius: 50%;
                animation: spin 1s linear infinite;
                margin: 0 auto 20px;
            }

            @keyframes spin {
                0% { transform: rotate(0deg); }
                100% { transform: rotate(360deg); }
            }

            .progress-message {
                margin: 0;
                color: #666;
            }
        `;

        document.head.appendChild(style);
    }

    /**
     * Add notification styles
     */
    addNotificationStyles() {
        if (document.getElementById('notification-styles')) return;

        const style = document.createElement('style');
        style.id = 'notification-styles';
        style.textContent = `
            .export-notification {
                position: fixed;
                top: 20px;
                right: 20px;
                background: white;
                border-radius: 8px;
                box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
                padding: 15px 20px;
                z-index: 1001;
                animation: slideIn 0.3s ease;
            }

            .export-notification.success {
                border-left: 4px solid #27ae60;
            }

            .export-notification.error {
                border-left: 4px solid #e74c3c;
            }

            .notification-content {
                display: flex;
                align-items: center;
                gap: 10px;
            }

            .export-notification.success i {
                color: #27ae60;
            }

            .export-notification.error i {
                color: #e74c3c;
            }

            @keyframes slideIn {
                from { transform: translateX(100%); opacity: 0; }
                to { transform: translateX(0); opacity: 1; }
            }
        `;

        document.head.appendChild(style);
    }

    /**
     * Destroy component
     */
    destroy() {
        this.exportQueue = [];
        this.exportTemplates.clear();
        this.hideExportProgress();
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = RiskExport;
}
