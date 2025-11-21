#!/usr/bin/env node

/**
 * Quick Build Test - Verifies production build completes successfully
 * This is a faster alternative to the full hydration test suite
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

console.log('\n🔨 Quick Production Build Test\n');

try {
  // Set environment variables
  const env = {
    ...process.env,
    NODE_ENV: 'production',
    NEXT_PUBLIC_API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080',
    ALLOW_LOCALHOST_FOR_TESTING: 'true',
  };

  console.log('Building production bundle...\n');
  
  execSync('npm run build', {
    stdio: 'inherit',
    env,
    cwd: path.join(__dirname, '..'),
  });

  // Check if .next directory exists
  const nextDir = path.join(__dirname, '..', '.next');
  if (fs.existsSync(nextDir)) {
    console.log('\n✅ Production build completed successfully!');
    console.log('✅ Build output exists in .next directory');
    console.log('\nNext steps:');
    console.log('1. Run: npm run start (to test production server)');
    console.log('2. Run: npm run test:hydration:manual (to test hydration)');
    process.exit(0);
  } else {
    console.error('\n❌ Build completed but .next directory not found');
    process.exit(1);
  }
} catch (error) {
  console.error('\n❌ Build failed:', error.message);
  process.exit(1);
}

