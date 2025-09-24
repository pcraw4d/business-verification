package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMerchantHubIntegration tests the integration between merchant portfolio and existing hub navigation
func TestMerchantHubIntegration(t *testing.T) {
	t.Run("Navigation Integration", func(t *testing.T) {
		// Test that merchant portfolio is properly integrated into navigation
		server := setupTestServer(t)
		defer server.Close()

		// Test dashboard hub includes merchant portfolio
		resp, err := http.Get(server.URL + "/dashboard-hub.html")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Check that merchant portfolio card is present
		body := readResponseBody(t, resp)
		assert.Contains(t, body, "Merchant Portfolio")
		assert.Contains(t, body, "merchant-portfolio.html")
		assert.Contains(t, body, "card-icon merchant")
	})

	t.Run("Navigation Component Integration", func(t *testing.T) {
		// Test that navigation component includes merchant management section
		server := setupTestServer(t)
		defer server.Close()

		resp, err := http.Get(server.URL + "/components/navigation.js")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := readResponseBody(t, resp)
		assert.Contains(t, body, "Merchant Management")
		assert.Contains(t, body, "merchant-portfolio")
		assert.Contains(t, body, "merchant-detail")
	})

	t.Run("Merchant Context Integration", func(t *testing.T) {
		// Test that merchant context component is available
		server := setupTestServer(t)
		defer server.Close()

		resp, err := http.Get(server.URL + "/components/merchant-context.js")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := readResponseBody(t, resp)
		assert.Contains(t, body, "MerchantContext")
		assert.Contains(t, body, "getCurrentMerchant")
		assert.Contains(t, body, "updateMerchantContext")
	})

	t.Run("Dashboard Integration", func(t *testing.T) {
		// Test that existing dashboards include merchant context
		server := setupTestServer(t)
		defer server.Close()

		dashboards := []string{
			"/dashboard.html",
			"/risk-dashboard.html",
			"/compliance-dashboard.html",
		}

		for _, dashboard := range dashboards {
			t.Run("Dashboard "+dashboard, func(t *testing.T) {
				resp, err := http.Get(server.URL + dashboard)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusOK, resp.StatusCode)

				body := readResponseBody(t, resp)
				assert.Contains(t, body, "merchant-context.js")
			})
		}
	})

	t.Run("Backwards Compatibility", func(t *testing.T) {
		// Test that existing functionality is not broken
		server := setupTestServer(t)
		defer server.Close()

		// Test that existing navigation still works
		resp, err := http.Get(server.URL + "/dashboard.html")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := readResponseBody(t, resp)
		// Check that existing navigation elements are still present
		assert.Contains(t, body, "navigation.js")
		// Note: dashboard.html has custom navigation, so we check for the script inclusion
		assert.Contains(t, body, "merchant-context.js")
	})

	t.Run("Merchant Portfolio Navigation", func(t *testing.T) {
		// Test that merchant portfolio page has proper navigation
		server := setupTestServer(t)
		defer server.Close()

		resp, err := http.Get(server.URL + "/merchant-portfolio.html")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := readResponseBody(t, resp)
		assert.Contains(t, body, "navigation.js")
		assert.Contains(t, body, "Merchant Portfolio")
	})

	t.Run("Merchant Detail Navigation", func(t *testing.T) {
		// Test that merchant detail page has proper navigation
		server := setupTestServer(t)
		defer server.Close()

		resp, err := http.Get(server.URL + "/merchant-detail.html")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := readResponseBody(t, resp)
		assert.Contains(t, body, "navigation.js")
		assert.Contains(t, body, "Merchant Detail")
	})
}

// TestMerchantContextIntegration tests the merchant context integration with existing dashboards
func TestMerchantContextIntegration(t *testing.T) {
	t.Run("Context Component Loading", func(t *testing.T) {
		// Test that merchant context component loads properly
		server := setupTestServer(t)
		defer server.Close()

		resp, err := http.Get(server.URL + "/components/merchant-context.js")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := readResponseBody(t, resp)
		assert.Contains(t, body, "class MerchantContext")
		assert.Contains(t, body, "init()")
		assert.Contains(t, body, "createContextUI()")
	})

	t.Run("Context Integration with Dashboards", func(t *testing.T) {
		// Test that merchant context integrates with existing dashboards
		server := setupTestServer(t)
		defer server.Close()

		dashboards := []string{
			"/dashboard.html",
			"/risk-dashboard.html",
			"/compliance-dashboard.html",
		}

		for _, dashboard := range dashboards {
			t.Run("Context in "+dashboard, func(t *testing.T) {
				resp, err := http.Get(server.URL + dashboard)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusOK, resp.StatusCode)

				body := readResponseBody(t, resp)
				// Check that merchant context script is included
				assert.Contains(t, body, "merchant-context.js")
			})
		}
	})

	t.Run("Context Auto-initialization", func(t *testing.T) {
		// Test that merchant context auto-initializes on dashboard pages
		server := setupTestServer(t)
		defer server.Close()

		resp, err := http.Get(server.URL + "/dashboard.html")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := readResponseBody(t, resp)
		// Check that merchant context script is included
		assert.Contains(t, body, "merchant-context.js")
	})
}

