package services

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/cache"
	"kyb-platform/internal/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockMerchantPortfolioService is a mock implementation for testing
type MockMerchantPortfolioService struct {
	merchants map[string]*Merchant
	errors    map[string]error
}

func NewMockMerchantPortfolioService() *MockMerchantPortfolioService {
	return &MockMerchantPortfolioService{
		merchants: make(map[string]*Merchant),
		errors:    make(map[string]error),
	}
}

func (m *MockMerchantPortfolioService) CreateMerchant(ctx context.Context, merchant *Merchant, userID string) (*Merchant, error) {
	if err, exists := m.errors["create"]; exists {
		return nil, err
	}
	m.merchants[merchant.ID] = merchant
	return merchant, nil
}

func (m *MockMerchantPortfolioService) GetMerchant(ctx context.Context, merchantID string) (*Merchant, error) {
	if err, exists := m.errors["get"]; exists {
		return nil, err
	}
	if merchant, exists := m.merchants[merchantID]; exists {
		return merchant, nil
	}
	return nil, database.ErrMerchantNotFound
}

func (m *MockMerchantPortfolioService) UpdateMerchant(ctx context.Context, merchant *Merchant, userID string) (*Merchant, error) {
	if err, exists := m.errors["update"]; exists {
		return nil, err
	}
	if _, exists := m.merchants[merchant.ID]; !exists {
		return nil, database.ErrMerchantNotFound
	}
	m.merchants[merchant.ID] = merchant
	return merchant, nil
}

func (m *MockMerchantPortfolioService) DeleteMerchant(ctx context.Context, merchantID string, userID string) error {
	if err, exists := m.errors["delete"]; exists {
		return err
	}
	if _, exists := m.merchants[merchantID]; !exists {
		return database.ErrMerchantNotFound
	}
	delete(m.merchants, merchantID)
	return nil
}

func (m *MockMerchantPortfolioService) SearchMerchants(ctx context.Context, filters *MerchantSearchFilters, page, pageSize int) (*MerchantListResult, error) {
	if err, exists := m.errors["search"]; exists {
		return nil, err
	}

	// Simple mock search - return all merchants
	merchants := make([]*Merchant, 0, len(m.merchants))
	for _, merchant := range m.merchants {
		merchants = append(merchants, merchant)
	}

	return &MerchantListResult{
		Merchants: merchants,
		Total:     len(merchants),
		Page:      page,
		PageSize:  pageSize,
		HasMore:   false,
	}, nil
}

func (m *MockMerchantPortfolioService) BulkUpdatePortfolioType(ctx context.Context, merchantIDs []string, portfolioType PortfolioType, userID string) (*BulkOperationResult, error) {
	if err, exists := m.errors["bulk_update"]; exists {
		return nil, err
	}

	successCount := 0
	for _, id := range merchantIDs {
		if merchant, exists := m.merchants[id]; exists {
			merchant.PortfolioType = portfolioType
			successCount++
		}
	}

	return &BulkOperationResult{
		OperationID: "bulk-update-123",
		Status:      "completed",
		TotalItems:  len(merchantIDs),
		Processed:   len(merchantIDs),
		Successful:  successCount,
		Failed:      len(merchantIDs) - successCount,
		Errors:      []string{},
		StartedAt:   time.Now(),
	}, nil
}

func (m *MockMerchantPortfolioService) BulkUpdateRiskLevel(ctx context.Context, merchantIDs []string, riskLevel RiskLevel, userID string) (*BulkOperationResult, error) {
	if err, exists := m.errors["bulk_update"]; exists {
		return nil, err
	}

	successCount := 0
	for _, id := range merchantIDs {
		if merchant, exists := m.merchants[id]; exists {
			merchant.RiskLevel = riskLevel
			successCount++
		}
	}

	return &BulkOperationResult{
		OperationID: "bulk-update-123",
		Status:      "completed",
		TotalItems:  len(merchantIDs),
		Processed:   len(merchantIDs),
		Successful:  successCount,
		Failed:      len(merchantIDs) - successCount,
		Errors:      []string{},
		StartedAt:   time.Now(),
	}, nil
}

