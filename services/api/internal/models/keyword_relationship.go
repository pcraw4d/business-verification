// Package models provides data structures for the KYB platform
package models

import (
	"time"
)

// KeywordRelationship represents a relationship between keywords (synonyms, abbreviations, etc.)
type KeywordRelationship struct {
	ID               int       `json:"id" db:"id"`
	PrimaryKeyword   string    `json:"primary_keyword" db:"primary_keyword"`
	RelatedKeyword   string    `json:"related_keyword" db:"related_keyword"`
	RelationshipType string    `json:"relationship_type" db:"relationship_type"` // synonym, abbreviation, related, variant
	ConfidenceScore  float64   `json:"confidence_score" db:"confidence_score"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// KeywordContext represents industry-specific context for keywords
type KeywordContext struct {
	ID            int       `json:"id" db:"id"`
	Keyword       string    `json:"keyword" db:"keyword"`
	IndustryID    int       `json:"industry_id" db:"industry_id"`
	ContextType   string    `json:"context_type" db:"context_type"` // primary, secondary, technical, business, general
	ContextWeight float64   `json:"context_weight" db:"context_weight"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// KeywordExpansionResult represents the result of keyword expansion
type KeywordExpansionResult struct {
	OriginalKeyword  string            `json:"original_keyword"`
	ExpandedKeywords []ExpandedKeyword `json:"expanded_keywords"`
	Confidence       float64           `json:"confidence"`
	TotalMatches     int               `json:"total_matches"`
}

// ExpandedKeyword represents an expanded keyword with its relationship
type ExpandedKeyword struct {
	Keyword          string  `json:"keyword"`
	RelationshipType string  `json:"relationship_type"`
	Confidence       float64 `json:"confidence"`
	ContextWeight    float64 `json:"context_weight"`
	IndustryRelevant bool    `json:"industry_relevant"`
}

// RelationshipType constants
const (
	RelationshipTypeSynonym      = "synonym"
	RelationshipTypeAbbreviation = "abbreviation"
	RelationshipTypeRelated      = "related"
	RelationshipTypeVariant      = "variant"
)

// ContextType constants
const (
	ContextTypePrimary   = "primary"
	ContextTypeSecondary = "secondary"
	ContextTypeTechnical = "technical"
	ContextTypeBusiness  = "business"
	ContextTypeGeneral   = "general"
)
