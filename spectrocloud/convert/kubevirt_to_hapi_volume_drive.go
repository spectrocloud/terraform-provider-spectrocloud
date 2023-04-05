package convert

import (
	"github.com/spectrocloud/hapi/models"
	"k8s.io/api/core/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ToHapiVmCloudInitConfigDrive(drive *kubevirtapiv1.CloudInitConfigDriveSource) *models.V1VMCloudInitConfigDriveSource {
	if drive == nil {
		return nil
	}

	return &models.V1VMCloudInitConfigDriveSource{
		NetworkData:          drive.NetworkData,
		NetworkDataBase64:    drive.NetworkDataBase64,
		NetworkDataSecretRef: ToHapiVmObjectReference(drive.NetworkDataSecretRef),
		SecretRef:            ToHapiVmObjectReference(drive.UserDataSecretRef),
		UserData:             drive.UserData,
		UserDataBase64:       drive.UserDataBase64,
	}
}

func ToHapiVmSysprep(sysprep *kubevirtapiv1.SysprepSource) *models.V1VMSysprepSource {
	if sysprep == nil {
		return nil
	}

	return &models.V1VMSysprepSource{
		ConfigMap: ToHapiVmObjectReference(sysprep.ConfigMap),
		Secret:    ToHapiVmObjectReference(sysprep.Secret),
	}
}

func ToHapiVmObjectReference(ref *v1.LocalObjectReference) *models.V1VMLocalObjectReference {
	if ref == nil {
		return nil
	}

	return &models.V1VMLocalObjectReference{
		Name: ref.Name,
	}
}
