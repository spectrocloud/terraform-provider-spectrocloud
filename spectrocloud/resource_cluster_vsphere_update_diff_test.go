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

// resourceClusterVsphereUpdate's cloud_config/machine_pool branches are
// gated behind d.HasChange(...), which the Set-then-Set pattern used by
// resource_cluster_vsphere_wave2_test.go never fires (see
// resource_cluster_eks_update_diff_test.go's buildEksUpdateResourceData for
// the rationale). buildVsphereUpdateResourceData builds a real InstanceState
// + config diff via Resource.Diff so HasChange/GetChange behave the way
// Terraform's own apply pipeline would produce them.
//
// Note: unlike EKS/GKE, vSphere's whole cloud_config block is ForceNew (see
// resource_cluster_vsphere.go's "cloud_config" schema entry), so any field
// change inside it produces a destroy/recreate diff via res.Diff() — not the
// in-place update resourceClusterVsphereUpdate's cloud_config HasChange
// branch expects. We drive those tests through buildUpdateResourceData's
// manual InstanceState+InstanceDiff construction instead (see
// vsphereCloudConfigBaseAttrs below), which bypasses schema-level ForceNew
// entirely — matching how runUpdateHasChangeSuite already exercises scalar
// fields.
func buildVsphereUpdateResourceData(t *testing.T, oldRaw, newRaw map[string]interface{}, configUID string) *schema.ResourceData {
	t.Helper()
	res := resourceClusterVsphere()

	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId(vsphereClusterID)
	require.NoError(t, oldRD.Set("cloud_config_id", configUID))
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId(vsphereClusterID)
	return finalRD
}

func vsphereCloudConfigRawMap(overrides map[string]interface{}) map[string]interface{} {
	cc := map[string]interface{}{
		"datacenter":    "test-datacenter",
		"folder":        "test-folder",
		"ssh_key":       "ssh-rsa AAAA",
		"network_type":  "VIP",
		"host_endpoint": "10.0.0.100",
	}
	for k, v := range overrides {
		cc[k] = v
	}
	return cc
}

func baseVsphereRaw(cloudConfig map[string]interface{}, machinePools []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":             "test-vsphere-cluster",
		"context":          "project",
		"cloud_account_id": vsphereCloudAccountUID,
		"cloud_config":     []interface{}{cloudConfig},
		"machine_pool":     machinePools,
	}
}

