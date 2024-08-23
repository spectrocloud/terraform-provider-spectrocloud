package spectrocloud

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func TestFlattenMachinePoolConfigsAws(t *testing.T) {
	testCases := []struct {
		name     string
		input    []*models.V1AwsMachinePoolConfig
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: []interface{}{},
		},
		{
			name:     "empty input",
			input:    []*models.V1AwsMachinePoolConfig{},
			expected: []interface{}{},
		},
		{
			name: "non-empty input",
			input: []*models.V1AwsMachinePoolConfig{
				{
					Name:                    "pool1",
					IsControlPlane:          types.Ptr(true),
					UseControlPlaneAsWorker: false,
					Size:                    3,
					MinSize:                 1,
					MaxSize:                 5,
					InstanceType:            "t2.micro",
					RootDeviceSize:          8,
					SubnetIds:               map[string]string{"us-west-2d": "subnet-87654321"},
					AdditionalSecurityGroups: []*models.V1AwsResourceReference{
						{
							ID: "sg-1234567890",
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":                       "pool1",
					"control_plane":              true,
					"control_plane_as_worker":    false,
					"additional_labels":          map[string]any{},
					"count":                      3,
					"min":                        1,
					"max":                        5,
					"instance_type":              "t2.micro",
					"disk_size_gb":               8,
					"az_subnets":                 map[string]string{"us-west-2d": "subnet-87654321"},
					"update_strategy":            "RollingUpdateScaleOut",
					"additional_security_groups": []string{"sg-1234567890"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := flattenMachinePoolConfigsAws(tc.input)
			if !cmp.Equal(result, tc.expected) {
				t.Errorf("Unexpected result (-want +got):\n%s", cmp.Diff(tc.expected, result))
			}
		})
	}
}
