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

// resourceClusterMaasUpdate's cloud_config/machine_pool branches are gated
// behind d.HasChange(...), which the Set-then-Set pattern used by
// resource_cluster_maas_wave2_test.go never fires (see
// resource_cluster_eks_update_diff_test.go's buildEksUpdateResourceData for
// the rationale). buildMaasUpdateResourceData builds a real InstanceState +
// config diff via Resource.Diff so HasChange/GetChange behave the way
// Terraform's own apply pipeline would produce them.
//
// Unlike vSphere, none of MaaS's cloud_config or machine_pool fields are
// ForceNew (see resource_cluster_maas.go's schema), so a plain res.Diff()
// round-trip works directly — no need for the manual
// InstanceState+InstanceDiff bypass vSphere requires.
func buildMaasUpdateResourceData(t *testing.T, oldRaw, newRaw map[string]interface{}, configUID string) *schema.ResourceData {
	t.Helper()
	res := resourceClusterMaas()

	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId(maasClusterID)
	require.NoError(t, oldRD.Set("cloud_config_id", configUID))
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId(maasClusterID)
	return finalRD
}

func maasCloudConfigRawMap(overrides map[string]interface{}) map[string]interface{} {
	cc := map[string]interface{}{
		"domain":        "test-domain",
		"enable_lxd_vm": false,
	}
	for k, v := range overrides {
		cc[k] = v
	}
	return cc
}

func baseMaasRaw(cloudConfig map[string]interface{}, machinePools []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":             "test-maas-cluster",
		"context":          "project",
		"cloud_account_id": maasCloudAccountUID,
		"cloud_config":     []interface{}{cloudConfig},
		"machine_pool":     machinePools,
	}
}

// maasMachinePoolRaw mirrors defaultMaasMachinePool (resource_cluster_maas_wave2_test.go)
// but keeps "azs" as a plain []interface{} instead of a pre-built *schema.Set
// — schema.TestResourceDataRaw / terraform.NewResourceConfigRaw need the raw
// Go shape, and the real Resource.Diff pipeline handles the TypeSet
// conversion for us.
func maasMachinePoolRaw(overrides map[string]interface{}) map[string]interface{} {
	pool := map[string]interface{}{
		"name":                    "cp-pool",
		"count":                   1,
		"control_plane":           true,
		"control_plane_as_worker": true,
		"instance_type": []interface{}{
			map[string]interface{}{
				"min_memory_mb": 8192,
				"min_cpu":       4,
			},
		},
		"azs": []interface{}{"us-east-1a"},
		"placement": []interface{}{
			map[string]interface{}{
				"resource_pool": "default",
			},
		},
	}
	for k, v := range overrides {
		pool[k] = v
	}
	return pool
}

