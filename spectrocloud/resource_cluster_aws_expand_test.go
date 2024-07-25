package spectrocloud

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func TestToMachinePoolAws(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]interface{}
		vpcId    string
		expected *models.V1AwsMachinePoolConfigEntity
	}{
		{
			name: "Test 1: Basic test case",
			input: map[string]interface{}{
				"control_plane":              false,
				"control_plane_as_worker":    false,
				"name":                       "testPool",
				"count":                      3,
				"instance_type":              "t2.micro",
				"min":                        1,
				"max":                        5,
				"capacity_type":              "on-demand",
				"update_strategy":            "RollingUpdateScaleOut",
				"disk_size_gb":               65,
				"azs":                        schema.NewSet(schema.HashString, []interface{}{"us-west-1a", "us-west-1b"}),
				"additional_security_groups": schema.NewSet(schema.HashString, []interface{}{"sg-12345", "sg-67890"}),
			},
			vpcId: "vpc-12345",
			expected: &models.V1AwsMachinePoolConfigEntity{
				CloudConfig: &models.V1AwsMachinePoolCloudConfigEntity{
					Azs:            []string{"us-west-1a", "us-west-1b"},
					InstanceType:   types.Ptr("t2.micro"),
					CapacityType:   types.Ptr("on-demand"),
					RootDeviceSize: int64(65),
					Subnets:        []*models.V1AwsSubnetEntity{}, // assuming no az_subnets provided
					AdditionalSecurityGroups: []*models.V1AwsResourceReference{
						{
							ID: "sg-12345",
						},
						{
							ID: "sg-67890",
						},
					},
				},
				PoolConfig: &models.V1MachinePoolConfigEntity{
					AdditionalLabels: map[string]string{},
					IsControlPlane:   false,
					Labels:           []string{"worker"},
					Name:             types.Ptr("testPool"),
					Size:             types.Ptr(int32(3)),
					UpdateStrategy: &models.V1UpdateStrategy{
						Type: "RollingUpdateScaleOut",
					},
					MinSize:                 int32(1),
					MaxSize:                 int32(5),
					UseControlPlaneAsWorker: false,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _ := toMachinePoolAws(tc.input, tc.vpcId)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Unexpected result (-want +got):\n%s", cmp.Diff(tc.expected, result))
			}
		})
	}
}
