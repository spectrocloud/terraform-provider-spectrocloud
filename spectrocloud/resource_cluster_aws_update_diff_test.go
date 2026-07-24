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

// resourceClusterAwsUpdate's cloud_config/machine_pool branches are gated
// behind d.HasChange(...), which the Set-then-Set pattern used by
// resource_cluster_aws_wave2_test.go never fires (see
// resource_cluster_eks_update_diff_test.go's rationale, and this session's
// AKS/MAAS equivalents). buildAwsUpdateResourceData builds a real
// InstanceState + config diff via Resource.Diff so HasChange/GetChange
// behave the way Terraform's own apply pipeline would produce them.
func buildAwsUpdateResourceData(t *testing.T, oldRaw, newRaw map[string]interface{}, configUID string) *schema.ResourceData {
	t.Helper()
	res := resourceClusterAws()

	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId("test-cluster-id")
	require.NoError(t, oldRD.Set("cloud_config_id", configUID))
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId("test-cluster-id")
	return finalRD
}

// awsCloudConfigRawMap's only non-ForceNew scalar field is
// override_cluster_api_config — every other cloud_config field
// (ssh_key_name, region, vpc_id, control_plane_lb) is ForceNew and would
// drive a destroy/re-create diff instead of the in-place update this test
// targets.
func awsCloudConfigRawMap(overrides map[string]interface{}) map[string]interface{} {
	cc := map[string]interface{}{
		"ssh_key_name": "test-key",
		"region":       "us-east-1",
		"vpc_id":       "vpc-test123",
	}
	for k, v := range overrides {
		cc[k] = v
	}
	return cc
}

func baseAwsRaw(cloudConfig map[string]interface{}, machinePools []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":             "test-aws-cluster",
		"context":          "project",
		"cloud_account_id": awsCloudAccountUID,
		"cloud_config":     []interface{}{cloudConfig},
		"machine_pool":     machinePools,
	}
}

// awsMachinePoolRaw mirrors defaultAwsMachinePool (resource_cluster_aws_wave2_test.go)
// but is used for building the raw Diff config, matching the shape
// schema.TestResourceDataRaw / terraform.NewResourceConfigRaw expect.
func awsMachinePoolRaw(overrides map[string]interface{}) map[string]interface{} {
	pool := map[string]interface{}{
		"name":          "worker-pool",
		"count":         2,
		"instance_type": "t3.large",
		"disk_size_gb":  65,
	}
	for k, v := range overrides {
		pool[k] = v
	}
	return pool
}

