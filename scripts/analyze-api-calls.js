#!/usr/bin/env node

/**
 * API Call Analysis Script
 * Analyzes JavaScript files to identify redundant API calls
 */

const fs = require('fs');
const path = require('path');

const STATIC_DIR = path.join(__dirname, '../cmd/frontend-service/static');
const API_PATTERNS = [
    /fetch\s*\(/g,
    /axios\.(get|post|put|delete|patch)/g,
    /\.get\(/g,
    /\.post\(/g,
    /XMLHttpRequest/g,
    /api\s*\./g,
    /\/api\//g
];

const results = {
    totalFiles: 0,
    totalCalls: 0,
    files: [],
    duplicates: [],
    recommendations: []
};

function findApiCalls(filePath, content) {
    const calls = [];
    const lines = content.split('\n');
    
    API_PATTERNS.forEach((pattern, index) => {
        let match;
        while ((match = pattern.exec(content)) !== null) {
            const lineNumber = content.substring(0, match.index).split('\n').length;
            const line = lines[lineNumber - 1] || '';
            
            calls.push({
                pattern: pattern.source,
                line: lineNumber,
                context: line.trim().substring(0, 100),
                match: match[0]
            });
        }
    });
    
    return calls;
}

function analyzeFile(filePath) {
    try {
        const content = fs.readFileSync(filePath, 'utf8');
        const calls = findApiCalls(filePath, content);
        
        if (calls.length > 0) {
            results.files.push({
                file: path.relative(STATIC_DIR, filePath),
                calls: calls.length,
                details: calls
            });
            results.totalCalls += calls.length;
        }
        
        return calls;
    } catch (error) {
        console.error(`Error reading ${filePath}:`, error.message);
        return [];
    }
}

function findDuplicates() {
    const urlMap = new Map();
    
    results.files.forEach(file => {
        file.details.forEach(call => {
            // Extract URL from context
            const urlMatch = call.context.match(/['"`]([^'"`]+)['"`]/);
            if (urlMatch) {
                const url = urlMatch[1];
                if (!urlMap.has(url)) {
                    urlMap.set(url, []);
                }
                urlMap.get(url).push({
                    file: file.file,
                    line: call.line
                });
            }
        });
    });
    
    urlMap.forEach((locations, url) => {
        if (locations.length > 1) {
            results.duplicates.push({
                url: url,
                count: locations.length,
                locations: locations
            });
        }
    });
}

function generateRecommendations() {
    // Group by file
    const fileGroups = {};
    results.duplicates.forEach(dup => {
        dup.locations.forEach(loc => {
            if (!fileGroups[loc.file]) {
                fileGroups[loc.file] = [];
            }
            fileGroups[loc.file].push(dup.url);
        });
    });
    
    Object.keys(fileGroups).forEach(file => {
        const urls = [...new Set(fileGroups[file])];
        if (urls.length > 0) {
            results.recommendations.push({
                file: file,
                suggestion: `Consider creating a shared API client for: ${urls.slice(0, 3).join(', ')}${urls.length > 3 ? '...' : ''}`,
                duplicateUrls: urls.length
            });
        }
    });
}

function scanDirectory(dir) {
    const entries = fs.readdirSync(dir, { withFileTypes: true });
    
    entries.forEach(entry => {
        const fullPath = path.join(dir, entry.name);
        
        // Skip node_modules, dist, and test files
        if (entry.name === 'node_modules' || 
            entry.name === 'dist' || 
            entry.name === '.build-temp' ||
            entry.name.endsWith('.test.js')) {
            return;
        }
        
        if (entry.isDirectory()) {
            scanDirectory(fullPath);
        } else if (entry.isFile() && entry.name.endsWith('.js')) {
            results.totalFiles++;
            analyzeFile(fullPath);
        }
    });
}

// Main execution
console.log('🔍 Analyzing API calls in frontend files...\n');

if (!fs.existsSync(STATIC_DIR)) {
    console.error(`❌ Static directory not found: ${STATIC_DIR}`);
    process.exit(1);
}

scanDirectory(STATIC_DIR);
findDuplicates();
generateRecommendations();

// Output results
console.log('==========================================');
console.log('API Call Analysis Results');
console.log('==========================================\n');

console.log(`📊 Statistics:`);
console.log(`   Total JS files scanned: ${results.totalFiles}`);
console.log(`   Files with API calls: ${results.files.length}`);
console.log(`   Total API calls found: ${results.totalCalls}`);
console.log(`   Duplicate API calls: ${results.duplicates.length}\n`);

if (results.duplicates.length > 0) {
    console.log('⚠️  Duplicate API Calls:');
    results.duplicates.slice(0, 10).forEach(dup => {
        console.log(`\n   URL: ${dup.url}`);
        console.log(`   Found ${dup.count} times:`);
        dup.locations.forEach(loc => {
            console.log(`     - ${loc.file}:${loc.line}`);
        });
    });
    
    if (results.duplicates.length > 10) {
        console.log(`\n   ... and ${results.duplicates.length - 10} more duplicates`);
    }
}

if (results.recommendations.length > 0) {
    console.log('\n💡 Recommendations:');
    results.recommendations.slice(0, 10).forEach(rec => {
        console.log(`\n   ${rec.file}:`);
        console.log(`   ${rec.suggestion}`);
        console.log(`   (${rec.duplicateUrls} duplicate URLs)`);
    });
}

// Save results to file
const outputFile = path.join(__dirname, '../Beta readiness/API_CALL_ANALYSIS.json');
fs.writeFileSync(outputFile, JSON.stringify(results, null, 2));
console.log(`\n📄 Full results saved to: ${outputFile}`);

console.log('\n✅ Analysis complete!');

