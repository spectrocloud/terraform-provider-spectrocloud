package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// (test file header) GKE CRUD coverage.
// Mirrors resource_cluster_gcp_wave2_test.go with two differences that
// come from the resource itself:
//   1. cloud_config only has project+region (no `network`).
//   2. machine pools don't declare control_plane — GKE manages the
//      control plane for you.

const (
	gkeCloudConfigUID  = "test-cloud-config-id"
	gkeCloudAccountUID = "test-gcp-account-id-1"
	gkeClusterProfile  = "cluster-profile-import-1"
)

func defaultGkeMachinePool(overrides map[string]interface{}) map[string]interface{} {
	pool := map[string]interface{}{
		"name":            "worker-pool",
		"instance_type":   "n1-standard-2",
		"count":           2,
		"disk_size_gb":    60,
		"update_strategy": "RollingUpdateScaleOut",
	}
	for k, v := range overrides {
		pool[k] = v
	}
	return pool
}

func gkeMachinePoolSet(pools ...map[string]interface{}) *schema.Set {
	items := make([]interface{}, len(pools))
	for i, p := range pools {
		items[i] = p
	}
	return schema.NewSet(resourceMachinePoolGkeHash, items)
}

func prepareGkeClusterResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceClusterGke().TestResourceData()
	require.NoError(t, d.Set("name", "test-gke-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_account_id", gkeCloudAccountUID))
	require.NoError(t, d.Set("cloud_config_id", gkeCloudConfigUID))
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": gkeClusterProfile},
	}))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"project": "test-gcp-project",
			"region":  "us-central1",
		},
	}))
	require.NoError(t, d.Set("machine_pool", gkeMachinePoolSet(defaultGkeMachinePool(nil))))
	return d
}

func TestResourceClusterGkeReadWithMock(t *testing.T) {
	d := prepareGkeClusterResourceData(t)
	d.SetId("test-gke-cluster-id")

	diags := resourceClusterGkeRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, gkeCloudConfigUID, d.Get("cloud_config_id"))
	assert.Equal(t, gkeCloudAccountUID, d.Get("cloud_account_id"))

	pools := d.Get("machine_pool").(*schema.Set)
	assert.Greater(t, pools.Len(), 0)
}

func TestFlattenCloudConfigGkeWithMock(t *testing.T) {
	d := prepareGkeClusterResourceData(t)
	c := mustUnitClient(t, false)

	diags := flattenCloudConfigGke(gkeCloudConfigUID, d, c)
	assert.False(t, diags.HasError())
	assert.Equal(t, gkeCloudAccountUID, d.Get("cloud_account_id"))

	cfg := d.Get("cloud_config").([]interface{})
	require.Len(t, cfg, 1)
	assert.Equal(t, "us-central1", cfg[0].(map[string]interface{})["region"])
	assert.Equal(t, "test-gcp-project", cfg[0].(map[string]interface{})["project"])
}

func TestResourceClusterGkeCreateWithMock(t *testing.T) {
	d := prepareGkeClusterResourceData(t)
	require.NoError(t, d.Set("tags", []interface{}{"skip_completion"}))

	diags := resourceClusterGkeCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-gke-cluster-id", d.Id())
}

func TestResourceClusterGkeUpdateMachinePoolWithMock(t *testing.T) {
	d := prepareGkeClusterResourceData(t)
	d.SetId("test-gke-cluster-id")

	oldPool := defaultGkeMachinePool(nil)
	newPool := defaultGkeMachinePool(map[string]interface{}{"count": 4})

	require.NoError(t, d.Set("machine_pool", gkeMachinePoolSet(oldPool)))
	require.NoError(t, d.Set("machine_pool", gkeMachinePoolSet(newPool)))

	diags := resourceClusterGkeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterGkeUpdateMachinePoolAddWithMock(t *testing.T) {
	d := prepareGkeClusterResourceData(t)
	d.SetId("test-gke-cluster-id")

	pool1 := defaultGkeMachinePool(nil)
	pool2 := defaultGkeMachinePool(map[string]interface{}{
		"name":  "worker-pool-2",
		"count": 3,
	})

	require.NoError(t, d.Set("machine_pool", gkeMachinePoolSet(pool1)))
	require.NoError(t, d.Set("machine_pool", gkeMachinePoolSet(pool1, pool2)))

	diags := resourceClusterGkeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterGkeUpdateMachinePoolDeleteWithMock(t *testing.T) {
	d := prepareGkeClusterResourceData(t)
	d.SetId("test-gke-cluster-id")

	pool1 := defaultGkeMachinePool(nil)
	pool2 := defaultGkeMachinePool(map[string]interface{}{
		"name":  "worker-pool-2",
		"count": 3,
	})

	require.NoError(t, d.Set("machine_pool", gkeMachinePoolSet(pool1, pool2)))
	require.NoError(t, d.Set("machine_pool", gkeMachinePoolSet(pool1)))

	diags := resourceClusterGkeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterGkeUpdateClusterProfileWithMock(t *testing.T) {
	d := prepareGkeClusterResourceData(t)
	d.SetId("test-gke-cluster-id")
	require.NoError(t, d.Set("tags", []interface{}{"skip_apply"}))

	setChangedClusterProfiles(t, d,
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-2"}},
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-1"}},
	)

	diags := resourceClusterGkeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}
