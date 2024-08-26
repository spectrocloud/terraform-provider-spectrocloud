package convert

import (
	"encoding/json"
	"fmt"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

func ToHapiVolume(volume *cdiv1.DataVolume, addVolumeOptions *models.V1VMAddVolumeOptions) (*models.V1VMAddVolumeEntity, error) {
	var GracePeriodSeconds int64
	if volume.DeletionGracePeriodSeconds != nil {
		GracePeriodSeconds = *volume.DeletionGracePeriodSeconds
	}

	Spec, err := ToHapiVolumeSpecM(volume.Spec)
	if err != nil {
		return nil, err
	}

	hapiVolume := &models.V1VMAddVolumeEntity{
		AddVolumeOptions: addVolumeOptions,
		DataVolumeTemplate: &models.V1VMDataVolumeTemplateSpec{
			Metadata: &models.V1VMObjectMeta{
				Annotations:                volume.Annotations,
				DeletionGracePeriodSeconds: GracePeriodSeconds,
				Finalizers:                 volume.Finalizers,
				GenerateName:               volume.GenerateName,
				Generation:                 volume.Generation,
				Labels:                     volume.Labels,
				ManagedFields:              ToHapiVmManagedFields(volume.ManagedFields),
				Name:                       volume.ObjectMeta.Name,
				Namespace:                  volume.ObjectMeta.Namespace,
				OwnerReferences:            ToHapiVmOwnerReferences(volume.OwnerReferences),
				ResourceVersion:            volume.ResourceVersion,
				UID:                        string(volume.UID),
			},
			Spec: Spec,
		},
		Persist: true,
	}
	return hapiVolume, nil
}

func ToHapiVolumeSpecM(spec cdiv1.DataVolumeSpec) (*models.V1VMDataVolumeSpec, error) {
	var hapiVolumeSpec models.V1VMDataVolumeSpec

	// Marshal the input spec to JSON
	specJson, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal kubevirtapiv1.VirtualMachineSpec to JSON: %v", err)
	}

	// Unmarshal the JSON to the desired HAPI VM spec
	err = json.Unmarshal(specJson, &hapiVolumeSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to models.V1ClusterVirtualMachineSpec: %v", err)
	}

	return &hapiVolumeSpec, nil
}
