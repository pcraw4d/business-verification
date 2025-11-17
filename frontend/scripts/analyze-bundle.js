#!/usr/bin/env node

/**
 * Bundle size analysis script
 * Analyzes Next.js build output and reports bundle sizes
 */

const fs = require('fs');
const path = require('path');

const BUILD_DIR = path.join(__dirname, '..', '.next');
const STATIC_DIR = path.join(BUILD_DIR, 'static');

function formatBytes(bytes) {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

function getFileSize(filePath) {
  try {
    const stats = fs.statSync(filePath);
    return stats.size;
  } catch (error) {
    return 0;
  }
}

function analyzeDirectory(dir, prefix = '') {
  const results = [];
  try {
    const items = fs.readdirSync(dir);
    
    for (const item of items) {
      const fullPath = path.join(dir, item);
      const stat = fs.statSync(fullPath);
      
      if (stat.isDirectory()) {
        results.push(...analyzeDirectory(fullPath, path.join(prefix, item)));
      } else {
        const size = stat.size;
        const relativePath = path.join(prefix, item);
        results.push({ path: relativePath, size });
      }
    }
  } catch (error) {
    // Directory doesn't exist or can't be read
  }
  
  return results;
}

function analyzeBundle() {
  console.log('📦 Analyzing Next.js bundle sizes...\n');
  
  if (!fs.existsSync(BUILD_DIR)) {
    console.error('❌ Build directory not found. Please run "npm run build" first.');
    process.exit(1);
  }
  
  // Analyze static chunks
  const chunksDir = path.join(STATIC_DIR, 'chunks');
  const chunks = analyzeDirectory(chunksDir);
  
  // Analyze pages
  const pagesDir = path.join(STATIC_DIR, 'pages');
  const pages = analyzeDirectory(pagesDir);
  
  // Group by type
  const jsFiles = [...chunks, ...pages].filter(f => f.path.endsWith('.js'));
  const cssFiles = [...chunks, ...pages].filter(f => f.path.endsWith('.css'));
  
  // Sort by size
  jsFiles.sort((a, b) => b.size - a.size);
  cssFiles.sort((a, b) => b.size - a.size);
  
  // Calculate totals
  const totalJSSize = jsFiles.reduce((sum, f) => sum + f.size, 0);
  const totalCSSSize = cssFiles.reduce((sum, f) => sum + f.size, 0);
  const totalSize = totalJSSize + totalCSSSize;
  
  console.log('📊 Bundle Size Summary');
  console.log('═'.repeat(50));
  console.log(`Total JavaScript: ${formatBytes(totalJSSize)}`);
  console.log(`Total CSS: ${formatBytes(totalCSSSize)}`);
  console.log(`Total: ${formatBytes(totalSize)}`);
  console.log('');
  
  console.log('🔝 Top 10 Largest JavaScript Files:');
  console.log('─'.repeat(50));
  jsFiles.slice(0, 10).forEach((file, index) => {
    const percentage = ((file.size / totalJSSize) * 100).toFixed(1);
    console.log(`${(index + 1).toString().padStart(2)}. ${file.path.padEnd(40)} ${formatBytes(file.size).padStart(10)} (${percentage}%)`);
  });
  console.log('');
  
  if (cssFiles.length > 0) {
    console.log('🎨 CSS Files:');
    console.log('─'.repeat(50));
    cssFiles.forEach((file) => {
      console.log(`   ${file.path.padEnd(40)} ${formatBytes(file.size).padStart(10)}`);
    });
    console.log('');
  }
  
  // Analyze by chunk type
  // Note: Turbopack may use different naming, so we check file contents/size patterns
  const vendorChunks = jsFiles.filter(f => 
    f.path.includes('vendor') || 
    f.path.includes('framework') ||
    f.path.includes('node_modules') ||
    (f.size > 200000 && !f.path.includes('pages')) // Large chunks likely vendor code
  );
  const chartChunks = jsFiles.filter(f => 
    f.path.includes('charts') || 
    f.path.includes('recharts') || 
    f.path.includes('d3') ||
    f.path.includes('export-libs')
  );
  const appChunks = jsFiles.filter(f => 
    !f.path.includes('vendor') && 
    !f.path.includes('charts') &&
    !f.path.includes('framework') &&
    !f.path.includes('export-libs')
  );
  
  console.log('📦 Chunk Analysis:');
  console.log('─'.repeat(50));
  console.log(`Vendor chunks: ${formatBytes(vendorChunks.reduce((s, f) => s + f.size, 0))}`);
  console.log(`Chart chunks: ${formatBytes(chartChunks.reduce((s, f) => s + f.size, 0))}`);
  console.log(`App chunks: ${formatBytes(appChunks.reduce((s, f) => s + f.size, 0))}`);
  console.log('');
  
  // Recommendations
  console.log('💡 Recommendations:');
  console.log('─'.repeat(50));
  
  if (totalJSSize > 500 * 1024) { // 500KB
    console.log('⚠️  Total JS bundle is large. Consider:');
    console.log('   - Further code splitting');
    console.log('   - Tree shaking unused code');
    console.log('   - Lazy loading more components');
  }
  
  const largestChunk = jsFiles[0];
  if (largestChunk && largestChunk.size > 200 * 1024) { // 200KB
    console.log(`⚠️  Largest chunk (${largestChunk.path}) is ${formatBytes(largestChunk.size)}. Consider splitting.`);
  }
  
  if (chartChunks.length > 0) {
    const chartSize = chartChunks.reduce((s, f) => s + f.size, 0);
    if (chartSize > 300 * 1024) {
      console.log(`⚠️  Chart libraries total ${formatBytes(chartSize)}. Ensure they are lazy-loaded.`);
    }
  }
  
  console.log('');
  console.log('✅ Analysis complete!');
}

analyzeBundle();

