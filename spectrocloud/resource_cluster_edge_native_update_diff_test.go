package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

// resourceClusterEdgeNativeUpdate's machine_pool branch is gated behind
// d.HasChange("machine_pool"), which schema.TestResourceData's
// Set-then-Set pattern never fires (see
// resource_cluster_apache_cloudstack_test.go's buildMachinePoolChangeResourceData
// for the rationale). buildEdgeNativeMachinePoolChangeResourceData builds a
// real InstanceState + config diff via Resource.Diff so HasChange/GetChange
// behave the way Terraform's own apply pipeline would produce them.
func buildEdgeNativeMachinePoolChangeResourceData(t *testing.T, oldPools, newPools []interface{}) *schema.ResourceData {
	t.Helper()
	res := resourceClusterEdgeNative()

	base := map[string]interface{}{
		"name":    "test-edge-native-cluster",
		"context": "project",
		"cloud_config": []interface{}{
			map[string]interface{}{
				"vip": "10.0.0.100",
			},
		},
	}

	oldRaw := map[string]interface{}{}
	for k, v := range base {
		oldRaw[k] = v
	}
	oldRaw["machine_pool"] = oldPools
	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId(edgeNativeClusterID)
	require.NoError(t, oldRD.Set("cloud_config_id", edgeNativeCloudConfigUID))
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newRaw := map[string]interface{}{}
	for k, v := range base {
		newRaw[k] = v
	}
	newRaw["machine_pool"] = newPools
	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)
	require.NotNil(t, diff)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId(edgeNativeClusterID)
	return finalRD
}

func edgeNativePoolRaw(name string, controlPlane bool, hostUID, hostName string) map[string]interface{} {
	return map[string]interface{}{
		"name":                    name,
		"control_plane":           controlPlane,
		"control_plane_as_worker": controlPlane,
		"edge_host": []interface{}{
			map[string]interface{}{
				"host_uid":  hostUID,
				"host_name": hostName,
			},
		},
	}
}

// TestResourceClusterEdgeNativeUpdate_MachinePoolDiff drives, in a single
// Update call, all three machine_pool branches gated by
// d.HasChange("machine_pool"): create (new-pool), update-with-host-deletion
// (pool-to-change), and full-pool removal (pool-to-remove, which also
// exercises the warningMessageForNodeDeletion diag.Warning append).
func TestResourceClusterEdgeNativeUpdate_MachinePoolDiff(t *testing.T) {
	oldPools := []interface{}{
		edgeNativePoolRaw("cp-pool", true, "edge-host-1", ""),
		edgeNativePoolRaw(routes.EdgeNativePoolChangeName, false, "host-a-uid", routes.EdgeNativePoolChangeHostName),
		edgeNativePoolRaw(routes.EdgeNativePoolRemoveName, false, "host-b-uid", routes.EdgeNativePoolRemoveHostName),
	}
	newPools := []interface{}{
		edgeNativePoolRaw("cp-pool", true, "edge-host-1", ""),
		edgeNativePoolRaw(routes.EdgeNativePoolChangeName, false, "host-a-uid", "host-kept"),
		edgeNativePoolRaw("new-pool", false, "host-c-uid", "host-new"),
	}

	d := buildEdgeNativeMachinePoolChangeResourceData(t, oldPools, newPools)

	old, cur := d.GetChange("machine_pool")
	require.Equal(t, 3, old.(*schema.Set).Len())
	require.Equal(t, 3, cur.(*schema.Set).Len())

	diags := resourceClusterEdgeNativeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)

	foundWarning := false
	for _, dg := range diags {
		if dg.Severity == diag.Warning {
			foundWarning = true
		}
	}
	assert.True(t, foundWarning, "expected a node-deletion warning diagnostic, got: %+v", diags)
}

// TestResourceClusterEdgeNativeUpdate_NewMachinePoolConversionError exercises
// the toMachinePoolEdgeNative error branch inside the main machine_pool
// diff loop (a brand-new pool, not present in the old set, whose
// control-plane node_repave_interval is invalid).
func TestResourceClusterEdgeNativeUpdate_NewMachinePoolConversionError(t *testing.T) {
	oldPools := []interface{}{
		edgeNativePoolRaw("cp-pool", true, "edge-host-1", ""),
	}
	newPools := []interface{}{
		edgeNativePoolRaw("cp-pool", true, "edge-host-1", ""),
		map[string]interface{}{
			"name":                    "bad-new-pool",
			"control_plane":           true,
			"control_plane_as_worker": false,
			"node_repave_interval":    30,
			"edge_host": []interface{}{
				map[string]interface{}{"host_uid": "bad-host-uid", "host_name": "bad-host"},
			},
		},
	}

	d := buildEdgeNativeMachinePoolChangeResourceData(t, oldPools, newPools)
	diags := resourceClusterEdgeNativeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterEdgeNativeUpdate_MachinePoolListError exercises the
// GetNodeListInEdgeNativeMachinePool API-error branch reached when a
// changed machine pool needs to reconcile deleted hosts.
func TestResourceClusterEdgeNativeUpdate_MachinePoolListError(t *testing.T) {
	oldPools := []interface{}{
		edgeNativePoolRaw(routes.EdgeNativePoolListErrorName, false, "host-a-uid", "host-removed"),
	}
	newPools := []interface{}{
		edgeNativePoolRaw(routes.EdgeNativePoolListErrorName, false, "host-a-uid", "host-kept"),
	}

	d := buildEdgeNativeMachinePoolChangeResourceData(t, oldPools, newPools)
	diags := resourceClusterEdgeNativeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterEdgeNativeUpdate_RepaveApprovalError exercises the
// validateSystemRepaveApproval error branch: cluster-uid-server-error makes
// GetCluster fail, which resourceClusterEdgeNativeUpdate must surface via
// diag.FromErr before touching machine_pool at all.
func TestResourceClusterEdgeNativeUpdate_RepaveApprovalError(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterEdgeNative(), "cluster-uid-server-error",
		baseEdgeNativeUpdateAttrs(), simpleDiff("cluster_timezone", "", "America/New_York"))

	diags := resourceClusterEdgeNativeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}
