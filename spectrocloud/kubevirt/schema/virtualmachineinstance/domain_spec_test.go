package virtualmachineinstance

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func TestFlattenDomainSpec(t *testing.T) {
	guestQuantity := resource.NewQuantity(2*1024*1024*1024, resource.BinarySI)

	testCases := []struct {
		input          kubevirtapiv1.DomainSpec
		expectedOutput []interface{}
	}{
		{
			input: kubevirtapiv1.DomainSpec{
				CPU:    &kubevirtapiv1.CPU{}, // empty CPU and Memory should be ignored
				Memory: &kubevirtapiv1.Memory{},
			},
			expectedOutput: []interface{}{
				map[string]interface{}{
					"devices": []interface{}{
						map[string]interface{}{
							"disk":      []interface{}{},
							"interface": []interface{}{},
						},
					},
					"resources": []interface{}{
						map[string]interface{}{
							"limits":                     map[string]interface{}{},
							"over_commit_guest_overhead": false,
							"requests":                   map[string]interface{}{},
						},
					},
				},
			},
		},
		{
			input: kubevirtapiv1.DomainSpec{
				CPU: &kubevirtapiv1.CPU{
					Cores:   2,
					Sockets: 1,
					Threads: 1,
				},
			},
			expectedOutput: []interface{}{
				map[string]interface{}{
					"cpu": []interface{}{
						map[string]interface{}{
							"cores":   uint32(2),
							"sockets": uint32(1),
							"threads": uint32(1),
						},
					},
					"devices":   []interface{}{map[string]interface{}{"disk": []interface{}{}, "interface": []interface{}{}}},
					"resources": []interface{}{map[string]interface{}{"limits": map[string]interface{}{}, "over_commit_guest_overhead": false, "requests": map[string]interface{}{}}},
				},
			},
		},
		{
			input: kubevirtapiv1.DomainSpec{
				Memory: &kubevirtapiv1.Memory{
					Guest: guestQuantity,
				},
			},
			expectedOutput: []interface{}{
				map[string]interface{}{
					"memory": []interface{}{
						map[string]interface{}{
							"guest": "2Gi",
						},
					},
					"devices":   []interface{}{map[string]interface{}{"disk": []interface{}{}, "interface": []interface{}{}}},
					"resources": []interface{}{map[string]interface{}{"limits": map[string]interface{}{}, "over_commit_guest_overhead": false, "requests": map[string]interface{}{}}},
				},
			},
		},
		{
			input: kubevirtapiv1.DomainSpec{
				Memory: &kubevirtapiv1.Memory{
					Hugepages: &kubevirtapiv1.Hugepages{
						PageSize: "1Gi",
					},
				},
			},
			expectedOutput: []interface{}{
				map[string]interface{}{
					"memory": []interface{}{
						map[string]interface{}{
							"hugepages": "1Gi",
						},
					},
					"devices":   []interface{}{map[string]interface{}{"disk": []interface{}{}, "interface": []interface{}{}}},
					"resources": []interface{}{map[string]interface{}{"limits": map[string]interface{}{}, "over_commit_guest_overhead": false, "requests": map[string]interface{}{}}},
				},
			},
		},
	}

	for _, tc := range testCases {
		output := flattenDomainSpec(tc.input)

		if diff := cmp.Diff(tc.expectedOutput, output, cmpopts.IgnoreUnexported(resource.Quantity{})); diff != "" {
			t.Errorf("Unexpected result (-want +got):\n%s", diff)
		}
	}
}

func TestExpandDomainSpec(t *testing.T) {
	testCases := []struct {
		input          []interface{}
		expectedOutput kubevirtapiv1.DomainSpec
	}{
		{
			input: []interface{}{
				map[string]interface{}{
					"devices": []interface{}{
						map[string]interface{}{
							"disk":      []interface{}{},
							"interface": []interface{}{},
						},
					},
					"resources": []interface{}{
						map[string]interface{}{
							"limits":                     map[string]interface{}{},
							"over_commit_guest_overhead": false,
							"requests":                   map[string]interface{}{},
						},
					},
				},
			},
			expectedOutput: kubevirtapiv1.DomainSpec{
				Resources: kubevirtapiv1.ResourceRequirements{
					OvercommitGuestOverhead: false,
					Requests:                map[v1.ResourceName]resource.Quantity{},
					Limits:                  map[v1.ResourceName]resource.Quantity{},
				},
				Devices: kubevirtapiv1.Devices{
					Disks:      []kubevirtapiv1.Disk{},
					Interfaces: []kubevirtapiv1.Interface{},
				},
			},
		},
		{
			input: []interface{}{
				map[string]interface{}{
					"cpu": map[string]interface{}{
						"cores":   2,
						"sockets": 1,
						"threads": 1,
					},
					"devices":   []interface{}{map[string]interface{}{"disk": []interface{}{}, "interface": []interface{}{}}},
					"resources": []interface{}{map[string]interface{}{"limits": map[string]interface{}{}, "over_commit_guest_overhead": false, "requests": map[string]interface{}{}}},
				},
			},
			expectedOutput: kubevirtapiv1.DomainSpec{
				CPU: &kubevirtapiv1.CPU{
					Cores:   2,
					Sockets: 1,
					Threads: 1,
				},
				Resources: kubevirtapiv1.ResourceRequirements{
					OvercommitGuestOverhead: false,
					Requests:                map[v1.ResourceName]resource.Quantity{},
					Limits:                  map[v1.ResourceName]resource.Quantity{},
				},
				Devices: kubevirtapiv1.Devices{
					Disks:      []kubevirtapiv1.Disk{},
					Interfaces: []kubevirtapiv1.Interface{},
				},
			},
		},
		{
			input: []interface{}{
				map[string]interface{}{
					"memory": []interface{}{
						map[string]interface{}{
							"guest": "2Gi",
						},
					},
					"devices":   []interface{}{map[string]interface{}{"disk": []interface{}{}, "interface": []interface{}{}}},
					"resources": []interface{}{map[string]interface{}{"limits": map[string]interface{}{}, "over_commit_guest_overhead": false, "requests": map[string]interface{}{}}},
				},
			},
			expectedOutput: kubevirtapiv1.DomainSpec{
				Memory: &kubevirtapiv1.Memory{
					Guest: resource.NewQuantity(2*1024*1024*1024, resource.BinarySI),
				},
				Resources: kubevirtapiv1.ResourceRequirements{
					OvercommitGuestOverhead: false,
					Requests:                map[v1.ResourceName]resource.Quantity{},
					Limits:                  map[v1.ResourceName]resource.Quantity{},
				},
				Devices: kubevirtapiv1.Devices{
					Disks:      []kubevirtapiv1.Disk{},
					Interfaces: []kubevirtapiv1.Interface{},
				},
			},
		},
	}

	for i, tc := range testCases {
		output, err := expandDomainSpec(tc.input)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(output, tc.expectedOutput) {
			if !compareDomainSpec(output, tc.expectedOutput) {
				t.Errorf("Test case %d:\nInput: %#v\nExpected output: %#v\nActual output: %#v", i, tc.input, tc.expectedOutput, output)
			}
		}

	}
}

