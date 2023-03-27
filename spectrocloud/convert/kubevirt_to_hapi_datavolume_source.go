package convert

import (
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

func ToHapiVmDataVolumeSource(source *cdiv1.DataVolumeSource) *models.V1VMDataVolumeSource {
	if source == nil {
		return nil
	}
	return &models.V1VMDataVolumeSource{
		Blank:    ToHapiVmDataVolumeBlankSource(source.Blank),
		HTTP:     ToHapiVmDataVolumeHTTPSource(source.HTTP),
		Imageio:  ToHapiVmDataVolumeImageioSource(source.Imageio),
		Pvc:      ToHapiVmDataVolumePvcSource(source.PVC),
		Registry: ToHapiVmDataVolumeRegistrySource(source.Registry),
		S3:       ToHapiVmDataVolumeS3Source(source.S3),
		//TODO: Upload:   ToHapiVmDataVolumeUploadSource(source.Upload),
		Vddk: ToHapiVmDataVolumeVddkSource(source.VDDK),
	}
}

func ToHapiVmDataVolumeUploadSource(upload *cdiv1.DataVolumeSourceUpload) models.V1VMDataVolumeSourceUpload {
	if upload == nil {
		return ""
	}
	return make(map[string]interface{})
}

func ToHapiVmDataVolumeS3Source(s3 *cdiv1.DataVolumeSourceS3) *models.V1VMDataVolumeSourceS3 {
	if s3 == nil {
		return nil
	}
	return &models.V1VMDataVolumeSourceS3{
		CertConfigMap: s3.CertConfigMap,
		SecretRef:     s3.SecretRef,
		URL:           ptr.StringPtr(s3.URL),
	}
}

func ToHapiVmDataVolumeRegistrySource(registry *cdiv1.DataVolumeSourceRegistry) *models.V1VMDataVolumeSourceRegistry {
	if registry == nil {
		return nil
	}

	var CertConfigMap string
	if registry.CertConfigMap != nil {
		CertConfigMap = *registry.CertConfigMap
	}

	var ImageStream string
	if registry.ImageStream != nil {
		ImageStream = *registry.ImageStream
	}

	var PullMethod string
	if registry.PullMethod != nil {
		PullMethod = string(*registry.PullMethod)
	}

	var SecretRef string
	if registry.SecretRef != nil {
		SecretRef = *registry.SecretRef
	}

	return &models.V1VMDataVolumeSourceRegistry{
		CertConfigMap: CertConfigMap,
		ImageStream:   ImageStream,
		PullMethod:    PullMethod,
		SecretRef:     SecretRef,
		URL:           ptr.String(registry.URL),
	}
}

func ToHapiVmDataVolumePvcSource(pvc *cdiv1.DataVolumeSourcePVC) *models.V1VMDataVolumeSourcePVC {
	if pvc == nil {
		return nil
	}
	return &models.V1VMDataVolumeSourcePVC{
		Namespace: ptr.StringPtr(pvc.Namespace),
		Name:      ptr.StringPtr(pvc.Name),
	}
}

func ToHapiVmDataVolumeVddkSource(vddk *cdiv1.DataVolumeSourceVDDK) *models.V1VMDataVolumeSourceVDDK {
	if vddk == nil {
		return nil
	}
	return &models.V1VMDataVolumeSourceVDDK{
		BackingFile:  vddk.BackingFile,
		InitImageURL: vddk.InitImageURL,
		SecretRef:    vddk.SecretRef,
		Thumbprint:   vddk.Thumbprint,
		URL:          vddk.URL,
		UUID:         vddk.UUID,
	}
}

func ToHapiVmDataVolumeImageioSource(imageio *cdiv1.DataVolumeSourceImageIO) *models.V1VMDataVolumeSourceImageIO {
	if imageio == nil {
		return nil
	}
	return &models.V1VMDataVolumeSourceImageIO{
		URL:           ptr.StringPtr(imageio.URL),
		SecretRef:     imageio.SecretRef,
		CertConfigMap: imageio.CertConfigMap,
		DiskID:        ptr.StringPtr(imageio.DiskID),
	}
}

func ToHapiVmDataVolumeHTTPSource(http *cdiv1.DataVolumeSourceHTTP) *models.V1VMDataVolumeSourceHTTP {
	if http == nil {
		return nil
	}
	return &models.V1VMDataVolumeSourceHTTP{
		CertConfigMap:      http.CertConfigMap,
		ExtraHeaders:       http.ExtraHeaders,
		SecretExtraHeaders: http.SecretExtraHeaders,
		SecretRef:          http.SecretRef,
		URL:                ptr.StringPtr(http.URL),
	}
}

func ToHapiVmDataVolumeBlankSource(blank *cdiv1.DataVolumeBlankImage) models.V1VMDataVolumeBlankImage {
	if blank == nil {
		return nil
	}
	return make(map[string]interface{})
}
