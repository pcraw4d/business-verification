package data_extraction

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"kyb-platform/internal/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ComplianceRegulatoryExtractor extracts compliance and regulatory information
type ComplianceRegulatoryExtractor struct {
	// Configuration
	config *ComplianceRegulatoryConfig

	// Observability
	logger *observability.Logger
	tracer trace.Tracer

	// Extraction components
	regulatoryDetector *RegulatoryBodyDetector
	licenseExtractor   *LicenseCertificationExtractor
	complianceMonitor  *ComplianceStatusMonitor
	entityTypeDetector *LegalEntityTypeDetector
	sanctionScreener   *SanctionListScreener
	complianceAnalyzer *ComplianceAnalyzer

	// Thread safety
	extractionMux sync.RWMutex
}

// ComplianceRegulatoryConfig configuration for compliance extraction
type ComplianceRegulatoryConfig struct {
	// Regulatory body detection settings
	RegulatoryDetectionEnabled bool
	RegulatoryBodies           map[string][]string
	RegulatoryKeywords         []string
	RegulatoryPatterns         []string

	// License and certification settings
	LicenseExtractionEnabled bool
	LicensePatterns          []string
	LicenseKeywords          []string
	CertificationPatterns    []string
	CertificationKeywords    []string

	// Compliance monitoring settings
	ComplianceMonitoringEnabled bool
	ComplianceIndicators        []string
	ComplianceKeywords          []string
	ComplianceThresholds        map[string]float64

	// Legal entity type settings
	EntityTypeDetectionEnabled bool
	EntityTypePatterns         []string
	EntityTypeKeywords         map[string][]string
	EntityTypeClassifications  map[string]string

	// Sanction screening settings
	SanctionScreeningEnabled bool
	SanctionPatterns         []string
	SanctionKeywords         []string
	SanctionThresholds       map[string]float64

	// Analysis settings
	AnalysisEnabled     bool
	ConfidenceThreshold float64
	MaxExtractionTime   time.Duration
}

// ComplianceRegulatoryData represents extracted compliance and regulatory data
type ComplianceRegulatoryData struct {
	// Regulatory body information
	RegulatoryInfo *RegulatoryBodyInfo

	// License and certification information
	LicenseInfo *LicenseCertificationInfo

	// Compliance status information
	ComplianceInfo *ComplianceStatusInfo

	// Legal entity type information
	EntityTypeInfo *LegalEntityTypeInfo

	// Sanction screening information
	SanctionInfo *SanctionScreeningInfo

	// Analysis results
	Analysis *ComplianceAnalysis

	// Metadata
	ExtractionTime time.Time
	Confidence     float64
	Sources        []string
}

// RegulatoryBodyInfo represents regulatory body information
type RegulatoryBodyInfo struct {
	RegulatoryBodies []string
	Jurisdictions    []string
	RegulatoryAreas  []string
	ComplianceLevel  string
	Confidence       float64
	Sources          []string
}

// LicenseCertificationInfo represents license and certification information
type LicenseCertificationInfo struct {
	Licenses        []License
	Certifications  []Certification
	ExpirationDates []time.Time
	Status          string
	Confidence      float64
	Sources         []string
}

// License represents a business license
type License struct {
	Type         string
	Number       string
	Issuer       string
	IssueDate    time.Time
	ExpiryDate   time.Time
	Status       string
	Jurisdiction string
}

// Certification represents a business certification
type Certification struct {
	Type       string
	Issuer     string
	IssueDate  time.Time
	ExpiryDate time.Time
	Status     string
	Standard   string
}

// ComplianceStatusInfo represents compliance status information
type ComplianceStatusInfo struct {
	ComplianceScore float64
	ComplianceLevel string
	ComplianceAreas []string
	RiskAreas       []string
	LastAssessment  time.Time
	NextAssessment  time.Time
	Confidence      float64
	Sources         []string
}

// LegalEntityTypeInfo represents legal entity type information
type LegalEntityTypeInfo struct {
	EntityType         string
	EntitySubtype      string
	Jurisdiction       string
	FormationDate      time.Time
	RegistrationNumber string
	TaxID              string
	Confidence         float64
	Sources            []string
}

// SanctionScreeningInfo represents sanction screening information
type SanctionScreeningInfo struct {
	ScreeningResult string
	SanctionMatches []SanctionMatch
	RiskScore       float64
	RiskLevel       string
	LastScreened    time.Time
	NextScreening   time.Time
	Confidence      float64
	Sources         []string
}

// SanctionMatch represents a sanction list match
type SanctionMatch struct {
	ListName     string
	MatchType    string
	MatchScore   float64
	SanctionType string
	Reason       string
	DateAdded    time.Time
}

// ComplianceAnalysis represents comprehensive compliance analysis
type ComplianceAnalysis struct {
	OverallCompliance string
	ComplianceScore   float64
	KeyStrengths      []string
	KeyRisks          []string
	Recommendations   []string
	Confidence        float64
}

