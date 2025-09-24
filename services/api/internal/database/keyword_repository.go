package database

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Industry represents an industry in the classification system
type Industry struct {
	ID                  int       `json:"id"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	Category            string    `json:"category"`
	ParentIndustryID    *int      `json:"parent_industry_id"`
	ConfidenceThreshold float64   `json:"confidence_threshold"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// IndustryKeyword represents a keyword associated with an industry
type IndustryKeyword struct {
	ID         int       `json:"id"`
	IndustryID int       `json:"industry_id"`
	Keyword    string    `json:"keyword"`
	Weight     float64   `json:"weight"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ClassificationCode represents industry classification codes
type ClassificationCode struct {
	ID          int       `json:"id"`
	IndustryID  int       `json:"industry_id"`
	CodeType    string    `json:"code_type"` // 'naics', 'mcc', 'sic'
	Code        string    `json:"code"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// KeywordRepository defines the interface for keyword classification operations
type KeywordRepository interface {
	// Industry operations
	GetIndustryByID(ctx context.Context, id int) (*Industry, error)
	GetIndustryByName(ctx context.Context, name string) (*Industry, error)
	ListIndustries(ctx context.Context, category string) ([]*Industry, error)
	CreateIndustry(ctx context.Context, industry *Industry) error
	UpdateIndustry(ctx context.Context, industry *Industry) error

	// Keyword operations
	GetKeywordsByIndustry(ctx context.Context, industryID int) ([]*IndustryKeyword, error)
	SearchKeywords(ctx context.Context, query string, limit int) ([]*IndustryKeyword, error)
	AddKeywordToIndustry(ctx context.Context, industryID int, keyword string, weight float64) error
	UpdateKeywordWeight(ctx context.Context, keywordID int, weight float64) error

	// Classification code operations
	GetClassificationCodesByIndustry(ctx context.Context, industryID int) ([]*ClassificationCode, error)
	GetClassificationCodesByType(ctx context.Context, codeType string) ([]*ClassificationCode, error)
	AddClassificationCode(ctx context.Context, code *ClassificationCode) error

	// Search and classification operations
	ClassifyBusiness(ctx context.Context, businessName, description, websiteURL string) (*Industry, float64, error)
	GetTopIndustriesByKeywords(ctx context.Context, keywords []string, limit int) ([]*Industry, error)
}

// SupabaseKeywordRepository implements KeywordRepository using Supabase
type SupabaseKeywordRepository struct {
	client *SupabaseClient
	logger *log.Logger
}

// NewSupabaseKeywordRepository creates a new Supabase keyword repository
func NewSupabaseKeywordRepository(client *SupabaseClient, logger *log.Logger) *SupabaseKeywordRepository {
	if logger == nil {
		logger = log.Default()
	}

	return &SupabaseKeywordRepository{
		client: client,
		logger: logger,
	}
}

// GetIndustryByID retrieves an industry by its ID
func (r *SupabaseKeywordRepository) GetIndustryByID(ctx context.Context, id int) (*Industry, error) {
	r.logger.Printf("🔍 Getting industry by ID: %d", id)

	var industry Industry
	_, _, err := r.client.GetPostgrestClient().
		From("industries").
		Select("*", "", false).
		Eq("id", fmt.Sprintf("%d", id)).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to get industry by ID %d: %w", id, err)
	}

	return &industry, nil
}

// GetIndustryByName retrieves an industry by its name
func (r *SupabaseKeywordRepository) GetIndustryByName(ctx context.Context, name string) (*Industry, error) {
	r.logger.Printf("🔍 Getting industry by name: %s", name)

	var industry Industry
	_, _, err := r.client.GetPostgrestClient().
		From("industries").
		Select("*", "", false).
		Eq("name", name).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to get industry by name %s: %w", name, err)
	}

	return &industry, nil
}

// ListIndustries retrieves industries, optionally filtered by category
func (r *SupabaseKeywordRepository) ListIndustries(ctx context.Context, category string) ([]*Industry, error) {
	r.logger.Printf("🔍 Listing industries, category: %s", category)

	query := r.client.GetPostgrestClient().
		From("industries").
		Select("*", "", false).
		Eq("is_active", "true")

	if category != "" {
		query = query.Eq("category", category)
	}

	var industries []*Industry
	_, _, err := query.Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list industries: %w", err)
	}

	return industries, nil
}

