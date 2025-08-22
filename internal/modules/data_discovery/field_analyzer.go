package data_discovery

import (
	"context"
	"sort"
	"strings"

	"go.uber.org/zap"
)

// FieldAnalyzer analyzes content to discover potential data fields
type FieldAnalyzer struct {
	config     *DataDiscoveryConfig
	logger     *zap.Logger
	fieldTypes []FieldTypeDefinition
	dataTypes  []DataTypeDefinition
	validators map[string]FieldValidator
}

// FieldTypeDefinition defines a type of field that can be discovered
type FieldTypeDefinition struct {
	FieldType       string   `json:"field_type"`
	DisplayName     string   `json:"display_name"`
	Description     string   `json:"description"`
	DataType        string   `json:"data_type"`
	Priority        int      `json:"priority"`
	BusinessValue   float64  `json:"business_value"`
	ContextKeywords []string `json:"context_keywords"`
	ValidationRules []string `json:"validation_rules"`
	Examples        []string `json:"examples"`
}

// DataTypeDefinition defines data type characteristics
type DataTypeDefinition struct {
	DataType        string   `json:"data_type"`
	DisplayName     string   `json:"display_name"`
	ValidationRules []string `json:"validation_rules"`
	FormatPatterns  []string `json:"format_patterns"`
	Examples        []string `json:"examples"`
}

// FieldValidator interface for field validation
type FieldValidator interface {
	ValidateField(value string) (bool, float64)
	GetDataType() string
	GetConfidenceBoost() float64
}

// NewFieldAnalyzer creates a new field analyzer
func NewFieldAnalyzer(config *DataDiscoveryConfig, logger *zap.Logger) *FieldAnalyzer {
	analyzer := &FieldAnalyzer{
		config:     config,
		logger:     logger,
		fieldTypes: getFieldTypeDefinitions(),
		dataTypes:  getDataTypeDefinitions(),
		validators: make(map[string]FieldValidator),
	}

	// Initialize validators
	analyzer.initializeValidators()

	return analyzer
}

// AnalyzeFields analyzes content to discover fields based on patterns and classification
func (fa *FieldAnalyzer) AnalyzeFields(ctx context.Context, content *ContentInput, patterns []PatternMatch, classification *ClassificationResult) ([]DiscoveredField, error) {
	var discoveredFields []DiscoveredField

	fa.logger.Debug("Starting field analysis",
		zap.Int("patterns_count", len(patterns)),
		zap.String("business_type", getStringValue(classification, "BusinessType")))

	// Group patterns by field type
	patternGroups := fa.groupPatternsByFieldType(patterns)

	// Analyze each field type
	for fieldType, fieldPatterns := range patternGroups {
		field := fa.analyzeFieldType(content, fieldType, fieldPatterns, classification)
		if field != nil {
			discoveredFields = append(discoveredFields, *field)
		}
	}

	// Discover additional fields based on business context
	contextFields := fa.discoverContextualFields(content, classification)
	discoveredFields = append(discoveredFields, contextFields...)

	// Enhance fields with business value and priority
	fa.enhanceFieldsWithBusinessIntelligence(discoveredFields, classification)

	// Sort fields by priority and business value
	fa.sortFieldsByImportance(discoveredFields)

	fa.logger.Debug("Field analysis completed",
		zap.Int("fields_discovered", len(discoveredFields)))

	return discoveredFields, nil
}

// groupPatternsByFieldType groups pattern matches by field type
func (fa *FieldAnalyzer) groupPatternsByFieldType(patterns []PatternMatch) map[string][]PatternMatch {
	groups := make(map[string][]PatternMatch)

	for _, pattern := range patterns {
		groups[pattern.FieldType] = append(groups[pattern.FieldType], pattern)
	}

	return groups
}

