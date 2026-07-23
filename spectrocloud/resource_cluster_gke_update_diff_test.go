package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

// TestResourceClusterGkeStateUpgradeV2 exercises the state-upgrade branches
// directly: cluster_profile present as a list is kept as-is, a non-list
// value is left alone, and a missing key is a no-op — all three log-only
// branches return rawState unchanged, so the assertions confirm we hit each
// branch without altering the map.
func TestResourceClusterGkeStateUpgradeV2(t *testing.T) {
	t.Run("cluster_profile present as list", func(t *testing.T) {
		raw := map[string]interface{}{
			"cluster_profile": []interface{}{
				map[string]interface{}{"id": "profile-1"},
			},
		}
		out, err := resourceClusterGkeStateUpgradeV2(context.Background(), raw, nil)
		require.NoError(t, err)
		list, ok := out["cluster_profile"].([]interface{})
		require.True(t, ok)
		assert.Len(t, list, 1)
	})

	t.Run("cluster_profile present but not a list", func(t *testing.T) {
		raw := map[string]interface{}{
			"cluster_profile": "not-a-list",
		}
		out, err := resourceClusterGkeStateUpgradeV2(context.Background(), raw, nil)
		require.NoError(t, err)
		assert.Equal(t, "not-a-list", out["cluster_profile"])
	})

	t.Run("cluster_profile absent", func(t *testing.T) {
		raw := map[string]interface{}{
			"name": "some-cluster",
		}
		out, err := resourceClusterGkeStateUpgradeV2(context.Background(), raw, nil)
		require.NoError(t, err)
		_, exists := out["cluster_profile"]
		assert.False(t, exists)
		assert.Equal(t, "some-cluster", out["name"])
	})
}

// resourceClusterGkeUpdate's machine_pool branch is gated behind
// d.HasChange("machine_pool"), which the Set-then-Set pattern used by
// resource_cluster_gke_wave2_test.go never fires (see
// resource_cluster_edge_native_update_diff_test.go's
// buildEdgeNativeMachinePoolChangeResourceData for the rationale).
// buildGkeUpdateResourceData builds a real InstanceState + config diff via
// Resource.Diff so HasChange/GetChange behave the way Terraform's own apply
// pipeline would produce them.
func buildGkeUpdateResourceData(t *testing.T, oldRaw, newRaw map[string]interface{}, configUID string) *schema.ResourceData {
	t.Helper()
	res := resourceClusterGke()

	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId(gkeClusterID)
	require.NoError(t, oldRD.Set("cloud_config_id", configUID))
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId(gkeClusterID)
	return finalRD
}

const gkeClusterID = "test-gke-cluster-id"

func baseGkeUpdateRaw(machinePools []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":             "test-gke-cluster",
		"context":          "project",
		"cloud_account_id": gkeCloudAccountUID,
		"cloud_config": []interface{}{
			map[string]interface{}{
				"project": "test-gcp-project",
				"region":  "us-central1",
			},
		},
		"machine_pool": machinePools,
	}
}

func gkeUpdateBaseAttrs() map[string]string {
	return map[string]string{
		"name":                       "test-gke-cluster",
		"context":                    "project",
		"cloud_account_id":           gkeCloudAccountUID,
		"cloud_config_id":            gkeCloudConfigUID,
		"cluster_profile.#":          "0",
		"cloud_config.#":             "0",
		"machine_pool.#":             "0",
		"cluster_rbac_binding.#":     "0",
		"namespaces.#":               "0",
		"host_config.#":              "0",
		"location_config.#":          "0",
		"scan_policy.#":              "0",
		"backup_policy.#":            "0",
		"cluster_meta_attribute":     "",
		"cluster_timezone":           "",
		"pause_agent_upgrades":       "",
		"os_patch_on_boot":           "false",
		"os_patch_schedule":          "",
		"os_patch_after":             "",
		"review_repave_state":        "",
		"renew_k8s_certificates_now": "false",
	}
}

