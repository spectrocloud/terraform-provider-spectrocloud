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

// TestToClusterProfilePackCreate tests the toClusterProfilePackCreate function.
// This function:
// 1. Extracts pack fields from input map
// 2. Validates pack UID or resolution fields
// 3. Handles different pack types (Spectro, Manifest)
// 4. Trims whitespace from values and manifest content
// 5. Returns V1PackManifestEntity
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
		{
			name: "Valid Spectro Pack with registry_name",
			input: map[string]interface{}{
				"name":          "test-pack",
				"type":          "spectro",
				"tag":           "v1.0",
				"uid":           "test-uid",
				"registry_name": "test-registry-name",
				"values":        "test-values",
				"manifest":      []interface{}{},
			},
			expectedError: "",
			expectedPack: &models.V1PackManifestEntity{
				Name:        types.Ptr("test-pack"),
				Tag:         "v1.0",
				RegistryUID: "", // registry_name is stored but not resolved here
				UID:         "test-uid",
				Type:        models.V1PackTypeSpectro.Pointer(),
				Values:      "test-values",
				Manifests:   []*models.V1ManifestInputEntity{},
			},
		},
		{
			name: "Valid Pack with multiple manifests",
			input: map[string]interface{}{
				"name":         "test-pack",
				"type":         "spectro",
				"tag":          "v1.0",
				"uid":          "test-uid",
				"registry_uid": "test-registry-uid",
				"values":       "test-values",
				"manifest": []interface{}{
					map[string]interface{}{
						"content": "manifest-content-1",
						"name":    "manifest-name-1",
					},
					map[string]interface{}{
						"content": "manifest-content-2",
						"name":    "manifest-name-2",
					},
				},
			},
			expectedError: "",
			expectedPack: &models.V1PackManifestEntity{
				Name:        types.Ptr("test-pack"),
				Tag:         "v1.0",
				RegistryUID: "test-registry-uid",
				UID:         "test-uid",
				Type:        models.V1PackTypeSpectro.Pointer(),
				Values:      "test-values",
				Manifests: []*models.V1ManifestInputEntity{
					{
						Content: "manifest-content-1",
						Name:    "manifest-name-1",
					},
					{
						Content: "manifest-content-2",
						Name:    "manifest-name-2",
					},
				},
			},
		},
		{
			name: "Valid Pack with values containing newlines (should be trimmed)",
			input: map[string]interface{}{
				"name":         "test-pack",
				"type":         "spectro",
				"tag":          "v1.0",
				"uid":          "test-uid",
				"registry_uid": "test-registry-uid",
				"values":       "test-values\n",
				"manifest":     []interface{}{},
			},
			expectedError: "",
			expectedPack: &models.V1PackManifestEntity{
				Name:        types.Ptr("test-pack"),
				Tag:         "v1.0",
				RegistryUID: "test-registry-uid",
				UID:         "test-uid",
				Type:        models.V1PackTypeSpectro.Pointer(),
				Values:      "test-values", // Should be trimmed
				Manifests:   []*models.V1ManifestInputEntity{},
			},
		},
		{
			name: "Valid Pack with manifest content containing newlines (should be trimmed)",
			input: map[string]interface{}{
				"name":         "test-pack",
				"type":         "spectro",
				"tag":          "v1.0",
				"uid":          "test-uid",
				"registry_uid": "test-registry-uid",
				"values":       "test-values",
				"manifest": []interface{}{
					map[string]interface{}{
						"content": "manifest-content\n",
						"name":    "manifest-name",
					},
				},
			},
			expectedError: "",
			expectedPack: &models.V1PackManifestEntity{
				Name:        types.Ptr("test-pack"),
				Tag:         "v1.0",
				RegistryUID: "test-registry-uid",
				UID:         "test-uid",
				Type:        models.V1PackTypeSpectro.Pointer(),
				Values:      "test-values",
				Manifests: []*models.V1ManifestInputEntity{
					{
						Content: "manifest-content", // Should be trimmed
						Name:    "manifest-name",
					},
				},
			},
		},
		{
			name: "Valid Pack with empty values",
			input: map[string]interface{}{
				"name":         "test-pack",
				"type":         "spectro",
				"tag":          "v1.0",
				"uid":          "test-uid",
				"registry_uid": "test-registry-uid",
				"values":       "",
				"manifest":     []interface{}{},
			},
			expectedError: "",
			expectedPack: &models.V1PackManifestEntity{
				Name:        types.Ptr("test-pack"),
				Tag:         "v1.0",
				RegistryUID: "test-registry-uid",
				UID:         "test-uid",
				Type:        models.V1PackTypeSpectro.Pointer(),
				Values:      "", // Empty values should be preserved
				Manifests:   []*models.V1ManifestInputEntity{},
			},
		},
		{
			name: "Valid Pack with empty manifest list",
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
			name: "Valid Spectro Pack with all resolution fields (no UID)",
			input: map[string]interface{}{
				"name":         "test-pack",
				"type":         "spectro",
				"tag":          "v1.0",
				"uid":          "",
				"registry_uid": "test-registry-uid",
				"values":       "test-values",
				"manifest":     []interface{}{},
			},
			expectedError: "",
			expectedPack: &models.V1PackManifestEntity{
				Name:        types.Ptr("test-pack"),
				Tag:         "v1.0",
				RegistryUID: "test-registry-uid",
				UID:         "", // UID will be resolved in toClusterProfilePackCreateWithResolution
				Type:        models.V1PackTypeSpectro.Pointer(),
				Values:      "test-values",
				Manifests:   []*models.V1ManifestInputEntity{},
			},
		},
		{
			name: "Valid Spectro Pack with registry_name instead of registry_uid",
			input: map[string]interface{}{
				"name":          "test-pack",
				"type":          "spectro",
				"tag":           "v1.0",
				"uid":           "",
				"registry_name": "test-registry-name",
				"values":        "test-values",
				"manifest":      []interface{}{},
			},
			expectedError: "",
			expectedPack: &models.V1PackManifestEntity{
				Name:        types.Ptr("test-pack"),
				Tag:         "v1.0",
				RegistryUID: "", // registry_name is stored but not resolved here
				UID:         "",
				Type:        models.V1PackTypeSpectro.Pointer(),
				Values:      "test-values",
				Manifests:   []*models.V1ManifestInputEntity{},
			},
		},
		{
			name: "Error - Spectro Pack missing tag without UID",
			input: map[string]interface{}{
				"name":         "test-pack",
				"type":         "spectro",
				"tag":          "",
				"uid":          "",
				"registry_uid": "test-registry-uid",
				"values":       "test-values",
				"manifest":     []interface{}{},
			},
			expectedError: "pack test-pack: either 'uid' must be provided, or all of the following fields must be specified for pack resolution: name, tag, registry_uid (or registry_name). Missing: tag",
			expectedPack:  nil,
		},
		{
			name: "Error - Spectro Pack missing name without UID",
			input: map[string]interface{}{
				"name":         "",
				"type":         "spectro",
				"tag":          "v1.0",
				"uid":          "",
				"registry_uid": "test-registry-uid",
				"values":       "test-values",
				"manifest":     []interface{}{},
			},
			expectedError: "pack : either 'uid' must be provided, or all of the following fields must be specified for pack resolution: name, tag, registry_uid (or registry_name). Missing: name",
			expectedPack:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function under test
			actualPack, err := toClusterProfilePackCreate(tt.input)

			// Check for errors
			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				assert.Nil(t, actualPack, "Pack should be nil on error")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, actualPack, "Pack should not be nil on success")
				if actualPack != nil && tt.expectedPack != nil {
					assert.Equal(t, tt.expectedPack.Name, actualPack.Name, "Name should match")
					assert.Equal(t, tt.expectedPack.Tag, actualPack.Tag, "Tag should match")
					assert.Equal(t, tt.expectedPack.RegistryUID, actualPack.RegistryUID, "RegistryUID should match")
					assert.Equal(t, tt.expectedPack.UID, actualPack.UID, "UID should match")
					assert.Equal(t, tt.expectedPack.Type, actualPack.Type, "Type should match")
					assert.Equal(t, tt.expectedPack.Values, actualPack.Values, "Values should match")
					assert.Equal(t, len(tt.expectedPack.Manifests), len(actualPack.Manifests), "Manifests length should match")
					for i, expectedManifest := range tt.expectedPack.Manifests {
						if i < len(actualPack.Manifests) {
							assert.Equal(t, expectedManifest.Name, actualPack.Manifests[i].Name, "Manifest name should match")
							assert.Equal(t, expectedManifest.Content, actualPack.Manifests[i].Content, "Manifest content should match")
						}
					}
				}
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

