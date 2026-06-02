package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustUnitClient(t *testing.T, negative bool) *client.V1Client {
	t.Helper()

	var raw interface{}
	if negative {
		raw = unitTestMockAPINegativeClient
	} else {
		raw = unitTestMockAPIClient
	}

	c, ok := raw.(*client.V1Client)
	require.True(t, ok, "expected mock client to be *client.V1Client")
	require.NotNil(t, c)
	return c
}

func setChangedClusterProfiles(t *testing.T, d *schema.ResourceData, oldProfiles, newProfiles []interface{}) {
	t.Helper()
	require.NoError(t, d.Set("cluster_profile", oldProfiles))
	require.NoError(t, d.Set("cluster_profile", newProfiles))
}

func TestUpdateProfilesSkipApply(t *testing.T) {
	c := mustUnitClient(t, false)
	d := resourceClusterEks().TestResourceData()
	d.SetId("test-cluster-id")

	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("tags", []interface{}{"skip_apply"}))

	setChangedClusterProfiles(t, d,
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-2"}},
		[]interface{}{
			map[string]interface{}{
				"id": "cluster-profile-import-1",
				"variables": map[string]interface{}{
					"region": "us-east-1",
				},
			},
		},
	)

	require.NoError(t, updateProfiles(c, d))
}

func TestUpdateCommonFieldsClusterProfilePath(t *testing.T) {
	d := resourceClusterEks().TestResourceData()
	d.SetId("test-cluster-id")
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("tags", []interface{}{"skip_apply"}))

	setChangedClusterProfiles(t, d,
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-2"}},
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-1"}},
	)

	diags, done := updateCommonFields(d, mustUnitClient(t, false))
	assert.False(t, done)
	assert.Empty(t, diags)
}

func TestUpdateProfilesMissingClusterID(t *testing.T) {
	c := mustUnitClient(t, false)
	d := resourceClusterEks().TestResourceData()
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": "cluster-profile-import-1"},
	}))

	err := updateProfiles(c, d)
	assert.Error(t, err)
}

func TestUpdateProfilesDeletesRemovedAddonProfile(t *testing.T) {
	c := mustUnitClient(t, false)
	d := resourceClusterEks().TestResourceData()
	d.SetId("test-cluster-id")
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("tags", []interface{}{"skip_apply"}))

	setChangedClusterProfiles(t, d,
		[]interface{}{
			map[string]interface{}{"id": "cluster-profile-import-1"},
			map[string]interface{}{"id": "cluster-profile-import-2"},
		},
		[]interface{}{
			map[string]interface{}{"id": "cluster-profile-import-1"},
		},
	)

	require.NoError(t, updateProfiles(c, d))
}

func TestEnrichClusterProfilesWithVariablesWithMock(t *testing.T) {
	c := mustUnitClient(t, false)
	d := resourceClusterEks().TestResourceData()
	d.SetId("test-cluster-id")

	profiles := []interface{}{
		map[string]interface{}{"id": "cluster-profile-import-1"},
	}

	enriched, err := enrichClusterProfilesWithVariables(c, d, d.Id(), profiles)
	require.NoError(t, err)
	require.Len(t, enriched, 1)
	vars := enriched[0].(map[string]interface{})["variables"].(map[string]interface{})
	assert.Equal(t, "us-east-1", vars["region"])
}

func TestUpdateClusterTemplateVariablesWithMock(t *testing.T) {
	c := mustUnitClient(t, false)
	d := resourceClusterEks().TestResourceData()
	d.SetId("test-cluster-id")

	oldTemplate := []interface{}{
		map[string]interface{}{
			"id": "cluster-template-1",
			"cluster_profile": schema.NewSet(schema.HashResource(&schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {Type: schema.TypeString, Required: true},
				},
			}), []interface{}{
				map[string]interface{}{"id": "cluster-profile-import-1"},
			}),
		},
	}
	newTemplate := []interface{}{
		map[string]interface{}{
			"id": "cluster-template-1",
			"cluster_profile": schema.NewSet(schema.HashResource(&schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {Type: schema.TypeString, Required: true},
					"variables": {
						Type: schema.TypeMap,
						Elem: &schema.Schema{Type: schema.TypeString},
					},
				},
			}), []interface{}{
				map[string]interface{}{
					"id": "cluster-profile-import-1",
					"variables": map[string]interface{}{
						"region": "us-west-2",
					},
				},
			}),
		},
	}

	require.NoError(t, d.Set("cluster_template", oldTemplate))
	require.NoError(t, d.Set("cluster_template", newTemplate))

	require.NoError(t, updateClusterTemplateVariables(c, d))
}

func TestSetReplaceWithProfileForExistingWithMock(t *testing.T) {
	c := mustUnitClient(t, false)
	cluster := &models.V1SpectroCluster{
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{UID: "cluster-profile-import-1", Name: "test-cluster-profile-1", Type: "cluster"},
			},
		},
	}
	profiles := []*models.V1SpectroClusterProfileEntity{
		{UID: "cluster-profile-import-1"},
	}

	require.NoError(t, setReplaceWithProfileForExisting(c, cluster, profiles))
	assert.Empty(t, profiles[0].ReplaceWithProfile)
}

func TestReadCommonFieldsSyncProfilesWhenAddonDisabled(t *testing.T) {
	t.Cleanup(func() {
		disableAddonDeploymentResource = false
	})
	disableAddonDeploymentResource = true

	cluster := prepareSpectroClusterModel()
	cluster.Spec.ClusterProfileTemplates = []*models.V1ClusterProfileTemplate{
		{UID: "cluster-profile-import-1"},
	}

	d := resourceClusterEks().TestResourceData()
	d.SetId("test-cluster-id")

	diags, hasError := readCommonFields(mustUnitClient(t, false), d, cluster)
	assert.False(t, hasError)
	assert.Empty(t, diags)

	profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
	require.Len(t, profiles, 1)
	assert.Equal(t, "cluster-profile-import-1", profiles[0].(map[string]interface{})["id"])
}

func TestReadCommonFieldsPaths(t *testing.T) {
	cluster := prepareSpectroClusterModel()
	cluster.Spec.ClusterConfig.Timezone = "UTC"
	cluster.Spec.ClusterConfig.UpdateWorkerPoolsInParallel = true
	cluster.Spec.ClusterConfig.ClusterMetaAttribute = "test-meta"
	cluster.Spec.ClusterConfig.HostClusterConfig.IsHostCluster = BoolPtr(true)

	t.Run("success path", func(t *testing.T) {
		d := resourceClusterEks().TestResourceData()
		d.SetId("test-cluster-id")
		require.NoError(t, d.Set("cluster_meta_attribute", "placeholder"))
		require.NoError(t, d.Set("cluster_timezone", "UTC"))

		diags, hasError := readCommonFields(mustUnitClient(t, false), d, cluster)
		assert.False(t, hasError)
		assert.Empty(t, diags)
		assert.Equal(t, "test-meta", d.Get("cluster_meta_attribute"))
		assert.Equal(t, "UTC", d.Get("cluster_timezone"))
		assert.Equal(t, true, d.Get("update_worker_pools_in_parallel"))
	})
}

