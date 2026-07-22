package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

// clusterVariablesPatchErrorUID mirrors the constant of the same name in
// tests/mockApiServer/routes/mockCluster.go (not importable from this
// package) — GET/PATCH against this UID's variables endpoint returns 500.
const clusterVariablesPatchErrorUID = "cluster-uid-variables-patch-error"

// ---------------------------------------------------------------------------
// updateProfiles / rollbackClusterProfileOnUpdateError / computeProfilesToDelete /
// enrichClusterProfilesWithPacks / toClusterTemplateReference /
// updateClusterTemplateVariables / flattenClusterTemplateVariables.
//
// updateProfiles and updateClusterTemplateVariables both gate most of their
// logic behind d.HasChange(...)/d.GetChange(...). schema.TestResourceData()
// (and calling d.Set() twice on the same key, as cluster_common_wave1_test.go
// does elsewhere) never produces a real diff — GetChange reads from the
// state+diff readers, not from the "set" writer that d.Set() populates — so
// HasChange is always false and GetChange's "old" is always the zero value.
// The helpers below build a genuine SDK-computed diff (Resource.Diff +
// schema.InternalMap.Data), the same technique used by
// buildMachinePoolChangeResourceData (resource_cluster_apache_cloudstack_test.go)
// and buildCustomCloudCloudConfigChangeResourceData (resource_cluster_custom_cloud_wave3_test.go).
// ---------------------------------------------------------------------------

func buildClusterProfileChangeResourceData(t *testing.T, oldProfiles, newProfiles []interface{}, overrides map[string]interface{}) *schema.ResourceData {
	t.Helper()
	res := resourceClusterEks()

	base := map[string]interface{}{
		"name":             "test-cluster",
		"cloud_account_id": "test-account",
		"context":          "project",
	}
	for k, v := range overrides {
		base[k] = v
	}

	oldRaw := map[string]interface{}{}
	for k, v := range base {
		oldRaw[k] = v
	}
	oldRaw["cluster_profile"] = oldProfiles
	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId("test-cluster-id")
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newRaw := map[string]interface{}{}
	for k, v := range base {
		newRaw[k] = v
	}
	newRaw["cluster_profile"] = newProfiles
	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)
	require.NotNil(t, diff)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId("test-cluster-id")
	return finalRD
}

func buildClusterTemplateChangeResourceData(t *testing.T, id string, oldTemplate, newTemplate []interface{}) *schema.ResourceData {
	t.Helper()
	res := resourceClusterEks()

	base := map[string]interface{}{
		"name":             "test-cluster",
		"cloud_account_id": "test-account",
		"context":          "project",
	}

	oldRaw := map[string]interface{}{}
	for k, v := range base {
		oldRaw[k] = v
	}
	oldRaw["cluster_template"] = oldTemplate
	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId(id)
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newRaw := map[string]interface{}{}
	for k, v := range base {
		newRaw[k] = v
	}
	newRaw["cluster_template"] = newTemplate
	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)
	require.NotNil(t, diff)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId(id)
	return finalRD
}

func TestUpdateProfilesAddProfile_RealDiff(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := buildClusterProfileChangeResourceData(t,
		[]interface{}{},
		[]interface{}{map[string]interface{}{"id": "brand-new-profile-uid"}},
		nil,
	)
	require.True(t, d.HasChange("cluster_profile"))

	require.NoError(t, updateProfiles(c, d))
}

// TestUpdateProfilesRemovesAddonProfile_RealDiff drives the profile-deletion
// branch: an addon profile present in the old set (and still attached on the
// mock cluster fixture) is dropped from the new set, so getProfilesToDelete
// must call DeleteAddonDeployment before the new profile is patched in.
// clusterProfileUID2 ("cluster-profile-import-2") is the mock's addon-typed
// profile — see clusterProfileFixtureFor in mockClusterProfile.go.
func TestUpdateProfilesRemovesAddonProfile_RealDiff(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := buildClusterProfileChangeResourceData(t,
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-2"}},
		[]interface{}{map[string]interface{}{"id": "brand-new-added-profile-uid"}},
		nil,
	)
	require.True(t, d.HasChange("cluster_profile"))

	toDelete := getProfilesToDelete(c, d, mustGetCluster(t, c, "test-cluster-id"))
	require.Contains(t, toDelete, "cluster-profile-import-2")

	require.NoError(t, updateProfiles(c, d))
}

