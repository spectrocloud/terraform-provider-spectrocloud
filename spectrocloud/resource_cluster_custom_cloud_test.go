package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"testing"

	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/palette-sdk-go/api/models"
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
	_ = mockResourceData.Set("name", "test-cluster")
	_ = mockResourceData.Set("context", "project")
	_ = mockResourceData.Set("tags", []string{"tf:test", "env:dev"})
	_ = mockResourceData.Set("cloud", "nutanix")
	_ = mockResourceData.Set("description", "test description")
	_ = mockResourceData.Set("cloud_config_id", "config123")
	_ = mockResourceData.Set("cluster_profile", []interface{}{
		map[string]interface{}{
			"id": "cluster-profile-id",
		},
	})
	_ = mockResourceData.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"values": "test-config",
		},
	})
	_ = mockResourceData.Set("machine_pool", []interface{}{
		map[string]interface{}{
			"node_pool_config": "test-config-yaml",
		},
	})
	var location []interface{}
	location = append(location, map[string]interface{}{
		"country_code": "ind",
		"country_name": "india",
		"region_code":  "MZ",
		"region_name":  "mumbai",
		"latitude":     0.0,
		"longitude":    0.0,
	})
	_ = mockResourceData.Set("location_config", location)

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
	_ = mockResourceData.Set("cloud_config", []interface{}{
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
	_ = mockResourceData.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"values": "test-values",
		},
	})
	_ = mockResourceData.Set("machine_pool", []interface{}{
		map[string]interface{}{
			"control_plane":           true,
			"control_plane_as_worker": false,
			"node_pool_config":        "test-node-pool-config",
		},
	})
	_ = mockResourceData.Set("context", "project")
	_ = mockResourceData.Set("cloud_account_id", "test-cloud-account-id")

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

func boolPtr(b bool) *bool {
	return &b
}

func TestFlattenMachinePoolConfigsCustomCloud(t *testing.T) {
	// Define test cases
	tests := []struct {
		name         string
		input        []*models.V1CustomMachinePoolConfig
		expectedSize int
		expected     []interface{}
	}{
		{
			name:         "Empty machine pool",
			input:        []*models.V1CustomMachinePoolConfig{},
			expectedSize: 0,
			expected:     []interface{}{},
		},
		{
			name: "Single machine pool",
			input: []*models.V1CustomMachinePoolConfig{
				{
					UseControlPlaneAsWorker: true,
					IsControlPlane:          boolPtr(true),
					Values:                  "node-pool-values",
					Name:                    "pool-1",
					Size:                    int32(3),
				},
			},
			expectedSize: 1,
			expected: []interface{}{
				map[string]interface{}{
					"control_plane_as_worker": true,
					"control_plane":           boolPtr(true),
					"node_pool_config":        "node-pool-values",
					"name":                    "pool-1",
					"count":                   int32(3),
				},
			},
		},
		{
			name: "Multiple machine pools",
			input: []*models.V1CustomMachinePoolConfig{
				{
					UseControlPlaneAsWorker: true,
					IsControlPlane:          boolPtr(true),
					Values:                  "node-pool-1-values",
					Name:                    "pool-1",
					Size:                    int32(3),
				},
				{
					UseControlPlaneAsWorker: false,
					IsControlPlane:          boolPtr(false),
					Values:                  "node-pool-2-values",
					Name:                    "pool-2",
					Size:                    int32(5),
				},
			},
			expectedSize: 2,
			expected: []interface{}{
				map[string]interface{}{
					"control_plane_as_worker": true,
					"control_plane":           boolPtr(true),
					"node_pool_config":        "node-pool-1-values",
					"name":                    "pool-1",
					"count":                   int32(3),
				},
				map[string]interface{}{
					"control_plane_as_worker": false,
					"control_plane":           boolPtr(false),
					"node_pool_config":        "node-pool-2-values",
					"name":                    "pool-2",
					"count":                   int32(5),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := flattenMachinePoolConfigsCustomCloud(tt.input)
			assert.Equal(t, tt.expectedSize, len(output))
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestParseResourceCustomCloudImportID(t *testing.T) {
	tests := []struct {
		id                  string
		expectedClusterID   string
		expectedScope       string
		expectedCloudName   string
		expectedError       bool
		expectedErrorString string
	}{
		{
			id:                "cluster456:project:nutanix",
			expectedClusterID: "cluster456",
			expectedScope:     "project",
			expectedCloudName: "nutanix",
			expectedError:     false,
		},
		{
			id:                "cluster789:tenant:oracle",
			expectedClusterID: "cluster789",
			expectedScope:     "tenant",
			expectedCloudName: "oracle",
			expectedError:     false,
		},
		{
			id:                  "cluster123:invalid:gcp",
			expectedClusterID:   "",
			expectedScope:       "",
			expectedCloudName:   "",
			expectedError:       true,
			expectedErrorString: "invalid cluster ID format specified for import custom cloud cluster123:invalid:gcp, Ex: it should cluster_id:context:custom_cloud_name, `cluster456:project:nutanix`",
		},
		{
			id:                  "cluster456:project",
			expectedClusterID:   "",
			expectedScope:       "",
			expectedCloudName:   "",
			expectedError:       true,
			expectedErrorString: "invalid cluster ID format specified for import custom cloud cluster456:project, Ex: it should cluster_id:context:custom_cloud_name, `cluster456:project:nutanix`",
		},
		{
			id:                  "cluster456:tenant:",
			expectedClusterID:   "",
			expectedScope:       "",
			expectedCloudName:   "",
			expectedError:       true,
			expectedErrorString: "invalid cluster ID format specified for import custom cloud cluster456:tenant:, Ex: it should cluster_id:context:custom_cloud_name, `cluster456:project:nutanix`",
		},
	}

	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			resourceData := resourceClusterCustomCloud().TestResourceData()
			resourceData.SetId(test.id)

			clusterID, scope, customCloudName, err := ParseResourceCustomCloudImportID(resourceData)

			if test.expectedError {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expectedErrorString)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedClusterID, clusterID)
				assert.Equal(t, test.expectedScope, scope)
				assert.Equal(t, test.expectedCloudName, customCloudName)
			}
		})
	}
}

