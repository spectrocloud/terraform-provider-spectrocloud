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

// resourceClusterGcpUpdate's cloud_config/machine_pool branches are gated
// behind d.HasChange(...), which the Set-then-Set pattern used by
// resource_cluster_gcp_wave2_test.go never fires (see
// resource_cluster_eks_update_diff_test.go's buildEksUpdateResourceData for
// the rationale). buildGcpUpdateResourceData builds a real InstanceState +
// config diff via Resource.Diff so HasChange/GetChange behave the way
// Terraform's own apply pipeline would produce them.
func buildGcpUpdateResourceData(t *testing.T, oldRaw, newRaw map[string]interface{}, configUID string) *schema.ResourceData {
	t.Helper()
	res := resourceClusterGcp()

	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId(gcpClusterID)
	require.NoError(t, oldRD.Set("cloud_config_id", configUID))
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId(gcpClusterID)
	return finalRD
}

const gcpClusterID = "test-gcp-cluster-id"

// defaultGcpMachinePoolRaw is the raw-config counterpart to wave2's
// defaultGcpMachinePool: "azs" here is a plain []interface{} (not a
// pre-built *schema.Set), matching what TestResourceDataRaw / NewResourceConfigRaw
// expect for a TypeSet field's raw value.
func defaultGcpMachinePoolRaw(overrides map[string]interface{}) map[string]interface{} {
	pool := map[string]interface{}{
		"name":                    "cp-pool",
		"instance_type":           "n1-standard-2",
		"count":                   1,
		"control_plane":           true,
		"control_plane_as_worker": true,
		"disk_size_gb":            65,
		"update_strategy":         "RollingUpdateScaleOut",
		"azs":                     []interface{}{"us-central1-a"},
	}
	for k, v := range overrides {
		pool[k] = v
	}
	return pool
}

func baseGcpRaw(cloudConfig map[string]interface{}, machinePools []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":             "test-gcp-cluster",
		"context":          "project",
		"cloud_account_id": gcpCloudAccountUID,
		"cloud_config":     []interface{}{cloudConfig},
		"machine_pool":     machinePools,
	}
}

func gcpCloudConfigRawMap(overrides map[string]interface{}) map[string]interface{} {
	cc := map[string]interface{}{
		"project": "test-gcp-project",
		"region":  "us-central1",
		"network": "test-network",
	}
	for k, v := range overrides {
		cc[k] = v
	}
	return cc
}

