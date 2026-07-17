package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	aksCloudConfigUID  = "test-cloud-config-id"
	aksCloudAccountUID = "test-azure-account-id-1"
	aksClusterProfile  = "cluster-profile-import-1"
	aksClusterID       = "test-aks-cluster-id"
)

func defaultAksMachinePool(overrides map[string]interface{}) map[string]interface{} {
	pool := map[string]interface{}{
		"name":                 "worker-pool",
		"count":                2,
		"instance_type":        "Standard_D2s_v3",
		"disk_size_gb":         128,
		"storage_account_type": "Premium_LRS",
		"is_system_node_pool":  false,
	}
	for k, v := range overrides {
		pool[k] = v
	}
	return pool
}

func aksMachinePoolSet(pools ...map[string]interface{}) *schema.Set {
	items := make([]interface{}, len(pools))
	for i, p := range pools {
		items[i] = p
	}
	return schema.NewSet(resourceMachinePoolAksHash, items)
}

func prepareAksClusterResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceClusterAks().TestResourceData()
	require.NoError(t, d.Set("name", "test-aks-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_account_id", aksCloudAccountUID))
	require.NoError(t, d.Set("cloud_config_id", aksCloudConfigUID))
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": aksClusterProfile},
	}))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"subscription_id": "test-subscription-id",
			"resource_group":  "test-rg",
			"region":          "eastus",
		},
	}))
	require.NoError(t, d.Set("machine_pool", aksMachinePoolSet(defaultAksMachinePool(nil))))
	return d
}

func TestResourceClusterAksReadWithMock(t *testing.T) {
	d := prepareAksClusterResourceData(t)
	d.SetId(aksClusterID)

	diags := resourceClusterAksRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, aksCloudConfigUID, d.Get("cloud_config_id"))
	assert.Equal(t, aksCloudAccountUID, d.Get("cloud_account_id"))

	pools, ok := d.Get("machine_pool").(*schema.Set)
	require.True(t, ok)
	assert.Equal(t, 1, pools.Len())
}

func TestResourceClusterAksCreateWithMock(t *testing.T) {
	d := prepareAksClusterResourceData(t)
	require.NoError(t, d.Set("tags", []interface{}{"skip_completion"}))

	diags := resourceClusterAksCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, aksClusterID, d.Id())
}

func TestResourceClusterAksUpdateCloudConfigWithMock(t *testing.T) {
	d := prepareAksClusterResourceData(t)
	d.SetId(aksClusterID)

	oldCfg := map[string]interface{}{
		"subscription_id": "test-subscription-id",
		"resource_group":  "old-rg",
		"region":          "eastus",
	}
	newCfg := map[string]interface{}{
		"subscription_id": "test-subscription-id",
		"resource_group":  "test-rg",
		"region":          "eastus",
	}
	require.NoError(t, d.Set("cloud_config", []interface{}{oldCfg}))
	require.NoError(t, d.Set("cloud_config", []interface{}{newCfg}))

	diags := resourceClusterAksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterAksUpdateMachinePoolWithMock(t *testing.T) {
	d := prepareAksClusterResourceData(t)
	d.SetId(aksClusterID)

	oldPool := defaultAksMachinePool(nil)
	newPool := defaultAksMachinePool(map[string]interface{}{"count": 4})

	require.NoError(t, d.Set("machine_pool", aksMachinePoolSet(oldPool)))
	require.NoError(t, d.Set("machine_pool", aksMachinePoolSet(newPool)))

	diags := resourceClusterAksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterAksUpdateMachinePoolAddWithMock(t *testing.T) {
	d := prepareAksClusterResourceData(t)
	d.SetId(aksClusterID)

	pool1 := defaultAksMachinePool(nil)
	pool2 := defaultAksMachinePool(map[string]interface{}{
		"name":  "worker-pool-2",
		"count": 1,
	})

	require.NoError(t, d.Set("machine_pool", aksMachinePoolSet(pool1)))
	require.NoError(t, d.Set("machine_pool", aksMachinePoolSet(pool1, pool2)))

	diags := resourceClusterAksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterAksUpdateMachinePoolDeleteWithMock(t *testing.T) {
	d := prepareAksClusterResourceData(t)
	d.SetId(aksClusterID)

	pool1 := defaultAksMachinePool(nil)
	pool2 := defaultAksMachinePool(map[string]interface{}{
		"name":  "worker-pool-2",
		"count": 1,
	})

	require.NoError(t, d.Set("machine_pool", aksMachinePoolSet(pool1, pool2)))
	require.NoError(t, d.Set("machine_pool", aksMachinePoolSet(pool1)))

	diags := resourceClusterAksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterAksUpdateClusterProfileWithMock(t *testing.T) {
	d := prepareAksClusterResourceData(t)
	d.SetId(aksClusterID)
	require.NoError(t, d.Set("tags", []interface{}{"skip_apply"}))

	setChangedClusterProfiles(t, d,
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-2"}},
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-1"}},
	)

	diags := resourceClusterAksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}
