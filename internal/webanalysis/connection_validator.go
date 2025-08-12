package webanalysis

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
)

// ConnectionValidator provides comprehensive business-website connection verification
type ConnectionValidator struct {
	nameMatcher         *BusinessNameMatcher
	addressVerifier     *AddressVerifier
	contactVerifier     *ContactVerifier
	registrationChecker *RegistrationChecker
	ownershipScorer     *OwnershipScorer
	confidenceAssessor  *ConfidenceAssessor
	config              ConnectionValidatorConfig
	mu                  sync.RWMutex
}

// ConnectionValidatorConfig holds configuration for connection validation
type ConnectionValidatorConfig struct {
	EnableNameMatching        bool          `json:"enable_name_matching"`
	EnableAddressVerification bool          `json:"enable_address_verification"`
	EnableContactVerification bool          `json:"enable_contact_verification"`
	EnableRegistrationCheck   bool          `json:"enable_registration_check"`
	EnableOwnershipScoring    bool          `json:"enable_ownership_scoring"`
	MinConfidenceScore        float64       `json:"min_confidence_score"`
	MaxFuzzyDistance          int           `json:"max_fuzzy_distance"`
	EnableCrossReference      bool          `json:"enable_cross_reference"`
	Timeout                   time.Duration `json:"timeout"`
}

// ConnectionValidationResult represents the result of connection validation
type ConnectionValidationResult struct {
	BusinessName            string                   `json:"business_name"`
	WebsiteURL              string                   `json:"website_url"`
	IsConnected             bool                     `json:"is_connected"`
	NameMatchResult         *NameMatchResult         `json:"name_match_result"`
	AddressMatchResult      *AddressMatchResult      `json:"address_match_result"`
	ContactMatchResult      *ContactMatchResult      `json:"contact_match_result"`
	RegistrationMatchResult *RegistrationMatchResult `json:"registration_match_result"`
	OwnershipScore          float64                  `json:"ownership_score"`
	OverallConfidence       float64                  `json:"overall_confidence"`
	ConnectionStrength      string                   `json:"connection_strength"` // strong, moderate, weak, none
	Evidence                []ConnectionEvidence     `json:"evidence"`
	Warnings                []string                 `json:"warnings"`
	Errors                  []string                 `json:"errors"`
	ValidationTime          time.Duration            `json:"validation_time"`
	LastChecked             time.Time                `json:"last_checked"`
}

// BusinessNameMatcher matches business names using various algorithms
type BusinessNameMatcher struct {
	fuzzyMatcher        *FuzzyMatcher
	exactMatcher        *ExactMatcher
	abbreviationMatcher *AbbreviationMatcher
	aliasMatcher        *AliasMatcher
	config              NameMatchingConfig
	mu                  sync.RWMutex
}

// NameMatchingConfig holds configuration for name matching
type NameMatchingConfig struct {
	MaxFuzzyDistance    int     `json:"max_fuzzy_distance"`
	MinSimilarityScore  float64 `json:"min_similarity_score"`
	EnableAbbreviations bool    `json:"enable_abbreviations"`
	EnableAliases       bool    `json:"enable_aliases"`
	CaseSensitive       bool    `json:"case_sensitive"`
	IgnorePunctuation   bool    `json:"ignore_punctuation"`
}

// NameMatchResult represents the result of business name matching
type NameMatchResult struct {
	IsMatch         bool     `json:"is_match"`
	SimilarityScore float64  `json:"similarity_score"`
	MatchType       string   `json:"match_type"` // exact, fuzzy, abbreviation, alias
	MatchedName     string   `json:"matched_name"`
	OriginalName    string   `json:"original_name"`
	Confidence      float64  `json:"confidence"`
	FuzzyDistance   int      `json:"fuzzy_distance"`
	Abbreviations   []string `json:"abbreviations"`
	Aliases         []string `json:"aliases"`
	Warnings        []string `json:"warnings"`
}

// FuzzyMatcher provides fuzzy string matching capabilities
type FuzzyMatcher struct {
	algorithms map[string]FuzzyAlgorithm
	mu         sync.RWMutex
}

// FuzzyAlgorithm represents a fuzzy matching algorithm
type FuzzyAlgorithm struct {
	Name        string
	Description string
	Calculator  func(s1, s2 string) float64
}

// ExactMatcher provides exact string matching capabilities
type ExactMatcher struct {
	normalizers map[string]Normalizer
	mu          sync.RWMutex
}

// Normalizer represents a string normalization function
type Normalizer struct {
	Name        string
	Description string
	Normalize   func(s string) string
}

// AbbreviationMatcher matches business names with abbreviations
type AbbreviationMatcher struct {
	abbreviations map[string][]string
	patterns      []*regexp.Regexp
	mu            sync.RWMutex
}

// AliasMatcher matches business names with known aliases
type AliasMatcher struct {
	aliases map[string][]string
	mu      sync.RWMutex
}

// AddressVerifier verifies business addresses
type AddressVerifier struct {
	addressParser *AddressParser
	geocoder      *Geocoder
	standardizer  *AddressStandardizer
	config        AddressVerificationConfig
	mu            sync.RWMutex
}

// AddressVerificationConfig holds configuration for address verification
type AddressVerificationConfig struct {
	EnableGeocoding       bool    `json:"enable_geocoding"`
	EnableStandardization bool    `json:"enable_standardization"`
	MaxDistance           float64 `json:"max_distance_km"`
	MinConfidence         float64 `json:"min_confidence"`
}

// AddressMatchResult represents the result of address matching
type AddressMatchResult struct {
	IsMatch             bool      `json:"is_match"`
	SimilarityScore     float64   `json:"similarity_score"`
	Distance            float64   `json:"distance_km"`
	StandardizedAddress string    `json:"standardized_address"`
	GeocodedLocation    *Location `json:"geocoded_location"`
	Confidence          float64   `json:"confidence"`
	Warnings            []string  `json:"warnings"`
}

// AddressParser parses and normalizes addresses
type AddressParser struct {
	patterns []*regexp.Regexp
	mu       sync.RWMutex
}

// Geocoder provides geocoding capabilities
type Geocoder struct {
	providers map[string]GeocodingProvider
	mu        sync.RWMutex
}

// GeocodingProvider represents a geocoding service
type GeocodingProvider struct {
	Name     string
	Endpoint string
	APIKey   string
	Timeout  time.Duration
}

// AddressStandardizer standardizes address formats
type AddressStandardizer struct {
	formats map[string]AddressFormat
	mu      sync.RWMutex
}

// AddressFormat represents a standardized address format
type AddressFormat struct {
	Country    string
	Format     string
	Components []string
}

// Location represents a geographic location
type Location struct {
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Country    string  `json:"country"`
	State      string  `json:"state"`
	City       string  `json:"city"`
	PostalCode string  `json:"postal_code"`
}

// ContactVerifier verifies business contact information
type ContactVerifier struct {
	phoneVerifier *PhoneVerifier
	emailVerifier *EmailVerifier
	config        ContactVerificationConfig
	mu            sync.RWMutex
}

// ContactVerificationConfig holds configuration for contact verification
type ContactVerificationConfig struct {
	EnablePhoneVerification bool `json:"enable_phone_verification"`
	EnableEmailVerification bool `json:"enable_email_verification"`
	ValidateFormat          bool `json:"validate_format"`
	CheckExistence          bool `json:"check_existence"`
}

