package enrichment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTechnologyStackAnalyzer_NewTechnologyStackAnalyzer(t *testing.T) {
	tests := []struct {
		name        string
		config      *TechnologyStackAnalyzerConfig
		expectError bool
	}{
		{
			name:        "valid config",
			config:      getDefaultTechnologyStackAnalyzerConfig(),
			expectError: false,
		},
		{
			name:        "nil config",
			config:      nil,
			expectError: false, // Should use default config
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewTechnologyStackAnalyzer(zap.NewNop(), tt.config)

			if tt.expectError {
				assert.Nil(t, analyzer)
			} else {
				assert.NotNil(t, analyzer)
				assert.NotNil(t, analyzer.config)
				assert.NotNil(t, analyzer.logger)
			}
		})
	}
}

func TestTechnologyStackAnalyzer_AnalyzeTechnologyStack(t *testing.T) {
	analyzer := NewTechnologyStackAnalyzer(zap.NewNop(), nil)
	require.NotNil(t, analyzer)

	tests := []struct {
		name           string
		content        string
		headers        map[string]string
		expectedResult *TechnologyStackResult
		expectError    bool
	}{
		{
			name:    "wordpress site",
			content: `<html><head><meta name="generator" content="wordpress 6.0"><link rel="stylesheet" href="/wp-content/themes/default/style.css"></head><body>WordPress site</body></html>`,
			headers: map[string]string{
				"Server": "Apache/2.4.41",
			},
			expectedResult: &TechnologyStackResult{
				StackType:       "traditional",
				Complexity:      "simple",
				ConfidenceScore: 0.95,
				TechnologyStack: &TechnologyStack{
					CMS: []Technology{
						{
							Name:            "WordPress",
							Category:        "cms",
							Type:            "fullstack",
							Version:         "6.0",
							ConfidenceScore: 0.95,
							Evidence:        []string{"WordPress meta tags or wp-content directory detected"},
							DetectedAt:      time.Now(),
							Source:          "html",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name:    "react app with analytics",
			content: `<html><head><script src="https://unpkg.com/react@18.2.0/umd/react.development.js"></script><script>gtag('config', 'GA_MEASUREMENT_ID');</script></head><body>React app</body></html>`,
			headers: map[string]string{
				"X-Vercel": "vercel",
			},
			expectedResult: &TechnologyStackResult{
				StackType:       "modern",
				Complexity:      "moderate",
				ConfidenceScore: 0.90,
				TechnologyStack: &TechnologyStack{
					Frameworks: []Technology{
						{
							Name:            "React",
							Category:        "framework",
							Type:            "frontend",
							Version:         "18.2.0",
							ConfidenceScore: 0.85,
							Evidence:        []string{"React.js scripts or react components detected"},
							DetectedAt:      time.Now(),
							Source:          "html",
						},
					},
					Analytics: []Technology{
						{
							Name:            "Google Analytics",
							Category:        "analytics",
							Type:            "tracking",
							Version:         "",
							ConfidenceScore: 0.95,
							Evidence:        []string{"Google Analytics scripts or ga() function detected"},
							DetectedAt:      time.Now(),
							Source:          "html",
						},
					},
					Hosting: []Technology{
						{
							Name:            "Vercel",
							Category:        "hosting",
							Type:            "platform",
							Version:         "",
							ConfidenceScore: 0.90,
							Evidence:        []string{"Vercel header detected: X-Vercel"},
							DetectedAt:      time.Now(),
							Source:          "headers",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name:    "shopify store",
			content: `<html><head><script src="https://cdn.shopify.com/s/files/1/0000/0000/t/1/assets/theme.js"></script></head><body>{{ product.title }}</body></html>`,
			headers: map[string]string{},
			expectedResult: &TechnologyStackResult{
				StackType:       "traditional",
				Complexity:      "simple",
				ConfidenceScore: 0.95,
				TechnologyStack: &TechnologyStack{
					CMS: []Technology{
						{
							Name:            "Shopify",
							Category:        "cms",
							Type:            "ecommerce",
							Version:         "",
							ConfidenceScore: 0.95,
							Evidence:        []string{"Shopify liquid templates or shopify.com domain detected"},
							DetectedAt:      time.Now(),
							Source:          "html",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name:    "complex modern stack",
			content: `<html><head><script src="https://unpkg.com/react@18.2.0/umd/react.development.js"></script><script src="https://unpkg.com/vue@3.2.0/dist/vue.global.js"></script><script>gtag('config', 'GA_MEASUREMENT_ID');</script><script>fbq('init', '123456789');</script></head><body>Modern app</body></html>`,
			headers: map[string]string{
				"X-Vercel": "vercel",
				"CF-Ray":   "cloudflare",
			},
			expectedResult: &TechnologyStackResult{
				StackType:       "modern",
				Complexity:      "moderate",
				ConfidenceScore: 0.89,
				TechnologyStack: &TechnologyStack{
					Frameworks: []Technology{
						{
							Name:            "React",
							Category:        "framework",
							Type:            "frontend",
							Version:         "18.2.0",
							ConfidenceScore: 0.85,
							Evidence:        []string{"React.js scripts or react components detected"},
							DetectedAt:      time.Now(),
							Source:          "html",
						},
						{
							Name:            "Vue.js",
							Category:        "framework",
							Type:            "frontend",
							Version:         "3.2.0",
							ConfidenceScore: 0.85,
							Evidence:        []string{"Vue.js scripts or vue components detected"},
							DetectedAt:      time.Now(),
							Source:          "html",
						},
					},
					Analytics: []Technology{
						{
							Name:            "Google Analytics",
							Category:        "analytics",
							Type:            "tracking",
							Version:         "",
							ConfidenceScore: 0.95,
							Evidence:        []string{"Google Analytics scripts or ga() function detected"},
							DetectedAt:      time.Now(),
							Source:          "html",
						},
						{
							Name:            "Facebook Pixel",
							Category:        "analytics",
							Type:            "tracking",
							Version:         "",
							ConfidenceScore: 0.90,
							Evidence:        []string{"Facebook Pixel scripts or fbq() function detected"},
							DetectedAt:      time.Now(),
							Source:          "html",
						},
					},
					Hosting: []Technology{
						{
							Name:            "Vercel",
							Category:        "hosting",
							Type:            "platform",
							Version:         "",
							ConfidenceScore: 0.90,
							Evidence:        []string{"Vercel header detected: X-Vercel"},
							DetectedAt:      time.Now(),
							Source:          "headers",
						},
						{
							Name:            "Cloudflare",
							Category:        "hosting",
							Type:            "cdn",
							Version:         "",
							ConfidenceScore: 0.90,
							Evidence:        []string{"Cloudflare header detected: CF-Ray"},
							DetectedAt:      time.Now(),
							Source:          "headers",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name:    "empty content",
			content: "",
			headers: map[string]string{},
			expectedResult: &TechnologyStackResult{
				StackType:       "unknown",
				Complexity:      "simple",
				ConfidenceScore: 0.0,
				TechnologyStack: &TechnologyStack{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := analyzer.AnalyzeTechnologyStack(ctx, tt.content, tt.headers)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Check basic fields
				assert.Equal(t, tt.expectedResult.StackType, result.StackType)
				assert.Equal(t, tt.expectedResult.Complexity, result.Complexity)

				// Check confidence with tolerance
				assert.InDelta(t, tt.expectedResult.ConfidenceScore, result.ConfidenceScore, 0.1)

				// Check technology stack structure
				assert.NotNil(t, result.TechnologyStack)

				// Check specific technologies based on expected result
				if len(tt.expectedResult.TechnologyStack.CMS) > 0 {
					assert.Len(t, result.TechnologyStack.CMS, len(tt.expectedResult.TechnologyStack.CMS))
					for i, expectedCMS := range tt.expectedResult.TechnologyStack.CMS {
						assert.Equal(t, expectedCMS.Name, result.TechnologyStack.CMS[i].Name)
						assert.Equal(t, expectedCMS.Category, result.TechnologyStack.CMS[i].Category)
						if expectedCMS.Version != "" {
							assert.Equal(t, expectedCMS.Version, result.TechnologyStack.CMS[i].Version)
						}
					}
				}

				if len(tt.expectedResult.TechnologyStack.Frameworks) > 0 {
					assert.Len(t, result.TechnologyStack.Frameworks, len(tt.expectedResult.TechnologyStack.Frameworks))
					for i, expectedFramework := range tt.expectedResult.TechnologyStack.Frameworks {
						assert.Equal(t, expectedFramework.Name, result.TechnologyStack.Frameworks[i].Name)
						assert.Equal(t, expectedFramework.Category, result.TechnologyStack.Frameworks[i].Category)
						if expectedFramework.Version != "" {
							assert.Equal(t, expectedFramework.Version, result.TechnologyStack.Frameworks[i].Version)
						}
					}
				}

				// For analytics, just check that the expected ones are present (may be more detected)
				if len(tt.expectedResult.TechnologyStack.Analytics) > 0 {
					assert.GreaterOrEqual(t, len(result.TechnologyStack.Analytics), len(tt.expectedResult.TechnologyStack.Analytics))
					// Check that expected analytics are present
					for _, expectedAnalytics := range tt.expectedResult.TechnologyStack.Analytics {
						found := false
						for _, actualAnalytics := range result.TechnologyStack.Analytics {
							if actualAnalytics.Name == expectedAnalytics.Name {
								assert.Equal(t, expectedAnalytics.Category, actualAnalytics.Category)
								found = true
								break
							}
						}
						assert.True(t, found, "Expected analytics %s not found", expectedAnalytics.Name)
					}
				}

				// For hosting, just check that the expected ones are present (may be more detected)
				if len(tt.expectedResult.TechnologyStack.Hosting) > 0 {
					assert.GreaterOrEqual(t, len(result.TechnologyStack.Hosting), len(tt.expectedResult.TechnologyStack.Hosting))
					// Check that expected hosting are present
					for _, expectedHosting := range tt.expectedResult.TechnologyStack.Hosting {
						found := false
						for _, actualHosting := range result.TechnologyStack.Hosting {
							if actualHosting.Name == expectedHosting.Name {
								assert.Equal(t, expectedHosting.Category, actualHosting.Category)
								found = true
								break
							}
						}
						assert.True(t, found, "Expected hosting %s not found", expectedHosting.Name)
					}
				}
			}
		})
	}
}

func TestTechnologyStackAnalyzer_DetectCMS(t *testing.T) {
	analyzer := NewTechnologyStackAnalyzer(zap.NewNop(), nil)
	require.NotNil(t, analyzer)

	tests := []struct {
		name          string
		content       string
		headers       map[string]string
		expectedCMS   []string
		expectedCount int
	}{
		{
			name:          "wordpress detection",
			content:       `<html><head><meta name="generator" content="wordpress 6.0"></head><body><link rel="stylesheet" href="/wp-content/themes/default/style.css"></body></html>`,
			headers:       map[string]string{},
			expectedCMS:   []string{"WordPress"},
			expectedCount: 1,
		},
		{
			name:          "drupal detection",
			content:       `<html><head><meta name="generator" content="drupal 9.0"></head><body><script src="/sites/default/files/drupal.js"></script></body></html>`,
			headers:       map[string]string{},
			expectedCMS:   []string{"Drupal"},
			expectedCount: 1,
		},
		{
			name:          "joomla detection",
			content:       `<html><head><meta name="generator" content="joomla 4.0"></head><body><script src="/media/system/js/joomla.js"></script></body></html>`,
			headers:       map[string]string{},
			expectedCMS:   []string{"Joomla"},
			expectedCount: 1,
		},
		{
			name:          "shopify detection",
			content:       `<html><body>{{ product.title }}<script src="https://cdn.shopify.com/s/files/1/0000/0000/t/1/assets/theme.js"></script></body></html>`,
			headers:       map[string]string{},
			expectedCMS:   []string{"Shopify"},
			expectedCount: 1,
		},
		{
			name:          "wix detection",
			content:       `<html><body><script src="https://static.wixstatic.com/sites/all/themes/base/js/wix.js"></script></body></html>`,
			headers:       map[string]string{},
			expectedCMS:   []string{"Wix"},
			expectedCount: 1,
		},
		{
			name:          "squarespace detection",
			content:       `<html><body><script src="https://static.squarespace.com/universal/scripts-compressed/squarespace-common-compressed.js"></script></body></html>`,
			headers:       map[string]string{},
			expectedCMS:   []string{"Squarespace"},
			expectedCount: 1,
		},
		{
			name:          "no cms detected",
			content:       `<html><body>Plain HTML site</body></html>`,
			headers:       map[string]string{},
			expectedCMS:   []string{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			technologies := analyzer.detectCMS(tt.content, tt.headers)

			assert.Len(t, technologies, tt.expectedCount)

			for i, expectedCMS := range tt.expectedCMS {
				assert.Equal(t, expectedCMS, technologies[i].Name)
				assert.Equal(t, "cms", technologies[i].Category)
				assert.Greater(t, technologies[i].ConfidenceScore, 0.8)
			}
		})
	}
}

func TestTechnologyStackAnalyzer_DetectFrameworks(t *testing.T) {
	analyzer := NewTechnologyStackAnalyzer(zap.NewNop(), nil)
	require.NotNil(t, analyzer)

	tests := []struct {
		name               string
		content            string
		headers            map[string]string
		expectedFrameworks []string
		expectedCount      int
	}{
		{
			name:               "react detection",
			content:            `<html><head><script src="https://unpkg.com/react@18.2.0/umd/react.development.js"></script></head><body><div id="root"></div></body></html>`,
			headers:            map[string]string{},
			expectedFrameworks: []string{"React"},
			expectedCount:      1,
		},
		{
			name:               "vue detection",
			content:            `<html><head><script src="https://unpkg.com/vue@3.2.0/dist/vue.global.js"></script></head><body><div id="app"></div></body></html>`,
			headers:            map[string]string{},
			expectedFrameworks: []string{"Vue.js"},
			expectedCount:      1,
		},
		{
			name:               "angular detection",
			content:            `<html><body><div ng-app="myApp"><div ng-controller="myCtrl">{{ message }}</div></div></body></html>`,
			headers:            map[string]string{},
			expectedFrameworks: []string{"Angular"},
			expectedCount:      1,
		},
		{
			name:               "laravel detection",
			content:            `<html><head><meta name="csrf-token" content="abc123"></head><body>Laravel app</body></html>`,
			headers:            map[string]string{},
			expectedFrameworks: []string{"Laravel"},
			expectedCount:      1,
		},
		{
			name:               "django detection",
			content:            `<html><head><input type="hidden" name="csrfmiddlewaretoken" value="abc123"></head><body>Django app</body></html>`,
			headers:            map[string]string{},
			expectedFrameworks: []string{"Django"},
			expectedCount:      1,
		},
		{
			name:               "rails detection",
			content:            `<html><head><meta name="csrf-token" content="abc123"><meta name="authenticity_token" content="def456"></head><body>Rails app</body></html>`,
			headers:            map[string]string{},
			expectedFrameworks: []string{"Ruby on Rails"},
			expectedCount:      1,
		},
		{
			name:               "multiple frameworks",
			content:            `<html><head><script src="https://unpkg.com/react@18.2.0/umd/react.development.js"></script><script src="https://unpkg.com/vue@3.2.0/dist/vue.global.js"></script></head><body>Multiple frameworks</body></html>`,
			headers:            map[string]string{},
			expectedFrameworks: []string{"React", "Vue.js"},
			expectedCount:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			technologies := analyzer.detectFrameworks(tt.content, tt.headers)

			assert.GreaterOrEqual(t, len(technologies), len(tt.expectedFrameworks))

			// Check that expected frameworks are present
			for _, expectedFramework := range tt.expectedFrameworks {
				found := false
				for _, tech := range technologies {
					if tech.Name == expectedFramework {
						assert.Equal(t, "framework", tech.Category)
						assert.Greater(t, tech.ConfidenceScore, 0.7)
						found = true
						break
					}
				}
				assert.True(t, found, "Expected framework %s not found", expectedFramework)
			}
		})
	}
}

func TestTechnologyStackAnalyzer_DetectAnalytics(t *testing.T) {
	analyzer := NewTechnologyStackAnalyzer(zap.NewNop(), nil)
	require.NotNil(t, analyzer)

	tests := []struct {
		name              string
		content           string
		headers           map[string]string
		expectedAnalytics []string
		expectedCount     int
	}{
		{
			name:              "google analytics detection",
			content:           `<html><head><script>gtag('config', 'GA_MEASUREMENT_ID');</script></head><body>Site with GA</body></html>`,
			headers:           map[string]string{},
			expectedAnalytics: []string{"Google Analytics"},
			expectedCount:     1,
		},
		{
			name:              "google tag manager detection",
			content:           `<html><head><script src="https://www.googletagmanager.com/gtm.js?id=GTM-XXXX"></script></head><body>Site with GTM</body></html>`,
			headers:           map[string]string{},
			expectedAnalytics: []string{"Google Tag Manager"},
			expectedCount:     1,
		},
		{
			name:              "facebook pixel detection",
			content:           `<html><head><script>fbq('init', '123456789');</script></head><body>Site with Facebook Pixel</body></html>`,
			headers:           map[string]string{},
			expectedAnalytics: []string{"Facebook Pixel"},
			expectedCount:     1,
		},
		{
			name:              "hotjar detection",
			content:           `<html><head><script>_hjSettings={hjid:123456,hjsv:6};</script></head><body>Site with Hotjar</body></html>`,
			headers:           map[string]string{},
			expectedAnalytics: []string{"Hotjar"},
			expectedCount:     1,
		},
		{
			name:              "multiple analytics",
			content:           `<html><head><script>gtag('config', 'GA_MEASUREMENT_ID');</script><script>fbq('init', '123456789');</script></head><body>Site with multiple analytics</body></html>`,
			headers:           map[string]string{},
			expectedAnalytics: []string{"Google Analytics", "Facebook Pixel"},
			expectedCount:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			technologies := analyzer.detectAnalytics(tt.content, tt.headers)

			assert.GreaterOrEqual(t, len(technologies), len(tt.expectedAnalytics))

			// Check that expected analytics are present
			for _, expectedAnalytics := range tt.expectedAnalytics {
				found := false
				for _, tech := range technologies {
					if tech.Name == expectedAnalytics {
						assert.Equal(t, "analytics", tech.Category)
						assert.Greater(t, tech.ConfidenceScore, 0.8)
						found = true
						break
					}
				}
				assert.True(t, found, "Expected analytics %s not found", expectedAnalytics)
			}
		})
	}
}

func TestTechnologyStackAnalyzer_DetectHosting(t *testing.T) {
	analyzer := NewTechnologyStackAnalyzer(zap.NewNop(), nil)
	require.NotNil(t, analyzer)

	tests := []struct {
		name            string
		content         string
		headers         map[string]string
		expectedHosting []string
		expectedCount   int
	}{
		{
			name:    "aws detection",
			content: `<html><body>AWS hosted site</body></html>`,
			headers: map[string]string{
				"X-Amz-Cf-Id": "abc123",
				"Server":      "AmazonS3",
			},
			expectedHosting: []string{"Amazon Web Services"},
			expectedCount:   1,
		},
		{
			name:    "cloudflare detection",
			content: `<html><body>Cloudflare protected site</body></html>`,
			headers: map[string]string{
				"CF-Ray": "1234567890abcdef",
				"Server": "cloudflare",
			},
			expectedHosting: []string{"Cloudflare"},
			expectedCount:   1,
		},
		{
			name:    "vercel detection",
			content: `<html><body>Vercel deployed site</body></html>`,
			headers: map[string]string{
				"X-Vercel": "vercel",
			},
			expectedHosting: []string{"Vercel"},
			expectedCount:   1,
		},
		{
			name:    "netlify detection",
			content: `<html><body>Netlify deployed site</body></html>`,
			headers: map[string]string{
				"X-NF-Request-ID": "abc123",
			},
			expectedHosting: []string{"Netlify"},
			expectedCount:   1,
		},
		{
			name:    "multiple hosting",
			content: `<html><body>Complex hosting setup</body></html>`,
			headers: map[string]string{
				"X-Vercel": "vercel",
				"CF-Ray":   "1234567890abcdef",
			},
			expectedHosting: []string{"Vercel", "Cloudflare"},
			expectedCount:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			technologies := analyzer.detectHosting(tt.content, tt.headers)

			assert.GreaterOrEqual(t, len(technologies), len(tt.expectedHosting))

			// Check that expected hosting are present
			for _, expectedHosting := range tt.expectedHosting {
				found := false
				for _, tech := range technologies {
					if tech.Name == expectedHosting {
						assert.Equal(t, "hosting", tech.Category)
						assert.Greater(t, tech.ConfidenceScore, 0.8)
						found = true
						break
					}
				}
				assert.True(t, found, "Expected hosting %s not found", expectedHosting)
			}
		})
	}
}

func TestTechnologyStackAnalyzer_VersionExtraction(t *testing.T) {
	analyzer := NewTechnologyStackAnalyzer(zap.NewNop(), nil)
	require.NotNil(t, analyzer)

	tests := []struct {
		name            string
		content         string
		headers         map[string]string
		expectedCMS     string
		expectedVersion string
	}{
		{
			name:            "wordpress version extraction",
			content:         `<html><head><meta name="generator" content="wordpress 6.0.1"></head><body>WordPress site</body></html>`,
			headers:         map[string]string{},
			expectedCMS:     "WordPress",
			expectedVersion: "6.0.1",
		},
		{
			name:            "drupal version extraction",
			content:         `<html><head><meta name="generator" content="drupal 9.3.0"></head><body>Drupal site</body></html>`,
			headers:         map[string]string{},
			expectedCMS:     "Drupal",
			expectedVersion: "9.3.0",
		},
		{
			name:            "joomla version extraction",
			content:         `<html><head><meta name="generator" content="joomla 4.1.0"></head><body>Joomla site</body></html>`,
			headers:         map[string]string{},
			expectedCMS:     "Joomla",
			expectedVersion: "4.1.0",
		},
		{
			name:            "react version extraction",
			content:         `<html><head><script src="https://unpkg.com/react@18.2.0/umd/react.development.js"></script></head><body>React app</body></html>`,
			headers:         map[string]string{},
			expectedCMS:     "React",
			expectedVersion: "18.2.0",
		},
		{
			name:            "vue version extraction",
			content:         `<html><head><script src="https://unpkg.com/vue@3.2.45/dist/vue.global.js"></script></head><body>Vue app</body></html>`,
			headers:         map[string]string{},
			expectedCMS:     "Vue.js",
			expectedVersion: "3.2.45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var technologies []Technology

			switch tt.expectedCMS {
			case "WordPress":
				technologies = analyzer.detectCMS(tt.content, tt.headers)
			case "Drupal":
				technologies = analyzer.detectCMS(tt.content, tt.headers)
			case "Joomla":
				technologies = analyzer.detectCMS(tt.content, tt.headers)
			case "React":
				technologies = analyzer.detectFrameworks(tt.content, tt.headers)
			case "Vue.js":
				technologies = analyzer.detectFrameworks(tt.content, tt.headers)
			}

			require.Len(t, technologies, 1)
			assert.Equal(t, tt.expectedCMS, technologies[0].Name)
			assert.Equal(t, tt.expectedVersion, technologies[0].Version)
		})
	}
}

func TestTechnologyStackAnalyzer_StackTypeDetermination(t *testing.T) {
	analyzer := NewTechnologyStackAnalyzer(zap.NewNop(), nil)
	require.NotNil(t, analyzer)

	tests := []struct {
		name            string
		technologyStack *TechnologyStack
		expectedType    string
	}{
		{
			name: "modern stack",
			technologyStack: &TechnologyStack{
				Frameworks: []Technology{
					{Name: "React", Category: "framework", Type: "frontend"},
				},
			},
			expectedType: "modern",
		},
		{
			name: "traditional cms",
			technologyStack: &TechnologyStack{
				CMS: []Technology{
					{Name: "WordPress", Category: "cms", Type: "fullstack"},
				},
			},
			expectedType: "traditional",
		},
		{
			name: "headless cms",
			technologyStack: &TechnologyStack{
				CMS: []Technology{
					{Name: "WordPress", Category: "cms", Type: "fullstack"},
				},
				Frameworks: []Technology{
					{Name: "React", Category: "framework", Type: "frontend"},
				},
			},
			expectedType: "modern",
		},
		{
			name: "static site",
			technologyStack: &TechnologyStack{
				Tools: []Technology{
					{Name: "Vite", Category: "tool", Type: "build"},
				},
			},
			expectedType: "static",
		},
		{
			name:            "unknown stack",
			technologyStack: &TechnologyStack{},
			expectedType:    "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stackType := analyzer.determineStackType(tt.technologyStack)
			assert.Equal(t, tt.expectedType, stackType)
		})
	}
}

func TestTechnologyStackAnalyzer_ComplexityDetermination(t *testing.T) {
	analyzer := NewTechnologyStackAnalyzer(zap.NewNop(), nil)
	require.NotNil(t, analyzer)

	tests := []struct {
		name               string
		technologyStack    *TechnologyStack
		expectedComplexity string
	}{
		{
			name: "simple stack",
			technologyStack: &TechnologyStack{
				CMS: []Technology{{Name: "WordPress"}},
			},
			expectedComplexity: "simple",
		},
		{
			name: "moderate stack",
			technologyStack: &TechnologyStack{
				CMS:        []Technology{{Name: "WordPress"}},
				Frameworks: []Technology{{Name: "React"}},
				Analytics:  []Technology{{Name: "Google Analytics"}},
				Hosting:    []Technology{{Name: "Vercel"}},
			},
			expectedComplexity: "moderate",
		},
		{
			name: "complex stack",
			technologyStack: &TechnologyStack{
				CMS:        []Technology{{Name: "WordPress"}},
				Frameworks: []Technology{{Name: "React"}, {Name: "Vue.js"}},
				Tools:      []Technology{{Name: "Webpack"}, {Name: "Vite"}},
				Analytics:  []Technology{{Name: "Google Analytics"}, {Name: "Facebook Pixel"}},
				Hosting:    []Technology{{Name: "Vercel"}, {Name: "Cloudflare"}},
				Database:   []Technology{{Name: "MySQL"}},
				Security:   []Technology{{Name: "AWS WAF"}},
			},
			expectedComplexity: "complex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complexity := analyzer.determineComplexity(tt.technologyStack)
			assert.Equal(t, tt.expectedComplexity, complexity)
		})
	}
}

func TestTechnologyStackAnalyzer_ConfidenceCalculation(t *testing.T) {
	analyzer := NewTechnologyStackAnalyzer(zap.NewNop(), nil)
	require.NotNil(t, analyzer)

	tests := []struct {
		name               string
		technologyStack    *TechnologyStack
		expectedConfidence float64
	}{
		{
			name:               "empty stack",
			technologyStack:    &TechnologyStack{},
			expectedConfidence: 0.0,
		},
		{
			name: "single technology",
			technologyStack: &TechnologyStack{
				CMS: []Technology{{ConfidenceScore: 0.95}},
			},
			expectedConfidence: 0.95,
		},
		{
			name: "multiple technologies",
			technologyStack: &TechnologyStack{
				CMS:        []Technology{{ConfidenceScore: 0.95}},
				Frameworks: []Technology{{ConfidenceScore: 0.85}},
				Analytics:  []Technology{{ConfidenceScore: 0.90}},
			},
			expectedConfidence: 0.90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := analyzer.calculateOverallConfidence(tt.technologyStack)
			assert.InDelta(t, tt.expectedConfidence, confidence, 0.001)
		})
	}
}

func TestTechnologyStackAnalyzer_GetDefaultConfig(t *testing.T) {
	config := getDefaultTechnologyStackAnalyzerConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 0.3, config.MinConfidenceScore)
	assert.Equal(t, 10, config.MaxTechnologiesPerCategory)

	// Check that all indicator maps are populated
	assert.NotEmpty(t, config.CMSIndicators)
	assert.NotEmpty(t, config.FrameworkIndicators)
	assert.NotEmpty(t, config.ToolIndicators)
	assert.NotEmpty(t, config.AnalyticsIndicators)
	assert.NotEmpty(t, config.HostingIndicators)

	// Check specific indicators
	assert.Contains(t, config.CMSIndicators, "wordpress")
	assert.Contains(t, config.FrameworkIndicators, "react")
	assert.Contains(t, config.AnalyticsIndicators, "google_analytics")
}
