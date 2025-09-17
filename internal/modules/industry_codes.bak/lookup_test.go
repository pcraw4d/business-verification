package industry_codes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestIndustryCodeLookup_Lookup(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)
	icl := NewIndustryCodeLookup(icdb, logger)

	ctx := context.Background()
	err := icdb.Initialize(ctx)
	require.NoError(t, err)

	// Insert test codes
	testCodes := []*IndustryCode{
		{
			ID:          "test-1",
			Code:        "5411",
			Type:        CodeTypeSIC,
			Description: "Legal Services",
			Category:    "Professional Services",
			Keywords:    []string{"legal", "law"},
			Confidence:  0.95,
		},
		{
			ID:          "test-2",
			Code:        "5412",
			Type:        CodeTypeSIC,
			Description: "Accounting Services",
			Category:    "Professional Services",
			Keywords:    []string{"accounting", "tax"},
			Confidence:  0.90,
		},
		{
			ID:          "test-3",
			Code:        "5413",
			Type:        CodeTypeNAICS,
			Description: "Architectural Services",
			Category:    "Professional Services",
			Keywords:    []string{"architecture", "design"},
			Confidence:  0.85,
		},
	}

	for _, code := range testCodes {
		err = icdb.InsertCode(ctx, code)
		require.NoError(t, err)
	}

	// Test basic lookup
	req := &LookupRequest{
		Query:      "Legal",
		MaxResults: 10,
	}

	response, err := icl.Lookup(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "Legal", response.Query)
	assert.Len(t, response.Results, 1)
	assert.Equal(t, "5411", response.Results[0].Code.Code)
	assert.Greater(t, response.Results[0].Relevance, 0.0)

	// Test lookup with code type filter
	req = &LookupRequest{
		Query:      "Services",
		CodeTypes:  []CodeType{CodeTypeSIC},
		MaxResults: 10,
	}

	response, err = icl.Lookup(ctx, req)
	require.NoError(t, err)
	assert.Len(t, response.Results, 2)
	assert.Equal(t, 2, response.CodeTypeStats["sic"])

	// Test lookup with confidence filter
	req = &LookupRequest{
		Query:         "Services",
		MinConfidence: 0.92,
		MaxResults:    10,
	}

	response, err = icl.Lookup(ctx, req)
	require.NoError(t, err)
	assert.Len(t, response.Results, 1)
	assert.Equal(t, "5411", response.Results[0].Code.Code)

	// Test lookup with category filter
	req = &LookupRequest{
		Query:      "Services",
		Categories: []string{"Professional"},
		MaxResults: 10,
	}

	response, err = icl.Lookup(ctx, req)
	require.NoError(t, err)
	assert.Len(t, response.Results, 3)
}

func TestIndustryCodeLookup_ExactCodeMatch(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)
	icl := NewIndustryCodeLookup(icdb, logger)

	ctx := context.Background()
	err := icdb.Initialize(ctx)
	require.NoError(t, err)

	testCode := &IndustryCode{
		ID:          "test-1",
		Code:        "5411",
		Type:        CodeTypeSIC,
		Description: "Legal Services",
		Category:    "Professional Services",
		Confidence:  0.95,
	}

	err = icdb.InsertCode(ctx, testCode)
	require.NoError(t, err)

	// Test exact code match
	req := &LookupRequest{
		Query:      "5411",
		MaxResults: 10,
	}

	response, err := icl.Lookup(ctx, req)
	require.NoError(t, err)
	assert.Len(t, response.Results, 1)
	assert.Equal(t, "5411", response.Results[0].Code.Code)
	assert.Equal(t, 1.0, response.Results[0].Relevance)
	assert.Equal(t, "exact_code", response.Results[0].MatchType)
}

