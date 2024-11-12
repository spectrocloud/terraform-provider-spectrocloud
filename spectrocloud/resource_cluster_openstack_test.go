package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
)

// Utility function to create a *string from a string
func strPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}
func TestToOpenStackCluster(t *testing.T) {
	// Setup test data
	d := schema.TestResourceDataRaw(t, resourceClusterOpenStack().Schema, map[string]interface{}{
		"cloud_config": []interface{}{
			map[string]interface{}{
				"region":      "RegionOne",
				"ssh_key":     "default",
				"domain":      "default",
				"network_id":  "network-1",
				"project":     "my_project",
				"subnet_id":   "subnet-1",
				"subnet_cidr": "192.168.1.0/24",
				"dns_servers": []interface{}{"server1", "server2"},
			},
		},
		"context":          "default-context",
		"cloud_account_id": "cloud-account-id",
		"machine_pool": []interface{}{
			map[string]interface{}{
				"name":                 "worker",
				"flavor":               "m1.small",
				"control_plane":        false,
				"worker":               true,
				"desired_size":         2,
				"availability_zones":   []interface{}{"zone-1"},
				"subnet_ids":           []interface{}{"subnet-1"},
				"node_pools":           []interface{}{},
				"node_os_type":         "linux",
				"initial_node_count":   2,
				"auto_scaling_group":   false,
				"spot_instance":        false,
				"spot_max_price":       "0.0",
				"max_size":             5,
				"min_size":             2,
				"desired_capacity":     2,
				"force_delete":         false,
				"on_demand_percentage": 100,
			},
		},
	})

	// Mock client
	c := &client.V1Client{}

	// Call the function
	cluster, err := toOpenStackCluster(c, d)

	// Check for unexpected error
	assert.NoError(t, err)

	// Define expected output
	expected := &models.V1SpectroOpenStackClusterEntity{
		Metadata: &models.V1ObjectMeta{
			Annotations: map[string]string{
				"description": "",
			},
		},
		Spec: &models.V1SpectroOpenStackClusterEntitySpec{
			CloudAccountUID: strPtr("cloud-account-id"),
			Profiles:        []*models.V1SpectroClusterProfileEntity{},
			Policies:        &models.V1SpectroClusterPolicies{},
			CloudConfig: &models.V1OpenStackClusterConfig{
				Region:     "RegionOne",
				SSHKeyName: "default",
				Domain: &models.V1OpenStackResource{
					ID:   "default",
					Name: "default",
				},
				//Domain: &models.V1OpenStackResource{},
				Network: &models.V1OpenStackResource{
					ID: "network-1",
				},
				//Network: &models.V1OpenStackResource{},
				Project: &models.V1OpenStackResource{
					Name: "my_project",
				},
				//Project: &models.V1OpenStackResource{},
				Subnet: &models.V1OpenStackResource{
					ID: "subnet-1",
				},
				//Subnet: &models.V1OpenStackResource{},
				NodeCidr: "192.168.1.0/24",
				DNSNameservers: []string{
					"server2",
					"server1",
				},
				//DNSNameservers: []string{},
			},
			Machinepoolconfig: []*models.V1OpenStackMachinePoolConfigEntity{
				{
					CloudConfig: &models.V1OpenStackMachinePoolCloudConfigEntity{
						Azs:     []string{},
						DiskGiB: 0,
						FlavorConfig: &models.V1OpenstackFlavorConfig{
							DiskGiB:   0,
							MemoryMiB: 0,
							Name:      strPtr(""),
							NumCPUs:   0,
						},
						Subnet: &models.V1OpenStackResource{
							ID:   "",
							Name: "",
						},
					},
					PoolConfig: &models.V1MachinePoolConfigEntity{
						AdditionalLabels: map[string]string{},
						//AdditionalTags:        map[string]string{},
						IsControlPlane: false,
						Labels:         []string{"worker"},
						//MachinePoolProperties: &models.V1MachinePoolProperties{},
						MaxSize:            0,
						MinSize:            0,
						Size:               int32Ptr(0),
						Name:               strPtr("worker"),
						NodeRepaveInterval: 0,
						Taints:             []*models.V1Taint{},
						UpdateStrategy: &models.V1UpdateStrategy{
							Type: "RollingUpdateScaleOut",
						},
						UseControlPlaneAsWorker: false,
					},
				},
			},
		},
	}

	// Compare the expected and actual output using assertions
	assert.Equal(t, expected.Metadata.Annotations, cluster.Metadata.Annotations)
	assert.Equal(t, expected.Spec.CloudAccountUID, cluster.Spec.CloudAccountUID)
	assert.Equal(t, expected.Spec.CloudConfig, cluster.Spec.CloudConfig)
	assert.Equal(t, expected.Spec.Machinepoolconfig, cluster.Spec.Machinepoolconfig)
}