// TestResourceClusterGcpUpdate_RepaveApprovalError exercises the
// validateSystemRepaveApproval error branch: cluster-uid-server-error makes
// GetCluster fail, which resourceClusterGcpUpdate must surface via
// diag.FromErr before touching machine_pool at all.
func TestResourceClusterGcpUpdate_RepaveApprovalError(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterGcp(), "cluster-uid-server-error",
		baseGcpUpdateAttrs(), simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterGcpUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterGcpUpdate_GetCloudConfigError exercises the
// unconditional GetCloudConfigGcp error branch at the top of
// resourceClusterGcpUpdate.
func TestResourceClusterGcpUpdate_GetCloudConfigError(t *testing.T) {
	base := baseGcpUpdateAttrs()
	base["cloud_config_id"] = routes.GcpCloudConfigGetErrorUID

	d := buildUpdateResourceData(resourceClusterGcp(), gcpClusterID,
		base, simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterGcpUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterGcpUpdate_CloudConfigDiff exercises the
// d.HasChange("cloud_config") branch's happy path: toCloudConfigGcp +
// UpdateCloudConfigGcp are invoked with the new override_cluster_api_config
// value. project/region/network are unchanged, so this must not force a
// replace.
func TestResourceClusterGcpUpdate_CloudConfigDiff(t *testing.T) {
	pool := []interface{}{defaultGcpMachinePoolRaw(nil)}
	oldRaw := baseGcpRaw(gcpCloudConfigRawMap(nil), pool)
	newRaw := baseGcpRaw(gcpCloudConfigRawMap(map[string]interface{}{
		"override_cluster_api_config": "gcpClusterConfig:\n  spec:\n    releaseChannel: regular\n",
	}), pool)

	d := buildGcpUpdateResourceData(t, oldRaw, newRaw, gcpCloudConfigUID)
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterGcpUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterGcpUpdate_CloudConfigUpdateError exercises the
// UpdateCloudConfigGcp API-error branch inside the cloud_config HasChange
// block.
func TestResourceClusterGcpUpdate_CloudConfigUpdateError(t *testing.T) {
	pool := []interface{}{defaultGcpMachinePoolRaw(nil)}
	oldRaw := baseGcpRaw(gcpCloudConfigRawMap(nil), pool)
	newRaw := baseGcpRaw(gcpCloudConfigRawMap(map[string]interface{}{
		"override_cluster_api_config": "gcpClusterConfig:\n  spec:\n    releaseChannel: regular\n",
	}), pool)

	d := buildGcpUpdateResourceData(t, oldRaw, newRaw, routes.GcpCloudConfigUpdateErrorUID)
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterGcpUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterGcpUpdate_MachinePoolDiff drives, in a single Update
// call, all three machine_pool branches gated by d.HasChange("machine_pool"):
// create (new-pool), update (pool-to-change, hash differs on count), removal
// (pool-to-remove), and the unchanged no-op branch (pool-keep).
//
// See resource_cluster_eks_update_diff_test.go's
// TestResourceClusterEksUpdate_MachinePoolDiff for why we don't assert exact
// Set lengths: schema.InternalMap(res.Schema).Data reconstructs a zero-valued
// phantom entry for removed set elements, which resourceClusterGcpUpdate's
// `if name != ""` guard already tolerates.
func TestResourceClusterGcpUpdate_MachinePoolDiff(t *testing.T) {
	worker := func(overrides map[string]interface{}) map[string]interface{} {
		base := map[string]interface{}{
			"control_plane":           false,
			"control_plane_as_worker": false,
		}
		for k, v := range overrides {
			base[k] = v
		}
		return defaultGcpMachinePoolRaw(base)
	}

	poolKeep := worker(map[string]interface{}{"name": "pool-keep"})
	poolChangeOld := worker(map[string]interface{}{"name": "pool-to-change", "count": 2})
	poolChangeNew := worker(map[string]interface{}{"name": "pool-to-change", "count": 5})
	poolRemove := worker(map[string]interface{}{"name": "pool-to-remove"})
	newPool := worker(map[string]interface{}{"name": "new-pool"})

	cc := gcpCloudConfigRawMap(nil)
	oldRaw := baseGcpRaw(cc, []interface{}{poolKeep, poolChangeOld, poolRemove})
	newRaw := baseGcpRaw(cc, []interface{}{poolKeep, poolChangeNew, newPool})

	d := buildGcpUpdateResourceData(t, oldRaw, newRaw, gcpCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterGcpUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterGcpUpdate_ValidateOverrideScalingError exercises the
// validateOverrideScaling error branch: a new machine pool declares
// update_strategy=OverrideScaling without an override_scaling block.
func TestResourceClusterGcpUpdate_ValidateOverrideScalingError(t *testing.T) {
	poolKeep := defaultGcpMachinePoolRaw(nil)
	invalidPool := defaultGcpMachinePoolRaw(map[string]interface{}{
		"name":                    "invalid-pool",
		"control_plane":           false,
		"control_plane_as_worker": false,
		"update_strategy":         "OverrideScaling",
	})

	cc := gcpCloudConfigRawMap(nil)
	oldRaw := baseGcpRaw(cc, []interface{}{poolKeep})
	newRaw := baseGcpRaw(cc, []interface{}{poolKeep, invalidPool})

	d := buildGcpUpdateResourceData(t, oldRaw, newRaw, gcpCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterGcpUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterGcpStateUpgradeV2 exercises the state-upgrade branches
// directly: cluster_profile present as a list is kept as-is, a non-list
// value is left alone, and a missing key is a no-op — all three log-only
// branches return rawState unchanged, so the assertions confirm we hit each
// branch without altering the map.
func TestResourceClusterGcpStateUpgradeV2(t *testing.T) {
	t.Run("cluster_profile present as list", func(t *testing.T) {
		raw := map[string]interface{}{
			"cluster_profile": []interface{}{
				map[string]interface{}{"id": "profile-1"},
			},
		}
		out, err := resourceClusterGcpStateUpgradeV2(context.Background(), raw, nil)
		require.NoError(t, err)
		list, ok := out["cluster_profile"].([]interface{})
		require.True(t, ok)
		assert.Len(t, list, 1)
	})

	t.Run("cluster_profile present but not a list", func(t *testing.T) {
		raw := map[string]interface{}{
			"cluster_profile": "not-a-list",
		}
		out, err := resourceClusterGcpStateUpgradeV2(context.Background(), raw, nil)
		require.NoError(t, err)
		assert.Equal(t, "not-a-list", out["cluster_profile"])
	})

	t.Run("cluster_profile absent", func(t *testing.T) {
		raw := map[string]interface{}{
			"name": "some-cluster",
		}
		out, err := resourceClusterGcpStateUpgradeV2(context.Background(), raw, nil)
		require.NoError(t, err)
		_, exists := out["cluster_profile"]
		assert.False(t, exists)
		assert.Equal(t, "some-cluster", out["name"])
	})
}
