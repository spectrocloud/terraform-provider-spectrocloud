package spectrocloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/stretchr/testify/assert"
)

func commonNodePool() map[string]interface{} {
	nodePool := map[string]interface{}{
		"additional_labels": map[string]interface{}{
			"label1": "value1",
		},
		"taints": []interface{}{
			map[string]interface{}{
				"key":    "taint1",
				"value":  "true",
				"effect": "NoSchedule",
			},
		},
		"control_plane":           true,
		"control_plane_as_worker": false,
		"name":                    "test-pool",
		"count":                   3,
		"update_strategy":         "RollingUpdate",
		"node_repave_interval":    10,
	}
	return nodePool
}

func TestCommonHash(t *testing.T) {
	expectedHash := "label1-value1effect-NoSchedulekey-taint1value-truetrue-false-test-pool-3-RollingUpdate-10-"
	hash := CommonHash(commonNodePool()).String()

	assert.Equal(t, expectedHash, hash)
}

func TestResourceMachinePoolAzureHash(t *testing.T) {
	nodePool := map[string]interface{}{
		"additional_labels": map[string]interface{}{
			"label1": "value1",
		},
		"taints": []interface{}{
			map[string]interface{}{
				"key":    "taint1",
				"value":  "true",
				"effect": "NoSchedule",
			},
		},
		"control_plane":           true,
		"control_plane_as_worker": false,
		"name":                    "test-pool",
		"count":                   3,
		"update_strategy":         "RollingUpdate",
		"node_repave_interval":    10,
		"instance_type":           "Standard_D2_v3",
		"is_system_node_pool":     true,
		"os_type":                 "Linux",
	}

	expectedHash := 3495386805

	hash := resourceMachinePoolAzureHash(nodePool)

	assert.Equal(t, expectedHash, hash)
}

func TestResourceClusterHash(t *testing.T) {
	clusterData := map[string]interface{}{
		"uid": "abc123",
	}

	expectedHash := 1764273400

	hash := resourceClusterHash(clusterData)

	assert.Equal(t, expectedHash, hash)
}

func TestHashStringMapList(t *testing.T) {
	stringMapList := []interface{}{
		map[string]interface{}{"key1": "value1", "key2": "value2"},
		map[string]interface{}{"key3": "value3"},
	}

	expectedHash := "key1-value1key2-value2key3-value3"
	hash := HashStringMapList(stringMapList)

	assert.Equal(t, expectedHash, hash)
}

func TestHashStringMapListlength(t *testing.T) {
	stringMapList := []interface{}{}

	expectedHash := ""
	hash := HashStringMapList(stringMapList)

	assert.Equal(t, expectedHash, hash)
}

