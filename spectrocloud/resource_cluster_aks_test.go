package spectrocloud

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToMachinePoolAks(t *testing.T) {
	// Simulate a machine pool configuration
	machinePoolConfig := map[string]interface{}{
		"control_plane":        true,
		"count":                3,
		"min":                  1,
		"max":                  5,
		"instance_type":        "Standard_D2s_v3",
		"disk_size_gb":         50,
		"storage_account_type": "Premium_LRS",
		"is_system_node_pool":  true,
		"name":                 "pool-1",
		"update_strategy":      "RollingUpdateScaleOut",
	}

	// Call the toMachinePoolAks function with the machine pool configuration
	mp := toMachinePoolAks(machinePoolConfig)

	// Assertions
	assert.NotNil(t, mp)
	assert.Equal(t, "Standard_D2s_v3", mp.CloudConfig.InstanceType)
	assert.Equal(t, int32(50), mp.CloudConfig.OsDisk.DiskSizeGB)
	assert.Equal(t, "Premium_LRS", mp.CloudConfig.OsDisk.ManagedDisk.StorageAccountType)
	assert.True(t, mp.ManagedPoolConfig.IsSystemNodePool)
	assert.Equal(t, "pool-1", *mp.PoolConfig.Name)
	assert.Equal(t, int32(3), *mp.PoolConfig.Size)
	assert.Equal(t, int32(1), mp.PoolConfig.MinSize)
	assert.Equal(t, int32(5), mp.PoolConfig.MaxSize)
}

func TestFlattenClusterConfigsAks(t *testing.T) {
	// Simulate the Azure cloud configuration
	azureCloudConfig := &models.V1AzureCloudConfig{
		Spec: &models.V1AzureCloudConfigSpec{
			ClusterConfig: &models.V1AzureClusterConfig{
				SubscriptionID: StringPtr("mySubscriptionID"),
				ResourceGroup:  "myResourceGroup",
				Location:       StringPtr("eastus"),
				SSHKey:         StringPtr("sshPublicKey"),
				APIServerAccessProfile: &models.V1APIServerAccessProfile{
					EnablePrivateCluster: true,
				},
				VnetName:          "myVnet",
				VnetResourceGroup: "myVnetResourceGroup",
				VnetCidrBlock:     "10.0.0.0/16",
				WorkerSubnet: &models.V1Subnet{
					Name:              "workerSubnet",
					CidrBlock:         "10.0.1.0/24",
					SecurityGroupName: "workerSecurityGroup",
				},
				ControlPlaneSubnet: &models.V1Subnet{
					Name:              "controlPlaneSubnet",
					CidrBlock:         "10.0.2.0/24",
					SecurityGroupName: "controlPlaneSecurityGroup",
				},
			},
		},
	}

	// Call the flattenClusterConfigsAks function with the simulated Azure cloud configuration
	flattened := flattenClusterConfigsAks(azureCloudConfig)

	// Assertions
	assert.NotNil(t, flattened)
	assert.Len(t, flattened, 1)

	m := flattened[0].(map[string]interface{})
	assert.Equal(t, StringPtr("mySubscriptionID"), m["subscription_id"])
	assert.Equal(t, "myResourceGroup", m["resource_group"])
	assert.Equal(t, "eastus", m["region"])
	assert.Equal(t, "sshPublicKey", m["ssh_key"])
	assert.True(t, m["private_cluster"].(bool))
	assert.Equal(t, "myVnet", m["vnet_name"])
	assert.Equal(t, "myVnetResourceGroup", m["vnet_resource_group"])
	assert.Equal(t, "10.0.0.0/16", m["vnet_cidr_block"])
	assert.Equal(t, "workerSubnet", m["worker_subnet_name"])
	assert.Equal(t, "10.0.1.0/24", m["worker_cidr"])
	assert.Equal(t, "workerSecurityGroup", m["worker_subnet_security_group_name"])
	assert.Equal(t, "controlPlaneSubnet", m["control_plane_subnet_name"])
	assert.Equal(t, "10.0.2.0/24", m["control_plane_cidr"])
	assert.Equal(t, "controlPlaneSecurityGroup", m["control_plane_subnet_security_group_name"])
}

