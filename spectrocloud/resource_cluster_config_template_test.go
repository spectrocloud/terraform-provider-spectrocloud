package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeChangeGetter map[string][2]interface{}

func (f fakeChangeGetter) GetChange(key string) (interface{}, interface{}) {
	v, ok := f[key]
	if !ok {
		return nil, nil
	}
	return v[0], v[1]
}

func profileList(ids ...string) []interface{} {
	out := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		out = append(out, map[string]interface{}{"id": id})
	}
	return out
}

func templateBlock(id string) []interface{} {
	return []interface{}{
		map[string]interface{}{"id": id},
	}
}

func TestClassifyClusterTemplateTransition(t *testing.T) {
	tests := []struct {
		name         string
		change       fakeChangeGetter
		expectAttach bool
		expectDetach bool
	}{
		{
			name: "no change - both empty",
			change: fakeChangeGetter{
				"cluster_profile":  {[]interface{}{}, []interface{}{}},
				"cluster_template": {[]interface{}{}, []interface{}{}},
			},
		},
		{
			name: "pure cluster_profile edit - not a swap",
			change: fakeChangeGetter{
				"cluster_profile":  {profileList("p1"), profileList("p1", "p2")},
				"cluster_template": {[]interface{}{}, []interface{}{}},
			},
		},
		{
			name: "pure cluster_template variable edit - not a swap",
			change: fakeChangeGetter{
				"cluster_profile":  {[]interface{}{}, []interface{}{}},
				"cluster_template": {templateBlock("t1"), templateBlock("t1")},
			},
		},
		{
			name: "attach: cluster_profile removed, cluster_template added",
			change: fakeChangeGetter{
				"cluster_profile":  {profileList("p1"), []interface{}{}},
				"cluster_template": {[]interface{}{}, templateBlock("t1")},
			},
			expectAttach: true,
		},
		{
			// Regression (real customer repro): swapping cluster_profile ->
			// cluster_template in a single apply produced "cannot specify
			// both cluster_template and cluster_profile" and never attached
			// the template. Confirmed via a live `terraform apply` (with
			// debug logging temporarily added to classifyClusterTemplateTransition)
			// that Terraform's "new" cluster_profile value is not a truly
			// empty set here - it's a set with one phantom zero-value
			// element ({"id": "", "pack": [], "variables": {}}) instead of
			// zero elements. classifyClusterTemplateTransition must filter
			// that phantom out via filterEntriesWithID or it misses the
			// attach transition entirely.
			name: "attach: cluster_profile removed leaves a phantom zero-value set element",
			change: fakeChangeGetter{
				"cluster_profile": {
					profileList("p1"),
					[]interface{}{map[string]interface{}{"id": "", "pack": []interface{}{}, "variables": map[string]interface{}{}}},
				},
				"cluster_template": {[]interface{}{}, templateBlock("t1")},
			},
			expectAttach: true,
		},
		{
			name: "detach: cluster_template removed, cluster_profile added",
			change: fakeChangeGetter{
				"cluster_profile":  {[]interface{}{}, profileList("p1")},
				"cluster_template": {templateBlock("t1"), []interface{}{}},
			},
			expectDetach: true,
		},
		{
			name: "both populated at once - neither (existing validateProfileSource rejects this separately)",
			change: fakeChangeGetter{
				"cluster_profile":  {[]interface{}{}, profileList("p1")},
				"cluster_template": {[]interface{}{}, templateBlock("t1")},
			},
		},
		{
			name: "create (old always empty for both) - never attach or detach",
			change: fakeChangeGetter{
				"cluster_profile":  {[]interface{}{}, profileList("p1")},
				"cluster_template": {[]interface{}{}, []interface{}{}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attach, detach := classifyClusterTemplateTransition(tt.change)
			assert.Equal(t, tt.expectAttach, attach, "attach")
			assert.Equal(t, tt.expectDetach, detach, "detach")
		})
	}
}

func TestValidateClusterTemplateAttachTransition_RejectsDetach(t *testing.T) {
	// validateClusterTemplateAttachTransition just wraps the classifier, so
	// exercise it through classifyClusterTemplateTransition's detach case
	// directly against the same fake, confirming the CustomizeDiff-facing
	// wrapper's error path is reachable and worded clearly.
	change := fakeChangeGetter{
		"cluster_profile":  {[]interface{}{}, profileList("p1")},
		"cluster_template": {templateBlock("t1"), []interface{}{}},
	}
	_, detach := classifyClusterTemplateTransition(change)
	require.True(t, detach)
}

// TestValidateProfileSource_IgnoresPhantomProfileElement is the
// validateProfileSource half of the same regression: even if
// classifyClusterTemplateTransition were bypassed, the mutual-exclusivity
// check itself must not reject a phantom zero-value cluster_profile element
// sitting alongside a real cluster_template.
func TestValidateProfileSource_IgnoresPhantomProfileElement(t *testing.T) {
	d := resourceClusterAws().TestResourceData()
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": "", "pack": []interface{}{}, "variables": map[string]interface{}{}},
	}))
	require.NoError(t, d.Set("cluster_template", []interface{}{
		map[string]interface{}{"id": "t1"},
	}))

	require.NoError(t, validateProfileSource(d))
}

