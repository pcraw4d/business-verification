package classification

import (
	"context"
	"fmt"
	"strings"

	"kyb-platform/internal/classification/repository"
)

// CodeValidator validates code-industry relationships
type CodeValidator struct {
	repo            repository.KeywordRepository
	codeMetadataRepo *repository.CodeMetadataRepository
}

// NewCodeValidator creates a new code validator
func NewCodeValidator(repo repository.KeywordRepository, codeMetadataRepo *repository.CodeMetadataRepository) *CodeValidator {
	return &CodeValidator{
		repo:            repo,
		codeMetadataRepo: codeMetadataRepo,
	}
}

// ValidateCodeIndustryMatch validates if a code is appropriate for an industry
// Returns a confidence score (0.0-1.0) indicating how well the code matches the industry
// - 1.0: Code is directly associated with the industry
// - 0.8: Code is associated with a parent industry
// - 0.0: Code is not found or not associated
func (v *CodeValidator) ValidateCodeIndustryMatch(
	ctx context.Context,
	code string,
	codeType string,
	industryName string,
) (float64, error) {
	if code == "" || codeType == "" || industryName == "" {
		return 0.0, fmt.Errorf("code, codeType, and industryName are required")
	}

	// Normalize industry name for matching
	normalizedIndustry := strings.ToLower(strings.TrimSpace(industryName))

	// Step 1: Get industry by name
	industry, err := v.repo.GetIndustryByName(ctx, industryName)
	if err != nil {
		// Industry not found - try case-insensitive search
		industries, listErr := v.repo.ListIndustries(ctx, "")
		if listErr != nil {
			return 0.0, fmt.Errorf("failed to query industries: %w", listErr)
		}

		// Find industry by case-insensitive match
		for _, ind := range industries {
			if strings.EqualFold(ind.Name, industryName) {
				industry = ind
				break
			}
		}

		if industry == nil {
			// Industry not found - return low confidence
			return 0.0, nil
		}
	}

	// Step 2: Get codes for this industry
	codes, err := v.repo.GetClassificationCodesByIndustry(ctx, industry.ID)
	if err != nil {
		return 0.0, fmt.Errorf("failed to get codes for industry: %w", err)
	}

	// Step 3: Check if code exists in industry's code list
	for _, c := range codes {
		if c.CodeType == codeType && c.Code == code {
			// Exact match - full confidence
			return 1.0, nil
		}
	}

	// Step 4: Check code metadata repository for industry mappings
	if v.codeMetadataRepo != nil {
		// Try to find code in code_metadata with industry mapping
		metadataCodes, err := v.codeMetadataRepo.GetCodesByIndustryMapping(ctx, industryName, codeType)
		if err == nil {
			for _, metaCode := range metadataCodes {
				if metaCode.Code == code {
					// Found in industry mapping - full confidence
					return 1.0, nil
				}
			}
		}
	}

	// Step 5: Check parent industry (if industry has a category/parent)
	// For now, we'll check if the industry category matches
	// This is a simplified check - can be enhanced with actual parent industry lookup
	if industry.Category != "" {
		// Try to find codes in parent category
		categoryIndustries, err := v.repo.ListIndustries(ctx, industry.Category)
		if err == nil {
			for _, catIndustry := range categoryIndustries {
				if catIndustry.ID == industry.ID {
					continue // Skip self
				}
				catCodes, err := v.repo.GetClassificationCodesByIndustry(ctx, catIndustry.ID)
				if err == nil {
					for _, c := range catCodes {
						if c.CodeType == codeType && c.Code == code {
							// Found in parent category - partial confidence
							return 0.8, nil
						}
					}
				}
			}
		}
	}

	// Step 6: Check crosswalks - code might be valid through crosswalk relationship
	// This is a fallback - if code exists in crosswalks, it might be valid
	// For now, we'll return 0.0 if not found
	// This can be enhanced to check crosswalk relationships

	// Code not found in industry or parent - return 0.0
	return 0.0, nil
}

// ValidateCodesForIndustry validates multiple codes for an industry
// Returns a map of code -> confidence score
func (v *CodeValidator) ValidateCodesForIndustry(
	ctx context.Context,
	codes []string,
	codeType string,
	industryName string,
) (map[string]float64, error) {
	results := make(map[string]float64)

	for _, code := range codes {
		confidence, err := v.ValidateCodeIndustryMatch(ctx, code, codeType, industryName)
		if err != nil {
			// Log error but continue with other codes
			results[code] = 0.0
			continue
		}
		results[code] = confidence
	}

	return results, nil
}

