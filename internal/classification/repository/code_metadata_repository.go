package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"kyb-platform/internal/database"
)

// CodeMetadataRepository provides access to the code_metadata table
// This table contains enhanced metadata including official descriptions,
// crosswalks, hierarchies, and industry mappings
type CodeMetadataRepository struct {
	client *database.SupabaseClient
	logger *log.Logger
}

// NewCodeMetadataRepository creates a new code metadata repository
func NewCodeMetadataRepository(client *database.SupabaseClient, logger *log.Logger) *CodeMetadataRepository {
	if logger == nil {
		logger = log.Default()
	}
	return &CodeMetadataRepository{
		client: client,
		logger: logger,
	}
}

// CodeMetadata represents a record from the code_metadata table
type CodeMetadata struct {
	ID                 string                 `json:"id"`
	CodeType           string                 `json:"code_type"`            // "NAICS", "SIC", or "MCC"
	Code               string                 `json:"code"`                 // The actual code (e.g., "541511")
	OfficialName       string                 `json:"official_name"`        // Official name from source
	OfficialDescription string                `json:"official_description"` // Official description
	OfficialCategory   string                 `json:"official_category,omitempty"`
	IndustryMappings   map[string]interface{} `json:"industry_mappings"`     // JSONB: primary/secondary industries
	CrosswalkData      map[string]interface{} `json:"crosswalk_data"`        // JSONB: links to other code types
	Hierarchy          map[string]interface{} `json:"hierarchy"`             // JSONB: parent/child relationships
	Metadata           map[string]interface{} `json:"metadata,omitempty"`   // JSONB: additional metadata
	IsOfficial         bool                    `json:"is_official"`
	IsActive           bool                    `json:"is_active"`
	CreatedAt          string                  `json:"created_at"`
	UpdatedAt          string                  `json:"updated_at"`
}

// GetCodeMetadata retrieves metadata for a specific code
func (r *CodeMetadataRepository) GetCodeMetadata(ctx context.Context, codeType, code string) (*CodeMetadata, error) {
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}

	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return nil, fmt.Errorf("PostgREST client not available")
	}

	response, _, err := postgrestClient.
		From("code_metadata").
		Select("*", "", false).
		Eq("code_type", codeType).
		Eq("code", code).
		Eq("is_active", "true").
		Single().
		Execute()

	if err != nil {
		// Not found is not an error - code may not have metadata
		if strings.Contains(err.Error(), "No rows") || strings.Contains(err.Error(), "not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get code metadata: %w", err)
	}

	var metadata CodeMetadata
	if err := json.Unmarshal(response, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal code metadata: %w", err)
	}

	return &metadata, nil
}

// GetCodeMetadataBatch retrieves metadata for multiple codes
func (r *CodeMetadataRepository) GetCodeMetadataBatch(ctx context.Context, codes []struct{ CodeType, Code string }) (map[string]*CodeMetadata, error) {
	result := make(map[string]*CodeMetadata)

	// Query each code (PostgREST doesn't support complex IN queries easily)
	for _, code := range codes {
		metadata, err := r.GetCodeMetadata(ctx, code.CodeType, code.Code)
		if err != nil {
			r.logger.Printf("⚠️ Failed to get metadata for %s %s: %v", code.CodeType, code.Code, err)
			continue
		}
		if metadata != nil {
			key := fmt.Sprintf("%s:%s", code.CodeType, code.Code)
			result[key] = metadata
		}
	}

	return result, nil
}

