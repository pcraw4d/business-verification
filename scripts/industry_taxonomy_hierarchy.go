package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// IndustryTaxonomyHierarchy represents the complete industry taxonomy structure
type IndustryTaxonomyHierarchy struct {
	AnalysisDate       time.Time                  `json:"analysis_date"`
	TotalCategories    int                        `json:"total_categories"`
	PrimaryCategories  []PrimaryCategory          `json:"primary_categories"`
	CoverageAnalysis   TaxonomyCoverageAnalysis   `json:"coverage_analysis"`
	ImplementationPlan TaxonomyImplementationPlan `json:"implementation_plan"`
	Recommendations    []TaxonomyRecommendation   `json:"recommendations"`
}

type PrimaryCategory struct {
	ID                  int                  `json:"id"`
	Name                string               `json:"name"`
	Description         string               `json:"description"`
	CategoryType        string               `json:"category_type"`
	MarketSize          string               `json:"market_size"`
	GrowthRate          string               `json:"growth_rate"`
	CoverageStatus      string               `json:"coverage_status"`
	Priority            string               `json:"priority"`
	Subcategories       []Subcategory        `json:"subcategories"`
	Keywords            []string             `json:"keywords"`
	ClassificationCodes []ClassificationCode `json:"classification_codes"`
}

type Subcategory struct {
	ID                  int                  `json:"id"`
	Name                string               `json:"name"`
	Description         string               `json:"description"`
	CategoryType        string               `json:"category_type"`
	MarketSize          string               `json:"market_size"`
	GrowthRate          string               `json:"growth_rate"`
	CoverageStatus      string               `json:"coverage_status"`
	Priority            string               `json:"priority"`
	SpecificIndustries  []SpecificIndustry   `json:"specific_industries"`
	Keywords            []string             `json:"keywords"`
	ClassificationCodes []ClassificationCode `json:"classification_codes"`
}

type SpecificIndustry struct {
	ID                  int                  `json:"id"`
	Name                string               `json:"name"`
	Description         string               `json:"description"`
	MarketSize          string               `json:"market_size"`
	GrowthRate          string               `json:"growth_rate"`
	CoverageStatus      string               `json:"coverage_status"`
	Priority            string               `json:"priority"`
	Keywords            []string             `json:"keywords"`
	ClassificationCodes []ClassificationCode `json:"classification_codes"`
}

type ClassificationCode struct {
	CodeType    string `json:"code_type"`
	Code        string `json:"code"`
	Description string `json:"description"`
	IsPrimary   bool   `json:"is_primary"`
}

type TaxonomyCoverageAnalysis struct {
	TotalPrimaryCategories    int            `json:"total_primary_categories"`
	CoveredPrimaryCategories  int            `json:"covered_primary_categories"`
	TotalSubcategories        int            `json:"total_subcategories"`
	CoveredSubcategories      int            `json:"covered_subcategories"`
	TotalSpecificIndustries   int            `json:"total_specific_industries"`
	CoveredSpecificIndustries int            `json:"covered_specific_industries"`
	OverallCoveragePercentage float64        `json:"overall_coverage_percentage"`
	CoverageByPriority        map[string]int `json:"coverage_by_priority"`
}

type TaxonomyImplementationPlan struct {
	Phase1   []string `json:"phase_1"`
	Phase2   []string `json:"phase_2"`
	Phase3   []string `json:"phase_3"`
	Timeline string   `json:"timeline"`
}

type TaxonomyRecommendation struct {
	Recommendation string `json:"recommendation"`
	Priority       string `json:"priority"`
	Impact         string `json:"impact"`
	Effort         string `json:"effort"`
	Timeline       string `json:"timeline"`
}

