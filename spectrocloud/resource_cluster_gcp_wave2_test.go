package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// (test file header) GCP CRUD coverage.
//
// Mirrors the shape of resource_cluster_aws_wave2_test.go so the two
// clouds share a testing style. What's exercised here:
//   - Read against the mock returns cloud_config_id / cloud_account_id
//   - Create rounds through the mock's POST /spectroclusters/gcp and
//     sets d.Id() to the returned UID
//   - Update cloud_config (region/network/vpc change)
//   - Update machine pool (count change → PUT machinePools)
//   - Update machine pool set (add worker pool → POST, remove → DELETE)
//   - Update cluster profile (patchProfiles path via skip_apply tag)

const (
	gcpCloudConfigUID  = "test-cloud-config-id"
	gcpCloudAccountUID = "test-gcp-account-id-1"
	gcpClusterProfile  = "cluster-profile-import-1"
)

// defaultGcpMachinePool returns a control-plane pool matching the mock's
// primary pool. overrides let a test pin one or two fields (e.g. count)
// without rebuilding the whole map.
func defaultGcpMachinePool(overrides map[string]interface{}) map[string]interface{} {
	pool := map[string]interface{}{
		"name":                    "cp-pool",
		"instance_type":           "n1-standard-2",
		"count":                   1,
		"control_plane":           true,
		"control_plane_as_worker": true,
		"disk_size_gb":            65,
		"update_strategy":         "RollingUpdateScaleOut",
		"azs":                     schema.NewSet(schema.HashString, []interface{}{"us-central1-a"}),
	}
	for k, v := range overrides {
		pool[k] = v
	}
	return pool
}

func gcpMachinePoolSet(pools ...map[string]interface{}) *schema.Set {
	items := make([]interface{}, len(pools))
	for i, p := range pools {
		items[i] = p
	}
	return schema.NewSet(resourceMachinePoolGcpHash, items)
}

// prepareGcpClusterResourceData is the shared fixture builder for the
// CRUD tests below. Populates every field the resource's Create/Read
// paths touch — matching the mock's echo-back payload so an
// Update leg after Read reports no drift.
func prepareGcpClusterResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceClusterGcp().TestResourceData()
	require.NoError(t, d.Set("name", "test-gcp-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_account_id", gcpCloudAccountUID))
	require.NoError(t, d.Set("cloud_config_id", gcpCloudConfigUID))
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": gcpClusterProfile},
	}))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"project": "test-gcp-project",
			"region":  "us-central1",
			"network": "test-network",
		},
	}))
	require.NoError(t, d.Set("machine_pool", gcpMachinePoolSet(defaultGcpMachinePool(nil))))
	return d
}

func TestResourceClusterGcpReadWithMock(t *testing.T) {
	d := prepareGcpClusterResourceData(t)
	d.SetId("test-gcp-cluster-id")

	diags := resourceClusterGcpRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, gcpCloudConfigUID, d.Get("cloud_config_id"))
	assert.Equal(t, gcpCloudAccountUID, d.Get("cloud_account_id"))

	pools := d.Get("machine_pool").(*schema.Set)
	assert.Greater(t, pools.Len(), 0)
}

func TestFlattenCloudConfigGcpWithMock(t *testing.T) {
	// Direct call into flattenCloudConfigGcp to cover the cloud-config
	// GET → set path. resourceClusterGcpRead exercises the outer edges
	// (cluster read + common fields); this one pins the GCP-specific
	// flatten.
	d := prepareGcpClusterResourceData(t)
	c := mustUnitClient(t, false)

	diags := flattenCloudConfigGcp(gcpCloudConfigUID, d, c)
	assert.False(t, diags.HasError())
	assert.Equal(t, gcpCloudAccountUID, d.Get("cloud_account_id"))

	cfg := d.Get("cloud_config").([]interface{})
	require.Len(t, cfg, 1)
	assert.Equal(t, "us-central1", cfg[0].(map[string]interface{})["region"])
	assert.Equal(t, "test-gcp-project", cfg[0].(map[string]interface{})["project"])
}

