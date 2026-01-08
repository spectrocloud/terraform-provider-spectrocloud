package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"github.com/stretchr/testify/assert"
)

func TestToMachinePoolAws(t *testing.T) {
	tests := []struct {
		name        string
		machinePool interface{}
		vpcId       string
		expected    *models.V1AwsMachinePoolConfigEntity
		expectedErr bool
	}{
		{
			name: "Control Plane Pool with Node Repave Interval",
			machinePool: map[string]interface{}{
				"name":                    "control-plane-pool",
				"count":                   3,
				"instance_type":           "t3.large",
				"disk_size_gb":            100,
				"control_plane":           true,
				"control_plane_as_worker": false,
				"azs":                     schema.NewSet(schema.HashString, []interface{}{"us-west-2a", "us-west-2b"}),
				"node_repave_interval":    0,
			},
			vpcId: "vpc-12345",
			expected: &models.V1AwsMachinePoolConfigEntity{
				CloudConfig: &models.V1AwsMachinePoolCloudConfigEntity{
					Azs:            []string{"us-west-2b", "us-west-2a"},
					InstanceType:   types.Ptr("t3.large"),
					CapacityType:   types.Ptr("on-demand"),
					RootDeviceSize: 100,
					Subnets:        []*models.V1AwsSubnetEntity{},
				},
				PoolConfig: &models.V1MachinePoolConfigEntity{
					Name:                  types.Ptr("control-plane-pool"),
					Size:                  types.Ptr(int32(3)),
					MinSize:               3,
					MaxSize:               3,
					IsControlPlane:        true,
					Labels:                []string{"control-plane"},
					UpdateStrategy:        &models.V1UpdateStrategy{Type: "RollingUpdateScaleOut"},
					AdditionalLabels:      map[string]string{},
					AdditionalAnnotations: map[string]string{},
				},
			},
			expectedErr: false,
		},
		{
			name: "Worker Pool with Spot Instances",
			machinePool: map[string]interface{}{
				"name":                    "worker-pool",
				"count":                   5,
				"instance_type":           "t3.medium",
				"disk_size_gb":            50,
				"control_plane":           false,
				"control_plane_as_worker": false,
				"az_subnets": map[string]interface{}{
					"us-west-1a": "subnet-1",
				},
				"azs":                  schema.NewSet(schema.HashString, []interface{}{"us-west-1a"}),
				"capacity_type":        "spot",
				"max_price":            "0.5",
				"node_repave_interval": 10,
			},
			vpcId: "vpc-67890",
			expected: &models.V1AwsMachinePoolConfigEntity{
				CloudConfig: &models.V1AwsMachinePoolCloudConfigEntity{
					Azs:            []string{"us-west-1a"},
					InstanceType:   types.Ptr("t3.medium"),
					CapacityType:   types.Ptr("spot"),
					RootDeviceSize: 50,
					Subnets: []*models.V1AwsSubnetEntity{
						{ID: "subnet-1", Az: "us-west-1a"},
					},
					SpotMarketOptions: &models.V1SpotMarketOptions{
						MaxPrice: "0.5",
					},
				},
				PoolConfig: &models.V1MachinePoolConfigEntity{
					Name:                  types.Ptr("worker-pool"),
					Size:                  types.Ptr(int32(5)),
					MinSize:               5,
					MaxSize:               5,
					IsControlPlane:        false,
					Labels:                []string{"worker"},
					NodeRepaveInterval:    10,
					UpdateStrategy:        &models.V1UpdateStrategy{Type: "RollingUpdateScaleOut"},
					AdditionalLabels:      map[string]string{},
					AdditionalAnnotations: map[string]string{},
				},
			},
			expectedErr: false,
		},
		{
			name: "Control Plane with Invalid Node Repave Interval",
			machinePool: map[string]interface{}{
				"name":                    "control-plane-invalid",
				"count":                   3,
				"instance_type":           "t3.large",
				"disk_size_gb":            100,
				"control_plane":           true,
				"control_plane_as_worker": false,
				"azs":                     schema.NewSet(schema.HashString, []interface{}{"us-west-2a"}),
				"node_repave_interval":    10, // Invalid for control plane
			},
			vpcId:       "vpc-12345",
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := toMachinePoolAws(tt.machinePool, tt.vpcId)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, output)
			}
		})
	}
}
