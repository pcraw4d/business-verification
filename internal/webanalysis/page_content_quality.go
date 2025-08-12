package webanalysis

import (
	"strings"
	"time"
)

// PageContentQuality represents comprehensive content quality assessment
type PageContentQuality struct {
	OverallQuality      float64                 `json:"overall_quality"`
	ComponentScores     map[string]float64      `json:"component_scores"`
	QualityFactors      []QualityFactor         `json:"quality_factors"`
	ReadabilityMetrics  ReadabilityMetrics      `json:"readability_metrics"`
	StructureMetrics    StructureMetrics        `json:"structure_metrics"`
	CompletenessMetrics CompletenessMetrics     `json:"completeness_metrics"`
	BusinessMetrics     BusinessContentMetrics  `json:"business_metrics"`
	TechnicalMetrics    TechnicalContentMetrics `json:"technical_metrics"`
	AssessedAt          time.Time               `json:"assessed_at"`
}

// QualityFactor represents a specific quality assessment factor
type QualityFactor struct {
	Factor     string  `json:"factor"`
	Score      float64 `json:"score"`
	Weight     float64 `json:"weight"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
	Details    string  `json:"details"`
}

// ReadabilityMetrics represents readability assessment metrics
type ReadabilityMetrics struct {
	FleschReadingEase     float64 `json:"flesch_reading_ease"`
	FleschKincaidGrade    float64 `json:"flesch_kincaid_grade"`
	GunningFogIndex       float64 `json:"gunning_fog_index"`
	SMOGIndex             float64 `json:"smog_index"`
	AverageSentenceLength float64 `json:"average_sentence_length"`
	AverageWordLength     float64 `json:"average_word_length"`
	ComplexWordRatio      float64 `json:"complex_word_ratio"`
	ReadabilityScore      float64 `json:"readability_score"`
}

// StructureMetrics represents content structure assessment
type StructureMetrics struct {
	HeadingStructure    float64 `json:"heading_structure"`
	ParagraphStructure  float64 `json:"paragraph_structure"`
	ListStructure       float64 `json:"list_structure"`
	TableStructure      float64 `json:"table_structure"`
	NavigationStructure float64 `json:"navigation_structure"`
	ContentOrganization float64 `json:"content_organization"`
	LogicalFlow         float64 `json:"logical_flow"`
	StructureScore      float64 `json:"structure_score"`
}

// CompletenessMetrics represents content completeness assessment
type CompletenessMetrics struct {
	ContentLength       int     `json:"content_length"`
	InformationDensity  float64 `json:"information_density"`
	FactualContent      float64 `json:"factual_content"`
	DescriptiveContent  float64 `json:"descriptive_content"`
	ContactInformation  float64 `json:"contact_information"`
	BusinessInformation float64 `json:"business_information"`
	ServiceInformation  float64 `json:"service_information"`
	CompletenessScore   float64 `json:"completeness_score"`
}

// BusinessContentMetrics represents business-specific content quality
type BusinessContentMetrics struct {
	BusinessNamePresence float64 `json:"business_name_presence"`
	ServiceDescription   float64 `json:"service_description"`
	CompanyHistory       float64 `json:"company_history"`
	TeamInformation      float64 `json:"team_information"`
	Certifications       float64 `json:"certifications"`
	Testimonials         float64 `json:"testimonials"`
	CaseStudies          float64 `json:"case_studies"`
	BusinessScore        float64 `json:"business_score"`
}

// TechnicalContentMetrics represents technical content quality
type TechnicalContentMetrics struct {
	HTMLValidity        float64 `json:"html_validity"`
	AccessibilityScore  float64 `json:"accessibility_score"`
	MobileOptimization  float64 `json:"mobile_optimization"`
	ImageOptimization   float64 `json:"image_optimization"`
	LinkQuality         float64 `json:"link_quality"`
	MetaTagCompleteness float64 `json:"meta_tag_completeness"`
	TechnicalScore      float64 `json:"technical_score"`
}

// PageContentQualityAssessor manages content quality assessment
type PageContentQualityAssessor struct {
	config               ContentQualityConfig
	readabilityAnalyzer  *ReadabilityAnalyzer
	structureAnalyzer    *StructureAnalyzer
	completenessAnalyzer *CompletenessAnalyzer
	businessAnalyzer     *BusinessContentAnalyzer
	technicalAnalyzer    *TechnicalContentAnalyzer
}

// ContentQualityConfig holds configuration for content quality assessment
type ContentQualityConfig struct {
	Weights                map[string]float64 `json:"weights"`
	QualityThresholds      map[string]float64 `json:"quality_thresholds"`
	ReadabilityThresholds  map[string]float64 `json:"readability_thresholds"`
	StructureThresholds    map[string]float64 `json:"structure_thresholds"`
	CompletenessThresholds map[string]float64 `json:"completeness_thresholds"`
	BusinessThresholds     map[string]float64 `json:"business_thresholds"`
	TechnicalThresholds    map[string]float64 `json:"technical_thresholds"`
}

// NewPageContentQualityAssessor creates a new content quality assessor
func NewPageContentQualityAssessor(config ContentQualityConfig) *PageContentQualityAssessor {
	return &PageContentQualityAssessor{
		config:               config,
		readabilityAnalyzer:  NewReadabilityAnalyzer(),
		structureAnalyzer:    NewStructureAnalyzer(),
		completenessAnalyzer: NewCompletenessAnalyzer(),
		businessAnalyzer:     NewBusinessContentAnalyzer(),
		technicalAnalyzer:    NewTechnicalContentAnalyzer(),
	}
}

// AssessContentQuality performs comprehensive content quality assessment
func (pcqa *PageContentQualityAssessor) AssessContentQuality(content *ScrapedContent, business string) *PageContentQuality {
	quality := &PageContentQuality{
		ComponentScores: make(map[string]float64),
		QualityFactors:  []QualityFactor{},
		AssessedAt:      time.Now(),
	}

	// Assess readability using existing methods
	quality.ReadabilityMetrics = pcqa.calculateReadabilityMetrics(content.Text)

	// Assess structure
	quality.StructureMetrics = pcqa.structureAnalyzer.AnalyzeStructure(content.HTML, content.Text)

	// Assess completeness
	quality.CompletenessMetrics = pcqa.completenessAnalyzer.AnalyzeCompleteness(content)

	// Assess business content
	quality.BusinessMetrics = pcqa.businessAnalyzer.AnalyzeBusinessContent(content, business)

	// Assess technical content
	quality.TechnicalMetrics = pcqa.technicalAnalyzer.AnalyzeTechnicalContent(content)

	// Calculate component scores
	pcqa.calculateComponentScores(quality)

	// Calculate overall quality
	quality.OverallQuality = pcqa.calculateOverallQuality(quality)

	return quality
}

// calculateComponentScores calculates individual component scores
func (pcqa *PageContentQualityAssessor) calculateComponentScores(quality *PageContentQuality) {
	// Readability score
	readabilityScore := quality.ReadabilityMetrics.ReadabilityScore
	quality.ComponentScores["readability"] = readabilityScore
	quality.QualityFactors = append(quality.QualityFactors, QualityFactor{
		Factor:     "readability",
		Score:      readabilityScore,
		Weight:     pcqa.config.Weights["readability"],
		Confidence: 0.9,
		Reason:     "Content readability and comprehension",
		Details:    "Based on Flesch Reading Ease and other readability metrics",
	})

	// Structure score
	structureScore := quality.StructureMetrics.StructureScore
	quality.ComponentScores["structure"] = structureScore
	quality.QualityFactors = append(quality.QualityFactors, QualityFactor{
		Factor:     "structure",
		Score:      structureScore,
		Weight:     pcqa.config.Weights["structure"],
		Confidence: 0.8,
		Reason:     "Content organization and logical flow",
		Details:    "Based on heading structure, paragraph organization, and navigation",
	})

	// Completeness score
	completenessScore := quality.CompletenessMetrics.CompletenessScore
	quality.ComponentScores["completeness"] = completenessScore
	quality.QualityFactors = append(quality.QualityFactors, QualityFactor{
		Factor:     "completeness",
		Score:      completenessScore,
		Weight:     pcqa.config.Weights["completeness"],
		Confidence: 0.7,
		Reason:     "Content completeness and information density",
		Details:    "Based on content length, factual information, and business details",
	})

	// Business content score
	businessScore := quality.BusinessMetrics.BusinessScore
	quality.ComponentScores["business_content"] = businessScore
	quality.QualityFactors = append(quality.QualityFactors, QualityFactor{
		Factor:     "business_content",
		Score:      businessScore,
		Weight:     pcqa.config.Weights["business_content"],
		Confidence: 0.8,
		Reason:     "Business-specific content quality",
		Details:    "Based on business information, services, and professional indicators",
	})

	// Technical content score
	technicalScore := quality.TechnicalMetrics.TechnicalScore
	quality.ComponentScores["technical_content"] = technicalScore
	quality.QualityFactors = append(quality.QualityFactors, QualityFactor{
		Factor:     "technical_content",
		Score:      technicalScore,
		Weight:     pcqa.config.Weights["technical_content"],
		Confidence: 0.6,
		Reason:     "Technical content quality and optimization",
		Details:    "Based on HTML validity, accessibility, and mobile optimization",
	})
}

// calculateOverallQuality calculates the weighted overall quality score
func (pcqa *PageContentQualityAssessor) calculateOverallQuality(quality *PageContentQuality) float64 {
	overallQuality := 0.0
	totalWeight := 0.0

	for factor, componentScore := range quality.ComponentScores {
		weight := pcqa.config.Weights[factor]
		overallQuality += componentScore * weight
		totalWeight += weight
	}

	if totalWeight > 0 {
		return overallQuality / totalWeight
	}
	return 0.0
}

// calculateReadabilityMetrics calculates comprehensive readability metrics
func (pcqa *PageContentQualityAssessor) calculateReadabilityMetrics(text string) ReadabilityMetrics {
	metrics := ReadabilityMetrics{}

	// Use existing readability analyzer methods
	metrics.ReadabilityScore = pcqa.readabilityAnalyzer.CalculateReadability(text)

	// Calculate additional readability metrics
	metrics.FleschReadingEase = pcqa.calculateFleschReadingEase(text)
	metrics.FleschKincaidGrade = pcqa.calculateFleschKincaidGrade(text)
	metrics.GunningFogIndex = pcqa.calculateGunningFogIndex(text)
	metrics.SMOGIndex = pcqa.calculateSMOGIndex(text)
	metrics.AverageSentenceLength = pcqa.calculateAverageSentenceLength(text)
	metrics.AverageWordLength = pcqa.calculateAverageWordLength(text)
	metrics.ComplexWordRatio = pcqa.calculateComplexWordRatio(text)

	return metrics
}

// calculateFleschReadingEase calculates Flesch Reading Ease score
func (pcqa *PageContentQualityAssessor) calculateFleschReadingEase(text string) float64 {
	// Simplified Flesch Reading Ease calculation
	sentences := len(strings.Split(text, "."))
	words := len(strings.Fields(text))
	syllables := pcqa.countSyllables(text)

	if sentences == 0 || words == 0 {
		return 0.0
	}

	// Flesch Reading Ease formula: 206.835 - 1.015 × (total words ÷ total sentences) - 84.6 × (total syllables ÷ total words)
	score := 206.835 - 1.015*float64(words)/float64(sentences) - 84.6*float64(syllables)/float64(words)

	// Clamp to 0-100 range
	if score < 0 {
		return 0.0
	}
	if score > 100 {
		return 100.0
	}
	return score
}

// calculateFleschKincaidGrade calculates Flesch-Kincaid Grade Level
func (pcqa *PageContentQualityAssessor) calculateFleschKincaidGrade(text string) float64 {
	// Simplified Flesch-Kincaid Grade calculation
	sentences := len(strings.Split(text, "."))
	words := len(strings.Fields(text))
	syllables := pcqa.countSyllables(text)

	if sentences == 0 || words == 0 {
		return 0.0
	}

	// Flesch-Kincaid Grade formula: 0.39 × (total words ÷ total sentences) + 11.8 × (total syllables ÷ total words) - 15.59
	grade := 0.39*float64(words)/float64(sentences) + 11.8*float64(syllables)/float64(words) - 15.59

	// Clamp to reasonable range
	if grade < 0 {
		return 0.0
	}
	if grade > 20 {
		return 20.0
	}
	return grade
}

// calculateGunningFogIndex calculates Gunning Fog Index
func (pcqa *PageContentQualityAssessor) calculateGunningFogIndex(text string) float64 {
	// Simplified Gunning Fog Index calculation
	sentences := len(strings.Split(text, "."))
	words := len(strings.Fields(text))
	complexWords := pcqa.countComplexWords(text)

	if sentences == 0 || words == 0 {
		return 0.0
	}

	// Gunning Fog Index formula: 0.4 × [(words ÷ sentences) + 100 × (complex words ÷ words)]
	index := 0.4 * (float64(words)/float64(sentences) + 100*float64(complexWords)/float64(words))

	// Clamp to reasonable range
	if index < 0 {
		return 0.0
	}
	if index > 20 {
		return 20.0
	}
	return index
}

// calculateSMOGIndex calculates SMOG Index
func (pcqa *PageContentQualityAssessor) calculateSMOGIndex(text string) float64 {
	// Simplified SMOG Index calculation
	sentences := len(strings.Split(text, "."))
	complexWords := pcqa.countComplexWords(text)

	if sentences == 0 {
		return 0.0
	}

	// SMOG Index formula: 1.043 × √(complex words × 30 ÷ sentences) + 3.1291
	index := 1.043*float64(complexWords)*30/float64(sentences) + 3.1291

	// Clamp to reasonable range
	if index < 0 {
		return 0.0
	}
	if index > 20 {
		return 20.0
	}
	return index
}

// calculateAverageSentenceLength calculates average sentence length
func (pcqa *PageContentQualityAssessor) calculateAverageSentenceLength(text string) float64 {
	sentences := strings.Split(text, ".")
	words := len(strings.Fields(text))

	if len(sentences) == 0 {
		return 0.0
	}

	return float64(words) / float64(len(sentences))
}

// calculateAverageWordLength calculates average word length
func (pcqa *PageContentQualityAssessor) calculateAverageWordLength(text string) float64 {
	words := strings.Fields(text)
	if len(words) == 0 {
		return 0.0
	}

	totalLength := 0
	for _, word := range words {
		totalLength += len(word)
	}

	return float64(totalLength) / float64(len(words))
}

// calculateComplexWordRatio calculates ratio of complex words
func (pcqa *PageContentQualityAssessor) calculateComplexWordRatio(text string) float64 {
	words := strings.Fields(text)
	if len(words) == 0 {
		return 0.0
	}

	complexWords := pcqa.countComplexWords(text)
	return float64(complexWords) / float64(len(words))
}

// countSyllables counts syllables in text (simplified)
func (pcqa *PageContentQualityAssessor) countSyllables(text string) int {
	// Simplified syllable counting
	words := strings.Fields(text)
	totalSyllables := 0

	for _, word := range words {
		// Simple heuristic: count vowel groups
		vowelGroups := 0
		prevVowel := false

		for _, char := range strings.ToLower(word) {
			isVowel := char == 'a' || char == 'e' || char == 'i' || char == 'o' || char == 'u' || char == 'y'
			if isVowel && !prevVowel {
				vowelGroups++
			}
			prevVowel = isVowel
		}

		// Ensure at least one syllable per word
		if vowelGroups == 0 {
			vowelGroups = 1
		}

		totalSyllables += vowelGroups
	}

	return totalSyllables
}

// countComplexWords counts complex words (3+ syllables)
func (pcqa *PageContentQualityAssessor) countComplexWords(text string) int {
	words := strings.Fields(text)
	complexCount := 0

	for _, word := range words {
		syllables := pcqa.countSyllables(word)
		if syllables >= 3 {
			complexCount++
		}
	}

	return complexCount
}
