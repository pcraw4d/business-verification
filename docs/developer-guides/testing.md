# KYB Platform Testing Guide

## Overview

This guide provides comprehensive testing strategies, guidelines, and best practices for the KYB Platform. The testing approach follows a pyramid structure with unit tests at the base, integration tests in the middle, and end-to-end tests at the top.

## Testing Strategy

### Testing Pyramid

```
        /\
       /  \
      / E2E \     (Few, high-level user journeys)
     /______\
    /        \
   /Integration\  (Some, API and component integration)
  /____________\
 /              \
/    Unit Tests   \  (Many, individual functions and components)
/_________________\
```

### Testing Levels

1. **Unit Tests**: Test individual functions, methods, and components in isolation
2. **Integration Tests**: Test interactions between components and external systems
3. **End-to-End Tests**: Test complete user workflows from start to finish
4. **Performance Tests**: Test system performance under various loads
5. **Security Tests**: Test security vulnerabilities and compliance

## Unit Testing

### Go Unit Testing

#### Test Structure

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name           string
        input          InputType
        expectedOutput OutputType
        expectedError  string
    }{
        {
            name:           "successful case",
            input:          validInput,
            expectedOutput: expectedResult,
            expectedError:  "",
        },
        {
            name:           "error case",
            input:          invalidInput,
            expectedOutput: nil,
            expectedError:  "expected error message",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            // Setup test data and mocks
            
            // Act
            result, err := FunctionToTest(tt.input)
            
            // Assert
            if tt.expectedError != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedOutput, result)
            }
        })
    }
}
```

#### Service Layer Testing

```go
func TestMerchantService_GetMerchants(t *testing.T) {
    tests := []struct {
        name           string
        filters        MerchantFilters
        mockSetup      func(*mocks.MerchantRepository)
        expectedResult []Merchant
        expectedError  string
    }{
        {
            name: "successful retrieval",
            filters: MerchantFilters{
                PortfolioType: "onboarded",
                Limit:         10,
            },
            mockSetup: func(repo *mocks.MerchantRepository) {
                repo.On("GetMerchants", mock.Anything, mock.Anything).
                    Return([]Merchant{
                        {ID: "1", Name: "Test Merchant", PortfolioType: "onboarded"},
                    }, nil)
            },
            expectedResult: []Merchant{
                {ID: "1", Name: "Test Merchant", PortfolioType: "onboarded"},
            },
        },
        {
            name: "repository error",
            filters: MerchantFilters{},
            mockSetup: func(repo *mocks.MerchantRepository) {
                repo.On("GetMerchants", mock.Anything, mock.Anything).
                    Return(nil, errors.New("database error"))
            },
            expectedError: "database error",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockRepo := &mocks.MerchantRepository{}
            if tt.mockSetup != nil {
                tt.mockSetup(mockRepo)
            }
            
            service := NewMerchantService(mockRepo, zap.NewNop())
            
            // Execute
            result, err := service.GetMerchants(context.Background(), tt.filters)
            
            // Assert
            if tt.expectedError != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedResult, result)
            }
            
            mockRepo.AssertExpectations(t)
        })
    }
}
```

#### Repository Testing

```go
func TestMerchantRepository_GetMerchant(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping database test")
    }
    
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    repo := NewMerchantRepository(db, zap.NewNop())
    
    t.Run("existing merchant", func(t *testing.T) {
        // Insert test data
        merchant := &Merchant{
            ID:           "test-123",
            Name:         "Test Merchant",
            BusinessType: "Retail",
        }
        err := repo.CreateMerchant(context.Background(), merchant)
        require.NoError(t, err)
        
        // Test retrieval
        result, err := repo.GetMerchant(context.Background(), "test-123")
        require.NoError(t, err)
        assert.Equal(t, merchant.ID, result.ID)
        assert.Equal(t, merchant.Name, result.Name)
    })
    
    t.Run("non-existent merchant", func(t *testing.T) {
        _, err := repo.GetMerchant(context.Background(), "non-existent")
        assert.Error(t, err)
        assert.Equal(t, ErrMerchantNotFound, err)
    })
}
```

#### Handler Testing

```go
func TestMerchantHandler_GetMerchants(t *testing.T) {
    tests := []struct {
        name           string
        queryParams    map[string]string
        mockSetup      func(*mocks.MerchantService)
        expectedStatus int
        expectedError  string
    }{
        {
            name: "successful request",
            queryParams: map[string]string{
                "page":  "1",
                "limit": "10",
            },
            mockSetup: func(service *mocks.MerchantService) {
                service.On("GetMerchants", mock.Anything, mock.Anything).
                    Return([]Merchant{{ID: "1", Name: "Test"}}, Pagination{}, nil)
            },
            expectedStatus: http.StatusOK,
        },
        {
            name: "invalid page parameter",
            queryParams: map[string]string{
                "page": "invalid",
            },
            expectedStatus: http.StatusBadRequest,
            expectedError:  "INVALID_REQUEST",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockService := &mocks.MerchantService{}
            if tt.mockSetup != nil {
                tt.mockSetup(mockService)
            }
            
            handler := NewMerchantHandler(mockService, zap.NewNop())
            
            // Create request
            req := httptest.NewRequest("GET", "/api/v1/merchants", nil)
            q := req.URL.Query()
            for key, value := range tt.queryParams {
                q.Add(key, value)
            }
            req.URL.RawQuery = q.Encode()
            
            // Create response recorder
            w := httptest.NewRecorder()
            
            // Execute
            handler.GetMerchants(w, req)
            
            // Assert
            assert.Equal(t, tt.expectedStatus, w.Code)
            
            if tt.expectedError != "" {
                var errorResp ErrorResponse
                err := json.Unmarshal(w.Body.Bytes(), &errorResp)
                require.NoError(t, err)
                assert.Equal(t, tt.expectedError, errorResp.Error.Code)
            }
            
            mockService.AssertExpectations(t)
        })
    }
}
```

### Frontend Unit Testing

#### JavaScript Component Testing

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
        
        it('should debounce search requests', async () => {
            // Arrange
            const query = 'test merchant';
            mockApiClient.searchMerchants.mockResolvedValue([]);
            
            // Act
            merchantSearch.search(query);
            merchantSearch.search(query);
            merchantSearch.search(query);
            
            // Wait for debounce
            await new Promise(resolve => setTimeout(resolve, 300));
            
            // Assert
            expect(mockApiClient.searchMerchants).toHaveBeenCalledTimes(1);
        });
    });
});
```

