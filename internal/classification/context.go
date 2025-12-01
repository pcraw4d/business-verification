package classification

import (
	"context"
	"sync"
	"time"

	"kyb-platform/internal/classification/repository"
)

// ClassificationContext stores shared data for a single classification request
// This prevents redundant keyword extraction and processing across multiple steps
type ClassificationContext struct {
	// Extracted keywords (extracted once, reused throughout pipeline)
	Keywords []string
	
	// Contextual keywords with metadata (using repository.ContextualKeyword)
	ContextualKeywords []repository.ContextualKeyword
	
	// Extracted entities from NER
	Entities []Entity
	
	// Topic scores from topic modeling
	TopicScores map[int]float64
	
	// Website content (cached to avoid re-scraping)
	WebsiteContent string
	
	// Structured data extracted from website
	StructuredData map[string]interface{}
	
	// Metadata
	ExtractionTime time.Time
	BusinessName   string
	WebsiteURL     string
	
	// Mutex for thread-safe access
	mu sync.RWMutex
}

// Entity represents a named entity extracted from text
type Entity struct {
	Text       string
	Type       string
	Confidence float64
	Source     string
}

// NewClassificationContext creates a new classification context
func NewClassificationContext(businessName, websiteURL string) *ClassificationContext {
	return &ClassificationContext{
		Keywords:          make([]string, 0),
		ContextualKeywords: make([]repository.ContextualKeyword, 0),
		Entities:          make([]Entity, 0),
		TopicScores:       make(map[int]float64),
		StructuredData:    make(map[string]interface{}),
		ExtractionTime:    time.Now(),
		BusinessName:      businessName,
		WebsiteURL:        websiteURL,
	}
}

// SetKeywords sets the extracted keywords (thread-safe)
func (cc *ClassificationContext) SetKeywords(keywords []string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.Keywords = keywords
}

// GetKeywords gets the extracted keywords (thread-safe)
func (cc *ClassificationContext) GetKeywords() []string {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.Keywords
}

// SetContextualKeywords sets the contextual keywords (thread-safe)
func (cc *ClassificationContext) SetContextualKeywords(keywords []repository.ContextualKeyword) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.ContextualKeywords = keywords
}

// GetContextualKeywords gets the contextual keywords (thread-safe)
func (cc *ClassificationContext) GetContextualKeywords() []repository.ContextualKeyword {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.ContextualKeywords
}

// SetEntities sets the extracted entities (thread-safe)
func (cc *ClassificationContext) SetEntities(entities []Entity) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.Entities = entities
}

// GetEntities gets the extracted entities (thread-safe)
func (cc *ClassificationContext) GetEntities() []Entity {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.Entities
}

// SetTopicScores sets the topic scores (thread-safe)
func (cc *ClassificationContext) SetTopicScores(scores map[int]float64) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.TopicScores = scores
}

// GetTopicScores gets the topic scores (thread-safe)
func (cc *ClassificationContext) GetTopicScores() map[int]float64 {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	// Return a copy to prevent external modification
	result := make(map[int]float64)
	for k, v := range cc.TopicScores {
		result[k] = v
	}
	return result
}

// SetWebsiteContent sets the cached website content (thread-safe)
func (cc *ClassificationContext) SetWebsiteContent(content string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.WebsiteContent = content
}

// GetWebsiteContent gets the cached website content (thread-safe)
func (cc *ClassificationContext) GetWebsiteContent() string {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.WebsiteContent
}

// SetStructuredData sets the structured data (thread-safe)
func (cc *ClassificationContext) SetStructuredData(data map[string]interface{}) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.StructuredData = data
}

// GetStructuredData gets the structured data (thread-safe)
func (cc *ClassificationContext) GetStructuredData() map[string]interface{} {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	// Return a copy to prevent external modification
	result := make(map[string]interface{})
	for k, v := range cc.StructuredData {
		result[k] = v
	}
	return result
}

// HasKeywords checks if keywords have been extracted (thread-safe)
func (cc *ClassificationContext) HasKeywords() bool {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return len(cc.Keywords) > 0 || len(cc.ContextualKeywords) > 0
}

// contextKey is a type for context keys to avoid collisions
type contextKey string

const (
	// ClassificationContextKey is the key for storing ClassificationContext in context.Context
	ClassificationContextKey contextKey = "classification_context"
)

// WithClassificationContext adds a ClassificationContext to the context
func WithClassificationContext(ctx context.Context, cc *ClassificationContext) context.Context {
	return context.WithValue(ctx, ClassificationContextKey, cc)
}

// GetClassificationContext retrieves the ClassificationContext from context
func GetClassificationContext(ctx context.Context) (*ClassificationContext, bool) {
	cc, ok := ctx.Value(ClassificationContextKey).(*ClassificationContext)
	return cc, ok
}

