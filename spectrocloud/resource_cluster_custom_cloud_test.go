package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/palette-api-go/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/stretchr/testify/assert"
)

func TestFlattenCloudConfigsValuesCustomCloud(t *testing.T) {
	// Test case 1: When config is nil
	result := flattenCloudConfigsValuesCustomCloud(nil)
	assert.Len(t, result, 0, "Expected empty slice when config is nil")

	// Test case 2: When config.Spec is nil
	config := &models.V1CustomCloudConfig{}
	result = flattenCloudConfigsValuesCustomCloud(config)
	assert.Len(t, result, 0, "Expected empty slice when config.Spec is nil")

	// Test case 3: When config.Spec.ClusterConfig is nil
	config.Spec = &models.V1CustomCloudConfigSpec{}
	result = flattenCloudConfigsValuesCustomCloud(config)
	assert.Len(t, result, 0, "Expected empty slice when config.Spec.ClusterConfig is nil")

	// Test case 4: When config.Spec.ClusterConfig.Values is not nil
	config.Spec.ClusterConfig = &models.V1CustomClusterConfig{
		Values: ptr.StringPtr("test-values"),
	}
	result = flattenCloudConfigsValuesCustomCloud(config)
	assert.Len(t, result, 1, "Expected one element in the slice")
	assert.Equal(t, "test-values", result[0].(map[string]interface{})["values"], "Values should match")
}

//func TestFlattenCloudConfigCustom(t *testing.T) {
//	// Mock resource data
//	mockResourceData := resourceClusterCustomCloud().TestResourceData()
//	mockResourceData.Set("context", "project")
//	mockResourceData.Set("cloud", "aws")
//	mockResourceData.Set("cloud_config_id", "config123")
//
//	var mps []*models.V1CustomMachinePoolConfig
//	mps = append(mps, &models.V1CustomMachinePoolConfig{
//		AdditionalLabels:        nil,
//		IsControlPlane:          ptr.BoolPtr(true),
//		Name:                    "cp-pool",
//		Size:                    1,
//		Taints:                  nil,
//		UseControlPlaneAsWorker: true,
//		Values:                  "-- test yaml string",
//	})
//
//	// Mock client
//	mockClient := &client.V1Client{}
//
//	// Call the function with mocked dependencies
//	diags, _ := flattenCloudConfigCustom("config123", mockResourceData, mockClient)
//
//	var emptyErr diag.Diagnostics
//	// Assertions
//	assert.Equal(t, emptyErr, diags)
//
//	// Assert resource data values
//	assert.Equal(t, "config123", mockResourceData.Get("cloud_config_id"))
//	assert.Equal(t, "account123", mockResourceData.Get("cloud_account_id"))
//
//}

func TestToMachinePoolCustomCloud(t *testing.T) {
	// Test case 1: Valid machine pool configuration
	machinePool := map[string]interface{}{
		"node_pool_config":        "config123",
		"control_plane":           true,
		"control_plane_as_worker": true,
	}

	expected := &models.V1CustomMachinePoolConfigEntity{
		CloudConfig: &models.V1CustomMachinePoolCloudConfigEntity{
			Values: "config123",
		},
		PoolConfig: &models.V1CustomMachinePoolBaseConfigEntity{
			IsControlPlane:          true,
			UseControlPlaneAsWorker: true,
			// Set other fields as expected
		},
	}

	actual := toMachinePoolCustomCloud(machinePool)

	assert.Equal(t, expected, actual)
}

func TestToCustomClusterConfig(t *testing.T) {
	// Create a mock schema.ResourceData with relevant data for testing
	mockResourceData := resourceClusterCustomCloud().TestResourceData()
	mockResourceData.Set("name", "test-cluster")
	mockResourceData.Set("context", "project")
	mockResourceData.Set("tags", []string{"tf:test", "env:dev"})
	mockResourceData.Set("cloud", "nutanix")
	mockResourceData.Set("description", "test description")
	mockResourceData.Set("cloud_config_id", "config123")
	mockResourceData.Set("cluster_profile", []interface{}{
		map[string]interface{}{
			"id": "cluster-profile-id",
		},
	})
	mockResourceData.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"values": "test-config",
		},
	})
	mockResourceData.Set("machine_pool", []interface{}{
		map[string]interface{}{
			"node_pool_config": "test-config-yaml",
		},
	})
	mockResourceData.Set("location_config", map[string]interface{}{
		"country_code": "ind",
		"country_name": "india",
		"region_code":  "MZ",
		"region_name":  "mumbai",
		"latitude":     "N12312",
		"longitude":    "S12312",
	})

	expected := &models.V1CustomClusterConfigEntity{
		Location:                toClusterLocationConfigs(mockResourceData),
		MachineManagementConfig: toMachineManagementConfig(mockResourceData),
		Resources:               toClusterResourceConfig(mockResourceData),
	}

	actual := toCustomClusterConfig(mockResourceData)

	assert.Equal(t, expected, actual)
}