// RegulatoryBodyDetector detects regulatory bodies
type RegulatoryBodyDetector struct {
	enabled          bool
	regulatoryBodies map[string][]string
	keywords         []string
	patterns         []*regexp.Regexp
	detectionMux     sync.RWMutex
}

// LicenseCertificationExtractor extracts license and certification information
type LicenseCertificationExtractor struct {
	enabled               bool
	licensePatterns       []*regexp.Regexp
	licenseKeywords       []string
	certificationPatterns []*regexp.Regexp
	certificationKeywords []string
	extractionMux         sync.RWMutex
}

// ComplianceStatusMonitor monitors compliance status
type ComplianceStatusMonitor struct {
	enabled       bool
	indicators    []string
	keywords      []string
	thresholds    map[string]float64
	monitoringMux sync.RWMutex
}

// LegalEntityTypeDetector detects legal entity types
type LegalEntityTypeDetector struct {
	enabled         bool
	patterns        []*regexp.Regexp
	keywords        map[string][]string
	classifications map[string]string
	detectionMux    sync.RWMutex
}

// SanctionListScreener screens against sanction lists
type SanctionListScreener struct {
	enabled      bool
	patterns     []*regexp.Regexp
	keywords     []string
	thresholds   map[string]float64
	screeningMux sync.RWMutex
}

// ComplianceAnalyzer performs comprehensive compliance analysis
type ComplianceAnalyzer struct {
	enabled             bool
	confidenceThreshold float64
	analysisMux         sync.RWMutex
}

// NewComplianceRegulatoryExtractor creates a new compliance extractor
func NewComplianceRegulatoryExtractor(config *ComplianceRegulatoryConfig, logger *observability.Logger, tracer trace.Tracer) *ComplianceRegulatoryExtractor {
	if config == nil {
		config = &ComplianceRegulatoryConfig{
			RegulatoryDetectionEnabled: true,
			RegulatoryBodies: map[string][]string{
				"financial":     {"SEC", "FINRA", "FDIC", "OCC", "CFTC", "CFPB"},
				"healthcare":    {"FDA", "CMS", "HIPAA", "HHS", "CDC"},
				"environmental": {"EPA", "DEQ", "DNR", "DEC"},
				"labor":         {"DOL", "OSHA", "EEOC", "NLRB"},
				"tax":           {"IRS", "State Tax Authorities"},
			},
			RegulatoryKeywords: []string{
				"regulated", "regulation", "compliance", "regulatory", "authority",
				"commission", "board", "agency", "department", "bureau",
			},
			RegulatoryPatterns: []string{
				`(?i)(regulated|regulation|compliance|regulatory)`,
				`(?i)(SEC|FINRA|FDIC|OCC|CFTC|CFPB|FDA|CMS|EPA|DOL|OSHA|IRS)`,
			},
			LicenseExtractionEnabled: true,
			LicensePatterns: []string{
				`(?i)(license|licensing|permit|authorization)`,
				`(?i)(license\s+number|permit\s+number|auth\s+number)`,
				`(?i)(licensed|permitted|authorized)`,
			},
			LicenseKeywords: []string{
				"license", "licensing", "permit", "authorization", "certified",
				"accredited", "approved", "registered", "compliant",
			},
			CertificationPatterns: []string{
				`(?i)(certification|certified|accreditation|accredited)`,
				`(?i)(ISO|SOC|PCI|HIPAA|GDPR|SOX)`,
				`(?i)(certification\s+number|accreditation\s+number)`,
			},
			CertificationKeywords: []string{
				"certification", "certified", "accreditation", "accredited",
				"ISO", "SOC", "PCI", "HIPAA", "GDPR", "SOX", "compliance",
			},
			ComplianceMonitoringEnabled: true,
			ComplianceIndicators: []string{
				"compliant", "compliance", "regulated", "certified", "licensed",
				"accredited", "approved", "registered", "audited", "monitored",
			},
			ComplianceKeywords: []string{
				"compliant", "compliance", "regulated", "certified", "licensed",
				"accredited", "approved", "registered", "audited", "monitored",
				"in compliance", "meets standards", "follows regulations",
			},
			ComplianceThresholds: map[string]float64{
				"min_compliance_score": 0.3,
				"max_risk_areas":       5.0,
			},
			EntityTypeDetectionEnabled: true,
			EntityTypePatterns: []string{
				`(?i)(LLC|Inc\.|Corp\.|Corporation|Limited|Partnership|Sole\s+Proprietorship)`,
				`(?i)(Limited\s+Liability\s+Company|Incorporated|Corporation)`,
				`(?i)(Partnership|Sole\s+Proprietorship|Non-Profit|Foundation)`,
			},
			EntityTypeKeywords: map[string][]string{
				"llc":                 {"LLC", "Limited Liability Company", "Ltd"},
				"corporation":         {"Inc", "Corp", "Corporation", "Incorporated"},
				"partnership":         {"Partnership", "LP", "LLP", "General Partnership"},
				"sole_proprietorship": {"Sole Proprietorship", "Sole Owner", "Individual"},
				"non_profit":          {"Non-Profit", "Foundation", "Charity", "501(c)"},
			},
			EntityTypeClassifications: map[string]string{
				"llc":                 "limited_liability_company",
				"corporation":         "corporation",
				"partnership":         "partnership",
				"sole_proprietorship": "sole_proprietorship",
				"non_profit":          "non_profit",
			},
			SanctionScreeningEnabled: true,
			SanctionPatterns: []string{
				`(?i)(sanction|embargo|restricted|prohibited|banned)`,
				`(?i)(OFAC|SDN|Specially\s+Designated\s+Nationals)`,
				`(?i)(sanctioned|embargoed|restricted|prohibited|banned)`,
			},
			SanctionKeywords: []string{
				"sanction", "embargo", "restricted", "prohibited", "banned",
				"OFAC", "SDN", "Specially Designated Nationals", "blacklist",
			},
			SanctionThresholds: map[string]float64{
				"max_risk_score":  0.7,
				"min_match_score": 0.8,
			},
			AnalysisEnabled:     true,
			ConfidenceThreshold: 0.6,
			MaxExtractionTime:   30 * time.Second,
		}
	}

	cre := &ComplianceRegulatoryExtractor{
		config: config,
		logger: logger,
		tracer: tracer,
	}

	// Initialize components
	cre.regulatoryDetector = &RegulatoryBodyDetector{
		enabled:          config.RegulatoryDetectionEnabled,
		regulatoryBodies: config.RegulatoryBodies,
		keywords:         config.RegulatoryKeywords,
		patterns:         compilePatterns(config.RegulatoryPatterns),
	}

	cre.licenseExtractor = &LicenseCertificationExtractor{
		enabled:               config.LicenseExtractionEnabled,
		licensePatterns:       compilePatterns(config.LicensePatterns),
		licenseKeywords:       config.LicenseKeywords,
		certificationPatterns: compilePatterns(config.CertificationPatterns),
		certificationKeywords: config.CertificationKeywords,
	}

	cre.complianceMonitor = &ComplianceStatusMonitor{
		enabled:    config.ComplianceMonitoringEnabled,
		indicators: config.ComplianceIndicators,
		keywords:   config.ComplianceKeywords,
		thresholds: config.ComplianceThresholds,
	}

	cre.entityTypeDetector = &LegalEntityTypeDetector{
		enabled:         config.EntityTypeDetectionEnabled,
		patterns:        compilePatterns(config.EntityTypePatterns),
		keywords:        config.EntityTypeKeywords,
		classifications: config.EntityTypeClassifications,
	}

	cre.sanctionScreener = &SanctionListScreener{
		enabled:    config.SanctionScreeningEnabled,
		patterns:   compilePatterns(config.SanctionPatterns),
		keywords:   config.SanctionKeywords,
		thresholds: config.SanctionThresholds,
	}

	cre.complianceAnalyzer = &ComplianceAnalyzer{
		enabled:             config.AnalysisEnabled,
		confidenceThreshold: config.ConfidenceThreshold,
	}

	return cre
}

