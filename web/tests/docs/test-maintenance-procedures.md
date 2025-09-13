# Visual Regression Test Maintenance Procedures

## Overview

This document outlines the procedures for maintaining the visual regression testing framework, including baseline updates, test maintenance, and troubleshooting.

## Table of Contents

1. [Baseline Update Procedures](#baseline-update-procedures)
2. [Test Maintenance Schedule](#test-maintenance-schedule)
3. [Test Review Process](#test-review-process)
4. [Performance Monitoring](#performance-monitoring)
5. [Documentation Updates](#documentation-updates)
6. [Emergency Procedures](#emergency-procedures)

## Baseline Update Procedures

### When to Update Baselines

Update baselines when:
- ✅ **Intentional UI changes** are made
- ✅ **New features** are added
- ✅ **Design updates** are implemented
- ✅ **Bug fixes** that change visual appearance
- ❌ **NOT for** temporary styling issues
- ❌ **NOT for** test failures due to bugs

### Baseline Update Process

#### 1. Local Development

```bash
# 1. Make your UI changes
# 2. Test locally to ensure changes work
npm run test:visual

# 3. Update baselines for intentional changes
npx playwright test --update-snapshots

# 4. Review the updated screenshots
npx playwright show-report

# 5. Commit the changes
git add web/tests/visual/*.spec.js-snapshots/
git commit -m "chore: update visual regression test baselines for [feature description]"
git push
```

#### 2. Using GitHub Actions

```bash
# 1. Go to GitHub Actions
# 2. Run "Update Visual Test Baselines" workflow
# 3. Select the test type to update:
#    - all: Update all test baselines
#    - baseline: Update baseline screenshots only
#    - interactive: Update interactive element tests
#    - state-based: Update state-based tests
#    - cross-browser: Update cross-browser tests
#    - responsive: Update responsive design tests
# 4. Review the workflow results
# 5. Check the committed changes
```

#### 3. Selective Baseline Updates

```bash
# Update specific test baselines
npx playwright test tests/visual/baseline-screenshots.spec.js --update-snapshots
npx playwright test tests/visual/interactive-element-tests.spec.js --update-snapshots
npx playwright test tests/visual/state-based-tests.spec.js --update-snapshots
npx playwright test tests/visual/cross-browser.spec.js --update-snapshots
npx playwright test tests/visual/responsive-design.spec.js --update-snapshots
```

### Baseline Review Checklist

Before committing baseline updates:

- [ ] **Visual Review**: All updated screenshots look correct
- [ ] **Intentional Changes**: Changes are intentional, not bugs
- [ ] **Consistency**: Screenshots are consistent across browsers
- [ ] **Quality**: Screenshots are clear and properly cropped
- [ ] **Documentation**: Changes are documented in commit message
- [ ] **Testing**: Updated baselines pass all tests

### Baseline File Management

#### File Structure
```
web/tests/visual/
├── baseline-screenshots.spec.js-snapshots/
│   ├── risk-dashboard-page-full-darwin.png
│   ├── enhanced-risk-indicators-page-full-darwin.png
│   └── ...
├── interactive-element-tests.spec.js-snapshots/
│   ├── interactive-button-hover-primary-darwin.png
│   ├── interactive-tooltip-hover-darwin.png
│   └── ...
└── ...
```

#### File Naming Convention
- Format: `{test-name}-{platform}.png`
- Platform: `darwin` (macOS), `linux` (Linux), `win32` (Windows)
- Examples:
  - `risk-dashboard-page-full-darwin.png`
  - `interactive-button-hover-primary-darwin.png`

#### File Size Management
- Monitor baseline file sizes
- Compress large screenshots if needed
- Remove obsolete baseline files
- Keep file sizes under 1MB when possible

## Test Maintenance Schedule

### Daily Maintenance

- [ ] **Monitor Test Results**: Check GitHub Actions for test failures
- [ ] **Review Failures**: Investigate any test failures
- [ ] **Update Documentation**: Document any new issues or solutions

### Weekly Maintenance

- [ ] **Test Stability Review**: Check test success rates
- [ ] **Performance Monitoring**: Review test execution times
- [ ] **Baseline Review**: Check for outdated baselines
- [ ] **Dependency Updates**: Check for Playwright updates

### Monthly Maintenance

- [ ] **Comprehensive Review**: Review all test files
- [ ] **Performance Optimization**: Optimize slow tests
- [ ] **Documentation Updates**: Update guides and procedures
- [ ] **Test Coverage Analysis**: Ensure adequate coverage

### Quarterly Maintenance

- [ ] **Framework Updates**: Update Playwright and dependencies
- [ ] **Test Architecture Review**: Review test structure and organization
- [ ] **Best Practices Review**: Update testing best practices
- [ ] **Training Updates**: Update team training materials

## Test Review Process

### Code Review Checklist

When reviewing test changes:

- [ ] **Test Logic**: Test logic is correct and comprehensive
- [ ] **Selectors**: Selectors are stable and semantic
- [ ] **Waits**: Proper waits are implemented
- [ ] **Error Handling**: Error scenarios are handled
- [ ] **Documentation**: Tests are well-documented
- [ ] **Performance**: Tests are optimized for performance

### Test Quality Metrics

Monitor these metrics:

- **Success Rate**: > 95% test success rate
- **Execution Time**: < 10 minutes for full test suite
- **Flakiness**: < 5% flaky test rate
- **Coverage**: > 90% visual coverage of UI components

### Test Review Process

1. **Automated Review**: GitHub Actions runs tests automatically
2. **Manual Review**: Review test results and screenshots
3. **Peer Review**: Team member reviews test changes
4. **Approval**: Approve changes after review
5. **Merge**: Merge changes to main branch

## Performance Monitoring

### Key Performance Indicators

- **Test Execution Time**: Monitor total test execution time
- **Individual Test Time**: Monitor slow tests
- **Resource Usage**: Monitor CPU and memory usage
- **Network Performance**: Monitor Railway connection times

### Performance Optimization

#### 1. Parallel Execution
```javascript
// In playwright.config.js
export default defineConfig({
  workers: 4, // Adjust based on available resources
});
```

#### 2. Test Optimization
```javascript
// Use efficient selectors
await page.click('[data-testid="submit-button"]'); // Good
await page.click('.container > div:nth-child(2) > button'); // Bad

// Use proper waits
await page.waitForSelector('.element', { state: 'visible' }); // Good
await page.waitForTimeout(5000); // Bad
```

#### 3. Screenshot Optimization
```javascript
// Optimize screenshot settings
await expect(page).toHaveScreenshot('my-page.png', {
  maxDiffPixels: 100, // Allow small differences
  threshold: 0.2, // Set appropriate threshold
});
```

### Performance Monitoring Tools

- **GitHub Actions**: Monitor test execution times
- **Playwright Reports**: Review detailed performance metrics
- **Railway Metrics**: Monitor application performance
- **Custom Scripts**: Monitor specific performance aspects

## Documentation Updates

### When to Update Documentation

Update documentation when:
- New test types are added
- Test procedures change
- New issues are discovered
- Best practices are updated
- Tools or frameworks are updated

### Documentation Maintenance

#### 1. Test Documentation
- Update test descriptions
- Add new test examples
- Update troubleshooting guides
- Maintain API documentation

#### 2. Procedure Documentation
- Update maintenance procedures
- Add new troubleshooting steps
- Update best practices
- Maintain configuration guides

#### 3. User Documentation
- Update user guides
- Add new feature documentation
- Update FAQ sections
- Maintain training materials

### Documentation Review Process

1. **Regular Review**: Review documentation monthly
2. **User Feedback**: Incorporate user feedback
3. **Expert Review**: Have experts review technical content
4. **Version Control**: Track documentation changes
5. **Publication**: Publish updated documentation

## Emergency Procedures

### Test Failure Emergency

When tests fail in production:

1. **Immediate Response**:
   - Check Railway deployment status
   - Verify application accessibility
   - Review error logs

2. **Investigation**:
   - Identify root cause
   - Check recent changes
   - Review test logs

3. **Resolution**:
   - Fix the issue
   - Update tests if needed
   - Update baselines if necessary

4. **Prevention**:
   - Update procedures
   - Improve monitoring
   - Add safeguards

### Baseline Corruption

When baselines are corrupted:

1. **Identification**:
   - Check baseline file integrity
   - Verify file permissions
   - Review recent changes

2. **Recovery**:
   - Restore from backup
   - Regenerate baselines
   - Update from main branch

3. **Prevention**:
   - Improve backup procedures
   - Add integrity checks
   - Monitor file changes

### Performance Degradation

When test performance degrades:

1. **Monitoring**:
   - Check execution times
   - Monitor resource usage
   - Review test logs

2. **Optimization**:
   - Optimize slow tests
   - Reduce parallel execution
   - Update test configuration

3. **Scaling**:
   - Add more resources
   - Optimize test distribution
   - Update infrastructure

## Maintenance Tools

### Automated Tools

- **GitHub Actions**: Automated test execution
- **Playwright**: Test execution and reporting
- **Custom Scripts**: Maintenance automation
- **Monitoring Tools**: Performance monitoring

### Manual Tools

- **Test Reports**: Review test results
- **Screenshot Comparison**: Visual diff analysis
- **Log Analysis**: Error investigation
- **Performance Profiling**: Performance analysis

### Maintenance Scripts

```bash
# Clean up old test results
npm run test:clean

# Update all dependencies
npm update

# Run performance benchmarks
npm run test:performance

# Generate maintenance report
npm run test:maintenance-report
```

## Best Practices

### 1. Regular Maintenance
- Schedule regular maintenance windows
- Monitor test health continuously
- Update documentation regularly
- Review and optimize tests

### 2. Change Management
- Document all changes
- Review changes before implementation
- Test changes thoroughly
- Monitor impact of changes

### 3. Quality Assurance
- Maintain high test quality standards
- Review test coverage regularly
- Monitor test stability
- Ensure proper error handling

### 4. Team Collaboration
- Share knowledge and best practices
- Conduct regular training sessions
- Maintain clear communication
- Document lessons learned

### 5. Continuous Improvement
- Regularly review and improve processes
- Incorporate feedback and lessons learned
- Stay updated with best practices
- Invest in tooling and automation

## Conclusion

Proper maintenance of the visual regression testing framework is essential for its long-term success. By following these procedures, maintaining regular schedules, and continuously improving the framework, we can ensure reliable and effective visual testing for the KYB Platform.

Remember to:
- Update baselines only for intentional changes
- Monitor test performance regularly
- Maintain comprehensive documentation
- Follow established procedures
- Continuously improve the framework