// TestToClusterProfileCreateWithResolution tests the toClusterProfileCreateWithResolution function.
// This function:
// 1. Creates a basic cluster profile using toClusterProfileBasic
// 2. Resolves and processes packs using toClusterProfilePackCreateWithResolution
// 3. Sets profile variables using toClusterProfileVariables
// 4. Returns the complete cluster profile entity
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
			name: "Successful creation with empty packs",
			setup: func() (*schema.ResourceData, *client.V1Client) {
				d := resourceClusterProfile().TestResourceData()
				_ = d.Set("name", "test-profile")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("cloud", "azure")
				_ = d.Set("type", "cluster")
				_ = d.Set("pack", []interface{}{})
				_ = d.Set("profile_variables", []interface{}{
					map[string]interface{}{
						"variable": []interface{}{},
					},
				})
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return d, c
			},
			expectError: false,
			description: "Should successfully create cluster profile with empty packs",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "test-profile", cp.Metadata.Name, "Name should match")
				assert.NotNil(t, cp.Spec.Template.Packs, "Packs should not be nil")
				assert.Equal(t, 0, len(cp.Spec.Template.Packs), "Packs should be empty")
			},
		},
		{
			name: "Successful creation with empty variables",
			setup: func() (*schema.ResourceData, *client.V1Client) {
				d := resourceClusterProfile().TestResourceData()
				_ = d.Set("name", "test-profile")
				_ = d.Set("version", "2.0.0")
				_ = d.Set("cloud", "gcp")
				_ = d.Set("type", "infra")
				_ = d.Set("pack", []interface{}{
					map[string]interface{}{
						"uid":          "test-pack-uid-2",
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
			description: "Should successfully create cluster profile with empty variables",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "test-profile", cp.Metadata.Name, "Name should match")
				// Variables can be nil or empty slice when no variables are provided
				if cp.Spec.Variables != nil {
					assert.Equal(t, 0, len(cp.Spec.Variables), "Variables should be empty if not nil")
				}
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
			name: "Error from pack resolution - missing registry_uid",
			setup: func() (*schema.ResourceData, *client.V1Client) {
				d := resourceClusterProfile().TestResourceData()
				_ = d.Set("name", "test-profile")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("cloud", "aws")
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
				_ = d.Set("profile_variables", []interface{}{
					map[string]interface{}{
						"variable": []interface{}{},
					},
				})
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return d, c
			},
			expectError: true,
			description: "Should return error when pack resolution fails due to missing registry_uid",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity, err error) {
				assert.Error(t, err, "Should have error")
				assert.Nil(t, cp, "Cluster profile should be nil on error")
				assert.Contains(t, err.Error(), "either 'uid' must be provided", "Error should mention missing fields")
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
		{
			name: "Successful creation with tags",
			setup: func() (*schema.ResourceData, *client.V1Client) {
				d := resourceClusterProfile().TestResourceData()
				_ = d.Set("name", "test-profile")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("cloud", "aws")
				_ = d.Set("type", "add-on")
				_ = d.Set("tags", []interface{}{"tag1:value1", "tag2:value2"})
				_ = d.Set("pack", []interface{}{})
				_ = d.Set("profile_variables", []interface{}{
					map[string]interface{}{
						"variable": []interface{}{},
					},
				})
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return d, c
			},
			expectError: false,
			description: "Should successfully create cluster profile with tags",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "test-profile", cp.Metadata.Name, "Name should match")
				assert.NotNil(t, cp.Metadata.Labels, "Labels should not be nil")
			},
		},
		{
			name: "Successful creation with description",
			setup: func() (*schema.ResourceData, *client.V1Client) {
				d := resourceClusterProfile().TestResourceData()
				_ = d.Set("name", "test-profile")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("description", "This is a test description")
				_ = d.Set("cloud", "azure")
				_ = d.Set("type", "cluster")
				_ = d.Set("pack", []interface{}{})
				_ = d.Set("profile_variables", []interface{}{
					map[string]interface{}{
						"variable": []interface{}{},
					},
				})
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return d, c
			},
			expectError: false,
			description: "Should successfully create cluster profile with description",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "test-profile", cp.Metadata.Name, "Name should match")
				assert.NotNil(t, cp.Metadata.Annotations, "Annotations should not be nil")
				if cp.Metadata.Annotations != nil {
					assert.Equal(t, "This is a test description", cp.Metadata.Annotations["description"], "Description should match")
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

// TestToClusterProfileBasic tests the toClusterProfileBasic function.
// This function:
// 1. Extracts description from ResourceData (can be nil or empty)
// 2. Creates V1ClusterProfileEntity with metadata (name, UID, annotations, labels)
// 3. Creates spec with Template (CloudType, Type) and Version
// 4. Returns the basic cluster profile entity
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
			name: "Successful creation without description",
			setup: func() *schema.ResourceData {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("test-profile-uid-2")
				_ = d.Set("name", "test-profile-2")
				_ = d.Set("version", "2.0.0")
				_ = d.Set("cloud", "azure")
				_ = d.Set("type", "cluster")
				// Don't set description
				return d
			},
			expectError: false,
			description: "Should successfully create basic cluster profile without description",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity) {
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "test-profile-2", cp.Metadata.Name, "Name should match")
				assert.Equal(t, "", cp.Metadata.Annotations["description"], "Description should be empty string")
				assert.Equal(t, "azure", cp.Spec.Template.CloudType, "Cloud type should match")
				assert.Equal(t, "cluster", string(*cp.Spec.Template.Type), "Type should match")
			},
		},
		{
			name: "Successful creation with empty description",
			setup: func() *schema.ResourceData {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("test-profile-uid-3")
				_ = d.Set("name", "test-profile-3")
				_ = d.Set("version", "3.0.0")
				_ = d.Set("description", "")
				_ = d.Set("cloud", "gcp")
				_ = d.Set("type", "infra")
				return d
			},
			expectError: false,
			description: "Should successfully create basic cluster profile with empty description",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity) {
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "test-profile-3", cp.Metadata.Name, "Name should match")
				assert.Equal(t, "", cp.Metadata.Annotations["description"], "Description should be empty string")
				assert.Equal(t, "gcp", cp.Spec.Template.CloudType, "Cloud type should match")
				assert.Equal(t, "infra", string(*cp.Spec.Template.Type), "Type should match")
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
		{
			name: "Successful creation without tags",
			setup: func() *schema.ResourceData {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("test-profile-uid-5")
				_ = d.Set("name", "test-profile-5")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("cloud", "vsphere")
				_ = d.Set("type", "system")
				// Don't set tags
				return d
			},
			expectError: false,
			description: "Should successfully create basic cluster profile without tags",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity) {
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "test-profile-5", cp.Metadata.Name, "Name should match")
				assert.NotNil(t, cp.Metadata.Labels, "Labels should not be nil")
			},
		},
		{
			name: "Successful creation with different cloud types",
			setup: func() *schema.ResourceData {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("test-profile-uid-6")
				_ = d.Set("name", "test-profile-6")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("cloud", "all")
				_ = d.Set("type", "add-on")
				return d
			},
			expectError: false,
			description: "Should successfully create basic cluster profile with 'all' cloud type",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity) {
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "all", cp.Spec.Template.CloudType, "Cloud type should be 'all'")
			},
		},
		{
			name: "Successful creation with custom cloud type",
			setup: func() *schema.ResourceData {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("test-profile-uid-7")
				_ = d.Set("name", "test-profile-7")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("cloud", "nutanix")
				_ = d.Set("type", "cluster")
				return d
			},
			expectError: false,
			description: "Should successfully create basic cluster profile with custom cloud type",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity) {
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "nutanix", cp.Spec.Template.CloudType, "Cloud type should be 'nutanix'")
			},
		},
		{
			name: "Successful creation with different profile types",
			setup: func() *schema.ResourceData {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("test-profile-uid-8")
				_ = d.Set("name", "test-profile-8")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("cloud", "aws")
				_ = d.Set("type", "system")
				return d
			},
			expectError: false,
			description: "Should successfully create basic cluster profile with system type",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity) {
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "system", string(*cp.Spec.Template.Type), "Type should be 'system'")
			},
		},
		{
			name: "Successful creation with different versions",
			setup: func() *schema.ResourceData {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("test-profile-uid-9")
				_ = d.Set("name", "test-profile-9")
				_ = d.Set("version", "10.20.30-beta.1")
				_ = d.Set("cloud", "aws")
				_ = d.Set("type", "add-on")
				return d
			},
			expectError: false,
			description: "Should successfully create basic cluster profile with complex version",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity) {
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "10.20.30-beta.1", cp.Spec.Version, "Version should match")
			},
		},
		{
			name: "Successful creation with empty UID",
			setup: func() *schema.ResourceData {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("") // Empty UID
				_ = d.Set("name", "test-profile-10")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("cloud", "aws")
				_ = d.Set("type", "add-on")
				return d
			},
			expectError: false,
			description: "Should successfully create basic cluster profile with empty UID",
			verify: func(t *testing.T, cp *models.V1ClusterProfileEntity) {
				assert.NotNil(t, cp, "Cluster profile should not be nil")
				assert.Equal(t, "", cp.Metadata.UID, "UID should be empty string")
				assert.Equal(t, "test-profile-10", cp.Metadata.Name, "Name should match")
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

// TestToClusterProfileUpdateWithResolution tests the toClusterProfileUpdateWithResolution function.
// This function:
// 1. Creates a V1ClusterProfileUpdateEntity with metadata (name, UID) and spec (Type, Version)
// 2. Resolves and processes packs using toClusterProfilePackUpdateWithResolution
// 3. Sets the packs on the cluster profile update entity
// 4. Returns the complete cluster profile update entity
func TestToClusterProfileUpdateWithResolution(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (*schema.ResourceData, *models.V1ClusterProfile, *client.V1Client)
		expectError bool
		description string
		verify      func(t *testing.T, cp *models.V1ClusterProfileUpdateEntity, err error)
	}{
		{
			name: "Successful update with packs",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile, *client.V1Client) {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("test-profile-uid")
				_ = d.Set("name", "test-profile")
				_ = d.Set("version", "2.0.0")
				_ = d.Set("type", "add-on")
				_ = d.Set("pack", []interface{}{
					map[string]interface{}{
						"uid":          "test-pack-uid-1",
						"type":         "spectro",
						"name":         "test-pack",
						"registry_uid": "test-registry-uid",
						"tag":          "v1.0.0",
						"values":       "test values",
						"manifest": []interface{}{
							map[string]interface{}{
								"name":    "manifest1",
								"content": "manifest content",
							},
						},
					},
				})
				cluster := &models.V1ClusterProfile{
					Spec: &models.V1ClusterProfileSpec{
						Published: &models.V1ClusterProfileTemplate{
							Packs: []*models.V1PackRef{
								{
									PackUID: "test-pack-uid-1",
									Name:    types.Ptr("test-pack"),
									Manifests: []*models.V1ObjectReference{
										{
											UID:  "manifest-uid-1",
											Name: "manifest1",
										},
									},
								},
							},
						},
					},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return d, cluster, c
			},
			expectError: false,
			description: "Should successfully create update entity with packs",
			verify: func(t *testing.T, cp *models.V1ClusterProfileUpdateEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, cp, "Cluster profile update entity should not be nil")
				assert.Equal(t, "test-profile", cp.Metadata.Name, "Name should match")
				assert.Equal(t, "test-profile-uid", cp.Metadata.UID, "UID should match")
				assert.Equal(t, "2.0.0", cp.Spec.Version, "Version should match")
				assert.Equal(t, "add-on", string(*cp.Spec.Template.Type), "Type should match")
				assert.NotNil(t, cp.Spec.Template.Packs, "Packs should not be nil")
			},
		},
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
			name: "Successful update with manifest pack type",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile, *client.V1Client) {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("test-profile-uid-5")
				_ = d.Set("name", "test-profile-5")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("type", "system")
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
				cluster := &models.V1ClusterProfile{
					Spec: &models.V1ClusterProfileSpec{
						Published: &models.V1ClusterProfileTemplate{
							Packs: []*models.V1PackRef{
								{
									PackUID: "spectro-manifest-pack",
									Name:    types.Ptr("manifest-pack"),
									Manifests: []*models.V1ObjectReference{
										{
											UID:  "manifest-uid-1",
											Name: "manifest1",
										},
									},
								},
							},
						},
					},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return d, cluster, c
			},
			expectError: false,
			description: "Should successfully create update entity with manifest pack type",
			verify: func(t *testing.T, cp *models.V1ClusterProfileUpdateEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, cp, "Cluster profile update entity should not be nil")
				assert.Equal(t, "test-profile-5", cp.Metadata.Name, "Name should match")
				assert.Equal(t, "system", string(*cp.Spec.Template.Type), "Type should match")
				assert.NotNil(t, cp.Spec.Template.Packs, "Packs should not be nil")
				if len(cp.Spec.Template.Packs) > 0 {
					assert.Equal(t, "spectro-manifest-pack", cp.Spec.Template.Packs[0].UID, "Manifest pack should have default UID")
				}
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
			name: "Panic when cluster is nil",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile, *client.V1Client) {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("test-profile-uid-7")
				_ = d.Set("name", "test-profile-7")
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
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return d, nil, c
			},
			expectError: true,
			description: "Should panic when cluster is nil",
			verify: func(t *testing.T, cp *models.V1ClusterProfileUpdateEntity, err error) {
				// Function will panic on nil pointer dereference
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
		{
			name: "Successful update with empty UID",
			setup: func() (*schema.ResourceData, *models.V1ClusterProfile, *client.V1Client) {
				d := resourceClusterProfile().TestResourceData()
				d.SetId("") // Empty UID
				_ = d.Set("name", "test-profile-10")
				_ = d.Set("version", "1.0.0")
				_ = d.Set("type", "add-on")
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
			description: "Should successfully create update entity with empty UID",
			verify: func(t *testing.T, cp *models.V1ClusterProfileUpdateEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, cp, "Cluster profile update entity should not be nil")
				assert.Equal(t, "", cp.Metadata.UID, "UID should be empty string")
				assert.Equal(t, "test-profile-10", cp.Metadata.Name, "Name should match")
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

// TestToClusterProfilePackCreateWithResolution tests the toClusterProfilePackCreateWithResolution function.
// This function:
// 1. Extracts pack fields from input map
// 2. Validates pack UID or resolution fields
// 3. Resolves registry_name to registry_uid if provided
// 4. Resolves pack UID for Spectro/Helm packs if not provided
// 5. Handles Manifest packs with default UID
// 6. Trims whitespace from values and manifest content
// 7. Returns V1PackManifestEntity
func TestToClusterProfilePackCreateWithResolution(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (map[string]interface{}, *client.V1Client)
		expectError bool
		description string
		verify      func(t *testing.T, pack *models.V1PackManifestEntity, err error)
	}{
		{
			name: "Successful creation with UID provided",
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
			description: "Should successfully create pack with UID provided",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, pack, "Pack should not be nil")
				assert.Equal(t, "test-pack", *pack.Name, "Name should match")
				assert.Equal(t, "test-uid", pack.UID, "UID should match")
				assert.Equal(t, "test-registry-uid", pack.RegistryUID, "RegistryUID should match")
				assert.Equal(t, "v1.0", pack.Tag, "Tag should match")
				assert.Equal(t, models.V1PackTypeSpectro, *pack.Type, "Type should be Spectro")
			},
		},
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
			name: "Successful creation with multiple manifests",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":         "test-pack",
					"type":         "spectro",
					"tag":          "v1.0",
					"uid":          "test-uid",
					"registry_uid": "test-registry-uid",
					"values":       "test-values",
					"manifest": []interface{}{
						map[string]interface{}{
							"content": "manifest-content-1",
							"name":    "manifest-name-1",
						},
						map[string]interface{}{
							"content": "manifest-content-2",
							"name":    "manifest-name-2",
						},
					},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return input, c
			},
			expectError: false,
			description: "Should successfully create pack with multiple manifests",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, pack, "Pack should not be nil")
				assert.Equal(t, 2, len(pack.Manifests), "Should have two manifests")
				assert.Equal(t, "manifest-name-1", pack.Manifests[0].Name, "First manifest name should match")
				assert.Equal(t, "manifest-content-1", pack.Manifests[0].Content, "First manifest content should match")
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
			name: "Error from validation - missing registry_uid",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":     "test-pack",
					"type":     "spectro",
					"tag":      "v1.0",
					"uid":      "",
					"values":   "test-values",
					"manifest": []interface{}{},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return input, c
			},
			expectError: true,
			description: "Should return error when validation fails due to missing registry_uid",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.Error(t, err, "Should have error")
				assert.Nil(t, pack, "Pack should be nil on error")
				assert.Contains(t, err.Error(), "either 'uid' must be provided", "Error should mention missing fields")
			},
		},
		{
			name: "Error from validation - missing tag",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":         "test-pack",
					"type":         "spectro",
					"tag":          "",
					"uid":          "",
					"registry_uid": "test-registry-uid",
					"values":       "test-values",
					"manifest":     []interface{}{},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return input, c
			},
			expectError: true,
			description: "Should return error when validation fails due to missing tag",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.Error(t, err, "Should have error")
				assert.Nil(t, pack, "Pack should be nil on error")
				assert.Contains(t, err.Error(), "either 'uid' must be provided", "Error should mention missing fields")
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
			name: "Successful creation with empty manifest list",
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
			description: "Should successfully create pack with empty manifest list",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, pack, "Pack should not be nil")
				assert.Equal(t, 0, len(pack.Manifests), "Manifests should be empty")
			},
		},
		{
			name: "Error from registry_name resolution",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":          "test-pack",
					"type":          "spectro",
					"tag":           "v1.0",
					"uid":           "",
					"registry_name": "non-existent-registry",
					"values":        "test-values",
					"manifest":      []interface{}{},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return input, c
			},
			expectError: true,
			description: "Should return error when registry_name resolution fails",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.Error(t, err, "Should have error")
				assert.Nil(t, pack, "Pack should be nil on error")
				// Error might be from registry resolution or pack UID resolution
			},
		},
		{
			name: "Error from pack UID resolution for Spectro pack",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":         "non-existent-pack",
					"type":         "spectro",
					"tag":          "v1.0",
					"uid":          "",
					"registry_uid": "test-registry-uid",
					"values":       "test-values",
					"manifest":     []interface{}{},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return input, c
			},
			expectError: true,
			description: "Should return error when pack UID resolution fails for Spectro pack",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.Error(t, err, "Should have error")
				assert.Nil(t, pack, "Pack should be nil on error")
				assert.Contains(t, err.Error(), "failed to resolve pack UID", "Error should mention pack UID resolution")
			},
		},
		{
			name: "Error from pack UID resolution for Helm pack",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":         "non-existent-pack",
					"type":         "helm",
					"tag":          "v1.0",
					"uid":          "",
					"registry_uid": "test-registry-uid",
					"values":       "test-values",
					"manifest":     []interface{}{},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return input, c
			},
			expectError: true,
			description: "Should return error when pack UID resolution fails for Helm pack",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.Error(t, err, "Should have error")
				assert.Nil(t, pack, "Pack should be nil on error")
				assert.Contains(t, err.Error(), "failed to resolve pack UID", "Error should mention pack UID resolution")
			},
		},
		{
			name: "Error from registry_name resolution (even with UID provided)",
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
			expectError: true,
			description: "Should return error when registry_name resolution fails (function resolves registry_name even if UID is provided)",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.Error(t, err, "Should have error")
				assert.Nil(t, pack, "Pack should be nil on error")
				// Error from registry_name resolution
			},
		},
		{
			name: "Error when both registry_uid and registry_name are provided",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":          "test-pack",
					"type":          "spectro",
					"tag":           "v1.0",
					"uid":           "test-uid",
					"registry_uid":  "test-registry-uid",
					"registry_name": "test-registry-name",
					"values":        "test-values",
					"manifest":      []interface{}{},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return input, c
			},
			expectError: true,
			description: "Should return error when both registry_uid and registry_name are provided (validation error)",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.Error(t, err, "Should have error")
				assert.Nil(t, pack, "Pack should be nil on error")
				assert.Contains(t, err.Error(), "only one of 'registry_uid' or 'registry_name' can be specified", "Error should mention both fields")
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
			name: "Successful creation with Helm pack type",
			setup: func() (map[string]interface{}, *client.V1Client) {
				input := map[string]interface{}{
					"name":         "test-helm-pack",
					"type":         "helm",
					"tag":          "v1.0",
					"uid":          "test-helm-uid",
					"registry_uid": "test-registry-uid",
					"values":       "test-values",
					"manifest":     []interface{}{},
				}
				c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
				return input, c
			},
			expectError: false,
			description: "Should successfully create Helm pack",
			verify: func(t *testing.T, pack *models.V1PackManifestEntity, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.NotNil(t, pack, "Pack should not be nil")
				assert.Equal(t, models.V1PackTypeHelm, *pack.Type, "Type should be Helm")
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
