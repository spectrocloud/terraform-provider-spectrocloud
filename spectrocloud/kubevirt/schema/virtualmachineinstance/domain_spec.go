package virtualmachineinstance

import (
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"k8s.io/apimachinery/pkg/api/resource"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

func domainSpecFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"resources": {
			Type:        schema.TypeList,
			Description: "Resources describes the Compute Resources required by this vmi.",
			MaxItems:    1,
			Required:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"requests": {
						Type:        schema.TypeMap,
						Description: "Requests is a description of the initial vmi resources.",
						Optional:    true,
					},
					"limits": {
						Type:        schema.TypeMap,
						Description: "Requests is the maximum amount of compute resources allowed. Valid resource keys are \"memory\" and \"cpu\"",
						Optional:    true,
					},
					"over_commit_guest_overhead": {
						Type:        schema.TypeBool,
						Description: "Don't ask the scheduler to take the guest-management overhead into account. Instead put the overhead only into the container's memory limit. This can lead to crashes if all memory is in use on a node. Defaults to false.",
						Optional:    true,
					},
				},
			},
		},
		"cpu": {
			Type:        schema.TypeList,
			Description: "CPU allows to specifying the CPU topology. Valid resource keys are \"cores\" , \"sockets\" and \"threads\"",
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cores": {
						Type:        schema.TypeInt,
						Description: "Cores is the number of cores inside the vmi. Must be a value greater or equal 1",
						Optional:    true,
					},
					"sockets": {
						Type:        schema.TypeInt,
						Description: "Sockets is the number of sockets inside the vmi. Must be a value greater or equal 1.",
						Optional:    true,
					},
					"threads": {
						Type:        schema.TypeInt,
						Description: "Threads is the number of threads inside the vmi. Must be a value greater or equal 1.",
						Optional:    true,
					},
				},
			},
		},
		"memory": {
			Type:        schema.TypeList,
			Description: "Memory allows specifying the vmi memory features.",
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"guest": {
						Type:        schema.TypeString,
						Description: "Guest is the amount of memory allocated to the vmi. This value must be less than or equal to the limit if specified.",
						Optional:    true,
					},
					"hugepages": {
						Type: schema.TypeString,
						// PageSize specifies the hugepage size, for x86_64 architecture valid values are 1Gi and 2Mi.
						Description: "Hugepages attribute specifies the hugepage size, for x86_64 architecture valid values are 1Gi and 2Mi.",
						Optional:    true,
					},
				},
			},
		},
		"devices": {
			Type:        schema.TypeList,
			Description: "Devices allows adding disks, network interfaces, ...",
			MaxItems:    1,
			Required:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"disk": {
						Type:        schema.TypeList,
						Description: "Disks describes disks, cdroms, floppy and luns which are connected to the vmi.",
						Required:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:        schema.TypeString,
									Description: "Name is the device name",
									Required:    true,
								},
								"disk_device": {
									Type:        schema.TypeList,
									Description: "DiskDevice specifies as which device the disk should be added to the guest.",
									Required:    true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"disk": {
												Type:        schema.TypeList,
												Description: "Attach a volume as a disk to the vmi.",
												Optional:    true,
												Elem: &schema.Resource{
													Schema: map[string]*schema.Schema{
														"bus": {
															Type:        schema.TypeString,
															Description: "Bus indicates the type of disk device to emulate.",
															Required:    true,
														},
														"read_only": {
															Type:        schema.TypeBool,
															Description: "ReadOnly. Defaults to false.",
															Optional:    true,
														},
														"pci_address": {
															Type:        schema.TypeString,
															Description: "If specified, the virtual disk will be placed on the guests pci address with the specifed PCI address. For example: 0000:81:01.10",
															Optional:    true,
														},
													},
												},
											},
										},
									},
								},
								"serial": {
									Type:        schema.TypeString,
									Description: "Serial provides the ability to specify a serial number for the disk device.",
									Optional:    true,
								},
							},
						},
					},
					"interface": {
						Type:        schema.TypeList,
						Description: "Interfaces describe network interfaces which are added to the vmi.",
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:        schema.TypeString,
									Description: "Logical name of the interface as well as a reference to the associated networks.",
									Required:    true,
								},
								"interface_binding_method": {
									Type: schema.TypeString,
									ValidateFunc: validation.StringInSlice([]string{
										"InterfaceBridge",
										"InterfaceSlirp",
										"InterfaceMasquerade",
										"InterfaceSRIOV",
									}, false),
									Description: "Represents the Interface model, One of: e1000, e1000e, ne2k_pci, pcnet, rtl8139, virtio. Defaults to virtio.",
									Required:    true,
								},
								"model": {
									Type:     schema.TypeString,
									Optional: true,
									ValidateFunc: validation.StringInSlice([]string{
										"",
										"e1000",
										"e1000e",
										"ne2k_pci",
										"pcnet",
										"rtl8139",
										"virtio",
									}, false),
									Description: "Represents the method which will be used to connect the interface to the guest.",
								},
							},
						},
					},
				},
			},
		},
	}
}

