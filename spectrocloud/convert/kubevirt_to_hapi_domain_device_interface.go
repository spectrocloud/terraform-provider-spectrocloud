package convert

import (
	"github.com/spectrocloud/hapi/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func ToHapiVmInterfaces(interfaces []kubevirtapiv1.Interface) []*models.V1VMInterface {
	var result []*models.V1VMInterface
	for _, iface := range interfaces {
		result = append(result, ToHapiVmInterface(iface))
	}

	return result
}

func ToHapiVmInterface(iface kubevirtapiv1.Interface) *models.V1VMInterface {
	/*bootOrder := int32(2) // Default value 1 is used for disks
	if iface.BootOrder != nil {
		bootOrder = int32(*iface.BootOrder)
	}*/

	return &models.V1VMInterface{
		AcpiIndex: int32(iface.ACPIIndex),
		// TODO: BootOrder:   bootOrder,
		Bridge:      ToHapiVMInterfaceBridge(iface.InterfaceBindingMethod.Bridge),
		DhcpOptions: ToHapiVMDHCPOptions(iface.DHCPOptions),
		MacAddress:  iface.MacAddress,
		Macvtap:     ToHapiVMInterfaceMacvtap(iface.InterfaceBindingMethod.Macvtap),
		Masquerade:  ToHapiVMInterfaceMasquerade(iface.InterfaceBindingMethod.Masquerade),
		Model:       iface.Model,
		Name:        types.Ptr(iface.Name),
		Passt:       ToHapiVMInterfacePasst(iface.InterfaceBindingMethod.Passt),
		PciAddress:  iface.PciAddress,
		Ports:       ToHapiVMPorts(iface.Ports),
		Slirp:       ToHapiVMInterfaceSlirp(iface.InterfaceBindingMethod.Slirp),
		Sriov:       ToHapiVMInterfaceSRIOV(iface.InterfaceBindingMethod.SRIOV),
		Tag:         iface.Tag,
	}
}

func ToHapiVMPorts(ports []kubevirtapiv1.Port) []*models.V1VMPort {
	var result []*models.V1VMPort
	for _, port := range ports {
		result = append(result, ToHapiVMPort(port))
	}

	return result
}

func ToHapiVMPort(port kubevirtapiv1.Port) *models.V1VMPort {
	return &models.V1VMPort{
		Name:     port.Name,
		Port:     types.Ptr(port.Port),
		Protocol: port.Protocol,
	}
}

func ToHapiVMInterfaceMasquerade(masquerade *kubevirtapiv1.InterfaceMasquerade) models.V1VMInterfaceMasquerade {
	if masquerade == nil {
		return nil
	}
	return make(map[string]interface{})
}

func ToHapiVMDHCPOptions(options *kubevirtapiv1.DHCPOptions) *models.V1VMDHCPOptions {
	if options == nil {
		return nil
	}
	return &models.V1VMDHCPOptions{
		BootFileName:   options.BootFileName,
		NtpServers:     options.NTPServers,
		PrivateOptions: ToHapiVmPrivateOptions(options.PrivateOptions),
		TftpServerName: options.TFTPServerName,
	}
}

func ToHapiVmPrivateOptions(options []kubevirtapiv1.DHCPPrivateOptions) []*models.V1VMDHCPPrivateOptions {
	var result []*models.V1VMDHCPPrivateOptions
	for _, option := range options {
		result = append(result, ToHapiVmPrivateOption(option))
	}

	return result
}

func ToHapiVmPrivateOption(option kubevirtapiv1.DHCPPrivateOptions) *models.V1VMDHCPPrivateOptions {
	return &models.V1VMDHCPPrivateOptions{
		Option: types.Ptr(int32(option.Option)),
		Value:  types.Ptr(option.Value),
	}
}

func ToHapiVMInterfaceSRIOV(sriov *kubevirtapiv1.InterfaceSRIOV) models.V1VMInterfaceSRIOV {
	if sriov == nil {
		return nil
	}
	return make(map[string]interface{})
}

func ToHapiVMInterfacePasst(passt *kubevirtapiv1.InterfacePasst) models.V1VMInterfacePasst {
	if passt == nil {
		return nil
	}
	return make(map[string]interface{})
}

func ToHapiVMInterfaceSlirp(slirp *kubevirtapiv1.InterfaceSlirp) models.V1VMInterfaceSlirp {
	if slirp == nil {
		return nil
	}
	return make(map[string]interface{})
}

func ToHapiVMInterfaceMacvtap(macvtap *kubevirtapiv1.InterfaceMacvtap) models.V1VMInterfaceMacvtap {
	if macvtap == nil {
		return nil
	}
	return make(map[string]interface{})
}

func ToHapiVMInterfaceBridge(bridge *kubevirtapiv1.InterfaceBridge) models.V1VMInterfaceBridge {
	if bridge == nil {
		return nil
	}
	return make(map[string]interface{})
}
