package services

import (
	"context"
	"fmt"
	"log"
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
	logger        *log.Logger
}

// NewMerchantAnalyticsService creates a new merchant analytics service
func NewMerchantAnalyticsService(
	analyticsRepo *database.MerchantAnalyticsRepository,
	merchantRepo *database.MerchantPortfolioRepository,
	logger *log.Logger,
) MerchantAnalyticsService {
	if logger == nil {
		logger = log.Default()
	}

	return &merchantAnalyticsService{
		analyticsRepo: analyticsRepo,
		merchantRepo:  merchantRepo,
		logger:        logger,
	}
}

// GetMerchantAnalytics retrieves comprehensive analytics data for a merchant
func (s *merchantAnalyticsService) GetMerchantAnalytics(ctx context.Context, merchantID string) (*models.AnalyticsData, error) {
	s.logger.Printf("Getting analytics for merchant: %s", merchantID)

	// Verify merchant exists
	_, err := s.merchantRepo.GetMerchant(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}

	// Get classification data
	classification, err := s.analyticsRepo.GetClassificationByMerchantID(ctx, merchantID)
	if err != nil {
		s.logger.Printf("Warning: failed to get classification data: %v", err)
		// Use default classification if not found
		classification = &models.ClassificationData{
			PrimaryIndustry: "",
			ConfidenceScore: 0.0,
			RiskLevel:       "medium",
			MCCCodes:        []models.IndustryCode{},
			SICCodes:        []models.IndustryCode{},
			NAICSCodes:      []models.IndustryCode{},
		}
	}

	// Get security data
	security, err := s.analyticsRepo.GetSecurityDataByMerchantID(ctx, merchantID)
	if err != nil {
		s.logger.Printf("Warning: failed to get security data: %v", err)
		// Use default security if not found
		security = &models.SecurityData{
			TrustScore:      0.5,
			SSLValid:        false,
			SecurityHeaders: []models.SecurityHeader{},
		}
	}

	// Get quality metrics
	quality, err := s.analyticsRepo.GetQualityMetricsByMerchantID(ctx, merchantID)
	if err != nil {
		s.logger.Printf("Warning: failed to get quality metrics: %v", err)
		// Use default quality if not found
		quality = &models.QualityData{
			CompletenessScore: 0.0,
			DataPoints:        0,
			MissingFields:     []string{},
		}
	}

	// Get intelligence data
	intelligence, err := s.analyticsRepo.GetIntelligenceDataByMerchantID(ctx, merchantID)
	if err != nil {
		s.logger.Printf("Warning: failed to get intelligence data: %v", err)
		intelligence = &models.IntelligenceData{}
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

