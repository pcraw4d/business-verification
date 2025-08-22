package enrichment

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// detectCMS detects Content Management Systems
func (tsa *TechnologyStackAnalyzer) detectCMS(content string, headers map[string]string) []Technology {
	var technologies []Technology

	// Check for WordPress
	if tsa.detectWordPress(content, headers) {
		version := tsa.extractWordPressVersion(content, headers)
		technologies = append(technologies, Technology{
			Name:            "WordPress",
			Category:        "cms",
			Type:            "fullstack",
			Version:         version,
			ConfidenceScore: 0.95,
			Evidence:        []string{"WordPress meta tags or wp-content directory detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	// Check for Drupal
	if tsa.detectDrupal(content, headers) {
		version := tsa.extractDrupalVersion(content, headers)
		technologies = append(technologies, Technology{
			Name:            "Drupal",
			Category:        "cms",
			Type:            "fullstack",
			Version:         version,
			ConfidenceScore: 0.90,
			Evidence:        []string{"Drupal meta tags or drupal.js detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	// Check for Joomla
	if tsa.detectJoomla(content, headers) {
		version := tsa.extractJoomlaVersion(content, headers)
		technologies = append(technologies, Technology{
			Name:            "Joomla",
			Category:        "cms",
			Type:            "fullstack",
			Version:         version,
			ConfidenceScore: 0.90,
			Evidence:        []string{"Joomla meta tags or joomla.js detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	// Check for Shopify
	if tsa.detectShopify(content, headers) {
		technologies = append(technologies, Technology{
			Name:            "Shopify",
			Category:        "cms",
			Type:            "ecommerce",
			Version:         "",
			ConfidenceScore: 0.95,
			Evidence:        []string{"Shopify liquid templates or shopify.com domain detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	// Check for Wix
	if tsa.detectWix(content, headers) {
		technologies = append(technologies, Technology{
			Name:            "Wix",
			Category:        "cms",
			Type:            "website_builder",
			Version:         "",
			ConfidenceScore: 0.90,
			Evidence:        []string{"Wix scripts or wix.com domain detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	// Check for Squarespace
	if tsa.detectSquarespace(content, headers) {
		technologies = append(technologies, Technology{
			Name:            "Squarespace",
			Category:        "cms",
			Type:            "website_builder",
			Version:         "",
			ConfidenceScore: 0.90,
			Evidence:        []string{"Squarespace scripts or squarespace.com domain detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	return technologies
}

// detectFrameworks detects frontend and backend frameworks
func (tsa *TechnologyStackAnalyzer) detectFrameworks(content string, headers map[string]string) []Technology {
	var technologies []Technology

	// Frontend frameworks
	if tsa.detectReact(content, headers) {
		version := tsa.extractReactVersion(content, headers)
		technologies = append(technologies, Technology{
			Name:            "React",
			Category:        "framework",
			Type:            "frontend",
			Version:         version,
			ConfidenceScore: 0.85,
			Evidence:        []string{"React.js scripts or react components detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	if tsa.detectVue(content, headers) {
		version := tsa.extractVueVersion(content, headers)
		technologies = append(technologies, Technology{
			Name:            "Vue.js",
			Category:        "framework",
			Type:            "frontend",
			Version:         version,
			ConfidenceScore: 0.85,
			Evidence:        []string{"Vue.js scripts or vue components detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	if tsa.detectAngular(content, headers) {
		version := tsa.extractAngularVersion(content, headers)
		technologies = append(technologies, Technology{
			Name:            "Angular",
			Category:        "framework",
			Type:            "frontend",
			Version:         version,
			ConfidenceScore: 0.85,
			Evidence:        []string{"Angular scripts or ng- attributes detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	// Backend frameworks
	if tsa.detectLaravel(content, headers) {
		technologies = append(technologies, Technology{
			Name:            "Laravel",
			Category:        "framework",
			Type:            "backend",
			Version:         "",
			ConfidenceScore: 0.80,
			Evidence:        []string{"Laravel CSRF tokens or laravel session detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	if tsa.detectDjango(content, headers) {
		technologies = append(technologies, Technology{
			Name:            "Django",
			Category:        "framework",
			Type:            "backend",
			Version:         "",
			ConfidenceScore: 0.80,
			Evidence:        []string{"Django CSRF tokens or django session detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	if tsa.detectRails(content, headers) {
		technologies = append(technologies, Technology{
			Name:            "Ruby on Rails",
			Category:        "framework",
			Type:            "backend",
			Version:         "",
			ConfidenceScore: 0.80,
			Evidence:        []string{"Rails CSRF tokens or rails session detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	return technologies
}

// detectTools detects development and deployment tools
func (tsa *TechnologyStackAnalyzer) detectTools(content string, headers map[string]string) []Technology {
	var technologies []Technology

	// Build tools
	if tsa.detectWebpack(content, headers) {
		technologies = append(technologies, Technology{
			Name:            "Webpack",
			Category:        "tool",
			Type:            "build",
			Version:         "",
			ConfidenceScore: 0.75,
			Evidence:        []string{"Webpack bundle or webpack scripts detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	if tsa.detectVite(content, headers) {
		technologies = append(technologies, Technology{
			Name:            "Vite",
			Category:        "tool",
			Type:            "build",
			Version:         "",
			ConfidenceScore: 0.75,
			Evidence:        []string{"Vite scripts or vite bundle detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	// Package managers
	if tsa.detectNPM(content, headers) {
		technologies = append(technologies, Technology{
			Name:            "npm",
			Category:        "tool",
			Type:            "package_manager",
			Version:         "",
			ConfidenceScore: 0.70,
			Evidence:        []string{"npm packages or node_modules detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	return technologies
}

// detectAnalytics detects analytics and tracking services
func (tsa *TechnologyStackAnalyzer) detectAnalytics(content string, headers map[string]string) []Technology {
	var technologies []Technology

	// Google Analytics
	if tsa.detectGoogleAnalytics(content, headers) {
		technologies = append(technologies, Technology{
			Name:            "Google Analytics",
			Category:        "analytics",
			Type:            "tracking",
			Version:         "",
			ConfidenceScore: 0.95,
			Evidence:        []string{"Google Analytics scripts or ga() function detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	// Google Tag Manager
	if tsa.detectGoogleTagManager(content, headers) {
		technologies = append(technologies, Technology{
			Name:            "Google Tag Manager",
			Category:        "analytics",
			Type:            "tag_management",
			Version:         "",
			ConfidenceScore: 0.95,
			Evidence:        []string{"Google Tag Manager scripts or gtm.js detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	// Facebook Pixel
	if tsa.detectFacebookPixel(content, headers) {
		technologies = append(technologies, Technology{
			Name:            "Facebook Pixel",
			Category:        "analytics",
			Type:            "tracking",
			Version:         "",
			ConfidenceScore: 0.90,
			Evidence:        []string{"Facebook Pixel scripts or fbq() function detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	// Hotjar
	if tsa.detectHotjar(content, headers) {
		technologies = append(technologies, Technology{
			Name:            "Hotjar",
			Category:        "analytics",
			Type:            "heatmap",
			Version:         "",
			ConfidenceScore: 0.90,
			Evidence:        []string{"Hotjar scripts or _hjSettings detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	return technologies
}

// detectHosting detects hosting and infrastructure services
func (tsa *TechnologyStackAnalyzer) detectHosting(content string, headers map[string]string) []Technology {
	var technologies []Technology

	// Check headers for hosting indicators
	for key, value := range headers {
		lowerKey := strings.ToLower(key)
		lowerValue := strings.ToLower(value)

		// AWS
		if strings.Contains(lowerKey, "x-amz") || strings.Contains(lowerValue, "aws") {
			technologies = append(technologies, Technology{
				Name:            "Amazon Web Services",
				Category:        "hosting",
				Type:            "cloud",
				Version:         "",
				ConfidenceScore: 0.85,
				Evidence:        []string{fmt.Sprintf("AWS header detected: %s", key)},
				DetectedAt:      time.Now(),
				Source:          "headers",
			})
		}

		// Cloudflare
		if strings.Contains(lowerKey, "cf-") || strings.Contains(lowerValue, "cloudflare") {
			technologies = append(technologies, Technology{
				Name:            "Cloudflare",
				Category:        "hosting",
				Type:            "cdn",
				Version:         "",
				ConfidenceScore: 0.90,
				Evidence:        []string{fmt.Sprintf("Cloudflare header detected: %s", key)},
				DetectedAt:      time.Now(),
				Source:          "headers",
			})
		}

		// Vercel
		if strings.Contains(lowerKey, "x-vercel") || strings.Contains(lowerValue, "vercel") {
			technologies = append(technologies, Technology{
				Name:            "Vercel",
				Category:        "hosting",
				Type:            "platform",
				Version:         "",
				ConfidenceScore: 0.90,
				Evidence:        []string{fmt.Sprintf("Vercel header detected: %s", key)},
				DetectedAt:      time.Now(),
				Source:          "headers",
			})
		}

		// Netlify
		if strings.Contains(lowerKey, "x-nf") || strings.Contains(lowerValue, "netlify") {
			technologies = append(technologies, Technology{
				Name:            "Netlify",
				Category:        "hosting",
				Type:            "platform",
				Version:         "",
				ConfidenceScore: 0.90,
				Evidence:        []string{fmt.Sprintf("Netlify header detected: %s", key)},
				DetectedAt:      time.Now(),
				Source:          "headers",
			})
		}
	}

	return technologies
}

// detectDatabases detects database technologies
func (tsa *TechnologyStackAnalyzer) detectDatabases(content string, headers map[string]string) []Technology {
	var technologies []Technology

	// Check for database indicators in content
	if strings.Contains(content, "mysql") || strings.Contains(content, "mysqli") {
		technologies = append(technologies, Technology{
			Name:            "MySQL",
			Category:        "database",
			Type:            "relational",
			Version:         "",
			ConfidenceScore: 0.70,
			Evidence:        []string{"MySQL functions or references detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	if strings.Contains(content, "postgresql") || strings.Contains(content, "postgres") {
		technologies = append(technologies, Technology{
			Name:            "PostgreSQL",
			Category:        "database",
			Type:            "relational",
			Version:         "",
			ConfidenceScore: 0.70,
			Evidence:        []string{"PostgreSQL references detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	if strings.Contains(content, "mongodb") || strings.Contains(content, "mongo") {
		technologies = append(technologies, Technology{
			Name:            "MongoDB",
			Category:        "database",
			Type:            "nosql",
			Version:         "",
			ConfidenceScore: 0.70,
			Evidence:        []string{"MongoDB references detected"},
			DetectedAt:      time.Now(),
			Source:          "html",
		})
	}

	return technologies
}

// detectSecurity detects security and CDN services
func (tsa *TechnologyStackAnalyzer) detectSecurity(content string, headers map[string]string) []Technology {
	var technologies []Technology

	// Check headers for security services
	for key, value := range headers {
		lowerKey := strings.ToLower(key)
		lowerValue := strings.ToLower(value)

		// Cloudflare (also security)
		if strings.Contains(lowerKey, "cf-") || strings.Contains(lowerValue, "cloudflare") {
			technologies = append(technologies, Technology{
				Name:            "Cloudflare",
				Category:        "security",
				Type:            "cdn_security",
				Version:         "",
				ConfidenceScore: 0.90,
				Evidence:        []string{fmt.Sprintf("Cloudflare security header: %s", key)},
				DetectedAt:      time.Now(),
				Source:          "headers",
			})
		}

		// AWS WAF
		if strings.Contains(lowerKey, "x-amz-waf") || strings.Contains(lowerValue, "waf") {
			technologies = append(technologies, Technology{
				Name:            "AWS WAF",
				Category:        "security",
				Type:            "waf",
				Version:         "",
				ConfidenceScore: 0.85,
				Evidence:        []string{fmt.Sprintf("AWS WAF header: %s", key)},
				DetectedAt:      time.Now(),
				Source:          "headers",
			})
		}
	}

	return technologies
}

// Helper detection methods
func (tsa *TechnologyStackAnalyzer) detectWordPress(content string, headers map[string]string) bool {
	patterns := []string{
		`wp-content`,
		`wp-includes`,
		`wordpress`,
		`wp-admin`,
		`wp-json`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	// Check meta tags
	if strings.Contains(content, `<meta name="generator" content="wordpress`) {
		return true
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectDrupal(content string, headers map[string]string) bool {
	patterns := []string{
		`drupal`,
		`drupal.js`,
		`drupal.css`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	// Check meta tags
	if strings.Contains(content, `<meta name="generator" content="drupal`) {
		return true
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectJoomla(content string, headers map[string]string) bool {
	patterns := []string{
		`joomla`,
		`joomla.js`,
		`joomla.css`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	// Check meta tags
	if strings.Contains(content, `<meta name="generator" content="joomla`) {
		return true
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectShopify(content string, headers map[string]string) bool {
	patterns := []string{
		`shopify`,
		`liquid`,
		`shopify.com`,
		`myshopify.com`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectWix(content string, headers map[string]string) bool {
	patterns := []string{
		`wix`,
		`wix.com`,
		`wixsite.com`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectSquarespace(content string, headers map[string]string) bool {
	patterns := []string{
		`squarespace`,
		`squarespace.com`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectReact(content string, headers map[string]string) bool {
	patterns := []string{
		`react`,
		`react.js`,
		`react-dom`,
		`__react`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectVue(content string, headers map[string]string) bool {
	patterns := []string{
		`vue`,
		`vue.js`,
		`v-`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectAngular(content string, headers map[string]string) bool {
	patterns := []string{
		`angular`,
		`angular.js`,
		`ng-`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectLaravel(content string, headers map[string]string) bool {
	patterns := []string{
		`laravel`,
		`csrf-token`,
		`laravel_session`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectDjango(content string, headers map[string]string) bool {
	patterns := []string{
		`django`,
		`csrfmiddlewaretoken`,
		`django_session`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectRails(content string, headers map[string]string) bool {
	patterns := []string{
		`rails`,
		`authenticity_token`,
		`_rails`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectWebpack(content string, headers map[string]string) bool {
	patterns := []string{
		`webpack`,
		`webpack.js`,
		`webpackJsonp`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectVite(content string, headers map[string]string) bool {
	patterns := []string{
		`vite`,
		`__vite_`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectNPM(content string, headers map[string]string) bool {
	patterns := []string{
		`node_modules`,
		`npm`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectGoogleAnalytics(content string, headers map[string]string) bool {
	patterns := []string{
		`google-analytics`,
		`ga(`,
		`gtag(`,
		`googletagmanager`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectGoogleTagManager(content string, headers map[string]string) bool {
	patterns := []string{
		`googletagmanager`,
		`gtm.js`,
		`gtag`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectFacebookPixel(content string, headers map[string]string) bool {
	patterns := []string{
		`facebook`,
		`fbq(`,
		`fbevents`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (tsa *TechnologyStackAnalyzer) detectHotjar(content string, headers map[string]string) bool {
	patterns := []string{
		`hotjar`,
		`_hj`,
		`hjsv`,
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

// Version extraction methods
func (tsa *TechnologyStackAnalyzer) extractWordPressVersion(content string, headers map[string]string) string {
	// Look for WordPress version in meta tags
	re := regexp.MustCompile(`<meta name="generator" content="wordpress ([^"]+)"`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func (tsa *TechnologyStackAnalyzer) extractDrupalVersion(content string, headers map[string]string) string {
	// Look for Drupal version in meta tags
	re := regexp.MustCompile(`<meta name="generator" content="drupal ([^"]+)"`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func (tsa *TechnologyStackAnalyzer) extractJoomlaVersion(content string, headers map[string]string) string {
	// Look for Joomla version in meta tags
	re := regexp.MustCompile(`<meta name="generator" content="joomla ([^"]+)"`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func (tsa *TechnologyStackAnalyzer) extractReactVersion(content string, headers map[string]string) string {
	// Look for React version in script tags
	re := regexp.MustCompile(`react@([^/]+)`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func (tsa *TechnologyStackAnalyzer) extractVueVersion(content string, headers map[string]string) string {
	// Look for Vue version in script tags
	re := regexp.MustCompile(`vue@([^/]+)`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func (tsa *TechnologyStackAnalyzer) extractAngularVersion(content string, headers map[string]string) string {
	// Look for Angular version in script tags
	re := regexp.MustCompile(`angular@([^/]+)`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// Helper methods for result analysis
func (tsa *TechnologyStackAnalyzer) determinePrimaryTechnology(technologies []Technology) *Technology {
	if len(technologies) == 0 {
		return nil
	}

	// Return the technology with highest confidence
	primary := &technologies[0]
	for i := 1; i < len(technologies); i++ {
		if technologies[i].ConfidenceScore > primary.ConfidenceScore {
			primary = &technologies[i]
		}
	}

	return primary
}

func (tsa *TechnologyStackAnalyzer) determineStackType(technologyStack *TechnologyStack) string {
	// Check for modern stack indicators
	if len(technologyStack.Frameworks) > 0 {
		for _, framework := range technologyStack.Frameworks {
			if framework.Name == "React" || framework.Name == "Vue.js" || framework.Name == "Angular" {
				return "modern"
			}
		}
	}

	// Check for headless CMS
	if len(technologyStack.CMS) > 0 && len(technologyStack.Frameworks) > 0 {
		return "headless"
	}

	// Check for traditional CMS
	if len(technologyStack.CMS) > 0 && len(technologyStack.Frameworks) == 0 {
		return "traditional"
	}

	// Check for static site
	if len(technologyStack.Tools) > 0 {
		for _, tool := range technologyStack.Tools {
			if tool.Name == "Vite" || tool.Name == "Webpack" {
				return "static"
			}
		}
	}

	return "unknown"
}

func (tsa *TechnologyStackAnalyzer) determineComplexity(technologyStack *TechnologyStack) string {
	totalTechnologies := len(technologyStack.CMS) + len(technologyStack.Frameworks) +
		len(technologyStack.Tools) + len(technologyStack.Analytics) +
		len(technologyStack.Hosting) + len(technologyStack.Database) +
		len(technologyStack.Security)

	if totalTechnologies <= 3 {
		return "simple"
	} else if totalTechnologies <= 8 {
		return "moderate"
	} else {
		return "complex"
	}
}

func (tsa *TechnologyStackAnalyzer) calculateOverallConfidence(technologyStack *TechnologyStack) float64 {
	allTechnologies := []Technology{}
	allTechnologies = append(allTechnologies, technologyStack.CMS...)
	allTechnologies = append(allTechnologies, technologyStack.Frameworks...)
	allTechnologies = append(allTechnologies, technologyStack.Tools...)
	allTechnologies = append(allTechnologies, technologyStack.Analytics...)
	allTechnologies = append(allTechnologies, technologyStack.Hosting...)
	allTechnologies = append(allTechnologies, technologyStack.Database...)
	allTechnologies = append(allTechnologies, technologyStack.Security...)

	if len(allTechnologies) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, tech := range allTechnologies {
		totalConfidence += tech.ConfidenceScore
	}

	return totalConfidence / float64(len(allTechnologies))
}

func (tsa *TechnologyStackAnalyzer) collectEvidence(technologyStack *TechnologyStack) []string {
	var evidence []string

	allTechnologies := []Technology{}
	allTechnologies = append(allTechnologies, technologyStack.CMS...)
	allTechnologies = append(allTechnologies, technologyStack.Frameworks...)
	allTechnologies = append(allTechnologies, technologyStack.Tools...)
	allTechnologies = append(allTechnologies, technologyStack.Analytics...)
	allTechnologies = append(allTechnologies, technologyStack.Hosting...)
	allTechnologies = append(allTechnologies, technologyStack.Database...)
	allTechnologies = append(allTechnologies, technologyStack.Security...)

	for _, tech := range allTechnologies {
		evidence = append(evidence, tech.Evidence...)
	}

	return evidence
}

// getDefaultTechnologyStackAnalyzerConfig returns default configuration
func getDefaultTechnologyStackAnalyzerConfig() *TechnologyStackAnalyzerConfig {
	return &TechnologyStackAnalyzerConfig{
		MinConfidenceScore:         0.3,
		MaxTechnologiesPerCategory: 10,
		CMSIndicators: map[string][]string{
			"wordpress":   {"wp-content", "wp-includes", "wordpress"},
			"drupal":      {"drupal", "drupal.js"},
			"joomla":      {"joomla", "joomla.js"},
			"shopify":     {"shopify", "liquid"},
			"wix":         {"wix", "wix.com"},
			"squarespace": {"squarespace", "squarespace.com"},
		},
		FrameworkIndicators: map[string][]string{
			"react":   {"react", "react.js", "react-dom"},
			"vue":     {"vue", "vue.js", "v-"},
			"angular": {"angular", "angular.js", "ng-"},
			"laravel": {"laravel", "csrf-token"},
			"django":  {"django", "csrfmiddlewaretoken"},
			"rails":   {"rails", "authenticity_token"},
		},
		ToolIndicators: map[string][]string{
			"webpack": {"webpack", "webpack.js"},
			"vite":    {"vite", "__vite_"},
			"npm":     {"node_modules", "npm"},
		},
		AnalyticsIndicators: map[string][]string{
			"google_analytics": {"google-analytics", "ga(", "gtag("},
			"gtm":              {"googletagmanager", "gtm.js"},
			"facebook_pixel":   {"facebook", "fbq(", "fbevents"},
			"hotjar":           {"hotjar", "_hj", "hjsv"},
		},
		HostingIndicators: map[string][]string{
			"aws":        {"x-amz", "aws"},
			"cloudflare": {"cf-", "cloudflare"},
			"vercel":     {"x-vercel", "vercel"},
			"netlify":    {"x-nf", "netlify"},
		},
	}
}
