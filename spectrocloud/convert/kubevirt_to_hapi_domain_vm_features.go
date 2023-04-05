package convert

import (
	"github.com/spectrocloud/hapi/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

// TODO: implement vm features.

func ToHapiVmFirmware(firmware *kubevirtapiv1.Firmware) *models.V1VMFirmware {
	if firmware == nil {
		return nil
	}

	return &models.V1VMFirmware{
		Bootloader: nil,
		KernelBoot: nil,
		Serial:     "",
		UUID:       string(firmware.UUID),
	}
}

func ToHapiVmFeatures(features *kubevirtapiv1.Features) *models.V1VMFeatures {
	// return stub
	return &models.V1VMFeatures{}
}
