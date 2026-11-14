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

// PLT-2476 adds `update_lock` to spectrocloud_cluster_edge_native's
// `cloud_config`. On Day 1 it is sent in the create body's cloud config; on
// Day 2 it is mapped to the API's dedicated profile-update-lock endpoint
// (PATCH /v1/cloudconfigs/edge-native/{uid}/updateLock), separate from the
// general cloud config PUT/PATCH used by the other cloud_config fields.
// Palette engages the lock automatically once deployment completes and
// releasing it is only possible from the device's Local UI (out of scope for
// this provider per the epic's comment thread) - this provider only supports
// reading the current state and explicitly enabling/disabling it.

// TestCloudConfig_UpdateLock_Schema is the schema-level guard: update_lock
// must be Optional with a static "disabled" Default (not ForceNew - a lock
// toggle must not replace the cluster). It is not Computed: Computed and
// Default cannot be set together (the SDK rejects that combination at
// InternalValidate time with "Default must be nil if computed").
func TestCloudConfig_UpdateLock_Schema(t *testing.T) {
	res := resourceClusterEdgeNative()
	require.False(t, res.Schema["cloud_config"].ForceNew)

	cloudConfigElem := res.Schema["cloud_config"].Elem.(*schema.Resource)
	fieldSchema, ok := cloudConfigElem.Schema["update_lock"]
	require.True(t, ok, "cloud_config must define update_lock")
	assert.False(t, fieldSchema.ForceNew)
	assert.True(t, fieldSchema.Optional)
	assert.False(t, fieldSchema.Computed)
	assert.Equal(t, "disabled", fieldSchema.Default)
}

// TestToUpdateLockState covers the helper directly: a valid enabled/disabled
// value round-trips with ok=true, and an unset/empty value returns ok=false
// so callers leave the backend's current/auto-engaged state untouched.
func TestToUpdateLockState(t *testing.T) {
	state, ok := toUpdateLockState(map[string]interface{}{"update_lock": "enabled"})
	require.True(t, ok)
	assert.Equal(t, models.V1UpdateLockStateEnabled, state)

	state, ok = toUpdateLockState(map[string]interface{}{"update_lock": "disabled"})
	require.True(t, ok)
	assert.Equal(t, models.V1UpdateLockStateDisabled, state)

	_, ok = toUpdateLockState(map[string]interface{}{"update_lock": ""})
	assert.False(t, ok)

	_, ok = toUpdateLockState(map[string]interface{}{})
	assert.False(t, ok)
}

// TestToEdgeNativeCluster_UpdateLockInCreateBody confirms a configured
// update_lock is sent on Day 1 as part of the cluster create body's
// cloud config.
func TestToEdgeNativeCluster_UpdateLockInCreateBody(t *testing.T) {
	d := prepareEdgeNativeClusterResourceData(t)
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"vip":                "10.0.0.100",
			"overlay_cidr_range": "",
			"update_lock":        "disabled",
		},
	}))

	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	cluster, err := toEdgeNativeCluster(c, d)
	require.NoError(t, err)
	require.NotNil(t, cluster.Spec.CloudConfig.UpdateLock)
	assert.Equal(t, models.V1UpdateLockStateDisabled, cluster.Spec.CloudConfig.UpdateLock.State)
}

// TestToEdgeNativeCluster_UpdateLockUnsetNotInCreateBody confirms an unset
// update_lock is omitted from the create body, leaving Palette's default.
func TestToEdgeNativeCluster_UpdateLockUnsetNotInCreateBody(t *testing.T) {
	d := prepareEdgeNativeClusterResourceData(t)

	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	cluster, err := toEdgeNativeCluster(c, d)
	require.NoError(t, err)
	assert.Nil(t, cluster.Spec.CloudConfig.UpdateLock)
}

// TestToClusterConfigEdgeNativeUpdate_UpdateLockNotIncluded confirms
// update_lock is never bundled into the general Day-2 cloud config PUT body
// either - it must go only through UpdateLockEdgeNative, so the permission
// to change the lock stays separate from the permission to update the rest
// of cloud_config (epic requirement: RBAC separation).
func TestToClusterConfigEdgeNativeUpdate_UpdateLockNotIncluded(t *testing.T) {
	d := resourceClusterEdgeNative().TestResourceData()
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"vip":                "10.0.0.100",
			"overlay_cidr_range": "",
			"ssh_keys":           []interface{}{"ssh-rsa AAAA..."},
			"update_lock":        "enabled",
		},
	}))

	entity, err := toClusterConfigEdgeNativeUpdate(d)
	require.NoError(t, err)
	assert.Nil(t, entity.ClusterConfig.UpdateLock)
}

