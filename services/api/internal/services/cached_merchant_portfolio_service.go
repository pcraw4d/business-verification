package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"kyb-platform/internal/cache"

	"go.uber.org/zap"
)

// MerchantPortfolioServiceInterface defines the interface for merchant portfolio service
type MerchantPortfolioServiceInterface interface {
	CreateMerchant(ctx context.Context, merchant *Merchant, userID string) (*Merchant, error)
	GetMerchant(ctx context.Context, merchantID string) (*Merchant, error)
	UpdateMerchant(ctx context.Context, merchant *Merchant, userID string) (*Merchant, error)
	DeleteMerchant(ctx context.Context, merchantID string, userID string) error
	SearchMerchants(ctx context.Context, filters *MerchantSearchFilters, page, pageSize int) (*MerchantListResult, error)
	BulkUpdatePortfolioType(ctx context.Context, merchantIDs []string, portfolioType PortfolioType, userID string) (*BulkOperationResult, error)
	BulkUpdateRiskLevel(ctx context.Context, merchantIDs []string, riskLevel RiskLevel, userID string) (*BulkOperationResult, error)
	StartMerchantSession(ctx context.Context, userID, merchantID string) (*MerchantSession, error)
	EndMerchantSession(ctx context.Context, userID string) error
	GetActiveMerchantSession(ctx context.Context, userID string) (*MerchantSession, error)
}

// CachedMerchantPortfolioService wraps the MerchantPortfolioService with caching
type CachedMerchantPortfolioService struct {
	underlyingService MerchantPortfolioServiceInterface
	cacheService      *cache.MerchantCacheService
	logger            *log.Logger
	zapLogger         *zap.Logger
}

// NewCachedMerchantPortfolioService creates a new cached merchant portfolio service
func NewCachedMerchantPortfolioService(
	underlyingService MerchantPortfolioServiceInterface,
	cacheService *cache.MerchantCacheService,
	logger *log.Logger,
	zapLogger *zap.Logger,
) *CachedMerchantPortfolioService {
	if logger == nil {
		logger = log.Default()
	}
	if zapLogger == nil {
		zapLogger = zap.NewNop()
	}

	return &CachedMerchantPortfolioService{
		underlyingService: underlyingService,
		cacheService:      cacheService,
		logger:            logger,
		zapLogger:         zapLogger,
	}
}

// CreateMerchant creates a new merchant and caches it
func (s *CachedMerchantPortfolioService) CreateMerchant(ctx context.Context, merchant *Merchant, userID string) (*Merchant, error) {
	s.logger.Printf("Creating merchant with caching: %s", merchant.Name)

	// Create merchant using underlying service
	createdMerchant, err := s.underlyingService.CreateMerchant(ctx, merchant, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create merchant: %w", err)
	}

	// Cache the created merchant
	if err := s.cacheMerchantDetail(ctx, createdMerchant); err != nil {
		s.logger.Printf("Warning: failed to cache created merchant: %v", err)
	}

	// Invalidate related caches
	if err := s.invalidateRelatedCaches(ctx, createdMerchant.ID); err != nil {
		s.logger.Printf("Warning: failed to invalidate related caches: %v", err)
	}

	s.logger.Printf("Successfully created and cached merchant: %s (ID: %s)", createdMerchant.Name, createdMerchant.ID)
	return createdMerchant, nil
}

