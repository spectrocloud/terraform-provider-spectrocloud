package spectrocloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
)

func prepareClusterGroupTestData() (*schema.ResourceData, error) {
	d := resourceClusterGroup().TestResourceData()
	d.SetId("")
	err := d.Set("name", "test-name")
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

func TestToClusterGroup(t *testing.T) {
	assert := assert.New(t)

	// Create a mock ResourceData object
	d, err := prepareClusterGroupTestData()
	if err != nil {
		t.Errorf(err.Error())
	}
	m := &client.V1Client{}
	// Call the function with the mock resource data
	output := toClusterGroup(m, d)

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
	assert.Equal("namespace: test-namespace", output.Spec.ClustersConfig.Values)
	assert.Equal("LoadBalancer", output.Spec.ClustersConfig.EndpointType)
	assert.Equal("test-cluster-uid", output.Spec.Profiles[0].UID)
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

	d, err := prepareClusterGroupTestData()
	if err != nil {
		t.Errorf(err.Error())
	}
	ctx := context.Background()

	diags := resourceClusterGroupCreate(ctx, d, m)
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}

	if d.Id() != "test-uid" {
		t.Errorf("Expected ID to be 'test-uid', got %s", d.Id())
	}
}

func TestResourceClusterGroupDelete(t *testing.T) {
	testUid := "unit_test_uid"
	testscope := "project"
	m := &client.V1Client{
		DeleteClusterGroupFn: func(uid string) error {
			if uid != testUid {
				return fmt.Errorf("this UID `%s` doesn't match with test uid `%s`", uid, testUid)
			}
			return nil
		},
	}
	e := m.DeleteClusterGroup(testUid, testscope)
	if e != nil {
		t.Errorf("Expectred nil, got %s", e)
	}
}

func TestResourceClusterGroupUpdate(t *testing.T) {
	d, err := prepareClusterGroupTestData()
	if err != nil {
		t.Errorf(err.Error())
	}
	clusterConfig := []map[string]interface{}{
		{
			"host_endpoint_type":       "LoadBalancer",
			"cpu_millicore":            5000,
			"memory_in_mb":             5096,
			"storage_in_gb":            150,
			"oversubscription_percent": 120,
		},
	}
	d.Set("config", clusterConfig)
	m := &client.V1Client{
		UpdateClusterGroupFn: func(uid string, cg *models.V1ClusterGroupHostClusterEntity) error {
			assert.Equal(t, int(cg.ClustersConfig.LimitConfig.MemoryMiB), clusterConfig[0]["memory_in_mb"])
			assert.Equal(t, int(cg.ClustersConfig.LimitConfig.StorageGiB), clusterConfig[0]["storage_in_gb"])
			assert.Equal(t, int(cg.ClustersConfig.LimitConfig.CPUMilliCore), clusterConfig[0]["cpu_millicore"])
			assert.Equal(t, int(cg.ClustersConfig.LimitConfig.OverSubscription), clusterConfig[0]["oversubscription_percent"])
			assert.Equal(t, cg.ClustersConfig.EndpointType, clusterConfig[0]["host_endpoint_type"])
			return nil
		},
	}
	ctx := context.Background()
	resourceClusterGroupUpdate(ctx, d, m)
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
		t.Errorf(err.Error())
	}
	cluster := d.Get("clusters").([]interface{})[0].(map[string]interface{})
	ret := toClusterRef(cluster)
	assert.Equal(t, ret.ClusterUID, cluster["cluster_uid"])
}

func TestToHostClusterConfigs(t *testing.T) {
	d, err := prepareClusterGroupTestData()
	if err != nil {
		t.Errorf(err.Error())
	}
	hostConfigs := d.Get("clusters").([]interface{})
	clusterUid := hostConfigs[0].(map[string]interface{})["cluster_uid"]
	hostDns := hostConfigs[0].(map[string]interface{})["host_dns"]
	hostClusterConfigs := toHostClusterConfigs(hostConfigs)
	assert.Equal(t, clusterUid, hostClusterConfigs[0].ClusterUID)
	assert.Equal(t, hostDns, hostClusterConfigs[0].EndpointConfig.IngressConfig.Host)
}
