package multi_site_aggregation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// =============================================================================
// In-Memory Storage Implementations
// =============================================================================

// InMemoryLocationStore implements LocationStore interface with in-memory storage
type InMemoryLocationStore struct {
	locations map[string]BusinessLocation
	mutex     sync.RWMutex
}

// NewInMemoryLocationStore creates a new in-memory location store
func NewInMemoryLocationStore() *InMemoryLocationStore {
	return &InMemoryLocationStore{
		locations: make(map[string]BusinessLocation),
	}
}

// GetLocationsByBusinessID retrieves all locations for a business
func (s *InMemoryLocationStore) GetLocationsByBusinessID(ctx context.Context, businessID string) ([]BusinessLocation, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var locations []BusinessLocation
	for _, location := range s.locations {
		if location.BusinessID == businessID {
			locations = append(locations, location)
		}
	}

	return locations, nil
}

// SaveLocation saves a new location
func (s *InMemoryLocationStore) SaveLocation(ctx context.Context, location *BusinessLocation) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if location.ID == "" {
		return fmt.Errorf("location ID is required")
	}

	if _, exists := s.locations[location.ID]; exists {
		return fmt.Errorf("location with ID %s already exists", location.ID)
	}

	s.locations[location.ID] = *location
	return nil
}

// UpdateLocation updates an existing location
func (s *InMemoryLocationStore) UpdateLocation(ctx context.Context, location *BusinessLocation) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if location.ID == "" {
		return fmt.Errorf("location ID is required")
	}

	if _, exists := s.locations[location.ID]; !exists {
		return fmt.Errorf("location with ID %s not found", location.ID)
	}

	location.UpdatedAt = time.Now()
	s.locations[location.ID] = *location
	return nil
}

// DeleteLocation deletes a location
func (s *InMemoryLocationStore) DeleteLocation(ctx context.Context, locationID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.locations[locationID]; !exists {
		return fmt.Errorf("location with ID %s not found", locationID)
	}

	delete(s.locations, locationID)
	return nil
}

// =============================================================================
// In-Memory Data Store Implementation
// =============================================================================

// InMemoryDataStore implements DataStore interface with in-memory storage
type InMemoryDataStore struct {
	siteData map[string]SiteData
	mutex    sync.RWMutex
}

// NewInMemoryDataStore creates a new in-memory data store
func NewInMemoryDataStore() *InMemoryDataStore {
	return &InMemoryDataStore{
		siteData: make(map[string]SiteData),
	}
}

// SaveSiteData saves site data
func (s *InMemoryDataStore) SaveSiteData(ctx context.Context, data *SiteData) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if data.ID == "" {
		return fmt.Errorf("site data ID is required")
	}

	if _, exists := s.siteData[data.ID]; exists {
		return fmt.Errorf("site data with ID %s already exists", data.ID)
	}

	s.siteData[data.ID] = *data
	return nil
}

// GetSiteDataByLocationID retrieves site data by location ID
func (s *InMemoryDataStore) GetSiteDataByLocationID(ctx context.Context, locationID string) ([]SiteData, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var data []SiteData
	for _, siteData := range s.siteData {
		if siteData.LocationID == locationID {
			data = append(data, siteData)
		}
	}

	return data, nil
}

// GetSiteDataByBusinessID retrieves site data by business ID
func (s *InMemoryDataStore) GetSiteDataByBusinessID(ctx context.Context, businessID string) ([]SiteData, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var data []SiteData
	for _, siteData := range s.siteData {
		if siteData.BusinessID == businessID {
			data = append(data, siteData)
		}
	}

	return data, nil
}

// DeleteSiteData deletes site data
func (s *InMemoryDataStore) DeleteSiteData(ctx context.Context, dataID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.siteData[dataID]; !exists {
		return fmt.Errorf("site data with ID %s not found", dataID)
	}

	delete(s.siteData, dataID)
	return nil
}

// =============================================================================
// Factory Functions for Easy Service Creation
// =============================================================================

// CreateMultiSiteAggregationService creates a complete multi-site aggregation service with default implementations
func CreateMultiSiteAggregationService(logger *zap.Logger) *MultiSiteDataAggregationService {
	config := DefaultMultiSiteAggregationConfig()
	extractor := NewWebsiteDataExtractor(logger)
	validator := NewWebsiteDataValidator(logger)
	aggregator := NewWebsiteDataAggregator(logger)
	locationStore := NewInMemoryLocationStore()
	dataStore := NewInMemoryDataStore()

	return NewMultiSiteDataAggregationService(
		config,
		logger,
		extractor,
		validator,
		aggregator,
		locationStore,
		dataStore,
	)
}

// CreateMultiSiteAggregationServiceWithConfig creates a multi-site aggregation service with custom configuration
func CreateMultiSiteAggregationServiceWithConfig(
	config *MultiSiteAggregationConfig,
	logger *zap.Logger,
) *MultiSiteDataAggregationService {
	extractor := NewWebsiteDataExtractor(logger)
	validator := NewWebsiteDataValidator(logger)
	aggregator := NewWebsiteDataAggregator(logger)
	locationStore := NewInMemoryLocationStore()
	dataStore := NewInMemoryDataStore()

	return NewMultiSiteDataAggregationService(
		config,
		logger,
		extractor,
		validator,
		aggregator,
		locationStore,
		dataStore,
	)
}
