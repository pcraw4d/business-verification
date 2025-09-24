package external

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// KeyPersonnelExtractor extracts information about key personnel and executive team members
type KeyPersonnelExtractor struct {
	config *KeyPersonnelConfig
	logger *zap.Logger
}

// KeyPersonnelConfig contains configuration for personnel extraction
type KeyPersonnelConfig struct {
	// Extraction settings
	EnableExecutiveExtraction bool `json:"enable_executive_extraction"`
	EnableTeamExtraction      bool `json:"enable_team_extraction"`
	EnableRoleDetection       bool `json:"enable_role_detection"`
	EnableLinkedInIntegration bool `json:"enable_linkedin_integration"`

	// Executive titles to look for
	ExecutiveTitles  []string `json:"executive_titles"`
	SeniorTitles     []string `json:"senior_titles"`
	DepartmentTitles []string `json:"department_titles"`

	// Extraction patterns
	NamePatterns  []string `json:"name_patterns"`
	RolePatterns  []string `json:"role_patterns"`
	EmailPatterns []string `json:"email_patterns"`

	// Quality settings
	MinConfidenceThreshold   float64 `json:"min_confidence_threshold"`
	EnableDuplicateDetection bool    `json:"enable_duplicate_detection"`
	EnableContextValidation  bool    `json:"enable_context_validation"`

	// Privacy and compliance
	EnableDataAnonymization bool     `json:"enable_data_anonymization"`
	ExcludedDomains         []string `json:"excluded_domains"`
	MaxPersonnelCount       int      `json:"max_personnel_count"`
}

// ExecutiveTeamMember represents a key personnel or executive team member
type ExecutiveTeamMember struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Title           string    `json:"title"`
	Department      string    `json:"department"`
	Level           string    `json:"level"` // executive, senior, manager, etc.
	Email           string    `json:"email"`
	LinkedInURL     string    `json:"linkedin_url"`
	Bio             string    `json:"bio"`
	ImageURL        string    `json:"image_url"`
	ConfidenceScore float64   `json:"confidence_score"`
	ExtractedAt     time.Time `json:"extracted_at"`

	// Validation and quality
	ValidationStatus ValidationStatus   `json:"validation_status"`
	DataQuality      DataQualityMetrics `json:"data_quality"`

	// Privacy compliance
	PrivacyCompliance PrivacyComplianceInfo `json:"privacy_compliance"`
}

// PersonnelExtractionResult contains the results of personnel extraction
type PersonnelExtractionResult struct {
	Executives       []ExecutiveTeamMember `json:"executives"`
	SeniorManagement []ExecutiveTeamMember `json:"senior_management"`
	TeamMembers      []ExecutiveTeamMember `json:"team_members"`
	TotalExtracted   int                   `json:"total_extracted"`

	// Statistics
	ExtractionStats PersonnelExtractionStats `json:"extraction_stats"`

	// Metadata
	ExtractionTime  time.Duration `json:"extraction_time"`
	SourceURL       string        `json:"source_url"`
	ConfidenceScore float64       `json:"confidence_score"`
}

// PersonnelExtractionStats contains statistics about the extraction process
type PersonnelExtractionStats struct {
	TotalMatches      int     `json:"total_matches"`
	ValidExecutives   int     `json:"valid_executives"`
	ValidSeniorMgmt   int     `json:"valid_senior_mgmt"`
	ValidTeamMembers  int     `json:"valid_team_members"`
	AverageConfidence float64 `json:"average_confidence"`
	DuplicateCount    int     `json:"duplicate_count"`
	AnonymizedCount   int     `json:"anonymized_count"`
}

