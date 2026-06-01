package spectrocloud

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func TestFlattenMachinePoolConfigsMaas(t *testing.T) {
	t.Run("Nil Input", func(t *testing.T) {
		expected := make([]interface{}, 0)
		result := flattenMachinePoolConfigsMaas(nil, nil)

		if !reflect.DeepEqual(expected, result) {
			t.Errorf("Expected empty array for nil input, got: %v", result)
		}
	})

	t.Run("Valid Input worker defaults skip_k8s_upgrade", func(t *testing.T) {
		var mockMachinePools []*models.V1MaasMachinePoolConfig
		mp := &models.V1MaasMachinePoolConfig{
			AdditionalLabels: map[string]string{
				"TF": "test_label",
			},
			AdditionalTags: map[string]string{
				"TF": "test_tag",
			},
			Azs: []string{"zone1", "zone2"},
			InstanceType: &models.V1MaasInstanceType{
				MinCPU:     int32(2),
				MinMemInMB: int32(500),
			},
			IsControlPlane: false,
			Labels:         []string{"Masslabel1", "Masslabel2"},
			MachinePoolProperties: &models.V1MachinePoolProperties{
				ArchType: models.V1ArchTypeAmd64.Pointer(),
			},
			MaxSize:            3,
			MinSize:            2,
			Name:               "mass_mp_worker",
			NodeRepaveInterval: 30,
			ResourcePool:       "maas_resource",
			Size:               2,
			Tags:               []string{"test"},
			Taints:             nil,
			UpdateStrategy: &models.V1UpdateStrategy{
				Type: "RollingUpdateScaleOut",
			},
			UseControlPlaneAsWorker: true,
			UseLxdVM:                false,
		}
		mockMachinePools = append(mockMachinePools, mp)
		config := &models.V1MaasClusterConfig{
			Domain: types.Ptr("maas_resource_pool"),
		}

		expected := []interface{}{
			map[string]interface{}{
				"control_plane":   false,
				"name":            "mass_mp_worker",
				"count":           2,
				"update_strategy": "RollingUpdateScaleOut",
				"max":             3,
				"additional_labels": map[string]string{
					"TF": "test_label",
				},
				"additional_annotations":  map[string]interface{}{},
				"node_repave_interval":    int32(30),
				"skip_k8s_upgrade":        "disabled",
				"control_plane_as_worker": true,
				"min":                     2,
				"instance_type": []interface{}{
					map[string]interface{}{
						"min_memory_mb": 500,
						"min_cpu":       2,
					},
				},
				"azs":        []string{"zone1", "zone2"},
				"node_tags":  []string{"test"},
				"use_lxd_vm": false,
				"placement": []interface{}{
					map[string]interface{}{
						"resource_pool": "maas_resource",
					},
				},
			},
		}

		result := flattenMachinePoolConfigsMaas(mockMachinePools, config)

		if diff := cmp.Diff(expected, result); diff != "" {
			t.Errorf("Unexpected result (-want +got):\n%s", diff)
		}
	})

	t.Run("worker pool skip_k8s_upgrade from API", func(t *testing.T) {
		mp := &models.V1MaasMachinePoolConfig{
			Azs: []string{"z1"},
			InstanceType: &models.V1MaasInstanceType{
				MinCPU:     2,
				MinMemInMB: 4096,
			},
			IsControlPlane: false,
			Labels:         []string{"worker"},
			Name:           "w",
			Size:           1,
			MinSize:        1,
			MaxSize:        1,
			ResourcePool:   "rp",
			SkipK8sUpgrade: types.Ptr("enabled"),
			Taints:         nil,
			UpdateStrategy: &models.V1UpdateStrategy{Type: "RollingUpdateScaleOut"},
		}
		config := &models.V1MaasClusterConfig{Domain: types.Ptr("d")}
		out := flattenMachinePoolConfigsMaas([]*models.V1MaasMachinePoolConfig{mp}, config)
		if len(out) != 1 {
			t.Fatalf("expected 1 pool, got %d", len(out))
		}
		oi := out[0].(map[string]interface{})
		if oi["skip_k8s_upgrade"] != "enabled" {
			t.Errorf("skip_k8s_upgrade: want enabled, got %v", oi["skip_k8s_upgrade"])
		}
	})

	t.Run("control plane flatten still exposes skip_k8s_upgrade default", func(t *testing.T) {
		mp := &models.V1MaasMachinePoolConfig{
			Azs: []string{"z1"},
			InstanceType: &models.V1MaasInstanceType{
				MinCPU:     4,
				MinMemInMB: 8192,
			},
			IsControlPlane: true,
			Labels:         []string{"control-plane"},
			Name:           "master",
			Size:           3,
			MinSize:        3,
			MaxSize:        3,
			ResourcePool:   "rp",
			Taints:         nil,
			UpdateStrategy: &models.V1UpdateStrategy{Type: "RollingUpdateScaleOut"},
		}
		config := &models.V1MaasClusterConfig{Domain: types.Ptr("d")}
		out := flattenMachinePoolConfigsMaas([]*models.V1MaasMachinePoolConfig{mp}, config)
		oi := out[0].(map[string]interface{})
		if oi["skip_k8s_upgrade"] != "disabled" {
			t.Errorf("skip_k8s_upgrade: want disabled default, got %v", oi["skip_k8s_upgrade"])
		}
	})
}

