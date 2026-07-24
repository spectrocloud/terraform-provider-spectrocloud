package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// (test file header) vSphere CRUD coverage.
// Pattern shared with the AWS/GCP wave2 tests.

const (
	vsphereCloudConfigUID  = "test-cloud-config-id"
	vsphereCloudAccountUID = "test-vsphere-account-id-1"
	vsphereClusterProfile  = "cluster-profile-import-1"
	vsphereClusterID       = "test-vsphere-cluster-id"
)

// defaultVspherePlacement returns the placement block every machine pool
// needs. Kept as a helper because the placement fields are the same
// across every pool the tests build.
func defaultVspherePlacement() map[string]interface{} {
	// static_ip_pool_id has to be present (even as an empty string) —
	// resourceMachinePoolVsphereHash does place["static_ip_pool_id"].(string)
	// unconditionally, so a missing key panics.
	return map[string]interface{}{
		"cluster":           "test-cluster",
		"resource_pool":     "test-pool",
		"datastore":         "test-datastore",
		"network":           "test-network",
		"static_ip_pool_id": "",
	}
}

func defaultVsphereMachinePool(overrides map[string]interface{}) map[string]interface{} {
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
		"placement": []interface{}{defaultVspherePlacement()},
	}
	for k, v := range overrides {
		pool[k] = v
	}
	return pool
}

func vsphereMachinePoolSet(pools ...map[string]interface{}) *schema.Set {
	items := make([]interface{}, len(pools))
	for i, p := range pools {
		items[i] = p
	}
	return schema.NewSet(resourceMachinePoolVsphereHash, items)
}

func prepareVsphereClusterResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceClusterVsphere().TestResourceData()
	require.NoError(t, d.Set("name", "test-vsphere-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_account_id", vsphereCloudAccountUID))
	require.NoError(t, d.Set("cloud_config_id", vsphereCloudConfigUID))
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": vsphereClusterProfile},
	}))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"datacenter":    "test-datacenter",
			"folder":        "test-folder",
			"ssh_key":       "ssh-rsa AAAA",
			"network_type":  "VIP",
			"host_endpoint": "10.0.0.100",
		},
	}))
	require.NoError(t, d.Set("machine_pool", vsphereMachinePoolSet(defaultVsphereMachinePool(nil))))
	return d
}

func TestResourceClusterVsphereReadWithMock(t *testing.T) {
	d := prepareVsphereClusterResourceData(t)
	d.SetId(vsphereClusterID)

	diags := resourceClusterVsphereRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, vsphereCloudConfigUID, d.Get("cloud_config_id"))
	assert.Equal(t, vsphereCloudAccountUID, d.Get("cloud_account_id"))

	pools := d.Get("machine_pool").(*schema.Set)
	assert.Greater(t, pools.Len(), 0)
}

// TestResourceClusterVsphereReadClusterDeleted exercises the "cluster == nil"
// branch: GetCluster returns (nil, nil) once the fixture's State is
// "Deleted", so Read must clear the resource's ID instead of erroring.
func TestResourceClusterVsphereReadClusterDeleted(t *testing.T) {
	d := prepareVsphereClusterResourceData(t)
	d.SetId("cluster-uid-deleted-state")

	diags := resourceClusterVsphereRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "", d.Id())
}

