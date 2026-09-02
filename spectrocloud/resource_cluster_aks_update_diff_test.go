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

// resourceClusterAksUpdate's cloud_config/machine_pool branches are gated
// behind d.HasChange(...), which the Set-then-Set pattern used by
// resource_cluster_aks_wave2_test.go never fires (see
// resource_cluster_eks_update_diff_test.go's buildEksUpdateResourceData for
// the rationale). buildAksUpdateResourceData builds a real InstanceState +
// config diff via Resource.Diff so HasChange/GetChange behave the way
// Terraform's own apply pipeline would produce them.
func buildAksUpdateResourceData(t *testing.T, oldRaw, newRaw map[string]interface{}, configUID string) *schema.ResourceData {
	t.Helper()
	res := resourceClusterAks()

	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId(aksClusterID)
	require.NoError(t, oldRD.Set("cloud_config_id", configUID))
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId(aksClusterID)
	return finalRD
}

// aksCloudConfigRawMap's only non-ForceNew scalar field is
// override_cluster_api_config — every other cloud_config field
// (subscription_id, resource_group, region, ssh_key, private_cluster,
// vnet_*, worker_*, control_plane_*) is ForceNew and would drive a
// destroy/re-create diff instead of the in-place update this test targets.
func aksCloudConfigRawMap(overrides map[string]interface{}) map[string]interface{} {
	cc := map[string]interface{}{
		"subscription_id": "test-subscription-id",
		"resource_group":  "test-rg",
		"region":          "eastus",
		"ssh_key":         "ssh-rsa AAAA",
	}
	for k, v := range overrides {
		cc[k] = v
	}
	return cc
}

func baseAksRaw(cloudConfig map[string]interface{}, machinePools []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":             "test-aks-cluster",
		"context":          "project",
		"cloud_account_id": aksCloudAccountUID,
		"cloud_config":     []interface{}{cloudConfig},
		"machine_pool":     machinePools,
	}
}

// aksMachinePoolRaw mirrors defaultAksMachinePool (resource_cluster_aks_wave2_test.go)
// but is used for building the raw Diff config, matching the shape
// schema.TestResourceDataRaw / terraform.NewResourceConfigRaw expect.
func aksMachinePoolRaw(overrides map[string]interface{}) map[string]interface{} {
	pool := map[string]interface{}{
		"name":                 "worker-pool",
		"count":                2,
		"instance_type":        "Standard_D2s_v3",
		"disk_size_gb":         128,
		"storage_account_type": "Premium_LRS",
		"is_system_node_pool":  false,
	}
	for k, v := range overrides {
		pool[k] = v
	}
	return pool
}

// TestResourceClusterAksUpdate_RepaveApprovalError exercises the
// validateSystemRepaveApproval error branch: cluster-uid-server-error makes
// GetCluster fail, which resourceClusterAksUpdate must surface via
// diag.FromErr before touching cloud_config/machine_pool at all.
func TestResourceClusterAksUpdate_RepaveApprovalError(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterAks(), "cluster-uid-server-error",
		baseAksUpdateAttrs(), simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterAksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterAksUpdate_GetCloudConfigError exercises the
// unconditional GetCloudConfigAks error branch that runs regardless of
// whether cloud_config itself changed.
func TestResourceClusterAksUpdate_GetCloudConfigError(t *testing.T) {
	base := baseAksUpdateAttrs()
	base["cloud_config_id"] = routes.AksCloudConfigGetErrorUID

	d := buildUpdateResourceData(resourceClusterAks(), aksClusterID,
		base, simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterAksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterAksUpdate_CloudConfigDiff exercises the
// d.HasChange("cloud_config") branch's happy path: toCloudConfigAks +
// UpdateCloudConfigAks are invoked with the new cloud_config values.
func TestResourceClusterAksUpdate_CloudConfigDiff(t *testing.T) {
	pools := []interface{}{aksMachinePoolRaw(nil)}
	oldRaw := baseAksRaw(aksCloudConfigRawMap(nil), pools)
	newRaw := baseAksRaw(aksCloudConfigRawMap(map[string]interface{}{"override_cluster_api_config": "kind: Cluster"}), pools)

	d := buildAksUpdateResourceData(t, oldRaw, newRaw, aksCloudConfigUID)
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterAksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterAksUpdate_CloudConfigUpdateError exercises the
// UpdateCloudConfigAks API-error branch inside the cloud_config HasChange
// block.
func TestResourceClusterAksUpdate_CloudConfigUpdateError(t *testing.T) {
	pools := []interface{}{aksMachinePoolRaw(nil)}
	oldRaw := baseAksRaw(aksCloudConfigRawMap(nil), pools)
	newRaw := baseAksRaw(aksCloudConfigRawMap(map[string]interface{}{"override_cluster_api_config": "kind: Cluster"}), pools)

	d := buildAksUpdateResourceData(t, oldRaw, newRaw, routes.AksCloudConfigUpdateErrorUID)
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterAksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterAksUpdate_MachinePoolDiff drives, in a single Update
// call, all three machine_pool branches gated by d.HasChange("machine_pool"):
// create (new-pool), update (pool-to-change, hash differs on count), removal
// (pool-to-remove), and the unchanged no-op branch (pool-keep).
//
// Note: schema.InternalMap(res.Schema).Data(oldState, diff) reconstructs a
// TypeSet's new value including a zero-valued phantom entry for any removed
// set element (a known quirk of rebuilding sets from a raw diff outside a
// live apply). resourceClusterAksUpdate already defends against this via its
// `if name != ""` guard, so we don't assert exact Set lengths here — only
// that the Update call completes successfully having exercised every
// create/update/delete/unchanged branch.
func TestResourceClusterAksUpdate_MachinePoolDiff(t *testing.T) {
	poolKeep := aksMachinePoolRaw(map[string]interface{}{"name": "pool-keep"})
	poolChangeOld := aksMachinePoolRaw(map[string]interface{}{"name": "pool-to-change", "count": 2})
	poolChangeNew := aksMachinePoolRaw(map[string]interface{}{"name": "pool-to-change", "count": 5})
	poolRemove := aksMachinePoolRaw(map[string]interface{}{"name": "pool-to-remove"})
	newPool := aksMachinePoolRaw(map[string]interface{}{"name": "new-pool"})

	cc := aksCloudConfigRawMap(nil)
	oldRaw := baseAksRaw(cc, []interface{}{poolKeep, poolChangeOld, poolRemove})
	newRaw := baseAksRaw(cc, []interface{}{poolKeep, poolChangeNew, newPool})

	d := buildAksUpdateResourceData(t, oldRaw, newRaw, aksCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterAksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterAksUpdate_ValidateOverrideScalingError exercises the
// validateOverrideScaling error branch: a new machine pool declares
// update_strategy=OverrideScaling without an override_scaling block.
func TestResourceClusterAksUpdate_ValidateOverrideScalingError(t *testing.T) {
	poolKeep := aksMachinePoolRaw(nil)
	invalidPool := aksMachinePoolRaw(map[string]interface{}{
		"name":            "invalid-pool",
		"update_strategy": "OverrideScaling",
	})

	cc := aksCloudConfigRawMap(nil)
	oldRaw := baseAksRaw(cc, []interface{}{poolKeep})
	newRaw := baseAksRaw(cc, []interface{}{poolKeep, invalidPool})

	d := buildAksUpdateResourceData(t, oldRaw, newRaw, aksCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterAksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}