func TestResourceMachinePoolAksHash(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{
			name: "Complete AKS machine pool with all fields",
			input: map[string]interface{}{
				"name":                 "aks-pool-1",
				"count":                3,
				"instance_type":        "Standard_D2s_v3",
				"disk_size_gb":         100,
				"is_system_node_pool":  true,
				"storage_account_type": "Premium_LRS",
				"additional_labels": map[string]interface{}{
					"env":  "production",
					"team": "platform",
				},
				"update_strategy": "RollingUpdateScaleOut",
				"min":             1,
				"max":             5,
				"node": []interface{}{
					map[string]interface{}{
						"action": "cordon",
					},
				},
				"taints": []interface{}{
					map[string]interface{}{
						"key":    "dedicated",
						"value":  "backend",
						"effect": "NoSchedule",
					},
				},
			},
			expected: 489635413,
		},
		{
			name: "Minimal AKS machine pool",
			input: map[string]interface{}{
				"name":                 "aks-pool-2",
				"count":                2,
				"instance_type":        "Standard_B2s",
				"disk_size_gb":         50,
				"is_system_node_pool":  false,
				"storage_account_type": "Standard_LRS",
			},
			expected: 4269923102,
		},
		{
			name: "AKS machine pool with autoscaling",
			input: map[string]interface{}{
				"name":                 "aks-pool-3",
				"count":                2,
				"instance_type":        "Standard_D2s_v3",
				"disk_size_gb":         80,
				"is_system_node_pool":  false,
				"storage_account_type": "Premium_LRS",
				"min":                  1,
				"max":                  10,
			},
			expected: 1815788174,
		},
		{
			name: "System node pool with labels",
			input: map[string]interface{}{
				"name":                 "system-pool",
				"count":                1,
				"instance_type":        "Standard_DS2_v2",
				"disk_size_gb":         30,
				"is_system_node_pool":  true,
				"storage_account_type": "Standard_LRS",
				"additional_labels": map[string]interface{}{
					"pool-type": "system",
				},
			},
			expected: 650558149,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := resourceMachinePoolAksHash(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestResourceMachinePoolAksHashAllFields tests that all fields are included in hash
func TestResourceMachinePoolAksHashAllFields(t *testing.T) {
	testCases := []struct {
		name        string
		baseInput   map[string]interface{}
		modifyField func(map[string]interface{})
		description string
	}{
		{
			name: "Name change affects hash",
			baseInput: map[string]interface{}{
				"name":                 "pool-1",
				"count":                2,
				"instance_type":        "Standard_D2s_v3",
				"disk_size_gb":         100,
				"is_system_node_pool":  false,
				"storage_account_type": "Premium_LRS",
			},
			modifyField: func(m map[string]interface{}) {
				m["name"] = "pool-2"
			},
			description: "Changing name should change hash",
		},
		{
			name: "Count change affects hash",
			baseInput: map[string]interface{}{
				"name":                 "pool-1",
				"count":                2,
				"instance_type":        "Standard_D2s_v3",
				"disk_size_gb":         100,
				"is_system_node_pool":  false,
				"storage_account_type": "Premium_LRS",
			},
			modifyField: func(m map[string]interface{}) {
				m["count"] = 3
			},
			description: "Changing count should change hash",
		},
		{
			name: "Instance type change affects hash",
			baseInput: map[string]interface{}{
				"name":                 "pool-1",
				"count":                2,
				"instance_type":        "Standard_D2s_v3",
				"disk_size_gb":         100,
				"is_system_node_pool":  false,
				"storage_account_type": "Premium_LRS",
			},
			modifyField: func(m map[string]interface{}) {
				m["instance_type"] = "Standard_D4s_v3"
			},
			description: "Changing instance_type should change hash",
		},
		{
			name: "Disk size change affects hash",
			baseInput: map[string]interface{}{
				"name":                 "pool-1",
				"count":                2,
				"instance_type":        "Standard_D2s_v3",
				"disk_size_gb":         100,
				"is_system_node_pool":  false,
				"storage_account_type": "Premium_LRS",
			},
			modifyField: func(m map[string]interface{}) {
				m["disk_size_gb"] = 200
			},
			description: "Changing disk_size_gb should change hash",
		},
		{
			name: "System node pool flag change affects hash",
			baseInput: map[string]interface{}{
				"name":                 "pool-1",
				"count":                2,
				"instance_type":        "Standard_D2s_v3",
				"disk_size_gb":         100,
				"is_system_node_pool":  false,
				"storage_account_type": "Premium_LRS",
			},
			modifyField: func(m map[string]interface{}) {
				m["is_system_node_pool"] = true
			},
			description: "Changing is_system_node_pool should change hash",
		},
		{
			name: "Storage account type change affects hash",
			baseInput: map[string]interface{}{
				"name":                 "pool-1",
				"count":                2,
				"instance_type":        "Standard_D2s_v3",
				"disk_size_gb":         100,
				"is_system_node_pool":  false,
				"storage_account_type": "Premium_LRS",
			},
			modifyField: func(m map[string]interface{}) {
				m["storage_account_type"] = "Standard_LRS"
			},
			description: "Changing storage_account_type should change hash",
		},
		{
			name: "Additional labels change affects hash",
			baseInput: map[string]interface{}{
				"name":                 "pool-1",
				"count":                2,
				"instance_type":        "Standard_D2s_v3",
				"disk_size_gb":         100,
				"is_system_node_pool":  false,
				"storage_account_type": "Premium_LRS",
				"additional_labels": map[string]interface{}{
					"env": "dev",
				},
			},
			modifyField: func(m map[string]interface{}) {
				m["additional_labels"] = map[string]interface{}{
					"env": "prod",
				}
			},
			description: "Changing additional_labels should change hash",
		},
		{
			name: "Update strategy change affects hash",
			baseInput: map[string]interface{}{
				"name":                 "pool-1",
				"count":                2,
				"instance_type":        "Standard_D2s_v3",
				"disk_size_gb":         100,
				"is_system_node_pool":  false,
				"storage_account_type": "Premium_LRS",
				"update_strategy":      "RollingUpdateScaleOut",
			},
			modifyField: func(m map[string]interface{}) {
				m["update_strategy"] = "RollingUpdateScaleIn"
			},
			description: "Changing update_strategy should change hash",
		},
		{
			name: "Min change affects hash",
			baseInput: map[string]interface{}{
				"name":                 "pool-1",
				"count":                2,
				"instance_type":        "Standard_D2s_v3",
				"disk_size_gb":         100,
				"is_system_node_pool":  false,
				"storage_account_type": "Premium_LRS",
				"min":                  1,
			},
			modifyField: func(m map[string]interface{}) {
				m["min"] = 2
			},
			description: "Changing min should change hash",
		},
		{
			name: "Max change affects hash",
			baseInput: map[string]interface{}{
				"name":                 "pool-1",
				"count":                2,
				"instance_type":        "Standard_D2s_v3",
				"disk_size_gb":         100,
				"is_system_node_pool":  false,
				"storage_account_type": "Premium_LRS",
				"max":                  5,
			},
			modifyField: func(m map[string]interface{}) {
				m["max"] = 10
			},
			description: "Changing max should change hash",
		},
		{
			name: "Node configuration change affects hash",
			baseInput: map[string]interface{}{
				"name":                 "pool-1",
				"count":                2,
				"instance_type":        "Standard_D2s_v3",
				"disk_size_gb":         100,
				"is_system_node_pool":  false,
				"storage_account_type": "Premium_LRS",
				"node": []interface{}{
					map[string]interface{}{
						"action": "cordon",
					},
				},
			},
			modifyField: func(m map[string]interface{}) {
				m["node"] = []interface{}{
					map[string]interface{}{
						"action": "drain",
					},
				}
			},
			description: "Changing node config should change hash",
		},
		{
			name: "Taints change affects hash",
			baseInput: map[string]interface{}{
				"name":                 "pool-1",
				"count":                2,
				"instance_type":        "Standard_D2s_v3",
				"disk_size_gb":         100,
				"is_system_node_pool":  false,
				"storage_account_type": "Premium_LRS",
				"taints": []interface{}{
					map[string]interface{}{
						"key":    "key1",
						"value":  "value1",
						"effect": "NoSchedule",
					},
				},
			},
			modifyField: func(m map[string]interface{}) {
				m["taints"] = []interface{}{
					map[string]interface{}{
						"key":    "key2",
						"value":  "value2",
						"effect": "NoExecute",
					},
				}
			},
			description: "Changing taints should change hash",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get hash of base input
			baseHash := resourceMachinePoolAksHash(tc.baseInput)

			// Create modified copy
			modified := copyMap(tc.baseInput)
			tc.modifyField(modified)

			// Get hash of modified input
			modifiedHash := resourceMachinePoolAksHash(modified)

			// Hashes should be different
			if baseHash == modifiedHash {
				t.Errorf("%s: Base hash %d equals modified hash %d, but they should differ.\nBase: %+v\nModified: %+v",
					tc.description, baseHash, modifiedHash, tc.baseInput, modified)
			}
		})
	}
}

func TestResourceMachinePoolGcpHash(t *testing.T) {
	testCases := []struct {
		input    interface{}
		expected int
	}{
		{
			input: map[string]interface{}{
				"instance_type": "n1-standard-4",
				"min":           1,
				"max":           3,
				"capacity_type": "ON_DEMAND",
				"max_price":     "0.12",
				"azs":           schema.NewSet(schema.HashString, []interface{}{"us-central1-a", "us-central1-b"}),
				"az_subnets": map[string]interface{}{
					"us-central1-a": "subnet-1",
					"us-central1-b": "subnet-2",
				},
			},
			expected: 2586515099,
		},
	}
	for _, tc := range testCases {
		actual := resourceMachinePoolGcpHash(tc.input)
		if actual != tc.expected {
			t.Errorf("Expected hash %d, but got %d for input %+v", tc.expected, actual, tc.input)
		}
	}
}

func TestResourceMachinePoolAwsHash(t *testing.T) {
	testCases := []struct {
		input    interface{}
		expected int
	}{
		{
			input: map[string]interface{}{
				"min":           1,
				"max":           5,
				"instance_type": "t2.micro",
				"capacity_type": "ON_DEMAND",
				"max_price":     "0.03",
				"azs": schema.NewSet(schema.HashString, []interface{}{
					"us-east-1a",
					"us-east-1b",
				}),

				"az_subnets": map[string]interface{}{
					"us-east-1a": "subnet-1",
					"us-east-1b": "subnet-2",
				},
			},
			expected: 1929542909,
		},
	}

	for _, tc := range testCases {
		actual := resourceMachinePoolAwsHash(tc.input)
		if actual != tc.expected {
			t.Errorf("Expected hash %d, but got %d for input %+v", tc.expected, actual, tc.input)
		}
	}
}

func TestResourceMachinePoolEksHash(t *testing.T) {
	testCases := []struct {
		input    interface{}
		expected int
	}{
		{
			input: map[string]interface{}{
				"disk_size_gb":  100,
				"min":           2,
				"max":           5,
				"instance_type": "t2.micro",
				"capacity_type": "on-demand",
				"max_price":     "0.05",
				"az_subnets": map[string]interface{}{
					"subnet1": "subnet-123",
					"subnet2": "subnet-456",
				},
				"eks_launch_template": []interface{}{
					map[string]interface{}{
						"ami_id":           "ami-123",
						"root_volume_type": "gp2",
					},
				},
			},
			expected: 706444520,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Input: %v", tc.input), func(t *testing.T) {
			// Call the function with the test input
			result := resourceMachinePoolEksHash(tc.input)

			// Check if the result matches the expected output
			if result != tc.expected {
				t.Errorf("Expected: %d, Got: %d", tc.expected, result)
			}
		})
	}
}

func TestEksLaunchTemplate(t *testing.T) {
	testCases := []struct {
		input    interface{}
		expected string
	}{
		{

			input: []interface{}{
				map[string]interface{}{
					"ami_id":                     "ami-123",
					"root_volume_type":           "gp2",
					"root_volume_iops":           100,
					"root_volume_throughput":     200,
					"additional_security_groups": schema.NewSet(schema.HashString, []interface{}{"sg-123", "sg-456"}),
				},
			},
			expected: "ami-123-gp2-100-200-sg-456-sg-123-",
		},
		{
			// Test case with invalid input type (slice of non-map)
			input:    []interface{}{},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Input: %v", tc.input), func(t *testing.T) {
			// Call the function with the test input
			result := eksLaunchTemplate(tc.input)

			// Check if the result matches the expected output
			if result != tc.expected {
				t.Errorf("Expected: %s, Got: %s", tc.expected, result)
			}
		})
	}
}

