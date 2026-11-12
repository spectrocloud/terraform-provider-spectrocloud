package spectrocloud

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clusterProfileIDs(d *schema.ResourceData) []string {
	ids := []string{}
	for _, p := range normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile")) {
		ids = append(ids, p.(map[string]interface{})["id"].(string))
	}
	return ids
}

func isClusterProfilePatched(d *schema.ResourceData) bool {
	_, ok := patchedClusterProfiles.Load(d)
	return ok
}

func TestWithClusterProfileRollbackOnError(t *testing.T) {
	disableAddonDeploymentResource = false
	meta := &ProviderMeta{}
	oldProfiles := []interface{}{map[string]interface{}{"id": "profile-old"}}
	newProfiles := []interface{}{map[string]interface{}{"id": "profile-new"}}

	t.Run("failure before the profile PATCH restores the pre-apply profile", func(t *testing.T) {
		d := buildClusterProfileChangeResourceData(t, oldProfiles, newProfiles, nil)
		update := withClusterProfileRollbackOnError(func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics {
			return diag.Errorf("machine pool update failed")
		})

		diags := update(context.Background(), d, meta)

		require.True(t, diags.HasError())
		assert.Equal(t, []string{"profile-old"}, clusterProfileIDs(d))
	})

	t.Run("failure after the profile PATCH keeps the new profile", func(t *testing.T) {
		d := buildClusterProfileChangeResourceData(t, oldProfiles, newProfiles, nil)
		update := withClusterProfileRollbackOnError(func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
			markClusterProfilePatched(d)
			return diag.Errorf("backup policy update failed")
		})

		diags := update(context.Background(), d, meta)

		require.True(t, diags.HasError())
		assert.Equal(t, []string{"profile-new"}, clusterProfileIDs(d))
		assert.False(t, isClusterProfilePatched(d), "the marker must be cleared once the update returns")
	})

	t.Run("successful update keeps the new profile", func(t *testing.T) {
		d := buildClusterProfileChangeResourceData(t, oldProfiles, newProfiles, nil)
		update := withClusterProfileRollbackOnError(func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics {
			return nil
		})

		require.False(t, update(context.Background(), d, meta).HasError())
		assert.Equal(t, []string{"profile-new"}, clusterProfileIDs(d))
	})

	t.Run("failure without a cluster_profile change leaves it alone", func(t *testing.T) {
		d := buildClusterProfileChangeResourceData(t, oldProfiles, oldProfiles, map[string]interface{}{"description": "x"})
		require.False(t, d.HasChange("cluster_profile"))
		update := withClusterProfileRollbackOnError(func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
			_ = d.Set("cluster_profile", newProfiles)
			return diag.FromErr(errors.New("unrelated failure"))
		})

		update(context.Background(), d, meta)

		assert.Equal(t, []string{"profile-new"}, clusterProfileIDs(d))
	})
}

func TestUpdateProfilesPatchRejectedIsNotMarkedPatched(t *testing.T) {
	disableAddonDeploymentResource = false
	c := castV1Client(t, unitTestMockAPIClient)
	d := buildClusterProfileChangeResourceData(t,
		[]interface{}{map[string]interface{}{"id": "k8s-1-35-profile-uid"}},
		[]interface{}{map[string]interface{}{"id": "k8s-1-37-profile-uid"}},
		nil,
	)
	d.SetId(clusterProfilesPatchErrorUID)
	t.Cleanup(func() { patchedClusterProfiles.Delete(d) })

	require.Error(t, updateProfiles(c, d))
	assert.False(t, isClusterProfilePatched(d))
	assert.Equal(t, []string{"k8s-1-35-profile-uid"}, clusterProfileIDs(d))
}

func TestUpdateProfilesVariablesErrorRestoresOldVariables(t *testing.T) {
	disableAddonDeploymentResource = false
	c := castV1Client(t, unitTestMockAPIClient)
	d := buildClusterProfileChangeResourceData(t,
		[]interface{}{map[string]interface{}{
			"id":        "variable-profile-uid",
			"variables": map[string]interface{}{"region": "us-east-1"},
		}},
		[]interface{}{map[string]interface{}{
			"id":        "variable-profile-uid",
			"variables": map[string]interface{}{"region": "us-west-2"},
		}},
		nil,
	)
	d.SetId(clusterVariablesPatchErrorUID)
	t.Cleanup(func() { patchedClusterProfiles.Delete(d) })

	require.Error(t, updateProfiles(c, d))

	assert.True(t, isClusterProfilePatched(d), "the profile PATCH succeeded before the variables PATCH failed")
	profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
	require.Len(t, profiles, 1)
	assert.Equal(t, map[string]interface{}{"region": "us-east-1"}, profiles[0].(map[string]interface{})["variables"])
}

func TestRestoreClusterProfileVariables(t *testing.T) {
	d := buildClusterProfileChangeResourceData(t,
		[]interface{}{map[string]interface{}{"id": "profile-old", "variables": map[string]interface{}{"a": "1"}}},
		[]interface{}{
			map[string]interface{}{"id": "profile-old", "variables": map[string]interface{}{"a": "2"}},
			map[string]interface{}{"id": "profile-added", "variables": map[string]interface{}{"b": "1"}},
		},
		nil,
	)
	oldRaw, _ := d.GetChange("cluster_profile")

	restoreClusterProfileVariables(d, normalizeInterfaceSliceFromListOrSet(oldRaw))

	profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
	require.Len(t, profiles, 2)
	vars := map[string]interface{}{}
	for _, p := range profiles {
		m := p.(map[string]interface{})
		vars[m["id"].(string)] = m["variables"]
	}
	assert.Equal(t, map[string]interface{}{"a": "1"}, vars["profile-old"])
	assert.Empty(t, vars["profile-added"], "a newly attached profile had no variables applied")
}

func TestSyncClusterProfileAfterPatchErrorKeepsNewProfileWhenFlagOff(t *testing.T) {
	disableAddonDeploymentResource = false
	d := buildClusterProfileChangeResourceData(t,
		[]interface{}{map[string]interface{}{"id": "profile-old"}},
		[]interface{}{map[string]interface{}{"id": "profile-new"}},
		nil,
	)

	syncClusterProfileAfterPatchError(castV1Client(t, unitTestMockAPIClient), d)

	assert.Equal(t, []string{"profile-new"}, clusterProfileIDs(d))
}

func TestSetClusterProfileNamesFromCluster(t *testing.T) {
	d := resourceClusterEks().TestResourceData()
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": "profile-a"},
		map[string]interface{}{"id": "profile-not-attached"},
	}))
	cluster := &models.V1SpectroCluster{Spec: &models.V1SpectroClusterSpec{
		ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
			{UID: "profile-a", Name: "infra-profile"},
			{UID: "addon-profile", Name: "attached-by-addon-deployment"},
		},
	}}

	require.NoError(t, setClusterProfileNamesFromCluster(d, cluster))

	names := map[string]interface{}{}
	for _, p := range normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile")) {
		m := p.(map[string]interface{})
		names[m["id"].(string)] = m["name"]
	}
	assert.Equal(t, map[string]interface{}{
		"profile-a":            "infra-profile",
		"profile-not-attached": "",
	}, names, "only existing entries get a name; attached profiles not in state must not be added")
}
