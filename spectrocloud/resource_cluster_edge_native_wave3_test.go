package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

// ---------------------------------------------------------------------
// resourceClusterEdgeNativeRead — cluster-gone / API-error branches.
// ---------------------------------------------------------------------

func TestResourceClusterEdgeNativeReadClusterNotFound(t *testing.T) {
	d := prepareEdgeNativeClusterResourceData(t)
	d.SetId("cluster-uid-not-found")

	diags := resourceClusterEdgeNativeRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "", d.Id(), "deleted cluster should clear the resource ID")
}

func TestResourceClusterEdgeNativeReadServerError(t *testing.T) {
	d := prepareEdgeNativeClusterResourceData(t)
	d.SetId("cluster-uid-server-error")

	diags := resourceClusterEdgeNativeRead(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// ---------------------------------------------------------------------
// flattenCloudConfigEdgeNative — GetCloudConfigEdgeNative API-error branch.
// ---------------------------------------------------------------------

func TestFlattenCloudConfigEdgeNativeAPIError(t *testing.T) {
	d := resourceClusterEdgeNative().TestResourceData()
	d.SetId("test-cluster-uid")
	require.NoError(t, d.Set("context", "project"))

	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	diags := flattenCloudConfigEdgeNative(routes.EdgeNativeCloudConfigErrorUID, d, c)
	assert.True(t, diags.HasError())
}

// ---------------------------------------------------------------------
// flattenEdgeNativePoolHost — nil-host / nil-HostUID branches.
// ---------------------------------------------------------------------

func TestFlattenEdgeNativePoolHostNil(t *testing.T) {
	assert.Nil(t, flattenEdgeNativePoolHost(nil))
}

func TestFlattenEdgeNativePoolHostNilHostUID(t *testing.T) {
	assert.Nil(t, flattenEdgeNativePoolHost(&models.V1EdgeNativeHost{HostUID: nil}))
}

// ---------------------------------------------------------------------
// flattenMachinePoolConfigsEdgeNative — override_kubeadm_configuration
// (worker pools only) branch.
// ---------------------------------------------------------------------

func TestFlattenMachinePoolConfigsEdgeNativeOverrideKubeadm(t *testing.T) {
	hostUID := "host-1"
	result := flattenMachinePoolConfigsEdgeNative([]*models.V1EdgeNativeMachinePoolConfig{
		{
			Name:                         "worker-pool",
			IsControlPlane:               false,
			OverrideKubeadmConfiguration: "kubeletExtraArgs:\n  v: \"4\"",
			Hosts: []*models.V1EdgeNativeHost{
				{HostUID: &hostUID},
			},
		},
	})
	require.Len(t, result, 1)
	assert.Equal(t, "kubeletExtraArgs:\n  v: \"4\"", result[0].(map[string]interface{})["override_kubeadm_configuration"])
}

func TestFlattenMachinePoolConfigsEdgeNativeOverrideKubeadmSkippedForControlPlane(t *testing.T) {
	hostUID := "host-1"
	result := flattenMachinePoolConfigsEdgeNative([]*models.V1EdgeNativeMachinePoolConfig{
		{
			Name:                         "cp-pool",
			IsControlPlane:               true,
			OverrideKubeadmConfiguration: "should-be-ignored",
			Hosts: []*models.V1EdgeNativeHost{
				{HostUID: &hostUID},
			},
		},
	})
	require.Len(t, result, 1)
	_, ok := result[0].(map[string]interface{})["override_kubeadm_configuration"]
	assert.False(t, ok, "override_kubeadm_configuration must not be set for control plane pools")
}

// ---------------------------------------------------------------------
// toEdgeNativeCluster — error branches: profile resolution, invalid CIDR,
// and a failing machine pool conversion.
// ---------------------------------------------------------------------

func TestToEdgeNativeClusterProfileResolutionError(t *testing.T) {
	d := resourceClusterEdgeNative().TestResourceData()
	d.SetId("cluster-uid-not-found")
	require.NoError(t, d.Set("name", "test-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{"vip": "10.0.0.100"},
	}))
	require.NoError(t, d.Set("machine_pool", schema.NewSet(resourceMachinePoolEdgeNativeHash, []interface{}{
		map[string]interface{}{
			"name":          "pool1",
			"control_plane": true,
			"edge_host": schema.NewSet(resourceEdgeHostHash, []interface{}{
				map[string]interface{}{"host_uid": "uid1"},
			}),
		},
	})))

	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	_, err := toEdgeNativeCluster(c, d)
	assert.Error(t, err)
}

func TestToEdgeNativeClusterInvalidOverlayCidrError(t *testing.T) {
	d := resourceClusterEdgeNative().TestResourceData()
	require.NoError(t, d.Set("name", "test-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{"overlay_cidr_range": "not-a-cidr"},
	}))
	require.NoError(t, d.Set("machine_pool", schema.NewSet(resourceMachinePoolEdgeNativeHash, []interface{}{
		map[string]interface{}{
			"name":          "pool1",
			"control_plane": true,
			"edge_host": schema.NewSet(resourceEdgeHostHash, []interface{}{
				map[string]interface{}{"host_uid": "uid1"},
			}),
		},
	})))

	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	_, err := toEdgeNativeCluster(c, d)
	assert.Error(t, err)
}

