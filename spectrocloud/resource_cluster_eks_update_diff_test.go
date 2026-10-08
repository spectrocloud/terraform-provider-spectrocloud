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

// resourceClusterEksUpdate's cloud_config/fargate_profile/machine_pool
// branches are all gated behind d.HasChange(...), which the Set-then-Set
// pattern used by resource_cluster_eks_wave2_test.go never fires (see
// resource_cluster_edge_native_update_diff_test.go's
// buildEdgeNativeMachinePoolChangeResourceData for the rationale).
// buildEksUpdateResourceData builds a real InstanceState + config diff via
// Resource.Diff so HasChange/GetChange behave the way Terraform's own apply
// pipeline would produce them.
func buildEksUpdateResourceData(t *testing.T, oldRaw, newRaw map[string]interface{}, configUID string) *schema.ResourceData {
	t.Helper()
	res := resourceClusterEks()

	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId(eksClusterID)
	require.NoError(t, oldRD.Set("cloud_config_id", configUID))
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId(eksClusterID)
	return finalRD
}

func eksCloudConfigRaw(overrides map[string]interface{}) map[string]interface{} {
	cc := map[string]interface{}{
		"region":          "us-east-1",
		"vpc_id":          "vpc-test123",
		"ssh_key_name":    "test-key",
		"endpoint_access": "public",
	}
	for k, v := range overrides {
		cc[k] = v
	}
	return cc
}

func baseEksUpdateRaw(cloudConfig map[string]interface{}, machinePools, fargateProfiles []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":             "test-eks-cluster",
		"context":          "project",
		"cloud_account_id": eksCloudAccountUID,
		"cloud_config":     []interface{}{cloudConfig},
		"machine_pool":     machinePools,
		"fargate_profile":  fargateProfiles,
	}
}