// analyzeFieldType analyzes a specific field type based on pattern matches
func (fa *FieldAnalyzer) analyzeFieldType(content *ContentInput, fieldType string, patterns []PatternMatch, classification *ClassificationResult) *DiscoveredField {
	if len(patterns) == 0 {
		return nil
	}

	// Find field type definition
	var fieldTypeDef *FieldTypeDefinition
	for _, def := range fa.fieldTypes {
		if def.FieldType == fieldType {
			fieldTypeDef = &def
			break
		}
	}

	if fieldTypeDef == nil {
		fa.logger.Warn("Unknown field type", zap.String("field_type", fieldType))
		return nil
	}

	// Extract sample values from patterns
	sampleValues := fa.extractSampleValues(patterns)

	// Calculate confidence score
	confidence := fa.calculateFieldConfidence(patterns, fieldTypeDef, classification)

	// Determine extraction method
	extractionMethod := fa.determineExtractionMethod(patterns, fieldTypeDef)

	// Generate validation rules
	validationRules := fa.generateValidationRules(fieldTypeDef, sampleValues)

	field := &DiscoveredField{
		FieldName:        fa.generateFieldName(fieldType, classification),
		FieldType:        fieldType,
		DataType:         fieldTypeDef.DataType,
		ConfidenceScore:  confidence,
		ExtractionMethod: extractionMethod,
		SampleValues:     sampleValues,
		ValidationRules:  validationRules,
		Priority:         fieldTypeDef.Priority,
		BusinessValue:    fieldTypeDef.BusinessValue,
		Metadata: map[string]interface{}{
			"pattern_count":   len(patterns),
			"field_type_def":  fieldTypeDef,
			"extraction_base": "pattern_matching",
		},
	}

	return field
}

// discoverContextualFields discovers additional fields based on business context
func (fa *FieldAnalyzer) discoverContextualFields(content *ContentInput, classification *ClassificationResult) []DiscoveredField {
	var contextFields []DiscoveredField

	if classification == nil {
		return contextFields
	}

	// Discover fields based on business type
	businessTypeFields := fa.discoverBusinessTypeFields(content, classification.BusinessType)
	contextFields = append(contextFields, businessTypeFields...)

	// Discover fields based on industry
	industryFields := fa.discoverIndustryFields(content, classification.IndustryCategory)
	contextFields = append(contextFields, industryFields...)

	// Discover fields based on content categories
	for _, category := range classification.ContentCategories {
		categoryFields := fa.discoverCategoryFields(content, category)
		contextFields = append(contextFields, categoryFields...)
	}

	return contextFields
}

// discoverBusinessTypeFields discovers fields specific to business type
func (fa *FieldAnalyzer) discoverBusinessTypeFields(content *ContentInput, businessType string) []DiscoveredField {
	var fields []DiscoveredField

	switch businessType {
	case "ecommerce":
		fields = append(fields, fa.discoverEcommerceFields(content)...)
	case "b2b_software":
		fields = append(fields, fa.discoverB2BFields(content)...)
	case "consulting":
		fields = append(fields, fa.discoverConsultingFields(content)...)
	case "healthcare":
		fields = append(fields, fa.discoverHealthcareFields(content)...)
	case "financial":
		fields = append(fields, fa.discoverFinancialFields(content)...)
	}

	return fields
}

// discoverIndustryFields discovers fields specific to industry
func (fa *FieldAnalyzer) discoverIndustryFields(content *ContentInput, industry string) []DiscoveredField {
	var fields []DiscoveredField

	switch industry {
	case "technology":
		fields = append(fields, fa.discoverTechnologyFields(content)...)
	case "manufacturing":
		fields = append(fields, fa.discoverManufacturingFields(content)...)
	case "retail":
		fields = append(fields, fa.discoverRetailFields(content)...)
	case "education":
		fields = append(fields, fa.discoverEducationFields(content)...)
	}

	return fields
}

// discoverCategoryFields discovers fields based on content category
func (fa *FieldAnalyzer) discoverCategoryFields(content *ContentInput, category string) []DiscoveredField {
	var fields []DiscoveredField

	switch category {
	case "corporate":
		fields = append(fields, fa.discoverCorporateFields(content)...)
	case "product":
		fields = append(fields, fa.discoverProductFields(content)...)
	case "service":
		fields = append(fields, fa.discoverServiceFields(content)...)
	}

	return fields
}

// Business type specific field discovery methods

func (fa *FieldAnalyzer) discoverEcommerceFields(content *ContentInput) []DiscoveredField {
	return []DiscoveredField{
		fa.createContextualField("payment_methods", "payment_methods", "array", 3, 0.8, "Payment methods accepted"),
		fa.createContextualField("shipping_options", "shipping_info", "array", 4, 0.7, "Shipping options"),
		fa.createContextualField("return_policy", "policy_text", "string", 5, 0.6, "Return policy information"),
	}
}

