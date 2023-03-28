package convert

import (
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

func ToKubevirtDataVolumeTemplate(dataVolTemplates []*models.V1VMDataVolumeTemplateSpec) []kubevirtapiv1.DataVolumeTemplateSpec {
	var dataVolumeTemplateSpec []kubevirtapiv1.DataVolumeTemplateSpec

	for _, dv := range dataVolTemplates {
		dataVolumeTemplateSpec = append(dataVolumeTemplateSpec, kubevirtapiv1.DataVolumeTemplateSpec{
			ObjectMeta: ToKubevirtDataVolumeMetadata(dv.Metadata),
			Spec:       ToKubevirtDataVolumeSpec(dv.Spec),
		})
	}
	return dataVolumeTemplateSpec
}

func ToKubevirtDataVolumeMetadata(dataVolMetadata *models.V1VMObjectMeta) metav1.ObjectMeta {
	dataVolumeMetadata := metav1.ObjectMeta{
		Name:      dataVolMetadata.Name,
		Namespace: dataVolMetadata.Namespace,
		//UID:       types.UID(dataVolMetadata.UID),
		//ResourceVersion: dataVolMetadata.ResourceVersion,
		//Generation:      dataVolMetadata.Generation,
		//Labels:          dataVolMetadata.Labels,
		//Annotations:     dataVolMetadata.Annotations,
	}
	return dataVolumeMetadata
}

func ToKubevirtDataVolumeSpec(hapiDataVolSpec *models.V1VMDataVolumeSpec) cdiv1.DataVolumeSpec {
	kubevirtDataVolSpec := cdiv1.DataVolumeSpec{
		Source:            nil,
		SourceRef:         nil,
		PVC:               ToKubevirtDataVolumeSpecPVC(hapiDataVolSpec.Pvc),
		Storage:           nil,
		PriorityClassName: "",
		ContentType:       "",
		Checkpoints:       nil,
		FinalCheckpoint:   false,
		Preallocation:     nil,
	}
	return kubevirtDataVolSpec
}

func ToKubevirtDataVolumeSpecPVC(pvc *models.V1VMPersistentVolumeClaimSpec) *corev1.PersistentVolumeClaimSpec {
	kubevirtPVC := corev1.PersistentVolumeClaimSpec{
		AccessModes: ToKubevirtPVCAccessMode(pvc.AccessModes),
		//Selector:         nil,
		//Resources:        corev1.ResourceRequirements{},
		//VolumeName:       "",
		StorageClassName: types.Ptr(pvc.StorageClassName),
		//VolumeMode:       nil,
		//DataSource:       nil,
		//DataSourceRef:    nil,
	}
	return &kubevirtPVC
}

func ToKubevirtPVCAccessMode(accessMode []string) []corev1.PersistentVolumeAccessMode {
	var PVCAccessMode []corev1.PersistentVolumeAccessMode
	for _, acc := range accessMode {
		PVCAccessMode = append(PVCAccessMode, corev1.PersistentVolumeAccessMode(acc))
	}
	return PVCAccessMode
}