// TestFlattenClusterConfigsEdgeNative_UpdateLock is the Read-side guard: a
// populated API UpdateLock.State must surface as cloud_config's update_lock
// in state.
func TestFlattenClusterConfigsEdgeNative_UpdateLock(t *testing.T) {
	config := &models.V1EdgeNativeCloudConfig{
		Spec: &models.V1EdgeNativeCloudConfigSpec{
			ClusterConfig: &models.V1EdgeNativeClusterConfig{
				UpdateLock: &models.V1UpdateLock{State: models.V1UpdateLockStateEnabled},
			},
		},
	}

	flattened := flattenClusterConfigsEdgeNative(config)
	require.Len(t, flattened, 1)
	m := flattened[0].(map[string]interface{})
	assert.Equal(t, "enabled", m["update_lock"])
}

// TestFlattenClusterConfigsEdgeNative_UpdateLock_Nil confirms a cluster with
// no update lock state doesn't synthesize the key into state.
func TestFlattenClusterConfigsEdgeNative_UpdateLock_Nil(t *testing.T) {
	config := &models.V1EdgeNativeCloudConfig{
		Spec: &models.V1EdgeNativeCloudConfigSpec{
			ClusterConfig: &models.V1EdgeNativeClusterConfig{},
		},
	}

	flattened := flattenClusterConfigsEdgeNative(config)
	require.Len(t, flattened, 1)
	m := flattened[0].(map[string]interface{})
	_, ok := m["update_lock"]
	assert.False(t, ok)
}

// TestFlattenCloudConfigEdgeNative_UpdateLock drives the full Read helper
// (GetCloudConfigEdgeNative -> flatten) against the dedicated mock fixture,
// proving the end-to-end wiring, not just the pure transform.
func TestFlattenCloudConfigEdgeNative_UpdateLock(t *testing.T) {
	d := resourceClusterEdgeNative().TestResourceData()
	d.SetId("test-cluster-uid")
	require.NoError(t, d.Set("context", "project"))

	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	diags := flattenCloudConfigEdgeNative(routes.EdgeNativeCloudConfigUpdateLockUID, d, c)
	assert.False(t, diags.HasError(), "diags: %+v", diags)

	cloudConfig := d.Get("cloud_config").([]interface{})
	require.Len(t, cloudConfig, 1)
	cc := cloudConfig[0].(map[string]interface{})
	assert.Equal(t, "enabled", cc["update_lock"])
}