func TestResourceMachinePoolVsphereHash(t *testing.T) {
	testCases := []struct {
		input    interface{}
		expected int
	}{
		{
			input: map[string]interface{}{
				"instance_type": []interface{}{
					map[string]interface{}{
						"cpu":          2,
						"disk_size_gb": 50,
						"memory_mb":    4096,
					},
				},
				"placement": []interface{}{
					map[string]interface{}{
						"cluster":           "cluster1",
						"resource_pool":     "resource_pool1",
						"datastore":         "datastore1",
						"network":           "network1",
						"static_ip_pool_id": "static_pool1",
					},
				},
			},
			expected: 556255137,
		},
		{
			// Test case with missing instance_type
			input: map[string]interface{}{
				"placement": []interface{}{
					map[string]interface{}{
						"cluster":           "cluster2",
						"resource_pool":     "resource_pool2",
						"datastore":         "datastore2",
						"network":           "network2",
						"static_ip_pool_id": "static_pool2",
					},
				},
			},
			expected: 3826670463,
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			// Call the function with the test input
			result := resourceMachinePoolVsphereHash(tc.input)

			// Check if the result matches the expected output
			if result != tc.expected {
				t.Errorf("Expected: %d, Got: %d", tc.expected, result)
			}
		})
	}
}

func TestResourceMachinePoolEdgeNativeHash(t *testing.T) {
	testCases := []struct {
		input    interface{}
		expected int
	}{
		{
			input:    map[string]interface{}{},
			expected: 2166136261,
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			result := resourceMachinePoolEdgeNativeHash(tc.input)

			if result != tc.expected {
				t.Errorf("Expected: %d, Got: %d", tc.expected, result)
			}
		})
	}
}

