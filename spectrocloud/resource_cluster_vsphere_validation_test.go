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

// TestSortPlacementStructs_TieBreakLevels exercises every comparator branch
// in sortPlacementStructs: sorting only kicks in at the first field where two
// entries differ, so each subtest below keeps every earlier field tied to
// force the comparator down to a specific field.
func TestSortPlacementStructs_TieBreakLevels(t *testing.T) {
	t.Run("differ at datastore (cluster tied)", func(t *testing.T) {
		in := []interface{}{
			map[string]interface{}{"cluster": "same", "datastore": "d2", "resource_pool": "r1", "network": "n1"},
			map[string]interface{}{"cluster": "same", "datastore": "d1", "resource_pool": "r1", "network": "n1"},
		}
		sortPlacementStructs(in)
		assert.Equal(t, "d1", in[0].(map[string]interface{})["datastore"])
		assert.Equal(t, "d2", in[1].(map[string]interface{})["datastore"])
	})

	t.Run("differ at resource_pool (cluster+datastore tied)", func(t *testing.T) {
		in := []interface{}{
			map[string]interface{}{"cluster": "same", "datastore": "same", "resource_pool": "r2", "network": "n1"},
			map[string]interface{}{"cluster": "same", "datastore": "same", "resource_pool": "r1", "network": "n1"},
		}
		sortPlacementStructs(in)
		assert.Equal(t, "r1", in[0].(map[string]interface{})["resource_pool"])
		assert.Equal(t, "r2", in[1].(map[string]interface{})["resource_pool"])
	})

	t.Run("differ only at network (cluster+datastore+resource_pool tied)", func(t *testing.T) {
		in := []interface{}{
			map[string]interface{}{"cluster": "same", "datastore": "same", "resource_pool": "same", "network": "n2"},
			map[string]interface{}{"cluster": "same", "datastore": "same", "resource_pool": "same", "network": "n1"},
		}
		sortPlacementStructs(in)
		assert.Equal(t, "n1", in[0].(map[string]interface{})["network"])
		assert.Equal(t, "n2", in[1].(map[string]interface{})["network"])
	})

	t.Run("fully tied entries stay stable", func(t *testing.T) {
		in := []interface{}{
			map[string]interface{}{"cluster": "same", "datastore": "same", "resource_pool": "same", "network": "same"},
			map[string]interface{}{"cluster": "same", "datastore": "same", "resource_pool": "same", "network": "same"},
		}
		assert.NotPanics(t, func() { sortPlacementStructs(in) })
	})
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
		newPlacement := append([]interface{}{}, basePlacement...)
		newPlacement = append(newPlacement, map[string]interface{}{
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

	t.Run("datastore value change is rejected", func(t *testing.T) {
		updatedPlacement := []interface{}{
			map[string]interface{}{
				"cluster":           "cluster-a",
				"datastore":         "ds-b",
				"resource_pool":     "rp-a",
				"network":           "net-a",
				"static_ip_pool_id": "",
			},
		}

		changed, err := ValidateMachinePoolChange(makeSet(basePlacement), makeSet(updatedPlacement))
		require.Error(t, err)
		assert.True(t, changed)
		assert.Contains(t, err.Error(), "DataStore")
	})

	t.Run("resource_pool value change is rejected", func(t *testing.T) {
		updatedPlacement := []interface{}{
			map[string]interface{}{
				"cluster":           "cluster-a",
				"datastore":         "ds-a",
				"resource_pool":     "rp-b",
				"network":           "net-a",
				"static_ip_pool_id": "",
			},
		}

		changed, err := ValidateMachinePoolChange(makeSet(basePlacement), makeSet(updatedPlacement))
		require.Error(t, err)
		assert.True(t, changed)
		assert.Contains(t, err.Error(), "resource_pool")
	})

	t.Run("network value change is rejected", func(t *testing.T) {
		updatedPlacement := []interface{}{
			map[string]interface{}{
				"cluster":           "cluster-a",
				"datastore":         "ds-a",
				"resource_pool":     "rp-a",
				"network":           "net-b",
				"static_ip_pool_id": "",
			},
		}

		changed, err := ValidateMachinePoolChange(makeSet(basePlacement), makeSet(updatedPlacement))
		require.Error(t, err)
		assert.True(t, changed)
		assert.Contains(t, err.Error(), "Network")
	})

	t.Run("no control-plane pool on either side is a no-op", func(t *testing.T) {
		workerOnly := schema.NewSet(resourceMachinePoolVsphereHash, []interface{}{
			map[string]interface{}{
				"name":          "worker",
				"control_plane": false,
				"placement":     basePlacement,
			},
		})

		changed, err := ValidateMachinePoolChange(workerOnly, workerOnly)
		require.NoError(t, err)
		assert.False(t, changed)
	})
}
