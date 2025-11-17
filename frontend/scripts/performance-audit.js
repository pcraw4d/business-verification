#!/usr/bin/env node

/**
 * Performance audit script
 * Measures and reports on frontend performance metrics
 */

const fs = require('fs');
const path = require('path');

const BUILD_DIR = path.join(__dirname, '..', '.next');

function analyzePerformance() {
  console.log('⚡ Performance Audit\n');
  
  if (!fs.existsSync(BUILD_DIR)) {
    console.error('❌ Build directory not found. Please run "npm run build" first.');
    process.exit(1);
  }
  
  // Check for build manifest
  const manifestPath = path.join(BUILD_DIR, 'build-manifest.json');
  if (fs.existsSync(manifestPath)) {
    const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf-8'));
    
    console.log('📋 Build Manifest Analysis:');
    console.log('─'.repeat(50));
    console.log(`Pages: ${Object.keys(manifest.pages || {}).length}`);
    console.log(`Root files: ${(manifest.rootMainFiles || []).length}`);
    console.log(`Root CSS files: ${(manifest.rootMainFiles || []).filter(f => f.endsWith('.css')).length}`);
    console.log('');
  }
  
  // Analyze route structure
  const appDir = path.join(BUILD_DIR, 'server', 'app');
  if (fs.existsSync(appDir)) {
    const routes = analyzeRoutes(appDir);
    console.log('🛣️  Route Analysis:');
    console.log('─'.repeat(50));
    console.log(`Total routes: ${routes.length}`);
    console.log(`Routes with layouts: ${routes.filter(r => r.hasLayout).length}`);
    console.log(`Routes with loading: ${routes.filter(r => r.hasLoading).length}`);
    console.log('');
  }
  
  // Check for optimizations
  console.log('✅ Optimization Checks:');
  console.log('─'.repeat(50));
  
  const nextConfigPath = path.join(__dirname, '..', 'next.config.ts');
  if (fs.existsSync(nextConfigPath)) {
    const config = fs.readFileSync(nextConfigPath, 'utf-8');
    
    const checks = [
      { name: 'Code splitting configured', pattern: /splitChunks/i },
      { name: 'Image optimization', pattern: /images/i },
      { name: 'Font optimization', pattern: /display.*swap/i },
      { name: 'Package imports optimization', pattern: /optimizePackageImports/i },
    ];
    
    checks.forEach(check => {
      const passed = check.pattern.test(config);
      console.log(`${passed ? '✅' : '❌'} ${check.name}`);
    });
  }
  
  console.log('');
  console.log('💡 Next Steps:');
  console.log('─'.repeat(50));
  console.log('1. Run Lighthouse audit: npm run lighthouse');
  console.log('2. Check bundle sizes: npm run analyze-bundle');
  console.log('3. Test load times in browser DevTools');
  console.log('4. Monitor Core Web Vitals');
}

function analyzeRoutes(dir, prefix = '') {
  const routes = [];
  try {
    const items = fs.readdirSync(dir);
    
    for (const item of items) {
      const fullPath = path.join(dir, item);
      const stat = fs.statSync(fullPath);
      
      if (stat.isDirectory()) {
        const routePath = path.join(prefix, item);
        const route = {
          path: routePath,
          hasLayout: fs.existsSync(path.join(fullPath, 'layout.js')) || fs.existsSync(path.join(fullPath, 'layout.tsx')),
          hasLoading: fs.existsSync(path.join(fullPath, 'loading.js')) || fs.existsSync(path.join(fullPath, 'loading.tsx')),
        };
        routes.push(route);
        routes.push(...analyzeRoutes(fullPath, routePath));
      }
    }
  } catch (error) {
    // Directory doesn't exist
  }
  
  return routes;
}

analyzePerformance();

