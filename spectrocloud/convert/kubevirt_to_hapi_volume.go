package convert

import (
	"github.com/spectrocloud/hapi/models"
	"k8s.io/api/core/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func ToHapiVmVolumes(volumes []kubevirtapiv1.Volume) []*models.V1VMVolume {
	var Volumes []*models.V1VMVolume
	for _, volume := range volumes {
		Volumes = append(Volumes, ToHapiVmVolume(volume))
	}
	return Volumes
}

func ToHapiVmVolume(volume kubevirtapiv1.Volume) *models.V1VMVolume {
	var DownwardMetrics map[string]interface{}
	if volume.DownwardMetrics != nil {
		DownwardMetrics = make(map[string]interface{})
	}

	return &models.V1VMVolume{
		CloudInitConfigDrive:  ToHapiVmCloudInitConfigDrive(volume.CloudInitConfigDrive),
		CloudInitNoCloud:      ToHapiVmCloudInitNoCloud(volume.CloudInitNoCloud),
		ConfigMap:             ToHapiVmConfigMap(volume.ConfigMap),
		ContainerDisk:         ToHapiVmContainerDisk(volume.ContainerDisk),
		DataVolume:            ToHapiVmDataVolume(volume.DataVolume),
		DownwardAPI:           ToHapiVmDownwardAPI(volume.DownwardAPI),
		DownwardMetrics:       DownwardMetrics,
		EmptyDisk:             ToHapiVmEmptyDisk(volume.EmptyDisk),
		Ephemeral:             ToHapiVmEphemeral(volume.Ephemeral),
		HostDisk:              ToHapiVmHostDisk(volume.HostDisk),
		MemoryDump:            ToHapiVmMemoryDump(volume.MemoryDump),
		Name:                  types.Ptr(volume.Name),
		PersistentVolumeClaim: ToHapiVmPersistentVolumeClaimKubevirt(volume.PersistentVolumeClaim),
		Secret:                ToHapiVMSecret(volume.Secret),
		ServiceAccount:        ToHapiVMServiceAccount(volume.ServiceAccount),
		Sysprep:               ToHapiVMSysprep(volume.Sysprep),
	}
}

func ToHapiVMSysprep(sysprep *kubevirtapiv1.SysprepSource) *models.V1VMSysprepSource {
	if sysprep == nil {
		return nil
	}

	return &models.V1VMSysprepSource{
		ConfigMap: ToHapiVmLocalObjectReference(sysprep.ConfigMap),
		Secret:    nil,
	}
}

func ToHapiVmLocalObjectReference(configMap *v1.LocalObjectReference) *models.V1VMLocalObjectReference {
	if configMap == nil {
		return nil
	}

	return &models.V1VMLocalObjectReference{
		Name: configMap.Name,
	}
}

func ToHapiVmPersistentVolumeClaimKubevirt(claim *kubevirtapiv1.PersistentVolumeClaimVolumeSource) *models.V1VMPersistentVolumeClaimVolumeSource {
	if claim == nil {
		return nil
	}

	return &models.V1VMPersistentVolumeClaimVolumeSource{
		ClaimName:    types.Ptr(claim.ClaimName),
		Hotpluggable: claim.Hotpluggable,
		ReadOnly:     claim.ReadOnly,
	}
}

func ToHapiVmMemoryDump(dump *kubevirtapiv1.MemoryDumpVolumeSource) *models.V1VMMemoryDumpVolumeSource {
	if dump == nil {
		return nil
	}

	return &models.V1VMMemoryDumpVolumeSource{
		ClaimName:    types.Ptr(dump.ClaimName),
		Hotpluggable: dump.Hotpluggable,
		ReadOnly:     dump.ReadOnly,
	}
}

func ToHapiVMServiceAccount(account *kubevirtapiv1.ServiceAccountVolumeSource) *models.V1VMServiceAccountVolumeSource {
	if account == nil {
		return nil
	}

	return &models.V1VMServiceAccountVolumeSource{
		ServiceAccountName: account.ServiceAccountName,
	}
}

func ToHapiVMPersistentVolumeClaimKubevirt(claim *kubevirtapiv1.PersistentVolumeClaimVolumeSource) *models.V1VMPersistentVolumeClaimVolumeSource {
	if claim == nil {
		return nil
	}

	return &models.V1VMPersistentVolumeClaimVolumeSource{
		ClaimName:    types.Ptr(claim.ClaimName),
		Hotpluggable: claim.Hotpluggable,
		ReadOnly:     claim.ReadOnly,
	}
}

func ToHapiVMSecret(secret *kubevirtapiv1.SecretVolumeSource) *models.V1VMSecretVolumeSource {
	if secret == nil {
		return nil
	}

	return &models.V1VMSecretVolumeSource{
		Optional:    false,
		SecretName:  "",
		VolumeLabel: "",
	}
}

func ToHapiVMMemoryDump(dump *kubevirtapiv1.MemoryDumpVolumeSource) *models.V1VMMemoryDumpVolumeSource {
	if dump == nil {
		return nil
	}

	return &models.V1VMMemoryDumpVolumeSource{
		ClaimName:    nil,
		Hotpluggable: false,
		ReadOnly:     false,
	}
}
