package spectrocloud

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func prepareVirtualClusterTestData() *schema.ResourceData {
	d := resourceClusterVirtual().TestResourceData()

	d.SetId("")
	d.Set("name", "virtual-picard-2")

	// Cluster Profile for Virtual Cluster
	cProfile := make([]map[string]interface{}, 0)
	cProfile = append(cProfile, map[string]interface{}{
		"id": "virtual-basic-infra-profile-id",
	})
	d.Set("cluster_profile", cProfile)
	d.Set("host_cluster_uid", "host-cluster-id")
	d.Set("cluster_group_uid", "group-cluster-id")

	// Cloud Config for Virtual Cluster
	cloudConfig := make([]map[string]interface{}, 0)
	vCloud := map[string]interface{}{
		"chart_name":    "virtual-chart-name",
		"chart_repo":    "virtual-chart-repo",
		"chart_version": "v1.0.0",
		"chart_values":  "default-values",
		"k8s_version":   "v1.20.0",
	}
	cloudConfig = append(cloudConfig, vCloud)
	d.Set("cloud_config", cloudConfig)

	return d
}

func TestToVirtualClusterResize(t *testing.T) {
	resources := map[string]interface{}{
		"max_cpu":           4,
		"max_mem_in_mb":     8192,
		"max_storage_in_gb": 100,
		"min_cpu":           2,
		"min_mem_in_mb":     4096,
		"min_storage_in_gb": 50,
	}

	expected := &models.V1VirtualClusterResize{
		InstanceType: &models.V1VirtualInstanceType{
			MaxCPU:        4,
			MaxMemInMiB:   8192,
			MaxStorageGiB: 100,
			MinCPU:        2,
			MinMemInMiB:   4096,
			MinStorageGiB: 50,
		},
	}

	result := toVirtualClusterResize(resources)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}

