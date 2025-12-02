package nlp

import (
	"context"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// TopicRepository defines interface for querying industry-topic relationships from database
// This interface matches methods from repository.KeywordRepository
type TopicRepository interface {
	GetIndustryTopicsByKeywords(ctx context.Context, keywords []string) (map[int]float64, error) // industry_id -> relevance_score
	GetTopicAccuracy(ctx context.Context, industryID int, topic string) (float64, error)
}

// TopicScore represents a topic-industry alignment score
type TopicScore struct {
	IndustryID int
	Score      float64
	Keywords   []string // Keywords that contributed to the score
}

// TopicIndustryMapping represents a cached topic-industry relationship
type TopicIndustryMapping struct {
	IndustryID     int
	RelevanceScore float64
	AccuracyScore  float64
	LastUpdated    time.Time
}

// TopicModeler performs topic modeling using TF-IDF for industry classification
// Enhanced with database-driven industry mapping and score calibration
type TopicModeler struct {
	industryTopics map[int][]string // industry_id -> topic keywords (in-memory fallback)
	idfScores      map[string]float64 // word -> IDF score
	mu             sync.RWMutex
	minScore       float64 // Minimum score threshold (default 0.3)
	
	// Database-driven topic mapping (optional)
	topicRepo      TopicRepository // Optional repository for database queries
	topicCache     map[string]map[int]TopicIndustryMapping // topic -> industry_id -> mapping
	cacheMu        sync.RWMutex
	cacheTTL       time.Duration // Cache TTL (default 1 hour)
	useDatabase    bool          // Whether to use database for topic mapping
}

// NewTopicModeler creates a new topic modeler with default industry topics
func NewTopicModeler() *TopicModeler {
	tm := &TopicModeler{
		industryTopics: make(map[int][]string),
		idfScores:      make(map[string]float64),
		minScore:       0.15, // Lower threshold for better recall
		topicCache:     make(map[string]map[int]TopicIndustryMapping),
		cacheTTL:       1 * time.Hour,
		useDatabase:    false,
	}
	tm.loadDefaultIndustryTopics()
	// Calculate IDF scores without lock (during initialization)
	tm.calculateIDFScores()
	return tm
}

// NewTopicModelerWithRepository creates a new topic modeler with database repository support
func NewTopicModelerWithRepository(repo TopicRepository) *TopicModeler {
	tm := NewTopicModeler()
	tm.topicRepo = repo
	tm.useDatabase = repo != nil
	return tm
}

// SetRepository sets the topic repository for database queries
func (tm *TopicModeler) SetRepository(repo TopicRepository) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.topicRepo = repo
	tm.useDatabase = repo != nil
}

// IdentifyTopics identifies topics from keywords and returns industry scores
func (tm *TopicModeler) IdentifyTopics(keywords []string) map[int]float64 {
	if len(keywords) == 0 {
		return make(map[int]float64)
	}

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	topicScores := make(map[int]float64)
	keywordSet := make(map[string]bool)
	for _, kw := range keywords {
		keywordSet[strings.ToLower(kw)] = true
	}

	// Calculate topic alignment score for each industry
	for industryID, topicKeywords := range tm.industryTopics {
		score := tm.calculateTopicAlignment(keywordSet, topicKeywords)
		if score >= tm.minScore {
			topicScores[industryID] = score
		}
	}

	return topicScores
}

// IdentifyTopicsWithDetails returns detailed topic scores with contributing keywords
// Enhanced with database-driven industry mapping and score calibration
func (tm *TopicModeler) IdentifyTopicsWithDetails(keywords []string) []TopicScore {
	return tm.IdentifyTopicsWithDetailsContext(context.Background(), keywords)
}