func TestIndustryCodeLookup_CalculateRelevance(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)
	icl := NewIndustryCodeLookup(icdb, logger)

	ctx := context.Background()
	err := icdb.Initialize(ctx)
	require.NoError(t, err)

	testCode := &IndustryCode{
		ID:          "test-1",
		Code:        "5411",
		Type:        CodeTypeSIC,
		Description: "Legal Services and Consulting",
		Category:    "Professional Legal Services",
		Subcategory: "Legal Consulting",
		Keywords:    []string{"legal", "law", "attorney"},
		Confidence:  0.95,
	}

	err = icdb.InsertCode(ctx, testCode)
	require.NoError(t, err)

	// Test exact code match
	req := &LookupRequest{
		Query:      "5411",
		MaxResults: 10,
	}

	response, err := icl.Lookup(ctx, req)
	require.NoError(t, err)
	assert.Len(t, response.Results, 1)
	assert.Equal(t, 1.0, response.Results[0].Relevance)
	assert.Contains(t, response.Results[0].MatchedOn, "exact_code")

	// Test description match
	req = &LookupRequest{
		Query:      "Legal",
		MaxResults: 10,
	}

	response, err = icl.Lookup(ctx, req)
	require.NoError(t, err)
	assert.Len(t, response.Results, 1)
	assert.Greater(t, response.Results[0].Relevance, 0.0)
	assert.Contains(t, response.Results[0].MatchedOn, "description")

	// Test category match
	req = &LookupRequest{
		Query:      "Professional",
		MaxResults: 10,
	}

	response, err = icl.Lookup(ctx, req)
	require.NoError(t, err)
	assert.Len(t, response.Results, 1)
	assert.Greater(t, response.Results[0].Relevance, 0.0)
	assert.Contains(t, response.Results[0].MatchedOn, "category")

	// Test keyword match
	req = &LookupRequest{
		Query:      "attorney",
		MaxResults: 10,
	}

	response, err = icl.Lookup(ctx, req)
	require.NoError(t, err)
	assert.Len(t, response.Results, 1)
	assert.Greater(t, response.Results[0].Relevance, 0.0)
	assert.Contains(t, response.Results[0].MatchedOn, "keywords")
}

func TestIndustryCodeLookup_IsExactCodeMatch(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)
	icl := NewIndustryCodeLookup(icdb, logger)

	// Test valid code formats
	assert.True(t, icl.isExactCodeMatch("5411"))    // SIC
	assert.True(t, icl.isExactCodeMatch("5411-0"))  // SIC with dash
	assert.True(t, icl.isExactCodeMatch("541100"))  // NAICS
	assert.True(t, icl.isExactCodeMatch("5411 00")) // NAICS with space
	assert.True(t, icl.isExactCodeMatch("5411-00")) // NAICS with dash

	// Test invalid code formats
	assert.False(t, icl.isExactCodeMatch("541"))   // Too short
	assert.False(t, icl.isExactCodeMatch("54111")) // Too long for SIC
	assert.False(t, icl.isExactCodeMatch("5411a")) // Contains letters
	assert.False(t, icl.isExactCodeMatch("legal")) // Not numeric
	assert.False(t, icl.isExactCodeMatch(""))      // Empty
}

func TestIndustryCodeLookup_GetTopCodesByType(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)
	icl := NewIndustryCodeLookup(icdb, logger)

	ctx := context.Background()
	err := icdb.Initialize(ctx)
	require.NoError(t, err)

	// Insert test codes with different confidence levels
	testCodes := []*IndustryCode{
		{
			ID:          "test-1",
			Code:        "5411",
			Type:        CodeTypeSIC,
			Description: "Legal Services",
			Category:    "Professional Services",
			Confidence:  0.95,
		},
		{
			ID:          "test-2",
			Code:        "5412",
			Type:        CodeTypeSIC,
			Description: "Accounting Services",
			Category:    "Professional Services",
			Confidence:  0.90,
		},
		{
			ID:          "test-3",
			Code:        "5413",
			Type:        CodeTypeNAICS,
			Description: "Architectural Services",
			Category:    "Professional Services",
			Confidence:  0.85,
		},
	}

	for _, code := range testCodes {
		err = icdb.InsertCode(ctx, code)
		require.NoError(t, err)
	}

	// Test getting top codes by type
	results, err := icl.GetTopCodesByType(ctx, 5)
	require.NoError(t, err)

	// Verify SIC codes are sorted by confidence
	sicCodes := results[CodeTypeSIC]
	assert.Len(t, sicCodes, 2)
	assert.Equal(t, "5411", sicCodes[0].Code) // Higher confidence first
	assert.Equal(t, "5412", sicCodes[1].Code)

	// Verify NAICS codes
	naicsCodes := results[CodeTypeNAICS]
	assert.Len(t, naicsCodes, 1)
	assert.Equal(t, "5413", naicsCodes[0].Code)
}

