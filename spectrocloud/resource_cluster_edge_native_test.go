package spectrocloud

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/stretchr/testify/assert"

	"github.com/google/go-cmp/cmp"

	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func ToSchemaSetFromStrings(strings []string) *schema.Set {
	set := schema.NewSet(schema.HashString, nil)
	for _, v := range strings {
		set.Add(v)
	}
	return set
}

func TestToEdgeHosts(t *testing.T) {
	hostUI1 := "uid1"
	hostUI2 := "uid2"

	tests := []struct {
		name        string
		input       map[string]interface{}
		expected    *models.V1EdgeNativeMachinePoolCloudConfigEntity
		expectedErr string
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
						"host_name":       "host1",
						"host_uid":        "uid1",
						"static_ip":       "ip1",
						"nic_name":        "test_nic",
						"default_gateway": "1.1.1.1",
						"subnet_mask":     "2.2.2.2",
						"dns_servers":     ToSchemaSetFromStrings([]string{"t.t.com"}),
					},
					map[string]interface{}{
						"host_name":       "host2",
						"host_uid":        "uid2",
						"static_ip":       "ip2",
						"nic_name":        "test_nic",
						"default_gateway": "1.1.1.1",
						"subnet_mask":     "2.2.2.2",
						"dns_servers":     ToSchemaSetFromStrings([]string{"t.t.com"}),
					},
				},
			},
			expected: &models.V1EdgeNativeMachinePoolCloudConfigEntity{
				EdgeHosts: []*models.V1EdgeNativeMachinePoolHostEntity{
					{
						HostName: "host1",
						HostUID:  &hostUI1,
						Nic: &models.V1Nic{
							IP:      "ip1",
							NicName: "test_nic",
							Gateway: "1.1.1.1",
							Subnet:  "2.2.2.2",
							DNS:     []string{"t.t.com"},
						},
					},
					{
						HostName: "host2",
						HostUID:  &hostUI2,
						Nic: &models.V1Nic{
							IP:      "ip2",
							NicName: "test_nic",
							Gateway: "1.1.1.1",
							Subnet:  "2.2.2.2",
							DNS:     []string{"t.t.com"},
						},
					},
				},
			},
		},
		{
			name: "Edge_host with empty host_name",
			input: map[string]interface{}{
				"edge_host": []interface{}{
					map[string]interface{}{
						"host_name":       "",
						"host_uid":        "uid1",
						"static_ip":       "ip1",
						"nic_name":        "test_nic",
						"default_gateway": "1.1.1.1",
						"subnet_mask":     "2.2.2.2",
						"dns_servers":     ToSchemaSetFromStrings([]string{"t.t.com"}),
					},
					map[string]interface{}{
						"host_name":       "",
						"host_uid":        "uid2",
						"static_ip":       "ip2",
						"nic_name":        "test_nic",
						"default_gateway": "1.1.1.1",
						"subnet_mask":     "2.2.2.2",
						"dns_servers":     ToSchemaSetFromStrings([]string{"t.t.com"}),
					},
				},
			},
			expected: &models.V1EdgeNativeMachinePoolCloudConfigEntity{
				EdgeHosts: []*models.V1EdgeNativeMachinePoolHostEntity{
					{
						HostName: "",
						HostUID:  &hostUI1,
						Nic: &models.V1Nic{
							IP:      "ip1",
							NicName: "test_nic",
							Gateway: "1.1.1.1",
							Subnet:  "2.2.2.2",
							DNS:     []string{"t.t.com"},
						},
					},
					{
						HostName: "",
						HostUID:  &hostUI2,
						Nic: &models.V1Nic{
							IP:      "ip2",
							NicName: "test_nic",
							Gateway: "1.1.1.1",
							Subnet:  "2.2.2.2",
							DNS:     []string{"t.t.com"},
						},
					},
				},
			},
		},
		{
			name: "Valid two node edge hosts",
			input: map[string]interface{}{
				"edge_host": []interface{}{
					map[string]interface{}{
						"host_name":     "",
						"host_uid":      "uid1",
						"static_ip":     "ip1",
						"two_node_role": "primary",
					},
					map[string]interface{}{
						"host_name":     "",
						"host_uid":      "uid2",
						"static_ip":     "ip2",
						"two_node_role": "secondary",
					},
				},
			},
			expected: &models.V1EdgeNativeMachinePoolCloudConfigEntity{
				EdgeHosts: []*models.V1EdgeNativeMachinePoolHostEntity{
					{
						HostName: "",
						HostUID:  &hostUI1,
						Nic: &models.V1Nic{
							IP: "ip1",
						},
						TwoNodeCandidatePriority: "primary",
					},
					{
						HostName: "",
						HostUID:  &hostUI2,
						Nic: &models.V1Nic{
							IP: "ip2",
						},
						TwoNodeCandidatePriority: "secondary",
					},
				},
			},
		},
		{
			name: "Invalid two node edge hosts: duplicate role",
			input: map[string]interface{}{
				"edge_host": []interface{}{
					map[string]interface{}{
						"host_name":     "",
						"host_uid":      hostUI1,
						"static_ip":     "ip1",
						"two_node_role": "primary",
					},
					map[string]interface{}{
						"host_name":     "",
						"host_uid":      hostUI2,
						"static_ip":     "ip2",
						"two_node_role": "primary",
					},
				},
			},
			expected:    nil,
			expectedErr: "two node role 'primary' already assigned to edge host 'uid2'; roles must be unique",
		},
		{
			name: "Invalid two node edge hosts: missing leader",
			input: map[string]interface{}{
				"edge_host": []interface{}{
					map[string]interface{}{
						"host_name":     "",
						"host_uid":      hostUI1,
						"static_ip":     "ip1",
						"two_node_role": "primary",
					},
					map[string]interface{}{
						"host_name": "",
						"host_uid":  hostUI2,
						"static_ip": "ip2",
					},
				},
			},
			expected:    nil,
			expectedErr: "primary edge host 'uid1' specified, but missing secondary edge host",
		},
		{
			name: "Invalid two node edge hosts: missing follower",
			input: map[string]interface{}{
				"edge_host": []interface{}{
					map[string]interface{}{
						"host_name":     "",
						"host_uid":      hostUI1,
						"static_ip":     "ip1",
						"two_node_role": "secondary",
					},
					map[string]interface{}{
						"host_name": "",
						"host_uid":  hostUI2,
						"static_ip": "ip2",
					},
				},
			},
			expected:    nil,
			expectedErr: "secondary edge host 'uid1' specified, but missing primary edge host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toEdgeHosts(tt.input)
			if err != nil && !reflect.DeepEqual(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
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
			edgeHosts, err := toEdgeHosts(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected an error but got none.")
				}
				return
			}
			if err != nil {
				t.Fatalf("Expected no error, but got an error: %v", err)
			}

			expected := &models.V1EdgeNativeMachinePoolConfigEntity{
				CloudConfig: edgeHosts,
				PoolConfig: &models.V1MachinePoolConfigEntity{
					AdditionalLabels:        toAdditionalNodePoolLabels(tt.input),
					Taints:                  toClusterTaints(tt.input),
					IsControlPlane:          true,
					Labels:                  []string{"control-plane"},
					Name:                    ptr.To("pool1"),
					Size:                    ptr.To(int32(len(edgeHosts.EdgeHosts))),
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
			if !cmp.Equal(result.PoolConfig.Labels[0], "control-plane") {
				t.Errorf("Unexpected result (-want +got):\n%s", cmp.Diff(result.PoolConfig.Labels[0], "control-plane"))
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
							Nic: &models.V1Nic{
								IP: "ip1",
							},
						},
						{
							HostName: "host2",
							HostUID:  &huid2,
							Nic: &models.V1Nic{
								IP: "ip2",
							},
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
							Nic: &models.V1Nic{
								IP: "ip3",
							},
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
					"edge_host": []map[string]interface{}{
						{
							"host_name":       "host1",
							"host_uid":        "uid1",
							"static_ip":       "ip1",
							"nic_name":        "",
							"default_gateway": "",
							"subnet_mask":     "",
							"dns_servers":     []string(nil),
						},
						{
							"host_name":       "host2",
							"host_uid":        "uid2",
							"static_ip":       "ip2",
							"nic_name":        "",
							"default_gateway": "",
							"subnet_mask":     "",
							"dns_servers":     []string(nil),
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
					"edge_host": []map[string]interface{}{
						{
							"host_name":       "host3",
							"host_uid":        "uid3",
							"static_ip":       "ip3",
							"nic_name":        "",
							"default_gateway": "",
							"subnet_mask":     "",
							"dns_servers":     []string(nil),
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

func TestGetFirstIPRange(t *testing.T) {
	// Test case 1: Valid CIDR
	cidrValid := "192.168.1.0/24"
	resultValid, errValid := getFirstIPRange(cidrValid)

	// Assertions for valid CIDR
	assert.NoError(t, errValid)
	assert.Equal(t, "192.168.1.1", resultValid)

	// Test case 2: Invalid CIDR
	cidrInvalid := "invalid_cidr"
	resultInvalid, errInvalid := getFirstIPRange(cidrInvalid)

	// Assertions for invalid CIDR
	assert.Error(t, errInvalid)
	assert.Equal(t, "", resultInvalid)
}

func TestFlattenClusterConfigsEdgeNative(t *testing.T) {
	// Test case 1: Valid Cloud Config and Config
	cloudConfig := map[string]interface{}{"vip": "192.168.1.1"}
	validConfig := &models.V1EdgeNativeCloudConfig{
		Spec: &models.V1EdgeNativeCloudConfigSpec{
			ClusterConfig: &models.V1EdgeNativeClusterConfig{
				ControlPlaneEndpoint: &models.V1EdgeNativeControlPlaneEndPoint{
					Host: "192.168.1.1",
				},
				NtpServers: []string{"ntp-server-1", "ntp-server-2"},
				OverlayNetworkConfiguration: &models.V1EdgeNativeOverlayNetworkConfiguration{
					Cidr: "10.0.0.0/16",
				},
				SSHKeys:  []string{"ssh-key-1", "ssh-key-2"},
				StaticIP: false,
			},
		},
	}

	resultValid := flattenClusterConfigsEdgeNative(cloudConfig, validConfig)

	// Assertions for valid Cloud Config and Config
	expectedValidResult := []interface{}{
		map[string]interface{}{
			"ssh_keys":           []string{"ssh-key-1", "ssh-key-2"},
			"vip":                "192.168.1.1",
			"ntp_servers":        []string{"ntp-server-1", "ntp-server-2"},
			"overlay_cidr_range": "10.0.0.0/16",
		},
	}
	assert.Equal(t, expectedValidResult, resultValid)

	// Test case 2: Missing Control Plane Endpoint Host
	missingHostConfig := &models.V1EdgeNativeCloudConfig{
		Spec: &models.V1EdgeNativeCloudConfigSpec{
			ClusterConfig: &models.V1EdgeNativeClusterConfig{
				ControlPlaneEndpoint: &models.V1EdgeNativeControlPlaneEndPoint{},
				OverlayNetworkConfiguration: &models.V1EdgeNativeOverlayNetworkConfiguration{
					Cidr: "",
				},
			},
		},
	}

	resultMissingHost := flattenClusterConfigsEdgeNative(cloudConfig, missingHostConfig)

	// Assertions for missing Control Plane Endpoint Host
	assert.Equal(t, []interface{}{map[string]interface{}{}}, resultMissingHost)

	// Test case 3: Missing Cluster Config
	missingConfig := &models.V1EdgeNativeCloudConfig{}

	resultMissingConfig := flattenClusterConfigsEdgeNative(cloudConfig, missingConfig)

	// Assertions for missing Cluster Config
	assert.Equal(t, []interface{}{}, resultMissingConfig)
}

func TestToOverlayNetworkConfigAndVip(t *testing.T) {
	// Test case 1: Valid cloudConfig with overlay_cidr_range and vip
	validCloudConfig := map[string]interface{}{
		"overlay_cidr_range": "10.0.0.0/16",
		"vip":                "192.168.1.1",
	}

	controlPlaneEndpointValid, overlayConfigValid, errValid := toOverlayNetworkConfigAndVip(validCloudConfig)

	// Assertions for valid cloudConfig
	assert.NoError(t, errValid)
	assert.Equal(t, &models.V1EdgeNativeControlPlaneEndPoint{
		Host: "192.168.1.1",
		Type: "VIP",
	}, controlPlaneEndpointValid)
	assert.Equal(t, &models.V1EdgeNativeOverlayNetworkConfiguration{
		Cidr:   "10.0.0.0/16",
		Enable: true,
	}, overlayConfigValid)

	// Test case 2: Valid cloudConfig with overlay_cidr_range only
	overlayConfigOnly := map[string]interface{}{
		"overlay_cidr_range": "10.0.0.0/16",
	}

	controlPlaneEndpointOverlayOnly, overlayConfigOverlayOnly, errOverlayOnly := toOverlayNetworkConfigAndVip(overlayConfigOnly)

	// Assertions for valid cloudConfig with overlay_cidr_range only
	assert.NoError(t, errOverlayOnly)
	assert.Equal(t, &models.V1EdgeNativeControlPlaneEndPoint{
		Host: "10.0.0.1", // Automatically generated VIP
		Type: "VIP",
	}, controlPlaneEndpointOverlayOnly)
	assert.Equal(t, &models.V1EdgeNativeOverlayNetworkConfiguration{
		Cidr:   "10.0.0.0/16",
		Enable: true,
	}, overlayConfigOverlayOnly)

	// Test case 3: Valid cloudConfig with vip only
	vipOnly := map[string]interface{}{
		"vip": "192.168.1.1",
	}

	controlPlaneEndpointVipOnly, overlayConfigVipOnly, errVipOnly := toOverlayNetworkConfigAndVip(vipOnly)

	// Assertions for valid cloudConfig with vip only
	assert.NoError(t, errVipOnly)
	assert.Equal(t, &models.V1EdgeNativeControlPlaneEndPoint{
		Host: "192.168.1.1",
		Type: "VIP",
	}, controlPlaneEndpointVipOnly)
	assert.Equal(t, &models.V1EdgeNativeOverlayNetworkConfiguration{
		Cidr:   "", // Empty CIDR since overlay_cidr_range is missing
		Enable: false,
	}, overlayConfigVipOnly)

	// Test case 4: Missing cloudConfig fields
	missingFields := map[string]interface{}{}

	controlPlaneEndpointMissingFields, overlayConfigMissingFields, errMissingFields := toOverlayNetworkConfigAndVip(missingFields)

	// Assertions for missing cloudConfig fields
	assert.NoError(t, errMissingFields)
	assert.Equal(t, &models.V1EdgeNativeControlPlaneEndPoint{
		DdnsSearchDomain: "",
		Host:             "",
		Type:             "",
	}, controlPlaneEndpointMissingFields)
	assert.Equal(t, &models.V1EdgeNativeOverlayNetworkConfiguration{
		Cidr:   "",
		Enable: false,
	}, overlayConfigMissingFields)
}