// TestResourceClusterMaasUpdate_RepaveApprovalError exercises the
// validateSystemRepaveApproval error branch: cluster-uid-server-error makes
// GetCluster fail, which resourceClusterMaasUpdate must surface via
// diag.FromErr before touching cloud_config/machine_pool at all.
func TestResourceClusterMaasUpdate_RepaveApprovalError(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterMaas(), "cluster-uid-server-error",
		baseMaasUpdateAttrs(), simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterMaasUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterMaasUpdate_GetCloudConfigError exercises the
// unconditional GetCloudConfigMaas error branch that runs regardless of
// whether cloud_config itself changed.
func TestResourceClusterMaasUpdate_GetCloudConfigError(t *testing.T) {
	base := baseMaasUpdateAttrs()
	base["cloud_config_id"] = routes.MaasCloudConfigErrorUID

	d := buildUpdateResourceData(resourceClusterMaas(), maasClusterID,
		base, simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterMaasUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterMaasUpdate_CloudConfigDiff exercises the
// d.HasChange("cloud_config") branch's happy path: toMaasCloudConfigUpdate +
// UpdateCloudConfigMaas are invoked with the new cloud_config values.
func TestResourceClusterMaasUpdate_CloudConfigDiff(t *testing.T) {
	pools := []interface{}{maasMachinePoolRaw(nil)}
	oldRaw := baseMaasRaw(maasCloudConfigRawMap(nil), pools)
	newRaw := baseMaasRaw(maasCloudConfigRawMap(map[string]interface{}{"enable_lxd_vm": true}), pools)

	d := buildMaasUpdateResourceData(t, oldRaw, newRaw, maasCloudConfigUID)
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterMaasUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterMaasUpdate_CloudConfigUpdateError exercises the
// UpdateCloudConfigMaas API-error branch inside the cloud_config HasChange
// block.
func TestResourceClusterMaasUpdate_CloudConfigUpdateError(t *testing.T) {
	pools := []interface{}{maasMachinePoolRaw(nil)}
	oldRaw := baseMaasRaw(maasCloudConfigRawMap(nil), pools)
	newRaw := baseMaasRaw(maasCloudConfigRawMap(map[string]interface{}{"enable_lxd_vm": true}), pools)

	d := buildMaasUpdateResourceData(t, oldRaw, newRaw, routes.MaasCloudConfigUpdateErrorUID)
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterMaasUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterMaasUpdate_MachinePoolDiff drives, in a single Update
// call, all three machine_pool branches gated by d.HasChange("machine_pool"):
// create (new-pool), update (pool-to-change, hash differs on count), removal
// (pool-to-remove), and the unchanged no-op branch (pool-keep). All pools
// here are workers so no control-plane placement edge cases are involved.
//
// Note: schema.InternalMap(res.Schema).Data(oldState, diff) reconstructs a
// TypeSet's new value including a zero-valued phantom entry for any removed
// set element (a known quirk of rebuilding sets from a raw diff outside a
// live apply). resourceClusterMaasUpdate already defends against this via
// its `if machinePoolResource["name"].(string) != ""` guard, so we don't
// assert exact Set lengths here — only that the Update call completes
// successfully having exercised every create/update/delete/unchanged branch.
func TestResourceClusterMaasUpdate_MachinePoolDiff(t *testing.T) {
	worker := func(overrides map[string]interface{}) map[string]interface{} {
		base := map[string]interface{}{
			"control_plane":           false,
			"control_plane_as_worker": false,
		}
		for k, v := range overrides {
			base[k] = v
		}
		return maasMachinePoolRaw(base)
	}

	poolKeep := worker(map[string]interface{}{"name": "pool-keep"})
	poolChangeOld := worker(map[string]interface{}{"name": "pool-to-change", "count": 2})
	poolChangeNew := worker(map[string]interface{}{"name": "pool-to-change", "count": 5})
	poolRemove := worker(map[string]interface{}{"name": "pool-to-remove"})
	newPool := worker(map[string]interface{}{"name": "new-pool"})

	cc := maasCloudConfigRawMap(nil)
	oldRaw := baseMaasRaw(cc, []interface{}{poolKeep, poolChangeOld, poolRemove})
	newRaw := baseMaasRaw(cc, []interface{}{poolKeep, poolChangeNew, newPool})

	d := buildMaasUpdateResourceData(t, oldRaw, newRaw, maasCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterMaasUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterMaasUpdate_ValidateOverrideScalingError exercises the
// validateOverrideScaling error branch: a new machine pool declares
// update_strategy=OverrideScaling without an override_scaling block.
func TestResourceClusterMaasUpdate_ValidateOverrideScalingError(t *testing.T) {
	poolKeep := maasMachinePoolRaw(nil)
	invalidPool := maasMachinePoolRaw(map[string]interface{}{
		"name":                    "invalid-pool",
		"control_plane":           false,
		"control_plane_as_worker": false,
		"update_strategy":         "OverrideScaling",
	})

	cc := maasCloudConfigRawMap(nil)
	oldRaw := baseMaasRaw(cc, []interface{}{poolKeep})
	newRaw := baseMaasRaw(cc, []interface{}{poolKeep, invalidPool})

	d := buildMaasUpdateResourceData(t, oldRaw, newRaw, maasCloudConfigUID)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterMaasUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterMaasUpdate_HyperShiftValidationError exercises the
// validateHyperShiftMaasConfig error branch inside resourceClusterMaasUpdate:
// an openshift-type hyper_shift_config without host_cluster_uid must be
// rejected before any cloud_config/machine_pool work happens.
func TestResourceClusterMaasUpdate_HyperShiftValidationError(t *testing.T) {
	base := baseMaasUpdateAttrs()
	base["hyper_shift_config.#"] = "1"
	base["hyper_shift_config.0.cluster_deployment_type"] = "openshift"
	base["hyper_shift_config.0.host_cluster_uid"] = ""

	d := buildUpdateResourceData(resourceClusterMaas(), maasClusterID, base,
		simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterMaasUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}