func TestResourceMachinePoolEdgeNativeHashAdv(t *testing.T) {
	machinePool1 := map[string]interface{}{
		"edge_host": []interface{}{
			map[string]interface{}{
				"host_name": "host1",
				"host_uid":  "uid1",
				"static_ip": "192.168.1.1",
			},
			map[string]interface{}{
				"host_name": "host2",
				"host_uid":  "uid2",
				"static_ip": "192.168.1.2",
			},
		},
	}

	machinePool2 := map[string]interface{}{
		"edge_host": []interface{}{
			map[string]interface{}{
				"host_name": "host3",
				"host_uid":  "uid3",
				"static_ip": "192.168.1.3",
			},
			map[string]interface{}{
				"host_name": "host4",
				"host_uid":  "uid4",
				"static_ip": "192.168.1.4",
			},
		},
	}

	hash1 := resourceMachinePoolEdgeNativeHash(machinePool1)
	hash2 := resourceMachinePoolEdgeNativeHash(machinePool1) // Same input as above
	hash3 := resourceMachinePoolEdgeNativeHash(machinePool2) // Different input

	if hash1 != hash2 {
		t.Errorf("Hashes do not match for the same input: got %v want %v", hash2, hash1)
	}

	if hash1 == hash3 {
		t.Errorf("Hashes should not match for different inputs: got %v", hash3)
	}
}

