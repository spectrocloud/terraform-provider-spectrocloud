package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	eksCloudConfigUID  = "test-cloud-config-id"
	eksCloudAccountUID = "test-aws-account-id-1"
	eksClusterProfile  = "cluster-profile-import-1"
	eksClusterID       = "test-eks-cluster-id"
)

// PLT-2435: adds a `dedicate_node_pool_for_system_pods` attribute to EKS and
// GKE machine pools, wired straight through to the SDK's
// V1MachinePoolConfigEntity.DedicateNodePoolForSystemPods field. AKS already
// covers this via its existing is_system_node_pool attribute, so it needs no
// changes here.

func TestToMachinePoolEks_DedicateNodePoolForSystemPods(t *testing.T) {
	t.Run("true is passed through", func(t *testing.T) {
		m := map[string]interface{}{
			"name":                               "system-pool",
			"count":                              1,
			"disk_size_gb":                       20,
			"dedicate_node_pool_for_system_pods": true,
			"az_subnets":                         map[string]interface{}{},
		}
		got := toMachinePoolEks(m)
		require.NotNil(t, got.PoolConfig)
		assert.True(t, got.PoolConfig.DedicateNodePoolForSystemPods)
	})

	t.Run("absent key defaults to false rather than panicking", func(t *testing.T) {
		m := map[string]interface{}{
			"name":         "regular-pool",
			"count":        1,
			"disk_size_gb": 20,
			"az_subnets":   map[string]interface{}{},
		}
		got := toMachinePoolEks(m)
		require.NotNil(t, got.PoolConfig)
		assert.False(t, got.PoolConfig.DedicateNodePoolForSystemPods)
	})
}

func TestFlattenMachinePoolConfigsEks_DedicateNodePoolForSystemPods(t *testing.T) {
	machinePools := []*models.V1EksMachinePoolConfig{
		{
			Name:                          "system-pool",
			DedicateNodePoolForSystemPods: true,
		},
	}
	result := flattenMachinePoolConfigsEks(machinePools)
	require.Len(t, result, 1)
	oi := result[0].(map[string]interface{})
	assert.Equal(t, true, oi["dedicate_node_pool_for_system_pods"])
}

func defaultEksMachinePool(overrides map[string]interface{}) map[string]interface{} {
	pool := map[string]interface{}{
		"name":            "worker-pool",
		"count":           2,
		"instance_type":   "m5.large",
		"ami_type":        "AL2023_x86_64_STANDARD",
		"disk_size_gb":    100,
		"capacity_type":   "",
		"max_price":       "",
		"min":             0,
		"max":             0,
		"update_strategy": "RollingUpdateScaleOut",
		"az_subnets": map[string]interface{}{
			"us-east-1a": "subnet-worker",
		},
	}
	for k, v := range overrides {
		pool[k] = v
	}
	return pool
}

func eksMachinePoolSet(pools ...map[string]interface{}) *schema.Set {
	items := make([]interface{}, len(pools))
	for i, p := range pools {
		items[i] = p
	}
	return schema.NewSet(resourceMachinePoolEksHash, items)
}

func prepareEksClusterResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()
	emptyCIDRs := schema.NewSet(schema.HashString, []interface{}{})
	d := resourceClusterEks().TestResourceData()
	require.NoError(t, d.Set("name", "test-eks-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_account_id", eksCloudAccountUID))
	require.NoError(t, d.Set("cloud_config_id", eksCloudConfigUID))
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": eksClusterProfile},
	}))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"region":               "us-east-1",
			"vpc_id":               "vpc-test123",
			"ssh_key_name":         "test-key",
			"endpoint_access":      "public",
			"public_access_cidrs":  emptyCIDRs,
			"private_access_cidrs": emptyCIDRs,
			"az_subnets": map[string]interface{}{
				"us-east-1a": "subnet-cp",
			},
		},
	}))
	require.NoError(t, d.Set("fargate_profile", []interface{}{}))
	require.NoError(t, d.Set("machine_pool", eksMachinePoolSet(defaultEksMachinePool(nil))))
	return d
}

