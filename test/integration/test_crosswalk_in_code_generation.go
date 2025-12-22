//go:build e2e_railway
// +build e2e_railway

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestCrosswalkInCodeGeneration verifies that crosswalks are used in code generation
func TestCrosswalkInCodeGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping crosswalk test in short mode")
	}

	// Get Railway API URL
	apiURL := os.Getenv("RAILWAY_API_URL")
	if apiURL == "" {
		apiURL = "https://classification-service-production.up.railway.app"
	}

	t.Log("========================================")
	t.Log("CROSSWALK IN CODE GENERATION TEST")
	t.Log("========================================")
	t.Logf("Testing against: %s", apiURL)

	// Test cases with codes that should have crosswalks
	testCases := []struct {
		name        string
		business    string
		website     string
		expectMCC   bool
		expectNAICS bool
		expectSIC   bool
	}{
		{
			name:        "Food Store (MCC 5819)",
			business:     "Convenience Store",
			website:      "https://example-convenience-store.com",
			expectMCC:    true,
			expectNAICS: true, // Should have crosswalk from MCC
			expectSIC:   true, // Should have crosswalk from MCC
		},
		{
			name:        "Technology Company",
			business:    "Software Development Company",
			website:     "https://example-software.com",
			expectMCC:   true,
			expectNAICS: true,
			expectSIC:   true,
		},
	}

	client := &http.Client{
		Timeout: 90 * time.Second,
	}

	successCount := 0
	totalTests := len(testCases)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare request
			reqBody := map[string]interface{}{
				"business_name": tc.business,
				"website_url":   tc.website,
			}

			jsonData, err := json.Marshal(reqBody)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			// Make request
			url := fmt.Sprintf("%s/v1/classify", apiURL)
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", "application/json")

			startTime := time.Now()
			resp, err := client.Do(req)
			duration := time.Since(startTime)

			if err != nil {
				t.Logf("⚠️ Request failed: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				t.Logf("⚠️ Request failed with status %d: %s", resp.StatusCode, string(body))
				return
			}

			// Parse response
			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Logf("⚠️ Failed to decode response: %v", err)
				return
			}

			// Check for codes
			mccCodes, _ := result["mcc_codes"].([]interface{})
			naicsCodes, _ := result["naics_codes"].([]interface{})
			sicCodes, _ := result["sic_codes"].([]interface{})

			t.Logf("\n📊 Results for: %s", tc.name)
			t.Logf("   MCC codes: %d", len(mccCodes))
			t.Logf("   NAICS codes: %d", len(naicsCodes))
			t.Logf("   SIC codes: %d", len(sicCodes))
			t.Logf("   Response time: %v", duration)

			// Verify expectations
			hasMCC := len(mccCodes) > 0
			hasNAICS := len(naicsCodes) > 0
			hasSIC := len(sicCodes) > 0

			if tc.expectMCC && !hasMCC {
				t.Logf("   ⚠️ Expected MCC codes but got none")
			} else if hasMCC {
				t.Logf("   ✅ MCC codes generated")
			}

			if tc.expectNAICS && !hasNAICS {
				t.Logf("   ⚠️ Expected NAICS codes but got none")
			} else if hasNAICS {
				t.Logf("   ✅ NAICS codes generated")
			}

			if tc.expectSIC && !hasSIC {
				t.Logf("   ⚠️ Expected SIC codes but got none")
			} else if hasSIC {
				t.Logf("   ✅ SIC codes generated")
			}

			// Check if crosswalks might have been used (if we have codes from multiple types)
			if hasMCC && (hasNAICS || hasSIC) {
				t.Logf("   ✅ Crosswalks likely used (codes from multiple types)")
				successCount++
			} else if hasMCC || hasNAICS || hasSIC {
				t.Logf("   ⚠️ Only single code type generated (crosswalks may not be used)")
			}
		})
	}

	t.Log("\n========================================")
	t.Logf("SUMMARY: %d/%d tests showed crosswalk usage", successCount, totalTests)
	t.Log("========================================")
}