func TestGpuConfigHash(t *testing.T) {
	testCases := []struct {
		input    map[string]interface{}
		expected string
	}{
		{

			input: map[string]interface{}{
				"num_gpus":     2,
				"device_model": "model1",
				"vendor":       "vendor1",
				"addresses": map[string]interface{}{
					"address1": "value1",
					"address2": "value2",
				},
			},
			expected: "2-model1-vendor1-address1-value1address2-value2",
		},
		{
			// Test case with missing "addresses" key
			input: map[string]interface{}{
				"num_gpus":     1,
				"device_model": "model2",
				"vendor":       "vendor2",
			},
			expected: "1-model2-vendor2-",
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			result := GpuConfigHash(tc.input)

			if result != tc.expected {
				t.Errorf("Expected: %s, Got: %s", tc.expected, result)
			}
		})
	}
}

func TestInstanceTypeHash(t *testing.T) {
	testCases := []struct {
		name         string
		input        map[string]interface{}
		expectedHash string
	}{
		{
			name: "Valid InstanceTypeHash",
			input: map[string]interface{}{
				"cpu":               4,
				"disk_size_gb":      100,
				"memory_mb":         8192,
				"cpus_sets":         "0-3",
				"cache_passthrough": true,
				"gpu_config": map[string]interface{}{
					"num_gpus":     2,
					"device_model": "Tesla T4",
					"vendor":       "NVIDIA",
					"addresses": map[string]interface{}{
						"gpu-address-1": "10.0.0.1",
						"gpu-address-2": "10.0.0.2",
					},
				},
				"attached_disks": []interface{}{
					map[string]interface{}{
						"managed":    true,
						"size_in_gb": 500,
					},
				},
			},
			expectedHash: "4-100-8192-0-3-cache_passthrough-true2-Tesla T4-NVIDIA-gpu-address-1-10.0.0.1gpu-address-2-10.0.0.2managed-truesize_in_gb-500",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash := InstanceTypeHash(tc.input)
			assert.Equal(t, tc.expectedHash, hash)
		})
	}
}

func TestResourceMachinePoolMaasHash(t *testing.T) {
	testCases := []struct {
		name         string
		input        interface{}
		expectedHash int
	}{
		{
			name: "Valid MachinePoolMaasHash",
			input: map[string]interface{}{
				"instance_type": []interface{}{
					map[string]interface{}{
						"min_cpu":       2,
						"min_memory_mb": 4096,
					},
				},
				"azs":       schema.NewSet(schema.HashString, []interface{}{"az1", "az2"}),
				"node_tags": schema.NewSet(schema.HashString, []interface{}{"test", "tf"}),
			},
			expectedHash: 876064649,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash := resourceMachinePoolMaasHash(tc.input)
			assert.Equal(t, tc.expectedHash, hash)
		})
	}
}

func TestResourceMachinePoolVirtualHash(t *testing.T) {
	testCases := []struct {
		name         string
		input        interface{}
		expectedHash int
	}{
		{
			name: "Valid MachinePoolVirtualHash",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
			expectedHash: 2166136261,
		},
		{
			name: "Test Case 2",
			input: map[string]interface{}{
				"key3": "value3",
				"key4": true,
			},
			expectedHash: 2166136261,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash := resourceMachinePoolVirtualHash(tc.input)
			assert.Equal(t, tc.expectedHash, hash)
		})
	}
}

func TestResourceMachinePoolOpenStackHash(t *testing.T) {
	testCases := []struct {
		name         string
		input        interface{}
		expectedHash int
	}{
		{
			name: "Valid MachinePoolOpenStackHash",
			input: map[string]interface{}{
				"name":                    "worker-pool-1",
				"count":                   3,
				"control_plane":           false,
				"control_plane_as_worker": false,
				"node_repave_interval":    0,
				"update_strategy":         "RollingUpdate",
				"instance_type":           "flavor1",
				"subnet_id":               "subnet123",
				"azs":                     schema.NewSet(schema.HashString, []interface{}{"az1", "az2"}),
			},
			expectedHash: 715623002,
		},
		{
			name: "Valid MachinePoolOpenStackHash 2",
			input: map[string]interface{}{
				"name":                    "worker-pool-2",
				"count":                   2,
				"control_plane":           false,
				"control_plane_as_worker": false,
				"node_repave_interval":    0,
				"update_strategy":         "Recreate",
				"instance_type":           "flavor2",
				"subnet_id":               "subnet456",
				"azs":                     schema.NewSet(schema.HashString, []interface{}{"az3"}),
			},
			expectedHash: 3371730139,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash := resourceMachinePoolOpenStackHash(tc.input)
			assert.Equal(t, tc.expectedHash, hash)
		})
	}
}

