package classification

import (
	"context"
	"fmt"
	"log"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"

	"kyb-platform/internal/classification/cache"
	"kyb-platform/internal/classification/nlp"
	"kyb-platform/internal/classification/repository"
)

// MultiStrategyClassifier combines multiple classification strategies
type MultiStrategyClassifier struct {
	keywordRepo      repository.KeywordRepository
	entityRecognizer *nlp.EntityRecognizer
	topicModeler     *nlp.TopicModeler
	logger           *log.Logger
	calibrator       *ConfidenceCalibrator
	predictiveCache  *cache.PredictiveCache // Phase 2.3: Predictive caching
}

// NewMultiStrategyClassifier creates a new multi-strategy classifier
func NewMultiStrategyClassifier(
	keywordRepo repository.KeywordRepository,
	logger *log.Logger,
) *MultiStrategyClassifier {
	if logger == nil {
		logger = log.Default()
	}

	// Create topic modeler with repository support (if repository implements TopicRepository interface)
	topicModeler := nlp.NewTopicModeler()
	// Check if keywordRepo implements TopicRepository methods
	if topicRepo, ok := keywordRepo.(interface {
		GetIndustryTopicsByKeywords(ctx context.Context, keywords []string) (map[int]float64, error)
		GetTopicAccuracy(ctx context.Context, industryID int, topic string) (float64, error)
	}); ok {
		// Create adapter to match TopicRepository interface
		adapter := &topicRepositoryAdapter{repo: topicRepo}
		topicModeler.SetRepository(adapter)
	}

	// Create classification result cache (Phase 2.3)
	resultCache := cache.NewClassificationResultCache(1*time.Hour, logger)
	
	// Create classifier instance first
	classifier := &MultiStrategyClassifier{
		keywordRepo:      keywordRepo,
		entityRecognizer: nlp.NewEntityRecognizer(),
		topicModeler:     topicModeler,
		logger:           logger,
		calibrator:       NewConfidenceCalibrator(logger),
	}
	
	// Create predictive cache with classifier adapter (after classifier is created)
	classifierAdapter := &classificationPredictorAdapter{classifier: classifier}
	predictiveCache := cache.NewPredictiveCache(resultCache, classifierAdapter, logger)
	classifier.predictiveCache = predictiveCache
	
	return classifier
}

// classificationPredictorAdapter adapts MultiStrategyClassifier to ClassificationPredictor interface
type classificationPredictorAdapter struct {
	classifier *MultiStrategyClassifier
}

func (a *classificationPredictorAdapter) Classify(ctx context.Context, businessName, description, websiteURL string) (*cache.ClassificationPrediction, error) {
	if a.classifier == nil {
		return nil, fmt.Errorf("classifier not initialized")
	}
	
	result, err := a.classifier.ClassifyWithMultiStrategy(ctx, businessName, description, websiteURL)
	if err != nil {
		return nil, err
	}
	
	return &cache.ClassificationPrediction{
		PrimaryIndustry: result.PrimaryIndustry,
		Confidence:      result.Confidence,
		Keywords:        result.Keywords,
		Reasoning:       result.Reasoning,
	}, nil
}

// topicRepositoryAdapter adapts repository methods to TopicRepository interface
type topicRepositoryAdapter struct {
	repo interface {
		GetIndustryTopicsByKeywords(ctx context.Context, keywords []string) (map[int]float64, error)
		GetTopicAccuracy(ctx context.Context, industryID int, topic string) (float64, error)
	}
}

func (a *topicRepositoryAdapter) GetIndustryTopicsByKeywords(ctx context.Context, keywords []string) (map[int]float64, error) {
	return a.repo.GetIndustryTopicsByKeywords(ctx, keywords)
}

func (a *topicRepositoryAdapter) GetTopicAccuracy(ctx context.Context, industryID int, topic string) (float64, error) {
	return a.repo.GetTopicAccuracy(ctx, industryID, topic)
}

