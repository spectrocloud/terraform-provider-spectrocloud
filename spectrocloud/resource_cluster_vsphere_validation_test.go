package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSortPlacementStructs(t *testing.T) {
	in := []interface{}{
		map[string]interface{}{"cluster": "b", "datastore": "d2", "resource_pool": "r2", "network": "n2"},
		map[string]interface{}{"cluster": "a", "datastore": "d1", "resource_pool": "r1", "network": "n1"},
	}

	sortPlacementStructs(in)

	first := in[0].(map[string]interface{})
	assert.Equal(t, "a", first["cluster"])
	assert.Equal(t, "d1", first["datastore"])
}

func TestValidateMachinePoolChange(t *testing.T) {
	makeSet := func(placement []interface{}) *schema.Set {
		return schema.NewSet(resourceMachinePoolVsphereHash, []interface{}{
			map[string]interface{}{
				"name":          "cp",
				"control_plane": true,
				"placement":     placement,
			},
		})
	}

	basePlacement := []interface{}{
		map[string]interface{}{
			"cluster":           "cluster-a",
			"datastore":         "ds-a",
			"resource_pool":     "rp-a",
			"network":           "net-a",
			"static_ip_pool_id": "",
		},
	}

	t.Run("same placement no change", func(t *testing.T) {
		changed, err := ValidateMachinePoolChange(makeSet(basePlacement), makeSet(basePlacement))
		require.NoError(t, err)
		assert.False(t, changed)
	})

	t.Run("placement length change is rejected", func(t *testing.T) {
		newPlacement := append(basePlacement, map[string]interface{}{
			"cluster":           "cluster-b",
			"datastore":         "ds-b",
			"resource_pool":     "rp-b",
			"network":           "net-b",
			"static_ip_pool_id": "",
		})

		changed, err := ValidateMachinePoolChange(makeSet(basePlacement), makeSet(newPlacement))
		require.Error(t, err)
		assert.True(t, changed)
		assert.Contains(t, err.Error(), "adding/removing placement")
	})

	t.Run("cluster value change is rejected", func(t *testing.T) {
		updatedPlacement := []interface{}{
			map[string]interface{}{
				"cluster":           "cluster-b",
				"datastore":         "ds-a",
				"resource_pool":     "rp-a",
				"network":           "net-a",
				"static_ip_pool_id": "",
			},
		}

		changed, err := ValidateMachinePoolChange(makeSet(basePlacement), makeSet(updatedPlacement))
		require.Error(t, err)
		assert.True(t, changed)
		assert.Contains(t, err.Error(), "ComputeCluster")
	})
}
