package virtualmachineinstance

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

func ExpandDomainSpec(d *schema.ResourceData) (*models.V1VMDomainSpec, error) {
	result := &models.V1VMDomainSpec{}

	if v, ok := d.GetOk("resources"); ok {
		resources, err := expandResources(v.([]interface{}))
		if err != nil {
			return result, err
		}
		result.Resources = resources
	}
	if devices, err := expandDevicesToVM(d); err == nil {
		result.Devices = devices
	} else {
		return result, err
	}
	if v, ok := d.GetOk("cpu"); ok {
		cpu, err := expandCPUToVM(v.([]interface{})[0].(map[string]interface{}))
		if err != nil {
			return result, err
		}
		result.CPU = cpu
	}
	if v, ok := d.GetOk("memory"); ok {
		memory, err := expandMemoryToVM(v.([]interface{}))
		if err != nil {
			return result, err
		}
		result.Memory = memory
	}
	if v, ok := d.GetOk("firmware"); ok {
		firmware, err := expandFirmwareToVM(v.([]interface{}))
		if err != nil {
			return result, err
		}
		result.Firmware = firmware
	}
	if v, ok := d.GetOk("features"); ok {
		features, err := expandFeaturesToVM(v.([]interface{}))
		if err != nil {
			return result, err
		}
		result.Features = features
	}

	return result, nil
}

