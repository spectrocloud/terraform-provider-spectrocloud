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

// resourceClusterAzureUpdate's cloud_config/machine_pool branches are gated
// behind d.HasChange(...), which the Set-then-Set pattern used by
// resource_cluster_azure_wave2_test.go never fires (see
// resource_cluster_eks_update_diff_test.go's rationale, and this session's
// AWS/AKS/MAAS equivalents). buildAzureUpdateResourceData builds a real
// InstanceState + config diff via Resource.Diff so HasChange/GetChange
// behave the way Terraform's own apply pipeline would produce them.
func buildAzureUpdateResourceData(t *testing.T, oldRaw, newRaw map[string]interface{}, configUID string) *schema.ResourceData {
	t.Helper()
	res := resourceClusterAzure()

	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId(azureClusterID)
	require.NoError(t, oldRD.Set("cloud_config_id", configUID))
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId(azureClusterID)
	return finalRD
}

// azureCloudConfigDiffRawMap builds the cloud_config raw map used for Diff
// tests. Unlike AWS/AKS, Azure's cloud_config block has NO ForceNew fields
// at all, so any scalar (e.g. storage_account_name) can safely drive the
// cloud_config HasChange branch without triggering a destroy/re-create.
func azureCloudConfigDiffRawMap(overrides map[string]interface{}) map[string]interface{} {
	cc := map[string]interface{}{
		"subscription_id": "test-subscription-id",
		"resource_group":  "test-rg",
		"region":          "eastus",
		"ssh_key":         "test-ssh-key",
	}
	for k, v := range overrides {
		cc[k] = v
	}
	return cc
}

func baseAzureRaw(cloudConfig map[string]interface{}, machinePools []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":             "test-azure-cluster",
		"context":          "project",
		"cloud_account_id": azureCloudAccountUID,
		"cloud_config":     []interface{}{cloudConfig},
		"machine_pool":     machinePools,
	}
}

// azureMachinePoolDiffRaw mirrors defaultAzureMachinePool
// (resource_cluster_azure_wave2_test.go) but is used for building the raw
// Diff config, matching the shape schema.TestResourceDataRaw /
// terraform.NewResourceConfigRaw expect. None of machine_pool's fields are
// ForceNew (all commented out in the schema), so any field is safe to diff.
func azureMachinePoolDiffRaw(overrides map[string]interface{}) map[string]interface{} {
	pool := map[string]interface{}{
		"name":          "worker-pool",
		"count":         2,
		"instance_type": "Standard_D2s_v3",
	}
	for k, v := range overrides {
		pool[k] = v
	}
	return pool
}