func (fa *FieldAnalyzer) discoverB2BFields(content *ContentInput) []DiscoveredField {
	return []DiscoveredField{
		fa.createContextualField("api_documentation", "url", "string", 2, 0.9, "API documentation link"),
		fa.createContextualField("integration_options", "integration_info", "array", 3, 0.8, "Integration capabilities"),
		fa.createContextualField("enterprise_features", "feature_list", "array", 4, 0.7, "Enterprise features"),
	}
}

func (fa *FieldAnalyzer) discoverConsultingFields(content *ContentInput) []DiscoveredField {
	return []DiscoveredField{
		fa.createContextualField("expertise_areas", "expertise", "array", 2, 0.9, "Areas of expertise"),
		fa.createContextualField("case_studies", "case_study_links", "array", 3, 0.8, "Case studies"),
		fa.createContextualField("certifications", "certification_list", "array", 4, 0.7, "Professional certifications"),
	}
}

func (fa *FieldAnalyzer) discoverHealthcareFields(content *ContentInput) []DiscoveredField {
	return []DiscoveredField{
		fa.createContextualField("medical_specialties", "specialty_list", "array", 2, 0.9, "Medical specialties"),
		fa.createContextualField("insurance_accepted", "insurance_list", "array", 3, 0.8, "Insurance plans accepted"),
		fa.createContextualField("appointment_booking", "booking_info", "string", 4, 0.7, "Appointment booking information"),
	}
}

func (fa *FieldAnalyzer) discoverFinancialFields(content *ContentInput) []DiscoveredField {
	return []DiscoveredField{
		fa.createContextualField("financial_products", "product_list", "array", 2, 0.9, "Financial products offered"),
		fa.createContextualField("regulatory_info", "regulation_text", "string", 3, 0.8, "Regulatory information"),
		fa.createContextualField("investment_options", "investment_list", "array", 4, 0.7, "Investment options"),
	}
}

// Industry specific field discovery methods

func (fa *FieldAnalyzer) discoverTechnologyFields(content *ContentInput) []DiscoveredField {
	return []DiscoveredField{
		fa.createContextualField("tech_stack", "technology_list", "array", 3, 0.8, "Technology stack"),
		fa.createContextualField("github_profile", "url", "string", 4, 0.7, "GitHub profile"),
		fa.createContextualField("open_source", "project_list", "array", 5, 0.6, "Open source projects"),
	}
}

func (fa *FieldAnalyzer) discoverManufacturingFields(content *ContentInput) []DiscoveredField {
	return []DiscoveredField{
		fa.createContextualField("production_capacity", "capacity_info", "string", 3, 0.8, "Production capacity"),
		fa.createContextualField("quality_standards", "standard_list", "array", 4, 0.7, "Quality standards"),
		fa.createContextualField("manufacturing_locations", "location_list", "array", 5, 0.6, "Manufacturing locations"),
	}
}

func (fa *FieldAnalyzer) discoverRetailFields(content *ContentInput) []DiscoveredField {
	return []DiscoveredField{
		fa.createContextualField("store_locations", "address", "array", 2, 0.9, "Store locations"),
		fa.createContextualField("product_categories", "category_list", "array", 3, 0.8, "Product categories"),
		fa.createContextualField("loyalty_program", "program_info", "string", 5, 0.6, "Loyalty program"),
	}
}

func (fa *FieldAnalyzer) discoverEducationFields(content *ContentInput) []DiscoveredField {
	return []DiscoveredField{
		fa.createContextualField("courses_offered", "course_list", "array", 2, 0.9, "Courses offered"),
		fa.createContextualField("accreditation", "accreditation_info", "string", 3, 0.8, "Accreditation information"),
		fa.createContextualField("enrollment_info", "enrollment_text", "string", 4, 0.7, "Enrollment information"),
	}
}

// Content category specific field discovery methods

func (fa *FieldAnalyzer) discoverCorporateFields(content *ContentInput) []DiscoveredField {
	return []DiscoveredField{
		fa.createContextualField("company_mission", "mission_text", "string", 4, 0.7, "Company mission statement"),
		fa.createContextualField("leadership_team", "team_info", "array", 3, 0.8, "Leadership team"),
		fa.createContextualField("company_history", "history_text", "string", 5, 0.6, "Company history"),
	}
}

