package spectrocloud

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func prepareClusterVsphereTestData() *schema.ResourceData {
	d := resourceClusterVsphere().TestResourceData()

	d.SetId("")
	d.Set("name", "vsphere-picard-2")
	cConfig := make([]map[string]interface{}, 0)
	cConfig = append(cConfig, map[string]interface{}{
		"id": "vmware-basic-infra-profile-id",
	})
	d.Set("cluster_meta_attribute", "{'nic_name': 'test', 'env': 'stage'}")
	d.Set("cluster_profile", cConfig)
	d.Set("cloud_account_id", "vmware-basic-account-id")

	// cloud config
	keys := []string{"SSHKey1", "SSHKey2"}
	cloudConfig := make([]map[string]interface{}, 0)
	con := map[string]interface{}{
		"ssh_keys":              keys,
		"datacenter":            "Datacenter",
		"folder":                "sc_test/terraform",
		"network_type":          "DDNS",
		"network_search_domain": "spectrocloud.dev",
	}
	cloudConfig = append(cloudConfig, con)
	d.Set("cloud_config", cloudConfig)

	mPools := make([]map[string]interface{}, 0)

	// Adding control-plane pool
	cpPlacement := make([]interface{}, 0)
	cpPlacement = append(cpPlacement, map[string]interface{}{
		"id":                "",
		"cluster":           "test cluster",
		"resource_pool":     "Default",
		"datastore":         "datastore55_2",
		"network":           "VM Network",
		"static_ip_pool_id": "testpoolid",
	})
	cpInstance := make([]interface{}, 0)
	cpInstance = append(cpInstance, map[string]interface{}{
		"disk_size_gb": 40,
		"memory_mb":    8192,
		"cpu":          4,
	})
	mPools = append(mPools, map[string]interface{}{
		"control_plane":           true,
		"control_plane_as_worker": true,
		"name":                    "cp-pool",
		"count":                   1,
		"placement":               cpPlacement,
		"instance_type":           cpInstance,
		"node":                    []interface{}{},
	})

	// Adding Worker pool
	workerPlacement := make([]interface{}, 0)
	workerPlacement = append(workerPlacement, map[string]interface{}{
		"id":                "",
		"cluster":           "test cluster",
		"resource_pool":     "Default",
		"datastore":         "datastore55_2",
		"network":           "VM Network",
		"static_ip_pool_id": "testpoolid",
	})

	workerInstance := make([]interface{}, 0)
	workerInstance = append(workerInstance, map[string]interface{}{
		"disk_size_gb": 40,
		"memory_mb":    8192,
		"cpu":          4,
	})

	mPools = append(mPools, map[string]interface{}{
		"control_plane":           false,
		"control_plane_as_worker": false,
		"name":                    "worker-basic",
		"count":                   1,
		"min":                     1,
		"max":                     3,
		"placement":               workerPlacement,
		"instance_type":           workerInstance,
		"node":                    []interface{}{},
	})
	d.Set("machine_pool", mPools)
	return d
}

func TestToCloudConfigUpdate(t *testing.T) {
	assert := assert.New(t)
	cloudConfig := map[string]interface{}{
		"ssh_key":               "ssh-rsa AAAAB3NzaC1y",
		"datacenter":            "Datacenter",
		"folder":                "sc_test/terraform",
		"network_type":          "DDNS",
		"host_endpoint":         "tt.tt.test.com",
		"network_search_domain": "spectrocloud.dev",
		"static_ip":             false,
	}
	cloudEntity := toCloudConfigUpdate(cloudConfig)
	assert.Equal("spectrocloud.dev", cloudEntity.ClusterConfig.ControlPlaneEndpoint.DdnsSearchDomain)
	assert.Equal("DDNS", cloudEntity.ClusterConfig.ControlPlaneEndpoint.Type)
	assert.Equal("Datacenter", cloudEntity.ClusterConfig.Placement.Datacenter)
	assert.Equal("sc_test/terraform", cloudEntity.ClusterConfig.Placement.Folder)
	assert.Equal("spectro-templates", cloudEntity.ClusterConfig.Placement.ImageTemplateFolder)
	assert.Equal("ssh-rsa AAAAB3NzaC1y", cloudEntity.ClusterConfig.SSHKeys[0])
	assert.Equal("tt.tt.test.com", cloudEntity.ClusterConfig.ControlPlaneEndpoint.Host)
	assert.Equal(false, cloudEntity.ClusterConfig.StaticIP)
}

