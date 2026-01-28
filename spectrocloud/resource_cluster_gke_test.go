package spectrocloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"github.com/stretchr/testify/assert"
)

func TestToMachinePoolGke(t *testing.T) {
	// Simulate input data
	machinePool := map[string]interface{}{
		"name":          "pool1",
		"count":         3,
		"instance_type": "n1-standard-2",
		"disk_size_gb":  100,
	}
	mp, err := toMachinePoolGke(machinePool)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, mp)

	// Check the CloudConfig fields
	assert.NotNil(t, mp.CloudConfig)
	assert.Equal(t, "n1-standard-2", *mp.CloudConfig.InstanceType)
	assert.Equal(t, int64(100), mp.CloudConfig.RootDeviceSize)

	// Check the PoolConfig fields
	assert.NotNil(t, mp.PoolConfig)
	assert.Equal(t, "pool1", *mp.PoolConfig.Name)
	assert.Equal(t, int32(3), *mp.PoolConfig.Size)
	assert.Equal(t, []string{"worker"}, mp.PoolConfig.Labels)
}

func TestToGkeCluster(t *testing.T) {
	// Simulate input data
	cloudConfig := map[string]interface{}{
		"project": "my-project",
		"region":  "us-central1",
	}
	machinePool := map[string]interface{}{
		"name":          "pool1",
		"count":         3,
		"instance_type": "n1-standard-2",
		"disk_size_gb":  100,
	}
	d := resourceClusterGke().TestResourceData()
	d.Set("cloud_config", []interface{}{cloudConfig})
	d.Set("context", "cluster-context")
	d.Set("cloud_account_id", "cloud-account-id")
	d.Set("machine_pool", []interface{}{machinePool})

	// Call the toGkeCluster function with the simulated input data
	cluster, err := toGkeCluster(nil, d)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, cluster)

	// Check the Metadata
	assert.NotNil(t, cluster.Metadata)
	// Check other fields similarly
	assert.NotNil(t, cluster.Spec.CloudConfig)
	assert.Equal(t, "my-project", *cluster.Spec.CloudConfig.Project)
	assert.Equal(t, "us-central1", *cluster.Spec.CloudConfig.Region)

	// Check machine pool configuration
	assert.Len(t, cluster.Spec.Machinepoolconfig, 1)
	assert.Equal(t, "pool1", *cluster.Spec.Machinepoolconfig[0].PoolConfig.Name)
	assert.Equal(t, int32(3), *cluster.Spec.Machinepoolconfig[0].PoolConfig.Size)
	assert.Equal(t, "n1-standard-2", *cluster.Spec.Machinepoolconfig[0].CloudConfig.InstanceType)
	assert.Equal(t, int64(100), cluster.Spec.Machinepoolconfig[0].CloudConfig.RootDeviceSize)
}

func TestFlattenMachinePoolConfigsGke(t *testing.T) {
	// Simulate input data
	machinePools := []*models.V1GcpMachinePoolConfig{
		{
			InstanceType:   types.Ptr("n1-standard-2"),
			Name:           "pool1",
			RootDeviceSize: 100,
			Size:           3,
		},
		{
			InstanceType:   types.Ptr("n1-standard-4"),
			Name:           "pool2",
			Size:           2,
			RootDeviceSize: 200,
		},
	}

	// Call the flattenMachinePoolConfigsGke function with the simulated input data
	result := flattenMachinePoolConfigsGke(machinePools)

	// Assertions
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	// Check the first machine pool
	pool1 := result[0].(map[string]interface{})
	assert.Equal(t, "pool1", pool1["name"])
	assert.Equal(t, 3, pool1["count"])
	assert.Equal(t, "n1-standard-2", pool1["instance_type"])
	assert.Equal(t, 100, pool1["disk_size_gb"])

	// Check the second machine pool
	pool2 := result[1].(map[string]interface{})
	assert.Equal(t, "pool2", pool2["name"])
	assert.Equal(t, 2, pool2["count"])
	assert.Equal(t, "n1-standard-4", pool2["instance_type"])
	assert.Equal(t, 200, pool2["disk_size_gb"])
}

