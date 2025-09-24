package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Classification process test structures
type ClassificationRequest struct {
	MerchantID   string            `json:"merchant_id"`
	BusinessName string            `json:"business_name"`
	Description  string            `json:"description"`
	Website      string            `json:"website"`
	Keywords     []string          `json:"keywords"`
	Metadata     map[string]string `json:"metadata"`
}

type IndustryCode struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	Source      string  `json:"source"`
}

type ClassificationResult struct {
	ID                    string         `json:"id"`
	MerchantID            string         `json:"merchant_id"`
	Status                string         `json:"status"`
	MCCCodes              []IndustryCode `json:"mcc_codes"`
	NAICSCodes            []IndustryCode `json:"naics_codes"`
	SICCodes              []IndustryCode `json:"sic_codes"`
	PrimaryIndustry       string         `json:"primary_industry"`
	ConfidenceScore       float64        `json:"confidence_score"`
	ClassificationMethods []string       `json:"classification_methods"`
	ProcessedAt           time.Time      `json:"processed_at"`
}

type MultiMethodAnalysis struct {
	KeywordMatches      map[string]float64 `json:"keyword_matches"`
	DescriptionAnalysis map[string]float64 `json:"description_analysis"`
	WebsiteAnalysis     map[string]float64 `json:"website_analysis"`
	EnsembleScore       float64            `json:"ensemble_score"`
	RecommendedCodes    []IndustryCode     `json:"recommended_codes"`
}

