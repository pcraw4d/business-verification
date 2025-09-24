package risk_assessment

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ReputationAnalyzer provides online reputation analysis capabilities
type ReputationAnalyzer struct {
	config *RiskAssessmentConfig
	logger *zap.Logger
}

// ReputationAnalysisResult contains comprehensive reputation analysis results
type ReputationAnalysisResult struct {
	BusinessName        string               `json:"business_name"`
	SocialMediaPresence *SocialMediaPresence `json:"social_media_presence,omitempty"`
	ReviewAnalysis      *ReviewAnalysis      `json:"review_analysis,omitempty"`
	SentimentAnalysis   *SentimentAnalysis   `json:"sentiment_analysis,omitempty"`
	BrandMentions       *BrandMentions       `json:"brand_mentions,omitempty"`
	OverallScore        float64              `json:"overall_score"`
	RiskFactors         []RiskFactor         `json:"risk_factors"`
	Recommendations     []string             `json:"recommendations"`
	LastUpdated         time.Time            `json:"last_updated"`
}

// SocialMediaPresence contains social media analysis results
type SocialMediaPresence struct {
	Platforms        map[string]*SocialMediaPlatform `json:"platforms"`
	TotalFollowers   int64                           `json:"total_followers"`
	TotalPosts       int64                           `json:"total_posts"`
	EngagementRate   float64                         `json:"engagement_rate"`
	ActivityScore    float64                         `json:"activity_score"`
	PresenceScore    float64                         `json:"presence_score"`
	LastActivityDate time.Time                       `json:"last_activity_date"`
	IsActive         bool                            `json:"is_active"`
}

// SocialMediaPlatform contains platform-specific data
type SocialMediaPlatform struct {
	PlatformName   string    `json:"platform_name"`
	ProfileURL     string    `json:"profile_url"`
	Followers      int64     `json:"followers"`
	Following      int64     `json:"following"`
	Posts          int64     `json:"posts"`
	EngagementRate float64   `json:"engagement_rate"`
	LastPostDate   time.Time `json:"last_post_date"`
	IsVerified     bool      `json:"is_verified"`
	IsActive       bool      `json:"is_active"`
	ActivityScore  float64   `json:"activity_score"`
	PresenceScore  float64   `json:"presence_score"`
}

// ReviewAnalysis contains review and rating analysis
type ReviewAnalysis struct {
	Platforms       map[string]*ReviewPlatform `json:"platforms"`
	OverallRating   float64                    `json:"overall_rating"`
	TotalReviews    int64                      `json:"total_reviews"`
	PositiveReviews int64                      `json:"positive_reviews"`
	NegativeReviews int64                      `json:"negative_reviews"`
	NeutralReviews  int64                      `json:"neutral_reviews"`
	RatingTrend     string                     `json:"rating_trend"`
	ReviewVelocity  float64                    `json:"review_velocity"`
	ReviewScore     float64                    `json:"review_score"`
	LastReviewDate  time.Time                  `json:"last_review_date"`
}

// ReviewPlatform contains platform-specific review data
type ReviewPlatform struct {
	PlatformName    string    `json:"platform_name"`
	Rating          float64   `json:"rating"`
	TotalReviews    int64     `json:"total_reviews"`
	PositiveReviews int64     `json:"positive_reviews"`
	NegativeReviews int64     `json:"negative_reviews"`
	NeutralReviews  int64     `json:"neutral_reviews"`
	LastReviewDate  time.Time `json:"last_review_date"`
	ReviewVelocity  float64   `json:"review_velocity"`
}

// SentimentAnalysis contains sentiment analysis results
type SentimentAnalysis struct {
	OverallSentiment string                    `json:"overall_sentiment"`
	SentimentScore   float64                   `json:"sentiment_score"`
	PositiveMentions int64                     `json:"positive_mentions"`
	NegativeMentions int64                     `json:"negative_mentions"`
	NeutralMentions  int64                     `json:"neutral_mentions"`
	TotalMentions    int64                     `json:"total_mentions"`
	SentimentTrend   string                    `json:"sentiment_trend"`
	TopKeywords      []SentimentKeyword        `json:"top_keywords"`
	SourceBreakdown  map[string]*SentimentData `json:"source_breakdown"`
}