func TestFlattenMachinePoolConfigsAks(t *testing.T) {
	// Simulate Azure machine pool configurations
	machinePool1 := &models.V1AzureMachinePoolConfig{
		Name:             "pool1",
		Size:             3,
		MinSize:          1,
		MaxSize:          5,
		IsSystemNodePool: false,
		InstanceType:     "Standard_DS2_v2",
		OsDisk: &models.V1AzureOSDisk{
			DiskSizeGB: 100,
			ManagedDisk: &models.V1ManagedDisk{
				StorageAccountType: "Standard_LRS",
			},
		},
		UpdateStrategy: &models.V1UpdateStrategy{
			Type: "RollingUpdate",
		},
	}

	machinePool2 := &models.V1AzureMachinePoolConfig{
		Name:             "pool2",
		Size:             5,
		MinSize:          2,
		MaxSize:          8,
		IsSystemNodePool: true,
		InstanceType:     "Standard_DS3_v2",
		OsDisk: &models.V1AzureOSDisk{
			DiskSizeGB: 200,
			ManagedDisk: &models.V1ManagedDisk{
				StorageAccountType: "Premium_LRS",
			},
		},
		UpdateStrategy: &models.V1UpdateStrategy{
			Type: "Recreate",
		},
	}

	machinePools := []*models.V1AzureMachinePoolConfig{machinePool1, machinePool2}

	// Call the flattenMachinePoolConfigsAks function with the simulated Azure machine pool configurations
	flattened := flattenMachinePoolConfigsAks(machinePools)

	// Assertions
	assert.NotNil(t, flattened)
	// Machine pool 1
	m1 := flattened[0].(map[string]interface{})
	assert.Equal(t, "pool1", m1["name"])
	assert.Equal(t, 3, m1["count"])
	assert.Equal(t, 1, m1["min"])
	assert.Equal(t, 5, m1["max"])
	assert.Equal(t, "Standard_DS2_v2", m1["instance_type"])
	assert.Equal(t, 100, m1["disk_size_gb"])
	assert.False(t, m1["is_system_node_pool"].(bool))
	assert.Equal(t, "Standard_LRS", m1["storage_account_type"])
	assert.Equal(t, "RollingUpdate", m1["update_strategy"])

}

func TestToClusterTemplateAks(t *testing.T) {
	tests := []struct {
		name        string
		templateUID string
		expectNil   bool
	}{
		{
			name:        "with template UID",
			templateUID: "template-123",
			expectNil:   false,
		},
		{
			name:        "empty template UID",
			templateUID: "",
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceData := map[string]interface{}{
				"name":             "test-aks-cluster",
				"cloud_account_id": "test-account",
				"cloud_config": []interface{}{
					map[string]interface{}{
						"region":          "eastus",
						"resource_group":  "test-rg",
						"subscription_id": "test-sub",
						"ssh_key":         "test-key",
					},
				},
			}

			if tt.templateUID != "" {
				resourceData["cluster_template"] = tt.templateUID
			}

			d := resourceClusterAks().TestResourceData()
			for k, v := range resourceData {
				d.Set(k, v)
			}

			result := toClusterTemplate(d)

			if tt.expectNil {
				assert.Nil(t, result, "Expected nil but got %+v", result)
			} else {
				assert.NotNil(t, result, "Expected result but got nil")
				assert.Equal(t, tt.templateUID, result.UID, "Expected UID '%s', got '%s'", tt.templateUID, result.UID)
			}
		})
	}
}

func TestFlattenClusterTemplateAks(t *testing.T) {
	tests := []struct {
		name            string
		clusterTemplate *models.V1SpectroClusterTemplateRef
		expectedResult  string
	}{
		{
			name: "with template UID",
			clusterTemplate: &models.V1SpectroClusterTemplateRef{
				UID: "template-123",
			},
			expectedResult: "template-123",
		},
		{
			name:            "nil template",
			clusterTemplate: nil,
			expectedResult:  "",
		},
		{
			name: "empty template UID",
			clusterTemplate: &models.V1SpectroClusterTemplateRef{
				UID: "",
			},
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenClusterTemplate(tt.clusterTemplate)
			assert.Equal(t, tt.expectedResult, result, "Expected '%s', got '%s'", tt.expectedResult, result)
		})
	}
}