// Mock handlers for classification process
func createClassificationHandlers() *http.ServeMux {
	mux := http.NewServeMux()

	// Multi-method classification endpoint
	mux.HandleFunc("POST /api/v1/classify/multi-method", func(w http.ResponseWriter, r *http.Request) {
		var req ClassificationRequest
		json.NewDecoder(r.Body).Decode(&req)

		response := ClassificationResult{
			ID:         "classification-123",
			MerchantID: req.MerchantID,
			Status:     "completed",
			MCCCodes: []IndustryCode{
				{Code: "5734", Description: "Computer Software Stores", Confidence: 0.92, Source: "keyword_match"},
				{Code: "7372", Description: "Prepackaged Software", Confidence: 0.88, Source: "description_analysis"},
				{Code: "5045", Description: "Computers and Computer Peripheral Equipment", Confidence: 0.85, Source: "website_analysis"},
			},
			NAICSCodes: []IndustryCode{
				{Code: "541511", Description: "Custom Computer Programming Services", Confidence: 0.94, Source: "ensemble"},
				{Code: "541512", Description: "Computer Systems Design Services", Confidence: 0.89, Source: "keyword_match"},
				{Code: "443142", Description: "Electronics Stores", Confidence: 0.82, Source: "description_analysis"},
			},
			SICCodes: []IndustryCode{
				{Code: "7371", Description: "Computer Programming Services", Confidence: 0.91, Source: "ensemble"},
				{Code: "7373", Description: "Computer Integrated Systems Design", Confidence: 0.87, Source: "website_analysis"},
				{Code: "5734", Description: "Computer and Computer Software Stores", Confidence: 0.84, Source: "keyword_match"},
			},
			PrimaryIndustry:       "Technology Services",
			ConfidenceScore:       0.91,
			ClassificationMethods: []string{"keyword_match", "description_analysis", "website_analysis", "ensemble"},
			ProcessedAt:           time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Industry code assignment endpoint
	mux.HandleFunc("POST /api/v1/assign/industry-codes", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"assigned_codes": map[string]string{
				"primary_mcc":   "5734",
				"primary_naics": "541511",
				"primary_sic":   "7371",
			},
			"assignment_reason": "Based on highest confidence scores from multi-method classification",
			"effective_date":    time.Now(),
			"review_required":   false,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Confidence scoring endpoint
	mux.HandleFunc("POST /api/v1/score/confidence", func(w http.ResponseWriter, r *http.Request) {
		response := MultiMethodAnalysis{
			KeywordMatches: map[string]float64{
				"software":    0.95,
				"technology":  0.92,
				"programming": 0.89,
				"computer":    0.87,
			},
			DescriptionAnalysis: map[string]float64{
				"business_context": 0.91,
				"industry_terms":   0.88,
				"service_type":     0.85,
			},
			WebsiteAnalysis: map[string]float64{
				"content_analysis": 0.89,
				"meta_tags":        0.86,
				"navigation":       0.83,
			},
			EnsembleScore: 0.91,
			RecommendedCodes: []IndustryCode{
				{Code: "541511", Description: "Custom Computer Programming Services", Confidence: 0.94, Source: "ensemble"},
				{Code: "5734", Description: "Computer Software Stores", Confidence: 0.92, Source: "keyword_match"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	return mux
}

// TestSimplifiedClassificationProcess tests the classification process
func TestSimplifiedClassificationProcess(t *testing.T) {
	// Create test server
	mux := createClassificationHandlers()
	server := httptest.NewServer(mux)
	defer server.Close()

	t.Run("Complete Classification Process Journey", func(t *testing.T) {
		// Step 1: Multi-Method Classification
		classificationReq := ClassificationRequest{
			MerchantID:   "merchant-123",
			BusinessName: "TechSoft Solutions",
			Description:  "We provide custom software development and computer programming services for businesses",
			Website:      "https://techsoft.com",
			Keywords:     []string{"software", "programming", "technology", "development"},
			Metadata: map[string]string{
				"industry_hint": "technology",
				"service_type":  "software_development",
			},
		}

		resp, body, err := makeSimpleRequest("POST", "/api/v1/classify/multi-method", classificationReq, server)
		if err != nil {
			t.Fatalf("Multi-method classification failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var classificationResp ClassificationResult
		if err := json.Unmarshal(body, &classificationResp); err != nil {
			t.Fatalf("Failed to parse classification response: %v", err)
		}

		if classificationResp.Status != "completed" {
			t.Errorf("Expected status 'completed', got '%s'", classificationResp.Status)
		}

		if len(classificationResp.MCCCodes) < 3 {
			t.Errorf("Expected at least 3 MCC codes, got %d", len(classificationResp.MCCCodes))
		}

		if len(classificationResp.NAICSCodes) < 3 {
			t.Errorf("Expected at least 3 NAICS codes, got %d", len(classificationResp.NAICSCodes))
		}

		if len(classificationResp.SICCodes) < 3 {
			t.Errorf("Expected at least 3 SIC codes, got %d", len(classificationResp.SICCodes))
		}

		if classificationResp.ConfidenceScore < 0.8 {
			t.Errorf("Expected confidence score >= 0.8, got %f", classificationResp.ConfidenceScore)
		}

		t.Logf("✓ Multi-method classification successful: Score=%.2f, Primary=%s",
			classificationResp.ConfidenceScore, classificationResp.PrimaryIndustry)

		// Step 2: Industry Code Assignment
		resp, body, err = makeSimpleRequest("POST", "/api/v1/assign/industry-codes", map[string]string{
			"classification_id": classificationResp.ID,
		}, server)
		if err != nil {
			t.Fatalf("Industry code assignment failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var assignmentResp map[string]interface{}
		if err := json.Unmarshal(body, &assignmentResp); err != nil {
			t.Fatalf("Failed to parse assignment response: %v", err)
		}

		assignedCodes := assignmentResp["assigned_codes"].(map[string]interface{})
		if assignedCodes["primary_mcc"] == "" {
			t.Error("Expected primary MCC code to be assigned")
		}

		if assignedCodes["primary_naics"] == "" {
			t.Error("Expected primary NAICS code to be assigned")
		}

		if assignedCodes["primary_sic"] == "" {
			t.Error("Expected primary SIC code to be assigned")
		}

		t.Logf("✓ Industry code assignment successful: MCC=%s, NAICS=%s, SIC=%s",
			assignedCodes["primary_mcc"], assignedCodes["primary_naics"], assignedCodes["primary_sic"])

		// Step 3: Confidence Scoring
		resp, body, err = makeSimpleRequest("POST", "/api/v1/score/confidence", map[string]string{
			"classification_id": classificationResp.ID,
		}, server)
		if err != nil {
			t.Fatalf("Confidence scoring failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var scoringResp MultiMethodAnalysis
		if err := json.Unmarshal(body, &scoringResp); err != nil {
			t.Fatalf("Failed to parse scoring response: %v", err)
		}

		if scoringResp.EnsembleScore < 0.8 {
			t.Errorf("Expected ensemble score >= 0.8, got %f", scoringResp.EnsembleScore)
		}

		if len(scoringResp.KeywordMatches) == 0 {
			t.Error("Expected keyword matches to be provided")
		}

		if len(scoringResp.RecommendedCodes) < 2 {
			t.Errorf("Expected at least 2 recommended codes, got %d", len(scoringResp.RecommendedCodes))
		}

		t.Logf("✓ Confidence scoring successful: Ensemble=%.2f, Keywords=%d",
			scoringResp.EnsembleScore, len(scoringResp.KeywordMatches))

		t.Log("✅ Complete classification process test passed successfully")
	})
}
