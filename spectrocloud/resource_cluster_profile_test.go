package spectrocloud

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"github.com/stretchr/testify/assert"
)

func TestToClusterProfileVariables(t *testing.T) {
	mockResourceData := resourceClusterProfile().TestResourceData()
	var proVar []interface{}
	variables := map[string]interface{}{
		"variable": []interface{}{
			map[string]interface{}{
				"default_value": "default_value_1",
				"description":   "description_1",
				"display_name":  "display_name_1",
				"format":        "string",
				"hidden":        false,
				"immutable":     true,
				"name":          "variable_name_1",
				"regex":         "regex_1",
				"required":      true,
				"is_sensitive":  false,
			},
			map[string]interface{}{
				"default_value": "default_value_2",
				"description":   "description_2",
				"display_name":  "display_name_2",
				"format":        "integer",
				"hidden":        true,
				"immutable":     false,
				"name":          "variable_name_2",
				"regex":         "regex_2",
				"required":      false,
				"is_sensitive":  true,
			},
		},
	}
	proVar = append(proVar, variables)
	_ = mockResourceData.Set("cloud", "edge-native")
	_ = mockResourceData.Set("type", "add-on")
	_ = mockResourceData.Set("profile_variables", proVar)
	result, err := toClusterProfileVariables(mockResourceData)

	// Assertions for valid profile variables
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Test case 2: Empty profile variables
	mockResourceDataEmpty := resourceClusterProfile().TestResourceData()
	_ = mockResourceDataEmpty.Set("cloud", "edge-native")
	_ = mockResourceDataEmpty.Set("type", "add-on")
	_ = mockResourceDataEmpty.Set("profile_variables", []interface{}{map[string]interface{}{}})
	resultEmpty, errEmpty := toClusterProfileVariables(mockResourceDataEmpty)

	// Assertions for empty profile variables
	assert.NoError(t, errEmpty)
	assert.Len(t, resultEmpty, 0)

	// Test case 3: Invalid profile variables format
	mockResourceDataInvalid := resourceClusterProfile().TestResourceData()
	_ = mockResourceDataInvalid.Set("cloud", "edge-native")
	_ = mockResourceDataInvalid.Set("profile_variables", []interface{}{
		map[string]interface{}{
			"variable": []interface{}{}, // Invalid format, should be a list
		},
	})
	resultInvalid, _ := toClusterProfileVariables(mockResourceDataInvalid)

	// Assertions for invalid profile variables format
	assert.Len(t, resultInvalid, 0) // No variables should be extracted on error
}

func TestFlattenProfileVariables(t *testing.T) {
	// Test case 1: Valid profile variables and pv
	mockResourceData := resourceClusterProfile().TestResourceData()
	var proVar []interface{}
	variables := map[string]interface{}{
		"variable": []interface{}{
			map[string]interface{}{
				"name":          "variable_name_1",
				"display_name":  "display_name_1",
				"description":   "description_1",
				"format":        "string",
				"default_value": "default_value_1",
				"regex":         "regex_1",
				"required":      true,
				"immutable":     false,
				"hidden":        false,
			},
			map[string]interface{}{
				"name":          "variable_name_2",
				"display_name":  "display_name_2",
				"description":   "description_2",
				"format":        "integer",
				"default_value": "default_value_2",
				"regex":         "regex_2",
				"required":      false,
				"immutable":     true,
				"hidden":        true,
			},
		},
	}
	proVar = append(proVar, variables)
	_ = mockResourceData.Set("cloud", "edge-native")
	_ = mockResourceData.Set("profile_variables", proVar)

	pv := []*models.V1Variable{
		{Name: StringPtr("variable_name_1"), DisplayName: "display_name_1", Description: "description_1", Format: models.NewV1VariableFormat("string"), DefaultValue: "default_value_1", Regex: "regex_1", Required: true, Immutable: false, Hidden: false},
		{Name: StringPtr("variable_name_2"), DisplayName: "display_name_2", Description: "description_2", Format: models.NewV1VariableFormat("integer"), DefaultValue: "default_value_2", Regex: "regex_2", Required: false, Immutable: true, Hidden: true},
	}

	result, err := flattenProfileVariables(mockResourceData, pv)

	// Assertions for valid profile variables and pv
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, []interface{}{
		map[string]interface{}{
			"variable": []interface{}{
				map[string]interface{}{
					"name":          StringPtr("variable_name_1"),
					"display_name":  "display_name_1",
					"description":   "description_1",
					"format":        models.NewV1VariableFormat("string"),
					"default_value": "default_value_1",
					"regex":         "regex_1",
					"required":      true,
					"immutable":     false,
					"hidden":        false,
					"is_sensitive":  false,
				},
				map[string]interface{}{
					"name":          StringPtr("variable_name_2"),
					"display_name":  "display_name_2",
					"description":   "description_2",
					"format":        models.NewV1VariableFormat("integer"),
					"default_value": "default_value_2",
					"regex":         "regex_2",
					"required":      false,
					"immutable":     true,
					"hidden":        true,
					"is_sensitive":  false,
				},
			},
		},
	}, result)

	// Test case 2: Empty profile variables and pv
	//mockResourceDataEmpty := schema.TestResourceDataRaw(t, resourceClusterProfileVariables().Schema, map[string]interface{}{})
	mockResourceDataEmpty := resourceClusterProfile().TestResourceData()
	_ = mockResourceDataEmpty.Set("cloud", "edge-native")
	_ = mockResourceDataEmpty.Set("profile_variables", []interface{}{map[string]interface{}{}})
	resultEmpty, errEmpty := flattenProfileVariables(mockResourceDataEmpty, nil)

	// Assertions for empty profile variables and pv
	assert.NoError(t, errEmpty)
	assert.Len(t, resultEmpty, 0)
	assert.Equal(t, []interface{}{}, resultEmpty)
}

