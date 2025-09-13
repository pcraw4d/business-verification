# Contributing to KYB Platform

## Overview

Thank you for your interest in contributing to the KYB Platform! This document provides guidelines and procedures for contributing to the project.

## Code of Conduct

### Our Commitment

We are committed to providing a welcoming and inclusive environment for all contributors. We expect all contributors to:

- Be respectful and constructive in all interactions
- Focus on what is best for the community
- Show empathy towards other community members
- Accept constructive criticism gracefully
- Help create a safe environment for everyone

### Unacceptable Behavior

The following behaviors are considered unacceptable:

- Harassment, discrimination, or offensive comments
- Personal attacks or trolling
- Public or private harassment
- Publishing private information without permission
- Any conduct that could reasonably be considered inappropriate

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- Go 1.22+ installed
- Docker and Docker Compose
- Git configured with your name and email
- A GitHub account
- Basic understanding of the project architecture

### Development Environment Setup

1. **Fork and Clone**:
   ```bash
   # Fork the repository on GitHub
   # Clone your fork
   git clone https://github.com/your-username/kyb-platform.git
   cd kyb-platform
   
   # Add upstream remote
   git remote add upstream https://github.com/original-owner/kyb-platform.git
   ```

2. **Install Dependencies**:
   ```bash
   # Install Go dependencies
   go mod download
   
   # Start development environment
   docker-compose -f docker-compose.dev.yml up -d
   ```

3. **Run Tests**:
   ```bash
   # Run all tests
   go test ./...
   
   # Run tests with coverage
   go test -cover ./...
   
   # Run frontend tests
   npm test
   ```

## Development Workflow

### Branch Strategy

We use a feature branch workflow:

- `main`: Production-ready code
- `develop`: Integration branch for features
- `feature/*`: Feature development branches
- `bugfix/*`: Bug fix branches
- `hotfix/*`: Critical production fixes

### Branch Naming Convention

- `feature/merchant-portfolio-management`
- `bugfix/session-timeout-issue`
- `hotfix/security-vulnerability`
- `refactor/database-optimization`

### Commit Message Format

We use conventional commits:

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
```
feat(merchant): add portfolio type filtering

- Add portfolio type filter component
- Implement filter state management
- Add unit tests for filtering logic

Closes #123
```

```
fix(auth): resolve session timeout issue

The session timeout was not being properly handled when users
were inactive for extended periods. This fix ensures proper
session cleanup and user notification.

Fixes #456
```

### Pull Request Process

1. **Create Feature Branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Changes**:
   - Write clean, well-documented code
   - Add appropriate tests
   - Update documentation as needed
   - Follow coding standards

3. **Test Your Changes**:
   ```bash
   # Run all tests
   go test ./...
   
   # Run linting
   golangci-lint run
   
   # Run frontend tests
   npm test
   
   # Test in development environment
   docker-compose -f docker-compose.dev.yml up -d
   ```

4. **Commit Changes**:
   ```bash
   git add .
   git commit -m "feat(component): add new feature"
   ```

5. **Push and Create PR**:
   ```bash
   git push origin feature/your-feature-name
   # Create pull request on GitHub
   ```

## Coding Standards

### Go Code Standards

**Naming Conventions**:
```go
// Package names: lowercase, no underscores
package merchantportfolio

// Constants: CamelCase with descriptive names
const (
    DefaultTimeout = 30 * time.Second
    MaxRetryAttempts = 3
)

// Variables: camelCase
var (
    serverPort = 8080
    databaseURL string
)

// Types: PascalCase
type Merchant struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

// Functions: PascalCase for exported, camelCase for private
func GetMerchant(id string) (*Merchant, error) {
    return getMerchantInternal(id)
}

func getMerchantInternal(id string) (*Merchant, error) {
    // Implementation
}
```

**Error Handling**:
```go
// Good: Proper error handling with context
func ProcessMerchant(data MerchantData) error {
    if err := validateMerchant(data); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    if err := saveMerchant(data); err != nil {
        return fmt.Errorf("failed to save merchant: %w", err)
    }
    
    return nil
}

// Bad: Ignoring errors
func ProcessMerchant(data MerchantData) error {
    validateMerchant(data) // Error ignored
    saveMerchant(data)     // Error ignored
    return nil
}
```

