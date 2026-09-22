package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
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

func TestToMachinePoolGke_DedicateNodePoolForSystemPods(t *testing.T) {
	t.Run("true is passed through", func(t *testing.T) {
		m := map[string]interface{}{
			"name":                               "system-pool",
			"instance_type":                      "e2-standard-2",
			"disk_size_gb":                       60,
			"count":                              1,
			"dedicate_node_pool_for_system_pods": true,
		}
		got, err := toMachinePoolGke(m)
		require.NoError(t, err)
		require.NotNil(t, got.PoolConfig)
		assert.True(t, got.PoolConfig.DedicateNodePoolForSystemPods)
	})

	t.Run("absent key defaults to false rather than panicking", func(t *testing.T) {
		m := map[string]interface{}{
			"name":          "regular-pool",
			"instance_type": "e2-standard-2",
			"disk_size_gb":  60,
			"count":         1,
		}
		got, err := toMachinePoolGke(m)
		require.NoError(t, err)
		require.NotNil(t, got.PoolConfig)
		assert.False(t, got.PoolConfig.DedicateNodePoolForSystemPods)
	})
}

func TestFlattenMachinePoolConfigsGke_DedicateNodePoolForSystemPods(t *testing.T) {
	machinePools := []*models.V1GcpMachinePoolConfig{
		{
			Name:                          "system-pool",
			InstanceType:                  types.Ptr("e2-standard-2"),
			DedicateNodePoolForSystemPods: true,
		},
	}
	result := flattenMachinePoolConfigsGke(machinePools)
	require.Len(t, result, 1)
	oi := result[0].(map[string]interface{})
	assert.Equal(t, true, oi["dedicate_node_pool_for_system_pods"])
}