func TestToClusterProfileVariablesRestrictionError(t *testing.T) {
	mockResourceData := resourceClusterProfile().TestResourceData()
	var proVar []interface{}
	variables := map[string]interface{}{
		"variable": []interface{}{
			map[string]interface{}{
				"default_value": "default_value_1",
				"description":   "description_1",
				"display_name":  "display_name_1",
				"format":        "string",
				"hidden":        false,
				"immutable":     true,
				"name":          "variable_name_1",
				"regex":         "regex_1",
				"required":      true,
				"is_sensitive":  false,
			},
			map[string]interface{}{
				"default_value": "default_value_2",
				"description":   "description_2",
				"display_name":  "display_name_2",
				"format":        "integer",
				"hidden":        true,
				"immutable":     false,
				"name":          "variable_name_2",
				"regex":         "regex_2",
				"required":      false,
				"is_sensitive":  true,
			},
		},
	}
	proVar = append(proVar, variables)
	_ = mockResourceData.Set("cloud", "all")
	_ = mockResourceData.Set("type", "infra")
	_ = mockResourceData.Set("profile_variables", proVar)
	result, err := toClusterProfileVariables(mockResourceData)

	// Assertions for valid profile variables
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	_ = mockResourceData.Set("cloud", "edge-native")
	_ = mockResourceData.Set("type", "infra")
	result, err = toClusterProfileVariables(mockResourceData)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	_ = mockResourceData.Set("cloud", "aws")
	_ = mockResourceData.Set("type", "add-on")
	result, err = toClusterProfileVariables(mockResourceData)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	_ = mockResourceData.Set("cloud", "all")
	_ = mockResourceData.Set("type", "add-on")
	result, err = toClusterProfileVariables(mockResourceData)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	_ = mockResourceData.Set("cloud", "aws")
	_ = mockResourceData.Set("type", "infra")
	result, err = toClusterProfileVariables(mockResourceData)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	_ = mockResourceData.Set("cloud", "edge-native")
	_ = mockResourceData.Set("type", "add-on")
	result, err = toClusterProfileVariables(mockResourceData)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestToClusterProfilePackCreate(t *testing.T) {
	tests := []struct {
		name          string
		input         map[string]interface{}
		expectedError string
		expectedPack  *models.V1PackManifestEntity
	}{
		{
			name: "Valid Spectro Pack",
			input: map[string]interface{}{
				"name":         "test-pack",
				"type":         "spectro",
				"tag":          "v1.0",
				"uid":          "test-uid",
				"registry_uid": "test-registry-uid",
				"values":       "test-values",
				"manifest":     []interface{}{},
			},
			expectedError: "",
			expectedPack: &models.V1PackManifestEntity{
				Name:        types.Ptr("test-pack"),
				Tag:         "v1.0",
				RegistryUID: "test-registry-uid",
				UID:         "test-uid",
				Type:        models.V1PackTypeSpectro.Pointer(),
				Values:      "test-values",
				Manifests:   []*models.V1ManifestInputEntity{},
			},
		},
		{
			name: "Spectro Pack Missing UID and registry_uid",
			input: map[string]interface{}{
				"name":     "test-pack",
				"type":     "spectro",
				"tag":      "v1.0",
				"uid":      "",
				"values":   "test-values",
				"manifest": []interface{}{},
			},
			expectedError: "pack test-pack: either 'uid' must be provided, or all of the following fields must be specified for pack resolution: name, tag, registry_uid (or registry_name). Missing: registry_uid or registry_name",
			expectedPack:  nil,
		},
		{
			name: "Valid Manifest Pack with Default UID",
			input: map[string]interface{}{
				"name":   "test-manifest-pack",
				"type":   "manifest",
				"tag":    "",
				"uid":    "",
				"values": "test-values",
				"manifest": []interface{}{
					map[string]interface{}{
						"content": "manifest-content",
						"name":    "manifest-name",
					},
				},
			},
			expectedError: "",
			expectedPack: &models.V1PackManifestEntity{
				Name:        types.Ptr("test-manifest-pack"),
				Tag:         "",
				RegistryUID: "",
				UID:         "spectro-manifest-pack",
				Type:        models.V1PackTypeManifest.Pointer(),
				Values:      "test-values",
				Manifests: []*models.V1ManifestInputEntity{
					{
						Content: "manifest-content",
						Name:    "manifest-name",
					},
				},
			},
		},
		{
			name: "Valid Manifest Pack with Provided UID",
			input: map[string]interface{}{
				"name":   "test-manifest-pack",
				"type":   "manifest",
				"tag":    "",
				"uid":    "custom-uid",
				"values": "test-values",
				"manifest": []interface{}{
					map[string]interface{}{
						"content": "manifest-content",
						"name":    "manifest-name",
					},
				},
			},
			expectedError: "",
			expectedPack: &models.V1PackManifestEntity{
				Name:        types.Ptr("test-manifest-pack"),
				Tag:         "",
				RegistryUID: "",
				UID:         "custom-uid",
				Type:        models.V1PackTypeManifest.Pointer(),
				Values:      "test-values",
				Manifests: []*models.V1ManifestInputEntity{
					{
						Content: "manifest-content",
						Name:    "manifest-name",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function under test
			actualPack, err := toClusterProfilePackCreate(tt.input)

			// Check for errors
			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPack, actualPack)
			}
		})
	}
}

func prepareBaseClusterProfileTestData() *schema.ResourceData {
	d := resourceClusterProfile().TestResourceData()
	_ = d.Set("context", "project")
	_ = d.Set("name", "test-cluster-profile")
	_ = d.Set("version", "1.0.0")
	_ = d.Set("description", "test unit-test")
	_ = d.Set("cloud", "all")
	_ = d.Set("type", "cluster")
	var variables []interface{}
	variables = append(variables,
		map[string]interface{}{
			"variable": []interface{}{map[string]interface{}{
				"name":          "test_variable",
				"display_name":  "Test Vat",
				"format":        "string",
				"description":   "test var description",
				"default_value": "test",
				"regex":         "*",
				"required":      false,
				"immutable":     false,
				"is_sensitive":  false,
				"hidden":        false,
			},
			},
		},
	)
	_ = d.Set("profile_variables", variables)
	_ = d.Set("pack", []interface{}{
		map[string]interface{}{
			"uid":          "test-pack-uid-1",
			"type":         "spectro",
			"name":         "k8",
			"registry_uid": "test-pub-reg-uid",
			"tag":          "test:test",
			"values":       "test values",
			"manifest": []interface{}{map[string]interface{}{
				"uid":     "test-manifest-uid",
				"name":    "test-manifest",
				"content": "value content",
			},
			},
		},
		map[string]interface{}{
			"uid":          "test-pack-uid-2",
			"type":         "spectro",
			"name":         "csi",
			"registry_uid": "test-pub-reg-uid",
			"tag":          "test:test",
			"values":       "test values",
			"manifest": []interface{}{map[string]interface{}{
				"uid":     "test-manifest-uid",
				"name":    "test-manifest",
				"content": "value content",
			},
			},
		},
		map[string]interface{}{
			"uid":          "test-pack-uid-3",
			"type":         "spectro",
			"name":         "cni",
			"registry_uid": "test-pub-reg-uid",
			"tag":          "test:test",
			"values":       "test values",
			"manifest": []interface{}{map[string]interface{}{
				"uid":     "test-manifest-uid",
				"name":    "test-manifest",
				"content": "value content",
			},
			},
		},
		map[string]interface{}{
			"uid":          "test-pack-uid-4",
			"type":         "spectro",
			"name":         "os",
			"registry_uid": "test-pub-reg-uid",
			"tag":          "test:test",
			"values":       "test values",
			"manifest": []interface{}{map[string]interface{}{
				"uid":     "test-manifest-uid",
				"name":    "test-manifest",
				"content": "value content",
			},
			},
		},
	})
	d.SetId("cluster-profile-1")
	return d
}

func TestResourceClusterProfileCreate(t *testing.T) {
	d := prepareBaseClusterProfileTestData()
	var ctx context.Context
	_ = d.Set("type", "add-on")
	diags := resourceClusterProfileCreate(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Equal(t, "cluster-profile-1", d.Id())
}

func TestResourceClusterProfileCreateError(t *testing.T) {
	d := prepareBaseClusterProfileTestData()
	var ctx context.Context
	diags := resourceClusterProfileCreate(ctx, d, unitTestMockAPINegativeClient)
	assert.NotEmpty(t, diags)
}

func TestResourceClusterProfileRead(t *testing.T) {
	d := prepareBaseClusterProfileTestData()
	var ctx context.Context
	diags := resourceClusterProfileRead(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Equal(t, "cluster-profile-1", d.Id())
}

func TestResourceClusterProfileUpdate(t *testing.T) {
	d := prepareBaseClusterProfileTestData()
	var ctx context.Context
	diags := resourceClusterProfileUpdate(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Equal(t, "cluster-profile-1", d.Id())
}

func TestResourceClusterProfileDelete(t *testing.T) {
	d := prepareBaseClusterProfileTestData()
	var ctx context.Context
	diags := resourceClusterProfileDelete(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestValidatePackUIDOrResolutionFields(t *testing.T) {
	tests := []struct {
		name        string
		packData    map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid with uid provided",
			packData: map[string]interface{}{
				"name":         "test-pack",
				"tag":          "1.0.0",
				"registry_uid": "registry-123",
				"uid":          "pack-uid-123",
				"type":         "spectro",
			},
			expectError: false,
		},
		{
			name: "valid with all resolution fields provided",
			packData: map[string]interface{}{
				"name":         "test-pack",
				"tag":          "1.0.0",
				"registry_uid": "registry-123",
				"uid":          "",
				"type":         "spectro",
			},
			expectError: false,
		},
		{
			name: "manifest type should pass validation",
			packData: map[string]interface{}{
				"name":         "test-manifest",
				"tag":          "",
				"registry_uid": "",
				"uid":          "",
				"type":         "manifest",
			},
			expectError: false,
		},
		{
			name: "missing tag without uid should fail",
			packData: map[string]interface{}{
				"name":         "test-pack",
				"tag":          "",
				"registry_uid": "registry-123",
				"uid":          "",
				"type":         "spectro",
			},
			expectError: true,
			errorMsg:    "either 'uid' must be provided, or all of the following fields must be specified for pack resolution",
		},
		{
			name: "missing registry_uid without uid should fail",
			packData: map[string]interface{}{
				"name":         "test-pack",
				"tag":          "1.0.0",
				"registry_uid": "",
				"uid":          "",
				"type":         "spectro",
			},
			expectError: true,
			errorMsg:    "either 'uid' must be provided, or all of the following fields must be specified for pack resolution",
		},
		{
			name: "missing all resolution fields without uid should fail",
			packData: map[string]interface{}{
				"name":         "test-pack",
				"tag":          "",
				"registry_uid": "",
				"uid":          "",
				"type":         "spectro",
			},
			expectError: true,
			errorMsg:    "either 'uid' must be provided, or all of the following fields must be specified for pack resolution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := schemas.ValidatePackUIDOrResolutionFields(tt.packData)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestResolvePackUID(t *testing.T) {
	// This test would require mocking the client
	// For now, we'll test the validation logic of the function inputs
	c := &client.V1Client{} // Mock client - in real tests this would be properly mocked

	tests := []struct {
		name        string
		packName    string
		tag         string
		registryUID string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty name should fail",
			packName:    "",
			tag:         "1.0.0",
			registryUID: "registry-123",
			expectError: true,
			errorMsg:    "name, tag, and registry_uid are all required for pack resolution",
		},
		{
			name:        "empty tag should fail",
			packName:    "test-pack",
			tag:         "",
			registryUID: "registry-123",
			expectError: true,
			errorMsg:    "name, tag, and registry_uid are all required for pack resolution",
		},
		{
			name:        "empty registry_uid should fail",
			packName:    "test-pack",
			tag:         "1.0.0",
			registryUID: "",
			expectError: true,
			errorMsg:    "name, tag, and registry_uid are all required for pack resolution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resolvePackUID(c, tt.packName, tt.tag, tt.registryUID)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				// Note: These tests would require proper mocking to test successful resolution
				// For now, we just test the input validation
				if err != nil && strings.Contains(err.Error(), "name, tag, and registry_uid are all required") {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestFlattenClusterProfileCommon(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (*schema.ResourceData, *models.V1ClusterProfile)
		expectError bool
		description string
		verify      func(t *testing.T, d *schema.ResourceData, err error)
	}{
		{
			name: "Successful flattening with all fields",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile) {
				d := resourceClusterProfile().TestResourceData()
				cp := &models.V1ClusterProfile{
					Spec: &models.V1ClusterProfileSpec{
						Published: &models.V1ClusterProfileTemplate{
							CloudType:      "aws",
							Type:           "add-on",
							ProfileVersion: "1.0.0",
						},
					},
				}
				return d, cp
			},
			expectError: false,
			description: "Should successfully set cloud, type, and version fields",
			verify: func(t *testing.T, d *schema.ResourceData, err error) {
				assert.NoError(t, err, "Should not have error on success")
				assert.Equal(t, "aws", d.Get("cloud"), "Cloud should be set to 'aws'")
				assert.Equal(t, "add-on", d.Get("type"), "Type should be set to 'add-on'")
				assert.Equal(t, "1.0.0", d.Get("version"), "Version should be set to '1.0.0'")
			},
		},
		{
			name: "Flatten with different cloud types",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile) {
				d := resourceClusterProfile().TestResourceData()
				cp := &models.V1ClusterProfile{
					Spec: &models.V1ClusterProfileSpec{
						Published: &models.V1ClusterProfileTemplate{
							CloudType:      "edge-native",
							Type:           "cluster",
							ProfileVersion: "2.5.3",
						},
					},
				}
				return d, cp
			},
			expectError: false,
			description: "Should handle different cloud types",
			verify: func(t *testing.T, d *schema.ResourceData, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.Equal(t, "edge-native", d.Get("cloud"), "Cloud should be set to 'edge-native'")
				assert.Equal(t, "cluster", d.Get("type"), "Type should be set to 'cluster'")
				assert.Equal(t, "2.5.3", d.Get("version"), "Version should be set to '2.5.3'")
			},
		},
		{
			name: "Flatten with different profile types",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile) {
				d := resourceClusterProfile().TestResourceData()
				cp := &models.V1ClusterProfile{
					Spec: &models.V1ClusterProfileSpec{
						Published: &models.V1ClusterProfileTemplate{
							CloudType:      "azure",
							Type:           "infra",
							ProfileVersion: "3.1.0",
						},
					},
				}
				return d, cp
			},
			expectError: false,
			description: "Should handle different profile types",
			verify: func(t *testing.T, d *schema.ResourceData, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.Equal(t, "azure", d.Get("cloud"), "Cloud should be set to 'azure'")
				assert.Equal(t, "infra", d.Get("type"), "Type should be set to 'infra'")
				assert.Equal(t, "3.1.0", d.Get("version"), "Version should be set to '3.1.0'")
			},
		},
		{
			name: "Flatten with system profile type",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile) {
				d := resourceClusterProfile().TestResourceData()
				cp := &models.V1ClusterProfile{
					Spec: &models.V1ClusterProfileSpec{
						Published: &models.V1ClusterProfileTemplate{
							CloudType:      "all",
							Type:           "system",
							ProfileVersion: "1.2.3",
						},
					},
				}
				return d, cp
			},
			expectError: false,
			description: "Should handle system profile type",
			verify: func(t *testing.T, d *schema.ResourceData, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.Equal(t, "all", d.Get("cloud"), "Cloud should be set to 'all'")
				assert.Equal(t, "system", d.Get("type"), "Type should be set to 'system'")
				assert.Equal(t, "1.2.3", d.Get("version"), "Version should be set to '1.2.3'")
			},
		},
		{
			name: "Flatten with nil Spec (should panic)",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile) {
				d := resourceClusterProfile().TestResourceData()
				cp := &models.V1ClusterProfile{
					Spec: nil,
				}
				return d, cp
			},
			expectError: true,
			description: "Should panic when Spec is nil",
			verify: func(t *testing.T, d *schema.ResourceData, err error) {
				// Function will panic on nil pointer dereference
				// This test verifies the function doesn't handle nil gracefully
			},
		},
		{
			name: "Flatten with nil Published (should panic)",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile) {
				d := resourceClusterProfile().TestResourceData()
				cp := &models.V1ClusterProfile{
					Spec: &models.V1ClusterProfileSpec{
						Published: nil,
					},
				}
				return d, cp
			},
			expectError: true,
			description: "Should panic when Published is nil",
			verify: func(t *testing.T, d *schema.ResourceData, err error) {
				// Function will panic on nil pointer dereference
				// This test verifies the function doesn't handle nil gracefully
			},
		},
		{
			name: "Flatten with empty Type string",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile) {
				d := resourceClusterProfile().TestResourceData()
				cp := &models.V1ClusterProfile{
					Spec: &models.V1ClusterProfileSpec{
						Published: &models.V1ClusterProfileTemplate{
							CloudType:      "gcp",
							Type:           "",
							ProfileVersion: "4.0.0",
						},
					},
				}
				return d, cp
			},
			expectError: false,
			description: "Should handle empty Type string",
			verify: func(t *testing.T, d *schema.ResourceData, err error) {
				assert.NoError(t, err, "Should not have error with empty type string")
				assert.Equal(t, "gcp", d.Get("cloud"), "Cloud should be set correctly")
				assert.Equal(t, "", d.Get("type"), "Type should be set to empty string")
				assert.Equal(t, "4.0.0", d.Get("version"), "Version should be set correctly")
			},
		},
		{
			name: "Flatten with custom cloud type",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile) {
				d := resourceClusterProfile().TestResourceData()
				cp := &models.V1ClusterProfile{
					Spec: &models.V1ClusterProfileSpec{
						Published: &models.V1ClusterProfileTemplate{
							CloudType:      "nutanix",
							Type:           "add-on",
							ProfileVersion: "1.0.0",
						},
					},
				}
				return d, cp
			},
			expectError: false,
			description: "Should handle custom cloud types",
			verify: func(t *testing.T, d *schema.ResourceData, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.Equal(t, "nutanix", d.Get("cloud"), "Cloud should be set to custom cloud type 'nutanix'")
				assert.Equal(t, "add-on", d.Get("type"), "Type should be set correctly")
				assert.Equal(t, "1.0.0", d.Get("version"), "Version should be set correctly")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, cp := tt.setup()

			var err error
			var panicked bool

			// Handle potential panics for nil pointer dereferences
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
						err = fmt.Errorf("panic: %v", r)
					}
				}()
				err = flattenClusterProfileCommon(d, cp)
			}()

			// Verify results
			if tt.expectError {
				if panicked {
					// Panic is expected for nil pointer cases
					assert.Error(t, err, "Expected panic/error for test case: %s", tt.description)
				} else {
					assert.Error(t, err, "Expected error for test case: %s", tt.description)
				}
			} else {
				if panicked {
					t.Logf("Unexpected panic occurred: %v", err)
					assert.Fail(t, "Unexpected panic for test case: %s", tt.description)
				} else {
					assert.NoError(t, err, "Should not have error for test case: %s", tt.description)
				}
			}

			// Run custom verify function if provided
			if tt.verify != nil {
				tt.verify(t, d, err)
			}
		})
	}
}

