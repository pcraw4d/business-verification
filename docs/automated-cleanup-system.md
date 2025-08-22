# Automated Cleanup System Documentation

## Overview

The KYB Platform Automated Cleanup System is a comprehensive solution for maintaining code quality and reducing technical debt through automated detection and removal of deprecated code, backup files, and other code smells.

## System Architecture

### Core Components

1. **Go Cleanup Tool** (`cmd/cleanup/main.go`)
   - Advanced code analysis engine
   - Pattern-based detection of deprecated code
   - Auto-fix capabilities for safe operations
   - Integration with technical debt monitoring

2. **Shell Scripts**
   - `scripts/cleanup-deprecated-code.sh`: Command-line interface
   - `scripts/run-cleanup.sh`: Comprehensive runner
   - `scripts/setup-scheduled-cleanup.sh`: Cron job management

3. **Configuration System** (`configs/cleanup-config.yaml`)
   - Flexible pattern definitions
   - Severity classifications
   - Auto-fix settings
   - Integration configurations

4. **Monitoring & Reporting**
   - Technical debt metrics integration
   - Prometheus metrics export
   - HTML/JSON/YAML report generation
   - Health monitoring scripts

## Features

### Code Detection Patterns

- **Deprecated Comments**: `// DEPRECATED`, `// TODO: remove`, `// FIXME: legacy`
- **Deprecated Functions**: Functions marked as deprecated
- **Problematic Imports**: Imports of deprecated or problematic packages
- **Backup Files**: `.bak`, `.backup`, `.old` files
- **Temporary Files**: `.tmp`, `.temp`, `.cache` files
- **Magic Numbers**: Hardcoded numeric values without context
- **Unused Imports**: Potentially unused import statements
- **Empty Catch Blocks**: Empty error handling blocks

### Severity Levels

- **Critical**: Immediate attention required (e.g., problematic imports)
- **High**: Important issues (e.g., deprecated functions)
- **Medium**: Moderate issues (e.g., deprecated comments)
- **Low**: Minor issues (e.g., magic numbers)

### Auto-Fix Capabilities

- **Safe Operations**: Backup files, temporary files
- **Manual Review Required**: Deprecated functions, comments
- **Never Auto-Fix**: Critical security issues

## Usage

### Manual Cleanup

```bash
# Run comprehensive cleanup
./scripts/run-cleanup.sh

# Target specific patterns
./scripts/run-cleanup.sh --pattern backup_files --auto-fix

# Dry run (preview changes)
./scripts/run-cleanup.sh --dry-run

# Generate HTML report
./scripts/run-cleanup.sh --format html --output report.html
```

### Command Line Options

```bash
./scripts/run-cleanup.sh [OPTIONS]

OPTIONS:
    -d, --dry-run           Perform a dry run without making changes
    -n, --non-interactive   Run without user interaction
    -a, --auto-fix          Automatically fix simple issues
    -p, --pattern PATTERN   Only process specific pattern
    -s, --severity LEVEL    Only process specific severity
    -f, --format FORMAT     Output format (json,yaml,text,html)
    --shell-only            Run only shell-based cleanup
    --go-only               Run only Go-based cleanup
    --no-summary            Skip summary generation
```

### Scheduled Cleanup

The system includes automated scheduling via cron jobs:

```bash
# Install scheduled cleanup
./scripts/setup-scheduled-cleanup.sh --install-cron

# List current jobs
./scripts/setup-scheduled-cleanup.sh --list-cron

# Remove scheduled jobs
./scripts/setup-scheduled-cleanup.sh --remove-cron
```

#### Schedule

- **Daily (2 AM)**: Backup file cleanup (auto-fix enabled)
- **Weekly (Sunday 1 AM)**: Deprecated comment review
- **Monthly (1st, midnight)**: Full scan with HTML report
- **Weekly (Saturday 3 AM)**: Report generation
- **Daily (6 AM)**: Health check

## Configuration

### Pattern Configuration

Edit `configs/cleanup-config.yaml` to customize detection patterns:

```yaml
patterns:
  deprecated_comments:
    regex: "(?i)(//|#|/\\*)\\s*(deprecated|fixme|todo.*remove|legacy|obsolete)"
    severity: "high"
    auto_fixable: false
    description: "Comments marking deprecated code"
    suggestion: "Review and remove or update deprecated comments"
```

### Auto-Fix Settings

