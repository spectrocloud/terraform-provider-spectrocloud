package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

// ---------------------------------------------------------------------
// resourceClusterEdgeVsphereCreate
//
// validateOverrideScaling (cluster_common_fields.go) unconditionally
// casts machine_pool to *schema.Set, but edge_vsphere declares
// machine_pool as TypeList (see the PE-255 comment in
// resource_cluster_edge_vsphere.go) — d.GetOk("machine_pool") returns
// ok=true (and thus panics on that cast) whenever machine_pool is
// non-empty. d.GetOk treats an empty list as the schema zero value
// (ok=false), so an EMPTY machine_pool sidesteps the panic entirely and
// lets the rest of Create run for real — this is the only way to drive
// Create through to CreateClusterEdgeVsphere/waitForClusterCreation/Read
// without hitting the known blocker (task_3f23c658, see
// resource_cluster_edge_vsphere_wave2_test.go).
// ---------------------------------------------------------------------

func TestResourceClusterEdgeVsphereCreate_EmptyMachinePool(t *testing.T) {
	d := prepareEdgeVsphereClusterResourceData(t)
	require.NoError(t, d.Set("machine_pool", []interface{}{}))
	require.NoError(t, d.Set("tags", []interface{}{"skip_completion"}))

	diags := resourceClusterEdgeVsphereCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "test-vsphere-cluster-id", d.Id())
}

func TestToEdgeVsphereClusterProfileResolutionError(t *testing.T) {
	d := resourceClusterEdgeVsphere().TestResourceData()
	d.SetId("cluster-uid-not-found")
	require.NoError(t, d.Set("name", "test-edge-vsphere-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"datacenter": "test-datacenter",
			"folder":     "test-folder",
			"ssh_key":    "ssh-rsa AAAA",
			"vip":        "10.0.0.100",
		},
	}))
	require.NoError(t, d.Set("machine_pool", []interface{}{}))

	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	_, err := toEdgeVsphereCluster(c, d)
	assert.Error(t, err)
}

func TestToEdgeVsphereClusterMachinePoolError(t *testing.T) {
	d := prepareEdgeVsphereClusterResourceData(t)
	// control-plane pool with a non-zero node_repave_interval is invalid
	// (ValidationNodeRepaveIntervalForControlPlane), so
	// toMachinePoolEdgeVsphere errors and toEdgeVsphereCluster must
	// propagate it.
	pool := defaultEdgeVsphereMachinePool(map[string]interface{}{"node_repave_interval": 45})
	require.NoError(t, d.Set("machine_pool", []interface{}{pool}))

	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	_, err := toEdgeVsphereCluster(c, d)
	assert.Error(t, err)
}

func TestToMachinePoolEdgeVsphereOverrideKubeadmWorker(t *testing.T) {
	placement := defaultEdgeVspherePlacement()
	placement["id"] = ""
	pool := defaultEdgeVsphereMachinePool(map[string]interface{}{
		"control_plane":                  false,
		"control_plane_as_worker":        false,
		"override_kubeadm_configuration": "kubeletExtraArgs:\n  v: \"4\"",
		"placement":                      []interface{}{placement},
	})
	mp, err := toMachinePoolEdgeVsphere(pool)
	require.NoError(t, err)
	assert.Equal(t, "kubeletExtraArgs:\n  v: \"4\"", mp.PoolConfig.OverrideKubeadmConfiguration)
}

// ---------------------------------------------------------------------
// flattenMachinePoolConfigsEdgeVsphere — update_strategy / override_scaling
// and override_kubeadm_configuration (worker pools only) branches.
// ---------------------------------------------------------------------

func TestFlattenMachinePoolConfigsEdgeVsphereUpdateStrategyAndOverrideKubeadm(t *testing.T) {
	worker := false
	result := flattenMachinePoolConfigsEdgeVsphere([]*models.V1VsphereMachinePoolConfig{
		{
			Name:                         "worker-pool",
			IsControlPlane:               &worker,
			OverrideKubeadmConfiguration: "kubeletExtraArgs:\n  v: \"4\"",
			UpdateStrategy: &models.V1UpdateStrategy{
				Type:           "OverrideScaling",
				MaxSurge:       "1",
				MaxUnavailable: "0",
			},
		},
	})
	require.Len(t, result, 1)
	oi := result[0].(map[string]interface{})
	assert.Equal(t, "OverrideScaling", oi["update_strategy"])
	assert.Equal(t, "kubeletExtraArgs:\n  v: \"4\"", oi["override_kubeadm_configuration"])
	require.Contains(t, oi, "override_scaling")
	overrideScaling := oi["override_scaling"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "1", overrideScaling["max_surge"])
	assert.Equal(t, "0", overrideScaling["max_unavailable"])
}