func TestToClusterProfileCreateWithResolution(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (*schema.ResourceData, *client.V1Client)
		expectError bool
		description string
		verify      func(t *testing.T, cp *models.V1ClusterProfileEntity, err error)
	}{
		{
			name: "Successful creation with packs and variables",
			setup: func() (*schema.ResourceData, *client.V1Client) {
				d := resourceClusterProfile().TestResourceData()
				_ = d.Set("name", "test-profile")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("description", "test description")
				_ = d.Set("cloud", "aws")
				_ = d.Set("type", "add-on")
				_ = d.Set("pack", []interface{}{
					map[string]interface{}{
						"uid":          "test-pack-uid-1",
						"type":         "spectro",
						"name":         "test-pack",
						"registry_uid": "test-registry-uid",
						"tag":          "v1.0.0",
						"values":       "test values",
						"manifest":     []interface{}{},
					},
				})
				_ = d.Set("profile_variables", []interface{}{
					map[string]interface{}{
						"variable": []interface{}{
							map[string]interface{}{
								"name":          "test_var",
								"display_name":  "Test Variable",
								"format":        "string",
								"description":   "test description",
								"default_value": "default",
								"regex":         "",
								"required":      false,
								"immutable":     false,
								"is_sensitive":  false,
								"hidden":        false,
							},
						},
					},
				})
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return d, c
			},
			expectError: false,
			description: "Should successfully create cluster profile with packs and variables",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "test-profile", cp.Metadata.Name, "Name should match")
				assert.Equal(t, "1.0.0", cp.Spec.Version, "Version should match")
				assert.Equal(t, "aws", cp.Spec.Template.CloudType, "Cloud type should match")
				assert.NotNil(t, cp.Spec.Template.Packs, "Packs should not be nil")
				assert.NotNil(t, cp.Spec.Variables, "Variables should not be nil")
			},
		},
		{
			name: "Successful creation with multiple packs",
			setup: func() (*schema.ResourceData, *client.V1Client) {
				d := resourceClusterProfile().TestResourceData()
				_ = d.Set("name", "test-profile")
				_ = d.Set("version", "3.0.0")
				_ = d.Set("cloud", "edge-native")
				_ = d.Set("type", "add-on")
				_ = d.Set("pack", []interface{}{
					map[string]interface{}{
						"uid":          "pack-uid-1",
						"type":         "spectro",
						"name":         "pack1",
						"registry_uid": "reg-uid",
						"tag":          "v1.0",
						"values":       "values1",
						"manifest":     []interface{}{},
					},
					map[string]interface{}{
						"uid":          "pack-uid-2",
						"type":         "spectro",
						"name":         "pack2",
						"registry_uid": "reg-uid",
						"tag":          "v2.0",
						"values":       "values2",
						"manifest":     []interface{}{},
					},
					map[string]interface{}{
						"uid":          "",
						"type":         "manifest",
						"name":         "manifest-pack",
						"registry_uid": "",
						"tag":          "",
						"values":       "",
						"manifest":     []interface{}{},
					},
				})
				_ = d.Set("profile_variables", []interface{}{
					map[string]interface{}{
						"variable": []interface{}{},
					},
				})
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return d, c
			},
			expectError: false,
			description: "Should successfully create cluster profile with multiple packs",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "test-profile", cp.Metadata.Name, "Name should match")
				assert.NotNil(t, cp.Spec.Template.Packs, "Packs should not be nil")
				assert.GreaterOrEqual(t, len(cp.Spec.Template.Packs), 1, "Should have at least one pack")
			},
		},
		{
			name: "Successful creation with manifest pack type",
			setup: func() (*schema.ResourceData, *client.V1Client) {
				d := resourceClusterProfile().TestResourceData()
				_ = d.Set("name", "test-profile")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("cloud", "vsphere")
				_ = d.Set("type", "cluster")
				_ = d.Set("pack", []interface{}{
					map[string]interface{}{
						"uid":          "",
						"type":         "manifest",
						"name":         "manifest-pack",
						"registry_uid": "",
						"tag":          "",
						"values":       "manifest values",
						"manifest": []interface{}{
							map[string]interface{}{
								"name":    "manifest1",
								"content": "manifest content",
							},
						},
					},
				})
				_ = d.Set("profile_variables", []interface{}{
					map[string]interface{}{
						"variable": []interface{}{},
					},
				})
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return d, c
			},
			expectError: false,
			description: "Should successfully create cluster profile with manifest pack type",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "test-profile", cp.Metadata.Name, "Name should match")
				assert.NotNil(t, cp.Spec.Template.Packs, "Packs should not be nil")
				if len(cp.Spec.Template.Packs) > 0 {
					assert.Equal(t, "spectro-manifest-pack", cp.Spec.Template.Packs[0].UID, "Manifest pack should have default UID")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, c := tt.setup()

			var cp *models.V1ClusterProfileEntity
			var err error
			var panicked bool

			// Handle potential panics
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
						err = fmt.Errorf("panic: %v", r)
					}
				}()
				cp, err = toClusterProfileCreateWithResolution(d, c)
			}()

			// Verify results
			if tt.expectError {
				if panicked {
					assert.Error(t, err, "Expected panic/error for test case: %s", tt.description)
				} else {
					assert.Error(t, err, "Expected error for test case: %s", tt.description)
				}
			} else {
				if panicked {
					t.Logf("Unexpected panic occurred: %v", err)
					assert.Fail(t, "Unexpected panic for test case: %s", tt.description)
				} else {
					assert.NoError(t, err, "Should not have error for test case: %s", tt.description)
				}
			}

			// Run custom verify function if provided
			if tt.verify != nil {
				tt.verify(t, cp, err)
			}
		})
	}
}