// SentimentKeyword contains keyword sentiment data
type SentimentKeyword struct {
	Keyword        string  `json:"keyword"`
	Frequency      int64   `json:"frequency"`
	SentimentScore float64 `json:"sentiment_score"`
	Sentiment      string  `json:"sentiment"`
}

// SentimentData contains source-specific sentiment data
type SentimentData struct {
	SourceName       string  `json:"source_name"`
	SentimentScore   float64 `json:"sentiment_score"`
	PositiveMentions int64   `json:"positive_mentions"`
	NegativeMentions int64   `json:"negative_mentions"`
	NeutralMentions  int64   `json:"neutral_mentions"`
	TotalMentions    int64   `json:"total_mentions"`
}

// BrandMentions contains brand mention analysis
type BrandMentions struct {
	TotalMentions    int64           `json:"total_mentions"`
	PositiveMentions int64           `json:"positive_mentions"`
	NegativeMentions int64           `json:"negative_mentions"`
	NeutralMentions  int64           `json:"neutral_mentions"`
	MentionTrend     string          `json:"mention_trend"`
	TopSources       []MentionSource `json:"top_sources"`
	RecentMentions   []BrandMention  `json:"recent_mentions"`
	MentionVelocity  float64         `json:"mention_velocity"`
	BrandScore       float64         `json:"brand_score"`
}

// MentionSource contains source-specific mention data
type MentionSource struct {
	SourceName     string  `json:"source_name"`
	MentionCount   int64   `json:"mention_count"`
	SentimentScore float64 `json:"sentiment_score"`
	SourceType     string  `json:"source_type"`
}