// ExtractComplianceRegulatory extracts compliance and regulatory data
func (cre *ComplianceRegulatoryExtractor) ExtractComplianceRegulatory(ctx context.Context, businessName, websiteContent, description string) (*ComplianceRegulatoryData, error) {
	ctx, span := cre.tracer.Start(ctx, "ComplianceRegulatoryExtractor.ExtractComplianceRegulatory")
	defer span.End()

	span.SetAttributes(
		attribute.String("business_name", businessName),
		attribute.Int("content_length", len(websiteContent)),
		attribute.Int("description_length", len(description)),
	)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, cre.config.MaxExtractionTime)
	defer cancel()

	// Combine all text for analysis
	combinedText := strings.Join([]string{businessName, websiteContent, description}, " ")

	// Detect regulatory bodies
	var regulatoryInfo *RegulatoryBodyInfo
	if cre.config.RegulatoryDetectionEnabled {
		regulatoryInfo = cre.regulatoryDetector.DetectRegulatoryBodies(ctx, combinedText)
	}

	// Extract license and certification information
	var licenseInfo *LicenseCertificationInfo
	if cre.config.LicenseExtractionEnabled {
		licenseInfo = cre.licenseExtractor.ExtractLicenseCertification(ctx, combinedText)
	}

	// Monitor compliance status
	var complianceInfo *ComplianceStatusInfo
	if cre.config.ComplianceMonitoringEnabled {
		complianceInfo = cre.complianceMonitor.MonitorComplianceStatus(ctx, combinedText)
	}

	// Detect legal entity type
	var entityTypeInfo *LegalEntityTypeInfo
	if cre.config.EntityTypeDetectionEnabled {
		entityTypeInfo = cre.entityTypeDetector.DetectEntityType(ctx, combinedText)
	}

	// Screen against sanction lists
	var sanctionInfo *SanctionScreeningInfo
	if cre.config.SanctionScreeningEnabled {
		sanctionInfo = cre.sanctionScreener.ScreenSanctions(ctx, combinedText)
	}

	// Perform comprehensive analysis
	var analysis *ComplianceAnalysis
	if cre.config.AnalysisEnabled {
		analysis = cre.complianceAnalyzer.AnalyzeCompliance(ctx, regulatoryInfo, licenseInfo, complianceInfo, entityTypeInfo, sanctionInfo)
	}

	// Calculate overall confidence
	confidence := cre.calculateOverallConfidence(regulatoryInfo, licenseInfo, complianceInfo, entityTypeInfo, sanctionInfo)

	// Collect sources
	sources := cre.collectSources(regulatoryInfo, licenseInfo, complianceInfo, entityTypeInfo, sanctionInfo)

	result := &ComplianceRegulatoryData{
		RegulatoryInfo: regulatoryInfo,
		LicenseInfo:    licenseInfo,
		ComplianceInfo: complianceInfo,
		EntityTypeInfo: entityTypeInfo,
		SanctionInfo:   sanctionInfo,
		Analysis:       analysis,
		ExtractionTime: time.Now(),
		Confidence:     confidence,
		Sources:        sources,
	}

	cre.logger.Info("compliance regulatory extraction completed", map[string]interface{}{
		"business_name":     businessName,
		"confidence":        confidence,
		"regulatory_bodies": regulatoryInfo != nil && len(regulatoryInfo.RegulatoryBodies) > 0,
		"has_licenses":      licenseInfo != nil && len(licenseInfo.Licenses) > 0,
		"compliance_score":  complianceInfo != nil && complianceInfo.ComplianceScore > 0,
		"entity_type":       entityTypeInfo != nil && entityTypeInfo.EntityType != "",
		"screening_result":  sanctionInfo != nil && sanctionInfo.ScreeningResult != "",
	})

	return result, nil
}

