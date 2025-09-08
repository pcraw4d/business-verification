# Security Scanning Guide for KYB Platform

This guide provides comprehensive information about the security scanning implementation in the KYB Platform, including setup, configuration, and best practices for using free and open-source security tools.

## Table of Contents

1. [Overview](#overview)
2. [Available Security Tools](#available-security-tools)
3. [Setup and Installation](#setup-and-installation)
4. [Configuration](#configuration)
5. [Running Security Scans](#running-security-scans)
6. [Understanding Results](#understanding-results)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)
9. [Integration with CI/CD](#integration-with-cicd)

## Overview

The KYB Platform implements a comprehensive security scanning strategy using only free and open-source tools. This approach ensures:

- **Cost-effective security**: No licensing fees for security tools
- **Transparency**: Open-source tools provide full visibility into scanning logic
- **Customization**: Ability to modify and extend tools as needed
- **Community support**: Large community of users and contributors

### Security Scanning Strategy

Our security scanning covers multiple layers:

1. **Static Code Analysis**: Identifies security vulnerabilities in source code
2. **Dependency Scanning**: Checks for known vulnerabilities in dependencies
3. **Secret Detection**: Prevents accidental exposure of sensitive information
4. **Container Security**: Scans Docker images for vulnerabilities
5. **Configuration Security**: Validates security configurations

## Available Security Tools

### 1. gosec - Go Security Checker

**Purpose**: Static analysis tool for Go code that identifies security vulnerabilities.

**What it scans**:
- Hardcoded credentials
- SQL injection vulnerabilities
- Insecure cryptographic practices
- File permission issues
- Network security problems

**Installation**:
```bash
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
```

**Usage**:
```bash
gosec -fmt json -out results.json ./...
```

### 2. govulncheck - Go Vulnerability Checker

**Purpose**: Identifies known vulnerabilities in Go dependencies.

**What it scans**:
- Known CVEs in Go modules
- Vulnerable function calls
- Binary-level vulnerability detection

**Installation**:
```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
```

**Usage**:
```bash
govulncheck ./...
```

### 3. Trivy - Comprehensive Security Scanner

**Purpose**: Multi-purpose security scanner for containers, filesystems, and dependencies.

**What it scans**:
- Container images
- Filesystem vulnerabilities
- Dependency vulnerabilities
- Secret detection
- Configuration issues

**Installation**:
```bash
curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin v0.50.0
```

**Usage**:
```bash
# Filesystem scan
trivy fs --format json --output results.json .

# Container scan
trivy image --format json --output results.json myimage:latest
```

### 4. Semgrep - Static Analysis Engine

**Purpose**: Fast, customizable static analysis tool with security-focused rules.

**What it scans**:
- Security vulnerabilities
- Code quality issues
- Custom rule violations
- Framework-specific issues

**Installation**:
```bash
python3 -m pip install semgrep
```

**Usage**:
```bash
semgrep --config=auto --json --output results.json .
```

### 5. OWASP ZAP - Web Application Security Scanner

**Purpose**: Dynamic application security testing (DAST) tool.

**What it scans**:
- Web application vulnerabilities
- API security issues
- Authentication problems
- Session management issues

**Installation**:
```bash
curl -sSL https://github.com/zaproxy/zaproxy/releases/download/v2.14.0/zap_2.14.0_linux.tar.gz | tar -xz -C /opt/
```

**Usage**:
```bash
/opt/zap_2.14.0/zap.sh -cmd -quickurl http://localhost:8080 -quickprogress -quickout results.json
```

## Setup and Installation

### Prerequisites

- Go 1.22 or later
- Python 3.8 or later
- Docker (for container scanning)
- Git

### Automated Installation

The security scanning tools are automatically installed in the GitHub Actions workflow:

```yaml
- name: Install security tools
  run: |
    # Install Go security tools
    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    go install golang.org/x/vuln/cmd/govulncheck@latest
    
    # Install Trivy
    curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin v0.50.0
    
    # Install Semgrep
    python3 -m pip install semgrep
```

### Manual Installation

For local development, install tools manually:

```bash
# Install Go tools
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest

# Install Trivy
curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin v0.50.0

# Install Semgrep
python3 -m pip install semgrep

# Install OWASP ZAP (optional)
curl -sSL https://github.com/zaproxy/zaproxy/releases/download/v2.14.0/zap_2.14.0_linux.tar.gz | tar -xz -C /opt/
```

## Configuration

### Security Scan Configuration

The security scanning behavior is controlled by `configs/security-scan-config.yaml`:

```yaml
# Global settings
global:
  output_dir: "./security-reports"
  fail_on_critical: false
  log_level: "info"
  parallel_scans: true

# Tool-specific configurations
tools:
  gosec:
    enabled: true
    rules: "all"
    severity: "medium"
    confidence: "medium"
  
  trivy:
    enabled: true
    severity: ["CRITICAL", "HIGH", "MEDIUM"]
    security_checks: ["vuln", "secret", "config"]
```

### Customizing Scan Rules

#### gosec Rules

Enable/disable specific security rules:

```bash
# Exclude specific rules
gosec -exclude-rule G204 -exclude-rule G304 ./...

# Include only specific rules
gosec -include-rule G101 -include-rule G102 ./...
```

#### Semgrep Rules

Use different rule sets:

```bash
# Security audit rules
semgrep --config=p/security-audit .

# OWASP Top 10 rules
semgrep --config=p/owasp-top-ten .

# Custom rules
semgrep --config=./custom-rules.yaml .
```

## Running Security Scans

### Using the Enhanced Security Scan Script

The main security scanning script provides comprehensive scanning capabilities:

```bash
# Run all security scans
./scripts/enhanced-security-scan.sh

# Run specific scan types
./scripts/enhanced-security-scan.sh --scan-type code
./scripts/enhanced-security-scan.sh --scan-type dependencies
./scripts/enhanced-security-scan.sh --scan-type secrets

# Customize output and behavior
./scripts/enhanced-security-scan.sh \
  --output-dir ./custom-reports \
  --fail-on-critical true \
  --enable-zap true
```

### Individual Tool Usage

#### gosec
```bash
# Basic scan
gosec ./...

# JSON output with specific severity
gosec -fmt json -out results.json -severity medium ./...

# Exclude specific directories
gosec -exclude-dir vendor -exclude-dir test ./...
```

#### govulncheck
```bash
# Check current module
govulncheck ./...

# Check specific package
govulncheck ./internal/api/...

# Binary mode (requires built binary)
govulncheck -mode binary ./kyb-platform
```

#### Trivy
```bash
# Filesystem scan
trivy fs --format json --output results.json .

# Container scan
trivy image --format json --output results.json kyb-platform:latest

# Specific severity levels
trivy fs --severity CRITICAL,HIGH .
```

#### Semgrep
```bash
# Auto-config (recommended)
semgrep --config=auto --json --output results.json .

# Security-focused scan
semgrep --config=p/security-audit .

# Custom configuration
semgrep --config=./semgrep-config.yaml .
```

### GitHub Actions Integration

Security scans run automatically on:

- Pull requests to main/develop branches
- Pushes to main/develop branches
- Scheduled daily scans
- Manual workflow dispatch

## Understanding Results

### Report Formats

Security scan results are generated in multiple formats:

1. **JSON**: Machine-readable format for integration
2. **Text**: Human-readable format for review
3. **SARIF**: Standard format for GitHub Security tab
4. **HTML**: Interactive format for detailed analysis

### Severity Levels

- **CRITICAL**: Immediate security risk requiring urgent attention
- **HIGH**: Significant security risk requiring prompt attention
- **MEDIUM**: Moderate security risk requiring attention
- **LOW**: Minor security risk or best practice violation
- **INFO**: Informational findings

### Common Security Issues

#### gosec Findings

- **G101**: Hardcoded credentials
- **G102**: Binding to all interfaces
- **G201**: SQL injection via format string
- **G202**: SQL injection via string concatenation
- **G301**: Poor file permissions
- **G401**: Weak cryptographic algorithms

#### Trivy Findings

- **CVE-XXXX-XXXX**: Known vulnerabilities in dependencies
- **Secret detection**: Exposed API keys, passwords, tokens
- **Configuration issues**: Insecure Docker configurations

#### Semgrep Findings

- **Security rules**: OWASP Top 10 violations
- **Code quality**: Best practice violations
- **Framework-specific**: Go-specific security issues

### Interpreting Results

1. **Review Critical and High severity issues first**
2. **Check for false positives** (common with static analysis)
3. **Prioritize based on exploitability and impact**
4. **Document decisions for ignored findings**

## Best Practices

### Development Workflow

1. **Pre-commit hooks**: Run basic security checks before commits
2. **Pull request reviews**: Require security scan approval
3. **Regular updates**: Keep security tools and dependencies updated
4. **Baseline establishment**: Document acceptable security posture

### Security Tool Configuration

1. **Tune rules**: Disable false positives, enable relevant rules
2. **Set appropriate thresholds**: Balance security vs. development velocity
3. **Regular rule updates**: Keep security rules current
4. **Custom rules**: Add project-specific security requirements

### Dependency Management

1. **Regular updates**: Keep dependencies current
2. **Vulnerability monitoring**: Subscribe to security advisories
3. **Minimal dependencies**: Reduce attack surface
4. **Dependency review**: Audit new dependencies before adding

### Secret Management

1. **Environment variables**: Use environment variables for secrets
2. **Secret scanning**: Regular scans to prevent accidental exposure
3. **Rotation**: Regular secret rotation
4. **Access control**: Limit access to secrets

### Container Security

1. **Base image selection**: Use minimal, security-focused base images
2. **Regular updates**: Keep container images updated
3. **Non-root users**: Run containers as non-root
4. **Resource limits**: Set appropriate resource constraints

## Troubleshooting

### Common Issues

#### Tool Installation Failures

```bash
# gosec installation issues
go clean -modcache
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Trivy installation issues
curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin v0.50.0

# Semgrep installation issues
python3 -m pip install --upgrade pip
python3 -m pip install semgrep
```

#### False Positives

1. **gosec**: Use `-exclude-rule` to disable specific rules
2. **Semgrep**: Create custom rules to exclude false positives
3. **Trivy**: Use `--skip-dirs` and `--skip-files` options

#### Performance Issues

1. **Parallel execution**: Enable parallel scans in configuration
2. **Selective scanning**: Scan only changed files
3. **Caching**: Use tool-specific caching mechanisms
4. **Resource limits**: Set appropriate memory and CPU limits

### Debug Mode

Enable debug logging for troubleshooting:

```bash
# Enhanced security scan with debug
./scripts/enhanced-security-scan.sh --log-level debug

# Individual tools with verbose output
gosec -fmt text -verbose ./...
trivy fs --debug .
semgrep --verbose --config=auto .
```

## Integration with CI/CD

### GitHub Actions

The security scanning is integrated into the CI/CD pipeline:

```yaml
# Security scan job
security-scan:
  name: Security Scan
  runs-on: ubuntu-latest
  steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Install security tools
      run: |
        # Tool installation commands
    
    - name: Run security scan
      run: |
        ./scripts/enhanced-security-scan.sh
    
    - name: Upload results
      uses: actions/upload-artifact@v4
      with:
        name: security-reports
        path: security-reports/
```

### Pre-commit Hooks

Set up pre-commit hooks for local development:

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: gosec
        name: gosec
        entry: gosec
        args: ["./..."]
        language: system
        pass_filenames: false
      
      - id: govulncheck
        name: govulncheck
        entry: govulncheck
        args: ["./..."]
        language: system
        pass_filenames: false
```

### IDE Integration

Configure your IDE to run security scans:

#### VS Code

```json
{
  "go.toolsEnvVars": {
    "GOSEC": "gosec"
  },
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--enable=gosec"]
}
```

#### GoLand/IntelliJ

1. Install gosec plugin
2. Configure external tools
3. Set up file watchers

## Security Scanning Checklist

### Before Deployment

- [ ] All critical and high severity issues resolved
- [ ] Dependencies updated to latest secure versions
- [ ] No hardcoded secrets in codebase
- [ ] Container images scanned and secure
- [ ] Security configurations validated

### Regular Maintenance

- [ ] Weekly security tool updates
- [ ] Monthly dependency vulnerability review
- [ ] Quarterly security rule review
- [ ] Annual security tool evaluation

### Incident Response

- [ ] Security scan results reviewed
- [ ] Critical issues triaged immediately
- [ ] False positives documented
- [ ] Remediation timeline established
- [ ] Security team notified if needed

## Conclusion

The KYB Platform's security scanning implementation provides comprehensive coverage using free and open-source tools. This approach ensures cost-effective security while maintaining high standards of code quality and vulnerability detection.

For questions or issues with security scanning, please refer to the troubleshooting section or contact the development team.

## Additional Resources

- [gosec Documentation](https://securecodewarrior.github.io/gosec/)
- [govulncheck Documentation](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)
- [Trivy Documentation](https://aquasecurity.github.io/trivy/)
- [Semgrep Documentation](https://semgrep.dev/docs/)
- [OWASP ZAP Documentation](https://www.zaproxy.org/docs/)
- [GitHub Security Tab](https://docs.github.com/en/code-security/security-overview)