func TestToEdgeNativeClusterMachinePoolError(t *testing.T) {
	d := resourceClusterEdgeNative().TestResourceData()
	require.NoError(t, d.Set("name", "test-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{"vip": "10.0.0.100"},
	}))
	// control_plane pool with a non-zero node_repave_interval is invalid
	// (ValidationNodeRepaveIntervalForControlPlane), so toMachinePoolEdgeNative
	// errors and toEdgeNativeCluster must propagate it.
	require.NoError(t, d.Set("machine_pool", schema.NewSet(resourceMachinePoolEdgeNativeHash, []interface{}{
		map[string]interface{}{
			"name":                 "pool1",
			"control_plane":        true,
			"node_repave_interval": 30,
			"edge_host": schema.NewSet(resourceEdgeHostHash, []interface{}{
				map[string]interface{}{"host_uid": "uid1"},
			}),
		},
	})))

	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	_, err := toEdgeNativeCluster(c, d)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------
// toMachinePoolEdgeNative — override_kubeadm_configuration (worker) and
// invalid edge_host type propagation branches.
// ---------------------------------------------------------------------

func TestToMachinePoolEdgeNativeOverrideKubeadmWorker(t *testing.T) {
	mp, err := toMachinePoolEdgeNative(map[string]interface{}{
		"name":                           "worker-pool",
		"control_plane":                  false,
		"control_plane_as_worker":        false,
		"node_repave_interval":           0,
		"override_kubeadm_configuration": "kubeletExtraArgs:\n  v: \"4\"",
		"edge_host": schema.NewSet(resourceEdgeHostHash, []interface{}{
			map[string]interface{}{"host_uid": "uid1"},
		}),
	})
	require.NoError(t, err)
	assert.Equal(t, "kubeletExtraArgs:\n  v: \"4\"", mp.PoolConfig.OverrideKubeadmConfiguration)
}

func TestToMachinePoolEdgeNativeInvalidEdgeHostType(t *testing.T) {
	_, err := toMachinePoolEdgeNative(map[string]interface{}{
		"name":                    "worker-pool",
		"control_plane":           false,
		"control_plane_as_worker": false,
		"edge_host":               "not-a-set-or-list",
	})
	assert.Error(t, err)
}

func TestToMachinePoolEdgeNativeControlPlaneNodeRepaveError(t *testing.T) {
	_, err := toMachinePoolEdgeNative(map[string]interface{}{
		"name":                    "cp-pool",
		"control_plane":           true,
		"control_plane_as_worker": false,
		"node_repave_interval":    60,
		"edge_host": schema.NewSet(resourceEdgeHostHash, []interface{}{
			map[string]interface{}{"host_uid": "uid1"},
		}),
	})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------
// toEdgeHosts — fallback list-conversion branch (edge_host as []interface{}
// instead of *schema.Set) and invalid-type error branch.
// ---------------------------------------------------------------------

func TestToEdgeHostsListFallback(t *testing.T) {
	result, err := toEdgeHosts(map[string]interface{}{
		"edge_host": []interface{}{
			map[string]interface{}{"host_uid": "uid1", "host_name": "host1"},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.EdgeHosts, 1)
	assert.Equal(t, "uid1", *result.EdgeHosts[0].HostUID)
}

func TestToEdgeHostsInvalidType(t *testing.T) {
	_, err := toEdgeHosts(map[string]interface{}{
		"edge_host": "not-a-set-or-list",
	})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------
// resourceClusterEdgeNativeStateUpgradeV2 — machine_pool already a plain
// []interface{} (rather than *schema.Set) branch.
// ---------------------------------------------------------------------

func TestResourceClusterEdgeNativeStateUpgradeV2MachinePoolList(t *testing.T) {
	raw := map[string]interface{}{
		"machine_pool": []interface{}{
			map[string]interface{}{
				"name": "pool-1",
				"edge_host": []interface{}{
					map[string]interface{}{"host_uid": "uid-1"},
				},
			},
		},
	}

	out, err := resourceClusterEdgeNativeStateUpgradeV2(context.Background(), raw, nil)
	require.NoError(t, err)

	machinePool, ok := out["machine_pool"].([]interface{})
	require.True(t, ok)
	require.Len(t, machinePool, 1)
	mp := machinePool[0].(map[string]interface{})
	edgeHosts, ok := mp["edge_host"].([]interface{})
	require.True(t, ok)
	assert.Len(t, edgeHosts, 1)
}