// ContactMatchResult represents the result of contact matching
type ContactMatchResult struct {
	IsMatch           bool     `json:"is_match"`
	PhoneMatch        bool     `json:"phone_match"`
	EmailMatch        bool     `json:"email_match"`
	PhoneConfidence   float64  `json:"phone_confidence"`
	EmailConfidence   float64  `json:"email_confidence"`
	OverallConfidence float64  `json:"overall_confidence"`
	Warnings          []string `json:"warnings"`
}

// PhoneVerifier verifies phone numbers
type PhoneVerifier struct {
	patterns []*regexp.Regexp
	mu       sync.RWMutex
}

// EmailVerifier verifies email addresses
type EmailVerifier struct {
	patterns []*regexp.Regexp
	mu       sync.RWMutex
}

// RegistrationChecker checks business registration data
type RegistrationChecker struct {
	databases map[string]RegistrationDatabase
	config    RegistrationCheckConfig
	mu        sync.RWMutex
}

// RegistrationCheckConfig holds configuration for registration checking
type RegistrationCheckConfig struct {
	EnableCrossReference bool `json:"enable_cross_reference"`
	CheckMultipleSources bool `json:"check_multiple_sources"`
	ValidateStatus       bool `json:"validate_status"`
}

// RegistrationMatchResult represents the result of registration matching
type RegistrationMatchResult struct {
	IsMatch           bool                  `json:"is_match"`
	RegistrationFound bool                  `json:"registration_found"`
	RegistrationData  *BusinessRegistration `json:"registration_data"`
	Confidence        float64               `json:"confidence"`
	Sources           []string              `json:"sources"`
	Warnings          []string              `json:"warnings"`
}

// RegistrationDatabase represents a business registration database
type RegistrationDatabase struct {
	Name     string
	Endpoint string
	APIKey   string
	Timeout  time.Duration
}

// BusinessRegistration represents business registration information
type BusinessRegistration struct {
	RegistrationNumber string    `json:"registration_number"`
	BusinessName       string    `json:"business_name"`
	LegalName          string    `json:"legal_name"`
	Status             string    `json:"status"`
	RegistrationDate   time.Time `json:"registration_date"`
	Address            string    `json:"address"`
	Phone              string    `json:"phone"`
	Email              string    `json:"email"`
	Website            string    `json:"website"`
	Industry           string    `json:"industry"`
	Source             string    `json:"source"`
}

// OwnershipScorer scores ownership evidence
type OwnershipScorer struct {
	evidenceTypes map[string]EvidenceType
	config        OwnershipScoringConfig
	mu            sync.RWMutex
}

// OwnershipScoringConfig holds configuration for ownership scoring
type OwnershipScoringConfig struct {
	EnableMultipleEvidence bool    `json:"enable_multiple_evidence"`
	MinEvidenceScore       float64 `json:"min_evidence_score"`
	WeightRecentEvidence   float64 `json:"weight_recent_evidence"`
}

// EvidenceType represents a type of ownership evidence
type EvidenceType struct {
	Name        string
	Weight      float64
	Description string
	Validator   func(evidence string) bool
}

// ConfidenceAssessor assesses overall connection confidence
type ConfidenceAssessor struct {
	weights map[string]float64
	config  ConfidenceAssessmentConfig
	mu      sync.RWMutex
}

// ConfidenceAssessmentConfig holds configuration for confidence assessment
type ConfidenceAssessmentConfig struct {
	EnableWeightedScoring   bool `json:"enable_weighted_scoring"`
	RequireMultipleEvidence bool `json:"require_multiple_evidence"`
	MinEvidenceCount        int  `json:"min_evidence_count"`
}

// ConnectionEvidence represents evidence of a business-website connection
type ConnectionEvidence struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Source      string                 `json:"source"`
	Timestamp   time.Time              `json:"timestamp"`
	Details     map[string]interface{} `json:"details"`
}

// NoClearConnectionDetector detects when there is no clear connection between business and website
type NoClearConnectionDetector struct {
	config NoClearConnectionConfig
	mu     sync.RWMutex
}

// NoClearConnectionConfig holds configuration for no clear connection detection
type NoClearConnectionConfig struct {
	MinConfidenceThreshold float64 `json:"min_confidence_threshold"`
	MaxNameSimilarity      float64 `json:"max_name_similarity"`
	MaxAddressSimilarity   float64 `json:"max_address_similarity"`
	MaxContactSimilarity   float64 `json:"max_contact_similarity"`
	RequireMultipleMatches bool    `json:"require_multiple_matches"`
	EnableHeuristics       bool    `json:"enable_heuristics"`
}

// NoClearConnectionResult represents the result of no clear connection detection
type NoClearConnectionResult struct {
	NoClearConnection bool     `json:"no_clear_connection"`
	Confidence        float64  `json:"confidence"`
	Reasons           []string `json:"reasons"`
	Evidence          []string `json:"evidence"`
	Recommendations   []string `json:"recommendations"`
	RiskLevel         string   `json:"risk_level"` // low, medium, high, critical
}

// ConnectionValidationDashboard provides dashboard and reporting capabilities for connection validation
type ConnectionValidationDashboard struct {
	results     []*ConnectionValidationResult
	statistics  *ConnectionValidationStatistics
	config      DashboardConfig
	mu          sync.RWMutex
}

// DashboardConfig holds configuration for the dashboard
type DashboardConfig struct {
	MaxResultsStored    int           `json:"max_results_stored"`
	StatisticsInterval  time.Duration `json:"statistics_interval"`
	EnableRealTimeStats bool          `json:"enable_real_time_stats"`
	EnableAlerts        bool          `json:"enable_alerts"`
	AlertThresholds     AlertThresholds `json:"alert_thresholds"`
}

// AlertThresholds defines thresholds for dashboard alerts
type AlertThresholds struct {
	LowConfidenceRate    float64 `json:"low_confidence_rate"`    // Alert if > X% of results have low confidence
	NoConnectionRate     float64 `json:"no_connection_rate"`     // Alert if > X% of results show no connection
	HighRiskRate         float64 `json:"high_risk_rate"`         // Alert if > X% of results are high risk
	AverageConfidence    float64 `json:"average_confidence"`     // Alert if average confidence < X
	ProcessingTime       float64 `json:"processing_time"`        // Alert if average processing time > X seconds
}

// ConnectionValidationStatistics holds statistical information about validation results
type ConnectionValidationStatistics struct {
	TotalValidations     int     `json:"total_validations"`
	SuccessfulConnections int     `json:"successful_connections"`
	FailedConnections    int     `json:"failed_connections"`
	NoClearConnections   int     `json:"no_clear_connections"`
	AverageConfidence    float64 `json:"average_confidence"`
	AverageProcessingTime float64 `json:"average_processing_time"`
	ConfidenceDistribution map[string]int `json:"confidence_distribution"`
	RiskLevelDistribution  map[string]int `json:"risk_level_distribution"`
	ConnectionStrengthDistribution map[string]int `json:"connection_strength_distribution"`
	LastUpdated          time.Time `json:"last_updated"`
}

// DashboardReport represents a comprehensive dashboard report
type DashboardReport struct {
	Statistics          *ConnectionValidationStatistics `json:"statistics"`
	RecentResults       []*ConnectionValidationResult   `json:"recent_results"`
	Alerts              []DashboardAlert                `json:"alerts"`
	Trends              []TrendData                     `json:"trends"`
	Recommendations     []string                        `json:"recommendations"`
	GeneratedAt         time.Time                       `json:"generated_at"`
}