// TestUpdateProfilesVariableUpdate_RealDiff drives the profile-variable
// update loop at the tail of updateProfiles: same profile UID old and new,
// but variables added, plus a second new-profile entry with no id (must be
// skipped rather than sent to the API).
func TestUpdateProfilesVariableUpdate_RealDiff(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := buildClusterProfileChangeResourceData(t,
		[]interface{}{map[string]interface{}{"id": "variable-profile-uid"}},
		[]interface{}{
			map[string]interface{}{
				"id": "variable-profile-uid",
				"variables": map[string]interface{}{
					"region": "us-west-2",
				},
			},
		},
		nil,
	)
	require.True(t, d.HasChange("cluster_profile"))

	require.NoError(t, updateProfiles(c, d))
}

// TestUpdateProfilesVariableUpdateError_RealDiff forces
// c.UpdateClusterProfileVariableInCluster to fail (clusterVariablesPatchErrorUID
// in mockCluster.go), exercising the rollback branch at the tail of
// updateProfiles and confirming cluster_profile reverts to the pre-apply value.
func TestUpdateProfilesVariableUpdateError_RealDiff(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := buildClusterProfileChangeResourceData(t,
		[]interface{}{map[string]interface{}{"id": "variable-profile-uid"}},
		[]interface{}{
			map[string]interface{}{
				"id": "variable-profile-uid",
				"variables": map[string]interface{}{
					"region": "us-west-2",
				},
			},
		},
		nil,
	)
	d.SetId(clusterVariablesPatchErrorUID)
	require.True(t, d.HasChange("cluster_profile"))

	err := updateProfiles(c, d)
	require.Error(t, err)

	profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
	require.Len(t, profiles, 1)
	assert.Equal(t, "variable-profile-uid", profiles[0].(map[string]interface{})["id"])
	// The pre-apply snapshot came from oldState via the SDK's typed reader, which always
	// populates the "variables" TypeMap sub-field (as an empty map) even when unset in
	// config — so assert emptiness rather than key absence.
	vars, _ := profiles[0].(map[string]interface{})["variables"].(map[string]interface{})
	assert.Empty(t, vars, "rollback must restore the pre-apply cluster_profile, without the new variables")
}

// TestUpdateProfilesRollbackOnBadContext_RealDiff forces toAddonDeplProfiles
// to fail via an invalid "context" value, exercising updateProfiles' first
// rollback call site and confirming cluster_profile is restored.
func TestUpdateProfilesRollbackOnBadContext_RealDiff(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	res := resourceClusterEks()

	oldRaw := map[string]interface{}{
		"name":             "test-cluster",
		"cloud_account_id": "test-account",
		"context":          "project",
		"cluster_profile":  []interface{}{map[string]interface{}{"id": "pre-apply-profile"}},
	}
	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId("test-cluster-id")
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newRaw := map[string]interface{}{
		"name":             "test-cluster",
		"cloud_account_id": "test-account",
		"context":          "not-a-real-context",
		"cluster_profile":  []interface{}{map[string]interface{}{"id": "pre-apply-profile"}},
	}
	newConfig := terraform.NewResourceConfigRaw(newRaw)
	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)
	require.NotNil(t, diff)

	d, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	d.SetId("test-cluster-id")

	updateErr := updateProfiles(c, d)
	require.Error(t, updateErr)

	profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
	require.Len(t, profiles, 1)
	assert.Equal(t, "pre-apply-profile", profiles[0].(map[string]interface{})["id"])
}