func (fa *FieldAnalyzer) discoverProductFields(content *ContentInput) []DiscoveredField {
	return []DiscoveredField{
		fa.createContextualField("product_features", "feature_list", "array", 2, 0.9, "Product features"),
		fa.createContextualField("pricing_info", "pricing", "string", 3, 0.8, "Pricing information"),
		fa.createContextualField("product_demos", "demo_links", "array", 4, 0.7, "Product demonstrations"),
	}
}

func (fa *FieldAnalyzer) discoverServiceFields(content *ContentInput) []DiscoveredField {
	return []DiscoveredField{
		fa.createContextualField("service_offerings", "service_list", "array", 2, 0.9, "Service offerings"),
		fa.createContextualField("service_areas", "area_list", "array", 3, 0.8, "Service areas"),
		fa.createContextualField("support_options", "support_info", "string", 4, 0.7, "Support options"),
	}
}

// Helper methods

func (fa *FieldAnalyzer) createContextualField(fieldName, fieldType, dataType string, priority int, businessValue float64, description string) DiscoveredField {
	return DiscoveredField{
		FieldName:        fieldName,
		FieldType:        fieldType,
		DataType:         dataType,
		ConfidenceScore:  0.6, // Default confidence for contextual fields
		ExtractionMethod: "contextual_analysis",
		SampleValues:     []string{},
		ValidationRules:  []ValidationRule{},
		Priority:         priority,
		BusinessValue:    businessValue,
		Metadata: map[string]interface{}{
			"description":     description,
			"extraction_base": "contextual",
			"discovery_type":  "business_context",
		},
	}
}

// extractSampleValues extracts sample values from pattern matches
func (fa *FieldAnalyzer) extractSampleValues(patterns []PatternMatch) []string {
	var samples []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		if !seen[pattern.MatchedText] && len(samples) < 5 {
			samples = append(samples, pattern.MatchedText)
			seen[pattern.MatchedText] = true
		}
	}

	return samples
}

// calculateFieldConfidence calculates confidence score for a discovered field
func (fa *FieldAnalyzer) calculateFieldConfidence(patterns []PatternMatch, fieldTypeDef *FieldTypeDefinition, classification *ClassificationResult) float64 {
	if len(patterns) == 0 {
		return 0.0
	}

	// Base confidence from patterns
	var totalConfidence float64
	for _, pattern := range patterns {
		totalConfidence += pattern.ConfidenceScore
	}
	avgConfidence := totalConfidence / float64(len(patterns))

	// Boost based on multiple matches
	multipleMatchBonus := float64(len(patterns)-1) * 0.05
	if multipleMatchBonus > 0.2 {
		multipleMatchBonus = 0.2
	}

	// Boost based on business context alignment
	contextBonus := 0.0
	if classification != nil {
		contextBonus = fa.calculateContextBonus(fieldTypeDef, classification)
	}

	finalConfidence := avgConfidence + multipleMatchBonus + contextBonus

	// Ensure within bounds
	if finalConfidence > 1.0 {
		finalConfidence = 1.0
	}

	return finalConfidence
}

// calculateContextBonus calculates confidence bonus based on business context
func (fa *FieldAnalyzer) calculateContextBonus(fieldTypeDef *FieldTypeDefinition, classification *ClassificationResult) float64 {
	bonus := 0.0

	// Check alignment with business type
	businessTypeAlignment := fa.checkBusinessTypeAlignment(fieldTypeDef.FieldType, classification.BusinessType)
	bonus += businessTypeAlignment * 0.1

	// Check alignment with industry
	industryAlignment := fa.checkIndustryAlignment(fieldTypeDef.FieldType, classification.IndustryCategory)
	bonus += industryAlignment * 0.05

	return bonus
}

