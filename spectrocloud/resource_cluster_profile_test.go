package spectrocloud

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
					"input_type":    "text",
					"options":       []interface{}(nil),
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
					"input_type":    "text",
					"options":       []interface{}(nil),
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

func TestToClusterProfileVariablesInputTypeAndOptions(t *testing.T) {
	// Test input_type and options (dropdown with options list)
	mockResourceData := resourceClusterProfile().TestResourceData()
	proVar := []interface{}{
		map[string]interface{}{
			"variable": []interface{}{
				map[string]interface{}{
					"name":          "env_var",
					"display_name":  "Environment",
					"format":        "string",
					"description":   "Select environment",
					"default_value": "dev",
					"required":      true,
					"immutable":     false,
					"hidden":        false,
					"is_sensitive":  false,
					"input_type":    "dropdown",
					"options": []interface{}{
						map[string]interface{}{
							"label":       "Development",
							"value":       "dev",
							"description": "Dev environment",
							"default":     true,
						},
						map[string]interface{}{
							"label":       "Production",
							"value":       "prod",
							"description": "Prod environment",
							"default":     false,
						},
					},
				},
				map[string]interface{}{
					"name":          "notes",
					"display_name":  "Notes",
					"format":        "string",
					"description":   "Multiline notes",
					"default_value": "",
					"required":      false,
					"immutable":     false,
					"hidden":        false,
					"is_sensitive":  false,
					"input_type":    "multiline",
				},
			},
		},
	}
	_ = mockResourceData.Set("cloud", "all")
	_ = mockResourceData.Set("type", "add-on")
	_ = mockResourceData.Set("profile_variables", proVar)

	result, err := toClusterProfileVariables(mockResourceData)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// First variable: dropdown with options
	assert.Equal(t, "env_var", *result[0].Name)
	assert.NotNil(t, result[0].InputType)
	assert.Equal(t, models.V1VariableInputTypeDropdown, *result[0].InputType)
	assert.Len(t, result[0].Options, 2)
	assert.Equal(t, "dev", *result[0].Options[0].Value)
	assert.Equal(t, "Development", result[0].Options[0].Label)
	assert.Equal(t, "Dev environment", result[0].Options[0].Description)
	assert.True(t, result[0].Options[0].Default)
	assert.Equal(t, "prod", *result[0].Options[1].Value)
	assert.Equal(t, "Production", result[0].Options[1].Label)
	assert.False(t, result[0].Options[1].Default)

	// Second variable: multiline, no options
	assert.Equal(t, "notes", *result[1].Name)
	assert.NotNil(t, result[1].InputType)
	assert.Equal(t, models.V1VariableInputTypeMultiline, *result[1].InputType)
	assert.Nil(t, result[1].Options)
}