// IdentifyTopicsWithDetailsContext returns detailed topic scores with context support
// Enhanced with database-driven industry mapping and score calibration
func (tm *TopicModeler) IdentifyTopicsWithDetailsContext(ctx context.Context, keywords []string) []TopicScore {
	if len(keywords) == 0 {
		return []TopicScore{}
	}

	tm.mu.RLock()
	keywordSet := make(map[string]bool)
	for _, kw := range keywords {
		keywordSet[strings.ToLower(kw)] = true
	}
	tm.mu.RUnlock()

	// Step 1: Calculate TF-IDF scores using in-memory topics
	var topicScores []TopicScore
	tm.mu.RLock()
	for industryID, topicKeywords := range tm.industryTopics {
		score, contributingKeywords := tm.calculateTopicAlignmentWithDetails(keywordSet, topicKeywords)
		if score >= tm.minScore {
			topicScores = append(topicScores, TopicScore{
				IndustryID: industryID,
				Score:      score,
				Keywords:   contributingKeywords,
			})
		}
	}
	tm.mu.RUnlock()

	// Step 2: Map topics to industries using database (if available)
	if tm.useDatabase && tm.topicRepo != nil {
		industryTopics := tm.mapTopicsToIndustries(ctx, keywords)
		// Merge database results with in-memory results
		topicScores = tm.mergeTopicScores(topicScores, industryTopics)
	}

	// Step 3: Calibrate scores based on historical accuracy
	calibratedScores := tm.calibrateScores(ctx, topicScores)

	// Sort by score (highest first)
	sort.Slice(calibratedScores, func(i, j int) bool {
		return calibratedScores[i].Score > calibratedScores[j].Score
	})

	return calibratedScores
}

// mapTopicsToIndustries maps topics to industries using database
// Returns enhanced topic scores with database-driven relevance
func (tm *TopicModeler) mapTopicsToIndustries(ctx context.Context, keywords []string) []TopicScore {
	if tm.topicRepo == nil {
		return []TopicScore{}
	}

	// Check cache first
	cacheKey := strings.Join(keywords, ",")
	tm.cacheMu.RLock()
	if cached, exists := tm.topicCache[cacheKey]; exists {
		// Check if cache is still valid
		now := time.Now()
		valid := true
		for _, mapping := range cached {
			if now.Sub(mapping.LastUpdated) > tm.cacheTTL {
				valid = false
				break
			}
		}
		if valid {
			tm.cacheMu.RUnlock()
			// Convert cached mappings to TopicScores
			var scores []TopicScore
			for industryID, mapping := range cached {
				scores = append(scores, TopicScore{
					IndustryID: industryID,
					Score:      mapping.RelevanceScore,
					Keywords:   keywords,
				})
			}
			return scores
		}
	}
	tm.cacheMu.RUnlock()

	// Query database for industry-topic relationships
	industryRelevance, err := tm.topicRepo.GetIndustryTopicsByKeywords(ctx, keywords)
	if err != nil {
		// Fallback to in-memory if database query fails
		return []TopicScore{}
	}

	// Convert to TopicScores and cache
	var scores []TopicScore
	cachedMappings := make(map[int]TopicIndustryMapping)
	for industryID, relevanceScore := range industryRelevance {
		// Get accuracy score if available
		accuracyScore := 0.75 // Default
		if acc, err := tm.topicRepo.GetTopicAccuracy(ctx, industryID, keywords[0]); err == nil {
			accuracyScore = acc
		}

		scores = append(scores, TopicScore{
			IndustryID: industryID,
			Score:      relevanceScore,
			Keywords:   keywords,
		})

		cachedMappings[industryID] = TopicIndustryMapping{
			IndustryID:     industryID,
			RelevanceScore: relevanceScore,
			AccuracyScore:  accuracyScore,
			LastUpdated:    time.Now(),
		}
	}

	// Update cache
	tm.cacheMu.Lock()
	tm.topicCache[cacheKey] = cachedMappings
	tm.cacheMu.Unlock()

	return scores
}