func TestToCustomCloudConfig(t *testing.T) {
	// Create mock schema.ResourceData with relevant data for testing
	mockResourceData := resourceClusterCustomCloud().TestResourceData()
	mockResourceData.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"values": "mock values YAML string",
		},
	})

	expectedValues := "mock values YAML string"

	// Call the toCustomCloudConfig function with the mock schema.ResourceData
	customCloudConfig := toCustomCloudConfig(mockResourceData)

	// Assert that the returned customCloudConfig has the expected values
	assert.NotNil(t, customCloudConfig)
	assert.Equal(t, expectedValues, *customCloudConfig.Values)
}

func TestToCustomCloudCluster(t *testing.T) {
	// Mock schema.ResourceData
	mockResourceData := resourceClusterCustomCloud().TestResourceData()
	mockResourceData.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"values": "test-values",
		},
	})
	mockResourceData.Set("machine_pool", []interface{}{
		map[string]interface{}{
			"control_plane":           true,
			"control_plane_as_worker": false,
			"node_pool_config":        "test-node-pool-config",
		},
	})
	mockResourceData.Set("context", "project")
	mockResourceData.Set("cloud_account_id", "test-cloud-account-id")

	// Mock client.V1Client
	mockClient := &client.V1Client{
		// Mock any required methods or behaviors
	}

	// Call the toCustomCloudCluster function with the mock objects
	cluster, err := toCustomCloudCluster(mockClient, mockResourceData)

	// Assertions
	assert.NoError(t, err) // Ensure no error occurred
	assert.NotNil(t, cluster)
	assert.Equal(t, ptr.StringPtr("test-cloud-account-id"), cluster.Spec.CloudAccountUID) // Verify CloudAccountUID
	assert.NotNil(t, cluster.Spec.CloudConfig)                                            // Verify CloudConfig
	assert.NotNil(t, cluster.Spec.ClusterConfig)                                          // Verify ClusterConfig
	assert.NotNil(t, cluster.Spec.Machinepoolconfig)                                      // Verify Machinepoolconfig
	assert.NotNil(t, cluster.Spec.Profiles)                                               // Verify Profiles
}

//func TestResourceClusterCustomCloudUpdate(t *testing.T) {
//	// Mock schema.ResourceData with necessary fields
//	mockResourceData := resourceClusterCustomCloud().TestResourceData()
//	mockResourceData.Set("cloud_config", []interface{}{
//		map[string]interface{}{
//			"values": "test-values",
//		},
//	})
//	mockResourceData.Set("machine_pool", []interface{}{
//		map[string]interface{}{
//			"control_plane":           true,
//			"control_plane_as_worker": false,
//			"node_pool_config":        "test-node-pool-config",
//		},
//	})
//	mockResourceData.Set("context", "project")
//	mockResourceData.Set("cloud", "custom-cloud")
//	mockResourceData.Set("cloud_account_id", "test-cloud-account-id")
//
//	var mps []*models.V1CustomMachinePoolConfig
//	mps = append(mps, &models.V1CustomMachinePoolConfig{
//		AdditionalLabels:        nil,
//		IsControlPlane:          ptr.BoolPtr(true),
//		Name:                    "cp-pool",
//		Size:                    1,
//		Taints:                  nil,
//		UseControlPlaneAsWorker: true,
//		Values:                  "-- test yaml string",
//	})
//
//	// Mock client.V1Client
//	mockClient := &client.V1Client{}
//
//	// Call the resourceClusterCustomCloudUpdate function with mock objects
//	diags := resourceClusterCustomCloudUpdate(context.Background(), mockResourceData, mockClient)
//
//	// Assertions
//	var d diag.Diagnostics
//	assert.Equal(t, d, diags)
//
//}

//func TestResourceClusterCustomCloudCreate(t *testing.T) {
//	// Mock schema.ResourceData with necessary fields
//	mockResourceData := resourceClusterCustomCloud().TestResourceData()
//	mockResourceData.Set("cloud_config", []interface{}{
//		map[string]interface{}{
//			"values": "test-values",
//		},
//	})
//	mockResourceData.Set("machine_pool", []interface{}{
//		map[string]interface{}{
//			"control_plane":           true,
//			"control_plane_as_worker": false,
//			"node_pool_config":        "test-node-pool-config",
//		},
//	})
//	mockResourceData.Set("context", "project")
//	mockResourceData.Set("cloud", "custom-cloud")
//	mockResourceData.Set("cloud_account_id", "test-cloud-account-id")
//	mockResourceData.Set("skip_completion", true)
//
//	// Mock client.V1Client
//	mockClient := &client.V1Client{}
//
//	// Call the resourceClusterCustomCloudCreate function with mock objects
//	diags := resourceClusterCustomCloudCreate(context.Background(), mockResourceData, mockClient)
//
//	// Assertions
//	var d diag.Diagnostics
//	assert.Equal(t, d, diags)
//}
