package convert

import (
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ToHapiVmDevices(devices kubevirtapiv1.Devices) *models.V1VMDevices {
	return &models.V1VMDevices{
		AutoattachGraphicsDevice: *DefaultBool(devices.AutoattachGraphicsDevice, true),
		AutoattachInputDevice:    *DefaultBool(devices.AutoattachInputDevice, false),
		AutoattachMemBalloon:     *DefaultBool(devices.AutoattachMemBalloon, true),
		AutoattachPodInterface:   *DefaultBool(devices.AutoattachPodInterface, true),
		AutoattachSerialConsole:  *DefaultBool(devices.AutoattachSerialConsole, true),
		AutoattachVSOCK:          *DefaultBool(devices.AutoattachVSOCK, false),
		BlockMultiQueue:          *DefaultBool(devices.BlockMultiQueue, false),
		// TODO: ClientPassthrough:          nil,
		DisableHotplug:             devices.DisableHotplug,
		Disks:                      ToHapiVmDisks(devices.Disks),
		Filesystems:                ToHapiVmFilesystems(devices.Filesystems),
		Gpus:                       ToHapiVmGpus(devices.GPUs),
		HostDevices:                ToHapiVmHostDevices(devices.HostDevices),
		Inputs:                     ToHapiVmInputs(devices.Inputs),
		Interfaces:                 ToHapiVmInterfaces(devices.Interfaces),
		NetworkInterfaceMultiqueue: *DefaultBool(devices.NetworkInterfaceMultiQueue, false),
		// TODO: Rng:                        nil,
		// TODO: Sound:                      nil,
		// TODO: Tpm:                        nil,
		UseVirtioTransitional: *DefaultBool(devices.UseVirtioTransitional, false),
		// TODO: Watchdog:                   nil,
	}
}

func DefaultBool(b *bool, def bool) *bool {
	if b == nil {
		return &def
	}
	return b
}

func ToHapiVmInputs(inputs []kubevirtapiv1.Input) []*models.V1VMInput {
	var result []*models.V1VMInput
	for _, input := range inputs {
		result = append(result, ToHapiVmInput(input))
	}

	return result
}

func ToHapiVmInput(input kubevirtapiv1.Input) *models.V1VMInput {
	return &models.V1VMInput{
		Bus:  string(input.Bus),
		Name: ptr.StringPtr(input.Name),
		Type: ptr.StringPtr(string(input.Type)),
	}
}

func ToHapiVmHostDevices(devices []kubevirtapiv1.HostDevice) []*models.V1VMHostDevice {
	var result []*models.V1VMHostDevice
	for _, device := range devices {
		result = append(result, ToHapiVmHostDevice(device))
	}

	return result
}

func ToHapiVmHostDevice(device kubevirtapiv1.HostDevice) *models.V1VMHostDevice {
	return &models.V1VMHostDevice{
		DeviceName: ptr.StringPtr(device.DeviceName),
		Name:       ptr.StringPtr(device.Name),
		Tag:        device.Tag,
	}
}

func ToHapiVmGpus(gpus []kubevirtapiv1.GPU) []*models.V1VMGPU {
	var result []*models.V1VMGPU
	for _, u := range gpus {
		result = append(result, ToHapiVmGpu(u))
	}

	return result
}

func ToHapiVmGpu(u kubevirtapiv1.GPU) *models.V1VMGPU {
	return &models.V1VMGPU{
		DeviceName:        ptr.StringPtr(u.DeviceName),
		Name:              ptr.StringPtr(u.Name),
		Tag:               u.Tag,
		VirtualGPUOptions: ToHapiVmVirtualGPUOptions(u.VirtualGPUOptions),
	}
}

func ToHapiVmVirtualGPUOptions(options *kubevirtapiv1.VGPUOptions) *models.V1VMVGPUOptions {
	if options == nil {
		return nil
	}

	return &models.V1VMVGPUOptions{
		Display: ToHapiVmVGPUOptionsDisplay(options.Display),
	}
}

func ToHapiVmVGPUOptionsDisplay(display *kubevirtapiv1.VGPUDisplayOptions) *models.V1VMVGPUDisplayOptions {
	if display == nil {
		return nil
	}

	return &models.V1VMVGPUDisplayOptions{
		Enabled: *display.Enabled,
		RAMFB:   ToHapiVmRamFB(display.RamFB),
	}
}

func ToHapiVmRamFB(fb *kubevirtapiv1.FeatureState) *models.V1VMFeatureState {
	if fb == nil {
		return nil
	}

	return &models.V1VMFeatureState{
		Enabled: *fb.Enabled,
	}
}