// ClassificationStrategy represents a single classification strategy result
type ClassificationStrategy struct {
	StrategyName string            `json:"strategy_name"`
	IndustryID   int               `json:"industry_id"`
	IndustryName string            `json:"industry_name"`
	Score        float64           `json:"score"`
	Confidence   float64           `json:"confidence"`
	Evidence     []string          `json:"evidence"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// MultiStrategyResult represents the combined classification result
type MultiStrategyResult struct {
	PrimaryIndustry   string                 `json:"primary_industry"`
	Confidence        float64                `json:"confidence"`
	Strategies        []ClassificationStrategy `json:"strategies"`
	CombinedScores    map[int]float64       `json:"combined_scores"` // industry_id -> combined score
	Reasoning         string                 `json:"reasoning"`
	ProcessingTime    time.Duration          `json:"processing_time"`
	Keywords          []string               `json:"keywords"`
	Entities          []nlp.Entity           `json:"entities"`
	TopicScores       []nlp.TopicScore        `json:"topic_scores"`
	Method            string                 `json:"method,omitempty"` // Phase 2: "fast_path_keyword", "multi_strategy", "ml_validated"
}

// ClassifyWithMultiStrategy performs classification using multiple strategies
// Enhanced with predictive caching (Phase 2.3)
func (msc *MultiStrategyClassifier) ClassifyWithMultiStrategy(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*MultiStrategyResult, error) {
	startTime := time.Now()
	msc.logger.Printf("🚀 [MultiStrategy] Starting multi-strategy classification for: %s", businessName)

	// Phase 2.3: Check predictive cache first
	if msc.predictiveCache != nil {
		if cached, found := msc.predictiveCache.Get(businessName, description, websiteURL); found {
			msc.logger.Printf("✅ [MultiStrategy] Cache HIT for: %s", businessName)
			return &MultiStrategyResult{
				PrimaryIndustry: cached.PrimaryIndustry,
				Confidence:      cached.Confidence,
				Keywords:        cached.Keywords,
				Reasoning:       cached.Reasoning + " (from cache)",
				ProcessingTime:  time.Since(startTime),
			}, nil
		}
		msc.logger.Printf("📊 [MultiStrategy] Cache MISS for: %s", businessName)
		
		// Trigger predictive preloading in background (non-blocking, best effort)
		// Use a separate goroutine to avoid blocking the main request
		go func() {
			defer func() {
				if r := recover(); r != nil {
					msc.logger.Printf("⚠️ Panic in predictive preload recovered: %v", r)
				}
			}()
			// Use background context with timeout to prevent goroutine leaks
			preloadCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			msc.predictiveCache.PreloadCache(preloadCtx, businessName, description, websiteURL)
		}()
	}

	// Phase 2: Try fast path for obvious cases (<100ms)
	if fastResult, isFastPath := msc.tryFastPath(ctx, businessName, description, websiteURL); isFastPath {
		fastResult.ProcessingTime = time.Since(startTime)
		msc.logger.Printf("⚡ [Phase 2] Fast path succeeded: %s (confidence: %.2f%%, duration: %v)",
			fastResult.PrimaryIndustry, fastResult.Confidence*100, fastResult.ProcessingTime)
		return fastResult, nil
	}

	// Step 1: Extract keywords and entities in parallel
	keywordsChan := make(chan []string, 1)
	keywordsErrChan := make(chan error, 1)
	entitiesChan := make(chan []nlp.Entity, 1)
	var extractionWg sync.WaitGroup

	// Extract keywords in parallel
	// Phase 2: Also extract obvious keywords from description as fallback
	extractionWg.Add(1)
	go func() {
		defer extractionWg.Done()
		keywords, err := msc.extractKeywords(ctx, businessName, websiteURL)
		if err != nil {
			keywordsErrChan <- err
			keywordsChan <- []string{}
			return
		}
		
		msc.logger.Printf("🔍 [MultiStrategy] Initial keywords from business name/website: %v (count: %d)", keywords, len(keywords))
		
		// Phase 2: Always check description for keywords, even if we have some from business name/website
		// This is critical for cases like "Cloud Services Inc" where business name gives generic "services"
		// but description has specific "cloud computing" keywords
		if description != "" {
			descriptionKeywords := msc.extractObviousKeywords("", description, "")
			msc.logger.Printf("🔍 [MultiStrategy] Extracted %d keywords from description '%s': %v", len(descriptionKeywords), description, descriptionKeywords)
			
			if len(descriptionKeywords) > 0 {
				if len(keywords) == 0 {
					// No keywords from business name/website, use description keywords
					msc.logger.Printf("🔍 [MultiStrategy] No keywords from business name/website, using %d keywords from description: %v", len(descriptionKeywords), descriptionKeywords)
					keywords = descriptionKeywords
				} else {
					// Merge description keywords with existing keywords, prioritizing description keywords
					msc.logger.Printf("🔍 [MultiStrategy] Merging %d description keywords with %d existing keywords: %v + %v", len(descriptionKeywords), len(keywords), descriptionKeywords, keywords)
					keywordMap := make(map[string]bool)
					// Add description keywords first (higher priority)
					mergedKeywords := make([]string, 0, len(keywords)+len(descriptionKeywords))
					for _, kw := range descriptionKeywords {
						lowerKw := strings.ToLower(kw)
						if !keywordMap[lowerKw] {
							mergedKeywords = append(mergedKeywords, kw)
							keywordMap[lowerKw] = true
						}
					}
					// Add existing keywords that aren't duplicates
					for _, kw := range keywords {
						lowerKw := strings.ToLower(kw)
						if !keywordMap[lowerKw] {
							mergedKeywords = append(mergedKeywords, kw)
							keywordMap[lowerKw] = true
						}
					}
					keywords = mergedKeywords
					msc.logger.Printf("🔍 [MultiStrategy] Merged keywords (description first): %v", keywords)
				}
			} else {
				msc.logger.Printf("⚠️ [MultiStrategy] No keywords found in description: '%s'", description)
			}
		}
		
		keywordsChan <- keywords
	}()

	// Extract entities in parallel (can start before keywords are ready)
	extractionWg.Add(1)
	go func() {
		defer extractionWg.Done()
		// Use empty keywords initially for entity extraction
		combinedText := msc.combineTextForAnalysis(businessName, description, []string{})
		entities := msc.entityRecognizer.ExtractEntities(combinedText)
		entitiesChan <- entities
	}()

	extractionWg.Wait()
	close(keywordsChan)
	close(keywordsErrChan)
	close(entitiesChan)

	// Get results
	keywords := <-keywordsChan
	if err := <-keywordsErrChan; err != nil {
		msc.logger.Printf("❌ [MultiStrategy] Failed to extract keywords: %v", err)
		return nil, fmt.Errorf("failed to extract keywords: %w", err)
	}

	entities := <-entitiesChan

	if len(keywords) == 0 {
		// Check if this is a known business - use business name to infer industry
		knownBusinessIndustry := msc.getKnownBusinessIndustry(businessName)
		if knownBusinessIndustry != "" {
			msc.logger.Printf("🔍 [MultiStrategy] No keywords extracted, but detected known business '%s' - using industry: %s", businessName, knownBusinessIndustry)
			return &MultiStrategyResult{
				PrimaryIndustry: knownBusinessIndustry,
				Confidence:      0.75, // High confidence for known businesses
				ProcessingTime:  time.Since(startTime),
				Keywords:        []string{},
				Reasoning:       fmt.Sprintf("Known business '%s' classified as %s based on business name", businessName, knownBusinessIndustry),
			}, nil
		}
		
		// Phase 2: Try to extract keywords from description if business name extraction failed
		// This helps with cases like "Cloud Services Inc" where description has "Cloud computing services"
		descriptionKeywords := msc.extractObviousKeywords("", description, "")
		if len(descriptionKeywords) > 0 {
			msc.logger.Printf("🔍 [MultiStrategy] No keywords from business name, but found %d keywords from description: %v", len(descriptionKeywords), descriptionKeywords)
			// Retry classification with description keywords
			keywords = descriptionKeywords
			// Continue with normal flow instead of early return
		} else {
			msc.logger.Printf("⚠️ [MultiStrategy] No keywords extracted from business name or description")
			result := &MultiStrategyResult{
				PrimaryIndustry: "General Business",
				Confidence:      0.30,
				ProcessingTime:  time.Since(startTime),
				Keywords:        []string{},
				Reasoning:       "No keywords extracted from business name or website",
			}
			msc.logger.Printf("✅ [MultiStrategy] Classification completed (early return): %s (confidence: %.2f%%)",
				result.PrimaryIndustry, result.Confidence*100)
			return result, nil
		}
	}

	// Ensure we have keywords before proceeding (from either business name or description)
	if len(keywords) == 0 {
		msc.logger.Printf("⚠️ [MultiStrategy] Still no keywords after description extraction")
		result := &MultiStrategyResult{
			PrimaryIndustry: "General Business",
			Confidence:      0.30,
			ProcessingTime:  time.Since(startTime),
			Keywords:        []string{},
			Reasoning:       "No keywords extracted from business name, description, or website",
		}
		return result, nil
	}

	msc.logger.Printf("📊 [MultiStrategy] Extracted %d keywords and %d entities", len(keywords), len(entities))

	// Step 2: Run all classification strategies in parallel
	strategyChan := make(chan ClassificationStrategy, 4)
	var strategyWg sync.WaitGroup

	// Strategy 1: Keyword-based classification (40% weight) with business name context
	strategyWg.Add(1)
	go func() {
		defer strategyWg.Done()
		strategyCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		strategy := msc.classifyByKeywords(strategyCtx, keywords, businessName)
		if strategy != nil {
			strategyChan <- *strategy
		}
	}()

	// Strategy 2: Entity-based classification (25% weight)
	strategyWg.Add(1)
	go func() {
		defer strategyWg.Done()
		strategyCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		strategy := msc.classifyByEntities(strategyCtx, entities, keywords)
		if strategy != nil {
			strategyChan <- *strategy
		}
	}()

	// Strategy 3: Topic-based classification (20% weight)
	strategyWg.Add(1)
	go func() {
		defer strategyWg.Done()
		strategyCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		// Identify topics (with context for database queries)
		topicScores := msc.topicModeler.IdentifyTopicsWithDetailsContext(strategyCtx, keywords)
		strategy := msc.classifyByTopics(strategyCtx, topicScores)
		if strategy != nil {
			strategyChan <- *strategy
		}
	}()

	// Strategy 4: Co-occurrence-based classification (15% weight)
	strategyWg.Add(1)
	go func() {
		defer strategyWg.Done()
		strategyCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		strategy := msc.classifyByCoOccurrence(strategyCtx, keywords, entities)
		if strategy != nil {
			strategyChan <- *strategy
		}
	}()

	// Wait for all strategies to complete
	strategyWg.Wait()
	close(strategyChan)

	// Collect strategies
	strategies := []ClassificationStrategy{}
	for strategy := range strategyChan {
		strategies = append(strategies, strategy)
	}

	msc.logger.Printf("📊 [MultiStrategy] Completed %d strategies in parallel", len(strategies))

	// Get topic scores for result (needed for metadata)
	topicScores := msc.topicModeler.IdentifyTopicsWithDetailsContext(ctx, keywords)

	// Step 5: Combine strategies with weighted scoring
	combinedScores, primaryIndustry, confidence, reasoning := msc.combineStrategies(strategies)

	// Step 5.5: Apply business name context boost for known businesses with low confidence
	knownBusinessIndustry := msc.getKnownBusinessIndustry(businessName)
	if knownBusinessIndustry != "" && strings.EqualFold(primaryIndustry, knownBusinessIndustry) {
		// Boost confidence for known businesses that match expected industry
		if confidence < 0.70 {
			// Calculate boost to reach at least 0.75 confidence for known businesses
			boost := 0.75 - confidence
			if boost < 0.15 {
				boost = 0.15 // Minimum boost
			}
			if boost > 0.25 {
				boost = 0.25 // Maximum boost to avoid overconfidence
			}
			confidence = confidence + boost
			if confidence > 1.0 {
				confidence = 1.0
			}
			msc.logger.Printf("🔍 [MultiStrategy] Applied known business confidence boost: %.2f%% -> %.2f%% (business: %s, industry: %s)",
				(confidence-boost)*100, confidence*100, businessName, knownBusinessIndustry)
		}
	}

	// Step 6: Apply Phase 2 confidence calibration with 5 factors
	// Extract strategy scores
	strategyScores := make(map[string]float64)
	for _, strategy := range strategies {
		strategyScores[strategy.StrategyName] = strategy.Score
	}

	// Estimate content quality based on confidence and strategy agreement
	contentQuality := 0.7 // Default
	if confidence > 0.8 {
		contentQuality = 0.85
	} else if confidence < 0.5 {
		contentQuality = 0.5
	}

	// Code agreement will be calculated later when codes are available (default to 0.7)
	codeAgreement := 0.7

	// Apply Phase 2 enhanced calibration
	calibratedConfidence := msc.calibrator.CalibrateConfidence(
		strategyScores,
		contentQuality,
		codeAgreement,
		"multi_strategy",
	)
	
	// Use calibrated confidence
	if calibratedConfidence != confidence {
		msc.logger.Printf("📊 [Phase 2] Confidence calibrated: %.2f%% -> %.2f%%",
			confidence*100, calibratedConfidence*100)
		confidence = calibratedConfidence
	}

	msc.logger.Printf("✅ [MultiStrategy] Classification completed: %s (confidence: %.2f%%)",
		primaryIndustry, confidence*100)
	
	result := &MultiStrategyResult{
		PrimaryIndustry:  primaryIndustry,
		Confidence:       confidence, // Use calibrated confidence
		Strategies:       strategies,
		CombinedScores:   combinedScores,
		Reasoning:        reasoning,
		ProcessingTime:    time.Since(startTime),
		Keywords:         keywords,
		Method:           "multi_strategy", // Phase 2: Add method field
		Entities:         entities,
		TopicScores:      topicScores,
	}

	// Phase 2.3: Cache the result for future requests
	if msc.predictiveCache != nil {
		cachedResult := &cache.CachedClassificationResult{
			PrimaryIndustry: result.PrimaryIndustry,
			Confidence:      result.Confidence,
			Keywords:        result.Keywords,
			Reasoning:       result.Reasoning,
		}
		msc.predictiveCache.Set(businessName, description, websiteURL, cachedResult)
	}
	
	return result, nil
}

// tryFastPath attempts fast path classification for obvious cases (Phase 2)
// Returns result and true if fast path succeeded, nil and false otherwise
func (msc *MultiStrategyClassifier) tryFastPath(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*MultiStrategyResult, bool) {
	// Extract obvious keywords from business name, description, and URL
	obviousKeywords := msc.extractObviousKeywords(businessName, description, websiteURL)

	if len(obviousKeywords) == 0 {
		return nil, false // No obvious keywords found
	}

	// Check each obvious keyword for direct industry match
	for _, keyword := range obviousKeywords {
		// Query for high-confidence keyword matches (lower threshold for fast path: 0.70+)
		matches := msc.keywordRepo.GetIndustriesByKeyword(ctx, keyword, 0.70)

		if len(matches) > 0 {
			// Found high-confidence match via obvious keyword
			industry := matches[0]

			// Create a strategy result for fast path
			strategy := ClassificationStrategy{
				StrategyName: "keyword",
				IndustryName: industry.Name,
				Score:        0.95,
				Confidence:  0.92,
				Evidence:    []string{keyword},
			}

			result := &MultiStrategyResult{
				PrimaryIndustry: industry.Name,
				Confidence:      0.92, // High confidence for fast path
				Method:          "fast_path_keyword",
				Keywords:        []string{keyword},
				Strategies:      []ClassificationStrategy{strategy},
				Reasoning:       fmt.Sprintf("Fast path: obvious keyword '%s' matched industry '%s'", keyword, industry.Name),
				ProcessingTime:  0, // Will be set by caller
			}

			msc.logger.Printf("⚡ [Phase 2] Fast path triggered by keyword '%s' -> industry '%s'", keyword, industry.Name)
			return result, true
		}
	}

	// No fast path match found
	return nil, false
}

// extractObviousKeywords extracts obvious industry keywords from business information (Phase 2)
func (msc *MultiStrategyClassifier) extractObviousKeywords(businessName, description, websiteURL string) []string {
	obviousKeywords := []string{}

	// Common obvious industry keywords map
	obviousKeywordMap := map[string]bool{
		// Food & Beverage
		"restaurant": true, "cafe": true, "coffee": true, "bakery": true,
		"bar": true, "pub": true, "brewery": true, "winery": true,
		"pizzeria": true, "diner": true, "bistro": true, "pizza": true,
		"dining": true, "food": true, "eatery": true,

		// Retail
		"shop": true, "store": true, "boutique": true, "market": true,
		"mall": true, "retail": true, "merchandise": true,

		// Professional Services
		"law firm": true, "attorney": true, "lawyer": true, "legal": true,
		"dentist": true, "dental": true, "orthodontist": true,
		"doctor": true, "physician": true, "medical": true, "clinic": true,
		"accountant": true, "accounting": true, "cpa": true,
		"consulting": true, "consultant": true,
		
		// Technology & Cloud
		"cloud": true, "computing": true, "software": true, "technology": true,
		"tech": true, "it": true, "saas": true, "platform": true,

		// Insurance & Finance
		"insurance": true, "brokerage": true, "bank": true, "banking": true,
		"credit": true, "mortgage": true, "loan": true, "financial": true,
		"investment": true, "wealth": true, "trading": true, "finance": true,

		// Healthcare
		"hospital": true, "healthcare": true, "pharmacy": true, "diagnostic": true,
		"laboratory": true, "dialysis": true, "veterinary": true, "vet": true,
		"health": true, "patient": true, "care": true,

		// Agriculture
		"farm": true, "farming": true, "agriculture": true, "agricultural": true,
		"crop": true, "livestock": true, "ranch": true, "agri": true,

		// Home Services
		"plumber": true, "plumbing": true, "electrician": true, "electrical": true,
		"contractor": true, "construction": true, "roofing": true,
		"hvac": true, "heating": true, "cooling": true, "landscaping": true,

		// Automotive
		"auto repair": true, "mechanic": true, "car wash": true,
		"dealership": true, "automotive": true, "auto": true,

		// Personal Services
		"salon": true, "barber": true, "spa": true, "gym": true,
		"fitness": true, "yoga": true, "wellness": true,

		// Hospitality
		"hotel": true, "motel": true, "inn": true, "resort": true,
		"lodging": true, "accommodation": true,

		// Education
		"school": true, "university": true, "college": true,
		"tutoring": true, "academy": true, "education": true,
	}

	// Combine all text sources
	// Ensure spaces around text for proper regex matching
	combinedText := strings.ToLower(" " + businessName + " " + description + " " + websiteURL + " ")

	// Check for keyword matches
	for keyword := range obviousKeywordMap {
		if strings.Contains(combinedText, keyword) {
			// Check if it's a whole word match (not part of another word)
			// Simple check: surrounded by word boundaries or at start/end
			pattern := "(^|[^a-z])" + keyword + "([^a-z]|$)"
			matched, _ := regexp.MatchString(pattern, combinedText)
			if matched {
				obviousKeywords = append(obviousKeywords, keyword)
			}
		}
	}

	// Also check individual words in business name (highest signal)
	businessNameWords := strings.Fields(strings.ToLower(businessName))
	for _, word := range businessNameWords {
		// Remove common suffixes/prefixes
		word = strings.Trim(word, ".,!?;:'\"()[]{}")
		if obviousKeywordMap[word] {
			// Check if not already added
			found := false
			for _, existing := range obviousKeywords {
				if existing == word {
					found = true
					break
				}
			}
			if !found {
				obviousKeywords = append(obviousKeywords, word)
			}
		}
	}

	return obviousKeywords
}

// extractKeywords extracts keywords using the repository and filters them for relevance
func (msc *MultiStrategyClassifier) extractKeywords(ctx context.Context, businessName, websiteURL string) ([]string, error) {
	// Use repository's ClassifyBusiness to get keywords (already includes NER and topic modeling)
	result, err := msc.keywordRepo.ClassifyBusiness(ctx, businessName, websiteURL)
	if err != nil {
		return nil, fmt.Errorf("failed to classify business: %w", err)
	}

	if result == nil {
		return []string{}, nil
	}

	// Filter keywords for relevance based on business name and context
	filteredKeywords := msc.filterRelevantKeywords(result.Keywords, businessName, websiteURL)
	
	if len(filteredKeywords) < len(result.Keywords) {
		msc.logger.Printf("🔍 [MultiStrategy] Filtered %d keywords to %d relevant keywords (removed %d misleading keywords)",
			len(result.Keywords), len(filteredKeywords), len(result.Keywords)-len(filteredKeywords))
	}

	return filteredKeywords, nil
}

// filterRelevantKeywords filters out misleading keywords based on business name and context
func (msc *MultiStrategyClassifier) filterRelevantKeywords(keywords []string, businessName, websiteURL string) []string {
	if len(keywords) == 0 {
		return keywords
	}

	// Normalize business name for matching
	businessNameLower := strings.ToLower(businessName)
	
	// Known multi-industry businesses that should be classified by primary industry
	knownBusinesses := map[string]string{
		"amazon":      "retail",
		"microsoft":   "technology",
		"google":      "technology",
		"apple":       "technology",
		"facebook":    "technology",
		"meta":        "technology",
		"walmart":     "retail",
		"target":      "retail",
		"costco":      "retail",
		"mayo clinic": "healthcare",
		"cleveland clinic": "healthcare",
		"johns hopkins": "healthcare",
	}
	
	// Check if this is a known business with a primary industry
	primaryIndustry := ""
	for knownBusiness, industry := range knownBusinesses {
		if strings.Contains(businessNameLower, knownBusiness) {
			primaryIndustry = industry
			msc.logger.Printf("🔍 [MultiStrategy] Detected known business '%s' with primary industry: %s", businessName, primaryIndustry)
			break
		}
	}

	// Keywords that are misleading when extracted from mobile apps or secondary services
	misleadingKeywords := map[string]string{
		// Mobile app keywords that don't represent primary business
		"mobile kitchen": "food_trucks",
		"mobile food":    "food_trucks",
		"mobile dining": "food_trucks",
		"food truck":     "food_trucks",
		"food delivery": "food_trucks",
		// Food service keywords that are misleading for retail businesses
		"food service": "food_beverage",
		"table service": "food_beverage",
		"quick service": "food_beverage",
		"self service": "food_beverage",
		"counter service": "food_beverage",
		// Add more as needed
	}

	// Filter keywords
	filtered := make([]string, 0, len(keywords))
	removedCount := 0
	
	for _, keyword := range keywords {
		keywordLower := strings.ToLower(keyword)
		shouldRemove := false
		
		// Remove misleading keywords if they don't match the primary industry
		if primaryIndustry != "" {
			for misleadingKW, misleadingIndustry := range misleadingKeywords {
				// Normalize industry names for comparison
				misleadingIndustryNormalized := strings.ToLower(strings.ReplaceAll(misleadingIndustry, "_", " "))
				primaryIndustryNormalized := strings.ToLower(primaryIndustry)
				
				if strings.Contains(keywordLower, misleadingKW) {
					// Check if the misleading industry matches the primary industry
					if !strings.Contains(primaryIndustryNormalized, misleadingIndustryNormalized) &&
					   !strings.Contains(misleadingIndustryNormalized, primaryIndustryNormalized) {
						msc.logger.Printf("🔍 [MultiStrategy] Removing misleading keyword '%s' (industry: %s, primary: %s)",
							keyword, misleadingIndustry, primaryIndustry)
						shouldRemove = true
						removedCount++
						break
					}
				}
			}
		}
		
		// Remove very generic keywords that don't add value
		genericKeywords := []string{"content", "media", "inc", "corp", "llc", "ltd", "company", "business"}
		for _, generic := range genericKeywords {
			if keywordLower == generic {
				shouldRemove = true
				removedCount++
				break
			}
		}
		
		if !shouldRemove {
			filtered = append(filtered, keyword)
		}
	}

	if removedCount > 0 {
		msc.logger.Printf("🔍 [MultiStrategy] Removed %d misleading/generic keywords", removedCount)
	}

	return filtered
}

// getKnownBusinessIndustry returns the industry for a known business based on business name
func (msc *MultiStrategyClassifier) getKnownBusinessIndustry(businessName string) string {
	businessNameLower := strings.ToLower(businessName)
	
	knownBusinesses := map[string]string{
		"amazon":      "Retail",
		"microsoft":   "Technology",
		"google":      "Technology",
		"apple":       "Technology",
		"facebook":    "Technology",
		"meta":        "Technology",
		"walmart":     "Retail",
		"target":      "Retail",
		"costco":      "Retail",
		"mayo clinic": "Healthcare",
		"cleveland clinic": "Healthcare",
		"johns hopkins": "Healthcare",
	}
	
	for knownBusiness, industry := range knownBusinesses {
		if strings.Contains(businessNameLower, knownBusiness) {
			return industry
		}
	}
	
	return ""
}

// combineTextForAnalysis combines all text sources for entity extraction
func (msc *MultiStrategyClassifier) combineTextForAnalysis(businessName, description string, keywords []string) string {
	var builder strings.Builder
	
	if businessName != "" {
		builder.WriteString(businessName)
		builder.WriteString(" ")
	}
	
	if description != "" {
		builder.WriteString(description)
		builder.WriteString(" ")
	}
	
	for _, kw := range keywords {
		builder.WriteString(kw)
		builder.WriteString(" ")
	}
	
	return builder.String()
}

// classifyByKeywords performs keyword-based classification with business name context
// Enhanced with database optimizations: trigram indexes, full-text search, context timeout
func (msc *MultiStrategyClassifier) classifyByKeywords(ctx context.Context, keywords []string, businessName string) *ClassificationStrategy {
	msc.logger.Printf("🔍 [MultiStrategy] Strategy 1: Keyword-based classification (enhanced)")
	
	if len(keywords) == 0 {
		return nil
	}

	// Execute with context timeout (2s as per plan)
	strategyCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Use repository's keyword classification with enhanced trigram similarity support
	// Note: The repository method will use trigram indexes via database function if available
	classification, err := msc.keywordRepo.ClassifyBusinessByKeywords(strategyCtx, keywords)
	if err != nil {
		// Check if error is due to timeout
		if strategyCtx.Err() == context.DeadlineExceeded {
			msc.logger.Printf("⚠️ [MultiStrategy] Keyword classification timed out after 2s")
		} else {
			msc.logger.Printf("⚠️ [MultiStrategy] Keyword classification failed: %v", err)
		}
		return nil
	}

	if classification == nil {
		return nil
	}

	// Get industry ID from classification
	industryID := 0
	if classification.Industry != nil {
		industryID = classification.Industry.ID
	}

	// Apply business name context boost/penalty
	adjustedScore := classification.Confidence
	adjustedScore = msc.applyBusinessNameContext(keywords, classification.Industry.Name, adjustedScore, businessName)

	return &ClassificationStrategy{
		StrategyName: "keyword",
		IndustryID:    industryID,
		IndustryName: classification.Industry.Name,
		Score:        adjustedScore,
		Confidence:   adjustedScore,
		Evidence:     keywords[:minIntValue(10, len(keywords))], // Top 10 keywords
		Metadata: map[string]string{
			"keyword_count": fmt.Sprintf("%d", len(keywords)),
			"original_confidence": fmt.Sprintf("%.2f", classification.Confidence),
			"adjusted_confidence": fmt.Sprintf("%.2f", adjustedScore),
			"timeout_seconds": "2",
		},
	}
}

// applyBusinessNameContext adjusts confidence based on business name and industry match
func (msc *MultiStrategyClassifier) applyBusinessNameContext(keywords []string, industryName string, baseConfidence float64, businessName string) float64 {
	// Known business-to-industry mappings for context-aware scoring
	businessIndustryMap := map[string]string{
		"amazon":      "Retail",
		"microsoft":   "Technology",
		"google":      "Technology",
		"apple":       "Technology",
		"facebook":    "Technology",
		"meta":        "Technology",
		"walmart":     "Retail",
		"target":      "Retail",
		"costco":      "Retail",
		"mayo":        "Healthcare",
		"clinic":      "Healthcare",
		"hospital":    "Healthcare",
		"medical":     "Healthcare",
	}
	
	// Check if business name matches known businesses
	businessNameLower := strings.ToLower(businessName)
	industryLower := strings.ToLower(industryName)
	
	// Boost confidence if industry matches known business patterns
	for knownBusiness, expectedIndustry := range businessIndustryMap {
		if strings.Contains(businessNameLower, knownBusiness) {
			if strings.Contains(industryLower, strings.ToLower(expectedIndustry)) {
				// Boost confidence for correct match
				boost := 0.15
				adjusted := baseConfidence + boost
				if adjusted > 1.0 {
					adjusted = 1.0
				}
				msc.logger.Printf("🔍 [MultiStrategy] Applied business name context boost: %.2f -> %.2f (business: %s, industry: %s)",
					baseConfidence, adjusted, businessName, expectedIndustry)
				return adjusted
			} else {
				// Strong penalty for mismatch - reduce confidence significantly
				penalty := 0.40 // Increased penalty from 0.20 to 0.40
				adjusted := baseConfidence - penalty
				if adjusted < 0.0 {
					adjusted = 0.0
				}
				msc.logger.Printf("🔍 [MultiStrategy] Applied business name context penalty: %.2f -> %.2f (business: %s, expected: %s, got: %s)",
					baseConfidence, adjusted, businessName, expectedIndustry, industryName)
				
				// If penalty is severe, try to find the correct industry by boosting retail keywords
				if strings.Contains(strings.ToLower(expectedIndustry), "retail") {
					// Check if keywords contain retail-related terms
					keywordsLower := strings.ToLower(strings.Join(keywords, " "))
					retailKeywords := []string{"retail", "store", "shop", "marketplace", "ecommerce", "e-commerce", "online", "merchandise", "product", "shopping"}
					hasRetailKeywords := false
					for _, retailKW := range retailKeywords {
						if strings.Contains(keywordsLower, retailKW) {
							hasRetailKeywords = true
							break
						}
					}
					
					if hasRetailKeywords {
						// Boost confidence if retail keywords are present
						boost := 0.25
						adjusted = adjusted + boost
						if adjusted > 1.0 {
							adjusted = 1.0
						}
						msc.logger.Printf("🔍 [MultiStrategy] Applied retail keyword boost after penalty: %.2f -> %.2f (retail keywords found)",
							adjusted-boost, adjusted)
					}
				}
				
				return adjusted
			}
		}
	}
	
	return baseConfidence
}

// classifyByEntities performs entity-based classification
func (msc *MultiStrategyClassifier) classifyByEntities(ctx context.Context, entities []nlp.Entity, keywords []string) *ClassificationStrategy {
	msc.logger.Printf("🔍 [MultiStrategy] Strategy 2: Entity-based classification")
	
	if len(entities) == 0 {
		return nil
	}

	// Extract entity keywords
	entityKeywords := msc.entityRecognizer.GetEntityKeywords(entities)
	
	// Classify using entity keywords
	classification, err := msc.keywordRepo.ClassifyBusinessByKeywords(ctx, entityKeywords)
	if err != nil {
		msc.logger.Printf("⚠️ [MultiStrategy] Entity classification failed: %v", err)
		return nil
	}

	if classification == nil {
		return nil
	}

	// Calculate entity-based confidence
	entityConfidence := msc.calculateEntityConfidence(entities, classification.Industry.Name)

	industryID := 0
	if classification.Industry != nil {
		industryID = classification.Industry.ID
	}

	return &ClassificationStrategy{
		StrategyName: "entity",
		IndustryID:   industryID,
		IndustryName: classification.Industry.Name,
		Score:        entityConfidence,
		Confidence:   entityConfidence,
		Evidence:     entityKeywords[:minIntValue(10, len(entityKeywords))],
		Metadata: map[string]string{
			"entity_count": fmt.Sprintf("%d", len(entities)),
			"entity_types": msc.getEntityTypes(entities),
		},
	}
}

// classifyByTopics performs topic-based classification
func (msc *MultiStrategyClassifier) classifyByTopics(ctx context.Context, topicScores []nlp.TopicScore) *ClassificationStrategy {
	msc.logger.Printf("🔍 [MultiStrategy] Strategy 3: Topic-based classification")
	
	if len(topicScores) == 0 {
		return nil
	}

	// Get top topic (highest score)
	topTopic := topicScores[0]
	
	// Map industry ID to industry name (this may require repository lookup)
	industryName := msc.getIndustryNameFromID(topTopic.IndustryID)
	if industryName == "" {
		industryName = fmt.Sprintf("Industry %d", topTopic.IndustryID)
	}

	return &ClassificationStrategy{
		StrategyName: "topic",
		IndustryID:   topTopic.IndustryID,
		IndustryName: industryName,
		Score:        topTopic.Score,
		Confidence:   topTopic.Score,
		Evidence:     topTopic.Keywords[:minIntValue(10, len(topTopic.Keywords))],
		Metadata: map[string]string{
			"topic_count": fmt.Sprintf("%d", len(topicScores)),
		},
	}
}

// classifyByCoOccurrence performs co-occurrence-based classification with relationship analysis
// Enhanced with database-driven pattern matching
func (msc *MultiStrategyClassifier) classifyByCoOccurrence(ctx context.Context, keywords []string, entities []nlp.Entity) *ClassificationStrategy {
	msc.logger.Printf("🔍 [MultiStrategy] Strategy 4: Co-occurrence-based classification (enhanced)")

	// Step 1: Analyze co-occurrence patterns
	patterns := msc.analyzeCoOccurrencePatterns(keywords, entities)
	if len(patterns) == 0 {
		msc.logger.Printf("⚠️ [MultiStrategy] No co-occurrence patterns found")
		return nil
	}

	msc.logger.Printf("📊 [MultiStrategy] Generated %d co-occurrence patterns", len(patterns))

	// Step 2: Query database for industry patterns
	patternResults, err := msc.keywordRepo.FindIndustriesByPatterns(ctx, patterns)
	if err != nil {
		msc.logger.Printf("⚠️ [MultiStrategy] Pattern query failed: %v, falling back to keyword classification", err)
		// Fallback to basic keyword classification
		return msc.classifyByCoOccurrenceFallback(ctx, keywords, entities)
	}

	if len(patternResults) == 0 {
		msc.logger.Printf("⚠️ [MultiStrategy] No industries found for patterns, falling back")
		return msc.classifyByCoOccurrenceFallback(ctx, keywords, entities)
	}

	// Step 3: Find best industry match based on pattern analysis
	bestResult := patternResults[0] // Results are already sorted by pattern_matches and avg_score

	// Calculate confidence based on pattern matches and scores
	// More pattern matches = higher confidence
	// Higher avg_score = higher confidence
	patternMatchRatio := float64(bestResult.PatternMatches) / float64(len(patterns))
	baseConfidence := (patternMatchRatio * 0.6) + (bestResult.AvgScore * 0.4)

	// Boost confidence if multiple patterns match
	if bestResult.PatternMatches >= 3 {
		baseConfidence = math.Min(1.0, baseConfidence*1.2) // 20% boost for 3+ matches
	} else if bestResult.PatternMatches >= 2 {
		baseConfidence = math.Min(1.0, baseConfidence*1.1) // 10% boost for 2 matches
	}

	// Ensure minimum confidence threshold
	if baseConfidence < 0.4 {
		baseConfidence = 0.4
	}

	// Get industry details
	industry, err := msc.keywordRepo.GetIndustryByID(ctx, bestResult.IndustryID)
	if err != nil {
		msc.logger.Printf("⚠️ [MultiStrategy] Failed to get industry %d: %v", bestResult.IndustryID, err)
		return msc.classifyByCoOccurrenceFallback(ctx, keywords, entities)
	}

	msc.logger.Printf("✅ [MultiStrategy] Co-occurrence classification: %s (confidence: %.2f, patterns: %d)",
		industry.Name, baseConfidence, bestResult.PatternMatches)

	return &ClassificationStrategy{
		StrategyName: "co_occurrence",
		IndustryID:   bestResult.IndustryID,
		IndustryName: industry.Name,
		Score:        baseConfidence,
		Confidence:   baseConfidence,
		Evidence:     bestResult.MatchedPatterns[:minIntValue(10, len(bestResult.MatchedPatterns))],
		Metadata: map[string]string{
			"keyword_count":     fmt.Sprintf("%d", len(keywords)),
			"entity_count":      fmt.Sprintf("%d", len(entities)),
			"pattern_count":     fmt.Sprintf("%d", len(patterns)),
			"pattern_matches":   fmt.Sprintf("%d", bestResult.PatternMatches),
			"avg_pattern_score": fmt.Sprintf("%.2f", bestResult.AvgScore),
		},
	}
}

// analyzeCoOccurrencePatterns generates keyword pairs and entity-keyword pairs for pattern analysis
func (msc *MultiStrategyClassifier) analyzeCoOccurrencePatterns(keywords []string, entities []nlp.Entity) []string {
	patterns := make([]string, 0)
	seen := make(map[string]bool)

	// Normalize function for consistent pair format
	normalizePair := func(kw1, kw2 string) string {
		kw1Lower := strings.ToLower(strings.TrimSpace(kw1))
		kw2Lower := strings.ToLower(strings.TrimSpace(kw2))
		if kw1Lower < kw2Lower {
			return kw1Lower + "|" + kw2Lower
		}
		return kw2Lower + "|" + kw1Lower
	}

	// Generate keyword pairs (keyword-keyword)
	for i := 0; i < len(keywords)-1; i++ {
		for j := i + 1; j < len(keywords); j++ {
			pair := normalizePair(keywords[i], keywords[j])
			if !seen[pair] {
				patterns = append(patterns, pair)
				seen[pair] = true
			}
		}
	}

	// Generate entity-keyword pairs
	entityKeywords := msc.entityRecognizer.GetEntityKeywords(entities)
	for _, entityKw := range entityKeywords {
		for _, keyword := range keywords {
			pair := normalizePair(entityKw, keyword)
			if !seen[pair] {
				patterns = append(patterns, pair)
				seen[pair] = true
			}
		}
	}

	// Generate entity-entity pairs (if multiple entities)
	if len(entities) > 1 {
		for i := 0; i < len(entities)-1; i++ {
			for j := i + 1; j < len(entities); j++ {
				pair := normalizePair(entities[i].Text, entities[j].Text)
				if !seen[pair] {
					patterns = append(patterns, pair)
					seen[pair] = true
				}
			}
		}
	}

	return patterns
}

// classifyByCoOccurrenceFallback provides fallback classification when pattern matching fails
func (msc *MultiStrategyClassifier) classifyByCoOccurrenceFallback(ctx context.Context, keywords []string, entities []nlp.Entity) *ClassificationStrategy {
	// Combine keywords and entity keywords
	allKeywords := make([]string, 0, len(keywords)+len(entities))
	allKeywords = append(allKeywords, keywords...)

	entityKeywords := msc.entityRecognizer.GetEntityKeywords(entities)
	allKeywords = append(allKeywords, entityKeywords...)

	// Use keyword classification as fallback
	classification, err := msc.keywordRepo.ClassifyBusinessByKeywords(ctx, allKeywords)
	if err != nil {
		msc.logger.Printf("⚠️ [MultiStrategy] Co-occurrence fallback failed: %v", err)
		return nil
	}

	if classification == nil {
		return nil
	}

	// Calculate co-occurrence confidence (slightly lower than keyword-only)
	coOccurrenceConfidence := classification.Confidence * 0.85

	industryID := 0
	if classification.Industry != nil {
		industryID = classification.Industry.ID
	}

	return &ClassificationStrategy{
		StrategyName: "co_occurrence",
		IndustryID:   industryID,
		IndustryName: classification.Industry.Name,
		Score:        coOccurrenceConfidence,
		Confidence:   coOccurrenceConfidence,
		Evidence:     allKeywords[:minIntValue(10, len(allKeywords))],
		Metadata: map[string]string{
			"keyword_count": fmt.Sprintf("%d", len(keywords)),
			"entity_count":  fmt.Sprintf("%d", len(entities)),
			"fallback":      "true",
		},
	}
}

// combineStrategies combines multiple strategies using simple weighted average
// Simplified implementation per Phase 1.5 - no complex fallbacks
func (msc *MultiStrategyClassifier) combineStrategies(strategies []ClassificationStrategy) (map[int]float64, string, float64, string) {
	// Strategy weights (fixed, based on accuracy)
	weights := map[string]float64{
		"keyword":       0.40, // 40% - highest accuracy
		"entity":        0.25, // 25% - good accuracy
		"topic":         0.20, // 20% - moderate accuracy
		"co_occurrence": 0.15, // 15% - supporting evidence
	}

	// Combine scores using weighted average
	combinedScores := make(map[int]float64)
	industryNames := make(map[int]string)
	totalWeight := 0.0

	// Track which strategies contributed to each industry
	strategyContributions := make(map[int][]string)

	for _, strategy := range strategies {
		weight := weights[strategy.StrategyName]
		if weight == 0 {
			// Skip unknown strategies (no default weight)
			continue
		}

		// Calculate weighted score: strategy score * confidence * weight
		score := strategy.Score * strategy.Confidence
		weightedScore := score * weight

		combinedScores[strategy.IndustryID] += weightedScore
		totalWeight += weight

		// Store industry name
		if industryNames[strategy.IndustryID] == "" {
			industryNames[strategy.IndustryID] = strategy.IndustryName
		}

		// Track strategy contributions
		strategyContributions[strategy.IndustryID] = append(
			strategyContributions[strategy.IndustryID],
			fmt.Sprintf("%s(%.2f)", strategy.StrategyName, strategy.Confidence),
		)
	}

	// Normalize scores by total weight
	if totalWeight > 0 {
		for industryID := range combinedScores {
			combinedScores[industryID] /= totalWeight
		}
	}

	// Phase 2: Boost specific industries and penalize generic ones
	combinedScores = msc.boostSpecificIndustries(combinedScores, industryNames)

	// Phase 2: If no strategies returned results, try to infer from keywords
	// This helps with ambiguous cases like "ABC Corporation" where no keywords match
	if len(combinedScores) == 0 {
		msc.logger.Printf("⚠️ [Phase 2] No strategies returned results, combinedScores is empty")
		// This will default to General Business below, which is expected
	}

	// Find primary industry (highest score)
	var primaryIndustryID int
	var maxScore float64
	for industryID, score := range combinedScores {
		if score > maxScore {
			maxScore = score
			primaryIndustryID = industryID
		}
	}

	// Get industry name
	primaryIndustry := industryNames[primaryIndustryID]
	if primaryIndustry == "" {
		primaryIndustry = "General Business"
		primaryIndustryID = 26 // Default industry ID
		// If we have no scores, set a minimum score for General Business
		if maxScore == 0 {
			maxScore = 0.30 // Minimum confidence for General Business
		}
	}

	// Phase 2: Prefer specific industries over generic
	primaryIndustryID, primaryIndustry, maxScore = msc.selectBestIndustry(
		combinedScores,
		industryNames,
		primaryIndustryID,
		primaryIndustry,
		maxScore,
	)

	// Calculate final confidence
	confidence := maxScore
	if confidence < 0.35 {
		confidence = 0.35 // Minimum confidence threshold
	}
	if confidence > 1.0 {
		confidence = 1.0 // Cap at 1.0
	}

	// Generate clear reasoning
	reasoning := msc.generateReasoning(strategies, primaryIndustryID, confidence, strategyContributions)

	return combinedScores, primaryIndustry, confidence, reasoning
}

// boostSpecificIndustries boosts confidence of specific industries (Phase 2)
func (msc *MultiStrategyClassifier) boostSpecificIndustries(
	combinedScores map[int]float64,
	industryNames map[int]string,
) map[int]float64 {
	// Define generic industries to avoid
	genericIndustries := map[string]bool{
		"General Business":      true,
		"Other Services":        true,
		"Miscellaneous":         true,
		"Business Services":     true,
		"Professional Services": true,
		"General Merchandise":   true,
		"Other":                 true,
	}

	// Boost specific industries by 0.10 (increased from 0.05) to make them more competitive
	specificBoost := 0.10
	for industryID, score := range combinedScores {
		industryName := industryNames[industryID]
		if !genericIndustries[industryName] {
			combinedScores[industryID] = math.Min(score+specificBoost, 0.95)
		}
	}

	return combinedScores
}

// selectBestIndustry selects the best industry, preferring specific over generic (Phase 2)
func (msc *MultiStrategyClassifier) selectBestIndustry(
	combinedScores map[int]float64,
	industryNames map[int]string,
	currentIndustryID int,
	currentIndustry string,
	currentScore float64,
) (int, string, float64) {
	// Define generic industries
	genericIndustries := map[string]bool{
		"General Business":      true,
		"Other Services":        true,
		"Miscellaneous":         true,
		"Business Services":     true,
		"Professional Services": true,
		"General Merchandise":   true,
		"Other":                 true,
	}

	// If current industry is generic, always try to find specific alternative (even with higher confidence)
	if genericIndustries[currentIndustry] {
		msc.logger.Printf("⚠️ [Phase 2] Top candidate is generic '%s' with confidence (%.2f), checking for specific alternatives",
			currentIndustry, currentScore)

		// Look for more specific alternative - be more aggressive in preferring specific industries
		bestSpecificID := 0
		bestSpecificName := ""
		bestSpecificScore := 0.0

		for industryID, score := range combinedScores {
			industryName := industryNames[industryID]
			if !genericIndustries[industryName] {
				// More aggressive: Prefer specific industry if:
				// 1. It's within 0.30 of generic confidence (even more lenient)
				// 2. OR if specific industry has at least 0.30 confidence (lowered from 0.50)
				// 3. OR if generic confidence is low (<0.60) and specific has any score > 0.20
				shouldPrefer := false
				if (currentScore - score) < 0.30 && score > bestSpecificScore {
					shouldPrefer = true
				} else if score >= 0.30 && score > bestSpecificScore {
					shouldPrefer = true
				} else if currentScore < 0.60 && score > 0.20 && score > bestSpecificScore {
					// If generic is low confidence, prefer any specific industry with reasonable score
					shouldPrefer = true
				}
				
				if shouldPrefer {
					bestSpecificID = industryID
					bestSpecificName = industryName
					bestSpecificScore = score
				}
			}
		}

		if bestSpecificName != "" {
			msc.logger.Printf("✅ [Phase 2] Preferring specific industry '%s' (%.2f) over generic '%s' (%.2f)",
				bestSpecificName, bestSpecificScore, currentIndustry, currentScore)
			return bestSpecificID, bestSpecificName, bestSpecificScore
		} else {
			// If no specific alternative found, but generic confidence is very low, try harder
			if currentScore < 0.60 {
				msc.logger.Printf("⚠️ [Phase 2] Generic '%s' has low confidence (%.2f) but no specific alternative found - may need better keyword matching",
					currentIndustry, currentScore)
				// Last resort: find ANY specific industry with score > 0.15
				for industryID, score := range combinedScores {
					industryName := industryNames[industryID]
					if !genericIndustries[industryName] && score > 0.15 && score > bestSpecificScore {
						bestSpecificID = industryID
						bestSpecificName = industryName
						bestSpecificScore = score
					}
				}
				if bestSpecificName != "" {
					msc.logger.Printf("✅ [Phase 2] Last resort: Using specific industry '%s' (%.2f) instead of generic '%s' (%.2f)",
						bestSpecificName, bestSpecificScore, currentIndustry, currentScore)
					return bestSpecificID, bestSpecificName, bestSpecificScore
				}
			}
		}
	}

	// For generic industries, require higher confidence threshold
	if genericIndustries[currentIndustry] {
		if currentScore < 0.70 {
			msc.logger.Printf("⚠️ [Phase 2] Generic industry '%s' requires ≥0.70 confidence, got %.2f - may need fallback",
				currentIndustry, currentScore)
			// Still return it, but log warning
		}
	}

	return currentIndustryID, currentIndustry, currentScore
}

// generateReasoning generates clear reasoning for the classification result
func (msc *MultiStrategyClassifier) generateReasoning(
	strategies []ClassificationStrategy,
	primaryIndustryID int,
	confidence float64,
	strategyContributions map[int][]string,
) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Combined %d classification strategies using weighted average. ", len(strategies)))

	// List strategy contributions
	if contributions, exists := strategyContributions[primaryIndustryID]; exists && len(contributions) > 0 {
		builder.WriteString("Contributions: ")
		for i, contrib := range contributions {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(contrib)
		}
		builder.WriteString(". ")
	}

	// Add strategy details
	builder.WriteString("Strategies: ")
	for i, strategy := range strategies {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%s(%.2f)", strategy.StrategyName, strategy.Confidence))
	}

	builder.WriteString(fmt.Sprintf(". Final confidence: %.2f", confidence))

	return builder.String()
}

// calculateEntityConfidence calculates confidence based on entity relevance
func (msc *MultiStrategyClassifier) calculateEntityConfidence(entities []nlp.Entity, industryName string) float64 {
	if len(entities) == 0 {
		return 0.0
	}

	// Count relevant entities (business types, industries, services)
	relevantCount := 0
	totalConfidence := 0.0

	for _, entity := range entities {
		// Check if entity type is relevant
		if entity.Type == nlp.EntityTypeBusinessType || entity.Type == nlp.EntityTypeIndustry || entity.Type == nlp.EntityTypeService {
			relevantCount++
			totalConfidence += entity.Confidence
		}
	}

	if relevantCount == 0 {
		return 0.0
	}

	// Average confidence weighted by relevance
	avgConfidence := totalConfidence / float64(relevantCount)
	
	// Boost confidence based on number of relevant entities
	entityBoost := minFloat(float64(relevantCount)/10.0, 0.2) // Up to 20% boost
	
	return minFloat(avgConfidence+entityBoost, 1.0)
}

// getEntityTypes returns a comma-separated list of entity types
func (msc *MultiStrategyClassifier) getEntityTypes(entities []nlp.Entity) string {
	typeSet := make(map[string]bool)
	for _, entity := range entities {
		typeSet[string(entity.Type)] = true
	}
	
	types := make([]string, 0, len(typeSet))
	for t := range typeSet {
		types = append(types, t)
	}
	
	return strings.Join(types, ",")
}

// getIndustryNameFromID maps industry ID to name
// This is a placeholder - in production, this should query the database
func (msc *MultiStrategyClassifier) getIndustryNameFromID(industryID int) string {
	// Default industry name mapping
	industryNames := map[int]string{
		1:  "Technology",
		2:  "Healthcare",
		3:  "Financial Services",
		4:  "Retail & Commerce",
		5:  "Food & Beverage",
		6:  "Manufacturing",
		7:  "Construction",
		8:  "Real Estate",
		9:  "Transportation",
		10: "Education",
		11: "Professional Services",
		12: "Agriculture",
		13: "Mining & Energy",
		14: "Utilities",
		15: "Wholesale Trade",
		16: "Arts & Entertainment",
		17: "Accommodation & Hospitality",
		18: "Administrative Services",
		19: "Other Services",
	}
	
	return industryNames[industryID]
}

// minIntValue returns the minimum of two int values
func minIntValue(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// minFloat returns the minimum of two float64 values
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