// NewKeyPersonnelExtractor creates a new key personnel extractor
func NewKeyPersonnelExtractor(config *KeyPersonnelConfig, logger *zap.Logger) *KeyPersonnelExtractor {
	if config == nil {
		config = getDefaultKeyPersonnelConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &KeyPersonnelExtractor{
		config: config,
		logger: logger,
	}
}

// ExtractKeyPersonnel extracts key personnel and executive team information from website content
func (kpe *KeyPersonnelExtractor) ExtractKeyPersonnel(ctx context.Context, content string, sourceURL string) (*PersonnelExtractionResult, error) {
	startTime := time.Now()

	result := &PersonnelExtractionResult{
		Executives:       make([]ExecutiveTeamMember, 0),
		SeniorManagement: make([]ExecutiveTeamMember, 0),
		TeamMembers:      make([]ExecutiveTeamMember, 0),
		SourceURL:        sourceURL,
	}

	// Check context timeout
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("extraction cancelled: %w", ctx.Err())
	default:
	}

	// Extract executives if enabled
	if kpe.config.EnableExecutiveExtraction {
		executives, err := kpe.extractExecutives(ctx, content)
		if err != nil {
			kpe.logger.Error("executive extraction failed", zap.Error(err))
		} else {
			result.Executives = executives
		}
	}

	// Extract senior management if enabled
	if kpe.config.EnableTeamExtraction {
		seniorMgmt, err := kpe.extractSeniorManagement(ctx, content)
		if err != nil {
			kpe.logger.Error("senior management extraction failed", zap.Error(err))
		} else {
			result.SeniorManagement = seniorMgmt
		}
	}

	// Extract team members if enabled
	if kpe.config.EnableTeamExtraction {
		teamMembers, err := kpe.extractTeamMembers(ctx, content)
		if err != nil {
			kpe.logger.Error("team member extraction failed", zap.Error(err))
		} else {
			result.TeamMembers = teamMembers
		}
	}

	// Apply duplicate detection if enabled
	if kpe.config.EnableDuplicateDetection {
		result.Executives = kpe.deduplicatePersonnel(result.Executives)
		result.SeniorManagement = kpe.deduplicatePersonnel(result.SeniorManagement)
		result.TeamMembers = kpe.deduplicatePersonnel(result.TeamMembers)
	}

	// Apply confidence threshold filtering
	result.Executives = kpe.filterByConfidence(result.Executives)
	result.SeniorManagement = kpe.filterByConfidence(result.SeniorManagement)
	result.TeamMembers = kpe.filterByConfidence(result.TeamMembers)

	// Apply data anonymization if enabled
	if kpe.config.EnableDataAnonymization {
		result.Executives = kpe.anonymizePersonnel(result.Executives)
		result.SeniorManagement = kpe.anonymizePersonnel(result.SeniorManagement)
		result.TeamMembers = kpe.anonymizePersonnel(result.TeamMembers)
	}

	// Calculate statistics
	result.TotalExtracted = len(result.Executives) + len(result.SeniorManagement) + len(result.TeamMembers)
	result.ExtractionStats = kpe.calculatePersonnelStats(result)
	result.ExtractionTime = time.Since(startTime)
	result.ConfidenceScore = kpe.calculateOverallConfidence(result)

	kpe.logger.Info("personnel extraction completed",
		zap.Int("total_extracted", result.TotalExtracted),
		zap.Int("executives", len(result.Executives)),
		zap.Int("senior_management", len(result.SeniorManagement)),
		zap.Int("team_members", len(result.TeamMembers)),
		zap.Duration("extraction_time", result.ExtractionTime),
		zap.Float64("confidence_score", result.ConfidenceScore))

	return result, nil
}

