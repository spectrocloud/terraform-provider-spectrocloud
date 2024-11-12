package spectrocloud

import (
	"sort"
	"testing"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func prepareAzureTestData() *schema.ResourceData {
	d := resourceClusterAzure().TestResourceData()
	cloudConfig := []interface{}{
		map[string]interface{}{
			"region":               "us-west-2",
			"ssh_key":              "dummy_ssh_key",
			"subscription_id":      "dummy_subscription_id",
			"resource_group":       "dummy_resource_group",
			"storage_account_name": "dummy_storage_account_name",
			"container_name":       "dummy_container_name",
		},
	}
	azsList := []string{"us-east-1a", "us-east-1b"}
	azsSet := schema.NewSet(schema.HashString, nil)
	for _, az := range azsList {
		azsSet.Add(az)
	}
	var mpools []interface{}
	mpool := map[string]interface{}{
		"name":                    "pool1",
		"count":                   3,
		"control_plane":           false,
		"control_plane_as_worker": false,
		"instance_type":           "Standard_DS2_v2",
		"os_type":                 "Windows",
		"azs":                     azsSet,
		"disk": []interface{}{
			map[string]interface{}{
				"size_gb": 50,
				"type":    "Standard_LRS",
			},
		},
		"is_system_node_pool":  false,
		"node_repave_interval": 10,
	}
	mpools = append(mpools, mpool)
	d.Set("cloud_config", cloudConfig)
	d.Set("context", "Tenant")
	d.Set("name", "dummy_name")
	d.Set("cloud_account_id", "dummy_account_id")
	d.Set("machine_pool", mpools)
	return d
}

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

func TestValidateCPPoolCount(t *testing.T) {
	// Test case 1: Even control-plane pool size
	cpConfig1 := &models.V1AzureMachinePoolConfigEntity{
		PoolConfig: &models.V1MachinePoolConfigEntity{
			IsControlPlane: true,
			Size:           ptr.To(int32(4)),
			Name:           ptr.To("cp1"),
		},
	}

	// Test case 2: Non-cp pool
	workerConfig := &models.V1AzureMachinePoolConfigEntity{
		PoolConfig: &models.V1MachinePoolConfigEntity{
			IsControlPlane: false,
			Size:           ptr.To(int32(6)),
			Name:           ptr.To("worker1"),
		},
	}

	// Test case 3: Non-control plane pool with odd size
	nonCPConfig := &models.V1AzureMachinePoolConfigEntity{
		PoolConfig: &models.V1MachinePoolConfigEntity{
			IsControlPlane: false,
			Size:           ptr.To(int32(7)),
			Name:           ptr.To("worker2"),
		},
	}

	machinePool := []*models.V1AzureMachinePoolConfigEntity{
		cpConfig1,
		workerConfig,
		nonCPConfig,
	}

	// Run the function and capture the diagnostics
	diagnostics := validateCPPoolCount(machinePool)

	// Test case 1 should return an error, so diagnostics should not be nil
	assert.NotNil(t, diagnostics, "Test case 1 failed: Expected diagnostics to be non-nil")

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
		"os_type":                 "Windows",
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
	assert.Equal(t, machinePool["control_plane"], result.PoolConfig.IsControlPlane)
	assert.Equal(t, machinePool["control_plane_as_worker"], result.PoolConfig.UseControlPlaneAsWorker)
	assert.Equal(t, machinePool["instance_type"], result.CloudConfig.InstanceType)
	assert.Equal(t, machinePool["os_type"], string(result.CloudConfig.OsDisk.OsType))
	assert.Equal(t, machinePool["is_system_node_pool"], result.CloudConfig.IsSystemNodePool)
	assert.Equal(t, machinePool["node_repave_interval"], int(result.PoolConfig.NodeRepaveInterval))
	assert.Equal(t, machinePool["count"], int(*result.PoolConfig.Size))

}

func TestFlattenMachinePoolConfigsAzure(t *testing.T) {
	// Sample V1AzureMachinePoolConfig data
	azsList := []string{"us-east-1a", "us-east-1b"}
	azsSet := schema.NewSet(schema.HashString, nil)
	for _, az := range azsList {
		azsSet.Add(az)
	}
	machinePools := []*models.V1AzureMachinePoolConfig{
		{
			AdditionalLabels:      nil,
			AdditionalTags:        nil,
			Azs:                   azsList,
			InstanceConfig:        nil,
			InstanceType:          "Standard_DS2_v2",
			IsControlPlane:        ptr.To(false),
			IsSystemNodePool:      false,
			Labels:                nil,
			MachinePoolProperties: nil,
			MaxSize:               2,
			MinSize:               5,
			Name:                  "worker_pool",
			NodeRepaveInterval:    2,
			OsDisk: &models.V1AzureOSDisk{
				DiskSizeGB:  50,
				ManagedDisk: &models.V1ManagedDisk{StorageAccountType: "test"},
				OsType:      "Linux",
			},
			OsType:                  "Linux",
			Size:                    5,
			SpotVMOptions:           nil,
			Taints:                  nil,
			UpdateStrategy:          nil,
			UseControlPlaneAsWorker: false,
		},
	}

	result := flattenMachinePoolConfigsAzure(machinePools)
	actual := result[0].(map[string]interface{})
	// Assert the flattened result
	assert.Len(t, result, len(machinePools), "Expected length of result to match number of machine pools")
	assert.Equal(t, machinePools[0].InstanceType, actual["instance_type"])
	assert.Equal(t, machinePools[0].OsType, actual["os_type"])
	assert.Equal(t, machinePools[0].Name, actual["name"])
	assert.Equal(t, machinePools[0].NodeRepaveInterval, actual["node_repave_interval"])
	assert.Equal(t, machinePools[0].Size, actual["count"])
}

func TestFlattenClusterConfigsAzure(t *testing.T) {
	t.Run("NilConfig", func(t *testing.T) {
		result := flattenClusterConfigsAzure(nil)
		assert.Equal(t, []interface{}{}, result)
	})

	t.Run("EmptyConfig", func(t *testing.T) {
		config := &models.V1AzureCloudConfig{}
		result := flattenClusterConfigsAzure(config)
		assert.Equal(t, []interface{}{}, result)
	})

	t.Run("PartialConfig", func(t *testing.T) {
		config := &models.V1AzureCloudConfig{
			Spec: &models.V1AzureCloudConfigSpec{
				ClusterConfig: &models.V1AzureClusterConfig{
					SubscriptionID:     ptr.To("test-subscription-id"),
					ResourceGroup:      "test-resource-group",
					Location:           ptr.To("test-location"),
					SSHKey:             ptr.To("test-ssh-key"),
					StorageAccountName: "test-storage-account",
					ContainerName:      "test-container",
					VnetResourceGroup:  "test-network-resource-group",
					VnetName:           "test-vnet",
					VnetCidrBlock:      "10.0.0.0/16",
					ControlPlaneSubnet: &models.V1Subnet{Name: "cp-subnet", CidrBlock: "10.0.1.0/24", SecurityGroupName: "cp-sg"},
					WorkerSubnet:       &models.V1Subnet{Name: "worker-subnet", CidrBlock: "10.0.2.0/24", SecurityGroupName: "worker-sg"},
				},
			},
		}

		expected := []interface{}{
			map[string]interface{}{
				"subscription_id":            ptr.To("test-subscription-id"),
				"resource_group":             "test-resource-group",
				"region":                     ptr.To("test-location"),
				"ssh_key":                    ptr.To("test-ssh-key"),
				"storage_account_name":       "test-storage-account",
				"container_name":             "test-container",
				"network_resource_group":     "test-network-resource-group",
				"virtual_network_name":       "test-vnet",
				"virtual_network_cidr_block": "10.0.0.0/16",
				"control_plane_subnet": []interface{}{
					map[string]interface{}{
						"name":                "cp-subnet",
						"cidr_block":          "10.0.1.0/24",
						"security_group_name": "cp-sg",
					},
				},
				"worker_node_subnet": []interface{}{
					map[string]interface{}{
						"name":                "worker-subnet",
						"cidr_block":          "10.0.2.0/24",
						"security_group_name": "worker-sg",
					},
				},
			},
		}

		result := flattenClusterConfigsAzure(config)
		sortSliceOfMaps(expected)
		sortSliceOfMaps(result)
		assert.Equal(t, expected, result)
	})

	t.Run("MissingFields", func(t *testing.T) {
		config := &models.V1AzureCloudConfig{
			Spec: &models.V1AzureCloudConfigSpec{
				ClusterConfig: &models.V1AzureClusterConfig{
					ResourceGroup: "test-resource-group",
					Location:      ptr.To("test-location"),
				},
			},
		}

		expected := []interface{}{
			map[string]interface{}{
				"resource_group": "test-resource-group",
				"region":         ptr.To("test-location"),
			},
		}

		result := flattenClusterConfigsAzure(config)
		assert.Equal(t, expected, result)
	})
}

func sortSliceOfMaps(slice []interface{}) {
	sort.SliceStable(slice, func(i, j int) bool {
		mapI := slice[i].(map[string]interface{})
		mapJ := slice[j].(map[string]interface{})
		return mapI["name"].(string) < mapJ["name"].(string)
	})
}
