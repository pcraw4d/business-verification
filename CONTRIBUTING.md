# Contributing to KYB Platform

Thank you for your interest in contributing to the KYB Platform! This document provides guidelines and information for contributors to help ensure a smooth and productive development process.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Development Workflow](#development-workflow)
3. [Code Standards](#code-standards)
4. [Testing Guidelines](#testing-guidelines)
5. [Documentation Standards](#documentation-standards)
6. [Security Guidelines](#security-guidelines)
7. [Performance Guidelines](#performance-guidelines)
8. [Review Process](#review-process)
9. [Release Process](#release-process)
10. [Community Guidelines](#community-guidelines)

## Getting Started

### Prerequisites

Before contributing, ensure you have the following installed:

- **Go**: 1.22 or higher
- **Git**: Latest version
- **Docker**: 20.10 or higher
- **PostgreSQL**: 14 or higher
- **Redis**: 6.0 or higher
- **Make**: 4.0 or higher

### Development Environment Setup

1. **Fork the Repository**
   ```bash
   # Fork on GitHub, then clone your fork
   git clone https://github.com/your-username/kyb-platform.git
   cd kyb-platform
   ```

2. **Set Up Development Environment**
   ```bash
   # Install dependencies
   go mod download
   
   # Install development tools
   go install github.com/cosmtrek/air@latest
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   
   # Set up pre-commit hooks
   make setup-hooks
   ```

3. **Configure Environment**
   ```bash
   # Copy environment template
   cp env.example .env
   
   # Edit environment variables
   nano .env
   ```

4. **Start Dependencies**
   ```bash
   # Start PostgreSQL and Redis
   docker-compose up -d postgres redis
   
   # Run database migrations
   make migrate
   ```

5. **Start Development Server**
   ```bash
   # Start with hot reload
   make dev
   ```

### Project Structure

```
kyb-platform/
├── cmd/                    # Application entrypoints
│   └── api/               # Main API server
├── internal/              # Private application code
│   ├── api/              # HTTP handlers and middleware
│   ├── auth/             # Authentication and authorization
│   ├── classification/   # Business classification logic
│   ├── compliance/       # Compliance checking
│   ├── config/           # Configuration management
│   ├── database/         # Database models and migrations
│   ├── observability/    # Logging, metrics, tracing
│   └── risk/             # Risk assessment logic
├── pkg/                   # Public packages
├── docs/                  # Documentation
├── test/                  # Test utilities and data
├── scripts/               # Build and deployment scripts
└── deployments/           # Deployment configurations
```

## Development Workflow

### Branch Strategy

We follow a **Git Flow** branching strategy:

- **main**: Production-ready code
- **develop**: Integration branch for features
- **feature/***: Feature development branches
- **hotfix/***: Critical bug fixes
- **release/***: Release preparation branches

### Creating a Feature Branch

```bash
# Ensure you're on develop branch
git checkout develop
git pull origin develop

# Create feature branch
git checkout -b feature/your-feature-name

# Make your changes
# ... code changes ...

# Commit your changes
git add .
git commit -m "feat: add your feature description"
```

### Commit Message Format

We follow the **Conventional Commits** specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples**:
```bash
git commit -m "feat: add business classification endpoint"
git commit -m "fix(auth): resolve JWT token validation issue"
git commit -m "docs: update API documentation"
git commit -m "test: add unit tests for classification service"
```

### Pull Request Process

1. **Create Pull Request**
   - Target the `develop` branch
   - Use descriptive title and description
   - Reference related issues

2. **Pull Request Template**
   ```markdown
   ## Description
   Brief description of changes

   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update

   ## Testing
   - [ ] Unit tests pass
   - [ ] Integration tests pass
   - [ ] Manual testing completed

   ## Checklist
   - [ ] Code follows style guidelines
   - [ ] Self-review completed
   - [ ] Documentation updated
   - [ ] Tests added/updated
   ```

3. **Code Review**
   - Address reviewer feedback
   - Ensure all checks pass
   - Update documentation if needed

## Code Standards

### Go Code Style

We follow **Go Code Review Comments** and **Effective Go**:

**Formatting**:
```bash
# Format code
go fmt ./...

# Organize imports
goimports -w .

# Run linter
golangci-lint run
```

**Naming Conventions**:
```go
// Package names: lowercase, single word
package auth

// Function names: camelCase
func validateToken(token string) error

// Variable names: camelCase
var userID string

// Constants: camelCase or UPPER_CASE
const (
    maxRetries = 3
    API_VERSION = "v1"
)

// Interface names: verb + noun
type TokenValidator interface {
    Validate(token string) error
}

// Struct names: noun
type User struct {
    ID    string `json:"id"`
    Email string `json:"email"`
}
```

**Error Handling**:
```go
// Always check errors
if err != nil {
    return fmt.Errorf("failed to process request: %w", err)
}

// Use custom error types for business logic
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}
```

**Context Usage**:
```go
// Always pass context to functions that may block
func (s *Service) ProcessRequest(ctx context.Context, req *Request) (*Response, error) {
    // Use context for cancellation and timeouts
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    case result := <-s.processAsync(req):
        return result, nil
    }
}
```

### Code Organization

**File Structure**:
```go
// File: internal/auth/service.go
package auth

import (
    "context"
    "time"
    
    "github.com/your-org/kyb-platform/internal/config"
    "github.com/your-org/kyb-platform/internal/database"
)

// Service provides authentication functionality
type Service struct {
    config *config.Config
    db     *database.DB
}

// NewService creates a new authentication service
func NewService(config *config.Config, db *database.DB) *Service {
    return &Service{
        config: config,
        db:     db,
    }
}

// Public methods first
func (s *Service) Authenticate(ctx context.Context, credentials Credentials) (*Token, error) {
    // Implementation
}

// Private methods last
func (s *Service) validateCredentials(credentials Credentials) error {
    // Implementation
}
```

**Package Organization**:
- Keep packages focused and cohesive
- Use interfaces for dependency injection
- Separate concerns (handlers, services, repositories)

### Security Guidelines

**Input Validation**:
```go
// Always validate input
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Validate request
    if err := req.Validate(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Process request
}
```

**SQL Injection Prevention**:
```go
// Use parameterized queries
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
    query := "SELECT id, email, password_hash FROM users WHERE email = $1"
    var user User
    err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.PasswordHash)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return &user, nil
}
```

**Authentication**:
```go
// Use secure password hashing
func HashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", fmt.Errorf("failed to hash password: %w", err)
    }
    return string(hash), nil
}

// Verify passwords securely
func VerifyPassword(password, hash string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
```

## Testing Guidelines

### Test Structure

**Unit Tests**:
```go
// File: internal/auth/service_test.go
package auth

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestService_Authenticate(t *testing.T) {
    // Arrange
    mockDB := &MockDB{}
    service := NewService(config, mockDB)
    
    credentials := Credentials{
        Email:    "test@example.com",
        Password: "password123",
    }
    
    mockDB.On("GetUserByEmail", mock.Anything, credentials.Email).Return(&User{
        ID:           "user-123",
        Email:        credentials.Email,
        PasswordHash: "$2a$10$hashedpassword",
    }, nil)
    
    // Act
    token, err := service.Authenticate(context.Background(), credentials)
    
    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, token)
    mockDB.AssertExpectations(t)
}
```

**Integration Tests**:
```go
// File: test/integration/auth_test.go
package integration

import (
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
)

func TestAuthEndpoint_Login(t *testing.T) {
    // Setup test server
    server := setupTestServer(t)
    defer server.Close()
    
    // Test request
    req := httptest.NewRequest("POST", "/v1/auth/login", strings.NewReader(`{
        "email": "test@example.com",
        "password": "password123"
    }`))
    req.Header.Set("Content-Type", "application/json")
    
    // Execute request
    resp, err := server.Client().Do(req)
    assert.NoError(t, err)
    defer resp.Body.Close()
    
    // Assert response
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    var result map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&result)
    assert.NoError(t, err)
    assert.Contains(t, result, "token")
}
```

### Test Coverage

**Coverage Requirements**:
- **Unit Tests**: > 90% coverage
- **Integration Tests**: Critical paths covered
- **Performance Tests**: For performance-critical code

**Running Tests**:
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test ./internal/auth -v

# Run benchmarks
go test ./internal/auth -bench=.
```

### Test Data Management

**Test Factories**:
```go
// File: test/testdata/factory.go
package testdata

import (
    "github.com/google/uuid"
    "github.com/your-org/kyb-platform/internal/database"
)

func NewTestUser() *database.User {
    return &database.User{
        ID:       uuid.New().String(),
        Email:    "test-" + uuid.New().String()[:8] + "@example.com",
        Password: "password123",
    }
}

func NewTestBusiness() *database.Business {
    return &database.Business{
        ID:   uuid.New().String(),
        Name: "Test Business " + uuid.New().String()[:8],
    }
}
```

## Documentation Standards

### Code Documentation

**Package Documentation**:
```go
// Package auth provides authentication and authorization functionality
// for the KYB Platform.
//
// This package includes:
//   - User authentication with JWT tokens
//   - Role-based access control (RBAC)
//   - Password hashing and validation
//   - Session management
package auth
```

**Function Documentation**:
```go
// Authenticate validates user credentials and returns a JWT token.
//
// The function performs the following steps:
//   1. Validates the provided credentials
//   2. Retrieves the user from the database
//   3. Verifies the password hash
//   4. Generates and returns a JWT token
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - credentials: User credentials (email and password)
//
// Returns:
//   - Token: JWT token for authenticated user
//   - error: Error if authentication fails
//
// Example:
//
//	token, err := authService.Authenticate(ctx, Credentials{
//	    Email:    "user@example.com",
//	    Password: "password123",
//	})
func (s *Service) Authenticate(ctx context.Context, credentials Credentials) (*Token, error) {
    // Implementation
}
```

### API Documentation

**OpenAPI Documentation**:
```yaml
# Update docs/api/openapi.yaml for new endpoints
paths:
  /v1/auth/login:
    post:
      summary: Authenticate user
      description: Validates user credentials and returns JWT token
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/LoginRequest'
      responses:
        '200':
          description: Authentication successful
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/LoginResponse'
```

### README Updates

**Feature Documentation**:
```markdown
## New Feature: Business Classification

### Overview
The business classification feature automatically categorizes businesses using industry codes.

### Usage
```bash
curl -X POST http://localhost:8080/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Acme Corporation"}'
```

### Configuration
Set the following environment variables:
- `CLASSIFICATION_CACHE_TTL`: Cache duration (default: 1h)
- `CLASSIFICATION_CONFIDENCE_THRESHOLD`: Minimum confidence (default: 0.8)
```

## Security Guidelines

### Security Review Process

**Security Checklist**:
- [ ] Input validation implemented
- [ ] SQL injection prevention
- [ ] XSS protection
- [ ] CSRF protection
- [ ] Authentication required
- [ ] Authorization checks
- [ ] Sensitive data encrypted
- [ ] Secrets management
- [ ] Rate limiting
- [ ] Audit logging

**Security Testing**:
```bash
# Run security scans
make security-scan

# Check for vulnerabilities
gosec ./...

# Run SAST tools
make sast-scan
```

### Secret Management

**Environment Variables**:
```bash
# Never commit secrets to version control
# Use .env files for local development
# Use secret management in production

# Example .env file (not committed)
JWT_SECRET=your-secret-key-here
DB_PASSWORD=your-db-password
API_KEY=your-api-key
```

**Production Secrets**:
```yaml
# Kubernetes secrets
apiVersion: v1
kind: Secret
metadata:
  name: kyb-secrets
type: Opaque
data:
  jwt-secret: <base64-encoded-secret>
  db-password: <base64-encoded-password>
```

## Performance Guidelines

### Performance Requirements

**Response Time Targets**:
- **API Endpoints**: < 500ms for 95% of requests
- **Database Queries**: < 100ms for 95% of queries
- **Classification**: < 200ms for single business
- **Batch Processing**: < 2s for 100 businesses

**Performance Testing**:
```bash
# Run performance tests
make perf-test

# Benchmark specific functions
go test -bench=. ./internal/classification/

# Load testing
make load-test
```

### Performance Optimization

**Database Optimization**:
```sql
-- Use indexes for frequently queried columns
CREATE INDEX CONCURRENTLY idx_businesses_name ON businesses(name);
CREATE INDEX CONCURRENTLY idx_classifications_business_id ON classifications(business_id);

-- Use connection pooling
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)
```

**Caching Strategy**:
```go
// Use Redis for caching
func (s *Service) GetClassification(ctx context.Context, businessName string) (*Classification, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("classification:%s", businessName)
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        return cached, nil
    }
    
    // Get from database
    classification, err := s.db.GetClassification(ctx, businessName)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    s.cache.Set(ctx, cacheKey, classification, time.Hour)
    return classification, nil
}
```

## Review Process

### Code Review Checklist

**Functionality**:
- [ ] Code works as intended
- [ ] Edge cases handled
- [ ] Error handling implemented
- [ ] Performance considered

**Code Quality**:
- [ ] Follows style guidelines
- [ ] Well-documented
- [ ] Tests included
- [ ] No security issues

**Architecture**:
- [ ] Follows design patterns
- [ ] Maintains separation of concerns
- [ ] No unnecessary dependencies
- [ ] Scalable design

### Review Guidelines

**For Reviewers**:
- Be constructive and respectful
- Focus on code, not the person
- Provide specific feedback
- Suggest improvements
- Approve when satisfied

**For Authors**:
- Respond to feedback promptly
- Make requested changes
- Ask questions if unclear
- Update documentation
- Re-request review when ready

## Release Process

### Release Preparation

**Version Bumping**:
```bash
# Update version in go.mod
# Update CHANGELOG.md
# Update documentation
# Run full test suite
make test-all

# Create release branch
git checkout -b release/v1.2.0
```

**Release Checklist**:
- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG updated
- [ ] Version bumped
- [ ] Security scan clean
- [ ] Performance tests pass

### Release Process

```bash
# Merge to main
git checkout main
git merge release/v1.2.0

# Tag release
git tag -a v1.2.0 -m "Release v1.2.0"

# Push to remote
git push origin main
git push origin v1.2.0

# Create GitHub release
# Update develop branch
git checkout develop
git merge main
git push origin develop
```

## Community Guidelines

### Communication

**Code of Conduct**:
- Be respectful and inclusive
- Use welcoming and inclusive language
- Be collaborative and constructive
- Focus on what is best for the community
- Show empathy towards other community members

**Communication Channels**:
- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and discussions
- **Pull Requests**: Code contributions
- **Email**: Security issues (security@kybplatform.com)

### Issue Reporting

**Bug Report Template**:
```markdown
## Bug Description
Brief description of the bug

## Steps to Reproduce
1. Step 1
2. Step 2
3. Step 3

## Expected Behavior
What you expected to happen

## Actual Behavior
What actually happened

## Environment
- OS: [e.g., Ubuntu 20.04]
- Go Version: [e.g., 1.22.0]
- KYB Platform Version: [e.g., v1.1.0]

## Additional Information
Any other relevant information
```

**Feature Request Template**:
```markdown
## Feature Description
Brief description of the feature

## Use Case
Why this feature is needed

## Proposed Solution
How you think this should be implemented

## Alternatives Considered
Other approaches you considered

## Additional Information
Any other relevant information
```

### Recognition

**Contributor Recognition**:
- Contributors listed in README.md
- Commit history preserved
- Release notes credit contributors
- Special recognition for significant contributions

**Contributor Levels**:
- **Contributor**: First contribution
- **Regular Contributor**: Multiple contributions
- **Maintainer**: Sustained contributions and reviews
- **Core Maintainer**: Project leadership and architecture

---

## Getting Help

If you need help with contributing:

1. **Check Documentation**: Review this guide and project documentation
2. **Search Issues**: Look for similar issues or discussions
3. **Ask Questions**: Use GitHub Discussions for general questions
4. **Contact Maintainers**: Reach out to project maintainers

Thank you for contributing to the KYB Platform! Your contributions help make the platform better for everyone.
