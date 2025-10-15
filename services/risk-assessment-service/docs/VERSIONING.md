# Versioning Policy

## Overview

This document outlines the versioning policy for the Risk Assessment Service. We follow [Semantic Versioning (SemVer)](https://semver.org/) to ensure clear communication about the nature of changes in each release.

## Semantic Versioning

We use the format `MAJOR.MINOR.PATCH` (e.g., `3.0.1`):

- **MAJOR** (X.0.0): Breaking changes that require code changes
- **MINOR** (X.Y.0): New features that are backward compatible
- **PATCH** (X.Y.Z): Bug fixes that are backward compatible

## Version Number Guidelines

### Major Version (X.0.0)

Major versions are released when we introduce breaking changes that require users to modify their code. Examples include:

- **API Changes**: Removing or renaming endpoints, changing request/response formats
- **Authentication Changes**: Changing authentication methods or token formats
- **Data Format Changes**: Modifying data structures or field names
- **SDK Breaking Changes**: Changing SDK method signatures or behavior
- **Database Schema Changes**: Breaking changes to database structure
- **Configuration Changes**: Removing or renaming configuration options

#### Major Version Release Process

1. **Planning Phase** (3 months before release)
   - Identify breaking changes
   - Create migration guides
   - Notify users of upcoming changes

2. **Deprecation Phase** (6 months before release)
   - Mark deprecated features
   - Add deprecation warnings
   - Provide migration documentation

3. **Release Phase**
   - Release major version
   - Provide comprehensive migration guide
   - Support both old and new versions for 3 months

4. **End of Life Phase**
   - Remove deprecated features
   - End support for previous major version

### Minor Version (X.Y.0)

Minor versions add new features while maintaining backward compatibility. Examples include:

- **New API Endpoints**: Adding new functionality
- **New SDK Methods**: Adding new capabilities
- **New Configuration Options**: Adding optional settings
- **Performance Improvements**: Optimizations that don't change behavior
- **New Features**: Additional functionality

#### Minor Version Release Process

1. **Development Phase**
   - Implement new features
   - Ensure backward compatibility
   - Add comprehensive tests

2. **Testing Phase**
   - Internal testing
   - Beta testing with select users
   - Performance validation

3. **Release Phase**
   - Release minor version
   - Update documentation
   - Announce new features

### Patch Version (X.Y.Z)

Patch versions fix bugs and security issues without changing functionality. Examples include:

- **Bug Fixes**: Resolving reported issues
- **Security Patches**: Fixing security vulnerabilities
- **Performance Fixes**: Resolving performance issues
- **Documentation Updates**: Fixing documentation errors
- **Dependency Updates**: Updating third-party libraries

#### Patch Version Release Process

1. **Issue Identification**
   - Bug reports from users
   - Security vulnerability reports
   - Internal testing findings

2. **Fix Development**
   - Implement fix
   - Add regression tests
   - Validate fix doesn't introduce new issues

3. **Release Phase**
   - Release patch version
   - Notify users of fixes
   - Update changelog

## Version Support Lifecycle

### Support Levels

| Level | Description | Duration | Updates |
|-------|-------------|----------|---------|
| **Current** | Latest major version | 18 months | Full support |
| **Maintenance** | Previous major version | 12 months | Security and critical fixes only |
| **End of Life** | Older versions | 0 months | No support |

### Support Timeline Example

```
Version 3.0.0 (Current)
├── 18 months full support
├── 12 months maintenance support
└── End of life

Version 2.0.0 (Maintenance)
├── 12 months maintenance support
└── End of life

Version 1.0.0 (End of Life)
└── No support
```

## API Versioning

### URL Versioning

We use URL path versioning for our API:

```
https://api.kyb-platform.com/v1/assess  # Version 1
https://api.kyb-platform.com/v2/assess  # Version 2
https://api.kyb-platform.com/v3/assess  # Version 3
```

### Version Header

We also support version specification via HTTP header:

```http
GET /assess HTTP/1.1
Host: api.kyb-platform.com
API-Version: 3.0
```

### SDK Versioning

SDK versions are synchronized with API versions:

- **Go SDK**: `v3.0.0` corresponds to API `v3.0.0`
- **Python SDK**: `3.0.0` corresponds to API `v3.0.0`
- **Node.js SDK**: `3.0.0` corresponds to API `v3.0.0`

## Backward Compatibility

### Guaranteed Compatibility

We guarantee backward compatibility for:

- **Minor Versions**: All minor versions within a major version
- **Patch Versions**: All patch versions within a minor version
- **API Responses**: Response format remains consistent
- **SDK Methods**: Method signatures remain unchanged
- **Configuration**: Configuration options remain valid

### Breaking Changes

Breaking changes are only introduced in major versions and include:

- **Removed Endpoints**: API endpoints that are no longer available
- **Changed Response Format**: Modifications to response structure
- **Authentication Changes**: Changes to authentication methods
- **Required Field Changes**: Making optional fields required
- **Data Type Changes**: Changing field data types

## Migration Strategy

### Deprecation Process

1. **Announcement**: 6 months advance notice
2. **Deprecation Warnings**: Add warnings to deprecated features
3. **Migration Guide**: Provide detailed migration instructions
4. **Support Period**: Continue support for 3 months after deprecation
5. **Removal**: Remove deprecated features in next major version

### Migration Support

We provide comprehensive migration support:

- **Migration Guides**: Step-by-step migration instructions
- **Code Examples**: Before/after code examples
- **SDK Updates**: Updated SDK versions with migration helpers
- **Support**: Dedicated support for migration issues
- **Tools**: Migration tools and scripts where applicable

## Release Schedule

### Regular Release Cycle

- **Major Releases**: Every 6 months (January, July)
- **Minor Releases**: Every 2 months
- **Patch Releases**: As needed (typically monthly)

### Release Calendar 2024

| Month | Version | Type | Focus |
|-------|---------|------|-------|
| January | 3.0.0 | Major | Advanced ML models |
| March | 3.1.0 | Minor | Performance improvements |
| May | 3.2.0 | Minor | New features |
| July | 4.0.0 | Major | Next generation |
| September | 4.1.0 | Minor | Enhancements |
| November | 4.2.0 | Minor | Year-end features |

### Emergency Releases

Emergency releases may be issued outside the regular schedule for:

- **Critical Security Issues**: Immediate release
- **Data Loss Bugs**: Release within 24 hours
- **Service Outages**: Release within 48 hours
- **Compliance Issues**: Release within 1 week

## Version Numbering Examples

### Major Version (Breaking Changes)

```
2.1.5 → 3.0.0
```

**Changes:**
- Changed API response format
- Removed deprecated endpoints
- Updated authentication method
- Modified SDK method signatures

### Minor Version (New Features)

```
3.0.2 → 3.1.0
```

**Changes:**
- Added new risk assessment endpoint
- Implemented batch processing
- Added webhook notifications
- Enhanced error handling

### Patch Version (Bug Fixes)

```
3.1.0 → 3.1.1
```

**Changes:**
- Fixed memory leak in predictions
- Resolved timeout issues
- Updated documentation
- Fixed SDK authentication bug

## Pre-release Versions

### Alpha Versions

Alpha versions are early development releases:

```
3.1.0-alpha.1
3.1.0-alpha.2
```

**Characteristics:**
- Unstable and may contain bugs
- Features may change significantly
- Not recommended for production use
- Internal testing only

### Beta Versions

Beta versions are feature-complete but may contain bugs:

```
3.1.0-beta.1
3.1.0-beta.2
```

**Characteristics:**
- Feature-complete
- May contain minor bugs
- Suitable for testing
- Limited production use

### Release Candidates

Release candidates are final testing versions:

```
3.1.0-rc.1
3.1.0-rc.2
```

**Characteristics:**
- Feature-complete and stable
- Final testing phase
- Production-ready
- May become final release

## Version Metadata

### Build Information

Each release includes build metadata:

```json
{
  "version": "3.1.0",
  "build": "20240115.1",
  "commit": "abc123def456",
  "timestamp": "2024-01-15T10:30:00Z",
  "environment": "production"
}
```

### Version Endpoints

API provides version information:

```http
GET /version HTTP/1.1
Host: api.kyb-platform.com

Response:
{
  "version": "3.1.0",
  "api_version": "v3",
  "build_date": "2024-01-15T10:30:00Z",
  "git_commit": "abc123def456",
  "features": [
    "risk_assessment",
    "predictions",
    "webhooks",
    "batch_processing"
  ]
}
```

## SDK Version Compatibility

### Version Matrix

| API Version | Go SDK | Python SDK | Node.js SDK | Ruby SDK | Java SDK | PHP SDK |
|-------------|--------|------------|-------------|----------|----------|---------|
| v3.0.x | v3.0.x | 3.0.x | 3.0.x | 3.0.x | 3.0.x | 3.0.x |
| v2.1.x | v2.1.x | 2.1.x | 2.1.x | 2.1.x | 2.1.x | 2.1.x |
| v2.0.x | v2.0.x | 2.0.x | 2.0.x | 2.0.x | 2.0.x | 2.0.x |

### Compatibility Rules

1. **Same Major Version**: SDK and API must have same major version
2. **Minor Version Flexibility**: SDK minor version can be different from API
3. **Patch Version Independence**: Patch versions are independent
4. **Backward Compatibility**: Newer SDK versions work with older API versions

## Version Testing

### Compatibility Testing

We test version compatibility across:

- **API Versions**: All supported API versions
- **SDK Versions**: All supported SDK versions
- **Browser Versions**: Modern browser compatibility
- **Operating Systems**: Windows, macOS, Linux
- **Programming Languages**: All supported languages

### Test Matrix

| Test Type | Frequency | Scope |
|-----------|-----------|-------|
| Unit Tests | Every commit | All versions |
| Integration Tests | Daily | Current + previous major |
| Compatibility Tests | Weekly | All supported versions |
| Performance Tests | Weekly | Current version |
| Security Tests | Weekly | All versions |

## Version Communication

### Release Announcements

We communicate version releases through:

- **Email Notifications**: Direct emails to users
- **In-app Notifications**: Notifications in admin dashboard
- **Blog Posts**: Detailed release announcements
- **Documentation Updates**: Updated documentation
- **GitHub Releases**: GitHub release notes
- **Social Media**: Twitter, LinkedIn announcements

### Breaking Change Notifications

Breaking changes are communicated through:

- **6-Month Advance Notice**: Early warning of breaking changes
- **Deprecation Warnings**: Warnings in API responses
- **Migration Guides**: Detailed migration instructions
- **Support Channels**: Dedicated support for migration
- **Webinars**: Live migration assistance sessions

## Best Practices

### For Users

1. **Stay Current**: Keep SDKs and integrations updated
2. **Test Early**: Test new versions in development first
3. **Monitor Deprecations**: Watch for deprecation notices
4. **Plan Migrations**: Plan major version migrations in advance
5. **Use Version Pinning**: Pin to specific versions in production

### For Developers

1. **Follow SemVer**: Use semantic versioning for releases
2. **Document Changes**: Document all changes in changelog
3. **Test Thoroughly**: Test all changes before release
4. **Maintain Compatibility**: Ensure backward compatibility
5. **Communicate Clearly**: Provide clear migration guidance

## Tools and Resources

### Version Management Tools

- **Git Tags**: Version tagging in Git
- **Semantic Release**: Automated version management
- **Conventional Commits**: Standardized commit messages
- **Changelog Generator**: Automated changelog generation
- **Version Bump**: Automated version bumping

### Documentation Tools

- **API Documentation**: Versioned API documentation
- **SDK Documentation**: Versioned SDK documentation
- **Migration Guides**: Step-by-step migration instructions
- **Release Notes**: Detailed release information
- **Version History**: Complete version history

## Contact and Support

### Version Support

- **Current Version**: Full support with new features
- **Previous Major**: Maintenance support only
- **Older Versions**: No support, upgrade required

### Support Channels

- **Documentation**: [https://docs.kyb-platform.com/versioning](https://docs.kyb-platform.com/versioning)
- **Email Support**: [support@kyb-platform.com](mailto:support@kyb-platform.com)
- **Migration Support**: [migration@kyb-platform.com](mailto:migration@kyb-platform.com)
- **GitHub Issues**: [https://github.com/kyb-platform/risk-assessment-service/issues](https://github.com/kyb-platform/risk-assessment-service/issues)

---

**Last Updated**: January 15, 2024  
**Next Review**: April 15, 2024
