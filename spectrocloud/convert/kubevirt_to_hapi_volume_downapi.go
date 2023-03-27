package convert

import (
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	"k8s.io/api/core/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ToHapiVmDownwardAPI(api *kubevirtapiv1.DownwardAPIVolumeSource) *models.V1VMDownwardAPIVolumeSource {
	if api == nil {
		return nil
	}

	return &models.V1VMDownwardAPIVolumeSource{
		Fields:      ToHapiVmDownwardAPIVolumeFile(api.Fields),
		VolumeLabel: api.VolumeLabel,
	}
}

func ToHapiVmDownwardAPIVolumeFile(fields []v1.DownwardAPIVolumeFile) []*models.V1VMDownwardAPIVolumeFile {
	if fields == nil {
		return nil
	}

	var result []*models.V1VMDownwardAPIVolumeFile
	for _, field := range fields {
		result = append(result, ToHapiVmDownwardAPIVolumeFileItem(field))
	}

	return result
}

func ToHapiVmDownwardAPIVolumeFileItem(field v1.DownwardAPIVolumeFile) *models.V1VMDownwardAPIVolumeFile {
	return &models.V1VMDownwardAPIVolumeFile{
		Path:     ptr.StringPtr(field.Path),
		FieldRef: ToHapiVmObjectFieldSelector(field.FieldRef),
		ResourceFieldRef: ToHapiVmResourceFieldSelector(
			field.ResourceFieldRef,
		),
		Mode: *field.Mode,
	}
}

func ToHapiVmResourceFieldSelector(ref *v1.ResourceFieldSelector) *models.V1VMResourceFieldSelector {
	if ref == nil {
		return nil
	}

	return &models.V1VMResourceFieldSelector{
		ContainerName: ref.ContainerName,
		Resource:      ptr.StringPtr(ref.Resource),
		Divisor:       ToHapiVmQuantityDivisor(ref.Divisor),
	}
}

func ToHapiVmObjectFieldSelector(ref *v1.ObjectFieldSelector) *models.V1VMObjectFieldSelector {
	if ref == nil {
		return nil
	}

	return &models.V1VMObjectFieldSelector{
		FieldPath: ptr.StringPtr(ref.FieldPath),
	}
}
