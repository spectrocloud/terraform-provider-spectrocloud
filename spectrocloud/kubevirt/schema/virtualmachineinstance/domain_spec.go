package virtualmachineinstance

import (
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/common"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

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
	if v, ok := d.GetOk("firmware"); ok {
		firmware, err := expandFirmware(v.([]interface{}))
		if err != nil {
			return result, err
		}
		result.Firmware = firmware
	}
	if v, ok := d.GetOk("features"); ok {
		features, err := expandFeatures(v.([]interface{}))
		if err != nil {
			return result, err
		}
		result.Features = features
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
		if v < 0 {
			return result, fmt.Errorf("cores value %d cannot be negative", v)
		}
		if v > math.MaxInt { // Cap to max representable int on this architecture
			return result, fmt.Errorf("cores value %d is out of range for uint32", v)
		}
		result.Cores = common.SafeUint32(v)
	}
	if v, ok := cpu["sockets"].(int); ok {
		if v < 0 {
			return result, fmt.Errorf("sockets value %d cannot be negative", v)
		}
		if v > math.MaxInt { // Cap to max representable int on this architecture
			return result, fmt.Errorf("sockets value %d is out of range for uint32", v)
		}
		result.Sockets = common.SafeUint32(v)
	}
	if v, ok := cpu["threads"].(int); ok {
		if v < 0 {
			return result, fmt.Errorf("threads value %d cannot be negative", v)
		}
		if v > math.MaxInt { // Cap to max representable int on this architecture
			return result, fmt.Errorf("threads value %d is out of range for uint32", v)
		}
		result.Threads = common.SafeUint32(v)
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

func expandFirmware(firmware []interface{}) (*kubevirtapiv1.Firmware, error) {
	if len(firmware) == 0 || firmware[0] == nil {
		return nil, nil
	}

	result := &kubevirtapiv1.Firmware{}
	in := firmware[0].(map[string]interface{})

	if v, ok := in["uuid"].(string); ok && v != "" {
		result.UUID = types.UID(v)
	}

	if v, ok := in["serial"].(string); ok && v != "" {
		result.Serial = v
	}

	if v, ok := in["bootloader"].([]interface{}); ok && len(v) > 0 {
		bootloader, err := expandBootloader(v)
		if err != nil {
			return nil, err
		}
		result.Bootloader = bootloader
	}

	return result, nil
}

func expandBootloader(bootloader []interface{}) (*kubevirtapiv1.Bootloader, error) {
	if len(bootloader) == 0 || bootloader[0] == nil {
		return nil, nil
	}

	result := &kubevirtapiv1.Bootloader{}
	in := bootloader[0].(map[string]interface{})

	if v, ok := in["bios"].([]interface{}); ok && len(v) > 0 {
		bios, err := expandBIOS(v)
		if err != nil {
			return nil, err
		}
		result.BIOS = bios
	}

	if v, ok := in["efi"].([]interface{}); ok && len(v) > 0 {
		efi, err := expandEFI(v)
		if err != nil {
			return nil, err
		}
		result.EFI = efi
	}

	return result, nil
}

func expandBIOS(bios []interface{}) (*kubevirtapiv1.BIOS, error) {
	if len(bios) == 0 || bios[0] == nil {
		return &kubevirtapiv1.BIOS{}, nil
	}

	result := &kubevirtapiv1.BIOS{}
	in := bios[0].(map[string]interface{})

	if v, ok := in["use_serial"].(bool); ok {
		result.UseSerial = &v
	}

	return result, nil
}

func expandEFI(efi []interface{}) (*kubevirtapiv1.EFI, error) {
	if len(efi) == 0 || efi[0] == nil {
		return &kubevirtapiv1.EFI{}, nil
	}

	result := &kubevirtapiv1.EFI{}
	in := efi[0].(map[string]interface{})

	if v, ok := in["secure_boot"]; ok {
		if secureBoot, ok := v.(bool); ok {
			result.SecureBoot = &secureBoot
		}
	}

	if v, ok := in["persistent"]; ok {
		if persistent, ok := v.(bool); ok {
			result.Persistent = &persistent
		}
	}

	return result, nil
}

func expandFeatures(features []interface{}) (*kubevirtapiv1.Features, error) {
	if len(features) == 0 || features[0] == nil {
		return nil, nil
	}

	result := &kubevirtapiv1.Features{}
	in := features[0].(map[string]interface{})

	if v, ok := in["acpi"].([]interface{}); ok && len(v) > 0 {
		acpi, err := expandFeatureState(v)
		if err != nil {
			return nil, err
		}
		result.ACPI = *acpi
	}

	if v, ok := in["apic"].([]interface{}); ok && len(v) > 0 {
		apic, err := expandFeatureState(v)
		if err != nil {
			return nil, err
		}
		result.APIC = &kubevirtapiv1.FeatureAPIC{
			Enabled: apic.Enabled,
		}
	}

	if v, ok := in["smm"].([]interface{}); ok && len(v) > 0 {
		smm, err := expandFeatureState(v)
		if err != nil {
			return nil, err
		}
		result.SMM = smm
	}

	return result, nil
}

func expandFeatureState(featureState []interface{}) (*kubevirtapiv1.FeatureState, error) {
	if len(featureState) == 0 || featureState[0] == nil {
		return &kubevirtapiv1.FeatureState{}, nil
	}

	result := &kubevirtapiv1.FeatureState{}
	in := featureState[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		result.Enabled = &v
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
		if v, ok := in["boot_order"].(int); ok && v > 0 {
			bootOrder := uint(v)
			result[i].BootOrder = &bootOrder
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
	if in.Firmware != nil {
		att["firmware"] = flattenFirmware(in.Firmware)
	}
	if in.Features != nil {
		features := flattenFeatures(in.Features)
		if len(features) > 0 {
			att["features"] = features
		}
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

func flattenFirmware(in *kubevirtapiv1.Firmware) []interface{} {
	att := make(map[string]interface{})

	if in.UUID != "" {
		att["uuid"] = string(in.UUID)
	}

	if in.Serial != "" {
		att["serial"] = in.Serial
	}

	if in.Bootloader != nil {
		bootloader := flattenBootloader(in.Bootloader)
		if len(bootloader) > 0 {
			att["bootloader"] = bootloader
		}
	}

	return []interface{}{att}
}

func flattenBootloader(in *kubevirtapiv1.Bootloader) []interface{} {
	att := make(map[string]interface{})

	if in.BIOS != nil {
		bios := flattenBIOS(in.BIOS)
		if len(bios) > 0 {
			att["bios"] = bios
		}
	}

	if in.EFI != nil {
		efi := flattenEFI(in.EFI)
		if len(efi) > 0 {
			att["efi"] = efi
		}
	}

	if len(att) == 0 {
		return []interface{}{}
	}

	return []interface{}{att}
}

func flattenBIOS(in *kubevirtapiv1.BIOS) []interface{} {
	att := make(map[string]interface{})

	if in.UseSerial != nil {
		att["use_serial"] = *in.UseSerial
	}

	if len(att) == 0 {
		return []interface{}{}
	}

	return []interface{}{att}
}

func flattenEFI(in *kubevirtapiv1.EFI) []interface{} {
	att := make(map[string]interface{})

	// Always include secure_boot - use default true if nil
	if in.SecureBoot != nil {
		att["secure_boot"] = *in.SecureBoot
	} else {
		att["secure_boot"] = false
	}

	// Always include persistent if it's non-nil (explicitly set by API)
	// If nil, set to false as default
	if in.Persistent != nil {
		att["persistent"] = *in.Persistent
	} else {
		// Set default to false when empty/nil
		att["persistent"] = false
	}

	if len(att) == 0 {
		return []interface{}{}
	}

	return []interface{}{att}
}

func flattenFeatures(in *kubevirtapiv1.Features) []interface{} {
	att := make(map[string]interface{})

	if in.ACPI.Enabled != nil {
		acpi := flattenFeatureState(&in.ACPI)
		if len(acpi) > 0 {
			att["acpi"] = acpi
		}
	}

	if in.APIC != nil && in.APIC.Enabled != nil {
		apic := flattenFeatureState(&kubevirtapiv1.FeatureState{
			Enabled: in.APIC.Enabled,
		})
		if len(apic) > 0 {
			att["apic"] = apic
		}
	}

	if in.SMM != nil && in.SMM.Enabled != nil {
		smm := flattenFeatureState(in.SMM)
		if len(smm) > 0 {
			att["smm"] = smm
		}
	}

	if len(att) == 0 {
		return []interface{}{}
	}

	return []interface{}{att}
}

func flattenFeatureState(in *kubevirtapiv1.FeatureState) []interface{} {
	att := make(map[string]interface{})

	if in.Enabled != nil {
		att["enabled"] = *in.Enabled
	}

	if len(att) == 0 {
		return []interface{}{}
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
		if v.BootOrder != nil {
			// Safe conversion from uint to int to avoid potential overflow
			if *v.BootOrder <= math.MaxInt32 {
				c["boot_order"] = int(*v.BootOrder)
			}
		}

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