// TestNavigationBackwardsCompatibility tests that existing navigation functionality is preserved
func TestNavigationBackwardsCompatibility(t *testing.T) {
	t.Run("Existing Navigation Elements", func(t *testing.T) {
		// Test that all existing navigation elements are still present
		server := setupTestServer(t)
		defer server.Close()

		resp, err := http.Get(server.URL + "/components/navigation.js")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := readResponseBody(t, resp)

		// Check that existing navigation sections are still present
		assert.Contains(t, body, "Platform")
		assert.Contains(t, body, "Core Analytics")
		assert.Contains(t, body, "Compliance")
		assert.Contains(t, body, "Market Intelligence")

		// Check that existing pages are still mapped
		assert.Contains(t, body, "dashboard")
		assert.Contains(t, body, "risk-dashboard")
		assert.Contains(t, body, "compliance-dashboard")
		assert.Contains(t, body, "market-analysis-dashboard")
	})

	t.Run("Existing Page Detection", func(t *testing.T) {
		// Test that existing page detection still works
		server := setupTestServer(t)
		defer server.Close()

		resp, err := http.Get(server.URL + "/components/navigation.js")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := readResponseBody(t, resp)

		// Check that page mapping includes existing pages
		assert.Contains(t, body, "business-intelligence")
		assert.Contains(t, body, "risk-assessment")
		assert.Contains(t, body, "compliance-status")
		assert.Contains(t, body, "market-analysis")
	})

	t.Run("Existing Dashboard Functionality", func(t *testing.T) {
		// Test that existing dashboards still function properly
		server := setupTestServer(t)
		defer server.Close()

		dashboards := []string{
			"/dashboard.html",
			"/risk-dashboard.html",
			"/compliance-dashboard.html",
			"/market-analysis-dashboard.html",
		}

		for _, dashboard := range dashboards {
			t.Run("Dashboard "+dashboard, func(t *testing.T) {
				resp, err := http.Get(server.URL + dashboard)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusOK, resp.StatusCode)

				body := readResponseBody(t, resp)
				// Check that navigation is still present
				assert.Contains(t, body, "navigation.js")
			})
		}
	})
}

// Helper functions for testing

func setupTestServer(t *testing.T) *httptest.Server {
	// Create a test server that serves static files
	mux := http.NewServeMux()

	// Serve static files from web directory
	mux.Handle("/", http.FileServer(http.Dir("../../web/")))

	server := httptest.NewServer(mux)
	return server
}

func readResponseBody(t *testing.T, resp *http.Response) string {
	body := make([]byte, 0)
	buffer := make([]byte, 1024)

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			body = append(body, buffer[:n]...)
		}
		if err != nil {
			break
		}
	}

	return string(body)
}

// TestMerchantHubIntegrationPerformance tests the performance of hub integration
func TestMerchantHubIntegrationPerformance(t *testing.T) {
	t.Run("Navigation Load Performance", func(t *testing.T) {
		// Test that navigation loads quickly
		server := setupTestServer(t)
		defer server.Close()

		start := time.Now()
		resp, err := http.Get(server.URL + "/components/navigation.js")
		require.NoError(t, err)
		defer resp.Body.Close()

		loadTime := time.Since(start)
		assert.Less(t, loadTime, 100*time.Millisecond, "Navigation should load quickly")
	})

	t.Run("Dashboard Load Performance", func(t *testing.T) {
		// Test that dashboards with merchant context load quickly
		server := setupTestServer(t)
		defer server.Close()

		dashboards := []string{
			"/dashboard.html",
			"/merchant-portfolio.html",
			"/merchant-detail.html",
		}

		for _, dashboard := range dashboards {
			t.Run("Performance "+dashboard, func(t *testing.T) {
				start := time.Now()
				resp, err := http.Get(server.URL + dashboard)
				require.NoError(t, err)
				defer resp.Body.Close()

				loadTime := time.Since(start)
				assert.Less(t, loadTime, 500*time.Millisecond, "Dashboard should load quickly")
			})
		}
	})
}