// checkBusinessTypeAlignment checks if field type aligns with business type
func (fa *FieldAnalyzer) checkBusinessTypeAlignment(fieldType, businessType string) float64 {
	alignments := map[string]map[string]float64{
		"ecommerce": {
			"payment_methods": 1.0,
			"shipping_info":   1.0,
			"product_info":    1.0,
		},
		"b2b_software": {
			"api_documentation": 1.0,
			"integration_info":  1.0,
			"tech_stack":        0.8,
		},
		"consulting": {
			"expertise":      1.0,
			"case_studies":   1.0,
			"certifications": 0.8,
		},
	}

	if businessMap, exists := alignments[businessType]; exists {
		if alignment, exists := businessMap[fieldType]; exists {
			return alignment
		}
	}

	return 0.0
}

// checkIndustryAlignment checks if field type aligns with industry
func (fa *FieldAnalyzer) checkIndustryAlignment(fieldType, industry string) float64 {
	alignments := map[string]map[string]float64{
		"technology": {
			"tech_stack":        1.0,
			"github_profile":    1.0,
			"api_documentation": 0.8,
		},
		"healthcare": {
			"medical_specialties": 1.0,
			"insurance_accepted":  1.0,
			"appointment_booking": 0.8,
		},
		"financial": {
			"financial_products": 1.0,
			"regulatory_info":    1.0,
			"investment_options": 0.8,
		},
	}

	if industryMap, exists := alignments[industry]; exists {
		if alignment, exists := industryMap[fieldType]; exists {
			return alignment
		}
	}

	return 0.0
}

// determineExtractionMethod determines the best extraction method for a field
func (fa *FieldAnalyzer) determineExtractionMethod(patterns []PatternMatch, fieldTypeDef *FieldTypeDefinition) string {
	if len(patterns) == 0 {
		return "contextual_analysis"
	}

	// Analyze pattern metadata to determine best method
	methods := make(map[string]int)
	for _, pattern := range patterns {
		if method, exists := pattern.Metadata["extraction_method"].(string); exists {
			methods[method]++
		}
	}

	// Find most common method
	maxCount := 0
	bestMethod := "pattern_matching"
	for method, count := range methods {
		if count > maxCount {
			maxCount = count
			bestMethod = method
		}
	}

	return bestMethod
}

// generateValidationRules generates validation rules for a field
func (fa *FieldAnalyzer) generateValidationRules(fieldTypeDef *FieldTypeDefinition, sampleValues []string) []ValidationRule {
	var rules []ValidationRule

	// Add basic validation rules based on field type
	for _, ruleType := range fieldTypeDef.ValidationRules {
		rule := ValidationRule{
			RuleType:        ruleType,
			ConfidenceBoost: 0.1,
			Metadata: map[string]interface{}{
				"source": "field_type_definition",
			},
		}

		// Customize rule based on type
		switch ruleType {
		case "email_format":
			rule.Pattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		case "phone_format":
			rule.Pattern = `^[\+]?[0-9\-\(\)\s]+$`
		case "url_format":
			rule.Pattern = `^https?://[^\s<>"{}|\\^` + "`" + `\[\]]+$`
		case "required":
			rule.MinLength = 1
		case "max_length":
			rule.MaxLength = 255
		}

		rules = append(rules, rule)
	}

	return rules
}

// generateFieldName generates a descriptive field name
func (fa *FieldAnalyzer) generateFieldName(fieldType string, classification *ClassificationResult) string {
	// Use field type as base name
	baseName := fieldType

	// Add business context prefix if relevant
	if classification != nil && classification.BusinessType != "unknown" {
		switch classification.BusinessType {
		case "ecommerce":
			if strings.Contains(fieldType, "payment") {
				baseName = "ecommerce_" + baseName
			}
		case "healthcare":
			if strings.Contains(fieldType, "medical") {
				baseName = "healthcare_" + baseName
			}
		}
	}

	return baseName
}

// enhanceFieldsWithBusinessIntelligence enhances fields with business-specific intelligence
func (fa *FieldAnalyzer) enhanceFieldsWithBusinessIntelligence(fields []DiscoveredField, classification *ClassificationResult) {
	for i := range fields {
		// Adjust priority based on business context
		fields[i].Priority = fa.adjustPriorityForBusinessContext(fields[i], classification)

		// Adjust business value based on industry
		fields[i].BusinessValue = fa.adjustBusinessValueForIndustry(fields[i], classification)

		// Add industry-specific metadata
		fa.addIndustryMetadata(&fields[i], classification)
	}
}