**Function Design**:
```go
// Good: Single responsibility, clear purpose
func ValidateMerchantEmail(email string) error {
    if email == "" {
        return errors.New("email is required")
    }
    
    if !isValidEmail(email) {
        return errors.New("email format is invalid")
    }
    
    return nil
}

// Good: Use context for cancellation and timeouts
func ProcessMerchantVerification(ctx context.Context, merchant Merchant) error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    return performVerification(ctx, merchant)
}
```

### Frontend Code Standards

**JavaScript/HTML Standards**:
```javascript
// Use descriptive variable names
const merchantSearchResults = await searchMerchants(query);
const isPortfolioFilterActive = portfolioFilter.isActive();

// Use async/await for promises
async function loadMerchantData(merchantId) {
    try {
        const response = await fetch(`/api/merchants/${merchantId}`);
        const merchant = await response.json();
        return merchant;
    } catch (error) {
        console.error('Failed to load merchant data:', error);
        throw error;
    }
}

// Use meaningful function names
function validateMerchantForm(formData) {
    const errors = [];
    
    if (!formData.name) {
        errors.push('Merchant name is required');
    }
    
    if (!formData.email || !isValidEmail(formData.email)) {
        errors.push('Valid email is required');
    }
    
    return errors;
}
```

**CSS Standards**:
```css
/* Use BEM methodology for class naming */
.merchant-card {
    border: 1px solid #e0e0e0;
    border-radius: 8px;
    padding: 16px;
}

.merchant-card__header {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.merchant-card__title {
    font-size: 18px;
    font-weight: 600;
    color: #333;
}

.merchant-card--highlighted {
    border-color: #007bff;
    box-shadow: 0 2px 8px rgba(0, 123, 255, 0.1);
}

.merchant-card--highlighted .merchant-card__title {
    color: #007bff;
}
```

### Documentation Standards

**Code Documentation**:
```go
// Package documentation
// Package merchantportfolio provides merchant portfolio management functionality.
//
// The package implements comprehensive merchant portfolio operations including:
//   - Merchant CRUD operations
//   - Portfolio type management
//   - Risk level assignment
//   - Session management
//
// Example usage:
//
//	service := merchantportfolio.NewService(repo, logger)
//	merchants, err := service.GetMerchants(ctx, filters)
//	if err != nil {
//	    log.Fatal(err)
//	}
package merchantportfolio

// Merchant represents a business entity in the portfolio.
//
// A merchant contains all necessary information for business verification
// and compliance tracking.
type Merchant struct {
    // ID is the unique identifier for the merchant
    ID string `json:"id"`
    
    // Name is the business name
    Name string `json:"name"`
    
    // BusinessType represents the type of business
    BusinessType string `json:"business_type"`
    
    // RiskLevel indicates the assessed risk level
    RiskLevel RiskLevel `json:"risk_level"`
}

// GetMerchants retrieves merchants based on the provided filters.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - filters: Filter criteria for merchant selection
//
// Returns:
//   - []Merchant: List of matching merchants
//   - error: Any error that occurred during retrieval
//
// Example:
//
//	filters := MerchantFilters{
//	    PortfolioType: "onboarded",
//	    RiskLevel:     "low",
//	}
//	
//	merchants, err := service.GetMerchants(ctx, filters)
//	if err != nil {
//	    return fmt.Errorf("failed to get merchants: %w", err)
//	}
func (s *Service) GetMerchants(ctx context.Context, filters MerchantFilters) ([]Merchant, error) {
    // Implementation
}
```

## Testing Guidelines

### Unit Testing

