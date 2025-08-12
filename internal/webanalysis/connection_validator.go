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
	// Simplified Levenshtein distance calculation
	// In a real implementation, this would calculate the actual edit distance
	return 0.0
}

func calculateJaroWinklerSimilarity(s1, s2 string) float64 {
	// Simplified Jaro-Winkler similarity calculation
	// In a real implementation, this would calculate the actual similarity
	return 0.0
}

func calculateCosineSimilarity(s1, s2 string) float64 {
	// Simplified cosine similarity calculation
	// In a real implementation, this would calculate the actual similarity
	return 0.0
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
