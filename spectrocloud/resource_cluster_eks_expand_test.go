package spectrocloud

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spectrocloud/hapi/models"
)

func TestSetAwsLaunchTemplate(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1AwsLaunchTemplate
	}{
		{
			name:     "nil eks_launch_template",
			input:    map[string]interface{}{},
			expected: nil,
		},
		{
			name: "empty eks_launch_template list",
			input: map[string]interface{}{
				"eks_launch_template": []interface{}{},
			},
			expected: nil,
		},
		{
			name: "non-empty eks_launch_template list",
			input: map[string]interface{}{
				"eks_launch_template": []interface{}{
					map[string]interface{}{
						"ami_id":                 "ami-12345678",
						"root_volume_type":       "gp2",
						"root_volume_iops":       100,
						"root_volume_throughput": 125,
					},
				},
			},
			expected: &models.V1AwsLaunchTemplate{
				Ami: &models.V1AwsAmiReference{
					ID: "ami-12345678",
				},
				RootVolume: &models.V1AwsRootVolume{
					Type:       "gp2",
					Iops:       100,
					Throughput: 125,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := setAwsLaunchTemplate(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Unexpected result (-want +got):\n%s", cmp.Diff(tc.expected, result))
			}
		})
	}
}