// ---------------------------------------------------------------------
// resourceClusterEdgeVsphereRead — cluster-gone / API-error branches.
// ---------------------------------------------------------------------

func TestResourceClusterEdgeVsphereReadClusterNotFound(t *testing.T) {
	d := prepareEdgeVsphereClusterResourceData(t)
	d.SetId("cluster-uid-not-found")

	diags := resourceClusterEdgeVsphereRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "", d.Id())
}

func TestResourceClusterEdgeVsphereReadServerError(t *testing.T) {
	d := prepareEdgeVsphereClusterResourceData(t)
	d.SetId("cluster-uid-server-error")

	diags := resourceClusterEdgeVsphereRead(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// ---------------------------------------------------------------------
// flattenCloudConfigEdgeVsphere — GetCloudConfigVsphere API-error branch.
// ---------------------------------------------------------------------

func TestFlattenCloudConfigEdgeVsphereAPIError(t *testing.T) {
	d := resourceClusterEdgeVsphere().TestResourceData()
	d.SetId("test-cluster-uid")
	require.NoError(t, d.Set("context", "project"))

	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	diags := flattenCloudConfigEdgeVsphere(routes.VsphereCloudConfigErrorUID, d, c)
	assert.True(t, diags.HasError())
}

// ---------------------------------------------------------------------
// resourceClusterEdgeVsphereUpdate
// ---------------------------------------------------------------------

func TestResourceClusterEdgeVsphereUpdate_RepaveApprovalError(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterEdgeVsphere(), "cluster-uid-server-error",
		baseEdgeVsphereUpdateAttrs(), simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterEdgeVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

func TestResourceClusterEdgeVsphereUpdate_CloudConfigError(t *testing.T) {
	attrs := baseEdgeVsphereUpdateAttrs()
	attrs["cloud_config_id"] = routes.VsphereCloudConfigErrorUID
	d := buildUpdateResourceData(resourceClusterEdgeVsphere(), "test-edge-vsphere-cluster-id",
		attrs, simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterEdgeVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// buildEdgeVsphereMachinePoolChangeResourceData builds a real
// InstanceState + config diff (see
// resource_cluster_apache_cloudstack_test.go's
// buildMachinePoolChangeResourceData for the rationale) so
// d.HasChange("machine_pool")/d.GetChange("machine_pool") behave the way
// Terraform's own apply pipeline would produce them. machine_pool here is
// a TypeList (not TypeSet, per the PE-255 comment in
// resource_cluster_edge_vsphere.go), so pool ordering is preserved as
// given.
func buildEdgeVsphereMachinePoolChangeResourceData(t *testing.T, oldPools, newPools []interface{}) *schema.ResourceData {
	t.Helper()
	res := resourceClusterEdgeVsphere()

	base := map[string]interface{}{
		"name":          "test-edge-vsphere-cluster",
		"context":       "project",
		"edge_host_uid": "test-edge-host-uid",
		"cloud_config": []interface{}{
			map[string]interface{}{
				"datacenter": "test-datacenter",
				"folder":     "test-folder",
				"ssh_key":    "ssh-rsa AAAA",
				"vip":        "10.0.0.100",
			},
		},
	}

	oldRaw := map[string]interface{}{}
	for k, v := range base {
		oldRaw[k] = v
	}
	oldRaw["machine_pool"] = oldPools
	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId(edgeVsphereClusterID)
	require.NoError(t, oldRD.Set("cloud_config_id", edgeVsphereCloudConfigUID))
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newRaw := map[string]interface{}{}
	for k, v := range base {
		newRaw[k] = v
	}
	newRaw["machine_pool"] = newPools
	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)
	require.NotNil(t, diff)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId(edgeVsphereClusterID)
	return finalRD
}

// TestResourceClusterEdgeVsphereUpdate_RemoveAllMachinePools drives the
// "pool removed entirely" branch of the machine_pool diff loop. This is
// the only reachable branch of that loop through the public Update entry
// point: a non-empty NEW machine_pool makes d.GetOk("machine_pool")
// return ok=true inside validateOverrideScaling, which then panics on
// its *schema.Set cast (task_3f23c658) — so the create/hash-changed
// branches (lines ~599-623) can't be exercised this way. An empty NEW
// list keeps GetOk's ok=false, skipping that call safely while still
// making d.HasChange("machine_pool") true (old was non-empty).
func TestResourceClusterEdgeVsphereUpdate_RemoveAllMachinePools(t *testing.T) {
	oldPools := []interface{}{defaultEdgeVsphereMachinePool(nil)}
	newPools := []interface{}{}

	d := buildEdgeVsphereMachinePoolChangeResourceData(t, oldPools, newPools)

	old, cur := d.GetChange("machine_pool")
	require.Len(t, old.([]interface{}), 1)
	require.Len(t, cur.([]interface{}), 0)

	diags := resourceClusterEdgeVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}
