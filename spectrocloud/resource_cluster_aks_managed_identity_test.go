package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// PLT-2489 adds user-assigned managed identity support to
// spectrocloud_cluster_aks's cloud_config: aks_service_identity_resource_id
// (AKS control plane) and kubelet_identity_resource_id (AKS kubelet). No
// clusterIdentity (IaaS-only) and no machine_pool-level fields
// (user_assigned_identities/exclude_cluster_identity are IaaS pools only
// per PEM-12252) - see resource_cluster_azure_managed_identity_test.go for
// those.

// TestAksCloudConfig_ManagedIdentity_ForceNew guards immutability.
func TestAksCloudConfig_ManagedIdentity_ForceNew(t *testing.T) {
	res := resourceClusterAks()
	cloudConfigElem := res.Schema["cloud_config"].Elem.(*schema.Resource)
	fieldSchema, ok := cloudConfigElem.Schema["managed_identity"]
	require.True(t, ok)
	assert.True(t, fieldSchema.ForceNew)

	managedIdentityElem := fieldSchema.Elem.(*schema.Resource)
	aksServiceIdentitySchema, ok := managedIdentityElem.Schema["aks_service_identity_resource_id"]
	require.True(t, ok)
	assert.True(t, aksServiceIdentitySchema.ForceNew)
	kubeletIdentitySchema, ok := managedIdentityElem.Schema["kubelet_identity_resource_id"]
	require.True(t, ok)
	assert.True(t, kubeletIdentitySchema.ForceNew)
}

// TestAksMachinePool_NoIdentityFields confirms the IaaS-only
// user_assigned_identities/exclude_cluster_identity fields were NOT added
// to AKS's machine_pool, per PEM-12252 ("IaaS pools only").
func TestAksMachinePool_NoIdentityFields(t *testing.T) {
	res := resourceClusterAks()
	machinePoolElem := res.Schema["machine_pool"].Elem.(*schema.Resource)
	_, hasUAI := machinePoolElem.Schema["user_assigned_identities"]
	assert.False(t, hasUAI)
	_, hasECI := machinePoolElem.Schema["exclude_cluster_identity"]
	assert.False(t, hasECI)
}

func TestToAksManagedIdentity(t *testing.T) {
	t.Run("both set", func(t *testing.T) {
		cloudConfig := map[string]interface{}{
			"managed_identity": []interface{}{
				map[string]interface{}{
					"aks_service_identity_resource_id": "test-aks-service-identity-id",
					"kubelet_identity_resource_id":     "test-kubelet-identity-id",
				},
			},
		}
		got := toAksManagedIdentity(cloudConfig)
		require.NotNil(t, got)
		require.NotNil(t, got.AksServiceIdentity)
		assert.Equal(t, "test-aks-service-identity-id", *got.AksServiceIdentity.ResourceID)
		require.NotNil(t, got.KubeletIdentity)
		assert.Equal(t, "test-kubelet-identity-id", *got.KubeletIdentity.ResourceID)
	})

	t.Run("only aks_service_identity set", func(t *testing.T) {
		cloudConfig := map[string]interface{}{
			"managed_identity": []interface{}{
				map[string]interface{}{"aks_service_identity_resource_id": "test-aks-service-identity-id"},
			},
		}
		got := toAksManagedIdentity(cloudConfig)
		require.NotNil(t, got)
		require.NotNil(t, got.AksServiceIdentity)
		assert.Nil(t, got.KubeletIdentity)
	})

	t.Run("unset", func(t *testing.T) {
		assert.Nil(t, toAksManagedIdentity(map[string]interface{}{}))
	})

	t.Run("empty list", func(t *testing.T) {
		assert.Nil(t, toAksManagedIdentity(map[string]interface{}{"managed_identity": []interface{}{}}))
	})

	t.Run("block present but both fields empty", func(t *testing.T) {
		cloudConfig := map[string]interface{}{
			"managed_identity": []interface{}{
				map[string]interface{}{"aks_service_identity_resource_id": "", "kubelet_identity_resource_id": ""},
			},
		}
		assert.Nil(t, toAksManagedIdentity(cloudConfig))
	})
}

