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
// Next.js App Router generates HTML files in nested directory structure
// For nested routes: compliance/gap-analysis -> compliance/gap-analysis.html
// For simple routes: dashboard -> dashboard.html
// For dynamic routes: merchant-details/[id] -> merchant-details/[id]/page.html
func (rc *RouteConfig) getNextJSPath(route string) string {
	// Remove leading slash
	route = strings.TrimPrefix(route, "/")
	
	// Handle root route - Next.js generates index.html
	if route == "" || route == "index" {
		return filepath.Join(rc.nextJSBuildPath, "index.html")
	}
	
	// Check if this is a dynamic route (merchant-details with ID)
	// Dynamic routes in Next.js App Router don't generate static HTML per ID
	// Instead, they generate a route handler at: merchant-details/[id]/page.html
	// For static serving, we need to serve the route template or fall back to root
	if strings.HasPrefix(route, "merchant-details/") {
		// Extract the ID part
		parts := strings.Split(route, "/")
		if len(parts) >= 2 {
			// This is a dynamic route: merchant-details/[id]
			// Next.js App Router generates the route handler at: merchant-details/[id]/page.html
			// Check if the dynamic route template exists
			dynamicPath := filepath.Join(rc.nextJSBuildPath, "merchant-details", "[id]", "page.html")
			if _, err := os.Stat(dynamicPath); err == nil {
				return dynamicPath
			}
			// Also try the actual ID path (in case Next.js pre-rendered it)
			idPath := filepath.Join(rc.nextJSBuildPath, route+".html")
			if _, err := os.Stat(idPath); err == nil {
				return idPath
			}
			// If neither exists, serve root index.html for client-side routing
			// Next.js will handle the routing client-side
			return filepath.Join(rc.nextJSBuildPath, "index.html")
		}
	}
	
	// For routes with slashes, Next.js generates files in nested directories
	// e.g., compliance/gap-analysis -> compliance/gap-analysis.html
	if strings.Contains(route, "/") {
		// Try nested structure first (correct for App Router)
		nestedPath := filepath.Join(rc.nextJSBuildPath, route+".html")
		return nestedPath
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
	
	// Check if the file exists
	if _, err := os.Stat(nextJSPath); err == nil {
		// Next.js page exists, serve it
		http.ServeFile(w, r, nextJSPath)
		return
	}
	
	// For dynamic routes or routes that don't have static HTML files,
	// serve the root index.html and let Next.js handle client-side routing
	indexPath := filepath.Join(rc.nextJSBuildPath, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		// Serve index.html for client-side routing
		http.ServeFile(w, r, indexPath)
		return
	}
	
	// Next.js page doesn't exist, serve 404
	// Legacy UI fallback removed in Phase 4
	http.NotFound(w, r)
}

