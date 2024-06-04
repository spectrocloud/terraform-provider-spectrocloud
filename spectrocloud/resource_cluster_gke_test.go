package spectrocloud

import (
	"errors"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"github.com/stretchr/testify/assert"
	"testing"
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
			InstanceType:   types.Ptr("n1-standard-2"),
			Name:           "pool1",
			RootDeviceSize: 100,
			Size:           3,
		},
		{
			InstanceType:   types.Ptr("n1-standard-4"),
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

func TestFlattenClusterProfileForImport(t *testing.T) {
	m := &client.V1Client{
		GetClusterFn: func(scope, uid string) (*models.V1SpectroCluster, error) {
			var profiles []*models.V1ClusterProfileTemplate
			p1 := &models.V1ClusterProfileTemplate{
				CloudType:        "",
				Name:             "",
				PackServerRefs:   nil,
				PackServerSecret: "",
				Packs:            nil,
				ProfileVersion:   "",
				RelatedObject:    nil,
				Type:             "",
				UID:              "profile-1",
				Version:          0,
			}
			p2 := &models.V1ClusterProfileTemplate{
				CloudType:        "",
				Name:             "",
				PackServerRefs:   nil,
				PackServerSecret: "",
				Packs:            nil,
				ProfileVersion:   "",
				RelatedObject:    nil,
				Type:             "",
				UID:              "profile-2",
				Version:          0,
			}
			profiles = append(profiles, p1)
			profiles = append(profiles, p2)

			cluster := &models.V1SpectroCluster{
				APIVersion: "",
				Kind:       "",
				Metadata:   nil,
				Spec: &models.V1SpectroClusterSpec{
					CloudConfigRef:          nil,
					CloudType:               "",
					ClusterConfig:           nil,
					ClusterProfileTemplates: profiles,
					ClusterType:             "",
				},
				Status: nil,
			}
			return cluster, nil
		},
	}

	// Test case: Successfully retrieve cluster profiles
	clusterContext := "project"
	clusterID := "test-cluster-id"
	clusterProfiles := []interface{}{
		map[string]interface{}{"id": "profile-1"},
		map[string]interface{}{"id": "profile-2"},
	}
	mockResourceData := resourceClusterGke().TestResourceData()
	err := mockResourceData.Set("cluster_profile", clusterProfiles)
	if err != nil {
		return
	}
	err = mockResourceData.Set("context", clusterContext)
	if err != nil {
		return
	}
	mockResourceData.SetId(clusterID)

	result, err := flattenClusterProfileForImport(m, mockResourceData)
	assert.NoError(t, err)
	assert.Equal(t, clusterProfiles, result)

	//Test case: Error retrieving cluster
	m = &client.V1Client{
		GetClusterFn: func(scope, uid string) (*models.V1SpectroCluster, error) {

			return nil, errors.New("error retrieving cluster")
		},
	}
	result, err = flattenClusterProfileForImport(m, mockResourceData)
	assert.Error(t, err)
	assert.Empty(t, result)
}
