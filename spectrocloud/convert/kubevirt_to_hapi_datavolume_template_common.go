package convert

import (
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

func ToHapiVmLabelSelector(selector *metav1.LabelSelector) *models.V1VMLabelSelector {
	if selector == nil {
		return nil
	}
	return &models.V1VMLabelSelector{
		MatchExpressions: ToHapiVmLabelSelectorRequirements(selector.MatchExpressions),
		MatchLabels:      selector.MatchLabels,
	}
}

func ToHapiVmLabelSelectorRequirements(expressions []metav1.LabelSelectorRequirement) []*models.V1VMLabelSelectorRequirement {
	var hapiRequirements []*models.V1VMLabelSelectorRequirement
	for _, expression := range expressions {
		hapiRequirements = append(hapiRequirements, ToHapiVmLabelSelectorRequirement(expression))
	}
	return hapiRequirements
}

func ToHapiVmLabelSelectorRequirement(expression metav1.LabelSelectorRequirement) *models.V1VMLabelSelectorRequirement {
	return &models.V1VMLabelSelectorRequirement{
		Key:      ptr.StringPtr(expression.Key),
		Operator: ptr.StringPtr(string(expression.Operator)),
		Values:   expression.Values,
	}
}

func ToHapiVmTypedObjectReference(ref *corev1.TypedObjectReference) *models.V1VMTypedLocalObjectReference {
	if ref == nil {
		return nil
	}
	return &models.V1VMTypedLocalObjectReference{
		APIGroup: *ref.APIGroup,
		Kind:     ptr.StringPtr(ref.Kind),
		Name:     ptr.StringPtr(ref.Name),
	}
}

func ToHapiVmCheckpoints(checkpoints []cdiv1.DataVolumeCheckpoint) []*models.V1VMDataVolumeCheckpoint {
	var hapiCheckpoints []*models.V1VMDataVolumeCheckpoint
	for _, checkpoint := range checkpoints {
		hapiCheckpoints = append(hapiCheckpoints, ToHapiVmCheckpoint(checkpoint))
	}
	return hapiCheckpoints
}

func ToHapiVmCheckpoint(checkpoint cdiv1.DataVolumeCheckpoint) *models.V1VMDataVolumeCheckpoint {
	return &models.V1VMDataVolumeCheckpoint{
		Current:  ptr.StringPtr(checkpoint.Current),
		Previous: ptr.StringPtr(checkpoint.Previous),
	}
}

func ToHapiVmResourceRequirements(resources corev1.ResourceRequirements) *models.V1VMCoreResourceRequirements {
	return &models.V1VMCoreResourceRequirements{
		Limits:   ToHapiVmResourceList(resources.Limits),
		Requests: ToHapiVmResourceList(resources.Requests),
	}
}

func ToHapiVmResourceList(limits corev1.ResourceList) map[string]models.V1VMQuantity {
	hapiLimits := make(map[string]models.V1VMQuantity)
	for k, v := range limits {
		if str := v.String(); str != "" {
			hapiLimits[string(k)] = models.V1VMQuantity(str)
		}
	}
	return hapiLimits
}

func ToHapiVmAccessModes(accessModes []corev1.PersistentVolumeAccessMode) []string {
	var hapiAccessModes []string
	for _, accessMode := range accessModes {
		hapiAccessModes = append(hapiAccessModes, string(accessMode))
	}
	return hapiAccessModes
}

func ToHapiVmTypedLocalObjectReference(ref *corev1.TypedLocalObjectReference) *models.V1VMTypedLocalObjectReference {
	if ref == nil {
		return nil
	}
	return &models.V1VMTypedLocalObjectReference{
		APIGroup: *ref.APIGroup,
		Kind:     ptr.StringPtr(ref.Kind),
		Name:     ptr.StringPtr(ref.Name),
	}
}