// mergeStrMap combines multiple string maps left-to-right; later maps
// overwrite earlier ones on key conflict.
func mergeStrMap(maps ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

// vsphereCloudConfigAttrs is the flatmap-attribute equivalent of
// vsphereCloudConfigRawMap, for use with buildUpdateResourceData's manual
// InstanceState+InstanceDiff construction (which bypasses ForceNew).
func vsphereCloudConfigAttrs(overrides map[string]string) map[string]string {
	base := map[string]string{
		"cloud_config.#":                             "1",
		"cloud_config.0.datacenter":                  "test-datacenter",
		"cloud_config.0.folder":                      "test-folder",
		"cloud_config.0.image_template_folder":       "",
		"cloud_config.0.ssh_key":                     "ssh-rsa AAAA",
		"cloud_config.0.ssh_keys.#":                  "0",
		"cloud_config.0.static_ip":                   "false",
		"cloud_config.0.network_type":                "VIP",
		"cloud_config.0.host_endpoint":               "10.0.0.100",
		"cloud_config.0.network_search_domain":       "",
		"cloud_config.0.ntp_servers.#":               "0",
		"cloud_config.0.override_cluster_api_config": "",
	}
	for k, v := range overrides {
		base[k] = v
	}
	return base
}

func vsphereCloudConfigBaseAttrs(overrides map[string]string) map[string]string {
	return mergeStrMap(baseVsphereUpdateAttrs(), vsphereCloudConfigAttrs(nil), overrides)
}

// TestResourceClusterVsphereUpdate_RepaveApprovalError exercises the
// validateSystemRepaveApproval error branch: cluster-uid-server-error makes
// GetCluster fail, which resourceClusterVsphereUpdate must surface via
// diag.FromErr before touching cloud_config/machine_pool at all.
func TestResourceClusterVsphereUpdate_RepaveApprovalError(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterVsphere(), "cluster-uid-server-error",
		baseVsphereUpdateAttrs(), simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterVsphereUpdate_GetCloudConfigError exercises the
// unconditional GetCloudConfigVsphere error branch that runs regardless of
// whether cloud_config itself changed.
func TestResourceClusterVsphereUpdate_GetCloudConfigError(t *testing.T) {
	base := baseVsphereUpdateAttrs()
	base["cloud_config_id"] = routes.VsphereCloudConfigErrorUID

	d := buildUpdateResourceData(resourceClusterVsphere(), vsphereClusterID,
		base, simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterVsphereUpdate_CloudConfigDatacenterChangeError exercises
// the explicit validation inside the cloud_config HasChange branch: changing
// datacenter after provisioning must be rejected.
func TestResourceClusterVsphereUpdate_CloudConfigDatacenterChangeError(t *testing.T) {
	base := vsphereCloudConfigBaseAttrs(nil)
	d := buildUpdateResourceData(resourceClusterVsphere(), vsphereClusterID, base,
		simpleDiff("cloud_config.0.datacenter", "test-datacenter", "changed-datacenter"))
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterVsphereUpdate_CloudConfigDiff exercises the
// d.HasChange("cloud_config") branch's happy path: toCloudConfigUpdate +
// UpdateCloudConfigVsphere are invoked with the new cloud_config values.
func TestResourceClusterVsphereUpdate_CloudConfigDiff(t *testing.T) {
	base := vsphereCloudConfigBaseAttrs(nil)
	d := buildUpdateResourceData(resourceClusterVsphere(), vsphereClusterID, base,
		simpleDiff("cloud_config.0.network_search_domain", "", "spectrocloud.dev"))
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterVsphereUpdate_CloudConfigUpdateError exercises the
// UpdateCloudConfigVsphere API-error branch inside the cloud_config
// HasChange block.
func TestResourceClusterVsphereUpdate_CloudConfigUpdateError(t *testing.T) {
	base := vsphereCloudConfigBaseAttrs(map[string]string{"cloud_config_id": routes.VsphereCloudConfigUpdateErrorUID})
	d := buildUpdateResourceData(resourceClusterVsphere(), vsphereClusterID, base,
		simpleDiff("cloud_config.0.network_search_domain", "", "spectrocloud.dev"))
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterVsphereUpdate_MachinePoolDiff drives, in a single Update
// call, all three machine_pool branches gated by d.HasChange("machine_pool"):
// create (new-pool), update (pool-to-change, hash differs on count), removal
// (pool-to-remove), and the unchanged no-op branch (pool-keep). All pools
// here are workers so ValidateMachinePoolChange's control-plane placement
// check trivially passes (no control-plane pools on either side).
//
// See resource_cluster_eks_update_diff_test.go's
// TestResourceClusterEksUpdate_MachinePoolDiff for why we don't assert exact
// Set lengths: schema.InternalMap(res.Schema).Data reconstructs a zero-valued
// phantom entry for removed set elements, which resourceClusterVsphereUpdate's
// `if name != ""` guard already tolerates.
func TestResourceClusterVsphereUpdate_MachinePoolDiff(t *testing.T) {
	worker := func(overrides map[string]interface{}) map[string]interface{} {
		base := map[string]interface{}{
			"control_plane":           false,
			"control_plane_as_worker": false,
		}
		for k, v := range overrides {
			base[k] = v
		}
		return defaultVsphereMachinePool(base)
	}

	poolKeep := worker(map[string]interface{}{"name": "pool-keep"})
	poolChangeOld := worker(map[string]interface{}{"name": "pool-to-change", "count": 2})
	poolChangeNew := worker(map[string]interface{}{"name": "pool-to-change", "count": 5})
	poolRemove := worker(map[string]interface{}{"name": "pool-to-remove"})
	newPool := worker(map[string]interface{}{"name": "new-pool"})

	cc := vsphereCloudConfigRawMap(nil)
	oldRaw := baseVsphereRaw(cc, []interface{}{poolKeep, poolChangeOld, poolRemove})
	newRaw := baseVsphereRaw(cc, []interface{}{poolKeep, poolChangeNew, newPool})

	d := buildVsphereUpdateResourceData(t, oldRaw, newRaw, vsphereCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterVsphereUpdate_ControlPlanePoolChange exercises the
// PEM-5013 branch: when a changed pool is the control plane, its placement's
// Datacenter/Folder get overwritten from the (unchanged) cloud_config block
// before UpdateMachinePoolVsphere is called.
func TestResourceClusterVsphereUpdate_ControlPlanePoolChange(t *testing.T) {
	oldCp := defaultVsphereMachinePool(map[string]interface{}{"count": 1})
	newCp := defaultVsphereMachinePool(map[string]interface{}{"count": 3})

	cc := vsphereCloudConfigRawMap(nil)
	oldRaw := baseVsphereRaw(cc, []interface{}{oldCp})
	newRaw := baseVsphereRaw(cc, []interface{}{newCp})

	d := buildVsphereUpdateResourceData(t, oldRaw, newRaw, vsphereCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterVsphereUpdate_ValidateOverrideScalingError exercises the
// validateOverrideScaling error branch: a new machine pool declares
// update_strategy=OverrideScaling without an override_scaling block.
func TestResourceClusterVsphereUpdate_ValidateOverrideScalingError(t *testing.T) {
	poolKeep := defaultVsphereMachinePool(nil)
	invalidPool := defaultVsphereMachinePool(map[string]interface{}{
		"name":                    "invalid-pool",
		"control_plane":           false,
		"control_plane_as_worker": false,
		"update_strategy":         "OverrideScaling",
	})

	cc := vsphereCloudConfigRawMap(nil)
	oldRaw := baseVsphereRaw(cc, []interface{}{poolKeep})
	newRaw := baseVsphereRaw(cc, []interface{}{poolKeep, invalidPool})

	d := buildVsphereUpdateResourceData(t, oldRaw, newRaw, vsphereCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterVsphereUpdate_MachinePoolChangeValidationError exercises
// the ValidateMachinePoolChange error branch: the control-plane pool's
// placement.cluster changes between old and new, which is rejected before
// any machine pool CRUD calls are made.
func TestResourceClusterVsphereUpdate_MachinePoolChangeValidationError(t *testing.T) {
	oldCp := defaultVsphereMachinePool(nil)
	newCp := defaultVsphereMachinePool(map[string]interface{}{
		"placement": []interface{}{map[string]interface{}{
			"id":                "",
			"cluster":           "changed-cluster",
			"resource_pool":     "test-pool",
			"datastore":         "test-datastore",
			"network":           "test-network",
			"static_ip_pool_id": "",
		}},
	})

	cc := vsphereCloudConfigRawMap(nil)
	oldRaw := baseVsphereRaw(cc, []interface{}{oldCp})
	newRaw := baseVsphereRaw(cc, []interface{}{newCp})

	d := buildVsphereUpdateResourceData(t, oldRaw, newRaw, vsphereCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}