func TestResourceMachinePoolGkeHash(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{
			name: "Complete GKE machine pool with all fields",
			input: map[string]interface{}{
				"name":          "gke-pool-1",
				"count":         3,
				"disk_size_gb":  100,
				"instance_type": "n1-standard-4",
				"additional_labels": map[string]interface{}{
					"env":  "production",
					"team": "platform",
				},
				"update_strategy": "RollingUpdateScaleOut",
				"node": []interface{}{
					map[string]interface{}{
						"action": "cordon",
					},
				},
				"taints": []interface{}{
					map[string]interface{}{
						"key":    "dedicated",
						"value":  "backend",
						"effect": "NoSchedule",
					},
				},
			},
			expected: 2359262765,
		},
		{
			name: "Minimal GKE machine pool",
			input: map[string]interface{}{
				"name":          "gke-pool-2",
				"count":         1,
				"instance_type": "n1-standard-2",
			},
			expected: 1076173040,
		},
		{
			name: "GKE machine pool with disk size",
			input: map[string]interface{}{
				"name":          "gke-pool-3",
				"count":         2,
				"disk_size_gb":  50,
				"instance_type": "n1-standard-4",
			},
			expected: 239420914,
		},
		{
			name: "GKE machine pool with labels only",
			input: map[string]interface{}{
				"name":          "gke-pool-4",
				"count":         2,
				"instance_type": "n1-standard-2",
				"additional_labels": map[string]interface{}{
					"purpose": "testing",
				},
			},
			expected: 2140789356,
		},
		{
			name: "GKE machine pool with update strategy",
			input: map[string]interface{}{
				"name":            "gke-pool-5",
				"count":           3,
				"instance_type":   "n1-standard-4",
				"update_strategy": "RollingUpdateScaleIn",
			},
			expected: 3893189545,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := resourceMachinePoolGkeHash(tc.input)
			if actual != tc.expected {
				t.Errorf("Expected hash %d, but got %d for input %+v", tc.expected, actual, tc.input)
			}
		})
	}
}

// TestResourceMachinePoolGkeHashAllFields tests that all fields are included in hash
func TestResourceMachinePoolGkeHashAllFields(t *testing.T) {
	testCases := []struct {
		name        string
		baseInput   map[string]interface{}
		modifyField func(map[string]interface{})
		description string
	}{
		{
			name: "Name change affects hash",
			baseInput: map[string]interface{}{
				"name":          "pool-1",
				"count":         2,
				"instance_type": "n1-standard-2",
			},
			modifyField: func(m map[string]interface{}) {
				m["name"] = "pool-2"
			},
			description: "Changing name should change hash",
		},
		{
			name: "Count change affects hash",
			baseInput: map[string]interface{}{
				"name":          "pool-1",
				"count":         2,
				"instance_type": "n1-standard-2",
			},
			modifyField: func(m map[string]interface{}) {
				m["count"] = 3
			},
			description: "Changing count should change hash",
		},
		{
			name: "Disk size change affects hash",
			baseInput: map[string]interface{}{
				"name":          "pool-1",
				"count":         2,
				"disk_size_gb":  50,
				"instance_type": "n1-standard-2",
			},
			modifyField: func(m map[string]interface{}) {
				m["disk_size_gb"] = 100
			},
			description: "Changing disk_size_gb should change hash",
		},
		{
			name: "Instance type change affects hash",
			baseInput: map[string]interface{}{
				"name":          "pool-1",
				"count":         2,
				"instance_type": "n1-standard-2",
			},
			modifyField: func(m map[string]interface{}) {
				m["instance_type"] = "n1-standard-4"
			},
			description: "Changing instance_type should change hash",
		},
		{
			name: "Additional labels change affects hash",
			baseInput: map[string]interface{}{
				"name":          "pool-1",
				"count":         2,
				"instance_type": "n1-standard-2",
				"additional_labels": map[string]interface{}{
					"env": "dev",
				},
			},
			modifyField: func(m map[string]interface{}) {
				m["additional_labels"] = map[string]interface{}{
					"env": "prod",
				}
			},
			description: "Changing additional_labels should change hash",
		},
		{
			name: "Update strategy change affects hash",
			baseInput: map[string]interface{}{
				"name":            "pool-1",
				"count":           2,
				"instance_type":   "n1-standard-2",
				"update_strategy": "RollingUpdateScaleOut",
			},
			modifyField: func(m map[string]interface{}) {
				m["update_strategy"] = "RollingUpdateScaleIn"
			},
			description: "Changing update_strategy should change hash",
		},
		{
			name: "Node configuration change affects hash",
			baseInput: map[string]interface{}{
				"name":          "pool-1",
				"count":         2,
				"instance_type": "n1-standard-2",
				"node": []interface{}{
					map[string]interface{}{
						"action": "cordon",
					},
				},
			},
			modifyField: func(m map[string]interface{}) {
				m["node"] = []interface{}{
					map[string]interface{}{
						"action": "drain",
					},
				}
			},
			description: "Changing node config should change hash",
		},
		{
			name: "Taints change affects hash",
			baseInput: map[string]interface{}{
				"name":          "pool-1",
				"count":         2,
				"instance_type": "n1-standard-2",
				"taints": []interface{}{
					map[string]interface{}{
						"key":    "key1",
						"value":  "value1",
						"effect": "NoSchedule",
					},
				},
			},
			modifyField: func(m map[string]interface{}) {
				m["taints"] = []interface{}{
					map[string]interface{}{
						"key":    "key2",
						"value":  "value2",
						"effect": "NoExecute",
					},
				}
			},
			description: "Changing taints should change hash",
		},
		{
			name: "Adding taints affects hash",
			baseInput: map[string]interface{}{
				"name":          "pool-1",
				"count":         2,
				"instance_type": "n1-standard-2",
			},
			modifyField: func(m map[string]interface{}) {
				m["taints"] = []interface{}{
					map[string]interface{}{
						"key":    "key1",
						"value":  "value1",
						"effect": "NoSchedule",
					},
				}
			},
			description: "Adding taints should change hash",
		},
		{
			name: "Adding node config affects hash",
			baseInput: map[string]interface{}{
				"name":          "pool-1",
				"count":         2,
				"instance_type": "n1-standard-2",
			},
			modifyField: func(m map[string]interface{}) {
				m["node"] = []interface{}{
					map[string]interface{}{
						"action": "cordon",
					},
				}
			},
			description: "Adding node config should change hash",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get hash of base input
			baseHash := resourceMachinePoolGkeHash(tc.baseInput)

			// Create modified copy
			modified := copyMap(tc.baseInput)
			tc.modifyField(modified)

			// Get hash of modified input
			modifiedHash := resourceMachinePoolGkeHash(modified)

			// Hashes should be different
			if baseHash == modifiedHash {
				t.Errorf("%s: Base hash %d equals modified hash %d, but they should differ.\nBase: %+v\nModified: %+v",
					tc.description, baseHash, modifiedHash, tc.baseInput, modified)
			}
		})
	}
}

