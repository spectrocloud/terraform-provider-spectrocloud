package convert

import (
	"github.com/spectrocloud/hapi/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ToHapiVmDomain(domain kubevirtapiv1.DomainSpec) *models.V1VMDomainSpec {
	var IOThreadPolicy string
	if domain.IOThreadsPolicy != nil {
		IOThreadPolicy = string(*domain.IOThreadsPolicy)
	}
	return &models.V1VMDomainSpec{
		Chassis:         ToHapiVmChassis(domain.Chassis),
		CPU:             ToHapiVmCPU(domain.CPU),
		Clock:           ToHapiVmClock(domain.Clock),
		Devices:         ToHapiVmDevices(domain.Devices),
		Features:        ToHapiVmFeatures(domain.Features),
		Firmware:        ToHapiVmFirmware(domain.Firmware),
		IoThreadsPolicy: IOThreadPolicy,
		LaunchSecurity:  ToHapiVmLaunchSecurity(domain.LaunchSecurity),
		Machine:         ToHapiVmMachine(domain.Machine),
		Memory:          ToHapiVmMemory(domain.Memory),
		Resources:       ToHapiVmResources(domain.Resources),
	}
}

func ToHapiVmLaunchSecurity(security *kubevirtapiv1.LaunchSecurity) *models.V1VMLaunchSecurity {
	if security == nil {
		return nil
	}

	return &models.V1VMLaunchSecurity{
		Sev: models.V1VMSEV(*security.SEV),
	}
}

func ToHapiVmMachine(machine *kubevirtapiv1.Machine) *models.V1VMMachine {
	if machine == nil {
		return nil
	}

	return &models.V1VMMachine{
		Type: machine.Type,
	}
}

func ToHapiVmResources(resources kubevirtapiv1.ResourceRequirements) *models.V1VMResourceRequirements {
	if resources.Requests == nil {
		return nil
	}

	return &models.V1VMResourceRequirements{
		Requests: ToHapiVmResourceList(resources.Requests),
	}
}

func ToHapiVmChassis(chassis *kubevirtapiv1.Chassis) *models.V1VMChassis {
	if chassis == nil {
		return nil
	}

	return &models.V1VMChassis{
		Manufacturer: chassis.Manufacturer,
		Serial:       chassis.Serial,
		Sku:          chassis.Sku,
		Version:      chassis.Version,
		Asset:        chassis.Asset,
	}
}
