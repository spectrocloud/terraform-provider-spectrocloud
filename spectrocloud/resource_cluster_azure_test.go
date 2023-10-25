package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToStaticPlacement(t *testing.T) {
	// Created a mock of models.V1SpectroAzureClusterEntity and cloudConfig
	c := &models.V1SpectroAzureClusterEntity{
		Metadata: nil,
		Spec: &models.V1SpectroAzureClusterEntitySpec{
			CloudAccountUID: nil,
			CloudConfig: &models.V1AzureClusterConfig{
				AadProfile:             nil,
				APIServerAccessProfile: nil,
				ContainerName:          "",
				ControlPlaneSubnet:     nil,
				EnablePrivateCluster:   false,
				InfraLBConfig:          nil,
				Location:               nil,
				ResourceGroup:          "",
				SSHKey:                 nil,
				StorageAccountName:     "",
				SubscriptionID:         nil,
				VnetCidrBlock:          "",
				VnetName:               "",
				VnetResourceGroup:      "",
				WorkerSubnet:           nil,
			},
			ClusterConfig:     nil,
			Machinepoolconfig: nil,
			Policies:          nil,
			Profiles:          nil,
		},
	}
	cloudConfig := map[string]interface{}{
		"static_placement": []interface{}{
			map[string]interface{}{
				"network_resource_group":     "rg",
				"virtual_network_name":       "vnet",
				"virtual_network_cidr_block": "10.0.0.0/16",
				"control_plane_subnet": []interface{}{
					map[string]interface{}{
						"cidr_block":          "10.0.0.0/24",
						"name":                "cp_subnet",
						"security_group_name": "cp_sg",
					},
				},
				"worker_node_subnet": []interface{}{
					map[string]interface{}{
						"cidr_block":          "10.0.1.0/24",
						"name":                "worker_subnet",
						"security_group_name": "worker_sg",
					},
				},
			},
		},
	}

	toStaticPlacement(c, cloudConfig)

	// Verify the values in c are set correctly
	expected := &models.V1SpectroAzureClusterEntity{
		Spec: &models.V1SpectroAzureClusterEntitySpec{
			CloudConfig: &models.V1AzureClusterConfig{
				VnetResourceGroup: "rg",
				VnetName:          "vnet",
				VnetCidrBlock:     "10.0.0.0/16",
				ControlPlaneSubnet: &models.V1Subnet{
					CidrBlock:         "10.0.0.0/24",
					Name:              "cp_subnet",
					SecurityGroupName: "cp_sg",
				},
				WorkerSubnet: &models.V1Subnet{
					CidrBlock:         "10.0.1.0/24",
					Name:              "worker_subnet",
					SecurityGroupName: "worker_sg",
				},
			},
		},
	}

	assert.Equal(t, expected, c)
}

func TestValidateMasterPoolCount(t *testing.T) {
	// Test case 1: Even master pool size
	masterConfig1 := &models.V1AzureMachinePoolConfigEntity{
		PoolConfig: &models.V1MachinePoolConfigEntity{
			IsControlPlane: true,
			Size:           int32Ptr(4),
			Name:           stringPtr("master1"),
		},
	}

	// Test case 2: Non-master pool
	workerConfig := &models.V1AzureMachinePoolConfigEntity{
		PoolConfig: &models.V1MachinePoolConfigEntity{
			IsControlPlane: false,
			Size:           int32Ptr(6),
			Name:           stringPtr("worker1"),
		},
	}

	// Test case 3: Non-control plane pool with odd size
	nonMasterConfig := &models.V1AzureMachinePoolConfigEntity{
		PoolConfig: &models.V1MachinePoolConfigEntity{
			IsControlPlane: false,
			Size:           int32Ptr(7),
			Name:           stringPtr("worker2"),
		},
	}

	machinePool := []*models.V1AzureMachinePoolConfigEntity{
		masterConfig1,
		workerConfig,
		nonMasterConfig,
	}

	// Run the function and capture the diagnostics
	diagnostics := validateMasterPoolCount(machinePool)

	// Test case 1 should return an error, so diagnostics should not be nil
	assert.NotNil(t, diagnostics, "Test case 1 failed: Expected diagnostics to be non-nil")

	// Test case 2 and 3 are not control plane pools, so they should pass, so diagnostics should be nil
}

func int32Ptr(i int32) *int32 {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

func TestToMachinePoolAzure(t *testing.T) {
	// Test case 1: Provide a valid machine pool
	azsList := []string{"us-east-1a", "us-east-1b"}
	azsSet := schema.NewSet(schema.HashString, nil)
	for _, az := range azsList {
		azsSet.Add(az)
	}
	machinePool := map[string]interface{}{
		"name":                    "pool1",
		"count":                   3,
		"control_plane":           true,
		"control_plane_as_worker": false,
		"instance_type":           "Standard_DS2_v2",
		"os_type":                 "Linux",
		"azs":                     azsSet,
		"disk": []interface{}{
			map[string]interface{}{
				"size_gb": 50,
				"type":    "Standard_LRS",
			},
		},
		"is_system_node_pool":  false,
		"node_repave_interval": 0,
	}

	result, err := toMachinePoolAzure(machinePool)

	// Check for non-nil result
	assert.NotNil(t, result)

	// Check for nil error
	assert.NoError(t, err)
	assert.Equal(t, machinePool["name"], *result.PoolConfig.Name)
	//assert.Equal(t, machinePool["count"].(int32), &result.PoolConfig.Size)
	assert.Equal(t, machinePool["control_plane"], result.PoolConfig.IsControlPlane)
	assert.Equal(t, machinePool["control_plane_as_worker"], result.PoolConfig.UseControlPlaneAsWorker)
	assert.Equal(t, machinePool["instance_type"], result.CloudConfig.InstanceType)

}
