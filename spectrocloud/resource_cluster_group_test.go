package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prepareTestData() *schema.ResourceData {
	d := resourceClusterGroup().TestResourceData()
	d.SetId("")
	d.Set("name", "test-name")
	d.Set("tags", []string{"key1:value1", "key2:value2"})
	d.Set("config", []map[string]interface{}{
		{
			"cpu_millicore":            4000,
			"memory_in_mb":             4096,
			"storage_in_gb":            100,
			"oversubscription_percent": 200,
		},
	})
	d.Set("clusters", []map[string]interface{}{
		{
			"cluster_uid": "test-cluster-uid",
		},
	})
	return d
}

func TestToClusterGroup(t *testing.T) {
	assert := assert.New(t)

	// Create a mock ResourceData object
	d := prepareTestData()

	// Call the function with the mock resource data
	output := toClusterGroup(d)

	// Check the output against the expected values
	assert.Equal("test-name", output.Metadata.Name)
	assert.Equal("", output.Metadata.UID)
	assert.Equal(2, len(output.Metadata.Labels))
	assert.Equal("hostCluster", output.Spec.Type)
	assert.Equal(1, len(output.Spec.ClusterRefs))
	assert.Equal("test-cluster-uid", output.Spec.ClusterRefs[0].ClusterUID)
	assert.Equal(int32(4000), output.Spec.ClustersConfig.LimitConfig.CPUMilliCore)
	assert.Equal(int32(4096), output.Spec.ClustersConfig.LimitConfig.MemoryMiB)
	assert.Equal(int32(100), output.Spec.ClustersConfig.LimitConfig.StorageGiB)
	assert.Equal(int32(200), output.Spec.ClustersConfig.LimitConfig.OverSubscription)
}

func TestToClusterGroupLimitConfig(t *testing.T) {
	resources := map[string]interface{}{
		"cpu_millicore":            4000,
		"memory_in_mb":             4096,
		"storage_in_gb":            100,
		"oversubscription_percent": 200,
	}

	limitConfig := toClusterGroupLimitConfig(resources)
	assert.Equal(t, limitConfig.CPUMilliCore, int32(4000))
	assert.Equal(t, limitConfig.MemoryMiB, int32(4096))
	assert.Equal(t, limitConfig.StorageGiB, int32(100))
	assert.Equal(t, limitConfig.OverSubscription, int32(200))
}

func TestResourceClusterGroupCreate(t *testing.T) {
	m := &client.V1Client{
		CreateClusterGroupFn: func(cluster *models.V1ClusterGroupEntity) (string, error) {
			return "test-uid", nil
		},
		GetClusterGroupFn: func(uid string) (*models.V1ClusterGroup, error) {
			return &models.V1ClusterGroup{
				Metadata: &models.V1ObjectMeta{
					UID: uid,
				},
			}, nil
		},
	}

	d := prepareTestData()
	ctx := context.Background()

	diags := resourceClusterGroupCreate(ctx, d, m)
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}

	if d.Id() != "test-uid" {
		t.Errorf("Expected ID to be 'test-uid', got %s", d.Id())
	}
}