func TestToMachinePoolMaas(t *testing.T) {
	t.Run("worker pool default skip_k8s_upgrade disabled", func(t *testing.T) {
		input := map[string]interface{}{
			"control_plane":   false,
			"name":            "mass_mp_worker",
			"count":           2,
			"update_strategy": "RollingUpdateScaleOut",
			"max":             3,
			"additional_labels": map[string]interface{}{
				"TF": "test_label",
			},
			"node_repave_interval":    30,
			"control_plane_as_worker": true,
			"min":                     2,
			"instance_type": []interface{}{
				map[string]interface{}{
					"min_memory_mb": 500,
					"min_cpu":       2,
				},
			},
			"placement": []interface{}{
				map[string]interface{}{
					"id":            "test_id",
					"resource_pool": "test_resource_pool",
				},
			},
			"azs":        schema.NewSet(schema.HashString, []interface{}{"zone1", "zone2"}),
			"node_tags":  schema.NewSet(schema.HashString, []interface{}{"test"}),
			"use_lxd_vm": false,
			"network": []interface{}{
				map[string]interface{}{
					"network_name":    "test_network",
					"parent_pool_uid": "test_pool_uid",
					"static_ip":       false,
				},
			},
		}
		rp := "test_resource_pool"
		size := int32(2)
		name := "mass_mp_worker"
		expectedMachinePool := &models.V1MaasMachinePoolConfigEntity{
			CloudConfig: &models.V1MaasMachinePoolCloudConfigEntity{
				Azs:          []string{"zone2", "zone1"},
				InstanceType: &models.V1MaasInstanceType{MinCPU: 2, MinMemInMB: 500},
				ResourcePool: &rp,
				Tags:         []string{"test"},
				UseLxdVM:     false,
				Network: &models.V1MaasNetworkConfigEntity{
					NetworkName:   types.Ptr("test_network"),
					ParentPoolUID: "test_pool_uid",
					StaticIP:      false,
				},
			},
			PoolConfig: &models.V1MachinePoolConfigEntity{
				AdditionalLabels:        map[string]string{"TF": "test_label"},
				AdditionalAnnotations:   map[string]string{},
				Labels:                  []string{"worker"},
				MaxSize:                 3,
				MinSize:                 2,
				Name:                    &name,
				NodeRepaveInterval:      30,
				Size:                    &size,
				SkipK8sUpgrade:          types.Ptr("disabled"),
				UpdateStrategy:          &models.V1UpdateStrategy{Type: "RollingUpdateScaleOut"},
				UseControlPlaneAsWorker: true,
			},
		}

		result, err := toMachinePoolMaas(input)

		if diff := cmp.Diff(expectedMachinePool, result); diff != "" {
			t.Errorf("Unexpected result (-want +got):\n%s", diff)
		}
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("Expected a non-nil result")
		}
	})

	t.Run("worker pool skip_k8s_upgrade enabled", func(t *testing.T) {
		input := map[string]interface{}{
			"control_plane":           false,
			"name":                    "worker-skew",
			"count":                   1,
			"update_strategy":         "RollingUpdateScaleOut",
			"max":                     1,
			"min":                     1,
			"node_repave_interval":    0,
			"control_plane_as_worker": false,
			"skip_k8s_upgrade":        "enabled",
			"instance_type": []interface{}{
				map[string]interface{}{
					"min_memory_mb": 4096,
					"min_cpu":       2,
				},
			},
			"placement": []interface{}{
				map[string]interface{}{
					"resource_pool": "pool",
				},
			},
			"azs":        schema.NewSet(schema.HashString, []interface{}{"az1"}),
			"node_tags":  schema.NewSet(schema.HashString, []interface{}{}),
			"use_lxd_vm": false,
			"network":    []interface{}{},
		}
		result, err := toMachinePoolMaas(input)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.PoolConfig.SkipK8sUpgrade == nil || *result.PoolConfig.SkipK8sUpgrade != "enabled" {
			t.Fatalf("SkipK8sUpgrade: want enabled, got %v", result.PoolConfig.SkipK8sUpgrade)
		}
	})

	t.Run("control plane does not set SkipK8sUpgrade on pool config", func(t *testing.T) {
		input := map[string]interface{}{
			"control_plane":           true,
			"name":                    "master-pool",
			"count":                   3,
			"update_strategy":         "RollingUpdateScaleOut",
			"max":                     3,
			"min":                     3,
			"node_repave_interval":    0,
			"control_plane_as_worker": false,
			"skip_k8s_upgrade":        "enabled",
			"instance_type": []interface{}{
				map[string]interface{}{
					"min_memory_mb": 8192,
					"min_cpu":       4,
				},
			},
			"placement": []interface{}{
				map[string]interface{}{
					"resource_pool": "pool",
				},
			},
			"azs":        schema.NewSet(schema.HashString, []interface{}{"az1"}),
			"node_tags":  schema.NewSet(schema.HashString, []interface{}{}),
			"use_lxd_vm": false,
			"network":    []interface{}{},
		}
		result, err := toMachinePoolMaas(input)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.PoolConfig.SkipK8sUpgrade != nil {
			t.Fatalf("control plane SkipK8sUpgrade should be unset, got %v", result.PoolConfig.SkipK8sUpgrade)
		}
	})
}