func TestToClusterProfileBasic(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *schema.ResourceData
		expectError bool
		description string
		verify      func(t *testing.T, cp *models.V1ClusterProfileEntity)
	}{
		{
			name: "Successful creation with all fields",
			setup: func() *schema.ResourceData {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("test-profile-uid")
				_ = d.Set("name", "test-profile")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("description", "Test description")
				_ = d.Set("cloud", "aws")
				_ = d.Set("type", "add-on")
				_ = d.Set("tags", []interface{}{"tag1:value1", "tag2:value2"})
				return d
			},
			expectError: false,
			description: "Should successfully create basic cluster profile with all fields",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity) {
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "test-profile", cp.Metadata.Name, "Name should match")
				assert.Equal(t, "test-profile-uid", cp.Metadata.UID, "UID should match")
				assert.Equal(t, "Test description", cp.Metadata.Annotations["description"], "Description should match")
				assert.Equal(t, "aws", cp.Spec.Template.CloudType, "Cloud type should match")
				assert.Equal(t, "add-on", string(*cp.Spec.Template.Type), "Type should match")
				assert.Equal(t, "1.0.0", cp.Spec.Version, "Version should match")
				assert.NotNil(t, cp.Metadata.Labels, "Labels should not be nil")
			},
		},
		{
			name: "Successful creation with tags",
			setup: func() *schema.ResourceData {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("test-profile-uid-4")
				_ = d.Set("name", "test-profile-4")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("cloud", "edge-native")
				_ = d.Set("type", "add-on")
				_ = d.Set("tags", []interface{}{"env:prod", "team:devops"})
				return d
			},
			expectError: false,
			description: "Should successfully create basic cluster profile with tags",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity) {
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "test-profile-4", cp.Metadata.Name, "Name should match")
				assert.NotNil(t, cp.Metadata.Labels, "Labels should not be nil")
				// Verify tags are converted to labels
				if cp.Metadata.Labels != nil {
					assert.Greater(t, len(cp.Metadata.Labels), 0, "Labels should contain tags")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setup()

			var cp *models.V1ClusterProfileEntity
			var panicked bool
			var err error

			// Handle potential panics for missing required fields
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
						err = fmt.Errorf("panic: %v", r)
					}
				}()
				cp = toClusterProfileBasic(d)
			}()

			// Verify results
			if tt.expectError {
				if panicked {
					// Panic is expected for missing required fields
					assert.Error(t, err, "Expected panic/error for test case: %s", tt.description)
				} else {
					assert.Fail(t, "Expected panic/error but got none for test case: %s", tt.description)
				}
			} else {
				if panicked {
					t.Logf("Unexpected panic occurred: %v", err)
					assert.Fail(t, "Unexpected panic for test case: %s", tt.description)
				} else {
					assert.NotNil(t, cp, "Cluster profile should not be nil for test case: %s", tt.description)
				}
			}

			// Run custom verify function if provided
			if tt.verify != nil && !panicked {
				tt.verify(t, cp)
			}
		})
	}
}