func getMachinePlacement() []*models.V1VspherePlacementConfig {
	network := new(string)
	*network = "test-net"
	var placement []*models.V1VspherePlacementConfig
	placement = append(placement, &models.V1VspherePlacementConfig{
		Cluster:             "test-cluster",
		Datacenter:          "vsphere",
		Datastore:           "vcenter",
		Folder:              "/test/",
		ImageTemplateFolder: "",
		Network: &models.V1VsphereNetworkConfig{
			IPPool:      nil,
			NetworkName: network,
			ParentPoolRef: &models.V1ObjectReference{
				UID: "test-pool-id",
			},
			StaticIP: false,
		},
		ResourcePool:      "",
		StoragePolicyName: "",
		UID:               "test-uid",
	})
	return placement
}

func getMPools() []*models.V1VsphereMachinePoolConfig {
	var mTaint []*models.V1Taint
	diskGb := new(int32)
	*diskGb = 23
	memMb := new(int64)
	*memMb = 120
	numCpu := new(int32)
	*numCpu = 2
	mTaint = append(mTaint, &models.V1Taint{
		Effect:    "start",
		Key:       "owner",
		TimeAdded: models.V1Time{},
		Value:     "siva",
	})
	var mPool []*models.V1VsphereMachinePoolConfig
	mPool = append(mPool, &models.V1VsphereMachinePoolConfig{
		AdditionalLabels: map[string]string{
			"type":  "unittest",
			"owner": "siva",
		},
		AdditionalTags: nil,
		InstanceType: &models.V1VsphereInstanceType{
			DiskGiB:   diskGb,
			MemoryMiB: memMb,
			NumCPUs:   numCpu,
		},
		IsControlPlane: nil,
		Labels:         nil,
		MaxSize:        0,
		MinSize:        0,
		Name:           "",
		Placements:     getMachinePlacement(),
		Size:           0,
		Taints:         mTaint,
		UpdateStrategy: &models.V1UpdateStrategy{
			Type: "",
		},
		UseControlPlaneAsWorker: false,
	})
	return mPool
}

func getCloudConfig() *models.V1VsphereCloudConfig {
	cloudConfig := &models.V1VsphereCloudConfig{
		APIVersion: "v1",
		Kind:       "",
		Metadata:   nil,
		Spec: &models.V1VsphereCloudConfigSpec{
			CloudAccountRef: &models.V1ObjectReference{
				Kind: "",
				Name: "",
				UID:  "vmware-basic-account-id",
			},
			ClusterConfig:     nil,
			EdgeHostRef:       nil,
			MachinePoolConfig: getMPools(),
		},
		Status: nil,
	}
	return cloudConfig
}

func TestFlattenClusterConfigsVsphere(t *testing.T) {
	inputCloudConfig := &models.V1VsphereCloudConfig{
		Spec: &models.V1VsphereCloudConfigSpec{
			ClusterConfig: &models.V1VsphereClusterConfig{
				SSHKeys:    []string{"SSHKey1", "SSHKey1"},
				StaticIP:   true,
				NtpServers: []string{"ntp1", "ntp2"},
				Placement: &models.V1VspherePlacementConfig{
					Datacenter: "Datacenter1",
					Folder:     "Folder1",
				},
				ControlPlaneEndpoint: &models.V1ControlPlaneEndPoint{
					Type:             "VIP",
					DdnsSearchDomain: "example.dev",
				},
			},
		},
	}
	d := prepareClusterVsphereTestData()
	flattenedConfig := flattenClusterConfigsVsphere(d, inputCloudConfig)

	flattenedConfigMap := flattenedConfig.([]interface{})[0].(map[string]interface{})
	if flattenedConfigMap["datacenter"].(string) != inputCloudConfig.Spec.ClusterConfig.Placement.Datacenter {
		t.Errorf("Failed to flatten 'datacenter' field correctly")
	}
	if flattenedConfigMap["folder"].(string) != inputCloudConfig.Spec.ClusterConfig.Placement.Folder {
		t.Errorf("Failed to flatten 'folder' field correctly")
	}
	if !reflect.DeepEqual(flattenedConfigMap["ssh_keys"].([]string), inputCloudConfig.Spec.ClusterConfig.SSHKeys) {
		t.Errorf("Failed to flatten 'ssh_keys' field correctly")
	}
	if flattenedConfigMap["static_ip"].(bool) != inputCloudConfig.Spec.ClusterConfig.StaticIP {
		t.Errorf("Failed to flatten 'static_ip' field correctly")
	}
	if flattenedConfigMap["network_type"].(string) != inputCloudConfig.Spec.ClusterConfig.ControlPlaneEndpoint.Type {
		t.Errorf("Failed to flatten 'network_type' field correctly")
	}
	if flattenedConfigMap["network_search_domain"].(string) != inputCloudConfig.Spec.ClusterConfig.ControlPlaneEndpoint.DdnsSearchDomain {
		t.Errorf("Failed to flatten 'network_search_domain' field correctly")
	}
	flattenedNtpServers := flattenedConfigMap["ntp_servers"].([]string)
	if !reflect.DeepEqual(flattenedNtpServers, inputCloudConfig.Spec.ClusterConfig.NtpServers) {
		t.Errorf("Failed to flatten 'ntp_servers' field correctly")
	}
}