func TestFlattenClusterConfigsMaas(t *testing.T) {
	t.Run("Nil Input", func(t *testing.T) {
		d := resourceClusterMaas().TestResourceData()
		result := flattenClusterConfigsMaas(d, nil)
		expected := make([]interface{}, 0)

		if !reflect.DeepEqual(expected, result) {
			t.Errorf("Expected empty array for nil input, got: %v", result)
		}
	})

	t.Run("Nil Spec", func(t *testing.T) {
		d := resourceClusterMaas().TestResourceData()
		config := &models.V1MaasCloudConfig{}
		result := flattenClusterConfigsMaas(d, config)
		expected := make([]interface{}, 0)

		if !reflect.DeepEqual(expected, result) {
			t.Errorf("Expected empty array for nil spec, got: %v", result)
		}
	})

	t.Run("Nil ClusterConfig", func(t *testing.T) {
		d := resourceClusterMaas().TestResourceData()
		config := &models.V1MaasCloudConfig{
			Spec: &models.V1MaasCloudConfigSpec{},
		}
		result := flattenClusterConfigsMaas(d, config)
		expected := make([]interface{}, 0)

		if !reflect.DeepEqual(expected, result) {
			t.Errorf("Expected empty array for nil cluster config, got: %v", result)
		}
	})

	t.Run("Valid Input Without NtpServers", func(t *testing.T) {
		d := resourceClusterMaas().TestResourceData()
		domain := "test.maas.local"
		config := &models.V1MaasCloudConfig{
			Spec: &models.V1MaasCloudConfigSpec{
				ClusterConfig: &models.V1MaasClusterConfig{
					Domain:      &domain,
					EnableLxdVM: false,
				},
			},
		}

		result := flattenClusterConfigsMaas(d, config)
		resultMap := result[0].(map[string]interface{})

		assert.Equal(t, "test.maas.local", resultMap["domain"])
		// enable_lxd_vm is not set when false (relies on schema default)
		assert.Nil(t, resultMap["enable_lxd_vm"])
		assert.Nil(t, resultMap["ntp_servers"])
	})

	t.Run("Valid Input With NtpServers", func(t *testing.T) {
		d := resourceClusterMaas().TestResourceData()
		domain := "test.maas.local"
		ntpServers := []string{"0.pool.ntp.org", "1.pool.ntp.org", "time.google.com"}
		config := &models.V1MaasCloudConfig{
			Spec: &models.V1MaasCloudConfigSpec{
				ClusterConfig: &models.V1MaasClusterConfig{
					Domain:      &domain,
					EnableLxdVM: true,
					NtpServers:  ntpServers,
				},
			},
		}

		result := flattenClusterConfigsMaas(d, config)
		resultMap := result[0].(map[string]interface{})

		assert.Equal(t, "test.maas.local", resultMap["domain"])
		assert.Equal(t, true, resultMap["enable_lxd_vm"])
		assert.Equal(t, ntpServers, resultMap["ntp_servers"])
	})

	t.Run("Valid Input With Empty NtpServers", func(t *testing.T) {
		d := resourceClusterMaas().TestResourceData()
		domain := "test.maas.local"
		ntpServers := []string{}
		config := &models.V1MaasCloudConfig{
			Spec: &models.V1MaasCloudConfigSpec{
				ClusterConfig: &models.V1MaasClusterConfig{
					Domain:      &domain,
					EnableLxdVM: false,
					NtpServers:  ntpServers,
				},
			},
		}

		result := flattenClusterConfigsMaas(d, config)
		resultMap := result[0].(map[string]interface{})

		assert.Equal(t, "test.maas.local", resultMap["domain"])
		// enable_lxd_vm is not set when false (relies on schema default)
		assert.Nil(t, resultMap["enable_lxd_vm"])
		// Empty slice should be set
		assert.Equal(t, ntpServers, resultMap["ntp_servers"])
	})
}