func prepareClusterCustomTestData() *schema.ResourceData {
	d := resourceClusterCustomCloud().TestResourceData()
	_ = d.Set("name", "test-custom-cluster")
	_ = d.Set("context", "project")
	_ = d.Set("cloud", "test-cloud")
	_ = d.Set("tags", []string{"tf:unit", "env:dev"})
	_ = d.Set("description", "test description")
	_ = d.Set("cluster_profile", []interface{}{
		map[string]interface{}{
			"id": "test-cluster-profile-id",
			"pack": []interface{}{
				map[string]interface{}{
					"uid":          "pack-uid-1",
					"type":         "spectro",
					"name":         "k8",
					"registry_uid": "test-regi-uid",
					"tag":          "test",
					"values":       "test-pack-value",
				},
				map[string]interface{}{
					"uid":          "pack-uid-2",
					"type":         "manifest",
					"name":         "csi",
					"registry_uid": "test-regi-uid",
					"tag":          "test",
					"values":       "test-pack-value",
					"manifest": []interface{}{
						map[string]interface{}{
							"uid":     "test-manifest-id-1",
							"name":    "test-csi",
							"content": "test-content",
						},
					},
				},
				map[string]interface{}{
					"uid":          "pack-uid-3",
					"type":         "manifest",
					"name":         "cni",
					"registry_uid": "test-regi-uid",
					"tag":          "test",
					"values":       "test-pack-value",
					"manifest": []interface{}{
						map[string]interface{}{
							"uid":     "test-manifest-id-2",
							"name":    "test-cni",
							"content": "test-content",
						},
					},
				},
			},
		},
	})
	_ = d.Set("apply_setting", "DownloadAndInstall")
	_ = d.Set("cloud_account_id", "test-cloud-id")
	_ = d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"values": "test-custom-cloud-config/values",
		},
	})
	_ = d.Set("machine_pool", []interface{}{
		map[string]interface{}{
			"name":                    "test-cp-pool",
			"count":                   1,
			"control_plane":           true,
			"control_plane_as_worker": false,
			"node_pool_config":        "node-pool-config-values",
		},
	})
	_ = d.Set("pause_agent_upgrades", "unlock")
	_ = d.Set("os_patch_on_boot", false)
	_ = d.Set("os_patch_schedule", "0 0 * * *")

	_ = d.Set("skip_completion", true)
	_ = d.Set("backup_policy", []interface{}{
		map[string]interface{}{
			"prefix":                    "test",
			"backup_location_id":        "backup-location-id",
			"schedule":                  "0 1 * * *",
			"expiry_in_hour":            5,
			"include_disks":             true,
			"include_cluster_resources": true,
			"namespaces":                []string{"default"},
			"include_all_clusters":      true,
		},
	})
	_ = d.Set("scan_policy", []interface{}{
		map[string]interface{}{
			"configuration_scan_schedule": "0 1 * * *",
			"penetration_scan_schedule":   "0 1 * * *",
			"conformance_scan_schedule":   "0 1 * * *",
		},
	})
	_ = d.Set("cluster_rbac_binding", []interface{}{
		map[string]interface{}{
			"type":      "RoleBinding",
			"namespace": "default",
			"role": map[string]interface{}{
				"kind": "test",
				"name": "test",
			},
			"subjects": []interface{}{
				map[string]interface{}{
					"type":      "User",
					"name":      "test-subject",
					"namespace": "default",
				},
			},
		},
	})

	return d
}

func TestResourceClusterCustomCloudCreate(t *testing.T) {
	d := prepareClusterCustomTestData()
	var diags diag.Diagnostics
	var ctx context.Context
	diags = resourceClusterCustomCloudCreate(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}

func TestResourceClusterCustomCloudInvalidCloud(t *testing.T) {
	d := prepareClusterCustomTestData()
	_ = d.Set("cloud", "1234")
	var diags diag.Diagnostics
	var ctx context.Context
	diags = resourceClusterCustomCloudCreate(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 1, len(diags))
}