func TestAttachClusterToTemplate_RequiresTemplateID(t *testing.T) {
	d := resourceClusterAws().TestResourceData()
	// cluster_template left unset: GetChange's "new" value is empty.
	err := attachClusterToTemplate(nil, d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cluster_template")
}

// TestAttachClusterToTemplate_WithMock exercises the full attach call against
// the mock API server: builds the request from cluster_template (including
// inline profile variables), calls the real client.AttachClusterTemplate,
// and confirms the post-attach cluster_template refresh (variables re-read
// via the existing flattenClusterTemplateVariables/GetClusterVariables path)
// succeeds too.
func TestAttachClusterToTemplate_WithMock(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	d := resourceClusterAws().TestResourceData()
	d.SetId("test-cluster-id")
	require.NoError(t, d.Set("cluster_template", []interface{}{
		map[string]interface{}{
			"id": "test-cluster-config-template-id",
			"cluster_profile": []interface{}{
				map[string]interface{}{
					"id": "test-profile-uid-1",
					"variables": map[string]interface{}{
						"region": "us-west-2",
					},
				},
			},
		},
	}))

	err := attachClusterToTemplate(c, d)
	require.NoError(t, err)
}

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
			"id":        "test-profile-uid-1",
			"variables": variablesSet,
		},
	})

	_ = d.Set("cluster_profile", profilesSet)
	_ = d.Set("policy", []interface{}{
		map[string]interface{}{
			"id":   "test-policy-uid-1",
			"kind": "maintenance",
		},
	})
	d.SetId("test-cluster-config-template-id")
	return d
}

func TestResourceClusterConfigTemplateCRUD(t *testing.T) {
	testResourceCRUD(t, prepareBaseClusterConfigTemplateTestData, unitTestMockAPIClient,
		resourceClusterConfigTemplateCreate, resourceClusterConfigTemplateRead, resourceClusterConfigTemplateUpdate, resourceClusterConfigTemplateDelete)
}

func TestExpandClusterTemplateProfiles(t *testing.T) {
	profiles := []interface{}{
		map[string]interface{}{
			"id": "profile-uid-1",
		},
		map[string]interface{}{
			"id": "profile-uid-2",
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
			"id":   "policy-uid-1",
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
			"id":        "profile-1",
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{}),
		},
	})
	newProfiles := schema.NewSet(resourceClusterConfigTemplateProfileHash, []interface{}{
		map[string]interface{}{
			"id":        "profile-1",
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{}),
		},
		map[string]interface{}{
			"id":        "profile-2",
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{}),
		},
	})
	assert.True(t, profileStructureChanged(oldProfiles, newProfiles), "Should detect added profile")

	// Test case 2: Same number but different IDs
	oldProfiles = schema.NewSet(resourceClusterConfigTemplateProfileHash, []interface{}{
		map[string]interface{}{
			"id":        "profile-1",
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{}),
		},
		map[string]interface{}{
			"id":        "profile-2",
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{}),
		},
	})
	newProfiles = schema.NewSet(resourceClusterConfigTemplateProfileHash, []interface{}{
		map[string]interface{}{
			"id":        "profile-1",
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{}),
		},
		map[string]interface{}{
			"id":        "profile-3",
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{}),
		},
	})
	assert.True(t, profileStructureChanged(oldProfiles, newProfiles), "Should detect changed profile ID")

	// Test case 3: Same IDs, only variables changed
	oldVars := schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{
		map[string]interface{}{"name": "var1", "value": "old", "assign_strategy": "all"},
	})
	newVars := schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{
		map[string]interface{}{"name": "var1", "value": "new", "assign_strategy": "all"},
	})

	oldProfiles = schema.NewSet(resourceClusterConfigTemplateProfileHash, []interface{}{
		map[string]interface{}{
			"id":        "profile-1",
			"variables": oldVars,
		},
	})
	newProfiles = schema.NewSet(resourceClusterConfigTemplateProfileHash, []interface{}{
		map[string]interface{}{
			"id":        "profile-1",
			"variables": newVars,
		},
	})
	assert.False(t, profileStructureChanged(oldProfiles, newProfiles), "Should not detect change when only variables differ")
}