// extractExecutives extracts executive-level personnel from content
func (kpe *KeyPersonnelExtractor) extractExecutives(ctx context.Context, content string) ([]ExecutiveTeamMember, error) {
	var executives []ExecutiveTeamMember

	// Executive title patterns
	executivePatterns := []string{
		`(?i)(CEO|Chief Executive Officer|President|Founder|Co-Founder|Managing Director)`,
		`(?i)(CTO|Chief Technology Officer|Chief Technical Officer)`,
		`(?i)(CFO|Chief Financial Officer)`,
		`(?i)(COO|Chief Operating Officer)`,
		`(?i)(CMO|Chief Marketing Officer)`,
		`(?i)(CHRO|Chief Human Resources Officer|VP of HR)`,
		`(?i)(CLO|Chief Legal Officer|General Counsel)`,
		`(?i)(CDO|Chief Data Officer|Chief Digital Officer)`,
	}

	for _, pattern := range executivePatterns {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) > 1 {
				title := strings.TrimSpace(match[1])

				// Look for name near the title
				name := kpe.extractNameNearTitle(content, title)
				if name != "" {
					executive := ExecutiveTeamMember{
						ID:              generateID(),
						Name:            name,
						Title:           title,
						Level:           "executive",
						ConfidenceScore: kpe.calculateExecutiveConfidence(name, title),
						ExtractedAt:     time.Now(),
						ValidationStatus: ValidationStatus{
							IsValid:          true,
							ValidationErrors: []string{},
						},
						DataQuality: DataQualityMetrics{
							Completeness: 0.8,
							Accuracy:     0.9,
							Consistency:  0.85,
							Timeliness:   1.0,
						},
						PrivacyCompliance: PrivacyComplianceInfo{
							IsGDPRCompliant: true,
							IsAnonymized:    kpe.config.EnableDataAnonymization,
						},
					}

					// Extract additional information
					executive.Email = kpe.extractEmailForPerson(name, content)
					executive.LinkedInURL = kpe.extractLinkedInURL(name, content)
					executive.Bio = kpe.extractBioForPerson(name, content)
					executive.Department = kpe.determineDepartment(title)

					executives = append(executives, executive)
				}
			}
		}
	}

	return executives, nil
}

// extractSeniorManagement extracts senior management personnel
func (kpe *KeyPersonnelExtractor) extractSeniorManagement(ctx context.Context, content string) ([]ExecutiveTeamMember, error) {
	var seniorMgmt []ExecutiveTeamMember

	// Senior management patterns
	seniorPatterns := []string{
		`(?i)(VP|Vice President|Senior Vice President|Executive Vice President)`,
		`(?i)(Director|Senior Director|Executive Director)`,
		`(?i)(Head of|Head|Lead|Principal)`,
		`(?i)(Manager|Senior Manager|General Manager)`,
	}

	for _, pattern := range seniorPatterns {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) > 1 {
				title := strings.TrimSpace(match[1])

				// Look for name near the title
				name := kpe.extractNameNearTitle(content, title)
				if name != "" {
					senior := ExecutiveTeamMember{
						ID:              generateID(),
						Name:            name,
						Title:           title,
						Level:           "senior",
						ConfidenceScore: kpe.calculateSeniorConfidence(name, title),
						ExtractedAt:     time.Now(),
						ValidationStatus: ValidationStatus{
							IsValid:          true,
							ValidationErrors: []string{},
						},
						DataQuality: DataQualityMetrics{
							Completeness: 0.7,
							Accuracy:     0.85,
							Consistency:  0.8,
							Timeliness:   1.0,
						},
						PrivacyCompliance: PrivacyComplianceInfo{
							IsGDPRCompliant: true,
							IsAnonymized:    kpe.config.EnableDataAnonymization,
						},
					}

					// Extract additional information
					senior.Email = kpe.extractEmailForPerson(name, content)
					senior.LinkedInURL = kpe.extractLinkedInURL(name, content)
					senior.Bio = kpe.extractBioForPerson(name, content)
					senior.Department = kpe.determineDepartment(title)

					seniorMgmt = append(seniorMgmt, senior)
				}
			}
		}
	}

	return seniorMgmt, nil
}

