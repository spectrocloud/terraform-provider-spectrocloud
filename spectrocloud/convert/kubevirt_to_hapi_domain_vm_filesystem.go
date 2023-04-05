package convert

import (
	"github.com/spectrocloud/hapi/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func ToHapiVmFilesystems(filesystems []kubevirtapiv1.Filesystem) []*models.V1VMFilesystem {
	var result []*models.V1VMFilesystem
	for _, filesystem := range filesystems {
		result = append(result, ToHapiVmFilesystemItem(filesystem))
	}

	return result
}

func ToHapiVmFilesystemItem(filesystem kubevirtapiv1.Filesystem) *models.V1VMFilesystem {
	return &models.V1VMFilesystem{
		Name: types.Ptr(filesystem.Name),
		// TODO: Virtiofs: ToHapiVmVirtiofs(filesystem.Virtiofs),
	}
}