func TestToMaasClusterWithNtpServers(t *testing.T) {
	t.Run("With NtpServers", func(t *testing.T) {
		d := resourceClusterMaas().TestResourceData()

		// Set basic required fields
		d.Set("name", "test-maas-cluster")
		d.Set("context", "project")
		d.Set("cloud_account_id", "test-account-id")

		// Set cluster profile
		cConfig := []map[string]interface{}{
			{"id": "test-profile-id"},
		}
		d.Set("cluster_profile", cConfig)

		// Set cloud config with NTP servers
		ntpServers := schema.NewSet(schema.HashString, []interface{}{
			"0.pool.ntp.org",
			"1.pool.ntp.org",
			"time.google.com",
		})
		cloudConfig := []map[string]interface{}{
			{
				"domain":        "test.maas.local",
				"enable_lxd_vm": false,
				"ntp_servers":   ntpServers,
			},
		}
		d.Set("cloud_config", cloudConfig)

		// Set machine pool
		machinePools := schema.NewSet(resourceMachinePoolMaasHash, []interface{}{
			map[string]interface{}{
				"name":                    "worker-pool",
				"count":                   3,
				"control_plane":           false,
				"control_plane_as_worker": false,
				"node_repave_interval":    0,
				"instance_type": []interface{}{
					map[string]interface{}{
						"min_cpu":       2,
						"min_memory_mb": 4096,
					},
				},
				"azs": schema.NewSet(schema.HashString, []interface{}{"zone1"}),
				"placement": []interface{}{
					map[string]interface{}{
						"resource_pool": "default",
					},
				},
				"node_tags": schema.NewSet(schema.HashString, []interface{}{}),
			},
		})
		d.Set("machine_pool", machinePools)

		// Mock client - pass nil for this unit test
		// In a real scenario, you'd mock the client properly
		result, err := toMaasCluster(nil, d)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Spec.CloudConfig)
		require.Len(t, result.Spec.Machinepoolconfig, 1)
		require.NotNil(t, result.Spec.Machinepoolconfig[0].PoolConfig)
		require.NotNil(t, result.Spec.Machinepoolconfig[0].PoolConfig.SkipK8sUpgrade)
		assert.Equal(t, "disabled", *result.Spec.Machinepoolconfig[0].PoolConfig.SkipK8sUpgrade)
		assert.NotNil(t, result.Spec.CloudConfig.NtpServers)
		assert.Equal(t, 3, len(result.Spec.CloudConfig.NtpServers))
		// Check that NTP servers are present (order may vary due to Set)
		ntpServerSlice := result.Spec.CloudConfig.NtpServers
		assert.Contains(t, ntpServerSlice, "0.pool.ntp.org")
		assert.Contains(t, ntpServerSlice, "1.pool.ntp.org")
		assert.Contains(t, ntpServerSlice, "time.google.com")
	})

	t.Run("Without NtpServers", func(t *testing.T) {
		d := resourceClusterMaas().TestResourceData()

		// Set basic required fields
		d.Set("name", "test-maas-cluster")
		d.Set("context", "project")
		d.Set("cloud_account_id", "test-account-id")

		// Set cluster profile
		cConfig := []map[string]interface{}{
			{"id": "test-profile-id"},
		}
		d.Set("cluster_profile", cConfig)

		// Set cloud config without NTP servers
		cloudConfig := []map[string]interface{}{
			{
				"domain":        "test.maas.local",
				"enable_lxd_vm": false,
			},
		}
		d.Set("cloud_config", cloudConfig)

		// Set machine pool
		machinePools := schema.NewSet(resourceMachinePoolMaasHash, []interface{}{
			map[string]interface{}{
				"name":                    "worker-pool",
				"count":                   3,
				"control_plane":           false,
				"control_plane_as_worker": false,
				"node_repave_interval":    0,
				"instance_type": []interface{}{
					map[string]interface{}{
						"min_cpu":       2,
						"min_memory_mb": 4096,
					},
				},
				"azs": schema.NewSet(schema.HashString, []interface{}{"zone1"}),
				"placement": []interface{}{
					map[string]interface{}{
						"resource_pool": "default",
					},
				},
				"node_tags": schema.NewSet(schema.HashString, []interface{}{}),
			},
		})
		d.Set("machine_pool", machinePools)

		// Mock client - pass nil for this unit test
		result, err := toMaasCluster(nil, d)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Spec.CloudConfig)
		require.Len(t, result.Spec.Machinepoolconfig, 1)
		require.NotNil(t, result.Spec.Machinepoolconfig[0].PoolConfig)
		require.NotNil(t, result.Spec.Machinepoolconfig[0].PoolConfig.SkipK8sUpgrade)
		assert.Equal(t, "disabled", *result.Spec.Machinepoolconfig[0].PoolConfig.SkipK8sUpgrade)
		// NtpServers should be empty slice when not provided
		assert.Equal(t, 0, len(result.Spec.CloudConfig.NtpServers))
	})
}

