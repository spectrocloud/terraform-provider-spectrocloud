package spectrocloud

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func TestFlattenMachinePoolConfigsTke(t *testing.T) {
	testCases := []struct {
		name     string
		input    []*models.V1TencentMachinePoolConfig
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: []interface{}{},
		},
		{
			name:     "empty input",
			input:    []*models.V1TencentMachinePoolConfig{},
			expected: []interface{}{},
		},
		{
			name: "non-empty input without control plane",
			input: []*models.V1TencentMachinePoolConfig{
				{
					Name:           "pool1",
					IsControlPlane: false,
					Size:           3,
					MinSize:        1,
					MaxSize:        5,
					InstanceType:   "m1.medium",
					RootDeviceSize: 8,
					Azs:            []string{"us-west-2a", "us-west-2b"},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":              "pool1",
					"count":             3,
					"min":               1,
					"max":               5,
					"instance_type":     "m1.medium",
					"disk_size_gb":      8,
					"azs":               []string{"us-west-2a", "us-west-2b"},
					"update_strategy":   "RollingUpdateScaleOut",
					"additional_labels": map[string]any{},
				},
			},
		},
		{
			name: "non-empty input with control plane",
			input: []*models.V1TencentMachinePoolConfig{
				{
					Name:           "pool1",
					IsControlPlane: true, // This should be excluded
					Size:           3,
					MinSize:        1,
					MaxSize:        5,
					InstanceType:   "m1.medium",
					RootDeviceSize: 8,
					Azs:            []string{"us-west-2a", "us-west-2b"},
				},
				{
					Name:           "pool2",
					IsControlPlane: false,
					Size:           2,
					MinSize:        1,
					MaxSize:        4,
					InstanceType:   "m2.large",
					RootDeviceSize: 10,
					Azs:            []string{"us-west-2c"},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":              "pool2",
					"count":             2,
					"min":               1,
					"max":               4,
					"instance_type":     "m2.large",
					"disk_size_gb":      10,
					"azs":               []string{"us-west-2c"},
					"update_strategy":   "RollingUpdateScaleOut",
					"additional_labels": map[string]any{},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := flattenMachinePoolConfigsTke(tc.input)
			if !cmp.Equal(result, tc.expected) {
				t.Errorf("Unexpected result (-want +got):\n%s", cmp.Diff(tc.expected, result))
			}
		})
	}
}

func TestToMachinePoolTke(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1TencentMachinePoolConfigEntity
	}{
		{
			name: "valid input - worker pool",
			input: map[string]interface{}{
				"name":          "worker-pool",
				"count":         3,
				"min":           1,
				"max":           5,
				"instance_type": "m1.medium",
				"disk_size_gb":  100,
				"az_subnets": map[string]interface{}{
					"us-west-2":  "subnet-123456",
					"us-west-22": "subnet-654321",
				},
				"control_plane": false,
				"taints":        []interface{}{},
			},
			expected: &models.V1TencentMachinePoolConfigEntity{
				CloudConfig: &models.V1TencentMachinePoolCloudConfigEntity{
					RootDeviceSize: 100,
					InstanceType:   "m1.medium",
					Azs:            []string{"us-west-2", "us-west-22"},
				},
				PoolConfig: &models.V1MachinePoolConfigEntity{
					Labels:         []string{"worker"},
					Name:           ptr.To("worker-pool"),
					Size:           ptr.To(int32(3)),
					MinSize:        1,
					MaxSize:        5,
					IsControlPlane: false,
					UpdateStrategy: &models.V1UpdateStrategy{
						Type: "RollingUpdateScaleOut",
					},
					Taints:           []*models.V1Taint{}, // Expected taints if any
					AdditionalLabels: map[string]string{},
				},
			},
		},
		{
			name: "valid input - control plane pool",
			input: map[string]interface{}{
				"name":          "control-plane-pool",
				"count":         3,
				"instance_type": "m1.large",
				"disk_size_gb":  150,
				"az_subnets": map[string]interface{}{
					"us-west-1a": "subnet-123456",
				},
				"control_plane": true,
				"taints":        []interface{}{},
			},
			expected: &models.V1TencentMachinePoolConfigEntity{
				CloudConfig: &models.V1TencentMachinePoolCloudConfigEntity{
					RootDeviceSize: 150,
					InstanceType:   "m1.large",
					Azs:            []string{"us-west-1a"},
				},
				PoolConfig: &models.V1MachinePoolConfigEntity{
					Labels:         []string{"control-plane"},
					Name:           ptr.To("control-plane-pool"),
					Size:           ptr.To(int32(3)),
					MinSize:        3,
					MaxSize:        3,
					IsControlPlane: true,
					UpdateStrategy: &models.V1UpdateStrategy{
						Type: "RollingUpdateScaleOut",
					},
					Taints:           []*models.V1Taint{}, // Expected taints if any
					AdditionalLabels: map[string]string{},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function with the test input
			result := toMachinePoolTke(tc.input)

			// Compare the actual output with the expected output
			assert.Equal(t, tc.expected, result, "Unexpected result in test case: %s", tc.name)
		})
	}
}