// mergeTopicScores merges in-memory and database topic scores
func (tm *TopicModeler) mergeTopicScores(inMemory []TopicScore, database []TopicScore) []TopicScore {
	// Create map of industry scores
	scoreMap := make(map[int]TopicScore)

	// Add in-memory scores (weight: 0.4)
	for _, score := range inMemory {
		score.Score *= 0.4
		scoreMap[score.IndustryID] = score
	}

	// Add database scores (weight: 0.6, higher weight for database accuracy)
	for _, score := range database {
		score.Score *= 0.6
		if existing, exists := scoreMap[score.IndustryID]; exists {
			// Merge: combine scores
			existing.Score += score.Score
			existing.Keywords = append(existing.Keywords, score.Keywords...)
			scoreMap[score.IndustryID] = existing
		} else {
			scoreMap[score.IndustryID] = score
		}
	}

	// Convert map back to slice
	result := make([]TopicScore, 0, len(scoreMap))
	for _, score := range scoreMap {
		result = append(result, score)
	}

	return result
}

// calibrateScores calibrates topic scores based on historical accuracy
func (tm *TopicModeler) calibrateScores(ctx context.Context, scores []TopicScore) []TopicScore {
	if !tm.useDatabase || tm.topicRepo == nil {
		// No calibration without database
		return scores
	}

	calibrated := make([]TopicScore, 0, len(scores))
	for _, score := range scores {
		// Get accuracy score for this industry-topic pair
		accuracyScore := 0.75 // Default
		if len(score.Keywords) > 0 {
			if acc, err := tm.topicRepo.GetTopicAccuracy(ctx, score.IndustryID, score.Keywords[0]); err == nil {
				accuracyScore = acc
			}
		}

		// Calibrate: adjust score based on historical accuracy
		// Higher accuracy = boost score, lower accuracy = reduce score
		calibrationFactor := 0.5 + (accuracyScore * 0.5) // Range: 0.5 to 1.0
		calibratedScore := score.Score * calibrationFactor

		calibrated = append(calibrated, TopicScore{
			IndustryID: score.IndustryID,
			Score:      math.Min(calibratedScore, 1.0), // Cap at 1.0
			Keywords:   score.Keywords,
		})
	}

	return calibrated
}

// calculateTopicAlignment calculates how well keywords align with industry topic keywords
// Uses TF-IDF weighted scoring
func (tm *TopicModeler) calculateTopicAlignment(keywordSet map[string]bool, topicKeywords []string) float64 {
	if len(topicKeywords) == 0 {
		return 0.0
	}

	matches := 0
	totalTFIDF := 0.0
	maxPossibleTFIDF := 0.0

	for _, topicKw := range topicKeywords {
		topicKwLower := strings.ToLower(topicKw)
		
		// Check for exact match
		if keywordSet[topicKwLower] {
			matches++
			idf := tm.getIDF(topicKwLower)
			tfidf := 1.0 * idf // TF = 1.0 for matched keyword
			totalTFIDF += tfidf
		}
		
		// Calculate max possible TF-IDF for normalization
		idf := tm.getIDF(topicKwLower)
		maxPossibleTFIDF += 1.0 * idf
	}

	if maxPossibleTFIDF == 0 {
		return 0.0
	}

	// Normalized TF-IDF score
	score := totalTFIDF / maxPossibleTFIDF

	// Boost score based on match ratio (more matches = higher score)
	matchRatio := float64(matches) / float64(len(topicKeywords))
	// Use a more generous scoring: base score + match ratio boost
	score = score*0.6 + matchRatio*0.4

	return math.Min(score, 1.0) // Cap at 1.0
}