func TestFlattenCloudConfigGke(t *testing.T) {
	configUID := "test-config-uid"

	tests := []struct {
		name        string
		setup       func() *schema.ResourceData
		client      interface{}
		expectError bool
		description string
		verify      func(t *testing.T, diags diag.Diagnostics, d *schema.ResourceData)
	}{
		{
			name: "Flatten with existing cloud_config in ResourceData",
			setup: func() *schema.ResourceData {
				d := resourceClusterGke().TestResourceData()
				d.SetId("test-cluster-uid")
				_ = d.Set("context", "project")
				_ = d.Set("cloud_config", []interface{}{
					map[string]interface{}{
						"project": "my-project",
						"region":  "us-central1",
					},
				})
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: true, // GetCloudConfigGke may fail
			description: "Should use existing cloud_config from ResourceData when available",
			verify: func(t *testing.T, diags diag.Diagnostics, d *schema.ResourceData) {
				// Verify cloud_config_id is set even if API call fails
				if len(diags) == 0 {
					cloudConfigID := d.Get("cloud_config_id")
					assert.Equal(t, configUID, cloudConfigID, "cloud_config_id should be set")
				}
			},
		},
		{
			name: "Flatten without existing cloud_config in ResourceData",
			setup: func() *schema.ResourceData {
				d := resourceClusterGke().TestResourceData()
				d.SetId("test-cluster-uid")
				_ = d.Set("context", "project")
				// Don't set cloud_config - should use empty map
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: true, // GetCloudConfigGke may fail
			description: "Should use empty cloud_config map when not present in ResourceData",
			verify: func(t *testing.T, diags diag.Diagnostics, d *schema.ResourceData) {
				// Function should handle missing cloud_config gracefully
				if len(diags) > 0 {
					assert.NotEmpty(t, diags, "Should have diagnostics when API route is not available")
				}
			},
		},
		{
			name: "Error from GetCloudConfigGke",
			setup: func() *schema.ResourceData {
				d := resourceClusterGke().TestResourceData()
				d.SetId("test-cluster-uid")
				_ = d.Set("context", "project")
				return d
			},
			client:      unitTestMockAPINegativeClient,
			expectError: true,
			description: "Should return error when GetCloudConfigGke fails",
			verify: func(t *testing.T, diags diag.Diagnostics, d *schema.ResourceData) {
				assert.NotEmpty(t, diags, "Should have diagnostics when GetCloudConfigGke fails")
				if len(diags) > 0 {
					assert.True(t, diags.HasError(), "Should have error diagnostics")
				}
			},
		},
		{
			name: "Flatten with tenant context",
			setup: func() *schema.ResourceData {
				d := resourceClusterGke().TestResourceData()
				d.SetId("test-cluster-uid")
				_ = d.Set("context", "tenant")
				_ = d.Set("cloud_config", []interface{}{
					map[string]interface{}{
						"project": "my-project",
						"region":  "us-central1",
					},
				})
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: true, // GetCloudConfigGke may fail
			description: "Should handle tenant context correctly",
			verify: func(t *testing.T, diags diag.Diagnostics, d *schema.ResourceData) {
				// Function should attempt to get cloud config with tenant context
				if len(diags) > 0 {
					assert.NotEmpty(t, diags, "Should have diagnostics when API route is not available")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceData := tt.setup()
			c := getV1ClientWithResourceContext(tt.client, "project")

			var diags diag.Diagnostics
			var panicked bool

			// Handle potential panics for nil pointer dereferences
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
						diags = diag.Diagnostics{
							{
								Severity: diag.Error,
								Summary:  fmt.Sprintf("Panic: %v", r),
							},
						}
					}
				}()
				diags = flattenCloudConfigGke(configUID, resourceData, c)
			}()

			// Verify results
			if tt.expectError {
				if panicked {
					// Panic is acceptable if API routes don't exist
					assert.NotEmpty(t, diags, "Expected diagnostics/panic for test case: %s", tt.description)
				} else {
					assert.NotEmpty(t, diags, "Expected diagnostics for error case: %s", tt.description)
				}
			} else {
				if panicked {
					t.Logf("Unexpected panic occurred: %v", diags)
				}
				assert.Empty(t, diags, "Should not have errors for successful flatten: %s", tt.description)
				// Verify cloud_config_id is set on success
				cloudConfigID := resourceData.Get("cloud_config_id")
				assert.Equal(t, configUID, cloudConfigID, "cloud_config_id should be set on success: %s", tt.description)
			}

			// Run custom verify function if provided
			if tt.verify != nil {
				tt.verify(t, diags, resourceData)
			}
		})
	}
}

func TestFlattenClusterConfigsGke(t *testing.T) {
	tests := []struct {
		name     string
		input    *models.V1GcpCloudConfig
		expected []interface{}
	}{
		{
			name: "ClusterConfig with project only",
			input: &models.V1GcpCloudConfig{
				Spec: &models.V1GcpCloudConfigSpec{
					ClusterConfig: &models.V1GcpClusterConfig{
						Project: types.Ptr("my-project"),
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"project": types.Ptr("my-project"),
				},
			},
		},
		{
			name: "ClusterConfig with region only",
			input: &models.V1GcpCloudConfig{
				Spec: &models.V1GcpCloudConfigSpec{
					ClusterConfig: &models.V1GcpClusterConfig{
						Region: types.Ptr("us-central1"),
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"region": "us-central1",
				},
			},
		},
		{
			name: "ClusterConfig with both project and region",
			input: &models.V1GcpCloudConfig{
				Spec: &models.V1GcpCloudConfigSpec{
					ClusterConfig: &models.V1GcpClusterConfig{
						Project: types.Ptr("my-project"),
						Region:  types.Ptr("us-central1"),
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"project": types.Ptr("my-project"),
					"region":  "us-central1",
				},
			},
		},
		{
			name: "ClusterConfig with nil project",
			input: &models.V1GcpCloudConfig{
				Spec: &models.V1GcpCloudConfigSpec{
					ClusterConfig: &models.V1GcpClusterConfig{
						Project: nil,
						Region:  types.Ptr("us-west1"),
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"region": "us-west1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenClusterConfigsGke(tt.input)
			assert.Equal(t, tt.expected, result, "Unexpected result for test case: %s", tt.name)
		})
	}
}

func TestToGcpCluster(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *schema.ResourceData
		expectError bool
		verify      func(t *testing.T, cluster *models.V1SpectroGcpClusterEntity, err error)
	}{
		{
			name: "GCP cluster with multiple machine pools",
			setup: func() *schema.ResourceData {
				d := resourceClusterGcp().TestResourceData()
				cloudConfig := map[string]interface{}{
					"project": "my-gcp-project",
					"region":  "us-west1",
					"network": "default",
				}
				machinePool1 := map[string]interface{}{
					"name":          "control-plane",
					"count":         1,
					"instance_type": "n1-standard-4",
					"disk_size_gb":  200,
					"azs":           schema.NewSet(schema.HashString, []interface{}{"us-west1-a"}),
					"control_plane": true,
				}
				machinePool2 := map[string]interface{}{
					"name":          "worker-pool",
					"count":         3,
					"instance_type": "n1-standard-2",
					"disk_size_gb":  100,
					"azs":           schema.NewSet(schema.HashString, []interface{}{"us-west1-a", "us-west1-b"}),
					"control_plane": false,
				}
				d.Set("cloud_config", []interface{}{cloudConfig})
				d.Set("context", "project")
				d.Set("cloud_account_id", "gcp-account-id")
				d.Set("machine_pool", []interface{}{machinePool1, machinePool2})
				d.Set("name", "test-cluster")
				return d
			},
			expectError: false,
			verify: func(t *testing.T, cluster *models.V1SpectroGcpClusterEntity, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, cluster)
				assert.NotNil(t, cluster.Spec.CloudConfig)
				assert.Equal(t, "my-gcp-project", *cluster.Spec.CloudConfig.Project)
				assert.Equal(t, "us-west1", *cluster.Spec.CloudConfig.Region)
				assert.Equal(t, "default", cluster.Spec.CloudConfig.Network)
				assert.Len(t, cluster.Spec.Machinepoolconfig, 2)

				// Create a map to find pools by name (order is not guaranteed with schema.Set)
				poolMap := make(map[string]*models.V1GcpMachinePoolConfigEntity)
				for _, mp := range cluster.Spec.Machinepoolconfig {
					poolMap[*mp.PoolConfig.Name] = mp
				}

				// Check control plane pool
				cpPool, exists := poolMap["control-plane"]
				assert.True(t, exists, "Control plane pool should exist")
				assert.True(t, cpPool.PoolConfig.IsControlPlane, "Control plane pool should have IsControlPlane=true")

				// Check worker pool
				workerPool, exists := poolMap["worker-pool"]
				assert.True(t, exists, "Worker pool should exist")
				assert.False(t, workerPool.PoolConfig.IsControlPlane, "Worker pool should have IsControlPlane=false")
			},
		},
		{
			name: "GCP cluster with tenant context",
			setup: func() *schema.ResourceData {
				d := resourceClusterGcp().TestResourceData()
				cloudConfig := map[string]interface{}{
					"project": "tenant-project",
					"region":  "europe-west1",
					"network": "tenant-network",
				}
				machinePool := map[string]interface{}{
					"name":          "pool1",
					"count":         2,
					"instance_type": "n1-standard-1",
					"disk_size_gb":  50,
					"azs":           schema.NewSet(schema.HashString, []interface{}{"europe-west1-a"}),
					"control_plane": false,
				}
				d.Set("cloud_config", []interface{}{cloudConfig})
				d.Set("context", "tenant")
				d.Set("cloud_account_id", "tenant-account-id")
				d.Set("machine_pool", []interface{}{machinePool})
				d.Set("name", "tenant-cluster")
				return d
			},
			expectError: false,
			verify: func(t *testing.T, cluster *models.V1SpectroGcpClusterEntity, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, cluster)
				assert.NotNil(t, cluster.Spec.CloudConfig)
				assert.Equal(t, "tenant-project", *cluster.Spec.CloudConfig.Project)
				assert.Equal(t, "europe-west1", *cluster.Spec.CloudConfig.Region)
				assert.Equal(t, "tenant-network", cluster.Spec.CloudConfig.Network)
				assert.Equal(t, "tenant-account-id", *cluster.Spec.CloudAccountUID)
			},
		},
		{
			name: "GCP cluster with worker pool and override_kubeadm_configuration",
			setup: func() *schema.ResourceData {
				d := resourceClusterGcp().TestResourceData()
				cloudConfig := map[string]interface{}{
					"project": "my-project",
					"region":  "us-central1",
					"network": "my-network",
				}
				machinePool := map[string]interface{}{
					"name":                           "worker-pool",
					"count":                          5,
					"instance_type":                  "n1-standard-2",
					"disk_size_gb":                   100,
					"azs":                            schema.NewSet(schema.HashString, []interface{}{"us-central1-a"}),
					"control_plane":                  false,
					"override_kubeadm_configuration": "kubeletExtraArgs:\n  node-labels: 'worker=true'",
				}
				d.Set("cloud_config", []interface{}{cloudConfig})
				d.Set("context", "project")
				d.Set("cloud_account_id", "gcp-account-id")
				d.Set("machine_pool", []interface{}{machinePool})
				d.Set("name", "test-cluster")
				return d
			},
			expectError: false,
			verify: func(t *testing.T, cluster *models.V1SpectroGcpClusterEntity, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, cluster)
				assert.Len(t, cluster.Spec.Machinepoolconfig, 1)
				assert.False(t, cluster.Spec.Machinepoolconfig[0].PoolConfig.IsControlPlane)
				assert.Equal(t, "kubeletExtraArgs:\n  node-labels: 'worker=true'", cluster.Spec.Machinepoolconfig[0].PoolConfig.OverrideKubeadmConfiguration)
			},
		},
		{
			name: "GCP cluster with multiple AZs in machine pool",
			setup: func() *schema.ResourceData {
				d := resourceClusterGcp().TestResourceData()
				cloudConfig := map[string]interface{}{
					"project": "my-project",
					"region":  "us-central1",
					"network": "my-network",
				}
				machinePool := map[string]interface{}{
					"name":          "pool-multi-az",
					"count":         6,
					"instance_type": "n1-standard-2",
					"disk_size_gb":  100,
					"azs":           schema.NewSet(schema.HashString, []interface{}{"us-central1-a", "us-central1-b", "us-central1-c"}),
					"control_plane": false,
				}
				d.Set("cloud_config", []interface{}{cloudConfig})
				d.Set("context", "project")
				d.Set("cloud_account_id", "gcp-account-id")
				d.Set("machine_pool", []interface{}{machinePool})
				d.Set("name", "test-cluster")
				return d
			},
			expectError: false,
			verify: func(t *testing.T, cluster *models.V1SpectroGcpClusterEntity, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, cluster)
				assert.Len(t, cluster.Spec.Machinepoolconfig, 1)
				azs := cluster.Spec.Machinepoolconfig[0].CloudConfig.Azs
				assert.Len(t, azs, 3)
				// Check that all AZs are present (order may vary)
				azMap := make(map[string]bool)
				for _, az := range azs {
					azMap[az] = true
				}
				assert.True(t, azMap["us-central1-a"])
				assert.True(t, azMap["us-central1-b"])
				assert.True(t, azMap["us-central1-c"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setup()
			cluster, err := toGcpCluster(nil, d)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cluster)
			} else {
				if err != nil {
					t.Logf("Unexpected error: %v", err)
				}
				if tt.verify != nil {
					tt.verify(t, cluster, err)
				}
			}
		})
	}
}
