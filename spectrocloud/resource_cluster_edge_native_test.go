package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"reflect"
	"testing"
)

func prepareEdgeNativeTestData() *schema.ResourceData {
	// Create a TestResourceData object for testing
	d := resourceClusterEdgeNative().TestResourceData()
	d.Set("name", "cluster-1")
	d.Set("context", "project")
	d.Set("tags", []string{"tag1:value1", "tag2:value2"})
	d.Set("apply_setting", "apply_setting_value")
	d.Set("cloud_account_id", "cloud_account_id_value")
	d.Set("os_patch_on_boot", true)
	d.Set("os_patch_schedule", "0 0 * * *")
	d.Set("os_patch_after", "2023-01-01T00:00:00Z")
	mp := map[string]interface{}{
		"name":                    "pool-1",
		"additional_labels":       map[string]interface{}{"label1": "value1"},
		"control_plane":           true,
		"control_plane_as_worker": true,
		"update_strategy":         "RollingUpdateScaleOut",
		"edge_host": []map[string]interface{}{
			{
				"host_name": "host-1",
				"host_uid":  "uid-1",
				"static_ip": "ip-1",
			},
		},
	}

	d.Set("machine_pool", mp)

	return d
}

func TestToEdgeHosts(t *testing.T) {

	hostUI1 := "uid1"
	hostUI2 := "uid2"
	// Test case 1: When 'edge_host' is an empty slice, the function should return nil
	result := toEdgeHosts(map[string]interface{}{"edge_host": []interface{}{}})
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}

	// Test case 2: When 'edge_host' contains valid data, the function should return the expected result
	input := map[string]interface{}{
		"edge_host": []interface{}{
			map[string]interface{}{
				"host_name": "host1",
				"host_uid":  "uid1",
				"static_ip": "ip1",
			},
			map[string]interface{}{
				"host_name": "host2",
				"host_uid":  "uid2",
				"static_ip": "ip2",
			},
		},
	}
	expected := &models.V1EdgeNativeMachinePoolCloudConfigEntity{
		EdgeHosts: []*models.V1EdgeNativeMachinePoolHostEntity{
			{
				HostName: "host1",
				HostUID:  &hostUI1,
				StaticIP: "ip1",
			},
			{
				HostName: "host2",
				HostUID:  &hostUI2,
				StaticIP: "ip2",
			},
		},
	}

	result = toEdgeHosts(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test case 3: When 'edge_host' contains valid data with host_name as empty string, the function should return the expected result
	input = map[string]interface{}{
		"edge_host": []interface{}{
			map[string]interface{}{
				"host_name": "",
				"host_uid":  "uid1",
				"static_ip": "ip1",
			},
			map[string]interface{}{
				"host_name": "",
				"host_uid":  "uid2",
				"static_ip": "ip2",
			},
		},
	}
	expected = &models.V1EdgeNativeMachinePoolCloudConfigEntity{
		EdgeHosts: []*models.V1EdgeNativeMachinePoolHostEntity{
			{
				HostName: "",
				HostUID:  &hostUI1,
				StaticIP: "ip1",
			},
			{
				HostName: "",
				HostUID:  &hostUI2,
				StaticIP: "ip2",
			},
		},
	}

	result = toEdgeHosts(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestToMachinePoolEdgeNative(t *testing.T) {
	// Test case 1: Valid input data
	input := map[string]interface{}{
		"control_plane":           true,
		"control_plane_as_worker": false,
		"name":                    "pool1",
		"edge_host": []interface{}{
			map[string]interface{}{
				"host_name": "",
				"host_uid":  "uid1",
				"static_ip": "ip1",
			},
			map[string]interface{}{
				"host_name": "",
				"host_uid":  "uid2",
				"static_ip": "ip2",
			},
		},
		"additional_labels": map[string]interface{}{
			"label1": "value1",
			"label2": "value2",
		},
		"taints": []interface{}{
			map[string]interface{}{
				"key":    "key1",
				"value":  "value1",
				"effect": "NoSchedule",
			},
			map[string]interface{}{
				"key":    "key2",
				"value":  "value2",
				"effect": "PreferNoSchedule",
			},
		},

		// Add other relevant fields as needed for your test
	}

	expected := &models.V1EdgeNativeMachinePoolConfigEntity{
		CloudConfig: toEdgeHosts(input),
		PoolConfig: &models.V1MachinePoolConfigEntity{
			AdditionalLabels:        toAdditionalNodePoolLabels(input),
			Taints:                  toClusterTaints(input),
			IsControlPlane:          true,
			Labels:                  []string{},
			Name:                    types.Ptr("pool1"),
			Size:                    types.Ptr(int32(len(toEdgeHosts(input).EdgeHosts))),
			UpdateStrategy:          &models.V1UpdateStrategy{Type: getUpdateStrategy(input)},
			UseControlPlaneAsWorker: false,
		},
	}

	result, err := toMachinePoolEdgeNative(input)
	if err != nil {
		t.Fatalf("Expected no error, but got an error: %v", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

}

func TestFlattenMachinePoolConfigsEdgeNative(t *testing.T) {
	// Test case 1: When 'machinePools' is nil, the function should return an empty slice
	result := flattenMachinePoolConfigsEdgeNative(nil)
	if len(result) != 0 {
		t.Errorf("Expected an empty slice, got %v", result)
	}

	// Test case 2: When 'machinePools' contains valid data, the function should return the expected result
	hui1 := "uid1"
	huid2 := "uid2"
	huid3 := "uid3"
	machinePool1 := &models.V1EdgeNativeMachinePoolConfig{
		AdditionalLabels:        map[string]string{"label1": "value1"},
		Taints:                  []*models.V1Taint{},
		UseControlPlaneAsWorker: false,
		Name:                    "pool1",
		Hosts: []*models.V1EdgeNativeHost{
			{
				HostName: "host1",
				HostUID:  &hui1,
				StaticIP: "ip1",
			},
			{
				HostName: "host2",
				HostUID:  &huid2,
				StaticIP: "ip2",
			},
		},
		UpdateStrategy: &models.V1UpdateStrategy{Type: "strategy1"},
	}

	machinePool2 := &models.V1EdgeNativeMachinePoolConfig{
		AdditionalLabels:        map[string]string{"label2": "value2"},
		Taints:                  []*models.V1Taint{},
		UseControlPlaneAsWorker: true,
		Name:                    "pool2",
		Hosts: []*models.V1EdgeNativeHost{
			{
				HostName: "host3",
				HostUID:  &huid3,
				StaticIP: "ip3",
			},
		},
		UpdateStrategy: &models.V1UpdateStrategy{Type: "strategy2"},
	}

	machinePools := []*models.V1EdgeNativeMachinePoolConfig{machinePool1, machinePool2}

	expected := []interface{}{
		map[string]interface{}{
			"additional_labels":       map[string]interface{}{"label1": "value1"},
			"control_plane_as_worker": false,
			"name":                    "pool1",
			"edge_host": []map[string]string{
				{
					"host_name": "host1",
					"host_uid":  "uid1",
					"static_ip": "ip1",
				},
				{
					"host_name": "host2",
					"host_uid":  "uid2",
					"static_ip": "ip2",
				},
			},
			"update_strategy": map[string]interface{}{"type": "strategy1"},
		},
		map[string]interface{}{
			"additional_labels":       map[string]interface{}{"label2": "value2"},
			"control_plane_as_worker": true,
			"name":                    "pool2",
			"edge_host": []map[string]string{
				{
					"host_name": "host3",
					"host_uid":  "uid3",
					"static_ip": "ip3",
				},
			},
			"update_strategy": map[string]interface{}{"type": "strategy2"},
		},
	}

	result = flattenMachinePoolConfigsEdgeNative(machinePools)
	if len(result) != len(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
	for i, expectedOi := range expected {
		if !mapsAreEqual(result[i].(map[string]interface{}), expectedOi.(map[string]interface{})) {
			t.Errorf("Expected %v, got %v", expectedOi, result[i])
		}
	}
}