func expandResources(resources []interface{}) (*models.V1VMResourceRequirements, error) {
	result := &models.V1VMResourceRequirements{}

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

func expandDevicesToVM(d *schema.ResourceData) (*models.V1VMDevices, error) {
	result := &models.V1VMDevices{}
	if v, ok := d.GetOk("disk"); ok {
		result.Disks = expandDisksToVM(v.([]interface{}))
	}
	if v, ok := d.GetOk("interface"); ok {
		result.Interfaces = expandInterfacesToVM(v.([]interface{}))
	}
	return result, nil
}

func expandDisksToVM(disks []interface{}) []*models.V1VMDisk {
	if len(disks) == 0 || disks[0] == nil {
		return nil
	}
	result := make([]*models.V1VMDisk, len(disks))
	for i, c := range disks {
		in := c.(map[string]interface{})
		disk := &models.V1VMDisk{}
		if v, ok := in["name"].(string); ok {
			disk.Name = &v
		}
		if v, ok := in["serial"].(string); ok {
			disk.Serial = v
		}
		if v, ok := in["boot_order"].(int); ok && v > 0 {
			disk.BootOrder = int32(v)
		}
		if v, ok := in["disk_device"].([]interface{}); ok && len(v) > 0 {
			expandDiskDeviceToVM(v, disk)
		}
		result[i] = disk
	}
	return result
}

func expandDiskDeviceToVM(diskDevice []interface{}, disk *models.V1VMDisk) {
	if len(diskDevice) == 0 || diskDevice[0] == nil {
		return
	}
	in := diskDevice[0].(map[string]interface{})
	if v, ok := in["disk"].([]interface{}); ok && len(v) > 0 {
		disk.Disk = expandDiskTargetToVM(v)
	}
	if v, ok := in["cdrom"].([]interface{}); ok && len(v) > 0 {
		disk.Cdrom = expandCDRomTargetToVM(v)
	}
	if v, ok := in["lun"].([]interface{}); ok && len(v) > 0 {
		disk.Lun = expandLunTargetToVM(v)
	}
}

func expandDiskTargetToVM(disk []interface{}) *models.V1VMDiskTarget {
	if len(disk) == 0 || disk[0] == nil {
		return nil
	}
	in := disk[0].(map[string]interface{})
	t := &models.V1VMDiskTarget{}
	if v, ok := in["bus"].(string); ok {
		t.Bus = v
	}
	if v, ok := in["read_only"].(bool); ok {
		t.Readonly = v
	}
	if v, ok := in["pci_address"].(string); ok {
		t.PciAddress = v
	}
	return t
}

func expandCDRomTargetToVM(cdrom []interface{}) *models.V1VMCDRomTarget {
	if len(cdrom) == 0 || cdrom[0] == nil {
		return nil
	}
	in := cdrom[0].(map[string]interface{})
	t := &models.V1VMCDRomTarget{}
	if v, ok := in["bus"].(string); ok {
		t.Bus = v
	}
	return t
}

func expandLunTargetToVM(lun []interface{}) *models.V1VMLunTarget {
	if len(lun) == 0 || lun[0] == nil {
		return nil
	}
	in := lun[0].(map[string]interface{})
	t := &models.V1VMLunTarget{}
	if v, ok := in["bus"].(string); ok {
		t.Bus = v
	}
	if v, ok := in["read_only"].(bool); ok {
		t.Readonly = v
	}
	return t
}

func expandInterfacesToVM(interfaces []interface{}) []*models.V1VMInterface {
	if len(interfaces) == 0 || interfaces[0] == nil {
		return nil
	}
	result := make([]*models.V1VMInterface, len(interfaces))
	for i, c := range interfaces {
		in := c.(map[string]interface{})
		iface := &models.V1VMInterface{}
		if v, ok := in["name"].(string); ok {
			iface.Name = &v
		}
		if v, ok := in["model"].(string); ok {
			iface.Model = v
		}
		if v, ok := in["interface_binding_method"].(string); ok {
			setVMInterfaceBindingMethod(v, iface)
		}
		result[i] = iface
	}
	return result
}

func setVMInterfaceBindingMethod(method string, iface *models.V1VMInterface) {
	switch method {
	// case "bridge":
	// 	iface.Bridge = nil
	// case "masquerade":
	// 	iface.Masquerade = nil
	// case "slirp":
	// 	iface.Slirp = nil
	// case "sriov":
	// 	iface.Sriov = nil
	// case "macvtap":
	// 	iface.Macvtap = nil
	// case "passt":
	// 	iface.Passt = nil
	case "InterfaceBridge", "bridge":
		iface.Bridge = struct{}{}
	case "InterfaceMasquerade", "masquerade":
		iface.Masquerade = struct{}{}
	case "InterfaceSlirp", "slirp":
		iface.Slirp = struct{}{}
	case "InterfaceSRIOV", "sriov":
		iface.Sriov = struct{}{}
	case "macvtap":
		iface.Macvtap = struct{}{}
	case "passt":
		iface.Passt = struct{}{}
	}
}

func expandCPUToVM(cpu map[string]interface{}) (*models.V1VMCPU, error) {
	result := &models.V1VMCPU{}
	if len(cpu) == 0 {
		return result, nil
	}
	if v, ok := cpu["cores"].(int); ok {
		if v < 0 {
			return result, fmt.Errorf("cores value %d cannot be negative", v)
		}
		result.Cores = int64(v)
	}
	if v, ok := cpu["sockets"].(int); ok {
		if v < 0 {
			return result, fmt.Errorf("sockets value %d cannot be negative", v)
		}
		result.Sockets = int64(v)
	}
	if v, ok := cpu["threads"].(int); ok {
		if v < 0 {
			return result, fmt.Errorf("threads value %d cannot be negative", v)
		}
		result.Threads = int64(v)
	}
	return result, nil
}

func expandMemoryToVM(memory []interface{}) (*models.V1VMMemory, error) {
	result := &models.V1VMMemory{}
	if len(memory) == 0 || memory[0] == nil {
		return result, nil
	}
	in := memory[0].(map[string]interface{})
	if v, ok := in["guest"].(string); ok && v != "" {
		result.Guest = models.V1VMQuantity(v)
	}
	if v, ok := in["hugepages"].(string); ok && v != "" {
		result.Hugepages = &models.V1VMHugepages{PageSize: v}
	}
	return result, nil
}

func expandFirmwareToVM(firmware []interface{}) (*models.V1VMFirmware, error) {
	if len(firmware) == 0 || firmware[0] == nil {
		return nil, nil
	}
	result := &models.V1VMFirmware{}
	in := firmware[0].(map[string]interface{})
	if v, ok := in["uuid"].(string); ok && v != "" {
		result.UUID = v
	}
	if v, ok := in["serial"].(string); ok && v != "" {
		result.Serial = v
	}
	if v, ok := in["bootloader"].([]interface{}); ok && len(v) > 0 {
		result.Bootloader = expandBootloaderToVM(v)
	}
	return result, nil
}

func expandBootloaderToVM(bootloader []interface{}) *models.V1VMBootloader {
	if len(bootloader) == 0 || bootloader[0] == nil {
		return nil
	}
	result := &models.V1VMBootloader{}
	in := bootloader[0].(map[string]interface{})
	if v, ok := in["bios"].([]interface{}); ok && len(v) > 0 {
		result.Bios = expandBIOSToVM(v)
	}
	if v, ok := in["efi"].([]interface{}); ok && len(v) > 0 {
		result.Efi = expandEFIToVM(v)
	}
	return result
}

func expandBIOSToVM(bios []interface{}) *models.V1VMBIOS {
	if len(bios) == 0 || bios[0] == nil {
		return nil
	}
	result := &models.V1VMBIOS{}
	in := bios[0].(map[string]interface{})
	if v, ok := in["use_serial"].(bool); ok {
		result.UseSerial = v
	}
	return result
}

func expandEFIToVM(efi []interface{}) *models.V1VMEFI {
	if len(efi) == 0 || efi[0] == nil {
		return nil
	}
	result := &models.V1VMEFI{}
	in := efi[0].(map[string]interface{})
	if v, ok := in["secure_boot"].(bool); ok {
		result.SecureBoot = &v
	}
	if v, ok := in["persistent"].(bool); ok {
		result.Persistent = &v
	}
	return result
}

func expandFeaturesToVM(features []interface{}) (*models.V1VMFeatures, error) {
	if len(features) == 0 || features[0] == nil {
		return nil, nil
	}
	result := &models.V1VMFeatures{}
	in := features[0].(map[string]interface{})
	if v, ok := in["acpi"].([]interface{}); ok && len(v) > 0 {
		result.Acpi = expandFeatureStateToVM(v)
	}
	if v, ok := in["apic"].([]interface{}); ok && len(v) > 0 {
		result.Apic = expandFeatureAPICToVM(v)
	}
	if v, ok := in["smm"].([]interface{}); ok && len(v) > 0 {
		result.Smm = expandFeatureStateToVM(v)
	}
	return result, nil
}

func expandFeatureStateToVM(featureState []interface{}) *models.V1VMFeatureState {
	if len(featureState) == 0 || featureState[0] == nil {
		return nil
	}
	result := &models.V1VMFeatureState{}
	in := featureState[0].(map[string]interface{})
	if v, ok := in["enabled"].(bool); ok {
		result.Enabled = v
	}
	return result
}

func expandFeatureAPICToVM(apic []interface{}) *models.V1VMFeatureAPIC {
	if len(apic) == 0 || apic[0] == nil {
		return nil
	}
	result := &models.V1VMFeatureAPIC{}
	in := apic[0].(map[string]interface{})
	if v, ok := in["enabled"].(bool); ok {
		result.Enabled = v
	}
	return result
}

// func ExpandDisks(disks []interface{}) []kubevirtapiv1.Disk {
// 	result := make([]kubevirtapiv1.Disk, len(disks))

// 	if len(disks) == 0 || disks[0] == nil {
// 		return result
// 	}

// 	for i, condition := range disks {
// 		in := condition.(map[string]interface{})

// 		if v, ok := in["name"].(string); ok {
// 			result[i].Name = v
// 		}
// 		if v, ok := in["disk_device"].([]interface{}); ok {
// 			result[i].DiskDevice = expandDiskDevice(v)
// 		}
// 		if v, ok := in["serial"].(string); ok {
// 			result[i].Serial = v
// 		}
// 		if v, ok := in["boot_order"].(int); ok && v > 0 {
// 			bootOrder := common.SafeIntToUint(v)
// 			result[i].BootOrder = &bootOrder
// 		}
// 	}

// 	return result
// }

// func expandDiskDevice(diskDevice []interface{}) kubevirtapiv1.DiskDevice {
// 	result := kubevirtapiv1.DiskDevice{}

// 	if len(diskDevice) == 0 || diskDevice[0] == nil {
// 		return result
// 	}

// 	in := diskDevice[0].(map[string]interface{})

// 	if v, ok := in["disk"].([]interface{}); ok {
// 		result.Disk = expandDiskTarget(v)
// 	}

// 	return result
// }

// func expandDiskTarget(disk []interface{}) *kubevirtapiv1.DiskTarget {
// 	if len(disk) == 0 || disk[0] == nil {
// 		return nil
// 	}

// 	result := &kubevirtapiv1.DiskTarget{}

// 	in := disk[0].(map[string]interface{})

// 	if v, ok := in["bus"].(string); ok {
// 		result.Bus = kubevirtapiv1.DiskBus(v)
// 	}
// 	if v, ok := in["read_only"].(bool); ok {
// 		result.ReadOnly = v
// 	}
// 	if v, ok := in["pci_address"].(string); ok {
// 		result.PciAddress = v
// 	}

// 	return result
// }

// func ExpandInterfaces(interfaces []interface{}) []kubevirtapiv1.Interface {
// 	result := make([]kubevirtapiv1.Interface, len(interfaces))

// 	if len(interfaces) == 0 || interfaces[0] == nil {
// 		return result
// 	}

// 	for i, condition := range interfaces {
// 		in := condition.(map[string]interface{})

// 		if v, ok := in["name"].(string); ok {
// 			result[i].Name = v
// 		}
// 		if v, ok := in["interface_binding_method"].(string); ok {
// 			result[i].InterfaceBindingMethod = expandInterfaceBindingMethod(v)
// 		}
// 		if v, ok := in["model"].(string); ok {
// 			result[i].Model = v
// 		}
// 	}

// 	return result
// }

// func expandInterfaceBindingMethod(interfaceBindingMethod string) kubevirtapiv1.InterfaceBindingMethod {
// 	result := kubevirtapiv1.InterfaceBindingMethod{}

// 	switch interfaceBindingMethod {
// 	case "InterfaceBridge":
// 		result.Bridge = &kubevirtapiv1.InterfaceBridge{}
// 	case "InterfaceSlirp":
// 		result.DeprecatedSlirp = &kubevirtapiv1.DeprecatedInterfaceSlirp{}
// 	case "InterfaceMasquerade":
// 		result.Masquerade = &kubevirtapiv1.InterfaceMasquerade{}
// 	case "InterfaceSRIOV":
// 		result.SRIOV = &kubevirtapiv1.InterfaceSRIOV{}
// 	}

// 	return result
// }

// func FlattenDomainSpec(in kubevirtapiv1.DomainSpec) []interface{} {
// 	att := make(map[string]interface{})

// 	att["resources"] = flattenResources(in.Resources)
// 	if in.CPU != nil && in.CPU.Cores != 0 {
// 		att["cpu"] = flattenCPU(in.CPU)
// 	}
// 	if in.Memory != nil && (in.Memory.Guest != nil || in.Memory.Hugepages != nil) {
// 		att["memory"] = flattenMemory(in.Memory)
// 	}
// 	if in.Firmware != nil {
// 		att["firmware"] = flattenFirmware(in.Firmware)
// 	}
// 	if in.Features != nil {
// 		features := flattenFeatures(in.Features)
// 		if len(features) > 0 {
// 			att["features"] = features
// 		}
// 	}
// 	att["devices"] = flattenDevices(in.Devices)

// 	return []interface{}{att}
// }

// FlattenDomainSpecFromVM flattens *models.V1VMDomainSpec to the same shape as FlattenDomainSpec.
func FlattenDomainSpecFromVM(in *models.V1VMDomainSpec) []interface{} {
	if in == nil {
		return []interface{}{map[string]interface{}{}}
	}
	att := make(map[string]interface{})
	if in.Resources != nil {
		att["resources"] = flattenResourcesFromVM(in.Resources)
	}
	if in.CPU != nil {
		att["cpu"] = flattenCPUFromVM(in.CPU)
	}
	if in.Memory != nil {
		att["memory"] = flattenMemoryFromVM(in.Memory)
	}
	if in.Firmware != nil {
		att["firmware"] = flattenFirmwareFromVM(in.Firmware)
	}
	if in.Features != nil {
		if f := flattenFeaturesFromVM(in.Features); len(f) > 0 {
			att["features"] = f
		}
	}
	if in.Devices != nil {
		att["devices"] = flattenDevicesFromVM(in.Devices)
	}
	return []interface{}{att}
}

func flattenResourcesFromVM(in *models.V1VMResourceRequirements) []interface{} {
	if in == nil {
		return []interface{}{map[string]interface{}{}}
	}
	att := make(map[string]interface{})
	if in.Requests != nil {
		if m, ok := in.Requests.(map[string]interface{}); ok {
			res := make(map[string]string)
			for k, v := range m {
				res[k] = fmt.Sprint(v)
			}
			att["requests"] = utils.FlattenStringMap(res)
		}
	}
	if in.Limits != nil {
		if m, ok := in.Limits.(map[string]interface{}); ok {
			res := make(map[string]string)
			for k, v := range m {
				res[k] = fmt.Sprint(v)
			}
			att["limits"] = utils.FlattenStringMap(res)
		}
	}
	att["over_commit_guest_overhead"] = in.OvercommitGuestOverhead
	return []interface{}{att}
}

func flattenCPUFromVM(in *models.V1VMCPU) []interface{} {
	if in == nil {
		return []interface{}{map[string]interface{}{}}
	}
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

func flattenMemoryFromVM(in *models.V1VMMemory) []interface{} {
	if in == nil {
		return []interface{}{map[string]interface{}{}}
	}
	att := make(map[string]interface{})
	if in.Guest != "" {
		att["guest"] = string(in.Guest)
	}
	if in.Hugepages != nil && in.Hugepages.PageSize != "" {
		att["hugepages"] = in.Hugepages.PageSize
	}
	return []interface{}{att}
}

func flattenFirmwareFromVM(in *models.V1VMFirmware) []interface{} {
	if in == nil {
		return []interface{}{}
	}

	// Don't persist server-generated firmware (uuid/serial only) to state when the user didn't
	// configure firmware — avoids drift on plan (config has no firmware block).
	if in.Bootloader == nil && in.KernelBoot == nil && (in.UUID != "" || in.Serial != "") {
		return []interface{}{}
	}
	att := make(map[string]interface{})
	if in.UUID != "" {
		att["uuid"] = in.UUID
	}
	if in.Serial != "" {
		att["serial"] = in.Serial
	}
	if in.Bootloader != nil {
		if bl := flattenBootloaderFromVM(in.Bootloader); len(bl) > 0 {
			att["bootloader"] = bl
		}
	}
	if len(att) == 0 {
		return []interface{}{}
	}
	return []interface{}{att}
}

func flattenBootloaderFromVM(in *models.V1VMBootloader) []interface{} {
	if in == nil {
		return []interface{}{}
	}
	att := make(map[string]interface{})
	if in.Bios != nil {
		if b := flattenBIOSFromVM(in.Bios); len(b) > 0 {
			att["bios"] = b
		}
	}
	if in.Efi != nil {
		if e := flattenEFIFromVM(in.Efi); len(e) > 0 {
			att["efi"] = e
		}
	}
	if len(att) == 0 {
		return []interface{}{}
	}
	return []interface{}{att}
}

func flattenBIOSFromVM(in *models.V1VMBIOS) []interface{} {
	if in == nil {
		return []interface{}{}
	}
	att := make(map[string]interface{})
	att["use_serial"] = in.UseSerial
	if len(att) == 0 {
		return []interface{}{}
	}
	return []interface{}{att}
}

func flattenEFIFromVM(in *models.V1VMEFI) []interface{} {
	if in == nil {
		return []interface{}{}
	}
	att := make(map[string]interface{})
	if in.SecureBoot != nil {
		att["secure_boot"] = *in.SecureBoot
	}
	if in.Persistent != nil {
		att["persistent"] = *in.Persistent
	}
	if len(att) == 0 {
		return []interface{}{}
	}
	return []interface{}{att}
}

func flattenFeaturesFromVM(in *models.V1VMFeatures) []interface{} {
	if in == nil {
		return []interface{}{}
	}
	att := make(map[string]interface{})
	if in.Acpi != nil {
		att["acpi"] = []interface{}{map[string]interface{}{"enabled": in.Acpi.Enabled}}
	}
	if in.Apic != nil {
		att["apic"] = []interface{}{map[string]interface{}{"enabled": in.Apic.Enabled}}
	}
	if in.Smm != nil {
		att["smm"] = []interface{}{map[string]interface{}{"enabled": in.Smm.Enabled}}
	}
	if len(att) == 0 {
		return []interface{}{}
	}
	return []interface{}{att}
}

func flattenDevicesFromVM(in *models.V1VMDevices) []interface{} {
	if in == nil {
		return []interface{}{map[string]interface{}{}}
	}
	att := make(map[string]interface{})
	att["disk"] = flattenDisksFromVM(in.Disks)
	att["interface"] = flattenInterfacesFromVM(in.Interfaces)
	return []interface{}{att}
}

func flattenDisksFromVM(in []*models.V1VMDisk) []interface{} {
	if len(in) == 0 {
		return nil
	}
	att := make([]interface{}, len(in))
	for i, v := range in {
		if v == nil {
			continue
		}
		c := make(map[string]interface{})
		if v.Name != nil {
			c["name"] = *v.Name
		}
		if v.Serial != "" {
			c["serial"] = v.Serial
		}
		if v.BootOrder > 0 {
			c["boot_order"] = int(v.BootOrder)
		}
		c["disk_device"] = flattenVMDiskDevice(v)
		att[i] = c
	}
	return att
}

func flattenVMDiskDevice(in *models.V1VMDisk) []interface{} {
	if in == nil {
		return []interface{}{map[string]interface{}{}}
	}
	att := make(map[string]interface{})
	if in.Cdrom != nil {
		att["cdrom"] = []interface{}{map[string]interface{}{"bus": in.Cdrom.Bus}}
	}
	if in.Disk != nil {
		d := map[string]interface{}{
			"bus":       in.Disk.Bus,
			"read_only": in.Disk.Readonly,
		}
		if in.Disk.PciAddress != "" {
			d["pci_address"] = in.Disk.PciAddress
		}
		att["disk"] = []interface{}{d}

		//att["disk"] = []interface{}{map[string]interface{}{"bus": in.Disk.Bus}}
	}
	if in.Lun != nil {
		att["lun"] = []interface{}{map[string]interface{}{"bus": in.Lun.Bus}}
	}
	return []interface{}{att}
}

func flattenInterfacesFromVM(in []*models.V1VMInterface) []interface{} {
	if len(in) == 0 {
		return nil
	}
	att := make([]interface{}, len(in))
	for i, v := range in {
		if v == nil {
			continue
		}
		c := make(map[string]interface{})
		if v.Name != nil {
			c["name"] = *v.Name
		}
		if v.Model != "" {
			c["model"] = v.Model
		}
		c["interface_binding_method"] = flattenVMInterfaceBindingMethod(v)
		att[i] = c
	}
	return att
}

// func flattenVMInterfaceBindingMethod(in *models.V1VMInterface) []interface{} {
func flattenVMInterfaceBindingMethod(in *models.V1VMInterface) string {
	if in == nil {
		return ""
	}
	// Hapi uses interface{} for binding types; check which is set (non-nil after JSON unmarshal).
	if in.Masquerade != nil {
		return "InterfaceMasquerade"
	}
	if in.Bridge != nil {
		return "InterfaceBridge"
	}
	if in.Slirp != nil {
		return "InterfaceSlirp"
	}
	if in.Sriov != nil {
		return "InterfaceSRIOV"
	}
	if in.Macvtap != nil {
		return "macvtap"
	}
	if in.Passt != nil {
		return "passt"
	}
	return ""
	// att := make(map[string]interface{})
	// // V1VMInterface has Bridge, Slirp, Masquerade, Sriov as value structs - add placeholder for the one that would be set from API
	// att["bridge"] = []interface{}{map[string]interface{}{}}
	// return []interface{}{att}
}

// func flattenCPU(in *kubevirtapiv1.CPU) []interface{} {
// 	att := make(map[string]interface{})

// 	if in.Cores != 0 {
// 		att["cores"] = in.Cores
// 	}
// 	if in.Sockets != 0 {
// 		att["sockets"] = in.Sockets
// 	}
// 	if in.Threads != 0 {
// 		att["threads"] = in.Threads
// 	}
// 	return []interface{}{att}
// }

// func flattenMemory(in *kubevirtapiv1.Memory) []interface{} {
// 	att := make(map[string]interface{})

// 	if in.Guest != nil {
// 		att["guest"] = in.Guest.String()
// 	}

// 	if in.Hugepages != nil {
// 		att["hugepages"] = in.Hugepages.PageSize
// 	}

// 	return []interface{}{att}
// }

// func flattenFirmware(in *kubevirtapiv1.Firmware) []interface{} {
// 	att := make(map[string]interface{})

// 	if in.UUID != "" {
// 		att["uuid"] = string(in.UUID)
// 	}

// 	if in.Serial != "" {
// 		att["serial"] = in.Serial
// 	}

// 	if in.Bootloader != nil {
// 		bootloader := flattenBootloader(in.Bootloader)
// 		if len(bootloader) > 0 {
// 			att["bootloader"] = bootloader
// 		}
// 	}

// 	return []interface{}{att}
// }

// func flattenBootloader(in *kubevirtapiv1.Bootloader) []interface{} {
// 	att := make(map[string]interface{})

// 	if in.BIOS != nil {
// 		bios := flattenBIOS(in.BIOS)
// 		if len(bios) > 0 {
// 			att["bios"] = bios
// 		}
// 	}

// 	if in.EFI != nil {
// 		efi := flattenEFI(in.EFI)
// 		if len(efi) > 0 {
// 			att["efi"] = efi
// 		}
// 	}

// 	if len(att) == 0 {
// 		return []interface{}{}
// 	}

// 	return []interface{}{att}
// }

// func flattenBIOS(in *kubevirtapiv1.BIOS) []interface{} {
// 	att := make(map[string]interface{})

// 	if in.UseSerial != nil {
// 		att["use_serial"] = *in.UseSerial
// 	}

// 	if len(att) == 0 {
// 		return []interface{}{}
// 	}

// 	return []interface{}{att}
// }

// func flattenEFI(in *kubevirtapiv1.EFI) []interface{} {
// 	att := make(map[string]interface{})

// 	// Always include secure_boot - use default true if nil
// 	if in.SecureBoot != nil {
// 		att["secure_boot"] = *in.SecureBoot
// 	} else {
// 		att["secure_boot"] = false
// 	}

// 	// Always include persistent if it's non-nil (explicitly set by API)
// 	// If nil, set to false as default
// 	if in.Persistent != nil {
// 		att["persistent"] = *in.Persistent
// 	} else {
// 		// Set default to false when empty/nil
// 		att["persistent"] = false
// 	}

// 	if len(att) == 0 {
// 		return []interface{}{}
// 	}

// 	return []interface{}{att}
// }

// func flattenFeatures(in *kubevirtapiv1.Features) []interface{} {
// 	att := make(map[string]interface{})

// 	if in.ACPI.Enabled != nil {
// 		acpi := flattenFeatureState(&in.ACPI)
// 		if len(acpi) > 0 {
// 			att["acpi"] = acpi
// 		}
// 	}

// 	if in.APIC != nil && in.APIC.Enabled != nil {
// 		apic := flattenFeatureState(&kubevirtapiv1.FeatureState{
// 			Enabled: in.APIC.Enabled,
// 		})
// 		if len(apic) > 0 {
// 			att["apic"] = apic
// 		}
// 	}

// 	if in.SMM != nil && in.SMM.Enabled != nil {
// 		smm := flattenFeatureState(in.SMM)
// 		if len(smm) > 0 {
// 			att["smm"] = smm
// 		}
// 	}

// 	if len(att) == 0 {
// 		return []interface{}{}
// 	}

// 	return []interface{}{att}
// }

// func flattenFeatureState(in *kubevirtapiv1.FeatureState) []interface{} {
// 	att := make(map[string]interface{})

// 	if in.Enabled != nil {
// 		att["enabled"] = *in.Enabled
// 	}

// 	if len(att) == 0 {
// 		return []interface{}{}
// 	}

// 	return []interface{}{att}
// }

// func flattenResources(in kubevirtapiv1.ResourceRequirements) []interface{} {
// 	att := make(map[string]interface{})

// 	att["requests"] = utils.FlattenStringMap(utils.FlattenResourceList(in.Requests))
// 	att["limits"] = utils.FlattenStringMap(utils.FlattenResourceList(in.Limits))
// 	att["over_commit_guest_overhead"] = in.OvercommitGuestOverhead

// 	return []interface{}{att}
// }

// func flattenDevices(in kubevirtapiv1.Devices) []interface{} {
// 	att := make(map[string]interface{})

// 	att["disk"] = flattenDisks(in.Disks)
// 	att["interface"] = flattenInterfaces(in.Interfaces)

// 	return []interface{}{att}
// }

// func flattenDisks(in []kubevirtapiv1.Disk) []interface{} {
// 	att := make([]interface{}, len(in))

// 	for i, v := range in {
// 		c := make(map[string]interface{})

// 		c["name"] = v.Name
// 		c["disk_device"] = flattenDiskDevice(v.DiskDevice)
// 		c["serial"] = v.Serial
// 		if v.BootOrder != nil {
// 			c["boot_order"] = common.SafeUintToInt(*v.BootOrder)
// 		}

// 		att[i] = c
// 	}

// 	return att
// }

// func flattenDiskDevice(in kubevirtapiv1.DiskDevice) []interface{} {
// 	att := make(map[string]interface{})

// 	if in.Disk != nil {
// 		att["disk"] = flattenDiskTarget(*in.Disk)
// 	}

// 	return []interface{}{att}
// }

// func flattenDiskTarget(in kubevirtapiv1.DiskTarget) []interface{} {
// 	att := make(map[string]interface{})

// 	att["bus"] = in.Bus
// 	att["read_only"] = in.ReadOnly
// 	att["pci_address"] = in.PciAddress

// 	return []interface{}{att}
// }

// func flattenInterfaces(in []kubevirtapiv1.Interface) []interface{} {
// 	att := make([]interface{}, len(in))

// 	for i, v := range in {
// 		c := make(map[string]interface{})

// 		c["name"] = v.Name
// 		c["interface_binding_method"] = flattenInterfaceBindingMethod(v.InterfaceBindingMethod)
// 		if v.Model != "" {
// 			c["model"] = v.Model
// 		}
// 		att[i] = c
// 	}

// 	return att
// }

// func flattenInterfaceBindingMethod(in kubevirtapiv1.InterfaceBindingMethod) string {
// 	if in.Bridge != nil {
// 		return "InterfaceBridge"
// 	}
// 	if in.DeprecatedSlirp != nil {
// 		return "InterfaceSlirp"
// 	}
// 	if in.Masquerade != nil {
// 		return "InterfaceMasquerade"
// 	}
// 	if in.SRIOV != nil {
// 		return "InterfaceSRIOV"
// 	}

// 	return ""
// }