func TestToClusterGroupUpdate(t *testing.T) {
	// Set up test data
	clusterRefs := []*models.V1ClusterGroupClusterRef{
		{

			ClusterName: "cluster-1",
			ClusterUID:  "cluster-uid-1",
		},
		{
			ClusterName: "cluster-2",
			ClusterUID:  "cluster-uid-2",
		},
	}
	hostClustersConfig := []*models.V1ClusterGroupHostClusterConfig{
		{
			ClusterUID: "cluster-uid-1",
			EndpointConfig: &models.V1HostClusterEndpointConfig{
				IngressConfig: &models.V1IngressConfig{
					Host: "host-1",
				},
			},
		},
		{
			ClusterUID: "cluster-uid-2",
			EndpointConfig: &models.V1HostClusterEndpointConfig{
				LoadBalancerConfig: &models.V1LoadBalancerConfig{
					// Add test cases for LoadBalancerConfig fields here
				},
			},
		},
	}
	limitConfig := &models.V1ClusterGroupLimitConfig{
		CPUMilliCore:     1000,
		MemoryMiB:        2048,
		OverSubscription: 200,
		StorageGiB:       100,
	}
	clustersConfig := &models.V1ClusterGroupClustersConfig{
		EndpointType:       "LoadBalancer",
		HostClustersConfig: hostClustersConfig,
		LimitConfig:        limitConfig,
		Values:             "my_values",
	}
	clusterGroupEntity := &models.V1ClusterGroupEntity{
		Spec: &models.V1ClusterGroupSpec{
			ClusterRefs:    clusterRefs,
			ClustersConfig: clustersConfig,
		},
	}

	// Call the function under test
	result := toClusterGroupUpdate(clusterGroupEntity)

	// Assert the result is correct
	assert.Equal(t, clusterRefs, result.ClusterRefs)
	assert.Equal(t, clustersConfig.EndpointType, result.ClustersConfig.EndpointType)
	assert.Equal(t, len(hostClustersConfig), len(result.ClustersConfig.HostClustersConfig))
	for i := range hostClustersConfig {
		assert.Equal(t, hostClustersConfig[i].ClusterUID, result.ClustersConfig.HostClustersConfig[i].ClusterUID)
		assert.Equal(t, hostClustersConfig[i].EndpointConfig.IngressConfig, result.ClustersConfig.HostClustersConfig[i].EndpointConfig.IngressConfig)
		assert.Equal(t, hostClustersConfig[i].EndpointConfig.LoadBalancerConfig, result.ClustersConfig.HostClustersConfig[i].EndpointConfig.LoadBalancerConfig)
	}
	assert.Equal(t, limitConfig, result.ClustersConfig.LimitConfig)
	assert.Equal(t, "my_values", result.ClustersConfig.Values)
}

func TestFlattenClusterGroup(t *testing.T) {
	// set up test data
	name := "test-cluster-group"
	uid := "1234"
	cpuLimit := 1000
	memoryLimit := 2000
	storageLimit := 3000
	overSubscription := 500
	host1 := "host1"
	host2 := "host2"
	clusterUID1 := "cluster-uid-1"
	clusterUID2 := "cluster-uid-2"

	clusterGroup := &models.V1ClusterGroup{
		Metadata: &models.V1ObjectMeta{
			Name:   name,
			UID:    uid,
			Labels: map[string]string{"key1": "value1", "key2": "value2"},
		},
		Spec: &models.V1ClusterGroupSpec{
			ClustersConfig: &models.V1ClusterGroupClustersConfig{
				LimitConfig: &models.V1ClusterGroupLimitConfig{
					CPUMilliCore:     int32(cpuLimit),
					MemoryMiB:        int32(memoryLimit),
					StorageGiB:       int32(storageLimit),
					OverSubscription: int32(overSubscription),
				},
				HostClustersConfig: []*models.V1ClusterGroupHostClusterConfig{
					{
						ClusterUID: clusterUID1,
						EndpointConfig: &models.V1HostClusterEndpointConfig{
							IngressConfig: &models.V1IngressConfig{
								Host: host1,
							},
						},
					},
					{
						ClusterUID: clusterUID2,
						EndpointConfig: &models.V1HostClusterEndpointConfig{
							IngressConfig: &models.V1IngressConfig{
								Host: host2,
							},
						},
					},
				},
			},
		},
	}

	d := resourceClusterGroup().TestResourceData()
	diags := flattenClusterGroup(clusterGroup, d)
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}

	// assert cluster group fields are set correctly
	assert.Equal(t, uid, d.Id())
	assert.Equal(t, name, d.Get("name"))
	tags := d.Get("tags").(*schema.Set)
	expectedTags := []string{"key1:value1", "key2:value2"}
	assert.ElementsMatch(t, expectedTags, tags.List())

	// assert config fields are set correctly
	configList := d.Get("config").([]interface{})
	assert.Len(t, configList, 1)
	config := configList[0].(map[string]interface{})
	assert.Equal(t, cpuLimit, config["cpu_millicore"])
	assert.Equal(t, memoryLimit, config["memory_in_mb"])
	assert.Equal(t, storageLimit, config["storage_in_gb"])
	assert.Equal(t, overSubscription, config["oversubscription_percent"])

	// assert clusters fields are set correctly
	clustersList := d.Get("clusters").([]interface{})
	assert.Len(t, clustersList, 2)
	cluster1 := clustersList[0].(map[string]interface{})
	assert.Equal(t, clusterUID1, cluster1["cluster_uid"])
	assert.Equal(t, host1, cluster1["host"])
	cluster2 := clustersList[1].(map[string]interface{})
	assert.Equal(t, clusterUID2, cluster2["cluster_uid"])
	assert.Equal(t, host2, cluster2["host"])
}