// BrandMention contains individual brand mention data
type BrandMention struct {
	Source         string    `json:"source"`
	URL            string    `json:"url"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	Sentiment      string    `json:"sentiment"`
	SentimentScore float64   `json:"sentiment_score"`
	PublishedDate  time.Time `json:"published_date"`
	Author         string    `json:"author"`
}

// NewReputationAnalyzer creates a new reputation analyzer
func NewReputationAnalyzer(config *RiskAssessmentConfig, logger *zap.Logger) *ReputationAnalyzer {
	return &ReputationAnalyzer{
		config: config,
		logger: logger,
	}
}

// AnalyzeReputation performs comprehensive reputation analysis
func (ra *ReputationAnalyzer) AnalyzeReputation(ctx context.Context, businessName string, websiteURL string) (*ReputationAnalysisResult, error) {
	ra.logger.Info("Starting reputation analysis",
		zap.String("business", businessName),
		zap.String("website", websiteURL))

	result := &ReputationAnalysisResult{
		BusinessName: businessName,
		LastUpdated:  time.Now(),
	}

	// Analyze social media presence if enabled
	if ra.config.SocialMediaAnalysisEnabled {
		socialMedia, err := ra.analyzeSocialMediaPresence(ctx, businessName, websiteURL)
		if err != nil {
			ra.logger.Warn("Social media analysis failed",
				zap.String("business", businessName),
				zap.Error(err))
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Category:    "reputation",
				Factor:      "social_media_analysis",
				Description: fmt.Sprintf("Social media analysis failed: %v", err),
				Severity:    RiskLevelMedium,
				Score:       0.5,
				Evidence:    err.Error(),
				Impact:      "Cannot verify social media presence",
			})
		} else {
			result.SocialMediaPresence = socialMedia
		}
	}

	// Analyze reviews and ratings if enabled
	if ra.config.ReviewAnalysisEnabled {
		reviewAnalysis, err := ra.analyzeReviewsAndRatings(ctx, businessName, websiteURL)
		if err != nil {
			ra.logger.Warn("Review analysis failed",
				zap.String("business", businessName),
				zap.Error(err))
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Category:    "reputation",
				Factor:      "review_analysis",
				Description: fmt.Sprintf("Review analysis failed: %v", err),
				Severity:    RiskLevelMedium,
				Score:       0.5,
				Evidence:    err.Error(),
				Impact:      "Cannot verify review ratings",
			})
		} else {
			result.ReviewAnalysis = reviewAnalysis
		}
	}

	// Analyze sentiment if enabled
	if ra.config.SentimentAnalysisEnabled {
		sentimentAnalysis, err := ra.analyzeSentiment(ctx, businessName, websiteURL)
		if err != nil {
			ra.logger.Warn("Sentiment analysis failed",
				zap.String("business", businessName),
				zap.Error(err))
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Category:    "reputation",
				Factor:      "sentiment_analysis",
				Description: fmt.Sprintf("Sentiment analysis failed: %v", err),
				Severity:    RiskLevelMedium,
				Score:       0.5,
				Evidence:    err.Error(),
				Impact:      "Cannot verify sentiment analysis",
			})
		} else {
			result.SentimentAnalysis = sentimentAnalysis
		}
	}

	// Analyze brand mentions
	brandMentions, err := ra.analyzeBrandMentions(ctx, businessName, websiteURL)
	if err != nil {
		ra.logger.Warn("Brand mention analysis failed",
			zap.String("business", businessName),
			zap.Error(err))
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Category:    "reputation",
			Factor:      "brand_mention_analysis",
			Description: fmt.Sprintf("Brand mention analysis failed: %v", err),
			Severity:    RiskLevelMedium,
			Score:       0.5,
			Evidence:    err.Error(),
			Impact:      "Cannot verify brand mentions",
		})
	} else {
		result.BrandMentions = brandMentions
	}

	// Calculate overall score
	result.OverallScore = ra.calculateOverallScore(result)

	// Generate recommendations
	result.Recommendations = ra.generateRecommendations(result)

	ra.logger.Info("Reputation analysis completed",
		zap.String("business", businessName),
		zap.Float64("score", result.OverallScore))

	return result, nil
}

// analyzeSocialMediaPresence analyzes social media presence across platforms
func (ra *ReputationAnalyzer) analyzeSocialMediaPresence(ctx context.Context, businessName string, websiteURL string) (*SocialMediaPresence, error) {
	ra.logger.Debug("Analyzing social media presence",
		zap.String("business", businessName))

	// In a real implementation, this would query social media APIs
	// For now, we'll simulate social media data
	platforms := make(map[string]*SocialMediaPlatform)

	// Simulate LinkedIn data
	linkedin := &SocialMediaPlatform{
		PlatformName:   "LinkedIn",
		ProfileURL:     fmt.Sprintf("https://linkedin.com/company/%s", ra.sanitizeBusinessName(businessName)),
		Followers:      1250,
		Following:      150,
		Posts:          45,
		EngagementRate: 0.045,
		LastPostDate:   time.Now().AddDate(0, 0, -3),
		IsVerified:     true,
		IsActive:       true,
	}
	linkedin.ActivityScore = ra.calculateActivityScore(linkedin)
	linkedin.PresenceScore = ra.calculatePresenceScore(linkedin)
	platforms["linkedin"] = linkedin

	// Simulate Twitter/X data
	twitter := &SocialMediaPlatform{
		PlatformName:   "Twitter",
		ProfileURL:     fmt.Sprintf("https://twitter.com/%s", ra.sanitizeBusinessName(businessName)),
		Followers:      890,
		Following:      200,
		Posts:          120,
		EngagementRate: 0.032,
		LastPostDate:   time.Now().AddDate(0, 0, -1),
		IsVerified:     false,
		IsActive:       true,
	}
	twitter.ActivityScore = ra.calculateActivityScore(twitter)
	twitter.PresenceScore = ra.calculatePresenceScore(twitter)
	platforms["twitter"] = twitter

	// Simulate Facebook data
	facebook := &SocialMediaPlatform{
		PlatformName:   "Facebook",
		ProfileURL:     fmt.Sprintf("https://facebook.com/%s", ra.sanitizeBusinessName(businessName)),
		Followers:      2100,
		Following:      50,
		Posts:          78,
		EngagementRate: 0.038,
		LastPostDate:   time.Now().AddDate(0, 0, -5),
		IsVerified:     true,
		IsActive:       true,
	}
	facebook.ActivityScore = ra.calculateActivityScore(facebook)
	facebook.PresenceScore = ra.calculatePresenceScore(facebook)
	platforms["facebook"] = facebook

	// Calculate aggregate metrics
	totalFollowers := int64(0)
	totalPosts := int64(0)
	totalEngagement := float64(0)
	platformCount := 0
	var lastActivityDate time.Time

	for _, platform := range platforms {
		totalFollowers += platform.Followers
		totalPosts += platform.Posts
		totalEngagement += platform.EngagementRate
		platformCount++

		if platform.LastPostDate.After(lastActivityDate) {
			lastActivityDate = platform.LastPostDate
		}
	}

	engagementRate := totalEngagement / float64(platformCount)
	activityScore := ra.calculateOverallActivityScore(platforms)
	presenceScore := ra.calculateOverallPresenceScore(platforms)
	isActive := time.Since(lastActivityDate) < 30*24*time.Hour // Active if posted in last 30 days

	return &SocialMediaPresence{
		Platforms:        platforms,
		TotalFollowers:   totalFollowers,
		TotalPosts:       totalPosts,
		EngagementRate:   engagementRate,
		ActivityScore:    activityScore,
		PresenceScore:    presenceScore,
		LastActivityDate: lastActivityDate,
		IsActive:         isActive,
	}, nil
}

// analyzeReviewsAndRatings analyzes reviews and ratings across platforms
func (ra *ReputationAnalyzer) analyzeReviewsAndRatings(ctx context.Context, businessName string, websiteURL string) (*ReviewAnalysis, error) {
	ra.logger.Debug("Analyzing reviews and ratings",
		zap.String("business", businessName))

	// In a real implementation, this would query review platform APIs
	// For now, we'll simulate review data
	platforms := make(map[string]*ReviewPlatform)

	// Simulate Google Reviews data
	googleReviews := &ReviewPlatform{
		PlatformName:    "Google Reviews",
		Rating:          4.2,
		TotalReviews:    156,
		PositiveReviews: 120,
		NegativeReviews: 15,
		NeutralReviews:  21,
		LastReviewDate:  time.Now().AddDate(0, 0, -2),
		ReviewVelocity:  2.5, // reviews per day
	}
	platforms["google"] = googleReviews

	// Simulate Yelp data
	yelpReviews := &ReviewPlatform{
		PlatformName:    "Yelp",
		Rating:          4.0,
		TotalReviews:    89,
		PositiveReviews: 65,
		NegativeReviews: 12,
		NeutralReviews:  12,
		LastReviewDate:  time.Now().AddDate(0, 0, -5),
		ReviewVelocity:  1.2, // reviews per day
	}
	platforms["yelp"] = yelpReviews

	// Simulate Trustpilot data
	trustpilotReviews := &ReviewPlatform{
		PlatformName:    "Trustpilot",
		Rating:          4.5,
		TotalReviews:    67,
		PositiveReviews: 55,
		NegativeReviews: 8,
		NeutralReviews:  4,
		LastReviewDate:  time.Now().AddDate(0, 0, -1),
		ReviewVelocity:  1.8, // reviews per day
	}
	platforms["trustpilot"] = trustpilotReviews

	// Calculate aggregate metrics
	totalReviews := int64(0)
	totalPositive := int64(0)
	totalNegative := int64(0)
	totalNeutral := int64(0)
	totalRating := float64(0)
	platformCount := 0
	var lastReviewDate time.Time
	totalVelocity := float64(0)

	for _, platform := range platforms {
		totalReviews += platform.TotalReviews
		totalPositive += platform.PositiveReviews
		totalNegative += platform.NegativeReviews
		totalNeutral += platform.NeutralReviews
		totalRating += platform.Rating
		totalVelocity += platform.ReviewVelocity
		platformCount++

		if platform.LastReviewDate.After(lastReviewDate) {
			lastReviewDate = platform.LastReviewDate
		}
	}

	overallRating := totalRating / float64(platformCount)
	reviewVelocity := totalVelocity
	ratingTrend := ra.determineRatingTrend(platforms)
	reviewScore := ra.calculateReviewScore(overallRating, totalReviews, totalPositive, totalNegative)

	return &ReviewAnalysis{
		Platforms:       platforms,
		OverallRating:   overallRating,
		TotalReviews:    totalReviews,
		PositiveReviews: totalPositive,
		NegativeReviews: totalNegative,
		NeutralReviews:  totalNeutral,
		RatingTrend:     ratingTrend,
		ReviewVelocity:  reviewVelocity,
		ReviewScore:     reviewScore,
		LastReviewDate:  lastReviewDate,
	}, nil
}

// analyzeSentiment analyzes sentiment across various sources
func (ra *ReputationAnalyzer) analyzeSentiment(ctx context.Context, businessName string, websiteURL string) (*SentimentAnalysis, error) {
	ra.logger.Debug("Analyzing sentiment",
		zap.String("business", businessName))

	// In a real implementation, this would use NLP services
	// For now, we'll simulate sentiment data
	topKeywords := []SentimentKeyword{
		{Keyword: "professional", Frequency: 45, SentimentScore: 0.8, Sentiment: "positive"},
		{Keyword: "reliable", Frequency: 32, SentimentScore: 0.7, Sentiment: "positive"},
		{Keyword: "responsive", Frequency: 28, SentimentScore: 0.6, Sentiment: "positive"},
		{Keyword: "expensive", Frequency: 15, SentimentScore: -0.3, Sentiment: "negative"},
		{Keyword: "slow", Frequency: 12, SentimentScore: -0.5, Sentiment: "negative"},
	}

	sourceBreakdown := map[string]*SentimentData{
		"social_media": {
			SourceName:       "Social Media",
			SentimentScore:   0.65,
			PositiveMentions: 120,
			NegativeMentions: 25,
			NeutralMentions:  35,
			TotalMentions:    180,
		},
		"reviews": {
			SourceName:       "Reviews",
			SentimentScore:   0.72,
			PositiveMentions: 240,
			NegativeMentions: 35,
			NeutralMentions:  37,
			TotalMentions:    312,
		},
		"news": {
			SourceName:       "News",
			SentimentScore:   0.58,
			PositiveMentions: 45,
			NegativeMentions: 12,
			NeutralMentions:  23,
			TotalMentions:    80,
		},
	}

	// Calculate aggregate metrics
	totalPositive := int64(0)
	totalNegative := int64(0)
	totalNeutral := int64(0)
	totalSentiment := float64(0)
	sourceCount := 0

	for _, source := range sourceBreakdown {
		totalPositive += source.PositiveMentions
		totalNegative += source.NegativeMentions
		totalNeutral += source.NeutralMentions
		totalSentiment += source.SentimentScore
		sourceCount++
	}

	totalMentions := totalPositive + totalNegative + totalNeutral
	overallSentimentScore := totalSentiment / float64(sourceCount)
	overallSentiment := ra.determineSentiment(overallSentimentScore)
	sentimentTrend := ra.determineSentimentTrend(sourceBreakdown)

	return &SentimentAnalysis{
		OverallSentiment: overallSentiment,
		SentimentScore:   overallSentimentScore,
		PositiveMentions: totalPositive,
		NegativeMentions: totalNegative,
		NeutralMentions:  totalNeutral,
		TotalMentions:    totalMentions,
		SentimentTrend:   sentimentTrend,
		TopKeywords:      topKeywords,
		SourceBreakdown:  sourceBreakdown,
	}, nil
}

// analyzeBrandMentions analyzes brand mentions across the web
func (ra *ReputationAnalyzer) analyzeBrandMentions(ctx context.Context, businessName string, websiteURL string) (*BrandMentions, error) {
	ra.logger.Debug("Analyzing brand mentions",
		zap.String("business", businessName))

	// In a real implementation, this would query web search APIs
	// For now, we'll simulate brand mention data
	topSources := []MentionSource{
		{SourceName: "LinkedIn", MentionCount: 45, SentimentScore: 0.7, SourceType: "social"},
		{SourceName: "Twitter", MentionCount: 38, SentimentScore: 0.6, SourceType: "social"},
		{SourceName: "Industry Blog", MentionCount: 12, SentimentScore: 0.8, SourceType: "news"},
		{SourceName: "Local News", MentionCount: 8, SentimentScore: 0.5, SourceType: "news"},
		{SourceName: "Reddit", MentionCount: 15, SentimentScore: 0.4, SourceType: "forum"},
	}

	recentMentions := []BrandMention{
		{
			Source:         "LinkedIn",
			URL:            "https://linkedin.com/posts/example",
			Title:          "Great experience working with " + businessName,
			Content:        "Highly recommend their services...",
			Sentiment:      "positive",
			SentimentScore: 0.8,
			PublishedDate:  time.Now().AddDate(0, 0, -1),
			Author:         "John Smith",
		},
		{
			Source:         "Twitter",
			URL:            "https://twitter.com/user/status/123",
			Title:          "Just completed project with " + businessName,
			Content:        "Very professional team...",
			Sentiment:      "positive",
			SentimentScore: 0.7,
			PublishedDate:  time.Now().AddDate(0, 0, -2),
			Author:         "@techuser",
		},
	}

	// Calculate aggregate metrics
	totalMentions := int64(0)
	totalPositive := int64(0)
	totalNegative := int64(0)
	totalNeutral := int64(0)
	totalSentiment := float64(0)

	for _, source := range topSources {
		totalMentions += source.MentionCount
		totalSentiment += source.SentimentScore * float64(source.MentionCount)
	}

	// Estimate sentiment breakdown based on sentiment scores
	totalSentimentScore := totalSentiment / float64(totalMentions)
	totalPositive = int64(float64(totalMentions) * 0.6)  // Assume 60% positive
	totalNegative = int64(float64(totalMentions) * 0.15) // Assume 15% negative
	totalNeutral = totalMentions - totalPositive - totalNegative

	mentionTrend := ra.determineMentionTrend(topSources)
	mentionVelocity := 3.2 // mentions per day
	brandScore := ra.calculateBrandScore(totalMentions, totalSentimentScore, mentionVelocity)

	return &BrandMentions{
		TotalMentions:    totalMentions,
		PositiveMentions: totalPositive,
		NegativeMentions: totalNegative,
		NeutralMentions:  totalNeutral,
		MentionTrend:     mentionTrend,
		TopSources:       topSources,
		RecentMentions:   recentMentions,
		MentionVelocity:  mentionVelocity,
		BrandScore:       brandScore,
	}, nil
}

// Helper methods for calculations and scoring

func (ra *ReputationAnalyzer) calculateActivityScore(platform *SocialMediaPlatform) float64 {
	score := 0.5 // Base score

	// Activity based on recent posts
	daysSinceLastPost := time.Since(platform.LastPostDate).Hours() / 24
	if daysSinceLastPost < 7 {
		score += 0.3
	} else if daysSinceLastPost < 30 {
		score += 0.2
	} else if daysSinceLastPost < 90 {
		score += 0.1
	}

	// Activity based on engagement rate
	if platform.EngagementRate > 0.05 {
		score += 0.2
	} else if platform.EngagementRate > 0.02 {
		score += 0.1
	}

	return ra.max(0.0, ra.min(1.0, score))
}

func (ra *ReputationAnalyzer) calculatePresenceScore(platform *SocialMediaPlatform) float64 {
	score := 0.5 // Base score

	// Presence based on followers
	if platform.Followers > 10000 {
		score += 0.3
	} else if platform.Followers > 1000 {
		score += 0.2
	} else if platform.Followers > 100 {
		score += 0.1
	}

	// Bonus for verification
	if platform.IsVerified {
		score += 0.1
	}

	// Bonus for active status
	if platform.IsActive {
		score += 0.1
	}

	return ra.max(0.0, ra.min(1.0, score))
}

func (ra *ReputationAnalyzer) calculateOverallActivityScore(platforms map[string]*SocialMediaPlatform) float64 {
	if len(platforms) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, platform := range platforms {
		totalScore += platform.ActivityScore
	}

	return totalScore / float64(len(platforms))
}

func (ra *ReputationAnalyzer) calculateOverallPresenceScore(platforms map[string]*SocialMediaPlatform) float64 {
	if len(platforms) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, platform := range platforms {
		totalScore += platform.PresenceScore
	}

	return totalScore / float64(len(platforms))
}

func (ra *ReputationAnalyzer) calculateReviewScore(rating float64, totalReviews int64, positive int64, negative int64) float64 {
	score := 0.5 // Base score

	// Rating contribution (40% weight)
	ratingScore := rating / 5.0
	score += ratingScore * 0.4

	// Volume contribution (30% weight)
	volumeScore := ra.min(1.0, float64(totalReviews)/100.0) // Normalize to 100 reviews
	score += volumeScore * 0.3

	// Sentiment contribution (30% weight)
	if totalReviews > 0 {
		sentimentRatio := float64(positive) / float64(totalReviews)
		score += sentimentRatio * 0.3
	}

	return ra.max(0.0, ra.min(1.0, score))
}

func (ra *ReputationAnalyzer) calculateBrandScore(mentions int64, sentimentScore float64, velocity float64) float64 {
	score := 0.5 // Base score

	// Mention volume contribution (40% weight)
	mentionScore := ra.min(1.0, float64(mentions)/50.0) // Normalize to 50 mentions
	score += mentionScore * 0.4

	// Sentiment contribution (40% weight)
	score += sentimentScore * 0.4

	// Velocity contribution (20% weight)
	velocityScore := ra.min(1.0, velocity/5.0) // Normalize to 5 mentions per day
	score += velocityScore * 0.2

	return ra.max(0.0, ra.min(1.0, score))
}

func (ra *ReputationAnalyzer) calculateOverallScore(result *ReputationAnalysisResult) float64 {
	score := 0.5 // Base score

	// Social media score (25% weight)
	if result.SocialMediaPresence != nil {
		socialScore := (result.SocialMediaPresence.ActivityScore + result.SocialMediaPresence.PresenceScore) / 2.0
		score += socialScore * 0.25
	}

	// Review score (30% weight)
	if result.ReviewAnalysis != nil {
		score += result.ReviewAnalysis.ReviewScore * 0.30
	}

	// Sentiment score (25% weight)
	if result.SentimentAnalysis != nil {
		score += result.SentimentAnalysis.SentimentScore * 0.25
	}

	// Brand mention score (20% weight)
	if result.BrandMentions != nil {
		score += result.BrandMentions.BrandScore * 0.20
	}

	return ra.max(0.0, ra.min(1.0, score))
}

// Helper methods for trend determination

func (ra *ReputationAnalyzer) determineRatingTrend(platforms map[string]*ReviewPlatform) string {
	// Simplified trend determination
	// In a real implementation, this would analyze historical data
	return "stable"
}

func (ra *ReputationAnalyzer) determineSentiment(sentimentScore float64) string {
	if sentimentScore >= 0.6 {
		return "positive"
	} else if sentimentScore <= 0.4 {
		return "negative"
	}
	return "neutral"
}

func (ra *ReputationAnalyzer) determineSentimentTrend(sources map[string]*SentimentData) string {
	// Simplified trend determination
	// In a real implementation, this would analyze historical data
	return "improving"
}

func (ra *ReputationAnalyzer) determineMentionTrend(sources []MentionSource) string {
	// Simplified trend determination
	// In a real implementation, this would analyze historical data
	return "increasing"
}

func (ra *ReputationAnalyzer) generateRecommendations(result *ReputationAnalysisResult) []string {
	var recommendations []string

	// Social media recommendations
	if result.SocialMediaPresence != nil {
		if !result.SocialMediaPresence.IsActive {
			recommendations = append(recommendations, "Increase social media activity to improve online presence.")
		}
		if result.SocialMediaPresence.EngagementRate < 0.03 {
			recommendations = append(recommendations, "Improve social media engagement through better content and interaction.")
		}
	}

	// Review recommendations
	if result.ReviewAnalysis != nil {
		if result.ReviewAnalysis.OverallRating < 4.0 {
			recommendations = append(recommendations, "Focus on improving customer satisfaction to increase ratings.")
		}
		if result.ReviewAnalysis.TotalReviews < 50 {
			recommendations = append(recommendations, "Encourage more customer reviews to build credibility.")
		}
	}

	// Sentiment recommendations
	if result.SentimentAnalysis != nil {
		if result.SentimentAnalysis.SentimentScore < 0.6 {
			recommendations = append(recommendations, "Address negative sentiment through improved customer service and communication.")
		}
	}

	// Brand mention recommendations
	if result.BrandMentions != nil {
		if result.BrandMentions.TotalMentions < 20 {
			recommendations = append(recommendations, "Increase brand visibility through content marketing and PR efforts.")
		}
	}

	return recommendations
}

// Utility methods

func (ra *ReputationAnalyzer) sanitizeBusinessName(businessName string) string {
	// Remove special characters and convert to lowercase
	reg := regexp.MustCompile("[^a-zA-Z0-9]")
	sanitized := reg.ReplaceAllString(businessName, "")
	return strings.ToLower(sanitized)
}

func (ra *ReputationAnalyzer) max(a, b float64) float64 {
	return math.Max(a, b)
}

func (ra *ReputationAnalyzer) min(a, b float64) float64 {
	return math.Min(a, b)
}