func TestFlattenClusterConfigsVsphereNil(t *testing.T) {
	d := prepareClusterVsphereTestData()
	flatCloudConfig := flattenClusterConfigsVsphere(d, nil)
	if flatCloudConfig == nil {
		t.Errorf("flattenClusterConfigsVsphere returning value for nill: %#v", flatCloudConfig)
	}
}

func TestFlattenMachinePoolConfigsVsphereNil(t *testing.T) {
	flatPool := flattenMachinePoolConfigsVsphere(nil)
	if len(flatPool) > 0 {
		t.Errorf("flattenMachinePoolConfigsVsphere returning value for nill: %#v", flatPool)
	}
}

func TestFlattenMachinePoolConfigsVsphere(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		input    []*models.V1VsphereMachinePoolConfig
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: []interface{}{},
		},
		{
			name:     "empty input",
			input:    []*models.V1VsphereMachinePoolConfig{},
			expected: []interface{}{},
		},
		{
			name: "valid input",
			input: []*models.V1VsphereMachinePoolConfig{
				{
					Name:                    "pool1", // Match this name with input data
					Size:                    int32(3),
					MinSize:                 1,
					MaxSize:                 5,
					IsControlPlane:          ptr.To(true),
					UseControlPlaneAsWorker: false,
					NodeRepaveInterval:      int32(24),
					UpdateStrategy: &models.V1UpdateStrategy{
						Type: "RollingUpdate",
					},
					InstanceType: &models.V1VsphereInstanceType{
						DiskGiB:   ptr.To(int32(100)),
						MemoryMiB: ptr.To(int64(8192)),
						NumCPUs:   ptr.To(int32(4)),
					},
					Placements: []*models.V1VspherePlacementConfig{
						{
							UID:          "placement1",
							Cluster:      "cluster1",
							ResourcePool: "resource-pool1",
							Datastore:    "datastore1",
							Network: &models.V1VsphereNetworkConfig{
								NetworkName: ptr.To("network1"),
								ParentPoolRef: &models.V1ObjectReference{
									UID: "pool1",
								},
							},
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":                    "pool1", // Match with the input data
					"count":                   int32(3),
					"min":                     1,
					"max":                     5,
					"control_plane_as_worker": false,
					"control_plane":           true, // Include additional fields returned by the function
					"instance_type": []interface{}{
						map[string]interface{}{
							"disk_size_gb": 100,
							"memory_mb":    8192,
							"cpu":          4,
						},
					},
					"placement": []interface{}{
						map[string]interface{}{
							"id":                "placement1",
							"cluster":           "cluster1",
							"resource_pool":     "resource-pool1",
							"datastore":         "datastore1",
							"network":           ptr.To("network1"), // Handle pointer or use (*string)(nil) if necessary
							"static_ip_pool_id": "pool1",
						},
					},
					"update_strategy":   "RollingUpdate",          // Include this field in expected
					"additional_labels": map[string]interface{}{}, // Include this field in expected
				},
			},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := flattenMachinePoolConfigsVsphere(tc.input)

			// Debugging output
			fmt.Printf("Expected: %+v\n", tc.expected)
			fmt.Printf("Result: %+v\n", result)

			assert.Equal(t, tc.expected, result, "Unexpected result in test case: %s", tc.name)
		})
	}
}
