package spectrocloud

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func TestFlattenEksLaunchTemplate(t *testing.T) {
	testCases := []struct {
		name     string
		input    *models.V1AwsLaunchTemplate
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: []interface{}{},
		},
		{
			name: "non-nil input",
			input: &models.V1AwsLaunchTemplate{
				Ami: &models.V1AwsAmiReference{
					ID: "ami-12345678",
				},
				RootVolume: &models.V1AwsRootVolume{
					Type:       "gp2",
					Iops:       100,
					Throughput: 125,
				},
				// add security group "sg-12345678"
				AdditionalSecurityGroups: []*models.V1AwsResourceReference{
					{
						ID: "sg-12345678",
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"ami_id":                 "ami-12345678",
					"root_volume_type":       "gp2",
					"root_volume_iops":       int64(100),
					"root_volume_throughput": int64(125),
					"additional_security_groups": []string{
						"sg-12345678",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := flattenEksLaunchTemplate(tc.input)
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Errorf("Unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFlattenMachinePoolConfigsEks(t *testing.T) {
	testCases := []struct {
		name     string
		input    []*models.V1EksMachinePoolConfig
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: []interface{}{},
		},
		{
			name:     "empty input",
			input:    []*models.V1EksMachinePoolConfig{},
			expected: []interface{}{},
		},
		{
			name: "non-empty input",
			input: []*models.V1EksMachinePoolConfig{
				{
					Name:           "pool1",
					Size:           3,
					MinSize:        1,
					MaxSize:        5,
					InstanceType:   "t2.micro",
					RootDeviceSize: 8,
					SubnetIds:      map[string]string{"us-west-2d": "subnet-87654321"},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":                "pool1",
					"additional_labels":   map[string]any{},
					"eks_launch_template": []any{},
					"count":               3,
					"min":                 1,
					"max":                 5,
					"instance_type":       "t2.micro",
					"disk_size_gb":        8,
					"az_subnets":          map[string]string{"us-west-2d": "subnet-87654321"},
					"update_strategy":     "RollingUpdateScaleOut",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := flattenMachinePoolConfigsEks(tc.input)
			if !cmp.Equal(result, tc.expected) {
				t.Errorf("Unexpected result (-want +got):\n%s", cmp.Diff(tc.expected, result))
			}
		})
	}
}

func TestFlattenClusterConfigsEKS(t *testing.T) {
	testCases := []struct {
		name     string
		input    *models.V1EksCloudConfig
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: []interface{}{},
		},
		{
			name: "non-empty input",
			input: &models.V1EksCloudConfig{
				Spec: &models.V1EksCloudConfigSpec{
					ClusterConfig: &models.V1EksClusterConfig{
						Region: ptr.To("us-west-2"),
						EndpointAccess: &models.V1EksClusterConfigEndpointAccess{
							PublicCIDRs: []string{"0.0.0.0/0"},
							Private:     true,
							Public:      true,
						},
						EncryptionConfig: &models.V1EncryptionConfig{
							IsEnabled: true,
							Provider:  "arn:aws:kms:us-west-2:123456789012:key/abcd1234-a123-456a-a12b-a123b4cd56ef",
						},
						VpcID:      "vpc-0abcd1234ef56789",
						SSHKeyName: "my-key-pair",
					},
					MachinePoolConfig: []*models.V1EksMachinePoolConfig{
						{
							Name:      "cp-pool",
							SubnetIds: map[string]string{"subnet-12345678": "subnet-87654321"},
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"region":                "us-west-2",
					"private_access_cidrs":  []string{},
					"public_access_cidrs":   []string{"0.0.0.0/0"},
					"az_subnets":            map[string]string{"subnet-12345678": "subnet-87654321"},
					"encryption_config_arn": "arn:aws:kms:us-west-2:123456789012:key/abcd1234-a123-456a-a12b-a123b4cd56ef",
					"endpoint_access":       "private_and_public",
					"vpc_id":                "vpc-0abcd1234ef56789",
					"ssh_key_name":          "my-key-pair",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := flattenClusterConfigsEKS(tc.input)
			if !cmp.Equal(result, tc.expected) {
				t.Errorf("Unexpected result (-want +got):\n%s", cmp.Diff(tc.expected, result))
			}
		})
	}
}

func TestFlattenClusterConfigsEKSPrivateCIDRS(t *testing.T) {
	testCases := []struct {
		name     string
		input    *models.V1EksCloudConfig
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: []interface{}{},
		},
		{
			name: "non-empty input",
			input: &models.V1EksCloudConfig{
				Spec: &models.V1EksCloudConfigSpec{
					ClusterConfig: &models.V1EksClusterConfig{
						Region: ptr.To("us-west-2"),
						EndpointAccess: &models.V1EksClusterConfigEndpointAccess{
							PrivateCIDRs: []string{"172.23.12.12/0"},
							Private:      true,
							Public:       false,
						},
						EncryptionConfig: &models.V1EncryptionConfig{
							IsEnabled: true,
							Provider:  "arn:aws:kms:us-west-2:123456789012:key/abcd1234-a123-456a-a12b-a123b4cd56ef",
						},
						VpcID:      "vpc-0abcd1234ef56789",
						SSHKeyName: "my-key-pair",
					},
					MachinePoolConfig: []*models.V1EksMachinePoolConfig{
						{
							Name:      "cp-pool",
							SubnetIds: map[string]string{"subnet-12345678": "subnet-87654321"},
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"region":                "us-west-2",
					"public_access_cidrs":   []string{},
					"private_access_cidrs":  []string{"172.23.12.12/0"},
					"az_subnets":            map[string]string{"subnet-12345678": "subnet-87654321"},
					"encryption_config_arn": "arn:aws:kms:us-west-2:123456789012:key/abcd1234-a123-456a-a12b-a123b4cd56ef",
					"endpoint_access":       "private",
					"vpc_id":                "vpc-0abcd1234ef56789",
					"ssh_key_name":          "my-key-pair",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := flattenClusterConfigsEKS(tc.input)
			if !cmp.Equal(result, tc.expected) {
				t.Errorf("Unexpected result (-want +got):\n%s", cmp.Diff(tc.expected, result))
			}
		})
	}
}
