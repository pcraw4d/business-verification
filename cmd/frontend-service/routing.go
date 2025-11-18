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

// getClientAppPath returns the path to the Next.js client app entry point
// For dynamic routes, we need to serve the client-side app which handles routing
func (rc *RouteConfig) getClientAppPath() string {
	// Next.js generates the client app in .next/static
	// But for routing, we need the HTML shell that loads the client bundle
	// Try the root index.html first
	indexPath := filepath.Join(rc.nextJSBuildPath, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		return indexPath
	}
	// Fallback: try to find any HTML file that can serve as the app shell
	// Next.js App Router might generate app.html or similar
	appPath := filepath.Join(rc.nextJSBuildPath, "app.html")
	if _, err := os.Stat(appPath); err == nil {
		return appPath
	}
	return indexPath
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
			// ALWAYS try to serve the dynamic route template first
			dynamicPath := filepath.Join(rc.nextJSBuildPath, "merchant-details", "[id]", "page.html")
			if _, err := os.Stat(dynamicPath); err == nil {
				return dynamicPath
			}
			// If dynamic route template doesn't exist, we need to serve a page that can handle it
			// Try serving the merchant-details layout page if it exists
			merchantDetailsPath := filepath.Join(rc.nextJSBuildPath, "merchant-details", "page.html")
			if _, err := os.Stat(merchantDetailsPath); err == nil {
				return merchantDetailsPath
			}
			// Last resort: serve root index.html but this should be handled by Next.js router
			// The issue is that index.html is the home page which auto-redirects
			// So we need to ensure the dynamic route template exists
			return dynamicPath // Return the expected path even if it doesn't exist - serveRoute will handle fallback
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

// findAvailablePage attempts to find an available Next.js page by checking multiple path structures
// This handles variations in how Next.js App Router might generate pages
func (rc *RouteConfig) findAvailablePage(routeName string) string {
	// Try multiple path structures that Next.js App Router might use
	possiblePaths := []string{
		// App Router nested structure: route/page.html
		filepath.Join(rc.nextJSBuildPath, routeName, "page.html"),
		// Flat structure: route.html
		filepath.Join(rc.nextJSBuildPath, routeName+".html"),
		// Alternative nested: route/index.html
		filepath.Join(rc.nextJSBuildPath, routeName, "index.html"),
	}
	
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	
	return ""
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
	
	// For dynamic routes (merchant-details/[id]), Next.js App Router doesn't generate static HTML
	// We need to serve a client-side app shell that Next.js can use for client-side routing
	if strings.HasPrefix(route, "merchant-details/") {
		// Try to find the dynamic route template first
		merchantDetailsBase := filepath.Join(rc.nextJSBuildPath, "merchant-details")
		if _, err := os.Stat(merchantDetailsBase); err == nil {
			// Directory exists, try to serve the [id] template
			dynamicTemplate := filepath.Join(merchantDetailsBase, "[id]", "page.html")
			if _, err := os.Stat(dynamicTemplate); err == nil {
				http.ServeFile(w, r, dynamicTemplate)
				return
			}
		}
		
		// Template doesn't exist - Next.js App Router doesn't generate static HTML for dynamic routes
		// We need to serve a page that loads the Next.js client bundle for client-side routing
		// Try serving merchant-portfolio page (it has Next.js client bundle and won't redirect)
		// Check multiple possible path structures
		merchantPortfolioPath := rc.findAvailablePage("merchant-portfolio")
		if merchantPortfolioPath != "" {
			// Serve merchant-portfolio page which loads Next.js client bundle
			// Next.js router will detect the URL (/merchant-details/{id}) and route client-side
			if os.Getenv("DEBUG_ROUTING") == "true" {
				log.Printf("DEBUG: Found merchant-portfolio at: %s", merchantPortfolioPath)
			}
			log.Printf("INFO: Serving merchant-portfolio page for dynamic route: %s (Next.js will route client-side)", route)
			http.ServeFile(w, r, merchantPortfolioPath)
			return
		}
		
		// Try add-merchant page as fallback (also has Next.js client bundle)
		addMerchantPath := rc.findAvailablePage("add-merchant")
		if addMerchantPath != "" {
			if os.Getenv("DEBUG_ROUTING") == "true" {
				log.Printf("DEBUG: Found add-merchant at: %s", addMerchantPath)
			}
			log.Printf("INFO: Serving add-merchant page for dynamic route: %s (Next.js will route client-side)", route)
			http.ServeFile(w, r, addMerchantPath)
			return
		}
		
		// Debug: Log what we're checking if DEBUG_ROUTING is enabled
		if os.Getenv("DEBUG_ROUTING") == "true" {
			log.Printf("DEBUG: Checking for fallback pages for route: %s", route)
			log.Printf("DEBUG: Next.js build path: %s", rc.nextJSBuildPath)
			// List what files actually exist
			if entries, err := os.ReadDir(rc.nextJSBuildPath); err == nil {
				log.Printf("DEBUG: Files in build path:")
				for _, entry := range entries {
					log.Printf("DEBUG:   - %s (dir: %v)", entry.Name(), entry.IsDir())
				}
			}
		}
		
		// Last resort: serve any available page that has Next.js client bundle
		// The client-side router will handle the actual route based on the URL
		clientAppPath := rc.getClientAppPath()
		if _, err := os.Stat(clientAppPath); err == nil {
			log.Printf("INFO: Serving client app for dynamic route: %s (template not found, Next.js will route client-side)", route)
			http.ServeFile(w, r, clientAppPath)
			return
		}
		
		log.Printf("WARNING: Dynamic route template not found for: %s, and no client app available", route)
	}
	
	// For other routes, try serving root index.html for client-side routing
	indexPath := filepath.Join(rc.nextJSBuildPath, "index.html")
	if _, err := os.Stat(indexPath); err == nil && !strings.HasPrefix(route, "merchant-details/") {
		// Serve index.html for client-side routing (but not for merchant-details to avoid redirect)
		http.ServeFile(w, r, indexPath)
		return
	}
	
	// Next.js page doesn't exist, serve 404
	// Legacy UI fallback removed in Phase 4
	http.NotFound(w, r)
}

