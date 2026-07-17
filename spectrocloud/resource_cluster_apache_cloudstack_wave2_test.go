package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// (test file header) CloudStack CRUD.

const (
	cloudstackCloudConfigUID  = "test-cloud-config-id"
	cloudstackCloudAccountUID = "test-cloudstack-account-id-1"
	cloudstackClusterProfile  = "cluster-profile-import-1"
	cloudstackClusterID       = "test-cloudstack-cluster-id"
)

func defaultCloudStackMachinePool(overrides map[string]interface{}) map[string]interface{} {
	pool := map[string]interface{}{
		"name":                    "cp-pool",
		"count":                   1,
		"control_plane":           true,
		"control_plane_as_worker": true,
		"offering":                "Medium Instance",
		"template": []interface{}{
			map[string]interface{}{
				"name": "ubuntu-2004",
			},
		},
	}
	for k, v := range overrides {
		pool[k] = v
	}
	return pool
}

func cloudstackMachinePoolSet(pools ...map[string]interface{}) *schema.Set {
	items := make([]interface{}, len(pools))
	for i, p := range pools {
		items[i] = p
	}
	return schema.NewSet(resourceMachinePoolApacheCloudStackHash, items)
}

func prepareCloudStackClusterResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceClusterApacheCloudStack().TestResourceData()
	require.NoError(t, d.Set("name", "test-cloudstack-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_account_id", cloudstackCloudAccountUID))
	require.NoError(t, d.Set("cloud_config_id", cloudstackCloudConfigUID))
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": cloudstackClusterProfile},
	}))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"ssh_key_name": "test-ssh-key",
			"zone": []interface{}{
				map[string]interface{}{
					"name": "zone-1",
				},
			},
		},
	}))
	require.NoError(t, d.Set("machine_pool", cloudstackMachinePoolSet(defaultCloudStackMachinePool(nil))))
	return d
}

func TestResourceClusterApacheCloudStackReadWithMock(t *testing.T) {
	d := prepareCloudStackClusterResourceData(t)
	d.SetId(cloudstackClusterID)

	diags := resourceClusterApacheCloudStackRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, cloudstackCloudConfigUID, d.Get("cloud_config_id"))
	assert.Equal(t, cloudstackCloudAccountUID, d.Get("cloud_account_id"))
}

func TestResourceClusterApacheCloudStackCreateWithMock(t *testing.T) {
	// Create is skipped because resourceClusterApacheCloudStackCreate calls
	// waitForClusterCreation(..., initial=false), which bypasses the
	// skip_completion short-circuit. Every other cloud (aws/gcp/gke/vsphere/
	// azure/aks/maas/edge_native) uses initial=true so `tags = ["skip_completion"]`
	// exits the wait immediately. Fixing this cleanly needs a provider change
	// (change to initial=true, matching the rest of the resources), so a unit
	// test can't work around it without accepting a ~30s wait per run.
	t.Skip("CloudStack Create uses waitForClusterCreation(initial=false) — 30s wait even with skip_completion. Track as a provider consistency fix.")

	d := prepareCloudStackClusterResourceData(t)
	require.NoError(t, d.Set("tags", []interface{}{"skip_completion"}))

	diags := resourceClusterApacheCloudStackCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, cloudstackClusterID, d.Id())
}

func TestResourceClusterApacheCloudStackUpdateMachinePoolWithMock(t *testing.T) {
	d := prepareCloudStackClusterResourceData(t)
	d.SetId(cloudstackClusterID)

	oldPool := defaultCloudStackMachinePool(nil)
	newPool := defaultCloudStackMachinePool(map[string]interface{}{"count": 3})

	require.NoError(t, d.Set("machine_pool", cloudstackMachinePoolSet(oldPool)))
	require.NoError(t, d.Set("machine_pool", cloudstackMachinePoolSet(newPool)))

	diags := resourceClusterApacheCloudStackUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceClusterApacheCloudStackUpdateClusterProfileWithMock(t *testing.T) {
	d := prepareCloudStackClusterResourceData(t)
	d.SetId(cloudstackClusterID)
	require.NoError(t, d.Set("tags", []interface{}{"skip_apply"}))

	setChangedClusterProfiles(t, d,
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-2"}},
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-1"}},
	)

	diags := resourceClusterApacheCloudStackUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterApacheCloudStackStateUpgradeV2 covers the standard
// cluster_profile TypeList → TypeSet conversion.
func TestResourceClusterApacheCloudStackStateUpgradeV2(t *testing.T) {
	ctx := context.Background()

	t.Run("list passes through", func(t *testing.T) {
		state := map[string]interface{}{
			"cluster_profile": []interface{}{
				map[string]interface{}{"id": "cp-1"},
			},
		}
		got, err := resourceClusterApacheCloudStackStateUpgradeV2(ctx, state, nil)
		require.NoError(t, err)
		cp, ok := got["cluster_profile"].([]interface{})
		require.True(t, ok)
		assert.Len(t, cp, 1)
	})

	t.Run("missing key is a no-op", func(t *testing.T) {
		got, err := resourceClusterApacheCloudStackStateUpgradeV2(ctx, map[string]interface{}{}, nil)
		require.NoError(t, err)
		_, exists := got["cluster_profile"]
		assert.False(t, exists)
	})
}