// calculateTopicAlignmentWithDetails returns score and contributing keywords
func (tm *TopicModeler) calculateTopicAlignmentWithDetails(keywordSet map[string]bool, topicKeywords []string) (float64, []string) {
	if len(topicKeywords) == 0 {
		return 0.0, []string{}
	}

	var contributingKeywords []string
	totalTFIDF := 0.0
	maxPossibleTFIDF := 0.0

	for _, topicKw := range topicKeywords {
		topicKwLower := strings.ToLower(topicKw)
		
		// Check for exact match
		if keywordSet[topicKwLower] {
			contributingKeywords = append(contributingKeywords, topicKwLower)
			idf := tm.getIDF(topicKwLower)
			tfidf := 1.0 * idf // TF = 1.0 for matched keyword
			totalTFIDF += tfidf
		}
		
		// Calculate max possible TF-IDF for normalization
		idf := tm.getIDF(topicKwLower)
		maxPossibleTFIDF += 1.0 * idf
	}

	if maxPossibleTFIDF == 0 {
		return 0.0, contributingKeywords
	}

	// Normalized TF-IDF score
	if maxPossibleTFIDF == 0 {
		return 0.0, contributingKeywords
	}
	score := totalTFIDF / maxPossibleTFIDF

	// Boost score based on match ratio (more matches = higher score)
	matches := len(contributingKeywords)
	matchRatio := float64(matches) / float64(len(topicKeywords))
	// Use a more generous scoring: base score + match ratio boost
	score = score*0.6 + matchRatio*0.4

	return math.Min(score, 1.0), contributingKeywords
}

// getIDF returns the IDF (Inverse Document Frequency) score for a word
func (tm *TopicModeler) getIDF(word string) float64 {
	if idf, exists := tm.idfScores[word]; exists {
		return idf
	}
	// Default IDF for unknown words (log of total industries)
	return math.Log(float64(len(tm.industryTopics)) + 1)
}

// calculateIDFScores calculates IDF scores for all topic keywords
// Note: This should be called without locks as it's called during initialization
func (tm *TopicModeler) calculateIDFScores() {
	// Count document frequency (how many industries contain each word)
	wordDocFreq := make(map[string]int)
	totalDocs := len(tm.industryTopics)

	if totalDocs == 0 {
		return
	}

	for _, topicKeywords := range tm.industryTopics {
		seen := make(map[string]bool)
		for _, kw := range topicKeywords {
			kwLower := strings.ToLower(kw)
			if !seen[kwLower] {
				seen[kwLower] = true
				wordDocFreq[kwLower]++
			}
		}
	}

	// Calculate IDF: log((total_docs + 1) / (doc_freq + 1)) to avoid division by zero
	tm.idfScores = make(map[string]float64)
	for word, docFreq := range wordDocFreq {
		if docFreq > 0 {
			// Use smoothed IDF to avoid extreme values
			idf := math.Log((float64(totalDocs) + 1.0) / (float64(docFreq) + 1.0))
			tm.idfScores[word] = idf
		}
	}
}

// SetMinScore sets the minimum score threshold
func (tm *TopicModeler) SetMinScore(minScore float64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.minScore = minScore
}

// AddIndustryTopics adds or updates topic keywords for an industry
func (tm *TopicModeler) AddIndustryTopics(industryID int, keywords []string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.industryTopics[industryID] = keywords
	// Recalculate IDF scores
	tm.calculateIDFScores()
}

// GetIndustryTopics returns topic keywords for an industry
func (tm *TopicModeler) GetIndustryTopics(industryID int) []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if keywords, exists := tm.industryTopics[industryID]; exists {
		return keywords
	}
	return []string{} // Return empty slice instead of nil
}