// adjustPriorityForBusinessContext adjusts field priority based on business context
func (fa *FieldAnalyzer) adjustPriorityForBusinessContext(field DiscoveredField, classification *ClassificationResult) int {
	if classification == nil {
		return field.Priority
	}

	// Higher priority (lower number) for business-critical fields
	switch classification.BusinessType {
	case "ecommerce":
		if strings.Contains(field.FieldType, "payment") || strings.Contains(field.FieldType, "shipping") {
			return field.Priority - 1
		}
	case "healthcare":
		if strings.Contains(field.FieldType, "medical") || strings.Contains(field.FieldType, "insurance") {
			return field.Priority - 1
		}
	case "financial":
		if strings.Contains(field.FieldType, "regulatory") || strings.Contains(field.FieldType, "financial") {
			return field.Priority - 1
		}
	}

	return field.Priority
}

// adjustBusinessValueForIndustry adjusts business value based on industry context
func (fa *FieldAnalyzer) adjustBusinessValueForIndustry(field DiscoveredField, classification *ClassificationResult) float64 {
	if classification == nil {
		return field.BusinessValue
	}

	// Industry-specific value adjustments
	switch classification.IndustryCategory {
	case "technology":
		if strings.Contains(field.FieldType, "tech") || strings.Contains(field.FieldType, "api") {
			return field.BusinessValue + 0.1
		}
	case "healthcare":
		if strings.Contains(field.FieldType, "medical") || strings.Contains(field.FieldType, "patient") {
			return field.BusinessValue + 0.15
		}
	case "financial":
		if strings.Contains(field.FieldType, "financial") || strings.Contains(field.FieldType, "investment") {
			return field.BusinessValue + 0.2
		}
	}

	return field.BusinessValue
}

// addIndustryMetadata adds industry-specific metadata to fields
func (fa *FieldAnalyzer) addIndustryMetadata(field *DiscoveredField, classification *ClassificationResult) {
	if classification == nil {
		return
	}

	if field.Metadata == nil {
		field.Metadata = make(map[string]interface{})
	}

	field.Metadata["business_type"] = classification.BusinessType
	field.Metadata["industry"] = classification.IndustryCategory
	field.Metadata["content_categories"] = classification.ContentCategories
}

// sortFieldsByImportance sorts fields by priority and business value
func (fa *FieldAnalyzer) sortFieldsByImportance(fields []DiscoveredField) {
	sort.Slice(fields, func(i, j int) bool {
		// Primary sort: priority (lower number = higher priority)
		if fields[i].Priority != fields[j].Priority {
			return fields[i].Priority < fields[j].Priority
		}
		// Secondary sort: business value (higher = better)
		if fields[i].BusinessValue != fields[j].BusinessValue {
			return fields[i].BusinessValue > fields[j].BusinessValue
		}
		// Tertiary sort: confidence (higher = better)
		return fields[i].ConfidenceScore > fields[j].ConfidenceScore
	})
}

// initializeValidators initializes field validators
func (fa *FieldAnalyzer) initializeValidators() {
	// Initialize basic validators
	fa.validators["email"] = &EmailValidator{}
	fa.validators["phone"] = &PhoneValidator{}
	fa.validators["url"] = &URLValidator{}
	fa.validators["address"] = &AddressValidator{}
}

// getStringValue safely gets string value from classification result
func getStringValue(classification *ClassificationResult, field string) string {
	if classification == nil {
		return "unknown"
	}

	switch field {
	case "BusinessType":
		return classification.BusinessType
	case "IndustryCategory":
		return classification.IndustryCategory
	default:
		return "unknown"
	}
}

// Validator implementations

type EmailValidator struct{}

func (v *EmailValidator) ValidateField(value string) (bool, float64) {
	if strings.Contains(value, "@") && strings.Contains(value, ".") {
		return true, 0.9
	}
	return false, 0.0
}

func (v *EmailValidator) GetDataType() string         { return "string" }
func (v *EmailValidator) GetConfidenceBoost() float64 { return 0.1 }

type PhoneValidator struct{}

func (v *PhoneValidator) ValidateField(value string) (bool, float64) {
	// Simple phone validation
	digits := strings.Count(value, "0") + strings.Count(value, "1") + strings.Count(value, "2") +
		strings.Count(value, "3") + strings.Count(value, "4") + strings.Count(value, "5") +
		strings.Count(value, "6") + strings.Count(value, "7") + strings.Count(value, "8") +
		strings.Count(value, "9")

	if digits >= 10 && digits <= 15 {
		return true, 0.8
	}
	return false, 0.0
}

