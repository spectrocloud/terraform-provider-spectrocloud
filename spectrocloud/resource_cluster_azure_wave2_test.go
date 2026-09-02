package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	azureCloudConfigUID  = "test-cloud-config-id"
	azureCloudAccountUID = "test-azure-account-id-1"
	azureClusterProfile  = "cluster-profile-import-1"
	azureClusterID       = "test-azure-cluster-id"
)

func defaultAzureMachinePool(overrides map[string]interface{}) map[string]interface{} {
	azs := schema.NewSet(schema.HashString, []interface{}{"eastus-1"})
	pool := map[string]interface{}{
		"name":                    "worker-pool",
		"count":                   2,
		"control_plane":           false,
		"control_plane_as_worker": false,
		"instance_type":           "Standard_D2s_v3",
		"os_type":                 "Linux",
		"is_system_node_pool":     false,
		"azs":                     azs,
		"disk": []interface{}{
			map[string]interface{}{
				"size_gb": 128,
				"type":    "Premium_LRS",
			},
		},
	}
	for k, v := range overrides {
		pool[k] = v
	}
	return pool
}

func azureMachinePoolSet(pools ...map[string]interface{}) *schema.Set {
	items := make([]interface{}, len(pools))
	for i, p := range pools {
		items[i] = p
	}
	return schema.NewSet(resourceMachinePoolAzureHash, items)
}

func prepareAzureClusterResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceClusterAzure().TestResourceData()
	require.NoError(t, d.Set("name", "test-azure-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_account_id", azureCloudAccountUID))
	require.NoError(t, d.Set("cloud_config_id", azureCloudConfigUID))
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": azureClusterProfile},
	}))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"subscription_id": "test-subscription-id",
			"resource_group":  "test-rg",
			"region":          "eastus",
			"ssh_key":         "test-ssh-key",
		},
	}))
	require.NoError(t, d.Set("machine_pool", azureMachinePoolSet(defaultAzureMachinePool(nil))))
	return d
}

func TestResourceClusterAzureReadWithMock(t *testing.T) {
	d := prepareAzureClusterResourceData(t)
	d.SetId(azureClusterID)

	diags := resourceClusterAzureRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, azureCloudConfigUID, d.Get("cloud_config_id"))
	assert.Equal(t, azureCloudAccountUID, d.Get("cloud_account_id"))

	pools, ok := d.Get("machine_pool").(*schema.Set)
	require.True(t, ok)
	assert.Greater(t, pools.Len(), 0)
}

func TestFlattenCloudConfigAzureWithMock(t *testing.T) {
	d := prepareAzureClusterResourceData(t)
	c := mustUnitClient(t, false)

	diags := flattenCloudConfigAzure(azureCloudConfigUID, d, c)
	assert.False(t, diags.HasError())
	assert.Equal(t, azureCloudAccountUID, d.Get("cloud_account_id"))
}

func TestResourceClusterAzureCreateWithMock(t *testing.T) {
	d := prepareAzureClusterResourceData(t)
	require.NoError(t, d.Set("tags", []interface{}{"skip_completion"}))

	diags := resourceClusterAzureCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, azureClusterID, d.Id())
}

func TestResourceClusterAzureUpdateMachinePoolWithMock(t *testing.T) {
	d := prepareAzureClusterResourceData(t)
	d.SetId(azureClusterID)

	oldPool := defaultAzureMachinePool(nil)
	newPool := defaultAzureMachinePool(map[string]interface{}{"count": 4})

	require.NoError(t, d.Set("machine_pool", azureMachinePoolSet(oldPool)))
	require.NoError(t, d.Set("machine_pool", azureMachinePoolSet(newPool)))

	diags := resourceClusterAzureUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterAzureUpdateMachinePoolAddWithMock(t *testing.T) {
	d := prepareAzureClusterResourceData(t)
	d.SetId(azureClusterID)

	pool1 := defaultAzureMachinePool(nil)
	pool2 := defaultAzureMachinePool(map[string]interface{}{
		"name":  "worker-pool-2",
		"count": 1,
	})

	require.NoError(t, d.Set("machine_pool", azureMachinePoolSet(pool1)))
	require.NoError(t, d.Set("machine_pool", azureMachinePoolSet(pool1, pool2)))

	diags := resourceClusterAzureUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterAzureUpdateMachinePoolDeleteWithMock(t *testing.T) {
	d := prepareAzureClusterResourceData(t)
	d.SetId(azureClusterID)

	pool1 := defaultAzureMachinePool(nil)
	pool2 := defaultAzureMachinePool(map[string]interface{}{
		"name":  "worker-pool-2",
		"count": 1,
	})

	require.NoError(t, d.Set("machine_pool", azureMachinePoolSet(pool1, pool2)))
	require.NoError(t, d.Set("machine_pool", azureMachinePoolSet(pool1)))

	diags := resourceClusterAzureUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterAzureUpdateClusterProfileWithMock(t *testing.T) {
	d := prepareAzureClusterResourceData(t)
	d.SetId(azureClusterID)
	require.NoError(t, d.Set("tags", []interface{}{"skip_apply"}))

	setChangedClusterProfiles(t, d,
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-2"}},
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-1"}},
	)

	diags := resourceClusterAzureUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}