// loadDefaultIndustryTopics loads default industry topic keywords
// Industry IDs based on common classification:
// 1: Technology, 2: Healthcare, 3: Financial Services, 4: Retail, 5: Food & Beverage, etc.
func (tm *TopicModeler) loadDefaultIndustryTopics() {
	tm.industryTopics = map[int][]string{
		// Technology (industry_id = 1)
		1: {
			"software", "technology", "tech", "digital", "computer", "app", "application",
			"platform", "system", "network", "internet", "web", "cloud", "data", "ai",
			"artificial intelligence", "machine learning", "development", "programming",
			"IT", "information technology", "cyber", "software development",
		},
		// Healthcare (industry_id = 2)
		2: {
			"health", "medical", "healthcare", "doctor", "clinic", "hospital", "patient",
			"medicine", "therapy", "wellness", "dental", "pharmacy", "nursing", "treatment",
			"diagnostic", "therapeutic", "medical device", "health services",
		},
		// Financial Services (industry_id = 3)
		3: {
			"financial", "finance", "banking", "bank", "credit", "loan", "mortgage",
			"investment", "insurance", "trading", "accounting", "audit", "tax",
			"fintech", "wealth management", "financial services", "money",
		},
		// Retail & Commerce (industry_id = 4)
		4: {
			"retail", "store", "shop", "selling", "merchandise", "products", "commerce",
			"ecommerce", "e-commerce", "online store", "shopping", "marketplace", "vendor",
			"retailer", "merchant", "sales", "point of sale", "POS",
		},
		// Food & Beverage (industry_id = 5)
		5: {
			"food", "beverage", "restaurant", "cafe", "dining", "kitchen", "catering",
			"bakery", "bar", "pub", "bistro", "eatery", "diner", "tavern", "gastropub",
			"wine", "beer", "coffee", "tea", "alcohol", "spirits", "liquor",
		},
		// Manufacturing (industry_id = 6)
		6: {
			"manufacturing", "production", "factory", "industrial", "machinery", "assembly",
			"production line", "manufacturing plant", "industrial production", "fabrication",
		},
		// Construction (industry_id = 7)
		7: {
			"construction", "building", "contractor", "builder", "construction firm",
			"construction company", "construction services", "building construction",
		},
		// Real Estate (industry_id = 8)
		8: {
			"real estate", "property", "realty", "realtor", "property management",
			"real estate services", "property development", "real estate investment",
		},
		// Transportation (industry_id = 9)
		9: {
			"transportation", "transport", "logistics", "shipping", "delivery", "freight",
			"courier", "transportation services", "logistics services", "shipping services",
		},
		// Education (industry_id = 10)
		10: {
			"education", "school", "university", "college", "academy", "institute",
			"training", "learning", "educational", "teaching", "student", "academic",
		},
		// Professional Services (industry_id = 11)
		11: {
			"consulting", "advisory", "professional services", "consulting services",
			"advisory services", "professional consulting", "business consulting",
		},
		// Agriculture (industry_id = 12)
		12: {
			"agriculture", "farming", "agricultural", "farm", "crop", "livestock",
			"agricultural production", "farming services", "agricultural services",
		},
		// Mining & Energy (industry_id = 13)
		13: {
			"mining", "energy", "oil", "gas", "petroleum", "mining services",
			"energy services", "oil and gas", "mining operations",
		},
		// Utilities (industry_id = 14)
		14: {
			"utilities", "power", "electricity", "water", "gas utility", "utility services",
			"public utility", "utility company",
		},
		// Wholesale Trade (industry_id = 15)
		15: {
			"wholesale", "wholesaler", "wholesale trade", "wholesale distribution",
			"wholesale services", "bulk distribution",
		},
		// Arts & Entertainment (industry_id = 16)
		16: {
			"entertainment", "arts", "music", "film", "video", "gaming", "sports",
			"recreation", "entertainment services", "arts and entertainment",
		},
		// Accommodation & Hospitality (industry_id = 17)
		17: {
			"hotel", "hospitality", "accommodation", "resort", "inn", "lodge", "motel",
			"hostel", "hospitality services", "accommodation services",
		},
		// Administrative Services (industry_id = 18)
		18: {
			"administrative", "administrative services", "office services", "business services",
			"administrative support", "office support",
		},
		// Other Services (industry_id = 19)
		19: {
			"services", "business services", "service provider", "service company",
			"service business", "service industry",
		},
	}
}