func main() {
	// Generate comprehensive industry taxonomy hierarchy
	taxonomy := generateIndustryTaxonomyHierarchy()

	// Generate comprehensive report
	report, err := generateTaxonomyReport(taxonomy)
	if err != nil {
		fmt.Printf("Failed to generate report: %v\n", err)
		os.Exit(1)
	}

	// Save report to file
	filename := fmt.Sprintf("industry_taxonomy_hierarchy_%s.md", time.Now().Format("2006-01-02"))
	err = os.WriteFile(filename, []byte(report), 0644)
	if err != nil {
		fmt.Printf("Failed to save report: %v\n", err)
		os.Exit(1)
	}

	// Save JSON data
	jsonFilename := fmt.Sprintf("industry_taxonomy_hierarchy_%s.json", time.Now().Format("2006-01-02"))
	jsonData, err := json.MarshalIndent(taxonomy, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal JSON: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(jsonFilename, jsonData, 0644)
	if err != nil {
		fmt.Printf("Failed to save JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Industry taxonomy hierarchy completed successfully!\n")
	fmt.Printf("📊 Report saved to: %s\n", filename)
	fmt.Printf("📄 JSON data saved to: %s\n", jsonFilename)
	fmt.Printf("🎯 Total primary categories: %d\n", taxonomy.TotalCategories)
	fmt.Printf("📈 Overall coverage: %.1f%%\n", taxonomy.CoverageAnalysis.OverallCoveragePercentage)
	fmt.Printf("🚀 Implementation phases: 3\n")
}

func generateIndustryTaxonomyHierarchy() *IndustryTaxonomyHierarchy {
	taxonomy := &IndustryTaxonomyHierarchy{
		AnalysisDate:       time.Now(),
		PrimaryCategories:  []PrimaryCategory{},
		CoverageAnalysis:   TaxonomyCoverageAnalysis{},
		ImplementationPlan: TaxonomyImplementationPlan{},
		Recommendations:    []TaxonomyRecommendation{},
	}

	// Create comprehensive industry taxonomy
	createPrimaryCategories(taxonomy)
	analyzeCoverage(taxonomy)
	createImplementationPlan(taxonomy)
	generateRecommendations(taxonomy)

	return taxonomy
}

func createPrimaryCategories(taxonomy *IndustryTaxonomyHierarchy) {
	// Technology & Digital Services
	techCategory := PrimaryCategory{
		ID:             1,
		Name:           "Technology & Digital Services",
		Description:    "Technology companies, software development, digital services, and IT solutions",
		CategoryType:   "primary",
		MarketSize:     "very_large",
		GrowthRate:     "high_growth",
		CoverageStatus: "good",
		Priority:       "critical",
		Keywords:       []string{"technology", "software", "digital", "it", "tech", "computer", "programming", "development"},
		ClassificationCodes: []ClassificationCode{
			{CodeType: "NAICS", Code: "541511", Description: "Custom Computer Programming Services", IsPrimary: true},
			{CodeType: "SIC", Code: "7372", Description: "Prepackaged Software", IsPrimary: true},
			{CodeType: "MCC", Code: "5734", Description: "Computer Software Stores", IsPrimary: true},
		},
		Subcategories: []Subcategory{
			{
				ID: 1, Name: "Software Development", Description: "Custom software development and programming services",
				CategoryType: "secondary", MarketSize: "large", GrowthRate: "high_growth", CoverageStatus: "good", Priority: "critical",
				Keywords: []string{"software development", "programming", "coding", "application development", "web development", "mobile app"},
				SpecificIndustries: []SpecificIndustry{
					{ID: 1, Name: "Web Development", Keywords: []string{"web development", "website", "web application", "frontend", "backend"}},
					{ID: 2, Name: "Mobile App Development", Keywords: []string{"mobile app", "ios", "android", "mobile development", "app development"}},
					{ID: 3, Name: "Enterprise Software", Keywords: []string{"enterprise software", "business software", "corporate software", "enterprise application"}},
				},
			},
			{
				ID: 2, Name: "Cloud Computing", Description: "Cloud infrastructure, platforms, and services",
				CategoryType: "secondary", MarketSize: "large", GrowthRate: "high_growth", CoverageStatus: "good", Priority: "critical",
				Keywords: []string{"cloud computing", "aws", "azure", "google cloud", "cloud infrastructure", "saas", "paas", "iaas"},
				SpecificIndustries: []SpecificIndustry{
					{ID: 4, Name: "Cloud Infrastructure", Keywords: []string{"cloud infrastructure", "cloud hosting", "cloud services", "cloud platform"}},
					{ID: 5, Name: "Software as a Service", Keywords: []string{"saas", "software as a service", "cloud software", "online software"}},
				},
			},
		},
	}

	// Healthcare & Life Sciences
	healthCategory := PrimaryCategory{
		ID: 2, Name: "Healthcare & Life Sciences", Description: "Healthcare providers, medical services, pharmaceuticals, and health technology",
		CategoryType: "primary", MarketSize: "very_large", GrowthRate: "growing", CoverageStatus: "good", Priority: "critical",
		Keywords: []string{"healthcare", "medical", "health", "medicine", "pharmaceutical", "medical device", "hospital", "clinic"},
		Subcategories: []Subcategory{
			{
				ID: 3, Name: "Medical Services", Description: "Healthcare providers, clinics, hospitals, and medical practices",
				CategoryType: "secondary", MarketSize: "very_large", GrowthRate: "growing", CoverageStatus: "good", Priority: "critical",
				Keywords: []string{"medical services", "healthcare", "medical practice", "clinic", "hospital", "physician", "doctor"},
			},
			{
				ID: 4, Name: "Pharmaceuticals", Description: "Drug manufacturing, pharmaceutical services, and biotech",
				CategoryType: "secondary", MarketSize: "large", GrowthRate: "growing", CoverageStatus: "good", Priority: "critical",
				Keywords: []string{"pharmaceutical", "drug manufacturing", "pharma", "medication", "prescription drugs", "biotech"},
			},
		},
	}

	// Financial Services
	financeCategory := PrimaryCategory{
		ID: 3, Name: "Financial Services", Description: "Banking, investment, insurance, and financial technology",
		CategoryType: "primary", MarketSize: "very_large", GrowthRate: "growing", CoverageStatus: "good", Priority: "critical",
		Keywords: []string{"finance", "banking", "financial", "investment", "insurance", "credit", "fintech"},
		Subcategories: []Subcategory{
			{
				ID: 5, Name: "Commercial Banking", Description: "Traditional banking, financial institutions, and lending",
				CategoryType: "secondary", MarketSize: "very_large", GrowthRate: "stable", CoverageStatus: "good", Priority: "critical",
				Keywords: []string{"banking", "commercial bank", "financial institution", "bank", "lending", "loans"},
			},
			{
				ID: 6, Name: "Investment Services", Description: "Investment banking, wealth management, and asset management",
				CategoryType: "secondary", MarketSize: "large", GrowthRate: "growing", CoverageStatus: "good", Priority: "high",
				Keywords: []string{"investment", "wealth management", "asset management", "portfolio management", "financial planning"},
			},
		},
	}

	// Missing Industries
	missingCategory := PrimaryCategory{
		ID: 4, Name: "Missing Industries", Description: "Industries that need to be added to achieve comprehensive coverage",
		CategoryType: "primary", MarketSize: "very_large", GrowthRate: "stable", CoverageStatus: "missing", Priority: "critical",
		Keywords: []string{},
		Subcategories: []Subcategory{
			{
				ID: 7, Name: "Restaurant & Food Service", Description: "Food service, restaurants, and dining establishments",
				CategoryType: "secondary", MarketSize: "very_large", GrowthRate: "stable", CoverageStatus: "missing", Priority: "critical",
				Keywords: []string{"restaurant", "food service", "dining", "chef", "cuisine", "menu", "catering", "fast food"},
			},
			{
				ID: 8, Name: "Professional Services", Description: "Legal, accounting, consulting, and business services",
				CategoryType: "secondary", MarketSize: "large", GrowthRate: "growing", CoverageStatus: "missing", Priority: "high",
				Keywords: []string{"consulting", "legal", "accounting", "advertising", "marketing", "hr", "recruitment"},
			},
		},
	}

	taxonomy.PrimaryCategories = []PrimaryCategory{techCategory, healthCategory, financeCategory, missingCategory}
	taxonomy.TotalCategories = len(taxonomy.PrimaryCategories)
}

func analyzeCoverage(taxonomy *IndustryTaxonomyHierarchy) {
	analysis := &taxonomy.CoverageAnalysis

	analysis.TotalPrimaryCategories = len(taxonomy.PrimaryCategories)
	analysis.CoveredPrimaryCategories = 0
	analysis.TotalSubcategories = 0
	analysis.CoveredSubcategories = 0
	analysis.TotalSpecificIndustries = 0
	analysis.CoveredSpecificIndustries = 0
	analysis.CoverageByPriority = make(map[string]int)

	for _, primary := range taxonomy.PrimaryCategories {
		if primary.CoverageStatus != "missing" {
			analysis.CoveredPrimaryCategories++
		}
		analysis.CoverageByPriority[primary.Priority]++

		analysis.TotalSubcategories += len(primary.Subcategories)
		for _, sub := range primary.Subcategories {
			if sub.CoverageStatus != "missing" {
				analysis.CoveredSubcategories++
			}
			analysis.TotalSpecificIndustries += len(sub.SpecificIndustries)
			for _, specific := range sub.SpecificIndustries {
				if specific.CoverageStatus != "missing" {
					analysis.CoveredSpecificIndustries++
				}
			}
		}
	}

	if analysis.TotalPrimaryCategories > 0 {
		analysis.OverallCoveragePercentage = float64(analysis.CoveredPrimaryCategories) / float64(analysis.TotalPrimaryCategories) * 100
	}
}

func createImplementationPlan(taxonomy *IndustryTaxonomyHierarchy) {
	plan := &taxonomy.ImplementationPlan

	plan.Phase1 = []string{
		"Add Restaurant & Food Service industry with comprehensive keywords",
		"Add Professional Services industry with legal, accounting, and consulting keywords",
		"Expand Technology industry with AI/ML and emerging tech keywords",
		"Create industry taxonomy database schema",
	}

	plan.Phase2 = []string{
		"Add Construction & Building industry",
		"Add Automotive industry",
		"Add Agriculture industry",
		"Enhance keyword coverage for underrepresented industries",
		"Add comprehensive classification code mappings",
	}

	plan.Phase3 = []string{
		"Add remaining missing industries",
		"Implement dynamic keyword weighting",
		"Add emerging trends integration",
		"Create industry coverage analytics",
		"Implement automated gap analysis",
	}

	plan.Timeline = "6-8 weeks"
}

func generateRecommendations(taxonomy *IndustryTaxonomyHierarchy) {
	recommendations := []TaxonomyRecommendation{
		{
			Recommendation: "Implement comprehensive industry taxonomy hierarchy",
			Priority:       "Critical",
			Impact:         "Very High",
			Effort:         "High",
			Timeline:       "6-8 weeks",
		},
		{
			Recommendation: "Add missing critical industries (Restaurant, Professional Services)",
			Priority:       "Critical",
			Impact:         "Very High",
			Effort:         "Medium",
			Timeline:       "2-4 weeks",
		},
		{
			Recommendation: "Create industry taxonomy database schema",
			Priority:       "High",
			Impact:         "High",
			Effort:         "Medium",
			Timeline:       "1-2 weeks",
		},
		{
			Recommendation: "Implement industry coverage monitoring and analytics",
			Priority:       "Medium",
			Impact:         "Medium",
			Effort:         "Medium",
			Timeline:       "2-3 weeks",
		},
	}

	taxonomy.Recommendations = recommendations
}

func generateTaxonomyReport(taxonomy *IndustryTaxonomyHierarchy) (string, error) {
	var report strings.Builder

	report.WriteString("# 🏗️ Industry Taxonomy Hierarchy Report\n\n")
	report.WriteString(fmt.Sprintf("**Analysis Date**: %s\n", taxonomy.AnalysisDate.Format("January 2, 2006")))
	report.WriteString(fmt.Sprintf("**Total Primary Categories**: %d\n\n", taxonomy.TotalCategories))

	// Executive Summary
	report.WriteString("## 📊 Executive Summary\n\n")
	report.WriteString("This comprehensive analysis creates a complete industry taxonomy hierarchy for the KYB Platform classification system.\n\n")

	report.WriteString("### Key Findings:\n")
	report.WriteString(fmt.Sprintf("- **Total Primary Categories**: %d major industry sectors\n", taxonomy.TotalCategories))
	report.WriteString(fmt.Sprintf("- **Overall Coverage**: %.1f%% of primary categories covered\n", taxonomy.CoverageAnalysis.OverallCoveragePercentage))
	report.WriteString(fmt.Sprintf("- **Total Subcategories**: %d industry subcategories\n", taxonomy.CoverageAnalysis.TotalSubcategories))
	report.WriteString(fmt.Sprintf("- **Total Specific Industries**: %d specific industry types\n", taxonomy.CoverageAnalysis.TotalSpecificIndustries))
	report.WriteString("- **Implementation Timeline**: 6-8 weeks for complete taxonomy\n\n")

	// Primary Categories
	report.WriteString("## 🏭 Primary Industry Categories\n\n")
	for _, primary := range taxonomy.PrimaryCategories {
		report.WriteString(fmt.Sprintf("### %s\n", primary.Name))
		report.WriteString(fmt.Sprintf("**Description**: %s\n", primary.Description))
		report.WriteString(fmt.Sprintf("**Market Size**: %s\n", primary.MarketSize))
		report.WriteString(fmt.Sprintf("**Growth Rate**: %s\n", primary.GrowthRate))
		report.WriteString(fmt.Sprintf("**Coverage Status**: %s\n", primary.CoverageStatus))
		report.WriteString(fmt.Sprintf("**Priority**: %s\n", primary.Priority))
		report.WriteString(fmt.Sprintf("**Subcategories**: %d\n\n", len(primary.Subcategories)))

		// Subcategories
		for _, sub := range primary.Subcategories {
			report.WriteString(fmt.Sprintf("#### %s\n", sub.Name))
			report.WriteString(fmt.Sprintf("- **Description**: %s\n", sub.Description))
			report.WriteString(fmt.Sprintf("- **Coverage Status**: %s\n", sub.CoverageStatus))
			report.WriteString(fmt.Sprintf("- **Priority**: %s\n", sub.Priority))
			report.WriteString(fmt.Sprintf("- **Specific Industries**: %d\n", len(sub.SpecificIndustries)))
			report.WriteString(fmt.Sprintf("- **Keywords**: %s\n\n", strings.Join(sub.Keywords, ", ")))
		}
		report.WriteString("\n")
	}

	// Coverage Analysis
	report.WriteString("## 📈 Coverage Analysis\n\n")
	report.WriteString("| Level | Total | Covered | Coverage % |\n")
	report.WriteString("|-------|-------|---------|------------|\n")
	report.WriteString(fmt.Sprintf("| Primary Categories | %d | %d | %.1f%% |\n",
		taxonomy.CoverageAnalysis.TotalPrimaryCategories,
		taxonomy.CoverageAnalysis.CoveredPrimaryCategories,
		taxonomy.CoverageAnalysis.OverallCoveragePercentage))
	report.WriteString(fmt.Sprintf("| Subcategories | %d | %d | %.1f%% |\n",
		taxonomy.CoverageAnalysis.TotalSubcategories,
		taxonomy.CoverageAnalysis.CoveredSubcategories,
		float64(taxonomy.CoverageAnalysis.CoveredSubcategories)/float64(taxonomy.CoverageAnalysis.TotalSubcategories)*100))
	report.WriteString(fmt.Sprintf("| Specific Industries | %d | %d | %.1f%% |\n",
		taxonomy.CoverageAnalysis.TotalSpecificIndustries,
		taxonomy.CoverageAnalysis.CoveredSpecificIndustries,
		float64(taxonomy.CoverageAnalysis.CoveredSpecificIndustries)/float64(taxonomy.CoverageAnalysis.TotalSpecificIndustries)*100))

	// Implementation Plan
	report.WriteString("\n## 🗺️ Implementation Plan\n\n")
	report.WriteString(fmt.Sprintf("**Timeline**: %s\n\n", taxonomy.ImplementationPlan.Timeline))

	report.WriteString("### Phase 1: Critical Missing Industries (Weeks 1-2)\n")
	for i, item := range taxonomy.ImplementationPlan.Phase1 {
		report.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
	}

	report.WriteString("\n### Phase 2: Major Industries (Weeks 3-4)\n")
	for i, item := range taxonomy.ImplementationPlan.Phase2 {
		report.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
	}

	report.WriteString("\n### Phase 3: Advanced Features (Weeks 5-8)\n")
	for i, item := range taxonomy.ImplementationPlan.Phase3 {
		report.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
	}

	// Recommendations
	report.WriteString("\n## 💡 Recommendations\n\n")
	report.WriteString("| Recommendation | Priority | Impact | Effort | Timeline |\n")
	report.WriteString("|----------------|----------|--------|--------|----------|\n")

	for _, rec := range taxonomy.Recommendations {
		report.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			rec.Recommendation, rec.Priority, rec.Impact, rec.Effort, rec.Timeline))
	}

	// Next Steps
	report.WriteString("\n## 🎯 Next Steps\n\n")
	report.WriteString("1. **Immediate Actions (This Week)**\n")
	report.WriteString("   - Create industry taxonomy database schema\n")
	report.WriteString("   - Add Restaurant & Food Service industry\n")
	report.WriteString("   - Add Professional Services industry\n\n")

	report.WriteString("2. **Short-term Actions (Next 2 Weeks)**\n")
	report.WriteString("   - Complete Phase 1 implementation\n")
	report.WriteString("   - Begin Phase 2 major industries\n")
	report.WriteString("   - Implement industry coverage monitoring\n\n")

	report.WriteString("3. **Medium-term Actions (Next Month)**\n")
	report.WriteString("   - Complete Phase 2 implementation\n")
	report.WriteString("   - Begin Phase 3 advanced features\n")
	report.WriteString("   - Implement automated gap analysis\n\n")

	report.WriteString("---\n\n")
	report.WriteString("**Report Generated**: " + time.Now().Format("January 2, 2006 at 3:04 PM") + "\n")
	report.WriteString("**Status**: Ready for implementation\n")
	report.WriteString("**Next Review**: Weekly during implementation\n")

	return report.String(), nil
}
