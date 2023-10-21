package spectrocloud

import (
	"github.com/spectrocloud/hapi/models"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
)

func prepareVirtualClusterTestData() *schema.ResourceData {
	d := resourceClusterVirtual().TestResourceData()

	d.SetId("")
	d.Set("name", "virtual-picard-2")

	// Cluster Profile for Virtual Cluster
	cProfile := make([]map[string]interface{}, 0)
	cProfile = append(cProfile, map[string]interface{}{
		"id": "virtual-basic-infra-profile-id",
	})
	d.Set("cluster_profile", cProfile)
	d.Set("host_cluster_uid", "host-cluster-id")
	d.Set("cluster_group_uid", "group-cluster-id")

	// Cloud Config for Virtual Cluster
	cloudConfig := make([]map[string]interface{}, 0)
	vCloud := map[string]interface{}{
		"chart_name":    "virtual-chart-name",
		"chart_repo":    "virtual-chart-repo",
		"chart_version": "v1.0.0",
		"chart_values":  "default-values",
		"k8s_version":   "v1.20.0",
	}
	cloudConfig = append(cloudConfig, vCloud)
	d.Set("cloud_config", cloudConfig)

	return d
}

func TestToVirtualCluster(t *testing.T) {
	assert := assert.New(t)
	// Create a mock ResourceData object
	d := prepareVirtualClusterTestData()

	// Mock the client
	mockClient := &client.V1Client{}

	// Create a mock ResourceData for testing
	vCluster, err := toVirtualCluster(mockClient, d)
	assert.Nil(err)

	// Check the output against the expected values

	// Verifying cluster name attribute
	assert.Equal(d.Get("name").(string), vCluster.Metadata.Name)

	// Verifying host cluster uid and cluster group uid attributes
	assert.Equal(d.Get("host_cluster_uid").(string), vCluster.Spec.ClusterConfig.HostClusterConfig.HostCluster.UID)
	assert.Equal(d.Get("cluster_group_uid").(string), vCluster.Spec.ClusterConfig.HostClusterConfig.ClusterGroup.UID)

	// Verifying cloud config attributes
	val, _ := d.GetOk("cloud_config")
	cloudConfig := val.([]interface{})[0].(map[string]interface{})
	assert.Equal(cloudConfig["chart_name"].(string), vCluster.Spec.CloudConfig.HelmRelease.Chart.Name)
	assert.Equal(cloudConfig["chart_repo"].(string), vCluster.Spec.CloudConfig.HelmRelease.Chart.Repo)
	assert.Equal(cloudConfig["chart_version"].(string), vCluster.Spec.CloudConfig.HelmRelease.Chart.Version)
	assert.Equal(cloudConfig["chart_values"].(string), vCluster.Spec.CloudConfig.HelmRelease.Values)
	assert.Equal(cloudConfig["k8s_version"].(string), vCluster.Spec.CloudConfig.KubernetesVersion)
}

func TestToVirtualClusterResize(t *testing.T) {
	resources := map[string]interface{}{
		"max_cpu":           4,
		"max_mem_in_mb":     8192,
		"max_storage_in_gb": 100,
		"min_cpu":           2,
		"min_mem_in_mb":     4096,
		"min_storage_in_gb": 50,
	}

	expected := &models.V1VirtualClusterResize{
		InstanceType: &models.V1VirtualInstanceType{
			MaxCPU:        4,
			MaxMemInMiB:   8192,
			MaxStorageGiB: 100,
			MinCPU:        2,
			MinMemInMiB:   4096,
			MinStorageGiB: 50,
		},
	}

	result := toVirtualClusterResize(resources)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}
