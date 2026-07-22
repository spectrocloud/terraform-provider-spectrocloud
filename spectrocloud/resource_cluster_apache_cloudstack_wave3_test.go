package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Wave 3 — pushing coverage on the six lowest-covered funcs in
// resource_cluster_apache_cloudstack.go:
//   toCloudStackCluster, resourceClusterApacheCloudStackCreate,
//   flattenClusterConfigsApacheCloudStack, toCloudStackCloudConfigWithResolution,
//   resourceClusterApacheCloudStackImport, resourceClusterApacheCloudStackUpdate.
//
// Mock fixtures used here (see tests/mockApiServer/routes):
//   - zone name "zone-name-1" -> id "zone-id-1", account "test-apache-cloudstack-account-id-1"
//     (mockCloudAccounts.go properties/zones — static regardless of account UID)
//   - cluster UID "test-cloudstack-cluster-id" -> CloudType "apache-cloudstack" (mockCluster.go)
//   - cluster UID "cluster-uid-server-error" -> GET /v1/spectroclusters/{uid} returns 500
// ---------------------------------------------------------------------------

func buildCloudStackBuilderResourceData(t *testing.T, zoneName string) *schema.ResourceData {
	t.Helper()
	d := resourceClusterApacheCloudStack().TestResourceData()
	require.NoError(t, d.Set("name", "cloudstack-cluster-1"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_account_id", "test-apache-cloudstack-account-id-1"))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"ssh_key_name": "test-ssh-key",
			"zone": []interface{}{
				map[string]interface{}{
					"name": zoneName,
				},
			},
		},
	}))
	require.NoError(t, d.Set("machine_pool", schema.NewSet(resourceMachinePoolApacheCloudStackHash, []interface{}{
		map[string]interface{}{
			"name":                    "cp-pool",
			"count":                   1,
			"control_plane":           true,
			"control_plane_as_worker": true,
			"offering":                "Medium Instance",
		},
	})))
	return d
}

// ---------------------------------------------------------------------------
// toCloudStackCluster
// ---------------------------------------------------------------------------

func TestToCloudStackClusterFields(t *testing.T) {
	t.Parallel()
	d := buildCloudStackBuilderResourceData(t, "zone-name-1")
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")

	got, err := toCloudStackCluster(c, d)
	require.NoError(t, err)
	require.NotNil(t, got)

	require.NotNil(t, got.Metadata)
	assert.Equal(t, "cloudstack-cluster-1", got.Metadata.Name)

	require.NotNil(t, got.Spec)
	require.NotNil(t, got.Spec.CloudAccountUID)
	assert.Equal(t, "test-apache-cloudstack-account-id-1", *got.Spec.CloudAccountUID)

	require.NotNil(t, got.Spec.CloudConfig)
	require.Len(t, got.Spec.CloudConfig.Zones, 1)
	assert.Equal(t, "zone-id-1", got.Spec.CloudConfig.Zones[0].ID)
	assert.Equal(t, "zone-name-1", got.Spec.CloudConfig.Zones[0].Name)

	require.Len(t, got.Spec.Machinepoolconfig, 1)
	require.NotNil(t, got.Spec.Machinepoolconfig[0].PoolConfig.Name)
	assert.Equal(t, "cp-pool", *got.Spec.Machinepoolconfig[0].PoolConfig.Name)

	require.NotNil(t, got.Spec.ClusterConfig)
	require.NotNil(t, got.Spec.Policies)
}

