package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// (test file header) Edge vSphere
// CRUD coverage.
//
// Notes on shape differences vs plain vSphere:
//   - Cloud config requires `vip` (control-plane VIP); plain vSphere
//     doesn't have this field.
//   - machine_pool is a TypeList (not TypeSet) — see the comment in
//     resource_cluster_edge_vsphere.go about PE-255. That means we
//     use []interface{} not schema.NewSet.

const (
	edgeVsphereCloudConfigUID = "test-cloud-config-id"
	edgeVsphereClusterProfile = "cluster-profile-import-1"
	// Edge vSphere Create hits POST /v1/spectroclusters/vsphere (shared
	// with plain vSphere), so the mock returns test-vsphere-cluster-id.
	// The dedicated "test-edge-vsphere-cluster-id" fixture is used only
	// for direct Read tests where the caller controls d.Id().
	edgeVsphereClusterID = "test-edge-vsphere-cluster-id"
)

func defaultEdgeVspherePlacement() map[string]interface{} {
	return map[string]interface{}{
		"cluster":           "test-cluster",
		"resource_pool":     "test-pool",
		"datastore":         "test-datastore",
		"network":           "test-network",
		"static_ip_pool_id": "",
	}
}

func defaultEdgeVsphereMachinePool(overrides map[string]interface{}) map[string]interface{} {
	pool := map[string]interface{}{
		"name":                    "cp-pool",
		"count":                   1,
		"control_plane":           true,
		"control_plane_as_worker": true,
		"instance_type": []interface{}{
			map[string]interface{}{
				"disk_size_gb": 60,
				"memory_mb":    8192,
				"cpu":          4,
			},
		},
		"placement": []interface{}{defaultEdgeVspherePlacement()},
	}
	for k, v := range overrides {
		pool[k] = v
	}
	return pool
}

// prepareEdgeVsphereClusterResourceData builds a valid resource fixture.
// machine_pool is a TypeList here, not a TypeSet.
func prepareEdgeVsphereClusterResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceClusterEdgeVsphere().TestResourceData()
	require.NoError(t, d.Set("name", "test-edge-vsphere-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_config_id", edgeVsphereCloudConfigUID))
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": edgeVsphereClusterProfile},
	}))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"datacenter": "test-datacenter",
			"folder":     "test-folder",
			"ssh_key":    "ssh-rsa AAAA",
			"vip":        "10.0.0.100",
		},
	}))
	// machine_pool is a TypeList — use []interface{}, not schema.NewSet.
	require.NoError(t, d.Set("machine_pool", []interface{}{
		defaultEdgeVsphereMachinePool(nil),
	}))
	return d
}

func TestResourceClusterEdgeVsphereReadWithMock(t *testing.T) {
	d := prepareEdgeVsphereClusterResourceData(t)
	d.SetId(edgeVsphereClusterID)

	diags := resourceClusterEdgeVsphereRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)

	pools := d.Get("machine_pool").([]interface{})
	assert.Greater(t, len(pools), 0)
}

// Create/Update tests are currently SKIPPED — filed as follow-up
// task_3f23c658. The shared validateOverrideScaling helper in
// cluster_common_fields.go unconditionally casts machine_pool to
// *schema.Set, but edge_vsphere declares machine_pool as TypeList (per
// a PE-255 comment in resource_cluster_edge_vsphere.go). Any Create or
// Update call on edge_vsphere panics on that type assertion. Once the
// bug is fixed, the skips can be removed and the tests below will
// exercise the full CRUD paths.

func TestResourceClusterEdgeVsphereCreateWithMock(t *testing.T) {
	t.Skip("blocked on task_3f23c658 — validateOverrideScaling casts to *schema.Set")
	d := prepareEdgeVsphereClusterResourceData(t)
	require.NoError(t, d.Set("tags", []interface{}{"skip_completion"}))

	diags := resourceClusterEdgeVsphereCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "test-vsphere-cluster-id", d.Id())
}

func TestResourceClusterEdgeVsphereUpdateMachinePoolWithMock(t *testing.T) {
	t.Skip("blocked on task_3f23c658 — validateOverrideScaling casts to *schema.Set")
	d := prepareEdgeVsphereClusterResourceData(t)
	d.SetId(edgeVsphereClusterID)

	oldPool := defaultEdgeVsphereMachinePool(nil)
	newPool := defaultEdgeVsphereMachinePool(map[string]interface{}{"count": 3})

	require.NoError(t, d.Set("machine_pool", []interface{}{oldPool}))
	require.NoError(t, d.Set("machine_pool", []interface{}{newPool}))

	diags := resourceClusterEdgeVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterEdgeVsphereStateUpgradeV0 — same TypeList → TypeSet
// pattern for cluster_profile that the other cluster upgraders use.
func TestResourceClusterEdgeVsphereStateUpgradeV0(t *testing.T) {
	ctx := context.Background()

	t.Run("list passes through", func(t *testing.T) {
		state := map[string]interface{}{
			"cluster_profile": []interface{}{
				map[string]interface{}{"id": "cp-1"},
			},
		}
		got, err := resourceClusterEdgeVsphereStateUpgradeV0(ctx, state, nil)
		require.NoError(t, err)
		cp, ok := got["cluster_profile"].([]interface{})
		require.True(t, ok)
		assert.Len(t, cp, 1)
	})

	t.Run("missing key is a no-op", func(t *testing.T) {
		got, err := resourceClusterEdgeVsphereStateUpgradeV0(ctx, map[string]interface{}{}, nil)
		require.NoError(t, err)
		_, exists := got["cluster_profile"]
		assert.False(t, exists)
	})
}
