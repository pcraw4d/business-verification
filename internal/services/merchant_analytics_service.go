package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"kyb-platform/internal/database"
	"kyb-platform/internal/models"
)

// MerchantAnalyticsService provides business logic for merchant analytics
type MerchantAnalyticsService interface {
	GetMerchantAnalytics(ctx context.Context, merchantID string) (*models.AnalyticsData, error)
	GetWebsiteAnalysis(ctx context.Context, merchantID string) (*models.WebsiteAnalysisData, error)
}

// merchantAnalyticsService implements MerchantAnalyticsService
type merchantAnalyticsService struct {
	analyticsRepo *database.MerchantAnalyticsRepository
	merchantRepo  *database.MerchantPortfolioRepository
	cache         Cache
	logger        *log.Logger
}

// Cache interface for dependency injection
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

// NewMerchantAnalyticsService creates a new merchant analytics service
func NewMerchantAnalyticsService(
	analyticsRepo *database.MerchantAnalyticsRepository,
	merchantRepo *database.MerchantPortfolioRepository,
	cache Cache,
	logger *log.Logger,
) MerchantAnalyticsService {
	if logger == nil {
		logger = log.Default()
	}

	return &merchantAnalyticsService{
		analyticsRepo: analyticsRepo,
		merchantRepo:  merchantRepo,
		cache:         cache,
		logger:        logger,
	}
}

// GetMerchantAnalytics retrieves comprehensive analytics data for a merchant
func (s *merchantAnalyticsService) GetMerchantAnalytics(ctx context.Context, merchantID string) (*models.AnalyticsData, error) {
	s.logger.Printf("Getting analytics for merchant: %s", merchantID)

	// Add timeout context
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Verify merchant exists and is active
	merchant, err := s.merchantRepo.GetMerchant(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}

	if merchant.Status != "active" {
		return nil, fmt.Errorf("merchant is not active")
	}

	// Check cache first
	if s.cache != nil {
		cacheKey := fmt.Sprintf("analytics:%s", merchantID)
		var cachedData models.AnalyticsData
		if err := s.cache.Get(ctx, cacheKey, &cachedData); err == nil {
			s.logger.Printf("Cache hit for merchant analytics: %s", merchantID)
			return &cachedData, nil
		}
	}

	// Fetch data in parallel using goroutines
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	var classification *models.ClassificationData
	var security *models.SecurityData
	var quality *models.QualityData
	var intelligence *models.IntelligenceData

	// Fetch classification data (critical)
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, err := s.analyticsRepo.GetClassificationByMerchantID(ctx, merchantID)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, fmt.Errorf("classification: %w", err))
			// Use default classification if not found
			classification = &models.ClassificationData{
				PrimaryIndustry: "",
				ConfidenceScore: 0.0,
				RiskLevel:       "medium",
				MCCCodes:        []models.IndustryCode{},
				SICCodes:        []models.IndustryCode{},
				NAICSCodes:      []models.IndustryCode{},
			}
		} else {
			classification = data
		}
	}()

	// Fetch security data
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, err := s.analyticsRepo.GetSecurityDataByMerchantID(ctx, merchantID)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, fmt.Errorf("security: %w", err))
			security = &models.SecurityData{
				TrustScore:      0.5,
				SSLValid:        false,
				SecurityHeaders: []models.SecurityHeader{},
			}
		} else {
			security = data
		}
	}()

	// Fetch quality metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, err := s.analyticsRepo.GetQualityMetricsByMerchantID(ctx, merchantID)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, fmt.Errorf("quality: %w", err))
			quality = &models.QualityData{
				CompletenessScore: 0.0,
				DataPoints:        0,
				MissingFields:     []string{},
			}
		} else {
			quality = data
		}
	}()

	// Fetch intelligence data
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, err := s.analyticsRepo.GetIntelligenceDataByMerchantID(ctx, merchantID)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, fmt.Errorf("intelligence: %w", err))
			intelligence = &models.IntelligenceData{}
		} else {
			intelligence = data
		}
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// If critical data (classification) is nil, return error
	if classification == nil {
		return nil, fmt.Errorf("failed to fetch analytics: %v", errs)
	}

	// Assemble analytics data
	analytics := &models.AnalyticsData{
		MerchantID:     merchantID,
		Classification: *classification,
		Security:       *security,
		Quality:        *quality,
		Intelligence:   *intelligence,
		Timestamp:      time.Now(),
	}

	// Cache result (5 minute TTL)
	if s.cache != nil {
		cacheKey := fmt.Sprintf("analytics:%s", merchantID)
		if err := s.cache.Set(ctx, cacheKey, analytics, 5*time.Minute); err != nil {
			s.logger.Printf("Warning: failed to cache analytics data: %v", err)
		}
	}

	return analytics, nil
}

// GetWebsiteAnalysis retrieves website analysis data for a merchant
func (s *merchantAnalyticsService) GetWebsiteAnalysis(ctx context.Context, merchantID string) (*models.WebsiteAnalysisData, error) {
	s.logger.Printf("Getting website analysis for merchant: %s", merchantID)

	// Verify merchant exists
	merchant, err := s.merchantRepo.GetMerchant(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}

	websiteURL := merchant.ContactInfo.Website
	if websiteURL == "" {
		return nil, fmt.Errorf("merchant has no website URL")
	}

	// Get security data (includes SSL info)
	security, err := s.analyticsRepo.GetSecurityDataByMerchantID(ctx, merchantID)
	if err != nil {
		s.logger.Printf("Warning: failed to get security data: %v", err)
		security = &models.SecurityData{
			TrustScore:      0.5,
			SSLValid:        false,
			SecurityHeaders: []models.SecurityHeader{},
		}
	}

	// Build SSL data from security data
	sslData := models.SSLData{
		Valid: security.SSLValid,
	}
	if security.SSLExpiryDate != nil {
		sslData.ExpiryDate = security.SSLExpiryDate
		sslData.Issuer = "Unknown" // Would be populated from actual SSL check
		sslData.Grade = "A"        // Would be calculated from SSL check
	}

	// For MVP, use default performance and accessibility data
	// In production, these would come from actual website analysis
	performance := models.PerformanceData{
		LoadTime: 1.5,  // Default values
		PageSize: 1024000,
		Requests: 50,
		Score:    80,
	}

	accessibility := models.AccessibilityData{
		Score:  0.85, // Default score
		Issues: []string{},
	}

	analysis := &models.WebsiteAnalysisData{
		MerchantID:      merchantID,
		WebsiteURL:      websiteURL,
		SSL:             sslData,
		SecurityHeaders: security.SecurityHeaders,
		Performance:     performance,
		Accessibility:   accessibility,
		LastAnalyzed:    time.Now(),
	}

	return analysis, nil
}