// GetMerchant retrieves a merchant by ID, checking cache first
func (s *CachedMerchantPortfolioService) GetMerchant(ctx context.Context, merchantID string) (*Merchant, error) {
	s.logger.Printf("Retrieving merchant with caching: %s", merchantID)

	// Try to get from cache first
	var cachedMerchant map[string]interface{}
	found, err := s.cacheService.GetMerchantDetail(ctx, merchantID, &cachedMerchant)
	if err != nil {
		s.logger.Printf("Warning: failed to check cache for merchant %s: %v", merchantID, err)
	} else if found {
		// Convert cached data back to Merchant struct
		merchant, err := s.mapToMerchant(cachedMerchant)
		if err != nil {
			s.logger.Printf("Warning: failed to convert cached data for merchant %s: %v", merchantID, err)
		} else {
			s.logger.Printf("Retrieved merchant from cache: %s", merchantID)
			return merchant, nil
		}
	}

	// Cache miss - get from underlying service
	s.logger.Printf("Cache miss for merchant %s, fetching from database", merchantID)
	merchant, err := s.underlyingService.GetMerchant(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve merchant from database: %w", err)
	}

	// Cache the retrieved merchant
	if err := s.cacheMerchantDetail(ctx, merchant); err != nil {
		s.logger.Printf("Warning: failed to cache retrieved merchant: %v", err)
	}

	s.logger.Printf("Retrieved and cached merchant: %s", merchantID)
	return merchant, nil
}

// UpdateMerchant updates an existing merchant and invalidates cache
func (s *CachedMerchantPortfolioService) UpdateMerchant(ctx context.Context, merchant *Merchant, userID string) (*Merchant, error) {
	s.logger.Printf("Updating merchant with cache invalidation: %s", merchant.ID)

	// Update merchant using underlying service
	updatedMerchant, err := s.underlyingService.UpdateMerchant(ctx, merchant, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update merchant: %w", err)
	}

	// Invalidate and refresh cache
	if err := s.invalidateRelatedCaches(ctx, merchant.ID); err != nil {
		s.logger.Printf("Warning: failed to invalidate related caches: %v", err)
	}

	// Cache the updated merchant
	if err := s.cacheMerchantDetail(ctx, updatedMerchant); err != nil {
		s.logger.Printf("Warning: failed to cache updated merchant: %v", err)
	}

	s.logger.Printf("Successfully updated and cached merchant: %s", merchant.ID)
	return updatedMerchant, nil
}

// DeleteMerchant deletes a merchant and invalidates cache
func (s *CachedMerchantPortfolioService) DeleteMerchant(ctx context.Context, merchantID string, userID string) error {
	s.logger.Printf("Deleting merchant with cache invalidation: %s", merchantID)

	// Delete merchant using underlying service
	if err := s.underlyingService.DeleteMerchant(ctx, merchantID, userID); err != nil {
		return fmt.Errorf("failed to delete merchant: %w", err)
	}

	// Invalidate all related caches
	if err := s.invalidateRelatedCaches(ctx, merchantID); err != nil {
		s.logger.Printf("Warning: failed to invalidate related caches: %v", err)
	}

	s.logger.Printf("Successfully deleted merchant and invalidated cache: %s", merchantID)
	return nil
}

// SearchMerchants searches merchants with caching
func (s *CachedMerchantPortfolioService) SearchMerchants(ctx context.Context, filters *MerchantSearchFilters, page, pageSize int) (*MerchantListResult, error) {
	s.logger.Printf("Searching merchants with caching (page: %d, size: %d)", page, pageSize)

	// Create cache key for search
	cacheKey := s.generateSearchCacheKey(filters, page, pageSize)

	// Try to get from cache first
	var cachedResults []map[string]interface{}
	filtersMap := s.filtersToMap(filters)
	found, err := s.cacheService.GetMerchantSearch(ctx, cacheKey, filtersMap, &cachedResults)
	if err != nil {
		s.logger.Printf("Warning: failed to check cache for search: %v", err)
	} else if found {
		// Convert cached data back to MerchantListResult
		result, err := s.mapToMerchantListResult(cachedResults, page, pageSize)
		if err != nil {
			s.logger.Printf("Warning: failed to convert cached search results: %v", err)
		} else {
			s.logger.Printf("Retrieved search results from cache")
			return result, nil
		}
	}

	// Cache miss - search using underlying service
	s.logger.Printf("Cache miss for search, fetching from database")
	result, err := s.underlyingService.SearchMerchants(ctx, filters, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to search merchants: %w", err)
	}

	// Cache the search results
	if err := s.cacheSearchResults(ctx, cacheKey, filters, result); err != nil {
		s.logger.Printf("Warning: failed to cache search results: %v", err)
	}

	s.logger.Printf("Retrieved and cached search results: %d merchants found", len(result.Merchants))
	return result, nil
}