func compareDomainSpec(a, b kubevirtapiv1.DomainSpec) bool {
	if a.Resources.OvercommitGuestOverhead != b.Resources.OvercommitGuestOverhead {
		return false
	}

	if !reflect.DeepEqual(a.Devices.Disks, b.Devices.Disks) {
		return false
	}

	if !reflect.DeepEqual(a.Devices.Interfaces, b.Devices.Interfaces) {
		return false
	}

	if a.Memory != nil && b.Memory != nil {
		if a.Memory.Guest.Cmp(*b.Memory.Guest) != 0 {
			return false
		}
	} else if a.Memory != b.Memory {
		return false
	}

	return true
}

func TestExpandDisks(t *testing.T) {
	testCases := []struct {
		name     string
		input    []interface{}
		expected []kubevirtapiv1.Disk
	}{
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: []kubevirtapiv1.Disk{},
		},
		{
			name: "multiple disks",
			input: []interface{}{
				map[string]interface{}{
					"name": "disk1",
					"disk_device": []interface{}{
						map[string]interface{}{
							"disk": []interface{}{
								map[string]interface{}{
									"bus":         "virtio",
									"read_only":   true,
									"pci_address": "0000:04:00.0",
								},
							},
						},
					},
					"serial": "123",
				},
				map[string]interface{}{
					"name": "disk2",
					"disk_device": []interface{}{
						map[string]interface{}{
							"disk": []interface{}{
								map[string]interface{}{
									"bus":         "sata",
									"read_only":   false,
									"pci_address": "",
								},
							},
						},
					},
					"serial": "456",
				},
			},
			expected: []kubevirtapiv1.Disk{
				{
					Name: "disk1",
					DiskDevice: kubevirtapiv1.DiskDevice{
						Disk: &kubevirtapiv1.DiskTarget{
							Bus:        "virtio",
							ReadOnly:   true,
							PciAddress: "0000:04:00.0",
						},
					},
					Serial: "123",
				},
				{
					Name: "disk2",
					DiskDevice: kubevirtapiv1.DiskDevice{
						Disk: &kubevirtapiv1.DiskTarget{
							Bus:        "sata",
							ReadOnly:   false,
							PciAddress: "",
						},
					},
					Serial: "456",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := expandDisks(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected: %#v\nActual: %#v", tc.expected, result)
			}
		})
	}
}

func TestExpandInterfaces(t *testing.T) {
	testCases := []struct {
		name     string
		input    []interface{}
		expected []kubevirtapiv1.Interface
	}{
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: []kubevirtapiv1.Interface{},
		},
		{
			name: "multiple interfaces",
			input: []interface{}{
				map[string]interface{}{
					"name":                     "interface1",
					"interface_binding_method": "InterfaceBridge",
					"model":                    "virtio",
				},
				map[string]interface{}{
					"name":                     "interface2",
					"interface_binding_method": "InterfaceSRIOV",
					"model":                    "e1000",
				},
			},
			expected: []kubevirtapiv1.Interface{
				{
					Name: "interface1",
					InterfaceBindingMethod: kubevirtapiv1.InterfaceBindingMethod{
						Bridge: &kubevirtapiv1.InterfaceBridge{},
					},
					Model: "virtio",
				},
				{
					Name: "interface2",
					InterfaceBindingMethod: kubevirtapiv1.InterfaceBindingMethod{
						SRIOV: &kubevirtapiv1.InterfaceSRIOV{},
					},
					Model: "e1000",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := expandInterfaces(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected: %#v\nActual: %#v", tc.expected, result)
			}
		})
	}
}
