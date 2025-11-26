package nlp

import (
	"math"
	"sort"
	"strings"
	"sync"
)

// TopicScore represents a topic-industry alignment score
type TopicScore struct {
	IndustryID int
	Score      float64
	Keywords   []string // Keywords that contributed to the score
}

// TopicModeler performs topic modeling using TF-IDF for industry classification
type TopicModeler struct {
	industryTopics map[int][]string // industry_id -> topic keywords
	idfScores      map[string]float64 // word -> IDF score
	mu             sync.RWMutex
	minScore       float64 // Minimum score threshold (default 0.3)
}

// NewTopicModeler creates a new topic modeler with default industry topics
func NewTopicModeler() *TopicModeler {
	tm := &TopicModeler{
		industryTopics: make(map[int][]string),
		idfScores:      make(map[string]float64),
		minScore:       0.15, // Lower threshold for better recall
	}
	tm.loadDefaultIndustryTopics()
	// Calculate IDF scores without lock (during initialization)
	tm.calculateIDFScores()
	return tm
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
func (tm *TopicModeler) IdentifyTopicsWithDetails(keywords []string) []TopicScore {
	if len(keywords) == 0 {
		return []TopicScore{}
	}

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var topicScores []TopicScore
	keywordSet := make(map[string]bool)
	for _, kw := range keywords {
		keywordSet[strings.ToLower(kw)] = true
	}

	// Calculate topic alignment score for each industry
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

	// Sort by score (highest first)
	sort.Slice(topicScores, func(i, j int) bool {
		return topicScores[i].Score > topicScores[j].Score
	})

	return topicScores
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