// DashboardAlert represents an alert from the dashboard
type DashboardAlert struct {
	Type        string    `json:"type"`        // confidence, connection, risk, performance
	Severity    string    `json:"severity"`    // low, medium, high, critical
	Message     string    `json:"message"`
	Threshold   float64   `json:"threshold"`
	CurrentValue float64  `json:"current_value"`
	Timestamp   time.Time `json:"timestamp"`
}

// TrendData represents trend information for the dashboard
type TrendData struct {
	Metric      string    `json:"metric"`
	Value       float64   `json:"value"`
	Change      float64   `json:"change"`      // Percentage change from previous period
	Direction   string    `json:"direction"`   // increasing, decreasing, stable
	Timestamp   time.Time `json:"timestamp"`
}

// NewConnectionValidator creates a new connection validator
func NewConnectionValidator() *ConnectionValidator {
	config := ConnectionValidatorConfig{
		EnableNameMatching:        true,
		EnableAddressVerification: true,
		EnableContactVerification: true,
		EnableRegistrationCheck:   true,
		EnableOwnershipScoring:    true,
		MinConfidenceScore:        0.7,
		MaxFuzzyDistance:          3,
		EnableCrossReference:      true,
		Timeout:                   time.Second * 30,
	}

	return &ConnectionValidator{
		nameMatcher:         NewBusinessNameMatcher(),
		addressVerifier:     NewAddressVerifier(),
		contactVerifier:     NewContactVerifier(),
		registrationChecker: NewRegistrationChecker(),
		ownershipScorer:     NewOwnershipScorer(),
		confidenceAssessor:  NewConfidenceAssessor(),
		config:              config,
	}
}

// NewBusinessNameMatcher creates a new business name matcher
func NewBusinessNameMatcher() *BusinessNameMatcher {
	nameConfig := NameMatchingConfig{
		MaxFuzzyDistance:    3,
		MinSimilarityScore:  0.8,
		EnableAbbreviations: true,
		EnableAliases:       true,
		CaseSensitive:       false,
		IgnorePunctuation:   true,
	}

	return &BusinessNameMatcher{
		fuzzyMatcher:        NewFuzzyMatcher(),
		exactMatcher:        NewExactMatcher(),
		abbreviationMatcher: NewAbbreviationMatcher(),
		aliasMatcher:        NewAliasMatcher(),
		config:              nameConfig,
	}
}

// NewFuzzyMatcher creates a new fuzzy matcher
func NewFuzzyMatcher() *FuzzyMatcher {
	return &FuzzyMatcher{
		algorithms: map[string]FuzzyAlgorithm{
			"levenshtein": {
				Name:        "Levenshtein Distance",
				Description: "Edit distance algorithm",
				Calculator:  calculateLevenshteinDistance,
			},
			"jaro_winkler": {
				Name:        "Jaro-Winkler Similarity",
				Description: "String similarity algorithm",
				Calculator:  calculateJaroWinklerSimilarity,
			},
			"cosine": {
				Name:        "Cosine Similarity",
				Description: "Vector-based similarity",
				Calculator:  calculateCosineSimilarity,
			},
		},
	}
}

// NewExactMatcher creates a new exact matcher
func NewExactMatcher() *ExactMatcher {
	return &ExactMatcher{
		normalizers: map[string]Normalizer{
			"case": {
				Name:        "Case Normalization",
				Description: "Convert to lowercase",
				Normalize:   strings.ToLower,
			},
			"punctuation": {
				Name:        "Punctuation Removal",
				Description: "Remove punctuation marks",
				Normalize:   removePunctuation,
			},
			"whitespace": {
				Name:        "Whitespace Normalization",
				Description: "Normalize whitespace",
				Normalize:   normalizeWhitespace,
			},
		},
	}
}

// NewAbbreviationMatcher creates a new abbreviation matcher
func NewAbbreviationMatcher() *AbbreviationMatcher {
	return &AbbreviationMatcher{
		abbreviations: map[string][]string{
			"corporation":  {"corp", "corp.", "corporation"},
			"incorporated": {"inc", "inc.", "incorporated"},
			"limited":      {"ltd", "ltd.", "limited"},
			"company":      {"co", "co.", "company"},
			"associates":   {"assoc", "assoc.", "associates"},
			"partners":     {"partners", "partnership"},
		},
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(corp|inc|ltd|co|assoc)\b`),
		},
	}
}

// NewAliasMatcher creates a new alias matcher
func NewAliasMatcher() *AliasMatcher {
	return &AliasMatcher{
		aliases: map[string][]string{
			"microsoft": {"ms", "microsoft corporation"},
			"apple":     {"apple inc", "apple computer"},
			"google":    {"alphabet", "google llc"},
		},
	}
}

// NewAddressVerifier creates a new address verifier
func NewAddressVerifier() *AddressVerifier {
	addressConfig := AddressVerificationConfig{
		EnableGeocoding:       true,
		EnableStandardization: true,
		MaxDistance:           10.0, // 10 km
		MinConfidence:         0.8,
	}

	return &AddressVerifier{
		addressParser: NewAddressParser(),
		geocoder:      NewGeocoder(),
		standardizer:  NewAddressStandardizer(),
		config:        addressConfig,
	}
}

// NewAddressParser creates a new address parser
func NewAddressParser() *AddressParser {
	return &AddressParser{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(\d+)\s+([a-z\s]+),\s*([a-z\s]+),\s*([a-z]{2})\s*(\d{5}(?:-\d{4})?)`),
		},
	}
}

// NewGeocoder creates a new geocoder
func NewGeocoder() *Geocoder {
	return &Geocoder{
		providers: map[string]GeocodingProvider{
			"nominatim": {
				Name:     "OpenStreetMap Nominatim",
				Endpoint: "https://nominatim.openstreetmap.org/search",
				Timeout:  time.Second * 10,
			},
		},
	}
}

// NewAddressStandardizer creates a new address standardizer
func NewAddressStandardizer() *AddressStandardizer {
	return &AddressStandardizer{
		formats: map[string]AddressFormat{
			"us": {
				Country:    "United States",
				Format:     "{street_number} {street_name}, {city}, {state} {postal_code}",
				Components: []string{"street_number", "street_name", "city", "state", "postal_code"},
			},
		},
	}
}

// NewContactVerifier creates a new contact verifier
func NewContactVerifier() *ContactVerifier {
	contactConfig := ContactVerificationConfig{
		EnablePhoneVerification: true,
		EnableEmailVerification: true,
		ValidateFormat:          true,
		CheckExistence:          false, // Would require external API calls
	}

	return &ContactVerifier{
		phoneVerifier: NewPhoneVerifier(),
		emailVerifier: NewEmailVerifier(),
		config:        contactConfig,
	}
}

