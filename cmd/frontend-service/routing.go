package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// RouteConfig holds configuration for routing decisions
type RouteConfig struct {
	useNewUI        bool
	nextJSBuildPath string
	legacyPath      string
}

// NewRouteConfig creates a new route configuration
func NewRouteConfig() *RouteConfig {
	// Default to new UI unless explicitly disabled
	// Check for explicit legacy UI flag first
	useLegacyUI := os.Getenv("USE_LEGACY_UI") == "true"
	
	// If legacy UI is explicitly requested, use it
	// Otherwise, default to new UI (check if explicitly enabled OR default to true)
	useNewUI := !useLegacyUI && (os.Getenv("NEXT_PUBLIC_USE_NEW_UI") != "false" && os.Getenv("USE_NEW_UI") != "false")
	
	return &RouteConfig{
		useNewUI:        useNewUI,
		nextJSBuildPath: "./static/.next/server/app",
		legacyPath:      "./static",
	}
}

// shouldUseNewUI checks if we should use the new UI for a given route
func (rc *RouteConfig) shouldUseNewUI(route string) bool {
	if !rc.useNewUI {
		return false
	}
	
	// Check if Next.js page exists for this route
	nextJSPath := rc.getNextJSPath(route)
	if _, err := os.Stat(nextJSPath); err == nil {
		return true
	}
	
	return false
}

// getNextJSPath converts a route to Next.js file path
// Next.js generates HTML files directly in the app directory (e.g., dashboard.html, merchant-hub.html)
func (rc *RouteConfig) getNextJSPath(route string) string {
	// Remove leading slash
	route = strings.TrimPrefix(route, "/")
	
	// Handle root route - Next.js generates index.html
	if route == "" || route == "index" {
		return filepath.Join(rc.nextJSBuildPath, "index.html")
	}
	
	// Next.js generates HTML files directly: route -> route.html
	// For routes with slashes, try multiple patterns:
	// 1. Replace slashes with dashes: compliance/gap-analysis -> compliance-gap-analysis.html
	// 2. Keep nested structure: compliance/gap-analysis -> compliance/gap-analysis/page.html
	// 3. Simple: route -> route.html
	
	// First try: replace slashes with dashes (common Next.js pattern)
	if strings.Contains(route, "/") {
		dashRoute := strings.ReplaceAll(route, "/", "-")
		dashPath := filepath.Join(rc.nextJSBuildPath, dashRoute+".html")
		return dashPath
	}
	
	// Simple route: just add .html
	return filepath.Join(rc.nextJSBuildPath, route+".html")
}

// getLegacyPath gets the legacy HTML file path for a route
func (rc *RouteConfig) getLegacyPath(route string) string {
	route = strings.TrimPrefix(route, "/")
	
	// Map routes to legacy HTML files
	routeMap := map[string]string{
		"":                    "index.html",
		"dashboard":           "dashboard.html",
		"dashboard-hub":       "dashboard-hub.html",
		"add-merchant":        "add-merchant.html",
		"merchant-portfolio":  "merchant-portfolio.html",
		"register":            "register.html",
		"risk-dashboard":      "risk-dashboard.html",
		"enhanced-risk-indicators": "enhanced-risk-indicators.html",
		"compliance":           "compliance-dashboard.html",
		"admin":               "admin-dashboard.html",
		"merchant-hub":         "merchant-hub.html",
		"business-intelligence": "business-intelligence.html",
		"monitoring":           "monitoring-dashboard.html",
		"compliance/gap-analysis": "compliance-gap-analysis.html",
		"compliance/progress-tracking": "compliance-progress-tracking.html",
		"compliance/summary-reports": "compliance-summary-reports.html",
		"compliance/alerts": "compliance-alert-system.html",
		"compliance/framework-indicators": "compliance-framework-indicators.html",
		"merchant-hub/integration": "merchant-hub-integration.html",
		"merchant/bulk-operations": "merchant-bulk-operations.html",
		"merchant/comparison": "merchant-comparison.html",
		"risk-assessment/portfolio": "risk-assessment-portfolio.html",
		"market-analysis": "market-analysis-dashboard.html",
		"competitive-analysis": "competitive-analysis-dashboard.html",
		"business-growth": "business-growth-analytics.html",
		"analytics-insights": "analytics-insights.html",
		"admin/models": "admin-models.html",
		"admin/queue": "admin-queue.html",
		"sessions": "sessions.html",
		"gap-analysis/reports": "gap-analysis-reports.html",
		"gap-tracking": "gap-tracking-system.html",
		"api-test": "api-test.html",
	}
	
	if mappedRoute, ok := routeMap[route]; ok {
		return filepath.Join(rc.legacyPath, mappedRoute)
	}
	
	// Default: try route.html
	return filepath.Join(rc.legacyPath, route+".html")
}

// serveRoute serves Next.js UI (legacy UI has been removed in Phase 4)
func (rc *RouteConfig) serveRoute(w http.ResponseWriter, r *http.Request, route string) {
	// Phase 4: Legacy UI removed - only serve Next.js
	// If legacy UI is explicitly requested, return 404 (no longer available)
	if !rc.useNewUI {
		http.NotFound(w, r)
		return
	}
	
	// Try Next.js page
	nextJSPath := rc.getNextJSPath(route)
	
	// Debug: Log the path we're looking for (only in development)
	if os.Getenv("DEBUG_ROUTING") == "true" {
		log.Printf("DEBUG: Route '%s' -> Looking for: %s", route, nextJSPath)
		if _, err := os.Stat(rc.nextJSBuildPath); err != nil {
			log.Printf("DEBUG: Next.js build path does not exist: %s", rc.nextJSBuildPath)
		}
	}
	
	if _, err := os.Stat(nextJSPath); err == nil {
		// Next.js page exists, serve it
		http.ServeFile(w, r, nextJSPath)
		return
	}
	
	// Next.js page doesn't exist, serve 404
	// Legacy UI fallback removed in Phase 4
	http.NotFound(w, r)
}