// extractTeamMembers extracts general team members
func (kpe *KeyPersonnelExtractor) extractTeamMembers(ctx context.Context, content string) ([]ExecutiveTeamMember, error) {
	var teamMembers []ExecutiveTeamMember

	// Team member patterns
	teamPatterns := []string{
		`(?i)(Developer|Engineer|Software Engineer|Full Stack Developer)`,
		`(?i)(Designer|UX Designer|UI Designer|Product Designer)`,
		`(?i)(Analyst|Business Analyst|Data Analyst|Product Analyst)`,
		`(?i)(Specialist|Coordinator|Associate|Assistant)`,
	}

	for _, pattern := range teamPatterns {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) > 1 {
				title := strings.TrimSpace(match[1])

				// Look for name near the title
				name := kpe.extractNameNearTitle(content, title)
				if name != "" {
					member := ExecutiveTeamMember{
						ID:              generateID(),
						Name:            name,
						Title:           title,
						Level:           "team",
						ConfidenceScore: kpe.calculateTeamConfidence(name, title),
						ExtractedAt:     time.Now(),
						ValidationStatus: ValidationStatus{
							IsValid:          true,
							ValidationErrors: []string{},
						},
						DataQuality: DataQualityMetrics{
							Completeness: 0.6,
							Accuracy:     0.8,
							Consistency:  0.75,
							Timeliness:   1.0,
						},
						PrivacyCompliance: PrivacyComplianceInfo{
							IsGDPRCompliant: true,
							IsAnonymized:    kpe.config.EnableDataAnonymization,
						},
					}

					// Extract additional information
					member.Email = kpe.extractEmailForPerson(name, content)
					member.LinkedInURL = kpe.extractLinkedInURL(name, content)
					member.Bio = kpe.extractBioForPerson(name, content)
					member.Department = kpe.determineDepartment(title)

					teamMembers = append(teamMembers, member)
				}
			}
		}
	}

	return teamMembers, nil
}

// extractNameNearTitle extracts a person's name near a given title
func (kpe *KeyPersonnelExtractor) extractNameNearTitle(content, title string) string {
	// Look for name patterns near the title
	namePatterns := []string{
		fmt.Sprintf(`(?i)([A-Z][a-z]+ [A-Z][a-z]+).*%s`, regexp.QuoteMeta(title)),
		fmt.Sprintf(`(?i)%s.*([A-Z][a-z]+ [A-Z][a-z]+)`, regexp.QuoteMeta(title)),
		fmt.Sprintf(`(?i)([A-Z][a-z]+ [A-Z][a-z]+ [A-Z][a-z]+).*%s`, regexp.QuoteMeta(title)),
		fmt.Sprintf(`(?i)%s.*([A-Z][a-z]+ [A-Z][a-z]+ [A-Z][a-z]+)`, regexp.QuoteMeta(title)),
	}

	for _, pattern := range namePatterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindStringSubmatch(content)
		if len(matches) > 1 {
			name := strings.TrimSpace(matches[1])
			// Basic name validation
			if len(name) > 5 && len(name) < 50 {
				return name
			}
		}
	}

	return ""
}