func TestToMaasCloudConfigUpdate(t *testing.T) {
	t.Run("Valid cloud config update with NTP servers", func(t *testing.T) {
		cloudConfig := map[string]interface{}{
			"domain":        "test.maas.local",
			"enable_lxd_vm": false,
			"ntp_servers": schema.NewSet(schema.HashString, []interface{}{
				"0.pool.ntp.org",
				"1.pool.ntp.org",
			}),
		}

		result := toMaasCloudConfigUpdate(cloudConfig)

		assert.NotNil(t, result)
		assert.NotNil(t, result.ClusterConfig)
		assert.Equal(t, "test.maas.local", *result.ClusterConfig.Domain)
		assert.Equal(t, false, result.ClusterConfig.EnableLxdVM)
		assert.Equal(t, 2, len(result.ClusterConfig.NtpServers))
		assert.Contains(t, result.ClusterConfig.NtpServers, "0.pool.ntp.org")
		assert.Contains(t, result.ClusterConfig.NtpServers, "1.pool.ntp.org")
	})

	t.Run("Valid cloud config update without NTP servers", func(t *testing.T) {
		cloudConfig := map[string]interface{}{
			"domain":        "test.maas.local",
			"enable_lxd_vm": true,
		}

		result := toMaasCloudConfigUpdate(cloudConfig)

		assert.NotNil(t, result)
		assert.NotNil(t, result.ClusterConfig)
		assert.Equal(t, "test.maas.local", *result.ClusterConfig.Domain)
		assert.Equal(t, true, result.ClusterConfig.EnableLxdVM)
		assert.Equal(t, 0, len(result.ClusterConfig.NtpServers))
	})

	t.Run("Update NTP servers only", func(t *testing.T) {
		cloudConfig := map[string]interface{}{
			"domain":        "production.maas.local",
			"enable_lxd_vm": false,
			"ntp_servers": schema.NewSet(schema.HashString, []interface{}{
				"time1.google.com",
				"time2.google.com",
				"time3.google.com",
			}),
		}

		result := toMaasCloudConfigUpdate(cloudConfig)

		assert.NotNil(t, result)
		assert.NotNil(t, result.ClusterConfig)
		assert.Equal(t, 3, len(result.ClusterConfig.NtpServers))
		assert.Contains(t, result.ClusterConfig.NtpServers, "time1.google.com")
		assert.Contains(t, result.ClusterConfig.NtpServers, "time2.google.com")
		assert.Contains(t, result.ClusterConfig.NtpServers, "time3.google.com")
	})
}
