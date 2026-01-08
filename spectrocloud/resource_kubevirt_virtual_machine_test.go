package spectrocloud

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vm "github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/virtualmachine"
	vmi "github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/virtualmachineinstance"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/test_utils"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/test_utils/expand_utils"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/test_utils/flatten_utils"
	"gotest.tools/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

// Domain Spec Tests start

func prepareBasicResourceData() *schema.ResourceData {
	rd := resourceKubevirtVirtualMachine().TestResourceData()
	rd.SetId("vm_name")
	rd.Set("name", "vm_name")
	rd.Set("namespace", "default")
	return rd
}

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
		output := vmi.FlattenDomainSpec(tc.input)

		if diff := cmp.Diff(tc.expectedOutput, output, cmpopts.IgnoreUnexported(resource.Quantity{})); diff != "" {
			t.Errorf("Unexpected result (-want +got):\n%s", diff)
		}
	}
}

func prepareExpandDomainSpecTD1() *schema.ResourceData {
	rd := prepareBasicResourceData()
	rd.Set("disk", []interface{}{})
	rd.Set("interface", []interface{}{})
	rd.Set("resources", []interface{}{
		map[string]interface{}{
			"limits":                     map[string]interface{}{},
			"over_commit_guest_overhead": false,
			"requests":                   map[string]interface{}{},
		},
	})
	return rd
}

func prepareExpandDomainSpecTD2() *schema.ResourceData {
	rd := prepareBasicResourceData()
	rd.Set("cpu", []interface{}{map[string]interface{}{
		"cores":   2,
		"sockets": 1,
		"threads": 1,
	}})
	rd.Set("disk", []interface{}{})
	rd.Set("interface", []interface{}{})
	rd.Set("resources", []interface{}{
		map[string]interface{}{
			"limits":                     map[string]interface{}{},
			"over_commit_guest_overhead": false,
			"requests":                   map[string]interface{}{},
		},
	})
	return rd
}

func prepareExpandDomainSpecTD3() *schema.ResourceData {
	rd := prepareBasicResourceData()
	rd.Set("memory", []interface{}{
		map[string]interface{}{
			"guest": "2Gi",
		},
	})
	rd.Set("disk", []interface{}{})
	rd.Set("interface", []interface{}{})
	rd.Set("resources", []interface{}{
		map[string]interface{}{
			"limits":                     map[string]interface{}{},
			"over_commit_guest_overhead": false,
			"requests":                   map[string]interface{}{},
		},
	})
	return rd
}

