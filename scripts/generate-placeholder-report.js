#!/usr/bin/env node

/**
 * Generate Placeholder Detection Report
 * 
 * Generates a formatted report from placeholder detection results,
 * including comparison with previous runs and recommendations.
 */

const fs = require('fs');
const path = require('path');

class PlaceholderReportGenerator {
    constructor(reportPath, previousReportPath = null) {
        this.reportPath = reportPath;
        this.previousReportPath = previousReportPath;
        this.report = this.loadReport();
        this.previousReport = previousReportPath ? this.loadReport(previousReportPath) : null;
    }
    
    loadReport(filePath = null) {
        const pathToLoad = filePath || this.reportPath;
        try {
            const content = fs.readFileSync(pathToLoad, 'utf8');
            return JSON.parse(content);
        } catch (error) {
            console.error(`Error loading report: ${error.message}`);
            return null;
        }
    }
    
    generateMarkdown() {
        let markdown = '# Placeholder Data Detection Report\n\n';
        markdown += `**Generated:** ${new Date(this.report.timestamp).toLocaleString()}\n\n`;
        
        // Summary
        markdown += '## Summary\n\n';
        markdown += `- **Total Findings:** ${this.report.summary.total_findings}\n`;
        markdown += `- **Critical:** ${this.report.summary.by_severity.critical}\n`;
        markdown += `- **High:** ${this.report.summary.by_severity.high}\n`;
        markdown += `- **Medium:** ${this.report.summary.by_severity.medium}\n`;
        markdown += `- **Low:** ${this.report.summary.by_severity.low}\n`;
        markdown += `- **Info:** ${this.report.summary.by_severity.info}\n\n`;
        
        // Comparison with previous run
        if (this.previousReport) {
            markdown += '## Comparison with Previous Run\n\n';
            const prev = this.previousReport.summary;
            const curr = this.report.summary;
            
            markdown += `| Metric | Previous | Current | Change |\n`;
            markdown += `|--------|----------|---------|--------|\n`;
            markdown += `| Total | ${prev.total_findings} | ${curr.total_findings} | ${curr.total_findings - prev.total_findings} |\n`;
            markdown += `| Critical | ${prev.by_severity.critical} | ${curr.by_severity.critical} | ${curr.by_severity.critical - prev.by_severity.critical} |\n`;
            markdown += `| High | ${prev.by_severity.high} | ${curr.by_severity.high} | ${curr.by_severity.high - prev.by_severity.high} |\n\n`;
        }
        
        // Findings by severity
        const severities = ['critical', 'high', 'medium', 'low', 'info'];
        severities.forEach(severity => {
            const findings = this.report.findings.filter(f => f.severity === severity);
            if (findings.length === 0) return;
            
            markdown += `## ${severity.toUpperCase()} Severity Findings (${findings.length})\n\n`;
            
            findings.forEach((finding, index) => {
                markdown += `### ${index + 1}. ${finding.file}:${finding.line}\n\n`;
                markdown += `- **Pattern:** ${finding.pattern}\n`;
                markdown += `- **Category:** ${finding.category}\n`;
                markdown += `- **Is Fallback:** ${finding.is_fallback}\n`;
                markdown += `- **Has Error Handling:** ${finding.has_error_handling}\n`;
                markdown += `- **Recommendation:** ${finding.recommendation}\n\n`;
                markdown += `\`\`\`\n${finding.context}\n\`\`\`\n\n`;
            });
        });
        
        return markdown;
    }
    
    saveMarkdown(outputPath) {
        const markdown = this.generateMarkdown();
        fs.writeFileSync(outputPath, markdown);
        console.log(`Markdown report saved to: ${outputPath}`);
    }
}

// Run if called directly
if (require.main === module) {
    const reportPath = process.argv[2];
    const previousReportPath = process.argv[3] || null;
    const outputPath = process.argv[4] || reportPath.replace('.json', '.md');
    
    if (!reportPath) {
        console.error('Usage: node generate-placeholder-report.js <report.json> [previous-report.json] [output.md]');
        process.exit(1);
    }
    
    const generator = new PlaceholderReportGenerator(reportPath, previousReportPath);
    generator.saveMarkdown(outputPath);
}

module.exports = PlaceholderReportGenerator;

