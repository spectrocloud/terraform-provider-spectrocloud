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
			name: "Test Valid ResourceMachinePoolAksHash",
			input: map[string]interface{}{
				"instance_type":        "Standard_D2s_v3",
				"disk_size_gb":         100,
				"is_system_node_pool":  true,
				"storage_account_type": "Premium_LRS",
			},
			expected: 380130606,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := resourceMachinePoolAksHash(tc.input)
			assert.Equal(t, tc.expected, result)
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
				"azs":           []string{"us-central1-a", "us-central1-b"},
				"az_subnets": map[string]interface{}{
					"us-central1-a": "subnet-1",
					"us-central1-b": "subnet-2",
				},
			},
			expected: 1198721703,
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
			expected: 456946481,
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

func TestResourceMachinePoolCoxEdgeHash(t *testing.T) {

	testCases := []struct {
		input    map[string]interface{}
		expected int
	}{
		{

			input:    commonNodePool(),
			expected: 513591628,
		},
		{
			// Test case with empty input
			input:    nil,
			expected: 2166136261,
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			// Call the function with the test input
			result := resourceMachinePoolCoxEdgeHash(tc.input)

			// Check if the result matches the expected output
			if result != tc.expected {
				t.Errorf("Expected: %d, Got: %d", tc.expected, result)
			}
		})
	}
}

func TestResourceMachinePoolTkeHash(t *testing.T) {
	testCases := []struct {
		input    map[string]interface{}
		expected int
	}{
		{

			input: map[string]interface{}{
				"az_subnets": map[string]interface{}{
					"subnet1": "10.0.0.1",
					"subnet2": "10.0.0.2",
				},
			},
			expected: 3634270287,
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			// Call the function with the test input
			result := resourceMachinePoolTkeHash(tc.input)

			// Check if the result matches the expected output
			if result != tc.expected {
				t.Errorf("Expected: %d, Got: %d", tc.expected, result)
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
			input: map[string]interface{}{
				"host_uids": []interface{}{"host1", "host2", "host3"},
			},
			expected: 456992116,
		},
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

func TestResourceMachinePoolLibvirtHash(t *testing.T) {
	testCases := []struct {
		name         string
		input        interface{}
		expectedHash int
	}{
		{
			name: "Valid MachinePoolLibvirtHash",
			input: map[string]interface{}{
				"xsl_template": "xsl-template-1",
				"instance_type": []interface{}{
					map[string]interface{}{
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
				},
			},
			expectedHash: 3451728783,
		},
		{
			name:         "Test Case 2",
			input:        map[string]interface{}{},
			expectedHash: 2166136261,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash := resourceMachinePoolLibvirtHash(tc.input)
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
				"azs": schema.NewSet(schema.HashString, []interface{}{"az1", "az2"}),
			},
			expectedHash: 3363048657,
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
				"instance_type":   "flavor1",
				"subnet_id":       "subnet123",
				"update_strategy": "RollingUpdate",
				"azs":             schema.NewSet(schema.HashString, []interface{}{"az1", "az2"}),
			},
			expectedHash: 3148662768,
		},
		{
			name: "Valid MachinePoolOpenStackHash 2",
			input: map[string]interface{}{
				"instance_type":   "flavor2",
				"subnet_id":       "subnet456",
				"update_strategy": "Recreate",
				"azs":             schema.NewSet(schema.HashString, []interface{}{"az3"}),
			},
			expectedHash: 4045757255,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash := resourceMachinePoolOpenStackHash(tc.input)
			assert.Equal(t, tc.expectedHash, hash)
		})
	}
}