// RegulatoryBodyDetector methods

func (rbd *RegulatoryBodyDetector) DetectRegulatoryBodies(ctx context.Context, text string) *RegulatoryBodyInfo {
	rbd.detectionMux.Lock()
	defer rbd.detectionMux.Unlock()

	text = strings.ToLower(text)
	var regulatoryBodies []string
	var jurisdictions []string
	var regulatoryAreas []string

	// Detect regulatory bodies by category
	for category, bodies := range rbd.regulatoryBodies {
		for _, body := range bodies {
			if strings.Contains(text, strings.ToLower(body)) {
				regulatoryBodies = append(regulatoryBodies, body)
				regulatoryAreas = append(regulatoryAreas, category)
			}
		}
	}

	// Detect regulatory keywords
	hasRegulatoryKeywords := false
	for _, keyword := range rbd.keywords {
		if strings.Contains(text, keyword) {
			hasRegulatoryKeywords = true
			break
		}
	}

	// Determine compliance level
	complianceLevel := rbd.determineComplianceLevel(regulatoryBodies, hasRegulatoryKeywords)

	// Calculate confidence
	confidence := rbd.calculateRegulatoryConfidence(regulatoryBodies, hasRegulatoryKeywords)

	return &RegulatoryBodyInfo{
		RegulatoryBodies: regulatoryBodies,
		Jurisdictions:    jurisdictions,
		RegulatoryAreas:  regulatoryAreas,
		ComplianceLevel:  complianceLevel,
		Confidence:       confidence,
		Sources:          []string{"text_analysis"},
	}
}

func (rbd *RegulatoryBodyDetector) determineComplianceLevel(bodies []string, hasKeywords bool) string {
	if len(bodies) > 3 {
		return "high"
	} else if len(bodies) > 1 {
		return "medium"
	} else if len(bodies) > 0 || hasKeywords {
		return "low"
	}
	return "none"
}

func (rbd *RegulatoryBodyDetector) calculateRegulatoryConfidence(bodies []string, hasKeywords bool) float64 {
	confidence := 0.0

	// Base confidence from regulatory body detection
	if len(bodies) > 0 {
		confidence += 0.6
	}

	// Additional confidence from keywords
	if hasKeywords {
		confidence += 0.2
	}

	// Additional confidence from number of bodies
	if len(bodies) > 1 {
		confidence += 0.2
	}

	return confidence
}

// LicenseCertificationExtractor methods

func (lce *LicenseCertificationExtractor) ExtractLicenseCertification(ctx context.Context, text string) *LicenseCertificationInfo {
	lce.extractionMux.Lock()
	defer lce.extractionMux.Unlock()

	text = strings.ToLower(text)

	// Extract licenses
	licenses := lce.extractLicenses(text)

	// Extract certifications
	certifications := lce.extractCertifications(text)

	// Determine overall status
	status := lce.determineOverallStatus(licenses, certifications)

	// Calculate confidence
	confidence := lce.calculateLicenseConfidence(licenses, certifications)

	return &LicenseCertificationInfo{
		Licenses:       licenses,
		Certifications: certifications,
		Status:         status,
		Confidence:     confidence,
		Sources:        []string{"text_analysis"},
	}
}

