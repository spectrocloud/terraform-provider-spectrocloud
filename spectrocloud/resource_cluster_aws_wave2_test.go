package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	awsCloudConfigUID  = "test-cloud-config-id"
	awsCloudAccountUID = "test-aws-account-id-1"
	awsClusterProfile  = "cluster-profile-import-1"
)

func defaultAwsMachinePool(overrides map[string]interface{}) map[string]interface{} {
	pool := map[string]interface{}{
		"name":                    "cp-pool",
		"instance_type":           "t3.large",
		"count":                   1,
		"control_plane":           true,
		"control_plane_as_worker": true,
		"capacity_type":           "",
		"max_price":               "",
		"azs":                     schema.NewSet(schema.HashString, []interface{}{}),
		"az_subnets":              map[string]interface{}{},
	}
	for k, v := range overrides {
		pool[k] = v
	}
	return pool
}

func awsMachinePoolSet(pools ...map[string]interface{}) *schema.Set {
	items := make([]interface{}, len(pools))
	for i, p := range pools {
		items[i] = p
	}
	return schema.NewSet(resourceMachinePoolAwsHash, items)
}

func prepareAwsClusterResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceClusterAws().TestResourceData()
	require.NoError(t, d.Set("name", "test-aws-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_account_id", awsCloudAccountUID))
	require.NoError(t, d.Set("cloud_config_id", awsCloudConfigUID))
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": awsClusterProfile},
	}))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"region":                      "us-east-1",
			"vpc_id":                      "vpc-test123",
			"ssh_key_name":                "",
			"control_plane_lb":            "",
			"override_cluster_api_config": "",
		},
	}))
	require.NoError(t, d.Set("machine_pool", awsMachinePoolSet(defaultAwsMachinePool(nil))))
	return d
}

func TestResourceClusterAwsReadWithMock(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	d.SetId("test-cluster-id")

	diags := resourceClusterAwsRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, awsCloudConfigUID, d.Get("cloud_config_id"))
	assert.Equal(t, awsCloudAccountUID, d.Get("cloud_account_id"))

	pools := d.Get("machine_pool").(*schema.Set)
	assert.Greater(t, pools.Len(), 0)
}

func TestFlattenCloudConfigAwsWithMock(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	c := mustUnitClient(t, false)

	diags := flattenCloudConfigAws(awsCloudConfigUID, d, c)
	assert.False(t, diags.HasError())
	assert.Equal(t, awsCloudAccountUID, d.Get("cloud_account_id"))

	cfg := d.Get("cloud_config").([]interface{})
	require.Len(t, cfg, 1)
	assert.Equal(t, "us-east-1", cfg[0].(map[string]interface{})["region"])
}

func TestResourceClusterAwsCreateWithMock(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	require.NoError(t, d.Set("tags", []interface{}{"skip_completion"}))

	diags := resourceClusterAwsCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-aws-cluster-id", d.Id())
}

func TestValidateSystemRepaveApprovalWithMock(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	d.SetId("test-cluster-id")
	c := mustUnitClient(t, false)

	cluster, err := c.GetCluster(d.Id())
	require.NoError(t, err)
	require.NotNil(t, cluster)
	require.NotNil(t, cluster.Status)
	require.NotNil(t, cluster.Status.Repave, "mock cluster should include repave status for AWS update tests")

	assert.NoError(t, validateSystemRepaveApproval(d, c))
}

func TestResourceClusterAwsUpdateCloudConfigWithMock(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	d.SetId("test-cluster-id")

	oldCfg := map[string]interface{}{
		"region":                      "us-east-1",
		"vpc_id":                      "vpc-old",
		"ssh_key_name":                "",
		"control_plane_lb":            "",
		"override_cluster_api_config": "",
	}
	newCfg := map[string]interface{}{
		"region":                      "us-east-1",
		"vpc_id":                      "vpc-test123",
		"ssh_key_name":                "new-key",
		"control_plane_lb":            "",
		"override_cluster_api_config": "",
	}
	require.NoError(t, d.Set("cloud_config", []interface{}{oldCfg}))
	require.NoError(t, d.Set("cloud_config", []interface{}{newCfg}))

	diags := resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterAwsUpdateMachinePoolWithMock(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	d.SetId("test-cluster-id")

	oldPool := defaultAwsMachinePool(nil)
	newPool := defaultAwsMachinePool(map[string]interface{}{"count": 3})

	require.NoError(t, d.Set("machine_pool", awsMachinePoolSet(oldPool)))
	require.NoError(t, d.Set("machine_pool", awsMachinePoolSet(newPool)))

	diags := resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterAwsUpdateMachinePoolAddWithMock(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	d.SetId("test-cluster-id")

	cpPool := defaultAwsMachinePool(nil)
	workerPool := defaultAwsMachinePool(map[string]interface{}{
		"name":                    "worker-pool",
		"count":                   2,
		"control_plane":           false,
		"control_plane_as_worker": false,
	})

	require.NoError(t, d.Set("machine_pool", awsMachinePoolSet(cpPool)))
	require.NoError(t, d.Set("machine_pool", awsMachinePoolSet(cpPool, workerPool)))

	diags := resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterAwsUpdateMachinePoolDeleteWithMock(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	d.SetId("test-cluster-id")

	cpPool := defaultAwsMachinePool(nil)
	workerPool := defaultAwsMachinePool(map[string]interface{}{
		"name":                    "worker-pool",
		"count":                   2,
		"control_plane":           false,
		"control_plane_as_worker": false,
	})

	require.NoError(t, d.Set("machine_pool", awsMachinePoolSet(cpPool, workerPool)))
	require.NoError(t, d.Set("machine_pool", awsMachinePoolSet(cpPool)))

	diags := resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterAwsUpdateClusterProfileWithMock(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	d.SetId("test-cluster-id")
	require.NoError(t, d.Set("tags", []interface{}{"skip_apply"}))

	setChangedClusterProfiles(t, d,
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-2"}},
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-1"}},
	)

	diags := resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}