func TestToClusterProfileUpdateWithResolution(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (*schema.ResourceData, *models.V1ClusterProfile, *client.V1Client)
		expectError bool
		description string
		verify      func(t *testing.T, cp *models.V1ClusterProfileUpdateEntity, err error)
	}{
		{
			name: "Successful update with multiple packs",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile, *client.V1Client) {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("test-profile-uid-3")
				_ = d.Set("name", "test-profile-3")
				_ = d.Set("version", "4.0.0")
				_ = d.Set("type", "infra")
				_ = d.Set("pack", []interface{}{
					map[string]interface{}{
						"uid":          "pack-uid-1",
						"type":         "spectro",
						"name":         "pack1",
						"registry_uid": "reg-uid",
						"tag":          "v1.0",
						"values":       "values1",
						"manifest":     []interface{}{},
					},
					map[string]interface{}{
						"uid":          "pack-uid-2",
						"type":         "spectro",
						"name":         "pack2",
						"registry_uid": "reg-uid",
						"tag":          "v2.0",
						"values":       "values2",
						"manifest":     []interface{}{},
					},
					map[string]interface{}{
						"uid":          "",
						"type":         "manifest",
						"name":         "manifest-pack",
						"registry_uid": "",
						"tag":          "",
						"values":       "",
						"manifest":     []interface{}{},
					},
				})
				cluster := &models.V1ClusterProfile{
					Spec: &models.V1ClusterProfileSpec{
						Published: &models.V1ClusterProfileTemplate{
							Packs: []*models.V1PackRef{
								{
									PackUID: "pack-uid-1",
									Name:    types.Ptr("pack1"),
								},
								{
									PackUID: "pack-uid-2",
									Name:    types.Ptr("pack2"),
								},
							},
						},
					},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return d, cluster, c
			},
			expectError: false,
			description: "Should successfully create update entity with multiple packs",
			verify: func(t *testing.T, cp *models.V1ClusterProfileUpdateEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, cp, "Cluster profile update entity should not be nil")
				assert.Equal(t, "test-profile-3", cp.Metadata.Name, "Name should match")
				assert.NotNil(t, cp.Spec.Template.Packs, "Packs should not be nil")
				assert.GreaterOrEqual(t, len(cp.Spec.Template.Packs), 1, "Should have at least one pack")
			},
		},
		{
			name: "Error from pack resolution - missing registry_uid",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile, *client.V1Client) {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("test-profile-uid-4")
				_ = d.Set("name", "test-profile-4")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("type", "add-on")
				_ = d.Set("pack", []interface{}{
					map[string]interface{}{
						"uid":          "",
						"type":         "spectro",
						"name":         "test-pack",
						"registry_uid": "",
						"tag":          "v1.0.0",
						"values":       "test values",
						"manifest":     []interface{}{},
					},
				})
				cluster := &models.V1ClusterProfile{
					Spec: &models.V1ClusterProfileSpec{
						Published: &models.V1ClusterProfileTemplate{
							Packs: []*models.V1PackRef{},
						},
					},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return d, cluster, c
			},
			expectError: true,
			description: "Should return error when pack resolution fails due to missing registry_uid",
			verify: func(t *testing.T, cp *models.V1ClusterProfileUpdateEntity, err error) {
				assert.Error(t, err, "Should have error")
				assert.Nil(t, cp, "Cluster profile update entity should be nil on error")
				assert.Contains(t, err.Error(), "either 'uid' must be provided", "Error should mention missing fields")
			},
		},
		{
			name: "Successful update with different profile types",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile, *client.V1Client) {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("test-profile-uid-6")
				_ = d.Set("name", "test-profile-6")
				_ = d.Set("version", "5.0.0")
				_ = d.Set("type", "cluster")
				_ = d.Set("pack", []interface{}{})
				cluster := &models.V1ClusterProfile{
					Spec: &models.V1ClusterProfileSpec{
						Published: &models.V1ClusterProfileTemplate{
							Packs: []*models.V1PackRef{},
						},
					},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return d, cluster, c
			},
			expectError: false,
			description: "Should successfully create update entity with cluster type",
			verify: func(t *testing.T, cp *models.V1ClusterProfileUpdateEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, cp, "Cluster profile update entity should not be nil")
				assert.Equal(t, "cluster", string(*cp.Spec.Template.Type), "Type should be 'cluster'")
				assert.Equal(t, "5.0.0", cp.Spec.Version, "Version should match")
			},
		},
		{
			name: "Panic when cluster.Spec.Published is nil",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile, *client.V1Client) {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("test-profile-uid-9")
				_ = d.Set("name", "test-profile-9")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("type", "add-on")
				_ = d.Set("pack", []interface{}{
					map[string]interface{}{
						"uid":          "test-pack-uid",
						"type":         "spectro",
						"name":         "test-pack",
						"registry_uid": "reg-uid",
						"tag":          "v1.0",
						"values":       "",
						"manifest":     []interface{}{},
					},
				})
				cluster := &models.V1ClusterProfile{
					Spec: &models.V1ClusterProfileSpec{
						Published: nil,
					},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return d, cluster, c
			},
			expectError: true,
			description: "Should panic when cluster.Spec.Published is nil",
			verify: func(t *testing.T, cp *models.V1ClusterProfileUpdateEntity, err error) {
				// Function will panic on nil pointer dereference
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, cluster, c := tt.setup()

			var cp *models.V1ClusterProfileUpdateEntity
			var err error
			var panicked bool

			// Handle potential panics
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
						err = fmt.Errorf("panic: %v", r)
					}
				}()
				cp, err = toClusterProfileUpdateWithResolution(d, cluster, c)
			}()

			// Verify results
			if tt.expectError {
				if panicked {
					// Panic is expected for nil pointer cases
					assert.Error(t, err, "Expected panic/error for test case: %s", tt.description)
				} else {
					assert.Error(t, err, "Expected error for test case: %s", tt.description)
				}
			} else {
				if panicked {
					t.Logf("Unexpected panic occurred: %v", err)
					assert.Fail(t, "Unexpected panic for test case: %s", tt.description)
				} else {
					assert.NoError(t, err, "Should not have error for test case: %s", tt.description)
				}
			}

			// Run custom verify function if provided
			if tt.verify != nil && !panicked {
				tt.verify(t, cp, err)
			}
		})
	}
}

