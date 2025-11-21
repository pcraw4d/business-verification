#!/usr/bin/env node

/**
 * Build Verification Script
 * 
 * Verifies that all required environment variables are set before building.
 * This is critical for Next.js as NEXT_PUBLIC_* variables are embedded at build time.
 */

const requiredEnvVars = [
  'NEXT_PUBLIC_API_BASE_URL',
];

const optionalEnvVars = [
  'USE_NEW_UI',
  'NEXT_PUBLIC_USE_NEW_UI',
  'NODE_ENV',
];

const errors = [];
const warnings = [];

// Check required variables
console.log('🔍 Checking required environment variables...\n');

requiredEnvVars.forEach((varName) => {
  const value = process.env[varName];
  if (!value) {
    errors.push(`❌ ${varName} is not set (REQUIRED)`);
  } else if (value.includes('localhost') && process.env.NODE_ENV === 'production' && !process.env.ALLOW_LOCALHOST_FOR_TESTING) {
    errors.push(`❌ ${varName} is set to localhost in production: ${value}`);
  } else {
    console.log(`✅ ${varName} is set: ${value}`);
  }
});

// Check optional variables
console.log('\n🔍 Checking optional environment variables...\n');

optionalEnvVars.forEach((varName) => {
  const value = process.env[varName];
  if (!value) {
    warnings.push(`⚠️  ${varName} is not set (optional)`);
  } else {
    console.log(`✅ ${varName} is set: ${value}`);
  }
});

// Validate API base URL format
const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL;
if (apiBaseUrl) {
  try {
    const url = new URL(apiBaseUrl);
    if (url.protocol !== 'https:' && process.env.NODE_ENV === 'production') {
      warnings.push(`⚠️  NEXT_PUBLIC_API_BASE_URL uses ${url.protocol} instead of https: in production`);
    }
    console.log(`\n✅ API Base URL is valid: ${url.href}`);
    console.log(`   Protocol: ${url.protocol}`);
    console.log(`   Host: ${url.host}`);
  } catch (error) {
    errors.push(`❌ NEXT_PUBLIC_API_BASE_URL is not a valid URL: ${apiBaseUrl}`);
  }
}

// Print summary
console.log('\n' + '='.repeat(60));
console.log('📊 Summary');
console.log('='.repeat(60));

if (warnings.length > 0) {
  console.log('\n⚠️  Warnings:');
  warnings.forEach((warning) => console.log(`   ${warning}`));
}

if (errors.length > 0) {
  console.log('\n❌ Errors:');
  errors.forEach((error) => console.log(`   ${error}`));
  console.log('\n💡 Fix these errors before building!');
  console.log('   Set environment variables in Railway dashboard or .env file');
  process.exit(1);
}

if (warnings.length === 0 && errors.length === 0) {
  console.log('\n✅ All checks passed! Ready to build.');
} else if (errors.length === 0) {
  console.log('\n✅ All required checks passed! (Some optional variables are missing)');
}

process.exit(0);