func domainSpecSchema() *schema.Schema {
	fields := domainSpecFields()

	return &schema.Schema{
		Type: schema.TypeList,

		Description: "Specification of the desired behavior of the VirtualMachineInstance on the host.",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func ExpandDomainSpec(d *schema.ResourceData) (kubevirtapiv1.DomainSpec, error) {
	result := kubevirtapiv1.DomainSpec{}

	if v, ok := d.GetOk("resources"); ok {
		resources, err := expandResources(v.([]interface{}))
		if err != nil {
			return result, err
		}
		result.Resources = resources
	}
	if devices, err := expandDevices(d); err == nil {
		result.Devices = devices
	} else {
		return result, err
	}
	if v, ok := d.GetOk("cpu"); ok {
		cpu, err := expandCPU(v.([]interface{})[0].(map[string]interface{}))
		if err != nil {
			return result, err
		}
		result.CPU = &cpu
	}
	if v, ok := d.GetOk("memory"); ok {
		memory, err := expandMemory(v.([]interface{}))
		if err != nil {
			return result, err
		}
		result.Memory = &memory
	}

	return result, nil
}

func expandResources(resources []interface{}) (kubevirtapiv1.ResourceRequirements, error) {
	result := kubevirtapiv1.ResourceRequirements{}

	if len(resources) == 0 || resources[0] == nil {
		return result, nil
	}

	in := resources[0].(map[string]interface{})

	if v, ok := in["requests"].(map[string]interface{}); ok {
		requests, err := utils.ExpandMapToResourceList(v)
		if err != nil {
			return result, err
		}
		result.Requests = *requests
	}
	if v, ok := in["limits"].(map[string]interface{}); ok {
		limits, err := utils.ExpandMapToResourceList(v)
		if err != nil {
			return result, err
		}
		result.Limits = *limits
	}
	if v, ok := in["over_commit_guest_overhead"].(bool); ok {
		result.OvercommitGuestOverhead = v
	}

	return result, nil
}

func expandDevices(d *schema.ResourceData) (kubevirtapiv1.Devices, error) {
	result := kubevirtapiv1.Devices{}

	if v, ok := d.GetOk("disk"); ok {
		result.Disks = ExpandDisks(v.([]interface{}))
	}
	if v, ok := d.GetOk("interface"); ok {
		result.Interfaces = ExpandInterfaces(v.([]interface{}))
	}

	return result, nil
}

func expandCPU(cpu map[string]interface{}) (kubevirtapiv1.CPU, error) {
	result := kubevirtapiv1.CPU{}

	if len(cpu) == 0 {
		return result, nil
	}

	if v, ok := cpu["cores"].(int); ok {
		if v < 0 || uint64(v) > math.MaxUint32 {
			return result, fmt.Errorf("cores value %d is out of range for uint32", v)
		}
		result.Cores = uint32(v)
	}
	if v, ok := cpu["sockets"].(int); ok {
		if v < 0 || uint64(v) > math.MaxUint32 {
			return result, fmt.Errorf("sockets value %d is out of range for uint32", v)
		}
		result.Sockets = uint32(v)
	}
	if v, ok := cpu["threads"].(int); ok {
		if v < 0 || uint64(v) > math.MaxUint32 {
			return result, fmt.Errorf("threads value %d is out of range for uint32", v)
		}
		result.Threads = uint32(v)
	}

	return result, nil
}

func expandMemory(memory []interface{}) (kubevirtapiv1.Memory, error) {
	result := kubevirtapiv1.Memory{}

	if len(memory) == 0 || memory[0] == nil {
		return result, nil
	}

	in := memory[0].(map[string]interface{})

	if v, ok := in["guest"].(string); ok {
		got, err := resource.ParseQuantity(v)
		if err != nil {
			return result, err
		}
		result.Guest = &got
	}

	if v, ok := in["hugepages"].(string); ok {
		if in["hugepages"].(string) != "" {
			result.Hugepages = &kubevirtapiv1.Hugepages{
				PageSize: v,
			}
		}
	}

	return result, nil
}

func ExpandDisks(disks []interface{}) []kubevirtapiv1.Disk {
	result := make([]kubevirtapiv1.Disk, len(disks))

	if len(disks) == 0 || disks[0] == nil {
		return result
	}

	for i, condition := range disks {
		in := condition.(map[string]interface{})

		if v, ok := in["name"].(string); ok {
			result[i].Name = v
		}
		if v, ok := in["disk_device"].([]interface{}); ok {
			result[i].DiskDevice = expandDiskDevice(v)
		}
		if v, ok := in["serial"].(string); ok {
			result[i].Serial = v
		}
	}

	return result
}

func expandDiskDevice(diskDevice []interface{}) kubevirtapiv1.DiskDevice {
	result := kubevirtapiv1.DiskDevice{}

	if len(diskDevice) == 0 || diskDevice[0] == nil {
		return result
	}

	in := diskDevice[0].(map[string]interface{})

	if v, ok := in["disk"].([]interface{}); ok {
		result.Disk = expandDiskTarget(v)
	}

	return result
}

func expandDiskTarget(disk []interface{}) *kubevirtapiv1.DiskTarget {
	if len(disk) == 0 || disk[0] == nil {
		return nil
	}

	result := &kubevirtapiv1.DiskTarget{}

	in := disk[0].(map[string]interface{})

	if v, ok := in["bus"].(string); ok {
		result.Bus = kubevirtapiv1.DiskBus(v)
	}
	if v, ok := in["read_only"].(bool); ok {
		result.ReadOnly = v
	}
	if v, ok := in["pci_address"].(string); ok {
		result.PciAddress = v
	}

	return result
}

func ExpandInterfaces(interfaces []interface{}) []kubevirtapiv1.Interface {
	result := make([]kubevirtapiv1.Interface, len(interfaces))

	if len(interfaces) == 0 || interfaces[0] == nil {
		return result
	}

	for i, condition := range interfaces {
		in := condition.(map[string]interface{})

		if v, ok := in["name"].(string); ok {
			result[i].Name = v
		}
		if v, ok := in["interface_binding_method"].(string); ok {
			result[i].InterfaceBindingMethod = expandInterfaceBindingMethod(v)
		}
		if v, ok := in["model"].(string); ok {
			result[i].Model = v
		}
	}

	return result
}

func expandInterfaceBindingMethod(interfaceBindingMethod string) kubevirtapiv1.InterfaceBindingMethod {
	result := kubevirtapiv1.InterfaceBindingMethod{}

	switch interfaceBindingMethod {
	case "InterfaceBridge":
		result.Bridge = &kubevirtapiv1.InterfaceBridge{}
	case "InterfaceSlirp":
		result.DeprecatedSlirp = &kubevirtapiv1.DeprecatedInterfaceSlirp{}
	case "InterfaceMasquerade":
		result.Masquerade = &kubevirtapiv1.InterfaceMasquerade{}
	case "InterfaceSRIOV":
		result.SRIOV = &kubevirtapiv1.InterfaceSRIOV{}
	}

	return result
}

func FlattenDomainSpec(in kubevirtapiv1.DomainSpec) []interface{} {
	att := make(map[string]interface{})

	att["resources"] = flattenResources(in.Resources)
	if in.CPU != nil && in.CPU.Cores != 0 {
		att["cpu"] = flattenCPU(in.CPU)
	}
	if in.Memory != nil && (in.Memory.Guest != nil || in.Memory.Hugepages != nil) {
		att["memory"] = flattenMemory(in.Memory)
	}
	att["devices"] = flattenDevices(in.Devices)

	return []interface{}{att}
}

func flattenCPU(in *kubevirtapiv1.CPU) []interface{} {
	att := make(map[string]interface{})

	if in.Cores != 0 {
		att["cores"] = in.Cores
	}
	if in.Sockets != 0 {
		att["sockets"] = in.Sockets
	}
	if in.Threads != 0 {
		att["threads"] = in.Threads
	}
	return []interface{}{att}
}

func flattenMemory(in *kubevirtapiv1.Memory) []interface{} {
	att := make(map[string]interface{})

	if in.Guest != nil {
		att["guest"] = in.Guest.String()
	}

	if in.Hugepages != nil {
		att["hugepages"] = in.Hugepages.PageSize
	}

	return []interface{}{att}
}

func flattenResources(in kubevirtapiv1.ResourceRequirements) []interface{} {
	att := make(map[string]interface{})

	att["requests"] = utils.FlattenStringMap(utils.FlattenResourceList(in.Requests))
	att["limits"] = utils.FlattenStringMap(utils.FlattenResourceList(in.Limits))
	att["over_commit_guest_overhead"] = in.OvercommitGuestOverhead

	return []interface{}{att}
}

func flattenDevices(in kubevirtapiv1.Devices) []interface{} {
	att := make(map[string]interface{})

	att["disk"] = flattenDisks(in.Disks)
	att["interface"] = flattenInterfaces(in.Interfaces)

	return []interface{}{att}
}

func flattenDisks(in []kubevirtapiv1.Disk) []interface{} {
	att := make([]interface{}, len(in))

	for i, v := range in {
		c := make(map[string]interface{})

		c["name"] = v.Name
		c["disk_device"] = flattenDiskDevice(v.DiskDevice)
		c["serial"] = v.Serial

		att[i] = c
	}

	return att
}

func flattenDiskDevice(in kubevirtapiv1.DiskDevice) []interface{} {
	att := make(map[string]interface{})

	if in.Disk != nil {
		att["disk"] = flattenDiskTarget(*in.Disk)
	}

	return []interface{}{att}
}

func flattenDiskTarget(in kubevirtapiv1.DiskTarget) []interface{} {
	att := make(map[string]interface{})

	att["bus"] = in.Bus
	att["read_only"] = in.ReadOnly
	att["pci_address"] = in.PciAddress

	return []interface{}{att}
}

func flattenInterfaces(in []kubevirtapiv1.Interface) []interface{} {
	att := make([]interface{}, len(in))

	for i, v := range in {
		c := make(map[string]interface{})

		c["name"] = v.Name
		c["interface_binding_method"] = flattenInterfaceBindingMethod(v.InterfaceBindingMethod)
		if v.Model != "" {
			c["model"] = v.Model
		}
		att[i] = c
	}

	return att
}

func flattenInterfaceBindingMethod(in kubevirtapiv1.InterfaceBindingMethod) string {
	if in.Bridge != nil {
		return "InterfaceBridge"
	}
	if in.DeprecatedSlirp != nil {
		return "InterfaceSlirp"
	}
	if in.Masquerade != nil {
		return "InterfaceMasquerade"
	}
	if in.SRIOV != nil {
		return "InterfaceSRIOV"
	}

	return ""
}
