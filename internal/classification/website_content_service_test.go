package classification

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// TestIsContentSufficient_Validation tests the content sufficiency check
func TestIsContentSufficient_Validation(t *testing.T) {
	// Create a simple logger that implements the interface
	logger := &websiteContentServiceTestLogger{}

	mockCache := &MockWebsiteContentCacher{}

	service := NewWebsiteContentService(
		nil, // scraper
		nil, // smartCrawler
		mockCache,
		logger,
	)

	tests := []struct {
		name           string
		textContent    string
		keywords       []string
		expectedResult bool
	}{
		{
			name:           "sufficient content and keywords",
			textContent:    strings.Repeat("content ", 100), // > 500 chars
			keywords:       []string{"keyword1", "keyword2", "keyword3", "keyword4", "keyword5", "keyword6", "keyword7", "keyword8", "keyword9", "keyword10"},
			expectedResult: true,
		},
		{
			name:           "insufficient content length",
			textContent:    "short",
			keywords:       []string{"keyword1", "keyword2", "keyword3", "keyword4", "keyword5", "keyword6", "keyword7", "keyword8", "keyword9", "keyword10"},
			expectedResult: false,
		},
		{
			name:           "insufficient keywords",
			textContent:    strings.Repeat("content ", 100),
			keywords:       []string{"keyword1", "keyword2"},
			expectedResult: false,
		},
		{
			name:           "both insufficient",
			textContent:    "short",
			keywords:       []string{"keyword1"},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isContentSufficient(tt.textContent, tt.keywords)
			if result != tt.expectedResult {
				t.Errorf("isContentSufficient() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

// websiteContentServiceTestLogger implements the logger interface for testing
type websiteContentServiceTestLogger struct{}

func (l *websiteContentServiceTestLogger) Printf(format string, v ...interface{}) {
	// No-op for testing
	_ = fmt.Sprintf(format, v...)
}

// MockWebsiteContentCacher for testing
type MockWebsiteContentCacher struct{}

func (m *MockWebsiteContentCacher) Get(ctx context.Context, url string) (*CachedWebsiteContent, bool) {
	return nil, false
}

func (m *MockWebsiteContentCacher) Set(ctx context.Context, url string, content *CachedWebsiteContent) error {
	return nil
}

func (m *MockWebsiteContentCacher) IsEnabled() bool {
	return true
}
