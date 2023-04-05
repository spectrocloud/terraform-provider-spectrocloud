package convert

import (
	"github.com/spectrocloud/hapi/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func ToHapiVmCloudInitNoCloud(cloud *kubevirtapiv1.CloudInitNoCloudSource) *models.V1VMCloudInitNoCloudSource {
	if cloud == nil {
		return nil
	}

	return &models.V1VMCloudInitNoCloudSource{
		NetworkData:          "",
		NetworkDataBase64:    "",
		NetworkDataSecretRef: nil,
		SecretRef:            nil,
		UserData:             "",
		UserDataBase64:       "",
	}
}

func ToHapiVmHostDisk(disk *kubevirtapiv1.HostDisk) *models.V1VMHostDisk {
	if disk == nil {
		return nil
	}

	return &models.V1VMHostDisk{
		Capacity: ToHapiVmQuantityDivisor(disk.Capacity),
		Path:     types.Ptr(disk.Path),
		Shared:   *disk.Shared,
		Type:     types.Ptr(string(disk.Type)),
	}
}

func ToHapiVmEmptyDisk(disk *kubevirtapiv1.EmptyDiskSource) *models.V1VMEmptyDiskSource {
	if disk == nil {
		return nil
	}

	return &models.V1VMEmptyDiskSource{
		Capacity: ToHapiVmQuantityDivisor(disk.Capacity),
	}
}

func ToHapiVmContainerDisk(disk *kubevirtapiv1.ContainerDiskSource) *models.V1VMContainerDiskSource {
	if disk == nil {
		return nil
	}

	return &models.V1VMContainerDiskSource{
		Image:           types.Ptr(disk.Image),
		ImagePullPolicy: string(disk.ImagePullPolicy),
		ImagePullSecret: disk.ImagePullSecret,
		Path:            disk.Path,
	}
}