// BulkUpdatePortfolioType updates portfolio type for multiple merchants with cache invalidation
func (s *CachedMerchantPortfolioService) BulkUpdatePortfolioType(ctx context.Context, merchantIDs []string, portfolioType PortfolioType, userID string) (*BulkOperationResult, error) {
	s.logger.Printf("Bulk updating portfolio type with cache invalidation: %d merchants", len(merchantIDs))

	// Perform bulk update using underlying service
	result, err := s.underlyingService.BulkUpdatePortfolioType(ctx, merchantIDs, portfolioType, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk update portfolio type: %w", err)
	}

	// Invalidate caches for all affected merchants
	for _, merchantID := range merchantIDs {
		if err := s.invalidateRelatedCaches(ctx, merchantID); err != nil {
			s.logger.Printf("Warning: failed to invalidate cache for merchant %s: %v", merchantID, err)
		}
	}

	s.logger.Printf("Successfully bulk updated portfolio type and invalidated caches")
	return result, nil
}

// BulkUpdateRiskLevel updates risk level for multiple merchants with cache invalidation
func (s *CachedMerchantPortfolioService) BulkUpdateRiskLevel(ctx context.Context, merchantIDs []string, riskLevel RiskLevel, userID string) (*BulkOperationResult, error) {
	s.logger.Printf("Bulk updating risk level with cache invalidation: %d merchants", len(merchantIDs))

	// Perform bulk update using underlying service
	result, err := s.underlyingService.BulkUpdateRiskLevel(ctx, merchantIDs, riskLevel, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk update risk level: %w", err)
	}

	// Invalidate caches for all affected merchants
	for _, merchantID := range merchantIDs {
		if err := s.invalidateRelatedCaches(ctx, merchantID); err != nil {
			s.logger.Printf("Warning: failed to invalidate cache for merchant %s: %v", merchantID, err)
		}
	}

	s.logger.Printf("Successfully bulk updated risk level and invalidated caches")
	return result, nil
}

// StartMerchantSession starts a merchant session (no caching needed)
func (s *CachedMerchantPortfolioService) StartMerchantSession(ctx context.Context, userID, merchantID string) (*MerchantSession, error) {
	return s.underlyingService.StartMerchantSession(ctx, userID, merchantID)
}

// EndMerchantSession ends a merchant session (no caching needed)
func (s *CachedMerchantPortfolioService) EndMerchantSession(ctx context.Context, userID string) error {
	return s.underlyingService.EndMerchantSession(ctx, userID)
}

// GetActiveMerchantSession gets the active merchant session (no caching needed)
func (s *CachedMerchantPortfolioService) GetActiveMerchantSession(ctx context.Context, userID string) (*MerchantSession, error) {
	return s.underlyingService.GetActiveMerchantSession(ctx, userID)
}

// =============================================================================
// Helper Methods
// =============================================================================

// cacheMerchantDetail caches a merchant detail
func (s *CachedMerchantPortfolioService) cacheMerchantDetail(ctx context.Context, merchant *Merchant) error {
	merchantMap := s.merchantToMap(merchant)
	return s.cacheService.CacheMerchantDetail(ctx, merchant.ID, merchantMap, 30*time.Minute)
}

// invalidateRelatedCaches invalidates all caches related to a merchant
func (s *CachedMerchantPortfolioService) invalidateRelatedCaches(ctx context.Context, merchantID string) error {
	return s.cacheService.InvalidateMerchant(ctx, merchantID)
}