func TestToClusterProfilePackCreateWithResolution(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (map[string]interface{}, *client.V1Client)
		expectError bool
		description string
		verify      func(t *testing.T, pack *models.V1PackManifestEntity, err error)
	}{
		{
			name: "Successful creation with manifest pack and default UID",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":   "test-manifest-pack",
					"type":   "manifest",
					"tag":    "",
					"uid":    "",
					"values": "test-values",
					"manifest": []interface{}{
						map[string]interface{}{
							"content": "manifest-content",
							"name":    "manifest-name",
						},
					},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return input, c
			},
			expectError: false,
			description: "Should successfully create manifest pack with default UID",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, pack, "Pack should not be nil")
				assert.Equal(t, "test-manifest-pack", *pack.Name, "Name should match")
				assert.Equal(t, "spectro-manifest-pack", pack.UID, "UID should be default for manifest pack")
				assert.Equal(t, models.V1PackTypeManifest, *pack.Type, "Type should be Manifest")
				assert.Equal(t, 1, len(pack.Manifests), "Should have one manifest")
			},
		},
		{
			name: "Successful creation with values and manifest content trimming",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":         "test-pack",
					"type":         "spectro",
					"tag":          "v1.0",
					"uid":          "test-uid",
					"registry_uid": "test-registry-uid",
					"values":       "test-values\n",
					"manifest": []interface{}{
						map[string]interface{}{
							"content": "manifest-content\n",
							"name":    "manifest-name",
						},
					},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return input, c
			},
			expectError: false,
			description: "Should successfully trim whitespace from values and manifest content",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, pack, "Pack should not be nil")
				assert.Equal(t, "test-values", pack.Values, "Values should be trimmed")
				assert.Equal(t, "manifest-content", pack.Manifests[0].Content, "Manifest content should be trimmed")
			},
		},
		{
			name: "Successful creation with empty values",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":         "test-pack",
					"type":         "spectro",
					"tag":          "v1.0",
					"uid":          "test-uid",
					"registry_uid": "test-registry-uid",
					"values":       "",
					"manifest":     []interface{}{},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return input, c
			},
			expectError: false,
			description: "Should successfully create pack with empty values",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, pack, "Pack should not be nil")
				assert.Equal(t, "", pack.Values, "Values should be empty string")
			},
		},
		{
			name: "Successful resolution: registry_name resolved even with UID provided",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":          "test-pack",
					"type":          "spectro",
					"tag":           "v1.0",
					"uid":           "test-uid",
					"registry_name": "test-registry-name",
					"values":        "test-values",
					"manifest":      []interface{}{},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return input, c
			},
			expectError: false,
			description: "Function resolves registry_name even if UID is provided (tests resolveRegistryNameToUID is called regardless of UID presence)",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				if !assert.NoError(t, err, "Should not have error") {
					return
				}
				if !assert.NotNil(t, pack, "Pack should not be nil") {
					return
				}
				assert.Equal(t, "test-registry-uid", pack.RegistryUID, "RegistryUID should be resolved from registry_name even when UID is provided")
				assert.Equal(t, "test-uid", pack.UID, "UID should remain as provided")
			},
		},
		{
			name: "Successful creation with Spectro pack type",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":         "test-pack",
					"type":         "spectro",
					"tag":          "v1.0",
					"uid":          "test-uid",
					"registry_uid": "test-registry-uid",
					"values":       "test-values",
					"manifest":     []interface{}{},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return input, c
			},
			expectError: false,
			description: "Should successfully create Spectro pack",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, pack, "Pack should not be nil")
				assert.Equal(t, models.V1PackTypeSpectro, *pack.Type, "Type should be Spectro")
			},
		},
		{
			name: "Panic when client is nil",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":         "test-pack",
					"type":         "spectro",
					"tag":          "v1.0",
					"uid":          "",
					"registry_uid": "test-registry-uid",
					"values":       "test-values",
					"manifest":     []interface{}{},
				}
				return input, nil
			},
			expectError: true,
			description: "Should panic when client is nil and pack UID needs resolution",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				// Function will panic on nil pointer dereference when trying to resolve pack UID
			},
		},
		{
			name: "Successful creation with manifest pack and custom UID",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":   "test-manifest-pack",
					"type":   "manifest",
					"tag":    "",
					"uid":    "custom-manifest-uid",
					"values": "test-values",
					"manifest": []interface{}{
						map[string]interface{}{
							"content": "manifest-content",
							"name":    "manifest-name",
						},
					},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return input, c
			},
			expectError: false,
			description: "Should successfully create manifest pack with custom UID",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, pack, "Pack should not be nil")
				assert.Equal(t, "custom-manifest-uid", pack.UID, "UID should be custom value")
				assert.Equal(t, models.V1PackTypeManifest, *pack.Type, "Type should be Manifest")
			},
		},
		{
			name: "Successful resolution: registry_name to registry_uid for Helm pack",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":          "k8",
					"type":          "helm",
					"tag":           "1.0",
					"uid":           "",
					"registry_name": "Public",
					"values":        "test-values",
					"manifest":      []interface{}{},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return input, c
			},
			expectError: false,
			description: "Should successfully resolve registry_name to registry_uid using GetHelmRegistryByName, then resolve pack UID",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, pack, "Pack should not be nil")
				assert.Equal(t, "k8", *pack.Name, "Name should match")
				assert.Equal(t, "test-registry-uid", pack.RegistryUID, "RegistryUID should be resolved from registry_name")
				assert.Equal(t, "test-pack-uid", pack.UID, "Pack UID should be resolved")
				assert.Equal(t, "1.0", pack.Tag, "Tag should match")
				assert.Equal(t, models.V1PackTypeHelm, *pack.Type, "Type should be Helm")
			},
		},
		{
			name: "Successful resolution: pack UID resolution with provided registry_uid",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":         "k8",
					"type":         "spectro",
					"tag":          "1.0",
					"uid":          "",
					"registry_uid": "test-registry-uid",
					"values":       "test-values",
					"manifest":     []interface{}{},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return input, c
			},
			expectError: false,
			description: "Should successfully resolve pack UID when registry_uid is directly provided (tests resolvePackUID function)",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, pack, "Pack should not be nil")
				assert.Equal(t, "k8", *pack.Name, "Name should match")
				assert.Equal(t, "test-registry-uid", pack.RegistryUID, "RegistryUID should match provided value")
				assert.Equal(t, "test-pack-uid", pack.UID, "Pack UID should be resolved via GetPacksByNameAndRegistry")
				assert.Equal(t, "1.0", pack.Tag, "Tag should match")
				assert.Equal(t, models.V1PackTypeSpectro, *pack.Type, "Type should be Spectro")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, c := tt.setup()

			var pack *models.V1PackManifestEntity
			var err error
			var panicked bool

			// Handle potential panics
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
						err = fmt.Errorf("panic: %v", r)
					}
				}()
				pack, err = toClusterProfilePackCreateWithResolution(input, c)
			}()

			// Verify results
			if tt.expectError {
				if panicked {
					// Panic is expected for nil client cases
					assert.Error(t, err, "Expected panic/error for test case: %s", tt.description)
				} else {
					assert.Error(t, err, "Expected error for test case: %s", tt.description)
				}
			} else {
				if panicked {
					t.Logf("Unexpected panic occurred: %v", err)
					assert.Fail(t, "Unexpected panic for test case: %s", tt.description)
				} else {
					assert.NoError(t, err, "Should not have error for test case: %s", tt.description)
				}
			}

			// Run custom verify function if provided
			if tt.verify != nil && !panicked {
				tt.verify(t, pack, err)
			}
		})
	}
}