// extractEmailForPerson extracts email address for a specific person
func (kpe *KeyPersonnelExtractor) extractEmailForPerson(name, content string) string {
	// Look for email patterns near the person's name
	emailPattern := fmt.Sprintf(`(?i)%s.*?([a-zA-Z0-9._%%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`, regexp.QuoteMeta(name))
	regex := regexp.MustCompile(emailPattern)
	matches := regex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

// extractLinkedInURL extracts LinkedIn URL for a person
func (kpe *KeyPersonnelExtractor) extractLinkedInURL(name, content string) string {
	// Look for LinkedIn patterns near the person's name
	linkedInPattern := fmt.Sprintf(`(?i)%s.*?(https?://[^\\s]*linkedin\\.com[^\\s]*)`, regexp.QuoteMeta(name))
	regex := regexp.MustCompile(linkedInPattern)
	matches := regex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

// extractBioForPerson extracts bio information for a person
func (kpe *KeyPersonnelExtractor) extractBioForPerson(name, content string) string {
	// Look for bio patterns near the person's name
	bioPattern := fmt.Sprintf(`(?i)%s.*?([^.!?]*[.!?])`, regexp.QuoteMeta(name))
	regex := regexp.MustCompile(bioPattern)
	matches := regex.FindStringSubmatch(content)
	if len(matches) > 1 {
		bio := strings.TrimSpace(matches[1])
		// Limit bio length
		if len(bio) > 200 {
			bio = bio[:200] + "..."
		}
		return bio
	}

	return ""
}

// determineDepartment determines the department based on title
func (kpe *KeyPersonnelExtractor) determineDepartment(title string) string {
	titleLower := strings.ToLower(title)

	departmentMap := map[string][]string{
		"Engineering": {"engineer", "developer", "cto", "technical", "software", "devops"},
		"Finance":     {"cfo", "finance", "accounting", "financial", "treasurer"},
		"Marketing":   {"cmo", "marketing", "brand", "communications", "pr"},
		"Sales":       {"sales", "business development", "bd", "revenue"},
		"HR":          {"hr", "human resources", "people", "talent", "recruiting"},
		"Legal":       {"legal", "counsel", "compliance", "regulatory"},
		"Operations":  {"coo", "operations", "operational", "process"},
		"Product":     {"product", "pm", "product manager", "design"},
		"Data":        {"data", "analytics", "ai", "ml", "machine learning"},
	}

	for dept, keywords := range departmentMap {
		for _, keyword := range keywords {
			if strings.Contains(titleLower, keyword) {
				return dept
			}
		}
	}

	return "General"
}

// calculateExecutiveConfidence calculates confidence score for executives
func (kpe *KeyPersonnelExtractor) calculateExecutiveConfidence(name, title string) float64 {
	confidence := 0.9 // Base confidence for executives

	// Adjust based on name quality
	if len(name) > 10 {
		confidence += 0.05
	}

	// Adjust based on title specificity
	if strings.Contains(strings.ToLower(title), "chief") {
		confidence += 0.05
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// calculateSeniorConfidence calculates confidence score for senior management
func (kpe *KeyPersonnelExtractor) calculateSeniorConfidence(name, title string) float64 {
	confidence := 0.8 // Base confidence for senior management

	// Adjust based on name quality
	if len(name) > 10 {
		confidence += 0.05
	}

	// Adjust based on title specificity
	if strings.Contains(strings.ToLower(title), "vp") || strings.Contains(strings.ToLower(title), "director") {
		confidence += 0.05
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// calculateTeamConfidence calculates confidence score for team members
func (kpe *KeyPersonnelExtractor) calculateTeamConfidence(name, title string) float64 {
	confidence := 0.7 // Base confidence for team members

	// Adjust based on name quality
	if len(name) > 10 {
		confidence += 0.05
	}

	// Adjust based on title specificity
	if strings.Contains(strings.ToLower(title), "senior") {
		confidence += 0.05
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// deduplicatePersonnel removes duplicate personnel entries
func (kpe *KeyPersonnelExtractor) deduplicatePersonnel(personnel []ExecutiveTeamMember) []ExecutiveTeamMember {
	seen := make(map[string]bool)
	var unique []ExecutiveTeamMember

	for _, person := range personnel {
		key := strings.ToLower(person.Name + "|" + person.Title)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, person)
		}
	}

	return unique
}

// filterByConfidence filters personnel by confidence threshold
func (kpe *KeyPersonnelExtractor) filterByConfidence(personnel []ExecutiveTeamMember) []ExecutiveTeamMember {
	var filtered []ExecutiveTeamMember

	for _, person := range personnel {
		if person.ConfidenceScore >= kpe.config.MinConfidenceThreshold {
			filtered = append(filtered, person)
		}
	}

	return filtered
}

// anonymizePersonnel anonymizes personnel data for privacy compliance
func (kpe *KeyPersonnelExtractor) anonymizePersonnel(personnel []ExecutiveTeamMember) []ExecutiveTeamMember {
	var anonymized []ExecutiveTeamMember

	for _, person := range personnel {
		// Anonymize name (keep first letter and last name)
		if person.Name != "" {
			parts := strings.Fields(person.Name)
			if len(parts) >= 2 {
				person.Name = string(parts[0][0]) + ". " + parts[len(parts)-1]
			}
		}

		// Remove email and LinkedIn for privacy
		person.Email = ""
		person.LinkedInURL = ""

		// Anonymize bio
		if person.Bio != "" {
			person.Bio = "Professional bio available"
		}

		person.PrivacyCompliance.IsAnonymized = true
		anonymized = append(anonymized, person)
	}

	return anonymized
}

// calculatePersonnelStats calculates statistics for personnel extraction
func (kpe *KeyPersonnelExtractor) calculatePersonnelStats(result *PersonnelExtractionResult) PersonnelExtractionStats {
	stats := PersonnelExtractionStats{}

	allPersonnel := append(append(result.Executives, result.SeniorManagement...), result.TeamMembers...)

	stats.TotalMatches = len(allPersonnel)
	stats.ValidExecutives = len(result.Executives)
	stats.ValidSeniorMgmt = len(result.SeniorManagement)
	stats.ValidTeamMembers = len(result.TeamMembers)

	// Calculate average confidence
	totalConfidence := 0.0
	for _, person := range allPersonnel {
		totalConfidence += person.ConfidenceScore
	}

	if len(allPersonnel) > 0 {
		stats.AverageConfidence = totalConfidence / float64(len(allPersonnel))
	}

	return stats
}

// calculateOverallConfidence calculates overall confidence score for the extraction
func (kpe *KeyPersonnelExtractor) calculateOverallConfidence(result *PersonnelExtractionResult) float64 {
	if result.TotalExtracted == 0 {
		return 0.0
	}

	// Weight executives more heavily
	executiveWeight := 0.5
	seniorWeight := 0.3
	teamWeight := 0.2

	totalScore := 0.0
	totalWeight := 0.0

	for _, exec := range result.Executives {
		totalScore += exec.ConfidenceScore * executiveWeight
		totalWeight += executiveWeight
	}

	for _, senior := range result.SeniorManagement {
		totalScore += senior.ConfidenceScore * seniorWeight
		totalWeight += seniorWeight
	}

	for _, team := range result.TeamMembers {
		totalScore += team.ConfidenceScore * teamWeight
		totalWeight += teamWeight
	}

	if totalWeight > 0 {
		return totalScore / totalWeight
	}

	return 0.0
}

// UpdateConfig updates the extractor configuration
func (kpe *KeyPersonnelExtractor) UpdateConfig(config *KeyPersonnelConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	kpe.config = config
	kpe.logger.Info("key personnel extractor config updated")
	return nil
}

// GetConfig returns the current configuration
func (kpe *KeyPersonnelExtractor) GetConfig() *KeyPersonnelConfig {
	return kpe.config
}

// getDefaultKeyPersonnelConfig returns default configuration
func getDefaultKeyPersonnelConfig() *KeyPersonnelConfig {
	return &KeyPersonnelConfig{
		EnableExecutiveExtraction: true,
		EnableTeamExtraction:      true,
		EnableRoleDetection:       true,
		EnableLinkedInIntegration: false,
		ExecutiveTitles: []string{
			"CEO", "CTO", "CFO", "COO", "CMO", "CHRO", "CLO", "CDO",
			"Chief Executive Officer", "Chief Technology Officer", "Chief Financial Officer",
			"Chief Operating Officer", "Chief Marketing Officer", "Chief Human Resources Officer",
			"Chief Legal Officer", "Chief Data Officer", "President", "Founder", "Co-Founder",
		},
		SeniorTitles: []string{
			"VP", "Vice President", "Senior Vice President", "Executive Vice President",
			"Director", "Senior Director", "Executive Director", "Head of", "Lead", "Principal",
		},
		DepartmentTitles: []string{
			"Engineering", "Finance", "Marketing", "Sales", "HR", "Legal", "Operations", "Product", "Data",
		},
		NamePatterns: []string{
			`[A-Z][a-z]+ [A-Z][a-z]+`,
			`[A-Z][a-z]+ [A-Z][a-z]+ [A-Z][a-z]+`,
		},
		RolePatterns: []string{
			`(?i)(CEO|CTO|CFO|COO|CMO|VP|Director|Manager|Lead|Head)`,
		},
		EmailPatterns: []string{
			`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
		},
		MinConfidenceThreshold:   0.6,
		EnableDuplicateDetection: true,
		EnableContextValidation:  true,
		EnableDataAnonymization:  false,
		ExcludedDomains:          []string{},
		MaxPersonnelCount:        50,
	}
}
