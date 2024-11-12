package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func TestToMachinePoolGcp(t *testing.T) {
	tests := []struct {
		name           string
		input          map[string]interface{}
		expectedOutput *models.V1GcpMachinePoolConfigEntity
		expectError    bool
	}{
		{
			name: "Control Plane",
			input: map[string]interface{}{
				"control_plane":           true,
				"control_plane_as_worker": true,
				"azs":                     schema.NewSet(schema.HashString, []interface{}{"us-central1-a"}),
				"instance_type":           "n1-standard-1",
				"disk_size_gb":            50,
				"name":                    "example-name",
				"count":                   3,
				"node_repave_interval":    0,
			},
			expectedOutput: &models.V1GcpMachinePoolConfigEntity{
				CloudConfig: &models.V1GcpMachinePoolCloudConfigEntity{
					Azs:            []string{"us-central1-a"},
					InstanceType:   ptr.To("n1-standard-1"),
					RootDeviceSize: int64(50),
				},
				PoolConfig: &models.V1MachinePoolConfigEntity{
					AdditionalLabels: map[string]string{},
					Taints:           nil,
					IsControlPlane:   true,
					Labels:           []string{"control-plane"},
					Name:             ptr.To("example-name"),
					Size:             ptr.To(int32(3)),
					UpdateStrategy: &models.V1UpdateStrategy{
						Type: "RollingUpdateScaleOut",
					},
					UseControlPlaneAsWorker: true,
				},
			},
			expectError: false,
		},
		{
			name: "Node Repave Interval Error",
			input: map[string]interface{}{
				"control_plane":           true,
				"control_plane_as_worker": false,
				"azs":                     schema.NewSet(schema.HashString, []interface{}{"us-central1-a"}),
				"instance_type":           "n1-standard-2",
				"disk_size_gb":            100,
				"name":                    "example-name-2",
				"count":                   2,
				"node_repave_interval":    -1,
			},
			expectedOutput: &models.V1GcpMachinePoolConfigEntity{
				CloudConfig: &models.V1GcpMachinePoolCloudConfigEntity{
					Azs:            []string{"us-central1-a"},
					InstanceType:   ptr.To("n1-standard-2"),
					RootDeviceSize: int64(100),
				},
				PoolConfig: &models.V1MachinePoolConfigEntity{
					AdditionalLabels: map[string]string{"example": "label"},
					Taints:           []*models.V1Taint{},
					IsControlPlane:   true,
					Labels:           []string{"control-plane"},
					Name:             ptr.To("example-name-2"),
					Size:             ptr.To(int32(2)),
					UpdateStrategy: &models.V1UpdateStrategy{
						Type: "RollingUpdate",
					},
					UseControlPlaneAsWorker: false,
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := toMachinePoolGcp(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestFlattenMachinePoolConfigsGcp(t *testing.T) {
	tests := []struct {
		name           string
		input          []*models.V1GcpMachinePoolConfig
		expectedOutput []interface{}
	}{
		{
			name: "Single Machine Pool",
			input: []*models.V1GcpMachinePoolConfig{
				{
					AdditionalLabels:        map[string]string{"label1": "value1", "label2": "value2"},
					Taints:                  []*models.V1Taint{{Key: "taint1", Value: "value1", Effect: "NoSchedule"}},
					IsControlPlane:          ptr.To(true),
					UseControlPlaneAsWorker: true,
					Name:                    "machine-pool-1",
					Size:                    int32(3),
					UpdateStrategy:          &models.V1UpdateStrategy{Type: "RollingUpdate"},
					InstanceType:            ptr.To("n1-standard-4"),
					RootDeviceSize:          int64(100),
					Azs:                     []string{"us-west1-a", "us-west1-b"},
					NodeRepaveInterval:      0,
				},
			},
			expectedOutput: []interface{}{
				map[string]interface{}{
					"additional_labels": map[string]string{
						"label1": "value1",
						"label2": "value2",
					},
					"taints": []interface{}{
						map[string]interface{}{
							"key":    "taint1",
							"value":  "value1",
							"effect": "NoSchedule",
						},
					},
					"control_plane":           true,
					"control_plane_as_worker": true,
					"name":                    "machine-pool-1",
					"count":                   3,
					"update_strategy":         "RollingUpdate",
					"instance_type":           "n1-standard-4",
					"disk_size_gb":            100,
					"azs":                     []string{"us-west1-a", "us-west1-b"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := flattenMachinePoolConfigsGcp(tt.input)
			assert.Equal(t, tt.expectedOutput, output)
		})
	}
}

func TestFlattenClusterConfigsGcp(t *testing.T) {
	tests := []struct {
		name           string
		input          *models.V1GcpCloudConfig
		expectedOutput []interface{}
	}{
		{
			name: "Valid Cloud Config",
			input: &models.V1GcpCloudConfig{
				Spec: &models.V1GcpCloudConfigSpec{
					ClusterConfig: &models.V1GcpClusterConfig{
						Project: ptr.To("my-project"),
						Network: "my-network",
						Region:  ptr.To("us-west1"),
					},
				},
			},
			expectedOutput: []interface{}{
				map[string]interface{}{
					"project": ptr.To("my-project"),
					"network": "my-network",
					"region":  "us-west1",
				},
			},
		},
		{
			name:           "Nil Cloud Config",
			input:          nil,
			expectedOutput: []interface{}{},
		},
		{
			name:           "Empty Cluster Config",
			input:          &models.V1GcpCloudConfig{},
			expectedOutput: []interface{}{},
		},
		{
			name:           "Empty Cluster Config Spec",
			input:          &models.V1GcpCloudConfig{Spec: &models.V1GcpCloudConfigSpec{}},
			expectedOutput: []interface{}{},
		},
		{
			name: "Missing Fields in Cluster Config",
			input: &models.V1GcpCloudConfig{
				Spec: &models.V1GcpCloudConfigSpec{
					ClusterConfig: &models.V1GcpClusterConfig{},
				},
			},
			expectedOutput: []interface{}{
				map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := flattenClusterConfigsGcp(tt.input)
			assert.Equal(t, tt.expectedOutput, output)
		})
	}
}