#### HTML Component Testing

```javascript
// portfolio-filter.test.js
describe('PortfolioTypeFilter', () => {
    let filter;
    let container;
    
    beforeEach(() => {
        container = document.createElement('div');
        document.body.appendChild(container);
        filter = new PortfolioTypeFilter(container);
    });
    
    afterEach(() => {
        document.body.removeChild(container);
    });
    
    describe('filter selection', () => {
        it('should emit filter change event when option selected', () => {
            // Arrange
            const onChangeSpy = jest.fn();
            filter.on('change', onChangeSpy);
            
            // Act
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            onboardedOption.click();
            
            // Assert
            expect(onChangeSpy).toHaveBeenCalledWith('onboarded');
        });
        
        it('should update visual state when filter selected', () => {
            // Act
            const onboardedOption = container.querySelector('[data-value="onboarded"]');
            onboardedOption.click();
            
            // Assert
            expect(onboardedOption.classList.contains('active')).toBe(true);
        });
    });
});
```

## Integration Testing

### API Integration Tests

```go
func TestMerchantAPI_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Setup test server
    server := setupTestServer(t)
    defer server.Close()
    
    t.Run("complete merchant workflow", func(t *testing.T) {
        // Create merchant
        merchantData := MerchantRequest{
            Name:         "Integration Test Merchant",
            BusinessType: "Technology",
            ContactInfo: &ContactInfo{
                Email: "test@integration.com",
            },
        }
        
        createResp, err := http.Post(
            server.URL+"/api/v1/merchants",
            "application/json",
            strings.NewReader(toJSON(merchantData)),
        )
        require.NoError(t, err)
        require.Equal(t, http.StatusCreated, createResp.StatusCode)
        
        var createdMerchant Merchant
        err = json.NewDecoder(createResp.Body).Decode(&createdMerchant)
        require.NoError(t, err)
        
        // Update merchant
        updateData := MerchantRequest{
            PortfolioType: "onboarded",
            RiskLevel:     "low",
        }
        
        updateReq, err := http.NewRequest(
            "PUT",
            server.URL+"/api/v1/merchants/"+createdMerchant.ID,
            strings.NewReader(toJSON(updateData)),
        )
        require.NoError(t, err)
        updateReq.Header.Set("Content-Type", "application/json")
        
        updateResp, err := http.DefaultClient.Do(updateReq)
        require.NoError(t, err)
        require.Equal(t, http.StatusOK, updateResp.StatusCode)
        
        // Verify update
        getResp, err := http.Get(server.URL + "/api/v1/merchants/" + createdMerchant.ID)
        require.NoError(t, err)
        require.Equal(t, http.StatusOK, getResp.StatusCode)
        
        var updatedMerchant Merchant
        err = json.NewDecoder(getResp.Body).Decode(&updatedMerchant)
        require.NoError(t, err)
        
        assert.Equal(t, "onboarded", updatedMerchant.PortfolioType)
        assert.Equal(t, "low", updatedMerchant.RiskLevel)
    })
}
```

