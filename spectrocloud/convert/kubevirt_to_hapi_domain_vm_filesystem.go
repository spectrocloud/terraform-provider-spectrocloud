package convert

import (
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
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
		Name: ptr.StringPtr(filesystem.Name),
		// TODO: Virtiofs: ToHapiVmVirtiofs(filesystem.Virtiofs),
	}
}
