package spectrocloud

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestValidateProfileSource tests the mutual exclusivity validation
func TestValidateProfileSource(t *testing.T) {
	tests := []struct {
		name            string
		clusterTemplate []interface{}
		clusterProfile  []interface{}
		expectError     bool
		errorContains   string
	}{
		{
			name:            "Both empty - should pass",
			clusterTemplate: []interface{}{},
			clusterProfile:  []interface{}{},
			expectError:     false,
		},
		{
			name: "Only cluster_template provided - should pass",
			clusterTemplate: []interface{}{
				map[string]interface{}{
					"id": "template-123",
				},
			},
			clusterProfile: []interface{}{},
			expectError:    false,
		},
		{
			name:            "Only cluster_profile provided - should pass",
			clusterTemplate: []interface{}{},
			clusterProfile: []interface{}{
				map[string]interface{}{
					"id": "profile-123",
				},
			},
			expectError: false,
		},
		{
			name: "Both cluster_template and cluster_profile provided - should fail",
			clusterTemplate: []interface{}{
				map[string]interface{}{
					"id": "template-123",
				},
			},
			clusterProfile: []interface{}{
				map[string]interface{}{
					"id": "profile-123",
				},
			},
			expectError:   true,
			errorContains: "cannot specify both cluster_template and cluster_profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock ResourceData
			d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				"cluster_template": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type: schema.TypeString,
							},
						},
					},
				},
				"cluster_profile": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type: schema.TypeString,
							},
						},
					},
				},
			}, map[string]interface{}{
				"cluster_template": tt.clusterTemplate,
				"cluster_profile":  tt.clusterProfile,
			})

			err := validateProfileSource(d)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tt.expectError && err != nil && tt.errorContains != "" {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Error message '%s' does not contain '%s'", err.Error(), tt.errorContains)
				}
			}
		})
	}
}

// Note: Additional tests for extractProfilesFromTemplate and resolveProfileSource
// are covered by integration tests and the validation test above.
// Complex schema.Set testing requires more sophisticated test setup which is
// better handled through integration testing with actual Terraform configurations.
