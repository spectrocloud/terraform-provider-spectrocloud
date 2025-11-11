package spectrocloud

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func TestValidateClusterProfileAndTemplate(t *testing.T) {
	tests := []struct {
		name          string
		profile       []interface{}
		template      string
		expectedError bool
		errorContains string
	}{
		{
			name:          "both profile and template specified - should error",
			profile:       []interface{}{map[string]interface{}{"id": "profile-123"}},
			template:      "template-456",
			expectedError: true,
			errorContains: "mutually exclusive",
		},
		{
			name:          "only profile specified - should pass",
			profile:       []interface{}{map[string]interface{}{"id": "profile-123"}},
			template:      "",
			expectedError: false,
		},
		{
			name:          "only template specified - should pass",
			profile:       []interface{}{},
			template:      "template-456",
			expectedError: false,
		},
		{
			name:          "neither profile nor template specified - should error",
			profile:       []interface{}{},
			template:      "",
			expectedError: true,
			errorContains: "either cluster_profile or cluster_template must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create resource data with minimal required fields
			resourceData := map[string]interface{}{
				"name":             "test-cluster",
				"cloud_account_id": "test-account",
				"cloud_config": []interface{}{
					map[string]interface{}{
						"region": "us-west-2",
					},
				},
			}

			// Add cluster_profile if provided
			if len(tt.profile) > 0 {
				resourceData["cluster_profile"] = tt.profile
			}

			// Add cluster_template if provided
			if tt.template != "" {
				resourceData["cluster_template"] = tt.template
			}

			d := schema.TestResourceDataRaw(t, resourceClusterEks().Schema, resourceData)

			err := validateClusterProfileAndTemplate(d)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateClusterTemplateUpdate(t *testing.T) {
	// Note: This test validates the logic, but testing HasChange() properly requires
	// integration tests or acceptance tests. For unit tests, we validate no error
	// when there's no change detected.
	t.Run("validates without change detection", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, resourceClusterEks().Schema, map[string]interface{}{
			"name":             "test-cluster",
			"cloud_account_id": "test-account",
			"cluster_template": "template-123",
			"cloud_config": []interface{}{
				map[string]interface{}{
					"region": "us-west-2",
				},
			},
		})

		// Without proper diff tracking, HasChange will return false
		// This validates that the function doesn't error when no change is detected
		err := validateClusterTemplateUpdate(d)
		if err != nil {
			t.Errorf("unexpected error when no change detected: %v", err)
		}
	})
}

func TestToClusterTemplate(t *testing.T) {
	tests := []struct {
		name           string
		templateUID    string
		expectedResult *models.V1ClusterTemplateRef
	}{
		{
			name:        "with template UID",
			templateUID: "template-123",
			expectedResult: &models.V1ClusterTemplateRef{
				UID: "template-123",
			},
		},
		{
			name:           "empty template UID",
			templateUID:    "",
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceData := map[string]interface{}{
				"name":             "test-cluster",
				"cloud_account_id": "test-account",
				"cloud_config": []interface{}{
					map[string]interface{}{
						"region": "us-west-2",
					},
				},
			}

			if tt.templateUID != "" {
				resourceData["cluster_template"] = tt.templateUID
			}

			d := schema.TestResourceDataRaw(t, resourceClusterEks().Schema, resourceData)

			result := toClusterTemplate(d)

			if tt.expectedResult == nil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
			} else {
				if result == nil {
					t.Errorf("expected result, got nil")
				} else if result.UID != tt.expectedResult.UID {
					t.Errorf("expected UID '%s', got '%s'", tt.expectedResult.UID, result.UID)
				}
			}
		})
	}
}

func TestFlattenClusterTemplate(t *testing.T) {
	tests := []struct {
		name            string
		clusterTemplate *models.V1SpectroClusterTemplateRef
		expectedResult  string
	}{
		{
			name: "with template UID",
			clusterTemplate: &models.V1SpectroClusterTemplateRef{
				UID: "template-123",
			},
			expectedResult: "template-123",
		},
		{
			name:            "nil template",
			clusterTemplate: nil,
			expectedResult:  "",
		},
		{
			name: "empty template UID",
			clusterTemplate: &models.V1SpectroClusterTemplateRef{
				UID: "",
			},
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenClusterTemplate(tt.clusterTemplate)

			if result != tt.expectedResult {
				t.Errorf("expected '%s', got '%s'", tt.expectedResult, result)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