func TestExpandDomainSpec(t *testing.T) {
	testCases := []struct {
		input          *schema.ResourceData //[]interface{}
		expectedOutput kubevirtapiv1.DomainSpec
	}{
		{
			input: prepareExpandDomainSpecTD1(),
			expectedOutput: kubevirtapiv1.DomainSpec{
				Resources: kubevirtapiv1.ResourceRequirements{
					OvercommitGuestOverhead: false,
					Requests:                map[v1.ResourceName]resource.Quantity{},
					Limits:                  map[v1.ResourceName]resource.Quantity{},
				},
				Devices: kubevirtapiv1.Devices{
					Disks:      nil,
					Interfaces: nil,
				},
			},
		},
		{
			input: prepareExpandDomainSpecTD2(),
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
					Disks:      nil,
					Interfaces: nil,
				},
			},
		},
		{
			input: prepareExpandDomainSpecTD3(),
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
					Disks:      nil,
					Interfaces: nil,
				},
			},
		},
	}

	for i, tc := range testCases {
		output, err := vmi.ExpandDomainSpec(tc.input)
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
					"serial":     "123",
					"boot_order": 1,
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
					"serial":     "456",
					"boot_order": 2,
				},
			},
			expected: []kubevirtapiv1.Disk{
				{
					Name:      "disk1",
					Serial:    "123",
					BootOrder: func() *uint { bo := uint(1); return &bo }(),
					DiskDevice: kubevirtapiv1.DiskDevice{
						Disk: &kubevirtapiv1.DiskTarget{
							Bus:        "virtio",
							ReadOnly:   true,
							PciAddress: "0000:04:00.0",
						},
					},
				},
				{
					Name:      "disk2",
					Serial:    "456",
					BootOrder: func() *uint { bo := uint(2); return &bo }(),
					DiskDevice: kubevirtapiv1.DiskDevice{
						Disk: &kubevirtapiv1.DiskTarget{
							Bus:        "sata",
							ReadOnly:   false,
							PciAddress: "",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := vmi.ExpandDisks(tc.input)
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
			result := vmi.ExpandInterfaces(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected: %#v\nActual: %#v", tc.expected, result)
			}
		})
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

// Domain Spec Tests End

// VM Spec Test's Start

func prepareExpandVirtualMachineSpec(input interface{}) *schema.ResourceData {
	rd := prepareBasicResourceData()
	in := input.([]interface{})[0].(map[string]interface{})
	rd.Set("data_volume_templates", in["data_volume_templates"])
	rd.Set("run_strategy", in["run_strategy"])
	rd.Set("annotations", in["annotations"])
	rd.Set("labels", in["labels"])
	rd.Set("generate_name", in["generate_name"])
	rd.Set("name", in["name"])
	rd.Set("namespace", in["namespace"])
	rd.Set("priority_class_name", in["priority_class_name"])
	rd.Set("resources", in["resources"])
	rd.Set("disk", in["disk"])
	rd.Set("interface", in["interface"])
	rd.Set("node_selector", in["node_selector"])
	//rd.Set("affinity", in["affinity"])
	rd.Set("scheduler_name", in["scheduler_name"])
	rd.Set("tolerations", in["tolerations"])
	rd.Set("eviction_strategy", in["eviction_strategy"])
	rd.Set("termination_grace_period_seconds", in["termination_grace_period_seconds"])
	rd.Set("volume", in["volume"])
	rd.Set("hostname", in["hostname"])
	rd.Set("subdomain", in["subdomain"])
	rd.Set("network", in["network"])
	rd.Set("dns_policy", in["dns_policy"])
	rd.Set("pod_dns_config", in["pod_dns_config"])

	return rd
}

func prepareExpandVirtualMachineSpecWorkingCase(input []interface{}) *schema.ResourceData {
	rd := prepareExpandVirtualMachineSpec(input)
	return rd
}

func prepareExpandVirtualMachineSpecBadTolerationSeconds(input []interface{}) *schema.ResourceData {
	rd := prepareExpandVirtualMachineSpec(input)
	tolerations := rd.Get("tolerations")
	tolerations.([]interface{})[0].(map[string]interface{})["toleration_seconds"] = "a5"
	rd.Set("tolerations", tolerations)
	return rd
}

func prepareExpandVirtualMachineSpecBadPVCRequest(input []interface{}) *schema.ResourceData {
	rd := prepareExpandVirtualMachineSpec(input)
	dt := expand_utils.GetBaseInputForDataVolume()
	dt.([]interface{})[0].(map[string]interface{})["spec"].([]interface{})[0].(map[string]interface{})["pvc"].([]interface{})[0].(map[string]interface{})["resources"].([]interface{})[0].(map[string]interface{})["requests"].(map[string]interface{})["storage"] = "a5"
	if err := rd.Set("data_volume_templates", dt); err != nil {
		return nil
	}
	return rd
}

func prepareExpandVirtualMachineSpecBadPVCLimits(input []interface{}) *schema.ResourceData {
	rd := prepareExpandVirtualMachineSpec(input)
	dt := expand_utils.GetBaseInputForDataVolume()
	dt.([]interface{})[0].(map[string]interface{})["spec"].([]interface{})[0].(map[string]interface{})["pvc"].([]interface{})[0].(map[string]interface{})["resources"].([]interface{})[0].(map[string]interface{})["limits"].(map[string]interface{})["storage"] = "a5"
	if err := rd.Set("data_volume_templates", dt); err != nil {
		return nil
	}
	return rd
}

func prepareExpandVirtualMachineSpecBadDomainResourceRequest(input []interface{}) *schema.ResourceData {
	rd := prepareExpandVirtualMachineSpec(input)
	resources := rd.Get("resources")
	resources.([]interface{})[0].(map[string]interface{})["requests"].(map[string]interface{})["storage"] = "a5"
	rd.Set("resources", resources)
	return rd
}

func prepareExpandVirtualMachineSpecBadDomainResourceLimits(input []interface{}) *schema.ResourceData {
	rd := prepareExpandVirtualMachineSpec(input)
	resources := rd.Get("resources")
	resources.([]interface{})[0].(map[string]interface{})["limits"].(map[string]interface{})["storage"] = "a5"
	rd.Set("resources", resources)
	return rd
}

func TestExpandVirtualMachineSpec(t *testing.T) {
	input := expand_utils.GetBaseInputForVirtualMachine()
	baseOutput := expand_utils.GetBaseOutputForVirtualMachine()

	cases := []struct {
		input                *schema.ResourceData
		name                 string
		shouldError          bool
		expectedOutput       []kubevirtapiv1.VirtualMachineSpec
		expectedErrorMessage string
	}{
		{
			name:        "working case",
			input:       prepareExpandVirtualMachineSpecWorkingCase([]interface{}{input}),
			shouldError: false,
			expectedOutput: []kubevirtapiv1.VirtualMachineSpec{
				baseOutput,
			},
		},
		{
			name:                 "bad toleration_seconds",
			shouldError:          true,
			input:                prepareExpandVirtualMachineSpecBadTolerationSeconds([]interface{}{input}),
			expectedErrorMessage: "invalid toleration_seconds must be int or \"\", got \"a5\"",
		},
		{
			name:                 "bad pvc requests",
			shouldError:          true,
			input:                prepareExpandVirtualMachineSpecBadPVCRequest([]interface{}{input}),
			expectedErrorMessage: "quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'",
		},
		{
			name:                 "bad pvc limits",
			shouldError:          true,
			input:                prepareExpandVirtualMachineSpecBadPVCLimits([]interface{}{input}),
			expectedErrorMessage: "quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'",
		},
		{
			name:                 "bad domain resource requests",
			shouldError:          true,
			input:                prepareExpandVirtualMachineSpecBadDomainResourceRequest([]interface{}{input}),
			expectedErrorMessage: "quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'",
		},
		{
			name:                 "bad domain resource limits",
			shouldError:          true,
			input:                prepareExpandVirtualMachineSpecBadDomainResourceLimits([]interface{}{input}),
			expectedErrorMessage: "quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := vm.ExpandVirtualMachineSpec(tc.input)

			if tc.shouldError {
				assert.Equal(t, tc.expectedErrorMessage, err.Error())
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, output, baseOutput)
			}
		})
	}
}

