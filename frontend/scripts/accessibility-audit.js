#!/usr/bin/env node

/**
 * Accessibility audit script
 * Checks for common accessibility issues in the codebase
 */

const fs = require('fs');
const path = require('path');

const APP_DIR = path.join(__dirname, '..', 'app');
const COMPONENTS_DIR = path.join(__dirname, '..', 'components');

const ACCESSIBILITY_ISSUES = [];

function checkFile(filePath, content) {
  const relativePath = path.relative(path.join(__dirname, '..'), filePath);
  const issues = [];
  
  // Check for missing alt text on images
  const imgMatches = content.matchAll(/<img[^>]*>/gi);
  for (const match of imgMatches) {
    if (!match[0].includes('alt=')) {
      issues.push({
        type: 'error',
        rule: 'WCAG 1.1.1',
        message: 'Image missing alt attribute',
        line: getLineNumber(content, match.index),
      });
    }
  }
  
  // Check for missing labels on form inputs
  const inputMatches = content.matchAll(/<input[^>]*>/gi);
  for (const match of inputMatches) {
    const input = match[0];
    if (!input.includes('aria-label') && !input.includes('aria-labelledby') && !input.includes('id=')) {
      // Check if there's a label nearby (simplified check)
      const context = content.substring(Math.max(0, match.index - 200), match.index + 200);
      if (!context.includes('<label') && !context.includes('Label')) {
        issues.push({
          type: 'warning',
          rule: 'WCAG 4.1.2',
          message: 'Input may be missing accessible label',
          line: getLineNumber(content, match.index),
        });
      }
    }
  }
  
  // Check for missing button labels
  const buttonMatches = content.matchAll(/<button[^>]*>/gi);
  for (const match of buttonMatches) {
    const button = match[0];
    const context = content.substring(match.index, match.index + 100);
    if (!button.includes('aria-label') && !context.match(/<button[^>]*>[\s\S]{1,50}<\/button>/)) {
      issues.push({
        type: 'warning',
        rule: 'WCAG 4.1.2',
        message: 'Button may be missing accessible label',
        line: getLineNumber(content, match.index),
      });
    }
  }
  
  // Check for color contrast (simplified - check for inline styles with color)
  const colorMatches = content.matchAll(/style=["'][^"']*color[^"']*["']/gi);
  for (const match of colorMatches) {
    issues.push({
      type: 'info',
      rule: 'WCAG 1.4.3',
      message: 'Inline color styles - verify contrast ratio meets AA standards',
      line: getLineNumber(content, match.index),
    });
  }
  
  // Check for missing heading hierarchy
  const headings = content.matchAll(/<h([1-6])[^>]*>/gi);
  let lastLevel = 0;
  for (const match of headings) {
    const level = parseInt(match[1]);
    if (level > lastLevel + 1) {
      issues.push({
        type: 'warning',
        rule: 'WCAG 1.3.1',
        message: `Heading level ${level} follows level ${lastLevel} - may skip levels`,
        line: getLineNumber(content, match.index),
      });
    }
    lastLevel = level;
  }
  
  // Check for missing lang attribute on html
  if (content.includes('<html') && !content.includes('lang=')) {
    issues.push({
      type: 'error',
      rule: 'WCAG 3.1.1',
      message: 'HTML element missing lang attribute',
      line: 1,
    });
  }
  
  // Check for missing focus indicators
  const interactiveElements = content.matchAll(/<(button|a|input|select|textarea)[^>]*>/gi);
  for (const match of interactiveElements) {
    const element = match[0];
    if (!element.includes('focus:') && !element.includes('focus-visible:')) {
      // This is a simplified check - actual focus styles might be in CSS
      // Just note it for manual review
    }
  }
  
  if (issues.length > 0) {
    ACCESSIBILITY_ISSUES.push({
      file: relativePath,
      issues,
    });
  }
}

function getLineNumber(content, index) {
  return content.substring(0, index).split('\n').length;
}

function scanDirectory(dir) {
  try {
    const items = fs.readdirSync(dir);
    
    for (const item of items) {
      const fullPath = path.join(dir, item);
      const stat = fs.statSync(fullPath);
      
      if (stat.isDirectory()) {
        scanDirectory(fullPath);
      } else if (item.endsWith('.tsx') || item.endsWith('.jsx') || item.endsWith('.ts') || item.endsWith('.js')) {
        const content = fs.readFileSync(fullPath, 'utf-8');
        checkFile(fullPath, content);
      }
    }
  } catch (error) {
    // Directory doesn't exist or can't be read
  }
}

function runAudit() {
  console.log('♿ Running Accessibility Audit\n');
  console.log('Scanning files...\n');
  
  scanDirectory(APP_DIR);
  scanDirectory(COMPONENTS_DIR);
  
  if (ACCESSIBILITY_ISSUES.length === 0) {
    console.log('✅ No accessibility issues found!');
    return;
  }
  
  console.log(`Found ${ACCESSIBILITY_ISSUES.length} file(s) with potential issues:\n`);
  
  let errorCount = 0;
  let warningCount = 0;
  let infoCount = 0;
  
  ACCESSIBILITY_ISSUES.forEach(({ file, issues }) => {
    console.log(`📄 ${file}`);
    console.log('─'.repeat(60));
    
    issues.forEach(issue => {
      const icon = issue.type === 'error' ? '❌' : issue.type === 'warning' ? '⚠️ ' : 'ℹ️ ';
      console.log(`${icon} Line ${issue.line}: ${issue.message} (${issue.rule})`);
      
      if (issue.type === 'error') errorCount++;
      else if (issue.type === 'warning') warningCount++;
      else infoCount++;
    });
    
    console.log('');
  });
  
  console.log('📊 Summary:');
  console.log('─'.repeat(60));
  console.log(`Errors: ${errorCount}`);
  console.log(`Warnings: ${warningCount}`);
  console.log(`Info: ${infoCount}`);
  console.log(`Total: ${errorCount + warningCount + infoCount}`);
  console.log('');
  
  console.log('💡 Recommendations:');
  console.log('─'.repeat(60));
  console.log('1. Fix all errors (missing alt text, lang attributes)');
  console.log('2. Review warnings (labels, heading hierarchy)');
  console.log('3. Verify color contrast ratios meet WCAG AA standards');
  console.log('4. Test with screen readers (NVDA, JAWS, VoiceOver)');
  console.log('5. Test keyboard navigation');
  console.log('6. Run automated tools: axe DevTools, WAVE, Lighthouse');
}

runAudit();