// NewPhoneVerifier creates a new phone verifier
func NewPhoneVerifier() *PhoneVerifier {
	return &PhoneVerifier{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^\+?1?\s*\(?([0-9]{3})\)?[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})$`),
		},
	}
}

// NewEmailVerifier creates a new email verifier
func NewEmailVerifier() *EmailVerifier {
	return &EmailVerifier{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		},
	}
}

// NewRegistrationChecker creates a new registration checker
func NewRegistrationChecker() *RegistrationChecker {
	regConfig := RegistrationCheckConfig{
		EnableCrossReference: true,
		CheckMultipleSources: true,
		ValidateStatus:       true,
	}

	return &RegistrationChecker{
		databases: map[string]RegistrationDatabase{
			"open_corporates": {
				Name:     "OpenCorporates",
				Endpoint: "https://api.opencorporates.com",
				Timeout:  time.Second * 30,
			},
		},
		config: regConfig,
	}
}

// NewOwnershipScorer creates a new ownership scorer
func NewOwnershipScorer() *OwnershipScorer {
	ownershipConfig := OwnershipScoringConfig{
		EnableMultipleEvidence: true,
		MinEvidenceScore:       0.6,
		WeightRecentEvidence:   1.2,
	}

	return &OwnershipScorer{
		evidenceTypes: map[string]EvidenceType{
			"domain_registration": {
				Name:        "Domain Registration",
				Weight:      0.8,
				Description: "Domain registered to business",
				Validator:   validateDomainRegistration,
			},
			"ssl_certificate": {
				Name:        "SSL Certificate",
				Weight:      0.6,
				Description: "SSL certificate issued to business",
				Validator:   validateSSLCertificate,
			},
			"contact_information": {
				Name:        "Contact Information",
				Weight:      0.7,
				Description: "Contact info matches business",
				Validator:   validateContactInformation,
			},
			"address_match": {
				Name:        "Address Match",
				Weight:      0.9,
				Description: "Address matches business registration",
				Validator:   validateAddressMatch,
			},
		},
		config: ownershipConfig,
	}
}

// NewConfidenceAssessor creates a new confidence assessor
func NewConfidenceAssessor() *ConfidenceAssessor {
	confidenceConfig := ConfidenceAssessmentConfig{
		EnableWeightedScoring:   true,
		RequireMultipleEvidence: true,
		MinEvidenceCount:        2,
	}

	return &ConfidenceAssessor{
		weights: map[string]float64{
			"name_match":         0.3,
			"address_match":      0.25,
			"contact_match":      0.2,
			"registration_match": 0.15,
			"ownership_score":    0.1,
		},
		config: confidenceConfig,
	}
}

// NewNoClearConnectionDetector creates a new no clear connection detector
func NewNoClearConnectionDetector(config NoClearConnectionConfig) *NoClearConnectionDetector {
	return &NoClearConnectionDetector{
		config: config,
	}
}

// ValidateConnection performs comprehensive business-website connection validation
func (cv *ConnectionValidator) ValidateConnection(businessName, websiteURL string) (*ConnectionValidationResult, error) {
	start := time.Now()

	result := &ConnectionValidationResult{
		BusinessName: businessName,
		WebsiteURL:   websiteURL,
		LastChecked:  time.Now(),
		Evidence:     []ConnectionEvidence{},
		Warnings:     []string{},
		Errors:       []string{},
	}

	// Perform name matching
	if cv.config.EnableNameMatching {
		nameResult, err := cv.nameMatcher.MatchName(businessName, websiteURL)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Name matching failed: %v", err))
		} else {
			result.NameMatchResult = nameResult
			if nameResult.IsMatch {
				result.Evidence = append(result.Evidence, ConnectionEvidence{
					Type:        "name_match",
					Description: "Business name matches website content",
					Confidence:  nameResult.Confidence,
					Source:      "name_matcher",
					Timestamp:   time.Now(),
				})
			}
		}
	}

	// Perform address verification
	if cv.config.EnableAddressVerification {
		addressResult, err := cv.addressVerifier.VerifyAddress(businessName, websiteURL)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Address verification failed: %v", err))
		} else {
			result.AddressMatchResult = addressResult
			if addressResult.IsMatch {
				result.Evidence = append(result.Evidence, ConnectionEvidence{
					Type:        "address_match",
					Description: "Business address matches website",
					Confidence:  addressResult.Confidence,
					Source:      "address_verifier",
					Timestamp:   time.Now(),
				})
			}
		}
	}

	// Perform contact verification
	if cv.config.EnableContactVerification {
		contactResult, err := cv.contactVerifier.VerifyContact(businessName, websiteURL)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Contact verification failed: %v", err))
		} else {
			result.ContactMatchResult = contactResult
			if contactResult.IsMatch {
				result.Evidence = append(result.Evidence, ConnectionEvidence{
					Type:        "contact_match",
					Description: "Contact information matches business",
					Confidence:  contactResult.OverallConfidence,
					Source:      "contact_verifier",
					Timestamp:   time.Now(),
				})
			}
		}
	}

	// Perform registration check
	if cv.config.EnableRegistrationCheck {
		registrationResult, err := cv.registrationChecker.CheckRegistration(businessName, websiteURL)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Registration check failed: %v", err))
		} else {
			result.RegistrationMatchResult = registrationResult
			if registrationResult.IsMatch {
				result.Evidence = append(result.Evidence, ConnectionEvidence{
					Type:        "registration_match",
					Description: "Business registration matches website",
					Confidence:  registrationResult.Confidence,
					Source:      "registration_checker",
					Timestamp:   time.Now(),
				})
			}
		}
	}

	// Calculate ownership score
	if cv.config.EnableOwnershipScoring {
		ownershipScore, err := cv.ownershipScorer.ScoreOwnership(businessName, websiteURL)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Ownership scoring failed: %v", err))
		} else {
			result.OwnershipScore = ownershipScore
			if ownershipScore > 0.7 {
				result.Evidence = append(result.Evidence, ConnectionEvidence{
					Type:        "ownership_evidence",
					Description: "Strong ownership evidence found",
					Confidence:  ownershipScore,
					Source:      "ownership_scorer",
					Timestamp:   time.Now(),
				})
			}
		}
	}

	// Calculate overall confidence
	result.OverallConfidence = cv.confidenceAssessor.AssessConfidence(result)
	result.IsConnected = result.OverallConfidence >= cv.config.MinConfidenceScore
	result.ConnectionStrength = cv.determineConnectionStrength(result.OverallConfidence)
	result.ValidationTime = time.Since(start)

	return result, nil
}

// MatchName matches business names using various algorithms
func (bnm *BusinessNameMatcher) MatchName(businessName, websiteURL string) (*NameMatchResult, error) {
	bnm.mu.RLock()
	defer bnm.mu.RUnlock()

	result := &NameMatchResult{
		OriginalName: businessName,
		Warnings:     []string{},
	}

	// Try exact matching first
	if exactMatch := bnm.exactMatcher.Match(businessName, websiteURL); exactMatch {
		result.IsMatch = true
		result.MatchType = "exact"
		result.SimilarityScore = 1.0
		result.Confidence = 1.0
		return result, nil
	}

	// Try fuzzy matching
	if fuzzyScore := bnm.fuzzyMatcher.Match(businessName, websiteURL); fuzzyScore > bnm.config.MinSimilarityScore {
		result.IsMatch = true
		result.MatchType = "fuzzy"
		result.SimilarityScore = fuzzyScore
		result.Confidence = fuzzyScore
		return result, nil
	}

	// Try abbreviation matching
	if abbrevMatch := bnm.abbreviationMatcher.Match(businessName, websiteURL); abbrevMatch {
		result.IsMatch = true
		result.MatchType = "abbreviation"
		result.SimilarityScore = 0.9
		result.Confidence = 0.9
		return result, nil
	}

	// Try alias matching
	if aliasMatch := bnm.aliasMatcher.Match(businessName, websiteURL); aliasMatch {
		result.IsMatch = true
		result.MatchType = "alias"
		result.SimilarityScore = 0.8
		result.Confidence = 0.8
		return result, nil
	}

	// No match found
	result.IsMatch = false
	result.SimilarityScore = 0.0
	result.Confidence = 0.0
	return result, nil
}

// Match performs exact string matching
func (em *ExactMatcher) Match(s1, s2 string) bool {
	em.mu.RLock()
	defer em.mu.RUnlock()

	// Apply all normalizers
	normalized1 := s1
	normalized2 := s2

	for _, normalizer := range em.normalizers {
		normalized1 = normalizer.Normalize(normalized1)
		normalized2 = normalizer.Normalize(normalized2)
	}

	return normalized1 == normalized2
}

// Match performs fuzzy string matching
func (fm *FuzzyMatcher) Match(s1, s2 string) float64 {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	// Use Jaro-Winkler similarity as default
	if algorithm, exists := fm.algorithms["jaro_winkler"]; exists {
		return algorithm.Calculator(s1, s2)
	}

	return 0.0
}

// Match performs abbreviation matching
func (am *AbbreviationMatcher) Match(businessName, websiteURL string) bool {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// This is a simplified implementation
	// In a real implementation, this would extract business names from the website
	// and check for abbreviation matches

	for _, pattern := range am.patterns {
		if pattern.MatchString(businessName) {
			return true
		}
	}

	return false
}

// Match performs alias matching
func (alm *AliasMatcher) Match(businessName, websiteURL string) bool {
	alm.mu.RLock()
	defer alm.mu.RUnlock()

	// This is a simplified implementation
	// In a real implementation, this would check against known aliases

	normalizedBusinessName := strings.ToLower(businessName)
	if aliases, exists := alm.aliases[normalizedBusinessName]; exists {
		// Check if any alias appears in the website URL or content
		for _, alias := range aliases {
			if strings.Contains(strings.ToLower(websiteURL), strings.ToLower(alias)) {
				return true
			}
		}
	}

	return false
}

// VerifyAddress verifies business addresses
func (av *AddressVerifier) VerifyAddress(businessName, websiteURL string) (*AddressMatchResult, error) {
	av.mu.RLock()
	defer av.mu.RUnlock()

	result := &AddressMatchResult{
		Warnings: []string{},
	}

	// This is a simplified implementation
	// In a real implementation, this would:
	// 1. Extract addresses from the website
	// 2. Parse and standardize addresses
	// 3. Geocode addresses
	// 4. Compare distances

	// Simulate address verification
	result.IsMatch = false
	result.SimilarityScore = 0.0
	result.Confidence = 0.0

	return result, nil
}

// VerifyContact verifies business contact information
func (cv *ContactVerifier) VerifyContact(businessName, websiteURL string) (*ContactMatchResult, error) {
	cv.mu.RLock()
	defer cv.mu.RUnlock()

	result := &ContactMatchResult{
		Warnings: []string{},
	}

	// This is a simplified implementation
	// In a real implementation, this would:
	// 1. Extract phone numbers and emails from the website
	// 2. Validate formats
	// 3. Compare with business registration data

	// Simulate contact verification
	result.IsMatch = false
	result.PhoneMatch = false
	result.EmailMatch = false
	result.PhoneConfidence = 0.0
	result.EmailConfidence = 0.0
	result.OverallConfidence = 0.0

	return result, nil
}

// CheckRegistration checks business registration data
func (rc *RegistrationChecker) CheckRegistration(businessName, websiteURL string) (*RegistrationMatchResult, error) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	result := &RegistrationMatchResult{
		Warnings: []string{},
	}

	// This is a simplified implementation
	// In a real implementation, this would:
	// 1. Query business registration databases
	// 2. Cross-reference with website information
	// 3. Validate registration status

	// Simulate registration check
	result.IsMatch = false
	result.RegistrationFound = false
	result.Confidence = 0.0

	return result, nil
}

// ScoreOwnership scores ownership evidence
func (os *OwnershipScorer) ScoreOwnership(businessName, websiteURL string) (float64, error) {
	os.mu.RLock()
	defer os.mu.RUnlock()

	// This is a simplified implementation
	// In a real implementation, this would:
	// 1. Check domain registration
	// 2. Verify SSL certificates
	// 3. Validate contact information
	// 4. Check address matches

	// Simulate ownership scoring
	return 0.0, nil
}

// AssessConfidence assesses overall connection confidence
func (ca *ConfidenceAssessor) AssessConfidence(result *ConnectionValidationResult) float64 {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	if !ca.config.EnableWeightedScoring {
		// Simple average of available scores
		scores := []float64{}

		if result.NameMatchResult != nil && result.NameMatchResult.IsMatch {
			scores = append(scores, result.NameMatchResult.Confidence)
		}
		if result.AddressMatchResult != nil && result.AddressMatchResult.IsMatch {
			scores = append(scores, result.AddressMatchResult.Confidence)
		}
		if result.ContactMatchResult != nil && result.ContactMatchResult.IsMatch {
			scores = append(scores, result.ContactMatchResult.OverallConfidence)
		}
		if result.RegistrationMatchResult != nil && result.RegistrationMatchResult.IsMatch {
			scores = append(scores, result.RegistrationMatchResult.Confidence)
		}
		if result.OwnershipScore > 0 {
			scores = append(scores, result.OwnershipScore)
		}

		if len(scores) == 0 {
			return 0.0
		}

		total := 0.0
		for _, score := range scores {
			total += score
		}
		return total / float64(len(scores))
	}

	// Weighted scoring
	totalScore := 0.0
	totalWeight := 0.0

	if result.NameMatchResult != nil && result.NameMatchResult.IsMatch {
		weight := ca.weights["name_match"]
		totalScore += result.NameMatchResult.Confidence * weight
		totalWeight += weight
	}

	if result.AddressMatchResult != nil && result.AddressMatchResult.IsMatch {
		weight := ca.weights["address_match"]
		totalScore += result.AddressMatchResult.Confidence * weight
		totalWeight += weight
	}

	if result.ContactMatchResult != nil && result.ContactMatchResult.IsMatch {
		weight := ca.weights["contact_match"]
		totalScore += result.ContactMatchResult.OverallConfidence * weight
		totalWeight += weight
	}

	if result.RegistrationMatchResult != nil && result.RegistrationMatchResult.IsMatch {
		weight := ca.weights["registration_match"]
		totalScore += result.RegistrationMatchResult.Confidence * weight
		totalWeight += weight
	}

	if result.OwnershipScore > 0 {
		weight := ca.weights["ownership_score"]
		totalScore += result.OwnershipScore * weight
		totalWeight += weight
	}

	if totalWeight > 0 {
		return totalScore / totalWeight
	}

	return 0.0
}

// Helper method for ConnectionValidator
func (cv *ConnectionValidator) determineConnectionStrength(confidence float64) string {
	if confidence >= 0.9 {
		return "strong"
	} else if confidence >= 0.7 {
		return "moderate"
	} else if confidence >= 0.5 {
		return "weak"
	} else {
		return "none"
	}
}

// Fuzzy matching algorithms
func calculateLevenshteinDistance(s1, s2 string) float64 {
	// Calculate Levenshtein distance between two strings
	len1, len2 := len(s1), len(s2)
	
	// Create a 2D slice to store distances
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}
	
	// Initialize first row and column
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}
	
	// Fill the matrix
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,    // deletion
				min(
					matrix[i][j-1]+1, // insertion
					matrix[i-1][j-1]+cost, // substitution
				),
			)
		}
	}
	
	distance := matrix[len1][len2]
	maxLen := max(len1, len2)
	if maxLen == 0 {
		return 1.0 // Both strings are empty, perfect match
	}
	
	// Return similarity score (1 - normalized distance)
	return 1.0 - float64(distance)/float64(maxLen)
}

func calculateJaroWinklerSimilarity(s1, s2 string) float64 {
	// Calculate Jaro-Winkler similarity
	if s1 == s2 {
		return 1.0
	}
	
	len1, len2 := len(s1), len(s2)
	if len1 == 0 || len2 == 0 {
		return 0.0
	}
	
	// Calculate matching window
	matchWindow := max(len1, len2)/2 - 1
	if matchWindow < 0 {
		matchWindow = 0
	}
	
	// Find matching characters
	s1Matches := make([]bool, len1)
	s2Matches := make([]bool, len2)
	matches := 0
	
	for i := 0; i < len1; i++ {
		start := max(0, i-matchWindow)
		end := min(len2, i+matchWindow+1)
		
		for j := start; j < end; j++ {
			if !s2Matches[j] && s1[i] == s2[j] {
				s1Matches[i] = true
				s2Matches[j] = true
				matches++
				break
			}
		}
	}
	
	if matches == 0 {
		return 0.0
	}
	
	// Calculate transpositions
	transpositions := 0
	k := 0
	for i := 0; i < len1; i++ {
		if s1Matches[i] {
			for !s2Matches[k] {
				k++
			}
			if s1[i] != s2[k] {
				transpositions++
			}
			k++
		}
	}
	
	// Calculate Jaro similarity
	jaro := (float64(matches)/float64(len1) + 
		float64(matches)/float64(len2) + 
		float64(matches-transpositions/2)/float64(matches)) / 3.0
	
	// Calculate Jaro-Winkler similarity
	prefix := 0
	for i := 0; i < min(4, min(len1, len2)); i++ {
		if s1[i] == s2[i] {
			prefix++
		} else {
			break
		}
	}
	
	winkler := jaro + 0.1*float64(prefix)*(1.0-jaro)
	if winkler > 1.0 {
		return 1.0
	}
	return winkler
}

func calculateCosineSimilarity(s1, s2 string) float64 {
	// Calculate cosine similarity between two strings
	// Convert strings to character frequency vectors
	freq1 := make(map[rune]int)
	freq2 := make(map[rune]int)
	
	for _, r := range strings.ToLower(s1) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			freq1[r]++
		}
	}
	
	for _, r := range strings.ToLower(s2) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			freq2[r]++
		}
	}
	
	// Calculate dot product and magnitudes
	dotProduct := 0.0
	mag1 := 0.0
	mag2 := 0.0
	
	// Get all unique characters
	allChars := make(map[rune]bool)
	for r := range freq1 {
		allChars[r] = true
	}
	for r := range freq2 {
		allChars[r] = true
	}
	
	for r := range allChars {
		count1 := freq1[r]
		count2 := freq2[r]
		dotProduct += float64(count1 * count2)
		mag1 += float64(count1 * count1)
		mag2 += float64(count2 * count2)
	}
	
	mag1 = sqrt(mag1)
	mag2 = sqrt(mag2)
	
	if mag1 == 0 || mag2 == 0 {
		return 0.0
	}
	
	return dotProduct / (mag1 * mag2)
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func sqrt(x float64) float64 {
	return float64(int(x*1000)) / 1000 // Simplified square root for performance
}

// String normalization functions
func removePunctuation(s string) string {
	var result strings.Builder
	for _, r := range s {
		if !unicode.IsPunct(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func normalizeWhitespace(s string) string {
	// Replace multiple whitespace characters with single space
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

// Evidence validation functions
func validateDomainRegistration(evidence string) bool {
	// Simplified domain registration validation
	return false
}

func validateSSLCertificate(evidence string) bool {
	// Simplified SSL certificate validation
	return false
}

func validateContactInformation(evidence string) bool {
	// Simplified contact information validation
	return false
}

func validateAddressMatch(evidence string) bool {
	// Simplified address match validation
	return false
}

// DetectNoClearConnection analyzes connection validation results to determine if there's no clear connection
func (nccd *NoClearConnectionDetector) DetectNoClearConnection(result *ConnectionValidationResult) *NoClearConnectionResult {
	nccd.mu.RLock()
	defer nccd.mu.RUnlock()
	
	detectionResult := &NoClearConnectionResult{
		Reasons:         []string{},
		Evidence:        []string{},
		Recommendations: []string{},
	}
	
	// Check overall confidence
	if result.OverallConfidence < nccd.config.MinConfidenceThreshold {
		detectionResult.Reasons = append(detectionResult.Reasons, 
			fmt.Sprintf("Overall confidence (%.2f) below threshold (%.2f)", 
				result.OverallConfidence, nccd.config.MinConfidenceThreshold))
		detectionResult.Evidence = append(detectionResult.Evidence, 
			fmt.Sprintf("Low confidence score: %.2f", result.OverallConfidence))
	}
	
	// Check name matching
	if result.NameMatchResult != nil {
		if result.NameMatchResult.SimilarityScore < nccd.config.MaxNameSimilarity {
			detectionResult.Reasons = append(detectionResult.Reasons,
				fmt.Sprintf("Business name similarity (%.2f) below threshold (%.2f)",
					result.NameMatchResult.SimilarityScore, nccd.config.MaxNameSimilarity))
			detectionResult.Evidence = append(detectionResult.Evidence,
				fmt.Sprintf("Business name: '%s' vs Website: '%s'", 
					result.NameMatchResult.OriginalName, result.NameMatchResult.MatchedName))
		}
	} else {
		detectionResult.Reasons = append(detectionResult.Reasons, "No business name matching performed")
		detectionResult.Evidence = append(detectionResult.Evidence, "Missing name match result")
	}
	
	// Check address matching
	if result.AddressMatchResult != nil {
		if result.AddressMatchResult.Confidence < nccd.config.MaxAddressSimilarity {
			detectionResult.Reasons = append(detectionResult.Reasons,
				fmt.Sprintf("Address similarity (%.2f) below threshold (%.2f)",
					result.AddressMatchResult.Confidence, nccd.config.MaxAddressSimilarity))
			detectionResult.Evidence = append(detectionResult.Evidence,
				fmt.Sprintf("Address match confidence: %.2f", result.AddressMatchResult.Confidence))
		}
	} else {
		detectionResult.Reasons = append(detectionResult.Reasons, "No address matching performed")
		detectionResult.Evidence = append(detectionResult.Evidence, "Missing address match result")
	}
	
	// Check contact matching
	if result.ContactMatchResult != nil {
		if result.ContactMatchResult.OverallConfidence < nccd.config.MaxContactSimilarity {
			detectionResult.Reasons = append(detectionResult.Reasons,
				fmt.Sprintf("Contact similarity (%.2f) below threshold (%.2f)",
					result.ContactMatchResult.OverallConfidence, nccd.config.MaxContactSimilarity))
			detectionResult.Evidence = append(detectionResult.Evidence,
				fmt.Sprintf("Contact match confidence: %.2f", result.ContactMatchResult.OverallConfidence))
		}
	} else {
		detectionResult.Reasons = append(detectionResult.Reasons, "No contact matching performed")
		detectionResult.Evidence = append(detectionResult.Evidence, "Missing contact match result")
	}
	
	// Check if multiple matches are required
	if nccd.config.RequireMultipleMatches {
		matchCount := 0
		if result.NameMatchResult != nil && result.NameMatchResult.IsMatch {
			matchCount++
		}
		if result.AddressMatchResult != nil && result.AddressMatchResult.IsMatch {
			matchCount++
		}
		if result.ContactMatchResult != nil && result.ContactMatchResult.IsMatch {
			matchCount++
		}
		if result.RegistrationMatchResult != nil && result.RegistrationMatchResult.IsMatch {
			matchCount++
		}
		
		if matchCount < 2 {
			detectionResult.Reasons = append(detectionResult.Reasons,
				fmt.Sprintf("Only %d matching criteria met, minimum 2 required", matchCount))
			detectionResult.Evidence = append(detectionResult.Evidence,
				fmt.Sprintf("Match count: %d", matchCount))
		}
	}
	
	// Apply heuristics if enabled
	if nccd.config.EnableHeuristics {
		nccd.applyHeuristics(result, detectionResult)
	}
	
	// Determine if there's no clear connection
	detectionResult.NoClearConnection = len(detectionResult.Reasons) > 0
	
	// Calculate confidence based on number and severity of reasons
	detectionResult.Confidence = nccd.calculateDetectionConfidence(detectionResult.Reasons, result.OverallConfidence)
	
	// Determine risk level
	detectionResult.RiskLevel = nccd.determineRiskLevel(detectionResult.Confidence, len(detectionResult.Reasons))
	
	// Generate recommendations
	detectionResult.Recommendations = nccd.generateRecommendations(detectionResult)
	
	return detectionResult
}

// applyHeuristics applies additional heuristics to detect no clear connection
func (nccd *NoClearConnectionDetector) applyHeuristics(result *ConnectionValidationResult, detectionResult *NoClearConnectionResult) {
	// Check for suspicious patterns
	if result.WebsiteURL != "" {
		// Check for generic domain names
		genericDomains := []string{"example.com", "test.com", "placeholder.com", "demo.com"}
		for _, domain := range genericDomains {
			if strings.Contains(result.WebsiteURL, domain) {
				detectionResult.Reasons = append(detectionResult.Reasons, 
					"Website uses generic/placeholder domain")
				detectionResult.Evidence = append(detectionResult.Evidence,
					fmt.Sprintf("Generic domain detected: %s", domain))
				break
			}
		}
		
		// Check for very short domain names (potential fake sites)
		urlParts := strings.Split(result.WebsiteURL, "/")
		if len(urlParts) > 2 {
			domain := urlParts[2]
			if len(domain) < 5 {
				detectionResult.Reasons = append(detectionResult.Reasons,
					"Website domain is suspiciously short")
				detectionResult.Evidence = append(detectionResult.Evidence,
					fmt.Sprintf("Short domain: %s", domain))
			}
		}
	}
	
	// Check for missing critical information
	if result.NameMatchResult == nil || !result.NameMatchResult.IsMatch {
		detectionResult.Reasons = append(detectionResult.Reasons,
			"Business name not found on website")
		detectionResult.Evidence = append(detectionResult.Evidence,
			"No business name match detected")
	}
	
	// Check for low ownership score
	if result.OwnershipScore < 0.3 {
		detectionResult.Reasons = append(detectionResult.Reasons,
			fmt.Sprintf("Low ownership score (%.2f) indicates weak connection", result.OwnershipScore))
		detectionResult.Evidence = append(detectionResult.Evidence,
			fmt.Sprintf("Ownership score: %.2f", result.OwnershipScore))
	}
}

// calculateDetectionConfidence calculates confidence in the no clear connection detection
func (nccd *NoClearConnectionDetector) calculateDetectionConfidence(reasons []string, overallConfidence float64) float64 {
	// Base confidence on number of reasons and overall confidence
	reasonWeight := float64(len(reasons)) * 0.2
	confidenceWeight := (1.0 - overallConfidence) * 0.8
	
	totalConfidence := reasonWeight + confidenceWeight
	
	// Cap at 1.0
	if totalConfidence > 1.0 {
		return 1.0
	}
	
	return totalConfidence
}

// determineRiskLevel determines the risk level based on confidence and number of reasons
func (nccd *NoClearConnectionDetector) determineRiskLevel(confidence float64, reasonCount int) string {
	if confidence >= 0.9 || reasonCount >= 5 {
		return "critical"
	} else if confidence >= 0.7 || reasonCount >= 3 {
		return "high"
	} else if confidence >= 0.5 || reasonCount >= 2 {
		return "medium"
	} else {
		return "low"
	}
}

// generateRecommendations generates recommendations based on detection results
func (nccd *NoClearConnectionDetector) generateRecommendations(result *NoClearConnectionResult) []string {
	recommendations := []string{}
	
	if result.NoClearConnection {
		recommendations = append(recommendations, 
			"Manual review required - insufficient evidence of business-website connection")
		
		if result.RiskLevel == "critical" || result.RiskLevel == "high" {
			recommendations = append(recommendations,
				"High risk detected - consider additional verification steps")
		}
		
		// Add specific recommendations based on reasons
		for _, reason := range result.Reasons {
			if strings.Contains(reason, "name similarity") {
				recommendations = append(recommendations,
					"Verify business name spelling and variations")
			}
			if strings.Contains(reason, "address") {
				recommendations = append(recommendations,
					"Verify business address and location information")
			}
			if strings.Contains(reason, "contact") {
				recommendations = append(recommendations,
					"Verify contact information and phone numbers")
			}
			if strings.Contains(reason, "confidence") {
				recommendations = append(recommendations,
					"Review all connection evidence manually")
			}
		}
	} else {
		recommendations = append(recommendations,
			"Connection appears valid - proceed with standard verification")
	}
	
	return recommendations
}

// NewConnectionValidationDashboard creates a new connection validation dashboard
func NewConnectionValidationDashboard(config DashboardConfig) *ConnectionValidationDashboard {
	dashboard := &ConnectionValidationDashboard{
		results: []*ConnectionValidationResult{},
		statistics: &ConnectionValidationStatistics{
			ConfidenceDistribution: make(map[string]int),
			RiskLevelDistribution:  make(map[string]int),
			ConnectionStrengthDistribution: make(map[string]int),
		},
		config: config,
	}
	
	// Start statistics update if real-time stats are enabled
	if config.EnableRealTimeStats {
		go dashboard.updateStatisticsPeriodically()
	}
	
	return dashboard
}

// AddResult adds a validation result to the dashboard
func (cvd *ConnectionValidationDashboard) AddResult(result *ConnectionValidationResult) {
	cvd.mu.Lock()
	defer cvd.mu.Unlock()
	
	// Add result to the list
	cvd.results = append(cvd.results, result)
	
	// Maintain max results limit
	if len(cvd.results) > cvd.config.MaxResultsStored {
		cvd.results = cvd.results[1:] // Remove oldest result
	}
	
	// Update statistics
	cvd.updateStatistics()
	
	// Check for alerts if enabled
	if cvd.config.EnableAlerts {
		cvd.checkAlerts()
	}
}

// GetDashboardReport generates a comprehensive dashboard report
func (cvd *ConnectionValidationDashboard) GetDashboardReport() *DashboardReport {
	cvd.mu.RLock()
	defer cvd.mu.RUnlock()
	
	report := &DashboardReport{
		Statistics:      cvd.statistics,
		RecentResults:   cvd.getRecentResults(10), // Last 10 results
		Alerts:          cvd.getActiveAlerts(),
		Trends:          cvd.calculateTrends(),
		Recommendations: cvd.generateRecommendations(),
		GeneratedAt:     time.Now(),
	}
	
	return report
}

// GetStatistics returns current statistics
func (cvd *ConnectionValidationDashboard) GetStatistics() *ConnectionValidationStatistics {
	cvd.mu.RLock()
	defer cvd.mu.RUnlock()
	
	return cvd.statistics
}

// GetRecentResults returns the most recent validation results
func (cvd *ConnectionValidationDashboard) GetRecentResults(count int) []*ConnectionValidationResult {
	cvd.mu.RLock()
	defer cvd.mu.RUnlock()
	
	return cvd.getRecentResults(count)
}

// updateStatistics updates the statistics based on current results
func (cvd *ConnectionValidationDashboard) updateStatistics() {
	stats := cvd.statistics
	stats.TotalValidations = len(cvd.results)
	
	// Reset counters
	stats.SuccessfulConnections = 0
	stats.FailedConnections = 0
	stats.NoClearConnections = 0
	stats.AverageConfidence = 0.0
	stats.AverageProcessingTime = 0.0
	
	// Clear distributions
	stats.ConfidenceDistribution = make(map[string]int)
	stats.RiskLevelDistribution = make(map[string]int)
	stats.ConnectionStrengthDistribution = make(map[string]int)
	
	totalConfidence := 0.0
	totalProcessingTime := 0.0
	
	for _, result := range cvd.results {
		// Count connections
		if result.IsConnected {
			stats.SuccessfulConnections++
		} else {
			stats.FailedConnections++
		}
		
		// Count no clear connections
		if result.ConnectionStrength == "none" {
			stats.NoClearConnections++
		}
		
		// Accumulate confidence and processing time
		totalConfidence += result.OverallConfidence
		totalProcessingTime += float64(result.ValidationTime.Milliseconds()) / 1000.0
		
		// Update distributions
		cvd.updateDistributions(result, stats)
	}
	
	// Calculate averages
	if stats.TotalValidations > 0 {
		stats.AverageConfidence = totalConfidence / float64(stats.TotalValidations)
		stats.AverageProcessingTime = totalProcessingTime / float64(stats.TotalValidations)
	}
	
	stats.LastUpdated = time.Now()
}

// updateDistributions updates the distribution maps
func (cvd *ConnectionValidationDashboard) updateDistributions(result *ConnectionValidationResult, stats *ConnectionValidationStatistics) {
	// Confidence distribution
	confidenceLevel := cvd.getConfidenceLevel(result.OverallConfidence)
	stats.ConfidenceDistribution[confidenceLevel]++
	
	// Connection strength distribution
	stats.ConnectionStrengthDistribution[result.ConnectionStrength]++
	
	// Risk level distribution (if available from no clear connection detection)
	// This would be populated if we integrate the NoClearConnectionDetector
}

// getConfidenceLevel categorizes confidence scores
func (cvd *ConnectionValidationDashboard) getConfidenceLevel(confidence float64) string {
	if confidence >= 0.9 {
		return "excellent"
	} else if confidence >= 0.7 {
		return "good"
	} else if confidence >= 0.5 {
		return "fair"
	} else {
		return "poor"
	}
}

// getRecentResults returns the most recent results
func (cvd *ConnectionValidationDashboard) getRecentResults(count int) []*ConnectionValidationResult {
	if count > len(cvd.results) {
		count = len(cvd.results)
	}
	
	start := len(cvd.results) - count
	return cvd.results[start:]
}

// checkAlerts checks for conditions that should trigger alerts
func (cvd *ConnectionValidationDashboard) checkAlerts() {
	stats := cvd.statistics
	thresholds := cvd.config.AlertThresholds
	
	// Check low confidence rate
	if stats.TotalValidations > 0 {
		lowConfidenceCount := stats.ConfidenceDistribution["poor"] + stats.ConfidenceDistribution["fair"]
		lowConfidenceRate := float64(lowConfidenceCount) / float64(stats.TotalValidations)
		
		if lowConfidenceRate > thresholds.LowConfidenceRate {
			cvd.createAlert("confidence", "high", 
				fmt.Sprintf("Low confidence rate: %.2f%% (threshold: %.2f%%)", 
					lowConfidenceRate*100, thresholds.LowConfidenceRate*100),
				thresholds.LowConfidenceRate, lowConfidenceRate)
		}
	}
	
	// Check no connection rate
	if stats.TotalValidations > 0 {
		noConnectionRate := float64(stats.NoClearConnections) / float64(stats.TotalValidations)
		if noConnectionRate > thresholds.NoConnectionRate {
			cvd.createAlert("connection", "high",
				fmt.Sprintf("No clear connection rate: %.2f%% (threshold: %.2f%%)",
					noConnectionRate*100, thresholds.NoConnectionRate*100),
				thresholds.NoConnectionRate, noConnectionRate)
		}
	}
	
	// Check average confidence
	if stats.AverageConfidence < thresholds.AverageConfidence {
		cvd.createAlert("confidence", "medium",
			fmt.Sprintf("Average confidence: %.2f (threshold: %.2f)",
				stats.AverageConfidence, thresholds.AverageConfidence),
			thresholds.AverageConfidence, stats.AverageConfidence)
	}
	
	// Check processing time
	if stats.AverageProcessingTime > thresholds.ProcessingTime {
		cvd.createAlert("performance", "medium",
			fmt.Sprintf("Average processing time: %.2fs (threshold: %.2fs)",
				stats.AverageProcessingTime, thresholds.ProcessingTime),
			thresholds.ProcessingTime, stats.AverageProcessingTime)
	}
}

// createAlert creates a new dashboard alert
func (cvd *ConnectionValidationDashboard) createAlert(alertType, severity, message string, threshold, currentValue float64) {
	// In a real implementation, this would be stored and retrieved
	// For now, we'll just log it
	fmt.Printf("DASHBOARD ALERT: %s - %s: %s\n", severity, alertType, message)
}

// getActiveAlerts returns currently active alerts
func (cvd *ConnectionValidationDashboard) getActiveAlerts() []DashboardAlert {
	// In a real implementation, this would return stored alerts
	// For now, return empty slice
	return []DashboardAlert{}
}

// calculateTrends calculates trend data
func (cvd *ConnectionValidationDashboard) calculateTrends() []TrendData {
	// In a real implementation, this would calculate trends over time
	// For now, return empty slice
	return []TrendData{}
}

// generateRecommendations generates recommendations based on current statistics
func (cvd *ConnectionValidationDashboard) generateRecommendations() []string {
	recommendations := []string{}
	stats := cvd.statistics
	
	if stats.TotalValidations == 0 {
		return []string{"No validation data available yet"}
	}
	
	// Check confidence levels
	if stats.AverageConfidence < 0.7 {
		recommendations = append(recommendations,
			"Average confidence is low - consider improving validation algorithms")
	}
	
	// Check connection success rate
	successRate := float64(stats.SuccessfulConnections) / float64(stats.TotalValidations)
	if successRate < 0.8 {
		recommendations = append(recommendations,
			"Connection success rate is low - review validation criteria")
	}
	
	// Check processing time
	if stats.AverageProcessingTime > 5.0 {
		recommendations = append(recommendations,
			"Average processing time is high - consider performance optimization")
	}
	
	// Check no clear connection rate
	noConnectionRate := float64(stats.NoClearConnections) / float64(stats.TotalValidations)
	if noConnectionRate > 0.3 {
		recommendations = append(recommendations,
			"High rate of no clear connections - review detection criteria")
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations,
			"All metrics are within acceptable ranges")
	}
	
	return recommendations
}

// updateStatisticsPeriodically updates statistics at regular intervals
func (cvd *ConnectionValidationDashboard) updateStatisticsPeriodically() {
	ticker := time.NewTicker(cvd.config.StatisticsInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		cvd.mu.Lock()
		cvd.updateStatistics()
		cvd.mu.Unlock()
	}
}
