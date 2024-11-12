package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func TestToMachinePoolGke(t *testing.T) {
	// Simulate input data
	machinePool := map[string]interface{}{
		"name":          "pool1",
		"count":         3,
		"instance_type": "n1-standard-2",
		"disk_size_gb":  100,
	}
	mp, err := toMachinePoolGke(machinePool)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, mp)

	// Check the CloudConfig fields
	assert.NotNil(t, mp.CloudConfig)
	assert.Equal(t, "n1-standard-2", *mp.CloudConfig.InstanceType)
	assert.Equal(t, int64(100), mp.CloudConfig.RootDeviceSize)

	// Check the PoolConfig fields
	assert.NotNil(t, mp.PoolConfig)
	assert.Equal(t, "pool1", *mp.PoolConfig.Name)
	assert.Equal(t, int32(3), *mp.PoolConfig.Size)
	assert.Equal(t, []string{"worker"}, mp.PoolConfig.Labels)
}

func TestToGkeCluster(t *testing.T) {
	// Simulate input data
	cloudConfig := map[string]interface{}{
		"project": "my-project",
		"region":  "us-central1",
	}
	machinePool := map[string]interface{}{
		"name":          "pool1",
		"count":         3,
		"instance_type": "n1-standard-2",
		"disk_size_gb":  100,
	}
	d := resourceClusterGke().TestResourceData()
	d.Set("cloud_config", []interface{}{cloudConfig})
	d.Set("context", "cluster-context")
	d.Set("cloud_account_id", "cloud-account-id")
	d.Set("machine_pool", []interface{}{machinePool})

	// Call the toGkeCluster function with the simulated input data
	cluster, err := toGkeCluster(nil, d)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, cluster)

	// Check the Metadata
	assert.NotNil(t, cluster.Metadata)
	// Check other fields similarly
	assert.NotNil(t, cluster.Spec.CloudConfig)
	assert.Equal(t, "my-project", *cluster.Spec.CloudConfig.Project)
	assert.Equal(t, "us-central1", *cluster.Spec.CloudConfig.Region)

	// Check machine pool configuration
	assert.Len(t, cluster.Spec.Machinepoolconfig, 1)
	assert.Equal(t, "pool1", *cluster.Spec.Machinepoolconfig[0].PoolConfig.Name)
	assert.Equal(t, int32(3), *cluster.Spec.Machinepoolconfig[0].PoolConfig.Size)
	assert.Equal(t, "n1-standard-2", *cluster.Spec.Machinepoolconfig[0].CloudConfig.InstanceType)
	assert.Equal(t, int64(100), cluster.Spec.Machinepoolconfig[0].CloudConfig.RootDeviceSize)
}

func TestFlattenMachinePoolConfigsGke(t *testing.T) {
	// Simulate input data
	machinePools := []*models.V1GcpMachinePoolConfig{
		{
			InstanceType:   ptr.To("n1-standard-2"),
			Name:           "pool1",
			RootDeviceSize: 100,
			Size:           3,
		},
		{
			InstanceType:   ptr.To("n1-standard-4"),
			Name:           "pool2",
			Size:           2,
			RootDeviceSize: 200,
		},
	}

	// Call the flattenMachinePoolConfigsGke function with the simulated input data
	result := flattenMachinePoolConfigsGke(machinePools)

	// Assertions
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	// Check the first machine pool
	pool1 := result[0].(map[string]interface{})
	assert.Equal(t, "pool1", pool1["name"])
	assert.Equal(t, 3, pool1["count"])
	assert.Equal(t, "n1-standard-2", pool1["instance_type"])
	assert.Equal(t, 100, pool1["disk_size_gb"])

	// Check the second machine pool
	pool2 := result[1].(map[string]interface{})
	assert.Equal(t, "pool2", pool2["name"])
	assert.Equal(t, 2, pool2["count"])
	assert.Equal(t, "n1-standard-4", pool2["instance_type"])
	assert.Equal(t, 200, pool2["disk_size_gb"])
}