func baseEksUpdateAttrs() map[string]string {
	return map[string]string{
		"name":                       "test-eks-cluster",
		"context":                    "project",
		"cloud_account_id":           eksCloudAccountUID,
		"cloud_config_id":            eksCloudConfigUID,
		"cluster_profile.#":          "0",
		"cloud_config.#":             "0",
		"machine_pool.#":             "0",
		"fargate_profile.#":          "0",
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

// TestResourceClusterEksUpdate_RepaveApprovalError exercises the
// validateSystemRepaveApproval error branch: cluster-uid-server-error makes
// GetCluster fail, which resourceClusterEksUpdate must surface via
// diag.FromErr before touching cloud_config/machine_pool at all.
func TestResourceClusterEksUpdate_RepaveApprovalError(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterEks(), "cluster-uid-server-error",
		baseEksUpdateAttrs(), simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterEksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterEksUpdate_GetCloudConfigError exercises the
// unconditional GetCloudConfigEks error branch that runs regardless of
// whether cloud_config itself changed.
func TestResourceClusterEksUpdate_GetCloudConfigError(t *testing.T) {
	base := baseEksUpdateAttrs()
	base["cloud_config_id"] = routes.EksCloudConfigGetErrorUID

	d := buildUpdateResourceData(resourceClusterEks(), eksClusterID,
		base, simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterEksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterEksUpdate_CloudConfigDiff exercises the
// d.HasChange("cloud_config") branch's happy path: toCloudConfigEks +
// UpdateCloudConfigEks are invoked with the new cloud_config values.
func TestResourceClusterEksUpdate_CloudConfigDiff(t *testing.T) {
	// override_cluster_api_config is the only non-ForceNew scalar field
	// inside cloud_config — every other field (region, vpc_id,
	// ssh_key_name, azs, az_subnets, endpoint_access,
	// encryption_config_arn) is ForceNew and would drive a
	// destroy/re-create diff instead of the in-place update this test
	// targets.
	pools := []interface{}{defaultEksMachinePool(nil)}
	oldRaw := baseEksUpdateRaw(eksCloudConfigRaw(nil), pools, []interface{}{})
	newRaw := baseEksUpdateRaw(eksCloudConfigRaw(map[string]interface{}{"override_cluster_api_config": "kind: Cluster"}), pools, []interface{}{})

	d := buildEksUpdateResourceData(t, oldRaw, newRaw, eksCloudConfigUID)
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterEksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterEksUpdate_CloudConfigUpdateError exercises the
// UpdateCloudConfigEks API-error branch inside the cloud_config
// HasChange block.
func TestResourceClusterEksUpdate_CloudConfigUpdateError(t *testing.T) {
	pools := []interface{}{defaultEksMachinePool(nil)}
	oldRaw := baseEksUpdateRaw(eksCloudConfigRaw(nil), pools, []interface{}{})
	newRaw := baseEksUpdateRaw(eksCloudConfigRaw(map[string]interface{}{"override_cluster_api_config": "kind: Cluster"}), pools, []interface{}{})

	d := buildEksUpdateResourceData(t, oldRaw, newRaw, routes.EksCloudConfigUpdateErrorUID)
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterEksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

func eksFargateProfileRaw(name string) map[string]interface{} {
	return map[string]interface{}{
		"name":            name,
		"subnets":         []interface{}{"subnet-1"},
		"additional_tags": map[string]interface{}{},
		"selector": []interface{}{
			map[string]interface{}{
				"namespace": "default",
				"labels":    map[string]interface{}{},
			},
		},
	}
}

// TestResourceClusterEksUpdate_FargateProfileDiff exercises the
// d.HasChange("fargate_profile") branch's happy path.
func TestResourceClusterEksUpdate_FargateProfileDiff(t *testing.T) {
	pools := []interface{}{defaultEksMachinePool(nil)}
	cc := eksCloudConfigRaw(nil)
	oldRaw := baseEksUpdateRaw(cc, pools, []interface{}{})
	newRaw := baseEksUpdateRaw(cc, pools, []interface{}{eksFargateProfileRaw("fargate-1")})

	d := buildEksUpdateResourceData(t, oldRaw, newRaw, eksCloudConfigUID)
	require.True(t, d.HasChange("fargate_profile"))

	diags := resourceClusterEksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterEksUpdate_FargateProfileUpdateError exercises the
// UpdateFargateProfilesEks API-error branch.
func TestResourceClusterEksUpdate_FargateProfileUpdateError(t *testing.T) {
	pools := []interface{}{defaultEksMachinePool(nil)}
	cc := eksCloudConfigRaw(nil)
	oldRaw := baseEksUpdateRaw(cc, pools, []interface{}{})
	newRaw := baseEksUpdateRaw(cc, pools, []interface{}{eksFargateProfileRaw("fargate-1")})

	d := buildEksUpdateResourceData(t, oldRaw, newRaw, routes.EksFargateUpdateErrorUID)
	require.True(t, d.HasChange("fargate_profile"))

	diags := resourceClusterEksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterEksUpdate_MachinePoolDiff drives, in a single Update
// call, all three machine_pool branches gated by
// d.HasChange("machine_pool"): create (new-pool), update (pool-to-change,
// hash differs on count), removal (pool-to-remove), and the unchanged
// no-op branch (pool-keep).
//
// Note: schema.InternalMap(res.Schema).Data(oldState, diff) reconstructs a
// TypeSet's new value including a zero-valued phantom entry for any removed
// set element (a known quirk of rebuilding sets from a raw diff outside a
// live apply). resourceClusterEksUpdate already defends against this via
// its `if name != ""` guard, so we don't assert exact Set lengths here —
// only that the Update call completes successfully having exercised every
// create/update/delete/unchanged branch.
func TestResourceClusterEksUpdate_MachinePoolDiff(t *testing.T) {
	poolKeep := defaultEksMachinePool(map[string]interface{}{"name": "pool-keep"})
	poolChangeOld := defaultEksMachinePool(map[string]interface{}{"name": "pool-to-change", "count": 2})
	poolChangeNew := defaultEksMachinePool(map[string]interface{}{"name": "pool-to-change", "count": 5})
	poolRemove := defaultEksMachinePool(map[string]interface{}{"name": "pool-to-remove"})
	newPool := defaultEksMachinePool(map[string]interface{}{"name": "new-pool"})

	cc := eksCloudConfigRaw(nil)
	oldRaw := baseEksUpdateRaw(cc, []interface{}{poolKeep, poolChangeOld, poolRemove}, []interface{}{})
	newRaw := baseEksUpdateRaw(cc, []interface{}{poolKeep, poolChangeNew, newPool}, []interface{}{})

	d := buildEksUpdateResourceData(t, oldRaw, newRaw, eksCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterEksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterEksUpdate_ValidateOverrideScalingError exercises the
// validateOverrideScaling error branch: a new machine pool declares
// update_strategy=OverrideScaling without an override_scaling block.
func TestResourceClusterEksUpdate_ValidateOverrideScalingError(t *testing.T) {
	poolKeep := defaultEksMachinePool(map[string]interface{}{"name": "pool-keep"})
	invalidPool := defaultEksMachinePool(map[string]interface{}{
		"name":            "invalid-pool",
		"update_strategy": "OverrideScaling",
	})

	cc := eksCloudConfigRaw(nil)
	oldRaw := baseEksUpdateRaw(cc, []interface{}{poolKeep}, []interface{}{})
	newRaw := baseEksUpdateRaw(cc, []interface{}{poolKeep, invalidPool}, []interface{}{})

	d := buildEksUpdateResourceData(t, oldRaw, newRaw, eksCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterEksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}
