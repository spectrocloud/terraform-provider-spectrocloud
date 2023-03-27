package convert

import (
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	corev1 "k8s.io/api/core/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

func ToHapiVmDataVolumeTemplates(templates []kubevirtapiv1.DataVolumeTemplateSpec) []*models.V1VMDataVolumeTemplateSpec {
	var hapiTemplates []*models.V1VMDataVolumeTemplateSpec
	for _, template := range templates {
		hapiTemplates = append(hapiTemplates, ToHapiVmDataVolumeTemplate(template))
	}
	return hapiTemplates
}

func ToHapiVmDataVolumeTemplate(template kubevirtapiv1.DataVolumeTemplateSpec) *models.V1VMDataVolumeTemplateSpec {
	return &models.V1VMDataVolumeTemplateSpec{
		APIVersion: template.APIVersion,
		Kind:       template.Kind,
		Metadata:   ToHapiVmObjectMeta(template.ObjectMeta),
		Spec:       ToHapiVmDataVolumeSpec(template.Spec),
	}
}

func ToHapiVmDataVolumeSpec(spec cdiv1.DataVolumeSpec) *models.V1VMDataVolumeSpec {
	return &models.V1VMDataVolumeSpec{
		Checkpoints:       ToHapiVmCheckpoints(spec.Checkpoints),
		ContentType:       string(spec.ContentType),
		FinalCheckpoint:   spec.FinalCheckpoint,
		Preallocation:     *DefaultBool(spec.Preallocation, false),
		PriorityClassName: spec.PriorityClassName,
		Pvc:               ToHapiVmPersistentVolumeClaimSpec(spec.PVC),
		Source:            ToHapiVmDataVolumeSource(spec.Source),
		SourceRef:         ToHapiVmDataVolumeReference(spec.SourceRef),
		Storage:           ToHapiVmDataVolumeStorage(spec.Storage),
	}
}

func ToHapiVmDataVolumeStorage(storage *cdiv1.StorageSpec) *models.V1VMStorageSpec {
	if storage == nil {
		return nil
	}

	var VolumeMode string
	if storage.VolumeMode != nil {
		VolumeMode = string(*storage.VolumeMode)
	}

	return &models.V1VMStorageSpec{
		AccessModes:      ToHapiVmAccessModes(storage.AccessModes),
		DataSource:       ToHapiVmTypedLocalObjectReference(storage.DataSource),
		Resources:        ToHapiVmResourceRequirements(storage.Resources),
		Selector:         ToHapiVmLabelSelector(storage.Selector),
		StorageClassName: *storage.StorageClassName,
		VolumeMode:       VolumeMode,
		VolumeName:       storage.VolumeName,
	}
}

func ToHapiVmDataVolumeReference(ref *cdiv1.DataVolumeSourceRef) *models.V1VMDataVolumeSourceRef {
	if ref == nil {
		return nil
	}
	return &models.V1VMDataVolumeSourceRef{
		Kind:      ptr.StringPtr(ref.Kind),
		Name:      ptr.StringPtr(ref.Name),
		Namespace: *ref.Namespace,
	}
}

func ToHapiVmPersistentVolumeClaimSpec(pvc *corev1.PersistentVolumeClaimSpec) *models.V1VMPersistentVolumeClaimSpec {
	if pvc == nil {
		return nil
	}

	var StorageClassName string
	if pvc.StorageClassName != nil {
		StorageClassName = *pvc.StorageClassName
	}
	var VolumeMode string
	if pvc.VolumeMode != nil {
		VolumeMode = string(*pvc.VolumeMode)
	}
	return &models.V1VMPersistentVolumeClaimSpec{
		AccessModes:      ToHapiVmAccessModes(pvc.AccessModes),
		DataSource:       ToHapiVmTypedLocalObjectReference(pvc.DataSource),
		DataSourceRef:    ToHapiVmTypedObjectReference(pvc.DataSourceRef),
		Resources:        ToHapiVmResourceRequirements(pvc.Resources),
		Selector:         ToHapiVmLabelSelector(pvc.Selector),
		StorageClassName: StorageClassName,
		VolumeMode:       VolumeMode,
		VolumeName:       pvc.VolumeName,
	}
}