// GetCrosswalkCodes retrieves related codes from other classification systems
// For example, given NAICS 541511, returns related SIC and MCC codes
func (r *CodeMetadataRepository) GetCrosswalkCodes(ctx context.Context, codeType, code string) ([]struct {
	CodeType string
	Code     string
	Name     string
}, error) {
	metadata, err := r.GetCodeMetadata(ctx, codeType, code)
	if err != nil || metadata == nil {
		return nil, err
	}

	var crosswalkCodes []struct {
		CodeType string
		Code     string
		Name     string
	}

	// Extract crosswalk data
	if crosswalkData, ok := metadata.CrosswalkData["naics"].([]interface{}); ok {
		for _, naicsCode := range crosswalkData {
			if codeStr, ok := naicsCode.(string); ok {
				// Get metadata for the crosswalk code to get official name
				naicsMeta, _ := r.GetCodeMetadata(ctx, "NAICS", codeStr)
				name := ""
				if naicsMeta != nil {
					name = naicsMeta.OfficialName
				}
				crosswalkCodes = append(crosswalkCodes, struct {
					CodeType string
					Code     string
					Name     string
				}{"NAICS", codeStr, name})
			}
		}
	}

	if crosswalkData, ok := metadata.CrosswalkData["sic"].([]interface{}); ok {
		for _, sicCode := range crosswalkData {
			if codeStr, ok := sicCode.(string); ok {
				sicMeta, _ := r.GetCodeMetadata(ctx, "SIC", codeStr)
				name := ""
				if sicMeta != nil {
					name = sicMeta.OfficialName
				}
				crosswalkCodes = append(crosswalkCodes, struct {
					CodeType string
					Code     string
					Name     string
				}{"SIC", codeStr, name})
			}
		}
	}

	if crosswalkData, ok := metadata.CrosswalkData["mcc"].([]interface{}); ok {
		for _, mccCode := range crosswalkData {
			if codeStr, ok := mccCode.(string); ok {
				mccMeta, _ := r.GetCodeMetadata(ctx, "MCC", codeStr)
				name := ""
				if mccMeta != nil {
					name = mccMeta.OfficialName
				}
				crosswalkCodes = append(crosswalkCodes, struct {
					CodeType string
					Code     string
					Name     string
				}{"MCC", codeStr, name})
			}
		}
	}

	return crosswalkCodes, nil
}

// GetHierarchyCodes retrieves parent and child codes for a given code (mainly for NAICS)
func (r *CodeMetadataRepository) GetHierarchyCodes(ctx context.Context, codeType, code string) (parent *CodeMetadata, children []*CodeMetadata, err error) {
	metadata, err := r.GetCodeMetadata(ctx, codeType, code)
	if err != nil || metadata == nil {
		return nil, nil, err
	}

	// Get parent code if available
	if parentCode, ok := metadata.Hierarchy["parent_code"].(string); ok {
		if parentType, ok := metadata.Hierarchy["parent_type"].(string); ok {
			parent, _ = r.GetCodeMetadata(ctx, parentType, parentCode)
		}
	}

	// Get child codes if available
	if childCodes, ok := metadata.Hierarchy["child_codes"].([]interface{}); ok {
		for _, childCode := range childCodes {
			if codeStr, ok := childCode.(string); ok {
				childMeta, _ := r.GetCodeMetadata(ctx, codeType, codeStr)
				if childMeta != nil {
					children = append(children, childMeta)
				}
			}
		}
	}

	return parent, children, nil
}

// GetCodesByIndustryMapping retrieves codes that match a specific industry
func (r *CodeMetadataRepository) GetCodesByIndustryMapping(ctx context.Context, industryName string, codeType string) ([]*CodeMetadata, error) {
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}

	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return nil, fmt.Errorf("PostgREST client not available")
	}

	// Query using JSONB contains operator
	// This searches for industry in the industry_mappings JSONB field
	response, _, err := postgrestClient.
		From("code_metadata").
		Select("*", "", false).
		Eq("code_type", codeType).
		Eq("is_active", "true").
		// Note: PostgREST JSONB queries may need special handling
		// For now, we'll get all codes and filter in memory
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to query code metadata: %w", err)
	}

	var allMetadata []CodeMetadata
	if err := json.Unmarshal(response, &allMetadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal code metadata: %w", err)
	}

	// Filter by industry mapping
	var results []*CodeMetadata
	for i := range allMetadata {
		meta := &allMetadata[i]
		
		// Check primary industry
		if primary, ok := meta.IndustryMappings["primary_industry"].(string); ok {
			if strings.EqualFold(primary, industryName) {
				results = append(results, meta)
				continue
			}
		}
		
		// Check secondary industries
		if secondaries, ok := meta.IndustryMappings["secondary_industries"].([]interface{}); ok {
			for _, sec := range secondaries {
				if secStr, ok := sec.(string); ok {
					if strings.EqualFold(secStr, industryName) {
						results = append(results, meta)
						break
					}
				}
			}
		}
	}

	return results, nil
}

// EnhanceCodeDescription enhances a code's description with official description from metadata
// Returns the enhanced description or the original if no metadata is found
func (r *CodeMetadataRepository) EnhanceCodeDescription(ctx context.Context, codeType, code, originalDescription string) string {
	metadata, err := r.GetCodeMetadata(ctx, codeType, code)
	if err != nil || metadata == nil {
		return originalDescription // Return original if no metadata
	}

	// Prefer official description if available
	if metadata.OfficialDescription != "" {
		return metadata.OfficialDescription
	}

	// Fall back to official name if description is empty
	if metadata.OfficialName != "" && metadata.OfficialName != code {
		return metadata.OfficialName
	}

	return originalDescription
}