func TestToCloudStackClusterZoneResolutionError(t *testing.T) {
	t.Parallel()
	d := buildCloudStackBuilderResourceData(t, "unresolvable-zone-name")
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")

	_, err := toCloudStackCluster(c, d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no CloudStack zone found")
}

func TestToCloudStackClusterProfileError(t *testing.T) {
	t.Parallel()
	d := buildCloudStackBuilderResourceData(t, "zone-name-1")
	// Non-empty d.Id() drives toProfiles -> toProfilesCommon into calling
	// GetClusterWithoutStatus(clusterUID); this UID's mock fixture returns
	// a 500, so the cluster-lookup branch errors out.
	d.SetId("cluster-uid-server-error")
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")

	_, err := toCloudStackCluster(c, d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be retrieved")
}

// ---------------------------------------------------------------------------
// resourceClusterApacheCloudStackCreate
// ---------------------------------------------------------------------------

func TestResourceClusterApacheCloudStackCreate_ValidateOverrideScalingError(t *testing.T) {
	t.Parallel()
	d := buildCloudStackBuilderResourceData(t, "zone-name-1")
	require.NoError(t, d.Set("machine_pool", schema.NewSet(resourceMachinePoolApacheCloudStackHash, []interface{}{
		map[string]interface{}{
			"name":                    "worker",
			"count":                   2,
			"control_plane":           false,
			"control_plane_as_worker": false,
			"offering":                "Medium Instance",
			"update_strategy":         "OverrideScaling",
			// override_scaling intentionally omitted -> validateOverrideScaling error.
		},
	})))

	diags := resourceClusterApacheCloudStackCreate(context.Background(), d, unitTestMockAPIClient)
	require.True(t, diags.HasError())
}

func TestResourceClusterApacheCloudStackCreate_ClusterBuildError(t *testing.T) {
	t.Parallel()
	d := buildCloudStackBuilderResourceData(t, "unresolvable-zone-name-2")

	diags := resourceClusterApacheCloudStackCreate(context.Background(), d, unitTestMockAPIClient)
	require.True(t, diags.HasError())
}

func TestResourceClusterApacheCloudStackCreate_Success(t *testing.T) {
	d := buildCloudStackBuilderResourceData(t, "zone-name-1")

	diags := resourceClusterApacheCloudStackCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "test-cloudstack-cluster-id", d.Id())
}

// ---------------------------------------------------------------------------
// flattenClusterConfigsApacheCloudStack
// ---------------------------------------------------------------------------

func TestFlattenClusterConfigsApacheCloudStackTable(t *testing.T) {
	t.Parallel()

	t.Run("nil config", func(t *testing.T) {
		out := flattenClusterConfigsApacheCloudStack(nil)
		assert.Equal(t, []interface{}{}, out)
	})

	t.Run("nil spec", func(t *testing.T) {
		out := flattenClusterConfigsApacheCloudStack(&models.V1CloudStackCloudConfig{})
		assert.Equal(t, []interface{}{}, out)
	})

	t.Run("nil cluster config", func(t *testing.T) {
		out := flattenClusterConfigsApacheCloudStack(&models.V1CloudStackCloudConfig{
			Spec: &models.V1CloudStackCloudConfigSpec{},
		})
		assert.Equal(t, []interface{}{}, out)
	})

	t.Run("zone without network", func(t *testing.T) {
		cfg := &models.V1CloudStackCloudConfig{
			Spec: &models.V1CloudStackCloudConfigSpec{
				ClusterConfig: &models.V1CloudStackClusterConfig{
					Zones: []*models.V1CloudStackZoneSpec{
						{ID: "zone-2", Name: "zone-two"},
					},
				},
			},
		}
		out := flattenClusterConfigsApacheCloudStack(cfg)
		require.Len(t, out, 1)
		m := out[0].(map[string]interface{})
		zones := m["zone"].([]interface{})
		require.Len(t, zones, 1)
		zone := zones[0].(map[string]interface{})
		_, hasNetwork := zone["network"]
		assert.False(t, hasNetwork)
	})

	t.Run("fully populated", func(t *testing.T) {
		cfg := &models.V1CloudStackCloudConfig{
			Spec: &models.V1CloudStackCloudConfigSpec{
				ClusterConfig: &models.V1CloudStackClusterConfig{
					Project:                  &models.V1CloudStackResource{ID: "proj-1", Name: "my-project"},
					SSHKeyName:               "ssh-key-1",
					ControlPlaneEndpoint:     "10.0.0.50",
					OverrideClusterAPIConfig: "kind: Cluster",
					SyncWithCKS:              true,
					Zones: []*models.V1CloudStackZoneSpec{
						{
							ID:   "zone-1",
							Name: "zone-one",
							Network: &models.V1CloudStackNetworkSpec{
								ID:          "net-1",
								Name:        "net-one",
								Type:        "Isolated",
								Gateway:     "10.0.0.1",
								Netmask:     "255.255.255.0",
								Offering:    "DefaultOffering",
								RoutingMode: "Static",
								Vpc: &models.V1CloudStackVPCSpec{
									ID:       "vpc-1",
									Name:     "vpc-one",
									Cidr:     "10.0.0.0/16",
									Offering: "VpcOffering",
								},
							},
						},
					},
				},
			},
		}

		out := flattenClusterConfigsApacheCloudStack(cfg)
		require.Len(t, out, 1)
		m := out[0].(map[string]interface{})
		assert.Equal(t, "ssh-key-1", m["ssh_key_name"])
		assert.Equal(t, "10.0.0.50", m["control_plane_endpoint"])
		assert.Equal(t, "kind: Cluster", m["override_cluster_api_config"])
		assert.Equal(t, true, m["sync_with_cks"])

		projList := m["project"].([]interface{})
		require.Len(t, projList, 1)
		proj := projList[0].(map[string]interface{})
		assert.Equal(t, "proj-1", proj["id"])
		assert.Equal(t, "my-project", proj["name"])

		zones := m["zone"].([]interface{})
		require.Len(t, zones, 1)
		zone := zones[0].(map[string]interface{})
		assert.Equal(t, "zone-1", zone["id"])
		assert.Equal(t, "zone-one", zone["name"])

		networks := zone["network"].([]interface{})
		require.Len(t, networks, 1)
		network := networks[0].(map[string]interface{})
		assert.Equal(t, "net-1", network["id"])
		assert.Equal(t, "net-one", network["name"])
		assert.Equal(t, "Isolated", network["type"])
		assert.Equal(t, "10.0.0.1", network["gateway"])
		assert.Equal(t, "255.255.255.0", network["netmask"])
		assert.Equal(t, "DefaultOffering", network["offering"])
		assert.Equal(t, "Static", network["routing_mode"])

		vpcs := network["vpc"].([]interface{})
		require.Len(t, vpcs, 1)
		vpc := vpcs[0].(map[string]interface{})
		assert.Equal(t, "vpc-1", vpc["id"])
		assert.Equal(t, "vpc-one", vpc["name"])
		assert.Equal(t, "10.0.0.0/16", vpc["cidr"])
		assert.Equal(t, "VpcOffering", vpc["offering"])
	})
}

// ---------------------------------------------------------------------------
// toCloudStackCloudConfigWithResolution
// ---------------------------------------------------------------------------

func TestToCloudStackCloudConfigWithResolutionZoneNotFoundError(t *testing.T) {
	t.Parallel()
	d := resourceClusterApacheCloudStack().TestResourceData()
	require.NoError(t, d.Set("cloud_account_id", "test-apache-cloudstack-account-id-1"))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"zone": []interface{}{
				map[string]interface{}{
					"name": "no-such-zone",
				},
			},
		},
	}))

	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	_, err := toCloudStackCloudConfigWithResolution(c, d, make(map[string]string))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no CloudStack zone found")
}