func (lce *LicenseCertificationExtractor) extractLicenses(text string) []License {
	var licenses []License

	// Simple license extraction - in production, use more sophisticated parsing
	for _, keyword := range lce.licenseKeywords {
		if strings.Contains(text, keyword) {
			license := License{
				Type:         "general",
				Number:       "unknown",
				Issuer:       "unknown",
				IssueDate:    time.Now(),
				ExpiryDate:   time.Now().AddDate(1, 0, 0),
				Status:       "active",
				Jurisdiction: "unknown",
			}
			licenses = append(licenses, license)
		}
	}

	return licenses
}

func (lce *LicenseCertificationExtractor) extractCertifications(text string) []Certification {
	var certifications []Certification

	// Simple certification extraction - in production, use more sophisticated parsing
	for _, keyword := range lce.certificationKeywords {
		if strings.Contains(text, keyword) {
			certification := Certification{
				Type:       "general",
				Issuer:     "unknown",
				IssueDate:  time.Now(),
				ExpiryDate: time.Now().AddDate(1, 0, 0),
				Status:     "active",
				Standard:   "unknown",
			}
			certifications = append(certifications, certification)
		}
	}

	return certifications
}

func (lce *LicenseCertificationExtractor) determineOverallStatus(licenses []License, certifications []Certification) string {
	if len(licenses) > 0 && len(certifications) > 0 {
		return "fully_licensed_certified"
	} else if len(licenses) > 0 {
		return "licensed"
	} else if len(certifications) > 0 {
		return "certified"
	}
	return "none"
}

func (lce *LicenseCertificationExtractor) calculateLicenseConfidence(licenses []License, certifications []Certification) float64 {
	confidence := 0.0

	// Base confidence from license detection
	if len(licenses) > 0 {
		confidence += 0.4
	}

	// Additional confidence from certification detection
	if len(certifications) > 0 {
		confidence += 0.4
	}

	// Additional confidence from multiple items
	if len(licenses) > 1 || len(certifications) > 1 {
		confidence += 0.2
	}

	return confidence
}

// ComplianceStatusMonitor methods

func (csm *ComplianceStatusMonitor) MonitorComplianceStatus(ctx context.Context, text string) *ComplianceStatusInfo {
	csm.monitoringMux.Lock()
	defer csm.monitoringMux.Unlock()

	text = strings.ToLower(text)

	// Calculate compliance score
	score := csm.calculateComplianceScore(text)

	// Determine compliance level
	level := csm.determineComplianceLevel(score)

	// Identify compliance areas
	areas := csm.identifyComplianceAreas(text)

	// Identify risk areas
	risks := csm.identifyRiskAreas(text)

	// Calculate confidence
	confidence := csm.calculateComplianceConfidence(text, score, areas)

	return &ComplianceStatusInfo{
		ComplianceScore: score,
		ComplianceLevel: level,
		ComplianceAreas: areas,
		RiskAreas:       risks,
		LastAssessment:  time.Now(),
		NextAssessment:  time.Now().AddDate(0, 6, 0), // 6 months from now
		Confidence:      confidence,
		Sources:         []string{"text_analysis"},
	}
}

func (csm *ComplianceStatusMonitor) calculateComplianceScore(text string) float64 {
	score := 0.5 // Base score

	// Positive indicators
	for _, keyword := range csm.keywords {
		if strings.Contains(text, keyword) {
			score += 0.1
		}
	}

	// Negative indicators
	negativeKeywords := []string{"non-compliant", "violation", "penalty", "fine", "suspended"}
	for _, keyword := range negativeKeywords {
		if strings.Contains(text, keyword) {
			score -= 0.2
		}
	}

	// Clamp score between 0 and 1
	if score < 0 {
		score = 0
	} else if score > 1 {
		score = 1
	}

	return score
}

func (csm *ComplianceStatusMonitor) determineComplianceLevel(score float64) string {
	if score >= 0.8 {
		return "excellent"
	} else if score >= 0.7 {
		return "good"
	} else if score >= 0.6 {
		return "fair"
	} else if score >= 0.5 {
		return "poor"
	} else {
		return "critical"
	}
}

func (csm *ComplianceStatusMonitor) identifyComplianceAreas(text string) []string {
	var areas []string

	for _, indicator := range csm.indicators {
		if strings.Contains(text, indicator) {
			areas = append(areas, indicator)
		}
	}

	return areas
}

func (csm *ComplianceStatusMonitor) identifyRiskAreas(text string) []string {
	riskKeywords := []string{"non-compliant", "violation", "penalty", "fine", "suspended", "investigation"}
	var risks []string

	for _, keyword := range riskKeywords {
		if strings.Contains(text, keyword) {
			risks = append(risks, keyword)
		}
	}

	return risks
}

