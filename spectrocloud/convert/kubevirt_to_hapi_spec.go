package convert

import (
	"fmt"

	"encoding/json"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ToHapiVmSpecM(spec kubevirtapiv1.VirtualMachineSpec) (*models.V1ClusterVirtualMachineSpec, error) {
	var hapiVmSpec models.V1ClusterVirtualMachineSpec

	// Marshal the input spec to JSON
	specJson, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal kubevirtapiv1.VirtualMachineSpec to JSON: %v", err)
	}

	// Unmarshal the JSON to the desired HAPI VM spec
	err = json.Unmarshal(specJson, &hapiVmSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to models.V1ClusterVirtualMachineSpec: %v", err)
	}

	return &hapiVmSpec, nil
}

// TemplateSpecToHapi converts KubeVirt VirtualMachineInstanceTemplateSpec to HAPI model (JSON-compatible).
func TemplateSpecToHapi(template *kubevirtapiv1.VirtualMachineInstanceTemplateSpec) (*models.V1VMVirtualMachineInstanceTemplateSpec, error) {
	if template == nil {
		return nil, nil
	}
	data, err := json.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal VirtualMachineInstanceTemplateSpec: %w", err)
	}
	var out models.V1VMVirtualMachineInstanceTemplateSpec
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to V1VMVirtualMachineInstanceTemplateSpec: %w", err)
	}
	return &out, nil
}

// DataVolumeTemplatesToHapi converts KubeVirt DataVolumeTemplateSpec slice to HAPI models (JSON-compatible).
func DataVolumeTemplatesToHapi(templates []kubevirtapiv1.DataVolumeTemplateSpec) ([]*models.V1VMDataVolumeTemplateSpec, error) {
	if len(templates) == 0 {
		return nil, nil
	}
	data, err := json.Marshal(templates)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DataVolumeTemplateSpec: %w", err)
	}
	var out []*models.V1VMDataVolumeTemplateSpec
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to V1VMDataVolumeTemplateSpec: %w", err)
	}
	return out, nil
}

// ObjectMetaToHapiVmMeta maps Kubernetes ObjectMeta to Spectro HAPI VM object metadata.
func ObjectMetaToHapiVmMeta(meta metav1.ObjectMeta) *models.V1VMObjectMeta {
	var gracePeriodSeconds int64
	if meta.DeletionGracePeriodSeconds != nil {
		gracePeriodSeconds = *meta.DeletionGracePeriodSeconds
	}
	return &models.V1VMObjectMeta{
		Annotations:                meta.Annotations,
		DeletionGracePeriodSeconds: gracePeriodSeconds,
		Finalizers:                 meta.Finalizers,
		GenerateName:               meta.GenerateName,
		Generation:                 meta.Generation,
		Labels:                     meta.Labels,
		ManagedFields:              ToHapiVmManagedFields(meta.ManagedFields),
		Name:                       meta.Name,
		Namespace:                  meta.Namespace,
		OwnerReferences:            ToHapiVmOwnerReferences(meta.OwnerReferences),
		ResourceVersion:            meta.ResourceVersion,
		UID:                        string(meta.UID),
	}
}