// TestResourceClusterEdgeNativeUpdate_UpdateLock_Day2 is the end-to-end day-2
// guard: changing only update_lock produces a cloud_config diff (not a
// destroy/recreate) and resourceClusterEdgeNativeUpdate calls
// UpdateLockEdgeNative successfully, through the dedicated endpoint rather
// than UpdateCloudConfigEdgeNative.
func TestResourceClusterEdgeNativeUpdate_UpdateLock_Day2(t *testing.T) {
	oldRaw := baseEdgeNativeRaw(edgeNativeCloudConfigRawMap(map[string]interface{}{
		"update_lock": "disabled",
	}))
	newRaw := baseEdgeNativeRaw(edgeNativeCloudConfigRawMap(map[string]interface{}{
		"update_lock": "enabled",
	}))

	d := buildEdgeNativeUpdateResourceData(t, oldRaw, newRaw, edgeNativeCloudConfigUID)
	require.True(t, d.HasChange("cloud_config.0.update_lock"))

	diags := resourceClusterEdgeNativeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterEdgeNativeUpdate_UpdateLockOnly_SkipsCloudConfigPut
// confirms a lock-only change does not also send the general cloud config
// PUT: the config UID makes that PUT fail, so success proves it was skipped.
func TestResourceClusterEdgeNativeUpdate_UpdateLockOnly_SkipsCloudConfigPut(t *testing.T) {
	oldRaw := baseEdgeNativeRaw(edgeNativeCloudConfigRawMap(map[string]interface{}{
		"update_lock": "disabled",
	}))
	newRaw := baseEdgeNativeRaw(edgeNativeCloudConfigRawMap(map[string]interface{}{
		"update_lock": "enabled",
	}))

	d := buildEdgeNativeUpdateResourceData(t, oldRaw, newRaw, routes.EdgeNativeCloudConfigUpdateRejectUID)
	require.True(t, d.HasChange("cloud_config"))
	require.False(t, hasCloudConfigChangeExceptUpdateLock(d))

	diags := resourceClusterEdgeNativeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterEdgeNativeUpdate_UpdateLockAndOtherField_SendsCloudConfigPut
// confirms a change to another cloud_config field alongside update_lock still
// goes through the general cloud config PUT.
func TestResourceClusterEdgeNativeUpdate_UpdateLockAndOtherField_SendsCloudConfigPut(t *testing.T) {
	oldRaw := baseEdgeNativeRaw(edgeNativeCloudConfigRawMap(map[string]interface{}{
		"update_lock": "disabled",
	}))
	newRaw := baseEdgeNativeRaw(edgeNativeCloudConfigRawMap(map[string]interface{}{
		"update_lock": "enabled",
		"ssh_keys":    []interface{}{"ssh-rsa BBBB..."},
	}))

	d := buildEdgeNativeUpdateResourceData(t, oldRaw, newRaw, routes.EdgeNativeCloudConfigUpdateRejectUID)
	require.True(t, hasCloudConfigChangeExceptUpdateLock(d))

	diags := resourceClusterEdgeNativeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError(), "expected the general cloud config PUT to be sent")
}

// TestResourceClusterEdgeNativeUpdate_UpdateLock_RejectedByAPI confirms a
// backend rejection of the lock-state change is surfaced as an apply-time
// error instead of being swallowed.
func TestResourceClusterEdgeNativeUpdate_UpdateLock_RejectedByAPI(t *testing.T) {
	oldRaw := baseEdgeNativeRaw(edgeNativeCloudConfigRawMap(map[string]interface{}{
		"update_lock": "disabled",
	}))
	newRaw := baseEdgeNativeRaw(edgeNativeCloudConfigRawMap(map[string]interface{}{
		"update_lock": "enabled",
	}))

	d := buildEdgeNativeUpdateResourceData(t, oldRaw, newRaw, routes.EdgeNativeUpdateLockRejectUID)
	require.True(t, d.HasChange("cloud_config.0.update_lock"))

	diags := resourceClusterEdgeNativeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError(), "expected the rejected update_lock change to surface as an error")
}

// TestResourceClusterEdgeNativeUpdate_UpdateLock_NoChangeNoOp confirms an
// unrelated cloud_config change (e.g. ssh_keys) doesn't also trigger an
// UpdateLockEdgeNative call when update_lock itself is unchanged.
func TestResourceClusterEdgeNativeUpdate_UpdateLock_NoChangeNoOp(t *testing.T) {
	oldRaw := baseEdgeNativeRaw(edgeNativeCloudConfigRawMap(map[string]interface{}{
		"update_lock": "enabled",
	}))
	newRaw := baseEdgeNativeRaw(edgeNativeCloudConfigRawMap(map[string]interface{}{
		"update_lock": "enabled",
		"ssh_keys":    []interface{}{"ssh-rsa BBBB..."},
	}))

	d := buildEdgeNativeUpdateResourceData(t, oldRaw, newRaw, edgeNativeCloudConfigUID)
	require.True(t, d.HasChange("cloud_config"))
	require.False(t, d.HasChange("cloud_config.0.update_lock"))

	diags := resourceClusterEdgeNativeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceClusterEdgeNativeCreate_UpdateLock exercises the day-1 path: an
// explicitly configured update_lock goes out in the create body and the
// create completes without a separate UpdateLockEdgeNative call.
func TestResourceClusterEdgeNativeCreate_UpdateLock(t *testing.T) {
	d := prepareEdgeNativeClusterResourceData(t)
	require.NoError(t, d.Set("tags", []interface{}{"skip_completion"}))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"vip":                "10.0.0.100",
			"overlay_cidr_range": "",
			"update_lock":        "disabled",
		},
	}))

	diags := resourceClusterEdgeNativeCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)

	cloudConfig := d.Get("cloud_config").([]interface{})
	require.Len(t, cloudConfig, 1)
	cc := cloudConfig[0].(map[string]interface{})
	assert.Equal(t, "disabled", cc["update_lock"])
}

// TestResourceClusterEdgeNativeCreate_UpdateLockUnset confirms Create succeeds
// when update_lock is left unset, leaving Palette's default untouched.
func TestResourceClusterEdgeNativeCreate_UpdateLockUnset(t *testing.T) {
	d := prepareEdgeNativeClusterResourceData(t)
	require.NoError(t, d.Set("tags", []interface{}{"skip_completion"}))

	diags := resourceClusterEdgeNativeCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}
