package spectrocloud

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func prepareClusterVsphereTestResourceData() *schema.ResourceData {
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
	d := prepareClusterVsphereTestResourceData()
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
	d := prepareClusterVsphereTestResourceData()
	flatCloudConfig := flattenClusterConfigsVsphere(d, nil)
	if flatCloudConfig == nil {
		t.Errorf("flattenClusterConfigsVsphere returning value for nill: %#v", flatCloudConfig)
	}
}

// TestFlattenClusterConfigsVsphereNilClusterConfig exercises the
// "cloudConfig.Spec.ClusterConfig == nil" branch — distinct from the
// whole-cloudConfig-nil branch covered above.
func TestFlattenClusterConfigsVsphereNilClusterConfig(t *testing.T) {
	d := prepareClusterVsphereTestResourceData()
	cloudConfig := &models.V1VsphereCloudConfig{
		Spec: &models.V1VsphereCloudConfigSpec{ClusterConfig: nil},
	}
	flat := flattenClusterConfigsVsphere(d, cloudConfig)
	assert.Equal(t, make([]interface{}, 0), flat)
}

// TestFlattenClusterConfigsVsphereOverrideClusterAPIConfig exercises the
// OverrideClusterAPIConfig-populated branch inside flattenClusterConfigsVsphere.
func TestFlattenClusterConfigsVsphereOverrideClusterAPIConfig(t *testing.T) {
	d := prepareClusterVsphereTestResourceData()
	yaml := "VSphereCluster:\n  spec:\n    identityRef:\n      kind: VSphereClusterIdentity\n"
	cloudConfig := &models.V1VsphereCloudConfig{
		Spec: &models.V1VsphereCloudConfigSpec{
			ClusterConfig: &models.V1VsphereClusterConfig{
				Placement:                &models.V1VspherePlacementConfig{Datacenter: "dc1", Folder: "f1"},
				OverrideClusterAPIConfig: yaml,
			},
		},
	}
	flat := flattenClusterConfigsVsphere(d, cloudConfig)
	flatMap := flat.([]interface{})[0].(map[string]interface{})
	assert.Equal(t, yaml, flatMap["override_cluster_api_config"])
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
					IsControlPlane:          types.Ptr(true),
					UseControlPlaneAsWorker: false,
					NodeRepaveInterval:      int32(24),
					UpdateStrategy: &models.V1UpdateStrategy{
						Type: "RollingUpdate",
					},
					InstanceType: &models.V1VsphereInstanceType{
						DiskGiB:   types.Ptr(int32(100)),
						MemoryMiB: types.Ptr(int64(8192)),
						NumCPUs:   types.Ptr(int32(4)),
					},
					Placements: []*models.V1VspherePlacementConfig{
						{
							UID:          "placement1",
							Cluster:      "cluster1",
							ResourcePool: "resource-pool1",
							Datastore:    "datastore1",
							Network: &models.V1VsphereNetworkConfig{
								NetworkName: types.Ptr("network1"),
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
							"network":           types.Ptr("network1"), // Handle pointer or use (*string)(nil) if necessary
							"static_ip_pool_id": "pool1",
						},
					},
					"update_strategy":        "RollingUpdate", // Include this field in expected
					"skip_k8s_upgrade":       "disabled",
					"additional_labels":      map[string]interface{}{}, // Include this field in expected
					"additional_annotations": map[string]interface{}{}, // Include this field in expected
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

func TestFlattenMachinePoolConfigsVsphereSkipK8sUpgrade(t *testing.T) {
	mp := &models.V1VsphereMachinePoolConfig{
		Name:           "worker-pool",
		Size:           1,
		MinSize:        1,
		MaxSize:        1,
		IsControlPlane: types.Ptr(false),
		Labels:         []string{"worker"},
		SkipK8sUpgrade: types.Ptr("enabled"),
		InstanceType: &models.V1VsphereInstanceType{
			DiskGiB:   types.Ptr(int32(50)),
			MemoryMiB: types.Ptr(int64(4096)),
			NumCPUs:   types.Ptr(int32(2)),
		},
		Placements: []*models.V1VspherePlacementConfig{
			{
				UID:          "p1",
				Cluster:      "cl",
				ResourcePool: "rp",
				Datastore:    "ds",
				Network: &models.V1VsphereNetworkConfig{
					NetworkName: types.Ptr("net"),
				},
			},
		},
		UpdateStrategy: &models.V1UpdateStrategy{Type: "RollingUpdateScaleOut"},
	}
	out := flattenMachinePoolConfigsVsphere([]*models.V1VsphereMachinePoolConfig{mp})
	require.Len(t, out, 1)
	oi := out[0].(map[string]interface{})
	assert.Equal(t, "enabled", oi["skip_k8s_upgrade"])
}

func TestToMachinePoolVsphereSkipK8sUpgrade(t *testing.T) {
	basePlacement := []interface{}{
		map[string]interface{}{
			"id":                "",
			"cluster":           "cl",
			"resource_pool":     "rp",
			"datastore":         "ds",
			"network":           "net",
			"static_ip_pool_id": "",
		},
	}
	baseInstance := []interface{}{
		map[string]interface{}{
			"disk_size_gb": 50,
			"memory_mb":    4096,
			"cpu":          2,
		},
	}

	t.Run("worker sets SkipK8sUpgrade enabled", func(t *testing.T) {
		m := map[string]interface{}{
			"control_plane":           false,
			"control_plane_as_worker": false,
			"name":                    "w",
			"count":                   1,
			"min":                     1,
			"max":                     1,
			"node_repave_interval":    0,
			"update_strategy":         "RollingUpdateScaleOut",
			"skip_k8s_upgrade":        "enabled",
			"placement":               basePlacement,
			"instance_type":           baseInstance,
		}
		mp, err := toMachinePoolVsphere(m)
		require.NoError(t, err)
		require.NotNil(t, mp.PoolConfig.SkipK8sUpgrade)
		assert.Equal(t, "enabled", *mp.PoolConfig.SkipK8sUpgrade)
	})

	t.Run("worker defaults SkipK8sUpgrade to disabled", func(t *testing.T) {
		m := map[string]interface{}{
			"control_plane":           false,
			"control_plane_as_worker": false,
			"name":                    "w",
			"count":                   1,
			"min":                     1,
			"max":                     1,
			"node_repave_interval":    0,
			"update_strategy":         "RollingUpdateScaleOut",
			"placement":               basePlacement,
			"instance_type":           baseInstance,
		}
		mp, err := toMachinePoolVsphere(m)
		require.NoError(t, err)
		require.NotNil(t, mp.PoolConfig.SkipK8sUpgrade)
		assert.Equal(t, "disabled", *mp.PoolConfig.SkipK8sUpgrade)
	})

	t.Run("control plane does not set SkipK8sUpgrade", func(t *testing.T) {
		m := map[string]interface{}{
			"control_plane":           true,
			"control_plane_as_worker": false,
			"name":                    "cp",
			"count":                   3,
			"min":                     3,
			"max":                     3,
			"node_repave_interval":    0,
			"update_strategy":         "RollingUpdateScaleOut",
			"skip_k8s_upgrade":        "enabled",
			"placement":               basePlacement,
			"instance_type":           baseInstance,
		}
		mp, err := toMachinePoolVsphere(m)
		require.NoError(t, err)
		assert.Nil(t, mp.PoolConfig.SkipK8sUpgrade)
	})
}

func TestFlattenMachinePoolConfigsVsphereOverrideClusterAPIConfig(t *testing.T) {
	yaml := "VSphereMachineTemplate:\n  spec:\n    template:\n      spec:\n        diskGiB: 80\n"
	mp := &models.V1VsphereMachinePoolConfig{
		Name:                     "worker-pool",
		Size:                     1,
		MinSize:                  1,
		MaxSize:                  1,
		IsControlPlane:           types.Ptr(false),
		Labels:                   []string{"worker"},
		OverrideClusterAPIConfig: yaml,
		InstanceType: &models.V1VsphereInstanceType{
			DiskGiB:   types.Ptr(int32(50)),
			MemoryMiB: types.Ptr(int64(4096)),
			NumCPUs:   types.Ptr(int32(2)),
		},
		Placements: []*models.V1VspherePlacementConfig{{
			UID: "p1", Cluster: "cl", ResourcePool: "rp", Datastore: "ds",
			Network: &models.V1VsphereNetworkConfig{NetworkName: types.Ptr("net")},
		}},
		UpdateStrategy: &models.V1UpdateStrategy{Type: "RollingUpdateScaleOut"},
	}
	out := flattenMachinePoolConfigsVsphere([]*models.V1VsphereMachinePoolConfig{mp})
	require.Len(t, out, 1)
	assert.Equal(t, yaml, out[0].(map[string]interface{})["override_cluster_api_config"])
}

func TestFlattenMachinePoolConfigsVsphereOverrideKubeadmConfiguration(t *testing.T) {
	kubeadmYaml := "kubeletExtraArgs:\n  v: \"4\"\n"
	mp := &models.V1VsphereMachinePoolConfig{
		Name:                         "worker-pool",
		Size:                         1,
		MinSize:                      1,
		MaxSize:                      1,
		IsControlPlane:               types.Ptr(false),
		Labels:                       []string{"worker"},
		OverrideKubeadmConfiguration: kubeadmYaml,
		InstanceType: &models.V1VsphereInstanceType{
			DiskGiB:   types.Ptr(int32(50)),
			MemoryMiB: types.Ptr(int64(4096)),
			NumCPUs:   types.Ptr(int32(2)),
		},
		Placements: []*models.V1VspherePlacementConfig{{
			UID: "p1", Cluster: "cl", ResourcePool: "rp", Datastore: "ds",
			Network: &models.V1VsphereNetworkConfig{NetworkName: types.Ptr("net")},
		}},
		UpdateStrategy: &models.V1UpdateStrategy{Type: "RollingUpdateScaleOut"},
	}
	out := flattenMachinePoolConfigsVsphere([]*models.V1VsphereMachinePoolConfig{mp})
	require.Len(t, out, 1)
	assert.Equal(t, kubeadmYaml, out[0].(map[string]interface{})["override_kubeadm_configuration"])
}

func TestToMachinePoolVsphereOverrideClusterAPIConfig(t *testing.T) {
	yaml := "VSphereCluster:\n  spec:\n    identityRef:\n      kind: VSphereClusterIdentity\n      name: my-identity\n"
	makeInput := func(controlPlane bool, override string) map[string]interface{} {
		return map[string]interface{}{
			"control_plane":           controlPlane,
			"control_plane_as_worker": false,
			"name":                    "pool",
			"count":                   1,
			"min":                     1,
			"max":                     1,
			"node_repave_interval":    0,
			"update_strategy":         "RollingUpdateScaleOut",
			"instance_type": []interface{}{
				map[string]interface{}{"disk_size_gb": 50, "memory_mb": 4096, "cpu": 2},
			},
			"placement": []interface{}{
				map[string]interface{}{
					"id": "", "cluster": "cl", "resource_pool": "rp",
					"datastore": "ds", "network": "net", "static_ip_pool_id": "",
				},
			},
			"override_cluster_api_config": override,
		}
	}

	t.Run("worker sets OverrideClusterAPIConfig", func(t *testing.T) {
		mp, err := toMachinePoolVsphere(makeInput(false, yaml))
		require.NoError(t, err)
		assert.Equal(t, yaml, mp.PoolConfig.OverrideClusterAPIConfig)
	})

	t.Run("control plane sets OverrideClusterAPIConfig", func(t *testing.T) {
		mp, err := toMachinePoolVsphere(makeInput(true, yaml))
		require.NoError(t, err)
		assert.Equal(t, yaml, mp.PoolConfig.OverrideClusterAPIConfig)
	})

	t.Run("empty passthrough leaves the field zero-valued", func(t *testing.T) {
		mp, err := toMachinePoolVsphere(makeInput(false, ""))
		require.NoError(t, err)
		assert.Empty(t, mp.PoolConfig.OverrideClusterAPIConfig)
	})
}

// TestToMachinePoolVsphereBoundsValidation exercises every negative/overflow
// guard clause in toMachinePoolVsphere. Each case starts from a valid worker
// (or control-plane) pool and overrides a single numeric field to trigger one
// specific error branch.
func TestToMachinePoolVsphereBoundsValidation(t *testing.T) {
	basePlacement := []interface{}{
		map[string]interface{}{
			"id": "", "cluster": "cl", "resource_pool": "rp",
			"datastore": "ds", "network": "net", "static_ip_pool_id": "",
		},
	}
	makeInput := func(controlPlane bool, overrides map[string]interface{}) map[string]interface{} {
		m := map[string]interface{}{
			"control_plane":           controlPlane,
			"control_plane_as_worker": false,
			"name":                    "pool",
			"count":                   1,
			"min":                     1,
			"max":                     1,
			"node_repave_interval":    0,
			"update_strategy":         "RollingUpdateScaleOut",
			"instance_type": []interface{}{
				map[string]interface{}{"disk_size_gb": 50, "memory_mb": 4096, "cpu": 2},
			},
			"placement": basePlacement,
		}
		for k, v := range overrides {
			m[k] = v
		}
		return m
	}

	t.Run("negative disk_size_gb errors", func(t *testing.T) {
		m := makeInput(false, nil)
		m["instance_type"] = []interface{}{map[string]interface{}{"disk_size_gb": -1, "memory_mb": 4096, "cpu": 2}}
		_, err := toMachinePoolVsphere(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be negative")
	})

	t.Run("negative memory_mb errors", func(t *testing.T) {
		m := makeInput(false, nil)
		m["instance_type"] = []interface{}{map[string]interface{}{"disk_size_gb": 50, "memory_mb": -1, "cpu": 2}}
		_, err := toMachinePoolVsphere(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be negative")
	})

	t.Run("negative cpu errors", func(t *testing.T) {
		m := makeInput(false, nil)
		m["instance_type"] = []interface{}{map[string]interface{}{"disk_size_gb": 50, "memory_mb": 4096, "cpu": -1}}
		_, err := toMachinePoolVsphere(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be negative")
	})

	t.Run("disk_size_gb out of int32 range errors", func(t *testing.T) {
		m := makeInput(false, nil)
		m["instance_type"] = []interface{}{map[string]interface{}{"disk_size_gb": math.MaxInt32 + 1, "memory_mb": 4096, "cpu": 2}}
		_, err := toMachinePoolVsphere(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "out of range")
	})

	t.Run("cpu out of int32 range errors", func(t *testing.T) {
		m := makeInput(false, nil)
		m["instance_type"] = []interface{}{map[string]interface{}{"disk_size_gb": 50, "memory_mb": 4096, "cpu": math.MaxInt32 + 1}}
		_, err := toMachinePoolVsphere(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "out of range")
	})

	t.Run("negative count errors", func(t *testing.T) {
		m := makeInput(false, map[string]interface{}{"count": -1})
		_, err := toMachinePoolVsphere(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "count value")
	})

	t.Run("count out of int32 range errors", func(t *testing.T) {
		m := makeInput(false, map[string]interface{}{"count": math.MaxInt32 + 1})
		_, err := toMachinePoolVsphere(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "out of range for int32")
	})

	t.Run("negative min errors", func(t *testing.T) {
		m := makeInput(false, map[string]interface{}{"min": -1})
		_, err := toMachinePoolVsphere(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "min value")
	})

	t.Run("min out of int32 range errors", func(t *testing.T) {
		m := makeInput(false, map[string]interface{}{"min": math.MaxInt32 + 1})
		_, err := toMachinePoolVsphere(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "min value")
	})

	t.Run("negative max errors", func(t *testing.T) {
		m := makeInput(false, map[string]interface{}{"max": -1})
		_, err := toMachinePoolVsphere(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "max value")
	})

	t.Run("max out of int32 range errors", func(t *testing.T) {
		m := makeInput(false, map[string]interface{}{"max": math.MaxInt32 + 1})
		_, err := toMachinePoolVsphere(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "max value")
	})

	t.Run("worker override_kubeadm_configuration is set", func(t *testing.T) {
		m := makeInput(false, map[string]interface{}{"override_kubeadm_configuration": "kind: X"})
		mp, err := toMachinePoolVsphere(m)
		require.NoError(t, err)
		assert.Equal(t, "kind: X", mp.PoolConfig.OverrideKubeadmConfiguration)
	})

	t.Run("negative node_repave_interval errors for worker", func(t *testing.T) {
		m := makeInput(false, map[string]interface{}{"node_repave_interval": -1})
		_, err := toMachinePoolVsphere(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "node_repave_interval")
	})

	t.Run("node_repave_interval out of int32 range errors for worker", func(t *testing.T) {
		m := makeInput(false, map[string]interface{}{"node_repave_interval": math.MaxInt32 + 1})
		_, err := toMachinePoolVsphere(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "node_repave_interval")
	})

	t.Run("negative node_repave_interval errors for control plane", func(t *testing.T) {
		m := makeInput(true, map[string]interface{}{"node_repave_interval": -1})
		_, err := toMachinePoolVsphere(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "node_repave_interval")
	})

	t.Run("node_repave_interval out of int32 range errors for control plane", func(t *testing.T) {
		m := makeInput(true, map[string]interface{}{"node_repave_interval": math.MaxInt32 + 1})
		_, err := toMachinePoolVsphere(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "node_repave_interval")
	})
}
