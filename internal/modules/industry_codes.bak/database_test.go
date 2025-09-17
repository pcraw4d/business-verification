package industry_codes

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func setupTestDatabase(t *testing.T) (*sql.DB, func()) {
	// Use an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

func TestIndustryCodeDatabase_Initialize(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)

	ctx := context.Background()
	err := icdb.Initialize(ctx)
	require.NoError(t, err)

	// Verify table was created by trying to insert a test record
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
}

func TestIndustryCodeDatabase_InsertCode(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)

	ctx := context.Background()
	err := icdb.Initialize(ctx)
	require.NoError(t, err)

	testCode := &IndustryCode{
		ID:          "test-1",
		Code:        "5411",
		Type:        CodeTypeSIC,
		Description: "Legal Services",
		Category:    "Professional Services",
		Subcategory: "Legal",
		Keywords:    []string{"legal", "law", "attorney"},
		Confidence:  0.95,
	}

	err = icdb.InsertCode(ctx, testCode)
	require.NoError(t, err)

	// Test duplicate insertion (should update)
	testCode.Description = "Updated Legal Services"
	err = icdb.InsertCode(ctx, testCode)
	require.NoError(t, err)

	// Verify the update
	retrieved, err := icdb.GetCodeByID(ctx, "test-1")
	require.NoError(t, err)
	assert.Equal(t, "Updated Legal Services", retrieved.Description)
}

func TestIndustryCodeDatabase_GetCodeByID(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)

	ctx := context.Background()
	err := icdb.Initialize(ctx)
	require.NoError(t, err)

	testCode := &IndustryCode{
		ID:          "test-1",
		Code:        "5411",
		Type:        CodeTypeSIC,
		Description: "Legal Services",
		Category:    "Professional Services",
		Keywords:    []string{"legal", "law"},
		Confidence:  0.95,
	}

	err = icdb.InsertCode(ctx, testCode)
	require.NoError(t, err)

	// Test successful retrieval
	retrieved, err := icdb.GetCodeByID(ctx, "test-1")
	require.NoError(t, err)
	assert.Equal(t, testCode.Code, retrieved.Code)
	assert.Equal(t, testCode.Type, retrieved.Type)
	assert.Equal(t, testCode.Description, retrieved.Description)
	assert.Equal(t, testCode.Keywords, retrieved.Keywords)

	// Test non-existent ID
	_, err = icdb.GetCodeByID(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "industry code not found")
}

func TestIndustryCodeDatabase_GetCodeByCodeAndType(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)

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

	// Test successful retrieval
	retrieved, err := icdb.GetCodeByCodeAndType(ctx, "5411", CodeTypeSIC)
	require.NoError(t, err)
	assert.Equal(t, testCode.ID, retrieved.ID)
	assert.Equal(t, testCode.Description, retrieved.Description)

	// Test non-existent code
	_, err = icdb.GetCodeByCodeAndType(ctx, "9999", CodeTypeSIC)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "industry code not found")
}

