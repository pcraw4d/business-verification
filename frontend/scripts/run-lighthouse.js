#!/usr/bin/env node

/**
 * Run Lighthouse audit and generate report
 * Usage: node scripts/run-lighthouse.js [url]
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const url = process.argv[2] || 'http://localhost:3000';
const outputDir = path.join(__dirname, '..', 'lighthouse-reports');

// Create output directory if it doesn't exist
if (!fs.existsSync(outputDir)) {
  fs.mkdirSync(outputDir, { recursive: true });
}

console.log('🔍 Running Lighthouse audit...\n');
console.log(`URL: ${url}\n`);

try {
  // Run Lighthouse
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  const htmlReport = path.join(outputDir, `lighthouse-report-${timestamp}.html`);
  const jsonReport = path.join(outputDir, `lighthouse-report-${timestamp}.json`);

  console.log('Running Lighthouse (this may take a minute)...\n');

  execSync(
    `npx lighthouse "${url}" ` +
    `--output html --output json ` +
    `--output-path "${htmlReport}" ` +
    `--chrome-flags="--headless" ` +
    `--only-categories=performance,accessibility,best-practices,seo`,
    { stdio: 'inherit' }
  );

  // Also generate JSON report
  execSync(
    `npx lighthouse "${url}" ` +
    `--output json ` +
    `--output-path "${jsonReport}" ` +
    `--chrome-flags="--headless"`,
    { stdio: 'pipe' }
  );

  // Parse and display summary
  if (fs.existsSync(jsonReport)) {
    const report = JSON.parse(fs.readFileSync(jsonReport, 'utf-8'));
    const categories = report.categories;

    console.log('\n📊 Lighthouse Results Summary:');
    console.log('─'.repeat(60));
    console.log(`Performance:      ${Math.round(categories.performance.score * 100)}/100`);
    console.log(`Accessibility:    ${Math.round(categories.accessibility.score * 100)}/100`);
    console.log(`Best Practices:   ${Math.round(categories['best-practices'].score * 100)}/100`);
    console.log(`SEO:              ${Math.round(categories.seo.score * 100)}/100`);
    console.log('─'.repeat(60));

    // Key metrics
    const metrics = report.audits;
    console.log('\n🎯 Key Metrics:');
    console.log(`First Contentful Paint:     ${metrics['first-contentful-paint'].displayValue || 'N/A'}`);
    console.log(`Largest Contentful Paint:   ${metrics['largest-contentful-paint'].displayValue || 'N/A'}`);
    console.log(`Total Blocking Time:        ${metrics['total-blocking-time'].displayValue || 'N/A'}`);
    console.log(`Cumulative Layout Shift:    ${metrics['cumulative-layout-shift'].displayValue || 'N/A'}`);
    console.log(`Speed Index:                 ${metrics['speed-index'].displayValue || 'N/A'}`);

    console.log('\n📄 Reports generated:');
    console.log(`HTML: ${htmlReport}`);
    console.log(`JSON: ${jsonReport}`);
    console.log('\n✅ Lighthouse audit complete!');
  }
} catch (error) {
  console.error('\n❌ Error running Lighthouse:', error.message);
  console.log('\n💡 Make sure:');
  console.log('1. The dev server is running (npm run dev)');
  console.log('2. The URL is accessible');
  console.log('3. Lighthouse is installed (npm install -D lighthouse)');
  process.exit(1);
}

