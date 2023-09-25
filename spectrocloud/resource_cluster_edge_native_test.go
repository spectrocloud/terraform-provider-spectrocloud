package spectrocloud

import (
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"reflect"
	"testing"
)

func prepareEdgeNativeResourceData() *schema.ResourceData {
	// Create a mock resource data
	resourceData := resourceClusterEdgeNative().TestResourceData()
	resourceData.Set("name", "sample-cluster")
	resourceData.Set("context", "project")
	resourceData.Set("cloud_config", []map[string]interface{}{
		{
			"vip": "192.168.1.1",
		},
	})
	resourceData.Set("tags", []string{"test"})
	resourceData.Set("cluster_profile", "test-cluster-uid")
	resourceData.Set("apply_setting", "test-settings")
	resourceData.Set("cloud_account_id", "test-cloud-account-uid")
	resourceData.Set("os_patch_on_boot", false)
	machinePool := map[string]interface{}{
		"control_plane":           true,
		"control_plane_as_worker": true,
		"name":                    "sample-pool",
		"edge_host": []interface{}{
			map[string]interface{}{
				"host_uid":  "host1",
				"static_ip": "192.168.1.1",
			},
			map[string]interface{}{
				"host_uid":  "host2",
				"static_ip": "192.168.1.2",
			},
		},
	}
	mp := make([]interface{}, 0)
	mp = append(mp, machinePool)
	resourceData.Set("machine_pool", mp)
	return resourceData
}

func TestFlattenCloudConfigEdgeNative(t *testing.T) {
	// Create a mock resource data
	resourceData := prepareEdgeNativeResourceData()

	// Create a mock V1Client
	client := &client.V1Client{
		GetCloudConfigEdgeNativeFn: func(uid string, clusterContext string) (*models.V1EdgeNativeCloudConfig, error) {
			return &models.V1EdgeNativeCloudConfig{
				Metadata: &models.V1ObjectMeta{
					UID: "cloudconfiguid",
				},
				Spec: &models.V1EdgeNativeCloudConfigSpec{
					ClusterConfig: &models.V1EdgeNativeClusterConfig{
						ControlPlaneEndpoint: nil,
						NtpServers:           nil,
						SSHKeys:              nil,
						StaticIP:             false,
					},
					MachinePoolConfig: []*models.V1EdgeNativeMachinePoolConfig{
						{
							AdditionalLabels:        map[string]string{"unit-test": "label1"},
							Taints:                  nil,
							IsControlPlane:          true,
							UseControlPlaneAsWorker: false,
							Name:                    "sample-pool",
							Hosts: []*models.V1EdgeNativeHost{
								{
									HostUID:  ptrString("host1"),
									StaticIP: "192.168.1.1",
								},
								{
									HostUID:  ptrString("host2"),
									StaticIP: "192.168.1.2",
								},
							},
							UpdateStrategy: &models.V1UpdateStrategy{
								Type: "rolling_update",
							},
						},
					},
				},
				Status: nil,
			}, nil
		},
	}

	configUID := "sample-config-uid"

	diags := flattenCloudConfigEdgeNative(configUID, resourceData, client)

	// Check if there are any errors
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}

	// Check if resource data is correctly set
	if uid := resourceData.Get("cloud_config_id").(string); uid != configUID {
		t.Errorf("Expected cloud_config_id %s, got %s", configUID, uid)
	}
}

func TestToEdgeNativeCluster(t *testing.T) {
	m := &client.V1Client{
		GetClusterWithoutStatusFn: func(uid string) (*models.V1SpectroCluster, error) {
			if uid != "cluster-123" {
				return nil, errors.New("unexpected cluster_uid")
			}
			return &models.V1SpectroCluster{
				Metadata: nil,
				Spec:     nil,
				Status: &models.V1SpectroClusterStatus{
					State: "Deleted",
				},
			}, nil
		},
	}
	resourceData := prepareEdgeNativeResourceData()
	result, err := toEdgeNativeCluster(m, resourceData)

	// Check if there are any errors
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check if result is not nil
	if result == nil {
		t.Errorf("Expected non-nil result, got nil")
	}

}

func TestToEdgeHosts(t *testing.T) {
	// Create a sample input map
	inputMap := map[string]interface{}{
		"edge_host": []interface{}{
			map[string]interface{}{
				"host_uid":  "host1",
				"static_ip": "192.168.1.1",
			},
			map[string]interface{}{
				"host_uid":  "host2",
				"static_ip": "192.168.1.2",
			},
		},
	}

	expectedResult := &models.V1EdgeNativeMachinePoolCloudConfigEntity{
		EdgeHosts: []*models.V1EdgeNativeMachinePoolHostEntity{
			{
				HostUID:  ptrString("host1"),
				StaticIP: "192.168.1.1",
			},
			{
				HostUID:  ptrString("host2"),
				StaticIP: "192.168.1.2",
			},
		},
	}

	result := toEdgeHosts(inputMap)

	// Check if the result matches the expected output
	if !compareEdgeConfigEntities(result, expectedResult) {
		t.Errorf("Expected %+v but got %+v", expectedResult, result)
	}
}