// GetKeywordsByIndustry retrieves keywords for a specific industry
func (r *SupabaseKeywordRepository) GetKeywordsByIndustry(ctx context.Context, industryID int) ([]*IndustryKeyword, error) {
	r.logger.Printf("🔍 Getting keywords for industry ID: %d", industryID)

	var keywords []*IndustryKeyword
	_, _, err := r.client.GetPostgrestClient().
		From("industry_keywords").
		Select("*", "", false).
		Eq("industry_id", fmt.Sprintf("%d", industryID)).
		Eq("is_active", "true").
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to get keywords for industry %d: %w", industryID, err)
	}

	return keywords, nil
}

// SearchKeywords searches for keywords containing the query string
func (r *SupabaseKeywordRepository) SearchKeywords(ctx context.Context, query string, limit int) ([]*IndustryKeyword, error) {
	r.logger.Printf("🔍 Searching keywords for query: %s", query)

	var keywords []*IndustryKeyword
	_, _, err := r.client.GetPostgrestClient().
		From("industry_keywords").
		Select("*", "", false).
		Ilike("keyword", "%"+query+"%").
		Eq("is_active", "true").
		Limit(limit, "").
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to search keywords for query %s: %w", query, err)
	}

	return keywords, nil
}

// ClassifyBusiness performs business classification based on input data
func (r *SupabaseKeywordRepository) ClassifyBusiness(ctx context.Context, businessName, description, websiteURL string) (*Industry, float64, error) {
	r.logger.Printf("🔍 Classifying business: %s", businessName)

	// This is a simplified implementation
	// In a real system, you would:
	// 1. Extract keywords from business name, description, and website
	// 2. Search for matching keywords in the database
	// 3. Calculate confidence scores based on keyword matches
	// 4. Return the best matching industry with confidence score

	// For now, return a default industry
	// TODO: Implement full keyword-based classification logic
	defaultIndustry := &Industry{
		ID:                  1,
		Name:                "General Business",
		Description:         "Default industry for unclassified businesses",
		Category:            "traditional",
		ConfidenceThreshold: 0.5,
		IsActive:            true,
	}

	return defaultIndustry, 0.5, nil
}

// GetTopIndustriesByKeywords finds industries that best match the given keywords
func (r *SupabaseKeywordRepository) GetTopIndustriesByKeywords(ctx context.Context, keywords []string, limit int) ([]*Industry, error) {
	r.logger.Printf("🔍 Finding top industries for keywords: %v", keywords)

	var industries []*Industry
	_, _, err := r.client.GetPostgrestClient().
		From("industries").
		Select("*", "", false).
		Eq("is_active", "true").
		Limit(limit, "").
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to get top industries for keywords: %w", err)
	}

	return industries, nil
}

// Placeholder implementations for remaining interface methods
func (r *SupabaseKeywordRepository) CreateIndustry(ctx context.Context, industry *Industry) error {
	return fmt.Errorf("CreateIndustry not implemented yet")
}

func (r *SupabaseKeywordRepository) UpdateIndustry(ctx context.Context, industry *Industry) error {
	return fmt.Errorf("UpdateIndustry not implemented yet")
}

func (r *SupabaseKeywordRepository) AddKeywordToIndustry(ctx context.Context, industryID int, keyword string, weight float64) error {
	return fmt.Errorf("AddKeywordToIndustry not implemented yet")
}

func (r *SupabaseKeywordRepository) UpdateKeywordWeight(ctx context.Context, keywordID int, weight float64) error {
	return fmt.Errorf("UpdateKeywordWeight not implemented yet")
}

func (r *SupabaseKeywordRepository) GetClassificationCodesByIndustry(ctx context.Context, industryID int) ([]*ClassificationCode, error) {
	return nil, fmt.Errorf("GetClassificationCodesByIndustry not implemented yet")
}

func (r *SupabaseKeywordRepository) GetClassificationCodesByType(ctx context.Context, codeType string) ([]*ClassificationCode, error) {
	return nil, fmt.Errorf("GetClassificationCodesByType not implemented yet")
}

func (r *SupabaseKeywordRepository) AddClassificationCode(ctx context.Context, code *ClassificationCode) error {
	return fmt.Errorf("AddClassificationCode not implemented yet")
}
