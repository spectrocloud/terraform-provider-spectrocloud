package spectrocloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

////
// All state-upgrader functions have the same shape:
//   (ctx, rawState, meta) → (rawState, error)
// They flip a field type (usually TypeList → TypeSet) and return the
// (possibly-unchanged) state map. Each test drives it with a populated
// map + a "missing" map to cover both branches.

// upgradeFn is the shared signature of every state upgrader.
type upgradeFn func(context.Context, map[string]interface{}, interface{}) (map[string]interface{}, error)

// runUpgraderPair drives an upgrader through two invocations: one with
// the expected input key populated (as an []interface{}) so the
// list-to-set branch fires, and one with the key absent so the
// "skipping" branch is exercised.
func runUpgraderPair(t *testing.T, name string, up upgradeFn, targetKey string) {
	t.Helper()
	t.Run(name+"/populated", func(t *testing.T) {
		state := map[string]interface{}{
			targetKey: []interface{}{
				map[string]interface{}{"name": "test"},
			},
		}
		got, err := up(context.Background(), state, nil)
		assert.NoError(t, err)
		assert.NotNil(t, got)
	})
	t.Run(name+"/missing_key", func(t *testing.T) {
		got, err := up(context.Background(), map[string]interface{}{}, nil)
		assert.NoError(t, err)
		assert.NotNil(t, got)
	})
	t.Run(name+"/wrong_type", func(t *testing.T) {
		// Not a []interface{} → "skipping conversion" branch.
		state := map[string]interface{}{targetKey: "not-a-list"}
		got, err := up(context.Background(), state, nil)
		assert.NoError(t, err)
		assert.NotNil(t, got)
	})
}

func TestStateUpgraders(t *testing.T) {
	cases := []struct {
		name string
		fn   upgradeFn
		key  string
	}{
		{"EksV2", resourceClusterEksStateUpgradeV2, "machine_pool"},
		{"GcpV2", resourceClusterGcpStateUpgradeV2, "machine_pool"},
		{"GkeV1", resourceClusterGkeStateUpgradeV1, "machine_pool"},
		{"GkeV2", resourceClusterGkeStateUpgradeV2, "machine_pool"},
		{"ClusterGroupV2", resourceClusterGroupStateUpgradeV2, "clusters"},
		{"FilterV2", resourceFilterStateUpgradeV2, "filter"},
		{"WorkspaceV2", resourceWorkspaceStateUpgradeV2, "clusters"},
	}
	for _, c := range cases {
		c := c
		runUpgraderPair(t, c.name, c.fn, c.key)
	}
}
