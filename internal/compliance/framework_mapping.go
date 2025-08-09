package compliance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// FrameworkMappingSystem provides regulatory framework mapping functionality
type FrameworkMappingSystem struct {
	logger       *observability.Logger
	mu           sync.RWMutex
	frameworks   map[string]*RegulatoryFramework // frameworkID -> framework
	mappings     map[string]*FrameworkMapping    // mappingID -> mapping
	mappingRules map[string][]MappingRule        // frameworkID -> rules
	crosswalks   map[string]*CrosswalkMapping    // crosswalkID -> crosswalk
	confidence   map[string]float64              // mappingID -> confidence score
}

// MappingRule represents a rule for framework mapping
type MappingRule struct {
	ID              string                 `json:"id"`
	SourceFramework string                 `json:"source_framework"`
	TargetFramework string                 `json:"target_framework"`
	RuleType        string                 `json:"rule_type"` // "exact", "partial", "related", "superseded"
	Confidence      float64                `json:"confidence"`
	Conditions      map[string]interface{} `json:"conditions"`
	Transformations map[string]interface{} `json:"transformations"`
	Notes           string                 `json:"notes"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// CrosswalkMapping represents a crosswalk between frameworks
type CrosswalkMapping struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	SourceFramework string                 `json:"source_framework"`
	TargetFramework string                 `json:"target_framework"`
	MappingType     MappingType            `json:"mapping_type"`
	Confidence      float64                `json:"confidence"`
	Mappings        []FrameworkMapping     `json:"mappings"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// MappingAnalysis represents analysis of framework mappings
type MappingAnalysis struct {
	ID                 string                  `json:"id"`
	SourceFramework    string                  `json:"source_framework"`
	TargetFramework    string                  `json:"target_framework"`
	TotalMappings      int                     `json:"total_mappings"`
	ExactMappings      int                     `json:"exact_mappings"`
	PartialMappings    int                     `json:"partial_mappings"`
	RelatedMappings    int                     `json:"related_mappings"`
	SupersededMappings int                     `json:"superseded_mappings"`
	AverageConfidence  float64                 `json:"average_confidence"`
	CoveragePercentage float64                 `json:"coverage_percentage"`
	Gaps               []MappingGap            `json:"gaps"`
	Recommendations    []MappingRecommendation `json:"recommendations"`
	AnalysisDate       time.Time               `json:"analysis_date"`
}

// MappingGap represents a gap in framework mapping
type MappingGap struct {
	ID                     string `json:"id"`
	SourceRequirementID    string `json:"source_requirement_id"`
	SourceRequirementTitle string `json:"source_requirement_title"`
	TargetRequirementID    string `json:"target_requirement_id"`
	TargetRequirementTitle string `json:"target_requirement_title"`
	GapType                string `json:"gap_type"` // "missing", "unmapped", "incomplete"
	Severity               string `json:"severity"` // "low", "medium", "high", "critical"
	Description            string `json:"description"`
	Impact                 string `json:"impact"`
	Recommendation         string `json:"recommendation"`
}

// MappingRecommendation represents a recommendation for framework mapping
type MappingRecommendation struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`     // "create", "update", "delete", "review"
	Priority    string    `json:"priority"` // "low", "medium", "high", "critical"
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Action      string    `json:"action"`
	Impact      string    `json:"impact"`
	Effort      string    `json:"effort"`
	Timeline    string    `json:"timeline"`
	AssignedTo  string    `json:"assigned_to"`
	Status      string    `json:"status"` // "open", "in_progress", "completed", "rejected"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewFrameworkMappingSystem creates a new framework mapping system
func NewFrameworkMappingSystem(logger *observability.Logger) *FrameworkMappingSystem {
	return &FrameworkMappingSystem{
		logger:       logger,
		frameworks:   make(map[string]*RegulatoryFramework),
		mappings:     make(map[string]*FrameworkMapping),
		mappingRules: make(map[string][]MappingRule),
		crosswalks:   make(map[string]*CrosswalkMapping),
		confidence:   make(map[string]float64),
	}
}

// RegisterFramework registers a regulatory framework
func (s *FrameworkMappingSystem) RegisterFramework(ctx context.Context, framework *RegulatoryFramework) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Registering regulatory framework",
		"request_id", requestID,
		"framework_id", framework.ID,
		"framework_name", framework.Name,
		"framework_type", framework.Type,
		"jurisdiction", framework.Jurisdiction,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.frameworks[framework.ID] = framework

	s.logger.Info("Regulatory framework registered successfully",
		"request_id", requestID,
		"framework_id", framework.ID,
		"framework_name", framework.Name,
		"requirement_count", len(framework.Requirements),
	)

	return nil
}

// CreateMapping creates a mapping between frameworks
func (s *FrameworkMappingSystem) CreateMapping(ctx context.Context, mapping *FrameworkMapping) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Creating framework mapping",
		"request_id", requestID,
		"source_framework", mapping.SourceFramework,
		"target_framework", mapping.TargetFramework,
		"source_requirement_id", mapping.SourceRequirementID,
		"target_requirement_id", mapping.TargetRequirementID,
		"mapping_type", mapping.MappingType,
		"confidence", mapping.Confidence,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate frameworks exist
	if _, exists := s.frameworks[mapping.SourceFramework]; !exists {
		return fmt.Errorf("source framework %s not found", mapping.SourceFramework)
	}
	if _, exists := s.frameworks[mapping.TargetFramework]; !exists {
		return fmt.Errorf("target framework %s not found", mapping.TargetFramework)
	}

	// Generate mapping ID if not provided
	if mapping.ID == "" {
		mapping.ID = fmt.Sprintf("mapping_%s_%s_%s_%s",
			mapping.SourceFramework, mapping.SourceRequirementID,
			mapping.TargetFramework, mapping.TargetRequirementID)
	}

	// Set timestamps
	now := time.Now()
	if mapping.CreatedAt.IsZero() {
		mapping.CreatedAt = now
	}
	mapping.UpdatedAt = now

	s.mappings[mapping.ID] = mapping
	s.confidence[mapping.ID] = mapping.Confidence

	s.logger.Info("Framework mapping created successfully",
		"request_id", requestID,
		"mapping_id", mapping.ID,
		"source_framework", mapping.SourceFramework,
		"target_framework", mapping.TargetFramework,
		"confidence", mapping.Confidence,
	)

	return nil
}

// GetMapping gets a specific framework mapping
func (s *FrameworkMappingSystem) GetMapping(ctx context.Context, mappingID string) (*FrameworkMapping, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting framework mapping",
		"request_id", requestID,
		"mapping_id", mappingID,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	mapping, exists := s.mappings[mappingID]
	if !exists {
		return nil, fmt.Errorf("mapping %s not found", mappingID)
	}

	return mapping, nil
}

// GetMappingsByFramework gets all mappings for a specific framework
func (s *FrameworkMappingSystem) GetMappingsByFramework(ctx context.Context, frameworkID string, direction string) ([]*FrameworkMapping, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting framework mappings",
		"request_id", requestID,
		"framework_id", frameworkID,
		"direction", direction,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	var mappings []*FrameworkMapping
	for _, mapping := range s.mappings {
		switch direction {
		case "source":
			if mapping.SourceFramework == frameworkID {
				mappings = append(mappings, mapping)
			}
		case "target":
			if mapping.TargetFramework == frameworkID {
				mappings = append(mappings, mapping)
			}
		case "both":
			if mapping.SourceFramework == frameworkID || mapping.TargetFramework == frameworkID {
				mappings = append(mappings, mapping)
			}
		}
	}

	return mappings, nil
}

// GetMappingsBetweenFrameworks gets mappings between two specific frameworks
func (s *FrameworkMappingSystem) GetMappingsBetweenFrameworks(ctx context.Context, sourceFramework, targetFramework string) ([]*FrameworkMapping, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting mappings between frameworks",
		"request_id", requestID,
		"source_framework", sourceFramework,
		"target_framework", targetFramework,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	var mappings []*FrameworkMapping
	for _, mapping := range s.mappings {
		if mapping.SourceFramework == sourceFramework && mapping.TargetFramework == targetFramework {
			mappings = append(mappings, mapping)
		}
	}

	return mappings, nil
}

// UpdateMapping updates an existing framework mapping
func (s *FrameworkMappingSystem) UpdateMapping(ctx context.Context, mappingID string, updates map[string]interface{}) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating framework mapping",
		"request_id", requestID,
		"mapping_id", mappingID,
		"updates", updates,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	mapping, exists := s.mappings[mappingID]
	if !exists {
		return fmt.Errorf("mapping %s not found", mappingID)
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "mapping_type":
			if mappingType, ok := value.(MappingType); ok {
				mapping.MappingType = mappingType
			}
		case "confidence":
			if confidence, ok := value.(float64); ok {
				mapping.Confidence = confidence
				s.confidence[mappingID] = confidence
			}
		case "notes":
			if notes, ok := value.(string); ok {
				mapping.Notes = notes
			}
		}
	}

	mapping.UpdatedAt = time.Now()

	s.logger.Info("Framework mapping updated successfully",
		"request_id", requestID,
		"mapping_id", mappingID,
	)

	return nil
}

// DeleteMapping deletes a framework mapping
func (s *FrameworkMappingSystem) DeleteMapping(ctx context.Context, mappingID string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Deleting framework mapping",
		"request_id", requestID,
		"mapping_id", mappingID,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.mappings[mappingID]; !exists {
		return fmt.Errorf("mapping %s not found", mappingID)
	}

	delete(s.mappings, mappingID)
	delete(s.confidence, mappingID)

	s.logger.Info("Framework mapping deleted successfully",
		"request_id", requestID,
		"mapping_id", mappingID,
	)

	return nil
}

// CreateCrosswalk creates a crosswalk between frameworks
func (s *FrameworkMappingSystem) CreateCrosswalk(ctx context.Context, crosswalk *CrosswalkMapping) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Creating framework crosswalk",
		"request_id", requestID,
		"crosswalk_name", crosswalk.Name,
		"source_framework", crosswalk.SourceFramework,
		"target_framework", crosswalk.TargetFramework,
		"mapping_type", crosswalk.MappingType,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate frameworks exist
	if _, exists := s.frameworks[crosswalk.SourceFramework]; !exists {
		return fmt.Errorf("source framework %s not found", crosswalk.SourceFramework)
	}
	if _, exists := s.frameworks[crosswalk.TargetFramework]; !exists {
		return fmt.Errorf("target framework %s not found", crosswalk.TargetFramework)
	}

	// Generate crosswalk ID if not provided
	if crosswalk.ID == "" {
		crosswalk.ID = fmt.Sprintf("crosswalk_%s_%s", crosswalk.SourceFramework, crosswalk.TargetFramework)
	}

	// Set timestamps
	now := time.Now()
	if crosswalk.CreatedAt.IsZero() {
		crosswalk.CreatedAt = now
	}
	crosswalk.UpdatedAt = now

	s.crosswalks[crosswalk.ID] = crosswalk

	s.logger.Info("Framework crosswalk created successfully",
		"request_id", requestID,
		"crosswalk_id", crosswalk.ID,
		"crosswalk_name", crosswalk.Name,
		"mapping_count", len(crosswalk.Mappings),
	)

	return nil
}

// GetCrosswalk gets a specific crosswalk
func (s *FrameworkMappingSystem) GetCrosswalk(ctx context.Context, crosswalkID string) (*CrosswalkMapping, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting framework crosswalk",
		"request_id", requestID,
		"crosswalk_id", crosswalkID,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	crosswalk, exists := s.crosswalks[crosswalkID]
	if !exists {
		return nil, fmt.Errorf("crosswalk %s not found", crosswalkID)
	}

	return crosswalk, nil
}

// GetFramework returns a registered regulatory framework by ID
func (s *FrameworkMappingSystem) GetFramework(ctx context.Context, frameworkID string) (*RegulatoryFramework, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting regulatory framework",
		"request_id", requestID,
		"framework_id", frameworkID,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	fw, ok := s.frameworks[frameworkID]
	if !ok {
		return nil, fmt.Errorf("framework %s not found", frameworkID)
	}
	return fw, nil
}

// AnalyzeFrameworkMapping analyzes mappings between frameworks
func (s *FrameworkMappingSystem) AnalyzeFrameworkMapping(ctx context.Context, sourceFramework, targetFramework string) (*MappingAnalysis, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Analyzing framework mapping",
		"request_id", requestID,
		"source_framework", sourceFramework,
		"target_framework", targetFramework,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get mappings between frameworks
	mappings, err := s.GetMappingsBetweenFrameworks(ctx, sourceFramework, targetFramework)
	if err != nil {
		return nil, err
	}

	// Get source framework requirements
	sourceFrameworkData, exists := s.frameworks[sourceFramework]
	if !exists {
		return nil, fmt.Errorf("source framework %s not found", sourceFramework)
	}

	// Get target framework requirements
	targetFrameworkData, exists := s.frameworks[targetFramework]
	if !exists {
		return nil, fmt.Errorf("target framework %s not found", targetFramework)
	}

	// Analyze mappings
	analysis := &MappingAnalysis{
		ID:              fmt.Sprintf("analysis_%s_%s_%d", sourceFramework, targetFramework, time.Now().Unix()),
		SourceFramework: sourceFramework,
		TargetFramework: targetFramework,
		AnalysisDate:    time.Now(),
	}

	// Count mappings by type
	for _, mapping := range mappings {
		analysis.TotalMappings++
		analysis.AverageConfidence += mapping.Confidence

		switch mapping.MappingType {
		case MappingTypeExact:
			analysis.ExactMappings++
		case MappingTypePartial:
			analysis.PartialMappings++
		case MappingTypeRelated:
			analysis.RelatedMappings++
		case MappingTypeSuperseded:
			analysis.SupersededMappings++
		}
	}

	// Calculate average confidence
	if analysis.TotalMappings > 0 {
		analysis.AverageConfidence /= float64(analysis.TotalMappings)
	}

	// Calculate coverage percentage
	totalSourceRequirements := len(sourceFrameworkData.Requirements)
	if totalSourceRequirements > 0 {
		analysis.CoveragePercentage = float64(analysis.TotalMappings) / float64(totalSourceRequirements) * 100.0
	}

	// Identify gaps
	analysis.Gaps = s.identifyMappingGaps(sourceFrameworkData, targetFrameworkData, mappings)

	// Generate recommendations
	analysis.Recommendations = s.generateMappingRecommendations(analysis)

	s.logger.Info("Framework mapping analysis completed",
		"request_id", requestID,
		"source_framework", sourceFramework,
		"target_framework", targetFramework,
		"total_mappings", analysis.TotalMappings,
		"coverage_percentage", analysis.CoveragePercentage,
		"gaps_count", len(analysis.Gaps),
		"recommendations_count", len(analysis.Recommendations),
	)

	return analysis, nil
}

// AutoMapFrameworks automatically maps frameworks based on similarity
func (s *FrameworkMappingSystem) AutoMapFrameworks(ctx context.Context, sourceFramework, targetFramework string, confidenceThreshold float64) ([]*FrameworkMapping, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Auto-mapping frameworks",
		"request_id", requestID,
		"source_framework", sourceFramework,
		"target_framework", targetFramework,
		"confidence_threshold", confidenceThreshold,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	sourceFrameworkData, exists := s.frameworks[sourceFramework]
	if !exists {
		return nil, fmt.Errorf("source framework %s not found", sourceFramework)
	}

	targetFrameworkData, exists := s.frameworks[targetFramework]
	if !exists {
		return nil, fmt.Errorf("target framework %s not found", targetFramework)
	}

	var autoMappings []*FrameworkMapping

	// Compare requirements and create mappings
	for _, sourceReq := range sourceFrameworkData.Requirements {
		bestMatch := s.findBestRequirementMatch(&sourceReq, targetFrameworkData.Requirements, confidenceThreshold)
		if bestMatch != nil {
			mapping := &FrameworkMapping{
				ID:                  fmt.Sprintf("auto_mapping_%s_%s_%s_%s", sourceFramework, sourceReq.RequirementID, targetFramework, bestMatch.Requirement.RequirementID),
				SourceFramework:     sourceFramework,
				SourceRequirementID: sourceReq.RequirementID,
				TargetFramework:     targetFramework,
				TargetRequirementID: bestMatch.Requirement.RequirementID,
				MappingType:         MappingTypePartial, // Auto-mapped are typically partial
				Confidence:          bestMatch.Confidence,
				Notes:               fmt.Sprintf("Auto-generated mapping based on similarity analysis"),
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			}
			autoMappings = append(autoMappings, mapping)
		}
	}

	s.logger.Info("Auto-mapping completed",
		"request_id", requestID,
		"source_framework", sourceFramework,
		"target_framework", targetFramework,
		"auto_mappings_count", len(autoMappings),
	)

	return autoMappings, nil
}

// Helper methods
func (s *FrameworkMappingSystem) identifyMappingGaps(sourceFramework, targetFramework *RegulatoryFramework, mappings []*FrameworkMapping) []MappingGap {
	var gaps []MappingGap

	// Create a map of mapped source requirements
	mappedSourceReqs := make(map[string]bool)
	for _, mapping := range mappings {
		mappedSourceReqs[mapping.SourceRequirementID] = true
	}

	// Identify unmapped source requirements
	for _, sourceReq := range sourceFramework.Requirements {
		if !mappedSourceReqs[sourceReq.RequirementID] {
			gap := MappingGap{
				ID:                     fmt.Sprintf("gap_%s_%s", sourceFramework.ID, sourceReq.RequirementID),
				SourceRequirementID:    sourceReq.RequirementID,
				SourceRequirementTitle: sourceReq.Title,
				GapType:                "unmapped",
				Severity:               s.calculateGapSeverity(&sourceReq),
				Description:            fmt.Sprintf("Requirement %s from %s has no mapping to %s", sourceReq.RequirementID, sourceFramework.Name, targetFramework.Name),
				Impact:                 "Potential compliance gap",
				Recommendation:         "Review and create appropriate mapping or mark as not applicable",
			}
			gaps = append(gaps, gap)
		}
	}

	return gaps
}

func (s *FrameworkMappingSystem) generateMappingRecommendations(analysis *MappingAnalysis) []MappingRecommendation {
	var recommendations []MappingRecommendation

	// Low coverage recommendation
	if analysis.CoveragePercentage < 50.0 {
		recommendations = append(recommendations, MappingRecommendation{
			ID:          fmt.Sprintf("rec_%s_low_coverage", analysis.ID),
			Type:        "create",
			Priority:    "high",
			Title:       "Low Framework Coverage",
			Description: fmt.Sprintf("Only %.1f%% of requirements are mapped between frameworks", analysis.CoveragePercentage),
			Action:      "Create additional mappings to improve coverage",
			Impact:      "High - Improves compliance understanding and reduces gaps",
			Effort:      "Medium - Requires manual review and mapping",
			Timeline:    "2-4 weeks",
			Status:      "open",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	// Low confidence recommendation
	if analysis.AverageConfidence < 0.7 {
		recommendations = append(recommendations, MappingRecommendation{
			ID:          fmt.Sprintf("rec_%s_low_confidence", analysis.ID),
			Type:        "update",
			Priority:    "medium",
			Title:       "Low Mapping Confidence",
			Description: fmt.Sprintf("Average mapping confidence is %.1f%%", analysis.AverageConfidence*100),
			Action:      "Review and improve low-confidence mappings",
			Impact:      "Medium - Improves mapping accuracy",
			Effort:      "Low - Review existing mappings",
			Timeline:    "1-2 weeks",
			Status:      "open",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	// Gap remediation recommendations
	for _, gap := range analysis.Gaps {
		recommendations = append(recommendations, MappingRecommendation{
			ID:          fmt.Sprintf("rec_%s_gap_%s", analysis.ID, gap.ID),
			Type:        "create",
			Priority:    gap.Severity,
			Title:       fmt.Sprintf("Address Mapping Gap: %s", gap.SourceRequirementTitle),
			Description: gap.Description,
			Action:      gap.Recommendation,
			Impact:      gap.Impact,
			Effort:      "Medium - Requires requirement analysis",
			Timeline:    "1-3 weeks",
			Status:      "open",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	return recommendations
}

func (s *FrameworkMappingSystem) findBestRequirementMatch(sourceReq *ComplianceRequirement, targetReqs []ComplianceRequirement, confidenceThreshold float64) *struct {
	Requirement *ComplianceRequirement
	Confidence  float64
} {
	var bestMatch *struct {
		Requirement *ComplianceRequirement
		Confidence  float64
	}
	bestConfidence := 0.0

	for i := range targetReqs {
		confidence := s.calculateRequirementSimilarity(sourceReq, &targetReqs[i])
		if confidence > bestConfidence && confidence >= confidenceThreshold {
			bestConfidence = confidence
			bestMatch = &struct {
				Requirement *ComplianceRequirement
				Confidence  float64
			}{
				Requirement: &targetReqs[i],
				Confidence:  confidence,
			}
		}
	}

	return bestMatch
}

func (s *FrameworkMappingSystem) calculateRequirementSimilarity(req1, req2 *ComplianceRequirement) float64 {
	// Simple similarity calculation based on title and description
	// In a real implementation, this would use more sophisticated NLP techniques

	similarity := 0.0

	// Title similarity (simple string matching)
	if req1.Title == req2.Title {
		similarity += 0.4
	} else if s.stringContains(req1.Title, req2.Title) || s.stringContains(req2.Title, req1.Title) {
		similarity += 0.2
	}

	// Description similarity
	if req1.Description == req2.Description {
		similarity += 0.3
	} else if s.stringContains(req1.Description, req2.Description) || s.stringContains(req2.Description, req1.Description) {
		similarity += 0.15
	}

	// Category similarity
	if req1.Category == req2.Category {
		similarity += 0.2
	}

	// Risk level similarity
	if req1.RiskLevel == req2.RiskLevel {
		similarity += 0.1
	}

	return similarity
}

func (s *FrameworkMappingSystem) stringContains(str1, str2 string) bool {
	// Simple string containment check
	// In a real implementation, this would use more sophisticated text analysis
	return len(str2) > 0 && len(str1) >= len(str2) &&
		(len(str1) == len(str2) && str1 == str2 ||
			len(str1) > len(str2) && (str1[:len(str2)] == str2 || str1[len(str1)-len(str2):] == str2))
}

func (s *FrameworkMappingSystem) calculateGapSeverity(requirement *ComplianceRequirement) string {
	// Calculate gap severity based on requirement risk level and priority
	switch requirement.RiskLevel {
	case ComplianceRiskLevelCritical:
		return "critical"
	case ComplianceRiskLevelHigh:
		return "high"
	case ComplianceRiskLevelMedium:
		return "medium"
	case ComplianceRiskLevelLow:
		return "low"
	default:
		return "medium"
	}
}