// TestResourceClusterAzureUpdate_RepaveApprovalError exercises the
// validateSystemRepaveApproval error branch: cluster-uid-server-error makes
// GetCluster fail, which resourceClusterAzureUpdate must surface via
// diag.FromErr before touching cloud_config/machine_pool at all.
func TestResourceClusterAzureUpdate_RepaveApprovalError(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterAzure(), "cluster-uid-server-error",
		baseAzureUpdateAttrs(), simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterAzureUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterAzureUpdate_GetCloudConfigError exercises the
// unconditional GetCloudConfigAzure error branch that runs regardless of
// whether cloud_config itself changed.
func TestResourceClusterAzureUpdate_GetCloudConfigError(t *testing.T) {
	base := baseAzureUpdateAttrs()
	base["cloud_config_id"] = routes.AzureCloudConfigGetErrorUID

	d := buildUpdateResourceData(resourceClusterAzure(), azureClusterID,
		base, simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterAzureUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterAzureUpdate_CloudConfigDiff exercises the
// d.HasChange("cloud_config") branch's happy path: azureClusterConfigFromMap
// + UpdateCloudConfigAzure are invoked with the new cloud_config values.
func TestResourceClusterAzureUpdate_CloudConfigDiff(t *testing.T) {
	pools := []interface{}{azureMachinePoolDiffRaw(nil)}
	oldRaw := baseAzureRaw(azureCloudConfigDiffRawMap(nil), pools)
	newRaw := baseAzureRaw(azureCloudConfigDiffRawMap(map[string]interface{}{"storage_account_name": "newaccount"}), pools)

	d := buildAzureUpdateResourceData(t, oldRaw, newRaw, azureCloudConfigUID)
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterAzureUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterAzureUpdate_CloudConfigUpdateError exercises the
// UpdateCloudConfigAzure API-error branch inside the cloud_config HasChange
// block.
func TestResourceClusterAzureUpdate_CloudConfigUpdateError(t *testing.T) {
	pools := []interface{}{azureMachinePoolDiffRaw(nil)}
	oldRaw := baseAzureRaw(azureCloudConfigDiffRawMap(nil), pools)
	newRaw := baseAzureRaw(azureCloudConfigDiffRawMap(map[string]interface{}{"storage_account_name": "newaccount"}), pools)

	d := buildAzureUpdateResourceData(t, oldRaw, newRaw, routes.AzureCloudConfigUpdateErrorUID)
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterAzureUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterAzureUpdate_MachinePoolDiff drives, in a single Update
// call, all three machine_pool branches gated by d.HasChange("machine_pool"):
// create (new-pool), update (pool-to-change, hash differs on count), removal
// (pool-to-remove), and the unchanged no-op branch (pool-keep). All pools
// are worker pools (control_plane=false) to keep validateCPPoolCount out of
// the picture for this happy-path test — see
// TestResourceClusterAzureUpdate_ValidateCPPoolCountError for that branch.
//
// Note: schema.InternalMap(res.Schema).Data(oldState, diff) reconstructs a
// TypeSet's new value including a zero-valued phantom entry for any removed
// set element (a known quirk of rebuilding sets from a raw diff outside a
// live apply). resourceClusterAzureUpdate already defends against this via
// its `if name != ""` guard, so we don't assert exact Set lengths here —
// only that the Update call completes successfully having exercised every
// create/update/delete/unchanged branch.
func TestResourceClusterAzureUpdate_MachinePoolDiff(t *testing.T) {
	poolKeep := azureMachinePoolDiffRaw(map[string]interface{}{"name": "pool-keep"})
	poolChangeOld := azureMachinePoolDiffRaw(map[string]interface{}{"name": "pool-to-change", "count": 2})
	poolChangeNew := azureMachinePoolDiffRaw(map[string]interface{}{"name": "pool-to-change", "count": 5})
	poolRemove := azureMachinePoolDiffRaw(map[string]interface{}{"name": "pool-to-remove"})
	newPool := azureMachinePoolDiffRaw(map[string]interface{}{"name": "new-pool"})

	cc := azureCloudConfigDiffRawMap(nil)
	oldRaw := baseAzureRaw(cc, []interface{}{poolKeep, poolChangeOld, poolRemove})
	newRaw := baseAzureRaw(cc, []interface{}{poolKeep, poolChangeNew, newPool})

	d := buildAzureUpdateResourceData(t, oldRaw, newRaw, azureCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterAzureUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterAzureUpdate_ValidateOverrideScalingError exercises the
// validateOverrideScaling error branch: a new machine pool declares
// update_strategy=OverrideScaling without an override_scaling block.
func TestResourceClusterAzureUpdate_ValidateOverrideScalingError(t *testing.T) {
	poolKeep := azureMachinePoolDiffRaw(nil)
	invalidPool := azureMachinePoolDiffRaw(map[string]interface{}{
		"name":            "invalid-pool",
		"update_strategy": "OverrideScaling",
	})

	cc := azureCloudConfigDiffRawMap(nil)
	oldRaw := baseAzureRaw(cc, []interface{}{poolKeep})
	newRaw := baseAzureRaw(cc, []interface{}{poolKeep, invalidPool})

	d := buildAzureUpdateResourceData(t, oldRaw, newRaw, azureCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterAzureUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterAzureUpdate_ValidateCPPoolCountError exercises the
// validateCPPoolCount diag branch inside the machine_pool HasChange block: a
// control-plane pool with an even size is rejected before any
// create/update/delete API calls are made.
func TestResourceClusterAzureUpdate_ValidateCPPoolCountError(t *testing.T) {
	poolKeep := azureMachinePoolDiffRaw(nil)
	evenCPPool := azureMachinePoolDiffRaw(map[string]interface{}{
		"name":          "cp-pool",
		"count":         4,
		"control_plane": true,
	})

	cc := azureCloudConfigDiffRawMap(nil)
	oldRaw := baseAzureRaw(cc, []interface{}{poolKeep})
	newRaw := baseAzureRaw(cc, []interface{}{poolKeep, evenCPPool})

	d := buildAzureUpdateResourceData(t, oldRaw, newRaw, azureCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterAzureUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}