// generateSearchCacheKey generates a cache key for search results
func (s *CachedMerchantPortfolioService) generateSearchCacheKey(filters *MerchantSearchFilters, page, pageSize int) string {
	// Create a simple cache key based on filters and pagination
	key := fmt.Sprintf("search:page_%d_size_%d", page, pageSize)
	if filters != nil {
		if filters.PortfolioType != nil {
			key += fmt.Sprintf(":portfolio_%s", *filters.PortfolioType)
		}
		if filters.RiskLevel != nil {
			key += fmt.Sprintf(":risk_%s", *filters.RiskLevel)
		}
		if filters.Industry != "" {
			key += fmt.Sprintf(":industry_%s", filters.Industry)
		}
		if filters.Status != "" {
			key += fmt.Sprintf(":status_%s", filters.Status)
		}
	}
	return key
}

// cacheSearchResults caches search results
func (s *CachedMerchantPortfolioService) cacheSearchResults(ctx context.Context, cacheKey string, filters *MerchantSearchFilters, result *MerchantListResult) error {
	// Convert merchants to maps for caching
	merchantMaps := make([]map[string]interface{}, len(result.Merchants))
	for i, merchant := range result.Merchants {
		merchantMaps[i] = s.merchantToMap(merchant)
	}

	filtersMap := s.filtersToMap(filters)
	return s.cacheService.CacheMerchantSearch(ctx, cacheKey, filtersMap, merchantMaps, 15*time.Minute)
}

// filtersToMap converts MerchantSearchFilters to a map for caching
func (s *CachedMerchantPortfolioService) filtersToMap(filters *MerchantSearchFilters) map[string]interface{} {
	if filters == nil {
		return nil
	}

	filtersMap := make(map[string]interface{})
	if filters.PortfolioType != nil {
		filtersMap["portfolio_type"] = *filters.PortfolioType
	}
	if filters.RiskLevel != nil {
		filtersMap["risk_level"] = *filters.RiskLevel
	}
	if filters.Industry != "" {
		filtersMap["industry"] = filters.Industry
	}
	if filters.Status != "" {
		filtersMap["status"] = filters.Status
	}
	if filters.SearchQuery != "" {
		filtersMap["search_query"] = filters.SearchQuery
	}

	return filtersMap
}

// merchantToMap converts a Merchant to a map for caching
func (s *CachedMerchantPortfolioService) merchantToMap(merchant *Merchant) map[string]interface{} {
	// Convert to JSON and back to map to ensure proper serialization
	jsonData, _ := json.Marshal(merchant)
	var merchantMap map[string]interface{}
	json.Unmarshal(jsonData, &merchantMap)
	return merchantMap
}

// mapToMerchant converts a map back to a Merchant struct
func (s *CachedMerchantPortfolioService) mapToMerchant(merchantMap map[string]interface{}) (*Merchant, error) {
	jsonData, err := json.Marshal(merchantMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal merchant map: %w", err)
	}

	var merchant Merchant
	if err := json.Unmarshal(jsonData, &merchant); err != nil {
		return nil, fmt.Errorf("failed to unmarshal merchant: %w", err)
	}

	return &merchant, nil
}

// mapToMerchantListResult converts cached search results back to MerchantListResult
func (s *CachedMerchantPortfolioService) mapToMerchantListResult(merchantMaps []map[string]interface{}, page, pageSize int) (*MerchantListResult, error) {
	merchants := make([]*Merchant, len(merchantMaps))
	for i, merchantMap := range merchantMaps {
		merchant, err := s.mapToMerchant(merchantMap)
		if err != nil {
			return nil, fmt.Errorf("failed to convert merchant map at index %d: %w", i, err)
		}
		merchants[i] = merchant
	}

	return &MerchantListResult{
		Merchants: merchants,
		Total:     len(merchants),
		Page:      page,
		PageSize:  pageSize,
		HasMore:   false, // We don't cache the hasMore flag, so assume false
	}, nil
}