func TestToCloudStackCloudConfigWithResolutionZoneCacheHit(t *testing.T) {
	t.Parallel()
	d := resourceClusterApacheCloudStack().TestResourceData()
	require.NoError(t, d.Set("cloud_account_id", "test-apache-cloudstack-account-id-1"))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"zone": []interface{}{
				map[string]interface{}{
					"name": "cache-only-zone",
				},
			},
		},
	}))

	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	// "cache-only-zone" does not exist in the mock's zone list, so a real
	// resolution call would fail; a pre-populated cache entry must short
	// circuit before that call happens.
	cache := map[string]string{"cache-only-zone": "cached-zone-id-777"}

	cfg, err := toCloudStackCloudConfigWithResolution(c, d, cache)
	require.NoError(t, err)
	require.Len(t, cfg.Zones, 1)
	assert.Equal(t, "cached-zone-id-777", cfg.Zones[0].ID)
}

// ---------------------------------------------------------------------------
// resourceClusterApacheCloudStackImport
// ---------------------------------------------------------------------------

func TestResourceClusterApacheCloudStackImport_Success(t *testing.T) {
	d := resourceClusterApacheCloudStack().TestResourceData()
	d.SetId("test-cloudstack-cluster-id:project")

	result, err := resourceClusterApacheCloudStackImport(context.Background(), d, unitTestMockAPIClient)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "test-cloudstack-cluster-id", result[0].Id())
}