func (m *MockMerchantPortfolioService) StartMerchantSession(ctx context.Context, userID, merchantID string) (*MerchantSession, error) {
	return &MerchantSession{
		ID:         "session-123",
		UserID:     userID,
		MerchantID: merchantID,
		StartedAt:  time.Now(),
	}, nil
}

func (m *MockMerchantPortfolioService) EndMerchantSession(ctx context.Context, userID string) error {
	return nil
}

func (m *MockMerchantPortfolioService) GetActiveMerchantSession(ctx context.Context, userID string) (*MerchantSession, error) {
	return nil, nil
}

// TestCachedMerchantPortfolioServiceBasicOperations tests basic operations with caching
func TestCachedMerchantPortfolioServiceBasicOperations(t *testing.T) {
	// Create mock underlying service
	mockService := NewMockMerchantPortfolioService()

	// Create memory cache
	cacheConfig := &cache.CacheConfig{
		Type:            cache.MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         1000,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	merchantCache := cache.NewMemoryCache(cacheConfig)

	// Create merchant cache service
	zapLogger := zap.NewNop()
	cacheService := cache.NewMerchantCacheService(merchantCache, nil, zapLogger)
	defer cacheService.Close() // This will also close the underlying cache

	// Create cached service
	logger := log.New(os.Stdout, "", 0)
	cachedService := NewCachedMerchantPortfolioService(mockService, cacheService, logger, zapLogger)

	ctx := context.Background()
	userID := "user-123"

	// Test CreateMerchant
	merchant := &Merchant{
		ID:            "merchant-123",
		Name:          "Test Merchant",
		PortfolioType: PortfolioTypeProspective,
		RiskLevel:     RiskLevelMedium,
		Status:        "active",
	}

	createdMerchant, err := cachedService.CreateMerchant(ctx, merchant, userID)
	require.NoError(t, err)
	assert.Equal(t, merchant.ID, createdMerchant.ID)
	assert.Equal(t, merchant.Name, createdMerchant.Name)

	// Test GetMerchant (should hit cache)
	retrievedMerchant, err := cachedService.GetMerchant(ctx, merchant.ID)
	require.NoError(t, err)
	assert.Equal(t, merchant.ID, retrievedMerchant.ID)
	assert.Equal(t, merchant.Name, retrievedMerchant.Name)

	// Test UpdateMerchant
	updatedMerchant := *retrievedMerchant
	updatedMerchant.Name = "Updated Merchant"

	updatedResult, err := cachedService.UpdateMerchant(ctx, &updatedMerchant, userID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Merchant", updatedResult.Name)

	// Test SearchMerchants
	portfolioType := PortfolioTypeProspective
	filters := &MerchantSearchFilters{
		PortfolioType: &portfolioType,
	}

	searchResult, err := cachedService.SearchMerchants(ctx, filters, 1, 10)
	require.NoError(t, err)
	assert.Len(t, searchResult.Merchants, 1)
	assert.Equal(t, "Updated Merchant", searchResult.Merchants[0].Name)

	// Test DeleteMerchant
	err = cachedService.DeleteMerchant(ctx, merchant.ID, userID)
	require.NoError(t, err)

	// Verify merchant is deleted
	_, err = cachedService.GetMerchant(ctx, merchant.ID)
	assert.Error(t, err)
}

// TestCachedMerchantPortfolioServiceBulkOperations tests bulk operations with cache invalidation
func TestCachedMerchantPortfolioServiceBulkOperations(t *testing.T) {
	// Create mock underlying service
	mockService := NewMockMerchantPortfolioService()

	// Create memory cache
	cacheConfig := &cache.CacheConfig{
		Type:            cache.MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         1000,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	merchantCache := cache.NewMemoryCache(cacheConfig)

	// Create merchant cache service
	zapLogger := zap.NewNop()
	cacheService := cache.NewMerchantCacheService(merchantCache, nil, zapLogger)
	defer cacheService.Close() // This will also close the underlying cache

	// Create cached service
	logger := log.New(os.Stdout, "", 0)
	cachedService := NewCachedMerchantPortfolioService(mockService, cacheService, logger, zapLogger)

	ctx := context.Background()
	userID := "user-123"

	// Create test merchants
	merchant1 := &Merchant{
		ID:            "merchant-1",
		Name:          "Merchant 1",
		PortfolioType: PortfolioTypeProspective,
		RiskLevel:     RiskLevelMedium,
		Status:        "active",
	}
	merchant2 := &Merchant{
		ID:            "merchant-2",
		Name:          "Merchant 2",
		PortfolioType: PortfolioTypeProspective,
		RiskLevel:     RiskLevelMedium,
		Status:        "active",
	}

	// Create merchants
	_, err := cachedService.CreateMerchant(ctx, merchant1, userID)
	require.NoError(t, err)
	_, err = cachedService.CreateMerchant(ctx, merchant2, userID)
	require.NoError(t, err)

	// Test bulk update portfolio type
	merchantIDs := []string{merchant1.ID, merchant2.ID}
	result, err := cachedService.BulkUpdatePortfolioType(ctx, merchantIDs, PortfolioTypeOnboarded, userID)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Successful)
	assert.Equal(t, 0, result.Failed)

	// Test bulk update risk level
	result, err = cachedService.BulkUpdateRiskLevel(ctx, merchantIDs, RiskLevelHigh, userID)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Successful)
	assert.Equal(t, 0, result.Failed)
}

// TestCachedMerchantPortfolioServiceErrorHandling tests error handling
func TestCachedMerchantPortfolioServiceErrorHandling(t *testing.T) {
	// Create mock underlying service with errors
	mockService := NewMockMerchantPortfolioService()
	mockService.errors["get"] = database.ErrMerchantNotFound

	// Create memory cache
	cacheConfig := &cache.CacheConfig{
		Type:            cache.MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         1000,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	merchantCache := cache.NewMemoryCache(cacheConfig)

	// Create merchant cache service
	zapLogger := zap.NewNop()
	cacheService := cache.NewMerchantCacheService(merchantCache, nil, zapLogger)
	defer cacheService.Close() // This will also close the underlying cache

	// Create cached service
	logger := log.New(os.Stdout, "", 0)
	cachedService := NewCachedMerchantPortfolioService(mockService, cacheService, logger, zapLogger)

	ctx := context.Background()

	// Test GetMerchant with error
	_, err := cachedService.GetMerchant(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to retrieve merchant from database")
}

// TestCachedMerchantPortfolioServiceSessionManagement tests session management (no caching)
func TestCachedMerchantPortfolioServiceSessionManagement(t *testing.T) {
	// Create mock underlying service
	mockService := NewMockMerchantPortfolioService()

	// Create memory cache
	cacheConfig := &cache.CacheConfig{
		Type:            cache.MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         1000,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	merchantCache := cache.NewMemoryCache(cacheConfig)

	// Create merchant cache service
	zapLogger := zap.NewNop()
	cacheService := cache.NewMerchantCacheService(merchantCache, nil, zapLogger)
	defer cacheService.Close() // This will also close the underlying cache

	// Create cached service
	logger := log.New(os.Stdout, "", 0)
	cachedService := NewCachedMerchantPortfolioService(mockService, cacheService, logger, zapLogger)

	ctx := context.Background()
	userID := "user-123"
	merchantID := "merchant-123"

	// Test StartMerchantSession
	session, err := cachedService.StartMerchantSession(ctx, userID, merchantID)
	require.NoError(t, err)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, merchantID, session.MerchantID)

	// Test EndMerchantSession
	err = cachedService.EndMerchantSession(ctx, userID)
	require.NoError(t, err)

	// Test GetActiveMerchantSession
	activeSession, err := cachedService.GetActiveMerchantSession(ctx, userID)
	require.NoError(t, err)
	assert.Nil(t, activeSession)
}