func TestToMachinePoolEdgeNative(t *testing.T) {
	// Create a sample machinePool input
	machinePool := map[string]interface{}{
		"control_plane":           true,
		"control_plane_as_worker": true,
		"name":                    "sample-pool",
		"edge_host": []interface{}{
			map[string]interface{}{
				"host_uid":  "host1",
				"static_ip": "192.168.1.1",
			},
			map[string]interface{}{
				"host_uid":  "host2",
				"static_ip": "192.168.1.2",
			},
		},
	}

	expectedResult := &models.V1EdgeNativeMachinePoolConfigEntity{
		CloudConfig: toEdgeHosts(machinePool),
		PoolConfig: &models.V1MachinePoolConfigEntity{
			AdditionalLabels: map[string]string{},
			Taints:           nil,
			IsControlPlane:   true,
			Labels:           []string{},
			Name:             types.Ptr("sample-pool"),
			Size:             types.Ptr(int32(2)),
			UpdateStrategy: &models.V1UpdateStrategy{
				Type: "RollingUpdateScaleOut",
			},
			UseControlPlaneAsWorker: true,
		},
	}

	result, err := toMachinePoolEdgeNative(machinePool)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !compareMachinePoolEntities(result, expectedResult) {
		t.Errorf("Expected %+v but got %+v", expectedResult, result)
	}
}

func TestFlattenMachinePoolConfigsEdgeNative(t *testing.T) {
	// Create a sample input array of V1EdgeNativeMachinePoolConfig
	machinePools := []*models.V1EdgeNativeMachinePoolConfig{
		{
			AdditionalLabels:        map[string]string{"unit-test": "label1"},
			Taints:                  nil,
			IsControlPlane:          true,
			UseControlPlaneAsWorker: false,
			Name:                    "sample-pool",
			Hosts: []*models.V1EdgeNativeHost{
				{
					HostUID:  ptrString("host1"),
					StaticIP: "192.168.1.1",
				},
				{
					HostUID:  ptrString("host2"),
					StaticIP: "192.168.1.2",
				},
			},
			UpdateStrategy: &models.V1UpdateStrategy{
				Type: "rolling_update",
			},
		},
	}

	expectedResult := []interface{}{
		map[string]interface{}{
			"additional_labels": map[string]string{"unit-test": "label1"},
			//"taints":                  nil,
			"control_plane":           true,
			"control_plane_as_worker": false,
			"name":                    "sample-pool",
			"edge_host": []map[string]string{
				{
					"host_uid":  "host1",
					"static_ip": "192.168.1.1",
				},
				{
					"host_uid":  "host2",
					"static_ip": "192.168.1.2",
				},
			},
			"update_strategy": map[string]interface{}{
				"type": "rolling_update",
			},
		},
	}

	result := flattenMachinePoolConfigsEdgeNative(machinePools)

	// Compare the result with the expected output
	if !compareSlices(result, expectedResult) {
		t.Errorf("Expected %+v but got %+v", expectedResult, result)
	}
}

func compareSlices(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for i, itemA := range a {
		itemA := itemA.(map[string]interface{})
		itemB := b[i].(map[string]interface{})
		if !reflect.DeepEqual(itemA["additional_labels"], itemB["additional_labels"]) ||
			!reflect.DeepEqual(itemA["taints"], itemB["taints"]) || itemA["control_plane"] != itemB["control_plane"] ||
			!reflect.DeepEqual(itemA["edge_host"], itemB["edge_host"]) {
			return false
		}
	}

	return true
}

func compareMachinePoolEntities(a, b *models.V1EdgeNativeMachinePoolConfigEntity) bool {
	// Compare CloudConfig
	if !compareCloudConfigs(a.CloudConfig, b.CloudConfig) {
		return false
	}

	// Compare PoolConfig
	if !comparePoolConfigs(a.PoolConfig, b.PoolConfig) {
		return false
	}

	return true
}

func compareCloudConfigs(a, b *models.V1EdgeNativeMachinePoolCloudConfigEntity) bool {
	if !compareEdgeHosts(a.EdgeHosts, b.EdgeHosts) {
		return false
	}
	return true
}

func compareEdgeHosts(a, b []*models.V1EdgeNativeMachinePoolHostEntity) bool {
	if len(a) != len(b) {
		return false
	}

	for i, hostA := range a {
		hostB := b[i]
		if *hostA.HostUID != *hostB.HostUID || hostA.StaticIP != hostB.StaticIP {
			return false
		}
	}

	return true
}

func comparePoolConfigs(a, b *models.V1MachinePoolConfigEntity) bool {
	// Compare AdditionalLabels, Taints, IsControlPlane, Labels, Name, Size, UpdateStrategy, and UseControlPlaneAsWorker
	if !reflect.DeepEqual(a.AdditionalLabels, b.AdditionalLabels) ||
		!reflect.DeepEqual(a.Taints, b.Taints) ||
		a.IsControlPlane != b.IsControlPlane ||
		!reflect.DeepEqual(a.Labels, b.Labels) ||
		*a.Name != *b.Name ||
		*a.Size != *b.Size ||
		!reflect.DeepEqual(a.UpdateStrategy, b.UpdateStrategy) ||
		a.UseControlPlaneAsWorker != b.UseControlPlaneAsWorker {
		return false
	}

	return true
}

func ptrString(s string) *string {
	return &s
}

func compareEdgeConfigEntities(a, b *models.V1EdgeNativeMachinePoolCloudConfigEntity) bool {
	if len(a.EdgeHosts) != len(b.EdgeHosts) {
		return false
	}

	for i, hostA := range a.EdgeHosts {
		hostB := b.EdgeHosts[i]
		if *hostA.HostUID != *hostB.HostUID || hostA.StaticIP != hostB.StaticIP {
			return false
		}
	}

	return true
}
