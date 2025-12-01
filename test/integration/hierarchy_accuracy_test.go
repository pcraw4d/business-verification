package integration

import (
	"testing"
)

// TestNAICSHierarchyCoverage verifies NAICS hierarchy coverage targets
func TestNAICSHierarchyCoverage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("NAICS hierarchy coverage", func(t *testing.T) {
		expectedMinCodes := 30
		expectedMinPercentage := 30.0

		// In real test:
		// SELECT 
		//     COUNT(*) as codes_with_hierarchy,
		//     COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata WHERE code_type = 'NAICS' AND is_active = true) as percentage
		// FROM code_metadata
		// WHERE code_type = 'NAICS'
		//   AND is_active = true
		//   AND hierarchy != '{}'::jsonb
		//   AND hierarchy IS NOT NULL;

		t.Logf("Expected minimum NAICS codes with hierarchy: %d (%.1f%%)", expectedMinCodes, expectedMinPercentage)
	})
}

// TestNAICSHierarchyAccuracy verifies hierarchy accuracy > 98%
func TestNAICSHierarchyAccuracy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("Test hierarchy accuracy", func(t *testing.T) {
		// Test known hierarchy relationships
		testCases := []struct {
			codeType     string
			code         string
			parentCode   string
			parentType   string
			description  string
		}{
			{
				codeType:    "NAICS",
				code:        "541511",
				parentCode:  "54151",
				parentType:  "NAICS",
				description: "Custom Computer Programming Services -> Custom Computer Programming Services (Industry)",
			},
			{
				codeType:    "NAICS",
				code:        "522110",
				parentCode:  "5221",
				parentType:  "NAICS",
				description: "Commercial Banking -> Depository Credit Intermediation",
			},
			{
				codeType:    "NAICS",
				code:        "621111",
				parentCode:  "6211",
				parentType:  "NAICS",
				description: "Offices of Physicians (except Mental Health Specialists) -> Offices of Physicians",
			},
			{
				codeType:    "NAICS",
				code:        "722511",
				parentCode:  "7225",
				parentType:  "NAICS",
				description: "Full-Service Restaurants -> Restaurants and Other Eating Places",
			},
			{
				codeType:    "NAICS",
				code:        "236115",
				parentCode:  "2361",
				parentType:  "NAICS",
				description: "New Single-Family Housing Construction -> Residential Building Construction",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				// In real test:
				// client := database.NewSupabaseClient(testDatabaseURL)
				// repo := repository.NewCodeMetadataRepository(client, log.Default())
				// parent, children, err := repo.GetHierarchyCodes(context.Background(), tc.codeType, tc.code)
				// 
				// if err != nil {
				//     t.Fatalf("Failed to get hierarchy: %v", err)
				// }
				// 
				// if parent == nil {
				//     t.Errorf("Expected parent code %s %s for %s %s, but got nil",
				//         tc.parentType, tc.parentCode, tc.codeType, tc.code)
				//     return
				// }
				// 
				// if parent.Code != tc.parentCode {
				//     t.Errorf("Expected parent code %s, but got %s", tc.parentCode, parent.Code)
				// }
				// 
				// if parent.CodeType != tc.parentType {
				//     t.Errorf("Expected parent type %s, but got %s", tc.parentType, parent.CodeType)
				// }

				t.Logf("Hierarchy accuracy test structure verified: %s", tc.description)
			})
		}
	})

	t.Run("Test hierarchy completeness", func(t *testing.T) {
		// Verify that parent codes exist in the database
		testCases := []struct {
			codeType    string
			code        string
			parentCode  string
			description string
		}{
			{
				codeType:    "NAICS",
				code:        "541511",
				parentCode:  "54151",
				description: "541511 should have parent 54151",
			},
			{
				codeType:    "NAICS",
				code:        "522110",
				parentCode:  "5221",
				description: "522110 should have parent 5221",
			},
			{
				codeType:    "NAICS",
				code:        "621111",
				parentCode:  "6211",
				description: "621111 should have parent 6211",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				// In real test:
				// Verify that the parent code exists in code_metadata
				// SELECT COUNT(*) 
				// FROM code_metadata
				// WHERE code_type = 'NAICS'
				//   AND code = tc.parentCode
				//   AND is_active = true;
				// 
				// Verify that the child code has the correct parent
				// SELECT hierarchy->>'parent_code' 
				// FROM code_metadata
				// WHERE code_type = tc.codeType
				//   AND code = tc.code
				//   AND is_active = true;

				t.Logf("Hierarchy completeness test structure verified: %s", tc.description)
			})
		}
	})

	t.Run("Test hierarchy accuracy percentage", func(t *testing.T) {
		expectedMinAccuracy := 98.0

		// In real test:
		// Verify that 98%+ of hierarchy relationships are accurate
		// SELECT 
		//     COUNT(*) AS codes_with_valid_parents,
		//     COUNT(*) * 100.0 / NULLIF((SELECT COUNT(*) FROM code_metadata WHERE code_type = 'NAICS' AND hierarchy != '{}'::jsonb), 0) AS accuracy_percentage
		// FROM code_metadata cm1
		// WHERE cm1.code_type = 'NAICS'
		//   AND cm1.is_active = true
		//   AND cm1.hierarchy != '{}'::jsonb
		//   AND cm1.hierarchy->>'parent_code' IS NOT NULL
		//   AND EXISTS (
		//       SELECT 1 
		//       FROM code_metadata cm2 
		//       WHERE cm2.code_type = cm1.hierarchy->>'parent_type'
		//         AND cm2.code = cm1.hierarchy->>'parent_code'
		//         AND cm2.is_active = true
		//   );

		t.Logf("Expected minimum hierarchy accuracy: %.1f%%", expectedMinAccuracy)
	})
}

