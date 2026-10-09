package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

// ---------------------------------------------------------------------
// resourceClusterEdgeNativeUpdate — cloud_config Day-2 update (PLT-2457).
//
// Prior to this fix, cloud_config was ForceNew and Update never called
// UpdateCloudConfigEdgeNative, so a vip/ssh_keys/ntp_servers change was
// either silently dropped or forced a full cluster replace. These tests
// verify the new d.HasChange("cloud_config") path calls the API and that
// a backend rejection (Hubble's PEM-10966 VIP validation) is surfaced as
// an apply-time error instead of being swallowed.
//
// cloud_config is a TypeList (nested block), so — like the AWS/AKS/MAAS
// equivalents in resource_cluster_aws_update_diff_test.go — a bare
// Set()-then-Set() on schema.TestResourceData never produces a real diff
// and d.HasChange("cloud_config") stays false. buildEdgeNativeUpdateResourceData
// mirrors buildAwsUpdateResourceData: build an old InstanceState + a new
// config, run it through Resource.Diff, then reconstruct ResourceData from
// that InstanceState+InstanceDiff pair so HasChange behaves the way
// Terraform's own apply pipeline would produce it.
// ---------------------------------------------------------------------

func edgeNativeCloudConfigRawMap(overrides map[string]interface{}) map[string]interface{} {
	cc := map[string]interface{}{
		"vip":                "10.0.0.100",
		"overlay_cidr_range": "",
		"ssh_keys":           []interface{}{"ssh-rsa AAAA..."},
	}
	for k, v := range overrides {
		cc[k] = v
	}
	return cc
}

func baseEdgeNativeRaw(cloudConfig map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":    "test-edge-native-cluster",
		"context": "project",
		"cloud_config": []interface{}{
			cloudConfig,
		},
		"machine_pool": []interface{}{
			map[string]interface{}{
				"name":          "cp-pool",
				"control_plane": true,
				"edge_host": []interface{}{
					map[string]interface{}{"host_uid": "edge-host-1"},
				},
			},
		},
	}
}

func buildEdgeNativeUpdateResourceData(t *testing.T, oldRaw, newRaw map[string]interface{}, configUID string) *schema.ResourceData {
	t.Helper()
	res := resourceClusterEdgeNative()

	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId(edgeNativeClusterID)
	require.NoError(t, oldRD.Set("cloud_config_id", configUID))
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId(edgeNativeClusterID)
	return finalRD
}

// TestResourceClusterEdgeNativeUpdateCloudConfigWithMock exercises the
// d.HasChange("cloud_config") branch's happy path: toClusterConfigEdgeNativeUpdate
// + UpdateCloudConfigEdgeNative are invoked with the new cloud_config values,
// and cloud_config is no longer ForceNew so this is an in-place update, not
// a replace.
func TestResourceClusterEdgeNativeUpdateCloudConfigWithMock(t *testing.T) {
	oldRaw := baseEdgeNativeRaw(edgeNativeCloudConfigRawMap(nil))
	newRaw := baseEdgeNativeRaw(edgeNativeCloudConfigRawMap(map[string]interface{}{
		"vip": "10.0.0.200",
	}))

	d := buildEdgeNativeUpdateResourceData(t, oldRaw, newRaw, edgeNativeCloudConfigUID)
	require.True(t, d.HasChange("cloud_config"))
	require.False(t, resourceClusterEdgeNative().Schema["cloud_config"].ForceNew,
		"cloud_config must not be ForceNew, or this diff would be a destroy/recreate instead of an update")

	diags := resourceClusterEdgeNativeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterEdgeNativeUpdateCloudConfigRejectedWithMock exercises
// the UpdateCloudConfigEdgeNative API-error branch, simulating Hubble's
// PEM-10966 VIP-immutability validation rejecting the change. Before this
// fix, this error was never even requested (Terraform never called this
// endpoint on Update at all).
func TestResourceClusterEdgeNativeUpdateCloudConfigRejectedWithMock(t *testing.T) {
	oldRaw := baseEdgeNativeRaw(edgeNativeCloudConfigRawMap(nil))
	newRaw := baseEdgeNativeRaw(edgeNativeCloudConfigRawMap(map[string]interface{}{
		"vip": "10.0.0.200",
	}))

	d := buildEdgeNativeUpdateResourceData(t, oldRaw, newRaw, routes.EdgeNativeCloudConfigUpdateRejectUID)
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterEdgeNativeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError(), "expected the backend VIP validation rejection to surface as an error")
}