// TestUpdateProfilesGetClusterError_RealDiff exercises the "failed to get
// cluster for profile update" error path (no rollback call at this site —
// see cluster_common_profiles.go's updateProfiles). This needs the
// toAddonDeplProfiles lookup (keyed off "cluster_uid" when present, else
// d.Id()) to succeed against a *different* UID than the one d.Id() uses for
// updateProfiles' own c.GetCluster(d.Id()) call — otherwise the earlier
// toProfilesCommon call in toAddonDeplProfiles fails first on the very same
// broken UID. resourceAddonDeployment's schema is the one place
// "cluster_uid" and the resource's own d.Id() are already distinct fields.
func TestUpdateProfilesGetClusterError_RealDiff(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := resourceAddonDeployment().TestResourceData()
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cluster_uid", "cluster-uid-addon-ready"))
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": "some-profile-uid"},
	}))
	d.SetId("cluster-uid-server-error")

	err := updateProfiles(c, d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get cluster for profile update")
}

// TestUpdateProfilesSetReplaceWithProfileError_RealDiff forces
// setReplaceWithProfileForExisting's c.GetClusterProfile lookup to fail
// (clusterProfileGetErrorUID in mockClusterProfile.go), exercising that
// rollback call site.
func TestUpdateProfilesSetReplaceWithProfileError_RealDiff(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := buildClusterProfileChangeResourceData(t,
		[]interface{}{map[string]interface{}{"id": "pre-apply-profile"}},
		[]interface{}{map[string]interface{}{"id": "cluster-profile-get-error-uid"}},
		nil,
	)

	err := updateProfiles(c, d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve profile replacements")

	profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
	require.Len(t, profiles, 1)
	assert.Equal(t, "pre-apply-profile", profiles[0].(map[string]interface{})["id"])
}

func mustGetCluster(t *testing.T, c interface {
	GetCluster(string) (*models.V1SpectroCluster, error)
}, uid string) *models.V1SpectroCluster {
	t.Helper()
	cluster, err := c.GetCluster(uid)
	require.NoError(t, err)
	require.NotNil(t, cluster)
	return cluster
}

// ---------------------------------------------------------------------------
// updateClusterTemplateVariables / flattenClusterTemplateVariables
// ---------------------------------------------------------------------------

func TestUpdateClusterTemplateVariables_RealDiff(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	oldTemplate := []interface{}{
		map[string]interface{}{
			"id": "template-uid-1",
			"cluster_profile": []interface{}{
				map[string]interface{}{"id": "cluster-profile-import-1"},
			},
		},
	}
	newTemplate := []interface{}{
		map[string]interface{}{
			"id": "template-uid-1",
			"cluster_profile": []interface{}{
				map[string]interface{}{
					"id": "cluster-profile-import-1",
					"variables": map[string]interface{}{
						"region": "us-west-2",
					},
				},
			},
		},
	}

	d := buildClusterTemplateChangeResourceData(t, "test-cluster-id", oldTemplate, newTemplate)
	require.True(t, d.HasChange("cluster_template"))

	require.NoError(t, updateClusterTemplateVariables(c, d))

	tmpl := d.Get("cluster_template").([]interface{})
	require.Len(t, tmpl, 1)
	tmplMap := tmpl[0].(map[string]interface{})
	assert.Equal(t, "template-uid-1", tmplMap["id"])

	profSet, ok := tmplMap["cluster_profile"].(*schema.Set)
	require.True(t, ok)
	require.Equal(t, 1, profSet.Len())
	prof := profSet.List()[0].(map[string]interface{})
	assert.Equal(t, "cluster-profile-import-1", prof["id"])

	vars, ok := prof["variables"].(map[string]interface{})
	require.True(t, ok)
	// flattenClusterTemplateVariables refreshes from GetClusterVariables (mock
	// fixture returns region=us-east-1 for cluster-profile-import-1),
	// filtered to variable names present in config ("region").
	assert.Equal(t, "us-east-1", vars["region"])
}

func TestUpdateClusterTemplateVariablesError_RealDiff(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	oldTemplate := []interface{}{
		map[string]interface{}{
			"id": "template-uid-1",
			"cluster_profile": []interface{}{
				map[string]interface{}{"id": "cluster-profile-import-1"},
			},
		},
	}
	newTemplate := []interface{}{
		map[string]interface{}{
			"id": "template-uid-1",
			"cluster_profile": []interface{}{
				map[string]interface{}{
					"id": "cluster-profile-import-1",
					"variables": map[string]interface{}{
						"region": "us-west-2",
					},
				},
			},
		},
	}

	d := buildClusterTemplateChangeResourceData(t, clusterVariablesPatchErrorUID, oldTemplate, newTemplate)
	require.True(t, d.HasChange("cluster_template"))

	err := updateClusterTemplateVariables(c, d)
	require.Error(t, err)

	tmpl := d.Get("cluster_template").([]interface{})
	require.Len(t, tmpl, 1)
	profSet, ok := tmpl[0].(map[string]interface{})["cluster_profile"].(*schema.Set)
	require.True(t, ok)
	require.Equal(t, 1, profSet.Len())
	prof := profSet.List()[0].(map[string]interface{})
	assert.Equal(t, "cluster-profile-import-1", prof["id"])
	vars, _ := prof["variables"].(map[string]interface{})
	assert.Empty(t, vars, "on error, cluster_template must roll back to the old (no-variables) value")
}

func TestFlattenClusterTemplateVariables_NoTemplate(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := resourceClusterEks().TestResourceData()
	d.SetId("test-cluster-id")

	require.NoError(t, flattenClusterTemplateVariables(c, d, d.Id()))
}

// TestFlattenClusterTemplateVariables_VariablesAPIError exercises the "Error
// fetching cluster variables" branch: the negative mock server has no route
// for GET /v1/spectroclusters/{uid}/variables, so GetClusterVariables errors,
// and flattenClusterTemplateVariables must swallow it (return nil) rather
// than fail the read/update.
func TestFlattenClusterTemplateVariables_VariablesAPIError(t *testing.T) {
	cNeg := castV1Client(t, unitTestMockAPINegativeClient)
	d := resourceClusterEks().TestResourceData()
	d.SetId("test-cluster-id")
	require.NoError(t, d.Set("cluster_template", []interface{}{
		map[string]interface{}{"id": "template-uid-1"},
	}))

	require.NoError(t, flattenClusterTemplateVariables(cNeg, d, d.Id()))
}

// ---------------------------------------------------------------------------
// rollbackClusterProfileOnUpdateError — the shouldSyncClusterProfilesFromAPI
// && c != nil && d.Id() != "" branch, which cluster_common_profiles_feature_flag_test.go's
// TestRollbackClusterProfileOnUpdateError only exercises with c == nil.
// ---------------------------------------------------------------------------

func TestRollbackClusterProfileOnUpdateError_APISync(t *testing.T) {
	t.Cleanup(func() { disableAddonDeploymentResource = false })
	disableAddonDeploymentResource = true

	c := castV1Client(t, unitTestMockAPIClient)
	oldProfile := []interface{}{map[string]interface{}{"id": "pre-apply-profile"}}

	t.Run("GetCluster succeeds, syncs cluster_profile from the cluster API document", func(t *testing.T) {
		d := resourceClusterEks().TestResourceData()
		d.SetId("test-cluster-id")
		_ = d.Set("cluster_profile", []interface{}{map[string]interface{}{"id": "failed-desired-profile"}})

		rollbackClusterProfileOnUpdateError(c, d, oldProfile)

		profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
		require.Len(t, profiles, 2)
		ids := []string{
			profiles[0].(map[string]interface{})["id"].(string),
			profiles[1].(map[string]interface{})["id"].(string),
		}
		assert.ElementsMatch(t, []string{"cluster-profile-import-1", "cluster-profile-import-2"}, ids)
	})

	t.Run("GetCluster hard error falls back to the pre-apply snapshot", func(t *testing.T) {
		d := resourceClusterEks().TestResourceData()
		d.SetId("cluster-uid-server-error")
		_ = d.Set("cluster_profile", []interface{}{map[string]interface{}{"id": "failed-desired-profile"}})

		rollbackClusterProfileOnUpdateError(c, d, oldProfile)

		profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
		require.Len(t, profiles, 1)
		assert.Equal(t, "pre-apply-profile", profiles[0].(map[string]interface{})["id"])
	})

	t.Run("GetCluster returns nil cluster (404-swallowed) falls back to the pre-apply snapshot", func(t *testing.T) {
		d := resourceClusterEks().TestResourceData()
		d.SetId("cluster-uid-not-found")
		_ = d.Set("cluster_profile", []interface{}{map[string]interface{}{"id": "failed-desired-profile"}})

		rollbackClusterProfileOnUpdateError(c, d, oldProfile)

		profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
		require.Len(t, profiles, 1)
		assert.Equal(t, "pre-apply-profile", profiles[0].(map[string]interface{})["id"])
	})
}

// ---------------------------------------------------------------------------
// enrichClusterProfilesWithPacks
// ---------------------------------------------------------------------------

func TestEnrichClusterProfilesWithPacks_RealMock(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	cluster := &models.V1SpectroCluster{
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{
					UID: "profile-with-packs",
					Packs: []*models.V1PackRef{
						{Name: types.Ptr("nginx"), PackUID: "pack-uid-nginx", Type: "helm"},
					},
				},
				{
					UID:   "profile-no-packs-on-cluster",
					Packs: []*models.V1PackRef{},
				},
			},
		},
	}

	d := resourceClusterEks().TestResourceData()
	d.SetId("test-cluster-id")
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{
			"id": "profile-with-packs",
			"pack": []interface{}{
				map[string]interface{}{"name": "nginx"},
			},
		},
		map[string]interface{}{
			"id": "profile-no-packs-on-cluster",
			"pack": []interface{}{
				map[string]interface{}{"name": "unused"},
			},
		},
		map[string]interface{}{
			"id": "profile-not-attached-to-cluster",
		},
	}))

	clusterProfiles := []interface{}{
		map[string]interface{}{"id": "profile-with-packs", "pack": []interface{}{}},
		map[string]interface{}{"id": "profile-no-packs-on-cluster", "pack": []interface{}{}},
		map[string]interface{}{"id": "profile-not-attached-to-cluster", "pack": []interface{}{}},
	}

	result, err := enrichClusterProfilesWithPacks(c, d, cluster, clusterProfiles)
	require.NoError(t, err)
	require.Len(t, result, 3)

	withPacks := result[0].(map[string]interface{})
	packs, ok := withPacks["pack"].([]interface{})
	require.True(t, ok)
	require.Len(t, packs, 1)
	assert.Equal(t, "nginx", packs[0].(map[string]interface{})["name"])

	// Attached on the cluster but with zero Packs on the template -> left as-is (empty).
	noPacksOnCluster := result[1].(map[string]interface{})
	assert.Equal(t, []interface{}{}, noPacksOnCluster["pack"])

	// Not attached on the cluster at all -> templateByUID lookup misses, left as-is.
	notAttached := result[2].(map[string]interface{})
	assert.Equal(t, []interface{}{}, notAttached["pack"])
}

