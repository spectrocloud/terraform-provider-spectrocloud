package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

// TestResourceClusterAzureRead_ServerError exercises resourceClusterAzureRead's
// non-404 GetCluster error branch (handleReadError → diag.FromErr).
func TestResourceClusterAzureRead_ServerError(t *testing.T) {
	d := prepareAzureClusterResourceData(t)
	d.SetId("cluster-uid-server-error")

	diags := resourceClusterAzureRead(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterAzureRead_ClusterDeleted exercises the "cluster == nil"
// branch: GetCluster returns nil when Status.State == "Deleted", and Read
// must clear the resource ID so Terraform recreates it.
func TestResourceClusterAzureRead_ClusterDeleted(t *testing.T) {
	d := prepareAzureClusterResourceData(t)
	d.SetId("cluster-uid-deleted-state")

	diags := resourceClusterAzureRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id())
}

// TestFlattenCloudConfigAzure_GetCloudConfigError exercises
// flattenCloudConfigAzure's GetCloudConfigAzure error branch.
func TestFlattenCloudConfigAzure_GetCloudConfigError(t *testing.T) {
	d := prepareAzureClusterResourceData(t)
	c := mustUnitClient(t, false)

	diags := flattenCloudConfigAzure(routes.AzureCloudConfigGetErrorUID, d, c)
	assert.True(t, diags.HasError())
}

// TestToAzureCluster_MachinePoolError exercises toAzureCluster's
// machinePoolConfigs error branch: a control-plane pool with a non-zero
// node_repave_interval is rejected by toMachinePoolAzure via
// ValidationNodeRepaveIntervalForControlPlane.
func TestToAzureCluster_MachinePoolError(t *testing.T) {
	d := prepareAzureClusterResourceData(t)
	d.SetId(azureClusterID)
	badPool := defaultAzureMachinePool(map[string]interface{}{
		"control_plane":        true,
		"node_repave_interval": 45,
	})
	require.NoError(t, d.Set("machine_pool", azureMachinePoolSet(badPool)))

	c := mustUnitClient(t, false)
	_, err := toAzureCluster(c, d)
	assert.Error(t, err)
}

// TestToMachinePoolAzure_ControlPlaneRepaveError exercises the control-plane
// node_repave_interval validation error branch directly.
func TestToMachinePoolAzure_ControlPlaneRepaveError(t *testing.T) {
	pool := defaultAzureMachinePool(map[string]interface{}{
		"control_plane":        true,
		"node_repave_interval": 15,
	})

	_, err := toMachinePoolAzure(pool)
	assert.Error(t, err)
}

// TestChooseIPMethod covers both branches of chooseIPMethod directly —
// TestToStaticPlacement (resource_cluster_azure_test.go) only ever exercises
// the empty-IP ("Dynamic") branch via the no-private_api_server code path.
func TestChooseIPMethod(t *testing.T) {
	assert.Equal(t, "Dynamic", *chooseIPMethod(""))
	assert.Equal(t, "Static", *chooseIPMethod("10.0.0.5"))
}

// TestResourceClusterAzureSchemaClosures exercises the DiffSuppressFunc and
// DefaultFunc closures embedded in the machine_pool schema (os_type,  azs).
// A bare TestResourceData()/Get() call never invokes these — they're
// wired into Terraform's diff/default-computation pipeline, not the
// Set/Get path — so they sit at 0% unless called directly (mirrors the
// resourceClusterCustomCloud / resourceClusterEks fix applied earlier this
// session).
func TestResourceClusterAzureSchemaClosures(t *testing.T) {
	mpSchema := resourceClusterAzure().Schema["machine_pool"].Elem.(*schema.Resource).Schema

	osType := mpSchema["os_type"]
	require.NotNil(t, osType.DiffSuppressFunc)
	assert.False(t, osType.DiffSuppressFunc("os_type", "Linux", "Windows", nil))

	azs := mpSchema["azs"]
	require.NotNil(t, azs.DefaultFunc)
	val, err := azs.DefaultFunc()
	require.NoError(t, err)
	assert.Equal(t, []string{""}, val)
}
