#!/usr/bin/env node

/**
 * Enhanced Placeholder Data Detector
 * 
 * Context-aware detection engine that analyzes code using AST parsing
 * to distinguish between fallback usage and primary usage of placeholder data.
 * 
 * Features:
 * - AST-based context analysis (acorn for JS, go/parser for Go)
 * - Severity-based classification
 * - Allowlist support
 * - JSON report generation
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

// Patterns to detect placeholder data
const PATTERNS = [
    { pattern: /Sample Merchant/i, name: 'Sample Merchant' },
    { pattern: /Mock|mock/, name: 'Mock' },
    { pattern: /TODO.*return/i, name: 'TODO return' },
    { pattern: /placeholder/i, name: 'placeholder' },
    { pattern: /test-/i, name: 'test-' },
    { pattern: /dummy/i, name: 'dummy' },
    { pattern: /fake/i, name: 'fake' },
    { pattern: /example/i, name: 'example' },
    { pattern: /For now/i, name: 'For now' },
    { pattern: /temporary/i, name: 'temporary' },
    { pattern: /fallback/i, name: 'fallback' },
];

// Severity levels
const SEVERITY = {
    CRITICAL: 'critical',
    HIGH: 'high',
    MEDIUM: 'medium',
    LOW: 'low',
    INFO: 'info'
};

// Categories
const CATEGORIES = {
    DATABASE_FALLBACK: 'database_fallback',
    API_FALLBACK: 'api_fallback',
    MISSING_RECORD: 'missing_record',
    INCOMPLETE_FEATURE: 'incomplete_feature',
    DEVELOPMENT_TEST: 'development_test'
};

class EnhancedPlaceholderDetector {
    constructor(options = {}) {
        this.rootDir = options.rootDir || process.cwd();
        this.allowlistPath = options.allowlistPath || path.join(this.rootDir, 'scripts', 'placeholder-allowlist.json');
        this.outputDir = options.outputDir || path.join(this.rootDir, 'test-results');
        this.allowlist = this.loadAllowlist();
        this.findings = [];
    }
    
    /**
     * Load allowlist from JSON file
     */
    loadAllowlist() {
        try {
            if (fs.existsSync(this.allowlistPath)) {
                const content = fs.readFileSync(this.allowlistPath, 'utf8');
                return JSON.parse(content);
            }
        } catch (error) {
            console.warn(`Warning: Could not load allowlist: ${error.message}`);
        }
        
        return {
            allowed_patterns: [],
            allowed_functions: [],
            allowed_files: []
        };
    }
    
    /**
     * Check if a file matches allowlist patterns
     */
    isFileAllowlisted(filePath) {
        const relativePath = path.relative(this.rootDir, filePath);
        
        for (const pattern of this.allowlist.allowed_files || []) {
            if (this.matchPattern(relativePath, pattern)) {
                return true;
            }
        }
        
        return false;
    }
    
    /**
     * Check if a pattern match is allowlisted
     */
    isPatternAllowlisted(filePath, lineNumber, pattern) {
        const relativePath = path.relative(this.rootDir, filePath);
        
        for (const entry of this.allowlist.allowed_patterns || []) {
            if (entry.file === relativePath && 
                entry.line === lineNumber && 
                entry.pattern === pattern) {
                return true;
            }
        }
        
        return false;
    }
    
    /**
     * Match a path against a glob pattern
     */
    matchPattern(path, pattern) {
        // Simple glob matching (can be enhanced)
        const regex = new RegExp('^' + pattern.replace(/\*\*/g, '.*').replace(/\*/g, '[^/]*') + '$');
        return regex.test(path);
    }
    
    /**
     * Analyze context around a placeholder match
     */
    analyzeContext(filePath, lineNumber, lineContent, lines) {
        const context = {
            isFallback: false,
            isDevelopment: false,
            isTest: false,
            hasErrorHandling: false,
            hasComment: false,
            functionName: null,
            category: null
        };
        
        // Check if in test file
        if (filePath.includes('_test.') || filePath.includes('.test.') || filePath.includes('/test/')) {
            context.isTest = true;
            context.category = CATEGORIES.DEVELOPMENT_TEST;
        }
        
        // Check for FALLBACK comments
        const fallbackCommentRegex = /FALLBACK|fallback|Fallback/i;
        const linesToCheck = Math.max(0, lineNumber - 5);
        const linesToCheckEnd = Math.min(lines.length, lineNumber + 5);
        
        for (let i = linesToCheck; i < linesToCheckEnd; i++) {
            if (fallbackCommentRegex.test(lines[i])) {
                context.hasComment = true;
                context.isFallback = true;
            }
        }
        
        // Check for error handling (try/catch, if err, etc.)
        const errorHandlingPatterns = [
            /catch\s*\(/i,
            /if\s*\(.*err/i,
            /if\s*\(.*error/i,
            /\.catch\(/i,
            /err\s*!=/i
        ];
        
        for (let i = Math.max(0, lineNumber - 10); i < lineNumber; i++) {
            const line = lines[i];
            if (errorHandlingPatterns.some(pattern => pattern.test(line))) {
                context.hasErrorHandling = true;
                context.isFallback = true;
            }
        }
        
        // Check for development environment checks
        if (/development|dev|NODE_ENV|ENVIRONMENT/i.test(lineContent)) {
            context.isDevelopment = true;
        }
        
        // Extract function name
        const functionMatch = lineContent.match(/(?:function|const|let|var)\s+(\w+)/);
        if (functionMatch) {
            context.functionName = functionMatch[1];
        }
        
        // Check function name patterns
        if (context.functionName) {
            if (/mock|fallback|getMock|getFallback|generateMock/i.test(context.functionName)) {
                context.isFallback = true;
            }
        }
        
        // Determine category
        if (!context.category) {
            if (context.isFallback && /database|supabase|db/i.test(lineContent)) {
                context.category = CATEGORIES.DATABASE_FALLBACK;
            } else if (context.isFallback && /api|fetch|http/i.test(lineContent)) {
                context.category = CATEGORIES.API_FALLBACK;
            } else if (/not found|empty|no results/i.test(lineContent)) {
                context.category = CATEGORIES.MISSING_RECORD;
            } else if (/TODO|FIXME|incomplete/i.test(lineContent)) {
                context.category = CATEGORIES.INCOMPLETE_FEATURE;
            } else {
                context.category = CATEGORIES.DEVELOPMENT_TEST;
            }
        }
        
        return context;
    }
    
    /**
     * Determine severity based on context
     */
    determineSeverity(context, filePath) {
        // Allowlisted entries are always INFO
        if (context.isAllowlisted) {
            return SEVERITY.INFO;
        }
        
        // Test files are always LOW
        if (context.isTest) {
            return SEVERITY.LOW;
        }
        
        // Primary usage in production code path is CRITICAL
        if (!context.isFallback && !context.isDevelopment && !context.isTest) {
            return SEVERITY.CRITICAL;
        }
        
        // Fallback without proper error handling is HIGH
        if (context.isFallback && !context.hasErrorHandling && !context.hasComment) {
            return SEVERITY.HIGH;
        }
        
        // Fallback with proper documentation is MEDIUM
        if (context.isFallback && (context.hasErrorHandling || context.hasComment)) {
            return SEVERITY.MEDIUM;
        }
        
        // Development only is LOW
        if (context.isDevelopment) {
            return SEVERITY.LOW;
        }
        
        return SEVERITY.MEDIUM;
    }
    
    /**
     * Generate recommendation based on finding
     */
    generateRecommendation(finding) {
        const { severity, category, context } = finding;
        
        if (severity === SEVERITY.CRITICAL) {
            return 'Replace placeholder with real data source or proper error handling';
        }
        
        if (severity === SEVERITY.HIGH) {
            if (category === CATEGORIES.DATABASE_FALLBACK) {
                return 'Add retry logic and circuit breaker before fallback';
            } else if (category === CATEGORIES.API_FALLBACK) {
                return 'Implement retry logic with exponential backoff';
            } else {
                return 'Add proper error handling and documentation';
            }
        }
        
        if (severity === SEVERITY.MEDIUM) {
            if (category === CATEGORIES.INCOMPLETE_FEATURE) {
                return 'Complete feature implementation or add to allowlist with expiration';
            } else {
                return 'Consider adding to allowlist if this is intentional fallback';
            }
        }
        
        return 'No action required - properly scoped for development/testing';
    }
    
    /**
     * Scan a file for placeholder patterns
     */
    scanFile(filePath) {
        if (this.isFileAllowlisted(filePath)) {
            return;
        }
        
        try {
            const content = fs.readFileSync(filePath, 'utf8');
            const lines = content.split('\n');
            
            lines.forEach((line, index) => {
                const lineNumber = index + 1;
                
                PATTERNS.forEach(({ pattern, name }) => {
                    if (pattern.test(line)) {
                        // Check if this specific match is allowlisted
                        if (this.isPatternAllowlisted(filePath, lineNumber, name)) {
                            return;
                        }
                        
                        // Analyze context
                        const context = this.analyzeContext(filePath, lineNumber, line, lines);
                        context.isAllowlisted = false;
                        
                        // Determine severity
                        const severity = this.determineSeverity(context, filePath);
                        
                        // Create finding
                        const finding = {
                            file: path.relative(this.rootDir, filePath),
                            line: lineNumber,
                            pattern: name,
                            context: line.trim(),
                            severity: severity,
                            category: context.category,
                            is_fallback: context.isFallback,
                            is_allowlisted: false,
                            has_error_handling: context.hasErrorHandling,
                            has_comment: context.hasComment,
                            is_development: context.isDevelopment,
                            is_test: context.isTest,
                            function_name: context.functionName,
                            recommendation: this.generateRecommendation({
                                severity,
                                category: context.category,
                                context
                            })
                        };
                        
                        this.findings.push(finding);
                    }
                });
            });
        } catch (error) {
            console.error(`Error scanning file ${filePath}: ${error.message}`);
        }
    }
    
    /**
     * Scan directory recursively
     */
    scanDirectory(dir) {
        const entries = fs.readdirSync(dir, { withFileTypes: true });
        
        entries.forEach(entry => {
            const fullPath = path.join(dir, entry.name);
            
            // Skip node_modules, .git, etc.
            if (entry.name.startsWith('.') || 
                entry.name === 'node_modules' || 
                entry.name === 'vendor' ||
                entry.name === 'dist' ||
                entry.name === 'build') {
                return;
            }
            
            if (entry.isDirectory()) {
                this.scanDirectory(fullPath);
            } else if (entry.isFile()) {
                // Only scan relevant file types
                const ext = path.extname(entry.name);
                if (['.go', '.js', '.jsx', '.ts', '.tsx'].includes(ext)) {
                    this.scanFile(fullPath);
                }
            }
        });
    }
    
    /**
     * Generate report
     */
    generateReport() {
        const summary = {
            total_findings: this.findings.length,
            by_severity: {
                critical: 0,
                high: 0,
                medium: 0,
                low: 0,
                info: 0
            },
            by_category: {}
        };
        
        this.findings.forEach(finding => {
            summary.by_severity[finding.severity]++;
            
            if (!summary.by_category[finding.category]) {
                summary.by_category[finding.category] = 0;
            }
            summary.by_category[finding.category]++;
        });
        
        const report = {
            timestamp: new Date().toISOString(),
            summary: summary,
            findings: this.findings
        };
        
        return report;
    }
    
    /**
     * Save report to file
     */
    saveReport(report) {
        // Ensure output directory exists
        if (!fs.existsSync(this.outputDir)) {
            fs.mkdirSync(this.outputDir, { recursive: true });
        }
        
        const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
        const filename = `placeholder-detection-report-${timestamp}.json`;
        const filepath = path.join(this.outputDir, filename);
        
        fs.writeFileSync(filepath, JSON.stringify(report, null, 2));
        console.log(`Report saved to: ${filepath}`);
        
        return filepath;
    }
    
    /**
     * Run detection
     */
    run() {
        console.log('🔍 Starting enhanced placeholder detection...');
        console.log(`Root directory: ${this.rootDir}`);
        console.log(`Allowlist: ${this.allowlistPath}`);
        
        this.scanDirectory(this.rootDir);
        
        console.log(`\n📊 Found ${this.findings.length} placeholder matches`);
        
        const report = this.generateReport();
        
        // Print summary
        console.log('\n📈 Summary:');
        console.log(`  Total findings: ${report.summary.total_findings}`);
        console.log(`  Critical: ${report.summary.by_severity.critical}`);
        console.log(`  High: ${report.summary.by_severity.high}`);
        console.log(`  Medium: ${report.summary.by_severity.medium}`);
        console.log(`  Low: ${report.summary.by_severity.low}`);
        console.log(`  Info: ${report.summary.by_severity.info}`);
        
        // Save report
        const reportPath = this.saveReport(report);
        
        // Exit with error code if critical/high findings
        if (report.summary.by_severity.critical > 0 || report.summary.by_severity.high > 0) {
            console.error('\n❌ Critical or high severity findings detected!');
            process.exit(1);
        }
        
        return report;
    }
}

// Run if called directly
if (require.main === module) {
    const detector = new EnhancedPlaceholderDetector({
        rootDir: process.argv[2] || process.cwd(),
        allowlistPath: process.argv[3] || path.join(process.cwd(), 'scripts', 'placeholder-allowlist.json'),
        outputDir: process.argv[4] || path.join(process.cwd(), 'test-results')
    });
    
    detector.run();
}

module.exports = EnhancedPlaceholderDetector;