func TestResourceClusterGcpCreateWithMock(t *testing.T) {
	d := prepareGcpClusterResourceData(t)
	// skip_completion short-circuits the 30-second wait-for-cluster
	// state machine (see waitForClusterCreation) — without it this
	// test would pay ~30s per run for the initial Delay.
	require.NoError(t, d.Set("tags", []interface{}{"skip_completion"}))

	diags := resourceClusterGcpCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-gcp-cluster-id", d.Id())
}

func TestResourceClusterGcpUpdateCloudConfigWithMock(t *testing.T) {
	d := prepareGcpClusterResourceData(t)
	d.SetId("test-cluster-id")

	// Simulate a network change. Set old first then new to give the
	// resource a diff to react to; the mock's PUT clusterConfig accepts
	// any body with 204.
	oldCfg := map[string]interface{}{
		"project": "test-gcp-project",
		"region":  "us-central1",
		"network": "old-network",
	}
	newCfg := map[string]interface{}{
		"project": "test-gcp-project",
		"region":  "us-central1",
		"network": "new-network",
	}
	require.NoError(t, d.Set("cloud_config", []interface{}{oldCfg}))
	require.NoError(t, d.Set("cloud_config", []interface{}{newCfg}))

	diags := resourceClusterGcpUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterGcpUpdateMachinePoolWithMock(t *testing.T) {
	d := prepareGcpClusterResourceData(t)
	d.SetId("test-cluster-id")

	// Bump count on the CP pool — should trigger the PUT
	// /machinePools/{name} branch of resourceClusterGcpUpdate.
	oldPool := defaultGcpMachinePool(nil)
	newPool := defaultGcpMachinePool(map[string]interface{}{"count": 3})

	require.NoError(t, d.Set("machine_pool", gcpMachinePoolSet(oldPool)))
	require.NoError(t, d.Set("machine_pool", gcpMachinePoolSet(newPool)))

	diags := resourceClusterGcpUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterGcpUpdateMachinePoolAddWithMock(t *testing.T) {
	// Adding a worker pool → POST /machinePools branch.
	d := prepareGcpClusterResourceData(t)
	d.SetId("test-cluster-id")

	cpPool := defaultGcpMachinePool(nil)
	workerPool := defaultGcpMachinePool(map[string]interface{}{
		"name":                    "worker-pool",
		"count":                   2,
		"control_plane":           false,
		"control_plane_as_worker": false,
	})

	require.NoError(t, d.Set("machine_pool", gcpMachinePoolSet(cpPool)))
	require.NoError(t, d.Set("machine_pool", gcpMachinePoolSet(cpPool, workerPool)))

	diags := resourceClusterGcpUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterGcpUpdateMachinePoolDeleteWithMock(t *testing.T) {
	// Removing the worker pool → DELETE /machinePools/{name} branch.
	d := prepareGcpClusterResourceData(t)
	d.SetId("test-cluster-id")

	cpPool := defaultGcpMachinePool(nil)
	workerPool := defaultGcpMachinePool(map[string]interface{}{
		"name":                    "worker-pool",
		"count":                   2,
		"control_plane":           false,
		"control_plane_as_worker": false,
	})

	require.NoError(t, d.Set("machine_pool", gcpMachinePoolSet(cpPool, workerPool)))
	require.NoError(t, d.Set("machine_pool", gcpMachinePoolSet(cpPool)))

	diags := resourceClusterGcpUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterGcpUpdateClusterProfileWithMock(t *testing.T) {
	// skip_apply short-circuits the download-and-install wait; we want
	// updateProfiles to fire the PATCH profiles branch but not block on
	// canBeApplied polling.
	d := prepareGcpClusterResourceData(t)
	d.SetId("test-cluster-id")
	require.NoError(t, d.Set("tags", []interface{}{"skip_apply"}))

	setChangedClusterProfiles(t, d,
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-2"}},
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-1"}},
	)

	diags := resourceClusterGcpUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}