### Database Integration Tests

```go
func TestMerchantRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    repo := NewMerchantRepository(db, zap.NewNop())
    
    t.Run("bulk operations", func(t *testing.T) {
        // Create multiple merchants
        merchants := []*Merchant{
            {ID: "bulk-1", Name: "Bulk Merchant 1", BusinessType: "Retail"},
            {ID: "bulk-2", Name: "Bulk Merchant 2", BusinessType: "Technology"},
            {ID: "bulk-3", Name: "Bulk Merchant 3", BusinessType: "Services"},
        }
        
        for _, merchant := range merchants {
            err := repo.CreateMerchant(context.Background(), merchant)
            require.NoError(t, err)
        }
        
        // Test bulk retrieval
        filters := MerchantFilters{
            Limit: 10,
        }
        
        results, err := repo.GetMerchants(context.Background(), filters)
        require.NoError(t, err)
        assert.Len(t, results, 3)
        
        // Test bulk update
        updateData := map[string]interface{}{
            "portfolio_type": "onboarded",
        }
        
        err = repo.BulkUpdateMerchants(context.Background(), []string{"bulk-1", "bulk-2"}, updateData)
        require.NoError(t, err)
        
        // Verify updates
        merchant1, err := repo.GetMerchant(context.Background(), "bulk-1")
        require.NoError(t, err)
        assert.Equal(t, "onboarded", merchant1.PortfolioType)
    })
}
```

## End-to-End Testing

### Playwright E2E Tests

