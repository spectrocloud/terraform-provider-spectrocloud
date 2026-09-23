package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func prepareClusterGroupTestData() (*schema.ResourceData, error) {
	d := resourceClusterGroup().TestResourceData()
	d.SetId("test-cg-1")
	err := d.Set("context", "project")
	if err != nil {
		return nil, err
	}
	err = d.Set("name", "test-name")
	if err != nil {
		return nil, err
	}
	err = d.Set("tags", []string{"key1:value1", "key2:value2"})
	if err != nil {
		return nil, err
	}
	err = d.Set("config", []map[string]interface{}{
		{
			"host_endpoint_type":       "LoadBalancer",
			"cpu_millicore":            4000,
			"memory_in_mb":             4096,
			"storage_in_gb":            100,
			"oversubscription_percent": 200,
			"values":                   "namespace: test-namespace",
		},
	})
	if err != nil {
		return nil, err
	}
	err = d.Set("clusters", []map[string]interface{}{
		{
			"cluster_uid": "test-cluster-uid",
			"host_dns":    "https://test.dev.spectro.com",
		},
	})
	if err != nil {
		return nil, err
	}
	err = d.Set("cluster_profile", []map[string]interface{}{
		{
			"id": "test-cluster-uid",
		},
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}

func TestDefaultValuesSet(t *testing.T) {
	clusterGroupLimitConfig := &models.V1ClusterGroupLimitConfig{}
	hostClusterConfig := []*models.V1ClusterGroupHostClusterConfig{{}}
	endpointType := "testEndpointType"
	nonEmptyValues := "testValues"
	emptyValues := ""

	t.Run("Test with non-empty values", func(t *testing.T) {
		result := GetClusterGroupConfig(clusterGroupLimitConfig, hostClusterConfig, endpointType, nonEmptyValues, "k3s")

		assert.Equal(t, endpointType, result.EndpointType)
		assert.Equal(t, clusterGroupLimitConfig, result.LimitConfig)
		assert.Equal(t, hostClusterConfig, result.HostClustersConfig)
		assert.Equal(t, nonEmptyValues, result.Values)
	})

	t.Run("Test with empty values", func(t *testing.T) {
		result := GetClusterGroupConfig(clusterGroupLimitConfig, hostClusterConfig, endpointType, emptyValues, "k3s")

		assert.Equal(t, endpointType, result.EndpointType)
		assert.Equal(t, clusterGroupLimitConfig, result.LimitConfig)
		assert.Equal(t, hostClusterConfig, result.HostClustersConfig)
		assert.Equal(t, "", result.Values)
	})
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
		Spec: &models.V1ClusterGroupSpecEntity{
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
	endpointType := "Ingress"
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
				EndpointType: endpointType,
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
	assert.Equal(t, endpointType, config["host_endpoint_type"])
	assert.Equal(t, cpuLimit, config["cpu_millicore"])
	assert.Equal(t, memoryLimit, config["memory_in_mb"])
	assert.Equal(t, storageLimit, config["storage_in_gb"])
	assert.Equal(t, overSubscription, config["oversubscription_percent"])

	// assert clusters fields are set correctly
	clustersList := d.Get("clusters").([]interface{})
	assert.Len(t, clustersList, 2)
	cluster1 := clustersList[0].(map[string]interface{})
	assert.Equal(t, clusterUID1, cluster1["cluster_uid"])
	assert.Equal(t, host1, cluster1["host_dns"])
	cluster2 := clustersList[1].(map[string]interface{})
	assert.Equal(t, clusterUID2, cluster2["cluster_uid"])
	assert.Equal(t, host2, cluster2["host_dns"])
}

func TestToClusterRef(t *testing.T) {
	d, err := prepareClusterGroupTestData()
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	cluster := d.Get("clusters").([]interface{})[0].(map[string]interface{})
	ret := toClusterRef(cluster)
	assert.Equal(t, ret.ClusterUID, cluster["cluster_uid"])
}

func TestToHostClusterConfigs(t *testing.T) {
	d, err := prepareClusterGroupTestData()
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	hostConfigs := d.Get("clusters").([]interface{})
	clusterUid := hostConfigs[0].(map[string]interface{})["cluster_uid"]
	hostDns := hostConfigs[0].(map[string]interface{})["host_dns"]
	hostClusterConfigs := toHostClusterConfigs(hostConfigs)
	assert.Equal(t, clusterUid, hostClusterConfigs[0].ClusterUID)
	assert.Equal(t, hostDns, hostClusterConfigs[0].EndpointConfig.IngressConfig.Host)
}
func TestResourceClusterGroupCRUD(t *testing.T) {
	testResourceCRUD(t, func() *schema.ResourceData {
		d, err := prepareClusterGroupTestData()
		if err != nil {
			t.Fatal(err)
		}
		return d
	}, unitTestMockAPIClient,
		resourceClusterGroupCreate, resourceClusterGroupRead, resourceClusterGroupUpdate, resourceClusterGroupDelete)
}

// PLT-2327: k8s_distribution = "k3s" is deprecated (PPD-1605) and must be
// rejected at Create time for new cluster groups, without ever reaching the
// API. Existing k3s cluster groups must remain updatable, which is why the
// check lives only in resourceClusterGroupCreate.
func TestResourceClusterGroupCreate_RejectsK3s(t *testing.T) {
	d, err := prepareClusterGroupTestData()
	if err != nil {
		t.Fatal(err)
	}
	config := d.Get("config").([]interface{})[0].(map[string]interface{})
	config["k8s_distribution"] = "k3s"
	if err := d.Set("config", []map[string]interface{}{config}); err != nil {
		t.Fatal(err)
	}

	testResourceCRUDNegative(t, "Create", func() *schema.ResourceData { return d }, unitTestMockAPIClient,
		resourceClusterGroupCreate, resourceClusterGroupRead, resourceClusterGroupUpdate, resourceClusterGroupDelete,
		false, "k3s")
}

// Sanity check: the default k8s_distribution ("vcluster-generic") and the
// other still-supported value must not trip the guard.
func TestValidateClusterGroupK8sDistribution(t *testing.T) {
	for _, distro := range []string{"", "vcluster-generic", "cncf_k8s"} {
		t.Run(distro, func(t *testing.T) {
			d, err := prepareClusterGroupTestData()
			if err != nil {
				t.Fatal(err)
			}
			config := d.Get("config").([]interface{})[0].(map[string]interface{})
			config["k8s_distribution"] = distro
			if err := d.Set("config", []map[string]interface{}{config}); err != nil {
				t.Fatal(err)
			}
			assert.NoError(t, validateClusterGroupK8sDistribution(d))
		})
	}

	t.Run("k3s is rejected", func(t *testing.T) {
		d, err := prepareClusterGroupTestData()
		if err != nil {
			t.Fatal(err)
		}
		config := d.Get("config").([]interface{})[0].(map[string]interface{})
		config["k8s_distribution"] = "k3s"
		if err := d.Set("config", []map[string]interface{}{config}); err != nil {
			t.Fatal(err)
		}
		err = validateClusterGroupK8sDistribution(d)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "k3s")
	})
}