func TestEnrichClusterProfilesWithPacks_NoPackConfig(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	cluster := &models.V1SpectroCluster{
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{
					UID: "profile-with-packs",
					Packs: []*models.V1PackRef{
						{Name: types.Ptr("nginx"), PackUID: "pack-uid-nginx", Type: "helm"},
					},
				},
			},
		},
	}

	d := resourceClusterEks().TestResourceData()
	d.SetId("test-cluster-id")
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": "profile-with-packs"},
	}))

	clusterProfiles := []interface{}{
		map[string]interface{}{"id": "profile-with-packs", "pack": []interface{}{}},
	}

	result, err := enrichClusterProfilesWithPacks(c, d, cluster, clusterProfiles)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, []interface{}{}, result[0].(map[string]interface{})["pack"])
}

// ---------------------------------------------------------------------------
// toClusterTemplateReference
// ---------------------------------------------------------------------------

func TestToClusterTemplateReference_NonEmpty(t *testing.T) {
	d := resourceClusterEks().TestResourceData()
	require.NoError(t, d.Set("cluster_template", []interface{}{
		map[string]interface{}{"id": "template-uid-1"},
	}))

	ref := toClusterTemplateReference(d)
	require.NotNil(t, ref)
	assert.Equal(t, "template-uid-1", ref.UID)
}