func TestResourceClusterEksReadWithMock(t *testing.T) {
	d := prepareEksClusterResourceData(t)
	d.SetId(eksClusterID)

	diags := resourceClusterEksRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, eksCloudConfigUID, d.Get("cloud_config_id"))
	assert.Equal(t, eksCloudAccountUID, d.Get("cloud_account_id"))

	pools, ok := d.Get("machine_pool").(*schema.Set)
	require.True(t, ok)
	assert.Equal(t, 1, pools.Len())
}

func TestFlattenClusterConfigsEKSWithMock(t *testing.T) {
	c := mustUnitClient(t, false)

	config, err := c.GetCloudConfigEks(eksCloudConfigUID)
	require.NoError(t, err)
	require.NotNil(t, config)

	flattened := flattenClusterConfigsEKS(config).([]interface{})
	require.Len(t, flattened, 1)
	assert.Equal(t, "us-east-1", flattened[0].(map[string]interface{})["region"])
	assert.Equal(t, "vpc-test123", flattened[0].(map[string]interface{})["vpc_id"])
}

func TestResourceClusterEksCreateWithMock(t *testing.T) {
	d := prepareEksClusterResourceData(t)
	require.NoError(t, d.Set("tags", []interface{}{"skip_completion"}))

	diags := resourceClusterEksCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, eksClusterID, d.Id())
}

func TestResourceClusterEksUpdateCloudConfigWithMock(t *testing.T) {
	d := prepareEksClusterResourceData(t)
	d.SetId(eksClusterID)

	oldCfg := map[string]interface{}{
		"region":               "us-east-1",
		"vpc_id":               "vpc-old",
		"ssh_key_name":         "test-key",
		"endpoint_access":      "public",
		"public_access_cidrs":  schema.NewSet(schema.HashString, []interface{}{}),
		"private_access_cidrs": schema.NewSet(schema.HashString, []interface{}{}),
	}
	newCfg := map[string]interface{}{
		"region":               "us-east-1",
		"vpc_id":               "vpc-test123",
		"ssh_key_name":         "new-key",
		"endpoint_access":      "public",
		"public_access_cidrs":  schema.NewSet(schema.HashString, []interface{}{}),
		"private_access_cidrs": schema.NewSet(schema.HashString, []interface{}{}),
	}
	require.NoError(t, d.Set("cloud_config", []interface{}{oldCfg}))
	require.NoError(t, d.Set("cloud_config", []interface{}{newCfg}))

	diags := resourceClusterEksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterEksUpdateMachinePoolWithMock(t *testing.T) {
	d := prepareEksClusterResourceData(t)
	d.SetId(eksClusterID)

	oldPool := defaultEksMachinePool(nil)
	newPool := defaultEksMachinePool(map[string]interface{}{"count": 4})

	require.NoError(t, d.Set("machine_pool", eksMachinePoolSet(oldPool)))
	require.NoError(t, d.Set("machine_pool", eksMachinePoolSet(newPool)))

	diags := resourceClusterEksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterEksUpdateMachinePoolAddWithMock(t *testing.T) {
	d := prepareEksClusterResourceData(t)
	d.SetId(eksClusterID)

	pool1 := defaultEksMachinePool(nil)
	pool2 := defaultEksMachinePool(map[string]interface{}{
		"name":  "worker-pool-2",
		"count": 1,
	})

	require.NoError(t, d.Set("machine_pool", eksMachinePoolSet(pool1)))
	require.NoError(t, d.Set("machine_pool", eksMachinePoolSet(pool1, pool2)))

	diags := resourceClusterEksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterEksUpdateMachinePoolDeleteWithMock(t *testing.T) {
	d := prepareEksClusterResourceData(t)
	d.SetId(eksClusterID)

	pool1 := defaultEksMachinePool(nil)
	pool2 := defaultEksMachinePool(map[string]interface{}{
		"name":  "worker-pool-2",
		"count": 1,
	})

	require.NoError(t, d.Set("machine_pool", eksMachinePoolSet(pool1, pool2)))
	require.NoError(t, d.Set("machine_pool", eksMachinePoolSet(pool1)))

	diags := resourceClusterEksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterEksUpdateClusterProfileWithMock(t *testing.T) {
	d := prepareEksClusterResourceData(t)
	d.SetId(eksClusterID)
	require.NoError(t, d.Set("tags", []interface{}{"skip_apply"}))

	setChangedClusterProfiles(t, d,
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-2"}},
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-1"}},
	)

	diags := resourceClusterEksUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}
