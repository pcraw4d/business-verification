package classification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"kyb-platform/internal/external"
)

func TestLLMClassifier_ClassifyWithLLM(t *testing.T) {
	// Create a mock LLM service
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/classify" {
			t.Errorf("Expected path /classify, got %s", r.URL.Path)
		}

		// Parse request
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Return mock response
		response := map[string]interface{}{
			"primary_industry": "Management Consulting Services",
			"confidence":       0.88,
			"reasoning":        "This business provides strategic advisory services, which aligns with management consulting.",
			"codes": map[string]interface{}{
				"mcc": []map[string]interface{}{
					{"code": "8742", "description": "Management Consulting Services", "confidence": 0.90},
					{"code": "8741", "description": "Commercial Physical and Biological Research", "confidence": 0.75},
					{"code": "8748", "description": "Business Consulting Services", "confidence": 0.70},
				},
				"sic": []map[string]interface{}{
					{"code": "8742", "description": "Management Consulting Services", "confidence": 0.90},
				},
				"naics": []map[string]interface{}{
					{"code": "541611", "description": "Administrative Management and General Management Consulting Services", "confidence": 0.90},
				},
			},
			"alternative_classifications": []string{"Business Services", "Professional Services"},
			"processing_time_ms":           2340,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Create LLM classifier
	classifier := NewLLMClassifier(mockServer.URL, nil)

	// Create test content
	content := &external.ScrapedContent{
		Title:     "Smith & Associates",
		MetaDesc:  "Professional services firm providing strategic advisory",
		AboutText: "We provide strategic consulting services to businesses.",
		Headings:  []string{"Services", "About Us", "Contact"},
		Domain:    "smith-associates.com",
	}

	// Create layer 1 result
	layer1Result := &MultiStrategyResult{
		PrimaryIndustry: "Business Services",
		Confidence:      0.68,
		Keywords:        []string{"professional", "services", "advisory"},
		Method:          "multi_strategy",
	}

	// Create layer 2 result
	layer2Result := &EmbeddingClassificationResult{
		TopMatch:      "Management Consulting Services",
		TopSimilarity: 0.78,
		Confidence:    0.74,
		Method:        "embedding_similarity",
	}

	// Test classification
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := classifier.ClassifyWithLLM(
		ctx,
		content,
		"Smith & Associates",
		"Professional services firm providing strategic advisory",
		layer1Result,
		layer2Result,
	)

	if err != nil {
		t.Fatalf("ClassifyWithLLM failed: %v", err)
	}

	// Verify result
	if result.PrimaryIndustry != "Management Consulting Services" {
		t.Errorf("Expected primary industry 'Management Consulting Services', got '%s'", result.PrimaryIndustry)
	}

	if result.Confidence < 0.85 || result.Confidence > 0.95 {
		t.Errorf("Expected confidence around 0.88, got %.2f", result.Confidence)
	}

	if len(result.MCC) == 0 {
		t.Error("Expected MCC codes, got none")
	} else if result.MCC[0].Code != "8742" {
		t.Errorf("Expected top MCC code '8742', got '%s'", result.MCC[0].Code)
	}

	if len(result.AlternativeClassifications) == 0 {
		t.Error("Expected alternative classifications, got none")
	}

	if result.ProcessingTimeMs <= 0 {
		t.Errorf("Expected positive processing time, got %d", result.ProcessingTimeMs)
	}
}

func TestLLMClassifier_PrepareWebsiteContent(t *testing.T) {
	classifier := NewLLMClassifier("http://test", nil)

	content := &external.ScrapedContent{
		Title:     "Test Business",
		MetaDesc:  "Test description",
		AboutText: "This is a very long about text that should be truncated if it exceeds 800 characters. " + string(make([]byte, 1000)),
		Headings:  []string{"Heading 1", "Heading 2", "Heading 3", "Heading 4", "Heading 5", "Heading 6", "Heading 7"},
		Domain:    "test.com",
	}

	prepared := classifier.prepareWebsiteContent(content)

	// Should be truncated to 2000 chars
	if len(prepared) > 2000 {
		t.Errorf("Expected content truncated to 2000 chars, got %d", len(prepared))
	}

	// Should contain title
	if !strings.Contains(prepared, "Test Business") {
		t.Error("Expected prepared content to contain title")
	}
}

func TestLLMClassifier_ParseCodes(t *testing.T) {
	classifier := NewLLMClassifier("http://test", nil)

	// Test valid codes
	codesInterface := []interface{}{
		map[string]interface{}{
			"code":        "1234",
			"description": "Test Code",
			"confidence":  0.90,
		},
		map[string]interface{}{
			"code":        "5678",
			"description": "Another Code",
			"confidence":  0.85,
		},
	}

	codes := classifier.parseCodes(codesInterface)

	if len(codes) != 2 {
		t.Errorf("Expected 2 codes, got %d", len(codes))
	}

	if codes[0].Code != "1234" {
		t.Errorf("Expected first code '1234', got '%s'", codes[0].Code)
	}

	if codes[0].Source != "llm_reasoning" {
		t.Errorf("Expected source 'llm_reasoning', got '%s'", codes[0].Source)
	}

	// Test nil input
	nilCodes := classifier.parseCodes(nil)
	if len(nilCodes) != 0 {
		t.Errorf("Expected empty codes for nil input, got %d", len(nilCodes))
	}
}

func TestLLMClassifier_ParseAlternatives(t *testing.T) {
	classifier := NewLLMClassifier("http://test", nil)

	// Test valid alternatives
	altInterface := []interface{}{
		"Alternative 1",
		"Alternative 2",
		"Alternative 3",
	}

	alternatives := classifier.parseAlternatives(altInterface)

	if len(alternatives) != 3 {
		t.Errorf("Expected 3 alternatives, got %d", len(alternatives))
	}

	if alternatives[0] != "Alternative 1" {
		t.Errorf("Expected first alternative 'Alternative 1', got '%s'", alternatives[0])
	}

	// Test nil input
	nilAlts := classifier.parseAlternatives(nil)
	if len(nilAlts) != 0 {
		t.Errorf("Expected empty alternatives for nil input, got %d", len(nilAlts))
	}
}

func TestLLMClassifier_ErrorHandling(t *testing.T) {
	// Create a server that returns an error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	classifier := NewLLMClassifier(mockServer.URL, nil)

	content := &external.ScrapedContent{
		Title:  "Test",
		Domain: "test.com",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := classifier.ClassifyWithLLM(
		ctx,
		content,
		"Test Business",
		"Test description",
		nil,
		nil,
	)

	if err == nil {
		t.Error("Expected error for server error response, got nil")
	}
}


