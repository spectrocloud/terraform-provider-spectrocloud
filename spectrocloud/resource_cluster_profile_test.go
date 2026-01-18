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

// TestFlattenClusterProfileCommon tests the flattenClusterProfileCommon function.
// This function:
// 1. Sets the "cloud" field from cp.Spec.Published.CloudType
// 2. Sets the "type" field from cp.Spec.Published.Type
// 3. Sets the "version" field from cp.Spec.Published.ProfileVersion
// 4. Returns an error if any Set operation fails
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
			name: "Flatten with empty string values",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile) {
				d := resourceClusterProfile().TestResourceData()
				cp := &models.V1ClusterProfile{
					Spec: &models.V1ClusterProfileSpec{
						Published: &models.V1ClusterProfileTemplate{
							CloudType:      "",
							Type:           "add-on",
							ProfileVersion: "",
						},
					},
				}
				return d, cp
			},
			expectError: false,
			description: "Should handle empty string values",
			verify: func(t *testing.T, d *schema.ResourceData, err error) {
				assert.NoError(t, err, "Should not have error with empty strings")
				assert.Equal(t, "", d.Get("cloud"), "Cloud should be set to empty string")
				assert.Equal(t, "add-on", d.Get("type"), "Type should still be set")
				assert.Equal(t, "", d.Get("version"), "Version should be set to empty string")
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
		{
			name: "Flatten with long version string",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile) {
				d := resourceClusterProfile().TestResourceData()
				cp := &models.V1ClusterProfile{
					Spec: &models.V1ClusterProfileSpec{
						Published: &models.V1ClusterProfileTemplate{
							CloudType:      "vsphere",
							Type:           "cluster",
							ProfileVersion: "10.20.30-beta.1+sha.abc123",
						},
					},
				}
				return d, cp
			},
			expectError: false,
			description: "Should handle long version strings with metadata",
			verify: func(t *testing.T, d *schema.ResourceData, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.Equal(t, "vsphere", d.Get("cloud"), "Cloud should be set correctly")
				assert.Equal(t, "cluster", d.Get("type"), "Type should be set correctly")
				assert.Equal(t, "10.20.30-beta.1+sha.abc123", d.Get("version"), "Version should preserve full version string")
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