// TestFlattenCloudConfigGke_GetCloudConfigError exercises
// flattenCloudConfigGke's GetCloudConfigGke error branch directly.
func TestFlattenCloudConfigGke_GetCloudConfigError(t *testing.T) {
	d := resourceClusterGke().TestResourceData()
	c := mustUnitClient(t, false)

	diags := flattenCloudConfigGke(routes.GkeCloudConfigGetErrorUID, d, c)
	assert.True(t, diags.HasError())
}

// TestResourceClusterGkeUpdate_RepaveApprovalError exercises the
// validateSystemRepaveApproval error branch: cluster-uid-server-error makes
// GetCluster fail, which resourceClusterGkeUpdate must surface via
// diag.FromErr before touching machine_pool at all.
func TestResourceClusterGkeUpdate_RepaveApprovalError(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterGke(), "cluster-uid-server-error",
		gkeUpdateBaseAttrs(), simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterGkeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterGkeUpdate_GetCloudConfigError exercises the
// unconditional GetCloudConfigGke error branch at the top of
// resourceClusterGkeUpdate.
func TestResourceClusterGkeUpdate_GetCloudConfigError(t *testing.T) {
	base := gkeUpdateBaseAttrs()
	base["cloud_config_id"] = routes.GkeCloudConfigGetErrorUID

	d := buildUpdateResourceData(resourceClusterGke(), gkeClusterID,
		base, simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterGkeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterGkeUpdate_MachinePoolDiff drives, in a single Update
// call, all three machine_pool branches gated by
// d.HasChange("machine_pool"): create (new-pool), update (pool-to-change,
// hash differs on count), removal (pool-to-remove), and the unchanged
// no-op branch (pool-keep).
//
// See resource_cluster_eks_update_diff_test.go's
// TestResourceClusterEksUpdate_MachinePoolDiff for why we don't assert
// exact Set lengths here: schema.InternalMap(res.Schema).Data reconstructs
// a zero-valued phantom entry for removed set elements, which
// resourceClusterGkeUpdate's `if name != ""` guard already tolerates.
func TestResourceClusterGkeUpdate_MachinePoolDiff(t *testing.T) {
	poolKeep := defaultGkeMachinePool(map[string]interface{}{"name": "pool-keep"})
	poolChangeOld := defaultGkeMachinePool(map[string]interface{}{"name": "pool-to-change", "count": 2})
	poolChangeNew := defaultGkeMachinePool(map[string]interface{}{"name": "pool-to-change", "count": 5})
	poolRemove := defaultGkeMachinePool(map[string]interface{}{"name": "pool-to-remove"})
	newPool := defaultGkeMachinePool(map[string]interface{}{"name": "new-pool"})

	oldRaw := baseGkeUpdateRaw([]interface{}{poolKeep, poolChangeOld, poolRemove})
	newRaw := baseGkeUpdateRaw([]interface{}{poolKeep, poolChangeNew, newPool})

	d := buildGkeUpdateResourceData(t, oldRaw, newRaw, gkeCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterGkeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterGkeUpdate_ValidateOverrideScalingError exercises the
// validateOverrideScaling error branch: a new machine pool declares
// update_strategy=OverrideScaling without an override_scaling block.
func TestResourceClusterGkeUpdate_ValidateOverrideScalingError(t *testing.T) {
	poolKeep := defaultGkeMachinePool(map[string]interface{}{"name": "pool-keep"})
	invalidPool := defaultGkeMachinePool(map[string]interface{}{
		"name":            "invalid-pool",
		"update_strategy": "OverrideScaling",
	})

	oldRaw := baseGkeUpdateRaw([]interface{}{poolKeep})
	newRaw := baseGkeUpdateRaw([]interface{}{poolKeep, invalidPool})

	d := buildGkeUpdateResourceData(t, oldRaw, newRaw, gkeCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterGkeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}