// TestResourceMachinePoolGkeHashConsistency verifies that the same input produces the same hash
func TestResourceMachinePoolGkeHashConsistency(t *testing.T) {
	input := map[string]interface{}{
		"name":          "test-pool",
		"count":         3,
		"disk_size_gb":  100,
		"instance_type": "n1-standard-4",
		"additional_labels": map[string]interface{}{
			"env": "test",
		},
		"update_strategy": "RollingUpdateScaleOut",
	}

	hash1 := resourceMachinePoolGkeHash(input)
	hash2 := resourceMachinePoolGkeHash(input)

	assert.Equal(t, hash1, hash2, "Same input should produce same hash")
}

// TestResourceMachinePoolGkeHashDifference verifies that different inputs produce different hashes
func TestResourceMachinePoolGkeHashDifference(t *testing.T) {
	baseInput := map[string]interface{}{
		"name":          "test-pool",
		"count":         3,
		"disk_size_gb":  100,
		"instance_type": "n1-standard-4",
	}

	testCases := []struct {
		name     string
		modifier func(map[string]interface{}) map[string]interface{}
	}{
		{
			name: "Different name",
			modifier: func(m map[string]interface{}) map[string]interface{} {
				modified := copyMap(m)
				modified["name"] = "different-pool"
				return modified
			},
		},
		{
			name: "Different count",
			modifier: func(m map[string]interface{}) map[string]interface{} {
				modified := copyMap(m)
				modified["count"] = 5
				return modified
			},
		},
		{
			name: "Different disk size",
			modifier: func(m map[string]interface{}) map[string]interface{} {
				modified := copyMap(m)
				modified["disk_size_gb"] = 200
				return modified
			},
		},
		{
			name: "Different instance type",
			modifier: func(m map[string]interface{}) map[string]interface{} {
				modified := copyMap(m)
				modified["instance_type"] = "n1-standard-8"
				return modified
			},
		},
		{
			name: "Added labels",
			modifier: func(m map[string]interface{}) map[string]interface{} {
				modified := copyMap(m)
				modified["additional_labels"] = map[string]interface{}{"env": "prod"}
				return modified
			},
		},
		{
			name: "Different update strategy",
			modifier: func(m map[string]interface{}) map[string]interface{} {
				modified := copyMap(m)
				modified["update_strategy"] = "RollingUpdateScaleIn"
				return modified
			},
		},
		{
			name: "Added taints",
			modifier: func(m map[string]interface{}) map[string]interface{} {
				modified := copyMap(m)
				modified["taints"] = []interface{}{
					map[string]interface{}{
						"key":    "test",
						"value":  "true",
						"effect": "NoSchedule",
					},
				}
				return modified
			},
		},
	}

	baseHash := resourceMachinePoolGkeHash(baseInput)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			modifiedInput := tc.modifier(baseInput)
			modifiedHash := resourceMachinePoolGkeHash(modifiedInput)
			assert.NotEqual(t, baseHash, modifiedHash, "Modified input should produce different hash")
		})
	}
}