// TestResourceClusterAwsUpdate_RepaveApprovalError exercises the
// validateSystemRepaveApproval error branch: cluster-uid-server-error makes
// GetCluster fail, which resourceClusterAwsUpdate must surface via
// diag.FromErr before touching cloud_config/machine_pool at all.
func TestResourceClusterAwsUpdate_RepaveApprovalError(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterAws(), "cluster-uid-server-error",
		baseAwsUpdateAttrs(), simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterAwsUpdate_GetCloudConfigError exercises the
// unconditional GetCloudConfigAws error branch that runs regardless of
// whether cloud_config itself changed.
func TestResourceClusterAwsUpdate_GetCloudConfigError(t *testing.T) {
	base := baseAwsUpdateAttrs()
	base["cloud_config_id"] = routes.AwsCloudConfigGetErrorUID

	d := buildUpdateResourceData(resourceClusterAws(), "test-cluster-id",
		base, simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterAwsUpdate_CloudConfigDiff exercises the
// d.HasChange("cloud_config") branch's happy path: toCloudConfigAws +
// UpdateCloudConfigAws are invoked with the new cloud_config values.
func TestResourceClusterAwsUpdate_CloudConfigDiff(t *testing.T) {
	pools := []interface{}{awsMachinePoolRaw(map[string]interface{}{"control_plane": true})}
	oldRaw := baseAwsRaw(awsCloudConfigRawMap(nil), pools)
	newRaw := baseAwsRaw(awsCloudConfigRawMap(map[string]interface{}{"override_cluster_api_config": "kind: Cluster"}), pools)

	d := buildAwsUpdateResourceData(t, oldRaw, newRaw, awsCloudConfigUID)
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterAwsUpdate_CloudConfigUpdateError exercises the
// UpdateCloudConfigAws API-error branch inside the cloud_config HasChange
// block.
func TestResourceClusterAwsUpdate_CloudConfigUpdateError(t *testing.T) {
	pools := []interface{}{awsMachinePoolRaw(map[string]interface{}{"control_plane": true})}
	oldRaw := baseAwsRaw(awsCloudConfigRawMap(nil), pools)
	newRaw := baseAwsRaw(awsCloudConfigRawMap(map[string]interface{}{"override_cluster_api_config": "kind: Cluster"}), pools)

	d := buildAwsUpdateResourceData(t, oldRaw, newRaw, routes.AwsCloudConfigUpdateErrorUID)
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterAwsUpdate_MachinePoolDiff drives, in a single Update
// call, all three machine_pool branches gated by d.HasChange("machine_pool"):
// create (new-pool), update (pool-to-change, hash differs on count), removal
// (pool-to-remove), and the unchanged no-op branch (pool-keep).
//
// Note: schema.InternalMap(res.Schema).Data(oldState, diff) reconstructs a
// TypeSet's new value including a zero-valued phantom entry for any removed
// set element (a known quirk of rebuilding sets from a raw diff outside a
// live apply). resourceClusterAwsUpdate already defends against this via its
// `if name != ""` guard, so we don't assert exact Set lengths here — only
// that the Update call completes successfully having exercised every
// create/update/delete/unchanged branch.
func TestResourceClusterAwsUpdate_MachinePoolDiff(t *testing.T) {
	poolKeep := awsMachinePoolRaw(map[string]interface{}{"name": "pool-keep", "control_plane": true})
	poolChangeOld := awsMachinePoolRaw(map[string]interface{}{"name": "pool-to-change", "count": 2})
	poolChangeNew := awsMachinePoolRaw(map[string]interface{}{"name": "pool-to-change", "count": 5})
	poolRemove := awsMachinePoolRaw(map[string]interface{}{"name": "pool-to-remove"})
	newPool := awsMachinePoolRaw(map[string]interface{}{"name": "new-pool"})

	cc := awsCloudConfigRawMap(nil)
	oldRaw := baseAwsRaw(cc, []interface{}{poolKeep, poolChangeOld, poolRemove})
	newRaw := baseAwsRaw(cc, []interface{}{poolKeep, poolChangeNew, newPool})

	d := buildAwsUpdateResourceData(t, oldRaw, newRaw, awsCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterAwsUpdate_ValidateOverrideScalingError exercises the
// validateOverrideScaling error branch: a new machine pool declares
// update_strategy=OverrideScaling without an override_scaling block.
func TestResourceClusterAwsUpdate_ValidateOverrideScalingError(t *testing.T) {
	poolKeep := awsMachinePoolRaw(map[string]interface{}{"control_plane": true})
	invalidPool := awsMachinePoolRaw(map[string]interface{}{
		"name":            "invalid-pool",
		"update_strategy": "OverrideScaling",
	})

	cc := awsCloudConfigRawMap(nil)
	oldRaw := baseAwsRaw(cc, []interface{}{poolKeep})
	newRaw := baseAwsRaw(cc, []interface{}{poolKeep, invalidPool})

	d := buildAwsUpdateResourceData(t, oldRaw, newRaw, awsCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterAwsUpdate_MachinePoolCreateError exercises the
// toMachinePoolAws error branch inside the "new machine pool" create path:
// a control-plane pool with a non-zero node_repave_interval is rejected by
// ValidationNodeRepaveIntervalForControlPlane.
func TestResourceClusterAwsUpdate_MachinePoolCreateError(t *testing.T) {
	poolKeep := awsMachinePoolRaw(map[string]interface{}{"control_plane": true})
	badPool := awsMachinePoolRaw(map[string]interface{}{
		"name":                 "bad-cp-pool",
		"control_plane":        true,
		"node_repave_interval": 30,
	})

	cc := awsCloudConfigRawMap(nil)
	oldRaw := baseAwsRaw(cc, []interface{}{poolKeep})
	newRaw := baseAwsRaw(cc, []interface{}{poolKeep, badPool})

	d := buildAwsUpdateResourceData(t, oldRaw, newRaw, awsCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}