```yaml
auto_fix:
  enabled_patterns:
    - "backup_files"
    - "temp_files"
  
  never_auto_fix:
    - "deprecated_functions"
    - "deprecated_types"
    - "deprecated_imports"
    - "problematic_imports"
```

## Integration

### Technical Debt Monitoring

The cleanup system integrates with the technical debt monitoring system:

- Automatic metrics updates after cleanup runs
- Prometheus metrics export
- Historical tracking of cleanup effectiveness

### CI/CD Integration

```yaml
# Example GitHub Actions workflow
- name: Run Code Cleanup
  run: |
    ./scripts/run-cleanup.sh --dry-run --non-interactive
    # Fail if critical issues found
    if grep -q "critical" cleanup-report.json; then
      echo "Critical cleanup issues found"
      exit 1
    fi
```

## Monitoring & Alerts

### Health Checks

The system includes automated health monitoring:

- Job execution monitoring
- Error rate tracking
- Disk space monitoring
- Log file size management

### Alert Thresholds

```yaml
alerts:
  critical_threshold: 1
  high_threshold: 10
  medium_threshold: 50
  low_threshold: 100
```

## Reports

### Report Formats

- **JSON**: Machine-readable format for integration
- **YAML**: Human-readable configuration format
- **HTML**: Rich web-based reports with charts
- **Text**: Simple console output

### Report Contents

- Summary statistics
- Detailed item listings
- Severity breakdown
- Recommendations
- Fix history

## Best Practices

### Development Workflow

1. **Pre-commit Hooks**: Run cleanup checks before commits
2. **Regular Reviews**: Schedule weekly cleanup reviews
3. **Gradual Adoption**: Start with safe auto-fix patterns
4. **Team Training**: Educate team on cleanup patterns

### Safety Guidelines

1. **Always Test**: Run tests after cleanup operations
2. **Backup First**: Ensure version control is up to date
3. **Review Changes**: Manually review auto-fixed items
4. **Monitor Results**: Track cleanup effectiveness over time

### Performance Considerations

1. **Incremental Scans**: Use pattern filters for large codebases
2. **Parallel Processing**: Leverage Go's concurrency for large scans
3. **Caching**: Cache scan results for repeated operations
4. **Resource Limits**: Set appropriate timeouts and memory limits

## Troubleshooting

### Common Issues

1. **Permission Errors**: Ensure scripts are executable
2. **Pattern Compilation**: Check regex syntax in configuration
3. **File Access**: Verify file permissions and paths
4. **Cron Job Failures**: Check log files for error details

### Debug Mode

```bash
# Enable debug logging
export CLEANUP_LOG_LEVEL=debug
./scripts/run-cleanup.sh
```

### Log Files

- **Main Log**: `logs/cleanup-cron.log`
- **Error Log**: `logs/cleanup-cron-error.log`
- **Reports**: `reports/` directory

## Metrics & Analytics

### Key Metrics

- **Total Items Found**: Overall deprecated code count
- **Auto-Fix Rate**: Percentage of items automatically fixed
- **Severity Distribution**: Breakdown by severity level
- **Trend Analysis**: Historical cleanup effectiveness

### Prometheus Metrics

- `kyb_cleanup_items_total`: Total items found
- `kyb_cleanup_fixes_applied`: Number of fixes applied
- `kyb_cleanup_scan_duration`: Time taken for scans
- `kyb_cleanup_errors_total`: Number of errors encountered

## Future Enhancements

### Planned Features

1. **IDE Integration**: VS Code and JetBrains plugins
2. **Machine Learning**: AI-powered pattern detection
3. **Custom Rules**: User-defined cleanup patterns
4. **Team Collaboration**: Shared cleanup configurations
5. **Advanced Reporting**: Interactive dashboards

### Roadmap

- **Q1 2025**: IDE integration and custom rules
- **Q2 2025**: Machine learning pattern detection
- **Q3 2025**: Advanced analytics and dashboards
- **Q4 2025**: Team collaboration features

## Support

### Documentation

- This document: `docs/automated-cleanup-system.md`
- Technical debt monitoring: `docs/technical-debt-monitoring.md`
- API documentation: `docs/api/technical-debt-monitor.md`

### Getting Help

1. Check the troubleshooting section
2. Review log files for error details
3. Run tests to verify system health
4. Consult the configuration documentation

---

*Last updated: August 19, 2025*
*Version: 1.0.0*