```javascript
// merchant-workflow.spec.js
import { test, expect } from '@playwright/test';

test.describe('Merchant Workflow', () => {
    test.beforeEach(async ({ page }) => {
        // Navigate to merchant portfolio
        await page.goto('/merchant-portfolio');
        
        // Wait for page to load
        await page.waitForSelector('[data-testid="merchant-list"]');
    });
    
    test('should complete merchant onboarding workflow', async ({ page }) => {
        // Search for merchant
        await page.fill('[data-testid="merchant-search"]', 'Test Merchant');
        await page.click('[data-testid="search-button"]');
        
        // Wait for search results
        await page.waitForSelector('[data-testid="merchant-card"]');
        
        // Select first merchant
        await page.click('[data-testid="merchant-card"]:first-child');
        
        // Verify merchant detail view
        await expect(page.locator('[data-testid="merchant-detail"]')).toBeVisible();
        await expect(page.locator('[data-testid="merchant-name"]')).toContainText('Test Merchant');
        
        // Update merchant status
        await page.selectOption('[data-testid="portfolio-type-select"]', 'onboarded');
        await page.click('[data-testid="save-button"]');
        
        // Verify status update
        await expect(page.locator('[data-testid="status-indicator"]')).toContainText('Onboarded');
        
        // Verify success message
        await expect(page.locator('[data-testid="success-message"]')).toBeVisible();
    });
    
    test('should handle bulk operations', async ({ page }) => {
        // Select multiple merchants
        await page.check('[data-testid="merchant-checkbox"]:nth-child(1)');
        await page.check('[data-testid="merchant-checkbox"]:nth-child(2)');
        
        // Open bulk operations panel
        await page.click('[data-testid="bulk-operations-button"]');
        
        // Select bulk operation
        await page.selectOption('[data-testid="bulk-operation-select"]', 'update_portfolio_type');
        await page.selectOption('[data-testid="bulk-value-select"]', 'onboarded');
        
        // Execute bulk operation
        await page.click('[data-testid="execute-bulk-operation"]');
        
        // Wait for operation to complete
        await page.waitForSelector('[data-testid="bulk-operation-complete"]');
        
        // Verify results
        await expect(page.locator('[data-testid="bulk-operation-results"]')).toContainText('2 merchants updated');
    });
    
    test('should handle merchant comparison', async ({ page }) => {
        // Select first merchant for comparison
        await page.check('[data-testid="merchant-checkbox"]:nth-child(1)');
        
        // Click compare button
        await page.click('[data-testid="compare-button"]');
        
        // Select second merchant for comparison
        await page.check('[data-testid="merchant-checkbox"]:nth-child(2)');
        
        // Open comparison view
        await page.click('[data-testid="open-comparison"]');
        
        // Verify comparison interface
        await expect(page.locator('[data-testid="comparison-view"]')).toBeVisible();
        await expect(page.locator('[data-testid="merchant-1-details"]')).toBeVisible();
        await expect(page.locator('[data-testid="merchant-2-details"]')).toBeVisible();
        
        // Generate comparison report
        await page.click('[data-testid="generate-report"]');
        
        // Verify report generation
        await expect(page.locator('[data-testid="comparison-report"]')).toBeVisible();
    });
});
```

### Cross-Browser Testing

```javascript
// cross-browser.spec.js
import { test, expect, devices } from '@playwright/test';

const browsers = ['chromium', 'firefox', 'webkit'];

browsers.forEach(browserName => {
    test.describe(`${browserName} browser`, () => {
        test.use({ ...devices[browserName] });
        
        test('merchant portfolio should work across browsers', async ({ page }) => {
            await page.goto('/merchant-portfolio');
            
            // Test basic functionality
            await expect(page.locator('[data-testid="merchant-list"]')).toBeVisible();
            
            // Test search functionality
            await page.fill('[data-testid="merchant-search"]', 'test');
            await page.click('[data-testid="search-button"]');
            
            // Wait for results
            await page.waitForSelector('[data-testid="merchant-card"]');
            
            // Test merchant selection
            await page.click('[data-testid="merchant-card"]:first-child');
            await expect(page.locator('[data-testid="merchant-detail"]')).toBeVisible();
        });
    });
});
```

## Performance Testing

### Load Testing