func TestFlattenProfileVariablesInputTypeAndOptions(t *testing.T) {
	// Test flatten with input_type and options from API
	mockResourceData := resourceClusterProfile().TestResourceData()
	proVar := []interface{}{
		map[string]interface{}{
			"variable": []interface{}{
				map[string]interface{}{
					"name":         "env_var",
					"display_name": "Environment",
				},
			},
		},
	}
	_ = mockResourceData.Set("profile_variables", proVar)

	pv := []*models.V1Variable{
		{
			Name:         StringPtr("env_var"),
			DisplayName:  "Environment",
			Description:  "Select env",
			Format:       models.NewV1VariableFormat("string"),
			DefaultValue: "dev",
			InputType:    models.V1VariableInputTypeDropdown.Pointer(),
			Options: []*models.V1VariableOption{
				{Value: types.Ptr("dev"), Label: "Development", Description: "Dev", Default: true},
				{Value: types.Ptr("prod"), Label: "Production", Description: "Prod", Default: false},
			},
		},
	}

	result, err := flattenProfileVariables(mockResourceData, pv)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	outer := result[0].(map[string]interface{})
	variables := outer["variable"].([]interface{})
	assert.Len(t, variables, 1)
	flat := variables[0].(map[string]interface{})
	assert.Equal(t, "dropdown", flat["input_type"])
	opts := flat["options"].([]interface{})
	assert.Len(t, opts, 2)
	opt0 := opts[0].(map[string]interface{})
	assert.Equal(t, "dev", opt0["value"])
	assert.Equal(t, "Development", opt0["label"])
	assert.Equal(t, "Dev", opt0["description"])
	assert.True(t, opt0["default"].(bool))
	opt1 := opts[1].(map[string]interface{})
	assert.Equal(t, "prod", opt1["value"])
	assert.Equal(t, "Production", opt1["label"])
	assert.False(t, opt1["default"].(bool))
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

// prepareClusterProfileWithVersionChange creates ResourceData with a state+diff
// so that HasChange("version") returns true. Additional field changes can be
// injected via the extraDiffAttrs map.
func prepareClusterProfileWithVersionChange(oldVersion, newVersion, profileName string, extraDiffAttrs map[string]*terraform.ResourceAttrDiff) *schema.ResourceData {
	state := &terraform.InstanceState{
		ID: "cluster-profile-1",
		Attributes: map[string]string{
			"name":                profileName,
			"version":             oldVersion,
			"context":             "project",
			"description":         "old description",
			"cloud":               "all",
			"type":                "add-on",
			"tags.#":              "0",
			"pack.#":              "0",
			"profile_variables.#": "0",
		},
	}

	diffAttrs := map[string]*terraform.ResourceAttrDiff{
		"version": {
			Old: oldVersion,
			New: newVersion,
		},
	}
	for k, v := range extraDiffAttrs {
		diffAttrs[k] = v
	}

	diff := &terraform.InstanceDiff{
		Attributes: diffAttrs,
	}

	d, _ := schema.InternalMap(resourceClusterProfile().Schema).Data(state, diff)
	return d
}

// TestResourceClusterProfileUpdateVersionNoFlag tests that without the
// immutable-clusterprofiles feature flag, version changes fall through to the
// legacy in-place update path (UpdateClusterProfile / PUT) instead of triggering
// a Create-path clone. The Terraform id stays stable because no replacement is
// planned.
//
// Note: this is the backward-compat path. When the flag IS enabled, version
// changes never reach Update at all -- CustomizeDiff marks them as ForceNew, so
// Terraform plans a replacement and the new version is produced by the Create
// function (see TestResourceClusterProfileCreate_ImmutableClusterprofiles_*).
func TestResourceClusterProfileUpdateVersionNoFlag(t *testing.T) {
	orig := ProviderFeaturePreview
	defer func() { ProviderFeaturePreview = orig }()
	ProviderFeaturePreview = map[string]bool{}

	d := prepareClusterProfileWithVersionChange("1.0.0", "2.0.0", "nonexistent-profile", nil)
	var ctx context.Context

	diags := resourceClusterProfileUpdate(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	// Without the flag, version change should NOT clone -- it falls through
	// to the legacy update-in-place path and the Terraform id stays stable.
	assert.Equal(t, "cluster-profile-1", d.Id())
}

// TestResourceClusterProfileCreateAdoptExisting verifies the SDK v2
// adopt-on-create pattern: when Create fails because the profile already
// exists in Palette (e.g. another root module created it, or it was created
// via the UI) AND the immutable-clusterprofiles flag is enabled, the function
// adopts the existing UID into Terraform state instead of returning an error.
func TestResourceClusterProfileCreateAdoptExisting(t *testing.T) {
	orig := ProviderFeaturePreview
	defer func() { ProviderFeaturePreview = orig }()
	ProviderFeaturePreview = map[string]bool{"immutable-clusterprofiles": true}

	d := prepareBaseClusterProfileTestData()
	// Use name+version matching mock metadata → adopt path
	_ = d.Set("name", "test-cluster-profile-1")
	_ = d.Set("version", "1.0.0")
	_ = d.Set("type", "add-on")
	var ctx context.Context
	diags := resourceClusterProfileCreate(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	// Should have adopted the existing UID from the mock metadata.
	assert.Equal(t, "existing-profile-uid-1", d.Id())
}

// TestResourceClusterProfileCreateNoAdoptWithoutFlag verifies that when the
// immutable-clusterprofiles feature flag is OFF, a Create failure returns the
// error instead of trying to adopt an existing profile. This preserves the
// legacy "create is not idempotent" behavior for users who haven't opted into
// the new flag.
func TestResourceClusterProfileCreateNoAdoptWithoutFlag(t *testing.T) {
	orig := ProviderFeaturePreview
	defer func() { ProviderFeaturePreview = orig }()
	ProviderFeaturePreview = map[string]bool{}

	d := prepareBaseClusterProfileTestData()
	_ = d.Set("name", "test-cluster-profile-1")
	_ = d.Set("version", "1.0.0")
	_ = d.Set("type", "add-on")
	var ctx context.Context
	// On the negative client, create fails. Without the flag, it should
	// return the error instead of trying to adopt.
	diags := resourceClusterProfileCreate(ctx, d, unitTestMockAPINegativeClient)
	assert.NotEmpty(t, diags)
}

// TestResourceClusterProfileSchema_HasSkipDestroy verifies that the new
// skip_destroy schema field is present with the expected type and default.
// This field is part of the standard Terraform Plugin SDK v2 immutable-versioned
// resource pattern -- it gates whether Delete actually calls the API or just
// removes the resource from Terraform state.
func TestResourceClusterProfileSchema_HasSkipDestroy(t *testing.T) {
	r := resourceClusterProfile()
	field, ok := r.Schema["skip_destroy"]
	assert.True(t, ok, "skip_destroy field must be present on the cluster_profile schema")
	assert.NotNil(t, field)
	assert.Equal(t, schema.TypeBool, field.Type)
	assert.True(t, field.Optional)
	assert.Equal(t, false, field.Default, "skip_destroy must default to false for backward compatibility")
}

// TestResourceClusterProfileSchema_NoCurrentUid verifies that the current_uid
// field was removed as part of the consolidation. It was only useful as a
// workaround for the stale-output bug on the clone-on-version-change path,
// which no longer exists.
func TestResourceClusterProfileSchema_NoCurrentUid(t *testing.T) {
	r := resourceClusterProfile()
	_, ok := r.Schema["current_uid"]
	assert.False(t, ok, "current_uid field must not be present; it was removed when clone-on-version-change was consolidated into immutable-clusterprofiles")
}

// TestResourceClusterProfileSchema_CustomizeDiffRegistered verifies that the
// CustomizeDiff hook is wired up on the resource. CustomizeDiff is what marks
// the version field as ForceNew when the immutable-clusterprofiles feature
// flag is enabled.
func TestResourceClusterProfileSchema_CustomizeDiffRegistered(t *testing.T) {
	r := resourceClusterProfile()
	assert.NotNil(t, r.CustomizeDiff, "CustomizeDiff must be registered to gate ForceNew on version under immutable-clusterprofiles")
}

// customizeDiffFixture drives Resource.Diff (which runs the registered
// CustomizeDiff function internally) with a version bump on an existing
// resource. skipDestroyInConfig controls whether the user's HCL sets
// skip_destroy = true, which is what the CustomizeDiff plan-time validation
// checks for.
func customizeDiffFixture(oldVersion, newVersion string, skipDestroyInConfig bool) (*terraform.InstanceDiff, error) {
	r := resourceClusterProfile()
	state := &terraform.InstanceState{
		ID: "cluster-profile-1",
		Attributes: map[string]string{
			"name":                "example-addon",
			"version":             oldVersion,
			"context":             "project",
			"description":         "",
			"cloud":               "all",
			"type":                "add-on",
			"skip_destroy":        "false",
			"tags.#":              "0",
			"pack.#":              "0",
			"profile_variables.#": "0",
		},
	}
	cfg := map[string]interface{}{
		"name":         "example-addon",
		"version":      newVersion,
		"context":      "project",
		"cloud":        "all",
		"type":         "add-on",
		"skip_destroy": skipDestroyInConfig,
	}
	return r.Diff(context.Background(), state, terraform.NewResourceConfigRaw(cfg), unitTestMockAPIClient)
}

// TestResourceClusterProfileCustomizeDiff_VersionBump_MissingSkipDestroy
// verifies that when the immutable-clusterprofiles flag is enabled and the
// user bumps the version without setting skip_destroy = true, plan fails with
// a clear error that tells the user which knobs to add. This is the guardrail
// against the common mistake of enabling the flag but forgetting the companion
// SDK v2 pattern attributes.
func TestResourceClusterProfileCustomizeDiff_VersionBump_MissingSkipDestroy(t *testing.T) {
	orig := ProviderFeaturePreview
	defer func() { ProviderFeaturePreview = orig }()
	ProviderFeaturePreview = map[string]bool{"immutable-clusterprofiles": true}

	_, err := customizeDiffFixture("1.0.0", "1.1.0", false)
	assert.Error(t, err, "plan must error when version changes under the flag without skip_destroy = true")
	assert.Contains(t, err.Error(), "skip_destroy = true")
	assert.Contains(t, err.Error(), "create_before_destroy = true")
	assert.Contains(t, err.Error(), "aws_lambda_layer_version")
}

// TestResourceClusterProfileCustomizeDiff_VersionBump_WithSkipDestroy
// verifies that when skip_destroy is set, plan succeeds and the version change
// is marked as a replacement (ForceNew). This is the intended happy path under
// the immutable-clusterprofiles flag.
func TestResourceClusterProfileCustomizeDiff_VersionBump_WithSkipDestroy(t *testing.T) {
	orig := ProviderFeaturePreview
	defer func() { ProviderFeaturePreview = orig }()
	ProviderFeaturePreview = map[string]bool{"immutable-clusterprofiles": true}

	diff, err := customizeDiffFixture("1.0.0", "1.1.0", true)
	assert.NoError(t, err)
	assert.NotNil(t, diff)
	versionAttr, ok := diff.Attributes["version"]
	assert.True(t, ok, "version attribute should be in diff")
	assert.True(t, versionAttr.RequiresNew, "version change must be marked ForceNew under the flag")
}

// TestResourceClusterProfileCustomizeDiff_VersionBump_FlagOff verifies that
// without the immutable-clusterprofiles flag, the CustomizeDiff validation is
// bypassed entirely and version changes behave like any other in-place update
// -- no ForceNew, no skip_destroy requirement. This is the backward-compat path.
func TestResourceClusterProfileCustomizeDiff_VersionBump_FlagOff(t *testing.T) {
	orig := ProviderFeaturePreview
	defer func() { ProviderFeaturePreview = orig }()
	ProviderFeaturePreview = map[string]bool{}

	diff, err := customizeDiffFixture("1.0.0", "1.1.0", false)
	assert.NoError(t, err, "without the flag, version changes must not require skip_destroy")
	assert.NotNil(t, diff)
	if versionAttr, ok := diff.Attributes["version"]; ok {
		assert.False(t, versionAttr.RequiresNew, "without the flag, version must not be ForceNew")
	}
}

// TestResourceClusterProfileDelete_SkipDestroy verifies that when
// skip_destroy=true, the Delete function returns successfully without calling
// the Palette delete API. This is the SDK v2 preservation pattern for
// immutable-versioned resources: replacement lifecycles remove the old
// resource from Terraform state via Delete, and skip_destroy makes that a
// no-op so the underlying versioned object stays in the upstream system.
func TestResourceClusterProfileDelete_SkipDestroy(t *testing.T) {
	d := prepareBaseClusterProfileTestData()
	d.SetId("some-uid-that-does-not-exist-in-mock")
	_ = d.Set("skip_destroy", true)
	_ = d.Set("context", "project")
	var ctx context.Context

	// Delete against the negative (failing) client -- if skip_destroy is
	// honored, we never call the API, so the negative client's failure
	// path doesn't trigger.
	diags := resourceClusterProfileDelete(ctx, d, unitTestMockAPINegativeClient)
	assert.Empty(t, diags, "skip_destroy=true should bypass the API call entirely, so negative-client failures should not surface")
}

// TestResourceClusterProfileDelete_NormalDestroy verifies that when
// skip_destroy is false (the default), Delete calls the Palette delete API.
// This preserves the legacy behavior for users who don't opt into the new
// immutable lifecycle.
func TestResourceClusterProfileDelete_NormalDestroy(t *testing.T) {
	d := prepareBaseClusterProfileTestData()
	d.SetId("test-cluster-profile-1")
	_ = d.Set("skip_destroy", false)
	_ = d.Set("context", "project")
	var ctx context.Context

	// The mock client's delete endpoint returns success for known UIDs.
	diags := resourceClusterProfileDelete(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

// TestFindAnyExistingProfileVersionUID_Found verifies that the helper finds
// an existing profile by name via the SDK's GetClusterProfiles listing endpoint.
// This helper is used by the immutable-clusterprofiles Create path to discover
// a clone source for a new version of an existing lineage.
func TestFindAnyExistingProfileVersionUID_Found(t *testing.T) {
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	// The mock metadata includes "test-cluster-profile-1" with a known stable UID.
	uid, err := findAnyExistingProfileVersionUID(c, "test-cluster-profile-1")
	assert.NoError(t, err)
	assert.NotEmpty(t, uid, "expected to find an existing UID for test-cluster-profile-1 in the mock metadata")
}

// TestFindAnyExistingProfileVersionUID_NotFound verifies that the helper
// returns an empty string (not an error) when no profile with the given name
// exists. The empty-string-no-error return is what the Create path uses to
// decide between "clone from existing lineage" and "create the very first
// version from scratch".
func TestFindAnyExistingProfileVersionUID_NotFound(t *testing.T) {
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	uid, err := findAnyExistingProfileVersionUID(c, "this-profile-definitely-does-not-exist-in-the-mock")
	assert.NoError(t, err, "not-found should return an empty string without an error")
	assert.Empty(t, uid)
}
