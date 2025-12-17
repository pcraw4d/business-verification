package classification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/external"
)

// EmbeddingClassifier handles embedding-based classification using vector similarity
type EmbeddingClassifier struct {
	embeddingServiceURL string
	supabaseRepo        repository.KeywordRepository
	httpClient          *http.Client
	logger              *log.Logger
}

// NewEmbeddingClassifier creates a new embedding classifier
func NewEmbeddingClassifier(
	embeddingServiceURL string,
	repo repository.KeywordRepository,
	logger *log.Logger,
) *EmbeddingClassifier {
	if logger == nil {
		logger = log.Default()
	}

	// Phase 5: Optimize HTTP client with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        100,              // Maximum idle connections
		MaxIdleConnsPerHost: 10,               // Maximum idle connections per host
		IdleConnTimeout:     90 * time.Second, // Timeout for idle connections
		DisableCompression:  false,            // Enable compression
		DisableKeepAlives:   false,            // Enable keep-alives for connection reuse
		ForceAttemptHTTP2:   true,             // Enable HTTP/2 support
		TLSHandshakeTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &EmbeddingClassifier{
		embeddingServiceURL: embeddingServiceURL,
		supabaseRepo:        repo,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
		logger: logger,
	}
}

// EmbeddingClassificationResult represents the result of embedding-based classification
type EmbeddingClassificationResult struct {
	MCC             []CodeResult
	SIC             []CodeResult
	NAICS           []CodeResult
	Confidence      float64
	Method          string
	TopMatch        string  // Description of top match
	TopSimilarity   float64
	ProcessingTimeMs int64
}

// Note: CodeResult is defined in classifier.go, reusing it here