```go
func TestMerchantAPI_Performance(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping performance test")
    }
    
    server := setupTestServer(t)
    defer server.Close()
    
    t.Run("concurrent requests", func(t *testing.T) {
        const numRequests = 100
        const concurrency = 10
        
        var wg sync.WaitGroup
        results := make(chan time.Duration, numRequests)
        
        for i := 0; i < concurrency; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                
                for j := 0; j < numRequests/concurrency; j++ {
                    start := time.Now()
                    
                    resp, err := http.Get(server.URL + "/api/v1/merchants")
                    require.NoError(t, err)
                    require.Equal(t, http.StatusOK, resp.StatusCode)
                    
                    duration := time.Since(start)
                    results <- duration
                    
                    resp.Body.Close()
                }
            }()
        }
        
        wg.Wait()
        close(results)
        
        // Analyze results
        var totalDuration time.Duration
        var maxDuration time.Duration
        count := 0
        
        for duration := range results {
            totalDuration += duration
            if duration > maxDuration {
                maxDuration = duration
            }
            count++
        }
        
        avgDuration := totalDuration / time.Duration(count)
        
        // Assert performance requirements
        assert.Less(t, avgDuration, 100*time.Millisecond, "Average response time should be less than 100ms")
        assert.Less(t, maxDuration, 500*time.Millisecond, "Max response time should be less than 500ms")
    })
}
```

### Memory Testing

```go
func TestMerchantService_MemoryUsage(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping memory test")
    }
    
    // Setup
    mockRepo := &mocks.MerchantRepository{}
    service := NewMerchantService(mockRepo, zap.NewNop())
    
    // Mock large dataset
    largeMerchantList := make([]Merchant, 10000)
    for i := 0; i < 10000; i++ {
        largeMerchantList[i] = Merchant{
            ID:   fmt.Sprintf("merchant-%d", i),
            Name: fmt.Sprintf("Merchant %d", i),
        }
    }
    
    mockRepo.On("GetMerchants", mock.Anything, mock.Anything).
        Return(largeMerchantList, Pagination{}, nil)
    
    // Measure memory usage
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Execute operation
    _, err := service.GetMerchants(context.Background(), MerchantFilters{Limit: 10000})
    require.NoError(t, err)
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    // Calculate memory increase
    memoryIncrease := m2.Alloc - m1.Alloc
    
    // Assert memory usage is reasonable (less than 10MB for 10k merchants)
    assert.Less(t, memoryIncrease, uint64(10*1024*1024), "Memory usage should be less than 10MB")
}
```

## Security Testing

### Input Validation Testing

```go
func TestMerchantHandler_Security(t *testing.T) {
    handler := NewMerchantHandler(&mocks.MerchantService{}, zap.NewNop())
    
    tests := []struct {
        name        string
        input       string
        expectedStatus int
    }{
        {
            name:        "SQL injection attempt",
            input:       "'; DROP TABLE merchants; --",
            expectedStatus: http.StatusBadRequest,
        },
        {
            name:        "XSS attempt",
            input:       "<script>alert('xss')</script>",
            expectedStatus: http.StatusBadRequest,
        },
        {
            name:        "Path traversal attempt",
            input:       "../../../etc/passwd",
            expectedStatus: http.StatusBadRequest,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", "/api/v1/merchants/search?q="+url.QueryEscape(tt.input), nil)
            w := httptest.NewRecorder()
            
            handler.SearchMerchants(w, req)
            
            assert.Equal(t, tt.expectedStatus, w.Code)
        })
    }
}
```

### Authentication Testing

```go
func TestAuthMiddleware_Security(t *testing.T) {
    middleware := AuthMiddleware("test-secret")
    
    tests := []struct {
        name           string
        authHeader     string
        expectedStatus int
    }{
        {
            name:           "valid token",
            authHeader:     "Bearer " + generateValidToken("test-secret"),
            expectedStatus: http.StatusOK,
        },
        {
            name:           "invalid token",
            authHeader:     "Bearer invalid-token",
            expectedStatus: http.StatusUnauthorized,
        },
        {
            name:           "missing token",
            authHeader:     "",
            expectedStatus: http.StatusUnauthorized,
        },
        {
            name:           "malformed header",
            authHeader:     "InvalidFormat token",
            expectedStatus: http.StatusUnauthorized,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", "/api/v1/merchants", nil)
            if tt.authHeader != "" {
                req.Header.Set("Authorization", tt.authHeader)
            }
            
            w := httptest.NewRecorder()
            handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
            }))
            
            handler.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expectedStatus, w.Code)
        })
    }
}
```