func TestToMachinePoolOpenStack(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		expected  *models.V1OpenStackMachinePoolConfigEntity
		expectErr bool
	}{
		{
			name: "Normal Case",
			input: map[string]interface{}{
				"control_plane":           true,
				"control_plane_as_worker": false,
				"azs":                     schema.NewSet(schema.HashString, []interface{}{"az1", "az2"}),
				"subnet_id":               "subnet-123",
				"instance_type":           "m4.large",
				"name":                    "control-plane",
				"count":                   3,
				"node_repave_interval":    0,
			},
			expected: &models.V1OpenStackMachinePoolConfigEntity{
				CloudConfig: &models.V1OpenStackMachinePoolCloudConfigEntity{
					Azs: []string{"az2", "az1"},
					Subnet: &models.V1OpenStackResource{
						ID: "subnet-123",
					},
					FlavorConfig: &models.V1OpenstackFlavorConfig{
						Name: ptr.To("m4.large"),
					},
				},
				PoolConfig: &models.V1MachinePoolConfigEntity{
					AdditionalLabels: map[string]string{},
					//Taints:           []*models.V1Taint{},
					IsControlPlane: true,
					Labels:         []string{"control-plane"},
					Name:           strPtr("control-plane"),
					Size:           int32Ptr(3),
					UpdateStrategy: &models.V1UpdateStrategy{
						Type: "RollingUpdateScaleOut",
					},
					UseControlPlaneAsWorker: false,
					NodeRepaveInterval:      0,
				},
			},
			expectErr: false,
		},
		{
			name: "Missing Optional Fields",
			input: map[string]interface{}{
				"control_plane":           false,
				"control_plane_as_worker": false,
				"azs":                     schema.NewSet(schema.HashString, []interface{}{"az1"}),
				"subnet_id":               "subnet-456",
				"instance_type":           "m4.large",
				"name":                    "worker",
				"count":                   2,
			},
			expected: &models.V1OpenStackMachinePoolConfigEntity{
				CloudConfig: &models.V1OpenStackMachinePoolCloudConfigEntity{
					Azs: []string{"az1"},
					Subnet: &models.V1OpenStackResource{
						ID: "subnet-456",
					},
					FlavorConfig: &models.V1OpenstackFlavorConfig{
						Name: ptr.To("m4.large"),
					},
				},
				PoolConfig: &models.V1MachinePoolConfigEntity{
					AdditionalLabels: map[string]string{},
					//Taints:           []*models.V1Taint{},
					IsControlPlane: false,
					Labels:         []string{"worker"},
					Name:           strPtr("worker"),
					Size:           int32Ptr(2),
					UpdateStrategy: &models.V1UpdateStrategy{
						Type: "RollingUpdateScaleOut",
					},
					UseControlPlaneAsWorker: false,
					NodeRepaveInterval:      0,
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toMachinePoolOpenStack(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFlattenMachinePoolConfigsOpenStack(t *testing.T) {
	testCases := []struct {
		name     string
		input    []*models.V1OpenStackMachinePoolConfig
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: []interface{}{},
		},
		{
			name:     "empty input",
			input:    []*models.V1OpenStackMachinePoolConfig{},
			expected: []interface{}{},
		},
		{
			name: "non-empty input",
			input: []*models.V1OpenStackMachinePoolConfig{
				{
					Name:                    "pool1",
					IsControlPlane:          true,
					UseControlPlaneAsWorker: false,
					Size:                    3,
					Subnet: &models.V1OpenStackResource{
						ID: "subnet-12345",
					},
					Azs: []string{"az1", "az2"},
					FlavorConfig: &models.V1OpenstackFlavorConfig{
						Name: strPtr("m1.medium"),
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":                    "pool1",
					"control_plane":           true,
					"control_plane_as_worker": false,
					"count":                   3,
					"subnet_id":               "subnet-12345",
					"azs":                     []string{"az1", "az2"},
					"instance_type":           strPtr("m1.medium"),
					"additional_labels":       map[string]interface{}{},
					"update_strategy":         "RollingUpdateScaleOut",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := flattenMachinePoolConfigsOpenStack(tc.input)
			if !cmp.Equal(result, tc.expected) {
				t.Errorf("Unexpected result for %s (-want +got):\n%s", tc.name, cmp.Diff(tc.expected, result))
			}
		})
	}
}
