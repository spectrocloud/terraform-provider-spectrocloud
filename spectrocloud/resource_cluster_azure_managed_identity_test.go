package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// PLT-2489 adds user-assigned managed identity support to
// spectrocloud_cluster_azure (IaaS): cloud_config.managed_identity
// (cluster_identity only - aks_service_identity/kubelet_identity are
// AKS-only) and machine_pool.user_assigned_identities /
// exclude_cluster_identity. All four fields are immutable after creation
// per the API contract (PEM-12252).

// TestAzureCloudConfig_ManagedIdentity_ForceNew and
// TestAzureMachinePool_IdentityFields_ForceNew guard the immutability
// requirement at the schema level.
func TestAzureCloudConfig_ManagedIdentity_ForceNew(t *testing.T) {
	res := resourceClusterAzure()
	cloudConfigElem := res.Schema["cloud_config"].Elem.(*schema.Resource)
	fieldSchema, ok := cloudConfigElem.Schema["managed_identity"]
	require.True(t, ok)
	assert.True(t, fieldSchema.ForceNew)

	managedIdentityElem := fieldSchema.Elem.(*schema.Resource)
	resourceIDSchema, ok := managedIdentityElem.Schema["cluster_identity_resource_id"]
	require.True(t, ok)
	assert.True(t, resourceIDSchema.ForceNew)
}

func TestAzureMachinePool_IdentityFields_ForceNew(t *testing.T) {
	res := resourceClusterAzure()
	machinePoolElem := res.Schema["machine_pool"].Elem.(*schema.Resource)

	uaiSchema, ok := machinePoolElem.Schema["user_assigned_identities"]
	require.True(t, ok)
	assert.True(t, uaiSchema.ForceNew)

	eciSchema, ok := machinePoolElem.Schema["exclude_cluster_identity"]
	require.True(t, ok)
	assert.True(t, eciSchema.ForceNew)
	assert.Equal(t, false, eciSchema.Default)
}

// TestToAzureManagedIdentity covers the cloud_config-level transform used by
// both Create (toAzureCluster) and Update (toCloudConfigAzure), since both
// funnel through azureClusterConfigFromMap.
func TestToAzureManagedIdentity(t *testing.T) {
	t.Run("set", func(t *testing.T) {
		cloudConfig := map[string]interface{}{
			"managed_identity": []interface{}{
				map[string]interface{}{"cluster_identity_resource_id": "test-identity-resource-id"},
			},
		}
		got := toAzureManagedIdentity(cloudConfig)
		require.NotNil(t, got)
		require.NotNil(t, got.ClusterIdentity)
		require.NotNil(t, got.ClusterIdentity.ResourceID)
		assert.Equal(t, "test-identity-resource-id", *got.ClusterIdentity.ResourceID)
	})

	t.Run("unset", func(t *testing.T) {
		assert.Nil(t, toAzureManagedIdentity(map[string]interface{}{}))
	})

	t.Run("empty list", func(t *testing.T) {
		cloudConfig := map[string]interface{}{"managed_identity": []interface{}{}}
		assert.Nil(t, toAzureManagedIdentity(cloudConfig))
	})
}

// TestAzureClusterConfigFromMap_ManagedIdentity confirms the cloud_config
// transform (shared by Create and Update) wires ManagedIdentity through.
func TestAzureClusterConfigFromMap_ManagedIdentity(t *testing.T) {
	cloudConfig := map[string]interface{}{
		"region":               "eastus",
		"ssh_key":              "test-ssh-key",
		"subscription_id":      "test-subscription-id",
		"resource_group":       "test-rg",
		"storage_account_name": "",
		"container_name":       "",
		"managed_identity":     []interface{}{map[string]interface{}{"cluster_identity_resource_id": "test-identity-resource-id"}},
	}
	got := azureClusterConfigFromMap(cloudConfig)
	require.NotNil(t, got.ManagedIdentity)
	require.NotNil(t, got.ManagedIdentity.ClusterIdentity)
	assert.Equal(t, "test-identity-resource-id", *got.ManagedIdentity.ClusterIdentity.ResourceID)
}