## Test Data Management

### Test Fixtures

```go
// test/fixtures/merchant_fixtures.go
package fixtures

import (
    "time"
    "github.com/kyb-platform/internal/models"
)

func GetTestMerchants() []models.Merchant {
    return []models.Merchant{
        {
            ID:           "merchant-1",
            Name:         "Acme Corporation",
            BusinessType: "Retail",
            IndustryCode: "5411",
            PortfolioType: "onboarded",
            RiskLevel:    "low",
            CreatedAt:    time.Now().Add(-24 * time.Hour),
            UpdatedAt:    time.Now().Add(-1 * time.Hour),
        },
        {
            ID:           "merchant-2",
            Name:         "Tech Startup Inc",
            BusinessType: "Technology",
            IndustryCode: "7372",
            PortfolioType: "pending",
            RiskLevel:    "medium",
            CreatedAt:    time.Now().Add(-12 * time.Hour),
            UpdatedAt:    time.Now().Add(-30 * time.Minute),
        },
    }
}

func GetTestMerchantRequest() models.MerchantRequest {
    return models.MerchantRequest{
        Name:         "Test Merchant",
        BusinessType: "Services",
        IndustryCode: "5611",
        ContactInfo: &models.ContactInfo{
            Email: "test@merchant.com",
            Phone: "+1-555-123-4567",
        },
        Address: &models.Address{
            Street:  "123 Test St",
            City:    "Test City",
            State:   "CA",
            Zip:     "12345",
            Country: "US",
        },
    }
}
```

### Mock Data Generation

```go
// test/mocks/mock_data_generator.go
package mocks

import (
    "math/rand"
    "time"
    "github.com/kyb-platform/internal/models"
)

type MockDataGenerator struct {
    rand *rand.Rand
}

func NewMockDataGenerator() *MockDataGenerator {
    return &MockDataGenerator{
        rand: rand.New(rand.NewSource(time.Now().UnixNano())),
    }
}

func (g *MockDataGenerator) GenerateMerchants(count int) []models.Merchant {
    merchants := make([]models.Merchant, count)
    
    businessTypes := []string{"Retail", "Technology", "Services", "Manufacturing", "Healthcare"}
    portfolioTypes := []string{"onboarded", "pending", "prospective", "deactivated"}
    riskLevels := []string{"low", "medium", "high"}
    
    for i := 0; i < count; i++ {
        merchants[i] = models.Merchant{
            ID:           g.generateID(),
            Name:         g.generateBusinessName(),
            BusinessType: businessTypes[g.rand.Intn(len(businessTypes))],
            IndustryCode: g.generateIndustryCode(),
            PortfolioType: portfolioTypes[g.rand.Intn(len(portfolioTypes))],
            RiskLevel:    riskLevels[g.rand.Intn(len(riskLevels))],
            CreatedAt:    time.Now().Add(-time.Duration(g.rand.Intn(365)) * 24 * time.Hour),
            UpdatedAt:    time.Now().Add(-time.Duration(g.rand.Intn(30)) * time.Hour),
        }
    }
    
    return merchants
}

func (g *MockDataGenerator) generateID() string {
    return fmt.Sprintf("merchant-%d", g.rand.Intn(1000000))
}

func (g *MockDataGenerator) generateBusinessName() string {
    prefixes := []string{"Acme", "Global", "Premier", "Elite", "Advanced"}
    suffixes := []string{"Corp", "Inc", "LLC", "Ltd", "Group"}
    
    return fmt.Sprintf("%s %s %s",
        prefixes[g.rand.Intn(len(prefixes))],
        g.generateRandomWord(),
        suffixes[g.rand.Intn(len(suffixes))],
    )
}

func (g *MockDataGenerator) generateIndustryCode() string {
    codes := []string{"5411", "7372", "5611", "3111", "6211"}
    return codes[g.rand.Intn(len(codes))]
}
```

