package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// (test file header) MaaS CRUD.
// Pattern shared with the GCP/AWS/vSphere wave2 tests.

const (
	maasCloudConfigUID  = "test-cloud-config-id"
	maasCloudAccountUID = "test-maas-account-id-1"
	maasClusterProfile  = "cluster-profile-import-1"
	maasClusterID       = "test-maas-cluster-id"
)

func defaultMaasMachinePool(overrides map[string]interface{}) map[string]interface{} {
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
		"azs": schema.NewSet(schema.HashString, []interface{}{"us-east-1a"}),
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

func maasMachinePoolSet(pools ...map[string]interface{}) *schema.Set {
	items := make([]interface{}, len(pools))
	for i, p := range pools {
		items[i] = p
	}
	return schema.NewSet(resourceMachinePoolMaasHash, items)
}

func prepareMaasClusterResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceClusterMaas().TestResourceData()
	require.NoError(t, d.Set("name", "test-maas-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_account_id", maasCloudAccountUID))
	require.NoError(t, d.Set("cloud_config_id", maasCloudConfigUID))
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": maasClusterProfile},
	}))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"domain":        "test-domain",
			"enable_lxd_vm": false,
		},
	}))
	require.NoError(t, d.Set("machine_pool", maasMachinePoolSet(defaultMaasMachinePool(nil))))
	return d
}

func TestResourceClusterMaasReadWithMock(t *testing.T) {
	d := prepareMaasClusterResourceData(t)
	d.SetId(maasClusterID)

	diags := resourceClusterMaasRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, maasCloudConfigUID, d.Get("cloud_config_id"))
	assert.Equal(t, maasCloudAccountUID, d.Get("cloud_account_id"))

	pools := d.Get("machine_pool").(*schema.Set)
	assert.Greater(t, pools.Len(), 0)
}

func TestResourceClusterMaasCreateWithMock(t *testing.T) {
	d := prepareMaasClusterResourceData(t)
	require.NoError(t, d.Set("tags", []interface{}{"skip_completion"}))

	diags := resourceClusterMaasCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, maasClusterID, d.Id())
}

func TestResourceClusterMaasUpdateMachinePoolWithMock(t *testing.T) {
	d := prepareMaasClusterResourceData(t)
	d.SetId(maasClusterID)

	oldPool := defaultMaasMachinePool(nil)
	newPool := defaultMaasMachinePool(map[string]interface{}{"count": 3})

	require.NoError(t, d.Set("machine_pool", maasMachinePoolSet(oldPool)))
	require.NoError(t, d.Set("machine_pool", maasMachinePoolSet(newPool)))

	diags := resourceClusterMaasUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceClusterMaasUpdateMachinePoolAddWithMock(t *testing.T) {
	d := prepareMaasClusterResourceData(t)
	d.SetId(maasClusterID)

	cpPool := defaultMaasMachinePool(nil)
	workerPool := defaultMaasMachinePool(map[string]interface{}{
		"name":                    "worker-pool",
		"count":                   2,
		"control_plane":           false,
		"control_plane_as_worker": false,
	})

	require.NoError(t, d.Set("machine_pool", maasMachinePoolSet(cpPool)))
	require.NoError(t, d.Set("machine_pool", maasMachinePoolSet(cpPool, workerPool)))

	diags := resourceClusterMaasUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceClusterMaasUpdateMachinePoolDeleteWithMock(t *testing.T) {
	d := prepareMaasClusterResourceData(t)
	d.SetId(maasClusterID)

	cpPool := defaultMaasMachinePool(nil)
	workerPool := defaultMaasMachinePool(map[string]interface{}{
		"name":                    "worker-pool",
		"count":                   2,
		"control_plane":           false,
		"control_plane_as_worker": false,
	})

	require.NoError(t, d.Set("machine_pool", maasMachinePoolSet(cpPool, workerPool)))
	require.NoError(t, d.Set("machine_pool", maasMachinePoolSet(cpPool)))

	diags := resourceClusterMaasUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterMaasStateUpgradeV2 pins the standard cluster_profile
// TypeList → TypeSet upgrader.
func TestResourceClusterMaasStateUpgradeV2(t *testing.T) {
	ctx := context.Background()

	t.Run("list passes through", func(t *testing.T) {
		state := map[string]interface{}{
			"cluster_profile": []interface{}{
				map[string]interface{}{"id": "cp-1"},
			},
		}
		got, err := resourceClusterMaasStateUpgradeV2(ctx, state, nil)
		require.NoError(t, err)
		cp, ok := got["cluster_profile"].([]interface{})
		require.True(t, ok)
		assert.Len(t, cp, 1)
	})

	t.Run("missing key is a no-op", func(t *testing.T) {
		got, err := resourceClusterMaasStateUpgradeV2(ctx, map[string]interface{}{}, nil)
		require.NoError(t, err)
		_, exists := got["cluster_profile"]
		assert.False(t, exists)
	})
}