func (v *PhoneValidator) GetDataType() string         { return "string" }
func (v *PhoneValidator) GetConfidenceBoost() float64 { return 0.1 }

type URLValidator struct{}

func (v *URLValidator) ValidateField(value string) (bool, float64) {
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return true, 0.95
	}
	return false, 0.0
}

func (v *URLValidator) GetDataType() string         { return "string" }
func (v *URLValidator) GetConfidenceBoost() float64 { return 0.1 }

type AddressValidator struct{}

func (v *AddressValidator) ValidateField(value string) (bool, float64) {
	// Simple address validation
	if strings.Contains(value, ",") && len(value) > 20 {
		return true, 0.7
	}
	return false, 0.0
}

func (v *AddressValidator) GetDataType() string         { return "string" }
func (v *AddressValidator) GetConfidenceBoost() float64 { return 0.1 }

// getFieldTypeDefinitions returns built-in field type definitions
func getFieldTypeDefinitions() []FieldTypeDefinition {
	return []FieldTypeDefinition{
		{
			FieldType:       "email",
			DisplayName:     "Email Address",
			Description:     "Email contact information",
			DataType:        "string",
			Priority:        1,
			BusinessValue:   0.9,
			ContextKeywords: []string{"email", "contact", "mail"},
			ValidationRules: []string{"email_format", "required"},
			Examples:        []string{"info@example.com", "contact@company.org"},
		},
		{
			FieldType:       "phone",
			DisplayName:     "Phone Number",
			Description:     "Phone contact information",
			DataType:        "string",
			Priority:        1,
			BusinessValue:   0.9,
			ContextKeywords: []string{"phone", "tel", "call"},
			ValidationRules: []string{"phone_format", "required"},
			Examples:        []string{"(555) 123-4567", "+1-555-123-4567"},
		},
		{
			FieldType:       "address",
			DisplayName:     "Street Address",
			Description:     "Physical business address",
			DataType:        "string",
			Priority:        1,
			BusinessValue:   0.8,
			ContextKeywords: []string{"address", "location", "street"},
			ValidationRules: []string{"required", "max_length"},
			Examples:        []string{"123 Main St, City, ST 12345"},
		},
		{
			FieldType:       "url",
			DisplayName:     "Website URL",
			Description:     "Website or web page URL",
			DataType:        "string",
			Priority:        2,
			BusinessValue:   0.7,
			ContextKeywords: []string{"website", "url", "link"},
			ValidationRules: []string{"url_format"},
			Examples:        []string{"https://example.com", "http://company.org"},
		},
		{
			FieldType:       "social_media",
			DisplayName:     "Social Media",
			Description:     "Social media profile links",
			DataType:        "string",
			Priority:        3,
			BusinessValue:   0.6,
			ContextKeywords: []string{"social", "facebook", "twitter", "linkedin"},
			ValidationRules: []string{"url_format"},
			Examples:        []string{"https://facebook.com/company"},
		},
	}
}

// getDataTypeDefinitions returns data type definitions
func getDataTypeDefinitions() []DataTypeDefinition {
	return []DataTypeDefinition{
		{
			DataType:        "string",
			DisplayName:     "Text String",
			ValidationRules: []string{"max_length"},
			FormatPatterns:  []string{},
			Examples:        []string{"Sample text", "Company Name"},
		},
		{
			DataType:        "array",
			DisplayName:     "Array/List",
			ValidationRules: []string{"min_items"},
			FormatPatterns:  []string{},
			Examples:        []string{"[item1, item2, item3]"},
		},
		{
			DataType:        "number",
			DisplayName:     "Numeric Value",
			ValidationRules: []string{"numeric_format"},
			FormatPatterns:  []string{`^\d+$`, `^\d+\.\d+$`},
			Examples:        []string{"123", "45.67"},
		},
		{
			DataType:        "boolean",
			DisplayName:     "True/False",
			ValidationRules: []string{"boolean_format"},
			FormatPatterns:  []string{`^(true|false|yes|no)$`},
			Examples:        []string{"true", "false", "yes", "no"},
		},
	}
}
