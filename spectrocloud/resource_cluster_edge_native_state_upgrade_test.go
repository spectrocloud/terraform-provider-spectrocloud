package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceClusterEdgeNativeStateUpgradeV3(t *testing.T) {
	raw := map[string]interface{}{
		"cluster_profile": []interface{}{
			map[string]interface{}{"id": "profile-1"},
			map[string]interface{}{"id": "profile-2"},
		},
	}

	out, err := resourceClusterEdgeNativeStateUpgradeV3(context.Background(), raw, nil)
	require.NoError(t, err)

	profiles, ok := out["cluster_profile"].([]interface{})
	require.True(t, ok)
	assert.Len(t, profiles, 2)
}

func TestResourceClusterEdgeNativeStateUpgradeV2(t *testing.T) {
	machinePoolSet := schema.NewSet(resourceMachinePoolEdgeNativeHash, []interface{}{
		map[string]interface{}{
			"name": "pool-1",
			"edge_host": []interface{}{
				map[string]interface{}{"host_uid": "uid-1"},
			},
		},
	})
	raw := map[string]interface{}{
		"machine_pool": machinePoolSet,
	}

	out, err := resourceClusterEdgeNativeStateUpgradeV2(context.Background(), raw, nil)
	require.NoError(t, err)

	machinePool, ok := out["machine_pool"].([]interface{})
	require.True(t, ok)
	assert.Len(t, machinePool, 1)
}
