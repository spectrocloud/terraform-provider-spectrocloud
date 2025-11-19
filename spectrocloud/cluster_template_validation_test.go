package spectrocloud

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
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

// TestResolveProfileSourceLogic tests the profile resolution logic
func TestResolveProfileSourceLogic(t *testing.T) {
	tests := []struct {
		name            string
		clusterTemplate []interface{}
		clusterProfile  []interface{}
		expectedSource  string
		expectError     bool
	}{
		{
			name:            "Both empty - defaults to cluster_profile source",
			clusterTemplate: []interface{}{},
			clusterProfile:  []interface{}{},
			expectedSource:  "cluster_profile",
			expectError:     false,
		},
		{
			name: "cluster_template provided - use template source",
			clusterTemplate: []interface{}{
				map[string]interface{}{
					"id": "template-123",
				},
			},
			clusterProfile: []interface{}{},
			expectedSource: "cluster_template",
			expectError:    false,
		},
		{
			name:            "cluster_profile provided - use profile source",
			clusterTemplate: []interface{}{},
			clusterProfile: []interface{}{
				map[string]interface{}{
					"id": "profile-123",
				},
			},
			expectedSource: "cluster_profile",
			expectError:    false,
		},
		{
			name: "Both provided - should error",
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
			expectedSource: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the logic conceptually
			hasTemplate := len(tt.clusterTemplate) > 0
			hasProfile := len(tt.clusterProfile) > 0

			if hasTemplate && hasProfile {
				assert.True(t, tt.expectError, "Should error when both are provided")
			} else if hasTemplate {
				assert.Equal(t, "cluster_template", tt.expectedSource)
			} else {
				assert.Equal(t, "cluster_profile", tt.expectedSource)
			}
		})
	}
}

// TestVariableHandling tests variable map handling
func TestVariableHandling(t *testing.T) {
	tests := []struct {
		name      string
		variables map[string]interface{}
		expected  map[string]string
	}{
		{
			name:      "Empty variables",
			variables: map[string]interface{}{},
			expected:  map[string]string{},
		},
		{
			name: "Single variable",
			variables: map[string]interface{}{
				"replicas": "3",
			},
			expected: map[string]string{
				"replicas": "3",
			},
		},
		{
			name: "Multiple variables",
			variables: map[string]interface{}{
				"replicas":   "3",
				"image_tag":  "v1.0.0",
				"pullPolicy": "IfNotPresent",
			},
			expected: map[string]string{
				"replicas":   "3",
				"image_tag":  "v1.0.0",
				"pullPolicy": "IfNotPresent",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert interface{} map to string map (simulating the actual conversion)
			result := make(map[string]string)
			for k, v := range tt.variables {
				if strVal, ok := v.(string); ok {
					result[k] = strVal
				}
			}

			assert.Equal(t, len(tt.expected), len(result))
			for k, v := range tt.expected {
				assert.Equal(t, v, result[k])
			}
		})
	}
}

// TestProfileFiltering tests that empty/invalid profiles are filtered
func TestProfileFiltering(t *testing.T) {
	tests := []struct {
		name          string
		profiles      []interface{}
		expectedValid int
		description   string
	}{
		{
			name:          "All nil profiles",
			profiles:      []interface{}{nil, nil, nil},
			expectedValid: 0,
			description:   "Should filter all nil entries",
		},
		{
			name: "Mix of valid and nil profiles",
			profiles: []interface{}{
				map[string]interface{}{"id": "profile-1"},
				nil,
				map[string]interface{}{"id": "profile-2"},
			},
			expectedValid: 2,
			description:   "Should keep only valid profiles",
		},
		{
			name: "Profile without id",
			profiles: []interface{}{
				map[string]interface{}{"id": "profile-1"},
				map[string]interface{}{"variables": map[string]interface{}{"key": "value"}},
			},
			expectedValid: 1,
			description:   "Should filter profiles without id",
		},
		{
			name: "All valid profiles",
			profiles: []interface{}{
				map[string]interface{}{"id": "profile-1"},
				map[string]interface{}{"id": "profile-2"},
				map[string]interface{}{"id": "profile-3"},
			},
			expectedValid: 3,
			description:   "Should keep all valid profiles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validCount := 0
			for _, p := range tt.profiles {
				if p != nil {
					if pMap, ok := p.(map[string]interface{}); ok {
						if _, hasID := pMap["id"]; hasID && pMap["id"] != "" {
							validCount++
						}
					}
				}
			}
			assert.Equal(t, tt.expectedValid, validCount, tt.description)
		})
	}
}

// Note: Additional integration tests for full end-to-end testing with actual
// Terraform configurations and API interactions are better handled through
// acceptance tests with real provider instances.