func TestToVirtualCluster(t *testing.T) {
	// Mock client
	mockClient := &client.V1Client{}

	// Define test cases
	testCases := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1SpectroVirtualClusterEntity
		err      error
	}{
		{
			name: "valid input with cloud config and resources",
			input: map[string]interface{}{
				"host_cluster_uid":  "host-cluster-uid-123",
				"cluster_group_uid": "cluster-group-uid-123",
				"context":           "project-context",
				"cloud_config": []interface{}{
					map[string]interface{}{
						"chart_name":    "test-chart",
						"chart_repo":    "test-repo",
						"chart_version": "1.0.0",
						"chart_values":  "test-values",
						"k8s_version":   "1.21.0",
					},
				},
			},
			expected: &models.V1SpectroVirtualClusterEntity{
				Metadata: &models.V1ObjectMeta{
					Name:        "", // Replace with expected values
					UID:         "", // Replace with expected values if applicable
					Labels:      map[string]string{},
					Annotations: map[string]string{"description": ""},
				},
				Spec: &models.V1SpectroVirtualClusterEntitySpec{
					CloudConfig: &models.V1VirtualClusterConfig{
						HelmRelease: &models.V1VirtualClusterHelmRelease{
							Chart: &models.V1VirtualClusterHelmChart{
								Name:    "test-chart",
								Repo:    "test-repo",
								Version: "1.0.0",
							},
							Values: "test-values",
						},
						KubernetesVersion: "1.21.0",
					},
					ClusterConfig: &models.V1ClusterConfigEntity{
						HostClusterConfig: &models.V1HostClusterConfig{
							ClusterGroup: &models.V1ObjectReference{
								UID: "cluster-group-uid-123",
							},
							HostCluster: &models.V1ObjectReference{
								UID: "host-cluster-uid-123",
							},
						},
						UpdateWorkerPoolsInParallel: true, // schema default
						Timezone:                    "",
					},
					Profiles:          []*models.V1SpectroClusterProfileEntity{}, // Adjust according to expected output of toProfiles
					Policies:          &models.V1SpectroClusterPolicies{},        // Adjust according to expected output of toPolicies
					Machinepoolconfig: []*models.V1VirtualMachinePoolConfigEntity{},
				},
			},
			err: nil,
		},
		{
			name: "missing cloud config",
			input: map[string]interface{}{
				"host_cluster_uid":  "host-cluster-uid-123",
				"cluster_group_uid": "cluster-group-uid-123",
				"context":           "project-context",
				"resources":         []interface{}{},
			},
			expected: &models.V1SpectroVirtualClusterEntity{
				Metadata: &models.V1ObjectMeta{
					Name:        "", // Replace with expected values
					UID:         "", // Replace with expected values if applicable
					Labels:      map[string]string{},
					Annotations: map[string]string{"description": ""},
				},
				Spec: &models.V1SpectroVirtualClusterEntitySpec{
					CloudConfig: &models.V1VirtualClusterConfig{
						HelmRelease: &models.V1VirtualClusterHelmRelease{
							Chart: &models.V1VirtualClusterHelmChart{
								Name:    "",
								Repo:    "",
								Version: "",
							},
							Values: "",
						},
						KubernetesVersion: "",
					},
					ClusterConfig: &models.V1ClusterConfigEntity{
						HostClusterConfig: &models.V1HostClusterConfig{
							ClusterGroup: &models.V1ObjectReference{
								UID: "cluster-group-uid-123",
							},
							HostCluster: &models.V1ObjectReference{
								UID: "host-cluster-uid-123",
							},
						},
						UpdateWorkerPoolsInParallel: true, // schema default
						Timezone:                    "",
					},
					Profiles:          []*models.V1SpectroClusterProfileEntity{}, // Adjust according to expected output of toProfiles
					Policies:          &models.V1SpectroClusterPolicies{},        // Adjust according to expected output of toPolicies
					Machinepoolconfig: []*models.V1VirtualMachinePoolConfigEntity{},
				},
			},
			err: nil,
		},
		// Add more test cases as necessary
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceClusterVirtual().Schema, tc.input) // Replace with correct schema
			result, err := toVirtualCluster(mockClient, d)

			if err != nil {
				assert.Equal(t, tc.err, err, "Unexpected error in test case: %s", tc.name)
			} else {
				assert.Equal(t, tc.expected, result, "Unexpected result in test case: %s", tc.name)
			}
		})
	}
}

func TestToMachinePoolVirtual(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1VirtualMachinePoolConfigEntity
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"max_cpu":           8,
				"max_mem_in_mb":     32768,
				"max_storage_in_gb": 500,
				"min_cpu":           2,
				"min_mem_in_mb":     8192,
				"min_storage_in_gb": 100,
			},
			expected: &models.V1VirtualMachinePoolConfigEntity{
				CloudConfig: &models.V1VirtualMachinePoolCloudConfigEntity{
					InstanceType: &models.V1VirtualInstanceType{
						MaxCPU:        int32(8),
						MaxMemInMiB:   int32(32768),
						MaxStorageGiB: int32(500),
						MinCPU:        int32(2),
						MinMemInMiB:   int32(8192),
						MinStorageGiB: int32(100),
					},
				},
			},
		},
		{
			name: "zero values input",
			input: map[string]interface{}{
				"max_cpu":           0,
				"max_mem_in_mb":     0,
				"max_storage_in_gb": 0,
				"min_cpu":           0,
				"min_mem_in_mb":     0,
				"min_storage_in_gb": 0,
			},
			expected: &models.V1VirtualMachinePoolConfigEntity{
				CloudConfig: &models.V1VirtualMachinePoolCloudConfigEntity{
					InstanceType: &models.V1VirtualInstanceType{
						MaxCPU:        int32(0),
						MaxMemInMiB:   int32(0),
						MaxStorageGiB: int32(0),
						MinCPU:        int32(0),
						MinMemInMiB:   int32(0),
						MinStorageGiB: int32(0),
					},
				},
			},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := toMachinePoolVirtual(tc.input)
			assert.Equal(t, tc.expected, result, "Unexpected result in test case: %s", tc.name)
		})
	}
}