## Test Configuration

### Test Environment Setup

```go
// test/setup.go
package test

import (
    "database/sql"
    "testing"
    "github.com/kyb-platform/internal/database"
    "github.com/kyb-platform/internal/config"
)

func setupTestDB(t *testing.T) *sql.DB {
    // Create test database
    testConfig := &config.Config{
        Database: config.DatabaseConfig{
            Host:     "localhost",
            Port:     5432,
            User:     "kyb_test",
            Password: "test_password",
            Name:     "kyb_test",
            SSLMode:  "disable",
        },
    }
    
    db, err := database.Connect(testConfig.Database)
    require.NoError(t, err)
    
    // Run migrations
    err = database.Migrate(db, "up")
    require.NoError(t, err)
    
    return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
    // Clean up test data
    _, err := db.Exec("DELETE FROM merchants")
    require.NoError(t, err)
    
    // Close connection
    err = db.Close()
    require.NoError(t, err)
}

func setupTestServer(t *testing.T) *httptest.Server {
    // Setup test server with mocked dependencies
    mockRepo := &mocks.MerchantRepository{}
    mockService := &mocks.MerchantService{}
    
    handler := NewMerchantHandler(mockService, zap.NewNop())
    
    mux := http.NewServeMux()
    mux.HandleFunc("/api/v1/merchants", handler.GetMerchants)
    
    return httptest.NewServer(mux)
}
```

### Test Configuration Files

```yaml
# test/config/test.yaml
database:
  host: localhost
  port: 5432
  user: kyb_test
  password: test_password
  name: kyb_test
  ssl_mode: disable

redis:
  host: localhost
  port: 6379
  db: 1

logging:
  level: debug
  format: json

api:
  port: 8080
  timeout: 30s
```

## Continuous Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Test Suite

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.22
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run unit tests
      run: go test -v -race -coverprofile=coverage.out ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  integration-tests:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: kyb_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.22
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run integration tests
      run: go test -v -tags=integration ./test/integration/...

  e2e-tests:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
    
    - name: Install dependencies
      run: npm ci
    
    - name: Install Playwright
      run: npx playwright install --with-deps
    
    - name: Run E2E tests
      run: npx playwright test
    
    - name: Upload test results
      uses: actions/upload-artifact@v3
      if: always()
      with:
        name: playwright-report
        path: playwright-report/
```

## Best Practices

### Test Organization

1. **Test Structure**: Follow the Arrange-Act-Assert pattern
2. **Test Naming**: Use descriptive test names that explain the scenario
3. **Test Isolation**: Each test should be independent and not rely on other tests
4. **Test Data**: Use fixtures and factories for consistent test data
5. **Test Cleanup**: Always clean up resources after tests

### Performance Considerations

1. **Parallel Testing**: Use `t.Parallel()` for independent tests
2. **Test Timeouts**: Set appropriate timeouts for long-running tests
3. **Resource Management**: Properly manage database connections and file handles
4. **Memory Usage**: Monitor memory usage in performance tests
5. **Test Data Size**: Use appropriate test data sizes for different test types

### Maintenance

1. **Regular Updates**: Keep test dependencies up to date
2. **Test Coverage**: Maintain high test coverage (aim for 90%+)
3. **Test Documentation**: Document complex test scenarios
4. **Test Reviews**: Include tests in code reviews
5. **Test Metrics**: Track test execution time and failure rates

## Conclusion

This testing guide provides comprehensive strategies for testing the KYB Platform. Follow these guidelines to ensure robust, reliable, and maintainable tests that support the platform's quality and reliability goals.

For additional information, refer to the architecture documentation and API development guide.
