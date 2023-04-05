package convert

import (
	"github.com/spectrocloud/hapi/models"
	"k8s.io/api/core/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func ToHapiVmConfigMap(configMap *kubevirtapiv1.ConfigMapVolumeSource) *models.V1VMConfigMapVolumeSource {
	if configMap == nil {
		return nil
	}

	return &models.V1VMConfigMapVolumeSource{
		Name:        configMap.Name,
		Optional:    *configMap.Optional,
		VolumeLabel: configMap.VolumeLabel,
	}
}

func ToHapiVmDataVolume(volume *kubevirtapiv1.DataVolumeSource) *models.V1VMCoreDataVolumeSource {
	if volume == nil {
		return nil
	}

	return &models.V1VMCoreDataVolumeSource{
		Name:         types.Ptr(volume.Name),
		Hotpluggable: volume.Hotpluggable,
	}
}

func ToHapiVmEphemeral(ephemeral *kubevirtapiv1.EphemeralVolumeSource) *models.V1VMEphemeralVolumeSource {
	if ephemeral == nil {
		return nil
	}

	return &models.V1VMEphemeralVolumeSource{
		PersistentVolumeClaim: ToHapiVmPersistentVolumeClaim(ephemeral.PersistentVolumeClaim),
	}
}

func ToHapiVmPersistentVolumeClaim(claim *v1.PersistentVolumeClaimVolumeSource) *models.V1VMPersistentVolumeClaimVolumeSource {
	if claim == nil {
		return nil
	}

	return &models.V1VMPersistentVolumeClaimVolumeSource{
		ClaimName: types.Ptr(claim.ClaimName),
		// TODO: Hotpluggable: claim.Hotpluggable, NO SUCH FIELD!
		ReadOnly: claim.ReadOnly,
	}
}