func TestIndustryCodeLookup_GetCodesByCategory(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)
	icl := NewIndustryCodeLookup(icdb, logger)

	ctx := context.Background()
	err := icdb.Initialize(ctx)
	require.NoError(t, err)

	// Insert test codes
	testCodes := []*IndustryCode{
		{
			ID:          "test-1",
			Code:        "5411",
			Type:        CodeTypeSIC,
			Description: "Legal Services",
			Category:    "Professional Services",
			Confidence:  0.95,
		},
		{
			ID:          "test-2",
			Code:        "5412",
			Type:        CodeTypeSIC,
			Description: "Accounting Services",
			Category:    "Professional Services",
			Confidence:  0.90,
		},
		{
			ID:          "test-3",
			Code:        "5413",
			Type:        CodeTypeNAICS,
			Description: "Architectural Services",
			Category:    "Professional Services",
			Confidence:  0.85,
		},
		{
			ID:          "test-4",
			Code:        "5414",
			Type:        CodeTypeSIC,
			Description: "Engineering Services",
			Category:    "Engineering Services",
			Confidence:  0.80,
		},
	}

	for _, code := range testCodes {
		err = icdb.InsertCode(ctx, code)
		require.NoError(t, err)
	}

	// Test getting codes by category
	results, err := icl.GetCodesByCategory(ctx, "Professional", []CodeType{}, 10)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	// Test with code type filter
	results, err = icl.GetCodesByCategory(ctx, "Professional", []CodeType{CodeTypeSIC}, 10)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Test with limit
	results, err = icl.GetCodesByCategory(ctx, "Professional", []CodeType{}, 2)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Verify results are sorted by relevance and confidence
	assert.GreaterOrEqual(t, results[0].Relevance, results[1].Relevance)
}

func TestIndustryCodeLookup_GetCodeSuggestions(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)
	icl := NewIndustryCodeLookup(icdb, logger)

	ctx := context.Background()
	err := icdb.Initialize(ctx)
	require.NoError(t, err)

	testCode := &IndustryCode{
		ID:          "test-1",
		Code:        "5411",
		Type:        CodeTypeSIC,
		Description: "Legal Services and Consulting",
		Category:    "Professional Services",
		Keywords:    []string{"legal", "law", "attorney"},
		Confidence:  0.95,
	}

	err = icdb.InsertCode(ctx, testCode)
	require.NoError(t, err)

	// Test getting suggestions
	suggestions, err := icl.GetCodeSuggestions(ctx, "leg", 10)
	require.NoError(t, err)
	assert.Contains(t, suggestions, "5411")
	assert.Contains(t, suggestions, "legal")

	// Test with short query
	suggestions, err = icl.GetCodeSuggestions(ctx, "l", 10)
	require.NoError(t, err)
	assert.Empty(t, suggestions)

	// Test with non-matching query
	suggestions, err = icl.GetCodeSuggestions(ctx, "xyz", 10)
	require.NoError(t, err)
	assert.Empty(t, suggestions)
}

func TestIndustryCodeLookup_GetCodeStats(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)
	icl := NewIndustryCodeLookup(icdb, logger)

	ctx := context.Background()
	err := icdb.Initialize(ctx)
	require.NoError(t, err)

	// Insert test codes
	testCodes := []*IndustryCode{
		{
			ID:          "test-1",
			Code:        "5411",
			Type:        CodeTypeSIC,
			Description: "Legal Services",
			Category:    "Professional Services",
			Confidence:  0.95,
		},
		{
			ID:          "test-2",
			Code:        "5412",
			Type:        CodeTypeSIC,
			Description: "Accounting Services",
			Category:    "Professional Services",
			Confidence:  0.90,
		},
		{
			ID:          "test-3",
			Code:        "5413",
			Type:        CodeTypeNAICS,
			Description: "Architectural Services",
			Category:    "Professional Services",
			Confidence:  0.85,
		},
	}

	for _, code := range testCodes {
		err = icdb.InsertCode(ctx, code)
		require.NoError(t, err)
	}

	// Test getting stats
	stats, err := icl.GetCodeStats(ctx)
	require.NoError(t, err)

	// Verify stats
	assert.Equal(t, 3, stats["total_codes"])

	sicStats := stats["sic"].(map[string]interface{})
	assert.Equal(t, 2, sicStats["count"])
	assert.Equal(t, 0.925, sicStats["avg_confidence"])

	naicsStats := stats["naics"].(map[string]interface{})
	assert.Equal(t, 1, naicsStats["count"])
	assert.Equal(t, 0.85, naicsStats["avg_confidence"])
}

func TestIndustryCodeLookup_EmptyDatabase(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)
	icl := NewIndustryCodeLookup(icdb, logger)

	ctx := context.Background()
	err := icdb.Initialize(ctx)
	require.NoError(t, err)

	// Test lookup with empty database
	req := &LookupRequest{
		Query:      "Legal",
		MaxResults: 10,
	}

	response, err := icl.Lookup(ctx, req)
	require.NoError(t, err)
	assert.Empty(t, response.Results)
	assert.Equal(t, 0, response.TotalFound)

	// Test getting top codes with empty database
	results, err := icl.GetTopCodesByType(ctx, 5)
	require.NoError(t, err)
	assert.Empty(t, results[CodeTypeSIC])
	assert.Empty(t, results[CodeTypeNAICS])

	// Test getting suggestions with empty database
	suggestions, err := icl.GetCodeSuggestions(ctx, "legal", 10)
	require.NoError(t, err)
	assert.Empty(t, suggestions)
}
