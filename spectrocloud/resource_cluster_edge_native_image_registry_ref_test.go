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

// PLT-2484 adds `image_registry_ref` to spectrocloud_cluster_edge_native's
// `cloud_config`, mapped to the API's clusterConfig.imageRegistryRef - the
// OCI registry (provider_type=image) edge hosts use to pull the Edge Native
// provider (OS/bootstrap) image. It must be settable on create (day 0) and
// changeable later (day 2) without replacing the cluster.

// TestCloudConfig_ImageRegistryRef_NotForceNew is the schema-level guard for
// the "no replace" requirement: neither the field itself nor its parent
// cloud_config block may be ForceNew, or any day-2 change would destroy and
// recreate the cluster instead of calling UpdateCloudConfigEdgeNative.
func TestCloudConfig_ImageRegistryRef_NotForceNew(t *testing.T) {
	res := resourceClusterEdgeNative()
	require.False(t, res.Schema["cloud_config"].ForceNew)

	cloudConfigElem := res.Schema["cloud_config"].Elem.(*schema.Resource)
	fieldSchema, ok := cloudConfigElem.Schema["image_registry_ref"]
	require.True(t, ok, "cloud_config must define image_registry_ref")
	assert.False(t, fieldSchema.ForceNew)
	assert.True(t, fieldSchema.Optional)
}

// TestToImageRegistryRef covers the helper directly: a non-empty UID builds
// a V1ObjectReference, and an unset/empty value returns nil so the
// (omitempty) API field is left out of Create/Update requests that don't use
// this feature at all.
func TestToImageRegistryRef(t *testing.T) {
	ref := toImageRegistryRef(map[string]interface{}{"image_registry_ref": "reg-uid-123"})
	require.NotNil(t, ref)
	assert.Equal(t, "reg-uid-123", ref.UID)

	assert.Nil(t, toImageRegistryRef(map[string]interface{}{"image_registry_ref": ""}))
	assert.Nil(t, toImageRegistryRef(map[string]interface{}{}))
}

// TestToEdgeNativeCluster_ImageRegistryRef is the day-0 (create) guard:
// cloud_config.image_registry_ref must land on
// cluster.Spec.CloudConfig.ImageRegistryRef in the Create request.
func TestToEdgeNativeCluster_ImageRegistryRef(t *testing.T) {
	d := prepareEdgeNativeClusterResourceData(t)
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"vip":                "10.0.0.100",
			"overlay_cidr_range": "",
			"image_registry_ref": routes.EdgeNativeImageRegistryRefUID,
		},
	}))

	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	cluster, err := toEdgeNativeCluster(c, d)
	require.NoError(t, err)
	require.NotNil(t, cluster.Spec.CloudConfig.ImageRegistryRef)
	assert.Equal(t, routes.EdgeNativeImageRegistryRefUID, cluster.Spec.CloudConfig.ImageRegistryRef.UID)
}

// TestToEdgeNativeCluster_ImageRegistryRef_Unset confirms clusters that
// don't use this field are unaffected (nil, not an empty reference object).
func TestToEdgeNativeCluster_ImageRegistryRef_Unset(t *testing.T) {
	d := prepareEdgeNativeClusterResourceData(t)
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	cluster, err := toEdgeNativeCluster(c, d)
	require.NoError(t, err)
	assert.Nil(t, cluster.Spec.CloudConfig.ImageRegistryRef)
}

// TestToClusterConfigEdgeNativeUpdate_ImageRegistryRef is the day-2 (update)
// guard: the same field must reach the
// PUT /v1/cloudconfigs/edge-native/{uid}/clusterConfig payload.
func TestToClusterConfigEdgeNativeUpdate_ImageRegistryRef(t *testing.T) {
	d := resourceClusterEdgeNative().TestResourceData()
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"vip":                "10.0.0.100",
			"overlay_cidr_range": "",
			"ssh_keys":           []interface{}{"ssh-rsa AAAA..."},
			"image_registry_ref": routes.EdgeNativeImageRegistryRefUID,
		},
	}))

	entity, err := toClusterConfigEdgeNativeUpdate(d)
	require.NoError(t, err)
	require.NotNil(t, entity.ClusterConfig.ImageRegistryRef)
	assert.Equal(t, routes.EdgeNativeImageRegistryRefUID, entity.ClusterConfig.ImageRegistryRef.UID)
}