// TestToAksClusterConfig_ManagedIdentity confirms the shared cloud_config
// transform (used by both Create and Update via toAksClusterConfig) wires
// ManagedIdentity through.
func TestToAksClusterConfig_ManagedIdentity(t *testing.T) {
	cloudConfigMap := map[string]interface{}{
		"region":                    "eastus",
		"resource_group":            "test-rg",
		"ssh_key":                   "test-ssh-key",
		"private_cluster":           false,
		"subscription_id":           "test-subscription-id",
		"control_plane_subnet_name": "",
		"control_plane_cidr":        "",
		"managed_identity": []interface{}{
			map[string]interface{}{"aks_service_identity_resource_id": "test-aks-service-identity-id"},
		},
	}
	got := toAksClusterConfig(cloudConfigMap)
	require.NotNil(t, got.ManagedIdentity)
	require.NotNil(t, got.ManagedIdentity.AksServiceIdentity)
	assert.Equal(t, "test-aks-service-identity-id", *got.ManagedIdentity.AksServiceIdentity.ResourceID)
}

func TestFlattenAksManagedIdentity(t *testing.T) {
	t.Run("both set", func(t *testing.T) {
		aksID := "test-aks-service-identity-id"
		kubeletID := "test-kubelet-identity-id"
		got := flattenAksManagedIdentity(&models.V1AzureManagedIdentity{
			AksServiceIdentity: &models.V1AzureIdentityRef{ResourceID: &aksID},
			KubeletIdentity:    &models.V1AzureIdentityRef{ResourceID: &kubeletID},
		})
		require.Len(t, got, 1)
		entry := got[0].(map[string]interface{})
		assert.Equal(t, aksID, entry["aks_service_identity_resource_id"])
		assert.Equal(t, kubeletID, entry["kubelet_identity_resource_id"])
	})

	t.Run("nil", func(t *testing.T) {
		assert.Nil(t, flattenAksManagedIdentity(nil))
	})

	t.Run("empty", func(t *testing.T) {
		assert.Nil(t, flattenAksManagedIdentity(&models.V1AzureManagedIdentity{}))
	})
}

// TestFlattenClusterConfigsAks_ManagedIdentity is the full Read-side guard
// through flattenClusterConfigsAks.
func TestFlattenClusterConfigsAks_ManagedIdentity(t *testing.T) {
	aksID := "test-aks-service-identity-id"
	config := &models.V1AzureCloudConfig{
		Spec: &models.V1AzureCloudConfigSpec{
			ClusterConfig: &models.V1AzureClusterConfig{
				APIServerAccessProfile: &models.V1APIServerAccessProfile{},
				ManagedIdentity: &models.V1AzureManagedIdentity{
					AksServiceIdentity: &models.V1AzureIdentityRef{ResourceID: &aksID},
				},
			},
		},
	}
	flattened := flattenClusterConfigsAks(config)
	require.Len(t, flattened, 1)
	m := flattened[0].(map[string]interface{})
	managedIdentity, ok := m["managed_identity"].([]interface{})
	require.True(t, ok)
	require.Len(t, managedIdentity, 1)
	entry := managedIdentity[0].(map[string]interface{})
	assert.Equal(t, aksID, entry["aks_service_identity_resource_id"])
}

// TestFlattenClusterConfigsAks_ManagedIdentity_Nil confirms a cluster with
// no managed identity doesn't synthesize the key into state.
func TestFlattenClusterConfigsAks_ManagedIdentity_Nil(t *testing.T) {
	config := &models.V1AzureCloudConfig{
		Spec: &models.V1AzureCloudConfigSpec{
			ClusterConfig: &models.V1AzureClusterConfig{
				APIServerAccessProfile: &models.V1APIServerAccessProfile{},
			},
		},
	}
	flattened := flattenClusterConfigsAks(config)
	require.Len(t, flattened, 1)
	m := flattened[0].(map[string]interface{})
	_, ok := m["managed_identity"]
	assert.False(t, ok)
}