// ClassifyByEmbedding performs classification using embedding-based vector similarity
func (e *EmbeddingClassifier) ClassifyByEmbedding(
	ctx context.Context,
	content *external.ScrapedContent,
) (*EmbeddingClassificationResult, error) {
	startTime := time.Now()

	e.logger.Printf("🔍 [Layer 2] Starting embedding-based classification for domain: %s", content.Domain)

	// Step 1: Prepare text for embedding
	text := e.prepareTextForEmbedding(content)

	if len(text) < 50 {
		return nil, fmt.Errorf("insufficient text for embedding: %d chars", len(text))
	}

	e.logger.Printf("📝 [Layer 2] Prepared text for embedding (length: %d chars)", len(text))

	// Step 2: Generate embedding
	embedding, err := e.getEmbedding(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	e.logger.Printf("✅ [Layer 2] Generated embedding (dimension: %d)", len(embedding))

	// Step 3: Search for similar codes (each type)
	mccMatches, err := e.searchSimilarCodes(ctx, embedding, "MCC", 0.70, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to search MCC codes: %w", err)
	}

	sicMatches, err := e.searchSimilarCodes(ctx, embedding, "SIC", 0.70, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to search SIC codes: %w", err)
	}

	naicsMatches, err := e.searchSimilarCodes(ctx, embedding, "NAICS", 0.70, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to search NAICS codes: %w", err)
	}

	// Step 4: Select top 3 from each type
	result := &EmbeddingClassificationResult{
		MCC:    e.selectTopCodes(mccMatches, 3),
		SIC:    e.selectTopCodes(sicMatches, 3),
		NAICS:  e.selectTopCodes(naicsMatches, 3),
		Method: "embedding_similarity",
	}

	// Step 5: Calculate overall confidence
	if len(mccMatches) > 0 {
		result.TopMatch = mccMatches[0].Description
		result.TopSimilarity = mccMatches[0].Confidence
		result.Confidence = e.calculateConfidence(mccMatches, sicMatches, naicsMatches)
	} else {
		result.Confidence = 0.0
	}

	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()

	e.logger.Printf("✅ [Layer 2] Embedding classification complete (confidence: %.2f%%, top_match: %s, duration: %dms)",
		result.Confidence*100, result.TopMatch, result.ProcessingTimeMs)

	return result, nil
}

// prepareTextForEmbedding prepares text from scraped content for embedding
func (e *EmbeddingClassifier) prepareTextForEmbedding(content *external.ScrapedContent) string {
	parts := []string{}

	// Priority order: Title > Meta > About > Headings > Navigation

	// Title (highest signal - include 2x)
	if content.Title != "" {
		parts = append(parts, content.Title)
		parts = append(parts, content.Title) // Repeat for emphasis
	}

	// Meta description (high signal)
	if content.MetaDesc != "" {
		parts = append(parts, content.MetaDesc)
	}

	// About section (most contextual info)
	if content.AboutText != "" {
		// Limit to 500 chars to keep embedding focused
		aboutText := content.AboutText
		if len(aboutText) > 500 {
			aboutText = aboutText[:500]
		}
		parts = append(parts, aboutText)
	}

	// Top headings (good signal)
	if len(content.Headings) > 0 {
		// Take first 5 headings
		headingCount := 5
		if len(content.Headings) < headingCount {
			headingCount = len(content.Headings)
		}
		headings := strings.Join(content.Headings[:headingCount], ". ")
		parts = append(parts, headings)
	}

	// Navigation (indicates business areas)
	if len(content.NavMenu) > 0 {
		// Take first 10 nav items
		navCount := 10
		if len(content.NavMenu) < navCount {
			navCount = len(content.NavMenu)
		}
		nav := strings.Join(content.NavMenu[:navCount], ", ")
		parts = append(parts, nav)
	}

	// Combine
	combined := strings.Join(parts, ". ")

	// Truncate to 5000 chars (model limit)
	if len(combined) > 5000 {
		combined = combined[:5000]
	}

	return combined
}

// getEmbedding gets embedding from embedding service
func (e *EmbeddingClassifier) getEmbedding(ctx context.Context, text string) ([]float64, error) {
	reqBody := map[string]interface{}{
		"text":            text,
		"truncate_length": 5000,
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		e.embeddingServiceURL+"/embed",
		bytes.NewReader(reqBodyJSON),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding service returned status %d", resp.StatusCode)
	}

	var result struct {
		Embedding        []float64 `json:"embedding"`
		Dimension        int       `json:"dimension"`
		ProcessingTimeMs int       `json:"processing_time_ms"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	e.logger.Printf("📊 [Layer 2] Embedding service response (dimension: %d, processing_time: %dms)",
		result.Dimension, result.ProcessingTimeMs)

	return result.Embedding, nil
}

// searchSimilarCodes searches for similar codes using vector search
func (e *EmbeddingClassifier) searchSimilarCodes(
	ctx context.Context,
	embedding []float64,
	codeType string,
	threshold float64,
	limit int,
) ([]CodeResult, error) {
	// Use repository method to call Supabase RPC
	matches, err := e.supabaseRepo.MatchCodeEmbeddings(
		ctx,
		embedding,
		codeType,
		threshold,
		limit,
	)
	if err != nil {
		return nil, err
	}

	// Convert to CodeResult
	results := make([]CodeResult, 0, len(matches))
	for _, match := range matches {
		results = append(results, CodeResult{
			Code:        match.Code,
			Description: match.Description,
			Confidence:  match.Similarity,
			Source:      "embedding_similarity",
		})
	}

	if len(results) > 0 {
		e.logger.Printf("📊 [Layer 2] Vector search results for %s: %d matches (top similarity: %.2f)",
			codeType, len(results), results[0].Confidence)
	}

	return results, nil
}

// selectTopCodes selects top N codes from matches
func (e *EmbeddingClassifier) selectTopCodes(matches []CodeResult, limit int) []CodeResult {
	if len(matches) == 0 {
		return []CodeResult{}
	}

	// Already sorted by similarity from database
	if len(matches) > limit {
		return matches[:limit]
	}

	return matches
}

// calculateConfidence calculates overall confidence from matches
func (e *EmbeddingClassifier) calculateConfidence(
	mccMatches, sicMatches, naicsMatches []CodeResult,
) float64 {
	// Start with top MCC match similarity
	baseConfidence := 0.0
	if len(mccMatches) > 0 {
		baseConfidence = mccMatches[0].Confidence
	}

	// Boost if we have strong matches across all types
	if len(mccMatches) > 0 && len(sicMatches) > 0 && len(naicsMatches) > 0 {
		// Check if top matches are all high similarity
		mccTop := mccMatches[0].Confidence
		sicTop := sicMatches[0].Confidence
		naicsTop := naicsMatches[0].Confidence

		if mccTop > 0.85 && sicTop > 0.85 && naicsTop > 0.85 {
			baseConfidence *= 1.10 // +10% boost for strong agreement
		} else if mccTop > 0.80 && sicTop > 0.80 && naicsTop > 0.80 {
			baseConfidence *= 1.05 // +5% boost for good agreement
		}
	}

	// Check agreement between top 3 MCC matches
	if len(mccMatches) >= 3 {
		similarities := []float64{
			mccMatches[0].Confidence,
			mccMatches[1].Confidence,
			mccMatches[2].Confidence,
		}

		// If top 3 are all similar (tight cluster), boost confidence
		maxDiff := similarities[0] - similarities[2]
		if maxDiff < 0.10 {
			baseConfidence *= 1.08 // +8% boost for tight cluster
		}
	}

	// Cap at 0.92 (embeddings alone shouldn't claim >92% confidence)
	if baseConfidence > 0.92 {
		baseConfidence = 0.92
	}

	return baseConfidence
}

