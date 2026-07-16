package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// (test file header) edge_native CRUD.

const (
	edgeNativeCloudConfigUID = "test-cloud-config-id"
	edgeNativeClusterProfile = "cluster-profile-import-1"
	edgeNativeClusterID      = "test-edge-native-cluster-id"
)

func defaultEdgeHost(hostUID string) map[string]interface{} {
	return map[string]interface{}{
		"host_uid":  hostUID,
		"host_name": "",
	}
}

func edgeHostSet(hosts ...map[string]interface{}) *schema.Set {
	items := make([]interface{}, len(hosts))
	for i, h := range hosts {
		items[i] = h
	}
	return schema.NewSet(resourceEdgeHostHash, items)
}

func defaultEdgeNativeMachinePool(overrides map[string]interface{}) map[string]interface{} {
	pool := map[string]interface{}{
		"name":                    "cp-pool",
		"control_plane":           true,
		"control_plane_as_worker": true,
		"edge_host":               edgeHostSet(defaultEdgeHost("edge-host-1")),
	}
	for k, v := range overrides {
		pool[k] = v
	}
	return pool
}

func edgeNativeMachinePoolSet(pools ...map[string]interface{}) *schema.Set {
	items := make([]interface{}, len(pools))
	for i, p := range pools {
		items[i] = p
	}
	return schema.NewSet(resourceMachinePoolEdgeNativeHash, items)
}

func prepareEdgeNativeClusterResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceClusterEdgeNative().TestResourceData()
	require.NoError(t, d.Set("name", "test-edge-native-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_config_id", edgeNativeCloudConfigUID))
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": edgeNativeClusterProfile},
	}))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"vip":                "10.0.0.100",
			"overlay_cidr_range": "",
		},
	}))
	require.NoError(t, d.Set("machine_pool", edgeNativeMachinePoolSet(defaultEdgeNativeMachinePool(nil))))
	return d
}

func TestResourceClusterEdgeNativeReadWithMock(t *testing.T) {
	d := prepareEdgeNativeClusterResourceData(t)
	d.SetId(edgeNativeClusterID)

	diags := resourceClusterEdgeNativeRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, edgeNativeCloudConfigUID, d.Get("cloud_config_id"))
}

func TestResourceClusterEdgeNativeCreateWithMock(t *testing.T) {
	d := prepareEdgeNativeClusterResourceData(t)
	require.NoError(t, d.Set("tags", []interface{}{"skip_completion"}))

	diags := resourceClusterEdgeNativeCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, edgeNativeClusterID, d.Id())
}

func TestResourceClusterEdgeNativeUpdateMachinePoolWithMock(t *testing.T) {
	d := prepareEdgeNativeClusterResourceData(t)
	d.SetId(edgeNativeClusterID)

	oldPool := defaultEdgeNativeMachinePool(nil)
	newPool := defaultEdgeNativeMachinePool(map[string]interface{}{
		"edge_host": edgeHostSet(
			defaultEdgeHost("edge-host-1"),
			defaultEdgeHost("edge-host-2"),
		),
	})

	require.NoError(t, d.Set("machine_pool", edgeNativeMachinePoolSet(oldPool)))
	require.NoError(t, d.Set("machine_pool", edgeNativeMachinePoolSet(newPool)))

	diags := resourceClusterEdgeNativeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceClusterEdgeNativeUpdateClusterProfileWithMock(t *testing.T) {
	d := prepareEdgeNativeClusterResourceData(t)
	d.SetId(edgeNativeClusterID)
	require.NoError(t, d.Set("tags", []interface{}{"skip_apply"}))

	setChangedClusterProfiles(t, d,
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-2"}},
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-1"}},
	)

	diags := resourceClusterEdgeNativeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}