**Go Unit Tests**:
```go
func TestMerchantService_GetMerchants(t *testing.T) {
    tests := []struct {
        name           string
        filters        MerchantFilters
        mockSetup      func(*mocks.Repository)
        expectedResult []Merchant
        expectedError  string
    }{
        {
            name: "successful retrieval",
            filters: MerchantFilters{
                PortfolioType: "onboarded",
            },
            mockSetup: func(repo *mocks.Repository) {
                repo.On("GetMerchants", mock.Anything, mock.Anything).
                    Return([]Merchant{
                        {ID: "1", Name: "Test Merchant"},
                    }, nil)
            },
            expectedResult: []Merchant{
                {ID: "1", Name: "Test Merchant"},
            },
        },
        {
            name: "repository error",
            filters: MerchantFilters{},
            mockSetup: func(repo *mocks.Repository) {
                repo.On("GetMerchants", mock.Anything, mock.Anything).
                    Return(nil, errors.New("database error"))
            },
            expectedError: "database error",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockRepo := &mocks.Repository{}
            if tt.mockSetup != nil {
                tt.mockSetup(mockRepo)
            }
            
            service := NewService(mockRepo, zap.NewNop())
            
            // Execute
            result, err := service.GetMerchants(context.Background(), tt.filters)
            
            // Assert
            if tt.expectedError != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedResult, result)
            }
            
            mockRepo.AssertExpectations(t)
        })
    }
}
```

**Frontend Unit Tests**:
```javascript
// merchant-search.test.js
describe('MerchantSearch', () => {
    let merchantSearch;
    let mockApiClient;
    
    beforeEach(() => {
        mockApiClient = {
            searchMerchants: jest.fn()
        };
        merchantSearch = new MerchantSearch(mockApiClient);
    });
    
    describe('search', () => {
        it('should search merchants with valid query', async () => {
            // Arrange
            const query = 'test merchant';
            const expectedResults = [
                { id: '1', name: 'Test Merchant 1' },
                { id: '2', name: 'Test Merchant 2' }
            ];
            mockApiClient.searchMerchants.mockResolvedValue(expectedResults);
            
            // Act
            const results = await merchantSearch.search(query);
            
            // Assert
            expect(mockApiClient.searchMerchants).toHaveBeenCalledWith(query);
            expect(results).toEqual(expectedResults);
        });
        
        it('should handle search errors gracefully', async () => {
            // Arrange
            const query = 'test merchant';
            const error = new Error('Search failed');
            mockApiClient.searchMerchants.mockRejectedValue(error);
            
            // Act & Assert
            await expect(merchantSearch.search(query)).rejects.toThrow('Search failed');
        });
    });
});
```

### Integration Testing

**API Integration Tests**:
```go
func TestMerchantAPI_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Setup test server
    server := setupTestServer(t)
    defer server.Close()
    
    t.Run("create and retrieve merchant", func(t *testing.T) {
        // Create merchant
        merchantData := MerchantData{
            Name:         "Test Merchant",
            BusinessType: "Retail",
            Email:        "test@merchant.com",
        }
        
        createResp, err := http.Post(
            server.URL+"/api/merchants",
            "application/json",
            strings.NewReader(toJSON(merchantData)),
        )
        require.NoError(t, err)
        require.Equal(t, http.StatusCreated, createResp.StatusCode)
        
        var createdMerchant Merchant
        err = json.NewDecoder(createResp.Body).Decode(&createdMerchant)
        require.NoError(t, err)
        
        // Retrieve merchant
        getResp, err := http.Get(server.URL + "/api/merchants/" + createdMerchant.ID)
        require.NoError(t, err)
        require.Equal(t, http.StatusOK, getResp.StatusCode)
        
        var retrievedMerchant Merchant
        err = json.NewDecoder(getResp.Body).Decode(&retrievedMerchant)
        require.NoError(t, err)
        
        assert.Equal(t, createdMerchant.ID, retrievedMerchant.ID)
        assert.Equal(t, createdMerchant.Name, retrievedMerchant.Name)
    })
}
```

### End-to-End Testing

**Playwright E2E Tests**:
```javascript
// merchant-workflow.spec.js
import { test, expect } from '@playwright/test';

test.describe('Merchant Workflow', () => {
    test('should complete merchant onboarding workflow', async ({ page }) => {
        // Navigate to merchant portfolio
        await page.goto('/merchant-portfolio');
        
        // Search for merchant
        await page.fill('[data-testid="merchant-search"]', 'Test Merchant');
        await page.click('[data-testid="search-button"]');
        
        // Wait for results
        await page.waitForSelector('[data-testid="merchant-list"]');
        
        // Select merchant
        await page.click('[data-testid="merchant-card"]:first-child');
        
        // Verify merchant detail view
        await expect(page.locator('[data-testid="merchant-detail"]')).toBeVisible();
        await expect(page.locator('[data-testid="merchant-name"]')).toContainText('Test Merchant');
        
        // Update merchant status
        await page.selectOption('[data-testid="portfolio-type-select"]', 'onboarded');
        await page.click('[data-testid="save-button"]');
        
        // Verify status update
        await expect(page.locator('[data-testid="status-indicator"]')).toContainText('Onboarded');
    });
});
```