func TestIndustryCodeDatabase_SearchCodes(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)

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
			Type:        CodeTypeSIC,
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

	// Test search by description
	results, err := icdb.SearchCodes(ctx, "Legal", nil, 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "5411", results[0].Code)

	// Test search by category
	results, err = icdb.SearchCodes(ctx, "Professional", nil, 10)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	// Test search with code type filter
	sicType := CodeTypeSIC
	results, err = icdb.SearchCodes(ctx, "Services", &sicType, 10)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	// Test search with limit
	results, err = icdb.SearchCodes(ctx, "Services", nil, 2)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestIndustryCodeDatabase_GetCodesByType(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)

	ctx := context.Background()
	err := icdb.Initialize(ctx)
	require.NoError(t, err)

	// Insert test codes of different types
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

	// Test getting SIC codes
	results, err := icdb.GetCodesByType(ctx, CodeTypeSIC, 10, 0)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Test getting NAICS codes
	results, err = icdb.GetCodesByType(ctx, CodeTypeNAICS, 10, 0)
	require.NoError(t, err)
	assert.Len(t, results, 1)

	// Test pagination
	results, err = icdb.GetCodesByType(ctx, CodeTypeSIC, 1, 0)
	require.NoError(t, err)
	assert.Len(t, results, 1)

	results, err = icdb.GetCodesByType(ctx, CodeTypeSIC, 1, 1)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestIndustryCodeDatabase_GetCodesByCategory(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)

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

	// Test search by category
	results, err := icdb.GetCodesByCategory(ctx, "Professional", nil, 10)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	// Test search with code type filter
	sicType := CodeTypeSIC
	results, err = icdb.GetCodesByCategory(ctx, "Professional", &sicType, 10)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Test search with limit
	results, err = icdb.GetCodesByCategory(ctx, "Professional", nil, 2)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestIndustryCodeDatabase_UpdateCodeConfidence(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)

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

	// Test updating confidence
	newConfidence := 0.98
	err = icdb.UpdateCodeConfidence(ctx, "test-1", newConfidence)
	require.NoError(t, err)

	// Verify the update
	retrieved, err := icdb.GetCodeByID(ctx, "test-1")
	require.NoError(t, err)
	assert.Equal(t, newConfidence, retrieved.Confidence)

	// Test updating non-existent code
	err = icdb.UpdateCodeConfidence(ctx, "non-existent", 0.5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "industry code not found")
}

func TestIndustryCodeDatabase_GetCodeStats(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)

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
	stats, err := icdb.GetCodeStats(ctx)
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

func TestIndustryCodeDatabase_KeywordsHandling(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)

	ctx := context.Background()
	err := icdb.Initialize(ctx)
	require.NoError(t, err)

	// Test with keywords
	testCode := &IndustryCode{
		ID:          "test-1",
		Code:        "5411",
		Type:        CodeTypeSIC,
		Description: "Legal Services",
		Category:    "Professional Services",
		Keywords:    []string{"legal", "law", "attorney", "lawyer"},
		Confidence:  0.95,
	}

	err = icdb.InsertCode(ctx, testCode)
	require.NoError(t, err)

	// Test with empty keywords
	testCode2 := &IndustryCode{
		ID:          "test-2",
		Code:        "5412",
		Type:        CodeTypeSIC,
		Description: "Accounting Services",
		Category:    "Professional Services",
		Keywords:    []string{},
		Confidence:  0.90,
	}

	err = icdb.InsertCode(ctx, testCode2)
	require.NoError(t, err)

	// Verify keywords are stored and retrieved correctly
	retrieved, err := icdb.GetCodeByID(ctx, "test-1")
	require.NoError(t, err)
	assert.Equal(t, []string{"legal", "law", "attorney", "lawyer"}, retrieved.Keywords)

	retrieved2, err := icdb.GetCodeByID(ctx, "test-2")
	require.NoError(t, err)
	assert.Equal(t, []string{}, retrieved2.Keywords)
}

func TestIndustryCodeDatabase_ConcurrentAccess(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	icdb := NewIndustryCodeDatabase(db, logger)

	ctx := context.Background()
	err := icdb.Initialize(ctx)
	require.NoError(t, err)

	// Test concurrent insertions with proper synchronization
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			testCode := &IndustryCode{
				ID:          fmt.Sprintf("test-%d", index),
				Code:        fmt.Sprintf("54%d", index),
				Type:        CodeTypeSIC,
				Description: fmt.Sprintf("Service %d", index),
				Category:    "Professional Services",
				Confidence:  0.9,
			}

			// Use mutex to ensure thread-safe database access
			mu.Lock()
			err := icdb.InsertCode(ctx, testCode)
			mu.Unlock()

			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		assert.NoError(t, err)
	}

	// Verify all codes were inserted
	results, err := icdb.GetCodesByType(ctx, CodeTypeSIC, 20, 0)
	require.NoError(t, err)
	assert.Len(t, results, 10)
}
