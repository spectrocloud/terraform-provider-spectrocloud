package spectrocloud

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
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
			result := setEksLaunchTemplate(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Unexpected result (-want +got):\n%s", cmp.Diff(tc.expected, result))
			}
		})
	}
}

func TestToEksCluster(t *testing.T) {
	// Setup a dummy ResourceData for testing
	d := resourceClusterEks().TestResourceData()

	/* set the values of the ResourceData */
	d.Set("name", "test-cluster")
	d.Set("context", "project")
	d.Set("tags", []interface{}{"tag1:value1", "tag2:value2"})
	d.Set("cloud_account_id", "test-cloud-id")
	d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"ssh_key_name":          "test-ssh-key",
			"region":                "us-west-1",
			"vpc_id":                "test-vpc-id",
			"endpoint_access":       "public",
			"public_access_cidrs":   []interface{}{"0.0.0.0/0"},
			"encryption_config_arn": "arn:test:encryption",
		},
	})
	d.Set("machine_pool", []interface{}{
		map[string]interface{}{
			"name":            "test-pool",
			"disk_size_gb":    10,
			"count":           2,
			"instance_type":   "t2.micro",
			"capacity_type":   "on-demand",
			"update_strategy": "RollingUpdateScaleOut",
		},
	})

	//client := &client.V1Client{}
	//
	//cluster, err := toEksCluster(client, d)
	//
	//assert.NoError(t, err, "Expected no error from toEksCluster")
	//assert.Equal(t, "test-cluster", cluster.Metadata.Name, "Unexpected cluster name")
	//
	//assert.NotNil(t, cluster.Spec.Machinepoolconfig, "Expected MachinePools to be non-nil")
	//assert.Equal(t, 2, len(cluster.Spec.Machinepoolconfig), "Expected one machine pool in the cluster")
	//
	//assert.Equal(t, "test-pool", *cluster.Spec.Machinepoolconfig[1].PoolConfig.Name, "Unexpected machine pool name")
	//assert.Equal(t, int64(10), cluster.Spec.Machinepoolconfig[1].CloudConfig.RootDeviceSize, "Unexpected disk size")
}

func TestToMachinePoolEks(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected *models.V1EksMachinePoolConfigEntity
	}{
		{
			name: "Basic machine pool",
			input: map[string]interface{}{
				"name":            "test-pool",
				"disk_size_gb":    10,
				"instance_type":   "t2.micro",
				"update_strategy": "RollingUpdateScaleIn",
				"count":           2,
				"az_subnets": map[string]interface{}{
					"us-west-1a": "subnet-12345",
				},
			},
			expected: &models.V1EksMachinePoolConfigEntity{
				CloudConfig: &models.V1EksMachineCloudConfigEntity{
					RootDeviceSize: 10,
					InstanceType:   "t2.micro",
					CapacityType:   ptr.To("on-demand"),
					Azs:            []string{"us-west-1a"},
					Subnets: []*models.V1EksSubnetEntity{
						{
							Az: "us-west-1a",
							ID: "subnet-12345",
						},
					},
				},
				PoolConfig: &models.V1MachinePoolConfigEntity{
					IsControlPlane:   false,
					Labels:           []string{},
					Name:             ptr.To("test-pool"),
					Size:             ptr.To(int32(2)),
					MinSize:          2,
					MaxSize:          2,
					AdditionalLabels: map[string]string{},
					UpdateStrategy: &models.V1UpdateStrategy{
						Type: "RollingUpdateScaleIn",
					},
				},
			},
		},
		{
			name: "Spot instance machine pool with max price",
			input: map[string]interface{}{
				"name":          "test-pool-spot",
				"disk_size_gb":  10,
				"instance_type": "t2.micro",
				"count":         2,
				"capacity_type": "spot",
				"max_price":     "0.5",
				"az_subnets": map[string]interface{}{
					"us-west-1a": "subnet-12345",
				},
			},
			expected: &models.V1EksMachinePoolConfigEntity{
				CloudConfig: &models.V1EksMachineCloudConfigEntity{
					RootDeviceSize: 10,
					InstanceType:   "t2.micro",
					CapacityType:   ptr.To("spot"),
					Azs:            []string{"us-west-1a"},
					Subnets: []*models.V1EksSubnetEntity{
						{
							Az: "us-west-1a",
							ID: "subnet-12345",
						},
					},
					SpotMarketOptions: &models.V1SpotMarketOptions{
						MaxPrice: "0.5",
					},
				},
				PoolConfig: &models.V1MachinePoolConfigEntity{
					IsControlPlane:   false,
					Labels:           []string{},
					Name:             ptr.To("test-pool-spot"),
					Size:             ptr.To(int32(2)),
					MinSize:          2,
					MaxSize:          2,
					AdditionalLabels: map[string]string{},
					UpdateStrategy: &models.V1UpdateStrategy{
						Type: "RollingUpdateScaleOut",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toMachinePoolEks(tt.input)
			assert.Equal(t, tt.expected.CloudConfig.RootDeviceSize, got.CloudConfig.RootDeviceSize)
			assert.Equal(t, tt.expected.PoolConfig.AdditionalLabels, got.PoolConfig.AdditionalLabels)
			assert.Equal(t, tt.expected.PoolConfig.UpdateStrategy.Type, got.PoolConfig.UpdateStrategy.Type)
		})
	}
}