func (csm *ComplianceStatusMonitor) calculateComplianceConfidence(text string, score float64, areas []string) float64 {
	confidence := 0.0

	// Base confidence from score calculation
	confidence += 0.4

	// Confidence from area identification
	if len(areas) > 0 {
		confidence += 0.3
	}

	// Confidence from text length
	if len(text) > 1000 {
		confidence += 0.3
	} else if len(text) > 500 {
		confidence += 0.2
	} else if len(text) > 100 {
		confidence += 0.1
	}

	return confidence
}

// LegalEntityTypeDetector methods

func (letd *LegalEntityTypeDetector) DetectEntityType(ctx context.Context, text string) *LegalEntityTypeInfo {
	letd.detectionMux.Lock()
	defer letd.detectionMux.Unlock()

	text = strings.ToLower(text)

	// Detect entity type
	entityType := letd.detectEntityType(text)

	// Determine entity subtype
	subtype := letd.determineEntitySubtype(text)

	// Extract jurisdiction
	jurisdiction := letd.extractJurisdiction(text)

	// Calculate confidence
	confidence := letd.calculateEntityConfidence(text, entityType)

	return &LegalEntityTypeInfo{
		EntityType:         entityType,
		EntitySubtype:      subtype,
		Jurisdiction:       jurisdiction,
		FormationDate:      time.Now(),
		RegistrationNumber: "unknown",
		TaxID:              "unknown",
		Confidence:         confidence,
		Sources:            []string{"text_analysis"},
	}
}

func (letd *LegalEntityTypeDetector) detectEntityType(text string) string {
	for entityType, keywords := range letd.keywords {
		for _, keyword := range keywords {
			if strings.Contains(text, strings.ToLower(keyword)) {
				if classification, exists := letd.classifications[entityType]; exists {
					return classification
				}
				return entityType
			}
		}
	}
	return "unknown"
}

func (letd *LegalEntityTypeDetector) determineEntitySubtype(text string) string {
	// Simple subtype determination - in production, use more sophisticated logic
	if strings.Contains(text, "professional") {
		return "professional"
	} else if strings.Contains(text, "holding") {
		return "holding"
	} else if strings.Contains(text, "subsidiary") {
		return "subsidiary"
	}
	return "general"
}

func (letd *LegalEntityTypeDetector) extractJurisdiction(text string) string {
	// Simple jurisdiction extraction - in production, use more sophisticated logic
	if strings.Contains(text, "delaware") {
		return "Delaware"
	} else if strings.Contains(text, "california") {
		return "California"
	} else if strings.Contains(text, "new york") {
		return "New York"
	}
	return "unknown"
}

func (letd *LegalEntityTypeDetector) calculateEntityConfidence(text string, entityType string) float64 {
	confidence := 0.0

	// Base confidence from entity type detection
	if entityType != "unknown" {
		confidence += 0.6
	}

	// Additional confidence from text length
	if len(text) > 500 {
		confidence += 0.2
	}

	// Additional confidence from multiple entity indicators
	entityIndicators := []string{"llc", "inc", "corp", "partnership", "proprietorship"}
	count := 0
	for _, indicator := range entityIndicators {
		if strings.Contains(text, indicator) {
			count++
		}
	}
	if count > 1 {
		confidence += 0.2
	}

	return confidence
}

// SanctionListScreener methods

func (sls *SanctionListScreener) ScreenSanctions(ctx context.Context, text string) *SanctionScreeningInfo {
	sls.screeningMux.Lock()
	defer sls.screeningMux.Unlock()

	text = strings.ToLower(text)

	// Perform sanction screening
	matches := sls.performSanctionScreening(text)

	// Calculate risk score
	riskScore := sls.calculateRiskScore(matches)

	// Determine risk level
	riskLevel := sls.determineRiskLevel(riskScore)

	// Determine screening result
	result := sls.determineScreeningResult(matches)

	// Calculate confidence
	confidence := sls.calculateScreeningConfidence(text, matches)

	return &SanctionScreeningInfo{
		ScreeningResult: result,
		SanctionMatches: matches,
		RiskScore:       riskScore,
		RiskLevel:       riskLevel,
		LastScreened:    time.Now(),
		NextScreening:   time.Now().AddDate(0, 1, 0), // 1 month from now
		Confidence:      confidence,
		Sources:         []string{"text_analysis"},
	}
}

func (sls *SanctionListScreener) performSanctionScreening(text string) []SanctionMatch {
	var matches []SanctionMatch

	// Simple sanction screening - in production, integrate with actual sanction databases
	for _, keyword := range sls.keywords {
		if strings.Contains(text, keyword) {
			match := SanctionMatch{
				ListName:     "internal_screening",
				MatchType:    "keyword_match",
				MatchScore:   0.8,
				SanctionType: "potential_match",
				Reason:       "keyword detected: " + keyword,
				DateAdded:    time.Now(),
			}
			matches = append(matches, match)
		}
	}

	return matches
}

