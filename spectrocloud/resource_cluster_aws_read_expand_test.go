package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

// TestResourceClusterAwsRead_ServerError exercises resourceClusterAwsRead's
// non-404 GetCluster error branch (handleReadError → diag.FromErr).
func TestResourceClusterAwsRead_ServerError(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	d.SetId("cluster-uid-server-error")

	diags := resourceClusterAwsRead(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterAwsRead_ClusterDeleted exercises the "cluster == nil"
// branch: GetCluster returns nil when Status.State == "Deleted", and Read
// must clear the resource ID so Terraform recreates it.
func TestResourceClusterAwsRead_ClusterDeleted(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	d.SetId("cluster-uid-deleted-state")

	diags := resourceClusterAwsRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id())
}

// TestResourceClusterAwsRead_TagsMap exercises the tags_map branch: when
// tags_map is already populated on the resource, Read flattens
// cluster.Metadata.Labels into tags_map (and blanks out tags).
func TestResourceClusterAwsRead_TagsMap(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	d.SetId("test-cluster-id")
	require.NoError(t, d.Set("tags_map", map[string]interface{}{"env": "test"}))

	diags := resourceClusterAwsRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)

	tagsMap := d.Get("tags_map").(map[string]interface{})
	assert.Equal(t, "test", tagsMap["env"])
}

// TestFlattenCloudConfigAws_GetCloudConfigError exercises
// flattenCloudConfigAws's GetCloudConfigAws error branch.
func TestFlattenCloudConfigAws_GetCloudConfigError(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	c := mustUnitClient(t, false)

	diags := flattenCloudConfigAws(routes.AwsCloudConfigGetErrorUID, d, c)
	assert.True(t, diags.HasError())
}

// TestToAwsCluster_TagsMap exercises toAwsCluster's tags_map branch:
// when tags_map is set, cluster.Metadata.Labels is populated from it
// instead of the default "tags" attribute.
func TestToAwsCluster_TagsMap(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	d.SetId("test-cluster-id")
	require.NoError(t, d.Set("tags_map", map[string]interface{}{"team": "platform"}))

	c := mustUnitClient(t, false)
	cluster, err := toAwsCluster(c, d)
	require.NoError(t, err)
	assert.Equal(t, "platform", cluster.Metadata.Labels["team"])
}

// TestToAwsCluster_MachinePoolError exercises toAwsCluster's
// machinePoolConfigs error branch: a control-plane pool with a non-zero
// node_repave_interval is rejected by toMachinePoolAws via
// ValidationNodeRepaveIntervalForControlPlane.
func TestToAwsCluster_MachinePoolError(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	d.SetId("test-cluster-id")
	badPool := defaultAwsMachinePool(map[string]interface{}{
		"control_plane":        true,
		"node_repave_interval": 45,
	})
	require.NoError(t, d.Set("machine_pool", awsMachinePoolSet(badPool)))

	c := mustUnitClient(t, false)
	_, err := toAwsCluster(c, d)
	assert.Error(t, err)
}

// TestToMachinePoolAws_HostResourceGroup exercises the
// capacity_type=="host-resource-group" branch (host_resource_group_arn +
// license_configuration_arns) inside toMachinePoolAws.
func TestToMachinePoolAws_HostResourceGroup(t *testing.T) {
	pool := defaultAwsMachinePool(map[string]interface{}{
		"name":                       "worker-pool",
		"control_plane":              false,
		"control_plane_as_worker":    false,
		"disk_size_gb":               65,
		"capacity_type":              "host-resource-group",
		"host_resource_group_arn":    "arn:aws:license-manager:us-east-1:123456789012:host-resource-group:test",
		"license_configuration_arns": schema.NewSet(schema.HashString, []interface{}{"arn:aws:license-manager:us-east-1:123456789012:license-configuration:test"}),
		"additional_security_groups": schema.NewSet(schema.HashString, []interface{}{"sg-12345"}),
	})

	mp, err := toMachinePoolAws(pool, "vpc-test123")
	require.NoError(t, err)
	assert.Equal(t, "arn:aws:license-manager:us-east-1:123456789012:host-resource-group:test", mp.CloudConfig.HostResourceGroupArn)
	require.Len(t, mp.CloudConfig.LicenseConfigurationArns, 1)
	require.Len(t, mp.CloudConfig.AdditionalSecurityGroups, 1)
}

// TestToMachinePoolAws_Spot exercises the capacity_type=="spot" branch
// (SpotMarketOptions.MaxPrice) inside toMachinePoolAws.
func TestToMachinePoolAws_Spot(t *testing.T) {
	pool := defaultAwsMachinePool(map[string]interface{}{
		"name":                    "spot-pool",
		"control_plane":           false,
		"control_plane_as_worker": false,
		"disk_size_gb":            65,
		"capacity_type":           "spot",
		"max_price":               "0.05",
	})

	mp, err := toMachinePoolAws(pool, "vpc-test123")
	require.NoError(t, err)
	require.NotNil(t, mp.CloudConfig.SpotMarketOptions)
	assert.Equal(t, "0.05", mp.CloudConfig.SpotMarketOptions.MaxPrice)
}

// TestToMachinePoolAws_ControlPlaneRepaveError exercises the control-plane
// node_repave_interval validation error branch directly.
func TestToMachinePoolAws_ControlPlaneRepaveError(t *testing.T) {
	pool := defaultAwsMachinePool(map[string]interface{}{
		"control_plane":        true,
		"disk_size_gb":         65,
		"node_repave_interval": 10,
	})

	_, err := toMachinePoolAws(pool, "vpc-test123")
	assert.Error(t, err)
}