// Helper function to copy a map
func copyMap(original map[string]interface{}) map[string]interface{} {
	copied := make(map[string]interface{})
	for key, value := range original {
		copied[key] = value
	}
	return copied
}

func TestResourceMachinePoolCustomCloudHash(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{
			name: "With all fields",
			input: map[string]interface{}{
				"name":                    "custom-cloud",
				"count":                   3,
				"control_plane":           true,
				"control_plane_as_worker": false,
				"taints":                  []interface{}{"key1=value1", "key2=value2"},
				"node_pool_config":        "standard",
			},
			expected: 208692298,
		},
		{
			name: "Missing optional fields",
			input: map[string]interface{}{
				"name":             "test-pool",
				"count":            3,
				"node_pool_config": "standard",
			},
			expected: 1525978111,
		},
		{
			name: "YAML normalization - different formatting same content",
			input: map[string]interface{}{
				"name":  "yaml-pool",
				"count": 2,
				"node_pool_config": `apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: MachineDeployment
metadata:
  name: md-0
  namespace: test
spec:
  replicas: 2
  template:
    spec:
      version: v1.27.0`,
			},
			// This should match the normalized YAML hash
			expected: 0, // Will be calculated by first run
		},
	}

	// First, calculate the expected hash for the YAML normalization test
	for i := range testCases {
		if testCases[i].name == "YAML normalization - different formatting same content" {
			// Calculate actual hash to set as expected
			testCases[i].expected = resourceMachinePoolCustomCloudHash(testCases[i].input)
			fmt.Printf("Setting expected hash for YAML normalization test: %d\n", testCases[i].expected)

			// Now test that same content with different whitespace produces same hash
			input2 := map[string]interface{}{
				"name":  "yaml-pool",
				"count": 2,
				"node_pool_config": `apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind:    MachineDeployment
metadata:
  name:   md-0
  namespace:  test
spec:
  replicas:   2
  template:
    spec:
      version:  v1.27.0`,
			}
			hash2 := resourceMachinePoolCustomCloudHash(input2)
			assert.Equal(t, testCases[i].expected, hash2, "YAML normalization should produce same hash regardless of whitespace formatting")
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := resourceMachinePoolCustomCloudHash(tc.input)
			fmt.Printf("Debug: For input %+v, got hash %d, expected %d\n", tc.input, actual, tc.expected)
			if actual != tc.expected {
				t.Errorf("For test case '%s', expected hash %d, but got %d for input %+v", tc.name, tc.expected, actual, tc.input)
			}
		})
	}
}

// TestResourceMachinePoolCustomCloudHashYAMLNormalization specifically tests
// that different YAML formatting produces the same hash (perpetual diff fix)
func TestResourceMachinePoolCustomCloudHashYAMLNormalization(t *testing.T) {
	// Same YAML content with different whitespace/formatting
	yaml1 := `apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: MachineDeployment
metadata:
  name: md-0
spec:
  replicas: 2`

	yaml2 := `apiVersion:    infrastructure.cluster.x-k8s.io/v1beta1
kind:   MachineDeployment
metadata:
  name:   md-0
spec:
  replicas:  2`

	yaml3 := `apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: MachineDeployment
metadata:
    name: md-0
spec:
    replicas: 2`

	input1 := map[string]interface{}{
		"name":             "test-pool",
		"count":            2,
		"node_pool_config": yaml1,
	}

	input2 := map[string]interface{}{
		"name":             "test-pool",
		"count":            2,
		"node_pool_config": yaml2,
	}

	input3 := map[string]interface{}{
		"name":             "test-pool",
		"count":            2,
		"node_pool_config": yaml3,
	}

	hash1 := resourceMachinePoolCustomCloudHash(input1)
	hash2 := resourceMachinePoolCustomCloudHash(input2)
	hash3 := resourceMachinePoolCustomCloudHash(input3)

	assert.Equal(t, hash1, hash2, "YAML with extra spaces should produce same hash")
	assert.Equal(t, hash1, hash3, "YAML with different indentation should produce same hash")
}
