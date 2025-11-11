package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func prepareBaseClusterConfigTemplateTestData() *schema.ResourceData {
	d := resourceClusterConfigTemplate().TestResourceData()
	_ = d.Set("name", "test-cluster-config-template")
	_ = d.Set("context", "project")
	_ = d.Set("description", "Test cluster config template")
	tags := schema.NewSet(schema.HashString, []interface{}{
		"env:test",
		"team:platform",
	})
	_ = d.Set("tags", tags)
	_ = d.Set("cloud_type", "aws")

	// Create variables set
	variablesSet := schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{
		map[string]interface{}{
			"name":            "region",
			"value":           "us-west-2",
			"assign_strategy": "all",
		},
		map[string]interface{}{
			"name":            "instance_type",
			"value":           "t3.medium",
			"assign_strategy": "all",
		},
	})

	// Create profiles set
	profilesSet := schema.NewSet(resourceClusterConfigTemplateProfileHash, []interface{}{
		map[string]interface{}{
			"uid":       "test-profile-uid-1",
			"variables": variablesSet,
		},
	})

	_ = d.Set("profiles", profilesSet)
	_ = d.Set("policies", []interface{}{
		map[string]interface{}{
			"uid":  "test-policy-uid-1",
			"kind": "maintenance",
		},
	})
	d.SetId("test-cluster-config-template-id")
	return d
}

func TestResourceClusterConfigTemplateCreate(t *testing.T) {
	d := prepareBaseClusterConfigTemplateTestData()
	var ctx context.Context
	diags := resourceClusterConfigTemplateCreate(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Equal(t, "test-cluster-config-template-id", d.Id())
}

func TestResourceClusterConfigTemplateRead(t *testing.T) {
	d := prepareBaseClusterConfigTemplateTestData()
	var ctx context.Context
	diags := resourceClusterConfigTemplateRead(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Equal(t, "test-cluster-config-template-id", d.Id())
}

func TestResourceClusterConfigTemplateUpdate(t *testing.T) {
	d := prepareBaseClusterConfigTemplateTestData()
	var ctx context.Context
	diags := resourceClusterConfigTemplateUpdate(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Equal(t, "test-cluster-config-template-id", d.Id())
}

func TestResourceClusterConfigTemplateDelete(t *testing.T) {
	d := prepareBaseClusterConfigTemplateTestData()
	var ctx context.Context
	diags := resourceClusterConfigTemplateDelete(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestExpandClusterTemplateProfiles(t *testing.T) {
	profiles := []interface{}{
		map[string]interface{}{
			"uid": "profile-uid-1",
		},
		map[string]interface{}{
			"uid": "profile-uid-2",
		},
	}

	result := expandClusterTemplateProfiles(profiles)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "profile-uid-1", result[0].UID)
	assert.Equal(t, "profile-uid-2", result[1].UID)
}

func TestExpandClusterTemplatePolicies(t *testing.T) {
	policies := []interface{}{
		map[string]interface{}{
			"uid":  "policy-uid-1",
			"kind": "maintenance",
		},
	}

	result := expandClusterTemplatePolicies(policies)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "policy-uid-1", result[0].UID)
	assert.Equal(t, "maintenance", result[0].Kind)
}

func TestProfileStructureChanged(t *testing.T) {
	// Test case 1: Different number of profiles
	oldProfiles := schema.NewSet(resourceClusterConfigTemplateProfileHash, []interface{}{
		map[string]interface{}{
			"uid":       "profile-1",
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{}),
		},
	})
	newProfiles := schema.NewSet(resourceClusterConfigTemplateProfileHash, []interface{}{
		map[string]interface{}{
			"uid":       "profile-1",
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{}),
		},
		map[string]interface{}{
			"uid":       "profile-2",
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{}),
		},
	})
	assert.True(t, profileStructureChanged(oldProfiles, newProfiles), "Should detect added profile")

	// Test case 2: Same number but different UIDs
	oldProfiles = schema.NewSet(resourceClusterConfigTemplateProfileHash, []interface{}{
		map[string]interface{}{
			"uid":       "profile-1",
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{}),
		},
		map[string]interface{}{
			"uid":       "profile-2",
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{}),
		},
	})
	newProfiles = schema.NewSet(resourceClusterConfigTemplateProfileHash, []interface{}{
		map[string]interface{}{
			"uid":       "profile-1",
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{}),
		},
		map[string]interface{}{
			"uid":       "profile-3",
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{}),
		},
	})
	assert.True(t, profileStructureChanged(oldProfiles, newProfiles), "Should detect changed profile UID")

	// Test case 3: Same UIDs, only variables changed
	oldVars := schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{
		map[string]interface{}{"name": "var1", "value": "old", "assign_strategy": "all"},
	})
	newVars := schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{
		map[string]interface{}{"name": "var1", "value": "new", "assign_strategy": "all"},
	})

	oldProfiles = schema.NewSet(resourceClusterConfigTemplateProfileHash, []interface{}{
		map[string]interface{}{
			"uid":       "profile-1",
			"variables": oldVars,
		},
	})
	newProfiles = schema.NewSet(resourceClusterConfigTemplateProfileHash, []interface{}{
		map[string]interface{}{
			"uid":       "profile-1",
			"variables": newVars,
		},
	})
	assert.False(t, profileStructureChanged(oldProfiles, newProfiles), "Should not detect change when only variables differ")
}