func (sls *SanctionListScreener) calculateRiskScore(matches []SanctionMatch) float64 {
	if len(matches) == 0 {
		return 0.0
	}

	// Calculate average match score
	totalScore := 0.0
	for _, match := range matches {
		totalScore += match.MatchScore
	}

	return totalScore / float64(len(matches))
}

func (sls *SanctionListScreener) determineRiskLevel(riskScore float64) string {
	if riskScore >= 0.8 {
		return "high"
	} else if riskScore >= 0.5 {
		return "medium"
	} else if riskScore >= 0.2 {
		return "low"
	} else {
		return "very_low"
	}
}

func (sls *SanctionListScreener) determineScreeningResult(matches []SanctionMatch) string {
	if len(matches) == 0 {
		return "clear"
	} else if len(matches) == 1 {
		return "potential_match"
	} else {
		return "multiple_matches"
	}
}

func (sls *SanctionListScreener) calculateScreeningConfidence(text string, matches []SanctionMatch) float64 {
	confidence := 0.0

	// Base confidence from screening
	confidence += 0.4

	// Confidence from match count
	if len(matches) > 0 {
		confidence += 0.3
	}

	// Confidence from text length
	if len(text) > 1000 {
		confidence += 0.3
	} else if len(text) > 500 {
		confidence += 0.2
	} else if len(text) > 100 {
		confidence += 0.1
	}

	return confidence
}

// ComplianceAnalyzer methods

func (ca *ComplianceAnalyzer) AnalyzeCompliance(ctx context.Context, regulatory *RegulatoryBodyInfo, license *LicenseCertificationInfo, compliance *ComplianceStatusInfo, entityType *LegalEntityTypeInfo, sanction *SanctionScreeningInfo) *ComplianceAnalysis {
	ca.analysisMux.Lock()
	defer ca.analysisMux.Unlock()

	// Calculate overall compliance score
	score := ca.calculateOverallComplianceScore(regulatory, license, compliance, entityType, sanction)

	// Determine overall compliance
	overallCompliance := ca.determineOverallCompliance(score)

	// Identify key strengths
	strengths := ca.identifyKeyStrengths(regulatory, license, compliance, entityType, sanction)

	// Identify key risks
	risks := ca.identifyKeyRisks(regulatory, license, compliance, entityType, sanction)

	// Generate recommendations
	recommendations := ca.generateRecommendations(regulatory, license, compliance, entityType, sanction)

	// Calculate confidence
	confidence := ca.calculateAnalysisConfidence(regulatory, license, compliance, entityType, sanction)

	return &ComplianceAnalysis{
		OverallCompliance: overallCompliance,
		ComplianceScore:   score,
		KeyStrengths:      strengths,
		KeyRisks:          risks,
		Recommendations:   recommendations,
		Confidence:        confidence,
	}
}

func (ca *ComplianceAnalyzer) calculateOverallComplianceScore(regulatory *RegulatoryBodyInfo, license *LicenseCertificationInfo, compliance *ComplianceStatusInfo, entityType *LegalEntityTypeInfo, sanction *SanctionScreeningInfo) float64 {
	score := 0.5 // Base score

	// Regulatory body contribution
	if regulatory != nil && len(regulatory.RegulatoryBodies) > 0 {
		score += 0.2
		if regulatory.ComplianceLevel == "high" {
			score += 0.1
		}
	}

	// License and certification contribution
	if license != nil && (len(license.Licenses) > 0 || len(license.Certifications) > 0) {
		score += 0.2
		if license.Status == "fully_licensed_certified" {
			score += 0.1
		}
	}

	// Compliance status contribution
	if compliance != nil {
		score += compliance.ComplianceScore * 0.3
	}

	// Entity type contribution
	if entityType != nil && entityType.EntityType != "unknown" {
		score += 0.1
	}

	// Sanction screening contribution (inverse)
	if sanction != nil && sanction.RiskScore > 0 {
		score -= sanction.RiskScore * 0.2
	}

	// Clamp score between 0 and 1
	if score < 0 {
		score = 0
	} else if score > 1 {
		score = 1
	}

	return score
}

func (ca *ComplianceAnalyzer) determineOverallCompliance(score float64) string {
	if score >= 0.8 {
		return "excellent"
	} else if score >= 0.7 {
		return "good"
	} else if score >= 0.6 {
		return "fair"
	} else if score >= 0.5 {
		return "poor"
	} else {
		return "critical"
	}
}