// TestFlattenClusterConfigsAzure_ManagedIdentity is the Read-side guard.
func TestFlattenClusterConfigsAzure_ManagedIdentity(t *testing.T) {
	resourceID := "test-identity-resource-id"
	config := &models.V1AzureCloudConfig{
		Spec: &models.V1AzureCloudConfigSpec{
			ClusterConfig: &models.V1AzureClusterConfig{
				ManagedIdentity: &models.V1AzureManagedIdentity{
					ClusterIdentity: &models.V1AzureIdentityRef{ResourceID: &resourceID},
				},
			},
		},
	}
	flattened := flattenClusterConfigsAzure(config)
	require.Len(t, flattened, 1)
	m := flattened[0].(map[string]interface{})
	managedIdentity, ok := m["managed_identity"].([]interface{})
	require.True(t, ok)
	require.Len(t, managedIdentity, 1)
	entry := managedIdentity[0].(map[string]interface{})
	assert.Equal(t, resourceID, entry["cluster_identity_resource_id"])
}

// TestFlattenClusterConfigsAzure_ManagedIdentity_Nil confirms a cluster with
// no managed identity doesn't synthesize the key into state.
func TestFlattenClusterConfigsAzure_ManagedIdentity_Nil(t *testing.T) {
	config := &models.V1AzureCloudConfig{
		Spec: &models.V1AzureCloudConfigSpec{
			ClusterConfig: &models.V1AzureClusterConfig{},
		},
	}
	flattened := flattenClusterConfigsAzure(config)
	require.Len(t, flattened, 1)
	m := flattened[0].(map[string]interface{})
	_, ok := m["managed_identity"]
	assert.False(t, ok)
}

// TestToMachinePoolAzure_IdentityFields covers the machine-pool-level
// transform: user_assigned_identities and exclude_cluster_identity must
// reach V1AzureMachinePoolCloudConfigEntity.
func TestToMachinePoolAzure_IdentityFields(t *testing.T) {
	m := defaultAzureMachinePool(map[string]interface{}{
		"exclude_cluster_identity": true,
		"user_assigned_identities": schema.NewSet(schema.HashString, []interface{}{"extra-identity-resource-id"}),
	})
	mp, err := toMachinePoolAzure(m)
	require.NoError(t, err)
	require.NotNil(t, mp.CloudConfig.ExcludeClusterIdentity)
	assert.True(t, *mp.CloudConfig.ExcludeClusterIdentity)
	require.Len(t, mp.CloudConfig.UserAssignedIdentities, 1)
	assert.Equal(t, "extra-identity-resource-id", *mp.CloudConfig.UserAssignedIdentities[0].ResourceID)
}

// TestToMachinePoolAzure_IdentityFields_Defaults confirms a pool that
// doesn't set these fields gets a non-nil false ExcludeClusterIdentity
// (matching the schema's Default: false) and no UserAssignedIdentities.
func TestToMachinePoolAzure_IdentityFields_Defaults(t *testing.T) {
	m := defaultAzureMachinePool(nil)
	mp, err := toMachinePoolAzure(m)
	require.NoError(t, err)
	require.NotNil(t, mp.CloudConfig.ExcludeClusterIdentity)
	assert.False(t, *mp.CloudConfig.ExcludeClusterIdentity)
	assert.Empty(t, mp.CloudConfig.UserAssignedIdentities)
}

// TestFlattenMachinePoolConfigsAzure_IdentityFields is the Read-side guard
// for the machine-pool-level fields.
func TestFlattenMachinePoolConfigsAzure_IdentityFields(t *testing.T) {
	excludeClusterIdentity := true
	resourceID := "extra-identity-resource-id"
	pools := []*models.V1AzureMachinePoolConfig{
		{
			Name:                   "worker-pool",
			ExcludeClusterIdentity: &excludeClusterIdentity,
			UserAssignedIdentities: []*models.V1AzureIdentityRef{{ResourceID: &resourceID}},
		},
	}
	flattened := flattenMachinePoolConfigsAzure(pools)
	require.Len(t, flattened, 1)
	oi := flattened[0].(map[string]interface{})
	assert.Equal(t, true, oi["exclude_cluster_identity"])
	uai, ok := oi["user_assigned_identities"].(*schema.Set)
	require.True(t, ok)
	assert.Equal(t, []interface{}{resourceID}, uai.List())
}
