#!/usr/bin/env node

/**
 * Code Analysis Tool - Find Unused Features
 * Analyzes the codebase to identify:
 * 1. Backend API endpoints not used in frontend
 * 2. Frontend components not rendered
 * 3. Data services not called
 * 4. Unused utility functions
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const colors = {
    reset: '\x1b[0m',
    red: '\x1b[31m',
    green: '\x1b[32m',
    yellow: '\x1b[33m',
    blue: '\x1b[34m',
    cyan: '\x1b[36m'
};

const results = {
    unusedEndpoints: [],
    unusedComponents: [],
    unusedServices: [],
    unusedUtilities: [],
    recommendations: []
};

// Directories to analyze
const dirs = {
    backend: ['services', 'internal', 'cmd'],
    frontend: ['web', 'services/frontend/public', 'cmd/frontend-service/static'],
    shared: ['web/shared']
};

// Find all Go files
function findGoFiles(dir) {
    const files = [];
    function walk(currentPath) {
        const entries = fs.readdirSync(currentPath, { withFileTypes: true });
        for (const entry of entries) {
            const fullPath = path.join(currentPath, entry.name);
            if (entry.isDirectory() && !entry.name.startsWith('.') && entry.name !== 'node_modules') {
                walk(fullPath);
            } else if (entry.isFile() && entry.name.endsWith('.go') && !entry.name.endsWith('_test.go')) {
                files.push(fullPath);
            }
        }
    }
    try {
        walk(dir);
    } catch (e) {
        // Ignore errors
    }
    return files;
}

// Find all JavaScript files
function findJSFiles(dir) {
    const files = [];
    function walk(currentPath) {
        try {
            const entries = fs.readdirSync(currentPath, { withFileTypes: true });
            for (const entry of entries) {
                const fullPath = path.join(currentPath, entry.name);
                if (entry.isDirectory() && !entry.name.startsWith('.') && entry.name !== 'node_modules') {
                    walk(fullPath);
                } else if (entry.isFile() && (entry.name.endsWith('.js') || entry.name.endsWith('.html'))) {
                    files.push(fullPath);
                }
            }
        } catch (e) {
            // Ignore errors
        }
    }
    try {
        walk(dir);
    } catch (e) {
        // Ignore errors
    }
    return files;
}

// Extract API endpoints from Go handlers
function extractEndpoints(goFiles) {
    const endpoints = new Set();
    const endpointPattern = /HandleFunc\(["']([^"']+)["']/g;
    const routePattern = /router\.(HandleFunc|Handle)\(["']([^"']+)["']/g;
    
    for (const file of goFiles) {
        try {
            const content = fs.readFileSync(file, 'utf8');
            let match;
            
            // Find HandleFunc patterns
            while ((match = endpointPattern.exec(content)) !== null) {
                endpoints.add(match[1]);
            }
            
            // Find router patterns
            while ((match = routePattern.exec(content)) !== null) {
                endpoints.add(match[2]);
            }
        } catch (e) {
            // Ignore errors
        }
    }
    
    return Array.from(endpoints);
}

// Extract API calls from JavaScript files
function extractAPICalls(jsFiles) {
    const apiCalls = new Set();
    const patterns = [
        /fetch\(["']([^"']+)["']/g,
        /\.get\(["']([^"']+)["']/g,
        /\.post\(["']([^"']+)["']/g,
        /endpoints\.(\w+)/g,
        /getEndpoints\(\)\.(\w+)/g
    ];
    
    for (const file of jsFiles) {
        try {
            const content = fs.readFileSync(file, 'utf8');
            
            for (const pattern of patterns) {
                let match;
                while ((match = pattern.exec(content)) !== null) {
                    if (match[1]) {
                        apiCalls.add(match[1]);
                    }
                }
            }
        } catch (e) {
            // Ignore errors
        }
    }
    
    return Array.from(apiCalls);
}

// Extract component definitions
function extractComponents(jsFiles) {
    const components = new Set();
    const patterns = [
        /class\s+(\w+)\s+extends/gi,
        /class\s+(\w+)\s*\{/gi,
        /function\s+(\w+)\s*\(/gi,
        /const\s+(\w+)\s*=\s*(?:class|function)/gi
    ];
    
    for (const file of jsFiles) {
        try {
            const content = fs.readFileSync(file, 'utf8');
            
            for (const pattern of patterns) {
                let match;
                while ((match = pattern.exec(content)) !== null) {
                    if (match[1] && !match[1].startsWith('_')) {
                        components.add(match[1]);
                    }
                }
            }
        } catch (e) {
            // Ignore errors
        }
    }
    
    return Array.from(components);
}

// Extract component usage
function extractComponentUsage(jsFiles, htmlFiles) {
    const usages = new Set();
    
    // Check JavaScript files
    for (const file of jsFiles) {
        try {
            const content = fs.readFileSync(file, 'utf8');
            const usagePattern = /new\s+(\w+)\s*\(/g;
            let match;
            while ((match = usagePattern.exec(content)) !== null) {
                usages.add(match[1]);
            }
        } catch (e) {
            // Ignore errors
        }
    }
    
    // Check HTML files
    for (const file of htmlFiles) {
        try {
            const content = fs.readFileSync(file, 'utf8');
            const scriptPattern = /<script[^>]*>([\s\S]*?)<\/script>/gi;
            let scriptMatch;
            while ((scriptMatch = scriptPattern.exec(content)) !== null) {
                const scriptContent = scriptMatch[1];
                const usagePattern = /new\s+(\w+)\s*\(/g;
                let match;
                while ((match = usagePattern.exec(scriptContent)) !== null) {
                    usages.add(match[1]);
                }
            }
        } catch (e) {
            // Ignore errors
        }
    }
    
    return usages; // Return Set, not Array
}

// Main analysis
console.log(`${colors.cyan}🔍 Analyzing Codebase for Unused Features...${colors.reset}\n`);

// Find all files
console.log(`${colors.blue}Scanning files...${colors.reset}`);
const goFiles = [];
for (const dir of dirs.backend) {
    const fullPath = path.join(process.cwd(), dir);
    if (fs.existsSync(fullPath)) {
        goFiles.push(...findGoFiles(fullPath));
    }
}

const jsFiles = [];
const htmlFiles = [];
for (const dir of dirs.frontend) {
    const fullPath = path.join(process.cwd(), dir);
    if (fs.existsSync(fullPath)) {
        const files = findJSFiles(fullPath);
        jsFiles.push(...files.filter(f => f.endsWith('.js')));
        htmlFiles.push(...files.filter(f => f.endsWith('.html')));
    }
}

console.log(`Found ${goFiles.length} Go files, ${jsFiles.length} JS files, ${htmlFiles.length} HTML files\n`);

// Extract endpoints
console.log(`${colors.blue}Extracting API endpoints...${colors.reset}`);
const endpoints = extractEndpoints(goFiles);
console.log(`Found ${endpoints.length} API endpoints`);

// Extract API calls
console.log(`${colors.blue}Extracting API calls from frontend...${colors.reset}`);
const apiCalls = extractAPICalls(jsFiles);
console.log(`Found ${apiCalls.size} API call patterns`);

// Find unused endpoints
for (const endpoint of endpoints) {
    let found = false;
    for (const call of apiCalls) {
        if (call.includes(endpoint) || endpoint.includes(call)) {
            found = true;
            break;
        }
    }
    if (!found && !endpoint.includes('health') && !endpoint.includes('test')) {
        results.unusedEndpoints.push(endpoint);
    }
}

// Extract components
console.log(`${colors.blue}Extracting components...${colors.reset}`);
const components = extractComponents(jsFiles);
console.log(`Found ${components.size} components`);

// Extract component usage
const componentUsages = extractComponentUsage(jsFiles, htmlFiles);
console.log(`Found ${componentUsages.size} component usages`);

// Find unused components
for (const component of components) {
    if (!componentUsages.has(component) && 
        !component.includes('Test') && 
        !component.includes('test') &&
        component.length > 3) {
        results.unusedComponents.push(component);
    }
}

// Generate report
console.log(`\n${colors.yellow}📊 Analysis Results${colors.reset}\n`);

if (results.unusedEndpoints.length > 0) {
    console.log(`${colors.red}Unused API Endpoints (${results.unusedEndpoints.length}):${colors.reset}`);
    results.unusedEndpoints.forEach(ep => console.log(`  - ${ep}`));
    console.log();
    results.recommendations.push(`Consider implementing UI for ${results.unusedEndpoints.length} unused API endpoints`);
}

if (results.unusedComponents.length > 0) {
    console.log(`${colors.yellow}Potentially Unused Components (${results.unusedComponents.length}):${colors.reset}`);
    results.unusedComponents.slice(0, 20).forEach(comp => console.log(`  - ${comp}`));
    if (results.unusedComponents.length > 20) {
        console.log(`  ... and ${results.unusedComponents.length - 20} more`);
    }
    console.log();
}

// Save report
const reportPath = path.join(process.cwd(), 'test-results', 'unused-features-analysis.json');
const reportDir = path.dirname(reportPath);
if (!fs.existsSync(reportDir)) {
    fs.mkdirSync(reportDir, { recursive: true });
}

const report = {
    timestamp: new Date().toISOString(),
    summary: {
        unusedEndpoints: results.unusedEndpoints.length,
        unusedComponents: results.unusedComponents.length
    },
    unusedEndpoints: results.unusedEndpoints,
    unusedComponents: results.unusedComponents,
    recommendations: results.recommendations
};

fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
console.log(`${colors.green}✅ Report saved to: ${reportPath}${colors.reset}\n`);

// Exit
process.exit(results.unusedEndpoints.length > 0 ? 1 : 0);

