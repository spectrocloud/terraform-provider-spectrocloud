package spectrocloud

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
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

func TestIsKarpenterManagedPool(t *testing.T) {
	testCases := []struct {
		name     string
		input    *models.V1EksMachinePoolConfig
		expected bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: false,
		},
		{
			name: "nil labels",
			input: &models.V1EksMachinePoolConfig{
				Name:   "pool1",
				Labels: nil,
			},
			expected: false,
		},
		{
			name: "empty labels",
			input: &models.V1EksMachinePoolConfig{
				Name:   "pool1",
				Labels: []string{},
			},
			expected: false,
		},
		{
			name: "worker label only",
			input: &models.V1EksMachinePoolConfig{
				Name:   "pool1",
				Labels: []string{"worker"},
			},
			expected: false,
		},
		{
			name: "Karpenter-managed pool",
			input: &models.V1EksMachinePoolConfig{
				Name:   "karpenter-pool",
				Labels: []string{"worker", "spectrocloud.com/managed-by:karpenter"},
			},
			expected: true,
		},
		{
			name: "Karpenter-managed pool with only Karpenter label",
			input: &models.V1EksMachinePoolConfig{
				Name:   "karpenter-pool",
				Labels: []string{"spectrocloud.com/managed-by:karpenter"},
			},
			expected: true,
		},
		{
			name: "control-plane label with Karpenter",
			input: &models.V1EksMachinePoolConfig{
				Name:   "karpenter-cp-pool",
				Labels: []string{"control-plane", "spectrocloud.com/managed-by:karpenter"},
			},
			expected: true,
		},
		{
			name: "similar label but not Karpenter",
			input: &models.V1EksMachinePoolConfig{
				Name:   "pool1",
				Labels: []string{"worker", "spectrocloud.com/managed-by:terraform"},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isKarpenterManagedPool(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
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
					AmiType:        "AL2_x86_64",
					SubnetIds:      map[string]string{"us-west-2d": "subnet-87654321"},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":                   "pool1",
					"additional_labels":      map[string]any{},
					"additional_annotations": map[string]interface{}{},
					"ami_type":               "AL2_x86_64",
					"eks_launch_template":    []any{},
					"count":                  3,
					"min":                    1,
					"max":                    5,
					"instance_type":          "t2.micro",
					"disk_size_gb":           8,
					"az_subnets":             map[string]string{"us-west-2d": "subnet-87654321"},
					"update_strategy":        "RollingUpdateScaleOut",
				},
			},
		},
		{
			name: "skip Karpenter-managed pool",
			input: []*models.V1EksMachinePoolConfig{
				{
					Name:           "pool1",
					Size:           3,
					MinSize:        1,
					MaxSize:        5,
					InstanceType:   "t2.micro",
					RootDeviceSize: 8,
					AmiType:        "AL2_x86_64",
					SubnetIds:      map[string]string{"us-west-2d": "subnet-87654321"},
					Labels:         []string{"worker"},
				},
				{
					Name:           "karpenter-pool",
					Size:           2,
					MinSize:        1,
					MaxSize:        10,
					InstanceType:   "t3.medium",
					RootDeviceSize: 20,
					AmiType:        "AL2023_x86_64_STANDARD",
					Labels:         []string{"worker", "spectrocloud.com/managed-by:karpenter"},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":                   "pool1",
					"additional_labels":      map[string]any{},
					"additional_annotations": map[string]interface{}{},
					"ami_type":               "AL2_x86_64",
					"eks_launch_template":    []any{},
					"count":                  3,
					"min":                    1,
					"max":                    5,
					"instance_type":          "t2.micro",
					"disk_size_gb":           8,
					"az_subnets":             map[string]string{"us-west-2d": "subnet-87654321"},
					"update_strategy":        "RollingUpdateScaleOut",
				},
			},
		},
		{
			name: "skip control plane pool",
			input: []*models.V1EksMachinePoolConfig{
				{
					Name:           "pool1",
					Size:           3,
					MinSize:        1,
					MaxSize:        5,
					InstanceType:   "t2.micro",
					RootDeviceSize: 8,
					AmiType:        "AL2_x86_64",
					SubnetIds:      map[string]string{"us-west-2d": "subnet-87654321"},
					IsControlPlane: types.Ptr(true),
				},
			},
			expected: []interface{}{},
		},
		{
			name: "skip both control plane and Karpenter pools",
			input: []*models.V1EksMachinePoolConfig{
				{
					Name:           "cp-pool",
					Size:           3,
					IsControlPlane: types.Ptr(true),
				},
				{
					Name:   "karpenter-pool",
					Size:   2,
					Labels: []string{"worker", "spectrocloud.com/managed-by:karpenter"},
				},
				{
					Name:           "regular-pool",
					Size:           5,
					MinSize:        2,
					MaxSize:        10,
					InstanceType:   "t3.large",
					RootDeviceSize: 50,
					AmiType:        "AL2023_x86_64_STANDARD",
					Labels:         []string{"worker"},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":                   "regular-pool",
					"additional_labels":      map[string]any{},
					"additional_annotations": map[string]interface{}{},
					"ami_type":               "AL2023_x86_64_STANDARD",
					"eks_launch_template":    []any{},
					"count":                  5,
					"min":                    2,
					"max":                    10,
					"instance_type":          "t3.large",
					"disk_size_gb":           50,
					"azs":                    []string(nil),
					"update_strategy":        "RollingUpdateScaleOut",
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
						Region: types.Ptr("us-west-2"),
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
						Region: types.Ptr("us-west-2"),
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

func TestFlattenFargateProfilesEks(t *testing.T) {
	testCases := []struct {
		name     string
		input    []*models.V1FargateProfile
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: []interface{}{},
		},
		{
			name:     "empty input",
			input:    []*models.V1FargateProfile{},
			expected: []interface{}{},
		},
		{
			name: "single fargate profile with all fields",
			input: []*models.V1FargateProfile{
				{
					Name:           types.Ptr("fargate-profile-1"),
					SubnetIds:      []string{"subnet-12345", "subnet-67890"},
					AdditionalTags: map[string]string{"Environment": "production", "Team": "platform"},
					Selectors: []*models.V1FargateSelector{
						{
							Namespace: types.Ptr("default"),
							Labels:    map[string]string{"app": "nginx", "version": "1.0"},
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":            types.Ptr("fargate-profile-1"),
					"subnets":         []string{"subnet-12345", "subnet-67890"},
					"additional_tags": map[string]string{"Environment": "production", "Team": "platform"},
					"selector": []interface{}{
						map[string]interface{}{
							"namespace": types.Ptr("default"),
							"labels":    map[string]string{"app": "nginx", "version": "1.0"},
						},
					},
				},
			},
		},
		{
			name: "fargate profile with multiple selectors",
			input: []*models.V1FargateProfile{
				{
					Name:      types.Ptr("fargate-profile-2"),
					SubnetIds: []string{"subnet-11111"},
					Selectors: []*models.V1FargateSelector{
						{
							Namespace: types.Ptr("kube-system"),
							Labels:    map[string]string{"k8s-app": "kube-dns"},
						},
						{
							Namespace: types.Ptr("default"),
							Labels:    map[string]string{"app": "web"},
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":            types.Ptr("fargate-profile-2"),
					"subnets":         []string{"subnet-11111"},
					"additional_tags": map[string]string(nil),
					"selector": []interface{}{
						map[string]interface{}{
							"namespace": types.Ptr("kube-system"),
							"labels":    map[string]string{"k8s-app": "kube-dns"},
						},
						map[string]interface{}{
							"namespace": types.Ptr("default"),
							"labels":    map[string]string{"app": "web"},
						},
					},
				},
			},
		},
		{
			name: "fargate profile with nil selectors",
			input: []*models.V1FargateProfile{
				{
					Name:           types.Ptr("fargate-profile-4"),
					SubnetIds:      []string{"subnet-33333"},
					AdditionalTags: map[string]string{},
					Selectors:      nil,
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":            types.Ptr("fargate-profile-4"),
					"subnets":         []string{"subnet-33333"},
					"additional_tags": map[string]string{},
					"selector":        []interface{}{},
				},
			},
		},
		{
			name: "fargate profile with empty subnets",
			input: []*models.V1FargateProfile{
				{
					Name:           types.Ptr("fargate-profile-5"),
					SubnetIds:      []string{},
					AdditionalTags: map[string]string{"CostCenter": "engineering"},
					Selectors: []*models.V1FargateSelector{
						{
							Namespace: types.Ptr("production"),
							Labels:    map[string]string{"env": "prod"},
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":            types.Ptr("fargate-profile-5"),
					"subnets":         []string{},
					"additional_tags": map[string]string{"CostCenter": "engineering"},
					"selector": []interface{}{
						map[string]interface{}{
							"namespace": types.Ptr("production"),
							"labels":    map[string]string{"env": "prod"},
						},
					},
				},
			},
		},
		{
			name: "multiple fargate profiles",
			input: []*models.V1FargateProfile{
				{
					Name:      types.Ptr("fargate-profile-7"),
					SubnetIds: []string{"subnet-55555"},
					Selectors: []*models.V1FargateSelector{
						{
							Namespace: types.Ptr("app1"),
							Labels:    map[string]string{"app": "app1"},
						},
					},
				},
				{
					Name:      types.Ptr("fargate-profile-8"),
					SubnetIds: []string{"subnet-66666"},
					Selectors: []*models.V1FargateSelector{
						{
							Namespace: types.Ptr("app2"),
							Labels:    map[string]string{"app": "app2"},
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":            types.Ptr("fargate-profile-7"),
					"subnets":         []string{"subnet-55555"},
					"additional_tags": map[string]string(nil),
					"selector": []interface{}{
						map[string]interface{}{
							"namespace": types.Ptr("app1"),
							"labels":    map[string]string{"app": "app1"},
						},
					},
				},
				map[string]interface{}{
					"name":            types.Ptr("fargate-profile-8"),
					"subnets":         []string{"subnet-66666"},
					"additional_tags": map[string]string(nil),
					"selector": []interface{}{
						map[string]interface{}{
							"namespace": types.Ptr("app2"),
							"labels":    map[string]string{"app": "app2"},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := flattenFargateProfilesEks(tc.input)
			if !cmp.Equal(result, tc.expected) {
				t.Errorf("Unexpected result (-want +got):\n%s", cmp.Diff(tc.expected, result))
			}
		})
	}
}
