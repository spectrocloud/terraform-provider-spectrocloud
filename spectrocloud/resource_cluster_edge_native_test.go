package spectrocloud

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spectrocloud/hapi/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func TestToEdgeHosts(t *testing.T) {
	hostUI1 := "uid1"
	hostUI2 := "uid2"

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1EdgeNativeMachinePoolCloudConfigEntity
	}{
		{
			name:     "Empty edge_host",
			input:    map[string]interface{}{"edge_host": []interface{}{}},
			expected: nil,
		},
		{
			name: "Valid edge_host",
			input: map[string]interface{}{
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
			},
			expected: &models.V1EdgeNativeMachinePoolCloudConfigEntity{
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
			},
		},
		{
			name: "Edge_host with empty host_name",
			input: map[string]interface{}{
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
			},
			expected: &models.V1EdgeNativeMachinePoolCloudConfigEntity{
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toEdgeHosts(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestToMachinePoolEdgeNative(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1EdgeNativeMachinePoolConfigEntity
		hasError bool
	}{
		{
			name: "Valid input data",
			input: map[string]interface{}{
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
			},
			expected: nil,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := &models.V1EdgeNativeMachinePoolConfigEntity{
				CloudConfig: toEdgeHosts(tt.input),
				PoolConfig: &models.V1MachinePoolConfigEntity{
					AdditionalLabels:        toAdditionalNodePoolLabels(tt.input),
					Taints:                  toClusterTaints(tt.input),
					IsControlPlane:          true,
					Labels:                  []string{},
					Name:                    types.Ptr("pool1"),
					Size:                    types.Ptr(int32(len(toEdgeHosts(tt.input).EdgeHosts))),
					UpdateStrategy:          &models.V1UpdateStrategy{Type: getUpdateStrategy(tt.input)},
					UseControlPlaneAsWorker: false,
				},
			}

			result, err := toMachinePoolEdgeNative(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected an error but got none.")
				}
				return
			}
			if err != nil {
				t.Fatalf("Expected no error, but got an error: %v", err)
			}
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("Expected %v, got %v", expected, result)
			}
		})
	}
}

func TestFlattenMachinePoolConfigsEdgeNative(t *testing.T) {
	hui1 := "uid1"
	huid2 := "uid2"
	huid3 := "uid3"

	tests := []struct {
		name     string
		input    []*models.V1EdgeNativeMachinePoolConfig
		expected []interface{}
	}{
		{
			name:     "When 'machinePools' is nil",
			input:    nil,
			expected: []interface{}{},
		},
		{
			name: "When 'machinePools' contains valid data",
			input: []*models.V1EdgeNativeMachinePoolConfig{
				{
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
				},
				{
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
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"additional_labels":       map[string]string{"label1": "value1"},
					"control_plane_as_worker": false,
					"control_plane":           false,
					"node_repave_interval":    int32(0),
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
					"update_strategy": "strategy1",
				},
				map[string]interface{}{
					"additional_labels":       map[string]string{"label2": "value2"},
					"control_plane_as_worker": true,
					"control_plane":           false,
					"node_repave_interval":    int32(0),
					"name":                    "pool2",
					"edge_host": []map[string]string{
						{
							"host_name": "host3",
							"host_uid":  "uid3",
							"static_ip": "ip3",
						},
					},
					"update_strategy": "strategy2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenMachinePoolConfigsEdgeNative(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %v, got %v", len(tt.expected), len(result))
				return
			}

			for i, expectedMap := range tt.expected {
				if diff := cmp.Diff(expectedMap, result[i]); diff != "" {
					t.Errorf("Test %s failed for item %d. Mismatch (-expected +actual):\n%s", tt.name, i, diff)
				}
			}
		})
	}
}

func TestValidationNodeRepaveIntervalForControlPlane(t *testing.T) {
	tests := []struct {
		name            string
		nodeRepaveValue int
		hasError        bool
	}{
		{
			name:            "Zero node repave interval for control plane",
			nodeRepaveValue: 0,
			hasError:        false,
		},
		{
			name:            "Non-zero node repave interval for control plane",
			nodeRepaveValue: 10,
			hasError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidationNodeRepaveIntervalForControlPlane(tt.nodeRepaveValue)
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected an error but got none.")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got an error: %v", err)
				}
			}
		})
	}
}
