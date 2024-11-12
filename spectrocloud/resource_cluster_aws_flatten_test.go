package spectrocloud

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
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
					IsControlPlane:          ptr.To(true),
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

func TestFlattenClusterConfigsAws(t *testing.T) {
	tests := []struct {
		name     string
		input    *models.V1AwsCloudConfig
		expected []interface{}
	}{
		{
			name:     "Nil config",
			input:    nil,
			expected: []interface{}{},
		},
		{
			name: "Nil Spec",
			input: &models.V1AwsCloudConfig{
				Spec: nil,
			},
			expected: []interface{}{},
		},
		{
			name: "Nil ClusterConfig",
			input: &models.V1AwsCloudConfig{
				Spec: &models.V1AwsCloudConfigSpec{
					ClusterConfig: nil,
				},
			},
			expected: []interface{}{},
		},
		{
			name: "Empty ClusterConfig",
			input: &models.V1AwsCloudConfig{
				Spec: &models.V1AwsCloudConfigSpec{
					ClusterConfig: &models.V1AwsClusterConfig{},
				},
			},
			expected: []interface{}{
				map[string]interface{}{},
			},
		},
		{
			name: "Partial ClusterConfig",
			input: &models.V1AwsCloudConfig{
				Spec: &models.V1AwsCloudConfigSpec{
					ClusterConfig: &models.V1AwsClusterConfig{
						SSHKeyName: "my-ssh-key",
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"ssh_key_name": "my-ssh-key",
				},
			},
		},
		{
			name: "Complete ClusterConfig",
			input: &models.V1AwsCloudConfig{
				Spec: &models.V1AwsCloudConfigSpec{
					ClusterConfig: &models.V1AwsClusterConfig{
						SSHKeyName:               "my-ssh-key",
						Region:                   ptr.To("us-west-2"),
						VpcID:                    "vpc-12345",
						ControlPlaneLoadBalancer: "lb-12345",
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"ssh_key_name":     "my-ssh-key",
					"region":           "us-west-2",
					"vpc_id":           "vpc-12345",
					"control_plane_lb": "lb-12345",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := flattenClusterConfigsAws(tt.input)
			assert.Equal(t, tt.expected, output)
		})
	}
}