## Code Review Process

### Review Checklist

**For Authors**:
- [ ] Code follows project coding standards
- [ ] All tests pass
- [ ] New features have appropriate tests
- [ ] Documentation is updated
- [ ] No sensitive data in commits
- [ ] Commit messages follow conventional format

**For Reviewers**:
- [ ] Code is readable and well-structured
- [ ] Logic is correct and efficient
- [ ] Error handling is appropriate
- [ ] Security considerations are addressed
- [ ] Performance implications are considered
- [ ] Tests are comprehensive and meaningful

### Review Guidelines

**Be Constructive**:
- Focus on the code, not the person
- Provide specific suggestions for improvement
- Explain the reasoning behind your feedback
- Acknowledge good practices and solutions

**Be Thorough**:
- Review all changed files
- Check for potential security issues
- Verify test coverage
- Consider edge cases and error scenarios

**Be Responsive**:
- Respond to review comments promptly
- Ask questions if feedback is unclear
- Be open to discussion and alternative approaches
- Update code based on feedback

## Issue Reporting

### Bug Reports

When reporting bugs, please include:

1. **Clear Description**: What happened vs. what you expected
2. **Steps to Reproduce**: Detailed steps to reproduce the issue
3. **Environment**: OS, browser, version information
4. **Screenshots**: If applicable, include screenshots
5. **Logs**: Relevant error logs or console output

**Bug Report Template**:
```markdown
## Bug Description
Brief description of the bug

## Steps to Reproduce
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

## Expected Behavior
What you expected to happen

## Actual Behavior
What actually happened

## Environment
- OS: [e.g. macOS 12.0]
- Browser: [e.g. Chrome 95.0]
- Version: [e.g. v1.2.3]

## Additional Context
Any other context about the problem
```

### Feature Requests

When requesting features, please include:

1. **Problem Statement**: What problem does this solve?
2. **Proposed Solution**: How should it work?
3. **Alternatives**: Other solutions you've considered
4. **Additional Context**: Any other relevant information

**Feature Request Template**:
```markdown
## Feature Description
Brief description of the feature

## Problem Statement
What problem does this feature solve?

## Proposed Solution
How should this feature work?

## Alternatives Considered
What other solutions have you considered?

## Additional Context
Any other context or screenshots about the feature request
```

## Release Process

### Version Numbering

We use semantic versioning (SemVer):

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality in a backwards compatible manner
- **PATCH**: Backwards compatible bug fixes

### Release Checklist

- [ ] All tests pass
- [ ] Documentation is updated
- [ ] Version number is incremented
- [ ] Changelog is updated
- [ ] Release notes are prepared
- [ ] Security review is completed

### Release Process

1. **Create Release Branch**:
   ```bash
   git checkout -b release/v1.2.0
   ```

2. **Update Version**:
   ```bash
   # Update version in go.mod
   # Update version in package.json
   # Update version in documentation
   ```

3. **Create Release**:
   ```bash
   git tag v1.2.0
   git push origin v1.2.0
   ```

4. **Deploy**:
   ```bash
   # Deploy to staging
   # Run integration tests
   # Deploy to production
   ```

## Getting Help

### Resources

- **Documentation**: Check the docs/ directory
- **Issues**: Search existing GitHub issues
- **Discussions**: Use GitHub Discussions for questions
- **Code Review**: Ask for help in pull requests

### Contact

- **Maintainers**: @kyb-platform-maintainers
- **Security Issues**: security@kyb-platform.com
- **General Questions**: Use GitHub Discussions

## Recognition

Contributors will be recognized in:

- **CONTRIBUTORS.md**: List of all contributors
- **Release Notes**: Recognition for significant contributions
- **GitHub**: Contributor badges and statistics

Thank you for contributing to the KYB Platform! Your contributions help make the platform better for everyone.