func TestFlattenVirtualMachineSpec(t *testing.T) {
	input := flatten_utils.GetBaseInputForVirtualMachine()
	output1 := flatten_utils.GetBaseOutputForVirtualMachine()

	cases := []struct {
		input          kubevirtapiv1.VirtualMachineSpec
		expectedOutput []interface{}
	}{
		{
			input: input,
			expectedOutput: []interface{}{
				output1,
			},
		},
	}

	for _, tc := range cases {
		output := vm.FlattenVirtualMachineSpec(tc.input, prepareBasicResourceData())

		//Some fields include terraform randomly generated params that can't be compared
		//so we need to manually remove them
		nullifyUncomparableFields(&output)
		nullifyUncomparableFields(&tc.expectedOutput)

		if diff := cmp.Diff(tc.expectedOutput, output); diff != "" {
			t.Errorf("Unexpected result (-want +got):\n%s", diff)
		}
	}
}

func nullifyUncomparableFields(output *[]interface{}) {
	accessModes := (*output)[0].(map[string]interface{})["data_volume_templates"].([]interface{})[0].(map[string]interface{})["spec"].([]interface{})[0].(map[string]interface{})["pvc"].([]interface{})[0].(map[string]interface{})["access_modes"]
	test_utils.NullifySchemaSetFunction(accessModes.(*schema.Set))

	vmAffinity := (*output)[0].(map[string]interface{})["template"].([]interface{})[0].(map[string]interface{})["spec"].([]interface{})[0].(map[string]interface{})["affinity"]

	podAntiAffinity := vmAffinity.([]interface{})[0].(map[string]interface{})["pod_anti_affinity"].([]interface{})[0].(map[string]interface{})

	podAntiAffinityPreferredNamespace := podAntiAffinity["preferred_during_scheduling_ignored_during_execution"].([]interface{})[0].(map[string]interface{})["pod_affinity_term"].([]interface{})[0].(map[string]interface{})["namespaces"]
	test_utils.NullifySchemaSetFunction(podAntiAffinityPreferredNamespace.(*schema.Set))

	podAntiAffinityRequiredNamespace := podAntiAffinity["required_during_scheduling_ignored_during_execution"].([]interface{})[0].(map[string]interface{})["namespaces"]
	test_utils.NullifySchemaSetFunction(podAntiAffinityRequiredNamespace.(*schema.Set))

	podAffinity := vmAffinity.([]interface{})[0].(map[string]interface{})["pod_affinity"].([]interface{})[0].(map[string]interface{})

	podAffinityPreferredNamespace := podAffinity["preferred_during_scheduling_ignored_during_execution"].([]interface{})[0].(map[string]interface{})["pod_affinity_term"].([]interface{})[0].(map[string]interface{})["namespaces"]
	test_utils.NullifySchemaSetFunction(podAffinityPreferredNamespace.(*schema.Set))

	podAffinityRequiredNamespace := podAffinity["required_during_scheduling_ignored_during_execution"].([]interface{})[0].(map[string]interface{})["namespaces"]
	test_utils.NullifySchemaSetFunction(podAffinityRequiredNamespace.(*schema.Set))

	nodeAffinity := vmAffinity.([]interface{})[0].(map[string]interface{})["node_affinity"].([]interface{})[0].(map[string]interface{})

	nodeSelector := nodeAffinity["required_during_scheduling_ignored_during_execution"].([]interface{})[0].(map[string]interface{})["node_selector_term"].([]interface{})[0].(map[string]interface{})

	nodeRequiredMatchExpressions := nodeSelector["match_expressions"].([]interface{})[0].(map[string]interface{})["values"]
	test_utils.NullifySchemaSetFunction(nodeRequiredMatchExpressions.(*schema.Set))

	nodeRequiredMatchFields := nodeSelector["match_fields"].([]interface{})[0].(map[string]interface{})["values"]
	test_utils.NullifySchemaSetFunction(nodeRequiredMatchFields.(*schema.Set))

	nodePreference := nodeAffinity["preferred_during_scheduling_ignored_during_execution"].([]interface{})[0].(map[string]interface{})["preference"].([]interface{})[0].(map[string]interface{})

	nodePreferredMatchExpressions := nodePreference["match_expressions"].([]interface{})[0].(map[string]interface{})["values"]
	test_utils.NullifySchemaSetFunction(nodePreferredMatchExpressions.(*schema.Set))

	nodePreferredMatchFields := nodePreference["match_fields"].([]interface{})[0].(map[string]interface{})["values"]
	test_utils.NullifySchemaSetFunction(nodePreferredMatchFields.(*schema.Set))
}

func TestFlattenVMMToSpectroSchema(t *testing.T) {
	input := expand_utils.GetBaseOutputForVirtualMachine()
	inter := expand_utils.GetBaseInputForVirtualMachine()

	cases := []struct {
		input          kubevirtapiv1.VirtualMachineSpec
		expectedOutput error
	}{
		{
			input:          input,
			expectedOutput: nil,
		},
	}
	for _, tc := range cases {
		err := vm.FlattenVMMToSpectroSchema(tc.input, prepareExpandVirtualMachineSpec([]interface{}{inter}))
		if diff := cmp.Diff(tc.expectedOutput, err); diff != "" {
			t.Errorf("Unexpected result (-want +got):\n%s", diff)
		}
	}
}

// VM Spec Test's End