func (ca *ComplianceAnalyzer) identifyKeyStrengths(regulatory *RegulatoryBodyInfo, license *LicenseCertificationInfo, compliance *ComplianceStatusInfo, entityType *LegalEntityTypeInfo, sanction *SanctionScreeningInfo) []string {
	var strengths []string

	if regulatory != nil && len(regulatory.RegulatoryBodies) > 0 {
		strengths = append(strengths, "regulated_entity")
		if regulatory.ComplianceLevel == "high" {
			strengths = append(strengths, "highly_regulated")
		}
	}

	if license != nil && len(license.Licenses) > 0 {
		strengths = append(strengths, "licensed")
	}

	if license != nil && len(license.Certifications) > 0 {
		strengths = append(strengths, "certified")
	}

	if compliance != nil && compliance.ComplianceScore > 0.7 {
		strengths = append(strengths, "compliant")
	}

	if sanction != nil && sanction.RiskScore < 0.2 {
		strengths = append(strengths, "sanction_clear")
	}

	return strengths
}

func (ca *ComplianceAnalyzer) identifyKeyRisks(regulatory *RegulatoryBodyInfo, license *LicenseCertificationInfo, compliance *ComplianceStatusInfo, entityType *LegalEntityTypeInfo, sanction *SanctionScreeningInfo) []string {
	var risks []string

	if regulatory == nil || len(regulatory.RegulatoryBodies) == 0 {
		risks = append(risks, "unregulated_entity")
	}

	if license == nil || (len(license.Licenses) == 0 && len(license.Certifications) == 0) {
		risks = append(risks, "no_licenses_certifications")
	}

	if compliance != nil && compliance.ComplianceScore < 0.5 {
		risks = append(risks, "low_compliance")
	}

	if sanction != nil && sanction.RiskScore > 0.5 {
		risks = append(risks, "sanction_risk")
	}

	return risks
}

func (ca *ComplianceAnalyzer) generateRecommendations(regulatory *RegulatoryBodyInfo, license *LicenseCertificationInfo, compliance *ComplianceStatusInfo, entityType *LegalEntityTypeInfo, sanction *SanctionScreeningInfo) []string {
	var recommendations []string

	if regulatory == nil || len(regulatory.RegulatoryBodies) == 0 {
		recommendations = append(recommendations, "assess_regulatory_requirements")
	}

	if license == nil || (len(license.Licenses) == 0 && len(license.Certifications) == 0) {
		recommendations = append(recommendations, "obtain_necessary_licenses_certifications")
	}

	if compliance != nil && compliance.ComplianceScore < 0.6 {
		recommendations = append(recommendations, "improve_compliance_program")
	}

	if sanction != nil && sanction.RiskScore > 0.3 {
		recommendations = append(recommendations, "conduct_detailed_sanction_screening")
	}

	return recommendations
}

func (ca *ComplianceAnalyzer) calculateAnalysisConfidence(regulatory *RegulatoryBodyInfo, license *LicenseCertificationInfo, compliance *ComplianceStatusInfo, entityType *LegalEntityTypeInfo, sanction *SanctionScreeningInfo) float64 {
	confidence := 0.0
	count := 0

	if regulatory != nil {
		confidence += regulatory.Confidence
		count++
	}

	if license != nil {
		confidence += license.Confidence
		count++
	}

	if compliance != nil {
		confidence += compliance.Confidence
		count++
	}

	if entityType != nil {
		confidence += entityType.Confidence
		count++
	}

	if sanction != nil {
		confidence += sanction.Confidence
		count++
	}

	if count > 0 {
		return confidence / float64(count)
	}

	return 0.0
}

// Helper methods

func (cre *ComplianceRegulatoryExtractor) calculateOverallConfidence(regulatory *RegulatoryBodyInfo, license *LicenseCertificationInfo, compliance *ComplianceStatusInfo, entityType *LegalEntityTypeInfo, sanction *SanctionScreeningInfo) float64 {
	confidence := 0.0
	count := 0

	if regulatory != nil {
		confidence += regulatory.Confidence
		count++
	}

	if license != nil {
		confidence += license.Confidence
		count++
	}

	if compliance != nil {
		confidence += compliance.Confidence
		count++
	}

	if entityType != nil {
		confidence += entityType.Confidence
		count++
	}

	if sanction != nil {
		confidence += sanction.Confidence
		count++
	}

	if count > 0 {
		return confidence / float64(count)
	}

	return 0.0
}

func (cre *ComplianceRegulatoryExtractor) collectSources(regulatory *RegulatoryBodyInfo, license *LicenseCertificationInfo, compliance *ComplianceStatusInfo, entityType *LegalEntityTypeInfo, sanction *SanctionScreeningInfo) []string {
	var sources []string

	if regulatory != nil {
		sources = append(sources, regulatory.Sources...)
	}

	if license != nil {
		sources = append(sources, license.Sources...)
	}

	if compliance != nil {
		sources = append(sources, compliance.Sources...)
	}

	if entityType != nil {
		sources = append(sources, entityType.Sources...)
	}

	if sanction != nil {
		sources = append(sources, sanction.Sources...)
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueSources []string
	for _, source := range sources {
		if !seen[source] {
			seen[source] = true
			uniqueSources = append(uniqueSources, source)
		}
	}

	return uniqueSources
}