func TestResourceClusterApacheCloudStackImport_InvalidIDFormat(t *testing.T) {
	d := resourceClusterApacheCloudStack().TestResourceData()
	d.SetId("invalid-id-without-scope")

	_, err := resourceClusterApacheCloudStackImport(context.Background(), d, unitTestMockAPIClient)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cluster ID format")
}

// ---------------------------------------------------------------------------
// resourceClusterApacheCloudStackUpdate
// ---------------------------------------------------------------------------

func buildCloudStackClusterUpdateResourceData(t *testing.T, oldCPEndpoint, newCPEndpoint string, oldPools, newPools []interface{}) *schema.ResourceData {
	t.Helper()
	res := resourceClusterApacheCloudStack()

	buildRaw := func(cpEndpoint string, pools []interface{}) map[string]interface{} {
		return map[string]interface{}{
			"name":             "test-cluster",
			"cloud_account_id": "test-account",
			"cloud_config_id":  "test-cloud-config-id",
			"cloud_config": []interface{}{
				map[string]interface{}{
					"control_plane_endpoint": cpEndpoint,
					"zone": []interface{}{
						map[string]interface{}{
							"name": "test-zone",
						},
					},
				},
			},
			"machine_pool": pools,
		}
	}

	oldRD := schema.TestResourceDataRaw(t, res.Schema, buildRaw(oldCPEndpoint, oldPools))
	oldRD.SetId("test-cloudstack-cluster-id")
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newConfig := terraform.NewResourceConfigRaw(buildRaw(newCPEndpoint, newPools))

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)
	require.NotNil(t, diff)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId("test-cloudstack-cluster-id")
	return finalRD
}

func TestResourceClusterApacheCloudStackUpdate_CloudConfigAndMachinePool(t *testing.T) {
	oldPools := []interface{}{
		map[string]interface{}{
			"name":                    "cp-pool",
			"count":                   1,
			"offering":                "small-instance",
			"control_plane":           true,
			"control_plane_as_worker": true,
		},
	}
	newPools := []interface{}{
		map[string]interface{}{
			"name":                    "cp-pool",
			"count":                   2,
			"offering":                "small-instance",
			"control_plane":           true,
			"control_plane_as_worker": true,
		},
	}

	d := buildCloudStackClusterUpdateResourceData(t, "10.0.0.1", "10.0.0.2", oldPools, newPools)
	require.True(t, d.HasChange("cloud_config"))
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterApacheCloudStackUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceClusterApacheCloudStackUpdate_RepaveApprovalError(t *testing.T) {
	d := resourceClusterApacheCloudStack().TestResourceData()
	require.NoError(t, d.Set("cloud_account_id", "test-account"))
	require.NoError(t, d.Set("context", "project"))
	d.SetId("cluster-uid-server-error")

	diags := resourceClusterApacheCloudStackUpdate(context.Background(), d, unitTestMockAPIClient)
	require.True(t, diags.HasError())
}