// TestFlattenClusterConfigsEdgeNative_ImageRegistryRef is the Read-side
// guard: a populated API ImageRegistryRef must surface as cloud_config's
// image_registry_ref in state.
func TestFlattenClusterConfigsEdgeNative_ImageRegistryRef(t *testing.T) {
	config := &models.V1EdgeNativeCloudConfig{
		Spec: &models.V1EdgeNativeCloudConfigSpec{
			ClusterConfig: &models.V1EdgeNativeClusterConfig{
				ImageRegistryRef: &models.V1ObjectReference{UID: routes.EdgeNativeImageRegistryRefUID},
			},
		},
	}

	flattened := flattenClusterConfigsEdgeNative(config)
	require.Len(t, flattened, 1)
	m := flattened[0].(map[string]interface{})
	assert.Equal(t, routes.EdgeNativeImageRegistryRefUID, m["image_registry_ref"])
}

// TestFlattenClusterConfigsEdgeNative_ImageRegistryRef_Nil confirms a
// cluster with no image registry ref doesn't synthesize the key into state.
func TestFlattenClusterConfigsEdgeNative_ImageRegistryRef_Nil(t *testing.T) {
	config := &models.V1EdgeNativeCloudConfig{
		Spec: &models.V1EdgeNativeCloudConfigSpec{
			ClusterConfig: &models.V1EdgeNativeClusterConfig{},
		},
	}

	flattened := flattenClusterConfigsEdgeNative(config)
	require.Len(t, flattened, 1)
	m := flattened[0].(map[string]interface{})
	_, ok := m["image_registry_ref"]
	assert.False(t, ok)
}

// TestFlattenCloudConfigEdgeNative_ImageRegistryRef drives the full Read
// helper (GetCloudConfigEdgeNative -> flatten) against the dedicated mock
// fixture, proving the end-to-end wiring, not just the pure transforms.
func TestFlattenCloudConfigEdgeNative_ImageRegistryRef(t *testing.T) {
	d := resourceClusterEdgeNative().TestResourceData()
	d.SetId("test-cluster-uid")
	require.NoError(t, d.Set("context", "project"))

	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	diags := flattenCloudConfigEdgeNative(routes.EdgeNativeCloudConfigImageRegistryRefUID, d, c)
	assert.False(t, diags.HasError(), "diags: %+v", diags)

	cloudConfig := d.Get("cloud_config").([]interface{})
	require.Len(t, cloudConfig, 1)
	cc := cloudConfig[0].(map[string]interface{})
	assert.Equal(t, routes.EdgeNativeImageRegistryRefUID, cc["image_registry_ref"])
}

// TestResourceClusterEdgeNativeUpdate_ImageRegistryRef_Day2NoReplace is the
// end-to-end day-2 guard: changing only image_registry_ref produces a
// cloud_config diff (not a destroy/recreate, since cloud_config isn't
// ForceNew) and resourceClusterEdgeNativeUpdate calls
// UpdateCloudConfigEdgeNative successfully. Mirrors
// TestResourceClusterEdgeNativeUpdateCloudConfigWithMock in
// resource_cluster_edge_native_cloud_config_update_test.go.
func TestResourceClusterEdgeNativeUpdate_ImageRegistryRef_Day2NoReplace(t *testing.T) {
	oldRaw := baseEdgeNativeRaw(edgeNativeCloudConfigRawMap(nil))
	newRaw := baseEdgeNativeRaw(edgeNativeCloudConfigRawMap(map[string]interface{}{
		"image_registry_ref": routes.EdgeNativeImageRegistryRefUID,
	}))

	d := buildEdgeNativeUpdateResourceData(t, oldRaw, newRaw, edgeNativeCloudConfigUID)
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterEdgeNativeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}