// TestResourceClusterVsphereReadCloudTypeMismatch exercises the
// ValidateCloudType error branch: the default cluster fixture reports
// CloudType="aws", which resourceClusterVsphereRead must reject.
func TestResourceClusterVsphereReadCloudTypeMismatch(t *testing.T) {
	d := prepareVsphereClusterResourceData(t)
	d.SetId("test-cluster-id")

	diags := resourceClusterVsphereRead(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterVsphereReadGetCloudConfigError exercises the
// GetCloudConfigVsphere error branch: the fixture's CloudConfigRef points at
// routes.VsphereCloudConfigErrorUID.
func TestResourceClusterVsphereReadGetCloudConfigError(t *testing.T) {
	d := prepareVsphereClusterResourceData(t)
	d.SetId("test-vsphere-cluster-cloudconfig-error-id")

	diags := resourceClusterVsphereRead(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

func TestResourceClusterVsphereCreateWithMock(t *testing.T) {
	d := prepareVsphereClusterResourceData(t)
	require.NoError(t, d.Set("tags", []interface{}{"skip_completion"}))

	diags := resourceClusterVsphereCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, vsphereClusterID, d.Id())
}

func TestResourceClusterVsphereUpdateMachinePoolWithMock(t *testing.T) {
	d := prepareVsphereClusterResourceData(t)
	d.SetId(vsphereClusterID)

	oldPool := defaultVsphereMachinePool(nil)
	newPool := defaultVsphereMachinePool(map[string]interface{}{"count": 3})

	require.NoError(t, d.Set("machine_pool", vsphereMachinePoolSet(oldPool)))
	require.NoError(t, d.Set("machine_pool", vsphereMachinePoolSet(newPool)))

	diags := resourceClusterVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceClusterVsphereUpdateMachinePoolAddWithMock(t *testing.T) {
	d := prepareVsphereClusterResourceData(t)
	d.SetId(vsphereClusterID)

	cpPool := defaultVsphereMachinePool(nil)
	workerPool := defaultVsphereMachinePool(map[string]interface{}{
		"name":                    "worker-pool",
		"count":                   2,
		"control_plane":           false,
		"control_plane_as_worker": false,
	})

	require.NoError(t, d.Set("machine_pool", vsphereMachinePoolSet(cpPool)))
	require.NoError(t, d.Set("machine_pool", vsphereMachinePoolSet(cpPool, workerPool)))

	diags := resourceClusterVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceClusterVsphereUpdateMachinePoolDeleteWithMock(t *testing.T) {
	d := prepareVsphereClusterResourceData(t)
	d.SetId(vsphereClusterID)

	cpPool := defaultVsphereMachinePool(nil)
	workerPool := defaultVsphereMachinePool(map[string]interface{}{
		"name":                    "worker-pool",
		"count":                   2,
		"control_plane":           false,
		"control_plane_as_worker": false,
	})

	require.NoError(t, d.Set("machine_pool", vsphereMachinePoolSet(cpPool, workerPool)))
	require.NoError(t, d.Set("machine_pool", vsphereMachinePoolSet(cpPool)))

	diags := resourceClusterVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceClusterVsphereUpdateClusterProfileWithMock(t *testing.T) {
	d := prepareVsphereClusterResourceData(t)
	d.SetId(vsphereClusterID)
	require.NoError(t, d.Set("tags", []interface{}{"skip_apply"}))

	setChangedClusterProfiles(t, d,
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-2"}},
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-1"}},
	)

	diags := resourceClusterVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterVsphereStateUpgradeV0 — pins the TypeList→TypeSet
// conversion for cluster_profile (matches the pattern in AKS V3 and
// Azure V0 upgraders).
func TestResourceClusterVsphereStateUpgradeV0(t *testing.T) {
	ctx := context.Background()

	t.Run("list passes through", func(t *testing.T) {
		state := map[string]interface{}{
			"cluster_profile": []interface{}{
				map[string]interface{}{"id": "cp-1"},
			},
		}
		got, err := resourceClusterVsphereStateUpgradeV0(ctx, state, nil)
		require.NoError(t, err)
		cp, ok := got["cluster_profile"].([]interface{})
		require.True(t, ok)
		assert.Len(t, cp, 1)
	})

	t.Run("missing key is a no-op", func(t *testing.T) {
		got, err := resourceClusterVsphereStateUpgradeV0(ctx, map[string]interface{}{}, nil)
		require.NoError(t, err)
		_, exists := got["cluster_profile"]
		assert.False(t, exists)
	})
}
